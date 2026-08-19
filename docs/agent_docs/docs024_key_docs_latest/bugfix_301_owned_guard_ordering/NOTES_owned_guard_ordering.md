# NOTES — bugfix 301 (owned-guard ordering) — append-only, newest at the bottom

## 2026-08-19 — session 1: validity, ownership, design

**Ownership check** (`scripts/who-owns.py 301`, ~11:50 UTC): verdict "OWNED or recently
active", but reading the two named lanes' handoffs settles it — `bugfix_277_required_fields_repair`
shipped Tier 1 (status change, live on v1.0.1314) and explicitly does not claim the ordering
fix; `vigilant_designer_offer_analysis` (the filer) offers it: "Take `bugs_open/301`" in
HANDOFF_2026-08-18. Taking it. `needs_diagnosis` queue: 0 rows (checked before filing 090).

**Bug still valid** [MEASURED 2026-08-19 ~12:00 UTC]:
- 18 `page-build-handler`/`page-content-writer` orchestrations since the roll (07:52Z);
  the 10:18 cluster shows the shape exactly: handler created 10:18:27 ends `complete_error`,
  its writer child 10:18:42 COMPLETED, resolver child 10:19:16 COMPLETED — full chain run,
  then refused.
- 146 open findings on owned pages routed at `handler_agent='page-build-handler'`
  (live table only — archive not needed for "queued now"; statuses
  detected/needs_human_review/unresolved/failed).
- Tier 1 IS behaviourally live now: 2 `wont_fix` rows since roll with
  `result ? 'owned_page_refusal'`, 1 new `owned_page_review`. (The 277 lane's handoff
  asked for this check with both controls — I have only run the positive half; noted for
  them rather than claimed as their verification.)

**Design facts established first-hand** (file:line in PLAN):
- Live workflow graph read from `agent_definitions` (dump in scratchpad, steps listed in
  PLAN): every write path ends at `save_sections`; no-op arms park without touching the page.
- `save_page_sections_action.go` guard: unconditional refuse on owned, error LEADS with
  `ownedPageSkipReasonPrefix`; emit added by 295.
- Tier 1 mechanism: `routeToErrorStep` copies the error verbatim to `__step_error.message`
  (`coordinator.go:3697`); `UpdateWorkItemStatusAction` matches the prefix and stamps the
  configured `owned_page_refusal_status` (`v3_site_actions.go:5527-5545`). An early refusal
  that errors with the same prefix through `error_step → mark_item_failed` inherits all of it.
- Carrier census: `load_page_record` is carried by exactly 2 live agents:
  `page-build-handler`, `tool-recreation-handler`. The latter is the tool pipeline —
  must never refuse owned pages — which forces the opt-in default-OFF shape.
- Order tolerance: `LoadPageRecordInputSpec` has `CheckConfig: true`, no `StrictConfig` →
  `platform/validation/workflow.go:185-195` warns once on an unknown key, never rejects.
  So config-first is safe; binary-first is dormant.
- RFC_022 budget: `cmd/config-key-audit/optionalbudget.go:14-21` counts `Optional` only,
  deliberately not `ConfigKeys`. New key goes in `ConfigKeys` (it is a setting); trade-off
  disclosed in the council submission. Budget run 2026-08-19: 0 shared actions over N=10;
  `load_page_record` not in the widest list (3 optional keys).

**090 filed** (~12:10 UTC): intake `7281193f-59c2-489a-a9f2-fd4d58408cf5`, run
`dd61df1b-0d93-46e6-9065-1e0b9623379a`. Question put to it: does any branch of
`page-build-handler` write page content other than through `save_page_sections`?
Budget ~30 min per CLAUDE.md; do not retry on a missing orchestration row.

**Missteps this session:** none yet worth a WRONG_CALLS row. (Watching for: quoting the
bug file's 39/74 figures as live — both are dated aggregates; the 74 is policy-join
contaminated per the 277 lane's §8 correction.)

## 2026-08-19 — session 1 (continued): built, tested, submitted

- Go change: `refuse_owned_page` in `LoadPageRecordInputSpec.ConfigKeys` +
  `refuseOwnedPageIfConfigured` on BOTH return paths (authoritative-id and name/id);
  marker-emitters comment added to `owned_page_guard.go`. Tests: 4 cases, incl. the
  key-absent control (mock proves the default path never reads `rebuild_policy`).
- **Proven against HEAD, not the dirty tree**: another session has dirty files in the SAME
  package (`check_endpoint_health_action.go`, `conditional_branch_action.go`, two test
  files), so the in-tree green proved nothing — re-ran `git archive HEAD` + only my 3 files:
  build + full `actions` suite green.
- Migration `488` + `_ROLLBACK`: dry-run apply+rollback in ONE rolled-back transaction
  (both NOTICEs fired); double-apply probe ABORTED on the already-set guard ("could the
  control have fired" — it did). 486 collides already; 488 was free.
- Parity test run (RFC_022 obligation): **FAILS, pre-existing, NOT mine** —
  `save_page_meta_description` (commit `aeccfc595`, the 320 lane) declares 5 optional keys;
  `deployments/kustomize/services/optional-key-budget-check/base/check.py` literal says 0.
  The daily budget cron is BLIND to that action. My key is in ConfigKeys (not counted), so
  my change cannot fix or worsen it. TO DO this session: regenerate the literal per
  check.py's own comment as a separate narrow commit + re-apply the overlay (CLAUDE.md's
  documented procedure), and tell the 320 lane.
- Council submitted ~12:35Z: **corr `c7bc1b9e-97c8-4f3e-8a4f-b3a7029505ee`** (save this;
  budget ~30 min; find the run by payload not printout). Committing with
  `Council-Submitted:` per the norm.
- Register: **PBP-045** appended (page-build-pipeline.md + index row), additions-only
  verified by `--numstat` (9/0, 1 long row in index).
- Noted in passing, post-roll: migration `473` is now APPLIED (`162ab424a`) and its
  pre-registered prediction was WRONG in a fourth-outcome way — the 184/277 lanes' story,
  not mine, but the re-route "created a pair OUR gate holds"; my fix does not change that
  either way (early refusal lands `wont_fix`, which the gate ignores).

> **CORRECTED 2026-08-19:** the timestamps in this file reading "~12:00 UTC"/"~12:10 UTC"/
> "~12:35Z" above are BST misread as UTC — the real times are ~1h earlier (the 488 ledger
> row, written immediately after the council submission, reads 11:05:25+00). The measured
> DATA is unaffected (queries carried their own timestamps); only my narration drifted.
> Caught by the schema_migrations `applied_at`.

## 2026-08-19 — 488 APPLIED and ledger-recorded (11:05:25Z)

`488 OK` NOTICE fired; ledger row `applied_by='bugfix_301_owned_guard_ordering lane'`.
The config half is now LIVE (old binary: warns once + ignores the key; the error-routing
alignment for genuine load errors is active now). The refusal activates at the next
chassis roll carrying commit `6be66bceb`.

## 2026-08-19 — checkpoint (session pausing on usage limit)

DONE: fix committed `6be66bceb` (Go + tests + 488 + PBP-045 + bug-file note + standing
five); 488 APPLIED + ledger-recorded 11:05:25Z; parity drift fixed separately
(`1c16eb692`, one line, overlay re-applied — cronjob `configured`).

STILL OPEN, in order:
1. **READ the council verdict** — corr `c7bc1b9e-97c8-4f3e-8a4f-b3a7029505ee`. Query in
   CLAUDE.md/097 output; act on REVISE/REJECTED (code is already on the shared branch).
2. **READ the 090 verdict** — run `dd61df1b`, was at call_diagnoser ~11:10Z. If it refutes
   the "no write path bypasses save_sections" premise, act BEFORE the next roll (rollback
   file exists and is exercised).
3. **Post-roll behavioural verification** — RUNBOOK, BOTH controls (owned → wont_fix +
   refused_by='load_page_record' + no writer child; generic → writer runs and saves).
   Then, and only then, 301 moves toward closed (fixed-AND-live bar).

## 2026-08-19 — session 2 (after the fresh roll): LIVE, PROVEN, council round 2 in

**Deploy proven at the artefact** (pods `-bfw5n`/`-nkdkl`, started 12:15Z): `refuse_owned_page`
PRESENT in `/proc/1/exe` on BOTH replicas (a literal that exists only in `6be66bceb`),
`OWNED_PAGE_GUARD` PRESENT (long-lived control), nonsense needle ABSENT (probe discriminates).
Provenance log line already scrolled — "not in range", never "unstamped".

**BEHAVIOURALLY PROVEN, both controls** [MEASURED ~16:05Z]:
- POSITIVE: 3 owned-page `content_rewrite` items refused at 13:37–13:38Z with
  `step load_page_record failed: ... OWNED_PAGE_GUARD: page learn-algorithms-...` — all 3
  `wont_fix` with `result.owned_page_refusal`; **0 writer orchestrations in their window**;
  3 handler runs for the 3 items, `complete_error` at `failed_step=load_page_record`.
- NEGATIVE: 6 generic-page builds post-roll spawned writers normally; they later failed at
  `validate_content` — a step this change cannot touch, at the historical rate (172 in the
  7-day live window, 14 today ≈ ~25/day baseline). One build in flight at `call_content_writer`.
  ⚠ HONEST LIMIT: no generic build has COMPLETED end-to-end post-roll yet (all fell to the
  pre-existing validation failures), so "writer runs" is proven and "page saves" is inferred
  from the change not touching the save path — re-check a completed build tomorrow.
- Save-path refusals: **0 since the roll** — backstop-only now, as designed.
- Review-row dedup as designed: 0 NEW `refused_by='load_page_record'` rows because all 3 pages
  already hold an open `owned_page_review` row (ON CONFLICT DO NOTHING converged).

**090 verdict recorded, as promised, either way:** `UNVERIFIABLE` — iteration cap, neither
confirmed nor refuted; best-effort trail attached, no fix proposed. Superseded by the
production discriminating pair above.

**Council round 1: REVISE** (11:11:54Z, gating: `debug_historian` HIGH — 488 lacked the
`snapshot_agent()` opener the runner's header requires). Conceded + remediated: pre-488 dump
committed (`PRE_488_page_build_handler_workflow_steps.json`, dumped BEFORE apply), post-apply
snapshot taken (source_id `8f35c080`), WRONG_CALLS entry written (the trap: hand-applying to
avoid the runner also bypassed the documentation living in the runner). 488 itself NOT edited —
its md5 is in the ledger. Other objections all answered by measurement or the committed code:
attempt_count incremented on BOTH paths (v3:5748); the divergence is only the re-triage ladder,
affected population = 0 genuine load errors in the live 7-day window; both return paths wired
(quoted); count(*)=1 and 480-precondition guards ARE in the applied SQL (sketch elision, my
error); census re-run (still exactly 2 carriers, trh key `(absent)`); the widening-precedent
landmine cited and distinguished (per-step opt-in is the reason the wider net cannot catch the
tool pipeline). The 146 figure re-derived as a split: 84 failed / 36 unresolved / 13
needs_human_review / 9 detected (=142, moves with traffic; reviewer's 64 = narrower status set).

**Round 2 RESUBMITTED** on the same correlation: `RESUBMIT_CORR=c7bc1b9e`, run orch
`6469c138-3e88-492d-b2ba-5b60ab63a1ea`. Verdict owed a read (~30 min budget).

## 2026-08-19 — session 3 (fresh, from the HANDOFF): both reads clean → CLOSED

**§1 council round 2: APPROVED** 16:19:04Z (`diagnosis_artifacts`, corr `c7bc1b9e`, kind
`council_report`): "approved with 3 advisory objection(s) — none high-severity", 3 abstained. Every
advisory answered by an independent check this session (not by re-asserting the submission):
- `bug_historian` medium — "rests on `bugs_open/086` (step-level error_step dropped by the plan
  converter), still open?" → `ls bugs_open bugs_closed | grep ^086` → **`bugs_closed/`, CLOSED
  2026-07-27**, both halves live. The seat's "OPEN case" premise was stale.
- `diagnosis_guardian` medium — "confirm independently that coordinator.go honours STEP-level
  error_step" → `platform/orchestration/coordinator.go:3671-3676` (`routeToErrorStepOrFail`):
  `if step.ErrorStep != "" { routeToErrorStep(..., step.ErrorStep) }` FIRST, then the
  `step.Config["error_step"]` fallback; `:3393-3399` same order in the other path. The seat's
  "standing discipline" (config-only) is wrong for THIS engine — it may be right for the diagnosis
  loop's coordinator, which is what the round-2 rebuttal said.
- `debug_historian` medium/low — which label/pods, and was the digest checked →
  `kubectl get pods -o jsonpath=…image…imageID` over the namespace: **22 pods on
  `agent-chassis:v1.0.1316`, ONE imageID (`sha256:2d0d3def…`)**, spread `agent-chassis` 2 /
  `dynamic-agent` 17 / `business-intel` 1 / `vet-intel` 1. Both `agent-chassis` replicas were
  REPLACED at 17:13Z (after the 12:15Z probe), so re-probed `-86nqf`/`-8jlqh`: `refuse_owned_page`
  PRESENT, `OWNED_PAGE_GUARD` PRESENT, `ZZQQ_ABSENT_NEEDLE` absent on both. (First exec timed out at
  2 min on the second pod — a kubectl flake; retried alone with `timeout 90`, clean.)
- `bug_historian`/`editquality` low — matcher comment not updated → it WAS, in `6be66bceb`
  (`owned_page_guard.go:74-77` EMITTERS block); grep of every `ownedPageSkipReasonPrefix` site
  (`save_page_sections_action.go:79,225,242`, `v3_site_actions.go:822,985,5781`, `multipage_actions.go:65`
  = AssemblePage's skip_reason emitter, `load_page_record_action.go:275`) agrees with the 3 matchers
  listed. No platform edit needed post-verdict.
- `tooling_provenance` medium — leave a NOTES trail on the agent's doc subject → no `doc_plans`
  row exists for `page-build-handler`; not created (convention question, not this lane's).

**§2 completion: CLOSED.** `current_step='complete', status='COMPLETED'`: **2** post-roll
(`78a7f1ea` 16:16:55→16:22:07Z, `214074b9` 20:06:11→20:16:13Z). `collected_data` keys on both
include `call_content_writer`, `save_sections`, `sections_saved`, `deploy_page`, `deploy_result`.
⚠ misstep caught in-session: my first check tested `collected_data ? 'save_page_sections'` (the
ACTION name) and read FALSE — the STEP key is `save_sections`. Read the step names off the live
row (`jsonb_object_keys(default_config->'workflow'->'steps')`) before keying on them.

**Refusals re-measured, by `__step_error` not `error`** (my first `error LIKE '%OWNED_PAGE_GUARD%'`
count read **0** — `error` is NULL on a routed failure, the 099 landmine; I knew it and still typed
it first): post-roll `complete_error` = 10 unmarked (validate_content class) + **4 marked**
(13:37:08, 13:37:59, 13:38:35, **20:37:46Z** new — `tool-prompt-permutator`). All 4 work items
`wont_fix` + `result.owned_page_refusal` (5b683259, 9996bc0c, d0a2a069, 629e4a36). Writer children
20:30–20:45Z: **0**. `owned_page_review` since roll: **1 row, `refused_by='load_page_record'`** —
the first direct new-row evidence (the 4th page had no open row to dedup onto). Save-path
refusals since roll: 0.

**Candidate 3 census for the owner decision** [MEASURED ~20:55Z]: owned-page queue at
`page-build-handler` = 84 failed / 36 unresolved / 13 needs_human_review / 9 detected (=142).
Producers hard-coding the handler: 12 grep hits in `platform/orchestration/actions/` (listed in
the bug file's closing section). `grep -l rebuild_policy bugs_open/*` → 146, 208, 232, 224, 283,
263 — none is the producer-routing defect (208 is the rebuild-route sibling). So it is the untaken
candidate in TWO closed files (295 + 301) and nowhere else. Flagged, not filed (HANDOFF said the
owner decides; README carries the question and the recommendation (a)).

**Close actions:** closing section appended to the bug file; `git mv` → `bugs_closed/` (same
commit); 016b §10 row 301 + §9 pattern "a guard at the LAST step…" (additions only, 39/0);
`MEMORY_closed.md` line + topic file `bugfix-301-owned-guard-ordering.md` (no line in the capped
index — this lane never had one, and the practice lives in 016b §9). No platform code touched
this session.
