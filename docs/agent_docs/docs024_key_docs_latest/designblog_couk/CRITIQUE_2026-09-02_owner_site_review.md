# Owner critique of designblog.co.uk — 2026-09-02, with verification

The owner reviewed the freshly-live designblog.co.uk (remake №4 of the portfolio
positioning lane's 22 hosted-site remakes, live 2026-09-02) and stated that **much
of this critique applies to the sibling remakes too** (advertise.co.uk,
websitepromotion.co.uk — same lane, same day, same pipeline).

The critique is reproduced verbatim below (§1), then each point verified against
the served site (§2). Verification method: `curl` of the public URLs on
2026-09-02 ~20:10Z; all pages returned HTTP 200. Commands in
`RUNBOOK_designblog_couk.md` (this directory).

## 1. The critique (owner's words, verbatim)

> designblog.co.uk
> overall
> the design is exactly the same as all the other sites, it should be different -
> every site doesn't have to have the same top and bottom nav and the big heroes
> with the same components. Please talk to the components thread, the experience
> loop, the themes kit, the designer thread and the vigilant designer thread about
> this. We should be able to make the designs more impactful. We are trying to
> make these sites best in class, please check that each of the relevant agents
> got that message.
>
> nav
> there is no tools nav link
>
> /inspiration/index.html
> This is explaining the brief and not answering it, please talk to the relevant
> threads to ascertain why this is still broken
>
> https://designblog.co.uk/the-design-feed/index.html
> the hero text is unformatted and long like boilerplate. please talk to the
> relevant threads to determine what went wrong here. The copy explains the brief
> and doesn't answer it.
>
> The tools links' copy is verbose and doesn't look like it's been through the
> copy agent, talk to copy quality two stage thread too. e.g. "Or run a colour
> pairing through the True-Contrast Smart Palette to check it against WCAG's AA
> (4.5:1) and AAA (7:1) contrast thresholds. When a pairing fails either level, it
> offers a fix that holds the original hue steady and adjusts only the lightness,
> and if no shade of that hue can reach the target on your chosen background, it
> says so plainly and falls back to black or white instead. Confirm any close call
> by eye, since the tool's reading is a starting point, not the final word."
>
> https://designblog.co.uk/tools/smart-contrast/index.html
> the components look the same as everywhere else, this has been addressed
> elsewhere.
> the copy at the bottom under free tools sounds like AI e.g. "Check your
> palette's contrast before your users have to" should say something like just
> "Check your palette's contrast"
> and " It's built for anyone choosing colours for a live interface, where a fail
> here means real users can't read the text. The ratio it returns is only as
> accurate as the hex values you feed it." should say someting like just "It's
> built for anyone choosing colours for a live interface. Good colours lead to
> more readily readable text. It can return a ratio that is representative of how
> readable text is against a readability scale"
>
> https://designblog.co.uk/uk-studios-directory/index.html
> the directory is empty
>
> https://designblog.co.uk/glossary.html
> the glossary has text about the brief and is not a glossary
>
> https://designblog.co.uk/inspiration/index.html
> has no inspiration. it is just about the brief.
>
> All text doesn't have enough images.
> The design doesn't have enough images and infographics and graphics.
>
> The site is not as delightful as the brief made out it would be.
>
> Please talk to the relevant threads on all these points to fix the underlying
> problems.

## 2. Verification against the served site (all as of 2026-09-02)

Every point checked; **every point confirmed.**

| # | Claim | Verified finding (2026-09-02) |
|---|---|---|
| 1 | No tools nav link | CONFIRMED. Header nav on `/` has exactly 6 links: Home, The Feed, Criticism, Inspiration, Studios Directory, Glossary. No Tools — while the site serves `/tools/smart-contrast/index.html` (200, 92KB) and the copy names a CSS Variable Architect and CSS Unit Converter. The tools are reachable only through body copy. |
| 2 | Directory is empty | CONFIRMED. `/uk-studios-directory/index.html`: after the intro paragraph the next `<p>` is the footer email. **0 studio entries, 0 `<article>`, 0 content `<h3>`** as of 2026-09-02. The page's own H1 promises "UK design studios, organised so you can actually find one". |
| 3 | Glossary is not a glossary | CONFIRMED. `/glossary.html`: **0 term entries** (no `<dt>`, no term markup). Its 3 content `<h3>`s are meta-headings: "How the entries are written", "What's covered", "Where the definitions lead" — a page about the glossary it intends to be. |
| 4 | Inspiration has no inspiration | CONFIRMED. `/inspiration/index.html`: **0 showcases**. `<h3>`s: "What gets included", "How to use a showcase", "Where to go from here". Body copy describes admission criteria for showcases that do not exist ("A piece earns a place here because…"). |
| 5 | Feed hero unformatted/boilerplate; explains the brief | CONFIRMED. `/the-design-feed/index.html`: **0 feed items**. Hero paragraph is a single long wall listing the site's tools; the section `<h3>`s are "What actually lands here", "How it's ordered", "Where things go once they've grown" — the shape of a brief, not a feed. |
| 6 | Smart-contrast copy sounds like AI | CONFIRMED. Both quoted sentences are verbatim in the served page: "Check your palette's contrast before your users have to" and "It's built for anyone choosing colours for a live interface, where a fail here means real users can't read the text. The ratio it returns is only as accurate as the hex values you feed it." |
| 7 | Not enough images/infographics | CONFIRMED. **Exactly 1 `<img>` tag per page** on all 6 pages fetched (/, inspiration, the-design-feed, tools/smart-contrast, uk-studios-directory, glossary), as of 2026-09-02. |
| 8 | Design identical to sibling sites | The shared-chrome/shared-hero composition is the pipeline's design (one composition library, one chrome pattern); the sibling remakes shipped through the same pipeline the same day. Routed to the five design threads (see §3) rather than re-measured here — it is a mechanism property, not a per-page defect. |

> **ADDENDUM 2026-09-02 (sharpens #7):** the one `<img>` per page is the **header
> logo** — chrome, not content (verified at the served bytes: all 6 pages'
> single img is `/assets/images/logo.png` inside `<header>`). The components
> thread's stored-markup census the same day agrees and adds controls:
> designblog.co.uk carries **ZERO images across all 50 of its component slots**;
> advertise.co.uk 1 of 62, websitepromotion.co.uk 1 of 37 — against controls of
> 4/43 (garden-tools.uk) and 5/48 (boxingonline.com), so the query does detect
> images. The three remakes are the most image-poor sites on the estate. ~~Content
> imagery on this site is not sparse — it is absent.~~
>
> **CORRECTED 2026-09-02 (same evening; caught by the editorial_design_uplift
> lane, re-verified here on the served bytes):** "absent" was FALSE — heroes on
> this estate render as CSS `background-image` URLs, invisible to any `<img>`
> census, and **every one of the 6 pages serves a real hero image**
> (`/assets/images/hero-*.jpg` behind a darkening gradient). Both measurements
> above were correct and both encoded "image = `<img>` tag" — two agreeing
> measurements sharing an encoding are one measurement (WRONG_CALLS 2026-09-02).
> The corrected statement: designblog serves **heroes + section icons + a logo
> (10 planned imagery rows, all active) and has ZERO illustration and ZERO
> infographic rows** — the owner's complaint survives intact; the "zero images"
> number does not. Also found in the recount: the 6 pages share only **3
> distinct hero files** (the feed and the contrast tool reuse the homepage's
> hero) — a sameness finding of its own.

### The pattern behind #2–#5

Four listing-type pages (directory, glossary, inspiration, feed) all shipped as
**prose ABOUT the intended content with zero content items**. The copy on each
restates the brief's promise instead of delivering it. This is exactly the class
the experience loop's two live detectors target (listing-class 08-31,
experience-promise 09-02) — put to that thread to establish whether the detectors
ran on the four remakes and what they said.

## 3. Routing (who was told, 2026-09-02)

Messages sent from this session ("designblog.co.uk") to the live sessions below;
ACK status tracked in `NOTES_designblog_couk.md`.

| Thread (session name) | What was routed |
|---|---|
| Portfolio positioning | Full critique — they built the site; empty listing pages + brief-echo copy + missing tools nav; class applies to the other 18 remakes before the next briefs fire |
| components | Design sameness: same top/bottom nav + big heroes with the same components on every site; best-in-class directive |
| experience loop | Empty listing pages / experience-promise gap on a day-old build; did the two live detectors run on the four remakes? |
| theme kits | Per-site visual differentiation; more impactful designs; best-in-class directive |
| site design planner | Composition sameness at the mechanism level (one composition vocabulary → identical site shapes) |
| offer analyser benefit analyser visual designer (vigilant designer lane) | Visual impact, images/infographics scarcity, delight gap |
| copy quality two stage | Verbose tool-link copy + AI-sounding copy with the owner's before/after examples + brief-echo copy class |
