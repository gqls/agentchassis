# NOTES — bugfix 184 (literal markdown reaches the page) — append-only, newest at the bottom

## 2026-08-18 ~19:35Z — lane opened; ownership check; bug re-validated

- **Ownership**: `who-owns.py 184` says "OWNED or recently active", but read the detail:
  every recent commit is a *contribution* from another lane (`bugfix_201` 08-07 — "No action
  requested of this lane"; `bugfix_277` 08-17 — "This lane is not claiming the bug"). Grepped
  live session transcripts (`~/.claude/projects/.../\*.jsonl`) for `literal_markdown`: the one
  session with real hits (15) is the `bugfix_277` promoter-floor lane (migrations 466/471),
  touching the pair as *data*, not working 184. **Taking the bug.** [VERIFIED — transcript grep,
  2026-08-18 19:30]
- **Bug still valid** [MEASURED, live DB 2026-08-18]:
  - `site_work_items` `item_type='literal_markdown'`: 34 unresolved / 24 failed / 10 detected
    (newest **2026-08-18**, parked by the migration-444 promoter floor) / 3 complete /
    3 needs_human_review. 71 open across 6 sites; webdesign.co.uk carries 46, robot-hands.com 16.
  - Open items' findings by pattern: heading 107 (content_data) + 106 (rendered_html),
    code_span 49+8, bold 8+8. Samples include `## [` — headings that *contain markdown links*.
  - Fleet content_data quick scan (my own regexes, cruder than the check's): bold 1,
    heading 5, **md_link 9**, code_span 1 of 1,723 components. **Markdown links are now the
    largest raw bucket and the detector has NO link pattern** (check_literal_markdown.go:107-114
    covers bold/code_span/heading only). Confirms the 08-17 contribution's "symptom has widened"
    warning at the code level.
- Migration 304's prompt rule IS live (`rule9_extended = t` on `page-content-writer`), and the
  writer still emits markdown (bugfix_201's 08-07 artefact evidence: 18 findings re-written into
  the very field it was dispatched to clean). A prompt is not a control — again.
- Repair pair `literal_markdown → page-build-handler`: 1 complete / 28 failed lifetime (worst in
  fleet), HELD by the migration-444 success floor. New findings park at `detected`.
- Dispatched two Explore agents: (1) component write path / repair seam / field-type schema,
  (2) writer generation path + bugfix_201 lane conclusions + CQ-019 register state.

## 2026-08-18 ~19:45Z — artefact confirmation + status semantics

- **Served defect confirmed at the artefact** [MEASURED, curl 2026-08-18 19:42Z]:
  `fundamentallyai.com/news/index.html` visible text carries **11 ATX headings and 2 raw
  markdown links** right now. Not a stale-DB claim.
- `status='unresolved'` semantics: the two-strike rule (complete_work_item_verification.go /
  create_tool_cross_link_items.go:88 comment) parks a RE-detection as a non-dispatchable
  zombie once the page has ≥2 terminal failures; a stale-triaged 48h reaper also writes it.
  So the 34 unresolved rows are the same defect re-found on pages whose repairs already
  failed twice — consistent with "no working repairer exists".
- Diagnosis queue checked: no open `needs_diagnosis` item on this mechanism. No duplicate filing.

## 2026-08-18 ~20:20Z — research done, design fixed, part 1 committed, council submitted

- Two Explore agents mapped the machinery. Load-bearing facts (each verified in code by me
  before use): RenderComponentAction's :1988-:2012 window has the LLM map + schema in hand
  and feeds BOTH surfaces (v3_site_actions.go:2007 "capture before merge"); the no-LLM
  rerender does NOT pass RenderComponentAction (renders via RenderTemplate from stored
  content_data :486-491, persists mergedContent :543-564) — so it needs its own hook, and
  that hook is what makes a plain rerender the repair; `page-rerender` is in
  knownHandlerAgents (:110) with a 5,044-complete lifetime as an item handler; the live
  check_rerender_mode condition already grew to four reasons (read from the live row, not
  the seed).
- **Design**: strip-only shared primitive in datahelpers (scan + strip, property test
  scan(strip(x))==∅), md_link widening, HandlerAgent → page-rerender with spec.reason,
  opt-in default-OFF strip flags at three seams, migrations 473/474. Full reasoning:
  PLAN_2026-08-18_mechanical_markdown_repair.md. CQ-019's deferral of normalise-on-write
  answered (strip-only ≠ markdown→HTML; not in the save action; opt-in).
- **Council**: submitted, corr `060bcc0a-1ba5-4525-8fea-03de021e26f5` (~30 min budget).
- **SAME-FILE PASSENGER incident, handled**: rerender_page_sections_action.go carries the
  299 lane's uncommitted KEEP #3 hunk whose helpers (links_tel.go) are NOT at HEAD. Their
  session messaged me proposing "either order works" — WRONG in one detail: my strip hunk
  calls datahelpers.StripLiteralMarkdownFromContentData, not at HEAD at that moment, so
  their commit-first would have broken HEAD's build via MY passenger hunk. Resolution
  agreed by direct message: I committed the datahelpers primitive first (`019fb0616`,
  gofmt follow-up `5fbe549f7`) so my hunk compiles as their passenger; they commit next
  (naming my block as a passenger); I land part 2 (check re-route + rerender hook) after.
  **The check re-route is deliberately NOT at HEAD yet** — re-route without the strip hook
  in one image burns literal_markdown attempts on an unequipped handler.
- Migrations 473 (+ROLLBACK) and 474 (+ROLLBACK) authored; anchors verified against LIVE
  rows (page-content-writer step `render_section` action `render_component`;
  section-editor step `apply_edit` action `apply_section_edit`). Both safe pre-image
  (flags unread, reason unemitted); intended post-image.
- Register CQ-019 updated (status was stale two weeks — said "inert until 303" while
  303/304 were live since 08-04); bugs_open/184 progress note appended; 016b §9 entry
  appended ("repair-by-regeneration cannot fix a defect the regenerator has the habit of
  producing").

## 2026-08-18 ~21:00Z — 299 commit landed; part 2 at HEAD; council round 1 = REVISE, answered

- 299 lane committed `757a0890a` with my strip block as a named passenger, exactly as
  agreed. Part 2 committed `763bb5d55` (md_link widening + re-route). IMAGE_TAG bumped
  `f9cf30974` (v1.0.1310 → v1.0.1311 — all three of makefile/overlay/live pod read 1310,
  the same-tag cached-image precondition).
- **Council round 1: REVISE**, gating objection = my own risks-§6 config-path doubt,
  flagged as HIGH by four seats independently. The committed code was already uniform
  (`params.StepConfig.Config` at all three seams — v3:1836, rerender:129+hook,
  section-editor switch); the SKETCH was the false half → WRONG_CALLS entry appended.
- Real improvements taken from the round, not defended:
  - migrations 473/474 reworked: `snapshot_agent()` instead of bespoke backup tables
    (reuse seat — and the rework fixed a flaw the seat didn't name: my restore-from-
    backup rollback would have clobbered intervening migrations; the surgical inverse
    doesn't), needle gate added to 474 (debug_historian).
  - true lifetime counts re-measured across live + archive (prior_art_librarian):
    rerender pair 13,993/142 (stronger than cited), build-handler pair 3/36.
  - RUNBOOK: binary probe for the strip gates before batch promotion; chrome/head
    scope note (items structurally name body sections only).
- bug_historian's HIGH (strip at call sites, not inside RenderTemplate itself) is
  answered, not adopted: a strip inside the shared render mechanism would mutate EVERY
  template render including non-LLM contexts — the blind-mutation shape the estate
  rejects; coverage of the REPAIR is universal anyway (any writer's markdown is caught
  by the fleet-wide check and healed by the rerender path, whatever produced it), so an
  unflagged caller is detected-and-repaired, not silent.
- architecture seat's ConfigKeys point: rerender spec HAS the declaration; creating a
  first-ever ActionInputSpec for render_component (the fleet's most-shared render
  action, no spec today) as a rider on this fix is a bigger blast radius than the fix —
  named as follow-up, not smuggled in. apply_section_edit likewise has no registered
  spec (its step config predates the spec system).

## 2026-08-18 ~22:00Z — round 2 REVISE: two HIGHs UPHELD (real defects), one refuted; fixed and resubmitted

Round 2 found genuine design defects — the REVISE-is-cheaper-than-the-defect rule proven again:

- **UPHELD (guardian, HIGH)**: my step-level flag made the strip fire on EVERY
  sections-branch rerender — image_landed, template_changed, cta_links_stale — a blanket
  change to the fleet's highest-volume pipeline framed as a scoped repair. FIXED
  (`6fa9f5673`): double gate, flag AND spec.reason=='literal_markdown'. I had privately
  considered blanket-hygiene "intended"; the council was right that the plan didn't say
  so and the blast radius didn't match the framing.
- **UPHELD (bug_historian, HIGH)**: Info logs are ~minutes of retention — not an audit
  trail for a content-mutating transform. FIXED: stripped field paths now ride every
  seam's action RESULT (`stripped_markdown_fields`, durable via collected_data), and
  the repair path's pre-strip state is archived (save_page_sections →
  page_component_history:753, first-handed).
- **UPHELD (prior_art, HIGH)**: `ApplySectionEditInputSpec` EXISTS — my "no registered
  spec" was a FALSE ABSENCE, second same-class miss of the session (grep before
  asserting an absence in a submission). WRONG_CALLS entry #2 filed. ConfigKeys added.
- **REFUTED by measurement (editquality, HIGH)**: the substeps-vs-sub_workflow landmine
  needs BOTH keys present; the live loop config carries ONLY sub_workflow
  (jsonb_object_keys measured), and loop_actions.go:87-104 falls back to
  sub_workflow.steps — 474's path is the executed path.
- Also measured: all three target agent types carry exactly ONE active row (the
  dormant-duplicate trap has no substrate); no existing consumer of the new
  `stripped_markdown_fields` key (3 hits = my 3 writers).

Round 3 submitted on the same correlation (RUN_ORCH_ID 381fc44e). All code committed:
`6fa9f5673` is the round-2 fix commit.

## 2026-08-18 ~23:15Z — round 3 REVISE answered (9c77e0dff), round 4 APPROVED

- Round 3's asks, all landed in `9c77e0dff`: RenderComponentInputSpec (the action's
  FIRST spec — ten read keys as ConfigKeys; `output_html` deliberately undeclared: live
  configs carry it, ZERO Go readers — leaving it undeclared lets the coverage report
  surface a dead key instead of hiding it); `shouldStripLiteralMarkdown` extracted with
  a direct containment test (every other reason + near-misses + mistyped flag values
  must not strip); NULL-direction analysis written into both migration headers (all
  verifies are positive-presence; absence RAISES; no `<>`-on-NULL green possible).
- The gating bug_historian objection was answered with the census it asked for: LLM
  content_data reaches page_components through exactly FOUR writers — three hooked,
  one (create_report_page) disclosed as the detector-covered residual with zero live
  findings (measured: no open item's page carries a report/dossier slot).
- **Round 4: APPROVED** ("approved with 2 advisory objection(s) — none high-severity").
  Advisories, dispositions:
  - debug_historian (medium): do not apply 473/474 before pod-verifying the binary —
    already RUNBOOK rollout step 2 (provenance + per-replica binary probe).
  - render_guardian (medium): a section whose render bails out can persist stripped
    content_data while carrying stale dirty rendered_html — the verifier scans both
    stored surfaces, so completion is still refused; honest, not false-green. Noted.
  - architecture (low): RenderComponentInputSpec sits at 10 ConfigKeys — at the N=10
    boundary; the audit (`audit-optional-key-budget.sh --json`) counts OPTIONAL keys
    and reports nothing over budget (apply_section_edit 6/10, rerender 3/10);
    `cmd/config-key-audit` parity test green.
  - prior_art (low, twice): the two absence claims ("first spec", "output_html unread")
    marked as content-search absences a human may re-confirm — both were fresh greps
    this session, stated as such.
- Trail complete: REVISE → REVISE → REVISE → **APPROVED**, one correlation
  (`060bcc0a`), every commit already carrying `Council-Submitted:` — 098 credits them
  automatically now the verdict is approved. Two of the four rounds caught REAL design
  defects (blast radius; audit trail) — the revise-is-cheaper rule, proven again.
- SUMMARY_2026-08-18_literal_markdown.md written (milestone: design+code+council done;
  rollout pending the owner's fleet release).

## 2026-08-19 — image live, migrations applied, canaries run: the seam WORKS and found the next layer down

- **Deploy verified at the binary**: v1.0.1314 both replicas — `strip_literal_markdown`,
  `shouldStripLiteralMarkdown`, `md_link` all present, nonsense control absent. (The
  provenance log line had scrolled; the binary probe has no shelf life.)
- **473 + 474 APPLIED 10:34:26Z / 10:34:35Z**, ledger-recorded via `--record-only`
  (the runner's pending list carries a dozen other threads' files — `--apply` untouchable).
  473 first got the 083 lane's CONTRIB folded in: SCOPE header (owned pages = expected
  residual, bugs_open/301's) + at-apply generic/owned split NOTICE. Open population at
  promotion time: **30 generic / 41 owned**.
- **Canaries (2 generic + 1 owned, per the CONTRIB's ask). The mechanism worked end to
  end and the verifier caught a REAL deeper layer:**
  - dartsonline news-index: strip FIRED (`stripped_markdown_fields:
    [news-listing:items[18].summary]`), save succeeded (3 sections) — and the verifier
    refused: the SAME field dirty again. Post-save DB read shows resolver output
    (`## Updated PDC…` + a raw markdown TABLE). **Query-resolved fields cannot converge
    by stripping the stored copy: `plan.ResolvedData` merges LAST and wins.**
  - Root measured: **`content_feed_items.source_summary` carries raw markdown in ~700
    of 10,855 rows** (699 headings, 272 links, 81 bold, 8 whole scraped tables). The
    resolver (news_items.go:95) feeds it verbatim on every render.
  - Owned canary (webdesign learn-index): verifier-refused on a `ported-page` slot's
    rendered_html code_span — it did NOT hit the ownership guard (no save attempt on the
    residue path); honest failure either way, exactly the 473 SCOPE prediction's shape.
    Ported/tool-slot findings are structurally out of this route's reach (defect lives in
    ported HTML, not content_data) — the tool-rebuild programme's remit.
- **Fix committed `f3939f27d`** (inert until next roll): gated strip extended to
  `plan.ResolvedData` (rerender) and the `merge_with` overlay (v3), both under the same
  double gate; PLUS an **unconditional producer-local strip in `projectNewsItems`**
  (strip-before-truncate — truncating mid-`[text](url)` leaves an unmatchable
  half-pattern; ingest record stays raw, render value cleaned, the directory_items
  EscapeString posture). Verified against git-archive HEAD + my 3 files — the working
  tree carries another session's WIP referencing undefined `refuseOwnedPageIfConfigured`.
- **Process miss, owned**: `f3939f27d` carries `Council-Reviewed:` but its content was
  NOT in the approved r1–r4 plan — premature trailer. Correction: **round 5 submitted
  the same hour** on the same correlation (RUN_ORCH a460f0f6) covering exactly these
  three edits + the honesty disclosure. WRONG_CALLS entry appended. A REVISE verdict
  will be acted on forward.
- **Markdown TABLES are outside the 184 pattern set** — 8 source rows serve pipe-text
  even after the strip; recorded as a feed-quality follow-up, not silently absorbed.
- Both canary generic items are terminally `failed` (3 attempts each) — **re-arm them
  after the next roll**, along with the remaining generic population.
- Cross-session: 083/277 lane's two messages folded in (their by-construction owned-page
  analysis was right; their canary ask ran); 8d lane confirmed 473/474 independently and
  contributed the shallow-jsonb query gotcha (now in RUNBOOK). All replies sent.

## 2026-08-19 ~13:00Z — CHECKPOINT (session ending): round 5 = REVISE (real objections, unanswered yet); new pair is promoter-HELD until first completion

- **Round 5 REVISE, gating = editquality**: (1) HIGH — the v3 merge_with strip is
  single-gated (flag only) while the rationale claimed "same double gate"; genuine
  sketch-vs-claim mismatch to resolve — either add a reason-equivalent gate at that seam
  or state plainly why the writer path is flag-only by design (there is no spec.reason
  at generation time — likely the honest answer, but it must be SAID, and guardian's
  no-kill-switch HIGH on the unconditional projectNewsItems strip needs weighing too:
  options are (a) defend producer-local with the directory_items precedent VERIFIED in
  code (reuse seat says the precedent claim is asserted, not checked — it escapes HTML,
  it does NOT strip markdown, so the precedent is posture-similar not behaviour-identical
  — concede precisely), or (b) put the news strip behind a config/env kill switch.
  (2) bug_historian MEDIUM: feed_triage + render_rss_feed still consume the dirty
  source — enumerate/disposition, don't leave implied. (3) editquality MEDIUM: check
  whether newsTopicalTokens/dedup reads pre- or post-strip titles (my strip happens in
  projectNewsItems AFTER capNewsItemsPerTool/topical — verify order in code). (4) MEDIUM:
  plan.ResolvedData aliasing — planSection returns a fresh local map per section
  (b6e374fc2 doc_notes correction says exactly this) — cite it. (5) render_guardian LOW:
  confirm strip-to-empty can't slip past the escalation check (letter-guards keep ≥1 char
  of text — but verify the bare-image-markdown case). **NONE of these are applied-config
  risks — 473/474 stay correctly applied; the REVISE concerns f3939f27d's code, which is
  INERT until the next roll, so there is time to fix before it ships.**
- **f3939f27d's premature Council-Reviewed trailer now definitely mismatches** (r5 came
  back REVISE) — the WRONG_CALLS entry stands; fix forward in the round-6 commit.
- **083 lane's post-apply notice (load-bearing for rollout)**: the NEW pair
  `(literal_markdown, page-rerender)` is **held by the promoter's known_good gate**
  (0 lifetime completes; 0/3 post-apply). In ~3 days `held-pair-canary-escalation` will
  escalate it to needs_human_review — that is THEIR mechanism working, not our fix
  failing. **The unblock = one successful hand-canary after the next roll** — already
  RUNBOOK step; now it also releases the promoter queue. General landmine (theirs):
  re-routing an item_type mints a history-less pair the promoter holds by default.
- Their canary-model updates recorded: an owned-page item can terminally fail WITHOUT
  touching the ownership guard (verifier blocks first on ported-slot rendered_html);
  and no verifier-registered type can false-green via "completed having written nothing".

## 2026-08-19 ~17:00Z — FIRST ARTEFACT-VERIFIED REPAIR; 083 notice RETRACTED; r5 code shipped by another session's roll

- **083 lane RETRACTED its held-pair claim** (measured by them): 473 touches no work
  items; their 0/3 rows pre-date the apply; the repair lands on `(page_rerender,
  page-rerender)` 14,454/142 — never held. **Delete the "promoter-held, escalates in 3
  days" bullet from the previous checkpoint; nothing blocks rollout.** The general
  property (re-pointing a handler mints a history-less pair) survives only as derived.
- **v1.0.1315 rolled at 12:15Z (another session's release) and CARRIES `f3939f27d`** —
  binary-probed both replicas for the resolved-strip literal, control clean. Review is
  after the fact by design (owner ruling 07-29); the r5 REVISE fixes become a follow-up
  roll, not a pre-ship gate. Stated plainly: code under REVISE is live.
- **Re-armed the fundamentallyai canary → `complete` at attempt 0** (first-ever for
  this item), verifier record: *"no literal markdown on either surface across 2
  component(s)"*. **Served page: 13 defects this morning → 0 now, 12,092 visible
  chars** (not emptied). **The mechanism is proven end-to-end at the artefact.**
- Next: answer r5 in code (dispositions in the previous section), promote the remaining
  generic items (~28: statuses detected/failed/unresolved on non-owned pages), verify a
  sample at the artefact, let retraction close leftovers. Owned/ported 41 → 301 / tool
  rebuilds. Bug closes when the founding + widened-symptom pages verify clean served.

## 2026-08-19 ~17:20Z — BATCH PROMOTION: 7 generic pages, 6 complete at attempt 0; artefact check found two NON-markdown anomalies

- 083 lane confirmed (against the promoter's predicate, not assumed) that the
  fundamentallyai completion opened the `known_good` door: new findings at page-rerender
  now promote automatically. ⚠ Their arithmetic worth keeping: the new pair is 1/2 with
  2 PRE-FIX failures banked forever; `floor_ok` binds at the 5th outcome — two more
  failures before another success would HOLD it. That is why the batch was promoted
  selectively (content_data-side findings, no ported slots, one row per page) rather than
  indiscriminately.
- Promoted 7 pages (6 re-armed + webdesign news `detected` flipped in place; the
  `idx_swi_dedup` index allows ONE indexed row per page — duplicates and pages with a
  `needs_human_review` sibling were left alone). **6 complete at attempt 0 within ~7
  min**, each verifier-certified; the 7th queued behind other sites' work.
- **Served-page verification of the six** (the only check that counts): fundamentallyai
  news 0, dartsonline news 0 (12,346 chars — tables' pipes not counted, disclosed),
  robot-hands news/index 0, webdesign news 0, robot-hands gripper-catalog 0 — **five pages
  clean with text intact.** Two anomalies, both OUTSIDE the markdown mechanism:
  - **robot-hands.com/news.html still SERVES 5 headings/links although its 3 stored
    components are clean** (rewritten 16:08Z today, verifier-certified): `pages.deployed_at`
    is still **2026-08-11** — the rerender's deploy step did not republish this page. The
    DB-is-not-the-website landmine, live: the repair is done and the publish is missing.
    Not a markdown defect; a rerender/deploy pipeline gap — filed as a follow-up for the
    rerender lane rather than re-armed here (re-arming would pass the verifier again and
    still not deploy).
  - **robot-hands.com/learning-center/index.html is a 404** (339 visible chars of
    not-found chrome). A page row + stored component exist for a URL nothing serves —
    a dead-page/deploy-path issue, not markdown. Noted for the robot-hands lane.
- Net for the bug: the repair converges for every page the route can reach AND deploy.
  Remaining open: ai-agent-orchestration news (behind a needs_human_review sibling — a
  human decision, left alone), the 24→sibling duplicates (retraction will close them on
  the next discovery pass now the pages are clean), and the 41 owned/ported (301).

## 2026-08-19 ~21:00Z — ROUND 6 written, committed (`f6d632291`), submitted; one self-inflicted duplicate submission

- **Context on pickup**: the bug was CLOSED at 17:25Z (`0ca143c2d`, owner note `a3497ddfc`);
  the one thing owed on the lane was the round-6 answer to the r5 REVISE on `f3939f27d`
  (live since v1.0.1315). Read the r5 seat payloads from `orchestration_states
  a460f0f6…` (11 reviewers, 6 abstained, gating = editquality) rather than my own
  summary — two seats (bug_historian ×2 objections, debug_historian ×2) had detail the
  checkpoint bullet had compressed.
- **What round 6 changes vs. states.** ONE behaviour change: `projectNewsItems` now reads
  `DISABLE_NEWS_MARKDOWN_STRIP` (ships ARMED; set = HTML-escaped raw text, the exact
  pre-strip output) and logs `items_stripped/items` when anything changed; signature
  gains the logger. Everything else is corrected comments and tests. Precedent REUSED,
  not minted: `DISABLE_UNREGISTERED_HANDLER_DEMOTION` in `load_work_item_actions.go` is
  a council-round-1 guardian ask with the owner's "no default-OFF switches" stance
  already reconciled in its comment — the guardian r5 text ("default-on is fine, but
  make it a flag") asks for exactly that shape.
- **Measurements that answered seats** (all live, 2026-08-19 ~20:50Z):
  - `render_component` with `merge_with`, every step any depth (`jsonb_path_query
    '$.**'`): exactly two live steps, both page-content-writer's — `render_section`
    (flag ON, 474) and **`render_from_template` (flag UNSET)**. The second one is the
    unflagged consumer the producer-side strip exists for — a fact I had asserted in
    f3939f27d's comment ("unflagged callers") without naming one.
  - `rerender_page_sections`: ONE live step (page-rerender). `reason='literal_markdown'`:
    ONE Go writer (`check_literal_markdown.go:376`) + operator hand-promotion.
  - dedup/cap order: `newsTopicalTokens`/`capNewsItemsPerTool` run in `QueryNewsItems`
    (:143-144) on RAW titles, before projection. Strip changes no clustering.
  - `directory_items.go` precedent: escapes only, no strip — reuse_agent was right that
    I asserted it; conceded precisely in the comment (posture-similar, behaviour-different).
  - strip-to-empty: every pattern keeps ≥1 letter/digit; `![alt](url)` → `!alt`; the
    stored strip runs BEFORE the pre-check whose predicate is `isEmptyContentValue`
    (`json_envelope.go:468`) — an emptied required field ESCALATES. Pinned by
    `TestStripToEmptyOnlyFromAlreadyEmptyInput`.
  - aliasing: `planSection` allocates `resolvedData := make(...)` per call
    (`plan_sections_action.go:2077`); readers after the strip are the rc merge and
    mergedContent in the same iteration; nested values may alias `sourceResolver.specs`
    / `storedContent` (per-invocation); Strip is a fixpoint.
- **Mutation check on the kill-switch test**: predicate forced to `true || …` →
  `TestProjectNewsItemsKillSwitchRestoresRawText` FAILS ("got 'Read the full
  report.'"); restored → ok. Build/vet/test in the WORKING TREE (no scratchpad archive):
  `go build ./platform/orchestration/...` 0; vet clean on the edited packages (one
  pre-existing vet warning in `load_component_library_actions.go:207`, not mine); `go
  test ./platform/orchestration/actions/... ./platform/orchestration/datahelpers/` all ok.
- **My first test fixture was wrong, not the code**: `TestProjectNewsItemsStripsBefore
  Truncating` asserted the link TEXT survived a 200-char cut while the fixture put the
  text itself astride the cut (190 + 14 chars). The load-bearing assertion (no `](`
  half-pattern) held throughout; shortened the prose to 180 so the stripped text sits
  before the boundary and the raw link would have straddled it.
- **⚠ SAME-FILE PASSENGER, outbound this time**: my `v3_site_actions.go` comment hunk
  (the "FLAG-ONLY HERE, BY DESIGN" block) was swept into the 315 lane's commit
  `f0dd97c71` (21:42:37 +0100) before I committed — `git show f0dd97c71 -- …v3_site_actions.go
  | grep -c 'FLAG-ONLY HERE'` → 1. Forward-only; named in my commit message and in the
  submission (edit 3, "already at HEAD"). Nothing lost; the trail just reads odd.
- **⚠ WRONG CALL — I DOUBLE-SUBMITTED ROUND 6.** The trigger ran once and printed its
  tail; wanting the HEAD of the output I re-invoked the script with `| head -40` instead
  of reading the output file — it published again before I killed it at 5 s. Two
  council-gate rows, identical payload md5: `60e14994…` (20:45:52Z) and `cf86d0db…`
  (20:46:20Z). No coordinator-level cancel exists that a mid-flight council honours and
  hand-editing a live orchestration row is how stuck rows are made, so BOTH run; cost =
  one duplicate round of seat credits. Either verdict is valid for the round (same
  payload); I'll record both. Logged in WRONG_CALLS. **Cheap check: a publishing script's
  output is in the task's output file / scrollback — READ it, never re-run the script
  to see it again.**
- Commit `f6d632291` carries `Council-Submitted: 060bcc0a-…` (not Reviewed — the r5
  premature-trailer lesson).
- Next: read the r6 verdict(s) (budget ~30 min; find by payload); act on REVISE if any;
  after the next roll, probe EVERY agent-chassis pod for the kill-switch literal with two
  controls (RUNBOOK); the kill switch itself is exercised only by the unit test until an
  operator ever needs it — stated, not hidden.

## 2026-08-19 ~21:20Z — ROUND 6 APPROVED (run `60e14994…`, 14 reviewers, 3 abstained, 4 advisory objections, none high); advisories acted on

- Decision text: "approved with 4 advisory objection(s) — none high-severity". Objecting
  seats: editquality (2), guardian (3), bug_historian (1), prior_art (4). Approving:
  architecture, constitution, debug_historian, diagnosis_guardian, guidelines,
  improvement_guardian, mission, render_guardian, reuse_agent, tooling_provenance.
- **Acted on:**
  - bug_historian LOW "named, not tracked" → **`bugs_open/332` filed** for the RSS sibling.
    Measured before filing, and the measurement changed the framing: the 574/2,382
    fleet-wide feed rows with headings are ALL on sites that publish no RSS; exactly one
    site has `rss_feed` enabled (relojistas.com) and its live feed serves 0 markdown
    descriptions of 25 (its own rows: 0/79). So 332 is LATENT with a named re-review
    trigger, not a live bug. Pointer added to `bugs_closed/184`'s residuals.
  - prior_art MEDIUM "no reproducible SQL" → the exhaustiveness queries are now verbatim in
    the RUNBOOK (deep `jsonb_path_query '$.**'` step enumeration; the grep for reason
    writers; the grep for projectNewsItems callers; the kill-switch precedent by name).
  - guardian LOW "confirm the two callers" → confirmed by grep: `news_items.go:334` and
    `:351`, nothing else outside tests.
- **Noted, no action:** editquality/guardian on edit 3 being a stale hunk already at HEAD —
  correct; it was listed for the record with that said; a no-op, not a conflict. prior_art
  on the f0dd97c71 git-state claim being outside its check tier — true of every git claim;
  the verification path is `git show f0dd97c71 -- …v3_site_actions.go | grep -c 'FLAG-ONLY HERE'` → 1.
- The duplicate run `cf86d0db…` is still executing; its verdict will be recorded here when
  it lands, whatever it says — the correlation is already approved by the first.
- `f6d632291` carries `Council-Submitted:`; 098 credits it at report time. NOT amending.
- SUMMARY_2026-08-19 written (the close + the approved follow-up is the inflection).

## 2026-08-19 ~21:35Z — duplicate run `cf86d0db…` also APPROVED (14 reviewers, 3 abstained, 3 advisory, none high); one advisory taken into the test

- Same payload, independent panel draw, same outcome — a small, unplanned measurement of
  the gate's consistency on this submission (n=2; says nothing about the gate in general).
- New and useful: bug_historian MEDIUM — the strip-to-empty property test checked `== ""`
  but not whitespace-only output. The property is the same (every pattern keeps alnum; the
  heading strip removes only its prefix, so the line's text stays), so the assertion is now
  `strings.TrimSpace(got) == ""` with five more inputs (`"#  x"`, `"# \n# y"`, `"**a** "`,
  `" `b` "`, `"#   "`) — passes. Committed below.
- debug_historian MEDIUM — "2 of ~41 pods": MEASURED: the chassis image runs in **93 pods**
  (94 incl. one more at the instant of counting), `-l app=agent-chassis` = 2. All on ONE
  tag (v1.0.1316), which is what makes a two-pod probe valid; the RUNBOOK step now asserts
  the tag count first and probes per tag. Probed v1.0.1316: r6 literal ABSENT, r5 literal
  PRESENT, both controls correct — 1316 is the 315 lane's build, cut before `f6d632291`.
  So round 6 is still inert, as recorded.
- editquality LOW ×2: `operation` said `add` for two test edits that modify existing files
  — fair; cosmetic, noted for the next submission.
- The remaining advisories repeat run 1's (stale edit-3 hunk, prose-not-SQL, precedent
  not code-verified by the seat) — all answered in the previous section and the RUNBOOK.

## 2026-08-20 ~07:15Z — v1.0.1317 CARRIES ROUND 6 (probed, controls ok, switch armed); discovery cadence measured: the retraction premise was resting on a hand-dispatch

- **Roll verification (RUNBOOK step, tag-based):** fleet is ONE tag, v1.0.1317 (9
  chassis-image pods this morning — the count swings with spawned agents; the TAG count
  is the assertion). Probe on a label pod: `DISABLE_NEWS_MARKDOWN_STRIP` PRESENT, the r6
  log literal PRESENT, ctrl+ (DISABLE_UNREGISTERED_HANDLER_DEMOTION) ok, ctrl−
  (nonsense) ok. `printenv DISABLE_NEWS_MARKDOWN_STRIP` exits 1 on the pod → env unset →
  **strip ARMED**. `f6d632291` IS LIVE. Gotcha recorded in the RUNBOOK: my commit's own
  sha is legitimately ABSENT from the binary — the provenance stamp carries only the
  build's HEAD sha, so the literal probe (with controls) is the decisive check, not a
  sha grep.
- **Strip-log demand control:** 0 strip lines since the roll, and 0 news resolutions of
  any kind across ALL chassis-image pods — the zero is demand-starved, not evidence.
  Induced one standard light rerender (049b, reason=section_data_resolved) on
  fundamentallyai.com news-index after simulating the resolver's own selection:
  **12 of its top-20 news_archive rows carry markdown right now**, so the run MUST strip
  and log if the mechanism works. corr `343edda2`. Pre-flighted: no NULL content_data,
  no open items on the page.
- **⚠ FINDING (cadence, not detection): quality-discovery-agent has run exactly TWICE in
  the month orchestration_states retains (min created_at 07-19) — both 2026-08-19
  10:26Z, both hand-dispatched** (initial_request_data is a bare cli-style orchestrate).
  So the closed bug's "retraction closes the duplicates on the next discovery pass" was
  really "on the next HAND-dispatched pass", which nobody owed. The
  detection-works/schedule-doesn't pattern, on this check's drain path. Confirmed the
  retraction REACHES the stale rows before dispatching: `work_item_retraction.go:103` —
  status predicate is workItemClosedStatuses, so `failed`/`unresolved`/`detected` all
  close; the gate is scannedPages (only pages the run positively re-scanned and found
  clean on both STORED surfaces).
- Dispatched scoped discovery for the three sites with drainable rows (same envelope as
  the 08-19 runs): dartsonline.com corr `5a9d4142` (3 unresolved, page verified clean
  yesterday and re-verified stored+served this morning: 0 defects, 8,402 visible chars),
  robot-hands.com corr `759ccae4` (9 failed), webdesign.co.uk corr `e3eb4252` (45 rows —
  only the repaired news page's should drain; the owned/ported pages scan dirty in
  stored rendered_html and correctly STAY). ai-agent-orchestration deliberately NOT
  dispatched: its news page is unrepaired behind a needs_human_review sibling — a
  discovery run there drains nothing and the human decision is not mine to nudge.
- **⚠ Stated for the robot-hands arm before the result arrives:** the check scans STORED
  surfaces, and robot-hands news.html is the STALE-DEPLOY page (stored clean, served
  stale since 08-11) — so retraction there closes items for a page whose SERVED copy
  still shows defects. That is the check's design (DB-is-not-the-website); the served
  gap remains the rerender lane's routed follow-up, and closing the repair rows is
  correct because the repair half genuinely succeeded.

## 2026-08-20 ~07:40Z — RETRACTION PROVEN LIVE: 16 stale rows drained, dirty rows correctly stayed; one kcat drop caught by three absences

- The three induced discovery runs COMPLETED within a minute of dispatch. **Retraction at
  the artefact:** dartsonline.com 3→0 open (all `complete`, result reason "literal_markdown
  re-scan: page's unlocked components carry no markdown syntax on either surface");
  robot-hands.com 9→0; webdesign.co.uk 4 closed, **42 correctly remain** (owned/ported
  pages scan dirty in stored rendered_html — the 301/tool-rebuild population, plus the
  human-review row). The closed bug's last self-closing claim is now witnessed, not
  assumed. My worry that the orchestrate wrapper was a no-op ("scheduled task pre_query
  already did the work") was wrong in the way that matters: the dispatch DID drive the
  checks; the artefact (the drained queue) is the proof, per house rules.
- The robot-hands stale-deploy caveat stands as pre-stated: its items closed on stored
  surfaces while news.html still serves the 08-11 file — the rerender lane's routed gap,
  not a retraction defect.
- **The FIRST rerender witness dispatch (corr `343edda2`) never arrived** — kcat -P
  silent-drop landmine, live: kcat exited clean, but (1) no row by corr, (2) zero
  page-rerender rows in a window where 32 other orchestrations spawned, (3) zero hits in
  either label pod's logs. One drop in four dispatches through the identical `-c 1`
  heredoc pattern — `-c 1` reduces, does not eliminate. **Re-dispatched as corr
  `5dc60934`** — a legitimate re-run because the first act provably never happened
  (contrast yesterday's council double-submit, where it had).
