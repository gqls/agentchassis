// FILE: internal/adapters/git/github_client_refrace_test.go
//
// CommitToRepo's ref-race retry (bugs_open/120, owner ruling 2026-07-28) has
// three properties a diff reader cannot see hold together: only a
// non-fast-forward re-bases (everything else must keep failing loudly), the
// base is RE-READ inside the retry loop (retrying on the stale head would
// loop on the same 422 forever), and blobs are created outside it (they are
// content-addressed; re-creating them per attempt would triple API traffic
// under contention). The first is a unit test against the production error
// text; the structural pair are source-reading tripwires in the
// state_locks_test.go style.

package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"go.uber.org/zap"
)

// TestRefRaceRetriesRebaseOnNewHead drives CommitToRepo through the FAILING
// branch against a fake GitHub: the first updateRef gets the verbatim 422 and
// the branch head moves (the concurrent winner); the retry must re-read the
// NEW head, rebuild tree and commit on it, and succeed — with the blobs
// created exactly once. This is the branch production has not yet exercised
// (deployed 2026-07-28; two 5-burst runs completed 16/16 without a natural
// race), so the induced proof lives here.
func TestRefRaceRetriesRebaseOnNewHead(t *testing.T) {
	var (
		headSHA     atomic.Value // moves when the "winner" lands
		refReads    atomic.Int32
		blobCreates atomic.Int32
		refPatches  atomic.Int32
		treeBases   []string // base_tree sent to createTree, in order
		commitBases []string // parent sent to createCommit, in order
	)
	headSHA.Store("base-1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, method := r.URL.Path, r.Method
		switch {
		case method == "GET" && path == "/repos/testorg/sites":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "sites", "html_url": "https://github.com/testorg/sites",
				"default_branch": "master",
				"owner":          map[string]string{"login": "testorg"},
			})
		case method == "GET" && strings.HasPrefix(path, "/repos/testorg/sites/git/ref/heads/"):
			refReads.Add(1)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"object": map[string]string{"sha": headSHA.Load().(string)},
			})
		case method == "POST" && strings.HasSuffix(path, "/git/blobs"):
			n := blobCreates.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"sha": fmt.Sprintf("blob-%d", n)})
		case method == "POST" && strings.HasSuffix(path, "/git/trees"):
			var body struct {
				BaseTree string `json:"base_tree"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			treeBases = append(treeBases, body.BaseTree)
			json.NewEncoder(w).Encode(map[string]string{"sha": fmt.Sprintf("tree-on-%s", body.BaseTree)})
		case method == "POST" && strings.HasSuffix(path, "/git/commits"):
			var body struct {
				Parents []string `json:"parents"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			if len(body.Parents) > 0 {
				commitBases = append(commitBases, body.Parents[0])
			}
			json.NewEncoder(w).Encode(map[string]string{"sha": "commit-x"})
		case method == "PATCH" && strings.HasPrefix(path, "/repos/testorg/sites/git/refs/heads/"):
			if refPatches.Add(1) == 1 {
				// The concurrent winner lands between our read and our update.
				headSHA.Store("base-2")
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte(`{"message":"Update is not a fast forward","status":"422"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request: %s %s", method, path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewGitHubClient("test-token", "testorg", server.URL, zap.NewNop())
	if err != nil {
		t.Fatalf("NewGitHubClient: %v", err)
	}

	htmlURL, err := client.CommitToRepo(context.Background(), GitCommitData{
		RepoName:      "sites",
		Branch:        "master",
		CommitMessage: "test rerender",
		Files:         map[string]interface{}{"index.html": "<html/>"},
	})
	if err != nil {
		t.Fatalf("CommitToRepo should succeed after re-basing, got: %v", err)
	}
	if htmlURL != "https://github.com/testorg/sites" {
		t.Errorf("unexpected html url %q", htmlURL)
	}

	if got := refPatches.Load(); got != 2 {
		t.Errorf("updateRef called %d times, want 2 (fail, then succeed)", got)
	}
	if got := refReads.Load(); got != 2 {
		t.Errorf("branch head read %d times, want 2 — the retry must RE-READ the head", got)
	}
	if got := blobCreates.Load(); got != 1 {
		t.Errorf("blobs created %d times, want exactly 1 — content-addressed, hoisted above the loop", got)
	}
	if len(treeBases) != 2 || treeBases[0] != "base-1" || treeBases[1] != "base-2" {
		t.Errorf("tree bases %v, want [base-1 base-2] — the second attempt must build on the WINNER'S head", treeBases)
	}
	if len(commitBases) != 2 || commitBases[1] != "base-2" {
		t.Errorf("commit parents %v, want second parent base-2", commitBases)
	}
}

func TestIsNonFastForwardMatchesProductionError(t *testing.T) {
	// Verbatim shape from the live failure this fix exists for
	// (orchestration 6aedced7…, 2026-07-28 11:26:28Z).
	live := errors.New(`github API request failed with status: 422 Unprocessable Entity - {"message":"Update is not a fast forward","documentation_url":"https://docs.github.com/rest/git/refs#update-a-reference","status":"422"}`)
	if !isNonFastForward(live) {
		t.Fatal("the production non-fast-forward error is not recognised — the retry would never fire")
	}

	for _, e := range []error{
		errors.New("github API request failed with status: 401 Unauthorized"),
		errors.New("github API request failed with status: 403 rate limit exceeded"),
		errors.New(`github API request failed with status: 422 Unprocessable Entity - {"message":"Reference already exists"}`),
		errors.New("context deadline exceeded"),
		nil,
	} {
		if isNonFastForward(e) {
			t.Errorf("non-retryable error classified as a ref race: %v", e)
		}
	}
}

func commitToRepoBody(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile("github_client.go")
	if err != nil {
		t.Fatalf("read github_client.go: %v", err)
	}
	body := string(src)
	i := strings.Index(body, "func (c *GitHubClient) CommitToRepo")
	if i < 0 {
		t.Fatal("CommitToRepo not found — renamed? update this test with it")
	}
	body = body[i:]
	if j := strings.Index(body, "\nfunc "); j >= 0 {
		body = body[:j]
	}
	return body
}

func TestBaseIsReReadInsideTheRetryLoop(t *testing.T) {
	body := commitToRepoBody(t)
	loop := strings.Index(body, "for attempt := 1")
	if loop < 0 {
		t.Fatal("the ref-race retry loop is gone from CommitToRepo")
	}
	inLoop := body[loop:]
	if !strings.Contains(inLoop, "getLatestCommitSHA") {
		t.Fatal("getLatestCommitSHA is no longer read INSIDE the retry loop — " +
			"retrying on a stale base loops on the same non-fast-forward forever")
	}
}

func TestBlobsAreCreatedOutsideTheRetryLoop(t *testing.T) {
	body := commitToRepoBody(t)
	loop := strings.Index(body, "for attempt := 1")
	blob := strings.Index(body, "createBlob")
	if blob < 0 || loop < 0 {
		t.Fatal("createBlob or the retry loop not found in CommitToRepo")
	}
	if blob > loop {
		t.Fatal("blob creation moved inside the retry loop — blobs are content-addressed and must be created once")
	}
}
