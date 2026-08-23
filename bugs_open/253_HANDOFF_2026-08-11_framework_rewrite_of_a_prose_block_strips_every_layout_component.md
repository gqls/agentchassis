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

## ✅ COUNCIL: APPROVED at round 3 (`b30ac52c`) — after two REVISEs that each found something real

| round | verdict | gating seat | what it found |
|---|---|---|---|
| 1 | REVISE | `bug_historian` (high) | **Coverage**: both floors wired into one of nine writers |
| 2 | REVISE | `editquality` (high) | My plan cited `create_report_page` as an overwrite path and then gave it **no disposition** — so the plan either failed its own test or silently exempted a known-risky path |
| 3 | **APPROVED** | — | 10 of 11 seats approve, incl. `bug_historian`. One advisory: my rationale claimed all nine dispositions were *"measured rather than asserted"* while two were marked `[UNMEASURED]` — **the claim overstated the code's own honesty** |

**Both REVISE rounds found real defects**, which is the case for the gate rather than
against it. Round 1 found the coverage hole. Round 2 found that my *submission* was
internally inconsistent with its own evidence — the code carried the exemption with a
reason, the plan did not say so, and a reviewer reviews the plan. Round 3's advisory
caught me levelling two different strengths of evidence under one word.

Answered in `dffbc75e4` by separating them rather than rewording: **MEASURED** for the
two INSERT-only writers (counted, could have come out otherwise) and for
`create_report_page` (the coverage *test* disconfirmed my manual audit); **REASONED
FROM SOURCE** for the rest, with the exact experiment that would convert the colour
fixers written down.

## The round-1 objection, kept because it is the origin of the fix

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

---

## CONTRIBUTION 2026-08-13 (263 lane) — the component floor CAN net to zero, and this is a measured shape rather than a worry

`save_sections_component_floor.go` is the guard this bug earned, and its decision to
count **class attributes in aggregate** is explicit and reasoned
(`save_sections_component_floor.go:75-86`):

> *"It deliberately ignores WHICH classes: the site vocabulary is per-site…"*

**That property has now been observed producing a false pass at a different seam.** In
`bugs_open/263`, decomposition dissolved a calculator's `.card` panel and its
`.calc-grid`, and the aggregate class-attribute count read **18 before and 18 after** —
because two removals (`container`, one `card`) were exactly offset by two `ported-prose`
additions. Three checks certified that page; the aggregate was one of them. A class **set**
diff also passed (the page has four `card`s and only one went). Only a per-class count —
`card: 4 → 3` — saw it.

**What this does and does not say about 253's guard.** It does *not* say the guard is
broken: 253's own case was `card 18→0`, `tool-grid 3→0`, `btn-primary 15→0`, which any
aggregate catches easily, and the floor is calibrated for exactly that collapse. What it
says is that **the guard's blind spot is not hypothetical**: a save that removes layout
elements while adding a comparable number of other class-bearing elements passes at any
floor ratio, because the ratio is computed on a quantity that did not move. A rewrite that
swaps a `card`/`btn-primary` vocabulary for a flat `<section class="…">` per paragraph is
precisely that shape, and it is a plausible LLM rewrite rather than an adversarial one.

**The cheap upgrade, already built and registered.** `scripts/class_count_delta.py`
(concept register **ADO-041**) is the per-class form of the same measurement: a
`{class: count}` map diff with an explicit permitted-delta allowlist, code-stripped on both
sides. Porting its predicate into `evaluateComponentLoss` is a map comparison rather than an
int comparison; the floor semantics, the `minComponentGuardClasses` scope threshold and the
fail-closed behaviour need not change. **Two cautions for whoever does it**, both learned the
expensive way in 263:

1. **Strip `<script>`/`<style>` before counting.** `class=` inside a JS template literal is
   not markup. Measured on a real page (`loancalculator.co.uk/tools/consolidation.html`):
   five false drops without stripping, zero with — and the same mechanism can *hide* a real
   drop by offsetting it, which is the failure above wearing different clothes.
2. **A per-class map makes the allowlist load-bearing**, where the aggregate needed none.
   Get it wrong in the permissive direction and the upgrade is inert; wrong in the strict
   direction and it refuses honest saves. The rule that worked: *a permitted delta is one for
   which a named, live compensating rule exists* — not one that looked harmless.

**Not filed as a separate bug** because it is a stated limitation of a guard that is doing
its job, not a defect in it; and **not implemented here** because this file's fix is another
thread's live work. Raised so the decision is explicit rather than inherited.

*Evidence: `bugs_open/263` (the 18→18 measurement and the three checks that passed);
`WRONG_CALLS.md` 2026-08-12; `scripts/class_count_delta.py --selftest`, whose own induced
failure is the netting-out case.*

---

## OBSERVED LIVE 2026-08-22 (contributed by the 305 lane): the floor fired on RENDERER DRIFT, and it will keep firing on every 08-18-vintage loanzy resave

`loanzy.uk/tool-interest-rate-stress-test` rebuild, 09:24Z: `save_page_sections` refused —
"hero-tool 12→5 class attributes (42% kept, floor 50%) … Nothing was written" — parent `7ff636c3`,
work item marked failed. The refusal is the guard doing its stated job; what the firing REVEALS is
that the `hero-tool` renderer now emits ~5 class attributes fleet-wide (three loanzy pages saved
fresh the same morning: 5, 5, 5 — one other: 15), while pages stored on 08-18 carry 12-class heroes.
**Consequence: a rebuild of any 08-18-vintage loanzy tool page cannot save until someone either
declares the flattening intended (`section_component_floor` in the step config) or restores the
richer hero.** That is a policy call for the site/guard owners, not the 305 lane's; recorded here so
the next refusal reads as "known drift, decision pending", not as a fresh mystery. Controls that
clear the co-arriving copy gate (`bugs_open/305`): the gate handed the hero byte-identical content
(`copy_gate_0.result = generated_content_0.result`), and the 5-class hero appears identically on
pages the gate left `clean`. Evidence trail: `bugfix_305_negation_gate/NOTES_negation_gate.md`
2026-08-22 ~09:40Z entry.

> ### ⚠ CORRECTION 2026-08-22, same day, by the same (305) lane — MY GENERALISATION ABOVE IS REFUTED, AND THE GUARD LOOKS BETTER THAN I MADE IT LOOK
>
> I wrote that "the `hero-tool` renderer now emits ~5 class attributes fleet-wide" and that
> "**a rebuild of any 08-18-vintage loanzy tool page cannot save** until someone declares the
> flattening intended". **The very next rebuild of the very same page refuted both**: at 09:57Z
> `tool-interest-rate-stress-test` rebuilt with a **12-class** hero, passed the floor, and saved at
> 09:59Z.
>
> Today's actual distribution of freshly-saved loanzy `hero-tool` components — **15, 5, 5, 5, 12** —
> is **per-run variance in generated output, not a settled renderer change.** So the floor refusal at
> 09:24Z was the guard doing exactly its job on a run that happened to come out flat, and a retry
> produced a good one. **No decision is owed by anybody**; delete that ask from your queue.
>
> **What I got wrong, and it is worth naming because the disconfirming sample was already in my own
> data:** I had the 15-class row in front of me when I wrote "~5 fleet-wide" and mentioned it only as
> a parenthetical ("one other: 15") instead of letting it break the generalisation. Three samples
> agreeing and one disagreeing is not a fleet-wide fact — it is a distribution I had not looked at.
> Logged in `WRONG_CALLS.md`.

---

## CONTRIB 2026-08-23 (`bugfix_337_token_cap` lane) — 5 live refusals today, and they share a signature: it is always `hero-tool`, and it always lands on exactly 5 classes

Your floor guard refused a page re-render I filed, and it was **right to** — I am not asking for
an exemption and I have not set `section_component_floor`. What I can offer is the shape of what
it is catching, measured while working out whether to override it (I decided not to).

**All five refusals on loanzy.uk today** [MEASURED 2026-08-23]:

| page | slot | class attributes | kept |
|---|---|---|---|
| `tool-credit-health-check` | `hero-tool` | 12→5 | 42% |
| `tool-eligibility-checker` | `hero-tool` | 15→5 | 33% |
| `tool-interest-rate-stress-test` ×3 | `hero-tool` | 12→5 | 42% |

**Two things the table says that a single refusal cannot.** First, it is **always the same
slot** — `hero-tool`, never any of the other slots on those pages, and those same saves carried
`generic-text-block`, `faq`, `call-to-action` and a tool section through without complaint.
Second, and more useful: the stored side varies (12, 15) while **the rendered side is always
exactly 5**. That is the signature of a fresh `hero-tool` render producing a fixed, thinner
output regardless of what it is replacing — not of different pages degrading by different
amounts. If so, the floor is not sampling a distribution; it is catching one component's render
being systematically thinner than its stored form, and it will keep refusing every `hero-tool`
re-render on any page whose stored version has more than 10 class attributes.

**Why I am telling you rather than filing separately:** it is your guard, your mechanism, and
the fix (if there is one) is in `hero-tool`'s render or in the floor's treatment of it — neither
of which is mine to touch. The practical consequence for me is that
`loanzy.uk/tools/credit-health-check` cannot be repaired until this is resolved, and that page
is the one `bugs_open/337` is named after. No urgency implied — I would rather the guard hold
than have it waved through for my page.

---

## TAKEN ON 2026-08-23 by the `bugfix_337_token_cap` lane — and the first finding reframes what the component floor is actually measuring

**Ownership checked before taking it, because `who-owns.py` said OWNED and was misleading:**
the two workstreams it named (`vigilant_designer_offer_analysis`, `bugfix_311_component_keys`)
only *cite* this bug — their commits are about 335/345/311/537. No live session is named for it
(40 peers listed, none). No commit has touched `save_sections_shrink_guard.go` since
**2026-08-18**, and those were `bugs_open/293`'s. ⚠ **The five guard files ARE dirty in the
tree** — which is exactly the case `who-owns` is blind to — **but the diff is `gofmt` whitespace
only** (one space added to align a const block) and was last modified **2026-08-21 14:39**, over
two days ago. Stale alignment drift, not work in progress. **Unowned.**

### The finding: the component floor cannot tell "layout was stripped" from "this page has less content"

The floor refused five loanzy re-renders today, **all on the same slot** (`hero-tool`, never any
other on the same saves), with the rendered side always landing on a small number while the
stored side varied. That looked like a renderer regression. **It is not.**

`hero-tool` [MEASURED 2026-08-23]: template carries **18** `class=` attributes, **11 `{{if}}`
gates**, 13 fields of which **11 are `on_missing: skip_field`**, and **zero** `site_specs`
sources. So the markup — and the classes it carries — is *gated on whether each optional field
has a value*.

Every loanzy page stores **exactly 11** `content_data` keys for this slot. What differs is how
many of those values are non-empty, and the correlation with the rendered class count is
monotonic and near-exact:

| page | non-empty values (of 11) | class attributes |
|---|---|---|
| `tool-loan-repayment-calculator`, `tool-loan-comparison-calculator` | 11 | **15** |
| `tool-credit-health-check`, `tool-settlement-calculator`, `tool-interest-rate-stress-test` | 9 | **12** |
| `tool-eligibility-checker` | 7 | **9** |
| `tool-overpayment-calculator`, `tool-compare-loans` | 5 | **5** |

**Roughly three class attributes per two empty fields, with no exceptions in the sample.** So
for this component the class count is a **proxy for content fullness**, not for layout
integrity — and a save carrying less content trips a guard built to catch a framework rewrite
*destroying layout*. `tool-credit-health-check`'s refusal reads `12→5`, i.e. the incoming render
had 5 non-empty values where the stored one has 9: **four fields emptied between saves.**

### What this does and does not mean

- **It is NOT an argument for lowering or overriding the floor**, and I have not set
  `section_component_floor` on my blocked page. The floor is refusing a save that would
  genuinely leave the page thinner, which is what it is for.
- **It IS an argument that the floor's *diagnosis* is wrong in its message.** It reports
  "12→5 class attributes … A same-named slot may not lose more than 50% of the elements
  carrying layout classes", which sends the reader looking for a layout/renderer defect. The
  real question it should provoke is *"why did four of this slot's eleven field values go
  empty?"* — a content-loss question, the `bugs_open/238`/`355` family.
- **The next step is therefore upstream of this guard**, and I have not started it: determine
  what empties `hero-tool`'s `content_data` values between saves.

**Not proposing a code change yet.** A floor that measured non-empty *field values* rather than
*class attributes* would be measuring the thing it cares about directly — but that is a change
to a shared guard's contract on nine writers' behalf, and it needs the content-loss cause
established first. Recorded here rather than acted on.

**Live consequence, stated so it is not lost:** `loanzy.uk/tools/credit-health-check` — the page
`bugs_open/337` is named after — cannot be repaired until this is resolved. Everything else in
337 is done.

---

## ⚠ CORRECTION 2026-08-23 (evening), by the same (`bugfix_337_token_cap`) lane that wrote the two sections above — MY OWN CHARACTERISATION IS REFUTED, AND I RE-MADE AN ERROR THIS FILE HAD ALREADY CORRECTED DIRECTLY ABOVE ME

**What I wrote, twice, and what is actually true.** My CONTRIB and take-on above say the
refusals are *"the signature of a fresh `hero-tool` render producing a fixed, thinner output
regardless of what it is replacing"*, and that the next step is to **"determine what empties
`hero-tool`'s `content_data` values"** — a writer census. **There is nothing to census. Nothing
empties them.** The premise is false, and the census would have burned a session finding no
such writer.

**The worst part is that the refutation was already in this file, immediately above my
contribution.** The `bugs_open/305` lane corrected exactly this generalisation on **2026-08-22**
— *"15, 5, 5, 5, 12 … per-run variance in generated output, not a settled renderer change"* —
and named its own failure mode as letting three agreeing samples outvote one disagreeing one.
**I then wrote the same generalisation from five agreeing samples the next day**, in the same
file, below the correction. Reading a file is not the same as reading the correction in it.
Logged in `WRONG_CALLS.md`.

### The census that settles it [MEASURED 2026-08-23 ~17:45Z, loanzy.uk, 11 pages carrying `hero-tool`]

**1. The empties are confined to ONE optional block, with no exceptions.** Of **40** empty
values across all 11 pages, **40 are `stat_*` keys**. Not one of the other five keys is ever
empty on any page. Every page stores exactly **11** keys.

**2. They empty in label/value PAIRS**, so the non-empty count moves in twos and maps exactly
onto "how many of three optional stat slots are filled":

| non-empty (of 11) | = 5 fixed + | stats filled | rendered `class=` |
|---|---|---|---|
| 11 | 6 | **3** | 15 |
| 9 | 4 | **2** | 12 |
| 7 | 2 | **1** | 9 |
| 5 | 0 | **0** | 5 |

**3. And the count moves in BOTH DIRECTIONS across successive writes** — which is what kills
the emptying-writer theory outright. From `page_component_history` (archived pre-write state)
joined to the current row:

| page | stats filled, in write order |
|---|---|
| `tool-loan-comparison-calculator` | 0 → **3** |
| `tool-loan-repayment-calculator` | 0 → 0 → **3** |
| `tool-settlement-calculator` | 0 → **2** |
| `tool-overpayment-calculator` | 1 → **0** |

Three up, one down. **A mechanism that empties values cannot fill them.** What this is: the
generator choosing, run to run, how many of three optional stat slots it populates — the
`on_missing: skip_field` gates then drop ~3 class attributes per unfilled stat. Per-run
variance, exactly as the 305 lane said on 08-22.

### What this means for your guard — less than my earlier note implied

- **The floor is not catching content loss.** It is catching a *fresh generation that filled
  fewer optional stats than the stored one did*. Both saves are legitimate outputs of the same
  writer; neither is damaged.
- **The refusals are therefore RETRYABLE, and they retry successfully.** Proven on the page I
  was blocked on: refused 14:03:06Z at 12→5, **saved cleanly at 14:23:29Z** with the stat count
  back at 2. I did not touch `section_component_floor`, and I did not need to.
- **So my "practical consequence" paragraph above was wrong too.** I wrote that
  `loanzy.uk/tools/credit-health-check` "cannot be repaired until this is resolved". It
  repaired itself 20 minutes later, on retry, while I was writing that it could not.
- **The one thing that stands from my earlier note is the diagnosis-vs-message gap**, and it is
  narrower than I framed it: the message says *"may not lose more than 50% of the elements
  carrying layout classes"*, which sends the reader hunting a renderer defect. For this
  component the honest reading is *"the incoming render filled fewer optional fields than the
  stored one"* — a **retry** signal, not an investigation signal.
- **I am still not proposing a code change**, and I am now proposing one *less* than before: the
  earlier note floated measuring non-empty field values instead of class attributes. On this
  evidence that would change what the guard refuses without making anything safer — it would
  refuse the same saves for a better-worded reason. **A shared-guard contract change affecting
  nine writers should not be spent on a message.**

### Why this is filed without a `090` run, stated plainly per the 2026-07-31 owner ruling

This is a **refutation** of a structural claim, not the assertion of a new root cause, and it
rests on a complete first-hand census rather than inference: all 40 empty values enumerated by
key, all 11 pages, all 22 available history rows read directly. **It could have come out
otherwise, and I expected it to** — a single empty non-`stat_*` key, or a monotone downward
trend, would have refuted me and sent me to the writer census I am here to cancel. It is also
independently corroborated by a separate lane (305) reaching the same conclusion from different
samples a day earlier. The disconfirming check is one query and it is in the tables above.

**Net effect on this bug: one open question is CANCELLED, none is added.** `hero-tool` needs no
content-loss investigation. 253's own subject — a framework rewrite stripping layout components —
is untouched by any of this and remains whatever it was.
