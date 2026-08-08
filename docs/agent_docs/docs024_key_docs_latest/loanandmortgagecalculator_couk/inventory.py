#!/usr/bin/env python3
"""inventory.py — dump each calculator's USER-FACING interface, from the live page.

WHY THIS EXISTS, AND WHY IT READS LABELS RATHER THAN SCRIPT.
The deliverable is an INDEPENDENT oracle: expected answers recomputed from the
published definition (the annuity formula, the HMRC bands), never transcribed
from the page's own `<script>`. But an oracle still has to know WHICH box to
type the principal into and WHICH element holds the monthly payment, and the
obvious way to learn that — read the page's JS and see what `calculateLoan`
touches — is exactly the contamination the brief forbids. Reading the source to
find `getElementById('rate')` is one keystroke away from reading the line below
it that says what `rate` is divided by.

So this reads the page the way a USER does: the visible <label> attached to each
control, the button's visible text, and the caption above each result box. That
is a genuinely independent channel — it is the site's own claim about what each
number means, and if the label and the arithmetic disagree, that disagreement is
itself the finding rather than something the harness silently absorbs.

It also records `value`/`min`/`max`/`step` so a boundary vector can be checked
against the field's declared domain before it is used: a vector the browser
would clamp is not the vector you think you drove.

Usage:
  python3 inventory.py --out inventory.json          # all 23 tool pages
  python3 inventory.py --out x.json mortgages/stamp-duty.html
"""
import argparse
import json
import os
import sys

from oracle_driver import Driver, TOOL_PAGES, BASE


# Label resolution, in the order a sighted user resolves it:
#   1. <label for=id>                      the explicit binding
#   2. an ancestor <label>                 the wrapping form
#   3. the nearest preceding <label>       these pages' actual house style
#   4. aria-label / placeholder            the accessibility fallback
# Recorded WITH which rule fired, because "rule 3 fired" is a much weaker claim
# about meaning than "rule 1 fired" and the report should be able to say so.
INVENTORY_JS = r"""
(() => {
  const norm = s => (s || '').replace(/\s+/g, ' ').trim().slice(0, 120);
  const labelFor = e => {
    if (e.id) {
      const l = document.querySelector('label[for="' + CSS.escape(e.id) + '"]');
      if (l) return [norm(l.textContent), 'for'];
    }
    const anc = e.closest('label');
    if (anc) return [norm(anc.textContent), 'ancestor'];
    // Nearest label PRECEDING this control in document order.
    //
    // The first version took `closest('div').querySelector('label')`, i.e. the
    // FIRST label in the enclosing div — which on stamp-duty.html labelled the
    // buyer-type <select> as "Property Purchase Price", because both controls
    // sit in one wrapper. A caption that names the wrong field is worse than a
    // blank one: it is what the oracle's per-tool spec gets authored from.
    const labs = Array.from(document.querySelectorAll('label'));
    let best = null;
    for (const l of labs) {
      if (l.compareDocumentPosition(e) & Node.DOCUMENT_POSITION_FOLLOWING) best = l;
    }
    if (best) return [norm(best.textContent), 'preceding'];
    if (e.getAttribute('aria-label')) return [norm(e.getAttribute('aria-label')), 'aria'];
    if (e.placeholder) return [norm(e.placeholder), 'placeholder'];
    return ['', 'none'];
  };
  const inMain = e => !e.closest('nav,header,footer,[id*=nav],[class*=nav],[id*=menu],[class*=menu]');

  const inputs = [];
  document.querySelectorAll('input,select,textarea').forEach(e => {
    if (!inMain(e)) return;
    const [lab, how] = labelFor(e);
    const rec = {id: e.id || null, name: e.name || null,
                 tag: e.tagName.toLowerCase(),
                 type: (e.type || '').toLowerCase(),
                 label: lab, label_rule: how,
                 value: e.type === 'checkbox' || e.type === 'radio'
                        ? String(e.checked) : String(e.value),
                 attr_value: e.getAttribute('value'),
                 min: e.getAttribute('min'), max: e.getAttribute('max'),
                 step: e.getAttribute('step')};
    if (e.tagName === 'SELECT') {
      rec.options = Array.from(e.options).map(o => ({v: o.value, t: norm(o.textContent)}));
    }
    inputs.push(rec);
  });

  const buttons = [];
  document.querySelectorAll('button,input[type=button],input[type=submit]').forEach(e => {
    if (!inMain(e)) return;
    buttons.push({id: e.id || null, text: norm(e.textContent || e.value),
                  onclick: (e.getAttribute('onclick') || '').slice(0, 60),
                  disabled: !!e.disabled});
  });

  // Result boxes: an id-bearing element whose text looks like an output slot,
  // together with the caption above it. `.result-box label` is this site's
  // house style; the generic sweep catches anything that is not.
  // Per-output caption. `<p>Total Interest: <strong id=total-interest>£0.00</strong></p>`
  // is this site's commonest shape, so the parent's text MINUS the output's own
  // text is the caption; fall back to the nearest preceding heading/label. The
  // box-level `querySelector('label,h3')` the first version used captioned all
  // three of standard-calc's outputs "Monthly Repayment".
  const capFor = e => {
    const p = e.parentElement;
    if (p) {
      const own = norm(e.textContent);
      const whole = norm(p.textContent);
      if (whole.length > own.length && whole.endsWith(own)) {
        const c = whole.slice(0, whole.length - own.length).replace(/[:\s]+$/, '');
        if (c) return c;
      }
    }
    const heads = Array.from(document.querySelectorAll('label,h2,h3,h4,strong,span'));
    let best = null;
    for (const h of heads) {
      if (h.contains(e)) continue;
      if (h.compareDocumentPosition(e) & Node.DOCUMENT_POSITION_FOLLOWING) best = h;
    }
    return best ? norm(best.textContent) : '';
  };

  const outputs = [];
  const seen = new Set();
  document.querySelectorAll('.result-box, .results-box, .result, [class*=result]').forEach(box => {
    if (!inMain(box)) return;
    box.querySelectorAll('[id]').forEach(e => {
      if (seen.has(e.id)) return;
      seen.add(e.id);
      outputs.push({id: e.id, caption: capFor(e),
                    text: norm(e.textContent), where: 'result-box'});
    });
  });
  document.querySelectorAll('[id]').forEach(e => {
    if (!inMain(e) || seen.has(e.id) || e.children.length > 2) return;
    if (e.matches('input,select,textarea,button,form,script,style,link')) return;
    const t = norm(e.textContent);
    if (!/[£%]|^-?[\d,]+(\.\d+)?$/.test(t)) return;
    seen.add(e.id);
    outputs.push({id: e.id, caption: capFor(e), text: t, where: 'generic'});
  });

  // Every remaining id-bearing, non-control element that could HOLD an answer
  // but is empty right now. A verdict box that is empty until a press is
  // exactly what a per-page oracle needs to know exists — and exactly what a
  // "looks like money" filter cannot see before the press.
  document.querySelectorAll('[id]').forEach(e => {
    if (!inMain(e) || seen.has(e.id)) return;
    if (e.matches('input,select,textarea,button,form,script,style,link,nav,header,footer')) return;
    if (e.children.length > 3) return;
    seen.add(e.id);
    outputs.push({id: e.id, caption: capFor(e), text: norm(e.textContent),
                  where: 'empty-or-prose'});
  });

  return JSON.stringify({inputs, buttons, outputs,
                         scripts: document.querySelectorAll('script').length,
                         title: norm(document.title)});
})()
"""


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", required=True)
    ap.add_argument("pages", nargs="*", help="default: all 23 tool pages")
    a = ap.parse_args()

    pages = a.pages or TOOL_PAGES
    d = Driver()
    result = {}
    try:
        for p in pages:
            url = BASE + p
            try:
                d.goto(url)
                raw = d.ev(INVENTORY_JS)
                result[p] = json.loads(raw)
                inv = result[p]
                print("%-44s %2d inputs %2d buttons %2d outputs" %
                      (p, len(inv["inputs"]), len(inv["buttons"]), len(inv["outputs"])))
            except Exception as e:      # a page that will not load is a finding
                result[p] = {"error": str(e)}
                print("%-44s ERROR %s" % (p, e))
    finally:
        d.close()

    with open(a.out, "w") as f:
        json.dump(result, f, indent=1, sort_keys=True)
    print("\nwrote %s (%d pages)" % (a.out, len(result)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
