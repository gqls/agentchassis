// FILE: platform/orchestration/datahelpers/negationtells_test.go
//
// Corpus rule: every TRIP case is copy that actually shipped (the owner's own
// quoted sentences, and live page_components content from
// ai-agent-orchestration.com), and every PASS case is prose from the same pages
// that a reader would not object to. A scanner tested only on invented examples
// is tested against its own author's idea of the fault.
//
// Mutation checks (run by hand, recorded in the lane's NOTES):
//   - delete negXNotYRe from negationShapes      -> TestOwnersOwnSentencesTrip fails
//   - delete the neighbour comparison in AcceptNegationRewrite -> TestRewriteRejectsDisplacement fails
//   - make NegationExempt take the whole prompt   -> TestExemptionIsSentenceScoped fails
//   - delete the dropped_name loop                -> TestRewriteRejectsDroppedName fails
//   - key dropped_name on the whole `from` instead of from[:protectFrom]
//                                                 -> TestRewriteRejectsDroppedName fails
//   - set headingMinWords to sentenceMinWords     -> TestHeadingFloorIsSeparateFromTheSentenceFloor fails

package datahelpers

import (
	"strings"
	"testing"
)

func shapesOf(hits []NegationHit) map[string]int {
	m := map[string]int{}
	for _, h := range hits {
		m[h.Shape]++
	}
	return m
}

// The two sentences the owner quoted. If either stops tripping, this bug is
// reopened by definition.
func TestOwnersOwnSentencesTrip(t *testing.T) {
	cases := []struct {
		name, text, want string
	}{
		{"headline", "170+ agents defined, 174 agent types. The registry shows you what's possible, not what survives production.", "x_not_y"},
		{"two_sentence_reveal", "A model directory tells you which agents exist. It doesn't tell you how they hold up under real Kafka throughput.", "negative_reveal"},
		{"hero_tagline", "Multi-agent systems deployed to production in days, not months on Kubernetes, Kafka, and Postgres", "x_not_y"},
		{"staccato", "Not a demo environment. Not a proof of concept. If the work fits, we can show you a running example.", "staccato"},
		{"comma_but", "This is not just a framework, but a system that ships.", "not_x_but_y"},
		{"rather_than", "The reader is persuaded rather than sold to.", "rather_than"},
	}
	for _, c := range cases {
		got := shapesOf(ScanDefineByNegation(c.text))
		if got[c.want] == 0 {
			t.Errorf("%s: expected shape %q to trip, got %v", c.name, c.want, got)
		}
	}
}

// The negative arm. Real sentences from the same three pages, none of which
// defines by negation. A scanner that fires here is unusable at 215 sections a
// day, whatever it catches.
func TestCleanProsePasses(t *testing.T) {
	clean := []string{
		"This directory catalogues every agent definition running in our production fleet, organised across our 8 internal departments.",
		"We run over 1,600 orchestrations a day across 13 live production systems on Kubernetes, Kafka, and Postgres.",
		"Agent counts move as agents are added or retired, so what you see here is a snapshot.",
		"Rolling several debts into one loan usually brings the monthly payment down, and that is normally why people do it.",
		"Paying extra off a loan saves you interest, because interest is charged on what you still owe.",
		"We support version 1.0 and e.g. the older protocol revisions too.",
		"Book a call to walk through your pipeline with the people who run it.",
	}
	for _, s := range clean {
		if hits := ScanDefineByNegation(s); len(hits) > 0 {
			t.Errorf("clean prose tripped %v: %q", shapesOf(hits), s)
		}
	}
}

// A statement of policy or limit is not the mannerism. The house voice asks for
// these ("name the limit, the failure mode, or what the thing cannot do"), and a
// check that flagged them would send a human to edit a company's stated policy.
// Found by running the fleet check on live briefs, not by reading the regex.
func TestFirstPersonLimitStatementsAreNotTheMannerism(t *testing.T) {
	for _, s := range []string{
		"We do not offer refunds.",
		"We do not invent figures and we do not publish prices we cannot source.",
		"We do not charge for the first call.",
		"We cannot tell you which lender will accept you.",
	} {
		if hits := ScanDefineByNegation(s); len(hits) > 0 {
			t.Errorf("a first-person limit statement must not trip: %q -> %v", s, shapesOf(hits))
		}
	}
	// The owner's own example still trips: it negates a THING, not a policy.
	if hits := ScanDefineByNegation("A model directory tells you which agents exist. It doesn't tell you how they hold up."); shapesOf(hits)["negative_reveal"] == 0 {
		t.Error("the third-person reveal must still trip — that is the owner's second quoted sentence")
	}
	if hits := ScanDefineByNegation("This is a working registry. It is not a vendor roadmap."); shapesOf(hits)["negative_reveal"] == 0 {
		t.Error("a third-person reveal about a thing must trip")
	}
}

// A factual comparison must not read as the construction — this is the arm that
// keeps the neighbour set from firing on ordinary numbers.
func TestNeighbourSetIgnoresFactualComparison(t *testing.T) {
	if hits := ScanContrastNeighbours("We run more than 30 agents in production."); len(hits) > 0 {
		t.Errorf("factual 'more than 30' counted as a neighbour: %v", shapesOf(hits))
	}
	if hits := ScanContrastNeighbours("This is more than just a catalogue."); shapesOf(hits)["more_than_just"] != 1 {
		t.Errorf("'more than just' should be a neighbour, got %v", shapesOf(hits))
	}
	// The neighbour set must NEVER be a trip. (`instead_of` no longer proves
	// this — owner Decision B, 2026-08-31, promoted it to a tripping shape, so
	// the never-a-trip control uses a remaining neighbour.)
	if hits := ScanDefineByNegation("Unlike a spreadsheet, this updates itself."); len(hits) > 0 {
		t.Errorf("a neighbour shape must not trip the gate: %v", shapesOf(hits))
	}
	// And the promotion itself is pinned: instead_of now TRIPS.
	if hits := ScanDefineByNegation("We ship it instead of talking about it."); shapesOf(hits)["instead_of"] != 1 {
		t.Errorf("instead_of must trip as a shape under Decision B, got %v", shapesOf(hits))
	}
	if hits := ScanDefineByNegation("This helps everyone, not just engineers, ship faster; it is not merely a demo."); shapesOf(hits)["not_just"] == 0 {
		t.Errorf("not_just must trip as a shape under Decision B, got %v", shapesOf(hits))
	}
}

// A hit has to be able to name its own sentence, or the repair cannot splice.
func TestSentenceAttributionAndSplice(t *testing.T) {
	text := "<p>Our agents run in production. The registry shows you what's possible, not what survives production.</p>"
	hits := ScanDefineByNegation(text)
	if len(hits) == 0 {
		t.Fatal("expected a hit")
	}
	h := hits[0]
	want := "The registry shows you what's possible, not what survives production."
	if h.Sentence != want {
		t.Errorf("sentence attribution: got %q want %q", h.Sentence, want)
	}
	// The offsets must address the RAW text, so an exact-substring splice works.
	if text[h.SentenceStart:h.SentenceStart+len(h.Sentence)] != h.Sentence {
		t.Errorf("SentenceStart does not address the raw text")
	}
	if got := h.Sentence[h.MatchInSent : h.MatchInSent+len(h.Matched)]; got != h.Matched {
		t.Errorf("MatchInSent wrong: %q", got)
	}
}

// A rich_text field is paragraphs; the second may carry no full stop at all.
func TestHTMLParagraphsSplitAndTagsTrimmed(t *testing.T) {
	text := "<p>We deploy on Kubernetes</p><p>The registry shows what's possible, not what survives</p>"
	hits := ScanDefineByNegation(text)
	if len(hits) != 1 {
		t.Fatalf("expected exactly one hit, got %d (%v)", len(hits), shapesOf(hits))
	}
	if hits[0].Sentence != "The registry shows what's possible, not what survives" {
		t.Errorf("tag edges not trimmed: %q", hits[0].Sentence)
	}
}

func TestExemptionIsSentenceScoped(t *testing.T) {
	brief := []string{
		"Emphasis: the canonical tagline 'Multi-agent systems deployed to production in days, not months' must appear in the homepage hero.",
	}
	// The whole rendered prompt contains "rather than" six times over; a
	// prompt-wide exemption would exempt this, which is the bug this test pins.
	housePrompt := []string{
		"HOUSE VOICE. The reader is persuaded rather than sold to. Say what a thing IS rather than what it is not.",
	}

	supplied := ScanDefineByNegation("Multi-agent systems deployed to production in days, not months on Kubernetes.")
	if len(supplied) == 0 {
		t.Fatal("expected the tagline to trip")
	}
	if ok, why := NegationExempt(supplied[0], brief); !ok {
		t.Errorf("a brief-supplied tagline must be exempt, got %v/%q", ok, why)
	}

	own := ScanDefineByNegation("The registry shows you what's possible, not what survives production.")
	if ok, _ := NegationExempt(own[0], brief); ok {
		t.Errorf("the writer's own construction must NOT be exempt")
	}

	rt := ScanDefineByNegation("We build the pipeline rather than describing it.")
	if ok, why := NegationExempt(rt[0], housePrompt); ok {
		t.Errorf("house-voice prose must not exempt a 'rather than' hit (%q) — that is the 43%% hole", why)
	}
}

func TestRegulatoryNegationsExempt(t *testing.T) {
	for _, s := range []string{
		"These figures are an estimate, not financial advice.",
		"We are a comparison service, not a lender.",
		"The tool can be wrong, so check the figure with your provider.",
		"This does not constitute a recommendation.",
	} {
		hits := ScanDefineByNegation(s)
		if len(hits) == 0 {
			continue // nothing to exempt
		}
		if ok, why := NegationExempt(hits[0], nil); !ok {
			t.Errorf("regulatory sentence not exempt (%q): %q", why, s)
		}
	}
}

func TestRewriteRejectsDisplacement(t *testing.T) {
	from := "The registry shows you what's possible, not what survives production."
	protect := ScanDefineByNegation(from)[0].MatchInSent

	bad := map[string]string{
		"displaced_instead_of":     "The registry shows you what's possible instead of what survives production.",
		"displaced_more_than_just": "The registry is more than just a list of what's possible.",
		"displaced_em_dash":        "The registry shows what's possible — never what survives production.",
		"still_x_not_y":            "It lists what is possible, not what is proven.",
	}
	for want, to := range bad {
		ok, why := AcceptNegationRewrite(from, to, protect)
		if ok {
			t.Errorf("%s: accepted a displaced rewrite: %q", want, to)
		} else if why != want {
			t.Logf("%s: rejected as %q (acceptable if still a rejection)", want, why)
		}
	}

	good := "The registry lists the agent definitions running in our production fleet today."
	if ok, why := AcceptNegationRewrite(from, good, protect); !ok {
		t.Errorf("a direct rewrite was rejected as %q: %q", why, good)
	}
}

func TestRewriteFactRules(t *testing.T) {
	from := "We run 1,600 orchestrations a day, not 12 demos a week."
	protect := ScanDefineByNegation(from)[0].MatchInSent

	// The figure BEFORE the construction is the claim: losing it is a content loss.
	if ok, why := AcceptNegationRewrite(from, "We run orchestrations across live production systems every day.", protect); ok {
		t.Errorf("dropping the protected figure was accepted (%q)", why)
	}
	// The figure AFTER it is the contrasted alternative: dropping it is the point.
	if ok, why := AcceptNegationRewrite(from, "We run 1,600 orchestrations a day across live production systems.", protect); !ok {
		t.Errorf("dropping the contrasted figure should be allowed, got %q", why)
	}
	// No invention, in either direction.
	if ok, _ := AcceptNegationRewrite(from, "We run 1,600 orchestrations a day for 42 clients.", protect); ok {
		t.Error("an invented figure was accepted")
	}
	if ok, _ := AcceptNegationRewrite("Read the guide at /guides/agents.html, not the marketing page.",
		"Read the guide for the full picture.", 40); ok {
		t.Error("a dropped link was accepted")
	}
	if ok, _ := AcceptNegationRewrite("We deploy on Kubernetes, not on bare metal.",
		"We deploy on Kubernetes and Heroku.", 24); ok {
		t.Error("an invented proper noun was accepted")
	}
	if ok, _ := AcceptNegationRewrite("<p>It shows what is possible, not what survives.</p>",
		"It shows what runs in production today.", 30); ok {
		t.Error("a rewrite that dropped the markup was accepted")
	}
	if ok, why := AcceptNegationRewrite("It shows what is possible, not what survives.",
		"It shows what is possible, not what survives.", 26); ok || why != "unchanged" {
		t.Errorf("an unchanged rewrite must be rejected, got %v/%q", ok, why)
	}
	if ok, why := AcceptNegationRewrite("It shows what is possible, not what survives production today.", "It shows.", 26); ok || why != "gutted" {
		t.Errorf("a gutted rewrite must be rejected, got %v/%q", ok, why)
	}

	// ── OWNER RULING 2026-09-03, relayed via the finetuning.uk lane ──────────
	// "Keeping the first clause and cutting at the comparison is the norm you
	// set. Loosen the judge so a truncation to a complete, true first clause is
	// accepted."
	//
	// These two are the cases that PROVOKED the ruling, and both were refused by
	// the old 40% proportion. The first is his own worked example (29.5%); the
	// second is a live homepage repair the gate rejected the same morning
	// (38.1%), leaving `rather than` on the served page. If either of these ever
	// reads "gutted" again, the floor has been put back and the ruling is lost.
	if ok, why := AcceptNegationRewrite(
		"We're not tied to one provider, so you get the model that fits the task, not the model we happen to sell.",
		"We're not tied to one provider.", 30); !ok {
		t.Errorf("the owner's own ruled truncation must be accepted, got %v/%q", ok, why)
	}
	if ok, why := AcceptNegationRewrite(
		"We pick the tool suited to each task rather than pushing one platform across everything you need.",
		"We pick the tool suited to each task.", 36); !ok {
		t.Errorf("a complete first clause must be accepted, got %v/%q", ok, why)
	}
	// The floor is on WORDS, so markup cannot buy a stub its way through: this
	// is the same 2-word stranded verb dressed in tags.
	if ok, why := AcceptNegationRewrite(
		"<p>It shows what is possible, not what survives production today.</p>",
		"<p><strong>It</strong> <em>shows</em>.</p>", 26); ok || why != "gutted" {
		t.Errorf("a markup-padded stub must still be gutted, got %v/%q", ok, why)
	}
}

// DROPPED_NAME — the LOSE half of the proper-name rule (bugs_open/420).
// Until this existed the guards were asymmetric: figures were protected both
// ways but names only against INVENTION, so a rewrite that DELETED an identifier
// passed everything. That was survivable only while the never-prose field list
// kept identity-bearing fields away from the model; 420 removed bare `name` from
// it, so the protection had to become a control in the judge — reachable by both
// mutating call sites and by any future caller that walks fields itself.
//
// Reasons are asserted, not just the bool: a later edit must not be able to
// satisfy this by tightening `gutted` instead.
func TestRewriteRejectsDroppedName(t *testing.T) {
	from := "We deploy on Kubernetes with Argo, not on bare metal."
	protect := ScanDefineByNegation(from)[0].MatchInSent

	// A name in the PROTECTED half is part of the claim and must survive.
	if ok, why := AcceptNegationRewrite(from, "We deploy with Argo for every service.", protect); ok || why != "dropped_name" {
		t.Errorf("losing the protected proper noun must be refused as dropped_name, got %v/%q", ok, why)
	}
	// The ruled truncation keeps from[:protectFrom] verbatim, so it is inert here.
	if ok, why := AcceptNegationRewrite(from, "We deploy on Kubernetes with Argo.", protect); !ok {
		t.Errorf("the ruled truncation must still be accepted, got %q", why)
	}
	// A name AFTER the construction is the contrasted alternative and may go —
	// this is the half that keying on protectFrom buys, exactly as dropped_figure.
	after := "We deploy on Kubernetes with Argo, not on Heroku."
	if ok, why := AcceptNegationRewrite(after, "We deploy on Kubernetes with Argo.",
		ScanDefineByNegation(after)[0].MatchInSent); !ok {
		t.Errorf("dropping the CONTRASTED name should be allowed, got %q", why)
	}
	// A sentence-initial capital is not a name, on this side as on the other.
	if ok, why := AcceptNegationRewrite("We ship weekly to production, not monthly.",
		"Releases go out weekly to production.", 28); !ok {
		t.Errorf("losing a sentence-initial capital is not losing a name, got %q", why)
	}
}

// The heading floor is a SEPARATE constant, and this test is the control that it
// is separate: the same rewrite must be refused by the sentence judge and
// accepted by the heading judge. Owner ruling 2026-09-03 — the 5-word floor was
// calibrated on body sentences and would have refused 25 of the 36 live heading
// repairs bugs_open/420 exposed, making them visible and then declining to fix
// them.
func TestHeadingFloorIsSeparateFromTheSentenceFloor(t *testing.T) {
	from := "Exact math, not simulation"
	protect := ScanDefineByNegation(from)[0].MatchInSent

	if ok, why := AcceptNegationRewrite(from, "Exact math.", protect); ok || why != "gutted" {
		t.Errorf("the SENTENCE floor must still refuse a 2-word survivor, got %v/%q — if this passes, the sentence floor has been loosened and that is the owner's ruling from 2026-09-03 being undone", ok, why)
	}
	if ok, why := AcceptNegationHeadingRewrite(from, "Exact math.", protect); !ok {
		t.Errorf("a heading is complete at two words; the heading judge must accept it, got %q", why)
	}
	// The proportional backstop still applies to BOTH arms, so the heading floor
	// is not a way round every size check.
	long := "A catalog with the figures behind it and the method stated in full, not just names"
	if ok, why := AcceptNegationHeadingRewrite(long, "A catalog.", ScanDefineByNegation(long)[0].MatchInSent); ok {
		t.Errorf("the 25%% backstop must still refuse a long original reduced to a stub, got %v/%q", ok, why)
	}
}

func TestShapeVocabularyIsStable(t *testing.T) {
	// Decision B (owner, 2026-08-31) added the last two. The register entry
	// (CQ-026) and the LANDMINES prose that named FIVE shapes are corrected in
	// the same commit as this line — that is what this ratchet exists to force.
	want := []string{"x_not_y", "not_x_but_y", "staccato", "rather_than", "negative_reveal", "instead_of", "not_just"}
	got := NegationShapeNames()
	if len(got) != len(want) {
		t.Fatalf("shape vocabulary changed: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("shape %d: got %q want %q — the migration, the register entry and the LANDMINES text all name these", i, got[i], want[i])
		}
	}
}

// Tag COUNTS equal is not markup preserved: inverted nesting keeps every tag and
// produces malformed HTML. (Council round 1, render_guardian seat.)
func TestRewriteRejectsInvertedNesting(t *testing.T) {
	from := "<b><i>It shows what is possible, not what survives.</i></b>"
	if ok, why := AcceptNegationRewrite(from, "<b><i>It shows what runs in production today.</b></i>", 30); ok {
		t.Errorf("inverted nesting was accepted (%q) — tag order, not tag counts", why)
	}
	if ok, why := AcceptNegationRewrite(from, "<b><i>It shows what runs in production today.</i></b>", 30); !ok {
		t.Errorf("well-formed markup preserved must be accepted, got %q", why)
	}
}

// The two-sentence reveal must be attributed to the sentence that CARRIES the
// negation, not to the clean sentence before it whose full stop the pattern
// happens to start at. Found by running the scanner over the owner's own live
// page: the repair would have been handed "A model directory tells you which
// agents exist." — a true, clean sentence — while the reveal stayed put.
func TestNegativeRevealAttributesToTheSentenceThatMustChange(t *testing.T) {
	text := "A model directory tells you which agents exist. It doesn't tell you how they hold up under real Kafka throughput."
	var reveal *NegationHit
	for _, h := range ScanDefineByNegation(text) {
		if h.Shape == "negative_reveal" {
			hh := h
			reveal = &hh
		}
	}
	if reveal == nil {
		t.Fatal("expected a negative_reveal hit")
	}
	if !strings.HasPrefix(reveal.Sentence, "It doesn't tell you") {
		t.Errorf("attributed to the wrong sentence: %q", reveal.Sentence)
	}
	// And the invariant the splice depends on must still hold.
	if got := reveal.Sentence[reveal.MatchInSent : reveal.MatchInSent+len(reveal.Matched)]; got != reveal.Matched {
		t.Errorf("MatchInSent no longer addresses Matched within Sentence: %q vs %q", got, reveal.Matched)
	}
	if text[reveal.SentenceStart:reveal.SentenceStart+len(reveal.Sentence)] != reveal.Sentence {
		t.Error("SentenceStart no longer addresses the raw text")
	}
}

// "Say what it IS" is the pressure that fills the slot the removed contrast
// leaves with an absolute. checkBannedClaims only catches patterns a site has
// ARMED, and the register is sparse — so an unarmed site would have had nothing
// between an invented superlative and the page. (Council round 4, compliance.)
func TestRewriteRejectsInventedSuperlatives(t *testing.T) {
	from := "The registry shows you what's possible, not what survives production."
	protect := ScanDefineByNegation(from)[0].MatchInSent
	for _, to := range []string{
		"The registry is the definitive record of what runs in production.",
		"The registry shows every single agent running in production.",
		"The registry gives you a fully verified view of production.",
		"The registry is always accurate about what runs in production.",
	} {
		if ok, why := AcceptNegationRewrite(from, to, protect); ok {
			t.Errorf("an invented absolute was accepted (%q): %q", why, to)
		}
	}
	// The word was the AUTHOR's, so keeping it is not an invention.
	keeps := "It always lists what is possible, not what survives."
	if ok, why := AcceptNegationRewrite(keeps, "It always lists what runs in production today.",
		ScanDefineByNegation(keeps)[0].MatchInSent); !ok {
		t.Errorf("a superlative the original already carried must not be a rejection, got %q", why)
	}
	// And a plain rewrite still passes.
	if ok, why := AcceptNegationRewrite(from, "The registry lists the agent definitions running in production today.", protect); !ok {
		t.Errorf("a plain rewrite was rejected as %q", why)
	}
}

// A sentence span must close at a table CELL boundary, header cells included.
//
// `</td` was in the boundary list from the start and `</th` was not, because
// "</th>" does not match the `</h` heading arm — it is `<`,`/`,`t`,`h`, so the
// third character already differs. The consequence was not a missed hit but a
// CORRUPT one: a header row scanned as a single sentence CONTAINING RAW MARKUP
// ("Real, not simulated</th><th>Throughput"). The captured sentence is exactly
// what a repair splices over, so a rewrite of that span would have replaced the
// cell tags with prose and broken the table.
//
// Reachable because migrations 594/595 retype five pass-through prose slots to
// `html` and instruct page-content-writer to emit <table> in them.
func TestSentenceSpansCloseAtEveryTableCellBoundary(t *testing.T) {
	for _, tc := range []struct{ name, html, want string }{
		{"header cells", "<table><tr><th>Real, not simulated</th><th>Throughput</th></tr></table>", "Real, not simulated"},
		{"data cells", "<table><tr><td>Ships in days, not months</td><td>Kubernetes</td></tr></table>", "Ships in days, not months"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hits := ScanDefineByNegation(tc.html)
			if len(hits) != 1 {
				t.Fatalf("got %d hits, want 1: %+v", len(hits), hits)
			}
			if hits[0].Sentence != tc.want {
				t.Errorf("sentence = %q, want %q", hits[0].Sentence, tc.want)
			}
			if strings.Contains(hits[0].Sentence, "<") || strings.Contains(hits[0].Sentence, ">") {
				t.Errorf("sentence carries RAW MARKUP (%q) — a repair splices over this span and would eat the tags", hits[0].Sentence)
			}
		})
	}
}

// The peer question that prompted the above (bugs_open/381, 2026-08-24): does the
// scanner split on tag boundaries or on punctuation only? On BOTH — so a run of
// list items with no full stops is many sentences, not one. Pinned because the
// 594/595 pair makes lists and subheads common in these slots for the first time.
func TestListItemsWithoutTerminatorsAreSeparateSentences(t *testing.T) {
	html := "<ul><li>Built for production, not demos</li><li>Kubernetes native</li><li>Kafka backed</li></ul>"
	hits := ScanDefineByNegation(html)
	if len(hits) != 1 {
		t.Fatalf("got %d hits, want exactly 1 — the construction is in the FIRST item only: %+v", len(hits), hits)
	}
	if hits[0].Sentence != "Built for production, not demos" {
		t.Errorf("sentence = %q — it bled across </li>, so the later items joined the first", hits[0].Sentence)
	}
}
