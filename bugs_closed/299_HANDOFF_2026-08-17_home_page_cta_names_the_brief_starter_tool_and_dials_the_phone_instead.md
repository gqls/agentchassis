# 299 — the home page's second CTA names the Website Brief Starter and its href DIALS THE PHONE; the tel: URI is malformed too, and the section was written AFTER the 268 fleet fix

**Filed 2026-08-17** by the `webdesign_uk_build_service` lane, **owner-reported**
("this text links nowhere"). It is worse than a dead link: it goes somewhere, and
somewhere wrong.

## The defect, verbatim from the served page

`https://preview.webdesign.uk/index.html`, the call-to-action section:

```html
<a href="tel:+44 (0) 7934 524 911" class="cta-btn cta-btn-secondary">Or answer a couple of
quick questions first with the Website Brief Starter, a tool that helps you set out what you
need before we talk.</a>
```

Three separate faults in one element:

1. **The copy names a TOOL; the href dials a PHONE.** A visitor who clicks the
   sentence about answering questions gets their dialler. This is the
   `cta_names_unknown_destination` / misdirected-CTA class (the `bugs_closed/268`
   family) — copy promising X, link delivering Y.
2. **The destination exists and is correctly linked elsewhere.**
   `/tools/website-brief-starter/index.html` is live and is linked properly from
   the nav and the footer on both `index` and `contact` (measured 2026-08-17).
   So this is not a missing page; it is a wrong href on one element.
3. **The `tel:` URI is itself malformed** — `tel:+44 (0) 7934 524 911` contains
   spaces and parentheses. Even read as a phone link it is not a valid `tel:` URI
   (`tel:+447934524911`).

Whole sentence as a button label is a fourth, milder problem: a 130-character CTA
is not a button, and no button copy in the house voice looks like this.

## Why this is a PLATFORM finding, not just bad copy

**The section was written AFTER the 268 fleet fix was live.** `page_components` for
`index`: `call-to-action` `updated_at = 2026-08-16 16:12:45`; the 268 CTA-destination
fleet fix shipped in `v1.0.1298+` and the sibling lane recorded it live before that.
So a chassis carrying the fix produced this element. Either the guard does not
recognise this shape (copy naming a TOOL, href a `tel:`), or it does not run on this
path. **That is the question the fixing thread must answer first** — the copy can be
corrected in one UPDATE and will be regenerated wrong again by the next rebuild if
the producer is not fixed.

Note the fleet already carries `cta_names_unknown_destination` rows for this site on
other pages (`what-you-get`, `how-it-works`) sitting in `needs_human_review`, and a
`misdirected_cta:index-rejected-v1-20260806` item that is `failed`. **No open item
covers this element**, which is why it reached the owner's eye rather than a queue.

## How to verify (and the control that stops a false pass)

```bash
curl -s https://preview.webdesign.uk/index.html | grep -o '<a[^>]*>[^<]*Website Brief Starter[^<]*</a>'
```
Expect, after a fix, an href of `/tools/website-brief-starter/index.html`.
**Control:** the SAME page's nav and footer already link that tool correctly, so a
check that merely greps the page for the correct URL PASSES TODAY, while the broken
button is untouched. Assert on the anchor whose TEXT contains "Brief Starter" and a
verb ("answer", "questions"), not on the presence of the URL anywhere in the page.

## Scope note — do not fix this in isolation right now

The owner is finalising a new plan for this site (2026-08-17, other session
`webdesign live web builder project`) under which **the whole site will be rewritten**,
the chat box moves to the home page, and the positioning copy is rewritten. This CTA
sits in exactly the section that work will touch. **File, do not patch**: a surgical
copy fix now is discarded by the rewrite, and the producer question above is the part
that survives it. Whoever does the rewrite must confirm the regenerated CTA points at
the tool page, using the control above.

---

## 2026-08-18 — TAKEN ON by the `bugfix_299_cta_dials_phone` lane; the producer question is ANSWERED

**The section was regenerated again (08-18 10:31, the chat-box rebuild) and the defect
survived in a new form:** the label is now "See how it works" (naming `/how-it-works.html`),
the href is still `tel:+44 (0) 7934 524 911`. Label and destination are written by two
different sources and nothing ties them together — a rewrite changes one and carries the
other.

**Why a chassis carrying the 268 fix still produces this — traced, not inferred
(→ `bugs_open/312`):** on that same rebuild the internal-link resolver computed the RIGHT
destination (both CTA fields → `/tools/website-brief-starter/index.html`, target titles
included) and `page-content-writer`'s `select_sections` step discarded it — its first
extract path names a `link_resolution` level the response does not have (0 of 150 retained
runs match), and the silent fallback fed the render the pre-resolver plan, whose
resolved_data is the 268 carry of the stored row. The carry then faithfully re-ships the
tel: through every rebuild. So: right answer computed, thrown away, damage carried — the
268 fix is part of the persistence mechanism, not a failed guard.

Two further halves, confirmed while validating the fix design: the misdirected-CTA check
CANNOT see this shape (`ClassifyLinkScope` files `tel:` under mailto and the scan skips it
before classification — it ran on this site 08-14 and 08-17 with this anchor live), and the
repair path would DELETE a genuine phone button rather than fix this one
(`applyCTARecompute`'s keep branch requires `validPages.Contains`, false for any non-page
href — the `LANDMINES.md` bug-203 trap in a second form; faq and how-it-works carry real
phone CTAs today).

**Fix in flight** (plan approved by the owner 08-18, working docs
`docs024_key_docs_latest/bugfix_299_cta_dials_phone/`): shared non-page-destination
vocabulary + tel: normaliser in `datahelpers`; keep branches in `setCTAField` and
`applyCTARecompute` (coordinated with `bugs_open/248`, which owns the page-scheme half); a
`cta_nonpage_destination` discovery check; the archived-page scan filter; a gated
destination stamp into the writer's field specs; and — LAST, under an interlock, because
fixing it first would arm the 248 clobber on every fresh build — the `select_sections`
path correction (312).

This file stays OPEN until the regenerated CTA's copy and href agree on the served page
(fixed AND live bar). Verification stays as §"How to verify" above — assert on the anchor
TEXT, never on the URL's presence in the page.

---

## 2026-08-19 — still OPEN, but the Go half is LIVE and two of the three switches are thrown

**The bug re-validated itself while being checked.** The served page still carries
`href="tel:+44 (0) 7934 524 911"` on `cta-btn-secondary`. The label has now changed a
**third** time — stored `updated_at` 2026-08-19 10:17:38, and the history reads:

| when | label | href |
|---|---|---|
| 08-16 16:12 | "Or answer a few short questions first with the Website Brief Starter…" | `tel:+44 (0) 7934 524 911` |
| 08-18 10:31 | "Or answer a couple of quick questions first with the Website Brief Starter…" | `tel:+44 (0) 7934 524 911` |
| 08-18 12:10 | "See how it works" | `tel:+44 (0) 7934 524 911` |
| 08-19 10:17 | "Read the full terms in our FAQ before you pay." | `tel:+44 (0) 7934 524 911` |

Four rewrites, four labels, one unmoved href. That is `bugs_open/312`'s carry, and it is
now the clearest single piece of evidence in this file: **whatever rewrites the copy is not
what writes the destination, and only one of them is changing.**

### What is live

- **The Go fix is in production**: chassis **v1.0.1316**, verified on BOTH pods by capability
  probe with a negative control absent each time (`NormalizeTelHref`,
  `IsAuthoredNonPageCTADestination`, `DescribeCTADestination`, `ctaTargetTitleField`, and
  248's `storedCTADestinationIsAuthored`). *Not* verified by commit ancestry — that check was
  unavailable, and why is now `LANDMINES.md` + `RFC_040`.
- **Migration 475 APPLIED** — `cta_nonpage_destination` armed on `completeness-discovery-agent`
  (checks 43 → 44, and on that agent only). The class is no longer invisible.
- **Migration 476 APPLIED** — the destination stamp is on, and **deliberately inert until 477**.
  A zero in `llm_call_log` for `Destination (fixed):` is the designed state right now; do not
  read it as a broken stamp. The migration header says so in place.
- **Migration 477 NOT applied.** It is the fleet-wide one. Both keep halves are now live, which
  satisfies its stated ancestry precondition — but its canary is still owed and the owner's
  decision #1 is unanswered. See below.

### What still has to happen for this file to close

The bar is unchanged and it is the served page: **the regenerated CTA's copy and href must
agree**, asserted on the anchor whose TEXT names a destination — never on the URL's presence
anywhere in the page, because nav and footer link the tool correctly and would pass that grep
today. In order:

1. Owner answers decision #1 (gate 477 on both keeps — now satisfied — or hold longer) and
   decision #2 (should this button end as a phone button with honest copy, or a Brief Starter
   link). Decision #3 (the intended number for the undialable `tel:+4407934524911`) is still
   open and is a one-line fix once answered.
2. Apply 477, canary `leopardessconsulting.co.uk` (authored `/contact.html` CTAs ×4 across
   /index and /how-it-works) — diff the CTA urls before and after. **Survival is the control
   that the keeps, not luck, made it safe.**
3. Rebuild webdesign.uk `index` through the normal pipeline (never by hand — the 2026-08-04
   owner ruling) and assert on the anchor text.

### Note for whoever picks this up

`bugs_open/312` is now confirmed at fleet scale rather than by a single trace: of 48 retained
`page-content-writer` runs carrying both structures, the resolver minted `*_target_title` on a
CTA section in 26 and **0 survived** into `sections_for_render`; 30 of 48 runs differ between
the two sides. The `*_target_title` keys are the sharp instrument here — only the resolver
mints them, so their absence downstream IS the discard, and it stays visible even on the runs
where the URLs coincidentally agree (18 of 48 are byte-identical, and a url-diff would score
those as healthy).

---

## 2026-08-20 — CLOSED: fixed AND live at the served page

```html
<a href="/faq.html" class="cta-btn cta-btn-secondary">Read the full terms in our FAQ before you pay.</a>
```

Fetched from `https://preview.webdesign.uk/index.html` (cache-busted), 2026-08-20. **The copy and
the destination agree.** Asserted on the anchor's TEXT, per this file's own control — never on
the URL's presence in the page, because nav and footer link the Brief Starter correctly and
would pass that grep regardless.

### What actually fixed it, in the order it happened

1. **The producer** (`bugs_open/312`): migration **477** repointed `select_sections` at the path
   the resolver's response really has. Before: 41 runs, 33 minted `*_target_title`, **0 survived**
   to the render. After: **4 of 4 survive, byte-identical**. The resolver's answers stopped being
   computed and thrown away.
2. **The keeps** (chassis v1.0.1317, capability-probed on both pods): the tel: was kept and
   **normalised** — `tel:+44 (0) 7934 524 911` → `tel:+447934524911` — with a computed destination
   title. Three of the five malformed tel: CTAs on this site self-repaired this way as their
   pages rerendered, with no human involved.
3. **The detector** (migration 475): `cta_nonpage_destination` went live and its **first run
   filed this very button** as `cta_names_nonpage_destination`. The element that had to reach the
   owner's eye because no queue could see it is now something a queue sees.
4. **The two a machine must not decide** (`SQL_2026-08-20_close_299_…sql`, owner-answered):
   the undialable `tel:+4407934524911` on contact/hero → `tel:+447934524911` (**owner confirmed
   the number**; `NormalizeTelHref` refuses this shape by design rather than inventing digits),
   and this button → `/faq.html`, matching its own copy. **The owner confirmed the phone button
   was never intentional** — and it would otherwise have been preserved for ever, because the fix
   teaches the framework that a tel: is *authored*, and the framework cannot tell
   authored-on-purpose from inherited-by-accident. It appears to have been a leftover from the
   2026-08-13 copy "Prefer to talk it through first? Call +44 (0) 7934 524 911 or email…".
5. **The rerender that made it live** — and it took three attempts, all of which reported
   success. See `LANDMINES.md`, *"A `page-rerender` dispatched without `spec.page_name` throws
   away everything it re-rendered…"*: without `spec.reason` it assembles stored HTML; with the
   reason but no `spec.page_name` it re-renders and then **discards** the result
   (`sections_saved: 0`, `success: true`) and deploys the stale assembly. Working envelope and
   the checks that catch it are in that entry and in the lane RUNBOOK.

### Not closed by this, and where it went instead

- **Two more instances of the same class on this site**, both now filed by the new detector and
  sitting in `needs_human_review`: `faq/hero` ("See what you get for it" → a phone) and
  `how-it-works/call-to-action` ("Still deciding? The FAQ page covers the full terms…" → a
  phone). Those tel: hrefs are believed **genuine** call-us buttons, so the fix there is the
  COPY, not the href — which is exactly what migration 476's destination stamp now feeds the
  writer. They are the first real test of it.
- `how-it-works/call-to-action` still carries the unnormalised `tel:+44 (0) 7934 524 911`; it
  will self-heal on its next rerender through the keep. No action needed.
- **`bugs_open/312` stays OPEN** on its own merits: its candidates 2 (a loud fallback) and 3 (a
  lockstep test binding the writer's configured path to the resolver's actual response shape)
  are unbuilt, and that seam has now failed silently in BOTH directions twice.

### Platform residue this bug leaves behind (all registered)

`LNK-034` (the non-page CTA vocabulary + detector), `BLD-023` / `RFC_040` (a binary publishes
what it can do — raised by the discovery that the estate's prescribed "is my fix live?" check
frequently cannot be performed at all), and three `LANDMINES.md` entries: the capability-probe
one, the rerender-dispatch one above, and LNK-034's ordering-dependency note.
