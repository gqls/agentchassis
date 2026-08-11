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

## 2026-08-10 — RFC_021 RULED; all three executed; the code half is LIVE and pod-proven

**Owner ruling** (recorded in RFC_021 §"OWNER RULING"): Q1 = **option C** (census at
adoption + an AUTOMATED check, not per-adoption producer inventories); Q2 = **KEEP** the
enforcement; Q3 = **proceed** with increment #5.

**Q2 proof — the code half is LIVE.** The fleet rolled past this lane's image: pods now run
**v1.0.1280**, and both replicas carry `carries REMOVED config key` and `bugs_open/234`
(negative control 0). So `RemovedConfigKeys` + `create_work_item` StrictConfig are in
production, not merely built.

**Q1 delivered:** `removed-config-keys-check` CronJob (06:25 UTC), `kubectl apply -k`'d and
**proven by a manual run before being trusted**: 181 definitions walked, 0 carriers,
`doc_notes` row written *on the clean result* (a missing row means the job did not run —
that must never look like "nothing is wrong"). Go gains `--removed-keys-in-use`; the Python
mirror's two drift risks are pinned by `removed_keys_cron_parity_test.go`, the declared-list
pin **mutation-proven** (drop the `update_page_status` entry → FAIL).

**Q3 delivered:** migration **370** applied + recorded (verify mutation-proven), seed 025
corrected, `UpdatePageStatusInputSpec` declares `notes_field`/`validation_issues_field`
retired with the unimplemented intent preserved in the messages. Data before code, per the
protocol. Rides **v1.0.1281** (built from HEAD, strings-verified incl. negative control,
pushed; **not** deployed — rolls are the owner's).

**Canary status: NOT PROVEN, and the reason is external.** Three firings produced no
validation event. The first poller *reported* a rejection but its evidence had been lost to
a pod-deletion race between capture and grep — a marker with no evidence, so it was
discarded and the poller rewritten (v2 greps the same capture it just took). Rounds 2 and 3
produced nothing because **the in-cluster fleet is stopped on an account-level Anthropic
cap** (another lane's finding, commit `5fb7c6ebe`; last orchestration 16:56Z). Nothing was
dispatched, so nothing could be rejected. The canary definition has been
**deactivated** so it cannot pollute the census that the Q1 protocol depends on — re-running
needs the row deleting first (the fire script refuses if the type exists in any state).
The strict flip remains pinned by unit test + the live pod-grep above; the canary would add
a behavioural witness and is owed when the fleet resumes.

**Two wrong calls this session, both caught before the commit that carried them, both in
WRONG_CALLS.md:** a fleet census written INTO a Go comment (false within the hour — 356's
data half landed under me), and a filtered count leaking into its own denominator
(`steps: 0` for the fleet's busiest work-item action; the true figure is 17).

**`commit_from` is now 0 carriers fleet-wide** and is a RemovedConfigKeys candidate —
deliberately NOT declared here: it belongs to the bugfix_136 lane, which was committing on
it the same day. Noted in that lane's deferred-items file with the protocol it now needs.

## 2026-08-11 — BOTH PROOFS LANDED. The lane's technical work is complete.

Fleet resumed (last orchestration 42s before the check) and **v1.0.1284** is deployed
carrying BOTH halves — `incr5=1` (increment #5's message), `bugs_open/234`=1, negative
control 0, on **both** replicas.

### Proof 1 — the STRICT CANARY fired, and the error names the fix

`witness_234_fire.sh` (corr `2ae40cde…`). Chassis log, `processor.go:293`, 09:35:29.553Z:

> `step 'file_witness_row' (action 'create_work_item') has unrecognised config keys
> [zzz_strict_witness_234] — this action declares its config contract as complete, so an
> unknown key is a definition error, not a no-op`

Classified `error_unrecoverable` / `permanent:code:WORKFLOW_INVALID`, **not retried**
("Validation error detected - NOT retrying to prevent infinite loop"), and recorded to
`agent_error_log` via `VALIDATION_ERROR_DROPPED`. No work item filed. So the strict flip
is not merely present in the binary — it **refuses live traffic, permanently, with the
key named**.

> **WHY THE FIRST THREE FIRINGS FOUND NOTHING, and it was never the fleet cap alone.**
> Both pollers hunted for a pod named `strict-witness-234`. **No such pod is ever
> created**: `ValidateWorkflow` runs in the CHASSIS processor (`processor.go:276`) BEFORE
> any agent is spawned, so a rejected workflow leaves no witness pod, no orchestration
> row, and its only trace is a chassis log line. The pollers were looking in a place the
> evidence could never be. The fleet cap masked this — it gave a sufficient-looking reason
> for silence, so I stopped looking. **A plausible external explanation for a null result
> is exactly when to re-examine the instrument.** Logged in WRONG_CALLS.md.

**The unplanned control that makes it airtight:** the spec witness below was fired minutes
later down the *identical* path and **did** produce an orchestration row. Same dispatch,
same lane, same minute — one ran, one was refused. The difference is the bogus key.

### Proof 2 — SPEC DELIVERY, at a FILED ROW

`witness_234_spec_fire.sh` (corr `5468a042…`), carrying improvement-loop's
`insert_rerender_item` config **verbatim** on the spec key:

```
spec_witness_234_eac60db8 | {"refresh_site_components": true} | cancelled
```

**That is the bug fixed, observed at a row** — the thing 17 previous rows could not do.
Disconfirmable by construction: `{}` would have meant migration 364 changed nothing.

Written this way rather than waiting for improvement-loop because it **has filed nothing
since 2026-08-09 14:56Z** — the step is conditional on an audit promoting findings, so
"~1.8 rows/day" (my earlier figure, an average over 8 days) was never a rate you could
wait on. Forcing it by dispatching improvement-loop at a real site would have run audits,
fixes and a rerender on live customer pages to prove a one-key change. The residue this
leaves is honest and small: the witness proves the MECHANISM delivers; that
improvement-loop *reaches* the step is established by reading its live config (identical
`spec_literal`, unchanged apart from this rename). The natural row will still appear.

**Both witnesses DELETED afterwards** (definitions and the witness work item); census
re-checked clean: 17 create_work_item steps, 0 unrecognised keys. A live witness carrying
a bogus key would poison the very census the RFC_021 Q1 protocol depends on.

### Postscript — the instrument blindness, demonstrated on the same event

The original pod-watching poller was still running against corr `2ae40cde…` while I found
the rejection in the chassis log, and it has now finished. Its verdict on the firing that
**demonstrably was refused** (chassis `processor.go:293`, 09:35:29.553Z, same correlation):

```
found=no
--- row must be ABSENT ---   0
--- orchestration record --- (empty)
```

So the two instruments were pointed at the same event and disagreed completely: the
pod-watcher reported nothing, the chassis-log grep reported the refusal 25ms after
dispatch. That converts the WRONG_CALLS entry from a reasoned explanation into a
**demonstration** — the check could not have fired on a healthy fleet either, and the fleet
outage merely supplied a believable reason not to ask.

Note also which fields the blind instrument DID get right: row absent, orchestration
absent. Both were true, both were consistent with a rejection, and **neither could
distinguish "refused" from "never ran"** — which is exactly why row-absence was already
recorded as TRAP 2 in the RUNBOOK before any of this. A watcher can be right about
everything it measures and still be unable to answer the question.

## 2026-08-11 (later) — commit_from retired here, 136 lane told, council round 2 in flight

**`commit_from` taken at the owner's direction** (it had been left with the bugfix_136 lane
as a matter of ownership etiquette; the owner reassigned it). Census at adoption, broadest
form — all depths, any row state, **all four workflow columns**, plus a whole-table text
match: **0 carriers**. Declared in `UpdatePageStatusInputSpec.RemovedConfigKeys` alongside
the 370 pair; cron literal updated in lockstep and the parity guard mutation-proven (drop
the entry → the test fails naming both maps). Audit + a manual CronJob run: 4 keys declared
across 2 actions, 0 carriers.

Worth keeping: `commit_from` survived for months because `coordinator.go`'s `dataRefKeys`
carried `"commit_from", // Used by update_page_status` — **false**; that list names keys
whose VALUE is rewritten under loop expansion, never what an action consumes. Three readers
took it as a statement of consumption. Its retirement message points at the build-provenance
stamp (`bugs_open/153`, BLD-019), because the intent — knowing which commit a build came
from — is now genuinely served there. Better than "deleted, sorry".

**bugfix_136 lane notified** in their own `HANDOFF_2026-08-09_deferred_items.md` (their item
is closed; and their migration 356 is applied but has **no `schema_migrations` row** — it
cost me an hour of "is `commit_from` 6 or 0?", so they should `--record-only` it). No live
session is on that lane — checked the transcripts — so the file is the only channel.

**Council round 2 submitted** under the same correlation `3eb0d1f1…` so the trail
accumulates. The rationale leads with why a resubmission is legitimate at all: this is NOT
"better measurements against a scope veto" (forbidden), it is the change as it stands after
the owner ruled the escalated scope question, plus the mechanism built to answer the
guardian's named gap.

> **The first round-2 attempt was INVALID and no seat saw it.** I put prose in an edit's
> `file` field (`…/{364,370}_*.sql + tests (…)`) to fit 14 files into an 8-edit cap;
> `editProblems` refuses any path containing whitespace. It presented as
> `complete_invalid / COMPLETED / error=NULL` — the `bugs_open/099` shape — and the tell was
> `execution_path=[]` with only the input in `collected_data`: **a run that validated nothing
> has nothing in it.** Overflow belongs in `rationale`/`sketch`, which are free text.
> WRONG_CALLS.md.

**Image v1.0.1285 built + pushed** (not deployed). Verified the *correct* way after the
`strings` recipe was retired under me: extracted the binary, found the STAMP by testing
recent commits with an absent-control, then `git merge-base --is-ancestor` — which is a
query, not an inference. **`git rev-parse HEAD` was NOT the stamp** even seconds after my own
build, because HEAD moved under me; that trap and the chassis's rotated-away provenance line
are both now in LANDMINES and the RUNBOOK.

## 2026-08-11 (final) — APPROVED, and the last unproven claim closed on v1.0.1286

**Council `3eb0d1f1`: APPROVED at round 4** — 13 approve, 2 advisory (both "I cannot verify
this from here"), none high. Trail: REJECTED → owner ruling (RFC_021) → REVISE → REVISE →
APPROVED. Both REVISE rounds found real defects; details in the case file and in
[[a-revise-round-is-cheaper-than-the-defect-it-finds]]. `098` coverage report: this lane's
commits are **REVIEWED by correlation, 0 MISMATCH**.

**Deploy verified the correct way (v1.0.1286, deployed by the owner):** extracted the binary,
found the stamp by testing recent commits (`c3b424c8e`, another session's), negative control
clean, then `git merge-base --is-ancestor` for each of my three commits — all three IN. Note
the stamp is not mine and not HEAD: that is normal on this tree and is exactly why ancestry,
not equality, is the question.

**The last unproven claim, now closed.** `commit_from`'s retirement had only ever been
asserted. Witness `cf-witness-234` (an `update_page_status` step carrying
`commit_from: page_deployed.commit_sha`, the real historical value) was **REFUSED by the live
validator**: `step 'mark' (action 'update_page_status') carries REMOVED config key(s): …`,
and **no orchestration row** — rejected pre-spawn, as the landmine says. So the mechanism is
proven on a SECOND action, not just on `create_work_item`. Witness deleted.

> Grep note: `grep -oE '"error":"[^"]*"'` truncates this message, because the message itself
> contains a `%q`-quoted key name. The truncation is the instrument, not the error — the same
> class as everything else in this lane's WRONG_CALLS entries.

**Four wrong calls this session, all logged with the check that would have caught each:** a
census written into a Go comment (false within the hour); a filtered count inside its own
denominator; prose in a schema-validated path field (invalidated a whole council round before
any seat saw it); and citing the wrong provenance mechanism for `commit_from` (caught by a
seat, and the deepest of the four).
