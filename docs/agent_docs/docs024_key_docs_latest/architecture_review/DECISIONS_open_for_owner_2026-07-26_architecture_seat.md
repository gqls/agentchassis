# Decisions open for the owner — an architecture seat / process

**Status: DRAFT — proposals only. Nothing built, no config or code changed.**
Written 2026-07-26 after the owner asked whether we had discussed an
architecture council member before, and asked for the choices and proposed
decisions in one place.

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

### D5 — Measure ossification before staffing anything.

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

### D7 — Open: does the veto survive, and is clause (d) narrowed?

**Not proposed — genuinely the owner's call.** Two sub-questions:

- **(a)** Given three overturns in two days, should the guardian's veto become a
  strong objection that escalates, rather than a block? The counter: the veto is
  what *caused* the RFC track to exist, so it has already earned its keep once.
- **(b)** If a forward seat is added, should clause (d) be narrowed so the
  guardian stops being asked to weigh **benefit** — which it does badly and
  which is not its remit — and judges blast radius and contract-breakage only?

My inclination is (b) yes, (a) no: keep the veto, remove the benefit judgement
from the seat that cannot measure benefit, and let the forward seat supply it.

---

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
