// FILE: platform/publish/b2worker_test.go
package publish

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/storage"
)

// fakeStore is a map-backed ObjectStore whose ListObjects derives ETags from
// content MD5, the same way B2's S3 gateway does for single-part uploads.
type fakeStore struct {
	objects map[string][]byte
	// corruptKey, when set, lies about that key's ETag in listings — it
	// simulates a damaged copy so the verification guard can be PROVEN to
	// fire, not just assumed (a fake's bookkeeping cannot assert a negative).
	corruptKey string
	// deleteSilentlyFails, when set, makes Delete return nil without
	// removing anything — the sweep's post-delete re-list verification must
	// catch it, same discipline as corruptKey for the copy half.
	deleteSilentlyFails bool
}

func newFakeStore() *fakeStore { return &fakeStore{objects: map[string][]byte{}} }

func (f *fakeStore) ListObjects(_ context.Context, prefix string) ([]storage.ObjectInfo, error) {
	var out []storage.ObjectInfo
	for k, v := range f.objects {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		sum := md5.Sum(v)
		etag := hex.EncodeToString(sum[:])
		if k == f.corruptKey {
			etag = "corrupted"
		}
		out = append(out, storage.ObjectInfo{Key: k, Size: int64(len(v)), ETag: etag})
	}
	return out, nil
}

func (f *fakeStore) Download(_ context.Context, key string) (io.ReadCloser, error) {
	v, ok := f.objects[key]
	if !ok {
		return nil, fmt.Errorf("no such key %q", key)
	}
	return io.NopCloser(bytes.NewReader(v)), nil
}

func (f *fakeStore) Upload(_ context.Context, key, _ string, body io.Reader) (string, error) {
	// B2's S3 gateway requires Content-Length, which the SDK derives only
	// from a seekable body — a bare stream fails live with HTTP 411
	// MissingContentLength (hit on the first canary, 2026-08-15). The fake
	// enforces the contract the real gateway enforces, so the unit tests
	// can no longer pass a body production would reject.
	if _, ok := body.(io.Seeker); !ok {
		return "", fmt.Errorf("fake gateway: upload body for %q is not seekable — the real B2 S3 gateway returns 411 MissingContentLength for this", key)
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	f.objects[key] = b
	return "s3://fake/" + key, nil
}

func (f *fakeStore) Delete(_ context.Context, key string) error {
	if f.deleteSilentlyFails {
		return nil
	}
	delete(f.objects, key)
	return nil
}

func seedSite(f *fakeStore, domain string, files map[string]string) []File {
	var out []File
	for k, v := range files {
		f.objects[domain+"/"+k] = []byte(v)
		sum := md5.Sum([]byte(v))
		out = append(out, File{Key: k, ETag: hex.EncodeToString(sum[:]), Size: int64(len(v))})
	}
	return out
}

func TestB2WorkerCopiesTreeUnderProjectPrefix(t *testing.T) {
	store := newFakeStore()
	files := seedSite(store, "example.com", map[string]string{
		"index.html":   "<html>home</html>",
		"css/site.css": "body{}",
	})
	src := NewS3Source(store, "example.com")

	res, err := NewB2Worker(store).Publish(context.Background(), Request{
		Domain: "example.com", Project: "canary.ugg2.com", Files: files, Source: src,
	})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if res.Published != 2 {
		t.Errorf("published %d files, want 2", res.Published)
	}
	if res.URL != "https://canary.ugg2.com/" {
		t.Errorf("URL %q — the worker serves by hostname, so this must be the project host", res.URL)
	}
	if got := string(store.objects["canary.ugg2.com/index.html"]); got != "<html>home</html>" {
		t.Errorf("copied index.html = %q", got)
	}
	if _, ok := store.objects["canary.ugg2.com/css/site.css"]; !ok {
		t.Error("nested key css/site.css was not copied under the project prefix")
	}
}

func TestB2WorkerSweepsOrphansAndNeverTheCopiedSet(t *testing.T) {
	store := newFakeStore()
	files := seedSite(store, "example.com", map[string]string{
		"index.html": "<html>home</html>",
		"about.html": "<html>about</html>",
	})
	// The motivating shape (bugs_open/429): a destination key whose source
	// was retracted before this publish.
	store.objects["canary.ugg2.com/contact.html"] = []byte("<html>retracted</html>")

	res, err := NewB2Worker(store).Publish(context.Background(), Request{
		Domain: "example.com", Project: "canary.ugg2.com", Files: files,
		Source: NewS3Source(store, "example.com"),
	})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if _, ok := store.objects["canary.ugg2.com/contact.html"]; ok {
		t.Error("orphaned destination key survived the publish — the mirror still cannot unpublish")
	}
	if res.Deleted != 1 || len(res.DeletedKeys) != 1 || res.DeletedKeys[0] != "contact.html" {
		t.Errorf("result must report the sweep for the caller's 404 acceptance, got Deleted=%d keys=%v", res.Deleted, res.DeletedKeys)
	}
	for _, key := range []string{"canary.ugg2.com/index.html", "canary.ugg2.com/about.html"} {
		if _, ok := store.objects[key]; !ok {
			t.Errorf("%s is in the source set and was deleted — the sweep must only remove keys absent from source", key)
		}
	}
	// The source itself must never be touched by a destination sweep.
	if _, ok := store.objects["example.com/index.html"]; !ok {
		t.Error("source tree was modified by the sweep")
	}
}

func TestB2WorkerRefusesAnEmptyFileSet(t *testing.T) {
	store := newFakeStore()
	store.objects["canary.ugg2.com/index.html"] = []byte("still hosted")

	_, err := NewB2Worker(store).Publish(context.Background(), Request{
		Domain: "example.com", Project: "canary.ugg2.com", Files: nil,
		Source: NewS3Source(store, "example.com"),
	})
	if err == nil {
		t.Fatal("an empty file set must refuse — with the sweep it would read as 'delete the whole mirror'")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("refusal should name the empty set, got: %v", err)
	}
	if _, ok := store.objects["canary.ugg2.com/index.html"]; !ok {
		t.Error("the refusal deleted anyway — the guard must fire before any sweep")
	}
}

func TestB2WorkerBulkFloorRefusesAndTheFlagOverrides(t *testing.T) {
	store := newFakeStore()
	files := seedSite(store, "example.com", map[string]string{"index.html": "x"})
	for i := 0; i < 25; i++ {
		store.objects[fmt.Sprintf("canary.ugg2.com/old-%02d.html", i)] = []byte("orphan")
	}

	req := Request{
		Domain: "example.com", Project: "canary.ugg2.com", Files: files,
		Source: NewS3Source(store, "example.com"),
	}
	_, err := NewB2Worker(store).Publish(context.Background(), req)
	if err == nil {
		t.Fatal("25 of 26 destination keys swept in one pass must refuse without allow_bulk_unpublish")
	}
	if !strings.Contains(err.Error(), "allow_bulk_unpublish") {
		t.Errorf("refusal should name the override, got: %v", err)
	}
	if _, ok := store.objects["canary.ugg2.com/old-00.html"]; !ok {
		t.Error("the floor refused but deleted anyway")
	}

	req.AllowBulkUnpublish = true
	res, err := NewB2Worker(store).Publish(context.Background(), req)
	if err != nil {
		t.Fatalf("publish with override: %v", err)
	}
	if res.Deleted != 25 {
		t.Errorf("override sweep deleted %d, want 25", res.Deleted)
	}
	if _, ok := store.objects["canary.ugg2.com/old-00.html"]; ok {
		t.Error("override was accepted but the orphan survived")
	}
}

func TestB2WorkerSweepVerificationCatchesASilentDeleteFailure(t *testing.T) {
	store := newFakeStore()
	files := seedSite(store, "example.com", map[string]string{"index.html": "x"})
	store.objects["canary.ugg2.com/contact.html"] = []byte("orphan")
	store.deleteSilentlyFails = true

	_, err := NewB2Worker(store).Publish(context.Background(), Request{
		Domain: "example.com", Project: "canary.ugg2.com", Files: files,
		Source: NewS3Source(store, "example.com"),
	})
	if err == nil {
		t.Fatal("a Delete that returns nil without deleting published without error — the post-sweep re-list verification is not live")
	}
	if !strings.Contains(err.Error(), "sweep verification") {
		t.Errorf("error should name the sweep verification, got: %v", err)
	}
}

func TestB2WorkerVerificationCatchesACorruptedCopy(t *testing.T) {
	store := newFakeStore()
	files := seedSite(store, "example.com", map[string]string{"index.html": "<html>home</html>"})
	store.corruptKey = "canary.ugg2.com/index.html"

	_, err := NewB2Worker(store).Publish(context.Background(), Request{
		Domain: "example.com", Project: "canary.ugg2.com", Files: files,
		Source: NewS3Source(store, "example.com"),
	})
	if err == nil {
		t.Fatal("a corrupted destination copy published without error — the ETag verification is not live")
	}
	if !strings.Contains(err.Error(), "etag mismatch") {
		t.Errorf("error should name the mismatch, got: %v", err)
	}
}

func TestB2WorkerRefusesBadProjects(t *testing.T) {
	store := newFakeStore()
	files := seedSite(store, "example.com", map[string]string{"index.html": "x"})
	src := NewS3Source(store, "example.com")
	w := NewB2Worker(store)

	for name, project := range map[string]string{
		"empty":           "",
		"path not host":   "canary.ugg2.com/sub",
		"same as domain":  "example.com",
		"whitespace only": "   ",
	} {
		_, err := w.Publish(context.Background(), Request{Domain: "example.com", Project: project, Files: files, Source: src})
		if err == nil {
			t.Errorf("%s project %q accepted — must refuse", name, project)
		}
	}
}

func TestS3SourceStripsThePrefixAndSkipsTheMarker(t *testing.T) {
	store := newFakeStore()
	store.objects["example.com/index.html"] = []byte("x")
	store.objects["example.com/a/b.css"] = []byte("y")

	files, err := NewS3Source(store, "example.com").List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	for _, f := range files {
		keys[f.Key] = true
	}
	if !keys["index.html"] || !keys["a/b.css"] || len(keys) != 2 {
		t.Errorf("relative keys wrong: %v", keys)
	}
}

func TestCFPagesRefusesLoudlyUntilArmed(t *testing.T) {
	_, err := NewCFPages().Publish(context.Background(), Request{Domain: "example.com", Project: "p"})
	if err == nil {
		t.Fatal("cfpages must refuse until the owner's token exists and the client is live-verified")
	}
	if !strings.Contains(err.Error(), "b2worker") {
		t.Errorf("the refusal should point the operator at the working backend, got: %v", err)
	}
}
