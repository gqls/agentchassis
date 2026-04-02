Let me trace the full lifecycle from build through audit:

**The three entry points and what they do:**

```
1. build-pipeline-trigger (scheduled task, every 30s)
   │
   │  Pre-query: finds sites with triaged items in site_work_items
   │  Fires: build-dispatch-loop per site
   │  Scope: processes existing triaged items only — does NOT create new items
   │
   │  This is the worker. It picks up and executes whatever's in the queue.

2. improvement-loop (scheduled via improvement-sweep, every 600s)
   │
   │  Pre-query: finds least-recently-audited site with queue depth < 20
   │  Fires: improvement-loop for that one site
   │  Scope: runs ALL discovery + audit agents, creates items, triages them,
   │         then calls build-dispatch-loop internally to process what it found
   │
   │  This is the full cycle — discover, triage, fix, rerender.

3. Manual trigger (your CLI: ./trigger-audit.sh)
   │
   │  Fires: improvement-loop OR individual agents directly
   │  Scope: whatever you target
```

**How they interact — the timeline for a new site:**

```
T=0    domain-submitter creates site + research work item
       ↓
T=30s  build-pipeline-trigger picks up research item
       → dispatch-loop → domain-research-classifier → creates strategy item
       ↓
T=60s  build-pipeline-trigger picks up strategy item
       → dispatch-loop → domain-strategist → creates briefing item
       ↓
       ... chain continues: briefing → planner → page content items ...
       ↓
T=5m   build-pipeline-trigger picks up page content items
       → dispatch-loop → page-build-handler × N pages
       → all pages deployed
       ↓
T=10m  improvement-sweep scheduled task fires
       Pre-query: "which site was built least recently?"
       → finds the newly built site (or cycles through all sites)
       → fires improvement-loop
       ↓
       improvement-loop runs:
         Step 2: quality-discovery-agent → broken_nav_links, placeholder_contact
         Step 3: design-discovery-agent  → missing_css, hardcoded_colors, stale_components
         Step 4: completeness-discovery  → empty_sections, contamination, templates
         Step 5: design-audit-agent      → LLM visual + content review
         Step 6: site-review-agent       → LLM strategic review
         Step 7: triage (detected → triaged)
         Step 8: check findings
         Step 9: insert rerender item (priority 99)
         Step 10: build-dispatch-loop    → processes all fixes → rerender last
       ↓
T=18m  Site is improved. Next improvement-sweep will pick another site.
```

**The build-pipeline-trigger and improvement-loop DON'T conflict** because they serve different purposes:

| | build-pipeline-trigger | improvement-loop |
|---|---|---|
| Creates items? | No — only processes existing ones | Yes — discovery + audit agents create items |
| Runs audits? | No | Yes — all 5 discovery/audit agents |
| Calls dispatch? | Yes — directly | Yes — internally at step 10 |
| When items exist | Processes them immediately (30s cycle) | Discovers more, then processes everything |
| Concurrency risk | Claims items atomically — two dispatchers skip already-claimed items | improvement-sweep pre-query checks queue depth < 20 |

**The potential overlap:**

If the build-pipeline-trigger fires while the improvement-loop's internal dispatch is running, both try to claim items. The `claim_work_item` action uses atomic UPDATE with `WHERE status = 'triaged'` — only one succeeds. The other skips the item. No duplication.

**Where your manual trigger fits:**

```
Manual: ./trigger-audit.sh improvement-loop <site_id> <domain>
  ↓
  Same as improvement-sweep but YOU choose the site and timing.
  Useful when:
    - You just fixed something and want to verify
    - You want to audit a specific site without waiting 600s
    - You want to test new discovery checks

Manual: trigger a single discovery agent directly
  ↓
  Creates items with status: detected
  They sit there until either:
    a) improvement-loop's triage step promotes them (next 600s cycle)
    b) You manually run: UPDATE site_work_items SET status = 'triaged' WHERE status = 'detected'
    c) build-pipeline-trigger won't touch them — it only sees triaged items
```

**The gap in the current setup:** If you manually trigger just a discovery agent (not the full improvement-loop), the items stay as `detected` until the next improvement-loop run triages them. The build-pipeline-trigger ignores `detected` items. You'd need to either run the full improvement-loop or manually promote them.

For the fixes we're applying today, the manual SQL approach bypasses all of this — we're inserting items directly as `triaged` and resetting stuck items. Once the Go deploy is done, the build-pipeline-trigger (every 30s) will pick them up and start processing.



