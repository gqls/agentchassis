# NOTES — bugfix 440 (append-only, newest at the bottom)

## 2026-09-02 — lane opened; evidence gathered; one wrong call caught

Spun out of 410's candidate 1 (owner decision). Evidence base assembled first-hand — census,
probe, migration reads — all `[MEASURED 2026-09-02]`, recorded in the bug file. Missteps:

- **Wrong call, caught before filing but after saying it in-session**: read "2 rows contain the
  warning string" as "warning fired twice in production". Both rows were the 404 lane's council
  runs QUOTING the string in their submission payload. Caught by reading one member row
  (`current_step = complete_revise/complete_approved`). Logged in WRONG_CALLS 2026-09-02. The
  corrected finding (zero production firings + prose minted via migrations that bypass the
  creator) became the load-bearing "many doors" argument — the error, chased, was worth more
  than the number.
- Side effect of the same read: learned 404's r4 is `complete_approved` — their design is
  through. Their session has not yet read/recorded the verdict; nothing of theirs touched here.

## 2026-09-02 (later) — phase 1a built, mutation-proved, submitted (r1 corr 55def842)

- `rerender_routing_key{,_test}.go`: three-state resolver + two paste-target clause renderers.
  All four tests green; both named mutations run and killed by exactly their named test
  (unknown→assemble mutation → UnknownRefuses red; absent→refuse mutation → AbsentAssembles
  red). Refusal proven with REAL census unknowns (`tool_retirement`,
  `light_palette_chrome_replaced`), not synthetic values alone.
- ⚠ `platform/livespec` package run FAILS regardless of this change:
  `TestNoNewMigrationFileReadersOutsideTheAllowList`, the 405 lane's committed breakage
  (`ffa1707b3`, 7 days, documented in 404's NOTES). Reproduced identically without my files.
  Not touched — their file, their lane.
- 097 gotchas hit, for the next submitter: `operation` is an enum (learned 08-31, remembered);
  NEW one — a sketch whose every line starts `#`/`//`/`--` is refused as COMMENT-ONLY, which a
  markdown `###` heading triggers. Sketch the content lines, not the heading.
- REB-008 registered in the same commit (status: BUILT AND INERT BY DESIGN, with the
  do-not-overread warning).

## 2026-09-02 (later still) — phase 1a verdict READ: APPROVED, 2 advisories none high, 4 abstained

Corr `55def842`. Dispositions, so nothing dangles:

- **editquality [medium] "nothing calls the resolver; the defect is unchanged by this commit"** —
  TRUE AND BY DESIGN, conceded in the submission's own first paragraph ("INERT foundation only").
  This round never claimed to change production behaviour; the flip is RFC_062 phase 3. Their
  MISSING items (gate untouched, creator untouched) are the stated phase plan, not omissions.
- **reuse_agent [medium] + prior_art_librarian [medium], the real ask: the three claimed reuse
  targets were not SHOWN to exist** — fair, and answered at the artefact `[MEASURED 2026-09-02]`:
  `platform/livespec/rerender_reasons.go` — `var RerenderSectionReasons` (symbol area ~:129),
  `func CheckRerenderModeConditionClause()` (~:157), `func RerenderSectionReasonByName` (~:168);
  and structurally, `rerender_routing_key.go` CALLS the first and third in the same package and
  the package compiles with all four tests green — a compile is an existence proof grep cannot
  fake. The gate-config read prior_art_librarian wanted attached is the RUNBOOK's own query; its
  output (single conditional, five-value `==` disjunction, then `rerender_sections` / else
  `render_page`) was captured in this session before RFC_062 was written and is quoted in the RFC.
  Lesson for r-next submissions: attach the grep/query OUTPUT, not the claim that it was run —
  the same submission-accuracy failure class 404's r3/r4 just paid for.
- **bug_historian [low]** (missing-vs-empty deferred): stated as a BLOCKING caveat in the file
  header, RFC_062's open list, and the submission risks; phase 3 cannot pass its own round
  without discharging it. Deliberate, not overlooked.
- **guardian [low]** (name that livespec is shared cross-pipeline; future wiring out of this
  round's scope): correct, and now explicit here — any wiring into `agent_definitions` workflow
  steps is phase 3, RFC_062, its own round.
- **architecture [low]** (register must forbid a second producer of `RoutingReasonSpecKey`
  pre-RFC): ACTIONED — REB-008 amended in place, this commit.
- **architecture MISSING ("is this the Nth deflection of the gate to a higher layer?")** —
  answered: it is the FIRST raise, not a deflection. The deferral was stated ONCE (the 404
  lane's livespec header, through their approved rounds) and this lane's response to it was to
  CREATE the RFC with an owner and build the foundation the same day. A deflection leaves the
  higher layer unwritten; RFC_062 exists.

Commit `a3758c399` carries `Council-Submitted:` and is auto-credited now the verdict is approved.

## 2026-09-02 (post-roll) — phase 1a SHIPPED (by ancestry), and the probe that said otherwise became a LANDMINES entry

Fresh chassis roll (ReplicaSet `8ddbf8958`). Verification `[MEASURED 2026-09-02]`:

- **Three-way probe on both pods: PHASE1A-ABSENT, controls clean** — and that reading is WRONG,
  by construction. Phase 1a has zero callers, so the Go linker's dead-code elimination strips
  the uncalled functions and their folded literals. Proven, not asserted: a local
  `go build ./cmd/agent-chassis` from a tree CONTAINING the module also greps 0 for
  `input_data.spec.routing_reason` while both called-code controls grep 1.
- **The honest check for inert code is ANCESTRY**: both pods stamped `0d2feee2ff61`
  (`service_binary_capabilities`), and `git merge-base --is-ancestor a3758c399 0d2feee2ff61`
  holds. **Phase 1a is in the running binaries' source.** The probe will spontaneously start
  reading PRESENT the day phase 1b adds the first caller — do NOT read that as "shipped in that
  roll" (mis-dating trap). Full entry: LANDMINES "A capability probe for INERT code reads
  ABSENT with clean controls".
- ⚠ **DEBT: the LANDMINES entry is appended to the FILE but `landmines-verify-dispatch.sh` is
  BLOCKED** — kubeconfig token expired mid-session (fleet-wide `Unauthorized`, the 3-day
  expiry; owner refreshes). First session after refresh: run
  `./scripts/landmines-verify-dispatch.sh` (NOT `landmines-sync.py --apply` alone — the
  documented consume-the-new-entry-status trap).
- 404 lane: no commits since our CONTRIB — their r4 verdict (approved) still unread by them, so
  **phase 1b stays gated**.

## 2026-09-03 — token refreshed, debts cleared, build verified; a third lane is in phase 1b's file

- **Landmines debt CLEARED, and the trap it predicted fired exactly as documented**: the token
  was back; `landmines-verify-dispatch.sh` reported "0 new … nothing needs verification" —
  another session's sync had consumed our entry's new-status overnight, so the verifier skipped
  it. Remedy applied: `trigger-landmine-verifier.sh 'LANDMINES.md#a-capability-probe-for-inert-
  code-reads-absent-with-clean-controls-the-linker-s-'` → dispatched, correlation `dd777a91`.
- **Fresh build verified by ANCESTRY** `[MEASURED 2026-09-03]`: both pods of ReplicaSet
  `75b987cbd7` stamped `7bf1ff674021` (built 09:24 BST), which carries `a3758c399`. No literal
  probe attempted — per our own landmine.
- **Phase 1b's file gained a third lane**: commit `8eca969cb` (315-reopen) edited
  `create_rerender_items_action.go` (producer stops filing guaranteed-skip rerenders). No
  conflict with our planned stamping edit on inspection of the commit message, but 1b must
  re-read the file fresh at write time and expect same-file-passenger risk from BOTH the 404
  and 315 lanes.
- 404 lane docs: still no verdict-recording commit. Phase 1b stays gated (unless the owner
  lifts the courtesy gate — put to them as decision D5 today).

## 2026-09-03 (later) — OWNER RULED D1–D5; phase 1b BUILT and submitted (corr `934327db`)

Rulings as recommended: **D1** refusal routes to `needs_human_review` · **D2** 404 lane co-signs
the gate migration · **D3** the write-door CHECK is IN scope · **D4** `spec.reason` is never
validated · **D5** phase 1b's courtesy gate LIFTED. Recorded in RFC_062 §Rulings (status flipped
from DRAFT to DESIGN RULED), the lane PLAN's phase table, and REB-008.

**Phase 1b, and the one decision inside it that mattered.** The naive implementation stamps
`routing_reason` whenever the reason is KNOWN. That would have been a behaviour change wearing a
foundation's clothes: `image_landed` WITHOUT a `component_id` deliberately stamps nothing
(REB-001's designed degrade to assemble), so a routing key there would make the phase-3 gate
route a page that assembles today — and it would have surfaced weeks later, at the flip, as
"the flip broke image_landed". The shipped form sets `RoutingKey` **in lockstep with
`KeyReason`**, inside the existing `StampReason` guard, which makes the flip provably
byte-neutral for every reason that works correctly now. Found by reading the vocabulary's own
comment (`RerenderSectionReasons`, the StampAlways doc) rather than by testing after the fact.

Invariants pinned, each mutation-proved:
- lockstep across the whole (reason × component_id) matrix — mutation A (hoist the assignment
  out of the guard) → red;
- an unknown reason yields NO routing key, by control flow (the `!known` branch returns before
  the assignment), so phase 3 can never refuse an item this action minted — mutation B (assign
  in the `!known` branch) → red, two tests;
- producer/consumer round trip against `livespec.ResolveRoutingReason` (the drift this
  vocabulary already suffered once);
- the dedup key is untouched (`pageRerenderItemKey` takes `keyReason` alone) — the one way an
  additive field could have changed live behaviour.

Green in isolation against HEAD `e663a1d06` (`verify-head-builds --with`, both files). Submitted
with the 404-r4 discipline applied to my own submission: a scripted check that every sketch
contains what its rationale claims — 5 PASS, before dispatch, not after a REVISE.

⚠ Shared file: three lanes in eight days (404 `ef4236b4d`, 384 `9a00a1ee9`, 315-reopen
`8eca969cb` TODAY). Read fresh, diff confined to three hunks (32+/1-), pathspec commit.

## 2026-09-03 (later still) — phase 1b verdict READ: APPROVED, 2 advisories; one objection RESIZED THE LANE

Corr `934327db`, `decided_by: approved with 2 advisory objection(s) — none high-severity`,
4 abstained, 12 seats. Dispositions:

**`bug_historian` [medium] — other producers of `page_rerender` items. CONFIRMED, MATERIAL, and
the most valuable thing this round produced.** My claim ("first and only producer") was true of
the ROUTING KEY and useless as reassurance. Census `[MEASURED 2026-09-03]`: the creator I fixed
mints **1 of 3,172** reason-bearing items; `completeness-discovery-agent` alone mints 1,882, and
**13 Go files** write in-vocabulary reasons directly, mostly as raw `{"reason":"x"}` literals
bypassing the vocabulary constants. Full table in the bug file; RFC_062's design gained a fourth
door (producer conversion) and phase 3 gained a gate condition. **The transition clause is now
load-bearing** — narrowing early would route ~3,100 items to assemble, this bug's own shape
inside its own fix.

**`reuse_agent` [medium] — does `site_work_items.pipeline` already carry this distinction?
CHECKED, and no** `[MEASURED 2026-09-03]`: `pipeline` holds `build` (12,106) and `content` (13)
— a coarse dispatch lane, orthogonal to render mode; the `build` lane alone contains **13
distinct reasons**. It cannot express "which render mode", and overloading it would put routing
in the queue-selection column. Right question, and the query is cheap enough that it should have
been in the submission.

**`guardian` [medium] — "livespec changes are not in the edit list; this will not compile."**
Answered: `RoutingReasonSpecKey` and `ResolveRoutingReason` shipped in **phase 1a**
(`a3758c399`, APPROVED `55def842`), live in build `7bf1ff674021` by ancestry. Same answer to
`prior_art_librarian`'s MISSING (the resolver exists). Both would have been unnecessary had the
submission named the prior round's commit in the edit rationale rather than only in `submitter`.

**`guidelines` [medium] — register file not in the edits array.** CONCEDED, my accuracy failure:
REB-008 *was* updated in the same commit, but an unlisted edit is invisible to a reviewer. Third
round running that a seat has caught a submission-accuracy gap rather than a design fault — the
pattern is now explicit in the handoff.

**`editquality` [low] ×2 — ACTIONED in code, this commit:** the unknown-branch test now asserts
the early-return path itself (`Scoped`/`StampReason`/`KeyReason` all empty), and asserts as a
PRECONDITION that each poison value is still out-of-vocabulary, failing loudly with "pick
another" if a later lane declares one — the 410 poison-row trap, pre-empted.

**`tooling_provenance` [low] — leave a durable record for the next lane.** Taken: RFC_062 §Rulings
+ REB-008 + this NOTES entry are that record; no hand-written `doc_notes` row (estate rule: the
file is the system of record, the sync writes the rows).

⚠ **Mutation re-proof against the HARDENED tests was interrupted twice by another session's
in-flight test files** (`undefined: censusExcludedOwnedPages`, `readRebuildPolicy` — files this
lane never touched). First background attempt gated on `go build`, which passes while TEST files
are broken, so its three FAILs were void, not evidence — re-run gated on `go vet`. Recorded
because "the gate I chose could not see the failure I cared about" is this estate's most-repeated
error, and I made it again in the checking apparatus rather than in the claim.

**RESOLVED, same session** `[MEASURED 2026-09-03]`. The `go vet` re-run named above was ALSO the
wrong gate — it fails on any lint finding anywhere in the package (`unreachable code` in a third
lane's file), so it could never open. Correct gate: `go test ./pkg -run ZzzNoSuchTestZzz`, which
builds the test binary and runs nothing (exit 0 iff the test files compile). With it: baseline
**ok** → mutation B **`--- FAIL: TestUnknownReasonProducesNoRoutingKey`** → reverted **ok**, 0
mutation lines remaining. The hardened tests are mutation-proved, not merely passing. Both wrong
gates are logged in WRONG_CALLS 2026-09-03 — one too weak to see the failure, one too strong to
see the success.

## 2026-09-03 (evening) — phase 2 part 1: the Go producers, converted through ONE helper (corr `c7dab2c1`)

**Enumeration was the work, and it corrected my own figure from this morning.** "13 Go files
write an in-vocabulary reason" was true and was NOT the conversion set: reading each site's ITEM
TYPE found **five** `page_rerender` producers. The other four file `needs_page` (×2),
`needs_rerender`, and `literal_markdown` — the last matching 404's own finding. A blanket sweep
would have stamped page-rerender routing decisions onto items no rerender gate reads. Corrected
visibly in the bug file and RFC_062 rather than quietly.

**The helper, and why it is not a key added at each site.** `livespec.RerenderReasonFields` /
`StampRerenderReason` / `RerenderReasonJSONPrefix` define the pair once, beside the vocabulary
they depend on. Adding the key by hand at five sites would restart, one level along, the exact
drift this lane closes — the vocabulary's own header records the last time a judgement was
copied to N sites (gate knew five values, Go knew three). The JSON form is DERIVED from the map
form by marshalling, and a test composes the two and requires agreement.

**A trap found and designed out mid-build.** The first cut returned a comma-free fragment for
templates written `{%s,"page_name":%q}`. That works only while every caller passes a compile-time
constant: the day someone passes an empty variable it emits `{,"page_name":…}` — invalid JSON,
written into a text column, discovered whenever something next parses that spec. Changed to a
TRAILING-COMMA prefix with templates `{%s"page_name":%q}`, so both states compose validly.
Proven by EXECUTION, not reading — a scratch program reproducing each converted template printed
all three cases, including the empty one (`{"page_name":"about"}`).

**One conversion deliberately NOT made, and it is the shared-tree rule doing its job.**
`refresh_evidence_base_action.go` is the fifth producer; another session has **245 uncommitted
lines** in it right now (a FactsUnverifiable/attestations feature, two of its own tests currently
red — which is also what those two unexplained failures in my full-package run were). I wrote the
two-line conversion, saw the diff, and **reverted it** rather than sweep their half-finished work
into my commit. Recorded as the one deferred conversion; phase 3's census catches it if forgotten.

**Isolation proven with a control**: `verify-head-builds --with` my six files → actions and
discovery_checks green; `platform/livespec` FAILs identically **with and without** my files
(405 lane's `TestNoNewMigrationFileReadersOutsideTheAllowList`), so the RED is theirs — the
control is what makes that a finding rather than an assumption.

Remaining in phase 2: the raw-SQL migration door (authoring rule + `pattern-check.py` advisory),
its own round — that file is in council scope (2026-08-24).

## 2026-09-03 (night) — phase 2 verdict READ: APPROVED, 3 advisories; two of them found real things

Corr `c7dab2c1`, `approved with 3 advisory objection(s) — none high-severity`, 4 abstained,
13 seats. Dispositions:

- **`guardian` [medium] — "item type asserted from source, not live data" (citing the landmine
  *a discovery check's NAME is not the item_type it FILES*). RIGHT, and it found a real gap in
  my evidence.** Live census `[MEASURED 2026-09-03]`: `misdirected_cta` → `page_rerender` (2,315
  items) ✓ as claimed; **`contact_form_undeliverable` → BOTH `page_rerender` (1) and
  `contact_form_undeliverable` (7)** — that file files TWO item types and my `grep -n ItemType:`
  had stopped at the first. The stamp is correct (the two blocks build SEPARATE specs — verified
  after) but by layout, not by the enumeration I described. ⚠ And acting on the objection I
  first "fixed" a shared-spec problem that did not exist and broke the build; re-reading showed
  the specs were always separate. **An objection being right about the gap in your evidence does
  not make it right about the defect.** Both halves in WRONG_CALLS 2026-09-03. The site now
  carries the live-verified two-item-type fact as a comment, with the instruction for whoever
  ever merges those specs.
- **`bug_historian` [medium] — an out-of-vocabulary reason produces no routing key AND NO
  SIGNAL; silence, not a loud failure. ACTIONED in code.** `StampRerenderReason` now returns
  `known bool`, restoring the signal its own sibling deliberately exposes
  (`RerenderSectionReasonByName`: *"the bool is the whole point"*). Constant-passing callers may
  ignore it; variable-passing callers must not, and the doc says so. Mutation-proved (return a
  constant `true` → the new test goes red).
- **`editquality` [medium] — the sketch showed `RerenderReasonFields` and
  `RerenderReasonJSONPrefix` but not `StampRerenderReason`'s body, though edits 3–4 call it.**
  CONCEDED. Fourth consecutive round where a seat catches submission ACCURACY rather than a code
  fault. The pre-dispatch self-check I added at phase 1b checks that sketches contain what their
  rationale CLAIMS; it cannot see a symbol named in `symbol` whose body is missing. Next round's
  check: every symbol listed in `symbol` must appear in the sketch.
- **`prior_art_librarian` [medium] — an unfetched landmine whose title names these exact reason
  values.** FETCHED and read (`LANDMINES.md`, the `create_rerender_items` entry): it warns that
  writing a resolved value into `spec` can flip a site-wide rerender into a component-scoped one,
  because `rerender-pages` reads `input_data.spec.component_id`. Not this change — that trap is
  about `component_id`; `routing_reason` is read by nothing until phase 3. ⚠ But the entry
  quotes `create_rerender_items_action.go:219` and a hardcoded
  `(reason == "section_data_resolved" || reason == "image_landed")` gate that **no longer
  exists** — the logic moved into `rerenderModeFor` via the vocabulary. Flagged for a correction
  pass; not silently edited, because it is another lane's entry.
- **`guidelines` MISSING (nested-field ruling) + `guardian` [low]**: REB-008 names `routing_reason`
  as a field riding inside `spec`, and `[MEASURED 2026-09-03]` **zero** live `agent_definitions`
  mention `routing_reason` — no config reads it yet, as designed.
- **`debug_historian` / `architecture` [low] — string-spliced JSON is fragile as a general
  pattern.** Agreed and bounded: two known call-site shapes, the empty case proven to compose,
  and the doc says a third shape should reconsider a structured builder rather than extend this.

**Architecture signal on the signature change, judged and recorded rather than waved through.**
The pre-commit hook fired "exported symbol removed/changed" on `StampRerenderReason` gaining its
`known bool` return. Assessed as a POINT FIX, not a shared-contract change, on three grounds
`[MEASURED 2026-09-03]`: the symbol is **hours old** (introduced in this same lane's phase-2
commit), it has **2 non-test callers, both converted by this lane**, and adding a return value
breaks no caller (Go permits ignoring it) and changes no behaviour for any existing one. RFC_022's
test is about new authority on a shared seam; this removes silence from a seam nobody else uses
yet. If a third caller outside this lane appears before RFC_062 lands, that judgement expires —
and REB-008's no-second-producer constraint is what should stop that happening anyway.

## 2026-09-03 (night) — phase 2 part 2: the raw-SQL door (corr `3b484a74`)

**The design problem was not detection, it was the vocabulary.** A Python checker needs the five
reason values; hardcoding them would be the THIRD copy of a list whose second copy is why
`bugs_open/404` exists. So `check_rerender_routing_key` READS them from
`platform/livespec/rerender_reasons.go` at run time (`_VOCAB_CONST_RE` over the `Reason*`
constants) — verified by execution to return exactly livespec's five — and when it cannot read
them it emits a LOUD finding and checks nothing. That branch is the important one: **a
vocabulary check running on an empty vocabulary passes every file and reads exactly like a clean
bill of health.** Both blind modes mutation-proved: source missing → "could not be read"; source
present but reshaped (pointed at `rerender_routing_key.go`, which has no `Reason*` constants) →
"the declaration shape changed". Neither passes silently.

**Behaviour proven across five fixtures**, including the two silences that matter: free prose is
NOT flagged (owner ruling D4 — the annotation stays free forever), and an in-vocabulary reason on
a `needs_page` item is ignored (the item-type precision phase 2 part 1 established).

**Noise measured before shipping, not after**: 11 findings across 844 lintable migrations — and
the check only sees files a commit TOUCHES, so the practical rate is lower. That blind spot is
stated in the entry rather than left implicit.

**A live finding fell out of the census**: `683_content_listing_rerender_after_roll_HOLD.sql` and
`701_retype_357_population_by_adoption_HOLD.sql` carry `reason` with no routing key and
**no `-- APPLIED` line** — unapplied. Applied after phase 3 they mint items that assemble. Named
in the LANDMINES entry so whoever applies them sees it first.

**The authoring door is the LANDMINES entry, not a new doc** — the estate's system of record for
prospective traps, synced into `doc_notes` so agents read it too. Appended and dispatched in the
required order (`landmines-verify-dispatch.sh`, not the sync alone).

**The self-check I promised last round earned its keep immediately.** Widened to "every symbol
named in `symbol` must appear in the sketch", it FAILED on `_ROUTING_SHAPED_RE` — the exact
class of gap (`StampRerenderReason` named, body never shown) that drew `editquality`'s medium at
round `c7dab2c1`. Fixed before dispatch instead of after a REVISE.

## 2026-09-03 (late) — phase 3's blocker DISCHARGED, and it found a defect in already-approved code

The blocker written into the module header, RFC_062's design and two approved submissions —
*"confirm the evaluator's behaviour on a MISSING key vs `''` before pasting; getting it wrong
inverts the guard for legacy items"* — was discharged by **executing the evaluator**, not reading
it. The assumption behind the shipped renderer was **FALSE**.

`compareValues` (`conditional_branch_action.go`) handles nil BEFORE it strips quotes, so the
quoted two-character `''` is compared against a raw nil and never matches. Measured
`[MEASURED 2026-09-03]`, through the real `resolveFieldValue` + `compareValues` pair:

| item state | `== null` | `== ''` | `== 'cta_links_stale'` |
|---|---|---|---|
| routing key ABSENT | **TRUE** | false | false |
| present, `""` | false | **TRUE** | false |
| present, known | false | false | **TRUE** |
| present, unknown | false | false | false ← the only refusing state |

**Consequence for phase 1a's `CheckRoutingKnownConditionClause()`, which is APPROVED and shipped:**
it emitted `== ''` and the five values, and no `== null`. Pasted into the phase-3 gate migration
it would have evaluated FALSE for **every `page_rerender` item minted before phase 2** — the
fleet's entire normal traffic — sending all of them to the refusal branch on flip day. Fixed
today; the header's warning is kept as a visible correction rather than deleted, because the
next person to test emptiness in a workflow condition will make the same assumption.

Pinned so nobody re-derives it: `rerender_routing_gate_clause_test.go` runs the RENDERED clause
through the real evaluator for all four states, plus the transition clause under both keys. The
`== null` disjunct is mutation-proved (delete it → "absent routing key … clause allowed=false,
want true"). The general trap is a LANDMINES entry (synced, `doc_notes` row confirmed).

**The estate lesson, and it is the whole argument for stated blockers:** this defect was
disclosed to the council as a risk, in two rounds, and approved both times — correctly, because
no seat can catch it from the submission text. What caught it was the blocker being *written
down* and then *discharged by execution* rather than by re-reading the code. A stated blocker is
only worth the run that discharges it.
