# HANDOFF — bugfix 006, 2026-07-26 (cold-start here)

**Supersedes `HANDOFF_2026-07-21_bugfix_006_start_here.md`**, which is history and carries its own
CLOSED banner. Read this one.

## TL;DR — the case is CLOSED and nothing is outstanding

`006` was three independent errors (A/B/C). All three are answered and the file has moved to
**`bugs_closed/006_HANDOFF_2026-07-16_idea_uk_infra_errors.md`**, which is the authoritative
technical record and states both residuals at the top.

| | state | anything owed? |
|---|---|---|
| **A** runner replica crash-loop | **RESOLVED** — symptom extinct, *how* is `[INFERRED]` | no. Reopen trigger written into the file |
| **B** contact forms deliver nothing | **FIXED, LIVE, PROVEN** — 3 of 12 healed | no. 9 heal organically, owner ruling 07-25 |
| **C** claim-timeout churn | **FIXED & LIVE** — migration `220`, config, survived the 21:03Z roll | **no — council APPROVED round 1, trailer added** |

## Nothing is owed. The workstream is finished.

Council **APPROVED** round 1, corr `5e3531f4-2815-4d4b-8328-b47ee877ddd4`: 13 reviewers, 3
abstained (relevance-filtered), **`unreadable: 0`** (so the `019` void-a-round trap is cleared), 2
advisory objections, none high-severity. Trailer `Council-Reviewed: 5e3531f4-…` is on commit
`16999d664`.

**The one advisory objection worth a successor's attention** (bug_historian, MEDIUM) is answered in
the bug file, not dismissed: the branch completes on orchestration status with no artifact check,
so the accidental repair that a re-run sometimes performed no longer happens. The trade is argued
there explicitly — disagree with the reasoning if you like, but it is written down. The council's
"missing" item (a fail-loud signal for an auto-completed-but-broken item) is **not built**; the
population is queryable via the branch's own marker, and a real detector belongs in
`discovery_checks` as its own reviewed change.

<details>
<summary>How the verdict was checked (kept for the method, not because anything is pending)</summary>

```sql
SELECT created_at, metadata->>'decision', metadata->>'decided_by'
FROM diagnosis_artifacts
WHERE correlation_id = '5e3531f4-2815-4d4b-8328-b47ee877ddd4' AND kind = 'council_report'
ORDER BY created_at;
```

- **APPROVED** → add `Council-Reviewed: 5e3531f4-2815-4d4b-8328-b47ee877ddd4` as a trailer on a
  follow-up commit and note it in the bug file. **Read `decided_by` first** — the trailer is earned
  by APPROVED only.
- **REVISE** → objections come with the reviewers' own read-only checks answered. The fix is
  already live and behaviourally proven, so treat them as follow-up commits and resubmit with
  `RESUBMIT_CORR=5e3531f4-2815-4d4b-8328-b47ee877ddd4`. Update the **sketch** fields, not just
  prose.
- **REJECTED** → a guardian veto. The fix is live, so this is a revert-or-amend decision — bring it
  to the owner. Rollback is one file: `220_..._ROLLBACK.sql`.
- **No row yet** → check the run, and read the queue depth before concluding anything (see the
  trap below):
  ```sql
  SELECT status, current_step FROM orchestration_states
  WHERE collected_data->'input_data'->>'fix_correlation_id' = '5e3531f4-2815-4d4b-8328-b47ee877ddd4';
  ```

**Read the counts, not the headline.** `metadata` carries `reviewers`/`abstained`/`unreadable`; a
non-zero `unreadable` voids the round (`bugs_closed/019`) however confident the decision string
reads.

</details>

---

## What C actually is, in one paragraph

`claimed-item-timeout` is the fleet's only self-heal for a claim whose handler finished but whose
`mark_complete` write was lost. Its auto-complete branch hand-coded an artifact test **per
`item_type`** and had **3 of 18**, so the other fifteen hit the 40-minute reset even when the work
had succeeded — 84 timeouts against 14 auto-completions over 14 days, 11 items recorded `failed`
with the work done. Migration **220** replaces that with **one** branch keyed on evidence the
platform already records for every item type: the handler's own orchestration reaching `COMPLETED`,
found via `initial_request_data->'input_data'->>'work_item_id'` (both dispatch loops put it there
through `call_handler.input_mapping`).

Design docs: `PLAN_2026-07-26_claim_timeout_generic_evidence.md`. Commands and gotchas:
`RUNBOOK_bugfix_006.md`. Missteps: `NOTES_bugfix_006.md`. Owner-facing prose:
`README_where_we_are.md`. Current read-out: `SUMMARY_2026-07-26_bugfix_006_closed.md`.

## Verification already done — do not redo it

- **Both branches proven through the RUNNING scheduler**, not psql: a planted claim with a
  `COMPLETED` handler orchestration was auto-completed with the new marker; the **negative control**
  with a `FAILED` one was correctly left untouched. Probes deleted, zero remaining.
- **Every guard fault-injected and watched to fail** — four in the migration, three in the Go
  lockstep test.
- **Survived the 2026-07-26 21:03Z chassis roll**: `pre_query` still 6131 bytes, byte-identical,
  all three markers present; and the **new** scheduler binary runs it (log line moved
  `scheduler/main.go:238` → `:272` while the row stayed the same).

## Commits

| commit | what |
|---|---|
| `d61b3ace1` | migration 220 + ROLLBACK twin + the Go lockstep test |
| `d5988a8ed` | the closure; bug file moved to `bugs_closed/`; both indexes |
| `7891b2ee1` | the standing five + closing summary |
| `810b1a95b` | `WRONG_CALLS.md` row |
| `c4f3e4f6f` | post-roll re-verification + the dropped-submission correction |
| `dd627c4cf` | this handoff |
| `16999d664` | council APPROVED recorded + the objection answered — **carries the trailer** |

The trailer is on `16999d664` only, which is correct: it is earned by the APPROVED verdict, and the
verdict post-dates every earlier commit. Do not back-fill it onto the others.

## Watch, don't work

1. **The new marker should start appearing.** If it stays empty while timeouts continue, the branch
   is not reaching them and the closure was premature:
   ```sql
   SELECT item_type, count(*) FROM site_work_items
   WHERE error = 'Auto-completed: handler orchestration completed after claim' GROUP BY 1;
   ```
   (Empty as of 21:00Z — expected; orphaned claims are sporadic and there were zero claimed items.)
2. **A contact form still `#contact` AFTER a discovery cycle ran on that site** is a real signal.
   Still broken because nothing has visited yet is not.
3. **`bugs_open/003`** is where the prevention lives. Every claim this net catches is one that
   should not have needed catching.

## Traps this session paid for — the two that will bite a successor

**1. "No orchestration row means QUEUED, not dropped" has no expiry, and I got it wrong.**
The first submission (`9962dd2b-…`) was reported as queued at 18:44Z on that rule. At 21:05Z it had
left **zero rows** across `orchestration_states.collected_data`, `initial_request_data` and
`diagnosis_artifacts`, while another thread's **later** submission had run twice and reached
`complete_revise`. It was dropped. **Queued and dropped are identical from your own row** — the
discriminator is whether something published *after* you has finished. Cause of the drop
**[UNDIAGNOSED]**; it is `bugs_open/003` territory, not this case's.

**2. Neither "6 seconds" nor "30 minutes" is the dispatch latency — queue depth is.**
The resubmission had its `council-gate` row **6 seconds** after publish. The runbook's ~30 minutes
was measured 07-20 under load. Same lane, same day, two orders of magnitude apart. Read
`./scripts/dispatch-queue-depth.sh` *before* interpreting your own silence.

**And one rule I broke ninety seconds after writing it:** resubmitted at 254 s with the chassis pod
at 4m14s, inside the ~300 s post-restart window where spawns are silently dropped. It landed, which
is getting away with it, not being right. The rule stands.

## If you need to undo C

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/220_claimed_item_timeout_generic_evidence_ROLLBACK.sql
```
It restores the pre-220 `pre_query` verbatim (captured from the live row, md5
`27ea3fd389f6843064d421d6a5833e30`) and refuses to complete if the generic branch is still present
or the restored string does not parse. **Do not hand-edit the column back** — it is an 84-line SQL
string in a text column with no compiler behind it, and a typo there silently kills the fleet's only
claim self-heal while every surface still reads healthy.
