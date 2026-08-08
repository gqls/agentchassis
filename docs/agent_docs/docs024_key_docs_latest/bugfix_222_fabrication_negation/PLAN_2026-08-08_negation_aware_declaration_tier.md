# PLAN — bugfix 222: make the fabrication gate's declaration tier negation-aware

**Workstream:** `bugfix_222_fabrication_negation` · **Bug:** `bugs_open/222_HANDOFF_2026-08-08_fabrication_declaration_tier_convicts_the_denial.md` · **Status:** plan, drafted with Fable, implementing next.

**The defect in one sentence:** Tier A of `DetectToolFabrication` (`platform/orchestration/actions/check_tool_fabrication_action.go:91–93`) convicts on a qualifier-near-data-noun proximity match with no negation awareness, so the comment `// In-memory portfolio store (no fabricated data — starts empty)` — a *denial* of fabrication, on a tool that genuinely starts empty — was convicted as a declaration of it, discarding a correct, paid-for recreation into the human-review queue. The recreate prompt's own Data Integrity section makes such denials *likely* (the `prompt-text-poisons-its-own-detector` class), so this recurs.

---

## 1. Chosen design, and the alternatives rejected

**Chosen: extract the negation-guard ALGORITHM from `datahelpers/claims.go` as one small exported, parameterised primitive; keep the claims layer's cue vocabulary, window and boundary exactly as they are (zero behaviour change there); give the fabrication gate its own domain-tuned cue vocabulary in the actions package, applied at the position of the qualifier token in each Tier A match.**

The load-bearing observation: this platform already has a negation guard for exactly this class of false positive — `negatedClaimMatch` in `platform/orchestration/datahelpers/claims.go:593`, registered as **CLM-017**, shared by two consumers (`ScanBannedClaims` at claims.go:494, `ScanAttributedUncitedStats` at claims_attributed.go:192) under the CLM-004 anti-drift doctrine ("one matching rule here, not two that diverge"). What is shareable is the **algorithm** (bounded backwards window, clause-boundary trim, cue-regex test — including the hard-won multibyte-rune handling at claims.go:599–602). What is *not* shareable is the **cue vocabulary**: `negationCueRe` deliberately excludes bare `no` and `without` (claims.go:558–568, with counter-examples: "Without exception, every claim is verified" / "There are no exceptions: …" — an intensifier, not a negator, in marketing prose), and that exclusion is pinned by `TestBareNoIsAKnownResidualOfTheSharedGuard` (claims_attributed_test.go:123). The motivating bug-222 payload is negated by exactly the cue the claims layer excludes: bare "no".

Alternatives considered and rejected:

1. **Naive reuse of `negatedClaimMatch` as-is** (export it, call it from the gate). Rejected because it does not fix the bug: "no fabricated data" contains no cue `negationCueRe` recognises, so the motivating payload would still be convicted. Verified against the regex at claims.go:573–577 — bare `no` and `without` are absent by design.
2. **Widen `negationCueRe` globally to add `no`/`without`.** Rejected. The exclusion is load-bearing at the claims layer, which runs at **blocker** severity on live marketing copy (CLM-015/CLM-017: the strongest fleet-wide pattern was re-armed only *because* the guard is narrow). Adding bare `no` there disarms superlative overclaims built ON "no" and would trip `TestBareNoIsAKnownResidualOfTheSharedGuard`, a test written precisely to force this to be a deliberate claims-layer decision. Making that decision to fix an unrelated tool-recreation gate is the wrong trade, made in the wrong place.
3. **A one-off local negation regex inside the actions package** (the bug file's candidate 1, implemented naively as a private helper with its own window/trim logic). Rejected on the repo's own convention ("reuse existing machinery before building new") and on CLM-004's evidence: a second independently-maintained implementation of "is this token negated in its clause" is exactly the drift class the claims layer documents. It would also silently re-import two subtle defects the shared implementation already fixed: the multibyte clause-boundary rune bug, and the negation-inside-the-match-span bug (CLM-019 §"the third"), both found by tests, not by reading.
4. **The bug file's candidate 2** — scope the declaration scan to code/comment-stripped prose. Rejected as this fix (not as a future one): the declaration that motivated the *gate* (the vetcomp "we generate a large, realistic, deterministic dataset" confession) was itself a code comment, so stripping comments guts Tier A's primary catch surface.

The chosen design is the structural fix at framework level: one algorithm, two domain vocabularies, each with its own documented inclusions/exclusions.

## 2. Exact signatures, and how `negatedClaimMatch` is rewired

In `platform/orchestration/datahelpers/claims.go`, immediately after the CLM-017 doctrine block:

```go
// NegationGuard is the clause-local negation test, parameterised by domain.
// The ALGORITHM is shared; the CUE VOCABULARY is deliberately NOT shared —
// each domain tunes its own. Do not merge vocabularies in either direction.
type NegationGuard struct {
	Cue      *regexp.Regexp
	Boundary string
	Window   int
}

func (g NegationGuard) NegatedAt(text string, pos int) bool {
	window := text[maxInt(0, pos-g.Window):pos]
	if i := strings.LastIndexAny(window, g.Boundary); i >= 0 {
		_, size := utf8.DecodeRuneInString(window[i:])
		window = window[i+size:]
	}
	return g.Cue.MatchString(window)
}

var claimNegationGuard = NegationGuard{Cue: negationCueRe, Boundary: negationClauseBoundary, Window: negationWindowBytes}

func negatedClaimMatch(block string, start int) bool {
	return claimNegationGuard.NegatedAt(block, start)
}
```

`negatedClaimMatch`'s two call sites and every claims-layer test pass **unmodified** — this is the zero-behaviour-change proof for the claims layer (T-0). `negationCueRe`/`negationClauseBoundary`/`negationWindowBytes` stay unexported.

## 3. The fabrication-domain cue vocabulary

Lives in `check_tool_fabrication_action.go`, next to the Tier A regexes:

```go
var fabNegationCueRe = regexp.MustCompile(
	`(?i)\b(?:no|not|never|nor|none|without|zero|cannot|instead of|rather than)\b` +
		`|[a-z]n['’‘]t\b` +
		`|\b(?:cant|dont|doesnt|didnt|isnt|arent|wasnt|werent|wont|couldnt|shouldnt|wouldnt|mustnt)\b`,
)
```

| cue | why |
|---|---|
| `not`, `never`, `nor`, `cannot` | carried from `negationCueRe` — negates a following predicate in any domain. |
| bare `no`, `none` | the motivating payload ("**no** fabricated data"). Claims-layer intensifier counter-examples don't transfer: a code comment saying "no X" about its own data is a denial. |
| `without` | "without fabricating data" — a direct prompt-echo shape. |
| `zero` | "zero fake data" — from the bug file's own candidate list. |
| `instead of`, `rather than` | "we fetch real data **instead of** generating it" — matches `fabDataNearQualifier`; negates the token it immediately precedes. |
| `n't` + apostrophe-less stems | shape copied from `negationCueRe`, plus `mustnt`; curly apostrophes kept for typographic copy. |

Deliberately absent: `avoid` ("to avoid an empty state we generate placeholder rows" is itself a declaration); `fails to`/`unable to` (the vetcomp class verbatim — a fetch *failing* is what licenses the fabrication); `rarely`/`seldom` (hedges that still admit the act).

## 4. The qualifier-position problem, handled per regex

Negation must be checked at the **qualifier token's** position, not the match start — a cue regex spanning subject-to-verb can put the negator *inside* the matched span (CLM-019). Per regex:

- `fabQualifierNearData` — qualifier is group 1, at match start.
- `fabDataNearQualifier` — data-noun is group 1, **qualifier is group 2**, mid-span ("records are never generated" — match starts at "records", "never" sits inside it).
- `fabGenerateVerbData` — wrap `generat\w+` as group 1 (currently unparenthesised; nothing else reads its groups today).

```go
type declPattern struct {
	re        *regexp.Regexp
	qualGroup int
}

var (
	declQualifierNearData = declPattern{fabQualifierNearData, 1}
	declDataNearQualifier = declPattern{fabDataNearQualifier, 2}
	declGenerateVerbData  = declPattern{fabGenerateVerbData, 1}
)

func firstAssertedDeclaration(text string, p declPattern) (asserted string, suppressed []string) {
	for _, m := range p.re.FindAllStringSubmatchIndex(text, -1) {
		qStart := m[2*p.qualGroup]
		if qStart >= 0 && fabNegationGuard.NegatedAt(text, qStart) {
			suppressed = append(suppressed, text[m[0]:m[1]])
			continue
		}
		if asserted == "" {
			asserted = text[m[0]:m[1]]
		}
	}
	return asserted, suppressed
}
```

Gating still reads only `declSignals`/`corpusSignals` exactly as now; suppressed matches are merged into `res.Signals` (prefixed `"negated declaration ignored: "`) after the gating decision — informational only, never counted toward `Fabricated`. The synthetic-PII arm (`fabSyntheticPII`) is untouched.

**Pinned residual (deliberate, CLM-017-style):** post-positioned denial — "Mock data is **not used** in this tool" — stays convicted; a backwards scan cannot see a negator that follows its qualifier. Handling it needs a second, forward-scanning algorithm with its own false-suppression surface. Cost is one human-review item (fail-safe direction), not a shipped fabrication, and the lane's existing prompt-side workaround already covers it. Pinned by test T-7.

## 5. Window and clause boundary for the fabrication domain

```go
const fabNegationWindowBytes = 32
const fabNegationClauseBoundary = ".!?;:,<>\n\r\t–—(){}[]"

var fabNegationGuard = datahelpers.NegationGuard{Cue: fabNegationCueRe, Boundary: fabNegationClauseBoundary, Window: fabNegationWindowBytes}
```

32 bytes clears the longest observed cue-to-qualifier gap ("instead of dynamically generating", 24 bytes) with slack; deliberately half the claims layer's 64, because this domain's cues include bare "no"/"without" whose false-suppression risk grows with distance. Boundary set adds brackets to the claims set — this domain scans raw JS/HTML, where a bracket ends a comment clause; each added boundary only shrinks the guard's reach (conservative in the convicting direction). The motivating fixture keeps cue and qualifier inside one bracket group, so it is unaffected.

## 6. Tests

All in `check_tool_fabrication_action_test.go` unless noted.

- **T-0** (no new code): `go test ./platform/orchestration/datahelpers/` green with **zero test-file edits** after the claims.go change — the mechanical proof of no claims-layer behaviour change.
- **T-1** `TestDetect_DeniedFabricationComment_NotGated` — the bug-222 payload. `Fabricated == false`; `Signals` carries a `negated declaration ignored:` note.
- **T-2** `TestDetect_RealDeclarationWithUnrelatedNegatorElsewhere_StillGated` — vetcomp confession plus an unrelated negated sentence in a different clause. Stays `Fabricated == true, Tier == "declaration"` — proves the guard is positional, not string-global.
- **T-3** `TestDetect_NegatorInsideMatchSpan_NotGated` — "records are never generated…" / "this tool does not generate records…" (the `fabDataNearQualifier`/`fabGenerateVerbData` shapes). Fails against a naive scan-from-`loc[0]` implementation — the direct CLM-019 pin.
- **T-4** `TestDetect_DenialVocabularySweep_NotGated` — table test over the cue set: "no mock data", "without fabricating any data", "never invents entries", "zero fake records", "instead of generating it", "doesn't seed the dataset".
- **T-5** — existing suite regression, unmodified: `TestDetect_VetcompFabrication_Gated`, `TestDetect_IntroducedSyntheticPII_Gated`, `TestCheckToolFabricationAction_ReadsDottedConfigPath`, all 12 current cases.
- **T-6** mutation-and-control triple, run during implementation, recorded in NOTES:
  - Mutation A (`NegatedAt` → always `false`): T-1/T-3/T-4 fail; T-2/T-5 pass.
  - Mutation B (`NegatedAt` → always `true`): T-2, vetcomp, dotted-config-path tests fail; PII test still passes.
  - Control (`fabLiteralRecordThreshold` 15→14, unrelated Tier B knob): everything above passes. Revert all three.
- **T-7** `TestPostPositionedDenialIsAKnownResidual` — "Mock data is not used in this tool." stays convicted; comment instructs a future session to delete+promote if the guard ever learns forward negation.
- **T-8** (datahelpers) `TestNegationGuardIsParameterisedByDomain` — a custom-cue `NegationGuard` (including bare "no") suppresses where `claimNegationGuard` does not, over the same string.

## 7. Risk register

- **R1 — a defensively-phrased real fabrication is now Tier A-blind.** Accepted: Tier B is unaffected (early return only fires on non-empty `declSignals`) and the vetcomp incident carries four independent Tier B signals plus corroboration; the synthetic-PII arm is untouched (proven by Mutation B). Residual: a fabrication with a defensive comment AND no PRNG/builder/fragment/fetch signature AND <15 literal records — small, and Tier A was never the sole defence against it.
- **R2 — vocabulary drifts wider over time, quietly disarming Tier A.** Mitigated structurally: DELIBERATELY ABSENT block with counter-examples, T-2 as a string-global tripwire, Mutation B on record, suppressed matches surfaced in `Signals`.
- **R3 — claims-layer regression via the extraction.** Zero by construction (§2: identical-body wrapper, unexported vocabulary), proven by T-0.
- **R4 — false-positive direction preserved.** Every failure mode routes to `needs_human_review`, never to deploying a fabrication; the guard can only *remove* convictions; the fail-safe empty-input branch is untouched.

**Out of scope, noted not folded in:** `invent\w+` matches "inventory"; `realistic` has no left `\b` (matches "unrealistic"). Pre-existing precision defects, unrelated to negation — file separately if they ever bite.

## 8. Council-gate framing

`check_tool_fabrication`'s declaration tier convicts denials as declarations (bugs_open/222). Fix: extract the CLM-017 negation-guard *algorithm* as an exported, parameterised `NegationGuard` (claims layer's cue regex/window/boundary/both consumers untouched — mechanically proven by T-0/T-8), applied in the fabrication gate with a domain-tuned cue vocabulary at the qualifier's submatch position (CLM-019 lesson). Not architecture-scope under the 2026-07-29 ruling: adds an opt-in helper reachable by nothing until a caller names it; changes no shared mechanism's guarantee; the one behaviour that changes is the tool-recreation gate's own false-positive rate, in the reduction direction only (mutation-tested). Consumers told: the claims layer (no guarantee change) and `mortgagecalculator_couk_adoption` (whose workaround this retires).

## 9. Execution order

0. Pre-flight: grep LANDMINES for this file/table footprints; re-check `git status` (stale by now); check `site_work_items` for open work on this gate.
1. `claims.go` — add `NegationGuard`/`NegatedAt`/`claimNegationGuard`; rewire `negatedClaimMatch`. Verify T-0.
2. `claims_test.go` — add T-8.
3. `check_tool_fabrication_action_test.go` — add T-1–T-4, T-7 **first**; confirm T-1/T-3/T-4 fail red against unfixed code.
4. `check_tool_fabrication_action.go` — cue regex, consts, `fabNegationGuard`, `declPattern`/`firstAssertedDeclaration`; parenthesise `generat\w+`; rewire Tier A chain; merge suppressed notes post-decision.
5. Mutation runs (T-6), revert each, record in NOTES.
6. `go test ./platform/...` against a clean `git archive HEAD` overlay if the shared tree is dirty elsewhere.
7. Council submission per §8; save `SUBMISSION_CORR`.
8. Commit by pathspec: the two source files, two test files, register file, `bugs_open/222_…`, workstream NOTES/README. `Council-Submitted: <corr>` trailer.
9. Ship: bump `IMAGE_TAG`, `make build-agent-chassis`, pod-grep the added cue-regex literal on both replicas; behavioural check same-day via `orchestration_states.collected_data` on a re-run recreation item.

## 10. Concept-register follow-up (not written yet)

One entry in `claims-verification.md`, CLM family, cross-linked from CLM-017: `NegationGuard` exported/domain-parameterised, first non-claims consumer is the fabrication gate. Update `000_concept_index.md` count. 016b §9's existing entry for this bug gets one annotation after the fix is live, pointing at the mechanism-level rule rather than restating it.
