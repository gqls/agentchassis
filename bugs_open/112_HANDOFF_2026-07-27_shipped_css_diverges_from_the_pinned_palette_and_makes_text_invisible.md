# 112 — the shipped CSS diverges from the site's pinned palette, and the divergence makes text invisible

**Filed:** 2026-07-27 by the brochure_component_library workstream, after the owner
looked at fundamentallyai.com on mobile and reported unreadable grey text, a missing
chart and thin imagery. Every complaint traced to one mechanism.
**Severity:** HIGH for any site it hits — this is not cosmetic. Headings render at
**1.21:1** contrast, i.e. invisible, on 5 of 10 pages of a live commercial site.
**Class:** structural (a pinned source of truth that the artefact does not follow, with
nothing comparing the two).
**Status:** OPEN, unowned. Cause established, not yet fixed.

---

## Symptom, measured on the served pages

`https://fundamentallyai.com/multi-agent-review-council.html`, the card
"Every decision leaves a record":

| element | colour | on | ratio | WCAG AA |
|---|---|---|---|---|
| card **title** | `--color-heading` → `#E4EAF2` | `--color-card-bg` `#ffffff` | **1.21:1** | fail (needs 4.5) |
| card **body** | `--color-text-muted` `#7E91A8` | `#ffffff` | **3.23:1** | fail |
| section **eyebrow** | `--color-primary` `#0E1B2E` | `--color-background` `#080E1C` | **1.11:1** | fail |
| section body | `#7E91A8` | `#080E1C` | 5.97:1 | pass |

Two invisibility faults in **opposite directions** on the same page: near-white text on
white cards, and near-black text on the near-black section background.

`info-card-grid` carries this on **5 of 10 pages** (index, capabilities,
model-fine-tuning, multi-agent-review-council, and via chrome elsewhere).
`evidence-chart` has it too — its `__figure` is `--color-card-bg` (white) and its
labels and values inherit `--color-text` (near-white). **The bars render and every
number is invisible**, which is exactly why the owner reported "I don't see the graph".

## Cause

The site pins its palette in `site_specs.design_intent.palette.reference_values`:

```json
{"primary":"#86ADDE","secondary":"#4A6C99","accent":"#C8902A","background":"#080E1C",
 "surface":"#111E33","text":"#E4EAF2","text_muted":"#7E91A8","border":"#1B2D47"}
```

A coherent dark scheme: **light** blue primary on a near-black background.

The **served** `/assets/css/styles.css` does not match it:

| slot | pinned | shipped | |
|---|---|---|---|
| `primary` | `#86ADDE` | `#0E1B2E` | **inverted** — light → near-black |
| `secondary` | `#4A6C99` | `#1A2E48` | **inverted** |
| `card_bg` | *(not in the pin)* | `#ffffff` | invented |
| `header_bg` | *(not in the pin)* | `#ffffff` | invented |

Two separate mechanisms produce this, and they compound:

1. **`render_css_from_spec_action.go` splits the palette in two.** Core slots
   (primary/secondary/accent/background/surface/text/text_muted/border) are
   *spec wins*; **specialised slots (heading, hero_title, cta_bg, card_bg, header_bg…)
   are *theme wins***. So a dark spec married to a light theme yields a dark core with
   **white cards** — and the text variables, calibrated against the dark core, are then
   used on those white cards by every component. No component is at fault; each
   correctly uses `--color-heading` and `--color-text-muted`.
2. **The stylesheet is written only by a `webdesign-agent` run and never regenerated**
   (`bugs_closed/072`). `render_css_from_spec` has exactly one caller, `webdesign-agent`.
   So the *pin is not applied on any other path*: once a stylesheet has shipped, a later
   `design_intent` change has no route to the artefact. That is why the core slots are
   stale as well as the specialised ones being wrong.

## Why nothing caught it

- The site has **zero** `discovery_check` / `design_issue` / `design_critique` items —
  ever. Fleet-wide only 4 sites have any, max 2 each: the design-discovery loop is
  effectively not running (`features_open/019`, parked by owner ruling).
- `check_forced_text_colors` exists and is enabled, but it looks for **dark background,
  no light text override** — dark-on-dark. It would catch the eyebrow and **miss the
  white-card case entirely**, which is the worse of the two. It also routes at a handler
  that has never existed (`bugs_open/077`).
- The contrast maths is already in the tree and unused here:
  `color_util.go` — `wcagContrastRatio`, `relativeLuminance`, `pickReadableOnBackground`.
- `validate_page_content` checks links, claims and numbers. **Nothing compares a
  rendered colour pair against a threshold**, so a page whose headings are invisible
  passes every gate and deploys.

## The compounding finding — the audit already said so, three days early

`site_work_items` for this site holds three `audit_finding_brief_fidelity` rows created
**2026-07-24**, all still `status='detected'`:

1. *"No chart component exists anywhere in the built inventory — zero of 27 components
   are chart components"* — we then built one by hand on 07-26 after the owner asked.
2. *"model-fine-tuning and multi-agent-review-council both share the identical component
   pattern… suggesting template repetition rather than meaningful differentiation."*
3. *"Only 2 of 27 components contain images — raising serious doubt that the
   illustration system is meaningfully present."*

Items 2 and 3 are, almost verbatim, two of the owner's four complaints on 07-27.

**Fleet-wide there have only ever been 3 rows of this item type, all on this site, all
`detected`.** `audit_finding_brief_fidelity` appears in the Go tree **only in
`discovery_checks/verifier_coverage_test.go`** — no agent definition consumes it, no
handler claims it. The audit runs, writes a correct finding, and the finding is
terminal on arrival. Same shape as `bugs_open/083` (detected findings never reach a
handler) and `bugs_open/077` (detection wired to a non-existent fixer) — **contribute
there, do not fork a third account.**

## Fix candidates, ordered by what closes the door

1. **Make the bad state unrepresentable: derive the specialised slots from the scheme,
   never from a theme of the opposite scheme.** If the core palette is dark, `card_bg`
   and `header_bg` must be computed from it (`surface`, or a lightened `background`),
   not inherited from a light theme. A "spec wins for core, theme wins for specialised"
   split has no defence against the two disagreeing about light vs dark — and the
   scheme guard (`bugs_closed/022`) already establishes the precedent that a scheme
   contradiction must be resolved, not merged.
2. **A contrast gate in `validate_page_content`.** The maths exists in `color_util.go`;
   the resolved variables are known at render time. Refuse to publish a section whose
   text/background pair is under 4.5:1 (3:1 for large). This is the check that would
   have made all of this impossible to ship, and it needs no screenshots, no model and
   no new machinery — only wiring. **Cheapest real win.**
3. **Give `audit_finding_brief_fidelity` a handler** (or route it to `capability_gap`
   as `077` did) so a correct finding stops dying in `detected`.
4. **Regenerate the stylesheet from the pin on a path other than a full webdesign run**,
   so a `design_intent` change can reach the artefact. Currently it cannot.
5. Re-run the brief-fidelity audit fleet-wide. It has run **once**, on one site, and was
   right about everything.

## How to verify a fix

Compute the ratios from the *served* CSS, not from the spec — the whole defect is that
those disagree. Script the four rows in the table above and require them all ≥4.5:1,
with a **negative control**: assert the check fails on today's `#E4EAF2` on `#ffffff`
before trusting it to pass on tomorrow's.

## What this is NOT

- Not a component defect. Every component correctly uses palette variables; the
  variables are wrong. Fixing components would be the wrong layer and would have to be
  repeated for each new one.
- Not `bugs_closed/072` (styles.css never regenerated) — that is *why the divergence
  persists*, and is a contributing mechanism, but 072 does not explain how the CSS came
  to disagree with the pin in the first place.
