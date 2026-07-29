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

---

## 2026-07-29 (later) — THE CENSUS WAS WRONG, and the reason is the most useful thing in this file

**A third of the "DEAD" verdicts were my probe's fault, not the tools'.**
Re-probing all 21 with corrected input: **7 flipped to OK, 14 stayed DEAD.**

### Fault 1 — I drove the tools with input they were right to reject

The probe typed `probe123` onto the end of whatever a text field held. For
`smart-contrast` that made `#999999probe123` — not a colour. The tool's
`hexToRgb` returned null, the tool correctly did nothing, and the probe wrote
DEAD. **The tool was working the whole time.** Proof, evaluated in the page:

```
{engineLoaded: "function", ratioText: "2.85 : 1"}   # #999 on #fff, correct WCAG
```

So the probe was scoring *input validation* as *breakage* — it punished the
tools that check their inputs, which are the better-written ones. Fixed by
inferring a plausible value from the field itself: a hex field gets a different
valid hex, a numeric field gets a different number, a JSON field gets JSON,
an HTML field gets markup.

### Fault 2 — comparing `innerHTML.length` instead of `innerHTML`

A tool that rewrites `3.5` as `2.8` leaves the length identical. Same verdict,
same cause: **the measurement could not see the change it was looking for.**
Now compares the full markup, plus live form values, plus canvas pixel data
(`toDataURL().length`), with the driven element excluded so the probe cannot
mistake its own keystroke for the tool's answer.

### Corrected census (2026-07-29)

| verdict | count |
|---|---:|
| **OK** (responds to valid input) | **37** |
| **DEAD** (responds to nothing) | **14** |
| **BROKEN** (console throws) | **10** |
| **UNVERIFIED** (probe cannot measure) | **2** |

Still failing, 24 of 63 — the repair queue:

- **BROKEN (10):** animated-favicon, asset-formatter, blueprint-compiler,
  insight-injector, logic-architect, micro-cms, mind-map, pasteboard,
  + cubic-bezier and vibe-equalizer once verified by hand.
- **DEAD (14):** aspect-ratio, blob-maker, clip-path, diff-checker,
  golden-ratio, head-architect, image-optimizer, json-cleaner, magic-outliner,
  meme-generator, monolith-splitter, privacy-redactor, sri-generator,
  svg-patterns.

### The lesson, stated plainly

**Three times today the measurement was wrong in a way that made the site look
worse than it is** (static-read vs browser; length-compare; invalid input), and
each time the error was invisible in the output — DEAD looked like DEAD.
A verdict is only as good as the drive that produced it, so the probe now
prints WHAT it typed (`value="..."` in the note) beside every verdict. **If a
tool is reported broken, the first question is what the prober did to it.**

None of this softens the owner's report. 24 tools genuinely fail, and "OK" here
still only means *responds* — correctness is untested, and the per-tool loop
tests the tool's actual claim.

---

## 2026-07-29 (third pass) — a FOURTH harness fault, and the census that finally holds

**`document.querySelector('a, b, c')` returns the first match in DOCUMENT
order, not in the order the selectors are written.** My "drive the first
meaningful control" list was written preference-first (range → number → text →
… → button) and I read it as a priority list. It is not. On any page whose
buttons sit above its inputs, the probe clicked a button — and on this site
those are usually radio-style type pickers (`setType('dot')`, already active)
or `Undo` with nothing to undo. Clicking them correctly does nothing, and the
probe wrote DEAD.

Caught on `svg-patterns`: the census called it DEAD, but evaluating in the page
showed `setType`, `copyCSS` and `update` all defined, `cssOutput` full of valid
CSS and `preview` carrying a live background image. **The tool had never been
broken.** Fixed by trying each selector in turn and skipping controls whose id,
class or label reads as undo/redo/reset/clear/copy/download/share/print — a
control that is *supposed* to be a no-op on a fresh page cannot test liveness.

Re-probing the 14 flipped **3 more to OK** (image-optimizer, privacy-redactor,
svg-patterns).

### Final census, 2026-07-29 — this is the one to work from

| verdict | count | tools |
|---|---:|---|
| **OK** (responds to valid input) | **40** | the 37 earlier, plus image-optimizer, privacy-redactor, svg-patterns |
| **DEAD** | **10** | aspect-ratio, blob-maker, clip-path, diff-checker, golden-ratio, head-architect, json-cleaner, magic-outliner, meme-generator, sri-generator |
| **BROKEN** (console throws) | **8** | animated-favicon, asset-formatter, blueprint-compiler, insight-injector, logic-architect, micro-cms, mind-map, pasteboard |
| **NO-CONTROL** | **3** | monolith-splitter, rls-architect, seo-injector — each has 0–1 controls where its own copy promises an interactive tool, so these are broken in a different way: the markup the script needs was never ported |
| **UNVERIFIED** | **2** | cubic-bezier, vibe-equalizer |

**23 of 63 fail. The repair queue is those 23.**

### The pattern across four faults — worth carrying off this workstream

Every one of my measurement errors made the site look **worse** than it is, and
every one was invisible in the output, because a false DEAD and a true DEAD
print the same word. The sequence: static read instead of browser (2 wrong) →
`innerHTML.length` instead of content → invalid input into validating tools
(7 wrong) → document-order control selection (3 wrong). **12 of 63 verdicts,
19%, were my harness rather than the site.**

The fix that would have caught all four earlier is the same one: **before
believing a negative verdict, open the page and ask it what it thinks.** One
`evalpage.py` call against `svg-patterns` refuted a verdict that four rounds of
static reasoning had reinforced. The probe now prints the value it typed beside
each verdict so the drive is auditable from the output alone.

**This does not overturn the owner's report.** 23 tools genuinely fail, and OK
still only means *responds* — the per-tool loop tests each tool's actual claim.

---

## 2026-07-29 (fourth pass) — a FIFTH fault, and the census stabilises at 20

**Most of these tools are paste-then-press, and the probe never pressed.**
`sri-generator` was scored DEAD; driven properly it produces
`integrity="sha384-vuz+yO71bcb30P4dMUNzy6/D2y+6d/n0KcOnt5clJtTBxEDoKAqGay0stFlC8Dpr"`,
which **matches an independent Python `hashlib.sha384` of the same input
byte-for-byte.** The tool was not merely alive — it was *correct*, while being
counted among the broken.

Added a second phase: if typing changed nothing, find the action button
(generate/run/convert/build/calculate/format/minify/compile/analyse/check/…,
excluding undo/copy/download), click it, wait 1.2s for async work
(`crypto.subtle`, FileReader, image decode) and re-measure. That flipped
`sri-generator`, `json-cleaner` and `head-architect` to OK.

### Census, stabilised (2026-07-29)

| verdict | count |
|---|---:|
| **OK** | **43** |
| **DEAD** | **7** — aspect-ratio, blob-maker, clip-path, diff-checker, golden-ratio, magic-outliner, meme-generator |
| **BROKEN** | **8** — animated-favicon, asset-formatter, blueprint-compiler, insight-injector, logic-architect, micro-cms, mind-map, pasteboard |
| **NO-CONTROL** | **3** — monolith-splitter, rls-architect, seo-injector |
| **UNVERIFIED** | **2** — cubic-bezier, vibe-equalizer |

**20 of 63 fail.** The number moved 34 → 24 → 23 → 20 as the harness improved,
**every time downward, every time because my measurement was wrong rather than
the site.** Five distinct faults: static-read-not-browser, length-not-content,
invalid-input-into-validating-tools, document-order-control-selection, and
never-pressing-the-button.

### What that means for this workstream, stated plainly

The correct response is not "the tools are fine". Twenty genuinely fail, and the
BROKEN eight throw `null`-reference errors that mean **markup the scripts need
was never ported** — a real, common, fixable cause. But the repair queue is a
third of what the first census said, and had I started fixing from that list I
would have "repaired" working tools, most likely breaking some.

**The rule this workstream now runs on: a negative verdict is a hypothesis
until the page has been asked directly.** `evalpage.py <url> <expr>` is one
command and it refuted four of my five faults on first use.
