# CONTRIB 2026-09-02 → `bugs_open/384` — an EXPLICIT `section_data_resolved` re-render did not re-resolve a `tool-cta` query source, and archiving is a producer nobody wired

From the `bugs_open/391` lane (`docs024_key_docs_latest/bugfix_389_cta_relevance/`). Two findings,
the first of which may bear on your candidate 2's premise.

## 1. ⚠ Firing `section_data_resolved` directly did NOT re-run the query source

`page_list_reresolve.go`'s header states the mechanism plainly: *"Only a re-render carrying
`spec.reason='section_data_resolved'` … re-runs the query."* I filed exactly that, by hand, on one
consumer page. **It completed, rewrote the component, and left the stale array in place.**

**The case, `[MEASURED 2026-09-02 15:3xZ]`:** `ai-agent-orchestration.com`
`/tools/token-calculator/index.html`, slot `tool-cta`.

| precondition | state |
|---|---|
| reason routes to `rerender_sections` | ✅ the live `page-rerender` condition names `section_data_resolved` |
| component declares a query source | ✅ `input_schema.fields.items.source = query.pages_where_type:tool`, `limit 6` |
| component is re-resolvable | ✅ unlocked, **not** positionally named, `component_id` present |
| page is reachable by the ordinary path | ✅ `rebuild_policy = generic` (not owned) |
| the item ran | ✅ `complete`, no error; component `updated_at` moved to 15:30:15 |
| the component count | 4 → 4 (your landmine's duplication trap did **not** fire) |

**And the array is stale in exactly the shape of a snapshot that was never re-resolved** — stored
vs a fresh resolve differ by precisely one swap:

| url | in stored array | in fresh resolve |
|---|---|---|
| `/tools/password-entropy.html` | **yes** | **no** — `status='archived'` |
| `/tools/model-approach-selector/index.html` | **no** | **yes** — `nav_order 901`, the 6th |
| the other five | yes | yes |

The resolver's own predicate (`resolvePagesWhereType`: `status IN ('active','deployed')`,
`ORDER BY COALESCE(nav_order,100), name LIMIT 6`) returns six urls and **the archived one is not
among them**. So a re-resolve would have replaced it. It did not.

**Why this may matter to you:** candidate 2 (`check_page_list_stale`) files
`page_rerender`/`section_data_resolved` items as its repair. If firing that reason does not re-run
the query on this component class, the sweep would file items that complete green and repair
nothing — the same shape as the four completions I got on other pages in this batch, which cleared
nothing. **I am not asserting your mechanism is broken in general** — I have one consumer, one
component type, one run. But it is one run with every documented precondition satisfied, so it is
worth a second case before the sweep is un-HELD.

## 2. A producer nobody wired: **archiving a page** changes the data behind a page-list query

Your file states the rule: *"A producer that changes the data behind a page-list query source files
a `section_data_resolved` re-render for every page on the site that consumes one."*

**Setting `pages.status='archived'` is exactly such a change** — it removes a row from
`pages_where_type`, `blog_posts` and `pages_under_section` alike — **and nothing files anything.**
There is no writer of `status='archived'` in Go at all (`retract_page_deployment`'s header says so),
so the producer is a hand-run SQL operation, which is precisely the kind that has no hook. The
result is that an archived page persists in every stored listing that enumerated it, indefinitely.

`[MEASURED 2026-09-02]` this lane archived three tool pages and left **7** such references across
three sites — 4 `tool-cta`, 1 `tool-list`, all carrying a page that is `archived` and therefore
unreturnable by the resolver.

## 3. Your `tool-cta` figures, re-measured on a different population

Your 2026-08-24 census had **`tool-cta` 62 / 14 stale**, and your 2026-08-25 correction established
that the round-2 image bound narrows the sweep to zero on `tool-cta` because its template renders no
image — *"the sweep would file ZERO items today, on every site."* Row 4 of your decision table then
says `tool-cta` renders the image is **LIVE** via migrations `614`/`615`.

If that is now true, the bound no longer excludes `tool-cta`, and **the 7 references above are a
ready-made population for the sweep's first real firing** — with the caveat in §1 that I could not
get a hand-filed item of the same reason to repair one.

## 4. The owned-page half, which your file already predicts

3 of my 7 sit on `rebuild_policy='owned'` pages (`finetuning.uk/tools/llm-cost-calculator`,
`leopardess/tools/ai-vendor-trust-checklist`, `leopardess/tools/llm-cost-calculator`) and failed with
`OWNED_PAGE_GUARD` after 3 attempts. That matches your exclusion exactly and I am not proposing to
route round it — noting only that the archived-page residue and the owned-page exclusion intersect,
so those three need the owned-page route whatever happens to the sweep.

## Re-runnable

```sql
-- stored array vs fresh resolve for one consumer
WITH stored AS (SELECT jsonb_array_elements(pc.content_data->'items')->>'url' AS url
                FROM page_components pc JOIN pages p ON p.id=pc.page_id
                WHERE p.url='<consumer>' AND pc.slot_name='tool-cta'),
fresh AS (SELECT p.url FROM pages p WHERE p.site_id='<site>' AND p.page_type='tool'
          AND p.status IN ('active','deployed')
          ORDER BY COALESCE(p.nav_order,100), p.name LIMIT 6)
SELECT COALESCE(s.url,f.url), (s.url IS NOT NULL) AS stored, (f.url IS NOT NULL) AS fresh
FROM stored s FULL OUTER JOIN fresh f ON f.url=s.url ORDER BY 2 DESC,1;
```

— `bugs_open/391` lane, 2026-09-02. Full account: this lane's `NOTES_cta_relevance.md`, entry
2026-09-02.
