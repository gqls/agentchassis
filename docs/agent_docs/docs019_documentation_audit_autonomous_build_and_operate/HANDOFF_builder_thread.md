# HANDOFF — Builder-Route Thread → fresh context (2026-07-06)

Written per the travelling-docs pattern this project uses (runbooks = PLAN,
NOTES = history): everything the next session needs, nothing it must re-derive.
(The literal contextkit — the diagnosis loop — diagnoses build faults, not
sessions; this document is the manual equivalent for the thread itself.)

## 1. The mission and the thread structure

The platform builds sophisticated multipage websites from a bare domain name:
research the domain AND the vertical's best exemplars, understand WHY they
succeed (reasons, not copies), then plan/design/write/build/deploy — with
tools, feeds, graphics — aiming best-in-class per vertical. Go chassis
(`agentchassis`), Kubernetes (`-n ai-persona-system`), Kafka, Postgres,
deploy = git → GitHub Actions → Backblaze S3. Every agent is an orchestrator
owning a JSON workflow of steps calling Go actions. Builds run as a WORK-ITEM
RELAY: each hop = a `site_work_items` row naming a `handler_agent`; a 30s
scheduler (`build-pipeline-trigger` → `build-dispatch-loop`) claims and
dispatches.

THREE PARALLEL THREADS with hard boundaries:
- **THIS thread (you)**: the relay/spine, the builder map, coordination, the
  §B4 vertical-exemplar-researcher, plus platform items this thread
  discovered. Working docs: RUNBOOK_builder_route.md + NOTES_running_synthesis_v4.md.
- **Tools chat**: tool-pipeline internals, tool travelling-docs, tool-auditor,
  and a drafted diagnose_load_runtime change — HANDS OFF all of it. The
  planned-tool-page seam (§B5) is a JOINT decision, currently ON HOLD (three
  options recorded in RUNBOOK_builder_route.md §B5).
- **Site-quality thread**: RUNBOOK_site_quality.md — everything page-facing on
  dartsonline (chrome/nav, stylesheet, imagery, content depth, links, feeds
  scope, improvement-loop enablement, CTA-resolution). Scope changes that
  alter relay hops route back through THIS thread.

## 2. Current position (verified DONE — do not redo)

- **Diagnosis route (§7) CLOSED**: read-only loop citing code+data+runtime;
  three-tier CONFIRMED diagnosis achieved (run 73ed55c6 of the OLD thread
  numbering). Available as an instrument. RUNBOOK_code_retrieval_route.md.
- **Option A CLOSED**: shared result contract (datahelpers/result_contract.go),
  response size guard, four agents on preferred `result_from`; deployed
  v1.0.1092+.
- **Builder route §B0–§B4 CLOSED**: spine decided = the work-item relay
  (§B3, pre-registered criteria); §B4 vertical-exemplar-researcher LIVE and
  quality-verified end to end on dartsonline.com (site
  5fe8785b-223d-41a3-88ee-c07187622381): three real exemplars → causal
  synthesis (confidence 0.82) → strategy QUOTING the landscape → the
  setup-builder differentiator planned as a page. First bare-domain→deployed-
  site milestone (index/about/contact/new-arrivals committed).
- **Incidents fixed en route** (mechanisms in the runbook): the seed missing
  spawn-consumed columns; image_tag DEFAULT 'latest' = an ancient
  generic.process-era build (the observed generic-boot pod). Snapshots of
  note: classifier e6ca8cca, strategist ffa6b2da, researcher 139362d5,
  diagnose pair 34f4afc8/e8e96d24.
- **PARTIALLY CONFIRMED THIS TURN — one grep left for you**: the makefile
  pins `IMAGE_TAG ?= v1.0.1096` at the head (the `latest` line is COMMENTED
  OUT) and `make release redeploy-agents` is the deploy command. But the
  `redeploy-agents` target body is ROLLOUT RESTARTS ONLY (verified — it never
  touches agent_definitions), so the demonstrated row bump (identical
  updated_at at v1.0.1096 across live rows) happens ELSEWHERE in `release` or
  a companion step — UNVERIFIED WHERE. First-five-minutes task:
  `grep -n "agent_definitions\|image_tag" makefile` and read the `release`
  target body. Either way the practical guard stands: seeds must copy image
  columns from a live donor (the amended seed does); pinned rows demonstrably
  ride deploys. Residual: one line in guidelines 001's New Agent checklist.

## 3. THE TASK QUEUE (this thread's non-overlapping work, in order)

**Q5 — MAIN: classifier consolidation.** USER'S BRIEF (honour it verbatim):
domain-research-classifier is the newer/base; look BOTH ways for what each
can add to the other; BUT site-classifier belongs to the pageflow-builder /
intake-orchestrator path which does NOT use the site-work-items triage
mechanism — differences there may be load-bearing and must be honoured;
**check hard all of this before making any changes.**
Read-first plan:
1. `outputagent_definitions3.txt` (IN THE MANIFEST) = both classifiers' full
   rows — diff their inputs, prompts, outputs, spec writes, chaining.
2. Intake-door usage evidence — the column EXISTS (schema pasted 2026-07-06):
   `SELECT count(*), max(created_at) FROM orchestration_states
    WHERE orchestration_name ILIKE '%intake%';`
   (Full \d orchestration_states is in this handoff's companion paste; note
   orchestration_name varchar(255), site_id, owner_agent_type all exist.)
3. Map pageflow/intake's dependency points on site-classifier's OUTPUT SHAPE
   (intake-orchestrator v3: hitl_confirm_type → confirmed_type.recommended_builder;
   per-builder questionnaire fetch keys off it).
4. THEN propose: merge additions into domain-research-classifier and/or align
   site-classifier — with snapshot migrations; deprecation only with usage
   evidence at zero.
Also sweeps: intake-orchestrator's unreferenced spawn_rerender/call_rerender
steps (hygiene).

**Q8 — taxonomy alignment** (behind Q5): classifier site_type set vs
strategist canonical set — one decision, two snapshot prompt migrations.

**CI-triggered indexing** (self-contained): GitHub Actions step firing the
index-orchestrator envelope with ${GITHUB_SHA} on push (runners in-cluster;
envelope pattern = TRIGGER_code_indexer_v2.sh style, target
index-orchestrator, EXPLICIT ref never HEAD) — retires the §7A staleness
class for the diagnosis corpus. Note the makefile now owns image tags; this
item is about the CODE CORPUS, unrelated to image bumps.

**Background**: Q2 content-researcher vs content_researcher twins (which do
live workflows reference); stale agent_definitions column cosmetics;
guideline-001 checklist line for image columns.

## 4. FILE MANIFEST for the new project folder

MUST-HAVE (working docs):
- RUNBOOK_builder_route.md — the map (§B0 inventory, §B1–§B3 spine, §B4
  researcher, §B5 seam options, milestones, queue) — the thread's memory.
- NOTES_running_synthesis_v4.md — decisions with rationale, chronological.
- HANDOFF_builder_thread.md — this file.

MUST-HAVE (task inputs):
- outputagent_definitions3.txt — BOTH classifier rows (Q5's primary input).
- output.txt — scheduled_tasks full dump + adoption pair + the six-agent §B2
  dump (strategist prompt lives here — Q8's input; intake/pageflow context).
- buildbriefingagentandbuildsiteplanner.txt — the relay middle (§B3 evidence).
- makefile — the bump mechanism + build targets.
- Guidelines: 000_documentation_index, 001_development_guide,
  003_contracts_and_standards (002 architecture optional).

BOUNDARY-AWARENESS (read-only reference):
- RUNBOOK_site_quality.md (what NOT to work here).
- RUNNING_NOTES_travelling_docs_10_.md + toolagentdefinitions.txt (the tools
  chat's territory + today's tool definitions).
- RUNBOOK_code_retrieval_route.md (the diagnosis instrument, if needed).

ON REQUEST ONLY (Go sources — Q5 is workflow-level; fetch fresh if code reads
become necessary): spawn_actions.go, load_work_item_actions.go,
workflow_actions.go, result_spec.go, apply_adoption_plan_action.go.

NOT for this thread's folder: the four darts HTML files (quality thread);
§B4/Option-A migration files (applied; provenance lives in the runbooks —
re-fetch from this thread's outputs only if a REVERT is ever needed).

## 5. Gotchas that cost time this session (read before touching anything)

- Pod label key is **agent-type** (hyphen); `agent_type` selects nothing.
- getAgentDefinition consumes image_repository/image_tag/command/resources/
  health_config/env_vars + gates on **is_active=true**; `command` empty is
  NORMAL (image default entrypoint); `status` column is redundant for spawn.
- `pipeline` on site_work_items = ROUTING namespace ("build") — the
  create_work_item config key is **item_domain**; page items are
  **needs_page** (not needs_content_page).
- Joins to `sites` need wi./ss./p. prefixes on status/created_at (ambiguous).
- Orchestrator can show COMPLETED while the child FAILED — check body.status.
- Result contracts: post-Option-A both readers honour result_from /
  output_fields; keep workflow keys in sync with action contracts (the
  dead-key class caused a 1.27MB Kafka rejection once).
- 0-rows is not decisive until the query itself is checked; converted-paste
  documents arrive EMPTY — upload .txt FILES (0-byte uploads happen; re-save).
- Standing rules: schema before SQL; snapshot_agent before agent_definitions
  updates (live rows filter COALESCE(is_snapshot,false)=false AND deleted_at
  IS NULL); reasonable step sizes; reuse before recreate; structural over
  quick fixes; explicit ref never HEAD; user runs all SQL/kubectl/builds.

## 6. Opening move for the new chat

Paste the constitution (as always) + this handoff + the manifest files. First
action: Q5 step 1 — read outputagent_definitions3.txt and produce the
two-classifier diff table against the user's brief, before any proposal.
