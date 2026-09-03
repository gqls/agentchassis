# DESIGN CRITIQUE 2026-09-03 — finetuning.uk, the design-critique-agent's report (verbatim) and what it means for the homepage cards

Run: `design_critique_run` item `204f1ff7`, label `owner_request_homepage_cards_and_infographics`, complete 21:38Z; 16 images (8 pages × desktop + mobile); no measured findings filed (contrast/broken images are a separate drain). Report body from `doc_notes` categories 'design-report', verbatim:

```
I was shown 16 images: 8 pages, each at a desktop and a mobile width.

Site-wide, three things repeat consistently. First, every page runs the identical visual rhythm of dark navy hero, cream three-column card section, dark navy CTA band, then the same enormous footer with what looks like 25+ links (the "Explore/Quick Links/Tools" mega-footer). By page four or five this footer becomes wallpaper — a visitor scrolling past it on the services or approach page gets no new information, only a repeated wall of link text that dwarfs the actual page content above it in scroll length. Second, the hero images: the homepage, your-own-model, use-cases, case-studies, and how-we-work pages all reuse the same dark grey/teal geometric triangle graphic as the hero background. Seeing it a third time (index 9, 11, 13, 15) makes the site feel templated rather than considered — these pages are otherwise about quite different topics (case studies vs. process vs. use-cases) but announce themselves identically. Third, the icon-in-circle treatment for card headers (magnifying glass, gear, document icons in pale circles) is used at near-identical size and style on homepage, services, and about pages — fine individually, but combined with the repeated hero and footer it flattens the distinctiveness between pages; a visitor could not tell from a mid-scroll screenshot alone which page they're on.

Homepage (1, 2): hierarchy is good — headline, subhead, dual CTA read cleanly. The "what we build" six-card grid is solid, but the cookie banner sits awkwardly overlapping the card grid boundary in the mobile capture. Further down, the "orange left-border" text blocks (A model you own / Your data stays / We're not tied to one provider, etc.) are a good distinctive device — this is the strongest section of the site.

Your-own-model (3, 4): the FAQ accordion list at the bottom is long (7 items) with generous vertical padding between each closed row — on mobile this stretches the page considerably before reaching the CTA; consider tightening row padding by roughly 20%.

Services (5, 6): the two-column then odd single-card row ("High-volume data acquisition") at the bottom leaves an orphaned card sitting alone against a full-width empty gutter on desktop — either add a 12th card or let that section resolve to a 2-column layout for the final row.

Case-studies (13, 14): four cards resolve to a 3+1 grid, leaving "Automated News Collection" isolated in its own row with a large empty right-hand gutter — same orphan problem as services, more visible here since the card is noticeably shorter than its siblings, drawing the eye to the imbalance.

How-we-work (15, 16): the five-icon-card step row is well composed and the accompanying numbered-list explainer below is a nice change of pace from the card monotony elsewhere on the site — this page reads as the most confident piece of layout in the set.

Nits: on the homepage mobile view (2) the hero copy block runs quite close to the following section with little visual breathing room compared to desktop; the about page (7, 8) "Head of Quality/Maintenance" role cards resolve to a 3+1 grid with the same orphan issue as services and case-studies — worth fixing site-wide with a consistent 3-up card count instead of numbers like 4 or 7 that don't divide evenly by 3.
```

## What it means for the owner's ask (cards, carousel-like structures, imaginative)

1. **The monotony is site-wide, not one grid:** navy hero → cream three-column cards → navy CTA →
   mega-footer, on every page; icon-in-circle card headers identical across pages; the same hero image on
   five pages. A carousel on the homepage alone buys distinctiveness for one page; the rhythm is the
   deeper issue and is a site-level composition question (`site-design-planner` / the uplift lane's
   imagery half for the hero reuse).
2. **The one homepage slot the critic actually faults is `case-studies-grid`: four cards resolve to 3+1,
   an orphan with a large empty gutter, more visible because the orphan is shorter.** A horizontally
   swipeable carousel removes the orphan by construction and is the "carousel-like structure" the owner
   named. That is the canary.
3. **The homepage's "what we build" six-card grid is "solid"** and the **orange-left-border text blocks
   (`differentiators`) are "the strongest section of the site"** — keep both as they are; the uplift
   lane's comparison infographic should sit WITH the differentiators device, not replace it.
4. **How-we-work's five-icon step row + numbered list is the most confident layout on the site** — a
   change of pace from cards. That is the shape a "three steps" section would take if the owner wants one.
5. Two chrome nits outside this ask: the cookie banner overlaps the card grid on mobile; the mega-footer
   dwarfs page content by page four.
