# SPEC 2026-09-04 — constraints any new CAROUSEL component must satisfy

**Written by `editorial_design_uplift` for the `finetuning_uk_service` lane to paste into a
`needs_new_component` brief as its non-negotiable half.** The *taste* half — what the carousels
should be, how many, which suits which grid — is deliberately not here; the design-critique report is
the input for that.

**Every constraint below was verified at the code on 2026-09-04, not quoted from this lane's PLAN.**
File and line are given so the creator can check rather than trust.

---

## 1. NO ARITHMETIC. A template that computes a slide position renders NOTHING.

The render funcmap is **seven functions and no more** (`component_library.go:914–920`):

```
default · eq · ne · lower · upper · isset · safe
```

There is no `add`, `sub`, `mul`, `div`, `index`, `len`, `printf`, or range arithmetic. **A missing
function is a Go template PARSE error, so the component does not degrade — it renders empty**, and an
empty section is what the estate's shrink floors and "assembled to nothing" guards exist to catch
after the fact.

**The rule:** every number a carousel needs — slide width, track offset, transition step, card count,
active index — is either **emitted by the generator into a CSS custom property** and divided by the
browser, or handled in the component's own JavaScript. **Never computed in the template.** Both live
chart components already work this way; copy their shape.

## 2. IT MUST SURVIVE BEING ON THE PAGE TWICE — and the check that catches this is NOT armed here

`InstanceCollisions` (`component_instance_scope.go:323–348`) names four failure classes. **A carousel
can hit all four**, and it is the most JavaScript-heavy shape in the library, so this is the section
to read twice:

| class | what it means for a carousel |
|---|---|
| `DuplicateElementIDs` | two carousels with the same ids — **every `getElementById` resolves to the first**, so carousel 2's buttons drive carousel 1 |
| `EmptyElementIDs` | **one is already a defect** — an empty `id` is addressable by nothing, and two are a collision the duplicate check cannot see |
| `WindowOnloadAssignments` | **more than one means all but the LAST component never initialises.** Two carousels on a page, each assigning `window.onload`, and one of them is simply dead |
| `UnscopedInlineScripts` | any inline script not wrapped in an IIFE puts its declarations in global scope, where a second instance — or a different component — replaces them by name |

**The rule:** ids derive from the per-instance token, never a literal; no `window.onload` (use a
scoped `DOMContentLoaded` listener or an IIFE that runs immediately); all script bodies wrapped in an
IIFE.

⚠ **Do not rely on the platform to stop you.** `enforceInstanceScope`
(`component_instance_scope.go:315`) returns the config key's value and defaults to **false**. Verified
live 2026-09-04: it is armed on `tool-deployer` and `tool-generator` and on **nothing on the section
render path** — so on the path a card carousel actually takes, a collision is **recorded and
shipped**, not refused.

## 3. THE DECORATIVE / ASSERTIVE BOUNDARY — the trap in the owner's own wording

The ask is *"decorative (relevant) graphics"*. That is safe right up to the moment a graphic carries a
word or a number, because **text inside `<svg>` is invisible to the claims gate**: `svg` is a member
of `nonAssertionElements` (`datahelpers/claims.go:499–503`, alongside `script`, `style`, `code`,
`pre`, `iframe`). The gate never sees it, so nothing checks it against the site's evidence.

**The rule, and it is one sentence:**

> **Decoration carries no words and no numbers. Anything that carries either is an ASSERTION, and it
> is rendered as HTML text resolving through a registered fact — never as `<svg>` text.**

A card graphic that says "40% faster" inside an SVG is not decoration; it is an unchecked claim that
has routed around the one mechanism built to check it. Shapes, rules, arrows, textures, silhouettes
and abstract marks are decoration and are unconstrained by this section.

**This is also the hand-over seam to the infographics lane's route B.** The moment a card graphic
needs to *say* something — a mechanism flow, a comparison, a checklist, a chart — it stops being a
carousel-decoration problem and becomes one of their components (`mechanism-flow` 14 live placements,
`evidence-chart` 10, `checklist` 9, `comparison-table` 8; five of six template-rendered). **Do not
reinvent those inside a carousel card.**

## 4. COLOURING RESOLVES THROUGH THE PALETTE, AND FURNITURE HAS A CONTRAST FLOOR

*"Relevant colouring"* must come from the site's palette tokens (`var(--color-…)`), never a hardcoded
hex. Two independent lanes run colour-fixer sweeps over hardcoded declarations; a component that
ships literals will be rewritten by them within the week, and the rewrite is not guaranteed to
preserve the design.

⚠ **Carousel furniture — arrows, dots, progress bars, card borders, focus rings — is a GRAPHICAL
OBJECT under WCAG non-text contrast**, which is a real floor and not a nicety. The intuitive token,
`--color-border`, is usually the FAILING one. Check contrast in the same pass that chooses the colour,
not afterwards.

⚠ **And a CSS `var(--x, #fallback)` literal is present in the source and never applied** if `--x`
resolves — so a fallback is not evidence of the colour a viewer sees. Ask `getComputedStyle`, and
measure the contrast of what is actually painted.

## 5. WHAT THIS SPEC DOES NOT COVER

- **Which carousels to create, and what they should look like.** Taste; the design-critique report is
  the input, and this lane declined it deliberately.
- **Where they get applied.** A slot swap changes one page; changing a component changes every page
  using it — `features` is on **41** `page_components` across **12** sites, and on finetuning.uk alone
  it is on **five** pages `[MEASURED 2026-09-04]`.
- **Whether a component may become a default.** `hero-card-carousel` and `image-hover-card-grid` each
  have **zero live renders** `[MEASURED 2026-09-04]`; a default is the wrong first outing for either.

## 6. THE ACCEPTANCE TEST, so "it works" is not a matter of opinion

1. Render the component **twice on one page** and run `DetectInstanceCollisions` over the assembled
   HTML. `Clean()` must be true — not "no visible problem", the function.
2. Render it with its image field **absent**. It must degrade to a card with no empty figure and no
   reserved blank space (the `bugs_open/111` gate-the-container-on-its-contents rule).
3. Grep the rendered output for `<svg>` containing any `[0-9]` or a word of the site's copy. Any hit
   is a §3 violation.
4. Confirm no hardcoded hex outside a `var()` fallback, and measure contrast on the arrows and dots
   with `getComputedStyle` rather than reading the stylesheet.
5. Confirm the template contains no arithmetic — the parse succeeding is the test, since a missing
   function fails at parse time.

— `editorial_design_uplift`, 2026-09-04
