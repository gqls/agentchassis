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

ADVISORY: exits non-zero on findings so a caller can decide; it is NOT wired into
the pre-commit hook (a DB round-trip has no place blocking a shared commit).
Exit 0 = clean, 1 = findings, 2 = could not reach the DB.

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
import re
import subprocess
import sys

QUERY = """
SELECT jsonb_agg(jsonb_build_object(
         'name', name, 'function', function,
         'active', is_active,
         'tpl', COALESCE(html_template, '')))
  FROM content_components WHERE is_active;
"""

PSQL = ["kubectl", "exec", "-n", "ai-persona-system", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-tAc"]

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
    try:
        out = subprocess.run(PSQL + [QUERY], capture_output=True, text=True, timeout=120)
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


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--json", help="read components from a file instead of the cluster "
                                   "(same shape as the QUERY above)")
    ap.add_argument("--quiet", action="store_true", help="findings only, no summary")
    args = ap.parse_args()

    components = load(args.json)
    findings = scan(components)

    if not findings:
        if not args.quiet:
            print("check_placeholder_fallbacks: clean — %d active components, "
                  "no template substitutes a business fact for an absent datum"
                  % len(components))
        return 0

    print("check_placeholder_fallbacks: %d finding(s) across %d active components\n"
          % (len(findings), len(components)))
    for shape, name, function, literal in sorted(findings):
        print("  FABRICATED_%-14s %s" % (shape.upper(), name)
              + (" (%s)" % function if function and function != name else ""))
        print("      renders %r when the site supplies no datum" % literal)
    print("\nA component must not invent a business fact. Gate the element on its own\n"
          "datum and delete the fallback — the schema's on_missing:\"skip_field\" is\n"
          "already the contract. See bugs_open/140 and migration 287 for the worked case.")
    return 1


if __name__ == "__main__":
    sys.exit(main())
