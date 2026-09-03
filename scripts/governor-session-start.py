#!/usr/bin/env python3
"""SessionStart hook — say LOUDLY, at the start of every session, when the spend governor is
shedding. Owner ruling 2026-09-03, verbatim: council reviews should be held back "first but
loudly so I can stop doing important things for a while". A doc_notes row is loud to an agent;
this is what makes it loud to the OWNER and to every session they open.

    echo '{}' | ./scripts/governor-session-start.py      # what a session would see

WHAT IT READS. The live governor, in one psql round trip: shed level, month-to-date spend, the
budget, and — via the ONE canonical predicate, not a re-spelling of the ladder — whether
council-gate runs are admitted right now. Prints NOTHING when the level is 0 and council is
admitted: a banner that appears every day is a banner nobody reads.

THE TRADE-OFF, stated. landmines-session-start.py deliberately reads a local file so that a
session start never waits on the cluster and never fails when the kubeconfig has expired
(every 3 days here). This hook cannot: the state lives only in the database. So it is capped
at a short timeout and IT MUST NEVER BREAK OR NOTICEABLY SLOW A SESSION — any failure, timeout
or missing kubeconfig exits 0 with no output. The cost of that posture is the honest one: when
the token has expired, the banner is silent, and silence is NOT "level 0". The daily 752/657
VERIFY habit and the level-change doc_note (mig 753) are the channels that do not depend on
this hook.
"""
import json
import subprocess
import sys

TIMEOUT_S = 6   # typical exec ~1-2 s; anything slower is a cluster problem, not this hook's

SQL = r"""
SELECT json_build_object(
  'enabled', gc.enabled,
  'level', gs.shed_level,
  'mtd', round(COALESCE(gs.mtd_usd, 0), 0),
  'budget', gc.monthly_budget_usd,
  'pct', round(100 * COALESCE(gs.mtd_usd, 0) / NULLIF(gc.monthly_budget_usd, 0), 0),
  'council_admitted', governor_admits_agent('council-gate'),
  'withheld_items', (SELECT count(*) FROM governor_withheld_now),
  'heartbeat_s', round(EXTRACT(epoch FROM now() - gs.computed_at)),
  'last_change', (SELECT left(body, 200) FROM doc_notes
                   WHERE subject_key = 'spend-governor' AND categories ? 'level-change'
                   ORDER BY created_at DESC LIMIT 1)
)::text
FROM governor_config gc, governor_state gs WHERE gc.id = 1 AND gs.id = 1;
"""


def read_governor():
    """One row as a dict, or None on ANY failure (expired token, timeout, no cluster, no row)."""
    try:
        out = subprocess.run(
            ["kubectl", "-n", "ai-persona-system", "exec", "postgres-clients-0", "--",
             "psql", "-U", "clients_user", "-d", "clients_db", "-tAc", SQL],
            capture_output=True, text=True, timeout=TIMEOUT_S, stdin=subprocess.DEVNULL,
        )
        if out.returncode != 0 or not out.stdout.strip():
            return None
        return json.loads(out.stdout.strip().splitlines()[-1])
    except Exception:
        return None


def banner(g):
    """None when there is nothing to shout about."""
    if not g or not g.get("enabled"):
        return None
    level = int(g.get("level") or 0)
    council_ok = bool(g.get("council_admitted", True))
    if level == 0 and council_ok:
        return None
    lines = ["⚠ SPEND GOVERNOR IS SHEDDING — level %d (month-to-date $%s of $%s, %s%%)." % (
        level, g.get("mtd"), g.get("budget"), g.get("pct"))]
    if not council_ok:
        lines.append("⚠ COUNCIL REVIEWS ARE WITHHELD. A council submission will COMPLETE at "
                     "`complete_withheld` without running — that is not the queue, do not retry. "
                     "OWNER RULING 2026-09-03: stop non-essential platform work until the level drops.")
    else:
        lines.append("Council reviews are still admitted at this level.")
    wi = g.get("withheld_items")
    if wi:
        lines.append("%s work item(s) withheld right now — `SELECT * FROM governor_withheld_now`." % wi)
    hb = g.get("heartbeat_s")
    if hb is not None and hb > 600:
        lines.append("⚠ governor heartbeat is %ss old — the state task may not be running; the level may be STALE." % hb)
    if g.get("last_change"):
        lines.append("Last level change: " + g["last_change"])
    lines.append("Runbook: docs/agent_docs/docs024_key_docs_latest/dispatch_throughput/RUNBOOK_dispatch_throughput.md §\"Spend governor\".")
    return "\n".join(lines)


def main():
    try:
        sys.stdin.read()          # the hook payload; unused
    except Exception:
        pass
    text = banner(read_governor())
    if text:
        print(json.dumps({"hookSpecificOutput": {"hookEventName": "SessionStart",
                                                 "additionalContext": text}}))
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:
        sys.exit(0)
