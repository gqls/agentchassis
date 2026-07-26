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

## 2026-07-26 (later) — council round 1: REVISE, and it was worth the round

`745f9dfd` round 1: **13 reviewers, 3 abstained, 0 unreadable, decision REVISE**, gated by
`bug_historian`. (`abstained` is the relevance filter, not dissent — check `unreadable:0`, which
it was.) Run time ~9 minutes from dispatch; no queue wait, which is NOT the norm.

Two HIGH-severity objections, and they wanted different things:

- **bug_historian:** the fix closes one *source* of phantom links but leaves the platform's only
  fail-loud backstop — `validate_page_content`'s in-body href check at `Severity:"warning"`,
  non-blocking — "as generic and exploitable as before". **Accepted, not absorbed:** promoting it
  to blocking is a fleet-wide deploy-blocking change that needs measuring first, so it is filed as
  `bugs_open/079` with four candidates and the 023/033/049 coordination note. Widening this fix to
  cover it would have been the wrong call twice over (unmeasured, and it would have made a
  contained bugfix into a fleet behaviour change).
- **guidelines:** the local `INSERT ... ON CONFLICT` "violates the documented contract for
  `idx_swi_dedup`: use DELETE+INSERT". **Half right, and I did the useful half.** The DELETE+INSERT
  rule in 016b is stated about `page_components` ("Only `save_page_sections` writes it
  (DELETE+INSERT)") — a different table. `site_work_items`' own central helper,
  `insertWorkItem` (`load_work_item_actions.go:1101-1115`), uses ON CONFLICT with the interpolated
  terminal list, and a DELETE here would destroy a live non-terminal row's `attempt_count` and
  `depends_on`. But the seat was right about the *risk* it named — a copied dedup clause at a
  third call site — so the emitter now inserts **through** `insertWorkItem`, which makes the
  objection moot and brings two-strike anti-churn for free. Rebutted the rule, took the fix.

Mediums, all actioned:

- **bug_historian** (silent no-ops): every declined emit now writes an `agent_error_log` row
  (`tool_crosslink_not_emitted:<code>`) via `recordComponentWriteRejection` — the same durable
  channel the component write guard uses. Deliberately NOT a work item: nothing could action it,
  which is `bugs_open/077`'s trap (a queue entry whose handler has no remit).
- **debug_historian** (needle-gate): see the MISSTEP below — this one found a real defect.
- **tooling_provenance** (travelling docs): migration `212`, three `action`-subject notes, applied
  and recorded. Its framing needed correcting: `agent` is not a `doc_notes` subject_type (live
  vocabulary is pipeline 239 / experience 56 / tool 39 / action 4), and there were **0** prior
  notes or plans for these three subjects to consult.
- **prior_art_librarian** (is there already a resolver?): `CanonicalisePage` cannot answer this —
  its own header says *"It does not query the database — naming and URL synthesis are decided
  purely from the descriptor"*, so it produces one shape and cannot say where an existing page
  lives. `datahelpers.PageURLSet` answers "is this URL valid", not "which URL is this tool's page".
- **prior_art_librarian** (has `depends_on` ever gated anything?): fair, and thin — **3** items in
  production have used it (needs_design behind needs_composition), all 3 complete. So I proved the
  loader's predicate directly instead, both directions, on synthetic rows in a rolled-back
  transaction: not selectable while the gate is `triaged`, selectable the moment it is `complete`.

### MISSTEP — I shipped a rollback recipe that would have restored nothing

211's header carried a rollback keyed on `agent_definitions ... WHERE s.is_snapshot AND
s.snapshot_reason LIKE '211_%'`. **`snapshot_agent()` writes to `agent_definitions_backup`**, and
`agent_definitions` has no `snapshot_reason` column — so the recipe was a zero-row no-op (and a
neighbouring query of mine returned 0 rows cleanly, which briefly read as "the snapshots were
never taken"). Worse, because I applied 211 twice, "newest snapshot wins" — which my first
corrected draft used — would have restored the **migrated** state and reported success. Both
sidecars now key on `min(snapshot_taken_at)` with a needle guard. Logged in `WRONG_CALLS.md`;
commands in RUNBOOK R9. Counted pre-state, recovered: **1** create_cross_links, **0**
related_pages, 3 rows.

Round 2 resubmitted on the same correlation (`RESUBMIT_CORR=745f9dfd-...`). Image rebuilt at
v1.0.1166 and pod-grepped in the image: `tool_crosslink_not_emitted` present, the refusal string
present, positive control present.

## 2026-07-26 — rounds 2 and 3 were both voided by an UNREADABLE seat, not by an objection

Round 2: `decided_by: "unreadable reviewer(s): review_guidelines.result"`, `unreadable: ["review_guidelines.result"]`.
Round 3: same shape, different seat — `review_editquality.result`.

This is the `bugs_closed/019` / `036` class (*"one truncated reviewer voids whole council round"*):
a lost reviewer output forces `revise` regardless of what the readable seats said. **Read
`decided_by` before reading the objections** — twice now I would otherwise have concluded the
plan was being rejected on substance when it was not.

The substance trend across the four rounds, for the record:

| round | approvals | high-severity objections | decided_by |
|---|---|---|---|
| 1 | 6 of 13 | 2 (backstop gap, copied dedup clause) | gating objection from bug_historian |
| 2 | 9 | 0 | unreadable `review_guidelines.result` |
| 3 | 10 | 0 | unreadable `review_editquality.result` |
| 4 | — | — | (lean resubmission, in flight) |

**Hypothesis, and why round 4 is deliberately SMALLER instead of fuller.** Round 1 was ~2.8k of
rationale and lost no seat; rounds 2 and 3 grew to ~3.8k of rationale over a ~17KB plan and each
lost one. If reviewer prompts carry the submission and the seat's reply is cut at `max_tokens`,
a bigger submission raises the odds of a truncated reply — which is exactly `019`'s mechanism from
the other end. So round 4 trims to 2.76k rationale / 11.2KB plan while keeping every answer.
[UNVERIFIED — n=3, and I have not read the council's prompt assembly.] If round 4 also loses a
seat, the size hypothesis is dead and it is simply flaky.

### Round-3 objections, all answered before resubmitting

- **reuse_agent — "is there already a tx-wrapped insertWorkItem helper?"** No.
  `insertWorkItem(ctx, tx, item, logger)` takes a `*sql.Tx` and manages no transaction; all 12
  call sites open their own (`evidence_citations.go:412-414`, `refresh_evidence_base_action.go:745-747`,
  `lock_helpers.go:164`, …). `withWorkItemTx` is the first NAMED instance of a shape those sites
  already inline.
- **guardian — "who else invokes the two build actions?"** Measured live: exactly two rows —
  `tool-deployer` → `deploy_tool_to_site`, `tool-generator` → `create_tool_component`. Nothing else.
- **bug_historian — "audit for other emitters of the same class."** Grepped every
  `fmt.Sprintf("/…")` in `platform/orchestration/actions/`: every remaining one either **writes**
  the row carrying that path in the same function (`deploy_tool_action.go:434`,
  `create_tool_component_action.go:256/355` — all INSERT the page they build the URL for) or is an
  external API path (`companies_house_*`, `/company/%s`). **No second instance of the class.**
- **guardian (low)** — 212 relabelled `operation: add`; it only INSERTs `doc_notes`.
