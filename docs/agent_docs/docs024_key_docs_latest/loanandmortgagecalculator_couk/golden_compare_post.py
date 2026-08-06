#!/usr/bin/env python3
"""golden_compare_post.py — compare a DECOMPOSED calculator page against the
pre-change golden, allowing exactly one known structural difference and
nothing else.

WHY NOT PLAIN `toolgolden --compare`. The golden's fingerprint is "the text
and display of every id-bearing element". On the hand-built pages the whole
content wrapper carried `id="content"` (the skip-link target), so its field
holds the entire page text at display:block. After decomposition that wrapper
is gone and the skip-link target lives in the site header as an EMPTY
`<span id="content" tabindex="-1">` — inline, no text. So every decomposed
calculator diverges on `content` and on nothing else, and a raw --compare
reports a red result for a page whose arithmetic is untouched. Measured on
loans/consolidation (canary 2, 2026-08-05): 11 of 12 fields byte-equal
including every money field and the verdict string; `content` the only
difference.

WHAT THIS ASSERTS INSTEAD — narrower than the raw compare, not looser:
  1. every field EXCEPT `content` matches the golden exactly, in all three
     vectors and both phases (this is the arithmetic);
  2. `content` on the new page is EMPTY and `inline` — i.e. it really is the
     chrome span, not a wrapper that lost its text. A page whose content
     field is non-empty but different FAILS, because that would mean prose
     changed inside a live wrapper rather than the wrapper being replaced.

The allowance is therefore a shape assertion, not an ignore-list: the one
value `content` is permitted to take is the one the new chrome produces.

POSITIVE CONTROL (run it, do not assume): --self-test compares one tool's
golden against a DIFFERENT tool's live page and must report FAIL. Without
that, a comparator that silently matched nothing would pass everything.

Usage (PYTHONPATH must include webdesign_tools_repair and the sibling lane):
  python3 golden_compare_post.py <golden.json> <url> [<url> ...]
  python3 golden_compare_post.py --self-test <golden.json>
"""
import json
import sys

from toolgolden import Runner  # noqa: E402  (sibling lane, via PYTHONPATH)

CHROME_FIELD = "content"
PHASES = ("before", "after_input", "after_press")


def compare_one(golden_page, live_page, url, allow_chrome=True):
    problems = []
    for vec in golden_page:
        if vec not in live_page:
            problems.append("vector %s missing from live capture" % vec)
            continue
        for phase in PHASES:
            want = golden_page[vec].get(phase, {}).get("ids", {})
            got = live_page[vec].get(phase, {}).get("ids", {})
            for field, wval in want.items():
                gval = got.get(field, "<ABSENT>")
                if field == CHROME_FIELD and allow_chrome:
                    # the ONE permitted shape: empty text, inline display
                    if gval not in ("|inline", "<ABSENT>"):
                        problems.append(
                            "%s/%s: %s is %r — expected the empty chrome span "
                            "('|inline'). A non-empty, changed content field "
                            "means prose moved inside a live wrapper."
                            % (vec, phase, field, gval))
                    continue
                if gval != wval:
                    problems.append("%s/%s: %s  %r -> %r"
                                    % (vec, phase, field, wval, gval))
    return problems


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    self_test = "--self-test" in sys.argv
    if not args:
        print(__doc__)
        return 2
    golden = json.load(open(args[0], encoding="utf-8"))["pages"]

    if self_test:
        keys = sorted(golden)
        if len(keys) < 2:
            print("self-test needs a golden with >= 2 tools")
            return 2
        a, b = keys[0], keys[1]
        r = Runner()
        try:
            live_b = r.capture(b)
        finally:
            r.close()
        problems = compare_one(golden[a], live_b, b)
        print("SELF-TEST: %s's golden vs %s's live page -> %d problem(s)"
              % (a.rsplit('/', 1)[-1], b.rsplit('/', 1)[-1], len(problems)))
        for p in problems[:4]:
            print("   " + p)
        if not problems:
            print("SELF-TEST FAILED: comparator found no difference between "
                  "two DIFFERENT tools — it is inert, do not trust a pass.")
            return 1
        print("SELF-TEST PASSED: the comparator can still fail.")
        return 0

    urls = args[1:]
    if not urls:
        print("name at least one url")
        return 2
    r = Runner()
    bad = 0
    try:
        for url in urls:
            if url not in golden:
                print("NO GOLDEN  %s" % url)
                bad += 1
                continue
            problems = compare_one(golden[url], r.capture(url), url)
            if problems:
                bad += 1
                print("DIVERGED   %s" % url)
                for p in problems:
                    print("   " + p)
            else:
                print("MATCHES    %s  (arithmetic exact; chrome span as expected)"
                      % url)
    finally:
        r.close()
    print("\n%d of %d tool(s) diverged" % (bad, len(urls)))
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
