# ASSESSMENT 2026-09-04 — the dartsonline guide family, against the selection rule

Full path: `docs/agent_docs/docs024_key_docs_latest/infographics/ASSESSMENT_2026-09-04_dartsonline_guide_family.md`

**Owner ask, relayed by `news_editorial_features` 2026-09-04:** infographics considered for the
**guides** specifically, dartsonline as the live example.
**This is the first application of the lane's selection rule to a real page family.** It is an
assessment, not a dispatch: `dartsonline_traffic` owns that site.

---

## 1. Headline — most of the ask is unblocked today, and needs neither composition nor an evidence base

Two assumptions were in the air and both turn out to be false for this family:

1. *"It needs 035 composition, because a graphic inside `article-body` dies at the next body
   rewrite."* **True for a figure INSIDE prose; false for a graphic BESIDE it.**
2. *"It needs an evidence base, because a figure must resolve through a registered fact."* **True for
   `evidence-chart`; false for the three components this family actually wants**, which have no
   numeric field at all.

## 2. The structural finding: a sibling section needs nothing from composition

`[MEASURED 2026-09-04]` every dartsonline guide has the identical shape:

```
position 1  hero
position 2  article-body      3,287 – 6,160 B, ONE llm-owned `content` blob
position 3  call-to-action
```

**A `comparison-table` inserted at position 3 (CTA → 4) is a sibling, not an inclusion.** It never
touches `content`, so no body rewrite can destroy it — the failure mode that motivated
`inline_guide_imagery`'s whole durability programme simply does not arise. This is the pattern
`finetuning_uk_service` is shipping the same day (mechanism-flow + evidence-chart as new
template-rendered sections, approved copy untouched).

> **So: composition is required for figures interleaved WITH paragraphs. It is NOT required to give a
> guide a table.** The first cut of the owner's ask is reachable now, and `features_open/035`'s
> migration stays demand-driven rather than becoming a blocker for this.

## 3. Per-guide assessment — read the heading structure of all ten

| guide | h2 | the graphic-shaped content | route |
|---|---|---|---|
| `/blog/tungsten-guide` | 5 | **h2s ARE the rows**: "80% — Club Standard" / "90% — Tournament Sweet Spot" / "95% — Maximum Density", each carrying the same four attributes in prose (barrel profile · durability · who it suits · the downside) | **`comparison-table`. Strongest case in the family.** ~2,000 words that are a 3×4 table |
| `tool-dart-weight-comparator` | 4 | explicit `<ul>` of ordered bands (18–20g / 21–23g / 24g+) with a description each; plus the whole guide compares two barrels on weight × tungsten % × length | **`comparison-table`**, twice over |
| `tool-setup-builder` | 5 | 5-item `<ul>` headed *"Where setups usually go wrong"*; opening is causal — *"change one and you've changed how the other two behave"* | **`checklist`** + **`mechanism-flow`** |
| `tool-checkout-calculator` | 5 | *"Two darts left versus three"* (comparison); working-backwards-from-the-finish (sequence) | **`mechanism-flow`**, `comparison-table` secondary |
| `tool-practice-scorer` | 4 | `<ul>` of three metrics, each defined (three-dart average · checkout % · ton count) | **`comparison-table`** or `checklist` |
| `tool-dartboard-zone-visualiser` | 5 | **the copy asks for the graphic in so many words**: *"A scoring chart tells you what each zone is worth. It doesn't tell you how much smaller the treble ring is than the single band next to it."* | **`evidence-chart` — BLOCKED, see §4** |
| `tool-brand-comparator` | **0** | 1 `<ul>`, no headings | comparison-shaped by title; **unpartitionable** |
| `tool-flight-shaft-matcher` | **0** | 1 `<ul>` | **unpartitionable** |
| `tool-leg-average-calculator` | **0** | 1 `<ul>` | **unpartitionable** |
| `tool-tungsten-diameter-visualiser` | **0** | 1 `<ul>` | **unpartitionable** |

**`<table>` count across all ten: ZERO** — on a family whose two strongest pages are explicit
comparisons.

## 4. Why `evidence-chart` is blocked here, and why it doesn't matter much

`[MEASURED 2026-09-04]` **dartsonline.com IS one of the 21 eligible sites** — current `site_plans`
row plus **9** registered facts. But the nine are **PDC tour calendars and news events** (Players
Championship counts, European Tour dates, a nine-darter, an award). **None is about equipment.**

So the zone-area chart — the most visually compelling idea in the family — cannot be drawn from the
register as it stands. Those areas would first have to be registered, and *geometry derived from
published board specifications* is a different provenance question from a tour calendar. **Not
proposed here.**

**This turns out to bound very little, and that is the important part:**

> `comparison-table`, `checklist` and `mechanism-flow` **deliberately have no numeric, price, rating,
> score or rank field** (VIZ-006's principle: the absence of the slot is the control). They are
> therefore the right tools for **structured domain knowledge that is not a registered figure** —
> which is what a darts guide is made of. The tungsten percentages are **row names**, not measured
> claims.

That is why the fleet-wide surface `news_editorial_features` measured (319 of 381 article bodies
carrying a `<ul>`) is real despite almost no site holding an evidence base.

## 5. ⚠ The precedent that constrains all of this

**`dartsonline_traffic` REVERTED grip-styles' per-section imagery** hours after `inline_guide_imagery`
proved the mechanism on it — seven near-identical sections worked against the search traffic that
lane exists to win. **That was their call and the right one.**

**I believe this case differs in kind, and I am recording it as a belief for them to rule on, not as
a finding:** what was reverted was *decomposition into near-identical prose sections*. Replacing
~2,000 words of tungsten prose with one table is the **opposite** move — it reduces duplicated prose
and adds structure, and a real `<table>` is ordinarily friendly to featured snippets. `[UNMEASURED]`
— I have no traffic data and this lane will not assert an SEO claim.

## 6. Proposal — one page, fully reversible

**`/blog/tungsten-guide.html` gains a `comparison-table` sibling section.** Three rows (80/90/95%),
columns: barrel profile · durability · who it suits · watch out for. All four already written in the
prose; nothing invented; no figure that is not a row name.

Why this one: strongest structural fit, no evidence base needed, body untouched, one page, and it
either helps `dartsonline_traffic`'s traffic or it does not — learned on one page rather than ten.

**`dartsonline_traffic`'s call to accept, defer or refuse.** This lane will not dispatch at their
site. If they prefer a page with less traffic at stake, `tool-practice-scorer` is the next-best fit.

## 7. Fed back to `news_editorial_features`

- **6 of 10, not 83%.** Their fleet figure (83% of article bodies carry `<h2>`, h2-partitioning
  byte-lossless on 360/360) is not in dispute; in **this family** it is **6 of 10**, and the other
  four have **zero** `<h2>` and nothing to partition at. Their demand-driven migration will meet
  pages it cannot serve; §2's sibling route works on all ten.
- **`generic-text-block` (262 rows) may be comparable in size to `article-body` (381)**, so the guides
  may not be where this surface mostly lives. Worth measuring before either lane assumes it.
- Requested: their per-domain breakdown of the 319 `<ul>`s, to see whether the surface concentrates in
  guide-shaped sites or spreads thin.

## 8. Open

- `[UNMEASURED]` What the four h2-less guides *are* structurally — one `<ul>` each and no headings may
  mean short, or may mean a different template. Not opened.
- `[UNMEASURED]` Whether a `comparison-table` renders acceptably at 3 rows × 4 columns on mobile;
  VIZ-017 records the stacking behaviour below 40rem as verified by markup inspection only, never on
  a device.
- Registering dartboard zone geometry as facts would unblock the one genuinely numeric idea in the
  family. **A real option, deliberately not proposed** — it is a provenance question (published specs
  vs measurement) that belongs with the claims-verification lane.
