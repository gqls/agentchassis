# NOTES — bugs_open/281 tool audit by instance (append-only, newest at the bottom)

## 2026-08-15 — pick, verify, research

- Picked 281 (`who-owns` → "(none identified)"). Live-transcript symbol grep found one session
  on `check_tool_health` — verified it was the 122 lane that FILED 281 (wrote the file 14:51Z,
  last activity 14:55Z). Re-ran the 30-min recency grep before first Write: 2 hits, both
  listing/LANDMINES noise.
- Bug valid at tree and live: `check_tool_health.go:68` `component_level='tool'`; live
  tool-auditor `load_tool` = seed 088 shape (`WHERE cc.id=$1 … LIMIT 1`, params component only).
- Census `[MEASURED]`: webdesign tool pages = 63 ported-page (section) + 4 tool-level;
  ported-page component `a7daa5c5…` = 115 instances / 2 sites; 0 tool PLANs
  (`doc_notes subject_type='tool' AND categories ? 'acceptance_criteria'`), 89 `needs_criteria`;
  improve_tool fleet: 196 unresolved / 45 complete; clause (b) matches 62 of 63 ported pages —
  the 63rd (`tool-ab-test-calculator`) carries a fork AND a ported instance, counted via (a).
- Three Explore agents' reports (audit machinery / ported+decomposition / governance) are
  summarised in the plan file; the load-bearing findings are restated in TL-042.

## 2026-08-15 — review pass (owner: "check over once more") — what it changed

- **Missed producer.** `check_tool_acceptance.go` already files improve_tool→tool-improver for
  ported instances. `component_versions` on the shared component: v1 (08-05, pre-edit
  snapshot = the 77-char passthrough) and v3 (08-14 18:48Z) both `changed_by=update_component_html`;
  the 08-14 trigger was `tool_acceptance:asset-formatter:<webdesign>` (complete). Bug 281's
  "none exist" is true only of `audit_tool` items. → edit 9 (gate that producer), and the new
  type is named for both.
- **D4 threshold from census** `[MEASURED]`: `update_component_html` has ONE live consumer
  (tool-improver). Multi-placed: 50 section-level (max 299 pages), 2 tool-level (2 pages /
  2 sites; `tool-llm-cost-calculator` rewritten 5× by tool-improver). Both incidents
  section-level → refuse `component_level<>'tool' AND >1 page`; WARN for multi-site forks.
- **section_edit not a viable per-instance fixer today**: owner-gate item spec is
  `{edit_type, page_name, slot_name, field_updates:{}}` — no instruction payload; ported-slot
  section_edit: 1 complete, 2 triaged.
- **Live latent hazard.** Shared template now 8,864 chars of asset-formatter markup
  (`{{.body}}` still present); all 115 instances `build_status='pending'` since 18:48:38Z.
  Verified NOT propagated: every `pc.updated_at` == 18:48:38.131 (the bulk flip), loancash
  `rendered_html` still loan content, pages deployed 08-12; no `component_template_corrupted`
  item since 08-14. Not repaired here (another lane's artefact); recorded in 281 + TL-042.
- **MISSTEP (mine).** First propagation census read "15 loancash instances carry the new
  template" from `rendered_html LIKE '%ported-page-section%'` — a class-name match with no
  control. The very next query (timestamps + a content check) refuted it. Cheap check that
  would have caught it: pair any "did X propagate" LIKE with a positive control on a row known
  untouched. → WRONG_CALLS.md.

## 2026-08-15 — implementation

- Go: `check_tool_health.go` rewritten on `toolEligibilityWhere`/`toolSubjectKeyExpr`
  (instance identity, split-scope cooldowns, `PageID` set, template checks gated by
  `templateIsFork`, ported branch → `ported_tool_fix`, Tier-2 cap 12); `check_tool_acceptance.go`
  ported branch → `ported_tool_fix`, cooldown split-scope; `tool_eligibility.go` header
  rewritten; `update_component_html_action.go` step 1b fence. `escapeJSON` had to stay — two
  sibling checks use it (build told me).
- Tests: `check_tool_health_test.go` (4 tests, incl. the Mind-Map-shaped fixture);
  `update_component_html_shared_fence_test.go` (5 sqlmock tests; **mutation-proven** — with the
  fence condition disabled the refusal test FAILS, restored → green); `verifier_coverage_test.go`
  gap entry + `liveItemTypes`. Whole `discovery_checks` package green; `go build ./platform/...` ok.
- Seeds 425/426: dry-run in aborted txn → pre-flight OK, post-condition OK. **Round-trip
  (apply body + ROLLBACK body, compare to `_before`) caught a real defect in both rollback
  files**: `to_jsonb('literal')` on an untyped string → "could not determine polymorphic type";
  fixed with `::text`; round-trip now `t` for both.
- Register: TL-042 written, TL-033 supersession note, index row.

## 2026-08-15 — commit, seeds applied, council

- Committed `25f92a967` (20 files, pathspec) with `Council-Submitted: 360ae540-8b64-41f9-94da-d7c316183398`;
  gofmt follow-up `a41d11e30`. WRONG_CALLS entry for this lane rode `e96055a03` (bugs_open/282
  session) as a same-file passenger — nothing lost, they flagged it in their message.
- Seeds 425/426 APPLIED 17:17Z (pre-flight re-proved 0 open items without spec.page_id at apply
  time); live rows verified by reading them back (params, gate condition, prompt needle, slot
  path); ledger rows present. Hand-ran the live load_tool query on the shared component with two
  different page_ids → two different tools (mind-map 18,242 chars / asset-formatter 9,222),
  `source_html == rendered_html` for a ported instance, display_name = subject key. The pin works.
- **Council round 1 died `complete_invalid` at `persist_submission` — my schema slip**: four edit
  `file` fields named two paths / carried whitespace ("a.go + b.go", "(+_ROLLBACK)"); the
  validator wants exactly one repo-relative path. Read `__step_error.failed_step` FIRST (RUNBOOK
  says so; it was the plan, not a seat). Fixed the four fields, moved companions into rationale,
  resubmitted under `RESUBMIT_CORR` (run orch `5464d7fd…`). Cheap check for next time: assert
  `re.fullmatch(r"[A-Za-z0-9_./-]+", edit["file"])` before firing — added to the build script.
- Advisory on the commit: "migration + platform code in one commit — needs a staged rollout
  order". Considered: config is live on apply, Go rides the roll; the seed headers state the
  ordering and why it is safe in either order (forks already carry spec.page_id; the fence rides
  the same image as the widening). Additive/opt-in shape → normal council scope per RFC_010 §1 /
  RFC_022, stated in the submission rationale.

## 2026-08-15 — council verdict and follow-up

- **APPROVED** round 2 (`360ae540…`, run `5464d7fd…`, ~13 min in the seats): "approved with 6
  advisory objection(s) — none high-severity", 4 abstained, `unreadable` NULL. Acted on rather
  than defended (a REVISE round is cheaper than the defect it finds; here the seats' points were
  cheap and two were real):
  - bug_historian: "is there a SECOND writer of html_template bypassing the fence?" — enumerated:
    six actions write it. `fix_component_template_action.go` is the one other PAGE-aware writer
    (takes `page_component_id`, reads the page's `rendered_html`, writes the component template)
    — NOT fenced (its shared write is sometimes the intended repair — it restored the wrapper
    after 08-05); recorded open in the guard file + TL-042. Four others take the component as
    subject (fan-out intended).
  - reuse_agent: `component_write_guard.go` is the home for component write guards — fence
    moved there as `sharedComponentWriteCheck` (+ writer census in the header); the action calls it.
  - guardian: census on every call + fail-closed = new fleet-wide failure mode — narrowed:
    fail-closed only when `component_level<>'tool'`; a tool fork's census error warns and
    proceeds (+test; the helper's decision line mutation-proven again).
  - guardian (double-active-row landmine) and debug_historian (needle-gate discipline): both
    already covered by the seeds' pre-flight (`target_count <> 1 → RAISE`) and gated UPDATEs —
    the seats saw truncated sketches. Answered in the RUNBOOK.
  - debug_historian: deploy-verification recipe missing — RUNBOOK gains the build-provenance +
    `git merge-base --is-ancestor 25f92a967` recipe (per CLAUDE.md, not a symbol grep).
  - prior_art: '0 PLANs' and 'one consumer' asserted — attached as queries in the RUNBOOK
    (the LIKE over `default_config::text` covers nested sub_workflows).
  - architecture: TL-042 now states what would close the `ported_tool_fix` sink — (a) decompose
    to a `component_level='tool'` fork, or (b) a per-instance fixer + instance-keyed PLAN — and
    the count query to tell whether the gap is closing.
  - editquality (low): per-check dedup key is intentional (a tool failing both checks is two
    findings). Not changed.
- Follow-up commit `d7b2d9994` with `Council-Reviewed: 360ae540-8b64-41f9-94da-d7c316183398`.
- STATE: fixed at source; seeds 425/426 live; Go (`25f92a967`, `a41d11e30`, `d7b2d9994`) rides
  the next chassis roll. 281 stays OPEN until the roll + first-sweep census (RUNBOOK).

## 2026-08-16 — the roll, and three corrections to my own record

- **Roll verified at the artefact:** pods `v1.0.1303` (started 2026-08-15 18:45Z). Provenance line
  had scrolled and the pod logs had rotated (start 09:35Z today), so: binary probe. The stamp is
  ONE full sha (the build HEAD) — my own shas returned 0, which is expected, not absent. Found the
  build HEAD by probing candidate commits in the window (fixed-string grep, `-F`, per sha; a regex
  grep over `/proc/1/exe` timed out): `5e075a6f9` (count 3), fake-sha control 0. `git merge-base
  --is-ancestor` → all three 281 commits ARE in the running image. (The makefile's 1303 bump is
  uncommitted; HEAD says 1299 — don't look for the tag in git.)
- **Fence seen refusing on the live image:** `agent_error_log` `component_write_shared_blocked`,
  2026-08-16 09:59:06Z, step `induce_write`, "Ported Page (webdesign.co.uk)" 115 pages / 2 sites —
  the 285 lane's induced test (their close step 1). Not repeated.
- **CORRECTIONS (mine, all applied visibly today — register, RUNBOOK, PLAN, bug file, webdesign
  NOTES, guard-file census, WRONG_CALLS):**
  1. "not propagated" was FALSE for `learn-ai-builders-content-first` (served ~23.5 h via the
     improver's arbitrary delivery target; the 285 lane found + restored it, seed 431). The single
     `deployed`+updated row in my own census was it; I named it away.
  2. "0 tool PLANs" counted `doc_notes`; PLANs are in `doc_plans` (143; 14 of the ported 63). The
     routing stands on "no per-instance writeback", not on a PLAN count.
  3. `fix_component_template` is component-scoped mechanical repair, not a page-aware LLM
     rewriter — census wording corrected in `component_write_guard.go`.
- **State of the shared component `[MEASURED 2026-08-16]`:** template 4,664 chars with
  `{{.body}}` (v3 content restored by the 285 lane; poison banked as v4); 0 of 115 pending;
  poison page restored (3,781 chars, no `portedPageAssetList`).
- **tool_health has NOT run since the roll** (last item 2026-08-15 14:51Z; 0 `ported_tool_fix`
  rows fleet-wide). The first-sweep census needs a design-discovery pass on webdesign — next.

## 2026-08-16 — first sweep on the live image (design-discovery fired by hand, corr `172ef9b3…`, 10:07:52Z → COMPLETED 10:09Z)

- Rotation had last visited webdesign 2026-08-09, so I fired `082_fire_design_discovery_any_site.sh`
  (live `run_checks.config.checks` includes `tool_health` + `tool_acceptance` — read from the row).
  Run: 44 items inserted, 3 skipped, `checks_failed []`.
- **Census `[MEASURED 10:12Z]`:** `ported_tool_fix` **13** (all `needs_human_review`, handler NULL,
  page_id set, 13 distinct keys == rows; all on the shared ported component; findings: 9× no @media,
  3× fetch(), 1× CDN — all `low`); `audit_tool` **12** = the per-run cap, alphabetical
  `animated-favicon … css-variables` (12 distinct keys, page_id set); `improve_tool` **0** — the 4
  forks all had items inside 7 days (cooldown working, unchanged semantics). Negative control:
  **0** new tool items on non-tool pages, demand control 34 non-tool ported pages exist. Register
  the census verdict: coverage widened as designed; instance identity holds.
- **The motivating case is NOT detected by the structural tier — and the bug file's premise for
  it was wrong.** `tool-mind-map` drew no `ported_tool_fix`; the bare-hex census on its
  `rendered_html` is **0** (2 hex strings anywhere, both non-colour): its styles are `var(--text-dim)`,
  `var(--surface)` … throughout. The illegibility the owner saw is the RESOLVED value of those
  variables (contrast — bug 122's ink-slot class), which no regex over the source can see. The 281
  file's "the existing `hardcoded_colors` check already looks for exactly this class of thing" is
  refuted at the artefact; my plan repeated it and I never ran the census I wrote as a step (the
  `>3 bare hex verified by regex census first` line) → WRONG_CALLS. Note the row was rewritten by
  hand at 18:06:05Z on 08-15 (owner-gate `section_edit` cancelled 18:06:38Z), so whether it was
  var-based BEFORE that edit is unknowable from the live row.
- What DOES reach the Mind Map's defect class: Tier-2 LLM audit (`audit_tool` — reaches `m…` in a
  later capped pass; **hand-filed one now, page-pinned, item `1bfd5d1e…` as the instance-pin
  proof — SAFE only because 425's pin is live**), `palette_contrast` (in the same discovery run —
  check what it filed for this page), and Tier-4 vision (bug 243, never executed for any tool).

## 2026-08-16 (later) — instance-pin proof, a pre-existing stall found, 281 CLOSED

- Hand-filed page-pinned `audit_tool` for the Mind Map (`1bfd5d1e…`, created_by
  `bugfix-281-instance-pin-proof`, `triaged`); claimed 10:11:53Z. Run `a6f7ac42…`: `load_tool` →
  `tool_data.page_id` = the item's page, `component_level=section`, `source_html` 18,608 = the row's
  `rendered_html`; llm_audit score 5/10, 17 findings (4 high: mouse-only drag/pan, hover-only node
  controls, `nodeLayer.prepend(svgLayer)` on every render, global `window.onmousemove` clobber); loop
  ran 10 iterations `check_target_class → create_review_item → done` — **0 improve_tool** (gate works
  live). PIN PROVEN.
- **But the run never left `create_items_loop_complete`** (10:16Z → roll at 10:41Z killed it →
  FAILED; the sweep re-dispatched at 10:54 and 11:37, both stalled the same way; item `failed`
  12:17Z). Compared against history: pre-425 tool-auditor runs show the identical shape (20 RUNNING /
  2 FAILED at that step) → PRE-EXISTING, not 425. Correlation: 43 runs with >10 findings (the
  `max_iterations` cap) stuck/failed vs 1 completing; 3 uncapped also stuck. Filed **`needs_diagnosis`
  `815322b9…`** (mechanism + pointers, no counts asserted). Only ONE `audit_review` row exists for
  the page — the per-page `item_key_suffix_field` collapses N findings to 1 (residual, noted).
- Roll #2 today: pods `v1.0.1304` at 10:41Z (`git merge-base` on the new stamp not re-run — the
  fix commits precede 1303's HEAD, so they are in 1304 too by construction).
- **281 CLOSED**: both mechanisms fixed+live+measured; close section appended; file `git mv`'d with
  both paths on the commit. Handoff written: `HANDOFF_2026-08-16_continue_here.md`.


## 2026-08-16 (later) — the 090 verdict came back REFUTED, and the real cause is a 2^N blow-up

- Read `815322b9…`: **REFUTED**, then `stopped_by: iteration-cap`, `status: UNVERIFIABLE`. It
  killed my hypothesis with one row — `ec046659…`, 14 findings vs `max_iterations` 10, i.e.
  truncated, sitting at `complete`. So truncation does NOT block the handoff. Its own revised
  lead (Kafka `failed to write message to kafka: context canceled`) I then checked and
  **discarded**: those `agent_error_log` rows carry `step_name = process_item_iter_N_call_handler`,
  which is the *parent* `build-dispatch-loop`'s loop, not `tool-auditor`'s `create_items_loop`;
  one row today against 31 dead ones. A verdict's next_scope is a lead, not a finding.
- **What I should have done first, and what actually found it:** compare the same-shaped loop
  consumers by payload size. `internal-linker` 22 kB avg, `tool-suggester` 447 kB,
  **`tool-auditor` 22 MB avg / 29 MB max**. Three orders of magnitude on the same machinery is
  not a subtle bug and it took one query.
- Per-key sizes then gave it away as an exact geometric series — `_iter_2_done` 70 kB, then
  141, 283, 567, 1134, 2269, 4538, **9076 kB** at `_iter_9_done`. Exactly 2.00x per iteration.
- **Mechanism, read in the data not inferred:** `create_items_loop_iter_3_done` has
  `results[0..9]` (all ten, though four iterations had run) and `results[0]`/`[1]`/`[2]` each
  contain a nested `done` key holding that iteration's own full aggregate. That IS the recursion.
- **In the code:** `handleLoopExpansion` injects the sub-workflow's `done` substep
  (`action: loop_complete`) once per iteration, without `total_iterations` or
  `substep_output_fields` in its config; `LoopCompleteAction` then falls back to
  `loop_metadata.total_iterations` (the whole loop) and, finding no Strategy-1 fields, falls
  through to the Strategy-3 generic scan that copies every `<loop>_iter_<i>_*` key — previous
  `_done` aggregates included.
- **The control, which is what makes this disconfirmable:** 3 of the 18 live loops have no
  `loop_complete` substep. `page-content-writer` has no `_iter_N_done` keys at all, flat 8 kB
  per-iteration outputs, one 32 kB outer aggregate. If the cause were anything else it would
  double too.
- Confirmed on a third type — **`build-dispatch-loop`, the fleet dispatcher**: 201 kB → 617 →
  1419 → 3023 → 6247 kB, avg 2.8 MB over 353 rows, max 13 MB. It is on the same cliff.
- Filed **`bugs_open/289_…`** (15 of 18 loops exposed; fix candidates ordered by what makes the
  bad state unrepresentable) and a fresh `090` on the correct mechanism, RUN_CORRELATION
  `12ffad7c-a7b2-4955-b531-554f07650598`. Separate defect noticed and NOT folded in:
  `complete_workflow … message validation failed` x6 — `ProduceWithValidation` validates
  headers, not size, so it is its own bug.


## 2026-08-17 — the fix is built, submitted and committed; inert until the roll

- Re-checked before touching anything: 66 commits landed overnight, HEAD moved, **the loop
  engine files were untouched** and no new bug file cites 289 — so nobody had taken it. My four
  08-16 commits are all ancestors of HEAD. (`git log -6` reads as though my work is missing —
  it is 67 commits back; ancestry is the check, not the log's first page.)
- **The bug is still live and still burning.** Since the `v1.0.1305` roll (08-16 22:07Z):
  25 `tool-auditor` runs, **1 COMPLETED**, 14 FAILED, the rest stuck at
  `create_items_loop_complete` / `_iter_9_done` with `awaited_requests = {}`.
- Second `090` (`12ffad7c`) finished **UNVERIFIABLE**, as predicted — bundle starvation again.
  Its iteration-1 REFUTED cited "no `_done` key appears at all" for `1a70fed3`; that row has
  **ten**, 20 kB → 10 MB. Both runs on this target refuted it on false absences.
- Fix committed `509e01e6a`: mark the injected terminal at expansion, return a marker instead
  of aggregating. Matched on **action**, never name (9 of 15 are called `done`). Registered
  WFA-015 in the same commit — it is a reserved key on a shared mechanism, which is the shape
  `bugs_closed/124` was vetoed for shipping unregistered.
- **Guard proven by mutation, not by a green tick:** disabling the early return makes the guard
  test fail and print `results[0].done.results[0].done.nested` — the recursion, reproduced in a
  unit test. The guard test's control strips only the terminal signals and demands the old
  swallowing behaviour back, so a function that merely never aggregates would not pass it.
- Blast radius measured rather than asserted, because the council would otherwise be asked to:
  all 15 `loop_complete` substeps declare `output_field = (none)` and `next_step = (none)`, so
  the per-iteration aggregate was addressable by nothing and read by nobody. That measurement
  also killed the cheap name-based candidate.
- Council `Council-Submitted: 7a3c4fb7-e8c1-4b5f-950e-7a826d5bebbe`, verdict pending.
- Two pattern-check advisories checked and dismissed with reasons (twin `LoopAction` has no
  aggregation path; the `:334` hit is a bare string literal). Noticed and NOT fixed:
  `LoopAction`'s `max_iterations` read is a bare `.(float64)` — inert today, same class as 193.


## 2026-08-17 11:10Z — the council round died on an ACCOUNT-WIDE Anthropic quota, not on my JSON

- Round `7a3c4fb7` reached `complete_invalid` at step `review_constitution`. **This was NOT a
  submission-schema failure** — the runbook's own warning applies, so I read
  `__step_error.failed_step` before touching the JSON, and the message is:
  `API request failed with status 400 … "You have reached your specified API usage limits.
  You will regain access on 2026-09-01 at 00:00 UTC."`
- **Measured, fleet-wide, not inferred from one seat:**
  - Last SUCCESSFUL LLM call anywhere: **11:08:03Z** (`council-gate`, `claude-sonnet-5`).
    First failure **11:08:37Z**. Every call since has failed.
  - Failures span **two agent types** (`council-gate`, `landmine-verifier`) and **two models**
    (`claude-sonnet-5`, `claude-opus-4-6`) → account-level, not model- or agent-scoped.
  - `agent_error_log` usage-limit rows: 7 in the 11:00Z hour, none between 08-14 16:00Z and
    today. Earlier spikes (08-14: 78, 08-10: 10) recovered; **this one names a fixed regain
    date**, which a transient rate-limit does not.
  - Provider concentration over 24 h: `claude-sonnet-5` 386 ok, `claude-sonnet-4-6` 50,
    `claude-opus-4-6` 42, `mistral-small3.1` **2**. So ~99.6% of fleet LLM work is Anthropic
    and there is no meaningful fallback — this is an effective full stop for every LLM-driven
    agent (auditors, writers, the diagnosis loop, the fix loop, the council gate).
- **Do NOT read my submission as the cause.** This is a monthly account cap and the fleet had
  already spent 478 calls in the preceding 24 h; my round contributed a handful. Stating it
  because the coincidence of timing invites exactly that wrong inference.
- **Casualties of mine:** the `7a3c4fb7` verdict (never rendered — an invalid run writes no
  artifacts, so polling `diagnosis_artifacts` by that corr would wait for ever), and both
  landmine-verifier runs armed yesterday. The `Council-Submitted:` trailer on `509e01e6a`
  stands and asserts nothing, and `098` will credit it automatically **if** the correlation is
  ever approved — which now cannot happen before the quota returns or is raised.
- **Nothing about the 289 fix changes.** It is committed and rides the next roll regardless;
  the gate is advisory and cannot block a commit. What is owed is unchanged: read a verdict
  when one is obtainable, and act on a REVISE/REJECTED.


## 2026-08-17 (later) — swept the corpses, and found why they were immortal

- Did NOT go straight to the `UPDATE`. Asked first why a reaper had not already taken rows from
  29 July, which turned out to be the more valuable half.
- **The reaper's live `pre_query` has no `RUNNING` arm** — `AWAITING_RESPONSES` 30m (dispatch)
  and 90m, `EXECUTING_STEP` 4h, and nothing else. `TimeoutMonitor` iterates `awaited_requests`,
  which on these rows is `{}`, so it has nothing to time out at any age. Two recovery paths,
  both structurally blind. Filed **`bugs_open/294`**.
- **The census that licensed the sweep, and that could have come out otherwise:** `RUNNING` rows
  fleet-wide by age — 0 under 15 min, 0 under 1 h, 0 under 4 h, 49 over. `RUNNING` is not a live
  state at all on this fleet, it is a graveyard. Had healthy agents been sitting there for
  minutes, a 4 h reap would be unsafe and 294's candidate (1) would be the wrong fix.
- **Two harms I had not expected, both found by reading what treats `RUNNING` as live:**
  `monitoring.go` counts it as an ACTIVE orchestration (so the fleet's active count was
  overstated by 49 corpses — the instrument that should have caught this was reporting them as
  healthy), and `getActiveOrchestrationTopics` protects their topics from cleanup, so the 49
  rows were pinning **98 Kafka topics** permanently. That is a direct feed into `bugs_open/240`.
- Swept 49 rows to `FAILED`, ids saved first. After: 0 RUNNING fleet-wide, 0 pinned topics.
- **Did NOT apply 294's fix.** It is a live-config change to a fleet-wide reaper (immediate, no
  roll to gate it) on a shared mechanism, and the council gate is down until the quota returns.
  Filed with the measurement and the exact SQL instead, plus the negative control the verify
  needs. A one-off sweep is not a class fix and this file should not read as though it were.


## 2026-08-17 16:40Z — CORRECTION: the "quota exhausted until 2026-09-01" claim was WRONG

I reported a fleet-wide 15-day LLM outage. **It lasted about three minutes.**

- `llm_call_log`, measured now: last success before **11:08:03Z**, then **4 failures in 76
  seconds** (11:08:37 → 11:09:53), then **successes resume 11:13:02Z** and keep going —
  `council-gate` itself succeeded at 11:13:22Z and 11:13:32Z. Hourly totals for the day:
  11:00 **101 ok / 4 failed**, 12:00 **255 / 1**, 13:00 **99 / 0**, 16:00 **55 / 1**.
- **What I did wrong.** I measured at ~11:10 — *inside* the 3-minute window — saw four
  consecutive failures across two agent types and two models, and took the duration from the
  API's own 400 body: *"You will regain access on 2026-09-01 at 00:00 UTC."* Then I reasoned
  explicitly that the named date distinguished this from the transient spikes on 08-14/08-10.
  **That reasoning was the error.** Four failures over 76 seconds is a sample that cannot
  discriminate a 3-minute blip from a 15-day cap, and a duration claim needs a second
  observation separated in time, which by construction I did not have.
- **And it was already written down.** `bugs_open/243` (2026-08-10) records the *identical*
  error text against an outage that lasted 3 h 20 m and ended because the owner added credit —
  **21 days before the date the message named** — and states: *"It did NOT auto-restore on
  2026-09-01; the owner acted."* It also gives the liveness query to confirm recovery
  (`SELECT max(created_at) FROM llm_call_log WHERE success`). **I never grepped for it**, which
  is the same failure I logged in WRONG_CALLS THIS MORNING over `bugs_closed/274`: grepping the
  code and the DB for a mechanism is not grepping the bug queue. Logged again, because a lesson
  repeated within hours is worth more than the first entry.
- **What it cost.** Nothing irreversible, but two real things: I told the owner they had a
  billing decision to make when they did not, and I declined to submit the tripwire to the
  council on the grounds the gate was down — when it had been available for five hours. Round 2
  went out at 16:38Z on the same correlation `7a3c4fb7-…`, carrying BOTH the loop fix and the
  tripwile in one round (`bugs_open/244`: council-gate is 87.8% of August LLM spend, so a second
  round for four lines of logging is a poor trade).
- Every doc I put the claim in is corrected in place and dated: `bugs_open/289`, `bugs_open/294`,
  the loop-engine handoff, register `WFA-016` and its index row.
