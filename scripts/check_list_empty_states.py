#!/usr/bin/env python3
"""check_list_empty_states.py — ADVISORY standing lint for bugs_closed/054.

Flags any ACTIVE content_component that ranges over a list with NO conditional
naming that list. Such a component renders a BLANK container (no "nothing here
yet" copy) when its query-sourced list resolves empty — which happens on a
freshly-built site whose entities/games/guides/tools have not populated yet. See
docs/agent_docs/sql_for_agents/185_list_empty_state_guards.sql for the fix shape
and cta_link_integrity/RUNBOOK R16 for context.

> **WIDENED 2026-09-02 (bugs_open/425). It was looking at 8 of 50 components.**
> The predicate was `html_template LIKE '%{{range .items}}%'` — a LITERAL. The
> component library ranges over whatever the schema field is called, and
> [MEASURED 2026-09-02] `.items` is the fourth most common spelling: `.entries`
> 13, `.items` 8, `.cards` / `.features` / `.products` 3 each, then a long tail
> of `.articles`, `.categories`, `.testimonials`, `.periods`, `.rows` … So a
> component was checked or not according to what its author named a variable,
> and `content-listing` — the component behind bugs_open/425 — had never been
> looked at once. On the day of the widening the check reported **1 unguarded
> component of the 8 it could see**; it then reported **29 unguarded {{range}}
> blocks of 72, across 55 active components** `[MEASURED 2026-09-02]`.
>
> ⚠ **THAT FIGURE IS DECORATIVE AND IT DRIFTS — re-run rather than quote it.** By
> 2026-09-03 it was **30 of 74 across 57**, one day later, purely because the
> library grew. Nothing in this file depends on the number, which is exactly why
> nobody would notice it going stale: *a count no inference loads is never tested
> by the argument carrying it.* The command is the answer, not the figure:
> `python3 scripts/check_list_empty_states.py`. Nothing regressed. The other 28
> were always there and the check could not see them. (The denominator changed
> too: it now counts every {{range}} BLOCK, because a component may carry more
> than one and the old per-component reading hid the second.)
>
> This is the failure mode this estate keeps meeting from the reassuring side: a
> narrowed corpus reports a clean-looking number, and the number is honest about
> the corpus and silent about the gap. The fix is that the collection name is
> now DERIVED from each template rather than assumed.

WHY A SCRIPT AND NOT scripts/pattern-check.py: content_components live only in
the database, never in the repo, so a diff-based file linter cannot see them.
This is an OPERATIONAL check — run it by hand or from a sweep/cron against the
live cluster.

WHY THE EMPTY RENDER IS REACHABLE (the load-bearing fact, do not "fix" the schema
here instead): plan_sections_action.go sets resolvedData[field]=value and
continues whenever the query result is non-nil; an empty slice is not nil, so the
required/on_missing/min_items branch never runs for a query array. That deeper
resolver gap was bugs_closed/054 fix-candidate-2, fixed and live on chassis
v1.0.1149 for `on_missing: skip_section`; this lint covers the template half.

SIBLING CHECK: scripts/check_card_slot_guards.py asks the other half of the same
question — this one asks whether the ARRAY can be empty, that one whether an
ITEM'S FIELD can be. A component can pass here and fail there.

ADVISORY: this reports and exits non-zero on findings so a caller can decide; it
is NOT wired into the pre-commit hook (a DB round-trip has no place blocking a
shared commit — see pattern-check.py's ADVISORY note). Exit 0 = clean,
1 = unguarded component(s) found, 2 = could not reach the DB.

The guard test is the same coarse regex bugs_closed/054 uses. GOTCHA: it only
proves *some* {{if …<collection>…}} exists in the template, not that it encloses
the range. Good enough to flag a candidate for review; confirm by reading the
template before editing (RUNBOOK R16).
"""
import argparse
import sys

sys.path.insert(0, __file__.rsplit("/", 1)[0])
from component_template_lib import (  # noqa: E402
    DBUnreachable, RANGE_START, active_templates, mentions_in_condition,
)


def unguarded_ranges(name, template):
    """(collection, guarded?) for every {{range .X}} in the template."""
    out = []
    for m in RANGE_START.finditer(template):
        coll = m.group(1)
        out.append((coll, mentions_in_condition(template, coll)))
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    # Kept for callers that pass them (RUNBOOK R16); the connection details now
    # live in component_template_lib and these are accepted, not required.
    ap.add_argument("--namespace", default="ai-persona-system")
    ap.add_argument("--pod", default="postgres-clients-0")
    ap.add_argument("--user", default="clients_user")
    ap.add_argument("--db", default="clients_db")
    args = ap.parse_args()

    try:
        rows = active_templates()
    except DBUnreachable as exc:
        print(f"check_list_empty_states: could not reach the DB: {exc}", file=sys.stderr)
        return 2
    if not rows:
        print("check_list_empty_states: no active {{range}} components found "
              "(the query SUCCEEDED, so the library is genuinely empty — this is "
              "not a connection failure)", file=sys.stderr)
        return 2

    findings, total = [], 0
    for name, tmpl in rows:
        for coll, guarded in unguarded_ranges(name, tmpl):
            total += 1
            if not guarded:
                findings.append((name, coll))

    if findings:
        print(f"UNGUARDED list components ({len(findings)} of {total} "
              f"{{{{range}}}} blocks across {len(rows)} active components have no "
              f"guard on the collection they range over):")
        for name, coll in sorted(findings):
            print(f"  - {name}  (ranges .{coll})")
        print("\nEach renders a blank container on an empty list. Fix per "
              "bugs_closed/054 / migration 185: wrap in "
              "{{if .X}}...{{else}}<empty-state>{{end}}.")
        return 1

    print(f"OK: all {total} {{{{range}}}} blocks across {len(rows)} active "
          f"components carry a guard on their collection.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
