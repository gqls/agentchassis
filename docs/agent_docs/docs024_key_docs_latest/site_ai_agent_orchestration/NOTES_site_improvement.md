# NOTES — ai-agent-orchestration.com improvement (images, carousels, contrast)

Append-only, newest at the bottom. Technical log: evidence, commands, what the system said,
and every misstep.

---

## 2026-08-17 — session opens on the owner's three asks

Owner's ask: continue from `HANDOFF_2026-08-05_rebuild_scope.md`, and improve the site —
**get images working, make the components into carousels, improve the text contrast**. Owner's
stated expectation: *"We have already fixed all these elements on other sites so it is probably
just a matter of running the improvement loop onto the site - except for the carousels."*

**That expectation is half right, and the half that is wrong is the expensive half.** Recorded
here before anything is dispatched.

### Re-measurement of the 08-05 handoff (it is 12 days stale)

| 08-05 claim | 08-17 measurement | verdict |
|---|---|---|
| 31 NULL `content_data` across 10 pages | **15 NULL across 8 pages** | partly repaired |
| 5 pages with no components | 5 pages with 1 empty component each | unchanged in substance |
| 42 queued `page_rerender` not moving | 20 `unresolved` + 2 `failed` | partly drained |
| site UNLOCKED, `locked_at` NULL | still NULL, `status='deployed'` | unchanged |

Site id `2a8ebf9c-20a2-4c39-b191-840b012371da`. Nothing in flight: every `orchestration_states`
row touching the site today reads `COMPLETED` (checked 2026-08-17 ~16:00Z). The site is busy with
scheduled automation, so re-check immediately before dispatching.

### Ask 1 — contrast. The loop will NOT fix this, and I can show it.

`scripts/render_audit.py` against 4 live pages: **47 contrast failures, 44 of them firm**
(non-`over_image`). Not marginal — the worst are **1.00:1, i.e. text painted in exactly its own
background colour**.

Only two foreground colours appear in all 44:

```
 20  fg=rgb(230, 237, 243)  bg=rgb(255,255,255)     <- LIGHT text on WHITE
 14  fg=rgb(13, 17, 23)     bg=rgb(13,17,23)        <- DARK text on DARK  (1.00:1)
  6  fg=rgb(13, 17, 23)     bg=rgb(8,11,16)
  4  fg=rgb(230, 237, 243)  bg=rgb(248,249,250)
```

`#E6EDF3` is the site's `--color-text`; `#0D1117` is its `--color-primary`. So every failure is
one of the site's own tokens painted onto a ground it cannot survive.

**[MEASURED at the browser, not inferred from source]** computed custom properties on
`/pricing.html`:

```
--color-primary      #0D1117     <- identical to --color-surface
--color-primary-ink  #768eb2     <- the legible-ink repair IS live and IS correct
--color-text         #E6EDF3
--color-background   #080B10
--color-surface      #0D1117
```

**So bugfix 122's ink mechanism is live and working on this site and the site is still broken.**
That is the finding. Two independent defects survive it:

**(A) Components paint text with the BARE token.** The culprit declaration, extracted from the
component's own embedded `<style>` in `page_components.rendered_html` (`pricing`/`differentiators`):

```css
.differentiator-item h3 { color: var(--color-primary, #1a1a2e); }
```

`--color-primary` is `#0D1117`; the section's ground is `#0D1117`. Result 1.00:1. The fix the
platform already owns is `var(--color-primary-ink, var(--color-primary, #1a1a2e))` —
`--color-primary-ink` is `#768eb2` here, which clears the floor.

⚠ **The `#1a1a2e` fallback is in the source and is NEVER applied**, because the variable is set.
This is `[[a-css-fallback-is-present-and-inoperative]]` firing exactly as written — a grep of the
stylesheet reads "dark navy heading" and the browser paints invisible. I did not trust the
literal; the colours above are `getComputedStyle` values.

**(B) Components hardcode a LIGHT ground on a DARK site.** Seven of them:

```
about | departments-grid          #fff
about | leadership-team           #fff
index | case-studies-grid         255,255,255
index | departments-grid          #fff
index | differentiators-section   #fff
index | latest-news               #fff
index | system-stats              255,255,255
```

A white card on a dark site keeps the site's light `--color-text`, so the heading vanishes.

**Why "just re-render" does not work — the disconfirming test.** If staleness were the cause, a
freshly rendered page would be clean. It is not:

| page | last render | firm failures |
|---|---|---|
| `services` | 08-15 | **0** |
| `index` | **08-17 (today)** | **17** |
| `about` | 08-11 | 19 |
| `pricing` | **2026-04-13** | 8 |

`index` was rendered today and still fails, so family (B) is immune to re-rendering. Meanwhile
`pricing` cannot be re-rendered **at all**: 5/5 components have NULL `content_data`, which is the
`bugs_closed/194` damage, and `rerender_page_sections` has nothing to rebuild from. Its 7
invisible headings are frozen there until the page is rebuilt.

**Fleet context — this site's palette is an outlier.** Of 23 sites carrying a
`design_intent.palette.reference_values`, only **2** have `primary == surface`:
ai-agent-orchestration.com and oufe.com. Every other site gives `primary` a value distinct from
its own surface; the healthy dark sites give it a genuinely light one (fundamentallyai `#86ADDE`
on `#111E33`, vonc `#7c3cff` on `#13121f`).

### Ask 2 — images. One component, and the live handlers do not do what the name suggests.

Every `<img>` on the whole site is `case-studies-grid`, and there are only ten:

```
enterprise-reference-deployment | case-studies-grid | /assets/images/case-study-*.png   (x5, all HTTP 404)
index                           | case-studies-grid | (EMPTY src)                       (x5)
```

`content_data` for that component is rich — five card titles, excerpts, stats and genuinely good
`cardN_image_alt` text — but there is **no `cardN_image_url` key at all**, which is why `src`
renders empty. The site's own `image_source_unsatisfiable` items say it in as many words:
*"sources field 'card1_image_url' from site_assets.image which nothing generates"*.

⚠ **MISSTEP AVOIDED — do not route these at the obvious handler.** The site's 6 `image_url_404`
and 3 `image_source_unsatisfiable` rows carry an **empty `handler_agent`**, which is why nothing
dispatches them, and the tempting fix is to fill it in: `image-url-404-handler` and
`image-source-unsatisfiable-handler` are both live. I checked what they actually do first.

- Their workflows are `query_database` → `create_work_item` → `checkpoint_for_review`.
  **Neither generates an image.** They triage.
- The only site they have ever run against is `mortgagecalculator.co.uk` (2026-08-14): 3 + 2
  `complete`, 15 `cancelled`. That site now has **zero `<img>` tags in any component**.

So the precedent for "run the image handler" is a site whose images were *removed*. Filling in
`handler_agent` here would most likely strip the five case studies rather than illustrate them —
the opposite of the ask. Real image generation lives at `image-generator` / `image-build-handler`.

⚠ **Separate, and time-critical: the existing image assets are expiring URLs.** All 9 hero /
content_hero rows in `assets` are pre-signed Backblaze URLs carrying `X-Amz-Expires=604800`
(7 days) stamped `20260811T16-17Z`. That is **2026-08-18**, i.e. tomorrow. `og-card.png` and
`favicon.png` are the only two stored as stable `/assets/images/` paths. Not yet established
whether anything still serves the pre-signed form — no page component references one today
(the img census above finds none), so the blast radius may be nil. **[UNVERIFIED]** whether any
other consumer (og tags, feeds, the asset renderer) holds one.

### Ask 3 — carousels. Nothing exists; owner's instinct is right.

- No carousel/slider component exists. `grep -rli carousel platform/ internal/` returns only two
  substantive hits, and neither is a component.
- `html_actions.go:527` already carries the guidance *"For carousels/sliders: Use CSS animation,
  NOT complex JavaScript"* — but it sits in a **whole-page** generation prompt, not the
  component-level path this site builds through.
- Prior art worth reading before designing one: `bind_site_experience_action.go:36` cites
  *"the four dead carousel destinations found by hand on 2026-07-26"* (`bugs_open/023`, `071`) —
  carousels here have previously shipped with CTAs pointing at pages that do not exist. Whatever
  hint gets written must not be able to promise a destination that is not in `pages`.

### Where that leaves the owner's premise

Stated plainly, because it changes the plan: **only the carousel half of the owner's expectation
survives contact with the measurement.** Contrast was not "already fixed on other sites" in a way
that transfers — the mechanism that fixed other sites is *already live here* and this site is
still broken, for two reasons that mechanism was never meant to catch. Images were not fixed
elsewhere either; the one site that ran the image handlers ended up with no images.
