#!/usr/bin/env python3
"""Static checks on the gauntlet round-record component.

Runs against a LOCAL file by default, or against the SERVED page with --url.
Both matter and they are not the same claim: the local file is what I wrote,
the served page is what a visitor gets, and a delivery can lose the difference.

    ./verify_round_record.py round_record_component.html
    ./verify_round_record.py --url https://vonc.com/tools/gauntlet/round.html

Every check prints PASS or FAIL and the script exits non-zero if any failed, so
a truncated run cannot be mistaken for a clean one. There is a deliberate
positive control (textContent must be PRESENT): a check suite that only ever
asserts absence passes just as happily against an empty file.
"""
import re
import subprocess
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parents[4].parent
GO_STORE = REPO / "internal/tools-api/store/rounds.go"

fails = []


def check(name, ok, detail=""):
    print(("PASS  " if ok else "FAIL  ") + name + (("  — " + detail) if detail else ""))
    if not ok:
        fails.append(name)


def main():
    args = [a for a in sys.argv[1:]]
    if args and args[0] == "--url":
        url = args[1]
        html = subprocess.run(
            ["curl", "-sS", "-A", "Mozilla/5.0 Chrome/126", url],
            capture_output=True, text=True, check=True).stdout
        src = "served " + url
    else:
        p = Path(args[0]) if args else Path(__file__).with_name("round_record_component.html")
        html = p.read_text()
        src = str(p)

    print("checking: " + src + "  (" + str(len(html)) + " bytes)\n")

    # 1. no template syntax anywhere, INCLUDING inside comments. A malformed
    #    action in a comment still fails the parse and silently drops the
    #    renderer to a regex engine.
    check("no template placeholders", html.count("{{") == 0,
          "found " + str(html.count("{{")))

    # 2. the one rule about untrusted text. position/defence are public prose
    #    typed by a stranger; nothing on this page may assign markup.
    for prop in ("innerHTML", "outerHTML", "insertAdjacentHTML", "document.write"):
        check("no " + prop, html.count(prop) == 0, "found " + str(html.count(prop)))

    # 3. POSITIVE CONTROL — absence checks alone would pass on an empty file.
    check("textContent is used (control)", html.count("textContent") >= 2,
          "found " + str(html.count("textContent")))

    # 4. the script parses as balanced. Not a substitute for a real parse, but
    #    it catches a truncated delivery, which is the failure this lane has
    #    actually had (a 10,272-char component saved back as 1,253).
    m = re.search(r"<script>(.*)</script>", html, re.S)
    check("script block present", m is not None)
    if m:
        js = m.group(1)
        check("script braces balance", js.count("{") == js.count("}"),
              str(js.count("{")) + " open vs " + str(js.count("}")) + " close")
        check("script parens balance", js.count("(") == js.count(")"),
              str(js.count("(")) + " open vs " + str(js.count(")")) + " close")

        # 5. markup hooks and queries must agree in both directions. A renamed
        #    attribute leaves a silently blank field otherwise.
        markup = html[: html.rindex("<script>")]
        attrs = set(re.findall(r"\bdata-gr-[a-z-]+", markup))
        queried = set(re.findall(r"\[(data-gr-[a-z-]+)\]", js))
        check("no unqueried markup hooks", not (attrs - queried), str(sorted(attrs - queried)))
        check("no queries without markup", not (queried - attrs), str(sorted(queried - attrs)))

        # 6. the page must talk to the island, not to vonc.com.
        check("island API base", "https://tools.apis.uk" in js)

        # 7. ANTI-DRIFT ACROSS LANGUAGES. The slug alphabet and length are
        #    declared once in Go and mirrored here; if the Go constant changes
        #    and this does not, the page rejects addresses the server issues.
        if GO_STORE.exists():
            go = GO_STORE.read_text()
            ga = re.search(r'SlugAlphabet\s*=\s*"([^"]+)"', go)
            gl = re.search(r"SlugLength\s*=\s*(\d+)", go)
            ja = re.search(r'SLUG_ALPHABET\s*=\s*"([^"]+)"', js)
            jl = re.search(r"SLUG_LENGTH\s*=\s*(\d+)", js)
            check("slug alphabet matches Go", bool(ga and ja) and ga.group(1) == ja.group(1),
                  "go=" + (ga.group(1) if ga else "?") + " js=" + (ja.group(1) if ja else "?"))
            check("slug length matches Go", bool(gl and jl) and gl.group(1) == jl.group(1),
                  "go=" + (gl.group(1) if gl else "?") + " js=" + (jl.group(1) if jl else "?"))
        else:
            check("Go store readable for slug cross-check", False, str(GO_STORE))

        # 8. the provocation's marketing fields must never be rendered. `stats`
        #    is an array of numbers and this page computes nothing.
        for banned in ("stats", "primary_cta", "secondary_cta"):
            check("provocation." + banned + " not rendered",
                  ("prov." + banned) not in js and ('"' + banned + '"') not in js)

    print()
    if fails:
        print("FAILED " + str(len(fails)) + ": " + ", ".join(fails))
        return 1
    print("all checks passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
