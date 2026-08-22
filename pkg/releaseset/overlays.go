// FILE: pkg/releaseset/overlays.go
//
// The FILESYSTEM half of the enumeration: what the deployment tree says each
// service runs. Kept separate from decl.go because the two halves fail
// differently — a makefile that cannot be parsed is "could not run", an empty
// overlay walk is "nothing to check", and conflating them is how a gate reports
// clean on a tree it never opened.
package releaseset

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Pin is one overlay's statement of what image a service runs.
type Pin struct {
	Service string // directory name under services/
	Image   string // fully qualified, e.g. docker.io/aqls/agent-chassis
	Tag     string // newTag, may be empty
	Path    string // the kustomization.yaml, for error messages
}

// Bare returns the image without its registry prefix, or "" when the image is
// not from the registry given.
func (p Pin) Bare(registry string) string {
	prefix := strings.TrimSuffix(registry, "/") + "/"
	if !strings.HasPrefix(p.Image, prefix) {
		return ""
	}
	return strings.TrimPrefix(p.Image, prefix)
}

// ScanOverlays walks every production overlay under root and reads EVERY image
// each one pins — one Pin per element of the `images:` block, not one per file.
//
// ⚠ IT WALKS TO ANY DEPTH, and that is a correction, not a generalisation.
// The shell gate globbed a fixed `overlays/$(OVERLAY_PATH)/kustomization.yaml`
// = `overlays/production/uk_001/...`. Measured 2026-08-22: `tools-api` has a
// production overlay at `overlays/production/kustomization.yaml` with NO region
// directory, which that glob can never see. It pins a placeholder today and has
// no workload, so nothing is broken — but a real service placed at that depth
// would have been invisible to the gate for exactly the reason this bug exists.
// The enumeration must not bake in the region depth.
//
// ⚠ IT PREFERS newName OVER name, which is kustomize's own semantics: `name`
// selects the image to rewrite, `newName` is what actually runs. The shell gate
// read `name` and would therefore have reported the ORIGINAL image for any
// overlay that redirected. Only one overlay on this estate uses newName today
// (`tools-api`, a placeholder), so this is a correctness fix for a LATENT case
// — stating it as a live defect would be an overclaim.
//
// ⚠ AND IT DOES NOT STOP AT THE FIRST IMAGE — see blockImages below. Both of
// these are cases where the honest description of the fix is "the old gate could
// not have seen this", not "the old gate was wrong about this".
func ScanOverlays(root string) ([]Pin, error) {
	servicesDir := filepath.Join(root, "deployments", "kustomize", "services")
	info, err := os.Stat(servicesDir)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot read %s: %w — refusing to report zero overlays when the tree was "+
				"never opened (bugs_open/318)", servicesDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", servicesDir)
	}

	var pins []Pin
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", servicesDir, err)
	}
	for _, svcDir := range entries {
		if !svcDir.IsDir() {
			continue
		}
		prodRoot := filepath.Join(servicesDir, svcDir.Name(), "overlays", "production")
		if _, err := os.Stat(prodRoot); err != nil {
			// No production overlay at all. Base-only services applied by hand
			// live here (site-discovery-staleness-check, site-locale-unset-check
			// as of 2026-08-22); they run upstream images and are outside this
			// gate's reach BY CONSTRUCTION. Named in the report header rather
			// than silently skipped.
			continue
		}
		walkErr := filepath.WalkDir(prodRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || d.Name() != "kustomization.yaml" {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			for _, im := range blockImages(string(raw)) {
				pins = append(pins, Pin{Service: svcDir.Name(), Image: im.image, Tag: im.tag, Path: rel})
			}
			return nil
		})
		if walkErr != nil {
			return nil, walkErr
		}
	}
	sort.Slice(pins, func(i, j int) bool {
		if pins[i].Service != pins[j].Service {
			return pins[i].Service < pins[j].Service
		}
		return pins[i].Path < pins[j].Path
	})
	return pins, nil
}

// blockImages reads EVERY element of the `images:` transformer block and
// returns, for each, the image that element actually selects — `newName` when
// present, else `name` — plus its `newTag`.
//
// ⚠ IT READS EVERY ELEMENT, AND THAT IS A COUNCIL FINDING ACTED ON, not a
// generalisation. The first cut stopped at the first element, inheriting the
// shell gate's `awk … exit`. The council's `bug_historian` seat objected
// (medium, corr 83442a5a): if an overlay pins two images — app plus sidecar, or
// two of our services in one kustomization — and OURS is the second, the scan
// would silently produce no Pin for it and the gate could never flag it. That is
// precisely the shape this whole change exists to close (a check reporting clean
// about the thing it never looked at), reproduced in miniature inside the fix.
// [MEASURED 2026-08-22: no kustomization anywhere under deployments/ has more
// than one element today, so the gap was LATENT, not live — but a cap that is
// only safe because of today's data is the wrong kind of safe, and a warning
// surface would have been a worse answer than simply not capping.]
//
// Two boundaries remain and each has a test:
//
//   - the block ENDS at the first line that is neither indented nor blank, so a
//     `patches:` or `labels:` section below cannot leak into the read;
//   - within an element `newName` WINS over `name` regardless of source order,
//     because that is kustomize's semantics: `name` selects what to rewrite,
//     `newName` is what runs.
type blockImage struct{ image, tag string }

func blockImages(text string) []blockImage {
	var out []blockImage
	var cur blockImage
	inImages, inElement := false, false

	flush := func() {
		if !inElement {
			return
		}
		img := cur.image
		if img != "" {
			out = append(out, blockImage{image: img, tag: cur.tag})
		}
		cur = blockImage{}
	}

	var name, newName string
	setImage := func() {
		if newName != "" {
			cur.image = newName
		} else {
			cur.image = name
		}
	}

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !inImages {
			if trimmed == "images:" {
				inImages = true
			}
			continue
		}
		// End of the images block: a non-blank line with no leading whitespace.
		if trimmed != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}
		if strings.HasPrefix(trimmed, "- ") || trimmed == "-" {
			setImage()
			flush()
			name, newName = "", ""
			inElement = true
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
		}
		if !inElement || trimmed == "" {
			continue
		}
		switch {
		case strings.HasPrefix(trimmed, "newName:"):
			newName = value(trimmed)
		case strings.HasPrefix(trimmed, "name:"):
			name = value(trimmed)
		case strings.HasPrefix(trimmed, "newTag:"):
			cur.tag = value(trimmed)
		}
	}
	setImage()
	flush()
	return out
}

func value(kv string) string {
	_, v, _ := strings.Cut(kv, ":")
	return strings.TrimSpace(strings.Trim(strings.TrimSpace(v), `"'`))
}
