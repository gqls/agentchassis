# RFC_062 — split `spec.reason` into routing key + annotation, then refuse a routing key nobody understands

**Status: DESIGN RULED BY THE OWNER 2026-09-03 — open questions closed, see §Rulings. Phases 1a, 1b and 2 are SHIPPED AND LIVE. Phase 3 is BUILT, PROVEN BY EXECUTION, and HELD at migration `741_..._HOLD.sql` pending the 404 lane's co-sign (D2) — see §Phase 3 as built** · raised by `bugs_open/440` (spun out of 410's candidate 1, owner
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
     and any door the CHECK misses. ⚠ ~~Implementation must confirm the condition evaluator's
     behaviour on a MISSING key vs empty string before this ships — flagged, not assumed.~~
     **DISCHARGED 2026-09-03 BY EXECUTING THE EVALUATOR, and the assumption was FALSE.** A
     missing key does NOT match `== ''` (`compareValues`' nil branch runs before quote-stripping,
     so the quoted `''` never equals nil). Measured truth table — absent: `== null` TRUE, `== ''`
     false; present-empty: `== ''` TRUE; present-known: the value TRUE; present-unknown: all
     false, the only refusing state. **The clause therefore needs BOTH `== null` and `== ''`
     disjuncts**; the renderer shipped in phase 1a carried only `== ''` and would have routed
     every item minted before phase 2 — the entire legacy population — to human review on flip
     day. Fixed, with the four-state table pinned against the real evaluator
     (`rerender_routing_gate_clause_test.go`) and the `== null` disjunct mutation-proved.
     LANDMINES 2026-09-03 carries the general trap.
  3. *Producer conversion — WIDENED 2026-09-03 by the phase-1b producer census
     (bugs_open/440)*: the routing key must be stamped by EVERY producer, not just the Go item
     creator. `[MEASURED 2026-09-03]` 13 Go files write an in-vocabulary reason directly, mostly
     as raw `{"reason":"x"}` literals bypassing the vocabulary constants, and agent/migration
     producers mint the rest — but ⚠ **CORRECTED same day: only FIVE of those 13 are
     `page_rerender` producers** (the others file `needs_page`, `needs_rerender` or
     `literal_markdown`, and stamping a routing key on them would put a page-rerender routing
     decision on an item no rerender gate reads — enumerate item types, never sweep); the creator phase 1b converted mints just **1 of 3,172**
     reason-bearing items. **Narrowing the gate before that conversion completes would route
     ~3,100 items to assemble — this bug's own shape, fleet-wide, inside the change meant to fix
     it.** So the transition clause is LOAD-BEARING, not a drain-window nicety, and phase 3 gains
     a gate condition: a census showing zero reason-bearing items from unconverted producers.
  4. *Authoring door (advisory)*: pattern-check warning on migration files INSERTing
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

## Phase 3 as built — 2026-09-03, held at the co-sign

The flip is written, executed against the live database inside a transaction and rolled back
(twice, then a full apply → VERIFY → ROLLBACK → compare round trip). It ships as
`docs/agent_docs/sql_for_agents/741_refuse_unknown_rerender_routing_key_HOLD.sql` with `_ROLLBACK`
and `_VERIFY` companions. `_HOLD` because of D2 and nothing else: the 404 lane has been dormant
since 2026-08-26 and the CONTRIB asking for the co-sign is unanswered, so the owner's decision of
2026-09-03 was **build it all, stop before applying**.

What it does: a new start step `check_routing_key_known` ahead of the gate (absent / empty / known
→ on to the gate; present-but-unknown → refusal); `refuse_unknown_routing_key` parking the item at
`needs_human_review` (D1) with a message that names the offending VALUE and the vocabulary;
`check_rerender_mode` moved to the TRANSITION condition; and the CHECK constraint added NOT VALID
(D3), scoped to `item_type='page_rerender'`.

**D1 needed a code change, and it is the opt-in shape.** `fail_work_item`'s `error_message` is a
config LITERAL with no interpolation — `[MEASURED 2026-09-03]` 7 live steps across 6 agents carry
it and 0 contain `{{`. A static message can name the field and the vocabulary but never the key
that was actually wrong. So `fail_work_item` gains an opt-in `error_message_template`, default OFF,
byte-inert for all seven live steps: RFC_022-exempt on all three conditions, ordinary council gate.
Both of its failure modes are loud (`missingkey=error` for an absent path, an explicit guard for
text/template's `<no value>` on a present-but-nil key) and both fall back to the static literal
with an `agent_error_log` entry, because a human must be able to find out that the refusal message
they are reading is the second-best one.

### ⚠ The flip leaves a NEW blind spot, and it is 404's own drift class

`[MEASURED 2026-09-03, by execution]` **both** of the 404 lane's live Declarations stay GREEN
through this flip, unchanged and un-edited:

- the `FragmentMatch` on `CheckRerenderModeConditionClause()` holds, because the old five-value
  clause is a substring of the transition clause **exactly once** (Min:1/Max:1 satisfied);
- the paired count still reads **5**, because it counts `input_data.spec.reason ==`, and
  `input_data.spec.routing_reason ==` does not contain that as a substring.

That is convenient and it is a defect. Five new `routing_reason ==` disjuncts arrive asserted by
**nothing**, so a sixth routing value appended to the live gate without touching Go would drift
exactly the way `bugs_open/404` drifted — inside the change built to fix it. The existing
Declaration's own comment states the principle it now fails: *"A fragment sees loss and mutation;
only a count sees ADDITION."*

The remedy is a `routing_reason ==` count Declaration plus entries for the guard step and the CHECK
constraint. Those edits are **deliberately not in the phase-3 commit**: the daily auditor's own
note says to fix a declaration *"in the same commit as the migration that moved it"*, and today's
row reads `probed 15 live object(s); 0 finding(s)` — committing declarations for a HELD migration
would turn that clean 0 red every morning until the co-sign landed, and a permanently-red auditor
masks real drift. No Declaration can be honestly green both before and after the flip (pre-flip the
probe returns NULL for a step that does not exist). So the five specific edits are enumerated in
741's own header, in the file the applier must open.

### The residual phase 3 does not close, stated

`keys_disagree` — an item carrying `routing_reason` and a DIFFERENT `reason` — is **0 as of
2026-09-03** and nothing enforces it. In that state the transition clause routes the item (it
matches on either key) but `rerender_sections` is still handed `input_data.spec.reason`, so the
single-value readers (`shouldStripLiteralMarkdown`, the CTA recompute) would see the annotation and
silently under-deliver: this bug's own shape in miniature. It is empty because producers stamp in
LOCKSTEP, which is a property of the producers, not of the gate. Per the lane PLAN those readers
move to `routing_reason` when the gate NARROWS; the `_VERIFY` companion counts `keys_disagree` so
the state cannot arrive unnoticed.

### Narrowing is still blocked, and the number moved

`[MEASURED 2026-09-03]` **1,804** pending `page_rerender` items carry an in-vocabulary `reason`
and **14** carry a `routing_reason` (12 this morning — the converted producer is live and
stamping). So the transition clause remains load-bearing and narrowing waits on the drain census.
Two facts that make the flip itself safe today: **0** pending items carry a present-but-unknown
routing key, so nothing would be refused on flip day; and **0** rows in the whole table (every
status) would fail the CHECK, so it can be VALIDATEd immediately after apply rather than eventually.

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
