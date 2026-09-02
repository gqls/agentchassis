# NOTES — bugfix_427_event_render (append-only, newest at the bottom)

2026-09-02. Session renamed `bugs_open/427` to resume this bug. Read the bug
file in full, ran `who-owns.py 427` (OWNED/recently active — `calendar_component`
filed it, but its own §7 disclaims the fix as "not this lane's build"). Checked
`ListAgents`: three sessions started in the last hour with directly relevant
names — `calendar` (the filing lane, already closed out its involvement per its
own NOTES commit `cce085c5f`), `feed lane`, `gap planner`. Messaged all three
plus `boxingonline.com` before doing anything else, per the task's instruction
to coordinate rather than duplicate.

`boxingonline.com` replied fast with two corrections to the bug file, both
independently verified against the live DB before folding in: (1) the fact
count for the site had moved 3→1 since filing — one of the two seeded facts was
the owner's own billing email, wrongly registered as a publishable claim, and
removed per `bugs_open/420`'s privacy ruling; makes the case stronger (of two
facts ever seeded, one was a defect), not weaker. (2) the page-role diagnosis
(`d6d350ec…`) the bug pointed at as "in progress" had already completed —
`UNVERIFIABLE`, iteration-cap, not a confirmation or refutation of anything.
Both corrections committed to the bug file (`b2f35f9c8`) before doing further
work — citing a stale figure I could have just re-derived would have been
exactly the mistake this estate's own memory warns about.

**`feed lane` had already opened `news_feed_ingestion` and started
implementing candidate #1** — found this by grepping docs for `entity_ids`
while researching (not by being told), then checking `git status` for the
actual dirty files (`evidence_citations.go`, `registry.go`, a new action file,
a migration). Read their PLAN in full. Their design: reuse
`VerifyAndRegisterCitationsAction` unchanged except extending its field
pass-through list with `event_date`/`venue`/`participants`/`broadcaster` —
better than what I'd been about to design (a whole new registration action),
because it's the estate's own "reuse before building new" idiom applied a
third time (after `verify_and_register_directory_claims`).

**Caught one real defect in their still-uncommitted diff**: their comment said
`kind="event"`. `EvidenceFact.Kind`'s canonical vocabulary
(`datahelpers/claims.go`, `bugs_open/105`) is closed —
`metric | capability | entity | attestation` — and was fixed by MEASURING live
usage across every site before ruling on it. `"event"` isn't in that set or its
aliases, so `CanonicalKind()` would silently demote it to `"metric"` for any
typed consumer, AND `UnrecognisedKinds()` (called from `validate_page_content.go`
on every build) would flag it as an anomaly needing a fix, forever, on every
site. Flagged it before they committed; they verified independently and fixed
it to `kind: "entity"` (`a7a134af7`), with their own mutation-tested pin.

**Independently confirmed `directory_entities`/`directory_claims` (the
`render_directory_action.go`/`directory_claims.go` subsystem) is UNRELATED** —
a global, cross-site registry for AI-model/company/protocol directories, a
closed profile set (`directoryPublishProfiles`), explicitly "NOT per-site:
every opted-in site renders the SAME entities/claims." Pure name collision
with the "entity-directory" PAGE ROLE this bug's §6 also names. `feed_lane`'s
own PLAN reached the same conclusion independently — worth having two
independent reads land on the same answer given how easy the name collision
would be to walk into.

Dispatched a `Plan` agent (model `fable`, per the task's instruction) to draft
a full technical plan. **First mistake this session, corrected immediately**:
tried to send it a follow-up with new context using the `Agent` tool again
instead of `SendMessage` to its existing agent id — this spawned a SECOND,
unrelated agent (and with `isolation: "worktree"`, which was also wrong for a
read-only planning task). Caught it from the tool result before any real work
happened; stopped it with `TaskStop`, confirmed no stray worktree was left
(`git worktree list`), and sent the correction to the actual running agent via
`SendMessage` instead. No lasting effect — logged here as a note-to-self, not
in `WRONG_CALLS.md`, because it's a tool-usage slip rather than a false claim
about the system (WRONG_CALLS' actual schema).

The `fable` plan agent's FIRST returned draft (before my correction messages
had been processed) designed a full parallel implementation of candidate #1,
unaware `feed_lane` had already committed part 1. Did not act on it — a second
notification from the SAME task id arrived shortly after with a properly
reconciled plan (the tool's own docs note a task id can notify more than once;
this is why). Read only the second, superseding one. Worth recording as a
concrete instance of the estate's own principle that a subagent's report
describes what it intended, not necessarily the final state — the first
"completed" notification was real, and still not the version to build from.

Verified the second plan's key technical claims directly before building on
any of them (queryHandlers/sourceDependencies/ConsumerPages all exist exactly
as described, at the cited shapes) rather than trusting the agent's citations
at face value.

Traced `refreshCitationFact`'s actual dispatch condition myself
(`refresh_evidence_base_action.go:407`, `if _, has := src["citation"]; has`) —
confirms it fires on ANY fact carrying `source.citation` regardless of `kind`,
which `feed_lane`'s committed field-pass-through extension guarantees every
event fact will carry. This is the finding that collapsed candidate #2 from "a
new dispatch arm" to "already covered, verify and move on" — checked against
the actual committed code, not the bug file's speculative fix-candidate
wording.

Traced what actually CONSUMES `event_date`/`venue`/`participants`/`broadcaster`
once a fact carries them, rather than stopping at "the fact gets registered
and re-verified" — found `composeWriterBlock` only substitutes `{value}`, so a
`writer_line` phrased the natural way for an event fact would ship unsubstituted
braces into the writer's prompt. This wasn't in bug 427's own fix candidates or
in either plan agent's draft; found by reading the actual consuming function
rather than assuming "the register handles it." Built, tested, mutation-tested,
committed (`f865153f8`), submitted for council review
(`d0442d50-e383-477f-9ed8-19eaaeea3d93`).

Built the render target: `query.upcoming_events` resolver + `DepEvidenceBase`
dependency class + `queueEvidenceBasePageRerenders` producer hook
(`da2ab0d44`, submitted `08f56b7e-61e4-42d1-a3b6-13d700dd833c`). Confirmed
`consumerSQL` already excludes `rebuild_policy='owned'`/unshipped/archived
pages centrally, so no page-eligibility logic needed writing here — checked by
reading the function rather than assuming from the plan's framing of it as an
open question.
