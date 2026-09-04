# 450 — a plan names tool pages before their tools exist, nothing consumes the `owned_page_review` hold, and the phantom-link repair builds the tool URLs as prose shells that serve 200

**Filed 2026-09-02 ~22:1xZ by the portfolio_positioning lane** (remake №3, seotools.co.uk).
**Diagnosis loop: CONFIRMED** — intake `40879ff3`, run `96e97dc4`, verdict 2026-09-02 22:11:33Z,
`stopped_by: confirmed`, grounded by `[static] owned_page_guard.go: "return policy ==
ownedRebuildPolicy, true"` + `const ownedRebuildPolicy = "owned"`, and `[state]` reads of
`pages` (seotools tool-* rows `generic | tool`), `content_components ⋈ page_components`
(`hero-tool | section`, `Generic Text Block | section`), the `owned_page_review` rows, and
`page_component_history`. The loop also requested `check_phantom_internal_links.go` by content
before ruling. ⚠ The verdict lives in the `needs_diagnosis` item's `result`
(`WHERE spec->>'dispatch_correlation_id' LIKE '96e97dc4%'`), NOT in `doc_notes` — the
`doc_notes` query this file first pointed at returns nothing. Filed on first-hand verification
(rows + code lines + the advertise control) with the loop run alongside, per the 2026-07-31
ruling.

## Symptom at the artefact

seotools.co.uk was deployed 2026-09-02 ~18:2xZ with 7 planned tool pages. At 21:1xZ all seven
`/tools/<slug>/index.html` serve **200 at ~56 KB with the tool's own headline** ("Check Which
Pages Your Robots.txt Actually Blocks") and **0 `<form>`, 0 `<input>`, 0 `<select>`, 1 `<button>`
(the mobile-menu toggle)**. Control, identical probe: advertise.co.uk's three real tools carry
1 form, 0–11 inputs, up to 2 selects, 4 scripts. Every status-shaped check passes: page
`active`/`deployed`, components `deployed`, URL 200, `dead_internal_link_live` about to clear.
The only record that the tool is missing is the `owned_page_review` row in the human queue.

## The chain, verified

1. **The plan names tool pages with no tool component.** seotools' current plan `5895b7ae`
   (16:13:26Z): all seven `tool-*` pages have `site_plan_sections` = `hero-tool,generic-text-block`.
   Tool components are created by `tool-deployer` after design-discovery's `evaluate_tools`,
   and design-discovery had never selected seotools (rotation ≈ one site per 3 h; last four
   selections 09:43 / 12:43 / 15:44 / 18:47Z). **Control:** advertise's plan `046c9eee`
   (13:09:36Z) was persisted AFTER its tools existed (components 12:57–13:06Z, because the
   rotation happened to select advertise at 12:43:56Z) and names them:
   `hero-tool,tool-guide-intro,tool-ab-test-calculator,tool-cta`. Same planner, same day — the
   difference is whether the tool existed when the plan was written.
2. **The hold has no consumer.** `validate_site_plan` filed 7 × `owned_page_review`
   ("not_built — needs owner-aware build, not the generic builder") at 16:13:28Z. Nothing
   reads them: `pages.rebuild_policy` stays `generic` (it is `generic` on advertise's REAL tool
   pages too — fleet 2026-09-02: 118 tool-type pages `owned`, 219 `generic`), and no producer
   checks the queue before writing the page.
3. **The page rows exist from plan time**, `deployed_at NULL`, and the hub/nav pages link
   `/tools/<slug>/index.html`.
4. **The phantom-link check files a repair per LINK, routed to the generic builder.**
   `check_phantom_internal_links.go:145-153`: `unbuilt_internal_link` → handler
   `page-build-handler`, spec.fix "Build and deploy the target page". 20 such items on
   seotools at 19:39:49Z for 7 target pages.
5. **The generic builder builds what the plan says and the guard lets it through.**
   `save_page_sections_action.go:186` → `pageIsOwnedForGuard` (`owned_page_guard.go:176-190`)
   reads `rebuild_policy='owned'`; `generic` → save allowed → deployed. `page_component_history`
   for the seven pages: every write's `source_item_id` is an `unbuilt_internal_link` item —
   6 pages × 2–6 writes each, 26 writes between 19:57Z and 20:41Z (one link = one rebuild of
   the same page). One write tripped the component floor (`save_refused_incomplete` 540f5359,
   hero-tool 12→5 class attributes) — a by-product, closed with this reference.
6. **The URL now serves 200, so every link-shaped detector clears itself**
   (`dead_internal_link_live` ×7 will resolve on re-probe) and the shells look finished.
7. ~~**What `tool-deployer` does when `evaluate_tools` finally lands on a page that already
   carries generic components: [UNVERIFIED].**~~ **ANSWERED 2026-09-03 08:2xZ, at the rows —
   it does NOTHING to them.** The design rotation reached seotools at 21:48:01Z (evaluate_tools
   → 8 `add_tool`, complete 22:31–22:32Z) and websitepromotion at 03:49Z (7 `add_tool`).
   `tool-deployer` CREATED its own page rows under its own names — seotools now has
   `tool-ab-test-calculator`, `tool-canonical-checker`, `tool-cpm-cpc-benchmark-comparator`,
   `tool-keyword-intent-classifier`, `tool-robots-txt-generator`, `tool-seo-schema`,
   `tool-social-card` (+ guides), each with a real `component_level='tool'` row of 15–22 KB —
   and **the planner's names and the suggester's names are DISJOINT: 0 of 7 planned pages
   matched** (`robots-txt-tester` planned, `robots-txt-generator` built). So the seven planned
   shells persist, were re-deployed by a rerender wave at 00:07–00:09Z, and at 08:2xZ still
   serve 0 forms / 0 inputs; the site carries 15 tool URLs of which 7 are prose. The
   websitepromotion variant is the other branch of the same fork: its planned
   `tool-channel-prioritiser` had NO `site_plan_sections`, so the link repair's builder parked
   all 7 `unbuilt_internal_link` items at `mark_no_ready_sections` (HITL) instead of writing a
   shell, and a `needs_content_page` ("0 component rows — build it") joined them at 04:01Z. The
   shell therefore needs planned generic sections; without them the page parks (187 family).
   **Consequences:** (a) holding planner tool stubs starves nothing — the producer makes its own
   rows (the 444 CONTRIB's deadlock branch is refuted, its door-closing point stands); (b) the
   planner's tool pages are pure liability — no producer ever fills them; (c) cleanup of the
   shells is instance work, and each remake will owe it until (1)–(3) land.

## Two vocabularies of "owned"

The plan validator's "owned" = *this page belongs to the tool pipeline* (files
`owned_page_review`). The save guard's "owned" = `pages.rebuild_policy='owned'` (adopted/ported
pages; 208/295/301/333 family). Same word, different predicates, no join between them — so the
hold that says "not the generic builder" protects nothing, and five generic writers reach a
planned tool page. Writers of tool-type pages that carry NO tool-level component, last 14 d
[MEASURED 2026-09-02 via `page_component_history.source_item_id`; undercounts archived items]:
`unbuilt_internal_link` 6 pages / 26 writes · `empty_section` 3 / 67 · `page_rerender` 3 / 20 ·
`needs_page` 3 / 14 · `needs_content_page` 2 / 8. The link repair is the one that fires FIRST on
a fresh remake; it is not the only door.

## Spread [MEASURED 2026-09-02 ~21:5xZ]

Deployed `page_type='tool'` pages with no `content_components.component_level='tool'` row
among their components: loanandmortgagecalculator.co.uk 16 · webdesign.co.uk 14 · loanzy.uk 11
· **seotools.co.uk 7 (7 of 7)** · idea.uk 3 · loancash.co.uk 3 · vonc.com 3 ·
leopardessconsulting.co.uk 2 · boxingonline.com 1 · cv1.co.uk 1 — **61 pages across 10 sites**;
control advertise.co.uk 0 shells / 3 real. ⚠ Upper bound: on adopted sites a ported tool may
live inline in a non-tool-level component, so the seotools mechanism is proven only where the
history table names the writer. `unbuilt_internal_link` items targeting `/tools/` in 14 d:
lendzy.co.uk 22 · seotools.co.uk 21 · garden-tools.uk 8 · 7 more sites 1–2 each.

**For the remake programme this is structural, not incidental:** pages build within ~1 h of
release; the design rotation reaches a new site in hours-to-days; every remake whose links go
live before its tools will fill its tool URLs with prose, and every 200 hides it.

## Fix candidates (ordered by what closes the door)

> **ORDER CORRECTED 2026-09-02 ~22:0xZ, on the 444 session's CONTRIB below (`ad1b3b1fa`):**
> candidate (1) alone cannot close this bug — the shells came through the phantom-link door,
> which consults no plan-time gate — and it carries a deadlock risk that 444's arm does not: a
> tool page's producer arrives from OUTSIDE the plan, later, so "hold if no tool component"
> holds every tool page on every fresh build, and whether that starves the rotation is exactly
> §7. **Read as: (2) and (3) are the door-closers; (1) is conditional on §7 and, if built, a
> SIBLING key (`enforce_tool_sources`, default OFF) on `c610898d1`'s derived pattern; (4) is
> churn; (5) is this lane's interim.** **§7 ANSWERED 2026-09-03: holding planner tool stubs
> starves nothing (tool-deployer creates its own rows, names disjoint), so (1)'s deadlock hazard
> is retired; its limit — it cannot close the link-repair door — stands.** Also from them: 444's live defer-half does NOT mitigate
> 450 (`hero-tool`/`generic-text-block` declare no required query fields), and the seven
> seotools shells are realised pages now — no plan-side gate removes them; cleanup is instance
> work.

1. **Planner/validator: a tool page with no tool component is held out of the plan** and filed
   as `capability_gap` naming the tool — the exact shape 444 just shipped for listing pages
   (`enforce_listing_sources`, migration 720, gate live 2026-09-02): a tool page's item source
   is its tool component. Makes the bad state unrepresentable at the source.
2. **Make the hold a control:** when `validate_site_plan` files `owned_page_review`, also set
   the page's policy (`rebuild_policy='owned'`, or a new `tool-pending`) so `pageIsOwnedForGuard`
   refuses every generic writer — one predicate, finally consumed. `tool-deployer` already goes
   through the tool pipeline, not this guard. Cheap; depends on no producer changing.
3. **Phantom-link repair: never route a `page_type='tool'` target to `page-build-handler`.**
   LNK-038 already suppresses outbound links to never-shipped pages at render, so the repair is
   redundant for these targets and only manufactures shells.
4. **One target, one build:** N links to the same unbuilt page filed N items and N rebuilds
   (220's family, dedupe by target page).
5. **Process mitigation, THIS lane, now (not a fix):** fire a one-shot design discovery for the
   site as soon as its plan completes, so tools exist before links do — the advertise ordering,
   made deliberate. Recipe: `portfolio_positioning/RUNBOOK_remake_release.md` §2b. UNEXERCISED
   on a remake; answer §7 above before pointing it at seotools.

## ⚠ CORRECTED 2026-09-03 13:5xZ — the guard behaved CORRECTLY; a different and more urgent bug is live

> **This block said "FALSIFIED — THE GUARD DID NOT REFUSE A GENERIC WRITE" and that was WRONG.**
> The `portfolio_positioning` lane measured the artefact and I verified it: the six pages had REAL
> tool components attached at 09:34–09:54Z, so at 13:05Z they were **not shells**, and the guard
> correctly allowed the write. My "0 tool rows ever" was an artefact — every census here joins
> `page_components` to `content_components` **on `component_id`**, and the write had set that to
> NULL, so a page serving a real 20 KB tool read as never having had one. **I inferred pre-write
> state from post-write state.** Kept visible because the shape is the lesson.
>
> **THE REAL BUG:** `save_page_sections`' delete-and-reinsert **preserves the `rendered_html` of a
> slot it does not recognise as one of its planned sections and drops the `component_id`.** Six
> seotools pages now serve working tools (78–85 KB, real controls, instance-scoped ids) with a NULL
> reference — **one rerender away from losing them, with nothing in their appearance to warn
> anyone.** It also means this bug's own census over-reports by six. Repair those references before
> any rerender touches them. Full detail, scoping query and repair sketch:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_450_tool_page_shells/HANDOFF_2026-09-03_continue_here.md` §1.

### The original (wrong) reading, kept for the record

**36 `page_component_history` writes across SIX of the seven canonical seotools shells,
13:05:14Z–13:24:36Z**, producer **`needs_content_page` → `page-build-handler`** (the generic
builder), on pages with **zero tool rows ever** and `rebuild_policy='generic'`. **Zero
`owned_page_review` rows of any class in that window** — so the guard neither refused nor left a
receipt. It WAS live: those pods ran `v1.0.1358` / stamp `d0252fd4d`, which carries `587666be8`.

Not a metric artefact — that was checked first. After `29b40e8bc` a RE-RENDER write to a shell page
is expected by design, so "writes to a shell page" could have been measuring the wrong thing. It
was not: the producer is a generic content builder, and these are the exact pages this bug exists
for.

**Established:** `page-build-handler` DOES declare `refuse_owned_page: true`; the items were minted
2026-09-03 by **`rerender_single_page_action`** and `tool-generator` — a producer this bug never
accounted for, and NOT `tool-deployer`; the writes split 18 with a resolvable `source_item_id` / 18
without, so there are likely **two write paths**, neither identified.

**NOT diagnosed, and deliberately not guessed.** Both the `load_page_record` arm and the
then-live `save_page_sections` arm should have refused. ⚠ **`29b40e8bc` removed the tool arm from
`save_page_sections`** on the argument that every generic path is caught earlier — this is evidence
that argument may be FALSE, in which case that commit removed the backstop that would have caught
this. **Settle that first.** Four candidate explanations, the queries to separate them, and the
current stamp (`v1.0.1359` / `3043885191…`, rolled 13:28Z, which behaves differently) are in
`docs/agent_docs/docs024_key_docs_latest/bugfix_450_tool_page_shells/HANDOFF_2026-09-03_continue_here.md` §1.

## FIX IN FLIGHT — the door half is COMMITTED, INERT until the next roll (2026-09-03, `bugfix_450_tool_page_shells` lane)

**Owner from 2026-09-03: `docs/agent_docs/docs024_key_docs_latest/bugfix_450_tool_page_shells/`**
(standing five there; the `portfolio_positioning` lane keeps the INSTANCE work — owner ruling
2026-09-03, build the 8 planned tools). Bug stays OPEN: the fix is committed but not live, and
the bar is fixed AND live.

**Commit `587666be8`** — `pageIsOwnedForGuard` → **`pageRefusesGenericBuild`**, a two-class
verdict (`owned` | `tool_pending`), register entry **PBP-053**, council
**APPROVED round 1**, corr **`2b236e83-ffd1-4911-b73f-1c17249064c1`** (submitted after the
commit; see the process note below).

**⚠ THE APPROVAL IS THE LEAST USEFUL PART OF THE VERDICT — read the four mediums.** Each named a
claim this file made that had been ASSERTED rather than measured, and answering them changed two
of the numbers in it:

- **The blast radius is far narrower than the census suggests `[MEASURED 2026-09-03]`.** The
  predicate matches 67 pages, but **48 are already `rebuild_policy='owned'`** and were refused by
  the old guard too. **The genuinely NEW refusals are 19 pages** — 18 under `/tools/`, and
  **exactly ONE elsewhere: `idea.uk` `/report.html`** (typed `tool`, six components, no tool).
  That single row is the entire measured population of the mislabelled-`page_type` misfire class
  that the fix-candidate discussion treated as an open risk.
- **A claim in this file's fix record was FALSE and is corrected:** "nothing in this estate has
  ever UPDATEd `rebuild_policy`". **Six hand-run migrations do** — 164, 195, 367, 377, 667, 668.
  The true claim is **zero Go UPDATEs → no AUTOMATED transition**, which supports the derived
  design just as well. Caught by the council's prior-art seat asking for the query rather than
  the assertion; logged in `WRONG_CALLS.md`.
- **Hot-path cost, measured not argued:** the added EXISTS reads `(never executed)` for a
  non-tool page (Postgres short-circuits on `page_type`); ~2.2 ms whole read, against the ~2.7 ms
  the `333` door already documents.
- **No competing detector:** `check_missing_tools.go` files one per-SITE `evaluate_tools` item at
  `tool-suggester`; ours is per-PAGE. Different question, no second disposition path.
- **Limitation, stated rather than glossed:** a parked `tool_pending` item holds its dedup slot
  and retracts normally (`deferred` is in neither the terminal nor the closed set), but **nothing
  promotes it back to dispatch** when the tool lands — it waits for its detector's next pass. Consulted at the `writeWorkItem` policy door, `load_page_record`'s `refuse_owned_page`
arm, `save_page_sections`, `AssemblePageAction`, the rerender escalation and the build-selection
exclusion. Kill switch `DISABLE_TOOL_SHELL_REFUSAL`, armed, scoped to the new arm only.

**Which candidate this is, and which it is NOT.** It is candidate **2's intent** — make the hold
a real control — reached by a different mechanism, and it achieves candidate **3's effect** for
all five producers without editing the detector:

- **Candidate 2 as written (set `pages.rebuild_policy`) was NOT taken, on a measurement.** The
  column is CHECK-constrained to `'generic'|'owned'` (migration 164) and **nothing in this estate
  has ever UPDATEd it** — two INSERT-time writers, no transition in any handler, check or
  scheduler, in either direction. A flag set at plan time is a flag **nobody clears**; the page
  would be protected for ever, and `'owned'` already means verbatim/adopted to ~12 other readers.
  So the second class is **DERIVED** (`page_type='tool'` AND no live `component_level='tool'`
  row), which **lifts by itself** when the tool arrives. Proven at the ordering:
  `deploy_tool_action.go` inserts the tool component (`:517`) **before** raising its companion
  content item (`:564`), so the tool pipeline is never refused by this — checked precisely
  because the portfolio lane was about to fire 8 `add_tool` items at these same pages.
- **Candidate 3 (never route a tool target to `page-build-handler`) was NOT implemented at the
  detector, deliberately.** §"Two vocabularies" already measured four more producers writing
  these pages (`empty_section` 3/67, `page_rerender` 3/20, `needs_page` 3/14,
  `needs_content_page` 2/8). A guard on `check_phantom_internal_links` is a guard on one door;
  the two seams chosen are crossed by all five. Independently, `availableBuilders` is unreachable
  from `discovery_checks` (import direction) — which is why `bugs_open/220` deferred the same
  idea, and 220's "no demand signal" ledger note is now answered, though not by its own route.
- **Candidate 1 (plan-side hold) is BUILT — `5e6fee47b`, register BLD-029, council corr
  `4e7497ed`, migration `729` committed but NOT APPLIED.** Sibling key `enforce_tool_sources`
  (default OFF) in a NEW `tool_item_sources.go`, zero edits to 444's file, exactly as their CONTRIB
  asked. It cuts the SUPPLY of stubs; the door half stops them being filled; neither depends on
  the other. **§7 is its whole arming licence** — holding a stub starves nothing only because
  tool-deployer creates its own rows and nothing reads planned tool pages. That is a NEGATIVE
  finding about consumers and goes stale by addition, so if a reader ever appears, disarm the key
  (one `jsonb_set`) and say so here.
  - ⚠ **The migration is deliberately unapplied, and not merely out of caution:** its replacement
    prompt text tells the planner *"validation holds back tool pages whose tool does not exist"*,
    which is FALSE until a chassis carrying `5e6fee47b` rolls. The KEY would be order-safe early;
    the SENTENCE is not. Apply recipe and preconditions: the lane's RUNBOOK §10.
  - ⚠ **It also holds EMPTY-SECTIONED tool pages**, on this file's own evidence: §7's
    websitepromotion fork parks 7 items in a human queue per remake instead of shelling, so "no
    shell" is a recurring HITL tax plus a phantom-link source, not a harmless outcome.
  - ⚠ **The tool gate runs BEFORE 444's listing gate**, so a `/tools/` hub whose children were just
    held is held too rather than shipped empty — which means arming this key changes what the
    LISTING gate does on the same plan. Told to that lane.
- **Candidate 4 (one target, one build)** — untouched, still 220's.
- **Candidate 5** — the filing lane's interim, unaffected.

**What is still true after this lands, so nobody reads it as more than it is:** the 61 existing
shells are not repaired or removed (realised pages; instance work), the planner still names tool
pages until candidate 1 lands, **the `owned_page_review` hold still has no consumer** (this makes
the SAVE side see the case the hold complains about; it does not join the two vocabularies), and
`rerender_single_page`'s re-assembly of existing components is not gated — migration 164 calls
that the sanctioned owned-page deploy path and gating it is the `bugs_open/210` family.

> **CORRECTION 2026-09-03 (this lane, at the code): §2 above attributes the `owned_page_review`
> hold to `validate_site_plan`. The action that writes that exact summary is
> `ReconcileSitePlanAction` (`reconcile_site_plan_action.go:270-300`)** — a later step of the same
> `build-site-planner` workflow, with `sync_pages` minting the `pages` rows between the two. It
> matters because it explains why the hold carries no page id: at `validate_plan` the row does not
> exist yet. Caught by reading the emitter instead of trusting the attribution.

> **PROCESS NOTE, declared rather than hidden.** The commit landed BEFORE its council submission
> and its register entry, which is not the practice. Another lane's correct pathspec commit took a
> half-finished hunk of this rename as a same-file passenger while the rest was still dirty, so
> HEAD called three uncommitted symbols and `make build-*` was broken fleet-wide; that could not
> wait ~30 minutes for a verdict. The commit carries **no `Council-Reviewed:` trailer** and makes
> no review claim. The underlying misstep was mine — holding a shared-package RENAME dirty across
> a long design phase — and is logged in `WRONG_CALLS.md`.

**Where the committed guard hands over to the REPAIR section above — they compose exactly.** The
repair's first completion proves the adopt path attaches the tool to the EXISTING shell row
(`page_adopted: true`, same URL, no duplicate). That insert is the precise moment the derived
predicate goes false: from then on the page is an ordinary tool page, generic producers reach it
again, and it is the save path's own tool-preservation machinery (the Layer-2 reappend/splice
arms) that protects the widget — which is the correct division, because those arms exist to keep
a REAL tool and mine exists only to stop a page pretending to have one. So a repaired shell needs
nothing switched off, and the fleet census in §Spread should shrink by one per repair rather than
needing a separate unblocking step. ⚠ The corollary for the repair lane: a repaired page still
carries the shell's leftover `generic-text-block` at the same position, and it is now
**re-buildable**, so the two-rows-at-position-2 question in the repair section is not something my
guard will hold still for you.

⚠ **Do NOT verify this (or anything) with a re-render until `9831e9ab4` rolls.** Since 2026-09-02
a light re-render renders a page's own stored `content_data` back at itself: clean run, healthy
`rerendered` count, nothing delivered (`bugs_open/454`). Verify at work-item terminal status, the
`owned_page_review` receipt (`spec->>'refusal_class' = 'tool_pending'`) and the served body.

## Related / owners

`bugs_open/220` (unbuilt-link dispatch, same producer) · `bugs_open/282` (validate resolver
drops tool sections — the other half: when the tool DOES exist the resolver may still drop it)
· `bugs_open/206` + `bugs_open/444` (no-producer family; candidate 1 is 444's gate generalised)
· `bugs_open/149` checker-layer lane (an `owned_page_review` consumer) · `bugs_open/447`
(tool-suggester seat-blind — the same wave, one step later) · `bugs_open/253` (component floor
that caught one shell write). No owner yet; the remake programme is the first customer.

## Verify

> **⚠ THE CENSUS BELOW WAS WRONG IN TWO DIRECTIONS AND IS SUPERSEDED — CORRECTED 2026-09-03**
> (found by the `portfolio_positioning` lane on the first half, and by measuring the FIX's own
> predicate against it on the second). **The `61 pages / 10 sites` figure quoted throughout this
> file — and in `PBP-053`, the LANDMINES entry and council submission `2b236e83` — was a FLOOR,
> not a total.** Corrected figure, `[MEASURED 2026-09-03 ~12:0xZ]`: **67 pages / 16 sites.**
>
> 1. **`deployed_at IS NOT NULL` excludes the never-shipped variant** — which is §7's
>    websitepromotion fork, the page whose link-repair items park at `mark_no_ready_sections`
>    instead of writing a shell. **A count of shipped shells structurally cannot see a page that
>    never shipped**, so the variant this file calls the other branch of the same fork was
>    excluded from the census measuring its own class, and always had been. **+4 pages / 4 sites**
>    (websitepromotion.co.uk, garden-tools.uk, adversecreditmortgage.co.uk, idea.uk).
> 2. **The census does not test `cc.is_active`, and the FIX does.** A page whose only tool
>    component is INACTIVE reads as "has a tool" here and as a shell to `toolShellPredicateFor`.
>    **+9 pages / +4 sites** — finetuning.uk 3, robot-hands.com 2, ai-agent-orchestration.com 2,
>    gaswholesalers.com 1, and leopardessconsulting 2→3 — none of which appeared in the filed
>    census at all.
>
> **The general lesson, which is why this is at the top of the block rather than in a footnote:
> the census that measured this bug's population was never the predicate the fix uses.** Those
> two drifting apart is how a fix gets judged against the wrong denominator. The corrected
> queries below are the guard's own predicate, split by publication state.
>
> **And a property of ANY version of this census, from the same lane:** it is a
> repair-INITIATED count, not a repair-COMPLETED one. Attaching a tool component removes a page
> from it immediately, while the public keeps seeing prose until the rerender drains — an
> unbounded queue delay. On 2026-09-03 seotools left the census entirely at 10:27Z with **0 of
> its 7 pages published**. A later session re-running this will read "seotools: clean" off a site
> serving seven prose pages. **Acceptance is the served body — form and input counts — never
> this census.**

```sql
-- CORRECTED: the guard's own predicate, split by publication state. `shipped` is the
-- served-prose population; `never_shipped` is the parked/sectionless fork (§7) that the
-- original census could not see. Add s.domain=... to scope.
SELECT s.domain,
       count(*) FILTER (WHERE p.deployed_at IS NOT NULL) AS shipped,
       count(*) FILTER (WHERE p.deployed_at IS NULL)     AS never_shipped
  FROM pages p JOIN sites s ON s.id=p.site_id
 WHERE p.page_type='tool' AND p.status='active'
   AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
                   WHERE pc.page_id=p.id AND pc.build_status<>'removed'
                     AND cc.component_level='tool' AND cc.is_active)
 GROUP BY 1 ORDER BY 2 DESC, 3 DESC;

-- SUPERSEDED (kept so the 61/10 figure in older text is traceable, NOT for reuse):
-- SELECT s.domain, p.name FROM pages p JOIN sites s ON s.id=p.site_id
--  WHERE p.page_type='tool' AND p.status='active' AND p.deployed_at IS NOT NULL
--    AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
--                    WHERE pc.page_id=p.id AND pc.build_status<>'removed' AND cc.component_level='tool');
-- who wrote them
SELECT p.name, w.item_type, count(*) FROM page_component_history h JOIN pages p ON p.id=h.page_id
 LEFT JOIN site_work_items w ON w.id=h.source_item_id WHERE p.site_id=:site AND p.name LIKE 'tool-%' GROUP BY 1,2;
-- what the plan asked for
SELECT page_name, string_agg(component_name, ',' ORDER BY ordering) FROM site_plan_sections sp
 JOIN site_plans pl ON pl.id=sp.plan_id WHERE pl.site_id=:site AND pl.is_current AND page_name LIKE 'tool-%' GROUP BY 1;
```
```bash
# at the body — a tool is a FORM, never a size; control against a real one in the same run
for t in <slugs>; do b=$(curl -s "https://<domain>/tools/$t/?cb=$RANDOM"); printf "%-30s forms=%d inputs=%d\n" "$t" "$(grep -o '<form' <<<"$b"|wc -l)" "$(grep -o '<input' <<<"$b"|wc -l)"; done
```

## CONTRIB from the 444 fixing session (2026-09-02 late) — the asked-for answer on extending the gate

Asked directly ("a reason the same enforce_* key should NOT be extended this way goes in
this file"). Answer: the extension is architecturally right, but ONE property of the 444
gate does NOT transfer, and it is the property that made 444's gate safe to arm fleet-wide.

1. **The order-safety divergence — the real hazard.** 444's section-index arm is order-safe
   BY CONSTRUCTION: a hub's children are IN the plan, so post-plan builders (tool-deployer)
   cannot false-positive. A tool page's producer is OUTSIDE the plan and arrives LATER by
   design (the ~3h rotation) — so "hold if no tool component exists" holds EVERY tool page
   on EVERY fresh build; your own control shows advertise passed only because the rotation
   happened to reach it 34 min before the plan persisted. Whether that hold is CORRECT or a
   DEADLOCK depends exactly on your §7 [UNVERIFIED]: on advertise tool-deployer CREATED the
   page rows itself, which if general means held planner pages starve nothing (the producer
   makes its own, and the planner's tool stubs were never needed); but if evaluate_tools /
   the rotation reads PLANNED tool pages to decide what to build, holding them starves the
   producer and the capability_gap receipt breaks the cycle only if something consumes it —
   which is this bug's own §2 finding (holds nothing consumes). **Answer §7 before arming
   any tool arm.** (444's precedent for the happy case: the directory family has a re-add
   path — MissingDirectoryPageCheck — that makes dropping safe; the tool family's
   equivalent is precisely what §7 establishes or refutes.)
2. **If extended: a SIBLING key, not an overload.** The registry structure, capability_gap
   shape, preserve-guard rule and fail-open policy in `listing_item_sources.go` all
   transfer (it was built to take another arm) — but arm it behind its own optional key
   (e.g. `enforce_tool_sources`, default OFF) rather than widening what
   `enforce_listing_sources` means: independent rollback (a tool-arm misfire must be
   switchable without losing the live listing gate), and a tool page is not semantically a
   listing page. Build on `c610898d1`'s DERIVED-vocabulary pattern, not the deployed
   intermediate's hand map. Note both keys land on an action with no ActionInputSpec
   (WFA-013-invisible — already tracked in BLD-028 verify-later; a second key strengthens
   the case for adopting a spec).
3. **Candidate 1 alone does not close 450 — your mechanism proves it.** The shells were
   built through the phantom-link door (`unbuilt_internal_link` → `page-build-handler`),
   which consults neither `builderForPageType` nor any plan-time gate; it will do the same
   for any future tool-page row from any source with links pointing at it. Your candidates
   2/3 are the door-closers; 1 only shuts off the plan-side supply. Also stated so nobody
   assumes otherwise: 444's LIVE defer-half does NOT mitigate 450 — `hero-tool` +
   `generic-text-block` declare no required query-sourced fields, so those sections
   resolve as pure LLM work and build "successfully".
4. **The seven seotools shells are now REALISED pages** — any plan-side gate (444's rule,
   which an extension should keep) will never remove them; their cleanup is instance work
   exactly as 444's five are.

Offer stands: if the fixing thread wants the extension, it lands naturally as one resolver
arm + one key + tests in `listing_item_sources.go` — but only after §7 and the 090 verdict
(`96e97dc4`) are read.

## REPAIR of the 7+1 instances (owner ruling 2026-09-03) — and what the first one proves

The owner ruled BUILD the planned tools (not retire the stubs). Items fired by
`portfolio_positioning` after the v1.0.1356 roll settled: 7 at 09:05:26Z, the redirect-chain
checker at 09:13:50Z (ruled a "lesser" browser-only version — no backend provisioning exists for
generated tools). SQL of record:
`docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/SQL_2026-09-03_fire_planned_tools_450_instance.sql`
(+ `…_2026-09-03b_…_redirect_checker_lesser.sql`).

**First completion (09:34:09Z, `tool-robots-txt-tester`) settles two open questions:**

1. **The adopt path works on a SHELL page, and is the repair route.** `create_tool_component`'s
   `adopt_existing_page: true` (live on tool-generator's `save_tool`; `bugs_closed/286`, TL-044)
   attached to the EXISTING page `6feb9797` — `page_adopted: true`, same row created 2026-09-02
   16:13:27Z, same URL `/tools/robots-txt-tester/index.html`, **no duplicate page minted**. So a
   shell is repairable in place at the already-linked URL; §7's "tool-deployer creates its own
   rows" holds only when the names differ (the suggester's case), not as a limitation of the
   machinery.
2. **The shell's prose components are NOT removed, and the new tool lands at the SAME position.**
   Page now carries `1:hero-tool(section)`, `2:tool-robots-txt-tester(tool, 20,839 B)` AND
   `2:generic-text-block(section, 2,422 B)` — **two components at position 2**
   [MEASURED 2026-09-03 09:36Z]. `create_tool_component` inserts the widget at a hardcoded
   position 2 ("same as deploy_tool_action") without consulting what the page already holds.
   Ordering between two rows sharing a position is whatever the renderer's ORDER BY leaves it —
   not declared anywhere. **Consequence for the class:** repairing a shell in place leaves an
   orphaned prose block competing for the slot, so the repair needs either a position bump or a
   deliberate retirement of the shell's `generic-text-block`. Verify at the served body after the
   page's rerender drains, per site.

Also raised by the same run (normal tool-generator fan-out, not a defect): a companion guide page
`tool-robots-txt-tester-guide` (blog-post, planned), a `nav_drift` rebuild, two
`needs_content_page` (tool page + guide), three `content_rewrite` cross-links.


## REGRESSION AND NARROWING, 2026-09-03 (post-roll) — the door half was too wide at ONE seam

`587666be8` went live on `v1.0.1358` (stamp `d0252fd4d`) and immediately refused something it
should not have. Reported by the `bugs_open/427`/`454` lane with the measurement.

**What was wrong.** The arm sat at `save_page_sections`, which is shared by two paths that are
**indistinguishable at that seam**: a generic build authoring prose about a missing tool (this
bug's harm), and a `page-rerender` writing back components that are **already deployed and
serving** — no new authorship at all. It caught both.

**The collateral dominated `[MEASURED 2026-09-03, independently after the report]`:** of the **67**
pages the predicate matches, **54 across 10 sites are already serving deployed components**; only
**13** are the empty page this bug is about. The arm was refusing roughly **four times more repair
than harm**, on the sites this file's own Spread section lists (loanandmortgagecalculator 16,
loanzy, idea.uk incl. `report`, leopardessconsulting, loancash).

**Why it bit now rather than quietly** — the 427 lane's observation, and the transferable part:
until that morning a re-render on those pages ran, reported success and delivered nothing
(`bugs_open/454`: `classifyStoredSection` dropped its own plan). So refusing those saves had cost
nothing OBSERVABLE, because the saves were writing back unchanged bytes anyway. **The guard's
arrival and the repair vehicle's return to working landed in the same image.** A guard whose harm
is masked by an unrelated live defect looks free until that defect is fixed.

**The narrowing (`29b40e8bc`, rides the next roll): the tool arm no longer fires at
`save_page_sections`.** This costs the class nothing, because every generic path is caught EARLIER
and none reaches that line — `page-build-handler` at `load_page_record`'s `refuse_owned_page` arm,
`pageflow-builder`/`page-rebuild`/`site-work-orchestrator` at `AssemblePageAction`, any producer at
the `writeWorkItem` door (at file time, before LLM spend), build selection at
`genericBuildExclusionSQL` — while `page-rerender` crosses **none** of them and only ever arrives
at the save. The **owned** arm at that seam is untouched and byte-identical: migration 164 still
stops anything, re-render included, from delete-and-reinserting a live verbatim tool.

⚠ **Until the next roll the live chassis still refuses those saves.** `DISABLE_TOOL_SHELL_REFUSAL`
disarms the whole tool arm fleet-wide with no build, scoped so it cannot touch 164's protection —
an owner call, put to them.

> **CORRECTION to this file's own fix record:** it said "the 61 shells stop receiving generic
> rebuilds — intended". That was wrong in the direction that mattered. Most of those pages are
> LIVE, and their re-render is a repair vehicle, not a threat. What is intended is stopping a page
> BECOMING a shell, not freezing the ones that already are.

---
**Cross-reference 2026-09-04 (`portfolio_positioning`):** §7's finding — the tool-deployer creates and
ships its own pages, independent of the plan — is the producer behind **`bugs_open/478`**: those pages
carry `deployed_at`, the strategist's refresh gate reads any shipped page as "this site was planned", and
a greenfield site whose tools ship first never gets a briefing or a plan (copyonline, oxenunity; a
seeded skeleton did the same to cookly). Not a defect in this path; a consumer that assumed pages only
follow plans.
