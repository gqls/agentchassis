# 080 — `content-gap-planner`'s `new_page` bypasses `CanonicalisePage`, so two creation surfaces disagree on a page's name and URL (OPEN)

**Found:** 2026-07-26, while doing the council gate's blast-radius homework for
`bugs_closed/015`. **Not** caused by 015's fix — it predates it and is identical
for `blog-index`. Filed separately because it is the exact failure class
`CanonicalisePage` was created to end, and 015's closing evidence happens to
contain a clean live example of it.

## Mechanism

The platform has two surfaces that create `pages` rows, and only one of them
canonicalises.

- **Planner path** — `write_site_plan_action.go:276` and `site_db_actions.go:281`
  both run the descriptor through `datahelpers.CanonicalisePage`, which maps the
  section-index family (`section-index`, `blog-index`, `news-index`,
  `entity-directory`) to `name=<section>-index`, `url=/<section>/index.html`.
- **Gap-planner path** — `apply_gap_plan_action.go` `applyNewPage` never calls it.
  It lowercases the LLM's name and synthesises the URL by hand:

```go
// apply_gap_plan_action.go:294
pageName = strings.ToLower(strings.ReplaceAll(pageName, " ", "-"))
// apply_gap_plan_action.go:355
url := "/" + pageName + ".html"
```

So for the same logical page the two surfaces produce different rows.

## Live evidence (queried 2026-07-26)

Three sites hold a `news-index` page. Two were created by the gap-planner, one by
hand:

| domain | name | url | how it was created |
|---|---|---|---|
| gaswholesalers.com | `news` | `/news.html` | gap-planner `new_page` |
| robot-hands.com | `news` | `/news.html` | gap-planner `new_page` |
| relojistas.com | `noticias-index` | `/noticias/index.html` | hand-fixed (015 workaround), i.e. the canonical shape |

A planner-emitted news listing on either of the first two canonicalises to
`news-index` at `/news/index.html`, which does **not** match the existing `news`
at `/news.html`.

## Why it matters

`pages` has `ON CONFLICT (site_id, name)`. A name that disagrees does not
conflict — it **inserts a second row**. That is precisely the duplicate-page
failure `CanonicalisePage`'s own header documents itself as having been written
to prevent:

> *"Without this convergence, both surfaces upserted into pages with disagreeing
> names, the `ON CONFLICT (site_id, name) DO UPDATE` never fired, and we got
> duplicate rows plus parallel work-item streams that `idx_swi_dedup` could not
> collapse."* — `page_canonical.go` header, Phase 0 (doc 029)

The convergence was built for adoption-vs-planner and the gap-planner surface was
never brought into it.

## Scope and severity — measured, not assumed

**No duplicate exists today.** [VERIFIED 2026-07-26] The two `/news.html` sites
are the only exposure, and neither has been re-planned since. This is a latent
trap, not live damage — which is why it is filed rather than hot-fixed.

**Reversal trigger:** a re-plan of `gaswholesalers.com` or `robot-hands.com` that
emits a news listing. After migration `206_planner_news_index_page_type.sql`
(applied 2026-07-25) the planner *can* now emit `news-index`, so the trigger is
live where before it was unreachable. Watch for a second news page appearing
beside the existing one.

## Fix candidates (none applied)

1. **Route `applyNewPage` through `CanonicalisePage`** — the structural fix, and
   the one consistent with the helper's stated purpose. Risk: it changes the
   name/URL of *future* gap-created pages, so existing rows must be left alone
   (no data migration) or the two sites above get re-pointed deliberately.
2. **Canonicalise on read** where the two surfaces meet, leaving creation alone.
   Cheaper, but leaves the divergence in the data.
3. **Leave it and add a detector** — a check for "two pages on one site whose
   canonical form collides". Weakest, but it converts a silent duplicate into a
   work item.

Candidate 1 is the honest one; it needs a decision about the two existing rows,
which is why this is filed rather than fixed.

## Related

- `bugs_closed/015` — where this was found; its §"CanonicalisePage callers"
  enumeration is the grep evidence.
- `page_canonical.go` header — the design intent this violates.
- `bugs_open/052`, `bugs_open/049` — listings/links advertising pages that do not
  resolve; a duplicated page row is one way to feed them.
