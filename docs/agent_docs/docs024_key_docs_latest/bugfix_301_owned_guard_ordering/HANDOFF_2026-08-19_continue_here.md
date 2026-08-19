# HANDOFF — 2026-08-19, fresh session starts here: the fix is LIVE and PROVEN; two reads and one move remain

**Lane:** `bugfix_301_owned_guard_ordering` — bug file
`bugs_open/301_HANDOFF_2026-08-18_page_build_handler_runs_the_llm_writer_and_link_resolver_before_the_owned_page_guard_so_the_work_is_thrown_away.md`
**Read this file, then `NOTES_owned_guard_ordering.md` from the bottom.** The PLAN holds the
design decisions and their reasons; the RUNBOOK holds every query below with its gotcha.

---

## 0. STATE IN ONE PARAGRAPH

`page-build-handler` used to run the LLM writer + link resolver and only THEN refuse an
owned page (at `save_sections`, the last step). The fix — opt-in `refuse_owned_page` on the
`load_page_record` action + migration `488` (key + `error_step: mark_item_failed` on exactly
that one step) — is **committed (`6be66bceb`), applied (488 ledger-recorded 11:05:25Z), LIVE
on both replicas of the 12:15Z roll (binary-probed with controls), and BEHAVIOURALLY PROVEN
on live traffic**: 3 owned-page items refused at load 13:37–13:38Z, all `wont_fix` +
`result.owned_page_refusal`, **zero writer orchestrations spawned for them**; 6 generic
builds spawned writers normally in the same window; `refused_by='save_page_sections'`
reviews since the roll: 0 (backstop-only now). The save-path guard was deliberately KEPT.

Commits this lane made today: `6be66bceb` (fix), `25ca816c7`/`5949d9ce3` (docs),
`1c16eb692` (side-find: budget-cron parity — done, closed, not this lane's problem any more).

## 1. REMAINING ITEM 1 — read the council ROUND-2 verdict and act on it

Round 1 was **REVISE** (11:11:54Z), gated by `debug_historian` HIGH: 488 omitted the
`snapshot_agent()` opener. Conceded + remediated (see §3). Round 2 was **resubmitted on the
same correlation** at ~16:1xZ with every objection answered by measurement or the committed
code (`COUNCIL_SUBMISSION_2026-08-19_refuse_owned_at_load_round2.json` — the round-2 text is
appended to the rationale). At handoff time the verdict had NOT yet landed (last poll
above/NOTES; publish→run-start can be ~29 min — **do not retrigger, find it by payload**):

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id LIKE 'c7bc1b9e%' AND kind='council_report' ORDER BY created_at;
-- full report body:
SELECT body FROM diagnosis_artifacts
WHERE correlation_id LIKE 'c7bc1b9e%' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```

- **APPROVED** → future commits on this work may carry
  `Council-Reviewed: c7bc1b9e-97c8-4f3e-8a4f-b3a7029505ee` (the two existing commits carry
  `Council-Submitted:` and are credited automatically at report time — no amend, forward-only).
- **REVISE again** → answer it the same way: measurement or committed code, resubmit with
  `RESUBMIT_CORR=c7bc1b9e-97c8-4f3e-8a4f-b3a7029505ee`. All round-1 objections and their
  answers are in NOTES (bottom) if a new seat re-raises one.
- **REJECTED** (unlikely — 11 of 14 seats approved round 1, and the gating seat's ask was
  remediation, not reversal) → the guardian notes name the contained alternative; the
  `_ROLLBACK` file exists and is exercised; talk to the owner before rolling anything back,
  because the mechanism is already demonstrably saving spend.

## 2. REMAINING ITEM 2 — observe ONE generic-page build complete end-to-end post-roll

The honest gap in the negative control: post-roll, generic builds spawn writers (observed,
6×) but none has COMPLETED end-to-end yet — every finished one fell at `validate_content`
("content validation failed: N blockers"), a **pre-existing** failure mode (172 in the 7-day
live window, ~25/day baseline) that `6be66bceb` cannot touch (the early refusal is step 2,
owned pages only; these are generic pages dying at step 10). So "the writer still runs" is
observed; "the page still saves" currently rests on the fix not touching the save path.
**Do NOT induce a build** — wait for demand (it runs several per hour in bursts):

```sql
SELECT current_step, status, count(*), max(created_at) FROM orchestration_states
WHERE owner_agent_type='page-build-handler' AND created_at > '2026-08-19 12:15Z'
GROUP BY 1,2 ORDER BY 3 DESC;
-- want: at least one row current_step='complete', status='COMPLETED'
```

If completions stay at zero for another day while writers keep running, that is the
`validate_content` failure rate, not this fix — check whether another lane already owns it
(it predates the roll) before filing anything.

## 3. THE SNAPSHOT MISS — closed, but know it before you touch 488

488 was applied WITHOUT the `snapshot_agent()` opener that
`scripts/migration/run-migrations.sh`'s header requires for every `agent_definitions`
migration (WRONG_CALLS 2026-08-19 has the full story; the trap: hand-applying to avoid the
runner's take-every-pending-file behaviour also bypassed the doc that lives in the runner).
Remediated: the TRUE pre-488 step object (dumped ~09:47Z, before apply) is committed at
`PRE_488_page_build_handler_workflow_steps.json` in this directory; a post-apply
`snapshot_agent('page-build-handler', …)` snapshot exists (source_id `8f35c080…`).
**Do NOT edit `488_*.sql`** — its md5 is in `schema_migrations`; editing an applied,
ledgered file orphans the ledger row. Any future `agent_definitions` migration from this
lane OPENS with `SELECT snapshot_agent('<type>', '<file>: pre-update');`.

## 4. CLOSING 301 — the procedure, and the one decision to put to the owner

The filed defect (the ordering) is **fixed AND live with behavioural proof** — the close bar
is met once §1 and §2 are read. Then:

1. `git mv` the bug file `bugs_open/…301…md` → `bugs_closed/` **with BOTH paths on the same
   commit** (the `git mv` landmine), and verify at HEAD with `git ls-tree`. Resolve by SLUG
   — several numbers name two bugs on this tree.
2. Append a dated closing section first: fixed by `6be66bceb` + migration 488, proven
   2026-08-19 (positive + negative controls, queries in this lane's RUNBOOK), council
   correlation `c7bc1b9e` with the final verdict named AND read.
3. Update 016b §10's index line for 301, and add the §9 pattern if you judge it transferable
   (the shape: "a guard at the LAST step makes every refusal cost the full pipeline — check
   WHERE a guard sits, not only THAT it exists").
4. Update `MEMORY_closed.md` / the memory index per its own budget rules if a line exists
   for 301 (check first — this lane never added one).

**THE OWNER DECISION (flag it, do not decide it):** the bug file's **candidate 3** — the
real upstream defect: producers hard-code `handler_agent='page-build-handler'` for content
findings without consulting `pages.rebuild_policy`, so owned pages keep accumulating
findings that can only ever be refused (~142 open today: 84 failed / 36 unresolved / 13
needs_human_review / 9 detected). Closing 301 buries that in a closed file. Options:
(a) file it as its own small `bugs_open` entry at close time (grep both bug dirs first —
it may overlap `bugs_open/115`'s findings-terminate-nowhere shape), or (b) hand it formally
to the Tier 2 / `copy_quality_two_stage` exchange that owns the adjacent repair question
(their reply and the 277 lane's §5–7 record that exchange). Recommendation: (a), with a
cross-reference to (b) — a routing defect and a repair-design question are different bugs.

## 5. THINGS THAT WILL MISLEAD YOU (all hit or nearly hit this lane)

- **Do not re-quote this lane's counts as live.** The 146/142 queued-findings figure moves
  with traffic and its status split matters (a reviewer's narrower count of 64 was also
  true). Re-run the query, keep the split.
- **Classify refusals by the guard's ERROR TEXT, never by joining `pages.rebuild_policy`**
  (mutable; ~3% disagreement measured by the 277 lane; the archive table drops `error`, so
  the discriminator only works on the live table).
- `owned_page_review` rows dedup per page (`ON CONFLICT DO NOTHING`) — a refusal with NO new
  review row usually means an open row already exists for that page (that is why the 3 live
  refusals produced 0 new rows). Check `item_key LIKE 'owned_page_review:%'` before reading
  absence as a bug.
- The binary provenance line SCROLLS — absence from logs means "not in range", never
  "unstamped". Probe `/proc/1/exe` for `refuse_owned_page` with the `OWNED_PAGE_GUARD`
  positive control and a nonsense-needle negative, per the RUNBOOK.
- The 090 run (`dd61df1b`) ended **UNVERIFIABLE at iteration cap** — it neither confirmed
  nor refuted; do not cite it as support OR refutation. The production discriminating pair
  (3 refusals with zero writers + 6 generic writer-runs) is the evidence.
- Timestamps in early NOTES entries drifted BST→UTC by an hour (corrected in place) — trust
  the queries' own timestamps, not narration.

## 6. Session-start checklist

`git log --oneline -10` · re-read this file from disk (another session may have advanced it)
· `scripts/who-owns.py 301` · §1 verdict query · §2 completion query · then §4 if both read
clean.
