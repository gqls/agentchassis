#!/usr/bin/env python3
"""update_component.py — update an already-loaded component's template + schema.

WHY THIS EXISTS SEPARATELY FROM load_components.py, WHICH DELIBERATELY HAS NO
--force. That refusal guards a real hazard: overwriting a row another lane
inserted, with no diff and no warning, recoverable only by restore. It is right,
and it is not the situation here.

The situation here is that a component ALREADY LOADED needs a genuine change.
`tool-loan-repayment` gained `{{if}}` guards around three optional copy fields so
that one component can serve both /tools/standard-calc.html and /index.html —
without them, blanking a field renders an empty styled box rather than nothing.

AND THE STALE ROW IS NOT COSMETIC. page_components.rendered_html is what ships,
so a page written from the new file looks right today either way. But
content_components.html_template is what a LATER re-render reads — the writer
loop, the tool-improver, any rebuild. Leave it stale and the next re-render of
the homepage silently resurrects the unguarded markup and puts two dated
market-rate claims back on a page the owner was told would not carry them. The
divergence would be invisible until it shipped.

WHAT IT REFUSES:
  - a function not named on the command line (no bulk update);
  - a function with no existing row (use load_components.py — inserting is a
    different decision from updating);
  - anything failing the same five validations load_components.py runs, re-used
    by import rather than restated, because two copies of a validation rule is
    the drift this repo reviews for.

It ALWAYS backs the old row up first, into content_components_bak_20260802_decomp.

Usage:  python3 update_component.py --check tool-loan-repayment
        python3 update_component.py --apply tool-loan-repayment
"""
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.dirname(HERE)
REWRITE = os.path.join(LANE, "rewrite")
sys.path.insert(0, REWRITE)

from load_components import TOOLS, dollar_tag, psql, psql_stdin, validate  # noqa: E402

BAK = "content_components_bak_20260802_decomp"


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

        got, err = validate(fn)
        if err:
            print("INVALID  %-28s %s" % (fn, err))
            return 1

        existing = psql("SELECT id, md5(html_template), octet_length(html_template) "
                        "FROM content_components WHERE function='%s';" % fn)
        if not existing.strip():
            print("REFUSE   %s has no row — inserting is load_components.py's job" % fn)
            return 1
        cid, old_md5, old_len = existing.split("|")

        import hashlib
        new_md5 = hashlib.md5(got["template"].encode("utf-8")).hexdigest()
        print("%-28s stored %s (%s b)\n%-28s file   %s (%d b)"
              % (fn, old_md5[:12], old_len, "", new_md5[:12], len(got["template"].encode())))
        if old_md5 == new_md5:
            print("   identical — nothing to do")
            continue
        if not apply:
            print("   --check: would update %s" % cid)
            continue

        tmpl = got["template"]
        schema_json = json.dumps(got["schema"], ensure_ascii=False)
        t = dollar_tag(tmpl, schema_json)
        psql_stdin("\n".join([
            "BEGIN;",
            "CREATE TABLE IF NOT EXISTS %s (LIKE content_components INCLUDING ALL);" % BAK,
            "INSERT INTO %s SELECT * FROM content_components WHERE id='%s' "
            "AND NOT EXISTS (SELECT 1 FROM %s WHERE id='%s');" % (BAK, cid, BAK, cid),
            "UPDATE content_components SET html_template={t}{h}{t}, "
            "input_schema={t}{s}{t}::jsonb, updated_at=now() WHERE id='{i}';".format(
                t=t, h=tmpl, s=schema_json, i=cid),
            "COMMIT;",
        ]))
        print("   updated; previous row kept in %s" % BAK)
        print(psql("SELECT function, md5(html_template), octet_length(html_template), "
                   "updated_at FROM content_components WHERE id='%s';" % cid))
    return 0


if __name__ == "__main__":
    sys.exit(main())
