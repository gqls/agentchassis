# RFC_030 — three bespoke single-item-type work-item routers now exist; the fourth should be one engine

**Status: OPEN — filed 2026-08-15 by the bugfix_277 lane, at the council's direction.**
Raised in bug 277's round-2 review (corr `7b0e2833`), where THREE seats independently pressed
the same point: `reuse_agent` (medium — "if the same clone-and-reseed pattern repeats for
every stuck item_type, the estate ends up with N nearly-identical single-purpose router
agents instead of one parameterized router"), `architecture` (medium — "prose advice to a
future author with no enforcement mechanism … will likely be missed"), and `guardian` (low —
"worth recording as a hard tripwire, not just a note"). This file is the enforcement
mechanism: a tracked, numbered obligation in the review track, cross-referenced from CQ-023,
instead of a register sentence.

## The population (2026-08-15)

| router | seed | item_type | shape |
|---|---|---|---|
| `image-url-404-handler` | 397 | `image_url_404` | query classifier → branch → convert (`needs_imagery`) / escalate |
| `image-source-unsatisfiable-handler` | 397 | `image_source_unsatisfiable` | same |
| `required-fields-missing-handler` | 410 | `required_fields_missing` | query classifier → 6-way branch → close / park×4 / convert×2 |

All three are pure-SQL DB-config agents: a `query_database` classification, a
`conditional_branch` cascade, and arms built from `create_work_item` /
`update_work_item_status` / `checkpoint_for_review`. **The classifier SQL is the only
genuinely type-specific part.** Everything else — the cascade shape, the park/convert/close
arm mechanics, the born-triaged convention, the verify-block pattern — is copied between
seeds, and 410's three council rounds hardened mechanics (park-in-place holding the dedup
key; single-active-row assert; asset-source splitting) that 397's two routers do NOT have and
will not get except by hand-porting.

## What the engine would be

One agent type (or one Go action) taking per-item_type config: `{item_type: {classifier_sql,
routes: {label: close|park(message)|convert(item_type, handler, spec_map, mode)}}}`. The
2026-08-02 §1 producer-set rule and CQ-023's landmines (verifier-later fail-closes converts;
parked rows hold dedup keys) become engine-level guarantees instead of per-seed copies.
Whether it is a config-driven generic agent or a Go action with a registry is the design
question for the round; RFC_022's accumulated-surface concern (ten inert opt-ins nobody
counts) applies to per-type config growth here too.

## The tripwire, stated as a rule

**Whoever needs a router for a FOURTH item_type must propose this engine (or explicitly
argue it down in this file) instead of cloning seed 410.** The 033 reframing ("the framework,
not a person, resolves every one of these classes") implies many more types will want
routing: the review queue held ~20 uncovered types on 2026-08-15. Cloning has been cheap
twice; the third clone (410) already needed three council rounds to harden — the marginal
clone is getting more expensive than the engine.

## Round-3 pressure, recorded (2026-08-15)

The same review trail's round 3 pressed harder: `reuse_agent` (medium) — *"duplication is
allowed to reach three live copies before the rule bites"* — wanted the engine NOW rather
than at the fourth instance; `constitution` accepted the deferral as stated but insisted it
stay "a tracked obligation, not a closed matter"; `guardian` noted this file "does not itself
block a fourth router". All true. The deferral grounds, for whoever picks this up: the owner
ruling of 2026-08-15 ordered a handler for ONE type that morning; the engine is a design
round of its own; and the consolidation is LARGER than not-building-a-fourth — **397's two
routers lack 410's hardening (park-in-place dedup-key semantics, single-active-row assert,
asset-source splitting) and are this engine's first two migration targets**, so the engine's
scope is three-into-one, not a prophylactic for a hypothetical fourth.

## Relations

- CQ-023 (the third router; its "proliferation watch" clause points here).
- IMG-071 (the first two; their seeds lack 410's hardening).
- bugs_open/033 (the queue whose drain keeps minting candidates for this pattern).
- RFC_022 (the accumulation-counter blind spot this repeats at a different layer).
