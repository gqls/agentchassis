# 298 — `internal-linker` picks link targets from at most 15 candidates, alphabetically — and whether it has ever reached that step is UNMEASURED

**Filed 2026-08-17** by the `bugfix_275_silent_row_caps` lane. **Same defect class as
`bugs_open/275`**, third instance. **Deliberately filed with a weaker severity claim than
`bugs_open/297`**, because the reachability evidence does not support a stronger one — see §Reachability.

## The defect

`internal-linker`'s `load_candidate_pages` step ends:

```sql
... FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL
WHERE p.site_id = $1 AND p.name != $2 AND p.status = 'active'
  AND p.page_type IN ('content','service','landing','tool')
GROUP BY p.name, p.url, p.title, p.page_type
HAVING COUNT(pc.id) > 0
ORDER BY p.name LIMIT 15
```

Those rows are the candidate set the `plan_links` LLM step chooses internal link targets from. **The
cut is `ORDER BY p.name` — alphabetical** — so which pages a site may link to is decided by page-name
ordering, not by relevance. Silent by construction: the model returns plausible links either way.

## Measured 2026-08-17 (live `clients_db`), with the query's OWN predicate

⚠ **Including `HAVING COUNT(pc.id) > 0`.** A first pass that omitted the HAVING over-counted; the
figures below use the real predicate.

| | value |
|---|---|
| cap | **15** |
| sites with any candidates | 24 |
| **sites where candidates EXCEED the cap** | **8 of 24** |
| median site | **12** — *under the cap; most sites are unaffected* |
| worst site | **68** |

So this bites on **a third of sites**, and at the worst one the model chooses from **15 of 68 (22%)**.
Unlike `bugs_open/297` (19 of 24 sites), the median site here is fine.

## ✅ REACHABILITY — ANSWERED 2026-08-18, and it inverts this ticket's priority

**Keep reading past this block for the original §Reachability text; it is left intact because its
reasoning was sound and its conclusion was not.**

Two of its claims do not survive measurement:

1. **"ZERO rows, all history" is true of `llm_call_log` ONLY.** The agent has **13
   `orchestration_states` rows** and 82 `site_work_items`, most recent run **2026-08-18 07:33Z**. It
   is dispatched and it completes. Reading one channel is the trap this lane's own handoff lists as
   landmine #2.
2. **"The detector will answer §Reachability by itself" is false.** LCO-009 writes a log line, and
   the chassis log retains 15–90 s (measured 2026-08-18). It cannot answer a question about the past.

**What the durable record says.** `load_candidate_pages` writes its result to
`collected_data->'candidate_pages'`, which survives rolls. Two runs reached that step:

| run | candidates loaded | `current_step` at completion |
|---|---|---|
| 2026-08-17 22:22:33Z | 7 | `complete_no_candidates` |
| 2026-08-18 01:01:46Z | **15 — exactly the cap** | `complete_no_candidates` |

**Both runs ended reporting no candidates, including the one holding fifteen of them.**

### Why: the branch after the fetch is unsatisfiable as configured

- `load_candidate_pages` declares `"output_format": "array"`. `QueryDatabaseAction` returns the bare
  slice for that format (`platform/orchestration/actions/database_actions.go:129`); the `count` key
  exists **only** in the `object` branch.
- `check_candidates` is `"condition": "candidate_pages.count > 0"`, `else_step: complete_no_candidates`.
- `resolveFieldValue` (`platform/orchestration/actions/conditional_branch_action.go:320`) tries five
  strategies. Strategy 5, the recursive one, requires the base to be a `map[string]interface{}` — an
  array is not — so all five fail and it returns `nil`.
- The numeric arm then fails `datahelpers.ToFloat64(nil)`, logs `Numeric comparison: left side is not
  numeric`, and **returns false**. Every run takes the `else`.

**So `plan_links` cannot execute, which is precisely why `llm_call_log` is empty — the empty table was
a symptom of this, not evidence of dormancy.** The cap in this ticket has never shaped a link decision
because **the internal linker has never made a link.**

### What this does to the ticket

- The **cap** (this file's subject) is **latent**, exactly as the original §Reachability suspected it
  might be — but for a reason that is worse, not better, than "the agent is idle".
- **Fix candidate 1 ("resolve reachability first") is DONE.** Do not size the cap fix until the branch
  is fixed: repairing the cap alone changes nothing observable, and repairing the branch alone
  immediately makes the cap live on the 8 sites that exceed it.
- The dead branch is a **different defect** from the cap and is being filed separately rather than
  folded in here.

**Filing basis for THIS block (owner ruling 2026-07-31):** put through the diagnosis loop before being
asserted — `RUN_CORRELATION_ID=c4aa3559-86b1-4356-a28b-c71dfa661465`, filed 2026-08-18 18:38Z.
**Verdict pending at the time of writing**; whoever reads this next should check it and record the
result here. The measurements above are first-hand and reproducible by the queries in
`docs/agent_docs/docs024_key_docs_latest/bugfix_275_silent_row_caps/RUNBOOK_275_silent_row_caps.md`.

---

## ⚠ Reachability — the part that keeps this ticket honest (ORIGINAL, 2026-08-17 — superseded above)

**`llm_call_log` for `agent_type='internal-linker'`: ZERO rows, all history.** The step that consumes
these candidates (`plan_links`, an `execute_llm_prompt`) has no logged call, ever. So although the cap
is structurally real, **whether it has actually shaped a single link decision in production is
UNMEASURED.**

What IS established:

- `internal-linker` is an **active** agent definition (created 2026-04-12) with a live workflow whose
  steps are `load_target_page → check_target_found → load_candidate_pages → check_candidates →
  plan_links → …`.
- **69 work items name it as `handler_agent`, the most recent dated 2026-08-17** (today), 38 of them
  `complete`.
- Of those 38: **13 have a `target_page` count > 0**, i.e. they passed `check_target_found` and would
  have proceeded to `load_candidate_pages`. **15 have `target_page` count = 0** and exited early. 8
  carry only the spawn record and cannot be read (the known `result`-is-the-spawn-record trap —
  `bugs_open/287`).

**So the step is reachable and has probably been reached 13 times, yet no LLM call is logged.** That
gap is not resolved here, and a fixing thread should resolve it BEFORE sizing this bug: if `plan_links`
never executes, the cap is latent and this is a low-priority cleanup; if it executes and logs
elsewhere, it is live and equal to `bugs_open/297` in kind.

**The check:** fire one linker run against a site with >15 candidates (8 qualify) and look for a
`plan_links` row in `llm_call_log`. Absence there, with a completed run, is itself a finding.

## Adjacent finding, recorded not chased

**15 of 38 completed `internal-linker` items found NO target page** (`target_page.count = 0`) and
completed anyway. A run that completes having linked nothing is indistinguishable, from `status`, from
one that linked successfully — a `complete` work item is not a repaired artefact. Not this bug's
subject; recorded so it is not lost, and a candidate for its own ticket if someone finds it biting.

## Fix candidates, ordered by what closes the door

1. **Resolve reachability first** (§ above). Sizing a fix before knowing whether the step runs is how
   effort goes to the wrong place.
2. **Then: bound the payload, not the row count** — `content_sample` is already `LEFT(..., 800)`, so
   the per-row cost is bounded and 68 rows is ~54kB worst case; measure before choosing. The shape that
   worked for 275 is: find the dominant column, bound it, drop the row cap. **Mark any truncation** —
   an unmarked cut is the same defect one level down (migration 446 is the worked remedy).
3. **Or rank before capping** — link candidates have a natural relevance order (shared entities, page
   type, existing link graph) and `p.name` is not it.
4. **Do NOT just raise the number.** 15 → 30 still cuts the worst site and re-files this in a month.

## How to verify a fix

- Census returns **0 sites over the cap**, or no cap.
- Disconfirming pair: on a site with >15 candidates, take a page sorting past position 15 by name and
  confirm it appears in `plan_links`' rendered prompt after the fix and cannot before.
- **Both require `plan_links` to actually run** — which is §Reachability, and is why that comes first.

## Filing basis (owner ruling 2026-07-31)

**No `090` run; substitution stated plainly.** No new mechanism is asserted — it is `bugs_open/275`'s,
council-approved (corr `b684a399`) and registered as LCO-009. What is new is arithmetic on live data,
every figure reproducible by one query, **plus an explicitly UNMEASURED reachability question that this
file declines to guess at**. Grepped both bug directories before filing; nothing covers this step.

## Related

`bugs_open/275` (fixed, approved) · `bugs_open/297` (the sibling instance, live and worse) · register
**LCO-009** (~~the detector will answer §Reachability by itself once the chassis rolls~~ — **it cannot;
its log line does not survive 90 seconds. Corrected 2026-08-18; the durable channel is
`orchestration_states.collected_data`**) · `bugs_open/287` (why 8 of the 38 results are unreadable) ·
`bugs_open/242` (same class, render audit).
