# HANDOFF — bugs_open/450, continue here (2026-09-03, late)

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_450_tool_page_shells/`
Bug: `bugs_open/450_HANDOFF_2026-09-02_planned_tool_pages_are_built_as_prose_shells_by_the_link_repair_before_their_tools_exist.md`
Standing five in this directory: PLAN · RUNBOOK · NOTES (a)–(t) · README_where_we_are · SUMMARY_2026-09-03.

**Read §1 first. It is a live, measured failure of this lane's own fix and it is NOT diagnosed.**

---

## §1 ⚠ THE GUARD DID NOT STOP A GENERIC WRITE — measured, undiagnosed, do this first

**What happened `[MEASURED 2026-09-03 ~13:3xZ]`.** Between **13:05:14Z and 13:24:36Z**, six of the
seven canonical seotools planner shells received **6 writes each (36 `page_component_history` rows
total)**:

```
seotools.co.uk  tool-core-web-vitals-checker        6 writes   tool rows ever: 0   policy: generic
seotools.co.uk  tool-keyword-difficulty-estimator   6 writes   tool rows ever: 0   policy: generic
seotools.co.uk  tool-meta-tag-checker               6 writes   tool rows ever: 0   policy: generic
seotools.co.uk  tool-redirect-chain-checker         6 writes   tool rows ever: 0   policy: generic
seotools.co.uk  tool-robots-txt-tester              6 writes   tool rows ever: 0   policy: generic
seotools.co.uk  tool-title-tag-scorer               6 writes   tool rows ever: 0   policy: generic
```

These are **exactly the pages this bug is about** — `page_type='tool'`, `rebuild_policy='generic'`,
**never** a `component_level='tool'` row. The producer was **`needs_content_page` →
`page-build-handler`**, the generic builder. **Zero `owned_page_review` rows of any class were
created in that window.** So the guard neither refused nor left a receipt.

**The guard WAS live at the time.** The pods running 13:05–13:24Z were the 12:06Z set,
`v1.0.1359`'s predecessor `v1.0.1358`, stamped `d0252fd4d`, and
`git merge-base --is-ancestor 587666be8 d0252fd4d` passes. (A newer chassis, **`v1.0.1359`, stamp
`3043885191b20a0e9b83594b2002e8805fbe95ec`**, rolled at **13:28Z** and carries everything through
`b1a3107e6`, including the `29b40e8bc` narrowing.)

**What is established:**

- `page-build-handler` **does** declare `refuse_owned_page: true` on its `load_page_record` step,
  so the early-refusal arm is configured, not missing.
- The `needs_content_page` items in play were created **2026-09-03** by
  **`rerender_single_page_action`** and by **`tool-generator`** — *not* by the link repair, and
  **not by `tool-deployer`**.
- Statuses seen on them: `claimed`, `complete`, `triaged`, `deferred`.
- The write set splits **18 rows with a resolvable `source_item_id`** and **18 with none**, which
  suggests **two write paths**, not one.

**What is NOT established — do not assert any of it without checking:**

1. **Why neither refusal fired.** At that image, BOTH `load_page_record`'s arm AND the
   `save_page_sections` arm were active (`29b40e8bc` removes only the save arm and was not yet
   aboard). Either should have refused. Candidate explanations, untested: the guard's page lookup
   returned `uuid.Nil` and it stood down (`checked=false` fail-open); the policy read errored
   (also fail-open, logged at ERROR); the write path does not cross `save_page_sections` at all;
   or the items were claimed by a path that skips `load_page_record`.
2. **Whether `29b40e8bc` widened this.** It removed the tool arm from `save_page_sections`
   *deliberately*, on the argument that every generic path is caught earlier. **§1 is evidence
   that argument may be wrong** — if the earlier seams did not catch a `needs_content_page` build,
   then removing the save backstop makes this failure permanent rather than incidental. **This is
   the single most important thing to settle.**
3. **`rerender_single_page_action` as a producer of `needs_content_page`.** This lane never
   accounted for it. The bug's own census listed `page_rerender` as a writer of shell pages
   (3 pages / 20 writes) but treated it as a *rerender* path, not as a *minter of generic build
   items*. If it raw-INSERTs (as `deploy_tool_action.go:674` does for guide pages) it bypasses the
   `writeWorkItem` door entirely.

**First moves, in order:**

```bash
# 1. Which action actually wrote those rows? Check the two write paths.
#    (18 rows carry a source item, 18 do not — find the second writer.)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT h.source_item_id IS NULL AS no_source, h.change_reason, count(*)
  FROM page_component_history h JOIN pages p ON p.id=h.page_id
 WHERE h.created_at BETWEEN '2026-09-03 13:00:00+00' AND '2026-09-03 13:30:00+00'
   AND p.name LIKE 'tool-%' GROUP BY 1,2;"

# 2. Did the guard stand down (fail-open) rather than pass the page?
#    agent_error_log has NO created_at column — use its own timestamp column (\d it first).
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "\d agent_error_log"
# then count error_code='OWNED_PAGE_GUARD_UNCHECKED' in the window.

# 3. Reproduce deliberately, on ONE page, and watch the new image refuse or not.
#    v1.0.1359 is live and carries the narrowing, so this also tests point (2) above.
```

⚠ **Do NOT verify by re-rendering** — see RUNBOOK §8b. And note the narrowing means a re-render
write to a shell page is now *expected*; §1's writes are **not** that case, because their producer
is `needs_content_page` at the generic builder.

---

## §2 What is live, what is committed, what is blocked

| thing | commit | state |
|---|---|---|
| Door half — derived refusal, 6 seams | `587666be8` | **live** (also in v1.0.1359) |
| Narrowing — tool arm off `save_page_sections` | `29b40e8bc` | **live** in v1.0.1359 ⚠ see §1(2) |
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

**Owner lever available:** `DISABLE_TOOL_SHELL_REFUSAL` disarms the tool arm fleet-wide with no
build, scoped so it cannot touch migration 164's owned protection. Relevant if §1 turns out to be
the arm misbehaving rather than under-firing.

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

1. **§1 — diagnose why the guard did not refuse, and re-test whether `29b40e8bc` was safe.**
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
