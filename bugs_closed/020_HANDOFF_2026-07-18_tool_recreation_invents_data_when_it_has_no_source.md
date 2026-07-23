# 020 — tool-recreation invents a dataset when the original tool was data-backed

> # ✅ CLOSED 2026-07-23 — FIXED, LIVE, and VERIFIED end-to-end.
> The mechanical fabrication gate is live on chassis **v1.0.1150** and **wired** into
> tool-recreation-handler (migration 189), backed by the **fixed** detector
> (commit `1a2718213`). Proven by an **induced-fault probe** on the live build: a stubbed
> fabrication driven through the wired path was HELD at `needs_human_review` via REAL
> detection (`tier:"declaration"`, signals `["realistic, deterministic dataset",
> "makePostcode"]`) and **never deployed**. Plus the live prompt contract (migration 183).
> Two-part fix, both live. The tool-imagery HOLD (`3f6f1febf`) can lift. Full journey below
> and in `docs/.../bug020_tool_recreation_data_integrity/`.

*Found 2026-07-18 by the vetcomparison thread. **Live fabrication shipped to a public site and
served to visitors.** Contained on the affected site; **the platform defect is now FIXED & LIVE
(2026-07-23) — see the CLOSED banner above.***

> **STATUS 2026-07-21 (bugfix 020 thread).** The platform fix is **half live, half built-and-inert.**
> - **Prompt half — LIVE NOW** (migration `183`, DB config, no image roll). `recreate_tool` gained a
>   prominent `## Data Integrity` section (self-contained = code not data; preserve the original
>   fetch source; honest empty state, never invent records) and rule 9 was rewritten to bind DATA
>   not just arithmetic; `analyze_tool` now captures the original tool's `data_source` verbatim.
>   This is candidates (1)+(2). It reduces the rate; it does not, alone, guarantee obedience.
> - **Mechanical gate — BUILT + UNIT-TESTED + COMMITTED, inert until an image roll** (candidate 3).
>   New Go action `check_tool_fabrication` (`platform/orchestration/actions/`) + 12 passing tests
>   (incl. **fail-SAFE** on un-inspectable output — a council find); workflow wiring staged
>   **image-first** (now needle-gated) at
>   `docs/.../bug020_tool_recreation_data_integrity/WIRING_check_tool_fabrication_APPLY_AFTER_IMAGE.sql`.
>   Council reviewed (`SUBMISSION_CORR 8eef369f`): **2× REVISE, all substantive objections addressed**
>   (wiring folded in-plan; fail-open→fail-safe; paths verified; needle-gate + RETURNING). No APPROVED
>   obtained → **no `Council-Reviewed:` trailer** (earned by APPROVED only). Loop stopped after round 4.
> - **UPDATE 2026-07-22 (2nd roll) — the induced-fault caught a REAL BUG; gate UNWIRED again.**
>   v1.0.1146 wired via migration 189, then v1.0.1149 rolled WITH the fail-safe (`uninspectable`
>   grep = 1). Re-wired and ran an **induced-fault probe** (scratch agent, live gate on a stubbed
>   fabrication). It reached the HELD terminal — but via `tier:"uninspectable"`, NOT `declaration`:
>   the fabrication content never reached the detector. **Root cause:** the action read `html_field`
>   via `datahelpers.ExtractActionInputs`, whose Strategy 0 **resolves dotted config values against
>   collected_data** — so `"completeness_check.clean_html"` became the HTML *content*, and extracting
>   again with that as a path returned `""` → empty → uninspectable. On the fail-OPEN detector this is
>   a **silent no-op (deploys everything, incl. real fabrications)**; the fail-safe turned it into a
>   loud over-HOLD, which is how it was caught. **Fixed** (`1a2718213`): read config paths directly
>   like `check_tool_completeness` does; new wrapper regression test. **Gate UNWIRED live** (snapshot
>   taken; `next_step` back to `save_training_data`, gate steps removed) so recreations are not
>   over-held meanwhile — the Half-A prompt fix still protects. `WRONG_CALLS.md` logged.
> - **✅ CLOSED 2026-07-23 (3rd roll).** v1.0.1150 carries the fix `1a2718213` (gate symbols
>   pod-verified). Re-wired via migration 189 (routing independently verified). Re-ran the
>   induced-fault probe: `terminal=complete_held`, `fabrication_check.tier="declaration"` (REAL
>   detection this time, not uninspectable), signals `["realistic, deterministic dataset",
>   "makePostcode"]`, `needs_human_review` item raised, **never deployed**. That is the end-to-end
>   proof the fix works on the live wired path. Scratch agent + probe item cleaned up. **Lesson kept:**
>   static "paths verified" ≠ correctness — the action was proven only by inducing the fault, and the
>   layer that RUNS (the wrapper) is now tested, not only the pure detector.
>   Workstream docs: `docs/agent_docs/docs024_key_docs_latest/bug020_tool_recreation_data_integrity/`.

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

## Containment applied (site side — the platform defect above is still unfixed)

Both layers, 2026-07-18:

1. **Published file** — verified homepage restored from `b2896815`, pushed, live-verified clean
   (0 generator symbols, 0 unsupported claims).
2. **Source** — `page_components.rendered_html` still held the fabrication, `deployed` and
   unlocked, so the next render would have republished it. Note *where*: the generator was in the
   **`hero`** slot (18,101 chars — the whole recreated tool), **not** `filtered-result-grid`.
   Hero's data layer rewritten to `fetch('/data/vet-full-index.json')` keeping the chassis's UI
   (region filter, pagination); demo-sample disclaimer, price-sort options, "pricing information /
   ownership data" claims and a false about-page differentiator removed. Four components set
   `lock_type='permanent'` (index: hero, filtered-result-grid, info-card-grid; about:
   differentiators).

> **UNVERIFIED (as at 2026-07-19):** no render has run against the corrected source — nothing has
> touched `vetcomparison.uk/` since the restore, and manual dispatch failed (`rerender-pages` is
> `experimental`; neither `system.agent.site-builder.requests` nor
> `system.agent.page-rerender.process` produced an orchestration state from kcat). **Whoever sees
> the first render must run the greps in "How to verify a fix" above.** If fabrication returns,
> the permanent locks did not hold — record that here, it changes the fix ranking.

Site-side containment does not fix the defect: any other adopted site with a data-backed tool is
still exposed. See `docs/agent_docs/docs024_key_docs_latest/vetcomparison/HANDOFF_2026-07-19_vetcomparison_uk.md`.

## Note on numbering

`016` is currently used by two unrelated cases (council-revise-prompts and ssh-HOME) — concurrent
threads collided. This file takes `020`; highest existing was `019`.

---

## Addendum 2026-07-21 — the `permanent` locks do NOT survive a full page rebuild

The remediation relied on `page_components.lock_type='permanent'` (set by
`vetcomparison-fabrication-fix-bug020`) to protect the corrected components. On
2026-07-21 08:08 a normal full-page render ran on vetcomparison.uk's index and
**stripped every lock** — all five components came back with `lock_type` NULL.

Verified it is a delete-and-recreate, not an in-place clear: the `hero` row's primary
key changed across the render (`c8df695e-…` → `24060593-…`). So the rebuild replaces
the component rows wholesale; a lock living on the old row cannot survive, by
construction. **A `permanent` lock protects against an in-place overwrite, not against
a rebuild that re-materialises the page's components from its manifest.**

The saving grace this time: the rebuild regenerated `hero` from a clean source — the
real `fetch('/data/vet-full-index.json')` is intact and there is **zero** fabrication
(`Mulberry32`/`makePostcode`/`representative sample` all absent from the live HTML).
So the anticipated test — the original handoff said "if fabrication returns, the locks
did not hold" — fired with a benign outcome: the locks did **not** hold, but fabrication
did not return either. The exposure is therefore latent: the day the tool-recreation
path (this bug's actual subject) runs as part of a rebuild, the locks that were meant to
stop it will already be gone.

Consequence for any lock-based mitigation on this platform: **do not rely on
`lock_type` to survive a rebuild.** Protection has to live in the component's data
source or the generator's contract, not in a per-row flag the rebuild discards.
