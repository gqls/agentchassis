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
	"errors"
	"os"
	"strings"
	"testing"
)

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
