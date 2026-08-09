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
