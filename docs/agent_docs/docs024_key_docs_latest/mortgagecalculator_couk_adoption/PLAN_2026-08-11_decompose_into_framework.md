# PLAN — decompose mortgagecalculator.co.uk into framework management (OWNER-DECIDED 2026-08-11)

Companion to `CONTRIB_2026-08-10_decompose_unmanaged_pages.md`, which holds the measured gap.
This file records the owner's decisions and the execution route. **Written by the
`bugfix_210_needs_logo_unhandleable` lane; not yet executed** — the port deserves a fresh
session, and this lane's session was already long when the decisions landed.

## The owner's decisions (2026-08-11, verbatim intent)

1. **The one-off hero run: "run it now"** — DONE. The rota's design lane was fired once for this
   site (stamp rewound, task enabled for one tick, then restored to `enabled=false` — the 08-10
   cost pause stands). The item filed at 10:09 with a full brand-identity default prompt,
   `prompt_source='default_from_brand_identity'`, and was promoted to `triaged`/`build` **by id**
   (the only detected→triaged promoter, `triage_findings`, is deliberately off — promoting by id
   avoids sweeping the site's 51 other detected items into dispatch).
2. **Decomposition: PORT the pages that have live HTML; REBUILD the never-built ones.**
3. **Show the owner the generated hero when it lands.**
4. **The defaulted-population report line** — DONE and live: the site-discovery staleness report
   now carries `brand images from the DEFAULT prompt: {"total": N, "last_7d": M}`, proven
   rendering with real data the same morning.

## What PORT covers (preserve visible bytes, make managed)

| page | source of faithful bytes |
|---|---|
| `index` | bucket only (`scratchpad/mc_bucket/index.html`, post-backlink-removal, 10,948 B) — **no DB copy exists** |
| `guide-how-banks-decide` | bucket `guides/how-banks-decide.html` |
| `guide-lender-restrictions` | bucket `guides/lender-restrictions.html` |
| `guide-market-structure` | bucket `guides/market-structure.html` |
| `guide-missed-payments` | bucket `guides/missed-payments.html` |
| `guide-mortgage-scorecard` / `scorecard-simulator` | bucket `guides/your-mortgage-scorecard.html` — **one file, two page rows; resolve which row owns it BEFORE porting, or two pages will claim one artefact** |

## What REBUILD covers (nothing to preserve — ordinary framework builds)

`about-index`, `contact-index`, `guides-index`, and whichever of the scorecard pair does NOT own
the live file. All already have `sections` plans; dispatch `needs_page` per page.

## Execution route — REUSE the loancalculator toolchain, do not rewrite it

`docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/` — the same job, done, proven
byte-identical across 27 pages, reversible per page. Read its
`HANDOFF_2026-07-31_continue_here.md` corrections FIRST; every one of them is a trap this site
will also hit. The tooling: `decompose/prepare_work.sh` → `decompose_prover.py` (the splitting
rule — import it, never fork) → `decompose_pages.py` → `assemble_mirror.py` (predicted bytes
BEFORE any row is written) → `load_decomposition.py` (with `--restore <page>` reversibility) →
`verify_shipped.py`.

**The one structural difference from loancalculator:** there, the faithful bytes were already in
`page_components.rendered_html`; here, the six port pages have **zero** components — the
faithful bytes are ONLY in the bucket. So the decomposer's input is the bucket file. Everything
downstream (prover, mirror, loader) is unchanged.

**Site facts that matter, verified 2026-08-11:**
- Chrome: `site_components` has **3 rows** for this site — do NOT recreate chrome; ask whether
  it RESOLVES (the loancalculator lesson: "are there rows" was never the right question).
- The homepage has a **working hero calculator** — page-local `<script>`/`<style>` must survive
  the split (loancalculator correction 4: the splitting rule never read an external
  `<script src>` and a whole results box decomposed as editable prose while the prover passed).
  `--assets` is mandatory.
- `pages.index` is `build_status='needs_rebuild'` with zero components: **flip it out of
  `needs_rebuild` only WITH the port** (the ported components make the rebuild meaningful);
  until then any site-wide rebuild trigger regenerates it from nothing and only the
  `pageHasComponents` guard stands between that and a blank deploy.

## Order

1. Resolve the scorecard page-pair ownership, and the six duplicate deployed paths
   (`guides/x.html` vs `guides/x/index.html` — bugs_open/215's collision shape kills whole
   replan writes). One decision, recorded here.
2. Port `index` alone; `assemble_mirror` proof, then ship, then diff the SERVED bytes
   (origin, not CDN) against the pre-port copy. Byte-identical or explain every byte.
3. Port the four unambiguous guides, one at a time, same proof.
4. Rebuild the never-built pages through the framework (`needs_page` per page).
5. Re-run the placeholder/asset discovery checks against the result.

## Coordination

The facts/tool-acceptance work (`PLAN_2026-08-09_facts_into_tool_acceptance.md`) touches
`tool-*` pages; this plan touches `index`, `guide-*`, and the never-built four — **disjoint page
sets**, so the two can proceed in parallel *provided* neither runs a site-wide replan or rerender.
Check for live sessions before starting (`grep live transcripts, not who-owns` — it reads commits
only).
