https://claude.ai/chat/3ef4aaf4-fb53-433e-a5c4-64835885b145
#!/bin/bash
# ============================================================================
# GETTING THE V2 SYSTEM RUNNING — step by step
# ============================================================================
# Test order: verify DB → fix compile errors → build → discovery → asset deploy → full orchestrator
# Each step is independently testable. Stop and debug before moving on.

# ============================================================================
# STEP 0: VERIFY DATABASE
# ============================================================================

# Check site_work_items table exists
kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db -c \
  "SELECT COUNT(*) FROM site_work_items;"

# Check agent definitions were inserted
kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db -c \
  "SELECT type, agent_category, status FROM agent_definitions WHERE type IN ('site-work-orchestrator', 'design-discovery-agent', 'completeness-discovery-agent', 'asset-deployer', 'webdesign-agent');"

# Should see 5 rows. If any missing, re-run the SQL files.
# Note: asset-deployer is the existing agent (not "asset-deploy-agent" — that was
# a duplicate we avoided creating, see 001_development_guide Step 0).

# Check finetuning.uk has known issues (good test target)
kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db -c "
  SELECT s.domain,
         s.style_collection_id IS NULL as no_style,
         (SELECT COUNT(*) FROM assets WHERE site_id = s.id) as asset_count,
         (SELECT COUNT(*) FROM page_components pc
          JOIN pages p ON pc.page_id = p.id
          WHERE p.site_id = s.id AND pc.build_status = 'deployed'
            AND (pc.rendered_html IS NULL OR LENGTH(pc.rendered_html) < 50)
            AND COALESCE(pc.slot_name, '') NOT IN ('header','footer','head')
         ) as empty_sections
  FROM sites s WHERE s.domain = 'finetuning.uk';
"

# ============================================================================
# STEP 1: FIX COMPILE ERRORS IN work_item_actions.go
# ============================================================================
# 6x inputs.GetString("x")  →  inputs.Get("x")
# 2x params.ExecutionContext.AgentType  →  params.ExecutionContext.Sender.AgentType
#
# Quick sed (run in your repo checkout):
#
#   sed -i 's/inputs\.GetString(/inputs.Get(/g' platform/orchestration/actions/work_item_actions.go
#   sed -i 's/params\.ExecutionContext\.AgentType/params.ExecutionContext.Sender.AgentType/g' platform/orchestration/actions/work_item_actions.go
#
# Verify:
#   grep -n "GetString\|ExecutionContext\.AgentType[^.]" platform/orchestration/actions/work_item_actions.go
#   (should return nothing)

# ============================================================================
# STEP 2: ADD REGISTRY ENTRIES
# ============================================================================
# In registry.go, add to the SITE section (after mark_maintenance_complete):
#
#   "write_build_items":      { Handler: WriteBuildItemsAction,      Category: "site", Description: "...", IsLocal: true },
#   "load_work_items":        { Handler: LoadWorkItemsAction,        Category: "site", Description: "...", IsLocal: true },
#   "complete_work_item":     { Handler: CompleteWorkItemAction,     Category: "site", Description: "...", IsLocal: true },
#   "fail_work_item":         { Handler: FailWorkItemAction,         Category: "site", Description: "...", IsLocal: true },
#   "run_discovery_checks":   { Handler: RunDiscoveryChecksAction,   Category: "site", Description: "...", IsLocal: true },
#
# Note: NO "load_undeployed_assets" entry — we reuse the existing asset-deployer
# agent with deploy_image_asset (already registered). The asset-deployer now
# resolves s3_uri from asset_id via resolveStorageURIFromAsset.

# ============================================================================
# STEP 3: BUILD AND DEPLOY
# ============================================================================
# Build the image (this is where compile errors surface)
#
#   go build ./...
#
# If clean, push and let CI build the image.
# Update docker tag in agent_definitions if you use explicit tags.

# ============================================================================
# STEP 4: TEST DISCOVERY (simplest — DB queries only)
# ============================================================================
# Run design-discovery-agent against finetuning.uk.
# It should find: missing_css, undeployed_assets, possibly duplicate_palette.
# No LLM calls, no git, no external services — just SQL queries + inserts.

# Using trigger_discovery.sh:
#   ./trigger_discovery.sh finetuning.uk design

# Or manually:
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

kubectl -n kafka run -i --rm kcat-disco-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_discovery","processing_mode":"orchestrator","timeout_seconds":120,"steps":{"spawn_discovery":{"action":"spawn_agent","config":{"role":"discoverer","agent_type":"design-discovery-agent"},"output_field":"discoverer","next_step":"call_discovery","description":"Spawn discovery agent"},"call_discovery":{"action":"call_agent","config":{"agent_type":"design-discovery-agent","target_role":"discoverer","input_mapping":{"domain":"input_data.domain"},"timeout_seconds":60},"output_field":"discovery_result","next_step":"complete","description":"Run discovery checks"},"complete":{"action":"complete_workflow","config":{"output_fields":["discovery_result"]},"description":"Discovery complete"}}}},"input_data":{"domain":"finetuning.uk"}}
JSON

echo "CORRELATION_ID=$CORRELATION_ID"

# Monitor:
#   kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep "$CORRELATION_ID"

# Check results:
kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db -c \
  "SELECT item_type, severity, summary, handler_agent, status, priority
   FROM site_work_items
   ORDER BY created_at DESC LIMIT 20;"

# Expected for finetuning.uk:
#   missing_css       | high   | Site has no custom stylesheet    | webdesign-agent | detected | 50
#   undeployed_asset  | high   | Asset 'logo' not deployed       | asset-deployer  | detected | 60
#   undeployed_asset  | high   | Asset 'hero' not deployed       | asset-deployer  | detected | 60
#   duplicate_palette | low    | Colour palette identical to ... | webdesign-agent | detected | 150

# ============================================================================
# STEP 5: TRIAGE + TEST INDIVIDUAL HANDLERS
# ============================================================================
# Only after discovery works.
# Discovery items start as 'detected'. Dispatch only picks up 'triaged'.
# Test each handler independently before running the full loop.
#
# See 075c_test_handlers_against_site_work_items.sh for handler tests.
# See 075d_trigger_maintenance.sh for the full dispatch loop.
#
# Quick manual triage:

kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db -c "
  UPDATE site_work_items SET status = 'triaged', updated_at = NOW()
  WHERE site_id = (SELECT id FROM sites WHERE domain = 'finetuning.uk')
    AND status = 'detected'
  RETURNING item_type, handler_agent, priority, spec->>'purpose' as purpose;
"

# Test webdesign-agent standalone (just needs domain):
#   ./075c_test_handlers_against_site_work_items.sh css finetuning.uk

# Test asset-deployer standalone (passes asset_id, agent resolves s3_uri itself):
#   ./075c_test_handlers_against_site_work_items.sh asset finetuning.uk hero
#   ./075c_test_handlers_against_site_work_items.sh asset finetuning.uk logo

# ============================================================================
# STEP 6: TEST FULL MAINTENANCE DISPATCH LOOP
# ============================================================================
# After individual handlers work, test the dispatch loop end-to-end.
# The site-work-orchestrator in maintenance mode:
#   1. Loads site context, selects style, renders chrome
#   2. Loads content items (likely empty for maintenance)
#   3. Loads ALL triaged fix items (no handler filter)
#   4. fix_items_loop: for each item, spawn→call handler dynamically
#      - missing_css → spawns webdesign-agent → CSS deployed
#      - undeployed_asset (hero) → spawns asset-deployer → hero.jpg deployed
#      - undeployed_asset (logo) → spawns asset-deployer → logo.png deployed
#   5. Final design pass, complete
#
# The orchestrator routes to handlers based on each item's handler_agent field.
# No pre-spawning, no hardcoded handler list.

#   ./075d_trigger_maintenance.sh maintain finetuning.uk

# Or manually:
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

kubectl -n kafka run -i --rm kcat-maint-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_orchestrator","processing_mode":"orchestrator","timeout_seconds":900,"steps":{"spawn_orchestrator":{"action":"spawn_agent","config":{"role":"site_orchestrator","agent_type":"site-work-orchestrator"},"output_field":"orchestrator","next_step":"call_orchestrator","description":"Spawn site work orchestrator"},"call_orchestrator":{"action":"call_agent","config":{"target_role":"site_orchestrator","input_mapping":{"domain":"input_data.domain","mode":"input_data.mode"},"timeout_seconds":600},"output_field":"orchestrator_result","next_step":"complete","description":"Run maintenance dispatch loop"},"complete":{"action":"complete_workflow","config":{"output_fields":["orchestrator_result"]},"description":"Maintenance complete"}}}},"input_data":{"domain":"finetuning.uk","mode":"maintenance"}}
JSON

echo "CORRELATION_ID=$CORRELATION_ID"

# Monitor:
#   kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep "$CORRELATION_ID"

# Check results:
#   ./075d_trigger_maintenance.sh status finetuning.uk
# Or:
kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db -c "
  SELECT item_type, handler_agent, status, priority,
         spec->>'purpose' as purpose,
         result->>'commit_sha' as commit_sha,
         updated_at::timestamp(0) as updated
  FROM site_work_items
  WHERE site_id = (SELECT id FROM sites WHERE domain = 'finetuning.uk')
  ORDER BY priority, created_at;
"

# Expected end state:
#   missing_css      | webdesign-agent | complete | 50  | (null) | abc123...
#   undeployed_asset | asset-deployer  | complete | 60  | hero   | def456...
#   undeployed_asset | asset-deployer  | complete | 60  | logo   | ghi789...

