// FILE: platform/orchestration/actions/discovery_checks/check_page_content_divergence_test.go
//
// bugs_open/315 — pins page_content_divergence.
//
// THIS FILE CARRIES THE SAME UNUSUAL BURDEN AS check_asset_reference_404_test.go,
// and for the same reason. The check it tests has NO live positive: [MEASURED
// 2026-08-21] all 228 active pages then carrying a content_hash served bytes that
// hashed to exactly their stored fingerprint, across 12 domains. So nothing in
// production can demonstrate that the check bites, and "it found nothing" is
// precisely what a check that is silently broken also reports (016b §9: a gate's 0
// findings has two causes with opposite fixes).
//
// The substitute is an induced fault per branch. Every row below was RUN — the
// mutation applied to the real source, the named test executed, the source
// restored — on 2026-08-21. It is a measurement, not an intention:
//
//	mutation                                            test that catches it
//	--------------------------------------------------  ----------------------------------
//	drop the served-hash agreement on confirmation      OriginServingTwoBodiesFilesNothing
//	delete the contentIntentUnchanged re-read           RedeployedDuringThePassFilesNothing
//	fetch a fragment url instead of refusing it         FragmentURLIsNeverFetched
//	drop the cache-buster                               EveryRequestCarriesAUniqueCacheBuster
//	give the item a handler_agent                       NeverRoutesToAHandler
//	stop retracting on a match                          MatchingHashRetractsAndFilesNothing
//	raise the per-pass cap silently                     PerPassCapIsEnforced
//	drop status='active' from the query                 QueryPinsItsFourGuards
//	set Accept-Encoding by hand (hashing gzip bytes)    FetchServedPageHashesTheDecodedBody
//	drop the publish_target guard                       QueryPinsItsSixGuards
//	re-type the shared shipped predicate inline         DivergenceQueryUsesTheSharedShippedPredicate
//
// TWO GUARDS ARE DEFENCE-IN-DEPTH AND THIS FILE SAYS SO RATHER THAN PRETENDING
// OTHERWISE. The non-200 branch and the oversize branch each have a SECOND guard
// in series — the worker only fires a confirmation fetch for a judgeable 200 — so
// deleting either branch ALONE changes no outcome and no test can be made to fail
// against it. Deleting a branch AND its confirmation gate together does file a
// spurious item, and that pair is what the two tests catch (measured: both
// double-mutations fail their named test on the fetch-count assertion). A guard
// whose removal changes nothing is not load-bearing, and claiming a test proves it
// would be exactly the false comfort this table exists to avoid.
//
// FOUR OF THESE ROWS WERE FALSE WHEN FIRST WRITTEN, which is the reason the table
// now records a run rather than an intention. Three tests were passing against a
// mutated source because a LATER guard absorbed the fault (the unscripted intent
// re-read erroring; the confirmation gate), and one — the cap — because the test
// sized its fixture and its assertion from the very const it was testing, so
// raising the cap raised the expectation with it and the check silently widened.
// Each is fixed at the test, and the fix is noted where it lives.
//
// A guard no test can be made to fail against is not verified, it is decoration.

package discovery_checks

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"go.uber.org/zap"
)

// stubFetcher replaces the network with a scripted table of results and records
// every URL requested, so a test can assert what was NOT fetched as well as what
// was — the only way to prove the fragment refusal and the cap.
type stubFetcher struct {
	mu sync.Mutex
	// seq maps a path (url without the cache-buster) to successive results, so
	// the confirmation branch can be given a different answer the second time.
	seq   map[string][]divergenceProbeResult
	calls []string
}

func newStubFetcher() *stubFetcher {
	return &stubFetcher{seq: map[string][]divergenceProbeResult{}}
}

func (s *stubFetcher) install(t *testing.T) {
	t.Helper()
	prev := fetchServedPage
	fetchServedPage = func(_ context.Context, target string) divergenceProbeResult {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.calls = append(s.calls, target)
		key := target
		if i := strings.Index(target, "?"); i >= 0 {
			key = target[:i]
		}
		results := s.seq[key]
		if len(results) == 0 {
			return divergenceProbeResult{err: errors.New("stub: no scripted result for " + key)}
		}
		r := results[0]
		if len(results) > 1 {
			s.seq[key] = results[1:]
		}
		return r
	}
	t.Cleanup(func() { fetchServedPage = prev })
}

func (s *stubFetcher) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

func (s *stubFetcher) requested() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.calls))
	copy(out, s.calls)
	return out
}

const (
	divTestDomain = "robot-hands.com"
	// A real fingerprint pair from the lane's own proof, so the test data is the
	// shape production actually produces.
	divStoredHash = "e9d7090facaaddd3733d11885982979b9710d855df97297c062099bb5b09940b"
	divOtherHash  = "1111111111111111111111111111111111111111111111111111111111111111"
)

var divDeployedAt = time.Date(2026, 8, 21, 9, 0, 0, 0, time.UTC)

func divergenceRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "url", "content_hash", "build_status", "deployed_at", "age",
	})
}

func newDivergenceCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.MustParse("00ff3af5-dad8-4770-9f70-3edc267a3c92"),
		Pipeline:  "build",
		AgentType: "availability-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

// expectDomain scripts the site lookup every Run starts with.
func expectDomain(mock sqlmock.Sqlmock, siteID uuid.UUID) {
	mock.ExpectQuery(`FROM sites WHERE id`).WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow(divTestDomain))
}

// expectIntentReRead scripts contentIntentUnchanged. hash/at are what the row
// holds NOW — pass the same values the finder saw for "unchanged".
func expectIntentReRead(mock sqlmock.Sqlmock, pageID, hash string, at time.Time) {
	mock.ExpectQuery(`SELECT COALESCE\(content_hash`).WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"content_hash", "deployed_at"}).AddRow(hash, at))
}

// TestPageContentDivergence_ConfirmedMismatchFilesOneItem is bugs_open/315's own
// instance: the commit succeeded, the fingerprint was recorded, and the origin is
// serving something else entirely.
func TestPageContentDivergence_ConfirmedMismatchFilesOneItem(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	pageID := uuid.New()
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(pageID.String(), "product-detail", "/product-detail.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)
	expectIntentReRead(mock, pageID.String(), divStoredHash, divDeployedAt)

	fetcher := newStubFetcher()
	// Both fetches agree with each other and disagree with the stored hash.
	fetcher.seq["https://"+divTestDomain+"/product-detail.html"] = []divergenceProbeResult{
		{hash: divOtherHash, status: 200, bytes: 41234},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d", len(res.WorkItems))
	}
	if len(res.Resolved) != 0 {
		t.Errorf("a diverged page must retract nothing, got %d", len(res.Resolved))
	}
	wi := res.WorkItems[0]
	if wi.ItemType != "page_content_divergence" {
		t.Errorf("ItemType = %q", wi.ItemType)
	}
	if wi.Severity != "high" {
		t.Errorf("Severity = %q, want high — a visitor is served content we believe we replaced", wi.Severity)
	}
	if wi.PageID == nil || *wi.PageID != pageID {
		t.Errorf("PageID must be carried; triage starts from the page")
	}
	if want := "page_content_divergence:" + pageID.String(); wi.ItemKey != want {
		t.Errorf("ItemKey = %q, want %q", wi.ItemKey, want)
	}
	// BOTH hashes in full: the item is triaged by comparing them.
	if !strings.Contains(wi.SpecJSON, divStoredHash) || !strings.Contains(wi.SpecJSON, divOtherHash) {
		t.Errorf("spec must carry both full hashes, got %s", wi.SpecJSON)
	}
	// The confirming fetch must have happened.
	if n := fetcher.callCount(); n != 2 {
		t.Errorf("want 2 fetches (probe + confirmation), got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestPageContentDivergence_MatchingHashRetractsAndFilesNothing — the healthy
// case, which is 228 of 228 pages today. It must file nothing AND retract, because
// a match is a positive observation and that is the only thing CheckResult.Resolved
// permits.
func TestPageContentDivergence_MatchingHashRetractsAndFilesNothing(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	pageID := uuid.New()
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(pageID.String(), "product-detail", "/product-detail.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)

	fetcher := newStubFetcher()
	fetcher.seq["https://"+divTestDomain+"/product-detail.html"] = []divergenceProbeResult{
		{hash: divStoredHash, status: 200, bytes: 41234},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("a healthy page must file nothing, got %d", len(res.WorkItems))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("want 1 retraction on a positive observation, got %d", len(res.Resolved))
	}
	if res.Resolved[0].ItemType != "page_content_divergence" ||
		res.Resolved[0].ItemKey != "page_content_divergence:"+pageID.String() {
		t.Errorf("retraction must be narrow and keyed to the page: %+v", res.Resolved[0])
	}
	if res.Resolved[0].AllOfType {
		t.Errorf("AllOfType must stay false — one page's health says nothing about the site's")
	}
	if res.Resolved[0].Reason == "" {
		t.Errorf("Resolved.Reason is required by its contract")
	}
	// No confirmation fetch on a match, and no intent re-read either.
	if n := fetcher.callCount(); n != 1 {
		t.Errorf("want exactly 1 fetch for a matching page, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestPageContentDivergence_OriginServingTwoBodiesFilesNothing — a sync in
// progress can answer with one body then another. That is a moving target, not a
// divergence. Catches both the deletion of the confirming fetch and a
// confirmation that does not compare the two hashes.
func TestPageContentDivergence_OriginServingTwoBodiesFilesNothing(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	pageID := uuid.New().String()
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(pageID, "index", "/index.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)

	// The intent re-read is scripted to say "unchanged" DELIBERATELY, even though a
	// healthy run never reaches it. Without it, deleting the hash-agreement guard
	// below would still file nothing — the unscripted query would error and the
	// candidate would be discarded by the NEXT guard in series, so the mutation
	// would pass and this test would be proving nothing. Scripting it removes that
	// second guard from the path and leaves the comparison as the only thing
	// standing between a moving target and a work item.
	//
	// mock.ExpectationsWereMet is therefore NOT asserted here: in the healthy case
	// this expectation is deliberately never consumed.
	expectIntentReRead(mock, pageID, divStoredHash, divDeployedAt)

	fetcher := newStubFetcher()
	fetcher.seq["https://"+divTestDomain+"/index.html"] = []divergenceProbeResult{
		{hash: divOtherHash, status: 200, bytes: 100},
		{hash: "2222222222222222222222222222222222222222222222222222222222222222", status: 200, bytes: 101},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("two different bodies must file nothing, got %d", len(res.WorkItems))
	}
	if len(res.Resolved) != 0 {
		t.Errorf("and must retract nothing either, got %d", len(res.Resolved))
	}
}

// TestPageContentDivergence_RedeployedDuringThePassFilesNothing — the race that
// would libel a healthy page. We read hash H1, a deploy writes H2 with new bytes,
// we fetch the new bytes. Without the intent re-read this files a work item
// against a page that is perfectly fine.
func TestPageContentDivergence_RedeployedDuringThePassFilesNothing(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	pageID := uuid.New()
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(pageID.String(), "index", "/index.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)
	// The row now carries the NEW deploy's fingerprint — the served bytes we just
	// fetched are the correct ones for it.
	expectIntentReRead(mock, pageID.String(), divOtherHash, divDeployedAt.Add(time.Minute))

	fetcher := newStubFetcher()
	fetcher.seq["https://"+divTestDomain+"/index.html"] = []divergenceProbeResult{
		{hash: divOtherHash, status: 200, bytes: 100},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("a page redeployed mid-pass must file nothing, got %d", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestPageContentDivergence_Non200IsNotOurFinding — availability belongs to
// check_site_unreachable, which files on it. Judging it here would report one
// fault as two defects, and a 522 would arrive as a content divergence.
func TestPageContentDivergence_Non200IsNotOurFinding(t *testing.T) {
	for _, status := range []int{404, 403, 500, 522} {
		dctx, mock := newDivergenceCtx(t)
		expectDomain(mock, dctx.SiteID)
		mock.ExpectQuery(`FROM pages p`).WillReturnRows(
			divergenceRows().AddRow(uuid.New().String(), "index", "/index.html",
				divStoredHash, "deployed", divDeployedAt, int64(7200)),
		)
		fetcher := newStubFetcher()
		fetcher.seq["https://"+divTestDomain+"/index.html"] = []divergenceProbeResult{
			{status: status},
		}
		fetcher.install(t)

		res, err := (&PageContentDivergenceCheck{}).Run(dctx)
		if err != nil {
			t.Fatalf("status %d: Run: %v", status, err)
		}
		if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
			t.Errorf("status %d: want nothing filed and nothing retracted, got %d items / %d retractions",
				status, len(res.WorkItems), len(res.Resolved))
		}
		// A non-200 must be abandoned AT the status, not carried into the
		// confirmation path to be stopped by a later guard. Asserting the fetch
		// COUNT is what makes this test fail when the status branch is removed —
		// without it the confirmation's own status check absorbs the mutation and
		// this test passes while proving nothing.
		if n := fetcher.callCount(); n != 1 {
			t.Errorf("status %d: want exactly 1 fetch (no confirmation for an unjudgeable status), got %d",
				status, n)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("status %d: unmet expectations: %v", status, err)
		}
	}
}

// TestPageContentDivergence_TransportErrorFilesNothing — a transport failure is
// not a hash and is not a status. curl's `000` has been mistaken for an HTTP
// result on this estate before.
func TestPageContentDivergence_TransportErrorFilesNothing(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(uuid.New().String(), "index", "/index.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)
	fetcher := newStubFetcher()
	fetcher.seq["https://"+divTestDomain+"/index.html"] = []divergenceProbeResult{
		{err: errors.New("dial tcp: i/o timeout")},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Errorf("a transport error must file and retract nothing, got %d / %d",
			len(res.WorkItems), len(res.Resolved))
	}
}

// TestPageContentDivergence_OversizeBodyIsNeverHashed — a sha256 over a truncated
// body is a different, confidently wrong answer. It must skip, not convict.
func TestPageContentDivergence_OversizeBodyIsNeverHashed(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(uuid.New().String(), "index", "/index.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)
	fetcher := newStubFetcher()
	fetcher.seq["https://"+divTestDomain+"/index.html"] = []divergenceProbeResult{
		{status: 200, bytes: divergenceMaxBodyBytes + 1, oversize: true},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Errorf("an unhashable body must file and retract nothing, got %d / %d",
			len(res.WorkItems), len(res.Resolved))
	}
	// As with the non-200 case: assert it stopped HERE. Without the fetch count
	// the confirmation's own oversize check absorbs the mutation.
	if n := fetcher.callCount(); n != 1 {
		t.Errorf("want exactly 1 fetch for an unhashable body, got %d", n)
	}
}

// TestPageContentDivergence_FragmentURLIsNeverFetched — idea.uk has a live page
// row at "/tools.html#audience-check" while a DIFFERENT page owns "/tools.html".
// Fetching the first compares one page against the other's bytes. The assertion
// that matters is that NO request was made at all.
func TestPageContentDivergence_FragmentURLIsNeverFetched(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(uuid.New().String(), "tool-audience-check",
			"/tools.html#audience-check", divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)
	fetcher := newStubFetcher()
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n := fetcher.callCount(); n != 0 {
		t.Errorf("a fragment url must never be fetched; got %d requests: %v", n, fetcher.requested())
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("and must file nothing, got %d", len(res.WorkItems))
	}
}

// TestPageContentDivergence_EveryRequestCarriesAUniqueCacheBuster — without it an
// edge can answer both the probe and its confirmation from one cached copy, and
// the check would report a stale edge as a divergence for ever.
func TestPageContentDivergence_EveryRequestCarriesAUniqueCacheBuster(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	pageID := uuid.New()
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(pageID.String(), "index", "/index.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)
	expectIntentReRead(mock, pageID.String(), divStoredHash, divDeployedAt)

	fetcher := newStubFetcher()
	fetcher.seq["https://"+divTestDomain+"/index.html"] = []divergenceProbeResult{
		{hash: divOtherHash, status: 200, bytes: 10},
	}
	fetcher.install(t)

	if _, err := (&PageContentDivergenceCheck{}).Run(dctx); err != nil {
		t.Fatalf("Run: %v", err)
	}
	reqs := fetcher.requested()
	if len(reqs) != 2 {
		t.Fatalf("want probe + confirmation, got %d", len(reqs))
	}
	for _, r := range reqs {
		if !strings.Contains(r, "?cb=") {
			t.Errorf("request %q carries no cache-buster", r)
		}
	}
	if reqs[0] == reqs[1] {
		t.Errorf("the confirmation reused the probe's buster (%q) — an edge can serve both from one cache", reqs[0])
	}
}

// TestPageContentDivergence_NeverRoutesToAHandler — D5 is explicit that v1 is
// flag-only. Re-filing a rerender is the loop that already failed four times in
// this bug's own instance.
func TestPageContentDivergence_NeverRoutesToAHandler(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	pageID := uuid.New()
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(pageID.String(), "index", "/index.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)
	expectIntentReRead(mock, pageID.String(), divStoredHash, divDeployedAt)

	fetcher := newStubFetcher()
	fetcher.seq["https://"+divTestDomain+"/index.html"] = []divergenceProbeResult{
		{hash: divOtherHash, status: 200, bytes: 10},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 item, got %d", len(res.WorkItems))
	}
	if res.WorkItems[0].HandlerAgent != "" {
		t.Errorf("HandlerAgent = %q, want empty — an item with a handler that cannot repair the "+
			"delivery boundary is marked blocked at claim", res.WorkItems[0].HandlerAgent)
	}
}

// TestPageContentDivergence_PerPassCapIsEnforced — webdesign.co.uk carries 124
// hashed pages today, so the cap DOES bite on the busiest site. A silent cap reads
// as "every page was checked".
func TestPageContentDivergence_PerPassCapIsEnforced(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	expectDomain(mock, dctx.SiteID)
	// THE LITERAL 60 IS DELIBERATE AND MUST NOT BECOME divergenceMaxPagesPerPass.
	// Sizing the fixture and the assertion from the const under test makes the test
	// self-referential: raising the const raises the expectation with it, the cap
	// silently widens, and this test goes on passing. Proven — that mutation passed
	// until the literal went in. Changing the cap should require editing this line,
	// which is the point.
	const wantCap = 60
	const fixtureRows = wantCap + 5
	rows := divergenceRows()
	for i := 0; i < fixtureRows; i++ {
		rows.AddRow(uuid.New().String(), "p", "/p"+string(rune('a'+i%26))+".html",
			divStoredHash, "deployed", divDeployedAt, int64(7200))
	}
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(rows)

	fetcher := newStubFetcher()
	fetcher.install(t) // every fetch errors: we are counting requests, not judging

	if _, err := (&PageContentDivergenceCheck{}).Run(dctx); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n := fetcher.callCount(); n != wantCap {
		t.Errorf("fetched %d of %d pages, want exactly the cap %d", n, fixtureRows, wantCap)
	}
	if divergenceMaxPagesPerPass != wantCap {
		t.Errorf("the cap moved to %d; this test's fixture and assertion are pinned to %d — "+
			"change them deliberately, together", divergenceMaxPagesPerPass, wantCap)
	}
}

// TestPageContentDivergence_QueryPinsItsSixGuards — each is a one-line deletion
// that leaves every behavioural test above green, because sqlmock returns the rows
// a test hands it whatever the WHERE clause says. The query text is the only place
// they can be pinned.
//
// Two of the six arrived from council objections (corr be85a6d3) rather than from
// the original design, and both close a stale-fingerprint route: the publish_site
// seam writes neither content_hash nor deployed_at, and the shared shipped
// predicate is the platform's line rather than this check's to redraw.
func TestPageContentDivergence_QueryPinsItsSixGuards(t *testing.T) {
	for _, guard := range []struct{ needle, why string }{
		{"p.status = 'active'", "retracted and archived pages keep deployed_at by design and are not served"},
		{"p.content_hash IS NOT NULL", "without a fingerprint there is nothing to compare the wire against"},
		{"p.deployed_at IS NOT NULL", "the shared shipped predicate (queryresolve.DeployedPageEligibilitySQL) must survive concatenation"},
		{"make_interval(secs => $2)", "the settle window keeps the check off deliveries still in flight"},
		{"s.publish_target IS NULL", "a site on the publish_site seam has no fingerprint authority here — publish_site writes neither content_hash nor deployed_at, so its pages carry stale ones"},
		{"JOIN sites s", "the publish_target guard needs the sites row; losing the join silently drops that guard with it"},
	} {
		if !strings.Contains(divergenceCandidatesQuery, guard.needle) {
			t.Errorf("divergenceCandidatesQuery lost %q — %s", guard.needle, guard.why)
		}
	}
}

// TestDivergenceQueryUsesTheSharedShippedPredicate pins the REUSE itself, not just
// the resulting text. Re-typing `AND p.deployed_at IS NOT NULL` inline would keep
// the test above green while silently forking the platform's shipped definition —
// which is the drift the council seat objected to in the first place.
func TestDivergenceQueryUsesTheSharedShippedPredicate(t *testing.T) {
	if !strings.Contains(divergenceCandidatesQuery, queryresolve.DeployedPageEligibilitySQL) {
		t.Errorf("the query no longer contains queryresolve.DeployedPageEligibilitySQL verbatim — "+
			"it has been re-typed inline and will not follow the shared definition when it changes.\nquery:\n%s",
			divergenceCandidatesQuery)
	}
}

// TestFetchServedPageHashesTheDecodedBody exercises the REAL fetcher against a
// real server, because the seam the other tests use cannot catch an encoding
// mistake — and there is a live one available: setting Accept-Encoding by hand
// disables Go's transparent gunzip, so the check would hash the COMPRESSED stream
// and report every gzip-serving origin (which is all of them) as diverged.
func TestFetchServedPageHashesTheDecodedBody(t *testing.T) {
	body := []byte("<!doctype html><html><body>robot-hands product detail</body></html>")
	want := sha256.Sum256(body)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "cb=") {
			t.Errorf("server saw no cache-buster: %q", r.URL.String())
		}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			t.Errorf("Go's transport should have offered gzip itself; got %q", r.Header.Get("Accept-Encoding"))
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "text/html")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		_, _ = gz.Write(body)
	}))
	defer srv.Close()

	got := fetchServedPage(context.Background(), cacheBust(srv.URL+"/product-detail.html"))
	if got.err != nil {
		t.Fatalf("fetch: %v", got.err)
	}
	if got.status != 200 {
		t.Fatalf("status = %d", got.status)
	}
	if got.hash != hex.EncodeToString(want[:]) {
		t.Errorf("hash = %s, want %s — the fetcher hashed the compressed stream, not the file",
			got.hash, hex.EncodeToString(want[:]))
	}
	if got.bytes != int64(len(body)) {
		t.Errorf("bytes = %d, want the DECODED length %d", got.bytes, len(body))
	}
}

// TestFetchServedPageDoesNotHashANon200Body — an error page is not this page's
// bytes, and hashing it would produce a confident divergence finding for every
// 404 on the estate.
func TestFetchServedPageDoesNotHashANon200Body(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("<html>404</html>"))
	}))
	defer srv.Close()

	got := fetchServedPage(context.Background(), cacheBust(srv.URL+"/gone.html"))
	if got.err != nil {
		t.Fatalf("fetch: %v", got.err)
	}
	if got.status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", got.status)
	}
	if got.hash != "" {
		t.Errorf("a non-200 body must not be hashed, got %q", got.hash)
	}
}

// TestCacheBustIsUniquePerCall — two calls that produce the same buster let an
// edge answer a probe and its own confirmation from one cached copy.
func TestCacheBustIsUniquePerCall(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		got := cacheBust("https://example.com/x.html")
		if !strings.Contains(got, "?cb=") {
			t.Fatalf("no buster in %q", got)
		}
		if seen[got] {
			t.Fatalf("cacheBust repeated itself after %d calls: %q", i, got)
		}
		seen[got] = true
	}
}
