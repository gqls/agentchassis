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

> # ⚠⚠ CORRECTED SAME DAY, ~09:30Z — READ §5 BELOW BEFORE ACTING ON THIS SECTION
>
> **The 36/1 measurement is right. The TARGET named in §2 is wrong, and it is wrong for the
> population this file is about.** All 7 findings are `source: rendered_html`, and the
> `ported-page` component's `content_data` holds **no prose** — so `section_edit` +
> `strip_literal_markdown`, which strips `content_data`, **cannot reach them**. §2's table asks
> whether the tool forks were safe; it never asks the prior question, **which is whether the
> component I was pointing at can be re-rendered at all.** It cannot: its template's only field
> is `{{.body}}` and `body` is not a key.
>
> **What survives:** `section_edit → section-editor` really is 36/1 on owned pages and really is
> the right route where `content_data` can fill the template. **What is retracted:** that it is
> the route for *these* rows, and that clause 1's blocker had narrowed to "one code question".
> **What caught it:** reading the items' actual `findings` array — which I had not done — while
> checking whether a producer could file a `section_edit`. Full evidence, and the render measured
> against production's own engine, in §5.

**This narrows clause 1's blocker from "nothing exists" to "one thing exists and one question
remains", which is a materially different bug.** ~~(Retracted — see the box above and §5.)~~ The 08-19 evening finding (recorded in the lane
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

## 5. 2026-08-20 ~09:30Z — CORRECTION to §§1–4 above: the content is in `rendered_html`, so NO content_data route reaches it, and clause 1's blocker has NOT narrowed

**§4 said the remaining unknown was "whether a producer can file a `section_edit` item". Answering
that question is what refuted §2.** The producer side turned out to be easy — and then reading the
items' own `findings` array, which I had never done, made the whole route moot.

### 5.1 What the findings actually say [MEASURED 2026-08-20 ~09:30Z]

| page | `source` | `pattern` | `slot` | `field` | `matched` |
|---|---|---|---|---|---|
| learn-design-physics-of-ui | **`rendered_html`** | `code_span` | `ported-page` | *(empty)* | `` `ease` `` |
| tool-cubic-bezier | **`rendered_html`** | `code_span` | `ported-page` | *(empty)* | `` `ease-in-out` `` |
| tool-grid-generator | **`rendered_html`** | `code_span` | `ported-page` | *(empty)* | `` `33%` `` |
| tool-head-architect / json-cleaner / text-extractor | **`rendered_html`** | `code_span` | `ported-page` | *(empty)* | `` `fetch()` `` |
| tool-noise-generator | **`rendered_html`** | `code_span` | `ported-page` | *(empty)* | `` `feTurbulence` `` |

**All 7 are `rendered_html`, not `content_data`.** The detector scans both surfaces —
`literalMarkdownFinding.Source` is documented `content_data | rendered_html`, and `Field` is
*"empty for rendered_html"* (`check_literal_markdown.go:135-140`). These are backticked code tokens
in ported technical prose, i.e. the mildest form: markdown that should have become `<code>`.

### 5.2 Why that kills BOTH routes this lane has proposed

The `ported-page` component's `content_data` is **215 bytes of metadata** —
`{schema, sha256, source, qa_tier, generator}` — and the template's **only** field is `{{.body}}`,
which is **not a key**. So:

- **`473`'s rerender** regenerates from `content_data` → nothing to regenerate from;
- **`section_edit` + `strip_literal_markdown`** calls `StripLiteralMarkdownFromContentData` on that
  same map → **no prose to strip**.

**Both are inapplicable BY CONSTRUCTION, not by policy.** This is the lane's own memory rule — *a
repro regenerated from source is destroyed by the render; it cannot reproduce a defect living in
`rendered_html`* — arriving as a **repair** problem rather than a reproduction one.

### 5.3 Proven, not reasoned — and the control could have come out the other way

Rendered the real template against the real `content_data`, with production's own engine and option
(`text/template`, `Option("missingkey=zero")`, `call_agent.go:1171` (`executeGoTemplate`, reached via `component_library.go:1062`)):

| payload | rendered | body region | visible non-whitespace chars | err |
|---|---|---|---|---|
| **owned** (`ported-page`, no `body`) | 4,665 B | 188 B | **0** | `<nil>` |
| **generic** control (has `body`) | 11,035 B | 6,568 B | full prose | `<nil>` |

Same component, same template, two real payloads, opposite results. The body region of the owned
render is `<section><div><article>` and nothing inside. **Note `err=<nil>` both times** — the empty
render *succeeds*.

### 5.4 What is NOT true, stated because I nearly wrote it

**This is not "100 pages are one edit from being blanked."** `apply_section_edit` calls
`enforceSingleSlotFloors` (`section_editor_actions.go:451` → `single_slot_floors.go:161`), whose axis
is **visible** text with style/script content excluded, engaging above `minShrinkGuardVisibleChars`
(200) on the existing side. Thousands existing against **zero** incoming → refused, *"nothing was
written and the existing component still stands."* The outcome would be a **third refusal mode**,
not damage.

### 5.5 The fleet census, which explains the whole pattern

`Ported Page (webdesign.co.uk)`: **115 instances; 100 lack the `body` key.** The split is **total**:

| site | policy | instances | missing `body` |
|---|---|---|---|
| webdesign.co.uk | **owned** | 97 | **97** |
| loancash.co.uk | generic | 15 | **0** |
| loancash.co.uk | **owned** | 3 | **3** |

**Every missing-`body` instance is `owned`; every instance that has it is `generic`.** That is why
generic `literal_markdown` repairs complete (16 did, in one hour this morning) and owned ones never
can. The operative property was never ownership — it is that the owned ported pages carry **no
regenerable content**. Ownership and un-regenerability happen to coincide here, and mistaking the
first for the cause is what produced three wrong routes in two days.

### 5.6 Where clause 1 actually stands

**Retract "the blocker has narrowed to one code question."** For this population it has not moved:
a `rendered_html`-only defect on a component with no regenerable source has **no route**, and the
08-19 evening claim was right about *these rows* even though it was too strong as a general law.

What a repair would need is an **HTML-level transform on `rendered_html`** (`` `x` `` → `<code>x</code>`),
which nothing currently does — a different shape from every route considered so far. Whether that is
worth building for 7 findings of the mildest pattern is a judgement for the owner, not an obvious
yes: these are backticks in developer-tool prose, not broken pages.

*Live config corrected again so no human is handed the wrong target at an escalation: migration
`499` replaces `498`'s named target with the TEST — read the finding's `source`, then ask whether
`content_data` can reproduce `rendered_html`.*

---

## 6. 2026-08-20 (later) — OWNER RULED YES on §5.6, and the route is BUILT

**Owner, in chat, same day:** *"Do those seven findings get a repair route? Building one means a
transform that edits finished HTML directly - I think yes."*

Built and committed the same session (register **CQ-028**, council corr `b72a4029`, migration `513`
applied + round-tripped): `apply_section_edit` gained an opt-in `rendered_html_transform` edit type
carrying `datahelpers.ConvertLiteralCodeSpansInHTML` (`` `x` `` → `<code>x</code>`, byte-splice,
detector's skip set, conversion strictly ⊆ detection), and `check_literal_markdown` now routes a
page to `section-editor` when — and only when — every finding is `source=rendered_html` ∧
`pattern=code_span` ∧ one once-occurring slot ∧ `ContentDataCanFillTemplate` is false (migration
499's test, automated). Everything else keeps today's route. Design and evidence:
`docs/agent_docs/docs024_key_docs_latest/bugfix_277_required_fields_repair/PLAN_2026-08-15_required_fields_router.md`
(2026-08-20 addendum).

### What clause 1 now waits on, in order
1. **The chassis roll** carrying the code (detector routing + action branch are one image).
2. **The detector's next sweep** over webdesign.co.uk files fresh `literal_markdown` items at the
   new shape (all 7 old rows are terminal, dedup free — checked).
3. **ONE CANARY, deliberately dispatched.** The new pair `literal_markdown → section-editor` has
   zero lifetime completes, so the 444 promoter's ≥1-complete door HOLDS it. Promote one row by
   hand (the 083 precedent, commit `8d77196ad`), watch it end to end, and **verify at the served
   bytes**: `curl` the page — `<code>` present, backticks gone, the tool's own `<script>` template
   literals untouched. That single completion opens the door for the remaining six.
4. Then clause 1's worked example exists and this file can weigh closing against the
   `no_content_data` half (§0), which none of this touches.

---

## 7. 2026-08-21 — **CLAUSE 1 IS MET: the worked example is REPAIRED and proven at the served bytes**

All four steps of §6 have now happened. Evidence, in the order the sequence required it:

**1. The roll.** Chassis pods carry `buildinfo.GitCommit=0483e7f4e…` and
`git merge-base --is-ancestor af0f00bb5 0483e7f4e` → YES. Binary probe with both controls:
`rendered_html_transform` 8, `code_span_to_code_tag` 5, `OWNED_PAGE_GUARD` 3 (positive),
`ZZQQ_NEEDLE_THAT_MUST_NOT_EXIST` 0 (negative). Config half at the live column: flag `true`,
`transform_name` whitelisted.

**2. The sweep — and it had to be FORCED, which §6 did not anticipate.** `literal_markdown` is only
ever run by `site-discovery-rotation-quality`, whose `pre_query` takes `LIMIT 1` site with
`last_selected_at < now() - 7 days`. webdesign.co.uk was stamped 2026-08-18 07:23Z and **the whole
rotation was idle** (oldest site in the fleet at 5d 01h), so its next natural sweep was
**≈2026-08-25 07:33Z** [CALCULATED from measured stamps]. Owner approved forcing it; a one-shot
task (no `pre_query`, so the rotation stamp was not consumed) fired at 13:19:01Z and the run is
confirmed at `orchestration_states 4cfdca1f-…` COMPLETED 13:19:15Z — a stamp is not a run.

**3. What the router filed: 8 rows, every one at the new shape** — `section-editor`,
`edit_type=rendered_html_transform`, `transform_name=code_span_to_code_tag`. No check filed
anything else. The 7 old `wont_fix` rows did not block re-filing (`idx_swi_dedup` excludes
`wont_fix` and `unresolved` — read at the index definition).

**4. The canary, then the door.** One row promoted by hand at 13:21:42Z (`ecd947c2…`,
tool-cubic-bezier, the `` `ease-in-out` `` finding, 444's own promote UPDATE). Claimed 13:24:18 →
**complete 13:25:03Z**, `result._verification` = *verified — no literal markdown on either surface
across 1 component(s)*. **The promoter released the other six on its very next tick (13:27),
unaided** — which is the claim §6 could only assert.

### The proof is the served bytes, and the control is the half that carried the risk

| check | before (13:22Z) | after (13:25Z) |
|---|---|---|
| backticks on the page | 6 | **4** (−2, exactly one span) |
| `` `ease-in-out` `` | 1 | **0** |
| `<code>ease-in-out</code>` | 0 | **1** |
| **backticks inside `<script>`** — the tool's own template literals | 4 | **4, untouched** |
| page bytes | 16683 | 16694 (**+11** = `<code></code>` 13 − 2 backticks) |

`diff` of the two cache-busted fetches changes **one line pair and nothing else on the page**. The
falling total alone would NOT have been a pass: a transform that ate the JS template literals shows
the same downward move, which is why the script count is quoted beside it.

### One row of the eight was born dead, and it is not this route's fault

`learn-index` (`2c4033b0…`) was filed **`unresolved`** — terminal, never dispatchable — labelled
*"[unresolved after 2 attempts]"*. Cause: `writeWorkItem`'s two-strike rule
(`load_work_item_actions.go:1373-1408`) counts `complete`/`failed` rows for the same
`(site_id, item_key)` over 7 days, and `item_key` is handler-agnostic **by design**. Its two strikes
are `46f356cf` (failed, page-build-handler, 08-14) and `6865c4b9` (failed, page-rerender, 08-18) —
**both routes this file has already shown to be inapplicable by construction.** So a re-route
inherits the strikes of the route it replaced, and the label says "tried twice" about a repair tried
zero times. CONTRIB filed into `bugs_open/333`, whose §2 reaches the same rule down a different road.
It self-heals on the rolling window (both strikes age out before the 08-25 sweep) — **do not hand-
flip the row**; the prediction is disconfirmable and worth more than the one page.

### Where 277 stands after this

- **Clause 1: MET.** A page in the `no_content_data` population has been repaired, mechanically, no
  LLM, and verified at what the visitor is served.
- **The `no_content_data` half is UNTOUCHED** and is what still holds this file open — 27 of 30
  parked rows, a different agent, and `473`'s deterministic route does not cover it. Nothing in
  today's work bears on it, and the good news must not be allowed to bleed across (§0's standing
  warning).

**Final tally 13:37Z: 7 of 7 dispatchable rows `complete`, all `verified`, all with zero prose
backticks at the served page and their `<script>` literals intact** — `tool-head-architect` kept
**44** script backticks while reaching zero in prose, which a transform leaking into script context
could not do. Per-page table: NOTES 2026-08-21 §7.

---

## 8. 2026-08-21 (later) — post-roll re-verification, and a CORRECTION to this file's own account of the remaining half

### 8.1 The new chassis build does not disturb clause 1 — checked at the artefact, not inferred

A fresh image rolled at **16:54Z** (`sha256:68075cf5…`, stamp `bac189921`). `af0f00bb5` and
`6011f9657` are both ancestors of it and `0483e7f4e → bac189921` is forward with no revert — but
ancestry cannot prove a later commit did not delete the code, so the capability was re-probed on the
NEW binary: `rendered_html_transform` **8**, `code_span_to_code_tag` **5**, negative control **0**.
The seven repaired rows are untouched (`complete`, last write 13:37:02Z) and nothing new has been
filed for the type.

### 8.2 ⚠ CORRECTION — "`no_content_data` … is a content-acquisition problem" is TOO STRONG for most of the population

§"So: 277 stays OPEN" says the 27-row majority *"is a content-acquisition problem, not a routing
one"*, and points at a finding-to-edit converter. **For most of these rows the content already
exists — it is on the page, just not in `content_data`.** Worked case, read at the row rather than
reasoned:

- Item: *"Component 'hero' on page `tool-ttk-calculator` is missing 1 schema-required value field(s):
  **headline**"*, route `no_content_data`, parked by `park_blob`.
- `page_components.77eaa64e…`: `content_data` **NULL**, `rendered_html` **16,106 bytes**, and its
  first element is `<h1>Time-To-Kill (TTK) Calculator</h1>` — i.e. the "missing" headline is being
  served to visitors right now. The page itself returns 200, 31KB, with no placeholder or
  unrendered-template marker anywhere.

**So these are not 27 broken pages.** They are 27 components whose stored data cannot reproduce what
they serve — the same property this lane measured for the Ported Page population, seen from the
other side.

### 8.3 How much of it is recoverable — MEASURED across all 27, 2026-08-21 ~17:05Z

| missing field(s) | rows | `content_data` empty | `rendered_html` > 200B | component has a real `<h1>` |
|---|---|---|---|---|
| `headline` | 18 | 18 | 18 | **15** |
| `headline, primary_cta` | 6 | 6 | 6 | 0 |
| `features, headline` | 2 | 2 | 2 | 0 |
| `content` | 1 | 1 | 1 | 0 |

**15 of 27 (56%) are recoverable by the single most obvious rule** (the component's own `<h1>` is the
headline). The other 12 are not — 3 `headline`-only rows carry no `<h1>` at all, and the 9 multi-field
rows would need a rule per field (`primary_cta`, `features`, `content`), which is a different and
much less deterministic job. **This is a COVERAGE figure, not an agreement figure**: it says how
often the cheap rule *resolves*, which is the question a plain wire cannot answer (WRONG_CALLS,
2026-08-20).

### 8.4 The hazard that makes this an OWNER decision rather than an obvious yes

A backfill would make these components **regenerable again** — which is the whole point, and also the
risk. `HANDOFF_2026-08-20b` §3 already named it: *"If someone later BACKFILLS `content_data` on ported
pages, the regenerate routes wake up and could reprint pre-transform content — whoever does that owns
re-checking these 7."* Today's seven repairs live in `rendered_html` only, **by design**, precisely
because nothing regenerates those components. A backfill removes that protection.

So the three options, costed honestly:

1. **Recovery/backfill** (extract the field from the component's own `rendered_html`). Cheapest, and
   deterministic for the 15. Owns the re-check of today's seven, and needs a rule for the other 12 or
   an explicit "these stay parked".
2. **Leave them parked with the facts.** Defensible as a terminal state: no visitor sees a defect,
   every row is labelled with why, and the route classification is correct. Then 277 closes on the
   ground that routing is delivered and clause 1 is proven — and the residual becomes a data-model
   debt filed elsewhere, not an open repair bug.
3. **Build the finding-to-edit converter** this file originally proposed. Right for genuine
   acquisition cases; **overkill for the 15**, and it is the expensive one.

**No session should pick between 2 and 1 on its own** — option 2 changes what "fixed" means for this
bug, and option 1 spends the protection today's repair relies on.

### 8.5 ⚠ CORRECTION to §8.4, 2026-08-22 — the hazard I attached to option 1 does NOT arise for THIS population, and that makes option 1 cheaper than I costed it

§8.4 says a backfill *"re-enables the regenerate routes on exactly the components whose
un-regenerability makes today's repair safe"*, and therefore *"owns re-checking the seven"*.
**I asserted the overlap instead of measuring it. Measured 2026-08-22 09:2xZ, it is ZERO.**

Where the 27 parked `no_content_data` rows actually live:

| site | parked rows |
|---|---|
| finetuning.uk | 10 |
| gamesdesign.co.uk | 8 |
| ai-agent-orchestration.com | 5 |
| gaswholesalers.com | 3 |
| mortgagecalculator.co.uk | 1 |

**Not one is on webdesign.co.uk**, and a direct join of the parked rows' pages against the seven
repaired pages returns **0**. The two populations do not touch.

**What survives, and what does not.** The *general* trap stands exactly as written in
`HANDOFF_2026-08-20b` §3 — backfilling `content_data` makes a component regenerable, and a component
that also carries a `rendered_html`-only repair can then be regenerated back to its pre-repair
content. What does **not** survive is my application of it here: repairing these 27 does not put
yesterday's seven at risk, and whoever does it does **not** inherit re-checking them.

**So the real control is SCOPE, not abstention.** A backfill written for *these 27 rows* is safe. A
backfill written as a general "fill any empty `content_data` fleet-wide" mechanism would eventually
reach webdesign.co.uk's ported pages, which DO carry the repair — and that is the version that owns
the re-check. Anyone building this should say which of the two they are building, in the commit.

### 8.6 2026-08-22 — THIS FILE'S OWN WORKED-EXAMPLE CRITERION IS MET, and it was met without anything being repaired

§"So: 277 stays OPEN" sets three things "done" would need. Item 2 is:

> *"The worked example served: `tool-gas-unit-converter` carrying real content, checked at the served
> page and not at the item's status."*

**Measured 2026-08-22 09:3xZ at the served page.** `https://gaswholesalers.com/tools/tool-gas-unit-converter.html`
→ **200, 23,774 bytes**, a **6-row** conversion table, and every one of the values the finding calls
missing is rendered: `Dekatherm` ×1, `MMBtu` ×6, `therm` ×9, `MWh` ×2. The row itself says
*"missing 9 schema-required value field(s): reference_table_heading, section_heading,
section_subheading, table_note, table_row_dekatherm_desc, table_row_gj_desc, table_row_mmbtu_desc,
table_row_mwh_desc, table_row_therm_desc"*.

**Both statements are true at once, and that is the whole point:** the finding is correct about the
DATA (those fields are genuinely absent from `content_data`) and the fear behind the criterion — that
the page therefore does not serve content — is FALSE. The page has been serving its table since it
deployed on 2026-08-15.

**What this does to the file's logic.** Criterion 1 (*"something that acts on `no_content_data`"*)
exists in order to produce criterion 2. Criterion 2 is already satisfied. So what is left is not "the
page is broken and nothing can fix it" but **a data-model debt with no visitor-facing symptom**: 27
components that serve correctly and cannot be rebuilt from their own stored data.

That is a different — and smaller — thing than this file has been describing since 08-19, and it is
the fact the close/no-close decision should actually turn on.

> ⚠ **Misstep worth recording, because it nearly became a finding.** I first fetched
> `/tools/gas-unit-converter/index.html`, built by analogy with webdesign.co.uk's URL shape, and got a
> **404 "This page has gone missing"**. Had I written that down I would have reported the worked
> example as a dead page and sent the next session chasing a deploy bug that does not exist. **A 404
> on a URL you CONSTRUCTED is evidence about your guess, not about the site.** The real URL is in
> `pages.url` (`/tools/tool-gas-unit-converter.html`) — one query, and it is the only acceptable
> source for a page address on a fleet whose sites do not share a URL convention.
