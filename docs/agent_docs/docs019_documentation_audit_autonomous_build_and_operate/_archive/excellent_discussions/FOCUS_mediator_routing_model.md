# FOCUS — Mediator Routing Model: Change → Consultees

**Status:** exploratory design. Companion to `FOCUS_self_development_coding_pipeline_reasoning.md` (coordination positions A/B/C, the mediator) and `FOCUS_best_practice_doc_tree.md` (the atomic standards and their metadata). This doc covers how the mediator decides, for a given change, which standards apply and which area-owners and concern agents to consult.

---

## 1. Premise: the doc tree's metadata is the routing table

Routing is not a separate intelligence. The standard atom's `applies_to`, `scope`, `severity`, `kind`, and `check` fields are the routing table; routing is matching a change descriptor against those tags, plus a trigger policy and a single arbiter. The intelligence is front-loaded into the tree's metadata, which is why the two designs are one system. Getting the tags right (see the doc-tree FOCUS) is what keeps routing mechanical.

---

## 2. Describing the change

A change is reduced to a descriptor: `{ change_types: [...], areas: [...], touched_subsystems: [...] }`.

- **From a concrete diff (precise, mostly deterministic):** touched paths → change types via globs.
  - `actions/*.go` → `go_action`
  - file containing `CREATE`/`ALTER`, or `migrations/*.sql` → `sql_migration`
  - workflow JSON in `default_config` → `workflow_json`
  - `agent_definitions` insert/update → `agent_definition`
  - `adapters/*` → `adapter`
  - frontend templates → `frontend_component`
  - (vocabulary finalized against the real change-type set — deferred item in the doc-tree FOCUS)
- **Touched paths → areas:** each objective-tree area declares the paths it owns; path→area is the reverse of that ownership lookup.
- **Before a diff exists (planning time):** a cheap classifier LLM maps intent → likely change types + areas, so the mediator can route ahead and assemble builder context. Once a diff exists, re-derive precisely from paths.

### Routing runs twice per change
- **Pre** (from intent, possibly): select standards + area context to compose the builder's prompt.
- **Post** (from the actual diff): select reviewers for the produced change.
Same manifest, two invocations.

---

## 3. Selecting standards (one manifest query)

Select every `active` standard where any of:
- `scope = constitution`, or
- `applies_to` ∩ `change.change_types` ≠ ∅, or
- `id` ∈ union(`standing_concerns`) for the areas in `change.areas`.

This set is the "change-type bundle" from the doc-tree FOCUS — the doc tree and the routing query are the same index read two ways.

---

## 4. Selecting who acts on each matched standard

Keyed off the atom's own fields, so the routing decision is a property of the data:

| Atom shape | Action | Tier |
|---|---|---|
| `check` set (deterministic) | Run that validator action in the write/validate loop | Cheap |
| `kind = reference` | Compose into the builder's prompt | Cheap |
| `kind = rule`, `check = null` (judgement-only) | Consult the concern agent for its `concern`, if one exists | Expensive |
| (per touched area) | Spawn the area-owner agent fresh: breaks-my-area / fits-objectives / better-approach | Expensive |

---

## 5. Two tiers (fan-out control for frequent mediation)

The expectation is frequent mediation. That is fine because most of it is the cheap tier.

- **Cheap tier — every change, no spawns:** constitution + matched `reference` standards composed into the builder's context; deterministic validators (`check`s) run in the loop. A local, single-area, low-risk change pays only this.
- **Expensive tier — spawns area-owners + judgement concern agents:** fires only when a trigger says so (§6).

So "frequent mediation" does not mean "every change wakes eight agents." It means the cheap tier runs constantly and the expensive tier runs selectively.

---

## 6. Trigger policy (cheap → expensive)

Escalate to the expensive tier when any of:
- the change touches more than one area, or
- it touches a high-risk subsystem (`messaging-and-contracts`, `data-and-schema`, the deploy path), or
- a matched `blocker` rule is judgement-only (`check = null`), or
- a deterministic validator failed in a way that needs judgement to resolve, not a mechanical fix.

This policy is the part most likely to need tuning against real change traffic (see §10).

---

## 7. Conflict resolution (one arbiter)

The mediator collects advisory verdicts and decides. It does not delegate the decision.

Inputs:
- **Area-owners:** breaks-my-area? / fits-objectives? / better-approach?
- **Concern agents:** which standard is violated, at what severity? / improvement suggestion?
- **Validators:** pass/fail + errors.

Decision rule:
- Any `blocker` violation, or any owner reporting "breaks my area" → send back to the builder with that specific feedback; loop.
- Only `should` / `advisory` flags or improvement suggestions → do not block; the mediator either applies now or raises follow-up work items (the side-effect-item pattern, for the non-blocking async tail).
- Anything the mediator cannot resolve, or a high-severity judgement call → escalate to a human at the merge gate.

Concern agents advise; they never arbitrate. "Who has final say" stays in exactly one place.

---

## 8. Reuse vs new

**Reuses what exists:**
- change-type → consultee routing is the existing `item_type` → handler routing with a richer descriptor.
- spawn-fresh consultees reuse the spawn machinery and the `coordinator` agent_category.
- matched-standards-as-context reuses the prompt-composition pattern.
- non-blocking follow-ups reuse side-effect work items.
- validators-in-loop reuse the existing validate → regenerate loop.

**Genuinely new:**
- the change classifier (paths/intent → descriptor).
- the trigger policy.

---

## 9. The fragility to design against

The path→change_type globs and the `applies_to` vocabulary are the coupling point between the repo's real layout and routing. If they drift, routing misfires silently — loads the wrong standards or misses a concern. Treat the glob map as a tested artifact, not hand-edited config. Checks (the routing analogue of the doc-tree drift check):
- every owned code path resolves to exactly one area, and
- every change type has at least one path glob.

---

## 10. Deferred / open

- **Trigger policy tuning (§6):** the cheap→expensive thresholds want validation against real change traffic before they're trusted. Likely the first thing to adjust in practice.
- **`applies_to` vocabulary (§2):** finalize against the real change-type set; shared deferred item with the doc-tree FOCUS.
- **Pre-routing accuracy:** how good the intent classifier needs to be before a diff exists, vs deferring all routing to post-diff and accepting a less-informed builder context.

---

## 11. One-line state

Routing = describe the change (paths→types+areas), query the manifest for matching standards, act on each by its own metadata (validator / prompt context / concern agent), run a cheap tier always and an expensive tier on trigger, arbitrate in the mediator with humans only at the gate. Mechanical by design, because the doc tree carries the intelligence.
