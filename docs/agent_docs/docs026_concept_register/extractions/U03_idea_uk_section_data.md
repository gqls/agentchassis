# EXTRACTION U03 — docs024_key_docs_latest/idea_uk_section_data_missing
Extracted 2026-07-13. Files in scope: 191. Concepts found: 48.

Unit character: two intertwined debugging/build threads on the idea.uk chassis site —
(1) the June 2026 "differentiators empty cards / section data" investigation, and
(2) the July 2026 "scheme reaches components" P0 (light-resolved site rendering dark),
which spawned the imagery-resolution, presigned-URL-expiry, and Go-batch sub-threads.
Both version families are rolling snapshots: `running_notes_scheme_to_components` is
strictly append-only (every earlier version is a byte-prefix of (55)); the RUNBOOK
family rewrites its position banner while appending, so early versions were diffed
against (50) — no concept present in an earlier version is absent from (50) (the only
divergences are rewritten position/status text, e.g. STEP C shrinking from a
five-minute seed check to a single grep between (49) and (50)).

Paths in the coverage table are relative to
`docs/agent_docs/docs024_key_docs_latest/idea_uk_section_data_missing/`.

## Coverage

| file | treatment |
|---|---|
| 001_bundling_context.md | full |
| 019_pcw_prompt_item_fields.sql | full |
| 019_pcw_prompt_item_fields_down.sql | full |
| HANDOFF_idea_uk_differentiators_section_data.md | full |
| HANDOFF_scheme_to_components_for_claude_code(1).md | family-latest |
| HANDOFF_scheme_to_components_for_claude_code.md | family-delta |
| PLAN_scheme_to_components(1).md | family-latest |
| PLAN_scheme_to_components.md | family-delta |
| README.md | full (zero-byte file) |
| README_001_flow_notes.md | full |
| RUNBOOK_pcw_item_fields_fix.md | full |
| RUNBOOK_scheme_to_components.md | family-delta |
| RUNBOOK_scheme_to_components(1).md | family-delta |
| RUNBOOK_scheme_to_components(2).md | family-delta |
| RUNBOOK_scheme_to_components(3).md | family-delta |
| RUNBOOK_scheme_to_components(4).md | family-delta |
| RUNBOOK_scheme_to_components(5).md | family-delta |
| RUNBOOK_scheme_to_components(6).md | family-delta |
| RUNBOOK_scheme_to_components(7).md | family-delta |
| RUNBOOK_scheme_to_components(8).md | family-delta |
| RUNBOOK_scheme_to_components(9).md | family-delta |
| RUNBOOK_scheme_to_components(10).md | family-delta |
| RUNBOOK_scheme_to_components(11).md | family-delta |
| RUNBOOK_scheme_to_components(12).md | family-delta |
| RUNBOOK_scheme_to_components(13).md | family-delta |
| RUNBOOK_scheme_to_components(14).md | family-delta |
| RUNBOOK_scheme_to_components(15).md | family-delta |
| RUNBOOK_scheme_to_components(16).md | family-delta |
| RUNBOOK_scheme_to_components(17).md | family-delta |
| RUNBOOK_scheme_to_components(19).md | family-delta |
| RUNBOOK_scheme_to_components(20).md | family-delta |
| RUNBOOK_scheme_to_components(21).md | family-delta |
| RUNBOOK_scheme_to_components(22).md | family-delta |
| RUNBOOK_scheme_to_components(23).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(24).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(25).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(26).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(27).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(28).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(29).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(30).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(31).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(32).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(33).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(34).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(35).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(36).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(37).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(38).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(39).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(40).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(41).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(42).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(43).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(44).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(45).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(46).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(47).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(48).md | family-delta (prefix of 50) |
| RUNBOOK_scheme_to_components(49).md | family-delta (STEP C rewritten in 50) |
| RUNBOOK_scheme_to_components(50).md | family-latest |
| SPEC_scheme_to_components.md | full |
| bundle3 | full |
| gobatch_01_plan_sections.md | full |
| gobatch_02_component_library.md | full |
| gobatch_03_deploy_asset_localise.md | family-delta |
| gobatch_03_deploy_asset_localise(1).md | family-latest |
| gobatch_04_fixer_reaim.md | family-delta (on-primary token guess, corrected in (2)) |
| gobatch_04_fixer_reaim(1).md | family-delta |
| gobatch_04_fixer_reaim(2).md | family-latest |
| gobatch_05_flag_section_scope.md | full |
| plan_pcw_item_fields_fix.md | family-delta (byte-identical to (1)) |
| plan_pcw_item_fields_fix(1).md | family-latest |
| running_notes_checkpoint_ss(0).md | family-delta (byte-identical to (1)) |
| running_notes_checkpoint_ss(1).md | family-latest |
| running_notes_checkpoint_uu.md | full |
| running_notes_scheme_to_components.md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(1).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(2).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(3).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(4).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(5).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(6).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(7).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(8).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(9).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(10).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(11).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(12).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(13).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(14).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(15).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(16).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(17).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(18).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(19).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(20).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(21).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(23).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(24).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(25).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(26).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(27).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(28).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(29).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(30).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(31).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(32).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(33).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(34).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(35).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(36).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(37).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(38).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(39).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(40).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(41).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(42).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(43).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(44).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(45).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(46).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(47).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(48).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(49).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(50).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(51).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(52).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(53).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(54).md | family-delta (prefix of 55) |
| running_notes_scheme_to_components(55).md | family-latest |
| slice4a_creator_prompt.sql | full |
| slice4a_rollback.sql | header-scan (mechanical inverse of slice4a) |
| slice4b_003_contract.md | full |
| slice4c_step_description.sql | full |
| stepD_and_pages_reads.sql | full |
| stepF_replan_read.sql | full |
| w1_00_before.sql | full |
| w1_01_add_cta_pair.sql | full |
| w1_02_verify.sql | full |
| w1_rollback.sql | full |
| w1b_01_contrast_batch.sql | full |
| w1b_02_verify.sql | full |
| w1b_rollback.sql | header-scan (mechanical inverse) |
| w2_00_before.sql | full |
| w2_01_footer_fix.sql | full |
| w2_02_verify.sql | header-scan (superseded by _fixed) |
| w2_02_verify_fixed.sql | full |
| w2_rollback.sql | header-scan (mechanical inverse) |
| w2b_00_trigger_check.sql | full |
| w2b_01_layouts_updated_at_trigger.sql | full |
| w3a_01_cta_conversion.sql | full |
| w3a_02_verify.sql | header-scan (read verify) |
| w3a_rollback.sql | header-scan (mechanical inverse) |
| w3b_00_before.sql | header-scan (gate read) |
| w3b_01_hero_conversion.sql | full |
| w3b_02_verify.sql | header-scan (read verify) |
| w3b_rollback.sql | header-scan (mechanical inverse) |
| w3c_00_before.sql | full |
| w3e_01_gate_and_fix.sql | full |
| w3e_02_verify.sql | header-scan (read verify) |
| w3e_rollback.sql | header-scan (mechanical inverse) |
| w4b_00_before.sql | full |
| w4b_01_repoint.sql | full |
| w4b_02_read_triggers.sql | full |
| w4b_04_trigger_item.sql | full |
| w4b_05_verify_chrome.sql | full |
| w4b_rollback_repoint.sql | header-scan (mechanical inverse) |
| w6_03_final_verify.sql | full |
| w6_04_check_dropped_section.sql | full |
| w6_05_section_data_read.sql | full |
| w7_00b_paste_as_text.sql | full |
| w7_00c_paste_as_text.sql | full |
| w7a_01_gate.sql | full |
| w7a_rollback.sql | header-scan (mechanical inverse) |
| w7b_01_imagery.sql | full |
| w7b_02_verify.sql | full |
| w8_01_post_deploy_rebuild.sql | full |
| w8_03_fingerprint.sql | full |
| w8_05_provenance_wide.sql | full |
| w8_06_experiment_OPTIONAL.sql | full |
| w8_07_fresh_index_build.sql | full |
| w8_08_corrected_probe.sql | full |
| w8_09_hero_exposure.sql | full |
| w9_00_localisation_read.sql | full |
| w9_01_deploy_step_and_hero_consumers.sql | full |
| w9_02_deployer_and_shadow.sql | full |
| w9_03_assets_schema_and_inventory.sql | full |
| w9_04_backfill_flip.sql | full |
| w9_05_rebuild_and_verify.sql | full |
| w9_06_verify_only.sql | full |

## Concepts

### Scheme-to-components P0: light-resolved site renders dark
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** PLAN "## CLOSED (2026-07-03) … closed on deployed evidence (RUNBOOK §SCHEME CLOSE: all nine grep checks pass; the stale-section fossil `var(--accent-color` is gone)"; RUNBOOK §SCHEME CLOSE lists the nine counts (site-header-section 32 / gradient 0 / footer 37 / --hero-ink 13 / color-mix 14 / cta pair consumed / white rgba 0 / fossil 0 / brief-explanation 0-expected).
- **what:** The defining P0 of this thread: the chassis resolves each site to a light or dark scheme and the scheme travels correctly through layout and palette variables, but the component library was written dark-first — components hardcode white text and dark `--section-*` context inline, so a light-resolved site (idea.uk, `tool-portal-light`) deployed dark chrome and dark sections. The winning mechanism was completion of the existing paired-variable standard rather than restructure: one layout patched, ten templates de-hardcoded, chrome repointed + force-rerendered, then a full page-build-handler rebuild.
- **sources:** PLAN_scheme_to_components(1).md#CLOSED; RUNBOOK_scheme_to_components(50).md#SCHEME-CLOSE; HANDOFF_scheme_to_components_for_claude_code(1).md#The-problem; running_notes_scheme_to_components(55).md#Tk
- **relations:** paired-variable standard; hero ink model; hazard/band declarer taxonomy; rebuild vs rerender semantics; chrome selection path.
- **verify-later:** deployed idea.uk B2 index.html greps; `content_components.html_template` for site-footer/call-to-action/hero; `layouts.css_template` for tool-portal-light.

### Scheme derivation and drop at render
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** PLAN §Confirmed-at-code-level: "`deriveSchemeFromDesignIntent(style_direction, suggested_style)` returns light/dark/''; … `buildResolvedCompositionSpec` records the layout/palette ids … but not the scheme value"; notes (Sb) traced it end-to-end 2026-06-30.
- **what:** Scheme (light/dark) is derived at composition from `design_intent.style_direction` by substring matching, used by `resolveLayoutByTags` as a near-hard constraint to pick the layout, then dropped: neither the CSS loader SELECT nor the component `RenderContext` reads `layouts.scheme` (check-constrained to light/dark/neutral). It survives only as the layout's curated property, recoverable via `sites.style_collection_id → style_collections.css_theme_id → css_themes.layout_id → layouts.scheme`. Light/dark variety is handled by paired layouts (tool-portal-light/-dark), not runtime component flipping. Corollary data fact: only 3 of 18 active layouts have `scheme` set — scheme metadata is sparsely curated.
- **sources:** PLAN_scheme_to_components(1).md#Confirmed-at-code-level; running_notes_scheme_to_components(55).md#Sb #Sf; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3e)
- **relations:** three-part styles.css assembly; paired-variable standard; explicit RenderContext.Scheme (abandoned).
- **verify-later:** `platform/orchestration/actions/` deriveSchemeFromDesignIntent, resolveLayoutByTags, buildResolvedCompositionSpec; `layouts.scheme` column + values.

### Three-part styles.css assembly and palette merge rules
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sc), read from `render_css_from_spec_action.go` bodies in full, 2026-06-30; HANDOFF §Established mechanism restates it as verified.
- **what:** `RenderCSSFromSpecAction` builds styles.css in three appended parts: (1) the layout `css_template` rendered as a Go text/template with `{{palette}}`/`{{typo}}`/`{{token}}` FuncMap helpers over merged maps — merge rules: 8 core palette slots spec-wins, specialised slots theme-wins, typography spec-wins, structure layout-only; (2) component CSS from the `css_snippets` table (name, css_content, applies_to jsonb) where applies_to overlaps the site's components — a third CSS surface distinct from inline `<style>` (C3 cleared it of dark-section treatment: all 21 snippets are utilities); (3) the `buildSectionDefaults` luminance block. Theme composition loads via style_collections → css_themes joined to palettes/layouts/typography_sets, hard-erroring on NULL FKs.
- **sources:** running_notes_scheme_to_components(55).md#Sc #Se; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; PLAN_scheme_to_components(1).md#Confirmed-at-code-level
- **relations:** buildSectionDefaults; layout CTA pair curation; scheme derivation.
- **verify-later:** render_css_from_spec_action.go, render_css_composition_loader.go/_helpers.go; `css_snippets`, `palettes.colours`, `typography_sets` tables.

### buildSectionDefaults: luminance-keyed dark-only --section-* defaults
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sf) 2026-07-01: "`--section-*` is a DARK-ONLY override; light is the fallback. `buildSectionDefaults` returns '' unless bg or surface is dark."
- **what:** The renderer's only live per-section adaptation: `buildSectionDefaults` (color_util.go, WCAG `isDarkHex`/`pickReadableOnBackground`) emits a `body { --section-* }` block only when the merged palette's background or surface is dark, plus a dark-surface variant on 5 hardcoded surface classes (`.features/.services/.differentiators/.about/.faq-section`). On a light palette it emits nothing and element rules fall through to `var(--color-*)`. Retained unchanged under the paired-variable decision as the whole-palette-darkness base/safety net.
- **sources:** running_notes_scheme_to_components(55).md#Sf; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; SPEC_scheme_to_components.md#Decision-record
- **relations:** Colour Inheritance Model; Phase 4.5 deferral (generalises the 5-class list); paired-variable standard.
- **verify-later:** color_util.go buildSectionDefaults/isDarkHex; emitted styles.css tail on a dark-palette site.

### SectionStyles per-section CSS mechanism (built, disconnected, retired)
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** Notes (Sf): "`SectionStyles` is DEAD for current sites. None of the 18 active layouts reference `{{range .SectionStyles}}` … computed-but-unused"; SPEC: "`SectionStyles` stays retired."
- **what:** A fully-built but never-connected renderer mechanism: `queryDarkSectionsForCSS` + `buildCSSsectionStyles` compute per-component `{Function, ClassName: function+"-section", IsDark}` entries from `content_components.is_dark_section` (fallback dark list hero/social-proof/call-to-action/testimonials) and pass them to the layout template — which no active layout consumes. Considered as the cheap renderer-owns vehicle (Alt B) and explicitly retired by the paired-variable decision. A textbook infrastructure-orphan: ~80% built, deliberately not revived.
- **sources:** running_notes_scheme_to_components(55).md#Sf #Si; SPEC_scheme_to_components.md#Decision-record; HANDOFF_scheme_to_components_for_claude_code(1).md#Established
- **relations:** superseded by paired-variable standard; related Phase 4.5 (the other renderer-owns design).
- **verify-later:** render_css_from_spec_action.go buildCSSsectionStyles/queryDarkSectionsForCSS still present and uncalled from layouts.

### {function}-section class contract and data-component naming contract
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** Notes (Sf): "The `{function}-section` contract is REAL + operative, honoured unevenly" — honoured by the 5 surface sections + footer-with-disclaimer; not by hero (`.hero`) or CTA (`.cta-section`).
- **what:** Layouts and `buildSectionDefaults` key structural rules and surface treatment on `.{function}-section` class names, but the compiler (`CompilePageSectionsAction`) concatenates component HTML without wrapping, so the class is each component's own responsibility and adoption is inconsistent — the mechanism misses non-adopters and their inline CSS wins. Separately, every component root does carry `data-component="{function}"` (kebab-case, enforced by component_validation.go), giving an attribute-selector escape hatch the class mismatch cannot break.
- **sources:** running_notes_scheme_to_components(55).md#Sc #Se #Sf; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; PLAN_scheme_to_components(1).md#Q4
- **relations:** Colour Inheritance Model; SectionStyles (dead consumer of the same names); section painting contract.
- **verify-later:** component_validation.go naming checks; class emission across `content_components.html_template`.

### Colour Inheritance Model (var(--section-*, var(--color-*)) chains)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** HANDOFF §Established: "Element rules follow the Colour Inheritance Model: `var(--section-*, var(--color-*))`."
- **what:** The base element rules in layouts/base CSS resolve text/heading/muted/border colours through a two-level chain: a section-scoped custom property if declared, else the palette-level colour. This is what makes "components declare no colours" viable: a non-painting section inherits page ink automatically; a painting section re-exports its context onto the `--section-*` layer and every child element follows. The W3e "ambient pass-through" fix (`--section-x: var(--color-x)`) exists because some internal consumers lack var() fallbacks — deletion would fall to currentColor/transparent.
- **sources:** HANDOFF_scheme_to_components_for_claude_code(1).md#Established; SPEC_scheme_to_components.md#The-contract; running_notes_scheme_to_components(55).md#Sx (fix rationale)
- **relations:** section painting contract; buildSectionDefaults.
- **verify-later:** base head CSS / layout css_templates element rules.

### Paired-variable ("on-colour") standard — the decision record
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** SPEC (2026-07-02) "The standard is **paired variables** … Checks 4b/4c show it is the existing library standard incompletely adopted: 18/18 layouts define `--color-primary-text`, 17/18 define `--color-cta-text`"; executed via W1–W6 and closed 07-03.
- **what:** Every paintable band colour has a matching text colour, curated per layout (and therefore per scheme), overridable per site through palette specialised slots (theme-wins merge), with per-instance control later available via `site_plan_directives` scope=section. Selected over four alternatives (Alt 0 stale-build, Alt A component-owned bands, Alt B renderer-owned via is_dark_section, Alt full-025) after the user's gating answer: "a light scheme must be able to render fully light, and may carry dark hero bands" — band darkness must be a choice, not a component constant. Existing names are reused (`--color-header-bg/-text`, `--color-footer-bg/-text`, `--color-cta-bg/-text`, `--color-primary/-text`, `--color-hero-title/-subtitle`); the direction is completion of existing architecture, not restructure.
- **sources:** SPEC_scheme_to_components.md#Decision-record; running_notes_scheme_to_components(55).md#Sn #So; HANDOFF_scheme_to_components_for_claude_code(1).md#The-Decision; RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS
- **relations:** section painting contract (its component-facing rules); layout CTA pair curation; is_dark_section demotion; supersedes SectionStyles and defers Phase 4.5.
- **verify-later:** layout css_templates pair definitions; palette specialised-slot merge in composition helpers.

### Section painting contract (003 item 6 rewrite: four painting models)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** slice4b delivers the 003 rewrite as a patched doc; RUNBOOK 07-06-night: "Slice 4b: DELIVERED as a patched doc … STEP A — copy `outputs/003_contracts_and_standards_7_.md` over the repo's 003 doc" — repo copy still a pending user step at last dated evidence.
- **what:** Replaces 003's "Dark Section Contract" (if is_dark_section=true, template MUST set `--section-*`) with: a template's appearance derives from what its own CSS paints; a painting section chooses exactly one model and re-exports `--section-*` AS REFERENCES ONLY — (a) pair band re-exporting the pair text, (b) palette band re-exporting the on-colour family (`--color-primary-text, var(--color-background)`), (c) image/layered background defining `--hero-ink` per branch, (d) ambient: no background of its own and NO `--section-*` at all. Literal colours in `--section-*` declarations are forbidden; muted/border/surface derive via `color-mix`. The old contract is the exact inverse — the concept records a full contract reversal.
- **sources:** SPEC_scheme_to_components.md#The-contract; slice4b_003_contract.md; running_notes_scheme_to_components(55).md#Sh (old 003 item 6) #Ui
- **relations:** paired-variable standard; component-creator prompt re-aim; fix_forced_text_colours re-aim (mechanical enforcer); image fields rule 6b.
- **verify-later:** repo `docs/.../003_contracts_and_standards*.md` item 6/6b current text; whether outputs copy landed.

### is_dark_section demoted to catalogue metadata
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** SPEC consequences: "`is_dark_section` is demoted to selection/imagery metadata (6 of 37 declarers contradict their own flag — never key styling on it)"; slice4a prompt text landed 07-06 ("catalogue metadata ONLY").
- **what:** `is_dark_section` is a component-level boolean authored by the component-creator LLM (`store_generated_component` extracts it from generated JSON), scheme-blind (the needs-new-component spec carries no scheme field, so the library skewed dark independent of sites), and unreliable (6/37 self-declarers contradict their own flag). The decision demotes it: nothing may style from it — styling derives from what the template's CSS paints. It survives only as selection/imagery metadata. The earlier Q5/E design question ("where does per-section contrast intent live — site_plan_sections?") dissolves under the paired-variable model.
- **sources:** SPEC_scheme_to_components.md#Decision-record; running_notes_scheme_to_components(55).md#Sh #Sn; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (flag hygiene); slice4a_creator_prompt.sql
- **relations:** section painting contract; fix_forced_text_colours re-aim (isDarkSection param kept-ignored).
- **verify-later:** store_generated_component_action.go extraction; component_selector.go non-use; creator prompt current text.

### Hero ink model and the structural-dark exception
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sv) "W3b COMPLETE (UPDATE 1; … ink in both inline branches, layered solid+single-hue color-mix gradient, five ink-referencing section vars…)"; W3d extended it to the five hero-* variants (UPDATE 5).
- **what:** Image/layered sections define a per-branch `--hero-ink` custom property and derive all text/context from it: the image branch sets `--hero-ink:#fff` under the structural-dark exception (an `rgba(0,0,0,x)` overlay guarantees darkness, so white text is always safe); the no-image branch sets `--hero-ink: var(--color-primary-text)` over a layered `var(--color-primary)` solid plus a single-hue gradient mixing 15% toward the ink (depth on both dark and light primaries; the solid layer doubles as the color-mix-less fallback). Buttons become the inverse pair (ink background, primary label). Chosen after data showed imageless heroes are the common case (80/114 hero, 26/26 hero-*), and it fixed a latent white-on-cyan failure on tool-portal-dark.
- **sources:** running_notes_scheme_to_components(55).md#St #Su #Sv #Sw; w3b_01_hero_conversion.sql; RUNBOOK_scheme_to_components(50).md#HERO-(c)-DESIGN
- **relations:** section painting contract model (c); paired-variable standard.
- **verify-later:** hero + hero-* `content_components.html_template` current bytes; rendered index hero.

### Hazard-class vs band-class declarer taxonomy (library blast radius)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** CHECK 3 RESULTS (2026-07-02): 84 active sections — 15 hex backgrounds, 37 self-declare `--section-*`, split ~18 hazard vs ~19 band; SCHEME CLOSE remaining work item 4: "~10 remaining surface-painting declarers + ~17 band-class components (non-idea.uk)" still open.
- **what:** The diagnostic taxonomy that sized every fix decision: hazard-class components declare dark `--section-*` while painting surface variables or nothing — live white-on-light bugs today (the footer, site-head, the five hero-* variants, brief-explanation etc.); band-class components paint from primary/secondary/accent with white text — coherent today but blocking "fully light" (CTA, hero, social-proof, testimonials…). Ten templates (the idea.uk-visible set) were fixed by hand-needles; the non-idea.uk tail awaits the re-aimed fixer (Step D decision).
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS #SCHEME-CLOSE; SPEC_scheme_to_components.md#W2 #W3; running_notes_scheme_to_components(55).md#Sn
- **relations:** fix_forced_text_colours re-aim (the tail vehicle); supervised fixer first-run.
- **verify-later:** re-run the 3c split query; count remaining literal declarers among active sections.

### Chrome selection path and the dead header_component_id column
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sl): "`install_site_composition` sets `style_collections.header_component_id`/`footer_component_id` = NULL … grep finds NO code that writes them non-NULL … effectively a DEAD column"; HANDOFF §Established restates the chain with line numbers.
- **what:** Page-compile chrome resolution: `CompilePageSectionsAction` → `InjectHeader/InjectFooter/InjectHead` → `RenderHeader/RenderFooter` reads `style_collections.header_component_id` (always NULL — inserted NULL with a "webdesign-agent populates these later" comment and never written) → falls to `GetComponentByFunction("site-header")`, the single library-wide active component per function → else the hardcoded-dark fallback. `RenderHead` looks up function `head` (the only head component is inactive, so builds always used the fallback head); `site-head` is section-level and unreachable as chrome. Five other active header/footer functions (`*_pre_037`) are unreachable on this path. The one-active-component-per-function convention holds for sections by data (C4: no function has >1 active) though the UNIQUE index only covers tools.
- **sources:** running_notes_scheme_to_components(55).md#Sl #Sq #Se(C4); HANDOFF_scheme_to_components_for_claude_code(1).md#Established; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3d)
- **relations:** four chrome default stores; scheme-aware fallback chrome; dual chrome render paths.
- **verify-later:** component_library.go RenderHeader/RenderFooter/RenderHead/GetComponentByFunction; install_site_composition_action.go NULL insert + comment fate (W4c chose deleting the comment).

### Four overlapping chrome default stores and the update_site_defaults linkage
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** Notes (Sh) F-section: intended chain documented in 003's Site Component Linkage Contract; SPEC W4(c): "keep function-lookup as the norm … Smallest honest change: the comment" — the linkage deliberately left unrepaired.
- **what:** Header/footer defaults coexist in four stores: `style_collections.header/footer_component_id` (the operative read, dead-NULL), `site_components` slots (copy target + pre-render cache; idea.uk's were pinned to inactive components), `sites.default_components` JSONB (UpdateSiteDefaultsAction's target — a tracking copy nothing reads on the render path), and `layouts.default_*_component_id` (FK, all NULL, nothing copies it onward). The intended chain — style_collections as source of truth, `update_site_defaults` copying into site_components — never runs in composition (003's documented failure mode #1 IS idea.uk's case). The fix chose function-lookup as the norm rather than reviving the chain; populating style_collections at install remains a possible per-site-variety feature.
- **sources:** running_notes_scheme_to_components(55).md#Sg #Sh; SPEC_scheme_to_components.md#W4; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3b)
- **relations:** chrome selection path; Q6's original layouts.default_* direction (superseded by this resolution); site_components repoint.
- **verify-later:** v3_site_actions.go UpdateSiteDefaultsAction; whether the misleading install comment was deleted; any later population of style_collections chrome ids.

### Scheme-aware fallback chrome (RenderFallbackHeader/Footer consume the pairs)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Ud) 2026-07-06: "slice 2+F live"; RUNBOOK 07-06-night State: "Deployed: slices 1 …, 2 (fallback chrome C/D + Debug tidy E)".
- **what:** The safety-net chrome functions hardcoded dark (`background: ctx.PrimaryColor` default `#1a1a2e`, literal white text) — so any site whose chain broke got dark chrome regardless of scheme. Edits C/D replace the whole functions: backgrounds become `var(--color-header-bg, var(--color-surface))`/footer equivalent, text `var(--color-header-text, var(--color-text))`, muted/borders via `color-mix` — safe library-wide because Check 3e proved all 18 layouts set all four chrome vars. `RenderFallbackHead` deliberately unchanged (its only colour use is a `<meta theme-color>` value where `var()` cannot work). Edit E swapped the file's eight `logger.Debug` calls to Info per the no-Debug rule.
- **sources:** gobatch_02_component_library.md; running_notes_scheme_to_components(55).md#Sl #Tq #Ud; SPEC_scheme_to_components.md#W4(a)
- **relations:** chrome selection path; paired-variable standard; no-logger.Debug convention.
- **verify-later:** component_library.go RenderFallbackHeader/RenderFallbackFooter current bodies; deployed image tag containing them.

### Dual chrome render paths and repoint-before-force_rerender ordering
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sy): "render_site_components_action.go:345–430: pinned-component join has NO is_active filter; non-force runs SKIP non-empty slots → repoint BEFORE force_rerender or the old dark chrome re-renders"; W4b executed and verified 2026-07-02 (header 3750→6258B, footer color-mix in).
- **what:** Chrome has two render paths: the page-compile path renders fresh via RenderHeader/Footer, while `render_site_components` writes `site_components.rendered_html`, which the RERENDER handler injects into pages. The pinned join ignores `is_active`, and without `force_rerender` non-empty slots are skipped — so stale renders of deactivated components persist indefinitely. The W4b remedy pattern: repoint `site_components.component_id` to the active components (guarded on the known old ids; `rendered_html` deliberately left in place so there is no chrome-less window), then trigger rerender-pages v6 with `spec.refresh_site_components: true`. A deliberate side-effect became a staging technique: the chrome-refresh deploys the whole site as an intermediate visual checkpoint (light chrome over old sections) before the full rebuild.
- **sources:** running_notes_scheme_to_components(55).md#Sy #Sz #Ta #Td; w4b_01_repoint.sql; w4b_04_trigger_item.sql; RUNBOOK_scheme_to_components(50).md#W4b-RESULTS
- **relations:** rerender-pages v6 workflow; rebuild vs rerender semantics; four chrome default stores.
- **verify-later:** render_site_components_action.go force_rerender/skip logic; idea.uk site_components rows point at active components.

### Rebuild vs rerender semantics and stale-render fossilisation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** CHECK 4 RESULTS: "the deployed pages are RERENDER output carrying stale stored renders … deployed hero consumes legacy `var(--accent-color, #0f3460)` … A full page-build-handler rebuild is required; `needs_rerender` would re-fossilise them" — settling the 016-vs-026 documented tension "by direct evidence".
- **what:** Two distinct page-refresh routes with different semantics: `needs_rerender` (handler rerender-pages) reassembles stored `page_components.rendered_html` and injects stored chrome — it does NOT re-render component templates, so template changes never land and old renders fossilise; a full build (`site_work_items` insert: pipeline=build, handler_agent=page-build-handler, status=triaged) re-runs plan_sections and re-renders everything. idea.uk had lived for weeks on reassemblies of early renders while the library advanced — the fossil tell is a legacy variable name in deployed HTML (`var(--accent-color`), and its count going to 0 is the proof a rebuild truly re-rendered. Docs 026 ("rerender reflects new template") and 016 ("patches in place") disagreed; evidence sided with 016. Related hazard from the migration sketch: a content rebuild can de-tool a tool page (page-content-writer regenerates from plan_sections, which does not know the interactive tool).
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS #Migration-backfill; running_notes_scheme_to_components(55).md#So #Sh(migration route); HANDOFF_scheme_to_components_for_claude_code(1).md#Invariant (item 5)
- **relations:** dual chrome render paths; work-item crafting conventions; deployed-binary-predates-disk class.
- **verify-later:** rerender-pages vs page-build-handler workflow definitions in agent_definitions; 016/026 doc reconciliation.

### rerender-pages v6 workflow (refresh_site_components gate)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Tb): "Workflow (v6) fully read: gate `input_data.spec.refresh_site_components == true` → FORCED chrome render (header/footer/head) → js snippets render+commit → blog listing → get_pages (deployed+active) → create_rerender_items (per-page) → update_site_status deployed → complete."
- **what:** The site-wide rerender agent: one work item fans out to a forced chrome re-render (the only workflow passing `force_rerender: true` — pageflow-builder and site-work-orchestrator pass false, which explains fossilised chrome surviving full builds), JS snippet rendering, blog-listing rebuild, then per-page rerender items that the build dispatch loop drains; it ends by marking the site deployed. `spec.function`/`component_id` are consumed nowhere in v6. The real producer of such items is `store_generated_component` on regeneration (one deduped `needs_rerender` per affected site, item_key `component_regen_rerender:<uuid>`).
- **sources:** running_notes_scheme_to_components(55).md#Ta #Tb; w4b_02_read_triggers.sql; w4b_04_trigger_item.sql
- **relations:** dual chrome render paths; rebuild vs rerender semantics; work-item crafting conventions.
- **verify-later:** agent_definitions type='rerender-pages' version 6 default_config; check_refresh_components step.

### Layout CTA pair curation with WCAG contrast gates
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Notes (Sq) "W1 complete + verified"; (Su) "W1b COMPLETE: five layouts curated"; w1b comments record the five hex swaps and expected values.
- **what:** W1 added the missing CTA pair to tool-portal-light via anchored `regexp_replace` on the verified `--color-footer-text` line (`{{palette "cta_bg" "#e9e2d3"}}` + `cta_text "#1a1a1a"`, contrast ≈13.5, mirroring tool-portal-dark's neutral elevated band; accent alternative offered). A sweep computed every layout's cta pair contrast; five seed layouts failed 4.5 with white text and got same-hue darker fallbacks (W1b, zero live impact — no site uses them). Pair values are deliberate per-layout design: several light layouts curate DARK footer bands — "light site, dark band by choice" is already a curated model in the library. Requirement carried into the contract: pair contrast ≥ 4.5.
- **sources:** w1_01_add_cta_pair.sql; w1b_01_contrast_batch.sql; RUNBOOK_scheme_to_components(50).md#W1-RESULTS #CHECK-4-RESULTS (4b); SPEC_scheme_to_components.md#W1
- **relations:** paired-variable standard; three-part styles.css assembly.
- **verify-later:** layouts css_template cta pair values for the six touched layouts.

### Phase 4.5 data-section-bg surface generalisation (deferred)
- **category:** styling-render-pipeline
- **status-signal:** aspirational
- **status-evidence:** SPEC consequences: "025 Phase 4.5 (`data-section-bg` surface generalisation) is deferred as a separate dark-site concern."
- **what:** Doc 025's already-designed decouple: components carry a `data-section-bg="surface"` attribute; the renderer replaces its hardcoded 5-surface-class list with an attribute selector; dual-write migration. The thread audited it seriously (Si prior-art pass) and then argued it down: it solves a dark-site generalisation idea.uk never hits, its blanket "never self-declare" conflates hazardous surface declarations with load-bearing band declarations, and renderer ownership reintroduces component intent one hop away. Remains the designed answer for dark sites with surface sections outside the hardcoded 5, if that ever bites.
- **sources:** running_notes_scheme_to_components(55).md#Si #Sk #Sm; SPEC_scheme_to_components.md#Decision-record; HANDOFF_scheme_to_components_for_claude_code(1).md#Questioning-025
- **relations:** buildSectionDefaults (the 5-class list it generalises); paired-variable standard (chosen instead).
- **verify-later:** docs 025 §427–505; any data-section-bg attribute usage in components.

### Scheme-coherence audit guard (Q8)
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** PLAN Q8 "A scheme-coherence check in the design auditor / improvement loop … Status: open (I)"; absent from the SCHEME CLOSE remaining-work map and every later position note — silently dropped after the paired-variable close.
- **what:** The proposed regression guard: an auditor/improvement-loop check flagging "section scheme does not match site scheme / unintended contrast" so the scheme→components fix cannot silently regress. Designed as fix-shape item 8, never specified or built; the eventual regression protection took a different form (contract in the creator prompt + the re-aimed fixer as mechanical enforcer), leaving no dedicated audit check.
- **sources:** PLAN_scheme_to_components(1).md#Q8 #Provisional-fix-shape; RUNBOOK_scheme_to_components(50).md#SCHEME-CLOSE (absence)
- **relations:** fixes-at-initial-render principle; fix_forced_text_colours re-aim (partial substitute).
- **verify-later:** design-auditor checks list — does any scheme-coherence rule exist?

### Explicit RenderContext.Scheme signal (Q1)
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** Notes (Sf): "explicit `RenderContext.Scheme` is SECONDARY … This revises the Q1 emphasis in the PLAN"; never implemented anywhere in the executed fix.
- **what:** The original leading design (Q1/Q3): plumb the resolved scheme explicitly into both render entry points — `l.scheme` in the CSS loader SELECT + `themeComposition.Scheme`, and a `Scheme` field on `RenderContext` exposed via `contextToInterfaceMap` — so component templates receive a light/dark signal. Overtaken when Check 1 showed the scheme already reaches components implicitly through the palette `:root` values and luminance defaults; the components were the only thing defeating an already-working system. No scheme field was ever added.
- **sources:** PLAN_scheme_to_components(1).md#Q1; running_notes_scheme_to_components(55).md#Sb #Sf #Sk
- **relations:** superseded by paired-variable standard + implicit palette mechanism.
- **verify-later:** RenderContext struct (component_library.go) — confirm no Scheme field exists.

### Improvement-loop colour/nav fixer suite (pre-re-aim state)
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** Notes (Sj) full read of the fixers; HANDOFF: "As aimed today they are scheme-blind and ENFORCE the component-owns contract (`fix_forced_text_colors` injects dark `--section-*` into is_dark_section components)."
- **what:** The established fixer infrastructure: color-variable-fixer agent runs `fix_hardcoded_colors` (dark background hex → `var(--color-primary)`; dark 2-stop gradients → primary/secondary; deliberately leaves `rgba(0,0,0,x)` overlays alone; fixes both template and rendered HTML) and `fix_forced_text_colors` (strips forced child text colours so elements inherit the chain, WCAG-validates ≥4.5, and — the superseded part — injected the white `--section-*` contract into is_dark_section components); nav-link-fixer runs `fix_nav_link_templates` (with the documented rule that `render_site_components force_rerender` must follow); `fix_component_template` routes on fix_type (inject_nav_flex_css, remove_element, align_slot_name, inject_responsive_css, repair_template_slots) — symptom fixes for exactly the dark fallback header's output. Running them as-was on idea.uk would have entrenched dark.
- **sources:** running_notes_scheme_to_components(55).md#Sj; HANDOFF_scheme_to_components_for_claude_code(1).md#Established
- **relations:** superseded by fix_forced_text_colours re-aim; fix_hardcoded_colors retained (its hex→primary mapping stays coherent with the paintPaletteBand class).
- **verify-later:** fix_harcoded_colours_action.go, fix_nav_link_templates_action.go, fix_component_template_action.go current behaviour.

### fix_forced_text_colours re-aim: painting classifier + declaration rewriter
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** gobatch_04(2) delivers whole-function Edit G + new classifier code, with "RESOLVED (2026-07-06): … switched to `var(--color-primary-text, …)`"; RUNBOOK 07-06 night: "Slice 3 (the re-aimed fixer): deploying per your note — confirm it built"; supervised first run still pending.
- **what:** The backstop fixer rebuilt around the new contract: a `paintClass` classifier (paintAmbient/paintPair/paintInk/paintPaletteBand) derives what a template's own CSS paints from regexes over its style blocks — never from `is_dark_section` (the parameter is kept but deliberately ignored so call sites stay unchanged); `rewriteSectionDeclarationsInHTML` converts literal `--section-*` declarations to the class-appropriate references (pair text, hero ink, on-colour family, color-mix derivatives) and deletes declarations from ambient non-painters; the proven literal-stripping machinery and the WCAG contrast gate are retained; the old contract-injector trio (`ensureSectionContractInHTML`/`injectSectionContract`/`sectionContractRe`) is deleted. `result.contractAdded` is repurposed as a rewrite counter (name kept per the no-rename rule, meaning shift noted).
- **sources:** gobatch_04_fixer_reaim(2).md; running_notes_scheme_to_components(55).md#Ud #Ue #Ug; SPEC_scheme_to_components.md#W5
- **relations:** section painting contract (the thing it enforces); supervised fixer first-run protocol; is_dark_section demotion.
- **verify-later:** fix_forced_text_colours_action.go deployed body (classifier present, injector trio gone); first live run's details JSON.

### Fixes-land-at-initial-render principle (loop fixers backstop only)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Notes (Sk) "User steer … Fix at INITIAL RENDER (library + composition/renderer). The improvement-loop fixers stay AVAILABLE as a backstop but must NOT be *required* for new builds"; SPEC W5: "The fixers remain improvement-loop backstops — never required for a correct first render."
- **what:** A governing platform principle set mid-thread and carried into every artefact: correctness must come from the library, composition, and renderer at first render; improvement-loop fixers are post-hoc safety nets dispatched on audit findings, never a required step for a new build. This ruled out "make the fixers scheme-aware" as the primary mechanism and shaped where each fix landed (templates and prompts, not loop passes).
- **sources:** running_notes_scheme_to_components(55).md#Sk; SPEC_scheme_to_components.md#W5; HANDOFF_scheme_to_components_for_claude_code(1).md#Established (USER DIRECTION)
- **relations:** section painting contract; component-creator prompt re-aim; scheme-coherence guard (abandoned alternative).
- **verify-later:** whether any build workflow invokes colour fixers as a required step.

### Supervised fixer first-run protocol (disposable specimen site)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK STEP D (2026-07-06 night+): "watch the re-aimed fixer act on a real site once, under supervision, before it is ever allowed near the improvement loop or a second site" — not yet run at last dated evidence.
- **what:** Rollout protocol for a re-aimed automated fixer: confirm the deployed pod carries it; capture the specimen's before-state (dartsonline, a disposable freshly-built non-target site, site_id 5fe8785b); spawn manually via the 016b harness; judge the returned per-component `details` JSON (declarations rewritten per class, literals stripped, contrast-gate skips) rather than the render; re-read components and diff; only then decide hand-needles vs fixer for the library tail, and only then allow the improvement loop near it. Includes the deferred plan for a guarded 7-table cascade delete of a messed-up disposable site (written against schema when needed, never ad-hoc). The sizing read found dartsonline a thin specimen (all components literal-free; only two literal text hexes).
- **sources:** RUNBOOK_scheme_to_components(50).md#STEP-D; stepD_and_pages_reads.sql; running_notes_scheme_to_components(55).md#Uj #Uk
- **relations:** fix_forced_text_colours re-aim; hazard/band tail; debugging harness (016b).
- **verify-later:** whether Step D ran; dartsonline component state; the cascade-delete script's existence.

### component-creator prompt re-aim (painting rules, vocabulary, image-fields rule)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Notes (Ui): "4a evidence: gate t/t/t/t/f → UPDATE 1 t/t/t/f" (2026-07-06); slice4a_creator_prompt.sql RETURNING confirms SECTION PAINTING + IMAGE FIELDS RULE in, old DARK SECTIONS block gone.
- **what:** Four targeted needle replaces inside `agent_definitions.default_config->>'prompt_template'` for component-creator: the dark-sections literal block becomes the four painting models (references only; is_dark_section reported honestly but "nothing may style from it"); the consumer chain replaces the dark line; item 7's vocabulary gains the cta pair and extended tokens (surface-alt, hairline, code-bg, callout pair); Tier C gains the image-fields rule (site_assets.* fields required:false + skip_field + gated markup — described rather than shown because the prompt is itself Go-template-rendered and literal if-syntax would execute). Root cause it addresses: the generated half-migrated footer and brief-explanation proved the prompt was emitting components that consume chrome vars while self-declaring dark text — drift continues until the contract lives in the prompt.
- **sources:** slice4a_creator_prompt.sql; running_notes_scheme_to_components(55).md#Uf #Uh #Ui; RUNBOOK_scheme_to_components(50).md#CHECK-2-RESULTS (corollary)
- **relations:** section painting contract; agent re-registration vs re-seed risk; image fields optional-with-gate.
- **verify-later:** agent_definitions component-creator prompt_template current text; the Step C grep ("DARK SECTIONS (if the section has a dark background)" in Go sources).

### Agent re-registration vs re-seed risk (DB row authoritative)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** RUNBOOK 07-06 night: "deploys bump `agent_definitions.updated_at` without overwriting `default_config` (the old prompt survived today's deploy)… the user confirms component-creator is a dynamically spawned/registered agent, not a YAML-seeded one. So the DB row is authoritative."
- **what:** A durability model for DB-edited agent prompts: chassis deploys re-register agent definitions (bumping `updated_at`) but do not overwrite `default_config`, so SQL edits to prompts survive deploys for dynamically registered agents. The residual risk is an in-code prompt template driving an upsert; the check is one grep for a literal fragment of the OLD prompt in Go sources — a hit means mirroring the edit in code, no hit means nothing can revert the row. (Earlier drafts had a heavier "seed check" over configs/deployments YAML; superseded by the user's confirmation.)
- **sources:** RUNBOOK_scheme_to_components(50).md#STEP-C; RUNBOOK_scheme_to_components(49).md (the earlier five-minute variant, family-delta); running_notes_scheme_to_components(55).md#Uf #Uj
- **relations:** component-creator prompt re-aim; 019 idempotent prompt migration pattern.
- **verify-later:** run the Step C grep; agent registration code path (upsert semantics on default_config).

### Section-scope imagery pipeline (plan → emit → generate → deploy → rebuild)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** RUNBOOK §W7: "The flow already exists (05-26 handoff + code): `write_site_plan` → `site_plan_imagery` rows … → `emit_imagery_items` … → `needs_imagery` (priority ≤98) → `image-build-handler` … → `flag_page_image_rebuild` → `needs_page` (99) → rebuild resolves the URL"; W7b exercised it end-to-end (assets active at B2 in ~3 minutes each, 2026-07-03).
- **what:** The dynamic imagery supply chain: the planner writes `site_plan_imagery` rows (scope site/page/section; scope_ref `page:ordinal` for sections; key; kind hero/logo/icon/illustration/infographic; authored prompt — the table has NO description column; ordering; source), the gap-driven `emit_imagery_items` emits `needs_imagery` items only where no asset exists, image-build-handler's 25-step workflow generates, stores, brand-checks, spawns the asset-deployer (download S3 → optimise by purpose → commit to the site git repo, key-named files `_`→`-`.jpg), then `flag_page_image_rebuild` emits `needs_page` so plan_sections re-resolves the now-present asset. For idea.uk it never fired for the brief-explanation illustration simply because the planner never requested one (16 rows: 5 heroes, 10 icons, logo). Ordinal-based scope_refs drift when plans reorder (hygiene note; resolution is by key).
- **sources:** RUNBOOK_scheme_to_components(50).md#W7 #W7-0.3/0.4-RESULTS; w7b_01_imagery.sql; running_notes_scheme_to_components(55).md#Th #Tj #Ty #Tz
- **relations:** ensureAssets resolution gap; flag_page_image_rebuild section scope; presigned-URL expiry.
- **verify-later:** site_plan_imagery schema + emit_imagery_items step in build-site-planner; image-build-handler workflow steps.

### ensureAssets section-scope resolution gap (Edit B)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Notes (Tq): "Edit B LIVE (tools t)"; (Tv): "BRIEF-EXPLANATION CLOSED … illustration renders on index AND tools; Edit B fine throughout."
- **what:** The structural gap that made section illustrations unresolvable: `ensureAssets` (plan_sections_action.go) loaded only the page hero and site logo into the resolver's assets map (plus a legacy content_data fallback), so `site_assets.<key>` for section-scope imagery could never resolve — the pipeline's "last inch" was never wired. Edit B adds a third query (spi scope='section', scope_ref LIKE page||':%', joined to active assets), mapping BOTH by key (per-key schema paths like icon sets) AND by kind first-wins alias (generic `site_assets.illustration` paths), modelled on the hero block. The two-day "index miss" after deployment turned out to be a probe artefact (grep for the asset key string, but objects are UUID-named), taught as a debugging lesson.
- **sources:** gobatch_01_plan_sections.md#Edit-B; RUNBOOK_scheme_to_components(50).md#W7-CODE-FINDING; running_notes_scheme_to_components(55).md#Ti #Tt #Tu #Tv
- **relations:** section-scope imagery pipeline; plan_sections field deferral; probe-blindness (SQL pitfalls).
- **verify-later:** plan_sections_action.go ensureAssets section query; rendered brief-explanation `<img>` src on index/tools.

### flag_page_image_rebuild section-scope mapping (Edit H)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** RUNBOOK 07-06 night: "Slice 4c: PENDING APPLY — `gobatch_05_flag_section_scope.md` now carries Edit H … AND Edit H2"; the companion step-description SQL landed (Uj: "4c step description UPDATE 1") but the code edit awaited the next commit/image.
- **what:** `flag_page_image_rebuild` no-ops for non-page scope, so section-scope imagery landings never triggered the page rebuild that would surface them (observed live: zero flag-created needs_page in 30h after the two illustrations landed). Section scope_refs carry the page as a prefix (`index:1`), so the fix is a prefix-split: map scope 'section' to its page and fall through to the existing page path — no new emit code. Edit H2 + slice4c align the file header comment and the agent-definition step description with the new behaviour (cosmetic-drift discipline: descriptions must match deployed behaviour).
- **sources:** gobatch_05_flag_section_scope.md; slice4c_step_description.sql; running_notes_scheme_to_components(55).md#Tp #Ui #Uj
- **relations:** section-scope imagery pipeline; work-item crafting conventions.
- **verify-later:** flag_page_image_rebuild_action.go deployed body; image-build-handler flag_rebuild step description text.

### Image fields optional-with-gate contract
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** User decision (Th, 2026-07-03): "imagery must not block section rendering"; W7a gate applied (Tm: "UPDATE 1, gated t/t"); the rule landed in the creator prompt (4a) and 003 6b; `illustration_url` was already required:false + skip_field in schema.
- **what:** Any `site_assets.*`-sourced component field MUST be `required: false` with `on_missing: skip_field`, and its markup MUST be gated with a template conditional (`{{if .illustration_url}}` around brief-explanation's image wrapper is the model — Go templates treat "" and missing as false, covering the src="" broken-image case). Imagery arrives asynchronously and must never block or defer a section; the section renders imageless and the image is added by the pipeline's own queued rebuild. Codified as 003 item 6b and the creator prompt's IMAGE FIELDS RULE.
- **sources:** w7a_01_gate.sql; slice4b_003_contract.md#Edit-1 (6b); slice4a_creator_prompt.sql (R4); running_notes_scheme_to_components(55).md#Th #Tl #Tm
- **relations:** plan_sections field deferral semantics; section-scope imagery pipeline; component-creator prompt re-aim.
- **verify-later:** brief-explanation html_template gate; input_schema on_missing values across image-consuming components.

### Presigned-URL expiry and deploy-time asset localisation
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Notes (Ud) 2026-07-06: "w9_06: t/f/f both pages … localisation verified on content; Edit F deployed → recurrence prevention live → THREAD CLOSED"; w9_04 RETURNING "UPDATE 18, every url now /assets/images/…".
- **what:** A whole failure class: `assets.url` stored the presigned B2/S3 URL from generation (X-Amz-Expires=604800 — dies in seven days), while the asset-deployer had already committed the optimised file into the site repo under a key-derived local name. Renders that resolve from assets.url therefore embed URLs that die; heroes escaped only by being shadowed by a legacy local path. The fix is two-sided: w9_04 backfill flips all 18 idea.uk rows to `/assets/images/<key-hyphenated>.jpg`, preserving the unsigned S3 object path into `storage_path` (+ storage_provider), then a rebuild; Edit F makes `deploy_image_asset` record the committed local URL on the asset row at every future deploy, for ALL kinds (best-effort — a failure must not fail the deploy). Applies platform-wide to any site without the legacy hero_url shadow.
- **sources:** w9_03_assets_schema_and_inventory.sql; w9_04_backfill_flip.sql; gobatch_03_deploy_asset_localise(1).md; running_notes_scheme_to_components(55).md#Tu #Tw #Tz #Ua #Ub #Ud
- **relations:** legacy hero_url shadow; section-scope imagery pipeline; storage-architecture (S3/B2 refs preserved in storage_path).
- **verify-later:** deploy_image_asset_action.go post-commit UPDATE; assets rows url vs storage_path forms across sites.

### Legacy site-level hero_url shadow (last-write-wins per purpose)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Notes (Tz/Ub): "hero_url lives at SITE level (sites.content_data …), merged beneath section data; template `or` picks it over the presigned background_image … last-write-wins per purpose → site-wide hero currently = hero-about's image everywhere."
- **what:** A legacy mechanism that both saved and distorts heroes: image deploys historically wrote `purpose+"_url"` keys (e.g. hero_url) into site-level `sites.content_data`, which the ContentData-priority merge (component_library.go ~736) supplies to templates ahead of the schema-resolved `site_assets.hero` value. Consequences: hero renders stayed local-path (immune to presigned expiry) but every page shows the same last-written hero image; the per-page hero assets sat unconsumed. Banked as a known quirk (per-page heroes = a later improvement); render-neutral to the localisation flip.
- **sources:** running_notes_scheme_to_components(55).md#Tx #Ty #Tz #Ub; w9_02_deployer_and_shadow.sql; w8_09_hero_exposure.sql
- **relations:** presigned-URL expiry; ensureAssets (content_data fallback is gap-fill, hero/logo only).
- **verify-later:** sites.content_data hero_url keys; component_library.go merge priority; whether per-page heroes were ever wired.

### plan_sections field deferral semantics and needs_section_data escalation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (To): "Section BACK + both escalations self-closed = the deployed skip_field behaviour works"; W6.4: "two `needs_section_data` items in `needs_human_review` — `plan_sections` could not resolve `illustration_url` … and built each page WITHOUT the section."
- **what:** plan_sections resolves each schema field per its declared `source`; unresolvable required fields defer the WHOLE section, escalate a `needs_section_data` work item into needs_human_review, and the page builds without the section — a loud drop, not silent (guide refinement: fossil pages had been hiding the unresolved dependency). `on_missing: skip_field` is the established optional pattern: omit the field, let the template gate handle it. Edit A fixed the smell that a REQUIRED field with on_missing:skip_field fell to the default defer branch instead of honouring the declared intent. `closeResolvedDataRequest` self-closes escalations once the field resolves post-deploy.
- **sources:** gobatch_01_plan_sections.md#Edit-A; RUNBOOK_scheme_to_components(50).md#W6.4 #W7-FINDINGS; running_notes_scheme_to_components(55).md#Tg #Tl #To; w6_05_section_data_read.sql
- **relations:** image fields optional-with-gate; section data source triad; deployed-binary-predates-disk.
- **verify-later:** plan_sections_action.go on_missing switch (required branch skip_field case present); needs_section_data item lifecycle.

### Deployed-binary-predates-disk failure class
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Notes (Tm): "Fork RESOLVED: extraction sound … → the on-disk code cannot produce the July-2 escalation → deployed predates disk; the skip_field fix exists and never shipped."
- **what:** A named diagnosis class: observed behaviour contradicts a correct code read because the running pod's image predates the working copy — the fix exists on disk and never shipped. Diagnostic: `git log -1 -- <file>` vs the running pod's image age; remedy: deploy the working copy. Sibling lessons from the same threads: verify the running image contains an edit before debugging it ("success path silent"), and prefer a forward test (one clean build + read both render and stored data) over forensic reconstruction of overlapping rebuild windows.
- **sources:** running_notes_scheme_to_components(55).md#Tl #Tm #Tt; RUNBOOK_scheme_to_components(50).md#W7-FINDINGS; w8_07_fresh_index_build.sql (the forward-probe pattern)
- **relations:** chassis deploy model; plan_sections deferral (the instance).
- **verify-later:** 016b guide entry for this class.

### Section data source triad and reconcile_section_data
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** HANDOFF (2026-06-19): "`reconcile_section_data` IS wired — registry.go line 914 … description 'Re-trigger pages whose deferred section data is now query-resolvable'" (correcting a stale note that it was not wired).
- **what:** A component's content comes from one of three sources, and fixes differ per case: (1) query-resolvable section data (the tools/guides-list kind — the reconciler's scope: `ReconcileSectionDataAction` re-triggers pages whose deferred data has become resolvable), (2) a human-entered spec field (e.g. pricing tier_1_* from `site_specs.pricing` — the reconciler correctly skips these), (3) page-content-writer prose (LLM-generated). The differentiators investigation established the triad as the diagnostic frame — and then found the actual fault was in none of the sources (a key-naming mismatch). Incidental same-thread finding: `write_site_spec` errors "missing required fields: [spec_data]" on persist_mission/roadmap — the action input is spec_data but the column is `data` (site_specs is aspect + data jsonb, UNIQUE(site_id,aspect) WHERE is_current).
- **sources:** HANDOFF_idea_uk_differentiators_section_data.md; bundle3; running_notes_scheme_to_components(55).md#Sa #Sh (corrected facts)
- **relations:** array item-fields contract (the real fault); plan_sections deferral.
- **verify-later:** reconcile_section_data_action.go scope logic; registry.go wiring; site_specs schema.

### Array item-fields prompt contract (019 migration + ItemFields)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu) 2026-06-21: "Prompt migration already applied"; 019 migration NOTICE "prompt patched"; checkpoint (ss) documents the root cause and fragments verified at positions 2330/3402.
- **what:** Root cause of the differentiators empty cards: the page-content-writer prompt listed array fields with type only, never element shape, so the LLM guessed item keys — `title`/`body` against a template reading `name`/`description` renders empty; FAQ worked only because the natural guess happened to match. Fix: `plan_sections` gains `ItemFields []string` on `llmFieldSpec` via `extractArrayItemFields` (reads both `items` and `item_schema`, sorted for stable prompts); the 019 migration patches the prompt's What-To-Write line and makes the Output-Format JSON skeleton type-aware (`[{ "k": "..." }]` for arrays). The migration is order-independent with the Go deploy ({{if .item_fields}} is simply false until populated), idempotent via a sentinel, aborts if fragments moved, and ships a paired down-migration.
- **sources:** running_notes_checkpoint_ss(1).md; 019_pcw_prompt_item_fields.sql; plan_pcw_item_fields_fix(1).md; RUNBOOK_pcw_item_fields_fix.md
- **relations:** render-time item-key reconciler; component schema-template invariant; SQL change-management pattern.
- **verify-later:** agent_definitions page-content-writer prompt_template markers; plan_sections_action.go ItemFields population.

### Render-time item-key reconciler (schema-sourced, non-fatal)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Checkpoint (uu): "Three artefacts now final in outputs … code awaits a chassis image bump" (2026-06-21); no later doc in this unit confirms the image bump for this specific change.
- **what:** A belt-and-braces safety net in `RenderComponentAction`: before the merge, `reconcileGeneratedItemKeys` remaps LLM-drifted array item keys onto the expected ones using case/separator-insensitive matching plus a synonym table (title/body → name/description etc.), never moving a synonym onto a key that is itself expected. Decision 1B hardened it to source expected keys from the component's own `input_schema` (fields with source:"llm" only) instead of the section plan — removing plan-freshness coupling and making the prompt change an optimisation, not a correctness requirement. Decision 2: unrecoverable misses ERROR-and-continue (a missing sub-field is cosmetic; failing a page build is higher blast-radius). Corrected content lands in both rendered HTML and persisted content_data. Cross-file deploy constraint: rides the same image as plan_sections' extractArrayItemFields.
- **sources:** running_notes_checkpoint_uu.md; running_notes_checkpoint_ss(1).md#Fix-delivered; RUNBOOK_pcw_item_fields_fix.md#4-Logs
- **relations:** array item-fields contract; component schema-template invariant; needs_llm routing.
- **verify-later:** v3_site_actions.go reconcileGeneratedItemKeys + wire-in; whether the carrying image shipped (log lines "reconcileGeneratedItemKeys" in writer pods).

### needs_llm routing via detectNeedsLLMContent
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (ss): "the writer sub-workflow … branches on `render_mode == 'agent' OR needs_llm == true`. `needs_llm` is computed by `detectNeedsLLMContent` (v3_site_actions.go ~4095), which returns true for any non-empty `input_schema`."
- **what:** How a section reaches the LLM generation path: the page-content-writer's `process_sections_loop` routes on render_mode OR the computed needs_llm flag, and because detectNeedsLLMContent returns true for any non-empty input_schema, template-mode components with schemas still get LLM content. This made an investigative render_mode flip harmless to revert (differentiators back to 'template') and explains why a 'template' component had generated content at all.
- **sources:** running_notes_checkpoint_ss(1).md#What-we-established #Correction-logged
- **relations:** section data source triad; array item-fields contract.
- **verify-later:** v3_site_actions.go detectNeedsLLMContent; writer sub-workflow branch config.

### Component schema-template consistency invariant
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu): "The governing invariant — a component's schema `items` must match its template tokens — is the right thing to hold; the reconciler enforces consistency toward the current schema."
- **what:** A component's `input_schema` (shape `{"fields": {...}}`, unmarshalled in component_library.go) is the contract for its `html_template` tokens: array item field names in the schema must match what the template reads. The reconciler, the prompt, and generation all derive from input_schema, so divergence breaks all three coherently. Known violation: info-card-grid's stored html_template literally contains `<no value>` (rendered-against-nil output apparently written back into the template column) — flagged as its own repair thread and never fixed inside this unit. services-grid shares differentiators' schema byte-identically and was healed by the same fix.
- **sources:** running_notes_checkpoint_uu.md#Confirmed-during-the-hardening-review; running_notes_checkpoint_ss(1).md#Root-cause-in-code; RUNBOOK_pcw_item_fields_fix.md#Follow-on
- **relations:** array item-fields contract; render-time reconciler.
- **verify-later:** info-card-grid html_template (still `<no value>`?); services-grid first-use spot check.

### No component-level regeneration trigger (whole-page rebuild remedy)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu): "**No component-level regeneration trigger exists** (user confirmed). So the remedy for the already-deployed broken cards is a whole-index `page-rebuild`, which regenerates *all* index sections … Accepted as the cost."
- **what:** A platform limitation shaping every content-fix decision: there is no mechanism to regenerate one component on one page; the only remedy for bad stored content is a full page rebuild, which rewrites every section's copy (copy churn on hero, FAQ, narrative accepted as cost). Repeatedly parked on the hygiene/backlog lists; interacts with rebuild-vs-rerender (rerender can't be used because it reassembles stored HTML).
- **sources:** running_notes_checkpoint_uu.md#Decisions-taken; RUNBOOK_pcw_item_fields_fix.md#3
- **relations:** rebuild vs rerender semantics; content-governance (regeneration).
- **verify-later:** whether any component-scoped regen item type has since appeared in site_work_items vocabulary.

### Planner re-plan union safety (normaliseRealisedToPlanPage)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Checkpoint (Un) 2026-07-07: "normaliseRealisedToPlanPage (v3_site_actions.go:4383) exists so a re-plan LOADS realised pages …, converts them to plan-page shape CARRYING their sections, and UNIONS with the LLM proposal — its own comment: without carrying sections the upsert would clobber built pages."
- **what:** Site composition is whole-plan and LLM-driven: build-site-planner (consuming needs_site_plan) supersedes the current site_plans row and rewrites site_plan_pages + site_plan_sections. Re-running it is safe by design because load_existing_pages surfaces realised pages and the normaliser unions them (with their sections) into the new plan — built pages keep their composition while catalogued-but-uncomposed pages get composed. This makes "emit needs_site_plan" the structural route for composing missing pages, versus hand-INSERTing plan rows (which drifts nav/plan/page consistency).
- **sources:** running_notes_scheme_to_components(55).md#Un; stepF_replan_read.sql
- **relations:** planned-but-uncomposed pages gap; work-item crafting conventions.
- **verify-later:** v3_site_actions.go normaliseRealisedToPlanPage; build-site-planner workflow steps.

### Planned-but-uncomposed pages gap (catalogued, never composed)
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** Checkpoint (Ul): "the three planned pages have NO site_plan_sections rows; their pages.sections = []. Catalogued, never composed"; (Un) ends with the replan-read staged — the emit had not run at the unit's last dated note (2026-07-07).
- **what:** A distinct failure shape: pages rows exist with page_type and nav intent set (news-index, guides-index, tool-audience-check on idea.uk), so navigation links to them and 404s, but they carry empty sections and no plan rows — the LLM plan behind the current site_plans row never included them. A W6-style needs_page emit would build an empty page; the correct route is two-phase: planner re-run composes them (union-safe), then needs_page builds and deploys. Also surfaced the distinction between query-backed index pages (news/guides may be fed by the blog-listing mechanism) and static pages, and reuse of the already-deployed audience-check tool component.
- **sources:** running_notes_scheme_to_components(55).md#Uk #Ul #Um #Un; RUNBOOK_scheme_to_components(50).md#PLANNED-PAGES; stepD_and_pages_reads.sql (block B/C)
- **relations:** planner re-plan union safety; navigation (nav 404s); rebuild vs rerender.
- **verify-later:** idea.uk pages rows for the three; site_plan_sections presence; whether the needs_site_plan emit ran.

### Work-item crafting conventions (real shapes, truthful provenance, dedup)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** w4b_04 comments: "crafted from the real rows … with truthful deviations noted: source 'manual' and created_by 'w4b_chrome_refresh' … lying in provenance columns costs later debugging"; every W6/W7/W8/W9 insert repeats the pattern.
- **what:** The discipline for hand-inserting `site_work_items`: copy the metadata of real rows produced by the owning code path (pipeline/severity/priority/handler_agent/status), deviate only truthfully in provenance columns (source=manual, created_by=<script name>), carry only spec fields the consuming workflow actually reads, and dedup check-first with a NOT EXISTS that mirrors `idx_swi_dedup` exactly (non-terminal statuses only — including 'unresolved', a status the index taught the thread it had missed). Item_key families are stable conventions: `page_rerender:<page>`, `chrome_refresh_rerender:<site_id>`, `needs_imagery:section:<scope_ref>:<key>`, `component_regen_rerender:<uuid>`, `section_data_*`. The check-first pattern is borrowed from CreateNeedsNewComponentItem.
- **sources:** w4b_04_trigger_item.sql; w7b_01_imagery.sql; w8_01_post_deploy_rebuild.sql; running_notes_scheme_to_components(55).md#Tb #Tc
- **relations:** rerender-pages v6; work-item claim/retry; scheduler-and-tasks.
- **verify-later:** idx_swi_dedup definition; site_work_items status vocabulary.

### Work-item claim/retry behaviour and the claim-timeout class
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** W6 FINAL VERIFY: "3.1 failure class: `Claim timed out — handler pod likely died` on all three retried items — dispatch infrastructure, not the template changes; retries recovered."
- **what:** Build items are claimed by the dispatch loop and retried on claim timeout; heavy page builds (19:18–22:45 for six pages) collide with claim durations, producing retried-then-complete items whose error text is retained — read the error class before calling retries healthy. Observed hygiene gaps: `site_work_items.updated_at` stays frozen at insert through claim/retry/completion (same family as the pre-trigger layouts.updated_at); a deploy can release claims mid-flight (claimed→triaged). All parked on the hygiene list, not actioned in-thread.
- **sources:** RUNBOOK_scheme_to_components(50).md#W6-FINAL-VERIFY; w6_03_final_verify.sql; running_notes_scheme_to_components(55).md#Te #Tf #Tp
- **relations:** work-item crafting conventions; debugging (pod health).
- **verify-later:** build dispatch loop claim timeout vs typical build durations; updated_at handling on site_work_items.

### SQL needle-gate surgery pattern (guarded, idempotent, reversible DB edits)
- **category:** NEW:sql-change-management
- **status-signal:** deployed
- **status-evidence:** Practised in every w1–w9/slice4a file; codified in notes (Sv): "the needle-gate rule REFINED in 016b_7 (count expectations mechanically from the dump; mismatch = drift OR bad expectation)"; slice4b: "needle discipline applies to docs too."
- **what:** The unit's dominant change-management method for production data edits (templates, layouts, prompts, docs): dump the current bytes; derive byte-exact needles per element (multi-line E'…\n…' needles to disambiguate repeated strings; `position()` where needles contain literal `%`); run a read-only GATE that asserts each needle's presence and mechanically-derived counts; apply nested exact-string `replace()` (or `\1`-anchored regexp_replace) UPDATEs guarded on a pre-state marker so re-runs are 0-row no-ops; RETURNING booleans as immediate post-conditions; separate verify and inverse-rollback files plus a full .bak dump before every mutation. The 019 migrations extend the pattern to agent prompts: sentinel-guarded idempotency, abort-if-fragments-moved, paired down-migration. Sibling rules: `\set ON_ERROR_STOP on` for dependent mutation files; run SQL as files, never pasted into interactive psql.
- **sources:** w2_01_footer_fix.sql; w3b_01_hero_conversion.sql; slice4a_creator_prompt.sql; 019_pcw_prompt_item_fields.sql; running_notes_scheme_to_components(55).md#Sp #Sv #Ss
- **relations:** SQL pitfall class; agent re-seed risk; documentation-system (needle discipline on docs).
- **verify-later:** 016b guide's needle-gate entries; whether the pattern is written up as a standing convention doc.

### Postgres/SQL pitfall class (016b lessons)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Notes (St): "Guide updated → `016b_debugging_guide_7.md` (v4 log + three §9 entries …)"; each pitfall has an owned in-thread instance.
- **what:** The accumulated instrument-error catalogue this thread wrote into the debugging guide: Postgres ARE regex quantifier bounds max at 255 (`.{0,420}` is invalid — use substr+position); `substring(… from '(pattern)')` returns the FIRST CAPTURE GROUP, not the match; LIKE treats a needle's literal `%` as a wildcard (use position()); regexes like `background:\s*#` miss gradient-embedded hexes; a `0 rows` result is not decisive until the query and live state are checked (applies to one's own verification queries too); probes that grep for a key string are blind when objects are UUID-named; naive brace-counting false-fails on regex literals. Plus data-vocabulary lessons: sites.status vocabulary is draft/building/review/published/deployed/archived/error with legacy 'active'/'system' strays — never filter blast radius on status='active'.
- **sources:** running_notes_scheme_to_components(55).md#Sr #Ss #St #Sv #Sw #Tu #Ue; w2_02_verify_fixed.sql; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3f)
- **relations:** SQL needle-gate surgery; debugging guide 016b (the home doc).
- **verify-later:** 016b_debugging_guide_7.md §9 entries.

### cmd/bundle read-only context composer
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** Used for bundles 1–3 and the differentiators handoff; notes (Sa) record its failure modes ("-step framing emitting structure rather than source — use -step debug"; -doc paths must be ls-verified).
- **what:** The investigation tooling that assembled evidence bundles for this thread: `go run ./cmd/bundle` with an analysis JSON, `-root`, `-constitution`, `-step debug|framing|implementation`, `-task` (one-sentence brief), `-scope file[:Symbol]` code selections, `-include`, `-doc` paths, `-psql` connection command, `-schema-tables`, `-runtime-site`/`-runtime-page` live evidence, `-out`. Operational lore: `-step framing` yields signatures only; doc paths silently fail if wrong; bundles can arrive as thin slices (runtime data excluded) so live queries still need running separately.
- **sources:** 001_bundling_context.md; bundle3; RUNBOOK_scheme_to_components(50).md#Bundle-command; running_notes_scheme_to_components(55).md#Sa #Sh
- **relations:** docs019 contextkit (its home); check-based investigation method.
- **verify-later:** cmd/bundle source under docs019 go_files/contextkit; flag semantics.

### Chassis and site deploy model (single IMAGE_TAG; git → Actions → B2)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** README_001_flow_notes: "There's a single global IMAGE_TAG (currently v1.0.1066) and one agent-chassis binary that runs every dynamic agent … Rollback is symmetric and cheap"; checkpoint (uu) repeats the make targets.
- **what:** Two deploy surfaces: (1) chassis code ships as one image tag running every dynamic agent via agent_definitions config — targeted path `make quick-agent-update IMAGE_TAG=…` (build → push → kustomize → DB image_tag → restart agent-chassis) plus `make update-and-restart-orchestrator` for the generic-orchestrator statefulset; full `make release` bumps every service; rollback repoints to the old existing image without rebuild. (2) Site content deploys git → GitHub Actions → Backblaze B2. Operational wrinkle: full-build deploys commit as "Rerender: <page>" — the shared message format no longer distinguishes build from rerender.
- **sources:** README_001_flow_notes.md; running_notes_checkpoint_uu.md#Deploy-rollback; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; running_notes_scheme_to_components(55).md#Th (hygiene)
- **relations:** deployed-binary-predates-disk; agent re-registration.
- **verify-later:** Makefile targets; agent_definitions.image_tag column use.

### Orchestrator-agent architecture conventions
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Notes (Um) context reset restates them as standing rules: "every agent is an orchestrator owning a workflow of ≥1 steps that call ACTIONS. Children respond to the PARENT's responses_topic … Do NOT create sub-workflows in SQL — spawn sub-agents."
- **what:** The platform's structural conventions, carried as strict constraints through both threads: every agent is an orchestrator owning a workflow of steps that call Go actions; workflows stay simple with complexity in action code; no sub-workflows in SQL — spawn sub-agents with their own workflows (clean logs, separated responsibilities); child agents respond on the parent's responses topic; workflow variable names stay in sync with action expectations; identifiers are never renamed silently; `logger.Debug` is banned (invisible in the log pipeline — use Info); reuse/alter existing functions and architecture before creating new.
- **sources:** running_notes_scheme_to_components(55).md#Architecture-conventions #Um; HANDOFF_idea_uk_differentiators_section_data.md#House-rules; HANDOFF_scheme_to_components_for_claude_code(1).md#Constraints
- **relations:** house rules / standing preferences; platform mission.
- **verify-later:** agent-creation guidelines doc in repo; logging config that swallows Debug.

### House rules and standing preferences (the working contract)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Repeated verbatim at the top of the journal ("Standing preferences (STRICT)") and in both HANDOFFs so fresh chats inherit them.
- **what:** The user's cross-thread working contract, treated as binding by every agent session: Go not Python; British English; plain language, no hype/flattery, banned words "perfect/critical/excellent", no congratulation; confirm live schema/data before asserting or writing SQL; schema-first (`\d` before SELECT/UPDATE); structural framework fixes over one-off patches; low risk appetite, reasonable step sizes, ≤1 question per reply; no summary documents unless asked; don't call fixes final; no `*-light`/`*-dark` component variants; keep runbook + journal current; honest caveats including correcting one's own reads ("corrections owned").
- **sources:** running_notes_scheme_to_components(55).md#Standing-preferences; HANDOFF_idea_uk_differentiators_section_data.md#House-rules; HANDOFF_scheme_to_components_for_claude_code(1).md#Constraints
- **relations:** running-notes journal discipline; orchestrator conventions.
- **verify-later:** n/a (convention, not code) — check for a canonical repo home for these rules.

### Running-notes checkpoint journal discipline
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Journal header: "Memory is OFF; this doc is the journal. **Present this file at the END OF EVERY TURN.**"; 55 versions of an append-only file with lettered checkpoints (Sa…Un) prove the practice.
- **what:** The documentation method that makes long multi-session agent threads coherent without model memory: an append-only running-notes journal presented every turn, lettered checkpoints, a carry-over state block (preferences, architecture conventions, project facts, "the fix in one line"), explicit AWAITING/NEXT lines, and a strong corrections-owned culture (every wrong assumption is named and corrected in-place). Companion structure: PLAN (forward map), RUNBOOK (commands + results, with superseding "WHERE WE ARE" position blocks), HANDOFF (cold-start brief), SPEC (decision record) — each with a defined role. Operational lore: attachments arrived unreadable repeatedly; pasted text and file uploads are the working channels.
- **sources:** running_notes_scheme_to_components(55).md (header + throughout); PLAN_scheme_to_components(1).md#header; RUNBOOK_scheme_to_components(50).md (position blocks); HANDOFF_scheme_to_components_for_claude_code(1).md
- **relations:** house rules; docs026's own charter (this journal family is a model input for the council).
- **verify-later:** n/a — documentary practice.

### idea.uk live-VM / chassis-staging duality
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Journal project facts: "idea.uk — LIVE Go service selling £29 reports; single binary under systemd on a Hetzner VM … DNS (Cloudflare) → the VM, so chassis B2 deploys are invisible to the live site. UNCHANGED." Every checkpoint ends "idea.uk live VM untouched."
- **what:** The genuinely site-specific arrangement underpinning the whole unit's risk model: the revenue-earning idea.uk product (£29 reports, live Stripe webhook, orders in a file, reserved paths /request /confirm /approve /decline /stripe/webhook /internal/* /order/*) is a separate Go binary on a Hetzner VM; the chassis-built idea.uk site deploys to Backblaze B2 while DNS still points at the VM — so all chassis work is invisible staging and the VM cutover is a separate future decision. Two chassis site_ids exist for idea.uk (97ed2f64-… in the June thread, 1244516d-… in the July thread) — treated as separate/earlier rows, confirm before relying on either.
- **sources:** running_notes_scheme_to_components(55).md#Project-facts; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; HANDOFF_idea_uk_differentiators_section_data.md#Key-facts
- **relations:** platform mission; chassis deploy model.
- **verify-later:** sites rows for idea.uk (both ids); DNS state if a cutover is ever planned.

### Platform mission restatement (plan and build websites from a domain)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Checkpoint (Um), restated by the user 2026-07-07: "intelligently PLAN and BUILD multipage websites from a domain name — targeted design/content per vertical, eventually parsing best-in-class exemplars (reasoning from why they work, not copying), adding tools/blog/news/infographics from wider-world reasoning."
- **what:** The system's purpose as the user states it mid-thread when re-anchoring scope: an agent system that plans and builds whole multipage sites from a domain name, with vertical-targeted design and content, future exemplar-parsing (reasoning about why best-in-class sites work rather than copying), and enrichment via tools, blog, news and infographics; supported by close agent/message logging, an agent-creation guidelines doc, and distinct low-overlap agent responsibilities with sub-agents for research-before-content.
- **sources:** running_notes_scheme_to_components(55).md#Um
- **relations:** orchestrator conventions; research-agents; adoption-pipeline (exemplar parsing kinship).
- **verify-later:** agent-creation guidelines doc; exemplar-researcher plans elsewhere in docs.

### layouts.updated_at trigger and the reuse-before-create gate
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Notes (Ss): "CREATE FUNCTION errored = gate firing (shared `set_updated_at` already exists, used by site_specs/site_plans/content_feed_items/training_runs); CREATE TRIGGER bound to the EXISTING function; bump proved it. Complete — the reuse path"; (Su) observed it firing on W1b.
- **what:** Small but doctrine-carrying: layouts.updated_at gained a BEFORE UPDATE trigger, written with a deliberate collision gate — plain CREATE FUNCTION (not OR REPLACE) so a name collision errors rather than silently overwriting different semantics, which fired and routed the change onto the shared existing `set_updated_at` function. Notes the codebase convention of explicit `updated_at = now()` in UPDATEs, with which the trigger coexists harmlessly. The same "columns that never move" hygiene family was later observed on site_work_items.updated_at (listed, not actioned).
- **sources:** w2b_00_trigger_check.sql; w2b_01_layouts_updated_at_trigger.sql; running_notes_scheme_to_components(55).md#Sr #Ss #Su
- **relations:** work-item claim/retry hygiene; SQL change-management.
- **verify-later:** pg_trigger rows for layouts; set_updated_at consumers.

## Proposed NEW categories

- **NEW:page-build-pipeline** — the plan_sections → page-content-writer → compile → deploy build path and its semantics: field resolution/deferral (on_missing/skip_field, needs_section_data escalation), LLM routing (needs_llm), array item-key contracts and the render-time reconciler, build-vs-rerender distinction and fossilisation, rerender-pages workflow, the no-component-level-regen limitation, the de-tool hazard. Nine concepts in this unit alone land here; no existing spine slug owns the build path itself (styling-render-pipeline owns CSS/render, site-plan-and-reconciler owns the plan domain).
- **NEW:sql-change-management** — the needle-gate surgery pattern, idempotent sentinel-guarded prompt migrations with paired down-migrations, backup/rollback discipline, run-as-files convention. A coherent expert competence distinct from debugging (which owns the pitfall catalogue) — the council agent for "how production data is changed safely".

## Cross-unit notes for consolidation

- The 016b debugging-guide lessons extracted here (SQL pitfalls, fossil tells, status vocabulary, needle-gate rules) will also surface from the debugging docs unit — merge there, keep this unit as provenance.
- Doc 025 (Phase 4.5), 003 (contracts), 026 (regeneration), 029/030 (site plans) concepts referenced here are anchored in their own units; this unit contributes status evidence (e.g. 003 item 6 rewritten, Phase 4.5 deferred, 026's rerender claim contradicted by direct evidence).
- The two idea.uk site_ids (97ed2f64 June vs 1244516d July) should be reconciled in stage 2.
