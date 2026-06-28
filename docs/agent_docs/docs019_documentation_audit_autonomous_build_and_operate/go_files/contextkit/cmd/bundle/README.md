
RUNBOOK_thin_slice.md

cd docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit
go run ./cmd/bundle \
-analysis /tmp/analysis_repo.json -root ~/projects/agentchassis \
-constitution thin_slice_constitution.md -step debug \
-task "…" \
-scope platform/orchestration/actions/save_page_sections_action.go:SavePageSectionsAction \
-scope platform/orchestration/actions/plan_sections_action.go \
-include platform/orchestration/actions/registry.go \
-doc docs/.../016_debugging_guide_v2_45.md \
-psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
-schema-tables page_components,pages,site_work_items \
-runtime-site gamesdesign.co.uk -runtime-page index \
-capabilities -df-filter snapshot \
-out /tmp/bundle_gamesdesign.md