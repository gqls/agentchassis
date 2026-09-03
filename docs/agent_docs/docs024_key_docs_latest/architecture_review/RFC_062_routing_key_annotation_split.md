# RFC_062 — split `spec.reason` into routing key + annotation, then refuse a routing key nobody understands

**Status: DESIGN RULED BY THE OWNER 2026-09-03 — open questions closed, see §Rulings; implementation phased** · raised by `bugs_open/440` (spun out of 410's candidate 1, owner
decision 2026-09-02) · owning lane `docs024_key_docs_latest/bugfix_440_unknown_routing_key/` ·
prior art: `bugs_open/404` (whose livespec header names this split as "the real repair" and
defers it here), owner rulings 2026-08-02 §2 (opt-in shape) and 2026-07-29 §1 (the RFC trigger).

## Why this is RFC-scope (and what is deliberately not)

The flip in phase 3 changes what the shared `page-rerender` gate GUARANTEES: today every
unrecognised `spec.reason` assembles (re-ships stored HTML, completes green); after, a
present-but-unknown ROUTING key refuses toward review. That is a guarantee change on a shared
mechanism → RFC per 2026-07-29 §1. Phases 1a–2 are additive and inert (opt-in, unsafe default
OFF, zero live consumers at introduction) → explicitly NOT architecture-scope per RFC_022; they
go through the ordinary council gate.

## The problem, in the data (all `[MEASURED 2026-09-02]`, bugs_open/440 for full evidence)

`spec.reason` is two fields wearing one name. On live `page_rerender` items: 5 vocabulary
routing keys; 1 deliberate assemble-only key (`verbatim_adoption_deploy`); free-prose
annotations minted THE SAME DAY by migrations 696/693 (which bypass the Go creator, so 404's
loud-warning guard — probe-verified live — has fired zero times in production); and
routing-SHAPED unknowns (`tool_retirement` ×16, `light_palette_chrome_replaced` ×13,
`listing_stale` ×1) that silently assembled. Migration 696 proves intent is unrecoverable from
data: it updated `content_data` AND `rendered_html` together, so ITS assembles are correct —
while an author who edited only `content_data` would mint identical rows whose assemble
re-ships the stale artefact. **You cannot refuse unknown values of a field whose legitimate
values include arbitrary sentences. Splitting is what makes refusal well-defined.**

## Design

- **`spec.reason` — the annotation.** Free prose, never validated, forever legal. Existing
  producers keep writing it unchanged.
- **`spec.routing_reason` — the routing key.** Only `livespec.RerenderSectionReasons` values.
  Absent → assemble (today's safe default, preserved: annotation-only items stay legal).
  Present-and-known → routes exactly as `reason` does today. **Present-but-unknown → REFUSE**:
  fail the item toward review, never silently assemble. Post-split there is no legitimate
  counter-example of an unknown routing key — `verbatim_adoption_deploy` and the prose live in
  `reason` and stamp nothing.
- **Refusal placement, strongest door first** (the census showed raw-SQL migrations are the
  dominant unguarded producer):
  1. *Write door*: CHECK constraint (added NOT VALID; validated after census) on
     `site_work_items` scoped to `item_type='page_rerender'`: `spec->>'routing_reason' IS NULL
     OR spec->>'routing_reason' IN (<vocabulary>)`. A bad key fails the INSERT at the author's
     desk, synchronously — unrepresentable beats detectable. Vocabulary changes already require
     a paste-from-livespec migration for the gate; the CHECK rides the same migration.
  2. *Read door*: a second `conditional` step ahead of `check_rerender_mode` (the live gate is
     one conditional with only then/else — read 2026-09-02): `routing_reason == '' OR
     routing_reason IN (<vocabulary>)` → continue; else → refusal step. Catches pre-CHECK rows
     and any door the CHECK misses. ⚠ Implementation must confirm the condition evaluator's
     behaviour on a MISSING key vs empty string before this ships — flagged, not assumed.
  3. *Authoring door (advisory)*: pattern-check warning on migration files INSERTing
     `page_rerender` items whose `reason` is routing-shaped (single snake_case token) and
     out-of-vocabulary — "did you mean routing_reason?".
- **Transition**: gate condition becomes `routing_reason == 'x' OR reason == 'x'` per value
  (rendered by livespec, pasted into the migration — the established idiom, declaration-audited
  daily) while producers move; narrows to `routing_reason` once in-flight items drain (they
  drain in days) and the single-value readers (`shouldStripLiteralMarkdown`, the CTA recompute)
  have moved to the shared constants against the new key.

## Consumers (2026-07-29 §3 — to be TOLD, not merely measured; the telling is in each lane's NOTES)

The full list with owners is in the lane PLAN. Notable: the 404 lane (livespec + creator files;
their r4 approved 2026-09-02 — CONTRIB posted in their NOTES before any shared-file edit); the
384 lane (their one-off `listing_stale` key — flagged to them to route or annotate
deliberately); raw-SQL migration authors (reached by the conventions rule + advisory, the only
mechanism that reaches a door Go cannot see).

## Alternatives considered

- **Blanket refusal on `spec.reason`**: impossible — live legitimate free-prose population,
  growing daily. This is why 404 shipped loud-not-refuse and why the split precedes any refusal.
- **Refuse at the Go creator only**: guards one door of five; the dominant unguarded producer is
  raw SQL. Rejected as primary; the creator keeps its warning and gains stamping.
- **A vocabulary TABLE read by a trigger** (instead of pasted CHECK): softer to update but adds
  a runtime dependency and a second source of truth beside livespec; the paste idiom is already
  established and declaration-audited. Rejected for now; revisit if vocabulary churn accelerates.

## Rulings — OWNER, 2026-09-03 (these close the questions below; kept for the trail)

| # | question | RULING |
|---|---|---|
| D1 | refusal destination | **Route the item to `needs_human_review`**, never a silent assemble and not a blunt orchestration failure — consistent with the 2026-07-31 never-silently-dropped precedent. The message must name the offending key AND the vocabulary. |
| D2 | who signs the phase-3 gate migration | **The 404 lane CO-SIGNS.** The declarations it rewrites are theirs and the daily auditor holds them; do not ship it unilaterally. |
| D3 | the write door | **YES — add the CHECK constraint** (NOT VALID first, validated after the census). This is the layer raw-SQL migration authors cannot bypass, which the census showed is the dominant unguarded producer. Every vocabulary change updates it in the same migration, exactly as the gate condition already requires. |
| D4 | policing the annotation | **NO.** `spec.reason` stays free prose forever; the authoring advisory nudges, nothing forbids. |
| D5 | phase 1b's courtesy gate on the 404 lane | **LIFTED** — proceed. (Their round is approved and they have been told; the wait was this lane's courtesy, not a rule.) |

Consequences already carried into the plan: phase 3 ships the CHECK (D3) and the review-routing
refusal (D1) together, co-signed (D2); nothing validates `spec.reason` (D4); phase 1b shipped
2026-09-03 (D5).

## Open questions for the seats / owner — ANSWERED ABOVE, retained for the trail

1. Refusal destination: fail the orchestration (loud, blunt) vs route the item to
   `needs_human_review` (the 2026-07-31 never-silently-dropped precedent)? Lane recommends
   review-routing.
2. Should the CHECK also forbid vocabulary values appearing in `spec.reason` once producers
   move (prevent regression to the old habit), or is that over-constraint on an annotation field?
   Lane recommends: no — annotation stays free, the advisory nudges instead.
3. Timing of phase 3 relative to 404's remaining work — their lane should co-sign the gate
   migration since the declarations it rewrites are theirs.
