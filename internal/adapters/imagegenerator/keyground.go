// FILE: internal/adapters/imagegenerator/keyground.go
//
// bugs_open/424 — the matte half of the background-key mechanism. The prompt
// half (LogoBackgroundKeyClause) lives in
// platform/orchestration/actions/discovery_checks/default_brand_prompt.go and
// asks the model to paint a flat, deterministic, saturated colour as the
// ground. This file removes that colour mathematically, producing a real PNG
// alpha channel — the property the model itself cannot express.
//
// WHY BORDER FLOOD-FILL, NOT "every pixel near this colour". A global colour-
// distance pass would punch a hole through any interior element that merely
// resembles the key, which is exactly the "clipped mark" failure mode a naive
// background-removal pass risks (see bugs_open/424's fix-candidate discussion).
// Flood-filling outward from the four image edges instead only erases
// background that is CONNECTED to the outside of the picture — a disconnected
// interior region of similar colour survives the border pass untouched. That
// leaves one narrower hole (an enclosed area — a ring, a counter — that is
// genuinely background but not reachable from the border), handled as a
// second, tighter pass; see KeyOutBackground's doc comment for why that pass's
// safety rests on the prompt rather than on this code.
package imagegenerator

import (
	"image"
	"image/color"
	"math"
)

// MatteStats reports what KeyOutBackground actually did, so a caller can
// decide whether the model honoured the key-colour instruction at all.
type MatteStats struct {
	// BorderKeyed is the fraction (0..1) of the image's outermost ring of
	// pixels whose FINAL alpha is 0 — genuinely, fully transparent in the
	// output. A model that ignored the instruction — painted a checkerboard,
	// a solid unrelated ground, a vignette — leaves this near zero.
	//
	// CORRECTED 2026-09-02 (round 2, contributed by the bugfix_417_420 lane's
	// live dynamic testing, CONTRIB round 3): this field's DOC COMMENT always
	// said "ended up fully transparent", but the CODE computed it from BFS
	// flood-fill REACHABILITY (dist <= outer) instead — a border wash that
	// merely sat within the graded band (nowhere near inner, so alpha stayed
	// well above 0 everywhere) still marked every border pixel "reachable"
	// and reported BorderKeyed=1.000, identical to a genuine 87.4%-transparent
	// success. Measured live: a 0.0%-actually-transparent failure and an
	// 87.4%-actually-transparent success both read 1.000 on the old code.
	// The fail-closed guard this field exists to drive was therefore
	// evaluating the wrong thing entirely — see keyOutBackground's threshold
	// check in dynamic_adapter.go, unchanged by this fix, which now finally
	// gates on what its own comment always claimed it gated on.
	BorderKeyed float64

	// Keyed is the total count of pixels with alpha < 255 after both passes.
	Keyed int
}

// KeyOutBackground returns a copy of img with key-coloured ground made
// transparent, plus stats a caller can use to decide whether the result is
// trustworthy (see MatteStats.BorderKeyed).
//
// inner and outer are Euclidean distances in 8-bit RGB space (each channel
// compared as if scaled 0-255, regardless of img's source bit depth — see
// colourDistance) from key. A pixel at distance <= inner is fully
// transparent; at distance >= outer, fully opaque; between the two, alpha is
// graded linearly and the foreground colour is unmixed against key (despill),
// so an anti-aliased edge keeps the mark's own colour rather than a fringe of
// the key colour.
//
// Two passes, deliberately asymmetric in what they are allowed to touch:
//
//  1. Border flood-fill (outer threshold): BFS from every edge pixel within
//     outer of key, through 4-connected neighbours also within outer. This is
//     the safe pass — it can only ever erase background that is connected to
//     the outside of the picture.
//  2. Enclosed holes (inner threshold only, whole image): a ring or a counter
//     is genuine background the border pass cannot reach. Any pixel ANYWHERE
//     within inner of key is also keyed, regardless of connectivity. This
//     pass has no structural safety property — a mark that legitimately used
//     a near-exact key colour would be holed by it too. Its safety rests
//     entirely on the prompt forbidding that (LogoBackgroundKeyClause's
//     negative-prompt belt, logoBackgroundNegatives in generate_image_actions.go).
//
// img is read through the generic image.Image/color.Color interface
// throughout — never assumed to be a concrete 8-bit type — because a real
// generated logo asset has been measured at 16-bit PNG depth (bugs_open/424
// verification), and color.Color.RGBA() already normalises any source depth
// to a 16-bit range.
func KeyOutBackground(img image.Image, key color.Color, inner, outer float64) (*image.NRGBA, MatteStats) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, w, h))

	kr, kg, kb := to8Bit(key)

	// Precompute distance and raw (unmixed) colour for every pixel once.
	dist := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bl := to8Bit(img.At(b.Min.X+x, b.Min.Y+y))
			d := colourDistance(r, g, bl, kr, kg, kb)
			dist[y*w+x] = d
			out.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: bl, A: 255})
		}
	}

	idx := func(x, y int) int { return y*w + x }
	keyedByBorder := make([]bool, w*h)

	// Pass 1: border flood-fill, BFS within `outer`.
	type pt struct{ x, y int }
	queue := make([]pt, 0, w+h)
	seed := func(x, y int) {
		if x < 0 || y < 0 || x >= w || y >= h {
			return
		}
		i := idx(x, y)
		if keyedByBorder[i] || dist[i] > outer {
			return
		}
		keyedByBorder[i] = true
		queue = append(queue, pt{x, y})
	}
	for x := 0; x < w; x++ {
		seed(x, 0)
		seed(x, h-1)
	}
	for y := 0; y < h; y++ {
		seed(0, y)
		seed(w-1, y)
	}
	for len(queue) > 0 {
		p := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		seed(p.x+1, p.y)
		seed(p.x-1, p.y)
		seed(p.x, p.y+1)
		seed(p.x, p.y-1)
	}

	stats := MatteStats{}

	// finalAlpha records what alpha each pixel actually ended up with, so the
	// border stat below can be computed from the REAL result rather than from
	// keyedByBorder (mere BFS reachability — see MatteStats.BorderKeyed's
	// corrected doc comment for why that was wrong). Defaults to 255
	// (untouched pixels never enter the loop below).
	finalAlpha := make([]uint8, w*h)
	for i := range finalAlpha {
		finalAlpha[i] = 255
	}

	// Pass 2 + edge grading, per pixel: border-flood pixels and any pixel
	// within `inner` anywhere (the enclosed-hole pass) become transparent or
	// partially transparent; everything else keeps alpha 255 with its own
	// colour, unchanged.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := idx(x, y)
			d := dist[i]
			enclosed := d <= inner
			if !keyedByBorder[i] && !enclosed {
				continue // untouched — outside both passes
			}

			c := out.NRGBAAt(x, y)
			switch {
			case d <= inner:
				out.SetNRGBA(x, y, color.NRGBA{A: 0})
				finalAlpha[i] = 0
				stats.Keyed++
			case d >= outer:
				// A border-pass pixel sitting exactly at the outer boundary
				// (BFS admits dist <= outer). Stays opaque — this is the
				// harmless edge of the graded range, not an error case.
			default:
				// Graded edge: alpha ramps from 0 (at inner) to 255 (at outer).
				alpha := (d - inner) / (outer - inner)
				a8 := uint8(math.Round(alpha * 255))
				fg := despill(c.R, c.G, c.B, kr, kg, kb, alpha)
				out.SetNRGBA(x, y, color.NRGBA{R: fg[0], G: fg[1], B: fg[2], A: a8})
				finalAlpha[i] = a8
				stats.Keyed++
			}
		}
	}

	// Border stat, computed from finalAlpha (what actually happened), not
	// from keyedByBorder (what was merely close enough to be eligible).
	borderRing := 0
	borderTransparent := 0
	countBorder := func(x, y int) {
		borderRing++
		if finalAlpha[idx(x, y)] == 0 {
			borderTransparent++
		}
	}
	for x := 0; x < w; x++ {
		countBorder(x, 0)
		if h > 1 {
			countBorder(x, h-1)
		}
	}
	for y := 1; y < h-1; y++ {
		countBorder(0, y)
		if w > 1 {
			countBorder(w-1, y)
		}
	}
	if borderRing > 0 {
		stats.BorderKeyed = float64(borderTransparent) / float64(borderRing)
	}

	return out, stats
}

// to8Bit reads a color.Color through its RGBA() method (always 16-bit range,
// alpha-premultiplied by definition — but every pixel this function is ever
// called on is fully opaque source data from a freshly generated image, so
// premultiplication is a no-op) and scales down to 8-bit per channel.
func to8Bit(c color.Color) (r, g, b uint8) {
	r32, g32, b32, _ := c.RGBA()
	return uint8(r32 >> 8), uint8(g32 >> 8), uint8(b32 >> 8)
}

// colourDistance is the Euclidean distance between two 8-bit RGB colours.
func colourDistance(r1, g1, b1, r2, g2, b2 uint8) float64 {
	dr := float64(r1) - float64(r2)
	dg := float64(g1) - float64(g2)
	db := float64(b1) - float64(b2)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// despill unmixes a graded-edge pixel against the key colour: assuming the
// stored pixel p is a linear blend of the true foreground colour and the key
// colour in proportion alpha, recovers fg = (p - (1-alpha)*key) / alpha,
// clamped to [0,255]. Without this, a partially-keyed edge pixel keeps a
// visible fringe of the key colour once composited over anything but the key.
func despill(pr, pg, pb, kr, kg, kb uint8, alpha float64) [3]uint8 {
	if alpha <= 0 {
		return [3]uint8{0, 0, 0}
	}
	unmix := func(p, k uint8) uint8 {
		v := (float64(p) - (1-alpha)*float64(k)) / alpha
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}
		return uint8(math.Round(v))
	}
	return [3]uint8{unmix(pr, kr), unmix(pg, kg), unmix(pb, kb)}
}
