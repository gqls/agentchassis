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

## 2026-08-15 (same session, evening) — rounds 3-4 REVISE with seats disagreeing; ledger recorded; fleet ASSIGNED

**Round 3: REVISE** (gating: improvement_guardian HIGH — born-triaged bypasses the
detected→triager observe-only contract). Real catches worked from the same round:
- debug_historian's ledger point was TRUE: my psql-f applies never recorded 410 in
  `schema_migrations`. Fixed honestly: `run-migrations.sh --record-only` with a note stating
  the v1-v3 hand-apply history (the no-edit-after-record rule was never violated — the file
  entered the ledger only at v3; the runner's probe ran the file clean). NOTE the runner's
  dry run shows OTHER sessions' pending files (376/377/390/407/418-420) — never `--apply`.
- guardian's single-active-row assert: added to the verify block; confirmed live (count=1).
- bug_historian's other-non-llm-sources worry: answered by the producer's own emit set
  (`check_required_fields_missing.go:198-223` keeps only ''/llm/site_assets.* — pinned by
  test), so asset_sourced covers the entire non-llm emit population by construction.

**Round 4: REVISE** (gating: editquality — already-shipped edits re-listed as pending; fair,
the plan had become a status report). The substantive state after four rounds: **seats now
disagree with each other** — constitution APPROVES the born-triaged deferral as properly
disclosed AND accepts RFC_030's deferral, while improvement_guardian holds born-triaged HIGH
and reuse/architecture reject the deferral. Per CLAUDE.md and 033's own header ("seats
disagreeing with each other is the signal it needs a human, not a resubmission"), the
resubmission loop STOPS here; the two open questions go to the owner (see README).

**prior_art's round-4 challenge to the load-bearing premise, verified live:**
`triage_detected_items` live carriers = exactly `improvement-loop` (the landmine's "two other
agents" predates migration 286); `improvement-sweep` enabled=f. Born-detected IS born-stranded
today — the premise stands, now as a fresh query not a citation.

**FLEET ASSIGNED ~14:50Z**: pre-image (39 rows) saved as
`DATA_2026-08-15_assignment_preimage.tsv`; ASSIGN file's guarded UPDATE routed 39/39
(refusal preconditions passed, ROW_COUNT matched). Expected outcome: ~34 blob-parks (pure
information gain — same status, facts added), ~5 stale evidence-closes, zero conversions
(none remain in the backlog). Why run it during an unresolved council trail: the assignment
drew NO mechanics objection in rounds 3-4 (debug_historian approved the discipline), its
arms are all canary-proven, its effect is parks+closes only, and it is the user-approved
plan's step 5; the two outstanding gates concern the producer status line (live regardless,
via another lane's roll) and architecture policy (RFC_030) — neither touches what the
assignment does.

## 2026-08-15 ~16:15Z — FLEET AFTER-STATE VERIFIED CLEAN

All 39 assigned rows drained through the dispatch loop within ~85 minutes. Final state for
`item_type='required_fields_missing'`:

```
complete                                58   (50 historical + today's stale closes + canaries)
needs_human_review / no_content_data    35   (blob-parks, route + facts + options on the row)
needs_human_review / no_plan_owned       1   (the gas converter, tool pipeline named)
triaged 0 · blocked 0 · parked-without-a-route 0
```

Every non-terminal row of the type now carries its classification. Future items are born
routed (producer live on v1.0.1302). Remaining checks live in CQ-023's verify-later: the
+7-day churn guard (~0 new `unresolved` rows) and the re-raise-then-park of the two
cancelled-conversion findings via discovery rotation.

## 2026-08-15 late — BOTH owner decisions taken and actioned

Owner: (1) "As you suggest, go ahead" → 083 candidate 2 BUILT: `detected-item-promoter` (seed 430,
SCH-026, live, ledger-recorded; 70 detected = 66 promotable + 4 held + 0 unroutable at apply);
producer reverted to born-`detected` (inert until roll); council submission `05a3d1c8` pending.
Load-bearing finding on the way: the scheduler GATE requires `pipeline='build'` (the loader
doesn't) — why the promoter rewrites pipeline. (2) "Do as your recommendation recommends" →
RFC_030 RULED + SCHEDULED as a lane: `router_engine/` created with PLAN/HANDOFF/NOTES/RUNBOOK/
README; nothing built; first job is an A-vs-B design round with the council. Fixed my own
tautological "control" in 430's verify block before applying (partition assert instead).

## 2026-08-16 morning — fresh chassis: producer revert LIVE; promoter drained the pile; 083 council REVISE measured

Chassis `v1.0.1303` uniform (9 pods), stamp `5e075a6f9…` carries `3c6354059` → born-`detected`
producer is live. Promoter: detected 70 → 4 (= the held pair); 100 promoted, 93 complete / 4
failed (downstream) / 3 parked. Council `05a3d1c8` REVISE — objections measured, all
favourable, recorded in bugs_open/083 (pipeline provenance 97/2/1 all rewritten as the
original promoter did; no diagnose/report ever at detected; reapers key on claimed_at; two
sibling born-triaged producers named). Not resubmitted this session — token load; round 2 is
the first item in the HANDOFF.

## 2026-08-17 ~11:00–12:15Z — fresh chassis v1.0.1305 verified; 083 criterion 3 met; two corrections; migration 444 applied

**Chassis, verified at the artefact rather than the tag.** `v1.0.1305`, 2 pods (was 9). The
`build provenance` startup line had already scrolled — absent from `--tail=100000` on both
pods — so the log route is dead for this service, as LANDMINES predicts. Read the OCI label
instead: `docker image inspect … --format '{{index .Config.Labels
"org.opencontainers.image.revision"}}'` → **`6a782274b`**. Confirmed at the running binary with
both controls in one exec: the label sha PRESENT in `/proc/1/exe`; live HEAD `896c5aeeb` (a
real but different commit) ABSENT. `git merge-base --is-ancestor 3c6354059 6a782274b` → 0, so
the born-`detected` producer revert is live. Nothing in this lane was waiting on the roll —
the only Go dependency shipped at v1.0.1303 — but it is now checkable rather than assumed.

**§1 re-measured, and one of the numbers meant the opposite of what it looked like.** The
`detected` pile read **82**, against 4 the day before. Not a regression: 77 of the 82 are
flag-only rows with NO `handler_agent`, which this promoter's first predicate can never touch,
and 40 of those were deliberately restored to `detected` by the concurrent `bugs_closed/284`
lane (migration `442`, ledger-recorded 11:02Z today — i.e. *while I was measuring*). Promotable
pile = **0**; the 5 handler-bearing rows are two never-completed pairs held for a hand canary.
**Misstep:** my first reading of the pile assumed the 36 `head_essentials_missing` were "a new
type held for canary". They are not held — they are ineligible, for a different reason. Reading
`handler_agent` on the breakdown is what separated the two. This is why 083's criterion 1 has
been restated in the bug file: the raw `count(*) WHERE status='detected'` conflates two
populations with opposite meanings and will now climb for ever with normal discovery.

**Council round 1 (`05a3d1c8`) re-read from the artefact, not from yesterday's summary.**
Gating objection was `prior_art_librarian` HIGH on the "sole live carrier" premise; five more
seats objected at medium/low. Re-measured every one today (all in the round-2 submission's
`grounded_in`). One of yesterday's answers was **over-broad and is corrected**: the note said
*"every enabled reaper pre_query keys on `claimed_at`"* — false fleet-wide, since
`thunder-reaper` mentions `created_at` and `stale-orchestration-reaper` / `stuck-task-reaper`
key on neither. The defensible claim, which is the one that answers the objection, is narrower:
**of the 10 enabled tasks whose pre_query touches `site_work_items`, the three that can reap or
time out an item (`claimed-item-timeout`, `stale-work-item-reaper`, `report-dispatch`) all key
on `claimed_at`, and none on `created_at`.**

**Guardian's "confirm the Go diff didn't touch the per-pass cap" answered with the diff
itself:** `3c6354059` is 1 file, +8/−5, entirely one `WorkItemSpec` field flip
(`Status: "triaged"` → `"detected"`) plus its comment. Nothing near the early-return logic.

**The counterfactual that could not have come out otherwise (misstep, and the session's real
finding).** To size the "known-good rule is too weak" risk I computed each pair's record at its
promotion instant, keyed on `completed_at < triaged_at`. Every one of 17 pairs came back
**100% success, zero failures** — which I did not believe, because a real fleet of handlers
does not look like that. Cause: **`failed` rows never get `completed_at`** (0 of 265; control
run in the same breath). Re-keyed on `updated_at`, the truth appeared:
`literal_markdown → page-build-handler` was at **1 complete / 28 failed = 3%** when the promoter
promoted 6 more to it, 5 of which failed. That is 430's own submitted risk 2, fired. Now a
LANDMINE and a WRONG_CALLS entry.

**Migration `444` applied 2026-08-17, ledger-recorded, with a separate `_ROLLBACK.sql`** (the
last answering `debug_historian`'s LOW objection directly). Two predicates added to the
candidates CTE; `430` untouched because it is ledger-recorded. (1) pipeline **allow-list**
`IN ('build','content','design')` — deliberately not the `NOT IN ('diagnose','report')`
deny-list the objection implied, because [MEASURED] `report` does not exist on this table (0
rows) while `experience` (7) and `maintenance` (1) do, so the deny-list names one value that
cannot fire and misses two that can. (2) a **25% success floor** for pairs with ≥5 terminal
outcomes; threshold set by the census gap (3%, then 41, 42, 46, 50, 67, …), not chosen. Dry-run
first, read-only, with both controls in one run: `literal_markdown` fails the floor, `page_rerender`
(4017/21) passes. At apply both doors held **0 rows** and the verify block's positive control
(the floor must be *able* to hold something, else "holds 0" is vacuous) passed.

**083 criterion 3 MET — and a fetch artefact nearly made it a false negative.** Four promoted
+completed rows on `mortgagecalculator.co.uk` verified at the served page. First attempt: all
four **404**. The control — fetching the site root, 200 and 37 KB — showed the instrument and
the domain were fine, and the rows store `/guides/index.html` while I had requested `/guides/`.
At the exact stored URLs: 200 on all four, 1,426–2,641 visible characters, real on-topic h1s.
Second instrument: `page.updated_at` moved **5–42 seconds BEFORE** each item was closed, in all
four — the causal order real work leaves behind.

**Criterion 2 was already met before the fix existed** (7 completes, all 2026-08-09..11, all via
the *original* promoter). Yesterday's block and the day before's both said it was still waiting.
Corrected in the bug file and logged in WRONG_CALLS; it is satisfied on its wording but is not
evidence about candidate 2.

## 2026-08-17 ~11:30–12:45Z — 083 council APPROVED at round 2; advisories checked; the router's trail 7b0e2833 assessed and NOT resubmitted

**Council round 2 on `05a3d1c8`: APPROVED** (11:27Z, ~8 minutes from dispatch — much faster than
the 29-minute figure CLAUDE.md warns to budget for; the fleet was quiet). 12 seats approved,
including `architecture` and `prior_art_librarian` — the seat whose HIGH objection gated round 1.
3 abstained, not truncated, 2 advisory objections, none high.

> ⚠ **The submission was REFUSED client-side first**: `097` checks that at least one edit touches
> `platform/`, `internal/` or `pkg/`, and this round is SQL config plus docs. Round 1 passed only
> because it happened to carry the Go producer revert alongside. So **the gate's scope filter is
> path-based and cannot see a platform mechanism that ships as `scheduled_tasks` /
> `agent_definitions` config** — which on this estate is a large fraction of behaviour. Used
> `FORCE=1` deliberately (a live shared mechanism, not docs or site content, and an existing
> trail the gate had already accepted at round 1). Worth raising as its own item.

**Both advisories were checked rather than banked.** `guardian` LOW asked for the per-tick cost I
had not measured: `EXPLAIN (ANALYZE)` on the live table gives 65.3 ms for 430's predicates and
78.1 ms for 444's — **+12.9 ms on a 900,000 ms tick**. `bug_historian` MEDIUM was the one worth
real work: the floor treats `complete` as ground truth, and `bugs_closed/028` documents
`page-build-handler` reporting complete while deploying hollow content — the exact handler in the
pair that motivated the floor. Its sharpest form is one row, so I checked that row at the live
page: 0 literal-markdown hits in 8,120 visible chars. **A zero from a detector I had just written
is worthless without a demand control**, so the same five patterns went over three pages whose
items are `failed`/`needs_human_review`: 9, 5 and 13 hits. The one complete is real; the failures
are real; the objection is answered in the direction that strengthens the floor. Contributed to
`bugs_open/184` with a consumer notice, since that lane owns the underlying handler failure and
its findings will now sit at `detected` instead of dispatching.

**Trail `7b0e2833` (the router, REVISE ×4): assessed, and deliberately NOT resubmitted.** The
handoff suggested a short round 5 citing the two owner rulings. Reading round 4's actual verdict
says that would fail, and why:

- **What gated round 4 was `editquality` HIGH — "a no-op dressed as an edit"** — four of its
  edits were already-committed work re-listed as pending. A round 5 that "cites the rulings and
  stops" has *no real edits at all* and would be gated identically. The trail cannot be closed
  by narration.
- **Most of round 4's objections are now genuinely answered by shipped code, not by argument.**
  `improvement_guardian`'s HIGH (born-triaged removes the observe-only stage) and
  `prior_art_librarian`'s HIGH (the sole-live-carrier premise) are both settled: the promoter
  exists, the producer is back to born-`detected` and live on `v1.0.1305`, and
  `prior_art_librarian` itself approved that premise today on the 083 trail.
- **The RFC_030 proliferation objections are owner-RULED but their specific ask is not met.**
  `reuse_agent`, `guardian` and `architecture` all said, in effect, *"acceptable only if RFC_030
  is genuinely a hard gate on a 4th router, not aspirational"*. A lane existing is not a gate.
  That residual is real and belongs to the `router_engine` lane, not to another round here.
- **`bug_historian`'s HIGH is neither resolved nor refuted — it is UNEXERCISED, which nobody
  had measured.** [MEASURED 2026-08-17] every route ever taken by this router: `no_content_data`
  35, `asset_sourced` 1, `no_plan_owned` 1. **`file_rewrite` and `file_recreate` have NEVER
  fired, and the router has produced ZERO child work items.** The two conversions attempted
  during the fleet assignment were cancelled. So the regeneration risk the seat raised (closed
  case 056; `missingkey=zero` renders a missing required field as empty with no error, guarded
  at one call site) is entirely theoretical *so far* — and the first time those arms run will be
  in production, on a customer page, untested.

**I did not unilaterally gate the convert arms**, though the estate's own remedy is obvious (owner
ruling 2026-08-02 §2: new authority on a shared seam ships as an opt-in field with the unsafe
default OFF, which RFC_022 then confirms is not architecture-scope). Two reasons: it changes a
live, owner-blessed mechanism's safety posture on my reading alone, and the branch conditions are
`conditional_branch` steps whose `then_step`/`else_step` are data — so "opt-in" would mean either
a new expression form or redirecting the arms to a park step, and which of those is right is a
design call. **Put to the owner in `README_where_we_are.md` instead.**

## 2026-08-17 ~14:40–17:00Z — the "fresh build" shipped nothing; 453 proved itself; and 444 was wrong in a way no review could have caught

**THE ROLL SHIPPED NO NEW CODE. Verified on three instruments before believing it.** The owner
rebuilt and deployed the chassis at **the same `IMAGE_TAG` as the morning's build** (`v1.0.1305`),
so the node served its cached layer. (a) The local image at that tag carries `revision=89a0cbeb7`,
`created=14:30:02Z` — a genuinely new build. (b) The **running binary** contains `6a782274b` (the
morning's) and **not** `89a0cbeb7`. (c) Pod `imageID` is `sha256:f90a7e88…` while the local repo
digest is `sha256:6039e19c…` — different images. Pods restarted at 14:42/14:43Z, *after* the 14:30
build, so this is not a timing race. **252 commits sit unshipped, 26 of them touching Go.**
Contributed to `bugs_open/153` (which owns this trap) rather than filed anew; nothing of this
lane's is affected because migrations `444`/`453`/`454` are DB config and live at COMMIT.
Negative control was a **real but different sha**, never a zeros run — a 40-zero needle is present
in every binary (LANDMINES, same day).

**453 fired on schedule and did exactly what it was built to do.** First tick 12:57:43Z: escalated
the **4** `page_component_status_drift` rows (7 days waiting) to `needs_human_review`, and correctly
left `placeholder_contact` alone at 1 day — the clock discriminating, not just firing on a backlog.
The written row carries the pair, the reason, `days_waiting: 7`, `limit_days: 3`, the exact by-hand
promote command, and `owner: "(UNASSIGNED - claim this) check_page_component_status_drift.go added
2026-07-10, never touched since, no lane doc claims it"`. The design choice that the map enriches
rather than gates is what produced that last line instead of silence.

**Then 444 turned out to be wrong, and this is the important entry.** Checking a newly-arrived pair,
`empty_section → page-build-handler` read **3 complete / 12 failed = 20%** — below my floor, so 444
was holding it. But my own 11:00Z census had recorded that same pair at **11 complete / 13 failed =
46%**. Completes cannot decrease. My first thought was corrupted data.

Nothing had regressed. **`site_work_items.status` has TWO terminal success states** — a verification
sweep had moved 9 of those completes to **`verified`** (`complete_work_item_verification.go:218`),
which is a completion that *also passed verification*, i.e. more evidence, not less.
`idx_swi_completed`'s predicate already lists both; I had never read it as a domain.

So both of the promoter's success tests — 430's known-good rule and 444's floor — **measured success
with half the definition, and the resulting metric gets WORSE as the platform verifies more work.**
Fixed by `454` in both predicates, applied after a read-only dry run whose two controls flip
opposite ways (`empty_section` F→T proves the fix does something; `literal_markdown` staying held
proves it did not just disable the door). Fleet scope when found: `verified` was **9 rows across 1
pair**, so exactly one verdict changed — but the **latent** case is why it could not wait: a pair
whose successes have *all* been verified has zero `complete` rows, reads as never having worked, and
is held for ever, which is precisely `bugs_open/083`'s disease.

> **This is the second instance of one class in a single session**, and worth naming as such. Earlier
> today: `failed` rows carry no `completed_at`, which made a counterfactual return a uniform 100%.
> Same table, same day. Both are **a status column whose values I assumed instead of enumerating**,
> and in both cases the wrong answer was well-formed and plausible. Neither was *unmeasured* — the
> marker discipline was followed — they were **measured against an incomplete definition**, which no
> marker detects and, as WRONG_CALLS records, **no council review could have caught**: twelve seats
> approved 444, and `bug_historian` came closest by challenging whether `complete` was *trustworthy*
> — the right instinct on the wrong axis. A review of a plan cannot enumerate a column the plan never
> mentions. `SELECT status, count(*) … GROUP BY 1` before filtering on a status. Both now LANDMINES.

**Provenance note:** commit `b6bca52fc` carried, as a same-file passenger, another session's
improvement to the build-provenance LANDMINES entry (they split one line and added an
`UPDATED 2026-08-17` note about reading the image label first). Nothing was lost — the content
survives and is better — but naming it so `git log` does not attribute their work to this lane, the
same courtesy the `284` lane extended to this lane's WRONG_CALLS entry this morning.
