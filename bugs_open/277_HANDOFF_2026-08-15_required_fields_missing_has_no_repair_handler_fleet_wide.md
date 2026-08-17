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

**Same-day progression (all 2026-08-15):** seed applied and hardened to **v3** through the
council trail (corr `7b0e2833`, FOUR REVISE rounds — two of them found real design errors,
both then measured and turned into routes: `asset_sourced` — the writer must never fill
schema-declared `site_assets.*` fields, proven by a live validate_content refusal; and
`no_plan_unbuildable` — index-family pages with no plan must not be generically recreated,
proven by a live `mark_no_ready_sections` no-op). Five canaries verified per-arm incl. the
gas converter (parked `no_plan_owned`, tool pipeline named). **Producer change LIVE on
`v1.0.1302`** (another lane's roll; uniform image on all 25 chassis pods; merge-base +
literal probe with controls). Seed ledger-recorded (`--record-only`, honest note). **Fleet
assignment executed ~14:50Z**: 39/39 remaining rows routed (pre-image saved) — expected
outcome ≈34 blob-parks-with-facts + ≈5 evidence-closes, zero conversions left in the backlog.

**Council state: REVISE ×4 with seats now disagreeing with each other** (constitution
approves both the born-triaged deferral and RFC_030's deferral; improvement_guardian holds
born-triaged HIGH; reuse/architecture reject the deferral). Per the estate norm this stops
the resubmission loop; **two OWNER decisions are queued** (README_where_we_are): (1)
born-triaged vs rebuilding the disabled detected-promoter (bugs_open/083) for this type;
(2) whether/when to schedule RFC_030's three-into-one router-engine consolidation.

**OPEN** pending the fleet after-state verification (rows drain through the dispatch loop
over the following hour) + the two owner decisions. The mechanism itself is fixed AND live.

## 2026-08-17 — the churn guard at day 2, and an INDEPENDENT confirmation of the router's central judgement

All measured 2026-08-17 against the live DB; chassis `v1.0.1305` (OCI `revision=6a782274b`,
verified at the binary with positive and negative controls), which carries the born-`detected`
producer revert `3c6354059`.

### Churn guard (the +7-day check, 2 of 7 days in) — passing so far

Rows of this type created since the fleet assignment (2026-08-15 14:50Z): **exactly one**
(2026-08-16 10:02), and it went straight to `needs_human_review` carrying a route. **Zero
`unresolved`, zero `triaged`, zero unrouted.** Current all-time partition: `complete` 64,
`needs_human_review` 31, nothing else. Re-check ~08-22 before closing.

### The full chain ran end-to-end on that one new finding

It is worth naming each hop, because this is the first finding to traverse the whole mechanism
as designed rather than being back-filled by the assignment:
producer files it born-`detected` (live on the chassis) → `detected-item-promoter` promotes it
(known-good pair) → `required-fields-missing-handler` routes it (`asset_sourced`) → it parks in
the review queue carrying its classification.

### An independent mechanism reached the SAME classification on the same rows

`bugs_open/033`'s auto-drain (`revalidate_review_queue_action.go`) has since acquired a
revalidator for `required_fields_missing` and swept the parked pile at 08:45Z today. It knows
nothing about this router; it re-evaluates each parked finding against currently-deployed state
from its own premises. The two agree row-for-row:

| router's route (2026-08-15) | revalidator's verdict (2026-08-17) | rows |
|---|---|---|
| `no_content_data` — *"serves from one stored HTML block; regenerating a template section would destroy the page"* | `unknown` — *"component carries no content_data; it renders from another source"* | **29** |
| `no_plan_owned` — the gas converter, tool pipeline | `unknown` | 1 |
| `asset_sourced` | `still_holds` — *"at least one reported-missing field is still empty on the deployed component"* | 1 |

**This is the load-bearing measurement of the whole bug, and it could have come out otherwise.**
277's central design call was that 35 of the 44 findings were not "missing content" at all but
blob-served pages an automatic repair would have *destroyed* — a judgement made by reading the
data, and the thing a reviewer would most reasonably doubt. A second mechanism, written by a
different lane for a different bug, independently declines to judge exactly those rows and gives
the same reason. It did not have to: a revalidator that disagreed would have returned `resolved`
on them and auto-closed the lot.

**And the queue is now honest.** Of 31 parked rows, exactly **one** is a live, actionable
finding — the new one. The other 30 are the two classes a machine must not touch. Before this
work, all 44 were indistinguishable.

### 56 rows of this type have been auto-closed as `resolved` by that revalidator

`result.revalidation.verdict='resolved'` on 56 rows, reason *"every field this item reports
missing is populated on the deployed component"* (headline 31, headline+primary_cta 18,
content+heading 3, features+headline 3). Those closes are safe by construction: every terminal
status is excluded from `idx_swi_dedup`, so a wrong close releases the dedup key and the
producer re-raises. Worth stating plainly for anyone reading the counts: **the drop in this
type's review backlog is mostly 033's work, not this router's.** This router's contribution is
that the survivors carry their classification.

### Still open before this moves to `bugs_closed/`

1. The churn guard's remaining 5 days (~08-22).
2. The two cancelled conversions re-raising and parking — no `cancelled` rows of the type remain,
   so this now depends on discovery rotation re-filing them; if not seen by ~08-22, re-file by
   hand.
3. **Watch for the interaction, which nobody designed:** 033's revalidator now writes to rows
   this router parked. Both write `result`, and the loop's `mark_complete` REPLACES `result` on
   completed rows (this lane's landmine). Today they compose correctly — `route` and
   `revalidation` sit side by side in the same object — but that is not guaranteed by anything,
   and a future writer of either mechanism could clobber the other's evidence. Named so it is a
   known seam rather than a surprise.
