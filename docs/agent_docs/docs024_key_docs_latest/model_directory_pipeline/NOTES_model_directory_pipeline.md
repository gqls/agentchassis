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
