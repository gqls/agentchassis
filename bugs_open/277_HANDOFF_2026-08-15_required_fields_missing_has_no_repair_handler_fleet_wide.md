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

---

# TAKEN UP 2026-08-15 (session "bugfix 033") — the handler is BUILT, as a router

Workstream: `docs024_key_docs_latest/bugfix_277_required_fields_repair/` (standing docs +
census + canary evidence). Register: **CQ-023**. Seed:
`sql_for_agents/410_required_fields_missing_router.sql` (+ ROLLBACK). Council submission
`7b0e2833-715f-4a9a-897b-efd913073582`.

**The population re-measured (fleet-wide, not just the 12 sites): 44 open items**, and the
census (saved, read-only) reframes the fix-shape above:

| class | n | what the router does |
|---|---|---|
| `no_content_data` — component serves 1–21KB rendered_html with EMPTY content_data (blob) | 35 | **parks in place** with the facts: auto-regeneration would REPLACE served HTML (bugs 263) |
| `stale` — page/component gone at (page_name, slot) | 6 | closes with evidence; discovery rotation re-raises if still real |
| `no_plan_generic` — sectionless, safe to rebuild | 1 | converts to `needs_content_page` / `mode=recreate`, born `triaged` |
| `no_plan_owned` — **the gas converter** (tool page, owned-page guard) | 1 | **parks** naming the tool pipeline as the repair route |
| `partial` — fields genuinely empty on populated content_data | 1 | converts to `content_rewrite` / `mode=edit_live` |

**Correction to this file's fix shape 1/2 ("re-derive the missing fields … regenerating the
page plan first"):** that is the right repair for exactly ONE of the 44 (the sectionless
generic page). For the gas converter it is forbidden by the owned-page guard (reconcile_site_plan
decision 3 — the generic builder clobbers tool pages), so the router routes it to a parked
decision naming `needs_tool_recreation`/tool-improver rather than overriding an owner ruling.
**"The page serves real content" (the Verify section above) is therefore the TOOL lane's bar,
not this handler's** — this handler's bar is: no item of this type ever again sits unread
without a classification, and every repairable one is dispatched.

Producer flipped (Go, `check_required_fields_missing.go`: born `triaged` at the router;
`handler_coverage_test.go` roster updated) — inert until the next chassis roll; the seed +
a canaried assignment UPDATE carry the live half meanwhile. Fix shape 3's register obligation:
done (CQ-023 names producer set + first consumer; PBP-028's edit_live producer-set clause
gains the third emitter).

**OPEN** until fixed-AND-live: seed applied + canary verified + fleet assigned + producer
change rolled.
