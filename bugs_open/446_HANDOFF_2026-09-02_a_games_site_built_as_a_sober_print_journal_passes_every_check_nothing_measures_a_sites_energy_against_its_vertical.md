# 446 — a GAMES site was built as a sober print journal with no games, no game imagery and no articles — and it passed every check, because nothing measures a site's energy against its vertical

**Filed 2026-09-02 ~20:45Z** by the `gamedesign.uk` lane, on the owner's instruction after his
review of the freshly-live site: *"It is a game design site — why isn't it full of games and images
and excitement? Please add that to the errors list, it is a major error."* **Status: OPEN, owned by
this lane for the instance; the CLASS (§5) is unowned.** Sibling filings the same evening:
`bugs_open/444` (brief-echo listing pages — this site is a fourth-site instance, §3.2),
`bugs_open/114` (imagery deployed and never referenced — adjacent, §3.4).

⚠ Number collision risk — 445 was the highest at filing; resolve by slug.

## 1. The one-paragraph version

gamedesign.uk went live at 18:00Z tonight through the framework, in the direction its brief asked
for — and the brief, the imagery guide and the plan all asked for the WRONG THING for the vertical.
The served site has one image (the logo) and three pencil-and-paper still lifes under a 60% black
overlay; an articles index with zero articles that instead describes what the articles would be
like, including a "What they avoid" list copied from the brief's don'ts; a hero on that page that
is a gradient over a 404; and copy that examines game design like a legal journal. **Every check
passed.** No detector in the estate asks whether a site is as alive as its subject demands. A
games site rendered as stationery is indistinguishable, to the pipeline, from a games site.

## 2. The owner's words (verbatim, 2026-09-02 ~20:30Z)

> this site needs to be seen again by the checkers. please run the improvement loop over it. it
> suffers from the same problems that designblog.co.uk etc suffered with. please correspond with
> that blog to determine the best fixes. We need to change the design and copy. hero images are
> missing e.g. articles/index.html that same page shows an explanation of the brief and so on.
> It is a game design site why isn't it full of games and images and excitement -please add that
> to the errors list it is a major error

## 3. What is actually wrong, measured at the served site 2026-09-02 ~20:20Z (cache-busted, control 404)

### 3.1 The images say "no game imagery" — because the spec told them to

Three hero images were generated, stored and deployed (`needs_imagery` ×3 complete, commits
`653949ef`/`c843cab3`/`bcd66e5e`, all three files 200 at ~150 KB). Their prompts, verbatim from
the items' results:

- home: *"An overhead view of a designer's working desk — printed document pages with diagrams
  and annotations spread across a warm wood surface, a pencil resting on a ruled margin…"*
- about: *"A close detail of a hand-drawn system diagram on off-white paper — boxes, arrows,
  margin notes in pencil…"*
- contact: *"A minimal still life of a folded document and a pencil on a warm linen surface…
  **no game imagery**. The register of a correspondence page for a print journal."*

Source: this lane's own `imagery_style_guide` (`SEED_2026-09-02_gamedesign_uk_site_and_specs.sql`),
whose `avoid` list reads *"screenshots or renders of real or invented games; game characters,
weapons, loot, coins, gems, health bars, controllers; cartoon or anime styling; saturated
primaries, neon…"* and whose `medium` is *"ink-and-wash… diagrams of process and flow… rather than
pictures of games"*. The pipeline obeyed exactly. **The filing lane wrote the ban; the owner has
ruled it wrong for the vertical.** That is the instance's root cause and it is a SPEC error, not a
pipeline error — recorded here so the class (§5) is not confused with it.

### 3.2 The articles page has no articles, and describes the brief instead (`bugs_open/444`, fourth mechanism)

`/articles/index.html` — 200, 8,396 B, 2,148 chars of main text, **zero articles**. Served body:
*"Articles written for readers who already know what a GDD is… What the pieces do… What they
avoid: Listicles framed as universal rules for every studio · Tool tutorials or software
comparisons · Writing that would read the same on a general product management blog…"* — the
mission brief's constraints rendered as page copy, including a **negative-identity section** of the
kind the owner banned on 2026-09-02.

Mechanism, distinct from 444's feed/directory/glossary trio: the site plan created **one** article
page (`/blog/article.html`) with **zero sections**; the build handler parked it honestly
(`mark_no_ready_sections`); the owner ruled it cancelled. So the content type the index lists has
**no producer because no content pages exist** — and the content writer, given an index with
nothing to index, wrote about the index. Plan-time validation (444's candidate 1) would have
caught "a section-index whose section contains zero content pages" before a word was written.

### 3.3 The one page the owner named has a hero over a 404

`/articles/index.html`'s hero: `style="background-image: linear-gradient(rgba(0,0,0,0.5),
rgba(0,0,0,0.6)), url('/assets/images/hero.jpg')"` — and `/assets/images/hero.jpg` is **404** on
this domain. `site_plan_imagery` for the site holds `hero_home`, `hero_about`, `hero_contact`,
`logo` — **no site-scope hero and none for the section-index page**, so the template's default
path had nothing behind it. **Control that stops this being a fleet claim:** the same default path
is referenced by deployed components on at least eight other sites (webdesign.co.uk 51,
finetuning.uk 30, mortgagecalculator.co.uk 27, ai-agent-orchestration.com 26, gamesdesign.co.uk
21, relojistas.com 20, lendzy.co.uk 19, gaswholesalers.com 16 — `[MEASURED 2026-09-02]`) and on
**every one of the six probed the file serves 200**. gamedesign.uk is the anomaly: a planner run
that requested no site-level hero. Small detector gap alongside: `check_image_url_404.go` inspects
`<img src>` only (three shapes, all `<img>`), so a CSS `url()` that 404s is invisible to it.

### 3.4 Even the working heroes are muted by the template

Home/about/contact reference their generated images correctly — as CSS backgrounds under
`linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6))` with `--hero-ink: #fff`. A 50–60% black wash
over every hero on a warm-paper site, with white text. This is the "same as every other site"
grammar the owner named on designblog the same evening (their `CRITIQUE_2026-09-02`): the hero
component's overlay is fixed, not palette-aware, so a light editorial palette gets a dark hero
regardless. `bugs_open/114` is adjacent (imagery deployed and never referenced); this is the
next case along — referenced, and flattened.

### 3.5 The copy is a legal journal

Homepage: *"Game design, examined the way it is actually practised in studios. Writing for design
leads, principal designers…"* — accurate to the brief, and to the brief's fault: the brief asked
for *"short sentences… do not call anything powerful, seamless…"* and got an essay about the site.
Zero interactive elements, zero games, zero references to any actual game. The sibling
`gamesdesign.co.uk` has 11 playable tools and 4 games; the practice seat has prose about process.

## 4. What passed

`site_unreachable` (clears), `page_content_divergence` (bytes match the commit), the CTA resolver
(gated the dead buttons — correctly), the claims layer (no banned pattern fired — the copy is
scrupulously unclaimable), `audit-rowless-serving-domains` (row + pages → OK), `image_url_404`
(no `<img>` fails). **A site with nothing to say, saying it carefully, is green.**

## 5. The CLASS — the detector gap the owner is naming

**Nothing in the estate measures a site's ENERGY against its VERTICAL.** The checks measure
presence (sections exist, images resolve, claims are registered, links work) and never
appropriateness (does a games site look like games; does a food site look like food). The
`experience-promise` family (live 2026-09-02) is the nearest: it checks that a page delivers what
its own headings promise — which this site does, immaculately, because it promises restraint.
A vertical-appropriateness check would need a referent per vertical (what "alive" means for
games vs. for restructuring finance — `oufe.com` is CORRECTLY sober) and would have to read the
served page, not the spec. `[ABSENCE CLAIM — basis: grep of `discovery_checks/` for
vertical/energy/vivid/appropriate and the register's improvement-loop.md; a fixing thread should
re-verify.]`

## 6. Fix candidates

**For the instance (this lane, in flight):**
1. Re-seed `imagery_style_guide` and the brief to ask for what the vertical wants — game imagery,
   playable/interactive elements where the practice seat can carry them, saturated where earned,
   energy — while keeping the seat distinct from the sibling (practice, not free tools). Regenerate
   imagery, rebuild. The designblog lane's routing (components / experience loop / theme kits /
   designer / vigilant designer / copy-quality-two-stage) is being joined, not duplicated.
2. Get real article pages into the plan — an editorial site with zero editorial is 444's shape by
   construction. Ask the planner for N articles with sections, not one slot.
3. Request a site-scope hero and a hero per section-index in `site_plan_imagery`, so the
   template's default path has a file behind it.

**For the class (unowned):**
4. **Plan-time validation** (444 candidate 1, extended): refuse a plan whose section-index lists a
   content type with zero planned pages; refuse a plan with a hero-bearing page and no imagery row
   for it and no site-scope hero.
5. **Palette-aware hero overlay**: the hero component's gradient is a constant; make its opacity
   and ink derive from the palette's `dark_light` so a light editorial palette gets a light hero.
6. **`image_url_404` reads CSS `url()`** in inline styles and stylesheets — the fourth shape.
7. **A vertical-appropriateness reviewer seat** in the experience loop: given the classifier's
   vertical and the served page, does the page look like its subject? Referent per vertical;
   flag-only; the owner's critique tonight is the first training case.

## 7. How to verify

Instance: the served site carries ≥1 game-related image per hero, ≥N article pages with content,
an articles index that LISTS them, no hero over a 404, no "What they avoid" section — read at the
bytes with the control. Class: candidate 4's refusal must fire on THIS site's 17:33Z plan
(1 article page, 0 sections; section-index with 0 members) when replayed — a plan that already
carries articles is not a test.
