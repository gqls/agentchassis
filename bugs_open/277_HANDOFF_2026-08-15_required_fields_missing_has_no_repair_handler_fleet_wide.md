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

## COUNCIL APPROVED — trail `7b0e2833`, round 5, 2026-08-17 12:42Z

After four REVISE rounds on 2026-08-15, the trail closes **APPROVED**: 9 seats in favour, 5
abstained, not truncated, 3 advisory objections, none high. The seats that had blocked it are
among the approvers — **`editquality`**, which gated round 4 at HIGH on "a no-op dressed as an
edit"; **`improvement_guardian`**, whose HIGH objection to the born-`triaged` producer ran through
rounds 3 and 4; and **`architecture`** and **`reuse_agent`**, who had objected on the router
proliferation.

**Why this round worked where a citation round would not have.** It carried exactly **one** edit —
commit `3c6354059`, `Status: "triaged"` → `"detected"` — committed 2026-08-15 **18:16Z against
round 4's verdict at 14:41Z**, i.e. 3h35m later, so no round of this trail had ever judged the
file in its current state. Everything else that shipped since (seed 430, migration 444, the
register and bug-file records) was cited as evidence and deliberately kept **out** of the edit
list, because listing already-committed work as pending edits is precisely what gated round 4.
`improvement_guardian`'s objection was answered by code in production rather than by argument.

### The advisories, and the one that mattered

**`guardian` (MEDIUM) — the objection worth taking seriously, because its failure mode is this
lane's worst case.** It could not read `scheduled_tasks` from its seat, so it flagged: *"if the
promoter's scope doesn't cover this item_type, reverting to born-`detected` silently reproduces
the exact stranding this trail was fixing."* Checked predicate by predicate against the live row:

| the promoter's gate | value for `(required_fields_missing, required-fields-missing-handler)` |
|---|---|
| handler is a live active agent definition | **1** |
| pair has ≥1 lifetime `complete` | **14** |
| pipeline is in `444`'s allow-list | `content` ∈ `(build, content, design)` — **yes** |
| pair passes `444`'s 25% success floor | **yes** |

> ⚠ **The third row was the one worth checking on myself, not for the seat.** Migration `444` is
> this lane's own door-closer, and had this producer filed with `experience` or `maintenance` — both
> real pipeline values on this table — **my allow-list would have silently stranded the very
> findings this bug exists to route.** It files `Pipeline: "content"`
> (`check_required_fields_missing.go:163`), which is inside the list. An action that silences the
> detector it was written to protect is a documented failure family; this one came out clean, but
> only because it was measured rather than assumed.

**End-to-end proof, one row, which refutes the feared failure mode directly.** The only finding
born since the revert went live:

```
b2f1c7d4  born 2026-08-16 10:02 at status='detected'   (producer, born-detected, live on v1.0.1305)
          promoted 10:44:44 — 42 minutes, ~3 promoter ticks
          spec.original_pipeline='content' stamped, pipeline rewritten to 'build'
          routed by required-fields-missing-handler -> route='asset_sourced'
          parked at needs_human_review carrying its classification
```

It did not strand. It moved in 42 minutes, through every hop of the mechanism this trail is about.

**`prior_art_librarian` (MEDIUM)** flagged that it could not verify my claim that *this seat* had
approved the identical "sole live carrier" premise at corr `05a3d1c8`. The claim was accurate and
is checkable at the artefact: that report's approver list is `editquality, reuse_agent, guidelines,
tooling_provenance, diagnosis_guardian, improvement_guardian, render_guardian, debug_historian,
constitution, mission, **prior_art_librarian**, architecture` — 12 seats, 2026-08-17 11:27Z.

**`bug_historian` (MEDIUM)** re-raised the convert arms, and explicitly credited the plan for
naming it: *"the plan's own risk 2 names this and defers to the owner rather than gating it —
which is honest, but it means this round is being asked to approve closing a trail whose central
unexercised mechanism sits squarely inside a pattern that has burned this platform at least twice
before … with no fail-loud guard shipped alongside it."* **That is not closed and must not be read
as closed by this approval.** It is owner decision 1 in `README_where_we_are.md`, and a second
council seat has now independently said the arms want a guard.

**Trailers:** the work went out with `Council-Submitted: 7b0e2833-…`, which `098` resolves to this
approval at report time. No amend (forward-only).

---

## 2026-08-19 — measured against THIS FILE'S OWN verify criterion: half of it is met, half is not, and 277 does not close

Checked at the roll (`agent-chassis:v1.0.1314`) while closing out the `083` half of this lane. The
criterion in §"Verify (when built)" has two clauses, and they have come out differently.

**Clause 2 — MET.** *"a fresh `required_fields_missing` filed anywhere gets claimed within the
dispatch loop's normal cadence instead of escalating."* The router is live (1 active definition),
and the type moves: **130 complete / 30 needs_human_review** lifetime, with handler activity as
recent as **2026-08-19 08:45**. Every one of the 30 parked rows carries a route —
`no_content_data` 27, `asset_sourced` 2, `no_plan_owned` 1, **zero unrouted**. Nothing strands
unclassified any more, which is what this bug was filed about.

**Clause 1 — NOT MET.** *"The gas converter's three items go `needs_human_review` → repaired → the
page serves real content."* The gas converter's item is still at `needs_human_review`, routed
`no_plan_owned`, updated today:

```
required_fields_missing:7e576bc4-fb8b-46a4-b035-2842c481f35a:tool-gas-unit-converter
  status=needs_human_review  route=no_plan_owned  updated 2026-08-19
```

**It has been classified, not repaired.**

### The general form, which is the part worth acting on

**Nothing repairs a `required_fields_missing` item.** Of the completions in the live table:

| how it reached `complete` | rows |
|---|---|
| `resolution_path='auto:revalidated'` — a sweep found the defect had gone | **44** |
| `handled_by='build-dispatch-loop'`, no resolution path | **37** |
| repaired by the router | **0** |

The 44 are the honest number to look at: those items closed because the page acquired content **by
some other route**, and the sweep noticed. That is the mechanism doing its job as an accountant, not
as a repairer — and it is why the queue looks healthier than the pages are.

**This is not a criticism of the router, which does exactly what it was built and approved to do**
(corr `7b0e2833`, round 5). It is that the owner's ruling — *"we should create a repair handler
fleet wide"* — has been half-delivered: the routing half exists and is proven; the repairing half
does not exist for the largest route, `no_content_data` (27 of the 30 parked).

### So: 277 stays OPEN, and this is what "done" would need

1. Something that acts on `no_content_data` — the 27-row majority. That is a content-acquisition
   problem, not a routing one, and it is plausibly the same missing piece as the owned-page repair
   route (`bugs_open/301`, Tier 2): **a finding-to-edit converter**. ⚠ Note `copy-editor` already
   produces `apply_section_edit`'s exact input shape and is owned by the
   `loanandmortgagecalculator_couk` lane — talk to them before designing anything.
2. The worked example served: `tool-gas-unit-converter` carrying real content, checked at the served
   page and not at the item's status (`bugs_closed/287`: a `complete` work item is not a repaired
   artefact — and here the items are not even complete).
3. Only then does clause 1 pass.

> ⚠ **CORRECTED 2026-08-19 — `copy-editor` is owned by the `copy_quality_two_stage` lane, NOT `loanandmortgagecalculator_couk`.** I got the wrong lane from a `grep -rl "copy-editor"` hit in LMC's `README_where_we_are.md` — a *mention* — and read it as ownership. `scripts/who-owns.py` exists to separate those two, and I did not run it. The defining evidence is what the commits shipping migrations `447`/`462` actually touch: `docs024_key_docs_latest/copy_quality_two_stage/`. Register entry **CQ-024**. A CONTRIB is filed in their lane dir (`CONTRIB_2026-08-19_from_the_277_083_lane_…`, commit `7574482c7`).


**Closing this on clause 2 alone would be exactly the error this estate keeps logging:** the queue
is tidy, every row is labelled, and the page the bug was filed about still does not serve content.

## CONTRIB 2026-08-19 ~21:30Z (from the `bugfix_301_owned_guard_ordering` lane) — the PRODUCER half of your "no route at all" finding is now `bugs_open/333`; this file keeps the ROUTE half

Your 08-19b handoff (21:00Z) reached the owned-page residual from the repair side: *"an owned
page with a real, mechanically-repairable defect has no route at all — the generic repair refuses
it and nothing else claims it."* The same evening the owner ruled on 301's close-out residual that
**the routing defect and the repair-design question are different bugs**, and the routing half is
now filed: `bugs_open/333_HANDOFF_2026-08-19_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy_so_owned_pages_queue_findings_that_can_only_be_refused.md`.

What 333 holds that touches you directly, measured not inferred: your converter
(`required-fields-missing-handler`, steps `file_rewrite`/`file_recreate` → `create_work_item` at
`page-build-handler`) filed **28 `content_rewrite` items on owned pages on 08-18 alone** — the
newest large producer of the class, and `create_work_item` → `writeWorkItem` reads no policy
either. 333's preferred fix is a **policy-routability check at the door of `writeWorkItem`,
beside 291's registration probe** (same `pageIsOwnedForGuard` predicate; demote, never refuse;
per finding, not per page) — which would make your converted items on owned pages land visibly
as "no route" rather than `wont_fix`. **It does NOT decide what route repairs an owned page** —
that stays here (and in the Tier 2 / `copy_quality_two_stage` exchange). You are named in 333 as a
consumer to tell; this note is the telling. If you build the 277-side route first, 333's check
should name it as the owned-page handler rather than demoting; say so in 333 if you do.

## 2026-08-20 — "an owned page has NO route at all" is TOO STRONG. There is a candidate, it is 92% on owned pages, and the landmine that should have killed it does not apply

**This narrows clause 1's blocker from "nothing exists" to "one thing exists and one question
remains", which is a materially different bug.** The 08-19 evening finding (recorded in the lane
NOTES and in `LANDMINES.md`) concluded that *an owned page carrying a real, mechanically-repairable
defect has no repair route at all — the generic repair refuses it and nothing else claims it.*
The first half is confirmed below at n=7. **The second half is wrong.**

### 1. The route was already named, in this platform's own config, and nobody had picked it up

`migration 466`'s `what_to_do` text — the prose the escalation task hands a human — already says it:

> *"If protective refusals dominate, the handler is behaving CORRECTLY … the PAGE needs a different
> ROUTE. See `bugs_closed/295` … **Its fix candidate 3 is UNTOUCHED and is the live remedy: route
> content findings on owned pages to `section_edit`, which demonstrably works on them (18
> completes).** ⚠ `apply_section_edit` is right for REWRITING an existing component and a DEAD END
> for ADDING a section to an owned page."*

**[MEASURED 2026-08-20, live + archive] the figure is now better than the one quoted:**

| `section_edit → section-editor`, by `pages.rebuild_policy` | complete | failed | total |
|---|---|---|---|
| **`owned`** | **36** | **1** | **39** |
| `generic` | 53 | 4 | 57 |
| (no page row) | 132 | 0 | 133 |

**92% on owned pages, lifetime.** Compare the generic repair on the same axis: `literal_markdown →
page-rerender` is **8 complete on `generic`, 1 failed on the single `owned` page it tried.** The two
routes are not both blocked on owned pages — one is refused by design and the other is the estate's
established way of editing them.

### 2. The landmine that should have killed this was checked, and its precondition is ABSENT

`LANDMINES.md` carries a severe entry: *a `section_edit` on a per-site TOOL FORK whose template
carries `{{.field}}` copy and whose `content_data` is `{}` re-renders every text node to EMPTY —
the class-attribute floors PASS, the item completes, and the raw-tag check reads clean.* **Six of
the seven pages in our population are `tool-*`**, so this is a direct hit on shape and would have
made "use `section_edit`" another confidently-actionable wrong answer.

**It does not fire, and the reason is the useful part — the trap needs BOTH halves.**
[MEASURED 2026-08-20, `page_components` ⋈ `content_components`]:

| page | slot | `component_level` | `content_data` empty | `{{.field}}` hits in template |
|---|---|---|---|---|
| all 7 | `ported-page` | **section** | **no** | 1 |
| grid-generator / json-cleaner / noise-generator | `tool-*` | **tool** | **yes** | **0** |
| cubic-bezier / head-architect / text-extractor / learn-design-physics | *(single component)* | section | no | 1 |

The `component_level='tool'` forks do have `content_data = '{}'` — but **zero** `{{.field}}` hits in
their templates, so there is nothing for an empty `content_data` to fail to fill. **And the literal
markdown is not in the tool fork at all:** it is in the `ported-page` slot, a `section`-level
component whose `content_data` **is** populated — the ordinary, well-trodden target, i.e. the 36/1
population above. The caveat in 466's own warning (`section_edit` REWRITES, cannot ADD) also lines
up: literal asterisks in existing prose are a rewrite.

> ⚠ **The first check I ran for this was NEARLY VACUOUS, and it is recorded so nobody repeats it.**
> I grepped `page_components.rendered_html` for `{{.` and got 0 on all seven pages, and briefly read
> that as "the trap does not apply". **It proves nothing:** `rendered_html` is the RENDERED OUTPUT,
> so a template field that resolved to empty leaves no `{{.` behind *either*. The measurement would
> have returned 0 whether or not the risk existed. **The template is the only place the question can
> be asked** — which is why the table above joins `content_components`.

### 3. The first half IS confirmed, at n=7, by the population repairing itself out from under us

**[MEASURED 2026-08-20 08:11Z]** The 7 held `literal_markdown → page-build-handler` rows are gone
from the held set. Between **07:20:42Z and 07:23:58Z** every one was dispatched, refused by
`OWNED_PAGE_GUARD`, and terminated **`wont_fix`** (`owned_page_refusal: true`,
`owned_page_refusal_replaced_status: "failed"`). Why they were released: the pair rose from
**3 ok / 34 failed (8.1%, floor-held)** to **19 ok / 24 failed (44%, promotable)** — 16 completions
across 3 sites inside the 07:00Z hour — so the promoter fed it, and the owned-page rows in it went
straight into the guard.

**So the 08-19 prediction "re-pointing them at a generic repair would produce 7 more failures and
repair nothing" was tested without anyone re-pointing anything, and it held at 7 of 7.** The one
half that did *not* materialise is "drag a healthy pair toward its floor" — `301` makes that
mechanically impossible, since `wont_fix` is excluded from both sides of the promoter rule.

**And that exclusion is the thing to notice**: the full cycle — released → refused → `wont_fix` →
re-filed by the detector (`idx_swi_dedup` excludes `wont_fix`) — **leaves no trace in the pair's
record**, so a healthy-looking ratio on `literal_markdown → page-build-handler` is *not* evidence
that owned pages are being repaired. Recorded as a CONTRIB in `bugs_open/333`, which owns that seam;
both properties were already known there and in `bugs_closed/301`, so this is the measured instance,
not a new claim.

### 4. What clause 1 now needs — ONE question, not a design

**The remaining unknown is whether a producer can file a `section_edit` item for a
`literal_markdown`-shaped finding at all** — what `spec`/`field_updates` it would carry, and which
`page_component_id` it targets (on these pages, the `ported-page` one). That is a code question in
`section_editor_actions.go` and the producers, not a design question. Three landmines apply to
whoever answers it and all three are already written down: the `field_updates` merge is **per-field
and reverts intervening edits**; `apply_section_edit` writes `rendered_html` with **no content
validation**; and `apply_section_edit` **cannot ADD** a component.

**Do not read this as clause 1 being closed.** Nothing has been repaired yet, and `no_content_data`
(27 of the 30 parked) is a different and larger hole — the generic-repair-refuses-owned-pages
finding does not touch it. What has changed is that clause 1's blocker is no longer "no route
exists".

*Escalation-task config updated to match, so the next human to read it is not told the refuted
version: migrations `497` (the owners map pointed at three dead destinations) and `498`
(de-volatilising `497`'s own figures, which went stale in twelve hours).*
