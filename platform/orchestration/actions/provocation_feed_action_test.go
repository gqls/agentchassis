// Tests for render_provocation_feed.
//
// These are a PORT of provocation_pipeline/builder/verify_rotation.py, which is
// the specification for this action. That script's rule is the right one and it
// carries over: rotation cannot be verified by looking at one day's output,
// because a hardcoded provocation and a rotated one are indistinguishable on any
// single day — that is how the original bug survived a month. So the mechanism is
// driven across every date in the schedule's span and the invariants are asserted
// on all of them.
//
// Keep these in step with verify_rotation.py. Two implementations of one feed is
// a real risk and the mitigation is that both are checked against the same named
// invariants.

package actions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// A schedule with the same SHAPE as the live pool: some entries with a case and
// a long-form today-shape, some historical ones with neither.
// ---------------------------------------------------------------------------

func day(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func testSchedule() []provocation {
	return []provocation{
		// Historical: archive-shape only, no case. Exercises both fallbacks.
		{Slug: "no-case", PublishOn: day("2026-06-28"),
			Title:  "Group chats replaced friendship maintenance",
			Teaser: "Presence without effort."},
		// Historical: case written, no today-shape.
		{Slug: "case-only", PublishOn: day("2026-06-29"),
			Title: "Nobody actually reads terms of service", Teaser: "The cost outweighs the power.",
			DetailBody: "Reading takes an hour.\n\nWhich moves the burden elsewhere."},
		{Slug: "case-only-2", PublishOn: day("2026-07-01"),
			Title: "Reading fiction makes you worse at facts", Teaser: "Narrative wants a tidy arc.",
			DetailBody: "A novel teaches you to expect that events connect.\n\nAgainst that: fiction helps."},
		// Fully authored, both shapes.
		{Slug: "full", PublishOn: day("2026-07-26"),
			Title:    "Nobody actually wants a personalised internet",
			Teaser:   "Personalisation removes what you'd have shared with a stranger.",
			CardDesc: "What gets sold as personalisation is the quiet removal of common ground.",
			Headline: "Nobody actually <em>wants</em> a personalised internet.",
			Body:     "Every feed is tuned to one person, and every conversation opens with have you seen."},
	}
}

func buildOn(t *testing.T, schedule []provocation, on string) map[string]interface{} {
	t.Helper()
	feed, _, err := buildProvocationFeed(schedule, day(on), "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("build %s: %v", on, err)
	}
	return feed
}

func todayOf(feed map[string]interface{}) map[string]interface{} {
	return feed["today"].(map[string]interface{})
}

func archiveSlugs(feed map[string]interface{}) []string {
	entries := feed["archive"].(map[string]interface{})["entries"].([]map[string]interface{})
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e["slug"].(string))
	}
	return out
}

// shortDateExpected is deliberately re-derived here rather than calling
// shortDate: a test that formats dates with the code under test cannot detect a
// formatting change, it can only agree with it. (Carried over from
// verify_rotation.py, where the same note appears for the same reason.)
func shortDateExpected(t time.Time) string {
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	return fmt.Sprintf("%d %s", t.Day(), months[int(t.Month())-1])
}

// ---------------------------------------------------------------------------
// The sweep — every invariant, on every date in the span plus ten days past it
// ---------------------------------------------------------------------------

func TestRotationInvariantsAcrossTheWholeSpan(t *testing.T) {
	schedule := testSchedule()
	first := schedule[0].PublishOn
	last := schedule[len(schedule)-1].PublishOn.AddDate(0, 0, 10)

	dateOf := map[string]time.Time{}
	titleOf := map[string]string{}
	for _, p := range schedule {
		dateOf[p.Slug] = p.PublishOn
		titleOf[p.Slug] = p.Title
	}

	var seenToday []string
	prevToday := ""
	prevArchiveLen := -1
	days := 0

	for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
		feed := buildOn(t, schedule, d.Format("2006-01-02"))
		today := todayOf(feed)
		slugs := archiveSlugs(feed)
		days++

		// 1. today can always BECOME an archive entry. Missing either field is
		//    what stranded the live archive on 5 Jul while the site kept
		//    promising a daily.
		slug, _ := today["slug"].(string)
		if slug == "" {
			t.Fatalf("%s: today has no slug", d.Format("2006-01-02"))
		}

		// 2. A provocation is NEVER today and archived at the same time. The
		//    owner's rule stated as an invariant instead of a hope.
		for _, s := range slugs {
			if s == slug {
				t.Errorf("%s: today %q is also in the archive", d.Format("2006-01-02"), slug)
			}
		}

		// 2b. ...and the date must be the RIGHT one, not merely present. Freezing
		//     today.date to a literal passes every other invariant here, and that
		//     date is what gets carried into the archive on promotion — so a frozen
		//     one would date every archived entry identically while looking
		//     entirely plausible. "Populated" is not "correct".
		if got, want := today["date"].(string), shortDateExpected(dateOf[slug]); got != want {
			t.Errorf("%s: today.date is %q, should be %q", d.Format("2006-01-02"), got, want)
		}

		// 2c. Promotion carries each entry's own date, for the same reason.
		entries := feed["archive"].(map[string]interface{})["entries"].([]map[string]interface{})
		for _, e := range entries {
			es := e["slug"].(string)
			if got, want := e["date"].(string), shortDateExpected(dateOf[es]); got != want {
				t.Errorf("%s: archived %s dated %q, should be %q", d.Format("2006-01-02"), es, got, want)
			}
		}

		// 3. The arena's first card must NOT name today's provocation, and must
		//    stay a usable card — both renderers drop one missing title or url,
		//    which would remove the only lobby route into today's round.
		cards := feed["arena"].(map[string]interface{})["cards"].([]map[string]interface{})
		card0 := cards[0]
		if card0["title"].(string) == titleOf[slug] {
			t.Errorf("%s: arena card 0 names today's provocation", d.Format("2006-01-02"))
		}
		if card0["title"].(string) == "" || card0["url"].(string) == "" {
			t.Errorf("%s: arena card 0 is not a usable card", d.Format("2006-01-02"))
		}

		// 3b. The engine's contract. round.go takes the whole `today` object
		//     server-side and persists it as the round's provocation, so these keys
		//     are load-bearing for the GAME, not the display. Asserted because the
		//     seal creates a standing temptation to "seal" today by emptying them.
		for _, k := range []string{"headline", "body", "slug", "date"} {
			if s, _ := today[k].(string); strings.TrimSpace(s) == "" {
				t.Errorf("%s: today.%s is empty — that breaks the round, it does not seal it",
					d.Format("2006-01-02"), k)
			}
		}

		// 3c. The display surfaces must have something real to show instead.
		seal, ok := feed["seal"].(map[string]interface{})
		if !ok || seal["headline"].(string) == "" || seal["body"].(string) == "" {
			t.Errorf("%s: feed carries no seal copy, so the display surfaces have nothing "+
				"to say in place of today's provocation", d.Format("2006-01-02"))
		}

		// 4. Archive holds exactly the entries published strictly before today's,
		//    newest first.
		var want []string
		for i := len(schedule) - 1; i >= 0; i-- {
			if schedule[i].PublishOn.Before(dateOf[slug]) {
				want = append(want, schedule[i].Slug)
			}
		}
		if strings.Join(slugs, ",") != strings.Join(want, ",") {
			t.Errorf("%s: archive is %v, want %v", d.Format("2006-01-02"), slugs, want)
		}

		// 5. The archive only ever grows, and only when today changes.
		if prevArchiveLen >= 0 {
			if slug == prevToday && len(slugs) != prevArchiveLen {
				t.Errorf("%s: archive changed while today did not", d.Format("2006-01-02"))
			}
			if slug != prevToday && len(slugs) != prevArchiveLen+1 {
				t.Errorf("%s: today changed but archive grew by %d, not 1",
					d.Format("2006-01-02"), len(slugs)-prevArchiveLen)
			}
		}
		prevToday, prevArchiveLen = slug, len(slugs)

		if len(seenToday) == 0 || seenToday[len(seenToday)-1] != slug {
			seenToday = append(seenToday, slug)
		}
	}

	// 6. The whole point: the same code on different dates yields different
	//    provocations. One distinct value would mean nothing rotates — which is
	//    exactly the state this action exists to end.
	if len(seenToday) != len(schedule) {
		t.Errorf("expected %d distinct provocations across the span, saw %d (%v)",
			len(schedule), len(seenToday), seenToday)
	}

	// 7. Every scheduled provocation gets its turn, in order.
	for i, p := range schedule {
		if i < len(seenToday) && seenToday[i] != p.Slug {
			t.Errorf("sequence position %d is %q, want %q", i, seenToday[i], p.Slug)
		}
	}

	t.Logf("dates checked: %d, distinct today-values: %d of %d scheduled",
		days, len(seenToday), len(schedule))
}

// 8. Before the schedule starts, the builder must FAIL rather than invent one.
func TestBuildBeforeScheduleStartsFailsLoud(t *testing.T) {
	schedule := testSchedule()
	_, _, err := buildProvocationFeed(schedule, day("2026-06-01"), "x")
	if err == nil {
		t.Fatal("built a feed for a date before the schedule starts; it should refuse")
	}
	if !strings.Contains(err.Error(), "no provocation is published") {
		t.Errorf("unhelpful error for a too-early date: %v", err)
	}
}

// ---------------------------------------------------------------------------
// The seal, in both directions. These are the two ways this feed has actually
// been got wrong, so they are asserted directly rather than left to the sweep.
// ---------------------------------------------------------------------------

// Emptying today's keys is how the first seal attempt broke the engine. The
// checker must call that a break, not a seal.
func TestEmptyingTodayIsRefusedNotTreatedAsSealed(t *testing.T) {
	feed := buildOn(t, testSchedule(), "2026-07-26")
	todayEntry := testSchedule()[3]

	for _, key := range []string{"headline", "body", "slug", "date"} {
		f := buildOn(t, testSchedule(), "2026-07-26")
		todayOf(f)[key] = ""
		problems := checkFeed(f, todayEntry)
		if len(problems) == 0 {
			t.Errorf("emptying today.%s was accepted; round.go would serve a blank question", key)
			continue
		}
		if !strings.Contains(strings.Join(problems, " "), "round.go") {
			t.Errorf("today.%s: the complaint should name why it breaks the round, got %v", key, problems)
		}
	}
	_ = feed
}

// A card that names today's provocation is the leak the seal exists to close.
func TestArenaCardNamingTodayIsRefused(t *testing.T) {
	schedule := testSchedule()
	todayEntry := schedule[3]

	for _, field := range []string{"title", "desc"} {
		feed := buildOn(t, schedule, "2026-07-26")
		cards := feed["arena"].(map[string]interface{})["cards"].([]map[string]interface{})
		cards[0][field] = todayEntry.Title
		problems := checkFeed(feed, todayEntry)
		if len(problems) == 0 {
			t.Errorf("arena card 0 %s carrying today's title was accepted — that is the leak", field)
		}
	}
}

// The card must survive being sealed: both renderers drop a card missing either
// field, which silently removes the only lobby route into the round. That is a
// worse page than the leak, and it would pass a naive "today is absent" check.
func TestSealedCardMustStayUsable(t *testing.T) {
	schedule := testSchedule()
	todayEntry := schedule[3]

	for _, field := range []string{"title", "url"} {
		feed := buildOn(t, schedule, "2026-07-26")
		cards := feed["arena"].(map[string]interface{})["cards"].([]map[string]interface{})
		cards[0][field] = ""
		problems := checkFeed(feed, todayEntry)
		if len(problems) == 0 {
			t.Errorf("arena card 0 with an empty %s was accepted; both renderers would drop it", field)
		}
	}
}

// The sample must be a PAST provocation. Showing today's in full would be the
// leak wearing the seal's clothes.
func TestSampleIsNeverToday(t *testing.T) {
	schedule := testSchedule()
	todayEntry := schedule[3]
	feed := buildOn(t, schedule, "2026-07-26")

	sample, ok := feed["sample"].(map[string]interface{})
	if !ok {
		t.Fatal("no sample built; the display surfaces have nothing to show in place of today")
	}
	if sample["slug"].(string) == todayOf(feed)["slug"].(string) {
		t.Fatal("sample is today's provocation")
	}
	inArchive := false
	for _, s := range archiveSlugs(feed) {
		if s == sample["slug"].(string) {
			inArchive = true
		}
	}
	if !inArchive {
		t.Error("sample is not in the archive, so it has not necessarily been argued yet")
	}

	// And the checker must catch it if it ever becomes today's.
	feed["sample"].(map[string]interface{})["slug"] = todayOf(feed)["slug"]
	if len(checkFeed(feed, todayEntry)) == 0 {
		t.Error("a sample equal to today's provocation was accepted")
	}
}

// An entry appearing as today AND in the archive would publish today's full case
// on the archive page — a third leak path.
func TestTodayInArchiveIsRefused(t *testing.T) {
	schedule := testSchedule()
	todayEntry := schedule[3]
	feed := buildOn(t, schedule, "2026-07-26")

	entries := feed["archive"].(map[string]interface{})["entries"].([]map[string]interface{})
	entries = append(entries, map[string]interface{}{
		"slug": todayOf(feed)["slug"], "title": todayEntry.Title,
		"teaser": todayEntry.Teaser, "date": "26 Jul",
	})
	feed["archive"].(map[string]interface{})["entries"] = entries

	if len(checkFeed(feed, todayEntry)) == 0 {
		t.Error("today's entry in the archive was accepted; the archive page would publish its case")
	}
}

// ---------------------------------------------------------------------------
// Fallbacks — the historical entries have no long-form shape and must still
// produce a round-worthy `today`.
// ---------------------------------------------------------------------------

func TestHistoricalEntriesStillProduceAPlayableToday(t *testing.T) {
	schedule := testSchedule()

	// An entry with neither headline/body nor a case: falls back to title/teaser.
	feed := buildOn(t, schedule, "2026-06-28")
	today := todayOf(feed)
	if today["headline"].(string) != schedule[0].Title {
		t.Errorf("headline fallback: got %q, want the title", today["headline"])
	}
	if today["body"].(string) != schedule[0].Teaser {
		t.Errorf("body fallback: got %q, want the teaser", today["body"])
	}

	// An entry with a case but no today-shape: body falls back to the case.
	feed = buildOn(t, schedule, "2026-06-29")
	today = todayOf(feed)
	if today["body"].(string) != schedule[1].DetailBody {
		t.Errorf("body should fall back to the full case, got %q", today["body"])
	}
}

// ---------------------------------------------------------------------------
// Publish guards
// ---------------------------------------------------------------------------

// A no-op must SKIP rather than commit. Committing would advance generated_at
// daily while the site repeated itself — the original bug wearing the fix as a
// disguise, and it would make the file's git history claim rotation that never
// happened.
func TestUnchangedFeedSkipsTheCommit(t *testing.T) {
	feed := buildOn(t, testSchedule(), "2026-07-26")
	a, err := summariseFeed(feed)
	if err != nil {
		t.Fatal(err)
	}

	// Same content, different timestamp — exactly what a daily rebuild produces
	// when nothing has rotated.
	later, _, err := buildProvocationFeed(testSchedule(), day("2026-07-27"), "2026-12-25T09:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	b, err := summariseFeed(later)
	if err != nil {
		t.Fatal(err)
	}

	skip, err := checkAgainstServed(a, b, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !skip {
		t.Error("a feed differing only in generated_at should skip the commit")
	}
}

// A real rotation must NOT skip. Without this, the test above is satisfied by an
// implementation that skips everything.
func TestRotatedFeedDoesNotSkip(t *testing.T) {
	before := buildOn(t, testSchedule(), "2026-07-01")
	after := buildOn(t, testSchedule(), "2026-07-26")

	a, _ := summariseFeed(before)
	b, _ := summariseFeed(after)
	if a.TodaySlug == b.TodaySlug {
		t.Fatal("test is not exercising a rotation")
	}

	skip, err := checkAgainstServed(a, b, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skip {
		t.Error("a genuine rotation was skipped — the site would never update")
	}
}

// A shrinking archive means rows were deleted or un-approved. Refuse unless it
// is stated deliberately.
func TestShrinkingArchiveIsRefusedUnlessAllowed(t *testing.T) {
	full := buildOn(t, testSchedule(), "2026-07-26")
	small := buildOn(t, testSchedule(), "2026-06-29")

	served, _ := summariseFeed(full)
	next, _ := summariseFeed(small)
	if next.ArchiveCount >= served.ArchiveCount {
		t.Fatal("test is not exercising a shrink")
	}

	if _, err := checkAgainstServed(served, next, false); err == nil {
		t.Error("a shrinking archive was accepted without allow_shrink")
	}
	if _, err := checkAgainstServed(served, next, true); err != nil {
		t.Errorf("allow_shrink should permit a deliberate shrink, got %v", err)
	}
}

// With nothing served, checkAgainstServed itself neither errors nor reports "no
// change" — it cannot distinguish a first publish from a failed fetch, so the
// decision belongs to the caller.
//
// CORRECTED after the council's round 1: this comment used to say "the guards must
// not block: the feed's content is fully determined by the pool, so publishing is
// still correct". That was the wrong call and a reviewer caught it — the served
// feed supplies the shrink guard's denominator, so publishing blind disables the
// only defence against silently losing provocations. The ACTION now refuses unless
// allow_unverified_publish is set (TestUnverifiedPublishDefaultsToRefusing); this
// function stays permissive because a refusal here would also break a genuine
// first publish.
// Layer note, because the name reads like the opposite of the round-2 guard
// below: this pins the COMPARISON layer only. A nil served feed means "nothing
// to compare against", so checkAgainstServed neither skips nor errors. The
// refusal is the FETCH layer's job, gated by allow_unverified_publish — see
// TestUnverifiedPublishDefaultsToRefusing. Both are true at once.
func TestMissingServedFeedDoesNotBlockPublishing(t *testing.T) {
	feed := buildOn(t, testSchedule(), "2026-07-26")
	next, _ := summariseFeed(feed)

	skip, err := checkAgainstServed(nil, next, false)
	if err != nil {
		t.Errorf("a missing served feed must not block the publish, got %v", err)
	}
	if skip {
		t.Error("a missing served feed must not be read as no change")
	}
}

// ---------------------------------------------------------------------------
// Round 2 — guards added after the council's first verdict
// ---------------------------------------------------------------------------

// An unreadable served feed must REFUSE, not publish blind.
//
// The first draft tolerated it, reasoning that content is determined by the pool.
// That misses what the comparison is for: the served feed supplies the shrink
// guard's denominator, so tolerating a failed fetch switched off the only defence
// against silently dropping provocations. This pins the corrected behaviour —
// checkAgainstServed with no served feed must not report "safe to publish" in a
// way the caller can mistake for a verified publish.
func TestNoServedFeedMeansUnverifiedNotVerified(t *testing.T) {
	feed := buildOn(t, testSchedule(), "2026-07-26")
	next, _ := summariseFeed(feed)

	// checkAgainstServed itself stays permissive by design — it cannot tell a
	// first publish from a failed fetch. The refusal is the CALLER's job, which is
	// why the caller now branches on AllowUnverifiedPublish. What must remain true
	// here is that a nil served feed never yields skip=true, because that would
	// silently drop the publish instead of performing it.
	skip, err := checkAgainstServed(nil, next, false)
	if err != nil {
		t.Errorf("checkAgainstServed should not itself error on a nil served feed: %v", err)
	}
	if skip {
		t.Error("a nil served feed must never read as 'no change' — that would skip a real publish")
	}
}

// The config flag must exist and default to fail-closed. A guard whose default is
// permissive is not a guard.
func TestUnverifiedPublishDefaultsToRefusing(t *testing.T) {
	fc, err := parseProvocationFeedConfig(map[string]interface{}{"domain": "vonc.com"})
	if err != nil {
		t.Fatalf("plain config must parse: %v", err)
	}
	if fc.AllowUnverifiedPublish {
		t.Error("allow_unverified_publish must default to false, or a failed fetch silently disables the shrink guard")
	}
	fc, err = parseProvocationFeedConfig(map[string]interface{}{
		"domain": "vonc.com", "allow_unverified_publish": true,
	})
	if err != nil {
		t.Fatalf("config with the override must parse: %v", err)
	}
	if !fc.AllowUnverifiedPublish {
		t.Error("allow_unverified_publish was not read from config, so the override is unreachable")
	}
}

// ---------------------------------------------------------------------------
// Categories (RFC_013, ratified 2026-08-05: one feed FILE per category)
// ---------------------------------------------------------------------------

// The whole safety case for option (a) rests on the default category's artefact
// being untouched. If this ever changes, every existing reader breaks at once —
// including the engine, which would 404 and 503 the Gauntlet.
func TestDefaultCategoryKeepsTheOriginalFilenameForever(t *testing.T) {
	fc, err := parseProvocationFeedConfig(map[string]interface{}{"domain": "vonc.com"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if fc.Category != "general" {
		t.Errorf("category must default to 'general' (what every existing row is), got %q", fc.Category)
	}
	if fc.Filename != "provocations.json" {
		t.Errorf("the default category MUST publish provocations.json — the engine reads that exact "+
			"path and nothing tells it otherwise; got %q", fc.Filename)
	}

	// Migration 283's live schedule row passes this explicitly. It must keep working.
	fc, err = parseProvocationFeedConfig(map[string]interface{}{
		"domain": "vonc.com", "filename": "provocations.json",
	})
	if err != nil {
		t.Fatalf("the existing live schedule row's config must still parse: %v", err)
	}
	if fc.Filename != "provocations.json" {
		t.Errorf("got %q", fc.Filename)
	}
}

func TestCategoryDerivesItsOwnFilename(t *testing.T) {
	fc, err := parseProvocationFeedConfig(map[string]interface{}{
		"domain": "vonc.com", "category": "pets",
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if fc.Filename != "provocations-pets.json" {
		t.Errorf("expected provocations-pets.json, got %q", fc.Filename)
	}
}

// The failure this prevents is silent and severe: copy migration 283's schedule
// row to add a category, keep its explicit filename, and pets' provocation is
// published OVER the general feed and served as everybody's daily question.
// Nothing downstream detects it — the engine does not read the category and the
// shrink guard compares unrelated archives.
func TestAFilenameContradictingItsCategoryIsRefused(t *testing.T) {
	_, err := parseProvocationFeedConfig(map[string]interface{}{
		"domain": "vonc.com", "category": "pets", "filename": "provocations.json",
	})
	if err == nil {
		t.Fatal("a pets category publishing provocations.json was ACCEPTED — that overwrites the " +
			"general feed with pets content and the engine serves it as today's question")
	}
	if !strings.Contains(err.Error(), "contradicts") {
		t.Errorf("the error should say what is wrong and how to fix it; got: %v", err)
	}
}

// The category reaches a filename and a URL path segment, so it cannot be free
// text. Traversal is the case that matters.
func TestCategoryRejectsAnythingUnsafeInAPath(t *testing.T) {
	for _, bad := range []string{
		"../../etc/passwd", "a/b", "Pets", "pets.json", "a b", "a?b", "",
		strings.Repeat("x", 41),
	} {
		_, err := parseProvocationFeedConfig(map[string]interface{}{
			"domain": "vonc.com", "category": bad,
		})
		// "" is the one that must be ACCEPTED-as-absent rather than rejected: an
		// absent category is the default, which is how every current row behaves.
		if bad == "" {
			if err != nil {
				t.Errorf("an empty category must mean 'default', not an error: %v", err)
			}
			continue
		}
		if err == nil {
			t.Errorf("category %q was accepted; it becomes part of a path", bad)
		}
	}
}

// The bootstrap widening must stay narrow. These assert the DECISION TABLE
// directly, because the dangerous direction is the permissive one and it is
// reached only through a 404, which no ordinary run exercises.
func TestBootstrapIsPermittedOnlyForAGenuineFirstPublish(t *testing.T) {
	// Calls the REAL decision function. An earlier draft of this test re-stated
	// the condition inline and so could not have failed for any edit to the code
	// — recorded because that is the trap this repo keeps paying for.
	notFound := fmt.Errorf("fetching: %w", errFeedNotPublished) // wrapped, as the action sees it
	other := fmt.Errorf("GET https://vonc.com/data/x.json: 500 Internal Server Error")

	cases := []struct {
		name         string
		ferr         error
		archiveCount int
		want         bool
	}{
		{"new category, nothing behind it", notFound, 0, true},
		{"new category, back-dated rows", notFound, 3, false},
		{"established feed, spurious 404", notFound, 8, false},
		{"500 with an empty archive", other, 0, false},
		{"500 against an established feed", other, 8, false},
	}
	for _, c := range cases {
		if got := shouldBootstrap(c.ferr, c.archiveCount); got != c.want {
			t.Errorf("%s: shouldBootstrap = %v, want %v", c.name, got, c.want)
		}
	}
}

// Mutation control for the test above: if the archive condition were dropped —
// the plausible "simplification" — the established-feed case would flip to
// permitting a publish. This asserts the two inputs genuinely disagree, so the
// table is not passing by accident.
func TestTheBootstrapTableWouldCatchDroppingTheArchiveCondition(t *testing.T) {
	notFound := fmt.Errorf("wrapped: %w", errFeedNotPublished)
	withCondition := shouldBootstrap(notFound, 8)
	withoutCondition := errors.Is(notFound, errFeedNotPublished) // the weakened version
	if withCondition == withoutCondition {
		t.Fatal("the archive condition makes no difference on an established feed, so the " +
			"bootstrap table cannot detect it being removed")
	}
}

// The domain is interpolated into an outbound URL, so anything that is not a
// bare hostname must be refused before it reaches the network.
func TestOutboundDomainRejectsNonHostnames(t *testing.T) {
	// No DB needed: every one of these must fail the shape check, which runs
	// before the allow-list query. A nil *sql.DB would panic if any reached it,
	// so this also proves the shape check comes first.
	for _, bad := range []string{
		"", "vonc.com/../etc", "user:pw@vonc.com", "vonc.com:8080",
		"localhost/latest/meta-data", "vonc.com?x=1", "vonc.com#f", "a b.com",
		"http://vonc.com", "vonc.com\\x",
	} {
		if err := assertKnownDomain(context.Background(), nil, bad); err == nil {
			t.Errorf("domain %q was accepted; it must be refused before any fetch", bad)
		}
	}
}
