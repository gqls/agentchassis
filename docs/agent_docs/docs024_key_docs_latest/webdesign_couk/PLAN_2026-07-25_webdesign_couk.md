# PLAN — webdesign.co.uk: merge two static sites into one chassis-managed site

**Started 2026-07-25.** Owner directive: adopt `website-design.com` and `websitedesign.com`
into `webdesign.co.uk`, "very closely for the design", carrying all but one feature —
websitedesign.com's client-side LLM builder is skipped.

## What we are building

One site, `webdesign.co.uk`, chassis-managed from the start, holding the union of two
hand-built static sites that live in `~/projects/sites`:

| source | content | design |
|---|---|---|
| `website-design.com` | 86 pages — 55 client-side tools, 23 articles, global fuzzy search | Swiss high-contrast: `#0055ff` on white, `#111` ink, system fonts, fluid `clamp()` tokens |
| `websitedesign.com` | 21 pages — 10 tools, 10 guides | warm minimalist on the homepage ONLY; all 20 sub-pages carry a legacy dark-terminal skin |

Net after curation: **~64 tools + ~31 learn pages + about + two indexes + home**.

## Decisions and their reasons

### D1 — Design target: warm minimalist, everywhere (owner, 2026-07-25)
Chosen over the Swiss system and over a blend. The palette is `websitedesign.com/css/main.css`:

```
primary      #5c6b5d   (hover #4a574b)      background  #f9f8f6
secondary    #8a9a86                        surface     #ffffff
accent       #d4a373                        text        #2b2b2b
border       #edece9   (hover #dbd9d4)      text_muted  #717171
radius 12px · shadows rgba(43,43,43,0.04/0.06/0.08) · Inter + Fira Code · LIGHT ONLY
```

Consequence, accepted knowingly: this is the *smaller* site's design applied to the
*larger* site's content, so ~86 pages get reskinned rather than ~20. That is what the
compat-stylesheet decision (D5) exists to make affordable.

### D2 — Chassis-managed from the start (owner, 2026-07-25)
Chosen over a static merge in the sites repo. So the platform owns chrome, nav, CSS and
the homepage; the ~97 content pages are **owned pages** (`rebuild_policy='owned'`) whose
HTML we author and the chassis assembles and publishes.

### D3 — Onboarding via 082 FRESH mode with a parked cascade
Not `--from` adopt: adoption seeds identity and `design_reference` from ONE crawled site
to replicate it faithfully. We are merging two sites into a new brand with an
owner-fixed palette, so adoption's fidelity machinery would encode a provenance that is
not true. Not hand-SQL either: spec shapes evolve with the agents, and hand-authoring
identity/classification/strategy/briefing exercises none of the pipeline. Hand-SQL stays
the documented fallback if the parked cascade misbehaves twice.

Control comes from **parking, not racing**: a watcher flips every new `pipeline='build'`
work item on the site to `status='blocked'` unless allowlisted, and legs are released one
at a time. `blocked` is a respected status — it has its own index
(`idx_work_items_blocked`) and `CompleteWorkItemAction` refuses to overwrite it
(`load_work_item_actions.go:808`).

> **CORRECTED 2026-07-25 (same day, before any of it ran):** this paragraph first
> claimed a blocked item is *"inert; excluded from the `idx_swi_dedup` live set"*.
> Half right. It is inert — the dispatch gate takes only `triaged`/`approved` — but
> `blocked` is **absent from that index's exclusion list**, so a parked item still
> **holds its dedup slot**. Benign here (it stops a duplicate appearing behind a
> park) but a forgotten park blocks re-emission of that `item_key`. Caught by
> reading `\d site_work_items` rather than trusting the paraphrase.

> The site lock is NOT a park — verified 2026-07-25 by reading both SQL texts, not
> just the first one. `build-pipeline-trigger`'s `pre_query` counts only sites with
> `locked_at IS NULL`, so the lock decides whether the trigger **fires**. But its
> `find_dispatchable_site` step then runs
> `SELECT DISTINCT ON (wi.site_id) … ORDER BY wi.site_id, wi.priority ASC LIMIT 1`
> with **no `locked_at` filter**, so a trigger firing for another site's backlog can
> select ours. `LoadWorkItemsAction` has no lock check either. The watcher is the
> real gate; the lock is belt.

### D4 — Homepage is the only chassis-composed page
Tools index, learn index and about are **generated from the port manifest**, not composed
by an LLM. A deterministic listing of 64 tools across 6 categories with disambiguating
subtitles is a data-rendering job, and the manifest already holds the data. The planner
therefore runs exactly once, scoped by the mission brief to home + nav.

### D5 — A compat stylesheet, not a mass `var()` rewrite
Ported fragments keep their page-local `<style>` blocks, which are written against two
different token vocabularies. `assets/css/port-compat.css` aliases Site A's Swiss token
names onto warm values (and keeps its neutral `clamp()` space/type scales), pins Site B's
warm names, and adds port tokens (`--code-bg`, `--code-accent`, `--callout-*`,
`--ok/--warn/--bad`, `--danger`). Only literal colours get swept. This turns thousands of
`var()` references into one file.

Dark mode dies **by omission**: Site A's `prefers-color-scheme` block lived in
`global.css`, which is not ported. A verify gate enforces that no fragment reintroduces it.

### D6 — Curation (owner, 2026-07-25)
- **Skip** `tools/local-builder/` (WebLLM/WebGPU, ~800MB model download). With it go
  `guides/local-ai.html` and `guides/hosting-economics.html` — the latter is a
  byte-identical copy of the former (verified by md5), so the hosting-economics content
  was never actually written. Three prose mentions elsewhere get reworded.
- **Link the orphans in**: 4 finished-but-unlinked tool dirs (`clip-path`,
  `cubic-bezier`, `magic-outliner`, `noise-generator`) and 10 unlinked Learn articles.
- **Fix** `vibe-equalizer`, which is dead as shipped: it `<script src>`s
  `../../js/state.js`, a file that does not exist.
- **Drop** stale variants (`animated-favicon/index_orig.html`,
  `magic-outliner/index-convex-hull.html`) and `showcase/deconstruction-template.html`
  — an uninstantiated template (`<title>Deconstructed: [Site Name]</title>`), zero
  inbound links; publishing it would ship placeholder copy.

### D7 — No duplicate tools (verified 2026-07-25)
Checked before committing to the union: all 64 slugs are unique, and the near-miss pairs
are genuinely different tools. Recorded because "surely some of these are the same tool"
is the obvious objection to a 64-tool merge:

| pair | why both stay |
|---|---|
| `seo-schema` vs `seo-injector` | Article/Product/FAQ JSON-LD snippet vs LocalBusiness JSON-LD wrapped in an AI-builder prompt |
| `text-sanitizer` vs `insight-injector` | cleans AI text after the fact vs constrains generation up front |
| `css-variables` vs `vibe-equalizer` | token/theme-file generator on a 1.5 ratio scale vs mood sliders driving a live preview |
| `shadow-stacker` vs `smooth-shadow` | manual per-layer editor (X/Y/blur/spread/colour) vs parametric alpha/distance/sharpness → auto-eased layers |
| 5 prompt tools | image prompts / prompt trees / permutation matrix / sequential site-build deck / refactor prompts |

The overlap is real enough to confuse a visitor, so the tools index carries
disambiguating subtitles on each of these.

### D8 — Search
The header chrome fork renders the search pill; the carried 63-line `search.js` engine
ships as that fork's `js_content` (publishes to `/tools/assets/site-header.js`; the
template carries the `<script src>` itself — the route fixed in bugs 018+041, live
v1.0.1146 and verified on idea.uk). `search.json` is regenerated from the manifest, never
merged by hand. **One delivery path only** — the `js_snippets` bundle is not used for this.

### D9 — Fonts via Google Fonts css2 link in the head fork
Exactly what the source does; zero new machinery. Self-hosted woff2 under `/assets/fonts/`
is the better end state (one fewer third party, UK GDPR posture) and is a one-`<link>`
change later. **[OWNER-DECISION-OPEN]**

## Sequence

| phase | what | blocks on |
|---|---|---|
| 0 | this dir | — |
| 1 | `cmd/webdesignport transform` + port data files; QA offline | nothing (repo-local) |
| 2 | mission brief, watcher, 082 submit, airlock the cascade legs | — |
| 3 | `design_intent` pin | classifier leg done |
| 4 | plan review, composition, CSS | pin verified |
| 5 | chrome forks (header/footer/head) | composition |
| 6 | seeds, static assets, `webdesignport import` | site row + chrome |
| 7 | `rerender-pages` assemble, then home | chrome + CSS final |
| 8 | verify, unlock, steady state | everything |

Phase 1 is ~80% of the work and runs parallel with 2–5.

## Corrections

*(corrections are appended above, dated, never edited away)*

## Corrections

> **CORRECTED 2026-07-26 — "blocked on dispatch" was wrong.** At the end of
> session 1 I recorded that the 98 page assemblies would not dispatch and called
> it the `bugs_open/003` spawn-loss class. There was no bug. First claim came
> 20m40s after the items were created (documented latency); all 98 completed in
> 3h28m at ~2.1 min each, which is what single-flight-per-site costs. I gave up
> 8 minutes early, and my "imagery is completing concurrently" evidence was the
> tail of work claimed *before* the page items existed. Full account in NOTES and
> `WRONG_CALLS.md`.
>
> **Planning consequence:** a site-wide rerender of ~100 pages is a **3.5-hour**
> operation that shows nothing for the first 20 minutes. Phase 7 should have said
> so. Anyone repeating this should queue it and walk away.

> **CORRECTED 2026-07-26 — D1's claim that the pin protects the design is too
> broad.** The pin governs colour *values* and it held flawlessly: every colour
> in the committed `styles.css` is a pinned one. It does **not** govern which
> *component* the planner selects, and a component can carry darkness of its own
> — the generic `hero` paints `rgba(0,0,0,0.5)` over its background image
> regardless of palette. `design_intent.avoid` is prose the planner's component
> choice never reads.
>
> So the honest statement of what a pin buys is: **right palette, guaranteed;
> right furniture, not guaranteed.** Reviewing the planner's *composition* (which
> the airlock made possible) is a separate and equally necessary step — and it is
> the step I did not do, because I checked the page LIST against the brief and
> not the SECTION list. Fixed by `SQL_p6` (per-site two-column hero).

> **CORRECTED 2026-07-26 — the tool count.** The brief and D7 said "~64 tools".
> The real figure is **63** (55 from website-design.com + 8 from
> websitedesign.com, the 9th being the skipped LLM builder). The wrong number
> originated in this plan, reached eight live specs, and was corrected by
> `SQL_p4` before the planner could write it into page copy. Counts are now
> substituted from the catalogue (`{{TOOL_COUNT}}`) rather than typed.
