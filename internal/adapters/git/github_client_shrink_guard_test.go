// FILE: internal/adapters/git/github_client_shrink_guard_test.go
//
// The git-writer shrink floor (bugs_open/198). The defect it guards produced
// nine clobbered sites across three waves with every log line green, so the
// properties worth pinning are the ones whose failure LOOKS LIKE SUCCESS:
//
//   * the guard measures the PREFIXED path (the unprefixed mutation 404s on
//     everything, reads every 404 as "new file", and passes every commit while
//     appearing to work);
//   * a measurement failure REFUSES (a guard that cannot measure has verified
//     nothing) and says so in different words from a real violation;
//   * a caller that did not opt in makes no extra API call at all;
//   * an absurd floor is clamped rather than honoured (a floor > 1 demands every
//     commit grow the file, refusing everything for ever).
//
// Each test names the mutation it fails under.

package git

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"go.uber.org/zap"
)

// shrinkHarness is the refrace harness plus a contents endpoint. sizes maps a
// repo path to its incumbent byte size; a path absent from the map 404s.
type shrinkHarness struct {
	server       *httptest.Server
	client       *GitHubClient
	contentsGets *atomic.Int32
	blobCreates  *atomic.Int32
	refPatches   *atomic.Int32
	contentPaths []string
}

func newShrinkHarness(t *testing.T, sizes map[string]int, contentsStatus int) *shrinkHarness {
	t.Helper()
	h := &shrinkHarness{
		contentsGets: &atomic.Int32{},
		blobCreates:  &atomic.Int32{},
		refPatches:   &atomic.Int32{},
	}

	h.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, method := r.URL.Path, r.Method
		switch {
		case method == "GET" && path == "/repos/testorg/sites":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "sites", "html_url": "https://github.com/testorg/sites",
				"default_branch": "master",
				"owner":          map[string]string{"login": "testorg"},
			})
		case method == "GET" && strings.HasPrefix(path, "/repos/testorg/sites/contents/"):
			h.contentsGets.Add(1)
			filePath := strings.TrimPrefix(path, "/repos/testorg/sites/contents/")
			h.contentPaths = append(h.contentPaths, filePath)
			if contentsStatus != 0 && contentsStatus != http.StatusOK {
				w.WriteHeader(contentsStatus)
				w.Write([]byte(`{"message":"boom"}`))
				return
			}
			size, ok := sizes[filePath]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message":"Not Found"}`))
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": filePath, "size": size, "type": "file",
			})
		case method == "GET" && strings.HasPrefix(path, "/repos/testorg/sites/git/ref/heads/"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"object": map[string]string{"sha": "base-1"},
			})
		case method == "POST" && strings.HasSuffix(path, "/git/blobs"):
			n := h.blobCreates.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"sha": fmt.Sprintf("blob-%d", n)})
		case method == "POST" && strings.HasSuffix(path, "/git/trees"):
			json.NewEncoder(w).Encode(map[string]string{"sha": "tree-1"})
		case method == "POST" && strings.HasSuffix(path, "/git/commits"):
			json.NewEncoder(w).Encode(map[string]string{"sha": "commit-1"})
		case method == "PATCH" && strings.HasPrefix(path, "/repos/testorg/sites/git/refs/heads/"):
			h.refPatches.Add(1)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request: %s %s", method, path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(h.server.Close)

	client, err := NewGitHubClient("test-token", "testorg", h.server.URL, zap.NewNop())
	if err != nil {
		t.Fatalf("NewGitHubClient: %v", err)
	}
	h.client = client
	return h
}

// ── the pure decision ─────────────────────────────────────────────────────────

// Mutation this fails under: any change to the comparison direction, the
// minimum-size scope, or the treatment of an absent incumbent.
func TestEvaluateFileShrink(t *testing.T) {
	const min = 2048
	cases := []struct {
		name        string
		floor       float64
		incumbent   map[string]int
		incoming    map[string]int
		wantRefused int
	}{
		{"floor zero disables", 0, map[string]int{"a": 20000}, map[string]int{"a": 10}, 0},
		{"negative floor disables", -1, map[string]int{"a": 20000}, map[string]int{"a": 10}, 0},
		{"the clobber: 428 over 23650", 0.5, map[string]int{"a": 23650}, map[string]int{"a": 428}, 1},
		{"new file is never a shrink", 0.5, map[string]int{}, map[string]int{"a": 10}, 0},
		{"growth allowed", 0.5, map[string]int{"a": 20000}, map[string]int{"a": 30000}, 0},
		{"identical allowed", 0.5, map[string]int{"a": 20000}, map[string]int{"a": 20000}, 0},
		{"tiny incumbent out of scope", 0.5, map[string]int{"a": 1000}, map[string]int{"a": 10}, 0},
		{"exactly at the floor is allowed", 0.5, map[string]int{"a": 20000}, map[string]int{"a": 10000}, 0},
		{"one byte under the floor refuses", 0.5, map[string]int{"a": 20000}, map[string]int{"a": 9999}, 1},
		{"only the offending path is named", 0.5,
			map[string]int{"a": 20000, "b": 20000},
			map[string]int{"a": 19000, "b": 100}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := evaluateFileShrink(tc.floor, min, tc.incumbent, tc.incoming)
			if len(got) != tc.wantRefused {
				t.Fatalf("refused %d paths, want %d (%+v)", len(got), tc.wantRefused, got)
			}
		})
	}
}

// Mutation this fails under: removing the clamp in clampShrinkFloor. A floor of
// 5.0 then demands every commit be 5× the file it replaces, which refuses every
// deploy for ever — and reads as "deploys are broken", not "the config is wrong".
func TestAbsurdFloorIsClampedNotRefuseEverything(t *testing.T) {
	// Same size in and out: no shrink at all. Under a clamped floor this passes.
	got := evaluateFileShrink(5.0, 2048, map[string]int{"a": 20000}, map[string]int{"a": 20000})
	if len(got) != 0 {
		t.Fatalf("an unshrunk file must pass under any floor once clamped, got %+v", got)
	}
	if f, on := clampShrinkFloor(5.0); !on || f != maxShrinkFloor {
		t.Fatalf("clampShrinkFloor(5.0) = (%v,%v), want (%v,true)", f, on, maxShrinkFloor)
	}
}

// Mutation this fails under: measuring len(content) on a base64 payload instead
// of its decoded length — which overstates the incoming size by ~33% and lets a
// genuine shrink through.
func TestIncomingBytesMeasuresDecodedBase64(t *testing.T) {
	raw := strings.Repeat("x", 300)
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	n, err := incomingFileBytes(map[string]interface{}{"content": encoded, "encoding": "base64"})
	if err != nil {
		t.Fatalf("incomingFileBytes: %v", err)
	}
	if n != len(raw) {
		t.Fatalf("base64 payload measured %d bytes, want the decoded %d", n, len(raw))
	}
	if len(encoded) == len(raw) {
		t.Fatal("test is vacuous: encoded and raw are the same length")
	}
}

// ── the wired guard ───────────────────────────────────────────────────────────

func commitWithFloor(h *shrinkHarness, files map[string]interface{}, floor float64) error {
	_, err := h.client.CommitToRepo(context.Background(), GitCommitData{
		RepoName:        "sites",
		Domain:          "cookly.uk",
		Branch:          "master",
		CommitMessage:   "test",
		Files:           files,
		FileShrinkFloor: floor,
	})
	return err
}

// The incident, reproduced: a 504-byte accumulation of patch rules deployed over
// a 17,462-byte stylesheet. Mutation this fails under: moving the enforcement
// call below the blob loop (the refusal would then leave orphan blobs behind), or
// deleting it from CommitToRepo entirely.
func TestShrinkFloorRefusesTheClobberBeforeAnyBlob(t *testing.T) {
	h := newShrinkHarness(t, map[string]int{"cookly.uk/assets/css/styles.css": 17462}, http.StatusOK)

	err := commitWithFloor(h, map[string]interface{}{
		"assets/css/styles.css": strings.Repeat("x", 504),
	}, 0.5)

	if err == nil {
		t.Fatal("a 504-byte replacement of a 17,462-byte file must be refused")
	}
	if !strings.Contains(err.Error(), shrinkRefusalFix) {
		t.Errorf("refusal must carry the violation sentence, got: %v", err)
	}
	if !strings.Contains(err.Error(), "17462") || !strings.Contains(err.Error(), "504") {
		t.Errorf("refusal must name both sizes so it is checkable, got: %v", err)
	}
	if got := h.blobCreates.Load(); got != 0 {
		t.Errorf("blobs created %d times before the refusal, want 0", got)
	}
	if got := h.refPatches.Load(); got != 0 {
		t.Errorf("ref updated %d times despite refusal, want 0", got)
	}
}

// Mutation this fails under: measuring the unprefixed path. Every contents
// lookup would 404, every 404 reads as "new file", and the guard would pass the
// very commit it exists to refuse — while still logging that it ran.
func TestShrinkFloorMeasuresThePrefixedPath(t *testing.T) {
	h := newShrinkHarness(t, map[string]int{"cookly.uk/assets/css/styles.css": 17462}, http.StatusOK)

	_ = commitWithFloor(h, map[string]interface{}{
		"assets/css/styles.css": strings.Repeat("x", 504),
	}, 0.5)

	if len(h.contentPaths) == 0 {
		t.Fatal("the guard made no contents call at all")
	}
	if h.contentPaths[0] != "cookly.uk/assets/css/styles.css" {
		t.Fatalf("guard measured %q — it must measure the DOMAIN-PREFIXED path that exists in the tree", h.contentPaths[0])
	}
}

// Mutation this fails under: running the guard unconditionally (dropping the
// `data.FileShrinkFloor > 0` condition in CommitToRepo, or the early return in
// enforceFileShrinkFloor). Every one of the 17 carrier agents would then pay an
// extra API call per file per commit, and a contents-API outage would start
// failing deploys that never asked for the guard.
func TestUnconfiguredCallerMakesNoContentsCall(t *testing.T) {
	h := newShrinkHarness(t, map[string]int{"cookly.uk/assets/css/styles.css": 17462}, http.StatusOK)

	if err := commitWithFloor(h, map[string]interface{}{
		"assets/css/styles.css": strings.Repeat("x", 10),
	}, 0); err != nil {
		t.Fatalf("an unconfigured caller must commit exactly as before, got: %v", err)
	}
	if got := h.contentsGets.Load(); got != 0 {
		t.Errorf("contents API called %d times with the guard off, want 0", got)
	}
	if got := h.blobCreates.Load(); got != 1 {
		t.Errorf("blobs created %d times, want 1 — the commit must still happen", got)
	}
}

// Mutation this fails under: treating a non-404 error as "absent" (the shape
// partitionExistingPaths was careful to avoid). A 500 would then read as "new
// file" and the guard would pass everything during an outage — blind, and
// indistinguishable from healthy.
func TestMeasurementFailureRefusesWithItsOwnSentence(t *testing.T) {
	h := newShrinkHarness(t, nil, http.StatusInternalServerError)

	err := commitWithFloor(h, map[string]interface{}{
		"assets/css/styles.css": strings.Repeat("x", 504),
	}, 0.5)

	if err == nil {
		t.Fatal("a guard that cannot measure must refuse, not pass")
	}
	if !strings.Contains(err.Error(), shrinkMeasurementErrorFix) {
		t.Errorf("want the MEASUREMENT sentence, got: %v", err)
	}
	if strings.Contains(err.Error(), shrinkRefusalFix) {
		t.Error("a measurement failure must not tell the operator to lower the floor — nothing shrank")
	}
	if got := h.blobCreates.Load(); got != 0 {
		t.Errorf("blobs created %d times, want 0", got)
	}
}

// A genuinely new file must deploy. Mutation this fails under: treating an
// absent incumbent as size 0 and comparing against it.
func TestNewFileIsAllowed(t *testing.T) {
	h := newShrinkHarness(t, map[string]int{}, http.StatusOK)

	if err := commitWithFloor(h, map[string]interface{}{
		"assets/css/styles.css": strings.Repeat("x", 10),
	}, 0.5); err != nil {
		t.Fatalf("a file that does not exist yet cannot have shrunk, got: %v", err)
	}
	if got := h.blobCreates.Load(); got != 1 {
		t.Errorf("blobs created %d, want 1", got)
	}
}

// The healthy case, which is the one that must never regress: a normal patch
// deploy grows the file and passes.
func TestHealthyDeployPasses(t *testing.T) {
	h := newShrinkHarness(t, map[string]int{"cookly.uk/assets/css/styles.css": 17462}, http.StatusOK)

	if err := commitWithFloor(h, map[string]interface{}{
		"assets/css/styles.css": strings.Repeat("x", 17600),
	}, 0.5); err != nil {
		t.Fatalf("an append that grows the file must pass, got: %v", err)
	}
	if got := h.blobCreates.Load(); got != 1 {
		t.Errorf("blobs created %d, want 1", got)
	}
	if got := h.refPatches.Load(); got != 1 {
		t.Errorf("ref updated %d times, want 1", got)
	}
}

// Whole-commit refusal, not per-file. Mutation this fails under: filtering the
// offending file out and committing the rest, which would leave the repo in a
// state no caller asked for.
func TestOneBadFileRefusesTheWholeCommit(t *testing.T) {
	h := newShrinkHarness(t, map[string]int{
		"cookly.uk/assets/css/styles.css": 17462,
		"cookly.uk/index.html":            20000,
	}, http.StatusOK)

	err := commitWithFloor(h, map[string]interface{}{
		"assets/css/styles.css": strings.Repeat("x", 504),   // the clobber
		"index.html":            strings.Repeat("x", 21000), // perfectly fine
	}, 0.5)

	if err == nil {
		t.Fatal("one refused file must refuse the whole commit")
	}
	if got := h.blobCreates.Load(); got != 0 {
		t.Errorf("blobs created %d, want 0 — nothing may land", got)
	}
}
