package git

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// F1.1b(c): CreateBranch / CreatePullRequest against a stub GitHub API.
// The stub covers the three paths that matter: a fresh branch, the
// already-exists fallback (a re-fired implementer run must not die on its own
// leftovers), and PR creation with base defaulting to the repo default branch.
func newStubGitHub(t *testing.T, branchExists bool) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// getRepo — never auto-creates.
	mux.HandleFunc("/repos/testorg/platform", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "platform", "html_url": "https://github.test/testorg/platform",
			"default_branch": "main",
			"owner":          map[string]string{"login": "testorg"},
		})
	})
	// head of main / head of the fix branch
	mux.HandleFunc("/repos/testorg/platform/git/ref/heads/main", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"object": map[string]string{"sha": "mainsha111"}})
	})
	mux.HandleFunc("/repos/testorg/platform/git/ref/heads/fix/e08c5b01", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"object": map[string]string{"sha": "existingsha222"}})
	})
	// ref creation
	mux.HandleFunc("/repos/testorg/platform/git/refs", func(w http.ResponseWriter, r *http.Request) {
		if branchExists {
			w.WriteHeader(422)
			w.Write([]byte(`{"message": "Reference already exists"}`))
			return
		}
		var body struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if !strings.HasPrefix(body.Ref, "refs/heads/") || body.SHA == "" {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{"object": map[string]string{"sha": body.SHA}})
	})
	// PR creation — echoes head/base so the test can assert base defaulting.
	mux.HandleFunc("/repos/testorg/platform/pulls", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Head string `json:"head"`
			Base string `json:"base"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.Base != "main" {
			w.WriteHeader(400)
			w.Write([]byte(`{"message": "unexpected base"}`))
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"html_url": "https://github.test/testorg/platform/pull/7", "number": 7,
		})
	})
	return httptest.NewServer(mux)
}

func testClient(t *testing.T, apiBase string) *GitHubClient {
	t.Helper()
	c, err := NewGitHubClient("test-token", "testorg", apiBase, zap.NewNop())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	return c
}

func TestCreateBranchFresh(t *testing.T) {
	srv := newStubGitHub(t, false)
	defer srv.Close()
	c := testClient(t, srv.URL)

	sha, created, err := c.CreateBranch(context.Background(),
		GitCreateBranchData{RepoName: "platform", Branch: "fix/e08c5b01"})
	if err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	if !created || sha != "mainsha111" {
		t.Fatalf("want created=true sha=mainsha111, got created=%v sha=%s", created, sha)
	}
}

func TestCreateBranchAlreadyExists(t *testing.T) {
	srv := newStubGitHub(t, true)
	defer srv.Close()
	c := testClient(t, srv.URL)

	sha, created, err := c.CreateBranch(context.Background(),
		GitCreateBranchData{RepoName: "platform", Branch: "fix/e08c5b01"})
	if err != nil {
		t.Fatalf("CreateBranch on existing: %v", err)
	}
	if created || sha != "existingsha222" {
		t.Fatalf("want created=false sha=existingsha222, got created=%v sha=%s", created, sha)
	}
}

func TestCreatePullRequestDefaultsBase(t *testing.T) {
	srv := newStubGitHub(t, false)
	defer srv.Close()
	c := testClient(t, srv.URL)

	url, number, err := c.CreatePullRequest(context.Background(),
		GitCreatePRData{RepoName: "platform", Title: "fix: guides", Head: "fix/e08c5b01"})
	if err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}
	if number != 7 || url == "" {
		t.Fatalf("want number=7 and a url, got number=%d url=%q", number, url)
	}
}

func TestCreateBranchRequiresFields(t *testing.T) {
	c := testClient(t, "http://unreachable.test")
	if _, _, err := c.CreateBranch(context.Background(), GitCreateBranchData{}); err == nil {
		t.Fatal("want error on empty repo/branch")
	}
}
