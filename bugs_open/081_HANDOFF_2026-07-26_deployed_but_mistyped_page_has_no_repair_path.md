# 081 — A **deployed** but mistyped page has no repair path: the retype arm cannot see it and `new_page` overwrites it without fixing `page_type` (OPEN, live instance)

**Found:** 2026-07-26 while closing `bugs_closed/015`. This is the half of the
mistyped-`page_type` class that 015's fix does **not** cover, and it has a live
instance with an open work item that has already failed once.

## The gap in one line

015's repair path (`retype_existing`) only considers **stranded** pages —
nav-visible, `sections=[]`, and **`build_status <> 'deployed'`**. A page that is
mistyped *and already deployed* matches none of that, so nothing can re-type it;
and the fallback arm silently makes things worse.

## Mechanism — the two doors are both shut

**Door 1, the retype arm, cannot see it.** `findStrandedNavPages`
(`check_news_feed.go:682-692`) requires `COALESCE(build_status,'') <> 'deployed'`:

```sql
WHERE site_id = $1
  AND COALESCE(page_type, '') <> 'news-index'
  AND (COALESCE(in_header, false) OR COALESCE(in_footer, false))
  AND jsonb_array_length(COALESCE(sections, '[]'::jsonb)) = 0
  AND COALESCE(build_status, '') <> 'deployed'
```

That clause is **correct and deliberate** — it was added by 015 after the
predicate without it flagged six live, working pages on
`ai-agent-orchestration.com` as "dead links today". The point here is not that the
clause is wrong; it is that excluding deployed pages leaves the deployed-mistyped
case with no owner. With no candidates the spec carries no `retype_candidates`,
and the prompt makes approach E conditional on exactly that: *"ONLY available when
the gap description lists stranded candidate pages"*.

**Door 2, `new_page`, overwrites the live page and leaves the mistype.** The
upsert at `apply_gap_plan_action.go:394-404`:

```sql
INSERT INTO pages (site_id, name, url, title, page_type, build_status, sections, ...)
VALUES ($1, $2, $3, $4, $5, 'planned', $6::jsonb, ...)
ON CONFLICT (site_id, name) DO UPDATE SET
        title = EXCLUDED.title,
        sections = EXCLUDED.sections,
        updated_at = NOW()
```

`page_type` is in the INSERT but **not in the `DO UPDATE SET`**. So when the
planner reuses the existing name:

- the live deployed page's `title` and `sections` are **overwritten** by the LLM's plan;
- its `page_type` is **left wrong**, so it stays orphaned from the news machinery;
- a `needs_content_page` item is filed, so the deployed page gets rebuilt;
- the site still has zero `news-index` pages, so `MissingNewsPageCheck` fires
  again on the next sweep. **The loop does not converge.**

If instead the LLM picks a *different* name, a duplicate page row is created —
the outcome 015's candidate 3 exists to prevent.

## Live evidence (queried 2026-07-26)

Two pages are mistyped and deployed, so neither is repairable today:

| domain | name | page_type | url | build_status | nav |
|---|---|---|---|---|---|
| ai-agent-orchestration.com | `news` | **`content`** | `/news.html` | deployed | in_footer |
| idea.uk | `news-index` | **`section-index`** | `/news/index.html` | deployed | header+footer |

And `ai-agent-orchestration.com` has the work item that will walk into door 2:

| item_type | status | spec.page_name | spec.page_type | has retype_candidates |
|---|---|---|---|---|
| `missing_news_page` | **detected** | `news` | `news-index` | **false** |
| `missing_news_page` | `unresolved` | `news` | `news-index` | false |

`spec.page_name` is `news`, which is exactly the existing row's name — so the
`ON CONFLICT` branch is the one that fires. **The second row is the same check
from 2026-05-01 that already ran out of attempts and went `unresolved`** — this
has been looping for roughly three months, which is the strongest evidence that
door 2 does not resolve it.

## Why this was invisible until now

015's scope measurement asked *"how many sites want a separate news page and have
no `news-index` page?"* — answer, one — and then *"how many stranded candidates
does that site have?"* — answer, zero. Both true, and together they read as "no
live damage". The question neither asked was *"is there a page already doing the
job under the wrong type that is **deployed**?"* — and there is, on that very site.

## Fix candidates (none applied)

1. **Add `page_type` to the `new_page` upsert's `DO UPDATE SET`.** One line, and it
   converts the silent no-op into a repair — the conflicting row gets re-typed to
   the planner's intended type. Cheapest by far. Risk: `new_page` is generic, so
   this lets any gap plan re-type any same-named existing page; that is a real
   widening of authority and wants the fail-closed treatment `retype_existing` got.
2. **Widen the candidate predicate to a second, separate class** — deployed pages
   whose `page_type` disagrees with the role the check is looking for — and pass
   them as `retype_candidates` with a flag saying "deployed, re-render after
   re-typing". Keeps the fail-closed authorisation model 015 established. More work,
   correct shape. Note the sections-overwrite must NOT be applied blindly to a
   deployed page with real content.
3. **Detector only** — flag "site wants role X, has a deployed page occupying it
   under type Y" as a work item for a human. Weakest, but it stops the silent loop.

Candidate 2 is the honest structural fix; candidate 1 alone would repair the live
instance but hands a broad mutation power to a generic arm.

## Immediate, contained option for the live instance

Same targeted shape as 015's original relojistas workaround — re-type the two rows
by hand and let the normal renderers pick them up:

```sql
UPDATE pages SET page_type='news-index' WHERE site_id=$1 AND name='news';       -- ai-agent-orchestration.com
UPDATE pages SET page_type='news-index' WHERE site_id=$2 AND name='news-index'; -- idea.uk
```

**Not done here — it needs an owner call**, because both pages are live and
re-typing them changes their behaviour immediately: `render_news_section` starts
emitting `data/news-archive.json` for them, and `MissingNewsPageCheck` stops
firing. For `ai-agent-orchestration.com` that is probably desirable (it closes a
three-month-old loop). For `idea.uk` note it is VM-served and its news page's
`content_data` is stale/empty (`bugs_closed/026` Defect B part 1), so re-typing
alone will not make it correct.

## Related

- `bugs_closed/015` — the mistyped-`page_type` class; fixed for the *stranded* and
  *newly-planned* cases, and its closing section names this as residual 1's sibling.
- `bugs_closed/026` — routed these same two pages here ("handed to `015`'s owner");
  this file is where that hand-off actually lands.
- `bugs_open/080` — the other 015 offshoot: `new_page` bypasses `CanonicalisePage`,
  which is what makes the "different name → duplicate row" outcome above possible.
