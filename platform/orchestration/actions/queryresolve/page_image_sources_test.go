// FILE: platform/orchestration/actions/queryresolve/page_image_sources_test.go
//
// bugs_open/384, generalised 2026-08-25 by RFC_052. sourceDependencies is a
// DECLARATION about queryHandlers: "this base reads that store". A declaration
// beside a map drifts the moment someone adds a resolver, so this test does not
// read the declaration back — it DRIVES every registered handler against a
// recording sqlmock and checks which store each one's SQL actually touches,
// then requires the declared and observed sets to be identical in both
// directions, for every dependency class.
//
// Why not a source scan: a scan would make the comments in queryresolve.go
// load-bearing (the a-source-scanning-test-makes-comments-load-bearing trap).
// Recording the SQL the handler executes is the behaviour, not a description
// of it.
//
// Every handler is called with a representative argument and an empty result
// set; handlers that error on ErrNoRows still execute their SQL first, which is
// all this test needs. A handler that returned before querying would be
// recorded as "reads nothing", so the positive controls also prove the harness
// sees SQL at all.

package queryresolve

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// dependencyNeedles maps each declared dependency to a fragment that appears in
// the SQL of a resolver reading that store and in no other resolver's SQL.
//
// These must stay DISCRIMINATING. `pages` and `assets` are read by nearly every
// resolver here, so neither could ever serve as a needle — which is exactly why
// the page-image needle is the entity-linked CARD PREDICATE rather than a table
// name. If a needle stops being unique the test fails loudly (a resolver gets
// classified into a class it does not read), which is the safe direction.
var dependencyNeedles = map[SourceDependency]string{
	DepPageCardImages:    "ca.purpose = 'card'",
	DepFeedItems:         "content_feed_items",
	DepDirectoryEntities: "directory_entities",
	DepBusinessIntel:     "business_intel",
	DepProducts:          "FROM products",
	DepEvidenceBase:      "evidence_base",
}

// representativeArgs gives arg-taking handlers something to query with; a
// handler that refuses an empty arg before querying would otherwise be
// misclassified as reading nothing.
var representativeArgs = map[string]string{
	"pages_where_type":    "tool",
	"pages_under_section": "guides",
	"section_index_for":   "guide",
	"products":            "gripper",
}

func recordedSQLFor(t *testing.T, base string) []string {
	t.Helper()
	var seen []string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(_, actual string) error {
		seen = append(seen, actual)
		return nil
	})))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// ⚠ A GATED RESOLVER MUST BE UNGATED, OR THIS HARNESS IS BLIND TO IT.
	// resolveBusinessDirectory looks up its exporter config FIRST and returns an
	// error when there is none — deliberately, so a misconfiguration cannot look
	// like "zero eligible businesses" (bugs_open/206). Under a mock that answers
	// every query with an empty row set, that early return means the resolver
	// NEVER ISSUES its business_intel query, and this test would record it as
	// "reads nothing" and then report the (correct) declaration as stale. That
	// happened on the first run of the generalised test, 2026-08-25, and the
	// answer is to feed the gate rather than to weaken the assertion: an
	// exclusion list would have hidden a real read, which is the one thing this
	// test exists to make impossible.
	if base == "business_directory" {
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"vertical", "business_type_ilike"}).AddRow("dentists", "dentist%"))
	}

	// Enough empty answers for any resolver's query chain; unused ones are fine
	// because ExpectationsWereMet is deliberately not asserted here.
	for i := 0; i < 8; i++ {
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{}))
	}
	_, _ = queryHandlers[base](context.Background(), db, uuid.New(), representativeArgs[base], 0, zap.NewNop())
	return seen
}

// TestSourceDependenciesMatchTheResolvers is the lockstep guard, generalised.
// Its predecessor asked only "does this resolver read the card join"; it now
// asks the same question of every declared store, so a producer of ANY of them
// can trust the declaration.
func TestSourceDependenciesMatchTheResolvers(t *testing.T) {
	// The card needle must still be inside the fragment it claims to detect,
	// or this test cannot see the join at all and would pass vacuously.
	if !strings.Contains(PageImageJoinsSQL, dependencyNeedles[DepPageCardImages]) {
		t.Fatalf("PageImageJoinsSQL no longer contains %q — update the needle to a fragment unique to the card join, or this test cannot see the join at all",
			dependencyNeedles[DepPageCardImages])
	}

	// observed[base][dep] = the resolver's SQL touched that store.
	observed := map[string]map[SourceDependency]bool{}
	for base := range queryHandlers {
		sqlSeen := recordedSQLFor(t, base)
		if len(sqlSeen) == 0 {
			t.Errorf("handler %q executed no SQL under the harness — give it a representative arg (representativeArgs) so it can be classified", base)
			continue
		}
		observed[base] = map[SourceDependency]bool{}
		for _, q := range sqlSeen {
			for dep, needle := range dependencyNeedles {
				if strings.Contains(q, needle) {
					observed[base][dep] = true
				}
			}
		}
	}

	// Both directions, per base. An undeclared read is bugs_open/384 one level
	// up: the store's consumers would never be told it changed. A stale
	// declaration is the opposite failure — consumers re-rendered for a change
	// that cannot reach them.
	for base, deps := range observed {
		declared := sourceDependencies[base]
		for dep := range deps {
			if _, ok := declared[dep]; !ok {
				t.Errorf("resolver %q reads %s (SQL contains %q) but does NOT declare it in sourceDependencies — its consumers will never be re-resolved when that store changes",
					base, dep, dependencyNeedles[dep])
			}
		}
		for dep := range declared {
			if !deps[dep] {
				t.Errorf("sourceDependencies declares %q reads %s, but its resolver issued no SQL containing %q — stale declaration, or the read moved",
					base, dep, dependencyNeedles[dep])
			}
		}
	}
}

// TestEveryRegisteredBaseDeclaresItsDependencies is the completeness half, and
// it is the one the old boolean map could not express. A base absent from
// sourceDependencies is INVISIBLE to every producer — it would silently answer
// "reads nothing" for every dependency, and no existing test would notice. A
// base with genuinely nothing to declare must say so with an empty map, so
// "nobody thought about it" and "there is nothing to think about" stay
// distinguishable.
func TestEveryRegisteredBaseDeclaresItsDependencies(t *testing.T) {
	for base := range queryHandlers {
		if _, ok := sourceDependencies[base]; !ok {
			t.Errorf("query base %q is registered in queryHandlers but absent from sourceDependencies — declare what it reads, or declare an EMPTY map if it has no notifiable dependency (as section_index_for does)", base)
		}
	}
	for base := range sourceDependencies {
		if _, ok := queryHandlers[base]; !ok {
			t.Errorf("sourceDependencies declares %q, which is not a registered query base", base)
		}
	}
}

// The page-image wrappers must keep behaving exactly as they did before the
// generalisation — these callers were not touched and must not change.
func TestSourceReadsPageImagesNormalisesLikeResolve(t *testing.T) {
	for name, want := range map[string]bool{
		"blog_posts":                 true,
		"  Blog_Posts ":              true,
		"pages_where_type:tool":      true,
		"pages_where_type:blog-post": true,
		"pages_under_section:guides": true,
		"section_index_for:tool":     false,
		"news_archive":               false,
		"latest_news":                false,
		"business_directory":         false,
		"model_directory_full":       false,
		"":                           false,
		"not_a_query":                false,
	} {
		if got := SourceReadsPageImages(name); got != want {
			t.Errorf("SourceReadsPageImages(%q) = %v, want %v", name, got, want)
		}
	}
	if got := PageImageSources(); strings.Join(got, ",") != "blog_posts,pages_under_section,pages_where_type" {
		t.Errorf("PageImageSources() = %v — sorted declared bases expected", got)
	}
}

// SourceReads is the general form and must normalise identically — the wrapper
// above delegates to it, so a divergence here would be invisible there.
func TestSourceReadsAnswersPerDependency(t *testing.T) {
	cases := []struct {
		name string
		dep  SourceDependency
		want bool
	}{
		{"latest_news", DepFeedItems, true},
		{"  News_Archive ", DepFeedItems, true},
		{"latest_news", DepPageCardImages, false},
		{"blog_posts", DepFeedItems, false},
		{"model_directory_full", DepDirectoryEntities, true},
		// business_directory reads business_intel, NOT directory_entities.
		// Declaring them as one class would tell the wrong consumers on every
		// publish, which is why they are separate constants.
		{"business_directory", DepDirectoryEntities, false},
		{"business_directory", DepBusinessIntel, true},
		{"model_directory", DepBusinessIntel, false},
		{"section_index_for:tool", DepPageCardImages, false},
		{"not_a_query", DepFeedItems, false},
		{"", DepFeedItems, false},
	}
	for _, c := range cases {
		if got := SourceReads(c.name, c.dep); got != c.want {
			t.Errorf("SourceReads(%q, %s) = %v, want %v", c.name, c.dep, got, c.want)
		}
	}
}

// The item-key list is what makes ConsumerPages apply — or NOT apply — a
// template filter, and the two cases are not symmetric. An empty list means
// "this dependency governs the whole item set, so every consumer is affected";
// if that were ever read as "match no keys" the filter would exclude every news
// and directory consumer and both producers would silently stop notifying.
func TestDependencyItemKeysDistinguishNamedKeysFromWholeSet(t *testing.T) {
	if got := strings.Join(DependencyItemKeys(DepPageCardImages), ","); got != "image" {
		t.Errorf("DependencyItemKeys(DepPageCardImages) = %q, want \"image\"", got)
	}
	for _, dep := range []SourceDependency{DepFeedItems, DepDirectoryEntities, DepBusinessIntel, DepProducts} {
		if keys := DependencyItemKeys(dep); len(keys) != 0 {
			t.Errorf("DependencyItemKeys(%s) = %v, want empty — a whole-item-set dependency must declare NO keys, or ConsumerPages will filter its consumers out by template", dep, keys)
		}
	}
}

// The SQL consequence of the line above, asserted directly: a named-key
// dependency gets the renders-key filter, a whole-set dependency must not.
func TestConsumerSQLAppliesTheTemplateFilterOnlyForNamedKeys(t *testing.T) {
	imageSQL := consumerSQL(DepPageCardImages)
	if !strings.Contains(imageSQL, `cc.html_template ~* '\.(image)\y'`) {
		t.Errorf("consumerSQL(DepPageCardImages) lost the renders-image filter — every tool-cta-shaped consumer would be re-resolved for an invisible change:\n%s", imageSQL)
	}
	for _, dep := range []SourceDependency{DepFeedItems, DepDirectoryEntities} {
		if got := consumerSQL(dep); strings.Contains(got, "cc.html_template ~*") {
			t.Errorf("consumerSQL(%s) applies a renders-key filter — news and directory templates render no such key, so this returns NOTHING and the producer silently stops notifying:\n%s", dep, got)
		}
	}
}

// bugs_open/098 at the DESTINATION. The predicate moved here from the two
// producers, and a fix that is only measured at its origin leaves the new home
// unguarded — so both ends assert it (see render_news_section_rerender_test.go).
func TestConsumerSQLCarriesBothPageLifecyclePredicates(t *testing.T) {
	for _, dep := range []SourceDependency{DepPageCardImages, DepFeedItems, DepDirectoryEntities} {
		got := consumerSQL(dep)
		if !strings.Contains(got, "p.status IN ('active', 'deployed')") {
			t.Errorf("consumerSQL(%s) lost the page-status filter — an ARCHIVED page would be re-rendered and re-published, and a retraction would be self-undoing (bugs_open/098)", dep)
		}
		if !strings.Contains(got, "p.deployed_at IS NULL") {
			t.Errorf("consumerSQL(%s) lost the has-shipped floor — a never-built page would be selected; both producers carried this before they migrated here (bugs_open/098)", dep)
		}
		if !strings.Contains(got, "COALESCE(p.rebuild_policy, 'generic') <> 'owned'") {
			t.Errorf("consumerSQL(%s) lost the owned-page exclusion — save_sections refuses an owned page and the run FAILS (bugs_open/208)", dep)
		}
	}
}

// An undeclared dependency must be an ERROR, never an empty result. Silence
// here is indistinguishable from "this site has no consumers", which a producer
// reads as "nobody to tell" — bugs_open/384 itself, one level up.
func TestConsumerPagesRefusesAnUndeclaredDependency(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	got, err := ConsumerPages(context.Background(), db, uuid.New(), SourceDependency("no_such_store"), zap.NewNop())
	if err == nil {
		t.Fatalf("ConsumerPages accepted an undeclared dependency and returned %d pages — a producer would read that as 'nobody consumes my data'", len(got))
	}
	if !strings.Contains(err.Error(), "no query base declares dependency") {
		t.Errorf("error does not name the cause: %v", err)
	}
}

// ConsumesAny is how a producer narrower than its dependency class expresses
// that narrowness (render_directory publishes one KIND out of a shared store).
// Getting it wrong in the permissive direction re-renders unrelated pages.
func TestConsumesAnyMatchesBasesNotFullDeclarations(t *testing.T) {
	page := ConsumerPage{Fields: []ConsumerField{
		{Component: "model-directory", Field: "entries", Source: "query.model_directory"},
		{Component: "tool-list", Field: "items", Source: "query.pages_where_type:tool"},
	}}

	for _, c := range []struct {
		bases []string
		want  bool
	}{
		{[]string{"model_directory"}, true},
		{[]string{"model_directory_full", "model_directory"}, true},
		{[]string{"  Model_Directory  "}, true},        // normalised like Resolve
		{[]string{"pages_where_type"}, true},           // :arg stripped
		{[]string{"model_directory_full"}, false},      // a DIFFERENT base, not a prefix match
		{[]string{"mortgage_lender_directory"}, false}, // the wrong kind must not match
		{[]string{"query.model_directory"}, false},     // takes BASES, not full declarations
		{nil, false},
		{[]string{""}, false},
	} {
		if got := page.ConsumesAny(c.bases...); got != c.want {
			t.Errorf("ConsumesAny(%v) = %v, want %v", c.bases, got, c.want)
		}
	}
}
