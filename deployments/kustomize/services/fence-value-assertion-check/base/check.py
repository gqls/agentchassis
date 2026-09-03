#!/usr/bin/env python3
"""audit-fence-value-assertions.py — which tool acceptance fences never assert a NUMBER?

WHY THIS EXISTS
---------------
`bugs_open/449`. The verification ladder implements one check type that compares a
VALUE (`computed_values`), and neither agent that authors fences has ever been told it
exists — the `tool-generator` prompt enumerates a closed vocabulary and ends "No other
check type exists for interactions". So a generated fence checks that the page loads,
logs no console errors, fits a phone, and produces *an element* after a click. **A
calculator that prints a confidently wrong figure passes every check in its fence**, and
its record reads PASSED.

Measured 2026-09-03: `tool-generator` had written **186** current fences, **116**
asserting no expected value of any kind, **0** using `computed_values`, and **55** that
FILL the tool's inputs and then check nothing about what came out.

WHY THE WINDOW IS THE HEADLINE AND THE TOTAL IS NOT
---------------------------------------------------
449 §6: "Compare by created_at window, not by total — the existing ones do not change
themselves." A fix to the AUTHORING side shows up as a fall in NEW fences and leaves the
total nearly flat for weeks, because ~116 blind fences sit there until something rewrites
them. A totals-only report would therefore read as "no improvement" for a month after a
working fix, and would be quietly abandoned as useless. So the window is the headline and
the totals are context.

WHY A CRONJOB AND NOT A COMMIT HOOK
------------------------------------
The thing this watches changes by a route no commit can carry: a fence is written into
`doc_plans` by an agent at build time, many times a day, with no commit in either
direction. At commit time there is nothing to look at.

THE DEMAND CONTROL IS BUILT IN, AND IT IS NOT DECORATION
---------------------------------------------------------
A census like this returns the same comfortable number two ways: the corpus is clean, or
the query has gone blind (a changed fence shape, a renamed column, a `LIKE` that stopped
matching). Those are the same bytes. So before reporting anything, the check requires that
SOME author still shows a non-zero `uses_computed_values` — several operator lanes carry
one on every fence — and exits 2 rather than reporting a clean corpus off a blind census.
The passing control PRINTS which author satisfied it, so a reader sees the evidence rather
than the word PASSED.

⚠ **The control is deliberately NAME-FREE, and that was a correction made within minutes
of the first run.** It originally named `operator:mortgagecalculator-lane-a4` and
`operator:bugfix224-session` explicitly, on the reasoning that a specific control is a
stronger one. It is — and `created_by` is a FREE-TEXT LANE LABEL, not an identity. The
mcalc lane re-keyed its eight fences the same morning and that author became
`operator:mortgagecalculator-lane-2026-09-03-701-rekey`; the control survived only because
the OTHER hard-coded name happened to still exist. A control that fails when a lane renames
itself does not detect a blind query, it detects a rename — and it fails CLOSED, so it
would have exited 2 and looked like a broken census on a day when nothing was wrong.
(`WRONG_CALLS.md`, 2026-09-03.)

THE TRIGGER IS READ OFF THE FENCE, NEVER FROM A CLASSIFIER
-----------------------------------------------------------
"Is this tool a calculator" needs a judgement about tool kinds and inherits that
judgement's gaps. The fence already carries the evidence: a check with a `fill` or
`select` step has DECLARED that the tool takes input. `drives_but_asserts_nothing` is
therefore a fact in the document, not an opinion about it.

⚠ RESIDUAL: THIS IS A SECOND SPELLING OF A GO RULE, AND THAT IS A REAL COST
----------------------------------------------------------------------------
`grade()` below mirrors `summariseCriteriaValueAssertions` in
`platform/orchestration/actions/criteria_value_assertions.go`. Two implementations of one
rule is the drift class this estate keeps getting bitten by, and it is not avoidable here:
the Go one runs per-fence inside the chassis at write and judge time; a fleet sweep has to
read the whole corpus offline. **The mitigation is that `--self-test`'s fixtures are the
same cases the Go test pins** (empty `expect_values`, click-only, `select`, unparseable),
so a divergence surfaces as a failing fixture on whichever side changed. Do not treat the
two as guaranteed identical — check the fixtures when you change either.

Usage:
    scripts/audit-fence-value-assertions.py                 # census, 7-day window
    scripts/audit-fence-value-assertions.py --days 1 --json
    scripts/audit-fence-value-assertions.py --self-test     # fixtures only, no cluster
    scripts/audit-fence-value-assertions.py --write-note    # one doc_notes row (the CronJob)

Exit: 0 = no NEW blind fences in the window · 1 = new blind fences found
      2 = could not determine (includes the demand control failing, and a self-test failure)
"""
import argparse
import json
import os
import re
import subprocess
import sys

# ── the grader — mirrors criteria_value_assertions.go ───────────────────────

GRADE_NONE, GRADE_PATTERN, GRADE_EXACT = "none", "pattern", "exact"

# The fence is the FIRST ```criteria block, non-greedy — mirroring
# extractCriteriaFence (check_tool_acceptance.go), which takes the first one. A greedy
# match swallows the rest of the document and every test after it silently changes
# meaning. DOTALL because a fence spans lines.
FENCE_RE = re.compile(r"```criteria(.*?)```", re.DOTALL)


def extract_fence(body):
    m = FENCE_RE.search(body or "")
    return m.group(1).strip() if m else ""


def grade(fence):
    """Return (parsed, exact, pattern, drives_inputs, asserting_ids).

    FAIL-OPEN, exactly like the Go: an unparseable or absent fence yields parsed=False and
    is never reported as "asserts nothing". A fence nobody could read has not been shown to
    assert nothing, and Tier 2 already reports that case separately as
    `criteria_unparseable` — a second, weaker report of it would be noise.
    """
    if not (fence or "").strip():
        return False, 0, 0, False, []
    try:
        doc = json.loads(fence)
    except Exception:
        return False, 0, 0, False, []
    if not isinstance(doc, dict):
        return False, 0, 0, False, []

    exact = pattern = 0
    drives = False
    exact_ids, pattern_ids = [], []
    for ch in doc.get("checks") or []:
        if not isinstance(ch, dict):
            continue
        for st in ch.get("steps") or []:
            # `click` and `reload` supply no value, so they do not make the tool an
            # input-taker; only fill and select do.
            if isinstance(st, dict) and st.get("action") in ("fill", "select"):
                drives = True
                break
        typ = ch.get("type")
        if typ == "computed_values":
            # An EMPTY expect_values is credited with NOTHING, and that is not an
            # oversight: the runner refuses it outright ("it would assert nothing and pass
            # on any page"), so counting it would credit a fence for a check that can only
            # ever fail.
            ev = ch.get("expect_values")
            if isinstance(ev, dict) and len(ev) > 0:
                exact += 1
                exact_ids.append(ch.get("id") or "?")
        elif typ == "interaction":
            exp = ch.get("expect")
            if isinstance(exp, dict) and str(exp.get("text_matches") or "").strip():
                pattern += 1
                pattern_ids.append(ch.get("id") or "?")
    return True, exact, pattern, drives, exact_ids + pattern_ids


def grade_name(exact, pattern):
    if exact > 0:
        return GRADE_EXACT
    if pattern > 0:
        return GRADE_PATTERN
    return GRADE_NONE


# ── database, two ways in ───────────────────────────────────────────────────

SQL = """
SELECT COALESCE(jsonb_agg(jsonb_build_object(
         'author', COALESCE(created_by,'(null)'),
         'created_at', to_char(created_at,'YYYY-MM-DD"T"HH24:MI:SSZ'),
         'age_days', EXTRACT(epoch FROM (now() - created_at))/86400.0,
         'subject_key', subject_key,
         'body', left(body, 60000))), '[]'::jsonb)
  FROM doc_plans
 WHERE subject_type='tool' AND is_current AND body LIKE '%```criteria%';
"""


def _psql_argv(sql):
    """Two ways in, chosen by environment — the SAME query either way.

    A session on the workstation reaches the database through `kubectl exec`; a CronJob
    inside the cluster has no pods/exec RBAC and dials postgres directly. Doing both here
    is what lets ONE file be the thing a session runs by hand and the thing the clock runs.
    Copied deliberately from audit-listing-class-promise.py rather than reinvented.
    """
    host = os.environ.get("PG_CLIENTS_HOST")
    pw = os.environ.get("CLIENTS_DB_PASSWORD")
    if host and pw:
        return (["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
                 "-At", "-v", "ON_ERROR_STOP=1", "-c", sql],
                {**os.environ, "PGPASSWORD": pw})
    return (["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
             "psql", "-U", "clients_user", "-d", "clients_db", "-At",
             "-v", "ON_ERROR_STOP=1", "-c", sql], None)


def fetch():
    argv, env = _psql_argv(SQL)
    out = subprocess.run(argv, env=env, capture_output=True, text=True, timeout=300)
    if out.returncode != 0:
        print((out.stderr or out.stdout).strip()[:2000], file=sys.stderr)
        sys.exit(2)
    body = out.stdout.strip()
    return json.loads(body) if body else []


def write_note(body):
    sql = ("INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by) "
           "VALUES ('tool','__fleet__',$note$" + body + "$note$,"
           "'[\"fence-value-assertions\"]'::jsonb,'fence-value-assertion-check','cronjob');")
    argv, env = _psql_argv(sql)
    subprocess.run(argv, env=env, check=True, capture_output=True, text=True, timeout=120)


# ── the census ──────────────────────────────────────────────────────────────

def census(rows, days):
    by_author = {}
    for r in rows:
        a = r["author"]
        parsed, exact, pattern, drives, _ = grade(extract_fence(r.get("body") or ""))
        asserts = (exact + pattern) > 0
        blind = parsed and not asserts
        fresh = float(r.get("age_days") or 1e9) <= days
        d = by_author.setdefault(a, dict(
            author=a, fences=0, asserts_no_value=0, drives_but_asserts_nothing=0,
            uses_computed_values=0, unparseable=0, created_in_window=0,
            new_blind_in_window=0, newest=""))
        d["fences"] += 1
        d["asserts_no_value"] += 1 if blind else 0
        d["drives_but_asserts_nothing"] += 1 if (blind and drives) else 0
        d["uses_computed_values"] += 1 if exact > 0 else 0
        d["unparseable"] += 0 if parsed else 1
        if fresh:
            d["created_in_window"] += 1
            d["new_blind_in_window"] += 1 if (blind and drives) else 0
        ca = (r.get("created_at") or "")[:10]
        if ca > d["newest"]:
            d["newest"] = ca
    return sorted(by_author.values(), key=lambda d: -d["fences"])


# ── self-test: the same cases the Go test pins ──────────────────────────────

LIVENESS_FENCE = json.dumps({"profiles": ["desktop", "mobile"], "checks": [
    {"id": "boots", "type": "selector_exists", "selector": "#calc"},
    {"id": "console", "type": "no_console_errors"},
    {"id": "calc", "type": "interaction",
     "steps": [{"action": "fill", "selector": "#amount", "value": "250000"},
               {"action": "click", "selector": "#go"}],
     "expect": {"selector": "#monthlyPayment"}}]})

FIXTURES = [
    # (name, fence, parsed, grade, drives, blind_and_drives)
    ("generated liveness fence that fills and asserts nothing",
     LIVENESS_FENCE, True, GRADE_NONE, True, True),
    ("the same fence with a text_matches is a PATTERN assertion",
     LIVENESS_FENCE.replace('"expect": {"selector": "#monthlyPayment"}',
                            '"expect": {"selector": "#monthlyPayment", "text_matches": "x"}'),
     True, GRADE_PATTERN, True, False),
    ("computed_values with expectations is an EXACT assertion",
     '{"checks":[{"id":"s","type":"computed_values","steps":[{"action":"fill","selector":"#a","value":"1"}],'
     '"expect_values":{"#p":"\\u00a31.00"}}]}', True, GRADE_EXACT, True, False),
    # THE MUTATION GUARD: credit computed_values by TYPE and this goes red.
    ("computed_values with an EMPTY expect_values is credited with nothing",
     '{"checks":[{"id":"s","type":"computed_values","steps":[{"action":"fill","selector":"#a","value":"1"}],'
     '"expect_values":{}}]}', True, GRADE_NONE, True, True),
    ("clicks only — asserts nothing, but does not DRIVE",
     '{"checks":[{"id":"o","type":"interaction","steps":[{"action":"click","selector":"#m"}],'
     '"expect":{"selector":"#p"}}]}', True, GRADE_NONE, False, False),
    ("select counts as driving, not just fill",
     '{"checks":[{"id":"p","type":"interaction","steps":[{"action":"select","selector":"#b","value":"f"}],'
     '"expect":{"selector":"#t"}}]}', True, GRADE_NONE, True, True),
    ("an unparseable fence is UNKNOWN, never a finding",
     '{"checks":[{"id":"x"},]}', False, GRADE_NONE, False, False),
    ("no fence at all", "   ", False, GRADE_NONE, False, False),
]


def self_test():
    failures = []
    for name, fence, want_parsed, want_grade, want_drives, want_blind_drives in FIXTURES:
        parsed, exact, pattern, drives, _ = grade(fence)
        got_grade = grade_name(exact, pattern)
        blind_drives = parsed and (exact + pattern) == 0 and drives
        if parsed != want_parsed:
            failures.append(f"{name}: parsed={parsed}, want {want_parsed}")
        if got_grade != want_grade:
            failures.append(f"{name}: grade={got_grade}, want {want_grade}")
        if drives != want_drives:
            failures.append(f"{name}: drives={drives}, want {want_drives}")
        if blind_drives != want_blind_drives:
            failures.append(f"{name}: drives_but_asserts_nothing={blind_drives}, want {want_blind_drives}")

    # The fence extractor: prose naming the fence in backticks hijacks the FIRST match
    # (LANDMINES). Pinned so a change to FENCE_RE that breaks the non-greedy or the
    # first-match rule cannot pass quietly.
    two = "```criteria\n{\"checks\":[]}\n```\nlater\n```criteria\n{\"checks\":[{\"id\":\"z\"}]}\n```"
    if extract_fence(two) != '{"checks":[]}':
        failures.append("extract_fence must take the FIRST fence, non-greedy — got %r"
                        % extract_fence(two))

    # PARITY WITH THE DEPLOYED COPY. base/check.py is a copy of this file (the
    # configMapGenerator needs a real file), so the pair is pinned here rather than
    # trusted — the drift this estate keeps getting bitten by.
    here = os.path.dirname(os.path.abspath(__file__))
    deployed = os.path.join(here, "..", "deployments", "kustomize", "services",
                            "fence-value-assertion-check", "base", "check.py")
    if os.path.exists(deployed):
        with open(os.path.abspath(__file__), "rb") as a, open(deployed, "rb") as b:
            if a.read() != b.read():
                failures.append(
                    "deployments/.../fence-value-assertion-check/base/check.py has DRIFTED from "
                    "this file — copy it across and re-apply the overlay, or the cluster keeps "
                    "the old ConfigMap while this file looks correct")
    else:
        failures.append("the deployed copy is MISSING at %s — the CronJob would run nothing"
                        % os.path.normpath(deployed))

    if failures:
        print("SELF-TEST FAILED:", file=sys.stderr)
        for f in failures:
            print("  - " + f, file=sys.stderr)
        return 2
    print("self-test: %d fixtures + extractor + deployed-copy parity — all pass" % len(FIXTURES))
    return 0


# ── main ────────────────────────────────────────────────────────────────────

def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--days", type=int, default=7, help="window for the headline (default 7)")
    ap.add_argument("--json", action="store_true")
    ap.add_argument("--self-test", action="store_true", help="fixtures only, no cluster")
    ap.add_argument("--write-note", action="store_true",
                    help="write ONE doc_notes row for this run, including a clean one")
    args = ap.parse_args()

    if args.self_test:
        sys.exit(self_test())

    rows = fetch()
    if not rows:
        print("audit-fence-value-assertions: the census returned NO ROWS. There is always at "
              "least one current tool fence, so this is a broken query, not a clean corpus.",
              file=sys.stderr)
        sys.exit(2)
    stats = census(rows, args.days)

    # ── the demand control, before any finding is reported ──────────────────
    control_author = next((d["author"] for d in stats if d["uses_computed_values"] > 0), None)
    if not control_author:
        print("audit-fence-value-assertions: DEMAND CONTROL FAILED — NO author anywhere shows a\n"
              "  computed_values fence. Several operator lanes carry one on every fence "
              "(bugs_open/449 §2),\n"
              "  so a fleet-wide zero means this query no longer sees what it is looking for.\n"
              "  Refusing to report a clean corpus off a blind census.", file=sys.stderr)
        sys.exit(2)

    new_blind = sum(d["new_blind_in_window"] for d in stats)
    total_blind = sum(d["drives_but_asserts_nothing"] for d in stats)

    if args.json:
        print(json.dumps({"window_days": args.days, "authors": stats,
                          "new_blind_in_window": new_blind,
                          "standing_drives_but_asserts_nothing": total_blind,
                          "demand_control": "passed", "control_author": control_author}))
    else:
        print("Tool acceptance fences that never assert a number — bugs_open/449")
        print("Window: last %d day(s). THE WINDOW IS THE HEADLINE; the totals are context, because"
              % args.days)
        print("the standing stock does not change itself and would read as 'no improvement'.\n")
        print("%-46s %7s %7s %9s %9s | %5s %9s  %s" %
              ("author", "fences", "blind", "drv+blind", "computed", "new", "new-blind", "newest"))
        print("-" * 118)
        for d in stats:
            print("%-46s %7d %7d %9d %9d | %5d %9d  %s" %
                  (d["author"][:46], d["fences"], d["asserts_no_value"],
                   d["drives_but_asserts_nothing"], d["uses_computed_values"],
                   d["created_in_window"], d["new_blind_in_window"], d["newest"]))
        print("\ndemand control: PASSED — %s still shows computed_values, so the census is not blind"
              % control_author)
        if new_blind:
            print("\nFINDING: %d fence(s) created in the last %d day(s) drive a tool's inputs and "
                  "assert NO value." % (new_blind, args.days))
            print("  The intake is still open. Each is a tool whose Tier-4 PASS will mean 'it "
                  "responded', not 'it is right'.")
            print("  Cause: neither fence-authoring agent knows the computed_values type (449 §3).")
        else:
            print("\nNo NEW blind fences in the window. Note this says nothing about the %d "
                  "standing ones above," % total_blind)
            print("  which are repaired per site with a per-site oracle, not by this check.")

    if args.write_note:
        # ONE row per run, INCLUDING a clean one — so a MISSING row means the job did not
        # run and can never read as "nothing is wrong". (The convention optional-key-budget
        # -check established; CLAUDE.md states it as a rule.)
        lines = ["## Fence value assertions — %d new blind fence(s) in %d day(s)"
                 % (new_blind, args.days),
                 "Observed: %d current tool fence(s) across %d author(s); %d DRIVE a tool's inputs "
                 "and assert no value of any kind; %d created in the window, %d of those blind."
                 % (sum(d["fences"] for d in stats), len(stats), total_blind,
                    sum(d["created_in_window"] for d in stats), new_blind),
                 "Root cause: neither fence-authoring agent knows the computed_values check type "
                 "(bugs_open/449 §3) — the type is absent from the prompt, so it is never a candidate.",
                 "Fix: none applied — this is a detector. The authoring fix is sequenced after "
                 "bugs_open/441 (see the 449 PLAN), because a value assertion on a stale selector "
                 "fails for the wrong reason.",
                 "Verified: demand control PASSED (%s still shows computed_values, so the census is "
                 "not blind). Detector: scripts/audit-fence-value-assertions.py (--self-test proves "
                 "the grading without cluster access)." % control_author,
                 "Categories: fence-value-assertions"]
        write_note("\n".join(lines))

    sys.exit(1 if new_blind else 0)


if __name__ == "__main__":
    main()
