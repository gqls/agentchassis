# PLAN — webdesign.co.uk Phase 2: the buyer track (D10) and the copy rewrite (W2)

> **SUPERSEDED IN PART, 2026-07-27 (same day) — read
> `PLAN_2026-07-27b_buying_design.md` first.** The owner redirected the buyer
> track within hours of this being written. **What is dead here:** the `Hire`
> nav label (rejected — reads as Upwork/Fiverr, the wrong end of the quality and
> price scale), the "how to judge a quote" page (rejected — focuses the section on
> money rather than design), and the implicit general-purpose small-business
> buyer. **What survives and is still load-bearing:** §1's measurement that the
> buyer track is 100% new writing, §2's "buyer uses our tools to check what they
> were sold" bridge, §5's no-figures rail (the owner strengthened it), and §6's
> British-English finding. Kept rather than deleted because the rejected version
> is evidence about how the target audience was initially mis-sized.

**Written 2026-07-27**, immediately after the owner ruled D10/D11/D12 (recorded in
`PLAN_2026-07-25_webdesign_couk.md`). This plan covers the work those rulings
created. It is a proposal to react to, not a settled design — the page inventory
in §4 especially.

---

## 1. What D10 actually costs, measured rather than asserted

The owner chose **two audiences, fully separated**. Before planning it I checked
whether any buyer-facing content already exists to build on. It does not:

- **63 tools**, every one a practitioner utility (colour pickers, shadow
  stackers, CSS generators, image optimisers).
- **31 guides**, every one a technical deep dive. The full title list is in the
  DB; a representative span: *The Physics of UI: Why Default Easing Feels Cheap*,
  *Fractional Layouts: The Math of CSS Grid*, *Why 90% of Websites are Vulnerable
  to XSS*, *The Bayesian Truth*, *Understanding Bolt.new*.
- **1 about page, 1 home, 2 section indexes** (`/tools/`, `/learn/`).

```sql
SELECT page_type, build_status, count(*) FROM pages
 WHERE site_id='6b49db8e-d447-4467-8277-4f3018af9897' GROUP BY 1,2;
--  tool 63 deployed | guide 31 deployed | section-index 2 | landing 1 | content 1
--  news-index 1 planned
```

**Not one page addresses a buyer.** So the buyer track is 100% new writing, with
no harvest to improve — unlike the practitioner side, where W2 is a rewrite of
copy that already exists. That asymmetry is the main thing to plan around: the
two halves of W2 are different *kinds* of work, and estimating them as one number
will be wrong.

## 2. The bridge — why this site can do a buyer track others can't

The obvious risk of a buyer section on a tool site is that it is generic advice
anyone could publish. This site has a specific, defensible answer: **the buyer can
use the tools to check what they were sold.**

"Is my site fast enough?", "are these buttons big enough to tap?", "does this
contrast pass?" — those are buyer questions with practitioner tools already built
and live on this domain, 63 of them. A buyer page that ends in *"here is the tool
that answers this, and here is what a good result looks like"* is not generic
advice, and it is the only version of this track worth building.

**This is also the safest form of buyer content**, which matters given §5.

## 3. Naming and placement

Proposal, for the owner to accept or rename:

- **Nav label: `Hire`** — a fifth primary item after Tools, Learn, About, News.
- **URL: `/hire/index.html`**, a new `section-index`, mirroring `/tools/` and
  `/learn/`.

`Hire` is short, unambiguous, and plain. Alternatives considered: *For Buyers*
(reads as jargon and labels the reader), *Getting a Website* (accurate but long in
a nav bar), *Commissioning* (formal, and unclear to the audience least likely to
know the term). **Owner's call — the label is outward-facing.**

The existing nav has room: four primary items at positions 10/20/30/40, so `Hire`
takes position 50 with no reshuffle.

## 4. Proposed page inventory (react to this — it is the least settled part)

Deliberately small. Eight strong pages beat twenty thin ones, and "add rather
than remove" means we can grow it.

| page | the buyer's actual question | bridges to |
|---|---|---|
| `/hire/` index | where do I start? | the whole track |
| Do I need a designer, or can AI build it? | genuinely live question in 2026 | the 5 AI-builder guides already written |
| How to brief a web designer | what do I even ask for? | mood-board guide |
| How to judge a quote | is this expensive or not? | — (see §5 — no figures) |
| Checking what you were given | is this any good? | contrast, 44px, image tools |
| Your legal duties, plainly | accessibility + cookies + UK GDPR | §5 rail applies hard |
| Who owns your site? | domain, hosting, source code, lock-in | — |
| What "done" looks like | handover checklist | performance + a11y tools |

The AI-builder question is the strongest opener: the owner's brief asks for a
renewed focus on AI, site B's half of the merge is exactly that material, and it
is the question a 2026 buyer is actually asking. **It is also the page where a
practitioner-facing library can say something honest that a web agency's own site
cannot.**

## 5. The rails — and the one page most likely to break them

**"How to judge a quote" is the highest-risk page this project has ever
planned.** The natural way to write it is with figures — *"a small business site
in the UK typically costs £X–£Y"* — and we have no such data. This project has
shipped invented statistics **twice** (the tool count reaching eight live specs;
the earlier `{{TOOL_COUNT}}` case), and a price range is exactly the shape that
gets typed because it feels like common knowledge rather than a claim.

**Rail: that page ships with no price figures at all**, or it does not ship. It
can be genuinely useful without them — what a quote should itemise, what a
suspiciously cheap quote has omitted, what is a one-off cost versus a recurring
one, which questions expose a bad fit. All of that is structural, checkable, and
needs no number.

Carried rails, unchanged:

- **No invented statistics, ever.** Counts are substituted from the catalogue
  (`{{TOOL_COUNT}}`); never typed.
- **UK legal and regulatory claims need a primary source or they do not ship** —
  Equality Act 2010, WCAG 2.2, UK GDPR/PECR, ICO guidance, cited to the source
  and not to a summary of it. This is the class the vetcomparison workstream was
  burned by. The "legal duties" page is one long instance of it.
- **D11: no affiliate, no paid placement, and the about page's "sells nothing,
  collects nothing" promise stands.** A buyer track is where the temptation to
  monetise arrives; the answer is already recorded as no.
- **British English throughout**, and the house style at
  `travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT.md`.

## 6. W2 practitioner side — one measured finding to fold in

The rewrite has a concrete, non-speculative starting item. On a UK-focused site,
**American spellings survive the merge in body copy on 23 of 98 pages**:

```sql
SELECT count(DISTINCT p.id) FROM pages p JOIN page_components pc ON pc.page_id=p.id
 WHERE p.site_id='6b49db8e-d447-4467-8277-4f3018af9897'
   AND pc.rendered_html ~* '(optimiz|visualiz|customiz|organiz|analyz|recogniz|minimiz|maximiz|realiz|utiliz|favorite)';
-- 23
```

The pattern deliberately **excludes `color`, `center`, `gray` and `behavior`** —
those are CSS tokens, and including them measures the stylesheet, not the prose.
(`behavior` was checked separately and contributes **zero**: every instance is
`scroll-behavior`.) This is the "grepping a generic CSS property passes" trap from
the travelling-docs workstream, and it is why the number is 23 and not a
meaningless 90-odd.

Three page **titles** also carry them: `learn-marketing-seo-for-llms`
("Optimizing"), `tool-image-optimizer`, `tool-oklch-picker` ("Color Mixer").

**Trap: two of those three are slugs as well as titles.** `tool-image-optimizer`
and `tool-oklch-picker` appear in live URLs. The *displayed title* can be
Britishised freely; **the slug must not be**, or 98 pages of internal links and
every external link break at once. Change the title, leave the URL.

## 7. Sequencing

1. **Now:** finish the news sequence (feed → build page → chrome). In flight.
2. **Owner:** the Cloudflare step (still outstanding — checked 13:05 UTC, no
   beacon on the live home page). Everything about ordering waits on it.
3. **Owner:** accept or rename `Hire`, and react to the §4 inventory.
4. **Then W2 practitioner rewrite** — including the 23-page British English pass,
   which is mechanical and can go first.
5. **Then the buyer track**, written to §4 once the inventory is agreed.
6. **Ordering by popularity — last, and only after stats accumulate against the
   rewritten content.** Unchanged and still right.

## 8. Open for the owner

1. **`Hire` as the nav label** — accept, or a better word?
2. **The §4 page list** — which of the eight earn their place, and what is
   missing? This is the part I am least confident in, because it is the part with
   no existing content to check against.
3. **How far does "no figures" bite?** §5 rules them out of the quote page. If
   you have real UK pricing data from your own work, that changes the page from
   structural advice to genuinely useful — but it has to be *data*, not memory.
