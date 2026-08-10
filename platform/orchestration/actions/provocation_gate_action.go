// FILE: platform/orchestration/actions/provocation_gate_action.go
//
// The PROVOCATION GATE. Judges a candidate provocation and decides, with no
// human in the loop, whether it may enter the pool as publishable.
//
// WHY IT EXISTS, AND WHY IT IS THE ONLY CONTROL
// The owner decided on 2026-07-31 that generated provocations publish WITHOUT
// human approval (PLAN_2026-07-31_provocation_pipeline.md §10; recorded as taken,
// not pending — do not re-open it). That single decision changes what this file
// is. With an approver, a gate miss costs a moment of someone's attention.
// Without one, a gate miss is a false statement on a live homepage, which is
// `bugs_open/149` C1, witnessed on another site on 2026-07-29.
//
// THE SPLIT THAT IS THE WHOLE DESIGN (PLAN §4 Phase 2)
// A provocation is a deliberately contestable assertion. The claims rail
// ("nothing claims a number that is not true by construction") would reject every
// good provocation we have. But the BODY smuggles in ordinary factual claims that
// ARE fully subject to the rail — the four-day-week entry asserts the pilots
// "measure self-reported output", which is either true or false and has nothing
// to do with the thesis.
//
//	=> THE THESIS IS EXEMPT BY DESIGN. EVERY SUPPORTING FACTUAL ASSERTION IS NOT.
//
// A single blanket "is this true?" rejects everything; a single blanket "it is
// opinion" lets falsehoods through. Both failures have shipped here before
// (`bugs_closed/043`, generated copy inventing quantitative claims).
//
// HOW §10.2 IS ENFORCED — STRUCTURALLY, NOT BY A GUARD SOMEBODY MUST REMEMBER
// The failure to design against is not the gate judging wrongly; it is the gate
// NOT RUNNING (timeout, API error, malformed response) and the caller reading "no
// objection returned" as "no objection exists". This platform has already shipped
// that exact bug: a `!= nil` guard turned *unknown* into *no rule*, so an
// unpublished product range scored `Match` on a live page (fixed 2026-07-29,
// chassis v1.0.1196).
//
// So `gateVerdict.Approved` is a bool whose ZERO VALUE IS REJECTION, and approval
// is only ever written at ONE place in this file, after every layer has returned
// an affirmative. Every early return — including every error path — yields a
// rejected verdict by construction. There is no code path that can produce an
// approval by omission, which is a stronger property than remembering to default
// a variable. `TestGateRejectsWhenTheJudgeNeverRan` asserts it on the timeout
// path specifically, because that is the path that will never be exercised by
// accident.
//
// DEFENCE IN DEPTH, DELIBERATELY
// Layers A and B are deterministic and cannot fail open; layer C is a model and
// can. A candidate must clear ALL THREE. That ordering also means the cheap
// deterministic layers reject most bad candidates before any token is spent.
//
//	A  form      — corpus-derived shape (deterministic)
//	B  claims    — fleet-wide banned-claim scan of the BODY ONLY (deterministic)
//	C  judgement — safety + smuggled factual assertions (model, fails closed)
//
// WHAT IS DELIBERATELY NOT DECISIVE (PLAN §10.7)
// Criteria (b) "interesting" and (c) "current" are unmeasurable until there are
// contestants. A model asked to score them emits a confident number nothing can
// check — the precise shape of every entry in `WRONG_CALLS.md`. They are
// therefore RECORDED in the verdict and excluded from the publish decision. When
// paired mode produces engagement data they can become decisive; until then,
// pretending to measure them would be the least honest part of this file.
//
// Actions:
//   - gate_provocation: judge candidates, persist every verdict, approve none by default.

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Corpus-derived form rules
//
// PLAN §4: "Derive the criteria from the corpus we already have, not from first
// principles. That corpus IS the specification." The nine approved provocations
// were measured on 2026-08-05 rather than estimated:
//
//	title length   23..59 chars  (shortest "Privacy is already over",
//	                              longest  "Nobody actually reads terms of
//	                              service — and that's rational")
//	teaser length  38..78 chars
//
// The accepted bands below are deliberately WIDER than the observed range. A
// bound fitted exactly to nine samples is not a rule, it is a memorisation of
// the sample, and the first good provocation outside it would be rejected for
// being new rather than for being bad. The bands exist to catch the obviously
// malformed — a one-word title, an essay pasted into the title field.
// ---------------------------------------------------------------------------

const (
	minTitleLen  = 15
	maxTitleLen  = 90
	minTeaserLen = 20
	maxTeaserLen = 160
	// RE-JUSTIFIED 2026-08-06, because the original reason has been retired.
	//
	// This floor used to read "a body shorter than this cannot have made a case AND
	// its counter-case". The owner's ruling that one-sided provocations are
	// preferred removes that argument entirely, and a floor whose stated reason has
	// been withdrawn is a number nobody can defend — exactly the kind of rule that
	// survives by inertia and rejects something good later.
	//
	// The floor stays, on a narrower and measurable ground: THERE MUST BE SOMETHING
	// TO JUDGE. Layer C asks a model whether prose is safe and whether it smuggles
	// in false factual claims; neither question has an answer for an empty body, and
	// "no problems found in no text" is not a pass.
	//
	// Sized against the live corpus rather than chosen: the eight non-empty bodies
	// run 326..607 characters (measured 2026-08-05), so 200 sits clear of the real
	// distribution and catches only genuinely absent content. The ninth,
	// `group-chats-replaced-friendship`, has a body of ZERO characters in the
	// production pool — it fails this floor, and that is a defect in the POOL, not
	// in the gate. It is recorded as such rather than papered over by lowering the
	// number to admit it.
	minBodyLen = 200
)

// Hedging defeats the form: the corpus states its claims flatly, as fact. "AI
// will never be funny on purpose", not "AI may struggle to be funny".
var hedgePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(might|maybe|perhaps|possibly|arguably|probably)\b`),
	regexp.MustCompile(`(?i)\b(could be|may be|tends to|seems to|appears to)\b`),
	regexp.MustCompile(`(?i)\b(some (say|argue|think)|many believe|it depends)\b`),
}

// Tribal-political exclusion.
//
// PLAN §4: none of the nine touch party politics or the standard culture-war
// set, and this "should be made an explicit rule rather than left as an accident
// of taste". A keyword list is crude and cannot be the whole control — layer C
// judges this too — but it is deterministic, it cannot fail open, and it makes
// the rule reviewable instead of implicit in a prompt.
var tribalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(labour|tory|tories|conservative party|lib ?dems?|reform uk|snp)\b`),
	regexp.MustCompile(`(?i)\b(republican|democrat|maga|left-?wing|right-?wing|woke|anti-?woke)\b`),
	regexp.MustCompile(`(?i)\b(abortion|immigration|trans (rights|people)|gun control|brexit)\b`),
	regexp.MustCompile(`(?i)\b(israel|palestine|gaza|ukraine war|putin|trump|biden|starmer|farage)\b`),
}

// Abuse exclusion — DETERMINISTIC AND PRE-JUDGE (owner ruling 2026-08-09, PLAN §12).
//
// WHY THIS EXISTS, because a keyword list is otherwise the obvious thing to
// object to. The nine-round calibration on v1.0.1267 found the safety verdict
// resting on ONE stochastic boolean: `applyJudgement` rejects only on `!j.Safe`,
// and on round 3 of 9 the judge described a candidate as "pure repeated insult
// with no actual argument" in its own note and returned `safe:true` anyway. The
// note is free text and nothing cross-checks it against the boolean.
// (`HANDOFF_2026-08-08b_continue_here.md` §4.)
//
// A pattern list is crude and cannot be the whole control — layer C still judges
// this. What this layer buys is the one thing layer C cannot: anything it kills
// NEVER REACHES the judge's discretion, so no amount of stochasticity can approve
// it. That is the whole point of the ruling — remove the randomness rather than
// sample it.
//
// A SECOND, COMPOSING RULING EXISTS AND IS NOT IMPLEMENTED HERE. The owner
// separately told a concurrent session "we can ask a model different ways,
// several times" — varied-framing sampling of the judge (PLAN §14.2). The two
// rulings compose and neither supersedes the other: this layer is the
// deterministic floor, that one reduces what survives above it. It is deliberately
// NOT referenced by name in this comment, because the function does not exist yet
// and a comment citing an unwritten symbol reads exactly like a comment citing a
// real one.
//
// IT ERRS TOWARD REJECTION, BY RULING, AND THAT HAS A KNOWN COST.
// Owner ruling 2026-08-09 (PLAN §13, ruling 1): a false negative publishes abuse
// under his name; a false positive silently starves the pool. He chose the pool.
// So:
//
//   - This list is NEGATION-BLIND, deliberately. A provocation ARGUING ABOUT
//     online abuse ("calling strangers idiots is the price of the internet") will
//     be rejected by its own subject matter. That is the same shape as
//     `bugs_open/222` — a proximity detector convicting the denial — and it is
//     accepted here rather than fixed, because the alternative is a negation
//     parser deciding whether an insult is being used or mentioned, which is
//     exactly the model judgement this layer exists to not depend on.
//   - The recorded rejection therefore names the matched text, so a human reading
//     the verdict can see instantly that it was a mention rather than a use, and
//     release it. The reversed §10 (a human CAN approve — PLAN §14) is what makes
//     that cheap; before that reversal this list would have been too blunt.
//
// SCOPE: pejoratives aimed at PEOPLE, and incitement. Not profanity — "this is
// shit" is rude, not abusive, and the corpus is allowed to be rude.
var abusivePatterns = []*regexp.Regexp{
	// Pejoratives aimed at a person or group. The corpus argues with positions;
	// it does not name-call, and a provocation never NEEDS one of these.
	regexp.MustCompile(`(?i)\b(idiots?|morons?|imbeciles?|cretins?|half-?wits?|dim-?wits?|numpties|numpty)\b`),
	regexp.MustCompile(`(?i)\b(brain-?dead|mouth-?breathers?|knuckle-?draggers?|sub-?human|vermin|scum)\b`),
	regexp.MustCompile(`(?i)\b(stupid|thick|pathetic|worthless|deluded|braindead)\s+(people|men|women|users?|fans?|lot|bunch)\b`),
	regexp.MustCompile(`(?i)\banyone who\b[^.!?]{0,60}\bis (an? )?(idiot|moron|fool|imbecile|cretin)\b`),
	// Incitement and dehumanisation. Fatal irrespective of target.
	regexp.MustCompile(`(?i)\b(should be (shot|hanged|killed|beaten)|deserve to (die|suffer)|hope (they|he|she) dies?)\b`),
	regexp.MustCompile(`(?i)\b(kill yourself|kys)\b`),
}

// ---------------------------------------------------------------------------
// The verdict
// ---------------------------------------------------------------------------

// gateReason is one recorded ground for the decision. Rejections carry these and
// so do approvals — an approval with no recorded reasoning is indistinguishable
// from a gate that did not run (§10.3).
type gateReason struct {
	Layer  string `json:"layer"`  // "form" | "claims" | "judgement"
	Rule   string `json:"rule"`   // machine-readable rule name
	Detail string `json:"detail"` // human-readable, quotes the offending text
	Fatal  bool   `json:"fatal"`  // did this reason alone block approval?
}

// advisory holds the judgements deliberately EXCLUDED from the decision (§10.7).
// Recorded so that when contestant data exists we can ask, retrospectively,
// whether these scores ever predicted anything.
type advisory struct {
	Interesting int    `json:"interesting"` // 0..10, NOT decisive
	Current     int    `json:"current"`     // 0..10, NOT decisive
	Note        string `json:"note"`
}

// gateVerdict is the persisted decision.
//
// THE ZERO VALUE IS A REJECTION. That is load-bearing, not incidental: every
// error path in this file returns a zero-ish verdict and is therefore a
// rejection without having to remember to say so.
type gateVerdict struct {
	Approved bool                       `json:"approved"`
	Reasons  []gateReason               `json:"reasons"`
	Claims   []datahelpers.ClaimFinding `json:"claims,omitempty"`
	Advisory advisory                   `json:"advisory"`
	JudgeRan bool                       `json:"judge_ran"`
	Model    string                     `json:"model,omitempty"`
	GatedAt  string                     `json:"gated_at"`
	GateVer  string                     `json:"gate_version"`
}

// gateVersion is stamped into every verdict so a calibration run can be tied to
// the rules that produced it. Bump it whenever a rule changes, or the corpus
// evidence for an old verdict silently stops meaning what it says.
//
// > **I broke this rule on 2026-08-06 and it cost real ambiguity two days later.**
// > The owner's rulings retired the two-sidedness rejection and added the
// > contestability one — a change to what the gate DECIDES — and I left the
// > version at "1". On 08-08 a calibration run appeared in the pool that I had not
// > dispatched, and `gate_version` could not tell me whether it came from the old
// > rules or the new ones. It had to be established indirectly, by grepping the
// > verdicts for a `one_sided` note that only the new code can emit. That worked,
// > but it is exactly the reconstruction this field exists to make unnecessary.
// >
// > **The lesson is about WHERE the bump belongs, not about remembering harder:**
// > it is not a release chore, it is part of the edit that changes a rule. If you
// > are editing `applyJudgement`, `checkForm` or the fatal/advisory split, this
// > line is in the same change or the change is not finished.
//
// 2 — one-sided provocations accepted (advisory `one_sided` note); contestability
// added as a fatal criterion; minBodyLen re-justified. Owner rulings 2026-08-06.
const gateVersion = "2"

func (v *gateVerdict) reject(layer, rule, detail string) {
	v.Approved = false
	v.Reasons = append(v.Reasons, gateReason{Layer: layer, Rule: rule, Detail: detail, Fatal: true})
}

func (v *gateVerdict) note(layer, rule, detail string) {
	v.Reasons = append(v.Reasons, gateReason{Layer: layer, Rule: rule, Detail: detail, Fatal: false})
}

func (v *gateVerdict) fatal() bool {
	for _, r := range v.Reasons {
		if r.Fatal {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// The candidate
// ---------------------------------------------------------------------------

type provocationCandidate struct {
	ID     string
	Slug   string
	Title  string // the THESIS — exempt from the truth check by design
	Teaser string
	Body   string // supporting prose — fully subject to the claims rail
}

// ---------------------------------------------------------------------------
// Layer A — form
// ---------------------------------------------------------------------------

// checkForm applies the corpus-derived shape rules. Deterministic: same input,
// same verdict, no network, no tokens. Everything it can decide is decided here
// so that a malformed candidate never reaches a paid call.
func checkForm(c provocationCandidate, v *gateVerdict) {
	title := strings.TrimSpace(c.Title)
	teaser := strings.TrimSpace(c.Teaser)
	body := strings.TrimSpace(c.Body)

	if n := len([]rune(title)); n < minTitleLen || n > maxTitleLen {
		v.reject("form", "title_length",
			fmt.Sprintf("title is %d chars, outside the accepted %d..%d", n, minTitleLen, maxTitleLen))
	}
	if n := len([]rune(teaser)); n < minTeaserLen || n > maxTeaserLen {
		v.reject("form", "teaser_length",
			fmt.Sprintf("teaser is %d chars, outside the accepted %d..%d", n, minTeaserLen, maxTeaserLen))
	}
	if n := len([]rune(body)); n < minBodyLen {
		v.reject("form", "body_too_short",
			fmt.Sprintf("body is %d chars; under %d it cannot have made a case AND its counter-case", n, minBodyLen))
	}

	// A question is not a provocation. The corpus asserts; it does not ask.
	if strings.HasSuffix(title, "?") {
		v.reject("form", "title_is_a_question",
			"a provocation states a contestable claim flatly; it does not ask one")
	}

	for _, re := range hedgePatterns {
		if m := re.FindString(title); m != "" {
			v.reject("form", "title_hedges",
				fmt.Sprintf("title hedges with %q; the corpus states its claims as fact", m))
			break
		}
	}

	// Tribal-political check runs over the WHOLE artefact, not just the title:
	// a neutral-sounding thesis with a party-political body is the shape this
	// rule exists to stop.
	whole := title + " " + teaser + " " + body
	for _, re := range tribalPatterns {
		if m := re.FindString(whole); m != "" {
			v.reject("form", "tribal_political",
				fmt.Sprintf("mentions %q; the corpus deliberately avoids party politics and the culture-war set", m))
			break
		}
	}

	// Abuse runs over the WHOLE artefact for the same reason tribal_political
	// does, and it is FATAL on first match. Layer, deliberately, is "form" and not
	// a new "safety" value: the layer vocabulary is read by the calibration scorer
	// and stored in every historical verdict, and widening it would be a shared-
	// vocabulary change for a rule the `rule` field already names precisely.
	//
	// The matched text is quoted into the detail because this rule is negation-
	// blind by design (see `abusivePatterns`): the quote is what lets a human tell
	// a USE from a MENTION in one glance and release a false positive.
	for _, re := range abusivePatterns {
		if m := re.FindString(whole); m != "" {
			v.reject("form", "abusive_language",
				fmt.Sprintf("contains %q; the corpus argues with positions, never with people "+
					"(owner ruling 2026-08-09 — this check errs toward rejection, so a provocation "+
					"that merely MENTIONS abuse is caught too and needs a human release)", m))
			break
		}
	}
}

// ---------------------------------------------------------------------------
// Layer B — the claims rail, on the BODY ONLY
// ---------------------------------------------------------------------------

// checkClaims scans the supporting prose against the fleet-wide banned-claim set
// — known falsehoods, each one a pattern some site already shipped and was
// corrected for.
//
// IT IS GIVEN THE BODY AND THE TEASER, NEVER THE TITLE. That asymmetry is the
// design: the title is the thesis and is exempt by construction.
//
// > **CORRECTED 2026-08-05, by running the mutation rather than reasoning about
// > it.** This comment first claimed that scanning the title "would make the gate
// > reject all nine calibration samples". It would NOT. Applying the mutation
// > `blocks := []string{c.Body, c.Teaser, c.Title}` leaves every calibration test
// > passing, because none of the nine titles happens to match a fleet-wide
// > pattern. **So the asymmetry currently holds by luck, not by enforcement** —
// > and the test that was supposed to guard it passed under the mutation too.
// > `TestClaimsRailIsNotGivenTheThesis` now supplies its own tripping title (one
// > verified to match a live pattern) so that it fails when the thesis is
// > scanned. Do not restore the old claim: the nine are not a control for this.
//
// eb is nil deliberately: a provocation belongs to a domain, not to a site
// register, so only the fleet-wide set applies. ScanAllBannedClaims documents
// nil as "this site has no register", not "do not scan".
func checkClaims(c provocationCandidate, v *gateVerdict) {
	blocks := []string{c.Body, c.Teaser}
	findings := datahelpers.ScanAllBannedClaims(blocks, nil)
	if len(findings) == 0 {
		return
	}
	v.Claims = findings
	for _, f := range findings {
		v.reject("claims", "banned_claim",
			fmt.Sprintf("%s: %q (%s)", f.Check, f.Matched, f.Reason))
	}
}

// ---------------------------------------------------------------------------
// Layer C — the judge
// ---------------------------------------------------------------------------

// judgement is the model's structured reply. Every decisive field defaults to
// the unsafe answer being FALSE, so a partially-parsed reply cannot approve.
type judgement struct {
	Safe            bool `json:"safe"`
	TwoSided        bool `json:"two_sided"`
	Contestable     bool `json:"contestable"`
	OrdinaryExp     bool `json:"arguable_from_ordinary_experience"`
	FactualProblems []struct {
		Quote  string `json:"quote"`
		Reason string `json:"reason"`
	} `json:"factual_problems"`
	Interesting int    `json:"interesting"`
	Current     int    `json:"current"`
	Note        string `json:"note"`
}

// buildJudgePrompt asks the model to judge a candidate.
//
// THE FACTUAL CRITERION WAS NARROWED TO *FABRICATION* ON 2026-08-08, and the
// evidence is the second live calibration rather than an argument.
//
// The owner's instruction after the first run was "can we just remove that
// rhetoric?", and the flagged clause was duly deleted from the pool. The re-run
// shows why that cannot work phrase by phrase: the objection simply MOVED to the
// adjacent clause of the same sentence, and a second provocation was newly
// rejected for "Reading takes an hour" (a figure of speech) and "every study that
// frames it as apathy" (a rhetorical generalisation). Three flags across two
// provocations, none of them a fabrication.
//
// So the check was not catching what PLAN §4 and `bugs_closed/043` aimed at —
// generated copy INVENTING quantitative claims — it was penalising the ordinary
// register of argumentative prose. A provocation that must cite every
// generalisation is not writable, and the failure mode is silent: the pool starves
// while the gate reports itself working.
//
// This is a CORRECTION toward the stated intent, not a relaxation. Fabrication —
// an invented study, a made-up statistic, a false checkable fact — is still fatal,
// and that is the half `bugs_closed/043` was about. What is no longer fatal is
// "unsupported", which was never the target.
//
// ALSO MEASURED, and it changes what "calibrated" can mean: the judge is
// STOCHASTIC. `nobody-reads-terms-of-service` drew no factual objection on 05 Aug
// and two on 08 Aug from byte-identical text. A single green run is therefore not
// evidence; the calibration has to pass repeatedly before it means anything.
func buildJudgePrompt(c provocationCandidate) string {
	var b strings.Builder
	b.WriteString(`You are the safety and form gate for a daily "provocation" — a deliberately contestable opinion published for people to argue with.

THE MOST IMPORTANT RULE, AND THE ONE MOST EASILY GOT WRONG:
The THESIS is an opinion and is EXEMPT from fact-checking. You must NOT reject a
provocation for being contestable, one-sided in its thesis, overstated, or
something you disagree with. That is what a provocation IS.

But the SUPPORTING PROSE often smuggles in ordinary factual claims, and those are
NOT exempt. Example of the distinction, from a real accepted entry:
  thesis  "The four-day week is a productivity myth"   <- opinion, exempt
  body    "the pilots measure self-reported output"    <- a factual claim, checkable

Judge ONLY the supporting prose for factual problems, and report a factual problem
ONLY for FABRICATION:
  - a named study, institution, report or person that does not exist;
  - a statistic, percentage, sample size or date presented as sourced when it is
    invented;
  - a specific, checkable statement of fact that is simply false.

DO NOT report any of the following. This prose is an argument, not a paper, and
these are the register it is written in:
  - rhetorical or idiomatic quantities ("Reading takes an hour", "it takes a
    lawyer") - these are figures of speech, not measurements;
  - sweeping generalisations about a category ("the pilots recruit believers",
    "every study that frames it as apathy") - these are the ARGUMENT, contestable
    by design, and the reader is invited to disagree;
  - anything merely UNSUPPORTED, UNCITED, UNVERIFIED or "overstated". An opinion
    piece is allowed to assert without citing. Absence of a citation is not
    evidence of falsehood.

The test is INVENTED, not UNCITED. If you would write "unsupported",
"unverified", "no source given" or "overgeneralised", that is NOT a factual
problem - say nothing. Report it only if you believe the thing named does not
exist or the claim is actually false.

Do NOT report the thesis at all.

Also judge:
  safe      - no slur, harassment, incitement, medical/legal/financial advice,
              targeting of a private individual, or party-political tribalism.
  two_sided - the prose makes the case AND genuinely puts the counter-case.
              RECORDED ONLY on this site; a one-sided piece is acceptable here.
  contestable - a reasonable, informed person could take the OPPOSITE view and
              argue it seriously. This is the test that separates a real
              provocation from filler. "Privacy is already over" is contestable.
              "AI is changing everything" is NOT: nobody disagrees, so there is
              nothing to argue. Vague trend statements, truisms and things
              everyone already accepts all fail this.
  arguable_from_ordinary_experience - an ordinary reader could disagree from
              their own life, without specialist knowledge.

Finally, score two things we CANNOT yet measure and which do NOT affect the
decision — answer honestly, they are recorded for later calibration:
  interesting 0-10, current 0-10.

Reply with ONLY a JSON object, no prose, no code fence:
{"safe":bool,"two_sided":bool,"contestable":bool,"arguable_from_ordinary_experience":bool,
 "factual_problems":[{"quote":"...","reason":"..."}],
 "interesting":0-10,"current":0-10,"note":"one sentence"}

--- CANDIDATE ---
`)
	b.WriteString("THESIS (exempt from fact-checking): ")
	b.WriteString(c.Title)
	b.WriteString("\nTEASER: ")
	b.WriteString(c.Teaser)
	b.WriteString("\nSUPPORTING PROSE (fact-checkable):\n")
	b.WriteString(c.Body)
	return b.String()
}

// parseJudgement decodes the model's reply.
//
// STRICT ON PURPOSE. Any failure here — empty reply, prose around the JSON,
// truncation, a renamed field — returns an error, and every caller treats an
// error as a rejection. A lenient parser that "recovers" a partial object is how
// a truncated reply becomes an approval, and `output_tokens == max_tokens` means
// the completion was CUT, not finished (CLAUDE.md).
func parseJudgement(raw string) (*judgement, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, fmt.Errorf("judge returned an empty reply")
	}
	// Tolerate a fenced block, which several providers add despite instruction;
	// tolerate nothing else.
	if i := strings.Index(s, "{"); i > 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 && j < len(s)-1 {
		s = s[:j+1]
	}
	var j judgement
	dec := json.NewDecoder(strings.NewReader(s))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&j); err != nil {
		return nil, fmt.Errorf("judge reply is not the agreed shape: %w", err)
	}
	return &j, nil
}

// applyJudgement folds a successfully-parsed judgement into the verdict.
func applyJudgement(j *judgement, v *gateVerdict) {
	v.JudgeRan = true
	v.Advisory = advisory{Interesting: j.Interesting, Current: j.Current, Note: j.Note}

	if !j.Safe {
		v.reject("judgement", "unsafe", "judge marked the candidate unsafe: "+j.Note)
	}
	// TWO-SIDEDNESS IS RECORDED, NOT ENFORCED — OWNER RULING 2026-08-06.
	//
	// This was a fatal rejection until the live calibration ran, and it was wrong
	// twice over.
	//
	// (1) Its stated justification — "every entry in the corpus does" — was FALSE.
	// PLAN §4 asserts the corpus is "genuinely two-sided … every detail_body makes
	// the case and then makes the counter-case", and that claim was evidently
	// written from the handful of entries that do. Measured across all nine on
	// 2026-08-05: five carry an explicit counter-case, four do not, and one has a
	// body of zero characters. I encoded a property of a SAMPLE as a rule about
	// the CORPUS, and the gate then rejected 5 of the owner's 9 published
	// provocations — a well-behaved gate enforcing a rule its own subject matter
	// does not follow.
	//
	// (2) The owner's answer to that finding was not "fix the rule" but a product
	// judgement: **"the one sided provocation is better, we can continue that way
	// for now at least."** So a one-sided provocation is not a defect to be
	// tolerated; on this site it is preferred.
	//
	// The verdict is still recorded, because "how many are one-sided" is exactly
	// the kind of thing worth being able to ask later, and because the ruling is
	// explicitly provisional ("for now at least"). It joins interesting/current in
	// the advisory block: observed, never decisive.
	if !j.TwoSided {
		v.note("judgement", "one_sided",
			"the prose does not put a counter-case — RECORDED ONLY, never fatal "+
				"(owner ruling 2026-08-06: one-sided provocations are preferred here)")
	}
	// CONTESTABILITY IS FATAL, AND IT IS WHAT NOW CATCHES TRENDING SLOP.
	//
	// Added 2026-08-06, forced by a real consequence of the owner's ruling rather
	// than chosen: demoting two-sidedness to advisory removed the ONLY criterion
	// that was rejecting §10.6's "trending slop" sample, and the calibration said
	// so immediately by approving "AI is changing everything". §10.6 still requires
	// that sample to be rejected, so the ruling did not license letting it through
	// — it created a gap that had to be filled somewhere else.
	//
	// Contestability is the right place, because it is precisely what distinguishes
	// the one-sided provocations the owner prefers from filler. "Privacy is already
	// over" takes a position a reasonable person can argue against; "AI is changing
	// everything" states something nobody disputes. The first is a provocation with
	// no counter-case; the second is not a provocation at all. All nine live
	// entries satisfy it, so it costs the corpus nothing.
	if !j.Contestable {
		v.reject("judgement", "not_contestable",
			"nobody could seriously argue the opposite — a provocation must take a "+
				"position worth disputing, not state a trend everyone accepts")
	}
	if !j.OrdinaryExp {
		v.reject("judgement", "needs_specialist_knowledge",
			"an ordinary reader could not disagree from their own experience")
	}
	for _, fp := range j.FactualProblems {
		v.reject("judgement", "factual_problem_in_body",
			fmt.Sprintf("%q — %s", fp.Quote, fp.Reason))
	}

	// Recorded, deliberately NOT decisive (§10.7).
	v.note("judgement", "advisory_scores",
		fmt.Sprintf("interesting=%d current=%d (recorded; excluded from the decision until contestant data exists)",
			j.Interesting, j.Current))
}

// ---------------------------------------------------------------------------
// The gate itself
// ---------------------------------------------------------------------------

// judgeFn is the seam the tests drive. Production passes a closure over the AI
// client; tests pass a stub that can return an error, a timeout or a truncated
// reply — the paths that decide whether §10.2 actually holds.
type judgeFn func(ctx context.Context, prompt string) (string, error)

// gateCandidate runs all three layers and returns a verdict.
//
// THERE IS EXACTLY ONE `Approved = true` IN THIS FUNCTION and it is the last
// statement before the return, guarded by every layer having spoken. Keep it
// that way: the property that makes this gate fail closed is that approval
// cannot be reached by any early exit.
func gateCandidate(ctx context.Context, c provocationCandidate, judge judgeFn, model string) gateVerdict {
	v := gateVerdict{
		GatedAt: time.Now().UTC().Format(time.RFC3339),
		GateVer: gateVersion,
		Model:   model,
	}

	checkForm(c, &v)
	checkClaims(c, &v)

	// Cheap layers already rejected it — do not spend a call to confirm.
	if v.fatal() {
		return v
	}

	if judge == nil {
		v.reject("judgement", "judge_unavailable",
			"no judge was configured; absence of a verdict is not a favourable verdict (§10.2)")
		return v
	}

	raw, err := judge(ctx, buildJudgePrompt(c))
	if err != nil {
		// §10.2, the whole point: the gate not running is a REJECTION.
		v.reject("judgement", "judge_error",
			fmt.Sprintf("judge did not return a verdict (%v); treated as a rejection, never as a pass", err))
		return v
	}
	j, perr := parseJudgement(raw)
	if perr != nil {
		v.reject("judgement", "judge_unparseable",
			fmt.Sprintf("%v; treated as a rejection", perr))
		return v
	}

	applyJudgement(j, &v)
	if v.fatal() {
		return v
	}

	v.Approved = true
	return v
}

// ---------------------------------------------------------------------------
// Persistence
// ---------------------------------------------------------------------------

// persistVerdict writes the decision back to the candidate row — ALWAYS,
// including rejections (§10.3).
//
// The rejections are the interesting half: "a gate that has rejected 100% or 0%
// for a week is broken in one of two directions and both are invisible without
// the log." A rejected row keeps status='draft' and carries its reasons, so the
// population is queryable after the fact.
func persistVerdict(ctx context.Context, db *sql.DB, id string, v gateVerdict) error {
	blob, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal verdict: %w", err)
	}

	// STRIP NUL BYTES BEFORE THE jsonb WRITE — raised by the council's
	// bug_historian seat (2026-08-05), citing `bugs_closed/056`: LLM-derived text
	// containing a NUL kills a jsonb UPDATE with 22P05. The verdict embeds the
	// judge's own quotes and reasons, so it is exactly the LLM-derived text that
	// case describes.
	//
	// The failure would not have been silent here — the caller propagates this
	// error and aborts the batch — but "aborts the whole batch on one poisoned
	// character" is a bad trade for a value that carries no meaning in JSON
	// anyway. Removing it keeps the rejection log complete, which is the whole
	// point of §10.3: a rejected-but-unlogged candidate is indistinguishable from
	// one that was never judged.
	clean := bytes.ReplaceAll(blob, []byte{0}, nil)
	// CORRECTED 2026-08-05, on reading the schema rather than assuming it. The
	// first version parked rejections back in 'draft'. They were inert (the
	// candidate query also requires gated_at IS NULL) but they were MISLABELLED,
	// and that defeats the requirement this function exists to satisfy: §10.3
	// wants the rejections observable, and `WHERE status='rejected'` — the query
	// anyone would actually write — would have returned zero for ever while the
	// gate was rejecting everything. The table's CHECK constraint offers exactly
	// four statuses and 'rejected' is one of them; using 'draft' for a judged
	// candidate was inventing a meaning the schema already had a word for.
	status := "rejected"
	if v.Approved {
		status = "approved"
	}
	_, err = db.ExecContext(ctx, `
		UPDATE provocations
		   SET gate_verdict = $2::jsonb,
		       gated_at     = now(),
		       status       = $3
		 WHERE id = $1::uuid`, id, string(clean), status)
	if err != nil {
		return fmt.Errorf("persist verdict for %s: %w", id, err)
	}
	return nil
}

// loadGateCandidates returns ungated draft rows for a domain, oldest first.
//
// Only 'draft' AND gated_at IS NULL is ever selected, which excludes both ends
// deliberately. An already-approved row is never re-gated: it may already be
// live, and re-judging it would let a model's drift silently retract a published
// provocation. A 'rejected' row is not re-gated either, so a rejection costs one
// call and not one per run for ever. Re-judging rejects after a rule change is a
// legitimate future need and stays possible — clear `gated_at` and set the row
// back to 'draft' — but it must be a deliberate act, not a side effect of the
// gate running on a schedule.
func loadGateCandidates(ctx context.Context, db *sql.DB, domain string, limit int) ([]provocationCandidate, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, slug, title, teaser,
		       COALESCE(NULLIF(body, ''), COALESCE(detail_body, ''))
		  FROM provocations
		 WHERE domain = $1
		   AND status = 'draft'
		   AND gated_at IS NULL
		 ORDER BY created_at ASC
		 LIMIT $2`, domain, limit)
	if err != nil {
		return nil, fmt.Errorf("query gate candidates: %w", err)
	}
	defer rows.Close()

	var out []provocationCandidate
	for rows.Next() {
		var c provocationCandidate
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Teaser, &c.Body); err != nil {
			return nil, fmt.Errorf("scan gate candidate: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GateProvocationAction is the registered entry point.
func GateProvocationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("GateProvocation: starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config
	if config == nil {
		config = make(map[string]interface{})
	}
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		for k, v := range inputData {
			config[k] = v
		}
	}

	domain, _ := config["domain"].(string)
	if strings.TrimSpace(domain) == "" {
		return nil, fmt.Errorf("gate_provocation requires an explicit domain")
	}
	limit := 20
	if n, ok := config["limit"].(float64); ok && n > 0 {
		limit = int(n)
	}

	candidates, err := loadGateCandidates(ctx, params.DB, domain, limit)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		params.Logger.Info("GateProvocation: no ungated drafts", zap.String("domain", domain))
		return map[string]interface{}{
			"status": "complete", "judged": 0, "approved": 0, "rejected": 0,
		}, nil
	}

	// Build the judge. A missing ai_service is NOT fatal to the action — it is
	// fatal to every candidate, which is the same outcome expressed as a
	// rejection with a recorded reason rather than as a silent skip.
	var judge judgeFn
	model := ""
	if aiCfg := getAIServiceConfig(params); aiCfg != nil {
		if m, ok := aiCfg["model"].(string); ok {
			model = m
		}
		if client, cerr := createAIClient(ctx, aiCfg); cerr == nil {
			judge = func(c context.Context, prompt string) (string, error) {
				return client.GenerateText(c, prompt, map[string]interface{}{})
			}
		} else {
			params.Logger.Warn("GateProvocation: AI client unavailable; every candidate will be rejected",
				zap.Error(cerr))
		}
	} else {
		params.Logger.Warn("GateProvocation: no ai_service configured; every candidate will be rejected")
	}

	approved, rejected := 0, 0
	for _, c := range candidates {
		v := gateCandidate(ctx, c, judge, model)
		if err := persistVerdict(ctx, params.DB, c.ID, v); err != nil {
			// A verdict we could not record is a verdict that did not happen.
			return nil, err
		}
		if v.Approved {
			approved++
			params.Logger.Info("GateProvocation: approved", zap.String("slug", c.Slug))
		} else {
			rejected++
			params.Logger.Info("GateProvocation: rejected",
				zap.String("slug", c.Slug), zap.Int("reasons", len(v.Reasons)))
		}
	}

	params.Logger.Info("GateProvocation: done",
		zap.String("domain", domain), zap.Int("approved", approved), zap.Int("rejected", rejected))
	return map[string]interface{}{
		"status": "complete", "judged": len(candidates),
		"approved": approved, "rejected": rejected,
	}, nil
}
