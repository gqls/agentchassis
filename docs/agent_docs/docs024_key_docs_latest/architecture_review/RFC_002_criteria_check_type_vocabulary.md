# RFC 002 — Who may add a check type to the shared criteria vocabulary, and on what terms

**Status: RATIFIED 2026-07-29** (owner; three answers in §8, now §9) · raised 2026-07-29 by the experience-register session, **at the owner's
explicit instruction** ("route it to a real architecture review") after the council gate's
`review_architecture` seat ruled on corr `99f2a5e6-e934-4ca1-addb-f16a29b38b0f`.

> **THIS RFC IS RETROSPECTIVE, AND THAT IS THE POINT.** The change it concerns is already
> live (chassis v1.0.1197, pod-verified). This is not a request to build something; it is the
> `bugs_closed/124` shape — *the code stays and the precedent gets fixed*. If the honest
> outcome is "that should not have shipped the way it did", the useful product of this RFC is
> the rule that stops the next one, not a revert.

---

## 1. Problem + evidence

### 1.1 What was added

Two check types, `attribute_absent` and `attribute_matches`, in the Tier 2 static criteria
evaluator (`platform/orchestration/actions/discovery_checks/check_tool_acceptance.go`,
`static_attribute_checks.go`) plus their entries in the validator's capability table
(`platform/orchestration/actions/experience_criteria.go`).

Why it was needed is not in dispute and is not what this RFC asks about. In short: the
experience register's central rule — the `no-inert-control` invariant, *"a control that cannot
do anything must not be presented as a control"*, found independently reinvented in six
implementations across two sites and five authors — was recorded and **could not be checked**.
It blocked 13 of the register's 38 deferred clauses, across 9 of 9 entries. Six entries had
already authored eight checks in this exact shape before any code existed.

Proven live before submission, through the real exported evaluator: 6 checks PASS over 18 real
elements on `fundamentallyai.com` and `vonc.com`, 1 correctly SKIPs, 1 FAILs on a genuinely
served `href="#"` (that last is `bugs_open/137`, not a regression).

### 1.2 What the architecture seat ruled

> *"attribute_absent/attribute_matches are new reserved keys added to experienceCheckTiers, a
> capability table with two read-sites across two systems. Per the plan's own cited 2026-07-28
> seam ruling, a new key on a shared vocabulary is architecture-scope even when additive,
> small, well tested, and measured at zero current collision."*

The `guardian` seat, separately and in the same round, **declined to veto** — *"a constrained,
well-fenced addition rather than a redesign"* — but made the observation this RFC treats as its
sharpest input:

> *"the tool-acceptance pipeline is a second consumer that did not appear to review this
> change, and I want that acknowledged as a live dependency, not just a measured one."*

**Two seats, same round, defensible in opposite directions.** That is the signal that this is a
governance question rather than a technical one, and CLAUDE.md is explicit that such a question
is not answered by resubmitting with better measurements.

### 1.3 The process finding, which is the uncomfortable part

CLAUDE.md's 2026-07-28 seam ruling permits a platform seam to ship ahead of its review only when
**both** hold: (1) a real, stated ordering constraint, and (2) registration in the concept
register in the same commit that ships it.

**Condition (2) was met** (TL-031, same commit). **Condition (1) was not, and I said so in the
submission**: *"There is NO ordering constraint claimed here."* The change shipped anyway,
because on this tree **committing is shipping**: HEAD is shared, `make build-*` builds from
committed HEAD, and another session built and rolled v1.0.1197 carrying these commits. The
verdict arrived after the code was already on HEAD.

So the exemption's first condition is, in practice, **unsatisfiable-by-choice** here: a thread
cannot hold a change out of the fleet by declining to deploy, because it does not control the
deploys. The only mechanism that actually holds a seam back is a **default-OFF switch**, and I
did not build one. `[STATED, not excused]`

## 2. Design — what the RFC proposes

Not a code change. A **rule with a mechanical trigger**, in three parts.

**2.1 A check type is a governed addition.** Adding a key to `experienceCheckTiers` (or a new
`case` to `evaluateStaticCriteria`) requires naming, in the commit, the complete set of
consumers and confirming each has been told. Not "measured" — *told*. The guardian's point is
that a measurement of zero current collisions says nothing about whether the owners of the other
pipeline would have objected.

**2.2 A new refuting type needs a default-OFF switch, or an RFC first.** The distinction that
matters is not "new type" but **does it change what an existing consumer's documents mean**. An
additive type reachable only by opt-in changes nothing for anyone until they opt in; a type that
alters an existing evaluator's *guarantee* does. Attribute assertion did the second thing: it
made Tier 2 refute, where its stated rule was "confirm, never refute".

**2.3 The trigger is already half-built and should be finished.** `TestEveryStaticCheckTypeIsClassified`
(shipped in this change) already fails the build when a type is added without classifying it as
confirming or refuting. Extending it to also fail when a **refuting** type is added without a
`GOVERNANCE:` line naming its RFC turns this rule from a convention into the same thing that
caught it: a build failure at the moment of the decision. That is a small, contained edit and
does not need this RFC to be ratified first — but it should not be written until the owner has
ruled, or it encodes my answer rather than theirs.

## 3. Alternatives considered

**A. Do nothing; treat this as a one-off.** Ruled out by evidence, not taste: the register has
**three further capability gaps already named and waiting** — event-listener assertion, fault
injection at the fetch boundary, per-row conditionals tied to source data. Each is another check
type in the same shared vocabulary. Doing nothing means having this argument three more times,
with the precedent set by whoever moves first.

**B. Require an RFC for every new check type.** Ruled out on cost: an opt-in-only type that no
existing document names, and that cannot change any existing document's outcome, is exactly the
"point fix that happens to be large" the process doc says does *not* need an RFC. Requiring one
would make the process the obstacle rather than the check.

**C. Split the vocabulary — give the experience register its own evaluator.** Ruled out, and
this is the one worth arguing rather than asserting. It would remove the shared-contract
question entirely. It would also create **two definitions of "this check passed"** — and
`exported_static_criteria.go`'s own header already names why that is worse than the problem it
solves: *"a divergent judgement produces confident disagreement rather than an error."* The
current design was chosen precisely to avoid this, and the seat's objection is the price of that
choice, not evidence against it.

**D. Ship it behind a default-OFF switch.** The option I should have taken and did not, and the
only one that would have made the exemption's first condition satisfiable. Its cost is real —
a switch nobody turns on is how a mechanism rots unexercised, which this workstream has already
written down as a hazard — but that cost is a week, and it is the price of the change being
reviewable before it is relied upon.

## 4. Blast radius, named — derived mechanically

**Binaries linking `discovery_checks`** (from `go list -deps` per `cmd/` target):
`agent-chassis`, `core-manager`, `config-key-audit`, `test-spawning`, `workflow-monitor` —
five link it; **one, `agent-chassis`, actually executes criteria**.

**Callers of `evaluateStaticCriteria` — exactly two**, enumerated rather than grepped:
- `check_tool_acceptance.go:212` — the tool-acceptance discovery check. Its criteria come from
  `loadCurrentCriteria`, whose body is `SELECT body FROM doc_plans` + `extractCriteriaFence`.
- `exported_static_criteria.go:53` — reached only from `verify_site_experience_action.go:224`.
  Its criteria come from `experience_patterns.criteria_template`.

Those two tables are therefore the **complete inventory** of documents that reach this switch,
which is what makes the following exhaustive rather than convenient:

| measurement | value |
|---|---|
| criteria fences in `doc_plans` fleet-wide | 78 |
| …containing `attribute_absent` | **0** |
| …containing `attribute_matches` | **0** |
| `experience_patterns` rows using either | **1** |

Both types previously fell to the switch's `default:` branch and were SKIPPED, so **no existing
document changes outcome**. `[VERIFIED by query 2026-07-28; re-run before relying on it — the
count moves as tools are created.]`

**What the measurement does NOT cover, per the guardian:** whether the tool-acceptance
pipeline's owners would have objected. Zero collisions is a fact about documents, not about
consent.

## 5. Staged rollout plan

The change is already live, so this section is the **acceptance watch** rather than a plan:

1. **Shipped** — chassis v1.0.1197, pod-verified with a negative control (`attribute_matches` 5,
   `attribute_absent` 3, `matched no elements in the served HTML` 1, nonsense string 0).
2. **First real exercise** — CC-001 re-seeded 2026-07-29 08:12Z with 2 executable checks. Watch
   whether any tool-acceptance run's outcome changes; the prediction is that none does, because
   no `doc_plans` fence names either type.
3. **The induced-fault tests that matter** are in the tree and were each proved to bite before
   submission: zero-matches-passes fails two tests; an unclassified new type fails the build; a
   per-type field no runner decodes fails the lockstep.

## 6. Rollback plan

Removing both `case` labels and both capability-table entries returns every document to the
`default:` SKIP branch. **No migration is involved and no schema tolerates-or-not question
arises** — migration 264 is a column COMMENT. The previous binary tolerates every stored row
today, because the only entry using the new types stores them as ordinary JSON either way.

The one thing rollback would NOT undo: `bugs_open/137`, the disagreement between the
dead-control sweep and attribute assertion, which the capability *revealed* rather than caused.
That disagreement predates this change and survives its removal.

## 7. Acceptance evidence

- **Already in:** the six live PASSes over 18 real elements; the pod-grep with negative control;
  the re-seed landing at exactly the locally-predicted counts (2 executable, 8 deferred).
- **Owed:** one tool-acceptance sweep after this ships, confirming no fence's outcome moved.
- **Owed:** the owner's ruling on §2, which is the actual product of this RFC.

## 8. The question for the owner, stated plainly

1. Was the architecture seat right that a new key in a shared check vocabulary is
   architecture-scope? (I think it is right about `attribute_*` specifically, because these
   changed Tier 2's *guarantee*, and wrong as a general rule about *any* new key.)
2. Should a change like this have shipped behind a default-OFF switch (alternative D)? On a
   shared HEAD that is the only thing that actually holds a seam back, and I did not do it.
3. Should tool-acceptance's owners be brought in retrospectively, per the guardian?

**Related:** `bugs_closed/124` (the precedent: code stays, precedent gets fixed);
`bugs_open/137`; concept register **TL-031** (the change, its landmines, and the ruling recorded
verbatim); council trail `99f2a5e6-e934-4ca1-addb-f16a29b38b0f`, two rounds.

---

## 9. RATIFIED — the owner's three answers, 2026-07-29

Asked as three plain questions and answered directly. Recorded here, and the two that
change standing practice are written into `CLAUDE.md`'s seam section in the same commit,
because a ruling that lives only in an RFC is a ruling nobody will find.

**Q1 — does a new key in a shared check vocabulary need an RFC? → ONLY WHEN IT CHANGES A
PROMISE.** Not merely because the vocabulary is shared. The distinction the owner adopted
is the one this RFC proposed in §2.2: `attribute_absent`/`attribute_matches` needed review
because they made Tier 2 able to **refute**, where its stated rule was "confirm, never
refute" — a change to what the shared thing guarantees. A type that only adds an opt-in
capability, reachable by nothing until a document names it, goes through the normal
council gate.

*Consequence, immediately:* of the **three capability gaps already queued** —
event-listener assertion, fault injection at the fetch boundary, per-row conditionals —
none obviously changes a guarantee, so on this ruling they are ordinary gated changes.
**But fault injection deserves a second look before anyone assumes that**: serving a
deliberately broken feed is not an assertion, it is a new power over the environment
under test, and "the runner may now break things on purpose" is plausibly a guarantee
change. Whoever builds it should ask, not assume.

**Q2 — should changes like this ship behind a default-OFF switch? → NO. Review here is
after the fact, by design.** This retires condition (1) of the ordering exemption, which
asked for a "real, stated ordering constraint" as the price of shipping ahead of review.
The condition assumed a thread could hold a change back and was choosing not to; on this
tree it cannot. This RFC's own case is the proof — submitted before it was committed,
explicitly disclaiming any ordering constraint, live anyway on another session's build.
The owner weighed the off-switch and declined it: a switch nobody turns on is how a
mechanism rots unexercised, and that cost was judged higher than the review-timing one.

*So the honest rule is now:* register in the same commit, submit to the gate before or
alongside committing, and **do not claim an ordering constraint you do not have.**
§1.3 of this RFC — written as a confession — is retired as a finding: it was describing
a rule that could not be complied with, not a lapse.

**Q3 — tell the tool-acceptance owners retrospectively? → YES.** Measuring zero affected
documents proves nothing breaks; it does not establish consent. Done in the same commit —
a dated note in the travelling-docs workstream, which owns Tier 2 acceptance. The message
is deliberately *"your evaluator can now fail a page for something it serves, where before
it could only confirm"*, not a list of two new keys: the guarantee is the part that
touches them.

**What this RFC does NOT decide, and nobody should read into it:** whether the served
`href="#"` on vonc's archive template is a defect. That is `bugs_open/137`, it is still
open, and it now blocks CC-001 from verifying — which is the correct pressure on it.
