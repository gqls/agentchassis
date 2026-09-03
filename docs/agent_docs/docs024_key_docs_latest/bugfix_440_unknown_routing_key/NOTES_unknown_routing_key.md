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
