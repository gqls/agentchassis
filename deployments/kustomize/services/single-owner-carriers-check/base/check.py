#!/usr/bin/env python3
"""
single-owner-carriers-check — the AUTOMATIC half of RFC 006.

WHY THIS EXISTS, and why it is a CronJob rather than a git hook.

`triage_detected_items` promotes every `detected` row on a site. It was a step in
three live agents, so whichever copy ran first took everything and each later copy
honestly reported `promoted: 0` — which the improvement loop read as "nothing to do",
ending on "No issues found — site is clean" over a site holding 67 findings
(bugs_closed/150). RFC 006's owner ruling gave the promoter ONE owner, and
`ActionInputSpec.SingleOwner` + `scripts/audit-single-owner-actions.sh` made a second
carrier reportable.

But that script only ran when somebody remembered to run it. The council's
`architecture` seat named the gap: "the detector is inert-by-omission the same way the
field is inert-by-design". This closes it.

**A pre-commit hook cannot do this job.** At commit time a migration has not been
applied, so the live definitions still look correct — and config on this platform is
routinely changed in the database with no commit at all. The only place the question
can be answered truthfully is against live `agent_definitions`, on a clock.

WHAT IT DOES. Walks every live, active, non-snapshot agent definition; counts how many
DISTINCT AGENTS carry each action declared single-owner; reports any with more than
one. Writes ONE `doc_notes` row per run — on findings AND on a clean result — so that
a missing row means THE JOB DID NOT RUN, which is a different thing from "nothing is
wrong" and must not be silently indistinguishable from it. Exits non-zero on findings
so the Job shows as failed.

TWO THINGS HERE COULD DRIFT FROM THE GO DETECTOR, AND BOTH ARE PINNED BY TESTS
(cmd/config-key-audit/cron_parity_test.go):

  1. DECLARED_SINGLE_OWNER below is a literal, because this container has no Go
     toolchain. `TestCronCheckDeclaredListMatchesTheRegistry` asserts it equals
     `datahelpers.ListSingleOwnerActions()`.
  2. walk_steps() is a SECOND implementation of validation.WalkSteps. That is exactly
     the shape bugs_open/144 warns about — two hand-written traversals go blind in the
     same direction and then agree with each other. So
     `TestCronCheckAgreesWithTheGoDetector` feeds the SAME fixtures to both and
     asserts byte-identical output, including the nested-substep case.

Modes:
  check.py --stdin   < live-workflows.json   # print findings JSON, exit 1 on findings
  check.py                                    # query the DB, report, write doc_notes
"""

import json
import os
import subprocess
import sys

# Mirrors platform/validation/subworkflow.go's maxSubWorkflowDepth. A deeper nest is
# not walked by the runtime either, so walking it here would report a step the
# platform can never execute.
MAX_SUB_WORKFLOW_DEPTH = 8

# Kept in lockstep with datahelpers.ListSingleOwnerActions() by a Go test. Adding an
# action here without the declaration makes this check fire on a state the platform
# does not consider wrong; adding the declaration without this makes it silently miss.
DECLARED_SINGLE_OWNER = [
    "triage_detected_items",
]


def walk_steps(workflow):
    """Yield (path, step, nested) for every step, mirroring validation.WalkSteps.

    Sorted by name at each level, like the Go version, so paths come out in the same
    order and the two outputs can be compared directly.
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


def find_violations(agents, declared):
    """Actions declared single-owner that more than one live agent carries.

    Counts DISTINCT AGENTS, not steps: one agent carrying the step twice is a
    duplicated-step defect, a different class, and reporting it here would put two
    problems behind one message. Paths list every site so a finding is actionable
    without a second query.
    """
    is_single_owner = set(declared)
    owners, paths, seen = {}, {}, set()

    for agent in agents:
        agent_type = agent.get("type") or ""
        for path, step, _nested in walk_steps(agent.get("workflow")):
            action = step.get("action") or ""
            if action not in is_single_owner:
                continue
            paths.setdefault(action, []).append(agent_type + "." + path)
            key = (action, agent_type)
            if key in seen:
                continue
            seen.add(key)
            owners.setdefault(action, []).append(agent_type)

    findings = []
    for action, carriers in owners.items():
        if len(carriers) < 2:
            continue
        findings.append({
            "action": action,
            "owners": sorted(carriers),
            "paths": sorted(paths[action]),
        })
    findings.sort(key=lambda f: f["action"])
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


# The SAME export shape scripts/audit-single-owner-actions.sh and
# audit-unregistered-actions.sh use. Deliberately does not descend into sub-workflows
# in SQL — walk_steps does that, once, the way the runtime does.
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
        "SINGLE-OWNER CARRIER CHECK (RFC 006)",
        "",
        f"live agent definitions walked: {len(agents)}",
        f"actions declared single-owner:  {len(DECLARED_SINGLE_OWNER)} "
        f"({', '.join(DECLARED_SINGLE_OWNER)})",
        f"violations:                     {len(findings)}",
        "",
    ]
    if not findings:
        lines += [
            "No declared single-owner action has a second carrier.",
            "",
            "This row exists on a clean run ON PURPOSE: a MISSING row means the job did",
            "not run, which is not the same as 'nothing is wrong', and the two must not",
            "look alike.",
        ]
        return "\n".join(lines)

    lines.append("VIOLATIONS — an action whose effect is wider than the run calling it")
    lines.append("has picked up a second carrier. The copies now race: whichever runs")
    lines.append("first takes everything and the rest report an honest zero.")
    lines.append("")
    for f in findings:
        lines.append(f"  {f['action']} — carried by {len(f['owners'])} agents:")
        for p in f["paths"]:
            lines.append(f"      {p}")
    lines += [
        "",
        "Decide which agent OWNS the effect and remove the step from the others.",
        "Do NOT add a second per-caller signal — that is what RFC 006 ruled against.",
        "Background: RFC 006, bugs_closed/150, register WFA-006.",
    ]
    return "\n".join(lines)


def write_doc_note(body, password, host):
    tag = "socbody"
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', 'single-owner-carriers', ${tag}${body}${tag}$, "
        "'[\"single-owner-carriers\"]'::jsonb, 'single-owner-carriers-check');"
    )
    path = "/tmp/single-owner-note.sql"
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
        findings = find_violations(agents, DECLARED_SINGLE_OWNER)
        json.dump(findings, sys.stdout, indent=2)
        sys.stdout.write("\n")
        sys.exit(1 if findings else 0)

    password = os.environ.get("CLIENTS_DB_PASSWORD")
    if not password:
        print("REFUSING TO RUN: CLIENTS_DB_PASSWORD is not set.", file=sys.stderr)
        sys.exit(2)
    host = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")

    # Refuse over an empty declaration set: with nothing declared, EVERY possible
    # fleet yields zero findings, so a clean report would be indistinguishable from a
    # broken check. Same refusal the Go tool makes.
    if not DECLARED_SINGLE_OWNER:
        print("REFUSING TO RUN: no actions declared single-owner — a clean report "
              "would be one no fleet could ever fail.", file=sys.stderr)
        sys.exit(2)

    raw = psql(EXPORT_SQL, password, host)
    agents = json.loads(raw) if raw else []
    if not agents:
        print("REFUSING TO RUN: 0 live agent definitions returned — the query failed "
              "or the fleet is empty; refusing to report a clean fleet over it.",
              file=sys.stderr)
        sys.exit(2)

    findings = find_violations(agents, DECLARED_SINGLE_OWNER)
    report = render_report(agents, findings)
    print(report)
    write_doc_note(report, password, host)
    print("\ndoc_notes row written (subject_type='pipeline', "
          "subject_key='single-owner-carriers').")
    sys.exit(1 if findings else 0)


if __name__ == "__main__":
    main()
