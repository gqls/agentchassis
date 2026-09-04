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
| **773** council (`8bf83b59`) | ⚖ **APPROVED at round 2** — 11 seats readable, **0 objections**, `decided_by: all reviewers approve`. Round 1's `revise` was an LLM OUTAGE, not a review — §7 |
| LANDMINE (new) | ⚠ first dispatch **DIED IN THE OUTAGE** (4 attempts, all credit-400). Re-fired: `c034492e` (loop entry) and `b68cf7b8` (council addendum) — §7 |
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

---

## 7. UPDATE 12:15Z — §5's "one thing owed" is DISCHARGED, and finding out why cost the most interesting hour of the day

§5 above says the one thing owed is the `8bf83b59` verdict. **It landed, and the story is worth the
four paragraphs.**

### It came back `revise` — and that was not a review

Round 1, 12:02:36Z: `decision: revise`, but `decided_by` reads **"unreadable reviewer(s):
review_editquality, review_reuse_agent, review_guidelines, review_constitution, review_mission,
review_prior_art"** — `unreadable: 6, reviewers: 3, abstained: 8`. **The three seats that WERE
readable all APPROVED** (`guardian`, `debug_historian`, `architecture`) and the round contains **not
one objection.**

### The cause: the estate's Anthropic credit balance ran out

`[MEASURED 2026-09-04]` Every council seat call 11:21:00Z → 11:53:42Z failed with HTTP 400
*"Your credit balance is too low to access the Anthropic API"*. **Fleet-wide, not council-only:
ZERO successful LLM calls of any kind, any agent type, between 11:25 and 11:50**; 117 credit-400
failures in the 11:00 hour. Two OTHER lanes' rounds took the identical mechanical `revise` in the
same window (`3e9e8ce8` 11:22, `5de01fd3` 11:28) — **three unrelated submissions failing the same
way at the same time is the estate, not your plan.** Recovered unaided; last failure 11:56:47Z.

⚠ `gated_by_truncation` was **`false`**, which is true and irrelevant — the seats were not
truncated, they never ran. The discriminator is `llm_call_log`, not the gate's own metadata.

### The control, and it is decisive

Resubmitted on the **same** correlation with the plan **byte-identical** — no revision, because
there was nothing to revise. Round 2, 12:14:47Z: **`approved`, `decided_by: all reviewers approve`,
11 readable seats, 0 unreadable, 0 objections.** Same plan, healthy estate, opposite verdict. That
is what makes "round 1 was mechanical" a measurement rather than an excuse.

**`Council-Reviewed: 8bf83b59-cda4-43e1-af07-838ea10c1df7`** is now legitimate (verdict read).
Commit `75f7b843d` already carries `Council-Submitted:` for the same correlation and **`098` credits
it automatically** — forward-only forbids an amend and none is needed.

### What was filed out of it, so nobody re-derives it

- **`bugs_open/243`** — the episode, plus the finding that this file's own title names a **different
  failure mode** from the one now firing: `usage limit` (15,315 failures) **ended 2026-08-31**;
  `credit balance` (934) is what has fired since. **Mutually exclusive by day across 31 affected
  days — not one day shows both.** Full mode-A history: 04-10 (78) · 08-08 (20) · **08-25/26 (711,
  ~10 h)** · 09-02 (8) · 09-04 (117). Two of those (04-10, 09-02) were not in the file, and 04-10 is
  **earlier than the 07-31 it calls the first**.
- **`LANDMINES.md`** — an addendum for the **partial** outage (the existing entry covers the total
  one, which at least leaves an absence; the partial case writes a positive `revise` artefact).
  ⚠ **It also corrects a sibling entry filed the same day by the bugfix_257 lane**, which calls this
  *rare* on a base rate of 3-in-225-councils/7-days: all three landed inside **one 40-minute
  window**, so it is **clustered, not rare**, and the base rate predicts nothing. The actionable
  form is "check estate health", not "this is unlikely".

### ⚠ Three process notes the next session should not have to learn again

1. **An armed dispatch dies silently in an outage.** My `landmines-verify-dispatch.sh` run at 11:56
   failed 4× on credit-400. **A verifier that never reports is a health question, not a patience
   question.** Both entries re-fired: `c034492e`, `b68cf7b8`.
2. **Another session swept my LANDMINES edit into their commit** (`d9e5c0b71`, the 384 lane). Nothing
   lost — verified all 8 markers present in HEAD — and forward-only holds. Expected on this tree;
   noting it so the commit trail reads straight.
3. **`grep '^### '` is not a prior-art search.** I missed the bugfix_257 sibling entry because it is a
   `##`. And when I then checked whether my own edit had dropped references, a **single-line** grep
   reported three false absences on a hard-wrapped file — `tr '\n' ' '` first, exactly as
   `MEMORY.md` warns.

**So: nothing is owed on 442 or 773.** The open items are §5's three unchanged ones, plus the two
landmine verdicts now in flight.

### 7a. ⚠ The landmine verifier returned `NEEDS_HUMAN_REVIEW` on the council addendum — that is its INDEX, not a challenge to the entry

Do not read it as doubt. Its own note says why: it confirmed the core infrastructure
(`diagnosis_artifacts`, `kind='council_report'`, `metadata->>'decision'`, `llm_call_log`) in Go
code, and reports that `metadata->>'unreadable'`, `metadata->>'reviewers'`, the two `097`/`098`
shell scripts and `agent_type='council-gate'` **"live outside the .go-only index and could not be
checked"** — *"9 checks ran; 4 matched indexed code; **2 NOT ANSWERABLE by this index**; 3 matched
nothing in scope … the indexed corpus holds only: .go"*.

**So the verdict is structural: an entry whose footprint is jsonb keys, SQL or shell can never do
better than `NEEDS_HUMAN_REVIEW`, however true it is.** The unverifiable half here is precisely the
half measured first-hand today — `unreadable: 6 / reviewers: 3` on round 1 and `0 / 11` on round 2,
both read straight off `diagnosis_artifacts`.

⚠ And note its own staleness disclosure: it answered against **indexed commit `de1b9a58`, committed
2026-09-03 09:51Z — "the last pushed tip, not the present tree"**. A day behind, and it says so; a
reader who skips that line will take its silence about a symbol as absence.

### 7b. ⚖ POST-ROLL: 442 survived `v1.0.1361`, verified at the binary — and the roll changed nothing this lane depends on

`[VERIFIED 2026-09-04 16:02Z]` The fleet rolled to **`v1.0.1361`** (cut `06c0b18f2`), announced
mid-session by the "inter thread comms" lane. §4.2 is why this was re-checked rather than assumed.

**Waited for the rollout to FINISH before probing.** At 16:00 there was one new pod not ready and
two old ones still serving — probing then would have read `v1.0.1360` and reported a pass about the
wrong binary. `kubectl rollout status deployment/agent-chassis` returned
`successfully rolled out` at 16:02:15Z; only then the probe.

```
agent-chassis-6f699988d5-kzklg  v1.0.1361  ready     agent-chassis-6f699988d5-s4gg9  v1.0.1361  ready
  PRESENT meta_description_refused · PRESENT meta-description-repair
  PRESENT candidate_looks_internal (positive control) · absent ZZZ_cannot_exist_9f3a (negative control)
```

Both pods stamp **`06c0b18f2`** in `service_binary_capabilities`, and
`git merge-base --is-ancestor 776511e70 06c0b18f2` passes while the same test on `HEAD` fails — so
the control discriminates and the "did my fix ship" answer is a query, not an inference.

⚠ **`service_binary_capabilities` shows FOUR chassis rows for TWO live pods** — the deployment
cycled two replicasets (`5bbd648694` then `6f699988d5`) within a minute, all on the same commit.
Order by `started_at DESC` and match `pod_name` against `kubectl get pods`; a row is not evidence
that its pod is alive.

**The roll changed nothing this lane's claims rest on.** Per the peer's own caveat that a stamp
answers *"did my fix ship"* and not *"does this function still behave as my claim assumes"*:
`git log 239ab3626..06c0b18f2` over `cmd/scheduler/main.go`, `loop_expansion_handler.go`,
`loop_actions.go`, `save_page_meta_description_action.go` and
`save_page_meta_description_refusal_item.go` returns **0 commits each**, against a control of **16
Go commits in the same range** — so the zeros are real and not an empty query. §11b, §11d, the
LANDMINE and migration 773's premise all stand.

**773 is untouched** (config; a Go roll cannot reach it, checked anyway): `names_series t`,
`warns_bare t`, negative control `f`, and §4.1's `overwrite_existing` still **undeclared** at the
nested path.

**Acceptance evidence: still none, and still correctly so.** `meta_description_refused` rows: 0.
Eligible pages right now: 0. That is §11b's drained queue, not dormancy — read the two together.

### 7c. ⚠ CORRECTION to §5's adjacent list — the 440 RED is FIXED, and HEAD is red for two DIFFERENT reasons now

§5 (and §11e of `bugs_open/442`) says HEAD's `actions` package is red from `83407cd37`. **That is no
longer true and the note would send the next reader after a closed problem.**

`[VERIFIED 2026-09-04 ~16:2xZ, `scripts/verify-head-builds.sh --test ./platform/orchestration/actions/...`,
HEAD `26b09b978`]`

```
ok    …/platform/orchestration/actions            8.174s   <- GREEN, and this lane's tests live here
FAIL  …/actions/discovery_checks   TestEveryCheckProducedItemTypeIsClassified
FAIL  …/actions/queryresolve       TestSourceDependenciesMatchTheResolvers
```

The 440 lane fixed both of theirs in `50046041d` — *"fix the two package tests my `83407cd37` left
RED — **reported by two other lanes**, and my two pre-commit test commands could not between them
have caught it"* (plus `3c2f25fb9` restoring `finding_code_registry.json`'s own 1-space indent after
a `json.dump` reformatted all 76 entries). **The cross-lane report worked**; that is the argument for
contributing into another lane's bug file rather than filing a duplicate or patching their guard.

**The two failures now are other lanes' too, and neither is ours:**
`discovery_checks/verifier_coverage_test.go` was last touched by lanes 469 / 436 / 114, and
`queryresolve/page_image_sources_test.go` by 427 / 384. ⚠ **`discovery_checks` looks like ours and is
not** — it is where §5's "no verifier for `meta_description_refused`" work would land, so the
coincidence invites a wrong attribution. Checked rather than assumed: `grep -rn
'meta_description_refused' platform/orchestration/actions/discovery_checks/` returns **nothing**.

**So: `actions` is green, and whoever picks up §5's verifier item starts from a green parent package
but a red `discovery_checks` that is not theirs to fix.**
