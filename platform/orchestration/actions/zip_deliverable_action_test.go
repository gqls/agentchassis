// FILE: platform/orchestration/actions/zip_deliverable_action_test.go
//
// Pins the zip_deliverable contract: the archive holds exactly the tree (entry
// count == listing, index.html byte-identical), the upload body is seekable
// (B2 411s a bare stream), the presigned URL is returned, an empty tree is a
// recorded skip that never touches an archive, and an oversize tree ALERTS
// and completes rather than truncating (the demand-control path).
package actions

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/publish"
	"go.uber.org/zap"
)

// zipFakeStore adds presigning to the seekability-enforcing fake gateway.
type zipFakeStore struct {
	pubFakeStore
	presigned []string
}

func (f *zipFakeStore) GetPresignedURL(_ context.Context, key string, expiryMinutes int) (string, error) {
	f.presigned = append(f.presigned, key)
	return fmt.Sprintf("https://fake.presigned/%s?expires=%dm", key, expiryMinutes), nil
}

func zipParams(t *testing.T, data map[string]interface{}) ActionParams {
	t.Helper()
	p := pubParams(t, data)
	p.ExecutionContext.StepName = "zip"
	return p
}

func TestZipDeliverableCutsUploadsVerifiesAndPresigns(t *testing.T) {
	index := []byte("<html>the site</html>")
	store := &zipFakeStore{pubFakeStore: pubFakeStore{objects: map[string][]byte{
		"example.com/index.html":   index,
		"example.com/css/site.css": []byte("body{}"),
		"example.com/about.html":   []byte("<html>about</html>"),
		"othersite.com/index.html": []byte("<html>not ours</html>"),
	}}}
	swapStore(t, func(context.Context, string, *zap.Logger) (publish.ObjectStore, error) { return store, nil })

	res, err := ZipDeliverableAction(context.Background(), zipParams(t, map[string]interface{}{"domain": "example.com"}))
	if err != nil {
		t.Fatalf("zip_deliverable: %v", err)
	}
	m := res.(map[string]interface{})
	if m["zipped"] != true {
		t.Fatalf("want zipped, got %v", m)
	}
	if m["files"] != 3 {
		t.Errorf("want 3 files from the tree (not the other site's), got %v", m["files"])
	}
	if m["size_alert"] != false {
		t.Errorf("small tree must not alert, got %v", m["size_alert"])
	}

	key, _ := m["zip_key"].(string)
	if !strings.HasPrefix(key, "deliverables/example.com/example.com-") || !strings.HasSuffix(key, ".zip") {
		t.Fatalf("zip key %q not under deliverables/<domain>/", key)
	}
	stored, ok := store.objects[key]
	if !ok {
		t.Fatalf("no object stored at %q", key)
	}
	if got := int64(len(stored)); got != m["zip_size_bytes"].(int64) {
		t.Errorf("stored %d bytes, result says %v", got, m["zip_size_bytes"])
	}

	// The acceptance itself: entry count == tree listing, index byte-equal.
	zr, err := zip.NewReader(bytes.NewReader(stored), int64(len(stored)))
	if err != nil {
		t.Fatalf("stored object is not a readable ZIP: %v", err)
	}
	if len(zr.File) != 3 {
		t.Errorf("archive holds %d entries, want 3", len(zr.File))
	}
	for _, zf := range zr.File {
		if zf.Name == "index.html" {
			rc, _ := zf.Open()
			got, _ := io.ReadAll(rc)
			rc.Close()
			if !bytes.Equal(got, index) {
				t.Errorf("archived index.html differs from origin")
			}
		}
	}

	url, _ := m["presigned_url"].(string)
	if !strings.Contains(url, key) {
		t.Errorf("presigned URL %q does not reference the zip key", url)
	}
	if len(store.presigned) != 1 || store.presigned[0] != key {
		t.Errorf("presign called for %v, want exactly [%s]", store.presigned, key)
	}
}

func TestZipDeliverableEmptyTreeSkipsWithoutArchiving(t *testing.T) {
	store := &zipFakeStore{pubFakeStore: pubFakeStore{objects: map[string][]byte{}}}
	swapStore(t, func(context.Context, string, *zap.Logger) (publish.ObjectStore, error) { return store, nil })

	res, err := ZipDeliverableAction(context.Background(), zipParams(t, map[string]interface{}{"domain": "example.com"}))
	if err != nil {
		t.Fatalf("empty tree must be a result, not an error: %v", err)
	}
	m := res.(map[string]interface{})
	if m["skipped"] != true || m["zipped"] != false {
		t.Errorf("want recorded skip, got %v", m)
	}
	if store.uploads != 0 {
		t.Errorf("empty tree performed %d uploads", store.uploads)
	}
}

func TestZipDeliverableOversizeAlertsAndStillCompletes(t *testing.T) {
	store := &zipFakeStore{pubFakeStore: pubFakeStore{objects: map[string][]byte{
		"example.com/index.html": []byte("<html>bigger than one byte</html>"),
	}}}
	swapStore(t, func(context.Context, string, *zap.Logger) (publish.ObjectStore, error) { return store, nil })

	// Induced oversize: threshold of 1 byte — the demand control for the
	// alert path. The cut must ALERT and complete, never truncate or fail.
	res, err := ZipDeliverableAction(context.Background(), zipParams(t, map[string]interface{}{
		"domain": "example.com", "size_alert_bytes": 1,
	}))
	if err != nil {
		t.Fatalf("oversize must alert, not fail: %v", err)
	}
	m := res.(map[string]interface{})
	if m["size_alert"] != true {
		t.Errorf("induced oversize did not raise size_alert: %v", m)
	}
	if m["zipped"] != true {
		t.Errorf("oversize cut must still complete: %v", m)
	}
	if store.uploads != 1 {
		t.Errorf("oversize cut performed %d uploads, want 1", store.uploads)
	}
}
