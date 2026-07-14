
<!-- SOURCE: U13_docs024_small_dirs.md -->
### Component asset coupling not enforced (external JS/data file existence is convention only)
- **category:** NEW:component-asset-pipeline
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "data_sources enforcement + inline-small-js_content — component data-file paths are convention, not enforced" (open, unresolved as of 2026-05-21)
- **what:** A component template can reference `<script src="/tools/assets/X.js">` or fetch `/data/X.json` with nothing in the pipeline guaranteeing those files exist or get produced. `content_components.data_sources` (text[]) exists to declare the dependency but isn't consistently populated or validated at deploy time. Two proposed fixes: inline js_content <5KB directly vs. enforce/auto-stub for larger payloads; same pattern applies to data files.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Known-gaps, js_snippets_news_gaswholesalers/old/component_asset_pipeline_concerns.md, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** News rendering three-layer architecture; two news components pattern
- **verify-later:** content_components.data_sources column, git tools/assets/ and data/ paths per site

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Component asset coupling not enforced (external JS/data file existence is convention only)
- **category:** NEW:component-asset-pipeline
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "data_sources enforcement + inline-small-js_content — component data-file paths are convention, not enforced" (open, unresolved as of 2026-05-21)
- **what:** A component template can reference `<script src="/tools/assets/X.js">` or fetch `/data/X.json` with nothing in the pipeline guaranteeing those files exist or get produced. `content_components.data_sources` (text[]) exists to declare the dependency but isn't consistently populated or validated at deploy time. Two proposed fixes: inline js_content <5KB directly vs. enforce/auto-stub for larger payloads; same pattern applies to data files.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Known-gaps, js_snippets_news_gaswholesalers/old/component_asset_pipeline_concerns.md, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** News rendering three-layer architecture; two news components pattern
- **verify-later:** content_components.data_sources column, git tools/assets/ and data/ paths per site
