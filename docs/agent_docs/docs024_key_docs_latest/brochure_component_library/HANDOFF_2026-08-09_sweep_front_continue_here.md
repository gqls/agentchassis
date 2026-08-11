# HANDOFF 2026-08-09 — fundamentallyai sweep front: cold-start for a fresh chat

**Supersedes `HANDOFF_2026-08-05b_improvement_sweep.md`** for this front. Does **NOT**
supersede `HANDOFF_2026-08-05_continue_here.md` (camera / contact-sheet / 151-checker,
another thread) — that is a parallel front and still current.

Read in this order: this file → `NOTES_brochure_component_library.md` tail (08-05, 08-08
entries) → `SUMMARY_2026-08-08_the_machine_noticed_and_nobody_was_listening.md` →
`RUNBOOK_brochure_component_library.md` §improvement sweep and §linking a newly built page.

Site id, needed by nearly every query: **`199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`**.

---

## 1. THE ONE THING OWED — read the council verdict

**`Council-Submitted: 9da24d85-d440-49de-9d0c-f861de83cac4`**, commit **`1c2e25c8f`**.
The code is already on the shared branch (forward-only; holding it was never available).

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '9da24d85-d440-49de-9d0c-f861de83cac4';
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 5;
```
Budget ~30 min from 2026-08-09; a missing orchestration row is latency, not a dropped
dispatch — **do not resubmit on that evidence.** On REVISE, revise and resubmit with
`RESUBMIT_CORR=9da24d85-…`. On APPROVED, nothing to do: `098` credits the commit
automatically at report time. **Never hand-write `Council-Reviewed:` on a verdict you have
not read.**

The three questions I explicitly asked reviewers to check are in the submission's `risks`
block; the second is the one I am least sure of (a fourth eligibility constant vs widening
one of the three existing ones — `bugs_open/185` already tracks whether the family merges).

## 2. What shipped, and what it is inert until

### 2a. Linkability fix — committed `1c2e25c8f`, INERT UNTIL THE CHASSIS ROLLS

`resolve_internal_links_action.go`'s three loaders (`loadContentHubs`,
`loadInteractivePages`, `loadResolverPageSet`) carried only the **lifecycle** arm
(`status`), never the **build** arm. A `pages` row is `active` from the moment the planner
creates it, so a never-built page was a valid CTA target **and** a member of `validPages` —
which is what `applyCTARecompute` consults to decide whether to KEEP an authored link. That
is why the cost-calculator guide served a link to `/platform-log/index.html` for 18 days
while it returned 404, and survived every rerender.

**The measurement is the load-bearing part, and it refuted my first attempt.** Applying the
existing `PageHasShippedPredicateFor` would drop 39 pages from link candidacy fleet-wide and
**11 of them serve HTTP 200** — nine mortgagecalculator.co.uk, idea.uk's ab-test-calculator,
webdesign.co.uk's llm-cost-calculator; nearly all `needs_rebuild`, i.e. built once, still
serving their last artefact, never stamped with `deployed_at`. So the change instead adds a
narrower named floor, `datahelpers.PageMayBeLinkedPredicateFor` = never BUILT (`planned` +
no `deployed_at`): 22 pages fleet-wide, **all 22 return 404, none returns 200**.

**Do not "tighten" this back to the shipped-predicate.** Both measurements are in the helper
comment for exactly that reason. Re-run them before changing it:
```sql
-- would-be-excluded set; HTTP-test every row before trusting either predicate
SELECT s.domain, p.url, p.build_status FROM pages p JOIN sites s ON s.id=p.site_id
 WHERE p.status NOT IN ('deleted','archived')
   AND p.deployed_at IS NULL AND COALESCE(p.build_status,'') = 'planned' ORDER BY 1,2;
```

**Verify after the next roll** (a roll is not evidence — grep a string the change ADDED and
one it did not):
```bash
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
 'grep -ac "PageMayBeLinkedPredicateFor\|deployed_at IS NULL AND COALESCE" /app/agent-chassis'
```
Then behaviourally: a `cta_links_stale` rerender on a site with a `planned` page must not
select it. There is no such page on fundamentallyai any more (§2b), so pick another site.

### 2b. The three duplicate 404 rows — ARCHIVED, done

Created by the 08-07 planning pass, `active` + `planned` + never deployed, each duplicating a
page that already served:

| archived | duplicate of |
|---|---|
| `/tools/llm-cost-calculator/index.html` | `/tools/llm-cost-calculator.html` (200) |
| `/tools/tools/index.html` | `/tools.html` (200) |
| `/blog/ai-readiness-checker-guide.html` | `/guides/tool-ai-readiness-checker-guide.html` (200) |

**Archived, not row-deleted, and that was deliberate.** `status='archived'` IS the platform's
delete for a page: it removes the row from every derivation and from re-rendering, and
`retract_page_deployment_action.go`'s own header states there is **no writer of
`status='archived'` anywhere in Go or the frontends** — archiving is a hand-run SQL operation
by design. A hard `DELETE` was rejected because three FKs onto `pages` are NO ACTION
(`link_registry.target_page_id`, `redirects.source_page_id`, `page_component_history.page_id`)
so it can fail rather than cascade, and `site_work_items.page_id` is not FK-constrained at
all, so four open work items would dangle at a vanished id. No file retraction was needed —
all three had `deployed_at IS NULL` and 0 components, so nothing ever shipped.
Their four `needs_human_review` work items were cancelled in the same transaction with the
reason recorded. Live counterparts re-checked afterwards: all three still 200.

### 2c. Earlier the same week (already committed)

Capabilities chart re-rendered from the register (`9136/7856/208/214/16`, verified live);
Platform Log linked across 25 of 28 pages via `nav-link-fixer` + assemble-mode propagation;
tools page "2 companion guides" → 3. Commits `bc5a91257`, `103fbce7e`.

## 3. OPEN — `bugs_open/210`, the logo failure. Diagnosed, deliberately NOT fixed

`needs_logo` items are **unhandleable fleet-wide**. `image-build-handler`'s `call_logo_gen`
maps `prompt` from `input_data.spec.image_prompts.logo`; `check_placeholder_image_in_use.go`
writes that key **only when it can find a planned prompt** and files the item anyway when it
cannot. `input_mapping` is a strict allow-list, so the step dies at input extraction.
`call_hero_gen` has the identical shape.

**The one-line fix is a TRAP — do not ship it.** `"prompt"` → `"prompt?"` makes the field
optional, but `image-generator` has **no `prompt_template`** in `default_config`, so the
chain falls through to `generate_image_actions.go:895` and returns the generic
`"Generate content based on the provided context."` — and `store_logo_asset` would save that
image **as the site's logo**. A loud failure becomes a silently wrong brand asset.

The producer's own comment claims "the handler will fall back to its default prompt
template". It never reaches any fallback. That comment is part of the bug.

Three costed options are in the bug file; my recommendation is **3 now, 2 later** (stop
filing an item the handler cannot consume; build the real prompt-synthesis capability
separately). **This is a design choice and I did not make it unilaterally.** It also has not
been through `090` — the bug file declares that substitution and why.

Second defect in the same row: the finding was probably a **false positive** — nothing local
references `/assets/images/logo.png`; the site's logo is `logo.jpg` (200), and the only
`logo.png` in served HTML is leopardessconsulting's, an external URL that works. That is the
`bugs_closed/128` basename-vs-path family resurfacing.

## 4. Still open, not started

- **`image_url_404:logo.png` is `blocked`** — *"No handler_agent set — item cannot be routed
  to any agent."* A real finding that can never drain. Same disease as the twelve-day silence
  that started this whole front. Decide: give it a handler, or close it.
- **`needs_content_page` FAILED** on the spawn→call handshake, carrying an unexamined claim:
  *"the site claims 'more than ten live production sites'"*. **Do not cancel pre-diagnosis.**
  Note the register's F1 says 18 and its `writer_line` instructs a FLOOR, never the exact
  count — so "more than ten" is correct as written and this may be a non-problem.
- **`audit_tool`** on the selector: "Claim timed out (attempts exhausted)".
- **`capability_gap:hardcoded_section_colors`** DEFERRED, "outside scope".
- **`deactivated_head`** — TRUE here but **17 of 19 sites** pin `head` to a deactivated
  component and all serve a correct head (the chrome join honours the pin with no `is_active`
  filter, REB-006). The deactivation is the anomaly. **Do not repoint this one site** — that
  hides a fleet condition.
- **A second improvement sweep** once the above settle: `./run_improvement_sweep_once.sh
  fundamentallyai.com`, blast-radius header first, pre-flight the `detected` queue every time.

## 5. Traps this front paid for — read before trusting your own checks

1. **HEAD is currently RED, and not from this work.** `TestValidDocSubjectTypes_LockstepWithMigrationCheck`
   and `TestEveryCheckProducedItemTypeIsClassified` fail at clean HEAD (proved via
   `git archive HEAD` into a temp dir and re-running). Another lane's `decision` subject-type
   work. Don't chase them, and don't let them mask your own.
2. **A finding's `matched`/`snippet` is a photograph taken at filing time** — and on this
   estate the artefact often moves *because of the run that filed it*. I told the owner the
   capabilities page showed 97 approved rounds vs a real 205; the sweep's own later re-renders
   had already moved it to 187 hours earlier. Re-read the artefact before quoting a work item.
3. **Never count bare integers in HTML.** `grep -o 202` matched inside every `2026` date.
   Extract the field (`grep -oE 'evidence-chart__value[^>]*>[^<]*'`), don't search for the value.
4. **Gate a `grep -c` on byte count.** An empty `curl` body and a page genuinely missing an
   element both give 0. This nearly produced a false "contact.html has no chrome" report.
5. **A reconciler's own failure list is a claim.** `reconcile_footer_nav.sh` reported 5
   missing; 2 were false negatives (polled before propagation settled) and 3 were pages that
   don't exist. **Identical response sizes across unrelated URLs (three at exactly 2696 bytes)
   mean you are measuring the error page.**
6. **Never silence stderr in a poll loop.** `orchestration_states` has **no `agent_type`
   column**; a 20-minute watch on that column printed nothing and read as a lost dispatch
   while the sweep had already completed. Now a LANDMINE entry.
7. **Measure the blast radius before submitting, not after.** The linkability fix's first
   form was refuted by its own measurement (11 live pages would have been delisted). Had I
   asked the council to check it, that would have been asking a reviewer to do arithmetic
   I could do myself.

All of these are in `WRONG_CALLS.md` / `LANDMINES.md` as checks rather than stories.

## 6. Commit trail (this front)

`074bcda5b` sweep + docs · `ce6496aa0` landmine · `bc5a91257` capabilities + Platform Log
link · `103fbce7e` summary · `1c2e25c8f` linkability fix + `bugs_open/210`
(**Council-Submitted 9da24d85**) · this file's commit.

---

## ADDENDUM 2026-08-11 (from the fact-assignment front) — the census replan ran; your three archived pages are BACK in the plan

Owner-authorised census replan of fundamentallyai ran 10:19–10:22Z today (corr
`e74974b3`, new plan `40a66d3a-b80e-4f92-9033-c6de1f43bcd1`). As predicted, the
planner re-planned the three pages you archived in §2b: `ai-readiness-checker-guide`
(reconcile queued an AUTO-BUILD, `needs_page`, claimed 10:29Z — it will deploy),
`tool-llm-cost-calculator` and `tool-tools` (both parked at `needs_human_review`
as `owned_page_review`, no build). **The hand-archive pass is yours per the owner's
08-11 ruling 1** (phantom cleanup accepted as a known cost of running the census
now). Note ai-readiness-checker-guide will need file retraction this time if it
deploys — unlike your §2b rows it will have shipped an artefact. Also for your
§2b table: the replan's write recorded 2 `PLAN_PAGE_MERGE_LOSSY` rows — the two
duplicate guide PAIRS (`automation-savings-estimator-guide` /
`model-approach-selector-guide` vs their `tool-` twins, all four rows still
active+deployed) are now on the owner's radar via the 215 revisit trigger.
