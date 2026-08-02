#!/usr/bin/env python3
"""backfill_content_data.py — add a component's NEW schema fields to the
page_components rows that already render it.

THE GAP THIS CLOSES, WHICH IS SILENT AND SHIPS.

`content_components.input_schema` carries a `fallback` for every field, and it is
tempting to assume the renderer consults it. IT DOES NOT. The render context in
`rerender_page_sections_action.go` is built as:

    base ⊕ page_components.content_data ⊕ plan.resolved_data

and the schema is not in that list at all. `RenderTemplate` resolves a field the
context has no key for to the EMPTY STRING and logs a Warn — it does not fail, it
does not fall back, and it does not mark the page.

So adding a required field to a schema and a `{{.field}}` to a template is only
two thirds of a change. Until the key is in the row's `content_data`, the next
re-render silently drops it. On 2026-08-03 that would have shipped:

  - `tool-loan-vs-savings` with an EMPTY accessibility badge — the entire fix,
    rendering to nothing, on a page whose acceptance check would still pass
    because the badge element is present and the numbers are unchanged;
  - `tool-consolidation-risk` with an invisible withheld-comparison notice and
    two empty inline colours.

Both would have looked like a successful deploy. This is the same shape as the
lane's other traps: the artefact is served, the status is `complete`, and the
thing you changed is not there.

WHAT IT DOES. For each named component: read the lane's schema file, find every
field the schema declares that the row's `content_data` does not carry, and add
it with the schema's own `fallback`. It NEVER overwrites a key that is already
present — a field whose live value was edited by the writer loop or by a human
must win over a fallback, and this tool has no way to tell an intentional edit
from a stale one.

It backs the rows up into page_components_bak_20260803_backfill first.

Usage:  python3 backfill_content_data.py --check tool-loan-vs-savings ...
        python3 backfill_content_data.py --apply tool-loan-vs-savings ...
"""
import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.dirname(HERE)
sys.path.insert(0, os.path.join(LANE, "rewrite"))

from load_components import TOOLS, dollar_tag, psql, psql_stdin  # noqa: E402

SITE_ID = "0162cde4-633e-45e9-8ca6-87a6b2fe1d26"
BAK = "page_components_bak_20260803_backfill"


def schema_fields(fn):
    path = os.path.join(LANE, "rewrite", fn + ".schema.json")
    with open(path, encoding="utf-8") as fh:
        return json.load(fh).get("fields", {})


def rows_for(fn):
    """Every page_components row on this site rendering this component."""
    out = psql(
        "SELECT pc.id, p.name, COALESCE(pc.content_data::text,'{}') "
        "FROM page_components pc "
        "JOIN pages p ON p.id = pc.page_id "
        "JOIN content_components cc ON cc.id = pc.component_id "
        "WHERE p.site_id = '%s' AND cc.function = '%s' ORDER BY p.name;" % (SITE_ID, fn))
    rows = []
    for line in out.splitlines():
        if not line.strip():
            continue
        rid, name, data = line.split("|", 2)
        rows.append((rid, name, json.loads(data)))
    return rows


def main():
    apply = "--apply" in sys.argv
    names = [a for a in sys.argv[1:] if not a.startswith("--")]
    if not names or (not apply and "--check" not in sys.argv):
        print(__doc__)
        return 2

    for fn in names:
        if fn not in TOOLS:
            print("REFUSE   %s is not one of this lane's components" % fn)
            return 1

    total_missing = 0
    for fn in names:
        fields = schema_fields(fn)
        rows = rows_for(fn)
        if not rows:
            print("REFUSE   %s renders on no page of this site" % fn)
            return 1

        for rid, page, data in rows:
            missing = {k: v.get("fallback") for k, v in fields.items()
                       if k not in data}
            # A schema field with no fallback cannot be backfilled and must not
            # be guessed at — report it and refuse rather than write a null.
            noval = sorted(k for k, v in missing.items() if v is None)
            if noval:
                print("REFUSE   %s/%s: schema field(s) with no fallback: %s"
                      % (fn, page, ", ".join(noval)))
                return 1

            print("%-26s %-28s %d field(s) missing%s"
                  % (fn, page, len(missing),
                     ": " + ", ".join(sorted(missing)) if missing else ""))
            if not missing:
                continue
            total_missing += len(missing)
            if not apply:
                continue

            patch = json.dumps(missing, ensure_ascii=False)
            t = dollar_tag(patch)
            psql_stdin("\n".join([
                "BEGIN;",
                "CREATE TABLE IF NOT EXISTS %s (LIKE page_components INCLUDING ALL);" % BAK,
                "INSERT INTO %s SELECT * FROM page_components WHERE id='%s' "
                "AND NOT EXISTS (SELECT 1 FROM %s WHERE id='%s');" % (BAK, rid, BAK, rid),
                # `stored || patch` would let the patch WIN on a collision. The
                # operands are the other way round on purpose: anything already
                # in content_data survives, and this can only ADD.
                "UPDATE page_components SET content_data = "
                "  {t}{p}{t}::jsonb || COALESCE(content_data, '{{}}'::jsonb) "
                "WHERE id='{i}';".format(t=t, p=patch, i=rid),
                "COMMIT;",
            ]))
            print("   added; previous row kept in %s" % BAK)

    if not apply:
        print("\n--check: %d field(s) would be added across %d component(s)"
              % (total_missing, len(names)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
