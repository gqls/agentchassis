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
