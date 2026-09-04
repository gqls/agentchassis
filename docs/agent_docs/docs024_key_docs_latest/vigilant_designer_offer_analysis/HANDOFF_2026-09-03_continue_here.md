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
   > **⚠⚠ ~~THE FINDING THAT MATTERS: THE ABSENCE IS *PRICE*~~ — REFUTED THE SAME DAY. I MEASURED
   > MY OWN PROMPT. Read the correction below before quoting any number in this block.**
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
   > answered (by the stretched joins), so effort-in-general is not the gap. ~~**Price is.**~~
   >
   > > **⚠⚠ CORRECTED 2026-09-03 evening — refuted by `copy_quality_two_stage` within the hour by
   > > reading the deciding arm of the prompt, which I never opened. I verified their reading
   > > independently; it holds. Full account: NOTES, "I MEASURED MY OWN PROMPT".**
   > >
   > > **`lead_with` — the ANSWER side of the join — is derived from EXACTLY FOUR named fields**
   > > (`satisfaction_condition`, `value_proposition`, `trust_threshold`, `recurring_value`), and
   > > **`money_flow` is not one of them.** `[MEASURED 2026-09-03, my own count over the live
   > > 9,643-char prompt]` occurrences of `money_flow`, `price`, `cost`, `pay`, `charge`, `£`:
   > > **zero, all six.** And the join rule says *"Do NOT stretch a point to cover a question it does
   > > not answer, and do NOT add a lead_with point merely to close a gap."*
   > > **So `unanswered: true` on a price question was GUARANTEED before any site was analysed.**
   > >
   > > **The defect is mine and it is sharper than "money_flow is missing": the two halves read
   > > DIFFERENT POOLS.** The question side is told *"Derive each question from a named field of the
   > > strategy you were shown"* — the WHOLE register — while the answer side reads four fields of
   > > it. **Any question sourced outside those four is unanswerable by construction**, which is
   > > exactly the three fields that carry unanswered questions (`money_flow`, `growth_path`,
   > > `recommended_page_types`).
   > > ⚠ **Disconfirmed in one place, stated rather than hidden:** `competitive_position` is also
   > > outside the four and has **0 unanswered of 8** at rank 2.9 — probably because
   > > `value_proposition` covers the same ground. So the rule is "outside the four is unanswerable
   > > **unless another of the four covers the same ground**".
   > >
   > > **My proposed CAUSE is refuted too.** I wrote that the model ranks by how well it can answer;
   > > the prompt says *"Rank 1 is the FIRST doubt, not the one most important to us."* Their better
   > > account is the **exemplar** — *"what will this actually get me and how much work is it to get
   > > it"* names outcome and effort, never price. ⚠ Hold the RANK half loosely (`recurring_value` IS
   > > named and sits at 4.7); hold the JOIN half firmly.
   > >
   > > **WHAT SURVIVES:** the model raised price questions on six sites **unprompted** — nothing in
   > > the prompt names price — and ranked them 4.6 rather than dropping them. That is evidence about
   > > the VISITOR, not about our copy, and it is why the defect is worth fixing.
   > > **WHAT DOES NOT:** "the register does not answer what this costs me." The measurement cannot
   > > distinguish that from "the instrument cannot represent a price answer", and the two want
   > > opposite responses — an editorial campaign across 18 sites, or a one-clause prompt fix.
   > >
   > > **⚠ OWED, NOT DONE: the prompt fix, and it needs the owner on one point.** Add `money_flow` to
   > > TASK 1's source fields, or name price in the exemplar, or both — my prompt, my agent, my
   > > change. But `lead_with` requires *"a benefit to the reader, never a description of us or of our
   > > inventory"*, so **someone must decide whether "£29, no subscription" is a benefit worth
   > > LEADING WITH, or whether price is a doubt we deliberately answer further down the page.** That
   > > is a judgement about how these sites sell and it wants his word, not ours.
   > > ⚠ And migration `723` edited this same prompt — anchored replaces on DISJOINT anchors, and see
   > > §2.1: `723` is not idempotent, so do not model the new one on it.
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
> > **⚠⚠ THE ILLUSTRATED TEXT BLOCK HALF: DO NOT WRITE THE MIGRATION — THE PREMISE IS STALE AND THE
> > COMPONENT STARTED BEING CHOSEN TODAY.** `[MEASURED 2026-09-03 15:35:02Z]` 08-24→09-02:
> > **2,382 sections built, ZERO** `illustrated-text-block`, nine days running. **09-03: 6 of 514
> > sections (1.2%), on 2 sites** — `dartsonline.com` (5) and `advertise.co.uk` (1) — all six created
> > between 14:08:00Z and 14:10:47Z. So the owner ruled "switch the switches" against a premise
> > (*effectively never chosen*) that was true on 09-02 and is **not true now**.
> > ⚠ Mind the filter: a bare count returns **12 / 2 sites**; pairing
> > `build_status='deployed' AND status='active'` returns **6**. Only the second is served.
> > ⚠ **I have NOT established what changed and did not guess.** Two migrations landed in the
> > preceding hour (`736` 12:42:25Z, `687` 13:45:54Z) and **neither names this component or its
> > selection**; the timing equally fits a roll or another lane's prompt work. Recorded as
> > unattributed.
> > **THE OPEN JUDGEMENT, and it is this seat's: is 1.2% enough?** Zero was plainly wrong; 1.2% may be
> > right, a first trickle, or two sites' idiosyncrasy. **One day is not a rate** — give it three or
> > four more days of the same table before deciding. The table and the query are in NOTES under "THE
> > SECOND FLIP'S PREMISE IS STALE".
> > ⚠ **Cross-lane contradiction for its owner, not for us:** §5 records `agentchassis-ff`'s
> > measurement that `dartsonline.com`'s 22 content pages are **zero illustration-capable**, "no page
> > can host a per-section figure regardless of rows" — and **5 of today's 6 landed on dartsonline**.
> > Either a rebuild overtook it or the two counts mean different things. **Nobody quotes either
> > figure until one of us reconciles them.**
> >
> > **⚡ COUNCIL r1 = REVISE (9 of 11 seats approving), r2 SUBMITTED 15:30Z, verdict pending on the
> > SAME corr `2ac895f3-ca82-4dbe-8f4e-3335a04b8925`.** The gating objection from `bug_historian` was
> > **right**, and answering it changed something in the estate rather than just in the file:
> >
> > **The LANDMINES entry "a `source: static` field OVERWRITES your stored `content_data` on every
> > section resolve" is HALF STALE, and I have corrected it in place.** It says the opposite of this
> > migration's central claim, at the identical call site. The dates settle it: the entry was
> > measured **2026-08-03**; `carryStored` entered `plan_sections_action.go` **2026-08-11**
> > (`d26c26a9a`); the renderer/**static** branch got it **2026-08-14** (`8f899cc8d`, `fix(268)`) —
> > **eleven days after the entry was written.** Corroborated in live data, not just git: 11
> > instances across 6 sites store a static-source value differing from its fallback, the
> > load-bearing ones written *after* the fix (`mortgagecalculator.co.uk` `tool-list.card_link_label`
> > = "Work it out" vs fallback "Open tool", `updated_at` 2026-09-01 02:44Z).
> > ⚠ **The entry's `query.*` half is STILL LIVE and is NOT retired.**
> > ⚠ **The generalisation is worth more than the entry:** a landmine is a snapshot of a DEFECT, and
> > a defect is the thing most likely to be FIXED — so **a landmine goes stale in exactly the way its
> > own advice cannot detect**, and reads as a live warning for ever. This one was right for 11 days
> > and misleading for 20. Before acting on any entry: `git log --since=<its added date> -S '<the
> > symbol it names>' -- <the file>`.
> > (⚠ My LANDMINES edit was swept into another session's commit `276d65655` before I could commit
> > it. Verified intact in HEAD; nothing lost. Forward-only holds.)
> >
> > **⚠⚠ AND THE THING THE WHOLE COUNCIL MISSED, found only because `tooling_provenance` objected
> > that I had not read the travelling docs: `info-card-grid` HAS AN ACCEPTANCE FENCE, and one of its
> > six checks is `no_horizontal_overflow`, desktop AND mobile** (`doc_plans`,
> > subject_type=`component`, subject_key=`info-card-grid`, 2026-08-05). A carousel is by
> > construction a horizontally overflowing track. **It should pass** — the checker exempts any
> > element with a scrollable ancestor
> > (`internal/adapters/browserrunner/run_checks_action.go:1094-1104`, "a scroll container makes the
> > width reachable"), and the carousel puts `overflow-x: auto` on the cards' direct parent.
> > **BUT THAT IS A MECHANISM READ, NOT A RUN:** the recorded acceptance pass (2026-08-05, 10 of 10)
> > was taken on `ai-agent-orchestration.com/services.html`, a **flag-unset GRID** placement, so
> > **the fence has never run against a carousel.** ⚠ **WHOEVER APPLIES 740 OWES ONE FENCE RUN
> > AGAINST A FLIPPED PLACEMENT.** It is the one check the migration cannot make for itself.
> >
> > Also answered in r2: the boolean-fallback footgun is measured fleet-wide and the class is
> > **EMPTY** (2 boolean fields, both correct, zero non-boolean) so the platform guard is declined
> > with its threshold stated; the UPDATE predicate now mirrors the drift guard; a real whole-schema
> > pre-image with an assertion mutation-proven by a mutant that passes every count check; and the
> > landing mechanism is the existing `page_rerender` queue (9,723 complete, latest 15:26:59Z today)
> > with the **mixed-state interim named as a real visible cost**.

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
- **`site_delivery_and_editor` (boxingonline.com, first paid build)** — NEW 2026-09-03. They asked
  this seat for the palette values and the pinned logo/wordmark call; **neither answer was the one
  asked for, and one of mine was WRONG.** Full record:
  `docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/CONTRIB_2026-09-03_from_the_designer_seat_the_palette_is_already_ruled_and_the_logo_is_the_real_defect.md`.
  · **Palette: no decision needed** — owner ruled 09-02, *"the cream off white decision is fine — no
  design churn"*, and the site already serves BOTH halves of the "self-contradicting" brief
  (`#0a0a0a` header AND footer, red, gold, on the off-white ground). They accepted and logged it.
  · **Wordmark: I said yes and I was WRONG** — the owner ruled *"(2) header stays LOGO-ONLY. Closed."*
  on 09-02, at `docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/NOTES_webdesign_uk_build_service.md:6905`.
  I asserted an absence from a phrase-grep over one lane's directory. **Fourth false absence from my
  own grep in a week**; logged in `WRONG_CALLS.md`, CONTRIB corrected in place.
  · **The finding that stands: the logo shares NO colour with the site.** 52.4% blue, 45.3% neutral,
  ONE red pixel of 16,372; zero within ±60 of brand red `#C0392B` or gold `#D4A017`. A raised
  protest fist in a diamond, illegible at the served 40px, losing 48.4% of its ink on its own
  near-black header. **They reproduced the census independently on a different artefact** (206,018
  opaque px, 53.2%/45.0%, same zeros) **and found the sharper half: their prompt asked for "a
  stylised boxing glove or ring ropes", so it is a subject-fidelity miss (the 417 family), not a
  palette or transparency one.** Going to the owner as a new question; nothing owed back here.
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

---

# UPDATE — end of 2026-09-03. Read this before §2 and §3; it supersedes parts of both.

**TWO MIGRATIONS ARE WRITTEN, COMMITTED, SUBMITTED AND *UNAPPLIED*. Neither is yours to re-derive.**

| | what | corr | state |
|---|---|---|---|
| `740_..._HOLD.sql` | `info-card-grid`'s `carousel` defaults ON at resolution time | `2ac895f3-ca82-4dbe-8f4e-3335a04b8925` | **r3 APPROVED 16:02:39Z** (r1 REVISE, r2 REVISE). Still `_HOLD`, still unapplied |
| `747_..._HOLD.sql` | `offer-analyser`: price may LEAD, and the exemplar names it | `aeaf9f88-4348-4453-8c9e-213e7fd548a7` | **r2 APPROVED 16:18:20Z** — and r2 reviewed the wording the file actually contains (r1 approved superseded text). Still `_HOLD`, unapplied |

> **⚠ BOTH ARE NOW `_HOLD.sql`, AND THE RENAME IS THE POINT.** An unapplied migration sitting in
> `sql_for_agents/` is **not held** — any session running `run-migrations.sh --apply` without scoping
> `MIGRATIONS_DIR` takes EVERY pending file. I was not holding these; I was hoping nobody ran the
> runner, and a peer lane caught it. Verified in the source, not taken on trust:
> `run-migrations.sh:65` sets `SIDECAR_RE='_[A-Z][A-Z0-9_]*\.sql$'`, `:284` filters the pending list
> with `grep -vE`, `:294` still LISTS it. **Control run, because a filter that excluded everything
> would look identical:** an ordinary migration filename still reads as would-be-applied.
> **Drop the `_HOLD` suffix at the moment of applying, never before.**
>
> **⚠⚠ `747`'s r1 came back APPROVED AND I AM NOT CLAIMING IT — and an APPROVED verdict is the
> DANGEROUS one here.** It approved *"rank it LAST"*; the file now says *"rank it LAST BY DEFAULT —
> unless it is genuinely this site's strongest reader benefit, which is rare"*, a semantic softening
> made after dispatch (the owner said "**mostly**", and as first written a site whose real
> differentiator IS price would have had its strongest benefit forced to the bottom). So
> `Council-Reviewed:` on that correlation would be a **MISMATCH**, which is the coverage report's
> dishonesty surface. **A REVISE would have forced the resubmission anyway; an approval creates both
> the temptation and the paper trail for "approved, ship it" over text the file no longer contains.**
> Commits carry `Council-Submitted:` until a verdict approves the text that exists.
>
> **`747`'s three r1 advisories are all answered by measurement**, in its header: the multi-row
> locator guard (`offer-analyser` measured at exactly ONE active row and ONE prompt-carrying step, so
> the premise is false today and the guard went in anyway, proven by inducing a 9-pair abort); and
> `723` is still rollbackable after `747` — **answered by running it in a transaction and probing**,
> not by inspection: 723's anchor still occurs exactly once and both its inserted blocks are intact.
>
> **⚠ AND MY OWN VERIFY CAUGHT MY OWN HALF-FINISHED EDIT**, which is the best evidence the guards are
> controls. I re-worded `repl_a` and left the verify's needle on the old phrase; the `UPDATE` ran, the
> `NOTICE` printed its confident summary, and only the independent live-row read failed. **A check
> that shares state with the thing it checks can only confirm the author's intention, never the
> outcome.** Now a LANDMINE in its own right — and note it surfaced on the CLEAN run after a change,
> which is exactly when nobody re-runs.

## ⚠ THE OWNER RULED ON THE PRICE QUESTION, AND IT WENT TO A DIFFERENT LANE

His words reached `copy_quality_two_stage`, not this lane's log. Verbatim:

> *"we can add price in the exemplar, we don't need to be so strict as to not enable us to say what we
> need to. It's not 'never a description of our inventory' its a **deprioritise and mostly leave it
> very brief or out altogether**. Currently there is nothing wrong with this sentence that I can see:
> 'why would I pay £29 when I can get AI to analyse my idea for free?' (**if we can't offer anything
> better than the free model then we should think of another tool to create**.)"*

Three things in that, and the last is the biggest:

1. **Add price to the exemplar** — the direct answer to what we asked. `747` edit B.
2. **The TASK 1 absolute is REPEALED.** *"a benefit to the reader, **never** a description of us or of
   our inventory"* becomes a deprioritise. `747` edit A. **These two must ship together** or the
   prompt argues with itself — naming price in the exemplar while the absolute four sentences earlier
   forbids inventory description would tell the model to raise a doubt it is barred from answering.
3. **The `idea.uk` question is FINE as written** — that corrects how it was framed, including by me.
   The defect was never the question; it was that it is unanswerable and ranks 4.6. **Say it that way:
   "the model ranked a bad question badly" and "the model ranked a GOOD question badly because it
   could not answer it" read very differently to a later reader.**

> **⚠⚠ AND THE ASIDE IS THE SHARPEST THING IN IT, AND IT IS NOT A COPY INSTRUCTION.** *"if we can't
> offer anything better than the free model then we should think of another tool to create."*
> **A doubt we cannot answer is information about the OFFER, not a copy defect to close.** That is
> what the question hierarchy is actually FOR, and it took a product-shaped answer to show it. Not
> this lane's to act on — `idea.uk` has its own lane — but carry it, because it reframes what the
> whole instrument is for.

## ⚠ MY ACCOUNT OF THE PRICE GAP WAS WRONG TWICE AND THE THIRD IS STRONG-BUT-THIN

Both wrong versions are recorded in `747`'s header **because each one implies a different, wrong
migration.** Do not resurrect either.

- ✗ **"our sites do not address price"** — a property of the prompt. `money_flow`, `price`, `cost`,
  `pay`, `charge`, `£` all occur **zero** times in its 9,591 chars. Implied an editorial campaign
  across 18 sites.
- ✗ **"unanswerable BY CONSTRUCTION, because the answer side only reads TASK 1's four fields"** —
  **this is the dangerous one, because the migration it implies is "add `money_flow` to the four" and
  that is the wrong edit.** `[MEASURED 2026-09-03]` `from_field` on `lead_with` points is **OPEN**:
  competitive_position 17, content_strategy 13, **money_flow 5 across 4 sites**, revenue_models 2,
  growth_path 1 — and **2 of 7 money_flow questions ARE answered.** A strong prior, not a wall.
- ✓ **What `747` acts on.** Answered-rate by the question's own field: trust_threshold 29/30 ·
  satisfaction_condition 24/24 · competitive_position 9/9 · recurring_value 15/16 · value_proposition
  9/10 · **money_flow 2 of 7**. Uniquely, **ZERO money_flow questions are answered by a money_flow
  point** though five exist. Exactly two sites carry both; **on both I read the pair and THE MODEL WAS
  RIGHT TO REFUSE** — finetuning's point answers "will I be pressured to spend", idea.uk's lists
  contents. **Both cite `money_flow` and neither states a price.** That is the absolute forcing the
  price to be laundered into a non-price.
- ⚠ **n=2 carries that mechanism.** "Zero of seven, uniquely" is a *very* clean number, and cleanness
  is the tell this lane got wrong twice today. **A pre-registered post-apply check is in `747`'s
  header**, including the outcome that would show the account is only half right: if money_flow points
  start stating prices while the questions stay unanswered, the other half is that nothing asks a
  point to engage the **alternative** the visitor names — a separate one-clause migration,
  deliberately not folded in.

## ⚠ THE `competitive_position` RULE, SETTLED PROPERLY RATHER THAN BY ASSUMPTION

`copy_quality_two_stage` challenged it: "a point from field A covering a question from field B" is
precisely the shape a stretch takes, so is competitive_position's clean record really stretch-in-
disguise? **Ran it: no.** `[MEASURED 2026-09-03 15:43:30Z]` competitive_position 1 of 9 in the reuse
set (11.1%) against a fleet 14 of 91 (15.4%) — **below**, not concentrated. n=9 cannot decide that, so
**all nine pairs were hand-read**: eight tight single-use joins, one honest reuse. The formulation
that survives is theirs, sharpened: **"answers the same doubt head-on", not "topical overlap"** —
`value_proposition` IS the differentiation field, so it is not moonlighting.
⚠ **Denominator trap, stated because I nearly published the ratio:** 5-of-86 counts reuse EVENTS,
14-of-91 counts QUESTIONS touching a reused point. Different questions; never compare them directly.
⚠ **OPEN, and not closed by implication:** the fleet stretch rate is still ~15% and now has one fewer
explanation. Where the remaining stretches come from is a question my own instrument raises. **Not
folded into either migration.**

## ⚠ 740 IS APPROVED — AND TWO OF ITS ADVISORIES CORRECT CLAIMS I HAD MADE

Both are folded into the migration header and into `doc_notes`
(`subject_type='component'`, `subject_key='info-card-grid'`, 2026-09-03 16:12:33Z).

**1. "The existing `page_rerender` queue lands it" was THE WRONG DENOMINATOR, and a seat had to tell
me.** `render_guardian`: **assemble-mode** `page_rerender` (a `page_id` with no `spec.reason`)
re-embeds each section's existing stored HTML and never re-renders the template against a re-resolved
schema — **so it can never apply this fallback.** I grounded the claim only on
`rerender_page_sections_action.go:1450`, the SCOPED path, and never checked the split.
`[MEASURED 2026-09-03]`

| status | items | scoped (`spec.reason`) | assemble-only |
|---|---|---|---|
| complete | 9,781 | 1,264 | **8,517 — 87%** |
| unresolved | 1,751 | **1,712 — 98%** | 39 |

**87% of the completes I cited are a mode that cannot land the change.** The queue's busy-ness was
real and was the wrong evidence. **The forward-looking news is better and is a DIFFERENT fact — the
pending queue is 98% scoped. Quote that; never quote the completes.** The mixed carousel/grid
interim's length therefore depends on scoped rerenders reaching each of the 21 sites, not on
throughput.

**2. I named the wrong vehicle for the fence run, then measured it.** Three seats asked for a work
item to force it, and my r3 called `acceptance_run` "the right vehicle (277 rows)" — **named from the
type name and a row count without checking what the rows are.** `[MEASURED 2026-09-03]` **all 277 are
`handler_agent='tool-acceptance-agent'` with `spec.check` of `tool_acceptance_due`/`manual` — a TOOL
vehicle** — and this component's own 2026-08-05 fence run was driven by **no work item at all**.
**So none was filed, deliberately:** routing a component fence at a handler I cannot show runs one is
what `bugs_open/395`'s routing rule 3b exists to stop, and an unhandled row is worse than a recorded
gap. ⚠ **Still owed: a component-capable acceptance vehicle, or a hand-run of the fence against
`leopardessconsulting.co.uk/services.html`** — the one live carousel placement, so **this can be
answered BEFORE apply.**

> **⚠⚠ THAT PAGE IS A PAYING CLIENT'S LIVE SITE — and it is STILL the right target.** Flagged by
> `copy_quality_two_stage`; confirmed in the owner's own record, not on relay
> (`docs/agent_docs/docs024_key_docs_latest/about_page_commercial/PLAN_2026-07-24_about_page_commercial.md`
> D4, verbatim: *"a paying client's site (leopardess)"*).
> **The right question is "does the fence WRITE?", and the answer is measured NO:**
> `internal/adapters/browserrunner/` has **zero** `INSERT INTO`/`UPDATE … SET`;
> `check_tool_acceptance.go` files no work item and triggers no rerender on failure; screenshots go
> to the runner's own store. What reaches the site is an **HTTP GET of an already-served page**, and
> these are static sites in a bucket.
> **AND THE RISK ORDERING INVERTS.** `[MEASURED 2026-09-03 16:15:50Z]` leopardess/services is the
> **only** live placement with the `carousel` key set at all — `finetuning.uk` has one
> `info-card-grid` instance and no flag — so a "safer" portfolio target needs a `content_data` write
> **plus a rerender** first. **Choosing the portfolio site to avoid touching a client site means
> performing the very write the concern is about, aimed elsewhere.** The client page is lower-impact
> because nothing has to change for it to be testable.
> ⚠ **The grounding EXPIRES if the fence gains a write** — screenshot store wired on, a failure path
> filing a repair item, an acceptance flow triggering a rerender. **Re-run the two greps before
> hand-running it**; if either stops coming back empty, create a carousel placement on a portfolio
> site and accept the write there. Full reasoning: `doc_notes`, `subject_type='component'`,
> `subject_key='info-card-grid'`, 2026-09-03 16:16Z.

> **That is the fourth time today I asserted from a name or a count instead of the thing itself** —
> a served count without the template, a missing ruling without the ruling headings, a model output
> without its prompt, and now a work-item type without its rows. **Same check every time: open the
> thing.**

## ✅ THE 2026-09-04 FLEET ROLL (v1.0.1361, cut `06c0b18f2`) DISTURBS NOTHING IN THIS FILE — CHECKED

The obvious worry on picking this up is that a roll landed after the handoff was written, so every
measured grounding below might be stale. **It is not, and here is the check rather than the
assurance.** `git log 239ab3626..06c0b18f2` over the five paths this lane's claims rest on:

| path | why this lane depends on it | commits in the roll |
|---|---|---|
| `internal/adapters/browserrunner/` | the fence-is-read-only grounding for running on a client page | **0** |
| `plan_sections_action.go` | `carryStored`/fallback precedence — `740`'s whole mechanism | **0** |
| `rerender_page_sections_action.go` | the path that actually applies the fallback | **0** |
| `datahelpers/content_type_violations.go` | `IsEmptyContentValue` — why a stored `false` beats the fallback | **0** |
| `discovery_checks/check_tool_acceptance.go` | no work item / no rerender on a failing fence | **0** |
| **control — all `*.go` in the same range** | proves the greps discriminate | **18** |

⚠ **And note WHY git is the right instrument here and the pod is not:** these groundings are claims
about **what the code does**, not about which binary is serving. A roll cannot change them without
changing those files. (For "did my fix ship?" the answer is the opposite — `service_binary_capabilities`
per service, and mind that it is a **two-hour window**, not a history.)

**Both migrations are unaffected in any case:** they are `content_components` / `agent_definitions`
config, applied by hand, and a Go roll neither applies nor alters a migration. Nothing was in flight
to lose — both council runs completed 2026-09-03.

## What the next session owes, in order

1. **BOTH ARE APPROVED — `740` at r3 (16:02:39Z), `747` at r2 (16:18:20Z) — and both advisory sets
   are folded into their headers. They are ready to apply.**
   **Order: `747` first** (it is the owner's ruling and nothing about price moves until it applies;
   then run its pre-registered post-apply check). **Then `740`, after the fence run** — which needs
   no apply and is the one thing the migration cannot check for itself.
   ⚠ `747`'s single r2 advisory is already discharged: `editquality` caught that the submission
   warning about unit-less lengths was **itself** carrying two versions of that length (9,590 vs
   9,591 chars). Reconciled to **9,590 characters / 9,642 bytes** throughout.
   ⚠ **Do not hand-write `Council-Reviewed:` anywhere.** Both migrations' commits carry
   `Council-Submitted:`, and `098` resolves a correlation to its **LATEST** verdict at report time —
   which is exactly why both rounds went back on the SAME correlation. **Had `747`'s r2 opened a NEW
   correlation, the committed trailer would have gone on resolving to r1's APPROVED — a review of
   superseded wording — while the real review sat on a correlation no commit names, and the coverage
   report would have shown a clean credit with no mismatch flag.** Same-correlation resubmission is
   what keeps the trailer honest when the file moves under an in-flight round.
2. **Apply `747`** once approved — it is the owner's ruling and nothing moves until it applies.
   Then **run the pre-registered check** in its header. ⚠ It edits the same prompt as `723`; `723` is
   NOT idempotent (§2.1), so never model a future edit on it.
3. **Apply `740`**, then ⚠ **run the component acceptance fence against a carousel placement** —
   `leopardessconsulting.co.uk/services.html` has carried `carousel: true` since ~07-31, so **this can
   be answered BEFORE apply** on an instance that already exists. `acceptance_run` is the item type.
4. **Do NOT write the Illustrated Text Block migration** — see §3.2; its premise went stale today.
