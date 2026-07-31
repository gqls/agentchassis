# HANDOFF — bugs_open/099 candidate 2: two steps from closing

> **SUPERSEDED 2026-07-31 18:09 — BOTH STEPS ARE DONE AND THE BUG IS CLOSED.**
> `v1.0.1216` carried the widening (verified both replicas); the induced run
> `53fff682` refused a plan at 18:08:26 and **persisted a repaired plan at 18:09:31
> — 3 stages carrying all 3 edits** where the original had 2 stages with one holding
> 2. All four conditions held. `bugs_open/099` → `bugs_closed/099`; `FIX-057` is
> `deployed`. Kept for the reasoning and for the "what NOT to redo" section, which
> still stands — `fix-proposer` (`bugs_open/162`) and the owed landmine verification
> are the live follow-ups.

**Read this first, then `NOTES` for the reasoning.** Everything is committed. The bug
is **deliberately still OPEN** and the remaining work is short and specified.

## State in one paragraph

The fix is **built, reviewed three council rounds, and LIVE** on chassis `v1.0.1215`
(pod-grepped both replicas). Migration `272` is applied, so `feature-designer` is opted
in. Every stage of the mechanism has been **observed working on the real system**: an
induced refusal preserved a 10,488-byte design, wrote an operator record, and routed to
`repair_plan`. One later improvement — found *by* that live run — is committed but
**not yet in a rolled image**, and that is the only thing standing between here and
closing.

## The two steps to close

### 1. Get commit `69d51e4dd` into a rolled image

It widens the recoverable class to include **schema mismatches** (well-formed JSON in
the wrong shape), which is what killed the first live repair, and populates
`orchestration_id` on the operator record.

**Do NOT roll while another session's council is progressing** — that kills it, and it
was done to this lane twice on 07-31 (both a council round and a verification run died
at `EXECUTING_STEP` with a frozen `updated_at` and no error). Check first:

```sql
SELECT count(*), string_agg(current_step,',') FROM orchestration_states
 WHERE status NOT IN ('COMPLETED','FAILED')
   AND (current_step LIKE 'review_%' OR current_step LIKE 'council_%' OR current_step LIKE 'gate_%')
   AND updated_at > now() - interval '300 seconds';
```

Councils here are near-continuous, so **the cheapest path is to let the next session's
natural roll carry it** — that is exactly how the first half shipped (this lane never
rolled anything). Then confirm:

```bash
kubectl exec -n ai-persona-system <chassis pod> -- sh -c '
  grep -ac "well formed but does not match" /app/agent-chassis;   # the widening — want >0
  grep -ac "diagnose_persist_fix_plan" /app/agent-chassis'        # control — want 11
```

⚠ **Pick a control your diff did not touch.** This lane's first control returned 0 on a
binary that unambiguously contained the change, because the refactor had removed that
exact literal. See `LANDMINES.md`.

### 2. One induced run, and require FOUR conditions

Full procedure in `RUNBOOK_…md` §"Verify the fix on the failing branch". In short:
check nobody else is mid-run on `feature-designer`, arm `max_edits=1`, fire at work
item `7b89fb35-f42c-45d1-b64d-214aff56d918`, **disarm immediately after**.

Require **all four** — the first two are the bug file's, the last two are what make it
discriminating:

1. a `fix_plan` artifact exists for the new correlation, **and**
2. the run does not end at `complete_refused`, **and**
3. a refusal note exists (`kind='iteration_note'`, `metadata->>'note_kind'='plan_validation_refusal'`), **and**
4. the run reached `repair_plan`.

⚠ **099's own stated verification is insufficient and returns a false PASS** — it was
written for candidate 1, which worked, so that run now takes the success path and
satisfies (1) and (2) while the repair loop never fires. Corrected in the bug file.

**Then close:** `git mv` the case to `/bugs_closed/`, update `016b` §10, and set
`FIX-057`'s status from *"built, not yet live"* to live.

## What is deliberately NOT being done, so nobody redoes it

- **`fix-proposer` is still exposed** — same defect, same shared action. The fix is
  written and dry-run clean at `sql_for_agents/273_…sql`, **unapplied on purpose**
  (another lane owns that agent). Filed as **`bugs_open/162`** at the council's request
  and indexed in `016b` §10.
- **`council-gate` is not opted in either**, and that is a positive choice: it is the
  gate that reviews this mechanism. Opting it in would additionally need `summary`
  added to the refusal result, because its prompt renders `plan_persisted.summary`.
- **A fourth council round.** Three rounds each drew a different gating seat; round 1's
  five objections did not recur. Everything from round 3 is acted on (`162` filed) or
  answered by measurement (contract granularity, `ai_service` depth, live `BEGIN`, the
  refactor diff) — see `NOTES`. The gate is advisory; commits carry `Council-Submitted:
  f4a4628f-3b90-4054-a875-f2cf72b83e72`.
- **Size caps stay unrepairable.** `max_stages` / `max_total_edits` ask for less scope
  while the repair prompt forbids dropping scope, so such a refusal burns its round and
  goes terminal — where the platform already lands. `FIX-057`'s open question.

## Owed

**Landmine verification for four entries.** `landmines-sync.py --apply` (the command
CLAUDE.md names) delivers to `doc_notes` but does **not** verify;
`landmines-verify-dispatch.sh` is the wrapper that fires the verifier, and sync only
reports entries new *in that run*, so the window has passed. Fire by hand:

```bash
./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>'
```

for `diagnosisartifacts-kind-is-check-constrained-…`,
`a-null-orchestrationid-never-satisfies-2-…`,
`your-pod-grep-positive-control-can-be-invalidated-…`, and
`outputcontract-is-a-required-field-that-nothing-reads-…`.
(The `snapshot_agent` stub was dispatched on 07-31.)

Not done here because four verifier runs are four in-flight agent runs, and this lane
was waiting on a roll window precisely so as not to destroy in-flight work.

**Worth raising with whoever owns CLAUDE.md:** it names `landmines-sync.py --apply`,
which skips verification. It should name the dispatch wrapper, or say so.

## The one thing to carry forward

Three separate times on this bug, a check would have reported success for a reason
unrelated to what it was checking — the bug file's own verification procedure, the
first cap chosen to trigger the test, and the string used to prove the deploy. A
fourth: a landmine filed that the file already contained three times, because the
footprint was a symbol and symbols are not covered by the SessionStart hook. None were
caught by being careful; each was caught by measuring the specific thing rather than
the thing next to it.
