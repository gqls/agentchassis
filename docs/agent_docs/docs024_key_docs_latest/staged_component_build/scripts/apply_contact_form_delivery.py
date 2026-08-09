#!/usr/bin/env python3
"""Wire `contact-block` and `contact-form` to a REAL destination (bugs_open/228).

WHAT THIS CHANGES, and why each edit is the smallest one that closes the door:

  contact-block
    template  add `action="{{.form_action}}" method="POST"` to the <form>.
              This is the load-bearing edit and it is NOT cosmetic:
              RenderTemplateReportingMissing only seeds `form_action`, and
              sanitiseFormAction (component_library.go) only runs, when the
              TEMPLATE MENTIONS `form_action`. The component never did — which
              is the mechanical reason the platform's own per-site address
              repair, live since 2026-07-24, has never touched this component
              while repairing its sibling. One attribute puts it inside the
              existing machinery instead of building a second one.
    js        replace the 1,200 ms setTimeout that printed "Your message has
              been sent" with delivery whose outcome is the destination's.

  contact-form
    template  add a status element (`#cf-status`) and a `<script src>` for the
              new asset, and an `id` on the form so the script can bind it.
              Its action was already `{{.form_action}}`; what it lacked was any
              way to tell the visitor what happened.
    js        NEW js_content — this component had none.

WHAT IT DOES NOT CHANGE: `sites.email`, any content_data, any rendered_html.
Addresses come from the platform, not from here.

SAFETY. Every edit is an exact string replacement asserted to occur EXACTLY
ONCE before it is applied, and every write is verified inside the transaction by
a DO/RAISE on the resulting length and on the presence of the new markers —
never a bare SELECT, because ON_ERROR_STOP ignores a non-empty result set and a
verify block made of SELECTs cannot stop the COMMIT.

USAGE
  apply_contact_form_delivery.py            -> dry run (prints SQL, ROLLBACK)
  apply_contact_form_delivery.py --apply    -> COMMIT
Pipe to psql; never let the shell see the bodies.
"""
import os
import subprocess
import sys

DIR = os.path.dirname(os.path.abspath(__file__))

PSQL = [
    "kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
    "psql", "-U", "clients_user", "-d", "clients_db",
]


def fetch(sql: str) -> str:
    return subprocess.run(PSQL + ["-tAc", sql], capture_output=True, text=True,
                          check=True).stdout


def read(name: str) -> str:
    with open(os.path.join(DIR, name), "r", encoding="utf-8") as fh:
        return fh.read()


def replace_once(haystack: str, needle: str, repl: str, what: str) -> str:
    n = haystack.count(needle)
    if n != 1:
        sys.stderr.write(f"ABORT: {what}: target occurs {n} times, expected exactly 1:\n  {needle!r}\n")
        sys.exit(2)
    return haystack.replace(needle, repl, 1)


def q(s: str) -> str:
    """Dollar-quote a body. The tag is checked against the content."""
    tag = "$cc$"
    if tag in s:
        sys.stderr.write("ABORT: dollar-quote tag collides with content\n")
        sys.exit(2)
    return tag + s + tag


def verify_block(fn: str, tpl_len: int, js_len: int, tpl_markers, js_markers) -> str:
    """Assert length AND that each marker is present in the column it belongs to.

    The two marker lists are separate deliberately: the first version of this
    checked every marker against the template, so a JS marker could only ever
    fail. A verify that can only fail is as useless as one that can only pass —
    it just fails loudly instead of quietly.
    """
    checks = "\n".join(
        [f"  IF position({q(m)} in t) = 0 THEN\n"
         f"    RAISE EXCEPTION 'marker missing from {fn} TEMPLATE: %', {q(m[:40])};\n"
         f"  END IF;" for m in tpl_markers] +
        [f"  IF position({q(m)} in j) = 0 THEN\n"
         f"    RAISE EXCEPTION 'marker missing from {fn} JS: %', {q(m[:40])};\n"
         f"  END IF;" for m in js_markers]
    )
    return f"""DO $$
DECLARE t text; j text; c int;
BEGIN
  SELECT html_template, coalesce(js_content,''), count(*) OVER ()
    INTO t, j, c
    FROM content_components
   WHERE function='{fn}' AND is_active AND component_level='section';
  IF c IS DISTINCT FROM 1 THEN
    RAISE EXCEPTION 'expected exactly 1 active {fn} row, found %', c;
  END IF;
  IF length(t) IS DISTINCT FROM {tpl_len} THEN
    RAISE EXCEPTION '{fn} template length is %, expected {tpl_len}', length(t);
  END IF;
  IF length(j) IS DISTINCT FROM {js_len} THEN
    RAISE EXCEPTION '{fn} js length is %, expected {js_len}', length(j);
  END IF;
{checks}
  RAISE NOTICE 'OK {fn}: template % bytes, js % bytes', length(t), length(j);
END $$;"""


def main() -> int:
    apply = "--apply" in sys.argv[1:]

    # ---------------- contact-block ----------------
    cb_tpl = fetch("SELECT html_template FROM content_components "
                   "WHERE function='contact-block' AND is_active AND component_level='section';")
    cb_tpl = cb_tpl[:-1] if cb_tpl.endswith("\n") else cb_tpl
    if "form_action" in cb_tpl:
        sys.stderr.write("ABORT: contact-block template already mentions form_action — "
                         "re-read it before re-running this.\n")
        return 2
    cb_tpl_new = replace_once(
        cb_tpl,
        '<form class="cb-form" id="cb-contact-form" novalidate aria-label="{{.form_heading}}">',
        '<form class="cb-form" id="cb-contact-form" action="{{.form_action}}" method="POST"'
        ' novalidate aria-label="{{.form_heading}}">',
        "contact-block form tag",
    )
    cb_js_new = read("contact_block.js")

    # ---------------- contact-form ----------------
    cf_tpl = fetch("SELECT html_template FROM content_components "
                   "WHERE function='contact-form' AND is_active AND component_level='section';")
    cf_tpl = cf_tpl[:-1] if cf_tpl.endswith("\n") else cf_tpl
    if "cf-status" in cf_tpl:
        sys.stderr.write("ABORT: contact-form template already carries #cf-status.\n")
        return 2
    cf_tpl_new = replace_once(
        cf_tpl,
        '<form class="contact-form" action="{{.form_action}}" method="POST">',
        '<form class="contact-form" id="cf-contact-form" action="{{.form_action}}" method="POST">',
        "contact-form form tag",
    )
    cf_tpl_new = replace_once(
        cf_tpl_new,
        '            <button type="submit" class="form-submit">',
        '            <div class="contact-form-status" id="cf-status" role="alert" aria-live="polite"></div>\n'
        '            <button type="submit" class="form-submit">',
        "contact-form status element",
    )
    # The <script src> goes after </section>, matching what
    # store_generated_component_action.go emits for a component with js_content.
    cf_tpl_new = replace_once(
        cf_tpl_new,
        "</section>\n<style>",
        '</section>\n<script src="/tools/assets/contact-form.js"></script>\n<style>',
        "contact-form script ref",
    )
    # Status styling, so an error is legible on every palette. Layout only,
    # colours from CSS custom properties, matching this template's own comment.
    cf_tpl_new = replace_once(
        cf_tpl_new,
        ".form-submit {\n    padding: 1rem 2rem;",
        ".contact-form-status {\n"
        "    min-height: 1.25rem;\n"
        "    font-size: 0.9375rem;\n"
        "    line-height: 1.5;\n"
        "}\n"
        ".contact-form-status.cf-error {\n"
        "    color: var(--color-error, #b91c1c);\n"
        "}\n"
        ".contact-form-status.cf-success {\n"
        "    color: var(--color-success, #15803d);\n"
        "}\n"
        ".form-submit {\n    padding: 1rem 2rem;",
        "contact-form status styling",
    )
    cf_js_new = read("contact_form.js")

    out = ["BEGIN;", ""]
    out.append("-- ---------- contact-block: give the form a destination, delete the timer ----------")
    out.append(f"UPDATE content_components SET html_template={q(cb_tpl_new)},"
               f" js_content={q(cb_js_new)}, updated_at=now()\n"
               f" WHERE function='contact-block' AND is_active AND component_level='section';")
    out.append(verify_block("contact-block", len(cb_tpl_new), len(cb_js_new),
        ['action="{{.form_action}}"'],
        # Markers must be SINGLE-LINE strings. The first attempt used a phrase
        # from the file's own header comment and it failed — the comment wraps,
        # so the phrase exists in the reader's mind and not in the bytes.
        ["window.location.href = url;",
         "setStatus('success', 'Your message has been sent."]))
    out.append("")
    out.append("-- ---------- contact-form: say what actually happened ----------")
    out.append(f"UPDATE content_components SET html_template={q(cf_tpl_new)},"
               f" js_content={q(cf_js_new)}, updated_at=now()\n"
               f" WHERE function='contact-form' AND is_active AND component_level='section';")
    out.append(verify_block("contact-form", len(cf_tpl_new), len(cf_js_new),
        ['id="cf-status"', '<script src="/tools/assets/contact-form.js"></script>'],
        ["window.location.href = url;",
         "setStatus('success', 'Your message has been sent."]))
    out.append("")
    out.append("COMMIT;" if apply else "ROLLBACK;  -- dry run: pass --apply to commit")
    print("\n".join(out))
    return 0


if __name__ == "__main__":
    sys.exit(main())
