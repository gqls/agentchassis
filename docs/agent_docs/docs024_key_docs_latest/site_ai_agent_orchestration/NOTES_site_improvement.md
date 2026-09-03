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

---

## 2026-08-17 (later) — owner picked "both routes"; A1 then FAILED its own safety check

Owner chose route **A1 + A2** for contrast, and approved the `pricing` framework rebuild.

### ⚠ MY OWN OPTION A1 WAS UNSAFE, AND THE CHECK THAT CAUGHT IT TOOK ONE QUERY

I offered A1 ("give this site's `primary` a visible value — one spec row") as the fast, safe,
site-scoped half. **It is not safe, and I should have measured before offering it.** Before
editing the palette I counted how `--color-primary` is actually consumed:

```
component CSS (rendered_html, this site)      site stylesheet
  color         37                              color         2
  background    24     <---- the problem        background    2
  border-color   6                              border-color  1
  accent-color   4
```

`--color-primary` is **dual-role**: 37 foreground uses and 24 *background* uses. Lightening it so
the headings read would put a light fill under the white/near-white labels that sit on those
fills — trading 20 failures for a fresh set. That is `render_audit.py`'s own defect family 2, *"a
token used in two roles — correct in one place, invisible in the other"*, which I had read in that
file's header earlier the same session and still proposed a fix that walks into it.

**There is no safe single-value palette fix here.** Foreground uses need a light colour on these
dark grounds; background uses need a dark colour under white labels. One token cannot be both —
which is precisely why the ink companion exists. **A1 is withdrawn; A2 is the whole fix.**
Recorded in `WRONG_CALLS.md`.

### A2, done properly — migration 456

`content_components.html_template` is the source: the rule in the template is **byte-identical**
to the one in the served `rendered_html`, checked rather than assumed.

Fleet scope of the defect, measured: **156 of 294 templates** carry a bare foreground
`--color-primary`; only **4** mention the ink companion at all. This site renders **12** of them.
Migration `456` repoints those 12, named explicitly so a concurrent placement cannot shift the set.

Dry-run simulation before applying (read-only, the whole point being that it could have come out
otherwise): 12 rows, 12 changed, **36 bare → 0**, 36 wrapped, **0** damage to
`--color-primary-text` (a different token, the label that sits ON a primary fill). Applied:
`UPDATE 12`, all five guards passed.

Shape is 415's: two-level fallback `var(--color-primary-ink,var(--color-primary,#X))`, never bare,
so it is **inert** where no companion is emitted and corrective where one is.

### The fix is applied at the template and CANNOT REACH THE SITE — known open bug

`rerender-pages` does **not** render. It files one `page_rerender` work item per page: my run
created **41** (`items_created: 41`), status `COMPLETED` within seconds. **A COMPLETED
orchestration here means "41 rows filed", not "41 pages rendered"** — I nearly reported the fix as
shipped on that status.

Checked at the artefact instead: **1** component re-rendered, and the 12 components carrying an
ink repoint are all `article-body` from **migration 415, rendered 08-15** — i.e. *none of mine*.

The queue then sat unchanged for 10+ minutes at `claimed=1, triaged=40`. It is **not this site**:

| site | signature, 2026-08-17 ~18:0x |
|---|---|
| ai-agent-orchestration.com | claimed 1, triaged 40 |
| fundamentallyai.com | claimed 1, triaged 43 |
| finetuning.uk | claimed 1, triaged 35 |
| idea.uk | claimed 1, triaged 32 |
| webdesign.co.uk | claimed 1, **failed 30** |
| loancalculator.co.uk | claimed 1, **failed 7** |

Every site wedged at **exactly one claimed item**, all claimed by `build-dispatch-loop`. That is
`bugs_open/029` — *hung spawns saturate dispatch group and halt builds fleetwide* — which is
**OPEN and owned**. Not re-diagnosed and not forked; recorded here as the blocker.

> **CORRECTED 2026-08-18, by the `bugs_open/029` lane** (`CONTRIB_2026-08-18_from_the_029_lane_what_wedged_your_queue.md`,
> commit `619c02474`). Two things above are wrong, and both came from quoting the bug's TITLE:
>
> 1. **The dispatch concurrency group is NOT saturated**, and has not been since that file's own
>    2026-07-21 correction. The title still says it; I repeated the title. The real mechanism is a
>    **per-site mutex** in `build-pipeline-trigger`'s `find_dispatchable_site`:
>    `NOT EXISTS (… active.site_id = wi.site_id AND active.status='claimed')`. **That is why it is
>    "exactly one" on every site** — one orphaned `claimed` row removes that whole site from
>    dispatch, and a second cannot form while the first is held. My table was that mutex
>    photographed, which is the one thing I did get right.
> 2. **"Blocked" overstates it: the cost is ~40 minutes per site per incident, not indefinite.**
>    `claimed-item-timeout` releases the claim at 40 min. My own 08-18 entry below — the queue
>    drained overnight with nobody intervening — is the evidence for the correction, and I had it
>    in hand while still writing "blocked".
>
> What put the claim there (their measurement, not mine): `build-pipeline-trigger` declares
> `timeout_seconds: 900`, but that is honoured on **attempt 0 only** — every retry is silently
> recomputed to 5 minutes, so the caller burns three retries in ~25 min instead of ~60 and its
> final replay lands on a loop that is still running. 11 of 12 wedged loops froze 11–22s after
> that send, in `EXECUTING_STEP` with an empty awaited set — invisible to `TimeoutMonitor` and the
> retry driver, which both key on awaited requests. ⚠ **Why the replay wedges the loop is
> [UNVERIFIED]** by that lane; the optimistic-lock race is their leading candidate and must not be
> repeated as fact.

**So: the contrast fix is correct, applied and committed at the source, and invisible to a visitor
until that queue moves.** Reported to the owner as blocked rather than done, because the
difference is the whole point.

The documented bypass ([[single-page-deploy-bypasses-stalled-queue]]) needs care and was **not**
fired blind: with no `reason` stamped, `page-rerender` takes the `render_page` branch, which
assembles from **stored `rendered_html`** and would therefore ship the OLD css — a bypass that
completes green and propagates nothing. Only the `rerender_sections` branch (a `reason` of
`image_landed` / `section_data_resolved` / `cta_links_stale`) regenerates from `content_data`.
⚠ This site has **2 locked components**, and firing `section_data_resolved` at a locked,
positionally-named section **duplicates** it (LANDMINES; `bugs_open/189` records the reversal SQL).
Count `page_components` for the page before and after if that route is taken.

---

## 2026-08-18 — the queue drained, the fix landed, and it introduced one failure of its own

Owner is working `bugs_open/029` in another thread. Blocker noted, not forked.

### The queue drained overnight and 456 propagated

69 of the 70 queued `page_rerender` rows completed (latest 08-18 08:36Z); 81 of 116 components
re-rendered. **Measured at the artefact, same 4 pages as the 08-17 baseline:**

| page | before | after |
|---|---|---|
| index | 17 | **10** |
| about | 19 | **15** |
| pricing | 8 | **8** (expected — cannot re-render, 5/5 NULL `content_data`) |
| services | 0 | 0 |
| **TOTAL firm** | **44** | **33** |

Family split confirms the fix did what it claimed and only what it claimed: dark-on-dark
**20 → 8** (and the 8 survivors are all on `pricing`, the unrenderable page); light-on-light
**24 → 24**, untouched, because family B is a hardcoded white ground and 456 never addressed it.

### ⚠ 456 INTRODUCED A FAILURE, and the net number hid it

The after-audit contained a colour pair that had never appeared before:
`rgb(118,142,178) on rgb(240,165,0)` — `#768eb2` (the ink) on `#F0A500` (the amber accent),
**1.61:1**, at `a.stats-cta` "View full report" on `/index.html`. Before 456 that label was
`#0D1117` on `#F0A500` — near-black on amber, perfectly legible. **I made it worse.**

Root cause, and it is a flaw in how I wrote 456 rather than in the ink mechanism:
**I repointed every foreground `--color-primary` regardless of the ground it sits on.**
`--color-primary-ink` is derived to clear the floor against the PAGE grounds — background,
surface, and the composited overlay — and carries **no guarantee against a fill**. `.stats-cta`
is `background: var(--color-accent, #7dd3fc)`, so the repoint swapped a well-contrasted colour
for one that was never measured against amber.

**Caught by re-auditing the same four pages, NOT by the net figure.** 44 → 33 reads as a clean
win and contains a regression. A before/after total cannot see a swap; only the per-pair
breakdown could, and only because the new pair had no counterpart in the baseline.

**The renderer already had the right token and the component library was already expected to use
it** — `palette_specialised_slots.go:105-107`: *"`--color-primary-text` on a primary-filled
button, `--color-accent-text` on an accent-filled one"*, with `--color-accent-text` derived from
the palette text colour against `palette["accent"]` **as its ground** (line 729). Verified live
before writing the fix: it computes to `#294155` here.

Migration `457` applies it. Census of how many of 456's 36 repointed declarations sat on a fill:
**exactly one** (a second apparent hit, `.tag.highlight`, was a false positive from my own crude
block parser — the repointed declaration there is `border-color`, and its `color` is
`var(--color-background)`, untouched).

> **Transferable rule, and the reason 457 exists as its own migration rather than a quiet edit:
> a foreground repoint is safe only when the declaration's own rule block sets no `background`.
> Census the BLOCK, not the declaration.** Whoever takes the remaining 144 templates inherits this.

### ⚠ RENDERED ≠ DEPLOYED — and the DB query is the one that looks like success

After applying 457 I checked `page_components.rendered_html`: **2 rows carry the accent-text
fix.** Then I checked the live page: **still 1.61:1.** Both true at the same time. The component
is correct in the database and wrong on the internet, because nothing has re-assembled and
published the page since.

Had I stopped at the SQL count I would have reported 457 as done, and it would have read as
verified — a count of fixed components is exactly the shape of evidence that looks like proof of
a live fix and is not.

Queue at 12:12–12:17Z: `triaged=41, claimed=0`, static. **Different signature from yesterday's**
fleet-wide wedge (`claimed=1` on every site = `bugs_open/029`) — zero claims rather than one
stuck claim, plausibly a consequence of the owner's 029 work in flight. Not diagnosed, not
forked, and NOT re-fired: the rows are queued, and a missing completion is not a lost message.

> **CORRECTED 2026-08-18 ~12:25Z — the speculation in that paragraph is REFUTED, twice, and the
> real answer is dull: ordinary congestion.**
>
> 1. **It was not the 029 lane's work.** That lane states it has run **no mutations** against
>    `site_work_items` or `orchestration_states` — reads only, plus one `090` intake anchored on
>    `system.internal`. So attributing my empty claim set to their activity was a guess, and a
>    wrong one. It was the cheapest kind of wrong: an unfalsifiable "plausibly X" about a
>    neighbouring thread's actions, written instead of a query.
> 2. **It was not a wedge at all.** Measured:
>    - `build-pipeline-trigger` is `enabled`, `interval_seconds=60`, and **completed 8 seconds
>      before I looked** (`last_completed_at` 12:23:36Z). The trigger is healthy.
>    - The build pipeline is **draining fleet-wide at 2–3 items/minute**, every minute, for the
>      last 45 (`GROUP BY date_trunc('minute', updated_at)` on `status='complete'`).
>    - **My site holds ZERO `claimed` rows of ANY item type**, so the per-site mutex is not
>      excluding it — it simply has not been selected yet.
>    - Fleet backlog: **15+ sites** with `triaged`/`pipeline='build'` items, all created within the
>      preceding hour, and mine (12:07–12:10) is among the **newest**, i.e. near the back.
>    - Claims are churning normally: 3 sites held one at 12:22, **1 site at 12:24 — a different
>      site**, 1 minute old.
>
> **My query could not have seen the answer**, which is the transferable part: I filtered on
> `item_type='page_rerender'` and concluded "zero claims". The mutex is per **SITE, across all
> item types**, so a claim held by any other type would have been invisible to me — as it happens
> there was none, but I could not have known that from the query I ran. Same shape as
> [[your-action-moves-you-to-the-back-of-the-selector]] and the standing rule that a measurement
> answers the question you ENCODED: a filtered count cannot rule out a cause the filter excludes.
>
> **So "delayed behind a fleet backlog" — not blocked, not 029, and nothing to act on.**

---

## 2026-08-18 (late) — family B diagnosed properly, and my earlier account of it was wrong twice

Went looking for the carousel seam, and the same investigation settled family B. Both results below.

### ⚠ CORRECTION — "seven components hardcode a light ground" was WRONG, twice over

My 08-17 entry listed **seven** components as hardcoding a white ground, from this query:

```sql
... WHERE pc.rendered_html ~* 'background[^;]*(#fff|255, *255, *255)'
```

**That query cannot distinguish a bare literal from a `var()` FALLBACK, and five of the seven were
fallbacks** — `background: var(--color-background, #fff)`, present in the source and never applied
because the token is set. The very landmine I cited in the same entry
([[a-css-fallback-is-present-and-inoperative]]), and I walked into it while writing about it.

Then I over-corrected. Reading `differentiators-section`'s template and finding proper `var()`
usage, I concluded the templates were clean and the white must come from elsewhere — chasing it
through the site stylesheet (`--color-background: #080B10`, no `team-member` rule at all) and a
browser probe. **That was wrong too**, and the reason is worth recording: I grepped

```
grep -oE '\.(team-section|team-member)[^{]*\{[^}]*\}'
```

which returned `.team-section { padding: 3rem 1.5rem; }` — a **media-query duplicate** of the
selector. The real rule, with the background, is elsewhere in the same file. **A grep that returns
"the rule" when a selector appears more than once answers about the first occurrence and reads as
though it answered about all of them.** Resolved by asking the database instead of the text:
`(html_template ~ 'background: *#f8f9fa') = true` on the row reached **through `component_id`**,
not by name.

### What family B actually is — narrower, and worse

**Two shared templates, not seven: `departments-grid` and `leadership-team`.** Their complete
colour surface:

```
background: #f8f9fa;     <- section ground
color:      #555;
background: #fff;        <- card
background: #e0e0e0;     <- icon circle
color:      #0f3460;     <- card heading
color:      #555;
```

**Not one themed value. These two components have NO theme support whatsoever** — they are pure
light-theme CSS in a library whose sibling section component (`differentiators-section`) already
does it correctly with `var(--color-background, #fff)` / `var(--color-surface, #f8f9fa)`. On this
site, the only DARK site using them, they were never going to work.

⚠ **This is why the fix is NOT "tokenise the backgrounds".** Doing only that would put `#555` and
`#0f3460` — mid-grey and dark navy — onto a `#0D1117` card, i.e. **a fresh set of invisible text,
which is migration 456's mistake repeated exactly.** The 457 landmine ("census the BLOCK, not the
declaration") caught it before I wrote anything.

**The whole block must move together, and every token it needs already exists** (read from the
served `:root`, 35 tokens):

| current | token | value here |
|---|---|---|
| `background: #f8f9fa` | `var(--color-background, #f8f9fa)` | `#080B10` |
| `background: #fff` | `var(--color-surface, #fff)` | `#0D1117` |
| `background: #e0e0e0` | `var(--color-border, #e0e0e0)` | `#21262D` |
| `color: #0f3460` | `var(--color-text, #0f3460)` | `#E6EDF3` |
| `color: #555` | `var(--color-text-muted, #555)` | `#8B949E` |

**Blast radius, checked before proposing (the 456 lesson): 3 sites, and the two light ones barely
move.** `departments-grid` on ai-agent-orchestration + finetuning.uk; `leadership-team` on those
plus leopardessconsulting.co.uk.

| site | scheme | `--color-surface` | effect on the card |
|---|---|---|---|
| ai-agent-orchestration.com | DARK `#080B10` | `#0D1117` | **fixed** |
| finetuning.uk | light `#F5F3EF` | `#FFFFFF` | **identical to today's `#fff`** |
| leopardessconsulting.co.uk | light `#FAF8F4` | `#FFFFFF` | **identical to today's `#fff`** |

So the light sites keep their appearance (their surface token *is* white) and the dark site is
repaired. Two-level fallback throughout, so it is inert anywhere the tokens are absent.

**This retires the "owner design decision" I recorded on 08-17 as an either/or.** I framed it as
*strip the white so the dark theme shows through* vs *keep light cards and darken the text inside*.
Neither is right: the components should consume the site's tokens, exactly as their sibling in the
same library already does, which gives each site its own answer instead of picking one for all
three. The owner still owns whether to spend it — but it is no longer a choice between two designs.

### Carousels — the seam, and a conflict in this site's own spec

`plan_sections_action.go:1101` is the planner→creator seam: **`design_intent.style_direction` is
the string passed to `needs_new_component` items "so the component-creator knows the visual
style"**. On this site it is the bare slug **`"professional-dark"`** — a very low-bandwidth channel.

⚠ **`layout_preference` is NOT read by that seam**, despite being the obvious place and despite
carrying rich, relevant prose here (*"Card-based layouts for services, case studies, and team
members"*). Anything written there today reaches nothing. **[MEASURED]** by reading the Go: the
extraction takes `style_direction` only.

⚠ **This site's `design_intent.avoid` already contains "Testimonial carousels with headshots of
fake people".** Narrow — it bans one *kind* of carousel, not the pattern — but it is the site's own
standing instruction and any carousel hint must be written so the two do not read as contradictory.
Flagging rather than deciding: the owner asked for carousels and the spec discourages one variety
of them.

### ⚠ A separate, unrelated defect found in passing — the spec carries TWO contradictory palettes

`design_intent` holds `palette.reference_values` (**dark**: background `#080B10`, surface `#0D1117`)
**and** a `color_scheme` block (**light**: background `#ffffff`, surface `#f8f9fa`, text `#333333`),
plus a `colour_mood` prose paragraph describing a **third** scheme (electric blue `#0EA5E9` on
slate). `color_scheme` is **live** — `render_css_from_spec_action.go:480` does
`palette = extract("color_scheme")`.

**Fleet-wide this site is the only contradictory one:** of the sites carrying both keys, only
ai-agent-orchestration has a light `color_scheme` against a dark palette (leopardess has
`#FAF8F4`/`#FAF8F4`, consistent).

**[UNVERIFIED] what this currently causes.** The served `:root` is dark and correct, so the palette
is evidently winning today — but the two blocks disagree and one of them feeds CSS generation, so a
future run could serve the light one. Recorded, not acted on, and NOT offered as family B's cause:
family B is the untokenised components above, which is measured.

---

## 2026-08-18 (evening session) — family B applied as migration **469**, not 459

### First: two corrections to what the handoff told me

**1. The number 459 was already taken.** The handoff of ~12:15Z says "ready to write as migration
`459`". By the time I looked, `459_zip_deliverer_agent_HOLD.sql` existed from another lane, as did
460–468 and (while I was working) 470. The runner's rule is *next number = highest in the directory
+ 1*, so this landed as **469**. ⚠ **Numbers in `sql_for_agents/` also COLLIDE** — 457 and 458 each
name two unrelated migrations from two lanes (`457_stats_cta_…` vs `457_tool_auditor_…`;
`458_aiao_imagery_…` vs `458_detected_item_promoter_…`). So a bare migration number is ambiguous in
this directory exactly as a bare bug number is in `/bugs_*/`. **Cite the filename.**

**2. "Both light sites are unchanged" was too strong, and I had written it myself.** The table
above compares only `--color-surface`. The migration moves five tokens, and on the light sites three
of them are NOT identical to today's literal:

| declaration | today | finetuning.uk | leopardess |
|---|---|---|---|
| `.team-section` ground | `#f8f9fa` | `#F5F3EF` | `#FAF8F4` |
| `.team-member` ground | `#fff` | `#FFFFFF` **identical** | `#FFFFFF` **identical** |
| icon well | `#e0e0e0` | `#D4CFC6` | `#E4DFD5` |
| `.member-title` | `#0f3460` | `#1A1A2E` | `#1A1A1A` |
| `.section-intro` / `.member-bio` | `#555` | `#6B6860` | `#5A5751` |

Only the card ground is genuinely unchanged. The others shift to each site's own tokens — which is
the *intent*, but it is a change, and the defensible claim is **"no element loses its contrast floor
on any site"**, not "nothing moves". [MEASURED] by simulating every element of the block on all
three sites; every light-site row passes before AND after, lowest after-value 5.02:1 against a 4.5
floor (finetuning `.section-intro`). The disconfirming result was available: any row going
PASS→FAIL would have stopped the migration.

### What I verified before touching anything

- **Live contrast, re-measured 19:14Z: 32 firm failures** (`render_audit.py`, overImage excluded —
  raw 35, of which 3 over-image). Composition matches §3 of the handoff exactly: 20 `#E6EDF3` on
  `#FFFFFF`, 4 on `#F8F9FA` (= family B, 24), 7 `#0D1117` on `#0D1117` + 1 on `#080B10` (= family A,
  8, all on `pricing`).
- **`a.stats-cta` produced ZERO findings.** So 457 IS deployed and live. The handoff carries two
  `## 7` headings and the second one's item 1 ("Finish 457 — applied and rendered but NOT DEPLOYED,
  still 1.61:1") is **stale**; §1 of the same file is right. Resolved, nothing to do.
- **All five tokens are SET on all three sites** — read by `getComputedStyle` against the live
  pages, never from a stylesheet (R2). No branch falls back in practice.
- **Fleet-wide, exactly TWO components define `.team-*`/`.member-*`.** This is what licenses the
  "both must move together" reasoning: they share class names, both are placed on `about.html`, and
  their `<style>` blocks land on one page, so migrating one alone would restyle the other through
  the cascade. Asserted as a guard, not just checked.

### Family B's real breakdown — and `departments-grid` is on INDEX too

The handoff describes these as "the two about-page components". `departments-grid` is placed on
**index as well**, which is where 9 of the 24 live:

```
index / departments-grid   9   (1 H2 + 8 department H3s)
about / departments-grid   9   (1 H2 + 8 department H3s)
about / leadership-team    6   (1 H2 + 1 stray P + 4 member H3s)
```

**The mechanism is not a bad foreground.** `.team-section h2` and `.team-member h3` set *no colour
at all* — they inherit `--color-text` (`#E6EDF3`). The defect is a **themed foreground inherited
onto an unthemed ground**. That is why tokenising only the backgrounds would have been wrong in the
other direction: it would have left `color: #555` and `color: #0f3460` on a `#0D1117` card.

### A guard of mine that was wrong, caught before it ran

I first wrote the "never introduce a bare `var()`" invariant as a **fleet-wide** assertion, copying
456/457's shape. 456/457 assert it only for the two *ink* tokens. Measured before applying:
**145 `content_components` fleet-wide already carry a bare `var(--color-*)`** with no fallback. The
guard would have aborted the migration on pre-existing state that has nothing to do with this
change — and the failure would have read as "my edit introduced a bare var". Narrowed to the two
rows I touch; kept the bare-**ink** invariant fleet-wide, which genuinely is 0.
**The lesson: an invariant copied from a neighbouring migration inherits its SCOPE, and that scope
was chosen for a different token set.** Check the count before you assert it.

### A separate defect found in passing — content is DOUBLE-WRAPPED in `<p>`

`about` / `leadership-team` renders:

```html
<p class="section-intro"><p>AI Agent Orchestration is built around a single principle: ...</p></p>
```

The content already carries its own `<p>` and the template wraps it in another. The HTML parser
auto-closes the outer tag, so the **inner bare `<p>` is a SIBLING with no class** — it inherits
`--color-text` instead of `.section-intro`'s muted colour. That stray paragraph is one of the 24
failures, and 469 does fix its contrast (it sits on `.team-section`, whose ground is now themed).
**The double-wrap itself is untouched and is a real content/render defect** — `.section-intro`'s
styling reaches nothing on this site. NOT filed as a bug yet; it needs a fleet-wide census first
(is this one component's content, or every `{{if .x}}<p>{{.x}}</p>{{end}}` against LLM-authored
prose that already contains block markup?).

### Applying it

- **Rehearsed first**: the whole file with its final `COMMIT` replaced by `ROLLBACK`. Guards passed,
  `INSERT 0 2` / `UPDATE 2`, then rolled back; confirmed afterwards that `migration_backups` held 0
  rows and neither template was tokenised. A dry-run `SELECT` proves the anchors match; only a
  rehearsal proves the **guards** pass.
- **Applied alone with `psql -f`, NOT `run-migrations.sh --apply`.** ⚠ `--apply` takes EVERY pending
  file: at the time that was **17 files from at least six other lanes** (456–458, 460–462, 467, 468,
  470, plus older ones). Recorded afterwards with `--record-only`.
- ⚠ **"Pending" in that runner does not mean "unapplied".** 460 was listed pending, yet
  `template_changed` is already live in `page-rerender.check_rerender_mode` — it was applied by hand
  and never recorded. I checked the live agent row rather than trusting the listing, which is what
  told me the propagation path below actually exists.
- **Byte-exact rollback**: 469 backs both rows up to `migration_backups` first, and the ROLLBACK
  file restores from those rows rather than reverse-regexing. It cannot drift from what was replaced.

### Propagation — page-scoped, and why not site-wide

Templates changed; placements keep the OLD html until re-rendered. Filed **two** `page_rerender`
items (index, about) with `spec.reason='template_changed'`, copying the live
`component-template-fixer.create_rerender` shape.

- ⚠ **`reason` is load-bearing.** `page-rerender` routes to `rerender_sections` ONLY for
  `image_landed | section_data_resolved | cta_links_stale | template_changed`; anything else
  **assembles stored HTML** and ships the old CSS with a green status.
- ⚠ **Page-scoped ON PURPOSE, not site-wide.** A site-wide `needs_rerender` fans out to every page
  including `privacy` and `terms`, whose `generic-text-block` components are **permanently locked**
  (`182_legal_pages`, 2026-07-21) — and firing a rerender at a locked positionally-named section
  *duplicates* it rather than protecting it (`bugs_open/189`).
- Both target pages are `rebuild_policy='generic'` and all three placements have non-null
  `content_data`, so neither the content-gating refusal nor the `pricing`-style "nothing to rebuild
  from" applies.

### The other two sites are deliberately NOT re-rendered

`finetuning.uk` (3 placements) and `leopardessconsulting.co.uk` (1) now carry a template their
rendered HTML does not reflect. **This is safe and non-urgent**: their live pages keep today's
literals, so nothing changes for them until they next re-render for any reason, at which point they
get the tokenised version — which the simulation above says passes on both. Not re-rendered because
they belong to other lanes and I have measured their contrast only by simulation, not at the
artefact. Whoever owns them should re-render when convenient.

### The double-wrapped `<p>` — censused, and it is LOCAL, not a fleet pattern

Above I said this "needs a fleet-wide census first — is this one component's content, or every
`{{if .x}}<p>{{.x}}</p>{{end}}` against LLM-authored prose that already contains block markup?"
**Answered: it is two placements, both on this site's `about` page, and nothing else fleet-wide.**

| what it carries | placements | sites |
|---|---|---|
| `content-block-about` / `body_text` | 1 | 1 |
| `leadership-team` / `section_intro` | 1 | 1 |

**How the number moved, because the intermediate answers were both wrong and each looked
authoritative:**

1. `rendered_html ~ '<p[^>]*>\s*<p[ >]'` → **2 placements, 1 page, 1 site.** I distrusted this,
   correctly in principle: `rendered_html` is a **lagging** indicator, stale wherever a placement
   has not re-rendered since its content changed. A clean result there can mean "not yet visible".
2. So I censused `content_data` instead, which is current → **359 placements, 26 sites, 8
   components.** That number is real but answers a different question: block markup in
   `content_data` is only a defect where the TEMPLATE wraps that field in a `<p>`. Five of the eight
   components render the field bare into a block context, which is correct and intended.
3. Narrowing to "component wraps *a* field in `<p>`" → **4 components, 26 placements.** Still wrong,
   and this is the instructive one: `about-content` (23 of those 26, across 16 sites) has
   `<p>{{.description}}</p>` in its template and block markup in its `content` key — **two different
   fields.** It renders `{{.content}}` bare. The regex proved *some* field is wrapped, never *the
   one carrying the markup*, and at that point I had a fleet-wide finding across 16 sites that was
   an artefact of the join.
4. Matching the p-wrapped field NAMES against the fields that actually carry markup, and checking
   the `{{range}}`-nested ones separately (`leadership-team.members[].bio`,
   `about-content.highlights[].description` — both **0**, and a top-level `jsonb_each_text` cannot
   see either) → **2.** Which is where step 1 started.

**So the lagging instrument and the precise one agree, and the two readings in between did not.**
Worth writing down because the wrong ones were the ones that looked like a discovery: "359
placements across 26 sites" is exactly the shape of a finding you carry into a bug file.

**Not filed as a bug.** Two placements on one page, no automated consumer, and the visible symptom
(the stray paragraph's contrast) is fixed by 469. What is NOT fixed is that `.section-intro`'s
styling reaches nothing on this site — the inner bare `<p>` is a sibling, so the muted colour,
centring and max-width are all inert there. Cosmetic today; it belongs with the `pricing` rebuild
work, where the copy is regenerated anyway.

### 469 VERIFIED AT THE ARTEFACT — 32 → 8 firm failures, index and about at ZERO

Both `page_rerender` items completed (about 18:34:36Z, index 18:35:12Z — ~5 minutes behind ~50
queued items, no wedge, no `claimed` row held on this site at any point). Verified in the order
that matters, because the first of these is the one that looks like success:

**(1) the components** — all three placements carry `var(--color-surface, #fff)` and have **zero**
bare hex colour declarations left in `rendered_html`.

**(2) the live pages** — `render_audit.py`, overImage excluded:

| page | before (19:14Z) | after (19:36Z) |
|---|---|---|
| index.html | 9 | **0** |
| about.html | 15 | **0** |
| pricing.html | 8 | 8 (expected — unrenderable, §3c) |
| services.html | 0 | 0 |
| **TOTAL FIRM** | **32** | **8** |

**Regression check by COLOUR PAIR — the check 456 lacked: NONE.** Every `(fg,bg)` in the after-set
was already in the before-set. The after-set is now only the two family-A pairs on `pricing`
(`#0D1117` on `#0D1117` ×7, `#0D1117` on `#080B10` ×1). Both family-B pairs are gone entirely.

**Control on the two light sites: their `rendered_html` is UNTOUCHED** — all four placements still
carry bare hex, no token, `updated_at` still 2026-08-17. So their appearance cannot have changed,
which is what "not re-rendered" should mean and is worth asserting rather than assuming. They pick
up the tokens whenever they next re-render, and the simulation says that is safe.

⚠ **This lane's remaining contrast work is now ONLY `pricing`, and no re-render can touch it.**
8 failures, 5/5 components with NULL `content_data`, last rendered 2026-04-13. It closes via the
owner-approved framework rebuild (§3c) or not at all.

### ⚠ The `pricing` rebuild is NOT "approved and unblocked" — it was DISPATCHED and REFUSED

The handoff's §3c says `pricing` is *"owner approved 2026-08-17, **not yet dispatched**"* and
*"✅ `pricing` is `rebuild_policy='generic'`, **so the rebuild will not be refused**"*. Both halves
are wrong, and the evidence predates the handoff by ~16 hours.

**A page-scoped rebuild was already filed and it FAILED**, item
`889a0687-cc0a-4f5e-8693-9ee6ca98751a`, filed by `page-rerender` at **2026-08-17 20:28:00Z**,
handler `page-build-handler`, `spec.reason='content_data_backfill'`, `attempt_count=1/3`:

```
step save_sections failed: failed to execute action save_page_sections:
save_page_sections: SECTION SHRINK REFUSED for page "pricing" —
call-to-action 483→213 chars of VISIBLE text, stylesheet and script content excluded
(44% kept, floor 50%). A same-named prose slot may not lose more than 50% of the text a
READER sees in one save; if this shrink is intended, set section_shrink_floor in the step
config. Nothing was written (bugs_open/178, axis corrected by bugs_open/293).
```

**The `generic` check was true but answered the wrong question.** `rebuild_policy` gates
*overwrite permission*; this refusal came from the **shrink floor**, a different guard entirely.
Confirming a page is not `owned` therefore predicts nothing about whether its rebuild will be
refused — and reading "✅ will not be refused" is how a session skips looking for the refusal that
already happened. **A green answer to one gate is not a green answer to the gate that fired.**

**The page's state is otherwise exactly as documented**: 5/5 components `content_data IS NULL`,
`rendered_html` last written 2026-04-09/13, `rebuild_policy='generic'`, `status='active'`. So it
still cannot re-render, and the 8 remaining firm failures still close only via a rebuild.

**Why this is a decision, not a next step.** The guard is not misfiring: the regenerated
call-to-action really is 56% shorter than the one on the page, and the floor exists for the failure
class where an agent rewrites a whole artefact and persists a fragment while reporting success
(`bugs_open/012`, and CLAUDE.md's own `output_tokens == max_tokens` rule). Three routes, and they
are not equivalent:

1. **Set `section_shrink_floor`** in the step config and re-file the page-scoped rebuild. Cheapest,
   page-scoped — but it is *turning down a safety control* to let a known-shorter rewrite land.
   Whether that is right depends on whether the shorter CTA is acceptable copy, which is an
   owner-facing judgement about the live page, not a technical one.
2. **Find out why the regeneration is so much shorter first.** The floor may be reporting a real
   content loss rather than a stylistic one. Nothing has been written, so there is no damage yet
   and no time pressure.
3. **`082_submit_domain_unified.sh`** — the route §3c names. ⚠ **This is WHOLE-SITE, not
   page-scoped**: `needs_domain_research → strategy → briefing → site plan → design cascade →
   needs_content_page × N → rerender`. On a live 40-page site it regenerates everything, including
   the `index` and `about` copy whose contrast was just fixed and verified, and re-runs the design
   cascade. It is the correct tool for adopting or building a domain; it is a very large hammer for
   one broken page, and the owner's approval was for *"the pricing rebuild"*.

**NOT ACTED ON.** Recorded and put to the owner. Nothing is queued, nothing is at risk, and the
page has been in this state since April.

### `pricing` shrink investigation — STARTED, first finding: the text the floor is PROTECTING is itself defective

Owner chose "investigate the shrink first" (2026-08-18 evening). First step only — the **existing**
call-to-action's visible text, the 483 chars the floor refused to let drop to 213:

> Ready to Talk Architecture, Not Estimates? If you're evaluating AI agent infrastructure for
> production, the real costs are in failure modes, re-architecture, and provider lock-in — not the
> initial build. Before we talk scope, use our **`[LLM Provider Cost Comparison Calculator](/tools/tool-llm-cost-calculator.html)`**
> to get a concrete number on token costs across providers and whether self-hosting makes sense at
> your scale. Then let's have a direct conversation about what your system actually needs to run
> reliably. Discuss Your Requirements Explore Our Services

⚠ **That is LITERAL MARKDOWN being served as visible text on the live page** — the link syntax is
not rendered, it is printed. So a meaningful slice of the 483 chars the guard is defending is a
defect, not copy: the URL and the bracket/paren punctuation are only "text" because the markdown
was never converted. This site already carries open `literal_markdown` work items (1 failed, 2
needs_human_review), so it is a known family here.

**Why this matters for the decision, and why it is NOT yet an answer:** it means the 44%-kept figure
is measured against an inflated baseline, so the regenerated CTA may be losing far less real copy
than 56%. It does **not** yet establish that the replacement is acceptable — that needs the
regenerated text itself, which was never written (the save was refused, correctly).

**NEXT STEP, not taken:** recover what the rebuild actually generated. It is not in
`page_components` (nothing was written). Look in the orchestration for item
`889a0687-cc0a-4f5e-8693-9ee6ca98751a` — `orchestration_states.collected_data` for the
`page-build-handler` run of 2026-08-17 20:28Z — and compare its `call-to-action` section against
the text above **on the visible-text axis**, not tag-stripped (LANDMINES: the two axes agree on
zero pairs in eight days). ⚠ Chassis log retention here is ~4 minutes, so a log grep cannot recover
this and its control will also be zero.

### Step 2 — the original run's output is UNRECOVERABLE, and the search for it found my own verifier

**The generated sections from the 2026-08-17 20:28Z rebuild are gone.** `orchestration_states` is
pruned by the `database-cleanup` task, and `COMPLETED` rows now reach back only to **2026-08-18
14:32** — the run predates that by ~18 hours. Nothing else holds it: `page_components` was never
written (the save refused), and chassis log retention here is ~4 minutes.

⚠ **The obvious search for it returns a CONFIDENT FALSE POSITIVE, and it is one I created.**

```sql
SELECT orchestration_id, status, created_at FROM orchestration_states
WHERE collected_data::text LIKE '%889a0687-cc0a-4f5e-8693-9ee6ca98751a%';
--> 1 row: 4700bcb3…, COMPLETED, 2026-08-18 18:41:52
```

That row is **my own `landmine-verifier` dispatch** (correlation
`f30975fb-d7da-4764-a79c-b64de97143a5`), which matched because the LANDMINES entry I had written
minutes earlier **quotes that item id**, and the verifier collects the entry text. Read quickly it
looks exactly like "found it, and it COMPLETED" — the opposite of the truth on both counts.
**A `collected_data` search finds runs that MENTION an id, not runs that ARE it**, and after you
document something, the estate's own machinery starts quoting you back. Same family as
[[your-action-moves-you-to-the-back-of-the-selector]] and
[[prompt-text-poisons-its-own-detector]]. The tell was the timestamp: it postdated the failure by a
day and landed inside the minute I ran the dispatch script.

A second near-miss in the same search: a `FAILED` row from 2026-08-19 09:28 matching `%pricing%`
turned out to be **`webdesign.co.uk`**, matching because its *site plan* names a pricing page.

### Step 3 — re-firing the rebuild is SAFE here, and that is a property of THIS workflow only

To see the copy the floor refuses, the only route left is to regenerate it. Checked before firing,
because the estate has a recorded trap in exactly this shape (LANDMINES: the composition loops
`assemble_page → deploy_page(git_commit) → save_sections`, where **freshly LLM-written HTML is
committed to the deploying repo one step BEFORE the DB refusal**, so a "refused" save can still
have shipped).

**`page-build-handler` is ordered the other way**, read from the live agent row:

```
validate_content → save_sections → update_status → spawn_rerender_agent → deploy_page → complete
```

`deploy_page` is strictly **after** `save_sections`, so a refusal fails the workflow before
anything deploys. Corroborated by the artefact: the 08-17 refusal left `pricing` serving its April
render, and all five components still read `updated_at` 2026-04-09/13.

⚠ **Do not generalise this.** Which side of the refusal `deploy_page` sits on is a property of the
handler, not of `save_page_sections`. `page-build-handler` is safe; `page-rebuild`,
`pageflow-builder` and `site-work-orchestrator` are the ones that deploy first.

Filed one evidence item (`created_by='aiao-shrink-investigation'`), expected to FAIL at
`save_sections`. **Baselined all five `pricing` components by `md5(rendered_html)` first**, so
"nothing was written" can be proven rather than assumed.

### The 17 parked `contrast_failure` items will NOT drain yet — and this site is now a clean test of `bugs_open/296`

The handoff's §3b says the parked items "drain by themselves" once the page is fixed and the site's
render audit runs. **Half true, and the missing half is the schedule.** [MEASURED 2026-08-19]

- The 17 rows are still `deferred`, `updated_at` **2026-08-11 12:31** — untouched for 8 days.
- The last render-audit-authored item on this site is **2026-08-10 22:00**. The instrument that
  files them, and the only one that can retract them, **has not run here since**.

So my fix cannot close them today, and their continued presence is **not** evidence that the fix
failed. `bugs_open/296` (owned, `bugs_open/113` lane) is the file for this and its §9 records that
the rotation swept 14 sites overnight on 08-18 and retracted **40** parked rows — correcting its own
earlier prediction that "approximately none" would close. **This site was not in that sweep**, which
its `updated_at` of 08-11 shows directly.

**Why that makes this site useful to them rather than just late.** It is now a natural experiment
with a prediction that can fail: `index` and `about` genuinely measure **0** firm failures at the
artefact, while `pricing` still measures **8**. So when the rotation next reaches
`ai-agent-orchestration.com`, the retraction should close the parked rows belonging to `index`/
`about` and correctly decline the `pricing` ones. If it closes none, the mechanism is not doing what
§9 says; if it closes all 17, it is retracting rows it should have kept.

⚠ **NOT contributed into `bugs_open/296` itself**: that file is currently dirty in the shared
working tree from another session, and a same-file edit would ride along with theirs on whoever
commits first (LANDMINES: a pathspec commit still takes a same-file passenger). Recorded here
instead; whoever picks that lane up can lift it.

### ✅ SHRINK INVESTIGATION ANSWERED — the floor was right, the remedy was to RE-RUN, and the real blocker is a stale number

Evidence run completed 2026-08-19 15:17Z (orchestration `1c438d16-d991-48d1-8619-9f045a9d2a3d`).

**FIRST, THE CONTROL: nothing was written.** All five `pricing` components' `md5(rendered_html)`
are byte-identical to the baseline taken before firing, and every `updated_at` still reads
2026-04-09/13. The ordering argument held in practice, not just on paper.

**1. The shrink floor is NOT the blocker, and lowering it would have been the WRONG call.**

| call-to-action | visible chars |
|---|---|
| currently live (April render) | **483** |
| 2026-08-17 generation — refused at 44% | **213** |
| 2026-08-19 generation — this run | **489** |

The regeneration is **non-deterministic**, and 08-17 caught a bad draw. Today's is *fractionally
longer than the page it would replace*, so the floor would not have fired at all. **Had we set
`section_shrink_floor` to admit the 213-char version, we would have permanently lowered a safety
control to accommodate a one-off, and shipped the worse of two drafts.** The floor did precisely
its job: it refused a bad generation, wrote nothing, and cost one re-run.

⚠ **The corollary is about method, not this page.** A guard that fires on a *sampled* output cannot
be assessed from one firing. The question "is this floor too strict?" is unanswerable from the
refusal alone — you have to re-draw. One observation of a stochastic step is not a measurement of
it, and the fix that looks obvious from a single failure (loosen the threshold) is the one that
removes the protection.

**And the new copy is better, not merely longer** — clean prose, no literal markdown:

> *"What does this actually cost against building it yourself? That is the real comparison, not a
> feature list. A Technical Discovery Call gets you a scoped estimate against your own agent count,
> message volume and compliance requirements, worked through by someone who has configured the
> Kafka consumer groups and Postgres locking for systems like it. No sales deck, just the numbers
> your board is asking for."*

Compare the live one, which serves `[LLM Provider Cost Comparison Calculator](/tools/…)` as visible
text. **So the shrink floor was defending the worse copy of the two** — which is the shape the
`copy_quality_two_stage` CONTRIB warned about, confirmed here from the other end.

**2. The ACTUAL blocker is one stale number, and the claims gate is right.**

The run died at `validate_content`, **not** `save_sections` — it never reached the shrink floor.
`0 blockers, 1 error`, recovered from `agent_error_log` (`CONTENT_VALIDATION_BLOCKER_DETAIL`, which
persists the structured issues precisely so this does not need pod logs):

```
type: unregistered_number   severity: error   category: claims
value: "170"
location: ", Kubernetes, Kafka, and Postgres, drawing on a registry of 170+ agents
           already built and running in production."
```

The site's `evidence_base` (7 facts) registers **196** — *"active agent definitions in the
registry"* / *"distinct agent types"* — and `17` backend services. **There is no 170.** So the
writer asserted a stale figure and the gate refused the page. The gate is correct; the copy is wrong.

⚠ **And the same stale number is ALREADY LIVE on this site.** `about` / `leadership-team` serves
*"The framework now coordinates 170+ agent types on Kubernetes and Kafka"*. That page passed
because it was rendered in April, long before this gate existed — so the gate is not inconsistent,
it is simply newer than the page. **A claims gate only ever guards writes; it cannot see what is
already served.** Anything asserting 170 on this site today predates the check.

**What this means for the owner's decision:** the choice put on 2026-08-18 (accept the shorter copy
vs investigate) is **retired** — there is no shorter copy to accept. The pricing rebuild needs one
of:
- **(a)** the `170` corrected to the registered `196` at source, so the writer stops asserting it —
  this also fixes the live `about` page's stale claim; or
- **(b)** `170+` registered as a fact if it is the figure the owner actually wants to publish,
  which given `196` is registered would be publishing a *lower* number than the truth.

**(a) looks right and is not mine to decide** — it changes a public claim about the business.
Nothing is queued; the evidence item is at `needs_human_review` and wrote nothing.

---

## 2026-08-22 — the claim fixed at source, carousels and images SHIPPED, and two plans corrected

Owner: *"Update the claim to 196 agents at source. Please go ahead with the remaining tasks…
carousels [and images]."* All three done and verified at the artefact. Two of the three needed the
plan changing first, and those corrections are the useful part of this entry.

### 1. The "196" instruction, executed in substance rather than literally

**Writing 196 would have reproduced the bug with a different number.** Three measurements, all
taken before anything was edited:

1. **The facts are live SQL and had already moved.** `aao-agent-definitions` /
   `aao-agent-types` are `SELECT count(*) FROM agent_definitions …`, re-run on every evidence
   refresh: **175/174** on 2026-07-26, **196** when I reported to the owner on 2026-08-19, **199**
   on 2026-08-22. Any literal committed to a spec is wrong within days.
2. **"170+" was never false.** Both facts carry `tolerance: "gte"` and
   `datahelpers.numberSupported` accepts any asserted value at or below the registered one
   (`claims.go:1007`). 170 ≤ 199, so the *number* passed.
3. **The rejection was the CONTEXT gate.** `numberSupported` skips a fact entirely unless one of
   its `context_terms` appears in the window (`claims.go:990-1001`). The sentence was *"…a
   registry of 170+ agents already built and running in production"*; the terms were
   `"agent definition"`, `"specialised ai agent"`, `"agents in the registry"`, `"ai agents"`.
   None match. **No fact was eligible to vouch for a number both facts would have accepted.**

**Root cause: the evidence base instructed the writer to produce a phrase its own facts could not
validate.** `writer_block` said, twice, Write `"170+ agents"` / `"170+ agent types"` — wordings the
supporting fact does not list. Obeying the instruction guaranteed refusal; the only way to pass was
to disobey it.

**Why it was stale at all, and this is the transferable bit:** `writer_block` is HAND-WRITTEN here
and nothing regenerates it. `refresh_evidence_base_action` rebuilds it from each fact's
`writer_line` with `{value}` interpolated — but **only where the site sets
`writer_block_managed: true`** (`:474`). This site never opted in, so the facts refreshed while the
prose the writer actually reads froze on 2026-07-27, still announcing "175 as of 2026-07-26" and
"14 live sites" against live values of 199 and 25.

⚠ **And opting in is NOT safe yet — deliberately not done.** `composeWriterBlock` (`:996`) builds
the block from `writer_line`s and `allowed_entities` and **nothing else**. Setting the flag today
would silently delete both NEVER-write bans, the entire NOT-TRACKED list, and the two "DO NOT state
a figure" cautions. The bans survive as `banned_claims` regexes so *enforcement* holds; what would
go is *prevention* — and the two cautions are enforced nowhere at all. Recorded in migration `557`
as the condition for ever setting that flag.

Shipped as **`557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate.sql`**:
`writer_block` carries no literal count (every figure delegated to the facts list, which is the only
thing that refreshes) and mandates a wording the checker recognises; five natural paraphrases added
to the two agent facts' `context_terms`. **Fact VALUES untouched — they belong to the refresh
action, and a migration that writes a count is the defect this file exists to remove.**

⚠ **The added terms are phrase-based, NOT the bare word "agents", and that choice was vindicated
the same day.** Adding `"agents"` would make every "N agents" sentence eligible for a `gte` fact of
199. Later that afternoon the index rebuild was correctly refused for *"Consumer Group
Misconfiguration Threatens a **40**-Agent Pipeline"* — a client-scale claim this site does not
measure. Had I broadened to `"agents"`, that false claim would have been waved through by the fact
about our own registry.

**Verified:** re-fired the pricing rebuild; the `170` error is **gone**.

### 2. `bugs_open/364` — a clock time reads as a business claim (found by the fix working)

With 170 fixed, pricing failed on something else entirely: `unregistered_number "2"` from
*"debug a failing agent chain at **2am**"*. `unitSuffixRe` (`claims.go:681`) excludes
px/rem/ms/sec/`min read`/kb/mb/ordinals/hyphenated durations but **not `am`/`pm`**, so the token
reaches the lexical gate, which matches on `agents?` in the same sentence. Third instance of one
shape (`bugs_closed/073` `Read time: 8–12 minutes`, `bugs_closed/102` page-type blindness): the
exclusions are an **allow-list**, so every unit it has never met costs one false refusal.

Fixed (`am\b|pm\b`), tested, filed, **council-submitted `39d04868-6ce3-472f-a976-49cd387a7860`**.
⚠ Go, so **inert until the next fleet roll** — do not read a later success as proof it shipped.

⚠ **MY FIRST TEST WAS VACUOUS AND PASSED.** It used the shared populated fixture, where a `gte`
fact already supported the value 2, so nothing was ever flagged and the test asserted nothing. It
**passed with the fix reverted**. Caught by mutating the regex out and re-running — the only reason
I know the test works. Rewritten against an empty `&EvidenceBase{}` (nothing supported ⇒ any number
reaching the scan is reported): all four cases now fail without the fix. **A test that cannot fail
is worse than no test, and only a mutation tells you which you have.**

⚠ Coverage gap, named rather than hidden: I committed before submitting, so that commit carries
neither trailer and will list as un-reviewed in the `098` report. Forward-only forbids an amend.

### 3. Carousels — the handoff's plan would not have produced a carousel

`HANDOFF_2026-08-18` §5 says the work is *"APPROVE + BIND, not design"*. The contracts do exist and
this implementation follows one. **But approving and binding would have put nothing on the site.**

**[MEASURED 2026-08-22] the experience register is a SPECIFICATION AND VERIFICATION system, not a
generator.** Only **3** Go files touch it: `write_experience_pattern_action.go` (records a
contract), `bind_site_experience_action.go` (records which page it applies to), and
`verify_site_experience_action.go`, whose own header says it *"run[s] a bound fork's criteria
against the deployed page"*. **Nothing renders from `site_experiences`.** Binding would have
produced a fork whose criteria then fail against a page with no carousel. And the trigger script
says the rest itself: *"nothing in this lane can write `status='approved'`. Applying a verdict is a
separate action that does not exist yet, deliberately."* Register state is unchanged from 08-18:
**11** patterns, **0** approved, `experience-approval-council` **0** runs ever, as of 2026-08-22.

So **`559_case_studies_grid_optional_scroll_snap_carousel.sql`** implements the
`arrow-and-swipe-card-carousel` contract directly: native scroll-snap track (works with **no JS**),
JS adds arrows only, reduced-motion honoured in CSS and re-checked at click time, init idempotent
against both a double include and a re-init on the same node. **Auto-advance deliberately NOT
implemented** — the contract makes it conditional, and it is the clause that drags in the
IntersectionObserver, hover/focus pause and re-derive-after-swipe rules. Nothing rotates, so none of
those failure modes exists.

⚠ **The `no-inert-control` invariant is met by OVERFLOW, not by counting cards**, and on this
component the difference is load-bearing: `case-studies-grid` ships a category filter that hides
cards with `display:none`, so a count taken at init says "show the arrows" and stays wrong after the
visitor filters to one card. Visibility is derived from `scrollWidth > clientWidth`, re-evaluated on
scroll, on resize, and via a `MutationObserver` on the cards' `style` attribute — the filter's own
mechanism.

**Opt-in, default OFF** (owner ruling 2026-08-02 on shared seams): the component sits on **4** pages
across **3** sites as of 2026-08-22. Verified at the artefact — both aiao pages carry the track,
controls and script; **finetuning.uk and leopardessconsulting.co.uk return ZERO carousel markers**.

⚠ **My first live probe said the arrows were broken and it was WRONG.** An inline end-of-body probe
executes before `DOMContentLoaded`, so it sampled the pre-init state (`controlsHidden: true`).
Re-measured after `load`: `hidden=false`, `display=flex`. **A probe that runs earlier than the code
it measures reports the code as absent.**

### 4. Images — two faults wearing one symptom, and a derived path that made the old URLs unfixable

⚠ **The handoff says "there is no `cardN_image_url` key at all". True of `index` ONLY.**
`enterprise-reference-deployment` HAS all five keys, pointing at files that never existed. A fix
aimed at the missing-key half would have left five 404s in place.

⚠ **And the old URLs could never have come good, whatever was generated** — they end `.png`, and
this purpose serves `.jpg`. The served path is **derived, not stored**: `DeployedWebPath` →
`DeployedAssetPath` → `AssetKeyFilename` (underscores → dashes) with the extension from
`ImagePurposes[purpose]`, which for `content_hero` is **jpg** (`url_helpers.go:363`). Predicted from
the code, then confirmed at the artefact: every `<key>.jpg` 200, every `<key>.png` 404.

Nine `needs_imagery` items → `image-build-handler` (the live path, **62** completions fleet-wide as
of 2026-08-22). **NOT** `image-url-404-handler` / `image-source-unsatisfiable-handler` — both look
like the obvious owner of these rows and both only TRIAGE. Prompts are the framework's own
`cardN_image_alt` prose plus the house-style clause from `design_intent` as `458` left it (owner
ruling 2026-08-06: the framework writes the content, not me). Imagery policy is satisfied **by
subject** — every prompt asks for an abstract architectural diagram, so no person is depicted and
458's one hard rule cannot be engaged.

**Result: 10/10 card images live at HTTP 200**, 0 broken images reported by the audit, contrast
unchanged at 8 with **no new colour pair** — so neither the carousel nor the images regressed
anything.

### 5. Three system behaviours worth carrying forward

- ⚠ **Generating N images for one site SERIALISES behind N page rebuilds.** Each landed asset makes
  `image-build-handler` file a `needs_page`, and a `needs_page` holds the **per-site mutex** — one
  `claimed` row of ANY type removes the whole site from dispatch (`029`). My imagery run stalled at
  5 of 9 for ~9 minutes behind a rerender the image pipeline had filed itself. It drains; do not
  poke it (a takeover re-stamps `last_activity` and resets the 4-hour reaper clock).
- ⚠ **Those `needs_page` rebuilds fail at `validate_content` and write nothing**, so the image
  pipeline's own propagation does not work here. Propagate with a **page-scoped `template_changed`
  `page_rerender`** (the MERGE path, RUNBOOK R8), which is what actually shipped the images.
- ✅ **The claims gate protected my work.** That failing rebuild would have regenerated `index`'s
  `content_data` and dropped both `carousel_enabled` and the ten image bindings — the exact
  durability caveat written into `559`. It refused on the "40-Agent Pipeline" claim and wrote
  nothing, so the bindings survived (verified: flag present, 5 URLs, `updated_at` still mine).
  **The risk is real and only did not fire because an unrelated gate happened to refuse.**

### 6. Housekeeping

Owner ruling of 2026-08-22 (counts carry the date they were counted) landed mid-session; applied to
this lane's files — `559`, `560`, `364` and my `LANDMINES` entry. ⚠ That edit removed 2 lines from
`LANDMINES.md`, which `pattern-check` flags as touching an append-only ledger. **Both lines were my
own entry's heading and mechanism line, replaced with dated versions** — the in-place dated
correction the ruling asks for, not another session's content. I should have said so in the commit
message; recording it here instead.

---

## 2026-08-24 — CONTRAST IS AT ZERO. The last 8 closed themselves, and I did not do it

Checked the ground first, as the thread was two days old and a chassis build had rolled. **Three
things had changed underneath this lane, and one of them finished the job.**

### 1. `bugs_open/364`'s fix IS live — verified by capability, not by grep

The roll happened: chassis pods run `v1.0.1332`, started 2026-08-24 09:39Z.

⚠ **I nearly used the trap.** My instinct was `grep` the binary for my own commit sha. That is the
documented wrong answer, and `platform/buildcapability/buildcapability.go` (RFC_040, owner-ratified
2026-08-20 — *after* my last session) says why in as many words: *"buildinfo.GitCommit is ONE
string, not an ancestry, so grepping the binary for your own commit returns ABSENT for a binary that
certainly contains it. Two lanes have now been burned by exactly this (`bugs_open/215` on
v1.0.1288; `bugs_open/299` on v1.0.1316)."* The `build provenance` log line was also already gone —
retention here is ~4 minutes, so a startup line is unreadable within the hour.

The mechanism that works is new since I last looked: **`service_binary_capabilities`**, one row per
pod carrying `git_commit`, heartbeated every 15 min with a 2-hour retention window. So the question
is a proper query:

```sql
SELECT DISTINCT git_commit FROM service_binary_capabilities
WHERE service='agent-chassis' AND last_seen_at > now() - interval '30 minutes';
--> 0b262ed5e1702127c7e1b8b035eae9e33bdc90f8   (ONE distinct stamp)
```
```bash
git merge-base --is-ancestor ebe8f4323 0b262ed5e   # ✅ contains the 364 fix
```

**Read `buildcapability.go` before ever asking "did my Go change ship?" again.**

### 2. `pricing` rebuilt itself on 2026-08-22, hours after I stopped

All five components carry new `rendered_html` and **non-NULL `content_data`** (they were 5/5 NULL
since April, which was the whole reason the page could not re-render). Two items did it:

- `aiao-557-verify` — **my own evidence item**, which I left at `needs_human_review` with the
  `2am` false positive. It **completed 2026-08-22 16:02:21**. Migration `557` had cleared the `170`
  blocker; the retry drew copy without a clock time and went through. **That is the
  non-determinism I documented, working in my favour for once** — and it means the page was fixed
  by the source change, not by the Go fix, which was not yet live.
- `backfill-353` — added a tool reference on 2026-08-23 17:40.

### 3. The result: **0 firm contrast failures on all four pages**

```
page              firm 08-22   firm now
about.html                 0          0
index.html                 0          0
pricing.html               8          0     <-- the whole remaining backlog
services.html              0          0
TOTAL FIRM                 8          0
```

Two findings remain and both are `overImage` (approximate by the adapter's own admission, and
excluded from every figure this lane has quoted): one on `index`, one on `pricing`
(`.btn btn-secondary`, white on grey, 3.95:1 over an image). **If anyone wants those, they are a
different piece of work — an over-image backdrop is unknowable to this instrument, not a measured
failure.**

**The arc, end to end: 44 → 32 → 8 → 0 firm failures.**

### 4. The literal markdown is gone too

The pricing CTA used to serve `[LLM Provider Cost Comparison Calculator](/tools/…)` as visible
text. The rebuild regenerated it as prose and the page now has **0** markdown-link artefacts. So
the contribution I sent `copy_quality_two_stage` on 08-19 — *"the shrink floor is defending
defective copy"* — is now historical for this page: the floor let a better draft through and the
defect went with it. Worth them knowing; the mechanism they care about is unchanged.

### 5. Carousel and images survived two days and a roll

`index` and `enterprise-reference-deployment` both still carry the carousel markup and **5 card
images each, all resolving**. `carousel_enabled` still set on 2 placements, 10 image URLs still
bound, migrations `469`/`557`/`559` all still in force. ⚠ The rebuild that touched `pricing` did
**not** touch these — the durability risk recorded in `559` has still never actually fired, which
is not the same as it being safe.

### 6. Told the two other consumers, per the owner ruling of 2026-07-29 §3

`finetuning.uk` and `leopardessconsulting.co.uk` place `case-studies-grid`. CONTRIB filed in both
lanes: what changed, that their sites are OFF by construction (guard + verified at their live
pages), how to opt in, and the three traps — the flag is not durable against a rebuild, the arrows
hide on overflow not card count, and **the experience register is not how you ship this**.
Measuring that nothing broke is not the same as their having agreed.

### 7. Still open, and none of it blocks anything

- **The 17 parked `contrast_failure` items are STILL `deferred`, untouched since 2026-08-11.** The
  site's render audit has not run here since 08-10, so the retraction cannot fire — their presence
  remains *not* evidence of a live defect. Now that all four pages measure 0, this site is a clean
  disconfirmable test for `bugs_open/296`: the next rotation should retract them. If it retracts
  none, §9 of that file is wrong.
- **`composeWriterBlock` must learn to carry negative guidance** before any site sets
  `writer_block_managed: true` — otherwise opting in deletes the NEVER-STATE list. Unchanged.
- `overImage` findings (2) — a different instrument problem, not this lane's contrast defect.

---

## 2026-08-25 — my own `557` was publishing `NNN+` to the live site, and another lane caught it

Checked the ground first (thread a day old, heavy automation on this site). **Two CONTRIBs had
arrived and one names this lane's own migration as the cause of a public defect.** Everything below
was verified here before being acted on — a subagent's or a peer lane's report is another doc, not
a measurement.

### 1. CONFIRMED at the artefact: `NNN+` was live

```
curl https://ai-agent-orchestration.com/model-directory.html
  "…against the NNN+ agent types already running in production…"     <- 1 occurrence
```
Censused the rest of the site in the same breath: `adoption-tracker`, `protocol-tracker`, `index`,
`about`, `pricing` — **0 occurrences each, all HTTP 200**. So it was one page, not a sweep.

⚠ **And it lives in `content_data`, not just `rendered_html`** — so a `template_changed` rerender
would NOT have cleared it. Only a copy regeneration does. That is the opposite of the propagation
route this lane has used all week, and reaching for the familiar one would have looked like a fix
and changed nothing.

### 2. The mechanism, and it is mine

`557` put this in `writer_block`: *"take the live value from the `aao-agent-definitions` fact.
Phrase it as **"NNN+ AI agents"**"*. Two independent errors:

1. **`NNN` has no substitution machinery behind it.** I wrote it the way a human reads a style
   guide, into a document whose reader is an LLM that copies what it is shown.
2. **The writer is never given the facts list.** On the unscoped path its prompt contains
   `writer_block` and not the values. So "take the live value from the fact" points at nothing.

Measured by the `387` session: **137** instructed calls since 08-22, **14** copied `NNN` verbatim,
**0** wrote the value; **zero** `NNN` in any writer response before 08-22. The owner-ruled lower
bound was effectively **unstated** on this site for three days.

**I had verified `557` at the artefact and still missed it, because I verified the thing I
CHANGED** — that the block no longer carried a stale number — **and never asked what the writer
would DO with the replacement.** The check was one query away:
`SELECT prompt FROM llm_call_log WHERE agent_type='page-content-writer' ORDER BY created_at DESC LIMIT 1;`
would have shown the facts list absent. WRONG_CALLS logged.

### 3. What the other lane fixed, and what was left to me

Migration `611` (theirs) landed 2026-08-25 11:20:26Z — floors instead of stand-ins, an explicit ban
on letter stand-ins, every ban and the NOT-TRACKED list preserved, and it even carries `557`'s own
history note forward. **It is a better block than mine.** Read it before touching this row.

They flagged one thing back as ours: the `writer_line` field, which `611` deliberately did not
touch. ⚠ **Their report named TWO defects; censusing all seven writer_lines found FIVE, in two
classes:**

| class | facts | why it matters |
|---|---|---|
| frozen date beside a live `{value}` | `aao-agent-definitions`, `aao-agent-types`, **`aao-orchestrations`** (not reported) | managed mode would publish a TRUE number under a FALSE date — worse than either error alone, because it reads as provenance |
| a line that publishes what the block FORBIDS | `aao-orchestrations`, `aao-work-items` | the block says "DO NOT state an exact daily figure" / "DO NOT state a figure"; both writer_lines substitute exactly that. **The block would contradict itself, and the contradiction would be GENERATED — so no author would ever see themselves write it** |

**Taking the report at its stated scope would have left three of the five in place.** Fixed as
migration `613`: five writer_lines repaired and aligned to `611`'s floors, `{value}` removed from
all five, `611`'s `writer_block` and `banned_claims` guard-asserted byte-identical, no fact value
touched. `aao-departments` and `aao-services` deliberately left alone — `611` decided those stay
exact and silently widening scope is what the file disclaims.

⚠ **This does NOT clear the way for `writer_block_managed`.** `composeWriterBlock` still builds from
`writer_line`s and `allowed_entities` and nothing else, so flipping the flag still deletes the
NEVER-STATE list. `613` removes ONE precondition of several. The `387` session has proposed the
missing `writer_block_guidance` carry to the `bugs_open/288` lane, which owns that file.

### 4. The 364 lane's "three of your pages are 404" was REFUTED, and not by me

Their 08-24 CONTRIB said `adoption-tracker`, `protocol-tracker` and `model-directory` were deployed
and 404. The `387` session refuted it: they curled the **extensionless** form, which 404s for every
page on this site by hosting design (`scripts/cloudflare/worker.js:40-44`) — `/about` 404s the same
way. Verified here today: all three serve **200** at `/<name>.html`. **Nothing about this site's
deploys was ever wrong.** Recorded because the refutation arrived from a third lane, and a reader
of the 08-24 CONTRIB alone would still believe it.

### 5. A process note: my WRONG_CALLS entry was swept into another session's commit

I wrote the entry, and before I committed it another session's commit (`001211abf`, the 364 lane)
carried it in. **Nothing was lost** — verified properly, not by grep: `git diff --numstat HEAD` on
the file returns EMPTY, i.e. the working tree is byte-identical to HEAD, so the entry landed whole.

> **⚠ CORRECTED 2026-08-25 by the `bugs_open/381` lane — IT WAS ME, NOT `001211abf`.** Your entry
> was swept in by **`3d31b86a9`** ("381: 'served' was FALSE — homegarden.uk is a parked domain…"),
> which named `WRONG_CALLS.md` in its pathspec while your entry sat uncommitted in the shared tree.
> I am correcting this rather than leaving it because it is my sweep, your session is finished and
> cannot answer, and the `364` lane is carrying blame for something it did not do — it declined to
> move that blame onto me sight-unseen, which was the right call and is why it reached me instead.
>
> **The evidence, and the reason your original attribution was reasonable:**
> ```
> git log --oneline -S 'I wrote an EXEMPLAR into a prompt and it shipped to the public as copy' \
>   -- docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md     # -> 3d31b86a9, alone
> git show 001211abf -- docs/…/WRONG_CALLS.md | grep -c 'EXEMPLAR into a prompt'   # -> 0
> ```
> **`git log -N -- <path>` answers "what last touched this file", never "what introduced this
> content"** — and on a shared tree those differ within minutes, with no tell, because a
> path-filtered log always returns something recent and plausible. Two lanes made that exact
> mistake on the same day and both landed on `001211abf`. Use the pickaxe (`-S`), and then
> `git show <sha> -- <path> | grep '^+' | grep -v '^+\s*//'` to check the match is added content
> rather than a prose mention.
>
> **Your byte-compare conclusion is untouched: nothing was lost, and the entry is intact in HEAD.**
> Only the sha was wrong.

⚠ **Two grep traps in that check, both of which I hit first.** A line-oriented `grep -F` for a long
phrase reports FALSE ABSENCES because these files are hard-wrapped — two of my four probe phrases
returned 0 and both were present. Unwrapping (`tr '\n' ' ' | tr -s ' '`) returned them. And a probe
phrase that differs from the file by so much as a `**` returns 0 for a reason that has nothing to
do with the question. **The byte-compare is the check; phrase-grepping is a guess.**

This is CLAUDE.md's documented shared-tree hazard, from the other side: committing per task stops
me sweeping others' work, and cannot stop theirs sweeping mine. Forward-only holds, nothing to undo.

### RESOLVED 12:41Z — and a correction to my own reasoning an hour earlier

**The incident is closed at the artefact.** `model-directory` regenerated at 12:40:57; `content_data`
and `rendered_html` both clean; all seven pages checked read `NNN=0` and HTTP 200. The hero now
says *"more than 150 distinct agent types"* — `611`'s floor phrasing, true at 200 and immune to
drift.

> **⚠ CORRECTED, same session, before it reached the handoff.** At ~12:35 I wrote that the
> 6-hourly `model-directory-publish` *"does scoped rerenders … so it can never clear the
> placeholder"*, and was about to file a `needs_page` myself to force it. **That was wrong**, and
> the evidence I used was the task's own `description` field plus a single observation that the
> 12:26 publish had regenerated nothing at 12:35.
>
> What I had not done was **look at the queue**. The publish DOES lead to a full rebuild: it filed
> a `needs_page` (`created_by='render_directory'`) at **12:25:13**, which was still `claimed` while
> I was concluding it would never happen. It completed at 12:41 and cleared the placeholder.
>
> **Two lessons, and the second is the one that generalises:**
> 1. I read a *description of a mechanism* and treated it as the mechanism. `scheduled_tasks.description`
>    is prose someone wrote; `site_work_items` is what the thing actually did.
> 2. **"It hasn't happened yet" and "it cannot happen" look identical for exactly as long as the
>    work is in flight.** My 9-minute observation window was inside one job's runtime. Had I filed
>    the forcing rebuild I intended, I would have duplicated a running non-idempotent page build —
>    the precise damage `bugs_open/029` records — and then credited the fix to my own item.
>
> **What would have caught it in one query, and did:**
> `SELECT item_type, status, created_by FROM site_work_items WHERE site_id=… AND status IN ('claimed','triaged')`
> — before concluding that nothing is happening, ask what is claimed.

**So handoff §5 item 1 is CLOSED**, and the disconfirming result it named ("if it is not 0 several
hours after 611/613, the regeneration is not picking up the new block") did not occur — the
regeneration picked it up on the first rebuild after the source fix.

---

## 2026-08-25 ~15:55–17:05Z — session "ai-agent-orchestration": re-verification, two stale handoff claims, the managed flip PREPARED (617 HOLD), a 296 measurement

### 1. Re-verified at the artefact (all [MEASURED 2026-08-25 ~16:00Z])
- 7/7 pages HTTP 200 with `NNN=0` — index, about, pricing, services, model-directory, adoption-tracker,
  protocol-tracker. Stylesheet 20,923 bytes (not clobbered — `bugs_open/296` §10.4's undercount caveat does
  not apply to this site).
- `evidence_base` current row = `created_by='613_migration'` (12:09:40Z); keys `facts, writer_block,
  banned_claims`; `writer_block_managed` unset; no `writer_block_guidance`. The refresher's last supersede was
  **09:06:24Z — BEFORE 611/613** — so tomorrow's ~09:06Z pass is the first that could disturb them (the 387
  lane's pinned check; copied into the handoff).
- Live chassis: `service_binary_capabilities` → `4c996e1b5` (one distinct sha, started 09:27Z). merge-base:
  `c17a18620`, `14ec48b89`, `cbadcba71` (the CLM-029 carry) are **NOT** ancestors.

### 2. Two claims in this morning's handoff were STALE WHEN WRITTEN (WRONG_CALLS logged)
(a) *"The 17 parked contrast_failure items — still deferred, untouched since 08-11."* → **9** deferred. Eight
were `cancelled` at **2026-08-24 19:11:22Z** — 18 hours before the handoff — with
`result.reason` = *"The item_key selector is TAG.TAG — invented by the render audit's class fallback for a
class-less element (bugs_open/352). It matches nothing on any…"*. The 17 was carried from the 08-22 handoff
without a count.
(b) *"because the site's render audit has not run here since 08-10."* → `site_discovery_rotation` has
`render-audit-agent.last_selected_at = 2026-08-24 02:23:11Z`. The inference came from
`max(created_at)` of `contrast_failure` rows (08-10): **an audit that files nothing leaves no row, so the
absence of rows is not the absence of runs** — the rotation table is the record of selection.

### 3. Handoff §3 is stale: the guidance carry IS BUILT — by the 387 lane, not 288
`writer_block_guidance` = **CLM-029** (commits `c17a18620` → `14ec48b89` → `cbadcba71`, council
`0de22385` APPROVED r2), built "as agreed with the 288 lane (option b, their window)". Seated once in
`composeWriterBlock` (`refresh_evidence_base_action.go:1339`) and copied through the scoped path
(`plan_sections_action.go:2593`). Contract (register): **NEGATIVE/PROHIBITIVE guidance ONLY**. Inert until
the roll. **The flip stays forbidden today, and I measured why rather than repeating it:** the live row run
through the REAL `composeWriterBlock` (a `go test -overlay` harness — the test file is mapped in from the
scratchpad, nothing touches the shared tree; recipe in RUNBOOK R10) produces the seven number lines and
**no prohibition at all**:
```
NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine):
- more than 150 active agent definitions in the production registry
- more than 150 distinct agent types
- 8 departments — Strategy, Research, Content, Design, Development, Quality, Operations, Data (the platform's OWN taxonomy, never 'departments served')
- more than 20 live sites in production, built and operated end-to-end by the platform
- 17 backend services
- automated work items completed: state NO figure — the ledger is reaped, so this count falls as well as rises
- over a thousand orchestrations a day (never an exact daily figure — a rolling 24-hour window is stale within hours)
```

### 4. The 296 test — measured, and it is a result for THAT lane
- Selected 02:23Z 08-24; the run's `orchestration_states` row is **beyond retention** (oldest row in the
  table: 08-24 13:09Z, 7,898 rows), so completed-vs-timed-out is unknowable. Fleet: 9 COMPLETED
  render-audit runs since 08-24 15:29Z, every one with a matching rotation stamp — the rotation fires.
- The 9 deferred rows are unmoved (`updated_at` 08-11 12:31Z). R1 over the six pages they name: **about 0,
  services 0**; news / index / tools report only an over-image `A.btn` at 3.95:1 (a different selector,
  approximate by the adapter's admission); **ai-readiness-quiz NOT MEASURED** ("probe produced no result").
- Element census on the served pages (`grep -o 'class="[^"]*\b<cls>\b'`): **7 of 9 selectors still present
  and now passing** — index `P.news-card-summary` ×5, about `A.cta-btn` ×2 / `SPAN.stat-value` ×3 /
  `SPAN.about-eyebrow` ×1, services `SPAN.info-card-grid__eyebrow` ×1 / `H3.info-card-grid__card-title` ×6,
  tools `H3.tl-card-title` ×6. news `A.cta-btn`: **0 elements** (gone). ai-readiness-quiz `BUTTON.btn` ×4,
  unmeasured. So these are REPAIRED (this lane's 08-17→08-24 work), not vanished.
- `retractResolvedContrastFindings` closes a parked row iff its page is in the run's `pages_audited` and the
  key is absent from the run's failing set. Eight of nine qualify today. **It did not fire.** Candidates, not
  distinguishable from what survives: (i) the 02:23Z run timed out (`site-render-audit-rotation`
  `timeout_seconds=1800`; 43 pages; `max_pages 60`) — 296 §10.2's class; (ii) `pages_audited` did not
  include these pages. **The next selection is due ≥ 2026-08-27 02:23Z** (3-day interval, `ORDER BY
  last_selected_at ASC`), inside retention — that run is the discriminating one. Contributed as
  `bugs_open/296` §11.
- Ruled out: the 352 skew guard (`ffa6e1c3d`, 08-24 13:45Z) postdates the run, but all nine keys are
  classed `TAG.class`, for which old and new selector compositions are identical.

### 5. The flip, PREPARED — `sql_for_agents/617_aiao_writer_block_managed_with_guidance_carry_HOLD.sql`
Council corr **`35ab8b23-5f22-457f-a8c8-92baad862422`** (admission dry-run passed; 2 edits; in scope because `_HOLD.sql` IS the change).
- **What it writes:** `writer_block_managed=true`; `writer_block_guidance` = every prohibition 611 carried,
  verbatim, plus two categories the site's `banned_claims` already enforces (systems shipped for clients,
  client sectors served) — stated, not silent; an 8th fact `aao-architecture`, **valueless** (`EvidenceFact.Value`
  is `*float64` — a string value would break every reader's parse), so 611's positive "Architecture:
  Kubernetes, Kafka, Postgres" line rides the composer's CAPABILITIES section instead of guidance (which may
  not carry positive statements); and `writer_block` = the EXACT 1,993-byte output of `composeWriterBlock`
  over the proposed document (same jsonb expression, run read-only, fed to the real function). 611's block
  retired. 7 existing facts byte-identical; `banned_claims` untouched. Both asserted by guard.
- **Why HOLD:** the carry is not live (§1). On today's binary the flip deletes the prohibitions at the next
  refresh (§3's measurement). The file's CHASSIS GUARD refuses: no sha; the pre-carry sha `4c996e1b5…`
  (hardcoded); a sha that is not the one heartbeating in `service_binary_capabilities` (30-min window); a
  mixed fleet. **It cannot compute git ancestry** — R10's one-liner runs `git merge-base --is-ancestor
  c17a18620 $LIVE` first and only then pipes the file. A guard that makes you type the sha you checked.
- **Rehearsed against the live DB, all doomed transactions:** (a) no `-v` → refused, exit 3; (b) `-v
  live_chassis=4c996e1b5…` → "PRE-CARRY chassis"; (c) unknown 40-hex → "not the chassis that is RUNNING
  (4c996e1b5…)"; (d) guard stripped, `COMMIT→ROLLBACK` → `NOTICE: 617 OK`; (e) migration THEN rollback in ONE
  transaction → both OK, prior document byte-identical; (d') both writer_block constants minus
  ", uptime percentages" → `617: the composed writer_block lost a phrase 611 carried: uptime percentages` —
  the phrase guard is live, not vacuous. Afterwards: current row still `613_migration`, 0 backup rows.
- ⚠ **Number collision, again:** `616_css_patch_agent_prompt…` appeared at 16:40:38, two minutes before my
  write at 16:42:34, both untracked. Caught by the "re-check immediately before writing" line in R9; renumbered
  617 (free), internal references `sed`-checked to zero strays.
- **The live disconfirming test, after application:** the first ~09:06Z refresh should supersede the 617 row
  with an `evidence-refresher` row whose `writer_block` is **byte-identical** to the constant
  (`existing != block` is the only condition for regeneration). A different block means my prediction of the
  composer's output was wrong — read the diff; do not touch the refresher. Query in R10.
- The composed block, for the record:
```
NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine):
- more than 150 active agent definitions in the production registry
- more than 150 distinct agent types
- 8 departments — Strategy, Research, Content, Design, Development, Quality, Operations, Data (the platform's OWN taxonomy, never 'departments served')
- more than 20 live sites in production, built and operated end-to-end by the platform
- 17 backend services
- automated work items completed: state NO figure — the ledger is reaped, so this count falls as well as rises
- over a thousand orchestrations a day (never an exact daily figure — a rolling 24-hour window is stale within hours)

CAPABILITIES (assert without inventing numbers):
- Architecture: Kubernetes, Kafka, Postgres — true and stated freely

NEVER copy a number out of a page, a template, or an older spec: the NUMBERS list above is the ONLY authority for figures about this business, and if a count is not written there, write no count at all. NEVER restate a listed floor as any other number. NEVER write letters as stand-ins for digits (a letter repeated where a number belongs): a stand-in is a defect that ships, not a placeholder. NEVER write "over 70 specialised AI agents", "70+ agents" or "30+ agent types" — true as lower bounds and understating the fleet by more than half (owner ruling 2026-07-27). NEVER frame the 8 departments as departments of external clients ("departments served"): they are the platform's OWN organisational taxonomy. NEVER state an exact daily orchestration figure, and NEVER state a total of automated work items completed. NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent instances" is not measured), uptime percentages, systems shipped for clients, client sectors served. None of these are measured; every such figure at any value is an invention.
```

### 6. Seen, not mine — recorded for the handoff
`tool-automation-savings-estimator`: two `page_rerender` rows **failed** (08-24 21:00Z from
completeness-discovery `cta_links_stale`; 08-25 14:30Z from `bugfix_384_toolcta_fanout`
`template_changed`), both at `save_page_sections: SECTION COMPONENT FLOOR REFUSED … 77→37 class attributes
(48% kept, floor 50%)`. The 384 lane has it (their NOTES:198): the refusing slot is the page's own bespoke
tool section, "that page already failed 3 times on 2026-08-24, before any of this lane's work" —
pre-existing divergence between stored HTML and what template+`content_data` regenerate. `bugs_open/253`
class. Until reconciled, **every rerender of that page fails the same way**, which also means the 08-24
misdirected-CTA fix for it never lands. Not investigated here.

### 7. Council verdict on 617 — APPROVED r1 (`35ab8b23`), 1 medium + 4 low advisories, all acted on
Run started the moment it was published (15:46Z; no queue today), `complete_approved` ~16:1xZ. Seven seats
abstained (out of footprint), five spoke: editquality (object, advisory), reuse_agent, guardian,
compliance, prior_art_librarian (approve with notes), plus debug_historian / constitution / mission /
adoption_guardian (approve). Actions, same session, same file:
- **editquality, medium — "durability is proven only at apply and the first refresh; the typed-struct
  landmine could strip the key on a LATER refresh."** Answered with evidence, not code: CLM-029's own round
  surveyed all **9** `ParseEvidenceBase` callers (readers/guards; the admin handler stores the client's
  bytes) and pinned both real write paths (`refresh_evidence_base`'s raw-map marshal;
  `write_site_spec`'s `siteSpecDeepMerge`) with round-trip tests in `writer_block_guidance_387_test.go`;
  `writer_block` itself is equally unlisted in the struct and has never been lost. Recorded in the file
  header and — because "expect it by looking" — R10 now says run the survival query after the SECOND
  refresh too (the first pass that re-reads a refresher-written row).
- **editquality, low — unused `bad int;`.** True; removed.
- **guardian + compliance, low — the guard cannot compute ancestry; an operator skipping merge-base after
  some other un-carried roll would pass.** Strengthened with a NECESSARY condition the DB can check:
  `min(started_at)` of the running sha's pods must be after `c17a18620`'s commit time (12:49:19Z 08-25).
  Refuses "applied before any post-carry roll" for ANY sha; cannot see a restart on an old image or a
  reverting roll — stated in the guard's own comment; merge-base stays the operator's. **Mutation-proven:**
  with the by-name refusal deleted and today's running sha passed, the file refuses on `started_at`
  (09:27:32Z < 12:49:19Z).
- **reuse_agent, low — do the 6 managed sites share a reusable opt-in tool this duplicates?** Measured: no
  file under `scripts/` or `cmd/` mentions `writer_block_managed`; the six current managed rows were
  written by `portfolio_positioning Wave 1` / `Phase C pilot seed` (hand SQL, 08-17/18),
  `brochure_component_library` (07-24), `claude-session-copyquality-20260815`, and the refresher carrying
  earlier hand seeds forward. **Per-site hand SQL is the estate's actual shape**; a shared opt-in helper
  would be a new mechanism and belongs to the 288/387 lanes if anyone wants it — not built here.
- **prior_art_librarian, low — the census claims are unverifiable from its schema list.** They were
  measured here (§1, §5); nothing in the plan hinges on them.
Re-rehearsed after the edits: (a)(b)(c) refuse; (f) NEW started_at refusal isolated; (d) OK; (e) apply +
rollback OK; (d') phrase mutation caught. Live row still `613_migration`, 0 backup rows. Committed with
`Council-Reviewed: 35ab8b23…` — the verdict was read first.

### 8. My WRONG_CALLS entry was swept by another lane's commit — attributed correctly this time
`git log -S'a count carried into a handoff without a query' -- WRONG_CALLS.md` → **`26572c627`** (the
`bugfix_390` lane, "bug sweep" session), alone. Their entry and mine were both uncommitted in the shared
file; their pathspec commit took both, as CLAUDE.md says it must. Nothing lost — my entry is at HEAD
byte-for-byte (`git show HEAD:…WRONG_CALLS.md | grep -c '<phrase>'` = 1). The pickaxe, not `git log --
<path>`, is how §5 of the earlier session went wrong; used it first here.

---

## 2026-08-25 (evening) — the four-page scope was never questioned, and the site has 42 pages

Checked the state again after another roll. **Three findings, and the first is a correction to
everything this lane has reported all week.**

### 1. ⚠ "Contrast is at zero" was true of FOUR pages. The site has FORTY-TWO.

Audited all 42 active pages (enumerated from `pages`, not guessed):

```
pages audited 40 of 42     NOT MEASURED 2
TOTAL FIRM 17              overImage 23 (excluded, as always)
```

| page | firm | state |
|---|---|---|
| `/tools/agent-complexity-estimator.html` | 6 | template HAS the ink fix, placement stale since **2026-05-01**; `owned` |
| `/contact.html` | 4 | 3 = the 456 fix never propagated (placement stale since 08-11); 1 = white-on-amber button |
| `/tools/automation-savings-estimator/index.html` | 3 | `html_ink` already TRUE — **different cause, needs diagnosis** |
| `/tools/password-entropy.html` | 2 | template never fixed (`#666` on `#0D1117`); `owned` |
| `/tools/build-vs-buy-analyzer/index.html` | 1 | `html_ink` TRUE — different cause |
| `/tools/tool-llm-cost-calculator.html` | 1 | template HAS ink, placement stale since 08-11; `owned` |

**Nothing regressed.** index/about/pricing/services are still clean. The failures are on pages this
lane never measured, because the four-page scope came from the originating handoff and **I never
questioned the denominator.** "44 → 0" was always "44 → 0 on these four pages", and I did not say so.

⚠ **Two pages return zero because they CANNOT be measured**, reproducibly, while serving HTTP 200:
`ai-readiness-quiz.html` and `tool-ai-agent-roi-estimator.html` — *"probe produced no result"*.
`render_audit.py` prints *"the zeros above are silence, not a pass"* and that is exactly right. Both
are tool pages; a JS-heavy page that replaces its body can lose the injected probe. **Any "the site
is clean" claim has a two-page hole in it until that is understood.**

### 2. The dominant remaining defect is ONE mechanism, and mostly needs no new code

`#0D1117` on `#0D1117` (×8) and `#0D1117` on `#080B10` (×3) — family A, the same ink defect
migrations 456/469 addressed. For three of the six pages the **template already carries the fix and
the placement was never re-rendered**. That is a propagation gap, not a code gap.

⚠ **But three of the six are `rebuild_policy='owned'`, and a re-render is REFUSED there.** Those
pages hold their whole tool — calculator and `<script>` — in a single verbatim component, which is
exactly why they are owned. **Do NOT flip them to `generic` to get the re-render through**: the
LANDMINE on this is explicit — the composition loop commits freshly-written HTML to the deploying
repo one step BEFORE `save_page_sections` refuses, so the tool is replaced with prose and shipped.
The route for those is the owned-page chrome path (`refresh_owned_page_chrome.sh`), not this lane's
usual rerender.

Filed the one safe fix: a `template_changed` rerender for `contact` (`generic`, template fixed,
placement stale). ⚠ It is **queued behind another lane's claimed work** — `bugfix_391_cta_relevance`
holds 2 `content_rewrite` rows `claimed` on this site, and one claimed row of any type holds the
per-site mutex (`029`). Not poked; it drains.

### 3. Two stale claims in MY handoff, corrected by another session before I re-read it

The `617 HOLD` session found and fixed both:

- **"17 parked `contrast_failure` items"** → there are **9**. Eight were `cancelled` on 2026-08-24
  19:11:22. My 17 was measured 08-18 and repeated for a week without re-measuring.
- **"the site's render audit has not run here since 2026-08-10"** → it visited **02:23Z on 08-24**.

Verified both here today. **Both were counts I carried forward instead of re-running** — which is
precisely what the owner's 2026-08-22 ruling on dating counts exists to prevent, and I had applied
that rule to *new* numbers while leaving the inherited ones undated.

⚠ **And the 9 that remain matter more than the 17 did**: 6 of them are on `about`/`services`/`index`,
which now measure **0**, so the retraction should close them; 3 are on `news`/`tools`/
`ai-readiness-quiz`. I measured those: news **0 firm**, tools **0 firm**, quiz **NOT MEASURABLE**.
So `bugs_open/296`'s test on this site is sharper than I described: **8 of 9 should retract and 1
cannot be decided by this instrument at all.**

### 4. Another session has prepared the managed-mode flip I said was blocked

`617` (HOLD, council APPROVED r1) carries `611`'s prohibitions into `writer_block_guidance` —
CLM-029's first consumer — pre-composes `writer_block` to the real composer's exact bytes, and adds
a chassis guard refusing any pod that predates the carry. **So the blocker my handoff describes is
being cleared by someone else.** Do not re-derive it, and do not flip managed mode by hand: `617` is
held for ordering and applies deliberately.

### Same evening, continued — 17 → 5 site-wide, via a route the ordinary one could not take

**`contact` re-render landed: 4 → 1.** The three ink failures cleared exactly as predicted; only
the white-on-amber button survives, which was the prediction.

**Then the three `owned` tool pages, 9 failures, as migration `625`.** This is the interesting part
because *both* obvious routes were wrong:

- **Flip `rebuild_policy` to `generic`** so a rerender goes through → the documented tool-clobber.
  The composition loop commits freshly-written HTML to the deploying repo **one step before**
  `save_page_sections` refuses, so the calculator ships as prose. Calculators have already been
  destroyed this way (367 → 377 re-lock).
- **`refresh_owned_page_chrome.sh`** (the leopardess route) → safe, but **inert for this**: it
  re-renders in ASSEMBLE mode, which re-assembles the **stored** HTML — exactly what carries the
  stale CSS. It fixes chrome, not section CSS.

**So the operation was a surgical patch of the stored artefact** (precedent: migration `393`), and
it is only safe *now* because `bugs_closed/229` (live since v1.0.1276) gave `rendered_html` writes a
comparison and an archive. Before that fix this would have been reckless.

⚠ **And then assemble mode became the right deploy vehicle after all** — once the stored HTML is
correct, re-assembling it is precisely what ships it. That inversion is the lesson: *the same
mechanism was useless and then necessary, and what changed was not the mechanism but the state it
reads.* Ran the script per page; **ownership restored on all three and asserted afterwards**
(`owned`, scripts present, lengths 22,732 / 5,271 / 35,940).

**Four defect shapes, each measured against its own ground before being changed:**

| shape | before | after | note |
|---|---|---|---|
| bare foreground `--color-primary` (`h2`, `.ace-legend`) | 1.00 / 1.04 | 5.66 / 5.90 | 456's defect on pages 456 never reached; `.ace-legend`'s TEMPLATE was already correct — only the artefact was stale, since **2026-05-01** |
| label on a primary fill (`.estimate-btn`, `.calc-btn`) | 1.04 | 18.92 | 457's shape in its primary variant |
| inline `style="… color:#666 …"` (×2) | 3.30 | 6.15 | **not a rule at all** |

⚠ **The inline-style case is why this page read as clean to every rule-shaped query I ran.** I
searched `{...}` blocks inside `<style>` and found nothing, twice. A colour can be applied by an
attribute, and a CSS-rule query cannot see it. **`--color-on-primary` is NOT emitted on this site**,
which is why `.calc-btn`'s existing fallback chain still failed — a chain is only as good as its
first *emitted* link.

### ⚠ `render_audit.py` can sample BEFORE the stylesheet applies — the same element read two different wrong colours

The `contact` button measured **2.08:1 white on `#F0A500`** in one run and **1.15:1 white on
`rgb(239,239,239)`** in the next. `rgb(239,239,239)` is the UA default `buttonface` — i.e. the
second run sampled before the custom background applied. I nearly wrote up a self-inflicted
regression.

`getComputedStyle` on the live page settles it: **`bg: rgb(240,165,0)`, `color: rgb(255,255,255)`** —
the button is amber, the true ratio is **2.08**, and the stored CSS rule is byte-identical either
side of my re-render, so nothing regressed. **A differing measurement of an unchanged thing is an
instrument reading, not a finding** — and this instrument's failure mode is to report a *plausible
alternative colour*, not an error.

### Where the site actually stands, all 42 pages

```
FIRM failures site-wide:  17  ->  5      (42 pages, 2 unmeasurable)
  3  /tools/automation-savings-estimator/index.html   SUMMARY + BUTTON at 1.00, A at 1.09
  1  /tools/build-vs-buy-analyzer/index.html          BUTTON at 1.00
  1  /contact.html                                    .form-submit, white on amber, 2.08
```

The two remaining tool pages already carry the ink token in their stored HTML, so their cause is
**not** the 456 defect and guessing would waste a cycle. The contact button's fix is known and
computed — `--color-accent-text` is `#294155` here, giving **5.09:1** on amber — but
`contact-form` is on **20 sites**, so it is a fleet change requiring the consumers to be told first.

## 2026-08-26 ~09:17Z — the survival check PASSED; the routed floor-refusal row judged (left at failed)

### 1. 611's block SURVIVED its first daily refresh — the 387 lane's pinned check, run and answered
Current row: `created_by='evidence-refresher'`, created 2026-08-26 **09:07:07Z**; `writer_block` md5
**`f7fd6efd737228e6505e5653b5ef93e9`** — byte-identical to BOTH the 611 and 613 rows' blocks; `NNN` regex
**f**; NOT-TRACKED list present; 7 facts; no guidance key (617 not applied — no roll yet). So the
refresher carries an unmanaged block forward unchanged, exactly as `refresh_evidence_base_action.go:36`
claims — now a measurement, not a claim about behaviour. Postscript added to our CONTRIB in the 387 dir.

### 2. The loanzy lane routed work item `0229af86` here — judged: real defect, correctly NOT re-fired
Their outage sweep (21 credit casualties re-fired) left one row on this site at `failed` for hand
judgement. Read in full: it is the **misdirected_cta rerender of `tool-automation-savings-estimator`**,
re-filed by completeness-discovery at 08-26 00:56Z (the 08-24 row `6db1db3d` being terminal frees the
dedup key), failed 02:49Z on the identical **SECTION COMPONENT FLOOR REFUSED — 77→37 class attributes
(48% kept, floor 50%)**. NOTES 08-25 §6's defect, not a credit casualty; the peer's triage was right.
**Judgement: leave at `failed`** — the state is honest (a wanted fix blocked by a real defect), re-firing
burns a cycle to produce the same refusal (4 identical failures since 08-24), and cancelling would hide
the one row that keeps this defect visible in "what is biting". The fix is reconciling the page's stored
HTML with what template+content_data regenerate (`bugs_open/253` class) — still unowned, still this
site's open item.
⚠ **Predictable follow-on, recorded so nobody reads it as news:** the 01:00:54Z `rerender-pages` batch
holds two more triaged rows for this same page (plus `derive_card_asset`'s 02:31Z card_landed row — the
384 mechanism, working). Any of them that reaches `save_page_sections` on this page will fail with the
same floor refusal. When they do, that is the KNOWN defect firing, not a new one.

### 3. Design-discovery rotation RESUMED fleet-wide (09:20Z, peer heads-up) — our defences checked, our slot estimated
The `webdesign-tool-rebuilds` seat says the design rotation (paused since the 08-11 cost scare,
`bugs_open/401`) was re-enabled 2026-08-26 09:20Z, ~1 site/3h, least-recently-visited first, findings born
`detected` and auto-promoted where (item_type, handler_agent) is known-good. Checked rather than filed away:
- **Queue position:** the six stalest design stamps are all 08-09 ≤14:52Z; ours is 08-09 16:53Z → **~8th**,
  so expect our visit **~2026-08-27 morning**. A surprise design item before then is NOT the rotation.
- **The colour-churn defence is in place:** `design_intent.palette.reference_values` present (seeded by this
  lane 2026-08-18) and it deliberately records the dark `primary: #0D1117` (= `surface` — the degenerate
  palette PLAN §0 describes, which is why component-side fixes, not palette churn, took contrast to 0).
  Spec row `pinned = f` [NOTED, not fixed — the landmine's named pin is the reference_values key itself].
- **6 `image_url_404` rows sit `detected` from the LAST design visit (08-09)** — stale: all 10 card images
  verified 200 on 08-25. If the resumed pipeline promotes or re-files these, PLAN §2's warning stands: do
  NOT hand them to the triage-only image handlers; a fresh visit should re-check/retract.
- **Today's item flurry is NOT the rotation:** `audit_tool`/`evaluate_tools`/`improve_tool`/`acceptance_run`
  (08-26) = the tool-rebuild lane working this site's tools; `needs_content_image`/`undeployed_asset`
  completes + `derive_card_asset` rerenders = the 384 card mechanism. Three concurrent producers — do not
  conflate.

## 2026-08-26 ~09:45Z — the roll landed overnight; 617 APPLIED per R10; one attribution corrected, one R10 claim corrected

> **CORRECTED 2026-08-26 (caught by the webdesign-tool-rebuilds seat, ~09:30Z):** §3 above attributed
> today's `audit_tool`/`evaluate_tools`/`improve_tool` items to "the tool-rebuild lane working this site's
> tools". **Wrong — nobody in that lane touched this site.** `created_by` says `design-discovery-agent`
> (verified): they are the Track 2 contract-rules checker (tool_health rules 16/17) running INSIDE
> design-discovery sweeps that the improvement-loop (owner re-enabled ~21:18Z 08-25) dispatches as
> children overnight. Loop visits do NOT stamp `site_discovery_rotation`, so the ~08-27 LRV slot estimate
> stands. Attribute such items to "the tool_health check, loop-driven" — the lane wrote the rule, the
> platform runs it everywhere. My error was inferring authorship from the messenger's lane instead of
> reading `created_by` — the column was one query away and I had already run its neighbour.

### The roll: chassis `2fb40a960` since 23:11Z 08-25 — the carry is LIVE
`service_binary_capabilities`: one distinct sha, pods started 2026-08-25 23:11:34Z. `git merge-base
--is-ancestor` — `c17a18620` ✓ and `cbadcba71` ✓ IN the running binary. R10's gate condition met.

### 617 APPLIED 09:41:16Z — every guard passed on the real path
`psql -v live_chassis=2fb40a960f…` → backup INSERT 1, supersede UPDATE 1, insert INSERT 1,
`NOTICE: 617 OK`. Verified at the row: `created_by='617_migration'`, `writer_block_managed=true`,
`writer_block_guidance` present, **8 facts**, `writer_block` md5 **`fa0a4710733590782c109d2971ef760d`** =
the file's `$WB$` constant md5 exactly, `NNN` regex f. **611's interim block is RETIRED; the prohibitions
now live in the CLM-029 carry.** The backup row (`migration_backups`,
migration_name='617_aiao_writer_block_managed_with_guidance_carry') holds the superseded 09:07Z refresher
document for the rollback sidecar.

> **CORRECTED 2026-08-26, my own R10:** it said "*a `--record-only` for a `_HOLD` file works (the record
> path takes any filename)*". **False — the runner REFUSED it**: "UPPERCASE-suffixed sidecar … recording
> one is meaningless." I asserted the record half without testing it (I had only tested the apply-side
> exclusion). No dangling state: a `_HOLD` file is also grepped OUT of the pending listing, and the
> durable application record is the DB itself (the `617_migration` spec row + the migration_backups row).
> R10 corrected in place.

### Owed next — the disconfirming tests, now LIVE tests
- **2026-08-27 ~09:06Z** — first MANAGED refresh: expect a fresh `evidence-refresher` row with
  `writer_block` **byte-identical** (md5 `fa0a4710…`). Different md5 + NOT-TRACKED still present = my
  composition prediction was off (diff it, record it); NOT-TRACKED absent = the carry is not doing its job
  — run the ROLLBACK sidecar and investigate.
- **2026-08-28 ~09:06Z** — day-2 (first re-read of a refresher-WRITTEN row; the typed-struct question).
- Also owed by the platform, not us: the next page regeneration writes copy from the COMPOSED block for
  the first time — same floors, so no visible change is the expected result.

---

## 2026-08-26 — the two remaining tool pages diagnosed, and half my fix was a no-op

Roll completed overnight: **one** chassis stamp now (`2fb40a960`, ~66,912 pods), not the two of
yesterday. Yesterday's `625` work is intact — all three owned tool pages still `owned`, scripts
present.

### 1. The diagnosis, done by asking which declaration WINS

`getComputedStyle` plus a CSSOM walk matching each failing element against
`document.styleSheets` — not a stylesheet grep, per this lane's own rule. Result: **none of the
four failures is the 456 defect**, which is exactly why they survived 456/469/625 and why their
stored HTML already carried the ink token.

**The root is 456's root wearing a different face.** On this site `--color-primary` and
`--color-surface` are **the same value** (`#0D1117`). So any rule pairing them collapses to 1.00:1
— and it reads as *perfectly sensible CSS*, because on every other site in the fleet a label in
`--color-surface` on a `--color-primary` fill is exactly right. That is why nobody wrote a bug: the
CSS is correct everywhere except where two tokens happen to coincide.

| declaration | measured | repoint |
|---|---|---|
| `.method-details summary` `var(--color-primary)` | 1.00 | ink 5.66 |
| `#…-calculate-savings-button` `var(--color-surface)` | 1.00 | primary-text 18.92 |
| `.cta-link` `var(--color-secondary)` | 1.09 | accent-ink 9.09 |
| `.bvb-btn-primary` `var(--color-surface)` | 1.00 | primary-text 18.92 |

⚠ **`.cta-link` is an OVERRIDE, not an omission.** The template already carries
`a { color: var(--color-accent-ink, var(--color-accent)) }` at 9.09:1; `.cta-link` has higher
specificity and replaces it with a near-black. The fix is convergence on the token the sibling rule
already uses.

⚠ **`.result-value` is a FIFTH defect the audit cannot see.** It lives in a result panel hidden
until the visitor runs the calculator, so `render_audit.py` — which reads the page as loaded —
never reports it. Same 1.00:1 collapse. **Found by censusing the template for the collapsing PAIR
rather than by fixing what the instrument listed. On a page with conditional UI, the audit's count
is a LOWER BOUND.**

### 2. ⚠ Half of migration `636` is a NO-OP, and I found it only when filing the propagation

- `tool-build-vs-buy-analyzer-ai-agent-orchestration-com` — **1 placement**. Fix correct; rerender
  filed and queued. ✅
- `tool-automation-savings-estimator-ai-agent-orchestration-com` — **ZERO placements.** The page
  renders `tool-automation-savings-estimator-**fundamentallyai-com**` instead, which is placed on
  **2 sites** (fundamentallyai.com and here).

**I selected the components by CSS CONTENT and never asked which component the PAGE uses.** The
template census answered "where does this rule live?" and I read it as "which component does this
page render?" — two different questions with the same-looking answer. An `-ai-agent-orchestration-com`
suffix made the wrong one look obviously right.

**What caught it:** filing the rerender returned `INSERT 0 1` when I expected 2. **The count was the
tell** — had both templates been placed, nothing would have been odd and I would have reported four
failures fixed when three were untouched.

⚠ **The three savings-estimator failures are UNTOUCHED**, and fixing them is now a **cross-site**
change (fundamentallyai.com shares that component), so it needs that lane told first — owner ruling
2026-07-29 §3 — not a unilateral repoint.

⚠ **And the diagnosis there is INCOMPLETE.** The fundamentallyai fork matches only the
`.result-value` defect on a regex census; the source of the other three live rules
(`.method-details summary`, the button id rule, `.cta-link`) is **not yet located** — possibly
different whitespace in that fork, possibly the site stylesheet. **Do not assume the orphan's rules
are the fork's rules.** That is the same mistake one layer down.

### 3. Numbering

Renumbered `626` → `636` mid-flight: another lane took 626 while I worked. ⚠ **My `625` had already
collided the same way** (`625_clear_audit_stamps…HOLD.sql`). The ledger keys on FILENAME so both are
safely recorded, but a bare number is ambiguous in `sql_for_agents/` exactly as it is in `/bugs_*/`.
Re-check the highest number immediately before writing, not when you start.

---

## 2026-09-02 — a week later: one fix landed, and a defect I had CLOSED was silently re-opened

Checked the ground after a week. `build-vs-buy-analyzer` is **0** (migration `636` landed and
propagated 08-26 12:22). `password-entropy` and `tool-llm-cost-calculator` are still **0** — `625`
holding. Then the interesting one.

### 1. `agent-complexity-estimator` was serving the calculator TWICE

Live, HTTP 200: **two** `<h2>Agent Architecture Complexity Estimator` headings, **two** estimator
UIs. `page_components` held two rows in the **same slot at the same position**:

| row | created | bytes | fieldsets | legends | **inputs** |
|---|---|---|---|---|---|
| `b2b7acbd` (incumbent) | 2026-04-09 | 22,732 | 4 | 4 | **12** |
| `9aa63fc0` (new) | **2026-08-26 14:48** | 19,964 | 1 | 1 | **1** |

⚠ **The byte counts would have waved it through** — −12% is inside any plausible size floor. **The
input count is what shows the loss.** A shrink guard on this class needs a structural axis, not a
length one.

⚠ **It silently re-opened a defect I had closed.** This page measured **0** firm failures on
08-25 after `625`; it measured **1** today, on the NEW component's button, which never received
`625`'s repoint. **A fix verified at the artefact was undone by a component that arrived
afterwards.** Nothing regressed my work — something was added beside it.

⚠ **And the estate's de-duplication rule was vacuous here.** The standing guidance is "act only
where `count(DISTINCT md5(content_data)) = 1`". Both rows carry `content_data = '{}'`, so both md5s
are `99914b93…` — the rule reported agreement it never established, exactly the shape the LANDMINE
warns about for NULL. `content_hash` is empty and cannot stand in. **The discriminator that worked
was the rendered artefact's structure.**

Repaired as migration `692` (removes the degraded PLACEMENT, leaves the component as evidence,
guarded to abort if the survivor drops below 12 inputs or loses `625`). Filed as `bugs_open/434`.
**Verified: 1 placement, 12 inputs, `owned`, 0 failures, one heading.**

### 2. ⚠ I filed 434 BEFORE grepping `/bugs_open/`, and the check changed the filing

`bugs_open/430`, filed the same day, is *"forking a tool component drops `js_content`"* —
`deploy_tool_action.go`'s fork-on-deploy INSERT, the obvious suspect for anything creating a
per-site tool component. **One column separates them:**

```
…-leopardessconsulting-co-uk-ai-agent-orchestration-com   forked_from IS NOT NULL  (incumbent)
…-ai-agent-orchestration-com                              forked_from IS NULL      (the new one)
```

**It was never forked — it was generated.** So 430's INSERT is not the producer, and copying fewer
columns cannot rewrite markup from 4 fieldsets to 1. ⚠ **Both have `js_content` length 0**, so that
signature does not discriminate and must not be used to link them.

The order was wrong and the outcome was luck: had I not grepped afterwards, 434 would have stood as
a probable duplicate of a same-day sibling in an adjacent mechanism, which is exactly what a
duplicate looks like from outside.

### 3. ⚠ I left an owned tool page on `rebuild_policy='generic'` for ~8 hours

`refresh_owned_page_chrome.sh` flips to `generic`, arms `trap restore_all EXIT INT TERM`, publishes,
then waits `24 × 5s`. **I ran it under a 2-minute foreground timeout. The harness killed it mid-wait
and the trap did not fire** — the page sat on `generic` from 05:56 until I checked at ~14:00.

**Nothing errored.** The migration before it had committed cleanly, the page still served, the tool
was still there. **The damage is a WINDOW, not a state you can see in the artefact**: the page was
simply eligible for the generic composition loop, which commits freshly-written HTML to the
deploying repo one step *before* `save_page_sections` refuses — the path that has already replaced
working calculators with prose.

Nothing hit it (verified: 12 inputs, 22,732 bytes, unchanged). **Restored to `owned` before doing
anything else**, then re-ran the script in the BACKGROUND so no harness timeout applies. LANDMINE
filed. **The check is one query and it belongs immediately after every run of that script,
successful or not.**

---

## 2026-09-03 — going upstream: the tool prompt never learned what the component prompt already teaches

Picked up from `HANDOFF_2026-09-02_continue_here.md` §7.1, which said per-page fixes were not
converging and the next useful move was upstream. It was, and the upstream turned out to be one
prompt line.

### 1. Re-enumerated first, as this lane's §9 requires

45 active pages (47 rows, 6 `owned`) — the handoff's denominator held for today. Site
`2a8ebf9c-20a2-4c39-b191-840b012371da`.

### 2. What the renderer already offers, and nobody asked for

`buildLegibleInkDefaults` (`platform/orchestration/actions/palette_specialised_slots.go`) emits
**paired ink tokens**, and its own header comment is emphatic about the distinction:

> `--color-<x>-text`   the ink that goes **ON** an `<x>` fill
> `--color-<x>-ink`    `<x>` ITSELF, made legible as an ink **on the page**
> *"TWO DIRECTIONS, AND CONFLATING THEM IS THE MISTAKE THIS COMMENT EXISTS TO STOP."*

Served on this site `[MEASURED 2026-09-03, https://ai-agent-orchestration.com/assets/css/styles.css]`:
`--color-primary-text: #ffffff`, `--color-primary-ink: #768eb2`, `--color-accent-ink: #F0A500`,
`--color-accent-text: #294155`. **All four already there.** Also there:
`--color-primary: #0D1117` and `--color-surface: #0D1117`, byte-identical, with
`--color-background: #080B10`.

### 3. The failing rules, read rather than inferred

| component | rule | ratio |
|---|---|---|
| `tool-model-approach-selector` | `.submit-btn { background: var(--color-primary); color: var(--color-background) }` | **1.04** |
| `tool-token-calculator` | `.stat-value { color: var(--color-primary) }` inside `.stat-box { background: var(--color-surface) }` | **1.00** |
| `tool-model-approach-selector` | `.error-msg { color: var(--color-primary); background: var(--color-surface) }` | **1.00** |

1.04 reproduces the handoff's reported button figure exactly, independently, from the CSS.

⚠ **`.error-msg` was reported by NO audit.** It only paints in an error state, and a render-time
probe measures the states the page actually renders. **So the audit's count is a floor, not a
census** — do not size this class from audit findings. This generalises to hover, `:focus`,
`:checked` and anything behind a JS branch.

⚠ **The `<TD> 1.14` the handoff recorded for `token-calculator` is not a rule I could find.** The
`.stat-value` rule measures 1.00. I could not reconcile 1.14 to a declaration and am not going to
pretend I did; the mechanism is the same either way, but the specific figure is unexplained.

### 4. The measurement that could have come out otherwise

If the producing prompt is the cause, the shape should sit in the tool population and not the
component one — `component-creator`'s prompt teaches
`--section-text: var(--color-primary-text, var(--color-background));` and lists
`--color-primary-text`, while `tool-generator`'s whole colour vocabulary is eight tokens with no
pairing rule.

```
[MEASURED 2026-09-03] active unforked content_components
                                        non-tool     tool
components                                   151      261
fills with --color-primary                    31      174
uses --color-primary-text                     59       31
primary fill inked with the page ground        0      148
```

**Zero and 148.** A 40/60 split would have refuted the prompt theory. It did not. This is the
single most load-bearing number of the day, and it is disconfirmable, which is why I ran it.

Fleet blast radius `[MEASURED 2026-09-03]`: **9 of 59** palettes score under 3.0:1 for that
pairing, **7 under 1.25:1** — loancash 1.00, aiao 1.04, agritec 1.11, dartsonline 1.11,
robot-hands 1.14, loanandmortgagecalculator 1.18, oufe 1.21. ⚠ **Stored palette, not served
stylesheet** — the overlay may serve something else and is allowed to, so that is a population to
check, not nine confirmed sites. Only aiao was confirmed at the artefact.

That table also **refreshes a stale census inside the code**: `warnUnusablePrimary` still says
*"Measured fleet-wide 2026-07-27: 4 of 31 palettes … below 1.25:1"* and names a membership that has
since changed. Grew by addition, read as current for five weeks — the owner's dating rule exactly.

### 5. The second half — the audit calls the correct repair "unknown"

`canonicalCSSTokens` (`component_validation.go`) carried `--hero-ink` and **none** of
`--color-primary-ink`, `--color-accent-ink`, `--color-cta-bg-ink`. So a component that opted into
the renderer's documented repair was reported as unknown drift. Warn-only —
`AuditTemplateTokens` *"NEVER rejects a template"* — so nothing was blocked. It is still worth
fixing because it points the signal backwards while adoption sits at **15 of 412**.

### 6. What I did NOT do, and why

**Did not touch the palette.** Making primary differ from surface is `RFC_059`, which the owner
**withdrew** on 2026-09-02: the overlay must keep full authority over its own colours. Anyone
planning to "stop the churn" here is re-proposing a withdrawn RFC. The paired-ink route was chosen
because it is palette-agnostic — it stays correct whatever the overlay picks.

### 7. Misstep: I called the diagnosis run stalled when it was running fine

Saw four bundles and no verdict, matched it against the LANDMINE *"a 090 run on a symbol in a file
over ~60KB returns bundles and NO verdict"* (my symptom named a 45KB file and a 9.5KB one), and
said the last bundle was "over an hour ago". **Wrong.** I compared the DB's UTC timestamps against
my own assumed wall-clock. `SELECT now()-created_at` showed the newest bundle was **35 seconds
old**, arriving every ~4.5 minutes. The run was healthy.

**The check is one column.** Never age a DB timestamp by subtracting it from a clock you did not
read out of the same database — ask the database for the age. The landmine was real and my file
sizes did sit near the budget, which is exactly why a plausible mechanism deserved a measurement
rather than a match.

### 8. Shipped

- `bugs_open/458` — the filing, with the nine-palette table and the verification recipe.
- `docs/agent_docs/sql_for_agents/732` (+ ROLLBACK) — appends the pairing rule to `tool-generator`
  and `tool-improver`, anchored verbatim, **aborts if either anchor moved**. The abort was
  **induced** (exit 3, no COMMIT), not merely written.
- `component_validation.go` — the three ink companions plus `--color-accent-text` made canonical.
- `component_validation_ink_lockstep_test.go` — derives its expectation **from the emitter**, so
  the next companion fails a test instead of silently becoming unknown drift. **Mutation-proven:**
  removing `--color-primary-ink` fails it; restoring passes.
- Commit `0325ddebb`, `Council-Submitted: 0fd2ca6b-f400-4452-8cac-25399f7d55ea`.

⚠ **None of this repairs a single existing page.** 148 tool components and this lane's four live
failures persist until a regeneration pass. A still-failing page is NOT evidence this did not land.

### 9. The council rounds — two of the three found defects in the shipped work, not the paperwork

**Round 1 → REVISE (gated by `editquality`).** Five seats objected to the *same* thing: the
tool-improver half of the migration was "a comment, not code". **They were right about the sketch
and the sketch was wrong** — the file has always carried both halves; I elided one behind
`-- the same shape again for tool-improver` to keep the submission short. `debug_historian`'s HIGH
was the same shape: my `replace()` anchor written `'..., var(--color-border)'` was shorthand, and it
warned the ellipsis would silently no-op if literal. **Reviewers judge the sketch; it is the only
view of the code they get, and I gave them a false one.** Round 2's sketches are generated FROM THE
FILES, which is the only version of this that cannot drift.

⚠ **The round-1 objection I did not expect found a red HEAD.** `guardian` asked whether
`canonicalCSSTokens` had any consumer beyond the audit function. I had checked the *function* and
mistaken that for checking the *map*. One `grep -rn` found `rendererGuaranteedTokens` in
`check_stylesheet_gutted.go`, kept in lockstep by a parity test that parses the `actions` source —
and **my round-1 commit had left it red for ~2 hours**. `go build ./platform/...` passes on a
failing test and `go test` on the edited package misses a break one package over.

⚠⚠ **And the fix the test named would have been worse than the bug.** It demands set equality, so
the obvious move is to add all four tokens to the other list. That puts `--color-cta-bg-ink` — which
`buildLegibleInkDefaults` emits only when `solidCTAFill` is non-empty, present in **1 of 7** served
stylesheets against **7 of 7** for the other three — into a **live, registered, severity-high**
check, filing 6 of those 7 sites as gutted. **A green test would have certified a fleet-wide
false-positive generator.** Measuring the four at the artefact rather than satisfying the test is
what separated them. In `016b` §9.

**Round 2 → REVISE again**, and both gating objections were real:

- `editquality` + `guardian` HIGH: my edit named ONE file and changed TWO. The irony is exact — the
  round's own thesis is *"I broke a parity test by not touching the paired file"*, and I shipped the
  fix for that as a single-file edit needing two. Split.
- **`bug_historian`, and this is the best objection anyone made:** after this lands, a
  tool-generator run that *ignores* the new prompt sentence produces the identical defect with
  **zero detection anywhere**. `check_stylesheet_gutted` sees a token's ABSENCE; the wrong pairing
  uses valid, PRESENT vars. I had conceded "a prompt rule is taught, not enforced" in the submission
  **and stopped at the concession.** Built in round 3: `scripts/audit-fill-ink-pairing.sh`
  (**STY-062**). First run — **25 findings in 7 days across 5 sites**, 6 of them same-day.

`reuse_agent` asked why I added a third parity mechanism instead of extending the existing one.
Answer: `actions` imports `discovery_checks` (75 refs), so the reverse is an import cycle — which is
*why* the existing test parses source as text — and `buildLegibleInkDefaults` is unexported. Six
`*_lockstep_test.go` files already sit in `actions` in this direction. **Structurally impossible, and
I should have said so unprompted rather than leaving it to be asked.**

`guardian` asked for the expected fleet-wide finding count from arming a live detector. **Zero** —
13 of 13 genuinely-served stylesheets carry all three. ⚠ **Measuring it caught a fault in my own
instrument:** a 14th domain read as missing all three, and it was a **404** at the conventional
stylesheet path — I was grepping an error page. The check itself refuses on any non-2xx fetch (its
comment cites 63 false findings that refusal avoided), so it could not file there anyway.

**Standing tally for this lane: three council rounds, four real defects found** (a hidden half, a
red HEAD, a two-file edit, and a missing detector) — none of them cosmetic, and the last two changed
what shipped.
