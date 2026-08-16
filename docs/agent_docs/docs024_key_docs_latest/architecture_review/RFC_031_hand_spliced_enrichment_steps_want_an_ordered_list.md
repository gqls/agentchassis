# RFC_031 — deterministic content_features enrichers are hand-spliced per agent; the third recommender should be a shared ordered list

**Status: FILED 2026-08-15 (trigger record, not yet ripe).** Filed by the
portfolio_positioning lane at the council's direction: the architecture seat flagged the
pattern in 432's round 1 (corr `47785bb5`, medium, not gating) and round 2's approval asked
that the trigger "not become the final resting place" in lane NOTES — *"an actual RFC
register entry should exist so a landmine-style stale-status trap doesn't swallow it."*
This file is that record. Nothing is to be built yet: **the trigger is the THIRD
deterministic starter-kind/feature recommender**, and only two exist.

## The pattern

A deterministic, no-LLM action reads the current classification spec, matches the site's
vertical against a closed map, and deep-merges a `content_features.<key>` flag the planner
then reads (opt-in, no-match-means-no-write). Each such action is wired into its consumers
by HAND-EDITING the workflow JSON: insert a step, re-point the predecessor's `next_step`
AND `config.error_step`, preserve the continue-on-failure posture bug 291 established.
Every splice is a migration with drift guards on the live edges (429/432 house style), and
every rollback must pin the same edges in reverse (432 round-2 advisory).

## The population (measured live, 2026-08-15 22:30Z)

| action | consumer | spliced between | by |
|---|---|---|---|
| `evaluate_news_feed` | improvement-loop | (upstream) → `enrich_directory_features`* | pre-existing (news lane) |
| `evaluate_directory_features` | improvement-loop | `enrich_news_feed` → `load_audit_state` | migration 432 |
| `evaluate_directory_features` | domain-research-classifier | `write_classification_spec` → `write_content_direction_spec` | migration 432 |

\* after 432, news's success/error edges point at the directory step — each new splice
re-points the PREVIOUS enricher's edges, which is exactly the accumulating fragility.

**Drift evidence already visible:** `evaluate_news_feed` is NOT spliced into
domain-research-classifier, so a greenfield build carries its directory flags at plan time
(432's stated purpose) but not its news flag — an asymmetry nobody decided, produced purely
by the hand-splice approach having no single place where "the enrichers a classification
pass runs" is stated.

## What the abstraction would be

A generic ordered enrichment-step list on the consuming agents: one step (or one action)
that runs a configured list of enrichment actions in order, each with the 291
continue-on-failure posture, so adding the Nth recommender is appending a name to a list —
not splicing edges, not re-pointing the (N−1)th enricher's error path, not a per-agent
migration with per-edge drift guards. Whether it is a chassis-level `run_enrichments`
action taking `["evaluate_news_feed", "evaluate_directory_features", …]` or a workflow
convention is the design question for the round that builds it.

## The tripwire, stated as a rule

**Whoever writes the THIRD deterministic content_features recommender must propose this
abstraction (or explicitly argue it down in this file) instead of hand-copying the splice a
third time.** The 432 file's own header notes are the worked example of what each hand copy
costs: edge pins forward, edge pins in rollback, a council round arguing about both, and a
growing chain where each enricher's error path names the next enricher rather than the
join point. Relations: RFC_022 (accumulated-surface counting — ten individually-inert
opt-in keys are a shared action nobody understands; the same accumulation logic applies to
splices), RFC_030 (same shape of ruling for work-item routers: the modular unit is the
per-type map, the shared engine runs them), NEWS-001 / DIR-001 (the two live pattern
instances), `portfolio_positioning/NOTES` 2026-08-15 closing entry (the round-1 flag this
file makes durable).
