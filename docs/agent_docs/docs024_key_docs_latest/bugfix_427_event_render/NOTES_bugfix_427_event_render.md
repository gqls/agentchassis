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

Registered the new mechanism (PBP-048 addendum,
`docs026_concept_register/register/page-build-pipeline.md`) and added a
status roundup (bug file §9) covering all three lanes' progress in one place,
so a cold reader doesn't have to reconstruct it from four sessions' docs.

Checked council verdicts: both this session's submissions' composeWriterBlock
fix (`d0442d50…`) and `feed_lane`'s candidate #1 part 1 (`4849c95f…`) came
back **APPROVED** clean (3 of 10 reviewers engaged, no objections). No
follow-up commit needed — `098` credits a `Council-Submitted:` commit
automatically once its correlation resolves approved, per CLAUDE.md. The
render-target submission (`08f56b7e…`) is still in the queue.

---

2026-09-02/03 (fresh session, resuming from the HANDOFF). Picked up exactly
the two items the prior session named as needing a person: attach `event-list`
to boxingonline's live page, and deploy the admin-dashboard frontend (that
second one is bug 428's, covered here only because one session did both).

**Read the pending council verdict on migration 712 (`ff91e666…`) before
touching anything — it was REVISE, not silent.** Full report pulled from
`diagnosis_artifacts.body` (it's JSON, not text — `-x -t` on a plain `SELECT
body` reads as two lines because of how psql wraps a single JSON value; ended
up piping to a file and reading it structured). `editquality` raised two
non-gating objections, both concretely checkable against the live row rather
than argued: (1) does the migration's `input_schema` use the legacy
`properties` dialect the CHECK constraint `chk_input_schema_no_legacy_dialect`
refuses — checked, migration 712 uses the native `fields` dialect, and the
row's mere EXISTENCE in the table (`SELECT * FROM content_components WHERE
name='event-list'`) is proof the INSERT already passed that constraint; (2) is
`component_level` set explicitly — checked, it defaulted to `'section'`
(matches the constant `'section'::text` column default), NOT `'tool'`, so
`check_tool_health`'s `component_level='tool'` selector will never pick this
row up. Both objections closed by reading the live table, not by arguing.

**The GATING objection was `prior_art_librarian`, HIGH severity, and it was
right and load-bearing.** Quoted in full because it is the reason the rest of
this session went the way it did: the migration's own rationale asserted "the
ONLY route to attach this component is the full page-build-handler pipeline"
and used that to justify deferring the attach step. The seat pointed at a
NAMED, EXISTING, narrower mechanism the schema/landmine record already
documents for exactly this shape — `apply_section_edit`'s `component_swap`
edit type, part of the `section-editor` agent — and said the real open
question was whether IT can attach a newly-added library component, not
whether *any* narrower-than-full-rebuild path exists. Read the actual Go
(`platform/orchestration/actions/section_editor_actions.go`,
`applyComponentSwap`/`updatePageComponentSwap`) rather than trusting the
objection's framing either: `component_swap` repoints an EXISTING
`page_components` row's `component_id`/`slot_name` and re-renders from that
row's OWN (or caller-supplied) `content_data` — it targets one row, no other
section on the page is touched, and it goes nowhere near
`check_unresolved_sections`/`needs_rebuild`/the full replan pipeline the
migration's own header was worried about. This is a genuinely narrower,
already-built, already-safety-gated (lock/tombstone/composition-parent/
decision-citation/floors/regulated-identity/link-repair — ten call sites deep)
mechanism, and it resolves the exact open question migration 712 left
standing without needing to answer "does page-build-handler carry forward
existing sections untouched" at all.

**Checked which existing page_component row to target, rather than assuming
one.** boxingonline's `tool-fight-calendar` page had exactly two rows:
`hero-tool` (deployed, live) and `generic-text-block` (id
`cf6182ec-a98e-4ab3-a685-eb590090ab3b`, `build_status='pending'`,
`rendered_html` NULL — never actually rendered). Verified this second fact
against the SERVED page before touching anything: `sites.publish_project =
'boxingonline.ugg2.com'` (the domain itself, `boxingonline.com`, is not
DNS-live — `handed_over_at` is NULL — so the only way to see what a visitor
would actually see is the `.ugg2.com` preview subdomain, not the customer's
own domain; `getent hosts boxingonline.com` returns nothing, which the first
few attempts misread as a sandbox DNS restriction rather than "this domain
isn't live yet"). Curled `boxingonline.ugg2.com/tools/fight-calendar/index.html`
and found only `data-component="hero-tool"` on the page — the
`generic-text-block` row's prose was never live, so swapping it away removes
nothing a visitor has ever seen.

**Dispatched the swap directly**, mirroring `scripts/fire-section-edit.sh`'s
proven envelope shape (fetch section-editor's own live workflow from
`agent_definitions`, build the kafka message by hand, publish through
`kafka-publish-lib.sh` for the checked-publish receipt) but with
`edit_type=component_swap`, `new_component_function=event-list`, targeting
that row. Script kept at
`/home/ant/.claude-scratch/.../scratchpad/fire-boxing-event-list-swap.sh`
(scratch, not committed — the pattern is what's worth keeping, and it's
recorded here). Completed clean: `page_components` row now
`component_id`→event-list, `build_status='approved'`, then the workflow's own
`deploy_page`→`update_page_status`→`trigger_deploy` steps ran (git commit,
status→'deployed', a `deployer-agent` spawn that found nothing further to
commit since the git step had already landed it).

**Verifying the deploy actually reached the served page took far longer than
the deploy itself, and is worth recording in full because every step of it
was a real trap avoided, not padding.** The git commit (`007b3a7a1…`) landed
in seconds. The served preview (`boxingonline.ugg2.com`) still showed the OLD
page 20+ minutes later — cache-busted, checked `cf-cache-status: DYNAMIC` (so
it wasn't a stale Cloudflare edge cache, the ORIGIN itself was stale). Rather
than guess "propagation lag" and wait, used `gh run list --repo gqls/sites`
(this session had a working `gh auth status`) to find the ACTUAL GitHub
Actions run for the commit (`33672753667`, matched by commit message +
timestamp, not by trusting the newest row — several OTHER sessions' section-editor
edits to a DIFFERENT site were interleaved in the same run list, confirmed by
reading each one's own `gh run view --log | grep upload` to rule out a
same-file collision). `gh run view 33672753667 --log` showed the real "Sync to
B2" step: `delete tools/fight-calendar/index.html (old version)` /
`upload tools/fight-calendar/index.html`, then a Cloudflare cache purge for
`boxingonline.com` — the deploy genuinely succeeded at the infrastructure
level. The mismatch was a THIRD mechanism: `sites.publish_target='b2worker'`,
`publish_project='boxingonline.ugg2.com'` is a SEPARATE reconciliation pass
(`platform/orchestration/actions/publish_site_action.go`, agent type
`site-publisher`) that syncs `portfolio-sites/<domain>` onward to the preview
worker's own target on drift — distinct from the per-commit "Deploy to B2" GH
Actions workflow, and NOT triggered by a plain git commit. Did not chase
manually triggering `site-publisher` (it needs a spawned, storage-credentialed
pod — the standing chassis deliberately carries no B2 creds per that file's
own header — and building a correct spawn+call_agent envelope for it was not
worth the risk this session, given `portfolio-sites/boxingonline.com` — the
artefact that actually matters, the one `handed_over_at` would eventually
point real DNS at — was already confirmed correct). **Named, not chased: if a
future session needs the PREVIEW subdomain to show a change sooner than its
own reconciler tick, `site-publisher` is the action, and it needs a spawned
pod, not a direct `action=process` dispatch to the standing chassis.**

**The one real open defect this session found and could NOT close: `query.
upcoming_events`'s `items` field never populates during a light rerender.**
Reproduced THREE times, the last one under a chassis build 650 commits newer
than the one this bug's fix landed on (`7bf1ff674021f2d57dfd0aa41324541070646c3a`,
confirmed both `987ed3b3b` — the citation-gate REVISE fix — and `d6a952249`
— the `event_fixture_completeness` check — are ancestors, so this is not a
stale-binary artefact). Each time: dispatched `page-rerender` with
`spec.reason=section_data_resolved` directly (same envelope-building pattern
as the swap, workflow pulled live from `agent_definitions` for type
`page-rerender`), confirmed COMPLETED with no `__step_error`, confirmed
`rerender_sections.escalated=false` and a real `rerendered:2` count in its own
output — and confirmed via the DB, not the log, that `page_components.content_data`
for the event-list row is BYTE-IDENTICAL every time to what the original swap
wrote: still the OLD `generic-text-block` fields (`content`/`heading`), never
`items`/`headline`/`empty_text`, and `rendered_html` stays exactly 1813 bytes
(the same git blob sha256 across three separate git commits). One evidence_base
fact genuinely qualifies — `CIT-5b2cc9894bfc475f`, Canelo Alvarez vs Christian
Mbilli, `event_date: 2026-10-31` (correctly future relative to `2026-09-02`/
`09-03`), citation `url`+`quote` both present — the other five registered event
facts are past results (`2026-08-29`/`08-30`, correctly excluded by the
resolver's own `date.Before(today)` rule, which IS working as designed) or one
1998 historical fact. Tried to catch it live in the chassis logs three separate
times, the last with `kubectl logs -f` STARTED before dispatch specifically to
avoid missing a fast-completing run (the whole workflow completes in under 10
seconds) — captured the complete step-by-step trace of every generic
`coordinator.go`/`processor.go` infra log line, and ZERO business-logic log
lines from either `plan_sections_action.go`'s query.* branch or
`queryresolve/upcoming_events.go`'s own `logger.Info`/`logger.Warn` calls,
which per the source SHOULD fire unconditionally on every call. Confirmed via
`git merge-base --is-ancestor` that both files' code IS in the deployed
commit. **Did not find the mechanism. This is exactly the kind of claim
CLAUDE.md's diagnosis-before-debugging section says should go through the
`090_TRIGGER_needs_diagnosis` loop rather than get one more guess** — named
here with the full reproduction recipe so the next session (or the loop) does
not have to re-derive it. Candidate next steps, NOT tried: (a) read
`resolveComponent`/`classifyStoredSection` in `rerender_page_sections_action.go`
line by line rather than by grep, to settle whether the row is being CARRIED
(stored HTML reused, which would explain the byte-identical output and the
missing logs in one stroke) rather than freshly rendered — the byte-identical
output is equally consistent with "freshly rendered, items genuinely empty"
and "carried, template never re-run", and this session never distinguished
the two; (b) a scratch `page-rerender` dispatch against a DIFFERENT, already-
working query-sourced component (e.g. a `content-listing` row) to establish
whether the light-rerender path EVER populates a `query.*` array field, as a
control — never run, and it would settle (a) as a side effect.

**Fixed the drift the swap left behind.** After the swap, `pages.sections`
still named `generic-text-block`; left alone, `check_unresolved_sections.go`'s
own next sweep would have read that as "declared but unresolved" (the
component still exists in the library generally, just no longer joined to
THIS page) and marked the page `needs_rebuild` — reopening exactly the
full-rebuild risk this whole exercise was built to avoid. Migration `719`
(`docs/agent_docs/sql_for_agents/719_boxingonline_fight_calendar_sections_sync_event_list.sql`)
replaces the array entry, guarded by two `RAISE EXCEPTION` pre-checks (page
exists and still names the old slot; the event-list `page_components` row is
actually live) and a post-write assertion that CAN fail the transaction
(CLAUDE.md's own landmine about a "VERIFY BEFORE COMMIT" section written as
plain `SELECT`s that cannot fail — this one uses `RAISE`, not a comment).
Applied by hand (`psql -f` under an explicit transaction), NOT through
`run-migrations.sh --apply` — this tree had visibly concurrent sessions
mid-work (another site's "Section edit via section-editor" commits
interleaved with mine in the same `gh run list` window; a cross-session
message arrived mid-session from `designblog.co.uk` about an unrelated
migration 718) and `--apply` takes every pending file in the directory, not
just your own (LANDMINES, the `MIGRATIONS_DIR=…` entry). Recorded via
`--record-only` with a note naming what was checked. `pages.sections` now
reads `["advertising", "hero-tool", "event-list"]`.

**Admin-dashboard frontend (bug 428's item, done in the same session because
one session held both).** `make build-dashboard IMAGE_TAG=v1.0.1355` — did
NOT bump the shared makefile default (another session had already staged an
uncommitted `v1.0.1188→v1.0.1354` bump to the admin-dashboard kustomize
overlay specifically, of unclear provenance/currency; overriding IMAGE_TAG at
the invocation rather than editing the makefile avoided colliding with
whatever that was). Verified the built image actually CONTAINS bug 428's
release-surface UI before pushing — `docker run --entrypoint sh … grep -c
"Record verdicts only" /usr/share/nginx/html/assets/*.js` — 1 match, not
assumed from "all layers CACHED". `docker push` was refused by this session's
own auto-mode permission classifier (a production push, correctly gated) —
handed to the user rather than worked around. **The user (or another session)
completed it**: `admin-dashboard` deployment is now on `v1.0.1356` (one tag
higher than what this session built and pushed — so the eventual push/deploy
was NOT simply this session's own v1.0.1355 artefact waved through; re-verified
independently on the live pods, `grep -c 'Record verdicts only'` = 1, so
whatever built v1.0.1356 carries the same UI). Bug 428's own remaining item
("a human uses the release surface on a real verdict") is now actually
reachable — nobody has yet.
