# PLAN — fleet-wide AI model directory (+ adoption-tracker follow-on)

Design, phasing, decisions and their reasons. Corrections to the brief live
here, marked as corrections — never silently edited away.

## Origin

Session renamed "ai-agent-orchestration" (2026-07-22). Owner's brief: a
prominently-placed, continually-updated AI model directory (open + closed
source models — what they do, cited cost, owner, where to find/use them,
links out, later finetuning.uk wrapper links + video links) on
`ai-agent-orchestration.com` (site_id `2a8ebf9c-20a2-4c39-b191-840b012371da`).
A second section — company AI-agent adoption tracker (ROI claims, rollout
breadth, cited examples, protocol adoption e.g. MCP) — follows on the same
schema.

> **Correction (mid-brief, same session):** the original brief read as a
> single-site build. The owner then said: *"I would like this pipeline to be
> available to any new domain that requests it. Most of the infrastructure is
> already there. I'd like to be able to create this sort of thing
> automatically from the site-specs."* This is not a scope tweak, it's a
> different shape of deliverable — a fleet capability opted into via
> `site_specs`, not a bespoke page. Everything below reflects the corrected
> scope.

## Decisions (owner, 2026-07-22)

- **D1 — Pipeline before publish.** Build the automated research/citation
  pipeline first; no hand-curated launch. *Reason (owner):* explicit choice
  over "hand-curated MVP first" and "schema now, content later."
- **D2 — Model directory ships before the adoption tracker**, on a schema
  designed so the tracker is a same-shape addition (new `kind`, new `field`
  values), not a rebuild.
- **D3 — Fleet capability, not a single site.** Opt-in via `site_specs`
  (mirrors the existing news-feed opt-in exactly), auto-discovered and
  auto-published by scheduled triggers — the same architecture class as
  `content-feed-trigger`/`content-feed-refresh`.

## Architecture (see full detail in the approved plan file,
`/home/ant/.claude/plans/valiant-roaming-dawn.md`, copied here for
permanence since plan files are session-scoped)

Two new global tables, `directory_entities` + `directory_claims`, hold cited
facts (model now; company/protocol later — zero schema change for Phase E).
Cost/spec facts are **claims** (cited, re-verified, can flip to
`citation_lost`), never a jsonb convenience cache on the entity — that
distinction is deliberate: an uncited "attributes" blob is exactly the kind of
fact that goes stale silently, which the whole claims machinery exists to
catch.

A researcher agent (`directory-researcher`, a cousin of `evidence-researcher`
— NOT a reuse of it, since its persistence target `site_specs.evidence_base`
is per-site and wrong for a cross-site registry) discovers new entities and
extracts atomic, verbatim-quoted claims. A deterministic verifier re-fetches
each citation URL and confirms the quote is still present, reusing
`datahelpers.QuoteFoundInText` / the `evidence_citations.go` verification
shape unchanged.

A publish action (cousin of `RenderNewsSectionAction`) renders the registry
to `data/model-directory.json` per opted-in site, commits via the existing
`git_commit` action, queues a scoped `page_rerender`; a
`query.model_directory` resolver feeds the same data server-side — the same
dual-layer (baked HTML + client JSON fetch) the news component already uses.

Page/section creation is NOT a new mechanism — it plugs into the real,
already-automated pipeline: `discovery_checks` → `content-gap-planner` →
`apply_gap_plan_action.go` → `page-build-handler` (verified live as the path
that creates `/news.html` today).

Opt-in reuses the identical `site_specs` shape news already uses:
`aspect='classification'`, sibling key `content_features.model_directory`
next to the existing `content_features.news_feed`, written via the existing
generic `write_site_spec` action — not a new aspect, not a bespoke
transaction.

## Phasing

- **A — Schema.** `directory_entities` + `directory_claims` tables + indexes.
  Migration `191_model_directory_schema.sql`. No seed content (D1).
- **B — Researcher agent + scheduled refresh.** New Go action file
  `directory_claims.go` (upsert/verify/refresh), `agent_definitions` row
  `directory-researcher`, two `scheduled_tasks` rows (discovery weekly,
  freshness daily). Requires image roll.
- **C — Component + resolver + publish action.** `render_model_directory_action.go`,
  `content_components` rows `model-directory`/`model-directory-listing`,
  `queryresolve/model_directory_items.go`. Requires image roll. Must also
  name the components in the build-site-planner/site-architect prompt or
  they're inert (bit the brochure-library workstream before).
- **D — Opt-in + auto page creation + pilot.** `discovery_checks/check_model_directory.go`,
  append checks to `completeness-discovery-agent`, add `"model-directory"` to
  `structuralPageTypes` (`page_growth_budget.go:37`), `model-directory-trigger`
  agent + `scheduled_tasks` row `model-directory-publish`. Pilot on aao only
  before the fleet-wide loop would pick it up.

  > **CORRECTED 2026-07-24:** the publish-trigger half of this phase (the
  > `model-directory-trigger`/`model-directory-publish` items named in the
  > sentence above) was planned here but **never implemented in the Phase D
  > work** — the checks, growth budget and enablement migration all shipped
  > on 2026-07-22 while the publish leg silently fell out of the work list.
  > Nothing caught it: the omission produced no error anywhere, because a
  > trigger that doesn't exist looks identical to a trigger that is idling.
  > Noticed only on 2026-07-24, while writing the milestone summary — i.e.
  > by re-reading THIS plan against what was actually seeded, which is the
  > check that should have run at the end of Phase D itself. Closed the same
  > day: `SEED_directory_publish_trigger.sql` (publisher agent + trigger
  > agent + 6h task, self-gating), applied live after a dry-run caught a
  > second miss (`check_ad_category` rejects invented category values; use
  > content-feed-trigger's `orchestrator`/`coordinator`).
- **E — Later.** Adoption tracker: `kind='company'`/`kind='protocol'` rows,
  new `field` values, same tables.

## Open forks (owner)

1. `model-directory-listing` as its own page vs. a section on an existing
   hub page — plan defaults to a dedicated page (doubles growth-budget/nav
   footprint per site).
2. Re-verification cadence for pricing claims specifically (short
   `staleness_days` for prices vs. long for structural facts) — needs
   numbers before Phase B seeds them fleet-wide.
3. Pilot scope — aao only, or also `finetuning.uk` immediately given the
   brief's mention of future wrapper links there.
