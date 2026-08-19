#!/usr/bin/env python3
"""
wedge-evidence-capture — preserves `bugs_open/029`'s evidence before retention eats it.

WHY THIS EXISTS. On 2026-08-17 an orchestration wedge fired 18 times in four hours:
a `build-dispatch-loop` freezing in EXECUTING_STEP at a `process_item_iter_N_spawn_handler`
step, seconds after its spawn handshake SUCCEEDED, holding no `waiting` awaited request and
so invisible to the timeout driver and the retry path. By 2026-08-19 every one of those rows
was GONE and the diagnosis that needed them had nothing to read.

Nothing was misconfigured — it is the ordinary retention. `database-cleanup` (a live
`scheduled_tasks` row, hourly) deletes:

    COMPLETED/FAILED  orchestration_states  older than 24 HOURS   (by updated_at)
    EXECUTING_STEP / AWAITING_RESPONSES     older than  4 HOURS   (by updated_at)

⚠ READ THE LIVE ROW, NOT THE SEED: `docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql`
says 7 days / 24 hours for those two arms. The LIVE `pre_query` says 24 hours / 4 hours.
The live one is what deleted the evidence.

THE RACE THIS JOB EXISTS TO WIN. The stale reaper terminates a frozen orchestration at
**4 hours**, and cleanup arm 4 DELETES an EXECUTING_STEP row at **4 hours**. Those are the
same threshold. When the reaper wins you get a FAILED row with `reaper: stale EXECUTING_STEP`
that survives 24 more hours; when cleanup wins the wedge is deleted having never been
recorded at all. **So the 18 instances we saw are the ones where the reaper happened to win,
and the true rate is a lower bound.** Capturing only reaped rows would inherit that bias,
which is why population (a) below is the important half.

WHAT IT CAPTURES, hourly:
  (a) LIVE wedges — EXECUTING_STEP/AWAITING_RESPONSES, no `waiting` awaited request, frozen
      longer than FROZEN_MINUTES. Caught while the row still exists, hours before either
      the reaper or cleanup reaches it. This is the half that is otherwise lost.
  (b) REAPED wedges — FAILED carrying the reaper's stale marker.

For each NEW orchestration it writes one `doc_notes` row holding the orchestration row AND
every one of its `awaited_requests` rows as JSON — the awaited rows are where the signature
lives (which iteration errored, the duplicate spawn registration, the retry windows), and
they are deleted by CASCADE when the parent goes.

It also writes ONE summary row per run, **on clean results too**, so a MISSING summary row
means THE JOB DID NOT RUN and can never be read as "nothing is wrong".

DEDUPE: an orchestration already captured is skipped, so a wedge sitting frozen for hours
is stored once, not once per tick.

Modelled on site-discovery-staleness-check (same image, same secret, same doc_notes
convention, same direct-Postgres constraint — no pods/exec RBAC here).
"""

import json
import os
import subprocess
import sys

# How long a row must be silent before it counts as frozen. Must sit well BELOW the 4-hour
# reaper/cleanup threshold so there is time to capture, and well ABOVE the longest legitimate
# gap between state writes while a local action runs. 30 minutes gives ~3.5h of lead time.
# ⚠ A step AWAITING a response is excluded by the no-waiting-row clause regardless of age, so
# a legitimately long await (call_diagnoser declares 1800s) is NOT a false positive here.
# Env-overridable so the capture path can be INDUCED and proven: a job that has only ever
# reported "nothing to capture" has not demonstrated it can capture. Set WEDGE_FROZEN_MINUTES=0
# in a one-off job and every in-flight orchestration between steps qualifies, which exercises
# the per-wedge render + insert + dedupe for real. Production leaves it unset.
FROZEN_MINUTES = int(os.environ.get("WEDGE_FROZEN_MINUTES", "30"))

CATEGORY = "wedge-evidence"
# ⚠ doc_notes has CHECK (subject_type IN ('tool','pipeline','experience','action',
# 'experience-pattern','landmine','component','decision')). 'orchestration' is NOT among them
# and the insert fails at RUNTIME, in-cluster, long after the SQL looked fine locally. The
# sibling check jobs all use 'pipeline'; so does this one.
SUBJECT_TYPE = "pipeline"
# The per-wedge key is PREFIXED rather than a bare uuid: a bare uuid in a shared notes table
# says nothing about what it keys, and this must not collide with anything else uuid-keyed.
KEY_PREFIX = "wedge-evidence:"
SUMMARY_SUBJECT_KEY = "wedge-evidence-capture"


def psql(sql, password, host):
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    out = subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-tA", "-v", "ON_ERROR_STOP=1", "-c", sql],
        env=env, check=True, capture_output=True, text=True,
    )
    return out.stdout.strip()


# Both populations in one round trip, each row carrying its full awaited_requests set.
# `already` is the dedupe set, read from doc_notes rather than kept in the job.
CANDIDATES_SQL = """
WITH already AS (
    SELECT subject_key FROM doc_notes
     WHERE subject_type = '%(subject_type)s' AND categories ? '%(category)s'
),
live AS (
    SELECT os.orchestration_id, 'live' AS kind
      FROM orchestration_states os
     WHERE os.status IN ('EXECUTING_STEP', 'AWAITING_RESPONSES')
       AND os.last_activity < now() - interval '%(frozen)d minutes'
       AND NOT EXISTS (
           SELECT 1 FROM awaited_requests ar
            WHERE ar.orchestration_id = os.orchestration_id AND ar.status = 'waiting')
),
reaped AS (
    SELECT os.orchestration_id, 'reaped' AS kind
      FROM orchestration_states os
     WHERE os.status = 'FAILED' AND os.error LIKE 'reaper: stale EXECUTING_STEP%%'
),
candidates AS (
    SELECT * FROM live UNION SELECT * FROM reaped
)
SELECT coalesce(json_agg(row_to_json(c))::text, '[]') FROM (
    SELECT c.kind,
           os.orchestration_id::text,
           os.owner_agent_type, os.current_step, os.status,
           os.correlation_id::text,
           os.created_at, os.last_activity, os.updated_at,
           os.error,
           (os.last_activity - os.created_at)::text AS ran_for,
           (SELECT coalesce(json_agg(row_to_json(a) ORDER BY a.sent_at), '[]')
              FROM (SELECT ar.step_name, ar.retry_version, ar.sent_at, ar.timeout_at,
                           (ar.timeout_at - ar.sent_at)::text AS window,
                           ar.status, ar.target_agent_type, ar.request_id
                      FROM awaited_requests ar
                     WHERE ar.orchestration_id = os.orchestration_id) a) AS awaited_requests
      FROM candidates c
      JOIN orchestration_states os USING (orchestration_id)
     WHERE ('%(key_prefix)s' || os.orchestration_id::text) NOT IN (SELECT subject_key FROM already)
) c;
"""

# Refuse-to-run guard: if the table is empty the queries above cannot find anything, and a
# clean report over a failed read is exactly the shape this estate keeps being bitten by.
TOTALS_SQL = "SELECT json_build_object('total', count(*))::text FROM orchestration_states;"

# Has a summary already been written today (UTC)? The job runs HOURLY so it can win the
# 4-hour race, but a clean summary every hour would put ~8,700 near-empty rows a year into
# doc_notes, which is a SHARED store council seats read. So: a capture ALWAYS writes its
# summary; a clean run writes one only if today has none yet. The invariant survives at daily
# granularity — NO summary row for a given UTC day still means the job did not run that day.
SUMMARY_TODAY_SQL = (
    "SELECT count(*) FROM doc_notes WHERE subject_type = %s AND subject_key = %s "
    "AND categories ? '%s' AND created_at >= date_trunc('day', now() at time zone 'utc');")


def sql_literal(text):
    return "'" + text.replace("'", "''") + "'"


def write_note(subject_key, body, password, host):
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by) "
        "VALUES (%s, %s, %s, '[\"%s\"]'::jsonb, 'wedge-evidence-capture', 'cronjob');"
        % (sql_literal(SUBJECT_TYPE), sql_literal(subject_key), sql_literal(body), CATEGORY)
    )
    path = "/tmp/wedge-note.sql"
    with open(path, "w") as f:
        f.write(sql)
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-v", "ON_ERROR_STOP=1", "-f", path],
        env=env, check=True,
    )


def render(row):
    awaited = row.get("awaited_requests") or []
    lines = [
        "WEDGE EVIDENCE CAPTURE (%s) — bugs_open/029" % row["kind"],
        "",
        "orchestration_id : %s" % row["orchestration_id"],
        "owner_agent_type : %s" % row["owner_agent_type"],
        "current_step     : %s" % row["current_step"],
        "status           : %s" % row["status"],
        "correlation_id   : %s" % row["correlation_id"],
        "created_at       : %s" % row["created_at"],
        "last_activity    : %s   <- the real freeze time, NEVER updated_at" % row["last_activity"],
        "updated_at       : %s" % row["updated_at"],
        "ran_for          : %s" % row["ran_for"],
        "error            : %s" % (row["error"] or ""),
        "",
        "awaited_requests (%d rows, ordered by sent_at) — the signature lives here:" % len(awaited),
    ]
    for a in awaited:
        lines.append(
            "  %-38s rv=%s window=%-10s status=%-10s sent=%s"
            % (a["step_name"], a["retry_version"], a["window"], a["status"], a["sent_at"]))
    lines += [
        "",
        "Captured because both retention arms would have destroyed it: cleanup deletes",
        "EXECUTING_STEP at 4h and FAILED at 24h (live database-cleanup pre_query), and",
        "awaited_requests goes with the parent by CASCADE.",
        "",
        "RAW JSON follows so a diagnosis run can consume it verbatim:",
        json.dumps(row, indent=2, default=str),
    ]
    return "\n".join(lines)


def main():
    password = os.environ.get("CLIENTS_DB_PASSWORD")
    if not password:
        print("REFUSING TO RUN: CLIENTS_DB_PASSWORD is not set.", file=sys.stderr)
        sys.exit(2)
    host = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")

    totals = json.loads(psql(TOTALS_SQL, password, host))
    if totals["total"] == 0:
        print("REFUSING TO RUN: orchestration_states returned 0 rows — the read failed or the "
              "fleet is empty; refusing to write a clean summary over it.", file=sys.stderr)
        sys.exit(2)

    rows = json.loads(psql(
        CANDIDATES_SQL % {"subject_type": SUBJECT_TYPE, "category": CATEGORY,
                          "frozen": FROZEN_MINUTES, "key_prefix": KEY_PREFIX},
        password, host))

    captured = []
    for row in rows:
        write_note(KEY_PREFIX + row["orchestration_id"], render(row), password, host)
        captured.append("%s %s at %s (%s)" % (
            row["kind"], row["owner_agent_type"], row["current_step"], row["orchestration_id"]))
        print("CAPTURED: " + captured[-1])

    summary = [
        "wedge-evidence-capture run summary — bugs_open/029",
        "",
        "orchestration_states rows visible : %d" % totals["total"],
        "frozen-threshold (minutes)        : %d" % FROZEN_MINUTES,
        "newly captured this run           : %d" % len(captured),
    ]
    summary += ["  - " + c for c in captured] if captured else ["  (none — no new wedge)"]
    summary += [
        "",
        "A run that finds nothing STILL writes this row. A MISSING summary row therefore means",
        "the job did not run, which must never be indistinguishable from 'nothing is wrong'.",
    ]
    report = "\n".join(summary)
    print("\n" + report)

    already_today = int(psql(SUMMARY_TODAY_SQL % (
        sql_literal(SUBJECT_TYPE), sql_literal(SUMMARY_SUBJECT_KEY), CATEGORY), password, host))
    if captured or already_today == 0:
        write_note(SUMMARY_SUBJECT_KEY, report, password, host)
        print("summary row written.")
    else:
        print("summary row SUPPRESSED: clean run, and today already has one "
              "(see SUMMARY_TODAY_SQL — the daily invariant is unaffected).")
    sys.exit(1 if captured else 0)


if __name__ == "__main__":
    main()
