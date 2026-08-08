#!/usr/bin/env python3
"""oracle_driver.py — drive a live calculator with EXPLICIT values, not scaled defaults.

RELATIONSHIP TO `toolgolden.py` (sibling lane). That harness deliberately knows
nothing about any individual calculator: it scales each numeric field's own
default by x1/x2/x0.5 so it stays in-domain for ANY tool with no per-tool
config. That generality is why it can fingerprint 23 pages unattended — and it
is also why it cannot test a boundary. A band edge is a specific number
(£125,000, £250,000, £500,000, £625,000, £925,000); no multiple of a page's own
default lands on one except by luck, and a defect that only exists between
£500,000 and £625,000 is invisible to every vector that harness can generate.
An oracle knows its tool, so it can name the number.

WHAT IS REUSED, DELIBERATELY. `Driver` subclasses that file's `Runner` for
chromium process management and tab creation, so there is one way to start a
browser in this estate rather than two that drift. `settle()` is RE-STATED here
rather than called, because there it is a closure inside `capture()` and is not
reachable — the logic is character-for-character the same, and it must be: a
bare `readyState == 'complete'` poll once certified a golden in which every
answer was £0.00, because the inline <script> holding `calculateLoan` had not
been parsed yet. A capture taken mid-parse SUCCEEDS and is silently worthless.

WHAT IS ADDED. Three guards that a consistency harness does not need but a
correctness oracle does:

  1. `goto()` REFUSES A DEPLOY-WINDOW BLOB. B2 answers a mid-deploy request with
     a ~7-line `NoSuchKey` JSON body at HTTP 200; every grep against it returns
     zero, which reads exactly like a clean pass (RUNBOOK §12). So the load
     asserts a leading `<!DOCTYPE` and a plausible byte count, and raises
     otherwise. A harness that cannot tell "the page says £0" from "there is no
     page" cannot report either honestly.

  2. `set()` VERIFIES THE VALUE LANDED. A `<input type=number min=... max=...>`
     silently clamps, and a `<select>` silently ignores an option value it does
     not have. Setting £625,000 into a field capped at £500,000 and then
     checking the answer against an oracle computed for £625,000 produces a
     confident FAIL against a correct calculator. So every write is read back
     and any divergence raises before a single expectation is compared.

  3. `read()` PARSES STRICTLY AND RAISES. `parse_money('')` returning 0.0 would
     turn "the output element was never written" into "the tool answered zero",
     and zero is a number an oracle can be wrong about quietly. Unparseable text
     is an error, not a value — this is the property the `--mutate parse`
     control exists to prove.

Not a standalone entry point; `inventory.py` and `oracle.py` drive it.
"""
import json
import os
import re
import sys
import time

# The sibling lane owns the browser harness; import it rather than fork it.
_SIB = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                    "..", "loancalculator_couk")
_HARNESS = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                        "..", "webdesign_tools_repair")
for _p in (_SIB, _HARNESS):
    if os.path.abspath(_p) not in sys.path:
        sys.path.insert(0, os.path.abspath(_p))

from toolgolden import Runner            # noqa: E402  (chromium + tab management)
from toolprobe import CDP                # noqa: E402

BASE = "https://loanandmortgagecalculator.co.uk/"

# The 23 tool pages. `guides/*`, `index`, `legal` and the three section indexes
# carry no calculator and are not in scope.
TOOL_PAGES = [
    "loans/standard-calc.html",
    "loans/compare-loans.html",
    "loans/consolidation.html",
    "loans/overpayment-calculator.html",
    "loans/interest-rate-stress-test.html",
    "loans/loan-vs-savings.html",
    "loans/settlement-calculator.html",
    "loans/car-finance-calculator.html",
    "loans/credit-health-check.html",
    "loans/damage-checker.html",
    "loans/application-tracker.html",
    "mortgages/repayment.html",
    "mortgages/simple.html",
    "mortgages/overpayment.html",
    "mortgages/stamp-duty.html",
    "mortgages/affordability.html",
    "mortgages/fee-analyser.html",
    "mortgages/investor.html",
    "mortgages/bridging-loan.html",
    "mortgages/equity-release.html",
    "mortgages/rate-forecaster.html",
    "mortgages/portfolio.html",
    "mortgages/fact-finder.html",
]

# A modal blocks the renderer and the NEXT evaluate simply times out with no
# indication why (toolgolden learned this on application-tracker's confirm()).
DIALOG_STUB = ("window.confirm=()=>true;window.alert=()=>undefined;"
               "window.prompt=()=>'';'stubbed'")
CLEAR_STATE = ("try{localStorage.clear();sessionStorage.clear()}catch(e){};'cleared'")

MIN_PAGE_BYTES = 3000


class PageLoadError(RuntimeError):
    pass


class DriveError(RuntimeError):
    pass


class ParseError(RuntimeError):
    pass


def parse_money(text, allow_blank=False):
    """'£1,234.56' -> 1234.56. Raises on anything it cannot read.

    The parse is stated because §4 of the brief requires it stated: strip a
    leading currency symbol, thousands separators, a trailing percent sign and
    surrounding whitespace; accept a leading minus; accept a bare integer.
    Anything else — including the empty string and 'Infinity' and 'NaN' — is a
    ParseError. Returning 0.0 for those is how a silent blank becomes a
    confident number.
    """
    if text is None:
        raise ParseError("no text (element missing?)")
    s = str(text).strip()
    if not s:
        if allow_blank:
            return None
        raise ParseError("empty text")
    s = s.replace("£", "").replace(" ", " ").strip()
    s = s.replace(",", "").replace("%", "").strip()
    # A result box sometimes reads "£1,234.56 per month" or "£0 / mo Rent"; take
    # the leading number only, but ONLY when the remainder is a unit phrase —
    # never when it contains a second number, which would mean the element holds
    # two answers and picking one is a guess. The trailing class therefore
    # excludes digits, which is what keeps '£1,234.56 and £99.00' a ParseError.
    #
    # The leading '/' was added after `--selftest-parse` refused '£0 / mo Rent',
    # a real reading from mortgages/portfolio.html's #d_rent: the first version
    # required the unit phrase to START with a letter, so that output would have
    # come back N/A — a check quietly not made, which reads like a check made.
    m = re.match(r"^(-?\d+(?:\.\d+)?)(?:\s*([A-Za-z/][A-Za-z /]*))?$", s)
    if not m:
        raise ParseError("unparseable as money/number: %r" % (text,))
    return float(m.group(1))


class Driver(Runner):
    """One chromium, many pages. Subclasses Runner for process + tab handling."""

    def __init__(self):
        super().__init__()
        self.cdp = CDP(self._tab()["webSocketDebuggerUrl"])
        self.cdp.call("Runtime.enable")
        self.cdp.call("Page.enable")
        self.url = None

    # ---- primitives -------------------------------------------------------

    def ev(self, expr, timeout=20):
        r = self.cdp.call("Runtime.evaluate",
                          {"expression": expr, "returnByValue": True,
                           "awaitPromise": True}, timeout=timeout)
        res = r.get("result", {})
        if res.get("exceptionDetails"):
            raise DriveError("JS threw: %s" %
                             json.dumps(res["exceptionDetails"])[:300])
        return res.get("result", {}).get("value")

    def _settle(self):
        """Wait for a FULLY PARSED document, not merely readyState.

        Re-stated from toolgolden.Runner.capture's closure of the same name; see
        the module docstring for why it is not simply called. Wait for
        `complete`, then require the script count AND the serialised DOM length
        to stop moving across consecutive reads.
        """
        for _ in range(80):
            if self.ev("document.readyState", timeout=10) == "complete":
                break
            time.sleep(0.1)
        last = None
        for _ in range(40):
            cur = self.ev("document.querySelectorAll('script').length + ':' + "
                          "document.documentElement.outerHTML.length", timeout=10)
            if cur is not None and cur == last:
                return cur
            last = cur
            time.sleep(0.15)
        return last

    def goto(self, url, clear_state=True):
        """Load a page, clear persisted state, reload, and REFUSE a NoSuchKey blob."""
        self.cdp.call("Page.navigate", {"url": url})
        self._settle()
        if clear_state:
            self.ev(CLEAR_STATE, timeout=10)
            self.cdp.call("Page.navigate", {"url": url})
            self._settle()
        self.ev(DIALOG_STUB, timeout=10)

        html_len = self.ev("document.documentElement.outerHTML.length")
        head = self.ev("document.documentElement.outerHTML.slice(0,40)")
        if not isinstance(html_len, int) or html_len < MIN_PAGE_BYTES:
            raise PageLoadError("%s: document is %s bytes (< %d) — mid-deploy "
                                "NoSuchKey blob? head=%r"
                                % (url, html_len, MIN_PAGE_BYTES, head))
        if "NoSuchKey" in (self.ev("document.body.textContent.slice(0,200)") or ""):
            raise PageLoadError("%s: body contains NoSuchKey — deploy in flight" % url)
        self.url = url
        return html_len

    # ---- driving ----------------------------------------------------------

    def set(self, sel, value):
        """Set a control and VERIFY the value landed (clamping/unknown-option)."""
        js = r"""
        (() => {
          const e = document.querySelector(%s);
          if (!e) return JSON.stringify({ok:false, why:'no such element'});
          const v = %s;
          if (e.type === 'checkbox' || e.type === 'radio') {
            const want = (v === true || v === 'true' || v === 1 || v === '1');
            if (e.checked !== want) e.click();
            return JSON.stringify({ok: e.checked === want, got: String(e.checked),
                                   want: String(want)});
          }
          e.value = String(v);
          ['input','change'].forEach(ev =>
            e.dispatchEvent(new Event(ev, {bubbles: true})));
          return JSON.stringify({ok: String(e.value) === String(v),
                                 got: String(e.value), want: String(v)});
        })()
        """ % (json.dumps(sel), json.dumps(value))
        r = json.loads(self.ev(js))
        if not r.get("ok"):
            raise DriveError("set(%s, %r) did not land: %s"
                             % (sel, value, r.get("why") or
                                "element holds %r, wanted %r" % (r.get("got"), r.get("want"))))
        return r

    def click(self, sel):
        js = r"""
        (() => {
          const e = document.querySelector(%s);
          if (!e) return 'MISSING';
          if (e.disabled) return 'DISABLED';
          e.click();
          return 'ok';
        })()
        """ % json.dumps(sel)
        r = self.ev(js, timeout=30)
        if r != "ok":
            raise DriveError("click(%s): %s" % (sel, r))

    def text(self, sel):
        js = r"""
        (() => {
          const e = document.querySelector(%s);
          if (!e) return null;
          return (e.textContent || '').replace(/\s+/g, ' ').trim();
        })()
        """ % json.dumps(sel)
        return self.ev(js)

    def read(self, sel, raw=False):
        """Read an output element as a number. Raises if missing or unparseable."""
        t = self.text(sel)
        if t is None:
            raise ParseError("read(%s): no such element" % sel)
        if raw:
            return t
        return parse_money(t)

    def seed_storage(self, key, value):
        self.ev("localStorage.setItem(%s, %s); 'seeded'"
                % (json.dumps(key), json.dumps(json.dumps(value))))

    def close(self):
        try:
            super().close()
        except Exception:
            pass
