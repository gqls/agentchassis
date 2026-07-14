| CTS-001 | Page-build-handler pipeline (plan_sections Layer 0 + validate_content) | deployed | ensure_site_record→plan_sections triage→writer→validate_content→save/deploy pipeline | contracts-and-standards.md |
| CTS-002 | Component input-schema source vocabulary (Tier A-D + renderer, proposed E) | partial | llm/static/site-data/query source tiers for component fields; feed.* Tier E undecided | contracts-and-standards.md |
| CTS-003 | content_data is the source of truth; HTML patching rejected | deployed | Sections store content_data (truth) + derived rendered_html; patching HTML is a bridge at best | contracts-and-standards.md |
| CTS-004 | Workflow result contract + dead-key stub bug class | deployed | result_from/output_fields contract; historical bug shipped entire collected_data, now fixed | contracts-and-standards.md |
| CTS-005 | Component naming contract (function = canonical kebab ID) | deployed | content_components.function is kebab, unique-per-active-row, matches data-component attr | contracts-and-standards.md |
| CTS-006 | String-value naming convention (snake identifiers, kebab data) | deployed | Go-identifier values snake_case, data-shaped values kebab-case; migration 051 applied | contracts-and-standards.md |
| CTS-007 | page_type vocabulary and "landing, not index" | deployed | Canonical kebab page_types; homepage TYPE=landing, NAME=index | contracts-and-standards.md |
| CTS-008 | JS content separation contract (js_content → assets) | deployed | Component JS split from html_template into js_content asset file; js_snippets for shared utils | contracts-and-standards.md |
| CTS-009 | Component creation & regeneration contract | deployed | StoreGeneratedComponentAction create/regenerate branches; regen keyed by LLM-emitted function | contracts-and-standards.md |
| CTS-010 | Site component linkage contract (slot_name↔function) | deployed | Missing component_id link falls to generic lookup then hardcoded fallback header | contracts-and-standards.md |
| CTS-011 | CSS colour inheritance model (--section-*, --color-* fallback) | deployed | "Single most important rule" — element colours resolve via two-level var() fallback chain | contracts-and-standards.md |
| CTS-012 | Section painting contract (four painting models) | deployed | Sections re-export --section-* as token references; literal colours forbidden | contracts-and-standards.md |
| CTS-013 | CSS theme template contract (renderer vs template ownership) | deployed | Renderer owns palette/luminance defaults; theme template owns layout/typography only | contracts-and-standards.md |
| CTS-014 | Query parameterisation contract ($1 + params) | partial | All new SQL must use $1 placeholders; tool-suggester/tool-improver still unmigrated | contracts-and-standards.md |
| CTS-015 | Schema enforcement: flexible vs strict mode | abandoned | Approval-locks-schema design; later shown the strict-mode trigger was stillborn | contracts-and-standards.md |
| CTS-016 | Handler dispatch input-path contract (input_data.spec.*) | deployed | Handlers must read spec via input_data.spec, not top-level flattening; rediscovered twice | contracts-and-standards.md |
| CTS-017 | Legal rules schema and content_direction | aspirational | Per-site legal_rules + page-level content_direction; legal-content-agent still planned | contracts-and-standards.md |
| CTS-018 | system.internal site convention | deployed | Never-deployed sites row hosting maintenance/library work items | contracts-and-standards.md |
| CTS-019 | {function}-section class contract + data-component naming | partial | Class convention honoured unevenly; data-component attribute is the reliable escape hatch | contracts-and-standards.md |
| CTS-020 | Paired-variable ("on-colour") standard | deployed | Every paintable band colour has a matching curated text colour, overridable per site | contracts-and-standards.md |
| CTS-021 | is_dark_section demoted to catalogue metadata | deployed | Unreliable self-declared flag; styling must derive from what CSS actually paints | contracts-and-standards.md |
| CTS-022 | component-creator prompt re-aim (painting rules, vocabulary) | deployed | Prompt rewritten from literal dark-section block to four painting models | contracts-and-standards.md |
| CTS-023 | Image fields optional-with-gate contract | deployed | site_assets.* fields must be required:false + skip_field + template-gated | contracts-and-standards.md |
| CTS-024 | Component schema/template/prompt three-way consistency invariant | deployed | Schema item fields, template tokens, and prompt output must agree; info-card-grid still violates | contracts-and-standards.md |
| CTS-025 | Numbered-flat-fields anti-pattern (25 components) | partial | postN_title schemas force LLM fabrication; only game-list migrated to Tier D so far | contracts-and-standards.md |
| CTS-026 | Anti-fabrication content path (llm_field_specs, merge_with) | deployed | plan_sections resolves query data pre-LLM; RenderComponentAction overlays it as authoritative | contracts-and-standards.md |
| CTS-027 | plan_sections required-field deferral trap | deployed | Required+unresolved field hits switch default, silently drops the whole section | contracts-and-standards.md |
| CTS-028 | Chrome templates must be variable-driven | aspirational | Header/footer LLM-hardcode links; pre-store hardcoded-link gate designed, not built | contracts-and-standards.md |
| CTS-029 | Thin-slice constitution (always-on rules) | deployed | Flat-file always-on rules doc; destined to become `standards` rows scope=constitution | contracts-and-standards.md |
| CTS-030 | CSS section-colour model evolution (inheritance→hardcoded→painting) | superseded | Five-source archive history of how dark-section styling was hardened over ~a year | contracts-and-standards.md |
| CTS-031 | Component Quality Contract (scoring formula) | abandoned | Full quality-scoring contract vanished from docs v6→v7; residual fields still in live JSON | contracts-and-standards.md |
| CTS-032 | query.{name} field-source resolution timing | superseded | query.* resolution moved from render-time to plan_sections-time | contracts-and-standards.md |
| CTS-033 | Adapter response envelope contract (single-sourced) | deployed | Typed-struct Kafka envelope resolved empirically; now single-sourced in 035_adapter_guide | contracts-and-standards.md |
| CTS-034 | Chassis conventions verified (text+CHECK, deleted_at) | deployed | Live-schema verification pass corrected contract docs to match reality | contracts-and-standards.md |
| CTS-035 | Priority profile (order not weights; sealed constraints) | aspirational | Exploratory mediator-framework design; adjacent-project material, not core platform | contracts-and-standards.md |
| CTS-036 | Atomic standard (generated-views doc tree) | aspirational | Rule-atoms as smallest unit; docs are generated views. Same exploratory track as CTS-035 | contracts-and-standards.md |
| CTS-037 | Input/output contracts on agent definitions | deployed | agent_definitions.input_contract/output_contract now enforced at call-site, not just docs | contracts-and-standards.md |
| CTS-038 | Call metadata vs response-data convention (output_field.response) | deployed | Call metadata at output_field; called agent's payload at output_field.response | contracts-and-standards.md |
| CTS-039 | Component render modes (template/agent/composite/standalone) | partial | render_mode field designed with 4 modes; live data shows only template/standalone used | contracts-and-standards.md |
| CTS-040 | Tier D items-array component schema shape | deployed | Single items array + sub-schema replaces numbered-flat anti-pattern; pre-store validator | contracts-and-standards.md |
| CTS-041 | Query-resolver list components (pages_where_type) | deployed | List components resolve items dynamically by page_type; no template change on page add | contracts-and-standards.md |
| CTS-042 | data-function contract + P1/P2/P3 fallback | superseded | Original structure/content decoupling; superseded by kebab function naming contract | contracts-and-standards.md |
| CTS-043 | Recursive component tree ("everything is a component") | abandoned | Slot-placeholder recursive RenderNode design; shipped system uses flat sections instead | contracts-and-standards.md |
| CTS-044 | Generation-time guards for dynamic components | deployed | Runtime-fill marker + no-inline-script baked in at generation, not patched post-hoc | contracts-and-standards.md |
| CTS-045 | CSS variable naming convention (--color-*) + STRICT RULE | deployed | LLM was emitting nonexistent --primary-color names; prompt now enforces real variable names | contracts-and-standards.md |
| CTS-046 | API documentation convention (OpenAPI + internal API.md) | deployed | Two-tier docs: external OpenAPI spec + internal per-service API.md with CI lint | contracts-and-standards.md |
| CTS-047 | Training-data export format (ChatML + metadata sidecar) | deployed | ChatML messages + ignored metadata sidecar; prose-not-JSON rows kept as DPO rejects | contracts-and-standards.md |
| CTS-048 | Local-step input resolution: input_mapping dead, key_path for loops | deployed | Coordinator doesn't resolve input_mapping for local action/loop substeps; use key_path | contracts-and-standards.md |
| CTS-049 | Capability gate D5 — requires-backend semantic tag | partial | Backend-requiring components gated by class tag; supersedes invented intent-probe site type | contracts-and-standards.md |
| CTS-050 | Class-level rename (probe → site-engine) and env-var churn | superseded | Service/paths/env-vars renamed from probe-specific to class-generic site-engine naming | contracts-and-standards.md |
| CTS-051 | /stats endpoint + INTERNAL_API_KEY | deployed | Key-gated per-host stats summary via X-Internal-Key header | contracts-and-standards.md |
| CTS-052 | /events export endpoint (P4 collector interface) | deployed | Key-gated NDJSON event stream with since/host/limit params, lock-free by design | contracts-and-standards.md |
| CTS-053 | Wrapper-orchestrator requirement finding (001:405-462) | partial | Scheduler-reached + substantive-work agents must not run in shared chassis pod | contracts-and-standards.md |
| CTS-054 | Adapter Response Envelope Contract — conditional traffic-probe application | superseded | Applicability decision demoted once P4 redesigned to need no adapter at all | contracts-and-standards.md |
| CTS-055 | Section resolvers override content_data on every render | deployed | Hero image/static fields re-resolve on every render, ignoring stored instance edits | contracts-and-standards.md |
| CTS-056 | Static-source schema fields force fleet-generic labels/suffixes | partial | source:static label fields re-apply generic fallback text on every render; fix is static→llm | contracts-and-standards.md |
| CTS-057 | Component creation contract (generator's embedded rulebook) | deployed | Full LLM-facing component-generation rulebook compiled from docs 003+018+schema v2 | contracts-and-standards.md |
| LOCK-001 | Pattern A (locked_at/locked_by) canonical; Pattern B (pinned) dead | deployed | locked_at/locked_by is the uniform lock mechanism; pinned boolean never wired | locks.md |
| LOCK-002 | Lock semantics: hard gate discovery, soft gate execution, read-only rerender | deployed | Lock means human-controls, not read-only; discovery skips locked, execution doesn't | locks.md |
| LOCK-003 | Site-level lock (sites.locked_at) | deployed | Master switch stopping all automated agent activity on a site | locks.md |
| LOCK-004 | Timed lock-expiry project and lock-model coherence plan | partial | lock_type/lock_expires_at added via migration 115 (schema only); Go predicate sweep pending | locks.md |
| LOCK-005 | Adoption faithfulness via 90-day timed locks | partial | Faithful-first-pass lock originates at first re-plan; re-plan-window enforcement undeployed | locks.md |
| LOCK-006 | auto_lock_on_deploy trigger — assumed live, later found stillborn | abandoned | Strict-mode lock trigger never functionally fired; dropped via migration 009 | locks.md |
| RAGR-001 | knowledge_base: shared pgvector RAG store | deployed | Shared embedded content store (vector(768)+trigram fallback) across industries/collections | rag-retrieval.md |
