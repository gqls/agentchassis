# Register — data-charts

1 concept, consolidated from 1 raw extraction across unit U25.

> Note on source duplication: this block appeared twice, byte-identical, in the cluster input file (the same mechanical whole-file duplication documented in register/imagery.md's header) — collapsed to one raw extraction before consolidation.

### CHRT-001 — Chart component: Go static-SVG emitter + JS progressive enhancement (A5/A7)
- **status:** aspirational
- **status-evidence:** PLAN phase table: "L7 Charts (Go SVG + JS) — not started"; H2 "resolved 2026-07-10 — confirmed, data layer first."
- **what:** The reusable data-chart capability honouring prior imagery decisions from the parent imagery best-in-class programme: D1 (charts are code-rendered from real data — the LLM may propose the story, but code owns the numbers; diffusion never plots data) and D3 (a chart is a Lane-B, content-driven asset, deliberately NOT added to the `site_plan_imagery` kind enum). The plan resolves a recorded tension between "go-echarts in-chassis" (confirmed as the rendering runtime, 2026-07-08) and "a static SVG/PNG must always exist as a fallback": a dependency-free Go SVG emitter produces the accessible static artifact (axes, caption, source line, query date), while an inline, self-contained JS renderer progressively enhances it — no CDN dependency, consistent with this codebase's no-external-script constraint on artifacts/templates. First charts would use provable in-DB numbers. Explicitly excludes, for now, a dedicated `data-chart-generator` agent and any external data-source APIs — those are deferred to the imagery programme's Phase I4 (see register/imagery.md IMG-046, the data-graph/chart pipeline concept, which states this same D1/D3 constraint from the imagery side and is the direct counterpart of this entry).
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#5; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-4; docs/leopardessconsulting/RUNBOOK.md#H2
- **relations:** imagery programme Phase I4 / FUTURE_data_graph_pipeline (register/imagery.md IMG-046); imagery kind→provider routing (register/imagery.md IMG-030 — charts are deliberately kept out of the diffusion/kind system this describes)
- **verify-later:** go.mod for charting dependencies (none expected yet); any SVG emitter package under the chassis; whether a data-chart-generator agent has since been scaffolded
