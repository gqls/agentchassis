# HANDOFF — bug sweep lane, 2026-09-04 (442 thread)

> ⚠ **THIS DIRECTORY HOLDS TWO CONCURRENT THREADS WITH NEAR-IDENTICAL FILENAMES.**
> `HANDOFF_2026-09-04_bugsweep4_continue_here.md` is a **different session** (bugs 361/366/400).
> **This file is the 442/464 thread**, continuing `HANDOFF_2026-09-03c_continue_here.md`
> (path: `docs/agent_docs/docs024_key_docs_latest/bugsweep_2026_08_26/`).

**Read this instead of `HANDOFF_2026-09-03c_continue_here.md`.** That file is still right about the
council, 464, and the four things in its §4. **Its §3 is wrong** and this file says why — the
correction is the main thing that happened today.

---

## 0. STATE IN ONE TABLE

| item | state 2026-09-04 12:0xZ |
|---|---|
| **442** Go half | ⚖ **STILL LIVE after the `v1.0.1360` roll** — re-probed at the binary on **both** pods, both controls behaving. §4.2 is why this needed re-checking |
| **442** §3 "nothing can be refused" | ❌ **WRONG, corrected.** It is a **drained queue**, not absent demand. 4 dispatches / 5 pages / 0 refusals in 25 h — §1 |
| **442** council (`76288ff9`) | APPROVED round 3, verdict read, advisories answered. **Nothing owed** |
| **442** NEW defect found + fixed | ⚖ **migration `773` APPLIED AND VERIFIED LIVE** — the message said "read each `save_result`" and that key holds only the LAST page. §2 |
| **773** council (`8bf83b59`) | ⏳ **SUBMITTED, verdict PENDING** — was at `gate_render` at 11:58Z. **This is the one thing owed.** §5 |
| LANDMINE (new) | dispatched for verification, corr `10d6ea46-320d-4ab0-8a55-c45453e75534`, verdict pending |
| **464** | CLOSED 09-03, unchanged |
| 338 / 320 / 407 / 359 / 404 | unchanged — not this lane's, see the 09-03 handoff §0 |

---

## 1. THE CORRECTION, AND IT IS THE REASON TO READ THIS FILE

The 09-03c handoff §3 says, in bold: **"no page can be refused and no item can exist."** The
measurement under it is real and I reproduced it (still `0` today). **The conclusion is wrong**,
and it is wrong in the direction that tells you not to bother looking.

**What I noticed.** The demand control says nothing is eligible — and the backfiller had *run twice
this morning and written a page each time*. Both are true, which is the tell.

**Why both are true.** Read at the deciding arm, `cmd/scheduler/main.go`, the `dynamicData == nil`
branch: when a scheduled task's `pre_query` finds nothing, the scheduler **stamps
`last_triggered_at`/`last_completed_at` and `continue`s** — no message, no orchestration row. So an
empty tick leaves **no trace in `orchestration_states` at all**, only a moved stamp. And
`meta-description-backfill`'s `pre_query` **is** §3's demand-control query, verbatim, grouped by
site with `LIMIT 1`. The two are one instrument read at different moments.

So the zero is measured **at rest, after the hourly drainer has cleared the queue.**

`[MEASURED 2026-09-04, whole 25 h retention window 09-03 10:21Z → 09-04 11:24Z]`

| | |
|---|---|
| dispatches | **4** (all reached `complete`; none `complete_nothing_to_do`) |
| pages offered / written / **refused** | 5 / 5 / **0** |
| eligible at rest, now | 0 |
| stamp at 10:56Z with no 10:56 orchestration | an **empty tick**, confirmed at the arm |

**The correct sentence: the gated action runs about five times a day and passes every time. The
refusal branch is UNTAKEN, not unreachable.** The first `meta_description_refused` row is still the
acceptance evidence and it will arrive on its own. **Do not force one.**

This is §9b's own misstep — a zero with two sufficient causes attributed to one — recurring one
section later in a file that already carries the warning about it.

---

## 2. WHAT WAS FOUND AND FIXED TODAY

**The defect.** 728 made the completion message name all seven refusal reasons. It says *"Read each
`save_result`"*. A key of exactly that name exists, is populated, is the right shape — and on a
multi-page run **holds only the LAST page**. The per-page truth is the `save_result_0`,
`save_result_1`, … series beside it. A run where page 0 is **refused** and page 1 is **written**
reads `{"updated": true}` at the key the message names. 1 of the 4 runs in the window was
multi-page.

**It is not this workflow's quirk, and not a uniform rule either.** `[MEASURED]` on
`page-content-writer`, deriving the last index **per row**, bare `copy_gate` == `copy_gate_<max N>`
on **20 of 20** runs and == `copy_gate_0` on **0 of 20**. But `section_output`, in that same loop,
has **no bare form at all** — so neither presence nor absence is informative and no reader can learn
the rule from one example.

**`[UNTRACED]` which code writes the bare key.** `makeIterationOutputField` rewrites each injected
step's `OutputField` to `{field}_{N}`, so the injected steps are demonstrably not the writer, and I
did not find what is. **Everything shipped says so explicitly.** If you trace it, the LANDMINE entry
is where it belongs — and do not let a later edit upgrade the observation to a cause.

**Shipped:** migration `773` (+ `_ROLLBACK`), commit `75f7b843d`, applied 11:57Z, ledger-recorded,
verified at the live row (`names_series t`, `warns_bare t`, `728 intact t`, negative control `f`),
and **re-running it now aborts on its own drift guard** — which is what proves the guard rather
than merely showing it exists. Round trip (`773` then `773_ROLLBACK`) was proven in one rolled-back
transaction *before* applying.

**Also filed:** a LANDMINE entry for the general shape, a `WRONG_CALLS` row (§4), and `442` §11a–f.

---

## 3. ⚠ THE 09-03c §4 LIST STILL STANDS — plus one it did not have

Everything in the previous handoff's §4 is still live and still true (`overwrite_existing` must stay
undeclared; a roll ≠ your code shipped; a roll kills an in-flight council; another session's edits
in this lane's test file; the tree may not compile). **Add this:**

**§4.1's landmine fired on MY OWN verification query, in the session that was quoting it.** I
checked the owner-ruling key on the **backfiller** at the top-level path
`{workflow,steps,save_description,config}` and got **blank**. The backfiller's save step is nested
in `backfill_loop.sub_workflow`, so that path does not exist — **blank is NULL, and NULL is not
`f`.** Read correctly, with a positive control proving the path resolves:

```sql
SELECT default_config#>'{workflow,steps,backfill_loop,config,sub_workflow,steps,save_description,config}' ? 'overwrite_existing' AS declares,
       default_config#>'{workflow,steps,backfill_loop,config,sub_workflow,steps,save_description,config}' ? 'page_id_field'      AS control
FROM agent_definitions WHERE type='meta-description-backfiller' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;     -- expect: f | t
```

I was checking for **absence**, so the false negative happened to agree with the right answer. Had I
been checking for presence, the blank would have read as good news. **Always pair a `?` check with a
control key that must come back `t`.**

---

## 4. THE MISSTEPS, BECAUSE THEY ARE THE PART THAT DOES NOT RE-DERIVE

1. **A fixed comparand fabricated counter-evidence to my own hypothesis.** Testing "does the bare
   key hold the last iteration", I compared against a hard-coded `copy_gate_4` and got **3 true, 2
   false** — "mostly holds", the most dangerous possible result, because a mostly-true rule invites
   a hedge instead of a re-check. The two were runs of 3 and 4 sections: the comparand did not
   exist, `->` returned NULL, and `jsonb = NULL` is NULL. **Postgres does not complain, so the row
   reads as genuine disagreement rather than as a broken query.** Deriving `max(N)` per row: 20/20.
   Logged to `WRONG_CALLS.md`. **General form: in any per-iteration comparison, derive the last
   index from the row, and make a missing comparand a LOUD third state.**
2. **My first hypothesis was wrong in its direction and still landed near something real.** Seeing
   one `save_result` on a two-page run I assumed the loop **overwrote** and earlier pages were lost.
   The code refuted it before it reached any document (`save_result_0/_1` are preserved). Had I
   written it up on sight I would have shipped a false mechanism attached to a true symptom — the
   hardest kind to unpick, because the symptom keeps confirming it.
3. **§3's zero** (§1 above) — mine to own: I wrote that sentence yesterday.

---

## 5. WHAT IS LEFT

**Owed, and it is the only thing owed:**
- **Read the `8bf83b59-cda4-43e1-af07-838ea10c1df7` verdict** and act on a REVISE/REJECTED. The
  migration is **already applied** and the branch is shared, so a verdict is answered by a *further*
  migration, not by holding this one back. `Council-Submitted:` is on `75f7b843d`; **`098` credits
  it automatically once approved — do not write `Council-Reviewed:` on a verdict you have not read.**
  ```sql
  SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
  WHERE correlation_id='8bf83b59-cda4-43e1-af07-838ea10c1df7' AND kind='council_report' ORDER BY created_at;
  ```
  ⚠ **Count the rows, do not `LIMIT 1`** if you resubmit — a resubmission reuses the correlation and
  the newest-first query hands you the OLD verdict (theme_kits lane, 09-03).
- **The landmine verifier's verdict** on `10d6ea46-320d-4ab0-8a55-c45453e75534`.

**Open on 442, unchanged from 09-03c — the file stays OPEN:**
- **No verifier** for `meta_description_refused`. Five build guards and a live migration amending the
  claimed-item-timeout sweep's `pre_query`, merged with any other lane's pending amendment.
  `discovery_checks/verify_required_fields_missing.go` documents the sequence. Not casual work.
- **`voice_gate_unreadable` is still silent**, correctly, and still strands a page for ever.
- **§9d's second silent path is still unbuilt** — a page the *writer* drops never reaches the action.
  `[MEASURED 2026-09-04]` 0 of 5 pages dropped, 0 of 4 runs — **and that rules out nothing below
  roughly a 45% omission rate**, on runs of one page each where the instruction has least occasion
  to fire. Two clean samples now, both underpowered, both with their denominators written down.

**Adjacent, other lanes' (measured today, not carried forward blindly):**
- ⚠ **HEAD's `platform/orchestration/actions` is RED** at `541193665` — build green, **two** tests
  failing, **both** from `83407cd37` (the `bugs_open/440` lane): `TestTemplateExecutorsAreDeclared`
  (undeclared executor `renderFailWorkItemMessage`) and `TestFindingCodeScanEveryWriteIsRegistered`
  (`FAIL_WORK_ITEM_MESSAGE_TEMPLATE_FALLBACK` undeclared). The `theme_kits` lane reported the
  **first** on 09-03; **the second appears unreported**, so a lane fixing only what it was told about
  stays red. Both guards are working as designed — do not patch them. **This lane's own 10 tests
  pass at HEAD** when run by name.
- **Migration number collisions:** the 09-03c handoff noted two `728`s. **There are also two `734`s**
  (`734_classifier_reads_the_positioning_register` and this lane's `734_meta_description_repair_agent`).
  No damage — the ledger keys on filename — but a bare number is ambiguous here too. `773` was free.
- `097` now WARNs on **two** unclassified migration suffixes, `_ISLAND` **and** `_RELOCK` (the
  09-03c handoff named only `_RELOCK`).
- **8 abandoned `~/.claude-scratch/head-verify/` trees, ~3.9 GB**, other sessions'. Disk at 48%, so
  not urgent. `scripts/scratch-report.py` is the tool; do not hand-roll an `rm`.

---

## 6. HOW TO RE-CHECK EVERYTHING ABOVE WITHOUT TRUSTING ME

```bash
# 1. Is 442's Go still in the running binary? BOTH CONTROLS ARE THE POINT.
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for s in meta_description_refused meta-description-repair candidate_looks_internal ZZZ_cannot_exist_9f3a; do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$s" /proc/1/exe && echo "PRESENT $s" || echo "absent  $s"
done   # expect PRESENT, PRESENT, PRESENT, absent
# ⚠ do NOT use `logs | grep 'build provenance'` on the CHASSIS — it matches the landmine
#   corpus ABOUT build provenance and returns megabytes.
```

```sql
-- 2. Is 773 live, with a negative control so the instrument can fail?
SELECT (default_config#>>'{workflow,steps,complete,config,result_message}') LIKE '%save_result_0%'      AS names_series,
       (default_config#>>'{workflow,steps,complete,config,result_message}') LIKE '%ONLY THE LAST PAGE%' AS warns_bare,
       (default_config#>>'{workflow,steps,complete,config,result_message}') LIKE '%voice_gate_unreadable%' AS r728_intact,
       (default_config#>>'{workflow,steps,complete,config,result_message}') LIKE '%ZZZ_not_present%'    AS negative_control
FROM agent_definitions WHERE type='meta-description-backfiller' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;   -- expect t | t | t | f

-- 3. The DEMAND reading that §1 is about. A 0 here means DRAINED, not dormant —
--    so read it together with the dispatch count below, never alone.
SELECT count(*) FROM pages WHERE status='active'
  AND COALESCE(meta_description,'')='' AND page_visible_text_len(id) > 200;

SELECT to_char(created_at,'MM-DD HH24:MI') AS run,
       (collected_data#>>'{pages_missing_meta,count}')::int                AS offered,
       jsonb_array_length(collected_data#>'{written,result,descriptions}') AS written
FROM orchestration_states WHERE owner_agent_type='meta-description-backfiller'
ORDER BY created_at DESC;

-- 4. Has the mechanism finally fired? (the acceptance evidence)
SELECT status, count(*) FROM site_work_items WHERE item_type='meta_description_refused' GROUP BY 1;
```
