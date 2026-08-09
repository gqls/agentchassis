# HANDOFF — 2026-08-09. 209 is VERIFIED LATENT; the fix is deliberately NOT taken. ~~Read this first on a cold start.~~

> **SUPERSEDED same day by `HANDOFF_2026-08-09b_phase1_live_phase2_next.md`** —
> the owner approved the into-line fix and Phase 1 (migration 348) is applied
> and live. Start there; this file remains as the record of the verification
> stage and the divergence proposal it replaced.

`NOTES_209_purpose_keyed_source.md` has the evidence and every misstep;
`RUNBOOK_209_purpose_keyed_source.md` has the commands with their gotchas;
`PLAN_2026-08-08_209_purpose_keyed_source.md` has the decisions and the corrected
fix ranking. The bug file itself (`bugs_open/209_HANDOFF_2026-08-06_...md`)
carries the full verification block + 08-09 addendum — it is the shared account;
contribute there.

## State, in one table

| | |
|---|---|
| bug | `bugs_open/209` — `findStorageURI` resolves an image source by PURPOSE from `collected_data`, last-write-wins, consulted before the `asset_id` path |
| verdict | **LATENT, not live** — real in code at HEAD, unreachable in live config (measured 08-08, re-verified by content on v1.0.1270 after the 08-09 08:49Z roll re-applied seeds) |
| fix | **deliberately not written** — no live exposure, and the bug file's top-ranked fix is measured harmful (below) |
| shipped | `ae990ee82` — four characterisation tests (`platform/orchestration/actions/deploy_image_asset_purpose_source_test.go`, all passing) + the lane's standing docs + a LANDMINES entry |
| landmine | `LANDMINES.md#a-step-with-no-inputfields-...` — verifier-confirmed **STILL_VALID** (its footprints are all `func`, which is why it could be) |
| other commits | `78ca0c348` (ratchet line), `202abc188` (221 loose end closed + third failure mode into `bugs_open/223` + 2 WRONG_CALLS entries) |
| ownership | unowned before this lane; `who-owns.py` pointing at `bugfix_221_...` is a **false positive** (its handoff cites 209). Now it will point here — this lane is 2 sessions, verification only |

## The three findings a fixing thread must hold

1. **No live workflow can currently trigger it.** Three definitions carry the
   action (`asset-deployer`, `pageflow-builder`, `site-work-orchestrator`);
   `image-build-handler` is the near-miss. The hero+logo pair use different
   purposes; image-build-handler's two `purpose:"hero"` stores are on mutually
   exclusive branches and it deploys by identity (`s3_uri: asset_stored.image_uri`);
   asset-deployer's `collected_data` has no `{purpose}_uri` key at all and the
   reader is a strict path walk. Re-check this census before fixing — it is
   config, live-mutable, and the roll re-seeds it (compare CONTENT, `updated_at`
   is bumped even when identical).
2. **The bug file's fix candidate 1 ("delete Priority 2, rely on asset_id") is
   measured harmful — do not take it as written.** The legacy steps have no
   `input_fields`, so `asset_id` resolves by randomised recursive search: the
   logo step got the HERO's asset_id **344/400 (86%)**. Precisely: candidate 1
   leaves their primary route (the `uri_field`→`s3_uri` bridge) intact but swaps
   their **fallback** from correct-by-purpose to 86%-wrong-by-recursive-search —
   conditional exposure, fires exactly when the primary hiccups, which is what a
   fallback is for. Also: priorities **3–7** are purpose-keyed too, so deleting
   Priority 2 alone does not even achieve the candidate's stated goal.
3. **Recommended ranking (in the bug file, replacing the original):**
   (1) additive identity key — write `asset_uris.<asset_key>` alongside
   `{purpose}_uri`, reader prefers `asset_key` when supplied. **A platform seam**:
   concept-register entry in the same commit, council submission alongside,
   consumers told (2026-07-28/29 rulings). (2) give the legacy steps explicit
   `input_fields` — config-only, live immediately, kills the recursive hazard.
   (3) only then retire purpose-keyed priorities.

## Open questions (also at the end of PLAN)

- **Are `pageflow-builder` / `site-work-orchestrator` dead or dormant?** No live
  definition spawns/calls them; they did not run on 08-08. But
  `orchestration_states` retains completed runs ~24h ONLY (13 rows >24h, 0 >7d —
  do NOT read retention off `min(created_at)`, see WRONG_CALLS 08-08), and
  `llm_call_log` is blind to these agent types (positive control failed). An
  **owner call** to retire them would unlock the clean fix. Until then they are
  live-supported consumers.
- **Wider hazard:** which other actions have a no-`input_fields` step whose spec
  field name occurs twice in `collected_data`? Same 86% shape, different action.
  Nobody has swept for it.

## Also closed in this lane (not 209)

- 221's outstanding landmine-verifier re-fire: done, verdict `NEEDS_HUMAN_REVIEW`
  is a **false alarm** — `code_symbols` has NO `var` kind (func/method/struct/
  alias/interface only, 5,755 total), so `metaCommentaryPatterns` /
  `placeholderPatterns` (both exist at HEAD, `validate_page_content.go:105`/`:1229`)
  can never resolve. Filed as a **third failure mode** into `bugs_open/223`
  (owned elsewhere — contributed, not taken). 221's entry NOT downgraded.

## Cold-start checks before continuing this lane

1. `git log --oneline <last-commit-here>..HEAD -- platform/orchestration/actions/deploy_image_asset_action.go platform/orchestration/datahelpers/ bugs_open/209_*` — empty means the ground has not moved.
2. Re-run the harness: RUNBOOK §6. The instability split varies (344→348/400 across two days); plurality is the signal, not the ratio.
3. Re-run the config census by CONTENT: RUNBOOK §1–2.
4. `./scripts/who-owns.py 209` + the transcript sweep in RUNBOOK §7 — another
   thread may have picked up the fix; if so, contribute the findings table above
   into their work, do not compete.
5. The package's one failing test (`TestValidDocSubjectTypes_LockstepWithMigrationCheck`)
   is pre-existing (`bugs_open/064` lockstep, migration 340 vs
   `doc_subjects_common.go:63`) — not this lane's, don't chase it here.
