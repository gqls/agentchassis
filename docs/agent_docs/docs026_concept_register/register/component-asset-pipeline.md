# Register — component-asset-pipeline

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 1 raw extraction across unit U13.

> Note on source duplication: this block appeared twice, byte-identical, in the cluster input file (the same mechanical whole-file duplication documented in register/imagery.md's header) — collapsed to one raw extraction before consolidation.

### CAP-001 — Component asset coupling not enforced (external JS/data file existence is convention only)
- **status:** aspirational
- **status-evidence:** TODO_remaining_work.md: "data_sources enforcement + inline-small-js_content — component data-file paths are convention, not enforced" (open, unresolved as of 2026-05-21).
- **what:** A component template can reference an external file — `<script src="/tools/assets/X.js">` or a `fetch('/data/X.json')` — with nothing in the pipeline guaranteeing that file actually exists or gets produced at deploy time. `content_components.data_sources` (a `text[]` column) exists specifically to declare this dependency, but it isn't consistently populated or validated when a component deploys, so a broken reference only surfaces as a runtime 404 in the browser, not a build-time failure. Two proposed fixes are on record: inline `js_content` directly into the page when the payload is small (<5KB), versus enforce-or-auto-stub the dependency for larger payloads; the same unresolved-convention problem applies equally to referenced data files.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Known-gaps; js_snippets_news_gaswholesalers/old/component_asset_pipeline_concerns.md; js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** News rendering three-layer architecture (cross-category); two news components pattern (cross-category)
- **verify-later:** content_components.data_sources column population/validation; presence of referenced files under git tools/assets/ and data/ paths per site
