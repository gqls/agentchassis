# Handoff — the per-page adoption lock does not exist, so `adoption_locked` is a per-SITE flag that is true only on a site's first plan

**Filed 2026-07-20**, found while closing out `/bugs_open/001`. No fire: the practical exposure today
is three pages on a test site. What is broken is a **documented protection mechanism that is not
built**, and a **code comment that justifies a design decision by appealing to it**. The cost of
leaving it is that threads keep reasoning from it — `/bugs_open/001` did, and so do the comments in
`v3_site_actions.go`.

This is the same class as `/bugs_open/025` (a column documenting behaviour that does not exist).

## Status — 2026-07-21 (bugfix-051 thread)

**Fix candidate 1 (correct the false premise everywhere it is written) is COMPLETE.**
- `v3_site_actions.go` comments — already done before this thread (`1a13e265d`, refs
  renumbered `49b3c46b7`), committed at HEAD. The Pass C2 header note already reads
  "first-plan-only", not "90-day window". The rename half of candidate 3 (the *comment*
  re-scoping) is therefore also already in.
- `bugs_closed/001` §"Deliberately NOT done" (the `:310` "90-day window" paragraph) — corrected
  this thread with an inline `> **CORRECTED 2026-07-21**` note.
- `053_build_site_planner.sql` — corrected this thread at the intent line (`:2537`) and at the
  head of the "adoption locks"/full-054 section, flagging branch (b) as never-built design
  history, not live behaviour.

Live query re-verified against `agent_definitions` on 2026-07-21: still the single-branch
"no current plan" form, `build_status` now surfaced (migration 173). No 90-day / per-page /
timed branch present. So there is nothing further to correct for candidate 1.

**What remains is a direction choice, and it is the owner's — candidates 2 and 3 compete:**
- **Candidate 3 (retire the concept):** rename the wire key `adoption_locked` →
  `site_has_no_current_plan` / `is_first_plan` so nobody reasons about a lock again. This is the
  *only* remaining code action, and it is a config-live / Go-inert coupling change: the alias
  lives in the live `load_existing_pages` query (`agent_definitions`) while the deployed pod reads
  `rm["adoption_locked"]`, so it must ship image-first (Go reads the new key with a fallback,
  roll, then re-seed the query) or a first-plan adoption preserve breaks in the gap. It also
  forecloses candidate 2.
- **Candidate 2 (build adoption faithfulness for real):** build BOTH halves of branch (b) — the
  query branch AND a writer that emits `scope='page'`, `category='preserve'`, `locked_by='adoption'`,
  `lock_type='timed'` directives at adoption time. A real feature, not a bug fix; keeps the
  `adoption_locked` name meaningful.

No production fire either way — measured exposure is ~nil (the 3 `planned` dartsonline pages are
001-verification artefacts; the 17 `needs_rebuild` pages belong to `/bugs_open/037`). Awaiting the
owner's direction on 2-vs-3 before any code lands.

## The claim in one line

`adoption_locked` is **not** a per-page, 90-day-expiring lock. It is a **per-site** boolean meaning
"this site has no current plan", so it is true for every page on a site's **first** plan run and
false for every page on every re-plan thereafter.

## Evidence

**1. The live query has one branch, not two.** Read from `agent_definitions`, not from a migration
file (2026-07-20):

```sql
SELECT default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query'
FROM agent_definitions WHERE type='build-site-planner' AND deleted_at IS NULL AND is_active=true;
```
```sql
       CASE
         WHEN NOT EXISTS (
             SELECT 1 FROM site_plans sp
             WHERE sp.site_id = p.site_id AND sp.is_current = true
         ) THEN true
         ELSE false
       END AS adoption_locked
```

There is no `lock_expires_at`, no `site_plan_directives` join, no per-page test.

**2. The designed second branch is absent.** `docs/agent_docs/sql_for_agents/053_build_site_planner.sql`
§054 specifies `adoption_locked` as true when EITHER (a) there is no current plan **OR** (b) the
current plan has a live adoption preserve-directive for that page (`scope='page'`,
`category='preserve'`, `locked_by='adoption'`, `lock_type='timed'`, `lock_expires_at > NOW()`). Its
own note: *"After the 90-day window, (b) goes false (expiry), (a) no longer applies (a plan exists),
so adoption_locked = false and the site develops normally."* **Branch (b) is not in the live query.**

**3. Branch (b) would match nothing even if restored.** Fleet-wide:

```sql
SELECT locked_by, lock_type, category, scope, count(*),
       count(*) FILTER (WHERE lock_expires_at > NOW()) AS still_live
FROM site_plan_directives GROUP BY 1,2,3,4;
--  locked_by | lock_type | category | scope | count | still_live
--  (null)    | (null)    | content  | site  |   266 |          0
--  (null)    | (null)    | design   | site  |   196 |          0
```

462 directive rows, **`locked_by` NULL on every one**, `lock_type` NULL on every one, and **zero
rows at `scope='page'` with `category='preserve'`**. No code writes one either: `write_site_plan_action.go`
only *transfers* `locked_at`/`locked_by` from a previous plan (`transferDirectiveLocks`, `:638`), so
there is nothing to transfer from. [UNVERIFIED] whether branch (b) was ever live — only one
`build-site-planner` row survives in `agent_definitions` (version 1, no snapshots retained), so the
query's history cannot be traced there. It does not matter for the consequences below.

## Consequence 1 — Pass C2 fires only on a site's first plan, never on a re-plan

`reconcilePlanWithRealised` (`v3_site_actions.go:4575`) builds `lockedPages` from `adoption_locked`
alone (`:4588-4591`). `lockedPages` has exactly one consumer: `itemStemSets` (`:4636`), which has
exactly one consumer: **Pass C2** (`:4684`), the item-topic dedup that drops an LLM page
re-proposing an adopted item under a different name.

A re-plan by definition runs on a site that **has** a current plan. So on every re-plan
`adoption_locked` is false for every page → `lockedPages` is empty → `itemStemSets` is empty →
**Pass C2 cannot fire.** It is live only during a site's first plan after adoption.

That is not necessarily wrong — but the comment justifying its scope is:

```go
// v3_site_actions.go:4556-4558
//	          "tool-pricing" beside a built "guide-pricing" shares the stem
//	          "pricing"). Bounded to the 90-day window that risk is acceptable;
//	          made permanent for every built page it is not [...]
```

and `/bugs_open/001`'s "deliberately NOT done" note repeats it: *"Bounded to the 90-day window that
is acceptable; permanent for every built page it is not."* **There is no 90-day window.** The real
bound is "the first plan only", which is tighter than the comment claims and reached by a different
mechanism. The decision may still be right; the stated reason for it is not a thing that exists.

## Consequence 2 — "adoption faithfulness for 90 days" is not delivered

`053_build_site_planner.sql:2537` states the intent: *"adoption faithfulness for 90 days using new
lock"*. What actually holds is: an adopted page is force-preserved during the **first** plan run,
and from the second plan onward only if `build_status='deployed'` (the `/bugs_open/001` guard).

Exposure measured 2026-07-20 — pages carrying a real composition, **not** deployed, on a site that
has a current plan, i.e. unprotected on the next re-plan:

```sql
SELECT p.build_status, count(*) AS pages, count(DISTINCT p.site_id) AS sites
FROM pages p WHERE p.status='active' AND p.build_status <> 'deployed'
  AND jsonb_array_length(coalesce(p.sections,'[]'::jsonb)) > 0
  AND EXISTS (SELECT 1 FROM site_plans sp WHERE sp.site_id=p.site_id AND sp.is_current)
GROUP BY 1 ORDER BY 2 DESC;
--  needs_rebuild | 17 | 5
--  planned       |  3 | 1
```

**The `needs_rebuild` slice (17 pages) belongs to `/bugs_open/037`, not here** — do not fix it twice.
The `planned` slice is three pages, all on `dartsonline.com` (`brands`, `guides`, `shop`), and all
three are artefacts of 001's own verification runs. So **today's live exposure is close to nil**;
this is filed for the wrong premise, not the damage.

Context: only **8 of 29 sites** currently have a current plan, so on the other 21 every page still
reads `adoption_locked=true`. [INFERRED] most of those are the 17 inert `status='pool'` sites from
the news-feed pooling workstream, which have never been planned.

## Fix candidates

1. **Correct the comments and delete the dead reasoning** (cheapest, and strictly an improvement
   whatever else is decided). `v3_site_actions.go:4531` and `:4556-4558`, plus the corresponding
   paragraph in `/bugs_open/001`. Say what the flag means: "true on a site's first plan only".
2. **Decide whether adoption faithfulness is still wanted.** If yes, branch (b) needs *both* halves
   built — the query branch **and** a writer that emits `scope='page'`, `category='preserve'`,
   `locked_by='adoption'`, `lock_type='timed'` directives at adoption time. Restoring the query
   branch alone changes nothing (see evidence 3) and would look like a fix while being inert — the
   `/bugs_open/032` shape.
3. **Or retire the concept**: rename `adoption_locked` to what it is (`site_has_no_current_plan` or
   `is_first_plan`) so nobody reasons about a lock again, and re-scope Pass C2 explicitly to the
   first-plan case in its own comment.

Recommendation: **1 now, 3 next, 2 only if the owner wants adoption faithfulness back** — it is a
real feature with a real writer to build, not a bug fix.

## How to verify any fix

- The comment fix is self-evidencing (read it).
- For 2: after building the writer, `SELECT count(*) FROM site_plan_directives WHERE scope='page'
  AND category='preserve' AND locked_by='adoption' AND lock_expires_at > NOW();` must be non-zero on
  a freshly adopted site, and a second re-plan on that site must preserve a non-deployed adopted
  page's composition.
- For either: a re-plan on a site with a current plan should log `dropped page duplicating an
  adopted item topic` **zero** times today, and non-zero only if Pass C2 is deliberately re-scoped.

## Key references

- `platform/orchestration/actions/v3_site_actions.go` — `reconcilePlanWithRealised` (:4575),
  `lockedPages` (:4582), `itemStemSets` (:4636), Pass C2 (:4684), the wrong comments (:4531, :4556).
- `docs/agent_docs/sql_for_agents/053_build_site_planner.sql` §054 (:2712) — the two-branch design.
- `docs/agent_docs/sql_for_agents/173_load_existing_pages_build_status.sql` — the live query; its own
  note already observed branch (b) was absent and called it "the minimal 054 variant".
- `/bugs_open/001` — carries the wrong premise in its diagnosis and in its "deliberately NOT done".
- `/bugs_open/037` — owns the `needs_rebuild` slice of consequence 2.
- Adjacent class: `/bugs_open/025`, `/bugs_closed/032`.
