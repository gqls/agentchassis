#!/usr/bin/env python3
"""check_list_empty_states.py — ADVISORY standing lint for bugs_open/054.

Flags any ACTIVE content_component whose html_template ranges over a list
({{range .items}}) with NO enclosing items guard ({{if ... items ...}}). Such a
component renders a BLANK container (no "nothing here yet" copy) when its
query-sourced list resolves empty — which happens on a freshly-built site whose
entities/games/guides/tools have not populated yet. See
docs/agent_docs/sql_for_agents/185_list_empty_state_guards.sql for the fix shape
and cta_link_integrity/RUNBOOK R16 for context.

WHY A SCRIPT AND NOT scripts/pattern-check.py: content_components live only in
the database, never in the repo, so a diff-based file linter cannot see them.
This is an OPERATIONAL check — run it by hand or from a sweep/cron against the
live cluster.

WHY THE EMPTY RENDER IS REACHABLE (the load-bearing fact, do not "fix" the schema
here instead): plan_sections_action.go:1288-1321 sets resolvedData[field]=value
and continues whenever the query result is non-nil; an empty slice is not nil, so
the required/on_missing/min_items branch never runs for a query array. items
{required:true, min_items:1} is therefore SILENTLY IGNORED and the empty list is
handed to the template. That deeper resolver gap is bugs_open/054 fix-candidate-2
and is out of scope for this lint, which only flags the missing template guard.

ADVISORY: this reports and exits non-zero on findings so a caller can decide; it
is NOT wired into the pre-commit hook (a DB round-trip has no place blocking a
shared commit — see pattern-check.py's ADVISORY note). Exit 0 = clean,
1 = unguarded component(s) found, 2 = could not reach the DB.

The guard test is the same coarse regex bugs_open/054 uses. GOTCHA: it only
proves *some* {{if ...items...}} exists in the template, not that it encloses the
range. Good enough to flag a candidate for review; confirm by reading the
template before editing (RUNBOOK R16).
"""
import argparse
import subprocess
import sys

QUERY = r"""
SELECT function,
       (html_template ~ '\{\{ *if [^}]*items *\}\}') AS has_if_guard
  FROM content_components
 WHERE is_active AND html_template LIKE '%{{range .items}}%'
 ORDER BY has_if_guard, function;
"""


def run(ns: str, pod: str, user: str, db: str) -> str:
    cmd = [
        "kubectl", "exec", "-n", ns, pod, "--",
        "psql", "-U", user, "-d", db, "-tAF|", "-c", QUERY,
    ]
    return subprocess.run(cmd, capture_output=True, text=True, timeout=60).stdout


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--namespace", default="ai-persona-system")
    ap.add_argument("--pod", default="postgres-clients-0")
    ap.add_argument("--user", default="clients_user")
    ap.add_argument("--db", default="clients_db")
    args = ap.parse_args()

    try:
        out = run(args.namespace, args.pod, args.user, args.db)
    except Exception as exc:  # noqa: BLE001 — operational tool, report and bail
        print(f"check_list_empty_states: could not reach the DB: {exc}", file=sys.stderr)
        return 2

    rows = [r for r in out.strip().splitlines() if "|" in r]
    if not rows:
        print("check_list_empty_states: no active {{range .items}} components found "
              "(unexpected — is the DB reachable and populated?)", file=sys.stderr)
        return 2

    unguarded = [r.split("|", 1)[0] for r in rows if r.rsplit("|", 1)[1] == "f"]
    total = len(rows)
    if unguarded:
        print(f"UNGUARDED list components ({len(unguarded)} of {total} active "
              f"{{{{range .items}}}} components have no items guard):")
        for fn in unguarded:
            print(f"  - {fn}")
        print("\nEach renders a blank container on an empty list. Fix per "
              "bugs_open/054 / migration 185: wrap in "
              "{{if .items}}...{{else}}<empty-state>{{end}}.")
        return 1

    print(f"OK: all {total} active {{{{range .items}}}} components carry an items guard.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
