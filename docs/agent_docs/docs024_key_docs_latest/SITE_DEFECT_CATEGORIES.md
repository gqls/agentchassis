# SITE_DEFECT_CATEGORIES.md — what to check on every newly built site, before it goes to anyone

**Created 2026-09-02 at the owner's direction**, from his review of boxingonline.com — the first
paid customer build. His instruction: *"please list those errors in a site error file somewhere so
we can check the category with each new site."*

**Fleet-wide. Append-only. Any thread may add a category.**

## What belongs here, and what does not

This is the **acceptance checklist for a built site**: the classes of defect a site can ship with
while every internal check passes. It is a sixth destination alongside the five in `CLAUDE.md`:

- how the SYSTEM fails, once you have a symptom → `016b` §9
- how WE were wrong → `WRONG_CALLS.md`
- what misleads you when you TOUCH something → `LANDMINES.md`
- what a workstream is doing → the standing five
- what EXISTS and is callable → the concept register
- **what to CHECK on a site before delivering it → HERE**

**The bar for an entry: a CATEGORY, not an instance, and it must carry a runnable check.** A
category with no check is a story. Where a check is already automated, say so and name it, so
nobody hand-runs what a cron already does.

> **Why this file exists at all.** Every defect below shipped on a site where *every page
> validated `valid=true, issues=0`*, every work item completed, and every build reported success.
> **Not one of them announced itself.** They were found by a human reading the site, and by
> sessions probing the served artefact. That is the premise of the whole file: our internal
> signals did not discriminate, so the checks have to run against what a reader actually gets.

---

## 0. FIRST — how to run any of these without fooling yourself

Get these wrong and every check below returns a confident, meaningless answer.

- **Probe the SERVING host, never the customer's domain.** A parked customer domain answers 200
  for every path with a stub. Get the real one from `sites.publish_target` / `publish_project`.
  Control: an invented URL must 404. `scripts/probe-page-url.sh`.
- **Every probe carries a must-be-present control** (a string that has to be in the page). A
  control of zero means the fetch failed — report it as BLIND, never count it as clean.
- **Enumerate pages from the database, not from memory.** `pages WHERE deployed_at IS NOT NULL`.
  A hand-listed subset silently under-reports and reads as a pass.
- **A completed work item is not a changed artefact.** Nor is a fresh `deployed_at`. Read the
  artefact.
- **Date the artefact before accepting "latency".** If the published object's timestamp is OLDER
  than your change, wait. If it is NEWER and still wrong, stop waiting and look upstream.
- **In a `while read` loop, give any inner `kubectl exec -i` a `</dev/null`** or it eats the
  loop's stdin and your sweep reports one row, cleanly.
- **In a Postgres regex `\b` is BACKSPACE, not a word boundary — use `\y`.** The wrong one returns
  a zero that reads as an absence.

---

## 1. CONTENT SAYS THE WRONG THING

### 1.1 Meta-copy — the page describes its own editorial policy instead of doing its job
Sections headed "How we cover it", "Keeping it accurate", "How entries get added". Often a
paraphrase of our own `vertical_landscape.lessons` — the researcher's instruction sheet emitted
as page copy.
**Check:** first-person-plural policy prose on a page whose job is to serve content —
`grep -Eci "we write|we'd rather|we cover|gets checked|we update"` on the served body.
**Seen:** boxingonline about page, articles index, and the fight-calendar tool page (~2,000 chars
of process description on the brief's core deliverable).

### 1.2 A title promises a specific, dated, checkable thing the body does not contain
"Last night's result…" containing no result. **Not mechanically checkable** — a proper-noun/date
count scores such a page ABOVE a correct short report (the boxingonline case names nine proper
nouns and two dates while containing no news). Refused as a regex by the experience lane on that
evidence; it needs a reader.
**Check:** human read, or the copy lane's title-promise gate when it lands.

### 1.3 The site holds the real material and the page does not use it
The sharpest form of 1.2. Boxingonline's news page carried the true dated result while the
article promising that story was written from the model's general knowledge.
**Check:** for any site with a news/feed surface, sample one article and grep the *other*
surfaces for entities the article should have named.

### 1.4 Raw source residue served as content
Literal markdown (`[text](url`), fragments truncated mid-word, scraped navigation from the source
site, off-topic items.
**Check:** `grep -c '](http'`, `grep -o '\.\.\.' | wc -l`, and an off-vertical term count.
**Related:** `bugs_open/332`.

### 1.5 AI-tell copy that went THROUGH the copy gate (added 2026-09-02, designblog review)
Verbose tool-link paragraphs, closers like *"before your users have to"* / *"a starting point,
not the final word"* / *"says so plainly"*, and a 450-char essay in a BUTTON field — **on a build
that ran the full gated copy path.** The copy lane's decomposition (measured at the build's own
markers): the gate SAW and TARGETED one example and the repair model returned
`no_answer_for_target`, so the original shipped by design; banned WORDS currently have
detection-only (nightly checker), no page-side repairer; the rest are ruled classes queued.
**Gate-ran is not gate-repaired.**
**Check:** grep served pages for the known tells (`plainly`, `honest`, `before your .* have to`,
`starting point, not the final word`); flag any control-labelled field (`button`, `label`,
`secondary_cta`) whose value exceeds ~120 chars. The nightly banned-words checker covers BRIEFS —
it does not repair pages.
**Seen:** designblog.co.uk /tools/smart-contrast + the-design-feed, 2026-09-02 (owner quoted two
verbatim); same-day siblings advertise/websitepromotion expected to carry the same classes.

---

## 2. LISTINGS AND NAVIGATION

### 2.1 A listing shows the wrong content class
A "Latest news" block listing explainer guides; a guides index listing blog articles. Cause is
usually that the resolver has one vocabulary and it is not the page's.
**AUTOMATED:** experience-loop rule C (nightly) — an index-role page whose own directory holds
active pages while its listing shows none of them.
**⚠ A PASSING listing may pass coincidentally** — it passes when the resolver's one vocabulary
happens to be the right one for that page. A zero is not evidence the page is well-founded.

### 2.2 An index with nothing to index writes a manifesto
Emptiness does not present as emptiness; it presents as well-written in-voice policy prose that
validates clean.
**Check:** listable-item count of every index-role page; zero is a finding on its own.
**Also seen (2026-09-02, designblog.co.uk — FOUR pages on one day-old site):** glossary 0 terms,
inspiration 0 showcases, feed 0 entries, studios directory 0 studios; the tell is meta-`<h3>`s
about the intended content ("What gets included", "How the entries are written"). Same shape on
advertise same day ⇒ class home is **`bugs_open/444`**.
**Upstream check (444) — ask whether anything could EVER fill it, per page type:** FEED → does the
site have `content_sources` rows (all four 09-02 remakes: 0)? DIRECTORY → is the entity kind
registered in the DIR-001 pipeline? GLOSSARY/INSPIRATION-type item pages → **no item producer
exists anywhere in the estate** as of 2026-09-02. A zero with no producer is a PLAN defect, not a
content delay; 444's fix candidate is plan validation refusing/degrading a listing page whose item
source resolves to zero.
**⚠ the automated rules cannot see this case:** experience-loop rule C drops an index with no
`articles`/`items` array **before the rule runs** — an index that lists NOTHING is invisible to it
by construction (empty-index rule in build, experience loop, 2026-09-02). Until it lands, this
check is manual.

### 2.3 Duplicate nav entries — BOTH directions
Two pages sharing one label, AND one page under two labels. A rule written for one half cannot
see the other.
**AUTOMATED (half):** experience-loop rule A covers two-pages-one-label.
**Check the other half by hand:** count distinct hrefs vs distinct labels in `<header>`.

### 2.4 Orphan pages — deployed, in the sitemap, linked from nothing
**Check:** for each deployed page, count inbound links from every other deployed page. Zero
inbound + present in `sitemap.xml` + `noindex=false` is the state to flag.

### 2.5 A correctly-configured nav row that never renders
The row is populated and ordered and still absent, because a *type bar* classified it elsewhere
(`page_type='tool'` is barred from primary nav). Debugging the row finds nothing wrong.
**Check:** compare `pages WHERE in_header` against the served `<header>`. Remedy is the
declaration mechanism (`site_specs.data->'chrome'->'header_slots'`), not the row.

### 2.6 A live surface with no nav path at all (added 2026-09-02, designblog review)
Working tool pages serving 200, reachable ONLY through body copy — no nav link, and no hub page
for the family. Distinct cause from 2.5: **tool pages arrive post-plan via tool-deployer, the nav
rebuild runs before they exist, and the plan never plans a tools hub** — so nothing was ever
misconfigured; the page the link would point at was never planned. On advertise the same day:
seven tools serving, no /tools/index.html, `nav_order` 4 conspicuously empty.
**Check:** for each deployed page-type family (`tool`, `guide`, …), does at least one served nav
label reach it (directly or via a hub)? A family with N serving pages and zero nav paths is the
finding. Cheap per-site remedy: a tools section-index page with `in_header`, then a nav rebuild.
Class ownership is split (site-planner: plan the hub when tools are coming; bugfix-149
nav-membership family).
**Seen:** designblog.co.uk (smart-contrast serving, 6 nav links, none tools), advertise.co.uk,
both 2026-09-02.

---

## 3. TOOLS AND DATA

### 3.1 A tool ships no data — the reader supplies everything
A "comparison tool" with 18 empty inputs. A calculator is not a tool if we bring nothing.
**AUTOMATED:** experience-loop rule B — a `page_type='tool'` page serving no control, no inline
data and no runtime fetch.
**Check by hand for the softer case:** name the reader-supplied inputs and the site-supplied data
separately. **An empty site-supplied set means it is a form, not a tool.**
⚠ Read STORED component markup, not the served page — site chrome carries a mobile-menu `<button>`
that makes every served page look like it has a control.

### 3.2 A page presents a set, has no set, and has no empty state
The visitor cannot tell whether there is nothing or whether it is broken.
**Check:** any page whose job is to present a collection must say something true when the
collection is empty.

### 3.3 The site's structured fact corpus is empty
The root cause of 3.1 and 3.2. `site_specs` aspect=`evidence_base`.
**Check:** `SELECT jsonb_array_length(data->'facts') FROM site_specs WHERE site_id=$1 AND
aspect='evidence_base' AND is_current;`
**Fleet baseline 2026-09-02:** 20 of 54 sites have a row at all; **42 of 54 hold ≤5 facts**.
**Related:** `bugs_open/427`.

---

## 4. IMAGERY AND ASSETS

### 4.1 Assets generated, deployed, and undisplayable
Six article hero images existed, served 200, and no article page referenced one — because the
article component has a single text field and cannot hold an image by construction.
**Check:** for each `assets` row, grep the page that should show it. **Generation success is not
display.** Fleet 2026-09-02: 65 of 429 components carry image markup; 364 do not.

### 4.2 Everything except the logo
**Check:** `curl … | grep -o '<img[^>]*>' | wc -l` per page. A count of 1 is the logo alone.
> **CORRECTED 2026-09-02 (designblog review; WRONG_CALLS same day, commit `5d76472d8`): the
> check above is HALF-BLIND — heroes on this estate render as CSS backgrounds, invisible to any
> `<img>` count.** An "images absent" verdict off `<img>` alone was published and was false. Run
> BOTH greps, always together:
> `grep -o '<img[^>]*>' | wc -l` **and** `grep -o "background-image:[^;}]*url([^)]*)"`.
> Two agreeing measurements that share an encoding are one measurement.
**Companion check — hero REUSE within a site:** count distinct `background-image` URLs against
pages carrying a hero. designblog 2026-09-02: 6 pages, **3 distinct hero files** (the feed page
and the contrast tool both wear the homepage's hero) — sameness inside one site.

### 4.3 Invented brand identity in a generated asset
A logo lettered with a brand name nobody chose ("BOXING NEWS" on Boxing Online; "Farm Shield
Info" elsewhere) — a model told to letter a wordmark and never told what it should read.
**Check:** LOOK AT THE IMAGE. Then compare any lettering against `sites.company_name` /
`logo_text` / the domain stem. **Related:** `bugs_open/417`.

### 4.4 An asset that is not the shape it claims
A two-panel design comp served as a single logo mark.
**Check:** compare the asset's dimensions against the generator's requested `kindDefaults` **at
the generator output** — a served file may be resized, and then dimension proves nothing.
**Related:** `bugs_open/421`.

### 4.5 A background baked into a logo
Invisible on the one surface it is used on, wrong everywhere else. Owner ruling 2026-09-02: *"the
background behind a logo shouldn't be part of the logo."*
**Check:** PNG colour type 6 or 4, **or** a `tRNS` chunk — test for both; either alone gives a
false negative. **Related:** `bugs_open/424`.

### 4.6 The imagery plan is chrome-only — and the planner is OBEYING, not failing (added 2026-09-02)
A site ships with heroes + icons + a logo and nothing that carries content: **zero illustration
rows, zero infographic rows planned.** Fleet totals ever requested (measured 2026-09-02): hero
399 / icon 211 / logo 50 / illustration 25 / **infographic 1**. The live `build-site-planner`
prompt already carries the full `kind` vocabulary; three things in it suppress the rest — a
verbatim *"use sparingly in v1 — most plans will have zero section-scope entries"* instruction, a
chrome-only stated minimum (rule 13), and a worked example whose sections block shows ONLY icons
(exemplars ship verbatim). **So the near-zero is compliance; no vocabulary or machinery fix moves
it — the remedy is a deliberate prompt edit (owner cost decision), keeping rule 16
(one-image-per-entry) in the same change.**
**Check:** `SELECT kind, count(*) FROM <planned imagery rows for the site> GROUP BY 1;` — a plan
with no illustration/infographic entries on a content site is this category.
**Source:** inline_guide_imagery `NOTES_inline_guide_imagery.md` §15 (verbatim prompt quotes);
designblog_couk lane NOTES 2026-09-02.
> **REMEDY LANDED 2026-09-02 (owner-directed): migration 718** flipped all four suppressing
> surfaces (bullet, rule-13 floor, worked example, skeleton; rule 16 kept, rule-3 pages exempt) —
> live at 19:59Z, council corr `2dae4f20`. **The check above stays**: for a site PLANNED after
> 2026-09-02 ~20:00Z, zero illustration/infographic entries is now a defect signal, not
> compliance; for older plans it remains expected history. First canary = portfolio
> positioning's next remake brief.

### 4.7 Article pages are one prose slab — there is no structure to place imagery INTO (added 2026-09-02)
Even with illustrations planned, in-article imagery has nowhere to land: measured 2026-09-02,
across all 462 blog/guide pages fleet-wide the maximum prose-section count on any article page is
**1**; of 9 illustration-capable sections fleet-wide, 8 are on LANDING pages and 0 on blog/guide.
**A prompt/plan fix for 4.6 puts pictures on landing pages only; in-article imagery additionally
needs composition change.** Keep the two asks separate.
**Check:** count prose sections per article-role page. **⚠ count the unit the claim is about** —
section ROWS include hero and CTA; the PROSE count is what decides this (WRONG_CALLS 2026-09-02,
inline_guide_imagery).
**Related:** `bugs_open/114` (three lanes converged on this the same day).

---

## 5. COMPONENTS AND RENDERING

### 5.1 Empty slots rendered as empty elements
Four of six card slots empty and still holding layout. Reads as bad design; is a data gap.
**Check:** `grep -c 'class="[^"]*__excerpt"></p>'`-style scans for empty slot elements.
**A component that renders an empty element cannot tell a missing input from an intentional
blank.** **Related:** `bugs_open/425`.

### 5.2 The same value under different names by different producers
Template reads `.excerpt`; the resolver writes `meta_description`; the deck was in the row the
whole time. Also: a raw `<title>` including the " | Site Name" suffix rendered as a card headline.
**Check:** dump one item of a listing's `content_data` and compare its keys against the
component's template interpolations.

### 5.3 Unsubstituted placeholders shipped to production
`mailto:OWNER_EMAIL_ADDRESS`, `Article | Boxing Online`, a stray field label ("Free Cost")
rendered as body text, `placehold.co` in an `og:image`.
**Check:** grep the served site for `placehold`, `EXAMPLE`, `TODO`, `_ADDRESS`, `Lorem`, and any
ALL_CAPS_UNDERSCORE token.

---

## 6. PRIVACY, IDENTITY AND DEAD CONTROLS

### 6.1 The ordering party's contact details published as the site's
The billing email became `sites.email` and thence the footer of every page — **and was minted as
a registered, renderable claim**, so a rebuild could re-emit it and pass validation.
**Check:** grep the served site for the customer's ordering email; then check `sites.email`,
`site_specs` (`briefing`, `evidence_base`), `site_components`, `page_components`, and
`build_queue.direction`. **Owner ruling: the identity of whoever commissions a site is
independent of the site's operation.** **Related:** `bugs_open/420`.

### 6.2 A form that submits nowhere
**Check:** `SELECT content_data->>'form_action', count(*) FROM page_components WHERE content_data
? 'form_action' GROUP BY 1;` **Fleet 2026-09-02: 22 components carry it and `#contact` is the
only value any of them has.** Every form the estate has built is a decoration.
**Related:** `static_site_form_endpoint/PLAN_2026-09-02_pre_plan_extensible_form_endpoint.md`.

### 6.3 A retracted page that still answers
Deleting a page and removing its links does not remove the published object — the mirror has no
delete capability, so the URL keeps serving.
**Check:** deletion closes on **404 AND zero inbound links**, never one of the two.
**Related:** `bugs_open/429`.

---

## 7. PLAN AND BRIEF FIDELITY

### 7.1 Page types the strategy specified and the plan never emitted
`strategy.recommended_page_types` named `entity-page` (fighter profiles) and `entity-directory`
(per-event pages) with reasoning; the plan emitted neither, and both are live roles on other
sites.
**Check:** diff `strategy.recommended_page_types` against `site_plan_pages.role`.

### 7.2 A brief promise with no page behind it
The paid brief named six editorial slots; the planner emitted one zero-section page and the link
validator then rewrote the dead CTAs so the absence looked clean.
**Check:** diff `build_queue.direction` against the page inventory, by hand. **Nothing else does.**
**Related:** `bugs_open/419`.

### 7.3 The site does not aim at its own research
`vertical_landscape` names what best-in-class looks like and a `differentiation_opportunity`; the
strategy carries it; the plan drops it. Measured: `vertical_landscape` is read by **two** agents
fleet-wide, and in Go only as an existence check.
**Check:** read the `differentiation_opportunity` and ask which page delivers it.

---

## 8. THE ONE THAT PASSES ITS OWN CHECK

**A work item can complete, carry the correct reason, and change nothing.** Measured 2026-09-02:
a card fix ran twice on one page with `reason='template_changed'` — the discriminator three
sessions had converged on that day — and the stored data was unchanged, while the same fix via
the BUILD path ten minutes earlier came out correct. Same binary, same site.
**A correct reason is necessary and not sufficient. Always close on the artefact.**

---

## 9. CROSS-SITE SAMENESS (added 2026-09-02, from the owner's designblog review)

**No single-site audit can see this class** — `visual-design-auditor` audits one site in
isolation and scores coherence, not distinctiveness, so a site can pass every per-site check and
still be the owner's *"the design is exactly the same as all the other sites"*. These checks
compare a NEW site against a cohort (its build-wave siblings, or the fleet). Owner directive
2026-09-02: sites should differ — different chrome, different hero treatments, different section
rhythms; *"we are trying to make these sites best in class."*

### 9.1 Component-set overlap with the cohort
Measured 2026-09-02: the fleet's top-10 components carry **78–87% of ALL slots** on typical
sites; two thirds of the 156-component section library is unused or single-site. Sameness is a
SELECTION property, not library poverty.
**Check:** the cohort-overlap query in
`designblog_couk/RUNBOOK_designblog_couk.md` (§ pre-flight) — anything on ALL cohort sites is the
sameness, its instance count is the size.
**⚠ Components carry THREE names** — 76% of active components have `name <> function`, and
`page_components.slot_name` agrees with neither reliably (LANDMINE `0b3151337`). Resolve through
`component_id`, and state WHICH column your count reads; a zero from the wrong column reads as
"does not exist" and licenses rebuilding a thing that is on half the estate.

### 9.2 Identical chrome by construction
36 of 37 sites render the same `site-header` + `site-footer` (measured 2026-09-02) while 10
chrome-eligible header functions sit unused — `ChromeSlotFunction()` hardcodes slot→function, and
only 6 of 40 `style_collections` pin `header_component_id`.
**Check:** does the new site's `style_collections` row pin a header/footer, or is it falling
through to the hardcoded default (NULL links)? A per-site `UPDATE` pinning an unused header is
the available-today remedy; `page_archetypes` (theme kits, committed-unapplied) is the mechanism.

### 9.3 Identical layout because the library has one answer for this site shape
All three 09-02 remakes resolved to `magazine-grid` — matcher verified correct; the library holds
exactly ONE professional-register editorial layout and NO "content hub with embedded tools"
archetype, so different briefs converge honestly. 9 of 18 layouts have never been chosen by any
live site; 3 carry 73% of the fleet.
**Check:** compare the new site's resolved layout against its cohort's; three-in-a-row identical
is this category. **Related:** `bugs_open/445` (the archetype gap is the fix that compounds).

### 9.4 One hero image wearing several pages (within-site sameness)
See the companion check under 4.2: distinct `background-image` URLs vs pages carrying a hero
(designblog: 6 pages, 3 files).

---

**Adding an entry:** category, what it looks like, a runnable check, and the instance that
produced it. If you cannot write the check, say so explicitly (see 1.2) — a named unmechanisable
category is worth more than a weak check that looks like the others and quietly does not work.

## 10. THE SITE'S TEMPERATURE CONTRADICTS ITS VERTICAL (added 2026-09-02, from the owner's gamedesign.uk review — "it is a major error")

A games site shipped as a sober print journal: one image (the logo) plus three pencil-and-paper
still lifes under a 60% black wash, zero games, zero interactive elements, an articles index with
zero articles that describes what the articles would be like, and copy that examines game design
like a legal periodical. **Every internal check passed.** The owner: *"It is a game design site —
why isn't it full of games and images and excitement?"* Full instance and mechanism:
`bugs_open/446`. The lesson is the file's premise again: presence checks (images resolve, sections
exist, claims registered) cannot see APPROPRIATENESS. A site with nothing to say, saying it
carefully, is green. Referent per vertical — `oufe.com` (restructuring finance) is CORRECTLY sober;
the defect is the mismatch, not the register.

### 10.1 The spec bans what the vertical is made of
The lane's own `imagery_style_guide.avoid` read *"screenshots or renders of real or invented games;
game characters… cartoon or anime styling; saturated primaries"*, and the generated hero prompts
carried it verbatim: *"…a folded document and a pencil on a warm linen surface… **no game imagery**."*
The pipeline obeyed. **Briefs govern** (designblog found the same: its restraint came from its brief).
**Check, before dispatch:** read the site's `imagery_style_guide` and `mission`/`mission_brief`
against the classifier's vertical nouns —
```sql
SELECT data->>'avoid' FROM site_specs WHERE site_id='<id>' AND aspect='imagery_style_guide' AND is_current;
SELECT data->>'industry', data->>'category' FROM site_specs WHERE site_id='<id>' AND aspect='classification' AND is_current;
```
A ban that names the vertical's own subject (games on a games site, food on a food site, faces on a
portrait site) is the defect. **Check, after build:** the prompts actually sent —
```sql
SELECT result->'response'->'image_result'->>'prompt' FROM site_work_items
 WHERE site_id='<id>' AND item_type='needs_imagery' AND status='complete';
```
`grep -ci "no <vertical noun>"` over them; any hit is 10.1 fired.

### 10.2 An index page with zero members of its own type (the brief-echo shape — `bugs_open/444`)
The served page is 200 at full weight, headed with meta-prose ("What the pieces do", "What they
avoid", "What gets included", "How the entries are written") and lists nothing. Mechanism here: the
plan created ONE article page with ZERO sections, so the type had no producer at all.
**Check, at the plan:** for every `section-index` page, count planned content pages in its section —
```sql
SELECT spp.name, (SELECT count(*) FROM site_plan_pages c WHERE c.plan_id=spp.plan_id AND c.parent_section=spp.slug AND c.role<>'section-index') AS members
FROM site_plan_pages spp WHERE spp.plan_id=(SELECT id FROM site_plans WHERE site_id='<id>' AND is_current) AND spp.role='section-index';
```
`members = 0` before a word is written is the whole defect; 444's plan-time refusal (in council
2026-09-02) is the automated form. **Check, at the artefact:**
`curl -s <index-url> | grep -ciE 'what (they|we) avoid|what gets included|how the entries are written|what the pieces do'`
— any hit on an index page is brief-echo. Negative-identity headings ("What they avoid") are ALSO
owner-banned copy in their own right (2026-09-02).

### 10.3 A hero over a 404 — the CSS-url shape 4.2 cannot see
`/articles/index.html` rendered `background-image: linear-gradient(…), url('/assets/images/hero.jpg')`
and that path was 404: the planner requested no site-scope hero and none for the section-index, so
the template's default had nothing behind it. `check_image_url_404` inspects `<img src>` only.
**Check:** extend 4.2's second grep with a resolve step —
```bash
curl -s "$URL" | grep -oE "url\('?[^')]+'?\)" | grep -oE "/[^')]+" | sort -u | while read -r u; do printf "%s %s\n" "$(curl -s -o /dev/null -w '%{http_code}' "https://$DOMAIN$u")" "$u"; done
```
Any non-200 is a hero (or section) painting a gradient over nothing. **Control:** the same default
path served 200 on all six other sites probed — do not read one site's 404 as a fleet defect.

### 10.3b The WRONG hero — a presence check passes on the homepage's picture
Two pages wore the homepage's hero while their own generated heroes sat orphaned (0 references,
files 200). Cause at the library: `hero-about` (28 sites) and `hero-contact` (25 sites) declare no
image field, so the per-page asset falls through. **Check:** a filename+extension-anchored census —
for each `assets`/`site_plan_imagery` hero row, `grep -c '<file>.jpg'` across the served pages, with
a known-referenced hero as the control; then count DISTINCT hero urls against pages carrying a hero
(4.2's companion). One url on N pages is this category. ⚠ **Do not census with an unanchored
`LIKE '%hero-about%'`** — it matches `data-component="hero-about"` and says the opposite of the truth.

### 10.4 Zero interactive or playable elements on an interactive vertical
The practice seat of a games pair had no game, no demo, no embed, no playable anything, and one
link to the sibling's tools. **Check:**
`curl -s "$URL" | grep -ciE '<canvas|<iframe|<video|data-game|data-tool|<button(?![^>]*type="submit")'`
across the site's pages, and `grep -c 'gamesdesign.co.uk'` (or the vertical's tool host) for cross-
links. On a games / tools / demo vertical, a site-wide zero on the first and ≤1 on the second is
this category. Not a rule for every vertical — a law firm scores zero and is right to.

### 10.5 The class has no detector, and the fix is a reviewer with a referent
Nearest instrument: the `experience-promise` family (live 2026-09-02) checks a page delivers what
its own headings promise — which a restrained site does, immaculately. `content-quality-auditor`
post-migration-694 carries a "does an index list its own items or write ABOUT itself" dimension
(filing_mode=record) and is the one that should catch 10.2. Nothing reads the served page against
the VERTICAL's temperature. Proposed in `bugs_open/446` §6.7: a flag-only reviewer seat given the
classifier's vertical and the served bytes; the owner's two critiques tonight (designblog,
gamedesign.uk) are its first two training cases. Until it exists, **a human reads the site** — which
is how both were found.
