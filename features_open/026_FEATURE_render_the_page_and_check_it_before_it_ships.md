# 026 — render the page and check it, because every check we have reads a source

**Raised:** 2026-07-27, from the owner's question after seeing fundamentallyai.com on
his phone: *"why weren't these bugs picked up, what do we need to do to ensure we
produce working sites… the build process should be correct and check."*
**Owner:** unowned. Groundwork committed (`scripts/render_audit.py`).
**Anchors:** `bugs_open/113` (palette composition), `bugs_open/114` (imagery).

---

## The finding, stated plainly

The platform has about fifty discovery checks. They are good checks. **Not one of them
renders a page.** Every check reads an input — a component template, a palette row, a
token vocabulary, an href, an `assets` row — and asks whether that input looks right.

Three defect families cannot be seen that way, because they are properties of the
composition and not of any input:

1. **a slot the layout fills with a literal because the palette omits it.** Palette
   valid, layout valid, merge unreadable. (113)
2. **one token used in two roles.** `--color-primary` is a foreground in 53 places and
   a fill in 26; a value can be correct for one and invisible in the other. (113)
3. **a component hard-coding an ink over a themed background.** The template reads
   fine; the pairing is what fails.

Measured cost of that gap on one site: **101 WCAG-AA failures across five pages**,
including every card title, the entire chart section, and the owner's own example
("every decision leaves a record", 1.21:1). Plus five images rendering as broken icons.
Every page's status said `deployed`; every check said nothing.

## What already exists, and what it proves

- `check_forced_text_colors` detects dark-on-dark — **within a `<style>` block**. It
  cannot see a colour that arrives through a CSS variable defined in the stylesheet,
  which is every case in 113. Its handler also never existed until `bugs_closed/077`.
- `check_image_url_404` states in its own header that the HTTP half is *"deferred until
  we have git-adapter integration on the discovery path"*. That deferred half is the
  half that catches a real 404.
- `check_hardcoded_section_colors`, `check_missing_css`, `check_duplicate_palette`,
  `check_placeholder_image_in_use` — all source-side, all sound, all blind to this.

The machinery is not missing. **The vantage point is.**

## Second finding: detection already outruns consumption

Making a new detector is only worth it if something drains it. Two measurements from
the same day say that is not currently true:

```sql
SELECT item_type, status, count(*) FROM site_work_items
 WHERE item_type LIKE 'audit_finding%' GROUP BY 1,2;
   audit_finding_brief_fidelity | detected | 3
```

Three rows, fleet-wide, all from fundamentallyai.com on 2026-07-24, all still
`detected`. One of them reads: *"Only 2 of 27 components contain images — raising
serious doubt that the illustration system is meaningfully present."* That is the
owner's complaint, filed by our own platform, three days before he made it.

```sql
SELECT status, count(*) FROM site_work_items WHERE spec->>'reason'='image_landed' GROUP BY 1;
   needs_human_review | 14
   complete           | 13
```

So: **the answer to "why wasn't this picked up" is partly "it was".**

## Proposed shape

### Phase 1 — the audit exists and runs (done)

`scripts/render_audit.py`. Renders in headless Chromium, composites the effective
background for every visible text node, reports pairs below AA, reports images that
failed to load, re-checks each failure over HTTP before reporting it, exits non-zero.
Usable today against any live site with no cluster access.

### Phase 2 — the cheap half, in the pipeline, no browser

> **STATUS 2026-07-27 — the shared package is BUILT and committed (`e43a3bda0`);
> the check itself is NOT.** Council `31bad59f-6366-4116-a41c-b6ece45bd634`.
>
> **Done.** `platform/colour` holds the WCAG maths once — `ParseHex`,
> `RelativeLuminance`, `ContrastRatio`, `IsDark`, `IsPerceptuallyDark`, plus
> `Pair` and `AuditPalette`. `actions/color_util.go` is now five thin wrappers,
> so no call site in `actions` changed and its palette tests pass untouched.
> This is the half this feature identified as *"the work"*, and the reason
> stands: `actions` imports `discovery_checks`, so a second copy of the formulas
> was the only alternative, and a checker that disagreed with the renderer's own
> dark/light classification would be worse than none.
>
> `AuditPalette` evaluates nine pairings, each carrying its provenance in a
> comment. The tests are a **regression corpus**, not a demonstration:
> fundamentallyai.com's real palette before and after today's repair.
> `TestAuditPaletteCatchesTheRepairRegression` pins the property that matters —
> **the repair fixed five pairings and broke a sixth**, so a checker run only
> *after* a palette change would report success. Control assertions carry equal
> weight: `text_muted/background` measured 5.97:1, passed throughout, and the
> test fails if the audit flags it.
>
> **Not done, and deliberately.** The discovery check that loads a site's
> **composed** palette and files the finding. Composing it means mirroring
> `buildPaletteMap` (`render_css_composition_helpers.go:72`) plus the
> core-vs-specialised authority rule, and a second loader that diverged would
> recreate the defect being removed. `AuditPalette` is therefore **dead code
> until that lands** — stated plainly because dead code that looks live is its
> own trap.
>
> **Two things the next author needs.** (1) The slot NAMES in `auditPairs`
> (`card_bg`, `text_muted`, `primary_text`) are assumed to match the renderer's
> and are **not yet verified against a live composition** — that is the first
> thing to test, and it may require renaming entries. (2) Alpha is dropped, not
> composited, so an 8-digit hex is judged on its opaque colour, which
> **over-reports** contrast. Callers must composite first.



A discovery check that composes the palette exactly as `RenderCSSFromSpecAction` does
and evaluates the pairs the layout actually emits: text on background, text on surface,
text on card_bg, primary as a foreground, and each ink on its own fill. This catches
families 1 and 2 — the systemic ones — deterministically and in milliseconds.

**Design note, and the reason this is a Phase rather than a paragraph:** the WCAG maths
lives in `platform/orchestration/actions/color_util.go`, and `discovery_checks` is
imported *by* `actions`. Copying the maths into a second package creates precisely the
two-things-that-must-agree drift this council reviews for. It wants a small shared
package (`platform/colour`) that both import, with `color_util.go` reduced to thin
wrappers. That refactor is the work; the check itself is small.

Route the finding as a `capability_gap` (remit.go): there is no palette-repair handler
and inventing one would re-create `bugs_closed/077`. Repainting a brand is an authoring
decision, and the honest output is a roadmap row, not a dispatch.

### Phase 3 — the rendered half, in the pipeline

`browser-runner-adapter` already exists. A post-deploy step that renders the changed
pages and files what it finds catches family 3 and broken images too. Gate it on the
deploy path, not the build path: the thing worth measuring is what the visitor gets.

### Phase 4 — close the consumption gap first if forced to choose

If only one phase is ever built, it should be **draining what is already detected**.
A detector whose output nobody reads is a more expensive way of not knowing.

Filed in its own right as **`bugs_open/115`**, by a concurrent thread in this same
workstream reading the same owner report. That the two of us found it independently,
hours apart, is itself evidence about how visible the gap is once anyone looks.

## Acceptance

- A dark site whose palette omits `card_bg` produces a finding before a human sees the
  page.
- `scripts/render_audit.py --sitemap` over the fleet's live sites is recorded here with
  a date, so the number is a baseline rather than an anecdote.
- The three parked `audit_finding_brief_fidelity` rows are either actioned or the item
  type stops being filed.

---

## 2026-07-27 (evening) — Phase 2 is BUILT, ENABLED, and the fleet baseline this file asked for

**Phase 2b done** (`6dd8667ea`): `check_palette_contrast` reads
`palettes.colours` via `site_specs.resolved_composition.palette_id` and audits
the pairings a stylesheet emits. **Enabled** in `design-discovery-agent`'s checks
array (22 checks now; backup `bak_ad_designdiscovery_20260727`). Safe to enable
before the roll — an unregistered check name logs a Warn and is skipped
(`discovery_checks.go:124`) — so it is inert until the binary carries it and then
starts working with nothing owed.

> **The deferral in the box above was wrong, and checking is what settled it.**
> I wrote that composing the palette needed a second loader that could diverge
> from `buildPaletteMap`. It does not: the merge RESULT is persisted in
> `palettes.colours`, so there is nothing to reimplement. *"I cannot do X yet"
> is a claim about the codebase* — this one did not survive one query.

**Which row it reads is the load-bearing decision, and the obvious choice is
wrong.** `design_intent.palette.reference_values` — the site's intent — was
CORRECT throughout the incident. Auditing intent would have reported a clean site
while the pages were unreadable. The defect was intent ≠ artefact, so the check
reads the composed artefact.

### The fleet baseline this file's acceptance section asked for

Measured 2026-07-27 across every site with a resolved composition — **7 failing
pairings on 7 of 10 sites**:

| site | worst pairing | ratio |
|---|---|---|
| dartsonline.com | primary as an ink on the background | **1.11:1** |
| robot-hands.com | same shape | **1.14:1** |
| oufe.com | same shape | **1.21:1** |
| webdesign.co.uk | accent as an ink | 2.13:1 |
| vetcomparison.uk | accent as an ink | 2.42:1 |
| relojistas.com | accent as an ink | 2.68:1 |
| idea.uk | muted text on the background | 3.35:1 |
| fundamentallyai.com | — | **clean** (repaired today) |
| gamesdesign.co.uk, vonc.com | — | clean |

**Three sites carry the identical near-black-ink-on-near-black-background defect**
that took three days and the owner's own eyes to find on the fourth. That is the
answer to "why wasn't this picked up": nothing was looking, and it is not rare.

### Acceptance, honestly

- ✅ *"A dark site whose palette omits card_bg produces a finding before a human
  sees the page."* Verified against fundamentallyai's real pre-repair palette:
  exactly the three failures the owner reported, silent on the repaired one.
- ✅ *"the fleet's live sites recorded here with a date"* — above.
- ❌ *"The three parked `audit_finding_brief_fidelity` rows are either actioned or
  the item type stops being filed."* **Untouched.** Still `bugs_open/115`, and
  still the phase this file argues should come first if only one is ever built.

### What Phase 2 still cannot see

A component that hard-codes an ink over a themed fill — family 3. It produced a
real regression the same day: repairing fundamentallyai's palette flipped
`primary` near-black → light blue and two components hard-coding `#fff` over it
went 17:1 → 2.32:1. **A green Phase 2 result does not mean the pages are legible.**
Phase 3 (`browser-runner-adapter` on the deploy path) is what closes that, and
`scripts/render_audit.py` already does the work standalone.

