# 020 — tool-recreation invents a dataset when the original tool was data-backed

*Found 2026-07-18 by the vetcomparison thread. **Live fabrication shipped to a public site and
served to visitors.** Contained on the affected site; the platform defect is unfixed.*

**Family:** same class as `001` (a rebuild resurrecting fabrication that had been audited out of
a site days earlier) but a **different mechanism** — 001 is `reconcilePlanWithRealised` no-op
during re-plan; this is the tool-recreation path having no concept of a data dependency. Read 001
first; do not merge them, fix both.

## What happened

`vetcomparison.uk` is a veterinary price-comparison site whose entire remediation history is
about *not publishing figures we cannot source* (997 fabricated price rows quarantined, a legal
factual record written, see `docs/agent_docs/docs024_key_docs_latest/vetcomparison/LEGAL_2026-07-15_*`).

Its homepage carried a hand-built practice directory: a search/filter widget reading
`/data/vet-full-index.json` (2,109 real verified practices, exported from
`business_intel.businesses`).

The site was adopted onto the chassis on 2026-07-17. The build cascade raised
`needs_tool_recreation`, handled by **`tool-recreation-handler`**, which recreated the widget as
a self-contained component that **generates synthetic practices in the browser**. Its own
comment, shipped to production:

> `// The original directory holds 2,100+ UK practices. For this recreation we`
> `// generate a large, realistic, deterministic dataset so search, filtering`
> `// and pagination behave exactly as they would against the real directory.`

Implementation: `TOWNS` / `PREFIXES` / `SUFFIXES` arrays crossed by a Mulberry32 seeded RNG
(`const rng = () => {...}`), `makePostcode()` inventing postcodes like `SW4 3PL`, `buildData()`
assembling the list, then `render()`. Fake practice names ("Abbey Veterinary Centre", "Oakwood
Vets") with invented postcodes were served to live visitors as a UK practice directory.

The same rebuild also emitted copy the site cannot support: "pricing information, ownership
data" (neither is published), `Price: Low to High` sort controls (no published prices), and a
disclaimer calling the real 2,109-practice directory "a representative sample for demonstration
and comparison purposes".

**Every work item reported `complete`.** `needs_tool_recreation`: complete. All `page_rerender`:
complete. Nothing failed, nothing warned.

## Root cause

Two defects compound:

**(a) The recreation path has no data-dependency contract.** `tool-recreation-handler`'s prompt
asks for self-contained HTML/CSS/JS for an interactive tool. There is no slot for "this tool
reads from <URL//path>", and adoption's `extract_interactive_fingerprint` step does not carry a
captured `fetch()` target through to it. A tool whose behaviour *is* its data therefore cannot be
recreated faithfully — the model must either emit a dead empty widget or invent records to make
search/filter/pagination demonstrably "work". It chose the latter, and said so in a comment.

**(b) The prohibition is scoped to arithmetic, not data.** The prompt's rule 9 reads:

> `9. No fake data or dummy outputs — calculations must be mathematically correct`

Read in context (rules 7–10 are all about completeness and correctness of *functions*), this is a
statement about calculators. It does not tell the model that inventing *records* is forbidden,
and the model evidently did not read it that way.

## Fix candidates

1. **Carry the data dependency through adoption.** `extract_interactive_fingerprint` should
   record each tool's data sources (fetch/XHR URLs, referenced `/data/*.json`), and
   `tool-recreation-handler` should receive them as a required contract: *recreate the widget
   against THIS source; do not embed data*.
2. **Rewrite rule 9 so it binds data, not just maths**, e.g.: *"Never generate, synthesise, seed
   or hard-code example records. If the tool needs data and you have not been given a source,
   render an empty state and stop — do not invent records."* Worth applying to every generative
   prompt that can emit list-shaped content, not just this one.
3. **Add a post-generation fabrication check** beside the existing completeness-marker check
   (the workflow already has `check_completeness`). Grep generated JS for data-invention tells —
   seeded PRNG (`Mulberry32`, `imul`, `seed`), name-fragment arrays crossed to build labels,
   `buildData`/`generate*`, literal record arrays over N entries — and fail the item to
   `needs_human_review` rather than deploying.
4. **Hard gate on audited-content sites.** Some sites (vetcomparison, leopardess) have an
   explicit no-unsourced-claims policy. That policy should be a machine-readable site flag that
   blocks *any* generated content deploy which introduces records or statistics, rather than
   living only in documentation a generative agent never reads.

Prefer (1)+(2) structurally; (3) is the cheap net that catches the next variant.

## How to verify a fix

- Re-run tool recreation against a data-backed tool and confirm the output **fetches** rather
  than embeds; no PRNG, no fragment arrays, empty state when the source is unreachable.
- Grep the rendered page, not the work item: `curl <page> | grep -iE 'Mulberry32|makePostcode|
  buildData|SUFFIXES'` must return nothing. **`complete` is not evidence** (CLAUDE.md).
- Regression case: adopt a site whose homepage tool reads a JSON file; assert the deployed tool
  requests that file.

## Containment already applied (site side, not a platform fix)

`sites` repo: verified homepage restored from `b2896815` (real directory, claim + opt-out
routes), pushed and live-verified clean — 0 generator symbols, 0 unsupported claims. The hand
edit takes a permanent lock, which protects *this* page only.
**Still exposed:** the `index` page spec still lists a `filtered-result-grid` section with no
data source, so a future render can regenerate this. Spec-level fix outstanding — see
`docs/agent_docs/docs024_key_docs_latest/vetcomparison/HANDOFF_2026-07-18_vetcomparison_uk.md`.

## Note on numbering

`016` is currently used by two unrelated cases (council-revise-prompts and ssh-HOME) — concurrent
threads collided. This file takes `020`; highest existing was `019`.
