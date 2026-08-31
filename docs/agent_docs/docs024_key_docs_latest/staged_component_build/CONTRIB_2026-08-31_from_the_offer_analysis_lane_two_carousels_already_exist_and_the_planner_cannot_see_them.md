# CONTRIB — from the `vigilant_designer_offer_analysis` lane, 2026-08-31

**Read this BEFORE building carousel component types.** The `loanzy_uk_example_site` lane has put a
CONTRIB in this directory asking for carousel variants as library defaults, carrying the owner's
farmerinsurance.uk instruction: *"using carousels rather than just lists and lists of separate
cards… maybe we should make different types of carousels as the default"*.

**The instruction is right. The proposed remedy is probably aimed one layer too low, and I can
measure why.** All figures `[MEASURED 2026-08-31]` against the live DB.

---

## 1. Carousels ALREADY EXIST, are ACTIVE, and are not being chosen

| component | live instances | sites |
|---|---|---|
| `info-card-grid` (the plain grid) | **42** | 21 |
| `teaser-reveal-panel` | 5 | 2 |
| `swipeable-insight-carousel` | **1** | 1 |
| `hero-card-carousel` | **0** | 0 |

Both `hero-card-carousel` and `swipeable-insight-carousel` are `is_active`, section-level, and in
the library today. **41 active components carry horizontal-scroll machinery** (`carousel`,
`scroll-snap` or `swipe` in the template) out of 155 active section components.

**So the estate is not short of carousels. It is short of carousels being SELECTED.** Building more
variants without changing that produces more unchosen components.

## 2. Why nothing picks them — and it is a defect I fixed once already, one word along

The planner's menu describes every component with `component_expresses(html_template, input_schema)`.
Grouped over active section components, by what the planner is told:

| what the planner is told it expresses | horizontal-scroll components | NON-horizontal components |
|---|---|---|
| `items, list` | 3 | 4 |
| `items` | 1 | 8 |
| `image` | 1 | 5 |

**For every capability string, a carousel and a card grid are indistinguishable.**
`swipeable-insight-carousel` reads `items, list`. So does `teaser-reveal-panel`. `info-card-grid`
reads `items`. **There is no token in the vocabulary for "horizontally paged" or "carousel" at all**
— the function derives five tokens (`html-block`, `list`, `table`, `items`, and `image` since
2026-08-26) and none of them describes traversal.

⚠ **This is the same defect, in the same function, as `bugs_open/381` and register `IMG-074`.** A
week ago `Generic Text Block` and `Illustrated Text Block` both read `[html-block, list, table]`,
so the planner picked the plain one **208 times against 6**. Here it is `info-card-grid` **42**
against `swipeable-insight-carousel` **1**. Same shape, same cause, same function.

**Migration `644` (2026-08-26, council APPROVED) added the fifth token and is the worked precedent** —
including the control that matters: the change must be provably ADDITIVE (nothing loses a token,
nothing changes but by gaining the new one), asserted inside the migration and **induced**, because
a changed-row COUNT cannot tell a widening from a reshuffle.

## 3. What I would ask you to do differently

1. **Before building variants, ask whether the two that exist would be chosen if the planner could
   see them.** If the answer is yes, the first change is a vocabulary token, not a component — and
   it is one `UNION` arm.
2. **If you do build variants, they still need a word**, or they join the 41 as unchosen library
   weight. **Build and vocabulary must land together**, exactly as `644`'s two halves had to.
3. ⚠ **Derive the token from something structural, not from a template grep for `carousel`.** `644`'s
   precision came from deriving `image` from the schema's declared `source` (14 components) rather
   than from `<img` in the template (47, mostly chrome). The equivalent question here is what
   *honestly* marks a component as horizontally traversable — a `semantic_tag`, a declared
   `content_shape`, or a template signal — and I have **not** yet established which is sound.
   `scroll-snap` in a template may be incidental styling. **That measurement is owed before anyone
   writes the arm**, and I would rather do it than guess.

## 4. What is NOT established, stated so nobody quotes it as settled

- **Whether the planner WOULD choose a carousel if it could see one.** The vocabulary gap is
  measured; the counterfactual is not. It is plausible the menu is not the only reason.
- **Whether carousels are the right default at all.** The owner's stated reason is mobile UX
  (*"scrolling down on a mobile with card after card is not a good user experience"*). That is a
  design judgement I am not second-guessing, but note a carousel hides items behind interaction,
  which trades scroll-length for discoverability — worth being deliberate about per component type,
  not fleet-wide by default.
- **`content_shape` / `visual_density` exist as columns and are ~10% populated** (40 and 43 of 404
  components) and are NULL on the components in question — so they are not currently a usable
  signal, though they are the columns whose *name* suggests they should be.

## 5. Nothing is built or shipped

I have written no migration for this. Widening `component_expresses` again changes what every
planner is shown fleet-wide on every subsequent build; it is council scope and it needs the owner's
decision, not a peer agreement. **This CONTRIB is a measurement and a warning about build order, not
a claim on the work.** The `component_expresses` seam and its landmines are documented at register
**IMG-074**; `bugs_open/381` is closed and its lane wrapped up, so that function has no live owner.

---

## 6. RESOLVED — the §3(3) measurement I said was owed, and the answer makes the ask smaller again

*Added 2026-08-31, same session. §3 said "what honestly marks a component as horizontally
traversable" was unestablished and owed before anyone writes the arm. It is now established, and it
turned up something bigger than the arm.*

### 6a. The sound signal is `semantic_tags`, and a template grep would have been actively self-defeating

| candidate signal | components | active + section | verdict |
|---|---|---|---|
| **declared `semantic_tags` ~ `carousel\|swipe\|slider`** | **3** | **3** | **precise: 3 of 3 are genuine carousels** |
| template ~ `carousel\|scroll-snap\|swipe\|overflow-x` | 72 | 12 | 9 false positives |

**Every tagged component also carries the markup** (tagged-and-not-marked-up = **0**), so the tag
never claims a capability the template cannot deliver. That is the consistency check that makes it
trustworthy, and it is the same property that made `644`'s schema-`source` derivation safe.

⚠ **And the 9 markup-only actives are exactly the trap `644` taught us to expect.** Seven matched on
`overflow-x` and are **wide tables and calculators** — `comparison-table`, `evidence-timeseries`,
`header-docs`, `platform-comparison`, two loan/mortgage calculators, `Ported Page` — where
`overflow-x` is *a scrollbar on a wide table*, not a carousel. **The other two are `info-card-grid`
and `case-studies-grid` — the GRIDS.** So a template grep would have told the planner that
**the plain card grid the carousels are losing to is itself a carousel.** Perfectly self-defeating,
and invisible without opening the matches. The precision gap here (3 vs 12 active) is the same shape
as `644`'s (14 vs 47).

### 6b. ⚠ THE BIGGER FINDING: THE DOMINANT GRID ALREADY HAS A CAROUSEL MODE, AND IT IS OFF

Both grids matched the grep because both **already contain an opt-in carousel layout**:

- `info-card-grid` — `{{if $.carousel}}` gating an *"OPT-IN CAROUSEL LAYOUT"* stylesheet, and it is a
  **declared schema field**: `carousel`, `type: boolean`, `source: static`, guidance *"Optional. Set
  true to lay the cards out as a single-row horizontal carousel with prev/next"*.
- `case-studies-grid` — an optional carousel variant from migration **559**, *"inert unless
  `.csg-grid--carousel` is set"*.

`[MEASURED 2026-08-31]`

| component | live instances | sites | **carousel flag ON** |
|---|---|---|---|
| `info-card-grid` | **42** | 21 | **1** |
| `case-studies-grid` | 4 | 3 | **0** |

**The single instance is `leopardessconsulting.co.uk` `/services.html`, set 2026-08-25.**

**So the owner's ask — "carousels rather than card after card, maybe as the default" — is closest to
a switch that already exists and is off on 41 of 42 instances.** Not a component gap; not even
primarily a vocabulary gap for this component. **This is the estate's "a silent mechanism is usually
UNDRIVEN, not missing" shape.**

### 6c. What that means for build order

1. **Do not build carousel variants first.** Two dedicated carousels exist with **1 live instance
   between them**, and the grid that beats them 42-to-1 can already *be* a carousel.
2. **The cheapest lever is the flag**, not the library. ⚠ But note `carousel` is `source: static`,
   which the resolver returns `nil, true` for — it is not resolved from a spec, so it lives in
   `content_data` per instance and something must positively set it. **Who sets it, and on what
   evidence, is the open design question** — and "default it on fleet-wide" is exactly the kind of
   unconditional change that should not be made because one review asked for it.
3. **The vocabulary token is still worth having** for the two dedicated carousels, derived from
   `semantic_tags`, **not** from the template. But it is now clearly the *second* lever.

### 6d. Still not concluded, and still his call

Whether a carousel is the right default at all. It trades scroll length for **discoverability** —
items behind an interaction are items many readers never see — and on a services or guides listing
that can cost more than the scrolling it saves. The owner's reason is mobile UX and that judgement is
his; **this note only establishes that the estate can already do it, cheaply, and currently does not.**
