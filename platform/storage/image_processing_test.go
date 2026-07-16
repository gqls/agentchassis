// FILE: platform/storage/image_processing_test.go
//
// CoverCropResize (Phase I3): the card derivation needs an EXACT-size,
// centre-cropped output regardless of the source aspect. Geometry only —
// resampling quality is nfnt/resize's business.

package storage

import (
	"image"
	"image/color"
	"testing"
)

// gradientImage builds a WxH image whose left half is red and right half is
// blue, top-left origin — enough structure to verify centre-cropping keeps
// the middle rather than an edge.
func gradientImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x < w/2 {
				img.Set(x, y, color.RGBA{R: 255, A: 255})
			} else {
				img.Set(x, y, color.RGBA{B: 255, A: 255})
			}
		}
	}
	return img
}

func TestCoverCropResize_ExactOutputSize(t *testing.T) {
	cases := []struct {
		name             string
		srcW, srcH       int
		targetW, targetH uint
	}{
		{"same aspect pure downscale (hero→card)", 1600, 900, 800, 450},
		{"wider than target (crops sides)", 2400, 900, 800, 450},
		{"taller than target (crops top/bottom)", 900, 1600, 800, 450},
		{"square source to 16:9", 768, 768, 800, 450},
		{"upscale smaller source", 400, 225, 800, 450},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := CoverCropResize(gradientImage(tc.srcW, tc.srcH), tc.targetW, tc.targetH)
			b := out.Bounds()
			if b.Dx() != int(tc.targetW) || b.Dy() != int(tc.targetH) {
				t.Fatalf("output %dx%d, want %dx%d", b.Dx(), b.Dy(), tc.targetW, tc.targetH)
			}
		})
	}
}

func TestCoverCropResize_CentreCropKeepsMiddle(t *testing.T) {
	// A 3000×450 source into 800×450: the crop must come from the horizontal
	// middle, so the output's left half is still red and right half blue —
	// an edge-anchored crop would be single-colour.
	out := CoverCropResize(gradientImage(3000, 450), 800, 450)
	r, _, _, _ := out.At(10, 225).RGBA()
	_, _, bl, _ := out.At(790, 225).RGBA()
	if r == 0 {
		t.Fatalf("left edge lost the red half — crop not centred")
	}
	if bl == 0 {
		t.Fatalf("right edge lost the blue half — crop not centred")
	}
}

func TestCoverCropResize_ZeroTargetIsNoop(t *testing.T) {
	src := gradientImage(100, 100)
	if out := CoverCropResize(src, 0, 450); out != src {
		t.Fatalf("zero target width should return the source unchanged")
	}
}

func TestCardPurposeRegistered(t *testing.T) {
	w, h, q, ext := GetImageConfig("card")
	if w != 800 || h != 450 || ext != "jpg" {
		t.Fatalf("card purpose = %dx%d %s, want 800x450 jpg", w, h, ext)
	}
	if q < 75 || q > 90 {
		t.Fatalf("card quality %d outside sane JPG range", q)
	}
	if got := DeployedWebPath("card_learning_center_hub", "card"); got != "/assets/images/card-learning-center-hub.jpg" {
		t.Fatalf("DeployedWebPath card = %q", got)
	}
}
