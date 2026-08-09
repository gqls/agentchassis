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
