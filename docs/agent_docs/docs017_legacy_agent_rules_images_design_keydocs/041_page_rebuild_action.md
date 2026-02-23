maintenance-triage (future orchestrator)
→ scans for stale pages, broken links, missing content, etc.
→ inserts work items into a triage queue table
→ dispatches appropriate specialist agents

page-rebuild (this agent)
→ picked up from triage queue (or spawned manually via generic agent)
→ input: { domain } or { site_id }
→ loads everything it needs from DB
→ rebuilds only needs_rebuild pages
→ deploys


The triage queue table could be something like:
maintenance_queue (
id uuid,
site_id uuid,
task_type text,        -- 'page_rebuild', 'css_update', 'nav_fix', ...
priority int,
payload jsonb,         -- { "pages": ["use-cases","privacy"], "reason": "stale_content" }
status text,           -- 'pending', 'in_progress', 'complete', 'failed'
claimed_by text,       -- agent_id
created_at, updated_at
)