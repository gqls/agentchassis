# Register — component-asset-pipeline

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

_Concept count retired 2026-08-09 — derived, not stored; run the drift pair in `000_concept_index.md`, or read `concept-register-drift-check`'s daily row (DOC-074)._ consolidated from 1 raw extraction across unit U13.

> Note on source duplication: this block appeared twice, byte-identical, in the cluster input file (the same mechanical whole-file duplication documented in register/imagery.md's header) — collapsed to one raw extraction before consolidation.

### CAP-001 — Component asset coupling not enforced (external JS/data file existence is convention only)
- **status:** aspirational
- **status-evidence:** TODO_remaining_work.md: "data_sources enforcement + inline-small-js_content — component data-file paths are convention, not enforced" (open, unresolved as of 2026-05-21).
- **what:** A component template can reference an external file — `<script src="/tools/assets/X.js">` or a `fetch('/data/X.json')` — with nothing in the pipeline guaranteeing that file actually exists or gets produced at deploy time. `content_components.data_sources` (a `text[]` column) exists specifically to declare this dependency, but it isn't consistently populated or validated when a component deploys, so a broken reference only surfaces as a runtime 404 in the browser, not a build-time failure. Two proposed fixes are on record: inline `js_content` directly into the page when the payload is small (<5KB), versus enforce-or-auto-stub the dependency for larger payloads; the same unresolved-convention problem applies equally to referenced data files.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Known-gaps; js_snippets_news_gaswholesalers/old/component_asset_pipeline_concerns.md; js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** News rendering three-layer architecture (cross-category); two news components pattern (cross-category)
- **verify-later:** content_components.data_sources column population/validation; presence of referenced files under git tools/assets/ and data/ paths per site

### CAP-002 — `ActionInputs.WasDefaulted`: a caller can now tell a spec DEFAULT from a caller's choice
- **status:** built, unit-tested (4 cases incl. the nil-map and no-default arms), mutation-tested, council-submitted `7f0c1535-25cb-4645-adba-f7429e357a79` — live in code at commit `930ace3bd`, INERT until the next chassis roll. One caller today: `deploy_image_asset`.
- **what:** `platform/orchestration/datahelpers.ActionInputs` gained a `Defaulted map[string]bool` and a `WasDefaulted(key) bool` accessor. `ExtractActionInputs` marks every field it fills from `spec.Defaults` and clears the mark in Strategy 0 — the only strategy that can overwrite a default. **Additive: no value in `Values` changes, so ignoring it preserves existing behaviour exactly.**
- **why it exists (the landmine it makes visible):** defaults are written into `Values` **before** Strategy 1/2 run, and every later strategy skips a field that already holds a value. So **a field WITH a default can only ever be set by a Strategy 0 explicit dot-path that resolves** — the recursive search that finds a nested `spec.<field>` for every other field never runs for it. The consequence is not a wrong value but an **unattributable** one: a consumer reads `"hero"` and cannot tell whether a caller asked for it. `deploy_image_asset` shipped every work-item-dispatched asset — logos and favicons included — with hero geometry and a `.jpg` extension for four months on exactly that (`bugs_open/248` finding (b)).
- **the bar for using it:** call it before treating a value as an *instruction*. A defaulted value is a fallback, not a statement. Where an authoritative source exists (for `deploy_image_asset`, the `assets` row's own `purpose`/`asset_key`), prefer it over a default — but never over a value the caller actually stated.
- **sources:** `platform/orchestration/datahelpers/action_inputs.go` (`ActionInputs.Defaulted`, `WasDefaulted`, Strategy 0's `delete`); `platform/orchestration/datahelpers/action_inputs_defaulted_test.go`; `bugs_open/248`; `platform/orchestration/actions/deploy_image_asset_action.go` (the one consumer).
- **relations:** `bugs_open/231` (a static config value for a spec-defaulted field is dead — the same mechanism, different field, and the reason this is a class rather than one action's bug); `bugs_open/248`; CAP-001 (sibling: both are "the pipeline does not enforce what it assumes").
- **verify-later:** whether any OTHER action with `spec.Defaults` is silently acting on a default it believes was chosen — the population is every `ActionInputSpec` with a non-empty `Defaults` map, and nobody has enumerated it.
