# 222 — the fabrication gate's declaration tier convicts a comment that DENIES fabrication, and the recreate prompt makes such comments likely

**Filed 2026-08-08 by the mortgagecalculator adoption lane.** Status: FIXED +
LIVE (chassis v1.0.1269, 2026-08-08 ~23:10 BST, pod-verified both replicas —
see "Fix landed" section below). Council round 2 in flight, round 1 blocked
on a fleet-wide credit outage, not on this submission. **Kept in `bugs_open/`
per owner direction (2026-08-06)** — do not move to `bugs_closed/`.
Severity: medium — it discards paid-for, correct recreations and files
`needs_human_review` items whose review burden lands on the owner; it does not
damage live pages (the artefact is refused, nothing is overwritten).

## Diagnosis provenance (per the 2026-07-31 owner ruling)

Not run through 090 — **stated substitution**: this filing reproduces the
conviction live, quotes the deciding regex, and preserves the convicted
payload whose single matching line is a negation. The mechanism claim is one
regex and one input; every leg is quoted below and independently checkable in
minutes. (The 090 loop would re-read `check_tool_fabrication_action.go:91` and
the payload and conclude the same; there is no cross-cutting uncertainty here.)

## The incident (2026-08-08 ~16:45 UTC, mortgagecalculator.co.uk)

The id-alignment recreation of `tool-portfolio` (work item
`07b7eca3-19a7-48f2-aa82-bdde24502958`) was convicted by `check_fabrication`:

```json
{"tier": "declaration", "fabricated": true,
 "signals": ["declared synthetic/fake data: fabricated data"],
 "detail": "The recreation declares or introduces invented data — this is fabrication regardless of the original tool."}
```

The recreation's ONLY text matching the detector (preserved payload, line 316;
copy at the lane's
`mortgagecalculator_couk_adoption/acceptance/` — see NOTES 08-08 evening):

```js
// In-memory portfolio store (no fabricated data — starts empty)
```

The tool starts empty and every record is user-entered — there is no invented
data anywhere in the payload (16,623 chars, id-complete 15/15 against the
golden, completion marker present). The recreation was discarded on a run that
otherwise passed every gate, and a `needs_human_review` item
(`3d11e960-1951-4b7a-841e-aa538666e3c2`) was filed.

## Root cause

`platform/orchestration/actions/check_tool_fabrication_action.go:91`:

```go
fabQualifierNearData = regexp.MustCompile(`(?is)(synthetic|fabricat\w+|\bfake\b|\bdummy\b|\bmock\b|placeholder|realistic|deterministic|invent\w+|made[-\s]?up)[^\n]{0,48}(dataset|data[-\s]?set|\bdata\b|records|entries|...)`)
```

A qualifier within 48 chars of a data-noun convicts — **with no negation
awareness**. "no fabricated data", "never invent records", "do not use mock
data" all match identically to the declarations the tier exists to catch.

**The prompt makes the false positive likely, not freak.** The recreate
prompt's Data Integrity section (`recreate_tool` step, `agent_definitions`
type `tool-recreation-handler`) spends ~9 lines insisting the model must not
fabricate/invent/seed data. A conscientious model echoes the prohibition as a
code comment — the more emphatic the prompt, the more likely the echo. This is
the `prompt-text-poisons-its-own-detector` class (memory topic file of that
name): the instruction's vocabulary becomes the detector's conviction.

Corroborating history: the SAME page's recreation was convicted by the same
gate on 2026-08-05 12:55 (item `aca92097`, signals purged unread — so
unknowable now, but the prior for "same benign echo" just went up sharply).
The 2026-08-08 morning re-run of the same page passed with `fabricated:false`
— the model happened not to write the comment that time. A gate whose verdict
flips on comment phrasing at temperature 0.1 is measuring prose style, not
fabrication.

## Fix candidates, ordered by what closes the door

1. **Make the declaration scan negation-aware at the match site**: reject a
   match whose qualifier is preceded (within ~16 chars, same line) by a
   negator (`no |not |never |without |don't |do not |zero `). Cheap, targeted,
   testable with the exact payload above as the fixture — and add the inverse
   fixture (a REAL declaration, e.g. "seeded with realistic data") to prove
   the scan still convicts. Note `mutate-the-code-to-prove-the-guard`.
2. Scan only code/comment-stripped PROSE for declarations (the 218 defect-A
   fix built exactly this discipline for the placeholder scan — assertion-text
   blocks); a declaration in a comment about *what the code does not do* never
   reaches a visitor. Bigger change; tier B (corroborated corpus) already
   covers actual generator code.
3. Do nothing and eat the review items — rejected: the gate's false positives
   land on the owner's review queue, and each discards a completed opus run.

## Verify

Unit: feed the preserved payload to `DetectToolFabrication` — must return
`fabricated:false` after the fix; feed a true declaration — must stay
convicted. Live: re-run the portfolio recreation item; `check_fabrication`
output in `orchestration_states.collected_data` (read same-day; it purges)
must show `fabricated:false` when the only match is a negation.

## Workaround (until fixed)

The lane's next portfolio recreation item adds to its spec's mandatory
requirements: write code comments that describe what the code DOES, and do not
use the words fabricated/synthetic/fake/mock/dummy/placeholder near the word
data in comments. This dodges the false positive without weakening the gate
(the gate still runs, and real fabrication is still code, which tier B reads).

## Fix landed 2026-08-08 evening, submitted to council, INERT until the next chassis roll

Picked up by a separate lane (`bugfix_222_fabrication_negation`); full plan,
notes and runbook at
`docs/agent_docs/docs024_key_docs_latest/bugfix_222_fabrication_negation/`.
Fix candidate 1 taken, but **not as a one-off local regex** — this platform
already had a negation guard for exactly this class (`negatedClaimMatch`,
CLM-017 in `datahelpers/claims.go`), so the algorithm was extracted as an
exported `NegationGuard` and reused, with a fabrication-domain cue vocabulary
(deliberately including `no`/`without`/`zero`, which the claims layer excludes
on purpose for unrelated reasons) applied at each Tier A regex's qualifier
submatch position — a second regex-position subtlety (CLM-019) that the
original candidate 1 sketch did not anticipate. Registered as CLM-020.

Reproduced first (five new tests run red against unfixed code, confirming
each fixture actually triggers the bug — two of my own fixture-authoring
mistakes were caught this way), then fixed, then mutation-and-control tested
against a clean `git archive HEAD` overlay. Zero behaviour change to the
claims layer, proven mechanically (its full suite passes with no test file
edited). Council submission `SUBMISSION_CORR=aa2d0d62-4aba-480e-aedc-8be264d53b01`.

**Verify once live:** re-run the portfolio recreation item (or feed the exact
preserved payload above to `DetectToolFabrication`) and confirm
`fabricated:false`. Once confirmed, the workaround clause above can come out
of the recreate spec.

**CONFIRMED LIVE 2026-08-08 ~23:10 BST.** Pod-verified both
`agent-chassis-778b7c77c7-*` replicas on chassis `v1.0.1269`: `NegationGuard`
and `negated declaration ignored` present (absent on the prior build,
v1.0.1268, checked before this roll — positive control `declared
synthetic/fake data` present on both, confirming the grep methodology).
`mortgagecalculator_couk_adoption` told; their comment-style workaround
clause can come out. Council round 2 submitted after the owner's credit
top-up (round 1 was blocked fleet-wide, not a finding against this
submission); verdict pending, correlation `aa2d0d62-4aba-480e-aedc-8be264d53b01`.
