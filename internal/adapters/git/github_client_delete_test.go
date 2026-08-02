// FILE: internal/adapters/git/github_client_delete_test.go
//
// The UNPUBLISH primitive (bugs_open/098). Four properties that a diff reader
// cannot check by eye, each proved against a fake GitHub by driving the real
// client:
//
//  1. a deletion is sent as a tree entry whose `sha` is JSON **null** — not "",
//     not an absent key. The struct field is a *string precisely for this, and
//     nothing else in the codebase would notice if someone "tidied" it back;
//  2. a path that is already absent is a SUCCESS and produces **no commit at
//     all** — a re-run of a repair must be safe, and an empty commit would fire
//     the deploy workflow and purge a Cloudflare zone for no change;
//  3. deletions are resolved INSIDE the ref-race retry, unlike blobs, so a
//     re-base re-asks what the new head holds;
//  4. a traversal path is REFUSED before any request is made.
//
// (2) and (4) are negative assertions, so they are made by MUTATING the input
// until the guard is the only thing that could have produced the outcome — a
// count of "commits that did not happen" proves nothing on its own.
package git

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"go.uber.org/zap"
)

// deleteStub is a fake GitHub that records what the client actually sent, and
// serves a repo whose only existing file is `existing`.
type deleteStub struct {
	server      *httptest.Server
	treeEntries [][]map[string]interface{} // raw tree payload per createTree call
	commits     atomic.Int32
	refPatches  atomic.Int32
	contentsGET atomic.Int32
	blobCreates atomic.Int32
	headSHA     atomic.Value
	// exists is consulted by the contents endpoint; the KEY is the full
	// repo-relative path as GitHub would see it.
	exists map[string]bool
	// onContents, if set, runs before each contents lookup — used to move the
	// world underneath the client mid-retry.
	onContents func()
}

func newDeleteStub(t *testing.T, exists map[string]bool) *deleteStub {
	t.Helper()
	s := &deleteStub{exists: exists}
	s.headSHA.Store("base-1")

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, method := r.URL.Path, r.Method
		switch {
		case method == "GET" && path == "/repos/testorg/sites":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "sites", "html_url": "https://github.com/testorg/sites",
				"default_branch": "master",
				"owner":          map[string]string{"login": "testorg"},
			})

		case method == "GET" && strings.HasPrefix(path, "/repos/testorg/sites/contents/"):
			s.contentsGET.Add(1)
			if s.onContents != nil {
				s.onContents()
			}
			p := strings.TrimPrefix(path, "/repos/testorg/sites/contents/")
			if s.exists[p] {
				json.NewEncoder(w).Encode(map[string]string{"path": p, "sha": "blob-existing"})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Not Found"}`))

		case method == "GET" && strings.HasPrefix(path, "/repos/testorg/sites/git/ref/heads/"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"object": map[string]string{"sha": s.headSHA.Load().(string)},
			})

		case method == "POST" && strings.HasSuffix(path, "/git/blobs"):
			n := s.blobCreates.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"sha": fmt.Sprintf("blob-%d", n)})

		case method == "POST" && strings.HasSuffix(path, "/git/trees"):
			// Decode the tree as RAW maps, not into []TreeEntry — decoding into
			// the struct would happily turn a missing key into a zero value and
			// hide the exact distinction (null vs "" vs absent) under test.
			var body struct {
				Tree []map[string]interface{} `json:"tree"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			s.treeEntries = append(s.treeEntries, body.Tree)
			json.NewEncoder(w).Encode(map[string]string{"sha": "tree-1"})

		case method == "POST" && strings.HasSuffix(path, "/git/commits"):
			s.commits.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"sha": "commit-1"})

		case method == "PATCH" && strings.HasPrefix(path, "/repos/testorg/sites/git/refs/heads/"):
			s.refPatches.Add(1)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))

		default:
			t.Errorf("unexpected request: %s %s", method, path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.server.Close)
	return s
}

func (s *deleteStub) client(t *testing.T) *GitHubClient {
	t.Helper()
	c, err := NewGitHubClient("test-token", "testorg", s.server.URL, zap.NewNop())
	if err != nil {
		t.Fatalf("NewGitHubClient: %v", err)
	}
	return c
}

// TestDeletionIsSentAsNullSHA is the load-bearing one. GitHub removes a path
// when the tree entry's sha is null; "" is rejected and an absent key is not a
// deletion at all. Asserted on the wire, because that is the only place the
// difference is visible.
func TestDeletionIsSentAsNullSHA(t *testing.T) {
	stub := newDeleteStub(t, map[string]bool{
		"robot-hands.com/learning-center/index.html": true,
	})

	_, err := stub.client(t).CommitToRepo(context.Background(), GitCommitData{
		RepoName:      "sites",
		Domain:        "robot-hands.com",
		Deletions:     []string{"learning-center/index.html"},
		CommitMessage: "retract",
	})
	if err != nil {
		t.Fatalf("CommitToRepo: %v", err)
	}

	if len(stub.treeEntries) != 1 {
		t.Fatalf("createTree called %d times, want 1", len(stub.treeEntries))
	}
	tree := stub.treeEntries[0]
	if len(tree) != 1 {
		t.Fatalf("tree has %d entries, want 1: %+v", len(tree), tree)
	}
	entry := tree[0]

	if got := entry["path"]; got != "robot-hands.com/learning-center/index.html" {
		t.Errorf("deletion path %q — the domain prefix must come from the same code the publish uses", got)
	}
	sha, present := entry["sha"]
	if !present {
		t.Fatalf(`the "sha" key is ABSENT from the deletion entry; GitHub needs it present and null`)
	}
	if sha != nil {
		t.Fatalf("sha = %#v, want nil (JSON null). An empty string is rejected by the API and is what a non-pointer field would send", sha)
	}
	if got := entry["mode"]; got != "100644" {
		t.Errorf("mode = %v, want 100644", got)
	}
}

// TestWriteAndDeleteTravelInOneCommit — a MOVE is one commit, which is the
// whole reason deletion lives in CommitToRepo rather than in a verb of its own.
// If these ever split into two commits, a page is momentarily absent from both
// the old path and the new one, which is bugs_closed/125's orphan class.
func TestWriteAndDeleteTravelInOneCommit(t *testing.T) {
	stub := newDeleteStub(t, map[string]bool{"site.com/old.html": true})

	_, err := stub.client(t).CommitToRepo(context.Background(), GitCommitData{
		RepoName:      "sites",
		Domain:        "site.com",
		Files:         map[string]interface{}{"new.html": "<html/>"},
		Deletions:     []string{"old.html"},
		CommitMessage: "move",
	})
	if err != nil {
		t.Fatalf("CommitToRepo: %v", err)
	}

	if got := stub.commits.Load(); got != 1 {
		t.Fatalf("createCommit called %d times, want exactly 1 — a move must be atomic", got)
	}
	tree := stub.treeEntries[0]
	if len(tree) != 2 {
		t.Fatalf("tree has %d entries, want 2 (one write, one removal): %+v", len(tree), tree)
	}
	var writes, removals int
	for _, e := range tree {
		if e["sha"] == nil {
			removals++
		} else {
			writes++
		}
	}
	if writes != 1 || removals != 1 {
		t.Errorf("tree carries %d writes / %d removals, want 1 / 1", writes, removals)
	}
}

// TestAbsentPathIsANoOpNotACommit proves the idempotency guard by MUTATION: the
// only difference from TestDeletionIsSentAsNullSHA is that the file is not
// there. If the guard were removed, this would commit — so "0 commits" here is
// attributable to the guard and to nothing else.
func TestAbsentPathIsANoOpNotACommit(t *testing.T) {
	stub := newDeleteStub(t, map[string]bool{ /* nothing exists */ })

	_, err := stub.client(t).CommitToRepo(context.Background(), GitCommitData{
		RepoName:      "sites",
		Domain:        "robot-hands.com",
		Deletions:     []string{"learning-center/index.html"},
		CommitMessage: "retract again",
	})
	if err != nil {
		t.Fatalf("re-running a retraction must SUCCEED, got: %v", err)
	}

	if got := stub.contentsGET.Load(); got != 1 {
		t.Errorf("existence checked %d times, want 1 — the guard must actually look", got)
	}
	if got := len(stub.treeEntries); got != 0 {
		t.Errorf("createTree called %d times, want 0", got)
	}
	if got := stub.commits.Load(); got != 0 {
		t.Errorf("createCommit called %d times, want 0 — an empty commit still fires the deploy workflow and purges a CDN zone", got)
	}
	if got := stub.refPatches.Load(); got != 0 {
		t.Errorf("updateRef called %d times, want 0", got)
	}
}

// TestPartialAbsenceStillCommitsThePresentOnes — the mixed batch. The absent
// path must not take the present one down with it.
func TestPartialAbsenceStillCommitsThePresentOnes(t *testing.T) {
	stub := newDeleteStub(t, map[string]bool{"site.com/here.html": true})

	_, err := stub.client(t).CommitToRepo(context.Background(), GitCommitData{
		RepoName:      "sites",
		Domain:        "site.com",
		Deletions:     []string{"here.html", "gone.html"},
		CommitMessage: "retract two",
	})
	if err != nil {
		t.Fatalf("CommitToRepo: %v", err)
	}
	if got := stub.commits.Load(); got != 1 {
		t.Fatalf("createCommit called %d times, want 1", got)
	}
	tree := stub.treeEntries[0]
	if len(tree) != 1 || tree[0]["path"] != "site.com/here.html" {
		t.Errorf("tree = %+v, want exactly the ONE path that exists", tree)
	}
}

// TestDeletionsAreResolvedInsideTheRetryLoop. Blobs are content-addressed and
// hoisted above the loop; a removal is a statement ABOUT the head, so it must be
// re-asked after a re-base. Driven by making the FIRST updateRef fail with the
// production 422 and deleting the file underneath the client at the same moment:
// the retry must notice it is gone and produce no removal entry.
func TestDeletionsAreResolvedInsideTheRetryLoop(t *testing.T) {
	stub := newDeleteStub(t, map[string]bool{"site.com/doomed.html": true})

	var patches atomic.Int32
	stub.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" && strings.HasPrefix(r.URL.Path, "/repos/testorg/sites/git/refs/heads/") {
			if patches.Add(1) == 1 {
				// A concurrent committer wins the ref race AND happens to have
				// removed the same file.
				delete(stub.exists, "site.com/doomed.html")
				stub.headSHA.Store("base-2")
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte(`{"message":"Update is not a fast forward","status":"422"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
			return
		}
		stub.serveDefault(t, w, r)
	})

	_, err := stub.client(t).CommitToRepo(context.Background(), GitCommitData{
		RepoName:      "sites",
		Domain:        "site.com",
		Files:         map[string]interface{}{"kept.html": "<html/>"},
		Deletions:     []string{"doomed.html"},
		CommitMessage: "race",
	})
	if err != nil {
		t.Fatalf("CommitToRepo should re-base and succeed, got: %v", err)
	}

	if got := stub.contentsGET.Load(); got != 2 {
		t.Fatalf("existence checked %d times, want 2 — once per attempt. If this is 1, the deletion set was hoisted above the retry loop and a re-base reuses a stale answer", got)
	}
	if len(stub.treeEntries) != 2 {
		t.Fatalf("createTree called %d times, want 2", len(stub.treeEntries))
	}
	if got := stub.blobCreates.Load(); got != 1 {
		t.Errorf("blobs created %d times, want 1 — content-addressed, still hoisted", got)
	}
	for _, e := range stub.treeEntries[1] {
		if e["sha"] == nil {
			t.Errorf("second attempt still carries a removal for a path the new head no longer holds: %+v", e)
		}
	}
}

// serveDefault is the stub's normal routing, reused by tests that wrap one
// endpoint. Kept in sync with newDeleteStub by delegation, not by copying.
func (s *deleteStub) serveDefault(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	path, method := r.URL.Path, r.Method
	switch {
	case method == "GET" && path == "/repos/testorg/sites":
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "sites", "html_url": "https://github.com/testorg/sites",
			"default_branch": "master",
			"owner":          map[string]string{"login": "testorg"},
		})
	case method == "GET" && strings.HasPrefix(path, "/repos/testorg/sites/contents/"):
		s.contentsGET.Add(1)
		p := strings.TrimPrefix(path, "/repos/testorg/sites/contents/")
		if s.exists[p] {
			json.NewEncoder(w).Encode(map[string]string{"path": p, "sha": "blob-existing"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	case method == "GET" && strings.HasPrefix(path, "/repos/testorg/sites/git/ref/heads/"):
		json.NewEncoder(w).Encode(map[string]interface{}{
			"object": map[string]string{"sha": s.headSHA.Load().(string)},
		})
	case method == "POST" && strings.HasSuffix(path, "/git/blobs"):
		n := s.blobCreates.Add(1)
		json.NewEncoder(w).Encode(map[string]string{"sha": fmt.Sprintf("blob-%d", n)})
	case method == "POST" && strings.HasSuffix(path, "/git/trees"):
		var body struct {
			Tree []map[string]interface{} `json:"tree"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		s.treeEntries = append(s.treeEntries, body.Tree)
		json.NewEncoder(w).Encode(map[string]string{"sha": "tree-1"})
	case method == "POST" && strings.HasSuffix(path, "/git/commits"):
		s.commits.Add(1)
		json.NewEncoder(w).Encode(map[string]string{"sha": "commit-1"})
	default:
		t.Errorf("unexpected request: %s %s", method, path)
		w.WriteHeader(http.StatusNotFound)
	}
}

// TestExistenceCheckErrorIsNotTreatedAsAbsent. A 500 or a rate limit must fail
// the retraction, not silently skip it — reporting success for a removal that
// did not happen is the exact failure mode bugs_open/098 is about.
func TestExistenceCheckErrorIsNotTreatedAsAbsent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/repos/testorg/sites":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "sites", "html_url": "https://github.com/testorg/sites",
				"default_branch": "master",
				"owner":          map[string]string{"login": "testorg"},
			})
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/repos/testorg/sites/git/ref/heads/"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"object": map[string]string{"sha": "base-1"},
			})
		case strings.HasPrefix(r.URL.Path, "/repos/testorg/sites/contents/"):
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Server Error"}`))
		default:
			t.Errorf("unexpected request: %s %s — nothing should be written after the check fails", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewGitHubClient("test-token", "testorg", server.URL, zap.NewNop())
	if err != nil {
		t.Fatalf("NewGitHubClient: %v", err)
	}
	if _, err := client.CommitToRepo(context.Background(), GitCommitData{
		RepoName:      "sites",
		Domain:        "site.com",
		Deletions:     []string{"page.html"},
		CommitMessage: "retract",
	}); err == nil {
		t.Fatal("a failing existence check must fail the retraction, not report success")
	}
}

// TestDeletionPathGuardRefusesTraversal. The guard is proved by mutation: the
// SAME call with a plain path succeeds against the same stub, so the refusal is
// attributable to the path shape rather than to the fixture.
func TestDeletionPathGuardRefusesTraversal(t *testing.T) {
	refused := []struct{ path, why string }{
		{"../other-site.com/index.html", "traversal out of the caller's domain directory"},
		{"/absolute.html", "absolute path would produce an empty first tree segment"},
		{"a//b.html", "empty segment"},
		{"./here.html", "non-canonical"},
		{`win\path.html`, "backslash is taken literally by GitHub tree paths"},
		{" padded.html", "padding"},
		{"", "empty"},
	}
	for _, tc := range refused {
		t.Run(tc.why, func(t *testing.T) {
			stub := newDeleteStub(t, map[string]bool{})
			_, err := stub.client(t).CommitToRepo(context.Background(), GitCommitData{
				RepoName:      "sites",
				Domain:        "site.com",
				Deletions:     []string{tc.path},
				CommitMessage: "retract",
			})
			if err == nil {
				t.Fatalf("path %q was accepted; expected refusal (%s)", tc.path, tc.why)
			}
			if got := stub.contentsGET.Load(); got != 0 {
				t.Errorf("guard ran AFTER %d network calls; it must refuse before anything is requested", got)
			}
		})
	}

	// The mutation control: the same shape, a legal path, same stub — succeeds.
	stub := newDeleteStub(t, map[string]bool{"site.com/legal.html": true})
	if _, err := stub.client(t).CommitToRepo(context.Background(), GitCommitData{
		RepoName:      "sites",
		Domain:        "site.com",
		Deletions:     []string{"legal.html"},
		CommitMessage: "retract",
	}); err != nil {
		t.Fatalf("control: a plain path must be accepted, got %v", err)
	}
}

// TestEmptyRequestIsStillRefused — relaxing the guard for deletion-only commits
// must not have made a nothing-at-all commit legal.
func TestEmptyRequestIsStillRefused(t *testing.T) {
	stub := newDeleteStub(t, map[string]bool{})
	if _, err := stub.client(t).CommitToRepo(context.Background(), GitCommitData{
		RepoName:      "sites",
		Domain:        "site.com",
		CommitMessage: "nothing",
	}); err == nil {
		t.Fatal("a commit with neither files nor deletions must be refused")
	}
}
