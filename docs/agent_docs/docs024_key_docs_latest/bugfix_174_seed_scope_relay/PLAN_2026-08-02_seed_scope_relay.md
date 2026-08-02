# PLAN — bugs_open/174: the diagnosis dispatch loop drops `seed_scope`

**Started** 2026-08-02 by the `bugfix_174_seed_scope_relay` lane.
**Ticket:** `bugs_open/174_HANDOFF_2026-08-01_dispatch_loop_drops_seed_scope_so_a_targeted_diagnosis_silently_becomes_an_untargeted_one.md`
(filed by the `bugfix_164` lane, unowned when picked up).

## The problem, as filed

A diagnosis fired with `SEED_SCOPE` naming specific symbols silently ran against
whatever the code search happened to return. Measured by the filing lane on
`site_work_items`: of four intakes that ever carried a non-empty `seed_scope`,
**three** were claimed by `diagnose-dispatch-loop` and lost it. Two of the three
are other lanes' real work (`bugs_open/155`, and a scheduler diagnosis from
07-28) — both aimed at chosen symbols, both silently re-aimed, **and neither
author had any way to know**.

The ticket's diagnosis: `diagnose-orchestrator` forwards `seed_scope`, but the
`diagnose-dispatch-loop` in front of it does not, and `input_mapping` is an
allow-list so the key is dropped in silence.

## CORRECTION to the filing — there are THREE gates, not one

> **This is the substance of this lane's work, and it was found by reading the
> path rather than by trusting the ticket.** The ticket's fix candidate 1 —
> "add `seed_scope?` and `runtime_page?` to `call_handler`'s `input_mapping`,
> sourced from `claimed.seed_scope`" — **would not have worked, and would have
> failed silently**, which is the same failure mode all over again.

| # | gate | why it drops the key | fixed by |
|---|---|---|---|
| 1 | `claim_item`'s SQL `RETURNING` clause | It is **also** an allow-list. It projects nine spec keys; `seed_scope` is not among them, so `claimed.seed_scope` **does not exist** for any mapping to read. The ticket did not name this. | migration 289 (config) |
| 2 | `call_handler`'s `input_mapping` | An allow-list, not a passthrough. An unlisted key is skipped at Info level. This is the gate the ticket named. | migration 289 (config) |
| 3 | **type** | `QueryDatabaseAction` scans every column into `interface{}` and stringifies any `[]byte` (`database_actions.go`). A jsonb column therefore reaches collected_data as the **string** `["a","b"]`. `ResolveInputMapping` does no coercion. `ExtractStringListHelper` returned **nil** for a string — indistinguishable from "the caller supplied nothing". | `data_helpers.go` (Go) |

Gate 1 makes candidate 1 map from a path that resolves to nothing; the key being
optional, it is dropped without a word. Gate 3 then drops it a third time even
once gates 1 and 2 are open. **All three had to move together.**

## Why it stayed invisible for three lanes

`diagnose_assemble_bundle`'s scope resolution is a **fallback chain by design**
(`:135-151`): loop scope → `input_data.seed_scope` → `code_results`. With the
seed confiscated, arm 2 finds nothing and arm 3 quietly supplies a *different,
plausible* scope. The action genuinely cannot tell "the caller gave no seed"
(correct, and the common case) from "the seed was taken in transit" (the bug),
so it correctly does not complain.

**A fallback chain converts a lost parameter into a successful run with
different inputs.** That is the transferable finding, and it is why fix 3 below
matters more than its size suggests.

## What was built

1. **Migration 289** (`sql_for_agents/289_...sql`, applied 2026-08-02 ~11:15,
   snapshot `f4055640`). Gates 1 and 2 together. Its final assertion checks the
   **invariant**, not the two keys: every key `diagnose-orchestrator`'s
   `input_contract` declares must be forwardable by the loop.
2. **`ExtractStringListHelper` accepts JSON-array text** (gate 3). Widening
   only — a string returned nil before, so no existing caller's answer can move.
   Correct whichever shape the driver returns, which is why no caller has to
   know what pgx does with jsonb.
3. **Scope provenance** in `diagnose_assemble_bundle`. `scope_source`
   (`route` | `seed` | `code_results`) on the result every run; a bundle note
   **only** on the ambiguous `code_results` arm, so seed/route bundles stay
   byte-identical and no archived baseline moves.
4. **`config-key-audit --relay-gaps`** + `scripts/audit-relay-gaps.sh` — the
   class-closing check. See the design note below, which is the part worth
   reading.

## Design decision: why the class-closing check is a REGISTRY, not a fleet rule

The ticket's fix candidate 2 asked for a lockstep test asserting "every key the
orchestrator accepts is forwardable by the loop". I wrote the **general** version
first and **measured** it before committing to it. It is not sound:

| version | findings | verdict |
|---|---|---|
| "every `call_agent` must forward every key its callee declares" | **31** of 75 resolvable call sites | Noise. Spot-checked `pageflow-builder.apply_site_design` → `webdesign-agent` omits `site_context`, and the callee has an explicit `else_step: load_site_context` to load it itself. Legitimate. |
| tightened: "…and the callee actually READS `input_data.<key>`" | **3** | Still cannot tell "the caller dropped it" from "the caller never had it" — which is the entire question. |
| **both versions** | — | **BLIND TO 174 ITSELF.** `call_handler` resolves its callee at runtime (`agent_type_field: claimed.handler_agent`), so a static resolver skips the one site the check exists for. |

That last row is the "narrowing a detector can make it inert" trap, and it would
have shipped a check that passed for ever on the bug that motivated it.

**So the check asserts where the question is answerable.** For a *dispatcher*,
the caller's envelope is not "whatever is in collected_data" — it is the work
item spec, whose shape is exactly what the handler's contract declares. Population
measured live: **3 dispatcher-shaped relays fleet-wide.** One is registered and
asserted; the other two are reported as **uncovered** rather than assumed fine,
because registering them would mean asserting something nobody has read.

The tool reports three different kinds of thing, deliberately kept apart:

- **findings** — a registered relay that cannot carry a declared key. A defect.
- **unmatched registry entries** — a relay we claim to check that no longer
  matches live config: *an assertion that stopped running*. **Worse than a
  finding**, and it exits 1 too. This earned its place on the very first live
  run (see NOTES).
- **uncovered relays** — dispatcher-shaped and unregistered. Advisory only.

## Deliberately NOT fixed, and the measurement that justifies it

`QueryDatabaseAction` stringifying every jsonb column is the **deeper** cause.
It is not fixed here: changing it alters the shape every consumer receives, and
that is a shared-mechanism change.

> **CORRECTED 2026-08-02 — the blast-radius figure I gave the council was wrong,
> in the cautious direction.** The submission said "14 live `query_database`
> steps project json/jsonb". That came from a loose regex over the whole query
> text, which also matched `->>` text casts and `->` inside WHERE predicates.
> **Re-measured properly: exactly ONE live step projects a JSON-typed value, and
> it is the one this fix added.** Every other one of the 14 either casts with
> `->>` (so its consumer already expects text) or uses the arrow in a predicate,
> which is not output at all. Verified by reading all three ambiguous queries in
> full, not by trusting the tightened filter.

So the deferred fix has **zero currently-affected consumers**. That makes the
trap genuinely *prospective* — a landmine for the next author who projects a
jsonb value expecting an object — which is exactly what `LANDMINES.md` is for,
and it is recorded there.

## Council

Submitted **before** committing, correlation `081d98b3-75e1-4926-a17a-b0c72e5ccece`.
**APPROVED at round 1**, 6 advisory objections, none high-severity. Four were
answered with work rather than prose — see NOTES § "what the council caught".
