# HANDOFF — 2026-08-09 (midday). PHASE 1 IS LIVE. Phase 2 (the Go deletion) is next. COLD-START HERE.

Supersedes `HANDOFF_2026-08-09_209_verified_latent_fix_not_taken.md` (whose
"fix not taken" is no longer true — the owner approved into-line the same day).
`NOTES_209_purpose_keyed_source.md` = evidence + missteps;
`RUNBOOK_209_purpose_keyed_source.md` = commands (§7 scoped migration apply,
§8 post-roll config re-verify are the two you'll need first);
`PLAN_2026-08-08_209_purpose_keyed_source.md` = decisions, incl. the two owner
rulings. Contribute findings INTO `bugs_open/209` and `bugs_open/231` — shared
accounts.

## State

| | |
|---|---|
| direction | **INTO-LINE, owner-approved 2026-08-09** ("carry on … Phase 1 first"), superseding the earlier divergence plan |
| Phase 1 | **DONE, LIVE, ROW-VERIFIED** — migration `348_pageflow_swo_deploy_steps_resolve_by_identity.sql`, applied+recorded 09:41:53; ROLLBACK sidecar exists |
| what 348 did | all 4 deploy steps (pageflow-builder + site-work-orchestrator × hero/logo) now resolve purpose/s3_uri/asset_id/domain by **Strategy-0 dotted paths** from their own store step's output (`{p}_stored.*`, `site_record.domain`); `uri_field` + static purpose REMOVED; `input_fields: ["purpose","domain","asset_id"]` added — **s3_uri deliberately excluded** (store-failure corner must skip, never aggressive-search a sibling's URI) |
| closes | bugs_open/231 shadow (logo-as-hero) for these workflows · the 86% recursive-asset_id hazard · their purpose-keyed source dependence |
| tests | 8/8 in `platform/orchestration/actions/deploy_image_asset_purpose_source_test.go` — incl. exact-live-shape determinism (100/100) and the store-failure corner (sibling never leaks) |
| commits | session 1: `ae990ee82`; session 2: `6287b198a`, `ae86ef660`, `2eaee2387`, + this session's Phase-1 commit |
| bugs | `209` OPEN (Phase 2 + behavioural proof owed) · `231` OPEN (fleet census + candidates 2/3 + behavioural proof) · contributions live in `223` (two consumers of var-blindness) and `idea_uk_vm_site/CONTRIB_…` (064 recurrence) |

## Next work, in order

1. **Behavioural proof (owed on BOTH bugs):** dispatch ONE sacrificial-domain
   run through each legacy workflow (owner ruling 2026-08-04: every site through
   the framework — these workflows ARE framework; still, pick a throwaway domain,
   not a client's). Assert: `hero.*` AND `logo.*` committed with **different
   bytes**; `content_data.logo_url` serves 200. This satisfies 209's own
   "verify on the workflow that really does it" bar and proves 348 end-to-end.
   Note the pair has NO dispatcher among live definitions — find the initial
   message/topic to fire them (082_submit or a direct kafka dispatch; NOT
   verified this session — check `scripts/initial_messages/` first).
2. **Phase 2 (Go, council):** delete `findStorageURI` + its call site in
   `deploy_image_asset_action.go` (~90 lines; priorities 1–7 all dead for live
   callers now — Priority 1's `uri_field` no longer exists in any config, but
   RE-VERIFY that census on pickup, it is config and drifts). Resolution becomes:
   `s3_uri` input → `asset_id` DB row → loud skip. Update the characterisation
   tests DELIBERATELY (the purpose-keyed ones flip). Council-submit before or
   alongside the commit (`Council-Submitted:` trailer). **Ordering: image must
   not roll before 348 is confirmed at the rows** — it is, but re-check on
   pickup (RUNBOOK §8; rolls re-stamp `updated_at` without changing content —
   the 341 `gate_next_item` control proved content survives).
3. **231's remainder:** the fleet-class census (which OTHER live configs carry a
   static value for a spec-defaulted field). The 090 (run `e952039b`) returned
   UNVERIFIABLE/scope-not-narrowing — it CONFIRMED the helper mechanism
   independently but could not fetch the spec **declaration** (package-level
   `var` — `code_symbols` has no var kind; bug 223, now with a second consumer).
   Do the census by hand: enumerate `Defaults: map` specs (grep under
   `platform/orchestration/actions/`), then per defaulted field query live
   `agent_definitions` step configs for a static non-dotted value ≠ default.
   Also consider 231's candidate 3 (CheckConfig flags shadowed statics — cheap,
   catches future authors).
4. **Phase 3 (optional, unowned):** retire the `{purpose}_uri` writers — first
   classify the 6-per-definition references to `hero_uri|logo_uri|hero_result|logo_result`
   beyond the deploy steps. Not needed to close 209.

## Traps already paid for (do not re-derive)

- `who-owns.py` on 209/223/064 returns citation-based false positives — re-check
  via live transcripts (RUNBOOK §9).
- `--apply` takes EVERY pending file; scope with `MIGRATIONS_DIR` (RUNBOOK §7).
  Pending set held other lanes' 342/345/346 this morning.
- `updated_at` on agent_definitions is re-stamped fleet-wide at every deploy,
  content-preserving. Compare CONTENT (RUNBOOK §8).
- `orchestration_states` keeps completed runs ~24h; `llm_call_log` is blind to
  these agent types (positive control failed). Neither can prove the pair's
  run history.
- The package's ONE failing test (`TestValidDocSubjectTypes_Lockstep…`) is the
  064-shape recurrence owned by `idea_uk_vm_site` (told via CONTRIB file) — not
  this lane's.
- Full-row reads, not column-selected ones, when verifying config (the
  `domain_field`/`output_mapping` keys were invisible to two-column queries).

## Cold-start checks

1. `git log --oneline <this-handoff's-commit>..HEAD -- platform/orchestration/actions/deploy_image_asset_action.go platform/orchestration/datahelpers/ bugs_open/209_* bugs_open/231_*` — empty = ground unmoved.
2. RUNBOOK §8 content re-verify of the four steps (348 shape intact?).
3. `go test ./platform/orchestration/actions/ -run 'TestFindStorageURI_|TestExtractActionInputs_|TestDeployImageAsset_|TestLegacyLogoStep_|TestPurposeFieldBridge_|TestStrategy0DottedPaths_|TestMigration348Shape_'` — 8 pass expected.
4. Ownership sweep (RUNBOOK §9) — another thread may have taken Phase 2.
