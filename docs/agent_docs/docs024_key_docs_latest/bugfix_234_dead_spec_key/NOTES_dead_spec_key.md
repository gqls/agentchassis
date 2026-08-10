# NOTES — bugfix 234 (append-only, newest at the bottom)

## 2026-08-10 — lane opened; premises re-verified; two owner decisions

Picked up `bugs_open/234` (filed 2026-08-09 by the bugfix_136 lane, which ended with "next
action is yours"; `who-owns.py` shows no owning workstream; no live transcript is working
it beyond citations).

**Bug re-verified live before planning** (the filing was hours old, but this tree changes
under you):
- All three carriers still present, correct values, none carries
  `spec_data`/`spec_paths`/`spec_literal` (recursive all-depths walk — see RUNBOOK).
- 16/16 `improvement_rerender_*` rows `spec='{}'` (2026-08-01 → 2026-08-09), positive
  control 5,040 non-empty rows fleet-wide.
- `dedupe_rerender%` and `capability_gap_audit%`: still 0 rows each — the other two
  carriers remain unexercised, so translating them changes no live behaviour.

**Two premises of the case file's owner-decision framing are stale, and both moved the
decision** (recorded in the case file as a dated correction):
1. The case file deferred restoring the flag pending "whatever guard 226 lands". **226's
   guard is LIVE**: both chassis replicas (v1.0.1274 at check time) carry
   `emitChromeDivergenceItem` (strings-proven, negative control clean), and 226's header
   says it has already refereed a real event (the dartsonline header fixer-vs-rebuild loop).
2. "Turning on full site-component reassembly that has not run from this path in months"
   is true of THIS PATH only. Fleet-wide the flag is routine: 8 producers file
   `refresh_site_components: true`, ~5–15 rows/day, latest same-day. Restoring the
   improvement-loop path adds ~1.8 rows/day to a daily behaviour, with the divergence
   guard watching.

**Owner decisions (AskUserQuestion at plan time):** RESTORE the flag via `spec_literal`;
ship BOTH `StrictConfig: true` on `create_work_item` AND the new `RemovedConfigKeys`
opt-in field.

**StrictConfig precondition measured**: after translating the three carriers,
`create_work_item` has ZERO unknown keys fleet-wide **at all depths** (recursive walk over
every live definition; every remaining key ∈ Required ∪ Optional ∪ ConfigKeys ∪
DeprecatedConfigKeys ∪ framework set). This is the "recognised set checked against every
live step" the action's own doc comment names as the gate for strict.

**Misstep (caught in-session, cost ~1 minute):** first read of the recursive census said
2 `spec` carriers, not 3 — I counted the GROUP BY rows instead of reading the count
column (improvement-loop's two steps collapse into one group row). Re-ran showing the
rows; 3 confirmed. The check that prevents it is in the RUNBOOK. Not a WRONG_CALLS entry:
never written down as a claim, refuted by my own next query.

**Prior-art checks for the framework half** (a quiet git log is not silence — searched
the tree, not just history): no `RemovedConfigKeys`/`RetiredConfigKeys` anywhere in Go;
migration 356 (another lane, same family: dead `commit_from`/`output_format` keys) shipped
CheckConfig opt-ins only, and explicitly left its two undeletable dead keys "to be
REPORTED by the detector" — i.e. that lane also had no mechanism for a key that must
hard-fail. Both lanes' leftovers become RemovedConfigKeys candidates once the field exists.

**Fable note:** the owner asked for fable to prepare the plan; the fable Plan agent died on
the session usage limit (reset 19:40). Plan prepared by the main (Opus) session instead,
owner approved it in plan mode.

## 2026-08-10 — Phase 1 DONE: migration 364 APPLIED + recorded; seeds corrected

- Numbering: 363 was free when planned, TAKEN by another session by write time (the 356
  lesson firing in real time). Landed as **364**.
- The migration RENAMES the key in place — `spec_literal`/`spec_paths` take their value
  from the live `spec` key (`jsonb_set(cfg #- old, new, cfg#>old)`), never retyped — so
  the prose value (em-dash included) cannot be drifted by transcription. The pre-guard
  asserts exact expected values (prose asserted by stable ends + key count) and RAISEs on
  drift or partial application.
- **Verify blocks proven disconfirmable BEFORE applying**: three mutated copies, each with
  one UPDATE removed, run against the live DB — each RAISEd `364 VERIFY: a spec key
  survived` naming exactly the skipped step (il/nc/qr flags), transaction aborted, live
  rows confirmed untouched after (3 carriers still present). Then the real apply:
  `BEGIN / DO / UPDATE 1 ×3 / DO / COMMIT`.
- Post-apply: recursive all-depths census = **0 `spec` carriers**; the three new spellings
  read back with exact values. (Definition-level checks only prove the rename; the BUG fix
  is proven at a filed row — Phase 5.)
- Recorded via `run-migrations.sh --record-only` with note. Seeds corrected: 054 (:166),
  291 (:117), 269 (:97 — with a comment on why spec_paths, not spec_literal).

## 2026-08-10 — Phase 2+3+4 DONE: SCR-007 built and mutation-proven, council submitted, v1.0.1278 built+pushed

**Code (commit `d278d7b25`, 10 files):** `ActionInputSpec.RemovedConfigKeys` (retired key →
hard validation error naming the replacement; checked BEFORE strict/unknown; independent of
the `checksConfig()` opt-in; excluded from UNKNOWN so no double-report under the softer
label) + `create_work_item` adopts it for `spec` AND flips `StrictConfig: true` + audit
surfaces (`REMOVED KEYS IN USE`, exit 1; `--specs` emits `removed_config_keys`) + register
SCR-007 same commit. Nine new tests across three files.

**Every guard proven by mutation, each reverted green:**
- `StrictConfig` → false ⇒ `TestCreateWorkItemIsStrict` FAILS.
- `spec` dropped from `RemovedConfigKeys` ⇒ `TestCreateWorkItemSpecKeyIsRemoved` FAILS
  naming it.
- validator branch disabled (`false &&`) ⇒ `TestRemovedConfigKeyFailsValidation` FAILS.
- audit classifier: induced synthetic hit (`LIVE='create_work_item\tspec\ttop'`) ⇒
  classified `removed_keys_in_use` (not unknown), **exit 1**. Against the real fleet:
  `none`, exit 0 — correct, 364 removed the carriers first.

**Ordering held:** migration applied → carriers re-checked 0 at all depths → THEN the code
committed. `git archive HEAD` build + tests green (the committed state is what any
session's roll ships). Live audit run post-change: `UNKNOWN KEYS` has no create_work_item
entry (the two listed actions, `process_data`/`update_page_status`, are the 356 lane's
adjudicated leftovers — the next RemovedConfigKeys candidates, named in SCR-007).

**Council:** corr `3eb0d1f1-6929-4131-bbef-c636256aa667`, submitted before the commit,
`Council-Submitted:` trailer on `d278d7b25`. Run found EXECUTING at `review_architecture`.

**Image:** IMAGE_TAG → v1.0.1278 (commit `0631d6996`; subsumed another session's
uncommitted 1275→1277 bump — same-file passenger, monotonic, noted in the message).
`make build-agent-chassis` from committed HEAD; image strings-verified BEFORE push:
`bugs_open/234` = 1, `carries REMOVED config key` = 1, nonsense control = 0. Pushed.
**Not deployed** — releases are whole-fleet; the roll is the owner's.

**A 17th empty-spec row** appeared pre-migration (`improvement_rerender_finetuning.uk`,
08-09 14:56Z, spec={}) — consistent with the bug, filed before the fix; the damage count
is 17/17, not 16/16, at fix time.

**PENDING (the two proofs that close this lane):**
1. **Filed-row proof (data half):** first `improvement_rerender_*` row created AFTER
   2026-08-10 ~00:30Z must carry `{"refresh_site_components": true}`. RUNBOOK query.
   ~1.8 rows/day natural rate — check within a day.
2. **Post-roll proof (code half):** pod-grep both replicas (`bugs_open/234` ≥1, nonsense
   control 0) + the strict canary (RUNBOOK) once the fleet rolls to ≥1278.

## 2026-08-10 — council round 1: REJECTED (guardian scope veto); RFC_021 filed

Verdict on `3eb0d1f1-6929-4131-bbef-c636256aa667`: **REJECTED, hard veto from
`guardian`** — 7 approve, 3 object (editquality, bug_historian, architecture), 5 abstain.
The veto is about SCOPE, not correctness: live hard-fail on the shared validator
(`ActionInputSpec` field + `checkStepConfigKeys` branch + strict flip on the busiest
work-item action) "dressed as a single-bug fix". Its contained alternative: ship the
offline audit only. `architecture` disagreed with the veto in the same round —
objected-but-approved, asking for a short RFC consolidating the four-state precedence
machine before increment #5. Seats in conflict = the RFC_002/124 shape; per the owner
ruling 2026-07-28 **the code stays, the precedent gets fixed, a human breaks it**.

Actions taken, same session:
- **RFC_021 filed** (`architecture_review/RFC_021_config_key_validation_states_one_contract.md`):
  the written four-state contract (architecture's "missing"), the guardian's veto stated
  fairly, and the two owner questions — adoption protocol for live hard-fail, and
  keep-vs-split for v1.0.1278's enforcement.
- **editquality M answered**: census re-run FRESH at 2026-08-10 11:15Z (post-commit):
  0 `spec` carriers, 0 unrecognised keys on any live create_work_item step, all depths.
  (It had also been re-run at the commit gate — the objection was right that the PLAN
  only cited the older run.)
- **bug_historian M answered**: the mark_page_needs_attention sibling keys are now
  TRACKED in `bugfix_136_config_key_aliases/HANDOFF_2026-08-09_deferred_items.md` as
  RemovedConfigKeys candidates, blocked on RFC_021 Q1.
- **prior_art M answered**: the "no config[\"spec\"] read" claim is grounded in a full
  body read (:247-296 composes from the three real spellings only) and pinned both
  directions by the contract test.
- Verdict recorded on SCR-007 and the case file. **No resubmission** — a scope veto is
  not answered by measurements. Commit `d278d7b25` keeps its `Council-Submitted:`
  trailer; 098 will bucket it against the rejection, which is honest.

**Standing constraint until RFC_021 Q1 is ruled: no new RemovedConfigKeys adoptions, no
new StrictConfig flips.**
