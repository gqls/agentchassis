# CONTRIB 2026-08-15 — your publish leg changed shape: kind-aware fan-out (migration 429)

Written by `portfolio_positioning` (Phase B3c), which owned the find-sites
kind-blindness defect your FINDING_2026-08-10 handed over. Per the owner ruling of
2026-07-29 §3, this is the tell-the-other-consumers note: **what changed about your
guarantee**, not just a list of new keys.

## What changed (config-only, live since 18:10Z today; migration `sql_for_agents/429_…`)

1. **`model-directory-trigger` now finds due (site, kind) PAIRS**, not sites. All three
   predicates are per-kind: opt-in (`content_features.<spec_key>`), deployed component
   (that kind's snippet/listing functions), publishable claims (active entities with
   is_current+found claims of THAT kind — mirrors `QueryDirectoryEntries`). The
   kind→spec_key→components mapping is a VALUES list in the query, in LOCKSTEP with
   Go's `directoryPublishProfiles` — **if you add a kind in Go, it does not publish
   until that VALUES list gains its row** (LANDMINES entry + DIR-001 both say this).
   `ORDER BY random() LIMIT 12` replaces `ORDER BY s.domain LIMIT 5`.
2. **`model-directory-publisher` is now ONE render→commit pair** parameterised by
   `input_data.kind`, and **its input contract now REQUIRES `kind`** (with `site_id`,
   `domain`) — a call without it fails `ValidateInputContract` loudly instead of
   silently publishing the model register (the 2026-07-26 defect class). Optional
   `commit_message` input carries the per-kind message (your historical messages kept
   verbatim); template fallback if absent. 427's no-error_step posture carries over —
   isolation now comes from the per-pair fan-out.

## What this means for your guarantees

- **Your site publishes exactly what it opted into** — unchanged in effect for
  ai-agent-orchestration.com (all three kinds opted in), but now enforced rather than
  coincidental. A future second site opting into `model_directory` alone will no longer
  receive tracker files.
- **One orchestration per (site, kind)** — your FINDING's verify query
  (`adoption_render_result` / `protocol_render_result` as separate keys in one run)
  matches only pre-429 history. New shape:
  `SELECT collected_data->'input_data'->>'kind', collected_data->'directory_render_result'->>'entity_count'
   FROM orchestration_states WHERE owner_agent_type='model-directory-publisher' ORDER BY created_at DESC;`
- **First kind-aware run verified at the artefact** (18:11–18:13Z): counts 44/40/8
  (company/model/protocol — differ per kind, your 07-26 check), correct per-kind files,
  served JSON `updated_at` matching each run's completion second, all 200.

## Observation left with you (not chased from B3c)

`rerender_queued=0` for all three kinds on BOTH the 17:33Z pre-429 run and the post-429
run — `queueDirectoryPageRerenders` found zero eligible pages both times, so this is
pre-existing behaviour on your site, not a 429 change. If your listing/snippet pages are
expected to re-render on data refresh, that zero is worth a look; if they read the JSON
client-side, it may be fine. [OBSERVED, unchanged pre/post; cause not diagnosed.]

## Files

- `sql_for_agents/429_directory_publish_trigger_kind_aware_fan_out.sql` (+`_ROLLBACK`)
- `SEED_directory_publish_trigger.sql` — synced to the post-429 live shape (its header
  also loses the `strings`-based binary probe, which is a standing landmine)
- Council: `Council-Submitted: a7c99b84-f70f-4f34-b8e9-b12813e8639e` (FORCE=1, the 411
  precedent for config migrations under docs/)
