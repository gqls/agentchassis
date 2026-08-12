# PLAN — separate WHAT THE PAGE SAYS from HOW IT READS: a two-stage copy process

**Opened 2026-08-12** at the owner's direction, out of the loanandmortgagecalculator
homepage rewrite. **This lane exists because AI-slop copy is a product defect, not a
polish item** — the owner's words: *"AI slop writing is extremely damaging for any
site."*

**Owner's proposal, verbatim:** *"Maybe we have a two stage process - the first is
writing the copy with the facts, and then another pass to look through it, perhaps
using the offer analysis loop and reorder and rewrite the copy using the same facts
but in a readable form. Two separate stages might help us focus better on the
problem."*

**Scope note from the owner:** this belongs in its own thread rather than inside
`loanandmortgagecalculator_couk`, *"although this is a good (not busy) site to test
on."* So: design and mechanism here, test fixtures on LMC, and the offer-ordering
half is a hand-in to `vigilant_designer_offer_analysis` (the owner already put
homepage messaging in their remit on 2026-08-11).

---

## 1. The evidence this is a process defect, not a model defect

Three rounds on one page, same model (`claude-sonnet-5`), same prompt, same
owner-approved 46KB site voice spec. What changed each time was the BRIEF.

| round | brief said | what came out |
|---|---|---|
| 1 | nothing (no `content_direction` on the page) | 235-word essay, no cards, no headings — refused by the shrink guard |
| 2 | structure: h1, sections, cards as h3+blurb+link | right structure, **zero design classes** (`bugs_open/253`) |
| 3 | structure **+ the design vocabulary** | design restored, and the copy the owner then rejected |
| 4 | framing + readability + "do not describe the site" | in flight at time of writing |

**The round-3 copy the owner rejected was my brief executed faithfully.** Compare:

> **My brief:** *"This site is loanandmortgagecalculator.co.uk: 23 free UK
> calculators covering loans AND mortgages together, because the two kinds of
> borrowing interact (a car loan changes what a mortgage lender offers; a
> remortgage changes what other debt costs). No sign-up, no credit check,
> everything runs in the browser and nothing is sent anywhere."*
>
> **What it wrote:** *"This site holds 23 calculators covering both sides: 12 for
> mortgages, 11 for loans and credit. Everything runs in your browser… There's no
> sign-up, no credit check, and nothing you enter is sent anywhere."* and *"Take on
> a car loan and a mortgage lender will usually offer you less."*

The owner's objections — *"they don't need to know that, we don't want to talk about
ourselves"* and *"the whole thing is based on negativity"* — land on **content I
specified**, in **an order I chose**. The writer's job was to render it; it did.

**That is the argument for two stages.** Whoever writes the brief embeds their own
framing and priority order, and cannot see it, because to them it is just the facts.
A second pass with **no stake in the brief** is the only place that framing becomes
visible. The same claim, differently: `CONTRIB_2026-08-08` in the offer lane already
named this — *"a true, well-evidenced claim can still be the wrong thing to lead
with."*

**Corollary that constrains the design:** stage 2 must not be handed the brief. If
it reads the brief it inherits the brief's framing, which is the failure being
fixed.

## 2. What already exists (measured 2026-08-12, do not rebuild these)

> **⚠ CORRECTED 2026-08-12 (same day, later) — FOUR ROWS OF THIS TABLE ARE WRONG, and
> the section's headline conclusion with them.** A prior-art sweep of the docs tree and
> the code found a whole earlier lane (`fleet_copy_quality`, 12 files, 6–9 Aug) and two
> shipped Go mechanisms (`ScanVoiceTells` CQ-020, `ScanDeployedClaims` CQ-021) that this
> section did not know about. Full evidence and queries in NOTES under *"the prior-art
> sweep"*. In short:
>
> - **"anything that CONSUMES that judgement → NO" is FALSE.** `content-quality-auditor`
>   has run **34 times, all COMPLETED, most recently today**, and its runs' own
>   `collected_data` shows findings being classified and items created. My query named
>   a producer string (`design-audit-agent`) that has never existed — the real one is
>   `design-audit`.
> - **"stage 2 — the editorial rewrite → NO" is FALSE.** `content_rewrite` holds **83
>   complete** items; 32 of them are `voiceh-rollout` rewriting a site's pages for voice
>   on 8–9 August.
> - **"the precedent for the missing piece"** — the copy equivalent of `css-patch-agent`
>   already exists and is named `page-build-handler`
>   (`write_audit_findings_action.go:375-386`).
> - **The real defect is ATTRIBUTION, and it is much smaller than a missing mechanism.**
>   The auditor's step sets `audit_source: "content-quality-audit"`; no work item in all
>   history carries that value, because `audit_source` is Optional with
>   `Defaults: {"audit_source": "design-audit"}` (`:43-44`). The copy auditor's output is
>   real, consumed, and **invisible as copy work to any query anyone would write.**
>
> **So the gap is NOT "the audit half runs and dies".** It is narrower and stranger: the
> findings that reach items are gap/CTA/differentiation-shaped — *what is missing* — while
> the readability-and-register axis has no producer with reach (`voice_tells` is opt-in and
> only leopardess has a gate; LMC has no `voice` spec at all) and, by deliberate design, no
> permitted applier. **§3 and §4 below are left as originally written and are superseded in
> the places NOTES names; they are kept because the reasoning in §1 and §3's rule 1 survives
> intact and is this lane's real contribution.**

| piece | exists? | state |
|---|---|---|
| stage 1 — write the facts | **yes** | `page-content-writer`, driven by `needs_page` |
| a copy-quality JUDGEMENT | **yes** | `content-quality-auditor` — tone, gaps, CTA, differentiation, one LLM call |
| anything that CONSUMES that judgement | **NO** | `SELECT item_type,status,count(*) FROM site_work_items WHERE created_by IN ('content-quality-auditor','design-audit-agent')` → **0 rows, all-history** |
| stage 2 — the editorial rewrite | **NO** | nothing rewrites prose for readability/order |
| a safe way to APPLY a copy edit | **yes** | `section-editor` — updates `content_data` then re-renders **one component**, so the page's markup and its neighbours survive (this is also the route that dodges `bugs_open/253`) |
| the precedent for the missing piece | **yes** | `css-patch-agent` exists precisely to apply **design** audit findings. There is no equivalent for **copy** findings. |

**So the gap is not "we need an AI to write better". It is that the audit half runs
and dies, and the apply half has no agent.** That asymmetry — design findings get an
applier, copy findings do not — is the whole of the missing mechanism.

## 3. The design

```
stage 1  needs_page → page-content-writer
         judged on: facts, coverage, structure, links, design classes.
         NOT judged on: prose quality, order, register.
              │
              ▼  (component written, page renders)
gate     content-quality-auditor  ─── raises ──▶  site_work_items
         judged on: does this read like a person; is the most useful
         thing first; does it talk about the site instead of the reader
              │                                   item_type='copy_needs_editing'
              ▼
stage 2  copy-editor agent  ── executes via ──▶  section-editor
         INPUT:  the rendered component + the site voice spec
                 + the reader-intent/benefit ORDER (offer lane)
         FORBIDDEN INPUT: the stage-1 brief
         CONTRACT: same facts, same links, same design classes, same
                   component boundaries. Reorder and rewrite only.
```

### The three rules that make stage 2 safe rather than another rewrite

1. **Same facts.** Stage 2 may reorder, merge, cut a sentence and rewrite any
   sentence. It may not introduce a fact, a number, a claim or a link. A fact
   inventory diff (numbers + links + claims in vs out) is the acceptance test, and
   it is mechanical.
2. **Same markup.** Class attributes and component boundaries in == out. Counted,
   not eyeballed — `bugs_open/253` is what happens when a text-level check governs a
   markup-level change, and my own comparison missed it the same way.
3. **One component at a time, through `section-editor`.** Never a page-level
   regenerate. This is what keeps a locked calculator row and its neighbours out of
   scope entirely, which matters because Track B is about to produce 22 pages shaped
   `[prose, LOCKED tool, prose]`.

### What the offer lane owns

The ordering judgement — *what is the most useful thing to this reader, and is it
first* — is their B4 question (*"what they are trying to achieve by visiting this
site"*), not a writing question. Stage 2 needs that as an INPUT it does not compute
itself. Hand-in written as `CONTRIB_2026-08-12` in their directory.

## 4. Phasing

> **⚠ SUPERSEDED 2026-08-12 by §6 below.** P2 asks for wiring that already exists (see
> the correction in §2). The revised phasing is §6; this list is kept as the record of
> what we believed before the prior-art sweep.

- **P0 (done, this session):** the LMC page brief rewritten against the owner's
  critique; round 4 run as the first fixture. Records whether a better BRIEF alone
  is sufficient — because if it is, stage 2 is a smaller job than it looks.
- **P1:** grade the four rounds against the owner's own critique, by hand, and turn
  his three objections into checkable rules (heading-not-a-negation;
  no-site-inventory-in-the-opening; no-stacked-consequences). Fixtures that came
  from real runs and a real owner rejection — **not composed by us**, which is the
  property the offer lane's HANDOFF says its fixtures lack.
- **P2:** wire the existing auditor's finding to an item (the `css-patch-agent`
  shape). Cheap, and it makes the gap visible in the queue instead of in a chat.
- **P3:** build the stage-2 copy-editor as config (agent definition + workflow),
  applying via `section-editor`. Seed migration + council gate if it touches shared
  Go, which it should not need to.
- **P4:** the fact-inventory and markup-parity acceptance checks, induced before
  they are trusted.

## 5. Open questions for the owner

1. **Does stage 2 run on every page, or only on a flagged one?** Every page is
   honest but expensive (one extra LLM call per page per build); flagged-only means
   trusting the auditor's judgement to notice slop, which is unmeasured.
2. **Who arbitrates when stage 2 disagrees with the site voice spec?** The spec is
   owner-approved; the editorial pass is a judgement. Today the page-level brief
   silently wins over the site spec because it renders later in the prompt with
   *"follow closely"* — that precedence is undocumented and is how a page brief
   overrode an owner-approved voice on 2026-08-11.
3. **Is "readable" gradeable at all, or does it stay a human call?** P1's fixtures
   are the cheapest way to find out, and the answer decides whether P3 has an
   acceptance test or only a reviewer.

---

## 6. REVISED PHASING (2026-08-12, after the prior-art sweep)

Ordered by what closes the door, not by what is interesting. **Phase 1 builds nothing** —
it delivers decisions already made and turns on mechanisms already shipped. On today's
evidence it is most of the available value, and it has to happen first or Phase 2 is
designed against a system nobody has actually seen working.

### Phase 1 — deliver what is already decided (no new mechanism)

**1a. Ship Voice H into the writer prompts. This is an undelivered owner decision, and
it is the highest-value item in the lane.** `[MEASURED 2026-08-12]` all seven writer
prompts still carry the old prescription *"Start with the fact"*; none carries H's
prohibition. The owner ruled on 2026-08-09 that H becomes the fleet default and that its
prohibition replaces that prescription. `page-content-writer` — the agent that wrote the
copy the owner rejected on 08-11 — is one of the seven.

Two things the fleet lane left open must be settled inside this, not around it:

- **How it ships.** Seven separate edits, or one shared carrier read at prompt-assembly
  time. The seven have already drifted from a common ancestor without anyone intending
  it, so seven edits will drift again. This is a revision of a live fleet-wide default,
  which makes it a council submission with both options written up, a concept-register
  entry in the same commit, and the other consumers **told rather than counted**
  (owner ruling 2026-07-29 §3).
- **The exemplars must change with the rule.** The fleet lane's clearest experimental
  result: they deleted a rule, left its three worked examples, and the behaviour did not
  move. *"The example is the instruction; the rule is commentary."* A submission that
  edits rule text and leaves the old exemplars in place is theatre, and we have direct
  evidence of that rather than a suspicion.

**1b. Fix the `audit_source` attribution defect.** The copy auditor's findings are
stamped `design-audit` because the configured literal never lands and the action's
default fires (`write_audit_findings_action.go:43-44`). Small, provable, and until it is
fixed **no query can distinguish a copy finding from a design one** — which is how this
lane came to believe the audit half was dead. File as a bug; it is a `bugs_open/` case,
not a mechanism.

**1c. Opt `loanandmortgagecalculator.co.uk` into `evidence_base`.** One `site_specs`
row turns on `ScanDeployedClaims` — shipped, live in `v1.0.1283`, council-approved, with
a claim-granular revalidation gate. NOTES records that the round-4 rewrite introduced a
new figure to the homepage and **the only thing that checked it was a hand-written
query**. This closes that hole today, for free, on a mechanism we already own. The same
question should then be asked of every finance site: 12 of ~29 sites carry the spec;
LMC, loancalculator and cookly do not.

### Phase 2 — the actual gap: a judgement with PAGE scope

The two-stage idea survives the sweep, but §3 rule 3 does not. **A section-scoped editor
cannot do stage 2's job**, and the fleet lane has the proof rather than the worry: the
arm test wrote "Amortisation" in one section and "amortization" in another, because *"each
section is written separately, so a rule about the whole page has nothing holding it
together between them."* Stage 2's whole remit — is the most useful thing first, does this
talk about the site instead of the reader, is it one name per thing — is page-level.

So the shape is **page-scoped READ, section-scoped WRITE**: stage 2 reads the entire
rendered page plus the site voice spec plus the offer lane's ordering input, is **denied
the stage-1 brief** (§1's corollary, which survives), and emits per-component edits
applied through `section-editor` one at a time. That split is the new thing this lane
would build, and it is the only part not already present somewhere in the estate.

**Before any of it is built, two constraints must be resolved by a human:**

- **An unsupervised copy rewriter changes a deliberate guarantee.** `voice_tells` was
  designed HITL-terminal (`HandlerAgent: ""`; the spec defers auto-rewrite; *"never an
  unreviewed auto-rewrite"*), and `bugs_open/033` cites that text as evidence it was
  filed correctly. Under the owner ruling of 2026-07-29 §1 this is architecture-scope —
  it changes what a shared mechanism guarantees — so it goes to an RFC or an owner
  ruling, not a seed migration.
- **Locks, not instructions, are what protect approved copy.** In the 08-09 arm test
  *both* prompt versions tried to overwrite the owner's personally-approved opening and
  were stopped only by the lock — *"not the instructions, not the care taken writing
  them."* Stage 2 must treat a locked component as out of scope structurally, in the
  selection, not by being told to leave it alone.

### Phase 3 — the ordering input (hand-in, already written)

The table-stakes/differentiator axis from `vigilant_designer_offer_analysis`
(`CONTRIB_2026-08-08`). Stage 2 consumes this; it does not compute it. Nothing in the
platform currently represents what a reader came for, and that is the axis the owner has
now judged on three times (loancalculator 08-08, mortgagecalculator 08-11, LMC 08-11).

### Phase 4 — acceptance checks, induced before they are trusted

Fact-inventory diff (numbers, links, claims in vs out) and markup parity (class attrs and
component boundaries in == out). **Induce a non-zero on both before believing a zero** —
a markup-parity check that cannot fail is what `bugs_open/253` looks like from the
inside, and the original §3 rule 2 was written after my own comparison missed it the same
way.

## 7. REVISED OPEN QUESTIONS FOR THE OWNER

The three in §5 stand. These come from the sweep and are sharper:

1. **Voice H shipped three days late — do we ship it now as seven edits, or build the
   shared carrier?** Seven is faster today and drifts again; the carrier is the structural
   fix and is a bigger council round. (Recommendation: the carrier, because the drift is
   already measured and this is the second time it has cost us.)
2. **Does an editorial pass get to change live copy without a human reading it?** The
   estate's current answer is a deliberate no, written into `voice_tells`. Stage 2 is only
   worth building if that answer changes, or if its output is queued for review — in which
   case it competes with `bugs_open/033`, the human-review queue that has no working
   surface.
3. **Do we opt the rest of the fleet into `evidence_base`, or only the finance sites?**
   17 sites in the pool carry no voice spec and no evidence base at all.

---

## 8. OWNER DECISIONS, 2026-08-12 (answering §7)

Taken in chat after the prior-art sweep was presented. All three are the recommended
option; recording the reasoning so a later thread does not reopen them cheaply.

**D1. Voice H ships as a SHARED CARRIER, not seven edits.** One place all seven writer
prompts read at prompt-assembly time. The argument that won it is measured, not
aesthetic: the seven have already drifted apart from a common ancestor without anyone
intending it, and this is the second time that drift has cost us. Seven edits would be
live sooner and would guarantee a third occasion.

- **This is a revision of a live fleet-wide default, so it goes through the council gate
  as one** (`fleet_copy_quality`'s own conclusion, 2026-08-09), with both options written
  up, a concept-register entry **in the same commit** (owner ruling 2026-07-28 condition
  2, which still stands), and the other consumers **told, not merely counted** (owner
  ruling 2026-07-29 §3). Six writer agents besides `page-content-writer` are consumers.
- **The exemplars change with the rule, or the change is theatre.** Not a style
  preference — the prior lane deleted a rule, left its three worked examples in place,
  and the behaviour did not move. *"The example is the instruction; the rule is
  commentary."*

**D2. NO unreviewed auto-rewrite. Stage 2's output queues for human review.** This
preserves the guarantee already written into `voice_tells` (`HandlerAgent: ""`, *"never
an unreviewed auto-rewrite"*), which means **stage 2 is no longer an architecture-scope
change** — it adds a producer to an existing HITL-terminal shape rather than changing
what a shared mechanism guarantees. That materially lowers what it costs to build.

> **⚠ THE KNOWN CONSEQUENCE, ACCEPTED WITH THE DECISION AND NOT DISCOVERED AFTER IT.**
> `bugs_open/033` — *the human review queue has no working surface*. Today a
> `needs_human_review` row is where work goes to park: `voice_tells` has **34 parked and
> 1 ever closed**, and that one closed by machine revalidation rather than by anyone
> reading it. **So D2 as stated routes stage 2's output into a queue nobody can read.**
> Stage 2 is not worth building until that surface exists, and `bugs_open/033` therefore
> moves from "a bug someone should fix" to **this lane's blocking dependency**. It is
> named here so the next thread meets it in the plan rather than in the backlog.

**D3. Review the three consumer-finance sites carrying `strategy.tone: "authoritative"`
— `loancalculator.co.uk`, `loancash.co.uk`, `lendzy.co.uk` — and no others.** Read them
before changing anything. The other five authoritative sites (`finetuning.uk`,
`fundamentallyai.com`, `gaswholesalers.com`, `oufe.com`, `vetcomparison.uk`) are left
alone: the mechanism argued for is specifically that "authoritative" turns adversarial
where **the reader is the weaker party**, and that condition does not hold on a B2B or
technical site. Reframing all eight would have edited eight specs on a theory
demonstrated on one site.

### Revised order of work, after these decisions

1. **D1 — the shared carrier** (council submission + register entry + exemplars). The
   only item that reaches every future page on every site.
2. **The `audit_source` attribution fix** (§6 1b) — file as a bug, small and provable.
3. **`evidence_base` for LMC** (§6 1c) — one row, turns on a shipped claims gate.
4. **D3 — read the three finance sites.** Diagnosis, not a sweep.
5. **`bugs_open/033`** — now blocking, per D2.
6. **Stage 2**, page-scoped read / section-scoped write, only after 5.

---

## 9. CONSTRAINT ON STAGE 2 FROM `bugs_open/260` (verified in code, 2026-08-12)

Handed in by the `brochure_component_library` front as `CONTRIB_2026-08-12b`; verified
independently and sized in `CONTRIB_2026-08-12c`. **This changes what §6 Phase 2 may
build, and it is not a risk note — it is a gate.**

**The mechanism.** A component declares a field as an array. The writer emits a prose
string instead. `range` over a string errors, the renderer silently falls back to a regex
renderer for a different dialect, and the component renders with `{{if}}`/`{{range}}`/
`{{end}}` verbatim in the page. The whole component is destroyed — every other field in it,
correctly written, goes with it.

**Why it lands here.** §6 Phase 2's executor is `section-editor`, and
`ApplySectionEditAction` has **three pre-persist guards, none of which is a type check**:
the link repair rewrites rather than refuses, the envelope normaliser refuses on envelope
*shape* only, and the artefact classifier is explicitly advisory. The single render-side
check is `if rendered == ""`, which non-empty template gibberish passes. **And this path
writes to pages that are already live**, where stage 1 is protected by `validate_content`
refusing before it persists.

**A readability pass is the likeliest thing in the estate to flatten a nested array into
prose.** That is literally the trigger in 260. So stage 2 is not merely exposed to this
hazard — it is the best candidate to cause it.

**D2's human review cannot catch it.** The failure is at render, downstream of the words a
reviewer reads. Approved copy and a destroyed component are the same artefact at review
time.

### The measured exposure `[MEASURED 2026-08-12]`

- **45 of 191 active components declare at least one `array` field** (49 fields), and
  **12 of those fields are `source: llm`** — authored by the writer, checked by nothing.
- The path has **132 completed `section-editor` runs and 0 of 1,454 stored components
  carrying directives**, so the hazard has not yet fired here — a real negative, since the
  path is heavily used, not idle.
- ⚠ **The library is not JSON Schema.** 4 of 191 components use `properties`; **140 use
  the house dialect `{"fields": {<name>: {"type": …}}}`**; 47 declare nothing at all. A
  gate written against `properties` would cover **4 components and report a clean sweep
  over the other 187.**

### What Phase 2 must therefore do

1. **Prefer `field_updates` over `replacement_content_data`.** `applyContentEdit` seeds its
   map from the existing row, so a field-scoped update preserves the types of every field
   it does not touch. Only a full replacement can retype one. If stage 2 can express its
   edit as field updates, the hazard is structurally out of reach — which is better than
   any gate.
2. **If a full replacement is unavoidable, a type gate stands in front of the write** — and
   it reads the **house dialect**, not JSON Schema, or it is inert on 187 of 191 components.
3. **The 47 schema-less components are not covered by either.** A green gate is not fleet
   coverage, and Phase 4 must say so where the result is reported.

**Phase 4 gains this as a third acceptance check**, and it is the same shape as the other
two: assert against the component's **own declaration**, which — like the required-link set
— cannot drift from the brief.

### Adopted from 12b §2, because it forecloses a round

The writer in 260 was given a **formal JSON Schema** (`type: array`,
`items: {type: object, required: [body]}`) and answered with a sentence. So *"prose
instructions failed, now hand the writer the schema"* has already been run on a live
component and failed. **Do not spend a round on it.** The check belongs at the boundary,
never in the prompt — which is the same conclusion this lane reached about set preservation
from an entirely different direction.
