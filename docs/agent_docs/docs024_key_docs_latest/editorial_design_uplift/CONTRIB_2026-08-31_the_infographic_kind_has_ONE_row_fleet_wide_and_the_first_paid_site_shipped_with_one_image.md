# CONTRIB 2026-08-31 — the `infographic` kind has ONE row fleet-wide, ever; the first paid site shipped carrying exactly one image

**To:** editorial_design_uplift (and cc the imagery lane + `inline_guide_imagery`).
**From:** the session reviewing boxingonline.com — `d2aa5206-73bc-4707-a69c-2702c1eb9152`,
order BR-9AUZ59, first paid customer build, 2026-08-31. Owner-raised the same evening:

> "there is not enough imagery on the pages and articles … please look at why we didn't use
> e.g. infographics to take the place of much of the explanatory copy."

All figures `[MEASURED 2026-08-31]`, queries inline.

---

## 1. What the site actually serves

**Every page carries exactly one `<img>`, and it is the logo.**

```
curl -s https://boxingonline.ugg2.com<path> | grep -o '<img[^>]*>' | wc -l
  /                                        1   (/assets/images/logo.png)
  /articles/index.html                     1   (same)
  /about.html                              1   (same)
  /guides/tool-fight-countdown-guide.html  1   (same)
  /tools/fight-calendar/index.html         1   (same)
```
Four hero images exist and serve (`hero-home.jpg` etc., HTTP 200) but as CSS backgrounds behind
text, not as content. So: **13 deployed pages, ~4,400–6,200 characters of guide prose apiece,
and not one explanatory image anywhere.**

## 2. What was requested — the whole imagery vocabulary this build used

```sql
SELECT spi.scope, spi.kind, spi.key, spi.scope_ref
  FROM site_plan_imagery spi JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
 WHERE sp.site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152';
```
→ **8 rows: 4 hero, 3 icon, 1 logo. Zero illustration. Zero infographic.**

All eight `needs_imagery` work items completed successfully with assets stored. **Nothing
failed.** The imagery pipeline did exactly what it was asked for, and it was asked for chrome.

## 3. The fleet number, which is the real finding

```sql
SELECT kind, count(*) AS rows, count(DISTINCT sp.site_id) AS sites
  FROM site_plan_imagery spi JOIN site_plans sp ON sp.id=spi.plan_id GROUP BY 1 ORDER BY 2 DESC;
```

| kind | rows | sites |
|---|---|---|
| hero | 359 | 29 |
| icon | 196 | 25 |
| logo | 45 | 28 |
| illustration | 19 | 5 |
| **infographic** | **1** | **1** |
| sprite_sheet | 1 | 1 |

**One infographic has ever been planned, on one site, in the whole estate.**

## 4. The capability is BUILT and READ — it is undriven, not missing

This is the "a silent mechanism is usually UNDRIVEN, not missing" shape, and every layer below
the planner is present:

- **Generation:** `platform/orchestration/actions/generate_image_actions.go:98` carries an
  `"infographic"` config block; `imagery_style_guide.go:316` handles it as a photographic kind.
- **Provider routing:** `internal/adapters/imagegenerator/routing.go:63` routes `infographic`
  to `providerBanana`; the stability provider has its own entry (`stability/provider.go:68`).
- **Plan admission:** `write_site_plan_action.go:213` — `"infographic": true`.
- **Consumption:** `plan_sections_action.go:476-492` selects section-scope imagery
  `WHERE spi.kind IN ('illustration','icon','infographic')`, joins the deployed asset and maps
  it into the render context by key and by kind.

So a planner that wrote an `infographic` row would get an image generated, deployed and rendered
today, with no code change. **Nothing writes the row.** The question for your lane is not "can
we?" — it is "what would ever ask?"

## 5. The owner asked for this seventeen days ago and it is still at "design"

`docs/agent_docs/docs024_key_docs_latest/inline_guide_imagery/PLAN_2026-08-14_durable_inline_guide_imagery.md`
opens with:

> **The ask (owner, 2026-08-13):** guide/blog articles should carry explanatory imagery inside
> the article body — between paragraphs and beside them — not just the header image.
> **Status: design, nothing implemented.**

That plan is a good piece of work (it correctly reframes the problem as *durability*, not
capability — in-body `<figure>` markup lives inside `article-body`'s single LLM `content` field
and is destroyed by any wholesale regeneration; it measured 93 instances / 18 sites and found
three dartsonline guides that had silently lost their figures). It has not moved since.

Meanwhile the first paid customer build shipped four explanatory guide pages, 4.4–6.2 KB each,
with no in-body imagery at all. **The lane's motivating case is no longer a dartsonline guide —
it is a customer deliverable.** Worth saying so in that lane's next handoff; I have filed a
short pointer there.

## 6. Why "infographic instead of copy" is the right instinct here, concretely

The owner's specific complaint about
`/guides/tool-fight-countdown-guide.html` (4,415 chars, six headed sections):

> "The only interesting thing it says in the whole page is that the match might be advertised at
> 9 but not start until 12. The whole of the page doesn't say much more than that."

Reading it, that is accurate. The page's real content is **one timeline**: doors → prelims →
main card advertised start → undercard drift → main event ring walk, with a note that the last
one moves. That is a diagram. It is currently ~4,400 characters of prose circling the same
point, because prose is the only medium the writer has.

Same for the weight-class finder guide (6,158 chars): the subject is a **table of divisions and
limits**. And the vertical research for this site says so independently — its `avoid` list warns
against *"generic stock photography"* and asks for *fight poster art and credited photography*,
and its `design` pattern note says the vertical runs on *"magazine-grid layouts — card-based
article displays with image, headline, and short deck"*. We built card grids with `"image": ""`.

## 7. What I would ask of your lane

1. **Own the question "what would ever write an `infographic` row?"** The answer is presumably
   the section planner, at the point where it decides a section is explanatory. One row would
   double the fleet-wide total.
2. **A structural-content tell**, in your composition work: a section whose subject is a
   sequence, a comparison or a set of thresholds should carry a diagram, not only prose. The
   three cases above (fight-night timeline, weight-class table, fighter comparison) are all in
   one 13-page site.
3. **Article/listing cards with no image are the vertical's own anti-pattern.** Cheap check:
   a `content-listing` where every item's `image` is empty.
4. **Pick up `inline_guide_imagery`'s durability finding or formally park it** — the design is
   written and the customer case now exists.

I am not dispatching anything at this site; the delivery lane
(`site_delivery_and_editor`) owns its pipeline and has work in flight.
