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

---

## 2026-08-17 evening — session `bugfix-083`: the canary ran, and the residual's remedy was inoperative

Picked up this lane's own `HANDOFF_2026-08-17` §5 items 1 and 3, from a fresh chat. Contributing
into the shared account rather than opening a `bugfix_083_*` lane, per CLAUDE.md.

**Owner decisions re-confirmed, not re-opened.** Decision 1 (gate the router's convert arms) came
back to me as *"gate them"* at ~16:15Z, which contradicted the *"leave as is"* this lane recorded at
12:53Z and shipped in `f1a5b6315`. I surfaced the contradiction with both timestamps rather than
acting on the later answer, re-explained the trade-off (including the 2026-07-29 counter-rule that a
gate nothing reaches is a mechanism rotting unexercised), and the owner confirmed **leave as is**.
No change to `file_rewrite` / `file_recreate`. Recorded here so a fourth session does not re-raise it.

**The canary (decision 2's other half).** `453` built the clock; nobody had been the human. Full
record in `bugs_open/083`. Three things worth having in the technical log:

1. **Validity-check first.** 1 of the 4 escalated rows named a `page_components.id` with **0 rows**
   in the table — repaired by an ordinary re-render five days after filing. `fixPageComponentStatus`
   returns a hard error on `sql.ErrNoRows`, so canarying that row would have written a `failed`
   against the pair, and post-`444`/`454` failures are *scored*. **The canary meant to qualify a
   pair can push it under the floor.** Filed as `bugs_open/300`; the row was closed by hand as
   `complete` / `resolution_path='manual:revalidated'` with the proof in `result.revalidation`.
2. **`430`'s prescribed canary UPDATE ends `AND status='detected'`** and matches **0 rows** once
   `453` has escalated — `UPDATE 0`, no error, reads like a completed step.
3. **`453` is a one-way door.** Nothing returns an escalated row to `detected`, so a newly
   known-good pair's remaining findings stay stranded. Returned by hand; the promoter then took
   them within 4 minutes. The mechanism fix belongs to `453`'s author.

**The residual, and why it was bigger than it looked.** 083 recorded the guardian's cheap control
as *"return a third column counting held rows so the scheduler's log makes it visible"*. Measured at
the running service, that column would have been **write-only**: for `fire_message=false` the
pre_query result is merged into `inputData` and then discarded — `fireTrigger` is never reached — so
`promoted` and `pairs` had never been readable either. Every tick emitted the same bare line.

So the fix went into `cmd/scheduler/main.go` (every CTE-only task logs its own result; the 2 KB cap
**announces itself**, because a silent truncation would rebuild the defect inside the fix) plus
migration `458` for the numbers. Commit `03012d862`, council `8dc58e2a-…` submitted with `FORCE=1`
— `097` scopes to `platform/`/`internal/`/`pkg/` and this is `cmd/` + config, the blind spot §5
item 6 of the handoff already names.

**Two findings from writing `458`, both of which change what I believe about this task:**

- **The row-suppression idiom never worked.** `... FROM promoted WHERE (SELECT COUNT(*) FROM
  promoted) > 0` suppresses nothing — aggregate-only target list, no `GROUP BY`, so exactly one row
  returns regardless. Verified read-only against an empty CTE (`WHERE` → `('0', NULL)`; `HAVING` →
  0 rows). Consequence: the promoter claimed its `maintenance` concurrency slot on **every** tick,
  defeating `048`'s release-on-no-op. Corrected to `HAVING COUNT(*) > 0 OR held > 0` — suppress on
  *both* being zero, because an idle tick that is holding rows is the state the residual is about.
  **Sized honestly: a door, not a repair** — no `maintenance` group-mate is starved today (all five
  within 1.04 intervals of due). ⚠ `453` carries the identical idiom; its lane's call.
- **Concept-register SCH-007 is stale and is what made the broken idiom look mandatory.** It says a
  pre_query "must return at least one row, or the timestamps never advance". False since
  `dc2e4b61a` (`bugs_open/048`): the zero-row path stamps both and logs *"Pre-query found no rows"*
  (`main.go:200-216`). Confirmed at the running service in the same hour — `feasibility-recheck`,
  `report-dispatch` and `diagnose-pipeline-trigger` all log exactly that line and keep their
  cadence. Entry corrected with the cost named.

**`458` reports `held=2` on its first run, and one of them is `444`'s own floor holding a live row
for the first time** (`literal_markdown → page-build-handler`). `444` recorded both doors as holding
zero and called them "doors, not repairs"; one is now load-bearing and nothing could see it. The
migration's verify block carries that as a **positive control** (held must be non-zero or the
counter is unproven) and a **negative control** (the promotable set must be unchanged by the
restructure — 0 before, 0 after).

**The counter inherits this file's own 18:30Z OPEN RISK** (14 completes left the table, cause
undiagnosed): `held` is a live `COUNT` over `site_work_items`, like the rule and the floor it
reports on, so a pair could be reported held with the confident-but-wrong reason *"has never
completed one"*. Flagged in `bugs_open/083` and in SCH-026 rather than left for the two accounts to
drift. If `090` finds the actor and the answer is a durable per-pair tally, this counter should read
that tally.

**A wrong turn, logged in full in `WRONG_CALLS.md`.** Before settling on `bugs_open/300`'s narrow
finding I tried to justify a *general* "is this finding still true?" predicate, using "was the page
redeployed after the finding was filed?" as the staleness proxy. It came out backwards — 86% success
for "stale" rows against 41% for fresh ones — because `pages.deployed_at` **is written by the
handler's own work**, so the instrument was moved by the thing it measured. No predicate built. The
tell was the result pointing the wrong way hard enough to be absurd; a confounded instrument that
had landed 60/40 in the expected direction would have gone straight into a bug file.

**One refuted hazard, recorded because it was plausible and specific.** A warning that
`component-template-fixer`'s `create_rerender` step INSERTs into a `domain` column dropped from
`site_work_items` — which would have failed *after* the repair was applied — is **false at the live
config**: the live agent definition names `pipeline`. The artefact that said otherwise
(`k8s/bk_agent_definitions_backup.sql`) is a snapshot dated 2026-03-12. Reading the live row rather
than the seed is what caught it, and it is the reason the canary was safe to run at all.

**Council round on the visibility fix: APPROVED at round 1** (corr `8dc58e2a`, 22:07:31Z), 4
advisory objections, none high; `architecture` returned `point_fix`. Three were checkable and were
checked (full record in `bugs_open/083` §7):

- `guardian` was right that I had **asserted** the other CTE-only tasks select nothing sensitive.
  Now measured over all 9: **zero** mention password/secret/credential/api_key/b2_/aws_/smtp/
  authorization, **zero** mention email, two mention "token" — both LLM *token-pressure* monitors
  ending `SELECT id::text AS note_id`. Clean today; the standing hazard is a future author, which
  is what the comment at the log site addresses.
- `guardian` also asked whether the non-suppressing `WHERE` exists elsewhere. **Exactly one other
  task**: `held-pair-canary-escalation` (`453`) — bounded at 2 of 9, both this lane's.
- `debug_historian` asked whether the held census had been gathered by running the live pre_query
  (which embeds `UPDATE … RETURNING`) and thereby promoted rows as a side effect. It had not: the
  census was a **read-only mirror** (`scored` alone), and the full query was only ever exercised
  inside `BEGIN … ROLLBACK`.
- **`editquality` and `tooling_provenance` were simply right about a submission defect:** I claimed
  in `grounded_in` that SCH-007 is corrected, but listed no edit touching the register. The
  correction was real and was made — it just was not in the edit list, so no seat could see it.
  **An edit list that omits work you have actually done reads exactly like a claim you cannot
  support.** Worth remembering next submission.

Two advisories accepted and not fixed, named so the silence is a decision: the forward migration
`RAISE`s on the already-applied case where a 0-row no-op would be tidier (the rollback has the
symmetric branch; erring loud is the safe direction), and `truncateForLog` is package-local —
checked, and all five existing `truncate*` helpers are unexported inside
`platform/orchestration/actions`, with only `fetchguard.LimitedRead` exported and unrelated.

**Told the `bugs_open/297` lane about the duplicate migration number** (their `453` is the unapplied
one, so renumbering is free; `459` was clear at 22:15Z). Not edited by me.

## 2026-08-18 — post-roll: the residual is CLOSED at the service, and it immediately found 14 rows nobody could see

`kafka-scheduler:v1.0.1309`, pod started 15:45:39Z, own `build provenance` line
`git_commit=f0117fb8b`; `03012d862` confirmed an ancestor. Full record in `bugs_open/083` §1-5.

**A control that could not fail, recorded because it briefly alarmed me.** My first negative control
was "my last commit must NOT be in the build" — it read as an ancestor. That is not a failure: the
build post-dates all of yesterday's work. A commit that cannot be absent is not a control. Re-run
against a commit made *after* the build revision (`3f1426a8d`), it behaves. Same discipline as
`444`'s positive control, applied to git rather than to SQL.

**The residual is closed on its real criterion** — not "the column exists" but a live tick's held
count legible in the service's own log:

```
"pre_query_result":"{\"held\":\"16\",\"held_detail\":\"… literal_markdown->page-build-handler
 (pair below the 25% success floor); placeholder_contact->page-build-handler (never completed) …\",
 \"pairs\":null,\"promoted\":\"0\"}"
```

and the two branches now discriminate (`thunder-reaper` takes `:218` "found no rows"; the promoter
takes `:286`). Before this build both emitted the identical sentence.

**Held: 2 rows → 16 rows across 5 pairs, overnight, invisible the whole time.** `literal_markdown →
page-build-handler` alone is 10 rows on 2 sites, held by `444`'s floor at 3 ok / 24 failed = 11%.
The other four pairs (6 rows) have never completed one and are waiting on a hand canary; two of them
(`dead_fragment_link`, `missing_conversion_path`) have never been dispatched at all, so their handler
is genuinely unmeasured rather than known-bad.

**⚠ CLOCK: `453`'s one-way door fires tomorrow.** `result ? 'held_pair_escalation'` is 0 across all
16 — the limit is 3 days and the oldest held row (`placeholder_contact`, 2026-08-16) is at 1.9 days,
so it crosses on **2026-08-19**. Those 3 rows then move to `needs_human_review`, which the promoter
never selects and nothing returns from. Yesterday this interaction had 5 rows behind it; now it has
16, first 3 due within a day. Not patched here — the door was deliberate and reclaiming escalated
rows after a successful canary is `453`'s author's call or the owner's.

### 2026-08-18, later — the two remaining canary candidates were assessed and NEITHER was run; both reasons are findings

Post-465 the held set is **15 rows in 4 pairs** (`literal_markdown` 10, `placeholder_contact` 3,
`dead_fragment_link` 1, `missing_conversion_path` 1). Of those, two pairs read 0 complete / 0 failed
— never dispatched at all — and looked like the obvious canaries. Both were refused, and not for
the same reason.

**`missing_conversion_path → content-gap-planner` — DO NOT CANARY. `bugs_open/255` already owns it.**
Filed 2026-08-11 by the `vigilant_designer_offer_analysis` lane (whose check it is), diagnosis loop
**CONFIRMED first iteration**: the type *"is routed at a handler that cannot read its spec, so it is
refused and released back to the detector"*. A canary would fail for that documented routing defect,
move the pair from 0/0 to 0/1, and record a route problem as handler incompetence. **A pair with no
history is not a pair with no known problems** — and the queue cannot tell you, because the absence
of rows is the whole point. New sub-bullet in `LANDMINES.md`: grep `/bugs_open/` for the item type
AND its handler before promoting. This is the same trap as the stale subject, arriving by a second
route.

**`dead_fragment_link → page-build-handler` — VALID but needs the owner, because the risk class
differs from yesterday's canary.** The finding is verified still true at the served page:
`vetcomparison.uk/tools/pet-treatment-cost-estimator/index.html` returns 200 and carries
`href="#directory"` with **no** `id="directory"` anywhere (the page holds three ids, all
tool-internal; negative control `id="zzyzx"` correctly absent). So a visitor clicking that control
gets nothing.

But yesterday's canary flipped a status column — reversible, `rendered_html` untouched. This handler
**regenerates the page**, and the page is a live tool, where `bugs_open/263` (decomposition dissolves
the tool's wrapper), `286` (the generator cannot rebuild a tool at an existing page) and
`208`/`226`/`238` (regeneration drops fields) all apply. Yesterday's approval was for a canary of a
different kind, so it does not obviously extend. Also worth the owner's attention: the spec's own fix
text says the site-surface remedy is *"re-render site components"* because *"the chrome link resolves
on NO page of this site"* — which suggests rebuilding this one page may not fix the class at all.
Second `LANDMINES` sub-bullet: read the handler's branch for *this* `fix_type`; "the last canary was
fine" does not carry across.

## 2026-08-18 ~15:45–17:00Z — the archiver found; 465 and 466 applied; four defects of my own closed

**Chassis `v1.0.1309` verified shipped** (label `f0117fb8b`, present in the binary, superseded
`a6d1c53c0` absent, 28 behind HEAD — ordinary churn). A real tag bump this time, so no repeat of
the 08-17 cached-layer no-op.

### The UNDIAGNOSED risk had a one-query answer, and I had run three searches that could not find it

`work-item-archiver` — enabled, daily, description *"Archives terminal work items older than 7 days
to `site_work_items_archive`"*. The archive holds **20,184 rows against 8,702 live**. So both of the
promoter's success tests, reading only `site_work_items`, meant **"in the last 7 days"**.

What found it was the archiver's name in an unrelated list of `maintenance` concurrency-group
members inside migration `458`'s header — peripheral vision, not a search. Why the searches failed
(all three logged in WRONG_CALLS + LANDMINES): it **moves** rows so there is no `DELETE`; its
`pre_query` is **NULL** because it is `fire_message=true`, so its SQL is in an agent definition and
not the column I was grepping; and its description says *"Archives"*, not *"retention"*. And my
reassuring control — *"the oldest surviving row is 2026-03-15"* — was blind by construction: that
row is **non-terminal** and the archiver only takes terminal ones.

**It was already causing damage, not latent.** `empty_internal_href → page-build-handler` read 0/1 =
0% live and **9/5 = 64%** true — held by the known-good rule as "never completed" while holding nine
lifetime successes. `empty_section → page-build-handler`, a **316**-success workhorse, read as
marginal at 43% against a true 91%.

**Migration `465`**: both tests now read live UNION archive. Three controls; the two negatives carry
the weight (`literal_markdown` must stay held — if the archive had rescued the pair `444` exists
for, the change would be dissolving the floor rather than fixing its scope; `placeholder_contact`
must stay unknown). Cost 78 ms → 134.6 ms per 900 s tick. A shared VIEW is the tidier answer and was
deliberately not taken — a new shared object is a shared-seam change (2026-07-28 ruling).

### Migration `466` — three corrections to my own `453`, and the third was a live contradiction

(a) the non-suppressing `WHERE` → `HAVING` (found by `458`'s author, left to me); (b) floor-held
pairs now escalate, with **opposite** remedy text — a canary on a floor-held pair adds a failure and
pushes it further under, which `bugs_open/300` is the live case of; (c) **the one I had not spotted**:
`453`'s held test asked `NOT EXISTS(status='complete')` over the live table alone, so after `454` and
`465` it disagreed with the promoter — for `empty_internal_href` it saw **0** successes where the
promoter saw **9**, and would have asked a human to canary a pair the promoter promotes unattended.
Both tests are now one test. Live after apply: watching 15 rows, 5 canary-held, 10 floor-held.

Shape is ONE `classified` CTE evaluated once, not a second negated copy of the promoter's
predicates — (c) is precisely that drift already realised, over three days and three migrations.

> **CORRECTION to commit `a62809d29`'s successor (the 466 commit): two phrases were eaten by
> unquoted backticks in `git commit -m`** — the documented shell trap, which I hit anyway. The
> message reads *"Shape is one  CTE evaluated once"* (lost the word `classified`) and *"the negative
> control first read , always zero"* (lost the formula). **The lost formula is the substance**: it
> was `promotable AND NOT promotable = 0`. Restated here because forward-only forbids an amend.
> Use single quotes in `-m`, or write the message to a file and use `-F`.

### THIRD tautological control caught in this lane — the count for one is now the habit

`466`'s negative control first read `promotable AND NOT promotable = 0`: always zero whatever the
code does. Replaced with a named discriminating case — `empty_internal_href` must be ABSENT from the
watched set, which under `453` it would not have been. Previous two: `453`'s draft
(`EXISTS(X) AND NOT EXISTS(X)`) and — per its own header — `430`'s, caught by its author before
applying. **The test is not "is this control true?" but "could it ever have come out non-zero?"**

### The family is now FIVE, and the shape never varied

`failed` rows carry no `completed_at` · `verified` is a second success status · the row set is not
stable · the row set is only a 7-day window · and a control that cannot come out otherwise. Every
one is **a population or a domain assumed rather than enumerated**, and none was caught by review —
twelve council seats approved `444`.

## 2026-08-18 ~18:15–18:40Z — `466` verified live; its clock re-predicted against TICKS not dates; and the floor's protective-refusal hole MEASURED

**Build:** `kafka-scheduler` on `0b185bad2a49c6e032352fa9e7d0b429f0a95104` (own `build provenance`
line, pod 17m old) — a roll newer than the handoff's v1.0.1309. Chassis provenance was **absent from
`--tail=3000`**, which is the documented "not in range" case for a busy service, not "unstamped".

### `466` is live and is the text that is running

Dumped `scheduled_tasks.pre_query` for `held-pair-canary-escalation` and read it: the `hist` CTE
UNIONs `site_work_items_archive`, `classified` carries `hold_kind`, the suppression is `HAVING`, and
the two remedy strings are opposite (canary → "run one by hand"; floor → "do NOT canary, FIX THE
HANDLER"). So all three of `466`'s corrections are in the deployed row, not just in the file.

### ⚠ CORRECTION to my own handoff §3.1 — its escalation dates are ONE TICK EARLY

The handoff predicted `placeholder_contact` escalates **~08-19** and `literal_markdown` **~08-20**.
Both are wrong, and the error is the one this lane keeps making: **computed on DATES, when the
mechanism runs on TIMESTAMPS.** The task is `interval_seconds=86400` and last fired
**12:57:48 UTC**, so it ticks at 12:57 each day; the predicate is
`min(created_at) < now() - interval '3 days'`. A row created at 19:17 is not 3 days old at 12:57 —
it misses by 6h20m and waits a further day.

[MEASURED 2026-08-18 18:19Z, read-only run of `466`'s own `classified` CTE]

| pair | hold_kind | rows | oldest (UTC) | 3d due | FIRST TICK THAT ESCALATES |
|---|---|---|---|---|---|
| `placeholder_contact → page-build-handler` | canary | 3 | 08-16 19:17 | 08-19 19:17 | **08-20 12:57** |
| `dead_fragment_link → page-build-handler` | canary | 1 | 08-18 01:38 | 08-21 01:38 | **08-21 12:57** |
| `literal_markdown → page-build-handler` | floor | 10 | 08-17 19:21 | 08-20 19:21 | **08-21 12:57** |
| `missing_conversion_path → content-gap-planner` | canary | 1 | 08-17 22:21 | 08-20 22:21 | **08-21 12:57** |

**So tomorrow's tick — the first under `466` — escalates NOTHING.** It should log `escalated=0` with
`watching=15`, which is `466`(a) working (a `HAVING` that still speaks on an idle tick). **Do not
read that zero as `466` being broken**; that misreading was set up by my own handoff. The effective
limit for a daily task with a 3-day predicate is 3–4 days, not 3. Conditional on the held set not
changing: if the oldest row leaves `detected`, `min(created_at)` moves later and pushes the date out.

### Handoff item 2 — the floor counts protective refusals as failure. MEASURED, fleet-wide.

Population: **948 `failed` rows**, `site_work_items` UNION `site_work_items_archive` (lifetime, which
is what the floor reads after `465`). Classifier is string-matching on `error`, with **D as the
residual bucket** — so an over-broad A/B rule would *shrink* D, and D is the number I lean on.

| class | rows | % | what it is |
|---|---|---|---|
| **A protective refusal** | 434 | 45.8% | handler correctly declined — 418 `rebuild_policy=owned`, 10 section-shrink, 6 overwrite |
| **B transient / infra** | 234 | 24.7% | 93 claim-timeout/pod-died, 78 LLM-call failed (**debatable**), 41 delivery, 22 other |
| **C housekeeping** | 110 | 11.6% | 52 `bugs_open/017` backfill, 30 no error at all, 28 manual cleanup — **never a handler attempt** |
| **D genuine non-repair** | 170 | **17.9%** | handler tried and produced nothing |

**Only ~18% of what the floor counts as handler failure is the handler failing** — and that is an
**upper bound**: I found ~9 rows misfiled INTO D (4 `diagnosis exceeded 75m — handler pod likely
died`, which my `Claim timed out%` pattern missed, and 5 `claims floor blocked`, which is another
protective refusal). Both corrections push the same way.

### But NOTHING is mis-gated today — and I nearly recorded that it was

First pass said `placeholder_contact → page-build-handler` **flips** to promotable under a corrected
floor (6 failures, **0** of them genuine). That was **my own tautology-family error, caught before
recording**: I had tested the FLOOR formula alone, but the promoter's predicate is
`c = 0 OR under-floor`, and this pair has **zero successes**, so the *canary* rule holds it whatever
the floor says. Re-run against the **full** predicate: **0 pairs flip, 0 with rows waiting.** The
handoff's "no pair is mis-held today" is confirmed — now on the full population rather than one pair.

`literal_markdown → page-build-handler` refines too: the handoff estimated 3/(3+14) ≈ 18% from 24
live failures; the true lifetime figure is **3 successes / 36 failures, of which 16 protective and 16
genuine → 3/(3+16) = 16%**. Still under 25%, still correctly held. Verdict survives, better number.

### The reason it is NOT merely theoretical: refusals are a RATCHET, and they are now the majority

[MEASURED] Protective share of `failed` by month: 04 **0%** · 05 **0%** · 06 **0%** · 07 **1.5%** ·
**08 61.6%**. The `rebuild_policy=owned` guard first fired 2026-07-17; it is **bursty**, not a slow
drift (08-08: 123 of 134; 08-09: 102 of 131; **today 66 of 74 = 89%**).

The floor reads **lifetime** with no window, so a refusal never ages out. **One batch sweep across N
owned pages permanently adds N to a pair's denominator, and a held pair gets no dispatches, so it can
never earn the successes to climb back out.** `literal_markdown → page-build-handler` needs **9** more
successes to clear 25% counting all 36 failures, but only **3** counting the 16 genuine ones — and
being held, it will get neither unattended.

> ### ⚠ DATED, FALSIFIABLE: the 08-21 escalation will MISDIRECT the human it pages
> `466`(b)'s floor-held remedy text says *"this handler is failing … FIX THE HANDLER, or decide the
> pair is wrong and retire the producer."* When `literal_markdown → page-build-handler` escalates on
> the **08-21 12:57** tick, **16 of its 36 lifetime failures are the handler correctly refusing to
> clobber owned pages** (`bugs_open/295`'s family). The payload will send someone to fix a handler
> that is, in 44% of its failures, behaving exactly as designed. This is my defect, in text I wrote
> yesterday, and it has a date on it.

### The fix is NOT to put this classifier in the gate — and that is the finding, not a dodge

Encoding `error ILIKE '%rebuild_policy=owned%'` into a live gate makes **an error message's wording
load-bearing fleet-wide**: anyone rewording that string silently changes what the promoter dispatches,
with no test that fails. That is the `a source-scan test makes your COMMENTS load-bearing` family,
one rung worse because the coupling crosses services. The sound fix is a **structured refusal
signal** — a distinct terminal status (`refused`) or a `result` key the handler sets deliberately —
so the gate reads an assertion rather than a sentence. That is a new shared vocabulary on a shared
seam, i.e. **architecture-scope** (owner rulings 2026-07-28 / 07-29), and belongs in the RFC track
with `bugs_open/295`, not in a fourth revision of this task's `pre_query`.

### Migration `471` applied — the floor-held remedy no longer misdirects (text-only)

Fixes the dated defect above. `466`(b)'s floor-held payload ended *"FIX THE HANDLER, or decide the
pair is wrong and retire the producer"*; it now leads with **"FIRST PARTITION THE FAILURES"**, carries
the fleet-wide numbers, hands the reader the partition query for their own pair, and says that if
protective refusals dominate then the handler is behaving correctly and the defect is
`bugs_open/295`, not the handler.

**Built as a single `replace()` against the LIVE `pre_query`, not as a new pasted query.** That is
the design decision, not an implementation detail: "text-only" then holds **by construction**, so I
did not have to write a control asserting it — and any control I *could* have written for it
(`replace(old,OLD,'') = replace(new,NEW,'')`) is **true by construction too**, which would have been
this lane's FOURTH tautological control. The way out of that trap was to make the property structural
rather than assert it.

Controls, both of which could have come out otherwise:
- **PRECONDITION** — the `466` anchor must occur **exactly once** in the live text (measured: 1). If
  another session had revised this task, the count changes and the migration stops instead of editing
  text it was not written against.
- **CONTROL 1, the one that matters** — `EXECUTE 'EXPLAIN ' || new_q`. **EXPLAIN plans without
  executing**, so it validates the whole rewritten statement — crucially that every apostrophe in the
  new prose is doubled, which is the realistic way to break a string edit nested inside a SQL literal
  — while mutating nothing. An un-doubled quote is a syntax error and aborts the COMMIT. (Running the
  query itself would have been the other option and was rejected: it is a `UPDATE`-in-CTE statement,
  so proving it parses by *running* it would escalate rows early.)
- **CONTROL 2, POSITIVE** — a floor-held pair with rows waiting must exist, or the corrected string is
  unreachable and the change untested. Result: **10 rows**, `literal_markdown → page-build-handler`.

Post-apply verification: `FIRST PARTITION THE FAILURES` present; both canary strings, the
`bugs_open/300` warning, the `HAVING` suppression and the archive UNION all intact; length 6044 →
7205. ⚠ **my first verification query reported `canary text intact: false` and that was MY OWN
misquote** — I probed for `run one by hand` when the live text says `runs one by hand`. A
substring probe returning false is evidence about the probe until you have checked the probe. Ledger:
recorded via `--record-only` (see RUNBOOK — `--apply` would have taken four other lanes' pending files).

**What 471 does NOT do, deliberately:** it does not change the gate's arithmetic. 0 pairs flip under
the full promoter predicate, and encoding `error ILIKE '%rebuild_policy=owned%'` into a live gate
would make an error message's wording load-bearing across services. The sound fix is a structured
refusal signal, which is architecture-scope and belongs with `bugs_open/295`.

### 2026-08-18, later still — path step 1 shipped: migration `479`, the escalation door opens both ways

Owner decision 2 ("fix the door"). Applied + ledger-recorded + committed `f95504674`, rollback file
alongside. Three things about HOW it was built are worth carrying forward more than the change:

**Surgical anchored replacement, not a rewrite.** The task body is **7,803 characters**, most of it
`466`'s `what_to_do` prose, and that lane shipped `465`, `466`, `471`, `472` in a single day. A
wholesale `SET pre_query = $Q$…$Q$` would have silently reverted whatever landed between my read and
my apply, and transcribing 7.8 KB by hand introduces its own errors. So: three verbatim anchors,
each **asserted to occur exactly once** before any replacement runs, plus a pre-image md5 guard that
**aborts** rather than clobbers. The verify block additionally asserts `466`'s parts survived
(archive scope, `hold_kind`, `resolution_path`, the failure-partition remedy, the 3-day limit) — a
negative control on a copy-paste error, since a rewrite that dropped one of those would still be
valid SQL and would still "work".

**The controls test the predicate, not the row count, and that was a deliberate choice.** Zero rows
are escalated today (oldest held row is at day 2 of 3), so `reclaimed = 0` and an assertion on that
number would pass identically if the predicate were `WHERE false` — vacuous, the exact defect this
lane keeps catching. Instead: positive control = `page_component_status_drift →
component-template-fixer` (hand-canaried 08-17, 4 ok / 0 failed) **must** satisfy the reclaim test;
negative control = `literal_markdown → page-build-handler` (3 ok / 36 failed = 8%) **must not** —
that being the pair `444` exists for, and the one that would wrongly return if the archive scope or
the floor arithmetic were wrong. Both behaved.

**Exercised end to end before applying.** The migration ran inside a transaction, the *resulting*
`pre_query` was then executed via `\gexec` — returning `escalated 0, reclaimed 0, watching 15` with
`watching_detail` naming each pair and its day count — and the whole thing rolled back. Testing the
migration is not the same as testing the query it produces; this catches a replacement that lands
cleanly and yields SQL that will not run.

**Residual, stated:** the arm has never reclaimed a real row, because none has been escalated yet.
`placeholder_contact` crosses the limit 2026-08-19 and `literal_markdown` on 08-20 — but both pairs
are *correctly* held, so neither will be reclaimed. The first genuine reclaim needs a pair to be
escalated and *then* qualify. Watch the daily tick's new `reclaimed` / `reclaimed_pairs` columns.

---

## 2026-08-18 evening (cont.) — path step 2: Tier 1, the refusal status. Commit `6aee22b00`, migration `480`, council corr `725b1f01`

### What was actually wrong, and where the fix had to go

The owner's decision was one word: an owned-page refusal must write something other than `failed`.
Finding the *place* to write it took most of the work.

`page-build-handler`'s workflow routes `save_sections`' error to `mark_item_failed`, which is
`update_work_item_status` with a hard-coded `"status": "failed"` [MEASURED — read from the live
`agent_definitions` row 2026-08-18]. Three shapes were considered and two were rejected on evidence:

1. **A different `error_step` for the refusal.** Impossible. `error_step` is a static config value
   resolved by the coordinator (`coordinator.go:3640`, `routeToErrorStepOrFail`); an action cannot
   name its own. Checked, not assumed.
2. **The guard writes the status itself, then returns the error.** Broken by construction:
   `mark_item_failed` still runs afterwards and `UpdateWorkItemStatusAction` has no terminal-state
   guard — it would overwrite `wont_fix` back to `failed` one step later.
3. **Discriminate at the routed step.** The only channel that survives the action → coordinator →
   error_step boundary is `collected_data.__step_error.message`, which `routeToErrorStep` copies
   **verbatim** from the action's error (`coordinator.go:3672`). A typed Go error does not cross it.
   So: put a marker in the message, read it at the routed step.

That is why the change is in two files rather than one, and why the marker is a string. It is not
the shape I would pick from scratch; it is the only one the seam supports today. Named as risk 1 in
the council submission rather than left for a reviewer to notice.

### The `wont_fix` choice, verified rather than inherited

`HANDOFF_2026-08-18c` §2b asserted two things about `wont_fix`. Both re-checked first-hand:

* **The floor.** [MEASURED 2026-08-18, from `scheduled_tasks.detected-item-promoter.pre_query`]
  `count(*) FILTER (WHERE h.status IN ('complete','verified')) AS c` and
  `count(*) FILTER (WHERE h.status = 'failed') AS f`, tested as `(c+f) < 5 OR c >= 0.25*(c+f)`.
  **CONFIRMED** — `wont_fix` is in neither bucket, so a refusal leaves numerator and denominator
  alone and the pair reads *never tested here*.
* **The dedup index.** [MEASURED — `pg_indexes.indexdef` for `idx_swi_dedup`] the partial index
  excludes `'complete','verified','rejected','wont_fix','failed','unresolved','cancelled'`.
  **PARTLY WRONG AS STATED, and it was my own claim.** 08-18c says *"`idx_swi_dedup` excludes
  `wont_fix`, so the dedup key is released"* — true, but it excludes `failed` too, so the key was
  **already** released and this is **not a difference between the two statuses**. It is not an
  argument for `wont_fix`; the floor is the whole argument. Recorded here rather than quietly
  dropped, because a true sentence that reads as a reason and is not one is the harder error to
  catch.

### Blast radius, measured before writing anything

The rule this lane keeps re-learning is "no collision is possible" is a query, not an argument. So,
every consumer that reads `wont_fix` *positively* (not as an exclusion):

| consumer | failed | wont_fix | differs? |
|---|---|---|---|
| `silentCoverageClause` (diagnose_silent_check) | covered | covered | no |
| `crossLinkFailedStatuses` | listed | listed | no |
| `check_page_canonical_collision` suppression | — | suppresses | scoped to its OWN item_type |
| `workItemClosedStatuses` (retraction) | open | **closed** | **YES** |
| any `scheduled_tasks.pre_query` | — | — | **none mentions `wont_fix` at all** |

One difference, and it is the right way round: a `wont_fix` refusal row will never be retracted.
That is correct for a row that is already closed, and the finding re-raises on its own because the
dedup key is free.

### Consumers enumerated, not asserted (owner ruling 2026-07-29 §3)

Every live agent with a `save_page_sections` step, queried from `agent_definitions`:

| agent | step | error routes to | affected? |
|---|---|---|---|
| `page-build-handler` | `save_sections` | `mark_item_failed` (`update_work_item_status`) | **yes — the one opt-in** |
| `page-rerender` | `save_sections` | *(no `error_step` at all)* | no |
| `tool-recreation-handler` | `save_sections` | `complete_error` → `complete_workflow` | no (writes no item status) |

⚠ **Loose end I did not chase.** `page-rerender` is 3754 ok / 89 failed on owned pages, which sits
oddly beside a guard that refuses generic saves on exactly those pages. Either rerender's saves
resolve the page differently, or the guard's `pageIsOwnedForGuard` lookup misses on that route.
[UNMEASURED] — not this task, but worth a look before anyone concludes the guard covers every save.

### Mutation-proving, because the tests were the point

Four assertions, and only as a **set**: the downgrade alone would pass on an implementation that
marks *every* failure `wont_fix`, which is strictly worse than the bug (the floor would go blind to
real incompetence). Proven by mutating the shipped code, not by reasoning:

* marker changed to one that never matches → the downgrade assertion fails (`got "failed", want
  "wont_fix"`);
* marker changed to `""` so it always matches → the default-OFF and genuine-save-failure assertions
  both fail (`got "wont_fix", want "failed"`).

Also caught by doing this: the first version of the test asserted on the `result` JSONB and **passed
vacuously**. `captureArg` only records a `driver.Value` that is a `string`, and `json.Marshal`
output arrives as `[]byte` — so every `strings.Contains` ran against `""`. Added `captureTextArg`.
The assertion looked exactly like a working one; it was the *positive* case failing loudly that
exposed it, which is luck, not method.

### Migration 480 — applied, and how it was scoped

`MIGRATIONS_DIR` pointed at a **scratch directory containing only `480`** (see RUNBOOK). The
unscoped dry run listed ~15 pending files belonging to other lanes, including two probing
*inconclusive* on drift (`467`, `468`) — `--apply` at repo scope would have taken all of them.

Exercised before applying, three ways:
1. the whole file with `COMMIT` → `ROLLBACK`: guard passed, `UPDATE 1`, verify `NOTICE` raised;
2. **probe: key pre-set** in the same transaction → the already-set guard **aborted**, as required;
3. **probe: key planted on `image-build-handler.mark_work_item_failed`** → the verify block's
   negative control **aborted** naming the leak. Without that control the positive assertion would
   pass identically on an `UPDATE` with no `WHERE` clause.

Ledger: `480_owned_page_refusal_is_not_a_handler_failure.sql` at `2026-08-18 20:24:01+00`.

### Order — why config went first here, contrary to the usual rule

The estate's rule is image first, because config naming an unregistered **action** fails at runtime.
That does not apply: this adds a config **key** to an action that already exists.
`update_work_item_status` reads its config by explicit lookup, has no `RegisterActionInputSpec`, and
there is no strict/unknown-key validator anywhere in the orchestration path [MEASURED — grepped for
one]. So a binary predating the Go half never looks the key up. **The config is live and the
behaviour is inert until the next chassis roll**, at which point it activates by itself with no
second visit. The pre-commit hook's architecture signal flagged exactly this ("needs a staged
rollout order") and the answer is in the migration header.

### Still open after this

Nothing about Tier 1 except the roll and the verdict (`725b1f01-f4b5-42fc-92b5-6de8fc0daa85`,
`Council-Submitted:` on `6aee22b00` — **not** `Council-Reviewed:`, the verdict is unread).
**Post-roll verification needs a positive AND a negative control**: an owned-page refusal must land
`wont_fix` carrying `result->'owned_page_refusal'`, and a genuine save failure must still land
`failed`. Without the second, "the refusals stopped counting" is equally consistent with having
broken the status write.

---

## 2026-08-18 late evening — the REVISE was right, and it found a defect I could not have found

### What the council objected to, and why my evidence was insufficient

Round 1 came back **REVISE**, gated by `editquality`. The objection, in one line: *you write the
item's status inside the handler, but the dispatch loop runs afterwards — does your write survive?*
It did not argue; it **cited two `LANDMINES.md` entries** keyed on `update_work_item_status
result_fields` and `site_work_items.result for any item completed by a loop sub-workflow`, both
saying the loop replaces what a handler wrote.

**The entries existed and I had not read them.** I grepped LANDMINES for the symbol I was *adding*
(`wont_fix`) and not for the mechanism I was *writing through* (`update_work_item_status`). Logged
in `WRONG_CALLS.md`.

### The answer — three writers, and they do not agree

| writer | guard |
|---|---|
| `CompleteWorkItemAction` (`load_work_item_actions.go:1017-1025`) | `status NOT IN ('needs_human_review','failed','unresolved','rejected','wont_fix','verified','blocked')` |
| `failUnverifiedCompletion` (`complete_work_item_verification.go:428-429`) | the **identical** list |
| `FailWorkItemAction` (`load_work_item_actions.go:1146-1160`) | **none — `WHERE id = $1`** |

`wont_fix` is in the first two lists, so on the `mark_complete` path the UPDATE matches **0 rows**
and replaces nothing — not the status, and not the result either, because both are in the same
statement. The landmine is real and does not reach this case. On the `mark_failed` path there is no
guard at all and the status is overwritten to `triaged` or `failed`.

**So which path does a refusal take?** [MEASURED 2026-08-18] over every `page-build-handler` item
whose `error` names the guard, split by `handled_by` — a column only `fail_work_item` and
`complete_work_item` ever write, so NULL means the handler's own write was never touched:

| `handled_by` | status | rows |
|---|---|---|
| NULL — handler's own write, untouched | `failed` | **113** |
| NULL | `cancelled` | 3 |
| `build-dispatch-loop` — via `fail_work_item` | `failed` | **2** |

**113 of 115 = 98.3% guarded; 2 not.** Positive control, because the first number alone reads as
"nothing ever overwrites anything": **122 rows sit at `needs_human_review` with
`handled_by='page-build-handler'`**, most recently today — a handler-set protected status
demonstrably surviving this exact loop, 122 times.

### The thing worth carrying, which is not "read LANDMINES"

**A blast-radius census of who READS a value is half the question. The other half is who else
WRITES it, and last.** I enumerated readers exhaustively and writers not at all, and the
exhaustive-looking table is precisely what made the omission invisible — to me and nearly to the
reviewer.

And: **today's rows could not have caught this either.** Both paths write `failed`, so no query on
the existing population distinguishes "the handler said failed and it stood" from "the loop
overwrote it". The 2 rows are visible *only* because `handled_by` separates the writers. A defect
that every possible query returns the same answer for is not absent — it is unobservable, and the
change that introduces a second possible value is what makes it visible.

### What I did with it

Contributed to **`bugs_open/307`** (owned by `staged_component_build`, active, same function) rather
than fixing it: `fail_work_item` is reached by every dispatch loop, `failed` is itself in the
sibling guard list so a naive copy would stop the retry ladder, and a shared-mechanism change inside
a bug patch is what the 2026-07-28 ruling vetoes. The split it probably wants is in the
contribution: statuses that record a DECISION protected, `failed`/`unresolved` left overwritable.
Round 2 resubmitted on the same correlation with the measurement and the residual stated.

---

## 2026-08-18 late evening — path step 3: `bugs_open/300`, and one figure in that bug had already moved

### The measurement is stronger than the bug file's, in the direction that matters

[MEASURED 2026-08-18, all 82 lifetime `page_component_status_drift` rows] `page_id` present 82/82,
`spec.slot_name` present 82/82, `spec.page_component_id` resolves **70/82** (12 dead, 15%),
`(page_id, slot_name)` resolves **82/82**. The bug filed it as "1 of 20".

**And the ageing it predicted has happened in one day.** The file recorded on 08-17 that *"every one
of their ids still resolves today"* for the 16 deferred rows. Today **11 of 16** do. Five died in a
queue nobody touched. That is the file's own hypothesis confirming itself, not my finding.

### The part that needed a test rather than a query

`(page_id, slot_name)` is **not unique**: [MEASURED] 17 such pairs on the estate carry more than one
component, worst case 4, of 1,730 components. **Zero of them are drift rows today** — which is
exactly the trap. Resolving by the pair alone is correct on every row that exists now and silently
wrong on the first one that is not, and no query against current data would ever say so. Hence the
stored id as a tiebreak *within* the pair's matches, and a refusal rather than a guess when the pair
is ambiguous and the stored id is not among them.

### A correction to this lane's own handoff

`HANDOFF_2026-08-18c` §5 says: *"⚠ `page_id` is **not** in the dispatch loop's `call_handler`
input_mapping (verified live)"*. **That is wrong.** [VERIFIED 2026-08-18, `build-dispatch-loop`'s
live `sub_workflow`] the mapping is `{spec, domain, issue?, source, site_id, page_id?, purpose?,
asset_id?, item_type, page_name?, current_page, work_item_id, component_id?, reviewed_brief?}`.
`page_id?` is there; the `?` makes it optional, so it is dropped whenever the row's `page_id` is
NULL — which is why the live `component-template-fixer` payload I sampled (an
`instance_scope_conversion` item) had none, and probably how the original claim was formed. The fix
does not depend on it: it joins through `work_item_id`, which is unconditionally mapped.

### Mutation matrix

Stable key never consulted → 4 tests RED. Tiebreak removed so an ambiguous pair takes the first
match → 2 RED. Fallback to the stored id removed → 2 RED. Full `platform/` suite green after
restore, and the working file byte-identical (`git diff --numstat` unchanged).

---

## 2026-08-18 late evening — path step 5: `bugs_open/314` filed, and the prior art nearly stopped me

Owner decision 6. The handoff's framing was *"097 scopes to platform//internal//pkg/, so config
cannot be submitted"*. Reading the script sharpened it: `SCOPE_RE='^(platform|internal|pkg)/'` is a
**path** test standing in for a **subject-matter** rule ("prose does not spend council credits"), and
this estate's config ships as SQL under `docs/agent_docs/sql_for_agents/`. **The gate is not
declining to review config; it cannot see that it is config.**

[MEASURED, 14 days] 227 commits ship a numbered migration; **152 (67%) carry no in-scope file** and
are refused by construction. Stated as a bound, not an enumeration — some of those are docs commits
carrying a migration, and some lanes used `FORCE=1` anyway.

⚠ **The prior art nearly stopped me filing, and reading it properly is what let me file.**
`architecture_review/DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` §8d cites this exact
line and concludes the refusal is **correct**. It is — *for prose*, which is what §8d was arguing
about (72 DESIGN/PLAN/SPEC docs in a month). Finding prior art that endorses the rule felt like
confirmation that there was no bug. **It was not: the question is whether the prior art was arguing
about your case.** That went into 016b §9 as the transferable half.

---

## 2026-08-18, end of session — APPROVED at round 3, and the approval's advisories corrected me twice more

**Verdict: `approved`, corr `725b1f01-f4b5-42fc-92b5-6de8fc0daa85`, 1 advisory none high, 4
abstained, `gated_by_truncation: false`.** Three rounds; **both REVISEs found real defects**, which
is now this lane's second data point that a revise round is cheaper than what it catches.

### The advisories that were checkable, and what checking them cost me

**`prior_art_librarian`, medium — "the numbers sizing Tier 2 are on your own `[MEASURED]` tag, not
quoted as query results."** Re-ran them. **Two were wrong.**

```sql
SELECT owner_agent_type, count(*), min(created_at)::date, max(created_at)::date
FROM orchestration_states WHERE owner_agent_type IN ('copy-editor','section-editor') GROUP BY 1;
--  section-editor | 18 | 2026-08-17 | 2026-08-18
--  copy-editor    |  2 | 2026-08-18 | 2026-08-18
```

1. **I compared orchestrations to work items** — "2 runs vs section-editor's 227". Like for like it
   is 2 vs 18 orchestrations (a ~24h retention window, not a lifetime) and **0 vs 227 work items**.
   The honest headline is not "barely used", it is **"nothing dispatches it at all"**.
2. **`copy-editor` is ONE DAY OLD, not dormant.** Seeded `2026-08-17 11:49`, updated
   `2026-08-18 17:59`, both runs today, **owned by the `loanandmortgagecalculator_couk` lane**
   (migrations `447`/`462`; `b04493b7b` — *"stage 2 BUILT and PROVEN on its proof case"*).

> ⚠ **CORRECTED 2026-08-19 — `copy-editor` is owned by the `copy_quality_two_stage` lane, NOT `loanandmortgagecalculator_couk`.** I got the wrong lane from a `grep -rl "copy-editor"` hit in LMC's `README_where_we_are.md` — a *mention* — and read it as ownership. `scripts/who-owns.py` exists to separate those two, and I did not run it. The defining evidence is what the commits shipping migrations `447`/`462` actually touch: `docs024_key_docs_latest/copy_quality_two_stage/`. Register entry **CQ-024**. A CONTRIB is filed in their lane dir (`CONTRIB_2026-08-19_from_the_277_083_lane_…`, commit `7574482c7`).


**That changes the Tier 2 advice, not just the record.** "Aim an existing producer at a finding"
still holds, but the producer has an owner iterating on it *right now*, so the next step is to talk
to that lane — not to draft an RFC around a `field_updates` contract that has changed twice in two
days.

> ⚠ **The lesson, and I made this error class twice tonight.** I found a mechanism I had wrongly
> declared nonexistent — and then characterised it ("2 runs, barely exercised") from figures I had
> not lined up. **Finding the thing you missed is not the end of the correction: the first
> description of it is written in the same hurry that produced the original claim.** The seat that
> caught the absence caught the description too, one round later.

**`prior_art_librarian`, low — "'no scheduled_task mentions `wont_fix`' is an absence claim, check
it."** Fair, and my original had no control. Re-run with one:

```sql
SELECT count(*) FROM scheduled_tasks WHERE COALESCE(pre_query,'') ILIKE '%wont_fix%'
   OR COALESCE(input_data::text,'') ILIKE '%wont_fix%';   -- 0
SELECT count(*) FROM scheduled_tasks WHERE COALESCE(pre_query,'') ILIKE '%failed%';  -- 5
```
The 5 is what makes the 0 mean something — the query can come out non-zero.

### Two residuals recorded rather than closed

- **`bug_historian`, low, and it is a genuinely new observation:** a `wont_fix` row is excluded from
  retraction **and** released by the dedup index, so it is never re-validated. **If a page's
  `rebuild_policy` later flips `owned → generic`, nothing revisits the closed refusal.** Harmless
  today — the finding re-raises and dispatches normally — but recorded, because a queue with no
  re-validation is exactly what `bugs_open/083` exists about.
- **`architecture`, low — do not spread the marker trick.** Deciding a terminal status by scanning
  error text is accepted here because the seam offers no other channel. **A second call site makes
  it a shared, stringly-typed contract nobody declared** — at that point the right move is
  structured error metadata on the coordinator, not a third `strings.Contains`. Written into WII-019
  so the next author meets it before writing the second one.

### `bug_historian` medium and `guardian` low both landed on the same thing, and agreed with the call

Both flagged `FailWorkItemAction`'s missing guard — `bug_historian` as *"one call site of a shared
judgement gets the rigorous fix; the sibling stays heuristic"* (a named 016b §9 shape), `guardian` as
*"correctly NOT fixed inline; a shared action getting a guard change belongs in its own reviewed
patch per the 2026-07-28 ruling"*. **Two seats, opposite emphases, same conclusion: right defect,
right containment.** It is in `bugs_open/307` with the split it probably wants.

## 2026-08-19 ~09:00Z — fresh chassis verified; and TWO other sessions landed on this lane's open question overnight

### Build: `agent-chassis` **`d3590ca46`**, verified at the BINARY with both controls

The startup `build provenance` line was **absent from `--tail=20000` on both pods** — the documented
"scrolled" case, not "unstamped". Probed `/proc/1/exe` instead, with a control in the same breath:

| sha | expectation | result |
|---|---|---|
| `d3590ca4638d49bb6a3874db681814c4b0a99bbe` | present if this is the fleet release | **PRESENT** |
| `db6ae7254` (current HEAD, built after) | must be ABSENT | absent |
| `0b185bad2a49…` (yesterday's build) | superseded, must be ABSENT | absent |

So a **real** build, not a same-tag cached no-op: **158 commits** on from yesterday's `0b185bad2`,
HEAD 3 further on. `kafka-scheduler` reports the same sha from its own log line.

### ⚠ `480` — another session BUILT the structured refusal signal I deferred, with an OWNER DECISION

**This supersedes my 2026-08-18 "left open deliberately" paragraph, and the next session must not
read that as still-open work.** `480_owned_page_refusal_is_not_a_handler_failure.sql` (applied
2026-08-18 20:24, Go half in this roll — `save_page_sections_action.go` + `v3_site_actions.go`)
makes `page-build-handler` record an ownership refusal as **`wont_fix`, not `failed`**. Owner
decision 2026-08-18 №1: *"do not switch the handler off for this — write something other than
`failed`."* `wont_fix` is in **neither** bucket of the floor, so a refusal now leaves numerator and
denominator alone. Shipped as an **opt-in field with the unsafe default OFF**
(`owned_page_refusal_status`), per the 2026-08-02 §2 ruling — which is exactly the shape I said the
sound fix would need.

Its author reached my conclusion independently and from a different case
(`phantom_internal_link → page-build-handler`: 101/46 = 69% on generic pages, 0/14 on owned ones,
blended 47%, with ~134 findings queued behind it). **Two lanes converging on the same defect from
different evidence is the signal that it was real.**

**What it does NOT do — and this is why `471`/`472` still earn their place:** it is **forward-only**
and backfills nothing. `literal_markdown → page-build-handler` re-measured today is **unchanged: 3
successes / 36 failures, 16 protective / 16 genuine, 8% raw and 16% corrected** — still correctly
held, and still needing **9** more successes rather than 3 to climb out. The ratchet I described is
not undone for existing rows; it only stops tightening. So when the pair escalates on **08-21**, the
human still needs the partition instruction.

**[UNPROVEN — no demand yet]** `480` has **not been exercised**. Zero `wont_fix` owned-refusal rows
since it applied, **and zero owned refusals written as `failed` since the roll either** — the second
half is the control that makes this "nothing tried", not "the fix failed". Do not record `480` as
behaviourally proven on the strength of an absence.

### `479` — the one-way door is fixed, and my `471`/`472` text SURVIVED

`479_escalation_reclaims_a_pair_that_has_since_qualified.sql` (owner decision №2, *"fix the door"*)
closes `453`'s one-way door: an escalated row whose pair later qualifies now rejoins the automated
path. Its author explicitly worked around this lane — **surgical replacement on three verbatim
anchors, each asserted to occur exactly once, guarded on the whole body's md5**, precisely because
"that lane is iterating fast (465, 466, 471, 472 in one day)". Verified after the fact: `FIRST
PARTITION THE FAILURES` present, `fix candidate 3 is UNTOUCHED` present, `bugs_open/295` absent.
Body 7,205 → 10,566 chars. **That is how to edit a live object another session is working on.**

### Escalation prediction RE-MEASURED on 08-19 and UNCHANGED

09:02Z, `held-pair-canary-escalation` last fired **08-18 12:57:48** — so today's tick had not yet
run and yesterday's correction is **still pending, not yet graded**. Held set identical to
yesterday: `placeholder_contact` **08-20 12:57**; `dead_fragment_link`, `literal_markdown`,
`missing_conversion_path` **08-21 12:57**.

### Misstep, caught inside the session: I measured promotions against a literal I INVENTED

Asked "is the promoter still promoting?" with
`WHERE resolution_path = 'auto:promoted_known_good'` → **0 rows for 4 days**, which reads as a dead
promoter. The promoter **sets no `resolution_path` at all**: it writes `status='triaged'`,
`triaged_at=now()` and `spec.original_pipeline`. A census of the column confirmed my literal exists
nowhere in 31k rows. Re-measured on the real signature: **301 / 100 / 71 / 335 / 316 promotions**
on 08-14…08-18, outcomes **946 complete vs 79 failed (92%)** — criterion 1 holds comfortably.
**The zero was well-formed, instant, and about a column the mechanism never writes.** Same family as
the five already logged; the check that caught it was censusing the column instead of trusting the
predicate. Read the mechanism's own `SET` clause before keying a health check on a value.

---

## 2026-08-19 — the roll landed. Both changes are LIVE and PROVEN at the binary; and verifying them corrected two of my own numbers, one of them load-bearing

### The deploy, proven at the artefact on both replicas

`agent-chassis:v1.0.1314`, pods `-l5h6l` (07:52Z) and `-nxmkf` (08:05Z). The `build provenance`
startup line had already scrolled out of `--tail=3000` — **which means "not in range", not
"unstamped"** — so the binary probe, single pass, four needles:

```
grep -aoE 'owned_page_refusal_status|resolveStatusRepairComponent|OWNED_PAGE_GUARD|ZZQQ_NEEDLE_THAT_MUST_NOT_EXIST' /proc/1/exe
  -> owned_page_refusal_status        PRESENT   (Tier 1's Go half)
  -> resolveStatusRepairComponent     PRESENT   (300's Go half)
  -> OWNED_PAGE_GUARD                 PRESENT   (long-lived control: the probe works)
  -> ZZQQ_NEEDLE_THAT_MUST_NOT_EXIST  ABSENT    (negative control: the probe discriminates)
```
Both replicas, same result. Config half intact (`owned_page_refusal_status: wont_fix` still on
`page-build-handler.mark_item_failed`).

### Behaviourally UNVERIFIED, and the demand control is why I am not claiming otherwise

Zero owned-page refusals since 07:52Z — **and zero `page-build-handler` orchestrations either.** So
the zero is explained by nothing having been dispatched, not by the fix working. Same for `300`:
`page_component_status_drift` has had no dispatch since 08-18. **A post-fix zero with no demand is
not evidence**, and refusals occur naturally at roughly 4/hour on live traffic (`bugs_open/301`
measured 59 in 14 hours), so the right move is to check again rather than to induce one and burn an
LLM chain — which is the exact waste `301` exists about.

### ⚠ CORRECTION 1 — "47% overall" for `phantom_internal_link` is WRONG. It is 62.7%.

I have repeated 47% in the council submission, register **WII-019**, both handoffs and my
`bugs_open/301` contribution. Measured today over live+archive, terminal outcomes only:

| `phantom_internal_link → page-build-handler` | ok | failed | % |
|---|---|---|---|
| on `generic` pages | 101 | 46 | **68.7%** |
| on `owned` pages | 0 | 14 | 0% |
| **TOTAL** | **101** | **60** | **62.7%** |

101/161 = 62.7%. **47% was arithmetic I got wrong and then carried**, and the two component figures
I quoted alongside it (69% and 0/14) were right the whole time — anyone could have divided them.

**Why it matters rather than being a typo:** the floor is 25%. From 101/60, crossing it needs
**243 more failures**. So *"a 69%-effective repair path is one bad stretch from switching off"* was
**overstated**. The change is still correct — a refusal is not incompetence — but the urgency was not.

### ⚠ CORRECTION 2 — and this one refutes a remedy I was about to propose

Having found that Tier 1 only affects FUTURE refusals, I went looking for the pairs it would release
if the historical rows were re-classified, and read `placeholder_contact` as held by 4 ownership
refusals. **It is not.** Its owned-page failures read:

```
step process_sections_loop_iter_0_generate_content failed: ...
```

— the **content generator** failing, not the guard refusing. I had inferred "owned page + failed =
ownership refusal" **from the page's policy column**, when the only thing that says so is the
guard's own error string. Discriminated properly (`error LIKE '%rebuild_policy=owned%'`), of 87
`owned`+`failed` rows **85 name the guard and 2 do not — and those 2 are exactly these.**

### So what Tier 1 actually buys, stated honestly

**It releases nothing that is held today**, and would not have even applied retroactively:

| held pair | why it is held | does Tier 1 touch it? |
|---|---|---|
| `literal_markdown` | 3 ok / 16 real failures = still below the floor with refusals excluded | **no** — real defect, `bugs_open/184` |
| `placeholder_contact` | never completed one; its owned failures are generator errors | **no** |
| `dead_fragment_link` | never completed one — awaiting a hand canary | no |
| `missing_conversion_path` | never completed one — and `bugs_open/255` says its handler cannot read its spec | no |

**Its value is PREVENTIVE and it is real:** 85 identified ownership refusals already sit in the
`failed` bucket in the live table alone, and ~134 findings are queued behind the refusal on owned
pages — every one of which would otherwise add to a denominator it has nothing to do with. But
"protects `phantom_internal_link`'s path immediately" is wrong on both words: not immediately, and
that pair was never close to the floor.

> **The transferable bit, and it is the same shape as last night's two:** I discriminated a category
> by the attribute that was *convenient to query* (`pages.rebuild_policy`) rather than the one that
> actually *defines* it (the guard's error text). The convenient column was 97% right, which is
> exactly why the 3% was invisible — and the 3% was the entire pair I was building a remedy for.

### ⚠ CORRECTIONS to this morning's 09:00Z entry — two figures, and I destroyed a file to find them

**I overwrote another session's `HANDOFF_2026-08-19_continue_here.md`** with a `cat >` heredoc,
having never read the path. It was written at 10:14Z (`917a5de9f`) and I clobbered it ~15 minutes
later. Recovered byte-identical (md5 `271f1df6…`) from `git show 917a5de9f:<path>` and restored
forward-only; my material is now an appendix below theirs. **This is the exact failure CLAUDE.md
names** — *read before write on any file you did not create; prefer the Write tool, which refuses an
unread file, over a shell redirect, which does not.* The tell I ignored: `git commit` printed **no
`create mode` line** and reported **137 deletions** on what I believed was a new file.

Their file corrects two things I wrote at 09:00Z above:

1. **`phantom_internal_link` is 62.7% lifetime, not 47%.** Generic **101/46 = 68.7%**, owned
   **0/14**, total **101/60 = 62.7%**. I quoted "blended 47%" **straight out of `480`'s own header
   without re-deriving it** — the same failure as the `bugs_open/295` path: repeating a figure from
   another document because it was in a document. Crossing the 25% floor from 62.7% needs **243 more
   failures**, so any "one bad stretch from being switched off" reading is overstated.
2. **"owned page + `failed`" is NOT the set "ownership refusal".** Of **87** `owned`+`failed` rows,
   **85 name the guard and 2 do not** — those 2 are `placeholder_contact`'s, failing in
   `process_sections_loop_iter_0_generate_content`, i.e. the **content generator**, not the guard.
   So discriminate on the guard's **error text**, never on `pages.rebuild_policy`. My own partition
   did key on the error text (`error ILIKE '%rebuild_policy=owned%'`), so my 434/16 figures stand —
   but the distinction is one I never stated, and stating it is what makes the number safe to reuse.

**What my figures got right, independently confirmed:** their `literal_markdown` line reads
*"3 ok / 16 REAL failures — still below floor with refusals excluded"*, which is exactly my
3/(3+16) = 16%. Two lanes, two methods, same number.

**And they are ahead of me on closure, which is the part that matters:** `277` is blocked not by the
churn-guard clock I quoted but by **its own verify clause 1 — the worked example must be REPAIRED,
and nothing repairs `no_content_data` at all** (44 completions are `auto:revalidated`, 37
`build-dispatch-loop`, **0 by the router**). `083` closes **~08-25** (owner decision 5), not 08-24.
**Their handoff is authoritative; read it, not my §3.**

---

## 2026-08-19 (cont.) — two peer lanes made contact, and one of them found a live hazard in MY machinery

### The CONTRIB got read the same hour, and my "not reachable" was wrong

I recorded in §5 of the handoff that a direct `SendMessage` to the `copy_quality_two_stage` lane
failed and the committed CONTRIB was therefore the only channel. **That was half wrong.** The
session IS reachable — its peer name is **derived** (`agentchassis-8d`), not lane-shaped, so
searching `ListAgents` for "copy quality two stage" misses it. The mapping is in
`~/.claude/sessions/*.json` (`name` + `sessionId` → `af212352`), and ⚠ **`ListAgents`' `[ref]` is
NOT the session-id prefix**, which is exactly what made my obvious lookup fail. Told to me by
`agentchassis-22`, verified by me before acting on it.

It resolved in the best way: **they reached me independently within the hour, and their reply cites
the CONTRIB.** So the durable file was still the right thing to file — but "not reachable" should
have been "I could not find the name", which is a different claim.

### The hazard they found, which is in the promoter and therefore mine

They asked a careful question: does the promoter need `copy_edit_proposed` recorded as
**never-promotable**, or is it safe merely by not being a known-good pair? Their words:
*"I would rather it be excluded on purpose than inert by accident, because inert by accident changes
the day someone canaries it without knowing the D2 constraint."*

**Measured, not reasoned. It is safe today, by TWO independent barriers:**

1. `checkpoint_for_review` files at `needs_human_review` (`checkpoint_for_review_action.go:223`) and
   the promoter's `scored` CTE takes `WHERE wi.status = 'detected'` only — so the rows are never
   looked at. Both live `copy_edit_proposed` rows are `complete`.
2. Even at `detected`, `handler_ok` would fail: `human-review` is **not** a live `agent_definitions`
   row (0).

**But their instinct was right, and the second barrier is the rotten one.** A held row's reason
string is *"handler not a live agent"* — indistinguishable from a broken routing config, and an
invitation to "fix" it. **And the next step is mine:** `held-pair-canary-escalation` escalates a
held row after 3 days **asking a human to hand-canary the pair**. So my own task would surface a
type whose entire design is "never auto-dispatch" to a person as *awaiting a canary*.

**The estate already has the right mechanism and they had already cited it without connecting it:**
`voice_tells` files `handler_agent = ''` (43 rows), and `scored` excludes empty handlers **outright**
— the pre_query's own comment says those rows belong at `detected` permanently and *"holding is not
what is happening to them"*. So:

| | `handler_agent` | what the promoter does | how it reads |
|---|---|---|---|
| `voice_tells` | `''` | excluded by construction | deliberate, and documented |
| `copy_edit_proposed` | `'human-review'` | selected, then HELD | **like a defect to fix** |

I recommended they move the label into `spec` and empty `handler_agent` — it makes the bad state
unrepresentable rather than merely refused, and needs nothing from my roster. I explicitly declined
to add an exclusion list to the promoter: a second roster to maintain is the drift class this estate
keeps filing bugs about. Offered as their call, with the alternative named.

`LANDMINES.md` entry added jointly, D2 cited at the point of enforcement, verifier dispatched.

> **The bit worth carrying:** the guarantee was never in danger, and the question was still worth
> asking. *"Safe by accident is indistinguishable from safe by design until the day someone acts on
> the hold reason"* is a better test than "is it currently dispatching?" — and it took an outside
> lane to ask it, because from inside my own mechanism the hold looked like the system working.

### Work handed to `agentchassis-22`, with the split agreed

They take handoff §4.5 — the two `[UNMEASURED]` loose ends nobody owns: (a) `page-rerender` saving
to owned pages while `page-build-handler` is refused, same guard; (b) `tool-…` pages carrying
`rebuild_policy='generic'`. Both read-only. I kept §4.1–§4.4 (post-roll checks, `083`'s close, the
`277` conversation I had just opened, and `314` which is the owner's call).

**Three things I gave them that they would otherwise have re-derived:**

- ⚠ **the 3,754 figure is not mine.** I inherited it from another session's table in `bugs_open/301`
  and repeated it in my handoff with no marker. **Re-derive before building on it.**
- ⚠ **the discriminator trap** — classify by the guard's error text, never by `pages.rebuild_policy`.
  85 of 87 vs 2 of 87; the 2 were the whole population I nearly built a remedy around.
- **`page-rerender`'s `save_sections` has no `error_step` at all**, so a refusal there fails the
  workflow rather than routing to a status write — a candidate explanation for refusals being
  *invisible* on that route rather than *absent*. All three agents call the same action, so the
  guard predicate is identical by construction; the difference is in what reaches it or what happens
  after.

**And the one thing in my scrollback that had never reached a file**, which is exactly the
duplication `who-owns.py` cannot see: today I measured **0 `page-build-handler` orchestrations since
the roll** while **20 of its work items were updated at 08:45:58** — a handler apparently acting with
no orchestration to show for it. I noticed, did not chase it, did not record it. Same shape as their
(a). Handed over with my zero explicitly marked unverified.

---

## 2026-08-19 — §4.5(a) and (b) MEASURED, by the `agentchassis-22` session (split agreed with the `bugfix 083` thread before starting)

Handed the 08-19 handoff by the owner and told to coordinate rather than duplicate. The lane thread
approved this split: §4.1–4.4 stay theirs, I take only the two `[UNMEASURED]` loose ends. Everything
below is mine; nothing above this line was edited.

### (a) — THE ASYMMETRY DOES NOT EXIST AS STATED. Both handlers are refused by the same guard

The handoff's §4.5 asks why `page-rerender` "saves to owned pages ~3,754 times without refusal while
`page-build-handler` is refused every time. Same guard." **Both halves of that sentence are wrong,
and so is the premise that they differ.**

**1. The 3,754 REPRODUCES but does not mean what it says.** [MEASURED 2026-08-19, live+archive]
Re-derived from scratch rather than inherited: `page-rerender` on `rebuild_policy='owned'` pages is
**3,818 complete+verified / 89 failed** (`bugs_open/301` line 196 said 3754/89 — same failures, ok
grown by 64 in a day, so the figure is sound). **But it counts WORK-ITEM OUTCOMES, not saves.** A
completed rerender item is not a save that reached the guard.

**2. ~89% of those items never call the guarded action at all.** [MEASURED] `page-rerender`'s live
workflow gates `save_sections` behind `check_rerender_mode`, whose condition is
`spec.reason IN (image_landed, section_data_resolved, cta_links_stale, template_changed)`; the
else-branch goes straight to `render_page` (assemble stored HTML), which never calls
`save_page_sections`. Distribution of the 4,171 owned-page items:

| spec.reason | reaches save_page_sections | ok | failed | rows |
|---|---|---|---|---|
| `(null)` | **no** | 3542 | 6 | 3681 |
| `cta_links_stale` | yes | 120 | **82** | 332 |
| `section_data_resolved` | yes | 127 | 1 | 129 |
| `verbatim_adoption_deploy` | no | 26 | 0 | 26 |
| `legal_page_publish` | no | 3 | 0 | 3 |

So the guard is **reached ~461 times, not 3,754**. It is not being bypassed; it is not being reached.

**3. `page-rerender` IS refused — 81 times.** [MEASURED, with a control that could have come out
otherwise] Discriminating on the guard's own error text (per this lane's 08-19 rule, never on
`pages.rebuild_policy`):

| handler | policy | error names guard | other error | failed rows |
|---|---|---|---|---|
| `page-rerender` | owned | **81** | 8 | 89 |
| `page-rerender` | generic | **0** | 46 | 46 |
| `page-build-handler` | owned | **100** | 10 | 112 |
| `page-build-handler` | generic | **0** | 151 | 161 |

The two `generic` rows are the control: zero guard-named errors where the guard cannot fire, against
46 and 151 real errors. The predicate discriminates.

**4. And `page-build-handler` is NOT refused every time — it completes 74 items on owned pages.**
[MEASURED] 74 ok / 112 failed lifetime. The "every time" reading came from looking only at failures.

**What is actually true:** both handlers call the same `save_page_sections` and are refused by it at
rates proportional to how often they reach it. `page-rerender` looks unrefused only because the
denominator was ~8x too big. **The peer's warning was the right one and it applies to their own
figure:** the route was invisible in the table it was being counted in.

**The one REAL anomaly, which is sharper than the question that was asked.** [MEASURED] Post-guard
(items created ≥ 08-08), on owned pages, through the *same* `save_sections` step:

| handler / reason | completed | guard-refused |
|---|---|---|
| `page-rerender` / `section_data_resolved` | 122 | **0** |
| `page-rerender` / `cta_links_stale` | 112 | **19** |
| `component-template-fixer` | 53 | **0** |

Same agent, same step, same guard, opposite outcomes by reason. **[INFERRED — mechanism read in
code, NOT measured per row]** `rerender_page_sections_action.go:401-419`: a section with no stored
`content_data`, or missing required LLM fields, escalates to the writer and returns
`escalated=true`; `check_escalated` then routes to `complete`, **skipping `save_sections` entirely**.
So those items complete having written nothing. `isSelfContainedSection` exempts tool sections from
that escalation, which is why the two reasons could diverge. **The check that would settle it:** per
row, whether `rerender_sections.escalated` was true — not attempted here.

**Corollary worth its own line:** `page-rerender` has **no `error_step` on any of its ten steps**
(verified in the live `agent_definitions` row, not the seed). A guard refusal there fails the whole
workflow rather than routing to a status write — so its refusals are real but land differently from
`page-build-handler`'s, which is the grain of truth in the original question.

### (b) — 89 tool pages are marked `generic` while 95 IDENTICALLY-SHAPED ones are marked `owned`

[MEASURED] Cross-tab of the two shape tests, kept independent so neither alone decides:

| name `tool-%` | url `/tools/%` | is `-guide` | policy | pages |
|---|---|---|---|---|
| yes | yes | no | **owned** | 95 |
| yes | yes | no | **generic** | **89** |
| yes | no | yes | generic | 70 |
| no | yes | no | owned | 10 |
| no | yes | no | generic | 4 |

The 70 `tool-…-guide` pages are prose guides and are **correctly** generic — the name test alone
over-counts, which is why the URL and guide axes are shown separately. **The defect is the top two
rows: 89 (+4) tool artefacts marked the opposite way from their twins.** Estate total is 172 owned
of 786 pages.

Seed case from the peer, **re-verified rather than inherited**: `tool-pet-treatment-cost-estimator`
(`vetcomparison.uk`, `/tools/pet-treatment-cost-estimator/index.html`, 1 component) is `generic`.
Its sibling `tool-pet-treatment-cost-estimator-guide` is also generic and legitimately so.

### ⚠ WHERE I FAILED MY OWN CONTROL, recorded because the failure is the useful part

I tried to price (b)'s damage as "completed generic saves that the guard would have refused had the
page been marked owned" and got 107 (of 688 completed runs, after correctly excluding 581
assemble-only rerenders — the same denominator error I had just caught in the 3,754).
**That number did not survive its own control and I am withdrawing it.** Splitting by guard era,
`owned` pages show **174 completions post-guard on the route I had classified as guarded** — which
is impossible if that classification were right. Two candidate confounds, neither tested:
`rebuild_policy` is **mutable and read at query time**, so a historical item is being judged against
today's marking (this lane's landmine 3, "the row set is not stable", in a new dress); and my
per-agent route classification is an assumption about workflow reachability, not a measurement —
`component-template-fixer` is 150/0 on owned pages, which alone shows it does not reach the guard.

**So: (b)'s POPULATION is established; (b)'s DAMAGE is NOT.** No exposure figure should be quoted
from this section. The clean way to price it is per-row orchestration evidence that
`save_page_sections` actually ran, not a join against today's policy column.

**This also gives a second, independent reason for the lane's 08-19 error-text rule:** the policy
column is wrong for historical rows not only because "owned+failed ≠ refusal", but because the
column can have changed since the run. The error text is what the run itself recorded.

### One stale comment found in passing, worth correcting where it lives

`save_page_sections_action.go:203` states *"Measured 2026-08-17: 0 `owned_page_review` rows have
ever carried `refused_by='save_page_sections'`"*. [MEASURED 2026-08-19] That is now **64 rows,
2026-08-17 → 08-18** — the `bugs_open/295` emit is live and filing. The comment is a dated
measurement so it is not wrong as written, but it reads as a live claim. Left for the code's owner.

### Not done, and deliberately left

- The peer's unrecorded observation (0 `page-build-handler` orchestrations vs 20 work items updated
  at 08:45:58) is the same "invisible in the table you would count it in" shape. **Not chased.**
- `feasibility-recheck`'s `pre_query` touching `handler_agent` + `agent_definitions` — not looked at.
- Whether (b) warrants a `/bugs_open/` file. Grepped both dirs first: `bugs_open/208`, `266`, `301`
  and `bugs_closed/295` are the neighbours; none of them is this finding. **Recommend filing, but
  the damage half must be priced first** — a bug filed on the population alone would assert a
  cross-cutting cause this measurement does not support (CLAUDE.md's 2026-07-31 ruling).

---

## 2026-08-19 later — §4.5(b) DAMAGE PRICED AT THE ARTEFACT: it is a ZERO, and the zero is the finding (`agentchassis-22`)

Following the lane thread's route 2 — *price the consequence, not the cause* — because it is immune
to the mutability trap that broke my first attempt: it asks the served page, not the label.

**VERDICT: population confirmed, damage NOT FOUND. No tool page has been clobbered. On this
evidence (b) should NOT be filed as a bug.**

### How the candidate set was narrowed, and why the aggregate lied twice

The clobber signature is *interactive markup replaced by prose* — that is precisely what the guard
exists to prevent. So a page still carrying its markup is not clobbered, by construction.

| | pages | still carry `<script>` in stored components | avg components |
|---|---|---|---|
| `generic`-marked tool pages | 89 | 72 | 3.2 |
| `owned`-marked tool pages | 95 | 94 | 1.3 |

That table looks damning — 81% vs 99%, and 2.5x the components, exactly the shape of one tool
component being replaced by several prose sections. **It is not damning, and reading the 17 rows
instead of the aggregate is what showed it.** [MEASURED]

Of the 17 `generic` tool pages with no script left: **7 are `archived`**, and **11 are on
`loanzy.uk`, every one created 2026-08-18** — a site built yesterday and still mid-build. That
leaves **3** older active candidates. **The aggregate had folded a new site's in-flight build into a
damage signal.** Same error class as the 3,754 and as my withdrawn 107: a population assumed rather
than enumerated, for the third time in one day, caught here only by printing the rows.

### The three real candidates, each checked at the served page

Controls in the same run: an `owned` tool page that must work, and a fabricated URL that must 404.

| page | served | verdict |
|---|---|---|
| `loancalculator.co.uk/tools/credit-roadmap.html` | 200, 35KB, 3 script, **0 inputs** | **NOT damage** — its slots are `hero / ported-prose / faq / tool-cta`. A *ported prose page about* a tool. Correctly `generic`; never was interactive |
| `vetcomparison.uk/tools/compliance-deadline-calculator/index.html` | **404**, 0 components | **NOT this defect** — `status='active'` but never built. A separate defect, unrelated to the guard |
| `adversecreditmortgage.co.uk/tools/eligibility-tool/index.html` | 200, 0 components in DB | in-build (page created 08-18) |
| **control** `gamesdesign.co.uk/tools/ttk-calculator/` (`owned`) | 200, 5 script, 4 inputs | works |
| **control** fabricated `/tools/definitely-not-a-real-tool-control/` | 404, **2690 bytes** | discriminates — and it is byte-identical to the 404 above, which is what makes that 404 a genuine absence rather than a fetch artefact |

**And the seed case settles it in the other direction.** `tool-pet-treatment-cost-estimator` — the
page this whole question started from, marked `generic` — serves **200, 21KB, 5 scripts, 2 inputs**.
**Its tool is fully intact.** The page most likely to show harm shows none.

### What this does to my own §4.5(b) population claim — it narrows it

"89 identically-shaped pages" was true of the two tests I ran (name + URL) and **false as an
implication**, because a third axis separates them: `ported-prose` landing pages sit under `/tools/`
with a `tool-` name and are *correctly* `generic`. Counting only pages carrying interactive
controls: **69 of 89 `generic` vs 83 of 95 `owned`**. So the marking really is inconsistent across
~69 genuinely interactive tools — **but inconsistent is not damaged, and nothing here is damaged.**

### Recommendation, and the reasoning rather than the caution

**Do not file.** The lane thread's precedent is that we do file on population-with-unpriced-damage
(`314`, `300`) **when a mechanism is read at source**. (b) has no mechanism — nothing shows the
marking is *wrong* rather than a deliberate distinction — and now it has a **priced consequence of
zero**. A bug file asserting a cross-cutting cause behind a harmless inconsistency is the 2026-07-31
ruling's exact target.

**What would reopen it:** a `generic`-marked interactive tool page found serving prose where its tool
used to be. The query above is the detector and it currently returns nothing. Recorded here so the
zero is dated and re-runnable rather than remembered as "someone looked once".

**A live defect found in passing, which is NOT mine and NOT this:**
`vetcomparison.uk/tools/compliance-deadline-calculator` is `status='active'` with zero components
and serves **404** since at least 2026-07-17. That is a page the estate believes it is publishing
and is not. Unowned as far as I can see; not filed by me because I have not looked for its cause.

> **UPDATE 2026-08-19, same session — the "live defect found in passing" above HAS A HOME, and it is
> a class, not a page. Do not re-file it.** I wrote "unowned as far as I can see"; that was true when
> I wrote it and is now wrong. The lane thread verified it first-hand and contributed it to
> **`bugs_open/315`** (`83459d0d2`) — *"`pages.deployed_at` is stamped whether or not the object is
> written"*, filed 08-18 by `webdesign_tool_rebuilds`, **OPEN and UNOWNED**. Grounded rather than
> relayed: I checked the commit and the file, and 315 §"Contribution, 2026-08-19" now names the page.
> **The page is a second instance one month older than 315's own**, and the class is **42 pages
> across 14 sites** at `active`+`planned`+zero components, plus **2 at `active`+`deployed`+zero
> components** — that pair being the sharper form, where the estate believes it is publishing and has
> nothing. My page also carries **3 completed `page_rerender` items** (08-11, 08-12, 08-18): three
> rerenders that completed on a page with nothing to render.
> **And the detector for this already exists and files nothing** — `diagnose_silent_check_action.go`
> carries both `gatherNavLinkedNeverBuilt` and `deployed_zero_components`, and a fleet-wide all-time
> query returns zero rows. Undriven, not missing, which decides whether the fix is code or a
> schedule. **Neither this session nor the lane thread owns 315; it is genuinely available.**

---

## 2026-08-19, end of day — `473` applied, my prediction was wrong, and the shared-ledger passenger problem happened to me in BOTH directions in one session

### The prediction failed, which is why it was worth writing down

`473`/`474` applied 10:34Z. I had pre-registered (handoff §7c) that an owned-page `literal_markdown`
item would land in **"exactly one of three outcomes"**. **It landed in a fourth** — refused by the
**completion verifier** on a ported slot's `rendered_html` code_span, *before* `pageIsOwnedForGuard`
was reached, because ported/tool slots have no `content_data` to strip and carry their HTML through.

**The enumeration felt exhaustive because I built it from the two mechanisms I happened to know
about** — my own guard, and the escalation path a peer had described that morning. I never asked
what else runs on a completion. Full account in handoff §7c-RESULT; the two model changes it forces
are there too, and the second (a registered verifier means the silent-completion path cannot
false-green) is a genuine floor under `bugs_open/315`.

### And a consequence that is ours, which I would not have found without getting the prediction wrong

Re-routing created a **new pair**: `(literal_markdown, page-rerender)` at **0 complete / 3 failed**,
while the old `page-build-handler` pair keeps its 3/35. **Our `known_good` gate needs ≥1 lifetime
completion, so our promoter now holds their fix**, and `held-pair-canary-escalation` will escalate it
in three days asking for a canary — which reads like an escalation *about* the fix. Told them, with
the remedy (their post-roll canary IS the unblock; one completion releases the queue). Written up as
a `LANDMINES.md` entry because they are the second lane it could bite.

### ⚠ THE SHARED-LEDGER PASSENGER PROBLEM, observed in BOTH directions within one session

CLAUDE.md warns that a pathspec commit cannot exclude a **same-file** edit. Both halves happened here
today, which is better evidence about the ledger's real behaviour than either alone:

1. **I carried someone else's work.** Committing my `copy_edit_proposed` landmine (`8af48db0c`) took
   another session's in-flight edit to the twin-identity entry — a "SHARPENED 2026-08-19" bullet and
   an anchor fix. **I declared it in the commit message and credited them**, which is the whole
   remedy available.
2. **Someone else carried mine.** My re-route landmine was appended, the verifier dispatched, and
   before I could commit it the file was committed by another session as part of `b2066634f`
   (a `pages.deployed_at` entry). **Nothing was lost** — the text is in HEAD and the verifier is
   armed — but `git log` for that entry now points at a commit about something else.

**What this means practically, and it is not "be more careful":** on a fleet-wide append-only ledger,
**authorship is not recoverable from `git log` and should not be relied on.** The remedy that works is
the one already in the format — every entry carries its own `- **added:** <date>, <lane>` line, which
survives whoever's commit happens to carry it. **That footer is the attribution, not the commit.**
I would not have believed how quickly both directions occur: the window between my append and my
commit attempt was under two minutes.

⚠ Corollary for anyone checking their own diff before committing a shared ledger: **a clean
`git diff` is not evidence your edit is uncommitted** — it may mean somebody has already committed it
for you. `git log -S '<a phrase from your entry>'` is the check that distinguishes the two, and it is
what I used here.

---

## 2026-08-19 ~16:20Z — the re-route landmine gets its FIRST MEASURED INSTANCE, and the half nobody wrote down is that re-pointing is PARTIAL

Session picked the lane up from `HANDOFF_2026-08-19_continue_here.md` and ran its §9 checklist. Chassis
still `v1.0.1315` (pods `bfw5n` 12:15:19Z, `nkdkl` 12:15:42Z) — **no new roll, so §1's binary probe
still describes the running binary** and was not re-run. Lane tree clean, everything committed.

### 1. Today's escalation tick: the load-bearing half of the §10.A prediction HELD

`held-pair-canary-escalation` fired **2026-08-19 12:58:16 UTC** [MEASURED, `kafka-scheduler` logs,
`pre_query_result`]:

```
escalated=0  reclaimed=0  watching=13
```

**§10.A predicted `escalated=0, watching=15`.** The `escalated=0` half — the one the row exists to
defend, *"ZERO IS CORRECT, not a failed migration"* — is **confirmed**. The count was 13 because the
held pile had already dropped 15 → 13, which §7e recorded earlier the same day. A stale count, not a
wrong prediction; recording the distinction because the row was written to stop a future reader
reading the zero as breakage, and it did its job.

⚠ **`watching` had moved again by the time I measured (13 at the tick, 12 at 16:15Z).** One
`literal_markdown` row left the held population in those three hours — see §3. **The held set is not
stable between a tick and your query**, which is landmine-family item 3 in §10.D and bit me inside
twenty minutes of reading the warning.

### 2. TWO DEFECTS IN THE INSTRUMENT §10.A NAMES — `watching_detail` cannot be read the way that row reads it

I went to the readout to check tomorrow's prediction and found the same pair listed **twice**:

```
placeholder_contact->page-build-handler (canary, day 1 of 3),
placeholder_contact->page-build-handler (canary, day 3 of 3)
```

Read the `pre_query` rather than guessing (`SELECT pre_query FROM scheduled_tasks WHERE
name='held-pair-canary-escalation'` — note the column is `name`, **not** `task_name`). Both defects
are in one line, and both make the readout say something milder than the truth:

```sql
string_agg(DISTINCT item_type||'->'||handler_agent||' ('||hold_kind
           ||', day '||(now()::date - created_at::date)||' of 3)', ', ') FROM classified
```

**(i) `DISTINCT` is applied to a string containing the PER-ROW `created_at`, while the clock runs on
`min(created_at)` per PAIR.** So a pair whose rows span two dates appears as two entries at two
different day-counts. Confirmed exactly [MEASURED]: `count(DISTINCT created_at::date)` per pair is
`2,1,1,1` and the readout has `5` entries for `4` pairs — the only pair with two dates is the only one
printed twice. **You cannot count held pairs off this line, and the low entry is a lie about the
pair**: the `overdue` CTE joins *every* classified row of the pair, so the "day 1 of 3" rows escalate
in the same tick as the "day 3 of 3" ones.

**(ii) The day counter is DATE arithmetic; the predicate is TIMESTAMP arithmetic.**
`(now()::date - created_at::date)` vs `HAVING min(created_at) < now() - interval '3 days'`. So
**"day 3 of 3" does NOT mean "fires this tick"** — `placeholder_contact` printed *day 3 of 3* at
12:58Z while its clock does not expire until **19:17:45Z**, six hours after the tick that said so.

⚠ This is §10.A's own "off by a full tick" landmine — **except it is not the reader's arithmetic, it
is inside the instrument**, which is why writing that warning did not protect against it. The
trustworthy reading is the clock itself, not the readout:

```sql
SELECT item_type||' -> '||handler_agent, count(*), min(created_at),
       min(created_at) + interval '3 days' AS escalates_after
FROM classified GROUP BY 1;   -- classified = the pre_query's own CTE, copied verbatim
```

**The four held pairs and their REAL clocks [MEASURED 2026-08-19 16:15Z]:**

| pair | kind | detected rows | oldest | clock expires | escalates at tick |
|---|---|---|---|---|---|
| `placeholder_contact → page-build-handler` | canary | **3** | 08-16 19:17:45 | **08-19 19:17:45Z** | **08-20 12:57** |
| `missing_conversion_path → content-gap-planner` | canary | 1 | 08-17 22:21:46 | 08-20 22:21:46Z | 08-21 12:57 |
| `dead_fragment_link → page-build-handler` | canary | 1 | 08-18 01:38:47 | 08-21 01:38:47Z | 08-21 12:57 |
| `literal_markdown → page-build-handler` | floor | **7** | 08-18 07:23:16 | 08-21 07:23:16Z | 08-21 12:57 |

**Every DATE in §10.A is confirmed by this.** One row count is stale: §10.A says `literal_markdown`
escalates **10** rows on 08-21 and §7f says **8**; it is **7** now. Not a contradiction — the
population drains (§3). **Do not carry a row count forward from any of the three; re-derive it at the
tick.**

⚠ **Do not canary `missing_conversion_path → content-gap-planner`** — `bugs_open/255` owns it, and
that has not changed.

### 3. `literal_markdown` — §7f's "one number to watch" is ANSWERED, and the answer is comfortable

§7f left the pair at 1 complete / 2 failed and flagged the arithmetic: `floor_ok` becomes binding at
the 5th outcome with 2 failures banked, so *"next two both fail → 1/4 → below the floor → HELD"*.

**[MEASURED 2026-08-19 16:16Z, live + archive per §8b]:**

| `literal_markdown` handler | complete/verified | failed | detected | % of outcomes |
|---|---|---|---|---|
| `page-rerender` (**the new route**) | **7** | **1** | 0 | **87.5%** |
| `page-build-handler` (the old route) | 3 | 34 | **7** | 8.1% |
| `page-content-writer` (older still) | 2 | 9 | 0 | 18.2% |

**Eight outcomes, 87.5% — the floor is not going to bind, and the worry in §7f is closed.** Seven of
those completions landed **between 16:00Z and 16:11Z today**, i.e. while I was reading the handoff.

⚠ **A third handler exists.** `page-content-writer` (2/9) is in neither §7e nor §7f, which both frame
this as a two-route story. The producer's own header records the chain —
`page-content-writer → page-build-handler` (08-05) `→ page-rerender` (08-18). **Three eras, three
pairs, and each one keeps its own record for ever.**

### 4. THE FINDING: the re-route landmine's missing instance is here, and re-pointing is PARTIAL

`LANDMINES.md` ("Re-routing an `item_type` to REPAIR it creates a NEW pair…") was retracted to a
**derived property with no measured instance**, and says explicitly: *"Nobody has yet been observed
doing it. If you hit a real instance, replace this bullet with it."*

**This is the instance, and the evidence is that rows OUTLIVED the literal that filed them.**
`check_literal_markdown.go:402` sets `HandlerAgent: "page-rerender"`, changed by `763bb5d55` on
**08-18 20:08** and live only since the **12:15Z** roll today. So any row created *before* that was
filed with `page-build-handler`. Yet [MEASURED]:

| created_at (batch) | on `page-build-handler` | on `page-rerender` |
|---|---|---|
| 2026-08-18 07:23:16.545362+00 | **7 `detected`** | 1 `failed`, 1 `triaged`→`claimed` |
| 2026-08-17 12:31:06.459751+00 | — | 4 `complete` |
| 2026-08-17 19:21:16 / 01:18:44 | — | 1 `complete` each |

**Rows from one detector run, to the microsecond, now sit on both handlers.** The producer cannot have
done that. **`handler_agent` was mutated on existing rows after creation.**

**And the dispatch provenance shows the landmine's own prescribed sequence being executed:**
4 of the new-route rows carry `pipeline='build'` + `spec.original_pipeline='content'` — **the
hand-canary recipe verbatim from migration `466`'s `what_to_do` text** — and 5 carry the promoter's
plain `pipeline='content'` with no `original_pipeline`. Ordered by `updated_at`, the `build` ones come
**first** and the `content` ones **after**. That is *re-point → hand-canary → `known_good` flips →
promoter takes the rest*, observed in the artefacts rather than predicted. **The landmine's remedy is
now measured, not just reasoned.**

#### The half that is NOT in the entry, and it is the one that costs something

**Re-pointing was partial, and the rows left behind cannot be reached by anything.** Seven rows of the
08-18 07:23:16 batch still carry `page-build-handler`, and their `updated_at` **equals their
`created_at`** — untouched for 33 hours while their siblings were repaired.

- The **promoter** will not dispatch them: their pair is 3/34 = 8.1%, held under the floor.
- The **producer** cannot re-file them onto the new route: `idx_swi_dedup` holds one open row per
  `(site_id, item_key)` and the key is `literal_markdown:<page_id>`, so a re-detection of the same
  page is dropped while the old row is open. **[INFERRED from the index semantics + the codebase's own
  "silently drop" language — I did not read the INSERT itself.]** What is **[MEASURED]** is that the
  rows have not moved in 33 hours.
- **Demand control, because this lane keeps shipping zeros that mean nothing:** the discovery
  machinery is emphatically alive — **60 rows across 11 item types filed since the 12:15Z roll**, most
  recent 16:14Z. It is not "the checks stopped running". **Zero `literal_markdown` rows have been
  filed since 08-18 07:23:16.**

**So on 08-21 those 7 rows escalate to `needs_human_review` carrying the reason *"the pair succeeds
below 25%, the promoter has stopped feeding it"* — which is TRUE of the route they are pinned to and
IRRELEVANT to their defect, which now has a working, artefact-verified repair completing at 87.5% on
the adjacent route.** That is `083`'s disease at its purest, and §7f called it before it was
measurable: *a real finding, a working repair that exists, and no path between the two.*

~~**The remedy is known and already proven in this very population** — an explicit `UPDATE … SET
handler_agent='page-rerender'` on the stranded rows, which is what somebody did by hand for the ones
that drained. **It is not ours to fire** (the type belongs to the `184`/`201` lanes, and the
escalation's own `owners` map names them), but it should be *decided* before the 08-21 tick rather
than after a human is invited to canary a route that has already been superseded.~~

> ## ⚠⚠ CORRECTED 2026-08-19 21:00Z, ~4½ HOURS LATER, BY ME — THE STRUCK-THROUGH REMEDY WOULD HAVE DONE DAMAGE
>
> **What caught it:** I went to route this finding at the owning lane and ran `scripts/who-owns.py`
> first, as the rules require. It said `bugs_open/184` is **CLOSED** — closed *today*, `0ca143c2d` —
> and its close-out reads *"residuals routed (owned/ported → 301/tool-rebuilds …)"*. **A lane does not
> route a residual it has not characterised**, so I finally asked the question I had skipped: what
> **are** the 7 rows? I had written a remedy for them into three files without ever looking.
>
> **[MEASURED 21:00Z] All 7 sit on `rebuild_policy='owned'` pages** — `tool-cubic-bezier`,
> `tool-grid-generator`, `tool-json-cleaner`, `tool-noise-generator`, `tool-text-extractor`,
> `tool-head-architect`, `learn-design-physics-of-ui`. **And the decisive contrast, the same query
> split by policy:**
>
> | `literal_markdown → page-rerender` | rows |
> |---|---|
> | `generic`, **complete** | **8** |
> | `owned`, **failed** | **1** |
>
> **Every one of the new route's successes is a generic page. Its single owned attempt failed.** The
> new route calls `save_page_sections` and is refused by the ownership guard on owned pages — **which
> is this lane's own §7b warning to the 184 lane, now confirmed at n=1 against my own finding.**
>
> **So the `UPDATE` I proposed would have converted 7 quiet rows into 7 loud failures**, dragged
> `(literal_markdown, page-rerender)` from 8/1 toward its floor, and repaired nothing. It is the
> worst kind of wrong: confidently actionable.
>
> **THE MECHANISM SURVIVES; THE INFERENCE DOES NOT.** Still true and still measured: existing open
> rows keep the old `handler_agent`, cannot be dispatched (old pair held below floor), cannot be
> re-filed (dedup). **False:** *"therefore a working repair exists next door and only routing
> separates them."* **Whether the new route can SERVE the old rows is a separate question with its own
> answer, and I never asked it** — I reasoned entirely about the transport and not at all about
> whether the destination would accept the cargo.
>
> ⚠ **And note how well-defended the wrong claim was.** It had a demand control, an `[INFERRED]`
> marker on the dedup half, a measured 33-hour gap, and a same-microsecond batch as its smoking gun.
> **Every one of those was true.** The error was upstream of all of them: a population I had counted
> but never *described*. This is the lane's own landmine-family item 1 — *a population assumed rather
> than enumerated* — for the seventh time, and the marker discipline cannot catch it, because nothing
> I wrote was unmeasured.
>
> **The corrected finding is better for this lane than the wrong one was.** These 7 rows are `277`'s
> subject: an owned page carrying a real, deterministically-repairable defect has **no repair route at
> all** — the generic mechanical fix refuses it, and nothing else claims it. `184` closed correctly
> and routed exactly this class **to us**. That is `no_content_data`'s hole reached from a new
> direction, and it strengthens `277`'s clause-1 blocker rather than adding a separate problem.

> **THE TRANSFERABLE PROPERTY, and it is the sharper half:** re-routing a producer fixes only
> **FUTURE** findings. Every open row filed under the old literal keeps the old `handler_agent` for
> ever — the producer cannot re-file over it (dedup) and the promoter will not dispatch it (old pair
> held). ~~**"Re-route the producer" is half a migration; the other half is an explicit UPDATE of the
> existing backlog, and nothing warns you it is missing**~~ — the new pair drains beautifully while
> the old rows sit still, which reads as success.
>
> > ⚠ **CORRECTED 2026-08-20 07:00Z — the struck-through sentence is the RETRACTED remedy, and until
> > now it was the LAST WORD in this file.** The correction box immediately above it already says an
> > `UPDATE` of the backlog would have produced 7 more failures and repaired nothing; this paragraph
> > then prescribed exactly that `UPDATE` as "the other half", eleven lines later. Anyone reading
> > NOTES from the bottom — which this lane's own handoff instructs — met the refuted version first
> > and the retraction second. **A correction that does not also fix the summary line is half a
> > correction**, and the summary is the half that travels.
> >
> > **The corrected general form** (it is the one already in the handoff §5 and in `LANDMINES.md`):
> > re-routing a producer strands its existing backlog, and the backlog is **not automatically
> > re-pointable**. Before proposing to move those rows, ask the question that was skipped here —
> > **would the new route even ACCEPT them?** Split the new route's record by whatever its guard keys
> > on (here `rebuild_policy`: 8 complete on `generic`, 1 failed on `owned`). If it would refuse
> > them, the residual belongs to whoever owns the **blocker**, not to whoever owns the item type.

### 5. Council round 2 on `301` was mid-flight while I wrote this

`RESUBMIT_CORR=c7bc1b9e`, orchestration `6469c138`, started **16:08:21Z**; at `review_bug_historian`
16:11Z, `review_architecture` 16:17Z. Round 1 came back `complete_revise` at 11:12Z. **Verdict not yet
read — do not write a `Council-Reviewed:` trailer for it until someone has actually read it**
(the commit that resubmitted correctly used `Council-Submitted:`). ⚠ The architecture seat is in this
round, and its known landmine is truncation — check the verdict is whole before believing it is mild.

---

## 2026-08-19 ~20:40Z — `v1.0.1316` rolled, `301` came back APPROVED, and its three medium advisories were checked rather than waved through — two are REFUTED and the third made our deploy proof better

> **⚠ CORRECTED — my own entry above, four hours old, is now stale on its first line.** It says
> *"Chassis still `v1.0.1315` … no new roll, so §1's binary probe still describes the running
> binary and was not re-run."* **A fresh build rolled at 17:13Z.** That sentence was true when
> written and is exactly the shape this lane keeps logging: a deploy fact with a shelf life of
> hours, written without one. **Caught by the owner telling me, not by any check of mine** — I had
> no re-check scheduled because I had concluded there was nothing to re-check. The cheap check that
> would have caught it is the same one I then ran: read the pods, not your last reading of them.

### 1. The new roll, re-probed on both replicas — and this time proven for all **57** pods, not 2

`agent-chassis:v1.0.1316`, pods `86nqf` (17:13:39Z) and `8jlqh` (17:14:01Z). The `build provenance`
startup line had **already scrolled** (this service emits ~3.7MB to `--tail=400`) — **"not in range",
never "unstamped"** — so the binary probe, which has no shelf life. Both replicas, one pass:

| needle | result | role |
|---|---|---|
| `owned_page_refusal_status` | **PRESENT** | Tier 1 (`083`/`301`) |
| `resolveStatusRepairComponent` | **PRESENT** | `300`'s fix |
| `refuse_owned_page` | **PRESENT** | `301`'s opt-in key (mig `488`) |
| `OWNED_PAGE_GUARD` | **PRESENT** | control: the probe works |
| `ZZQQ_NEEDLE_THAT_MUST_NOT_EXIST` | **ABSENT** | control: the probe discriminates |

⚠ **The probe takes ~2–7 minutes per pod** (it scans the whole binary); a 2-minute timeout kills it
mid-scan and the partial output looks like a clean negative. Give it 420s.

### 2. `301` COUNCIL ROUND 2 — **APPROVED**, 3 advisory objections, none high-severity

`RESUBMIT_CORR=c7bc1b9e-97c8-4f3e-8a4f-b3a7029505ee`, orchestration `6469c138`, `complete_approved`
**COMPLETED 16:19:28Z** — **11 minutes end to end**, against the ~30 minutes CLAUDE.md budgets.
`gated_by_truncation: false`, so the architecture seat's truncation landmine did **not** fire; its
verdict is whole and reads `ARCHITECTURE_SIGNAL: point_fix`, explicitly calling this *"the RFC_022
exception applied correctly, not just claimed"*.

**Round 1 (`b5b85e4b`) was `complete_revise` at 11:12Z; round 2 approved at 16:19Z.** The lane's
standing claim that a REVISE round is cheaper than the defect it finds holds again.

**No trailer action is owed on the code.** `6be66bceb` already carries `Council-Submitted: c7bc1b9e`,
and `098` resolves the correlation at **report** time, so the commit is credited automatically once
the verdict turns approved — no amend, which forward-only forbids anyway.

> ⚠ **My own bookkeeping slip, recorded rather than left to confuse the coverage report.** I put
> `Council-Reviewed: c7bc1b9e…` on the **LANDMINES docs commit** (`895df4e2f`). It is not a false
> claim — I read the full verdict and every objection before writing it — but the trailer exists to
> join a **platform-code** commit to its verdict, and docs are refused by the gate client-side. So it
> credits a commit that never needed review. Forward-only means it stands; **do not repeat it**, and
> if `098` shows an odd row for that commit, this is why.

### 3. The three MEDIUM advisories, each checked first-hand. Two were WRONG.

The lane's record is that advisories catch real defects, so these were checked rather than accepted.

**(a) `diagnosis_guardian` — REFUTED AT SOURCE. The seat's standing discipline is stale.**
Its objection: migration `488` sets `error_step` at the **step** level, but *"the coordinator reads
ONLY `step.config.error_step`; a step-level one is parsed and silently inert"* — making our routing
*"real or coincidental"*, unresolvable from 3 production refusals. **The code says the opposite, in a
comment written for exactly this question** (`platform/orchestration/coordinator.go:3667`):

```go
// Check step-level first (parallel to NextStep) — preferred location
if step.ErrorStep != "" { return s.routeToErrorStep(...) }
// Fallback to config-level for backward compatibility
if errorStep, ok := step.Config["error_step"].(string); ok && errorStep != "" { ... }
```

**Step level is the PREFERRED location and is checked FIRST; `step.config` is the backward-compatible
fallback.** The seat has the precedence exactly inverted. **This is worth telling them** — a stale
standing discipline in a council seat mis-fires on every future submission that does the right thing,
which is the same "following the rule draws the objection" pathology RFC_022 was narrowed to fix.

**(b) `bug_historian` — PREMISE VOID. The case it rests on is CLOSED.** It cited *"an OPEN case with
exactly this shape: `bugs_open/086` step_level_error_step_dropped_by_the_plan_converter"* and called
our 3-row sample thin against *"a documented, still-open drop bug"*. **086 is in `bugs_closed/`,
closed 2026-07-27** — and closed on precisely the evidence the objection asks for: *"the persisted
plans show a clean 0 → 10 step across the roll boundary"*, i.e. the converter was proven on data, not
on `strings`. ⚠ Note it landed the ambiguity trap too: a bare `086` is checkable in one `ls` across
**both** directories, which is why the rule is to resolve by slug and grep both.
**Their underlying point still stands as a distinction worth keeping**: the *coordinator* honouring
step-level `ErrorStep` (a) and the *plan converter* preserving it (b) are two different layers, and
only (b) was ever the bug. Both are now answered, by different evidence.

**(c) `debug_historian` — VALID, ALREADY A LANDMINE, AND NOW ANSWERED BETTER THAN IT ASKED.**
Its objection: *"both replicas verified"* may not cover the pods capable of running the step, citing
lore of 41 pods vs 2. **It is right, and it is `LANDMINES.md` line 5909** — which I had not read,
because the SessionStart hook only matches entries against files already dirty in the tree and a
`kubectl` footprint matches no path. **[MEASURED 2026-08-19 20:35Z]: 57 pods run the chassis image;
only 2 carry `app=agent-chassis`.** The count has grown 41 → 57.

**But the decisive check is the DIGEST, which also answers their second objection (same-tag cache) in
the same command:** all 57 pods resolve to **one** `imageID`, `sha256:2d0d3def…`, `distinct digests:
1`. **That upgrades "I probed 2 replicas" into "those 2 replicas' bytes are every pod's bytes"**
without exec'ing 57 pods. Added to that landmine entry (`895df4e2f`) with the one-liner, because the
entry prescribed enumerating by *image*, which proves the tag and cannot prove the bytes.
⚠ Digest identity is **not** reachability — every pod runs the code; whether any executes it is the
entry's existing positive-control query.

**The pattern across all three: every objection was cheap to check and none needed the author's
word.** (a) and (b) were refuted by one `grep` and one `ls`. That is the argument for reading
advisories on an APPROVED verdict rather than filing them — two of these would otherwise have entered
the record as unresolved doubts about a mechanism that is fine.

---

## 2026-08-20 ~08:05Z — the escalation task's `owners` map was pointing at three dead or wrong destinations, and it is a ONE-SHOT write. `497` applied before today's tick

Session-start checklist (handoff §9) ran clean and produced one new finding, which became the
session's work. Recording the checklist results first because two of them are the kind that go
stale in hours.

### 1. Checklist: the roll, the clocks, the owners

- **`v1.0.1316` still live, and this is the cheap way to say so.** `distinct digests: 1`,
  `sha256:2d0d3def…` — **the same digest §2 probed last night.** So last night's 5-needle binary
  probe still describes the running bytes and did **not** need re-running: digest identity to an
  already-probed digest transfers the probe. That is worth having as practice — the probe costs
  ~420s per pod, the digest check costs one `kubectl get pods`.
  ⚠ **Pod count moved again overnight: 94 now (`Job` 90, `ReplicaSet` 4).** Last night three
  sessions read 22/57/85 within half an hour. **Do not quote a pod total**; quote
  `distinct digests` and the `ReplicaSet` count, exactly as the handoff warns.
- **`scripts/who-owns.py` by slug on `277`, `083`, `300`, `314`: no competing owner.** Every commit
  against all four files is this lane's.
- **§6's four clocks re-derived from the live `pre_query`: confirmed to the microsecond**, and
  `placeholder_contact → page-build-handler` (3 rows) is now **overdue** — its clock expired
  2026-08-19 19:17:45Z, so it escalates at today's tick. The task is `enabled`, `interval_seconds
  86400`, `last_triggered_at 2026-08-19 12:58:16Z` — so **~12:58Z today**.

### 2. Two things in last night's handoff that today's commits had already moved

- **`301` is CLOSED** — the mid-move the handoff said to keep hands off completed. At HEAD (checked
  with `git ls-tree`, never `ls`) it is `bugs_closed/301_…`.
- **`bugs_open/333` was filed at 22:02Z**, after the handoff was written, and it takes the
  **producer/routing half** of §5's finding. The division is now explicit and is written into
  `bugs_open/277` as a CONTRIB: **333 = routing** (producers don't read `rebuild_policy`),
  **277 = repair design** (what actually repairs an owned page). That CONTRIB also measured
  something that belongs to us: **our own converter filed 28 `content_rewrite` items on owned
  pages on 08-18 alone**, making it the newest large producer of the class.

### 3. THE FINDING — all three entries in the `owners` map are dead or wrong, and the map is written ONCE

`held-pair-canary-escalation` maps `item_type → owning lane` and stamps it into
`result.held_pair_escalation.owner` **at escalation time, never revisiting it.** So a wrong entry is
not "stale config" — it is a wrong instruction handed to a person, once, permanently, at the only
moment they were ever going to read it. **[MEASURED 2026-08-20 06:49Z]**

| entry | pointed at | actually |
|---|---|---|
| `placeholder_contact` | `bugs_open/201 lane …` | **`bugs_open/201` DOES NOT EXIST** — closed |
| `literal_markdown` | `bugs_open/184 + bugs_open/201 lane …` | **NEITHER EXISTS** — both closed |
| `page_component_status_drift` | `(UNASSIGNED - claim this) … no lane doc claims it` | **half true** — see below |

Verified at HEAD, **not with `ls`**: `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ |
grep -E '/(184|201)_'` returns three paths, **all** under `bugs_closed/`. (184 is one of the
documented ambiguous numbers — it names two unrelated cases, and both are closed.)

**The third is the one with a PROVEN consumer, and it is why this was worth doing today rather than
noting.** The only escalation this family has ever produced —
`page_component_status_drift → component-template-fixer`, `result…'at' = 2026-08-17T12:57:43Z` —
carries **exactly** the string `(UNASSIGNED - claim this) … no lane doc claims it`. **A human was
already told "nobody owns this" about an item type that has an open, council-approved lane
(`bugs_open/300`).** So the defect is not prospective; it has already fired once, at n=1 of 1.

> ⚠ **And it is HALF true, so it is corrected rather than replaced.** `check_page_component_status_
> drift.go` really is untouched since 2026-07-10 (`git log` on the file: one commit, `7813f3eb9`).
> What changed is that the **ITEM TYPE** acquired an owner, which it did not have when `466` was
> written. Replacing the whole string would have destroyed a true statement to fix a false one.

### 4. The structural half — WHY it rotted, which is the transferable part

The map hardcodes a **`bugs_open/` PREFIX**. That prefix is precisely the thing that **flips when
the lane you are pointing at SUCCEEDS.** So the map was built to go stale at exactly the moment its
pointer starts to matter — and it did, while findings were still queued against 184 and 201.

Correcting three strings would have re-armed the same trap for the next close, so `497`'s
replacement 4 writes the rule **into the map itself**: name a bug by **NUMBER and SLUG**, never by
directory; a bare number is ambiguous; name the **lane directory**, which does not move.
Now in `LANDMINES.md`.

### 5. `497` — applied 2026-08-20 ~08:00Z, verified at the live column

`docs/agent_docs/sql_for_agents/497_escalation_owners_map_points_at_three_dead_destinations.sql`
(+ `_ROLLBACK.sql`), commit `a888224f0`. Same surgical-anchor pattern as `479`: four anchors each
asserted to occur exactly once, guarded on the whole text's md5, transcribing nothing.

**Behavioural surface NIL, and proven so rather than asserted.** Four string literals in a lookup
CTE; no predicate, clock, threshold or row selection. `51547db8…` (10,566) → `406bd757…` (13,106).

Three things about the controls, because two of them are the reusable lesson:

- **The load-bearing control is the REVERSE REPLACEMENT.** Put the new strings back and assert the
  md5 returns to the pre-image. It is the **only** check in the file that can catch collateral
  damage to the 10 KB of `what_to_do` prose — every presence/absence check passes identically
  whether that prose survived or was mangled. **MUTATION-PROVEN:** damaging one byte of unrelated
  prose (`'Promote ONE row by hand'` → `'Promote TWO rows by hand'`) makes it fail with
  `got 11eeb896… len 10567`; the clean file passes. So it is disconfirmable, not decorative.
- **Forward and reverse are driven from the SAME eight variables, declared once.** The first draft
  wrote the undo out by hand and it was **asymmetric** — `new4` was a prefix of the line it was
  meant to remove, so the undo left ` That\n` behind. The control caught it **before the file was
  ever run**. A control transcribed a second time tests the transcription, not the change.
- ⚠ **MY OWN ASSERTION FIRED ON MY OWN REMEDY, and it was right to.** I asserted `bugs_open/`
  occurs exactly twice after the fix. It found **three**: the rule comment I had just inserted says
  *"NEVER write a `bugs_open/` … PREFIX"*, so **the remedy text contains the token the check was
  counting.** Counting a token cannot distinguish a live pointer, a dead pointer, and a prohibition
  against writing one. Replaced with the three specific claims — `bugs_open/201` = 0,
  `bugs_open/184` = 0, and **`bugs_open/300` = 2 as a POSITIVE control** that the two genuinely-live
  citations in `what_to_do` survived untouched. This is the memory-index family *"declaring a key
  silences your own detector"* arriving from the other side: here the remedy **tripped** the
  detector rather than silencing it, which is the lucky direction.

**Verified independently at the live column, not from the migration's own NOTICE:** `md5 406bd757…
len 13106`, `bugs_open/201` = 0, `bugs_open/184` = 0, `bugs_open/300` = **2** (both survive). The one
remaining `(UNASSIGNED` is the `COALESCE(ow.owner, …)` **fallback** for item types absent from the
map — correct, and it must stay.

**And the bytes being right does not prove the SQL still runs.** The `pre_query` is text the
scheduler executes, so a broken literal would fail every tick. So the corrected query was
**EXECUTED**, rolled back, against live data — `CREATE TEMP TABLE … AS <pre_query>` inside the open
transaction. It parses, runs, and previews today's tick exactly:
**`escalated=3, reclaimed=0, watching=12`**, the 3 `placeholder_contact` rows each stamped with the
corrected `083 …` string. Round trip also exercised: `51547db8 → 406bd757 → 51547db8`, and the
reverted text re-applies cleanly.

> **No council submission owed, and that is itself evidence for `314`.** A migration under
> `docs/agent_docs/sql_for_agents/` is refused by the gate **client-side** — which is precisely
> `bugs_open/314`'s complaint (*"the council gate refuses config changes because migrations live
> under a docs path"*). Third lane to hit it.

### 6. `reason` was LEFT ALONE — a decision, not an omission

For the `literal_markdown` 7, `reason` will still say *"the pair succeeds below 25%, so the promoter
has stopped feeding it"* — **true of the PAIR, and not why THOSE rows are stuck.** Making `reason`
say so requires reading `rebuild_policy` **inside this task**: real new logic, on the very seam
`bugs_open/333` was filed for hours earlier. So the correction rides in the `owner` prose (free) and
the logic question goes to 333 (where the seam is). The new `literal_markdown` owner string says
this explicitly, so the human who reads it is not misled by the `reason` beside it.

### 7. §6's `watching_detail` defect reproduced INDEPENDENTLY, by accident

The rolled-back preview printed:
`placeholder_contact->page-build-handler (canary, day 2 of 3)` **and**
`placeholder_contact->page-build-handler (canary, day 4 of 3)` — **the same pair, twice, at two
day-counts**, five entries for four pairs. That is exactly §6(i): `DISTINCT` applied to a string
containing the **per-row** `created_at` while the clock runs on `min(created_at)` per **pair**. And
all 3 rows escalated stamped `days_waiting = 4`, confirming §6's other half — `overdue` joins
**every** row of the pair, so a "day 2 of 3" row escalates in the same tick as a "day 4 of 3" one.
**I did not set out to test this; the preview showed it.** Recording it because a defect confirmed
by an unrelated run is stronger evidence than one confirmed by the query written to find it.

### 8. A LEAD on `277`'s clause 1, sitting in this task's own prose — [UNVERIFIED by me]

While reading the `pre_query` I found that `466`'s `what_to_do` **already names a candidate route
for the owned-page hole**, and it is not one this lane's docs have picked up:

> *"If protective refusals dominate, the handler is behaving CORRECTLY, the floor is mis-holding
> this pair, and the handler does not need repairing — the PAGE needs a different ROUTE. See
> `bugs_closed/295` … **Its fix candidate 3 is UNTOUCHED and is the live remedy: route content
> findings on owned pages to `section_edit`, which demonstrably works on them (18 completes).**
> ⚠ `apply_section_edit` is right for REWRITING an existing component and a DEAD END for ADDING a
> section to an owned page."*

**Why this matters to us:** last night's corrected finding was *"an owned page with a real,
mechanically-repairable defect has **no route at all**"*. That may be **too strong** — there is a
named candidate with a claimed 18 completes on owned pages. **And the caveat lines up with our
population:** `literal_markdown` is asterisks reaching the page as literal text, i.e. **REWRITING
existing content**, not adding a section — the side of the caveat that is supposed to work.

**[UNVERIFIED — I have not checked the 18, not read `apply_section_edit`, and not confirmed it
accepts a `literal_markdown` finding.]** Marked because this is exactly the shape that burned this
lane last night: a plausible route inferred from a claim about a route, with the "would the
destination accept the cargo?" question **unasked**. The three checks, in order, are: (1) does
`apply_section_edit` accept `literal_markdown`; (2) re-measure the 18 completes on owned pages
lifetime incl. archive; (3) `grep LANDMINES for apply_section_edit`, which `466` itself instructs.
**Next session's first job**, and it is a better lead than anything in the handoff's §8.

### 9. Fixed the retracted remedy that was still the LAST WORD of this file

The 21:00Z correction box retracted the *"explicit `UPDATE` of the existing backlog"* remedy — and
then, **eleven lines below it**, the "THE TRANSFERABLE PROPERTY" paragraph prescribed exactly that
`UPDATE` as *"the other half"*, unmarked. Anyone reading this file **from the bottom**, which the
lane's own handoff §0 instructs, met the refuted version first and the retraction second. Struck
through and corrected in place. **A correction that does not also fix the summary line is half a
correction, and the summary is the half that travels.**

---

## 2026-08-20 ~09:00Z — the population repaired itself out from under me while I was writing about it, `497`'s own figures went stale in twelve hours (`498`), and "no route at all" is REFUTED

Three things happened in the four hours after the entry above, and they are connected: **the thing
I had just finished measuring stopped existing.**

### 1. THE 7 ROWS ARE GONE — terminal, not escalated, and the 08-21 clock is VOID

**[MEASURED 2026-08-20 08:11Z]** The 7 held `literal_markdown → page-build-handler` rows are no
longer `detected`. Between **07:20:42Z and 07:23:58Z** — one every ~30 seconds, so a serial
dispatch loop — every one was dispatched, refused, and terminated **`wont_fix`**:

```
completed_by_step: mark_item_failed · owned_page_refusal: true
owned_page_refusal_marker: OWNED_PAGE_GUARD · owned_page_refusal_replaced_status: "failed"
```

**Why they were released:** the pair rose from **3 ok / 34 failed (8.1%, FLOOR-HELD)** at 21:14Z last
night to **19 ok / 24 failed (44%, PROMOTABLE)** — **16 completions across 3 sites inside the 07:00Z
hour alone.** Promotable pair → promoter feeds it → the owned-page rows in it go straight into the
guard.

**So `literal_markdown` has dropped out of the held set entirely and the escalation this lane had
docketed for 2026-08-21 12:57Z will NOT fire.** Today's tick (~12:58Z) is `escalated=4,
reclaimed=0, watching=6` — **4** `placeholder_contact` rows, not the 3 the handoff and my own
earlier preview said, because a 4th joined the pair overnight.

> ⚠ **This is §6's "the population drains; re-derive at the tick" landing on me, and note WHICH
> claim it broke.** Not a figure — a *docket*. I had a dated clock, correctly derived twice from the
> live `pre_query`, and the rows left by a door I had not modelled. **A clock is only a prediction
> if nothing else can reach the rows first**, and `wont_fix` reached them in 3 minutes 16 seconds.

### 2. `497` WROTE SELF-STALING FIGURES INTO THE SAME CONFIG WHOSE SELF-STALING POINTER IT FIXED

This is the day's real lesson and it is uncomfortable. `497` (08:00Z) corrected a map that had rotted
because it hardcoded a `bugs_open/` prefix — *a value guaranteed to go stale the moment the lane it
named succeeded.* In the same migration I wrote into the same live config:

> *"[MEASURED 2026-08-19] **all 7 rows held** on page-build-handler sit on rebuild_policy=owned
> pages … re-pointing these 7 would produce 7 more failures, **drag a healthy pair toward its
> floor**, and repair nothing."*

**Three clauses, all true when measured, all false within twelve hours:** the 7 are not held; the
pair is not floor-held (44%, promotable); and it cannot be dragged toward its floor at all, because
**`301` makes that mechanically impossible** — a refusal terminates `wont_fix`, which is excluded
from *both* sides of the promoter rule. **The prediction's mechanism had already been prevented by a
fix that shipped two days earlier, and I did not check.**

> **THE DISTINCTION, and it is NOT measured-vs-unmeasured** — every word of it was measured, dated
> and marked, exactly as this lane's rules require:
> - a **DATED OBSERVATION** stays true for ever: *"on 08-20 07:20Z, 7 rows were refused"*;
> - a **DESCRIPTION OF CURRENT STATE** is false the moment the state moves: *"all 7 rows are held"*;
> - **and a reader cannot tell which kind they are looking at.** The `[MEASURED <date>]` marker
>   is worn by both.
>
> A **one-shot annotation** — read by a human at escalation time, possibly months later — must carry
> the first kind plus the durable **mechanism**, never the second. The marker discipline does not
> reach this; only the tense does.

**`498` applied ~08:40Z** and de-volatilises that string: durable mechanism, one dated worked
instance, and the candidate route. It asserts each of the three stale claims is **individually
gone** (named, so it cannot pass by accident), keeps `497`'s other two corrections as **positive
controls**, and its reverse-replacement control is mutation-proven the same way. Verified at the live
column: `ddd0c894… len 13746`, and the corrected query **executed** rolled back.

### 3. "AN OWNED PAGE HAS NO ROUTE AT ALL" IS REFUTED — and the route was in this task's own prose

Following up §8 of the previous entry (the `[UNVERIFIED]` lead) instead of leaving it for next time.

**[MEASURED 2026-08-20, live + archive] `section_edit → section-editor`, split by `rebuild_policy`:**

| policy | complete | failed | total |
|---|---|---|---|
| **`owned`** | **36** | **1** | **39** |
| `generic` | 53 | 4 | 57 |
| (no page row) | 132 | 0 | 133 |

**92% on owned pages** — against `466`'s own conservative "18 completes". Compare the generic repair
on the same axis: `literal_markdown → page-rerender` is 8 complete on `generic`, **1 failed on the
single `owned` page it tried.** The two routes are not both blocked on owned pages: **one is refused
by design and the other is how this estate already edits them.**

**The severe landmine was CHECKED, and its precondition is absent.** `LANDMINES.md` warns that a
`section_edit` on a per-site **TOOL FORK** whose template carries `{{.field}}` copy and whose
`content_data` is `{}` re-renders every text node to **EMPTY** while every floor passes. Six of our
seven pages are `tool-*` — a direct hit on shape, and the thing that would have made "use
`section_edit`" another confidently-actionable wrong answer.

| page / slot | `component_level` | `content_data` empty | `{{.field}}` hits in TEMPLATE |
|---|---|---|---|
| all 7, `ported-page` | **section** | no | 1 |
| grid-generator, json-cleaner, noise-generator — `tool-*` slot | **tool** | **yes** | **0** |

**The trap needs BOTH halves.** The tool forks do have `content_data='{}'` — and **zero** template
fields, so there is nothing for an empty `content_data` to fail to fill. **And the literal markdown
is not in the tool fork at all:** it is in `ported-page`, a section-level component whose
`content_data` **is** populated — the ordinary 36/1 target. `466`'s caveat (`section_edit` REWRITES,
cannot ADD) fits too: literal asterisks in existing prose are a rewrite.

> ⚠ **MY FIRST CHECK FOR THAT WAS NEARLY VACUOUS AND I BRIEFLY BELIEVED IT.** I grepped
> `page_components.rendered_html` for `{{.` — 0 on all seven pages — and read it as "the trap does
> not apply." **It proves nothing:** `rendered_html` is the **RENDERED OUTPUT**, so a template field
> that resolved to empty leaves no `{{.` behind *either*. **The measurement returns 0 whether or not
> the risk exists** — the memory-index rule *"a `[MEASURED]` figure is only evidence if it could
> have come out otherwise"*, hit twice in two days. The template is the only place the question can
> be asked, which is why the table above joins `content_components`.

### 4. What I did NOT claim, deliberately

- **Clause 1 is not closed.** Nothing has been repaired. `no_content_data` (27 of the 30 parked) is
  untouched by any of this — the generic-repair-refuses-owned-pages finding does not reach it.
- **The remaining unknown is one code question**, not a design: can a producer file a `section_edit`
  item for a `literal_markdown`-shaped finding at all — what `spec`/`field_updates`, targeting which
  `page_component_id` (here, `ported-page`). Three further landmines already written down apply to
  whoever answers: the `field_updates` merge is **per-field and reverts intervening edits**;
  `apply_section_edit` writes `rendered_html` with **no content validation**; it **cannot ADD**.
- **The `wont_fix` blind spot is NOT my finding.** `bugs_closed/301` §(~285) and `bugs_open/333`
  §§96–109 already record both halves — exclusion from the floor, and dedup re-filing. **Prior art
  checked before writing, and it changed what I wrote:** what is new is the joined-up cycle measured
  once, and the consequence for a reader — *a healthy ratio on a MIXED-POLICY pair is not evidence
  its owned-page rows are being repaired, it is evidence their failures stopped counting.* Filed as
  a CONTRIB into 333, not as a bug.

---

## 2026-08-20 ~10:15Z — I answered §4's "one code question" and it REFUTED the entry above. The content is in `rendered_html`; no `content_data` route reaches it

**The entry above is wrong in its central recommendation and I have corrected it in place in
`bugs_open/277` (§5) and in live config (`499`).** Recording the sequence here because the way it
came apart is more useful than the conclusion.

### 1. The question that did it

§4 above listed the remaining unknown as *"can a producer file a `section_edit` item for a
`literal_markdown`-shaped finding?"* Chasing that answered itself quickly and encouragingly:

- **`ApplySectionEditInputSpec.ConfigKeys` contains `strip_literal_markdown`** — the route has a
  purpose-built setting for exactly this defect;
- **it is LIVE and `true`** on `section-editor.apply_edit` (migration `474`), read from the live
  `agent_definitions` row, not the seed;
- **an empty `field_updates` is a legal payload.** `applyContentEdit` takes the merge branch on any
  non-nil value, iterates zero times, and **then** strips — so `{"edit_type":"content_edit",
  "page_name":X,"slot_name":Y,"field_updates":{}}` is a pure mechanical repair with no LLM. The 229
  existing `section_edit` items have exactly that shape, and 12 runs in production show
  `content_edit_mode: field_updates, updated_field_count: 0`.

**Everything pointed the right way.** Then I read the items' own `findings` array, which I had never
done in three days of measuring them.

### 2. What the findings say, and why it ends the route

**All 7: `source: rendered_html`, `pattern: code_span`, `slot: ported-page`, `field:` empty.**
The matches are backticked code tokens in ported technical prose — `` `fetch()` ``,
`` `feTurbulence` ``, `` `ease-in-out` ``, `` `33%` ``. The detector scans **both** surfaces
(`literalMarkdownFinding.Source` is `content_data | rendered_html`); these fired on the second.

And the `ported-page` component's `content_data` is **215 bytes of provenance metadata** —
`{schema, sha256, source, qa_tier, generator}` — while its template's only field is `{{.body}}`,
which is **not a key**. So `StripLiteralMarkdownFromContentData` would strip a map holding no prose,
and `473`'s rerender would regenerate from nothing. **Both routes this lane has proposed are
inapplicable BY CONSTRUCTION.**

### 3. Measured with a control that could have come out the other way

Rendered the real template against the real `content_data`, production's own engine and option
(`text/template`, `Option("missingkey=zero")`, `call_agent.go:1171` (`executeGoTemplate`, reached via `component_library.go:1062`)):

| payload | rendered | body region | visible non-ws chars | err |
|---|---|---|---|---|
| owned `ported-page` (no `body`) | 4,665 B | 188 B | **0** | `<nil>` |
| generic control (has `body`) | 11,035 B | 6,568 B | full prose | `<nil>` |

**Same component, same template, two real payloads, opposite results.** After two days of running
checks that could only return one answer, this one is the shape I should have been building all
along — and note `err=<nil>` on both: the empty render *succeeds*.

### 4. The census that explains three days of confusion

`Ported Page (webdesign.co.uk)`: **115 instances, 100 missing the `body` key.** The split is
**total** — all 100 missing-`body` are `owned` (webdesign.co.uk 97, loancash.co.uk 3); all 15 that
have it are `generic`.

> **So "owned pages cannot be repaired" was the wrong diagnosis all along.** Ownership and
> un-regenerability *coincide* on this component. The ownership guard therefore takes the blame,
> and every fix anyone proposes — including both of mine — is about routing around it. **Route
> around it successfully and you still repair nothing.** Ownership is the correlate; the operative
> property is whether the content is reachable from `content_data`. Now a `LANDMINES.md` entry.

### 5. What I did NOT claim, and nearly did

**"100 pages are one edit from being blanked" is FALSE.** `apply_section_edit` calls
`enforceSingleSlotFloors` → `evaluateSectionShrink` on the **VISIBLE**-text axis (style and script
content excluded, `minShrinkGuardVisibleChars` = 200 on the existing side). Thousands existing
against **zero** incoming → refused, *"nothing was written and the existing component still
stands."* The realistic outcome is a **third refusal mode**, not damage. I wrote the alarming
version first and checked it before it left the scratchpad; the guard is real and calibrated on
exactly this failure (`bugs_closed/285`).

### 6. Where clause 1 actually stands — RETRACTED to where it was

The entry above said the blocker had narrowed to "one code question". **It has not moved.** A
`rendered_html`-only defect on a component with no regenerable source has **no route**, and the
08-19 evening claim was right about *these rows* even while too strong as a general law. Repairing
them needs an **HTML-level transform on `rendered_html`** (`` `x` `` → `<code>x</code>`), which
nothing does — a different shape from every route considered. Whether that is worth building for 7
findings of the mildest pattern is the owner's call and not an obvious yes.

**What survives and is worth keeping:** `section_edit → section-editor` really is 36/1 on owned
pages, really is the right route where `content_data` can fill the template, and `473`'s own header
already said the owned-page refusal was by construction. None of that was wasted; it was aimed at
the wrong property.

### 7. Live config corrected a THIRD time in one day — and the shape of the fix changed

`497` fixed a dead pointer. `498` fixed stale figures. Both were fixed by writing a better **value**.
`499` is different: the value was *right* and its **application** was wrong, so the fix is a
**question the reader must answer about their own rows** — read the finding's `source`, then ask
whether `content_data` can reproduce `rendered_html`. **A named target is only ever right for the
population its author was looking at; a test travels.** That is the day's most durable output and it
came from being wrong three times in the same string.

### 8. The landmine VERIFIER caught a real error in my own entry, and the error is a nice shape

`landmines-verify-dispatch.sh` returned **NEEDS_HUMAN_REVIEW** on the new entry: 8 of 10 footprint
items resolved, and it named the two that did not — `missingkey=zero` (a string literal, unsearchable
by a symbol index) and **the cited line `component_library.go:861`**. I checked it rather than waving
it through, and **the citation was wrong.**

Line 861 is `missingBareFields`, a **scanning helper**; the real chain is
`RenderTemplate` → `RenderTemplateReportingMissing` → `component_library.go:1062` →
`executeGoTemplate` at **`call_agent.go:1171`**. The conclusion is untouched — the real renderer does
use `Option("missingkey=zero")`, so the measurement was made with production's actual option — but
the pointer sent readers to the wrong function.

> ⚠ **Why it was invisible: I grepped for the VALUE and the first hit had the right value for the
> wrong reason.** `missingkey=zero` appears in both the scanner and the renderer. Had the scanner
> used a different option I would have noticed instantly; because it agreed, nothing looked wrong.
> **A grep for a value cannot tell you that you are in the right function** — it can only tell you
> the value exists somewhere. This is the memory rule *cite the ARM, not the function* with the
> failure one level further out: I cited an arm, it was simply an arm of something else.
> Corrected in `bugs_open/277` §5.3, here, and in the `LANDMINES.md` footprint. It also stands
> uncorrected in `499`'s file header, which is committed and applied — forward-only, and the LIVE
> config string carries no line number, so there is no live impact.

**And the process point, since this lane keeps asking whether these mechanisms pay:** a
NEEDS_HUMAN_REVIEW verdict that names exactly which two claims it could not check is worth more than
a green one. It cost one `sed` to act on and it found a real defect.

---

## 2026-08-20 ~17:30Z — the owner ruled YES on the 7, and the route is BUILT (code committed-pending, config LIVE, canary owed)

**Owner (chat, this session):** *"Do those seven findings get a repair route? Building one means a
transform that edits finished HTML directly - I think yes."* Everything below implements that.

**What was built** (design + reasons in `PLAN_2026-08-15_...md`'s 2026-08-20 addendum; register `CQ-028`):
- `datahelpers.ConvertLiteralCodeSpansInHTML` — `` `x` `` → `<code>x</code>`, tokenizer byte-splice,
  detector's own skip set (same package, same `nonAssertionElements` map). Conversion pattern is
  MDCodeSpanRe with `<`/`>` also excluded: **conversion ⊆ detection**, property-tested.
- `apply_section_edit` edit_type `rendered_html_transform` — opt-in (`allow_rendered_html_transform`,
  default OFF), floors + regulated guard wired, HTML-only persist (nil content_data → the
  pre-existing html-only UPDATE branch). Every refusal is an ERROR into the attempt ladder,
  including `converted==0` (no no-op deploys reported as edits).
- `datahelpers.ContentDataCanFillTemplate` — migration 499's test as code, coarse TOWARD "can fill".
- `check_literal_markdown.transformRouteSlot` — routes a page to section-editor IFF all findings are
  rendered_html-source code_span on ONE once-occurring slot that cannot regenerate. 11 refusal
  directions unit-tested, each landing on today's route.
- Migration `513` (+ROLLBACK): flag + `input_fields += transform_name` on section-editor's apply_edit.
  **`input_fields` is ExtractActionInputs' Strategy-1 WHITELIST (`action_inputs.go:831`) — checked
  before writing the migration; without that entry the input is silently never extracted.**

**Measured/proven this session:**
- Round-trip: 513 apply → rollback → `md5(default_config)` = `fdb8cb4d…` **identical to pre-image**;
  then applied for real → `b6076c7d…`, flag `true` + `transform_name` whitelisted, read at the column.
- Mutation runs on the transform: (1) reorder `Raw()`-write after `TagName()` → the MIXED-CASE skip
  test fails — **real landmine found building this: `Tokenizer.TagName()` lower-cases the tag bytes
  IN PLACE in the buffer `Raw()` aliases** (x/net/html `escape.go lower`), so an all-lowercase test
  suite would never see it; (2) skip disabled → 3 tests fail incl. the REAL fixture's
  `` `left 1s ${css}` `` JS template literal (testdata = tool-cubic-bezier's actual 8,743 bytes,
  which carry both the live finding and that adversarial literal — a composed fixture would not).
- RFC_022 parity test fired on the 7th optional key exactly as designed; `check.py` literal bumped;
  **overlay re-apply still owed** (cluster keeps the old literal until `apply -k` on
  optional-key-budget-check).
- Promoter interaction measured, not assumed: 444's doors are `EXISTS(≥1 complete per pair)` AND
  `(sample<5 OR ≥25%)`. New pair `literal_markdown → section-editor` = 0 completes → **HELD until one
  canary completes.** All 7 current rows are terminal (wont_fix wave), so the detector's next sweep
  files fresh items at the new shape (dedup index excludes all their statuses — checked).

**Council:** submitted, corr `b72a4029-f925-48bc-81d6-1552b7d25099` (`submission_277_rendered_html_transform.json`).
Trigger schema pushed back twice before accepting — `operation` enum has no `create` (a new file is
`add`), and a sketch whose every line starts `#` is refused as comment-only. Committed with
`Council-Submitted:`, verdict to be read later (~30 min budget).

**Missteps this session, with their checks:**
- The first cut of the transform called `TagName()` before writing `Raw()` — caught by READING the
  tokenizer's source before trusting the API (the check: when two accessors share a buffer, grep the
  package source for in-place mutation before assuming Raw is raw).
- My scratchpad backup path for the mutation runs didn't exist (`cp A B 2>/dev/null || cp A C` — the
  first cp succeeded so C was never made, then a later restore read C). Harmless here because the
  /tmp copy existed; the check: after any `||`-fallback copy, `test -f` the path you intend to
  restore from BEFORE mutating.
- Council submission: I combined three files into one edit entry (refused: ONE EDIT = ONE FILE) —
  the client-side validator caught both this and the enum; cost three cheap round-trips, no credits.

### ~18:20Z — council round 1: REVISE (gating: editquality HIGH on edit 1), and the objection was RIGHT about the sketch

The gating objection read my sketch's regex as unanchored — because I quoted it **using backticks
as delimiters, which swallowed the pattern's own backtick anchors**. The code was never wrong; the
sketch was genuinely misleading, and a reviewer implementing from it would have built the
destructive version. The r2 sketch carries the verbatim source line with the anchors spelled out,
plus the two tests that pin them. **Check for next time: never quote a regex whose own syntax
includes your quoting character — spell the delimiter.**

The round also earned four real improvements (commit `25d00cfe9`): the
`ContentDataCanFillTemplate` ↔ `missingBareFields` cross-references (bug_historian was right that
two uncoordinated detectors for the missingkey=zero family is how they diverge — the import-cycle
constraint is now stated in both), the three-HTML-instruments record in the transform header
(reuse_agent asked for the shared-utility search to be run, and running it produced a better
header), the transform-registry accumulation rule at the switch + CQ-028 (architecture's point:
one key, N transforms, invisible to WFA-013's counter), and the RUNBOOK's post-roll binary needle
probe (debug_historian: the config DO/RAISE proves nothing about the binary).

Two objections were answered by checks the reviewers had themselves already run: section-editor
has exactly ONE active row (their query), and their 104-completes count corroborates the pair
history. Round 2 resubmitted on the same corr (`submission_277_rendered_html_transform_r2.json`,
8 edits, 14 grounded_in). Monitor armed for the report.

### ~18:35Z — round 2: **APPROVED** ("3 advisory objections — none high-severity", 3 abstained)

Verdict READ (diagnosis_artifacts council_report 18:26:44Z). Acted on the advisories that named real
work rather than banking them:
- **editquality medium** (the "structurally unable to write content_data" claim rested on one
  helper branch with no test): `section_editor_html_only_persist_test.go` now pins it — a capturing
  QueryMatcher asserts the nil-branch statement does not NAME content_data (arity alone would miss
  an inline `= NULL`), with a positive control, **mutation-proven** (adding `content_data = NULL`
  to the nil branch fails exactly that test; restore verified byte-clean vs HEAD).
- **prior_art_librarian low** (novelty overstated): CQ-028 rescoped to "first markup-INSERTION
  repair" — the dead-link repair is standing precedent for rewriting. **debug_historian low**: the
  entry now names it a third MODE (edit the output when nothing can regenerate). **render_guardian
  medium**: the by-design divergence note added — digest stamped same-statement, audit tooling must
  read ported-population divergence as expected.
- **guardian medium** (metadata): a migration edit should be `operation: config_change`, not `add`
  — noted for future submissions; no code consequence this round.
- Banked as stated risks, no action owed: bug_historian's missingkey=zero lineage (the routing
  helper now documents it), guardian's new-divergence-class and small-architecture-shape flags
  (both explicitly "not blocking given the gating"), architecture's registry-growth advisory
  (rule stands at the switch).

**The trail: REVISE (r1, sketch defect) → APPROVED (r2).** Both code commits already carry
`Council-Submitted:` and are credited automatically by 098; this round's follow-up carries
`Council-Reviewed:` — the verdict is read, so the strong trailer is now honest.

---

## 2026-08-21 (session "bugfix 083") — clause 1 REPAIRED at the served bytes, and the sweep it needed was FOUR DAYS AWAY, not "next"

### 1. The roll HAD happened; proven per service, with both controls

Chassis pods (`agent-chassis-68ff4d794c-{46xhw,xg5mb}`, one digest `sha256:3ed50651…`, started
2026-08-20 19:51Z) stamp `buildinfo.GitCommit=0483e7f4e410168d30ed2d86dcbb820d8c28a383`.
`git merge-base --is-ancestor af0f00bb5 0483e7f4e` → **YES** (and `6011f9657` too), so the whole
CQ-028 change is aboard. Binary probe [MEASURED 2026-08-21 ~12:0xZ]: `rendered_html_transform` 8,
`code_span_to_code_tag` 5, `OWNED_PAGE_GUARD` 3 (positive control), `ZZQQ_NEEDLE_THAT_MUST_NOT_EXIST`
0 (negative control). Config half at the live column: `allow_rendered_html_transform=true`,
`input_fields @> ["transform_name"]` → `t`.

⚠ **The `build provenance` startup line was NOT readable** — a chassis pod's log history is ~90
seconds under load, and `--tail=100000` returned 1.9MB of orchestration state without it. The three
"matches" my grep did find were the *phrase* quoted inside a council submission's collected_data.
The `buildinfo.GitCommit=` probe of `/proc/1/exe` is what actually answered it.

### 2. THE REAL FINDING — §2 step 2 ("the detector's next sweep") was ~4 days out, and nothing said so

`literal_markdown` is a **quality**-discovery check, and the only thing that drives it is
`site-discovery-rotation-quality` (mig 346, SCH-025). Its `pre_query` gates on
`COALESCE(r.last_selected_at,'-infinity') < now() - interval '7 days'`, `LIMIT 1`. [MEASURED
2026-08-21 12:1xZ] webdesign.co.uk's stamp is **2026-08-18 07:23:15Z**, and *the whole rotation was
idle*: the oldest stamp in the fleet (robot-hands.com) was at **5d 01h**, so no site was due at all
until 2026-08-23 10:16Z. Walking the 3h tick order forward, webdesign.co.uk's next natural quality
sweep was **≈2026-08-25 07:33Z** [CALCULATED from measured stamps] — the same day as `083`'s door
soak, by coincidence, and four days after the handoff implied "next".

**The general shape, and it is not specific to this check:** *a fix to a discovery check is not
observable until that site's rotation slot comes round, and the rotation goes IDLE — not "slow" —
whenever no site has aged past the floor.* `last_triggered_at` on the task keeps advancing every 3h
throughout, so the task looks healthy while examining nothing. Filed as a LANDMINE.

### 3. Forced the sweep (owner said fire it), and it did not consume the rotation stamp

Seeded `oneshot-quality-discovery-wdcouk-20260821` on the precedent shape of
`oneshot-quality-discovery-wdcouk-20260810` (same agent, same topic, `input_data` = domain +
site_id, **no `pre_query`**, `interval_seconds` 86400), fired at **13:19:01Z**, disabled at
13:19:5xZ. Because a one-shot carries no `pre_query`, it never touches `site_discovery_rotation` —
so webdesign.co.uk's 08-25 slot is still there, unspent. Preconditions checked first, including the
rotation's own courtesy gate (`0` claimed build items on the site).

**A trigger stamp is not a run** (`bugfix_230_discovery_driver`'s five-stamps CONTRIB). Verified the
run happened: `orchestration_states` `4cfdca1f-…`, quality-discovery-agent, webdesign.co.uk,
COMPLETED, 13:19:01.24Z → 13:19:15Z.

### 4. The router did exactly what it was built to do — 8 rows, ALL new-shape

Everything the sweep filed, fleet-wide, was `literal_markdown → section-editor` with
`edit_type=rendered_html_transform`, `transform_name=code_span_to_code_tag`: the 6 `tool-*` pages,
`learn-design-physics-of-ui`, and `learn-index`. **No other check filed anything** (dedup held the
rest). The 7 old `wont_fix` rows did not block re-filing — `idx_swi_dedup` excludes `wont_fix` AND
`unresolved`, checked at the index definition, not assumed.

### 5. The canary — one row, complete in 3m21s, and PROVEN at the served bytes

Promoted **one** row by hand (`ecd947c2…`, tool-cubic-bezier, the `` `ease-in-out` `` finding) with
444's own promote UPDATE at 13:21:42Z. claimed 13:24:18 → section-editor run `8ddb55b0…` →
**complete 13:25:03Z**, `result._verification` = *"verified — no literal markdown on either surface
across 1 component(s)"*.

Served bytes, cache-busted, before → after [MEASURED 13:22Z / 13:25Z]:

| check | before | after |
|---|---|---|
| backticks on the page | 6 | **4** (−2, exactly one span) |
| `` `ease-in-out` `` | 1 | **0** |
| `<code>` tags | 0 | **1** |
| `<code>ease-in-out</code>` | 0 | **1** |
| **backticks inside `<script>`** (adversarial control) | 4 | **4 — untouched** |
| page bytes | 16683 | 16694 (**+11** = `<code></code>` 13 − 2 backticks) |

`diff` of the two fetches: **one line pair changed, nothing else on the page.** The fixture's
adversarial case — a tool page whose own JS uses template literals — held in production, which is
the half that could not be proven from a test.

### 6. A DEFECT FOUND BY DOING IT — the two-strike rule makes a NEW route inherit the OLD route's strikes

1 of the 8 rows (`learn-index`, `2c4033b0…`) was **born `unresolved`** — terminal, never
dispatchable — carrying the summary *"[unresolved after 2 attempts]"*. Cause is
`writeWorkItem`'s two-strike rule (`load_work_item_actions.go:1373-1408`): it counts
`status IN ('complete','failed')` rows for the same `(site_id, item_key)` in the last 7 days, and at
≥2 births the new row `unresolved`.

Evaluated the predicate exactly as the code does: **`terminal_count_7d = 2`** — `46f356cf` (failed,
**page-build-handler**, 08-14) and `6865c4b9` (failed, **page-rerender**, 08-18). **Both are OLD
routes**, and this lane has already established that both are inapplicable to this class by
construction. So the first-ever attempt by the route that CAN repair it was counted as the third
attempt of routes that never could, and the label says "we tried twice" about a repair that has
been tried zero times.

**The general form:** `item_key` is handler-agnostic **by design** (that is what makes dedup work),
so the strike count is handler-agnostic too — **a re-route inherits the strikes of the route it
replaced.** `bugs_open/333` §"the finding terminates wont_fix" already names the two-strike rule
reaching a false "we tried twice", but by a different road (refusal loops on owned pages). This is
the *route-change* road and it needs no refusal at all. CONTRIB filed into 333.

**It self-heals here, and the arithmetic is worth stating rather than acting on:** the window is
rolling 7 days, so `46f356cf` (08-14 16:03) ages out at 08-21 16:03 today and `6865c4b9` (08-18
07:23) at 08-25 07:23 — just *before* the natural 08-25 07:33 sweep. So learn-index's next filing
should be born `detected`. **Do not hand-flip the row to test that** — the prediction is
disconfirmable at the next sweep and worth more than the one page.

### 7. Final tally, 13:37Z — 7 of 7 dispatchable rows complete, all `verified`, all clean at the served bytes

`tool-cubic-bezier` 13:25:04 (the canary) · `tool-grid-generator` 13:30:57 · `tool-head-architect`
13:31:45 · `tool-json-cleaner` 13:32:28 · `tool-noise-generator` 13:33:09 ·
`learn-design-physics-of-ui` 13:33:55 · `tool-text-extractor` 13:37:02. Eighth row = `learn-index`,
born `unresolved` (§6).

Independent check at the artefact on all seven, cache-busted, **not** trusting the verifier's status
[MEASURED 13:34–13:38Z] — prose backticks (total minus those inside `<script>`) and `<code>` tags:

| page | prose backticks | script backticks | `<code>` |
|---|---|---|---|
| tool-cubic-bezier | 0 | 4 | 1 |
| tool-grid-generator | 0 | 8 | 3 |
| tool-head-architect | 0 | **44** | 2 |
| tool-json-cleaner | 0 | 2 | 2 |
| tool-noise-generator | 0 | 6 | 2 |
| tool-text-extractor | 0 | 2 | 2 |
| learn-design-physics-of-ui | 0 | 0 | 1 |

`tool-head-architect` is the strongest single result: **44** backticks live inside that page's own
`<script>` and every one survived, while its prose reached zero. A transform that leaked into script
context could not produce that row.

### 8. ~14:00–14:45Z — the owed `diagnosis_guardian` message became two migrations, because the seat's own text is where the discipline lives

The handoff had carried *"tell the `diagnosis_guardian` seat its `error_step` discipline is
INVERTED"* since 08-19 as the largest un-started item a session could do alone. There is no inbox
for a seat: its `prompt_template` **is** the discipline, so "telling it" is a config change.

**Re-verified at source first** (`coordinator.go:3666-3679`, HEAD `91cd28919`) rather than trusting
08-19b's account: `routeToErrorStepOrFail` checks `step.ErrorStep` **first** — comment *"Check
step-level first (parallel to NextStep) — preferred location"* — then falls back to
`step.Config["error_step"]` *"for backward compatibility"*. The seat asserted the exact reverse and
judged on it (clause `(d)`).

**The second defect was found by reading both rosters side by side, and it is not prose drift.**
council-gate's copy read `## The author's stated rationale loop's load-bearing disciplines`.
`099_SYNC_gate_roster.py:85` did an unanchored
`p.replace("## The diagnosis", "## The author's stated rationale")` — meant for the diagnosis
CONTEXT block, and it also hits any longer heading beginning with those words. So the head of the
list of disciplines that seat defends had been nonsense for as long as the mirror has run.
**Fixed in three places, deliberately:** the live gate text (531), the live fix-proposer text is
unaffected (it was never mangled), and the script's substitution is now
`re.sub(r"^## The diagnosis[ \t]*$", …, flags=re.M)` so it cannot recur.

**What was exercised before applying — three ways, per the RUNBOOK, plus a round trip:**

| | 530 (fix-proposer) | 531 (council-gate) |
|---|---|---|
| whole file, COMMIT→ROLLBACK | guards pass, 1 row, NOTICE | same |
| anchor pre-broken | aborts *"does not occur exactly once"* | aborts on the heading anchor |
| the negative control's own needle PLANTED elsewhere | aborts *"`silently inert` still appears"* | — |
| 377 fragmentation induced (text prepended) | n/a | aborts *"377 shared prefix FRAGMENTED — 2 distinct prefixes across 17 marked seats"* |
| round trip with `_ROLLBACK` | `cbfcc981… len=4248` restored exactly | `191a7bbc… len=4301` restored exactly |
| live md5 after apply | `99bf2e45…` | `347a20cf…` |

⚠ **One guard in 531 is structurally unfirable and the header says so.** The marker-offset check
(`position(marker in v_new) <> v_mark_old`) cannot fail for this file's anchors — they are all
downstream of the breakpoint, and an upstream anchor would trip the occurs-once check first. It is
kept for the next author, but it is **not** a control that passed; the fragmentation check is the
one that can fail, and it did when induced. Writing "both 377 guards passed" would have been the
`[MEASURED]`-but-undisconfirmable shape this lane keeps catching in itself.

Council: `Council-Submitted: c00fbfd8-c459-4e8a-ac04-0997aca98477` on commit `086f3af35`
(3 edits, 6 `grounded_in`). ⚠ **The seat under repair is one of the seats that will review this**,
which is fine — it reviews on the text it holds at run time, and that text is now correct.
`DRY_RUN=1` refused the first draft: `plan.risks` must be a **string**, not an array.

### 9. Sharpening §6's prediction — the strike clock and the rotation clock are THE SAME CLOCK, which says when the two-strike rule can and cannot bite

Checked the arithmetic rather than leaving "it ages out" as a feeling. The two-strike window is
`created_at > now() - interval '7 days'`; the rotation's eligibility is
`last_selected_at < now() - interval '7 days'`. For a strike filed **by a rotation sweep**, the row's
`created_at` is the stamp plus a second or two — so the strike expires about a second *before* the
same site's next sweep can even be selected, and always before the filing itself. **A finding whose
only failures came from rotation-filed rows can therefore never be born `unresolved`.**

`learn-index` was born dead because its strikes did **not** both come from the rotation: `46f356cf`
(08-14 16:03, page-build-handler) sits between rotation sweeps, and its key carries rows on 08-10,
08-11 ×2 and 08-14 — a cadence no 7-day rotation produces. **So the rule bites exactly when a
SECOND producer files the same `item_key` off-rotation**, which is `bugs_open/333`'s territory
(producers routing the same finding by different paths), not a property of the re-route alone.

Two consequences worth carrying: the 08-25 prediction in §6 is **stronger** than I first wrote (both
strikes are out of window, one of them by construction, not by luck), and a fix for the rule that
merely lengthened or shortened the window would miss the case — what distinguishes these rows is
**who filed them**, which is the thing the count does not record.
