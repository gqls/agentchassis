# PROPOSAL — decompose webdesign.co.uk's 63 ported tools into first-class framework components (owner decision)

Raised by the owner alongside bugs_open/281: *"please consider decomposing all the tools so
we can manage them in the framework properly."* Researched 2026-08-15; NOT executed. This is
the recommendation with its preconditions, so the decision can be made on evidence.

## What "decomposed" means here

Today each ported tool is ONE `page_components` row on a `page_type='tool'` page, pointing at
the shared section-level `ported-page` component (`a7daa5c5…`); the whole tool (markup, CSS,
JS) lives in that row's `rendered_html`; `content_data` holds only provenance (no fields), so
nothing in the framework can edit it through the content path. Decomposed means: one
`content_components` row per tool with the machinery in `html_template` and the copy in
`input_schema` fields, the page's `page_components` replaced by one row per block, the tool
row `component_level='tool'` — the shape the framework already understands.

## What already exists (do not rewrite it)

- The B2 conversion chain in `docs024_key_docs_latest/loanandmortgagecalculator_couk/`
  (`b2_build.py` → `b2_load.py` → `b2_verify.py`) did exactly this for 23 calculators on
  2026-08-15 — byte-identical through Go's own template engine, 154 editable copy fields.
- The loancalculator lane's `decompose/` toolchain (ADO-039/040/041), and its
  `rewrite/load_components.py`, which DOES set `component_level='tool'`.
- The markup-preservation gate `scripts/class_count_delta.py` (ADO-041) — the gate any
  conversion must run; NOT `gate_wrapper_parity.py`, which does not strip script/style and
  false-drops on script-bearing pages (i.e. every webdesign tool).
- `cmd/webdesignport` is idempotent and lock-respecting, so the pre-state is recoverable by
  re-running the importer.

Webdesign is EASIER than the calculator lanes in one respect: its ports go through assembly
(no `deploy_mode='verbatim'`), so ADO-039's row-count landmine ("adding a row beside a
verbatim row is catastrophic") does not apply.

## Preconditions — each one is a reason not to run it today

1. **bugs_open/204 first (open, unowned).** `plan_sections` resolves sections by name/function
   only, so positional slots (`prose-0`, `tool-1`) never resolve and a decomposed page cannot
   be REBUILT; the selector then files `needs_new_component` junk (114 items on one site).
   Today `pages.sections=["ported-page"]` resolves. Converting the 63 without 204 turns 63
   rebuildable pages into 63 unrebuildable ones.
2. **Set `component_level='tool'` on the tool row.** B2's INSERT omits it and the column
   defaults to `'section'` — measured live 2026-08-15: only **1 of 18** LMC tool pages has a
   tool-level component after B2. A decomposed tool that is not marked `tool` fails eligibility
   clause (a) AND clause (b) (the page now has prose rows beside the tool row, so it is no longer
   sole-component) — i.e. decomposition as run this week REDUCES audit coverage. Handed to the
   B2 lane as a finding.
3. **`forked_from` / unique index.** `idx_cc_tool_function_unique` claims a `function`
   fleet-wide when `forked_from IS NULL`. Per-site tool rows must carry `forked_from` (a library
   ancestor, or a deliberate library row) or take a globally unique function.
4. **PLAN subject keys.** Ported tools are keyed `p.name` minus `tool-` (e.g. `animated-favicon`);
   a real tool component is keyed by `cc.function`, which by convention KEEPS the prefix. The 13
   existing criteria fences and their notes must be re-keyed or the function chosen as the bare
   slug — decide before the first conversion, or the fences go silent.
5. **13 tools keep logic in external `<script src>` files** (TL-032) — invisible to every
   DB-side splitting rule; the prover has been blind on exactly this shape before (ADO-039).
6. **Wrapper vocabulary.** 8+ wrapper idioms on 22 calculator pages (bugs_open/263); webdesign's
   tools came from a hand-built site with no shared idiom — expect more, and expect the
   `keep_widget_wrapper` opt-in to need per-tool judgement.
7. **Ownership.** No lane owns webdesign's tools today (`webdesign_tools_repair` dormant since
   07-31); the decomposition toolchains are ACTIVELY worked (loancalculator, LMC B2) — any change
   to `split_ordered` / `b2_*.py` lands under a live session.

## Recommendation

Do it, as its own owned lane, in this order: (204) → one pilot tool with a PLAN and an
external-script-free body, converted with `component_level='tool'` + `forked_from` decided,
verified through `class_count_delta.py` and by a Tier-1/2/4 pass drawing items under the
fork's `function` key → then the batch. Track 1 (bugs_open/281) works on the ported tools as
they stand and remains correct after decomposition — the audit is not waiting on this.

Owner questions: (a) proceed as a lane, and who owns it; (b) function naming convention for
promoted tools (`tool-<slug>` vs bare slug) — it decides the PLAN re-keying; (c) whether the
promoted rows are per-site forks or deliberate library rows.
