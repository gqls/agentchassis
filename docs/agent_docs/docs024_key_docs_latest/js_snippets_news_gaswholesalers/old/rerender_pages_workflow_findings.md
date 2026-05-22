# rerender-pages workflow — findings from 2026-05-19 investigation

Context: investigating how to cleanly rerender gaswholesalers.com after CSS
update, specifically ensuring news.html and index.html pick up the new
news component CSS classes.

## Finding 1: `rebuild_blog_listing` does NOT handle news-index pages

The `rebuild_blog_listing` step in the `rerender-pages` workflow v6 calls
`RebuildBlogListingAction`, which uses `findBlogPage` to locate the page
to rebuild. `findBlogPage` has two strategies:

```go
// Strategy 1: canonical page_type
SELECT id, name FROM pages
WHERE site_id = $1 AND page_type = 'blog-index'

// Strategy 2: page named 'blog' created as content type
SELECT id, name FROM pages
WHERE site_id = $1 AND name = 'blog' AND page_type = 'content'
```

Neither matches `page_type = 'news-index'` or `name = 'news'`.

**Consequence for sites with news (not blog):** the step is a silent no-op.
It logs `"No blog page found, skipping"` and returns. The news page is
still re-rendered as a regular `page_rerender` work item later in the
flow, so news visuals do get updated. But there is no equivalent
news-listing rebuild that would refresh the data shape if news feed
content changed.

**For future work:** the news-listing component renders client-side from
data files in `gaswholesalers.com/data/`, so a server-side rebuild may
not be needed. But if it ever is — e.g. to bake the latest items into
HTML for SEO — we'd need a parallel `rebuild_news_listing` action with
a `findNewsPage` helper that mirrors `findBlogPage` for news-index page
type and the `news-listing` component slot priority.

**Verification query:**
```sql
SELECT name, page_type FROM pages
WHERE site_id = '<site_id>'
  AND (name = 'news' OR name = 'blog'
       OR page_type IN ('blog-index', 'news-index'));
```

If results show `news-index` page type, this site is in the gap.

## Finding 2: JS snippets refresh is coupled to site components refresh

The workflow has three refresh phases tied behind one flag:

```
refresh_site_components=true →
    render_site_components →   (re-renders header, footer, head)
    render_js_snippets →        (builds snippets.js)
    deploy_js_snippets →        (commits snippets.js to git)
    rebuild_blog_listing → ...

refresh_site_components=false →
    rebuild_blog_listing → ... (skips all three of the above)
```

Conceptually these are independent:
- Site components are header/footer/head HTML stored in content_components
- JS snippets are a single concatenated file in the sites repo
- CSS is handled entirely by the webdesign-agent (not by rerender-pages)

A workflow consumer might want to refresh just JS (e.g. after editing
the js_snippets table) without touching site components. Currently they
have to pass `refresh_site_components=true` and accept the header/footer
re-rendering as a side effect.

**Future improvement (note for backlog):** split the single flag into
three:
```
refresh_site_components: bool
refresh_js_snippets:     bool
refresh_blog_listing:    bool   (currently always runs)
```

This is a workflow JSON edit in `agent_definitions` plus a small step
re-ordering. Low effort, modest value, do it next time we touch
rerender-pages versions.

## Finding 3: Migration C concern about rerender-pages was a false alarm

In an earlier session I noted that the rerender-pages workflow used
`input_data.site_id` directly rather than going through an
`ensure_site_record` step like other workflows. I'd flagged this as a
patch-needed item.

On closer inspection of v6:

| Step | site_id source | Issue? |
|---|---|---|
| check_refresh_components | conditional on refresh flag | none — doesn't need site_id |
| render_site_components | `input_fields: ["site_id", "domain"]` | extracted from input_data — fine |
| render_js_snippets | `site_id_field: "input_data.site_id"` | fine |
| deploy_js_snippets | `domain_field: "site_record.domain"` | needs site_record but only runs when refresh=true; verified working in earlier runs |
| rebuild_blog_listing | `site_id: "input_data.site_id"` | fine |
| get_pages | `input_fields: ["site_id", "domain"]` | fine |
| create_rerender_items | `site_id: "rerender_pages.site_id"` | derived from get_pages output |

The original patch I remembered was for a different workflow (the
single-page rerender invoked when content_components were modified by
update_site_content). That workflow had a real field-path bug that was
fixed. rerender-pages v6 is correctly set up for its current usage and
does not need this patch.

Note: this clarification matters because seeing "rerender-pages doesn't
use ensure_site_record" might trigger an instinct to patch it — don't,
the workflow works as designed.

## Trigger recommendation for clean full rerender

```json
{
  "action": "orchestrate",
  "config": {"agent_type": "rerender-pages"},
  "input_data": {
    "site_id": "<uuid>",
    "domain": "<domain>",
    "refresh_site_components": true
  }
}
```

Setting `refresh_site_components: true` produces the most complete
refresh:
- Header/footer/head re-rendered fresh from content_components
- snippets.js rebuilt and committed to git
- Each page re-assembled with current site components and current
  page_components
- One git commit per page

Set to `false` only if you specifically want to skip the site-level
refresh (e.g. you JUST refreshed it and don't want the second pass).
