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
