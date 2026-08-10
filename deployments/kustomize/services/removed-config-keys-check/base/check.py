#!/usr/bin/env python3
"""
removed-config-keys-check — the AUTOMATIC half of the RFC_021 Q1 owner ruling
(2026-08-10).

WHY THIS EXISTS. `ActionInputSpec.RemovedConfigKeys` (register SCR-007,
bugs_open/234) makes a RETIRED config key a HARD validation error: any live
definition carrying one is rejected on every message from the moment a binary
with the declaration rolls — the agent stops working until the definition is
fixed. That is the design (the warn-only alternative is how a lost flag survived
for months), but it leaves a window: a definition seeded or edited AFTER the
adoption census fails as an agent outage rather than a report line. The council
guardian vetoed the adoption round over exactly this window; the owner's RFC_021
ruling closed it with THIS check rather than per-adoption producer inventories:
census at adoption, plus this job on a clock.

**A pre-commit hook cannot do this job** — same two reasons as
single-owner-carriers-check, on which this is modelled: at commit time a
migration has not been applied, and live config is routinely changed in the
database with no commit at all. The only place the question has a true answer is
live `agent_definitions`, on a clock.

WHAT IT DOES. Walks every live, active, non-snapshot agent definition (all
depths); reports any step whose action declares a config key REMOVED while the
step still carries it. Writes ONE `doc_notes` row per run — on findings AND on a
clean result — so a missing row means THE JOB DID NOT RUN, which must never look
like "nothing is wrong". Exits non-zero on findings so the Job shows failed.

TWO THINGS HERE COULD DRIFT FROM THE GO DETECTOR, AND BOTH ARE PINNED BY TESTS
(cmd/config-key-audit/removed_keys_cron_parity_test.go):

  1. DECLARED_REMOVED below is a literal (no Go toolchain in this container).
     `TestRemovedKeysCronDeclaredListMatchesTheRegistry` asserts it equals the
     action→key-set shape of `datahelpers.ListRemovedConfigKeys()`. Keys only,
     not the replacement messages — the message lives in the Go declaration and
     the register; syncing prose here would be a third copy that drifts.
  2. walk_steps() is a copy of the mirrored traversal in
     single-owner-carriers-check/base/check.py (itself pinned against
     validation.WalkSteps). `TestRemovedKeysCronAgreesWithTheGoDetector` feeds
     the same fixtures to this script and the Go helper, including the
     nested-substep and substeps-beats-sub_workflow cases.

Modes:
  check.py --stdin   < live-workflows.json   # print findings JSON, exit 1 on findings
  check.py                                    # query the DB, report, write doc_notes
"""

import json
import os
import subprocess
import sys

# Mirrors platform/validation/subworkflow.go's maxSubWorkflowDepth.
MAX_SUB_WORKFLOW_DEPTH = 8

# Kept in lockstep with datahelpers.ListRemovedConfigKeys() (action -> sorted keys)
# by a Go test. A key here without the Go declaration makes this fire on a state
# the platform does not reject; a Go declaration missing here makes the ONLY
# automatic guard silently stop asking about it.
DECLARED_REMOVED = {
    "create_work_item": ["spec"],
    "update_page_status": ["notes_field", "validation_issues_field"],
}


def walk_steps(workflow):
    """Yield (path, step, nested) for every step, mirroring validation.WalkSteps.

    Copied from single-owner-carriers-check/base/check.py — the pinned mirror.
    Sorted by name at each level, like the Go version.
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

    # substeps WINS when both are present and substeps is non-empty — the runtime
    # takes substeps and ignores the other half (subWorkflowsOf, subworkflow.go).
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


def find_carriers(agents, declared):
    """Steps whose action declares a key REMOVED while the step still carries it.

    Every carrier is an agent one roll away from refusing to run (or already
    refusing, if the declaring binary is live) — so paths name agent, step path
    and key, actionable without a second query.
    """
    findings = []
    for agent in agents:
        agent_type = agent.get("type") or ""
        for path, step, _nested in walk_steps(agent.get("workflow")):
            action = step.get("action") or ""
            removed_keys = declared.get(action)
            if not removed_keys:
                continue
            config = step.get("config")
            if not isinstance(config, dict):
                continue
            for key in removed_keys:
                if key in config:
                    findings.append({
                        "agent": agent_type,
                        "path": path,
                        "action": action,
                        "key": key,
                    })
    findings.sort(key=lambda f: (f["agent"], f["path"], f["key"]))
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


# The SAME export shape the sibling checks use.
EXPORT_SQL = """
SELECT COALESCE(jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow')), '[]'::jsonb)
FROM agent_definitions
WHERE deleted_at IS NULL
  AND COALESCE(is_snapshot,false) = false
  AND is_active
  AND default_config ? 'workflow';
"""


def render_report(agents, findings):
    declared_desc = ", ".join(
        f"{action}.{key}" for action in sorted(DECLARED_REMOVED)
        for key in DECLARED_REMOVED[action])
    lines = [
        "REMOVED CONFIG KEYS IN USE CHECK (bugs_open/234, RFC_021 Q1 ruling)",
        "",
        f"live agent definitions walked: {len(agents)}",
        f"keys declared removed:         {declared_desc}",
        f"carriers found:                {len(findings)}",
        "",
    ]
    if not findings:
        lines += [
            "No live definition carries a removed config key.",
            "",
            "This row exists on a clean run ON PURPOSE: a MISSING row means the job did",
            "not run, which is not the same as 'nothing is wrong', and the two must not",
            "look alike.",
        ]
        return "\n".join(lines)

    lines.append("CARRIERS — each of these definitions is REJECTED at validation on every")
    lines.append("message by any binary carrying the declaration (the agent stops working")
    lines.append("until the definition is fixed):")
    lines.append("")
    for f in findings:
        lines.append(f"  {f['agent']}.{f['path']}: {f['action']} carries removed key '{f['key']}'")
    lines += [
        "",
        "Fix the DEFINITION (and its seed, or a reseed replays the key): the",
        "replacement spelling is in the action's RemovedConfigKeys message and in",
        "register SCR-007. Do NOT un-declare the key to silence this.",
        "Background: bugs_open/234, RFC_021, register SCR-007.",
    ]
    return "\n".join(lines)


def write_doc_note(body, password, host):
    tag = "rckbody"
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', 'removed-config-keys', ${tag}${body}${tag}$, "
        "'[\"removed-config-keys\"]'::jsonb, 'removed-config-keys-check');"
    )
    path = "/tmp/removed-config-keys-note.sql"
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
        agents = json.load(sys.stdin)
        findings = find_carriers(agents, DECLARED_REMOVED)
        json.dump(findings, sys.stdout, indent=2)
        sys.stdout.write("\n")
        sys.exit(1 if findings else 0)

    password = os.environ.get("CLIENTS_DB_PASSWORD")
    if not password:
        print("REFUSING TO RUN: CLIENTS_DB_PASSWORD is not set.", file=sys.stderr)
        sys.exit(2)
    host = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")

    # Refuse over an empty declaration set: with nothing declared, EVERY possible
    # fleet yields zero findings, so a clean report would be indistinguishable
    # from a broken check. Same refusal the sibling checks make.
    if not DECLARED_REMOVED:
        print("REFUSING TO RUN: no keys declared removed — a clean report would be "
              "one no fleet could ever fail.", file=sys.stderr)
        sys.exit(2)

    raw = psql(EXPORT_SQL, password, host)
    agents = json.loads(raw) if raw else []
    if not agents:
        print("REFUSING TO RUN: 0 live agent definitions returned — the query failed "
              "or the fleet is empty; refusing to report a clean fleet over it.",
              file=sys.stderr)
        sys.exit(2)

    findings = find_carriers(agents, DECLARED_REMOVED)
    report = render_report(agents, findings)
    print(report)
    write_doc_note(report, password, host)
    print("\ndoc_notes row written (subject_type='pipeline', "
          "subject_key='removed-config-keys').")
    sys.exit(1 if findings else 0)


if __name__ == "__main__":
    main()
