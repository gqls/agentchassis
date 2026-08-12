# HANDOFF — bug 122 lane. START HERE. Written 2026-08-12 (evening).

Supersedes `HANDOFF_2026-08-12b_continue_here.md` for **state**: its §3 task is **DONE and
committed** (`5639a1103`). That file stays the reference for **why the design is shaped this way**
(§3.1's two constraints, §3.3's four hazards) — all of it still true, and §3.2 has **one correction**
recorded below that you must read before touching the code. `HANDOFF_2026-08-12` remains the
reference for the owner's two decisions and the 📅 2026-08-16 action, **which is unchanged and
still owed**.

**Nothing in this lane is on fire.** The queue is empty, no site is locked out, the 226 are parked
and stable, and the code that just shipped **does nothing at all** until two services roll.

---

## 1. State, measured 2026-08-12 evening

| thing | state |
|---|---|
| the retraction | **BUILT + COMMITTED `5639a1103`**, 8 files, HEAD verified to build and test green in a clean `git archive` tree |
| council | **SUBMITTED, verdict UNREAD** — corr `a43b63d6-da35-4136-9471-88ec6ace799a`, committed with `Council-Submitted:` |
| is it live? | **NO, and a chassis-only roll will not make it live.** Needs BOTH `agent-chassis` (producer) and `browser-runner-adapter` (adapter). A fresh chassis was being built as this was written |
| 226 `contrast_failure` items | still **parked** `deferred`, `parked_by='migration_389'`, `max(attempt_count)=0`, **0 ever completed** — unchanged, and re-measured this session |
| register | `WII-016` added (work-item-integrity), index row added. Nothing to drop from `102_coverage_ratchet.txt` |
| `LANDMINES.md` | one entry added + synced: *a parked (`deferred`) work item is retractable* |
| `site-render-audit-rotation` | enabled, weekly per site, zero LLM spend — this is what will drive the retraction |
| `site-discovery-rotation-quality` | enabled, **inert until 2026-08-16 09:49Z** — correct and waiting, do not "fix" it |

---

## 2. ⚠ THE ONE CORRECTION TO `12b` §3.2 — read this before editing the retraction

`12b` §3.2 step 1 said to build the still-failing set from *"exactly the keys already computed at
`:266`"*. **That is wrong, it was not followed, and the code does the other thing.**

`:266` is reached only by findings that survive two filters above it: the **locked-component skip**
(we can see the fault but a human has locked that component, so we never file it) and the
**`max_items` cap** (default 60, worst-ratio-first, remainder dropped). A finding removed by either
was **measured and is still failing**. Building the retraction set from the FILED items reads "not
filed" as "fixed" and closes those tickets — the false completion the park exists to prevent.

So the set is built from `payload.Contrast` **before every filter that decides what to file**, and
`over_image` approximations count as still-observed (an unknown backdrop is "I could not tell", not
evidence of health). **If you change one thing in this function, do not change that.** It is pinned
by `TestWriteRenderAuditFindings_MeasuredButUnfiledFindingsAreNotRetracted`, and reverting it to
the handoff's spec makes that test RED — which is how it was proven rather than argued.

---

## 3. NEXT, in order. Do not skip to the unpark.

**(a) Read the council verdict.** `a43b63d6-da35-4136-9471-88ec6ace799a`. The code is already on the
shared branch, so a REVISE or REJECTED must be answered forward-only, in a new commit.
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'a43b63d6-da35-4136-9471-88ec6ace799a'
 ORDER BY created_at DESC LIMIT 1;                        -- COMPLETED, not EXECUTING_STEP
SELECT created_at, body FROM doc_notes WHERE categories ? 'council-gate'
 AND body LIKE '%a43b63d6%' ORDER BY created_at DESC LIMIT 3;
```
Two things the submission asks the council to rule on explicitly, so read for these: **(1)** that
retraction closes PARKED rows, and **(2)** the §2 correction above. If the verdict turns APPROVED,
`098` credits the commit automatically — **do not write a `Council-Reviewed:` trailer by hand.**

**(b) Roll BOTH services, and check the stamp PER SERVICE.** Go is inert until rebuilt and rolled;
`make build-*` builds from committed HEAD.
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
kubectl -n ai-persona-system logs -l app=browser-runner-adapter --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 5639a1103 <the stamp>   # "did my fix ship?" is a QUERY
```
An empty grep means **"not in range"** (it is a startup line and scrolls), never "unstamped" — fall
back to the binary probe with a known-present and a known-absent control. A release can straddle
sessions' commits, so read the stamp of the service you actually mean (`bugs_open/249`).

**(c) Watch ONE weekly audit, then grade it at the artefact.** The first retraction will come from
a site's own `site-render-audit-rotation` fire. Grade the ROWS, never the run's status:
```sql
-- did anything retract, and did the mechanism say so?
SELECT item_key, status, result->>'resolved_by', result->>'reason', result->>'resolved_at'
FROM site_work_items WHERE item_type='contrast_failure' AND result ? 'resolved_at'
ORDER BY result->>'resolved_at' DESC LIMIT 20;
-- the park, which will now drain on its own
SELECT status, count(*), count(*) FILTER (WHERE result ? 'resolved_at') AS retracted
FROM site_work_items WHERE item_type='contrast_failure' GROUP BY status;
```
The action's own result carries `retracted`, `retracted_parked`, `retraction_scope_pages`, and
`retraction_unavailable:true` when the adapter is too old to say what it measured. **If you see
`retraction_unavailable`, the browser-runner-adapter has not rolled** — that is the version-skew
branch working, not a bug.

**(d) THEN unpark, and only then.** One UPDATE at the foot of migration `389`, predicated on
`spec->>'parked_by' = 'migration_389'`. Row-level backup at
`scratchpad/backups/backup_park_contrast_failure_20260811.tsv`. Expect the remainder to be
**smaller than 226** by the time you get here — that is the point, not a discrepancy.

**📅 2026-08-16 — unchanged and still owed.** Price the discovery-rotation ramp: calls AND tokens
(baseline ~248k input tok/h idle; the driven sweep was ~806k/h). Queries at the foot of
`sql_for_agents/395_enable_quality_discovery_rotation_slow_ramp.sql`.

---

## 4. Also still open in this lane

| | item | status |
|---|---|---|
| 1 | `bugs_open/212` §8 — component-painted grounds (~24 failures) | **Owner's.** Architecture, not a bug patch. Unchanged |
| 2 | `bugs_open/242` — the silent 25-page cap | **DONE, live `v1.0.1288`.** It was the PRECONDITION for this session's work: it shipped the page COUNTS, this shipped the IDENTITIES |
| 3 | `dark_section_audit` | Straddles the same hole with the same unsound reason, which is now the ONLY entry still making that argument — the cross-reference in `verifier_coverage_test.go` was corrected without deciding it. **`bugs_open/213`'s call or the owner's, not ours** |
| 4 | Free cross-check | if a lane re-renders robot-hands `/selection-guide.html`, the audit filed `info-card-grid__card-link` + `__eyebrow` failures and migration `368` should close both. **Grade at the next audit, never at the item status** |

---

## 5. Standing traps this lane has paid for

- **Grade per selector, never by fleet total.** It rose 109 → 112 while every targeted failure closed.
- **A filed count is not a found count.** "34 findings" was 171 firm — 111 dropped by a cap. And
  **226 is a FLOOR, not a census**: the audit was capped at 25 pages until `v1.0.1288`.
- **Read the selection before asserting it excludes your rows.**
- **A `file:line` in a handoff is a pointer, not a quotation** — this fired TWICE in two days now.
  Yesterday on `:171` (stale line), today on `:266` (right line, wrong SET — see §2). Open the file.
- **A pathspec commit still takes a same-file passenger.** Expect it; verify at HEAD, not the tree.
- **`pages.sections` is an array of plain strings**; an object-shaped census returns 0 rows silently.
- **Never run `run-migrations.sh --apply` on this tree.**
- **A call count does not price an LLM loop.** Threshold the expensive unit.
- **Put a row COUNT you could be wrong about in your post-check.**
- **NEW — a passing mock cannot assert a negative; MUTATE.** All four retraction guards were proven
  by reverting each to its plausible-but-wrong form and confirming the matching test went RED. The
  first mutation IS the handoff's own spec, which is how §2 was established rather than argued.
- **NEW — `deferred` is in NEITHER status list.** Not terminal (holds its dedup slot), not closed
  (any retraction may close it). "Parked" does not mean "nothing will touch this". See LANDMINES.
