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

**Adding an entry:** category, what it looks like, a runnable check, and the instance that
produced it. If you cannot write the check, say so explicitly (see 1.2) — a named unmechanisable
category is worth more than a weak check that looks like the others and quietly does not work.
