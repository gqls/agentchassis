
<!-- SOURCE: U25_leopardess_social.md -->
### Chart component: Go static-SVG emitter + JS progressive enhancement (A5/A7)
- **category:** NEW:data-charts
- **status-signal:** aspirational
- **status-evidence:** PLAN phase table: "L7 Charts (Go SVG + JS) — not started"; H2 "resolved 2026-07-10 — confirmed, data layer first".
- **what:** The reusable data-chart capability honouring prior imagery decisions: D1 (charts are code-rendered from real data — the LLM proposes the story, code owns the numbers; diffusion never plots data), D3 (chart is a Lane-B asset, deliberately NOT a site_plan_imagery kind). Resolves the recorded conflict between "go-echarts in-chassis" (confirmed 2026-07-08) and "static SVG must always exist": a dependency-free Go SVG emitter produces the accessible static artifact (axes, caption, source line, query date); an inline self-contained JS renderer progressively enhances it (no CDN). First charts use provable in-DB numbers. Explicitly excludes the data-chart-generator agent and external data APIs (deferred to imagery Phase I4).
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#5; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-4; docs/leopardessconsulting/RUNBOOK.md#H2
- **relations:** imagery programme I4 / FUTURE_data_graph_pipeline (docs024); imagery kind routing (charts kept out of diffusion)
- **verify-later:** go.mod for charting deps (none expected); any SVG emitter under the chassis

<!-- SOURCE: U25_leopardess_social.md -->
### Chart component: Go static-SVG emitter + JS progressive enhancement (A5/A7)
- **category:** NEW:data-charts
- **status-signal:** aspirational
- **status-evidence:** PLAN phase table: "L7 Charts (Go SVG + JS) — not started"; H2 "resolved 2026-07-10 — confirmed, data layer first".
- **what:** The reusable data-chart capability honouring prior imagery decisions: D1 (charts are code-rendered from real data — the LLM proposes the story, code owns the numbers; diffusion never plots data), D3 (chart is a Lane-B asset, deliberately NOT a site_plan_imagery kind). Resolves the recorded conflict between "go-echarts in-chassis" (confirmed 2026-07-08) and "static SVG must always exist": a dependency-free Go SVG emitter produces the accessible static artifact (axes, caption, source line, query date); an inline self-contained JS renderer progressively enhances it (no CDN). First charts use provable in-DB numbers. Explicitly excludes the data-chart-generator agent and external data APIs (deferred to imagery Phase I4).
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#5; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-4; docs/leopardessconsulting/RUNBOOK.md#H2
- **relations:** imagery programme I4 / FUTURE_data_graph_pipeline (docs024); imagery kind routing (charts kept out of diffusion)
- **verify-later:** go.mod for charting deps (none expected); any SVG emitter under the chassis
