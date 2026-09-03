# 468 — `create_blog_posts` never passes `ParentSection`, so every article it writes lands under `/blog/` and no hub outside `/blog/` can ever be filled by it

**Filed 2026-09-03 by the `news_feed_ingestion` (feed) lane. UNOWNED.**
**Severity: it is the reason reviving `bugs_open/460`'s producer would not fill a section
index, even after `bugs_open/463`'s two fixes land.**

> **On the 2026-07-31 owner ruling** (a `bugs_open/` file asserting a structural cause is
> not filed until it has been through the `090` diagnosis loop, or the filing session says
> plainly why it substituted equivalent first-hand verification): **substituted, and here
> is the substitution.** The claim is a single missing struct field at named call sites,
> established by reading the code, not inferred from a symptom. I censused **every**
> non-test `CanonicalisePage` call site and read the function's own switch. The
> `bugs_open/463` lane then **independently repeated that census and confirmed it**,
> including the part that limits it. Two independent code reads agreeing on an
> enumerable, closed set is what `090` would have been asked to produce. Separately, the
> `090` trigger is a Kafka dispatch and this session's auto-mode classifier has been
> denying dispatches it judges the user did not name; that is a reason to state the
> substitution, not to skip the filing.

## 1. What the mechanism is, in plain terms

`datahelpers.CanonicalisePage` turns a role plus a slug into a page's canonical name, URL
and type. For the roles that live in a directory it takes a `ParentSection` telling it
*which* directory. Give it none and it falls back to a per-role default — for a
`blog-post`, the literal `blog`.

`create_blog_posts` is the action that writes article pages. It is the single action of
the `blog-content-planner` agent, which is the estate's second route to an article.

## 2. The defect, read in the code

`platform/orchestration/actions/create_blog_posts_action.go:196`:

```go
name, url, pageType := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
    Role: rawPageType,
    Slug: rawName,
})
```

A two-field struct literal. `ParentSection` is never set, so inside `CanonicalisePage`
`parent` is always `""`, and `page_canonical.go:217-220` takes the fallback:

```go
dir := parent
if dir == "" {
    dir = "blog"
}
return slug, "/" + dir + "/" + slug + ".html", "blog-post"
```

**So the URL is unconditionally `/blog/<slug>.html`.** Not by configuration, not by
anything the planning LLM can emit, not by a populated `site_plan_pages.parent_section` —
the field is never read on this path. The action cannot be told where to put an article.

## 3. The census — the whole closed set, verified twice

`[MEASURED 2026-09-03]` every non-test `CanonicalisePage` call site:

| call site | passes `ParentSection`? |
|---|---|
| `write_site_plan_action.go:494` | yes — `ParentSection: v.ParentSection` |
| `site_db_actions.go:314` | yes — `ParentSection: v.ParentSection` |
| `deploy_tool_action.go:736` | yes — **hardcoded** `ParentSection: "guides"` |
| `create_blog_posts_action.go:196` | **no** |
| `apply_gap_plan_action.go:340` | **no** |
| `apply_adoption_plan_action.go:547` | **no** |
| `apply_adoption_plan_action.go:855` | **no** |
| `deploy_tool_action.go:768`, `create_tool_component_action.go:399`, `write_site_plan_imagery_scope.go:101`, `check_tool_recreation_needed.go:251`, `page_identity.go:100` | n/a — roles that take no directory, or name-only uses |

Independently repeated and confirmed by the `bugs_open/463` lane the same day.

**The roles for which an absent `ParentSection` is a real defect** are the ones whose arm
reads `parent`: `tool`, `guide`, `game`, `blog-post`, `entity-page`
(`page_canonical.go:181, 192, 203, 217, 233`).

> ⚠ **`section-index` / `news-index` are NOT in that set, and a census that includes them
> overstates the case.** `CanonicalisePage` deliberately ignores `ParentSection` for the
> index family — a section index *is* its own section — so those rows carrying an empty
> `parent_section` is correct behaviour, not evidence. This file's first draft made
> exactly that mistake; caught by the 463 lane. See `WRONG_CALLS.md` 2026-09-03.

`[MEASURED 2026-09-03]` `site_plan_pages`: **109 of 109** `blog-post` rows carry an empty
`parent_section`. That figure carries the argument on its own.

## 4. Why this is not covered by `bugs_open/463`, in the 463 lane's own words

463 fixes the two defects between "plan a child page" and "a child page exists under the
prefix": Pass C deleting legitimate children, and the plan write path defaulting the
directory. **Its fix derives `parent_section` from the page's own URL inside
`ValidateRoles`.** `create_blog_posts` has no URL to derive from — it builds a page from
an LLM's blog-post spec carrying a name and a type, and **the URL is the OUTPUT of
canonicalisation, not an input to it**. There is nothing for that mechanism to read.
Recorded as a named residual in `bugs_open/463` §9; filed separately here because a
residual inside another bug is forgotten when that bug closes, which this estate has
already been bitten by (`LANDMINES`: a closed blocker keeps being obeyed).

## 5. Why it matters now rather than in the abstract

- **It is the third lock on `bugs_open/460`.** Reviving `blog-content-planner` — dormant
  since 2026-04-24, cause still unestablished — would, with 463's fixes fully landed,
  still write every article to `/blog/`.
- **It is why `designblog.co.uk`'s `/the-design-feed/` cannot be filled by that route.**
  The owner's 2026-09-03 re-scope keeps that page a `section-index` filled by child pages
  under its own prefix. Articles written to `/blog/` are not under the prefix, so the hub
  resolves zero children and the page stays empty, indistinguishably from today.
- Any future site wanting articles under `/insights/`, `/guides/` or a vertical prefix
  hits the same wall.

## 6. Fix candidates, ordered by what closes the door

1. **Take the target directory from the work item's spec and pass it.** `deploy_tool_action.go:736`
   is the precedent inside this same helper — it hardcodes `ParentSection: "guides"` and
   ships tool guides under `/guides/` today, so the field is proven end to end. The
   `needs_blog_posts` work item is where a caller would name the section. Makes the bad
   state unrepresentable for this producer.
2. **Default it from the hub that triggered the work.** `check_empty_blog.go` files the
   item because a blog/section index exists; that page's own URL stem is the directory its
   children belong in. Removes the need for anyone to remember to set it.
3. (Weakest) **Leave it and require callers to post-correct the URL.** Rejected on the
   estate's own rule that "operators must remember X" is a defect, and because the row is
   already written by then.

Whoever takes it should decide the same question for `apply_gap_plan_action.go:340` and
`apply_adoption_plan_action.go:547`/`:855`, which have the same shape. They are listed
here as the same class, **not** as measured damage — I have not checked what roles reach
those two in practice, and a fix should not assume they need it.

## 7. How to verify a fix

Not at the plan and not at the status. At the row, then the page:

```sql
-- after driving the producer for a site whose hub is NOT /blog/
SELECT url, page_type FROM pages
 WHERE site_id = :site AND page_type = 'blog-post'
 ORDER BY created_at DESC LIMIT 10;
```

Every URL must sit under the intended prefix. Then confirm the hub actually lists them,
which is a **different** mechanism and a different bug — `bugs_open/457`
(`rebuild_blog_listing` appending orphan `page_components` rows) decides whether a filled
hub renders its children. A filled hub that renders nothing looks exactly like an empty
one, so do not read a rendered-empty hub as this fix having failed.

## 8. Ownership / routing

**Unowned.** Found by the `news_feed_ingestion` lane while re-costing the routes for
`designblog.co.uk`'s `/the-design-feed/`; not taken, because this lane owns feed
ingestion rather than page production and the revival it unblocks belongs to whoever takes
`bugs_open/460`. Related: **460** (the dormancy), **463** (the two plan-path defects, in
flight), **457** (the hub render half), **444** (the listing-page gate that holds a
childless hub).
