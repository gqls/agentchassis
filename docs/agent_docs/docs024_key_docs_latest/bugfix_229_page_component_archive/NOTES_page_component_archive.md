# NOTES — bugfix 229 page component archive (append-only, newest at the bottom)

## 2026-08-09 — session 1 (ruling, measurements, build)

- **Owner ruling received in-session**: candidate 1 — extend the 344 shape.
  Ownership checked first: who-owns clean (only the 226 lane's commits touch
  the file); three live transcripts MENTION 229 but all are passive reads of
  the 226 status header. Ruling recorded in the bug file before any work.
- **Measurements that changed the design** (all in the PLAN with figures):
  DELETE dominates page-side (19,054 vs 4,928 updates all-time) so the DELETE
  arm is the load-bearing one; write rate 27-290 rows/day kills the fail-open
  argument; save_page_sections's own snapshot drops the artefact exactly when
  content_data exists (COALESCE at :714) — the house archive records the
  wrapper, 14,831/14,863 rows.
- **FK archaeology that forced two design points**: history.component_id is
  ON DELETE SET NULL, but a NEW row referencing a dying component is rejected
  ⇒ delete-rows carry component_id NULL, identity in new slot_name/position
  columns. history.page_id FK + page_components' ON DELETE CASCADE ⇒ archiving
  during a full-page cascade is structurally impossible ⇒ the one soft path:
  trigger skips when the pages row is gone (740 pages deleted all-time; a
  deliberate page deletion is not the silent-section-wipe class).
- **Misstep, caught by my own probe (and worth the WRONG_CALLS-adjacent
  note): the first apply of mig 357 FAILED its own DO/RAISE verify** —
  "patched overwrite classified machine_made". The trigger was right; the
  PROBE was wrong: every row written in one transaction shares created_at =
  now() (xact start), so ORDER BY created_at DESC LIMIT 1 among probe rows is
  ARBITRARY. Rewrote each check to select its row BY ITS BYTES (the
  row-identity rule applied to my own verification). Second apply: all four
  arms passed (negative no-op, machine_made, hand_patched, delete-with-NULL-
  component), probe self-cleaned, both triggers enabled 'O'. The failure
  itself is evidence the verify block can fail — an induced-check requirement
  satisfied by accident.
- **Go half**: page_component_divergence.go (classify mirrors the destructive
  statements' predicate INCLUDING pageComponentAgentWritableSQL — a locked
  row survives the rebuild and must not be counted as destroyed; ledger
  read-back built in from the start rather than earned in round 2 like
  chrome's). Stamps: save_page_sections INSERT, rebuild_blog_listing both
  arms, section_editor both arms, create_report_page both arms — all
  md5(html) same-statement. adopt_verbatim deliberately NOT stamped (ported
  bytes are not reproducible from content_data) with a test pinning the
  ABSENCE. Loud: save_page_sections (classify before DELETE, emit after
  RowsAffected>0) + rebuild_blog_listing UPDATE arm (filtered to its one
  component). 14 tests green (-count=1); one pin needed its anchor widened to
  span the VALUES clause (a match region that stops before the stamp cannot
  see it).
- **Council**: corr `eee2888b-20dc-46ba-9b1f-53e592374cba`, submitted ~19:15Z
  before the commits (the 07-30 rule); schema note for the next session — the
  097 trigger wants plan as an OBJECT {summary, edits[], grounded_in, risks},
  not a bare edits array; it refuses otherwise, client-side, before spending.
- Migration numbering: the directory moved 351→356 during THIS session's 226
  work (five numbers taken by other sessions in ~2h) — re-ls at write time is
  not paranoia, it is the base rate.

## 2026-08-09 — session 1 (council round 1 → REVISE → round 2)

- Round 1 verdict ~20:21Z (step `complete_revise` — read the REPORT, not the
  step name): **REVISE**, gated by `reuse_agent` (HIGH). 6 abstained.
- **The gating objection is an artefact of after-the-fact review, and the
  evidence is decisive**: the seat read the LIVE schema — which already
  carried THIS submission's own applied migration — and concluded
  `page_components.rendered_html_digest` pre-existed ("the plan risks
  duplicating an existing stamping mechanism"). Three independent proofs it
  is self-created: (1) this session's own `\d page_components` BEFORE the
  design (~17:52Z, in-conversation verbatim) lists no such column; (2) the
  ABORTED first apply's verification query returned count=0 for the new
  columns at ~19:0xZ — the column did not exist between the failed and
  successful applies; (3) the seat's own read-only check returned
  `non_null=1331, populated=0` — a column with zero stamps has no "other
  writer" to collide with, and the fleet grep shows exactly 7 Go files
  touching the name: my five page-side + chrome's two (different table).
- **Objections CONCEDED and fixed (the code asks)**:
  - bug_historian MEDIUM: loudness widened from 2 to 4 writers —
    `apply_section_edit` + `create_report_page` UPDATE arm now classify+emit
    (each filtered to its one component, after success; the locked path
    returns early). Still quiet by design: adopt_verbatim + non-Go writers
    (trigger archives; no Go seam). PLAN carries the revision block.
  - bug_historian LOW: predicate-parity pin added
    (`TestSavePageSectionsDeleteUsesSameWritablePredicate`) — classifier and
    DELETE both anchored on the shared helper, drift now breaks the build.
  - guardian MEDIUM: consumers-told WIDENED to the full eight-class writer
    inventory in the PLAN — the told-set now matches the affected-set.
- **Objections ANSWERED with evidence (no change needed)**:
  - guardian HIGH (procedural, "not a veto" by its own note): the 07-29
    owner ruling is quoted verbatim in the resubmission — condition (1)
    retired ("review here is after the fact, by design"), condition (2)
    registration-same-commit satisfied (STY-056 in `4f4189290`). Plus the
    operational answer: the probe caught its own bug BEFORE commit, the
    rollback is one statement, and `agent_error_log` monitoring for
    `page_component_history` mentions reads 0.
  - architecture MEDIUM: STY-056 LITERALLY restates the threshold — "a
    THIRD adopter needs the shared-abstraction RFC" appears in both the
    register entry and the 000 index line, quoted in the resubmission.
  - prior_art MEDIUM ×2: the cffbfec4 round-2 architecture seat text
    fetched VERBATIM from diagnosis_artifacts and quoted — it not only says
    what was claimed, it NAMES page_components as the anticipated second
    instance ("site_components now, page_components proposed next via
    bugs_open/229. Fine at two instances; a third table adopting the same
    shape without a shared abstraction would be the point this needs an
    RFC").
  - guardian LOW (digest as freshness signal elsewhere): the 7-file grep —
    no reader outside the two divergence mechanisms exists.
- All tests green after revisions; round 2 submitted on the same trail
  (`RESUBMIT_CORR=eee2888b…`).

## 2026-08-09 — session 1 (round 2 → REVISE → round 3)

- Round 2 ~21:0xZ: **REVISE, gated by editquality (HIGH)** — and the defect
  was in MY SUBMISSION, not the code: the round-1→schema restructuring script
  split the combined edit entry `"section_editor_actions.go +
  create_report_page_action.go"` on " + " and KEPT ONLY THE FIRST file, so
  the round-2 edit list promised four loud writers and named three. The code
  ships all four (`ebb879fc1`); the JSON under-declared. A reviewer can only
  see the submission — an edit list is a CLAIM, and mine was false by
  transformation. Round 3 fixes the list (create_report_page as its own
  edit; the migration + rollback entries merged to stay within 8).
- **reuse_agent: APPROVE** — the round-1 gating objection (column already
  exists) accepted the three-proof answer. The after-the-fact-review artefact
  class is now demonstrated on this trail: the reviewer's schema contains
  the submission's own applied half.
- **Guardian mediums, both EXECUTED not argued**:
  - Migration ledger: 351 AND 357 were applied by single-file psql but never
    RECORDED — a later `--apply` would have re-run both (the exact NNN/ledger
    landmine). Both now `--record-only`'d with notes naming the probe story.
    A real catch; the 344 practice said "record" and I did it for neither.
  - Rollback drill: the sidecar's statements executed inside a
    BEGIN…ROLLBACK — DROP×3 clean, guard NOTICE fired ("0 triggers remain
    inside txn"), and after ROLLBACK both triggers verified still live and
    enabled ('O'). The drill proves execution + guard without ever exposing
    the fleet to an unguarded window.
- **bug_historian medium (unenumerated writers archive-but-never-surface)**:
  filed as STY-056 open-review (e) WITH the sweep query written out — the
  ledger's trigger-written verdicts make it one indexed query; its natural
  driver is the 230 rotation once 083 gives detection a drain. Not folded
  in: a sweep with no driver is a helper with no callers, and file-don't-fold
  is the disposition the owner ratified for this bug's own birth.

## 2026-08-09 — session 1 (round 3 → REVISE → round 4)

- Round 3 ~21:4xZ: **REVISE, gated by editquality again — and again the
  submission, not the code.** The plan's RISKS field still said "loudness
  contained to save_page_sections + rebuild_blog_listing...
  section_editor/create_report_page overwrites archive but do not emit" —
  ROUND-1 text, carried through two rewrites while the edit list shipped
  four emitters. Three seats (editquality, guardian ×2) flagged the
  self-contradiction. **The lesson generalises: a multi-field submission is
  a document, and every field is a claim — I updated rationale, summary and
  edits each round and never re-read risks.** Round 4 rewrites risks to
  state the four-emitter reality once, WITH the blast radius sized (zero
  emits on day one — 0/1331 rows stamped; steady state bounded by
  hand-patch frequency, not write volume; the chrome sibling's production
  day measured ~1 item).
- debug_historian medium (no pod-grep step): answered honestly — the Go
  half is inert until the owner's whole-fleet release, so today's pod-grep
  would verify an absence we assert (1274 predates the code). The
  verification plan is now IN the submission: post-release pod-grep by
  image+ownerKind enumeration with positive AND temporal-negative controls,
  then the e2e protocol. Already the bug file's close criterion.
- Round 4 submitted on the trail. No code change since round 2.

## 2026-08-09 — session 1 (round 4: APPROVED)

- **Round 4 APPROVED** (~22:1xZ, orch `d99f987d`): "approved with 1 advisory
  objection(s) — none high-severity", 3 abstained. Verified at the report
  body in diagnosis_artifacts, not the step name or a watcher echo.
- The one advisory is bug_historian's standing unenumerated-writer concern —
  already filed as STY-056 open-review (e) with the sweep query and its
  driver named; it carries forward as a watch item, same shape as the chrome
  sibling's guardian advisory.
- **Trail summary, 4 rounds**: r1 REVISE (reuse_agent read the submission's
  own applied column as pre-existing — the after-the-fact-review artefact
  class, answered with three proofs and approved in r2) · r2 REVISE
  (editquality: my restructuring script dropped a file from the edit list —
  the code shipped, the claim didn't) · r3 REVISE (editquality: the risks
  field was round-1 prose carried through two rewrites — every field of a
  submission is a claim) · r4 APPROVED. **The code was council-clean from
  round 2; rounds 2-4 were the paperwork being made to tell the truth about
  it.** Two real catches along the way that were not paperwork: the
  unrecorded migrations (351+357, now ledgered) and the undrilled rollback
  (now drilled).
- This commit carries `Council-Reviewed:` legitimately — verdict read before
  the trailer was written. The three code commits keep `Council-Submitted:`;
  098 resolves them at report time.
- **Remaining, blocked on the owner's fleet release (v1.0.1275)**: pod-grep
  (positive + temporal-negative), then the e2e protocol. Both in the handoff.

## 2026-08-09 — session 1 (release landed, pod-verified, e2e RUN AND PASSED)

- **The release landed as v1.0.1276** (not the 1275 I bumped — the release
  machinery or another bump moved it on; the TAG identity mattered less than
  the binary proof). Pod-verify by image + ownerReferences.kind: 4
  ReplicaSet-backed pods at 1276 (both main agent-chassis replicas grepped
  `classifyPageComponentArtefacts`=4, WARN string=1,
  `page_divergence_overwritten`=2, chrome control=2); 18 Job-owned
  stragglers at 1274 are pre-roll spawns, noted not claimed.
- **E2E protocol run end-to-end on dartsonline "beginners" (~19:47Z), every
  signal by row identity**:
  - (a) page-rerender with `spec.reason='section_data_resolved'` (the mode
    gate: only image_landed/section_data_resolved/cta_links_stale reach
    `save_sections`; any other reason assembles stored HTML and stamps
    NOTHING — worth knowing before waiting on stamps that cannot come).
    Result: 3/3 sections stamped-and-matching; the trigger's DELETE arm
    archived the 3 outgoing pre-fix rows (`op='delete'/unstamped`) — its
    first production rows.
  - (b) psql patch (append probe comment to hero) → the patch itself drew
    `overwrite/machine_made/psql` archiving the pre-patch machine bytes
    (87e7c66eee4f) — the raw-psql writer class proven visible page-side.
  - (c) rerender → **the WARN fired** (page_component_divergence.go:154, on
    the rendering pod, in my orch), item
    `page_divergence_overwritten:page_component:5009f5c8:1:d267b8ea64b5`
    (exact predicted key: page8/position/digest12), ledger row
    `delete/hand_patched` with archived md5 == patched md5, component_id
    NULL, identity via slot+position. All 3 rows re-stamped, probe gone.
  - (d) negative control: untouched rerender → NO WARN, NO new item, 3 more
    `delete/machine_made` rows (the by-design delete-arm archive, correctly
    silent). DELETE recoverability is (c)'s own row — the hand_patched
    delete row IS the byte-exact recovery copy.
  - Census after all passes: 3 unstamped + 5 machine_made + 1 hand_patched
    deletes, 1 machine_made overwrite, 1 item, 0 trigger errors, 3/3
    orchestrations COMPLETED. Probe item `d89fcb4b` cancelled with a note.
- **Bug 229 is DONE IN SUBSTANCE** — stays in `bugs_open/` per the owner
  08-06 ruling. Remaining watches: STY-056 open-review (a) volume, (e)
  unsurfaced-writer sweep (driver = 230 rotation once 083 drains).

## 2026-08-10 — first standing-watch reading (both clean)

- **Watch (a) volume**: `page_component_history` total 30MB (was 29MB at
  design time — growth consistent with the ~0.9MB/day worst-case projection,
  no pruning urgency). Trigger-arm rows: 08-09 (partial day from mig apply)
  109; 08-10 so far 63 (38 unstamped-delete + 21 machine_made-delete +
  4 unstamped-overwrite). Save-path snapshot rows continue at their historic
  rate (59 today vs 100–550/day over the prior fortnight). Note the stamped
  share of delete-arm rows is already overtaking unstamped on 08-10 (21 vs 38,
  was 11 vs 91 on 08-09) — the restamp-through-natural-churn curve doing what
  the plan predicted.
- **Watch (e) unsurfaced-writer sweep** (query verbatim from STY-056 open
  review (e)): **0 unsurfaced**. The zero is meaningful: precheck counted 1
  `hand_patched` trigger row fleet-wide (the 08-09 E2E row) and the sweep
  excluded it, so the NOT-EXISTS join is proven to MATCH at least once
  (item_key digest join exercised, not vacuously empty). Driver for wiring
  the sweep remains the 230 rotation once 083 drains — nothing wired today,
  by design (a sweep with no driver is a helper with no callers).

## 2026-08-19 — second standing-watch reading (new session picking the lane up); closure re-verified; FILE MOVED; watch (a) TRIPPED

- **Closure re-verified before anything else** (the fixed-AND-live bar was
  restored by the owner on 08-12, superseding the 08-06 keep-in-bugs_open
  direction, so the first question was whether the move is earned TODAY):
  both fix commits ancestors of the live chassis build `d3590ca46`
  (provenance verified at the binary with controls this morning by another
  session, `c0ed34c13`); both triggers enabled 'O'; zero fail-closed
  refusals in `agent_error_log` since 08-09; 20 open
  `page_divergence_overwritten` items prove the loud half fires. Full
  evidence written into the bug file as a dated section; file moved
  `bugs_open/` → `bugs_closed/` with both paths named on the commit.
- **Watch (a) volume: TRIPPED.** 63MB total (was 30MB on 08-10) — ~3.5MB/day
  against the ~0.9MB/day worst-case projection, 4×. Cut by op/divergence
  since 08-10: delete/machine_made **4,085** (75%), delete/unstamped 1,075,
  overwrite/machine_made 189, overwrite/unstamped 101, delete/hand_patched
  22, overwrite/hand_patched 6. By site: robot-hands.com 660 rows / 4.7MB
  is the biggest single contributor; the rest spread across the loan/vet
  calculators, vonc, webdesign. The driver is the delete arm archiving
  machine_made bytes on routine DELETE+INSERT saves — the projection
  underestimated save frequency, not a defect. **Retention design now due
  on the design's own terms** ("volume/pruning decided on page-side
  measurements" — we have 9 days of them). Being taken up as its own
  council-gated change, this lane, next entry.
- **Watch (e) sweep: 6 unsurfaced** (was 0). All `application_name='psql'`,
  all webdesign.uk hero, two three-row same-transaction bursts (08-13
  16:40:58, 08-14 20:14:19) — a psql session iterating its own hand-patch;
  each write archived-but-did-not-surface its predecessor. No net loss (the
  final patches' destruction by stamped writers raised items — webdesign's
  17), but two lessons recorded in STY-056: the sweep counts same-actor
  self-iterations, and raw-psql patching is live behaviour, not a
  hypothetical. Driver for wiring remains 230-rotation-once-083-drains;
  083 (`detected_findings_never_reach_a_handler`) still open, so still not
  wired — and the 20 undrained items above are that bug's live cost here.
- **Reader census for the retention design** (do NOT prune without this):
  runtime readers of `page_component_history` = save_page_sections_action,
  rerender_page_sections_action, content_data_envelope_guard,
  section_visible_text, save_sections_prune_floor (calibration),
  page_component_divergence (ledger read-back), save_component_history_action,
  core-manager page_admin_handlers; plus one-off restore recipes (mig 287,
  378, 431) which read **content_data** from history rows. Key precedent for
  what may be dropped: the save-path snapshot's own COALESCE policy has
  always dropped the artefact when content_data exists (14,831/14,863 rows)
  — machine-reproducible artefact bytes are the class the platform already
  declines to keep, and the trigger arm was built to catch the DIVERGENT
  class. Retention that nulls only old machine_made trigger payloads (never
  hand_patched/unstamped, never content_data, never the ledger row) is
  inside the established semantics.
- **Misstep to own**: my first `agent_error_log` query guessed `created_at`
  as the timestamp column (it is `occurred_at`) — caught by the error, fixed
  by `\d` per the schema-first rule. In-flight, never asserted anywhere, so
  recorded here rather than WRONG_CALLS.

## 2026-08-19 — retention SHIPPED (mig 489): open-review (a) resolved on the watch's own trip

- **Design**: PLAN_2026-08-19_history_retention.md (measurements, preservation
  contract, reader census, why machine_made is the droppable class — the
  save-path snapshot's own COALESCE policy, not this session's invention).
- **Mechanism**: scheduled task `page-component-history-retention` — the
  `database-cleanup` pure-pre_query shape, fire_message=false, daily. NULLs
  `rendered_html` on trigger-arm rows machine_made AND >30d; everything else
  survives (ledger row, content_data, digest, slot/position). One doc_notes
  row per run, on zero too (WFA-013 precedent). Partial index
  `idx_pch_retention_candidates` keeps the daily scan bounded.
- **Applied + verified 2026-08-19 ~11:05Z**: single-file psql, ON_ERROR_STOP,
  in-transaction DO/RAISE probe (three synthetic rows on a real page FK: old
  machine_made pruned with content_data surviving; old hand_patched untouched;
  recent machine_made untouched; self-cleaned). Ledger row recorded
  (`--record-only` with note). **Run 0 induced manually with the pre_query
  verbatim**: `payloads_nulled=0, bytes_freed=0, note_written=1` — the zero is
  the predicted one (the trigger is younger than the 30-day window; first
  eligible row ages ~2026-09-08), and the report row proves the mechanism
  executes end-to-end. **Verify-later, dated**: (1) within 24h, a
  SCHEDULER-driven doc_notes row should appear —
  `SELECT created_at, body FROM doc_notes WHERE
  subject_key='page-component-history-retention' ORDER BY created_at DESC;`
  (2 rows = manual + scheduled; a missing second row means the scheduler is
  not running the task); (2) ~2026-09-09, the daily row should show non-zero
  prunes and table growth should flatten (`pg_size_pretty(
  pg_total_relation_size('page_component_history'))`, 63MB baseline today).
- **Council: NOT submitted, disposition stated** (the norm asks for the gate
  on platform-code changes; this is config-only): 097 refuses client-side —
  no edit touches platform/, internal/, pkg/ (owner scope ruling 07-17) —
  and, decisively, `scheduled_tasks` is OUTSIDE every seat's hardcoded
  11-table schema view, which the RUNBOOK itself records as making such a
  round "unwinnable by construction". Review surfaces used instead: the
  induced probe, run-0 induction, the ledger note, STY-056's updated entry,
  a LANDMINES entry (NULL payload after 09-08 = retention, not a broken
  trigger), and the ~20-day structural abort runway
  (`UPDATE scheduled_tasks SET enabled=false WHERE
  name='page-component-history-retention'` is live immediately).
- **Rollback**: 489_ROLLBACK removes task + index; nulled payloads are not
  recoverable — stated in both files.
