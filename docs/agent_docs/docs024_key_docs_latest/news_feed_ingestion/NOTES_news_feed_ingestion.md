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

Built migration 684 (`event_extracted_at` on `content_feed_items`, additive,
DO/RAISE drift guard + verify block per the 682 style precedent) and the two new
actions (`load_feed_items_for_event_extraction`, `mark_feed_items_event_extracted`
in `feed_event_extraction_actions.go`), extended `VerifyAndRegisterCitationsAction`'s
field pass-through list, registered both new actions. Committed (a7a134af7)
along with a correction to the PLAN doc's design section — see below.

**Correction, caught by a peer session before commit, not after.** The `bugs_open/427`
peer session read this lane's dirty tree (it's a shared working tree — anyone can
see uncommitted files) and messaged that the design's `kind="event"` was wrong:
`EvidenceFact.Kind` (`platform/orchestration/datahelpers/claims.go`) has a CLOSED
vocabulary — `metric | capability | entity | attestation`, `count`/`metrics`/`counts`
as the only aliases — and `"event"` is in neither. Verified directly against the
source before accepting: `CanonicalKind()` would silently map an unrecognised kind to
`"metric"`, and `UnrecognisedKinds()` (called from `validate_page_content.go` on every
build) would log a warning naming it, on every site, forever. Fixed to
`kind="entity"` (`datahelpers.FactKindEntity`) in both the code comment and the PLAN
doc (correction written visibly, not silently). This was still only a design/comment
value at that point — no LLM prompt existed yet to have baked the wrong literal in,
so nothing shipped wrong, but it would have on the next step.

Added a test (`TestVerifyAndRegisterCitations_EventFieldsPassThrough`) that decodes
the actual JSON written to `site_specs.data` via a `driver.Value` matcher
(`registeredEventFactHas`, modelled on `write_build_items_routing_test.go`'s
`specHandlerIs` — a real JSON decode, not a substring check) rather than trusting
the action's return value. Mutation-tested it myself: temporarily replaced the
four-field addition with a dummy string via `sed`, confirmed the test fails with a
clear "arguments do not match" error naming the wrong JSON, restored the correct
list, confirmed green again. (Note for future self: doing this via `sed` rather than
Edit means the harness can't tell a self-inflicted mutation from an external edit —
it flagged the file as "changed on disk since you last read it." Harmless here since
I immediately reverted to the intended state, but worth using Edit for temporary
mutations when the file matters, to avoid the false "another session touched this"
signal.)

Full test suite for the package: clean except pre-existing failures in
`provocation_gate_action_test.go`, confirmed via `git status` to be another
session's dirty uncommitted WIP (files I never touched), not caused by anything here.

`TestNoNewSilentScanLoss` (the bugs_open/410 scan-swallow ratchet) flagged my new
loader's `continue`-on-scan-error as an uncounted new instance of the silent-loss
shape. Fixed properly rather than suppressed: wired an `offered` counter through
`datahelpers.ScanShortfall` after the loop, following `loadStoredSections`'
worked example exactly (offered++ per `rows.Next()`, `// scan-loss:accepted:
counted` comment on the continue, `ScanShortfall(offered, len(items), subject)`
returned as an error). This action isn't a wholesale-replace like
`loadStoredSections` (nothing gets deleted if a row is dropped, it's just excluded
from this extraction pass), but strict-refuse-on-any-loss is still the safer
default and matches the established idiom — didn't invent a graded policy for a
case with no evidence yet that strict is too aggressive.

Commit's own hooks flagged two more things, both handled:
- **Optional-key-budget parity drift** (RFC_022/WFA-013): the new action's one
  optional key (`max_items`) wasn't in `check.py`'s `OPTIONAL_KEY_COUNTS` literal
  yet. Regenerated via the command in the file's own comment, diffed to confirm a
  single-line addition, committed separately, then re-applied the kustomize
  overlay (`kubectl apply -k .../optional-key-budget-check/overlays/production/uk_001`)
  so the cluster isn't running the stale literal — per CLAUDE.md's explicit
  instruction that editing `check.py` without re-applying leaves the old literal live.
- **"Architecture signal: migration + platform code in one commit"** — read the
  2026-07-29 owner ruling before treating this as an RFC trigger: an addition to a
  shared vocabulary needs an RFC only when it changes what the mechanism
  GUARANTEES, not merely because it's additive and shared. This change is opt-in
  (existing candidates that don't set the four new fields are unaffected) and
  reachable by nothing yet (no LLM prompt exists to emit them) — the RFC_002/
  2026-07-29 precedent names exactly this shape as NOT architecture-scope.
  Proceeded as a normal council-gate submission, noted the reasoning here rather
  than silently deciding it alone.

Submitted to council review: `SUBMISSION_CORR=4849c95f-2594-48e6-87b9-acee6341b0f8`
(the four edits: the new action file, the `evidence_citations.go` field-list
extension, the `registry.go` entries, migration 684). Committed BEFORE submitting
(the commits above), so neither commit carries a `Council-Submitted:` trailer —
missed doing that in real time. Forward-only forbids amending, so the trail is
imperfect: when the verdict lands, the `Council-Reviewed: <corr>` trailer will go
on a *later* commit (the workflow-wiring one) with a note that it covers these
earlier commits too, rather than on the commits it actually reviews. Flagging this
here so it isn't silently lost — a small process gap, not a technical one.

Replied to the peer confirming the fix, the mutation test, and the council
submission; gave them the go-ahead on fix candidate #2 (refresh_evidence_base
extension) and the interim-render design, both of which this lane's PLAN already
deferred, so no overlap.

Peer's `composeWriterBlock` fix (candidate #2 part 1, `f865153f8`) verified
independently: read the diff, ran its own tests, confirmed field names match
this lane's extraction fields exactly. Recorded in PLAN.

Council verdict for candidate #1 came back: verified directly against
`diagnosis_artifacts` (not taken on the peer's report, though they flagged it
first) — **APPROVED**, 5 advisory objections, none high-severity. The one worth
real follow-up: `architecture` seat flags that the four event fields are written
only via raw map keys, and the typed `EvidenceFact` struct used elsewhere for
parsing may silently drop them on any round-trip ("consistent with the existing
struct-roundtrip landmine") — checking this directly next.

> **Mistake, recorded visibly rather than left implicit.** Committing the
> approval record, I ran `git commit --allow-empty -m "..."` with **no
> pathspec** — reasoning (wrongly) that `--allow-empty` meant "this commits
> nothing new." It doesn't: `--allow-empty` only permits a commit whose diff
> against HEAD is empty; without a pathspec, `git commit` still commits
> whatever is in the INDEX, exactly the bare-`git commit -m` shape CLAUDE.md
> explicitly forbids. The index held other sessions' staged files left over
> from session start (`git status` at the very first turn of this session
> already showed them staged) — two `bugs_open/`→`bugs_closed/` deletions
> (359, 364) and three `sql_for_agents/` files (a new 213 migration, two
> already-superseded 648/654 `_HOLD` file deletions). **Checked for actual
> harm before writing this off as harmless**: `git log` shows 359/364's
> `bugs_closed/` halves and 648/654's renamed non-HOLD names were ALL already
> committed in earlier commits (`f5108dd47`, `ace31f564`) — what rode along was
> only the STALE leftover deletions of the old paths, already-orphaned by
> those earlier commits, not a live half-completed move. The 213 file is
> complete and coherent (read in full — proper header, BEGIN/COMMIT, guard
> blocks), not a half-written passenger. **No data lost, no half-move landmine
> triggered** — but the commit message still doesn't mention any of it, which
> is exactly the "four threads' work under one thread's message" harm CLAUDE.md
> names, regardless of luck. Lesson: `--allow-empty` is not a substitute for a
> pathspec — it changes whether git ALLOWS a no-diff commit, not what gets
> swept into it. Run `git status --short` immediately before any commit that
> has no explicit file list, empty-diff or not, and if genuinely nothing of
> mine needs a pathspec (a pure record-keeping commit), check the index is
> actually clean before trusting `--allow-empty` to mean that.
