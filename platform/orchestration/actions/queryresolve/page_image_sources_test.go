// FILE: platform/orchestration/actions/queryresolve/page_image_sources_test.go
//
// bugs_open/384. pageImageSources is a DECLARATION about queryHandlers: "these
// bases read the card/hero join". A declaration beside a map drifts the moment
// someone adds a resolver, so this test does not read the declaration back —
// it DRIVES every registered handler against a recording sqlmock and checks
// which of them actually issue SQL containing the card join, then requires the
// two sets to be identical in both directions.
//
// Why not a source scan: a scan would make the comment above pageImageJoins
// load-bearing (the a-source-scanning-test-makes-comments-load-bearing trap).
// Recording the SQL the handler executes is the behaviour, not a description
// of it.
//
// Every handler is called with a representative argument and an empty result
// set; handlers that error on ErrNoRows still execute their SQL first, which is
// all this test needs. A handler that returned before querying would be
// recorded as "reads nothing", so the positive controls (pages_where_type,
// blog_posts, pages_under_section) also prove the harness sees SQL at all.

package queryresolve

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// cardJoinNeedle is the one fragment of pageImageJoins that no other query in
// this package carries: the entity-linked card predicate.
const cardJoinNeedle = "ca.purpose = 'card'"

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
	// Enough empty answers for any resolver's query chain; unused ones are fine
	// because ExpectationsWereMet is deliberately not asserted here.
	for i := 0; i < 8; i++ {
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{}))
	}
	_, _ = queryHandlers[base](context.Background(), db, uuid.New(), representativeArgs[base], 0, zap.NewNop())
	return seen
}

func TestPageImageSourcesMatchTheResolversThatReadCards(t *testing.T) {
	if !strings.Contains(pageImageJoins, cardJoinNeedle) {
		t.Fatalf("pageImageJoins no longer contains %q — update cardJoinNeedle to a fragment unique to the card join, or this test cannot see the join at all", cardJoinNeedle)
	}

	reads := map[string]bool{}
	for base := range queryHandlers {
		sqlSeen := recordedSQLFor(t, base)
		if len(sqlSeen) == 0 {
			t.Errorf("handler %q executed no SQL under the harness — give it a representative arg (representativeArgs) so it can be classified", base)
			continue
		}
		for _, q := range sqlSeen {
			if strings.Contains(q, cardJoinNeedle) {
				reads[base] = true
			}
		}
	}

	// Both directions. A resolver that reads cards and is undeclared is the
	// bugs_open/384 shape one level up: its consumers would never be told.
	for base := range reads {
		if !pageImageSources[base] {
			t.Errorf("resolver %q reads the card join but is NOT declared in pageImageSources — its consumers will never be re-resolved when a card lands", base)
		}
	}
	for base := range pageImageSources {
		if _, registered := queryHandlers[base]; !registered {
			t.Errorf("pageImageSources declares %q, which is not a registered query base", base)
			continue
		}
		if !reads[base] {
			t.Errorf("pageImageSources declares %q but its resolver never issued SQL containing %q — stale declaration, or the join moved", base, cardJoinNeedle)
		}
	}
}

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
