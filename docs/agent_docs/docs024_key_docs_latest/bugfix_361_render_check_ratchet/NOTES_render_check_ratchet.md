# NOTES — bugfix_361 render-check ratchet (append-only, newest at the bottom)

## 2026-09-03 — lane opens

**Why this bug.** Picked from `bugs_open/` as OPEN + UNOWNED. Ownership established three
ways, because no single instrument answers it (and a peer session logged a WRONG_CALL today
for trusting one that does not):
1. the bug file's own status line — `**Status: OPEN, UNOWNED**`;
2. `scripts/who-owns.py 361` → originating lane `bugfix_140_contact_info_fabrication`,
   whose last commit is **2026-08-07** (27 days cold), and `git log` on the bug file itself:
   last touched 2026-08-22;
3. **`ListAgents` — 65 live peer sessions, none on 361.** This is the check no commit-reading
   tool can supply, and CLAUDE.md says so explicitly of `who-owns.py` ("reads COMMITS, so a
   session mid-fix is invisible").

**The bug re-verified first-hand before any code was touched** (it was 12 days old):
- `kubectl get cronjob component-render-check -o json` → `lastSuccessfulTime =
  2026-08-09T06:55:21Z`. **25 consecutive red days**, not the 12 the file records.
- `doc_notes` series (`source='component_render_check'`, 2 rows/day = red, 1 = green): NEW
  findings **227 → 478** since filing; active components **282 → 497**. The bug has grown.

**Instrument.** Dumped the live library to a JSON fixture and ran the tool offline:
`--json <fixture> --compare` reports **478 NEW**, which **matches the cluster's own row for
this morning exactly**. That agreement is what licenses every later measurement offline.
⚠ The plain `kubectl exec … psql` stream **truncated at 6.5 MB** — the documented flake, and
it fails as a JSON parse error, not as a non-zero exit. Route used instead:
`… | gzip -9 | base64 -w0` inside the pod, decode locally. See RUNBOOK.

## 2026-09-03 — the bug file's own fix candidate 1 has a hole, and the artefact proves it

361 §4 candidate 1: *"a finding whose component owns **zero keys** in the baseline is not NEW
— it is unbaselined"*. Not sufficient. `baseline.json`'s own note says **"1023 findings across
139 analysed components"**, and its keys span only **115 distinct components** ⇒ **24
components were analysed and CLEAN** at baseline-cut time. A keys-derived covered set cannot
see those 24, so a component that was clean and later **regresses** would be filed as
unbaselined and would **not** fail — the exact event a ratchet exists for.

`[MEASURED 2026-09-03]` 0 of the 24 regress today, so this is **latent, not live damage**.
Recording it anyway because it is 24 wide now and grows with every regeneration (analysed-and-
clean is the healthy majority: today 470 analysed, ~115 with findings).

**So the fix is not "scope by component" but "make the baseline record COVERAGE, not just
findings"** — which is what makes the bad state unrepresentable rather than unlikely, per
CLAUDE.md's ordering rule.

## 2026-09-03 — result, measured

| | HEAD | after the fix |
|---|---|---|
| NEW (fails the run) | **478** | **18 REGRESSION** across 5 components |
| growth, reported not failed | — | **460 unbaselined** across 62 components |
| exit code | 1 | 1 |

The job **stays red**, and that is the correct outcome: it is now red for 18 findings in 5
named components instead of 478 it manufactured. Clearing it is the debt decision 361 §4
reserves for the owner. The five: `blog-listing_pre_037` (361 §2(c) already established its
template was rewritten), `social_proof`, `tool-ab-test-calculator_pre_037`,
`tool-equity-release_pre_037`, `tool-gas-unit-converter-gaswholesalers-com`.

## 2026-09-03 — MISSTEP: a mutation-proof test that could not fail

Full entry in `WRONG_CALLS.md`. Short version: the canonical-coverage test asserted through
`loadBaseline`, which **re-canonicalises what it reads**, so a write-side mutation (record raw
names) was repaired by the read side before the assertion saw it — two guards in series, and
the test passed under the very defect it was named for. Three other mutations discriminated
correctly. Fixed by asserting on the **written JSON artefact** instead of the round-trip; the
mutation now fails. This is the tool's own register warning (*"do not verify with a
self-comparison"*) reproduced one layer up, in the test.

## 2026-09-03 — a second instrument of mine that could not come out false

Checking whether `cmd/component-render-check/` is in council scope, I first ran
`grep -qE "$COUNCIL_SCOPE_RE"` — a variable that **does not exist** in
`scripts/council-scope.sh` (the real names are `COUNCIL_SCOPE_CODE_RE` /
`COUNCIL_SCOPE_MIGRATION_RE`). An empty pattern matches everything, so it printed **"IN
SCOPE"** for all three paths I gave it, including a deliberate control. Redone with the
script's real `in_council_scope` function and both a positive control (`cmd/config-key-audit`,
`platform/`) and a negative one (a `.md`): my path is **OUT of scope**, correctly, so no
council submission is owed and none would have spent credits.

**Also noted, not mine to fix:** CLAUDE.md's council-scope paragraph lists
`cmd/config-key-audit/` and `scripts/pattern-check.py` as the `cmd/`+`scripts/` widenings; the
live `COUNCIL_SCOPE_CODE_RE` also carries **`^cmd/brief-negation-check/`**. The script is the
single source for the decision, so the doc is what has drifted.

## 2026-09-03 — the Fable review found TWO real defects in my own first cut

I committed `051c73d1e` and *then* the plan review landed. It was right twice, and both were
verified first-hand at the code before I acted on them (a subagent report is another doc, not a
measurement).

**Defect 1 — the covered set missed STATIC templates.** I collected coverage at `checked++`,
which is *after* `if len(an.root.children) == 0 { continue }` (`:747-750`). So a template with
no actions was never vouched for, and one later rewritten to reference a field and render a
hole would have been filed as unbaselined and failed nothing — the check's own stated signal,
exempted by my fix. **27 components today, 37 at baseline cut — larger than the 24-component
hole the whole fix exists to close.** Coverage is now recorded *before* the static skip.
`[MEASURED]` end-to-end on a 2-component fixture: `1 analysed, 2 covered`.

**Defect 2 — canonical storage exempted an EDITED clone.** I stored coverage canonicalised, so
a clone collapsed onto its representative. But an edited clone stops matching the hash and gets
its own identity — the tool's own note says its findings reporting as NEW "is correct" — so
nothing vouched for it and it silently stopped being watched. Coverage is now RAW, with a
`covers()` that reads raw first then the representative. **My test had pinned the wrong
behaviour** (asserting clone and representative collapse to one entry) — corrected in place
with the reason written into it, not deleted.

**Why the live numbers did not move** (18/460, exit 1, before and after): against a *legacy*
baseline coverage is derived from keys either way, so raw-vs-canonical can only bite once a
baseline is regenerated. That is exactly why it would have shipped silently — the arm was
unreachable on today's artefact, and a run today cannot distinguish the two versions. **A
measurement that cannot come out false was available and I nearly took it as reassurance.**

**Also adopted from the review:** `"components"` present-but-EMPTY is now refused (exit 2),
distinct from absent. Absent = legacy artefact, falls back; empty = a ratchet switched off by
hand, where nothing can ever fail. The tool cannot produce it — `omitempty` writes an empty
slice as *absent* — which my test discovered by failing, and is why that fixture is raw JSON.

**Not adopted, deliberately:** the review's suggestion to rename the JSON field `new_findings`
→ `regressions`. The human-facing labels DID change (`NEW` → `REGRESSION` in the summary,
stdout and doc_notes), so the discontinuity is visible in the series where a reader will meet
it. Renaming the JSON key as well buys nothing today — grep finds no machine consumer — and
costs the one continuity a future reader might rely on. Recorded so the choice is visible.

## 2026-09-03 — an operational finding that blocks the green run

`--write-baseline` on today's live library **refuses**: *"2 component(s) failed to parse and
are therefore uncovered"*. That guard is correct and predates this fix (baselining a run that
silently dropped components would bake the drop in as clean). But it means **step 3 of the
sequence — regenerate, so the job can go green — cannot be done until those two templates are
fixed.** They are listed by a plain `--json <fixture>` run under `unanalysed`. Handing this to
the owner rather than working around it: suppressing the guard to get a green baseline is
precisely the "fix that turned the check off" shape.
