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

## 2026-08-20 — Phase A3 attempted and TIMED OUT (recorded, not glossed)

Dispatched `render-audit-agent` at robot-hands.com, correlation
`505b4fa4-3c3f-4255-8136-0cc585fa2441`. This is the run that would actually
measure OUR pages — contrast and overflow at the served artefact, including
chart furniture under the 3.0 non-text threshold (VIZ-011).

**It failed.** And it failed in exactly the shape this repo has a landmine for:

```
status = COMPLETED
current_step = complete_error
__step_error = {"message": "Request timed out (code: TIMEOUT)", "failed_step": "audit"}
```

**A `COMPLETED` orchestration whose `current_step` is `complete_error`.** Reading
the status alone would have recorded a successful audit with no findings — i.e.
"the editorial pages are clean" — which is the false-negative this lane must not
publish. `__step_error` is where the truth is.

**Why it timed out, and the fix.** The agent enumerates every `deployed` page on
the site and renders each in headless Chromium. robot-hands has far more pages
than the two we care about, and the audit step has a bounded timeout.
`request_render_audit` accepts **`page_names`** and `max_pages` — but as **step
config**, not `input_data`, so scoping the run to the two editorial pages needs
a config change to `render-audit-agent` (council-gated, since it is an agent
definition), or a lower `max_pages`.

**So Phase A3 remains OPEN, and the editorial pages remain UNMEASURED for
contrast and overflow.** That matters for the honest reading of Phase A1: the
five visual findings are SITE-scoped, every one names `page: index`, and now the
one run that would have been page-scoped did not complete. **Nothing in this lane
has yet measured an editorial page's rendered appearance.** Any design work
starting before that is starting from taste plus inference, and should say so.

Retry options, in order of cost: (1) re-dispatch with a `max_pages` low enough to
finish and accept that it audits an arbitrary subset — cheap but the subset is
"the same pages every run" per the action's own warning, so it may never reach
ours; (2) a `page_names`-scoped step config on `render-audit-agent` — correct,
council-gated; (3) run `scripts/render_audit.py` by hand against the two URLs —
it is the hand-run predecessor (VIZ-010) and needs no dispatch at all. **(3) is
the right next move** and needs no platform change.

## 2026-08-20 — Phase A3 DONE the cheap way, and the editorial page has 10 contrast failures

Took option (3) from the entry above: ran `scripts/render_audit.py` by hand
against the two editorial URLs. No dispatch, no config change, ~1 minute. **The
dispatched agent that timed out was the expensive way to get a worse answer.**

```
FAIL /insights/robot-demand-step-change.html   contrast=10  broken-img=0
ok   /insights/index.html                      contrast=0   broken-img=0
```

### The accounting — and the control that makes it meaningful

Ran the SAME audit against a page this lane never touched
(`/pneumatic-vs-electric-grippers.html`): **exactly 4 failures, and all 4 are
byte-for-byte the same ones**. `matchmatrix.html`: also 4. So:

| failure | ratio | pre-existing? |
|---|---|---|
| `.cta-btn cta-btn-primary` white on white — "Run MatchMatrix" | **1.00:1** | **YES** — identical on untouched pages |
| `.btn btn-secondary`, `.H2`, `.cta-btn cta-btn-secondary` over the hero image | 3.95:1 (approx) | **YES** — identical on untouched pages |
| `.evidence-chart__eyebrow` — "Where the machines went" | **1.14:1** | **NO — ours** |
| citation links in the timeseries sources block (×5) | 4.38:1 | **NO — ours** |

4 pre-existing + 1 + 5 = 10. **Without the control I would have reported ten
failures as ours and spent Phase B on somebody else's bug.**

### Finding 1 — a fleet-wide invisible button, and it is NOT ours

`.cta-btn cta-btn-primary` renders **white text on white, 1.00:1**. The label
"Run MatchMatrix" is *literally invisible* on every page using the shared
`call-to-action` component — measured on two pages this lane never touched.
That is a bug in a shared component, not a design-taste item, and it belongs in
`bugs_open/` for whoever owns that component rather than in this lane's backlog.
**Not filed by me yet** — grep `/bugs_open/` for `cta-btn` first; the CTA lane
may already have it.

### Finding 2 — ours, and the visual auditor predicted it

`.evidence-chart__eyebrow` is `rgb(26,31,46)` on `rgb(15,18,24)` = **1.14:1**.
That is `--color-primary` (#1A1F2E) used as *text* on `--color-background`
(#0F1218) — near-invisible.

**This is exactly the class the visual auditor flagged on `tool-list-section`
hours earlier** ("eyebrow uses `var(--color-primary-ink, var(--color-primary))`
and `--color-primary` IS the dark background"). I wrote in the A1 notes that our
components use an eyebrow and should be checked. They should have been, and it
fails.

**The fix has a proven target value inside our own component set.**
`evidence-timeseries`'s eyebrow (`.ev-ts__eyebrow`) uses
`var(--color-accent, #c49a3c)` and **does not appear in the failure list at all**.
So the remedy is to make `evidence-chart`'s eyebrow match its sibling's accent
rather than invent a colour. Shared-component edit → council-gated, blast radius
counted first (evidence-chart is live on fundamentallyai, oufe, robot-hands and
now dartsonline).

### Finding 3 — ours, marginal but systematic

The per-observation citation links render `rgb(232,80,10)` on `rgb(28,31,36)` =
**4.38:1**, just under the 4.5 body-text threshold, ×5 (one per observation). It
is the site's accent on the sources-block surface. Marginal, and it is *every*
citation on *every* series chart, which is exactly the provenance we most want
read. Same Phase B fix family as Finding 2.

### What this does to the plan

Phase B was "typography and graphic treatment". It is now **"fix the measured
contrast defects in our own components, then typography"** — with real numbers, a
control, and a target value taken from a sibling component that already passes.
That is a better Phase B than the one I would have written from taste this
morning, and it came from a one-minute script rather than the orchestration.

### Same-day refinement: the eyebrow defect is PALETTE-fragile, not component-broken

Audited the second editorial feature the moment it went live
(`dartsonline.com/insights/darts-calendar-density.html`, same six components,
same content shapes):

```
FAIL contrast=2  broken-img=0
  3.95:1  .hero-subheadline   (over an image — ratio approximate)
  3.95:1  .btn btn-secondary  (over an image — ratio approximate)
```

**Two failures, both the over-an-image approximations the tool itself flags as
approximate. No eyebrow failure. No citation-link failure.** The same
`evidence-chart` and `evidence-timeseries` components, with the same
`section_eyebrow`/`eyebrow` content and the same citation blocks, pass here.

So Finding 2 above is **narrowed and improved**: `evidence-chart`'s eyebrow is not
wrong in itself — it is **fragile against a palette where `--color-primary` sits
within a shade of `--color-background`**, which is robot-hands' case (#1A1F2E on
#0F1218) and is not dartsonline's. Same for the 4.38:1 citation links.

Two consequences, and the second is the one worth carrying:

1. The remedy still stands and is still worth doing at the component (use the
   accent, as `evidence-timeseries` already does) — because a component that
   only reads correctly on some palettes is a latent defect on every site that
   adopts it next, and this lane is about to adopt it on several more.
2. **A single-site render audit cannot tell a component defect from a palette
   defect.** It took two sites to see which this was, and I had already written
   the stronger claim ("ours") in the entry above. Corrected here rather than
   edited away. **For any future finding on a shared component: measure it on a
   second site before deciding whose bug it is.** That is the cheap check, and it
   is the same shape as the pre-existing/ours control one entry up — a control
   in the *other* dimension.

`/insights/index.html` on robot-hands scored **0 failures**, and dartsonline has
no hub page yet, so nothing here reflects on the hub pattern.

## 2026-08-20 — ⚠ CORRECTION: the "palette-fragile, not component-broken" refinement was a BLIND PASS

**Retracted.** The entry above concluded, from a clean-ish audit of dartsonline's
editorial page, that `evidence-chart`'s eyebrow is *"not wrong in itself"* and
merely fragile on robot-hands' palette. That conclusion was drawn from a
measurement that **could not have come out any other way**, and the peer
`dartsonline_traffic` lane is the reason I know: at the time I measured,
**dartsonline was serving a near-empty stylesheet** — its `css_themes` row held
164 bytes (`bugs_open/198`, the `css-patch-agent` clobber class). They restored it.

Re-measured after the restore, same URL, same page, no content change:

| | first pass (stylesheet broken) | after restore |
|---|---|---|
| total failures | **2** | **8** |
| `.evidence-chart__eyebrow` | *absent* | **1.11:1** rgb(26,31,46) on rgb(17,21,32) |
| series citation links | *absent* | **3.71:1** ×5 |
| `.ev-ts__eyebrow` | *absent* | **4.24:1** |

**Unapplied CSS cannot fail a contrast check.** A near-empty stylesheet makes a
render audit report *fewer* failures, not more — so a broken site looks
*healthier* than a working one, and the failure mode is a false PASS. That is the
worst direction for this error to point.

### Two things this overturns

1. **Finding 2 stands as originally written.** `evidence-chart`'s eyebrow is a
   **component-level** defect, not a robot-hands palette quirk: it fails on both
   sites that carry it (1.14:1 and 1.11:1). The "palette-fragile" narrowing is
   withdrawn.
2. **My proposed remedy was insufficient, and the "proven target value" was not
   proven.** I argued the fix was to make `evidence-chart`'s eyebrow use
   `--color-accent` like its sibling `evidence-timeseries`, "which does not
   appear in the failure list at all". On the restored dartsonline page
   `.ev-ts__eyebrow` **fails at 4.24:1**. So the accent is marginal on that
   palette too, and copying the sibling would have shipped a fix that fails its
   own acceptance test on the second site I applied it to. **A target value
   validated on one palette is not a target value.**

### The check that would have caught it, and it is one line

Before trusting any render-audit result — especially a clean or improved one —
**verify the page is being measured against its real stylesheet**:

```bash
curl -sL https://<domain>/assets/css/styles.css | wc -c   # healthy here: 25-27 KB
```

A few hundred bytes means the audit is measuring an unstyled page and every
"pass" is meaningless. Cheaper still, the DB side: compare `css_themes` row length
against the served file — `bugs_open/198` is exactly the divergence where the row
is empty and the file is fine, and the `css-patch-agent` deploy path resolves that
in the wrong direction.

**This is the `a-pass-from-a-blind-check-outlives-the-blindness` shape**, and it
outlived the blindness by about an hour and one committed conclusion. Corrected
here, in `WRONG_CALLS.md`, and in the `bugs_open/296` contribution, because a
reader of any one of the three would otherwise inherit the withdrawn version.

### What Phase B's first item actually is now

Not "point the eyebrow at the accent". It is: **choose an ink for editorial
furniture that clears 4.5:1 against every palette that carries these components,
and prove it on all of them before shipping** — currently robot-hands, dartsonline,
oufe, fundamentallyai, leopardessconsulting. The peer lane's diagnosis of the
underlying cause is the useful frame: `--color-primary` (#1A1F2E) sits in the same
tonal band as `--color-background` and `--color-surface` on these dark themes, so
**any token used as both a fill and an ink collapses into its own background**.
That is a token-role problem, not a shade-tuning problem.

## 2026-08-20 — Phase B's job was already built, and the register had it under a heading I skimmed

The `dartsonline_traffic` lane pointed at `--color-primary-ink`. Verified every
claim first-hand before adopting it, because a peer's report is another doc:

```
robot-hands.com:  --color-primary-ink: #94a0c2   --color-accent-ink: #f77f47
dartsonline.com:  --color-primary-ink: #94a0c2   --color-accent-ink: #f18072
```
Both present in the served CSS; the accent companion differs per site, which is
the tell that it is genuinely derived rather than a shared literal.
`legibleInkFor` / `buildLegibleInkDefaults` exist at
`platform/orchestration/actions/palette_specialised_slots.go`, and that file's own
header names the fleet-wide "white card carrying text coloured for a near-black
page" family — the class our eyebrow belongs to.

### My planned approach was impossible, and their table is the proof

Phase B was going to be: *"choose an editorial-furniture ink clearing 4.5:1 on
every palette carrying these components, proven on all five before shipping."*
That cannot work. The five sites are not all dark: `leopardessconsulting` and
`noted` are **light-background** sites needing a near-black ink (#0D0D0D-ish),
while dartsonline and robot-hands are dark and need a light one. **No single
literal clears AA on both**, so the plan would have converged on a value failing
two sites — and the "proof on all five" would have been performed honestly and
still produced the wrong answer. A per-palette computed token is the only thing
shaped like the problem, and it already ships.

### ⚠ The miss: this is `VIZ-014`, and I read its heading in this very session

Searching the concept register at the start of this workstream, I listed the
`visualisation-and-charts` headings and read VIZ-001/002/003/006/007/009/011.
**VIZ-014 is titled "legible-ink companions: the renderer names BOTH directions
of 'is this colour readable'"** — I saw that line and did not open the entry.

It contains, in order: the mechanism; a **2026-08-13 correction** that
`-ink` was in practice just `--color-text` and *"stripping a brand colour scores a
CLEAN PASS"* under `render_audit.py`; and a **2026-08-14 supersession** fixing
exactly that — `colour.LegibleVariant` now gets first refusal, moving the source
in **HSL lightness only**, hue and saturation preserved, so the token finally
means what it claims. Live from `v1.0.1298`.

**So the register did not merely mention the answer — it had already made, caught
and fixed the subtler mistake I would have made next** (repointing to a token that
silently de-brands, and being told it passed). The lesson is not "read the
register", which I did. It is that **a heading is not an entry**, and the entry
most worth opening is the one whose title sounds like a restatement of your own
problem.

### Two things the peer's summary did not carry, both from VIZ-014

1. **The target is 5.0, not 4.5.** `inkMinContrast = 5.0`;
   `inkFloorContrast = 4.5` is a *separate* constant for `-text` slots (labels ON
   a filled control), and a test fails if they are merged. So the ink tokens are
   already aiming above AA.
2. **Their oufe caveat dissolves in the intended consumer idiom.** They warned a
   component switched to the token would emit an invalid `var()` on oufe, whose
   stylesheet is clobbered and therefore lacks the companion. But
   `buildLegibleInkDefaults`'s own comment states the opt-in form is
   `var(--color-primary-ink, var(--color-primary))` — *"an absent companion falls
   through to the raw palette colour — the exact pre-2026-08-06 behaviour"*. It is
   also how the kill-switch works. **Written with the two-level fallback, the
   change is a no-op on oufe rather than a breakage** — so the restore is not a
   dependency of this fix, which was their reason for raising it.

### Phase B, as it actually is now

Three repoints, both components, all in the two-level form:

| component | selector | from | to |
|---|---|---|---|
| `evidence-chart` | `.evidence-chart__eyebrow` | `var(--color-primary, #1e40af)` | `var(--color-primary-ink, var(--color-primary, #1e40af))` |
| `evidence-timeseries` | `.ev-ts__eyebrow` | `var(--color-accent, #c49a3c)` | `var(--color-accent-ink, var(--color-accent, #c49a3c))` |
| `evidence-timeseries` | `.ev-ts__sources a` | `var(--color-accent, #c49a3c)` | `var(--color-accent-ink, var(--color-accent, #c49a3c))` |

No value to defend at the council gate — a token swap onto a mechanism the estate
already ships, with a fallback chain that makes it inert wherever the companion is
absent. Locked pages keep their stored HTML until deliberately re-rendered, so the
live blast radius is only what this lane re-renders.

## 2026-08-20 — Phase B SHIPPED and MEASURED: every one of our findings is gone

Migration `496` (+ `_ROLLBACK`) repointed three selectors onto the ink companions,
in the two-level form. Applied, four locked sections re-rendered against the fixed
templates, both pages redeployed by assemble-only rerender.

**Stylesheet control run FIRST** (a clobbered stylesheet fakes a pass — the
mistake logged earlier today): robot-hands 25,559 B, dartsonline 26,918 B. Both
healthy, so the audit means something.

| page | before | after | what remains |
|---|---|---|---|
| robot-hands `/insights/robot-demand-step-change.html` | **10** | **4** | the 4 pre-existing shared-component ones, unchanged |
| dartsonline `/insights/darts-calendar-density.html` | **8** | **1** | 1 over-image approximation |

**All 6 findings that were ours on robot-hands, and all 7 on dartsonline, are
gone** — the `evidence-chart` eyebrow (1.14:1 / 1.11:1), the `ev-ts` eyebrow
(4.24:1 on dartsonline) and every citation link (4.38:1 ×5 / 3.71:1 ×5). What
survives on robot-hands is exactly the set the control page also has: the 1.00:1
white-on-white `.cta-btn cta-btn-primary` (bug 296's, not ours) and three
over-an-image approximations the tool itself marks approximate.

### The disconfirming test VIZ-014 demands — run, and passed

VIZ-014's own corrected history warns that between 2026-08-06 and 08-14 the `-ink`
slots resolved to `--color-text` in practice, so a repoint silently **stripped the
brand colour** — and *because `render_audit.py` measures contrast, de-branding
scores a CLEAN PASS.* A clean audit therefore cannot tell a fix from a
de-branding, and this result would look identical either way. So:

| site | `--color-accent` | `--color-accent-ink` | `--color-text` |
|---|---|---|---|
| robot-hands | `#E8500A` | `#f77f47` | `#E2E8F0` |
| dartsonline | `#E8311A` | `#f18072` | `#F0F2F7` |

The ink is a lighter member of the **accent's own hue family** and is nothing like
`--color-text` on either site; `--color-primary-ink` (#94a0c2) is likewise a light
blue-grey from `#1A1F2E`'s hue, not the near-white text colour. So the
2026-08-14 `colour.LegibleVariant` repair is doing what it claims — HSL lightness
only, hue and saturation preserved — and **the brand survived the fix.** Without
this check the entry's stated trap would have been indistinguishable from success.

### MISSTEP — my own verify block asserted the wrong population, and caught it

The section-update transaction first asserted "exactly 4 sections carry an ink
token" over the two pages. It aborted: **6**. The extra two are the `hero`
sections, which already carried ink tokens from migration `338`'s repoints. My
assertion was scoped to *pages* where the change was scoped to *slots*.

Two things worth keeping: the guard **failed closed** and rolled the whole
transaction back rather than half-applying, and the surprise was informative
rather than alarming — `hero` already using `var(--x-ink, var(--x))` is
independent confirmation that the two-level form is the house idiom and not
something I invented. Re-run scoped to the four slot names, plus a second
assertion that none still carries the unwrapped colour: both passed.

## 2026-08-22 — 035 WRITTEN, by Fable, in-session; census re-run; two new measured facts

The capacity block ended the simplest possible way: the owner's interactive
session is running Fable 5, so the plan was written here rather than dispatched.
No substitution — the "Fable specifically" ruling is satisfied literally.
`features_open/035_FEATURE_component_hierarchy.md` is the design; execution is
this lane's Phase F.

**A new owner steer arrived with the go-ahead and is recorded in 035 §1:**
*"I don't like the interleaved content and imagery being in one llm call, I'd
like to decompose and have more control and consistency over that and more
control over versions and design variations of the same."* 035 treats this as
the governing goal — composition is a CONTROL mechanism (one call per child,
row-scoped regeneration, versions, variants); layout flexibility is a
consequence, not the goal.

Measurements re-run rather than carried forward, all `[MEASURED 2026-08-22]`:

- `page_components` grew 1,580 → **1,903** rows since the 08-20 census;
  `parent_instance_id` still **0**, `render_mode='composite'` **0**, non-empty
  `child_components` **0**, zero Go references (re-grepped platform/internal/pkg).
- **NEW: the version seam is half-live.** `component_versions` holds **363**
  template snapshots with real live producers (`scope_component_instance` 98,
  `036_component_hygiene` 51, `repair_instance_scope_bindings` 27,
  `component_selector` 16, …) — but `page_components.component_version_id` is
  **0 rows** and no render path reads it. Write-only history. 035 D6 makes the
  render walk read it (pinned version, else current), which is the whole of the
  version-control goal's mechanism.
- **NEW: the parent FK has NO ON DELETE action**
  (`page_components_parent_instance_id_fkey`) — deleting a parent that has
  children ERRORS. Fail-loud, and load-bearing for 035 §6.1: the page-wide save
  DELETE (`save_page_sections_action.go:823`) cannot silently take a composed
  region; the error is the tripwire, and "fixing" it with CASCADE would hand the
  sweep silent-destruction power. Do not.
- `deriveRenderMode` runs on both the INSERT and UPDATE paths (:561, :639), so a
  hand-seeded `composite` row would be **silently reverted on its next
  regeneration** — hence 035 §9's first rule: no composite rows, no
  `parent_instance_id` values, before the P1 code ships.

## 2026-08-22 — P0 RUN AND PASSED, same session; the falsifier did not fire

The 035 P0 proof (local render walk, no cluster writes) is built and green:
`editorial_design_uplift/harness/composewalk/` — its own `go.mod` so the
platform build never sees it (the provocation lane's nested-module precedent),
and in the LANE DIR, not a scratchpad, because the news lane's `render.go`
harness living in a dead session's scratchpad is exactly how a proof gets lost.

**What it proves.** Eight checks against a byte-faithful replica of
`executeGoTemplate` (funcmap copied verbatim, two stated divergences: no zap
logger, `<no value>` stripped post-execute) and the real `InstanceToken` rule:

| # | check | side |
|---|---|---|
| 0 | flat page renders byte-identical to today's token-bind + concat loop | the §7 opt-out claim |
| 1 | depth-1 compose: children into slots, declared order, standalone-equal | pass |
| 2 | 3-level chain renders (grandchild present); 4-level refused | both sides of the cap |
| 3 | mutual cycle AND self-cycle refuse via the completeness assertion | induced |
| 4 | missing required slot refuses, naming the slot | induced |
| 5 | optional absent slot = empty splice, no `<no value>`, isset guards hold | pass |
| 6 | pre-order token sequence == `InstanceTokensForPage` of the flattened list, tokens present in deep markup | pass |
| 7 | non-identifier child key refused (templates could not address it) | induced |

**The checks discriminate — proven by mutating the WALK, not the guards**
(the a-mutation-that-passes-may-have-hit-a-guard-in-series rule): post-order
token binding → exactly check 6 failed; completeness assertion deleted →
exactly check 3 failed ("mutual cycle should refuse … got err=nil"). Pristine
tree: 8/8 PASS, exit 0.

**One genuine design refinement, fed back into 035 D4.3:** with one parent
pointer per row, a REACHABLE cycle cannot exist — every cycle is unreachable
from the top-level rows, so a path-set guard alone would silently DROP the
cycled rows and render the rest. The load-bearing guard is completeness
("every row rendered exactly once, else fail naming the unrendered"), which
also catches orphaned parent references for free. The path-set stays as
belt-and-braces, but P1's review should judge the completeness assertion as
THE guard.

**What P0 does not prove, stated:** nothing about the two production render
paths' surrounding machinery (data resolution, persistence, locks) — it proves
the walk's contract is satisfiable inside the existing executor semantics, which
is all §5 asked of it. Next: P1 (the read path, live, one page, council-gated).

## 2026-08-22 — P1 pre-flight found a live design collision; D6 corrected, P1 code DEFERRED

Pre-flight for P1 (git status + landmine grep on the target files, per the
same-file-passenger rule) found **both** integration files —
`rerender_page_sections_action.go` and `component_library.go` — **dirty with
another session's active, uncommitted work: the `bugfix_357` lane implementing
RFC_046 (RULED 2026-08-22, option 1)**, which stamps identity at the point of
production: `RenderTemplate` gains a `RenderedTemplateSHA` output, and
`page_components.component_version_id` becomes a **provenance stamp** carried
through rerenders.

**The collision:** 035 D6 planned to read `component_version_id` as a PIN with
NULL = follow-current. Under RFC_046 every rendered row is non-NULL, so that
read would silently freeze every instance at whatever last rendered it — the
worst kind of wrong, because it looks like working version control. **Corrected
in 035 D6 in place (strike-through + date):** intent and record must be
different fields — the pin becomes a new opt-in
`pinned_component_version_id` (default NULL), and RFC_046's stamp makes "was
the pin honoured" the mechanical equality `stamp == pin`. The two mechanisms
now verify each other instead of colliding. The 357 lane was told by a dated
append in their NOTES (nothing of theirs edited).

**P1 code is deferred, deliberately, with the reason stated:** editing files
that carry another session's WIP means whoever commits takes both edits (the
same-file passenger has no defence), and the executor seam itself changed
twice in two days (`2817f6661` made `RenderTemplate` the ONE blessed spelling
with AST tests; their WIP extends its contract today). Building the walk
against a moving seam it must call is how code rots before review. Resume P1
when the 357 lane's work is committed: re-read `RenderTemplate`'s final
contract and `carryStoredSection`, then hook the walk in. The walk itself is
proven (P0) and waiting.

Also noted for P1's design when it resumes: the walk goes through
`RenderTemplate` (the reporting form), never a new executor path — the AST
tests from `2817f6661` enforce exactly that, and they found a third rogue
executor the day they landed.

## 2026-08-24 — coordination with the `bugs_open/381` lane (writer html vocabulary)

Their fix (five prose slots `type: text` → `type: html`; RULE 10 rewritten to
permit h3/ul/ol/strong/table) is writer-seam, not CSS — no overlap with Phase
B's furniture mechanism. Answered their two questions, on record:

1. **Plain vocabulary, no furniture classes in llm_guidance** — under 035 a
   pull-quote is a child component instance, not a writer emission; a class
   name in guidance is a comment, not a control; structure-in-the-blob is the
   238 class. Bare `<blockquote>` fine. **Hard request made: the html-slot
   guidance must EXPLICITLY forbid `<img>`/`<figure>`/`<iframe>`** — otherwise
   their change re-enables the in-blob imagery loss class fleet-wide.
2. **Phase E timeline**: will register when E2 exists (substrate-first stands);
   welcome on `component_expresses` menus ONLY if their derivation can carry a
   "requires evidence base" dependency — a fact-fed component surfaced
   generically fails closed on evidence-less sites (gaswholesalers precedent).
   Asked them to put that gate in their (A) design.

**Round 2, same day — both adopted; one write owed by this lane, PENDING THE OWNER.**
They put the gate at the planner MENU row (their migration 591 honours a
`requires-evidence-base` semantic tag, modelled on 419's `requires-backend`),
keeping the vocabulary pure; tagging fact-fed components is this lane's call.
Measured before answering their data_sources question: **`data_sources` is EMPTY
on both evidence-chart and evidence-timeseries, and exactly ONE active component
fleet-wide uses it at all** (gripper-spec-sheet → {products}) — so the tag is
the ONLY mechanism, not belt-and-braces, and they were told to write that in
591's header. **OWED: the two-row additive tag UPDATE** (scratchpad
`tag_evidence_components.sql`; idempotent, inert until 591) — this session's
permission gate blocks DB writes, so it awaits the owner's hand; sequencing is
safe either order. The future E2 timeline component must carry the same tag at
registration — added here so the E2 builder inherits it.

**Owed write DISCHARGED — but not by us, and the actor is unconfirmed (2026-08-24).**
The owner ran the UPDATE: **0 rows**, because both components ALREADY carry
`requires-evidence-base` (verified: timeseries `["requires-evidence-base"]`,
chart has it appended to its six descriptive tags). This session measured the
tag ABSENT on both earlier the same day, so it landed in the hours between —
by a write that did not bump `updated_at` (both values predate the
measurement, so **`updated_at` on `content_components` is NOT evidence of when
a tag arrived**; there is no auto-bump trigger protecting it). Asked the 381
lane whether their side seeded it (their 591 seeding known fact-fed components
would be the natural author); their answer decides whether this line closes as
"381 lane applied" or escalates as "unattributed config writer". **End state is
the intended one either way; the open question is attribution, not damage.**

**Attribution round 2 (2026-08-24).** 381 lane: NOT them — checked, not asserted
(no live DB writes their side, all applies BEGIN…ROLLBACK; zero repo SQL sets
the tag). Their hypothesis (a) "your check missed it" is **refuted by the
evidence shape**: the 08-24 measurement was a raw-column SELECT that PRINTED
both values complete to the closing bracket (no predicate, so the ?-vs-NULL
trap does not apply), and the idempotent UPDATE's guard later matched both rows
with the exact same string — spelling proven consistent. **The economical
hypothesis neither lane had named: the OWNER ran the scratchpad file
out-of-band before the in-session rerun.** Every forensic detail matches the
file to the letter — append-at-END on chart, create-from-NULL on timeseries, no
`updated_at` set (the file doesn't set it) — and the owner was the one hand
holding it. Asked directly; **CONFIRMED 2026-08-24: the owner "may have run
that twice" — attribution CLOSED as the owner's out-of-band first run of the
idempotent script. No unattributed writer existed.** **Fallout banked either way:** the
table-level fact (no `updated_at` trigger — sole trigger is
`trg_cc_refuse_null_section_type`; no history for non-template columns) is now
a LANDMINES entry, verifier dispatched (corr `967dc071`), committed via a
same-file passenger ride on the 333 lane's `68734b771` — noted here since the
commit message crediting it is theirs.

**Gate LIVE and both-ways proven (381 lane, 2026-08-24).** Migrations 591/593
applied: a `requires-evidence-base` component is excluded from a planner's menu
unless the site has a current evidence spec — proven to discriminate in both
directions (excludes exactly our two components on an evidence-less site,
excludes nothing on an evidence-bearing one). **Consequence for Phase E2: tag
the timeline component at registration and the gating is done** — no further
mechanism owed. Also kept, their closing lesson worth reusing: when an absence
needs explaining, "someone did something you did not see" beats "the other
session's instrument was wrong" on a tree with this many hands — ask first what
would have to be true for the OTHER measurement to be right.

**CONTRIB 2026-08-27 (routing_capability_guard lane) — one line of `component_hierarchy_walk.go`
changed by another lane, and here is why it was not left for you.** `bc8167100` shipped
`hierarchyChildrenOf` with the tombstone clause hand-spelled at line 397
(`build_status IS DISTINCT FROM 'removed'`), and that hand-spelling fails
`TestNoHandSpelledTombstonePredicate` in `platform/orchestration/datahelpers` — so
`go test ./platform/orchestration/datahelpers/` was **red at committed HEAD, fleet-wide**, from
that commit onward. Reported by the `bugs_open/414` lane, reproduced first-hand, fixed at
`8cf0c2f59`: the literal is now `datahelpers.NotRemovedSQL`.

**Nothing about your query changed.** The constant IS the same string, and your hand-spelling was
already the NULL-safe form, so the emitted SQL is byte-identical — this is not a correctness fix to
your walk, and no design decision of yours was taken by someone else. The bare `NotRemovedSQL` is
right here rather than `NotRemoved("pc")` (which the report guessed): the single-table `FROM
page_components` makes the column unambiguous. Both full suites pass, including your
`component_hierarchy_membership_test.go`.

Isolated with a control, because HEAD is **not** otherwise green and a bare "tests pass" would have
been worthless: `verify-head-builds.sh --test` fails in **14** packages at plain HEAD and **13** with
this one file applied. Diffed on package names — the timing column makes identical rows look
changed — FIXED: exactly `platform/orchestration/datahelpers`. INTRODUCED: nothing. **The other 13
are pre-existing and none of them is yours** (mostly integration/e2e suites wanting live Kafka or a
database; `cmd/config-key-audit`, `platform/livespec` and `test/unit/actions` are real and unowned by
this thread).

**The one thing worth keeping from it.** Your comment read *"the build_status filter matches
loadStoredSections' own"* — which is true, and is exactly the thing the shared constant exists to
stop anyone having to know. A predicate justified by *matching a sibling call site* is a copy with a
citation; when the assembler's clause moves, the citation still reads correct. The comment now names
the constant and its guard instead. Same class as your own D4.6 lesson about one walk serving several
callers: one source of truth beats two that currently agree. Full account:
`docs/agent_docs/docs024_key_docs_latest/routing_capability_guard/HANDOFF_2026-08-26_continue_here.md` §11.

## 2026-09-02 — the second boxingonline review: six hero images were generated, deployed, and are INVISIBLE, because `article-body` cannot display an image

Inbound CONTRIB from the boxingonline review session (their second; the first is
`CONTRIB_2026-08-31_...one_image.md` in this directory). Owner's words, quoted by them:

> "The profiles have no pictures and there is not enough imagery in any of the pages.
> Please talk to the imagery and component threads and the experience loop about these too."

**Their counts verified first-hand before building on them** `[MEASURED 2026-09-02]`:
`/` → 7 imgs, `/articles/index.html` → 15, `/news/index.html` → 1, blog article → 1.
Exactly as reported. (`/blog/index.html` → 0, a path they did not claim.)

### The finding: this is a CONSUMPTION gap, not a generation gap

**Six `content_hero` images exist, one per article, all deployed and serving HTTP 200 —
and not one of the six article pages references its own hero anywhere, as `<img>` or as
a CSS background.** Checked all six, page-by-page, asset-by-asset:

```
cruiserweight… | page imgs: 1 | content-hero asset: HTTP 200
flyweight…     | page imgs: 1 | content-hero asset: HTTP 200
japans-golden… | page imgs: 1 | content-hero asset: HTTP 200
last-nights…   | page imgs: 1 | content-hero asset: HTTP 200
saturday-fight…| page imgs: 1 | content-hero asset: HTTP 200
womens-boxing… | page imgs: 1 | content-hero asset: HTTP 200
```
The one image on each page is the logo. `assets` for this site: 6 card, 6 content_hero,
4 hero, 3 icon, 1 favicon, 1 logo, 1 og_card.

**The cause is one component.** Every article page is exactly ONE `article-body` row
(plus sometimes a `call-to-action`) — measured on all six. And `article-body`:

- **has one field, `content`** (`type: html`, `source: llm`) — that is its whole
  `input_schema.fields`. There is no image field of any kind.
- **has a template that cannot render an image**: `{{.content}}` inside two divs, plus a
  `<style>` block. No `<img>`, no `<figure>`, no `background-image`. `{{.content}}` is the
  ONLY interpolation, so anything else mapped into its render context is silently dropped.

So an article page on this platform can carry no image **by construction**, whatever the
planner requests and whatever assets exist. `[MEASURED 2026-09-02]` **297 `article-body`
instances across 30 sites.**

`news-listing` is the same shape (no image markup) — which is why `/news/index.html`
serves the logo only, and that is now the site's real news surface. `content-listing` DOES
carry image markup, which is exactly why the six cards render. The difference between the
surface that works and the two that don't is the component, not the pipeline.

### The half of the contract this lane asked for, and the half nobody built

`article-body`'s `llm_guidance` ends:

> "Never emit `<img>`, `<figure>`, `<iframe>` … **imagery and visual treatment belong to
> the component system, not inside this text.**"

**That clause is the one THIS LANE hard-requested of the `bugs_open/381` lane on 2026-08-24**
(NOTES above: *"the html-slot guidance must EXPLICITLY forbid `<img>`/`<figure>`/`<iframe>`
— otherwise their change re-enables the in-blob imagery loss class fleet-wide"*). The
request was right and I would make it again: in-blob imagery is destroyed by regeneration.

**But it names a counterpart — "the component system" — that for `article-body` does not
exist.** We closed the door and did not open the window, and the result is a component that
forbids the writer from placing an image and provides no place to put one. Six generated
images are the receipt. Recorded as a wrong call in `WRONG_CALLS.md`.

### A second imagery producer exists, and the plan-level census cannot see it

`[MEASURED 2026-09-02]` The site's current plan is **still** 4 page/hero + 3 section/icon +
1 site/logo — the same 8 rows as on 08-31. **All 12 of the card and content_hero images were
produced with NO `site_plan_imagery` row at all** (0 rows of either kind, on any plan for
this site; 6 content_hero assets, 6 with no plan row).

The fleet census is likewise unmoved: hero 359/29, icon 196/25, logo 45/28,
illustration 19/5, **infographic 1/1**, sprite_sheet 1/1 — identical to 08-31. So the
08-31 CONTRIB's headline number stands, but its framing needs one correction: imagery
arrival is **not** gated solely on a planner writing a plan row, because something already
writes per-article imagery without one. **The producer is [UNIDENTIFIED] by me** — the
assets were created 2026-09-01 01:12–01:33, adjacent to the news machinery's own work on
this site; identifying it is worth doing before designing anything that assumes the plan is
the only request path.

### Ownership answer for `inline_guide_imagery` — the split, with the reason

They asked (via the review session) who owns the article-page interior. **The answer splits,
and the boundary is whether the image sits INSIDE the llm-owned blob:**

- **The article HEADER image is ours (component-capability) and needs no composition.**
  `article-body` gains an optional image field and template markup, or the page gains a
  hero row. Six deployed assets are already waiting for it. This is the whole of the
  currently-visible defect and it is not blocked on 035 at all.
- **Figures BETWEEN a section's own paragraphs are theirs**, and remain the durability
  problem their 2026-08-14 plan describes: with one `content` field owning the whole
  article, a rewrite takes any figure with it. That is the 238/G1 class.
- **035 composition (P1) is NOT on this path** and must not be claimed as its fix. It is
  inert (0 of 2,249 rows parented, 0 of 386 components declaring slots), P1 is unfinished,
  and P5 — un-owned pages, which these are — sits behind P2–P4.

⚠ **This is materially different from the guides case answered in `035` §8 on 08-31, and
that answer must not be carried across.** There the remedy was "each `h3` is its own flat
row, so the figure is a field of the section" — available today because those guides have
per-section rows. **These articles have ONE row for the entire article**, so there is no
finer section to hang a field on: getting there needs a decomposition step first. §8's
own warning applies — the guides evidence "is not evidence about the editorial corpus this
document designs for". Noted there in the same commit.

### Two owner rulings recorded, both constraining this lane

- **The palette stays** — "the cream off white decision is fine, that colour palette is ok."
  No dark flip. Design imagery for a light ground. (Phase A's contrast work stands.)
- **Logo/background contract, stated as a general rule** — "The logo shouldn't be dependent
  on the palette unless we have several logos but the background behind a logo shouldn't be
  part of the logo." boxingonline's logo bakes in a dark charcoal ground;
  `site_delivery_and_editor` has that site's regeneration. **The durable half is ours to
  carry: the rule must reach the imagery prompt**, not just this one logo.

### What is owed, and what I deliberately did NOT do

1. **`article-body` gains an image capability** — optional field + template markup, default
   absent so every existing instance renders byte-identically. **Blast radius is 297
   instances / 30 sites and `content_components` is live config, applied the moment it is
   written**, so this goes through the council gate with a byte-identity test, not straight
   into the DB. Not written yet.
2. **`news-listing`** — same class, and it carries the real news. Same treatment.
3. **Identify the unplanned imagery producer** (above) before assuming the plan is the
   request path.
4. **The logo/background rule into the imagery prompt** — durable half of ruling 2.
5. **NOT dispatched anything at boxingonline** — `site_delivery_and_editor` owns that
   pipeline and has work in flight. **NOT touched P1 code** — separate track, and the
   extraction (`2a0bdb001`) is one commit old.

## 2026-09-02 (later) — migration 686 WRITTEN, rehearsed and submitted: `article-body` gains an image

State re-checked before starting, because this thread had been open a while: 84 commits
had landed since the morning's record. `article-body` **unchanged** (`updated_at`
2026-08-24 16:57, md5 `002cbcd9cada6a37bf4a5158fd1e5f22`, len 1378, no image markup), the
defect still live at the artefact (both sampled articles and `/news/index.html` still 1
img = logo), and **nobody else on it** — no commit, no migration, no open work item
mentions `article-body`.

### The fix is smaller than the finding, because the producer half already exists

The morning's NOTES left "identify the unplanned imagery producer" as an open question.
**Answered by reading, and it changes the fix**: `plan_sections_action.go:463-476` already
looks up `imageryplan.ContentHeroKey(pageName)` in `assets` and binds it to
`r.assets["hero"]`. Its own comment says this is *"what makes the article page show the
same image family as its listing card"* and calls it a *"convention with no plan row"* —
which is exactly why my morning census found 12 images and 0 `site_plan_imagery` rows and
could not explain it. **So no producer work is owed and no plan row is needed.**

Three more facts that made this a component-only change, each checked rather than assumed:

- `site_assets.hero` as a field source is **already precedented** — the live `hero`
  component's `background_image` uses that exact source.
- optional-image-plus-alt is precedented by `content-block-about` (`image_src` =
  `site_assets.image`, `image_alt` = `llm`, both `required:false`, `on_missing:skip_field`).
  I copied that shape rather than inventing one.
- **`rerender_page_sections_action.go:464` builds the resolver WITH the page name**, so
  existing pages acquire the field on their next rerender. No backfill, no re-plan. (The
  render-time gap-fill in `render_site_components_action.go:1007` passes `""` for the page,
  so it could NOT have served this — worth knowing before anyone reaches for it.)

### Written: `686_article_body_hero_image_capability.sql` (+ `_ROLLBACK`)

Two optional fields (`hero_image_url` ← `site_assets.hero`; `hero_image_alt` ← llm) and one
`{{if .hero_image_url}}`-guarded figure, plus 176 bytes of CSS. Guarded on the template's
exact md5, idempotent, anchor-uniqueness asserted (`replace()` replaces EVERY occurrence),
post-conditions in `DO`/`RAISE` because a verify block of `SELECT`s cannot stop the `COMMIT`.

**Equivalence stated precisely instead of as "byte-identical":** markup is byte-for-byte
today's for any instance with no hero; the only difference in emitted bytes is the 176
characters of CSS, whose selectors match no element such a page contains. Saying
"byte-identical" would have been false, and the harness is what forced the honest wording.

### Proof, with controls — `harness/articlehero/` (durable, not scratch)

Renders the OLD live template and the NEW one against the **real stored `content_data` of a
live article**. **14/14 PASS.** Two controls prove the checks discriminate: (13) the
unguarded `alt` spelling **does** emit the literal `<no value>`, and (14) an unguarded block
**does** change the no-hero markup. Without those, checks 11/12 and 1/2/4 would pass while
proving nothing.

**Then rehearsed the migration itself** under `BEGIN`/`ROLLBACK` against the live database —
both `DO` blocks verified — and a full **apply-then-reverse round-trip in one transaction**,
proving `_ROLLBACK` returns the template to its exact pre-686 md5 with both fields gone and
`content` intact. Live template confirmed untouched afterwards.

### Three missteps, all caught by the checks rather than by review

1. **My byte figure was wrong.** Check 3 asserted 166 bytes of CSS; the real figure is 176.
   It FAILED on the first run and I corrected both the test and the migration header. A
   figure I had typed twice, in two files, confidently.
2. **The first cut of the migration was not valid PL/pgSQL** — `IF <expr> FROM
   content_components WHERE ...`. `IF` takes an expression, not a `FROM` clause. **The
   BEGIN/ROLLBACK rehearsal is the only thing that would have caught this**; writing SQL and
   submitting it unrun is how three P1 council rounds were spent on symbols that did not
   compile, and the same trap applies to SQL. The counts now come from the initial
   `SELECT INTO`. ⚠ **Rehearse every migration before submitting it, not after approval.**
3. **I truncated my own submission file by redirecting into the file I was reading**
   (`python3 ... < f > f`): the shell opens the target for writing *before* the reader runs,
   so it read an empty file and the content was gone. Rebuilt to a new path. **Never
   redirect into a path that appears in the same command's input.**

### Submitted to the council gate

`SUBMISSION_CORR = 4bf6c48f-9cd6-440f-9257-a5668b6635fc`. Admission tested free with
`DRY_RUN=1` first (which is also how I found the submission envelope is **nested** —
`plan.summary` / `plan.edits` / `plan.grounded_in` / `plan.risks`, not the flat shape I
wrote first). The `risks` block separates what I have already measured (so no seat re-derives
it) from the three things a reviewer should genuinely judge — the `site_assets.hero` role
choice and its site-brand-hero fallback, figure-at-head versus a separate hero row, and
alt-text provenance. Budget ~30 minutes; a missing orchestration row is latency, not a
dropped dispatch.

**NOT applied.** `content_components` is live config the moment it is written, so 686 waits
on the verdict and the owner. Nothing dispatched at boxingonline — the delivery lane owns
that pipeline. `news-listing` has the same defect and is deliberately NOT in this change;
it follows once this shape is agreed.

### 2026-09-02 — the council run was KILLED BY A ROLL, and the "existing pages pick it up" claim is now properly grounded

**The run died; it did not run slowly.** `4bf6c48f` sat at `review_prior_art`,
`EXECUTING_STEP`, `error` NULL, for 47 minutes. Both `agent-chassis` pods were **created
at 12:28:03Z and 12:28:24Z** and the run's last activity is **12:28:32Z**. Two other runs
died in the same window (`2979c27f` = the `inline_guide_imagery` lane, submitted three
seconds after mine; `84b51f16`, unnamed).

**The control is what makes it conclusive rather than suggestive:** every run submitted
AFTER the roll completed normally — `bd469ba1` 3m48s, `1dd3d298` 5m29s, `38be9226` 5m48s,
`3f9cdfea` 19m55s, all COMPLETED. The gate is healthy; only the runs in flight across
12:28 died. Without that control the honest reading would have been "the gate is slow
today", which is the reading that produces a duplicate retry.

Already a landmine (`LANDMINES.md:1563`, "A chassis roll KILLS an in-flight council") so
nothing new to file — **checked before appending, not after.** Resubmitted with
`RESUBMIT_CORR` so the trail stays under `4bf6c48f`; new run correlation `e0bec270`,
orchestration `c27ec7bd`. The `inline_guide_imagery` lane was told, with the evidence, since
a dead row of theirs looks exactly like latency and they had no reason to suspect it.

**Separately — the one claim in the submission that rested on a single line is now
grounded properly.** I had written "rerender_page_sections_action.go:464 builds the
resolver WITH the page name, so existing pages acquire the field on their next rerender".
True, but that is an inference from a constructor argument, and inferring a capability from
a nearby line is the exact class I logged a wrong call for this morning. The file's own
header settles it outright:

> "RerenderPageSectionsAction re-renders ALL of a page's sections from their STORED
> content_data plus **FRESHLY re-resolved dynamic fields**, WITHOUT invoking the content
> writer (no LLM). It is the lightweight path for 'a resolved field changed, re-render the
> page' — **specifically an image asset landing (hero/section image)**"

and the merge comment at :1325 adds that the path "merges stored content_data with FRESHLY
resolved fields, **resolved last**" — so a newly resolvable field wins over stored absence
rather than being carried past. **This action exists for precisely the case 686 creates.**
So the operational promise is: apply 686, fire a rerender at the six pages, and the images
appear — no re-plan, no writer call, no backfill.

### 2026-09-02 — 686 vs the per-section imagery binding: STRUCTURALLY immune, not merely currently safe

The `inline_guide_imagery` lane raised a real hazard against 686 (their IMG-075 work): adding
an image field to `article-body` makes it a third component able to reach the **per-section**
imagery resolver, and on a page carrying TWO instances of one component the binding can
attach a real figure to the WRONG section — which renders, deploys and looks correct. Their
guard for it is written but **not in the running binary** (their round-2 commit missed the
12:28 roll by ~2.5 minutes).

They then measured the duplicate count themselves and reported zero, i.e. moot. **Reproduced
here — both their spelling and a tombstone-filtered version return 0 pages with two
`article-body` instances.** But "zero pages today" is a DATA fact, and their own guides design
(one illustrated block per `h3`) is built to falsify it. So I checked the structure instead,
and it is a better answer:

```
-- plan_sections_action.go, the query that populates r.sectionAssets:
 WHERE sp.site_id = $1 AND spi.scope = 'section'
   AND spi.scope_ref LIKE $2 || ':%'
   AND spi.kind IN ('illustration', 'icon', 'infographic')
```
and `sectionAssetFor` is only a lookup into that map (`byKind[path]`). **686's field is
sourced `site_assets.hero`, and `hero` cannot enter that map — excluded by a kind filter in
CODE, not by the absence of data.** Hero resolves in the separate page-scope branch
(`ContentHeroKey`), site brand hero as last resort. Census agrees and shows the filter is
load-bearing rather than decorative `[MEASURED 2026-09-02]`: page/hero 369, section/icon 200,
section/illustration 9, section/infographic 1, site/hero 3, site/logo 46 — **no section/hero
row exists and no code path would read one.**

**Kept, because it bites LATER:** their `slot_name` carry-forward trap. `save_page_sections`
writes `slot_name = component name`, and `plan_sections`' carry-forward drops any slot whose
rows repeat with differing `content_data`. So on a page with repeated same-component sections,
a resolver-sourced field that resolves to nothing is **not** rescued by the carry — it is
dropped, and `on_missing: skip_field` makes that silent. That is how apis.uk's six
illustrations became one content-write away from vanishing. **Inapplicable to `article-body`
while the duplicate count is zero; live the day the guides work makes it non-zero** — which is
that lane's own stated direction, so this is a dependency between the two lanes, not a
curiosity.

**A figure of theirs did not reproduce, and it is the same shape as a correction they had just
sent me.** They twice quoted "93 instances across 18 sites" for `article-body`. Measured with
**their own query's spelling**: **297 rows, 1 distinct component, 30 sites** — and by
`component_id` against the row 686 edits, the same 297/30. Told them; not guessing at their
mechanism, but 3× matters when the reasoning rests on it.

⚠ **And a census-drift datum worth having: it moved 297 → 298 (30 → 31 sites) between two
queries about a minute apart, and `site_plan_imagery` hero went 359 → 369 since this morning.**
The estate plans heroes continuously. The submission's "297 as of 2026-09-02" is correct-as-of
and dated, which is the whole reason the dating rule exists — but nobody should quote it bare
next week.

### 2026-09-02 — council round 1: REVISE, and both HIGH objections were worth the round

8 approve / 3 object, gated by `editquality`. Decision `revise`, correlation `4bf6c48f`.
**Neither HIGH was answerable by argument — both named a landmine I had not consulted, and
both were right that I had asserted by analogy.** Answered by measurement and resubmitted
(round 2 envelope `f54dde26`, orchestration `a9a0e83d`).

**editquality (HIGH, gating) — the landmine is REAL: `LANDMINES.md:6018`,** *"A component
field whose `source` is not \"llm\" makes the content writer skip the LLM ENTIRELY and
re-render from template — the run reports success and the copy never changes."* If that fired,
686 would stop `article-body`'s TEXT being written on 297 instances, silently — catastrophically
worse than the bug. **It cannot fire, for the mechanism the entry itself names:**
`llmFieldSpecs` is appended only `if source == "llm"`; the tag is
`json:"llm_field_specs,omitempty"`, so an EMPTY list serialises as **absent**; and
`check_render_mode` branches on `current_section.llm_field_specs != null`. **The trap fires
when a component has NO llm field at all.** `article-body` keeps `content` (source llm), so
the list holds one entry and is never empty. The partition is per-FIELD, not per-component.

**prior_art_librarian (HIGH) — `LANDMINES.md:18413`,** *"A COMPONENT FIELD SOURCED
`site_assets.image` RENDERS THE PAGE'S OWN HERO, NOT A CONTENT IMAGE."* Their fear: all six
articles show ONE shared site hero, replacing the measured defect with a different one.
**Proven not to, two independent ways.** (a) The landmine's mechanism is ALIAS resolution —
`imageRoleAliases` maps `image` → `hero`, taken because *"the literal key ALWAYS misses,
because nothing anywhere populates `r.assets["image"]`"*. 686 sources the **literal** key
`hero`, which IS populated, so the alias branch is never reached. (b) `r.assets["hero"]` is
filled by three arms in order, and I measured which fires `[MEASURED 2026-09-02]`:

| arm | source | fires for the six blog pages? |
|---|---|---|
| 1 | planner page hero | **NO** — page-scope plan hero rows exist for exactly four page names (about, contact, index, tool-fight-calendar), none of them blog pages |
| 2 | `ContentHeroKey(pageName)` | **YES, 6 of 6** — each has an ACTIVE asset at exactly `'content_hero_'||replace(name,'-','_')`, which is `ContentHeroKey` verbatim (`imageryplan.go:233`) |
| 3 | site brand hero | **CANNOT** — this site has ZERO site-scope hero rows |

So each article resolves its OWN image, and the one arm that could serve a shared image is
absent. Arm order fences it even if arm 2 missed.

**A third landmine, raised by NO seat, which is this change's honest residual:**
`LANDMINES.md:9271` — a true content **regeneration** (`tone_shift`/`content_rewrite`), unlike
the merging re-render, DROPS non-llm keys when the resolver resolves nothing at that moment,
and *"a gated field fails more quietly than an ungated one"* — exactly this field's shape
(optional, `skip_field`, behind `{{if}}`). Not a reason not to ship: it is the standing
behaviour of every non-llm field in the estate, and the key is rewritten while the asset is
active. But **if the content hero is ever deactivated, the next regeneration drops the field
silently — no error, no failed work item.** Now in the migration header. Same family as the
`inline_guide_imagery` lane's `slot_name` carry-forward trap.

**One of my round-1 claims was simply WRONG and is corrected:** I wrote that the
`content-block-about` precedent has "both fields `required:false`, `on_missing: skip_field`".
Measured: `image_src` has `on_missing=skip_field`; **`image_alt`'s `on_missing` is UNSET.**
Behaviourally identical (`LANDMINES.md:9274` — it defaults to `skip_field` when omitted), but
the claim as written was false. 686 sets it explicitly on both, i.e. stricter than the
precedent. `news-listing` "same shape" is now verified too, not analogised: `items`
(query.news_archive), `headline`, `subheadline`, `loading_text`, all llm — **no image field at
all.**

**Accepted without argument:** `bug_historian`'s medium — this is instance N+1 of the generic
`missingkey=zero` class (016b §9 item 7), correctly guarded HERE and closing nothing
fleet-wide. **Kept as a naming decision rather than conceded:** `reuse_agent`'s low objection
that `hero_image_url` diverges from the precedent's `image_src`. It does, deliberately — that
vocabulary is the subject of landmine 18413 precisely because it names a content slot and
delivers chrome. Reusing it would import the confusion the landmine exists to prevent.

**⚠ FOR WHOEVER APPLIES THIS (`render_guardian`, non-blocking but operationally load-bearing):
existing instances need a path that RE-RESOLVES against `content_data` — a scoped rerender or
a resave. A plain assemble-mode page-rerender redeploys the stored HTML unchanged**, so the
field would never appear despite the assets existing, and the rollout would read as a no-op.
Carried into the migration header, because the lane that applies this is not the lane that
wrote it.

**MISSTEP 4 (2026-09-02) — I executed my own commit message.** `f3f81ba39`'s message used
backticks around three identifiers inside a double-quoted `git commit -m "..."`, so bash ran
them as command substitution: `bash: content: command not found`, and the same for `image` and
`hero`. The commit succeeded with those three words **deleted** from the message. They are the
three the argument turns on — the reader is left with *"article-body keeps ."*, *"the literal
key  is never populated"* and *"sources the literal , which IS populated"*, which is worse than
useless because it reads as fluent prose with the discriminating term silently removed.
Corrected in the follow-up commit; forward-only, so no amend.

**This is a documented trap** (`MEMORY.md` → shell-tool-traps: *"backticks in `-m` execute"*)
and I walked into it anyway, because the message was long prose and the backticks were doing
markdown work, not shell work. **The durable fix is mechanical, not vigilance: write commit
messages with `git commit -F <file>`** (or a quoted heredoc into a file), so the shell never
parses the message at all. Adopted here from now on. The tell is free and I ignored it —
`command not found` on STDERR, immediately above a successful commit line.

### 2026-09-02 — council round 2: APPROVED, and the two advisories closed with measurements

`4bf6c48f` → **approved**, "approved with 2 advisory objection(s) — none high-severity". Both
round-1 HIGHs are gone: `editquality` now approves outright ("both round-2 HIGH objections are
answered with mechanism-level arguments, not analogy"), and `prior_art_librarian` moved from
HIGH to approve. **The migration is written, reviewed and NOT applied** — `content_components`
is live config, so it waits on the owner.

Two advisories were worth closing rather than filing:

**1. `render_guardian` MEDIUM — I elided the CSS in the sketch, so they could not check it.**
Their objection is procedural and correct: the submission wrote `<176 chars of
.article-body__hero CSS>` as a placeholder, so nobody could verify it against the
`--section-*`/`--color-*` `var()` fallback-chain contract. The literal, for the record:

```css
.article-body-section .article-body__hero{margin:0 0 2rem}
.article-body-section .article-body__hero img{width:100%;height:auto;display:block;border-radius:var(--radius-md,8px)}
```
**Zero colour declarations** — no `background`, no `color`, so it cannot bypass the var chain;
the one variable used carries a fallback. The concern does not bite, but **they could not know
that, and that is my fault, not theirs. LESSON: never elide the actual bytes in a submission
sketch — an elision converts a checkable claim into a trust request.**

**2. `prior_art_librarian` MEDIUM — my llm-dispatch defence was "asserted from source reading,
not verified by execution". Fair, and now MEASURED in live data** `[MEASURED 2026-09-02]`.
`content-block-about` carries **11 llm fields and 2 non-llm fields** — the exact mixed shape
686 creates. Across its **15** live instances:

| llm `heading` written | llm `body_text` | llm `eyebrow_text` | non-llm `image_src` resolved |
|---|---|---|---|
| 15/15 | 15/15 | 15/15 | 14/15 |

So a component holding non-llm fields demonstrably **still gets its LLM copy written**. The
landmine's skip-the-LLM branch is not reached where an llm field remains — now shown by data,
not only by reading `omitempty`.

**MISSTEP 5 — my first attempt at that very probe returned a false zero.** I queried
`content_data->>'eyebrow'`; the key is **`eyebrow_text`**. I had taken the field name from
`brief-explanation`'s schema and applied it to `content-block-about`. The result — "0 of 15
have LLM copy" — was alarming and completely wrong, and it would have read as *evidence for the
very landmine I was refuting*. **A probe for a key you have not confirmed exists returns
absence, and absence is indistinguishable from the failure you are looking for.** The fix was
one query: list `jsonb_object_keys` first, then probe. Same family as every other misstep today
— asking a question whose encoding I had not checked.

Also carried, from `debug_historian`'s advisories: the 016 back-catalogue has titled cases for
both "deployed hero images exist but the page renders the fallback" and "two rebuild routes —
only page-build-handler re-resolves section sources". The second reinforces the rollout note
already in the header; the first is worth a pointer for whoever revisits this class.

### 2026-09-02 — 686 APPLIED and RECORDED (owner go-ahead), and bug 426 changed how I applied it

**Applied 2026-09-02 ~14:55 BST**, on the owner's explicit instruction. Pre-check first: template
still at md5 `002cbcd9…`, len 1378, `already_applied=false` — so the guard's precondition held and
no other session had touched the row in the two hours since approval.

```
NOTICE:  686: article-body updated (id 5835b2e1-50d7-4f20-8a9c-8da4d270ae3d)
NOTICE:  686 VERIFY: OK — guarded block present, content intact, both new fields optional
COMMIT
```

**Verified independently of the migration's own VERIFY block**, at the row:

| new md5 | len | guarded block | `<no value>` literal | url source | alt source | content source |
|---|---|---|---|---|---|---|
| `80d531ce726798c2d5ce2b5f3542f3e2` | **1730** | present | absent | `site_assets.hero` | `llm` | `llm` (intact) |

**1730 is an independent cross-check**: the local harness computed exactly 1730 for the new
template hours earlier, from a completely different code path (Go `text/template` over the dumped
old template). Two routes to the same byte count.

**Control on the blast radius:** `page_components` untouched — 301 live `article-body` instances,
**0** carrying `hero_image_url`, most recent row write `13:46:47Z`, i.e. before my apply. Existing
pages therefore render exactly as before, which is what the {{if}} guard promised. ⚠ Note the
instance count is now **301**, not the 297 the submission quoted this morning — that is the census
drift already logged (297 → 298 → 301 in one day). The submission's figure was correct-as-of and
dated; **do not quote it bare.**

**BUG 426 CHANGED THE METHOD, and it is the reason this is not a defect.** That bug — filed today
by the `bugfix_314_council_scope` lane — is exactly *"a migration applied by hand and never
recorded in the ledger looks, to the runner, like one that has never been applied; the next
`--apply` replays it"*. Seven live instances accumulated in one lane over a month. **I hand-applied
686** (the runner has no single-file apply, and `--apply` takes EVERY pending file — the documented
trap), which is precisely the shape that produces instance #8. So the second step was not optional:

```
./scripts/migration/run-migrations.sh --record-only 686_article_body_hero_image_capability.sql \
  --note 'applied by hand 2026-09-02 … md5 80d531ce…, len 1730 … content field still source llm …'
```
→ `== 686_article_body_hero_image_capability.sql recorded (applied_by='record-only')`

**Reading a bug before acting is what made this two commands instead of a month-long invisible
drift.** The user asked for 426's state in the same breath as the apply, which is the only reason I
read it first.

**Also ran the mandated dry-run** (CLAUDE.md: "dry-run per SESSION and after every roll") — which
is 426's own §8 first action, the fleet-wide `Pending (N)` census it explicitly leaves
`[UNMEASURED]`. ⚠ Two process notes for anyone repeating it: it is SLOW (probes each pending file
in a doomed transaction, minutes not seconds), and **piping it through `tail` withholds ALL output
until it finishes**, which reads exactly like a hang. Another session was running the same census
concurrently.

**STILL NOT DONE, and it is not mine:** the six boxingonline articles do not show their heroes yet.
The capability now exists; the pages need a **re-resolving** path (scoped rerender or resave), not
a plain assemble-mode page-rerender, which would redeploy the stored HTML unchanged and make this
look like it failed. That is `site_delivery_and_editor`'s call. `news-listing` still has the
identical defect and still has no change written.

### 2026-09-02 — 686 ROLLED BACK 69 minutes after it was applied: the remedy was wrong for 97% of the population

**Nothing was ever rendered with it — 0 of 301 instances acquired the field before the rollback,
and the template is back at its exact pre-686 md5 `002cbcd9…`, len 1378.** That is a short window
plus luck, not a control I had built.

**What happened.** The `inline_guide_imagery` lane sent an unrelated finding: two live pages where
an aliased image field renders the page's own hero — *"the same file twice on one page,
immediately under the hero that already shows it"* (`vonc.com/about`). I recognised the shape as
something 686 could cause and went to count `[MEASURED 2026-09-02]`:

> **292 of the 301 pages carrying `article-body`, across 31 sites, ALSO carry a `hero` component
> whose `background_image` reads the SAME `site_assets.hero` key.**

So on **97%** of the population, 686's field would have rendered the same image twice — hero at
the top of the page, the identical file again at the top of the article body. **The six
boxingonline pages that motivated the whole change are in the nine-page minority with no hero
component at all.**

**The reframe, and it is the useful part.** Peer pages are not broken and never needed
`article-body` to carry an image. They show imagery through the **`hero` component from a
page-scope plan row** — verified at the artefact: `agritec.uk/blog/insect-bioconversion.html`
renders `background-image: …url('/assets/images/hero-bsf.jpg')`, has **1** page-scope plan hero
row, and its only `<img>` is the logo. **The fleet pattern is hero-section-for-imagery,
article-body-for-prose, and 292 pages follow it.**

**So the real defect is one level up and is a page-composition question:** the six boxingonline
blog pages have **no hero component and no page-scope plan hero** — which is exactly the case the
`ContentHeroKey` convention generates per-article images FOR (its own doc comment: *"a per-article
image GENERATED from the article's own content … for a page the planner gave no hero of its
own"*). The images are generated for pages that have nowhere to put them. **The question for
whoever takes this on is why the planner composed those blog pages without a hero when 292 peers
have one — not how to teach `article-body` to hold an image.**

**What this cost and what it did not.** Two council rounds, an apply, a rollback, and a wrong
paragraph propagated into `035` §8 (corrected in place). It cost no rendered page, no customer
artefact and no other lane's work. **The council could not have caught it** — eleven seats
reviewed the change as I described it, and my description was accurate about the change and
silent about the population.

**The check I did not run, which is one query:** before adding a field to a shared component, ask
**what the other instances already do**. I measured the component, the assets, all three resolver
arms, the llm-dispatch mechanism, the alias map and the precedents — **every layer except the
neighbours.** Logged in `WRONG_CALLS.md` as
*generalised-a-remedy-from-the-motivating-case-to-the-population*.

**Ledger state, deliberate:** `schema_migrations` still carries the 686 row **on purpose**, so
`--apply` cannot replay the defective file; its `notes` now records the rollback, the reason and
the live md5. The file itself carries a DO-NOT-APPLY header. Any superseding fix takes a **new
migration number** (forward-only).

**Also measured while here, for `bugs_open/426` §8** — which explicitly leaves it `[UNMEASURED]`:
the mandated dry run reports **`Pending (164)`** fleet-wide. ⚠ Two process notes: it takes **over
five minutes** (my first run was killed by `timeout 300`), and piping it through `tail` withholds
all output until it finishes, which reads exactly like a hang. That runtime is a plausible part of
why a free, mandated check goes unrun — worth handing to that lane.

### 2026-09-02 — outcome: the 189-census became a fleet detector, and the ROLLBACK shaped it more than the census did

Closing the loop in this lane's own record, because the result otherwise lives only in another
lane's files. The `bugfix_114_imagery_wiring` lane turned this lane's one-shot census into a
standing state: **`check_unrendered_page_imagery` (IMG-077, commit `a87746b77`, migration 708,
inert until the next chassis roll)** files one flag-only rollup per site for exactly the
population I measured (state `no_image_slot`), counted per site **with the census date in the
spec** and retracted automatically when a site's population empties. Their CONTRIB is in this
directory.

**Two design choices they made because of migration 686's failure, not because of its success —
which is the part worth remembering:**

1. **`no_image_slot` is FLAG-ONLY, with no prescribed remedy.** Their stated reason is my
   97%-double-image finding: 686 was a prescribed remedy that was right for nine pages and wrong
   for 292, and a detector that had named it **would have propagated the error faster than I
   could retract it.** A detector carrying a wrong remedy is worse than no detector, because it
   arrives with authority.
2. **The generator is deliberately NOT gated.** They measured something I had not: a content hero
   on a slotless page **still feeds listing-card derivation** (193/193 event-convergence
   fleet-wide since 08-26). So those images are not waste, which retires the "why generate them at
   all?" question I would probably have asked next — and gating would have industrialised the
   opposite error.

**One correction I sent and they took:** the composition reframe is a **QUESTION, not a finding.**
I established that 292 of 301 article-body pages carry a hero component, that their lane measured
330 of 432 guide/blog pages the same way, and that `ContentHeroKey`'s doc comment says it exists
for *"a page the planner gave no hero of its own"*. **I never read the planner, so I did not
establish WHY.** Their first cut paraphrased it toward asserting a cause; the remedy text now
states it as open and points at `bugs_open/412`. ⚠ **Watch for this shape when a finding of yours
is adopted elsewhere** — a paraphrase in someone else's document is where a measured absence
quietly becomes a claimed cause, and the lane that measured it is the only one who can see the
difference.

**Also settled, and it corrects forty of my own probes:** their `boxingonline.com` HTTP 000 was
not an outage. That domain has **no DNS record** (bare and `www`); the site serves at its publish
target `boxingonline.ugg2.com` (200), control `agritec.uk` 200, and `sites` says
`publish_project = boxingonline.ugg2.com`, cutover pending approval. **I probed that site ~40
times today, every one against the publish target, and never once stated why** — so a peer's
entirely correct observation looked like a contradiction. **State the target you are probing and
why, the first time you probe it.** Their 000 failed SAFE; the dangerous direction is the parked
customer domain that 200s every path including invented ones.

### 2026-09-02 — ANSWERED: "what would ever write an infographic row?" — the planner is OBEYING AN INSTRUCTION, and I verified it at the live prompt

This lane's 2026-08-31 CONTRIB asked the question and could not answer it: the infographic capability
is built, routed, admitted and rendered end-to-end, and **1 row has ever been planned fleet-wide**.
The `inline_guide_imagery` lane found the answer in the live `build-site-planner` prompt and routed
it here. **I read the prompt myself rather than relay it** — all four claims verified verbatim
against `agent_definitions` id `f263eaa1-…`, 2026-09-02:

1. **The vocabulary is COMPLETE, not missing.** The `kind` field is documented
   *"one of: `logo`, `hero`, `illustration`, `icon`, `infographic`. No other values permitted"*, and
   rule 15 repeats the enum. So nothing needed teaching — the word was always there.
2. **Section-scope imagery is told to be rare, in terms:**
   > *"**`sections`** — for icons, illustrations, or infographics attached to a specific section.
   > **Use sparingly in v1 — most plans will have zero section-scope entries.** Only emit a section
   > entry when a specific section's imagery need is not covered by the page hero."*
3. **The stated minimum is CHROME ONLY** — verbatim: *"one site-scope `logo` entry, one page-scope
   `hero` entry under `pages.index`, and one page-scope `hero` entry for every other page whose
   `sections` array contains a hero-class component"*. **There is no floor for illustration or
   infographic anywhere.**
4. **`infographic` occurs exactly THREE times in the whole 34,781-byte config, and all three are
   rule/schema text** — the field-table enum, the `sections` description, and rule 15's repeat.
   **It appears nowhere in the worked example**, while the other four kinds do.

**So the census is the prompt working exactly as written.** hero 399 / icon 211 / logo 50 /
illustration 25 / **infographic 1** — *"most plans will have zero"* produced almost exactly zero.
The 08-31 finding (*"the capability is built, routed and rendered end-to-end and nothing has ever
asked"*) was right in every particular; the cause is **one sentence of English that nobody had read
against the numbers.** And infographic's 1-of-1 has its own explanation: it is the only kind that is
**permitted, never required, and never exemplified.**

⚠ **A silent mechanism is usually UNDRIVEN, not missing — and this is the sharpest instance this
lane has seen.** Four layers of working machinery, a complete vocabulary, and a rate of ~0, all
produced by an instruction telling the model to expect zero.

**NOT TOUCHED, deliberately.** The prompt is read by the build path for every new site, the cost is
real generated images per section, and 18 remakes are queued behind it — it is the planner owners'
call, not this lane's. The other lane gave the designblog lane the verbatim quotes so whoever edits
it argues with the text rather than a summary. **If it is edited, one condition rides with it**
(their insistence, and it is right): rule 16's *"each entry produces exactly ONE image"* discipline
must be in the same edit, because under-decomposition is what produces unusable multi-panel images.

**⚠ THE DISTINCTION THAT MUST SURVIVE TO THE OWNER — two asks, not one.** A prompt change lands
pictures **where there is structure to hold them**, and on article pages there is none:
`[MEASURED 2026-09-02]` max prose sections on any of **462** article pages is **1**; of the 9 pages
carrying an illustration-capable section, **8 are landing pages with exactly one and ZERO are
`blog-post` or `guide`**. So it would improve the landing pages the owner is looking at, and would
not put a single image inside article text. **Progress on the first must not be reported as progress
on the second** — that is this lane's one-slab finding (`bugs_open/114`) arriving from the supply
side.
