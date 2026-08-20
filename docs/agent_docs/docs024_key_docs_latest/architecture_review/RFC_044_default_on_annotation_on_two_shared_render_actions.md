# RFC_044 — a default-ON annotation on `render_component` and `compile_page_sections` is a shared-contract change, and it arrived inside a bug fix

**Status:** FILED 2026-08-20 by the `bugfix_305_negation_gate` lane. **Needs an owner or architecture
decision.** The code is live-on-next-roll and additive; the question is whether it should be.

**Why this is here rather than in the bug file.** Two council seats objected to the same thing in
round 2 of `c48b7612` — `guardian` at HIGH and `architecture` at medium — and per the owner ruling of
2026-07-28, *"a veto on SCOPE is not answered by resubmitting with better measurements. It is a
judgement about how a capability reached production. Record it where the change lives, route the seam
to architecture review on its own merits."* This is that routing. Verbatim:

> **guardian (HIGH):** *"annotateSectionNegation/annotatePageNegation wrap the GlobalActionRegistry
> entries for render_component and compile_page_sections — not agent-scoped, not workflow-scoped.
> Every pipeline that invokes these two actions (page-build-handler, page-rerender, section-editor,
> css-patch-agent, adopt_verbatim, and any other consumer of the shared render/compile path) now runs
> the negation scanner on every section of every page build, fleet-wide, by default ON, forever — not
> just on page-content-writer's output."*

> **architecture (medium):** *"…adding new result keys other workflows' output_field/logging may come
> to depend on — this is a shared-contract change to two of the most-invoked actions in the platform,
> arriving inside a single-agent bug fix."*

Both are correct as statements of fact. This RFC states the case for and against, and what would have
to be true for each option.

## 1. What was actually done

Two wrappers registered at the `GlobalActionRegistry` entries for `render_component` and
`compile_page_sections` (`platform/orchestration/actions/copy_gate_annotation.go`). They:

- run `datahelpers.ScanContentDataForNegation` over the content the section was rendered from, and
  over each section's `content_data` on the compiled page;
- attach `copy_gate_findings` / `copy_gate_page_hits` + `copy_gate_page_fields` to the result map
  **only when non-empty**;
- pass the inner action's error and non-map results through untouched;
- change nothing else. No refusal, no LLM call, no write, no behaviour change.

**Why default ON was chosen:** the REPAIR is opt-in and today runs on one agent. If the COUNT were
also opt-in, the fleet number would be "whatever the wired agents happen to produce", and *"the copy
improved"* would be indistinguishable from *"the check was not wired here"*. The rule invoked is the
meta-description gate's own (`save_page_meta_description_action.go:44-48`): *"a gate a workflow author
can forget to wire is a comment"*.

**Why RFC_022's narrow exception does NOT cover it, stated plainly:** that exception requires the
unsafe side to be the DEFAULT-OFF one. This is default ON. The exception does not apply and the
architecture seat is right to say so.

## 2. The case that it is not architecture-scope

- It adds no authority. The wrapper cannot change what is rendered, cannot refuse, cannot write.
  Deleting both keys from the result changes no downstream behaviour today.
- It has **zero consumers**. Nothing reads `copy_gate_findings` — measured by grep at the time of
  filing. It is a measurement surface, not a contract anyone is holding.
- The cost is a regex sweep over a content map already in memory (~µs per section).

## 3. The case that it IS, and it is the stronger one

- **"Zero consumers today" is exactly how a contract starts.** RFC_034 records the same seat naming
  this expiry for `InstanceID`: the moment something consumes the key, it becomes load-bearing across
  a shared surface, and that is the PR where the architecture pass belongs. Here the consumer is
  already foreseeable — the verification plan for `bugs_open/305` reads these keys.
- **`ActionInputSpec` does not govern it.** Neither action declares one, so the RFC_022 optional-key
  budget cannot see the key, and nothing counts it toward the accumulation the budget exists to
  notice. It is invisible to the one mechanism that would flag the tenth such key.
- **The blast radius is every page build in the estate**, which is not the blast radius of the bug
  it was filed under.

## 4. Options

**(a) Leave it.** Cheapest, and honest only if this RFC is decided rather than filed and forgotten.
Requires: a register entry (exists, **CQ-026**) and an accepted statement that the key is
observational and unconsumed.

**(b) Give both actions an `ActionInputSpec` and declare the key.** Makes the surface visible to the
optional-key budget, so the tenth key on these actions is noticed by the mechanism built for it.
Costs a declaration of every config key those two actions already read — `render_component` reads
around a dozen, and getting that list wrong is worse than the key (`ActionInputSpec`'s own header:
*"an over-strict validator is a considerably worse bug than the inert key it is chasing"*).

**(c) Make the annotation opt-in per step, default OFF, enabled by the same migration.** Satisfies
RFC_022's exception exactly. **Gives up the property the annotation exists for** — the fleet count
becomes a count of the wired agents, and the estate has a measured record of what happens to
mechanisms nobody remembers to wire (`copy-editor`: live since 2026-08-17, dispatched by nothing).

**(d) Scope the wrapper to the writer's own steps** (wrap only when the step config names it).
Equivalent to (c) with extra machinery.

## 5. What this lane recommends, and what it is NOT deciding

**(a) if the estate accepts that an unconsumed observational key on a shared action is not a
contract; otherwise (b).** Not (c): an unwired counter is the failure this whole bug is an instance
of — a rule that exists and does not hold.

**This lane is not deciding it.** The code is committed and inert until the next roll, so there is
time. What would make the decision concrete: whoever adds the FIRST reader of `copy_gate_findings`
outside this lane closes this RFC or implements (b) in the same commit.

## 6. Cross-links

- `bugs_open/305` §10–§11 · register **CQ-026** · council correlation `c48b7612-3ecc-4345-912e-5966c079cb91` rounds 1–2
- RFC_022 (the narrow exception this does not qualify for) · RFC_034 (the same seat, the same expiry, stated for `InstanceID`)
- owner ruling 2026-07-28 (scope vetoes are routed, not re-argued) · owner ruling 2026-08-02 §2 (new authority ships opt-in default-OFF — cited here for why it does NOT apply: no authority is added)
