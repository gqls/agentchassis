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

## Related / owners

`bugs_open/220` (unbuilt-link dispatch, same producer) · `bugs_open/282` (validate resolver
drops tool sections — the other half: when the tool DOES exist the resolver may still drop it)
· `bugs_open/206` + `bugs_open/444` (no-producer family; candidate 1 is 444's gate generalised)
· `bugs_open/149` checker-layer lane (an `owned_page_review` consumer) · `bugs_open/447`
(tool-suggester seat-blind — the same wave, one step later) · `bugs_open/253` (component floor
that caught one shell write). No owner yet; the remake programme is the first customer.

## Verify

```sql
-- shells: tool-type pages with no tool-level component (add s.domain=... to scope)
SELECT s.domain, p.name FROM pages p JOIN sites s ON s.id=p.site_id
 WHERE p.page_type='tool' AND p.status='active' AND p.deployed_at IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
                   WHERE pc.page_id=p.id AND pc.build_status<>'removed' AND cc.component_level='tool');
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

