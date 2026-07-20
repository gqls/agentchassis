# Handoff — a section naming a component that does not exist renders an empty stub, and the build calls it success

**Filed 2026-07-20**, found while verifying `/bugs_open/001` live. Two related things: a **naming
convention that is easy to get wrong** (and that anyone comparing plans to pages must know), and a
**defect** — a section name that resolves to nothing produces a hollow `<section>` on a deployed
page instead of failing.

## Part 1 — the convention (read this before comparing a plan to a page)

`pages.sections` stores the component's **`function`**. `page_components` reference the component
**row**, whose `name` is usually different:

```sql
SELECT name, function FROM content_components
WHERE name IN ('about-hero','differentiators-section');
--  about-hero              | hero-about
--  differentiators-section | differentiators
```

So a page whose `sections` read `["hero-about", …, "differentiators", …]` renders
`page_components` `about-hero, …, differentiators-section, …`. **That is correct and matching**, not
drift — but compared naively it looks like the page rendered the wrong components. This cost real
time on 2026-07-20: `about` on dartsonline.com was briefly read as a regression when it was fine.

Names are also normalised before lookup — `NormalizeComponentFunction`
(`platform/orchestration/actions/component_validation.go:78`) converts `call_to_action` →
`call-to-action` and `SocialProof` → `social-proof`. So a literal string comparison against
`content_components` is wrong in *two* ways; normalise, then match on `function`.

Fleet-wide, after normalising (743 section entries on active pages):

| | count |
|---|---|
| resolve by `function` (the convention) | 724 |
| resolve only by `name` (namespace mixed) | 8 |
| **resolve to nothing** | **11** |

The 8 name-not-function entries (`differentiators-section`, `services-hero`, `contact-hero`,
`case-studies-hero`, and one literal `faq section` **with a space**) are the LLM writing a component
*name* where a *function* belongs. They happen to work only when a component's name and function
coincide, so they are latent, not currently broken.

## Part 2 — the defect: an unresolvable section renders a hollow stub

The 11 that resolve to nothing do not fail. They render an empty section, and the page deploys.

**`finetuning.uk` `/ai-guides`, `build_status='deployed'`:**

```sql
SELECT sections FROM pages WHERE name='ai-guides' AND site_id=(SELECT id FROM sites WHERE domain='finetuning.uk');
-- ["hero", "featured_article", "category_section", "article_grid", "testimonials", "call_to_action"]
```

```
 comp           | function       | position | rendered_html len
 hero           | hero           | 1        | 2368
 (null)         | (null)         | 2        | 208
 (null)         | (null)         | 3        | 208
 (null)         | (null)         | 4        | 208
 testimonials   | testimonials   | 5        | 3245
 call-to-action | call-to-action | 6        | 2078
```

Positions 2–4 are `featured_article`, `category_section`, `article_grid` — none of which exist as a
component under any normalisation. Each produced a `page_components` row with **`component_id IS
NULL`** and this 208-byte body:

```html
<section id="…" class="section section--generic">
  <div class="container">
    <h2 class="section__title"></h2>
    <div class="section__content"></div>
  </div>
</section>
```

An empty heading and an empty content div. On a guides index page, the three missing slots are the
featured article, the category section and the article grid — i.e. **the entire point of the page**.

**Scale**: 7 stub slots (all exactly 208 bytes) across deployed pages on 3 sites; `finetuning.uk`
and `ai-agent-orchestration.com` carry most of it. A separate 35 orphan slots have
`component_id IS NULL` but real content (842–12,592 bytes) — that is a **different** situation
(content written without a live component link) and is NOT this bug; do not conflate them.

```sql
SELECT CASE WHEN length(pc.rendered_html) < 500 THEN 'stub' ELSE 'has content' END, count(*)
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE pc.component_id IS NULL AND p.build_status='deployed' GROUP BY 1;
--  has content | 35
--  stub        | 7
```

## Why it survives

It IS detected. `check_empty_sections.go:136` has an `empty_heading` rule
(`<(h[1-6])[^>]*>\s*</\1>`) that these stubs match exactly, and `empty_section` items exist —
including one literally named `hero-tool`, one of the unresolvable names:

```
finetuning.uk | empty_section | unresolved | [stale: triaged 48h+] Empty section 'hero-tool' on page llm-cost-calculator
```

**Every one of them is `unresolved`**, most stamped `[stale: triaged 48h+]`, newest 2026-05-01. So
this is not a detection gap — it is the same delivery gap as `/bugs_open/023` and `/bugs_open/033`:
the finding is raised, nothing consumes it, it ages out, and the hollow section stays live for
months. Fixing the emitter below without fixing that routing just produces more unresolved items.

## Fix candidates

1. **Fail loudly at resolution.** When a section name resolves to no component, the page build
   should refuse that page (or emit a `needs_new_component` / `needs_human_review` item naming the
   unresolvable section) rather than writing a `component_id IS NULL` row and reporting success.
   This is the platform's recurring shape — CLAUDE.md's "trust the rendered artefact, not the
   status" — a build that produced three hollow sections should not be `complete`.
2. **Validate at plan time.** `validate_site_plan` already normalises and strips section names; have
   it check each proposed section resolves to a real component function and drop/flag the ones that
   do not, so an unbuildable name never reaches `pages.sections`. Cheaper and earlier, but will not
   help the 11 already stored.
3. **Normalise the namespace.** Accept a component `name` as well as a `function` at resolution
   time (which would fix the 8 latent ones), or stop the LLM proposing names by giving it only
   functions in the prompt's component list. Check what `load_components` currently passes — if it
   offers `name`, the LLM is being invited to make exactly this mistake.
4. **Clean up the 7 existing stubs** — but only after 1 or 2, or they will come back on the next
   rebuild.

## How to verify a fix

1. Plan a page with a deliberately bogus section name. Assert the build does **not** report
   `complete` with a hollow section — it refuses or raises an actionable item.
2. Re-run the fleet query above: `stub` count should reach 0 and stay there after a rebuild.
3. Assert the 8 name-not-function entries still render (do not fix this by breaking them).
4. Check the artefact, not the row: fetch the live page and confirm no empty `<section>` remains.

## Related

- `/bugs_open/023`, `/bugs_open/033` — the delivery gap that lets the detected items rot.
- `/bugs_open/038` — a fix there needs to compare plan sections to `pages.sections` correctly; Part 1
  is the reason a naive comparison would false-negative.
- `/bugs_open/001` — where this was found; its "VERIFIED LIVE" section records the near-miss.
- `docs024_key_docs_latest/empty_sections_loop_integrity/` — the existing workstream on this class.
