# NOTES — bugfix 285 (section-list assembler is lock-blind). Technical log, append-only, newest at the bottom.

## 2026-08-15 evening — session `390a1ae1` (named "bugfix bugs_open/285")

### Ownership / validity checks (before touching anything)
- `who-owns.py 285` → AMBIGUOUS number; OWNED (webdesign_uk_build_service, filing lane). Live
  transcripts grepped for `LoadPageSectionsFromSpec|load_page_sections_from_spec|lock-blind`:
  the filing session `6b041b7a` (last activity 17:51Z = the "case filed" summary), one stale
  hit in `65c5cffd` (09:55Z task notification). No session editing the loader; loader/plan/save
  files clean in `git status`. Owner then stamped the bug file: this fix is a separate lane's.
- Live (19:00 BST): `pages` `4ff10911-ede0-4ba2-943b-547f66859cac` (webdesign.uk/contact)
  `sections=["hero","contact-info"]`; rows hero(1) contact-info(2) chat-input-box(3,
  `lock_type=permanent`, locked_by "webdesign_uk_build_service lane (restore after
  improvement-sweep wiped it 2026-08-11)", component `d6a8f57b-…` tool/active/all-static
  fields). `site_plan_sections` current plan for contact = hero, contact-info. Item
  `a4cd5dc8` `lock_blocked_change:contact:chat-input-box` blocked_action=remove still
  `needs_human_review`.
- Fleet census (query in RUNBOOK §C1): 26 locked rows (all permanent, 0 `removed`); 13 on
  tier-1 pages whose plan omits them: contact + 12 loancalculator.co.uk tool pages (slots
  `tool-1..4`, positions mostly 5 = already exiled by the guard's tail pass, one at 3, one at 4).
  `lock_blocked_change` remove-blocked items: 6 open, five filed 17:11–17:48Z TODAY on
  loancalculator by `page-build-handler` (corr `366d1ab4`, `b61d3568`;
  `spec_sections.source=site_plan_tables`, sections `[hero, ported-prose, faq, tool-cta]`).

### Mechanism traced end-to-end (each link read, not grepped)
- Loader `load_page_sections_from_spec_action.go` (424 lines, all read): 4 tiers, no
  `page_components` read; three per-tier cache syncs (L169, L230, L380) guarded by
  `sections::text IS DISTINCT FROM $1` — **always true** (measured:
  `SELECT sections::text IS DISTINCT FROM '["hero","contact-info"]' … → t`, jsonb form → f).
- Only live consumer of the loader: `page-build-handler.load_spec_sections` (055) →
  `plan_sections` (`spec_sections.sections`) → `resolve_internal_links` (mutates in place,
  returns all) → page-content-writer loop (`llm_field_specs != null` ? LLM : `render_from_template`;
  `render_component` emits `component_id`, `component_function`, `stored_slot_name` via
  `slot_name_from: current_section.name`) → `compile_page_sections` → `save_page_sections`.
- `plan_sections` Path 0 (`loadPageSlotComponentIDs`, slot→component_id, no lock/level filter)
  and Path 1 (`loadSectionComponents`, no level filter) resolve a merged slot name.
  **`loadComponentNameResolver` (282) is NOT called by plan_sections** — callers are
  `v3_site_actions.go:3407` (validate) and `apply_gap_plan_action.go:222/354/882`. → bug-file
  correction #1 (WRONG_CALLS row written).
- Save: DELETE all agent-writable (`pageComponentAgentWritableSQL("")` →
  `datahelpers.AgentWritableSQLFor`, `chrome_render_inputs.go:91`) + INSERT proposal at i+1;
  `loadActiveLockedRows(pageID)`; `matchLockedRow` identity → slot → kebab, consume-once;
  in-proposal → keep row + `overwrite` item; not-in-proposal → tail exile + `remove` item
  (`save_page_sections_action.go:987-1009`) = the symptom.
- Consumers of `pages.sections` that shift when the cache carries locked slots:
  `check_section_source_drift` (ENABLED live) would flag every fixed page → changed in the
  same commit; prune floor `planned − suppressed − locked` becomes exact (was under-counting);
  `check_unresolved_sections` cannot fire (merged names have rows); `rerender_single_page`
  PlannedSections is diagnostic only. Package direction: actions → discovery_checks →
  datahelpers ⇒ shared helper lives in datahelpers.

### Design decisions (reasons in PLAN)
- Merge only when a tier served: a locked-only list is neither plan nor page and a rebuild on
  it would delete unlocked siblings (prune floor would probably refuse, but don't rely on it).
- Membership excludes `build_status='removed'` (0 rows today) — a MEMBERSHIP condition, the
  LOCK predicate is the guard's verbatim.
- Pairing mirrors `matchLockedRow` arms with consume-once so a snake_case plan entry and a
  kebab-case locked slot cannot become two rows (the 189 class).
- One consolidated cache sync after the merge, jsonb-compared (fixes the always-true guard).
- Missteps: none yet this session. (The Plan-agent adversarial pass was killed by the API
  session limit — design not independently critiqued; mutation test stands in.)

## 2026-08-16 morning — implementation session (same session `390a1ae1`, resumed after the session-limit pause)

- Clock note: the "2026-08-15 evening" entries above were written on the same session; the
  implementation landed 2026-08-16 ~09:30–10:30Z. HEAD at start `52db410bc`; by commit time HEAD
  had moved to `67996ebf1` (other lanes) — the archive-HEAD check was re-run against it.
- Standing five created (this dir). Code written per PLAN with one refinement: siteID typed
  `uuid.UUID` (datahelpers already imports uuid); `IsKebabCase` exported so
  `ValidateComponentFunction` needs no second regex.
- Tests: datahelpers 12 merge cases + predicate pin + normaliser; loader ×3 (+ marshal-shape
  doc test); drift ×3. **Mutation runs (recorded verbatim):** `specSections = merged` →
  `_ = merged`: `TestLoadPageSectionsFromSpec_MergesLockedLiveRowAtItsPosition` FAILS on
  `sections = [hero contact-info]`; facts-insertion `if` → `if false`: same test FAILS on
  `section_facts missing or misaligned: <nil>`. Restored, `diff -q` identical.
- Shared-tree trap met: the `actions` test package would not compile in the working tree —
  two OTHER sessions' untracked files (`agent_definition_nullable_columns_test.go` 11:01,
  `component_template_writer_coverage_test.go` 11:00) both declare `stripLineComments`. Not
  mine; verified instead against `git archive HEAD` + my 7 files (build ./... OK; datahelpers,
  discovery_checks, actions all `ok`), twice (HEAD `52db410bc`, then `67996ebf1`).
- Live SQL dry-run of the helper (predicate spelled out): contact → `chat-input-box|3|permanent`;
  loancalculator site-wide → the 12 calculators + index/tool-3. **New measurement:** positions
  had moved since the 19:00 census — `index/tool-3` 4→6, `tool-settlement-calculator/tool-2`
  3→5 — the tail-exile degrading positions during the afternoon's rebuilds.
- Register: LOCK-008 written into `locks.md`; the `000_concept_index.md` row was swept into
  another lane's commit `67996ebf1` minutes before mine (their message says so honestly:
  "two same-file passengers from other lanes (LOCK-008 new row, CTS-060 status)"). Harmless
  and expected (MEMORY [[a-pathspec-commit-still-takes-a-same-file-passenger]]); drift pair
  clean. Noted in my commit message.
- Council submitted: corr `79f70435-fadc-4e1b-b9d3-6d41f437f7fd` (run reached
  `gate_bug_historian` EXECUTING at 10:07Z, ~6 min after publish). Committed `7d9b7334a` with
  `Council-Submitted:` (8 files, pathspec).
- Docs: LANDMINES ×2 (locked slots in `pages.sections` are not drift; `jsonb::text IS DISTINCT
  FROM` always true) → `landmines-verify-dispatch.sh` dispatched 2/2; WRONG_CALLS row (the
  inherited 282 co-requisite claim); 016b §9 addendum + §10 row; bug file "Implementation"
  section + CORRECTED marker on the 282 bullet; cross-reference appended to 282's file.
- Missteps this session: (1) the datahelpers merge test's "input not mutated" check was
  written vacuous (`len(x) != len(x[:len(x)])`) — caught on re-read, replaced with a real
  before/after `DeepEqual`; (2) first live dry-run of the helper SQL failed on my own shell
  substitution (`$1`/`$2` sed after the Go string concat left `AND NOT` empty) — re-ran with
  the predicate spelled out; (3) the plan-agent adversarial pass was lost to the API session
  limit — the design's edge cases rest on my own review + tests, stated in the council risks.
- Same-file passengers, the other direction: my LANDMINES ×2 and WRONG_CALLS appends were swept
  into another lane's `605ab9b1b` (declared in its message) before my docs commit — append-only
  files, content intact at HEAD (grep-verified). Docs commit therefore carries 016b, the two bug
  files and this lane dir only.

## 2026-08-16 afternoon — council round 1 REVISE → round 2 resubmitted (same corr `79f70435`)

- Round 1 verdict (10:26Z): **REVISE**, decided by a gating objection from `prior_art_librarian`
  (HIGH: "the pairing mirrors matchLockedRow arm-for-arm is the load-bearing claim — confirm the
  arm ORDER against the symbol"). Everything else approve/object-medium: bug_historian (best-effort
  skip is log-only — durable trace wanted), reuse_agent + architecture (second locked-row loader
  beside `loadActiveLockedRows`; mirrored arms), guardian (no-locked-rows regression test; live
  consumer census not shown; loud drift-check failure to be told to that pipeline), editquality
  (register/WRONG_CALLS edits not listed as edits), tooling_provenance (no doc_notes row for the
  action), guidelines (LOCK-008 must name the seam's shape), improvement_guardian (item_key
  dedup shape), debug_historian (tmpfs archive caveat; name the pod check). Reviewers' own checks
  measured 25 locked rows (my 26 was 26 at 19:00 the day before — one unlocked or removed since;
  not re-derived) and 7 remove-blocked items (my 6 + one new: `tool-settlement-calculator:tool-2`
  17:58Z — the defect kept firing after my census, as predicted).
- Facts re-verified for the seats: matchLockedRow read verbatim (identity → slot exact → slot
  kebab, consume-once); live consumer census → `page-build-handler v1 load_spec_sections` ONLY;
  `git grep` at `7d9b7334a~1` → no test ever called the action; `kebabCaseRe` had three uses all in
  component_validation.go; no external importer of `actions.NormalizeComponentFunction`; 090
  d9f97c15 outcome CONFIRMED (`collected_data->'verdict'->'result'->>'outcome'`); item_key has no
  timestamp (`lock_helpers.go:162`).
- Round-2 code `57336c127`: durable `LOCKED_MERGE_SKIPPED` agent_error_log row on a skipped merge
  (+ test); `activeLockedRowsSQL` hoisted in save + `TestRowAndListLockedLoadersNegateTheSamePredicate`;
  `TestLoadPageSectionsFromSpec_NoLockedRowsIsOneSyncNoMerge`. Verified against `git archive HEAD`
  + the four files; scratch archive removed after (tmpfs landmine). The `actions` test package
  compiles in the shared tree again (the two other sessions' duplicate `stripLineComments`
  files were resolved by them).
- doc_notes row for the action written: `b8faa498-8234-4546-a005-6180c19f075d`
  (`subject_type=action`, `subject_key=load_page_sections_from_spec`).
- Round 2 published 15:1xZ with `RESUBMIT_CORR` (run envelope `f53c39c2`, orch `e79f31d9`);
  trigger output tee'd to scratch and NOT re-run (LANDMINES: a TRIGGER script publishes on every
  invocation).
- Misstep: my first round-2 JSON edit appended a 9th edit and then truncated to 8 — dropping the
  very lockstep edit the seats asked for; caught by printing the file list; folded into edit 5.
- 15:11Z the first round-2 envelope (`e79f31d9`) went `complete_invalid`: `edit 5: file path must be
  repo-relative with no traversal or whitespace` — my three-files-in-one `file` field. Fixed (single
  path; other files named in symbol/rationale) and republished ONCE (`RUN_ORCH_ID=0b7c937d`,
  envelope `eca48e3f`). WRONG_CALLS row written. Check the run's `current_step` at +60 s next time.
- **Round 2 verdict 15:31Z: APPROVED** ("approved with 3 advisory objection(s) — none high-severity",
  3 abstained). prior_art_librarian → approve. Advisory: bug_historian medium (other writers of
  `pages.sections` can still write a lock-blind list between builds — true, the loader repairs it
  at the next build; recorded in LOCK-008 as residual 1) + low (best-effort skip is still a
  proceed — the durable row is the mitigation, stated); guardian medium (the save-file hoist rode
  in the same edit slot as an unrelated delegate — bundling in the SUBMISSION, one-line SQL hoist
  in the CODE; and the drift check's loud failure is fleet-wide — accepted, residual 2) + low
  (jsonb sync changes updated_at cadence fleet-wide — intended); architecture medium
  (cross-package contract without an enforced call path — residual 3, source-scan lockstep named
  as the mechanical answer if a fourth reader appears); reuse_agent low ×2 (mirrored arms, two
  loaders — answered in round 2, accepted as advisory); editquality low (says arm 3 is not IN
  matchLockedRow — correct: it stands in for the identity arm the list cannot apply; the
  resubmission said so, and the register entry says so). Both code commits already carry
  `Council-Submitted: 79f70435`; 098 credits them at report time. Register/016b/bug file updated
  with the verdict.

## 2026-08-16 late afternoon — APPROVED, and round 1 is LIVE (unexercised)

- **Council round 2: APPROVED** (`0b7c937d` / envelope `eca48e3f`, `complete_approved` 15:31:26Z).
  `council_decide`: `decision=approved`, `round=1` (the re-published envelope restarts the
  counter), 14 reviewers, 3 abstained, `decided_by = "approved with 3 advisory objection(s) —
  none high-severity"`, `unreadable=0`, `gated_by_truncation=false`. Verdict read from
  `orchestration_states.collected_data->'council_decide'` and the `doc_notes` council-gate row;
  full report `diagnosis_artifacts kind=council_report`, corr `79f70435`, 15:31:14Z.
  The three advisory objections: **architecture** (medium, `ARCHITECTURE_SIGNAL: needs_rfc`) —
  `MergeLockedPageSlots`/`LockedPageSlotsSQL` are a cross-package contract with no enforced call
  path, so a future section-list reader can still be written lock-blind; the seat explicitly
  recommends *"proceeding but opening the RFC in parallel rather than gating this round"*.
  **bug_historian** (medium) — the census asked "who calls this ACTION", not "who else WRITES
  `pages.sections`"; (low) the best-effort skip is a new instance of the silent-fallback family,
  name it as such. **guardian** (medium ×2, low ×1) — the `save_page_sections` hoist rode in a
  bundled edit slot; the drift check's new loud failure is fleet-wide; the cache-sync
  consolidation changes write behaviour for every build. **editquality** (low) — "arm for arm"
  overstates: the merge's third arm is a *functional stand-in* for the guard's identity arm, not
  the same arm. **prior_art_librarian** approved but asks a human to diff the quoted
  `matchLockedRow` body against the committed file (the code index lags HEAD).
- **No new trailer, deliberately.** `7d9b7334a` and `57336c127` already carry
  `Council-Submitted: 79f70435`; CLAUDE.md says `098` resolves the correlation at REPORT time and
  credits the commits automatically once the verdict turns approved. Forward-only forbids an
  amend, and writing `Council-Reviewed:` onto a later unrelated commit would be a false join.
- **Liveness, measured at the artefact (15:45Z).** Pods `agent-chassis-5d95ddddfd-{48lv6,vtfdx}`
  run `v1.0.1304`, started 10:41Z. `build provenance` had already scrolled past `--tail=3000`
  (the documented shelf life), so: binary probe on 48lv6 — `locked_sections_merged` PRESENT,
  `LOCKED_MERGE_SKIPPED` absent, `load_page_sections_from_spec` PRESENT (positive control),
  `deadbeefcontrolstring` absent (negative control). Round 1 in, round 2 out — consistent with
  the commit times (10:09Z vs 15:10Z) either side of the pod start. Corroborated at the data:
  20/20 post-roll `page-build-handler` runs carry the new `spec_sections` keys.
- **The merge has NOT fired in production.** All 20 runs: `locked_merge_count = 0`. Correct
  behaviour (no locked rows on those pages) and NOT evidence the merge works.
- **The `remove` counter is currently a blind zero.** 0 new items since the roll, but 0 locked
  pages rebuilt — no demand. Recorded before quoting it anywhere; the near-miss is exactly the
  memory-file lesson *"a post-fix ZERO needs a DEMAND control"*, which is why the register status
  line says it in terms rather than reporting "0 items = fixed".
- **bug_historian's objection, answered by measurement** (now in LOCK-008 residual 1): writers of
  `pages.sections` = `apply_gap_plan` ×3, `ensure_page_section_layout` (empty-source rescue only),
  the loader, and 7 INSERT-a-new-page paths that are benign by construction (no `page_components`
  rows exist yet). Callers of `save_page_sections` over 30 days: `page-build-handler` 98 (all with
  the loader) and `page-rerender` 625 (none) — and rerender composes from the stored ROWS
  (`loadStoredSections`/`carryStoredSection`), so it cannot drop a locked row; the fleet's 38
  `overwrite` vs 7 `remove` split is that difference showing up in the data. The other four
  callers: 0 runs in 30 days. `page-rebuild` is the one to re-check if it wakes (plans from
  `current_page.sections`, no loader in its chain).
- **Open, not done by this session:** the architecture seat's RFC (single mandatory entrypoint
  for section-list assembly) is not written; the acceptance run is not fired (it redeploys a live
  shopfront page and the owner assigned it to the filing lane); round 2 is inert until the next
  roll.

## 2026-08-17 morning — round 2 live, and the merge PROVEN on a natural rebuild

- **Both rounds live on `v1.0.1305`** (pods `agent-chassis-5657f446c7-{q7b82,r6sf2}`, started
  2026-08-16 22:07Z). Probe on q7b82: `LOCKED_MERGE_SKIPPED` PRESENT (round 2 shipped this time),
  `locked_sections_merged` PRESENT, `load_page_sections_from_spec` PRESENT (positive control),
  `deadbeefcontrolstring` absent (negative). RUNBOOK C6.
- **The merge FIRED, unprompted.** `page-build-handler` corr `b45d9965-1efb-4d58-aa05-447ce4bc83a8`,
  2026-08-16 **16:09:17Z** — 24 minutes after yesterday's "unexercised" entry, under round 1 on
  v1.0.1304 — on loancalculator.co.uk `index`. `locked_merge_count=1`,
  `locked_sections_merged=["tool-3"]`, `source=site_plan_tables`, proposed list = the plan's 5
  names + the locked calculator at the tail (its live position, 6). All five owner criteria hold;
  the evidence table is in the bug file. The one I care most about: **the locked row's
  `updated_at` is still 2026-08-09**, i.e. the save consumed-and-kept it without a write, while
  the five unlocked siblings all carry `created_at = updated_at = 16:23:23Z` — the guard's own
  control (do not "fix" this by never rebuilding anything) passing in the same row set.
- **The `remove` counter now has a demand control and still reads 0.** Newest `remove` item
  fleet-wide is still 2026-08-15 17:58:27Z. Yesterday that zero was blind; today a locked page
  has actually rebuilt, so the zero answers a question that was asked.
- **`section_source_drift` filed nothing** since the roll — the drift-check half is doing its job
  (without it, this page would have been a false positive the moment the cache gained `tool-3`).
- **MISSTEP, caught before it became a false alarm and worth the space:** my first artefact check
  grepped the served page for `tool-3` and `data-component="tool-3"` → 0 and 0, which reads
  exactly like "the locked calculator is not being served", and I was one step from filing that as
  a new defect. Both markers were mine, not the artefact's: `tool-3` is a `page_components.slot_name`
  (a DB label that never appears in markup) and this tool's template emits no `data-component` at
  all. Taking the literal FROM the row (`substring(rendered_html …)` → `tool-loan-repayment-section`)
  and re-running with a present-control (`tool-list`=25) and an absent-control
  (`zzz-absent-control-zzz`=0) gives 16 hits: the tool IS served. WRONG_CALLS row written. Same
  family as the memory-file lesson "your measurement answers the question you ENCODED" — and note
  the failure mode was a FALSE POSITIVE for a bug, which is the direction that wastes a lane.
- **Where that leaves the case:** fixed, live, and no longer reproducible — the stated
  `bugs_closed/` bar. Not moved, because the owner's acceptance text names webdesign.uk `contact`
  and the chat-box lock is the filing lane's; that page has not rebuilt since the roll and its
  evidence would be identical in shape (tool-level locked row, tier-1 page, plan omits it).
  Recommendation recorded in the bug file: move on the owner's word.

## 2026-08-17 — CLOSED on the owner's word; what is left of the lane

- Case file moved `bugs_open/` → `bugs_closed/` (both paths named on the commit; verified at HEAD
  with `git ls-tree`, not at the tree — the `git mv` + pathspec trap ships a COPY otherwise).
  Header now carries the CLOSED banner, the evidence pointer, and the one thing the closure does
  NOT assert (contact/chat box never rebuilt through the fix).
- LOCK-008 status → **deployed**; 016b §10 row → CLOSED with the `/bugs_closed/` path and the
  resolve-by-slug warning (both 285s are now in `bugs_closed/`, so a bare number is worse than
  ever).
- **Open after closure, and none of it blocks:** (1) `RFC_033` — the architecture seat's
  needs_rfc, awaiting an owner ruling; this lane's recommendation is the source-scan lockstep,
  not a single entrypoint. (2) LOCK-008's open review question — should a merged locked section
  be marked "carry the stored row" so the writer/render loop spends nothing on it and the
  `overwrite` item stops recurring? Cost today is one template render per locked section per pass
  (all locks are `static`-field components, no LLM spend measured); the benefit is silence, not
  correctness. (3) The three advisory council residuals, recorded in LOCK-008: the between-builds
  window where another `pages.sections` writer can leave a lock-blind list, the drift check's
  fleet-wide loud failure, and the unenforced cross-package contract (= RFC_033). (4) `page-rebuild`
  is the one dormant save path that would re-open the class if it woke — 0 runs in 30 days,
  plans from `current_page.sections` with no loader in its chain. (5) webdesign.uk `contact`
  acceptance + the chat-box lock: the filing lane's, unchanged.
