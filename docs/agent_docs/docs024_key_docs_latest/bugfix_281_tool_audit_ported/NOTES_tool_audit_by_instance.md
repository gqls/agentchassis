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

---

## 2026-08-18 (evening) — picked the lane up from the 08-17 handoff; both its blockers were already done, and the live gap is one status over

**First move was to ground the handoff rather than work it, and that was the whole session's value.**
Its §2 listed five open items with `bugs_open/294` as "the only thing genuinely waiting on a human".
Two of the five had been closed by other sessions in the meantime:

- **294 CLOSED** (migration `463`, council APPROVED round 2). I did not take the file's word: read
  the reaper's live `pre_query` and confirmed the fourth arm `failed_running` is present, and that
  `RUNNING` is 0 rows fleet-wide.
- **`a436d898f` SHIPPED** — `git merge-base --is-ancestor a436d898f f0117fb8b` is true, and
  `f0117fb8b` is the commit stamped into the `v1.0.1309` binary.

**The 289 fix still holds, but the first number I looked at would have said otherwise.** Aggregate
`collected_data` for `build-dispatch-loop` is now 450 kB avg / 1,642 kB max over 159 runs, against
the 104 kB / 229 kB over 10 runs recorded right after the roll — a 7× rise in the max. **Size is not
the test and I nearly treated it as one.** The test is the per-lap ratio: on the three largest runs
in a 6 h window, every `_iter_N_done` key is 77 bytes flat, ratio 1.00. The growth is legitimate
per-item payload. Demand control: 160 multi-lap runs in the window, so the flatness is not silence.

**Residual (6)'s precondition is now measurable and met.** 81 `loop_complete` steps across 34 runs
carry `loop_iteration` without the explicit `loop_iteration_terminal` flag — and every one belongs
to a run created before the 17:05Z roll (latest 16:56:21Z, nine minutes before the pods started),
all terminal. So no non-terminal run needs the fallback. Left undone deliberately: it needs a build
and roll, and the code is inert either way.

### The live gap: `INITIALIZED`, filed as `bugs_open/310`

`294`'s post-close flagged it; I verified it end to end. Neither the reaper (four arms) nor
`database-cleanup` (two clauses) names `INITIALIZED`, and zero scheduled tasks mention it — so such
a row is reaped by nothing and pruned by nothing. Four `StatusInitialized` references exist in Go:
the constant, one live writer (`state.go:734`), one reader (`coordinator.go:741`), and a dormant
writer inside `OrchestratorHelper`, whose constructor has zero callers (re-verified by grep, which
is what `294`'s Correction 3 asks the next reader to do rather than cite it).

**The thing that made the census interpretable was noticing the table is pruned asymmetrically.**
`COMPLETED`/`FAILED` reach back 24 h; `INITIALIZED` and `CANCELLED` are never deleted. So "2 rows"
is a lifetime figure while "964 completed" is a one-day figure — same table, same `GROUP BY`, two
different populations. Both numbers were needed: without the 964 the two rows read as a dead
pipeline, and without the pruner predicate the two rows read as a 24-hour explosion. Filed as a
landmine.

### Missteps, in the order I made them

- **I nearly reported a regression from an aggregate.** See above — the 7× max-size rise. Caught by
  asking what the disconfirming result would look like (a ratio near 2.0) before interpreting the
  number, rather than after.
- **My branch-B guard test failed, and the guard was fine — my harness was wrong.** I rebuilt the
  new `pre_query` inside a shell heredoc to simulate "already applied", and `echo` added a leading
  newline, so the md5 differed and Guard 1 correctly REFUSED. For about a minute that looked like a
  guard defect. **Reconstructing a byte-exact payload through shell quoting is not a test of the
  thing; it is a test of your quoting.** Redone the honest way — run the migration twice in one
  transaction — which is also the real-world repeat-run case, and it passed (`UPDATE 0`, "already
  present — this run is a no-op").
- **Migration number collision, live.** I built the pair as `469` and another lane created
  `469_render_audit_rotation_three_day_window.sql` **two minutes** before I wrote mine. Renumbered
  to `470`. The `sed 's/469/470/g'` that fixes the references also runs over the **embedded SQL
  payload**, so I re-verified both `pre_query` texts byte-exact by md5 *after* the rename. They were
  clean, but the check is the point — a global sed over a file whose payload is SQL is exactly how
  a payload acquires a silent edit.
- **I misread the wall clock** and briefly thought the `090` run had been stalled 80 minutes. The DB
  clock was 18:27, not 19:35. Nothing followed from it because I checked the orchestration rows
  before saying anything, which is the only reason it is a footnote rather than an entry in
  `WRONG_CALLS.md`.

### What is proven about `470`, and what is not

Proven, all inside rolled-back transactions, live row unchanged throughout (`md5 91ba9704`):
parse check passes on the new text and **is caught** on a corrupted copy; all three Guard-1
branches (apply / declared no-op / REFUSED); the rollback sidecar restores byte-exactly; and the
induced test fires **both ways in one tick** — the 870 h row FAILED with the right error, the
10-minute negative control survived, siblings all 0.

**Not proven, and cannot be without applying:** that the *scheduled* tick fires it. Every test above
executes the `pre_query` directly and bypasses the scheduler. That is the gap the owner's apply
decision opens, and it is stated as such in `310`.

---

## 2026-08-19 (morning) — post-roll verification on `v1.0.1314`, and the lane's remaining item is a judgement call, not a task

**The build.** `IMAGE_TAG` is `v1.0.1314`; chassis pods started 07:52:27Z / 08:05:39Z, image
digest `sha256:d0257576…`. Read the commit from the **image's own label** rather than the tag or
the pod age (`docker inspect … org.opencontainers.image.revision`) → **`d3590ca46`**, and
`git merge-base --is-ancestor 50467bc74 d3590ca46` is true, so the build contains everything this
lane committed. The `build provenance` startup line had already scrolled out of `--tail=3000`,
which is "not in range", **not** "unstamped" — the label is the durable answer.

**Nothing of this lane's was waiting on the build.** Everything committed 08-18 was docs plus one
SQL config change; the only Go item left (residual 6) is not written.

**All three fixes re-verified against the new binary, each with the control that makes the zero mean
something:**

- **289 (loop doubling).** Per-lap `_iter_N_done` keys on the largest multi-lap runs since the roll:
  **77 bytes flat, ratio 1.00.** Demand control: 123 runs across 6 agent types since 07:52Z, so the
  flatness is not an idle fleet. Size alone remains the wrong test — see the 08-18 entry.
- **294 (`RUNNING`) + 310 (`INITIALIZED`).** Both reaper arms present in the live `pre_query`;
  the task last ticked 09:01:23Z. **Non-terminal rows older than 4 h fleet-wide: 0.** In fact
  non-terminal rows of *any* age were 0 at 09:02Z — which needed the same demand control before it
  could be read as health rather than silence.
- **Residual 6's precondition, re-measured and now cleaner than on 08-18.** `loop_complete` steps
  carrying `loop_iteration`: **1,196 across 357 COMPLETED runs all carry the explicit
  `loop_iteration_terminal` flag** (newest 08:42 today), and **0 need the fallback**. The only
  unflagged steps left are 22 in 6 **CANCELLED** runs from 2026-07-24 — terminal, never re-executed,
  and never pruned because `database-cleanup` does not name that status. The 08-18 cohort (19
  COMPLETED + 40 FAILED) has aged out entirely.

**A correction to my own 08-18 reading of the lane's scope.** I wrote that `281` was the lane's
originating bug and treated it as open. **It is CLOSED** — `bugs_closed/281`. I had run
`ls bugs_open/ bugs_closed/ | grep 281`, which spans both directories and prints only the basename,
so it cannot tell you which one the file is in. `git ls-tree -r --name-only HEAD -- bugs_open/
bugs_closed/ | grep '/281_'` answers it in one command and is the same check the `git mv` landmine
already prescribes for verifying a move.

**Chasing the per-page `audit_review` residual, and stopping.** The loop-engine handoff lists it as
"the 281 lane's item". Looked for it and could not find the artefact: `item_type='audit_review'`
returns **0 rows in `site_work_items` AND in `site_work_items_archive`** — and I checked the archive
because of the standing 7-day-window landmine, so the zero is not a retention artefact. The real
types are `audit_tool` (39) and `audit_finding_{audience,tone,brief_fidelity}` (39 total), all with
`design-audit_*` keys dated 2026-03→08-09, i.e. a different and older mechanism. `audit_review` is
an item **KEY prefix**, not an item_type — that is what 289's close was quoting. **The row 289 cited
(`734fa16b`, created 08-16) is in neither table.** Not investigated further: it belongs to the
285/281 sub-lane, not the loop engine. Flagged because **the residual is recorded only inside CLOSED
bug files**, which is where a residual goes to be forgotten.

### Residual 6: the precondition is met and I am NOT taking it unilaterally

One injection site sets the flag — `loop_expansion_handler.go:182`, guarded by
`if substep.Action == "loop_complete"` — and `loop_iteration` is set on every injected step at
`:153`. So deleting the `loop_iteration`-presence fallback in `isLoopIterationTerminal` is safe
*given* the precondition, which holds.

**But the trade-off the council's architecture seat did not weigh is that the fallback is no longer
merely redundant — it is a second, independent discriminator against the exact failure that caused
`289`,** a 2^N blow-up that reached 22 MB rows and left `tool-auditor` completing 1 run in 63.
Deleting it makes the system depend on one line being right. The gain is structural tidiness; the
seat itself framed the wider `action`-overloading observation as the RFC signal and "**not** this
fix". Two lines of defence in depth against a recently-severe bug is a poor thing to trade for
tidiness, so this goes to the owner rather than into a commit.
