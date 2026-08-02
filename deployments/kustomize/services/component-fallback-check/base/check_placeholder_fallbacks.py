#!/usr/bin/env python3
"""check_placeholder_fallbacks.py — ADVISORY standing lint for bugs_open/140.

Flags any ACTIVE content_component whose template SUBSTITUTES A BUSINESS FACT when
the site's datum is absent — the defect the `contact-info` component shipped from
birth until 2026-08-02:

    {{if .phone}}…{{else}}+1234567890{{end}}                  <- tel: href
    {{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm{{end}}

Eight live sites served those invented hours, and vetcomparison.uk served the
invented phone, styled identically to their real details. On a platform whose
whole claims apparatus exists to stop unsourced assertions reaching a page, the
component library itself asserted unverifiable business facts by default.

THE DISTINCTION THIS SCRIPT DRAWS, and it is the whole design:

  A LABEL default is legitimate. "Read more", "Get Started", "Send Message",
  "Contact Information", "Enable JavaScript to see the adoption tracker." — these
  name a control or a section. A site that does not override them is not thereby
  making a claim about itself. The component library is FULL of these and they are
  fine; a lint that reported them would be switched off within a day.

  A FACT default is a fabrication. A phone number, an email address, a postal
  address, a price, a domain, a set of opening hours. Nobody stated it, no evidence
  register holds it, and it renders in the same style as the real thing.

The component's own input_schema already knows the difference — `section_title`
carries "fallback": "Contact Us" with on_missing:use_fallback (a label), while
`phone`/`hours`/`address` carry on_missing:"skip_field" (facts). Only the template
ignored it. This script is the mechanical form of that distinction.

WHY A SCRIPT AND NOT scripts/pattern-check.py: content_components live only in the
database, never in the repo, so a diff-based file linter cannot see them. This is
an OPERATIONAL check — run it by hand, after any component seed, and from a sweep.
Same reasoning, invocation and exit codes as scripts/check_cta_gates.py and
scripts/check_list_empty_states.py.

WHY IT EXISTS AT ALL, given check_placeholder_contact.go already looks for
fabricated contact details: that check matches a ROSTER OF LITERALS against
rendered HTML, so a new component with a NEW invented default is invisible to it
until a human remembers to add the literal — which is exactly how bugs_open/140
survived from the library's birth (its nine patterns matched 1 row fleet-wide
while missing 8 live fabrications, because none of them was our own). This script
reads the LIVE LIBRARY instead of a list, so new entries are measured rather than
remembered. The two are complements: this one guards the source, that one the
artefact.

SECOND CLASS, added 2026-08-02 (owner ruling C on RFC_009): a field whose schema
declares `on_missing: "skip_field"`, which the template REFERENCES but never
GATES. When the datum is absent that renders a BLANK — an empty cell in
platform-comparison, an empty row in product-specs, a missing subheadline —
instead of skipping the element as the contract says. **Reported separately and
deliberately NOT failing the exit code**: it is a milder defect (a blank asserts
nothing untrue, a fabricated literal asserts a falsehood), 68 of them predate
this check, and a red result nobody can clear is a red result everybody learns to
ignore.

WHERE IT RUNS. Two places, one file:
  * by hand, by a session, via kubectl exec (the default);
  * from the `component-fallback-check` CronJob, daily, which CANNOT use kubectl
    (no pods/exec RBAC for ai-persona-app) and so connects to Postgres directly —
    set PG_CLIENTS_HOST + CLIENTS_DB_PASSWORD and it switches automatically.
  `--report` writes ONE doc_notes row per run, on findings AND on a clean result,
  so "looked and found nothing" is distinguishable from "stopped running". That
  ambiguity is exactly what hid bugs_open/140 for months in the sibling detector
  (`placeholder_contact`: 0 items all-history, no way to tell which).

ADVISORY: it is NOT wired into the pre-commit hook (a DB round-trip has no place
blocking a shared commit, and content_components changes routinely land with no
commit at all).
Exit 0 = clean OR ungated-only · 1 = a FABRICATED FACT · 2 = could not reach the DB.

DELIBERATE EXCLUSIONS — report nothing here, on purpose:

  * A bare path or fragment ({{else}}/contact.html, {{else}}#) is a DESTINATION,
    not a fact, and it is scripts/check_cta_gates.py's PLACEHOLDER class. Two
    checks reporting one finding teaches people to ignore both.
  * CSS declarations and style fragments ({{else}}background: linear-gradient(…),
    --hero-ink: var(…)). A default colour is not a claim about the business.
  * Attribute values (lazy, false, true, _blank) and single punctuation.
  * "Contact manufacturer for pricing" — deliberately NOT a price finding. It
    states no figure; it is an honest non-claim, which is the correct thing for a
    component to say when it has no datum.

KNOWN LIMIT, stated rather than implied: the {{else}}…{{end}} match is
non-greedy and does not maintain a block stack, so a fallback containing a nested
{{if}} is reported by its leading text only. It has no false NEGATIVES on the
current library (every literal fallback is flat); if that changes, lift the
block-stack parse from check_cta_gates.py rather than widening the regex.
"""
import argparse
import json
import os
import re
import subprocess
import sys

QUERY = """
SELECT jsonb_agg(jsonb_build_object(
         'name', name, 'function', function,
         'active', is_active,
         'tpl', COALESCE(html_template, ''),
         'fields', COALESCE(input_schema->'fields', '{}'::jsonb)))
  FROM content_components WHERE is_active;
"""

# Two ways to reach the database, because this runs in two places.
#   interactive  — a session at a terminal, via kubectl exec (the default).
#   in-cluster   — the CronJob, which CANNOT use kubectl: the ai-persona-app
#                  service account has no pods/exec RBAC in this namespace, so a
#                  kubectl-only script fails there in a way that looks like a
#                  clean run if you are not watching exit codes. Same constraint
#                  single-owner-carriers-check and bugs-open-staleness-sweep hit.
# Selected by PG_CLIENTS_HOST being set, which is how the CronJob declares itself.
PSQL_KUBECTL = ["kubectl", "exec", "-n", "ai-persona-system", "postgres-clients-0", "--",
                "psql", "-U", "clients_user", "-d", "clients_db", "-tAc"]


def psql_argv():
    """Return (argv_prefix, env) for a -tAc query, choosing the reachable route."""
    host = os.environ.get("PG_CLIENTS_HOST")
    if not host:
        return PSQL_KUBECTL, None
    env = os.environ.copy()
    pw = os.environ.get("CLIENTS_DB_PASSWORD")
    if not pw:
        print("PG_CLIENTS_HOST is set but CLIENTS_DB_PASSWORD is not — refusing to "
              "guess a connection", file=sys.stderr)
        sys.exit(2)
    env["PGPASSWORD"] = pw
    return (["psql", "-h", host, "-p", "5432", "-U", "clients_user",
             "-d", "clients_db", "-tAc"], env)

# {{if …}} IF_BRANCH {{else}} LITERAL {{end}} — captured together, because whether
# the literal is a SUBSTITUTE is decidable only against the branch it replaces.
FALLBACK_BLOCK = re.compile(
    r"\{\{-?\s*(?:if|else\s+if)\s+[^}]*?-?\}\}(.*?)\{\{-?\s*else\s*-?\}\}(.*?)\{\{-?\s*end\s*-?\}\}",
    re.S)
TAG = re.compile(r"<[^>]+>")

# ---------------------------------------------------------------------------
# FACT shapes. Each one asserts something checkable about a business.
# ---------------------------------------------------------------------------
FACT_SHAPES = [
    ("phone",
     # +1234567890 / +1 (234) 567-890 / (555) 123-4567 / 07934 524 911
     re.compile(r"(?:\+\d[\d\s().-]{7,}\d)|(?:\(\d{3}\)\s*\d{3}[-\s]?\d{3,4})|(?:\b0\d{3,4}[\s-]?\d{3}[\s-]?\d{3,4}\b)")),
    ("email",
     re.compile(r"\b[\w.+-]+@[\w-]+\.[A-Za-z]{2,}\b")),
    ("domain",
     # a bare hostname with a real TLD — one site's domain as every site's default
     re.compile(r"\b(?:[\w-]+\.)+(?:com|co\.uk|uk|org|net|io|ai|dev)\b")),
    ("address",
     re.compile(r"\b\d+[a-zA-Z]?\s+[A-Z][\w'-]*\s+(?:St|Street|Ave|Avenue|Rd|Road|Ln|Lane|Way|Close|Drive|Dr|Blvd)\b")),
    ("price",
     re.compile(r"[£$€]\s?\d")),
    # Two forms, because a day name is not required to state opening hours and
    # relying on one was this script's first false NEGATIVE: the control
    # "Weekdays 8am to 5pm" — a plain fabrication — slipped through a
    # day-name-anchored pattern.
    ("opening_hours",
     re.compile(r"(?i)"
                # a time RANGE: "9am – 6pm", "8am to 5pm", "08:00-17:00"
                r"(?:\b\d{1,2}(?::\d{2})?\s*(?:am|pm)\b.{0,20}?\b\d{1,2}(?::\d{2})?\s*(?:am|pm)\b)"
                r"|"
                # or a day/period word carrying a time
                r"(?:\b(?:mon|tue|wed|thu|fri|sat|sun|weekday|weekend|daily)[a-z]*\b"
                r"[^|]{0,40}?\b\d{1,2}\s*(?:am|pm|:\d{2}))")),
]

# Shapes that are NOT facts even though they may contain digits or dots.
CSS_DECL = re.compile(r"(?:^|;)\s*(?:--)?[a-z-]+\s*:\s*\S")
BARE_DEST = re.compile(r"^[#/]")
ATTR_VALUE = {"lazy", "eager", "false", "true", "_blank", "_self", "none", "auto"}


def classify(literal):
    """Return a fact-shape name, or None when the literal is a legitimate label."""
    text = TAG.sub(" ", literal).strip()
    text = re.sub(r"\s+", " ", text)

    if len(text) < 3:
        return None                      # '.', '#', punctuation
    if text.lower() in ATTR_VALUE:
        return None
    if BARE_DEST.match(text):
        return None                      # a destination — check_cta_gates.py's class
    if CSS_DECL.search(text):
        return None                      # a style fragment, not a claim

    for name, pat in FACT_SHAPES:
        if pat.search(text):
            return name
    return None


def flatten(html):
    return re.sub(r"\s+", " ", TAG.sub(" ", html)).strip()


# ---------------------------------------------------------------------------
# Second finding class, added 2026-08-02 on the owner's C ruling for RFC_009.
#
# A field whose input_schema declares "on_missing": "skip_field", which the
# template REFERENCES but never GATES. When the datum is absent this renders a
# blank — an empty cell in platform-comparison, an empty row in product-specs, a
# missing subheadline — rather than skipping the element as the contract says.
#
# THIS IS A MILDER CLASS THAN THE FABRICATIONS ABOVE and the report says so.
# An ungated field with no fallback renders NOTHING, which is ugly; an ungated
# field WITH a fabricated literal renders a FALSE FACT, which is the bug this
# script was written for. They are reported separately so the second can never be
# lost in the noise of the first — 68 of the milder against 0 of the severe, when
# this was written.
#
# WHY THIS MATTERS MORE THAN IT LOOKS: RFC_009's measured finding is that ~90% of
# fields declare no on_missing at all, so a render-time enforcement gate would be
# inert for nine fields in ten. This check does not depend on the declaration
# being widespread — it reports on the 10% that DO declare, which is exactly the
# population where the contract is checkable.
# ---------------------------------------------------------------------------

def field_used(tpl, field):
    return re.search(r"\{\{[^}]*\.%s\b" % re.escape(field), tpl) is not None


def field_gated(tpl, field):
    """APPROXIMATE — a gate ANYWHERE in the template, not a block-stack scope check.

    Over-counts as gated a field guarded in one place and used bare in another;
    it does not under-count, so a clean result here is meaningful and a finding is
    worth reading rather than trusting blindly. If this ever needs to be exact,
    lift the block-stack parse from check_cta_gates.py rather than widening the
    regex — that parse already exists and was written for this shape of question.
    """
    return re.search(r"\{\{-?\s*(?:if|with)\s+[^}]*\.%s\b" % re.escape(field), tpl) is not None


def scan_ungated_skip_fields(components):
    """Return [(component, function, [fields])] declaring skip_field but never gating it."""
    out = []
    for c in components:
        tpl = c.get("tpl") or ""
        fields = c.get("fields") or {}
        if not isinstance(fields, dict):
            continue
        bad = [f for f, spec in fields.items()
               if isinstance(spec, dict)
               and spec.get("on_missing") == "skip_field"
               and field_used(tpl, f)
               and not field_gated(tpl, f)]
        if bad:
            out.append((c["name"], c.get("function") or "", sorted(bad)))
    return sorted(out)


def scan(components):
    """Return findings: list of (fact_shape, component, function, literal)."""
    findings = []
    for c in components:
        tpl = c.get("tpl") or ""
        for m in FALLBACK_BLOCK.finditer(tpl):
            if_branch, literal = m.group(1), m.group(2)
            if "{{" in literal:
                continue                 # not a literal fallback; a nested action

            flat = flatten(literal)

            # A fallback that renders the SAME text as the branch it replaces is
            # not inventing anything — it is one constant rendered two ways, the
            # commonest being "link it if we have a URL, otherwise print it".
            # about-commercial-block's builder attribution is exactly this:
            #   {{if .built_by_url}}<a href=…>fundamentallyai.com</a>
            #   {{else}}fundamentallyai.com{{end}}
            # Reporting that as a fabricated domain was this script's first false
            # positive, caught by reading the template before believing the tool.
            if flat and flat in flatten(if_branch):
                continue

            shape = classify(literal)
            if shape:
                findings.append((shape, c["name"], c.get("function") or "", flat[:90]))
    return findings


def load(path):
    if path:
        with open(path) as fh:
            return json.load(fh)
    argv, env = psql_argv()
    try:
        out = subprocess.run(argv + [QUERY], capture_output=True, text=True,
                             timeout=120, env=env)
    except (subprocess.TimeoutExpired, FileNotFoundError) as exc:
        print("could not reach the database: %s" % exc, file=sys.stderr)
        sys.exit(2)
    if out.returncode != 0:
        print("could not reach the database: %s" % out.stderr.strip(), file=sys.stderr)
        sys.exit(2)
    body = out.stdout.strip()
    if not body:
        print("no active components returned — is this the right database?", file=sys.stderr)
        sys.exit(2)
    return json.loads(body)


def write_doc_note(body):
    """Record the run in doc_notes so agents and later sessions can read it.

    Written on EVERY run, clean or not — a check that only speaks when it fails is
    indistinguishable from one that has stopped running, which is the exact
    ambiguity bugs_open/140 hit with `placeholder_contact` (0 items all-history,
    and no way to tell 'looked and found nothing' from 'never ran').
    Best-effort: a failure to record must never become a failure to report.
    """
    argv, env = psql_argv()
    sql = ("INSERT INTO doc_notes (subject_type, subject_key, body, categories, source) "
           "VALUES ('pipeline', 'component-library', %s, '[\"component-integrity\"]'::jsonb, "
           "'check_placeholder_fallbacks')" % literal_sql(body))
    try:
        subprocess.run(argv + [sql], capture_output=True, text=True, timeout=60, env=env)
    except Exception as exc:                                   # noqa: BLE001
        print("(could not write doc_notes: %s)" % exc, file=sys.stderr)


def literal_sql(s):
    return "'" + s.replace("'", "''") + "'"


FIXTURE = os.path.join(os.path.dirname(os.path.realpath(__file__)),
                       "..", "..", "..", "..", "..",
                       "platform", "orchestration", "actions", "testdata",
                       "component_fallback_fixtures.json")


def selftest():
    """Check this implementation against the SHARED fixture.

    The same file is read by component_fallback_guard_test.go. Two implementations
    of one rule is the drift class this very rule detects, so they are pinned to a
    single set of cases rather than to each other's good intentions. RFC_009 option
    B records why both exist: the gate must run in the Go write path with no DB,
    the lint must read the live library on a schedule.
    """
    try:
        with open(FIXTURE) as fh:
            cases = json.load(fh)["cases"]
    except (OSError, ValueError) as exc:
        print("cannot read the shared fixture %s: %s" % (FIXTURE, exc), file=sys.stderr)
        return 2

    fired = quiet = 0
    bad = []
    for c in cases:
        got = scan([{"name": c["name"], "function": "", "tpl": c["template"], "fields": {}}])
        want = c.get("expect")
        if want is None:
            quiet += 1
            if got:
                bad.append("%s\n     want: allowed (legitimate label)\n     got:  refused as %r"
                           % (c["name"], got[0][0]))
        else:
            fired += 1
            if not got:
                bad.append("%s\n     want: refused as %r\n     got:  allowed" % (c["name"], want))
            elif got[0][0] != want:
                bad.append("%s\n     want shape %r, got %r" % (c["name"], want, got[0][0]))

    if not fired or not quiet:
        print("the fixture must exercise BOTH arms; fired=%d quiet=%d" % (fired, quiet),
              file=sys.stderr)
        return 1
    if bad:
        print("selftest FAILED — this implementation disagrees with the shared fixture,\n"
              "so it has drifted from the Go birth gate (component_fallback_guard.go):\n")
        for b in bad:
            print("  " + b)
        return 1
    print("selftest OK — %d must-refuse and %d must-allow cases agree with the shared "
          "fixture (%s)" % (fired, quiet, os.path.normpath(FIXTURE)))
    return 0


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--selftest", action="store_true",
                    help="check this implementation against the shared Go/Python fixture")
    ap.add_argument("--json", help="read components from a file instead of the cluster "
                                   "(same shape as the QUERY above)")
    ap.add_argument("--quiet", action="store_true", help="findings only, no summary")
    ap.add_argument("--report", action="store_true",
                    help="also write a doc_notes row (used by the CronJob)")
    args = ap.parse_args()

    if args.selftest:
        return selftest()

    components = load(args.json)
    findings = scan(components)
    ungated = scan_ungated_skip_fields(components)
    lines = []

    if findings:
        lines.append("FABRICATED FACTS — %d finding(s). A component must not invent a "
                     "business fact." % len(findings))
        for shape, name, function, literal in sorted(findings):
            lines.append("  FABRICATED_%-14s %s%s" % (
                shape.upper(), name,
                (" (%s)" % function if function and function != name else "")))
            lines.append("      renders %r when the site supplies no datum" % literal)
        lines.append("  Gate the element on its own datum and DELETE the fallback — the "
                     "schema's on_missing:\"skip_field\" is already the contract.")
        lines.append("  Worked case: bugs_open/140, migration 287.")

    if ungated:
        n = sum(len(f) for _, _, f in ungated)
        if lines:
            lines.append("")
        lines.append("UNGATED skip_field — %d field(s) across %d component(s). MILDER: these "
                     "render a BLANK, not a false fact." % (n, len(ungated)))
        for name, function, fields in ungated:
            lines.append("  %-38s %s" % (
                name + (" (%s)" % function if function and function != name else ""),
                ", ".join(fields)))
        lines.append("  Each declares on_missing:\"skip_field\" and is referenced but never "
                     "gated, so an absent datum renders empty instead of being skipped.")
        lines.append("  [APPROXIMATE — tests for a gate anywhere, not a scope check: "
                     "over-counts as gated, never under-counts.] See RFC_009.")

    if not lines:
        msg = ("check_placeholder_fallbacks: clean — %d active components; no template "
               "substitutes a business fact for an absent datum, and every declared "
               "skip_field that is rendered is gated." % len(components))
        if not args.quiet:
            print(msg)
        if args.report:
            write_doc_note(msg)
        return 0

    header = ("check_placeholder_fallbacks: %d fabricated-fact finding(s) and %d ungated "
              "skip_field(s), across %d active components"
              % (len(findings), sum(len(f) for _, _, f in ungated), len(components)))
    body = header + "\n\n" + "\n".join(lines)
    print(body)
    if args.report:
        write_doc_note(body)
    # Exit 1 only for the SEVERE class. The ungated-blank class is reported but must
    # not fail a gate on its own: 68 of them predate this check, and a red result
    # nobody can clear is a red result everybody learns to ignore.
    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main())
