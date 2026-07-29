# PLAN — §4c panels-as-carousels with a deliberate cliffhanger (design + first slice)

**Status 2026-07-29:** design settled, first slice BUILT and placed on
fundamentallyai.com's home page. Owner decision this session: *"design + build a
first slice"*. This document is the design; the slice is the proof, and what the
slice cost is recorded honestly at the end.

Originating direction (owner, 2026-07-28, recorded verbatim in
`HANDOFF_2026-07-28_continue_here.md` §4c):

> *"as a whole almost all the panels could be carousels of one sort or another,
> especially on the home page, and have a really short first sentence and a small
> potentially incomplete second sentence to be completed when they click through"*

---

## 1. The decision that shaped everything else: this shape already had a name

`experience_patterns` already holds **`teaser-detail-deeplink`** (kind
`micro-journey`, status `draft`), built before this workstream existed:

> *"A list of teasers where activating one opens its full text in place — no page
> load — while the address bar gains a parameter that reproduces the same open
> state on a cold load. Rows with no full text are not offered as controls at all."*

That is the owner's idea, already written down. So the build is **a second
component implementing an existing shape**, not a new shape. This matters beyond
tidiness: §4e's whole argument is that the value of a shape vocabulary is that the
count stays small, and *"two visually different components can share a shape"* is
the property that makes it worth having. A panel and vonc's provocations archive
look nothing alike and are the same micro-journey.

The component declares it in the markup: `data-experience-pattern="teaser-detail-deeplink"`.

**Not done, deliberately:** the pattern's `section_types` array does not yet name
`teaser-reveal-panel`. Those rows belong to the experience-register thread and were
edited twice on 07-28. Proposed to them rather than taken — see §7.

## 2. The content contract

Three fields per item, and the split is the design:

| field | what it is | rule |
|---|---|---|
| `hook` | one very short complete sentence, under 12 words | must stand alone |
| `continuation` | the deliberately unfinished second sentence, under 20 words | must be genuinely completed by `body`; **never** an ellipsis |
| `body` | the full text revealed on activation | **optional** |

`body` being optional is the honest half of the contract. An item with no body
renders as a plain statement with **no control at all** — never a button that opens
nothing. That is the pattern's own `absence_semantics` (*"an entry without body …
MUST degrade to the non-openable state, never to a dead link"*), and it is also the
owner's **replace-before-deleting** rule in a different costume: do not put a
control on the page pointing at something that is not there.

It follows that an item with no body must not carry a cliffhanger either — you may
only tease what you can deliver. The template enforces this structurally: the
`data-continues` mark is emitted **only** on the branch that has a body.

## 3. The three hazards §4c named, and what was actually done about each

**Hazard 1 — the claims gate's ±70-character window.** `datahelpers/claims.go`
requires a fact's `context_terms` within ±70 characters of the number, so a split
that separates a figure from its supporting words makes a *true* figure look
unverified. Amelioration as designed in §4e ("refuse to split a sentence containing
a figure at all") is now a hard rule in `llm_guidance`: *"Never write a figure,
percentage, count or date in hook or continuation. Any number and the words that
give it meaning must sit together inside body."* The live slice obeys it — every
figure ("more than a dozen", "100% to around 97%") sits wholly inside a `body`.

**Hazard 2 — a cliffhanger looks like truncation.** `output_tokens == max_tokens`
detection, and every checker built on it, reads an unfinished sentence as damage.
The amelioration is the one §4e specified: **mark it in the data, not the prose.**
The continuation is wrapped in `<span class="trp__continuation" data-continues="true">`.
A checker can therefore distinguish intent from damage by reading an attribute, and
the guidance forbids the ellipsis, which is precisely what a truncated completion
also produces. The render harness asserts both (`no ellipsis anywhere`, and the
mark appearing exactly on the openable items).

**Hazard 3 — an LLM on the render path.** Not built, and the design says do not
build it. `rerender_page_sections` is this workstream's only LLM-free repair route;
it is how the palette, the inks, the chart and this week's copy corrections were all
fixed without regenerating a word. The splitter, if it is ever built, **splits once
at plan/build time and persists into `content_data`**. The slice needed no splitter
at all: its content is existing approved copy, rearranged by hand.

## 4. A fourth hazard, found while building, that §4c did not anticipate

**A JS-driven reveal hides site copy from the only checkers that read it.** vonc's
implementation of this shape populates its detail region from JavaScript, and its
contract requires the region be *emptied* on close — because `innerText` on a
`display:none` element falls back to `textContent`, so an acceptance check can pass
without the interaction ever happening.

Copy that only exists after a JS call is invisible to the claims gate and to
crawlers. This is the same class as the finding already recorded for this lane —
**text inside `<svg>` is invisible to the claims gate** (`claims.go:137`
`nonAssertionElements`) — and the body text here is exactly the assertive prose that
most needs checking.

So this component uses **native `<details>`/`<summary>`**: the body is always in the
DOM, the `open` attribute is the honest state signal, and the reveal works with
JavaScript disabled. The JS snippet adds *only* URL addressability. With it blocked,
missing or broken, every teaser still opens and nothing becomes a dead control.

That is a real departure from the pattern's stated contract (which says the region
must be emptied), and it is a departure the pattern should probably absorb: its
"emptied" clause is a property of *one implementation*, not of the shape.

## 5. What was built

```
components/teaser-reveal-panel/
  template.html      Go template + inline CSS. Native <details>; scroll-snap track
                     (swipe on mobile, grid on wide screens); degraded branch.
  input_schema.json  the contract above, with the figure rule in llm_guidance
  behaviour.js       progressive enhancement ONLY: ?open=<key>, popstate,
                     open-on-cold-load, siblings close
  sample_data.json   three items: one full, one exercising the label fallback,
                     one with NO body to exercise the degraded branch
  register.sql       generated install (template + schema inlined)
scripts/render_teaser_reveal_panel.go   14 assertions, run before any DB write
```

Every CSS variable was checked against live `css_themes` **before** use, because
`--color-surface`, `--spacing-section` and `--container-max-width` are defined by no
active theme and the fallback silently wins (L-landmine, 07-26). All twelve used
here are defined.

## 6. How it was proven, including the parts that failed first

- **Template renders, 14 checks pass**, run against `html/template` exactly as the
  platform render path does, before anything touched the database.
- **The checks are non-vacuous, proven by two mutants.** Giving the bodyless item a
  body fails 6 checks including `degraded branch fired at all`; changing a
  continuation to end in "..." fails `no ellipsis anywhere`. A green check nobody
  has seen go red is not evidence.
- **The first version of the harness was wrong, and the mutants caught it twice.**
  It counted `.trp__card` inside the `<style>` block and reported four failures
  against a correct template — *a check that cannot tell a CSS rule from an element
  is measuring the wrong thing.* Then mutant A panicked it on an unguarded
  `strings.Index` returning -1. Both fixed; the harness now slices the markup away
  from the style block and reports a missing degraded branch as a failure.
- **Live verification** is recorded in NOTES against the served page, not the DB row.

## 7. What is deliberately NOT done

1. **The splitter** (owner's *"another small cpu based llm that splits the text"*).
   Designed above, not built. It only earns its place once there is a second and
   third panel to feed, and it must never touch the render path.
2. **`experience_patterns.section_types` does not yet name this component.** Those
   rows belong to the experience-register thread. Proposed, not taken.
3. **The planner does not choose this component.** Registration makes it reachable
   (`load_component_library` returns all active section components); whether the
   planner *selects* it is `features_open/017`, still unobserved. The slice is placed
   explicitly.
4. **The rest of the panels.** One panel, one page. "Almost all the panels" is a
   decision to take after the owner has looked at one.

## 8. The reversal lever

The slice **replaced** the home page's `differentiators` grid with the same six
pieces of approved copy in the new treatment — which is exactly the operation the
owner described ("changing a panel from a grid to a carousel"), performed without
re-running the writer, so no new claim entered the site. The displaced row is
archived in `page_component_history` with source `operator_trp_slice_2026-07-29`.
To revert: re-insert the archived `content_data` against the `differentiators`
component at position 4, delete the panel row, re-render index.

> **Landmine paid for during placement:** `page_components.id` is **not stable
> across re-renders**. A placement keyed on an id read forty minutes earlier
> silently matched nothing — `INSERT 0`, `DELETE 0` — and left both the old grid and
> the new panel at position 4. Key placement edits on `(page_id, function)`, never
> on a `page_components.id` you read before an intervening re-render.
