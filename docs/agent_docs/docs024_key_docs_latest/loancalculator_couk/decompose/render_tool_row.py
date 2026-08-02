#!/usr/bin/env python3
"""render_tool_row.py — re-render ONE tool row's rendered_html offline.

WHY THIS EXISTS, AND IT IS NOT A CONVENIENCE.

**The platform cannot re-render any section of this site.** `rerender_sections`
resolves each section's component by passing `page_components.slot_name` to
`loadComponentSchemas`, which looks components up **by name or function**. This
site's slots are POSITIONAL — `prose-0`, `prose-1`, `tool-2` — so nothing ever
matches, every section takes the `component not found, carrying stored HTML`
branch, and the work item completes reporting success having changed nothing.

Measured 2026-08-03, work item `b0c2265d`, orchestration `439489b6`:

    rerender_sections -> rerendered: 0, carried: 4, skipped: false

The four calculator fixes were already live in `content_components` at the time.
The page came back byte-for-byte as before (±1 byte from the save round-trip) and
the item said `complete`. **Nothing in that outcome distinguishes it from a
successful render.** See `bugs_open/` for the case.

So the route this site actually has — and the one its 27 pages were shipped
through — is: render the component offline with the SAME Go template engine, write
`rendered_html` directly, then let `render_page` (assemble-only, no `spec.reason`)
stitch the stored bytes. That path is proven: 27 of 27 pages byte-identical to the
mirror's predictions.

WHAT MAKES THIS SAFE RATHER THAN A HAND-EDIT. It renders from
`page_components.content_data` — the row's OWN values, not the schema fallbacks —
so what ships is what the row says it is. And `--check` runs a **CONTROL** first:
it re-renders the row from the template as committed at a baseline ref and
requires that to reproduce the bytes currently stored. If the control fails, the
offline renderer and whatever produced the live row disagree, and nothing should
be written until that is understood.

Usage:
  python3 render_tool_row.py --check <function> [...]
  python3 render_tool_row.py --apply <function> [...]
  python3 render_tool_row.py --check --control-ref <sha> <function>
"""
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.dirname(HERE)
REWRITE = os.path.join(LANE, "rewrite")
sys.path.insert(0, REWRITE)
sys.path.insert(0, HERE)

from load_components import TOOLS, dollar_tag, psql, psql_stdin  # noqa: E402
from assemble_mirror import render_component                     # noqa: E402

REPO = os.path.abspath(os.path.join(LANE, "..", "..", "..", ".."))
SITE_ID = "0162cde4-633e-45e9-8ca6-87a6b2fe1d26"
BAK = "page_components_bak_20260803_toolfix"
DEFAULT_CONTROL_REF = "6e8098022"   # last commit before the 2026-08-03 fixes


def live_rows(fn):
    out = psql(
        "SELECT pc.id, p.name, COALESCE(pc.content_data::text,'{}'), "
        "       octet_length(pc.rendered_html), md5(pc.rendered_html) "
        "FROM page_components pc "
        "JOIN pages p ON p.id = pc.page_id "
        "JOIN content_components cc ON cc.id = pc.component_id "
        "WHERE p.site_id = '%s' AND cc.function = '%s' ORDER BY p.name;" % (SITE_ID, fn))
    rows = []
    for line in out.splitlines():
        if not line.strip():
            continue
        rid, page, cd, nbytes, md5 = line.split("|", 4)
        rows.append((rid, page, json.loads(cd), int(nbytes), md5))
    return rows


def overrides_from(content_data, fn, rewrite_dir):
    """content_data minus anything that is not a schema field.

    `_provenance` is written by the decomposer and is deliberately NOT a
    component field — render_component raises on an unknown override, which is
    the behaviour we want everywhere except here.
    """
    schema = json.load(open(os.path.join(rewrite_dir, fn + ".schema.json"),
                            encoding="utf-8"))
    fields = schema.get("fields", {})
    return {k: v for k, v in content_data.items() if k in fields}


def checkout_ref(ref, tmpdir):
    """Materialise every component template+schema as committed at `ref`."""
    os.makedirs(tmpdir, exist_ok=True)
    for name in os.listdir(REWRITE):
        if not (name.endswith(".html.tmpl") or name.endswith(".schema.json")):
            continue
        rel = os.path.relpath(os.path.join(REWRITE, name), REPO)
        r = subprocess.run(["git", "show", "%s:%s" % (ref, rel)],
                           capture_output=True, text=True, cwd=REPO)
        if r.returncode != 0:
            continue                     # a file that did not exist at that ref
        open(os.path.join(tmpdir, name), "w", encoding="utf-8").write(r.stdout)
    # render_tool.go is resolved relative to the render dir, so it must be there.
    for helper in ("render_tool.go",):
        src = os.path.join(REWRITE, helper)
        if os.path.exists(src):
            open(os.path.join(tmpdir, helper), "w", encoding="utf-8").write(
                open(src, encoding="utf-8").read())
    return tmpdir


def main():
    apply = "--apply" in sys.argv
    argv = sys.argv[1:]
    ref = DEFAULT_CONTROL_REF
    if "--control-ref" in argv:
        ref = argv[argv.index("--control-ref") + 1]
        argv = [a for i, a in enumerate(argv)
                if i not in (argv.index("--control-ref"), argv.index("--control-ref") + 1)]
    names = [a for a in argv if not a.startswith("--")]
    if not names or (not apply and "--check" not in sys.argv):
        print(__doc__)
        return 2
    for fn in names:
        if fn not in TOOLS:
            print("REFUSE   %s is not one of this lane's components" % fn)
            return 1

    control_dir = checkout_ref(ref, "/tmp/render-tool-row-control")

    for fn in names:
        rows = live_rows(fn)
        if not rows:
            print("REFUSE   %s renders on no page of this site" % fn)
            return 1

        for rid, page, cd, cur_bytes, cur_md5 in rows:
            ov = overrides_from(cd, fn, REWRITE)

            # ── CONTROL: the row as it stands should be reproducible from the
            #    BASELINE template + this row's own content_data. If it is not,
            #    the offline renderer and whatever wrote the live row disagree,
            #    and writing a new render would be shipping an unexplained diff.
            ctl_ov = overrides_from(cd, fn, control_dir)
            try:
                control = render_component(control_dir, fn, ctl_ov)
            except Exception as exc:                       # noqa: BLE001
                print("CONTROL  %-26s %-26s could not render at %s: %s"
                      % (fn, page, ref, exc))
                control = None

            import hashlib
            ctl_md5 = hashlib.md5(control.encode()).hexdigest() if control else None
            ctl_ok = ctl_md5 == cur_md5
            drift = (len(control.encode()) - cur_bytes) if control else None

            # TRAILING NEWLINES ONLY — an explicitly named tolerance, not a fuzzy
            # comparison. `carryStoredSection` -> `save_page_sections` is NOT
            # byte-preserving: it trims the trailing "\n" after "</script>".
            # Measured 2026-08-03 on this very row, which went 8893 -> 8892 bytes
            # through the failed re-render while its CONTENT was untouched. The
            # control must still match everywhere else, so a real drift of one
            # character anywhere in the body still fails.
            ctl_note = ""
            if control and not ctl_ok:
                stored_txt = psql("SELECT rendered_html FROM page_components "
                                  "WHERE id='%s';" % rid)
                if control.rstrip("\n") == stored_txt.rstrip("\n"):
                    ctl_ok = True
                    ctl_note = ("  (identical but for %+d trailing newline(s) — "
                                "the save round-trip trims them)" % drift)

            fresh = render_component(REWRITE, fn, ov)
            new_md5 = hashlib.md5(fresh.encode()).hexdigest()

            print("%-26s %-26s" % (fn, page))
            print("   stored   %6d b  %s" % (cur_bytes, cur_md5[:12]))
            print("   control  %6s b  %s  %s%s"
                  % (len(control.encode()) if control else "-",
                     (ctl_md5 or "-")[:12],
                     "REPRODUCES" if ctl_ok else "DIFFERS by %s byte(s)" % drift,
                     ctl_note))
            print("   fresh    %6d b  %s" % (len(fresh.encode()), new_md5[:12]))

            if new_md5 == cur_md5:
                print("   identical — nothing to do")
                continue
            if not ctl_ok:
                print("   ⛔ CONTROL DID NOT REPRODUCE — not writing. Understand the "
                      "drift first; the offline renderer and the live row disagree.")
                return 1
            if not apply:
                print("   --check: would rewrite %s" % rid)
                continue

            t = dollar_tag(fresh)
            psql_stdin("\n".join([
                "BEGIN;",
                "CREATE TABLE IF NOT EXISTS %s (LIKE page_components INCLUDING ALL);" % BAK,
                "INSERT INTO %s SELECT * FROM page_components WHERE id='%s' "
                "AND NOT EXISTS (SELECT 1 FROM %s WHERE id='%s');" % (BAK, rid, BAK, rid),
                "UPDATE page_components SET rendered_html={t}{h}{t}, updated_at=now() "
                "WHERE id='{i}';".format(t=t, h=fresh, i=rid),
                "COMMIT;",
            ]))
            back = psql("SELECT octet_length(rendered_html), md5(rendered_html) "
                        "FROM page_components WHERE id='%s';" % rid)
            print("   written: %s  (previous row kept in %s)" % (back, BAK))
    return 0


if __name__ == "__main__":
    sys.exit(main())
