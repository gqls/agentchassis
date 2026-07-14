| SYS-001 | Kafka topic model evolution | deployed | Three eras of topic naming; current model has generic entry, per-spawn job topics, fixed adapter topics | system-architecture.md |
| SYS-002 | Agent message contract & "agent = row" orchestrator convention | deployed | Agent = DB row with default_config.workflow; spawn-before-call; reply to caller's topic; house rules | system-architecture.md |
| SYS-003 | Orchestration state and collected_data as the workflow data bag | deployed | orchestration_states row holds workflow_plan/collected_data/current_step/status | system-architecture.md |
| SYS-004 | Stale orchestration sweeper | deployed | Periodic DB sweep classifies/repairs expired awaited_requests instead of in-memory timeouts | system-architecture.md |
| SYS-005 | Work-item relay spine / dispatch-loop pattern | deployed | Baton = site_work_items row; 30s pump seeds queue; dispatch loop claims and spawns handlers | system-architecture.md |
| SYS-006 | Entity data model | aspirational | State-based lifecycle entities driving pages, news triggers, client-side real-time data | system-architecture.md |
| SYS-007 | Maintenance profile per-site configuration | deployed | sites.settings.maintenance_profile controls per-domain cadence, budgets, audit config | system-architecture.md |
| SYS-008 | Idle timeout for spawned agents + topic cleanup strategy | deployed | idle_timeout_seconds env var exits idle pods; CronJob + Kafka retention clean up orphan topics | system-architecture.md |
| SYS-009 | business-intel shared-pod pattern | deployed | Multiple agent defs share one static pod via message routing; ai_service must live in step config | system-architecture.md |
| SYS-010 | CollectedData pathologies | deployed | Overloaded single-channel data structure with documented duplication/namespace pathologies | system-architecture.md |
| SYS-011 | Flat-namespace collision risk and compensating-mechanism accretion | deployed | Shared flat map lets actions silently pick up wrong fields; workarounds accreting | system-architecture.md |
| SYS-012 | Response-topic consumer group race | unknown | Per-pod consumer groups fan every response out to every pod, causing version races | system-architecture.md |
| SYS-013 | Kafka empty partition assignment on simultaneous pod join | unknown | Simultaneous deploy join can leave a partition unassigned; workaround is killing a pod | system-architecture.md |
| SYS-014 | Observability gaps: owner_agent_type "generic" | unknown | Rerouted generic workflows keep misleading owner_agent_type, breaking searches | system-architecture.md |
| SYS-015 | Four overlapping chrome default stores | partial | Header/footer defaults split across 4 stores; intended chain deliberately left unrepaired | system-architecture.md |
| SYS-016 | Coordinator result-extraction contract (resolveResultSpec) & silent-stub bug family | deployed | Singular output_field silently dropped results; centralised result_spec.go fix shipped | system-architecture.md |
| SYS-017 | Hosting split: static-serverless front + small always-on backend | deployed | Pure-static sites serverless on B2; multi-LLM/webhook jobs need a small always-on service | system-architecture.md |
| SYS-018 | Oversize-result delivery: fail-loud hardening + size guards | deployed | Oversize completions now error loudly instead of shipping full collected_data as a stub | system-architecture.md |
| SYS-019 | sites.status informational lifecycle label (system-architecture angle) | deployed | No build-time code filters on sites.status; build dispatch keys on site_work_items instead | system-architecture.md |
| SYS-020 | Aspiration: agent-creation & message logging workstream | aspirational | Stated desire to log agent creation/inter-agent messages as its own workstream; never built | system-architecture.md |
| SYS-021 | pages.sections is the build-read field | deployed | Page build resolves sections via site_specs.site_plan then falls back to pages.sections | system-architecture.md |
| SYS-022 | Chassis config location bugs | partial | max_tokens shadowing, dead step-level temperature, dropped error_step field | system-architecture.md |
| SYS-023 | work-site-orchestrator (monolith) vs build-site-planner (thin planner) | deployed | Old inline monolith replaced by a thin planner that delegates via work items | system-architecture.md |
| SYS-024 | Snapshot-shadowing agent-definition loader defect | deployed | Unfiltered ORDER BY version reads let snapshots silently shadow live agent rows | system-architecture.md |
| SYS-025 | Quality Assurance Agent Architecture | superseded | Three-layer QA model folded into the main system-architecture doc, not abandoned | system-architecture.md |
| SYS-026 | site_work_items domain → pipeline column rename | superseded | Work-routing column renamed to eliminate collision with the site's own domain name | system-architecture.md |
| SYS-027 | Dispatch-loop input_mapping path mismatch | unknown | spec JSONB mapped nested but handlers read flat, causing path-resolution errors | system-architecture.md |
| SYS-028 | Asset self-resolving storage URI (dispatch loop) | superseded | asset-deployer resolves its own s3:// URI from asset_id instead of orchestrator pre-resolving | system-architecture.md |
| SYS-029 | Self-spawning flat dispatch-loop (pre-scheduler, superseded) | superseded | Early one-item-then-respawn design replaced by scheduler tick + in-workflow loop | system-architecture.md |
| SYS-030 | claim_work_item atomic claim + load_work_items first_item | deployed | Atomic UPDATE...RETURNING claim prevents double-processing of work items | system-architecture.md |
| SYS-031 | collected_data growth causing OOM-kills | partial | 18MB collected_data blew the 512Mi pod limit, causing phantom-completed orchestrations | system-architecture.md |
| SYS-032 | Page content-creation build pipeline trace | deployed | Hop-by-hop trace from load_page_record through SavePageSectionsAction | system-architecture.md |
| SYS-033 | extractResponseContent flat-string hypothesis (superseded) | superseded | Writer-can't-populate-structured-fields hypothesis disproven by an isolated build test | system-architecture.md |
| SYS-034 | Site-chrome rendering gap | partial | Relay build path may never invoke the chrome-rendering step; zero <nav> measured live | system-architecture.md |
| SYS-035 | Generic orchestrate envelope as universal manual trigger | deployed | kcat-produce shape for hand-running any agent via the generic entry point | system-architecture.md |
| SYS-036 | Parent/child result key = step-name convention | deployed | Child response is stored under the calling step's own name, never a synthetic key | system-architecture.md |
| SYS-037 | Workflow default_config location convention | deployed | Agent workflow lives in default_config, never the separate *_workflow columns | system-architecture.md |
| SYS-038 | Autonomous Build-and-Operate — trust-not-capability thesis | aspirational | Umbrella vision bounding LLM uncertainty to progressively remove the human | system-architecture.md |
| SYS-039 | Build-vs-operate asymmetry | aspirational | Build work is sandboxable/competition-safe; operate work is live and risk-gated | system-architecture.md |
| SYS-040 | Lifecycle map by verifiability + containment (Tier A/B/C) | aspirational | Autonomy ceiling set by verifiability and containment, independent of agent capability | system-architecture.md |
| SYS-041 | Autonomous control loop | aspirational | route-produce-verify-gate-apply-feedback wraps the existing orchestrator unchanged | system-architecture.md |
| SYS-042 | Mediator routing model | aspirational | Change descriptor matched against doc-tree metadata routing table to select consultees | system-architecture.md |
| SYS-043 | Wrapper-orchestrator pattern (pod lifecycle) | deployed | Thin spawn→call→complete wrapper gives real work its own dedicated Job pod | system-architecture.md |
| SYS-044 | Loop mechanisms | deployed | Loop steps expand into N×M dynamic workflow steps, not Go for-loops | system-architecture.md |
| SYS-045 | Architectural tensions catalogue | partial | Living catalogue of genre-level design tensions (infer-and-repair; page identity) | system-architecture.md |
| SYS-046 | Site / area / page component hierarchy | partial | Three-level slot resolution: area_components → site_components → assembly | system-architecture.md |
| SYS-047 | Pages / page_components split | deployed | Structure/workflow in pages; actual rendered content in page_components | system-architecture.md |
| SYS-048 | awaited_requests global request/response registry | deployed | DB-backed registry matching Kafka responses to waiting orchestrations | system-architecture.md |
| SYS-049 | Message deduplication (processed_messages) | deployed | Dedup key + composite PK blocks duplicate delivery within a retry generation | system-architecture.md |
| SYS-050 | Orchestration ↔ site linkage (orchestration_states.site_id) | deployed | Direct nullable site_id column replaces JSONB spelunking for per-site orchestrations | system-architecture.md |
| SYS-051 | Sites contact-identity denormalisation | deployed | Identity/contact fields promoted from content_data JSONB to first-class columns | system-architecture.md |
| SYS-052 | Universal orchestration principle & agent_group_definitions elimination | deployed | No orchestrator/worker distinction; groups became agents with spawn/call workflows | system-architecture.md |
| SYS-053 | Stateless-first agent principle + DB-backed orchestration state | deployed | Agents are ephemeral executors; all workflow state persists in the database | system-architecture.md |
| SYS-054 | ExecutionContext unified message envelope and ID semantics | deployed | correlation/orchestration/request id semantics; sender constructs, receiver trusts | system-architecture.md |
| SYS-055 | Two-phase agent lifecycle (spawn + initialize handshake) | deployed | Spawn creates the pod; a separate initialize handshake precedes real work | system-architecture.md |
| SYS-056 | SagaCoordinator engine: embedded, distributed, no central orchestrator | deployed | Every pod embeds a full orchestrator loading JSON workflows from the DB | system-architecture.md |
| SYS-057 | Reply-to metadata (__work_request__) convention | deployed | Stores request_id + caller's responses topic at receipt time, used at completion | system-architecture.md |
| SYS-058 | Perspective transformation | deployed | Receiver's own orchestration becomes primary; sender is responsible for headers | system-architecture.md |
| SYS-059 | MessageType semantics | deployed | request = actively working now; response = reporting back, not a history marker | system-architecture.md |
| SYS-060 | Fuel budget resource management | unknown | Per-orchestration compute budget plumbed through headers, enforcement unconfirmed | system-architecture.md |
| SYS-061 | Child-orchestration timeout monitor | partial | In-memory per-child timeout goroutine; pod-restart recovery was a known gap | system-architecture.md |
| SYS-062 | Fan-out and awaited-response correlation | deployed | fan_out step dispatches parallel sub-tasks matched back via causation_id | system-architecture.md |
| SYS-063 | Early long-term platform ambitions | aspirational | Founding roadmap: self-organising teams, marketplace, multi-tenant, cross-cluster | system-architecture.md |
| SYS-064 | Environment variable validation framework (abandoned) | abandoned | Planned pre-spawn env var validation framework, silently dropped | system-architecture.md |
| SYS-065 | relationships table — first-class entity relationships | partial | Generic relationship entity modelled on links, later earmarked for semantic page links | system-architecture.md |
| SYS-066 | Agent families architecture | partial | Eight specialist-agent families each owning a data domain, mixed completion | system-architecture.md |
| SYS-067 | "Database is source of truth, Git is the deployment artifact" | deployed | Pivotal data-ownership doctrine enabling rerendering and granular editing | system-architecture.md |
| SYS-068 | Layer-1 / Layer-2 hack-resistance model | deployed | Core cluster never serves inbound traffic; Layer 2 is static-on-S3 with nothing to compromise | system-architecture.md |
| SYS-069 | Gateway proxy pattern (auth-service → core-manager) | deployed | auth-service is the only HTTP ingress; core-manager re-validates JWTs independently | system-architecture.md |
| SYS-070 | site-engine (API-only capture backend) | deployed | stdlib-only Go binary capturing intent events server-side for VM-hosted sites | system-architecture.md |
| SYS-071 | Standalone "probe-go" service (abandoned) | abandoned | Forked multi-vhost service rejected as too far from the website-building chassis | system-architecture.md |
| SYS-072 | Layer-4-build + thin-Layer-5-VM-deploy framing | deployed | Traffic probe reuses the existing build+deploy pipeline, swapping only the target | system-architecture.md |
| SYS-073 | Phased plan P0–P5 (traffic probe) | partial | Structural decisions through off-box collection and a future registry adapter | system-architecture.md |
| SYS-074 | VM-Hosted Backend Sites class (proposed) | aspirational | New persistent internet-facing VM class with DNS/TLS/data-return lifecycle | system-architecture.md |
| SYS-075 | Pull architecture / no collector VM | deployed | Serving boxes buffer JSONL; the cluster pulls over key-gated HTTPS, no push | system-architecture.md |
| SYS-076 | idea.uk topology exception | deployed | Deliberate exception to the serverless-edge default: a small always-on backend | system-architecture.md |
| SYS-077 | Agent chassis — generic configurable agent executor | deployed | One reusable Go binary becomes any agent type via database configuration | system-architecture.md |
| SYS-078 | Local vs remote actions and the action registry | deployed | Workflow steps run synchronously in-process or dispatch to another agent's topic | system-architecture.md |
| SYS-079 | Message header contract | deployed | Rich sender identity, retry, and status-enum header set on every message | system-architecture.md |
| SYS-080 | Orchestration-as-identity model (AgentID = PodName) | deployed | The orchestration record, not the pod, is the persistent "agent doing a task" | system-architecture.md |
| SYS-081 | Optimistic locking on orchestration state | unknown | Version-column CAS design specified but unconfirmed as shipped | system-architecture.md |
| SYS-082 | Retry semantics | unknown | Same request_id with incremented retry_version, unconfirmed as shipped | system-architecture.md |
| SYS-083 | Agent-centric architecture: steps call agents, not topics | deployed | call_agent with agent_type is the primary abstraction over raw topic addressing | system-architecture.md |
| SYS-084 | Inter-agent invocation protocol v1 (superseded) | superseded | Early invoke_agent/agent_invocations design replaced by call_agent + headers | system-architecture.md |
| SYS-085 | Project Manager / User Representative agent hierarchy (abandoned) | abandoned | Top-level PM/user-rep persona hierarchy vanished; review intent moved to HITL | system-architecture.md |
| SYS-086 | HTML-first progressive enhancement delivery | deployed | Deliberate plain HTML/CSS/JS generation strategy that survived into the renderer | system-architecture.md |
| SYS-087 | Workflow status state machine | deployed | RUNNING/AWAITING_RESPONSES/COMPLETED/FAILED vocabulary, minor drift across eras | system-architecture.md |
| SYS-088 | Human-readable orchestration and correlation names | deployed | Generated readable names alongside UUIDs for narrative-style debugging | system-architecture.md |
| SYS-089 | Agent teams: composite/family/service-agent patterns (abandoned) | abandoned | Complex team-composition designs abandoned in favour of simpler agent groups | system-architecture.md |
| DBI-001 | Snapshot-before-mutate discipline | deployed | snapshot_agent/take_site_snapshot + backup naming convention before any mutation | database-and-infrastructure.md |
| DBI-002 | Migrations ledger system | deployed | schema_migrations table + guarded run-migrations.sh runner, live since 2026-07-10 | database-and-infrastructure.md |
| DBI-003 | API keys logged in plaintext — exposure & rotation | deployed | STABILITY/BANANA keys exposed in logs for 7 weeks, then scrubbed and rotated | database-and-infrastructure.md |
| DBI-004 | Three-database architecture | deployed | MySQL auth + PostgreSQL clients_db + PostgreSQL templates_db via pgbouncer | database-and-infrastructure.md |
| DBI-005 | Schema-per-client multi-tenancy (create_client_schema) | deployed | Each client gets an isolated client_<id> schema created by a SQL function | database-and-infrastructure.md |
| DBI-006 | agent_instances/agent_spawn_history column-shape drift & correction | superseded | Archived manual-DDL column list didn't match spawn_actions.go; corrected twice | database-and-infrastructure.md |
| DBI-007 | Clients → networks → sites hierarchy (early spine) | superseded | Original multi-tenancy spine; networks/clients now rarely referenced vs sites | database-and-infrastructure.md |
| DBI-008 | sites.build_status vestigial column | unknown | Defaults to pending and is never advanced by any code path | database-and-infrastructure.md |
| DBI-009 | layouts.updated_at trigger and reuse-before-create gate | deployed | Plain CREATE FUNCTION collision gate routed the trigger onto a shared function | database-and-infrastructure.md |
| DBI-010 | system.internal pseudo-site anchor pattern | deployed | Platform-wide work items anchor to a fixed pseudo-site instead of a null site_id | database-and-infrastructure.md |
| DBI-011 | Ownership hierarchy reuse for entitlement scoping | deployed | Existing clients→networks→sites FK chain reused for billing/entitlement instead of new ownership | database-and-infrastructure.md |
| DBI-012 | Model-written SQL under a three-guard read-only substrate | deployed | Prompt guard + parse-lint + read-only transaction let a model emit arbitrary SELECT SQL safely | database-and-infrastructure.md |
| DBI-013 | QueryDatabaseAction parameterised queries & schema-drift discipline | deployed | $1 placeholders required; live DB, not dumps, is the source of truth for columns | database-and-infrastructure.md |
| DBI-014 | Database cleanup and log retention policy | deployed | Uniform per-table retention functions with distinct success/error windows | database-and-infrastructure.md |
| DBI-015 | Early schema inventory and since-dropped tables | superseded | Point-in-time \dt+ snapshot preserving tables since dropped or absorbed elsewhere | database-and-infrastructure.md |
| DBI-016 | Auth database provisioning | deployed | Hand-provisioned auth_db/auth_user; file preserves a live credential (hygiene finding) | database-and-infrastructure.md |
| DBI-017 | Database password rotation runbook | deployed | Three-holder password chain rotated in a safe PG→secret→PgBouncer order | database-and-infrastructure.md |
| DBI-018 | sites.status vocabulary and the blast-radius filter trap | deployed | Validated status vocabulary; heuristic against scoping blast-radius queries on 'active' | database-and-infrastructure.md |
| DBI-019 | training_exports Postgres schema | deployed | Versioned ChatML dataset snapshots in Postgres TOAST instead of S3 | database-and-infrastructure.md |
| DBI-020 | clients_db vs templates_db agent_definitions source-of-truth | deployed | Rich-schema agent_definitions load from clients_db, not templates_db | database-and-infrastructure.md |
| DBI-021 | setup.sh box provisioning script | deployed | Multi-vhost box installer: nginx, certbot, systemd, hardening, deploy hook | database-and-infrastructure.md |
| DBI-022 | Deploy privilege model (site-engine-deploy sudo hook) | deployed | Scoped sudoers rule lets a deploy user swap the binary with no root CI key | database-and-infrastructure.md |
| DBI-023 | VM sizing / Hetzner box selection | deployed | Boxes sized by disk/log headroom; EU-only Hetzner, x86-only build caveat | database-and-infrastructure.md |
| DBI-024 | No tenant isolation today; dedicated-cluster-per-client as offered capability | partial | Single shared Postgres/Kafka with no RLS; per-client cluster positioned as buildable | database-and-infrastructure.md |
