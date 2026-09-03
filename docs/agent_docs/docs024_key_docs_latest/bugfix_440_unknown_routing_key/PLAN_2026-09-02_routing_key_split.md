# PLAN 2026-09-02 — bugs_open/440: split `spec.reason`, then refuse what nobody understands

Design and phasing for the refusal half of 410's candidate 1. The design document proper is
`architecture_review/RFC_062_routing_key_annotation_split.md`; this file is the lane's working
plan and decision log. Owner directive 2026-09-02: robust, framework-wide over point fix;
coordinate, never work in isolation; council for the code.

## The decision that shapes everything (and its reason)

**Refusal lives at the doors data actually enters through, not only at the one door with a Go
guard.** Measured 2026-09-02: the dominant producer of out-of-vocabulary reasons is raw-SQL
migration INSERTs, which no Go code can see. So the layers are, strongest first:

1. **Write door (the eventual CHECK)**: `page_rerender` items may carry `spec.routing_reason`
   only from the vocabulary; a bad key fails the INSERT at apply time, at the author's desk,
   synchronously. Unrepresentable > detectable.
2. **Read door (the gate)**: a second condition step — routing_reason present-but-unknown →
   refusal branch (fail toward review), absent → assemble as today. Catches rows that predate
   the CHECK and any door the CHECK cannot cover.
3. **Authoring door (advisory)**: pattern-check warning on migrations INSERTing `page_rerender`
   items with routing-shaped unknown `reason` values ("did you mean routing_reason?").

`spec.reason` itself is NEVER validated — it is the annotation, free prose is legal forever.
That is what makes refusal possible at all: post-split, "present but unknown" is a mistake by
construction, with zero legitimate counter-examples (`verbatim_adoption_deploy` and the
migration-696 prose live happily in `reason`, stamp no routing key, and assemble as intended).

## Phases

| phase | content | ships | gate |
|---|---|---|---|
| 1a (tonight) | `platform/livespec/rerender_routing_key.go`: the spec-key constant, resolver, gate-clause renderer for BOTH transition and final forms, doc header pinning opt-in/default-OFF and naming RFC_062. Tests incl. the negative (unknown key resolves false). INERT — zero callers, zero behaviour change | council gate (097) + commit | not architecture-scope (RFC_022: opt-in, unsafe default OFF, zero live consumers — enumerated, not asserted: `grep -rn RoutingReasonSpecKey` returns the definition and tests only) |
| 1b | creator stamps `routing_reason` alongside `reason` for in-vocabulary values (`create_rerender_items_action.go`); additive, gate still reads `reason` | council gate | ~~AFTER the 404 lane has read its r4 verdict~~ **GATE LIFTED by owner ruling D5, 2026-09-03; SHIPPED same day** — lockstep with `reason` (not merely-known), so the flip is byte-neutral for REB-001's designed degrade |
| 2 | producers: fixer SQL template + migration-authoring rule in sql_for_agents conventions + pattern-check advisory (layer 3) | council gate | — |
| 3 | THE FLIP, RFC_062's subject (**owner-ruled 2026-09-03: refusal routes to `needs_human_review` (D1); the CHECK constraint IS in scope (D3); the 404 lane CO-SIGNS the gate migration (D2); `spec.reason` is never validated (D4)**): gate reads `routing_reason` (compat OR during drain), second condition step refuses present-unknown toward review; optionally CHECK NOT VALID at the write door. Requires: census of in-flight items clean, consumers told (2026-07-29 §3), RFC accepted | RFC + council + owner | changes what the shared gate GUARANTEES — full architecture track |

## Consumers (2026-07-29 §3: told, not merely measured) — the list, owners in brackets

- `check_rerender_mode` gate config [page-rerender agent; declaration-audited]
- `create_rerender_items_action.go` + livespec vocabulary [404 lane — r4 approved 2026-09-02; CONTRIB posted in their NOTES before any shared-file edit]
- raw-SQL migration authors [estate-wide; the conventions doc + advisory reach them]
- `adopt_verbatim.go` [annotation-only; unaffected by design — verify, not assume, at phase 3]
- 615-shape fixer fan-out [stamps in-vocabulary `template_changed`; gains `routing_reason` in phase 2]
- `rerender-pages` agent [passes `spec.reason` through — 404's census; passes routing_reason through identically at phase 2]
- single-value readers `shouldStripLiteralMarkdown`, CTA recompute [read the ROUTING meaning → move to routing_reason at phase 3, constants already shared]
- 384 lane [their `listing_stale` key appeared once, 08-24 — flagged in CONTRIB so they route it into the vocabulary or the annotation deliberately]

## Not doing, stated

- No blanket refusal on `spec.reason` (live legitimate counter-examples; the whole point of the split).
- No new work-item type, no new table.
- No touching 404-lane files until their verdict is read and recorded by them (1b's gate).
