#!/usr/bin/env python3
"""
optional-key-budget-check — the AUTOMATIC half of RFC 022's counter.

WHY THIS EXISTS. The owner's RFC 022 ruling (2026-08-11) made individually-inert
opt-in fields exempt from architecture review, deliberately accepting a blind
spot: nothing would notice a shared action accumulating its tenth one. The
counter (cmd/config-key-audit --optional-key-budget, register WFA-013) closes
that — but only when it runs, and the count changes by a route no commit hook
can see: an agent's live config gaining a carrier happens straight in the
database. Measured while this was being built: the shared-action count moved
21 -> 22 overnight with nobody acting. So the check runs on a clock, like its
RFC 006 sibling, and for the same reason.

THE RULING IT ENFORCES (owner, 2026-08-14): budget N = 10, on SHARED actions
(2 or more distinct live carriers). And the framing is part of the ruling:
sharing is estate design — agents are deliberately reusable across workflows —
so a finding here means an action's ACCUMULATED OPTIONAL SURFACE owes one
review as a whole, never that its reuse is a problem. After that one review the
acknowledged level is the action's baseline (ACKED_LEVELS below), and only
growth PAST the baseline pages again.

WHAT IT DOES. Walks every live, active, non-snapshot agent definition; counts
DISTINCT carriers per action; joins that against the declared optional-key
counts; reports any shared action whose count exceeds both the budget and its
acknowledged baseline. Writes ONE doc_notes row per run — on findings AND on a
clean result — so a missing row means THE JOB DID NOT RUN, which is different
from "nothing is wrong" and must not look like it. Exits non-zero on findings.

THREE THINGS HERE COULD DRIFT FROM THE GO DETECTOR, ALL PINNED BY TESTS
(cmd/config-key-audit/optional_budget_cron_parity_test.go):

  1. OPTIONAL_KEY_COUNTS is a literal (no Go toolchain in this container).
     Pinned to the registry: every action with a non-empty Optional list, at
     its declared count.
  2. ACKED_LEVELS is a literal. Pinned to the repo's source of truth,
     docs/agent_docs/docs024_key_docs_latest/architecture_review/optional_key_budget_acks.json.
  3. walk_steps() is a copy of the single-owner check's mirror of
     validation.WalkSteps (bugs_open/144: hand-written traversals go blind the
     same way and then agree). Pinned by feeding the SAME fixtures to this
     script and the Go detector, including the substeps-wins nesting.
  Plus BUDGET is pinned to scripts/audit-optional-key-budget.sh's default, so
  the daily job and a hand run cannot silently disagree about N.

Modes:
  check.py --stdin [budget]  < live-workflows.json  # findings JSON, exit 1 on findings
  check.py                                          # query the DB, report, write doc_notes

The optional --stdin budget exists FOR THE PARITY TEST: once every over-budget
action carries an acknowledged baseline (the healthy end-state), the ruled
BUDGET can never produce a finding, and the traversal comparison would go
unexercisable — a silent unpinning of the third walk copy. The scheduled run
never passes it; the daily job always enforces BUDGET.
"""

import json
import os
import subprocess
import sys

# The owner's ruled budget (2026-08-14). Pinned to the wrapper's default by a
# parity test — change both or the build fails.
BUDGET = 10

# Shared means at least this many DISTINCT live carriers. A single consumer's
# surface is its own business (the ruling taxes shared seams only).
SHARED_MIN = 2

# Mirrors platform/validation/subworkflow.go's maxSubWorkflowDepth.
MAX_SUB_WORKFLOW_DEPTH = 8

# Every registered action with a non-empty ActionInputSpec.Optional, at its
# declared count. Kept in lockstep with the Go registry by
# TestBudgetCronCountsLiteralMatchesTheRegistry — regenerate with:
#   go run ./cmd/config-key-audit --specs | python3 -c "import json,sys; \
#     [print(f'    \"'+a+f'\": {len(s[\"optional\"])},') for a,s in \
#     sorted(json.load(sys.stdin).items()) if s.get('optional')]"
OPTIONAL_KEY_COUNTS = {
    "analyse_repo_local": 12,
    "append_doc_note": 11,
    "apply_adoption_plan": 1,
    "apply_feed_scores": 2,
    "apply_gap_plan": 3,
    "apply_section_edit": 7,
    "apply_theme_kit": 1,
    "assemble_upload_manifest": 5,
    "bind_site_experience": 6,
    "check_endpoint_health": 1,
    "check_tool_completeness": 1,
    "check_tool_fabrication": 3,
    "checkpoint_for_review": 7,
    "cleanup_stale_topics": 3,
    "collect_external_orders": 3,
    "complete_work_item": 2,
    "compute_checkpoint_keys": 2,
    "compute_component_quality": 5,
    "create_blog_posts": 1,
    "create_rerender_items": 7,
    "create_tool_component": 5,
    "create_tool_cross_link_items": 1,
    "create_work_item": 5,
    "deploy_image_asset": 4,
    "deploy_tool_to_site": 4,
    "derive_brand_head_assets": 1,
    "derive_card_asset": 3,
    "diagnose_assemble_bundle": 16,
    "diagnose_build_gate": 8,
    "diagnose_code_lookup": 5,
    "diagnose_council_decide": 3,
    "diagnose_dormant_agents": 3,
    "diagnose_emit": 4,
    "diagnose_escalate": 4,
    "diagnose_load_runtime": 21,
    "diagnose_persist_fix_plan": 7,
    "diagnose_prepare_fix_commit": 11,
    "diagnose_read_repo_files": 6,
    "diagnose_route": 14,
    "diagnose_run_checks": 6,
    "diagnose_silent_check": 5,
    "diagnose_triage": 9,
    "dispatch_feed_sources": 2,
    "dispatch_thunder_prepare_object_urls": 1,
    "dispatch_verifiers": 3,
    "emit_report_status_files": 1,
    "emit_sprite_css": 1,
    "enrich_fingerprint_with_css": 2,
    "ensure_collection_tasks": 2,
    "execute_vision_prompt": 5,
    "extract_design_fingerprint": 1,
    "extract_interactive_fingerprint": 1,
    "feature_stage_route": 9,
    "fetch_llm_news": 1,
    "fetch_news_search": 1,
    "fetch_rss": 1,
    "fetch_scrape": 1,
    "fix_component_template": 3,
    "fixloop_digest": 4,
    "flag_page_image_rebuild": 2,
    "fork_theme_from_site": 4,
    "git_adapter_request": 2,
    "judge_acceptance_results": 4,
    "load_current_section_content": 2,
    "load_doc_context": 4,
    "load_due_sources": 1,
    "load_edit_context": 4,
    "load_existing_content": 2,
    "load_feed_items_for_event_extraction": 1,
    "load_feed_items_for_triage": 2,
    "load_page_record": 3,
    "load_page_sections_from_spec": 1,
    "load_pending_verifications": 2,
    "load_site_for_rebuild": 1,
    "load_unswept_areas": 2,
    "mark_maintenance_complete": 1,
    "mark_training_run_running": 1,
    "normalize_to_feed_items": 3,
    "persist_diagnosis_note": 10,
    "plan_sections": 8,
    "populate_nav_tables": 1,
    "prepare_rebuild_dispatches": 3,
    "prepare_scrape_batches": 2,
    "prepare_training_data": 2,
    "process_area_sweep": 3,
    "promote_candidates": 2,
    "publish_site": 4,
    "reconcile_section_data": 1,
    "reconcile_site_plan": 1,
    "reconcile_superseded_reviews": 2,
    "record_vision_finding": 4,
    "refresh_directory_claims": 1,
    "refresh_evidence_base": 1,
    "refresh_product_specs": 4,
    "remove_duplicate_page_sections": 1,
    "rename_tool_identity": 1,
    "render_directory": 3,
    "render_model_directory": 3,
    "render_news_section": 4,
    "render_rss_feed": 2,
    "request_browser_run": 6,
    "request_component_browser_run": 7,
    "request_render_audit": 7,
    "rerender_page_sections": 3,
    "resolve_internal_links": 2,
    "retract_asset_files": 4,
    "retract_page_deployment": 6,
    "revalidate_review_queue": 4,
    "rewrite_negations": 1,
    "save_page_meta_description": 5,
    "scan_sites_for_maintenance": 2,
    "score_grippers": 6,
    "select_representative_content": 2,
    "store_crawl_batch": 1,
    "store_generated_component": 6,
    "training_data_export": 6,
    "update_site_spec_from_item": 2,
    "update_source_timestamps": 1,
    "verify_acceptance_predicates": 1,
    "verify_site_experience": 7,
    "write_audit_findings": 1,
    "write_build_items": 2,
    "write_doc_plan": 8,
    "write_experience_pattern": 4,
    "write_feed_items": 2,
    "write_render_audit_findings": 2,
    "zip_deliverable": 3,
}

# Reviewed baselines. Source of truth is the repo's
# architecture_review/optional_key_budget_acks.json; this literal mirrors its
# `count` fields and a parity test fails the build if they drift. An action at
# or under its baseline is quiet; growth PAST the baseline pages again.
ACKED_LEVELS = {
    "analyse_repo_local": 12,
    "append_doc_note": 11,
    "diagnose_prepare_fix_commit": 11,
    "git_commit": 11,
}


def walk_steps(workflow):
    """Yield (path, step, nested) for every step, mirroring validation.WalkSteps.

    A copy of single-owner-carriers-check's mirror — parity-tested against the
    Go traversal on fixtures that include the substeps-wins nesting.
    """
    steps = (workflow or {}).get("steps") or {}
    if not isinstance(steps, dict):
        return
    for name in sorted(steps):
        step = steps[name]
        if not isinstance(step, dict):
            continue
        path = "steps." + name
        yield path, step, False
        yield from _walk_nested(path, step, 0)


def _walk_nested(path, step, depth):
    if depth >= MAX_SUB_WORKFLOW_DEPTH:
        return
    config = step.get("config")
    if not isinstance(config, dict) or not config:
        return

    raw_substeps = config.get("substeps")
    raw_sub_workflow = config.get("sub_workflow")

    # substeps WINS when both are present and substeps is non-empty — the
    # runtime takes substeps and ignores the other half (subworkflow.go).
    if isinstance(raw_substeps, dict) and raw_substeps:
        sub_path, sub_steps = path + ".substeps", raw_substeps
    elif isinstance(raw_sub_workflow, dict):
        sub_path = path + ".sub_workflow"
        sub_steps = raw_sub_workflow.get("steps")
        if not isinstance(sub_steps, dict):
            return
    else:
        return

    for name in sorted(sub_steps):
        nested = sub_steps[name]
        if not isinstance(nested, dict):
            continue
        nested_path = sub_path + "." + name
        yield nested_path, nested, True
        yield from _walk_nested(nested_path, nested, depth + 1)


def find_over_budget(agents, budget=None):
    """Shared actions whose optional-key count exceeds both the budget and their
    acknowledged baseline. Counts DISTINCT carriers, not steps — one agent
    calling an action from three steps is one consumer's design."""
    if budget is None:
        budget = BUDGET
    carriers, seen = {}, set()
    for agent in agents:
        agent_type = agent.get("type") or ""
        for _path, step, _nested in walk_steps(agent.get("workflow")):
            action = step.get("action") or ""
            if not action or (action, agent_type) in seen:
                continue
            seen.add((action, agent_type))
            carriers.setdefault(action, []).append(agent_type)

    findings = []
    for action, count in OPTIONAL_KEY_COUNTS.items():
        agents_for = sorted(carriers.get(action, []))
        if len(agents_for) < SHARED_MIN:
            continue
        if count <= max(budget, ACKED_LEVELS.get(action, 0)):
            continue
        findings.append({
            "action": action,
            "optional_keys": count,
            "consumers": len(agents_for),
            "agents": agents_for,
        })
    findings.sort(key=lambda f: (-f["optional_keys"], f["action"]))
    return findings


def psql(sql, password, host):
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    out = subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-tA", "-v", "ON_ERROR_STOP=1", "-c", sql],
        env=env, check=True, capture_output=True, text=True,
    )
    return out.stdout.strip()


# The SAME export shape every live-join mode of config-key-audit reads.
# Deliberately no sub-workflow descent in SQL — walk_steps does that, once,
# the way the runtime does.
EXPORT_SQL = """
SELECT COALESCE(jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow')), '[]'::jsonb)
FROM agent_definitions
WHERE deleted_at IS NULL
  AND COALESCE(is_snapshot,false) = false
  AND is_active
  AND default_config ? 'workflow';
"""


def render_report(agents, findings):
    lines = [
        "OPTIONAL-KEY BUDGET CHECK (RFC 022, owner ruling 2026-08-14: N = 10)",
        "",
        f"live agent definitions walked:   {len(agents)}",
        f"actions declaring optional keys: {len(OPTIONAL_KEY_COUNTS)}",
        f"acknowledged baselines:          {len(ACKED_LEVELS)}",
        f"findings:                        {len(findings)}",
        "",
    ]
    if not findings:
        lines += [
            "No shared action's optional-key count exceeds the budget or its baseline.",
            "",
            "This row exists on a clean run ON PURPOSE: a MISSING row means the job did",
            "not run, which is not the same as 'nothing is wrong', and the two must not",
            "look alike.",
        ]
        return "\n".join(lines)

    lines.append("OVER BUDGET — a shared action's accumulated optional surface has grown")
    lines.append("past the ruled budget (or past its reviewed baseline). Sharing itself is")
    lines.append("estate design and is NOT the finding; the accumulated surface is.")
    lines.append("")
    for f in findings:
        lines.append(f"  {f['action']} — {f['optional_keys']} optional keys, "
                     f"carried by {f['consumers']} agents:")
        lines.append(f"      {', '.join(f['agents'])}")
    lines += [
        "",
        "Each flagged action owes ONE architecture review of its surface as a whole;",
        "the acknowledged level then goes into optional_key_budget_acks.json (and this",
        "check's ACKED_LEVELS mirror) as its baseline. Do NOT propose de-sharing as",
        "the remedy. Background: RFC 022, register WFA-013.",
    ]
    return "\n".join(lines)


def write_doc_note(body, password, host):
    tag = "okbbody"
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', 'optional-key-budget', ${tag}${body}${tag}$, "
        "'[\"optional-key-budget\"]'::jsonb, 'optional-key-budget-check');"
    )
    path = "/tmp/optional-key-budget-note.sql"
    with open(path, "w") as f:
        f.write(sql)
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-v", "ON_ERROR_STOP=1", "-f", path],
        env=env, check=True,
    )


def main():
    if "--stdin" in sys.argv:
        budget = None
        rest = [a for a in sys.argv[1:] if a != "--stdin"]
        if rest:
            budget = int(rest[0])
        agents = json.load(sys.stdin)
        findings = find_over_budget(agents, budget)
        json.dump(findings, sys.stdout, indent=2)
        sys.stdout.write("\n")
        sys.exit(1 if findings else 0)

    password = os.environ.get("CLIENTS_DB_PASSWORD")
    if not password:
        print("REFUSING TO RUN: CLIENTS_DB_PASSWORD is not set.", file=sys.stderr)
        sys.exit(2)
    host = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")

    # Refuse over an empty declaration map: with nothing declared, EVERY fleet
    # yields zero findings, so a clean report would be indistinguishable from a
    # broken literal. Same refusal the Go tool makes over an empty registry.
    if not OPTIONAL_KEY_COUNTS:
        print("REFUSING TO RUN: OPTIONAL_KEY_COUNTS is empty — a clean report would "
              "be one no fleet could ever fail.", file=sys.stderr)
        sys.exit(2)

    raw = psql(EXPORT_SQL, password, host)
    agents = json.loads(raw) if raw else []
    if not agents:
        print("REFUSING TO RUN: 0 live agent definitions returned — the query failed "
              "or the fleet is empty; refusing to report a clean fleet over it.",
              file=sys.stderr)
        sys.exit(2)

    findings = find_over_budget(agents)
    report = render_report(agents, findings)
    print(report)
    write_doc_note(report, password, host)
    print("\ndoc_notes row written (subject_type='pipeline', "
          "subject_key='optional-key-budget').")
    sys.exit(1 if findings else 0)


if __name__ == "__main__":
    main()
