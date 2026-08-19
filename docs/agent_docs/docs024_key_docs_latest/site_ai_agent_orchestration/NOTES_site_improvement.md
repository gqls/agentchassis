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
