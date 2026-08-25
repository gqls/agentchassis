package main

// b2_test.go — the B2 path, proven against a local stub that speaks the four
// B2 endpoints the engine uses. The stub COUNTS objects, so the tests can
// assert the property that matters most for a paid store: no code path leaks
// an object (a quota refusal deletes what it just uploaded; a delete deletes).

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type b2Stub struct {
	mu      sync.Mutex
	objects map[string][]byte // key -> bytes
	fileIDs map[string]string // fileId -> key
	nextID  int
	srv     *httptest.Server
}

func newB2Stub(t *testing.T) *b2Stub {
	t.Helper()
	st := &b2Stub{objects: map[string][]byte{}, fileIDs: map[string]string{}}
	mux := http.NewServeMux()
	// v4 shape, matching the live probe of 2026-08-25 (v2/v3 are refused by
	// the real service, so the stub does not offer them either).
	mux.HandleFunc("GET /b2api/v4/b2_authorize_account", func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := r.BasicAuth(); !ok {
			w.WriteHeader(401)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"authorizationToken": "acct-tok",
			"apiInfo": map[string]any{"storageApi": map[string]any{
				"apiUrl":      st.srv.URL,
				"downloadUrl": st.srv.URL,
				"allowed": map[string]any{"buckets": []map[string]any{
					{"id": "bkt1", "name": "personae-noted-media"}}},
			}},
		})
	})
	mux.HandleFunc("POST /b2api/v4/b2_get_upload_url", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"uploadUrl": st.srv.URL + "/upload", "authorizationToken": "up-tok"})
	})
	mux.HandleFunc("POST /upload", func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Bz-File-Name")
		if key == "" {
			w.WriteHeader(400)
			return
		}
		data, _ := readAll(r.Body)
		st.mu.Lock()
		st.nextID++
		id := fmt.Sprintf("fid-%d", st.nextID)
		st.objects[key] = data
		st.fileIDs[id] = key
		st.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]any{"fileId": id, "fileName": key})
	})
	mux.HandleFunc("GET /file/personae-noted-media/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/file/personae-noted-media/")
		st.mu.Lock()
		data, ok := st.objects[key]
		st.mu.Unlock()
		if !ok {
			w.WriteHeader(404)
			return
		}
		// ServeContent gives the stub real Range behaviour for free.
		http.ServeContent(w, r, "", time.Time{}, strings.NewReader(string(data)))
	})
	mux.HandleFunc("POST /b2api/v4/b2_delete_file_version", func(w http.ResponseWriter, r *http.Request) {
		var in struct{ FileName, FileId string }
		json.NewDecoder(r.Body).Decode(&in)
		st.mu.Lock()
		defer st.mu.Unlock()
		if _, ok := st.fileIDs[in.FileId]; !ok {
			w.WriteHeader(400)
			w.Write([]byte(`{"code":"file_not_present"}`))
			return
		}
		delete(st.objects, in.FileName)
		delete(st.fileIDs, in.FileId)
		json.NewEncoder(w).Encode(map[string]any{})
	})
	st.srv = httptest.NewServer(mux)
	t.Cleanup(st.srv.Close)
	return st
}

func (st *b2Stub) count() int {
	st.mu.Lock()
	defer st.mu.Unlock()
	return len(st.objects)
}

func newTestServerB2(t *testing.T) (*Server, *b2Stub) {
	t.Helper()
	s := newTestServer(t)
	stub := newB2Stub(t)
	s.B2 = &B2{keyID: "k", appKey: "s", apiBase: stub.srv.URL,
		HTTP: &http.Client{Timeout: 10 * time.Second}}
	return s, stub
}

func TestB2MediaRoundTrip(t *testing.T) {
	s, stub := newTestServerB2(t)
	c := signUp(t, s, "b2@example.com")
	rec := do(t, s, "POST", "/api/notes", c, `{"title":"b2","content":""}`)
	var n Note
	json.Unmarshal(rec.Body.Bytes(), &n)

	rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=video", n.ID), c, "0123456789")
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body.String())
	}
	var up struct{ ID int64 }
	json.Unmarshal(rec.Body.Bytes(), &up)
	if stub.count() != 1 {
		t.Fatalf("B2 should hold 1 object, holds %d", stub.count())
	}

	// The row must point at B2 and carry NO inline bytes.
	var storageKey string
	var bytesLen any
	if err := s.Store.DB.QueryRow(
		`SELECT COALESCE(storage_key,''), octet_length(bytes) FROM media WHERE id=$1`, up.ID).
		Scan(&storageKey, &bytesLen); err != nil {
		t.Fatal(err)
	}
	if storageKey == "" || bytesLen != nil {
		t.Fatalf("row not B2-backed: storage_key=%q bytes_len=%v", storageKey, bytesLen)
	}

	// Served back through the engine (auth stays ours), Range included.
	rec = do(t, s, "GET", fmt.Sprintf("/api/media/%d", up.ID), c, "")
	if rec.Code != 200 || rec.Body.String() != "0123456789" {
		t.Fatalf("serve: %d %q", rec.Code, rec.Body.String())
	}
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/media/%d", up.ID), nil)
	req.AddCookie(c)
	req.Header.Set("Range", "bytes=5-8")
	out := httptest.NewRecorder()
	s.Routes().ServeHTTP(out, req)
	if out.Code != http.StatusPartialContent || out.Body.String() != "5678" {
		t.Fatalf("range through the proxy: %d %q", out.Code, out.Body.String())
	}

	// Delete removes the OBJECT and the row, and hands the quota back.
	rec = do(t, s, "DELETE", fmt.Sprintf("/api/media/%d", up.ID), c, "")
	if rec.Code != 200 {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body.String())
	}
	if stub.count() != 0 {
		t.Fatalf("B2 object leaked after delete: %d", stub.count())
	}
	var mediaBytes int64
	s.Store.DB.QueryRow(`SELECT media_bytes FROM accounts WHERE email='b2@example.com'`).Scan(&mediaBytes)
	if mediaBytes != 0 {
		t.Fatalf("quota not freed: %d", mediaBytes)
	}
}

func TestB2QuotaRefusalDeletesTheUploadedObject(t *testing.T) {
	s, stub := newTestServerB2(t) // 1 MB quota
	c := signUp(t, s, "b2quota@example.com")
	rec := do(t, s, "POST", "/api/notes", c, `{"title":"q","content":""}`)
	var n Note
	json.Unmarshal(rec.Body.Bytes(), &n)

	big := strings.Repeat("x", 800*1024)
	if rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", n.ID), c, big); rec.Code != http.StatusCreated {
		t.Fatalf("first upload: %d", rec.Code)
	}
	if rec = do(t, s, "POST", fmt.Sprintf("/api/notes/%d/media?kind=image", n.ID), c, big); rec.Code != http.StatusInsufficientStorage {
		t.Fatalf("second should hit the quota: %d", rec.Code)
	}
	// The property that keeps a PAID store honest: the refused upload's object
	// must be gone from B2, not orphaned there for ever.
	if stub.count() != 1 {
		t.Fatalf("quota refusal leaked a B2 object: count=%d, want 1", stub.count())
	}
}

func TestLayoutRidesTheSave(t *testing.T) {
	s := newTestServer(t)
	c := signUp(t, s, "layout@example.com")

	layout := `{"v":1,"items":[{"id":"t1","kind":"text","x":0.1,"y":0.1,"w":0.5,"h":0.2,"z":1}]}`
	rec := do(t, s, "POST", "/api/notes", c, fmt.Sprintf(`{"title":"board","content":"txt","layout":%s}`, layout))
	if rec.Code != 200 {
		t.Fatalf("save with layout: %d %s", rec.Code, rec.Body.String())
	}
	var saved Note
	json.Unmarshal(rec.Body.Bytes(), &saved)
	if saved.Layout == nil || !strings.Contains(string(saved.Layout), `"t1"`) {
		t.Fatalf("layout not returned: %s", rec.Body.String())
	}

	// A save WITHOUT layout (an old client, or a text-only save) must not
	// erase the arrangement the row already holds.
	rec = do(t, s, "POST", "/api/notes", c, fmt.Sprintf(`{"id":%d,"title":"board","content":"txt2"}`, saved.ID))
	if rec.Code != 200 {
		t.Fatalf("save without layout: %d", rec.Code)
	}
	rec = do(t, s, "GET", "/api/notes", c, "")
	if !strings.Contains(rec.Body.String(), `"t1"`) {
		t.Fatalf("layout erased by a layout-less save: %s", rec.Body.String())
	}

	// An explicit null CLEARS it (that is how "back to linear" is said).
	rec = do(t, s, "POST", "/api/notes", c, fmt.Sprintf(`{"id":%d,"title":"board","content":"txt3","layout":null}`, saved.ID))
	if rec.Code != 200 {
		t.Fatalf("save with null layout: %d", rec.Code)
	}
	rec = do(t, s, "GET", "/api/notes", c, "")
	if strings.Contains(rec.Body.String(), `"t1"`) {
		t.Fatalf("explicit null did not clear the layout")
	}

	// And a silly-sized layout is refused.
	huge := `{"v":1,"pad":"` + strings.Repeat("x", 300*1024) + `"}`
	rec = do(t, s, "POST", "/api/notes", c, fmt.Sprintf(`{"title":"big","content":"","layout":%s}`, huge))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized layout accepted: %d", rec.Code)
	}
}

// TestB2LiveRoundTrip runs ONLY when real credentials are in the environment —
// it is the check that the stub above agrees with the actual service, which is
// exactly what the v2→v4 authorize surprise (2026-08-25) proved cannot be
// assumed. It touches the real bucket: one small object, deleted at the end.
func TestB2LiveRoundTrip(t *testing.T) {
	b := NewB2FromEnv()
	if b == nil {
		t.Skip("NOTED_B2_KEY_ID/NOTED_B2_APP_KEY not set — skipping the live check")
	}
	data := []byte("live-roundtrip " + time.Now().UTC().Format(time.RFC3339))
	sum := sha1.Sum(data)
	key := "test/live-roundtrip.txt"

	fileID, err := b.Upload(key, "text/plain", data, hex.EncodeToString(sum[:]))
	if err != nil {
		t.Fatalf("live upload: %v", err)
	}
	res, err := b.Download(key, "bytes=0-3")
	if err != nil {
		t.Fatalf("live download: %v", err)
	}
	part, _ := readAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusPartialContent || string(part) != "live" {
		t.Fatalf("live range download: %d %q", res.StatusCode, part)
	}
	if err := b.Delete(key, fileID); err != nil {
		t.Fatalf("live delete: %v", err)
	}
	// Deleting again must converge, not error — the retry-after-crash path.
	if err := b.Delete(key, fileID); err != nil {
		t.Fatalf("second delete should be already-gone success: %v", err)
	}
}
