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
