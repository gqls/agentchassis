package main

// engine_test.go — the tests that matter for a service holding other people's
// private writing. Isolation first, then the quota, then the migration path.
//
// Needs a real Postgres: NOTED_TEST_DSN=postgres://... go test ./...
// Skips (loudly) rather than passing when that is absent — a test suite that
// silently tests nothing is worse than one that fails.

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	dsn := os.Getenv("NOTED_TEST_DSN")
	if dsn == "" {
		t.Skip("NOTED_TEST_DSN not set — skipping (this suite proves nothing without a database)")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	// Fresh schema per run so one test's rows cannot explain another's pass.
	if _, err := db.ExecContext(ctx,
		`DROP TABLE IF EXISTS media, sessions, notes, accounts CASCADE`); err != nil {
		t.Fatalf("reset: %v", err)
	}
	st := &Store{DB: db, QuotaBytes: 1024 * 1024} // 1 MB, so the quota is reachable in a test
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &Server{Store: st, SecureCookies: false, SessionTTL: time.Hour, MaxUploadBytes: 5 << 20}
}

// signUp returns the session cookie for a new account.
func signUp(t *testing.T, s *Server, email string) *http.Cookie {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"password":"a-long-enough-password"}`, email)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/register", strings.NewReader(body))
	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s: got %d — %s", email, rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookie {
			return c
		}
	}
	t.Fatal("no session cookie set on register")
	return nil
}

func do(t *testing.T, s *Server, method, path string, c *http.Cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if c != nil {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	return rec
}

// The single most important property of this service.
func TestOneAccountCannotSeeAnothersNotes(t *testing.T) {
	s := newTestServer(t)
	alice := signUp(t, s, "alice@example.com")
	bob := signUp(t, s, "bob@example.com")

	rec := do(t, s, "POST", "/api/notes", alice, `{"title":"Alice private","content":"her diary"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("alice save: %d %s", rec.Code, rec.Body.String())
	}
	var saved Note
	json.Unmarshal(rec.Body.Bytes(), &saved)

	rec = do(t, s, "GET", "/api/notes", bob, "")
	if strings.Contains(rec.Body.String(), "Alice private") || strings.Contains(rec.Body.String(), "her diary") {
		t.Fatalf("LEAK: bob's note list contains alice's note: %s", rec.Body.String())
	}

	// Bob knows the id and asks directly — the id is guessable, so knowing it
	// must not be enough.
	rec = do(t, s, "POST", "/api/notes", bob,
		fmt.Sprintf(`{"id":%d,"title":"hijacked","content":"overwritten"}`, saved.ID))
	if rec.Code == http.StatusOK {
		t.Fatalf("LEAK: bob overwrote alice's note by id")
	}
	rec = do(t, s, "DELETE", fmt.Sprintf("/api/notes/%d", saved.ID), bob, "")
	if rec.Code == http.StatusOK {
		t.Fatalf("LEAK: bob deleted alice's note by id")
	}

	// And alice's note is untouched by both attempts.
	rec = do(t, s, "GET", "/api/notes", alice, "")
	if !strings.Contains(rec.Body.String(), "her diary") {
		t.Fatalf("alice's note was damaged: %s", rec.Body.String())
	}
}

func TestMediaIsAccountScoped(t *testing.T) {
	s := newTestServer(t)
	alice := signUp(t, s, "a@example.com")
	bob := signUp(t, s, "b@example.com")

	rec := do(t, s, "POST", "/api/notes", alice, `{"title":"with audio","content":"x"}`)
	var n Note
	json.Unmarshal(rec.Body.Bytes(), &n)

	// Bob may not attach to alice's note even with a valid session of his own.
	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=audio", n.ID), bob, "pretend audio")
	if rec.Code == http.StatusCreated {
		t.Fatalf("LEAK: bob attached media to alice's note")
	}

	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=audio", n.ID), alice, "real audio bytes")
	if rec.Code != http.StatusCreated {
		t.Fatalf("alice upload: %d %s", rec.Code, rec.Body.String())
	}
	var up struct{ ID int64 }
	json.Unmarshal(rec.Body.Bytes(), &up)

	rec = do(t, s, "GET", fmt.Sprintf("/api/media/%d", up.ID), bob, "")
	if rec.Code == http.StatusOK {
		t.Fatalf("LEAK: bob downloaded alice's recording — body %q", rec.Body.String())
	}
	rec = do(t, s, "GET", fmt.Sprintf("/api/media/%d", up.ID), alice, "")
	if rec.Code != http.StatusOK || rec.Body.String() != "real audio bytes" {
		t.Fatalf("alice cannot read her own media: %d %q", rec.Code, rec.Body.String())
	}
}

func TestUnauthenticatedIsRefused(t *testing.T) {
	s := newTestServer(t)
	for _, p := range []string{"/api/notes", "/api/me"} {
		if rec := do(t, s, "GET", p, nil, ""); rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s without a cookie returned %d, want 401", p, rec.Code)
		}
	}
	bogus := &http.Cookie{Name: sessionCookie, Value: "not-a-real-token"}
	if rec := do(t, s, "GET", "/api/notes", bogus, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("forged cookie returned %d, want 401", rec.Code)
	}
}

func TestQuotaIsEnforced(t *testing.T) {
	s := newTestServer(t) // 1 MB quota
	c := signUp(t, s, "quota@example.com")
	rec := do(t, s, "POST", "/api/notes", c, `{"title":"big","content":""}`)
	var n Note
	json.Unmarshal(rec.Body.Bytes(), &n)

	chunk := strings.Repeat("x", 400*1024) // 400 KB
	codes := []int{}
	for i := 0; i < 4; i++ {
		rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", n.ID), c, chunk)
		codes = append(codes, rec.Code)
	}
	// 400 KB * 3 = 1.2 MB > 1 MB, so the third or fourth must be refused.
	if codes[2] != http.StatusInsufficientStorage && codes[3] != http.StatusInsufficientStorage {
		t.Fatalf("quota never bit: codes %v", codes)
	}
	if codes[0] != http.StatusCreated {
		t.Fatalf("first upload should have been accepted, got %d", codes[0])
	}
}

// The migration path off the browser-only app. The shape here is fixed by files
// already on users' disks, so this test is a contract, not a preference.
func TestImportOfTheRealBackupFormat(t *testing.T) {
	s := newTestServer(t)
	c := signUp(t, s, "import@example.com")

	audio := base64.StdEncoding.EncodeToString([]byte("voice recording bytes"))
	image := base64.StdEncoding.EncodeToString([]byte("photo bytes"))
	backup := fmt.Sprintf(`{
      "format":"noted.co.uk/full-backup","version":1,
      "notes":[{"id":"abc-123","title":"Shopping","content":"milk"},
               {"id":"def-456","title":"Ideas","content":"a thought"}],
      "audio":{"abc-123":["data:audio/webm;base64,%s"]},
      "images":{"abc-123":["data:image/jpeg;base64,%s"]}}`, audio, image)

	rec := do(t, s, "POST", "/api/import", c, backup)
	if rec.Code != http.StatusOK {
		t.Fatalf("import: %d %s", rec.Code, rec.Body.String())
	}
	var res struct{ Notes, Recordings, Photos, Skipped int }
	json.Unmarshal(rec.Body.Bytes(), &res)
	if res.Notes != 2 || res.Recordings != 1 || res.Photos != 1 || res.Skipped != 0 {
		t.Fatalf("import counts wrong: %+v — body %s", res, rec.Body.String())
	}

	rec = do(t, s, "GET", "/api/notes", c, "")
	for _, want := range []string{"Shopping", "milk", "Ideas", "a thought"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("imported note missing %q: %s", want, rec.Body.String())
		}
	}

	// Importing the SAME file again must not duplicate: people re-import when
	// they are anxious about whether it worked, which is exactly when a
	// duplicate would be most upsetting.
	rec = do(t, s, "POST", "/api/import", c, backup)
	if rec.Code != http.StatusOK {
		t.Fatalf("second import: %d %s", rec.Code, rec.Body.String())
	}
	rec = do(t, s, "GET", "/api/notes", c, "")
	var listed struct{ Notes []Note }
	json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed.Notes) != 2 {
		t.Fatalf("re-import duplicated notes: now %d, want 2", len(listed.Notes))
	}
}

func TestLoginDoesNotRevealWhetherAnAccountExists(t *testing.T) {
	s := newTestServer(t)
	signUp(t, s, "real@example.com")

	wrongPw := do(t, s, "POST", "/api/login", nil, `{"email":"real@example.com","password":"wrong-password-here"}`)
	noAcct := do(t, s, "POST", "/api/login", nil, `{"email":"ghost@example.com","password":"wrong-password-here"}`)

	if wrongPw.Code != noAcct.Code {
		t.Fatalf("status differs: existing=%d absent=%d — that is an account oracle", wrongPw.Code, noAcct.Code)
	}
	if wrongPw.Body.String() != noAcct.Body.String() {
		t.Fatalf("body differs:\n existing: %s absent:   %s", wrongPw.Body.String(), noAcct.Body.String())
	}
}

func TestPasswordHashingRoundTrips(t *testing.T) {
	h, err := hashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(h, "correct horse") {
		t.Fatal("the password appears in its own hash")
	}
	if !verifyPassword(h, "correct horse battery staple") {
		t.Fatal("correct password rejected")
	}
	if verifyPassword(h, "correct horse battery stapl") {
		t.Fatal("wrong password accepted")
	}
	// Two hashes of the same password must differ, or the salt is not working.
	h2, _ := hashPassword("correct horse battery staple")
	if h == h2 {
		t.Fatal("identical hashes for the same password — salt is not being applied")
	}
}
