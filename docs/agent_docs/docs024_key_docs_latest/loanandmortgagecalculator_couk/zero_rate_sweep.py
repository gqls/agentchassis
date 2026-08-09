#!/usr/bin/env python3
"""zero_rate_sweep.py — find the `bugs_open/224` defect class on ANY calculator site.

WHY A SECOND TOOL, WHEN `oracle.py` ALREADY FINDS THIS.
`oracle.py` needs a per-tool spec: which control takes the principal, which
element holds the answer, what the right answer is. That is what makes it able
to say "£143.47 should be £166.67" — and it is also what makes it cost an hour
per site. This file needs NO spec and no oracle at all, because both failure
modes of the zero-rate defect are visible without knowing what the right answer
is:

  MODE 1  the tool prints `NaN` or `Infinity`. Self-evident — a money field
          containing "NaN" is wrong whatever the tool is for.
  MODE 2  the tool writes nothing and leaves the PREVIOUS answer on screen. Not
          self-evident from one reading, but decidable without an oracle: drive
          the same final inputs by two different routes and compare. The two
          readings must agree whatever the tool computes, so a disagreement is a
          defect with no reference to correctness at all.

So this is the cheap fleet-wide sweep and `oracle.py` is the expensive per-site
proof. Use this to find out WHICH sites are affected; use that to fix one.

WHAT IT DRIVES. Numeric inputs are classified from their VISIBLE LABEL, not from
the page's script (same discipline as `inventory.py`, and the same reason): a
field whose label mentions rate/APR/AER/interest/% is a rate field and is set to
0; every other numeric field keeps its own default, which is in-domain by
construction. Buttons matching the reset pattern are excluded — clicking
"Start Over" mid-walk destroys the state you just established, a lesson already
paid for in `toolgolden.PRESS_JS` and re-paid in `invariants.py`.

Usage:
  python3 zero_rate_sweep.py https://loancalculator.co.uk/ tools/compare-loans.html ...
  python3 zero_rate_sweep.py --from-repo /home/ant/projects/sites/loancalculator.co.uk
  python3 zero_rate_sweep.py --from-repo <dir> --json out.json
"""
import argparse
import json
import os
import sys

from oracle_driver import Driver, DriveError, PageLoadError

RESET_RE = r"/reset|clear|start over|start again|cancel|close|download|backup/i"

# Set every rate-ish numeric field to `rateval`, leave the rest at their own
# defaults, and report what was driven so a null result can be told from a
# "there was nothing to drive" result.
DRIVE_JS = r"""
(rateval) => {
  const norm = s => (s || '').replace(/\s+/g, ' ').trim().slice(0, 90);
  const inMain = e => !e.closest('nav,header,footer,[id*=nav],[class*=nav],[id*=menu],[class*=menu]');
  const labelFor = e => {
    if (e.id) {
      const l = document.querySelector('label[for="' + CSS.escape(e.id) + '"]');
      if (l) return norm(l.textContent);
    }
    const anc = e.closest('label');
    if (anc) return norm(anc.textContent);
    const labs = Array.from(document.querySelectorAll('label'));
    let best = null;
    for (const l of labs) {
      if (l.compareDocumentPosition(e) & Node.DOCUMENT_POSITION_FOLLOWING) best = l;
    }
    return best ? norm(best.textContent) : norm(e.placeholder || '');
  };
  const RATE = /rate|apr|aer|interest|%/i;
  const fire = e => ['input','change'].forEach(
    ev => e.dispatchEvent(new Event(ev, {bubbles:true})));

  const driven = [];
  document.querySelectorAll('input[type=number],input[type=range]').forEach(e => {
    if (e.disabled || !inMain(e)) return;
    const lab = labelFor(e);
    const isRate = RATE.test(lab) || RATE.test(e.id || '') || RATE.test(e.name || '');
    if (isRate) {
      e.value = String(rateval); fire(e);
      driven.push({sel: e.id ? '#' + e.id : null, label: lab, set: rateval, rate: true});
    } else {
      // Keep the field's own default; a blank field is not a vector, so fill
      // an obviously in-domain value only where there is no default at all.
      if (e.value === '') { e.value = '1000'; fire(e); }
      driven.push({sel: e.id ? '#' + e.id : null, label: lab, set: e.value, rate: false});
    }
  });
  return JSON.stringify(driven);
}
"""

PRESS_JS = r"""
(() => {
  const all = Array.from(document.querySelectorAll(
                'button,input[type=button],input[type=submit]'))
    .filter(e => !e.disabled)
    .filter(e => !e.closest('nav,header,footer,[id*=nav],[class*=nav],[id*=menu],[class*=menu]'))
    .filter(e => !%s.test((e.textContent || e.value || '') + ' ' + (e.id || '')));
  const b = all.find(e => e.getAttribute('onclick') || e.onclick) || all[0];
  if (!b) return 'none';
  b.click();
  return b.id || (b.textContent || b.value || '').trim().slice(0, 30);
})()
""" % RESET_RE

READ_JS = r"""
(() => {
  const norm = s => (s || '').replace(/\s+/g, ' ').trim().slice(0, 120);
  const inMain = e => !e.closest('nav,header,footer,[id*=nav],[class*=nav],[id*=menu],[class*=menu]');
  const out = {};
  document.querySelectorAll('[id]').forEach(e => {
    if (!inMain(e) || e.children.length > 2) return;
    if (e.matches('input,select,textarea,button,form,script,style,link')) return;
    const t = norm(e.textContent);
    if (!t) return;
    // Only elements that look like they hold a computed value.
    if (!/[£%]|\d/.test(t)) return;
    out[e.id] = t;
  });
  return JSON.stringify(out);
})()
"""


def sweep_page(d, url):
    """Return {nan: {...}, stale: {...}, driven: [...], pressed: str}."""
    # Route A: straight to a 0% rate from the page's own defaults.
    d.goto(url)
    driven = json.loads(d.ev("(%s)(%s)" % (DRIVE_JS.strip(), 0)) or "[]")
    pressed_a = d.ev(PRESS_JS, timeout=30)
    read_a = json.loads(d.ev(READ_JS) or "{}")

    # Route B: the SAME final inputs, reached via a high rate first. Any
    # difference means the output depends on input history, i.e. the tool did
    # not recompute — mode 2, detected with no oracle.
    d.goto(url)
    d.ev("(%s)(%s)" % (DRIVE_JS.strip(), 17.5))
    d.ev(PRESS_JS, timeout=30)
    d.ev("(%s)(%s)" % (DRIVE_JS.strip(), 0))
    d.ev(PRESS_JS, timeout=30)
    read_b = json.loads(d.ev(READ_JS) or "{}")

    nan = {k: v for k, v in read_a.items()
           if "NaN" in v or "Infinity" in v or "undefined" in v}
    stale = {k: (read_a.get(k), read_b.get(k))
             for k in set(read_a) | set(read_b)
             if read_a.get(k) != read_b.get(k)}
    return {"nan": nan, "stale": stale, "driven": driven,
            "pressed": pressed_a,
            "rate_fields": sum(1 for x in driven if x.get("rate"))}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("base", nargs="?", help="e.g. https://loancalculator.co.uk/")
    ap.add_argument("pages", nargs="*")
    ap.add_argument("--from-repo", help="local site dir; sweeps its live twin")
    ap.add_argument("--json")
    a = ap.parse_args()

    base, pages = a.base, list(a.pages)
    if a.from_repo:
        domain = os.path.basename(a.from_repo.rstrip("/"))
        base = base or "https://%s/" % domain
        for root, _dirs, files in os.walk(a.from_repo):
            for f in files:
                if not f.endswith(".html"):
                    continue
                rel = os.path.relpath(os.path.join(root, f), a.from_repo)
                # Only pages that actually compute something.
                src = open(os.path.join(root, f), errors="replace").read()
                if 'type="number"' in src or "type='number'" in src:
                    pages.append(rel)
    if not base or not pages:
        ap.error("need a base URL and at least one page (or --from-repo)")
    pages = sorted(set(pages))

    d = Driver()
    results, affected = {}, []
    try:
        for p in pages:
            url = base.rstrip("/") + "/" + p
            try:
                r = sweep_page(d, url)
            except (PageLoadError, DriveError) as e:
                results[p] = {"error": str(e)}
                print("  %-44s ERROR %s" % (p[:44], str(e)[:60]))
                continue
            results[p] = r
            bad = bool(r["nan"]) or bool(r["stale"])
            if bad:
                affected.append(p)
            print("  [%s] %-44s rate-fields:%d  NaN:%d  history-dependent:%d"
                  % ("FAIL" if bad else " ok ", p[:44], r["rate_fields"],
                     len(r["nan"]), len(r["stale"])))
            for k, v in list(r["nan"].items())[:4]:
                print("         NaN   #%-22s %s" % (k, v[:60]))
            for k, (x, y) in list(r["stale"].items())[:4]:
                print("         HIST  #%-22s %r vs %r" % (k, (x or "")[:26], (y or "")[:26]))
    finally:
        d.close()

    print("\n%s\n%d of %d pages affected%s" %
          ("-" * 72, len(affected), len(pages),
           (": " + ", ".join(affected)) if affected else ""))
    if a.json:
        with open(a.json, "w") as f:
            json.dump({"base": base, "affected": affected, "results": results},
                      f, indent=1)
        print("wrote %s" % a.json)
    return 1 if affected else 0


if __name__ == "__main__":
    sys.exit(main())
