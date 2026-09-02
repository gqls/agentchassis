// FILE: platform/orchestration/actions/publish_site_action_test.go
//
// Pins the publish_site contract that the council/register entry states:
// NULL publish_target is a recorded no-op that never touches storage; no
// drift is a no-op; and published_hash is written ONLY after served-bytes
// acceptance — a publish whose served copy disagrees with origin leaves the
// drift standing for the next reconciler tick.
package actions

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/gqls/agentchassis/platform/publish"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

type pubFakeStore struct {
	objects map[string][]byte
	uploads int
}

func (f *pubFakeStore) ListObjects(_ context.Context, prefix string) ([]storage.ObjectInfo, error) {
	var out []storage.ObjectInfo
	for k, v := range f.objects {
		if strings.HasPrefix(k, prefix) {
			sum := md5.Sum(v)
			out = append(out, storage.ObjectInfo{Key: k, Size: int64(len(v)), ETag: hex.EncodeToString(sum[:])})
		}
	}
	return out, nil
}

func (f *pubFakeStore) Download(_ context.Context, key string) (io.ReadCloser, error) {
	v, ok := f.objects[key]
	if !ok {
		return nil, fmt.Errorf("no such key %q", key)
	}
	return io.NopCloser(bytes.NewReader(v)), nil
}

func (f *pubFakeStore) Upload(_ context.Context, key, _ string, body io.Reader) (string, error) {
	// Mirrors the real B2 S3 gateway: a non-seekable body fails live with
	// 411 MissingContentLength (first canary, 2026-08-15).
	if _, ok := body.(io.Seeker); !ok {
		return "", fmt.Errorf("fake gateway: upload body for %q is not seekable", key)
	}
	b, _ := io.ReadAll(body)
	f.objects[key] = b
	f.uploads++
	return "s3://fake/" + key, nil
}

func (f *pubFakeStore) Delete(_ context.Context, key string) error {
	delete(f.objects, key)
	return nil
}

// pubParams mirrors the production shape: the step config carries dotted
// paths, the values live in collected input_data (Strategy 0 resolution).
func pubParams(t *testing.T, data map[string]interface{}) ActionParams {
	t.Helper()
	config := map[string]interface{}{}
	for k := range data {
		config[k] = "input_data." + k
	}
	return ActionParams{
		Context:          context.Background(),
		ExecutionContext: &types.ExecutionContext{Action: "process", StepName: "publish"},
		StepConfig:       models.Step{Config: config},
		CollectedData:    map[string]interface{}{"input_data": data},
		Logger:           zap.NewNop(),
	}
}

func swapStore(t *testing.T, fn func(context.Context, string, *zap.Logger) (publish.ObjectStore, error)) {
	t.Helper()
	prev := newPortfolioStore
	newPortfolioStore = fn
	t.Cleanup(func() { newPortfolioStore = prev })
}

func swapServedFetch(t *testing.T, fn func(context.Context, string) ([]byte, int, error)) {
	t.Helper()
	prev := servedFetch
	servedFetch = fn
	t.Cleanup(func() { servedFetch = prev })
}

func siteRows(target, project, publishedHash string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "publish_target", "publish_project", "published_hash"}).
		AddRow("11111111-1111-1111-1111-111111111111", target, project, publishedHash)
}

func TestPublishSiteNullTargetSkipsWithoutTouchingStorage(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id::text").WillReturnRows(siteRows("", "", ""))

	swapStore(t, func(context.Context, string, *zap.Logger) (publish.ObjectStore, error) {
		t.Error("a NULL-target site must never construct a storage client — most reconciler passes are skips and need no credentials")
		return nil, fmt.Errorf("unreachable")
	})

	params := pubParams(t, map[string]interface{}{"domain": "example.com"})
	params.DB = db
	res, err := PublishSiteAction(context.Background(), params)
	if err != nil {
		t.Fatalf("NULL target must be a result, not an error: %v", err)
	}
	m := res.(map[string]interface{})
	if m["skipped"] != true || m["published"] != false {
		t.Errorf("want skipped no-op, got %v", m)
	}
	if !strings.Contains(m["reason"].(string), "opt-in") {
		t.Errorf("reason should say the seam is opt-in default OFF, got %q", m["reason"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestPublishSiteNoDriftSkipsWithoutUploading(t *testing.T) {
	store := &pubFakeStore{objects: map[string][]byte{"example.com/index.html": []byte("<html>x</html>")}}
	files, _ := publish.NewS3Source(store, "example.com").List(context.Background())
	currentHash := publish.TreeHash(files)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id::text").WillReturnRows(siteRows("b2worker", "canary.ugg2.com", currentHash))

	swapStore(t, func(context.Context, string, *zap.Logger) (publish.ObjectStore, error) { return store, nil })

	params := pubParams(t, map[string]interface{}{"domain": "example.com"})
	params.DB = db
	res, err := PublishSiteAction(context.Background(), params)
	if err != nil {
		t.Fatalf("no-drift must be a result, not an error: %v", err)
	}
	m := res.(map[string]interface{})
	if m["skipped"] != true || m["reason"] != "no drift" {
		t.Errorf("want no-drift skip, got %v", m)
	}
	if store.uploads != 0 {
		t.Errorf("no-drift pass performed %d uploads — the reconciler must be free when nothing changed", store.uploads)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestPublishSiteDriftPublishesAcceptsAndRecords(t *testing.T) {
	origin := []byte("<html>new version</html>")
	store := &pubFakeStore{objects: map[string][]byte{"example.com/index.html": origin}}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id::text").WillReturnRows(siteRows("b2worker", "canary.ugg2.com", "th1:stale"))
	mock.ExpectExec("UPDATE sites SET published_hash").
		WillReturnResult(sqlmock.NewResult(0, 1))

	swapStore(t, func(context.Context, string, *zap.Logger) (publish.ObjectStore, error) { return store, nil })
	var fetchedURL string
	swapServedFetch(t, func(_ context.Context, url string) ([]byte, int, error) {
		fetchedURL = url
		return origin, 200, nil
	})

	params := pubParams(t, map[string]interface{}{"domain": "example.com"})
	params.DB = db
	res, err := PublishSiteAction(context.Background(), params)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	m := res.(map[string]interface{})
	if m["published"] != true || m["accepted"] != true {
		t.Errorf("want published+accepted, got %v", m)
	}
	if !strings.HasPrefix(fetchedURL, "https://canary.ugg2.com/index.html?pub=") {
		t.Errorf("acceptance fetched %q — must hit the served copy with a cache-buster", fetchedURL)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("published_hash was not recorded after acceptance: %v", err)
	}
}

func TestPublishSiteSweepGetsThe404HalfOfAcceptance(t *testing.T) {
	origin := []byte("<html>home</html>")
	store := &pubFakeStore{objects: map[string][]byte{
		"example.com/index.html": origin,
		// The bugs_open/429 shape: a mirror copy whose source was retracted.
		"canary.ugg2.com/contact.html": []byte("<html>retracted</html>"),
	}}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id::text").WillReturnRows(siteRows("b2worker", "canary.ugg2.com", "th1:pre-sweep"))
	mock.ExpectExec("UPDATE sites SET published_hash").
		WillReturnResult(sqlmock.NewResult(0, 1))

	swapStore(t, func(context.Context, string, *zap.Logger) (publish.ObjectStore, error) { return store, nil })
	var sweptProbe string
	swapServedFetch(t, func(_ context.Context, url string) ([]byte, int, error) {
		if strings.Contains(url, "contact.html") {
			sweptProbe = url
			return []byte("not found"), 404, nil
		}
		return origin, 200, nil
	})

	params := pubParams(t, map[string]interface{}{"domain": "example.com"})
	params.DB = db
	res, err := PublishSiteAction(context.Background(), params)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	m := res.(map[string]interface{})
	if m["published"] != true || m["accepted"] != true {
		t.Errorf("want published+accepted, got %v", m)
	}
	if m["deleted"] != 1 {
		t.Errorf("result must record the sweep, got deleted=%v", m["deleted"])
	}
	if !strings.HasPrefix(sweptProbe, "https://canary.ugg2.com/contact.html?pub=") {
		t.Errorf("the swept key must be probed at the served copy with a cache-buster, fetched %q", sweptProbe)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("published_hash was not recorded after the pair passed: %v", err)
	}
}

func TestPublishSiteSweptKeyStillServingFailsAcceptance(t *testing.T) {
	origin := []byte("<html>home</html>")
	store := &pubFakeStore{objects: map[string][]byte{
		"example.com/index.html":       origin,
		"canary.ugg2.com/contact.html": []byte("<html>retracted</html>"),
	}}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// No ExpectExec: a published_hash write would fail ExpectationsWereMet —
	// the drift must stand so the next tick retries the (idempotent) sweep.
	mock.ExpectQuery("SELECT id::text").WillReturnRows(siteRows("b2worker", "canary.ugg2.com", "th1:pre-sweep"))

	swapStore(t, func(context.Context, string, *zap.Logger) (publish.ObjectStore, error) { return store, nil })
	swapServedFetch(t, func(_ context.Context, url string) ([]byte, int, error) {
		// Everything serves 200 — including the key the sweep deleted: the
		// hosted copy did not actually converge (an edge or origin lying).
		return origin, 200, nil
	})

	params := pubParams(t, map[string]interface{}{"domain": "example.com"})
	params.DB = db
	res, err := PublishSiteAction(context.Background(), params)
	if err != nil {
		t.Fatalf("acceptance failure must be a result, not an error: %v", err)
	}
	m := res.(map[string]interface{})
	if m["published"] != true || m["accepted"] != false {
		t.Errorf("want published-but-not-accepted, got %v", m)
	}
	if !strings.Contains(m["reason"].(string), "want 404") {
		t.Errorf("reason should name the 404 expectation, got %q", m["reason"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("published_hash must NOT be written when a swept key still serves: %v", err)
	}
}

func TestPublishSiteAcceptanceFailureLeavesDriftStanding(t *testing.T) {
	origin := []byte("<html>new version</html>")
	store := &pubFakeStore{objects: map[string][]byte{"example.com/index.html": origin}}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// No ExpectExec: any UPDATE would fail ExpectationsWereMet below.
	mock.ExpectQuery("SELECT id::text").WillReturnRows(siteRows("b2worker", "canary.ugg2.com", "th1:stale"))

	swapStore(t, func(context.Context, string, *zap.Logger) (publish.ObjectStore, error) { return store, nil })
	swapServedFetch(t, func(context.Context, string) ([]byte, int, error) {
		return []byte("<html>a stale edge copy</html>"), 200, nil
	})

	params := pubParams(t, map[string]interface{}{"domain": "example.com"})
	params.DB = db
	res, err := PublishSiteAction(context.Background(), params)
	if err != nil {
		t.Fatalf("acceptance failure must be a result, not an error: %v", err)
	}
	m := res.(map[string]interface{})
	if m["published"] != true || m["accepted"] != false {
		t.Errorf("want published-but-not-accepted, got %v", m)
	}
	if !strings.Contains(m["reason"].(string), "sha256") {
		t.Errorf("reason should carry the hash disagreement, got %q", m["reason"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("published_hash must NOT be written on acceptance failure: %v", err)
	}
}
