package main

// upload_chunked_test.go — the chunked upload path
// (PLAN_2026-08-26_large_uploads.md), driven through the real HTTP surface
// against the stub that verifies part sha1s like the live service. The
// properties that keep a PAID store honest are the ones asserted: quota is
// reserved from begin and released on abort, a refusal never leaks a billed
// B2 file, finish refuses gaps by name, and the reaper cleans what nobody
// will finish — including orphans only B2 knows about.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// chunkedTestServer: quota wide enough for multi-part files; the single-request
// cap stays at newTestServer's 5 MB so the chunked path is the only route for
// the sizes these tests use.
func chunkedTestServer(t *testing.T) (*Server, *b2Stub) {
	s, stub := newTestServerB2(t)
	s.Store.QuotaBytes = 64 << 20
	return s, stub
}

func mkNoteID(t *testing.T, s *Server, c *http.Cookie) int64 {
	t.Helper()
	rec := do(t, s, "POST", "/api/notes", c, `{"title":"big","content":""}`)
	if rec.Code >= 300 {
		t.Fatalf("create note: %d %s", rec.Code, rec.Body.String())
	}
	var n Note
	decodeBody(t, rec, &n)
	return n.ID
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
}

func TestChunkedUploadHappyPath(t *testing.T) {
	s, stub := chunkedTestServer(t)
	c := signUp(t, s, "chunk@example.com")
	noteID := mkNoteID(t, s, c)
	lift(t, s, "chunk@example.com", 64<<20)

	const size = 12 << 20 // 12 MB -> part_size 6 MB, 2 parts
	rec := do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media/uploads", noteID), c,
		fmt.Sprintf(`{"kind":"video","mime":"video/mp4","size":%d}`, size))
	if rec.Code != http.StatusCreated {
		t.Fatalf("begin: %d %s", rec.Code, rec.Body.String())
	}
	var begin struct {
		UploadID   int64 `json:"upload_id"`
		PartSize   int64 `json:"part_size"`
		PartsTotal int64 `json:"parts_total"`
	}
	decodeBody(t, rec, &begin)
	if begin.PartSize != 6<<20 || begin.PartsTotal != 2 {
		t.Fatalf("begin arithmetic: part_size=%d parts_total=%d", begin.PartSize, begin.PartsTotal)
	}

	p1 := strings.Repeat("a", int(begin.PartSize))
	p2 := strings.Repeat("z", size-int(begin.PartSize))
	if rec = do(t, s, "PUT", fmt.Sprintf("/api/uploads/%d/parts/1", begin.UploadID), c, p1); rec.Code != 200 {
		t.Fatalf("part 1: %d %s", rec.Code, rec.Body.String())
	}
	if rec = do(t, s, "PUT", fmt.Sprintf("/api/uploads/%d/parts/2", begin.UploadID), c, p2); rec.Code != 200 {
		t.Fatalf("part 2: %d %s", rec.Code, rec.Body.String())
	}
	rec = do(t, s, "POST", fmt.Sprintf("/api/uploads/%d/finish", begin.UploadID), c, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("finish: %d %s", rec.Code, rec.Body.String())
	}
	var fin struct {
		ID      int64 `json:"id"`
		ByteLen int64 `json:"byte_len"`
	}
	decodeBody(t, rec, &fin)
	if fin.ByteLen != size {
		t.Fatalf("byte_len: %d", fin.ByteLen)
	}

	// The assembled object is the concatenation, byte-verified at the edges.
	stub.mu.Lock()
	var stored []byte
	for _, b := range stub.objects {
		stored = b
	}
	stub.mu.Unlock()
	if len(stored) != size || stored[0] != 'a' || stored[size-1] != 'z' {
		t.Fatalf("assembled object wrong: len=%d", len(stored))
	}
	if stub.unfinishedCount() != 0 {
		t.Fatalf("unfinished large file left behind: %d", stub.unfinishedCount())
	}

	// Quota charged once, reservation gone.
	var mediaBytes, pendingCount int64
	s.Store.DB.QueryRow(`SELECT media_bytes FROM accounts WHERE email='chunk@example.com'`).Scan(&mediaBytes)
	s.Store.DB.QueryRow(`SELECT count(*) FROM pending_uploads`).Scan(&pendingCount)
	if mediaBytes != size || pendingCount != 0 {
		t.Fatalf("quota/reservation after finish: media_bytes=%d pending=%d", mediaBytes, pendingCount)
	}
}

func TestChunkedBeginRefusalsDoNotLeak(t *testing.T) {
	s, stub := chunkedTestServer(t) // 64 MB quota
	c := signUp(t, s, "chunkrefuse@example.com")
	noteID := mkNoteID(t, s, c)

	// Beyond the per-file cap (env default 5 MB, no override) → 413, and
	// nothing was started on B2 to leak.
	rec := do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media/uploads", noteID), c,
		fmt.Sprintf(`{"kind":"video","size":%d}`, 70<<20))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("beyond max_upload should 413: %d %s", rec.Code, rec.Body.String())
	}

	// At or under the single-request cap the chunked path refuses outright.
	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media/uploads", noteID), c,
		fmt.Sprintf(`{"kind":"image","size":%d}`, 1<<20))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("small file should be told to use the normal path: %d", rec.Code)
	}

	// Over QUOTA (cap lifted past it): the begin refuses 507 AND cancels the
	// large file it had already started — the property that keeps a refusal
	// from leaking billed storage.
	lift(t, s, "chunkrefuse@example.com", 100<<20)
	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media/uploads", noteID), c,
		fmt.Sprintf(`{"kind":"video","size":%d}`, 70<<20))
	if rec.Code != http.StatusInsufficientStorage {
		t.Fatalf("beyond quota should 507: %d %s", rec.Code, rec.Body.String())
	}
	if stub.unfinishedCount() != 0 {
		t.Fatalf("a refused begin leaked an unfinished large file: %d", stub.unfinishedCount())
	}
}

// lift raises the account's per-file cap so multi-part sizes are reachable in
// tests (the default cap is newTestServer's 5 MB); quota stays the server's.
func lift(t *testing.T, s *Server, email string, maxUpload int64) {
	t.Helper()
	if _, err := s.Store.DB.Exec(
		`UPDATE accounts SET max_upload_override_bytes=$2 WHERE email=$1`, email, maxUpload); err != nil {
		t.Fatal(err)
	}
}

func TestReservationBlocksOtherUploadsUntilAbort(t *testing.T) {
	s, stub := chunkedTestServer(t) // 64 MB quota
	c := signUp(t, s, "reserve@example.com")
	noteID := mkNoteID(t, s, c)
	lift(t, s, "reserve@example.com", 64<<20)

	rec := do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media/uploads", noteID), c,
		fmt.Sprintf(`{"kind":"video","size":%d}`, 60<<20))
	if rec.Code != http.StatusCreated {
		t.Fatalf("begin: %d %s", rec.Code, rec.Body.String())
	}
	var begin struct {
		UploadID int64 `json:"upload_id"`
	}
	decodeBody(t, rec, &begin)

	// A small-path upload that would fit an empty account must now refuse:
	// 60 MB of the 64 are promised away.
	small := strings.Repeat("x", 5<<20)
	if rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", noteID), c, small); rec.Code != http.StatusInsufficientStorage {
		t.Fatalf("reservation should block the small path: %d %s", rec.Code, rec.Body.String())
	}

	// Abort releases the reservation and cancels the B2 half...
	if rec = do(t, s, "DELETE", fmt.Sprintf("/api/uploads/%d", begin.UploadID), c, ""); rec.Code != 200 {
		t.Fatalf("abort: %d %s", rec.Code, rec.Body.String())
	}
	if stub.unfinishedCount() != 0 {
		t.Fatalf("abort left an unfinished large file: %d", stub.unfinishedCount())
	}
	// ...and the same small upload now fits.
	if rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", noteID), c, small); rec.Code != http.StatusCreated {
		t.Fatalf("after abort the small upload should fit: %d %s", rec.Code, rec.Body.String())
	}
}

func TestFinishRefusesGapsByName(t *testing.T) {
	s, _ := chunkedTestServer(t)
	c := signUp(t, s, "gaps@example.com")
	noteID := mkNoteID(t, s, c)
	lift(t, s, "gaps@example.com", 64<<20)

	const size = 12 << 20
	rec := do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media/uploads", noteID), c,
		fmt.Sprintf(`{"kind":"video","size":%d}`, size))
	var begin struct {
		UploadID int64 `json:"upload_id"`
		PartSize int64 `json:"part_size"`
	}
	decodeBody(t, rec, &begin)

	// Wrong-size part refused before any bytes reach B2's ledger.
	if rec = do(t, s, "PUT", fmt.Sprintf("/api/uploads/%d/parts/1", begin.UploadID), c, "tiny"); rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong-size part should 400: %d %s", rec.Code, rec.Body.String())
	}

	p1 := strings.Repeat("a", int(begin.PartSize))
	if rec = do(t, s, "PUT", fmt.Sprintf("/api/uploads/%d/parts/1", begin.UploadID), c, p1); rec.Code != 200 {
		t.Fatalf("part 1: %d", rec.Code)
	}
	rec = do(t, s, "POST", fmt.Sprintf("/api/uploads/%d/finish", begin.UploadID), c, "")
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "part 2") {
		t.Fatalf("finish should name the missing part: %d %s", rec.Code, rec.Body.String())
	}
}

func TestAccountDeletionCancelsPendingUploads(t *testing.T) {
	s, stub := chunkedTestServer(t)
	c := signUp(t, s, "delpending@example.com")
	noteID := mkNoteID(t, s, c)
	lift(t, s, "delpending@example.com", 64<<20)

	rec := do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media/uploads", noteID), c,
		fmt.Sprintf(`{"kind":"video","size":%d}`, 20<<20))
	if rec.Code != http.StatusCreated {
		t.Fatalf("begin: %d", rec.Code)
	}
	if stub.unfinishedCount() != 1 {
		t.Fatalf("expected 1 unfinished, got %d", stub.unfinishedCount())
	}
	rec = do(t, s, "DELETE", "/api/account", c, `{"password":"a-long-enough-password"}`)
	if rec.Code != 200 {
		t.Fatalf("delete account: %d %s", rec.Code, rec.Body.String())
	}
	if stub.unfinishedCount() != 0 {
		t.Fatalf("account deletion left an unfinished large file: %d", stub.unfinishedCount())
	}
}

func TestReaperReleasesStaleAndOrphanedUploads(t *testing.T) {
	s, stub := chunkedTestServer(t)
	c := signUp(t, s, "reap@example.com")
	noteID := mkNoteID(t, s, c)
	lift(t, s, "reap@example.com", 64<<20)

	// A stale reservation: begun, then abandoned 25 hours ago.
	rec := do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media/uploads", noteID), c,
		fmt.Sprintf(`{"kind":"video","size":%d}`, 20<<20))
	if rec.Code != http.StatusCreated {
		t.Fatalf("begin: %d", rec.Code)
	}
	if _, err := s.Store.DB.Exec(
		`UPDATE pending_uploads SET created_at = now() - interval '25 hours'`); err != nil {
		t.Fatal(err)
	}

	// Two rowless B2 orphans (a crash between start and insert): one aged —
	// reapable — and one young, which must be LEFT (it may be a begin whose
	// row is milliseconds away).
	agedID, _ := s.B2.StartLargeFile("media/acct_999/orphan-aged", "video/mp4")
	stub.backdateUnfinished(agedID, 25*time.Hour)
	if _, err := s.B2.StartLargeFile("media/acct_999/orphan-young", "video/mp4"); err != nil {
		t.Fatal(err)
	}

	s.ReapAbandonedUploads(context.Background())

	var pendingCount int64
	s.Store.DB.QueryRow(`SELECT count(*) FROM pending_uploads`).Scan(&pendingCount)
	if pendingCount != 0 {
		t.Fatalf("stale reservation survived the reaper: %d", pendingCount)
	}
	if got := stub.unfinishedCount(); got != 1 {
		t.Fatalf("want exactly the young orphan left, got %d unfinished", got)
	}
	// And the released quota is usable again.
	small := strings.Repeat("x", 1<<20)
	if rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", noteID), c, small); rec.Code != http.StatusCreated {
		t.Fatalf("quota not released by the reaper: %d %s", rec.Code, rec.Body.String())
	}
}

func TestOverridesAreReportedAndHonoured(t *testing.T) {
	s, _ := chunkedTestServer(t)
	c := signUp(t, s, "paid@example.com")
	if _, err := s.Store.DB.Exec(
		`UPDATE accounts SET media_quota_override_bytes=$1, max_upload_override_bytes=$2
		 WHERE email='paid@example.com'`, int64(200<<20), int64(100<<20)); err != nil {
		t.Fatal(err)
	}
	rec := do(t, s, "GET", "/api/me", c, "")
	var me struct {
		MediaQuota     int64 `json:"media_quota"`
		MaxUpload      int64 `json:"max_upload"`
		SmallUploadMax int64 `json:"small_upload_max"`
	}
	decodeBody(t, rec, &me)
	if me.MediaQuota != 200<<20 || me.MaxUpload != 100<<20 || me.SmallUploadMax != 5<<20 {
		t.Fatalf("/api/me overrides: %+v", me)
	}

	// A 90 MB begin — far past the env cap — is allowed under the override.
	noteID := mkNoteID(t, s, c)
	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media/uploads", noteID), c,
		fmt.Sprintf(`{"kind":"video","size":%d}`, 90<<20))
	if rec.Code != http.StatusCreated {
		t.Fatalf("override should allow the begin: %d %s", rec.Code, rec.Body.String())
	}
}
