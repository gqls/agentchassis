# PLAN — bugs_open/172, the agent-state gather's two caps

**Opened** 2026-08-02. **Lane** `bugfix_172_agent_state_cap`.
**Target** `platform/orchestration/actions/diagnose_load_runtime_action.go`,
`gatherAgentState`.

## What we are trying to do

Make both caps in the diagnosis bundle's agent-state section report what they
discard, and make the survivors reproducible. The section exists so the diagnosis
loop's verdicter can reason about **which agent misbehaved** — so an agent that is
silently absent is read as an agent that did nothing, and that is a wrong-answer
mechanism, not an untidiness.

## How the scope changed on contact — the important part

The ticket names **one** cap: `matched = matched[:typeCap]`, measured at filing as
**latent** (max 4 agent types ever listed, default cap 5).

Sizing the fix meant measuring the section as it actually renders, and that found a
**second cap in the same function that is not latent**. The `llm_call_log` gather
issued ONE query for ALL matched types under a single shared
`ORDER BY created_at DESC LIMIT n`. Rows are therefore allocated by **global
recency across the whole named set**, so the chattiest agent takes the entire
budget and every quieter named agent renders zero lines.

**Measured 2026-08-02** over the 72 retained bundles carrying this section: of the
23 that named more than one type *and* returned any rows, **every single one shows
exactly ONE agent type** in its log lines (10 named 4 types, 10 named 3, 3 named 2).

So the ticket's own framing — "latent, one short of firing" — was **true of the cap
it named and false of the function**. The decision was to fix both, because they
are one defect class in one gather, and fixing only the named one would leave the
firing one behind a file that now looks audited. This is the third time this loop
family has been audited by the shape someone happened to grep for
(`bd003f67a` → `164` → this); narrowing again would have been the fourth.

## Design decisions, and why

1. **`ORDER BY type` on the DISTINCT listing.** The cap slices the *tail*, and
   `matchAgentTypes` preserves input order, so without a sort the survivors were
   whatever the plan produced. Two identical runs could gather different agents,
   both reporting success. Determinism is also what makes the cap notice checkable.
2. **Name the dropped types, don't just count them.** The sibling marker at
   `diagnose_assemble_bundle_action.go:328` can only count; here the casualties are
   known strings, and naming them is what lets a verdicter re-ask for one.
3. **The heading is CONDITIONAL.** It counts kept-vs-named only when the cap fires.
   Deliberate trade: the ticket asks for a byte-identical negative control below the
   cap, and always counting would move every existing bundle's baseline. When
   nothing was dropped the original sentence is simply true.
4. **Per-type allocation, not a bigger shared limit.** Raising the limit moves the
   cliff without making it visible — the candidate `145` was refused for. A window
   function makes starvation-by-another-agent's-chattiness **unrepresentable**,
   which is the ordering rule from `order-fix-candidates-by-what-closes-the-door`.
5. **"No rows" and "at the cap" are worded differently on purpose.** A type with no
   rows is an ANSWER about the table and is safe to reason from; a type that filled
   its budget has older calls that were not gathered. Collapsing them is exactly the
   mistake `bd003f67a` corrected at the sibling cap.
6. **No shared helper.** A "cap and report" helper across this family would be a
   platform seam, and a seam arriving inside a bug patch is what the guardian seat
   vetoed `bugs_closed/124` for. The two new functions are file-local and pure.

## Phasing

- [x] Verify the ticket is still valid, and that no live session owns it
- [x] Measure the section as rendered (found the second, firing cap)
- [x] File the structural claim with the diagnosis loop **before** asserting it
- [x] Fix both halves; reuse existing marker conventions
- [x] Induce the type cap (it cannot fire in production) + mutation-test the guards
- [x] Council gate: `d47b826e-6fc6-42ad-a2ef-62b1f1ba0b88`
- [x] Commit (`3761a04ca`) with `Council-Submitted:`
- [ ] Live on a rolled chassis + pod-verified, **then** close to `bugs_closed/`
