# PILOT — "Guidelines Agent" council reviewer (stage 3, seat #3 of the extended roster)

**Status: LIVE as of 2026-07-17.** Applied to `clients_db` via
`fixloop_eg_dartsonline/0NN_fix_proposer_v8_guidelines.sql`. Pre-flight
confirmed no in-flight fix-proposer/council orchestrations and that the live
state was v7 (reuse present, guidelines absent) before overwriting. Prior row
snapshotted (`f9d90a2d-...`). Verified live: `review_guidelines` present,
wired `review_reuse_agent → review_guidelines → review_guardian`, both
`review_fields` arrays carry all 5 reviewers, prompt intact (2,774 chars).
**The council is now 5 sequential reviewers — the last always-on seat before
the relevance-filter.**

---

## 1. Why this seat

FIX-036's own next-named roster member after reuse: "a guidelines agent
(adherence to 000-0xx, or did the guideline fall short)." The two clauses are
both load-bearing and give this seat a shape no other reviewer has:

- **Adherence:** does the proposed fix follow the platform's own documented
  conventions and contracts — the numbered guides (`001_development_guide`,
  `002_system_architecture`, `003_...` etc.) and the hard contracts they
  describe?
- **"Did the guideline fall short":** distinctively, when a fix reveals that a
  documented *rule itself* is wrong or stale, that is a **guideline-gap**, and
  the fix-loop's own design (FIX-016, FIX-042) says it should **lean
  side-task, not block** — the fix is correct; the rule needs amending. This
  is not a hypothetical: `MDL-039` (BUG B, found this week) proved a live
  runbook rule about `max_tokens` placement was literally *backwards*, and the
  `idx_swi_dedup` ↔ `workItemTerminalStatuses` contract-drift (the dedup-index
  memory) is a second live case of a documented contract that a fix must keep
  in lockstep or break the fleet.

## 2. Grounding concepts (the curated context)

Unlike the bug-historian (a single recurring failure family) and the
reuse-agent (a single discipline), the guidelines-agent's context is a small
set of the platform's most-rediscovered *rules* — the ones the register shows
people keep having to relearn:

| Concept | Category | The rule |
|---|---|---|
| `DEV-005` | development-guide | Wrapper-orchestrator pattern: anything doing substantive work (LLM calls, crawls, heavy DB) needs a spawned pod via a parent, not the shared chassis slots. Confirmed independently across 6+ workstreams. |
| `DEV-027` | development-guide | Work-item dedup: `idx_swi_dedup` UNIQUE (site_id, item_key) over non-terminal statuses; use DELETE+INSERT not ON CONFLICT; the terminal-status set is a contract. |
| `DEV-018` | development-guide | Manual work-item crafting: truthful provenance (`source='manual'`, real `created_by`), real shapes from the owning code path, URLs from `pages.url` never invented. |
| `CTS-037` | contracts-and-standards | Input/output contracts: any input a workflow reads MUST be declared in `input_contract`; call-site `input_mapping` must satisfy the callee's contract. |
| `CTS-002` | contracts-and-standards | Component input-schema source tiers (A–D + renderer); the trap: `required:true` with `on_missing` left as `skip_field` hits the switch default and silently defers the whole section. |

Plus the meta-rule this seat exists to apply: **`FIX-016`/`FIX-042` — a
guideline-gap is a side-task (a note/amendment recommendation), not a
violation-block.**

## 3. Charter and the critical design decision

**The guidelines-agent judges: does this fix follow the platform's documented
conventions and contracts — and if it doesn't follow one, is that because the
plan is wrong, or because the *rule* is?**

This distinction drives how it uses the fixed output contract
(`{reviewer, verdict, objections, missing, checks, notes}`), where `verdict:
object` routes the plan to a revise round:

- **Plan VIOLATES a live rule** (e.g. adds a work-item insert that ignores
  `idx_swi_dedup`, or reads an undeclared input) → **`verdict: object`**, name
  the specific rule in `objections`. This correctly triggers a revise.
- **Diagnosis reveals a GUIDELINE ITSELF is wrong/stale** (the fix is sound
  but exposes a bad rule — the `MDL-039` shape) → **`verdict: approve`**, put
  the guideline-gap in `notes` as a recommended side-task/amendment. It must
  **NOT** object — forcing a correct fix to revise because the underlying rule
  is bad is exactly the failure FIX-016 designed against. This encodes "lean
  side-task, not block" within the existing contract, with no new machinery.

**Verdicts: `approve | object`, no `veto`** — same advisory design and same
reason as seats #1 and #2 (any reviewer's veto rejects outright regardless of
`hard_veto_from`, confirmed in `diagnose_council_decide_action.go`).

## 4. Prompt template (matches the existing reviewers' contract)

```
# Council reviewer: GUIDELINES AGENT

You judge two things about this fix plan: (1) does it FOLLOW the platform's
documented conventions and contracts; (2) where it appears not to, is that
because the PLAN is wrong, or because the RULE is? You change nothing; you
judge.

## The platform's load-bearing rules (the ones people keep relearning)
- WRAPPER-ORCHESTRATOR: anything doing substantive work (LLM calls, crawls,
  heavy DB, minutes of runtime) must run in a spawned pod via a parent
  (processing_mode:"orchestrator" + spawn_agent), never inline on a shared
  chassis slot; file writes from non-spawned actions die with a random pod.
- WORK-ITEM DEDUP: site_work_items dedup is idx_swi_dedup UNIQUE(site_id,
  item_key) over NON-TERMINAL statuses; the terminal-status set is a contract
  (drift between it and the Go list breaks every keyed insert fleet-wide);
  use DELETE+INSERT, not ON CONFLICT.
- TRUTHFUL PROVENANCE: hand-made work items copy the real owning path's
  metadata, deviate only truthfully (source='manual', real created_by), and
  take URLs from pages.url — never invent a path.
- DECLARED CONTRACTS: any input a workflow reads must be declared in the
  agent's input_contract; a call site's input_mapping must satisfy the
  callee's contract.
- SCHEMA-SOURCE TIERS: a component field with required:true must set on_missing
  deliberately — leaving it skip_field/empty hits the switch default and
  silently defers the whole section.

## The meta-rule for THIS seat (important)
A GUIDELINE-GAP is not a violation. If the diagnosis shows the fix is correct
but exposes a documented rule that is itself wrong or stale (this happens — a
runbook rule about max_tokens placement was recently found to be backwards),
say so in notes as a recommended side-task / guideline amendment, and APPROVE.
Do NOT object: forcing a correct fix to revise because the underlying rule is
bad is the wrong move. Object ONLY when the PLAN breaks a rule that is right.

CHECKS: if a verdict hinges on a fact a read-only SQL query could settle
(does a contract column exist, does an agent declare an input), put it in
checks as {"sql": "SELECT ...", "why": "..."} — SELECT/WITH only. Write checks
ONLY against the tables/columns in the Schema section below.

## Schema (the ONLY tables available to checks)
{{.schema_hint.text}}

## The diagnosis
{{.diagnosis_row.conclusion}}

## The plan
{{.plan_persisted.plan_json}}

## Output — ONLY this JSON
{"reviewer": "guidelines", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the specific rule violated", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "..."}], "notes": "any guideline-gap goes HERE (approve + note), not in objections"}
```

## 5. Exact wiring (extends v7, becomes v8)

Chain: `... → review_bug_historian → review_reuse_agent → review_guidelines
(NEW) → review_guardian → council_decide`.

Five edits, identical shape to the v6/v7 patches:
1. `review_reuse_agent.next_step`: `'review_guardian'` → `'review_guidelines'`
2. New step `review_guidelines`, `next_step: 'review_guardian'`
3. `council_decide.review_fields` and `escalate.review_fields`: add
   `'review_guidelines.result'`
4. `repropose.input_fields` and its prompt: add `review_guidelines`

## 6. This is the last always-on seat before the relevance-filter

Per the scaling concern raised at seat #2 and your direction ("do the
guidelines member then we can look at the relevance filtering mechanism"):
this brings the council to **5 sequential reviewers**. The remaining 7
candidates from the "ten more" list are narrower specialists (pipeline
guardians, an LLM-reliability specialist, a compliance eye) — exactly the
seats the relevance-filter is meant to activate only when relevant, rather
than run on every decision. Design for that filter follows this seat; the
narrow specialists get built behind it, not as more always-on steps.
