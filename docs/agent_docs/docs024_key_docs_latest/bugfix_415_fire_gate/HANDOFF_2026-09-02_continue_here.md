# HANDOFF 2026-09-02 — bugfix 415: the fire-gate `pre_query` is narrower than the selector — ~~CONTINUE HERE~~ DONE

> **COMPLETE 2026-09-02 (same day, fixing session = the resumed "bugs_open/413" session).
> Do not start work from this file — it is a record now.** Migration `688` (+`_ROLLBACK`/
> `_VERIFY`) committed `59e722812`, applied 13:28Z, VERIFY green, trigger fired 6 s
> post-apply, all three narrownesses simulation-proven load-bearing. Bug file moved to
> `bugs_closed/415` with the full closing evidence. Register WDS-002/WDS-003, LANDMINES
> (three corrections + re-sync) and the throughput lane's 08-30 handoff all updated.
> Council `Council-Submitted: 5f0cb450-e40f-4ffd-ac8e-01534caeac25`.
> Two traps below went STALE between the handoff cut and the fix, both re-checked at the
> artefact: (trap 3) the `213` file had become TRACKED (passenger in `0f3721c6e`) — a live
> whole-dir `--apply` hazard, renamed `_SUPERSEDED` in `5cd756b99`; (trap 6) the fleet LLM
> outage had CLEARED (76 ok / 0 failed in the hour before submission), so the council ran
> normally. Everything else below was verified still true before acting on it.

Written by the session that filed 415 and fixed/closed its sibling 413 (session name
"bugs_open/413"). That session's arc is done; this file is everything the fixing session
needs without re-deriving it. **The bug file is the shared account — contribute into
`bugs_open/415_HANDOFF_2026-08-26_fire_gate_prequery_spelling_drifted_from_the_selector_so_approved_only_backlogs_never_fire_the_trigger.md`,
and read it first** (mechanism, severity honestly stated, prior art, candidates ranked).

## The one-paragraph state

`scheduled_tasks.pre_query` on `build-pipeline-trigger` (and its disabled `-2` sibling —
kept as ruling B's rollback path) gates whether the dispatcher fires at all. It counts sites
with a `status='triaged' AND pipeline='build'` row under a bare `s.locked_at IS NULL` — while
the selector it gates (`find_dispatchable_site`, post-657) admits `('triaged','approved')`,
has NO pipeline filter, and honours the lock-EXCEPTION arm. So the gate is narrower than what
it gates in **three** independent ways (the lock arm is the third — found after filing, see
below), and a backlog that is approved-only, non-build-pipeline, or entirely on a
lock-excepted site never fires the trigger. No error anywhere; the damage would be an
absence (the 413 meter-blindness class). **Theoretical at today's volume** — the fleet
almost always holds triaged build rows — but the door is reachable (end-of-backlog is
exactly where the throughput work drives), the fix is cheap, and this drift class has now
bitten three times on this seam (078→285 eligibility, 413→657 ordering, this).

## Verified live 2026-09-02 ~2xZ (by the filing session, at the artefact)

- Both trigger rows share `md5(pre_query) = 200246f7ede3e33b14be2fc064efa7da` — **use this
  as the migration preflight anchor**, and re-read it at your session start (it can drift
  beneath you; refuse on mismatch, never blind-replace).
- Gate text byte-identical to the 08-26 filing read (quoted in full in the bug file §"The
  drift"): `s.locked_at IS NULL` bare (no `lock_except_item_ids` arm — **confirmed, this is
  the third narrowness**; the selector's cross-site spelling is
  `(s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[])))`),
  `wi.status = 'triaged'` only, `wi.pipeline = 'build'` present.
- 413/657 context: CLOSED 2026-09-02, measured PASS (+2h and +3.3d). Selector md5 is
  `d29807313a8f6ed543a541c35c1626c4`, VERIFY = `657_selector_ranks_sites_by_loadable_work_VERIFY.sql`
  (bare-named since 09-02; historical docs naming `_HOLD` were true when written).

## The chosen fix (candidate 1; owner NOT blocked)

**Widen the gate so its admission is a strict superset of the selector's.** A gate may be
WIDER than its selector — a spare fire is one cheap no-op tick (the selector returns 0 rows,
`check_has_site` ends the run); it must never be narrower. Concretely, the new pre_query
keeps the cheap per-site EXISTS shape and:

- `wi.status IN ('triaged', 'approved')` — was `= 'triaged'`;
- **drops** `wi.pipeline = 'build'` (neither the selector nor the loop's `load_items` config
  filters on pipeline — verified during the 413 work);
- keeps `attempt_count < max_attempts` and the retry_after arm (selector has both);
- **adds the lock-exception arm** in the selector's own cross-site spelling (⚠⚠ do NOT
  reuse the per-site Go fragment — `work_items_common.go:851-870` explains why the two
  spellings must stay different);
- deliberately does NOT add approval_mode/depends_on/busy-skip (selector-side narrowings;
  omitting them keeps the gate wider, which is the safe direction — say so in the migration
  header so the next reader doesn't "fix" it).

The alternative (candidate 2, drop the pre_query entirely) costs ~1.3s of selector execution
per 60s tick even on an idle fleet; presented to the owner as the simpler option, proceeding
with candidate 1 unless he says otherwise. Candidate rationale + verification shape:
the bug file §"Fix candidates" / §"How to verify".

## Traps, each one load-bearing (read before writing SQL)

1. **UPDATE BOTH ROWS in one statement** (`WHERE name LIKE 'build-pipeline-trigger%'`, or
   two explicit names) — the disabled sibling is the rollback path, and a by-name UPDATE on
   just the enabled row desyncs it silently (LANDMINES 2026-08-25 sibling entry). Assert
   ROW_COUNT = 2.
2. **The 584 daily VERIFY's assertion 1/7 pins PARITY, not text** — `count(DISTINCT
   (md5(coalesce(pre_query,'')), target_agent_type, target_topic, fire_message)) = 1` across
   the rows (`584_dispatch_sibling_C_insert_trigger_2_VERIFY.sql:24-33`, read 2026-09-02).
   So a both-rows migration keeps it green and **no lockstep edit is owed** — but tell the
   dispatch_throughput lane anyway before applying (session "throughput"; the VERIFY is
   their monitoring commitment to the guardian, and they run it daily).
3. **`213_dispatch_gate_matches_dispatcher.sql` is UNTRACKED-but-staged by some session** —
   it was written 08-12 for exactly this alignment and was never applied. Do NOT apply it
   (pre-633: no lock-exception arm; by-name UPDATE that misses the sibling; pre-657 world).
   Treat as prior art; coordinate a `_SUPERSEDED` rename with whoever staged it rather than
   deleting another session's work.
4. **`regexp_replace(..., 'n')` over multi-line text silently replaces nothing and still
   reports UPDATE 1** (LANDMINES, keyed `scheduled_tasks.pre_query`, 08-24). The pre_query
   is multi-line: replace the WHOLE VALUE (md5-preflighted), never regex-edit it.
5. DB config is live the moment it applies — but this change only makes the trigger fire
   MORE, so there is no measurement-window coordination owed (unlike 657). Still ping the
   throughput lane at commit + apply; their floor meters would show any surprise.
6. **A fleet LLM outage has been live since ~08-28** (99%+ fail; owner pinged re top-up).
   Irrelevant to this DB-only change, but it makes the queue quiet — a quiet gate is partly
   demand-starvation, so don't read post-apply fire counts as the fix's effect without a
   demand control.
7. Migration numbering: **check the next free number at write time** (`ls | grep -oE
   '^[0-9]{3}' | sort -n | tail`) — numbers are not a mutex and the directory moved past 658
   during the last week. House style for scheduled_tasks edits: migration 637 (guarded
   UPDATEs, preflight, DO/RAISE post-check, rerun-safe refusal); note `snapshot_agent()` is
   for agent_definitions and does not apply here — quote the old value in the ROLLBACK
   instead.
8. Sidecars: `_ROLLBACK` restores the exact old text (anchor md5 200246f7...) on BOTH rows;
   `_VERIFY` asserts — parity across rows, `'approved'` present, `pipeline` ABSENT,
   `lock_except_item_ids` present — and must be **mutation-proved** (it should FAIL against
   the pre-fix text; run it before apply expecting failure, after expecting pass).
9. **Council**: migrations are in scope (widened 08-19). One round per coherent task;
   `DRY_RUN=1` tests admission free; budget ~30 min; `Council-Submitted:` on the commit,
   `Council-Reviewed:` only after reading an APPROVED verdict. Lessons from 657's rounds
   that transfer: answer objections with run checks, not assertions; when your SQL must pick
   "the row the system uses", copy the system's own selection rule verbatim; make every
   in-query cast total (the guardian will ask).

## Verification (from the bug file, sharpened)

Induce, don't wait: with the fix live, hold one row at `status='approved'` as a site's only
eligible work during a window where triaged-build count is zero fleet-wide (or simulate: run
the OLD gate text and the NEW side by side against live data and show a population the old
one misses — the cheap, non-invasive proof; the 391 lane's mirror-the-code-exactly rule
applies). Disconfirming result: with only approved rows pending, the trigger fires and the
site is served within one interval.

## Docs owed as you go

Bug file 415 (the account) · WDS-002 register bullet (the "three spellings" line becomes
"aligned by NNN") · LANDMINES: the "Two queries decide dispatchable and DISAGREE" entry and
the 08-20 `pre_query` retry_after entry both reference the gate — correct in place +
`landmines-verify-dispatch.sh` after · WRONG_CALLS if you get something wrong ·
this dir gets NOTES/README if the work outlives one session. Commit per task by pathspec;
grep LANDMINES for `scheduled_tasks`/`pre_query`/`build-pipeline-trigger` at session start
(the hook only shows dirty-file matches).

## Owner decisions (context, none blocking — UPDATED 2026-09-02 late)

Candidate 2 for 415 (delete the gate) only if he prefers it over widening — and the
throughput lane has confirmed 415 is admission-correctness, not ordering, so it stays this
lane's call: proceed with widening. **The adjacent ordering questions are now RULED
(provisionally): 413's candidate 2 age-floor DECLINED** ("no need to reorder — flow +
capacity"; recorded with its mechanical revisit trigger in `bugs_closed/413`'s closing
section), and the same decline-by-default covers any other reordering variant. **Ordering
decisions route to the dispatch_throughput lane** (the 08-26 split), not to the owner
directly. D4 / credit top-ups remain with that lane.
