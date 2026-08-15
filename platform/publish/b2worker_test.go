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
	b, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	f.objects[key] = b
	return "s3://fake/" + key, nil
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
