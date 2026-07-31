# PROPOSAL — building a site part step by step, with a travelling doc and a gate per stage

**Written 2026-07-30 for the owner**, from the session `fundamentallyai.com 4`
(`brochure_component_library` lane), at his request: *"write a detailed provenance
and step by step history of how we got to this reasonably good stage with the
carousel — I think if we are to build more and more complicated components we need
to do it step by step and follow the doc traveller idea for each small part of a
site build … We could for instance have a set of build tools and acceptance checks
— perhaps a bit like the checkers that we have now but are responsible for checking
a particular stage in development — some may even be created dynamically at the
start or at different stages of the project."*

**Status: ADOPTED 2026-07-30 by the `brochure_component_library` lane** (owner
direction: *"This provenance and ladder project is now this lane's project"*). It is
no longer waiting for a separate thread — this lane owns it, and its standing five
live beside this file in `staged_component_build/`. Nothing is BUILT yet; the first
unknown is now resolved (see the box below). Part 1 is history and is evidenced;
Part 2 is a design argument and is marked where it is not.

> **RESOLVED 2026-07-30 — the question this document said to ask first.** The
> proposal closed by marking `[UNVERIFIED]` whether `doc_plans` fits a component
> without schema change, and named that as the next thread's first action. Asked
> and answered: **no, and the reason is a one-line constraint, not a design
> mismatch.** Both tables carry a CHECK on `subject_type` —
> `doc_plans_subject_type_check` allows `tool|pipeline|experience|action|experience-pattern`
> and `doc_notes_subject_type_check` the same plus `landmine`. **Neither allows
> `component`**, so a component PLAN cannot be inserted today. Extending it is the
> smallest possible platform change, it is purely additive, and **there is a
> four-times precedent** (migrations 163, 184, 218, 270) — 270 is the template.
>
> **The better half of the answer, which changes the design for the better:**
> `doc_notes` has a `site_id` column and `doc_plans` does not. That is exactly the
> split a component needs and it was already there. A component's *contract* is
> fleet-wide (one `content_components` row serves 11 sites for `info-card-grid`),
> so the PLAN being site-less is correct. But S4–S7 are **per-site, per-page**
> facts, and the verdicts land in NOTES, which can carry `site_id`. **The PLAN is
> the fleet-wide contract; the NOTES are the per-site verdicts.** The schema
> already encodes the distinction the ladder needs, which is mild evidence these
> tables were built for something of this shape.

**How to read it.** Part 1 is the provenance: what actually happened to the carousel,
in order, with commits. Part 2 is the proposal, and it is *derived* from Part 1 rather
than imported from theory — every stage in the proposed ladder exists because Part 1
shows a stage where something was either caught or missed.

> ## ⚠ SUPERSEDED IN SCOPE 2026-07-31 — read this before building anything from Part 2
>
> **The owner cut the ladder from eight gates to three** (PLAN **D8**), and this document
> was NOT rewritten, deliberately: what it argued and how that survived contact is the
> useful record. Treat Part 2's eight stages as **a checklist to read and ignore where it
> does not fit**, not a build spec.
>
> **Machinery is being built for three things only:** the claim written before the build;
> verification through the visitor's real gesture; and every check proven able to fail
> (including *a mutant counts only if the artefact provably changed*).
>
> **Why**, in one line each, all evidenced below or in the REPLY: the mutation harness cost
> the first forward-run lane ~40 minutes and found nothing in their product · S0 was "a
> five-minute grep that prevented nothing" · S7 cannot be completed while `bugs_open/157`
> is open · and decisively, **two of these eight gates were wrong on first contact and S4
> would have BLOCKED a correct build.** Eight gates is eight chances to be confidently
> wrong at someone else's expense.
>
> **The current plan of record is the PLAN's D8 and phasing**, and the first item is not in
> this document at all: the three-way naming-contract check (`CHECK_naming_contract.sh`),
> because until it passes every other gate's green is untrustworthy.
> Plain-prose account: `SUMMARY_2026-07-31_we_cut_the_ladder_down.md`.

---

# PART 1 — the carousel, step by step, as it actually happened

The component is `teaser-reveal-panel`, live on four pages of fundamentallyai.com.
It took five rounds over roughly 20 hours. The fifth round found a bug that had
been present since the first, and that is the single most important fact in this
document.

Source documents, all current:
`brochure_component_library/PLAN_2026-07-29_teaser_reveal_panel.md` (the design),
`NOTES_brochure_component_library.md` (the technical log — the five rounds are the
entries from `## 2026-07-29 (later)` at line 2713 to the end),
`SUMMARY_2026-07-30_the_panel_is_finished_and_two_new_fronts_open.md` (the read-out),
`components/teaser-reveal-panel/` (the version-controlled source of the DB rows).

## Round 0 — the ask (2026-07-28)

The owner's direction, recorded verbatim in `HANDOFF_2026-07-28_continue_here.md` §4c:

> *"as a whole almost all the panels could be carousels of one sort or another,
> especially on the home page, and have a really short first sentence and a small
> potentially incomplete second sentence to be completed when they click through"*

Two things about this that mattered later. It was **written down verbatim rather
than paraphrased**, so every subsequent round could be checked against the original
words rather than against someone's summary of them. And it named a *shape*, not a
component — which is what made Round 1's first decision possible.

## Round 1 — design and first slice (2026-07-29 14:38, `d42ebc44f`)

**The decision that shaped everything: the shape already had a name.**
`experience_patterns` already held `teaser-detail-deeplink` (kind `micro-journey`,
status `draft`), written before this workstream existed:

> *"A list of teasers where activating one opens its full text in place — no page
> load — while the address bar gains a parameter that reproduces the same open state
> on a cold load. Rows with no full text are not offered as controls at all."*

That is the owner's idea, already in the system. So the build became *a second
component implementing an existing shape*, not a new shape — declared in the markup
as `data-experience-pattern="teaser-detail-deeplink"`. The check that produced this
was simply *look for the shape before inventing one*, and it took minutes.

**The content contract** — three fields, and the split is the design:

| field | what it is | rule |
|---|---|---|
| `hook` | one very short complete sentence, under 12 words | must stand alone |
| `continuation` | the deliberately unfinished second sentence, under 20 words | must be genuinely completed by `body`; **never** an ellipsis |
| `body` | the full text revealed on activation | **optional** |

`body` being optional is the honest half: an item with no body renders as a plain
statement with **no control at all**, never a button that opens nothing. That is the
pattern's own `absence_semantics`, and it is the owner's replace-before-deleting rule
in a different costume. It follows that a bodyless item must not carry a cliffhanger
either — you may only tease what you can deliver — and the template enforces this
structurally: the `data-continues` mark is emitted only on the branch that has a body.

**Three hazards were named in advance, in writing, each answered concretely:**

1. *The claims gate's ±70-character window.* `datahelpers/claims.go` requires a
   fact's `context_terms` within ±70 characters of its number, so splitting a figure
   away from its supporting words makes a **true** figure look unverified. Answer: a
   hard rule in `llm_guidance` — no figure, percentage, count or date in `hook` or
   `continuation`; any number and the words that give it meaning stay together inside
   `body`.
2. *A cliffhanger looks like truncation.* `output_tokens == max_tokens` detection, and
   every checker built on it, reads an unfinished sentence as damage. Answer: mark it
   in the **data**, not the prose — `<span class="trp__continuation"
   data-continues="true">` — and forbid the ellipsis, which is exactly what a
   truncated completion also produces.
3. *An LLM on the render path.* Not built, and the design says do not build it.
   `rerender_page_sections` is this lane's only LLM-free repair route. The slice
   needed no splitter at all: its content was existing approved copy, rearranged by
   hand.

**A fourth hazard was found while building, that the brief had not anticipated.**
A JS-driven reveal hides site copy from the only checkers that read it: copy that
only exists after a JS call is invisible to the claims gate and to crawlers — the
same class as the already-recorded finding that *text inside `<svg>` is invisible to
the claims gate* (`claims.go:137` `nonAssertionElements`). And the body text here is
exactly the assertive prose that most needs checking. So the reveal uses **native
`<details>`/`<summary>`**: the body is always in the DOM, the `open` attribute is the
honest state signal, and the reveal works with JavaScript disabled. The JS adds
*only* URL addressability.

**What was built:**

```
components/teaser-reveal-panel/
  template.html      Go template + inline CSS; native <details>; scroll-snap track
  input_schema.json  the contract above, with the figure rule in llm_guidance
  behaviour.js       progressive enhancement ONLY: ?open=<key>, popstate, siblings close
  sample_data.json   three items: one full, one label-fallback, one with NO body
  register.sql       generated install (template + schema inlined)
scripts/render_teaser_reveal_panel.go   14 assertions, run BEFORE any DB write
```

Every CSS variable was checked against live `css_themes` **before** use, because
`--color-surface`, `--spacing-section` and `--container-max-width` are defined by no
active theme and the fallback silently wins. All twelve used here are defined.

**How it was proven, including the parts that failed first:**

- The template renders and **14 checks pass**, run against `html/template` exactly as
  the platform render path does, before anything touched the database.
- **The checks are non-vacuous, proven by two mutants.** Giving the bodyless item a
  body fails 6 checks including *degraded branch fired at all*; changing a continuation
  to end in `...` fails *no ellipsis anywhere*. A green check nobody has seen go red is
  not evidence.
- **The first version of the harness was wrong, and the mutants caught it twice.** It
  counted `.trp__card` inside the `<style>` block and reported four failures against a
  correct template — *a check that cannot tell a CSS rule from an element is measuring
  the wrong thing*, and it failed in the direction that would have made me "fix" a
  working template. Then mutant A **panicked** it on an unguarded `strings.Index`
  returning −1. Both fixed.

Registered in the concept register as **CLC-012**.

**Four missteps this round, each caught by a check rather than by luck** — these are
the reason Part 2 proposes stage gates at all:

- `page_components.id` **is not stable across re-renders.** A placement keyed on an id
  read ~40 minutes earlier silently matched nothing (`INSERT 0`, `DELETE 0`) and left
  both the old grid and the new panel at position 4. Key placement edits on
  `(page_id, function)`. The landmine was already recorded by another lane; I hit it anyway.
- **The big one: a page-level placement does not survive a re-render, and the work item
  said `complete`.** The first index re-render **dropped the panel entirely** and
  renumbered the remaining sections, because the rebuild resolves sections against
  `site_plan_sections` and the plan still said `differentiators` at ordering 3. The green
  status was accurate about the rerender and silent about the section it discarded.
  **Placement is a plan fact; a `page_components` row alone is a render artefact.**
- A **UTC/BST clock trap** nearly produced a false stall report: my poll printed local
  BST while the DB reports UTC, so I read "10 minutes queued" off a two-minute-old row.
- The harness misread CSS as markup, as above.

**Verified live** against the served page, not the DB row: 5 openable `<details>`,
1 static card with no control, 5 `data-continues` marks, pattern declared, 0 ellipses,
0 unrendered vars. Contrast measured *in the state that only exists after a click* —
and the first attempt at that proved nothing, because `render_audit.py` renders a
**local** copy so `?open=review-council` never reached `window.location`. Wrote
`probe_reveal_open_state.py`, which forces every `<details>` open in the DOM: 5
revealed bodies, all 13.19:1. **It prints the count it measured**, so a probe that
opened nothing is distinguishable from a clean result.

## Round 2 — the deduplication detour (2026-07-29 18:30, `2e97f8929`)

The owner said the home and capabilities pages "seem very similar". The observation
was right and undersold, and this round is in the carousel's history because it
changed what the panel was *for* on three of its four eventual pages.

**Measured before touching anything.** A per-section fact census across all ten
served pages, against `site_specs.evidence_base` (9 owner-seeded facts) as the ruler:
**18 sections across 5 pages each restated 3 or more of the same 9 facts.** Home had
six capability-listing sections out of eight, three of them consecutive, two sharing
the literal heading "What this platform demonstrably does" with a third instance on
`/capabilities.html`. The `info-card-grid` on home versus capabilities was only **18%
textually similar** (`difflib.SequenceMatcher`) while asserting the identical six
facts — independently generated near-duplicate content, not a copy-paste, which is
the worse kind because it reads as independent content to a crawler.

Fixed on all five repeating pages (owner chose "one list per page", "all five"): kept
the section most on-topic for each page, archived and removed the rest — 7
`page_components` rows archived under `operator_dedup_capability_lists_2026-07-29`,
then removed from `page_components`, `pages.sections` **and** `site_plan_sections` in
one transaction, per the `(page_id, function)` lockstep Round 1 learned the hard way.
Did not trust the `complete` status: refetched all five pages live afterwards.

**The mechanism was then diagnosed properly rather than guessed, and filed as
`bugs_open/151`.** Per-section copy is written in total isolation — `page-content-writer`'s
`process_sections_loop` calls `generate_content` once per section with no sibling
section visible — and every one of those isolated calls receives the identical
whole-site `writer_block`, built once per site by `composeWriterBlock`
(`refresh_evidence_base_action.go:582-637`) with no per-fact usage tracking.
`EvidenceFact` (`claims.go:74-96`) has no such field at all — not a dead one, absent.
`build-site-planner` is the one place with full cross-page visibility and its only
duplication guard is page-level topic dedup (`053_build_site_planner.sql:2461`);
nothing about facts. `SelectComponentByType` (`component_selector.go:150-193`) has no
notion of a component being roster-shaped versus single-topic, so nothing stops two
full-roster components landing on one page — which is what home did three times over.

Three fix candidates, ordered by what makes the bad state unrepresentable: (1) scope
facts to sections at plan time, since the planner already has the visibility to do it
in one pass; (2) tag component shape so the selector stops pairing two rosters —
weaker, does nothing about two *narrative* sections restating a fact; (3) a post-build
fact-repetition census as a permanent gate, the only candidate that also protects the
nine already-deployed sites. Transferable pattern added to 016b §9: *a shared fact
pool handed unchanged to N isolated writers restates itself everywhere it is plugged in.*

## Round 3 — images, and rollout to four pages (2026-07-29 20:16, `4b2e157df`)

Owner: *"implement the carousels on almost every component block, with images."*

**Full inventory before touching anything**, because the component templates already
vary in what they support: `info-card-grid` (11 sites) already carries an optional
decorative `icon_image`; `image-hover-card-grid` (this site only) has full images but
a hover-based reveal that does not work on touch; `swipeable-insight-carousel` (this
site only) is a genuine swipe carousel with no image support and no reveal. Neither
`services-grid` (live on 6 sites) nor `info-card-grid`'s shared template was touched
— **converting a page away from a shared component is safe; editing the shared
component is not**, and wasn't done.

**The image inventory is a lesson in not trusting the obvious reading.** `assets.url`
for 10 of 15 icon rows showed a literal broken placeholder
(`/assets/images/input-data.asset-key.jpg`), which looked exactly like the "52 broken
asset rows" finding from `bugs_open/114`. It wasn't: the actual stable serving path is
a different convention than the `assets.url` column tracks, and curling all 25 icon +
hero + tool-hero paths found **25 of 25 already live (200)**, most already referenced
somewhere on the site. `assets.url` being wrong is a real, separate finding; it means
nothing about the images themselves. **12 were verified visually** — downloaded and
viewed — against their `site_plan_imagery.prompt` before any `image_alt` was written,
because the schema rule written in Round 1 forbids inventing one.

Extended, not replaced: the template gained an edge-to-edge `.trp__media` rendered in
**either** branch, open or degraded-static, so an image never depends on the reveal
firing. Harness **14 → 18 checks**; two new mutants (alt echoing the hook; stripping
every image) each fail exactly the check they should and nothing else.

Four placements, four different histories, one reversible archive pattern. The
capabilities set fit so exactly (`icon_service_*` maps 1:1 onto the section's six
existing headings) that it looks like it was generated for that section and never
wired in. On the fine-tuning page, **the Round 2 bug turned up again while comparing
two sections rather than by looking for it**: `info-card-grid` and
`image-hover-card-grid` independently restated 4 of their 6 facts each — the exact
`bugs_open/151` pattern, missed by that bug's own census because the census was scoped
to the nine company-wide facts and these were fine-tuning-specific claims.

Every hook and continuation was checked programmatically against the figure rule (no
digit in either field) before writing to the DB, plus ellipses, missing alt text, and
alt echoing hook. All four real payloads rendered through the **actual template**, not
the harness's hardcoded sample, before the SQL ran. **One SQL mistake, caught by the
transaction rather than by luck:** the rollout script wrote `SELECT page_id FROM pages p`
(`page_id` is `page_components`' column); `ON_ERROR_STOP` halted mid-script and the
whole transaction rolled back on connection close — verified rolled back before fixing.

## Round 4 — owner feedback: padding, ellipsis, open-state merge, real arrows (2026-07-29 21:50, `763ceb81c`)

Four requests, **one of which directly collided with a rule the component was built
to enforce** — so it needed a substitute, not silent compliance and not a silent refusal.

- **Padding**: `--spacing-lg` → `--spacing-xl`.
- **"Put an ellipsis at the end of the cut-off text."** Did *not* put a real `…` into
  the stored text — that is precisely what the component's own `llm_guidance` forbids,
  because a truncation checker built on `output_tokens == max_tokens` reads a trailing
  ellipsis as a cut-off generation. Substitute: **CSS `content: "\2026"` on
  `.trp__continuation::after`**, hidden once `[open]`. It exists only in the rendered
  pixel, never in the HTML text node, so it is invisible to the claims gate and to any
  truncation heuristic in exactly the way the stored character would not have been.
- **"Make the text read as one section when opened."** Read literally rather than as
  "relabel the control": `.trp__card[open] .trp__control { display: none; }` — the
  control disappears entirely once there is nothing left to invite.
- **"Six cards on two lines — make it a real left/right-arrow carousel."** Root cause
  found rather than patched: a `@media (min-width: 60rem)` rule switched the track from
  a horizontal scroll-snap row to a **wrapping grid** at desktop widths. That rule is
  gone. Arrows added — and **reused, not reinvented: the exact
  `goTo`/`nearestIndex`/`scrollBy` pattern already live on `hero-card-carousel`**, same
  keyboard support, same `aria-live` "Card N of M" convention.

**The genuine open question was answered with a stated default rather than left
hanging.** The owner invited discussion on how a dropdown should behave inside a
horizontally-scrolling carousel. Answer: `.trp__body { max-height: 12rem; overflow-y:
auto; }`, unconditionally — because without a cap, opening a card grows that grid row
and drags the fixed-position overlaid arrows out of alignment every time. In practice
no body written so far approaches 12rem, so it is a safety bound that is essentially
never active. **Flagged to the owner as a choice, not a foreclosed decision**, with the
alternative named and the reversal cost stated (one CSS rule).

Harness **18 → 23 checks**. Three new mutants — stripping the arrows, re-adding a
literal `...`, removing the open-state hiding rule — each caught by exactly the check
it should fail and nothing else.

## Round 5 — the round that changes how we test (2026-07-30 09:35, `1ec9e8cf6`)

Owner: the arrows did nothing, opening a second card left the first open, and the text
should replace the whole closed block rather than appending under it.

**The cause was found before anything was fixed, by simulating real clicks in a
headless browser.** And here is the finding that justifies this entire proposal:

> Every verification this component had ever passed — the render harness, the
> `probe_reveal_open_state.py` contrast checks, even this session's own earlier
> "verified live" claims about the deep-link and sibling-close — exercised either the
> **static markup** or forced `.open = true` **directly on DOM nodes**. None of them
> ever actually clicked anything, so none of them could have caught a JS
> initialisation bug.

Once `.click()` was actually called: `track.scrollBy` was never invoked, and opening
card 1 left card 0's `.open` still `true`.

**Root cause:** `<script src="/assets/js/snippets.js">` sits in `<head>`, plain, no
`defer`. It executes synchronously at that point in parsing — **before the panel
markup later in `<body>` exists**. `behaviour.js`'s very first line,
`document.querySelectorAll('[data-component="teaser-reveal-panel"]')`, therefore found
zero panels and the whole file's `if (!panels.length) return;` exited immediately.
**This had been true since the very first version, Round 1 included** — the deep-link
and sibling-close logic had never once run client-side on the live site. Only the
native `<details>` element, which needs no JS, and the CSS ever did anything. That is
why four rounds of checks all passed: the parts that worked needed no JS, and nothing
ever tested the parts that did.

Fixed by gating the per-panel init on `document.readyState`, the same shape
`hero-card-carousel` already used. Re-tested with the same real-click harness:
`scrollByCalled: true`, `scrollLeftAfterClick: 272` (was `0`), and opening card 1
correctly sets card 0's `.open` to `false`. Both reported bugs, one cause.

**Checked whether this was a platform-wide gap before assuming it was.**
`site_components` (`slot_name='head'`) confirms the non-deferred `snippets.js`
placement is fleet-wide — **13 of 13 sites**. That sounds structural, but the
`js_snippets` table narrows it sharply: **6 of 7 active snippets already guard on
`document.readyState`/`DOMContentLoaded`** (`hero-card-carousel`, `lobby-grid-loader`,
`provocation-card-loader`, `provocations-archive-loader`, `stat-band`, and now
`teaser-reveal-panel`). This is an established, widely-followed convention, **not an
unknown platform gap — the failure was mine**, for not checking the convention before
writing the component's first version. The one other exception is
`news-date-formatter`, unguarded, not investigated; flagged rather than chased.

Text merge implemented literally as *replace the whole block*:
`.trp__card[open] .trp__text { display: none; }` removes hook + continuation + control
as one unit, and what replaces it lives permanently in `.trp__body` — still always in
the DOM for the claims gate and crawlers — as
`<strong class="trp__body-lead">{{hook}}</strong> {{continuation}} {{body}}`,
concatenated with a literal space **in the template**, so it reads as one paragraph
because it *is* one `<p>`. Verified against the real rendered text, not just presence.

Harness **23 → 24 checks**, and one dead check retired honestly (the old open-state
control rule no longer exists; it is covered by the parent hide). Two new mutants.

**One loose end disclosed rather than hidden:** the real-click test reported two
generic `"Script error."` entries, which is what a browser reports for an uncaught
exception in a cross-origin script when the test document's origin (`file://`) differs
from the script's — very likely an artefact of the test rig, not the live page. Not
chased; flagged.

## What actually produced the good result — and the one thing that didn't

Reading the five rounds back, the practices that earned their keep were:

1. **The shape existed before the build.** Looking for it cost minutes and prevented
   a tenth vocabulary entry.
2. **Hazards named in advance, in writing, each with a concrete answer.** Three were
   named before building; a fourth was found during, and written into the same place.
3. **A harness run before any DB write** — the template proven against `html/template`
   exactly as the render path does it.
4. **Mutants.** A green check nobody has seen go red is not evidence. The mutants
   caught two harness defects, not just template ones.
5. **Verification against the served artefact**, never the `complete` status — which
   was directly vindicated: a `complete` rerender had silently discarded the panel.
6. **A probe that prints what it measured**, so "opened nothing" is distinguishable
   from "clean".
7. **Reuse over reinvention** — the arrows, and then the readyState guard, both came
   from `hero-card-carousel`.
8. **Blast radius measured before an edit** — shared templates left alone; the
   fleet-wide check in Round 5 that stopped a personal mistake being filed as a
   platform defect.

And the thing that did not work, stated plainly: **the check ladder had a hole exactly
where no check simulated a real interaction, and four rounds of honest, non-vacuous,
served-artefact verification all passed straight through it.** The bug was in the first
commit and survived until the owner clicked the thing. No amount of *more* checks of
the kinds already being run would have found it, because they were all sound about
what they measured. What was missing was not rigour. It was a **stage**.

---

# PART 2 — the proposal

## The problem, stated precisely

There is no such thing as a build *stage* for a site part today. There is a component
that does or does not exist, and a page that is or is not deployed. Everything in
between — is the shape right, is the contract sound, does the template render, is it
registered, is it placed durably, does it serve, does it *operate*, does it still
operate after the next roll — is carried in a session's head and in prose notes.

That is why Round 5 happened. Not because the checks were weak, but because "does it
operate when a human clicks it" was never a named thing that had to pass.

**The bet of this proposal:** name the stages, give each one a gate, and make the
travelling doc the thing that carries the gate. Then a more complicated component is
not a bigger leap — it is more stages of the same size.

## What already exists — the reuse inventory

This matters more than the design, because the honest finding from
`webdesign_tools_repair/REPORT_2026-07-29_concepts_for_a_working_tools_chain.md` is
that the platform already holds nearly the whole chain for **tools**, and the distance
to a working loop was five small pieces of wiring, not a new system. The same is
likely true here.

**Live and directly reusable:**

- **Travelling docs (TL-017, DOC-003).** Per-subject `PLAN` + `NOTES` rows in
  `doc_plans`/`doc_notes`, where the PLAN carries a fenced ```criteria block that *is*
  the machine-readable definition of working, and the NOTES accumulate every repair.
  Written by the agents themselves. **This exists for tools and not for components** —
  the single biggest reuse opportunity in this proposal.
- **The verification ladder, Tiers 0–4 (TL-008).** Cheap-to-expensive, each tier
  catching a different class. Tier 2 is static under the **anchor rule** — *static
  checks confirm, never refute* — and Tier 4 drives real headless Chromium via
  `browser-runner-adapter`, desktop and mobile profiles, with interaction checks and
  overflow attribution.
- **The criteria vocabulary**, read from the two evaluators' own switch statements:
  `selector_exists`, `selector_count`, **`interaction` (fill/click/select + expect
  `text_matches`)**, `page_status_ok`, `asset_loads`, `attribute_absent`,
  `attribute_matches`, `no_console_errors`, `no_horizontal_overflow`.
- **`has_visible_area` (TL-034), added 2026-07-30 by the `webdesign_tools_repair`
  lane** — measures `getBoundingClientRect()` against a floor (default 24×24,
  overridable per check), and **fails** on both a collapsed box and a missing
  element. It exists because three tools served work areas measuring **1146×0**:
  present in the DOM, invisible, and `selector_exists` passed all three. It is
  **Tier-4-only by necessity** — it measures rendered layout, which no static read
  of HTML can compute, so a Tier-2 equivalent is not unbuilt but impossible. This is
  the single most valuable import for **S2 and S6**: it is the primitive that
  separates *exists* from *usable*, and the carousel's own Round 5 was a
  reachability defect of the same family. **Committed but NOT YET ROLLED — see the
  hazard below before authoring any gate against it.**
  > **CORRECTED 2026-07-30 (evening) — it is also BROKEN, and the first forward run of
  > this ladder is what found it.** `bugs_open/157`: `has_visible_area` reports **0 for
  > any axis whose rendered size is a whole number.** `chromiumPage.VisibleArea`
  > (`run_checks_action.go:718-719`) does `w, _ := m["w"].(float64)` — and
  > `playwright-go` returns an integral JS number as Go `int`, so the assertion fails,
  > **the `, _` swallows it**, and `w` keeps its zero value. A `24px × 24px` checkbox
  > measures `0×0`; a `0×94` reading on mobile (only the integral axis zero) is the
  > observation that identifies the bug.
  >
  > **Two lessons for this ladder, and the second is the more important:**
  > (1) The primitive I called the most valuable import fails **in the direction of
  > reporting a defect that is not there** — the worst direction, because it invites you
  > to deform a correct product to satisfy it. The lane that hit it wrote *"DO NOT make
  > the checkbox size fractional to turn the gate green"*, which is exactly right: the
  > `24px` is a deliberate WCAG 2.2 target size.
  > (2) **A gate can be wrong, and the ladder must say what to do then.** The rule is now
  > explicit: **when a gate fails, the first question is whether the gate is right** —
  > file against the gate, keep the subject as the reproducer, and never tune the subject
  > to a green. This is the mirror image of S2's mutation rule: mutants prove a gate can
  > go red; 157 proves a red gate can be wrong.
- **The ten-rule criteria validator** (P1–P10) from the experience register — exported,
  and currently applied only to `experience_patterns`.
- **`experienceVerdict`** — the platform's only anti-vacuous verdict function
  (*"≥1 PASS and 0 FAIL, else inconclusive"*).
- **Session-level tools this lane already wrote:**
  `scripts/render_component_template.go`, `render_teaser_reveal_panel.go` (the 24-check
  harness), `probe_reveal_open_state.py`, `rebundle_js_snippets_direct.sh`,
  `rerender_page_sections_direct.sh`, `gen_component_register_sql.py`.
- **`scripts/render_audit.py`** — renders in headless Chromium, composites effective
  backgrounds, reports sub-AA pairs and failed images (from `features_open/026`).

**The single most important line in that inventory:** `interaction` + `text_matches`,
evaluated by `browser-runner-adapter`, **already does what Round 5's hand-rolled
real-click test did** — and it was proven end to end on 2026-07-29, the same day, by
the `smart-contrast` pilot: 11 of 11 checks in real Chromium across desktop and
mobile, asserting the tool's actual arithmetic against known answers, dispatched
through the platform's own agent rather than a session harness. So the missing stage
is **not new construction**. It is pointing a proven mechanism at components instead
of only at tools.

**Known gaps and hazards this proposal must respect** (all from the tools-chain report,
verified there against the live system on 2026-07-29):

- **G1** — the fixer never reads the PLAN or NOTES. The failing criteria are *in* its
  inputs as `acceptance_test`; the prompt just never references them. So a repair agent
  can undo a deliberate decision every time it fires. Any staged scheme that writes
  down "deliberate decisions — do not re-fix" needs this fixed or the writing is inert.
- **G3** — criteria fences are never linted, and the composer has been recorded
  inventing selectors twice (TL-016). The remedy TL-016 already concluded was
  *validation, not sterner prompts*.
- **G4** — the Tier-4 judge can pass on nothing: an all-skipped result set yields
  `len(Failed)==0` → PASS plus a 7-day cooldown.
- **G5 / enablement** — discovery passes are currently **manual-fire**, and the
  improvement loop is stopped by owner ruling. A ladder nothing fires is a mechanism
  rotting unexercised.
- **`bugs_open/149`** — only 2 of 22 discovery handler agents run
  `validate_page_content`; six registered checks are configured in no agent and have
  filed zero items ever. **Proliferating checkers is a measured failure mode here**,
  not a hypothetical one.
- **`bugs_open/126`** — Tier 4 navigates once per profile and state accumulates across
  checks, so a consent-gated subject cannot pass.

**The hazard that threatens every gate in this ladder, found 2026-07-30 and filed to
`LANDMINES.md`.** An unknown check type is **skipped, not failed**: the Tier-4 type
switch ends `default: skip(ch.ID, ch.Type+" not implemented")`
(`internal/adapters/browserrunner/run_checks_action.go`). Combine that with **G4** —
an all-skipped result set yields `len(Failed)==0` → **PASS note plus a 7-day
cooldown** — and a gate written against a check type the running binary does not
carry **passes vacuously and then suppresses its own re-check for a week.**

This is not hypothetical and it is live right now. `has_visible_area` was committed
at 07-30 15:19 (`1850acb07`); the running `browser-runner-adapter` pod is older and
does not contain it (two long markers unique to the change grep 0 against three long
pre-existing controls at 1 each, pod `browser-runner-adapter-8646cddb79-qfcmr`). So
the newest and most useful check type for this ladder is, at the time of writing,
**exactly the one that would silently skip.**

Three consequences for the design, and they are requirements rather than notes:

1. **No gate may be authored against a check type without first proving the type is
   in the running binary** — and the marker must be LONG, because Go compiles short
   string literals to immediate comparisons that never reach rodata (`grep -ac
   "selector_count"` returns **0** on a binary that fully supports it). A negative
   from a short marker is worthless.
2. **`skip` is the wrong default for a stage gate.** A stage that cannot evaluate its
   own question has not passed it; it is *inconclusive*. This is the same fix G4
   already names (adopt `experienceVerdict`), and this ladder needs it more than the
   tool chain does, because a ladder's whole value is that stage N's pass licenses
   stage N+1.
3. **S7 is therefore not optional.** A gate that passed against one binary is not
   evidence against the next one, which is precisely what S7 exists to say.

## The proposed unit: a build step with a travelling doc

One small part of a site build — a component, a section treatment, a page's
interactive behaviour — gets **one PLAN and one NOTES stream**, in the same tables the
tool lane already uses, with the same discipline:

- the PLAN holds **aim, contract, hazards-and-their-answers, deliberate decisions
  ("do not re-fix"), and the ```criteria fence**;
- the NOTES are append-only and hold every repair, dead end and correction;
- both are written by whoever does the work, at the time, not at handoff;
- **the fence is authored one criterion at a time, and never written until its author
  has watched it pass by hand** — the pilot's own rule, and the reason `smart-contrast`
  passed first complete run.

The `teaser-reveal-panel` directory is already 80% of this shape in files. The proposal
is to make it a DB row so the platform can read it, which is exactly what TL-017 did
for tools.

## The proposed stage ladder

Cheap to expensive, same principle as the tool ladder. Each stage has **one question**
and **one gate**, and the gate is a check that can go red.

| stage | the question | the gate | what it would have caught |
|---|---|---|---|
| **S0 shape** | does this shape already exist? | a named `experience_pattern`, or a written justification for a new one | *(passed in Round 1 — this is the stage that worked)* |
| **S1 contract** | is the contract sound, and are the hazards answered? | every field has `llm_guidance`; every named hazard has a concrete answer or an explicit accept; fence drafted | Round 1's four hazards; Round 4's ellipsis collision would have been a *known* collision |
| **S2 template** | does it render, and are the checks real? | harness green **and ≥1 mutant red per assertion class**, where **a mutant counts only if the harness proves the artefact changed** (report mutants *applied*, not *attempted*) | the harness counting CSS as markup; the `strings.Index` panic; **a `sed` mutation that silently applied to nothing** |
| **S3 register** | is it reachable? | present in `content_components`, returned by `load_component_library`; JS delivered by the route that page type actually uses (see correction) | the `js_content` publishes-but-inert trap |
| **S4 place** | is the placement protected by the mechanism that governs THIS page type? | `site_plan_sections` for planned sections; **`pages.rebuild_policy='owned'` for owned tool pages** (see correction) | **the panel silently dropped by a `complete` re-render** |
| **S5 serve** | does the visitor get it? | fetched page, `<style>` sliced away before counting; 0 unrendered `{{`; contrast measured in the state that needs a click | the local-copy probe that proved nothing |
| **S6 operate** | does it *work* when driven? | **real clicks in real Chromium on the live URL**, desktop + mobile, via `browser-runner-adapter` and the fence's `interaction` checks | **Round 5. The whole reason for this document.** |
| **S7 regress** | does it still work? | S5 + S6 re-run after any roll, rebundle or rerender | a same-tag roll or rebundle silently reverting a DB-side repair |

Two properties are load-bearing:

**A stage gate must be able to fail.** Generated or hand-written, an assertion nobody
has seen go red is not evidence — and G4 shows the platform has already shipped a judge
that passes on an empty result set. Every gate needs the anti-vacuous rule: report the
count you measured, not the absence of failures. `probe_reveal_open_state.py` already
does this and it should be the house style.

> **And the mutation itself needs the same rule applied one level up — found by the
> first forward run, 2026-07-30.** That lane's template harness asserted
> `if mutated == tpl { ERROR: mutant did not change the template }`, but its *ad hoc*
> fence mutations went through `sed`, and **one silently applied to nothing** because the
> pattern spanned two lines and `sed` is line-based. **A mutation suite that mutates
> nothing reports a full set of green checks.** They caught it only because the gate
> prints the count it measured; the verdict line alone said SATISFIED and looked fine.
> **So: a mutant counts only if the harness proves the artefact changed, and the gate
> reports mutants APPLIED, not attempted.** This is the fifth instance in one day of the
> single class this ladder exists to defeat — a check whose failing branch was never
> exercised — and the first one found by somebody other than me.

**A later stage may not be assumed from an earlier one.** Round 5 is the proof: S2–S5
all genuinely passed while S6 had never once succeeded. This is the same argument as
TL-012 — *"completeness + validation passed" ≠ working* — one level down.

### THE LADDER HAS NOW BEEN RUN FORWARDS ONCE — and it corrected itself twice

**2026-07-30, `leopardessconsulting`, `ai-vendor-trust-checklist`** (`0bfdf5b2e`,
`docs/leopardessconsulting/tools/ai-vendor-trust-checklist/`). A **tool**, not a
component — which is why it could run the same day: `doc_plans` already accepts
`subject_type='tool'`, so this ladder's blocked P1 migration was never on its path. That
lane read the substrate status correctly and did not wait for us.

**What it did:** authored the fence *before* building (S1, `fence_check.go`, 7 rules, all
7 proven able to fail), a render harness with **12 checks and 12 mutants, all red** (S2),
then placed it and drove it in real Chromium on both profiles (S6):
**18 pass, 3 fail, 0 unexpected skips.** The S2 mutation requirement — the gate I claimed
nothing in the tools chain had — was adopted in full, on the first try, by a different
lane. That is the strongest available evidence the ladder is transmissible rather than
personal.

**It corrected two of my eight gates, and both corrections are better than what I wrote:**

- **S3 was too narrow.** I generalised from *section* components, where `js_content`
  publishes to `/tools/assets/` but the assemble injects no `<script>` tag, so the working
  route is a `js_snippets` row. **For a tool page the opposite is true:**
  `rerender_single_page_action.collectJSAssets` reads `content_components.js_content` and
  emits `tools/assets/{function}.js` as part of the page's own commit. So the asset path is
  **derived from `function`** rather than typed into a template — which makes a
  `<script src>` mismatch *structurally impossible* rather than merely checked for, and
  that exact mismatch is a live defect on `llm-cost-calculator`. **S3's gate is therefore
  "delivered by the route this page type actually uses", not one named table.**
- **S4 did not apply at all, and asserting it would have blocked the build.** Their site's
  `site_plans` row has **zero** `site_plan_sections` rows; the mechanism actually
  protecting its four other tool pages is `pages.rebuild_policy='owned'`, a hard refusal in
  `save_page_sections_action.go`. And because `page-rerender`'s `save_sections` step *is*
  the generic save that guard refuses, **`owned` blocks the initial render** — so the order
  is forced: render with `generic`, then flip to `owned`. Verified as a red/green pair
  (with `owned` set the same render refuses and the served page stays byte-identical).
  **A gate that names one mechanism when the platform has two, in tension, is a gate that
  will be wrong half the time.**

**And it confirmed D3 independently, with a fleet measurement I did not have.** The
skip-reads-as-pass hazard turns out to have a second, entirely different trigger: three
values must be equal —
`doc_plans.subject_key == pages.name == content_components.function` — or `load_docs`
returns an empty fence and `request_browser_run` **SKIPS with `needs_criteria`**, which is
honest but is *not a failure either*, so it reads as a clean run that asserted nothing.
Measured: **6 of 22 hosted tools fleet-wide cannot be acceptance-tested at all** until
renamed, across five sites — including two on the site they were working on. D3 was
reasoned; this is measured, and the population is large.

### Two rules for authoring a gate, adopted from the tools lane

The `webdesign_tools_repair` lane reached the same conclusion this ladder is built on,
**on the same day, from a different direction, and its two rules are better stated than
mine.** Adopted verbatim in substance:

1. **Verify through the visitor's gesture, never through the subject's internal
   functions.** If the entry point is a paste, dispatch a paste. **If the vocabulary
   cannot express the gesture, that is a MISSING CHECK TYPE to record as a deferral —
   not a licence to substitute a function call.**
2. **A gate must assert the TERMINAL value, not the first observable state change.**
   *"Status reads LIVE EDITING"* is a waypoint; *"text can be edited and emphasised"*
   is the point. A fence asserting the waypoint passed while the tool was unusable.

**Why the convergence is worth recording rather than just the rules.** That lane
verified `pasteboard` by calling `addItem()` and `logic-architect` by calling
`loadTemplate()`; both returned the right answer, so both "passed", while a visitor
could reach neither. This lane forced `.open = true` directly on DOM nodes and called
it verified for four rounds. Those are not two similar mistakes — they are **one defect
class: verifying through a privileged path the visitor does not have.** Two lanes, two
subject types, one day, arriving independently. That is about as strong as evidence for
a rule gets, and it is the reason S6 is a named stage rather than a line in a checklist.

The corollary generalises past both lanes and is the strongest single argument for the
browser tier: **a property of the COMPOSITION can only be checked in the composition.**
`has_visible_area` exists because an element can be in the DOM and measure 1146×0;
`features_open/026` exists because a colour can be valid in the palette and invisible
on the page. Same shape, two altitudes, and both invisible to everything that reads a
source.

## On dynamically created checks (the owner's question, answered with evidence)

**It is already precedented, and it already has a measured failure mode.** The
tool-generator writes the criteria fence **at birth**, by LLM, under composer rules
(migs 131/158/162) — so "a check created at the start of the project" exists in
production today. And TL-016 records that same composer **inventing selectors twice**.
The anchor rule exists precisely because a generated static check cannot be trusted to
refute.

So the honest answer is: yes, generate them, under three conditions that the platform
can already enforce.

1. **Every generated criterion passes the exported ten-rule validator before it is
   stored.** This is G3's fix and it is the same code the experience register already
   runs. Validation, not sterner prompts.
2. **No criterion is written until someone or something has watched it pass.** The
   pilot's rule. A fence authored against a page nobody looked at is a fence about an
   imagined page.
3. **A generated criterion must come with the mutation that makes it go red**, or it
   does not count as a gate. This is the one thing the carousel did right from the
   first commit and it is what caught the harness's own defects.

Stage-scoped generation then falls out naturally: at S1 the fence is drafted from the
contract; at S4 the placement assertions are generated from the actual
`site_plan_sections` row; at S6 the `interaction` checks are generated from the
controls the template actually emits. Each is generated from a real artefact that
exists by then — which is what makes it groundable rather than invented.

## Ordered proposal — what to build, by what closes the door

1. **Component travelling docs.** Give a component a PLAN with a criteria fence and a
   NOTES stream, in `doc_plans`/`doc_notes`. Reuses TL-017's tables and write path
   wholesale. **Without this there is nothing for any stage gate to read**, so it is
   first regardless of everything else.
   **SUBMITTED 2026-07-30** — council `e5673868-7c5b-489c-931a-7ba59b959b91`, commit
   `c659e312b`, migration **not applied** (image first).
   > **CORRECTED, and the correction is the most useful thing on this page.** I costed
   > this as "one migration extending two CHECK constraints" and staged the DDL outside
   > `sql_for_agents/` to keep it from being applied by accident. **Both halves of that
   > were wrong.** `subject_type` has a **second enforcement point in Go** —
   > `validDocSubjectTypes`, gating every doc action — so the DDL alone would have
   > reproduced **`bugs_open/064` a third time** (163 missed a gate; 184 moved the DB
   > CHECKs only and left its own seeded docs unreachable). And the migration **must** be
   > numbered, because `TestValidDocSubjectTypes_LockstepWithMigrationCheck` parses the
   > newest numbered file and fails on drift — so withholding it does not protect
   > anything, it reddens HEAD. Shipped as one commit carrying both halves, mutation-proven
   > (Go half alone → the test fails naming 184's failure mode).
   >
   > **Why this belongs in a document about stage gates:** I priced a platform change by
   > reading *one* enforcement point and calling the result "the smallest possible
   > change". The ladder exists because that class of miss is normal, and the fix is
   > never "be more careful" — it is a gate that reads the other point for you. There is
   > an existing checklist for exactly this
   > (`experience_register/design/subject_type_addition.md`) and it names all four
   > enforcement points; I found it only because a code comment pointed at it.
   Still true and worth keeping: the change is additive and inert, so under the
   2026-07-29 owner ruling §1 it is **normal council-gate scope, not an RFC**.
2. **S6 via the existing browser-runner.** A component-scoped acceptance run that
   dispatches the fence to `browser-runner-adapter`, exactly as `tool-acceptance-agent`
   does. Reuses the mechanism the `smart-contrast` pilot proved end to end. Closes the
   class that cost five rounds.
3. **Route component fences through the ten-rule validator** — G3's fix, same exported
   code, applied to a second corpus.
4. **Fix G1** so a repair agent reads the PLAN's deliberate decisions before touching
   anything. One prompt line plus one migration, per the tools-chain report. Cheap, and
   it is what makes "do not re-fix" mean something.
5. **A generic component render harness**, so each component does not hand-roll its own
   Go file that grows 14 → 24 checks by hand. Fix `gen_component_register_sql.py`'s
   hardcoding to evidence-chart's description while in there.
6. **Only then** stage-scoped dynamic check generation, and the anti-vacuous verdict
   rule (G4) — noting that G4 changes what a shared mechanism *guarantees* and therefore
   goes to the council gate and plausibly an RFC under RFC_002's ratified trigger.

Steps 1–3 are additive opt-ins: normal council gate, register in the same commit, no
RFC needed. That is worth knowing before anyone starts.

**One clarification the tools lane's revision forces, and it sharpens this ladder.**
The owner's validation-versus-judgement correction (recorded in that lane's §4, which
split its G2 into a cadence fix and a separate judgement seat) applies here directly:
**every gate in S0–S7 is validation** — a closed question with a fixed rule, the same
answer every time it is asked. *"Is this component any good? Is this treatment right
for this page?"* is **judgement**, and it belongs to a reviewer seat, not to a stage
gate. Keeping them apart matters because conflating them produces the worst of both:
a gate that drifts per component and a judgement boxed into a checklist. So this ladder
deliberately has **no aesthetic or editorial gate**, and that is not an omission.

## Open questions the next thread should decide, not inherit

1. **One checker agent parameterised by stage, or one per stage?** `bugs_open/149` is
   the cautionary evidence: 22 discovery handler agents exist and only 2 run
   `validate_page_content`; six registered checks sit in no agent at all. Proliferation
   here has already produced dead checks.
2. **Who fires the stages, and when?** This is G5 and it is the question most likely to
   make the whole thing inert. Discovery passes are manual-fire; the improvement loop is
   stopped by owner ruling. A ladder with no trigger is the "mechanism rotting
   unexercised" cost the owner has already ruled against paying.
3. **Does a stage gate get to *refuse*?** A gate that can block is a guarantee change
   under the 2026-07-29 owner ruling and goes to architecture review. A gate that only
   reports is additive. These are genuinely different features and the answer changes
   the build order.
4. **Relationship to `features_open/015`, the staged site maturity ladder.** That is
   this same idea one level up — sites climbing named rungs against worked examples,
   with per-rung promotion criteria, *"ideally checkable the way discovery checks
   already work"*. It is REQUESTED and undesigned. **These two should probably be one
   design with two altitudes rather than two designs**, and deciding that early is
   cheaper than reconciling later.
5. **Relationship to `features_open/026`.** That feature's finding is that no check
   *renders* — the vantage point is missing. Its Phase 3 (browser-runner on the deploy
   path) is a sibling of S6 here. **Do not build two of these.** 026 is page-and-palette
   scoped, S6 is part-and-interaction scoped; the overlap is the dispatch, and the
   dispatch should be shared.
6. **Does `bugs_open/151`'s candidate (3) belong in this ladder?** A post-build
   fact-repetition census is exactly a stage gate, and it is the only 151 candidate that
   also protects the nine already-deployed sites. It may be cheaper to build it *as* the
   first content-stage gate than as a standalone check.

> **Questions 4 and 5 now have a proposed answer, from the tools lane rather than from
> here — recorded as PROPOSED, not settled** (it is the owner's call, and 015 has never
> been designed by anyone). Its formulation: **`015` is the rung vocabulary, `027` is the
> gate mechanism, `026` is the missing instrument, and the existing tool chain is what
> all three point at.** That is a better decomposition than "these should probably be one
> design", because it makes them composable instead of merged — three things at three
> scales that share a dispatch, rather than one document trying to be all of them. It
> also means this lane does **not** need to own 015 in order to proceed, which unblocks
> everything here. What still needs the owner: whether 015 stays a separate thread at all.

## What this proposal deliberately does not claim

- **It is not measured.** Every figure in Part 1 is evidenced and dated; Part 2 is a
  design argument. The claim that stages would have caught Round 5 is a **reasoned
  inference [INFERRED]**, not an experiment — though it is a strong one, since Round 5's
  fix was found the first time anyone ran the S6-shaped check.
  > **PARTIALLY DISCHARGED 2026-07-30 (evening), and stated precisely so it is not
  > overclaimed.** The first forward run (above) proves the ladder *catches real defects
  > nobody was looking for* — an S2 check found a two-file id contract that the platform's
  > own `OrphanElementRefs` returns nil on, which is how `llm-cost-calculator` shipped
  > pointing at another tool's JS filename; and S6 surfaced `bugs_open/157`. **It does not
  > discharge the Round 5 counterfactual**, which is a different claim about a different
  > defect, and no run has tested it. The marker stays. What *has* changed is that
  > "would a stage have caught anything?" is now answered yes, by someone other than the
  > author, on their first attempt.
- ~~**It does not know that the reuse is as clean as it looks.** Whether the component
  case fits `doc_plans`' shape without schema change is **[UNVERIFIED]**.~~
  > **RESOLVED 2026-07-30, and the marker did its job.** Read the two tables: the reuse
  > is clean in shape and blocked by a one-line CHECK constraint, with a four-times
  > precedent for extending it. Full answer in the box at the top. Recording this as a
  > worked example of the practice rather than deleting it: **the `[UNVERIFIED]` marker
  > is what made this the first thing anyone did**, and the answer arrived in two
  > queries. An unmarked assumption would have been designed against for a week.
- **It does not propose re-enabling the improvement loop.** That is ruled stopped.

---

## Cross-links

- **The carousel itself:** `brochure_component_library/PLAN_2026-07-29_teaser_reveal_panel.md`,
  `NOTES_brochure_component_library.md` (entries from line 2713),
  `SUMMARY_2026-07-30_the_panel_is_finished_and_two_new_fronts_open.md`,
  `components/teaser-reveal-panel/`, `components/README.md` (the per-component
  acceptance checklist this proposal generalises).
- **The tool provenance / travelling-docs research:**
  `docs024_key_docs_latest/travelling_docs/OVERVIEW_self_verifying_tools.md` (the
  plain-language tour), `RUNBOOK_travelling_docs(38).md` §0 (the Stage 0–6 tracker),
  `PLAN_travelling_docs(7).md`, `tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md`,
  `037_TOOL_DOCS_convention(1).md`.
- **The current state of the tools chain, with the gaps measured:**
  `webdesign_tools_repair/REPORT_2026-07-29_concepts_for_a_working_tools_chain.md`.
- **The dedup work:** `bugs_open/151_HANDOFF_2026-07-29_section_writer_has_no_memory_of_facts_already_used.md`,
  plus 016b §9's transferable pattern.
- **Register entries:** TL-008 (the ladder), TL-017 (criteria in the PLAN), TL-012,
  TL-016, TL-033, DOC-003/DOC-010 (travelling doc conventions), CLC-012 (this carousel).
- **Adjacent features:** `features_open/015` (site maturity ladder), `026` (render before
  ship), `017` (component adoption — the planner never selects the new components).
