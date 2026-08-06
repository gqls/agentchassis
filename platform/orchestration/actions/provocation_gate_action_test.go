// FILE: platform/orchestration/actions/provocation_gate_action_test.go
//
// CALIBRATION — and PLAN §10.6 is explicit that this file is not "good practice"
// but "the only evidence the gate works at all", because the owner removed the
// human approver who would otherwise have been the backstop.
//
// The bar it sets: the gate must pass all 9 real provocations AND reject a
// deliberately bad set — a bare insult, a factual claim dressed as opinion, a
// one-sided political take, and a piece of trending slop.
//
// The nine below are the REAL corpus, copied from the live pool on 2026-08-05
// (domain vonc.com, status approved). They are the specification, so they are
// reproduced verbatim rather than paraphrased: a paraphrase would calibrate the
// gate against my idea of a provocation instead of against the owner's.
//
// WHAT THIS FILE DELIBERATELY DOES NOT DO: it does not call a model. The judge is
// stubbed, so these tests pin the DETERMINISTIC layers and the fail-closed
// wiring. A model's verdicts are not reproducible and a test that depended on
// them would fail for reasons unrelated to the code. The model's own calibration
// is a separate, live exercise — `cmd/provocation-gate-calibrate` — and it must
// be run before the gate is wired to anything that publishes.

package actions

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// ---------------------------------------------------------------------------
// The real corpus (live pool, 2026-08-05)
// ---------------------------------------------------------------------------

// realProvocations is GENERATED MECHANICALLY from the live pool, never typed.
//
// OWNER RULING 2026-08-06: "we want it all to be done through the framework, so we
// don't want you writing things yourself."
//
// The previous version of this block claimed to be "reproduced verbatim" and was
// not: eight of the nine live rows have an empty `body` column, so I had composed
// long-form bodies — complete with tidy "The counter is..." turns — and then tested
// whether the gate detects counter-cases. It passed 9/9 against prose I had written
// to contain the very property under test. The live calibration scored 4/9 on the
// same nine provocations and is what exposed it (WRONG_CALLS.md 2026-08-05).
//
// So this is now the output of:
//
//	SELECT slug, title, teaser, COALESCE(NULLIF(body,''), COALESCE(detail_body,''))
//	  FROM provocations WHERE domain='vonc.com' AND status='approved' ORDER BY publish_on;
//
// serialised straight to Go. The COALESCE mirrors loadGateCandidates exactly, so the
// judged text here is the judged text in production. A comment asserting "verbatim"
// is not a control on verbatimness; generating the fixture is.
//
// NOTE the first entry: `group-chats-replaced-friendship` has a body of ZERO
// characters in the production pool. It is kept, not quietly dropped — see
// TestGateAcceptsTheRealProvocations, which treats it as the pool defect it is.
var realProvocations = []provocationCandidate{
	{
		Slug:   "group-chats-replaced-friendship",
		Title:  "Group chats replaced friendship maintenance",
		Teaser: "Presence without effort. The bar has never been lower.",
		Body:   "",
	},
	{
		Slug:   "nobody-reads-terms-of-service",
		Title:  "Nobody actually reads terms of service — and that's rational",
		Teaser: "The cost of reading outweighs the power to change anything.",
		Body:   "Reading takes an hour. Understanding takes a lawyer. Refusing takes the service away. Given those three prices, not reading is the correct decision, and every study that frames it as apathy has mistaken a rational calculation for a character flaw.\nWhich moves the burden somewhere else entirely. If consent is only ever given unread, then consent is not the thing doing the work, and we should stop pretending that it is.",
	},
	{
		Slug:   "four-day-week-productivity-myth",
		Title:  "The four-day week is a productivity myth",
		Teaser: "The pilots that prove it were self-selected true believers.",
		Body:   "The pilots recruit organisations that already believed, run them for six months with everyone watching, and measure self-reported output. That is a design which cannot return a negative result. It tells you what motivated people do under observation, not what a four-day week does.\nThe counter is that the effect may well be real regardless, and demanding a hostile trial of something people obviously want is its own motivated reasoning. Possibly. Run it on a sceptical workforce for two years and the argument ends.",
	},
	{
		Slug:   "fiction-makes-you-worse-at-facts",
		Title:  "Reading fiction makes you worse at facts",
		Teaser: "Narrative trains you to want a tidy arc. Reality doesn't have one.",
		Body:   "A novel teaches you to expect that events connect, that behaviour has motive, and that the ending explains the beginning. None of that is true of a pandemic, an election or a market. The better you get at narrative, the more confidently you impose one.\nAgainst that: fiction is the main way most people practise holding a mind that is not their own, which is hardly nothing when the facts in dispute are about other people. Perhaps the trade is worth making. But it is a trade, and it is almost always sold as a free gain.",
	},
	{
		Slug:   "data-driven-decisions-arent",
		Title:  "Most 'data-driven' decisions aren't",
		Teaser: "The numbers get picked after the gut already chose.",
		Body:   "Watch the sequence. Someone forms a view, then commissions the analysis, then reads the analysis for the part that agrees. The dashboard is not an input to the decision. It is the receipt.\nThe defence is that this still beats nothing — that even a motivated search for evidence occasionally turns up the number that stops you. Fair enough. But then say that is what the dashboard is for, and stop calling the output data-driven.",
	},
	{
		Slug:   "privacy-is-already-over",
		Title:  "Privacy is already over",
		Teaser: "You traded it years ago. The fight now is who profits.",
		Body:   "You cannot claw back a decade of location history, contact graphs and purchase records by changing a setting. The data exists, it has been copied, and the copies are the asset. Every privacy control shipped since governs what happens next, never what already happened.\nSo the honest question stops being whether privacy survives and becomes who is permitted to profit from its absence. That is a distribution argument rather than a technical one, and it has an entirely different set of winners.",
	},
	{
		Slug:   "remote-work-killed-mentorship",
		Title:  "Remote work killed mentorship",
		Teaser: "You can't absorb judgement over a video call.",
		Body:   "Judgement is not transferred in meetings. It is absorbed in the two minutes after one — the aside, the raised eyebrow, the way someone rewrites your paragraph while you watch. None of those moments has an agenda item, so no scheduled call contains them.\nThe rebuttal is that this was always a story senior people told about their own value. Plenty of people learned their craft alone, from documents, badly lit, and turned out fine. So which is it: a genuine transmission loss, or nostalgia for the office as a stage?",
	},
	{
		Slug:   "ai-never-funny-on-purpose",
		Title:  "AI will never be funny on purpose",
		Teaser: "The machine can recombine a million jokes and still not know why any land.",
		Body:   "A model can hold every joke ever written and still not know which one to tell. Humour is a social risk instrument: it needs a target, a shared assumption to break, and a real chance of the room going cold. A system tuned never to offend and never to fail has removed all three ingredients before it starts.\nThe counter-case is that funniness is only a pattern in the data, and the machine is a better pattern-finder than you are. If that holds, the failure is temporary and the punchlines improve. If it does not, then everything an AI has ever produced that made you laugh was written by a person it read.",
	},
	{
		Slug:   "nobody-wants-personalised-internet",
		Title:  "Nobody actually wants a personalised internet",
		Teaser: "What gets sold as personalisation is the quiet removal of what you'd have shared with a stranger.",
		Body:   "Every feed is tuned to one person, and every conversation now opens with “have you seen” and closes with a shrug. What gets sold as personalisation is mostly the quiet removal of whatever you would have had in common with a stranger. The engine is not serving you — it is dividing the room so each half can be sold separately.",
	},
}

// ---------------------------------------------------------------------------
// The deliberately bad set (PLAN §10.6 names these four kinds)
// ---------------------------------------------------------------------------

var badProvocations = []struct {
	name      string
	c         provocationCandidate
	wantRule  string // the rule that MUST fire; "" = any deterministic rejection
	judgeSays *judgement
}{
	{
		name: "a bare insult",
		c: provocationCandidate{
			Slug: "bad-insult", Title: "People who use tabs are idiots",
			Teaser: "They are stupid and everyone knows it, no argument needed.",
			Body:   strings.Repeat("They are simply morons and anyone defending them is also a moron. ", 6),
		},
		judgeSays: &judgement{Safe: false, TwoSided: false, Contestable: true, OrdinaryExp: true, Note: "abusive"},
		wantRule:  "unsafe",
	},
	{
		name: "a factual claim dressed as opinion",
		c: provocationCandidate{
			Slug: "bad-factual", Title: "Remote work destroyed British productivity",
			Teaser: "The figures since 2020 settle what used to be a matter of taste.",
			Body: "Office attendance fell by exactly 62 percent after 2020 and national output fell with " +
				"it, according to the 2023 Whitfield Institute study of 41,000 firms. That is not an " +
				"opinion, it is the record. " +
				"Some argue the causation runs the other way, but the study controlled for it.",
		},
		judgeSays: &judgement{
			Safe: true, TwoSided: true, Contestable: true, OrdinaryExp: true,
			FactualProblems: []struct {
				Quote  string `json:"quote"`
				Reason string `json:"reason"`
			}{{Quote: "the 2023 Whitfield Institute study of 41,000 firms", Reason: "no such study; invented source and quantity"}},
		},
		wantRule: "factual_problem_in_body",
	},
	{
		name: "a one-sided political take",
		c: provocationCandidate{
			Slug: "bad-political", Title: "The Labour government has ruined the economy",
			Teaser: "Every measure since the election has made ordinary households poorer.",
			Body:   strings.Repeat("This is the predictable result of left-wing policy and nobody serious disputes it. ", 5),
		},
		// Never reaches the judge: the deterministic tribal rule fires first.
		wantRule: "tribal_political",
	},
	{
		name: "trending slop",
		c: provocationCandidate{
			Slug: "bad-slop", Title: "AI is changing everything",
			Teaser: "The pace of change is unprecedented and we must all adapt now.",
			Body:   strings.Repeat("Artificial intelligence is transforming every industry at unprecedented speed. ", 5),
		},
		judgeSays: &judgement{Safe: true, TwoSided: false, Contestable: false, OrdinaryExp: true, Note: "a trend nobody disputes; nothing to argue against"},
		wantRule:  "not_contestable",
	},
}

// aGoodCandidate returns a real provocation that clears every deterministic layer,
// for tests whose subject is the JUDGE rather than the corpus.
//
// It exists because `realProvocations[0]` used to serve this purpose and stopped:
// regenerating the fixture from the pool (ORDER BY publish_on) moved
// `group-chats-replaced-friendship` — the row with a ZERO-character body — into
// index 0, and seven fail-closed tests started rejecting on `body_too_short`
// before the judge was ever consulted. They still passed as "rejected", which is
// the dangerous part: a test asserting "this is refused" cannot tell refusal-for-
// the-right-reason from refusal-for-an-unrelated-one unless it checks the reason.
// Ours did check, so they failed loudly instead of silently testing nothing.
//
// A positional reference into data generated from a live table is a silent
// dependency on that table's ordering. This is the fix: state the property.
func aGoodCandidate(t *testing.T) provocationCandidate {
	t.Helper()
	for _, c := range realProvocations {
		if len(strings.TrimSpace(c.Body)) >= minBodyLen {
			return c
		}
	}
	t.Fatal("no provocation in the live corpus clears the deterministic layers; " +
		"every judge-focused test below would be testing the form rules instead")
	return provocationCandidate{}
}

// ---------------------------------------------------------------------------
// Judge stubs
// ---------------------------------------------------------------------------

func judgeReturning(j judgement) judgeFn {
	return func(context.Context, string) (string, error) {
		b, _ := marshalJudgement(j)
		return b, nil
	}
}

func marshalJudgement(j judgement) (string, error) {
	type fp struct {
		Quote  string `json:"quote"`
		Reason string `json:"reason"`
	}
	probs := make([]fp, 0, len(j.FactualProblems))
	for _, p := range j.FactualProblems {
		probs = append(probs, fp{p.Quote, p.Reason})
	}
	var b strings.Builder
	b.WriteString(`{"safe":`)
	b.WriteString(boolStr(j.Safe))
	b.WriteString(`,"two_sided":`)
	b.WriteString(boolStr(j.TwoSided))
	b.WriteString(`,"contestable":`)
	b.WriteString(boolStr(j.Contestable))
	b.WriteString(`,"arguable_from_ordinary_experience":`)
	b.WriteString(boolStr(j.OrdinaryExp))
	b.WriteString(`,"factual_problems":[`)
	for i, p := range probs {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"quote":"` + escapeJSON(p.Quote) + `","reason":"` + escapeJSON(p.Reason) + `"}`)
	}
	// The scores must come from j, not be hardcoded. They were literals here in
	// the first draft, so TestAdvisoryScoresDoNotAffectTheDecision fed the gate
	// 5/5 while believing it had fed it 0/0 — the assertion that the scores are
	// RECORDED caught it. A stub that silently substitutes its own values makes
	// every test built on it vacuous.
	b.WriteString(`],"interesting":` + itoa(j.Interesting) +
		`,"current":` + itoa(j.Current) +
		`,"note":"` + escapeJSON(j.Note) + `"}`)
	return b.String(), nil
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}

// goodJudgement is what the model should say about every entry in the real corpus.
var goodJudgement = judgement{Safe: true, TwoSided: true, Contestable: true, OrdinaryExp: true, Interesting: 7, Current: 6}

// ---------------------------------------------------------------------------
// THE CALIBRATION TESTS
// ---------------------------------------------------------------------------

// The first half of §10.6's bar. If this fails, the gate would have rejected
// provocations the owner actually published — the false-positive direction, which
// is the one that silently starves the site.
func TestGateAcceptsTheRealProvocations(t *testing.T) {
	if len(realProvocations) != 9 {
		t.Fatalf("the corpus is 9 provocations, got %d — calibration is against the real set, not a sample", len(realProvocations))
	}

	emptyBodied := 0
	for _, c := range realProvocations {
		c := c
		t.Run(c.Slug, func(t *testing.T) {
			v := gateCandidate(context.Background(), c, judgeReturning(goodJudgement), "stub")

			// THE ONE HONEST EXCEPTION, AND IT IS A FINDING, NOT AN EXEMPTION.
			// `group-chats-replaced-friendship` has a body of zero characters in the
			// production pool. There is nothing for the gate to judge, so it MUST be
			// rejected — and it must be rejected for exactly that reason, not for
			// some other rule that happens to catch it. Lowering minBodyLen to admit
			// an empty provocation would be fixing the checker to agree with a
			// broken row; the row is what needs fixing, and the owner's ruling that
			// provocations are written by the framework rather than by hand means
			// that repair belongs to the generator, not to me.
			if strings.TrimSpace(c.Body) == "" {
				emptyBodied++
				if v.Approved {
					t.Fatalf("APPROVED a provocation with an EMPTY body — there is no prose to " +
						"judge, so safety and factual checks both returned 'no problems found in " +
						"no text', which is not a pass")
				}
				if !hasRule(v, "body_too_short") {
					t.Errorf("rejected the empty-bodied row, but not as body_too_short; got: %s", ruleList(v))
				}
				t.Logf("POOL DEFECT (expected): %s has no body in production, so it cannot be judged", c.Slug)
				return
			}

			if !v.Approved {
				for _, r := range v.Reasons {
					if r.Fatal {
						t.Errorf("REJECTED a real provocation: [%s/%s] %s", r.Layer, r.Rule, r.Detail)
					}
				}
				t.Fatalf("%s was rejected; the corpus IS the specification", c.Slug)
			}
		})
	}

	// Guard the guard: if the pool is repaired and this row gains a body, the
	// exception above becomes dead code that would silently start exempting
	// nothing — or, worse, would exempt a DIFFERENT row that later loses its body.
	if emptyBodied != 1 {
		t.Errorf("expected exactly 1 empty-bodied row in the live corpus, found %d — "+
			"either the pool was repaired (delete the exception) or another row lost its "+
			"body (a new defect worth reporting)", emptyBodied)
	}
}

// The second half of §10.6's bar.
func TestGateRejectsTheDeliberatelyBadSet(t *testing.T) {
	for _, tc := range badProvocations {
		t.Run(tc.name, func(t *testing.T) {
			judge := judgeFn(nil)
			if tc.judgeSays != nil {
				judge = judgeReturning(*tc.judgeSays)
			} else {
				// Must be rejected before the judge is consulted at all.
				judge = func(context.Context, string) (string, error) {
					t.Fatalf("%s reached the judge; a deterministic layer should have rejected it first", tc.name)
					return "", nil
				}
			}
			v := gateCandidate(context.Background(), tc.c, judge, "stub")
			if v.Approved {
				t.Fatalf("APPROVED %s — this is the false-negative direction, which puts a false statement on a live homepage", tc.name)
			}
			if tc.wantRule != "" && !hasRule(v, tc.wantRule) {
				t.Errorf("%s was rejected, but not for %q; reasons: %s", tc.name, tc.wantRule, ruleList(v))
			}
		})
	}
}

func hasRule(v gateVerdict, rule string) bool {
	for _, r := range v.Reasons {
		if r.Rule == rule && r.Fatal {
			return true
		}
	}
	return false
}

func ruleList(v gateVerdict) string {
	var out []string
	for _, r := range v.Reasons {
		if r.Fatal {
			out = append(out, r.Layer+"/"+r.Rule)
		}
	}
	if len(out) == 0 {
		return "(none fatal)"
	}
	return strings.Join(out, ", ")
}

// ---------------------------------------------------------------------------
// §10.2 — the paths that will never be exercised by accident
// ---------------------------------------------------------------------------

// The single most important test in this file. A gate that cannot run must not
// be read as a gate with no objection. Each sub-case is a way the judge fails to
// produce a verdict, and every one must come out REJECTED.
func TestGateRejectsWhenTheJudgeNeverRan(t *testing.T) {
	good := aGoodCandidate(t)

	cases := []struct {
		name  string
		judge judgeFn
		rule  string
	}{
		{"judge is nil (no ai_service configured)", nil, "judge_unavailable"},
		{"judge returns an error", func(context.Context, string) (string, error) {
			return "", errors.New("connection reset")
		}, "judge_error"},
		{"judge times out", func(ctx context.Context, _ string) (string, error) {
			return "", context.DeadlineExceeded
		}, "judge_error"},
		{"judge returns an empty reply", func(context.Context, string) (string, error) {
			return "", nil
		}, "judge_unparseable"},
		{"judge returns prose instead of JSON", func(context.Context, string) (string, error) {
			return "Sure! This looks like a fine provocation to me.", nil
		}, "judge_unparseable"},
		{"judge reply is TRUNCATED mid-object", func(context.Context, string) (string, error) {
			// output_tokens == max_tokens means the completion was CUT, not
			// finished. A lenient parser would recover {"safe":true} and approve.
			return `{"safe":true,"two_sided":true,"arguable_from_ordinary_experience":tr`, nil
		}, "judge_unparseable"},
		{"judge reply renames a field", func(context.Context, string) (string, error) {
			return `{"is_safe":true,"two_sided":true,"contestable":true,"arguable_from_ordinary_experience":true,` +
				`"factual_problems":[],"interesting":5,"current":5,"note":""}`, nil
		}, "judge_unparseable"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := gateCandidate(context.Background(), good, tc.judge, "stub")
			if v.Approved {
				t.Fatalf("APPROVED on %q — absence of a verdict was read as a favourable verdict, "+
					"which is the bug this platform shipped on 2026-07-29 (chassis v1.0.1196)", tc.name)
			}
			if !hasRule(v, tc.rule) {
				t.Errorf("rejected, but not for %q; got: %s", tc.rule, ruleList(v))
			}
			if v.JudgeRan {
				t.Errorf("JudgeRan is true although the judge produced no usable verdict")
			}
		})
	}
}

// THE COUNCIL'S llm_reliability OBJECTION, TURNED INTO A TEST (2026-08-05).
//
// The seat's concern was precise and worth taking seriously: strict JSON decoding
// is a proxy for truncation, not the signal itself, so "a max_tokens cutoff that
// happens to land after a syntactically-complete JSON object would pass Decode()
// and be treated as a genuine verdict".
//
// INVESTIGATED, AND THE GATE IS ALREADY PROTECTED — but by the AI client layer,
// not by the parser, which is why the seat could not see it (judgeFn is opaque in
// a submission). Every provider client in the estate turns a truncation into a
// NON-NIL error before the gate ever sees the text:
//
//	anthropic.go  stop_reason == "max_tokens"  -> &TruncatedError{...}
//	gemini.go     finishReason MAX_TOKENS      -> &TruncatedError{...}
//	ollama.go     done_reason == "length"      -> &TruncatedError{...}
//
// So `raw, err := judge(...)` yields err != nil and the gate rejects, regardless
// of whether the partial text parses. The case below drives EXACTLY the seat's
// scenario — a truncation whose partial is perfectly valid, complete, approving
// JSON — and requires a rejection.
//
// This test exists because the protection lives in someone else's file. A future
// "improvement" that reaches for aiservice.IsTruncated to salvage a parseable
// partial would reintroduce the gap silently, and this is what would stop it.
func TestGateRejectsATruncationWhosePartialIsValidJSON(t *testing.T) {
	good := aGoodCandidate(t)

	// The nastiest possible shape: the cut landed after a complete object, and
	// that object says "approve". Only the error distinguishes it from a real
	// verdict, so if the gate ignored the error it would approve here.
	completeApprovingJSON := `{"safe":true,"two_sided":true,"contestable":true,` +
		`"arguable_from_ordinary_experience":true,"factual_problems":[],` +
		`"interesting":9,"current":9,"note":"looks good"}`

	if _, err := parseJudgement(completeApprovingJSON); err != nil {
		t.Fatalf("test setup is wrong: the partial must be VALID for this to model "+
			"the objection, but it failed to parse: %v", err)
	}

	truncating := func(context.Context, string) (string, error) {
		return completeApprovingJSON, &aiservice.TruncatedError{
			Partial:      completeApprovingJSON,
			OutputTokens: 4096,
			Reason:       "stop_reason=max_tokens",
			Provider:     "anthropic",
		}
	}

	v := gateCandidate(context.Background(), good, truncating, "stub")
	if v.Approved {
		t.Fatal("APPROVED a truncated response whose partial happened to be valid, " +
			"approving JSON — this is precisely the llm_reliability objection, and it " +
			"means the gate is trusting the parser instead of the provider's stop signal")
	}
	if !hasRule(v, "judge_error") {
		t.Errorf("rejected, but not as a judge error; got: %s", ruleList(v))
	}
	if v.JudgeRan {
		t.Error("JudgeRan is true for a truncated call; no complete verdict was produced")
	}
}

// A zero verdict must be a rejection. This pins the property that makes every
// early return in gateCandidate safe, independent of any code path.
func TestZeroVerdictIsARejection(t *testing.T) {
	var v gateVerdict
	if v.Approved {
		t.Fatal("the zero value of gateVerdict is APPROVED; every error path in the gate now fails open")
	}
}

// MUTATION CONTROL. A test suite that only ever sees a correct implementation
// cannot tell you it would notice a broken one. This drives the exact defect
// §10.2 describes — treating a judge error as "no objection" — and requires the
// calibration to catch it.
func TestTheFailClosedTestsWouldCatchAFailOpenGate(t *testing.T) {
	good := aGoodCandidate(t)

	// Simulate the broken gate: judge errors, and the caller ignores it and
	// approves. If our assertion below did not fire, TestGateRejectsWhenTheJudge
	// NeverRan would be vacuous.
	brokenVerdict := gateVerdict{Approved: true} // what a fail-OPEN gate returns
	if !brokenVerdict.Approved {
		t.Fatal("mutation setup is wrong; this must model an approval")
	}

	realVerdict := gateCandidate(context.Background(), good,
		func(context.Context, string) (string, error) { return "", errors.New("boom") }, "stub")

	if realVerdict.Approved == brokenVerdict.Approved {
		t.Fatal("the real gate agrees with a fail-open gate on an errored judge — the calibration is vacuous")
	}
}

// ---------------------------------------------------------------------------
// The thesis/body split — the design's central asymmetry
// ---------------------------------------------------------------------------

// The claims rail must never see the title.
//
// THIS TEST WAS VACUOUS IN ITS FIRST FORM AND THE MUTATION CHECK CAUGHT IT.
// It used the real title "Privacy is already over" and asserted that no rejection
// quoted it. But no fleet-wide pattern matches that string, so the assertion held
// whether the title was scanned or not — it passed under the deliberate mutation
// `blocks := []string{c.Body, c.Teaser, c.Title}`. A test that passes in both
// directions measures nothing.
//
// The same run refuted a claim I had written into the source: that scanning the
// title "would make the gate reject all nine calibration samples". It would not —
// none of the nine trips a fleet-wide pattern, so the asymmetry currently holds
// BY LUCK rather than by enforcement. That is precisely why this test has to
// supply the luck-free case itself.
//
// So the title below is chosen to MATCH a real fleet-wide pattern
// (`(you|readers?) can rely on (this|it|us|these|our)`, verified 2026-08-05) while
// remaining a plausible rhetorical provocation. Correct code: no finding. Title
// scanned: a finding, and the test fails. It now discriminates.
// WIDENED 2026-08-05 after the council's `compliance` seat objected that one case
// was thin coverage for a load-bearing exemption, and named the class it watches:
// overclaimed-reliability phrasing ("fully verified", "authoritative"). It was
// right — the exemption is what stops the gate rejecting every good provocation,
// and it was defended by a single string.
func TestClaimsRailIsNotGivenTheThesis(t *testing.T) {
	trippingTitles := []string{
		// The reliability-overclaim class the compliance seat tracks.
		"You can rely on this: every expert forecast is worthless",
		"Every claim is verified, and that is exactly the problem",
		"Guaranteed accurate is a marketing phrase, not a standard",
	}

	for _, title := range trippingTitles {
		t.Run(title[:24], func(t *testing.T) {
			// Control: prove this title really does trip the rail when scanned, so a
			// future change to the fleet-wide set cannot quietly make the test
			// vacuous again in the way the first version was.
			if got := datahelpers.ScanAllBannedClaims([]string{title}, nil); len(got) == 0 {
				t.Fatalf("this case no longer discriminates: %q trips no fleet-wide "+
					"pattern, so it would pass even if the thesis were scanned", title)
			}

			c := provocationCandidate{
				Title:  title,
				Teaser: "Forecasting is a confidence trade, not an information trade.",
				Body: "Ordinary supporting prose with nothing checkable in it at all, long enough to clear the floor. " +
					strings.Repeat("The counter-case is put here at length. ", 20),
			}
			var v gateVerdict
			checkClaims(c, &v)
			if len(v.Claims) != 0 || v.fatal() {
				t.Fatalf("the claims rail was given the thesis and rejected it: %s", ruleList(v))
			}
		})
	}
}

// The advisory scores are recorded but must never block. A model that scores a
// perfectly good provocation 0/10 for "interesting" must not be able to veto it —
// §10.7, because that judgement has no data behind it.
func TestAdvisoryScoresDoNotAffectTheDecision(t *testing.T) {
	c := aGoodCandidate(t)
	dull := goodJudgement
	dull.Interesting = 0
	dull.Current = 0

	v := gateCandidate(context.Background(), c, judgeReturning(dull), "stub")
	if !v.Approved {
		t.Fatalf("a zero advisory score blocked approval; §10.7 says those scores are recorded, not decisive (%s)", ruleList(v))
	}
	if v.Advisory.Interesting != 0 || v.Advisory.Current != 0 {
		t.Errorf("advisory scores were not recorded: %+v", v.Advisory)
	}
}
