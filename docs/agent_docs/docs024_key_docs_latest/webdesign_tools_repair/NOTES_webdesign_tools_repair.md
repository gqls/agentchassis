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

---

## 2026-07-29 — Repair #1: insight-injector (the loop, proven end to end)

First tool through the full §3 loop. Timings and evidence, so the next one can
be estimated:

1. **Measured** — census verdict BROKEN, console: `TypeError: Cannot set
   properties of null (setting 'innerHTML')` at load.
2. **Diagnosed** — the script addressed five elements (`#biz-name`, `#biz-fact`,
   `#biz-story`, `#biz-tone`, `#banned-tags`); the page contained one (`#output`).
   **The port dropped the whole left-hand panel.** Its CSS survived intact
   (`.tool-layout`, `.controls-panel`, `.input-group`, `.banned-words-box`,
   `.banned-tag`) and so did every line of its JS. The page even still told the
   visitor to "fill out the insights on the left" — and there was no left.
3. **Fixed** — rebuilt the panel to the contract the surviving CSS and JS
   already described. **Nothing redesigned and nothing invented**: every class
   name and every element id came from code already on the page.
4. **Verified in a browser, twice** — locally before shipping and on the live
   URL after: 15 banned-word tags render, and a generated prompt carries the
   business name, the hard fact, the customer story and the ban list. Screenshot
   checked too, because "the JS returns the right string" is not the same as
   "the page looks right".
5. **PLAN + NOTES written to the database** (`SQL_t01`) — aim, delivery
   mechanism, three *deliberate decisions — do not re-fix* (blank inputs are
   allowed on purpose; the ban list is deliberately not user-editable; the
   two-column layout is load-bearing because the copy refers to it), and a
   `criteria` fence with eight checks including two interaction tests.
   **This is the step that makes the repair durable** — the tool is now
   something the acceptance ladder can test rather than a page nobody watches.
6. **Shipped** — `gqls/sites`, deployed, re-probed live: **OK**.

**The transferable finding for the remaining BROKEN seven:** all eight throw
`null`-reference errors, and this one's cause was *missing markup, with working
CSS and JS still in place*. If that holds for the others, the repairs are
restorations rather than rewrites — cheap, and low-risk, because the surviving
code states exactly what the markup must provide. Check that assumption per
tool rather than assuming it: **the ids the script asks for are the
specification.**

---

## 2026-07-29 (fifth pass) — build the detector FIRST, then repair from what it says

Owner instruction: *"please fix the tools that you know are broken, using the
method of building them up step by step through the chassis. Please try and
repair the chassis so broken tools don't get deployed."*

I did the chassis half first, and that turned out to be the right order for a
reason I did not anticipate: **the detector's output is the specification for
each repair**, and the first thing it found was a defect in my own repair #1.

### The chassis half

`datahelpers.OrphanElementRefs` decides, from the artefact alone, which element
ids a page's script addresses that the page never contains and never creates.
Two consumers: a discovery check over deployed pages, and a hard pre-deploy
refusal in `DeployToolToSiteAction`. Plus `tool_eligibility.go`, which fixes the
structural reason none of this was ever watched.

**The eligibility finding is the one worth carrying off this workstream.** Every
tier of the acceptance ladder opened with `cc.component_level = 'tool'`. All 63
tools here are `'section'`. But widening THAT alone would not have worked:

```sql
SELECT cc.id, cc.function, count(*) FROM page_components pc
  JOIN content_components cc ON cc.id=pc.component_id ... GROUP BY 1,2;
-- a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef | ported-page | 97
```

**One component row is shared by 97 pages.** The ladder keys its subject on
`cc.function`, so all 63 would have collided onto one PLAN. A ported tool's
identity is its PAGE, not its component. That is why the widening derives the
key from `p.name` minus a leading `tool-`.

### THE MISSTEP THAT MATTERS: I confirmed a finding against a page I invented

The check first reported **10** findings. I "confirmed" one of them,
`css-filter-playground`, in a browser: six sliders, all missing, `rangeCount: 0`.
That went into a commit message, a register entry, a Go file header and a **live
council submission**.

It was false. I fetched `/tools/css-filter-playground.html`; the URL the database
gives is `/tools/css-filter-playground/index.html`. The first 404s, so I ran my
evidence against Chrome's error page. On the real URL: nine range inputs, all six
ids present. **The tool has never been broken.** Its sliders are built with
`id="${f.name}"` from a data array — present in the browser, absent from the
source. A real false-positive class, now handled.

What caught it was not review. I re-read that output twice while quoting it into
four documents. It fell out of a *different* question — two other pages appeared
to 404, which was implausible enough that I finally read `pages.url` out of the
database instead of assuming its shape.

Full account and the three cheap checks that would have caught it:
`docs024_key_docs_latest/WRONG_CALLS.md`, 2026-07-29.

**And I repeated a fault from this very file.** I nearly filed
"animated-favicon's code generator never runs" — there is a Generate Script
button ten lines above in the same file and I had not pressed it. That is the
fifth harness fault documented in the fourth pass above, repeated by the person
who wrote it down. Writing a lesson down does not install it.

### The repairs — nine tools, all driven, not merely loaded

| tool | what was missing | second defect found by driving it |
|---|---|---|
| animated-favicon | the hidden 32×32 resize canvas | — |
| asset-formatter | the whole input column + `#asset-list` | — |
| insight-injector | the input panel | **fixed only in the repo, not at source** |
| logic-architect | the toolbar | `saveHistory(true)` — no such function |
| mind-map | the toolbar | `saveHistory(true)` — same |
| monolith-splitter | the input column | — |
| pasteboard | the app bar | none — it *defines* `saveHistory` |
| rls-architect | the input column | — |
| seo-injector | the input column | — |

`saveHistory(true)` threw on **every** project load in two tools, so the
`render()` and `renderSidebar()` calls after it never ran. Nothing static would
have found it; it came from pressing the buttons. And **pasteboard genuinely
defines `saveHistory`** — assuming the family shared the bug would have broken a
working tool, which is why each was checked rather than swept.

Verification was behavioural, not structural: a favicon built from two uploaded
frames through the real file input; `seo-injector`'s output parsed back as JSON
and asserted field by field; undo actually stepping node counts back down.

### The other thing the detector found: my own repair #1 was not fixed at source

`tool-insight-injector` appeared in the very first run — a tool I had "repaired"
that morning. The repo file was fixed, deployed and re-probed OK; the DATABASE
still held the broken copy, last written 2026-07-27. For a ported page,
`page_components.rendered_html` **is** the artefact — `content_data` holds only
port provenance — so a repo-only fix is undone by the next publish. It had
already been undone: a `Rerender:` commit at 14:43 UTC reverted the file.

**So the loop now ends at the database, and `toolsource.py` enforces it** —
it refuses a push that still has orphan refs, that unbalances the section tags,
or that shrinks the stored HTML by more than a third without `SHRINK_OK=1`
(a truncated completion looks exactly like a small successful edit,
`bugs_open/012`).

A second, smaller trap the same hour: when the nine pages were spliced into the
site repo, insight-injector's produced **bytes identical to the local working
tree**, so it had no diff, so it was not in the commit — and the rebase onto
origin then took the remote's reverted copy. **On a shared tree, "no diff
locally" is not "already correct remotely."**

### Where the count stands

- **Before:** 9 pages fleet-wide whose scripts addressed absent elements.
- **After:** `0 of 98`, measured by running the REAL Go function over a fresh
  dump of the live database, not a Python mirror of it.
- All nine now carry a `doc_plans` PLAN with a ```criteria fence and a `doc_notes`
  repair entry, so the widened ladder can test them from here. That is the step
  that makes the repair durable; without it they go back to being unwatched.

### Still open

- **The DEAD seven** (aspect-ratio, blob-maker, clip-path, diff-checker,
  golden-ratio, magic-outliner, meme-generator) and the **UNVERIFIED two**
  (cubic-bezier, vibe-equalizer) are untouched. They are a different defect
  class — the orphan check does not flag them, and it was never going to.
- **[UNVERIFIED] whether a future `Rerender:` will revert these again.** The one
  that did it read a database that was still broken at the time, so it is not
  evidence either way. The next rerender of a tool page is the test. If it strips
  a panel again, the cause is the rerender regenerating from `content_data`
  rather than republishing `rendered_html`.
- **The discovery check is not enabled anywhere.** It is inert until
  `orphan_element_refs` is added to a discovery agent's `checks` array. That is
  deliberate (the image must be live first) but it means nothing is watching yet.

---

## 2026-07-29 (sixth pass) — the owner opened two tools I had scored OK, and both were broken

> *"the first tool I tested doesn't work properly, the fluid typography composer
> text doesn't change size when I resize the viewport… The second tool I checked
> doesn't work - /tools/micro-cms/index.html it has No Project Loaded and shows a
> blank box. Please check all tools again thoroughly and establish a workflow
> that checks everything we can check."*

Two out of two. That is a fair verdict on the bar I was using: **"OK" meant the
page reacted, not that the tool does what it claims.** I had written that caveat
into this file in the first pass and then let the census stand on it anyway.

### What each of the two actually was

**micro-cms** — the port kept the editor stage and dropped the ENTIRE sidebar.
The tool's logic lives in FOUR EXTERNAL FILES (`js/storage.js`, `editor.js`,
`ui.js`, `app.js`), and **that is why the static orphan-reference check never saw
it**: the check reads a page's own `rendered_html`, so all fourteen ids addressed
from those files were invisible. Thirteen tools on this site load relative
scripts the same way. Repairing it then exposed a SECOND bug underneath:
`createProject()` builds `history: []`, `loadProject()` reads `history[0].html`,
which threw before `setStatus()` could run — so a freshly created project still
read "No Project Loaded". The first defect had been hiding the second.

**fluid-typography** — not broken; *useless where people are*. It emits a
correct `clamp()`, and the preview carries it. But the preview is styled in `vw`,
and `vw` means the browser window. Measured across five widths before touching
anything:

```
 360px -> 16.7273px    900px -> 26.5455px   1600px -> 32px
 600px -> 21.0909px   1200px -> 32px
```

Above the max-width setting the clamp is pinned, so on any ordinary desktop
nothing moves — while the copy told the visitor to resize their window and
watch. Fixed by giving the preview its OWN viewport: a width slider driving an
iframe, since `vw` resolves against the iframe. Verified live, and it reproduces
the real-browser numbers exactly.

### The workflow — `toolaudit.py`, and every rule in it was wrong first

Nine checks. **Six of them I had to correct against a browser**, each after it
produced a false verdict on a working tool. That sequence IS the record:

| the rule as first written | what it wrongly condemned | the correction |
|---|---|---|
| fetch external scripts with `urllib` | every page — Cloudflare refuses a bare Python request | fetch **through the page**; it also sees exactly what the tool sees |
| any unresolved id is a defect | `regex-tester`, which replaces its own editor via `outerHTML` and works perfectly | only if the id is in **neither** the served source **nor** the DOM |
| flag any large empty region | 4 tools whose output boxes are *supposed* to start empty | measure **after** the interaction, not before |
| ...and any region with no text | `clip-path`, `community-growth`, `shadow-stacker` — a clip polygon, 12 chart bars, a live box-shadow | require `childElementCount === 0`; empty means *nothing there*, not *nothing written* |
| skip undo/copy/reset as no-ops | `cubic-bezier` — its first button is a preset named **Default**, a no-op on a fresh page | add `default` to the no-op list |
| an element with a background-image is not empty | `svg-patterns` | already fixed in pass three; kept |

**Twelve of 63 verdicts, before this, had been my harness rather than the site.
Six more would have been.** Every one made the site look worse than it is, and
every one was invisible in the output, because a false BROKEN prints exactly
like a true one.

### Repaired this pass, all driven live afterwards

- **micro-cms** — sidebar restored (14 ids read off the four JS files), plus the
  `history[0]` guard and a starter page. Now: `● LIVE EDITING: Test site`,
  `designMode: on`, 188 characters of editable content in the iframe.
- **blueprint-compiler** — opened at "2. Sitemap Architecture" with section 1,
  the compile button and the whole output panel missing. Now compiles 2,478
  characters naming the business and phone entered.
- **vibe-equalizer** — the card said "adjust the sliders" and there were none.
  Now the radius slider moves the card 20px → 36px with the readout tracking it.
  `#prompt-output` is a TEXTAREA, not a div: `state.js` copies `output.value`, so
  a div would have copied nothing while appearing to work.
- **fluid-typography** — simulated viewport, above.

### The rule this pass adds

**A tool is not verified until something has asserted its CLAIM, not its
liveness.** The claim is what the PLAN's ```criteria fence is for, which is why
every repaired tool gets one. `toolaudit.py` is the floor — it can only prove a
tool is *not obviously broken*. It cannot tell you the contrast ratio is right,
only that a number appeared.

### Closing state, 2026-07-29

Full audit over all 63 live tool pages: **63 RESPONDS, 0 flagged.** No BROKEN, no
DEAD, no EMPTY, no NO-CONTROL.

The number moved 34 → 24 → 23 → 20 → 9 → 5 → 0 across the day. **Every downward
step but one was my measurement being corrected, not a repair.** The repairs
themselves account for thirteen tools; the other thirty-odd "failures" never
existed.

Caveat, stated because the number invites the wrong reading: this run used the
eight-check harness, BEFORE the two-band responsive check was wired in. And
RESPONDS still only means *not obviously broken*. `fluid-typography` would have
scored RESPONDS on the morning of the day the owner reported it, because it did
respond — it just demonstrated nothing at the widths its visitors use. **The
audit is the floor. The criteria fence is the claim.** 13 tools have one; 50 do
not, and that is the whole of the remaining work.

### The tally of harness faults, because it is the most useful thing here

Nine, across two harnesses, every one making the site look worse than it was:

1. static source read instead of a browser (2 tools wrong)
2. `innerHTML.length` instead of content
3. invalid input typed into validating tools (7 wrong)
4. document-order control selection (3 wrong)
5. never pressing the action button (3 wrong)
6. `urllib` fetch of external scripts, refused by Cloudflare (would have been 63)
7. unresolved id reported without asking whether the PAGE removed it (1 wrong)
8. emptiness measured before the interaction, and without requiring no children (7 wrong)
9. one button pressed instead of each in turn — an already-selected option reads
   exactly like a dead control (2 wrong)

Plus two transient network errors that printed in the verdict column.

**The rule that would have caught all of them: before believing a negative
verdict, ask the page directly — and prove you asked the right page.**

---

## 2026-07-29 (seventh pass) — the concepts report, challenged before it was written

Owner: *"go over this report again challenging your findings in order to improve
it."* The findings came from three subagent sweeps; the challenge pass re-ran
every load-bearing claim against the live system. Five survived intact, four
were corrected, and one entirely new gap fell out — which is the strongest
argument yet for the challenge habit.

**Corrected before publication:**
- fence count 78 → **23** (78 was RFC_002's point-in-time measure over all rows)
- TL-014's "continuous sweep not in the binary" → STALE; `tool_acceptance_due`
  is in the running binary [pod-grep]
- G1 sharpened: the failing criteria DO reach the fixer's inputs
  (`spec.acceptance_test`, dispatch passes spec whole) — the prompt just never
  references them. Part of the fix is one prompt line.
- G4 narrowed: the judge guards no-criteria and no-results; only the
  all-individually-skipped case reads as PASS.
- G5 reshaped: improvement-sweep is off DELIBERATELY (owner ruling) — the
  draft would have recommended re-enabling something ruled stopped. The
  acceptance checks live on design-discovery-agent; `orphan_element_refs` sits
  on completeness-discovery-agent, which no enabled schedule targets.

**Found by the challenge, fixed, and then found wanting once more:**
**G6** — `request_browser_run` resolved pages by `name = function`, so every
acceptance run for a ported tool (13 now carry fences) would have hard-errored;
my morning eligibility widening had only reached the discovery side. Fixed
(two-candidate lookup), rolled v1.0.1205 — and then **the pilot's first live
run failed inside my fix**: mixing a bare `$2` with `'tool-' || $2` makes
Postgres deduce inconsistent types (42P08). `go build` cannot see that class.
Corrected with explicit casts, the statement proven by PREPARE/EXECUTE against
the live DB before shipping, rolled v1.0.1206. Two lessons, both cheap:
**the first live exercise of new SQL is part of the change, not part of the
future**, and a PREPARE against the live schema costs one second.

**The pilot** (smart-contrast, correlation fcb58019 failed on the 42P08; re-run
after the 300s spawn window on 1206): the first fence on this estate asserting
a tool's ARITHMETIC — known-answer pairs watched passing by hand first
(#767676/#ffffff → "4.54 : 1", #000000 → "21.00 : 1") per migration 148's rule.

---

## 2026-07-30 — four more tools broken, and the lookup failure behind the report

**The owner opened four tools this workstream had called repaired. All four were
unusable.** Not near misses:

- **micro-cms** — editable body 248px inside a 743px frame, so the caret could
  not be placed in the bottom two thirds; and a query for formatting buttons
  returned an EMPTY LIST. `designMode` was on and `execCommand('bold')` worked
  when called directly — the capability was there with no way to reach it. The
  false copy ("Click anywhere... Everything on this page is editable") was mine.
- **pasteboard, logic-architect, mind-map** — work areas measuring **1146x0**.
  One cause: each was ported from a standalone page whose `body` WAS the flex
  container, so `flex: 1` had no flex parent and the height collapsed, while the
  same rule restyled the host page (`overflow:hidden` on the site's own body).
  Measured fleet-wide: exactly 3 pages carry that leak; these three.

### How they passed me, which is the finding

I verified pasteboard by calling `addItem(src)` and logic-architect by calling
`loadTemplate('code')`. **A visitor cannot call a function.** Their real entry
points are a paste event and a click on a visible control, and the areas those
act on had no height. Two rules:

1. **Verify through the visitor's GESTURE, never the tool's internal functions.**
   If the vocabulary cannot express the gesture, that is a missing check type to
   record as a deferral — not a licence to substitute a function call.
2. **A fence asserts the TERMINAL value, not the first observable state change.**
   "Status reads LIVE EDITING" is a waypoint; "text can be edited and
   emphasised" is the point. My micro-cms fence asserted the waypoint.

Built so the rule is enforced rather than written: **TL-034 `has_visible_area`**,
a Tier-4 check measuring `getBoundingClientRect()` against a floor. Tier-4-only
by necessity — it measures rendered layout, so a Tier-2 equivalent is impossible.

### The lookup failure, recorded because it is the reusable part

The owner had asked me to find a prior discussion of staged tool building. I
searched `docs/` thoroughly, found two adjacent statements, and reported it not
found. **It was in `features_open/015` — the repo root, not `docs/`.** A search
agent I launched for it also died on a session limit, so nothing corrected me.

**`features_open/` is a first-class source. Read it.** Three entries turned out
to matter, at three scales: **015** the site maturity ladder (rungs, reference
examples, measurable progression criteria — a criteria fence one rung at a
time), **027** the component stage ladder S0–S7 raised the same day by another
session, and **026** the render-before-ship gate (fifty checks, none renders a
page; 101 WCAG-AA failures on one site while every status said deployed).

**027 reached this workstream's conclusion independently, on the same day:**
*"They all measured static markup or forced DOM state; not one ever fired a real
click. What was missing was not rigour — it was a stage."* Two lanes, one day,
one finding. Its S2 gate is the one thing neither lane had: **≥1 mutant red per
assertion class** — prove a check can fail before trusting it to pass. Nine
harness faults in this workstream are the argument for adopting it.

---

## 2026-07-31 — harness fault TEN, found by pointing `toolaudit.py` at another site

Contributed by the `loanandmortgagecalculator_couk` lane, which borrowed this
harness to verify 26 calculators it was porting. Recorded here because the fault
is in the harness, not in that lane's site.

**`toolaudit.py` scored 14 of 14 working calculators `NO-CONTROL`** — "nothing a
visitor can touch" — on `mortgagecalculator.co.uk`. Every one of those pages has
number inputs and a Calculate button, and every one computes correctly.

**Cause: every check scoped its query to `main …`, and that site has no `<main>`
element.** It wraps content in `<div class="container">`. Four sites in the file:
`DRIVE_JS` (controls + the output snapshot), the `empty_regions` query, and the
`_responsive` target search. With no `<main>`, `document.querySelectorAll('main
input, …')` returns `[]`, and an empty control set is reported as a site defect.

This is the **exact fault family the file's own docstring was written about** — a
harness failure printing in the same column as a site verdict — and it is the
worst direction to fail in, because it condemns working work. `--all` only ever
pointed at `webdesign.co.uk`, where every tool page does have `<main>`, so the
blindness was invisible for 63 pages.

**Fix: a `SCOPE_FN` prelude shared by all four sites.** Root is
`main` → `[role=main]` → `body`. It is deliberately **asymmetric**: when a
`<main>` exists, behaviour is unchanged, so no existing verdict can move. Only on
the `body` fallback is site chrome subtracted (`!e.closest('header,nav,footer')`)
— without that, a header of links and a nav of buttons would read as "controls"
and a dead tool inside a well-built shell would score PASS. Writing the fallback
symmetrically would have traded a false negative for a false positive.

**Both halves measured, not asserted** (S2's rule — prove it can fail, and prove
it did not break what passed):

| claim | check | result |
|---|---|---|
| the fix makes the blind pages visible | 14 pages re-audited | 13 calculators `RESPONDS`; the 14th is the hub page, whose "buttons" are all `<a>`-wrapped navigation, so `NO-CONTROL` is **correct** there |
| the fix moves nothing that had a `<main>` | original extracted from `git show HEAD:…`, both versions run over the same 4 `webdesign.co.uk` tools, **9 result fields** diffed each | **0 of 36 fields drifted** — verdict, controls, changed, pressed, unresolved, bad_subresources, empty_regions, responsive, console all identical |

The two pages chosen for the control are `fluid-typography` and `micro-cms` — the
two that motivated this harness's hardest checks — precisely because they are the
ones a scoping change would be most likely to disturb.

**Real findings on that site, which the blind harness had hidden behind
`NO-CONTROL`:** six of nine guides link `Home` to `index.html` from inside
`/guides/`, resolving to `/guides/index.html` → live 404; the homepage links to
`guides/mortgage-scorecard.html`, but the file is `your-mortgage-scorecard.html`
→ live 404; two guides are orphans; `sitemap.xml` is 404 while `robots.txt` still
carries the unfilled placeholder comment `# Sitemap location (replace with your
actual domain)`; no `favicon.ico`, so every page load logs a 404.

**The transferable rule: `--all` is a single-site sample, and a harness verified
against one site's markup conventions has only been verified against those
conventions.** The cheap check is to point it at any second site before trusting
a verdict from it — which is what happened here, by accident, and it cost one run
to find a fault that had been latent for 63 pages.
