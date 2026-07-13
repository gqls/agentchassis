# Session Summary — 8 March 2026

## What Was Done

### Model Upgrades
- Classifier + planner → claude-opus-4-6
- Content writer, auditors, tool agents → claude-sonnet-4-6
- Extended thinking support added to `anthropic.go` (Go change deployed)
- Extended thinking SQL ready but not enabled (waiting on API version validation)

### Anthropic API Version Fix
- Deployed `anthropic.go` had `2025-01-01` which is invalid
- Must be `2023-06-01` — this caused empty LLM results for blog planner
- **Bug #11**

### Site Specs
- `content_writer_specs_and_blog_block.sql` — content writer now loads site_specs via `read_site_spec` step, LLM prompt includes content_direction/identity/design_intent
- Backfilled site_specs for ai-agent-orchestration.com and leopardessconsulting.co.uk
- Classifier and planner prompts updated to produce design_intent and content_direction
- Classifier writes 4 spec aspects: identity, classification, content_direction, design_intent

### New Agents Created (SQL definitions)
| Agent | Purpose | Status |
|-------|---------|--------|
| domain-submitter | Entry point for new domains via Kafka message | Deployed |
| blog-content-planner | Plans blog posts from spec, creates pages + items | Deployed |
| content-gap-planner | Plans how to fill audit-identified content gaps | Ready to deploy |
| spec-updater | Applies spec field updates from audit findings | Ready to deploy |

### Improvement Loop Updated
- Added design-audit-agent and site-review-agent to the improvement loop chain
- Flow: structural checks → LLM audits → triage → dispatch
- Updated 009_improvement_loop.md with full documentation

### Dispatch Loop Fix
- Batch processing: changed from 50 items to 5 per batch
- Loop back: process_items.next_step changed from "complete" to "load_items"
- Memory limit increased to 1Gi
- **Bug: OOM at 512Mi with 54 items**

### Audit Classification Overhaul
- `write_audit_findings_action.go` rewritten with algorithmic classification
- Loads existing pages upfront, classifies by 6 deterministic rules
- Routes gaps to content-gap-planner instead of page-build-handler
- Routes metadata issues to spec-updater instead of page-build-handler
- No more "page_name: new page needed" items going to page-build-handler

### Page Build Handler Fixes
- Added `load_page_record` Go action (replaces inline SQL)
- Added `check_page_found` guard after page lookup
- Added `check_content_produced` guard after content writer
- Content writer input_mapping uses page_record (DB data) not audit spec
- Deploy step uses page_record.id not sections_saved.page_id

### Scheduler Fix
- Added client_id header to scheduler messages
- Use "system" (with client schema created) or "demo_client"
- **Bug #15: messages rejected by validator**

### Kubernetes Env Var Fix
- Scheduler patch-deployment.yaml: CLIENTS_DB_PASSWORD must come before CLIENTS_DATABASE_URL
- **Bug #12: $(VAR) substitution requires prior definition**

---

## Files Ready to Deploy

### Go (agent-chassis build)
| File | Location | Type |
|------|----------|------|
| `load_page_record_action.go` | `platform/orchestration/actions/` | New |
| `write_audit_findings_action.go` | `platform/orchestration/actions/` | Replace |
| `apply_gap_plan_action.go` | `platform/orchestration/actions/` | New |
| `update_site_spec_from_item_action.go` | `platform/orchestration/actions/` | New |
| `create_blog_posts_action.go` | `platform/orchestration/actions/` | New |
| `check_empty_blog.go` | `platform/orchestration/actions/discovery_checks/` | New |
| `check_empty_sections.go` | `platform/orchestration/actions/discovery_checks/` | Replace |
| `anthropic.go` | `platform/aiservice/` | Verify version header is 2023-06-01 |

### Go (scheduler build)
| File | Change |
|------|--------|
| `cmd/scheduler/main.go` | Add `"client_id": "demo_client"` to fireTrigger headers |

### Registry entries needed
```go
"load_page_record":             { Handler: LoadPageRecordAction, Category: "site", IsLocal: true },
"apply_gap_plan":               { Handler: ApplyGapPlanAction, Category: "site", IsLocal: true },
"update_site_spec_from_item":   { Handler: UpdateSiteSpecFromItemAction, Category: "site", IsLocal: true },
"create_blog_posts":            { Handler: CreateBlogPostsAction, Category: "site", IsLocal: true },
```

### SQL (run after Go deploy)
| File | Run order | Purpose |
|------|-----------|---------|
| `update_page_build_handler_workflow.sql` | 1 | Replace inline SQL with load_page_record action |
| `content_gap_planner.sql` | 2 | Create content-gap-planner agent |
| `spec_updater_agent.sql` | 3 | Create spec-updater agent |
| `enable_extended_thinking.sql` | Later | Add budget_tokens to classifier/planner (optional) |

### SQL (already applied this session)
- Model upgrades (opus-4-6 / sonnet-4-6) ✓
- Classifier prompt: design_intent + content_direction output ✓
- Planner prompt: design_intent + content_direction output ✓
- Classifier workflow: write_content_direction_spec + write_design_intent_spec steps ✓
- Content writer: load_site_specs step + prompt update ✓
- Domain-submitter agent + store_submission_spec step ✓
- Blog-content-planner agent ✓
- Dispatch loop: batch processing (5 items, loop back) ✓
- Improvement loop: design-audit-agent + site-review-agent steps ✓
- Site specs backfill for ai-agent-orchestration.com + leopardessconsulting.co.uk ✓

---

## Bug Tally (add to dev guide)

| # | Bug | Root cause | Rule |
|---|-----|-----------|------|
| 11 | Invalid Anthropic API version header | Changed to 2025-01-01 which doesn't exist | Use 2023-06-01, verify against docs before changing |
| 12 | K8s env var ordering | $(VAR) only resolves vars defined earlier | Define secrets before connection strings that reference them |
| 13 | Data path mismatch between steps | Audit spec passed as current_page, missing sections | Trace every path through workflow, check skip shapes |
| 14 | Audit findings missing page structure | write_audit_findings created items without sections | Handlers load structural data from DB, don't rely on spec |
| 15 | Scheduler missing client_id | fireTrigger didn't include it in headers | All Kafka messages need client_id in headers |
| 16 | Wrong column in query (purpose) | Didn't check \d pages before writing SQL | Always \d table_name before writing queries |

---

## Documents Produced
| File | Purpose |
|------|---------|
| `009_improvement_loop_v2.md` | Full improvement loop documentation |
| `014_entity_feed_livedata_architecture.md` | Design rationale for entity/feed/live data |
| `014_entity_feed_livedata_implementation.md` | Step-by-step build guide for phases 2-4 |
| `015_workflow_data_path_validation.md` | Mandatory checklist before deploying workflows |
| `session_bugs_and_fixes.md` | Bugs 1-10 from earlier sessions |
| `submit-domain.sh` | Kafka script to submit new domains |

---

## What's Running Now
- Build pipeline triggered for all sites
- Dispatch loop processing work items in batches of 5
- Auditors ran for all 4 sites, findings in queue
- Page-build-handler failing on `purpose` column (fix in deploy queue)
- Scheduler running but messages rejected (client_id fix in deploy queue)
- dartsonline.com submitted via domain-submitter (pending classifier run)

---

## Next Steps After Deploy

1. **Verify scheduler** — check messages accepted after client_id fix
2. **Verify page builds** — existing pages (services, contact, case-studies) should build with sections from DB
3. **Run audits again** — with updated write_audit_findings, check classification stats
4. **Verify gap routing** — "new page needed" items go to content-gap-planner
5. **Check dartsonline.com** — classifier should have identified sports/entertainment vertical
6. **Blog planner** — retrigger for finetuning.uk after anthropic version fix confirmed
7. **Entity data agent** — Phase 2, Sprint 1 from the implementation guide
