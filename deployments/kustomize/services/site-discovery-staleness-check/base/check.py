#!/usr/bin/env python3
"""
site-discovery-staleness-check — the watchdog half of bugs_open/230.

WHY THIS EXISTS. Bug 230's mechanism was a silence: every scheduled row targeting
the three site-discovery agents was a disabled one-off, so a site was examined
exactly as often as a human pointed a trigger at it, and nothing anywhere said so.
Migration 346 gives detection a clock (the site-discovery-rotation-* tasks + the
site_discovery_rotation stamp table). This job watches the clock, daily, and makes
the two failure shapes that could quietly recreate the bug VISIBLE:

  1. COVERAGE — a deployed/active site whose stamp (for any of the three agents)
     is missing or older than 2x the rotation period. Catches: the tasks disabled
     (the exact quiet grave 230 documents), a site perpetually deferred by the
     busy-skip, a new site the rotation cannot reach.

  2. CLOSERS vs PRODUCERS — stamps advanced in the last 24h with ZERO discovery
     orchestrations observed in the same window. The stamps are written by the
     pre_query BEFORE the Kafka fire, so they are fire-and-forget: they prove
     selection, never execution (LANDMINES: "last_triggered_at keeps advancing
     while nothing runs"). orchestration_states retains COMPLETED/FAILED rows for
     24h (database-cleanup step 3), which is exactly long enough for a daily job
     to answer "did yesterday's selections actually run?". Caveat, stated rather
     than hidden: a hand-fired discovery run in the same window satisfies this
     check, so it detects a DEAD delivery path, not a lossy one.

Writes ONE doc_notes row per run — on findings AND on a clean result — so a
missing row means THE JOB DID NOT RUN, which must never be indistinguishable from
"nothing is wrong". Exits non-zero on findings. Modelled on
single-owner-carriers-check (same image, same secret, same doc_notes convention,
same direct-Postgres constraint — no pods/exec RBAC here).

GRACE ON INSTALL: for the first rotation period after the tasks are created, a
missing stamp means "queued, still catching up", not "unreachable" — the first
run after migration 346 must not fire 60 false findings. After one full period,
a missing stamp is a finding.
"""

import json
import os
import subprocess
import sys

# Keep in lockstep with migration 346's pre_query interval ('7 days') and the
# task set it seeds. If the owner retunes the period there, retune it here.
ROTATION_PERIOD_DAYS = 7
STALE_AFTER_DAYS = 2 * ROTATION_PERIOD_DAYS

DISCOVERY_AGENTS = [
    "quality-discovery-agent",
    "design-discovery-agent",
    "completeness-discovery-agent",
]


def psql(sql, password, host):
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    out = subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-tA", "-v", "ON_ERROR_STOP=1", "-c", sql],
        env=env, check=True, capture_output=True, text=True,
    )
    return out.stdout.strip()


AGENT_LIST_SQL = "', '".join(DISCOVERY_AGENTS)

# One round trip, one JSON document. site_discovery_rotation is only referenced
# when it exists (guarded below), so this check reports the pre-migration state
# as a finding instead of crashing on it.
STATE_SQL = f"""
SELECT jsonb_build_object(
  'rotation_installed', to_regclass('public.site_discovery_rotation') IS NOT NULL,
  'tasks', (SELECT COALESCE(jsonb_agg(jsonb_build_object(
                'name', name, 'enabled', enabled,
                'created_days_ago', round(extract(epoch FROM now() - created_at) / 86400.0, 1))
              ORDER BY name), '[]'::jsonb)
            FROM scheduled_tasks WHERE name LIKE 'site-discovery-rotation-%'),
  'site_count', (SELECT count(*) FROM sites WHERE status IN ('active','deployed')),
  'orch_last_24h', (SELECT COALESCE(jsonb_object_agg(owner_agent_type, n), '{{}}'::jsonb)
                    FROM (SELECT owner_agent_type, count(*) AS n
                          FROM orchestration_states
                          WHERE owner_agent_type IN ('{AGENT_LIST_SQL}')
                            AND created_at > now() - interval '24 hours'
                          GROUP BY 1) o)
)::text;
"""

ROTATION_SQL = f"""
SELECT jsonb_build_object(
  'stale', (SELECT COALESCE(jsonb_agg(jsonb_build_object(
                'domain', x.domain, 'agent', x.agent,
                'stamp_days_ago', x.days_ago) ORDER BY x.domain, x.agent), '[]'::jsonb)
            FROM (
              SELECT s.domain, a.agent,
                     round(extract(epoch FROM now() - r.last_selected_at) / 86400.0, 1) AS days_ago
              FROM sites s
              CROSS JOIN (VALUES ('{DISCOVERY_AGENTS[0]}'), ('{DISCOVERY_AGENTS[1]}'), ('{DISCOVERY_AGENTS[2]}')) a(agent)
              LEFT JOIN site_discovery_rotation r ON r.site_id = s.id AND r.agent_type = a.agent
              WHERE s.status IN ('active','deployed')
                AND (r.last_selected_at IS NULL
                     OR r.last_selected_at < now() - interval '{STALE_AFTER_DAYS} days')
            ) x),
  'stamps_last_24h', (SELECT COALESCE(jsonb_object_agg(agent_type, n), '{{}}'::jsonb)
                      FROM (SELECT agent_type, count(*) AS n
                            FROM site_discovery_rotation
                            WHERE last_selected_at > now() - interval '24 hours'
                            GROUP BY 1) g)
)::text;
"""


def find_findings(state, rotation):
    findings = []

    tasks = state["tasks"]
    enabled = [t for t in tasks if t["enabled"]]
    if not state["rotation_installed"] or len(enabled) < len(DISCOVERY_AGENTS):
        findings.append({
            "kind": "driver_missing",
            "detail": (
                f"rotation table installed: {state['rotation_installed']}; "
                f"enabled site-discovery-rotation tasks: {len(enabled)}/{len(DISCOVERY_AGENTS)} "
                f"({', '.join(t['name'] for t in tasks) or 'none'}). "
                "This is bugs_open/230's original state: detection has no clock."
            ),
        })
        return findings  # everything downstream would only restate this

    # Grace: within the first rotation period after install, NULL stamps are
    # catch-up, not coverage failure.
    youngest_task_age = min(t["created_days_ago"] for t in tasks)
    in_grace = youngest_task_age < ROTATION_PERIOD_DAYS

    for row in rotation["stale"]:
        if row["stamp_days_ago"] is None and in_grace:
            continue
        findings.append({
            "kind": "site_stale",
            "detail": (
                f"{row['domain']} not selected by {row['agent']} for "
                f"{'ever (no stamp)' if row['stamp_days_ago'] is None else str(row['stamp_days_ago']) + ' days'}"
                f" (threshold {STALE_AFTER_DAYS}d)"
            ),
        })

    stamps = rotation["stamps_last_24h"]
    orchs = state["orch_last_24h"]
    for agent in DISCOVERY_AGENTS:
        if stamps.get(agent, 0) > 0 and orchs.get(agent, 0) == 0:
            findings.append({
                "kind": "fires_without_runs",
                "detail": (
                    f"{agent}: {stamps[agent]} site(s) stamped in the last 24h but zero "
                    "orchestrations observed in the same window — selections are being "
                    "made and the runs are not happening (dead delivery path)."
                ),
            })

    return findings


def render_report(state, rotation, findings):
    lines = [
        "SITE-DISCOVERY STALENESS CHECK (bugs_open/230)",
        "",
        f"active/deployed sites:            {state['site_count']}",
        f"rotation tasks enabled:           {len([t for t in state['tasks'] if t['enabled']])}/{len(DISCOVERY_AGENTS)}",
        f"stamps advanced last 24h:         {json.dumps(rotation['stamps_last_24h']) if rotation else 'n/a'}",
        f"discovery orchestrations last 24h: {json.dumps(state['orch_last_24h'])}",
        f"findings:                         {len(findings)}",
        "",
    ]
    if not findings:
        lines += [
            "Every active/deployed site has been selected by all three discovery agents",
            f"within {STALE_AFTER_DAYS} days, and selections are producing runs.",
            "",
            "This row exists on a clean run ON PURPOSE: a MISSING row means the job did",
            "not run, which is not the same as 'nothing is wrong', and the two must not",
            "look alike.",
        ]
        return "\n".join(lines)

    for f in findings:
        lines.append(f"  [{f['kind']}] {f['detail']}")
    lines += [
        "",
        "Remedies: bugfix_230_discovery_driver/RUNBOOK_discovery_driver.md.",
        "A driver_missing finding means detection has no clock again — that is the",
        "bug itself, not an operational hiccup.",
    ]
    return "\n".join(lines)


def write_doc_note(body, password, host):
    tag = "sdscbody"
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', 'site-discovery-staleness', ${tag}${body}${tag}$, "
        "'[\"site-discovery-staleness\"]'::jsonb, 'site-discovery-staleness-check');"
    )
    path = "/tmp/site-discovery-staleness-note.sql"
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
    password = os.environ.get("CLIENTS_DB_PASSWORD")
    if not password:
        print("REFUSING TO RUN: CLIENTS_DB_PASSWORD is not set.", file=sys.stderr)
        sys.exit(2)
    host = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")

    state = json.loads(psql(STATE_SQL, password, host))

    if state["site_count"] == 0:
        print("REFUSING TO RUN: 0 active/deployed sites returned — the query failed "
              "or the fleet is empty; refusing to report a clean fleet over it.",
              file=sys.stderr)
        sys.exit(2)

    # Only touch the rotation table when it exists; its absence is a finding,
    # not a crash.
    rotation = {"stale": [], "stamps_last_24h": {}}
    if state["rotation_installed"]:
        rotation = json.loads(psql(ROTATION_SQL, password, host))

    findings = find_findings(state, rotation)
    report = render_report(state, rotation, findings)
    print(report)
    write_doc_note(report, password, host)
    print("\ndoc_notes row written (subject_type='pipeline', "
          "subject_key='site-discovery-staleness').")
    sys.exit(1 if findings else 0)


if __name__ == "__main__":
    main()
