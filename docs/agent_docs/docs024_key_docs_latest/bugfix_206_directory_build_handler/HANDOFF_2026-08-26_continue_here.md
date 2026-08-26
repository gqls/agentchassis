# HANDOFF 2026-08-26 — continue here (`bugs_open/206`)

**Supersedes `HANDOFF_2026-08-25b_continue_here.md`.** Keep it; accurate for its own day.
Read `bugs_open/206` **bottom-up** — its last five sections are 2026-08-25/26, in order.

---

## State in one paragraph

**All the code this bug needs is shipped, reviewed and LIVE.** Two council approvals, one routing
authority instead of three copies, `section-index` routed, the gap row aligned, a permanent
provenance stamp, and 9 tests across two doors that had none. **Nothing is outstanding in code.**
The bug stays OPEN for exactly one reason: **nobody has yet watched the fixed code route a page.**
That needs a greenfield build carrying an `entity-directory` or `entity-page` page — the only two
types where the fix and the bug behave differently. **§5 explains why I recommend against closing on
"it's live", and it is the owner's call.**

---

## 1. Everything DONE — do not redo

| commit | what | live? |
|---|---|---|
| `d1aa231aa`, `200d54bdf` | `builderForPageType` created; `reconcile_site_plan` calls it. Council **APPROVED** r6 (`52dbd067`). | ✅ `v1.0.1334` |
| `efec862f4` | The swap: `WriteBuildItemsAction`'s inline maps deleted; `section-index` → `directory-build-handler` in the same commit; `capability_gap` `handler_agent` → EMPTY. Council **APPROVED r1** (`b92e624d`), 13 reviewers. | ✅ `v1.0.1339` |
| `0777eb297` | Two coverage gaps an adversarial review found — one the swap itself created. | ✅ `v1.0.1339` |
| `1887a116b` | **Routing provenance**: both doors stamp `spec.handler`. Council **APPROVED** (`9ff151d6`), 9 reviewers. | ✅ **`v1.0.1341`** |
| `3dda3b191` | Answers that round's 3 objections **by measurement**; declares `spec.handler` in **BLD-027** (owed by the 2026-08-11 nested-field ruling); fixes a comment `1887a116b` had falsified. | ✅ `v1.0.1341` |

`098` credits all of them; **MISMATCH: 0**.

**Live verified 2026-08-26, four ways** (the startup log line had scrolled out of *both* pods even at
`--tail=200000`): image label `2fb40a960`, ancestry of all four commits, a **control** (two commits
made after the build correctly NOT ancestors), and a known-value probe on the running binary
confirming the label describes what is actually running. RUNBOOK §7c has the sequence.

**Docs**: `bugs_open/206` (5 sections), RUNBOOK §7/§7a/§7b/§7c/§7d, NOTES (missteps 1–6), README
(4 owner entries), **BLD-027** + index row, `LANDMINES` (1 new entry, 3 corrections/extensions),
`WRONG_CALLS` (4 entries), 3 memory lessons.

---

## 2. What is LEFT

### (a) THE closure proof — the only open item, and it is not schedulable by this lane

**Precondition, which outranks everything:** the build must carry an `entity-directory` or
`entity-page` page. `section-index` is **not** sufficient — every other type routes to
`page-build-handler` under *both* the bug and the fix, so the mint cannot discriminate. This is not
theoretical: `homegarden.uk` (2026-08-25) had 17 `section-index` pages and settled nothing.

```sql
-- run this against the plan BEFORE treating a build as the closure artefact
SELECT page_type, count(*) FROM pages
WHERE site_id=(SELECT id FROM sites WHERE domain='<domain>') GROUP BY 1 ORDER BY 2 DESC;
```

Then **RUNBOOK §7d** — the permanent gate, now armed:
```sql
  AND swi.spec->>'handler' = swi.handler_agent   -- the column still holds what the emit wrote
```
`[MEASURED 2026-08-26]` **0 of 231** `reconcile_site_plan` rows carry `spec.handler` yet, because no
reconcile has run since the roll. **That zero is expected and is also indistinguishable from a broken
stamp** — so the first stamped row deserves both §7d and §7's older gates, agreeing, before anyone
calls it proof.

### (b) The nine parked rows — ⚠ **and the count was wrong until today**

`[MEASURED 2026-08-26, unfiltered]` **9**, not the five/six this lane kept quoting — every earlier
figure came from a query filtered to the four domains I already knew about, so **three were never in
any of my counts**.

| domain | page | type | status | note |
|---|---|---|---|---|
| garden-tools.uk | brand-directory-index | entity-directory | needs_human_review | ⛔ **DO NOT TOUCH** |
| garden-tools.uk | brand-profile | entity-page | needs_human_review | ⛔ **DO NOT TOUCH** |
| garden-tools.uk | buying-guides-index | section-index | needs_human_review | ⛔ **DO NOT TOUCH** |
| dartsonline.com | brand-detail | entity-page | needs_human_review | not covered by the ruling |
| loanzy.uk | guides-index | section-index | needs_human_review | not covered by the ruling |
| adversecreditmortgage.co.uk | blog-index | section-index | **unresolved** | **unreachable** — §2(c) |
| mortgagecalculator.co.uk | about-index | section-index | deferred | newly counted |
| mortgagecalculator.co.uk | contact-index | section-index | deferred | newly counted |
| robot-hands.com | learning-center-index | section-index | needs_human_review | newly counted |

⛔ **`garden-tools.uk`: NOTHING IS TO BE CLEARED (owner ruling, 2026-08-25).** The authorisation was
**retracted** — `CONTRIB_2026-08-25_from_loanzy_lane_the_owner_retracted_the_parked_row_authorisation.md`
in this directory. It is an unassisted greenfield build four lanes measure against.

⚠ **A hand re-triage proves NOTHING about this fix, and is a known false-PASS path** — it sets
`handler_agent`, which is the column the closure test reads. **This is why `spec.handler` exists.**

**None of these can self-heal**: a parked row holds its own `item_key`, so reconcile skips the page
as queued, and nothing schedules reconcile anyway.

### (c) Follow-ups — named, evidenced, other lanes'

- **Residual (b), the bigger and better fix**: `page-build-handler` cannot fill a *missing* layout
  for **any** type — `ensure_page_section_layout` lives only in `directory-build-handler`'s workflow.
  `blog-post`/`blog-index` casualties on four sites. Right shape: make the layout-ensuring step
  reachable from the generic path, **not** route more types to `directory-build-handler`. Own round.
- **The `unresolved` divergence**: `loadOpenPageItems` (`:713`) treats `unresolved` as OPEN so
  reconcile skips the page for ever, while `idx_swi_dedup` doesn't cover it and both claim gates
  exclude it. Nothing can free such a row. One live casualty. Changes a dedup contract → own round.
- **`emitted++` without `RowsAffected`** (`reconcile_site_plan_action.go:484`) while the gap arm four
  lines up reads it and cites `bugs_open/091`. One line.
- **`needs_directory` is write-only**: 0 rows ever, 0 Go readers outside `builder_routing.go`, 0 live
  configs. Retiring it touches `create_tool_cross_link_items.go:263`'s gate.

---

## 3. The traps this lane actually fell into — read before measuring anything

Each produced a confident, plausible, wrong answer. Full accounts in `WRONG_CALLS.md` and NOTES.

1. **`handler_agent` has two causes** — the fix writes it, and so does the documented operator
   repair. All three rows fleet-wide that would have passed the closure test were **hand repairs**.
   → that is what `spec.handler` fixes.
2. **A test can pass under mutation for more than one reason.** Fixing the first reason did not make
   it fail. *If a mutation still passes after you fix why it passed, there was another reason.*
3. **A fixture producing zero rows passes every assertion about those rows** (`nav_order: nil` failed
   a scan; the action logged "no per-page builds needed" and returned success).
4. **Three broken deploy probes in one command**, with a 40-zeros "control" that came back PRESENT.
5. **I warned a peer lane about a mechanism without checking its precondition on their data**, and
   they had already adopted it into their acceptance guide. *A check that stops an investigation is
   more dangerous than one that starts a wrong one.*
6. **A count you kept is not a census** — "five parked rows" was a filtered query, repeated for days;
   the real number is nine.
7. **Three council seats caught me asserting an absence rather than measuring it** ("nothing reads
   `spec.handler`"). If you write "nothing reads X" in a submission, **attach the query.**

**The meta-lesson, which cost most:** a correction is not exempt from the discipline it enforces. The
measurement that caught trap 1 computed the very discriminator its own first fix omitted.

---

## 4. Cross-lane

- `bugs_open/381` is a good correspondent — captured mint evidence on request, and **caught a stale
  caveat of mine**. Their bug (pages thinner than the plan promised) overlaps this one at the
  *symptom* and not the cause; their `homegarden.uk` build is the worked example.
- `bugs_closed/187` deliberately ruled the reconcile emitter unguarded on 2026-08-03 — read before
  touching that arm.
- `bugs_open/220` owns 42 of the 87 no-op items (tool pages, different cause).
- **Grep by the SYMPTOM, not the bug number** — this population is named under 206, 220, 328 and
  closed 187, and `who-owns.py` cannot find it.

---

## 5. Can this close? — my recommendation, and the owner's call

**The engineering is finished.** Every line is shipped, reviewed by two councils, and live.
CLAUDE.md's bar for `bugs_closed/` is *"fixed AND live"*, and a literal reading is now satisfied.

**I recommend keeping it OPEN, for one reason that I think outweighs the tidiness:**

> This file's own title is *"'the machinery is proven live' (vetcomparison PLAN, 07-26) was an
> **unverified inference**, now falsified."*

**This bug exists because a session concluded the machinery worked without watching it work** — and it
was re-opened on 2026-08-24 for the same reason a second time, when the 08-08 fix turned out to be
live at one producer and absent at another for fifteen days with nobody looking. Closing it now on
"the code is live and the tests pass" would be the **third** instance of precisely the inference this
file was filed to refute.

**The engineering being finished and the behaviour being verified are different claims.** This is the
one bug in the estate where conflating them *is* the documented defect.

**If the owner prefers to close on the CLAUDE.md bar**, the honest way is to close it **stating
plainly that no artefact was ever observed**, and to carry forward (a) the nine parked rows and
(b) the unobserved-routing gap somewhere they will not be lost — because a closed file's caveats
stop being read, which is `a-pass-from-a-blind-check-outlives-the-blindness`.

**Cheapest honest path to a real close:** wait for any lane to build a greenfield site with an
`entity-directory` page — free, no site disturbed — then run §2(a)'s two queries. There is no cheap
way to force it: `reconcile_site_plan` runs only inside `build-site-planner`, whose earlier steps are
a full LLM re-plan, which is `bugs_closed/001`'s named hazard and would destroy the
`garden-tools.uk` baseline the owner has ruled untouchable.
