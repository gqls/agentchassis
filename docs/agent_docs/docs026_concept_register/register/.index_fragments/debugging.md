| DBG-001 | The 016/016b debugging guide: assumption checklist + durable invariants | deployed | Canonical, versioned operational runbook + heuristics; extracted piecemeal by nearly every unit | debugging.md |
| DBG-002 | Agent is a DB row; trust default_config over prose | deployed | default_config.workflow is the real behaviour; description can lie; two possible source DBs | debugging.md |
| DBG-003 | LLM step config field-path shadowing (ai_service/max_tokens/temperature) | partial | Top-level ai_service shadows step overrides; misplaced max_tokens silently defaults to ~2048 | debugging.md |
| DBG-004 | Timeout chain ordering contract (claim > call_handler > workflow) | deployed | Three nested timeouts must strictly decrease or duplicate/orphaned handling results | debugging.md |
| DBG-005 | Claimed-item-timeout evidence-based auto-completion (false-positive family) | partial | v1→v2 fix history for false-completing stuck claims; homepage zero-component incident | debugging.md |
| DBG-006 | jsonb && operator class bug | deployed | No jsonb&&jsonb operator; css_snippets errored silently for months | debugging.md |
| DBG-007 | Silent-completion / "trust the artefact, not the status" family | deployed | The platform's most recurring failure shape: success status masking no-op work | debugging.md |
| DBG-008 | save_page_sections: sole writer, HTML-fallback bug, guard laundered to success | partial | content-regression guard's refusal routed through a SUCCESS-labelled complete_error step | debugging.md |
| DBG-009 | Presign/awaited-loop O(K²) state bloat, fixed by batch adapter calls | deployed | Per-item awaited loops re-persist full state; batch adapter call is the structural fix | debugging.md |
| DBG-010 | Hand-applied agent-def migrations have no ledger; re-run reverts later ones | deployed | No schema_migrations ledger; idempotent only vs own prior application | debugging.md |
| DBG-011 | CrashLoop exec "./X" — image lacks the binary | deployed | Wrong/stale Docker image content; "no guard between built and running" recurred 3x | debugging.md |
| DBG-012 | Open problem: nav-updater never spawns | unknown | Definition/topics exist but no pod ever starts; nav_drift items always claim-timeout | debugging.md |
| DBG-013 | Zero-planned-sections silent no-op success (planning gap) | partial | Page with 0 sections routes to complete_error, a success-labelled step | debugging.md |
| DBG-014 | Content↔template key-contract drift (system-stats class) | partial | Rewritten template shares zero keys with stored content_data; visible-content filter drops it | debugging.md |
| DBG-015 | agent_error_log/llm_call_log/http_request_log as primary forensic sources | deployed | Persistent DB logs outlive pod stdout; read them before kubectl logs | debugging.md |
| DBG-016 | SQL/DB template-and-data surgery method (needle-gate + Postgres pitfalls) | deployed | Dominant safe method for mutating production templates/prompts/configs by SQL | debugging.md |
| DBG-017 | sites.status is informational; never scope blast-radius by status='active' | deployed | 'active' is legacy; dispatch keys on site_work_items instead | debugging.md |
| DBG-018 | Kafka trigger payload discipline: multi-line kcat bodies mis-route | deployed | kcat -P is line-delimited; multi-line JSON silently mis-pairs headers/body | debugging.md |
| DBG-019 | Discovery-checks list maintenance and the workflow-replace landmine | deployed | jsonb array append is safe; whole-workflow jsonb_set is a future silent-erase risk | debugging.md |
| DBG-020 | Deployed-binary-predates-disk failure class | deployed | Observed behaviour contradicts correct code because the deployed image predates the repo | debugging.md |
| DBG-021 | LLM API shape disciplines (server-tool injection, per-model shapes, timeouts) | deployed | Anthropic API wire-shape differences across model generations; long-call timeout sizing | debugging.md |
| DBG-022 | Operator/assistant division-of-labour + DB-change safety conventions | deployed | Snapshot-before-change, verified replace(), workflow-vs-Go deploy distinction | debugging.md |
| DBG-023 | Send-before-register await race (preRegisterAwaitedRequest fix) | deployed | Fast adapter reply beat awaited_requests insert; fixed by register-before-send | debugging.md |
| DBG-024 | agent_definitions source-of-truth is clients_db, not templates_db | deployed | templates_db has only the legacy 8-agent old-schema catalogue | debugging.md |
| DBG-025 | CLI/ops data-transfer pitfalls (COPY/psql jsonb, kubectl exec/cp, tnr scp) | deployed | A cluster of verified transfer traps beyond the kcat heredoc issue | debugging.md |
| DBG-026 | configOrInput numeric config coercion (expiry_minutes silently dropped) | deployed | Shared helper type-asserted to string; JSON numbers silently fell through to defaults | debugging.md |
| DBG-027 | Scheduler-fired chassis-resident agents report owner_agent_type='generic' | deployed | Monitoring filters must key on collected_data config.agent_type, not owner_agent_type | debugging.md |
| DBG-028 | Kafka topic-creation race self-heal (transient "Topic not yet on broker") | deployed | First-publish race on new per-spawn topics self-heals on retry; not a real fault | debugging.md |
| DBG-029 | Loose dispatch item-status semantics (complete ≠ done) | aspirational | Seven dated sightings of dispatch bookkeeping bugs; fix parked as hygiene | debugging.md |
| DBG-030 | F2 tiered guard-verification methodology (unit→integration→live fixtures) | deployed | Tier 1/2/3 verification with KEEP/REJECT fixtures; evidence discriminator ordering | debugging.md |
| DBG-031 | R6c artifact-forensics method: cache-busted, metric-consistent comparisons | deployed | Only compare artefacts with identical metrics; md5 before concluding stale-cache | debugging.md |
| DBG-032 | Code-ahead-of-DB schema drift (SQLSTATE 42703, latent until first caller) | deployed | Binary referenced columns before migration ran; migration file was mis-parked | debugging.md |
| DBG-033 | Prompt-template rendering resolvers differ by output_format | deployed | text→bare string, json→map, action-config→different resolver keeping .result | debugging.md |
| DBG-034 | EXECUTING_STEP frozen forever means the worker pod died (OOMKill triage) | deployed | orchestration_states written by the worker; a dead pod freezes the row, not a hang | debugging.md |
| DBG-035 | chunkContent() infinite loop — the OOM root cause | deployed | start=end-overlap stepped backwards forever; fixed with forward-progress guard | debugging.md |
| DBG-036 | Env-prefix trap: VAR=x on its own line never reaches the child process | deployed | Missing export/same-line prefix silently uses defaults; banner-tell mitigation | debugging.md |
| DBG-037 | Two failure envelopes: COMPLETED parent ≠ child succeeded | deployed | sendWorkflowResponse hides failure in body; notifyParentOfFailure is the other shape | debugging.md |
| DBG-038 | Pod label agent-type (hyphen) vs log field agent_type (underscore) | deployed | Underscore selector silently matches zero pods; type-wide selectors mix in stale runs | debugging.md |
| DBG-039 | 0-rows rule: zero rows decisive only after query AND run completion ruled in | deployed | State-dump substitute for evidence past the idle-reaper's 3600s capture window | debugging.md |
| DBG-040 | Untracked-file deploy trap: verify by ancestry, not tag/commit message | deployed | git commit -a misses untracked files; verify via git merge-base --is-ancestor | debugging.md |
| DBG-041 | Convergence inertness: []map[string]interface{} vs []interface{} assertion | deployed | Type assertion always failed; whole convergence feature silently dead since deploy | debugging.md |
| DBG-042 | Defect-catalogue discipline: enumerate by root cause, read-pin-confirm-fix | deployed | Post-deployment audit method grouping defects by root-cause family, not symptom | debugging.md |
| DBG-043 | Kafka consumer-group recovery: restart-to-rejoin, never replay-from-earliest | deployed | Topic wipe broke group membership; replay-from-earliest risked duplicate adoptions | debugging.md |
| DBG-044 | Manual work-item insertion as an operational rebuild lever | deployed | needs_page/needs_content_page hand-inserts are claimed normally by dispatch | debugging.md |
| DBG-045 | Kafka per-spawn response-topic partition race (adapter reply lost) | partial | LeastBytes balancer picks out-of-range partition on fresh per-spawn topics | debugging.md |
| DBG-046 | Work-item re-drive and zombie-claim operational semantics | partial | Stuck claims block a whole site; re-drive requires resetting attempt_count + claim metadata | debugging.md |
| DBG-047 | Pipeline field as a soft routing label (needs_imagery excluded by pipeline='design') | deployed | Discovery checks inherited wrong pipeline label; dispatcher filter also loosened | debugging.md |
| DBG-048 | Early pipeline-failure triage priorities dropped by root-cause diagnosis | abandoned | First-pass symptom triage superseded within a day by deeper root-cause fixes | debugging.md |
| DBG-049 | Probe-project debugging-guide entries #24-#28 | deployed | Five field-earned checklist entries: runtime-path config, harness input, export, interfaces, UNIQUE | debugging.md |
| DBG-050 | gamesdesign silent-staleness: result-contract stub (output_field vs output_fields) | deployed | SagaCoordinator honoured only plural key; resolveResultSpec fix shipped 2026-06-18 | debugging.md |
| DBG-051 | Assumed-status-values trap | deployed | Never assume a status column's vocabulary; always SELECT DISTINCT first | debugging.md |
| DBG-052 | "Renders empty" diagnostic method (data-binding, not template) | deployed | 5-step method proving an empty shell is a data-binding bug, not a template failure | debugging.md |
| DBG-053 | rendered_html is a snapshot, not a live view | deployed | Template migrations don't retroactively affect already-built pages' frozen renders | debugging.md |
| DBG-054 | Isolated build test methodology (throwaway test-page pattern) | deployed | Minimal test page through the full build path attributes a bug to one pipeline layer | debugging.md |
| DBG-055 | error_step must live inside step.Config, not at step level | deployed | Step-level error_step silently ignored; dormant instances existed across tool agents | debugging.md |
| DBG-056 | Stage-by-stage rebuild verification and the false-complete rule | deployed | Five-stage A-E method; complete + unchanged components = the old false-complete | debugging.md |
| DBG-057 | Code-retrieval corpus staleness masquerading as retrieval-quality problem | deployed | code_symbols index built from a year-old stale checkout, not a retrieval bug | debugging.md |
| DBG-058 | Spawn-consumed columns lesson: seeds must copy infra columns from a live donor | deployed | Missing command column boots generic entrypoint; dispatcher's call goes unheard | debugging.md |
| DBG-059 | orchestration_state_audit: temporary attachable trigger for state races | deployed | AFTER UPDATE trigger capturing every transition; explicitly removed after use | debugging.md |
| DBG-060 | Message-flow logging / observability plan (never fully built) | aspirational | Early MessageFlowLogger aspiration; only zap logs + processing_history ever built | debugging.md |
| DBG-061 | Orchestration environment reset runbook (clean-slate test-cycle procedure) | deployed | Standard truncate/scale/topic-delete procedure repeated across early docs | debugging.md |
| DBG-062 | Early message-routing failure-mode catalogue | deployed | Seven traced early bugs behind every core architectural convention | debugging.md |
| DBG-063 | Parent-timeout vs child-HITL race | deployed | Parent call_agent timeout fires before child HITL answered, losing the pause | debugging.md |
| DBG-064 | Orchestration debug log taxonomy (early ancestor of the formal guide) | superseded | DEBUGaa grep targets + pg_stat_activity lock triage; ancestral to 016/016b | debugging.md |
| DBG-065 | Mode A/Mode B broken-template taxonomy + pre-extraction JS-shell class | deployed | <no value> repair vs regeneration routing; pre-separateInlineJS shells never got JS | debugging.md |
| DBG-066 | Snapshot-shadowing defect (version+1000 outranks active row) | superseded | snapshot_agent() rows sorted ahead of active in naive ORDER BY version loaders | debugging.md |
| DBG-067 | Secret hygiene: image-provider API keys logged in plaintext | partial | STABILITY/BANANA keys in logs; rotation repeatedly deferred across sessions | debugging.md |
| DBG-068 | Adapter-vs-chassis deployment drift | partial | Separate K8s resources; chassis rebuild doesn't refresh the adapter binary | debugging.md |
| DBG-069 | Launcher reply-topic own-vs-parent derivation (Decision D4) | deployed | Intermediate replies must use own ResponsesTopic, never parent's | debugging.md |
| DBG-070 | gpu-provisioner output-shape flattening (output_fields plural vs singular) | deployed | Same output_field/output_fields bug class as DBG-050 on a different agent | debugging.md |
| DBG-071 | Marker/attribute REPLACE anchoring + hidden-vs-author-CSS landmines | deployed | Bare-string attribute replace corrupts inline querySelector; hidden loses to author CSS | debugging.md |
| DBG-072 | Problem-category taxonomy for component/tool defects | deployed | Shared greppable vocabulary tagging incidents so patterns roll up to the guide | debugging.md |
| DBG-073 | Workflow monitoring REST endpoints (built but apparently unused) | unknown | /monitor/* endpoints documented as built but never seen used; psql/db-inspector instead | debugging.md |
| DBG-074 | kcat + db-inspector operational runbook | deployed | Early ops playbook for triggering/tracing workflows in the live cluster | debugging.md |
| RSH-001 | Dual-signal self-heal on missing spec dependency | deployed | Loud error log + queued recovery work item; two-strike rule caps retries | resilience-self-heal.md |
| RSH-002 | Composition resolver orphan-rows policy | aspirational | Tolerate cheap orphaned rows from failed installs; sweep via database-cleanup | resilience-self-heal.md |
| MIGG-001 | Proposed migration runner/ledger for hand-applied agent-def changes | aspirational | Never built; only manual "2d state check" stands in; responds to DBG-010's incident | migration-governance.md |
| SQLC-001 | SQL needle-gate surgery pattern (guarded, idempotent, reversible DB edits) | deployed | Same method as DBG-016, extracted independently under this proposed category | sql-change-management.md |
