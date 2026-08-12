#!/usr/bin/env python3
"""class_count_delta.py — PER-CLASS count diff between two HTML documents.

The one measurement that catches a decomposition or a rewrite silently deleting a
page's design, and the reason it has to be per-class rather than any of the three
cheaper things people reach for first. All three were tried on the same page, on
the same day, and all three passed it (bugs_open/263, WRONG_CALLS.md 2026-08-12):

  a class SET diff        loans-standard-calc carries four `card`s. The calculator's
                          was dissolved and three prose cards survived, so the SET
                          was unchanged. Reported as proven.
  an AGGREGATE count      18 class attributes before, 18 after — two removals
                          (`container`, one `card`) exactly offset by two
                          `ported-prose` additions. This is the shape that makes an
                          aggregate floor unsafe: offsetting additions hide drops.
  a byte-for-byte diff    against a prediction generated from the same manifest that
                          dropped the wrapper. It proves fidelity to the model, not
                          preservation of the original, and for this class of defect
                          it is structurally incapable of failing.

Only `card: 4 -> 3` says it. So: count each class NAME separately, and compare.

WHAT COUNTS AS PERMITTED. A permitted delta is one for which a NAMED, LIVE
COMPENSATING RULE exists — not one that looked harmless. On the decomposed
calculator sites exactly one qualifies: the page's own `.container` is dissolved
by the descent and `chrome/head.html` compensates by styling `<main>` as the
container (`main .container { max-width:none; margin:0; padding:0 }`). Everything
else that disappears is uncompensated design loss. The allowlist deliberately does
NOT live in this module: policy is per-lane, and a shared allowlist means the first
lane to widen it weakens the gate for every other lane. Same reasoning as
scripts/truncation_registry.py, pointing the same way.

TWO TRAPS THIS MODULE EXISTS TO NOT FALL INTO, both measured on real pages:

  * `class=` INSIDE A <script> BODY IS NOT MARKUP. loancalculator.co.uk's
    tools/consolidation.html builds rows in a JS template literal, so `d-bal`,
    `d-months`, `d-name`, `d-rate` and `remove-btn` each appear once in real markup
    and once in JS. Decomposition drops the inline script, so a naive count reports
    five drops that are not drops — and the natural fix (allowlist them) punches a
    real hole. Worse, the same mechanism can HIDE a real drop: a JS template
    emitting `class="card"` offsets a dissolved `.card` exactly the way
    `ported-prose` did in 263. So both sides are stripped of <script>, <style> and
    comments before counting.
  * AN EMPTY "AFTER" READS AS TOTAL LOSS, AND AN EMPTY "BEFORE" READS AS A PASS.
    A scope regex that fails to match, or a `git show` from the wrong directory,
    yields an empty string and a confident number. Both sides are asserted
    non-empty and the scope is asserted to have matched; the failure is `NOT
    CHECKED` (exit 2), never a pass. This module was written after exactly that
    read produced a "0 occurrences" answer during its own planning.

Exit 0 = no unpermitted deltas, 1 = findings, 2 = could not measure (unreachable
URL, unreadable file, empty scope, ambiguous markup). Exit 2 is deliberately
distinct: "the check did not run" must never be mistaken for "the check passed".

Usage:
  class_count_delta.py --before A.html --after B.html [--permit container=-1] ...
  class_count_delta.py --before A.html --after https://site/page --scope main \\
                       --permit container=-1 --permit ported-prose=+ --allow-additions
  class_count_delta.py --before A.html --after B.html --permit-none   # rollback oracle
  class_count_delta.py --selftest

Importers: pass HTML strings to check() and keep your allowlist at your own call
site. `counts`, `strip_code`, `delta` and `check` are the whole surface.
"""

import argparse
import re
import sys
import urllib.request

CODE_RE = re.compile(r"<(script|style)\b.*?</\1\s*>", re.S | re.I)
COMMENT_RE = re.compile(r"<!--.*?-->", re.S)
CLASS_RE = re.compile(r'\bclass\s*=\s*"([^"]*)"', re.I)
# Anything spelled another way is a silent under-count, so it is an error, not a
# best effort. Verified zero occurrences across both calculator sites 2026-08-12.
ODD_CLASS_RE = re.compile(r"""\bclass\s*=\s*(?!")""", re.I)


class Unmeasurable(Exception):
    """The check could not run. Never report this as a pass."""


def strip_code(html):
    """Remove <script>, <style> and comments — their text is not page markup."""
    return COMMENT_RE.sub(" ", CODE_RE.sub(" ", html))


def counts(html):
    """{class name: occurrences} over real markup only."""
    stripped = strip_code(html)
    odd = ODD_CLASS_RE.search(stripped)
    if odd:
        raise Unmeasurable(
            "class attribute is not double-quoted at offset %d (%r) — this counter "
            "would silently under-count it on BOTH sides, which is invisible. Widen "
            "CLASS_RE deliberately rather than let it read low."
            % (odd.start(), stripped[odd.start():odd.start() + 40]))
    out = {}
    for value in CLASS_RE.findall(stripped):
        for name in value.split():
            out[name] = out.get(name, 0) + 1
    return out


def delta(before, after):
    """{class: (before, after)} for every class whose count changed."""
    return {c: (before.get(c, 0), after.get(c, 0))
            for c in set(before) | set(after)
            if before.get(c, 0) != after.get(c, 0)}


def scope_to(html, tag):
    """Narrow to <tag>…</tag>. Refuses rather than returning an empty scope."""
    m = re.search(r"<%s\b[^>]*>(.*)</%s\s*>" % (tag, tag), html, re.S | re.I)
    if not m or not m.group(1).strip():
        raise Unmeasurable(
            "scope <%s> did not match or is empty — every class would read as "
            "dropped (or as fine, depending on which side it is). NOT CHECKED." % tag)
    return m.group(1)


def check(before_html, after_html, permitted=None, allow_unlisted_additions=False):
    """Findings, one string per class whose change is not permitted.

    `permitted` maps a class name to either an int (the EXACT delta allowed, e.g.
    container=-1) or "+" (any number of additions, no drops). With
    allow_unlisted_additions=False an unlisted addition is itself a finding — that
    is the strict form, and it is the one that closes the netting-out hole by
    construction at a seam where every emitted block is a byte slice of the source
    and additions are therefore impossible.
    """
    if not (before_html or "").strip():
        raise Unmeasurable("the BEFORE document is empty — nothing to compare against")
    if not (after_html or "").strip():
        raise Unmeasurable("the AFTER document is empty — this would read as total loss")
    permitted = permitted or {}
    b, a = counts(before_html), counts(after_html)
    findings = []
    for cls, (was, now) in sorted(delta(b, a).items()):
        d = now - was
        spec = permitted.get(cls)
        if spec == "+":
            if d < 0:
                findings.append("%s: %d -> %d (%+d) — permitted to be ADDED, not dropped"
                                % (cls, was, now, d))
            continue
        if isinstance(spec, int):
            if d != spec:
                findings.append("%s: %d -> %d (%+d) — permitted delta is exactly %+d"
                                % (cls, was, now, d, spec))
            continue
        if d > 0 and allow_unlisted_additions:
            continue
        findings.append("%s: %d -> %d (%+d) — %s"
                        % (cls, was, now, d,
                           "no compensating rule is named for this class"
                           if d < 0 else "unexpected addition"))
    return findings


def _read(source):
    if source.startswith(("http://", "https://")):
        try:
            with urllib.request.urlopen(source, timeout=30) as fh:
                return fh.read().decode("utf-8", "replace")
        except Exception as exc:                       # noqa: BLE001 — reported, not raised
            raise Unmeasurable("could not fetch %s: %s" % (source, exc))
    try:
        with open(source, encoding="utf-8", errors="replace") as fh:
            body = fh.read()
    except OSError as exc:
        raise Unmeasurable("could not read %s: %s" % (source, exc))
    if not body.strip():
        raise Unmeasurable(
            "%s is EMPTY. A `git show <ref>:<path>` run from a subdirectory returns "
            "empty on stdout and its error on stderr, so this reads as a clean pass. "
            "Paths are repo-root-relative." % source)
    return body


def _parse_permits(pairs):
    out = {}
    for p in pairs:
        if "=" not in p:
            raise SystemExit("--permit takes name=delta, e.g. container=-1 or ported-prose=+")
        name, value = p.split("=", 1)
        out[name] = "+" if value.strip() == "+" else int(value)
    return out


def selftest():
    doc = '<div class="container"><div class="card"><div class="calc-grid">' \
          '<div class="form-group"></div></div></div></div>'
    assert counts(doc) == {"container": 1, "card": 1, "calc-grid": 1, "form-group": 1}
    # identical input must be silent, or every later zero is meaningless
    assert check(doc, doc) == [], "identical documents reported a delta"
    # INDUCED FAILURE: a check that cannot fail is not a check
    dissolved = doc.replace('<div class="card">', "").replace("</div></div></div>", "</div></div>")
    assert check(doc, dissolved), "a dissolved .card was NOT reported — the gate is inert"
    # the netting-out shape 263 was reported clean on: aggregate 2 -> 2, per-class sees it
    netted = '<section class="ported-prose"></section><div class="card"></div>' \
             '<div class="calc-grid"></div><div class="form-group"></div>'
    assert len(counts(doc)) == 4 and sum(counts(netted).values()) == 4
    assert any(f.startswith("container:") for f in check(doc, netted)), \
        "the offset-by-additions case was not reported"
    # a class inside a <script> body is not markup
    assert counts('<div class="a"></div><script>x.innerHTML=\'<i class="a">\'</script>') == {"a": 1}
    # additions: permitted only when named, unless explicitly allowed
    assert check(doc, doc + '<section class="ported-prose"></section>') != []
    assert check(doc, doc + '<section class="ported-prose"></section>',
                 {"ported-prose": "+"}) == []
    # exact-delta specs
    assert check(doc, doc.replace(' class="container"', "")) != []
    assert check(doc, doc.replace(' class="container"', ""), {"container": -1}) == []
    # unmeasurable inputs must raise, never return []
    for bad in ("", "   "):
        try:
            check(bad, doc)
        except Unmeasurable:
            pass
        else:
            raise AssertionError("an empty BEFORE was accepted")
    try:
        counts("<div class='x'></div>")
    except Unmeasurable:
        pass
    else:
        raise AssertionError("a single-quoted class attribute was silently under-counted")
    print("selftest: OK (including the induced failure and the netting-out case)")
    return 0


def main():
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--before")
    ap.add_argument("--after", help="file path or http(s) URL")
    ap.add_argument("--scope", help="narrow BOTH sides to this tag, e.g. main")
    ap.add_argument("--scope-after", help="narrow only the AFTER side to this tag")
    ap.add_argument("--permit", action="append", default=[],
                    metavar="NAME=DELTA", help="container=-1, ported-prose=+")
    ap.add_argument("--permit-none", action="store_true",
                    help="no delta of any kind is permitted (the rollback oracle)")
    ap.add_argument("--allow-additions", action="store_true",
                    help="unlisted ADDITIONS are not findings (use at an assembling seam)")
    ap.add_argument("--selftest", action="store_true")
    args = ap.parse_args()

    if args.selftest:
        return selftest()
    if not args.before or not args.after:
        ap.error("--before and --after are required")

    permitted = {} if args.permit_none else _parse_permits(args.permit)
    allow_add = args.allow_additions and not args.permit_none
    try:
        before, after = _read(args.before), _read(args.after)
        if args.scope:
            before, after = scope_to(before, args.scope), scope_to(after, args.scope)
        elif args.scope_after:
            after = scope_to(after, args.scope_after)
        findings = check(before, after, permitted, allow_add)
    except Unmeasurable as exc:
        print("NOT CHECKED: %s" % exc, file=sys.stderr)
        return 2

    if not findings:
        print("class counts: OK (%s)"
              % ("no delta of any kind" if args.permit_none
                 else "no unpermitted delta; permitted %s" % (permitted or "{}")))
        return 0
    print("UNPERMITTED CLASS DELTAS (%d):" % len(findings))
    for f in findings:
        print("  " + f)
    return 1


if __name__ == "__main__":
    sys.exit(main())
