#!/usr/bin/env python3
"""check_card_slot_guards.py — ADVISORY standing lint for bugs_open/425.

Flags any ACTIVE content_component that, INSIDE a {{range}}, renders an element
whose entire content is a single per-item interpolation ( <p>{{.excerpt}}</p> )
with no conditional naming that field. When the producer does not write the key,
such an element renders EMPTY — and an empty element is not nothing: it keeps
its padding, its margin and its place in the flex or grid row.

WHY THIS IS A CLASS AND NOT A SITE BUG. An empty-slot-tolerant component
silently converts a DATA GAP into a DESIGN FLAW, and the design flaw is what
gets reported. On boxingonline.com the owner asked for "better card designs";
the cards were structurally fine and were shipping four empty slots per card,
because `query.blog_posts` never wrote category, excerpt, date or read_time and
the template rendered all four unconditionally. Nobody looks for a missing
excerpt. They look at the card.

HOW IT DIFFERS FROM check_list_empty_states.py — they are complements, and the
distinction is the whole reason this is a second script:
  * that one asks whether the ARRAY can be empty (no items at all → a blank
    container where an empty-state should be);
  * this one asks whether an ITEM'S FIELD can be empty (items present, slots
    blank). A component can pass that check and fail this one, which is exactly
    what content-listing did from the library's birth until 2026-09-02.

WHY A SCRIPT AND NOT scripts/pattern-check.py: content_components live only in
the database, never in the repo, so a diff-based file linter cannot see them.
This is an OPERATIONAL check — run it by hand or from a sweep.

ADVISORY: reports and exits non-zero on findings so a caller can decide; NOT
wired into the pre-commit hook (a DB round-trip must not block a shared commit).
Exit 0 = clean, 1 = unguarded slot(s) found, 2 = could not reach the DB.

NOT REPORTED, deliberately: a slot whose own open tag interpolates an item
field (<a href="{{.url}}">{{.title}}</a>). That is a structural anchor — a
missing .url makes the card broken rather than blank, which is bugs_open/309's
class and has its own detector. See ATTR_INTERPOLATION.

GOTCHA, stated rather than hidden: the guard test proves a conditional NAMING
the field exists somewhere in the range body, not that it encloses the element.
It can therefore MISS a slot that is named by an unrelated conditional. It is a
candidate flagger; read the template before editing.

AND IT CANNOT TELL A REQUIRED SLOT FROM AN OPTIONAL ONE — that is a real limit
with a structural cause worth knowing. `content_components.input_schema` carries
`required` / `on_missing` / `fallback` per TOP-LEVEL field, and the estate's
resolver honours them; it has NO vocabulary for the fields INSIDE an array item.
So `articles {required: true, on_missing: skip_section}` governs whether the
listing renders at all and says nothing whatever about `articles[].excerpt`.
There is therefore no declaration to join against, and a headline slot and a
read-time slot look identical from here. Closing that gap — per-item field
declarations, so `on_missing` reaches inside a {{range}} — is the durable fix
this check is a stand-in for (bugs_open/425 §fix-candidate 4).
"""
import argparse
import re
import sys

sys.path.insert(0, __file__.rsplit("/", 1)[0])
from component_template_lib import (  # noqa: E402
    DBUnreachable, active_templates, mentions_in_condition, range_bodies,
)

# An element whose ENTIRE content is one per-item interpolation. Attributes are
# allowed; anything else between the tags is not, because a slot with sibling
# text still renders that text and does not collapse to nothing.
SOLE_SLOT = re.compile(
    r"<([a-zA-Z][\w-]*)\b([^>]*)>\s*\{\{-?\s*\.([A-Za-z_]\w*)\s*-?\}\}\s*</\1\s*>")

# A slot whose OPEN TAG itself interpolates an item field — <a href="{{.url}}">
# — is STRUCTURAL, not decorative, and is deliberately not reported here.
# Guarding it is the wrong remedy: an anchor exists in order to be a link, so a
# missing .url is a broken card rather than a slot to collapse, and that is
# already bugs_open/309's class (cards rendered unclickable) with its own
# detector. Reporting it too would put an unactionable line above every real
# finding, which is how a check stops being read.
ATTR_INTERPOLATION = re.compile(r"=\s*[\"\']?\{\{-?\s*\.")

# Wrappers whose children are all guarded but which are not themselves guarded
# still occupy layout. Same defect one level up — how content-listing's
# section__header survived a fix aimed at its children.
WRAPPER = re.compile(
    r"<(div|section|footer|header|ul|ol)\b[^>]*>((?:\s*\{\{-?\s*if[^}]*\}\}.*?\{\{-?\s*end\s*-?\}\}\s*)+)</\1\s*>",
    re.S)


def _wrappers(fragment, scope_label, name, out):
    """Wrappers whose children are ALL conditional but which are not themselves.

    Such a wrapper still renders — empty — with its own margin and padding, so
    guarding only its children moves the defect up one level rather than
    removing it.
    """
    for m in WRAPPER.finditer(fragment):
        tag = m.group(1)
        fields = re.findall(r"\{\{-?\s*if[^}]*?\.([A-Za-z_]\w*)", m.group(2))
        if not fields:
            continue
        if re.search(r"\{\{-?\s*if[^}]*\}\}\s*<%s\b" % re.escape(tag), fragment):
            continue
        out.append((name, scope_label,
                    f"<{tag}> wrapping only guarded children ({', '.join(sorted(set(fields)))})",
                    "wrapper"))


def findings_for(name, template):
    """Per-item slots inside every {{range}}, plus empty-able wrappers ANYWHERE.

    THE WRAPPER SCAN IS NOT LIMITED TO RANGE BODIES, and that is the point.
    The first cut of this check scanned only inside {{range}}, which made it
    blind to the very defect the bugs_open/425 render proof found: content-listing's
    `section__header` sits OUTSIDE the range, both its children were guarded, the
    wrapper was not, and a section with neither title nor subtitle rendered an
    empty <div> carrying its own margin. A detector that cannot find one of the
    two defects its own bug fixed is not a detector for that bug — caught by the
    council's editquality seat (round 84b51f16) pointing at the migration's
    thinner positive control, which is the same blind spot one file along.
    """
    out = []
    for coll, body in range_bodies(template):
        for m in SOLE_SLOT.finditer(body):
            tag, attrs, field = m.group(1), m.group(2), m.group(3)
            if ATTR_INTERPOLATION.search(attrs):
                continue
            if not mentions_in_condition(body, field):
                out.append((name, coll, f"<{tag}>{{{{.{field}}}}}</{tag}>", "slot"))
        _wrappers(body, coll, name, out)

    # Whole-template pass for wrappers, with the range bodies blanked so the
    # per-item wrappers already reported above are not counted twice.
    outside = template
    for _, body in range_bodies(template):
        outside = outside.replace(body, " " * len(body), 1)
    _wrappers(outside, "(section level)", name, out)
    return out


# ---------------------------------------------------------------------------
# SELF-TEST. The point is that the detector can come out BOTH ways: it must fire
# on the pre-682 content-listing template and go quiet on the post-682 one. A
# check only ever run against a tree that already carries its own fix reports
# the same number whatever is true (WRONG_CALLS.md, 2026-08-03).
_PRE_682 = """<div class="section__header">
  {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
  {{if .section_subtitle}}<p>{{.section_subtitle}}</p>{{end}}
</div>
{{range .articles}}
<article>
  {{if .image}}<div class="c__image">
    <img src="{{.image}}" alt="{{.title}}">
    <span class="c__category">{{.category}}</span>
  </div>{{end}}
  <h3><a href="{{.url}}">{{.title}}</a></h3>
  <p class="c__excerpt">{{.excerpt}}</p>
  <div class="c__meta">
    <span class="c__date">{{.date}}</span>
    <span class="c__read-time">{{.read_time}}</span>
  </div>
</article>
{{end}}"""

_POST_682 = """{{if or .section_title .section_subtitle}}<div class="section__header">
  {{if .section_title}}<h2>{{.section_title}}</h2>{{end}}
  {{if .section_subtitle}}<p>{{.section_subtitle}}</p>{{end}}
</div>{{end}}
{{range .articles}}
<article>
  {{if .image}}<div class="c__image">
    <img src="{{.image}}" alt="{{.title}}">
    {{if .category}}<span class="c__category">{{.category}}</span>{{end}}
  </div>{{end}}
  <h3><a href="{{.url}}">{{.title}}</a></h3>
  {{if .excerpt}}<p class="c__excerpt">{{.excerpt}}</p>{{end}}
  {{if or .date .read_time}}<div class="c__meta">
    {{if .date}}<span class="c__date">{{.date}}</span>{{end}}
    {{if .read_time}}<span class="c__read-time">{{.read_time}}</span>{{end}}
  </div>{{end}}
</article>
{{end}}"""


def self_test() -> int:
    failures = []

    pre = findings_for("pre-682", _PRE_682)
    pre_fields = sorted(f[2] for f in pre if f[3] == "slot")
    expected = ['<p>{{.excerpt}}</p>', '<span>{{.category}}</span>',
                '<span>{{.date}}</span>', '<span>{{.read_time}}</span>']
    if pre_fields != expected:
        failures.append(f"pre-682 slots: got {pre_fields}, want {expected}")

    # The section-level wrapper: pre-682 must report it, post-682 must not.
    # This arm is what the first cut of the check could not do at all.
    if not any(f[3] == "wrapper" and f[1] == "(section level)" for f in pre):
        failures.append("pre-682 section__header wrapper was NOT reported — the "
                        "wrapper scan is blind outside {{range}} again")

    post = findings_for("post-682", _POST_682)
    if post:
        failures.append(f"post-682 should be clean, got {post}")

    # The structural anchor must be excluded in BOTH, or the exclusion is
    # silently doing nothing.
    if any(".title" in f[2] for f in pre + post):
        failures.append("structural anchor <a href={{.url}}>{{.title}}</a> was reported")

    # The scanner must not read past a range's own {{end}}.
    tail = findings_for("tail", "{{range .xs}}<p>{{.a}}</p>{{end}}<span>{{.b}}</span>")
    if any(".b" in f[2] for f in tail):
        failures.append("scanner ran past the range's {{end}}")

    # A nested {{if}} inside the body must not end the body early.
    nested = findings_for("nested",
                          "{{range .xs}}{{if .g}}<i>x</i>{{end}}<p>{{.late}}</p>{{end}}")
    if not any(".late" in f[2] for f in nested):
        failures.append("a nested {{if}} truncated the range body")

    for f in failures:
        print(f"  FAIL: {f}")
    if failures:
        print(f"self-test: {len(failures)} failure(s)")
        return 1
    print("self-test: OK (fires on pre-682, quiet on post-682, anchor excluded, "
          "nesting handled)")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--component", help="only check this component name")
    ap.add_argument("--quiet", action="store_true", help="print findings only")
    ap.add_argument("--self-test", action="store_true",
                    help="prove the detector fires and goes quiet; no DB access")
    args = ap.parse_args()

    if args.self_test:
        return self_test()

    try:
        rows = active_templates()
    except DBUnreachable as exc:
        print(f"check_card_slot_guards: could not reach the DB: {exc}", file=sys.stderr)
        return 2
    if not rows:
        print("check_card_slot_guards: no active components with a {{range}} found "
              "(unexpected — the query succeeded, so the library is genuinely empty)",
              file=sys.stderr)
        return 2

    if args.component:
        rows = [r for r in rows if r[0] == args.component]
        if not rows:
            print(f"check_card_slot_guards: no active component named {args.component!r}",
                  file=sys.stderr)
            return 2

    findings = []
    for name, tmpl in rows:
        findings.extend(findings_for(name, tmpl))

    if not findings:
        print(f"OK: all {len(rows)} active {{{{range}}}} components guard every "
              f"single-interpolation slot.")
        return 0

    slots = [f for f in findings if f[3] == "slot"]
    wrappers = [f for f in findings if f[3] == "wrapper"]
    affected = sorted({f[0] for f in findings})
    print(f"UNGUARDED per-item slots: {len(slots)} slot(s) and {len(wrappers)} "
          f"wrapper(s) across {len(affected)} of {len(rows)} active "
          f"{{{{range}}}} components.\n")
    current = None
    for name, coll, what, kind in sorted(findings):
        if name != current:
            print(f"  {name}")
            current = name
        print(f"    range .{coll}: {what}")
    if not args.quiet:
        print("\nThese are CANDIDATES, not confirmed defects: input_schema has no per-item\n"
              "field vocabulary, so this check cannot tell a required headline slot from an\n"
              "optional read-time one. Confirm against what the component's producer writes.\n")
        print("Each renders an EMPTY element when its producer does not write the key —\n"
              "empty, but still occupying layout. Fix per bugs_open/425 / migration 682:\n"
              "  {{if .excerpt}}<p class=\"...\">{{.excerpt}}</p>{{end}}\n"
              "and guard a wrapper whose children are all optional as a pair:\n"
              "  {{if or .date .read_time}}<div class=\"...__meta\">...</div>{{end}}\n"
              "Then ask whether the PRODUCER should be writing the key at all — a\n"
              "collapsed slot hides the gap; it does not fill it.")
    return 1


if __name__ == "__main__":
    sys.exit(main())
