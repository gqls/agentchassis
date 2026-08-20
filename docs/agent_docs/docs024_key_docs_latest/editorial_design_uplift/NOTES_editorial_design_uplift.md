# NOTES — editorial design uplift

Running record, append-only, newest at the bottom. Evidence, commands, what the
system actually said, and every misstep.

---

## 2026-08-20 — lane opened; Phase A1 findings are in

### The correspondence, and what it actually returned

**Eleven live design/experience agents** [MEASURED 2026-08-20]:
`visual-designer`, `brand-designer`, `feature-designer`, `site-design-planner`,
`design-audit-agent`, `design-discovery-agent`, `visual-design-auditor`,
`webdesign-agent`, `experience-planner`, `experience-approval-council`,
`experience-register-writer`. (§9 of the inline-imagery plan noted there is **no
agent literally called "vigilant designer"** — `visual-designer` is the closest.)

**Dispatched `design-audit-agent` at robot-hands.com**, correlation
`51404b33-5287-42cf-b74e-93b5f8d3ea29`. Input contract is just
`{site_id, domain}` (read from `call_visual_auditor.config.input_mapping` before
dispatching, not guessed). It spawns `visual-design-auditor` + a content auditor,
runs algorithmic checks AND an LLM visual audit. Three orchestration rows, all
COMPLETED.

**Algorithmic results:** `unlinked_components: 0`, `slot_mismatches: 2`,
`nav_stacked: 1`.

**Visual audit: 5 findings** — 2 high (colour, dark_section), 3 medium
(typography, spacing, colour).

### The finding that lands directly on what this lane just shipped

The auditor's **first high-severity finding is the hero's hardcoded colour**:

> "Hero section uses hardcoded rgba and hex values in inline style rather than CSS
> variables, breaking theme consistency and making global colour changes
> impossible" — current value
> `style="background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url('/assets/images/hero-home.jpg'); --hero-btn-ink: #0F1115;"`
> — suggestion: define `--hero-overlay-start` / `--hero-overlay-end` and
> `--hero-btn-ink` in the theme and reference them; `#0F1115` "does not exist in
> the declared palette and should map to `var(--color-primary)` (#1A1F2E)".

**This is the same overlay the owner just ruled should be the editorial default,
and the auditor is right about it.** The two are not in conflict: the ruling is
about *what the hero should look like* (image + semi-transparent overlay, not a
flat gradient), and the finding is about *where those two rgba values should
live* (the theme, not an inline style). So the ruling stands and this becomes
**Phase B's first concrete item** — tokenise the overlay in the shared `hero`
template, which improves every hero on the fleet, not just editorial pages.

Worth stating plainly because it is the argument for having asked at all: I would
have written a typography-first plan out of my own taste. The platform's own
auditor pointed at the component I had touched forty minutes earlier, for a
reason I had not considered.

### The other four, and which are ours

| finding | severity | ours? |
|---|---|---|
| hero hardcodes rgba/hex instead of theme variables | high | **yes** — shared `hero`, used by every editorial page |
| `brief-explanation-section` fallback `#0d0d0d` is the `code_bg` token, not `background` (#0F1218) | high | no — not on editorial pages, but it is a **palette-token discipline** finding whose class our components must not repeat |
| `surface_alt`/`background_alt` are both `#1a1a1a`, clashing with the dark blue-grey palette | medium | no, but same class |
| `tool-list-section` eyebrow uses `var(--color-primary-ink, var(--color-primary))`, and `--color-primary` is the dark background — **near-invisible text** | medium | no — but **`evidence-timeseries` and `evidence-chart` both use an eyebrow**, so check ours before shipping typography work |
| `--spacing-section` used inconsistently (shorthand vs single value) | medium | no, same class |

**The pattern across four of the five is one thing: fallback values that do not
come from the palette.** That is a discipline our new components must adopt from
the start rather than acquire as debt — and it is checkable, which makes it a
better Phase B acceptance criterion than "looks nicer".

### Honest limits of this evidence

- **It audited the SITE, not our pages.** Every finding names `page: index`, so
  none of them is a measurement of the editorial pages themselves. The editorial
  relevance is by *component* (`hero`) and by *class* (palette fallbacks), which
  is a weaker claim than "the auditor found these on our page" and must not be
  written up as the stronger one.
- **Not yet run:** `render-audit-agent` over the two editorial pages (contrast and
  overflow at the served artefact, including chart furniture under WCAG 3.0 per
  VIZ-011) and `compute_component_quality` on the editorial components. Those are
  Phase A2/A3 and they are the ones that would actually measure our pages.
- 5 findings is a small sample from one run on one site; the audit is LLM-backed,
  so a second run may not return the same five.

### Blocked, and not worked around

`features_open/035_FEATURE_component_hierarchy.md` — Fable-only by owner ruling
(twice). Dispatched with a full brief; failed on **"You've reached your Fable 5
limit"**. Fourth failure of the same kind across sessions. No substitution made.
Everything it needs to read is catalogued in the PLAN §2 and in
`news_editorial_features/NOTES` — it is blocked on capacity, not knowledge.
