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

// Stage 1 of the media pasteboard (PLAN_2026-08-24): video joins the kinds,
// media can be deleted (freeing quota), and Range requests work so a <video>
// can seek. Each property below was absent before 2026-08-24.

func TestVideoKindAcceptedAndUnknownKindRefused(t *testing.T) {
	s := newTestServer(t)
	c := signUp(t, s, "video@example.com")
	rec := do(t, s, "POST", "/api/notes", c, `{"title":"clip","content":""}`)
	var n Note
	json.Unmarshal(rec.Body.Bytes(), &n)

	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=video", n.ID), c, "mp4 bytes")
	if rec.Code != http.StatusCreated {
		t.Fatalf("video upload refused: %d %s", rec.Code, rec.Body.String())
	}
	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=document", n.ID), c, "pdf bytes")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown kind accepted: %d %s", rec.Code, rec.Body.String())
	}

	// The unified array carries every kind, in the order things were added.
	do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=audio", n.ID), c, "ogg bytes")
	rec = do(t, s, "GET", "/api/notes", c, "")
	var listed struct{ Notes []Note }
	json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed.Notes) != 1 || len(listed.Notes[0].Media) != 2 {
		t.Fatalf("unified media array wrong: %s", rec.Body.String())
	}
	if listed.Notes[0].Media[0].Kind != "video" || listed.Notes[0].Media[1].Kind != "audio" {
		t.Fatalf("unified media order wrong: %+v", listed.Notes[0].Media)
	}
}

func TestMediaDeleteIsAccountScopedAndFreesQuota(t *testing.T) {
	s := newTestServer(t) // 1 MB quota
	alice := signUp(t, s, "del-a@example.com")
	bob := signUp(t, s, "del-b@example.com")
	rec := do(t, s, "POST", "/api/notes", alice, `{"title":"quota","content":""}`)
	var n Note
	json.Unmarshal(rec.Body.Bytes(), &n)

	big := strings.Repeat("x", 800*1024) // 800 KB — two cannot fit in 1 MB
	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", n.ID), alice, big)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first upload: %d %s", rec.Code, rec.Body.String())
	}
	var up struct{ ID int64 }
	json.Unmarshal(rec.Body.Bytes(), &up)
	if rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", n.ID), alice, big); rec.Code != http.StatusInsufficientStorage {
		t.Fatalf("second 800 KB should have hit the 1 MB quota: %d", rec.Code)
	}

	// Bob may not delete alice's media, and his attempt must not free her quota.
	if rec = do(t, s, "DELETE", fmt.Sprintf("/api/media/%d", up.ID), bob, ""); rec.Code == http.StatusOK {
		t.Fatalf("LEAK: bob deleted alice's media")
	}
	if rec = do(t, s, "GET", fmt.Sprintf("/api/media/%d", up.ID), alice, ""); rec.Code != http.StatusOK {
		t.Fatalf("alice's media gone after bob's refused delete: %d", rec.Code)
	}

	// Alice deletes; the bytes come back to her quota and the row is gone.
	if rec = do(t, s, "DELETE", fmt.Sprintf("/api/media/%d", up.ID), alice, ""); rec.Code != http.StatusOK {
		t.Fatalf("alice delete: %d %s", rec.Code, rec.Body.String())
	}
	if rec = do(t, s, "GET", fmt.Sprintf("/api/media/%d", up.ID), alice, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("deleted media still served: %d", rec.Code)
	}
	if rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", n.ID), alice, big); rec.Code != http.StatusCreated {
		t.Fatalf("delete did not free the quota: %d %s", rec.Code, rec.Body.String())
	}
}

func TestMediaRangeRequestsSeek(t *testing.T) {
	s := newTestServer(t)
	c := signUp(t, s, "range@example.com")
	rec := do(t, s, "POST", "/api/notes", c, `{"title":"seek","content":""}`)
	var n Note
	json.Unmarshal(rec.Body.Bytes(), &n)
	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=video", n.ID), c, "0123456789")
	var up struct{ ID int64 }
	json.Unmarshal(rec.Body.Bytes(), &up)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/media/%d", up.ID), nil)
	req.AddCookie(c)
	req.Header.Set("Range", "bytes=5-8")
	out := httptest.NewRecorder()
	s.Routes().ServeHTTP(out, req)
	if out.Code != http.StatusPartialContent || out.Body.String() != "5678" {
		t.Fatalf("range not honoured: %d %q (Content-Range %q)",
			out.Code, out.Body.String(), out.Header().Get("Content-Range"))
	}
}

// Immediate account deletion (owner ruling 2026-08-25) and captions (stage 3).

func TestAccountDeletionIsCompleteAndScoped(t *testing.T) {
	s := newTestServer(t)
	alice := signUp(t, s, "del-me@example.com")
	bob := signUp(t, s, "stays@example.com")

	rec := do(t, s, "POST", "/api/notes", alice, `{"title":"mine","content":"private"}`)
	var n Note
	json.Unmarshal(rec.Body.Bytes(), &n)
	do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", n.ID), alice, "pixels")
	do(t, s, "POST", "/api/notes", bob, `{"title":"bobs","content":"kept"}`)

	// Wrong password: refused, everything intact.
	rec = do(t, s, "DELETE", "/api/account", alice, `{"password":"not-the-password"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password should refuse: %d %s", rec.Code, rec.Body.String())
	}
	if rec = do(t, s, "GET", "/api/notes", alice, ""); !strings.Contains(rec.Body.String(), "private") {
		t.Fatalf("refused delete damaged the account: %s", rec.Body.String())
	}

	// Right password: gone — sessions, notes, media, the row itself.
	rec = do(t, s, "DELETE", "/api/account", alice, `{"password":"a-long-enough-password"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body.String())
	}
	if rec = do(t, s, "GET", "/api/notes", alice, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("old session survived deletion: %d", rec.Code)
	}
	for _, q := range []string{
		`SELECT count(*) FROM accounts WHERE email='del-me@example.com'`,
		`SELECT count(*) FROM notes n JOIN accounts a ON a.id=n.account_id WHERE a.email='del-me@example.com'`,
	} {
		var c int
		s.Store.DB.QueryRow(q).Scan(&c)
		if c != 0 {
			t.Fatalf("rows survived deletion: %q -> %d", q, c)
		}
	}
	var orphans int
	s.Store.DB.QueryRow(`SELECT count(*) FROM media m LEFT JOIN accounts a ON a.id=m.account_id WHERE a.id IS NULL`).Scan(&orphans)
	if orphans != 0 {
		t.Fatalf("media rows survived the cascade: %d", orphans)
	}

	// Bob is untouched; alice's email is free again and starts EMPTY.
	if rec = do(t, s, "GET", "/api/notes", bob, ""); !strings.Contains(rec.Body.String(), "kept") {
		t.Fatalf("deleting alice damaged bob: %s", rec.Body.String())
	}
	fresh := signUp(t, s, "del-me@example.com")
	rec = do(t, s, "GET", "/api/notes", fresh, "")
	var listed struct{ Notes []Note }
	json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed.Notes) != 0 {
		t.Fatalf("re-registered account is not empty: %s", rec.Body.String())
	}
}

func TestMediaCaptionIsAccountScoped(t *testing.T) {
	s := newTestServer(t)
	alice := signUp(t, s, "cap-a@example.com")
	bob := signUp(t, s, "cap-b@example.com")
	rec := do(t, s, "POST", "/api/notes", alice, `{"title":"c","content":""}`)
	var n Note
	json.Unmarshal(rec.Body.Bytes(), &n)
	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", n.ID), alice, "img")
	var up struct{ ID int64 }
	json.Unmarshal(rec.Body.Bytes(), &up)

	if rec = do(t, s, "PATCH", fmt.Sprintf("/api/media/%d", up.ID), bob, `{"caption":"graffiti"}`); rec.Code == http.StatusOK {
		t.Fatalf("LEAK: bob captioned alice's media")
	}
	if rec = do(t, s, "PATCH", fmt.Sprintf("/api/media/%d", up.ID), alice, `{"caption":"the garden in May"}`); rec.Code != http.StatusOK {
		t.Fatalf("caption: %d %s", rec.Code, rec.Body.String())
	}
	rec = do(t, s, "GET", "/api/notes", alice, "")
	if !strings.Contains(rec.Body.String(), "the garden in May") {
		t.Fatalf("caption not in the list payload: %s", rec.Body.String())
	}
	long := strings.Repeat("x", 501)
	if rec = do(t, s, "PATCH", fmt.Sprintf("/api/media/%d", up.ID), alice, fmt.Sprintf(`{"caption":%q}`, long)); rec.Code != http.StatusBadRequest {
		t.Fatalf("overlong caption accepted: %d", rec.Code)
	}
}
