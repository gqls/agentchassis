# PLAN — `bugs_open/175`: close the "upsert drops `page_type`" class, not four instances

**Opened:** 2026-08-02. **Bug:** `bugs_open/175` (filed 2026-08-01 by the
`bugfix_081` lane at the council gate's request; status was OPEN, unowned).

## What the bug is

Four `pages` upserts name `page_type` in the INSERT and **not** in the
`DO UPDATE SET`. On a name collision the arm silently stops being a CREATE and
becomes a PARTIAL update: the new content lands under the OLD role, `RETURNING id`
returns an id either way, and the caller cannot tell which happened.

| site | role written | `DO UPDATE SET` |
|---|---|---|
| `create_report_page_action.go:164` | `report` | url, title, sections |
| `deploy_tool_action.go:376` | `tool` | url, title, sections |
| `deploy_tool_action.go:514` | `blog-post` (companion guide) | title |
| `create_tool_component_action.go:416` | `blog-post` (companion guide) | title |

`bugs_closed/081` is the same shape on the gap-planner's `new_page` arm; it was
fixed there alone, and the `bug_historian` seat's objection ("this plan does not
claim to audit siblings … it should not be read as closing the class") is what
produced 175.

## Validity re-checked before starting (2026-08-02)

- All four sites still carry the shape — re-grepped
  `ON CONFLICT (site_id, name)` today, quoted above from the live tree.
- The census in 175 was **incomplete**, and that matters for scope. Today's grep
  finds five more `DO UPDATE` arms it does not list: `site_db_actions.go:1141`,
  `create_blog_posts_action.go:219`, `adopt_verbatim.go:470`,
  `apply_adoption_plan_action.go:532` (175 has this one) and
  `cmd/webdesignport/import.go:182`. **Every one of them already carries
  `page_type = EXCLUDED.page_type`**, so they are in the *opposite* camp 175
  describes — they re-type on collision rather than dropping the type. They are
  out of scope here, for 175's own stated reason: the two failure modes are
  opposite and "make all six identical" is the wrong resolution.

## Exposure — measured today, and it is a real surface rather than a live fire

175 left this `[UNMEASURED]` and said the next thread should measure before
choosing a fix. Run against the live DB, 2026-08-02:

```sql
-- names the constant-role arms would claim, held under a DIFFERENT page_type
SELECT 'guide-arm', s.domain, p.name, p.page_type, p.build_status
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.name LIKE '%-guide' AND COALESCE(p.page_type,'') <> 'blog-post'
UNION ALL SELECT 'tool-arm', ... WHERE p.name LIKE 'tool-%' AND page_type <> 'tool'
UNION ALL SELECT 'report-arm', ... WHERE p.name LIKE 'report-%' AND page_type <> 'report';
```

| arm | domain | name | page_type | build_status |
|---|---|---|---|---|
| guide-arm | robot-hands.com | `gripper-selection-guide` | content | **deployed** |
| guide-arm | robot-hands.com | `selection-guide` | content | **deployed** |
| report-arm | idea.uk | `report-example` | content | **deployed** |
| report-arm | lendzy.co.uk | `report-loan-shark` | content | **deployed** |

Four live, deployed pages sit on names these arms claim. **A collision has not
been observed** — it needs a tool whose page name is exactly `gripper-selection`
or `selection`, and the report arm names pages `report-<uuid>` so its two rows are
unreachable in practice. So: the surface is real and one arm's is plausible
(robot-hands.com is a gripper site), the fire is not lit. `[MEASURED 2026-08-02]`.

## The fix — candidate 2 from 175, one seam instead of four patches

175's candidates were (1) per-call-site patches, (2) a shared
`upsertPageForRole`, (3) a detector. The owner's standing instruction is a robust
framework fix over the individual case, and 175 itself says (2) is "the only
candidate that stops a seventh arm being written with the same bug".

`UpsertPageForRole` (new file `platform/orchestration/actions/page_role_upsert.go`)
owns the whole write for **an arm whose role is a compile-time constant**, and
answers the collision in one place:

| collision | answer | why |
|---|---|---|
| no collision | CREATE with every declared column | unchanged |
| row holds the SAME role | refresh the caller's declared `Refresh` subset | this is what the arms did, and it is right — the page is doing this arm's job |
| row holds a DIFFERENT role and **has been live** | **REFUSE.** mutate nothing; file `mistyped_deployed_page` (`needs_human_review`); tell the caller | `081`'s answer. Re-typing a live page changes what it serves the instant it happens, and 081 measured that no predicate can tell a real listing from a hub that embeds one |
| row holds a DIFFERENT role and has **never** been live | **ADOPT.** take the row over completely — every declared column *including* `page_type` | nothing is served, and the arm's role is a constant it owns by construction. Leaving the type wrong IS the defect |

Two decisions here are deliberately **not** copies of 081, and both are recorded
in NOTES with their evidence:

1. **"Has been live" is `build_status IN ('deployed','needs_rebuild') OR
   deployed_at IS NOT NULL`, not `build_status = 'deployed'`.** 081 used the
   narrow test. `bugs_closed/037` is a whole case about `needs_rebuild` falling
   outside a `= 'deployed'` guard, and the live census says 35 of 46
   `needs_rebuild` rows have a non-null `deployed_at` — they *have* shipped.
2. **Adoption is a WHOLE takeover, not a partial one.** A partial adopt (type but
   not url, say) would mint a fresh hybrid — the same defect wearing different
   columns. Whole takeover of an unshipped row is the only shape that cannot
   leave a half-claimed page behind.

## Prevention — so a seventh arm cannot be written

A `scripts/pattern-check.py` rule: any `ON CONFLICT (site_id, name) DO UPDATE`
whose `SET` list omits `page_type` while the INSERT names it. That is the exact
signature of the class, it is cheap, and it fires at commit time on the file where
the mistake is made. Without it, the helper only fixes today's four.

## Out of scope, stated rather than silently dropped

- `apply_gap_plan_action.go` (081's arm) is **not** converted. Its role comes from
  an LLM plan, not a constant, so the ADOPT branch above would hand a generic arm
  the authority 081 argued it must not have. It keeps its own resolver. What IS
  shared is the refusal *filing*, extracted so both produce one item shape.
- The five `page_type = EXCLUDED.page_type` arms (adoption, blog posts, site sync,
  verbatim adoption, the port CLI) — 175 says explicitly not to make them
  identical, and the adoption path's authority to re-type is arguably correct.
- A page whose collision row is `status='archived'` keeps today's behaviour on the
  same-role path (refresh, stays archived). Logged as a warning so it is visible;
  changing it is a different bug (`bugs_open/098` territory).

## Verification bar

Per 175 and 081: induce both branches, and **break the guard and watch the test
fail** — a refusal-only test is satisfied by a helper that refuses everything, and
`mock.ExpectationsWereMet()` is not "no database call happened" (LANDMINES).
