# RFC 055 — completion gating has no extension point, and gate 1c is the fourth hand-wired one

**Status: DRAFT, raised 2026-08-25 by the `vigilant_designer_offer_analysis` lane.
RETROSPECTIVE, in the shape RFC_002 established: the change is committed (`69479bcf6`,
`Council-Reviewed: 064841bd-58fc-46a1-a77d-6b0a6309d0ba`) and inert until the next fleet roll.**

It is here because the council's **architecture** seat emitted `ARCHITECTURE_SIGNAL` on that
round and said, in terms, that this **should be named rather than absorbed**:

> *"This plan is the THIRD/FOURTH independent gate bolted onto `complete_work_item_verification.go`'s
> shared completion pipeline — gate 1a, gate 1b/`noChangeGates`, gate 2/`RegisterVerifier` registry,
> now gate 1c/`acceptancePredicateGates` — each with its own roster, its own three-valued outcome
> enum, and its own hand-wired promotion-guard lockstep test. There is no general 'completion policy'
> extension point; every new check re-derives the same shape."*

The `reuse_agent` seat objected at MEDIUM to the same shape independently, calling it *"the same
shape as the tool-creation novel/fork duplication the platform already regrets"*. **Two seats
reaching the same conclusion by different routes is the condition the 2026-07-29 ruling was written
for**, so this is filed rather than answered in a commit message.

Neither seat vetoed, and the verdict was APPROVED. **This RFC does not ask to reverse anything.**

---

## 1. What the thing IS, before any rule about it

A **work item** is one recorded defect on one page. A **handler** is the agent sent to repair it.
When the handler's saga returns successfully, the dispatch loop calls `complete_work_item`, and
`CompleteWorkItemAction` stamps the row `complete`.

A **completion gate** is a check that runs in the moment between "the handler came back" and "the
row says complete", and can refuse. There are now four, all inside one function
(`verifyBeforeComplete`), all added by different lanes at different times:

| gate | asks | added for | its roster |
|---|---|---|---|
| 1 | did the handler's own saga report FAILURE? | `bugs_open/017` | none — universal |
| 1b | did the handler report it changed NOTHING? | `bugs_open/213` D1 | `noChangeGates` |
| 1c | is the item's OWN stated criterion still false? | `bugs_open/395` | `acceptancePredicateGates` |
| 2 | does the TYPE's registered verifier still see the defect? | `bugs_open/017` | `RegisterVerifier` registry |

Three of the four are **opt-in per `item_type`**, and each opt-in was built separately.

## 2. The observation, stated as narrowly as it is true

**The four gates are not redundant — they ask genuinely different questions, and none subsumes
another.** That is worth saying first, because the objection is easy to over-read. Gate 1b needs
the handler's REPLY, which only this position in the code has. Gate 2 needs a type-scoped
predicate. Gate 1c needs the item's own spec plus current page state. Merging them would be wrong.

What repeats is not the question — it is the **WIRING**. Each new gate independently invents:

1. an opt-in roster `map[string]someRule`, keyed on `item_type`;
2. a rule struct carrying its own evidence fields (`Why`, `UnreadableWhy`, `RefusalWhy`,
   `PromotionOwes`, `LicenceVoided`, `MeasuredAgainstHandler`…);
3. a three-valued outcome enum whose zero value is deliberately not a policy;
4. a roster test asserting each entry carries the evidence its declaration implies;
5. a hand-written clause in `TestClaimTimeoutExclusionCoversBothCompletionGates`, because the
   claimed-item-timeout sweep writes rows directly and **no gate runs for it**;
6. a distinct `status` string plus a branch in `blockedCompletionReason`;
7. optionally, an `agent_error_log` code and a `finding_code_registry.json` declaration for the
   case where the gate goes blind.

Gate 1c reproduced **all seven**. Items 1–4 are near-identical to gate 1b's by design — I copied
them deliberately, because matching the house idiom is better than inventing a fifth one — and
that is exactly the seat's point: **the correct local decision produces the accumulation.**

⚠ **Item 5 is the one with teeth**, and it is the reason this is an architecture question rather
than a tidiness one. The lockstep contract is now:

```
excluded  ⇔  (has a registered verifier)
             OR (has a noChangeGates entry)
             OR (has an acceptancePredicateGates entry that REFUSES)
```

That disjunction grows by one clause per gate, **and each clause has its own subtlety** — gate 1c's
counts only its REFUSING entries, because a recording entry blocks nothing and would trip the
reverse direction. A fifth gate adds a fourth clause with a fourth subtlety, in a test whose whole
job is to be the thing nobody has to remember. `bugs_closed/317` exists because that contract was
once stated one clause too narrow.

## 3. What is NOT the problem

- **It is not that the gates are unsafe.** Each is opt-in with the unsafe default OFF, each roster
  is small and evidenced, and the accumulated behaviour on any type that has not opted in is
  byte-identical to 2026-08-01.
- **It is not that they should be one gate.** See §2 — the questions differ in what evidence they
  can even see.
- **It is not urgent.** The cost is coordination, and it grows with the number of gates (4), not
  with traffic.

## 4. The question for the owner

**Is a general "completion policy" extension point worth building, and if so, when?**

The shape it would take, sketched only far enough to be argued about:

```go
type CompletionPolicy interface {
    Name() string                       // the status string and the blocked-reason branch
    Grades(itemType string) bool        // the opt-in roster, per policy
    CanBlock(itemType string) bool      // what the claim-timeout lockstep reads — ONE clause, for ever
    Evaluate(ctx, GateInput) GateResult // spec + handler reply + db, so no gate is starved of evidence
}
```

The honest costs, because this estate has been bitten by premature shared mechanisms before:

- **Four is a small N**, and a framework derived from four instances tends to fit the fourth.
- **`GateInput` would have to carry the union of what all gates need** — the handler's reply, the
  spec, the site and page ids, a DB handle — which hands every future gate more authority than it
  asked for. That is the accumulation RFC_022's optional-key budget counts, one level up.
- **The evidence fields do NOT generalise.** `UnreadableWhy` licenses a claim about payload shapes;
  `PromotionOwes` licenses a claim about a missing negative control. Flattening them into one
  `Why string` would delete the distinction that makes each roster test worth running — and those
  distinctions are the most valuable thing in either file.
- Against all that: **item 5 does generalise, exactly and without loss.** `CanBlock` is one method,
  and it collapses the growing disjunction to a single clause permanently.

**So there is a cheap partial available that is NOT the full interface**, and it is what this lane
would recommend if asked: extract only the claim-timeout contract — a registry of "things that can
refuse a completion, per item_type" that each gate registers into — and leave the rosters, the rule
structs and the outcome enums exactly where they are, per gate, with their own evidence fields.
That takes the one item with teeth and pays none of the four costs above.

## 5. The secondary point the same seat raised, recorded because it is a real reading

> *"`content_rewrite` is an EXISTING, high-volume item type (1,637+ historical completions). Arming
> it — even at record-only — is a behaviour change for an existing caller, which per the RFC_022
> ruling text falls back to 'architecture-scope exactly as before' rather than the narrow opt-in
> exception (whose third condition wants zero named consumers, not a same-commit first consumer)."*

**This is a fair reading of RFC_022's third condition and this lane does not contest it.** The
narrow exception requires that *zero live consumers name the field*; gate 1c names `content_rewrite`
in the commit that ships it, so the exception does not cleanly apply and the change is
architecture-scope by the letter.

What is offered against it, for the owner to weigh rather than as a rebuttal:

- the armed outcome is `predicateRecords`, which **refuses nothing** — the only observable change on
  a `content_rewrite` completion is an extra key under `result._verification`;
- it is inert on any item carrying no `acceptance_predicate`, which `[MEASURED 2026-08-25]` is
  **1,637 of the 1,638** completions of that type — exactly ONE `content_rewrite` completion in
  the estate's whole history carries one, and it is `bugs_open/395`'s worked case. (Written first
  as "1,635 of 1,638" by subtracting all three live predicates before checking; two of the three
  sit on `wont_fix` rows, not completions. The query is
  `count(*) FILTER (WHERE spec ? 'acceptance_predicate')` over the live table UNION the archive,
  filtered to `status='complete'` — a figure worth re-running rather than deriving, since it moves
  every time the producer runs.);
- one page-metadata `SELECT` is added per armed item that carries a predicate — **3 rows today**.

⚠ **But "the blast radius is small" is not the same as "the exception applies", and the seat was
objecting to the second.** Recorded as its own question so it is not settled by the size of the
first: **does arming an EXISTING type at a non-blocking outcome count as naming a live consumer for
RFC_022's purposes?** If yes, the rule as written has no way to ship a first consumer at all, which
is worth knowing either way.

## 6. What happens if nobody answers this

Nothing breaks. Gate 1c is live, opt-in and non-blocking. The next lane that needs a completion gate
copies gate 1b or gate 1c, gets a correct result, and adds a fifth clause to the lockstep
disjunction. **That is the outcome this RFC exists to make visible, not to prevent** — and the §4
partial is available to whoever wants to stop it cheaply.

## 7. Sources

- `platform/orchestration/actions/complete_work_item_verification.go` (all four gates in one function)
- `platform/orchestration/actions/complete_work_item_no_change.go` (gate 1b's roster and its evidence fields)
- `platform/orchestration/actions/complete_work_item_acceptance_predicate.go` (gate 1c)
- `platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go` (the growing disjunction)
- council report, correlation `064841bd-58fc-46a1-a77d-6b0a6309d0ba` — `architecture` and
  `reuse_agent` seats, both MEDIUM, verdict APPROVED
- register **WII-033**; `bugs_open/395`; `bugs_closed/317` (why item 5 has teeth)
- RFC_022 (the narrowing §5 is measured against)
