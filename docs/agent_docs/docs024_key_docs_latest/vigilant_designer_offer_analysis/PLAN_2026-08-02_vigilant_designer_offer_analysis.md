# PLAN — vigilant senior designer + offer/benefit analyser

**Opened:** 2026-08-02, at the owner's direction ("a constant vigilant designer… and the offer
analyser and benefit analyser… all of the above can also be made into checkers and handlers
too for continual improvement. It needs a strong focus or we'll keep tripping up on the detail.")

**Organising rule:** every detector ships with its drain (handler + promotion + trigger).
The platform's proven failure mode is detection without consumption — `bugs_open/115`
(a correct brief-fidelity audit predicted the owner's complaints three days early and died at
`status='detected'`), `bugs_open/083`, and the 019 owner ruling that deferred sweep enrolment
for exactly this reason.

## Owner decisions (2026-08-02, at planning time)

| decision | answer |
|---|---|
| Cadence | **Manual triggers for now.** Build everything drainable; fire per site by hand. `improvement-sweep` stays `enabled=false`; flipping it on (G1) is a separate owner go — it reverses the 07-29 deliberate stop. |
| Build order | **Designer first**, then the offer analyser. Shared Phase 0 (drain) first. |
| Critic model | **Trial both** — vision seam implemented for Gemini AND Claude, same critique over 2–3 sites with each, compare actionability, record the ruling (supersedes the untested 07-24 Gemini call). |
| Authority | **Broad autonomy** — auto-apply at every altitude, locks respected. Two platform-constrained exceptions (below). |

## Autonomy model

- Auto-applies: CSS/detail fixes, spacing/responsive, CTA, imagery, bounded page recomposition,
  directed palette change (webdesign-agent + reference_values pinned in the same operation so
  the change sticks instead of churning), content/tone rewrites, strategy refresh.
- **Exception 1 — shared chrome**: 3 head components serve all 14 live domains; an automated
  chrome edit is fleet-wide by construction. Auto path = fork-to-site-scope first (pin the fork
  via `style_collections`, the pin predicate deliberately admits forks), then edit the fork.
  Built as A4.4, last; until then chrome findings file `capability_gap`.
- **Exception 2 — layout-triple swap**: no runtime re-compose path exists by deliberate design
  (DES-048/049). Stays a `capability_gap` review item in v1; a gated auto re-compose is
  architecture-scope, revisit once the critic demonstrates demand.
- Locks always respected (`locked_at IS NOT NULL` never touched, counted as `skipped_locked`);
  byte-preserved adopted sites → `needs_human_review` only.

## The full approved plan

The complete phase-by-phase plan (Programme A phases 0–5, Programme B phases B1–B5, the
item_type/handler/verifier table, verification acceptance list) was approved by the owner on
2026-08-02 and is reproduced in full below. It was designed against three exploration passes +
two planning passes over the live system the same day; the "facts that bind" list is what was
verified then — re-verify anything acted on later, this tree moves in minutes.

### Facts that bind the design (verified live 2026-08-02)

- Promotion is single-owner: migration 286 makes `improvement-loop.triage_findings` the sole
  `triage_detected_items` owner. Never a second triage step; `scripts/audit-single-owner-actions.sh` stays clean.
- `AIService` is text-only (`platform/aiservice/interface.go`). No vision path exists in the
  chassis; the screenshot critic needs a new seam (council + concept-register, same commit).
- `render-audit-agent` (VIZ-012) is live but its browser findings stop in `collected_data` —
  no write tail. `browserrunner/screenshots.go` exists but fires only on check failure.
- The 3-pass audit cap lies: capped sites report "clean" (`bugs_open/171`); the JSON path is
  `sites.settings.maintenance_profile.audit_pass_count`.
- `domain-strategist.create_next_item` unconditionally enqueues `needs_briefing` → rebuild
  chain. Must be gated (B2) before `check_premise_incomplete` ever enables (B3).
- Adopted finance sites (loancash class) have no `strategy` row; `bugs_open/115` is down to
  2 open findings — finding 2 (template repetition) is Programme A's acceptance test.
- Webdesign LLM may only change hue (8 core slots); `install_site_composition` refuses re-runs;
  sites lacking `design_intent.palette.reference_values` re-roll their palette every webdesign
  run (3 sites exposed; the generic_theme repair IS the churn until A4.3).
- Checker recipe: registry init() + sqlmock test + handler/verifier/registration coverage
  tests; check names into an agent's checks array only after the image rolls; IMP-016 order
  (observe-only → handler live → one clean cycle → enable).
- Active adjacent lane — do not collide: portfolio_positioning owns premise→writer wiring for
  new builds (their gap 1, lendzy shadow build). This workstream owns the read-back side of
  the same spec fields. Do not enrol lendzy until their marker test passes.

### Programme A — the vigilant designer

**Phase 0 — make findings flow.**
0.1 Repoint `scheduled_tasks.improvement-sweep.target_topic` → `system.agent.scheduled.requests`
(dead-topic fix), leave `enabled=false` (G1 owner gate). Its pre_query already self-enrols all
active/deployed sites with backpressure, so 019's enrolment is structurally solved by the same
future flip.
0.2 Replace the 3-pass cap with convergence + cooldown: pg function `site_audit_fingerprint(site_id)`
(md5 over ordered `pages.content_hash` + composed palette + chrome rendered_html hashes); state
in `settings.maintenance_profile.last_audit {fingerprint, at, passes_at_fingerprint}`; audit due
iff fingerprint changed OR 14d cooldown expired; **`triage_findings` runs on every path** (the
cap only ever protected LLM spend — promotion is one UPDATE); non-convergence brake at 3 passes
on an unchanged fingerprint → one `audit_not_converging` capability_gap row; `audit_skipped`
recorded so skipped ≠ clean.
0.3 Render-audit write tail: new Go action `write_render_audit_findings` — firm ContrastFinding
→ `contrast_failure` → css-patch-agent (item_key `render:contrast:<page>:<class>`; dedup + next
audit is the de-facto verifier); over_image reported never filed; BrokenImage → existing
undeployed-asset machinery or `needs_imagery` → image-build-handler (never the deliberately
flag-only `image_url_404`); Overflow → `responsive_fix` → component-template-fixer; locked
culprits skipped and counted. Config tail step added to render-audit-agent after the image rolls.
0.4 The ~205 detected items: no bulk promote; per-site hand-fired sweeps starting with one
specimen site; cancel provably-stale rerender rows first; hold 115's two rows for A3.2's
acceptance run.

**Phase 1 — eyes.**
1.1 browserrunner: `capture_screenshots` + `profiles [desktop, mobile]` on the render-audit
path; full-page PNG per (url, profile) unconditionally when asked; uploads under
`design-critique/<site>/<run>/…`; `Screenshots` in the result; `max_pages` caps count.
1.2 Vision seam (new shared seam — council + register same commit): `AIService.GenerateWithImages`
implemented for anthropic AND gemini (the model trial is a config switch; ollama typed
not-supported); new `execute_vision_prompt` action (URIs → bytes via storage client; same
overlay/model config + llm_call_logger path as `execute_llm_prompt`).

**Phase 2 — the 018 critic.** New `design-critique-agent`, third spawn/call pair in
design-audit-agent, internal due-gate (fingerprint/14d → else `complete_not_due`):
`ensure_site_record → check_critique_due → request_render_audit (screenshots, ≤8 pages × 2
viewports) → load_design_context (design_intent, composed palette, section inventory, fleet
homepage-skeleton summary) → critique (execute_vision_prompt) → write_findings
(audit_source='design-critique') → record_critique → complete`. Three-altitude prompt (site max
3 / component max 6 / detail max 3, 12/site cap), closed category vocabulary matching the
router exactly, findings on the proven contract (acceptance_test, affected_component,
max_fix_attempts…). Routing: component_css/detail_css → `design_css_fix` → css-patch-agent
(+ real verifier); spacing/responsive/cta → existing fixers; imagery_density → needs_imagery;
page_composition → `needs_page_recompose` (enable only once A3.1 lands); brand/palette →
`needs_design_review` → webdesign-agent with pin-on-apply; chrome → capability_gap until A4.4;
new `unknown_category_policy: capability_gap` for this audit_source. Model trial across 2–3
sites, ruling recorded here.

**Phase 3 — recompose drain + 016/017.**
3.1 `page-recompose-handler`: claim → load page/sections/candidates (join tables, never
usage_count) → LLM bounded recompose (≤2 section changes, locks retained, archetype constraints)
→ Go `validate_recompose_plan` → apply via existing plan/section path → rerender item → complete.
Real verifier: sections differ as acceptance names. DES-038 boundary: selects from registered
components only.
3.2 Revive 016: routable categories only; due-gate (brief change + 30d); wire into
design-audit-agent. **Acceptance = bug 115's finding 2 travelling detected → triaged →
recompose → a page whose section plan differs.** Then experimental → active.
3.3 017's checks: `check_interactivity_promised` → needs_page_recompose;
`check_generic_section_sole_candidate` → capability_gap; fleet dormancy inventory feeds A4.1
steering + the fixloop digest, not work items.

**Phase 4 — anti-brochure compose-time.**
4.1 Per-archetype structural constraints in build-site-planner prompt (required AND forbidden
sections per site_type; anti-sameness instruction; dormant-component steering by exemplar).
Council/owner-flagged: site_type vocabulary extension, core-slot changes.
4.2 `check_composition_convergence`: homepage section-sequence similarity ≥0.8 vs a different-
proposition site, or shared layout+palette+typography triple → needs_page_recompose naming the
twin; template-level residue → capability_gap.
4.3 Churn fix: `check_palette_unpinned` → `palette-pinner` handler (composed core slots →
reference_values; gated on contrast-clean). Lands before any sweep-driven webdesign volume.
4.4 Chrome fork-to-site-scope handler (broad-autonomy completion; riskiest, last; pin predicate
not pool predicate; LANDMINES :690/:2801 first; council).

**Phase 5 — enablement (manual-trigger mode).** Hand-fire per site; checks into
design-discovery-agent's array only after their image rolls; per-category critic enablement
follows its handler; **G1 (flip improvement-sweep on) deferred to a separate owner go.**

### Programme B — offer/benefit analyser (after A)

B1 Widen `site-review-agent.load_strategic_context` to load strategy/identity/content_direction/
mission_brief/audience; prompt judges against primary_model. Verify by planted marker reaching
the assembled prompt.
B2 De-mine domain-strategist (refresh_mode/deployed → complete WITHOUT create_next_item — else
every refresh enqueues a rebuild chain) + four Q-fields in the strategy shape
(satisfaction_condition, money_flow, recurring_value_hook, trust_threshold) + refresh-preserves
instruction. Verify on loancalculator.co.uk: new strategy row, zero needs_briefing items.
B3 `check_premise_incomplete` → `needs_strategy` → domain-strategist (fills the "none"/"implied"
premise on the live estate); `check_revenue_shape` (doc 028's rule: tools/advertising → no
service-selling CTA lexicon → revenue_shape_cta → component-template-fixer, residue →
capability_gap; lead-gen/direct → nav-reachable contact/quote page with a form →
missing_conversion_path → content-gap-planner; affiliate → one capability_gap, machinery absent;
locked/adopted → needs_human_review only). Into quality-discovery-agent's array after the roll,
observe-only first.
B4 `offer-analyser` agent (config-only): load_premise → load_offer_surface (ALL active pages) →
run_offer_analysis → write_findings (audit_source 'offer-analysis'). Judges need↔offer fit,
the benefit table (promised → delivered? surfaced?), monetisation shape, improvements — closed
vocabulary emitting EXISTING item_types only. Grades against the site's own proposition
coordinates. Prompt forbids "users want…" phrasing — artefact-vs-premise only (zero outcome
data; cheapest future instrument = the disabled intent-collection task, named not scoped).
Due-gated inside improvement-loop before triage_findings.
B5 End-to-end proof on webdesign.co.uk + both planted-marker proofs; enrolment order = owner
calls at the time; lendzy only after the positioning lane's marker test passes.

### Item-type / handler / verifier summary (new or changed)

| item_type | detector | handler | verifier |
|---|---|---|---|
| contrast_failure | render-audit tail | css-patch-agent | classified (next audit + dedup key) |
| design_css_fix | critic css categories | css-patch-agent | NEW (patch CSS contains named property) |
| needs_page_recompose | critic, 016, interactivity, convergence | page-recompose-handler (NEW) | NEW (sections differ) |
| palette_unpinned | check_palette_unpinned | palette-pinner (NEW) | NEW (reference_values present) |
| needs_design_review (directed palette) | critic brand | webdesign-agent + pin-on-apply | existing |
| needs_strategy | check_premise_incomplete | domain-strategist (live) | NEW (strategy row with primary_model) |
| revenue_shape_cta | check_revenue_shape | component-template-fixer | NEW (lexicon re-scan) |
| missing_conversion_path | check_revenue_shape | content-gap-planner | NEW (page+form query) |
| capability_gap | chrome (until A4.4), layout swap, affiliate, sole-candidate, non-convergence, off-vocabulary | none — roadmap by design | classified |
| offer LLM findings | offer-analyser | existing handlers only | existing |

### Programme-level acceptance

1. Bug 115's own test: brief-fidelity finding 2 → a page whose section plan differs.
2. A render-audit contrast failure → css-patch → next render audit clean.
3. A critic design_css_fix → applied → verifier passes → re-critique does not re-raise.
4. Two archetype-different sites structurally diverge after A4.1/A4.2 (section lists compared).
5. One offer finding end-to-end on webdesign.co.uk + both planted-marker proofs.

## Corrections to Programme B (2026-08-08, from `REVIEW_2026-08-08_…`)

Marked here rather than edited into §B above, so the original text stays readable.

- **B2's four Q-fields are a RESTORATION, not an invention.** `satisfaction_condition`,
  `trust_threshold`, `recurring_value` and the money-flow field are the shape
  **gaswholesalers.com** has carried since 2026-04-17 (its whole strategy spec is those six
  keys — `satisfaction_condition, trust_threshold, recurring_value, monetisation,
  primary_intent, visitor_type` — and it is the only site with it; the other 16 carry the
  current 12-key shape). Copy the live row rather than designing from the plan text.
- **`primary_model` is NOT a defect in this plan — do not "fix" §B1 or §B3.** A first pass of
  the 08-08 review claimed the field had never been written and called those two lines
  defective. **That was wrong**: it exists on **16 of 17** sites, nested at
  `revenue_models.primary_model` (the shape `domain-strategist`'s own prompt declares). The
  reviewer had read the top level only. Recorded because a future reader who finds only the
  first claim would "correct" two correct lines. `WRONG_CALLS.md` 2026-08-08.
  - Distribution, worth having: `direct_business` **10**, `saas_tools` 3,
    `display_advertising` 2, `lead_generation` 1, `sponsored_listings` 1, absent 1. So §B3's
    `check_revenue_shape` has a real population on day one.
- **`needs_strategy` already has a live producer** (`vertical-exemplar-researcher` →
  `domain-strategist`; 3 complete rows). §B3 makes this lane the **second** producer on an
  existing `item_type` ⇒ the **owner ruling 2026-08-02 / RFC_010 §1** applies: no architecture
  round, **but the concept-register entry must name the full producer set and state the shared
  `item_key` shape**. Not optional.
- **B2's hazard is verified three links deep, not one**: `domain-strategist.create_next_item`
  (unconditional) → `needs_briefing` → `build-briefing-agent` → `needs_site_plan` →
  `build-site-planner`. All 3 historical `needs_briefing` rows are greenfield builds, where
  the chain is correct; it has never run against a deployed site.
  [UNVERIFIED beyond the planner — `build-site-planner` files no further work items.]

## Decision log

- **2026-08-02** — programme approved by the owner with the four decisions tabled above.
  Chosen over: re-enabling the sweep now (owner kept the 07-29 stop in force); offer-first
  ordering; single-model critic. Broad autonomy chosen over detail-auto/structure-review —
  the two stated exceptions are platform constraints, not authority limits.
- **2026-08-08 — OWNER DECISION: B1 and B2 jump the queue.** This **partially reverses the
  08-02 "designer first" build order**, for these two items only; the rest of Programme B
  stays behind Programme A, and A's own order is unchanged. Reason: neither B1 nor B2 depends
  on anything in Programme A, both are `agent_definitions` config (live on apply, no image
  roll, no committed-but-inert window), and each fixes something already wrong — B1 because
  `site-review-agent` asks the offer question ~16×/fortnight with no premise in context, B2
  because a premise refresh on a deployed site would re-plan it. Taken after
  `REVIEW_2026-08-08_…`; the owner also directed the wider scope be written up as
  `features_open/030_FEATURE_offer_and_benefit_analyser.md`, which now holds the
  correspondence surface, the council question and the open questions.
