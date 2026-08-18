# NOTES — bugfix 146 (ported tool pages outside every acceptance tier)

Append-only, newest at the bottom. Bug: `bugs_open/146_HANDOFF_2026-07-29_ported_tool_pages_are_outside_every_acceptance_tier.md`
(⚠ 146 is an AMBIGUOUS number — the OTHER 146 is `eleven_tool_pages_are_unreachable…`. This lane is the
`ported_tool_pages_are_outside_every_acceptance_tier` slug ONLY.)

## 2026-08-17 — lane opened; picked as the next unowned open bug

Session was asked to take on 080/081; both are CLOSED at the fixed-AND-live bar (081 on
v1.0.1225 2026-08-01; 080 on v1.0.1250 2026-08-04, detector induced). Bulk-scanned all 108
`bugs_open/` files for commits touching the file in 14d: only 085 (done in substance, residual
owned by brochure lane), 096 (its own header says CLOSED 07-28, never moved), 131-gauntlet
(vonc lane owns the site) and 146-ported-tools survive. who-owns on the 146 slug: no commits
touching or about it, "owning" hits are stale citations. Live-transcript grep: the only session
that read it recently was the 281 lane closing ITS bug (08-16), which cited 146 as prior art and
left follow-ons to others. → 146 taken.

## 2026-08-17 — re-validation against the live DB (queries run this session)

The bug was filed 07-29; the terrain has moved. Measured today:

- **Eligible population for `tool_acceptance_due` is now 72 rows (was 17)** — same query as the
  bug file (component_level='tool', active, deployed, LEFT JOIN doc_plans criteria). webdesign.co.uk
  now contributes **5** tool-level components (was ZERO): aspect-ratio, css-specificity-calculator,
  css-unit-converter, html-minifier, markdown-tables. None of the 7 overflowing pages is among them.
- **All 7 named pages are unchanged**: deployed, active, function='ported-page',
  component_level='section'. Still structurally outside the tier population.
- **Fleet-wide uncovered set: 89 deployed active `/tools/…` pages with NO active tool-level
  component**, by site: webdesign.co.uk 59, mortgagecalculator.co.uk 10, leopardessconsulting 4,
  finetuning.uk 3, loancash 3, ai-agent-orchestration 2, robot-hands 2, fundamentallyai 2,
  idea.uk 2, vonc 1, gaswholesalers 1. [MEASURED today; note the predicate counts pages, and
  `/tools/index.html` was excluded — some rows may be guides/index pages, not tools; refine
  before quoting as "89 tools".]
- **A `tool_health` check now EXISTS and files `ported_tool_fix` items** (19 at
  needs_human_review): webdesign.co.uk 13 (08-16, created_by `generic`), gamesdesign.co.uk 6
  (08-17, created_by `design-discovery-agent`). So "no tier can ever see them" is no longer
  bluntly true — but the check's findings are static heuristics (no @media breakpoint / external
  CDN / fetch() / needs_rebuild state).
- **The heuristics miss 5 of the bug's 7 measured overflow pages.** Only vibe-equalizer and
  smooth-shadow drew `ported_tool_fix` items; recommender-engine, layout-generator, css-variables,
  social-card, blob-maker (fixed-width overflow, the geometry-only signal) drew nothing. The
  discriminating instrument (`no_horizontal_overflow` at 390×844 in the browser-runner) still
  never runs against ported pages. **This is the sharpened residual claim — the bug survives.**
- `needs_diagnosis` queue: EMPTY (dedup check done).
- Also on these pages in the queue: many `misdirected_cta` page_rerender items (CTA lane's class),
  a `vision_finding` on vibe-equalizer, 3 failed `page_rerender … _assemble` items — other lanes'
  machinery, not this bug.

Adjacent lanes (must not compete):
- **webdesign_tool_rebuilds** (owner-directed 08-15): native replacement of the 63 ported tools on
  webdesign.co.uk; blocked on a roll (`bugs_open/286` fix inert until roll + seed 435 HOLD). When it
  completes, webdesign's instance half retires by attrition. It does NOT address the fleet-level
  structural gap (gamesdesign, mortgagecalculator etc. keep ported/uncovered tools).
- **281 lane** closed 08-16 (tool audit per-instance); its sweep is what censused ported_tool_fix.
- **291 lane**: ported_tool_fix/audit findings were re-parked `needs_human_review` because the
  named review handler was a phantom; phase 2 committed, inert until roll.
- **mortgagecalculator adoption** A4: writing 12 tool PLANs (criteria fences) — their tools' route
  into the acceptance population is their lane's work.

## 2026-08-17 — live re-scan: ALL 7 STILL OVERFLOW; and my stored-HTML read was the wrong artefact

Re-ran the bug's own scanner (`gauntlet_dead_cta/p4_sources/scan_clipped_tools_2026-07-29.py`,
clause extracted from `run_checks_action.go` at runtime, 390×844) over the 7 live URLs:
**0 clean · 7 FLAGGED · 0 errors** — same culprits as filed (e.g. social-card `strong (455px)`,
vibe-equalizer now `over=361` vs 11 filed, forcedBy `div.card-img` 324px). Output:
scratchpad `scan_146_rerun_2026-08-17.txt` (copy below in this dir if needed).

**Near-miss worth keeping:** the previous entry noted the 7 `ported-page` components'
`rendered_html` was updated 08-15 and no longer contains the filed width literals — I was one
step from concluding "instance half repaired by other work". The live scan refutes that:
the served pages still carry the fixed widths, so **`page_components.rendered_html` is NOT where
a ported page's tool markup lives** (the served bytes come from elsewhere — establish where
before ever asserting a ported page's state from that column). The check that saved it: measure
at the ARTEFACT the visitor gets, never the stored row ([[a-print-statement-is-not-a-config-row]]
family; "trust the rendered artefact, not the status").

**Bug 146 re-validated in full: VALID.** Live damage (7 pages overflowing on mobile, ~3 weeks
after filing) + the structural gap (geometry acceptance checks never run on ported pages;
tool_health's static heuristics caught 2 of 7).

## 2026-08-17 — research complete; TWO corrections to this file's own earlier entries

> **CORRECTED 2026-08-17:** the first entry's bullet "Eligible population for
> `tool_acceptance_due` is now 72 rows (was 17) … still structurally outside the tier
> population" measured the WRONG predicate. I re-ran the BUG FILE's query
> (`component_level='tool'`), but the live gate has not been that since `ac9f75a0c`
> (2026-07-29 17:19, TL-033 — **three hours after 146 was filed**): the shared
> `toolEligibilityWhere` (tool_eligibility.go:97-117) admits ported sole-component
> `page_type='tool'` pages, keyed `regexp_replace(p.name,'^tool-','')`. My measurement
> encoded the question the bug file asked in July, not the code's question today —
> exactly the `[[narrow-filter-defines-the-conclusion]]` trap the bug file itself
> documents. Caught by the code-research pass before it reached the bug file or a
> commit. Logged in WRONG_CALLS.md.

**The live mechanism, measured with the RIGHT predicate (clause-b census, this session):**
67 ported sole-component tool pages fleet-wide: webdesign.co.uk 59 (14 with a current
```criteria``` PLAN), loanandmortgagecalculator 3 (3 fenced), loancash 3 (0),
mortgagecalculator 1 (1), vonc 1 (1). The 14 fenced webdesign ported tools each drew exactly
one `acceptance_run` (53 complete fleet-wide) and verdict notes. So ported tools DO enter
Tier 4 — when fenced.

**The two live doors that keep 146's symptom alive:**
1. **No fence → no run.** 6 of the 7 overflowing pages have no criteria fence → 0 runs,
   0 notes, invisible to the only instrument (`no_horizontal_overflow`, Tier-4-only) that
   can measure their defect. `tool_health`'s static proxies caught 2 of 7.
2. **Fenced + FAILED → silent sink.** `JudgeAcceptanceResultsAction`
   (tool_acceptance_actions.go:866-873) resolves the component by
   `WHERE cc.function = <subject key>`; a ported instance's component is `ported-page`, so
   `componentID=""` → the else-arm (:1029-1040) logs "route manually", writes the
   acceptance-fail note, files NOTHING. **Live evidence: `pasteboard` FAILED 08-05 AND
   08-14; `vibe-equalizer` FAILED 08-05 AND 08-14 (the `no_horizontal_overflow` defect this
   very bug measured, div#preview-card 380px forced by div.card-img 324px) — four failing
   verdicts, zero work items.** The vibe-equalizer `ported_tool_fix` row that DOES exist
   came from tool_health's static `no_responsive` heuristic on 08-16, not from either
   acceptance failure. This is 281's Finding B, which their handoff leaves OPEN, UNOWNED.

**Fix chosen (framework-first):** make the Tier-4 judge the THIRD producer of
`ported_tool_fix` (after check_tool_health.go:278-293 and check_tool_acceptance.go:292-307),
key `ported_tool_fix:tool_acceptance_tier4:<subjectKey>:<siteID>`, status
needs_human_review, no handler — the owner-ruled converging-producers shape (2026-08-02 §1:
no RFC needed provided the register names the producer set + key shape; TL-042 updated in
the same commit). The fence gap (door 1) is an OWNER DECISION on cost (auto-fence ported
tools with Tier-4-only baseline criteria vs let the rebuild lane retire them) — written up
in the PLAN, not taken unilaterally.

Diagnosis-loop note (owner ruling 2026-07-31): 090 NOT filed for the door-2 root cause, and
here is the stated substitution — the mechanism was independently diagnosed by the 281 lane
(their Finding B, recorded in bugs_closed/281 and their 08-16 handoff), and this session
re-verified it first-hand at BOTH halves: the deciding arm read at
tool_acceptance_actions.go:1029-1040, and the failing branch observed live (4 FAILED
verdict notes whose own Fix: line says "route this manually", 0 resulting items). The claim
is local, self-evidencing, and twice-independently established.

## 2026-08-17 — CHECKPOINT: re-validation + plan committed; code NOT started (tree left clean)

Session hit its usage checkpoint mid-implementation. State: everything above is done and
committed; `platform/orchestration/actions/tool_acceptance_actions.go` is UNTOUCHED (one
exploratory line was added and reverted in the same minute — the tree compiles as before).

**Exact resume point** (all reads already done, cited in PLAN):
1. Code: in `JudgeAcceptanceResultsAction`'s `componentID==""` else-arm
   (tool_acceptance_actions.go:1029-1040), add the ported route per PLAN § "Door 2":
   declare `portedRouted := false` beside `escalated`; helper
   `routePortedAcceptanceFailure` reads `input_data.spec.component_id` (fixed path, NO new
   config key — keeps RFC_022 surface flat), resolves
   `SELECT component_level FROM content_components WHERE id=$1::uuid AND is_active`, files
   `ported_tool_fix` ONLY when level <> 'tool' (acceptance_stuck's exact ON CONFLICT
   DO-UPDATE-merge idiom, handler '', status needs_human_review, key
   `ported_tool_fix:tool_acceptance_tier4:<fn>:<siteID>`); new `case portedRouted:` in the
   fixLine switch BEFORE the noAutoFix/stuck cases; `out["ported_tool_fix_filed"]=true`
   key-presence-only.
2. Tests: new `tool_acceptance_ported_sink_test.go` reusing `judgeRun`/`captureArg` from
   tool_acceptance_convergence_test.go — its driver must ADD `component_id`/`page_id`/
   `page_name` to `input_data.spec` and mock the extra component-level query between the
   fork lookup and the insert. Firing arm + two negative controls (no spec.component_id →
   byte-identical today-behaviour; level='tool' → no item). Existing suite stays green by
   construction (its driver never supplies spec.component_id). Mutation-check the guard.
3. Register: TL-042 in docs026 register/tool-lifecycle.md — add the judge as THIRD
   producer with the new key segment, and correct its stale "first sweep census pending"
   clause (sweep ran 08-16 10:08Z: 13 ported_tool_fix + 12 audit_tool).
4. 016b §9: the transferable pattern — a subject-key resolver left keyed on the OLD
   population after a widening silently drops the new members' outcomes (kin of 149's
   "widening what REACHES a function breaks it unedited").
5. Council: submit via 097 (rationale = PLAN), `Council-Submitted:` trailer if committing
   before the verdict. Commit Go+test+register by pathspec.
6. Owner decision (PLAN § Door 1, fences) + dated pointer in webdesign_tool_rebuilds NOTES
   re the seven pages — outstanding, not started.

## 2026-08-18 — door 2 BUILT, committed `1549dc58b`, council corr `d2edf61d`; fresh build verified first

- **Fleet baseline re-verified at the binary before anything else**: pods on v1.0.1309
  (started 15:45Z), full sha `f0117fb8b93ea3e1f32298daeb9751bcff4b90c7` present 3× in
  /proc/1/exe, invented-sha control 0. A real fresh build (not the same-tag cache trap).
  My fix postdates it → inert until the NEXT roll.
- **Implementation** exactly per PLAN: `routePortedAcceptanceFailure` helper +
  else-arm call + fixLine case + `ported_tool_fix_filed` result key (key-presence).
  No new step-config keys (fixed input paths) → RFC_022 surface unchanged.
- **The working tree's actions package would not compile the tests** — another session's
  in-flight edit (`component_write_guard.go` referencing undefined `balancedPairs`).
  Tested against `git archive HEAD` + only my two files in a scratch overlay: full
  `actions` package ok (1.9s), `discovery_checks` ok.
- **Mutation checks, both directions:** guard inverted → firing test FAILS. Guard
  deleted → **first attempt SURVIVED the fork control** — sqlmock's recorder never sees
  an unexpected statement, so "no insert recorded" was vacuously true (the
  mock-bookkeeping-cannot-assert-a-negative landmine, met live). Fixed by observing the
  judge's OWN log stream (zap observer; every ported-route utterance contains "ported",
  no pre-existing judge log does — grepped). Re-induced: guard deleted now fails the fork
  control via the logs and nowhere else.
- **Council**: corr `d2edf61d-87af-4195-bcce-c5717afc2d9e`, submitted 2026-08-18 (three
  schema bounces first: plan must be an OBJECT; 'create'→'add'; a sketch starting `###`
  reads as comment-only). Committed pre-verdict with `Council-Submitted:` per the norm.
  Verdict owed: budget ~30 min from submission; find the run by payload
  (`collected_data->'input_data'->>'fix_correlation_id'`), never retry on a missing row.
- Bug file updated with the dated FIXED-IN-CODE section + close conditions.

## 2026-08-18 — council APPROVED ROUND 1 (`d2edf61d`), 13 seats approve / 4 object advisory, none high; all four mediums answered with evidence

- **guidelines (medium)** claimed the work-item dedup contract is "DELETE+INSERT, never
  ON CONFLICT". **The file's own pinned contract says the opposite**:
  `TestConvergenceGuard_EscalatesAfterTwoFailedCycles` asserts "must not DELETE+INSERT —
  no such dedup contract exists" and the acceptance_stuck comment (council-reviewed in
  081's round) states the ON CONFLICT arbiter matching idx_swi_dedup IS the canonical
  idiom (a DELETE+INSERT would disturb a row a human is triaging). The real risk the seat
  gestures at — Go terminal-list vs DB partial-index drift — is the standing lockstep
  contract ([[dedup-index-go-list-lockstep]]), unchanged by this fix. No change.
- **guardian (medium, retraction seams)**: measured — NO retraction or revalidation
  mechanism touches `ported_tool_fix` (only Go refs outside the three producers are a
  comment; `reviewRevalidators` covers other types only). Cross-producer closing is
  impossible today, and the check-segmented keys keep the three producers' items
  disjoint by construction. No change.
- **guardian (low, single-caller)**: census re-run INCLUDING snapshots and deleted rows —
  still exactly one agent_definitions row (`tool-acceptance-agent`, active, not snapshot,
  not deleted). Conclusive.
- **guardian (low, log-text fragility)**: self-protecting — the firing test asserts
  `saidPorted()` TRUE, so a reword that drops "ported" from the arm's log lines fails the
  FIRING test immediately; the negative controls cannot silently rot without it.
- **debug_historian (medium, post-deploy proof)**: already committed to in the bug file's
  close conditions; sharpened there — probe the binary for `routePortedAcceptanceFailure`
  with an invented-symbol control, per service, THEN the natural induction.
- **bug_historian (medium, the still-silent branches)**: disposition recorded — the
  residue is (i) run items lacking `spec.component_id` (only pre-425/bespoke shapes;
  every current producer writes it), (ii) transient lookup errors, (iii) renamed forks
  (a different defect, not a ported instance). Deliberately NOT given a fourth
  handler-less queue: three human-review sinks already exist for ported pages
  (`ported_tool_fix`, `fact_drift_review`, tool-auditor's `needs_human_review`) and the
  033/083 cadence work owns their drain. Named as a residual in the bug file.
- **editquality (low, priority parity)**: a real catch — checked the LIVE rows:
  ported_tool_fix carries severity→priority high→30 / low→80 (dispatch is ASCENDING);
  my medium→60 sits exactly on the sibling mapping. Source differs deliberately
  ('acceptance' vs the checks' 'discovery' — producer-truthful, as acceptance_stuck).

**Trailer status:** the commit carries `Council-Submitted:` (written pre-verdict,
forward-only forbids an amend); 098 credits it automatically now the correlation is
approved. Verdict READ in full this session — this section is the record of that read.

## 2026-08-18 (evening) — FIX LIVE on v1.0.1310; OWNER RULING answers Door 1; induction FIRED

- **v1.0.1310 (pods 18:00Z) carries the fix — verified at BOTH replicas' binaries**:
  `grep -ac routePortedAcceptanceFailure /proc/1/exe` → 2, invented-symbol control → 0,
  on `-l8r76` and `-xdsz6`. Close condition 1 of the bug file: DONE.
- **OWNER RULING 2026-08-18: ported tools should ALL be rewritten and be fully
  customisable by the framework like any other tool.** This decides PLAN §Door 1 as
  none-of-the-above: not fences over ported blobs — native rebuilds, fleet-wide (the
  webdesign_tool_rebuilds approach extended beyond webdesign). Fencing-as-remedy is dead;
  the judge fix remains the safety net for as long as any port is still serving.
- **loancash measured against the owner's "should have been built by the framework"**:
  half right — `tool-affordability-complaint-checker` and `tool-cpa-cancellation-checker`
  ARE framework forks (built 2026-08-12, canonical URLs, fenced, mobile-checkable). But
  THREE genuine ports remain from the 2026-08-01 adoption (`tool-complaint-deadline-calculator`,
  `tool-price-cap-checker`, `tool-true-cost-calculator` — `ported-page` component, flat
  `.html` URLs) plus a ported `tools-index`. Those four fall under the rewrite ruling.
- **Fleet rewrite backlog (clause-b census, any build_status), 72 pages / 6 sites**:
  webdesign 56 (55 deployed + 1 needs_rebuild — down from 59, their lane is moving),
  gamesdesign 6 (ALL needs_rebuild), loancash 3, loanandmortgagecalculator 3 (their lane
  has its own D5 authority programme — coordinate, don't direct), vonc 3 (1 deployed +
  2 needs_rebuild; includes island-backed special cases — per-tool call), mortgagecalculator 1.
  ⚠ flat-URL ports rebuilt natively will meet the canonical-URL convention — the 080
  `page_canonical_collision` detector watches exactly this; expect its items, don't
  suppress them.
- **Induction FIRED 18:4xZ**: hand-inserted the due-sweep's exact ported item shape
  (`cc37a4d7-5019-401b-bb5b-7db5a23c4c2b`, acceptance_run:vibe-equalizer, status triaged —
  the dispatcher-cannot-see-detected landmine). NOTE: `tool_acceptance_run.sh` REFUSES
  ported tools (resolves by `cc.function`, which is 'ported-page') — worth a fix in that
  script if manual ported runs become routine; the hand-insert with the proven 08-14 spec
  shape is the workaround. Expected outcome: mobile-fit FAILS (page still overflowed at
  the live URL 08-17) → the new arm files
  `ported_tool_fix:tool_acceptance_tier4:vibe-equalizer:6b49db8e…`. Monitor armed.

## 2026-08-18 (later) — induction queued behind 131 items; priority bumped on my own item

First monitor window (40 min) expired with `cc37a4d7` still `triaged`, unclaimed. Not a
defect — queue physics: 131 triaged build items sat AHEAD of it (dispatch is ASCENDING
priority, 5/tick; 96 of the 131 are `page_rerender` at priority 80, which other lanes
refill), drain measured at ~1 claim/min, so a priority-90 item could starve for hours.
Bumped MY OWN induction item to priority 20 (behind only 2 items) — a dispatch-order
hint on a row this lane created, not a semantic change; the run itself is identical.
Monitor re-armed. If the SECOND window also expires unclaimed, stop assuming and read
the dispatcher: the ai-endpoint health gate (`claim_work_item_action.go` ~:218,
LANDMINES `ai_endpoint_health`) can hold ALL claims while endpoints read unhealthy.

## 2026-08-18 (final) — INDUCTION PROVEN END-TO-END; 146 (ported-tools slug) CLOSED

Item `cc37a4d7` claimed by build-dispatch-loop after the priority bump, run completed;
the Tier-4 run FAILED mobile-fit on the live page (same 380px/324px attribution as the
07-29 filing and the 08-17 re-scan) and the new arm filed
`ported_tool_fix:tool_acceptance_tier4:vibe-equalizer:6b49db8e…` at `needs_human_review`,
created 19:24:22Z, spec carrying subject/page/component/failing_checks/criteria/screenshot
(s3 URI + presigned)/forced_by/fix hint. The verdict note's `Fix:` line records the
filing ("no automated fixer may act — this is a PORTED instance … filed ported_tool_fix
at needs_human_review carrying mobile-fit for a human to route") — the note-from-outcome
half verified too. **Every close condition met: fixed + council-approved + live on both
replicas + production-induced. File moved to `bugs_closed/` with the closing banner.**
The induced item stays in the queue deliberately — it is a real finding and the rebuild
lane's queue-ordering signal.
