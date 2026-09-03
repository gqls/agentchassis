# HANDOFF — vigilant designer + offer/benefit analyser (2026-09-03)

**COLD-START = this file, then `HANDOFF_2026-09-02_continue_here.md` (still the authority on this
lane's longer history), then register `CQ-036` / `CQ-034` / `IMG-074` / `WII-033` / `CLM-024`.**

> **Re-run every number here before acting.** This branch takes hundreds of commits a day. Every
> figure below carries the date it was taken; a count does not go wrong, it goes **stale by
> addition**, and it reads as current for ever.

---

## §0 — THE FOUR THINGS THAT CHANGED, AND ONE IS A TRAP

**1. A FRESH CHASSIS ROLLED: `7bf1ff674021f2d57dfd0aa41324541070646c3a`** (`[VERIFIED 2026-09-03
09:49Z]` from `service_binary_capabilities`, with a control — HEAD is correctly **not** an ancestor).

**2. ⚠⚠ `BANNED_REGISTER` v2 IS NOW LIVE, SO THE 23% BASELINE IS NO LONGER COMPARABLE.** `0c11a8818`
is an ancestor of the deployed build. v2 widened `x_not_y` (em/en dash separators) and added a
`plain_words` class: **8 patterns → 9**. **Any post-roll rate measured with v2 and compared against
the pre-roll 23% is measuring the instrument, not the producer.** Measured both ways so the
discontinuity is visible rather than inferred:

| day | minted | dirty by **v1** | dirty by **v2** |
|---|---|---|---|
| 08-26 | 249 | 53 (21%) | 65 (26%) |
| 08-27 | 342 | 90 (26%) | 102 (29%) |
| 08-31 | 137 | 39 (28%) | 50 (36%) |
| 09-01 | 153 | 41 (26%) | 56 (36%) |
| **09-02** (gate wired 17:14Z) | 149 | **24 (16%)** | 34 (22%) |
| **09-03** (first full gated day) | 76 | **9 (11%)** | 15 (19%) |

**Compare v1-to-v1 only.** On that basis: baseline **23%** → **11%** today, n=76, which is a real
drop (≈p 0.009). **Do not quote the v2 column against the v1 baseline.**

**3. THE PRODUCER GATE IS FIRING ON EVERYTHING.** `[MEASURED 2026-09-03]` **13 of 13**
`offer_ordering` rows written today carry the gate's own `register_repairs` record, and **11 of 13**
performed at least one repair. ⚠ **The residual 11% is NOT the gate failing to run** — it never drops
a point and never fails a run, so an unrepairable point keeps its original text **and is recorded**.
**The record is the instrument; read it, not the absence of dirt.**

**4. MIGRATION `723` IS APPLIED AND THE COUNCIL RETURNED `REVISE`.** See §2 — that is the first thing
you owe, and one objection is a real latent defect.

## §1 — WHAT IS DONE. Do not re-take any of this

| | state |
|---|---|
| **Decision D — the question hierarchy** | **SHIPPED**, migration `723`, applied + ledger-recorded 2026-09-02. Owner authorised **directly** (this lane declined the same ruling twice by relay, correctly) |
| register **CQ-036** + index row | live, with the council response recorded in it |
| the producer register gate (`CQ-034`) | **live and firing** — see §0.3 |
| `BANNED_REGISTER` v2 | live in the rolled binary; lockstep suite green at `0c11a8818` |
| `IMG-074` (planner imagery vocabulary) + LANDMINE | live since 08-26 |
| LANDMINES ×4 this week | the `site_assets.image`→hero alias trap · widening-vs-reshuffle · **config-wired vs call-wired guards** · **`name` vs `function` on 336 of 442 components** |
| `WRONG_CALLS` ×3 | incl. the joint two-lane row, and **my three false absences from my own greps** |
| vetcomparison design critique | delivered, corrected twice, reconciled — see §5 |

## §2 — ⚠ WHAT YOU OWE FIRST: the `723` REVISE

**Corr `fdd0f52c-256d-4e62-b02b-1bdab0ddc98a`, gating objection from `editquality`. Verdict read and
each objection answered in `CQ-036`.** The migration is **already applied**, so these land against a
live change. In priority order:

1. **⚠ `723` IS NOT IDEMPOTENT — DO NOT RE-APPLY IT.** (`debug_historian`, HIGH, **real, not fixed**.)
   The replacement text **re-embeds its own anchor** (`OUTPUT. Return ONE JSON object and nothing
   else`) at the tail of the inserted block, and `replace()` replaces **all** occurrences — so a
   second run **stacks a second copy of the guidance** instead of no-opping. `[VERIFIED 2026-09-03]`
   the applied row is correct (anchor ×1, guidance ×1, `spec_version 2`, 9,590 chars), **so the risk
   is a re-run, not the current state.** Ledger-recorded, so the runner will not re-run it; a hand
   re-apply would. **The correct form was a pre-state gate plus an occurrence COUNT** rather than
   `position(...) = 0` presence checks.
2. **`architecture` (MEDIUM) — CONCEDED and unresolved: no consumer is wired and no follow-up is
   named.** This is this lane's own recurring finding turned on itself. The consumption side is
   `copy_quality_two_stage`'s by the agreed seam split, and D-before-axis sequencing means there is
   nothing for their writers to read until the hierarchy actually runs.
3. **`llm_reliability` (MEDIUM) — ANSWERED, no action.** `[MEASURED 2026-09-03]` 388 `offer-analyser`
   calls over 10 days: max output **3,525** tokens, mean 2,222, cap **8,000**, **zero at or over**.
   ~4,475 headroom against a ~6-item hierarchy.
4. **`editquality`'s gating HIGH is factually wrong about the file — and the fault is mine.** It says
   the verify never checks `spec_version` or the `question_hierarchy` key; **both assertions are in
   the file.** My `edits[4]` sketch showed only part of the verify block. ⚠ **A sketch that omits
   assertions is read as a verify block that lacks them — the submission is the artefact under
   review, not the file.** Carry the whole block next time.
5. **`guardian` + `prior_art_librarian` (LOW, same point, both right):** load-bearing claims should be
   **re-checkable at review time**, not cited from a prior measurement. Each of mine was one query.

## §3 — WHAT TO DO NEXT, in the owner's confirmed order

1. **Watch the first hierarchy land.** Nothing changes until a site is re-analysed.
   ```sql
   SELECT s.domain, jsonb_array_length(ss.data->'question_hierarchy') AS questions,
          count(*) FILTER (WHERE (q->>'unanswered')::bool) AS unanswered
     FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
          LATERAL jsonb_array_elements(ss.data->'question_hierarchy') q
    WHERE ss.is_current AND ss.aspect='offer_ordering' GROUP BY 1,2 ORDER BY 1;
   ```
   ⚠ **A zero means "no site re-analysed since apply", NOT "the model ignored it"** — confirm the
   demand side first. ⚠⚠ **AND A MOSTLY-`unanswered` TOP IS THE PRE-REGISTERED RESULT, NOT A
   FAILURE** — it restates the 19-of-186 measurement per site. **A pass that answers everything means
   the model stretched points to cover questions**, which is the failure mode.

   > **✅ DONE 2026-09-03 — IT LANDED, AND THE PRE-REGISTERED CRITERION WAS REFUTED. Full read and
   > every figure: `NOTES_vigilant_designer_offer_analysis.md`, entry "DECISION D HAS LANDED".**
   > `[MEASURED 2026-09-03 12:38:58Z, pinned in one query]` **18 sites, 93 questions**, first row
   > 09-02 23:03Z. **It is a MOVING population** — 15 sites at 12:19Z, 18 at 12:36Z, and 19 of the 37
   > current `offer_ordering` rows are still un-re-analysed. Re-take before quoting.
   >
   > **The join is mechanically sound:** zero dangling `answered_by`, zero `unanswered`/`answered_by`
   > contradictions, and `from_field` is the **strategy** register (`trust_threshold` 28,
   > `satisfaction_condition` 22, …) not the offer points — so the questions are not
   > reverse-engineered from the answers.
   >
   > **But 86 of 93 (92%) came back ANSWERED**, which by the letter of the criterion above IS the
   > failure mode. **So I read pairs instead of quoting the count:** 36 of the 93 questions across 7
   > of the 18 sites, question text against the text of the point claimed to answer it.
   > **Verdict: the joins are mostly real** — on `garden-tools.uk` and `farmerinsurance.uk` all ten
   > pairs are tight and one-to-one. **The prediction was wrong, not the mechanism.**
   >
   > **⚠ THE FAILURE MODE IS REAL BUT NARROW, AND IT HAS A ONE-LINE SCREEN.** It shows up as **one
   > point claimed twice**, not as a site answering everything:
   > `count(answered_by) > count(DISTINCT answered_by)` per site. `[MEASURED 2026-09-03]` it fires on
   > **4 of 18 sites / 5 of 86 refs**, and **4 of those 5 reuses are stretches or half-answers**
   > (`idea.uk` offers a privacy point as the answer to "do I have to sign up?"), while **1 of 5 is
   > honest** (`relojistas.com` — two ways of asking about cadence). **Of the 21 single-use joins I
   > read, none was a stretch.** So it is a screen, not a verdict — but it turns a 93-pair editorial
   > job into a 5-pair one.
   >
   > **⚠⚠ THE FINDING THAT MATTERS: THE ABSENCE IS *PRICE*, NOT "EFFORT/PRACTICALITY".**
   > `[MEASURED 2026-09-03 12:38:58Z]` **5 of 7 `money_flow` questions are unanswered. Every other
   > source field is a clean ZERO** (`trust_threshold` 0/28, `satisfaction_condition` 0/22,
   > `recurring_value` 0/15, `value_proposition` 0/9, `competitive_position` 0/8). The register
   > answers trust, differentiation and satisfaction on every site and does not answer *what this
   > costs me*. **And the model ranks the price question LAST** — mean rank **4.6** against 2.3–2.9
   > for trust and satisfaction. For `idea.uk`, a £29 product whose competitor is free, *"why would I
   > pay £29 when AI is free"* at rank 5 is **an ordering judgement to reject**, and ordering is
   > `copy_quality_two_stage`'s half of the seam. **That is the first thing the hierarchy has produced
   > that is worth sending them, and it is unsent.**
   >
   > This also sharpens H4 honestly: H4's "19 of 186 points address effort or practicality" was a
   > regex proxy, and the entry above it in NOTES records this lane nearly reporting a **widened**
   > regex as an improvement. `idea.uk` and `remortgagecalculator.uk`'s effort questions **were**
   > answered (by the stretched joins), so effort-in-general is not the gap. **Price is.**
2. **The two switched-off things — ruled, built, unstarted, and mine.**
   `[MEASURED 2026-09-02]` `info-card-grid`'s `carousel` flag is ON for **1 of 49** instances while
   the owner has **ruled it default-on**; and `Illustrated Text Block` is still chosen on **one site**
   post-`IMG-074`. Both are the cheapest impact wins on the table and are named to the owner as such.
   ⚠ For the carousel, the open design question is **where the default lives** (schema default +
   backfill of the 49, or resolution-time) — `carousel` is `source: static`, so **nothing derives it**
   and something must positively set it.
> **⚠ OWNER RULED BOTH FLIPS 2026-09-03 ("switch the switches"), relayed via `designblog.co.uk`, and
> the PRE-FLIGHT GATE IS ASSIGNED TO THIS LANE.** Both flips sit at #2 in the order the owner
> confirmed directly, so they are next — but **the effect of both is UNVERIFIED**, and the carried
> caveat (editorial design uplift, endorsed by designblog) is: **land them with a before/after read on
> served BYTES, not config.**
>
> **THE BEFORE-READ IS CAPTURED — `[MEASURED 2026-09-03]`, and it comes with its own controls.**
> `leopardessconsulting.co.uk/services.html` is the **positive control**: it is the ONE instance with
> `carousel: true` already set, so its served markup is what "on" looks like. `designblog.co.uk/index.html`
> is a flag-unset comparator.
>
> | served signature | flag **ON** (leopardess) | flag **unset** (designblog) |
> |---|---|---|
> | `carousel` | **19** | **0** |
> | `scroll-snap` | **6** | **0** |
> | `icg-` | **6** | **0** |
> | `prev` / `next` | **10 / 10** | **0 / 0** |
> | `overflow-x` | **2** | **2** | ~~must NOT change~~ **← WRONG, see below** |
>
> > **⚠⚠ CORRECTED 2026-09-03 — THE `overflow-x` NEGATIVE CONTROL IS FALSE, AND FOLLOWING IT WOULD
> > MAKE A CORRECT FLIP READ AS A DEFECT.** The struck text said *"it reads 2 on both, because it is
> > the wide-table styling … a flip that moves `overflow-x` is doing something other than what it
> > says."* Both halves are wrong.
> >
> > **The template's ONLY `overflow-x` is INSIDE the flag's own gated block.** `info-card-grid`'s
> > `html_template` line 204 sits between `{{if $.carousel}}<style>` (line 165) and its `{{end}}`
> > (line 298), alongside `--icg-track-gap` and `scroll-snap-type: x mandatory`. **So a correct flip
> > ADDS ONE `overflow-x` per flipped instance.**
> >
> > **The equal 2 was a coincidence of unrelated CSS on two different pages of two different sites** —
> > which is what made it look like a control. `[MEASURED 2026-09-03, at the served bytes]`:
> > `leopardess/services.html` = 1 from another component's `--trp-track-gap` + **1 from the
> > info-card-grid carousel block itself**; `designblog/index.html` = **2 from `.category-strip`**,
> > an unrelated component whose CSS is emitted twice. Neither number is the wide-table styling the
> > struck text names, and the two 2s have nothing in common.
> >
> > **THE CORRECTED ACCEPTANCE TEST — before/after on the SAME page, never across two pages:**
> > * **POSITIVE, must move 0 → n:** `data-hcc-track`, `data-hcc-prev`, `data-hcc-next`,
> >   `data-hcc-slide`, `info-card-grid__grid--carousel`, `scroll-snap-type`.
> > * **EXPECTED TO MOVE:** `overflow-x`, **+1 per flipped instance**. A flip that leaves it
> >   unchanged has NOT emitted the carousel stylesheet.
> > * **NEGATIVE, must NOT move:** the count of `info-card-grid__card` articles, and the card titles.
> >   This is a **layout** change; the content must be byte-stable. **A flip that changes the card
> >   count is the one doing something other than what it says.**
> >
> > What caught it: reading the template before running the test, because the flip was mine to make.
> > The general form is in `WRONG_CALLS.md` — **a count summed over a whole document is not a control
> > unless you know what each occurrence IS**, and two documents agreeing on a total is not agreement.
>
> **The rest of the acceptance test stands:** pick a flag-unset site, flip it, re-fetch, and confirm
> the four discriminating counts move from the right column to the left. ⚠ **Config alone is not
> evidence.**
>
> > **⚠ AND THE "must be positively set per instance" CLAUSE IS ALSO RETIRED — CORRECTED 2026-09-03.**
> > It said `carousel` is `source: static` "so nothing derives it and a flip must be positively set per
> > instance". **A `source: static` field with a declared `fallback` IS derived, at resolution time**:
> > `plan_sections_action.go:2886` reads `if !carryStored() && fallback != nil { resolvedData[field] =
> > fallback }`, and the re-render path runs the same `planSection`
> > (`rerender_page_sections_action.go:1450`). **So no per-instance backfill is needed** — that is what
> > migration `740` does, and the design question §3.2 left open ("schema default + backfill, or
> > resolution-time?") is answered: **resolution-time, with the mechanism that already exists.**
>
> ⚠ **AND IF THE ILLUSTRATED-BLOCK FIX IS A PLANNER-PROMPT CHANGE: migration `718` JUST EDITED THE SAME
> PROMPT'S IMAGERY BLOCK.** Use anchored replaces on **disjoint** anchors, per the 591/595/598/718
> discipline — and note this lane's own `723` idempotency defect (§2.1) is exactly what a careless
> anchored replace produces.
>
> **PRE-FLIGHT GATE (assigned):** the two queries live in `designblog.co.uk`'s RUNBOOK with the
> three-names landmine and the both-greps caveat baked in. **18 remakes are queued, so the payoff shape
> is wiring it to run per-remake before ship.** ⚠ Post-`721` counts measure a **repaired** population —
> date everything.
>
> > **⚡ STATE 2026-09-03 — THE CAROUSEL HALF IS WRITTEN, SUBMITTED AND *NOT YET APPLIED*.**
> > `docs/agent_docs/sql_for_agents/740_info_card_grid_carousel_defaults_on.sql` (+ `_ROLLBACK`),
> > committed `e14fff5a0`, council corr **`2ac895f3-ca82-4dbe-8f4e-3335a04b8925`** — verdict pending,
> > `Council-Submitted:` trailer, so `098` credits it automatically on approval.
> > **WHAT THE NEXT SESSION OWES: read the verdict, then APPLY it, then re-render one page and run the
> > CORRECTED acceptance test above.**
> >
> > **Re-take the census first** — `[MEASURED 2026-09-03 12:41:06Z]` it is **40** live instances across
> > **21** sites, not the 49 in the paragraph above (that figure was taken 09-02 on a different filter;
> > the pairing here is `build_status='deployed' AND status='active'`, both arms, per the standing
> > landmine). **1 stores `carousel: true`, ZERO store `false`, 39 carry no key.**
> > ⚠ **That zero is load-bearing:** `IsEmptyContentValue`'s default arm makes a bool `false`
> > NON-empty, so a stored `false` is carried and **beats** the fallback. Had the 39 stored an explicit
> > false, `740` would apply cleanly, verify green, and change nothing.
> >
> > **The pre-flight gate for THIS flip is measured and GREEN, at the artefact:** `[MEASURED
> > 2026-09-03]` all **21/21** carrier sites serve `/assets/js/snippets.js` at HTTP 200 with **15**
> > `data-hcc` occurrences, so the arrows will not be inert. **NEGATIVE CONTROL** (a constant 15 could
> > mean the grep matches something always present): **6/6 non-carrier sites serve 0**, one of them
> > `fundamentallyai.com` with a **10,928-byte** bundle — so it is not "small bundle = zero". The
> > bundle follows the COMPONENT, not the flag: `js_snippets` has one active row whose `applies_to`
> > is `["hero-card-carousel", "info-card-grid"]`.
> >
> > **The Illustrated Text Block half is NOT started.**

3. **boxingonline cards** — design against image + headline + deck; category/date/read-time collapse
   by default after migration 682. ⚠ **Do not add a short display-headline field**: `nav_label` is
   empty or unusable on ~5 of 6 pages fleet-wide.
4. **vonc** — cut the click (the home CTA says "File Your Position" and the landing page will not let
   you file one until you click again), and the theme pass **only in a deliberate window**.
5. **The logo rule** — "the background shouldn't be part of the logo" must reach the imagery prompt.

## §4 — WATCH-OUTS (new first)

- **⚠ `reference_values` IS NO LONGER A PIN.** Owner ruling 2026-09-02: the classifier has **full
  authority to ignore our themes**. `RFC_059` proposed a structural pin and **he withdrew it on that
  ground**. **Palette churn is authorised behaviour, not a defect to arrest** — any plan to
  de-garish vonc by pinning colours is against the ruling. ⚠ **And measure at the SERVED stylesheet**:
  `loanzy.uk` serves colours matching **neither** its palette row **nor** its `reference_values`.
- **⚠ `content_components.name` ≠ `.function` on 336 of 442 active components, and SIX pairs are the
  same words REVERSED** (`contact-hero` / `hero-contact`, 25 live instances). **A census on the wrong
  column returns a confident ZERO for a component that exists** — and a zero licenses building it.
  `page_components.slot_name` is a **third** spelling agreeing with neither. LANDMINE has the query.
- **⚠ MY OWN REPEATED FAILURE MODE, three times in a week — read this before trusting one of my
  absence claims.** (a) `grep | head -8` read as an absence (real count 55) → wrote a duplicate of an
  existing file; (b) a regex requiring `var(--x)` to close immediately, so **every** usage carrying a
  fallback was invisible → "the accent is never applied" when it is applied six times, one visibly;
  (c) **no instrument at all** — assumed three parked contrast defects were the accent's because a
  neighbouring one was. **Count before you read; match the PREFIX not the closed token; a defect's
  neighbour is not its family.** All three were caught by other lanes, not by me.
- **⚠ A GATE THAT RUNS AND WHOSE OUTPUT IS DISCARDED IS A FALSE GREEN.** `723`'s verify asserts the
  chain end-to-end for this reason. The sibling shape: migration `667` washed three points into a row
  the producer superseded **55 seconds later** — guards passed, NOTICE printed, ledger stamped, live
  table unchanged.
- **⚠ `offer-analyser` HAS NO ROOT `ai_service` BLOCK.** A step without a step-level block is live,
  fires, and **repairs nothing while recording why**. It happened for real between migrations `681`
  and `682`.
- **⚠ The `ordering` object is DEEP-MERGED**, so a key the model omits leaves the previous run's value
  standing and **looking current** (`bugs_open/327`). Every key is required on every run.

## §5 — CROSS-LANE STATE

- **`copy_quality_two_stage`** — the close partner. Seam split **agreed both ways**: production
  (`question_hierarchy` + `answered_by`, through the gate) is **ours**; writers and ordering are
  **theirs**. Register v2 landed `0c11a8818` and is live. ⚠ Two numbers **never to quote loosely**:
  the 10%→30% effort move is **not** an improvement (**the regex was widened**), and the post-gate
  1-of-12 is **not** evidence (`P(≤1 of 12 | unchanged) = 0.193`; ~25 points needed to detect a
  halving).
- **`vetcomparison` + `site design planner`** — critique delivered and **corrected twice by them**.
  Final state: the accent is **vestigial, not absent** (one static `::before`, two hover states, three
  dead rules); the three parked `contrast_failure` items are **grey body text at 4.14–4.33:1, real and
  separate** — ⚠ **my bundling them with the accent was an over-reach and the sharper of my two
  errors**; only `d6da17b4` is superseded; **the primary at 4.94:1 is the sharpest finding** (static,
  load-bearing, 0.44 headroom).
- **`designblog.co.uk`** — owner's fleet-sameness directive. The visual auditor **exists, is live, and
  scores coherence not impact**: zero mentions of imagery/infographic/visual-impact/distinct, never
  sees the rendered page, single-site by construction. Cohort sameness **is** cheaply measurable and
  the two queries are with them.
- **`agentchassis-ff` (dartsonline)** — per-section binding (`IMG-075`) shipped 09-01. ⚠ Their
  measurement: **all 22 content pages are `hero` + `article-body` + `call-to-action`, zero
  illustration-capable** — no page can host a per-section figure regardless of rows.
- **`inline_guide_imagery`** owns imagery placement; **`editorial design uplift`** now owns the
  **imagery supply** gap (two lanes routed it there).
- **`bugs_open/395`'s lane has CLOSED.** `pageFieldWriters` is total and live; `title` /
  `content-gap-planner` now answers TRUE. ⚠ **Read the `CLM-024` `doc_notes` row before wiring the
  emit-side stamp** — that stamp is still unbuilt.

## §6 — RESIDUALS, stated plainly

1. **`723`'s idempotency defect is unfixed** (§2.1). Latent; do not re-apply.
2. **`question_hierarchy` has no consumer and no named follow-up** — the `architecture` seat's
   concession, and the fourth instance of this lane's own pattern.
3. **The 23% baseline is retired by the v2 roll.** Future comparisons need a v2 baseline; §0.2 has
   both columns so one can be built without re-deriving.
4. **The pattern claim needs re-measuring before it is repeated.** "Five for five built-but-undriven"
   was pressed to the owner twice and is **three of five** — two were driven within days, and one is a
   third state: **built, approved, seeded, never executed**.
5. **Imagery supply** is the common root of three separate owner complaints and now has an owner
   elsewhere; **page structure is a second precondition** nobody owned as of 09-02.
6. **The parked-defect audit was offered twice and dropped** at the owner's implicit no. The
   vetcomparison case suggests those queues are worth less than their length.
