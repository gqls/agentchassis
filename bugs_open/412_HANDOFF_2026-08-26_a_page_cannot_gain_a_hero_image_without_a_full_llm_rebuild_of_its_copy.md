# 412 — a page cannot gain a hero image without a FULL LLM REBUILD of its copy, and the item that triggers it is disguised as the light path

**Filed 2026-08-26** by the `finetuning_uk_service` lane, after the coupling blocked a legitimate
owner request: *"please also run the improvement loop over the site carefully to fix the missing
images"*, then, on being shown the cost, *"I don't want the copy in the register I rejected"*.
**Status: OPEN, unowned. Severity: MEDIUM-HIGH** — it makes a purely visual change unavailable to
any site whose copy is under review, and the rebuild it forces has a recorded history of inventing
facts.

## 1. The coupling

The only route from "this page has no hero image" to "this page shows a hero image":

```
needs_imagery (scope=page)  →  image-build-handler
   → generate → store_hero_asset → spawn_asset_deployer
   → flag_page_image_rebuild
        → item_type 'needs_page'  →  page-build-handler  →  FULL LLM REBUILD
```

`flag_page_image_rebuild_action.go:175-193` — `itemType: "needs_page"`,
`handlerAgent: "page-build-handler"`.

**There is no light route, and the reason is structural rather than an oversight.** `page-rerender`
*does* honour a landed image — `check_rerender_mode`'s condition includes
`spec.reason == 'image_landed'` and routes it to `rerender_sections`, no LLM. But nothing files a
`page_rerender` for a landed hero, because the wiring step — putting `hero_url` into the render
context from `hero_deployed.image_url` — lives inside `BuildRenderContextAction`
(`v3_site_actions.go:1358`), which only the build path runs. So the value the light path would
need is produced only by the heavy path.

## 2. ⚠ THE ITEM IS DISGUISED AS THE LIGHT PATH, and someone has already been caught by it

The emitted item carries:

| field | value | what it suggests |
|---|---|---|
| `spec.reason` | `"image_landed"` | one of the two no-LLM reasons `page-rerender` accepts |
| `item_key` | `page_rerender:<page>` | a `page_rerender` item |
| `item_type` | **`needs_page`** | …a full rebuild |
| `handler_agent` | **`page-build-handler`** | …by the writer |

Two of the four fields say "light" and the two that decide say "heavy". This is recorded in-tree
already, from the other direction — `render_news_section_html.go:39-56`:

> *"a session pointed [the news emitter] at flag_page_image_rebuild **on the belief that spec.reason
> selected a scoped no-LLM branch there. It does not.** … `needs_page` meant a FULL LLM REBUILD of
> every news page on every feed cycle — copy-regeneration roulette 4x/day. On 2026-07-24 a roll of
> that roulette re-invented two phantom links and **FABRICATED A CONTACT EMAIL** on the relojistas
> homepage."*

That comment fixed the news emitter. **The disguise itself was left in place**, so the next reader
meets it again — this lane did, today, and read `spec.reason` as protective before checking the
handler.

## 3. Why it matters beyond one site

Any site whose copy is deliberately frozen — under review, awaiting a rewrite, legally checked,
owner-approved — **cannot receive a hero image at all.** The change a site owner perceives as
purely visual is, in the platform, a full regeneration of the page's words by an LLM.

Motivating case `[MEASURED 2026-08-26]`: finetuning.uk has **9** deployed, active pages carrying
neither `hero_url` nor `background_image` (`about`, `approach`, `careers`, `case-studies`,
`contact`, `services`, `use-cases`, 2 tool pages) — and they are exactly the population that falls
to the CSS colour-band branch in `bugs_open/398`. Its other **26** pages all share one image,
`/assets/images/hero.jpg`.

## 4. Fix candidates, ordered by what closes the door

1. **Hoist the wiring out of the build path.** Make the deployed hero's URL land in
   `page_components.content_data.hero_url` at deploy time (asset-deployer or a small action), so
   the value exists independently of `BuildRenderContextAction`. `flag_page_image_rebuild` then
   emits `page_rerender` / `reason=image_landed`, which `check_rerender_mode` **already** routes to
   `rerender_sections`. No new mechanism, no LLM, and it makes the heavy path unnecessary rather
   than merely discouraged.
2. **Emit the light item when the page needs no backfill.** `needs_page` is genuinely right when a
   deferred field is absent and only the writer can supply it (the comment above says so). Decide
   per page: if the ONLY missing field is the hero URL, file `page_rerender`; otherwise
   `needs_page`. Narrower than 1, and it leaves the decision in a branch someone can get wrong.
3. **At minimum, stop the disguise.** If the item is a full rebuild, its `item_key` should not be
   `page_rerender:<page>` and its `spec.reason` should not be one of `page-rerender`'s no-LLM
   reasons. Costs nothing, fixes no behaviour, and would have saved this lane an hour.

## 5. How to verify a fix

```bash
# a page with no hero, given one, must keep its words
#   1. capture: served HTML + page_components.content_data
#   2. file needs_imagery (scope=page) for it
#   3. after the image lands, diff the copy — it must be byte-identical
```
A worked baseline of exactly this shape is committed at
`docs/agent_docs/docs024_key_docs_latest/finetuning_uk_service/baselines/2026-08-26_pre_hero_rebuild/`
(9 pages' served HTML + 66 components' `content_data`), with the 9 `needs_imagery` items that were
filed and then cancelled in `hero_items_filed.sql` — re-file them to reproduce the case.

## 6. Sources

`platform/orchestration/actions/flag_page_image_rebuild_action.go:175-193` ·
`platform/orchestration/actions/v3_site_actions.go:1358` (`BuildRenderContextAction`) ·
`platform/orchestration/actions/render_news_section_html.go:39-56` (the recorded prior victim) ·
`agent_definitions` `page-rerender` step `check_rerender_mode` (the light branch that already
accepts `image_landed`) · `agent_definitions` `image-build-handler` (`flag_rebuild` step) ·
`bugs_open/398` (the colour-band defect whose population is the same 9 pages).
