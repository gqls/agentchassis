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

Owner said "please go ahead" on building + rolling the image and wiring the
workflow. Built `agent-chassis` from committed HEAD — **not** at the working
tree's current `IMAGE_TAG` (`v1.0.1351`): that value was itself an UNCOMMITTED
edit (another session's WIP), and `docker manifest inspect` confirmed
`v1.0.1351` already exists in the registry and is already the LIVE deployed
tag — reusing it would have overwritten someone else's already-shipped image
under the same tag with different content. Built+pushed+deployed at a fresh
`v1.0.1352` instead (confirmed unused first). Verified at the artefact, not
the roll: both new pods' `/proc/1/exe` contain my exact build commit
(`a2732c72...`, positive control) and do NOT contain the current (later, by
then) HEAD sha (negative control — deliberately NOT the "40 zeros" control,
which LANDMINES.md already documents as matching Go's binary padding and
hiding a stale build; used a real, meaningful sha instead). Both new action
names (`load_feed_items_for_event_extraction`, `mark_feed_items_event_extracted`)
present in the binary.

Wired the extraction chain into `feed-triage`'s live workflow config
(`agent_definitions.default_config`, DB-live, image already rolled so the
actions exist): `apply_scores` → `check_for_events` (gate) →
`load_for_event_extraction` → `extract_event_facts` (new LLM step) →
`register_event_facts` (`verify_and_register_citations`) →
`mark_event_extracted` → `complete`. Extraction prompt modelled directly on
`evidence-researcher`'s `extract_claims` step (same verbatim-quote,
machine-checked discipline: "the url will be re-fetched and rejected unless
your quote appears in it verbatim") — never invents `event_date`/`venue`/
`participants`/`broadcaster`, omits any field the source text doesn't state
rather than guessing. Applied via a guarded DO/RAISE transaction (drift check
before, full verify after each field), same discipline as the SQL migrations.

**Caught a real design gap by testing against real data before declaring this
done, not by inspection.** boxingonline.com already has 29 `content_feed_items`
rows at `status='relevant'` with `event_extracted_at IS NULL` (a real backlog
this mechanism should process on its first run). But `check_has_items` — the
step BEFORE `apply_scores` — short-circuits straight to `complete` whenever
there are zero `status='ingested'` items that cycle, which bypasses
`apply_scores` (and therefore my whole new chain) entirely. In steady state
(bursty ingestion, most triage cycles finding nothing new to score) this would
mean the extraction chain almost never runs, backlog or no backlog — the
opposite of what candidate #1 is for. **Fixed before it could misbehave in
production**: restructured so BOTH paths (fresh triage work this cycle, or an
empty ingest queue) converge on `load_for_event_extraction`, which gates
purely on its OWN live count (`extraction_items.count`) rather than on
whether this cycle's triage scored anything — `check_has_items`'s zero-branch
and `apply_scores`'s next_step both now point at the loader; `check_for_events`
moved to after it and keys on `extraction_items.count` instead of
`triage_result.relevant`. Verified the corrected shape with a DO/RAISE block
reading back all six fields, not just re-running the same UPDATE and trusting
it worked.

Then actually tested it, not just wired it: dispatched a real `feed-triage`
run for boxingonline.com (`scripts/kafka-publish-lib.sh`, the OPP-009 pattern
from `idea_uk_vm_site/scripts/dispatch_content_feed_orchestrator.sh`, adapted
to `agent_type=feed-triage`) — landed clean, correlation `bf4556cf-1ed4-4400-
b963-65b6e8d289d7`, watching progress via Monitor rather than a sleep loop.

**Second real mistake this session, caught immediately by the test rather than
by review — migration 684 was never actually applied.** I wrote it, verified
its own DO/RAISE guards were correct by reading them, and moved on — but never
ran it against the live DB. The dispatched run FAILED at
`load_for_event_extraction`: `column cfi.event_extracted_at does not exist
(SQLSTATE 42703)`. Exactly the value of testing the mechanism end-to-end
instead of stopping at "the code is written and the config is wired": a code
review would not have caught this (the Go code and the migration file were
both individually correct), only running it did. Fixed: applied the migration
by hand directly (`psql < 684_....sql`, its own verify block went green),
then registered it properly — `run-migrations.sh --record-only`, not a
hand-written ledger row — with a note naming exactly how it was caught.
Deliberately did NOT run the bulk `--apply` (CLAUDE.md: it takes EVERY pending
file, and "pending" on this tree almost never means "yours" — a background
dry-run I started to check the queue was still running minutes later against
what is evidently a long queue from other concurrent sessions; stopped it once
the hand-apply path made it moot).

**Re-ran the same dispatch after the fix — full chain green, real data, real
output.** `score_relevance → load_for_event_extraction → check_for_events →
extract_event_facts → register_event_facts → mark_event_extracted → complete
→ COMPLETED`. Checked the actual artefact, not the status: boxingonline.com's
`evidence_base` gained **6 new dated event facts** (kind="entity", `CIT-`
prefixed ids, e.g. a real, verified fight result — "Filip Hrgovic stopped
Moses Itauma in round 9... on August 30, 2026" at "The O2 Arena", citation
url `boxing247.com/...`, quote confirmed live by `verifyCitationLive`).
`participants` populated on all 6; `broadcaster` on 0 of 6 — checked this is
correct, not a bug: no source article in this batch stated a broadcaster, and
the prompt's "never invent" instruction held rather than fabricating one. 49
`content_feed_items` rows now carry `event_extracted_at` (the 29-item
pre-existing backlog plus 20 freshly triaged this run) — the backlog-draining
behaviour the `check_has_items` restructure was specifically fixed to enable.

This is candidate #1, live, working, against the actual motivating site.
What's left for 427 fully: candidate #2 (peer, in progress) and candidate #3
(blocked on the separate page-role diagnosis). This lane's build is done.

Moved to this lane's priority #2: `bugs_open/316`. Re-checked `who-owns.py` —
same read as before (real work clustered 08-22, everything since is
CONTRIB/routing shape). The ORDERING and CAPACITY halves are already fixed
and live (migrations 554, 556) — confirmed by reading the bug file's own §
history, not re-derived. Two riders were parked in the lane: the UK-news
TLD-default feature (real, owner-requested, unbuilt) and "unenrolled sites
never reach the seeder" (`CONTRIB_2026-08-25`).

**The second rider does not describe a real, current defect — checked before
building anything, per this lane's own established discipline.** The CONTRIB's
core claim: `find_news_sites`'s WHERE clause structurally excludes any site
with zero `content_sources` rows, so such a site can never reach
`content-feed-orchestrator`'s `seed_content_sources` step. Read the CONTRIB's
own quoted SQL closely first — it is truncated mid-clause ("`… FROM sites s
JOIN site…`"), cut off exactly where the eligibility predicate begins. Checked
the LIVE query (`SELECT ... FROM agent_definitions WHERE
type='content-feed-trigger'`): it has an explicit `NOT EXISTS (... content_sources
...) OR EXISTS (...)` arm — a zero-source site IS eligible. Traced this back
through history: `090_b_content_feed_trigger.sql` (the ORIGINAL seed) already
carries this exact `NOT EXISTS` arm; `554_news_feed_trigger_orders_by_the_schedule...sql`
(2026-08-22, three days BEFORE the CONTRIB) explicitly documents it by name
("arm A... is the PROVISIONING path... a newly-classified news site with no
sources yet is picked up here and seeded") and states `MEASURED 2026-08-22:
the state has ZERO live instances`. Empirically re-measured today
`[MEASURED 2026-09-02]`: 15 sites carry `news_feed.recommended=true`; 14 of
them have `content_sources` rows; the ONE exception
(`adversecreditmortgage.co.uk`) was classified 40 minutes before I checked —
`content-feed-refresh` last fired 08:58Z, next fire ~14:58Z, so it simply
hasn't had a cycle yet, not evidence of exclusion. No accumulating backlog of
stuck zero-source sites anywhere in the fleet. **Verdict: the mechanism works
and has worked since before the CONTRIB was filed** — the CONTRIB's diagnosis
was built on an incomplete read of the query.

**There IS a real, smaller, already-acknowledged residual — not urgent, not
what was filed.** `554`'s own author flagged it explicitly, not hidden:
"the real defect underneath is that provisioning and refresh share one capped
queue... recorded as a follow-up" (same wording repeated in the lane's own
PLAN). Because an unprovisioned site's `due_at` is `NULL` and sorts `NULLS
LAST`, it competes at the back of the same `LIMIT 10` queue as routine
refreshes — at today's fleet scale (15 recommended sites, well under the
limit) this has never bitten, but it theoretically could if the recommended
population grows well past the per-cycle limit. Recording this here rather
than building anything for it: no live defect, already tracked, not a lever
worth pulling on unmeasured risk.

---

**Session switch, resumed here.** Continued lane priority #2's third rider:
UK-news default for `.uk`/`.co.uk` sites. Confirmed via `who-owns.py 316`
that no one else picked this up in the interim and the tree had no dirty
files touching the target areas.

Finished the read the handoff left mid-way: `web_search_action.go` in full,
then the actual provider layer it hands off to
(`internal/adapters/websearch/adapter.go` + `providers/*.go`) — the handoff
had only read the orchestration-action half and correctly flagged "check
what the underlying search API actually calls it" as the next step.

**Found the exact mechanism, not assumed:** `providers.SearchOptions`
already declares `Region string // us, uk, etc.` — but nothing populates it
(`adapter.go`'s `SearchOptions{...}` construction sets only `SearchType`/
`TimeRange`) and nothing reads it except a dead comment in `duckduckgo.go`.
Checked the live deployment: `PRIMARY_SEARCH_PROVIDER=firecrawl`
(`kubectl get deploy web-search-adapter` env) — Firecrawl is the operative
provider, not a fallback. **Fetched Firecrawl's own `/v2/search` API docs**
(WebFetch, not assumed from memory): a `country` parameter geo-targets both
`web` and `news` sources and **defaults to `"US"`** when absent — almost
certainly the literal cause of "the news is from America". Also fetched
ScrapingBee's docs (the first fallback): `country_code`, same convention.
Both use `"UK"` as their value for the United Kingdom, matching
`SearchOptions.Region`'s own doc comment — no translation table needed.

**DuckDuckGo turned out to be a dead end for this specific bug, checked not
assumed:** `TestDuckDuckGoDeclinesNews` already proves DDG declines
`search_type: "news"` outright, and `FetchNewsSearchAction` always forces
`search_type: "news"` for a `news_search` source — so DDG structurally never
serves this pipeline's requests regardless of region. Its own hardcoded,
unconditional `kl=uk-en` (a separate, real, pre-existing oddity — every
*web*-type DDG fallback search is silently UK-locked today) is real but out
of scope: fixing it would be scope creep with zero effect on this bug. Left
it alone rather than touching a file "while I'm in there" for no benefit.

**Confirmed no existing TLD-derivation mechanism** before building a second
one (the lane's own open design question #4) — grepped
`platform/orchestration/actions/*.go` for TLD/suffix logic; nothing exists.
Reused `isBlockedDomain`'s `strings.HasSuffix` pattern (`helpers.go`) for the
new `regionForDomain` helper.

**Measured blast radius before writing any code**
(`[MEASURED 2026-09-02]`): 6 sites / 26 of 52 fleet-wide `news_search`
source rows carry a `.uk`/`.co.uk` domain (`farmerinsurance.uk`, `idea.uk`,
`loanandmortgagecalculator.co.uk`, `mortgagecalculator.co.uk`,
`remortgagecalculator.uk`, `webdesign.co.uk`), none carrying any
region-shaped key. This mattered for the design, not just as a sanity check:
`seedNewsSearchSources` inserts `ON CONFLICT (site_id, name) DO NOTHING`, so
a seed-time-only fix would never touch these 26 already-provisioned rows —
the owner asked for "all .co.uk and .uk sites", so a backfill migration is
required alongside the code change, not optional polish.

Full design written into PLAN before any code was touched (this file's own
lane discipline).

**Built, following the design exactly:**
- `seed_content_sources_action.go`: new `regionForDomain` helper +
  `seedNewsSearchSources` now takes `domain`, sets `config["region"]` for
  `.uk` sites.
- `feed_fetch_async_actions.go`: `FetchNewsSearchAction` passes
  `source_config.region` through, mirroring the existing `time_range` block.
- `web_search_action.go`: reads `config["region"]`, threads it into the
  adapter request body, the log line, and `Metadata` — parity with every
  other passthrough param.
- `internal/adapters/websearch/adapter.go`: `RequestPayload.Data.Region`,
  threaded into `SearchOptions.Region`.
- `firecrawl.go` / `scrapingbee.go`: set `country` / `country_code`
  respectively when `opts.Region != ""`; absent stays absent (provider
  default applies), so this is opt-in per source, not a global default flip.

**Tests written to the house pattern** (matched `search_options_test.go`'s
existing style at both layers rather than inventing a new one): region
passthrough at the action layer (`web_search_options_test.go`), at the
adapter layer including a real JSON-wire round trip
(`extractRequestPayload` parsing `"region":"uk"`, not just the struct
literal), and at each provider (`country`/`country_code` sent when Region is
set, omitted — not defaulted to empty-string sent — when it isn't), plus a
table test for `regionForDomain` including a deliberate near-miss
(`notreally.uk.com` — a different TLD ending in "uk.com", not ".uk").

**Verified against committed HEAD, not the working tree** — the working
tree currently has an unrelated pre-existing build break in
`apply_theme_kit_action.go` (another session's uncommitted WIP, confirmed via
`git status` on that file, not caused by anything here). Used
`scripts/verify-head-builds.sh --with <file> ... --test` (CLAUDE.md's own
prescribed tool for exactly this situation) rather than trusting a broken
`go build ./...` on the shared tree. All 9 overlaid files, both changed
packages, full test suite: green against HEAD.

**Migration `691_uk_news_search_region_default.sql`** written (+ `_ROLLBACK`
+ `_VERIFY`), following `608`'s single-statement-backfill style scaled to a
count-guarded multi-row UPDATE, matching `690`'s DO/RAISE discipline (a bare
`SELECT` verify cannot stop a `COMMIT`). Guard checks for exactly 26 pending
rows, treats 0 as "already applied" (matches the migration runner's own
probe convention), and refuses on any other count — the population may have
moved since the census. Not yet applied — that's the next step, following
candidate #1's exact playbook: commit, submit to council, THEN apply/build/
roll, never before.

Committed (`0a408f8db`) — pathspec-only, 14 files, commit-scope report showed
no passenger. Noted the pattern-check's `silent-reply-drop` flag at
`adapter.go:617` (`sendResponse`): pre-existing, nowhere near this change's
edits (which touch lines 44-49 and 217-227), not this task's scope — it's
already a tracked, named pattern (016b §9 / bugs_closed/062) whose fix is
its own RFC-scale item per the pattern-check's own text.

**Missed the `Council-Submitted:` trailer on that commit** — submitted to
the gate after committing, and forward-only forbids amending to add it
retroactively. Recording the correlation in this follow-up commit's message
instead (both name the same 691/691-adjacent Go+migration work; 098 will
join on the trailer in THIS commit, not the missing one on the code commit).

Submitted to council gate: dry-run validated clean first (`DRY_RUN=1`), then
the real dispatch. `SUBMISSION_CORR=8842fe96-9a71-4ea5-9993-2483f10712cb`,
orchestration `5947e320-9974-4a4f-b45c-00b727e7ffae`. Dispatch lane showed
LAG 0 at submit time (queue clear, several other councils in flight
concurrently — normal per the runbook). Following candidate #1's proven
order: wait for the verdict before building/rolling anything.

**Verdict landed: APPROVED, round 1, 3 advisory objections, none high-severity**
(verified directly against `diagnosis_artifacts`, not taken from the trigger
script's own printout). Checked the substantive ones rather than waving them
through — none required a design change, but the checking surfaced one real
fleet-wide finding:

- `reuse_agent` (medium): had the plan searched for an existing per-site
  locale/country signal before minting a new `region` key? Checked: `sites`
  has no market/locale/country column; `site_specs` is a generic
  aspect+jsonb store with nothing region-shaped found. Did find a related,
  non-colliding precedent — `vet-pipeline-orchestrator` and
  `area-sweep-orchestrator` already carry a `country: "GB"` config key, ISO
  form, for a different pipeline (business-directory area search) feeding a
  different external API. Different key name (`country` vs this lane's
  `region`), different consuming API, no collision — but the ISO-vs-vendor
  code choice ("GB" there, "uk"/"UK" here) is now visibly inconsistent
  fleet-wide for the same underlying concept, for the concrete reason that
  each was grounded in ITS OWN target API's documented convention rather
  than an estate-wide standard. Worth a future unification if a third
  region-aware pipeline appears; not blocking, not touched here.
- `guardian` (medium): named the blast radius rather than asserting it.
  `[MEASURED 2026-09-02]`: 11 agent types invoke `web_search`/
  `fetch_news_search` (`adoption-researcher`, `area-sweep-discoverer`,
  `brief-writer`, `directory-researcher`, `domain-research-classifier`,
  `evidence-researcher`, `feed-ingester`, `finance-directory-researcher`,
  `grounded-explainer`, `research-agent`, `vet-practice-verifier`); zero of
  them carry a `region` key in their live config today, confirming the
  change is inert for all 11 by measurement, not by assumption.
- `guidelines`/`architecture` (minor/low): should `region` be registered on
  an `ActionInputSpec`, and does it push `web_search` over the RFC_022
  optional-key budget? Checked both — and found something worth recording
  for its own sake, not just to close this objection: **`web_search` has
  NO `ActionInputSpec` registered at all** (grepped
  `RegisterActionInputSpec("web_search"` — zero hits), so
  `scripts/audit-optional-key-budget.sh` lists it under **"NOT COUNTED — no
  ActionInputSpec, so the optional surface is UNKNOWABLE, not zero"**
  (10 carriers, alongside `execute_llm_prompt`, `call_agent`, and 100 other
  live actions). This predates this change entirely — none of `web_search`'s
  existing 5 config keys (`query`/`num_results`/`search_type`/`time_range`/
  `provider`) were ever registered either, so adding one just for `region`
  would be inconsistent, not completeness. **The RFC_022 budget mechanism
  cannot see this action's surface at all, for reasons unrelated to this
  fix** — recording this here because a session running the audit script
  later could misread "not counted" as "under budget", which is the exact
  opposite of what it means. Not this task's fix (retrofitting an
  `ActionInputSpec` onto a 10-carrier shared action is its own
  properly-scoped piece of work); flagging for whoever next touches
  `web_search_action.go` or the RFC_022 tooling.

`debug_historian`'s two objections were about process visibility, not the
code: it read only the submission JSON's edit list, which named the main
migration file but not that `_ROLLBACK.sql`/`_VERIFY.sql` sidecars already
exist (they do — written before submission, following the `690` precedent
exactly), and it correctly flagged that no pod-binary verification step was
named as an explicit action. Noted for next submission: name sidecar files
explicitly in the edit sketch, and list the post-roll verification as its
own plan edit even when it's "just" a check.

Committing with `Council-Reviewed: 8842fe96-9a71-4ea5-9993-2483f10712cb`
next, then building both images.

Committed (`2f8411b7e`). Built both images from committed HEAD at a fresh,
confirmed-unused tag (`docker manifest inspect` first, per this lane's own
practice): `agent-chassis` (commit `f7dbb529...`) and `web-search-adapter`
(commit `1947e9bd...` — HEAD had moved between the two builds, another
session committing concurrently; confirmed both build points are still
descendants of this lane's own commit via `git merge-base --is-ancestor`
before treating that as fine). Pushed both to `v1.0.1353`. Bumped
`IMAGE_TAG` (makefile) and both services' production overlay `newTag` in
one commit (`886d693bc`) — re-verified the makefile hadn't moved again
first.

**Stopped before applying to the cluster to ask the owner**, rather than
rolling unilaterally. Reason: an auto-loaded memory
([[releases-are-whole-fleet-make-release]], dated 2026-08-03) records that
the owner previously blocked an unauthorised single-service `kubectl apply
-k` and asked that deploys go through him. This session's OWN earlier
candidate #1 fix did single-service deploy — but only after the owner said
"please go ahead" for that specific task; that authorisation doesn't carry
over silently to a different change. Asked via AskUserQuestion.

**Owner deployed a fresh chassis build himself** and asked for docs +
a session-switch handoff. Checked what's checkable without cluster access
(the request implies wrapping up now, not continuing to chase
verification): both `agent-chassis:v1.0.1355` and
`web-search-adapter:v1.0.1355` genuinely exist in the registry (`docker
manifest inspect`, doesn't need kubectl) — real builds, not a bare tag
edit. The overlay files carrying `newTag: v1.0.1355` are UNCOMMITTED in the
working tree as of this note (git history's last commit on those files is
still this lane's own `886d693bc` at `v1.0.1353`) — not this lane's file to
commit on the owner's behalf without knowing his intended message; noted,
not acted on.

**Could NOT verify at the binary — kubectl is down fleet-wide, confirmed as
the known 3-day token expiry, not a bug**: `kubectl get pods` returned
`Unauthorized` on every call; decoded the JWT `exp` claim directly
(`[[kubeconfig-token-expires-every-3-days]]`'s prescribed check) — expired
2026-09-02 22:08Z. Only the owner can refresh it. **So the following are
UNVERIFIED as of this note, not confirmed**: whether the live pods are
actually running v1.0.1355 (vs. a stuck rollout), whether that build
contains this lane's commit specifically (near-certain given the overlay
commit's ancestry and the standard `git archive HEAD` build process, but
"near-certain" is not "verified" — no binary sha grep was possible),
whether migration 691 has been applied (needs the same DB access), and the
live-`.uk`-site verification (Firecrawl request actually carrying
`country: "UK"`, results skewing UK) that this fix's own PLAN calls for
before declaring it done. **Next session's first job, once kubectl access
is restored**: all four of the above, in that order — pod rollout status,
binary sha grep (positive control this lane's own commit, negative control
a sha that should be absent), migration 691's pre-check query (apply by
hand + `run-migrations.sh --record-only` if still pending), then the real
`.uk` site dispatch. Full commands in `RUNBOOK_news_feed_ingestion.md`.

**Three cross-session messages arrived during this work, none yet acted
on** (in-scope for this lane, routed here correctly):
1. `portfolio_positioning` — owner wants `webpronews.com`'s RSS feed
   (`https://www.webpronews.com/feed/`, verified 200/~1.08MB/100
   items/multiple-per-hour) recorded as a candidate news source; write-up
   already committed to this dir as
   `CONTRIB_2026-09-02_from_portfolio_positioning_webpronews_feed_candidate.md`
   (commit `ebc050732`, not by this session). Caution: the owner endorsed
   the FEED's content, not the old advertise.co.uk Drupal consumer's
   wholesale-import pattern — needs this pipeline's normal editorial
   treatment, not a straight copy.
2. `designblog.co.uk` — `/the-design-feed/` serves zero items (0
   `content_sources` rows), routed here as "candidate-source work"; wants a
   DESIGN-vertical source when sources get wired (explicitly not
   WebProNews). Asked for an ACK.
3. `bugs_open/444` (empty listing pages) — diagnosed that `advertise`'s
   empty `/news/` page needs BOTH halves to fix: `content_features.news_feed`
   authored in its classification spec (no key exists at all today —
   `idea.uk`'s 2026-08-25 entry is the worked example) AND a `content_sources`
   row (rss, the WebProNews URL). Their plan-time gate (council pending,
   corr `c0990eb3`) will hold un-enabled news pages on a `capability_gap`
   named `news_source_enablement` — this lane's enablement work is what
   turns that receipt green. Their plan doc:
   `bugfix_444_empty_listing_pages/PLAN_2026-09-02_listing_source_gate.md`.

All three land in this lane's existing priority #3 gap (source enablement
for sites with zero `content_sources` rows) — genuinely new work, not
started this session, written into the new HANDOFF's own priority list
rather than actioned here given the owner's ask to wrap up now.
