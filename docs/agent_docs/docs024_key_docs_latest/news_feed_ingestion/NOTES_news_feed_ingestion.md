# NOTES — news_feed_ingestion (append-only, newest at the bottom)

2026-09-02. Lane opened. Received a charter handoff from the `calendar` session
(cross-session message) proposing this lane own the raw ingestion pipeline, with
`bugs_open/427` as the urgent pickup. Before accepting: read `bugs_open/427` in
full, ran `python3 scripts/who-owns.py 427` and `... 316`, checked `git status`
for any dirty feed-related files (none), read `news-feed-pooling-workstream.md`
memory and `bugfix_410_feed_phase_lock`'s commit log. All consistent with the
peer's account — accepted the charter, replied confirming.

Also separately fielded an earlier cross-session message (peer session
`bugs_open/427`, likely the same `calendar` session or a sibling working the same
bug) asking whether this lane's work overlapped 427 — at that point this session
had done nothing yet, so answered "no conflict, nothing started" truthfully, and
that stands: the overlap only exists now because this lane is *choosing* to pick
up 427's fix candidate #1, not because of any prior collision.

Read `platform/orchestration/actions/feed_triage_actions.go` in full —
`apply_feed_scores` sets `relevance_score`/`status`/`topics`/`credibility` on
`content_feed_items`, never touches `entity_ids`/`duplicate_of`. Confirms bug
427's writer-audit independently.

Checked `entity_ids`' actual column shape before trusting bug 427's fix-candidate
wording literally: `uuid[]`, no FK, no documented target (`\d content_feed_items`,
2026-09-02). `directory_entities` exists but is an unrelated feature (kind+slug
business/model directory, migration 192-era). The `news_editorial_features` lane
already found this and left a LANDMINE about it. **Decision recorded in PLAN**:
don't populate `entity_ids` for this fix — bug 427's own wording allows "a
comparable typed field," and the actual target is `evidence_base` per its §6.

Read `evidence_citations.go` (`VerifyAndRegisterCitationsAction`) and
`directory_claims.go` (`verify_and_register_directory_claims`) in full.
`directory_claims.go`'s own header states the reusable idiom explicitly: "This
file adds NOTHING to how a citation is verified... reused UNCHANGED" — same
`verifyCitationLive` + supersede-write pattern, just a different target store.
This is the third instance of the same shape (site evidence_base / cross-site
directory / now feed-derived events), which is why the design in PLAN reuses
`VerifyAndRegisterCitationsAction` directly (extend its field pass-through list)
rather than writing a fourth near-duplicate.

Read `feed-triage`'s live workflow config from `agent_definitions` (steps only,
not the full prompt text — see RUNBOOK for the query that avoids flooding the
terminal with a 273KB prompt dump, which is what the first attempt did).
Confirmed the step-chaining shape (`next_step`, `evaluate_condition`'s
`condition_field`/`conditions`/`default`) and that `apply_scores`'s `next_step`
is currently `complete` — that's the splice point for the new extraction chain.

Design written up in PLAN. Next: migration for `event_extracted_at`, the two new
Go actions, the field-list extension, tests, then the workflow-config wiring
(DB-live, done only after the image carrying the new actions is built and
rolled — sequencing matters, per CLAUDE.md).
