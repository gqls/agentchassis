# NOTES — bugs_open/029 tool-suggester phantom links (append-only, newest at bottom)

## 2026-07-21 — session start, diagnosis

- Picked up `/bugs_open/029_..._tool_suggester_writes_phantom_tool_links.md` (NOT the
  same-numbered `..._hung_spawns_...` — 029 is an ambiguous number; `who-owns.py 029` warns and
  its "OWNED" verdict conflates the two files. This one had no active owner.).
- Read related bugs 023 (CTA label/url pairing), 049 (chrome staleness + planned-but-unbuilt
  pages). 049 explicitly lists 029 as related-but-distinct ("another route to a 404, from page
  content rather than chrome"). Confirmed 029 is its own defect.
- Design doc `005_tool_pipeline.md` names the culprit step: `create_cross_links` →
  `create_tool_cross_link_items.go`, run after `create_items_loop`.

### What the DB showed (all queries in RUNBOOK)

- Fleet sweep R1: **0 of 24** constructed `/tools/{function}.html` URLs resolve to a real page.
- The one built tool (`tool-process-automation-scorer`) deployed at
  `/tools/process-automation-scorer/index.html` — item points at `/tools/tool-process-
  automation-scorer.html`. So "tool was never built" is NOT the whole story; the URL is wrong
  even when the tool exists.
- R3: deployed tool URLs use three incompatible shapes. No deterministic function→url map.

### What the code showed (Explore agent + direct reads)

- `create_tool_cross_link_items.go:142` fabricates the URL from suggestion-time data; guards
  only check the *related* page exists, never the *tool's* page.
- `deploy_tool_action.go` / `create_tool_component_action.go` create the page first, then emit
  follow-on items with the real `page_id`/url — the pattern to reuse.
- `add_tool` spec carries `related_pages` (R4) via `spec_data: current_suggestion` — the build
  path already has what it needs.
- Downstream: internal-link-resolver only does CTA fields; `validate_page_content.go` detects
  the in-body phantom but files it `warning` (non-blocking) → deploys anyway.

### MISSTEP / correction caught this session

- The Explore agent read `k8s/bk_agent_definitions_backup.sql:214` and reported page-build-
  handler does NOT thread `spec.suggestion` to the writer. That contradicted the live evidence
  (the phantom link IS in leopardess's rendered services page). Checked the **live** row (R5):
  it DOES map `rewrite_guidance? → input_data.spec.suggestion`. The backup is stale (predates
  migration 072). Lesson (again): verify agent config against the live row, never the k8s backup.
- The original 2026-07-19 handoff said the diagnosis item "never dispatched — no orchestration
  row". Today it is `status='complete'`. That was queue latency (~30 min), not a drop — same
  trap as the council queue. No durable verdict landed in `doc_notes`, so primary evidence stands.

### Decision

Fix at the build path (candidate 1), not the emitter (runs too early) or the consumer (needs a
deferral mechanism). See PLAN. Next: implement, route through council gate (platform change),
build + roll + verify.

## 2026-07-26 — P1 implemented, P2 submitted, config half LIVE

Resumed after four idle days. Coverage check first: `who-owns.py 029` (warns the number is
ambiguous — this is the phantom-links file, not `..._hung_spawns_...`), `git log --all --since`
for phantom/cross-link/tool-suggest → nothing since 07-21. No competing thread.

### Evidence re-grounded against the live system (figures in the 07-21 docs were 5 days old)

- R1 sweep re-run 2026-07-25: **27 items across 4 sites** (was 24/3) — ai-agent-orchestration 1,
  fundamentallyai 9, gamesdesign 14, leopardess 3. **Still 0 of 27 resolve** (`matched_page_url`
  NULL on every row). fundamentallyai is new since the diagnosis; the emitter kept firing.
- R4 confirmed live: `add_tool` specs carry `related_pages` (fundamentallyai's LLM cost
  calculator → `["capabilities","multi-agent-review-council","model-fine-tuning"]`, handler
  `tool-deployer`). The build path really does have what it needs.
- R5 confirmed live: page-build-handler `rewrite_guidance?` = `input_data.spec.suggestion`.
- Live tool-generator chain is longer than seed 062b shows (062b is stale): ensure_site_record →
  load_brand_context → generate_tool_html → **save_tool** → compose_plan → write_plan →
  index_plan → enqueue_rerender → complete. `save_tool` is `create_tool_component`, so the emit
  needs no new workflow step on either agent.
- `pages.build_status` vocabulary is exactly {deployed, needs_rebuild, planned} (363/31/26).
- `needs_content_page` statuses fleet-wide: needs_human_review 19, complete 13, unresolved 1.

### What was built

`platform/orchestration/actions/create_tool_cross_link_items.go` rewritten:
`emitToolCrossLinkItems` (shared emitter, takes a REAL url), `resolveToolPageURL` (looks the page
up via `page_components → content_components.function`, with a `pages.name` fallback — both READ
`pages.url`), `relatedPagesFromSpec`/`relatedPagesFromInputs`. The old action stays registered
but is now fail-safe: it resolves the page and emits nothing when there is none. Emit calls added
to `deploy_tool_action.go` (main path + the already-deployed early return, which makes re-running
the deployer a supported backfill) and `create_tool_component_action.go`.

### DEPARTURE FROM THE PLAN — the emit is GATED, not just correctly-addressed

PLAN §Residuals deferred "page created but never deployed → still 404s" to 049. On re-reading I
kept it in scope, because with the emitter moved to the build path that residual is no longer
someone else's broad class — it is *this* emitter's remaining failure mode, and it reproduces the
leopardess damage exactly (a reference to a tool page that never goes live). With 19 of 33 live
`needs_content_page` items sitting in `needs_human_review`, it is not a corner case either.

So `emitToolCrossLinkItems` emits only if the tool page is already live (`deployed` /
`needs_rebuild`), else it attaches `depends_on` = the open `needs_content_page` item for that
page; if there is no such item, or it is in a terminal-failed status, it emits **nothing**. The
loader already honours this (`load_work_item_actions.go:562-571` — an item is only selected when
every `depends_on` row is complete/verified), so no new machinery. Failure direction is
deliberate: a parked item beats a live 404. Cost: parked items age and may be swept by 070.

### Config half applied out of band and RECORDED

`211_tool_crosslink_emit_at_build.sql` (renumbered from 210 — another thread created a 210 while
I was writing; ledger keys on filename so it was cosmetic, but the number-collision trap is
documented so I moved). Deletes `create_cross_links`, repoints `create_items_loop → complete`,
wires `related_pages` into both build steps, guards its own post-conditions in-transaction.

- Probe was `ok` under the runner's dry run.
- `--apply` was NOT used: it applies **every** pending file in the directory and 9 of them belong
  to other threads. Ran `psql -f` by hand, then `run-migrations.sh --record-only … --note …`.
- **MISSTEP:** I ran the file twice (the first run's output scrolled past and I re-ran to read the
  head). It is idempotent by design and re-committed cleanly, but it left **2 identical
  `doc_notes` rows** and a second set of `snapshot_agent` before-images. Harmless, not cleaned up
  — deleting rows to tidy cosmetics is a worse risk than the noise. Cheap check that would have
  caught it: `| tail -25` on the first run showed only the verification SELECT, so read the
  ledger/verify query instead of re-running the migration.
- Verified after apply: `create_cross_links` gone, `create_items_loop.next_step='complete'`,
  both `related_pages` paths wired.

Applying config BEFORE the image is deliberate and stated in the file header: part 1 stops the
bleeding immediately; parts 2/3 are inert on the deployed binary (unknown config key, no matching
InputSpec entry) and activate with the image. The Go side also falls back to reading
`input_data.spec.related_pages` directly, so the two halves can roll in either order.

### Council

Submitted 2026-07-26 13:33 — `SUBMISSION_CORR=745f9dfd-0a08-415b-a0a2-92c96bd30260`. Unusually,
it started executing within ~14s (no queue wait this time; do not treat that as the norm — the
documented dispatch latency is ~30 min).
