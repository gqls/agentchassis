#!/usr/bin/env python3
"""
instance-scope-sweep — files conversion work for tools born without the per-instance
id namespace (bugs_open/283's FLOW half; owner ruling 2026-08-21).

WHY THIS EXISTS.

The 283 conversion programme fixed every component that EXISTED (the stock). The
estate keeps minting new interactive tools daily — `tool-generator` produced 23
between 2026-08-18 and 2026-08-20, seven of them on the day the backlog conversion
finished — so without a standing sweep the estate un-converts itself
(WRONG_CALLS.md 2026-08-21, the producer-census entry).

Two upstream controls exist and this sweep is the BACKSTOP behind both:
  1. the generator's prompt teaches the scoping rules (migration 520), and
  2. `create_tool_component`'s birth guard mechanically converts or refuses
     (tool_birth_instance_scope.go, armed by the same migration once a binary
     carrying it has rolled).
A prompt is a request and a guard only guards the door it stands at — templates
also arrive by hand-authored SQL, migrations, the admin UI and older actions
(deploy_tool forks). Only a clock against live `content_components` sees all of
them; same decisive argument as every sibling check.

WHAT IT DOES. Finds active, PLACED components whose html_template uses
getElementById but never references {{.InstanceID}}, and files ONE
`instance_scope_conversion` work item per row (`fix_type: scope_component_instance`,
dedup on item_key `instance-scope:<first 8 of row id>` — the exact shape the proven
batches used). The component-template-fixer does the rest: mechanical conversion,
the judged LLM branch for scripts it cannot prove, refusal to a human for the
residue, and reason-carrying rerenders/section_edits for delivery. The sweep
carries NO conversion intelligence of its own on purpose — the pipeline is the
intelligence, and it is council-approved (07635a2f round 9).

It writes ONE doc_notes row per run — when it files work AND when the estate is
clean — so a MISSING row means THE JOB DID NOT RUN, which must never look like
"nothing needed converting".

ESCALATION (owner ruling 2026-08-21, second half: "ensure all these errors are detectable
and fixable through the improvement loop"). A row that is STILL UNCONVERTED after two or more
terminal conversion items carries evidence the pipeline cannot convert it as it stands -
whether the items FAILED loudly (the chartTitle class: the converter errors) or COMPLETED as
polite no-op refusals (the dynamic-id class: "convert through the judged pool", which the
mechanical arm refuses without routing anywhere). Grinding it daily fixes nothing.
> CORRECTED 2026-08-22: the first cut matched status='failed' only and was BLIND to the
> completed-no-op class - three rows ground for a day under it (aria-builder,
> economy-flow-modeller, shadow-stacker, each 3 no-op completions). "Still unconverted after
> N terminal items" is the status-agnostic evidence: a fixed:true item converts the template
> in the same transaction, so an unconverted row's terminal items are all non-conversions. Such rows are
ESCALATED: the sweep files ONE add_tool item with replace_existing=true (the owner-ruled
rebuild route: full pipeline, fresh generation, incumbent snapshotted), spec derived from the
incumbent's own description and tool-doc header, rate-limited to one escalation per row per
14 days so a rebuild that does not cure the defect loops slowly and visibly instead of
thrashing. needs_human_review rows are DELIBERATELY NOT escalated: the judged gate parked
them precisely because no machine could prove a rewrite of a WORKING live tool safe, and a
from-scratch rebuild discards the incumbent's behaviour - that stays a human's call
(the owner made exactly that call for the 2026-08-21 six by hand).

THE DEMAND CONTROL. This check's healthy long-run answer is "0 to file", and a
broken query also returns 0. So every run counts converted templates
({{.InstanceID}} adopters — 90+ live since 2026-08-20) through the same matching in
the same statement. If THAT is zero, the matching is broken and the run REFUSES
rather than reporting a quiet result it cannot stand behind (the lesson its
predecessor, instance-token-adoption-check, wrote into its own docstring).
"""
import json
import os
import subprocess
import sys

CENSUS_SQL = """
SELECT json_build_object(
  'unconverted', (SELECT COALESCE(json_agg(json_build_object(
                     'id', x.id, 'function', x.function) ORDER BY x.function), '[]'::json)
                    FROM (SELECT DISTINCT c.id::text AS id, c.function
                            FROM content_components c
                            JOIN page_components pc ON pc.component_id = c.id
                           WHERE c.is_active
                             AND c.html_template ~ 'getElementById'
                             AND c.html_template NOT LIKE '%{{.InstanceID}}%') x),
  'converted_control', (SELECT count(*) FROM content_components
                          WHERE is_active AND html_template LIKE '%{{.InstanceID}}%'),
  'active_total',      (SELECT count(*) FROM content_components WHERE is_active),
  'open_items',        (SELECT count(*) FROM site_work_items
                          WHERE item_type = 'instance_scope_conversion'
                            AND status NOT IN ('complete','verified','rejected','wont_fix',
                                               'failed','unresolved','cancelled')),
  'parked',            (SELECT count(*) FROM site_work_items
                          WHERE item_type = 'instance_scope_conversion'
                            AND status = 'needs_human_review')
);
"""

# One item per unconverted row. The NOT EXISTS dedup means a re-run (or an item
# already filed by a session or an earlier sweep) costs nothing; rows whose items
# were REFUSED to needs_human_review are NOT re-filed (that status is
# non-terminal in the dedup list precisely so a parked row stays parked for its
# human rather than being re-ground daily).
FILE_ITEMS_SQL = """
INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary,
  priority, handler_agent, status, created_by, spec, item_key
)
SELECT DISTINCT ON (c.id)
  s.id, 'automated_check', 'build', 'instance_scope_conversion', 'medium',
  'instance-scope-sweep: ' || c.function || ' (row ' || left(c.id::text, 8)
    || ') uses getElementById without {{.InstanceID}} — born after or outside the 283 conversion; converting through the proven pipeline',
  45, 'component-template-fixer', 'triaged', 'instance-scope-sweep',
  jsonb_build_object(
    'fix_type', 'scope_component_instance',
    'component_id', c.id::text,
    'category', 'seam',
    'note', 'filed by the standing sweep (bugs_open/283 flow half); mechanical or judged per the fixer''s own routing'),
  'instance-scope:' || left(c.id::text, 8)
FROM content_components c
JOIN page_components pc ON pc.component_id = c.id
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE c.is_active
  AND c.html_template ~ 'getElementById'
  AND c.html_template NOT LIKE '%{{.InstanceID}}%'
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items w
    WHERE w.item_key = 'instance-scope:' || left(c.id::text, 8)
      AND w.status NOT IN ('complete','verified','rejected','wont_fix',
                           'failed','unresolved','cancelled'))
RETURNING item_key;
"""


# Escalate rows with >=2 FAILED conversion items and no add_tool rebuild filed in 14 days.
# Spec derived from the incumbent (its display_name, description, tool-doc header) - the
# framework's own words, never composed here.
ESCALATE_SQL = """
WITH failed_rows AS (
  SELECT c.id, c.function, c.display_name, c.description, c.html_template,
         s.domain, s.id AS site_id,
         count(w.id) AS failed_n
  FROM content_components c
  JOIN page_components pc ON pc.component_id = c.id
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
  JOIN site_work_items w ON w.item_key = 'instance-scope:' || left(c.id::text, 8)
                        AND w.status IN ('failed', 'complete')
  WHERE c.is_active
    AND c.html_template ~ 'getElementById'
    AND c.html_template NOT LIKE '%{{.InstanceID}}%'
  GROUP BY c.id, c.function, c.display_name, c.description, c.html_template, s.domain, s.id
  HAVING count(w.id) >= 2
)
INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary,
  priority, handler_agent, status, created_by, spec, item_key
)
SELECT DISTINCT ON (f.site_id, f.function)
  f.site_id, 'automated_check', 'build', 'add_tool', 'medium',
  'instance-scope-sweep ESCALATION: ' || f.function || ' is still unconverted after ' || f.failed_n
    || ' terminal conversion items (failed or refused-as-no-op); rebuilding through the full pipeline per the 2026-08-21 owner ruling',
  40, 'tool-generator', 'triaged', 'instance-scope-sweep',
  jsonb_build_object(
    'name', f.display_name,
    'function', f.function,
    'complexity', 'moderate',
    'replace_existing', true,
    'description',
      COALESCE(NULLIF(f.description,''), f.display_name)
      || E'\n\nSTRUCTURAL REQUIREMENT (the save gate enforces this and refuses the tool otherwise): every document.getElementById call must take a quoted string literal written directly at the call site. Never iterate over arrays of element-id strings, never pass an id through a variable or parameter, and never build an id by concatenation. If the tool builds repeated rows dynamically, give each row''s elements ids with a static hyphen-ended prefix (for example ''row-'' + n) and look them up the same composed way consistently - or better, keep references to created elements in variables instead of looking them up by id at all.'
      || CASE WHEN position('=== tool-doc ===' in f.html_template) > 0
               AND position('=== /tool-doc ===' in f.html_template) > position('=== tool-doc ===' in f.html_template)
         THEN E'\n\nBehaviour contract of the incumbent (from its own tool-doc header; preserve these invariants):\n'
              || substr(f.html_template,
                        position('=== tool-doc ===' in f.html_template) + 16,
                        position('=== /tool-doc ===' in f.html_template) - position('=== tool-doc ===' in f.html_template) - 16)
         ELSE '' END),
  'add_tool_rebuild_' || f.function || '_' || f.domain
FROM failed_rows f
WHERE NOT EXISTS (
  SELECT 1 FROM site_work_items w2
  WHERE w2.item_key = 'add_tool_rebuild_' || f.function || '_' || f.domain
    AND (w2.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled')
         OR w2.created_at > now() - interval '14 days'))
RETURNING item_key;
"""


def psql(sql, password, host, capture=True):
    env = dict(os.environ)
    env["PGPASSWORD"] = password
    out = subprocess.run(
        ["psql", "-h", host, "-p", "5432", "-U", "clients_user", "-d", "clients_db",
         "-t", "-A", "-v", "ON_ERROR_STOP=1", "-c", sql],
        env=env, check=True, capture_output=True, text=True,
    )
    return out.stdout.strip()


def render_report(c, filed):
    lines = [
        "instance-scope-sweep — the standing backstop for bugs_open/283's flow half.",
        "",
        f"active components:                  {c['active_total']}",
        f"converted ({{{{.InstanceID}}}}, control): {c['converted_control']}",
        f"unconverted + placed (the corpus):  {len(c['unconverted'])}",
        f"conversion items open before run:   {c['open_items']}",
        f"parked at needs_human_review:       {c['parked']}",
        f"items FILED this run:               {len(filed)}",
        "",
    ]
    if not c["unconverted"]:
        lines += [
            "CLEAN — every placed interactive component carries the per-instance namespace.",
            "This row exists on a clean run ON PURPOSE: a MISSING row means the sweep did",
            "not run, which is not the same thing and must not look alike.",
        ]
    elif not filed:
        lines += [
            "Nothing filed: every unconverted row already has an open conversion item",
            "(in flight, or parked for a human). The corpus list follows so a reader can",
            "route the parked ones without re-deriving it:",
        ] + [f"    {r['function']} ({r['id'][:8]})" for r in c["unconverted"][:40]]
    else:
        lines += ["Filed for conversion through the proven pipeline:"] + [
            f"    {k}" for k in filed[:40]]
    return "\n".join(lines)


def write_doc_note(body, password, host):
    tag = "issweep"
    sql = (
        "INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
        f"VALUES ('pipeline', 'instance-scope-sweep', ${tag}${body}${tag}$, "
        "'[\"instance-scope-sweep\",\"bugs-open-283\"]'::jsonb, "
        "'instance-scope-sweep');"
    )
    path = "/tmp/instance-scope-sweep-note.sql"
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
    # --stdin renders a report from a census on stdin so both branches can be
    # exercised without a database (the exercise-both-branches rule).
    if "--stdin" in sys.argv:
        c = json.load(sys.stdin)
        print(render_report(c, c.get("simulated_filed", [])))
        sys.exit(0)

    password = os.environ.get("CLIENTS_DB_PASSWORD")
    if not password:
        print("REFUSING TO RUN: CLIENTS_DB_PASSWORD is not set.", file=sys.stderr)
        sys.exit(2)
    host = os.environ.get("PG_CLIENTS_HOST", "postgres-clients")

    raw = psql(CENSUS_SQL, password, host)
    if not raw:
        print("REFUSING TO RUN: the census returned nothing.", file=sys.stderr)
        sys.exit(2)
    c = json.loads(raw)

    if c["active_total"] == 0:
        print("REFUSING TO RUN: 0 active components — the query failed or the library is "
              "empty; refusing to report over it.", file=sys.stderr)
        sys.exit(2)

    # THE DEMAND CONTROL: converted templates are matched by the same LIKE through
    # the same escaping in the same statement, and are known non-zero (90+ live
    # since 2026-08-20). Zero here means the MATCHING broke, not the estate.
    if c["converted_control"] == 0:
        print("REFUSING TO RUN: the {{.InstanceID}} converted-count control matched 0 "
              "templates, against 90+ known live adopters. The pattern matching is "
              "broken; every number this run would report is worthless. Do NOT 'fix' "
              "this by removing the control.", file=sys.stderr)
        sys.exit(2)

    filed = []
    if c["unconverted"]:
        out = psql(FILE_ITEMS_SQL, password, host)
        filed = [ln for ln in out.splitlines() if ln.strip() and not ln.startswith("INSERT")]

    out = psql(ESCALATE_SQL, password, host)
    escalated = [ln for ln in out.splitlines() if ln.strip() and not ln.startswith("INSERT")]

    report = render_report(c, filed)
    if escalated:
        report += ("\n\nESCALATED to full-pipeline rebuild (>=2 failed conversions, "
                   "14-day rate limit; owner ruling 2026-08-21):\n"
                   + "\n".join("    " + k for k in escalated))
    print(report)
    write_doc_note(report, password, host)
    sys.exit(0)


if __name__ == "__main__":
    main()
