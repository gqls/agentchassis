#!/bin/bash
# ============================================================================
# trigger-audit.sh — Run an audit/review agent against a specific site
#
# Usage:
#   ./trigger-audit.sh <agent_type> <site_id> <domain>
#
# Examples:
#   ./trigger-audit.sh design-audit-agent 1368e337-dd1d-4799-bbb3-8221a1b79bcc finetuning.uk
#   ./trigger-audit.sh site-review-agent 5fe15466-4e2e-4ff2-981e-98c1b7074002 gaswholesalers.com
#   ./trigger-audit.sh visual-design-auditor 1368e337-dd1d-4799-bbb3-8221a1b79bcc finetuning.uk
#
# Available agents:
#
#   TOP-LEVEL (run these — they spawn the group agents automatically):
#     design-audit-agent      — Runs visual-design-auditor + content-quality-auditor
#     site-review-agent       — Runs content-quality-auditor + strategic alignment review
#
#   GROUP AGENTS (can also run individually for targeted checks):
#     visual-design-auditor   — Colour, spacing, typography, dark sections, responsive
#     content-quality-auditor — Tone, content gaps, CTA, differentiation
#
#   FIX AGENTS (not auditors, but useful for manual triggering):
#     site-component-linker     — Links site_components to style collection templates
#     component-template-fixer  — CSS injection, element removal, slot alignment
#     page-build-handler        — Generates + persists page content (wraps page-content-writer)
#
# SAFETY: Each agent operates on ONE site only (the site_id you pass).
# It will NOT affect other sites. Work items are created with status 'detected'
# and need triage before the dispatch loop processes them.
# ============================================================================

AGENT_TYPE="${1:?Usage: $0 <agent_type> <site_id> <domain>}"
SITE_ID="${2:?Usage: $0 <agent_type> <site_id> <domain>}"
DOMAIN="${3:?Usage: $0 <agent_type> <site_id> <domain>}"

# design-audit-agent
# site-review-agent
# visual-design-auditor
# content-quality-auditor
# site-component-linker
# component-template-fixer
# page-build-handler
# blog-content-planner

# main ones
AGENT_TYPE="design-audit-agent"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

AGENT_TYPE="design-audit-agent"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

AGENT_TYPE="design-audit-agent"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

AGENT_TYPE="design-audit-agent"
SITE_ID="2a8ebf9c-20a2-4c39-b191-840b012371da"
DOMAIN="ai-agent-orchestration.com"

AGENT_TYPE="site-review-agent"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

AGENT_TYPE="site-review-agent"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

AGENT_TYPE="site-review-agent"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

AGENT_TYPE="site-review-agent"
SITE_ID="2a8ebf9c-20a2-4c39-b191-840b012371da"
DOMAIN="ai-agent-orchestration.com"

---

 eac60db8-b032-432b-b36d-76f37632045d | system.internal
 1368e337-dd1d-4799-bbb3-8221a1b79bcc | finetuning.uk
 345e15cf-2679-415e-8b11-6e36031ef82f | gamesdesign.co.uk
 00ff3af5-dad8-4770-9f70-3edc267a3c92 | robot-hands.com
 5fe15466-4e2e-4ff2-981e-98c1b7074002 | gaswholesalers.com
 4851f6fc-71cf-4160-a270-e03d6d3e0732 | leopardessconsulting.co.uk
 e1e22a7d-0552-405a-85b3-1b1e51384df5 | vonc.com
 2a8ebf9c-20a2-4c39-b191-840b012371da | ai-agent-orchestration.com

 "visual-design-auditor"
 "site-review-agent"
 "design-audit-agent"
 "content-quality-auditor"
 "site-component-linker"
 "blog-content-planner"
 "design-discovery-agent"

AGENT_TYPE="visual-design-auditor"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

AGENT_TYPE="visual-design-auditor"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

AGENT_TYPE="content-quality-auditor"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

AGENT_TYPE="content-quality-auditor"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

AGENT_TYPE="site-component-linker"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

AGENT_TYPE="site-component-linker"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

AGENT_TYPE="blog-content-planner"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

AGENT_TYPE="design-discovery-agent"
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
DOMAIN="robot-hands.com"

--

./trigger-audit.sh design-audit-agent 4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk
./trigger-audit.sh design-audit-agent 2a8ebf9c-20a2-4c39-b191-840b012371da ai-agent-orchestration.com
./trigger-audit.sh design-audit-agent 1368e337-dd1d-4799-bbb3-8221a1b79bcc finetuning.uk
./trigger-audit.sh design-audit-agent 5fe15466-4e2e-4ff2-981e-98c1b7074002 gaswholesalers.com
./trigger-audit.sh blog-content-planner 1368e337-dd1d-4799-bbb3-8221a1b79bcc finetuning.uk

./trigger-audit.sh design-discovery-agent 00ff3af5-dad8-4770-9f70-3edc267a3c92
=====
-- working copies



#AGENT_TYPE="design-audit-agent"
#SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
#DOMAIN="leopardessconsulting.co.uk"

#AGENT_TYPE="design-audit-agent"
#SITE_ID="2a8ebf9c-20a2-4c39-b191-840b012371da"
#DOMAIN="ai-agent-orchestration.com"

#AGENT_TYPE="design-audit-agent"
#SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
#DOMAIN="finetuning.uk"

#AGENT_TYPE="design-audit-agent"
#SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
#DOMAIN="finetuning.uk"

#AGENT_TYPE="design-audit-agent"
#SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
#DOMAIN="gaswholesalers.com"

#AGENT_TYPE="site-review-agent"
#SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
#DOMAIN="gaswholesalers.com"

#AGENT_TYPE="site-review-agent"
#SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
#DOMAIN="finetuning.uk"

#AGENT_TYPE="visual-design-auditor"
#SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
#DOMAIN="gaswholesalers.com"

#AGENT_TYPE="visual-design-auditor"
#SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
#DOMAIN="finetuning.uk"

#AGENT_TYPE="content-quality-auditor"
#SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
#DOMAIN="gaswholesalers.com"

#AGENT_TYPE="content-quality-auditor"
#SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
#DOMAIN="finetuning.uk"

#AGENT_TYPE="site-component-linker"
#SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
#DOMAIN="gaswholesalers.com"

#AGENT_TYPE="site-component-linker"
#SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
#DOMAIN="finetuning.uk"

AGENT_TYPE="blog-content-planner"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

AGENT_TYPE="design-discovery-agent"
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
DOMAIN="robot-hands.com"

AGENT_TYPE="improvement-loop"
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
DOMAIN="robot-hands.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P -c 1 \
-b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H message_type=request \
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"



-----------------

{
  "action": "orchestrate",
  "config": { "agent_type": "${AGENT_TYPE}" },
  "input_data": {
    "site_id": "${SITE_ID}",
    "domain": "${DOMAIN}"
  }
}

-- See findings created by the audit
SELECT item_type, severity, handler_agent, status, summary
FROM site_work_items
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND source = 'discovery'
  AND created_at > NOW() - INTERVAL '1 hour'
ORDER BY priority;

Step 4: Process the findings
The audit creates items with status: detected. Triage promotes them to triaged. The dispatch loop processes them. You can either:

Run the build-pipeline-trigger (which includes triage + dispatch)
Or triage manually and trigger dispatch:

-- Triage detected items to triaged
UPDATE site_work_items
SET status = 'triaged', triaged_at = NOW()
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND status = 'detected'
  AND domain = 'build';

The triage_detected_items action is already a step in both design-audit-agent and site-review-agent workflows — it runs as their second-to-last step, after write_audit_findings. So items will be promoted to triaged automatically when the audit completes.
If you run the group agents directly (visual-design-auditor or content-quality-auditor), they don't have a triage step — they just write findings as detected. In that case you'd triage manually:


UPDATE site_work_items
SET status = 'triaged', triaged_at = NOW()
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND status = 'detected'
  AND domain = 'build';

-- Work items created by the audit
SELECT item_type, severity, handler_agent, status, LEFT(summary, 80) as summary
FROM site_work_items
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND source = 'discovery'
  AND created_at > NOW() - INTERVAL '10 minutes'
ORDER BY priority;

-- Orchestration status
SELECT owner_agent_type, status, current_step, LEFT(error, 100) as error
FROM orchestration_states
WHERE created_at > NOW() - INTERVAL '10 minutes'
  AND owner_agent_type IN ('content-quality-auditor', 'visual-design-auditor', 'design-audit-agent', 'site-review-agent', 'generic')
ORDER BY created_at DESC LIMIT 5;

---

get them fixed

-- Review what was found before triaging
SELECT item_type, severity, LEFT(summary, 120) as summary
FROM site_work_items
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND status = 'detected'
ORDER BY priority;

-- When ready, promote to triaged
UPDATE site_work_items
SET status = 'triaged', triaged_at = NOW()
WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND status = 'detected';


  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)
    FROM site_work_items
    WHERE created_at > NOW() - INTERVAL '1 hour'
    ORDER BY priority;

---

-- All pending items across all sites
SELECT s.domain, wi.item_type, wi.severity, wi.status, wi.handler_agent,
       LEFT(wi.summary, 70) as summary
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status IN ('detected', 'triaged')
ORDER BY s.domain, wi.priority;


-- Triage all CSS/structural fixes (safe, mechanical)
UPDATE site_work_items
SET status = 'triaged', triaged_at = NOW()
WHERE status = 'detected'
  AND item_type IN ('hardcoded_section_colors', 'needs_design_review');

-- Or triage everything for a specific site
UPDATE site_work_items
SET status = 'triaged', triaged_at = NOW()
WHERE status = 'detected'
  AND site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc';

-- Or triage by severity
UPDATE site_work_items
SET status = 'triaged', triaged_at = NOW()
WHERE status = 'detected' AND severity = 'high';

then trigger pipeline


==============================
===============================
================================
===============================
==============================

AGENT_TYPE="blog-content-planner"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
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
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"



AGENT_TYPE="design-audit-agent"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
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
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"


AGENT_TYPE="design-audit-agent"
SITE_ID="2a8ebf9c-20a2-4c39-b191-840b012371da"
DOMAIN="ai-agent-orchestration.com"


CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
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
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"


AGENT_TYPE="design-audit-agent"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"


CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
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
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"




AGENT_TYPE="design-audit-agent"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"


CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
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
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"




AGENT_TYPE="site-review-agent"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
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
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"



AGENT_TYPE="site-review-agent"
SITE_ID="1368e337-dd1d-4799-bbb3-8221a1b79bcc"
DOMAIN="finetuning.uk"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
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
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"



AGENT_TYPE="site-review-agent"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
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
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"


AGENT_TYPE="site-review-agent"
SITE_ID="2a8ebf9c-20a2-4c39-b191-840b012371da"
DOMAIN="ai-agent-orchestration.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
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
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"

----

image build handler
AGENT_TYPE="image-build-handler"
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
DOMAIN="robot-hands.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo "  Time:           ${TIMESTAMP}"
echo ""

kubectl -n kafka run -i --rm kcat-audit-$(date +%s) \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P -c 1 \
-b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H message_type=request \
-H client_id=$CLIENT_ID \
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"image-build-handler"},"input_data":{"site_id":"00ff3af5-dad8-4770-9f70-3edc267a3c92","domain":"robot-hands.com","item_type":"needs_logo","spec":{"purpose":"logo","image_prompts":{"logo":"A precise, technical logomark for Robot-Hands.com — a stylised robotic gripper or end-effector silhouette rendered in clean geometric lines, suggesting precision engineering and industrial automation. Monochrome or two-tone (electric blue accent on dark background). No cartoonish elements. The mark should read as a technical icon, like something you would see on an engineering schematic or HMI panel. Pair with clean sans-serif wordmark ROBOT-HANDS.COM. Professional, minimal, authoritative."}}}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=${AGENT_TYPE} --tail=50"
echo ""
echo "Check findings:"
echo "  SELECT item_type, severity, handler_agent, status, LEFT(summary, 80)"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > NOW() - INTERVAL '1 hour'"
echo "  ORDER BY priority;"


orig:
{"action":"orchestrate","config":{"agent_type":"image-build-handler"},"input_data":{"site_id":"00ff3af5-dad8-4770-9f70-3edc267a3c92","domain":"robot-hands.com","item_type":"needs_logo","spec":{"purpose":"logo","image_prompts":{"logo":"A precise, technical logomark for Robot-Hands.com — a stylised robotic gripper or end-effector silhouette rendered in clean geometric lines, suggesting precision engineering and industrial automation. Monochrome or two-tone (electric blue accent on dark background). No cartoonish elements. The mark should read as a technical icon, like something you would see on an engineering schematic or HMI panel. Pair with clean sans-serif wordmark ROBOT-HANDS.COM. Professional, minimal, authoritative."}}}}
