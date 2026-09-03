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
	"os"
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

// ---------------------------------------------------------------------------
// Category-aware scheduling (owner ruling 2026-08-09, PLAN §13 ruling 9)
// ---------------------------------------------------------------------------

// TestGroupPendingByCategory_IsANoOpWhileEveryRowIsGeneral is the claim the
// ruling was taken on: fix it NOW, while every live row is still `general`, so
// the change cannot bite the day a second category appears. If this fails, the
// fix was not safe to land early and the ruling's premise was wrong.
func TestGroupPendingByCategory_IsANoOpWhileEveryRowIsGeneral(t *testing.T) {
	in := []pendingProvocation{
		{id: "1", slug: "oldest", category: "general"},
		{id: "2", slug: "middle", category: "general"},
		{id: "3", slug: "newest", category: "general"},
	}
	order, by := groupPendingByCategory(in)

	if len(order) != 1 || order[0] != "general" {
		t.Fatalf("order = %v, want exactly [general] — one category in, one group out", order)
	}
	got := by["general"]
	if len(got) != len(in) {
		t.Fatalf("group holds %d rows, want %d — rows were dropped by grouping", len(got), len(in))
	}
	for i := range in {
		if got[i].slug != in[i].slug {
			t.Errorf("position %d = %q, want %q — the oldest-first order did not survive grouping, "+
				"which is the anti-starvation property", i, got[i].slug, in[i].slug)
		}
	}
}

// TestGroupPendingByCategory_SeparatesCategoriesAndKeepsEachInOrder is the
// behaviour the fix exists for. Before it, a mixed batch was handed consecutive
// dates from a whole-domain high-water mark, so a new category inherited the
// busiest one's future and was silently never scheduled near-term.
func TestGroupPendingByCategory_SeparatesCategoriesAndKeepsEachInOrder(t *testing.T) {
	// Interleaved on purpose: the query returns oldest-first ACROSS categories.
	in := []pendingProvocation{
		{id: "1", slug: "music-old", category: "music"},
		{id: "2", slug: "food-old", category: "food"},
		{id: "3", slug: "music-new", category: "music"},
		{id: "4", slug: "film-only", category: "film"},
		{id: "5", slug: "food-new", category: "food"},
	}
	order, by := groupPendingByCategory(in)

	// First-appearance order, NOT map order: the category holding the
	// longest-waiting row must be scheduled first.
	want := []string{"music", "food", "film"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v — first-appearance order is what makes this "+
				"deterministic between runs on an over-subscribed pool", order, want)
		}
	}
	for cat, wantSlugs := range map[string][]string{
		"music": {"music-old", "music-new"},
		"food":  {"food-old", "food-new"},
		"film":  {"film-only"},
	} {
		got := by[cat]
		if len(got) != len(wantSlugs) {
			t.Errorf("%s: %d rows, want %d", cat, len(got), len(wantSlugs))
			continue
		}
		for i := range wantSlugs {
			if got[i].slug != wantSlugs[i] {
				t.Errorf("%s[%d] = %q, want %q", cat, i, got[i].slug, wantSlugs[i])
			}
		}
	}

	// Every input row must appear exactly once: a grouping that loses a row
	// silently un-schedules a provocation, which is this bug in a new costume.
	total := 0
	for _, g := range by {
		total += len(g)
	}
	if total != len(in) {
		t.Errorf("grouped %d rows, want %d — grouping lost or duplicated work", total, len(in))
	}
}

// TestNextPublishDates_StartsTomorrowForACategoryWithNoHistory pins the arithmetic
// half of the fix at the seam the SQL feeds. A brand-new category reads
// max(publish_on) = NULL WITHIN ITS OWN CATEGORY, so it must start tomorrow — not
// after whatever the busiest category has already queued.
func TestNextPublishDates_StartsTomorrowForACategoryWithNoHistory(t *testing.T) {
	today := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)

	fresh := nextPublishDates(nil, today, 2)
	if len(fresh) != 2 {
		t.Fatalf("got %d dates, want 2", len(fresh))
	}
	if want := today.AddDate(0, 0, 1); !fresh[0].Equal(time.Date(want.Year(), want.Month(), want.Day(), 0, 0, 0, 0, time.UTC)) {
		t.Errorf("a category with no history starts %v, want tomorrow (%v)", fresh[0], want)
	}

	// The contrast that makes the point: the OLD whole-domain behaviour would have
	// passed this busy category's high-water mark for the fresh one above.
	busy := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)
	crowded := nextPublishDates(&busy, today, 1)
	if crowded[0].Before(busy) {
		t.Errorf("a category with history must continue after it, got %v", crowded[0])
	}
	if !fresh[0].Before(crowded[0]) {
		t.Errorf("the fresh category (%v) must schedule BEFORE the crowded one (%v) — "+
			"if these are equal, the high-water mark is being read across categories again",
			fresh[0], crowded[0])
	}
}

// ---------------------------------------------------------------------------
// The options map — the config key that looked live and was not (2026-08-10)
// ---------------------------------------------------------------------------

// TestConfiguredTokenBudgetReachesTheOptionsMap is the test whose absence cost
// three config migrations.
//
// `ai_service.max_tokens` was set to 8000 on the generate step and the next run
// still died at output_tokens=2048, because the action handed `GenerateText` an
// empty options map and the provider client reads the budget from nowhere else.
// The config was right, the request was wrong, and nothing in between said so.
//
// This asserts the ONE property that distinguishes those two states: the number
// in the config comes out in the map that is actually passed to the client.
func TestConfiguredTokenBudgetReachesTheOptionsMap(t *testing.T) {
	// jsonb decodes numbers as float64 — the shape a live config really has.
	aiCfg := map[string]interface{}{"model": "claude-sonnet-5", "max_tokens": float64(8000)}

	opts := llmOptionsFromConfig(nil, aiCfg, nil, "test")

	got, ok := opts["max_tokens"]
	if !ok {
		t.Fatal("max_tokens is configured on ai_service but absent from the options map — " +
			"the provider client will silently use its hardcoded 2048 (anthropic.go:109)")
	}
	if got != 8000 {
		t.Fatalf("max_tokens reached the options map as %v, want 8000", got)
	}
}

// The step's own config wins, so a step can raise its budget without editing a
// shared ai_service block. Same precedence as ExecuteAIStepAction.
func TestStepConfigOutranksTheServiceBlockForTheTokenBudget(t *testing.T) {
	stepCfg := map[string]interface{}{"max_tokens": float64(16000)}
	aiCfg := map[string]interface{}{"max_tokens": float64(8000)}

	if got := llmOptionsFromConfig(stepCfg, aiCfg, nil, "test")["max_tokens"]; got != 16000 {
		t.Fatalf("step config max_tokens=16000 lost to ai_service; got %v", got)
	}
}

// Thinking must be opt-in. The client enables extended thinking on the PRESENCE
// of budget_tokens, so forwarding a zero would turn it on for every call.
func TestThinkingIsNotRequestedUnlessBudgeted(t *testing.T) {
	for _, cfg := range []map[string]interface{}{
		nil,
		{"budget_tokens": float64(0)},
	} {
		if _, present := llmOptionsFromConfig(nil, cfg, nil, "test")["budget_tokens"]; present {
			t.Fatalf("budget_tokens forwarded for config %v — the client would enable thinking", cfg)
		}
	}
	if got := llmOptionsFromConfig(nil, map[string]interface{}{"budget_tokens": float64(4000)}, nil, "test")["budget_tokens"]; got != 4000 {
		t.Fatalf("an explicit budget_tokens was dropped; got %v", got)
	}
}

// ---------------------------------------------------------------------------
// The prompt must not misdescribe the gate (owner ruling 2026-08-06)
// ---------------------------------------------------------------------------

// TestGeneratorPromptDoesNotDemandTheCounterCase.
//
// The prompt told the model "The body MUST put the counter-case. A one-sided
// piece is rejected." for four days after the gate stopped rejecting one-sided
// pieces and the owner said he preferred them. A prompt that misstates the gate
// spends the model's effort on a rule nobody enforces and produces the shape the
// owner rejected.
//
// This test fails if that instruction comes back, which is the point: the
// sentence reads as an obviously sensible rule, so it will look like an
// improvement to whoever re-adds it.
func TestGeneratorPromptDoesNotDemandTheCounterCase(t *testing.T) {
	p := buildGeneratorPrompt(4, []exemplar{{Title: "T", Teaser: "Te", Body: "B"}}, nil, nil)

	for _, banned := range []string{
		"A one-sided piece is rejected",
		"MUST put the counter-case",
	} {
		if strings.Contains(p, banned) {
			t.Fatalf("the generator prompt still contains %q — the gate has recorded rather than "+
				"enforced two-sidedness since 2026-08-06 (applyJudgement: one_sided is a note), "+
				"and the owner ruled one-sided is preferred", banned)
		}
	}

	// The criterion that REPLACED it must be stated, or the prompt now describes
	// no fatal judgement criterion at all.
	if !strings.Contains(p, "CONTESTABLE") {
		t.Fatal("the prompt does not ask for a contestable claim — that is the fatal judgement " +
			"criterion (not_contestable) that took two-sidedness's place")
	}
}

// The exemplars are the specification (PLAN §4). If they do not reach the
// prompt verbatim, the model is imitating whatever prose is hardcoded here
// instead of what the site actually publishes — which is what the old prompt
// did, and it described the corpus wrongly.
func TestGeneratorPromptCarriesTheExemplarsVerbatim(t *testing.T) {
	ex := []exemplar{{
		Title:  "Group chats replaced friendship maintenance",
		Teaser: "Presence without effort. The bar has never been lower.",
		Body:   "A distinctive body that exists only in this test fixture.",
	}}

	p := buildGeneratorPrompt(4, ex, nil, nil)

	for _, want := range []string{ex[0].Title, ex[0].Teaser, ex[0].Body} {
		if !strings.Contains(p, want) {
			t.Fatalf("the prompt does not carry the exemplar %q — the corpus is the specification, "+
				"and a prompt that drops it specifies nothing", want)
		}
	}
}

// TestNoProvocationActionCallsAModelWithAnEmptyOptionsMap binds the CALL SITES.
//
// Without this, the three tests above are vacuous in the way that matters: they
// prove `llmOptionsFromConfig` computes the right map, and would all still pass
// if both actions went back to handing `GenerateText` an empty one — which is
// exactly the bug. The helper being correct and the helper being USED are
// independent facts, and only the second one sends a token budget to the API.
//
// A source scan is a blunt instrument and this estate has been bitten by making
// comments load-bearing, so it is deliberately narrow: only lines that actually
// call GenerateText, only in the two provocation actions, matching the literal
// empty map. Prose about empty options maps lives on its own lines and in
// llm_options.go, neither of which is scanned.
//
// > **NOTE 2026-09-03 — this test is now PARTLY subsumed, and kept on purpose.**
// > `llm_budget_call_sites_test.go` enforces the same rule across the whole
// > package, on the AST, and additionally refuses a hardcoded budget literal. It
// > exists because the two-named-files scope here could not see two callers
// > written in August that each hardcoded 2000 (`bugs_open/257` §2026-09-03b).
// > What this test still adds is the per-FILE alarm below: it fails if EITHER of
// > these two files stops calling GenerateText at all, which the package-wide
// > version cannot notice among ten call sites. The empty-map half is redundant.
// > Retire this deliberately if you disagree — do not let it rot silently.
func TestNoProvocationActionCallsAModelWithAnEmptyOptionsMap(t *testing.T) {
	for _, file := range []string{
		"provocation_generator_action.go",
		"provocation_gate_action.go",
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		var sawCall bool
		for i, line := range strings.Split(string(src), "\n") {
			if !strings.Contains(line, "GenerateText(") {
				continue
			}
			sawCall = true
			if strings.Contains(line, "map[string]interface{}{}") {
				t.Errorf("%s:%d calls GenerateText with an empty options map:\n\t%s\n"+
					"Since bugs_open/257 that no longer pins the call to 2048 — the client "+
					"resolves ai_service.max_tokens from its own construction config "+
					"(aiservice.DefaultMaxOutputTokens is the floor). But an empty map still "+
					"drops this STEP's own max_tokens and its budget_tokens, neither of which "+
					"a client constructor ever sees. Pass llmOptionsFromConfig(...) instead.",
					file, i+1, strings.TrimSpace(line))
			}
		}
		// A scan that finds nothing to check passes for the wrong reason — if the
		// call moves or is renamed, this test must stop claiming to guard it.
		if !sawCall {
			t.Fatalf("%s contains no GenerateText call — this test is no longer "+
				"watching anything; find where the model is called and repoint it", file)
		}
	}
}

// TestGeneratorPromptRulesOutSpecialistVocabulary.
//
// Owner, 2026-08-10, rejecting a gate-approved candidate on its title alone:
// *"noone knows what scales means - that's a techie term"*. The gate had passed it
// — correctly, by its own rules, which say nothing about register — so this is a
// generator-side rule or it is nothing.
//
// It is deliberately a rule about the READER, not a banned-word list. The plain-copy
// work of 2026-07-28 ended on exactly this point: a phrase-level patch resurfaces in
// new grammar the following round, because the instinct is what repeats, not the
// word. Banning "scales" would have produced "compounds" or "does not degrade at
// volume" next time.
func TestGeneratorPromptRulesOutSpecialistVocabulary(t *testing.T) {
	p := buildGeneratorPrompt(4, []exemplar{{Title: "T", Teaser: "Te", Body: "B"}}, nil, nil)

	// The rule must be stated as a property of the reader's own speech. If this
	// check is what fails, someone has narrowed it to a word list.
	if !strings.Contains(p, "would this reader have said it") {
		t.Fatal("the prompt no longer states the register rule as a test about the reader — " +
			"a banned-word list was tried before and the instinct simply reappears in new " +
			"vocabulary (plain-copy work, 2026-07-28 round 3)")
	}
	// The worked example is what makes an abstract rule actionable, and it is the
	// owner's own rejection.
	if !strings.Contains(p, "scales") {
		t.Fatal("the prompt dropped the worked example the rule came from")
	}
}

// TestGeneratorPromptAsksForBritishEnglish.
//
// The generator produced "Long showers are self-care theater" on 2026-08-10 — an
// American spelling on a British site. Nothing anywhere in the pipeline was asking
// for British English: CLAUDE.md states it as a platform convention, the gate does
// not judge it, and a convention no prompt carries is a convention the model has
// never been told about.
//
// It sits with the register rule on purpose. Both are the same failure from the
// reader's side — a word that says the writing was not meant for them.
func TestGeneratorPromptAsksForBritishEnglish(t *testing.T) {
	p := buildGeneratorPrompt(4, []exemplar{{Title: "T", Teaser: "Te", Body: "B"}}, nil, nil)

	if !strings.Contains(p, "BRITISH ENGLISH") {
		t.Fatal("the generator prompt does not ask for British English — the platform " +
			"convention is stated in CLAUDE.md and enforced by nothing the model can read")
	}
	// The worked pair the defect came from. Examples are what make a spelling rule
	// actionable; "use British English" alone did not exist anywhere and would have
	// been just as absent in effect.
	if !strings.Contains(p, "theatre not theater") {
		t.Fatal("the prompt dropped the worked spelling example the rule came from")
	}
}

// ---------------------------------------------------------------------------
// Readability (owner, 2026-08-11: "readable by a 5 year old or something like that")
// ---------------------------------------------------------------------------

// TestReadabilityCatchesTheEntryThatPromptedTheRule uses the real text, because a
// threshold tuned against invented prose is tuned against nothing.
func TestReadabilityCatchesThePoolsWorstEntry(t *testing.T) {
	// Verbatim from `cooking-from-scratch-every-night-isnt-worth-it`, the generator's
	// own output and the worst-scoring body in the pool on 2026-08-11 (34.5 words
	// per sentence). If the rail does not catch this, it catches nothing.
	body := "The hour you spend chopping vegetables after work is an hour you are not " +
		"spending with the people you did all of it for, and the meal that results is " +
		"rarely so much better than the one you could have assembled in fifteen minutes " +
		"that the difference repays what it cost you to make it."

	r := measureReadability(body)
	if len(r.Failures) == 0 {
		t.Fatalf("the rail passed the pool's worst entry: grade %.1f, %.1f words/sentence, "+
			"longest %d words", r.Grade, r.AvgWords, r.LongestWords)
	}
	if r.LongestWords <= maxSentenceWords {
		t.Fatalf("longest sentence measured %d words, which is under the %d limit — "+
			"the sentence splitter is breaking where it should not",
			r.LongestWords, maxSentenceWords)
	}
}

// The other direction, and the one that decides whether the rail is usable: short
// plain sentences must PASS. A check that fails everything gets switched off, and
// then it protects nothing.
func TestReadabilityPassesPlainProse(t *testing.T) {
	body := "You clean the bath when someone is coming. You do not clean it for you. " +
		"That tells you who the house is for. It is not you. Think about that."

	r := measureReadability(body)
	if len(r.Failures) != 0 {
		t.Fatalf("plain prose failed the rail: %v (grade %.1f, %.1f words/sentence)",
			r.Failures, r.Grade, r.AvgWords)
	}
}

// THE RAIL IS NOW FATAL (2026-09-02). This test was `TestReadabilityIsAdvisoryNotFatal`
// and it asserted the opposite; it failed when the rail was flipped, which is exactly
// what it was built to do — the flip is a one-line change and the failing test is what
// forces whoever makes it to come and read the reasoning.
//
// The two conditions the old test demanded were both met before flipping: a run showing
// the prompt can pass (8/8 across two rounds on 2026-08-12, against 0/28 for the older
// pool), and a reason to need it — the owner removed the human-approval stamp from the
// publish path in the same change, so this rail is now the only non-stochastic check
// between the generator and the live site.
//
// If you are here because this test failed, you have made the rail advisory again.
// Before doing that, note what else changed on 2026-09-02: nobody reads the text before
// it is served. An advisory note in that world is recorded and never read by anyone.
func TestReadabilityIsFatal(t *testing.T) {
	v := gateVerdict{}
	checkReadability(provocationCandidate{
		Body: "This sentence is deliberately far longer than twenty words in order to " +
			"trip the readability rail and prove that doing so produces a rejection " +
			"rather than a mere note on the candidate under judgement here.",
	}, &v)

	if !v.fatal() {
		t.Fatal("checkReadability did NOT produce a fatal reason. Since 2026-09-02 this " +
			"rail is the only deterministic thing standing between the generator and the " +
			"live site: the human-approval stamp was removed from the publish path the " +
			"same day, and the LLM judge is documented-stochastic on this corpus. If you " +
			"made it advisory again, say so in the owner's log — it is a change to what " +
			"the site can publish unread, not a tuning tweak.")
	}
	var found bool
	for _, r := range v.Reasons {
		if r.Rule == "hard_to_read" {
			found = true
		}
	}
	if !found {
		t.Fatal("a 30-word sentence produced no hard_to_read reason — the rail is not wired in")
	}
}

// A candidate that READS EASILY must still pass cleanly now the rail rejects. Without
// this arm the test above passes for a rail that fails everything, which would starve
// the site rather than filter it — the exact failure the 2026-08-11 comment predicted.
func TestReadabilityFatalDoesNotRejectPlainProse(t *testing.T) {
	v := gateVerdict{}
	checkReadability(provocationCandidate{
		Body: "You did not catch up on rest. You gave yourself jet lag for free. " +
			"The lie-in feels kind. It shifts your body clock by two hours. " +
			"On Monday you wake in the dark and call it a bad night.",
	}, &v)

	if v.fatal() {
		t.Fatalf("plain short prose was REJECTED by the rail: %v.\n"+
			"A rail that rejects everything is indistinguishable from a broken generator "+
			"and starves the site. Check the thresholds before shipping this.", v.Reasons)
	}
}

// The exemplar loop must pull UPWARD. Ordering by date made this round's output the
// next round's specification, and on 2026-08-11 that meant the pool's worst entry was
// teaching the model. Excluding failures is what makes it a ratchet rather than a
// ranking of a bad set.
func TestExemplarSelectionExcludesRatherThanRanks(t *testing.T) {
	src, err := os.ReadFile("provocation_generator_action.go")
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	s := string(src)
	if strings.Contains(s, "ORDER BY publish_on DESC NULLS LAST\n\t\t LIMIT $2") {
		t.Fatal("loadExemplars is selecting exemplars by DATE again. That makes each " +
			"round's output the next round's specification with nothing pulling upward — " +
			"measured 2026-08-11, it fed the model the worst-written entry in the pool.")
	}
	if !strings.Contains(s, "if len(r.Failures) > 0 {") {
		t.Fatal("loadExemplars no longer EXCLUDES exemplars that fail the readability " +
			"bar. Ranking alone returns the best of a bad set — which is exactly the " +
			"state the pool was in when every one of its 28 entries failed.")
	}
}
