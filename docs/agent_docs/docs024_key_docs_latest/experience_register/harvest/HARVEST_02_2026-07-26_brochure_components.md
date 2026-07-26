# HARVEST 02 — the brochure component library (2026-07-26)

Owner ruling, same day: harvest the brochure set before building P2, on the reasoning that
harvesting four entries changed the schema in five structural ways and a *different kind* of
component may change it again — cheaper to find out before the migration than after.

It did. Five entries, one structural change, and one live defect found by applying a harvested
clause. Entries: `harvest/entries/CC-003…CC-007`.

## 1. What was harvested, and the live proof for each

All five components are live on `fundamentallyai.com`, one per page. Every page fetched this
session; both behaviours confirmed in the **served** bundle by strings they create.

| entry | component | live page (HTTP 200) | behaviour |
|---|---|---|---|
| `CC-003 arrow-and-swipe-card-carousel` | `hero-card-carousel` | `/capabilities.html` — 4 cards, `data-hcc-autoplay="false"`, live region present, **no pause control shipped** | JS in bundle (`data-hcc-track`, the "Card N of M" string) |
| `CC-004 hover-reveal-card-grid` | `image-hover-card-grid` | `/model-fine-tuning.html` — 4 `:focus-visible` rules in the shipped CSS | CSS only |
| `CC-005 scroll-snap-card-track` | `swipeable-insight-carousel` | `/multi-agent-review-council.html` — snap track, links on 2 cards only | CSS only |
| `CC-006 count-up-stat-band` | `stat-band` | `/index.html` — 4 count-up figures, each with its authored value on `aria-label` | JS in bundle (`__statBandInit` ×2) |
| `CC-007 illustrated-statement-block` | `people-feature-block` | `/about.html` — exactly one link | none, by design |

The brochure workstream's own `components/README.md` still lists four of these as "not yet
placed" (a 07-22 status). They are all placed and live now. Not a criticism — a status line in
a README decays exactly the way a handoff's diagnosis does; it is why every claim above names
the URL I fetched.

## 2. The structural finding: one invariant, six independent sightings

The same rule keeps arriving from different authors, in different mechanisms, on different
sites:

| sighting | mechanism |
|---|---|
| archive row with no case (`CC-001`, vonc) | JS strips the template's `href="#"` and the tabindex |
| feed CTA with no destination (`CC-002`, vonc) | the control is not rendered |
| carousel with fewer than two cards (`CC-003`) | JS hides the arrows and the pause control |
| pause control when auto-advance is off (`CC-003`) | the element is never emitted — confirmed on the live page |
| card with no link URL (`CC-005`) | template conditional `{{if .link_url}}` |
| statement block with no link (`CC-007`) | same conditional |

**One rule: a control that cannot do anything must not be presented as a control.** Six
implementations, five of them written without reference to each other.

So the register needs **named invariants** — clauses held once and *referenced* by entries
(`requires_invariant: ["no-inert-control"]`), not restated per entry. Restating is how six
subtly different versions of one rule came to exist in the first place, and a register that
copies the clause into every entry would industrialise the drift rather than end it. A second
invariant is already visible: **`pointer-behaviour-has-a-keyboard-equal`** (`CC-004`'s paired
`:hover`/`:focus-visible` rules; `CC-003`'s arrow keys).

This is a **new field and a new table row type** for P2 — see §5.

## 3. Further corrections to the design (continuing HARVEST_01 §3's numbering in spirit)

### 3.1 Not every trigger has a human actor — `automatic_triggers`
The contract shape assumes a visitor operates a control. Four of the behaviours here are
triggered by something else entirely: **the viewport** (count-up on scroll-into-view; the
carousel suspending auto-advance when scrolled away), **a system preference**
(`prefers-reduced-motion` suppressing motion in three of the five), **the visitor's own
scrolling** (the carousel re-deriving which card is current), and **focus movement**
(auto-advance yielding while anything inside is focused). These are contract — they are what
makes the component polite — and there was nowhere to put them.

### 3.2 I recorded stat-band as having no interaction. It has the most interesting one.
`design/taxonomy_seed.md` listed `stat-band` as the example of "a component with no
interaction — a contract recording 'none' is still worth an entry". It ships a count-up
animation with an honesty rule I would not have thought to write: *the animation never invents
digits, it counts only to a value already rendered from authored data, and the true figure sits
on the accessible name throughout so assistive technology never hears the fake intermediate
numbers.* Corrected in the taxonomy. The "inert by design" slot passes to `CC-007`, which
genuinely has nothing.

### 3.3 Accessibility clauses are contract, not decoration
The carousel announces "Card N of M" into a live region on every step; its pause control's
accessible name flips with its state; arrow keys do what the arrows do. The hover grid's reveal
is bound to focus as well as hover. The snap track hides its "Swipe or scroll →" hint from
assistive technology because it describes a gesture a screen-reader user will not make. **Every
one of these is the first thing lost when behaviour is re-invented**, and each is a checkable
outcome rather than a style preference. They belong in the contract, as outcomes with a named
channel — not in a separate "a11y notes" field nobody reads.

### 3.4 "Inert by design" needs to be assertable, and is not
`CC-007` exists to say *nothing here is a control*. No check we have can express "zero
additional links or buttons inside this container" — `selector_count` has no zero/equals form.
So the one component whose contract is emptiness cannot be verified, and a dead-control sweep
cannot tell "correct and inert" from "unfinished". Add to the HARVEST_01 §5.4 dependency list:
**zero-count and scoped-negation assertions**.

### 3.5 Behaviour survives without JavaScript — and which patterns do is worth knowing
Two of the five have no JavaScript at all; the carousel's own swipe/snap works with the bundle
off (JS only adds arrows, keyboard stepping and auto-advance). Given `bugs_open/084` (published
pages can silently lose their JavaScript), "what still works when the bundle fails" is not
academic. It fits `degraded_states`, which HARVEST_01 already added — this is the second,
independent reason for that field.

### 3.6 Idempotence, because behaviour ships in a shared bundle
`stat-band` guards with `window.__statBandInit`; the carousel instead scopes to its own root
and re-queries. Two mechanisms for one requirement — a site-wide bundle can be evaluated more
than once. Worth a clause so it is a decision rather than an accident.

## 4. Applying a harvested clause found a live defect — the register's case, demonstrated

`CC-003`'s contract says a card's outcome is *a real page load of the card's destination*. So I
checked the destinations on the live page. Measured this session:

```
https://fundamentallyai.com/capabilities        -> 404   (all four carousel cards link here)
https://fundamentallyai.com/capabilities.html   -> 200
fragments #review-council #verification #rapid-delivery #embeddings on capabilities.html -> none exist
image-hover-card-grid cards -> /model-fine-tuning.html#evaluation, #review-council  (page 200, fragments absent)
```

Every card in the harvested carousel is dead twice over: an extension-less path that 404s, and
a fragment that does not exist. The hover grid's cards reach a real page and then a fragment
that isn't there.

**This is not a new bug and I have not filed one.** It is exactly the class of
`bugs_open/071_…validate_gate_detects_every_broken_link_then_discards_the_finding` — which
already carries the extension-less class (8 targets, handed over from 049 on 2026-07-26) and
names the fragment blind spot (24 of 25 anchored links fleet-wide). Evidence contributed to
that file; the fix belongs to whoever owns it. Two things this session adds: these are
**component-card** links rather than nav/chrome, and they sit on a site whose link audit
reported 43 targets and zero broken on 2026-07-25 — an audit that resolves paths and drops
fragments will call a page sound while every card on it is dead.

And the point for this workstream: **a registered contract with a declared destination role
turns "is this link right?" into a check that runs, instead of a judgement someone has to
remember to make.** Today nothing expresses what those four cards were *supposed* to reach.

## 5. What P2 must build differently (adds to HARVEST_01 §5)

6. **Invariants are a first-class thing**: a small `experience_invariants` table (name, clause,
   rationale, sightings) plus `requires_invariant jsonb` on `experience_patterns`. Seed with
   `no-inert-control` (6 sightings) and `pointer-behaviour-has-a-keyboard-equal` (2). The
   pattern council checks entries against invariants rather than re-arguing them.
7. **`automatic_triggers jsonb`** on `experience_patterns` (§3.1) — viewport, system
   preference, focus movement, page lifecycle. Distinct from `contract`, whose actor is the
   visitor.
8. **Accessibility outcomes stay inside `contract`**, with a `trigger_channels` /
   announcement field, rather than a separate section (§3.3).
9. Criteria dependency list gains **zero-count and scoped-negation** (§3.4), alongside
   HARVEST_01's four.

## 6. Coordination

The brochure component library belongs to another session
(`brochure_component_library/`, owner-green-lit chart component in flight). Everything here is
read-only: their repo sources, their DB rows read, their live pages fetched. Nothing was
edited, no dispatch fired. The one thing handed over is the dead-destination evidence in §4,
appended to `bugs_open/071` where the class already lives.
