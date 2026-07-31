#!/usr/bin/env python3
"""load_components.py — install the eleven proven components into content_components.

SAFE BY ORDERING, and that is the design. A component with no `page_components`
row is inert: nothing renders it, nothing serves it, and loancalculator.co.uk
keeps serving its verbatim pages byte-for-byte. So this step can be verified at
leisure before anything is attached to a page. Attaching them is a separate,
reversible step that must be taken deliberately.

WHAT IT REFUSES TO DO:
  - overwrite an existing row. If `function` is already present the tool is
    skipped and named. There is no --force: a clobber here would destroy another
    lane's component with no diff and no warning, and the recovery is a restore.
  - install anything that fails validation (below). Validation runs over ALL
    eleven and the transaction is only opened if every one passes, so a bad
    template cannot leave a half-loaded set behind.

WHAT IT VALIDATES, and why each check exists rather than being assumed:
  1. The Go template PARSES. A parse failure makes the renderer fall back
     silently to a regex engine (LANDMINE), so the component would appear to work
     and drift from what text/template would produce. render_tool.go has already
     parsed each of these, but it parsed the file on disk — this parses what is
     about to be STORED, which is the thing that matters.
  2. The tool-doc header is WELL FORMED (both sentinels, opener first).
     `StripToolDocHeader` deliberately leaves a malformed block alone rather than
     risk truncating a script, so an unterminated opener SHIPS to the public page.
  3. No `<no value>` and no residual `{{` after rendering with the fallbacks —
     the TL-030 corruption class.
  4. Structural tag balance on script/style/section/div/fieldset, the same
     five-pair predicate the birth-write guard uses.
  5. The input_schema is `{fields: {...}}` with every field carrying
     type/source/fallback/llm_guidance, matching the live convention read off
     tool-mortgage-overpayment.

Usage:  python3 load_components.py --check     (validate only, no DB writes)
        python3 load_components.py --apply
"""
import json
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db"]

OPEN, CLOSE = "/* === tool-doc ===", "=== /tool-doc === */"
BALANCED = [("<script", "</script>"), ("<style", "</style>"), ("<section", "</section>"),
            ("<div", "</div>"), ("<fieldset", "</fieldset>")]

# function -> (name, category). `name` is NOT NULL in the schema.
TOOLS = {
    "tool-loan-repayment":        ("Standard Loan Repayment Calculator", "finance"),
    "tool-credit-health-check":   ("Credit Health Check", "finance"),
    "tool-rate-stress-test":      ("Variable Rate Stress Test", "finance"),
    "tool-early-settlement":      ("Early Settlement Estimator", "finance"),
    "tool-overpayment-impact":    ("Overpayment Impact Tool", "finance"),
    "tool-loan-vs-savings":       ("Pay Off Loan or Save?", "finance"),
    "tool-return-damage-checker": ("Car Return Damage Checker", "finance"),
    "tool-compare-loan-offers":   ("Compare Loan Offers", "finance"),
    "tool-car-finance-pcp-hp":    ("Car Finance: PCP vs HP", "finance"),
    "tool-consolidation-risk":    ("Debt Consolidation Risk Checker", "finance"),
    "tool-application-tracker":   ("Loan Application Tracker", "finance"),
}


def psql(sql):
    r = subprocess.run(PSQL + ["-tA", "-v", "ON_ERROR_STOP=1", "-c", sql],
                       capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError((r.stderr or r.stdout).strip()[:400])
    return r.stdout.strip()


def psql_stdin(sql):
    r = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1"], input=sql,
                       capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError((r.stderr or r.stdout).strip()[:600])
    return r.stdout.strip()


def render(fn):
    """Render with the REAL engine, and return the rendered HTML."""
    out = os.path.join("/tmp", fn + ".load.html")
    r = subprocess.run(["go", "run", os.path.join(HERE, "render_tool.go"),
                        os.path.join(HERE, fn + ".html.tmpl"),
                        os.path.join(HERE, fn + ".schema.json"), out],
                       capture_output=True, text=True, cwd=HERE)
    if r.returncode != 0:
        return None, (r.stderr or r.stdout).strip()[:300]
    return open(out, encoding="utf-8").read(), None


def validate(fn):
    tmpl_path = os.path.join(HERE, fn + ".html.tmpl")
    schema_path = os.path.join(HERE, fn + ".schema.json")
    for p in (tmpl_path, schema_path):
        if not os.path.exists(p):
            return None, "missing " + os.path.basename(p)
    tmpl = open(tmpl_path, encoding="utf-8").read()
    schema = json.load(open(schema_path, encoding="utf-8"))

    # 2. tool-doc header well formed
    o, c = tmpl.find(OPEN), tmpl.find(CLOSE)
    if o < 0 or c < 0 or c < o:
        return None, "tool-doc header missing or malformed (opener=%d closer=%d)" % (o, c)

    # 5. schema shape
    fields = schema.get("fields")
    if not isinstance(fields, dict) or not fields:
        return None, "input_schema has no non-empty `fields` object"
    for name, f in fields.items():
        for key in ("type", "source", "fallback", "llm_guidance"):
            if key not in f:
                return None, "field %s missing %s" % (name, key)

    # 1 + 3. real engine parse + render (render_tool.go also enforces the
    # no-copy-inside-<script> rule and the <no value>/{{ residue checks)
    html, err = render(fn)
    if err:
        return None, "render/parse failed: " + err

    # 4. tag balance on the STORED artefact
    low = html.lower()
    for op, cl in BALANCED:
        if low.count(op) > low.count(cl):
            return None, "unbalanced %s (%d open, %d close)" % (op, low.count(op), low.count(cl))

    return {"template": tmpl, "schema": schema, "rendered": len(html)}, None


def dollar_tag(*bodies):
    """A dollar-quote tag that appears in none of the bodies."""
    for i in range(1000):
        tag = "$ld%d$" % i
        if all(tag not in b for b in bodies):
            return tag
    raise RuntimeError("could not find a free dollar-quote tag")


def main():
    apply = "--apply" in sys.argv
    if not apply and "--check" not in sys.argv:
        print(__doc__)
        return 2

    print("== validating %d component(s) ==" % len(TOOLS))
    loaded, bad = {}, []
    for fn in sorted(TOOLS):
        got, err = validate(fn)
        if err:
            bad.append((fn, err))
            print("  INVALID  %-28s %s" % (fn, err))
        else:
            loaded[fn] = got
            print("  ok       %-28s %d fields, renders %d bytes"
                  % (fn, len(got["schema"]["fields"]), got["rendered"]))
    if bad:
        print("\n%d invalid — nothing written." % len(bad))
        return 1

    existing = set(x for x in psql(
        "SELECT function FROM content_components WHERE function IN (%s);"
        % ",".join("'%s'" % f for f in sorted(TOOLS))).splitlines() if x)
    todo = [f for f in sorted(TOOLS) if f not in existing]
    for f in sorted(existing):
        print("  SKIP     %-28s already present — not overwritten" % f)
    if not todo:
        print("\nnothing to insert.")
        return 0

    if not apply:
        print("\n--check: would insert %d component(s): %s" % (len(todo), ", ".join(todo)))
        return 0

    stmts = ["BEGIN;"]
    for fn in todo:
        name, category = TOOLS[fn]
        tmpl = loaded[fn]["template"]
        schema_json = json.dumps(loaded[fn]["schema"], ensure_ascii=False)
        t = dollar_tag(tmpl, schema_json, name)
        # NOT ON CONFLICT DO UPDATE: the pre-check above is the guard, and an
        # upsert here would quietly clobber a row another lane inserted between
        # the check and the write. A conflict must be a loud failure.
        stmts.append(
            "INSERT INTO content_components "
            "(function, name, display_name, description, category, component_level, "
            " render_mode, created_from, is_active, html_template, input_schema) VALUES ("
            "{t}{fn}{t}, {t}{name}{t}, {t}{name}{t}, "
            "{t}Rewritten from the hand-built loancalculator.co.uk original 2026-07-31; "
            "proven numerically identical across three input vectors.{t}, "
            "{t}{cat}{t}, 'tool', 'template', 'manual', true, "
            "{t}{tmpl}{t}, {t}{schema}{t}::jsonb);".format(
                t=t, fn=fn, name=name, cat=category, tmpl=tmpl, schema=schema_json))
    stmts.append("COMMIT;")

    print("\n== inserting %d component(s) in ONE transaction ==" % len(todo))
    psql_stdin("\n".join(stmts))

    # Verify by reading back, not by trusting the exit code.
    rows = psql(
        "SELECT function, component_level, is_active, length(html_template), "
        "(SELECT count(*) FROM jsonb_object_keys(input_schema->'fields')) "
        "FROM content_components WHERE function IN (%s) ORDER BY function;"
        % ",".join("'%s'" % f for f in todo))
    print(rows)
    n = len([l for l in rows.splitlines() if l.strip()])
    print("\n%d of %d component(s) read back from the database." % (n, len(todo)))
    print("They are INERT until a page_components row points at one — the site still")
    print("serves its verbatim pages, unchanged.")
    return 0 if n == len(todo) else 1


if __name__ == "__main__":
    sys.exit(main())
