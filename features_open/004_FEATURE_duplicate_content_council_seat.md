# 004 FEATURE — duplicate-content council seat

**Raised:** 2026-07-20, by the owner ("we might need a duplicate content member
on the council"), during news-feed-pooling design.
**Status:** specified, not seated. Becomes worth seating the moment pooled
selection exists — which is also the moment it has something to veto.
**Related:** `features_open/002_RISK_portfolio_duplicate_content.md` (the risk it
polices), `docs024_key_docs_latest/news_feed_pooling/PLAN_*` Decisions 4/10/11.

## Why a seat, when 002 already specifies a metric

The metric (pairwise similarity across a pool's rendered output) catches the
problem **after** it renders. A council seat catches it **at review time** — when
a thread proposes a change to selection, ranking, rendering or package
generation, before it ships. They are complementary: the check is the smoke
alarm, the seat is the building inspector.

The deeper reason: duplicate content is a **cross-item property**. This repo's
review machinery — like its verification machinery — naturally evaluates the
thing in front of it, and every incident class this platform has recorded
(one-branch fixes reading as done, per-site checks passing while the set is
broken) comes from that blindness. A seat whose *only* lens is "what does this
change do to the portfolio as a set?" is the structural fix at review altitude.

## What the seat reviews (footprint)

Fires when the edited paths touch:

- `platform/orchestration/actions/feed_*` — selection, triage, normalise
- `platform/orchestration/actions/render_news_section_action.go`,
  `render_rss_feed_action.go` — rendered feed surface
- anything writing the `audience` aspect or pool membership — profile convergence
  upstream *is* selection convergence downstream
- packaged-feature substrate/angle generation, when built (`features_open/001`)
- pool-site seeds and pool membership changes

## Review posture

The seat asks one question in four forms:

1. **Selection**: does this change make two same-pool sites *more* likely to pick
   the same top-K? (e.g. weakening recency jitter, sharpening a shared profile,
   removing a diversity cap.)
2. **Rendering**: does this change increase the shared byte-surface? (shared
   summaries replacing per-site ones; a template change that flattens per-site
   variation.)
3. **Profile**: does this change make audience profiles converge? (a pool default
   edited to say more, forks folded back into the base, a seed that copies one
   profile to N sites.)
4. **Measurement**: does the change alter what 002's check would see, and has the
   proposer re-baselined? A change that *improves* real similarity but breaks the
   metric's comparability should say so.

Verdict style follows the guardian pattern: REVISE with the specific pair-level
concern named; REJECTED only when the change structurally guarantees convergence
(e.g. deleting the per-site ranking layer, pointing multiple sites at one
pre-rendered block).

## Mechanism (standard, per CLAUDE.md)

Seat `fix-proposer` as usual, then run the roster mirror — never hand-patch the
gate:

```
python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py
```

(dry run first; `--apply` writes with a snapshot). Cost is relevance-gated: the
seat fires only when edited paths match its footprint, so it is free for every
thread not touching feeds.

## Inputs the seat needs to exist first

- The 002 similarity check running, so the seat can cite a baseline rather than
  intuit ("this change moves worst-decile pair similarity from X to Y" beats
  "this feels riskier").
- Pool membership readable from the DB (pool synthetic sites — PLAN Decision 8),
  so the seat can enumerate the affected set.

## Not yet decided

- Whether the seat is advisory-with-veto (guardian-class) or advisory-only. Given
  002's "not cleanly revertable" property — footprint damage attaches to domains,
  not deployments — there is a case for guardian-class on the rendering surface
  specifically.
- Whether the same seat also reviews *packaged-feature* substrate updates for
  angle staleness (the blast-radius concern in 001), or whether that is a
  separate freshness seat. Leaning same seat: both are "what does this shared
  change do to the derived set".
