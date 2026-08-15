# NOTES — bugfix 277 (append-only, newest at the bottom)

## 2026-08-15 — session "bugfix 033": research, design, build

**Ownership checks before starting.** `who-owns.py 033` → review_queue_drain (dormant since
07-28; the one commit in 14d was the 168 lane contributing in). `who-owns.py 277` → only the
filing commit, no owning workstream. Live `.jsonl` transcripts: the staged_component_build
session (filed 277 this morning, says "not this lane's build"), bugfix 213, and copy/fact
lanes — nobody on 277. Taken.

**Bug validity re-measured.** Queue at 735 `needs_human_review` (was 325 on 07-28) at ~10:30Z;
NOTE another lane's CTA re-run resolved 183 `cta_names_unknown_destination` items while this
session ran (commit `39aadb590`) — re-measure before quoting any queue total.
`required_fields_missing`: 44 open (50 lifetime closed, 153 total `auto:revalidated` closes
across types by the 033 drain).

**Population measured (the load-bearing finding).** Of the 44 open items:
- 36 point at components with **NULL content_data and substantial serving rendered_html**
  (953–21,797 bytes; content_item_id/data_path/schema_mode all NULL — pure blobs). The
  producer's reason string ("the template renders them as empty strings") is FALSE for these
  today: they do not render through the template. The real risk is latent (a regeneration
  replaces served HTML — bugs 263/238 class), and the render gate bounds the empty-strings
  scenario (RenderComponentAction refuses; rerender_page_sections escalates to writer).
- 7 ghosts by `spec.component_id`; but 38/44 resolve by (page_name, slot) — the component-id
  join under-resolves, exactly as 016b §9 records.
- 1 genuinely-partial row (`4fa5b019`, ai-agent-orchestration index case-studies-grid,
  5 × `cardN_image_url` — STRING-typed so the producer's image-type skip missed them; the
  writer minting image paths is a canary watch-item).
- 4 items on pages with `sections='[]'`; only the gas converter's page is owned/tool-typed.
- 35 of 44 name `headline` — concentration explained by hero-slot blobs, not a writer bug.

**Precedent found and reused.** IMG-071 (seed 397, owner-instructed 08-12): router-not-repairer
handlers. Live check: items HAVE completed through them (5), so the router shape has executed
in production — the planning agent believed otherwise from the seed's 0-assigned assert; the
live rows are ground truth (assignment happened later, post-248).

**Design correction (my own first design was wrong — worth keeping).** I proposed
checkpoint_for_review escalation + complete for the blob/owned classes. The planning agent's
review refuted it: checkpoint writes NO item_key (verified `checkpoint_for_review_action.go:198-207`
— column absent from the INSERT) and completing the original releases the dedup key → producer
re-raises → two-strike births endless `unresolved` rows (5 keys of this type already at
1 strike, 2026-08-04, `work_items_common.go:123-125`). Park-in-place replaced it: the original
row goes back to `needs_human_review` with the triage in `error`, holding the key. Also
corrected by the same review: route order (owned before blob before no_plan_generic — a
sectionless page with a blob component must park, not convert to recreate) and a missing
`resolved` route (the revalidator's own closure predicate).

**Mechanics verified in code before building** (file:line in the planning transcript, spot
re-verified here):
- `complete_work_item` on a self-parked item: guarded UPDATE excludes needs_human_review,
  0 rows → SUCCESS payload `{completed:false, reason:"already_flagged_or_terminal"}` — never
  mark_failed (`load_work_item_actions.go:956-978`).
- No verifier registered for `required_fields_missing` → completion gate 2 passes; a verifier
  added later would fail-closed the converted arm (register landmine, CQ-023).
- `update_work_item_status` supports `error_message` + `result_fields` literals
  (`v3_site_actions.go:5280-5416`).
- `create_work_item` StrictConfig: data inputs (site_id/page_id/parent_item_id/summary) go in
  config as dotted paths; `spec` RETIRED → spec_paths/spec_literal; spec_paths unresolved =
  HARD ERROR (all 44 rows have the uniform spec dialect, so all paths resolve).
- conditional_branch: `==` cascades only (a missing field makes `!=` true); final else_step =
  mark_failed so unknown routes fail loudly.
- query_database: params auto-prefix `input_data.`, nil = hard error, `output_format: object`
  flattens the FIRST row — classifier built to always return exactly one row.
- Pure-SQL workflow: AI-endpoint claim gate is a non-issue (`extractAIEndpointFromHandler`
  returns "" → check skipped).
- Bug 230 (discovery recurring driver) FIXED 08-09 → a wrong close's re-raise path is
  measured (~9 site-examinations/day), not aspirational — this answers the council objection
  recorded in 016b about the 033 drain's safety case.
- Bug 238's two fix commits (`d26c26a9a`, `51f56d0c9`) verified ancestors of the running
  chassis stamp `a2a691213dfbe11d38549f128870ef41cbf24a83` (extracted from /proc/1/exe via
  the `buildinfo.GitCommit=` marker; NOT `strings`).

**Seed SQL proven before apply.** The exact embedded one-line query was extracted from the
seed file (not retyped), `''`→`'` unescaped, `$1/$2` substituted per row, and run against the
five canary candidates: `332bb3f6`→stale, `4fa5b019`→partial, `e512af8a`→no_content_data,
`483fb749`→no_plan_owned, `7ed472ab`→no_plan_generic — all five match the census.

**Go change + tests.** Producer emits `HandlerAgent: requiredFieldsHandlerAgent` +
`Status: "triaged"`; constant declared in `const ( … )` block form because
handler_coverage_test's const resolver only matches block-form declarations (first run failed
on a bare `const x = "…"` — the sensor read it as a runtime route). Full
discovery_checks suite green; revalidator contract tests green.

**Council.** Submitted `7b0e2833-715f-4a9a-897b-efd913073582` before committing; verdict
pending at time of writing (budget ~30 min dispatch latency — do not re-trigger on a missing
orchestration row).

**Misstep log (this session):** (1) my first live-session ownership grep piped `grep -l` into
`wc -l` — every file trivially "matched"; caught within a minute, redone with `grep -c` per
file. (2) The checkpoint-escalation design above — caught by the planning review before any
build. (3) A binary probe greped for the FIX commit's sha inside the binary — a binary
carries only its own build stamp; corrected to extract `buildinfo.GitCommit=` and use
`git merge-base --is-ancestor`.

## 2026-08-15 (same session, later) — seed applied, canary verified per-arm, council round 1 REVISE, round 2 resubmitted

**Seed 410 applied** ~11:02Z (verify block passed, 0 assigned). **Canary (4 rows) assigned**
and dispatched within two cadences:
- `332bb3f6` (stale) → `complete`, orch `0177ce18` route=stale. **The row's `result` was
  OVERWRITTEN by the loop's mark_complete with spawn bookkeeping** — the close arm's evidence
  survives only in `orchestration_states.collected_data.triage`. RUNBOOK corrected (its
  original completed-row verification query was wrong).
- `e512af8a` (blob) → parked at `needs_human_review`, route + message on the row. ✓
- `483fb749` (gas converter) → parked, route=no_plan_owned (orch `8dd51e7e`, n=9 fields,
  html_len 13248). ✓
- `4fa5b019` (partial) → `complete`/converted; minted
  `content_rewrite:from_rfm:_ai-agent-orchestration.com_dda5fbdf…` @ page-build-handler,
  `mode=edit_live`, priority 30, source AND created_by = the router, `depends_on` = the
  original (create_work_item's `parent_item_id` input feeds the depends_on DISPATCH GATE, not
  a provenance column — the gate held seconds until close_converted completed the original).
  Rebuild still executing at last check.

**Council round 1: REVISE** (14 seats, 3 abstained; gating: editquality — the assignment
UPDATE must be a reviewed edit, not prose). Real catches among the rest:
- **prior_art_librarian caught a stale figure**: "exactly one ever terminal" was a 07-25
  code-comment measurement; live lookup = 50 complete (the 033 drain's own closes).
  → corrected in PLAN + WRONG_CALLS.md entry.
- **guardian's component-id key objection** → answered with analysis + a measured instance
  (the partial row's stored spec.component_id was already dead; classification by (page,slot)
  found the live one) — written into CQ-023.
- **debug_historian's needle-gate demand** → `ASSIGN_2026-08-15_fleet_assignment.sql` built
  with pre-image step, execution-time refusal conditions, ROW_COUNT assert, revert.
- **reuse_agent/architecture proliferation warning** → CQ-023 tripwire: the FOURTH router
  author should build the generalised engine.
- render_guardian's objection cites a pre-178-fix landmine; PBP-028/migration 299 is the
  edit_live channel — answered by citation.
- guardian's scheduled-deps question: 0 scheduled_tasks pre_queries name the type or status.

**Round 2 resubmitted** on the same trail (`RESUBMIT_CORR=7b0e2833…`, run orch `a687f2be`).
**Fleet assignment (remaining ~40 rows) HELD** for the verdict + the partial canary's
artefact check.

## 2026-08-15 (same session, round 2 worked) — two REAL refinements found by the round, both measured; seed at v3; producer already LIVE

**Round 2: REVISE again** (gating: bug_historian — prove the edit_live writer path is guarded
against the missingkey=zero blanking family). Working the objections found two genuine
defects in my v1 arms, both settled by MEASUREMENT rather than argument:

1. **The "partial" canary was a category error, and the round's gating question exposed it.**
   The five missing fields (`cardN_image_url`) are `source: "site_assets.image"` in the
   component schema — NOT llm fields. The prose writer can only MINT urls for those, and the
   live run proved the outcome: validate_content refused ("0 blockers, 1 errors"), nothing
   shipped, artefact byte-identical. → **v2 adds the `asset_sourced` route**: any still-empty
   field with `source site_assets.*` parks with the asset-pipeline routing fact. The
   mis-minted conversion item was cancelled with a resolution note. (Also answers the gate:
   where fields ARE llm-sourced, the conversion path renders through `render_component` —
   the guarded call site, `missingRequiredLLMFields` v3_site_actions.go:1843/2066 + the
   PBP-032 envelope render guard — and validate_content stands between writer and save,
   measured failing CLOSED.)
2. **The no_plan_generic conversion no-ops, exactly as bug_historian predicted — measured,
   not argued.** Canary #5 (leopardess blog, the only no_plan_generic row): router converted
   → page-build-handler → `mark_no_ready_sections` ("no sections ready to build"). And the
   round-2 "committed fallback" (ensure_page_section_layout) was investigated and REJECTED:
   `defaultSectionsForPage` has NO blog-index archetype — it would rebuild a listing page as
   hero+generic prose. → **v3 splits the route**: `no_plan_generic` converts only for
   page_type ''/content/landing; index-family pages park as `no_plan_unbuildable` naming the
   blog machinery (needs_blog_posts → blog-content-planner; bugs 015/206). The no-op
   conversion item was cancelled with a note.

**Other round-2 answers, measured:** provenance on minted items is truthful (source AND
created_by = the router); `parent_item_id` feeds `depends_on` (a dispatch gate, held seconds
— documented in CQ-023); 0 scheduled_tasks pre_queries name the type or needs_human_review;
the only other agent_definitions row naming the type is the producer's carrier
(completeness-discovery-agent); guardian's single-active-row assert added to the seed's
verify block; debug_historian's needle-gate discipline shipped as the ASSIGN file;
prior_art's IMG-071 evidence measured (rows active, 5 terminal items) and now superseded by
this router's own five live executions.

**The proliferation tripwire is now a tracked artifact**: `RFC_030` filed (three seats asked
for enforcement beyond a register sentence); CQ-023 points at it.

**THE PRODUCER GO CHANGE IS ALREADY LIVE** — the fleet rolled to `v1.0.1302` mid-session
(uniform across all 25 chassis-image pods; stamp `194907d5b…` carries `5ad81182b`; literal
probe 1/control 0). "Your commit is a deploy" in action: new required_fields_missing items
are born routed NOW. The seed's inertness assert was converted to a report (a re-apply after
the deliberate canary would otherwise wrongly abort; inertness is structural — the file
contains no site_work_items UPDATE).

**Seed v3 proven** the same way as v1/v2: exact embedded string extracted and run —
7ed472ab→no_plan_unbuildable, 4fa5b019→asset_sourced, e512af8a→no_content_data. Census v3
re-run and saved. **Fleet assignment still HELD** for round 3's verdict.
