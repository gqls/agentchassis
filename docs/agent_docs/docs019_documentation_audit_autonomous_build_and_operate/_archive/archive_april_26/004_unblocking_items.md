There are three ways things get blocked, and each has a different unblock path:
1. Handler doesn't exist — the claim action marks it blocked because the agent type isn't in agent_definitions. Unblocked automatically by the feasibility-recheck scheduled task (already created) when the agent is deployed. No human action needed.
2. Spec says it's blocked — the classifier marks a feature as status: blocked in the spec. These aren't work items yet — they're entries in the spec's pages/features array. When the classifier or a human changes the status to planned, a work item gets created. The feasibility-recheck doesn't apply here because there's no work item to promote.
3. Manually blocked — a human or the audit agent decides something shouldn't be actioned. Unblocked by manual SQL or a future admin API.
   For cases 1 and 3, the unblock is:
   sql-- Manual unblock of specific items
   UPDATE site_work_items
   SET status = 'triaged', error = NULL, updated_at = NOW()
   WHERE id = '<item_id>' AND status = 'blocked';

-- Unblock all items for a specific handler (after deploying that handler)
UPDATE site_work_items
SET status = 'triaged', error = NULL, updated_at = NOW()
WHERE handler_agent = 'tool-deployer' AND status = 'blocked';
For case 2, promoting a spec entry to a work item:
sql-- Create work item from a previously blocked spec entry
INSERT INTO site_work_items (
site_id, source, domain, item_type, priority,
handler_agent, status, created_by, summary, spec
) VALUES (
'<site_id>', 'spec', 'build', 'needs_content_page', 60,
'page-build-handler', 'triaged', 'human',
'Build pricing page (previously blocked)',
'{"page_name": "pricing", "sections": ["hero", "pricing-tiers", "faq", "call-to-action"]}'::jsonb
);
Now let me add the blocked-item filtering to write_audit_findings:Contemplated integrating blocked-item validation into deduplication logicContemplated integrating blocked-item validation into deduplication logicNow I'll update the file — add blocked-item checking to the existing dedup section:Debugged database connection handling and resource managementDebugged database connection handling and resource managementLet me fix a potential issue — the defer blockedRows.Close() keeps the result set open while we make more queries in the loop:Synthesized unblock mechanisms and documented modificationsSynthesized unblock mechanisms and documented modificationsUpdated write_audit_findings_action.go with two layers of blocked-item filtering:
Layer 1: Bulk preload — loads all blocked item keys for the site upfront, checks each finding against the set. Fast O(1) lookup, catches items blocked by the same audit source.
Layer 2: Broader match — for each finding, also checks if a blocked item exists with the same item_type and page, regardless of which audit source created it. Catches items blocked by the claim action (handler doesn't exist) or manually blocked.
Layer 3: Dedup status list — the existing dedup check now includes blocked alongside detected, triaged, claimed. So even if layers 1 and 2 miss something, the dedup key catches it.
The three unblock paths:
How it got blockedHow it gets unblockedClaim action: handler agent not in agent_definitionsAutomatic: feasibility-recheck scheduled task (every 10 min) promotes to triaged when agent appearsSpec says status: blockedManual: human or classifier changes spec status to planned, creates work itemManual block by humanManual: UPDATE site_work_items SET status = 'triaged' WHERE id = '...'
A future improvement is adding a load_blocked_items step to the audit agent workflows that passes blocked item summaries into the LLM prompt — so the LLM doesn't even analyze gaps it knows are blocked. That saves tokens but isn't needed for correctness, since the Go-level filtering prevents the work items from being created regardless.

