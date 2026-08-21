package git

// The git-writer shrink floor — bugs_open/198.
//
// WHY THIS LIVES HERE AND NOT IN THE CHASSIS. The defect it guards is a commit
// that replaces a live site asset with a fraction of itself: css-patch-agent
// deployed a ~100-byte accumulation of patch rules over 17-26KB stylesheets on
// nine sites across three waves, every run reporting success. The chassis-side
// caller cannot see that, because it has no idea what the file it is replacing
// currently contains — GitCommitAction assembles a payload and produces it to
// Kafka. This client is the only place in the estate that holds BOTH the new
// bytes and a cheap way to ask what the old ones were, so it is the only place
// the comparison can be made at all.
//
// WHY IT IS OPT-IN, DEFAULT OFF. `git_commit` is a shared seam with 17 carrier
// agents. New authority on a shared seam ships as an opt-in field whose unsafe
// side is the default (owner ruling 2026-08-02 §2), so that the decision is
// visible to a reviewer of the CALLER rather than buried in a helper every
// caller inherits. FileShrinkFloor == 0 means every existing caller behaves
// byte-identically and makes no extra API call — asserted by test, not by
// inspection.
//
// WHY THIS PLACEMENT IS NOT THE ONE LANDMINES WARNS ABOUT. That entry ("the
// owned-page guard belongs on assemble_page, NOT on git_commit") is about a
// guard that refused a CLASS OF CALLER: it would have stopped tool pages
// deploying at all, and the failure would have looked like "tools mysteriously
// stopped updating". This guard refuses a SHAPE OF PAYLOAD, only for a step that
// asked for it, and says exactly what it refused and why. That is the
// archived_page_guard argument — a prohibition belongs at the seam every
// producer passes through — restricted further by the opt-in.
//
// FAIL CLOSED, WITH A DIFFERENT SENTENCE. A guard that cannot measure has not
// observed a healthy commit; it has observed nothing. So a measurement failure
// refuses too, and says so in its own words: "lower the floor" is not a remedy
// for a contents API that returned 500, and sharing one sentence between the two
// cases is how an operator ends up disabling a working guard to fix a network
// fault (the split the council forced on the 293 guard, for the same reason).

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"sort"
	"strings"

	"go.uber.org/zap"
)

const (
	// minShrinkGuardFileBytes is the smallest INCUMBENT file the floor judges.
	// Below it a ratio is noise — a 200-byte file dropping to 90 says nothing
	// about intent, and the artefacts this exists to protect (site stylesheets,
	// scripts, feeds) are all orders of magnitude larger. Distinct from, and
	// deliberately smaller than, migration 542's 4096-byte "is this a
	// stylesheet at all" floor: that one asks whether a css_themes row is a
	// plausible stylesheet, this one asks whether a file is big enough for a
	// ratio to mean anything.
	minShrinkGuardFileBytes = 2048

	// maxShrinkFloor clamps an absurd configured floor instead of honouring it.
	// A floor above 1.0 demands that every commit GROW the file, which refuses
	// every write for ever and reads as "the deploy is broken" rather than "the
	// config is wrong" — the exact trap a missing clamp set for the page-total
	// text floor (found by using it, commit d5b40c4eb).
	maxShrinkFloor = 0.95
)

// fileShrink is one refused path, carrying the numbers that make the refusal
// checkable rather than merely assertive.
type fileShrink struct {
	Path     string
	OldBytes int
	NewBytes int
	Ratio    float64
}

// clampShrinkFloor normalises a configured floor. Returns (floor, enabled).
func clampShrinkFloor(floor float64) (float64, bool) {
	if floor <= 0 {
		return 0, false
	}
	if floor > maxShrinkFloor {
		return maxShrinkFloor, true
	}
	return floor, true
}

// evaluateFileShrink is the whole decision, with no I/O in it so it can be
// driven by a table test. incumbent holds the measured size of every path that
// already exists (absent paths are simply not in the map — a new file is never a
// shrink); incoming holds the byte length of what is about to be written.
func evaluateFileShrink(floor float64, minIncumbentBytes int, incumbent, incoming map[string]int) []fileShrink {
	f, enabled := clampShrinkFloor(floor)
	if !enabled {
		return nil
	}

	var refused []fileShrink
	for path, newBytes := range incoming {
		oldBytes, exists := incumbent[path]
		if !exists {
			continue // a file that did not exist cannot have shrunk
		}
		if oldBytes < minIncumbentBytes {
			continue // too small for a ratio to carry meaning
		}
		if float64(newBytes) >= f*float64(oldBytes) {
			continue // within the floor, or growing
		}
		refused = append(refused, fileShrink{
			Path:     path,
			OldBytes: oldBytes,
			NewBytes: newBytes,
			Ratio:    float64(newBytes) / float64(oldBytes),
		})
	}
	// Deterministic order: the refusal message is read by humans and compared by
	// tests, and map iteration order would make both flaky.
	sort.Slice(refused, func(i, j int) bool { return refused[i].Path < refused[j].Path })
	return refused
}

// incomingFileBytes measures what will actually land in the blob, which for a
// base64 payload is the DECODED length — the same convention sha256OfFiles
// documents and for the same reason: the bytes in the repo are what a later
// reader sees, not their transport encoding.
func incomingFileBytes(fileData interface{}) (int, error) {
	var content, encoding string

	switch v := fileData.(type) {
	case string:
		content, encoding = v, "utf-8"
	case map[string]interface{}:
		if s, ok := v["content"].(string); ok {
			content = s
		}
		if e, ok := v["encoding"].(string); ok {
			encoding = e
		}
		if encoding == "" {
			encoding = "utf-8"
		}
	default:
		return 0, fmt.Errorf("invalid file data type: %T", fileData)
	}

	if strings.EqualFold(encoding, "base64") {
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return 0, fmt.Errorf("base64 payload could not be decoded for measurement: %w", err)
		}
		return len(decoded), nil
	}
	return len(content), nil
}

// pathSize reports the size in bytes of ref:path. Same call pathExists makes —
// the contents API already returns `size` and that function discards it — with
// 404 the only status meaning "absent" and every other non-2xx an error, so a
// blinded guard can never be mistaken for a passing one.
//
// The contents endpoint returns a JSON ARRAY for a directory and refuses files
// over ~1MB with a non-2xx; both surface here as an error, which fails closed.
func (c *GitHubClient) pathSize(ctx context.Context, owner, repo, ref, path string) (int, bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s",
		c.apiBase, owner, repo, pathEscapeSegments(path), neturl.QueryEscape(ref))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		_, _ = io.Copy(io.Discard, resp.Body)
		return 0, false, nil
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		var payload struct {
			Size int    `json:"size"`
			Type string `json:"type"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return 0, false, fmt.Errorf("github contents API response for %q could not be read as a file: %w", path, err)
		}
		if payload.Type != "" && payload.Type != "file" {
			return 0, false, fmt.Errorf("github contents API returned a %q, not a file, for %q", payload.Type, path)
		}
		return payload.Size, true, nil
	default:
		_, _ = io.Copy(io.Discard, resp.Body)
		return 0, false, fmt.Errorf("github contents API returned %s for %q", resp.Status, path)
	}
}

// enforceFileShrinkFloor refuses the WHOLE commit when any opted-in file would
// replace an existing file with less than floor × its bytes. Whole-commit and
// not per-file: the files in one request are one intended state, and letting the
// unrefused half land would leave the repo in a state nobody asked for.
//
// Called with paths ALREADY domain-prefixed, because that is what is in the tree
// and therefore what the contents API can be asked about. Measuring the
// unprefixed path is the dangerous mutation — every lookup 404s, every 404 reads
// as "new file", and the guard passes everything while looking like it works.
// There is a test whose only job is to fail under it.
func (c *GitHubClient) enforceFileShrinkFloor(ctx context.Context, owner, repo, branch string, files map[string]interface{}, floor float64) error {
	if _, enabled := clampShrinkFloor(floor); !enabled {
		return nil
	}

	incoming := make(map[string]int, len(files))
	incumbent := make(map[string]int, len(files))

	for path, fileData := range files {
		newBytes, err := incomingFileBytes(fileData)
		if err != nil {
			return fmt.Errorf("%s: %w", shrinkMeasurementErrorFix, err)
		}
		incoming[path] = newBytes

		size, exists, err := c.pathSize(ctx, owner, repo, branch, path)
		if err != nil {
			return fmt.Errorf("%s: could not measure %q at %s/%s@%s: %w",
				shrinkMeasurementErrorFix, path, owner, repo, branch, err)
		}
		if exists {
			incumbent[path] = size
		}
	}

	refused := evaluateFileShrink(floor, minShrinkGuardFileBytes, incumbent, incoming)
	if len(refused) == 0 {
		// The allowed path says so once, with the numbers. This line is the
		// post-roll verification surface: seeing it on a real deploy proves the
		// field arrived AND the guard measured AND a healthy commit passes —
		// three things a refusal alone would not establish.
		c.log.Info("file_shrink_floor: commit passed the shrink floor",
			zap.String("repo", repo),
			zap.String("branch", branch),
			zap.Float64("floor", floor),
			zap.Int("files_measured", len(incoming)),
			zap.Int("files_pre_existing", len(incumbent)))
		return nil
	}

	parts := make([]string, 0, len(refused))
	for _, r := range refused {
		parts = append(parts, fmt.Sprintf("%s: %d bytes would replace %d (%.1f%% of it)",
			r.Path, r.NewBytes, r.OldBytes, r.Ratio*100))
	}
	c.log.Warn("file_shrink_floor: commit REFUSED",
		zap.String("repo", repo),
		zap.String("branch", branch),
		zap.Float64("floor", floor),
		zap.Strings("refused", parts))

	return fmt.Errorf("%s (floor %.2f): %s", shrinkRefusalFix, floor, strings.Join(parts, "; "))
}

// The two aftermath sentences, deliberately NOT shared. Each names the remedy
// that is actually available in its own case; an operator reading the wrong one
// takes the wrong action, and the likeliest wrong action — disabling the guard —
// is unrecoverable in the sense that matters, because the next clobber lands
// while it is off.
const (
	shrinkRefusalFix = "git commit refused: a file would be replaced by a fraction of itself " +
		"(bugs_open/198). If this shrink is intended — a genuine redesign or minification — lower or " +
		"remove `file_shrink_floor` on the committing step for this run. If it is NOT intended, the " +
		"content being committed is probably built on an empty or stale base: check the source it was " +
		"assembled from before committing again"

	shrinkMeasurementErrorFix = "git commit refused: the shrink floor could not MEASURE the file it " +
		"was asked to protect (bugs_open/198), so nothing was verified and the commit was not made. " +
		"Nothing shrank here — lowering `file_shrink_floor` is not the remedy. Retry once the contents " +
		"API is reachable; a persistent failure on one path is a repo/permission fault, not a size fault"
)
