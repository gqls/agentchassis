# 253 — a framework rewrite of a decomposed prose block strips every layout component

**Filed 2026-08-11. GUARD SHIPPED 2026-08-12** (`0c8e08ccb`,
`Council-Submitted: b30ac52c`); **the live page was already repaired by then, by a
different route — see "Two remedies" below.** Observed live on
`loanandmortgagecalculator.co.uk/index.html` within four hours of that page being
decomposed. **This is the finding that governs Track B**, and it was not predicted
by any of the decomposition briefs.

> ### ⚠ NUMBER COLLISION — `253` names TWO unrelated bugs
> The other is `253_..._label_match_overlap_count_ties_on_incidental_nav_label_words`.
> Its fix commits (`c6dcbcaa8`, `6ea633cea`, `9b7811d4b`) are **not** this bug's, and
> a `git log` by number will hand you the wrong case. **Refer to this one by slug.**

## ⛔ COUNCIL: REVISE (round 1, `b30ac52c`) — and the objection is RIGHT

**Gating objection, `bug_historian`, severity HIGH, on the wiring edit:**

> *"The guard is wired only into SavePageSectionsAction… several OTHER writers of
> `page_components.rendered_html` … are documented as capable of writing this column
> without going through save_page_sections. This is the '016b §9' pattern 'one call
> site of a shared judgement gets the rigorous fix; the sibling stays heuristic' — if
> any of those paths bypass SavePageSectionsAction, they bypass BOTH the pre-existing
> text shrink floor and this new component floor, and a flattening save through one
> of them will fail exactly as silently as the bug this plan fixes."*

**Audited 2026-08-13, as the seat asked. It is worse than the objection states.**

Nine Go writers touch `page_components.rendered_html`; **one is guarded**:

```
adopt_verbatim · create_report_page · create_tool_component · deploy_tool
fix_forced_text_colours · fix_harcoded_colours · rebuild_blog_listing
section_editor_actions (ApplySectionEditAction)   ← 3 UPDATE sites
save_page_sections                                ← the ONLY guarded one
```

**The one that matters is `ApplySectionEditAction`**, and the reason is not its row
count. It is **live** (`section-editor` agent definition), it does
`UPDATE page_components SET rendered_html = $2` directly, and **it is precisely the
per-component edit path that decomposition exists to enable** — 10c §3's stated
benefit is *"after decomposition you can rewrite one prose block without touching the
calculator"*, and that is this action. So the guard covers the door the observed
incident happened to come through, and misses the one the whole design steers future
edits toward. Also live and unguarded: `rerender-pages`, `report-builder`,
`tool-generator`.

**The seat also identified something neither of us can close from here**, and it is
correct: *"Whether the existing text shrink guard is wired into every
page_components writer, or only into save_page_sections_action — this plan's
coverage question is inherited wholesale from that guard."* It is inherited. The
`bugs_open/178` floor has exactly the same single-call-site coverage, so **both
floors have been protecting one door of nine since 08-02**, and nobody noticed
because the incident that motivated each of them came through the guarded one.

**This is the same defect I filed against other people's code in `251`/`252`** — the
landmine that `injectCanonicalLink`/`injectPageJSONLD`/`injectRobotsNoindex` live on
one head producer only. I cited it, then reproduced it. The memory entry for it is
literally *"a guard only guards the door you walk through"*.

### ✅ ROUND 2 DONE AND RESUBMITTED (2026-08-13, same correlation `b30ac52c`)

- **`enforceSingleSlotFloors`** (`single_slot_floors.go`) — the single-row form of
  **both** floors, wired into `ApplySectionEditAction`'s `content_edit` branch.
  **One function composing the two existing pure decisions, not a second copy**:
  pasting the logic into a second call site would reproduce the very defect
  objected to, with an extra copy to drift. `component_swap` deliberately NOT
  guarded — it changes `component_id`, `slot_name` and `html` together, so its
  markup is *supposed* to differ.
- **⛔ The part that matters, and it came from an induction that FAILED.** After
  wiring the second call site I deleted that wiring again to check something would
  catch it. **Nothing did** — the whole package still passed, because the unit tests
  exercise the decision functions and are blind to whether anyone calls them. *A
  guard nothing proves is reached is the same defect one level up.* So the class is
  now a test (`page_component_writer_coverage_test.go`): every file that `UPDATE`s
  `rendered_html` must enforce a floor or sit in `exemptWriters` **with a reason**,
  and a tenth writer fails it until its author decides in writing. Re-induced —
  unwiring the section editor now fails it **by name**.
- **It earned its keep immediately**: it caught `create_report_page_action.go`,
  which my *manual* audit had filed as create-only and which in fact looks up and
  **overwrites** its own report row. Classified, not waved through.
- Exemptions are decisions with reasons; two are marked `[UNMEASURED]` (the colour
  fixers are believed structure-preserving on a code reading, not an experiment).
- **Stated weakness**: the coverage test reads SOURCE, so it proves wiring EXISTS,
  not that it EXECUTES. Strictly more than the zero we had; the behavioural half
  belongs in each action's own test.

The seat's second, low-severity objection is also fair and is now stated as residual
exposure rather than left implicit: `minComponentGuardClasses=10` means a flattening
of a slot just under the threshold produces the same silent no-refusal this bug
exists to close.

## Two remedies, and the one that actually repaired the page was not the code

**The live homepage is no longer flattened.** Re-measured 2026-08-12: `class="card"`
**12**, `tool-grid` **2**, `btn-primary` **12**, `highlight-box` **1**, `hero` **1**
(prose-0 rewritten 16:03). It was repaired by the lane session seeding a
`content_direction` telling the writer the cards are good and stay, then re-running —
**not** by any change to the platform.

That is the important lesson and it should not be lost in the fix: **the writer was
not malfunctioning, it was uninstructed.** Handed a block of markup with no
description of what the markup means, it produced clean prose. Told what the page's
vocabulary was, it kept it. So the primary remedy for a flattening is
`content_direction`, and the guard below is the **safety net for the case where
nobody thought to write one** — which is exactly the case that occurred here, and
will occur again on the next site decomposed by someone who does not know to.

The guard's own refusal sentence says this, deliberately: it directs the reader to
give the writer the component vocabulary rather than to lower the floor.

---

## What happened

Track A decomposed the LMC homepage at 11:16Z into a single `prose-0` component
holding the page's hand-built body byte-for-byte. At **15:47Z** the generic pipeline
rewrote that component. The rewrite kept the words and the links. It removed the
site's entire visual vocabulary.

| markup in the block | before (Track A, byte-identical to the hand-built page) | after the rewrite |
|---|---|---|
| `class="card"` | 18 | **0** |
| `tool-grid` | 3 | **0** |
| `btn-primary` | 15 | **0** |
| `highlight-box` | 1 | **0** |
| `class="hero"` | 1 | **0** |

Links survived — 14 calculator links before and after, and *more* internal links
overall (28 → 34). So this is **not** a content or navigation loss, and the first
read of the diff (mine) wrongly suggested it was. It is a **presentation** loss: the
site's shopfront went from a styled calculator directory to a flat run of headings
with bare "Open calculator" links. `prose-0` went 6,958 → 5,832 bytes.

## Why this is the important one for Track B

A decomposed calculator page is `["prose-0", "tool-1", "prose-2"]`. The tool row is
**locked**, so the calculator itself is safe — `bugs_open/058`'s lock holds it, and
the matching rule is now pinned by
`save_sections_positional_tool_slot_test.go`. **The prose rows around it are not
locked and are exactly what was rewritten here.** So Track B's expected outcome is:
the calculator keeps working and keeps its markup, while the cards, buttons and
grid framing it are silently flattened on the next generic rebuild.

**That was a materially different risk from the one Track B was authorised against.**
Every brief to date framed the calculator page danger as "the widget gets replaced
by prose or moved to the bottom". Both of those were already guarded; this one was
not, and it lands on 22 live consumer-finance pages.
**As of 2026-08-12 it is guarded** — see fix candidate 1 below — so Track B's
per-page prose rows now refuse a flattening rather than absorb it silently. **The
guard is not yet ROLLED**: it is committed, not running, until the next chassis
build. Do not treat Track B as protected before a pod is serving it, and prove that
the way this lane proves everything else — not by the tag.

## What is NOT wrong here

- **The framework rewriting this copy is correct and intended.** Owner ruling
  2026-08-06: the framework writes the content, not a CLI session. Decomposition
  exists precisely so it *can*. Do not read this bug as an argument against
  decomposition.
- **The shrink guard worked.** An earlier attempt at 15:24Z was REFUSED —
  `prose-0 3776→1334 chars (35% kept, floor 50%)`, `bugs_open/178`, nothing written,
  and it raised `save_refused_incomplete:index` for a human. The 15:47Z save was
  within the floor (84% kept) and proceeded. **The guard measures TEXT VOLUME, and
  is blind to markup.** A rewrite can keep 84% of the words and 0% of the components.
- **No other Track A page was touched** — only `/index.html` has a `page_components`
  row updated after 12:00Z. This is one occurrence, not a sweep.

## Root cause — where to look

Not yet established, and this is the honest state of it. `[INFERRED]` The writer is
handed the section's content and asked to produce the section's content; nothing in
that contract obliges it to preserve component classes it did not author, and a
`ported-prose` block has no schema describing what its markup means. The likely fix
sites are the `page-content-writer` prompt/config and whatever validates a section
write — but **that is a hypothesis, not a diagnosis, and it should go through `090`
before anyone asserts it.** The observation above is solid; the cause is not.

## Fix candidates, ordered by what closes the door

> **(1) IMPLEMENTED 2026-08-12** — `platform/orchestration/actions/save_sections_component_floor.go`,
> commit `0c8e08ccb`, `Council-Submitted: b30ac52c-e42d-4110-bd22-fce5598b3bf7`
> (verdict not yet read — do **not** upgrade that trailer to `Council-Reviewed:`
> without reading it). Calibrated on the real before/after rather than invented:
> prose-0 class attributes **43 before → 1 flattened (0.02) → 31 on the good rewrite
> (0.72)**, i.e. the bad and good cases are **35× apart**, so 0.25/0.34/0.50 all
> separate them. Default 0.5 mirrors the text floor. Scope threshold 10 class
> attributes, from a fleet distribution of median 5 / p90 35 over 1,422 unlocked
> slots (~31% of slots in scope). Counts class ATTRIBUTES, not tokens, and is
> deliberately blind to WHICH classes — a rewrite swapping one valid vocabulary for
> another passes.
>
> **Stated weakness:** the safety evidence is **one** good rewrite. The floor is
> DEFAULT ON, so it changes behaviour for every `save_page_sections` caller on the
> first roll; that call is flagged for the council explicitly rather than buried,
> and its sibling shipping default-on at the same 0.5 is the precedent relied on.

1. **A markup-preservation floor beside the text floor.** The shrink guard already
   exists and already raises a human-reviewable item; it is the natural home. Assert
   that a same-named prose slot may not lose more than N% of its *component class
   occurrences* in one save. This makes the bad state detectable by the mechanism
   that already stops the analogous text case, and it is the smallest change that
   would have caught this.
2. **Lock prose rows that carry layout components.** Blunt: it also stops the
   legitimate rewrites decomposition exists to enable, so it trades the whole benefit
   for the protection. Not recommended except as a stopgap on specific pages.
3. **Give `ported-prose` a schema the writer must satisfy.** Correct in the long run,
   much larger, and it is really a question about how verbatim-adopted markup gets
   described to a writer at all.

**Do not "fix" this by restoring the page and calling it done** — it will be rewritten
again by the next rebuild, and the second occurrence will look like a new bug.

## ~~Immediate decision owed on the live page~~ — RESOLVED 2026-08-12

~~The LMC homepage is currently serving the flattened version…~~ **No longer true.**
The page was repaired by `content_direction` before the guard shipped (see the top of
this file). `load_lmc.py --apply index` would still REFUSE — its pre-write guard sees
the stored md5 has moved from the 08-09 baseline, which is that guard working, not a
problem — but there is nothing left to repair.

## See also

- `bugs_open/178` — the shrink floor that fired at 15:24 and correctly held.
- `platform/orchestration/actions/save_sections_positional_tool_slot_test.go` — why
  the locked tool row itself is safe.
- `loanandmortgagecalculator_couk/NOTES_…md`, 2026-08-11 — Track A's full record,
  including that a `predicted/` file is only valid until the framework next writes
  the page. That is how this was caught: a post-roll mirror check flagged the
  homepage as differing, and the difference turned out not to be the roll at all.
