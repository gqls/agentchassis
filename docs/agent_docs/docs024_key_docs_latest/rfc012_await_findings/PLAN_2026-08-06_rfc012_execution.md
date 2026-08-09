# PLAN 2026-08-06 — executing RFC_012's second-sitting rulings

Owner rulings (recorded in the RFC, commit `3851e90b5`), three pieces, this lane owns all:

1. **(d) → a STANDING check, online if the framework allows.** Bound by addendum 1
   (13-key routing graph, different-action discriminator; both naive detectors return 0
   on the known bug — a candidate must prove itself against that case). Vehicle to be
   chosen against the RFC_006 precedent: live config is guarded by a scheduled check,
   never a commit-time hook.
2. **The §3(a) reader census — COMMISSIONED.** Deliverable: an artefact in
   `architecture_review/` naming every reader of an awaited step's key (config side from
   live `agent_definitions`; Go side mechanically-findable readers, honestly bounded)
   with a per-reader verdict under merge-not-replace. This is what (a)/(a′) stays gated
   behind; the census does not decide (a).
3. **Option B implementation — ASSIGNED here, now.** A leaf-package shared
   `agent_error_log` writer importable from `actions` (retiring the hand-copied INSERT
   column lists) + a named findings-plus-await helper generalising 098 debt-5b's proven
   direct-DB-row pattern. Through the council gate; concept-register entry in the same
   commit as the seam.

## Decisions and their reasons

- Research delegated to three parallel agents (helper ground truth; online-check
  machinery precedents; the census itself) — this session's context is deep, and the
  census in particular is exactly the shape an agent can run end-to-end with the method
  written into the artefact for re-running.
- Order: census lands independently (its own commit); helper + check are code and go
  through one council round as a coherent task (they share the RFC and the register
  category) unless research shows they are better split.

## Verification (to be completed as research lands)

- Helper: unit tests + the debt-5b shape re-expressed through it; `go build ./...` on
  `git archive HEAD`; the import-cycle claim proven by the build itself.
- Standing check: MUST detect the known-bug case addendum 1 names (the disconfirmable
  half); a no-finding run on today's fleet is only meaningful beside that.
- Census: queries verbatim in the artefact so it can be re-run; totals stated;
  [UNVERIFIED] marked where a grep cannot see.

---

## 2026-08-09 — §1a, the last piece, and two decisions that resized it

§1a (hoist the `RunAgentType` ladder) was not in the three owner rulings above. It was carved out
of the RSH-008 hardening deliberately, because this lane's round 2 was REJECTED for bundling, and
it is now shipped as `1bc08d1ce` / `RSH-009` / `RFC_019`.

**DECISION 1 — the `os.Getenv("AGENT_TYPE")` rung stays coordinator-side; the shared method stops
at the context.** The 08-08 handoff called this "the design decision of the job" and left it open.
Reasons, in the order they weigh: `AGENT_TYPE` is a property of the process and this type is
deserialised from Kafka headers, so a method reading it would answer differently per pod,
invisibly at the call site; and the two consumers' tails *legitimately differ* — the actions door's
floor (`params.AgentType` = `state.OwnerAgentType`) is strictly more specific than the pod, and it
must never reach for the `"generic"` filler because that is the value RSH-008 chose `unattributed`
to avoid colliding with. `ResolvedAgentType(fallback string)` was considered and rejected: it
looks like one ladder while two callers pass different fallbacks. `t.Setenv` pins the exclusion so
the decision is enforced rather than documented.

**CORRECTION to the brief this work was commissioned on, recorded here rather than silently
absorbed.** The 08-08 handoff sized §1a at "**559** rows across 25 distinct `step_name`s" plus 25
`REVIEW_SUPERSEDED_BY_PASSING_SAVE` rows. Measured before building: **499 of 555 predate
2026-07-26**, the day `RunAgentType` shipped; the 25 rows are all from **one day**, 2026-07-23.
Live residue: **~36 rows in 13 days**. The change is justified structurally — one question, two
ladders, in packages that cannot import each other — and **not** by volume. `RFC_019` §3 therefore
lists "do nothing" as a serious option. The class of error (a retention-bounded table prices a
fixed defect identically to a live one) is filed in `LANDMINES.md` and `WRONG_CALLS.md`.

**DECISION 2 — the change went to BOTH forums, and the gate REJECTED it on scope.** RSH-008's
round-1 `architecture` seat had ruled that change `point_fix` explicitly because it *"stays inside
`platform/orchestration/actions`"*; this one does not, so the precedent was not claimed. Round 1
(corr `6186ab10-a006-4c34-b9ea-ecedfde8ea2d`): **REJECTED**, hard veto from `guardian`, with
`architecture` returning `ARCHITECTURE_SIGNAL: needs_rfc` — so my own §8 argument that this was
gate scope was wrong, and is corrected in `RFC_019` §10.

**It is deliberately NOT resubmitted and NOT reverted.** The guardian's contained alternative
(duplicate the two-line read locally, leave `types` and `coordinator.go` alone) is the second
ladder this change exists to retire, and the `architecture`, `reuse_agent` and `constitution`
seats say so in the same round. Per CLAUDE.md 2026-07-28 a scope veto is not answered by
resubmitting with better measurements, especially when seats disagree with each other; the
guardian itself routes the call to `RFC_019`. **The open item is an owner decision, not a task.**

**Still open, and it is a measurement not an argument:** whether §1a is a partial no-op on resumed
steps (`ensureFullExecutionContext` never backfills `RunAgentType`). Undecidable before the roll —
`orchestration_states` keeps ~24h and every affected row is weeks old. The post-roll query and its
36-row baseline are in `RSH-009`'s `verify-later`.
