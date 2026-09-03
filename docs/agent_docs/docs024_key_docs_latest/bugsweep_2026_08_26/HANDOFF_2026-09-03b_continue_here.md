# HANDOFF — bug sweep lane, 2026-09-03b

> ⚖ **UPDATED LATER THE SAME DAY — THE OWNER RULED ON §2: "yes, make them loud."**
> Route B is **BUILT**. Config half live (migration `734`), Go half committed `776511e70`,
> inert until the chassis rolls (a roll was expected within the hour of that commit —
> **check whether it landed before trusting anything below about the Go half**).
> §2's three inputs all held and one of them changed the design; see §7, added below.
> `bugs_open/442` §10 is the full account. **§2 is now history, not a decision.**

**Read this instead of `HANDOFF_2026-09-03_continue_here.md`.** That file is still correct about
338, 320, 404, 407 and the four owner rulings. What changed is **§5's leftover: `bugs_open/442`
has been worked.** Where it is now wrong I say so here rather than editing it — except for a
one-line banner added at its top, because a reader who acts on its §5 would redo shipped work.

---

## 0. STATE IN ONE TABLE

| item | state 2026-09-03b |
|---|---|
| **442** silent refusal | **candidate 3 SHIPPED, applied, live, council APPROVED** (`2ed33c57`, 10/10 seats). Candidates 1 & 2 **OPEN**, and 1 is now known to cover only half the mechanism |
| **442 route choice** | still the **owner's**, and now better informed — see §2 |
| 338 / 320 / 407 / 359 | unchanged from the 09-03 handoff |
| **404** r4 objections | still owed a read. Not this lane's |

---

## 1. WHAT SHIPPED, AND HOW TO CHECK IT WITHOUT TRUSTING ME

Migration `728_meta_description_backfill_result_message_names_the_copy_gates.sql` (+ `_ROLLBACK`),
commit `5a8728db9`, **applied and recorded in `schema_migrations`**. Config-only — no image, no
roll, live on apply. Council `2ed33c57-b49a-4b1b-ad1e-7e23ce6c477a`: `approved`, `decided_by:
all reviewers approve`, zero objections, not truncated.

The backfiller's `complete` step told a reader a refusal *"carries a named reason (empty_candidate
/ candidate_looks_internal / candidate_too_long / already_has_description)"*. There are **seven**.
The three missing were the copy-gate ones — the expensive refusals, the ones that leave a page
blank on an hourly schedule, the ones that need a person.

Verify at the artefact, with the must-be-absent control (this is the check, not the migration's
own verify block):

```sql
WITH m AS (SELECT default_config->'workflow'->'steps'->'complete'->'config'->>'result_message' AS msg
           FROM agent_definitions WHERE type='meta-description-backfiller' AND is_active
             AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL)
SELECT r AS reason, (position(r in (SELECT msg FROM m)) > 0) AS named
FROM unnest(ARRAY['empty_candidate','candidate_looks_internal','candidate_too_long',
                  'already_has_description','voice_tell','banned_claim','voice_gate_unreadable',
                  'THIS_REASON_DOES_NOT_EXIST']) r;
-- expect: seven t, and THIS_REASON_DOES_NOT_EXIST -> f
```

⚠ **The message deliberately does NOT just list seven.** A seven-item list is the same defect one
birthday later — `442` §4 makes exactly that point about enumerations rotting by addition with no
equivalent of the 2026-08-22 date-your-counts rule. It lists the seven, **splits them by what they
ask of a reader**, and says it is a copy, where the authoritative set lives, and that finding it
takes two greps. **Do not "tidy" it shorter.**

---

## 2. ⚖ THE OPEN OWNER QUESTION IS UNCHANGED, BUT THE EVIDENCE UNDER IT HAS MOVED

The 09-03 handoff §0.4 put Route A (make the surface honest) and Route B (make the refusal loud)
to the owner. **Route A is now done.** Route B is still his call, and three things measured today
should be in front of him when he takes it:

1. **The volume objection HOLDS.** `voice_tells` work items, counting the archive as well as the
   live table: **66** parked in `needs_human_review` against **5** ever complete (3 live + 2
   archived), nothing filed since 2026-08-27. Filing gate refusals there relocates the silence.
2. **Route B as specified covers HALF the mechanism** (`442` §9d, new). The writer step's prompt
   *instructs* omission — *"omit that page entirely rather than inventing one. Returning fewer
   entries than you were given is a correct answer"* — and nothing compares
   `pages_missing_meta.count` with the length of `written.result.descriptions`. A page the model
   drops leaves **less** trace than a gate refusal: no `save_result` at all, because the loop never
   reaches the action. Candidate 1 files from **inside** that action, so it cannot see this path.
   A **fifth candidate** — compare the two integers already sitting in `collected_data` — covers
   both and files nothing into a queue.
3. **Nothing is broken today, and that is measured, not assumed.** Active pages: **37** blank
   (avg **8** chars of visible text), **0** clearing the backfiller's `page_visible_text_len > 200`
   gate; **1,171** described (avg **4,381**), **1,137** clearing it. No page is both eligible and
   blank. Both silent paths are **latent**. That is an argument about urgency, not soundness.

---

## 3. ⚠ TWO CORRECTIONS TO THINGS THIS LANE HAD WRITTEN DOWN

### 3a. `442` §2 blamed a zero on retention, and §6 turned that into an instruction
It said `orchestration_states` *"returns zero rows carrying a `save_result.reason` fleet-wide —
the rows age out"*, so §6 said *"do not verify at `orchestration_states`"*.

`[MEASURED 2026-09-03]` the table holds **9,277** rows over a **~26-hour** window (oldest
`2026-09-02 09:41`). Five backfiller runs survive in it and **all five carry a `save_result`** —
5/5 `updated:true`, **0/5** carrying a `reason`. The action only writes `reason` when it refuses,
so a window with no refusal returns zero either way. **Two sufficient causes, one measured**, and
the wrong one became guidance telling the next session away from the evidence.

**The demand control is one column: is `save_result` present AT ALL.** Corrected in `442` §9b;
landmine written, footprinted on `orchestration_states` / `save_result` / `__step_error`, because
there is no tell — both answers are `(0 rows)`.

### 3b. Line numbers are a citation, not evidence — third instance, caught by a seat
The submission justified the two-greps asymmetry with line numbers. `prior_art_librarian` filed a
`missing` item on an otherwise unanimous approval, and was right. Answered in `442` §9g at
committed HEAD with the output and a must-be-empty control — and the answer is better than the
claim was: the gate path writes `"reason":  reason`, a **variable**, so a grep shaped for
`"reason": "<literal>"` matches exactly the already-documented set and reports a **complete**
list. **Auditing a vocabulary by grepping its MEMBERS instead of its WRITERS** is the general
defect. `WRONG_CALLS.md`, three rows now on this fault at three altitudes.

---

## 4. WHAT IS LEFT ON THIS LANE

**Owed, small:**
- **Read 404's three r4 objections** — still open, still not this lane's, still worth it.
- **`442` candidates 1/2 + the new fifth** — blocked on the owner's route choice (§2). Do **not**
  build Route B without answering *who reads that queue*; §2.1 is the measurement that says why.

**Adjacent, none of them this lane's** (carried forward unverified today except where noted):
- `platform/livespec` RED at committed HEAD (405 lane) — unchanged, 9 days.
- `TestNoHandSpelledTombstonePredicate` RED at committed HEAD (114/IMG-077 lane).
- `_RELOCK` unclassified migration suffix — **still WARNed** by the 097 trigger today.
- `WII-035` duplicate row id in the concept index.

---

## 5. FIRST COMMANDS

```bash
cd /home/ant/projects/agentchassis
git log --oneline -20

# 1. Is 728 live? (the §1 query above, with its control)
# 2. Is the damage still zero? — run BOTH rows; the described-pages row is the demand control
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT CASE WHEN COALESCE(meta_description,'')='' THEN 'BLANK' ELSE 'has description' END AS grp,
       count(*), round(avg(page_visible_text_len(id))) AS avg_len,
       count(*) FILTER (WHERE page_visible_text_len(id) > 200) AS eligible
FROM pages WHERE status='active' GROUP BY 1;"
# 3. The queue behind Route B, live AND archive — the archive half is not optional
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT 'live' src, status, count(*) FROM site_work_items WHERE item_type='voice_tells' GROUP BY 1,2
UNION ALL SELECT 'archive', status, count(*) FROM site_work_items_archive WHERE item_type='voice_tells' GROUP BY 1,2;"
```

⚠ **`collected_data ? '<key>'` on `orchestration_states` is a sequential scan and takes minutes.**
Run it in the background with `SET statement_timeout` raised. A timeout is not evidence the rows
are gone — that is the same mistake as §3a wearing a third hat.

---

## 6. WHAT SHIPPED THIS SESSION, FOR THE RECORD

`5a8728db9` migration 728 (+ROLLBACK) · `a3e092368` 442 §9 CONTRIB · `b811964f7` WRONG_CALLS ×2 +
LANDMINES ×1 · `652f32f74` 016b §9 extension · `699b6e439` NOTES + the lane's first
`README_where_we_are.md` · `75fa195a2` 442 §9g verdict · `6815e3cb0` WRONG_CALLS third instance.
Council `2ed33c57-b49a-4b1b-ad1e-7e23ce6c477a` **APPROVED**. Landmine verifier dispatched
(`28364a1b`), verdict not awaited.

---

## 7. ⚖ ADDED 2026-09-03b — the ruling, what was built, and the two things that must not be lost

### 7.1 The measurement that changed the design — the transferable part
§2.1 said the volume objection held. Before building I asked the question behind it: **is that
queue a graveyard because of people, or because of the shape of the row?**

`[MEASURED 2026-09-03, site_work_items UNION site_work_items_archive]` items **WITH** a
`handler_agent`: **56,315 / 83% complete**. Items with **NO** handler: **6,699 / 17% complete**,
989 parked. `voice_tells`: **69 rows, every one handler-less.**

**It is the row.** A flag-only `needs_human_review` item is not "loud" — it IS the 17%. Any lane
about to file findings should run that pair before choosing where to file them.

### 7.2 What was built
`save_page_meta_description` files `meta_description_refused` at **`meta-description-repair`**
(migration `734`), which re-asks **with the refusal quoted back** and saves through the **same
gated action**. Second refusal parks at `needs_human_review` carrying both attempts — deliberately,
because `fail_work_item`'s ladder brands a two-striker `unresolved`, which is silent again.
Registered **SEO-008** with its index row. Council `76288ff9-3cde-46e6-b65a-22564fac8f6d`.

### 7.3 ⚠ TWO THINGS THE NEXT SESSION MUST NOT LOSE
1. **`overwrite_existing` is undeclared on that agent and MUST STAY UNDECLARED.**
   `102_coverage_ratchet.txt` line 105 and `bugs_open/320` §15 record that authority — rewriting
   published copy on an automated finding — as **explicitly withheld by the owner on 2026-08-21**,
   granted once for a one-off dispatch and never armed on any agent. Adding that one key is a
   **one-line, innocuous-looking config edit that crosses an owner ruling.** Verified undeclared
   2026-09-03 with a control; re-verify before trusting it:
   ```sql
   SELECT default_config#>'{workflow,steps,save_description,config}' ? 'overwrite_existing'
   FROM agent_definitions WHERE type='meta-description-repair' AND is_active AND deleted_at IS NULL;
   -- must be f
   ```
2. **"No rows" is not "it does not work", and needs a demand control.** `[MEASURED 2026-09-03]`
   ZERO active pages are both blank and clearing the backfiller's `> 200` gate, so **nothing can
   have filed**. Run this before concluding anything:
   ```sql
   SELECT count(*) FROM pages WHERE status='active'
     AND COALESCE(meta_description,'')='' AND page_visible_text_len(id) > 200;
   ```

### 7.3b ⚖ COUNCIL ROUND 1 = REVISE (gated, guardian HIGH). Round 2 resubmitted, verdict OWED A READ
The gating objection was correct: the edit list omitted the shared action file the plan admitted
editing. **And one objection changed the code** — `bug_historian` [medium] found that the filing's
own six failure branches were bare `logger.Warn`s, so a failed write of the LOUD RECORD put the
refusal back to being a log line, *"the exact defect this plan exists to close, now one hop deeper
and harder to notice because the design narrative says it's already fixed"*. The filing now
returns `(filed, fileError)` and the action reports both (`356196fe9`). Two asserted claims were
measured in response — callers of the action are **exactly two, both this lane's** (control: the
same query for `ensure_site_record` returns 50), and there is **no shared refusal-parking helper**
to reuse (all three prior instances are package-private with no common signature). Full account:
`bugs_open/442` §10g.
⚠ **Both `776511e70` (r1) and `356196fe9` (r2) are ancestors of HEAD**, and the chassis was still
on `v1.0.1356` (pods 08:57Z) as of this writing — so the pending build carries BOTH. Verify at the
artefact, not the clock: `git merge-base --is-ancestor 356196fe9 <the chassis build-provenance stamp>`.

### 7.4 Still open on 442, so the file stays OPEN
- **No verifier** for `meta_description_refused` — five build guards plus a live
  claimed-item-timeout migration merged with other lanes' amendments. Named follow-up.
- **`voice_gate_unreadable` is still silent**, correctly for a rewrite handler (wrong actor) but
  it still leaves a page blank for ever. Needs an operational surface.
- **§9d's second silent path** — the writer omitting a page — is unbuilt. The fifth candidate
  (compare `pages_missing_meta.count` with the length of `written.result.descriptions`) covers it.
- **Read the `76288ff9` verdict**, and note it did **not** see §7.3.1: I found the withheld-authority
  ruling after submitting.

### 7.5 ⚠ The working tree does not compile, and it is not this lane's
Another session has an **untracked** `criteria_value_assertions.go` whose `itoa` collides with the
one in `provocation_gate_action_test.go` at HEAD. `go test ./platform/orchestration/actions/`
fails in the tree and proves nothing about HEAD. Everything this lane ran went through
`scripts/verify-head-builds.sh --test`, and **HEAD 7289af111 was green** including the new code.
