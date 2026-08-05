// FILE: platform/orchestration/actions/provocation_generator_action_test.go
//
// The scheduler's one-per-day rule (PLAN §10.5) is the property that bounds how
// much damage a broken generator can do: "At most one rotation per day means a
// broken generator can produce at most one bad day before anyone notices, rather
// than a flood. Do not add a catch-up mode that publishes several at once to fill
// a gap."
//
// nextPublishDates is pure and total precisely so that rule can be tested without
// a database, and the temptation it exists to resist is real and concrete: on
// 2026-08-05 the pool's newest entry was ten days old, and "helpfully" filling
// those ten days is exactly the catch-up the ruling forbids.

package actions

import (
	"strings"
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("bad test date %q: %v", s, err)
	}
	return d
}

// The real situation on the day this was written: the pool's latest entry is
// 2026-07-26 and today is 2026-08-05. Ten days are missing. The scheduler must
// start TOMORROW and must not touch the hole.
func TestSchedulerDoesNotBackfillTheGap(t *testing.T) {
	latest := mustDate(t, "2026-07-26")
	today := mustDate(t, "2026-08-05")

	got := nextPublishDates(&latest, today, 3)
	want := []string{"2026-08-06", "2026-08-07", "2026-08-08"}

	if len(got) != len(want) {
		t.Fatalf("got %d dates, want %d", len(got), len(want))
	}
	for i, w := range want {
		if g := got[i].Format("2006-01-02"); g != w {
			t.Errorf("date %d: got %s, want %s", i, g, w)
		}
	}
	for _, d := range got {
		if !d.After(today) {
			t.Errorf("%s is not in the future; back-dating publishes straight into the archive "+
				"and spends the buffer that makes one bad day survivable (§10.5)", d.Format("2006-01-02"))
		}
	}
}

// One per calendar day, with no repeats — the property the partial unique index
// also enforces in the database. Two provocations sharing a date would make one
// of them unpublishable and the insert would fail.
func TestSchedulerAssignsExactlyOnePerDay(t *testing.T) {
	today := mustDate(t, "2026-08-05")
	got := nextPublishDates(nil, today, 7)

	if len(got) != 7 {
		t.Fatalf("got %d dates, want 7", len(got))
	}
	seen := map[string]bool{}
	for i, d := range got {
		k := d.Format("2006-01-02")
		if seen[k] {
			t.Fatalf("date %s assigned twice", k)
		}
		seen[k] = true
		if i > 0 {
			gap := d.Sub(got[i-1])
			if gap != 24*time.Hour {
				t.Errorf("gap between %s and %s is %v, want 24h",
					got[i-1].Format("2006-01-02"), k, gap)
			}
		}
	}
}

// When the pool already runs ahead of today, the scheduler must continue from
// the end of the queue, not from tomorrow — otherwise it would collide with
// dates already assigned.
func TestSchedulerContinuesFromTheEndOfAFutureQueue(t *testing.T) {
	latest := mustDate(t, "2026-08-20") // already scheduled well ahead
	today := mustDate(t, "2026-08-05")

	got := nextPublishDates(&latest, today, 2)
	if got[0].Format("2006-01-02") != "2026-08-21" {
		t.Fatalf("got %s, want 2026-08-21 — scheduling must resume after the queue, not overwrite it",
			got[0].Format("2006-01-02"))
	}
}

// An empty pool must still schedule forward, never today or earlier.
func TestSchedulerOnAnEmptyPoolStartsTomorrow(t *testing.T) {
	today := mustDate(t, "2026-08-05")
	got := nextPublishDates(nil, today, 1)
	if got[0].Format("2006-01-02") != "2026-08-06" {
		t.Fatalf("got %s, want 2026-08-06", got[0].Format("2006-01-02"))
	}
}

func TestSchedulerReturnsNothingForNonPositiveCounts(t *testing.T) {
	today := mustDate(t, "2026-08-05")
	if d := nextPublishDates(nil, today, 0); len(d) != 0 {
		t.Errorf("n=0 returned %d dates", len(d))
	}
	if d := nextPublishDates(nil, today, -3); len(d) != 0 {
		t.Errorf("n=-3 returned %d dates", len(d))
	}
}

// MUTATION CONTROL. If nextPublishDates started at the day AFTER the pool's
// latest without the "tomorrow at the earliest" floor, it would backfill — and
// TestSchedulerDoesNotBackfillTheGap must be the thing that notices. This models
// that broken implementation and requires the two to disagree, so the test above
// cannot be passing for an unrelated reason.
func TestBackfillWouldBeCaught(t *testing.T) {
	latest := mustDate(t, "2026-07-26")
	today := mustDate(t, "2026-08-05")

	// The broken version: continue from the pool's latest, ignoring today.
	broken := latest.AddDate(0, 0, 1) // 2026-07-27 — ten days in the past
	real := nextPublishDates(&latest, today, 1)[0]

	if real.Equal(broken) {
		t.Fatal("the scheduler agrees with a backfilling implementation; the no-backfill test is vacuous")
	}
	if !broken.Before(today) {
		t.Fatal("mutation setup is wrong: the broken date must be in the past to model the defect")
	}
}

// ---------------------------------------------------------------------------
// Generation guards
// ---------------------------------------------------------------------------

// validateGenerated must drop the unusable and keep the rest — it is not a
// second gate, so it must not reject on anything evaluative.
func TestValidateGeneratedDropsOnlyTheUnusable(t *testing.T) {
	in := []generatedProvocation{
		{Slug: "good-one", Title: "A flat contestable claim", Teaser: "Short hook.", Body: "Body."},
		{Slug: "Bad Slug", Title: "Has spaces", Teaser: "t", Body: "b"},
		{Slug: "", Title: "No slug", Teaser: "t", Body: "b"},
		{Slug: "missing-body", Title: "No body", Teaser: "t", Body: ""},
		{Slug: "good-one", Title: "Duplicate slug in batch", Teaser: "t", Body: "b"},
		{Slug: "second-good", Title: "Another claim", Teaser: "Hook.", Body: "Body."},
	}
	ok, dropped := validateGenerated(in)
	if len(ok) != 2 {
		t.Fatalf("kept %d, want 2 (good-one, second-good); dropped: %s", len(ok), strings.Join(dropped, "; "))
	}
	if len(dropped) != 4 {
		t.Errorf("dropped %d, want 4: %s", len(dropped), strings.Join(dropped, "; "))
	}
	if ok[0].Slug != "good-one" || ok[1].Slug != "second-good" {
		t.Errorf("kept the wrong rows: %+v", ok)
	}
}

// The generator's parser must be as strict as the gate's, and for the same
// reason: a truncated completion that a lenient parser recovers puts half a
// provocation into the pool.
func TestParseGeneratedRejectsMalformedReplies(t *testing.T) {
	cases := []struct{ name, raw string }{
		{"empty", ""},
		{"prose", "Here are some provocations you might like!"},
		{"truncated mid-object", `[{"slug":"a","title":"T","teaser":"x","body":"y"},{"slug":"b","tit`},
		{"empty array", `[]`},
		{"unknown field", `[{"slug":"a","title":"T","teaser":"x","body":"y","score":9}]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseGenerated(tc.raw); err == nil {
				t.Fatalf("accepted a malformed reply (%s)", tc.name)
			}
		})
	}
}

func TestParseGeneratedAcceptsAFencedReply(t *testing.T) {
	raw := "```json\n" + `[{"slug":"a-claim","title":"A claim","teaser":"Hook.","body":"Body."}]` + "\n```"
	got, err := parseGenerated(raw)
	if err != nil {
		t.Fatalf("rejected a well-formed fenced reply: %v", err)
	}
	if len(got) != 1 || got[0].Slug != "a-claim" {
		t.Fatalf("parsed wrongly: %+v", got)
	}
}

// The generator must never be able to write a publishable row. This is the
// containment property, so it is asserted against the SQL text itself rather
// than trusted to review.
func TestGeneratorInsertsDraftsOnly(t *testing.T) {
	// insertDrafts' statement is the single place a generated row is created.
	// If this ever gains 'approved' or a publish_on value, the generator can
	// publish without the gate, which PLAN §10 forbids.
	const forbidden = "'approved'"
	if strings.Contains(generatorInsertSQL, forbidden) {
		t.Fatalf("the generator's INSERT mentions %s — it must write drafts only", forbidden)
	}
	if !strings.Contains(generatorInsertSQL, "'draft'") {
		t.Fatal("the generator's INSERT no longer writes 'draft'")
	}
	if strings.Contains(generatorInsertSQL, "publish_on") {
		t.Fatal("the generator's INSERT sets publish_on — dating is the scheduler's job, and a dated " +
			"approved row is publishable")
	}
}
