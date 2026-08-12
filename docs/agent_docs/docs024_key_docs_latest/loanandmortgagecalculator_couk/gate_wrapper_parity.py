#!/usr/bin/env python3
"""gate_wrapper_parity.py — refuse a manifest that drops a layout class.

WHY THIS EXISTS (`bugs_open/263`). Decomposition dissolved the calculator's own
presentation wrapper on 12 of 21 pages — `.card` (background, 30px padding, radius,
shadow), `.calc-grid` / `.input-grid` (`grid-template-columns: 1fr 1fr`), `.fact-grid`.
Two live calculator pages shipped with their panel gone and their inputs stacked before
anyone noticed, and the two checks that were supposed to notice both passed:

  * `deploy_pages.py`'s byte diff compares the served page against a prediction built
    from THE SAME MANIFEST that dropped the wrapper. It proves fidelity to the model,
    never preservation of the original — it is structurally incapable of catching this.
  * a class-SET comparison saw nothing, because `loans/standard-calc` carries four
    `.card`s: the calculator's was dissolved and three prose cards survived. The
    aggregate attribute count was 18 before and 18 after, two losses exactly offset by
    two `ported-prose` additions.

So the only check that works is a PER-CLASS COUNT, and it must run BEFORE the rows are
written — which is why this reads the manifest rather than a deployed page.

WHAT IT ASSERTS. For each page in the manifest, every class in the source's `#content`
subtree must appear at least as many times across the manifest's blocks. Losing one of
four `.card`s is a failure; a set comparison would call that clean.

PERMITTED LOSSES, and each one is compensated somewhere real:
  container / container-tight  — decomposition dissolves the single page wrapper by
                                 design and the assembled-layout shim makes <main> the
                                 container (PLAN_2026-08-05). `container-tight`'s 740px
                                 box is re-wrapped by load_lmc.py.
Nothing else is permitted. If a page legitimately needs another exemption, add it HERE
with the compensation named, so the next reader can check the claim.

Run:  python3 gate_wrapper_parity.py <manifest.json>
      python3 gate_wrapper_parity.py <manifest.json> --self-test   # MUST fail
"""
import json
import os
import re
import subprocess
import sys

SITE_DIR = "loanandmortgagecalculator.co.uk"
SITES_REPO = os.path.expanduser("~/projects/sites")
PERMITTED = {"container", "container-tight"}


def class_counts(html):
    c = {}
    for m in re.findall(r'class="([^"]+)"', html):
        for x in m.split():
            c[x] = c.get(x, 0) + 1
    return c


def source_content(ref, relpath):
    """The #content subtree of the pinned source — the same slice decomposition reads."""
    r = subprocess.run(["git", "show", "%s:%s/%s" % (ref, SITE_DIR, relpath)],
                       capture_output=True, text=True, cwd=SITES_REPO)
    if r.returncode != 0:
        sys.exit("git show failed for %s: %s" % (relpath, r.stderr.strip()[:200]))
    html = r.stdout
    m = re.search(r'<div id="content"[^>]*>', html)
    if not m:
        return None
    i = m.end()
    depth, j = 1, i
    while depth > 0:
        nm = re.search(r"<(/?)div\b[^>]*>", html[j:])
        if not nm:
            break
        depth += -1 if nm.group(1) else 1
        j += nm.end()
    return html[i:j]


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    path = sys.argv[1]
    self_test = "--self-test" in sys.argv
    man = json.load(open(path))
    ref, pages = man["pinned_ref"], man["pages"]

    failures, checked = 0, 0
    for name, entry in sorted(pages.items()):
        relpath = entry.get("path") or entry.get("relpath")
        if not relpath:
            relpath = ("index.html" if name == "index"
                       else name.replace("-", "/", 1) + ".html")
        src = source_content(ref, relpath)
        if src is None:
            print("SKIP %-38s no div#content at the pin (already decomposed?)" % name)
            continue
        want = class_counts(src)
        got = class_counts("".join(b["html"] for b in entry["blocks"]))
        if self_test:
            # Pretend the source carried one more of a class the page really has, so a
            # correct manifest MUST come out short. Uses the page's own vocabulary, so
            # the control cannot pass by naming something that never existed.
            k = next((k for k in want if k not in PERMITTED), None)
            if k:
                want = dict(want, **{k: want[k] + 1})
        drops = {k: (v, got.get(k, 0)) for k, v in want.items()
                 if k not in PERMITTED and got.get(k, 0) < v}
        checked += 1
        if drops:
            failures += 1
            print("FAIL %-38s %s" % (name, drops))
        else:
            print("ok   %-38s %d classes preserved" % (name, len(want)))

    print("\n%d page(s) checked, %d failing." % (checked, failures))
    if self_test:
        if failures != checked:
            sys.exit("CONTROL FAILED: %d of %d pages passed with an induced shortfall. "
                     "The gate is not comparing counts." % (checked - failures, checked))
        print("CONTROL OK: every page failed under an induced one-class shortfall.")
        return 0
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
