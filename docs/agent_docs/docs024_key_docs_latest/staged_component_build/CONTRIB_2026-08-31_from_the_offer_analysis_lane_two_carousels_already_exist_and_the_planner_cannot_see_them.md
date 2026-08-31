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
