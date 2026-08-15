# 277 — `required_fields_missing` has no repair handler anywhere in the fleet, so a page with no plan strands its items at `needs_human_review` forever

## STATUS: OPEN — OWNER RULED 2026-08-15: "we should create a repair handler fleet wide." Not yet built.

Filed by the `staged_component_build` lane while closing `bugs_closed/248`, because the worked
example (`tool-gas-unit-converter`) has been parked on this gap since before 2026-08-12 and the
owner's ruling this morning turns it from an observation into a build task.

## The mechanism, plainly

When discovery or a build step finds a page whose required content fields are absent, it files
a `site_work_items` row with `item_type='required_fields_missing'`. Work items are repaired by
handler agents that claim their `item_type`. **No handler in the fleet claims this type** — so
every such item can only escalate to `needs_human_review` and sit. This is not a stalled queue;
it is a queue with no consumer, by construction.

## The evidence (measured 2026-08-15, re-measure before building)

- The worked example: `tool-gas-unit-converter` (webdesign.co.uk tools) — three items at
  `needs_human_review`, two touched by a sweep on 08-14 without repair. The page carries
  `sections=[]` and no plan, so `page-build-handler` **correctly** no-ops; the failure is
  upstream of rendering and no handler owns it. (Detail: 14c handoff §3, and
  `HANDOFF_2026-08-14c_continue_here.md`'s re-verification.)
- Fleet count of open `required_fields_missing` at `needs_human_review` on just the 12
  248-affected sites: **32** (finetuning 11, gamesdesign 8, ai-agent-orchestration 8,
  robot-hands 3, leopardess 1, idea 1) — from this session's coordination query. A fleet-wide
  count needs running fresh.

## What the owner ruled

A repair handler, fleet-wide — not a per-site hand fix. The ruling was given in chat
2026-08-15 in the context of the gas converter ("we should create a repair handler fleet
wide"), i.e. the class gets a mechanism, the instance gets repaired by that mechanism.

## Fix shape (for the building session — candidates, not a design)

1. A handler agent (or an arm of an existing build handler) claiming
   `required_fields_missing`: re-derive the missing fields from the site plan / specs and
   re-dispatch the page build — the "re-dispatch through the build pipeline" path the owner
   was previously asked to decide per-instance.
2. It must handle the `sections=[]`-and-no-plan case (the gas converter's): that likely means
   regenerating the page plan first, not just re-rendering.
3. Registration + dispatch wiring is platform code → council gate; the handler naming a new
   shared `item_type` consumer should follow the 2026-08-02 §1 register-the-producer-set rule
   (the type exists; this adds its first CONSUMER — say so in the register entry).

## Verify (when built)

The gas converter's three items go `needs_human_review` → repaired → the page serves real
content; and a fresh `required_fields_missing` filed anywhere gets claimed within the
dispatch loop's normal cadence instead of escalating.

## Relations

- `bugs_closed/248` (the lane that carried the ruling here), 14c handoff §3.
- `bugs_open/033` (human-review queue has no working surface — the queue this type currently
  dies in).
- Owner rulings 2026-08-02 §1 (producer-set registration), council-gate norms (CLAUDE.md).
