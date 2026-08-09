# NOTES — bugfix 220 (append-only, newest at the bottom)

## 2026-08-08 ~18:45Z — lane opened, ownership swept

- who-owns 220: commits are the filing (206 lane) + a contrib (116 lane). Transcript
  grep for FIX-SITE symbols (`load_page_record|availableBuilders|check_phantom_internal_links`),
  not the bug number: b5a58a2b = 206 lane, closing message says 220 is a follow-up it
  is NOT taking; 63f97914 = fragments lane (reads 220, works dead_fragment_link);
  005ad3d6 = 201/RFC_017 lane — touched `complete_work_item_verification.go` TODAY
  (fail-closed flip, committed). Conclusion: free. [MEASURED — greps + tail reads]
- Re-verified the three mechanism legs live (see PLAN § Validity). The vetcomparison
  instance itself is healed (206 lane built directory-index 17:02Z) — do NOT re-use it
  as the acceptance case without re-checking for a fresh unbuilt target.

## 2026-08-08 ~19:00Z — the finding the bug file does not state

`load_page_record_action.go` resolves page_name BEFORE page_id (`:7-10`, `:174`).
Candidate 1 as written in the bug file ("map page_id and have load_page_record prefer
it") would be INERT: spec.page_name (container) is always present for this item type,
so the name always wins. Confirmed by reading the live step config:
`page-build-handler.load_page_record.config` = `{page_id: input_data.spec.page_id,
page_name: input_data.spec.page_name, site_id: site_record.site_id}` — note BOTH
config paths point at the CONTAINER for this item type (spec.page_id is the container;
only the COLUMN carries the target). Fix shape per RFC_010 §2: opt-in field
`authoritative_page_id`, default absent = today's behaviour.

- Precedent found for the mapping half: `site-work-orchestrator.call_handler` already
  maps `"page_id?": "current_fix_item.page_id"`. [MEASURED — live agent_definitions]
- Zero live agents map `current_item.page_id`; `input_data.page_id` is read only by
  `page-retraction` + `deduplicate-sections`, both with 0 work items ever. [MEASURED]
- Only build-dispatch-loop maps `current_item.spec.page_name` (jsonb_path query over
  live rows). The sibling loop does not share the defect. [MEASURED]
- `ExtractActionInputs` ignores undeclared config keys at runtime (declared-field
  iteration; UnknownConfigKeys is the offline audit) → no image/config ordering
  constraint. [VERIFIED — read action_inputs.go:198-247]

## 2026-08-08 ~19:30-20:00Z — implemented, tested, submitted

- Go: `authoritative_page_id` on load_page_record (body refactored into shared
  `queryPageRecordRow` so the id path cannot drift from the name path);
  `VerifyUnbuiltInternalLinkResolved` + registration; coverage-map entry removed.
- A guard I did not know existed caught the missing half:
  `TestRegisteredVerifiersMatchClaimTimeoutExclusion` — registering a verifier obliges
  the claim-timeout lockstep (declared list in sql_for_agents/220 + live column). Done
  both halves (220 edit same commit, mig 341 on 331's template; live list read first:
  8 entries, no drift). [VERIFIED — test failed, then green]
- Tests green on the dirty tree AND against a clean `git archive HEAD` overlay (the
  tree carries another session's WIP in this same package). [MEASURED]
- MISSTEP (full entry in WRONG_CALLS.md 2026-08-08): ran `landmines-sync.py --apply`
  directly after appending the LANDMINES entry — the documented wrapper
  `landmines-verify-dispatch.sh` should have been used; recovered by hand-firing
  `trigger-landmine-verifier.sh`, correlation `f70fb3af`. Check the verdict later:
  `SELECT created_at, left(body,120) FROM doc_notes WHERE subject_key LIKE
  'LANDMINES.md#loadpagerecord%' AND categories ? 'landmine-verification';`
- Council: submitted r1, correlation `def4441c-df3a-460a-b2ce-208da04f4023`
  (submission JSON in this dir). Committing with `Council-Submitted:` per the
  2026-07-30 trailer rule; budget ~30 min for the verdict, find the run by payload
  not by printed id.

## 2026-08-08 ~19:20 UTC — committed, migrations applied+recorded, council executing

- Timestamp correction for the entries above: earlier stamps written "Z" were BST
  (local); UTC is one hour earlier. This entry and later ones use UTC.
- Commit `a60a13cbb` (20 files) with `Council-Submitted: def4441c`; gofmt followup
  `03433f4b5` (the pre-commit pattern check caught verifier_coverage_test.go — the
  map realigned after the entry removal).
- Same-file passengers, disclosed in the commit message: `000_concept_index.md`
  carried FTW-042 + SQAM-003 rows from concurrent lanes; my header paragraph
  accounts for all three counts. Also observed the reverse: my WRONG_CALLS.md entry
  was committed by ANOTHER session's sweep (`ee945d7da`) minutes after I appended it
  — nothing lost, entry is in HEAD under their message, findable by path.
- Migs 340+341 applied by hand and verified live (mapping key, load_page_record key,
  9-entry claim-timeout list — all read back), recorded with --record-only in the
  same motion. [MEASURED — the three read-backs are in the ledger notes]
- Council run found by payload: `current_step=review_reuse_agent, EXECUTING_STEP`
  at 18:14 UTC — dispatch was ~5 min this time, not 29.

## 2026-08-08 ~18:20-18:50 UTC — round 1 REVISE; both real catches fixed; round 2 in

- REVISE, gating objection from bug_historian. Round completed in ~6 min from
  submission (dispatch was ~5 min, not the budgeted 29). Full report:
  `diagnosis_artifacts kind='council_report', correlation def4441c` (column is
  `body`, not `content` — the first read errored).
- **The HIGH was a real catch**: a valid uuid matching no row returned {found:false}
  — the soft-miss contract — routing the saga through the success-labelled
  complete_error path, the exact silent no-op the input exists to close. FIXED
  (fatal error + a test that also pins "no second query" via ExpectationsWereMet).
- **The LIKE catch was real too**: raw href into LIKE-concatenation; `_` is a
  wildcard, over-match refuses a resolved item. FIXED with position(). ⚠ the
  sibling VerifyDeadFragmentLinkResolved carries the same LIKE shape — flagged in
  the code comment and in bugs_open/220, NOT edited (active lane's file).
- Everything else answered by measurement in submission_220_r2.json: ONE producer
  at source (the fragments arm `continue`s on unbuilt targets by design); ledger
  rows quoted; 340's pre/post guards quoted (the seat had only the sketch);
  zero-history reads verbatim.
- Fixes committed e55cbfa64. Round 2 resubmitted on the SAME correlation
  (RESUBMIT_CORR), run orch id e3df060f.

## 2026-08-08 ~19:30 UTC — round 2 blocked on FLEET-WIDE API credit exhaustion; lane parked clean

- Round 2 terminated at `complete_invalid` — NOT a verdict and NOT a validation
  refusal of the plan: `review_editquality`'s LLM call failed with the Anthropic
  API "credit balance is too low" 400. **Fleet-wide: 31 such failures 18:25–19:20
  UTC** — every LLM-driven pipeline is down until the owner tops up billing.
  [MEASURED — agent_error_log]
- Re-fire when credits are restored (same command, same correlation):
  `RESUBMIT_CORR=def4441c-df3a-460a-b2ce-208da04f4023 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh docs/agent_docs/docs024_key_docs_latest/bugfix_220_unbuilt_link_dispatch/submission_220_r2.json`
- Landmine-verifier verdict on my entry: NEEDS_HUMAN_REVIEW — "authoritative_page_id
  returns 0 hits in the content index". Explained: the verifier ran 18:07 UTC against
  the CODE INDEX; the commit carrying the symbol (a60a13cbb) landed minutes later.
  `git grep authoritative_page_id HEAD -- platform/` returns hits (load_page_record_action.go
  + the test). The index-lag shape is bugs_open/108's; the verifier's own blindness to
  non-Go footprints is bugs_open/223's. No entry defect; no action needed beyond this note.
- Lane state at park: substance COMPLETE (Go committed a60a13cbb + 03433f4b5 +
  e55cbfa64; migs 340/341 applied+verified+recorded; register WII-012; LANDMINES;
  docs). OWED: (a) round-2 verdict once credits return; (b) post-roll pod-grep +
  behavioural acceptance (RUNBOOK); (c) candidate 4, deferred on record.

## 2026-08-08 ~22:06 UTC — round 2 APPROVED (3 advisory objections, none high; 3 abstained)

Credits refilled by the owner; the saved re-fire command ran unchanged (orch
`f42cff55`). Verdict: **approved**. The advisory objections, answered on the record:
- **Two-gates lore** (bug_historian + debug_historian): "input_mapping is an
  ALLOW-LIST and so is a claim query's RETURNING". The second gate is OPEN for this
  loop: `LoadWorkItemsAction`'s SELECT includes `wi.page_id`
  (`load_work_item_actions.go:646`) and exposes it on `current_item` (`:776-777`) —
  which is the very column-exposure the 154 lane built; round 1's own answered
  checks confirmed it live ("mentions_page_id: true" + the claim-step read). No
  second edit needed.
- **Operation label** (guardian): migration edits should be labelled
  `config_change`, not `add` — bookkeeping for FUTURE submissions, adopted.
- **tool-recreation-handler stays name-first** (guardian): deliberate — the opt-in
  is per-caller by the RFC_010 ruling; its lane opts in when its items need it. The
  landmine entry names both carrying agents so the next reader sees the split.
The platform commits carry `Council-Submitted: def4441c` and are auto-credited by
098 now the verdict is approved. Lane's remaining owed items: post-roll pod-grep +
behavioural acceptance (RUNBOOK), candidate 4 on its demand signal.

## 2026-08-09 — fresh chassis roll: deploy PROVEN at the pod; behavioural acceptance in flight

- Both replicas (`agent-chassis-5c8776654c-wml6w`/`-zhz2g`, ~5 min old), one exec
  each: `authoritative_page_id` = **3**, `names no page row for site` (a string only
  the NEWEST commit e55cbfa64 adds, so the build postdates all three commits) =
  **1**, `unbuilt_internal_link verifier` = **5**, invented-string negative control
  = **0**. No natural negative control exists (additive change; the LIKE string my
  r2 removed also lives in the fragments verifier, so it cannot serve). [MEASURED]
- Behavioural acceptance target chosen by census: **dartsonline.com / grip-styles**
  (blog-post, 3 plan sections, 5 live linking pages, site unlocked, page-build-handler's
  own family — a directory/tool target would exercise the DEFERRED candidate 4 and
  loud-fail by design, wrong page for an acceptance PASS).
- One-shot improvement loop fired at dartsonline with the safe kcat pattern
  (payload in container COMMAND + PUBLISH_OK receipt — the 07-26 kcat landmine):
  **PUBLISH_OK**, correlation `110acf5a-c0d5-48b6-8d2a-84e60f842171`.
- ⚠ the shipped trigger script `060improvement_loop/076_improvement_loop_trigger.sh`
  hardcodes SITE_ID/DOMAIN to robot-hands.com AFTER its own arg parsing — passing
  args does NOT work; a payload built by hand was used instead. Left unfixed (not
  this lane's file); noted here so the next user does not fire at the wrong site.

## 2026-08-09 ~10:30 UTC — acceptance read-out, containment, mig 342; session parking

- Full mechanism + evidence in the bug file's 2026-08-09 addendum (single source;
  not restated here). Register WII-012 updated with the acceptance + a
  contamination LANDMINE. WRONG_CALLS: the 342 numbering collision (I checked the
  ledger for 340/341 and carried that freshness to 342 — the thunder lane's 342
  already existed).
- Council for 342: submitted, **REFUSED client-side by the scope filter** — a
  config-only change carries no platform/ file and docs paths never spend credits
  (owner ruling 2026-07-17). Not forced. The def4441c round reviewed the census
  342 rides on; this refusal + reasoning recorded here is the review trail.
- Items ledger for the acceptance run: 338deb27 complete/verified (disjunct b,
  stored substrate); a8327624 failed (regression floor, loud, correct);
  91c005ed/a3d21774/2330e479/af6ffa42 cancelled pre-dispatch (contamination
  class); 47ba8f2c/3c10ab6c (beginners rerenders) cancelled (would publish
  contamination); 3cb732b1 = the repair, triaged at priority 30.

## 2026-08-09 ~13:15 UTC — repair PROVEN at the served artefact; convergence run re-fired

**Priority 1 (the repair) is done and checked at the artefact, not at the status.**
`3cb732b1` read `complete` at 10:14:32 — which on its own proves nothing (a
`complete` work item is not a repaired artefact). The checks that do:

- Beginners' three components (`hero`/`article-body`/`call-to-action`) all
  rewritten 10:13:58 and hold BEGINNERS copy again — hero headline "Everything You
  Need To Start Throwing", body opening "You've thrown a few house darts in the
  pub…". The contaminating grip-styles hero ("Your Grip Decides More Than Your
  Barrel Does") is gone. [MEASURED]
- `pages.deployed_at` for beginners = 12:31:31Z, i.e. AFTER the repair — so the
  rerender that published it published the repaired copy, not the contamination.
  The two held rerenders stayed cancelled; this deploy came from the queue drain.
  [MEASURED]
- **Served**: `curl https://dartsonline.com/blog/beginners.html` → **200**, 29,183
  bytes, carries both beginners signature phrases. Note the URL shape — the page
  `name` is `beginners` but the url is `/blog/beginners.html`; curling `/beginners`
  404s and would read as a broken page. [MEASURED]
- The 12 "grip" hits on the served page are all legitimate (grip is one of the
  three specs a beginners guide covers) and one of them is **the unbuilt link
  itself**: `<a href="/blog/grip-styles.html">grip styles guide</a>`, still live,
  still 404. So the defect condition is intact and re-detectable — which is what
  the convergence proof needs.

**Observation on disjunct (b), recorded because it reads oddly in the ledger.**
`338deb27` verified at 10:01 via disjunct (b) — "href no longer rendered on the
container". That was HONEST about the substrate it read and simultaneously an
artefact of the damage: the contamination had replaced beginners' copy, taking the
href with it. The repair restored beginners' copy and therefore restored the link,
so that 'verified' is now stale and the finding is re-mintable. Not a defect in
the verifier (the link genuinely was absent; and the item's own fix text accepts
link-removal as a remedy), and mig 342 closes the route by which a handler can
damage a container at all — but worth knowing that **disjunct (b) cannot
distinguish a link legitimately removed from a container whose content was
destroyed**, and the two differ in that the second one comes back.

**Pre-flight for the re-fire, all re-checked at the moment of use** (the lane's own
rule — migration numbers, censuses and item ids expire in hours):
- All three config legs still live: dispatcher `"page_id?": "current_item.page_id"`
  (note it sits under `call_handler.config.input_mapping`, one level deeper than
  the step — my first query read the step and returned empty, which looks exactly
  like a missing leg); `load_page_record.authoritative_page_id` =
  `input_data.page_id`; `save_sections.page_name_field` = `page_record.name`.
  [MEASURED]
- Binary re-greped on the CURRENT pods (`agent-chassis-5c5bbf8548-khpl4`/`-mkdjp`,
  created 12:23Z — a NEWER roll than the acceptance run, so the earlier pod-grep
  had expired): `authoritative_page_id` = 3, `unbuilt_internal_link` = 7, invented
  negative control = 0, both replicas. [MEASURED]
- No discovery had run since 08:58; the 10:09→12:57 orchestration stream is the
  dispatch loop draining that run's queue, not a new discovery.
- `completeness-discovery-agent` is the agent carrying `phantom_internal_links`
  (live config query), so the improvement loop is the trigger that re-mints.

**Re-fired the one-shot improvement loop at dartsonline**, safe kcat pattern
(payload in container COMMAND, `PUBLISH_OK` receipt): corr
`576f0ab9-5a17-4449-9bbc-ee1983576433`. Script kept at
`scratchpad/fire_improvement_loop_dartsonline.sh` — it asserts the payload
contains no single quote before firing, because the whole payload is single-quoted
inside `sh -c`. The shipped `076_improvement_loop_trigger.sh` still hardcodes
robot-hands.com after its arg parsing; still not this lane's file to fix.

## 2026-08-09 ~13:20 UTC — the run's ledger so far, plus two incidental findings

**Convergence run (corr `576f0ab9`) progress**, keyed on item ids, never a rolling
window: discovery re-minted **10** `unbuilt_internal_link` items at 13:12:45 (up
from 6 on the morning run — the repaired beginners page restored its link, and
other containers were rerendered meanwhile), all `detected` → all `triaged` by
13:19:36. Targets split exactly as the lane predicted:
- **6 → `grip-styles`** (blog-post, `planned`, never deployed): `4151471c`
  (barrel-weight), `1874c63f` (beginners), `1ad68d52` (flight-shapes), `d1398df9`
  (shaft-length), `cc008ad4` (tool-setup-builder-guide), `e0289053`
  (tungsten-guide). This is the acceptance family — page-build-handler's own type.
- **4 → `section-index` directories** (`brands-index` ×3, `shop-index` ×1):
  `69818add`, `0469f44f`, `6e1b562b`, `b4184d0f`. These are **deferred candidate
  4's demand signal** and are expected to fail LOUDLY rather than converge; that is
  the designed outcome, not a regression. Watch whether they land `failed`.

**That `cancelled` re-minted at all is itself a checked fact**, not an assumption:
the 08-09 morning items were cancelled, and a cancelled row would hold the dedup
slot for ever if the Go list and the index disagreed. `cancelled` is in
`workItemTerminalStatuses` (`work_items_common.go:47`, joined the closed set in
migration 157) — so the slot frees and discovery re-mints. If a future run mints
nothing, check that lockstep FIRST.

**Incidental 1 — my LANDMINES.md entry was swept into another session's commit.**
`190ee4568` ("landmine(205): …", 14:17:27 BST) carries my 26 lines and nothing
else; my own commit `37f1a88ec` (14:18:25) therefore contains only the bug file,
while its message describes the landmine too. Nothing is lost — both entries are at
HEAD, verified with `git show HEAD:<path> | grep -c` — and forward-only forbids an
amend, so this note IS the correction. Textbook instance of the hazard CLAUDE.md
describes; recording it because a reader of `37f1a88ec` will otherwise look for a
landmine that is not in it.

**Incidental 2 — `scripts/trigger-landmine-verifier.sh:84` uses the UNSAFE kcat
pattern** (`kubectl -n kafka run -i --rm … kcat -P` fed from a heredoc) and prints
no receipt, which is the ~4-in-5 silent-drop trap already recorded fleet-wide. My
dispatch (corr `ce13e13c`, for the new URL landmine) DID land — 2 orchestration
rows, `EXECUTING_STEP` — so it got lucky, and "it worked for me" is exactly how
this trap survives. **Not fixed: not this lane's file**, same call the previous
session made about `076_improvement_loop_trigger.sh`. The fix is mechanical
(payload into the container COMMAND, `&& echo PUBLISH_OK`) and
`fire_improvement_loop_dartsonline.sh` in this session's scratchpad is a working
template for whoever picks it up.
