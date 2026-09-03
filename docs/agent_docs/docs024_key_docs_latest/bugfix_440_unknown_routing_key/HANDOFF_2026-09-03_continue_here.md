# HANDOFF 2026-09-03 — bugfix 440 (unknown routing key): foundation approved+shipped+verified; five owner decisions pending; phases 1b–3 are the remaining work

Written for a session with none of this context. Every claim carries its check; cite symbols,
never line numbers. Supersedes `HANDOFF_2026-09-02_continue_here.md` (kept for the trail).
Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_440_unknown_routing_key/`.

## The bug in one paragraph

`spec.reason` on `page_rerender` items is TWO fields wearing one name: the gate's routing key
AND a free-prose human annotation (evidence in `bugs_open/440_HANDOFF_2026-09-02_…`, every
figure dated). The live gate (`check_rerender_mode`: five-value `==` disjunction, `else_step:
render_page`) silently ASSEMBLES anything it doesn't recognise, so a routing key nobody
understands completes green having changed nothing. Refusal is impossible until the fields
split (`spec.routing_reason` = vocabulary-only; `spec.reason` = annotation, free forever) —
the split is RFC_062; the refusal flip is its phase 3.

## State — done, approved, shipped, verified

| what | state | re-check |
|---|---|---|
| bug filed; parent 410 closed | done | `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ \| grep three_seams` → one line, bugs_closed |
| RFC_062 | DRAFTED, awaiting owner decisions D1–D4 below | `architecture_review/RFC_062_routing_key_annotation_split.md` |
| phase 1a (inert foundation) | **APPROVED r1 (corr `55def842`) + SHIPPED + VERIFIED 2026-09-03** | `platform/livespec/rerender_routing_key{,_test}.go`, REB-008, commit `a3758c399`. Verified by ANCESTRY: pods of ReplicaSet `75b987cbd7` stamped `7bf1ff674021`, `git merge-base --is-ancestor a3758c399 7bf1ff674021` holds. ⚠ NEVER literal-probe this while it has zero callers — DCE strips it (LANDMINES 2026-09-02; verifier dispatched `dd777a91`, verdict unread) |
| coordination | done | CONTRIBs in 404's and 384's NOTES (`5b5c669dd`); consumers enumerated in lane PLAN |
| debts | **cleared 2026-09-03** | landmines sync in doc_notes; per-entry verifier dispatched after another session's sync consumed the new-status (the documented trap, remedy applied) |

## FIVE OWNER DECISIONS — the lane is blocked on these for phase 3 (and D5 unblocks 1b)

Explained in plain terms, rule, then recommendation (owner note 2026-08-12 format):

- **D1 — where does a refused item go?** Fail the rebuild loudly, or park the item in
  `needs_human_review` with a message naming the bad key and the vocabulary. Precedent: the
  2026-07-31 ruling (odd orders are never silently dropped — they go to review). **Recommend
  review-routing** — a bad key is usually a typo or an undeclared value; a human routes it in
  minutes and nothing is lost.
- **D2 — does the 404 lane co-sign the phase-3 gate migration?** The migration rewrites gate
  declarations that lane wrote and the daily auditor holds. **Recommend yes.**
- **D3 — do we add the database-door lock?** A CHECK constraint (NOT VALID first) so even
  hand-written migration SQL cannot INSERT a bad routing key — the strongest layer, because the
  census shows raw SQL is the dominant unguarded producer. Cost: every vocabulary change must
  update the constraint in the same migration (already true of the gate condition — same paste
  idiom). **Recommend yes.**
- **D4 — is the annotation field ever policed?** Once producers move, should vocabulary values
  be forbidden from appearing in `spec.reason`? **Recommend no** — the annotation stays free
  forever; the authoring advisory nudges instead.
- **D5 — may phase 1b proceed before the 404 lane records their r4 verdict?** The gate was this
  lane's courtesy, not a rule: their code is committed and their verdict is `complete_approved`
  (they have been told, CONTRIB 2026-09-02). **Recommend: proceed now**, with the same-file
  care below.

## What REMAINS, in order (what "closed" requires)

1. **Phase 1b** — creator stamps `routing_reason` alongside `reason` for in-vocabulary values
   (`create_rerender_items_action.go`). Unblocked by D5. ⚠ THREE lanes have touched this file
   (404; 315-reopen's `8eca969cb` on 2026-09-03; us next) — re-read it fresh at write time,
   expect same-file passengers, pathspec commit. Council gate, not RFC.
2. **Phase 2** — reach the unguarded doors: migration-authoring rule in sql_for_agents
   conventions + `scripts/pattern-check.py` advisory (routing-shaped unknown `reason` in a
   migration INSERT → "did you mean routing_reason?"). ⚠ pattern-check.py is IN council scope
   (2026-08-24 widening). Independent of 1b.
3. **Phase 3 — the flip (RFC_062)**: gate reads `routing_reason` (paste
   `TransitionRerenderModeConditionClause()` for the drain window), second conditional refuses
   present-but-unknown (paste `CheckRoutingKnownConditionClause()`), plus D3's CHECK if ruled
   yes. Blockers beyond D1–D4: (a) **confirm the condition evaluator's behaviour on a MISSING
   key vs ''** — inverting it refuses every legacy item (flagged in module header, RFC, and the
   approved submission's risks); (b) drain census clean; (c) consumers told (404/384 done; rest
   via the conventions rule).
4. **Close bug 440** when the refusal is **fixed AND live**: an induced unknown routing key
   REFUSES at the live gate (the bug file's verification-trap section says exactly how — prove
   it can fail, both directions), annotation-only prose still assembles unwarned, and the
   census shows no stranded in-flight items. Then `git mv` to bugs_closed (both paths on the
   commit; verify at HEAD).
5. Housekeeping riders: read the landmine-verifier verdict (`dd777a91`); REB-008 flips from
   "BUILT AND INERT" to live wording at phase 3 — and its no-second-producer constraint lifts
   only when RFC_062 lands.

## Known cross-lane facts

- `platform/livespec` package run FAILS at HEAD regardless of this lane
  (`TestNoNewMigrationFileReadersOutsideTheAllowList`, 405 lane's `ffa1707b3`). Theirs.
- Submission craft: attach query OUTPUT, not claims it was run (404's r3/r4 lesson; our r1's two
  mediums were exactly this). 097: `operation` is an enum; all-comment sketches are refused.
- Counting warning emissions: exclude `fix_correlation_id IS NOT NULL` or you count council
  payloads quoting the string (WRONG_CALLS 2026-09-02).

## Key artefacts

| what | where |
|---|---|
| bug file | `bugs_open/440_HANDOFF_2026-09-02_a_routing_key_nobody_understands_completes_green.md` |
| design + open decisions | `architecture_review/RFC_062_routing_key_annotation_split.md` (+ D1–D5 above) |
| lane docs | this directory (PLAN = phases+consumers · NOTES = evidence+dispositions+missteps, newest at bottom · RUNBOOK = census, emission-counting, inert-verification recipes) |
| code | `platform/livespec/rerender_routing_key{,_test}.go` (REB-008) |
| commits | `ec2efc06e` · `a3758c399` · `0600eb6b3` · `5b5c669dd` · `544de50e0` · `35de364dd` |
| council | `55def842` APPROVED r1 (dispositions in NOTES) |
