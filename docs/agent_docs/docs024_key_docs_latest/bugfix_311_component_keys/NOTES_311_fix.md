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
