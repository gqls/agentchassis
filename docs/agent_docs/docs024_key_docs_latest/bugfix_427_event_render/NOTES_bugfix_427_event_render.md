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

---

## 2026-09-03 (later) — the open defect from this morning is a fleet-wide re-render regression, and it is not ours

**Result first: `bugs_open/454` filed, fix committed `9831e9ab4`, council submission
`075cfedd-aef0-4230-b4f1-909ecf68959d`.** `classifyStoredSection` computes a section plan,
uses it to decide the row renders, and returns without carrying it. The struct field
`sectionClassification.plan` is **read at exactly one line in the repository and written at
none**. So `renderPlannedSection` gets a zero value, `plan.ResolvedData` is nil, and every
light re-render since 2026-09-02 12:27 BST has rendered `base ⊕ stored content_data` and
persisted the stored map unchanged.

**What I actually did, in order, because the order is the lesson.** The morning handoff
named two untried next steps and nominated a `090` run. I took the first of the two — read
`classifyStoredSection` line by line rather than by grep — before spending anything. Three
reads in:

```
grep -n '\.plan\b' platform/orchestration/actions/rerender_page_sections_action.go
1501:	comp, plan, htmlTemplate := cls.comp, cls.plan, cls.htmlTemplate
```

One hit. A read, no write. That is the whole diagnosis, and it took about four minutes from
opening the file. The `090` run was not fired — reason stated in `bugs_open/454` §7 rather
than silently omitted, per the 2026-07-31 owner ruling: there was no hypothesis for the loop
to refute, because the claim is a property of the source text rather than an inference about
behaviour.

**Then I wrote the test BEFORE the fix**, which mattered more than it usually does here.
`platform/orchestration/actions/rerender_page_sections_resolved_data_test.go` builds
boxingonline.com's real `event-list` component and a real register fact and drives
`rerenderFlatSections`. Unfixed, it produces — with no cluster involved —
`<section class="event-list"><p class="event-list-empty">No confirmed fixtures yet.</p></section>`,
which is the live page, byte for byte in shape. That is what turned "I have found something
odd in the source" into "this is the bug we have been chasing since yesterday".

**Where my own earlier claims in this file were wrong.**

> **CORRECTED 2026-09-03.** This morning's entry says I "captured the complete step-by-step
> trace … and ZERO business-logic log lines from either `plan_sections_action.go`'s query.\*
> branch or `queryresolve/upcoming_events.go`'s own `logger.Info`/`logger.Warn` calls, which
> per the source SHOULD fire unconditionally on every call". The source reading was right and
> the conclusion I drew from the absence was wrong. Under the real cause `planSection` runs in
> full — the resolver IS called, the query IS executed, the log line IS emitted; only the
> *result* is thrown away afterwards. So the log capture missed it. Three separate careful
> captures, one with `logs -f` started before dispatch, and I read the absence as evidence
> about the code instead of as evidence about my instrument. **The check I skipped:** a
> positive control — grep the same capture for a log line I KNEW must be there (the
> `coordinator.go` infra lines were there, but they come from a different logger and a
> different pod path than the action's own). An absence in a capture that never demonstrated
> it could see a present line is not an absence.

> **CORRECTED 2026-09-03.** This morning's entry framed the open question as carry-vs-fresh-
> render and said the byte-identical output "is equally consistent with 'freshly rendered,
> items genuinely empty' and 'carried, template never re-run'". Two options, and the truth was
> a third: **freshly rendered from stale inputs** — no carry, no empty resolve, the resolver
> working perfectly and its output discarded one function later. Worth noticing that I stated
> the disjunction as exhaustive without checking that it was. The *action* I proposed was
> still correct and is what found it.

**Numbers, measured today rather than carried.** `[MEASURED 2026-09-03]`, `clients_db`,
`page_components` at `build_status='deployed'` joined to `content_components`: **206** rows /
**196** pages / **21** component functions declare a `query.*`-sourced field; **1,855** rows /
**838** pages / **82** functions declare any non-`llm` source at all. The second is the real
blast radius — `plan.ResolvedData` carries every non-LLM resolution, not just `query.*`.
Query in `bugs_open/454` §4; re-run it before quoting these, a census goes stale by addition.

**Two things that masked it, both worth remembering as a shape.** `planSection`'s own
`carryStored` (bugs_open/238) re-supplies a non-LLM field from the page's stored
`content_data` when its source resolves to nothing — so on an already-populated page,
"resolved" and "stored-only" produce identical bytes. And the `cta_links_stale` recompute
allocates its own map when it finds `plan.ResolvedData` nil, so CTA repair kept working
throughout: the one re-render reason anyone was actively watching was the one still
functioning.

**A cost I incurred and should record.** My pathspec commit of the one-line fix took a
**same-file passenger**: the `bugs_open/450` session had an uncommitted rework of
`escalateRerenderToWriter` in the same file, and `git commit <path>` takes the file from the
working tree, so their half went into HEAD with mine. HEAD stopped building —
`pageRefusesGenericBuild`, `refusalToolPending` and an 8-arg `emitOwnedPageReviewItem`, all
still uncommitted. This is the documented trap (LANDMINES, "a pathspec commit still takes a
same-file passenger") and no hook can prevent it; what I could have done is **read `git
status` for my own target file immediately before committing and seen it was already dirty**,
which would have told me to expect this and to warn the other lane in the same breath rather
than after the fact. I measured the minimal closure that restores HEAD (six files, all
theirs: `owned_page_guard.go`, `multipage_actions.go`, `save_page_sections_action.go`,
`load_work_item_actions.go`, `get_pages_to_build_actions.go`, `load_page_record_action.go`,
verified with `verify-head-builds.sh --with`) and messaged the `bugs_open/450` session with
it rather than committing six files of another lane's in-flight refactor on my own judgement
of its readiness.

**What is still owed on 427 itself:** nothing new. The fix is Go, so it is inert until a
chassis image carrying `9831e9ab4` rolls; both live builds still carried the defect at
09:54 UTC today. When it rolls, re-dispatch the page-rerender and read the artefact.

---

## 2026-09-03 (later still) — verdict APPROVED, ff91e666 resubmitted, and a defect found in this lane's own migration while writing the resubmission

**454's fix is council-APPROVED** (`075cfedd`, 10:05:35Z, round 1): 12 of 13 seats approve, 5
abstain, no truncation, `decided_by: "approved with 1 advisory objection(s) — none
high-severity"`. All three advisories were about my SUBMISSION, not the code, and are
adjudicated in `bugs_open/454` §11. The one worth carrying:

> **`editquality` MEDIUM was right on the merits and wrong about what shipped.** I sketched
> ONE test where I had written TWO, so the seat could not see the classification-level
> assertion — and objected, correctly, that without it a future refactor could re-break
> `classifyStoredSection` while leaving the merge intact. That is precisely why the first test
> exists. **The runbook already says reviewers judge the sketch, because it is the only view of
> the code they get, and I did not follow it.** A second sketch block would have cost four
> lines and saved a MEDIUM. The other two advisories cost nothing: `plan`'s scope is settled by
> compilation, and `reuse_agent`'s "put it in the existing test file" is refused on a
> measurement — there IS no `rerender_page_sections_action_test.go`; this action already has
> four concern-named test files, so a fifth follows the convention.

**`ff91e666` round 2 is submitted** (dispatch `c46cf6c2`, trail keyed on `ff91e666`; verdict
still pending at the time of writing — the only `council_report` on that correlation remains
the 09-02 REVISE). It answers `prior_art_librarian`'s gating HIGH with events rather than a
better argument, and widens the edit list from `712` alone to `712 + 719 + 727`.

**The find of that exercise was a defect in this lane's OWN migration 719 — surfaced by
writing an honest rationale, not by any detector.** 719's UPDATE rebuilt `pages.sections` with
`jsonb_agg(DISTINCT x)` and no `ORDER BY`, which does not preserve input order:

```
before 719   ["hero-tool", "generic-text-block", "advertising"]
after  719   ["advertising", "hero-tool", "event-list"]      <- arbitrary order
intended     ["hero-tool", "event-list", "advertising"]
```

`pages.sections` is order-bearing BY INDEX (`save_page_sections_action.go:1979` states it;
`adopt_fragment_section.go` uses `planned[Position-1]`; `section_editor_actions.go` matches on
`p.sections->(pc.position - 1)`), so position 1 indexed to `"advertising"` and position 2 to
`"hero-tool"`. **Bounded, not overstated: not live damage** — that section-editor arm is gated
on an empty `slot_name` and both rows carry non-empty ones. Latent until a build leaves one
empty, and silent either way, because a wrong-but-present name matches nothing rather than
erroring.

Fixed by **migration 727**, applied by hand and `--record-only`'d. Two `RAISE` pre-checks, and
a verify block that asserts **the index alignment** rather than the array's literal value —
so it also catches drift it did not cause. **Rehearsed under `BEGIN`/`ROLLBACK`** (clean;
post-rollback control unchanged) and **induced-failure-proven**: writing a deliberately wrong
order made the verify `RAISE` and abort with the row untouched. That second step is the one
that is easy to skip and the only one that shows the guard can fire.

**Two things named and deliberately not done**, so neither reads as handled: `"advertising"`
is still declared in that array with no `page_components` row (pre-dates 719 and this lane);
and nobody has censused whether other pages fleet-wide carry the same index misalignment — 727
restores ONE page and claims nothing wider.

**Cross-lane traffic this session, recorded because two of the three were useful and one
corrected me.** (1) The `bugs_open/450` lane committed the six-file closure that restored HEAD
after my pathspec commit took their passenger — HEAD green at `13aac933f`, my fix and both
tests intact and part of that green, independently re-verified rather than taken on trust.
They also volunteered that `escalateRerenderToWriter` now returns a second refusal disposition
(`skipped_tool_pending_page`), which is in the RUNBOOK now: any count reading only
`skipped_owned_page` will undercount from the next roll. (2) The `gamedesign.uk` lane's
post-687 CONTRIB, relayed twice (once directly, once via `site-design-planner` through a name
collision), adjudicated into `bugs_open/428` §11 as a new residual of 687 rather than a
reopening of 428. (3) Answering their invited refutation turned up a real second producer of
`blog-post` pages (`create_blog_posts_action.go:238`, driven by a live `blog-content-planner`
row) — **and nearly had me repeat this file's own mistake of the morning**: my first check was
`orchestration_states`, which returned 0 runs and spans **24 hours**. A rolling window cannot
establish that a thing never happened. `llm_call_log` has the real memory: 10 calls, all
2026-04-03 → 2026-04-24, none since.

---

## 2026-09-03 (afternoon) — the roll landed, the fix is proven at the artefact, and the page is blocked one step further on

**Chassis `d0252fd4d` / `v1.0.1358` rolled 12:18Z carrying `9831e9ab4`.** Verified by
`git merge-base --is-ancestor` against the commit the STANDING pods report — and the documented
stale-row landmine fired on the way: a plain newest-first read of
`service_binary_capabilities` returned six rows for a **spawned**
`agent-image-build-handler` pod still on the old `7bf1ff67`. Filtering to
`pod_name LIKE 'agent-chassis-<rs>-%'` gave two standing pods agreeing on `d0252fd4d`, and that
agreement is the signal. **Had I trusted the first read I would have concluded the fix had not
shipped and stopped.**

**The dispatch and what it proves.** `page-rerender`, `reason=section_data_resolved`, correlation
`be75b209`. Counts came back `rerendered:2 carried:0 escalated:false` — **which is exactly what
the broken runs reported for a fortnight, so the counts are not the evidence and I did not treat
them as such.** The evidence is the `sections_metadata`, read against a control captured minutes
before the dispatch:

| | before | after |
|---|---|---|
| `event-list` content_data keys | `content, heading` | `content, heading, **items**` |
| `event-list` items | 0 | **1** |
| `event-list` rendered_html | 1,813 B (`md5 ee2ec068…`) | **2,498 B** |
| `hero-tool` content_data | 11 keys | +`hero_url`, +`background_image` |

**The `hero-tool` row is the finding I did not go looking for and the one that matters most.**
Nothing in 427 or 454 was about hero images. That section regained `planSection`'s authoritative
hero aliasing in the same pass, from a completely different non-`llm` source — which is the
1,855-row blast-radius claim demonstrated rather than asserted. One line, two sections, two
unrelated sources.

**Then the save refused**, and it is not this lane's defect:
`OWNED_PAGE_GUARD: page tool-fight-calendar is page_type=tool with no tool component`. That is
`bugs_open/450`'s `pageRefusesGenericBuild`, shipped in the SAME image (`587666be8`). Its
predicate genuinely holds — `page_type='tool'`, both components `component_level='section'`.
`[MEASURED 2026-09-03]` **58** tool pages refused across **12** sites, **53** of them on **9**
sites already serving. Concentrated: loanandmortgagecalculator.co.uk 16, loanzy.uk 5+, idea.uk 3.

**The interaction is the interesting part, and neither lane could have predicted it.** Until this
morning a re-render on those 53 pages ran, reported success and delivered nothing — so refusing
its save cost nothing observable. 454's fix and 450's guard arrived in the same image, and the
same refusal now blocks a real repair. Raised with the 450 lane with the measurement (they had
asked to be told), and **not routed around** — the scope call is theirs.

**A claim of mine to correct, from this morning.** §14 of the bug file said the page "fills in on
its own once 454 ships". 454 shipped, the data resolves, and the page still shows the empty
state. Corrected in place there. The lesson is one this lane has now hit twice in a day:
**"it will work once X lands" is a prediction about a chain, and I had verified only my own
link.**

---

## 2026-09-03 (15:10 UTC) — the fixture is on the page, and a shared-tree race taught the sharpest lesson of the day

**Both prerequisites finally lined up.** `29b40e8bc` (450's guard fix) rolled in a second chassis
build (`v1.0.1359`), and this time the check that mattered was NOT "did the pods restart" — it
was "did the guard fix's commit actually get shipped, checked by arithmetic against the running
commit's own timestamp", because the FIRST reported "fresh build" this afternoon (`v1.0.1358`)
turned out to be no roll at all: same standing pods, same commit, started before `29b40e8bc` even
existed. Recorded that as a dated negative in the handoff rather than acting on the claim.
**This time it was real** — new pods, new commit, `merge-base --is-ancestor` clean on both fixes.

Re-dispatched immediately. `COMPLETED`, all the way through `save_sections` (the step that had
refused twice today) into `render_page` and `deploy_page`. `items` 0→1, html 1,813→2,498 bytes.
Traced the deploy past the job status to the actual GitHub Actions "Sync to B2" step —
`delete .../index.html (old version)` then `upload .../index.html` — which is the standard this
lane set for itself yesterday and should not have been tempted to skip today given the time
pressure of a "wrap up for a new chat" request.

**Closed `bugs_open/454`.** Genuinely fixed, live, and its effect proven through the real
production pipeline rather than a test — that meets CLAUDE.md's bar cleanly. Moving it to
`bugs_closed/` produced the most instructive incident of the whole day, worth recording in full
because every step of it was a real trap, not padding.

**The race, exactly.** I appended a closure section and ran `git mv` to relocate the file.
Unbeknownst to me, the `components` lane — working `bugs_open/425`, an unrelated two-day-old
regression that turned out to BE this bug — had appended their own CONTRIB to the same file
**seconds before** my move, so `git mv` correctly carried their content along with mine to the
new path. Then their own commit, independently in flight, named the OLD path as its pathspec.
By the time it executed, the file no longer existed there (I'd moved it), so a pathspec commit —
which reads the **working tree**, not the index — recorded a clean 477-line **deletion** with no
corresponding add anywhere. For a few minutes, `HEAD` held **zero** copies of a file two
sessions were actively writing to.

**Nobody's fault, and nobody made it worse.** The `components` lane did exactly the right thing:
noticed via `git ls-tree -r HEAD`, diagnosed it precisely (parent commit has the content, my
index still holds the full file staged), and **declined to fix it themselves** — "the move is
your close decision" — flagging it instead. That restraint is what kept it a two-minute repair
instead of a second collision. I restored by naming only the surviving path (the old one no
longer matches anything git knows, so naming it errors) and verified at `HEAD`, not the tree,
before saying anything was fixed.

**The transferable shape, distinct from the already-logged same-file-passenger landmine**: that
landmine is about a `git mv`'s OWN two-sided commit dropping half of itself. This is different —
a THIRD PARTY's ordinary, correctly-formed pathspec commit, aimed at a path that used to be
correct, executing during the exact seconds a `git mv` invalidated it. Worth its own line in
LANDMINES rather than folded into the existing entry, because the fix is different: the existing
entry says "name both paths on your own move commit"; this one says "a pathspec commit can be
made wrong by SOMEONE ELSE'S concurrent move, and the tell is `git ls-tree -r HEAD` returning
zero rows for a file you know exists somewhere."

**Two content corrections folded into the closed file, both from the same peer, both real:** my
§14 had framed a still-unconfirmed canary as evidence for a question it could never have
answered (the value predated the regression entirely, planted by a build, not a re-render) —
caught before it could mislead a future reader into citing an unconfirmed line as settled. And
the closure claim needed one enumerated exception (`component_id NULL` rows, structurally
unreachable by this fix) rather than reading as an unqualified "every light re-render now
works".

**427 itself**: everything upstream of the artefact is now proven. What is left is not code —
it is a decision about `site_plan_sections`' immutability (§19) and a downstream detector's own
schedule (the nightly `experience_loop` reclassification). Written up as the lane's actual
remaining state rather than left implicit.

---

## 2026-09-03 evening — session "427" (lane resumed from HANDOFF_2026-09-03d)

### The handover, and why it was not taken on trust

The prior session was messaged before anything was written. It replied that it was not
still on it, and — better — it re-verified both of my findings at the source before
agreeing, rather than accepting handoff continuity. It also volunteered a second gap in its
own closeout (that it named the experience check as the closing signal without checking
whether that check *could* fire given what it had built). Recording that because a clean
handover is not the norm this file usually documents.

### Misstep 1 (mine): I said "this settles it" before I had read the reverting path

I read the re-plan path first, found `reconcilePlanWithRealised`'s snap
(`v3_site_actions.go:7701-7724`), confirmed live that `load_existing_pages` really does
supply `p.sections` and `p.build_status`, and concluded — out loud, to the owner — that
§19.2 was refuted and the migrations were safe. That was the same error §19.2 made, one
file along: I had read a **writer** and had not read the **reader** of the store that was
left stale.

What caught it: opening `discovery_checks/check_section_source_drift.go` and reading its
file header, which states the whole trap in six lines and names migration 153/154 as the
worked case. It was in the tree the whole time, and it is the header of the detector that
had already filed a work item about this exact page at 12:24 that morning.

The cheap check, now in `WRONG_CALLS.md`: **ask which code path READS the store you left
stale, and grep for it.** `grep -rn "FROM site_plan_sections" --include=*.go platform/`.

### Misstep 2 (mine): my triage query encoded the wrong definition of "authority"

First pass joined only tier 1 and reported `leopardessconsulting.co.uk/index` as a LIVE
DIVERGENCE. It is served by **tier 2** (the `site_specs.site_plan` aspect) and agrees. The
check's precedence is `COALESCE(tier1, tier2)`; mine was `tier1`. A NULL tier-1 read as
"divergent" instead of "look one tier down".

The tell was the **shape** of the output — `live_authority` came back NULL rather than a
list — not the number. Had I not noticed, migration 753 would have left a resolved item
open on the grounds that it was divergent.

### Misstep 3 (mine): a migration-number collision I created and then had to undo

Checked the highest migration number, got 747, wrote 750 (correct). Wrote the second
migration as 751 — and by then two other sessions had taken 751 and 752. No file was
overwritten (different basenames) but the number was duplicated, so I renumbered to 753
including its `$pre$`/`$post$` dollar-quote tags and every `751 ABORT` / `751 VERIFY`
string. **The number must be re-checked immediately before writing the file, not once at
the start of the session** — mine went stale inside forty minutes. Three separate 750s now
exist in the directory.

### What was verified, with the queries

Pre-state, all `[MEASURED 2026-09-03]`:
- `site_plan_sections` for the current plan: 3 rows, ids `d7bdc4c8` / `d74518a8` /
  `16a18d39`, `assigned_fact_ids='[]'`, `subject` NULL, and **all four** of
  `component_version_id`/`palette_id`/`layout_id`/`typography_set_id` NULL.
- exactly **one** `site_plans` row for boxingonline.com, `is_current`, and equal to the
  page's `built_from_plan_version`.
- **zero** `site_specs` rows with `aspect='site_plan'` — tier 2 does not exist for this
  site, so creating one (as migration 153 did for robot-hands) would have handed tier 2 a
  permanent say over tier 3 on every page.
- **zero** locked `page_components` rows — so `MergeLockedPageSlots` is the identity here
  and a raw list comparison is exactly what the drift check does.

Post-apply: the loader's own query returns `hero-tool, event-list`; its sync-down guard
(`:562`) would update 0 rows; **artefact byte-identical** — `hero-tool`/3,859 B,
`event-list`/2,498 B, `pages.updated_at` still `15:10:36.708975+00`.

### The induced failures — both fired

- **750**: pointed the post-write needle at `["hero-tool","advertising"]`, a value no write
  produces. Result: `UPDATE 1`, `DELETE 1`, `UPDATE 1`, then `ERROR: 750 VERIFY FAILED`,
  transaction rolled back, tier 1 still the stale triple. So the guard fires *and* the
  transaction is genuinely atomic.
- **753**: removed the "stores agree" predicate so it would close every open item. Result:
  `UPDATE 5` — including `apis.uk` — then `ERROR: 753 VERIFY FAILED: apis.uk/index should
  still be open (found 0 open items)`, rolled back. That is the guard that matters, because
  the whole design of 753 is "close by predicate so another lane's live case is excluded by
  the data".

### Council: `b290bef5` APPROVED round 1 — and the objections were worth reading

12 reviewers, 5 abstained, 4 advisory objections, none high-severity, no truncation.
Two were genuinely actionable and I checked both rather than banking the approval:

1. **`editquality`, MEDIUM: "727/728 are tagged in LANDMINES as instances of the
   `jsonb_agg(DISTINCT)` idiom that silently reorders — so tier 3's order may itself be
   corrupt, undermining your premise that you are aligning to a *correct* live page."**
   A very good catch about a real hazard, and **already closed**: the scrambling was done
   by **719** (`["hero-tool","generic-text-block","advertising"]` →
   `["advertising","hero-tool","event-list"]`) and **727 exists to repair it**. 728
   explicitly avoids the idiom. Confirmed the current order is right by agreement with
   `page_components` position order (1 `hero-tool`, 2 `event-list`). Premise holds — and
   now on evidence rather than assumption. Note 727's own header says that defect was found
   *"not by any detector"*.
2. **`guardian` LOW + `guidelines` missing: "closing the work item by direct UPDATE bypasses
   the verifier registry."** Checked: **no verifier is registered** for
   `section_source_drift`. It sits in `verifier_coverage_test.go`'s `catMechanical` backlog
   (`"predicate is check_section_source_drift"`), and `grep RegisterVerifier` finds nothing
   for it. So there was nothing to bypass. Worth knowing this is a *gap*, not an absence of
   risk — a mechanical item type with no verifier is exactly how the backlog in
   `bugs_open/469` accumulated.

Objections answered by the **file** but not by my **sketch** (a lesson in itself — the
council reviews the sketch): `editquality`'s "no `site_id` filter on the work-item UPDATE"
(the file has one), `guardian`'s "rollback should verify `plan_id`" (it aborts if the
current plan is not `bba66eda`), and `editquality`'s "assert non-zero row counts" (the
pre-check asserts the exact three-row aggregate). **Write the sketch as the file, not as a
paraphrase of it.**

Recorded and not actioned: `guardian` MEDIUM — the safety case rests on `decideEmit`'s
evaluation order, which SQL cannot enforce. True. Accepted as residual; the framework fix
is where a guard for it belongs. `debug_historian` MEDIUM — no dump/backup before a
production mutation. Fair; the induced-failure run and a hand-reconstructing rollback are
weaker than a dump.

**The `architecture` seat's objection is the one that matters, and it was already
satisfied:** *"file the RFC now rather than defer again"*, plus — the sharpest line in the
round — *"a detector that fires and does not prevent the loss it detects is not a working
safeguard; it is a log."* `RFC_064` was committed the same session.
