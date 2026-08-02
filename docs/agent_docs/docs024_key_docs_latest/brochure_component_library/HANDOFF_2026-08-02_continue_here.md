# HANDOFF 2026-08-02 — brochure component library / fundamentallyai.com

**Supersedes `HANDOFF_2026-07-30b_continue_here.md`.** Read this one first; that one
is still accurate about everything it describes except TL-035's status.

---

## 1. Where this lane actually is, in one paragraph

TL-035 — *photograph a page that PASSES, not only one that fails* — is **armed end to
end and not yet demonstrated.** The adapter half went live 07-31. The caller half went
live tonight on **v1.0.1229** (another session's roll; on a shared HEAD that is normal,
not a courtesy). The config flag is **set**. What does not yet exist is a **photograph**:
no acceptance run has completed since the flag was armed, so nothing has been proved at
the artefact. **Until a `Rendered:` line appears in an `acceptance-run` doc_note, the
honest claim is "the wire is connected", not "the capability works."** One run is queued
and waiting in the fleet queue.

## 2. The single next action

Check whether the queued run has completed and whether its note carries a `Rendered:` line.

```sql
-- the queued run
SELECT status, claimed_by, attempt_count, left(coalesce(error,'-'),120)
  FROM site_work_items WHERE id='e8756d1e-1f3d-48ba-9418-88ed70ca1b3b';

-- THE ARTEFACT CHECK -- this is the one that matters
SELECT created_at, subject_key,
       body LIKE '%Rendered:%' AS has_render_line,
       substring(body from 'Rendered:.{0,300}') AS the_line
  FROM doc_notes WHERE categories ? 'acceptance-run'
 ORDER BY created_at DESC LIMIT 3;
```

**Three outcomes, three different follow-ups — do not conflate them:**

| what you see | what it means | what to do |
|---|---|---|
| note exists, **has** a `Rendered:` line with `s3://` URIs | TL-035 is proven. | Mark TL-035 proven in the register (`docs026_concept_register/register/tool-lifecycle.md`, the `verify-later` bullet names this as open item (a)); write the lane SUMMARY — one is owed and this is the milestone that earns it. |
| note exists, **no** `Rendered:` line | The flag did not reach the adapter, **or** object storage rejected the upload. These are different faults. | Read `collected_data->'request_run'->'response'` on the orchestration row — if the request payload shows `capture_renders: true` the chassis did its half and the fault is downstream (adapter or S3). A run that FAILED its checks also legitimately has no render for the failing profile — check the verdict before concluding anything. |
| no note at all | The run has not completed. | It is queue latency, not a fault — see §4. Do not re-queue; check for a duplicate first. |

## 3. What was done tonight, and the two things that were done differently from plan

**Pod-verified before writing the key, both replicas, one exec each.** The ordering is
load-bearing (DB config is live instantly, Go is not) and it was honoured rather than
asserted:

```
capture_renders                 1   target
Rendered: full-page screenshot  1   target (added prose)
request_browser_run             8   POSITIVE control
.response.data.screenshots      0   NEGATIVE control
```

> **The negative control is the point, and the RUNBOOK's original instruction was
> weak.** It said to use `judge_acceptance_results` as a control — that is a *positive*
> control, and it reads identically on a binary built before the change. It proves the
> grep works, never that your code shipped. `.response.data.screenshots` was a real
> concatenated literal in the old binary, deleted when four hand-built path strings
> became one `envelopePaths(field, key)`, so **0 is only reachable post-change.**
> Derive candidates with `git diff <before>^ <after> -- <file> | grep '^-'`, then check
> each against HEAD before trusting it — `criteria_json` appeared in that removed-lines
> list and is still present (it moved), so it would have been a *false* negative control.

**Change 1: it went in as a numbered seed, not a bare `UPDATE`.**
`sql_for_agents/292_acceptance_runs_photograph_a_page_that_passes.sql`. A DB-only write
leaves a key in `default_config` with no provenance whatsoever — no who, no when, no
against-which-binary, no why. Two `DO`/`RAISE` guards; the second asserts a **neighbour**
key (seed 147's `profiles`) because a guard checking only what it just wrote cannot tell
a surgical `jsonb_set` from a write that flattened the whole `config` object. **Both were
induced before the real apply.** Watch for the self-satisfying mutant: `sed`-ing
`'true'::jsonb` globally flips the guard's own expectation too and passes.

**Change 2: it is numbered 292, not 291, and that was my error.** I applied by hand with
`psql` — correct, because `--apply` takes *every* pending file and 17 belong to other
threads — but a hand run writes no `schema_migrations` row, so **the number was never
claimed.** Another session recorded a different `291_` five minutes later. Now recorded
with `run-migrations.sh --record-only … --note '<what was verified>'`. Filed fleet-wide in
`LANDMINES.md`. The two `291:` strings in NOTES are left as printed — that is a
transcript, and tidying it would make it false.

## 4. Why the run is waiting, and why you must not speed it up

Two mechanisms, both working as designed:

1. **`tool_acceptance_due` has a 7-day cooldown per subject.** Every candidate ran 2–4
   days ago, so **nothing fires on its own** and a run must be queued by hand. The
   work-item shape is in the RUNBOOK under *"Dispatching an acceptance run by hand"*.
2. **`build-dispatch-loop` takes items fleet-wide, strictly by priority**, and acceptance
   runs are deliberately **priority 90** so they test the *new* page rather than one about
   to be replaced. At 19:02 there were **19** higher-priority items ahead — other lanes'
   repairs, including the `bugs_open/140` contact-info fix and the `178` content restore.
   It drained to 8 within ten minutes, so this is a wait of tens of minutes, not days.

**Do not raise the priority to jump the queue.** Those are repairs to pages serving wrong
content to real visitors. A verification run is not more urgent than that, and the
ordering is the design working.

## 5. Open, in the order I would take them

1. **The artefact check above.** Everything else on TL-035 is downstream of it.
2. **The lane SUMMARY.** Owed since 07-30 and deliberately not written yet — the five
   headings would have said "the wire is connected" twice running. Once a photograph
   exists the read-out genuinely changes, and that is the milestone.
3. **`verify-later` item (c) on TL-035, which is the one that decides whether any of this
   was worth building: does anyone LOOK?** Renders currently land as `s3://` URIs in a
   note body. There is no surface that puts them in front of a human. A photograph nobody
   opens is the same as no photograph. This is a real design question, not a chore.
4. **Not chosen by the owner but still open from the 07-30b handoff:** `llm-cost-calculator`'s
   two dead CTA blocks; the `/tools/decision-record/index.html` 404 stub and the missing
   `/tools.html` (both stale rows, nothing links to them); a guide page for the
   review-council simulator; a `tool-guide-intro` section (note: a `tool-guide-intro`
   component now appears on `llm-cost-calculator` — check what it is before rebuilding it).

## 6. Things that will mislead you on this lane

- **A roll is not evidence your fix shipped.** Grep a string your change ADDED *and* one
  it REMOVED, on **every** replica, same exec. `strings` is absent from several of these
  images — use `grep -acF` on the binary directly, or you get a confident 0 for target and
  control alike.
- **A render is a look, never a verdict.** `Renders` carries no `failing_checks` by
  construction. If anything starts branching on a render's presence, the two-list design
  has been undone and the whole point is lost.
- **`Renders` is on the FAILED note too, and that is not a slip.** The adapter captures per
  (url, profile), so a two-profile run where desktop fails and mobile passes files a shot
  **and** a render.
- **A verify block of `SELECT`s cannot stop a `COMMIT`.** Use `DO`/`RAISE`, and induce it —
  a guard that has never fired is decoration.
- **Check `git log` before assuming HEAD is yours**, and `git ls-files --error-unmatch <path>`
  before reaching for `add`: `git commit <path>` reads the working tree and ignores the
  index, so a *tracked* file never needs `add`, and adding one leaves it in the shared index
  for another session's bare commit to collect. This lane has been bitten by both halves of
  that in one day.

## 7. Commits

- `97dacc5e8` — arm(TL-035): seed 292, pod verification, RUNBOOK/NOTES/README/register
- `02990d88d` — landmine: a migration number is yours when the LEDGER says so

Earlier, for context: `9cc63c775` (caller half), `72463f51e` (r1 objection fix),
`606c5aa3a` (register), `3162f0271` (lane docs), `62e72ff77` (wrong call).
Council `2c895dd1-adae-4f8e-8acf-4592b8ca3981` — APPROVED r1, 11 approve / 1 object; the
one medium objection changed the code and is written up in
`EVIDENCE_2026-07-31b_TL-035_caller_half.md`.
