UPDATE maintenance_findings
SET status = 'in_progress', fixed_by = 'content-maintenance'
WHERE site_id = $1
AND domain = 'content'
AND status = 'triaged'
AND resolution_path IN ('auto_fix', 'approved')
AND id = (SELECT id FROM maintenance_findings
WHERE site_id = $1 AND domain = 'content'
AND status = 'triaged' AND resolution_path IN ('auto_fix', 'approved')
ORDER BY priority DESC LIMIT 1
FOR UPDATE SKIP LOCKED)
RETURNING *;
```

The `FOR UPDATE SKIP LOCKED` prevents two processes claiming the same row. The orchestrator takes the highest priority item, fixes it, marks it `fixed`, moves to the next. If it runs out of time on this heartbeat cycle, remaining items wait for next cycle.

But this only works if the `domain` field cleanly maps to an orchestrator. What about findings that cross domains? "This page has a broken external link" — is that a links finding or a content finding? The links agent detected it, but the fix might be a content rewrite (replace the sentence that contained the broken link). 

I think the answer is: the discovery agent sets the domain based on who found it. The triage step can reassign the domain based on who should fix it. The links agent finds "broken external link" and writes `domain = 'links'`. Triage assesses it and if the fix is "rewrite the paragraph," it changes `domain = 'content'` and sets `suggested_action = 'rewrite_section'`. Now the content orchestrator picks it up.

This means triage isn't just enriching — it's also routing. Which makes sense. The discovery agents know what's wrong but not always who should fix it.

**The catch-all agent:**

This handles three things:

1. **Unclaimed findings** — triaged items where no domain orchestrator picked them up. Maybe the domain was set wrong, or the resolution path doesn't match any orchestrator's auto-fix capability, or the domain orchestrator isn't enabled for this site. The catch-all scans for findings that have been in `triaged` status for longer than X hours.

2. **Stale findings** — items that were detected or triaged cycles ago and haven't moved. The date reference from 2024 that nobody fixed. The catch-all can escalate severity, send a HITL notification ("these findings have been sitting for 2 weeks"), or auto-set to `ignore` if they're low severity and old enough.

3. **Reclassification** — findings that don't fit neatly into a domain. The catch-all can use LLM judgment to figure out who should own it, reassign the domain, and let the next heartbeat cycle pick it up.

The catch-all runs less frequently — maybe once a day or every few days. It's the safety net that prevents the system from silently accumulating unactioned findings.

Its logic would be roughly:
```
1. Find all findings with status 'triaged' older than 24 hours
   → Check if domain orchestrator is enabled for this site
   → If not enabled: send HITL notification, mark as 'flagged_for_human'
   → If enabled but unclaimed: bump priority, leave for next cycle

2. Find all findings with status 'detected' older than 48 hours (triage never ran)
   → Run basic triage itself (or flag for manual triage)

3. Find all findings with status 'triaged' + resolution_path = 'suggest' older than 7 days
   → Send HITL reminder: "these suggestions are waiting for your approval"

4. Find all findings with status 'triaged' + resolution_path = 'flag' older than 14 days
   → Downgrade to 'monitor' or send escalation HITL

5. Find all findings with status 'monitor' older than configured threshold
   → Re-evaluate: has the situation changed? Re-run discovery on this specific item