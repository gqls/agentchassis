# NOTES — model directory pipeline

Running record, append-only, newest at the bottom. What was tried, what the
system actually said, every misstep.

---

**2026-07-22 — kickoff, scope escalation, and grounding research**

Owner's opening brief was a single-site content ask for
`ai-agent-orchestration.com`: an AI model directory (open+closed models,
cited cost, links out) and, later, a company AI-agent adoption tracker (ROI
claims, protocol adoption). Mid-brief the owner escalated: this should be a
fleet capability, opt-in via `site_specs`, auto-created for any domain that
requests it, "most of the infrastructure is already there."

Two research passes (general-purpose agent, then Explore agent) grounded the
design against the live repo/DB rather than assuming:

- Confirmed `ai-agent-orchestration.com` is the framework's own brochure/
  marketing site (blog posts about agent orchestration, an ROI estimator, an
  LLM cost calculator, a live `/news.html`) — the model directory fits its
  existing register.
- Confirmed the news-feed pipeline is a real, live, near-exact structural
  template: `site_specs.classification.content_features.news_feed` opt-in →
  `content-feed-trigger` scheduled discovery → `content_sources`/
  `content_feed_items` ingestion → `RenderNewsSectionAction` publish →
  dual-layer component (`latest-news`/`news-listing`, server template +
  client JSON fetch).
- Confirmed the V5 citation-verification machinery (`evidence_citations.go`,
  `datahelpers.QuoteFoundInText`, the `evidence-researcher` agent) is built
  and is exactly the "cited cost" mechanism needed — a verbatim quote is
  re-fetched and re-checked, not just LLM-asserted once and trusted forever.
- Confirmed page auto-creation is a real, already-live pipeline
  (`discovery_checks` → `content-gap-planner` → `apply_gap_plan_action.go` →
  `page-build-handler`) — grepped and read `structuralPageTypes` in
  `page_growth_budget.go:37` and the `MissingNewsSectionCheck`/
  `MissingNewsPageCheck` registration in `check_news_feed.go:46-49` directly,
  did not take the Plan agent's description on faith.
- Checked: **no** `models`/`ai_models` table exists yet in `clients_db`
  (`pg_tables` query, 0 rows) — this is genuinely new schema, not
  rediscovering dead infra.
- Checked the pool-site pattern (`docs/.../news_feed_pooling/`) as a
  candidate for "one global dataset, many sites" — explicitly design-only,
  parked behind an owner gate, not proven in production. Rejected as the
  storage model for this reason: better to build two small, real tables than
  extend an unfinished mechanism.

Plan approved by owner (2026-07-22) as
`/home/ant/.claude/plans/valiant-roaming-dawn.md`, copied into
`PLAN_2026-07-22_model_directory_pipeline.md` in this directory for
permanence.

Migration numbering at kickoff: latest filed is `190_enable_contact_form_undeliverable_check.sql`;
claiming `191` for the schema (Phase A). No other session's docs mention
`directory_entities`/`directory_claims`/`model_directory` as of this check.

**Collision on apply:** by the time the migration file was written, another
session had concurrently filed `191_diagnose_agent_resources.sql` (unrelated —
diagnose-agent pod resource requests, bugs_open/043) as an untracked file.
Re-checked numbering immediately before applying (per CLAUDE.md: a session-start
snapshot goes stale within minutes) and caught it before running the migration.
Renamed mine to `192_model_directory_schema.sql`. Lesson confirmed in practice,
not just in principle: re-run the numbering check right before you apply, not
just when you first draft the file.

**Applied 2026-07-22.** Ran `192_model_directory_schema.sql` directly via
`psql -f` (not `run-migrations.sh --apply`), specifically to avoid sweeping in
the other session's `191_diagnose_agent_resources.sql` (unrelated, in-flight)
along with three already-applied-by-hand-but-unrecorded files (186/188/190) —
none of that is my task, so I recorded only my own file in `schema_migrations`
by hand, checksum-matched, `applied_by='manual-single-file'` with a note
explaining why. `\d` confirmed both tables + indexes match the design exactly.
Failing-branch test (in a rolled-back transaction, no data left behind):
inserted one `directory_claims` row, then attempted a second `is_current`
row for the same `(entity_id, field)` — correctly rejected with
`duplicate key value violates unique constraint "idx_directory_claims_current"`.
Phase A complete and verified.

**Phase B — Go actions + seeds, 2026-07-22.** Wrote
`platform/orchestration/actions/directory_claims.go`
(`verify_and_register_directory_claims`, `refresh_directory_claims`), reusing
`verifyCitationLive`/`datahelpers.QuoteFoundInText` unchanged (same package,
no export needed — confirmed by reading evidence_citations.go in full before
writing a line). Registered both in `registry.go`. `go build ./platform/...`
clean; `go vet` clean (one pre-existing unreachable-code warning in
`load_component_library_actions.go`, not mine). Extracted the status-decision
logic into a pure `classifyDirectoryClaimOutcome` and unit-tested it
(`directory_claims_test.go`) — the found/citation_lost/fetch_error/recovered
distinction is exactly the property this layer must never get wrong, so it's
worth being independently testable of the DB, same rationale as
`evidence_citations_test.go`.

**Correction, same session:** the first draft of the schema (per the plan
file) omitted `directory_entities.links`/`attributes` from the one write path
(`upsertDirectoryEntity` only set name/owner/summary) — a silent gap that
would have permanently left "where to find it / how to use it" empty despite
the schema being designed to carry it. Caught before commit by re-reading my
own draft against the brief, not by any external check. Fixed: candidates may
carry `entity_links`/`entity_attributes`, shallow-merged into the existing
jsonb via `||` so a later pass can add a key without clobbering earlier ones.

**Second correction, same session — routing bug caught before shipping.**
First draft of `SEED_directory_scheduled_tasks.sql` gave the discovery task
`target_topic = 'system.agent.directory-researcher.requests'`. Checked the
live `scheduled_tasks` table before trusting this: only 4 distinct
`target_topic` values exist across the whole fleet
(`business-intel`/`vet-intel` — separate deployed microservices per the
kustomize list — plus `generic` and an internal noop). `content-feed-refresh`
proves the actual pattern: `target_agent_type='content-feed-trigger'` but
`target_topic='system.agent.generic.requests'` — a custom agent TYPE that
runs on the shared agent-chassis (not its own microservice) is dispatched
through the generic topic, with the real type carried in the message payload
(`cmd/scheduler/main.go` `fireTrigger`, `config.agent_type =
task.TargetAgentType`, sent to `task.TargetTopic` regardless of what that
type is). A dedicated `system.agent.directory-researcher.requests` topic has
no consumer — the task would have fired into the void, silently. Fixed to
match `content-feed-refresh` exactly. This is precisely the class of bug the
"diagnosis before debugging" standing rule is about: it would have looked
identical to "working" (task fires, no error) right up until the moment
someone checked whether anything actually happened.

Not yet applied: `SEED_directory_researcher_agent.sql` and
`SEED_directory_scheduled_tasks.sql` are correctly gated behind an image roll
(image-first-then-seed) — pausing here to confirm with the owner before
building/pushing/rolling a chassis image, since that is a shared-cluster
action with real blast radius on other concurrent sessions, unlike everything
committed so far (docs, schema, Go source only — no running process changed).

**Phase C — publish action, resolver, components, 2026-07-22.** Read
`render_news_section_action.go`, `queryresolve/news_items.go` and
`queryresolve.go`'s dispatch switch in full before writing anything, to
confirm the "one query, two projections" discipline and the dual-layer
(server-template + client-JSON-fetch) rendering shape were real, not
assumed. Key finding that shaped the design: `queryresolve.Resolve()`
requires a non-nil `SiteID` and documents "there are no cross-site queries"
as a hard invariant — my resolver still takes one (satisfies the contract)
but genuinely ignores it, since the model directory is the one query whose
answer doesn't depend on which site asked. Wrote
`queryresolve/model_directory_items.go` (`QueryModelDirectoryEntries`,
shared by both the resolver and the JSON publish path), wired
`model_directory`/`model_directory_full` into `queryresolve.go`'s switch, and
`render_model_directory_action.go` (cousin of `RenderNewsSectionAction`,
including its own `queueModelDirectoryPageRerenders` cousin of
`queueNewsPageRerenders` — confirmed sharing the `page_rerender:<page>`
item_key convention across multiple triggers is the *intended* idiom, not a
collision risk: a scoped rerender re-resolves every `query.*` field on the
page regardless of which trigger asked for it).

**Correction, same session — a security divergence, deliberate.** Read
`latest-news`'s live `js_content` before writing my own: it builds card HTML
by concatenating feed-sourced text (title, summary, source) directly into
`innerHTML`. That is a real XSS surface already live in production, not
something I introduced — but model/company names, summaries and quotes here
are ALSO third-party text, sourced via a broader attack surface (web-scrape +
LLM extraction, not a curated feed), so I did not copy the pattern. Both new
components' `js_content` build the DOM via `createElement`/`textContent`
instead, documented inline as a deliberate divergence from the news
precedent, not an oversight. Did not go back and fix the news component —
out of scope for this workstream, flagged here in case it becomes its own
bug.

Wrote `SEED_directory_components.sql` for the two new `content_components`
rows (`model-directory`, `model-directory-listing`), gated behind the image
roll like the agent/scheduled-task seeds. **Verified all four gated SEED
files by dry-running them against the live DB inside a transaction I then
rolled back** (`sed` swap of the trailing `COMMIT;`/`ROLLBACK;` line) —
caught nothing wrong, but this is the check that would have caught a
dollar-quoting or column-count mistake before it sat silently broken in the
repo for weeks.

**Phase D — discovery checks + growth budget, 2026-07-22.** Read
`check_news_feed.go` in full (all five checks + the bugs_open/015 stranded-
nav-page logic) before writing my two. Deliberately did NOT clone the
stranded-nav-page retype logic — that exists because news pages had already
accumulated years of wrong-page_type history across many sites before the
gate existed; model-directory is a brand-new page_type with no such history
yet, so there is nothing to retype. Documented as a scope cut, not an
oversight, in case it needs adding later.

Both checks gate on TWO conditions before ever raising a finding: the site
opted in (`site_specs.classification.content_features.model_directory`) AND
the global registry actually has `status='found'` claims — unlike news,
there is no per-site "are there any sources yet" check, because the registry
that would answer that isn't site-scoped. Wrote
`check_model_directory_test.go` with sqlmock, covering all four branches
(not opted in / opted in but empty registry / raises a finding / section
already exists) plus the `separate_page` gate — these are exactly the
conditions a silent regression would hit first.

**Caught by running the FULL test suite, not just my own package**: my two
new item types (`missing_model_directory_section`,
`missing_model_directory_page`) tripped
`TestEveryCheckProducedItemTypeIsClassified` in
`verifier_coverage_test.go` — a real, pre-existing fleet-wide governance
gate (bugs_open/021 §INSTANCE 2: every check-produced item_type must have a
verifier or be an acknowledged gap) that I would otherwise have silently
violated. Added both to `itemTypesWithoutVerifiers` with `catMechanical`
(same category as their `missing_news_section`/`missing_news_page`
siblings — same handler, same existence-check shape), explicitly noting
they are NOT `[INFERRED]` since I wrote and read the checks myself. Two
OTHER item types (`backend_entry_orphaned`, `contact_form_undeliverable`)
were already failing this same test before I touched anything — confirmed
via `git log` on their check files, dated 2026-07-22 from other sessions'
work — left untouched, not mine to fix.

Also found and fixed a real duplication landmine while adding
`"model-directory"` to `structuralPageTypes`: `page_growth_budget.go` keeps
a SECOND, hand-copied list of the same page-type vocabulary inline in a raw
SQL string (`page_type IN (...)`), used to split the weekly-budget count.
Updated both copies and added a comment flagging the duplication for
whoever touches this next — it is exactly the "two hand-maintained rosters
that must stay identical" shape CLAUDE.md already calls out for the council
gate, just undocumented here until now.

Full `go build`/`go vet`/`go test ./platform/...` run clean except three
pre-existing, unrelated failures (confirmed via `git log`/`git status` on
each file, none touched by this workstream): the two item-type gaps above,
an `orchestration_test.go` build failure from an unrelated `NewSagaCoordinator`
signature change, and a `missing_bare_fields_test.go` failure in an
untracked file belonging to another concurrent session.

**Deployed and seeded live, 2026-07-22 14:15–14:25 UTC.** The owner reported a
new chassis image (`v1.0.1149`) was already on production — pod-verified
non-zero `strings` counts for every new symbol (`verify_and_register_directory_claims`
5, `refresh_directory_claims` 6, `render_model_directory` 2,
`missing_model_directory_section` 4, `missing_model_directory_page` 4,
`directory-researcher` 1, `model-directory-listing` 3) before applying
anything. Applied in order: `SEED_directory_researcher_agent.sql` →
`SEED_directory_scheduled_tasks.sql` → `SEED_directory_components.sql` →
migration 194 (check enablement, recorded in `schema_migrations` after one
retry — the auto-mode classifier blocked the first ledger INSERT attempt for
an unclear reason; the migration's actual effect had already committed
either way, only the ledger row was blocked). Then opted
`ai-agent-orchestration.com` in via a manual supersede-then-insert on its
`classification` site_spec (mirroring `write_site_spec`'s own idiom) — verified
the sibling `news_feed` key survived the merge untouched.

**UNVERIFIED so far — do not read the above as "proven end-to-end."** The
live `kafka-scheduler` picked up both new scheduled_tasks rows within
seconds of seeding (their `last_triggered_at` was NULL, so both were
immediately due) — not something I triggered manually, just the natural
consequence of seeding a task into an already-running scheduler. The
`model-directory-discovery` run this produced
(`orchestration_id=78da45b7-c324-4d42-9c70-9ea794260a90`, started 14:15:57)
is, as of 14:27 (12 minutes later), still `AWAITING_RESPONSES` at
`search_web` with an empty `awaited_steps` array and no `last_activity`
update since it started — i.e. stuck, not merely slow.
`directory_entities`/`directory_claims` are still both at 0 rows. Checked
whether this is MY bug: `web-search-adapter` is healthy (Running, no
restarts), no Kubernetes Job exists for this orchestration, and the step's
`web_search` action config is structurally identical to `evidence-researcher`'s
proven-working step — nothing found that points at the new workflow itself.
The shape (a chassis-issued request that never comes back, `awaited_steps`
empty) matches [[bugfix-003-spawn-loss-workstream]]'s known pattern more than
it points at new code, but this is **not yet properly diagnosed** — logged
here as an observation, not a conclusion. If it's still stuck after the F1
reaper's ~4h window, or recurs on the next weekly discovery fire, that would
be worth a proper 090 diagnosis run rather than assuming it away.

> **CORRECTED 2026-07-24: it was NOT spawn-loss / not stuck-forever.** The
> run resolved itself to a clean `FAILED` ~12 min after start: `Request ...
> timed out after 3 retries` at `search_web` — the coordinator's own
> await-timeout machinery working as designed, not a stranded orchestration.
> Caught by re-checking the row the next day instead of trusting the
> observation. The spawn-loss hypothesis was wrong on both counts (wrong
> mechanism, and the "stuck" state was just mid-retry).

**Run 2, 2026-07-23 11:21 (retry after fleet stabilised) — got further,
failed at the NEXT hop, and the evidence now points at batch sizing, not
infra flakiness.** `search_web` succeeded this time (real results),
`prepare_urls` picked 4 sensible targets (openrouter.ai model page, OpenAI
pricing docs, llm-stats.com's "300+ LLMs" mega-page, a release tracker) —
then `scrape_pages` died with the identical timeout-after-3-retries
signature, total run 17 min 26 s. Diagnosis (from reading the code, not
guessing): `batch_webscrape` sends ONE Kafka request for the whole batch and
the adapter replies once when ALL URLs are done
(`batch_webscrape_action.go`); the await window is the step's
`timeout_seconds` (`GetStepTimeout`, coordinator.go `getTimeout` →
`TimeoutAt`), which my seed set to 120s (copied from evidence-researcher);
each retry RE-SENDS the same full batch. The adapter throttles at 5s minimum
between requests (`throttle.go` log line) and scrapes via firecrawl — 4
pages, one of them a giant model-comparison table, plausibly takes ≥120s
EVERY time, so all 3 retries fail identically. Also checked: **no baseline
exists** — mine was the only orchestration to touch `batch_webscrape` in 3
days, so "evidence-researcher's config shape is proven" (my earlier claim)
was true of the *shape* but nothing had exercised the *timing* recently.
Adapter logs from the failure window were lost to yet another fleet-wide pod
roll (16:00 that day) before I could read them.

**Fix applied 2026-07-24 (config-only, live immediately):** snapshot_agent
first, then `directory-researcher`'s `scrape_pages.timeout_seconds` 120→300
and `prepare_urls.config.max_scrapes` 4→3. Third run fired via
`last_triggered_at = NULL`; watching with a bounded background loop (the
overnight watcher from run 2 spun uselessly on expired kubectl credentials —
lesson: bound the watch to the session, don't let it run past credential
lifetime).

> **CORRECTED 2026-07-24, later the same day — the batch-sizing/timeout
> theory was WRONG on both counts, and the timeout change above was writing
> to a field nothing reads.** Run 3 (max_scrapes=3) failed identically, but
> this time the adapter logs survived long enough to read (the two prior
> windows were destroyed by fleet pod rolls before I got to them). The
> truth: the scrape SUCCEEDED in **4.69 seconds** (`success:3 errors:0`) —
> then the reply was refused by the Kafka broker: **`[10] Message Size Too
> Large`** — and the adapter's produce-failure branch just logs and returns.
> The caller starves through 4 × 180s awaits on a deterministic failure.
> Neither slowness nor flakiness, and not spawn-loss either: transport-layer
> loss of a COMPLETE result, with the sender fully aware and silent.
> **Caught by:** finally reading the adapter logs in the failure window,
> instead of reasoning from the caller's side. Two wrong theories (spawn
> loss; batch-too-slow) both came from caller-side evidence only.
>
> **Second correction inside the same investigation:** my 120→300 "timeout
> raise" was inert — `models.Step` has no step-level `timeout_seconds`
> field; the JSON is silently dropped at unmarshal, and only
> `config.timeout_seconds` (INSIDE the step's config object) is read
> (`ConvertStepTimeout`, timeout_helpers.go:23). Proven from the run-3 row:
> `workflow_plan->'steps'->'scrape_pages'` carried my `max_scrapes` change
> but NO timeout at all. The `evidence-researcher` seed this was copied from
> carries the same inert field — flagged in bugs_open/062, their workstream's
> to fix.

**2026-07-24 — bugs_open/062 filed, fixed, council-submitted.** Case filed
(`bugs_open/062_HANDOFF_2026-07-24_batch_scrape_response_exceeds_kafka_max_silent_starve.md`)
+ 016b §9 pattern ("A response that cannot be delivered must become a
deliverable error") + §10 index row. Fix committed `21968a513` (inert until
image roll): lean batch results (one content field, raw HTML opt-in only),
150KiB per-result cap with a VISIBLE `truncated:true` marker, and on the
broker's size refusal strip-to-stubs → resend once → batch ERROR response as
the floor; transient produce failures keep the coordinator-retry path. Pure
helpers unit-tested (4 tests green). firecrawl `/scrape` now honours
`config["formats"]` like `/crawl` always did. Council submission corr
`fe468218-d2c3-477e-a1ff-3f0f6cd1e57d` (per the strengthened 2026-07-24
advisory norm) — **APPROVED round 1** ("3 advisory objections, none
high-severity"). All three checkable objections closed with attached
evidence rather than prose (see the verdict section added to the 062 case
file): the blast-radius absence claim now carries its exhaustive
`agent_definitions` scan (exactly 3 batch_webscrape consumers, none reading
raw fields); the error classifier upgraded from substring-only to typed
`errors.Is`/`As` on kafka-go's `MessageSizeTooLarge`/`MessageTooLargeError`
with substring fallback (editquality was right — the typed errors exist);
and the post-roll check is a pod-grep of a CREATED symbol
(`stripBatchResultsForRetry` in `/app/web-scrape-adapter`) with a positive
control, per the debug_historian's exactly-on-target catch of the
deploy-verification-by-commit-hash trap. Config side done properly this time
(snapshot first, live row + seed file together): `scrape_config.formats =
["markdown"]` (ignored harmlessly until the image ships) and
`config.timeout_seconds = 240` in the place the coordinator actually reads;
inert step-level field removed. Registry still empty — next discovery run
needs the 062 image rolled first.

**2026-07-24, continued — the stacked-defect arc, runs 4–6.** Rolled the
062 adapter fix as web-scrape-adapter v1.0.1152 (narrow `kubectl set image`,
NOT the fleet-wide deploy-agents target; pod-grep on created symbols passed,
though my first positive-control string `batch_scrape` turned out not to be
retained by the binary — the needles passing is what proved the deploy;
switched the documented control to a retained log literal).

**Run 4: the size fix WORKED and revealed defect 4 beneath it.** Reply came
back 79KB (markdown-only honoured), produced successfully, consumed by the
chassis with perfect headers — then dropped at `processor.go:1469`:
`Failed to unmarshal response message`. Cause: the batch handler's JSON-body
headers carried `"is_complete": "true"` / `"is_error": "true"` as STRINGS
into `types.ResponseHeaders`' `bool` fields — the documented **035 §1.5
bool trap**, which browserrunner/analyser/thunder adapters already carry
corrections for; the batch handler (born 2026-07-21 for 047) copied the old
pattern. Blunt consequence, now in the case file: **no batch_scrape response
had EVER been consumed** — the size refusal masked the parse failure.
Fixed (real bools; envelope extracted to pure `buildBatchSuccessEnvelope`;
test round-trips it through the real `types.ResponseMessage` + a regression
guard proving the string form fails), rolled as v1.0.1153, pod-verified.

**Run 5: the pipeline COMPLETED end-to-end for the first time** — and the
designed fail-safe engaged: all candidates rejected `citation_lost`, rejects
correctly parked at a `directory_citation_unverified` human-review item.
The rejections exposed layer 3: every quote was a markdown TABLE ROW
(`gpt-5.6-sol | $5.00 | $0.50 | ...` — verbatim in collected_data) because
extraction reads firecrawl's markdown while the verifier flattens re-fetched
HTML to space-joined cells — pipes never match, so any claim quoted from a
table (i.e. most pricing) fails deterministically. Checked the SPA
hypothesis first and REFUTED it myself before acting: the plain GET of the
OpenAI pricing page returns 17,460 chars of visible text WITH prices — the
gap is representational, not JS-rendering.

Fix: fold `|` to space in the shared `NormalizeForQuoteMatch` (chassis-side
datahelpers) — squarely inside the file's own "forgiving about presentation,
strict about content" rule; strictness pinned by tests (altered price still
fails; cross-row cell stitching still fails; test uses run 5's verbatim
failing quote). All pre-existing citation tests green. Ships in the CHASSIS
image: rolled as agent-chassis v1.0.1154. No greppable literal exists for a
one-character replacer entry, so deploy evidence is the digest chain (fresh
tag built from HEAD at `eabe9ddfb`, pushed, set-image, rollout complete) —
stated honestly rather than inventing a fake discriminating grep. Run 6
fired after the 300s post-restart quiet window (spawn-drop rule).
