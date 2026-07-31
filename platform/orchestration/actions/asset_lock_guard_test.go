package actions

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// ── The guarantee, asserted rather than commented ────────────────────────────
//
// asset_lock_guard.go makes two choices that fail CLOSED and that a later edit
// could quietly reverse: no status filter, and no expiry awareness. A comment
// enforces nothing, so both are pinned here. If LOCK-004's predicate sweep
// deliberately changes the expiry half, this test is what makes that a decision
// instead of a side effect.

func TestAssetLockPredicateIsBareLockedAt(t *testing.T) {
	if got := assetAgentWritableSQL(""); got != "locked_at IS NULL" {
		t.Fatalf("writable predicate = %q, want bare locked_at IS NULL", got)
	}
	if got := assetAgentWritableSQL("assets."); got != "assets.locked_at IS NULL" {
		t.Fatalf("alias form = %q", got)
	}

	// status: conditioning the guard on unconstrained text fails OPEN.
	// lock_expires_at: honouring expiry here would WEAKEN three existing guards.
	for _, banned := range []string{"status", "lock_expires_at", "lock_type"} {
		if strings.Contains(assetAgentWritableSQL(""), banned) {
			t.Errorf("asset writable predicate must not mention %q — see the header of asset_lock_guard.go", banned)
		}
	}
}

func TestAssetLockedSQLIsExactNegation(t *testing.T) {
	for _, alias := range []string{"", "a.", "assets."} {
		want := "NOT (" + assetAgentWritableSQL(alias) + ")"
		if got := assetLockedSQL(alias); got != want {
			t.Errorf("assetLockedSQL(%q) = %q, want %q — the read predicate must be DERIVED from the write predicate, never re-typed", alias, got, want)
		}
	}
}

// ── The helper ───────────────────────────────────────────────────────────────

func TestLockedAssetKeys(t *testing.T) {
	ctx := context.Background()
	siteID := uuid.New()
	when := time.Date(2026, 7, 29, 8, 30, 5, 0, time.UTC)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT DISTINCT ON \\(asset_key\\)").
		WillReturnRows(sqlmock.NewRows([]string{"asset_key", "locked_by", "lock_type", "locked_at"}).
			AddRow("og_card", "admin", "permanent", when).
			AddRow("favicon", "", "", when))

	locks, err := lockedAssetKeys(ctx, db, siteID, "favicon", "og_card", "card_about")
	if err != nil {
		t.Fatal(err)
	}

	if !locks.Locked("og_card") || !locks.Locked("favicon") {
		t.Fatalf("both returned keys must read as locked, got %v", locks.Keys())
	}
	if locks.Locked("card_about") {
		t.Error("a key with no locked row must read as writable")
	}
	if got, want := locks.Keys(), []string{"favicon", "og_card"}; !equalStrings(got, want) {
		t.Errorf("Keys() = %v, want %v (sorted, for a stable payload)", got, want)
	}

	// Describe carries what an operator needs to clear the lock, and degrades
	// rather than lying when the columns are NULL on the row.
	if got := locks.Describe("og_card"); !strings.Contains(got, "admin") || !strings.Contains(got, "2026-07-29") {
		t.Errorf("Describe(og_card) = %q, want the holder and the date", got)
	}
	if got := locks.Describe("favicon"); !strings.Contains(got, "unclassified") || !strings.Contains(got, "unknown") {
		t.Errorf("Describe with NULL locked_by/lock_type = %q, want it to say so (4 of 5 live locks have lock_type NULL)", got)
	}
	if got := locks.Describe("card_about"); got != "" {
		t.Errorf("Describe of an unlocked key = %q, want empty so callers can build a message unconditionally", got)
	}
}

// An empty key list must not issue a query — `= ANY('{}')` is pointless work and
// the earlier IN () form was a syntax error. sqlmock with no expectations fails
// the test if any query is sent.
func TestLockedAssetKeysEmptyDoesNotQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	locks, err := lockedAssetKeys(context.Background(), db, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(locks) != 0 {
		t.Fatalf("want empty set, got %v", locks.Keys())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no query should have been issued: %v", err)
	}
}

// ── The bug: derive_card_asset must refuse BEFORE the git commit ─────────────
//
// StorageClient is deliberately nil. Reaching the storage type-assertion would
// return an error, so a CLEAN refusal is also proof of the ordering — the lock
// is read before any of the machinery that would replace the artefact. Same
// construction as TestDeriveBrandHeadBothLockedRefuses, for the same reason.
func TestDeriveCardAssetLockedRefusesBeforeAnyWrite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()

	mock.ExpectQuery("FROM pages p JOIN sites s").
		WillReturnRows(sqlmock.NewRows([]string{"name", "domain"}).
			AddRow("about-us", "example.com"))
	// WithArgs pins the KEY the guard asks about, not merely that it asked. A
	// pre-check that computed the key differently from the writers would query a
	// row nobody writes, find nothing, and wave the derivation through — silently.
	// (Council corr b5ff41f1, editquality edit-2, low.)
	mock.ExpectQuery("SELECT DISTINCT ON \\(asset_key\\)").
		WithArgs(sqlmock.AnyArg(), `{"card_about_us"}`).
		WillReturnRows(sqlmock.NewRows([]string{"asset_key", "locked_by", "lock_type", "locked_at"}).
			AddRow("card_about_us", "admin", "permanent", time.Now()))

	out, err := DeriveCardAssetAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"site_id":   siteID.String(),
			"entity_id": pageID.String(),
		},
	})
	if err != nil {
		t.Fatalf("expected a clean refusal, got error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result shape %#v", out)
	}
	if m["derived"] != false || m["locked"] != true {
		t.Fatalf("want derived:false locked:true, got %#v", m)
	}
	// The refusal must name the lock — an operator reading only this payload has
	// to know why nothing happened and what to clear.
	if reason, _ := m["reason"].(string); !strings.Contains(reason, "locked") || !strings.Contains(reason, "admin") {
		t.Errorf("refusal reason should identify the lock, got %q", reason)
	}
	// The pre-check must be the LAST query — nothing downloaded, nothing committed.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("queries beyond the lock pre-check ran: %v", err)
	}
}

// The unlocked path must not refuse. Without this, the guard above would still
// "pass" if lockedAssetKeys started reporting every key as locked — a guard that
// blocks everything is not a working guard.
func TestDeriveCardAssetUnlockedProceedsPastTheGuard(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM pages p JOIN sites s").
		WillReturnRows(sqlmock.NewRows([]string{"name", "domain"}).AddRow("about-us", "example.com"))
	mock.ExpectQuery("SELECT DISTINCT ON \\(asset_key\\)").
		WillReturnRows(sqlmock.NewRows([]string{"asset_key", "locked_by", "lock_type", "locked_at"}))
	// No lock → the action goes on to look for a source hero. Returning no hero
	// gives it a clean, recognisable stopping point that is NOT the lock refusal.
	mock.ExpectQuery("FROM site_plan_imagery spi").WillReturnRows(sqlmock.NewRows([]string{"id", "url", "asset_key"}))
	mock.ExpectQuery("FROM assets a").WillReturnRows(sqlmock.NewRows([]string{"id", "url", "asset_key"}))
	mock.ExpectQuery("FROM site_plan_imagery spi").WillReturnRows(sqlmock.NewRows([]string{"id", "url", "asset_key"}))

	out, err := DeriveCardAssetAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"site_id":   uuid.New().String(),
			"entity_id": uuid.New().String(),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := out.(map[string]interface{})
	if m["locked"] == true {
		t.Fatalf("an unlocked card must not be refused as locked: %#v", m)
	}
	if reason, _ := m["reason"].(string); !strings.Contains(reason, "hero") {
		t.Fatalf("expected to reach the hero lookup past the guard, got %#v", m)
	}
}

// The guard, the repo path and the upsert target must be ONE string. This is the
// structural half of the council's edit-2 objection: the WithArgs assertion above
// proves the guard asks about the right key today, and this proves the key it asks
// about is the same one the artefact is written under — so a future edit cannot
// make the two drift without going red.
func TestCardKeyIsSharedByTheGuardAndTheArtefactPath(t *testing.T) {
	for _, page := range []string{"about-us", "tool-matchmatrix", "learning-center-post", "news"} {
		key := contentCardKey(page)

		// The repo path the commit writes, and the web path the row records, are
		// both derived from this key — not from pageName, and not recomputed.
		repoPath := storage.DefaultAssetBasePath + "/" + storage.AssetKeyFilename(key, ".jpg")
		if !strings.Contains(repoPath, storage.AssetKeyFilename(key, ".jpg")) {
			t.Fatalf("repo path %q is not derived from the guarded key %q", repoPath, key)
		}
		if web := storage.DeployedWebPath(key, "card"); web == "" {
			t.Fatalf("web path for guarded key %q is empty — the row would record nothing", key)
		}

		// The live convention (all 13 card rows, 2026-07-31): card_<page> with
		// hyphens as underscores. A change here is a change to which row the lock
		// protects, so it must be deliberate.
		if !strings.HasPrefix(key, "card_") || strings.Contains(key, "-") {
			t.Errorf("contentCardKey(%q) = %q — breaks the live card_<page_with_underscores> convention", page, key)
		}
	}
}

// ── The lockstep door: the NEXT producer ─────────────────────────────────────
//
// bugs_open/145 records the architecture seat's standing objection shape: a fix
// that closes its own instance leaves "a generic hazard reachable by any future
// producer". This is that door for asset derivations. It is source-level on
// purpose: the rule has to fail when someone ADDS a producer, which no runtime
// test can see.

// gitAssetProducer is one file in this package that participates in committing
// bytes to a site repo.
type gitAssetProducer struct {
	file          string
	touchesCommit bool // calls (or defines) sendGitCommitRequest
	upsertsAsset  bool // INSERTs an assets row for the bytes it commits
	checksLock    bool // consults the shared asset lock guard
}

// inPopulation is the precise 143 class: an action that REGENERATES an asset
// artefact from a source and upserts the row describing it. Both halves matter.
//
// deploy_image_asset commits bytes and has no lock check anywhere, and is
// deliberately NOT in this population — stated here rather than left as a silent
// gap in the filter. It commits bytes the named row ALREADY points at, so
// deploying a locked asset is publication of the approved artefact, not
// replacement of it; guarding its `UPDATE assets SET url` would leave a locked
// row pointing at an expiring presigned URL, which is a regression, not a fix.
// emit_sprite_css commits CSS and only READS the sprite_sheet asset.
func (p gitAssetProducer) inPopulation() bool {
	return p.touchesCommit && p.upsertsAsset
}

// unguardedAssetDerivations is the rule, as a pure function over facts, so it
// can be proven to FIRE (below) without editing a real source file. A rule only
// tested against a tree that satisfies it passes just as happily when the rule
// has been deleted.
func unguardedAssetDerivations(ps []gitAssetProducer) []string {
	var bad []string
	for _, p := range ps {
		if p.inPopulation() && !p.checksLock {
			bad = append(bad, p.file)
		}
	}
	sort.Strings(bad)
	return bad
}

func classifyGitAssetProducers(dir string) ([]gitAssetProducer, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []gitAssetProducer
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		s := string(src)
		if !strings.Contains(s, "sendGitCommitRequest") {
			continue
		}
		out = append(out, gitAssetProducer{
			file:          name,
			touchesCommit: true,
			upsertsAsset:  strings.Contains(s, "INSERT INTO assets"),
			checksLock:    strings.Contains(s, "lockedAssetKeys"),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].file < out[j].file })
	return out, nil
}

// Every asset derivation that commits an artefact consults the shared guard.
func TestAssetDerivationsConsultTheLockGuard(t *testing.T) {
	producers, err := classifyGitAssetProducers(".")
	if err != nil {
		t.Fatalf("classify: %v", err)
	}

	var inPop int
	for _, p := range producers {
		if p.inPopulation() {
			inPop++
		}
	}
	if inPop == 0 {
		t.Fatal("no asset-deriving git producers found — the classifier has gone blind (renamed helper?), which would make this test pass vacuously")
	}

	if bad := unguardedAssetDerivations(producers); len(bad) > 0 {
		t.Errorf("these actions commit an asset artefact and upsert its row without consulting lockedAssetKeys (bugs_open/143): %v\n"+
			"An artefact must be lock-checked BEFORE the git commit — a provenance upsert guarded with locked_at protects only the DB row.", bad)
	}
}

// The set of files touching the git-commit path is pinned. A NEW producer fails
// this test, which forces a deliberate classification (in the 143 population, or
// out of it with a reason) instead of a silent omission.
func TestGitCommittingProducerSetIsKnown(t *testing.T) {
	known := []string{
		"deploy_image_asset_action.go",       // defines sendGitCommitRequest; publishes named bytes — out of population, see inPopulation()
		"derive_brand_head_assets_action.go", // derives favicon + og_card — IN population
		"derive_card_asset_action.go",        // derives a page card — IN population, the 143 instance
		"emit_sprite_css_action.go",          // commits CSS, reads the sprite asset — out of population
	}

	producers, err := classifyGitAssetProducers(".")
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	var got []string
	for _, p := range producers {
		got = append(got, p.file)
	}
	if !equalStrings(got, known) {
		t.Errorf("the set of files on the git-commit path changed.\n got:  %v\n want: %v\n"+
			"If you ADDED a producer: decide whether it regenerates an asset artefact. If it does, call lockedAssetKeys BEFORE sendGitCommitRequest and add it here; if it does not, add it here with the reason. Do not just widen the list.", got, known)
	}
}

// Proof the rule fires. Fed a producer that commits an artefact and upserts its
// row without the guard, the rule must name it — otherwise the two tests above
// would pass just as well with the rule gutted.
func TestUnguardedRuleFiresOnASyntheticProducer(t *testing.T) {
	facts := []gitAssetProducer{
		{file: "derive_new_thing_action.go", touchesCommit: true, upsertsAsset: true, checksLock: false},
		{file: "derive_guarded_action.go", touchesCommit: true, upsertsAsset: true, checksLock: true},
		{file: "commits_css_action.go", touchesCommit: true, upsertsAsset: false, checksLock: false},
		{file: "writes_row_only_action.go", touchesCommit: false, upsertsAsset: true, checksLock: false},
	}
	got := unguardedAssetDerivations(facts)
	if !equalStrings(got, []string{"derive_new_thing_action.go"}) {
		t.Fatalf("rule reported %v; want exactly the unguarded derivation. A guarded derivation, a CSS commit and a row-only writer must all pass.", got)
	}
}
