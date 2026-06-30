# RUNNING NOTES — scheme reaches components (dedicated P0 thread)

The journal for the dedicated "make a site's scheme reach its components" thread. Memory is OFF; this doc is the journal. **Present this file at the END OF EVERY TURN.** Continues from the parent chat's `running_notes_2.md` checkpoint (qqq); the parent will resume its own notes once this fix is done. Companions: `PLAN_scheme_to_components.md`, `RUNBOOK_scheme_to_components.md`, and (source) `REPORT_scheme_does_not_reach_components.md` + `HANDOFF_scheme_to_components.md`.

═══════════════════════════════════════════════════════════════════════
CARRY-OVER STATE
═══════════════════════════════════════════════════════════════════════

## Standing preferences (STRICT)
Go not Python; plain human language, no LLM-hype, no flattery; confirm live API/schema/data facts before asserting or coding (`0 rows` not decisive until the query/state is checked); reuse and adapt before rebuild; fix the framework structurally over one-offs; honest caveats and pushback, including correcting my own reads; British English; low risk appetite; reasonable steps; ≤1 question per reply; don't create summary docs unless asked; minimal formatting (prose over bullets) in replies; banned words "perfect"/"critical"/"excellent"; no `logger.Debug` (it doesn't show — use `logger.Info`); don't call a fix "final"/"last"; no congratulation. Keep the runbook + this journal current. SQL run as FILES via `kubectl … < file`.

## Architecture conventions
Every agent is an orchestrator owning a workflow of steps that call actions. Keep workflows simple; put complexity in Go actions. Don't build sub-workflows in SQL — spawn sub-agents with their own workflows. Agents respond to the caller's (parent's) responses topic. Keep workflow variable names in sync with what the actions expect. Check DB schemas before writing SQL.

## Project facts
- **idea.uk** — LIVE Go service selling £29 reports; single binary under systemd on a Hetzner VM, nginx + LetsEncrypt, 127.0.0.1:8080; orders in a file (no DB); live Stripe webhook. Reserved tool paths: `/request /confirm /approve /decline /stripe/webhook /internal/* /order/*`. DNS (Cloudflare) → the VM, so chassis B2 deploys are invisible to the live site. UNCHANGED.
- **Chassis website-builder** — multi-agent Go/Kafka/Postgres in k8s; domain → multipage site → static → Backblaze B2 (github → GH Actions → B2).
- idea.uk chassis staging site_id `1244516d-014d-421c-88c6-090bb1e9552a`. (Note: the parent runbook's general-commands block also shows `97ed2f64-…`; treat as a separate/earlier row and confirm before relying on either.)
- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`. k8s: `-n ai-persona-system`, `-n kafka`; cluster `personae-kafka-cluster`.

## The fix in one line
Make the site's scheme reach its components as a variable-value override consumed by one component (palette + `--section-*`), not `*-light`/`*-dark` duplication; framework-level; design before code; migrate + back-fill without regressing dark sites.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sa) — bundle 1 assembled and read; two setup issues
═══════════════════════════════════════════════════════════════════════
Built bundle 1 for investigations A (render path) + B (scheme signal) + G (`--section-*` mechanism), scoped to settle the gating Q4, via the `cmd/bundle` read-only composer (analyser + assembler + dbcontext). Two issues with the run, recorded so the next bundle is clean:
1. The two `-doc` paths (`docs/agent_docs/docs024_key_docs_latest/003…`, `…002…`) did not resolve — the bundle recorded `no such file or directory`. The standards are NOT in the bundle; read 002/003 from project knowledge instead. Fix the path for bundle 2 (`ls` the docs024 dir first).
2. The in-scope code came through as SIGNATURES ONLY, no function bodies — consistent with `-step framing` emitting structure rather than source (the parent example used `-step debug`); a wrong `-root` could also do it. Use `-step debug` next. The six schemas and the runtime evidence came through fully; the `render_css_*` bodies were the gap and were supplied separately afterward.
Runtime evidence (idea.uk/index): the only recent errors are 2026-06-21 `write_site_spec` "missing required fields: [spec_data]" on `persist_mission`/`persist_roadmap`/`persist_roadmap_brief` (a separate spec-write issue, not the scheme problem); the page's work-item history is three completed `page_rerender`/`needs_page` items from 2026-06-21 (the stale dark build).
No code/DB changes. idea.uk untouched (staging).

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sb) — investigation B settled: where the scheme signal stops
═══════════════════════════════════════════════════════════════════════
Traced scheme end to end from the bodies:
- `deriveSchemeFromDesignIntent(style_direction, suggested_style)` → `light`/`dark`/`""` (substring match: dark; or light/warm/editorial/paper/bright/soft → light).
- `resolveLayoutByTags` uses scheme as a near-hard constraint to pick the layout.
- `buildResolvedCompositionSpec` records `css_theme_id`/`palette_id`/`layout_id`/`typography_set_id` (+ names, lineage, reasoning, resolved_by/at) — **NOT the scheme value.** So scheme survives downstream only on `layouts.scheme` (schema-confirmed; check constraint `light`/`dark`/`neutral`; no scheme column on `style_collections`).
- It reaches `styles.css` (palette `:root` + the section mechanism, see Sc) but **stops before the component render context**: `RenderContext` (component_library.go) has palette-colour fields (`PrimaryColor`/`AccentColor`/`TextColor`/`BackgroundColor`) and content but **no scheme field and no `--section-*`**; `BuildRenderContextAction` merges sources into that struct, so a source carrying "scheme" has nowhere to land; `contextToInterfaceMap` exposes only struct fields. So a component template (`RenderTemplate(template, ctx)`) gets colour values but no light/dark signal — which is exactly why dark components hardcode their darkness inline.
Injection point (Q1/Q3 lean): carry the base scheme as an explicit value into the render context (add `RenderContext.Scheme`, populated from `layouts.scheme` via `sites.style_collection_id → style_collections.css_theme_id → css_themes.layout_id → layouts.scheme`) and into the CSS loader (add `l.scheme` to its SELECT + a `themeComposition.Scheme`), and let the renderer own per-section `--section-*`. Reuse, not a parallel mechanism.
Aside (standing-pref violation in existing code, noted not actioned): `BuildRenderContextAction` uses `logger.Debug` for its source-merge tracing — those lines won't show, which will hamper debugging the very plumbing we're about to add. Worth switching to `logger.Info` when we touch it.

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sc) — investigation G settled: the CSS render path + the class-name contract
═══════════════════════════════════════════════════════════════════════
Read `render_css_from_spec_action.go` + loader + helpers in full. `RenderCSSFromSpecAction` assembles `styles.css` in three appended parts:
1. The layout `css_template` rendered as a Go `text/template` with FuncMap helpers `{{palette}}`/`{{typo}}`/`{{token}}` over merged maps. Merge rules (helpers): palette core slots (primary/secondary/accent/background/surface/text/text_muted/border) spec-wins; specialised slots theme-wins; typography spec-wins; structure layout-only. Theme composition loaded via `sites.style_collection_id → style_collections.css_theme_id → css_themes` joined to `palettes.colours` + `layouts.css_template`/`structure_tokens` + `typography_sets.fonts`/`scale`; hard-errors on NULL FKs.
2. **Component CSS from a `css_snippets` table** (`name`, `css_content`, `applies_to` jsonb), where `applies_to` overlaps the site's component list — appended to `styles.css`. This is a THIRD CSS surface distinct from `content_components.html_template` inline `<style>`. Where the dark hero/CTA/footer treatment actually lives (inline vs snippet vs both) is NOT yet known → inventory in C.
3. A luminance-based `--section-*` defaults block via `buildSectionDefaults(bg, surface, mergedPalette, backgroundIsDark, surfaceIsDark)` — renderer-owned, picks readable section colours from the palette. (`buildSectionDefaults`/`isDarkHex` are referenced but defined in a file not yet pulled — locate next.)
The section-context mechanism has two halves: `SectionStyles` (a `[]{Function, ClassName, IsDark}` passed to the layout template, with **`ClassName = function + "-section"`** and `IsDark` from `content_components.is_dark_section` via `queryDarkSectionsForCSS`) and the `buildSectionDefaults` global block. **The class contract is `{function}-section`** — the stylesheet assumes components emit `hero-section`/`call-to-action-section`, but the report's evidence is they emit `.hero`/`.cta-section`/`.tl-card`. That mismatch (+ the components' inline dark CSS winning) is the concrete, code-level form of Q4. Neither render path reads `layouts.scheme` (loader SELECT omits it; RenderContext has no scheme field) — so the section dark/light today keys off per-component `is_dark_section` and the palette's background luminance, never the site's resolved scheme.
Q4 leaning: **(a) components adopt `{function}-section`**, because `SectionStyles` + `buildSectionDefaults` are already built for that contract (maximum reuse, matches the report). Decision gated on an actual layout `css_template` (D) — does its `SectionStyles` block emit section STRUCTURE + context, or only context?

═══════════════════════════════════════════════════════════════════════
## CHECKPOINT (Sd) — docs written; next is bundle 2 (C + D → Q4)
═══════════════════════════════════════════════════════════════════════
Wrote this thread's three working docs (per request): `PLAN_scheme_to_components.md` (problem, thesis, Q1–Q8 status, fix shape, investigation status, phasing), `RUNBOOK_scheme_to_components.md` (corrected bundle command, the C and D SQL, verification queries, gated migration sketch, rollback/safety), and this journal. No code or DB changes this thread — still the design phase.
Status of the investigations: A mostly done (remaining: `buildSectionDefaults`/`isDarkHex` bodies; assembly detail), B done, G done. Open and gating: Q4, needing the C inventory (incl. `css_snippets` and the section-uniqueness check) and the D class-name pull. Then E (per-section contrast storage; `is_dark_section`'s role), F (header/footer wiring + `update_site_defaults`), H (migration/backfill), I (scheme-coherence guard).
NEXT: run bundle 2 — the C and D SQL in the runbook (against `clients_db`, as files) — and bring back the results to settle Q4. idea.uk unchanged; live £29 VM tool untouched.
