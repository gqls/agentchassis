# NOTES — bugfix 311 (append-only, newest at the bottom)

## 2026-08-19 — session start, research pass

- Bug still VALID [MEASURED 2026-08-19]: class counts re-run — 26 active base section
  rows with NULL section_type (unchanged), 89 with section_type=function, 26 differing
  (was 25), 80 tool-level base (was 79). Live failures through 2026-08-18 23:26 UTC
  (`needs_new_component` failed items for loans-settlement/overpayment/credit-health/
  standard-calc/compare-loans/interest-rate-stress-test/car-finance).
- No fix in flight elsewhere: `who-owns` → portfolio_positioning lane documented it and
  explicitly parked the fix pending "a council round and the owner's say-so"
  (HANDOFF_2026-08-18c §4); no uncommitted platform-code changes touch the three files;
  no open work item routes at the fix.
- **Mechanism refinement (the load-bearing find):** Path 1 of plan_sections already
  matches by function (`loadSectionComponents` pass 2, no level filter). The incumbents
  are dropped there by `sectionTemplateValid` (no `</section>`; they are tool-shaped,
  `created_from='manual'`, ending `</script>` — measured on all three). The drop, not
  the NULL section_type, routes the flow to the selector. Independent artefact:
  work item `3d775f99` (2026-08-15) defers with "stored component 824e3309 … failed the
  template guard" — the same rows, named by the guard.
- Corollary: bug file fix candidate 2 (backfill section_type) is REFUTED for the
  guard-dropped subclass (creates a `selector_error` → "ready as-is" degrade) and a
  NO-OP for guard-passing rows (Path 1 already resolves them by function). Details in
  PLAN §"The mechanism".
- 090 filed on the refinement (per the 2026-07-31 owner ruling — durable structural
  claim): intake `1306e72c-c725-4c3b-b0c3-8a63137f35fb`, run corr
  `f1433782-6ba7-4304-a7f9-8bd830dfb7c9`. Verdict pending at time of writing.
- Adjacent machinery read before designing: CLC-004 (pre-generation field-name
  preservation — dormant for NULL-section_type rows, which is why every generation drew
  fresh field names and the guard kept firing), CLC-006 (regen-vs-create keyed on
  LLM-chosen function — this bug is its cross-site worst case), RFC_034 (DECIDED
  2026-08-17; convert-by-id programme; adjacent, not conflicting),
  `check_unresolved_sections` (the retry engine that kept re-arming the failing pages;
  converges once a valid component exists).
- `input_data` for component-creator runs carries `site_id` AND `domain` (verified on
  run for item `7a2219bc`) — the diversion needs no new plumbing.
- Schema facts that shaped the design: `content_components.name` UNIQUE (global);
  `function` unique only for tool-level base active rows; no site ownership column.
- Live chassis: v1.0.1314, pods restarted 2026-08-19 ~08:00 UTC — last night's plan
  logs gone; provenance line scrolled (expected; not evidence of anything).

### Missteps / corrections this session

- (none yet beyond inherited ones; the candidate-2 refutation is a correction to the
  BUG FILE's account, recorded there and in WRONG_CALLS.md once the 090 verdict lands.)

## 2026-08-19 — build, tests, submission (same session, later)

- 090 refinement run FAILED on infrastructure at the `verdict` step: Anthropic API
  status 400 "You have reached your specified API usage limits" — the neighbouring
  diagnosis run (`6f900e18`, another lane) failed identically at 10:25 UTC, so it is the
  fleet cap, not this run. Intake `c1d726ad` reset to `triaged` for automatic re-claim.
  Substitute first-hand verification stated in the bug file per the 2026-07-31 ruling.
- **Cross-lane arrival mid-session:** commit `73c2505e2` added to the 311 bug file —
  RFC_036 is the same wall at TOOL level (`create_tool_component`, unique index) with a
  forked_from remedy. Read RFC_036: OPEN, no code proposed, owner holds a contained
  interim. Decision: this round fixes the SECTION writer only, stated plainly in the bug
  file with the reason the remedies differ (a section-level fork would be invisible to
  every selection path; a tool-level fork escapes the partial unique index and deploy
  links pages itself).
- Built: `component_storage_identity.go` (+ tests), wiring + section_type self-heal in
  `store_generated_component_action.go`. Full actions suite green; mutation proof run
  (diversion deleted → foreign-collision test fails on uncovered incumbent UPDATE →
  restored); `go build ./...` clean (one PRE-EXISTING vet finding in
  `load_component_library_actions.go:207`, untouched file, not ours); archive-HEAD
  overlay build + full suite green (my commit cannot break shared HEAD).
- Council submitted: corr `fc3ac5f4-ee3a-4e27-88ab-a8b2536b2c1d`. The round may fail on
  the same API cap — see RUNBOOK before re-firing anything.
- Register CLC-020 appended; LANDMINES 311 entry corrected in place (the "invisible to
  the selector" mechanism was incomplete — Path 1 + template guard is what hides the
  incumbents); WRONG_CALLS row for the refuted candidate 2.
- **Declared same-file passengers at commit time:** `component-lifecycle.md` carries the
  283 lane's uncommitted CLC-021/CLC-022 entries + a CLC-017 correction; `LANDMINES.md`
  carries their new instance-scope-bindings entry. Both complete and coherent; declared
  in the commit message rather than waiting (their entries say "committed this session",
  i.e. they are mid-cycle; a pathspec commit takes the whole file either way).

## 2026-08-19 — council verdict: APPROVED round 1 (corr fc3ac5f4), 4 advisory objections, none high

12 reviewers, 5 abstained, no truncation gating. Full report:
`diagnosis_artifacts kind='council_report' correlation fc3ac5f4`. Triage of every advisory,
with what was DONE about each:

- **editquality medium — `domainSlug` asserted, not verified:** answered by the shipped code
  (the council judges sketches): `domainSlug` exists in `deploy_tool_action.go` (same
  package, read this session: `strings.ReplaceAll(domain, ".", "-")`); the committed helper
  compiles and its tests pass against it.
- **editquality low — LogActionFindings provenance merge:** the diversion finding inherits
  the running step's provenance (component-creator), which IS the diverting actor — intended.
  The RUNBOOK's post-roll demand check reads the actual row; verify provenance columns there.
- **editquality low — "register entry ships in same commit" not an edit:** it did ship in
  `17d883333` (CLC-020 + index row in `0bc2d6162`). Claim was true; council could not see it.
- **bug_historian medium / prior_art medium — RFC_036 divergence asserted, not demonstrated:**
  now MEASURED and recorded in RFC_036 §10 (appended this session, the architecture seat's
  tracking ask): `idx_cc_tool_function_unique` read live — UNIQUE btree(function) WHERE
  component_level='tool' AND forked_from IS NULL AND is_active — partial on forked_from,
  which is exactly why the tool fork escapes it and a section fork would just be invisible.
- **bug_historian low — census read error degrades silently:** it degrades with a Warn
  ("dependent census unreadable — proceeding as legacy regeneration") and the field guard
  as backstop; behaviour equals today's, never below. Accepted as designed.
- **reuse_agent medium / architecture medium — two parallel collision conventions:** tracked
  in RFC_036 §10 as asked; §9.3's implementer is pointed at `foreignDependents` for reuse.
- **reuse_agent low — census may duplicate an existing pattern:** nearest sibling is
  `markPagesPendingRebuild`'s site enumeration (same file) — different question
  (affected-sites-for-rerender vs foreign-ownership with requester exclusion + domains).
  Noted; not merged, deliberately: the rerender one has no exclusion and no site_components.
- **guardian medium — "only caller" asserted:** now MEASURED: exactly one live
  agent_definition names the action (`component-creator`, active; snapshots/deleted
  excluded), and zero direct Go callers outside tests (grep). Recorded here.
- **guardian medium — concurrent-divert race:** two racers mint the SAME suffixed name;
  `content_components_name_key` (global UNIQUE on name) makes the second INSERT fail LOUDLY;
  its retry finds the row and takes the own-site regeneration path. Self-converging, visible,
  no silent duplication. (A silent duplicate would need different names for the same
  function+site, which the deterministic suffix rules out.)
- **guardian low — COALESCE self-heal is scope creep:** acknowledged in the plan's risks
  block pre-verdict; accepted.
- **debug_historian low — no deploy-verification step:** RUNBOOK carries the binary probe
  with positive AND negative controls; council could not see lane docs.
- **debug_historian medium — no re-trigger for already-parked items:** convergence path
  exists without manual action and is now stated in the bug file: the failed
  `needs_new_component` items are TERMINAL, so `CreateNeedsNewComponentItem`'s dedup
  (status NOT IN terminal) permits a fresh item; loanzy's pages already sit at
  `needs_rebuild`; `check_unresolved_sections` keeps re-arming deployed pages because the
  incumbent matches by function — the exact loop that today livelocks becomes the
  convergence engine once the diverted row exists. RUNBOOK also scripts a manual re-drive.

## 2026-08-19 — the re-run 090 verdict: UNVERIFIABLE (iteration-cap), with one fair nuance

The re-armed diagnosis (dispatch corr `a8ec7411`) completed `UNVERIFIABLE` — stopped at the
iteration cap, explicitly "NOT CONFIRMED", not refuted, "hand to a human with the full
trail". Its last hypothesis raises a branch nuance my contribution should carry: work item
`3d775f99`'s "repair that component; do not create a new one" message is emitted by
**Path 0** (the stored-slot `byIDDrops` branch — loancalculator's page PINS incumbent
`824e3309` in `page_components`), not by Path 1's name lookup. That does not contradict the
refinement — `3d775f99` was cited as proof the GUARD DROPS these templates at load (the
loop: "the load-time guard-drop mechanism itself may still be real"), and the greenfield
sites (loanzy, remortgagecalculator — no stored pin, so Path 0 is skipped) are the ones
whose artefacts show the Path-1-drop → selector → `needs_new_component` route. Two branches
consume the same drop: pinned slots defer LOUDLY (human review), unpinned ones fall through
to the deadlock. The corollary (backfill → selector_error → "as-is") remains a code-read
claim, cited line-by-line, unrefuted and unconfirmed by the loop. Status of the record:
the refinement stands on the stated first-hand verification (2026-07-31 ruling); the loop's
independent check ran out of budget before reaching a verdict, twice for different reasons
(API cap, then iteration cap) — recorded here so nobody reads "090 was run" as "090
confirmed it".

## 2026-08-19 — OWNER RULING arrived (via portfolio_positioning lane, cross-session): scope widens to BOTH writers, as a PRECONDITION

Verbatim intent: **"For the calculators — one submission covering both writers and treat it
as a precondition."** Meaning: (1) ONE council submission covering both
`store_generated_component` (done: `17d883333`, APPROVED `fc3ac5f4`) AND
`create_tool_component` (RFC_036 §9.3 — NOT YET BUILT); (2) the pair gates the ~50-site
portfolio wave 1. The ruling arrived AFTER the section-half round was approved, so
compliance path (forward-only, cannot unsubmit): build §9.3 now and submit a round whose
rationale explicitly names both writers, cites this ruling and `fc3ac5f4`, and frames the
two rounds as one logical change — no half may ever read as the whole. §9.5's
honest-fallback line restated in RFC_036 (the peer's ask); my duplicate "§10" header
renumbered to §11. Technical trap the peer flagged for §9.3: its lookup predicate is
`component_level='tool' AND forked_from IS NULL` — a LIBRARY row; a section-level incumbent
claims nothing in that predicate (correct for the index, but never claim §9.3 covers 311).
Peer's offer: their pilot sits frozen under a build-halt as a stable failing specimen for
on-demand queries — use it for post-roll verification. THE §9.3 IMPLEMENTATION IS THE NEXT
WORK ITEM IN THIS LANE and is not started at the time of this note.

## 2026-08-19 (afternoon) — THE FIX IS LIVE on v1.0.1315, binary-proven with controls

Asked by the loanzy_uk lane (their probe for the fix sha came back ABSENT — correctly
distrusted: the stamp is the BUILD commit's sha, so an absent ancestor sha proves nothing).
The definitive chain, run 2026-08-19 ~16:00 UTC:
- Both v1.0.1315 replicas (started 12:15Z) stamp `590ca3a20cca99e0f6e9c6a2545bd8e94c11b9ae`
  in /proc/1/exe; fake-sha negative control ABSENT on the same pod.
- `git merge-base --is-ancestor 17d883333 590ca3a20` → TRUE. Fix is in the build.
- Second positive probe on the fix's own literal: `COMPONENT_COLLISION_DIVERTED` PRESENT.
- Behavioural evidence: none yet — ZERO needs_new_component attempts fleet-wide since the
  roll (loanzy lane measured), so this is a no-demand zero; the first real build is the test.

Incumbent baselines PINNED before any real-world test (both-halves verification —
values, not refs):
- b89f91e1 (mortgages-repayment): html a2c00f1c66ce6f4ef72b48083f1e3da6 / schema 8265ae5a931b735305b1fe007b148acb
- 7d8b0503 (loans-car-finance-calculator): html 5f9534982e7f2bd776605ed78e755010 / schema 8e2cfe0afb1863b178390d6a048409b0
- 824e3309 (loans-credit-health-check): html e6ee4b07f11d0b43c1c5a62667f4999f / schema dd8f9863c84f8a5a7ec3e99154241f43

The loanzy lane will re-run a build as the free real-world test and report either way.
Owner context relayed by them: ALL sites must be capable of having tools — "pick a
vertical without calculators" is off the table as a workaround.

## 2026-08-19 (late afternoon) — the TOOL-writer half BUILT + SUBMITTED (the owner's precondition pair is now code-complete)

- Answered the loanzy lane's gating question first: their clean-domain test is NOT premature —
  their path is entirely the section writer (live); create_tool_component only runs on the
  add_tool route their builds never trigger. Two caveats sent: the fix guarantees storage +
  linkage, not LLM generation quality (their zero-`<input>` served-page check stays); and the
  tool half still gates wave 1.
- Built RFC_036 §9.3 as written: library-claim lookup (the idx_cc_tool_function_unique
  predicate verbatim, so the check and the index cannot drift) → INSERT carries
  `forked_from = <library id>`; no claim → byte-identical write (pinned); lookup read error →
  fail-OPEN to today's loud index refusal (pinned). Mutation proof run: fork wiring deleted →
  `TestCreateToolComponent_LibraryClaimForcesFork` failed on "argument 9 expected [library id]
  … actual [nil]" → restored → suite green. Full actions suite + archive-HEAD overlay green.
- `expectPreamble` in the 286 lane's adopt tests updated to pin the new lookup — without it
  those ordered walks would "pass" by silently exercising the fail-open branch.
- Council round `ceae30f2-b03f-42b4-a5dc-a07837d7bbe0` submitted, rationale naming BOTH
  writers + the ruling + fc3ac5f4's state, per the one-submission ruling. Verdict pending.
- The pair's state: section half LIVE (v1.0.1315, zero real exercises); tool half committed
  alongside this note, inert until the next roll. Wave 1's precondition is code-complete,
  pending the ceae30f2 verdict + a roll + the loanzy real-world test.
- Post-commit advisory triage (`e24bc9c0f`): pattern-check flagged
  `unrepaired-component-write` (page_components.rendered_html with no link repair) in
  create_tool_component_action.go — PRE-EXISTING (3 occurrences in the parent commit,
  zero in this commit's hunks); that class is `bugs_open/136`'s, owned elsewhere. Noted,
  not acted on: pulling an unrelated behaviour change into this round would be scope creep
  on a council-submitted change.

## 2026-08-19 — ceae30f2 verdict: APPROVED round 1 (tool-writer half; 10 reviewers, 3 advisories none high)

Triage, with the measurement each advisory asked for:
- **bug_historian medium — prove fork-invisibility is irrelevant at tool level:** MEASURED:
  16 active tool-level FORKS already exist beside 84 base rows (deploy_tool_to_site has been
  minting them all along), so §9.3 adds rows to an ALREADY-POPULATED class every tool reader
  already lives with; and a repo-wide grep finds exactly TWO code sites pairing
  `component_level='tool'` with `forked_from IS NULL` — my new lookup and deploy_tool's
  library-vs-fork definition — i.e. the fork mechanism itself; no discovery/health/
  eligibility path (check_tool_health, check_missing_tools, …) filters tool rows on
  forked_from. Whatever treatment deploy forks get today, generated forks get identically.
- **reuse_agent medium / guardian medium / architecture low — the double-copy tail needs a
  TRACKED follow-up, not an accepted risk:** now tracked concretely in RFC_036 §11 (this
  commit): widen deploy_tool_to_site's existing-fork lookup from `forked_from=$1 AND name=$2`
  to also match the generator's name shape (`<function>-<slug>`), so the two fork producers
  recognise each other's copies. Follow-up, not this round: it changes deploy behaviour and
  deserves its own small round if/when the tail fires (it fails loudly at the page layer
  today — pages_site_id_name_key).
- **guardian low — forked_from's second meaning (site copy, not byte lineage):** stands as
  stated in the risks block; RFC_036 §9.3's own definition; consumers that assume byte
  lineage would misread deploy forks equally (they copy then drift via rerender).
- **debug_historian low — name the deploy-verification recipe:** added to the RUNBOOK below
  the section-half recipe: same two-control probe, ancestry on `e24bc9c0f`, and the
  demand signal is the §9.3 Info log line / a successful save_tool for a function the
  library claims (tool-ab-test-calculator is the natural first case).

## 2026-08-19 ~16:20Z — new session: the real-world test of the SECTION half is being driven from THIS lane

- State on pickup [MEASURED 16:15Z]: chassis still v1.0.1315 (both pods started 12:15Z) — the tool
  half (`e24bc9c0f`) has NOT rolled; `needs_new_component` items created/updated since the roll:
  0; `COMPONENT_COLLISION_DIVERTED` findings: 0; scoped `*-loanzy-uk` rows: 0. The loanzy lane's
  16:08Z page touches were another lane's meta-description backfiller, not a build. Neither the
  loanzy lane (no next domain chosen) nor the portfolio lane (sites locked under the owner's
  halt) has run the agreed test, so a no-demand zero is all we have. Driving it here, per the
  PLAN's own verification section.
- Specimen chosen: `loans-car-finance-calculator`, NOT `loans-credit-health-check` — both
  loanzy attempts at credit-health-check died UPSTREAM of the store (`generate_template`,
  `stop_reason=max_tokens`, 47k chars at the 16000 cap), so it cannot exercise the diversion;
  car-finance generated fine three times and was refused at the store each time (the
  10-field contract guard). Incumbent `7d8b0503-0446-456c-a85e-d398264e8136`, dependents =
  loanandmortgagecalculator.co.uk only (1 page_components row, 0 site_components).
- Baselines RE-PINNED 16:18Z for all seven incumbents (values, not refs) — all three previously
  pinned are UNCHANGED; the other four added:
  7d8b0503 car-finance html 5f9534982e7f2bd776605ed78e755010 / schema 8e2cfe0afb1863b178390d6a048409b0 ·
  9cbfe279 compare-loans f3fd6e9cc9980c2a2eb971e9484cbbb0 / 3bba8e7d9d13338ea0370971f9ef487c ·
  824e3309 credit-health e6ee4b07f11d0b43c1c5a62667f4999f / dd8f9863c84f8a5a7ec3e99154241f43 ·
  2cf33f06 stress-test 6ca2074d036c177920101a1a3f97c46b / a805b2af699f1c28a9d7833ff35405e6 ·
  b7a499f4 overpayment 5291eedc497dc1cd5338d55f1cf217b7 / fd2a6336dd159833892afdad62863f19 ·
  70b72b3e settlement 66ea98791a002de25870cf072e1d313f / b7a1e6090d00f0bc1f17178d9ade3a45 ·
  b420389f standard-calc 85d673794c5ac0638595817760b59794 / a5790bcfeb1d46da94cb8ef3d9fc5fdc ·
  b89f91e1 mortgages-repayment a2c00f1c66ce6f4ef72b48083f1e3da6 / 8265ae5a931b735305b1fe007b148acb.
- Re-drive mechanism: a FRESH item, not a reset — `idx_swi_dedup` excludes `failed`, and
  `CreateNeedsNewComponentItem` inserts `status='triaged'`, `handler_agent='component-creator'`,
  `pipeline='build'`, priority 50; mirrored by hand with `created_by='bugfix_311_redrive'`
  (dispatch filters on status/attempts/approval/pipeline/handler, never created_by; `source`
  kept `component_selector` because it becomes `component_versions.change_source`).
  Item **`9d16951e-439d-4c8c-8c9c-8ee9251d83e6`**, 16:20:04Z. Pre-flight: loanzy `locked_at`
  NULL, no claimed items on the site, no other unlocked dispatchable site ahead of it,
  build-pipeline-trigger ticking every 60s (last 16:19:17Z).
- Caveat stated BEFORE the result: the diversion only fires if the LLM picks the incumbent's
  function name again (CLC-006: keyed on the LLM-chosen function). `loans-loan-vs-savings`
  succeeded yesterday precisely because the LLM chose `loan-vs-savings` — a plain creation,
  not a collision. If this run creates a fresh-named row, that is a pass for loanzy but NOT a
  diversion exercise; read `diverted_from` in the result / the finding row, not the status.

## 2026-08-19 16:23Z — RESULT leg 1+2: THE DIVERSION FIRED ON A REAL CASE, and no incumbent moved

- Item `9d16951e` claimed 16:21Z by build-dispatch-loop, **complete 16:23:00Z on attempt 0**
  (child orchestration `c399e282`). `stored_component`: `status=created`,
  `requested_function=loans-car-finance-calculator`, `function=loans-car-finance-calculator-loanzy-uk`,
  `diverted_from_component_id=7d8b0503-0446-456c-a85e-d398264e8136`, `section_type=loans-car-finance-calculator`,
  quality_score 100, template 13,738 chars, has_js. So the LLM DID pick the incumbent's name
  (the caveat above did not bite) — this is a genuine collision exercise.
- New row `2e497429-b2de-46f6-9e1e-9799b33912a3`: base (`forked_from` NULL), active,
  `created_from='generated'`, section_type = the request vocabulary, html contains `</section>`
  (so `sectionTemplateValid` will NOT guard-drop it — unlike the incumbents). The selector's
  own predicate (`section_type='loans-car-finance-calculator' AND component_level='section' AND
  is_active AND forked_from IS NULL`) returns EXACTLY this one row [MEASURED].
- Durable record: ONE `agent_error_log` row, `error_code='COMPONENT_COLLISION_DIVERTED'`,
  severity warning, agent_type component-creator, step store_component, work_item_id 9d16951e,
  context {incumbent_id, requested_function, final_function, section_type} — the council's
  editquality-low provenance question answered at the row: provenance is the diverting actor.
- **No collateral damage [MEASURED 16:24Z]**: all EIGHT incumbents' `md5(html_template)` and
  `md5(input_schema::text)` equal the 16:18Z baselines; 7d8b0503's `updated_at` is still
  2026-08-13 14:18:58Z.
- Loud-failure check: the `error` column on the item is empty; no SQLSTATE 23505; no
  `selector_error`.

### CORRECTION to my own earlier claim (this file, "council verdict" entry, debug_historian medium)

> **CORRECTED 2026-08-19 20:00Z:** I wrote that `check_unresolved_sections` "keeps re-arming
> deployed pages … the exact loop that today livelocks becomes the convergence engine once the
> diverted row exists." Half true. The check DOES flip the page to `build_status='needs_rebuild'`
> (loanzy `tool-car-finance-calculator` sits there now), but **nothing consumes `needs_rebuild`
> on its own**: the `page-rebuild` agent that reads it has NO scheduled task and ZERO
> orchestrations in history, and `seed_build_queue`/`write_build_items` key on `planned`. The
> page's two real builds were both `needs_page` items → `page-build-handler` (the second one
> filed by `image-build-handler`'s `flag_page_image_rebuild`, key `page_rerender:<page>`). So the
> convergence path is: a `needs_page` item must be FILED (by image-landed, reconcile, or by
> hand) — the diverted row makes that build succeed, it does not cause the build. Caught by
> reading the consumer (`grep needs_rebuild` → page-rebuild → scheduled_tasks: none) before
> claiming the page would heal; memory's "needs_rebuild is a DEAD queue" (bugfix 161) was right.
> What this means for the 6 other loanzy tool pages: each needs its own needs_new_component
> re-drive AND a needs_page re-render — or the loanzy lane's full build, which files both.

## 2026-08-19 20:04Z — leg 3 in flight: needs_page re-render filed for the car-finance page

- Mirrored `flag_page_image_rebuild`'s shape exactly: `needs_page`, spec
  `{"reason":"bugfix_311_redrive_relink","page_name":"tool-car-finance-calculator"}`, priority
  99, handler `page-build-handler`, status triaged, item_key `page_rerender:tool-car-finance-calculator`
  (previous holder `25c73782` is complete → dedup permits). Item **`1f8e8563-0ff1-485b-8c35-ace3ad05bd1c`**.
- Served-page baseline BEFORE [MEASURED 20:05Z]: `https://loanzy.uk/tools/car-finance-calculator/index.html`
  → HTTP 200, 25,703 bytes, **0 `<input>`**, md5 42b9e15d8ce02c09ff013016a587bc2a. The pass
  condition for leg 3 is a `page_components` row linking `2e497429` with build_status deployed
  AND a served page with >0 `<input>` — the second independently of the first (loanzy lane's
  protocol: stored+linked ≠ a good calculator).
- 20:06Z: item `1f8e8563` claimed; page-build-handler `214074b9`. **plan_sections resolved the
  slot [MEASURED from collected_data.section_plan]:** `name=loans-car-finance-calculator`,
  `status=ready`, `component_id=2e497429…`, `function=loans-car-finance-calculator-loanzy-uk`,
  render_mode agent — the selector-visibility claim (section_type unsuffixed → Path 2 finds it)
  is now demonstrated on a real build, not asserted from the predicate. No new
  needs_new_component item filed. Content writer running.

## 2026-08-19 20:16Z — RESULT leg 3: the page links the diverted row AND serves a real calculator — ALL THREE LEGS PASS

- page-build-handler `214074b9` COMPLETED 20:16:13Z (steps: plan_sections → content writer
  `60dc0d61` → save_sections → rerender → deploy_page). Item `1f8e8563` complete, attempt 0.
- `page_components` for `41ad4a72` (tool-car-finance-calculator) [MEASURED 20:17Z]: position 2 =
  **`2e497429` (loans-car-finance-calculator-loanzy-uk), build_status deployed, rendered_html
  14,002 chars**; hero-tool / generic-text-block / faq re-linked around it. `pages.build_status`
  = deployed (was needs_rebuild).
- Served artefact, independently [MEASURED 20:17Z]: `https://loanzy.uk/tools/car-finance-calculator/index.html`
  → 200, **38,912 bytes (was 25,703), 4 `<input>` (was 0)** — `#car-price`, `#deposit`,
  `#loan-term` range sliders + `#interest-rate` number; md5 12b488fcd83cf1b942df4c30574180f3
  (before 42b9e15d…). Zero `{{` / Go control syntax (the 260 class did NOT fire here). The
  extracted JS reference carries the FINAL (suffixed) name — `/tools/assets/loans-car-finance-calculator-loanzy-uk.js`
  → 200, 3,516 B = the store result's `js_size` — which is the E2 ordering point (identity
  resolved before `separateInlineJS`) demonstrated live.
- Incumbents re-read AFTER the whole run: all eight md5 pairs unchanged, `updated_at` untouched.
- Cost: one component-creator run + one page build. Wall clock: 16:20→16:23 (store) and
  20:04→20:16 (page), the gap between being this session's usage-limit pause, not the system.

**Verdict for the SECTION half: fixed AND live AND exercised on a real collision, all three
protocol legs, with the "before" pinned.** What keeps 311 OPEN: the owner's precondition pair —
the TOOL half (`e24bc9c0f`, council ceae30f2 APPROVED) has not rolled (chassis still v1.0.1315
at 20:17Z); and the other six loanzy tool pages are still hollow until each gets a re-drive +
re-render (recipe in RUNBOOK) or the loanzy lane's full rebuild. Not done here, deliberately:
the PLAN's verification scope was one item, and the rest is the loanzy lane's site and their
stated "next run" — pointed at, not pre-empted.
- (The needs_rebuild dead-queue trap already has a LANDMINES entry — "A data repair RACES the
  sweep that publishes it", trap 1, footprint `pages.build_status` — so no new entry; the
  WRONG_CALLS row for my claim points there.)

## 2026-08-19 20:35Z — v1.0.1316 IS LIVE and carries BOTH halves (binary-proven, both replicas, controls both ways)

- Pods `agent-chassis-5ddd9744-86nqf` / `-8jlqh`, image v1.0.1316, started **17:13Z / 17:14Z**.
  Provenance line already scrolled (>5000 lines) → candidate-sha probe: stamp
  **`07eeba4a1eecbe809f518b5d0b7f9fc5f75e71ed`** (2026-08-19 17:21:10 +0100, "handoff(320)…")
  PRESENT in /proc/1/exe on BOTH replicas; fake sha `deadbeef…` ABSENT (negative control);
  `git merge-base --is-ancestor e24bc9c0f 07eeba4a1` TRUE and `… 17d883333 07eeba4a1` TRUE;
  literals `library tool claims this function` (tool half) and `COMPONENT_COLLISION_DIVERTED`
  (section half) both PRESENT. **The owner's precondition pair is LIVE.**

> **CORRECTED 2026-08-19 20:40Z — a stale figure carried forward.** The 20:16Z entry above, the
> bug-file status line, and the portfolio/loanzy contribs said "chassis still v1.0.1315 at
> 20:17Z". I had measured the pods ONCE, at 16:15Z, and repeated it four hours later without
> re-running `kubectl get pods`. The roll happened at 17:13Z. Consequence for the evidence: the
> store diversion (16:23Z) ran on v1.0.1315, but the page rebuild (20:06–20:16Z) ran on
> **v1.0.1316** — fine for the claim (the section code is in both), wrong as a statement of
> fact. Caught by the owner saying a build had rolled; the cheap check is CLAUDE.md's own:
> re-run the status snapshot before acting on it — a pod list is a snapshot like `git status`.
> WRONG_CALLS row added.

- Tool-half DEMAND test: NOT run from this lane, on purpose. The natural specimen
  (`tool-ab-test-calculator`, webdesign.co.uk) belongs to the `webdesign_tool_rebuilds` lane,
  which is ACTIVE right now (its NOTES last entry 20:35Z, `page_rerender` items filed 20:36Z)
  and whose handoff still reads "RFC_036 §9 — nobody has built it / Phase D runs whenever
  RFC_036 lands". Phase D is now unblocked; told them (CONTRIB in their dir, this commit) with
  the assertion list and baselines. Firing an `add_tool` on a site another session is
  actively dispatching at would contend on the one-site-per-tick dispatcher and muddle their
  artefact reads.
- Baselines PINNED for that test [MEASURED 20:38Z]: library row `8c9a6e06-e2b2-4f21-baf6-651585375f0c`
  (`tool-ab-test-calculator_pre_037`, base, active) html `8673be08f969504f5a9ceb46e45d7656`,
  schema `688e1188b91ccef0674cd527daa05ec3`, updated_at 2026-05-06; existing forks: `cd60486c`
  (…-webdesign-co-uk, INACTIVE, html 8208eb17…) and `58da6570` (…-idea-uk, active, 2169c654…).
  Pass = a NEW row `forked_from='8c9a6e06…'`, `component_level='tool'`, name
  `tool-ab-test-calculator-webdesign-co-uk`, save_tool COMPLETES (no SQLSTATE 23505), the
  §9.3 Info log line, and 8c9a6e06's md5s unchanged. One thing for them to read first: the
  action's "already exists" check (`create_tool_component_action.go:228-237`) returns early if
  an ACTIVE tool component with that function is linked to one of the site's pages — so the
  ported slot's state at build time decides whether the fork branch is even reached.

## 2026-08-19 21:05Z — CONTRIB IN from the `webdesign_tool_rebuilds` lane: the tool-half DEMAND test RAN, and §9.3 PASSES

Fired from that lane (its own NOTES,
`docs/agent_docs/docs024_key_docs_latest/webdesign_tool_rebuilds/NOTES_native_rebuild_of_ported_tools.md`,
entries 20:52Z and 20:58Z, hold the full working). Specimen: `tool-ab-test-calculator` on
webdesign.co.uk, work item `3a3e480c-321c-46fb-adae-0d58c05bf2aa`, generator orchestration
`d6ce0591-6906-472a-8302-9a598b0f4789`, build completed **20:57:04Z**.

**Result: PASS on four of your five assertions, and the fifth is unanswerable for a reason worth
recording.**

1. ✅ New row **`8a315006-2170-4ba7-b517-4abaf9619e45`**, name
   `tool-ab-test-calculator-webdesign-co-uk`, `component_level='tool'`, `is_active`,
   **`forked_from='8c9a6e06-e2b2-4f21-baf6-651585375f0c'`**, `created_from='generated'`,
   `source_agent_type='tool-generator'`.
2. ✅ `save_tool` COMPLETED. No SQLSTATE 23505 on the item (`error` NULL) or the orchestration
   (`current_step=complete`, `status=COMPLETED`, `page_adopted=true`, `already_exists` NULL,
   `__step_error` NULL). The `create_result.component_id` equals the component the page actually
   carries, so this is provably the run that built the artefact.
3. ⚠ **The §9.3 Info log line could NOT be checked, and its absence is not evidence.** Both
   replicas return 0 matches for `library tool claims this function` — and the control says why:
   **`kubectl logs --since=6h` on either chassis replica returns lines beginning at 21:01:50Z /
   21:02:07Z, a retention window of ~2.2 minutes** (458 / 1,110 lines). The 20:57 build had already
   rotated out. Worth knowing for any future demand test on this service: **that grep is only valid
   inside ~2 minutes of the event, and it fails silently.**
   The DB is the better evidence anyway: `forked_from` is written from `libraryToolFork`
   (`create_tool_component_action.go:262-285`), which is nil unless your branch runs, and nothing
   else sets that column on a `created_from='generated'` row. The column proves the branch executed;
   the log line would only have proved it spoke.
4. ✅ **Library row `8c9a6e06` UNTOUCHED after the build**: html md5 `8673be08f969504f5a9ceb46e45d7656`,
   schema md5 `688e1188b91ccef0674cd527daa05ec3`, `updated_at` 2026-05-06 18:12:16 — all three
   identical to your 20:38Z pins. Both existing forks untouched.
5. ✅ Artefact: component 10,086 chars, `{{\.` 0, `onclick=` 0, `alert(` 0, zero bare hex, page
   gained the slot, ported slot retired 94 s later with its bytes byte-identical (md5
   `6b99651c11b7dbfa939c5296bdb5704b` before and after). Served-page grade follows when the page's
   rerender drains (`ad2a2dc4-4fbb-489f-9b74-bcd00e6f09ff`, still queued at 21:05Z).

**On your "one thing to read first" — it mattered, and the answer was favourable.** The
`already_exists` probe needs an ACTIVE tool-level component with that function linked to a page of
the site. At build time: `cd60486c` had a placement on the site but `is_active=false`; `58da6570` was
active but its only placement is on idea.uk; the library row is active with no placement anywhere. So
the probe returned nothing and the fork branch was reached. Anyone repeating this on a tool whose
local fork is still ACTIVE must deactivate it first (the RUNBOOK's "Before REFILING a tool that
already has a native component" step) or the run short-circuits and the fork branch is never tested.

**Consequence for your precondition pair: the tool half is now demand-proven, not just built.**
Phase D's second tool (`tool-meme-generator`, library claim `6ae53f32-be86-4c29-bc52-983c35d23b18`)
is the remaining one and is a Phase B rich app, so it goes later by the owner's standing instruction.
§11's known follow-up (`deploy_tool_to_site`'s fork lookup) did not bite: this route never calls it.

## 2026-08-20 ~08:00Z — new session: both halves RE-VERIFIED on v1.0.1317, the tool half's served page GRADED, and the residual measured honestly (it is not "six pages")

**Nothing named `HANDOFF_2026-08-19_continue_here.md` exists in this lane** — the pickup pointed at
one and it was never written (`git log` on the directory confirms it). Picked up from
`NOTES` + `RUNBOOK` + `README_where_we_are` instead. Worth one line in a future handoff commit:
a lane that never wrote a cold-start file sends its next session hunting.

### 1. Both halves are live on the CURRENT image, not just the one they shipped on

The fleet rolled again overnight: chassis **v1.0.1317**, pods `agent-chassis-c7d6d875b-{67cgh,x5tgn}`,
started **2026-08-19 22:26Z** — i.e. every version statement in this file below 20:35Z is about a
retired image. Re-verified at the binary rather than assumed forward [MEASURED 08:12Z]:

- stamp **`2d13d530d2943831641ff6e51e4c92d8eb4b6c10`** PRESENT in `/proc/1/exe`; the two other
  candidate shas from the same build window (`5022305cf`, `d4950c53c`) **ABSENT** — so the probe
  discriminates, which a single positive can never show on its own;
- `git merge-base --is-ancestor` TRUE for **both** `17d883333` (section half) and `e24bc9c0f`
  (tool half) against that stamp;
- capability literals `COMPONENT_COLLISION_DIVERTED` and `library tool claims this function`
  both PRESENT on **both** replicas; invented literal `zzzz_no_such_literal_311_control` ABSENT.

### 2. The tool half's LAST open assertion is now answered — at the served page

Yesterday's 21:05Z contrib closed four of five assertions and left "served-page grade follows when
the page's rerender drains". It drained (item `ad2a2dc4` complete 21:36Z) and the page grades
[MEASURED 08:15Z]: `https://webdesign.co.uk/tools/ab-test-calculator/index.html` → 200,
**16,172 bytes, 5 `<input>`, 1 `<button>`, zero `{{`**, ids `a-visitors` / `b-visitors` /
`a-conversions` / `b-conversions` / `panel-a` / `panel-b` / `results` / `verdict`.

**Discriminated, not just observed.** Three components have carried that page. Querying which one
holds the served markup: `a-visitors` appears in `8a315006`'s `rendered_html` (the forked row,
`build_status=deployed`) and in **neither** of the two `removed` slots (`a7daa5c5` the ported
original, `cd60486c` the old inactive fork). `id="verdict"` alone would NOT have discriminated —
the ported slot has it too. So the served calculator is provably the row that `forked_from` points
at the library claim, which is the whole of RFC_036 §9.3.

**Consequence: the owner's precondition pair is demand-proven on both halves, at the artefact.**

### 3. The residual is NOT "six hollow pages" — measured, it is four different things

Every loanzy tool page fetched and its `<input>` count taken [MEASURED 08:16Z], then cross-read
against the eight `failed` `needs_new_component` items. The RUNBOOK's list was close but wrong in
three ways, and the differences change what each page needs:

| class | pages | what it needs |
|---|---|---|
| **fixed** | `tool-car-finance-calculator` (4 inputs) | done 08-19 |
| **collision class — `store_component` died, 311's exact defect** | `tool-interest-rate-stress-test`, `tool-loan-comparison-calculator` (`loans-compare-loans`), `tool-loan-repayment-calculator` (`loans-standard-calc`), `tool-overpayment-calculator`, `tool-settlement-calculator` | re-drive + re-render (both legs) |
| **upstream `max_tokens`, NOT this bug** | `tool-credit-health-check` AND `tool-eligibility-checker` — *two* pages, both planned the same `loans-credit-health-check` section | a cap decision; see §5 |
| **render leg only** | `tool-loan-vs-savings` — its component was created fine on 08-19 (plain creation, no collision) and the page still serves **0 inputs** | one `needs_page`, no generation |
| **never planned a section** | `tool-compare-loans`, `tool-is-a-loan-right-for-me` (both 404, zero `page_components`) | not a 311 case at all |

Three corrections to this lane's own earlier record: the count was **five** collision-class pages
plus a render-only one, not six; `tool-eligibility-checker` was missed entirely (it is the second
victim of the `max_tokens` failure, same section type, different page); and `tool-loan-vs-savings`
was being counted as needing a re-drive when it only ever needed the render leg.

### 4. Two measurement traps hit while pinning the baselines, both worth the ink

**(a) An "incumbent untouched" control can be broken by a mechanism that is nothing to do with
you.** Re-pinning all eight incumbents before this run, `b420389f` (`loans-standard-calc`) had
MOVED: html md5 `85d67379…` → **`a9dea7cd…`**, 2,469 → 2,852 chars, `updated_at` **2026-08-20
07:02:57Z** — one hour before I looked, and after yesterday's pin. Attributed rather than
suspected: `component_versions` holds exactly one archived row for it, `change_source =
**scope_component_instance_judged**`, and the archived bytes ARE yesterday's md5. So a concurrent
component-scoping mechanism rewrote a shared incumbent this morning. Every other loans incumbent
has zero version rows. **Do not re-use a day-old baseline on a shared row, and if one has moved,
read `component_versions.change_source` before blaming your own run.**

**(b) A single 404 is not a baseline.** My first sweep read `tool-loan-comparison-calculator` as
**404**; two later reads, same URL from `pages.url`, both **200 / 22,600 bytes / 0 inputs**, and
nothing deployed in between (`max(page_components.updated_at)` for it is 08-18 22:31Z). Recorded
as an unexplained transient, NOT as a mechanism — but note which way it would have cut: had I kept
the 404 as the "before", the fix would have been credited with publishing a page it never touched.
Pinned baseline is the reproducible one.

### 5. Filed the owner's choice: the five collision-class re-drives (08:18Z)

Owner ruled "repair the 5 collision-class pages" (credit-health-check's pair and the two
never-planned pages left alone). Preconditions checked before inserting, not after: site
`55213ded` **unlocked**, zero claimed items, all five `needs_new_component:<section>` keys held
only by `failed` rows and all five `page_rerender:<page>` keys held only by `complete` rows —
`idx_swi_dedup` excludes both statuses (index definition read, not recalled), so a fresh insert
cannot 23505. Items, all `triaged`/`component-creator`/priority 50/`created_by=bugfix_311_redrive`:
`8ce63159` stress-test · `dcaead88` compare-loans · `a1daabc8` standard-calc ·
`9a249bfa` overpayment · `d1add6d3` settlement.

### 6. Interim result [08:26Z] — 2 of 5 diverted cleanly, 0 incumbents moved

| section | status | stored as | diverted from |
|---|---|---|---|
| `loans-interest-rate-stress-test` | complete, attempt 0 | `…-loanzy-uk` | `2cf33f06` |
| `loans-compare-loans` | complete, attempt 0 | `…-loanzy-uk` | `9cbfe279` |
| `loans-standard-calc` | claimed | — | — |
| `loans-overpayment-calculator` | triaged | — | — |
| `loans-settlement-calculator` | triaged | — | — |

The dispatcher takes one item per ~60s tick per site, so they serialise; ~1–2 min each.
**All five incumbents re-read after the two completions: md5s equal the 08:12Z pins**, including
`loans-standard-calc`'s post-07:02Z value `a9dea7cd…` — i.e. the diversion left even the row that
another mechanism had just rewritten alone.

### 7. `311` candidate 3 (the deploy gate) VERIFIED FIRST-HAND, and it is filable on its own

Read rather than inherited from the bug file. `loadSectionComponents`
(`platform/orchestration/actions/v3_site_actions.go:4936-4954`): every section name that resolves
to nothing gets a **stub** appended (`needs_llm:true`, empty description) behind a single
`logger.Warn("loadSectionComponents: stubs for unresolved sections")`, and the build proceeds.
`plan_sections` records the gap as `needs_section_data` (`plan_sections_action.go:2746`).
Readers of that item exist — `reconcile_section_data_action.go` re-attempts it,
`revalidate_review_queue_action.go` revalidates parked ones, `loadOpenSectionDataRequests`
(`plan_sections_action.go:2561`) reads it to avoid repeat LLM spend — **but none of them gates a
deploy**, so the accurate statement is "detection and repair exist; refusal does not", which is
sharper than 311's original "nothing reads it".

Census [MEASURED 08:24Z]: **47 open `needs_section_data` items fleet-wide; 12 pages are
`build_status='deployed'` while carrying one** — robot-hands.com 4, leopardessconsulting.co.uk 3,
ai-agent-orchestration.com 2, finetuning.uk / gaswholesalers.com / loanandmortgagecalculator.co.uk
1 each. (All 47 carry `spec->>'page_name'`, checked, so the join is sound.)
`discovery_checks/check_unresolved_sections.go` DOES detect the recovered case and flips the page
to `needs_rebuild` — the status this lane already established has **no consumer**, which is why
detection has never converted into a repair.

### 8. ⚠ CORRECTION 09:15Z to §7 above, WRITTEN 40 MINUTES EARLIER IN THIS SAME SESSION — a deploy gate DOES exist, and I asserted a negative from a one-symbol grep

> **CORRECTED 2026-08-20 09:15Z.** §7 says *"none of them gates a deploy … detection and repair
> exist; refusal does not"*, and the bug file and its commit message carry the same sentence. **It
> is wrong.** `UpdatePageStatusAction` (`v3_site_actions.go:819-960`) refuses the `deployed` stamp
> **twice**: once when `pageHasComponents` is false (flip to `needs_rebuild`, clear
> `built_from_plan_version`, from the gamesdesign 0-component case) and again when
> `pageSectionShortfall` reports `rendered < planned` — `bugs_open/040`'s partial-build case,
> whose own comment states the rule I claimed did not exist: *"A partial build must be treated
> exactly like a 0-component one."* `bugs_open/210` widened it to any assembly skip, with a
> bounded retry and an `agent_error_log` refusal row.
>
> **What caught it:** going to write the bug file, and reading the deploy path instead of citing my
> own summary. **The cheap check I skipped:** I grepped for readers of `needs_section_data` and
> concluded "nothing gates the deploy" — but a gate does not have to mention the work item, and
> this one does not. The refuting code was two functions from something I had already read this
> session (`loadSectionComponents`, §7). **A negative about "nothing anywhere does X" cannot be
> established by grepping one symbol.** Also note which way it cut: it made the estate look
> *worse* than it is, and a wrong claim in that direction is just as expensive — it would have
> spent a bug number and a reviewer's round on a gate that shipped weeks ago.

**What survives, and it is narrower and sharper than what I wrote.** Both gates compare **counts**:
`pageSectionShortfall` (`v3_site_actions.go:1214-1233`) is `count(sections) - count(suppressed)`
versus **`count(*) FROM page_components`** — every row, whatever component it points at, whatever
its `build_status`. So the gate can see a slot that is EMPTY and cannot see a slot filled by the
WRONG THING. And `loadSectionComponents` guarantees the second shape rather than the first: an
unresolved section name is given a **stub** (`needs_llm:true`), which is what keeps the count whole.
That is a plausible mechanism for the owner's original symptom — a page that ships without its
calculator and is stamped `deployed` — **and it is a hypothesis, not a finding.**

**My census cannot carry it, either** [checked 09:12Z]: of the 12 "deployed pages with an open
`needs_section_data`", the cleanest specimen (`finetuning.uk` / `password-entropy`, planned 1,
rendered 1, gap = that very section) turns out to have **the right component** —
`tool-password-entropy-finetuning-uk`, a tool-level fork — sitting in the slot at
`build_status='pending'`. The item is STALE, not a live hole. Two more
(`leopardessconsulting.co.uk` case-study, planned 4 / rendered 0 excluding `removed`) are deployed
with every slot removed AFTER the stamp, which no stamp-time gate can catch. **So "12 pages are
serving a hole" is not established, and I am not filing it.** An open work item is not a live
defect; the artefact is.

**Therefore: no bug filed for candidate 3 today, deliberately.** What a filer needs next, in order:
(1) whether a stub actually becomes a `page_components` row on a real build — read
`save_page_sections`' writer, or watch one build; (2) a per-page artefact check (does the served
page contain the planned section's markup) rather than a work-item join; (3) then `090`, because
the claim is structural and CLAUDE.md's default applies. Candidate 3 stays a named residual on
`311` with this framing, which is why 311 is NOT being moved to `bugs_closed/` today even though
its titled defect is fixed, live and demand-proven on both halves.

## 2026-08-20 09:35Z — RESULT of the five-page repair: 5 of 5 diverted, 3 of 5 pages now serve real calculators, and the two misses are both OTHER mechanisms working correctly

### Component leg — 5 of 5, clean

All five complete on **attempt 0**, all five diverted, one per incumbent, and **all five incumbents
byte-identical** to md5s re-pinned at 08:12Z:

| section | new row | diverted from | new template |
|---|---|---|---|
| `loans-interest-rate-stress-test` | `950ac9db` | `2cf33f06` | 14,546 chars |
| `loans-compare-loans` | `3b08b9e9` | `9cbfe279` | 16,759 chars |
| `loans-standard-calc` | `2b2c79a8` | `b420389f` | 15,050 chars |
| `loans-overpayment-calculator` | `dc808c49` | `b7a499f4` | 17,149 chars |
| `loans-settlement-calculator` | `95788047` | `70b72b3e` | 12,129 chars |

Every new row: base (`forked_from` NULL), active, `section_type` = the request vocabulary,
`created_from='generated'`, and contains `</section>` so `sectionTemplateValid` will not guard-drop
it — unlike all five incumbents. `COMPONENT_COLLISION_DIVERTED` findings now total **6**.

### Page leg — 3 landed, graded at the served artefact against the 08:17Z pins

| page | before | after |
|---|---|---|
| `tool-loan-comparison-calculator` | 200, 22,600 B, **0** `<input>` | 200, **42,791 B, 6** `<input>` |
| `tool-overpayment-calculator` | 200, 23,705 B, **0** | 200, **42,089 B, 5** |
| `tool-settlement-calculator` | 200, 18,205 B, **0** | 200, **32,151 B, 5** |

Zero `{{` on all three. `page_components` for loan-comparison: position 2 = `3b08b9e9`
(`loans-compare-loans-loanzy-uk`), `build_status=deployed`, 17,497 chars rendered.

### Miss 1 — `tool-interest-rate-stress-test`: the `253` floor guard, on an unrelated slot

`save_sections` refused: `hero-tool 12→5 class attributes (42% kept, floor 50%)`. **Nothing was
written**, so the page is untouched (re-fetched: byte-identical to its 18,312 B / 0-input baseline).
The refusal is about `hero-tool`, not the calculator — a `page_rerender` regenerates every section,
so any slot the content writer flattens takes the whole save down. Retry filed 09:35Z (`84e586b9`),
because a fresh item is the only route: **the failed one is terminally parked at attempt 1 of 3.**
Read, then corroborated: `page-build-handler`'s `mark_item_failed` is
`update_work_item_status {"status":"failed"}`, and `ClaimWorkItemAction` takes only
`triaged`/`approved` — and the row's `handled_by` is **empty**, which is the discriminating tell,
since `FailWorkItemAction` (the arm that WOULD have returned it to `triaged`) writes `handled_by`.

### Miss 2 — `tool-loan-repayment-calculator`: I spent a full build on an ARCHIVED page

Component diverted, page build **complete on attempt 0**, four slots rendered including the new
calculator at 14,991 chars — and the page still 404s, because `pages.status='archived'`. The
archived-page guard refused the deploy stamp at the last step (two `ARCHIVED_PAGE_DEPLOY_REFUSED`
rows, 09:21:00–09:21:01Z), leaving `build_status='planned'` and `deployed_at` NULL. **The guard did
its job; my pre-flight did not** — I checked `build_status` and never `status`. Cost: one generation
plus one page build. It is also `bugs_open/266` demonstrated live: every producer up to the final
stamp did full work on an archived page. Pre-flight query added to the RUNBOOK. Unarchiving is the
loanzy lane's call, not a repair to take from here.

> **CORRECTION to §3's matrix in this same file, caught by the above.** §3 said
> `tool-compare-loans` and `tool-is-a-loan-right-for-me` "never planned a section at all", inferred
> from zero `page_components` rows plus a 404. **Both are wrong:** they carry **5 and 4 planned
> sections** in `pages.sections`, and the 404 is because they are **archived**. Loanzy has **four**
> archived tool pages (`tool-compare-loans`, `tool-eligibility-checker`,
> `tool-is-a-loan-right-for-me`, `tool-loan-repayment-calculator`). A 404 plus zero slots cannot
> distinguish "never planned" from "planned then archived"; `pages.status` can, and it is one
> column I did not read. **This also narrows `bugs_open/337`** (filed this session): of its three
> cap-failed pages, `tool-eligibility-checker` is archived, so the live loss is **two** pages on two
> sites, not three — corrected in that file before anyone acts on it.

### What this leaves on loanzy

3 tool pages now serve real calculators (plus car-finance from 08-19 = **4**). One retry in flight.
One blocked by archival (owner/lane call). `tool-credit-health-check` blocked by `337`.
`tool-loan-vs-savings` still needs only a re-render — not filed here, it is outside the owner's
chosen five and costs nothing whenever that lane wants it.

## 2026-08-20 14:05Z — the stress-test retry FAILED IDENTICALLY, which refutes my own prediction and settles the page as blocked

Retry `84e586b9` refused at the same step with **the same figures as the first run**:
`hero-tool 12→5 class attributes (42% kept, floor 50%)`, twice. I had written in the RUNBOOK that
"a floor refusal is content-writer variance and may well pass on a second run" — **wrong, and
wrong against a diagnostic I had added to `016b` §9 four hours earlier in this same session**
("N identical failures with IDENTICAL numbers is a deterministic refusal, not a flaky one"). I
wrote the check and then reasoned straight past it on the first case it applied to. Corrected in
the RUNBOOK; the cheap check is literally the one in my own entry — extract the numbers from both
errors and compare them before spending the second build.

**No damage from either attempt, confirmed at the artefact:** the page still serves 200 /
**18,312 bytes / 0 `<input>` / md5 `4374adb383d3270bdcfd184e42c361ef`** — byte-identical to the
08:17Z baseline. The guard writes nothing when it refuses, and that held twice.

**Why it is deterministic and what the remedy is.** The page's stored `hero-tool` carries 12
class-carrying elements; the content writer regenerates it with 5, every time, from the same
inputs. The guard's own guidance (`save_sections_component_floor.go:225-231`) names the fix and it
is not a knob: *"give it the component vocabulary in `content_direction` rather than lowering the
floor, which is what fixed the motivating page"*, with `section_component_floor=0` marked "the
deliberate escape hatch, not a fix". The floor is **step config** (default 0.5, slots under 10
class attributes out of scope), so there is no per-item override. **Not taken here:** editing
loanzy's `content_direction` changes how every page on that site is written, and that is the
loanzy lane's decision, not this lane's repair. **Stopped at two attempts — no third.**

### Final state of the owner's five

| page | outcome |
|---|---|
| `tool-loan-comparison-calculator` | ✅ serves 6 `<input>` (was 0) |
| `tool-overpayment-calculator` | ✅ serves 5 (was 0) |
| `tool-settlement-calculator` | ✅ serves 5 (was 0) |
| `tool-interest-rate-stress-test` | ⛔ blocked by the `253` floor guard on an unrelated slot — deterministic, page undamaged, remedy is a content-direction change on that site |
| `tool-loan-repayment-calculator` | ⛔ `pages.status='archived'` — component and page built fine, deploy stamp correctly refused; unarchiving is the loanzy lane's call |

**Component leg: 5 of 5.** **Page leg: 3 of 5, both misses attributable to other mechanisms working
correctly rather than to 311's fix.** Loanzy now serves four working tool calculators
(these three plus car-finance from 08-19), against one before this lane started.

## 2026-08-20 14:45Z — v1.0.1319, and CANDIDATE 3's mechanism finally pinned: the gate FIRES, and the page keeps serving the hole anyway

### Roll check first (the lane's own discipline: re-probe, never carry forward)

Fleet is on **v1.0.1319**, pods `agent-chassis-86b95b967b-*`, started **10:18Z** — so the afternoon
of this session's own work already ran on it, not on v1.0.1317. Probed at the binary: both
capability literals (`COMPONENT_COLLISION_DIVERTED`, `library tool claims this function`)
**PRESENT**; invented literal `zzzz_no_such_literal_311_control` **ABSENT**; two candidate stamp
shas from the build window (`759eea9d6`, `0cb95eb9d`) absent, i.e. the probe discriminates. Both
halves live on the current image.

### The stub hypothesis from §8 is NOT supported — measured, and it could have come out otherwise

§8 proposed that an unresolved section's **stub** keeps `pageSectionShortfall`'s count whole and so
hides a hole. Tested at the only place it could be true — rows with no component:
**11 of 1,855 `page_components` rows have `component_id IS NULL`, across 8 pages** [MEASURED
14:40Z]. They are not holes: two are tool pages on `lendzy.co.uk`
(`tool-complaint-deadline-calculator`, `tool-price-cap-checker`) carrying 13,262 and 14,747 chars,
and their served pages return **2 and 3 `<input>`** with working buttons. A componentless row is a
section rendered without a library component, not a stub standing in for a missing one. **So the
count is not being inflated, and the shortfall gate is not blind in the way I guessed.**

### What IS happening — read on the originating case, which is live right now

`remortgagecalculator.uk`, the site whose failure opened this bug ("left out the actual tools"):

- `pages.sections` for `index` = `["hero","mortgages-repayment","brief-explanation",
  "info-card-grid","mortgage-lender-directory","call-to-action"]` — **six** planned.
- `page_components` holds **five**, and the missing one is **`mortgages-repayment`** — literally
  the section named in this bug file's own step 1.
- `build_status = 'needs_rebuild'`, `deployed_at` cleared. **So the gate FIRED and was RIGHT**:
  5 < 6, refuse the stamp, flip to needs_rebuild.
- And the page **still serves**: `https://remortgagecalculator.uk/index.html` → 200,
  **40,726 bytes, 0 `<input>`** [MEASURED 14:42Z].

**The mechanism, stated precisely, and it is not "there is no gate":**
> **The refusal is a status write. It neither retracts the already-published artefact nor reaches a
> worker.** The previous deploy's file keeps serving — missing its calculator — while the row says
> `needs_rebuild`; and `needs_rebuild` has no consumer (established by this lane on 08-19: the
> `page-rebuild` agent has no scheduled task and zero orchestrations in history). So a page refused
> for a missing section serves the version without it **indefinitely**, and the DB looks correct
> the whole time.

That is a **convergence** defect, not a gate defect, and it is the surviving half of the owner's
original symptom. Nothing else in `bugs_open/` carries it: `210` is the inverse case (a stamp
wrongly APPLIED after a content failure); `208`, `219`, `220`, `226`, `333` are different triggers.

### Repair of the originating page: BLOCKED, and the pre-flight is why I know

The two-item recipe would fix that page — 311's fix means the store would now divert
`mortgages-repayment` to a `-remortgagecalculator-uk` row. **Not filed:** the site is **LOCKED**,
`locked_at 2026-08-18`, `locked_by = "portfolio_positioning: owner HALT 2026-08-18 pending
classifier register-input (RFC) + builder-flow decision"`. An owner halt is not mine to step
around. **Third time today the pre-flight earned its place** — and this time it cost nothing
instead of a wasted build. Incumbent `b89f91e1` re-pinned anyway for whoever does it:
html `a2c00f1c66ce6f4ef72b48083f1e3da6`, schema `8265ae5a931b735305b1fe007b148acb`, unchanged since
08-15. The `needs_new_component:mortgages-repayment` key is held only by `cancelled`/`failed` rows,
so a fresh item is insertable the moment the halt lifts.

**Next: `090`, not a bug file.** The claim above is structural (a refusal path that converges
nowhere) and CLAUDE.md's default applies — and this session has already been wrong twice today
asserting things about this exact code path. Filing it after a verdict, not before.

## 2026-08-20 15:30Z — `090` verdict: **UNVERIFIABLE (stopped: scope-not-narrowing)** — and it is worth more than a CONFIRMED would have been, because it named a discriminator I skipped

Run `e9555fad-5b25-46bc-9908-f40db98e16a4`, five evidence bundles over five iterations, item
`e0e0d7cb` complete. Verdict **UNVERIFIABLE**, *"Hand to a human with the full trail; do NOT
auto-conclude."* It cited `pageSectionShortfall`'s counting query and `UpdatePageStatusAction`'s
`} else if rendered < planned {` arm at Tier 0 — so the arm exists, as I read it — and then listed
three gaps. **Two of them land on claims I made in §"14:45Z" of this file an hour ago.**

> ### ⚠ CORRECTION to my own 14:45Z entry, within the hour
>
> **(a) "The gate FIRED and was RIGHT" on `remortgagecalculator.uk` is NOT established — it is an
> attribution I had no evidence for.** The loop's objection: *"the two guards write identical page
> rows, so the state alone is 'consistent with' the hypothesis, not direct confirmation of which
> guard fired"* — it looked for `agent_error_log` rows to attribute three shortfall pages and got
> **zero**. `refuseDeployStampOnSkip` (`page_build_failure_guard.go`), the shortfall arm,
> `check_unresolved_sections`, and `flagPagesForRebuild` (`maintenance_actions.go`) all leave the
> same `build_status='needs_rebuild'` row. I inferred the shortfall arm because planned(6) >
> rendered(5) — suggestive, and not proof. **`needs_rebuild` + cleared stamp is a row with at least
> four possible authors and no attribution column.**
>
> **(b) "`needs_rebuild` has no consumer" is REFUTED by a case that was in my own output this
> morning.** The loop noticed repeated `page_rerender` completions on `tool-ab-test-calculator`
> while `build_status` stayed `needs_rebuild`. Checked [15:28Z]: that page's rerender item
> `ad2a2dc4` is **`complete`**, handler **`page-rerender`** (not `page-build-handler`), and the page
> **does serve the new forked calculator** (5 `<input>`, graded at 08:15Z) — while
> `build_status='needs_rebuild'` and `deployed_at` is still **2026-08-14 22:10:57**. So the page WAS
> reprocessed and republished; the flag simply was never cleared. Contrast today's five loanzy
> repairs, filed at `page-build-handler`, which all flipped to `deployed` with a fresh `deployed_at`.
> **Two rebuild paths, only one of which maintains the status columns.** I had this page's
> `needs_rebuild`-while-serving in front of me at 09:05Z and read straight past it.

**What still stands, because it is an artefact fact and not an inference:**
`remortgagecalculator.uk/index.html` serves 200 / 40,726 bytes / **0 `<input>`**; its planned
`mortgages-repayment` section has no `page_components` row; and 311's fix now makes that section
creatable. Everything about *why* the page is in that state is unestablished.

**Consequences, and they are the useful output of the run:**
1. **Candidate 3 stays OPEN and UNFILED**, with a sharper next step than the one I wrote at 14:45Z:
   the missing piece is **attribution**, not observation. Either the guards must log distinguishably
   (`refuseDeployStampOnSkip` already writes an `agent_error_log` row; the shortfall arm appears not
   to), or a filer must catch one in the act. The loop's own "still needed" list also asks for
   `sites.deploy_config` / `published_hash` / `published_at`, since nothing in the bundle can even
   see the deploy bucket — **the "does the refusal retract the published file" half is not
   answerable from the Go code at all.**
2. **The status-column half belongs to `bugs_open/315`'s family, not to a new bug.** 315 is
   "`deployed_at` stamped WITHOUT publishing"; this is the inverse — **published without the stamp,
   and the flag left set.** Same defect class (the `pages` status columns do not track the
   artefact in either direction), different direction. Contributed there rather than filed here.
3. **A `090` that comes back UNVERIFIABLE is not a wasted run.** It cost one dispatch and it killed
   two of my claims — one of which I had already committed to a bug file. That is the cheapest
   place for those to die.
