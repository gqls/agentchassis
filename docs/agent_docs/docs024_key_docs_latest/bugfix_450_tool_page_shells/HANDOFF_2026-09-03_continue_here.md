# HANDOFF — bugs_open/450, continue here (2026-09-03, late)

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_450_tool_page_shells/`
Bug: `bugs_open/450_HANDOFF_2026-09-02_planned_tool_pages_are_built_as_prose_shells_by_the_link_repair_before_their_tools_exist.md`
Standing five in this directory: PLAN · RUNBOOK · NOTES (a)–(t) · README_where_we_are · SUMMARY_2026-09-03.

**Read §1 first.** It corrects a wrong conclusion of mine (the guard did NOT fail) and names the
real live bug: six pages serve working tools whose `component_id` is NULL, so they are one rerender
away from losing them. **That repair is time-sensitive and is the first job.**

---

## §1 ⚠ CORRECTED — THE GUARD DID NOT FAIL. A DIFFERENT AND WORSE BUG IS LIVE: `save_page_sections` KEEPS A TOOL'S HTML AND DROPS ITS `component_id`

> **This section said "THE GUARD DID NOT STOP A GENERIC WRITE" for about twenty minutes and that
> was WRONG.** Corrected 2026-09-03 ~13:5xZ after the `portfolio_positioning` lane measured the
> artefact and I verified it. The error was mine and it is the day's recurring one, in its most
> consequential form: **I measured POST-write state and inferred PRE-write state.** Left uncorrected
> it would have sent the next session hunting a guard bug that does not exist. Kept visible rather
> than deleted, because the shape is the lesson.

### What actually happened

The `portfolio_positioning` lane's repair wave attached real tool components to these six pages at
**09:34:09, 09:37:24, 09:40:35, 09:46:14, 09:49:05 and 09:54:14Z** — one library component each
(`tool-robots-txt-tester-seotools-co-uk` etc., `component_level='tool'`, `is_active=true`, still
present and untouched).

So at **13:05–13:24Z**, when `needs_content_page` → `page-build-handler` wrote to them, **these
pages were NOT shells. They carried live tools.** The guard therefore **correctly allowed** the
write, and its silence was right rather than a failure.

**`save_page_sections` then deleted all three rows and re-inserted them ~80 ms later, preserving
the tool slot's `rendered_html` (17,938–23,953 bytes) and setting its `component_id` to NULL.**
Verified directly:

```
tool-robots-txt-tester        hero-tool           cid=set    3,500 B   13:05:15
tool-robots-txt-tester        generic-text-block  cid=set    3,142 B   13:05:15
tool-robots-txt-tester        tool-robots-txt-…   cid=NULL  20,839 B   13:05:15   <-- reference gone
```

`page_component_history` for that slot shows six writes at 13:05:14, **every one already
`cid=NULL`**.

### Why my "0 tool rows ever" reading was an artefact

`toolShellPredicateFor` — and every census in this lane, and the query I used to declare the
falsification — joins `page_components` to `content_components` **on `component_id`** and filters
`component_level='tool'`. **With the id NULL the join drops the row**, so a page serving a real
20 KB tool reads as having *no tool component, ever*. The zero I measured was created by the very
write I was trying to explain.

**Consequences, and (2) is the one that bites this lane's own instruments:**

1. **Serving is correct today.** All six serve working tool controls with instance-scoped ids
   (`id="c-tool-robots-txt-tester-fetch-domain-input"`), 78–85 KB bodies, 6 scripts. A visitor sees
   a working tool. So the writes landed and the outcome was survivable.
2. **⚠ THE CENSUS AND THE GUARD NOW BOTH OVER-REPORT BY SIX.** These six count as shells in the
   66/67 population and are not. The guard will also now *refuse* future generic writes to them —
   which is accidentally protective, given (3), but for a reason that has nothing to do with what
   the predicate is meant to express. **This is the census-versus-predicate divergence written up
   in `016b` §9 this morning, reappearing inside my own instrument one field along.**
3. **THE REAL EXPOSURE IS THE NEXT REBUILD.** Anything that regenerates from `component_id` rather
   than from stored HTML has nothing to regenerate. **These six pages are one rerender away from
   losing their tools for real, and nothing about their current appearance would warn anyone.**

### What this means for `29b40e8bc` — the worry is RETIRED

The previous version of this section warned that removing the tool arm from `save_page_sections`
may have removed a needed backstop. **It did not, and could not have:** the pages were not shells
at write time, so neither the narrowed nor the un-narrowed guard would have refused. The narrowing
is irrelevant to this incident in both directions. (It is also not the cause: `29b40e8bc` was NOT
aboard `v1.0.1358`, the image running at 13:05Z, and it changes only the guard condition, not the
delete-and-reinsert.)

### The actual open bug — likely deserves its own `bugs_open/` file

**`save_page_sections`' delete-and-reinsert preserves the `rendered_html` of a slot it does not
recognise as one of its own planned sections, and drops the `component_id`.** The plan for these
pages names `hero-tool` and `generic-text-block`; the tool slot is not in it, so the save carries
the bytes forward without the reference.

First moves:

```bash
# 1. Scope it — how many pages fleet-wide have a tool-level slot with a NULL component_id?
#    (This cannot use the component_level join — that is the whole point. Match on slot_name.)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT s.domain, p.name, pc.slot_name, length(pc.rendered_html) AS bytes
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.component_id IS NULL AND pc.build_status<>'removed'
   AND pc.slot_name LIKE 'tool-%' ORDER BY 1,2;"

# 2. Read the re-insert in save_page_sections_action.go: which branch omits component_id for a
#    slot that is not in the plan's section list, and is the omission deliberate anywhere?

# 3. Decide the repair: re-point component_id from content_components by matching slot_name to
#    cc.function (the acceptance coupling — pages.name == cc.function, sanitiseFunction guarantees
#    the tool- prefix). Cheap and reversible. Do this BEFORE any rerender touches these six.
```

⚠ **`page_type='tool'` is not the only exposure.** Any page whose plan omits a slot that was
inserted by another producer can lose that slot's reference the same way.

## §2 What is live, what is committed, what is blocked

| thing | commit | state |
|---|---|---|
| Door half — derived refusal, 6 seams | `587666be8` | **live** (also in v1.0.1359) |
| Narrowing — tool arm off `save_page_sections` | `29b40e8bc` | **live** in v1.0.1359 (§1: the worry about it is RETIRED) |
| Receipt wording corrected (stated fact, not inference) | `196319707` | live |
| Plan-side gate `enforce_tool_sources` | `5e6fee47b` | live but **KEYLESS** — inert |
| Migration 729 arming the key | `681190083` | **committed, NOT APPLIED — BLOCKED** |
| Register PBP-053 / BLD-029, finding code, landmine, 016b ×2 | various | committed |

**Councils: both APPROVED.** Door `2b236e83-ffd1-4911-b73f-1c17249064c1`; gate
`4e7497ed-62ed-4426-a814-8361754c2352`. All mediums actioned (see NOTES (i), (k)).

**BLOCKED, needs the owner:** applying `729` was **refused by the session permission classifier**
(a live-DB write). Not worked around. Preconditions are otherwise met — verdict read, code live.
Recipe and the reason it waited: RUNBOOK §10. Until it applies, **the planner keeps naming tool
pages whose tools do not exist.**

**Owner lever, still available and now probably NOT wanted:** `DISABLE_TOOL_SHELL_REFUSAL` disarms
the tool arm fleet-wide with no build, scoped so it cannot touch migration 164's owned protection.
⚠ Per §1 it would also disarm the accidental protection those six `component_id`-NULL pages are
currently getting, so **do not pull it until their references are repaired.**

---

## §3 Numbers you can trust, and the ones you cannot

**Use the guard's own predicate — `toolShellPredicateFor` in `owned_page_guard.go` — as the
census. RUNBOOK §1 carries a copy and says to diff it against the function first.** Four
measurement errors in this lane came from paraphrasing it (016b §9 has the pattern).

- shell pages: **66–67 / 15–16 sites**, stable across independent readings by two lanes.
- of those, **~54 already serving** deployed components; only **~13** empty.
- genuinely NEW refusals versus the pre-existing owned population: **19** (48 of 67 were already
  `rebuild_policy='owned'`).
- ⚠ **`61 / 10` appears in older text and is a FLOOR twice over** (missing `deployed_at IS NULL`
  and missing `cc.is_active`). Superseded; the bug file's Verify block explains both.
- ⚠ the census is **repair-INITIATED, not repair-COMPLETED**: a page leaves it when a tool
  component attaches, while the public still sees prose until the rerender drains.
- ⚠ drain rate: **NOT established.** "39 repairs in 12h" over-counts — the predicate cannot
  separate a first tool arriving from a tool being *regenerated*. NOTES (q).

**The harm metric, and why its earlier zero meant nothing:** historical share of writes hitting
shell pages = **275 / 17,205 = 1.60%** of all `page_component_history` writes. Condition on fleet
activity, not wall-clock. §1's 36 writes are far above that share and are the falsification.

---

## §4 Peer lanes — live dependencies

- **`portfolio_positioning`** — owns the INSTANCE repair. All 8 tools built, adopting existing
  pages at existing URLs. **Owes this lane a served-body reading of the seven seotools pages**
  (whether the tool and the leftover prose visibly compete at position 2). Their finding, recorded
  in NOTES (s): the **sectionless** fork repairs *cleaner* — this inverts my own argument for the
  plan-side gate **in its favour**.
- **`bugs_open/427` / `454`** — reported the regression §2 row 2 fixes; wrote the measurement
  pattern into 016b (`80f74b23d`). Their `9831e9ab4` is live. ⚠ Their standing warning: until it
  rolled, every re-render served stored data back at itself.
- **`bugs_open/444`** — the gate frame this siblings. Told (bug file follow-up) that arming
  `enforce_tool_sources` **changes what their listing gate does on the same plan** (ordering).
- **`428`** — adding a record-only reconciliation block BELOW both gates and a page-type
  external-producer registry. **Warned** that the set of things that WRITE a tool page is wider
  than the set that PRODUCES its tool — §1 is the evidence.
- **`apis.uk`** — owns 640's rule 17; confirmed the anchor 729 defends and added an EXTERNAL
  READERS note at their end.

---

## §5 Open, ordered

1. **§1 — repair the six NULL `component_id` references BEFORE any rerender reaches those pages**,
   then scope the class fleet-wide and read the re-insert branch that drops the reference. The
   guard question is CLOSED (it behaved correctly); this is a distinct and more urgent bug.
2. **Apply migration 729** once the permission question is settled (owner).
3. **`bug_historian`'s standing objection (council, low, accepted):** nothing PINS the §7
   assumption that nothing reads planned tool pages. It is a negative finding in a code comment.
   A periodic "has a reader appeared" check is the real answer. Named, not built.
4. **Residual, explicitly out of scope:** the 61+ existing shells (instance work), the
   `owned_page_review` hold still having no consumer, `rerender_single_page`'s re-assembly path
   (bug 210 family), and N-links-one-page churn (220's own candidate).

## §6 Traps this lane paid for — read before touching anything

- **RUNBOOK §8b** — do not verify with a re-render.
- **RUNBOOK §10** — 729's apply preconditions; and while 729 is applied, `720_ROLLBACK` refuses by
  design (LANDMINES entry; unwind newest-first).
- **RUNBOOK §1** — copy the guard's predicate; do not paraphrase it.
- **016b §9 ×2** — the measurement pattern, and "a correct predicate wrapped in untested
  inferences" (a check can fire on the right rows and tell the operator something false; no test
  sees it because every test asserts the predicate).
- **`WRONG_CALLS.md`** — six entries from this lane today, all under my own name. The recurring
  one: *the predicate was right, and every sentence I wrapped around it was an untested inference.*
- Timestamps here are **UTC from the database clock**. `agent_error_log` has **no `created_at`** —
  `\d` it before querying.
