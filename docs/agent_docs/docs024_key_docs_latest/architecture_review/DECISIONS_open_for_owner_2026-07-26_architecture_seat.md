# Decisions open for the owner — an architecture seat / process

**Status: MOSTLY BUILT AND LIVE as of 2026-07-27.** This file began as proposals;
do not read the header of any section as current state without checking its own
status line. Current state in prose:
`SUMMARY_2026-07-27b_architecture_seat_built.md`.

| decision | state |
|---|---|
| D1 forward seat argues the forward side only | BUILT |
| D2 seated at `feature-designer` (design stage) | **LIVE** |
| D3 advisory, no veto, routes to the RFC track | **LIVE** (twice corrected — see the seat script) |
| D4 mechanical RFC trigger from the staged diff | **LIVE** in `scripts/commit-scope-report.sh` |
| D5 measure ossification first | **RUN — confirmed** |
| D6 name the asymmetric bar | **DONE 07-27** — `PROCESS_architecture_review.md` |
| D7(a) does the veto survive | **RULED: yes** (owner) |
| **D7(b) should the guardian weigh benefit** | **⇦ THE ONE DECISION OPEN FOR THE OWNER.** My recommendation **REVERSED 07-27 late**: do **not** narrow it — defer, with a named reversal trigger (see D7) |
| D8a′ council reads its own minutes | **LIVE** on all three councils |
| D8e-1 generated case index for the historians | **LIVE** |
| D9 does the FIX lane need the forward seat too | **NEW 07-27, arose from evidence.** Recommend DEFER, do not seat a second copy; countable reversal trigger named |

**The owner's decision, in one line:** D7(b) — *should the guardian keep being
asked to weigh **benefit**, or be narrowed to blast radius and contract-breakage
only?* Everything else on this workstream is built, live, or deferred behind a
named trigger. My advice has **changed** since this file last put it to you, and
the argument for the change is inside D7; the short version is that the guardian's
failure looks like **ignorance rather than remit**, and we have now fixed the
ignorance, so narrowing the remit would be the wrong repair — and an irreversible
one. It is also **safe to leave open**: nothing is blocked on it and no code waits
on the answer.

**Written 2026-07-26** after the owner asked whether we had discussed an
architecture council member before, and asked for the choices and proposed
decisions in one place.

**Updated 2026-07-27.** Owner ruled **D7(a): the veto survives**; D7(b) left open,
with the argument *against* narrowing it now recorded alongside my inclination.
The owner also asked whether the council looks at the missteps, which became
**D8** — and then pointed at the concept register, which caught a defect that
would have broken the feature-build lane (see the seat script's second correction).

**Baseline for judging all of this, captured 2026-07-27 before any post-change
council ran** (`scripts/council-adoption-report.sh`):

| seat | reviews | invoked stability pref. | cited precedent |
|---|---|---|---|
| guardian | 205 | 87 | **3** |
| bug_historian | 139 | — | 37 cited a source |
| debug_historian | 173 | — | 19 cited a source |
| architecture | 0 | — | — |

~~**3 of 87 is the number to beat.**~~ That is how often the seat that most needs its
own history actually referred to it, while invoking the very preference that needs
it. *(An earlier run of the report said 4; the query carried stray literals that
matched spuriously. Corrected here — the committed script gives 3.)*

> **CORRECTED AGAIN 2026-07-27 late, and this time the defect was structural, not
> a stray literal.** The sentence above describes an **intersection** ("referred to
> it *while* invoking the preference") that the query never computed: `invoked` and
> `cited` were two **independent** `FILTER`s over all guardian reviews, so neither
> "3 of 87" nor the later "6 of 90" was ever a subset. On the same corpus, **4 of
> the 6 cited precedent without invoking the preference at all**, and the true
> intersection is **2 of 90 = 2.2%**. Fixed in
> `scripts/council-adoption-report.sh` (headline is now `both_invoked_and_cited`,
> with `cited_but_did_not_invoke` kept visible). **Take 2 of 90 as the baseline.**
> The correction makes the D5 case *stronger*, not weaker — the seat referred to
> its own history less often than we had said. Caught by two figures both labelled
> "pre-change" disagreeing about a population that could not have changed; full
> entry in `WRONG_CALLS.md`.

Named `DECISIONS_open_for_owner_*` rather than `SUMMARY_*` deliberately: the
content is choices awaiting a ruling, which is the shape of
`fixloop_eg_dartsonline/DECISIONS_open_for_owner_2026-07-18.md`, not the
five-part milestone read-out that `SUMMARY_` means in this repo. A SUMMARY
follows if and when something is decided and built.

---

## 1. The question

> A process — which might be a council member — that ensures the architecture is
> robust, doesn't change too often beneath us, and is sufficient and good, even
> best, for our current and anticipated plans. There will be conflict between
> rewriting to make it the best for each new decision, and keeping what we
> already have that we know works and has been battle-tested.

---

## 2. What already exists

### 2a. An architecture-review track — built, used once

`architecture_review/PROCESS_architecture_review.md`, created 2026-07-25 by
owner decision (commits `7ed97d36a`, `29b3bab94`, `8b4538b1c`). It carries a
four-condition trigger test, a seven-section RFC template (problem · design ·
alternatives · blast radius · staged rollout · rollback · acceptance evidence),
and owner-as-authority. `RFC_001_at_least_once_delivery.md` is RATIFIED.

So the *process* half of the question is largely answered already. What follows
is mostly about staffing and triggering it.

### 2b. One council seat already holds the conservative half — verbatim

`review_guardian`, live in `agent_definitions` (type `council-gate`), is the
only hard-veto seat of sixteen. Its charter, quoted exactly:

> **(b)** architecture-change signals — edits to shared contracts, wire formats,
> message shapes, exported signatures, or MANY packages at once mean this is not
> a constrained fix: veto and say it needs an architecture review.

> **(d) STABILITY PREFERENCE** — long-stable, load-bearing infrastructure (the
> orchestrator, Kafka/messaging, agent spawning, the core work-item dispatch) is
> battle-tested; PREFER a fix at a higher, less-foundational layer over editing
> it. An edit to this core is itself a strong architecture-change signal: object
> and ask whether the cause can be addressed above it, and reserve veto for a
> genuine architecture change to foundational plumbing dressed as a point fix.

Clause (d) is already, almost word for word, the "doesn't change too often
beneath us" half of the question. Clause (b) is already the RFC trigger.
**Neither clause asks whether the architecture is sufficient for what we plan to
do next.** That absence is the gap.

### 2c. The gap, stated in the repo today

`bugs_open/086_HANDOFF_2026-07-26_...:254`:

> **There is no architecture-review agent.** The council gate is the review
> system; the fleet's six `*-architect` agents design websites, not code.
> "Architecture review" means the owner.

Verified: the six are `site-architect`, `site-component-architect` (×2 rows),
`content-site-architect`, `landing-page-architect`, `portfolio-architect` — all
website assembly, none platform code.

### 2d. There are THREE council rosters, at three lifecycle points

> **CORRECTION 2026-07-26.** An earlier draft of this analysis asserted that the
> council "reviews a plan that already exists — by then the shape is decided, so
> an architecture seat there can only ever say no," and concluded the forward
> half had nowhere to live. **The owner corrected this mid-session: the council
> is used in advance of the build too, and in the diagnosis loop.** That is
> right, and it changes the proposal materially — the seam I claimed was missing
> already exists. What caught it: the owner, not a check I ran. The cheap check
> I skipped was one query listing `review_*` steps across all agent types, not
> just `council-gate`.

| # | lane | agent | seats | when it fires |
|---|---|---|---|---|
| 1 | experience | `experience-planner` | `journeys`, `contracts`, `honesty`, `feasibility`, `mvp` | composing an EXPERIENCE_PLAN — earliest |
| 2 | feature build | `feature-designer` | `guardian`, `editquality`, `bug_historian`, `guidelines`, `reuse_agent` | design from an owner-approved capability spec, **before any code** |
| 3 | fix / manual | `fix-proposer`, `council-gate` | 16 (relevance-gated) | an edit plan, before commit |

Two observations that matter more than anything else in this file:

- **Lane 1 already contains forward-looking seats** — `review_feasibility` and
  `review_mvp` judge whether a plan is buildable and correctly sized. That is the
  closest existing thing to a forward advocate. It exists only in the experience
  lane, and never sees platform code.
- **The guardian's hard veto operates with the fewest counterweights at exactly
  the point where forward judgement matters most.** At lane 3 it sits with 15
  others including `mission`, `constitution` and `prior_art`. At lane 2 —
  designing a feature before it is built — it sits with only four, and *none* of
  them is a mission, constitution or forward-fitness seat. The earlier the
  decision, the more one-sided the review.

---

## 3. What we measured — and a correction to the premise

The worry "doesn't change too often beneath us" is testable from git. It had not
been tested. It has now.

```
git log --since="60 days ago" --oneline -- <path> | wc -l
```

| scope | commits, 60d |
|---|---|
| **all commits in repo** | **2,123** |
| commits touching `platform/` at all | 377 (17.8%) |
| `platform/orchestration/` | 366 |
| ├─ `platform/orchestration/actions/` | 348 |
| └─ orchestration core, excluding `actions/` | 55 |
| `platform/orchestration/coordinator.go` alone | 9 |
| `platform/storage/` | 8 |
| `platform/agentbase/` | 7 |
| `platform/messaging/`, `platform/kafka/` | 6 each |
| `platform/health/`, `platform/aiservice/` | 5 each |
| 11 further `platform/*` packages | 0 |

Top churning files inside `orchestration` are `actions/registry.go` (52) and
individual action files (`v3_site_actions.go` 22, `plan_sections_action.go` 16,
various `diagnose_*_action.go` 11–15).

**Read correctly, this refutes the premise it was meant to test.** The apparent
366-commit churn in the orchestrator is 95% growth in a **plug-in action
registry** — new actions being registered, which is precisely what a registry is
for and touches no contract. The actual core moved 55 times in 60 days, and the
coordinator itself 9 times, against 2,123 total commits.

> **CORRECTION 2026-07-26.** In discussion I said "churn is measurable and
> currently unmeasured", implying measurement would show a problem. It shows the
> opposite: **the load-bearing core is remarkably stable, and the architecture's
> own shape is what makes it so** — an action registry absorbs feature growth
> without the core moving. Caught by running the measurement instead of
> asserting it. This is the second premise in this file corrected by evidence
> rather than argument, which is itself an argument for D5 below.

**Therefore the live risk is more likely the opposite of the one named.** Not
churn — **ossification**: a core that is politically expensive to touch (one
seat can veto edits to it, by charter) accumulating workarounds in the layer
above it, which is exactly where all the movement is. This is a hypothesis, not
a finding: **[INFERRED]** from the churn distribution plus clause (d)'s explicit
"prefer a fix at a higher, less-foundational layer". The check that would settle
it is D5.

---

## 4. The conflict, restated against the record

The conflict the owner anticipates is real, already running, and **structurally
one-sided**: stability holds a hard veto; sufficiency-for-anticipated-plans has
no seat, no veto, and no trigger anywhere in three rosters.

Guardian veto record, from `doc_notes` (`categories ? 'council-gate'`) and the
workstream docs:

| case | guardian | what happened next |
|---|---|---|
| `003` delivery redesign | veto ×2 | shipped (accidentally, via a ride-along build); 19 retries / 7 end-to-end recoveries in first 4.5 h; owner ratified keep-live |
| `030` dispatch lane | veto, "naming no defect" | contested with evidence over three rounds → **veto lifted, APPROVED**; publish→start 18 min → 1 s |
| `086` `error_step` converter | veto | owner overruled; measurement showed 109 real failures/30 d would have been handled instead of fatal |

Three `REJECTED — hard veto from guardian` rows on 2026-07-26 alone.

**Being fair to the seat.** None of this proves the guardian was wrong ex ante. A
risk that did not materialise was still a risk; in `003` the deploy was
accidental, so the cautious counterfactual was never run; and in `086` the
veto's *shape* was judged right even as its sizing was judged wrong. The honest,
narrower claim is about the process, not the seat:

> **The veto has never yet been sustained when escalated.** Every time it was
> tested against a measurement, it was overturned. A block that is always
> overturned on review is not calibrated — it is imposing a round-trip through
> the owner as the price of every structural change.

That is the thing to fix, and it is fixable without weakening the seat: give the
veto a counterparty that argues cost-of-not-changing with the same rigour, and
give it a measurement to be calibrated against.

---

## 5. Proposed decisions

Each is stated as a choice with a recommendation and the argument against it.

### D1 — Do NOT add a second brake. If a seat is added, it argues the forward side.

**Proposal.** Any new seat's remit is *sufficiency for current and anticipated
plans* and *cost of not changing*. The conservative remit stays where it is,
with the guardian.

**Why.** A single seat holding both remits collapses into the conservative one,
because "this is battle-tested, don't touch it" is always cheaper to evidence
than "this will not carry us in three months." That collapse is observable —
clause (d) is the forward-looking language, and in practice it functions purely
as a brake. Adding a second veto-holder makes the gate unpassable.

**Against.** If the true risk is ossification (§3, unproven), a forward advocate
is right; if the true risk is churn, this makes it worse. D5 settles which.

---

### D2 — Put the architecture question at lane 2 (`feature-designer`), not lane 3.

**Proposal.** The forward seat joins the **feature-designer** roster — design
time, before code exists — rather than the fix gate.

**Why.** Lane 2 is the earliest point platform code is shaped, and it is where
the guardian currently sits with the fewest counterweights (§2d): four seats,
none of them mission, constitution or forward-fitness. It is also the only lane
where the answer can still change the design rather than merely block it.

**Against.** Lane 2 fires rarely (feature builds are infrequent) so the seat
would get little exercise, and most platform change still arrives via lane 3.
A variant worth considering: seat it in **both**, with veto in neither.

---

### D3 — Advisory, no veto; its verdict is an RFC trigger, not an objection.

**Proposal.** The seat outputs `needs_rfc` / `fine_as_point_fix` plus its
reasoning, rather than approve/object/veto.

**Why.** It converts the RFC track's currently-manual trigger test into a
routine one, and it cannot deadlock against the guardian. Note that `RFC_001`
was written *after* the code was already in production — the trigger did not
fire because nothing fires it.

**Against.** An advisory verdict nobody must act on may be ignored. Mitigated by
D4, which makes the trigger mechanical and therefore visible whether or not the
seat is consulted.

---

### D4 — Make the RFC trigger mechanical, computed from the diff.

**Proposal.** Extend `scripts/commit-scope-report.sh` (already a `pre-commit`
hook, already reads `git diff --cached --name-only`, `scripts/commit-scope-report.sh:26`)
to also print `ARCHITECTURE SIGNAL` when the staged diff meets the
`PROCESS_architecture_review.md` trigger test: ≥3 top-level `platform/*`
packages, an exported-symbol change, or a migration touching a shared contract.
Advisory, never blocking — same posture as the existing scope block.

**Why.** Cheapest real change on this list; reuses machinery that exists; and it
catches the failure mode that actually occurred (RFC written after the fact)
without requiring anyone to remember a four-condition test.

**Against.** Exported-symbol detection needs more than filenames, so the first
version would be package-count and migration-path only. That is still most of
the value. **[UNMEASURED]** — how often the signal would fire on real history
has not been computed; worth doing before building it.

---

### D5 — RUN 2026-07-27. Result: ossification is real, and the exercise is justified.

**Owner approved the measurement 2026-07-27. It has been run. §3's `[INFERRED]`
hypothesis is now measured, and it holds.**

The corpus: 259 `council_report` artifacts, giving **204 guardian verdicts —
139 object, 53 approve, 12 veto** (veto rate 5.9%, objection rate 68%), and
**437 individual guardian objections**, of which **29 invoke the stability
preference** explicitly ("higher layer", "less-foundational", "battle-tested",
"foundational").

The decisive figure is not the count but the **recurrence**:

| core site deflected upward | distinct submissions | first | last |
|---|---|---|---|
| `coordinator.go` / `SagaCoordinator.ProcessResponse` | **6** | 2026-07-20 | 2026-07-26 |
| `spawn_actions.go` / `spawnAgentKubernetesJobFromDefinition` | **4** | 2026-07-21 | 2026-07-26 |
| `platform/kafka` consume lane / `processRequests` | 2 | 2026-07-26 | 2026-07-26 |

**Six independent submissions needed to change `ProcessResponse` in seven days,
and each was told to fix it somewhere higher.** That is the ossification
signature: if the higher-layer fixes had held, the same site would not keep
coming back. Four bugs currently sit open *in that core* — `075` (ownership
discard), `086` (converter, coordinator.go), `034` (validation, coordinator.go),
`096` (council head-of-line blocking).

Set against §3's churn split — `coordinator.go` moved **9 times in 60 days**
while the action registry above it moved **348** — the two measurements agree:

> **Pressure on the core is high and rising; actual change to the core is near
> zero; the difference is being absorbed as workarounds in the layer above.**

One objection in the set is the guardian recording its own deflection failing:
*"The round-2 veto's named higher-layer alternative is refuted by hard evidence
(scheduler has no k8s capability…)"* — i.e. the safest contained alternative did
not exist. That is the mechanism failing in a single row.

**Verdict on D5's own question — "does this justify staffing anything?" — YES.**
It also re-weights D7(b): the guardian is deflecting to a higher layer 29 times
without any instrument for whether those deflections hold. It has no way to know
`ProcessResponse` is on its sixth visit, because (per D8) it cannot see its own
minutes.

*Method, for re-running:* `jsonb_array_elements(body::jsonb->'reviews')` over
`diagnosis_artifacts WHERE kind='council_report'`, filtered to
`r->>'reviewer'='guardian'`, then `jsonb_array_elements(r->'objections')`.
Site tagging is by `ILIKE` on the objection text — **[APPROXIMATE]**: it
undercounts anything phrased without the symbol name, so 6/4/2 are floors.

---

### D5 (original proposal, retained for the record) — Measure ossification before staffing anything.

**Proposal.** Before adding a seat, answer: **are we extending sideways because
the core is untouchable?** Concretely — for each of the last N guardian
objections/vetoes that recommended "a fix at a higher layer", did the higher
-layer fix hold, or did the same defect recur? Plus: how many known defects
currently sit in the orchestration core, deferred?

**Why.** §3 makes ossification a plausible hypothesis and nothing more. This is
the query that either justifies the whole exercise or kills it — and this file
has already had two premises corrected by measurement (§2d, §3), which is the
pattern.

**Against.** Costs a session before anything ships. That is the point.

---

### D6 — Name the asymmetric bar, so conflict is a bar and not a debate.

**Proposal.** Record in `PROCESS_architecture_review.md` that keeping
battle-tested code is the default and needs no evidence, while replacing it must
show all four of:

1. a defect the current design **cannot express a fix for** (not merely one it
   handles badly);
2. blast radius derived mechanically (`go list -deps`, compile-proof for
   removals) — never qualitatively;
3. stages that are **each independently valuable**, so a halt mid-rollout leaves
   a better system than before;
4. a rollback that does not require a schema migration (image-first).

**Why.** `RFC_001` already satisfies 2, 3 and 4 — this names the bar it met
rather than inventing one. Item 1 is the load-bearing addition: it is the test
that distinguishes "the architecture is insufficient" from "this instance is
broken", which is precisely the distinction the guardian keeps being asked to
make with no criterion to make it against.

**Against.** Criterion 1 is a judgement, not a measurement, and a determined
author can always argue it. It still beats no criterion.

---

### D7 — The veto, and whether the guardian weighs benefit

Two sub-questions. **(a) is now ruled; (b) remains open.**

- **(a) — RULED, owner 2026-07-27: THE VETO SURVIVES.** The question put was
  whether three overturns in two days meant the veto should become a strong
  objection that escalates rather than a block. It does not. The veto stays a
  hard block. (Supporting argument on the record: the veto is what *caused* the
  RFC track to exist, so it has already earned its keep once.)
- **(b) — OPEN, owner undecided 2026-07-27.** If a forward seat is added, should
  clause (d) be narrowed so the guardian stops being asked to weigh **benefit** —
  which it has no instrument to measure — and judges blast radius and
  contract-breakage only?

**My inclination on (b) stays yes**, and D8 below strengthens it: the guardian
is not merely un-instrumented for benefit, it is un-instrumented for *history*
too. But (b) is genuinely load-bearing and the owner is right to hold it — see
the argument against, which I had not made properly:

**Against narrowing (b).** Blast radius and benefit are not separable in
practice. "Is this worth the risk" is the question a veto *answers*; a seat
judging blast radius alone would have to veto every wide change, which is
strictly more conservative than today, not less. Narrowing (b) only works if
the forward seat actually exists and is read — so **(b) should not be decided
before D5, and cannot be implemented before D1/D2.** It is downstream of both.

> **UPDATED 2026-07-27 late — my inclination has REVERSED, on first evidence.
> Recommendation is now: do NOT narrow (b) yet. Defer, with the reversal trigger
> named below.**
>
> The argument for narrowing was that the guardian has no instrument for benefit
> and had been overturned on every escalation. The first post-cutover council
> (14:18:19, `b64141e5`) supplies the first evidence of how it behaves *with* an
> instrument, and it undercuts that argument. Given its own deflection history,
> the guardian invoked the stability preference and then reasoned its way **out**
> of deflecting, unprompted:
>
> > *"The recurrence across three rounds is evidence of a genuinely scattered
> > defect (multiple independent RenderContext producers), **not** evidence that
> > this fix belongs at a higher layer … so that preference does not bite here."*
>
> That is the exact judgement D5 measured it failing to make six times on
> `ProcessResponse`. **So the observed failure mode was ignorance, not remit** —
> it kept deflecting because it could not see that it had already deflected, and
> the fix for that was the minutes, which are now live. Narrowing a remit is the
> wrong repair for a missing instrument, and it is irreversible in the direction
> that matters: a blast-radius-only seat cannot recover the judgement above even
> when it is right.
>
> **Reversal trigger — the condition under which narrowing IS justified.** Re-run
> `scripts/council-adoption-report.sh` once there are **≥20 post-cutover guardian
> reviews that invoke the stability preference**, and narrow (b) if BOTH hold:
> (i) precedent-citation stays near the **2-of-90 (2.2%)** baseline, i.e. the minutes
> changed nothing; and (ii) deflections still recur on the same core sites
> (`coordinator.go`/`ProcessResponse`, `spawn_actions.go`). If instead citation
> rises and recurrence falls, the instrument fixed it and clause (d) should stand
> unchanged.
>
> **Two honest caveats on this reversal.** `n = 1` — one review, and it approved,
> so it was never a hard case. And §2 of the report scored that very review
> `cited_precedent = 0`, because the metric counts a precedent *citation* and this
> one reasoned about recurrence without quoting a past report: **the metric
> undercounts the behaviour I am citing as evidence.** Fix the metric before
> reading trigger condition (i), or it will read as "the minutes changed nothing"
> when they did.

---

### D9 — NEW, arose from evidence 2026-07-27: does the FIX lane need the forward seat too?

**Not a proposal — a question the evidence opened, recorded so it is not lost.**
D1/D2/D3 deliberately placed `review_architecture` on `feature-designer` only
(design time, before code exists). On the same 14:18 council — a **fix-lane** run,
which has no architecture seat — `bug_historian` opened its note:

> *"**Architecture-level concern for a human:** this bug class … has now needed
> three council rounds to chase down instances empirically … Recommend a human
> confirm whether a single shared render-context-builder refactor … is on the
> roadmap, versus continuing to fix drop points one live-test at a time."*

A seat not commissioned for forward fitness made a forward-fitness judgement, on
the lane with no home for one, and escalated it to a human because nothing there
owns it. That is the D2 gap reappearing one lane down.

**My recommendation: DEFER, and do not seat a second copy.** Three reasons, two
of them already on the record here:
- §8c's doctrine — reserve the seat budget for genuine judgement, and do not add
  a seat where the remit is already held or the instrument is the problem.
- Cost: the gate is relevance-gated, but a 17th seat fires on every submission
  whose paths match it, on the highest-volume lane (36 runs/day vs
  `feature-designer`'s ~3).
- **The seat we already have has said nothing yet** (0 reviews — see the rate
  limit below). Staffing a second copy of an unmeasured seat is exactly the
  mistake D5 was built to prevent.

**Reversal trigger:** if the fix lane produces ≥3 more architecture-level
escalations routed to "a human" from seats not commissioned for them, the remit
is unowned in practice and D9 should be decided rather than deferred. Countable —
`body ILIKE '%rchitecture-level concern%'` over `council_report` — so this does
not need a judgement call to re-open.

**`[UNMEASURED]`:** one instance, noticed by reading, not counted. The count above
is the check, and it has not been run.

---

### D8 — NEW: give the council a misstep corpus, or accept that it has none

**Discovered 2026-07-27, prompted by the owner asking directly: does the council
look at the missteps? It does not, and it structurally cannot.** Three
independent blocks, each verified:

| tier | what it can reach | verified by |
|---|---|---|
| `code_checks` | **Go only — 4,535 symbols, 0 markdown** | `SELECT ... FROM code_symbols GROUP BY path suffix` |
| `code_checks` | `WRONG_CALLS.md` specifically: **0 rows** | `SELECT count(*) FROM code_symbols WHERE path ILIKE '%WRONG_CALLS%'` |
| SQL `checks` | ten tables only — `pages, sites, site_plans, site_plan_pages, site_work_items, content_components, page_components, agent_definitions, diagnosis_artifacts, agent_error_log` | `load_schema_hint.config.query` |

Consequences, in ascending order of seriousness:

1. **`WRONG_CALLS.md`, `/bugs_open/`, `/bugs_closed/` and every working doc are
   invisible to every seat.** The entire written record of how we get things
   wrong cannot be consulted by the system built to catch us getting things
   wrong.
2. **`doc_notes` is not in the schema hint** — so the headline/summary layer of
   past verdicts is unreachable.

   > **CORRECTED 2026-07-27, same day, before anything was built on it.** This
   > point originally read: *"so the council cannot read its own verdicts …
   > each run starts with no memory of any previous run."* **That is wrong, and
   > it is wrong in the direction that matters.** The council's own history *is*
   > reachable: `diagnosis_artifacts` **is** one of the ten hinted tables, and it
   > holds **259 `council_report` rows** whose body is the full verdict JSON —
   > every reviewer's objections verbatim, with severities — plus 248 `fix_plan`
   > rows. The memory is on the shelf and always has been.
   >
   > What is actually true is narrower and more embarrassing: **no seat is told
   > it is there.** Of 32 `review_*` prompts across `council-gate` and
   > `fix-proposer`, **zero** mention `council_report`; only two
   > (`diagnosis_guardian`, `tooling_provenance`) mention `diagnosis_artifacts`
   > at all. The schema hint lists table names and columns with no indication
   > that one of them is the council's own minutes.
   >
   > **What caught it:** running the D5 measurement, which needed the guardian's
   > historical objections and therefore had to find where they were stored —
   > the query I would have skipped had I gone straight to implementing D8a.
   > **The cheap check:** `SELECT kind, count(*) FROM diagnosis_artifacts GROUP BY 1`
   > — five seconds, against a table already named in the hint I had read.
   > **Consequence for the plan:** D8a as written (add `doc_notes`) would have
   > bought the derivative summary layer while leaving the full corpus still
   > unmentioned. The fix is a **prompt** change, not a schema change — cheaper
   > than proposed and strictly better. See D8a′.
3. **The `bug_historian`'s "documented history" is seven narrative items
   hard-coded into its prompt template** — a hand-written constant ending
   "MOST RECENT: Go's template engine…". The seat named *historian* holds a
   frozen list, not a corpus. It goes stale silently, and nothing in its output
   distinguishes "no matching history" from "history exists but is not in my
   prompt".

**Why this belongs in this file rather than a bug.** It changes the proposal:

- **A forward architecture seat would inherit exactly this blindness.** Judging
  "sufficient for anticipated plans" needs a corpus even more than "has this
  failed before" does. Building D1/D2 without D8 produces a seat that reasons
  from its prompt, confidently, about a roadmap it cannot read.
- **It supplies the mechanism §6 was missing.** §6 said the seat needs a single
  roadmap artefact. D8 says *where it must live*: one of the ten hinted tables,
  or the prompt itself. **A roadmap in markdown is unreadable to the reviewer**,
  so "write a roadmap doc" would not have worked.
- **It weakens the case for D3.** An advisory seat trained to spot repeats,
  which cannot see the record of repeats, is close to decorative.

**Options, cheapest first.** *(All [UNMEASURED] for cost; none built.)*

- **D8a — add `doc_notes` to the schema hint.** One-line config change, live
  immediately, no image. Gives every seat its own verdict history. Risk: seats
  spend rounds querying it; the hint doubles as a budget.
- **D8b — index markdown into `code_symbols`** (or a sibling table) so
  `code_checks` reaches `WRONG_CALLS.md`, `bugs_open/`, `bugs_closed/`. Larger;
  needs whatever builds that index. Highest value per §1–3 above.
- **D8c — replace the `bug_historian`'s hard-coded seven with a query.** Depends
  on D8a or D8b. Until then the list should at least carry the date it was
  written, so staleness is visible.
- **D8d — do nothing, but say so.** Record in `PROCESS_architecture_review.md`
  that the council reviews *plans against prompts*, not plans against history,
  and that the misstep corpus is a **human** instrument. Honest, and cheap.

### D8a′ — SUPERSEDES D8a. **APPLIED AND LIVE 2026-07-27.**

**Owner ran it 2026-07-27.** `fix-proposer` patched (`UPDATE 1`, 5 seats), then
`099_SYNC_gate_roster.py --apply` mirrored to `council-gate` — dry run showed no
seats added or removed, routing OK, drift exactly the five patched steps, snapshot
taken. **Verified live: 5 seats × 2 agents = 10 rows** carrying `council_report`,
with the deflection check present on the guardian in both. Config is live
immediately; no image, no roll. Commands and the three gotchas are in
`RUNBOOK_architecture_seat.md`.

The five seats now told their minutes exist: `review_guardian`,
`review_bug_historian`, `review_debug_historian`, `review_prior_art`,
`review_reuse_agent`. The guardian additionally carries D5's finding — before
deflecting a change to a higher layer it is told to count how often that site has
already been sent upward, and that a site which keeps returning is evidence the
deflections are not holding.

**Not yet measured [UNMEASURED]:** whether the seats actually use it. The check
is whether `council_report` appears in any seat's `checks` array on subsequent
runs, and whether the guardian's stability-preference objections start citing
prior deflections. Worth reading after the next handful of council runs — and it
is the honest test of this change, not the fact that the text is present.

---

### D8a′ (as proposed) — Tell the seats their minutes exist.

Given the correction above, the right first move is **not** a schema change:

- The corpus is already reachable (`diagnosis_artifacts`, hinted, 259
  `council_report` rows with full verdict text).
- What is missing is **one paragraph in the seat prompts** naming the table, the
  `kind='council_report'` filter, and the `body::jsonb->'reviews'` shape.
- **Implementation path is the documented one**, not a hand-patch: edit
  `fix-proposer`'s `review_*` steps, then run
  `099_SYNC_gate_roster.py --apply` (snapshots first). **Landmine confirmed
  2026-07-27: the mirror copies `review_*`/`gate_*` steps only — it does NOT
  copy `load_schema_hint`** (099 line 117 carries non-review steps over from the
  gate's own copy). So a hint change would be a four-place edit across
  `council-gate`, `fix-proposer`, `feature-designer`, `experience-planner`; a
  prompt change rides the mirror for two of them.
- **`doc_notes` in the hint is now optional and secondary** — it holds the
  headline summaries, which are derivative of `council_report`. Worth adding for
  the guardian's containment notes, but it is no longer the load-bearing change.

**Highest-value single line, from D5:** the guardian should be able to ask *"how
many times has this core site already been deflected upward?"* Six visits to
`ProcessResponse` is a fact about its own past behaviour that it currently
cannot see.

---

### D8e — the debugging guides: the owner is right, and they are far too big

**Asked 2026-07-27: does the `bug_historian` look at the old debug docs?** No —
they are markdown, and per D8 no seat can read markdown. Sizes measured:

| corpus | bytes | files |
|---|---|---|
| `016b_debugging_guide_8_consolidated.md` | 455,474 | 1 |
| `016_debugging_guide_v2_58_consolidated.md` | 251,931 | 1 |
| `WRONG_CALLS.md` | 386,455 | 1 |
| `bugs_open/` | 359,862 | 32 |
| `bugs_closed/` | 1,841,386 | 89 |
| **total** | **~3.3 MB** | 124 |

**~3.3 MB is roughly 800k tokens against a seat budget of `max_tokens: 8000`.**
So "or if they're too big then some other agent should" is the correct read —
inlining is not on the table for any seat.

**But they are already chunked, and that is the opening.** `016b` §9 is a flat
list of `### <one-line pattern assertion>` headings, each self-contained and
dated — *"A mistyped routing key produces silence in every gate at once, not one
loud failure (2026-07-18)"*, *"An action that exists in code but in no registry
fails as 'requires a topic' — and the failure is stamped 'complete'"*. The
headings alone are a usable index at roughly **10–20 KB**, which *is* promptable.

Three shapes, cheapest first — **not yet decided**:

- **D8e-1 — heading index only.** Extract §9 headings + `WRONG_CALLS` `##`
  headings + `bugs_*` filename slugs into a table; give the historians the list
  and let them `code_check`/SQL for detail. Smallest, and it directly fixes the
  frozen-seven problem (D8c) by making the list generated rather than typed.
- **D8e-2 — a retrieval seat.** A dedicated agent that takes the plan, searches
  the full 3.3 MB, and returns the three most relevant prior cases to the
  council. This is the owner's "some other agent should" — correct shape, more
  build.
- **D8e-3 — index markdown into `code_symbols`** (was D8b). Most general,
  largest, and it also serves the reuse/prior-art seats.

**My recommendation: D8e-1 now** — it is small, it makes the historian's list
self-maintaining, and it is a prerequisite for judging whether D8e-2 is worth
it. Note the sequencing trap: D8e-2 built first would be a retrieval agent over
a corpus nobody has indexed.

---

**Revised recommendation (supersedes the line below): D8a′ first** — a prompt
paragraph, mirrored via 099, exposing the council's own 259 minutes; then
**D8c's date-stamp**; then **D8e-1**; with D8e-2/D8e-3 sized only after D8e-1
shows what the index looks like. Original recommendation, retained: *D8a now
(one line, immediate, and it is the seat's own memory), D8c's date-stamp now,
and D8b sized but deferred until D5 says whether any of this is worth staffing.*
D5 has now reported and says it is.

**Against.** Every option here widens what the council can query, and the
prompts are already long; more corpus may mean more rounds, and round count is
what makes the gate cost ~30 minutes. D8d is a legitimate answer if the corpus
is judged a human instrument by design.

---

### D9 — landmines as a footprinted corpus — DRAFTED BY ANOTHER THREAD, awaiting your fold-in

**Not written by this thread.** Session *"bugfix 61"* drafted it 2026-07-27 at the
owner's direction and deliberately did not edit this document beyond this pointer,
because it was being actively edited at the time. Full text:
`architecture_review/PROPOSAL_D9_landmines_as_a_footprinted_corpus.md`.

In one line: **D8's defect has a second symptom.** Knowledge that cannot be queried
must be broadcast, so landmines pile into the auto-loaded `MEMORY.md` — measured at 76
entries / mean 223 chars, compacted twice inside one hour on 07-26 and re-inflated past
its starting size in 46 minutes. The proposal reuses `doc_notes` (370 rows, already
carrying landmine-shaped categories such as `do-not-lock-derived`) keyed on the guarded
path/symbol, and revives **D8a's original "add `doc_notes` to the schema hint"** — which
D8a′ superseded only for the *minutes* case, leaving the landmine case open. It is
explicit that delivery to a *session* (as opposed to a seat) is unsolved, and that
draining `MEMORY.md` before that is built would remove protection we have today.
Distinct from `bugs_open/108`, which fixes the *code* index; this is the prose corpus
`code_checks` will never cover by design.

## 6. Dependency this proposal has

**"Sufficient for anticipated plans" requires the anticipated plans to be
readable by the reviewer.** They are currently spread across ~40 workstream
directories, `features_open/`, `MEMORY.md` and doc 028. A reviewer with no
roadmap artefact to read will invent one, confidently — which is the failure
mode CLAUDE.md's diagnosis section already documents. **Any version of D1/D2
depends on there being a single roadmap document the seat is pointed at.** That
may be the real first task.

## 7. What I could not verify

- Whether the ossification hypothesis holds — **[INFERRED]** only, see D5.
- How often D4's mechanical signal would fire on real history — **[UNMEASURED]**.
- Whether `experience-planner`'s `feasibility`/`mvp` seats are transferable to
  platform code, or are experience-specific — **[ASSUMED]** transferable;
  their prompts were not read.
- The 003/030/086 veto record is from the workstream docs and `doc_notes`
  headlines; individual verdict bodies were not read in full.

---

## 8. Evidence from a live near-miss — added 2026-07-27 by the gripper-dossier thread

*Appended, not restructured; §§1–7 are untouched. Owner directed this thread to
feed evidence here and leave the seat question with you.*

**The incident.** On 2026-07-24 the gripper-dossier design specified a new service
`cmd/gripper-intake/` on the island VM (own DB, own key, own limiter, own CORS).
On 2026-07-25 another thread shipped `cmd/tools-api` to the same VM — multi-tool,
multi-site, with all four already in it. Caught 2026-07-26 by the **owner asking an
integration question**. No mechanism caught it. Full write-up:
`WRONG_CALLS.md` 2026-07-27; correction in
`robot_hands_gripper_dossier/DESIGN_2026-07-24_…md` §2.

### 8a. D4 is no longer `[UNMEASURED]`

§5's D4 says the mechanical RFC trigger's fire rate on real history *"has not been
computed; worth doing before building it."* Computed, for the **doc-surface**
variant — a staged `.md` adding a line that names a `cmd/<x>/` absent from `cmd/`:

```
git show --format= --unified=0 --diff-filter=AM <sha> -- '*.md'   # added lines only
commits scanned: 1500   firing: 10   rate: 0.67%   (2026-07-19 → 07-27)
  12fa24e6b 07-24 gripper workstream opened   -> gripper-intake   ← the incident, day one
  e9fb8a174 07-25 · ce97c8bca 07-25 · 79fd07caa 07-26 -> gripper-intake
  9658d3921 · af07067df · fc0652ce8 · d7b8f34d9  07-20 -> assembler
```

0.67% sits inside `pattern-check.py`'s accepted band (SUMMARY 2.0%, README 0.7%).

Two cautions for whoever builds D4. **A whole-tree scan is not the same predicate**
— it fires on ~190 docs, almost all archived copies naming the retired `cmd/bundle`;
only the staged-diff form is usable. And **"new compose service" and "new route
prefix" are not buildable** from what a pre-commit hook sees: compose files live
under `docs/`, and three router idioms (gin, gorilla, stdlib) are live in this tree
simultaneously.

**The property that justifies it is not the first fire.** On 07-24 `tools-api` did
not exist, so the peer list would have been correct and the author right to proceed.
It is that the check is **free and idempotent**, so it re-fires on 07-25 and 07-26
with `cmd/tools-api` newly in the peer list. Recommend printing peers annotated by
recency — the failure mode is a peer that arrived *after* you looked.

### 8b. D8 is worse than stated — the Go tier is broken too, not merely narrow

§5 D8 records that `code_checks` reaches "Go only — 4,535 symbols, 0 markdown."
Both halves of that sentence understate it. Filed as `bugs_open/108`:

- **`content` never contains function bodies.** `composeSymbolContent`
  (`code_symbols_actions.go:336-352`) builds it from `kind+symbol+signature+doc+path`.
  Live, the entire `content` of one row is three lines: `func init` / `func init()` /
  the path. So every `content` check for a route, registry key, table name or string
  literal returns zero. Verified: `'%stop_reason%'`, `'%/api/v1/tools/gauntlet%'`,
  `'%med_export_json%'` → **0, 0, 0**. The contract at
  `diagnose_code_lookup_action.go:29-31` promises body matching and **its own
  documented example cannot work.**
- **The freshness guard reports FRESH while 667 commits behind.**
  `codeIndexFreshness` computes age from `updated_at`. Live: rows written
  2026-07-26 13:36 (→ FRESH) describing commit `e19aa5d` from 07-24. The index
  tracks a **pushed** ref; `origin/086_experience_loop` has not moved since 07-24
  while the branch gained 667 commits. `internal/tools-api/` therefore has **zero
  rows** — the service the design duplicated was invisible to the very index a
  prior-art check would query.

**Consequence for D8's own reasoning:** D8a (add `doc_notes` to the schema hint) and
D8b (index markdown) both assume the Go tier works and only the corpus is missing.
It does not work. **A forward architecture seat built on this index would inherit an
instrument that manufactures absence in the direction that approves the plan.**
`bugs_open/108` candidate 2 (index bodies from the `[line_start, line_end]` span
already stored) also settles D8b's schema question as a side effect.

### 8c. Evidence bearing on D1 and D3 — do *not* seat duplicate-capability

This incident is the strongest case yet for D1's "no second brake", but it also
argues that **duplicate-capability should not be a seat at all**:

- "Does this already exist?" is **factual**, and `pattern-check.py`'s founding
  doctrine is *"spend the LLM council on judgement, not on what a string comparison
  can settle."* A seat on a factual question is the mistake that file exists to stop.
- Two seats already hold the remit (`reuse_agent`, `prior_art_librarian`). Their
  problem is a broken instrument (8b), not absent judgement.
- **No seat would have caught this one anyway.** On 07-24 the absence was real; any
  reviewer would have approved correctly on the evidence. Reserve the seat budget
  for D1/D2's forward-fitness remit, which is genuinely judgement.

### 8d. The gap §6 names, restated from the incident

§6 says the seat needs a readable roadmap. The incident points at something
narrower and more tractable:

> **A decision with architectural consequence was recorded in the one medium no
> mechanism reads, and the mechanism that would object refuses that medium by
> design** (`097_TRIGGER_council_review_v1.sh:53`, `SCOPE_RE='^(platform|internal|pkg)/'`).

That refusal is *correct* — 72 DESIGN/PLAN/SPEC docs were created in
`docs024_key_docs_latest` in July, and reviewing them would cost real credits. The
route around it is not to widen the gate's scope but to let a **grep** read the docs
and hand the council nothing. Recommend against building a capability-ledger table:
`ls cmd/` cannot drift, and D8 already showed that anything requiring manual upkeep
goes stale silently (the `bug_historian`'s hard-coded seven).

### 8e. A doctrine offered for ratification, if the owner wants one sentence

> **Divergence is allowed when it is parameterised and forbidden when it is copied.**
> A second implementation is fine as a row in a table or a profile; it is not fine as
> a second copy of the code.

Generalises `vm_estate`'s *"merge the generator, not the trust boundary."* Worked
example: `med_export_json` (`registry.go:1691`) sits **ten lines above**
`directory_export_json` (`:1701`), whose own header reads *"nothing site-specific
may be hardcoded here."* Both registered, both live, nobody saw it.
