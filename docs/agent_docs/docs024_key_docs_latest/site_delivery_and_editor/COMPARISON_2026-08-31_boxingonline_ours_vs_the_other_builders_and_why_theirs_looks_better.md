# COMPARISON 2026-08-31 — boxingonline.com: ours vs the page another builder produced, and why theirs looks better

**Occasion:** the owner had the same brief built by a different system (a single-page AI site
builder — the pasted HTML carries `ai-site-builder-preview` instrumentation) and sent us the
full source. His ruling:

> "their design is much more vibrant, they have player profiles - we should. and much more too.
> **We want what we have and what they have, not instead**, but our quality can and must be
> improved."

**The verdict this file argues, in one line:** their advantage is almost entirely **presentation
and content-type inventory** — and *we specified both, in writing, and built neither*. Their
content half is a mock-up we must not copy. So "what we have and what they have" is achievable
without inventing anything, and most of the gap is a configuration and planning failure rather
than a capability we lack.

All our-side figures `[MEASURED 2026-08-31]`, queries inline. Their-side facts are read
directly off the HTML he pasted.

---

## 1. The design gap — and the sentence in our own spec that asked for their design

Their palette, from their `:root`:
```
--black:#0c0c0f   --charcoal:#16161b   --panel:#1c1c22
--red:#e0263c     --gold:#f2b21a       --off:#f4f1ea
headings: 'Arial Black', Impact — uppercase, letter-spaced, line-height 1.05
```
A near-black page, red and gold accents, fight-poster headline type.

**Our `design_intent` spec for this site says, in prose:**

> `colour_mood`: "**Deep red and near-black as the dominant palette** — boxing is a sport with
> theatrical, high-contrast visual heritage… **Gold accent for highlights mirrors championship
> belts** and broadcast graphics. This is a deliberate bold-creative choice over light-default
> because the vertical is combat sports… **Light would feel too soft for the subject matter.**"
>
> `avoid[0]`: "**Pastel or muted colour schemes — too soft for the sport**"
>
> `typography_mood`: "Headings in a wide, confident condensed sans-serif — **the kind of
> letterform used on fight posters**."
>
> `style_direction`: "bold-creative"

That is a description of the page he was sent, written by us, before we built ours.

**What we actually serve** (`curl https://boxingonline.ugg2.com/assets/css/styles.css`):
```
--color-background: #F7F3EE      <- warm off-white
--color-surface:    #FFFFFF
--color-text:       #1A0A0A
--color-secondary:  #C0392B
--color-accent:     #D4A017
--color-header-bg:  #0a0a0a      (dark chrome only)
--font-heading: 'Barlow Condensed', 'Arial Narrow', sans-serif
```
A light page with dark chrome. The **exact thing its own spec's `avoid` list forbids.**

> **CORRECTED 2026-08-31, same day, by the session that wrote it — and the corrected version
> is the stronger finding.** My quotation of `colour_mood` above uses an ellipsis that drops a
> sentence arguing the OTHER WAY. The full field reads: *"Deep red and near-black as the dominant
> palette… **A warm off-white background keeps it readable without feeling clinical.** Gold accent
> for highlights… This is a deliberate bold-creative choice over light-default… Light would feel
> too soft for the subject matter."*
>
> So the brief does not simply say "dark". **It says both, in one field**: near-black dominant AND
> a warm off-white background AND light-would-be-too-soft. It contradicts itself, and my elision
> hid the half that the build actually followed. What caught it: reading the whole field again
> while tracing the palette provenance, rather than trusting my own excerpt.
>
> **The finding survives and sharpens.** It is not "the prose said dark and the values said
> light". It is: **the design brief is self-contradictory, and `palette.reference_values`
> silently resolved the contradiction toward light — because it is the only half anything reads.**
> A brief that cannot be satisfied both ways needs the conflict caught where it is written, not
> resolved by whichever key happens to be load-bearing.

**Where the contradiction lives — this is the finding.** The same spec row carries both halves:

| `design_intent` key | says | load-bearing? |
|---|---|---|
| `colour_mood` (prose) | near-black dominant, light too soft | **no** — prose |
| `palette.reference_values` | `background #F7F3F0`, `surface #FFFFFF` | **yes** — this is what renders |

`palette.reference_values` is the documented pinning mechanism (it exists precisely because
`generic_theme` misfires — see the colour-churn landmine). Here it was populated with a **light**
palette that contradicts the mood prose sitting six lines above it in the same row. Nothing
compares the two, so nothing noticed.

**PROVENANCE CONFIRMED AT THE ARTEFACT `[MEASURED 2026-08-31]`**, because "two documents agree
with me" is not evidence of a causal chain. Three links, each checked:

1. **The picker's cascade puts `reference_values` second, behind only a human-set mission
   palette** — `resolve_composition_pallette_action.go:16-20`: *"1. mission.preferred_palette
   (human pre-specified) · 2. design_intent.palette.reference_values (semantic brief)"*. This
   site has no mission palette, so link 2 is the live one.
2. **The site's palette row is BYTE-IDENTICAL to `reference_values`** — not merely similar:
   ```sql
   SELECT name, origin, colours FROM palettes WHERE source_domain='boxingonline.com';
   -- palette-boxingonline-com | adopted |
   -- {"text":"#1A0A0A","accent":"#D4A017","border":"#E0D5D0","primary":"#1A0A0A",
   --  "surface":"#FFFFFF","secondary":"#C0392B","background":"#F7F3F0","text_muted":"#6B5B55"}
   ```
   All eight slots match `design_intent.palette.reference_values` exactly.
3. **The served stylesheet carries those values through** (`--color-background: #F7F3EE`, a small
   render-time adjustment of `#F7F3F0`; `--color-secondary: #C0392B` and `--color-accent:
   #D4A017` verbatim).

So changing `reference_values` is the right lever, and this is safe to act on rather than infer.

**One thing that makes it a CLASS question, not a site question.** The library contains a palette
literally named `boxing` — the cascade's step-4 fallback for a high-energy layout — and **its
background is `#fafaf9`, light too.** Its chrome is right (`primary #dc2626`, `header_bg
#000000`, `footer_bg #1c1917`, `accent #fbbf24`, a red gradient CTA) but the page body is still
off-white. **So there was no path through this system that produced the near-black page the brief
also asked for.** Picking the library palette instead would not have saved us.

`[UNVERIFIED]` whether other sites show the same self-contradicting `colour_mood`, and whether
any site in the estate serves a genuinely dark-dominant page. That census is cheap, it is the
difference between a one-row fix and a class fix, and it should run before anyone hand-edits
this site.

**Second design point, and it is the cheap one:** their vibrancy costs nothing in assets.
**Their page contains ZERO `<img>` elements.** Every "photo" is a CSS gradient box with a text
label — `Ringside Photo`, `Weigh-In`, `Press Room`, `Training Camp`. The profile "photos" are
gradient panels with the fighter's initials at 15% opacity. So on imagery they are not ahead of
us at all; they are ahead on **composition, contrast and type**. That is reachable with a palette
row and a component pass, and it does not wait on image generation.

## 2. The content-type gap — and the spec that named every missing type

He is right that they have player profiles and we do not. What he could not know is that
**our strategy spec specified them, with reasoning, and named the research behind them.**

`site_specs.aspect='strategy' -> recommended_page_types` for this site names **seven** types.
Verbatim on the two he noticed:

> **`entity-page`** — "Fighter profile pages — one per fighter mentioned across articles —
> aggregate all site content about that boxer, display their current record, and link to
> upcoming fights on the calendar. These pages capture long-tail search traffic ('Canelo Alvarez
> next fight', 'Tyson Fury boxing news')… The vertical_landscape research confirms all three
> exemplars use this pattern as a structural SEO layer."
>
> **`entity-directory`** — "An event directory — one page per major upcoming fight — gives each
> bout a permanent URL with full details: fighters, date, venue, broadcast, undercard, and a
> brief preview. This serves the high-intent 'Fury vs Usyk date and time' search…"

**What the planner then built:**
```sql
SELECT p.name, p.role FROM site_plan_pages p JOIN site_plans sp ON sp.id=p.plan_id
 WHERE sp.site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152';
```
→ `index` landing · `about` content · `contact` content · `articles-index` section-index ·
`tool-fight-calendar` tool · `article` blog-post (zero sections, never built).

**Zero `entity-page`. Zero `entity-directory`.** The two types that would have produced fighter
profiles and event pages — the two things he singled out — are the two the plan dropped.

**And the roles are not hypothetical.** Fleet-wide, `site_plan_pages` carries `entity-page` on
**30 pages across 11 plans** and `entity-directory` on **6 pages across 5 plans**. The
vocabulary exists, is admitted, and is in production elsewhere.

> **Filed to the diagnosis loop rather than asserted here:** run correlation
> `d6d350ec-e16b-4792-9282-ca5155369791` asks why the planner omits roles the strategy names.
> **Do not quote a mechanism for this until that reports.** What is measured is the mismatch,
> not its cause.

Also missing on our side and present on theirs, all of them named in our own research or
strategy:

| they have | our spec that asked for it |
|---|---|
| Boxer profile cards with record, division, nickname, country | `strategy.entity-page`; `vertical_landscape.adopt` — "Fighter tag/profile pages as a structural content layer" |
| Editorial section with **named bylines** ("J. Alvarado, Senior Columnist") | `vertical_landscape.adopt` — "Named site voice or columnist identity — even a single consistent byline creates loyalty that anonymous aggregation cannot" |
| News grid with one large feature card + four smaller | `design_intent.layout_preference` — "Magazine grid — editorial card grid… full-width hero banner for the latest big fight" |
| Calendar as fixture rows: date chip, badge, venue, local time, CTA | `design_intent.layout_preference` — "dedicated calendar section with clear event rows. The calendar should feel like a **proper fixture list, not a generic table**" |
| Newsletter capture ("Never Miss a Bell") | `strategy.growth_path` — "building a bookmark habit and potential email list" |
| Per-event detail links | `strategy.entity-directory` |

**Every single row of that table was specified by us and not built.** That is the honest shape of
the gap: not a missing capability, a missing execution.

## 3. What we must NOT copy from it

Their page is a **mock-up wearing a finished site's clothes**, and on a boxing site the
difference is not cosmetic — our own research names factual precision as the vertical's number
one trust mechanism and stale dates as its number one destroyer.

Objectively verifiable on their page:

1. **An unsubstituted template placeholder shipped to the rendered page.** The contact form is
   `action="mailto:OWNER_EMAIL_ADDRESS"` — the literal variable name. (Worth noting given the
   owner's other complaint tonight: their builder was going to publish an owner address too, and
   only a bug stopped it.)
2. **A placeholder image service in the social card:** `og:image` and `twitter:image` both point
   at `https://placehold.co/1200x630`.
3. **Every date is 2024** — "Oct 6, 2024", "© 2024" — two years stale as of today.
4. **Every "view all" and "details" link is `href="#"`.** There is no second page. The nav is
   five in-page anchors. Depth: one page.
5. **No provenance for any fact on the page.** Fighters (Marcus Reyes 28-0-0, Kenji Okada 22-1-0,
   Amara Cole 17-0-0, Diego Silva 31-2-1), fixtures (T-Mobile Arena Oct 18, Ariake Arena Oct 25),
   bylines and quotes carry no source. `[UNVERIFIED]` whether any of it corresponds to real
   people or events; what *is* certain is that the page cites nothing, and our platform would
   refuse to publish records and fixtures with no `evidence_base` entry — correctly.

**So the comparison is not "they built a better site".** They built a better-looking *design
comp*, populated with invented content, one page deep. We built a thinner-looking real site with
real interactive tools, real pages, and a claims-gating pipeline — and then failed to apply our
own design and content-type decisions to it.

## 4. What we have that they do not

Worth stating so the improvement work does not throw it away:

- **Working interactive tools** (four of them) rather than static markup — though see the
  standing complaint that they ask the reader for the data.
- **Real pages at real URLs**, a sitemap, canonicals, and a publish pipeline, versus five
  anchors on one document.
- **Real generated image assets** — logo, four heroes, three icons, actually produced and
  deployed, where theirs has none.
- **A claims and evidence pipeline** that is the reason we do not have four invented fighter
  records on the homepage.
- **Accessibility already present** on both sides (they have a skip link and `aria-labelledby`;
  so do we) — no gap either way.

## 5. The concrete list, in the order I would do it

1. **Fix `palette.reference_values` to match the spec's own `colour_mood`** (near-black dominant,
   red + gold accents). One row, live immediately, no image build. Biggest visible change per
   unit of work, and it makes the site obey a decision we already made. Census the fleet for the
   same prose-vs-values contradiction before or alongside.
2. **Emit the two missing page roles** — `entity-page` (fighter profiles) and `entity-directory`
   (event pages). Both are live roles elsewhere; the strategy already says what goes on them.
   **Gated on the diagnosis run above** so we fix the planner rather than hand-adding pages.
3. **Magazine-grid the article listing and give the calendar fixture rows** — both are already
   written in `layout_preference`; this is a component/composition job, which is the
   editorial_design_uplift lane's territory.
4. **Named byline on editorial.** Our research says it is a vertical standard; it costs a field.
5. **Newsletter capture**, for the bookmark habit the strategy names.
6. **In-body imagery and infographics** — the standing gap, `inline_guide_imagery` and the
   1-row-fleet-wide infographic finding.
7. **Give the tools real data** so they answer rather than ask.

Items 1–5 are all *executing decisions we already recorded*. Only 6 and 7 are new build.

---

**Cross-references:** `OWNER_REVIEW_2026-08-31_boxingonline_what_he_found_and_what_each_finding_actually_is.md`
(this directory) · `bugs_open/419` (zero-section blog-post page; mechanism deliberately
unsettled) · diagnosis run `d6d350ec-e16b-4792-9282-ca5155369791` (planner drops recommended
roles) · the five lane CONTRIBs committed as `a5fa9909a`.
