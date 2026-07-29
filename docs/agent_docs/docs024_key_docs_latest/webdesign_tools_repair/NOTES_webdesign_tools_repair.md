# NOTES — webdesign.co.uk tools repair

Append-only, newest at the bottom. Evidence, commands, and every misstep.

---

## 2026-07-29 — Phase A: the census, measured in a real browser

**The instruction was "none of the tools actually work". The measurement says
34 of 63 fail and 29 respond.** Both facts matter and neither is the whole
picture — see the caveat below, which is the important part.

### How it was measured

`scratchpad/toolprobe.py` (this session's; see RUNBOOK for the copy that ships).
Headless chromium over the DevTools protocol, one live URL at a time:
loads the page, records console errors and thrown exceptions, counts the
controls inside `<main>`, **drives the first meaningful control the way a
visitor would** (range/number → move it; text/textarea → type; select → change;
button → click), then compares a snapshot of the page's output regions before
and after. Four verdicts:

| verdict | meaning |
|---|---|
| **OK** | a control was driven and the page changed |
| **DEAD** | controls exist, driving one changes nothing |
| **BROKEN** | the console threw |
| **NO-CONTROL** | nothing to drive |

### The result (all 63, 2026-07-29)

| verdict | count | tools |
|---|---:|---|
| **OK** | 29 | ab-test-calculator, bg-remover, community-growth, csp-builder, css-filter-playground, entropy-meter, favicon-maker, fluid-typography, focus-ring, grid-generator, jwt-inspector, layout-generator, noise-generator, oklch-picker, parallax-generator, performance-budget, prompt-architect, prompt-permutator, recommender-engine, regex-tester, seo-schema, shadow-stacker, smooth-shadow, social-card, svg-optimizer, text-sanitizer, token-calculator, touch-target, white-balance |
| **DEAD** | 21 | aspect-ratio, bayesian-rank, blob-maker, clip-path, css-variables, diff-checker, golden-ratio, head-architect, html-minifier, image-optimizer, json-cleaner, magic-outliner, markdown-tables, meme-generator, mesh-gradient, monolith-splitter, privacy-redactor, smart-contrast, sri-generator, svg-patterns, text-extractor |
| **BROKEN** | 11 | animated-favicon, aria-builder, asset-formatter, blueprint-compiler, cubic-bezier, insight-injector, logic-architect, micro-cms, mind-map, pasteboard, vibe-equalizer |
| **NO-CONTROL** | 2 | rls-architect, seo-injector |

Recurring console errors among the BROKEN, verbatim:

```
TypeError: Cannot read properties of null (reading 'getContext')      animated-favicon
TypeError: Cannot read properties of null (reading 'appendChild')     asset-formatter
TypeError: Cannot read properties of null (reading 'addEventListener') blueprint-compiler
TypeError: Cannot set properties of null (setting 'innerHTML')        insight-injector, micro-cms
TypeError: Cannot set properties of null (setting 'value')            logic-architect, mind-map
```

Every one is the same shape: **the script addresses an element the page does not
contain.** The static pass agreed — 11 tools reference ids that appear nowhere
in their own HTML (`css-filter-playground` 6, `insight-injector` 5,
`seo-injector` 5, `logic-architect` 4). These pages were ported as HTML blobs
from two source sites; whatever trimmed them dropped markup the scripts need.

### THE CAVEAT, and it is the point

**"OK" means the page reacted, not that the tool is correct.** The probe drives
one control and asks whether anything changed. A tool that reacts with a wrong
answer passes. So:

- **34 failing is a floor, not a total.** Correctness has not been tested for any
  of the 29.
- The owner's "none of them work" is not contradicted by this census — it is a
  report about *usefulness*, which is a higher bar than *responds to input*, and
  the per-tool loop must test the tool's actual claim (does the contrast tool
  compute the right ratio?), not just its liveness.
- Where the census and the owner's report differ, this file records both. The
  census is what a machine could see; the report is what a person got.

### Misstep, recorded

**I nearly reported the static read as the census.** The first pass grepped the
HTML for inline scripts and missing ids and produced a tidy list of ~23 broken
tools. Two of its "no inline JS at all" entries (`community-growth`,
`csp-builder`) drive fine in a real browser — the static pass had only looked
inside `<main>` and missed scripts elsewhere in the document. **The browser is
the witness; the source read is a hypothesis.** Had I published the static list,
two working tools would have been queued for repair and the real count would
have been wrong in both directions.

### Second misstep: the probe harness itself

Chromium's DevTools HTTP endpoint stops answering at around the 14th tab in one
instance — the first full run died at tool 13 with a socket timeout and the
second at 36 with the run's own wall-clock timeout. Fixed by restarting the
browser every 6 URLs and retrying tab creation once. **A harness failure looks
exactly like a site failure in the output** (`probe failed: timed out` sat in
the same column as real verdicts); `aria-builder` and `cubic-bezier` are marked
BROKEN from such a timeout and **need a re-probe before anyone believes their
verdict** — flagged here rather than quietly corrected, because that is the
class of error this file exists to catch.

### CORRECTION to the table above (same day, after re-probing)

- **`aria-builder` is OK, not BROKEN.** Its BROKEN verdict was the harness
  timeout described above. Re-probed clean: `ctl=5 changed=True`. Corrected
  totals: **30 OK, 21 DEAD, 10 BROKEN, 2 NO-CONTROL — 33 of 63 failing.**
- **`cubic-bezier` and `vibe-equalizer` are UNVERIFIED, not BROKEN.** They time
  out on `Runtime.evaluate` across three harness variants (own-deadline reads,
  readiness polling, generous timeouts). **I nearly filed "two tool pages lock
  the browser"** — an independent mechanism refuted it first:
  `chromium --headless --dump-dom` renders both fully (10,972 and 15,340 bytes),
  as it does a known-good control. So the pages are fine and my probe is not;
  cause still unknown (not `alert()` — `vibe-equalizer` has none). **They get a
  hand check when their turn comes in the loop, and until then they carry no
  verdict at all.** Writing UNVERIFIED where I cannot measure is the whole point
  of the marker.
