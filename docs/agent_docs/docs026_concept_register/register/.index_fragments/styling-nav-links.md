| STY-001 | Styling render pipeline reference: two assembly paths and the scheme gap | deployed | Umbrella finding: CSS render and section render are separate paths meeting only in the browser | styling-render-pipeline.md |
| STY-002 | CSS assembly pipeline (composable theme → styles.css) | deployed | analyze_design → render_css_from_spec (deterministic) → deploy_css → CDN sync | styling-render-pipeline.md |
| STY-003 | Component quality tracking (quality_score et al.) | deployed | Scored fields on content_components drive planner/auditor regeneration targeting | styling-render-pipeline.md |
| STY-004 | Pre-store component validation gates + planning deferrals + empty-section filter | deployed | Three layers stop broken components/sections reaching pages | styling-render-pipeline.md |
| STY-005 | Scheme-to-components P0: light-resolved site renders dark | deployed | idea.uk deployed dark chrome/sections despite resolving light; fixed via paired-variable standard | styling-render-pipeline.md |
| STY-006 | Three-part styles.css assembly and core/specialised palette merge rules | deployed | 8 core slots spec-wins, specialised slots theme-wins; caused mixed output on leopardess | styling-render-pipeline.md |
| STY-007 | buildSectionDefaults: luminance-keyed dark-only --section-* defaults | deployed | Renderer's only live per-section adaptation; emits nothing on light palettes | styling-render-pipeline.md |
| STY-008 | SectionStyles: built-but-disconnected per-section CSS mechanism, retired | abandoned | ~80% built renderer mechanism no active layout ever consumes | styling-render-pipeline.md |
| STY-009 | Hero ink model and the structural-dark exception | deployed | Per-branch --hero-ink variable drives hero contrast; imageless heroes are the common case | styling-render-pipeline.md |
| STY-010 | Hazard-class vs band-class declarer taxonomy (library blast radius) | partial | Diagnostic split sizing every scheme-fix decision; non-idea.uk tail still open | styling-render-pipeline.md |
| STY-011 | Chrome selection path and the dead header_component_id column | deployed | Chrome resolution chain always falls through a NULL-forever column to fallbacks | styling-render-pipeline.md |
| STY-012 | Scheme-aware fallback chrome (RenderFallbackHeader/Footer) | deployed | Safety-net chrome functions rewritten from hardcoded-dark to var()-driven | styling-render-pipeline.md |
| STY-013 | Dual chrome render paths (build-fresh vs stale rerender-injected) | deployed | Repoint-before-force_rerender ordering prevents stale dark chrome re-render | styling-render-pipeline.md |
| STY-014 | Phase 4.5 data-section-bg surface generalisation (deferred) | aspirational | Designed attribute-based surface decouple, deferred as unneeded dark-site concern | styling-render-pipeline.md |
| STY-015 | Explicit RenderContext.Scheme signal (Q1) — abandoned design | abandoned | Explicit scheme plumbing dropped once implicit palette signal was found sufficient | styling-render-pipeline.md |
| STY-016 | Exact-field-name template binding with silent empty on miss (RenderTemplate) | deployed | `<no value>` strip is why renamed/missing fields fail silently | styling-render-pipeline.md |
| STY-017 | Section readiness model (planSection source tiers, spec resolver) | deployed | Field source tiers + required/fallback semantics decide defer-vs-carry | styling-render-pipeline.md |
| STY-018 | Stored⊕resolved merge writes resolved values back into content_data | deployed | Durable recoveries, but also a contamination carrier for bad fallbacks | styling-render-pipeline.md |
| STY-019 | Visible-content filter (≤10 chars) + data-runtime-fill assembler exemption | deployed | Base filter plus marker exempting intentionally-empty interactive shells | styling-render-pipeline.md |
| STY-020 | Assembly membership and chrome model (page_components by position) | deployed | pages.sections is metadata only; three coexisting head shapes confuse forensics | styling-render-pipeline.md |
| STY-021 | R6f — theming vocabulary drift (defined vs consumed CSS custom properties) | deployed | 11 gap names between template var() usage and generated styles.css :root | styling-render-pipeline.md |
| STY-022 | D2a — buildTokenAliases renderer-enforced compatibility bridge | deployed | Step-11 post-pass appends missing canonical/alias :root definitions | styling-render-pipeline.md |
| STY-023 | D2b — canonical-token prevention (contract rule 11 + lint) | partial | Warn-only lint + contract rule stop new orphan tokens at the source | styling-render-pipeline.md |
| STY-024 | Ambient pass-through pattern for surface painters | deployed | Sanctioned --section-x: var(--color-x) pass-through for fallback-less consumers | styling-render-pipeline.md |
| STY-025 | Interactive-section clobber + interactivity-aware save guard | partial | Full rebuilds silently discard interactive tools stored only as rendered_html | styling-render-pipeline.md |
| STY-026 | Theme-layer render resolution (style_collection → css_theme) | deployed | Render path resolves colour exclusively via style_collection, ignoring site_specs | styling-render-pipeline.md |
| STY-027 | Render-off-build_status debt (planned-vs-rendered diff) | partial | Rebuilds skip planned-but-missing sections on already-deployed pages | styling-render-pipeline.md |
| STY-028 | site-asset-renderer: deterministic per-site JS snippet bundling | deployed | Bundles js_snippets into assets/js/snippets.js per site, closing the loader gap | styling-render-pipeline.md |
| STY-029 | CSS component-list fallback bug (fake 5-item list) | deployed | Bad status filter emptied component list, triggering a hardcoded fallback; fixed | styling-render-pipeline.md |
| STY-030 | CSS applies_to granularity mismatch (known issue, unfixed) | partial | Exact-text overlap matching means only 2 of ~21 snippets ever ship | styling-render-pipeline.md |
| STY-031 | Rerender pipeline (rerender-pages/page-rerender/render-site-components) | deployed | Assembly/deployment half of the system, separated from content generation | styling-render-pipeline.md |
| STY-032 | CSS responsibility barrier + colour inheritance model | deployed | Global CSS owns colour/typography; component CSS owns layout only | styling-render-pipeline.md |
| STY-033 | Section-contrast model (is_dark_section + --section-* variables) | deployed | Dark-background components must define the --section-* contract on container | styling-render-pipeline.md |
| STY-034 | JS delivery paths & the js_snippets loader gap (historical) | partial | Four coexisting JS paths catalogued; loader gap later closed by STY-028 | styling-render-pipeline.md |
| STY-035 | Inline JS extraction contract (separateInlineJS / js_content) | deployed | Bare script blocks extracted to external per-component JS files at store time | styling-render-pipeline.md |
| STY-036 | aggregate_webpage HTML assembly action (first-gen renderer) | superseded | Earliest page renderer, one action call per page, long since replaced | styling-render-pipeline.md |
| STY-037 | Content/structure separation: JSON content + html-assembler | superseded | Defined the modern content/template split; taxonomy ancestor of current pipeline | styling-render-pipeline.md |
| STY-038 | HTML action architecture (generate → process → validate) | superseded | Three-action LLM page pipeline, replaced by component-template rendering | styling-render-pipeline.md |
| STY-039 | Batched multipage generation (assemble_multipage_site) | superseded | Batch-of-3-5 generation strategy, replaced by sequential loop-based generation | styling-render-pipeline.md |
| STY-040 | Asset bubble-up deduplication (proposal, never shipped) | abandoned | Recursive per-component asset merge proposal; barrier model shipped instead | styling-render-pipeline.md |
| STY-041 | Assembly action consolidation (3 clear actions) | partial | Rationalising 6 overlapping assembly actions down to assemble_page and siblings | styling-render-pipeline.md |
| STY-042 | Component library unification (component_library.go) | deployed | Shared Go module: one source of truth for component ops and chrome rendering | styling-render-pipeline.md |
| STY-043 | page_components: component instances as the page's stored form | deployed | Single most consequential schema decision enabling rerender/edit/lock | styling-render-pipeline.md |
| STY-044 | Head-inside-body bug and positional injection fixes | deployed | Dedup-by-size heuristic kept the wrong misplaced head; fixed to dedup by position | styling-render-pipeline.md |
| STY-045 | Slot-based modular page assembly (proposal, partially adopted) | partial | Pure-concatenation assembly proposal; site_components shipped, JSON-first landed differently | styling-render-pipeline.md |
| STY-046 | CSS generation bug (webdesign-agent design_spec not applied) | partial | Deployed CSS reverts to default blue template despite a correct design_spec | styling-render-pipeline.md |
| STY-047 | http2 deprecation fix at the nginx conf generator | deployed | setup.sh now emits version-neutral listen directives | styling-render-pipeline.md |
| STY-048 | page-rerender mode contract and site-uniformity reconcile pattern | deployed | Two page-rerender modes with different skip semantics; idempotent reconcile scripts | styling-render-pipeline.md |
| NAV-001 | Nav agent family and the three-tier authority model | partial | Only Tier 1 (strategist, new-build) is fully implemented of the three tiers | navigation.md |
| NAV-002 | Two nav systems and the GetNavItems fallback | deployed | site_nav tables vs legacy pages.in_header flags; partial population mixes both | navigation.md |
| NAV-003 | Stale pages polluting nav + config-driven deactivation fix | deployed | build_status filter + deactivate_stale_pages flag close the stale-nav gap | navigation.md |
| NAV-004 | Nav discovery checks and fix agents | deployed | Discovery/fixer pairs for anchor-slug links, stacked nav, unlinked components, orphans | navigation.md |
| NAV-005 | Duplicate header/footer pathology | partial | Site-level components leaking into pages.sections cause double-rendered chrome | navigation.md |
| NAV-006 | Nav quality mechanisms of 2026-04-17 | deployed | Tiered priority, child-page exclusion, label trust, footer quick links shipped together | navigation.md |
| NAV-007 | Hardcoded fallback nav/header defaults inventing structure | partial | Brochure-default fallbacks fabricate URLs; later found not the live-path cause for one incident | navigation.md |
| NAV-008 | Tool nav integration | partial | Per-tool nav entries work but grouping/label-length design remains open | navigation.md |
| NAV-009 | Navigation maintenance: nav-updater and nav-link-fixer | deployed | Algorithmic, no-LLM agents refreshing nav tables and fixing anchor-slug templates | navigation.md |
| NAV-010 | Navigation tables (site_nav_groups / site_nav_items) | deployed | First-class nav model replacing scattered pages-table queries and the old cache | navigation.md |
| NAV-011 | Global context injection for navigation (superseded) | superseded | Earlier Global.Sitemap design, superseded by nav tables + GetNavItems | navigation.md |
| NAV-012 | Header nav from pages.in_header + nav-label hygiene | deployed | Nav membership is a data flag; label hygiene is a companion defect | navigation.md |
| LNK-001 | link_registry as first-class link index + planned links-orchestrator family | partial | Registry/extraction/validation shipped; the agent family on top remains unbuilt | link-management.md |
| LNK-002 | Internal linking machinery and its defects | partial | pages table as link-target authority, catalogued against known defects | link-management.md |
| LNK-003 | Hero CTA brochure-default defect (text↔destination mismatch) | deployed | Generic hero schemas defaulted every CTA to /contact.html and phantom /services.html | link-management.md |
| LNK-004 | sourceResolver `pages` fabrication bug — the phantom generator | deployed | Root cause of hero phantoms; also corrected an earlier wrong-cause diagnosis | link-management.md |
| LNK-005 | Correct-or-absent principle + loud-but-non-blocking phantom policy | deployed | Structural rule: never fabricate a link target; absence is a warning, not an error | link-management.md |
| LNK-006 | Step 1 / Layer 1a hero+CTA schema/template hardening | deployed | skip_field + gated buttons so an unresolved CTA renders nothing | link-management.md |
| LNK-007 | Layer 1b header/footer phantom fix (shared site components) | deployed | Hardcoded ContentData in render_site_components fixed at the Go source | link-management.md |
| LNK-008 | datahelpers/links.go — canonical link classification library | deployed | Single shared normaliser used by both deploy gate and post-deploy audit | link-management.md |
| LNK-009 | check_phantom_internal_links post-deploy audit + surface routing | partial | Discovery check built and gate-cleared but deliberately not yet enabled | link-management.md |
| LNK-010 | B4/B5 Browse-All hub links via `section_index_for` queryresolve verb | deployed | New queryresolve verb replaces empty *_index_url specs for hub buttons | link-management.md |
| LNK-011 | internal-link-resolver agent | deployed | Build-time sub-agent resolving hero/CTA destinations to real pages, not a patcher | link-management.md |
| LNK-012 | unresolved_cta build-time HITL signal | deployed | Only place a correctly-dropped CTA button's absence is detectable pre-deploy | link-management.md |
| LNK-013 | page-content-writer ↔ resolver wiring with regression-safe fallback chain | deployed | Fallback-to-prior-behaviour wiring later masked a bug for two weeks | link-management.md |
| LNK-014 | select_sections path-mismatch bug (phantom CTA root cause) | deployed | Wrong JSON path silently discarded resolver output; one-line jsonb_set fix | link-management.md |
| LNK-015 | link_registry — records but never validates (dormant substrate, abandoned) | abandoned | target_page_id never populated; live audit reads rendered_html directly instead | link-management.md |
| LNK-016 | nav-link-fixer agent (template-anchor scope only) | deployed | Fixes #{{.slug}} anchors in templates; cannot reach hardcoded ContentData | link-management.md |
| LNK-017 | prepare_link_context available_pages gap on the work-item path | partial | Work-item rebuild path leaves the LLM's link-context constraint empty | link-management.md |
| LNK-018 | Semantic linking domain decomposition (5 link types) | partial | Taxonomy splitting link work by lifecycle/complexity; most agents unbuilt | link-management.md |
| LNK-019 | Links agent family (algorithmic, no-LLM link health) | partial | Judgment-free crawler/validator/redirect-manager family, still unimplemented | link-management.md |
| LNK-020 | site_specs `cta` aspect + CTA graph audit (parked) | partial | Shared CTA URL source fixed dependants; graph found circular, retarget parked | link-management.md |
| LNK-021 | link registry, cached navigation structures, and redirects (foundation) | deployed | Original MVP schema: link_registry + versioned nav cache + redirects table | link-management.md |
| LNK-022 | CTA-graph integrity (dead-end and circular primary actions) | partial | Every primary CTA once 404'd, then became circular; retarget deliberately parked | link-management.md |
