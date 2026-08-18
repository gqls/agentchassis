# 309 — fundamentallyai.com's Platform Log index renders SIX cards with ZERO anchors, so every article it advertises is unreachable from the site

**Filed 2026-08-18** by the 279/284/290 thread, while investigating an owner
observation about the same page. **Status: OPEN. Symptom MEASURED at the served
artefact with a positive control; ROOT CAUSE NOT DIAGNOSED — no `090` run yet, and
this file does not assert a mechanism.** Prior art checked: no bug in `bugs_open/`
or `bugs_closed/` mentions `bl-card`, unlinked cards, or a section index that does
not link its members.

## The symptom, measured at the served page

`https://fundamentallyai.com/platform-log/index.html` (HTTP 200, 32.6 KB) renders
**6 `<article class="bl-card">` cards, and not one contains an `<a>` element**:

| # | card title | anchors |
|---|---|---|
| 1 | Self-Correction: leopardessconsulting.co.uk | 0 |
| 2 | How to Use the LLM Cost Calculator | 0 |
| 3 | How to Use the Review Council Simulator | 0 |
| 4 | How to Use the AI Readiness Checker | 0 |
| 5 | How to Use the Automation Savings Estimator | 0 |
| 6 | How to Use the Model Approach Selector | 0 |

Each card is otherwise complete — image with alt text, category chip, date, read
time, title, excerpt. The page has 31 anchors in total (chrome, footer) and **2 in
`<main>`, both pointing at `/tools/…`**, neither at any article.

**It is not "empty hrefs" and not "phantom links":** there is no anchor element to
carry an href. So `empty_internal_href` (13 rows, page-build-handler) and
`phantom_internal_link` (51 rows) are both the WRONG item type for it, and filing
either would send it to a remit that does not cover it.

⚠ **A measurement correction inside this filing, because it nearly became the
finding:** a first pass on one truncated card string reported "anchors in card: 1".
Re-running the same regex over ALL SIX cards returned 0 for every one. A regex over
a `[:900]` slice is not a measurement of the card.

## The control — the template CAN do this, so it is not fleet-wide

`[MEASURED 2026-08-18]` another site's section index, same card idiom:

| page | cards | cards containing an anchor |
|---|---|---|
| `mortgagecalculator.co.uk/investor/index.html` | 6 | **6** |
| `fundamentallyai.com/platform-log/index.html` | 6 | **0** |
| `idea.uk/news/index.html` | 0 | – (different markup; inconclusive, claims nothing) |
| `vetcomparison.uk/guides/index.html` | 0 | – (same) |

So the card component renders working links elsewhere. Whatever is wrong is about
this page's data, its template variant, or its render — not the component as such.

## Why it matters

Five of the six cards are the tool guides. The site sells the tools; the guides are
the writing that explains them; and **nothing anywhere on the site links any of
them** — checked at `/`, `/platform-log/index.html`, `/tools.html` and
`/capabilities.html`, none of which contains a single `/blog/…-guide` or
`/guides/…` href (the homepage links exactly one blog post, and it is not a guide).
So the guides are reachable only by direct URL or search, while the index that
exists to route readers to them lists them as inert text.

**Card 4 has a second problem behind the first:** "How to Use the AI Readiness
Checker" corresponds to `pages.name = 'ai-readiness-checker-guide'`, whose status is
**`archived`** (its live sibling is `/guides/tool-ai-readiness-checker-guide.html`).
So the index is advertising an archived page, and simply restoring anchors would
give that card a 404 unless it is repointed.

## What this bug is NOT (an owner hypothesis, measured and corrected)

The owner asked about "duplicate guides", believing three tools had two guides each
and that duplication was unintended. **They are not duplicates.** Each pair is two
DIFFERENT articles about the same tool:

| tool | `/blog/…` article | `/guides/tool-…` article |
|---|---|---|
| Automation Savings Estimator | "How to Use the Automation Savings Estimator" | "How the AI Automation Time Savings Estimator Works" |
| Model Approach Selector | "How the Model Approach Selector weighs fine-tuning, RAG…" | "Prompting, RAG, or fine-tuning: a decision guide…" |

A usage guide and a conceptual/decision guide. **Archiving either would destroy real
content, not remove a copy** — so no de-duplication work was filed. Whether the site
WANTS two pieces per tool is a content-strategy question for the owner, not a defect.
Inventory as at 2026-08-18: 7 active guides (4 under `/blog/`, 3 under `/guides/`)
plus 1 archived; only the `/blog/` "How to Use" set appears on the index.

## Candidate mechanisms — all `[UNVERIFIED]`, which is why this is not a fix

1. The listing template variant used by this page emits a title but no wrapping
   anchor (compare against whatever `mortgagecalculator.co.uk/investor/index.html`
   renders with).
2. The listing data carries no resolvable URL per item, and the template omits the
   anchor rather than emitting an empty one — which would make the archived card 4
   the visible tip of a wider resolution failure.
3. A rebuild regenerated this index from a source that lost the per-item hrefs.
   Every guide page's `updated_at` is 2026-08-17, so something rebuilt them all
   recently; the index's own render date has not been established.

**Next step: a `090` diagnosis run** (symptom: the six-card / zero-anchor
measurement above plus the working control; point it at the platform-log-index
page row, `page_components.rendered_html` for that page, and whatever renders
`bl-card`). Check the `needs_diagnosis` queue first.

## How to verify a fix

At the SERVED page, never the stored HTML (this thread's own landmine): all six
cards contain an anchor whose href resolves to an active page — and card 4 points at
the live `/guides/tool-ai-readiness-checker-guide.html`, not the archived
`/blog/ai-readiness-checker-guide.html`. The one-liner that produced the table above
is in this file's history; it counts `<article class="bl-card">` blocks and the
subset containing `<a`.
