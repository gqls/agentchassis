# 175 — (renumbered from 172 on 2026-08-01: another session had already filed a different 172 — the agent-state-cap truncation case. Resolve by slug, per CLAUDE.md.)

# 175 — four more `pages` upserts silently drop `page_type` on a name collision, and one does the opposite

**Filed:** 2026-08-01 by the `bugfix_081` lane, **at the council gate's request**.
Reviewing `081`'s fix (correlation `ccd4384c-aff9-45ed-80b2-01c3ced573bb`), the
`bug_historian` seat objected that the plan fixes one arm of a shape it has seen
recur across subsystems, and that nothing confirmed there was no third write path:

> *"no check confirms there is no THIRD write path (replan/rerender/other planner
> arms — cf. `bugs_closed/037`, `bugs_closed/038`, the replan-clobbers-built-pages
> history) still doing a blind title/sections overwrite keyed only on name,
> unguarded by build_status+page_type."*
>
> *"This plan fixes one arm well; it does not claim to audit siblings, and the plan
> should not be read as closing the class."*

It was right, and the audit it asked for takes one grep. **There are four more.**

**Severity:** unmeasured, and that is the honest headline — see § Exposure. The
*shape* is confirmed by reading; the *live damage* is not established for any of
the four, unlike `081` where three months of looping was measured.

**Status:** OPEN, unowned. Filed as a finding with the census done, not a fix.

---

## The shape (from `bugs_closed/081`, and 016b §9)

An upsert names SOME columns in its `DO UPDATE SET`. `page_type` is in the INSERT
and not in the SET, so on a name collision the arm silently becomes a PARTIAL
update: the new content lands under the OLD role. Nothing errors, `RETURNING id`
yields a page id either way, and the caller cannot tell a create from an
overwrite. On a **deployed** page that overwrites live content AND leaves the
routing key wrong, so whatever asked for the page asks again next sweep.

## The census — `grep -rn "ON CONFLICT (site_id, name)" --include=*.go`, 2026-08-01

| site | INSERT sets `page_type` | `DO UPDATE SET` | verdict |
|---|---|---|---|
| `apply_gap_plan_action.go:427` | `$5` (planner's) | **fixed** — `DO NOTHING` + explicit branch | `bugs_closed/081` |
| `create_report_page_action.go:164` | `'report'` | `url, title, sections, updated_at` | **same shape** |
| `deploy_tool_action.go:376` | `'tool'` | `url, title, sections, updated_at` | **same shape** |
| `deploy_tool_action.go:514` | `'blog-post'` | `title, updated_at` | **same shape** |
| `create_tool_component_action.go:416` | `'blog-post'` | `title, updated_at` | **same shape** |
| `apply_adoption_plan_action.go:532` | `$5` | `title, url, **page_type**, sections, meta_description` | **opposite risk** — see below |

**Four share `081`'s omission.** They differ from `081` in one way that matters:
each hardcodes a *constant* `page_type` (`report`, `tool`, `blog-post`) rather
than taking one from an LLM plan. So the collision they can produce is narrower —
"a page named X already exists under a different type, and a tool/report/guide
deploy is about to write its content into it" — but it is the same partial update
and the same silent overwrite of whatever was there.

`deploy_tool_action.go:514` and `create_tool_component_action.go:416` are the
**same companion-guide upsert duplicated in two files** (byte-identical SQL,
different callers). Whatever is decided applies to both, and they should probably
be one function — flagging, not claiming.

## The sixth is the opposite risk, and should not be "made consistent"

`apply_adoption_plan_action.go:532` **does** carry `page_type = EXCLUDED.page_type`.
That is `081`'s fix candidate 1 — the one `081` declined, because it hands an arm
authority to re-type any live page it collides with. On the adoption path that may
well be correct (adoption is explicitly importing an existing site's structure and
is entitled to say what a page is), but it is the same authority `081` argued a
*generic* arm should not have.

**Do not resolve this file by making all six identical.** The two failure modes
are opposite — omit and you get a silent partial update; include and you get an
unguarded re-type — and the right answer differs per call site. `081`'s answer was
a third thing: refuse the collision and route the decision to a human.

## Exposure — NOT measured, and the query to do it

`081` earned its severity from three months of a measured loop. **This file has no
equivalent and does not pretend to.** What is established: the shape, by reading
the SQL. What is not: whether any of the four has ever actually collided.

Each of these arms is keyed on a name it computes itself (a tool slug, a report
name, a guide name), so a collision needs a pre-existing page of that name under a
different type. That is exactly what `bugs_closed/015`'s and `081`'s history shows
does happen, but it has not been counted here.

```sql
-- pages whose type disagrees with what a tool/report/guide arm would write.
-- Start here; a non-zero count is the trigger for taking this file seriously.
SELECT s.domain, p.name, p.page_type, p.build_status, p.status
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE (p.name LIKE '%-guide' OR p.name LIKE '%-report' OR p.page_type IN ('tool','report'))
  AND p.page_type NOT IN ('tool','report','blog-post')
ORDER BY 1, 2;
```

`[UNMEASURED]` deliberately — running it against production is one command and the
next thread should do it *before* choosing a fix, not after.

## Fix candidates (none applied)

1. **Per-call-site decision, following `081`'s shape where it fits.** For an arm
   whose `page_type` is a constant, the question "may I overwrite a live page of a
   different type with my content?" has an obvious answer (no), and the `081`
   treatment — read the row, refuse, file for review — transfers directly.
2. **A shared `upsertPageForRole` helper** that takes the role and does the
   read-decide-write once. Closes the class rather than six instances, and it is
   the only candidate that stops a seventh arm being written with the same bug.
   Bigger, and it is a shared seam — architecture-scope, so it needs the RFC route
   rather than a bug patch (CLAUDE.md, platform seams).
3. **Detector only:** a check for `pages` rows whose `page_type` disagrees with
   the role their `sections`/`url` imply. Weakest, but it would give this file the
   exposure number it currently lacks.

**Do not take (2) without deciding the adoption-path question above**, or it will
quietly impose one answer on six call sites that may want two.

## How to verify a fix

Same as `081`: induce both branches. A collision with a live page of a different
type must not mutate it; a clean create and a same-type refresh must still work.
And per `081`'s own correction — **break the guard and watch the test fail**, or
the test is a decoration (`LANDMINES.md`, "`mock.ExpectationsWereMet()` is NOT
'no database call happened'").

## Related

- `bugs_closed/081` — the parent. Its § "What was done" is the shape to copy, and
  its scope measurement is the discipline: it declined a wider fix on a count.
- 016b §9 — "An `ON CONFLICT DO UPDATE` that names SOME columns turns a CREATE
  into a silent PARTIAL update".
- `bugs_closed/015` — mistyped `page_type`, the stranded/newly-planned half.
- `bugs_open/080` — `new_page` bypassing canonicalisation; same file, and the
  reason a disagreeing name INSERTS a second row instead of colliding at all.
- Concept register `WII-008`.

---

# FIXED 2026-08-02 — candidate 2, and the exposure this file asked for

**Fixing lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_175_page_role_upsert/`
(PLAN / NOTES / RUNBOOK / README_where_we_are). **Commit:** `cbbecb021`.
**Concept register:** PBP-027. **Council:** submitted `e78c62e3-7f01-48f1-b083-924eaccd195a`.

## The exposure, measured — § Exposure's `[UNMEASURED]` is discharged

The query in § Exposure returns 2 rows. Written per-arm — "is a page holding a name
THIS arm would claim, under a different type?" — it returns **4, every one
`deployed`**:

| arm | domain | name | page_type | build_status |
|---|---|---|---|---|
| guide-arm | robot-hands.com | `gripper-selection-guide` | content | deployed |
| guide-arm | robot-hands.com | `selection-guide` | content | deployed |
| report-arm | idea.uk | `report-example` | content | deployed |
| report-arm | lendzy.co.uk | `report-loan-shark` | content | deployed |

**And the honest qualification, which the row count alone would hide:** a hit is a
*surface*, not a collision. Tool page names canonicalise to `tool-<slug>`, so the
guide arm's reachable name is `tool-gripper-selection-guide`; the bare form is
produced only by `deploy_tool_action`'s `TrimPrefix` branch or a
`pageNameOverride`. The report arm names pages `report-<uuid>`, so its two rows
are unreachable. **No collision has been observed on any of the four arms.** The
shape is confirmed by reading; this fix is prevention, and the severity line at
the top of this file stands.

## The census here was incomplete (does not change the answer)

Re-grepped 2026-08-02: eleven `ON CONFLICT (site_id, name)` sites, not six. The
five extra `DO UPDATE` arms — `site_db_actions.go:1141`,
`create_blog_posts_action.go:219`, `adopt_verbatim.go:470`,
`cmd/webdesignport/import.go:182` (+ the `seed_content_sources` `DO NOTHING` pair)
— **all already carry `page_type = EXCLUDED.page_type`**, so they belong to the
"opposite risk" camp this file describes and § "The sixth is the opposite risk"
governs them unchanged.

## What was done

Candidate 2, **not** candidate 1, for this file's own stated reason: it is the only
one that stops a seventh arm. `UpsertPageForRole`
(`platform/orchestration/actions/page_role_upsert.go`) owns the write for any arm
whose role is a **compile-time constant**, and answers the collision once:

| collision | outcome |
|---|---|
| none | **created** — every declared column |
| row holds the SAME role | **refreshed** — the caller's declared `Refresh` subset, nothing else |
| DIFFERENT role, never served | **adopted** — the arm takes the row over completely, `page_type` included |
| DIFFERENT role, **has been served** | **refused** — nothing mutated, `mistyped_deployed_page` filed `needs_human_review` |

Converted: `create_report_page_action.go`, `deploy_tool_action.go` (both arms),
`create_tool_component_action.go`. **`apply_gap_plan_action.go` deliberately NOT
converted** — its role comes from an LLM plan, so the ADOPT branch would hand a
generic arm the authority `081` declined. What IS shared is the refusal *filing*
(`fileMistypedLivePageItem`), so both refusal sites now produce one item shape and
one `item_key` and dedupe onto a single open human decision.

**§ "How to verify a fix" is answered on its own terms.** Both branches induced,
and each guard broken and watched to fail (five mutations, five red tests,
unmutated green) — the table is in NOTES. The assertions that catch three of the
five read the **SQL the helper actually built**, because the defect IS a column
list and no statement-ordering expectation can see it.

## Two deliberate divergences from `bugs_closed/081`, both measured

1. **"Has been live" is `build_status IN ('deployed','needs_rebuild') OR
   deployed_at IS NOT NULL`**, wider than 081's `build_status = 'deployed'`.
   `bugs_closed/037` is an entire case about `needs_rebuild` falling outside that
   predicate, and 35 of 46 `needs_rebuild` rows carry a non-null `deployed_at`.
   Now a `LANDMINES.md` entry in its own right.
2. **ADOPT re-types a never-served row**, where 081 refreshes and leaves the type.
   081's declined widening was about *repairing existing mistypes* and was declined
   on a count; this is a different question — a collision on a row nothing has ever
   served — and leaving the type wrong there is the defect itself.

## Prevention, so this file cannot recur as `bugs_open/2xx`

`scripts/pattern-check.py` → `check_partial_page_upsert`: fires when `page_type` is
in the INSERT list and absent from the SET list, silent on the `page_type =
EXCLUDED.page_type` camp. **MEASURED over 1,120 Go files: 4 findings at pristine
HEAD (exactly this file's census), 0 after the fix.** The first run of that
measurement was against a tree that already had the fix and reported 0/0 — logged
in `WRONG_CALLS.md`, because a check nobody has seen fire is not a check.

## Left open, deliberately

- **`create_tool_component_action.go:288`** creates its own tool page with a plain
  `INSERT` and **no `ON CONFLICT` at all**: a collision raises a unique violation,
  deletes the component and errors. Loud and fail-closed, so outside this file's
  silent class. Converting it would make re-runs idempotent — a real behaviour
  change nobody asked for. **Follow-up candidate, not a defect of this fix.**
- **Two arms now return an error on refusal** where they previously overwrote a
  live page (`DeployToolToSiteAction`'s tool page, `CreateReportPageAction`); the
  companion-guide arms stay non-fatal and log. Raised as the open question in the
  council submission and in PBP-027.

---

# CLOSED 2026-08-02 — LIVE on chassis `v1.0.1233`, pod-verified on both replicas, council APPROVED

**Council verdict: APPROVED** — `e78c62e3-7f01-48f1-b083-924eaccd195a`, *"approved with 4
advisory objection(s) — none high-severity"*, 5 seats abstained (relevance-gated), 12 reviewed.

## Live proof (not the tag, not git — the running binary)

`v1.0.1233`, replicas `agent-chassis-85b6bd9777-8pwxz` and `-r4rhf`, identical results:

| grep | expect | got |
|---|---|---|
| `UpsertPageForRole: refused` (ADDED) | >0 | **1 / 1** |
| `adopted an unshipped page into this role` (ADDED) | >0 | **1 / 1** |
| `new_page refused` (POSITIVE CONTROL, 081's string) | >0 | **1 / 1** |
| `url = EXCLUDED.url, title = EXCLUDED.title` (REMOVED — the report arm's own statement) | **0** | **0 / 0** |

The negative control matters and had to be chosen carefully: `ON CONFLICT (site_id, name) DO
UPDATE SET` still returns **4** in the binary, and that is CORRECT — the five arms that carry
`page_type = EXCLUDED.page_type` keep the statement deliberately. A negative control has to
name a string only the removed code had.

The image was also verified **before** the push (`docker run --rm --entrypoint sh …:v1.0.1232`),
which is cheaper than discovering it after a roll. `v1.0.1232` was superseded by another
session's `v1.0.1233` build from the same HEAD before it was applied; both carry the commit.

## The four advisory objections, and what was done about each

| seat | objection | disposition |
|---|---|---|
| `prior_art_librarian` (med) | *"'exactly one producer and no automated consumer' is sourced from a test docstring, not a live check"* | **Right, and now checked live:** `0` active `agent_definitions` reference `mistyped_deployed_page` in any config column; `0` rows in `site_work_items`. The docstring's claim holds — but it was folklore until this query. |
| `debug_historian` (med) | *"no post-deploy verification step — nothing confirms the running pod carries `UpsertPageForRole`"* | **Discharged by the table above**, both replicas. On the `git archive` tmpfs concern: the extraction yielded 1,120 `.go` files and a full `go build ./...` succeeded, which a truncated tree cannot do. |
| `tooling_provenance` (med) | *"no `doc_notes` row keyed to the subject; the next fixer has only a markdown trail"* | **Two `LANDMINES.md` entries added and synced** via `scripts/landmines-sync.py --apply` (the sanctioned path — CLAUDE.md forbids hand-writing landmine rows into `doc_notes`): the `build_status` predicate, and the three-helpers trap below. |
| `guardian` (med) | *"confirm the other consumers tolerate 3 producers; get explicit sign-off on ADOPT and on refusal-becomes-error, not a risk footnote"* | Consumer check done (row above); `verifier_coverage_test.go`'s description was updated in the shipping commit. The sign-off is **not mine to give** → `architecture_review/RFC_010`. |
| `architecture` (med, objecting on record, not blocking) | *"this is architecture-scope … the blast-radius/rollback framing belongs in the architecture_review track so the next class-fix has a citable precedent"* | **`architecture_review/RFC_010_who_may_answer_a_page_name_collision.md` filed**, with the three questions costed and a recommendation. |
| `editquality` (low) | *"`recurrenceExpected` is load-bearing on repeated `item_key`s — three producers could hit the two-strike drop"* | **Checked in the code, and it cannot:** the two-strike probe counts only `status IN ('complete','failed')` rows. A `needs_human_review` decision is non-terminal, so while it is open `idx_swi_dedup` collapses repeats and neither branch fires. And at two strikes the item is still **inserted**, labelled `unresolved` (`load_work_item_actions.go:1190-1198`) — it is a relabel, not a drop. |
| `reuse_agent` (low) | *"confirm no existing `pattern-check.py` rule already scans `pages` upserts"* | **Checked:** the only other `ON CONFLICT` rule is `check_unguarded_migration_insert`, scoped to migration files. No overlap. |
| `prior_art_librarian` (low) | *"no check was run for an existing page-upsert helper before building one"* | **Checked, and it found something worth keeping:** `site_db_actions.go:1090 upsertPage` and `cmd/webdesignport/import.go:163 upsertPage` already exist — the plan-sync path, which carries `page_type = EXCLUDED.page_type` and therefore has the **opposite** collision policy. Not a duplicate (its role comes from the plan, not a constant), but a genuine trap → now a `LANDMINES.md` entry. |

## Where everything lives

- **Fix:** `cbbecb021` · **docs:** `9897cb82f` · **016b §9 follow-up:** `588adb6f1`
- **Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_175_page_role_upsert/`
- **Register:** PBP-027 · **RFC:** `architecture_review/RFC_010` · **Landmines:** two entries, synced
- **Recurrence check:** `check_partial_page_upsert` in `scripts/pattern-check.py` (4 → 0)

## What is deliberately NOT closed by this

- **`architecture_review/RFC_010` is OPEN** — the ADOPT branch's authority and the
  producer-set expansion want an owner ruling, not a bug closure.
- **`create_tool_component_action.go:288`** — the tool-page `INSERT` with no `ON CONFLICT`
  at all. Loud and fail-closed, so outside this class; a follow-up if anyone wants
  re-runs to be idempotent.
