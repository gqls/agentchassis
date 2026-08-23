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
//	move the settle window off 60 min                   DivergenceSettleWindowIsPinnedAndReachesTheQuery
//	stop binding the window to the query ($2)           DivergenceSettleWindowIsPinnedAndReachesTheQuery
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
	seq     map[string][]divergenceProbeResult
	calls   []string
	accepts []string
}

func newStubFetcher() *stubFetcher {
	return &stubFetcher{seq: map[string][]divergenceProbeResult{}}
}

func (s *stubFetcher) install(t *testing.T) {
	t.Helper()
	prev := fetchServedPage
	fetchServedPage = func(_ context.Context, target, accept string) divergenceProbeResult {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.calls = append(s.calls, target)
		s.accepts = append(s.accepts, accept)
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

func (s *stubFetcher) acceptHeaders() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.accepts...)
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
	// The confirming fetch AND the raw-object probe must both have happened.
	if n := fetcher.callCount(); n != 3 {
		t.Errorf("want 3 fetches (probe + confirmation + raw-object probe), got %d", n)
	}
	// The headers are the point of the third fetch: two as a browser, then one
	// asking for the object itself. A raw probe that repeated the HTML Accept
	// would return the injected body and could never exonerate anything.
	if got := fetcher.acceptHeaders(); len(got) != 3 ||
		got[0] != divergenceAcceptHTML || got[1] != divergenceAcceptHTML || got[2] != divergenceAcceptRaw {
		t.Errorf("Accept headers = %v, want [%s %s %s]",
			got, divergenceAcceptHTML, divergenceAcceptHTML, divergenceAcceptRaw)
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
		{hash: divOtherHash, status: 200, bytes: 41000},
		{hash: "2222222222222222222222222222222222222222222222222222222222222222", status: 200, bytes: 41001},
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
		{hash: divOtherHash, status: 200, bytes: 41000},
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
		{hash: divOtherHash, status: 200, bytes: 41000},
	}
	fetcher.install(t)

	if _, err := (&PageContentDivergenceCheck{}).Run(dctx); err != nil {
		t.Fatalf("Run: %v", err)
	}
	reqs := fetcher.requested()
	if len(reqs) != 3 {
		t.Fatalf("want probe + confirmation + raw-object probe, got %d", len(reqs))
	}
	for _, r := range reqs {
		if !strings.Contains(r, "?cb=") {
			t.Errorf("request %q carries no cache-buster", r)
		}
	}
	// All three must be distinct, the raw probe included: it is fetched to be
	// compared against the fingerprint, so an edge answering it from the cache
	// of an earlier fetch would make it agree with whatever it just served.
	if reqs[0] == reqs[1] || reqs[1] == reqs[2] || reqs[0] == reqs[2] {
		t.Errorf("two fetches shared a cache-buster — an edge can serve both from one cache: %v", reqs)
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
		{hash: divOtherHash, status: 200, bytes: 41000},
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

// TestDivergenceSettleWindowIsPinnedAndReachesTheQuery does two things a reader
// might think are one.
//
// (1) It pins the LITERAL 60 minutes. The value is a safety constant traded off
// against measured delivery behaviour (see the const's comment: a ~17-minute
// propagation tail, a 1h07 case, and a 9-hour true positive), so moving it should
// require editing this line and reading that evidence — not be a one-character
// change nothing notices. Deliberately a literal and NOT `divergenceSettleWindow`:
// sizing an assertion from the constant under test is self-referential and cannot
// fail, which is exactly how the per-pass cap test passed against a mutated cap
// until it was pinned to a literal.
//
// (2) It asserts the value actually REACHES the query as $2. A const nothing
// passes is a documented intention, not a behaviour — and the finder is the only
// place that binds it.
func TestDivergenceSettleWindowIsPinnedAndReachesTheQuery(t *testing.T) {
	const wantMinutes = 60
	if divergenceSettleWindow != wantMinutes*time.Minute {
		t.Errorf("divergenceSettleWindow = %v, want %dm — if this move is deliberate, read the "+
			"measurements in the const's comment and change this literal with it",
			divergenceSettleWindow, wantMinutes)
	}

	dctx, mock := newDivergenceCtx(t)
	expectDomain(mock, dctx.SiteID)
	// WithArgs pins the BOUND VALUE: site id, then the window in seconds.
	mock.ExpectQuery(`FROM pages p`).
		WithArgs(dctx.SiteID, float64(wantMinutes*60)).
		WillReturnRows(divergenceRows())

	if _, err := (&PageContentDivergenceCheck{}).Run(dctx); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the settle window did not reach the query as $2: %v", err)
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

	got := fetchServedPage(context.Background(), cacheBust(srv.URL+"/product-detail.html"), divergenceAcceptHTML)
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

	got := fetchServedPage(context.Background(), cacheBust(srv.URL+"/gone.html"), divergenceAcceptRaw)
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

// TestPageContentDivergence_EdgeInjectedBodyFilesNothing is the false positive
// this check actually produced in production, reduced to a unit test.
//
// vetcomparison.uk had Cloudflare Web Analytics enabled on its zone. Cloudflare
// injects a ~359-byte beacon <script> into anything it treats as browser HTML,
// so the body on the wire could never hash to the fingerprint — while the object
// in the bucket was byte-perfect and every visitor was served the current page.
// The check flagged it on six consecutive passes and a whole session was spent
// believing a live customer-facing fault existed. It did not.
//
// PAIRED CONTROL: this is TestPageContentDivergence_ConfirmedMismatchFilesOneItem
// with ONE value changed — the raw probe's hash. There it repeats the served
// hash and an item files; here it equals the stored hash and nothing files. Two
// opposite outcomes from one differing input is what makes this a test of the
// guard rather than a test of the fixture.
//
// It deliberately does NOT assert ExpectationsWereMet: the intent re-read IS
// scripted so that deleting the guard lets the candidate run all the way to a
// filed item and fail this test. Without that scripted row a deleted guard would
// merely error on an unexpected query, discard the candidate, and pass — the
// test would be green for the wrong reason.
func TestPageContentDivergence_EdgeInjectedBodyFilesNothing(t *testing.T) {
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
		// Two browser-Accept fetches agree: the injected body.
		{hash: divOtherHash, status: 200, bytes: 44857},
		{hash: divOtherHash, status: 200, bytes: 44857},
		// The raw-object probe gets the object itself, and it is exactly what we
		// stamped. The delivery worked; the edge added to it afterwards.
		{hash: divStoredHash, status: 200, bytes: 44498},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("an edge-injected body is not a delivery fault; want 0 work items, got %d", len(res.WorkItems))
	}
	if len(res.Findings) != 0 {
		t.Errorf("want 0 findings, got %d", len(res.Findings))
	}
	// Deliberately does not retract either. The origin object is right, but this
	// check observes the wire, and an edge that can add bytes could in principle
	// serve a stale body to browsers while answering the raw probe correctly.
	// Skipping asserts nothing; retracting would assert health we did not see.
	if len(res.Resolved) != 0 {
		t.Errorf("want 0 retractions from a body we could not judge, got %d", len(res.Resolved))
	}
	if n := fetcher.callCount(); n != 3 {
		t.Errorf("want 3 fetches, got %d", n)
	}
}

// TestPageContentDivergence_UnusableRawProbeFilesNothing — the raw probe is the
// exonerating evidence, so failing to READ it is not the same as failing to be
// exonerated by it. With no information either way the candidate is discarded:
// the sweep re-probes the whole fleet every few hours, so a missed pass costs
// one cycle, while a false item costs a human a triage.
//
// Same paired-control construction as the test above: only the raw probe's
// outcome differs from the filing case.
func TestPageContentDivergence_UnusableRawProbeFilesNothing(t *testing.T) {
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
		{hash: divOtherHash, status: 200, bytes: 44857},
		{hash: divOtherHash, status: 200, bytes: 44857},
		{err: errors.New("connection reset by peer")},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("want 0 work items when the raw probe could not be read, got %d", len(res.WorkItems))
	}
	if len(res.Resolved) != 0 {
		t.Errorf("want 0 retractions, got %d", len(res.Resolved))
	}
}

// TestPageContentDivergence_EmptyTwoHundredFilesNothing — PLAN D10. An edge can
// answer 200 with a zero-length body, and this lane OBSERVED exactly that during
// its 2026-08-22 watch (sha256 e3b0c44298fc…, the hash of the empty string).
//
// The reason it needs its own guard is that it defeats the one that looks like it
// should catch it: an empty body hashes STABLY, so the confirmation fetch AGREES
// with the first and the moving-target guard passes it straight through. Two
// consecutive empty 200s would file a work item against a healthy page.
//
// Scripts three results so that DELETING the floor lets the candidate run all the
// way to a filed item — otherwise this would pass green for the wrong reason.
func TestPageContentDivergence_EmptyTwoHundredFilesNothing(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	pageID := uuid.New()
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(pageID.String(), "index", "/index.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)
	expectIntentReRead(mock, pageID.String(), divStoredHash, divDeployedAt)

	const emptyBodyHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	fetcher := newStubFetcher()
	fetcher.seq["https://"+divTestDomain+"/index.html"] = []divergenceProbeResult{
		{hash: emptyBodyHash, status: 200, bytes: 0},
		{hash: emptyBodyHash, status: 200, bytes: 0},
		{hash: emptyBodyHash, status: 200, bytes: 0},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("an empty 200 is not a page and cannot convict one; want 0 work items, got %d", len(res.WorkItems))
	}
	if len(res.Findings) != 0 {
		t.Errorf("want 0 findings, got %d", len(res.Findings))
	}
	if len(res.Resolved) != 0 {
		t.Errorf("an unjudgeable body must not retract either, got %d", len(res.Resolved))
	}
}

// TestPageContentDivergence_SmallBodyThatMatchesStillRetracts pins the floor's
// PLACEMENT, which is the part of it that could silently go wrong.
//
// The floor sits inside the candidate branch, AFTER the hash-match arm. If it
// were hoisted above that arm — the obvious "skip unjudgeable bodies early"
// refactor — then a small page that is perfectly healthy would stop retracting,
// and its open item would never close. That failure is invisible in production
// (a retraction that does not happen looks like a page that has not recovered),
// so it is pinned here instead.
func TestPageContentDivergence_SmallBodyThatMatchesStillRetracts(t *testing.T) {
	dctx, mock := newDivergenceCtx(t)
	pageID := uuid.New()
	expectDomain(mock, dctx.SiteID)
	mock.ExpectQuery(`FROM pages p`).WillReturnRows(
		divergenceRows().AddRow(pageID.String(), "tiny", "/tiny.html",
			divStoredHash, "deployed", divDeployedAt, int64(7200)),
	)

	fetcher := newStubFetcher()
	// Below the floor, and it MATCHES. Health does not need to clear a size bar.
	fetcher.seq["https://"+divTestDomain+"/tiny.html"] = []divergenceProbeResult{
		{hash: divStoredHash, status: 200, bytes: 12},
	}
	fetcher.install(t)

	res, err := (&PageContentDivergenceCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("a matching hash is a positive observation whatever the body size; want 1 retraction, got %d", len(res.Resolved))
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("want 0 work items, got %d", len(res.WorkItems))
	}
	// One fetch: a match is not a candidate, so nothing is confirmed or re-probed.
	if n := fetcher.callCount(); n != 1 {
		t.Errorf("want 1 fetch for a matching page, got %d", n)
	}
}
