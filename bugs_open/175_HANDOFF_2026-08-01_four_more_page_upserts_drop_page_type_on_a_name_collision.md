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
