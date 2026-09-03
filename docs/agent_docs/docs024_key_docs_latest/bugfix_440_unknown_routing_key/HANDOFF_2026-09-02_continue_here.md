# HANDOFF 2026-09-02 — bugfix 440 (unknown routing key): phase 1a APPROVED + SHIPPED; what phase 1b–3 need; one new landmine and one debt

> **SUPERSEDED 2026-09-03** — debts cleared, build re-verified, owner decisions enumerated:
> `HANDOFF_2026-09-03_continue_here.md` (same directory). This file kept for the trail.

Written for a session with none of this context. Every claim carries its check; cite symbols,
never line numbers. This lane was spun out of `bugs_closed/410_…three_seams…`'s candidate 1 by
owner decision 2026-09-02. ⚠ The number 410 is AMBIGUOUS (the phase-lock case is unrelated and
still open); the number 440 is not, yet — check before assuming.

## What this bug is, in one paragraph

`spec.reason` on `page_rerender` items is TWO fields wearing one name: the gate's routing key
AND a free-prose human annotation (11 prose items minted 2026-09-02 alone, by migrations 696/693
— raw SQL that bypasses every Go guard). The gate (`check_rerender_mode`, read live: a single
conditional, five-value `==` disjunction, `else_step: render_page`) silently ASSEMBLES anything
it does not recognise — so a routing key nobody understands completes green having changed
nothing. Refusal is impossible until the fields are split; the split is RFC_062; the refusal is
its phase 3. Full evidence: the bug file (`bugs_open/440_HANDOFF_2026-09-02_…`), each figure
dated and disconfirmable.

## State — done and verified

| what | state | re-check |
|---|---|---|
| bug filed, 410 closed | done | `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ \| grep three_seams` → exactly one line, in bugs_closed |
| RFC_062 (the design) | DRAFTED | `architecture_review/RFC_062_routing_key_annotation_split.md` — split; refusal at WRITE door (CHECK NOT VALID) > gate refuse-branch > authoring advisory; 2 open questions for the owner at its tail |
| phase 1a code | **APPROVED r1 (corr `55def842`, 2 advisories none high) + SHIPPED** | `platform/livespec/rerender_routing_key{,_test}.go`, commit `a3758c399`; REB-008 (+index row `0600eb6b3`); verdict dispositions in NOTES. **Shipped = by ANCESTRY, not by probe — see the landmine below** |
| coordination | done | CONTRIB in 404's NOTES (their r4 is `complete_approved`, told them; also told them their warning has 0 production firings) and 384's NOTES (`listing_stale`); commit `5b5c669dd` |
| wrong calls | logged | WRONG_CALLS 2026-09-02: text-LIKE counts strings QUOTED in council payloads as emissions — exclude `fix_correlation_id IS NOT NULL` |

## ⚠ NEW LANDMINE (filed today, and it will bite the next phase-verifier)

**A capability probe for INERT code reads ABSENT with clean controls.** Phase 1a has zero
callers, so dead-code elimination strips its functions and folded literals: both fresh pods
probed ABSENT for `input_data.spec.routing_reason` while their stamp `0d2feee2ff61` carries
`a3758c399` by ancestry — and a local `go build ./cmd/agent-chassis` from source CONTAINING the
module also greps 0 (controls 1). **Verify inert phases by ancestry** (`service_binary_capabilities`
→ `git merge-base --is-ancestor`); the literal probe becomes valid only when phase 1b adds a
caller, and its first PRESENT dates the CALLER's roll. RUNBOOK has the exact commands.

**⚠ DEBT: the LANDMINES entry is in the FILE but NOT synced** — the kubeconfig token expired
mid-session (fleet-wide `Unauthorized` = the 3-day expiry; owner refreshes). First session after
refresh MUST run `./scripts/landmines-verify-dispatch.sh` (never `landmines-sync.py --apply`
alone — it consumes the new-entry status and the verifier then skips it; documented trap).

## What REMAINS — the phases, and what gates each

1. **Phase 1b — creator stamps `routing_reason` alongside `reason`** (in-vocabulary values only).
   File: `create_rerender_items_action.go` (the 404 lane's). **GATED on the 404 lane reading and
   recording their r4 verdict** (approved 2026-09-02, orch `40639f27`, still unread by them at
   handoff time — check `git log --oneline -3 -- docs/…/bugfix_404_rerender_reason_vocabulary/`
   for movement). Additive; council gate, not RFC.
2. **Phase 2 — reach the unguarded doors**: migration-authoring rule in sql_for_agents
   conventions + `scripts/pattern-check.py` advisory (routing-shaped unknown `reason` in a
   migration INSERT → "did you mean routing_reason?"). ⚠ if the advisory lands in
   pattern-check.py it is IN council scope (2026-08-24 widening). Independent of 1b.
3. **Phase 3 — the flip (RFC_062 proper)**: gate migration pastes
   `TransitionRerenderModeConditionClause()` (drain window), then the refuse-branch second
   conditional pasting `CheckRoutingKnownConditionClause()`; optionally the CHECK NOT VALID at
   the write door. **Blockers, all stated in the RFC**: (a) owner's two open questions (refusal
   destination — lane recommends `needs_human_review`; whether 404 co-signs the gate migration —
   lane recommends yes); (b) the missing-key-vs-'' evaluator behaviour MUST be confirmed before
   any paste (inverting it refuses every legacy item — flagged in the module header, the RFC,
   and the approved submission's risks); (c) census of in-flight items clean; (d) consumers told
   (list in lane PLAN; 404+384 already done).
4. **REB-008 constraint (architecture seat, round 55def842)**: `RoutingReasonSpecKey` must gain
   NO second producer before RFC_062 lands; phase 1b's creator is the FIRST sanctioned one.

## Known cross-lane facts a fresh session needs

- `platform/livespec` package test run FAILS at HEAD regardless of this lane:
  `TestNoNewMigrationFileReadersOutsideTheAllowList`, the 405 lane's committed breakage
  (`ffa1707b3`, ~7 days). Reproduces without our files. Theirs.
- 404's four rounds never drew a design objection — every gate was submission accuracy. When
  submitting here: attach query OUTPUT, not the claim it was run (their r3/r4 lesson, and our
  round's two medium advisories were exactly this).
- 097 gotchas: `operation` is an enum (modify|add|remove|config_change); a sketch whose every
  line starts `#`/`//`/`--` is refused as comment-only (a markdown `###` heading triggers it).

## Key artefacts

| what | where |
|---|---|
| bug file | `bugs_open/440_HANDOFF_2026-09-02_a_routing_key_nobody_understands_completes_green.md` |
| design | `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_062_routing_key_annotation_split.md` |
| lane docs | `docs/agent_docs/docs024_key_docs_latest/bugfix_440_unknown_routing_key/` (PLAN = phases+consumers · NOTES = evidence, dispositions, missteps · RUNBOOK = census/emission-counting/inert-verification recipes) |
| code | `platform/livespec/rerender_routing_key{,_test}.go` (REB-008; three-state resolver, two paste-target clause renderers, both mutations killed) |
| commits | `ec2efc06e` (close+file) · `a3758c399` (phase 1a) · `0600eb6b3` (index row) · `5b5c669dd` (CONTRIBs) · `544de50e0` (verdict close-out) |
| council | `55def842-9874-480b-991f-ce6d1f54154c` APPROVED r1 |
| parent lane (all closed) | `docs024_key_docs_latest/bugfix_410_silent_scan_loss/` + `bugs_closed/410_…three_seams…` |
