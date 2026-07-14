
<!-- SOURCE: U03_idea_uk_section_data.md -->
### SQL needle-gate surgery pattern (guarded, idempotent, reversible DB edits)
- **category:** NEW:sql-change-management
- **status-signal:** deployed
- **status-evidence:** Practised in every w1–w9/slice4a file; codified in notes (Sv): "the needle-gate rule REFINED in 016b_7 (count expectations mechanically from the dump; mismatch = drift OR bad expectation)"; slice4b: "needle discipline applies to docs too."
- **what:** The unit's dominant change-management method for production data edits (templates, layouts, prompts, docs): dump the current bytes; derive byte-exact needles per element (multi-line E'…\n…' needles to disambiguate repeated strings; `position()` where needles contain literal `%`); run a read-only GATE that asserts each needle's presence and mechanically-derived counts; apply nested exact-string `replace()` (or `\1`-anchored regexp_replace) UPDATEs guarded on a pre-state marker so re-runs are 0-row no-ops; RETURNING booleans as immediate post-conditions; separate verify and inverse-rollback files plus a full .bak dump before every mutation. The 019 migrations extend the pattern to agent prompts: sentinel-guarded idempotency, abort-if-fragments-moved, paired down-migration. Sibling rules: `\set ON_ERROR_STOP on` for dependent mutation files; run SQL as files, never pasted into interactive psql.
- **sources:** w2_01_footer_fix.sql; w3b_01_hero_conversion.sql; slice4a_creator_prompt.sql; 019_pcw_prompt_item_fields.sql; running_notes_scheme_to_components(55).md#Sp #Sv #Ss
- **relations:** SQL pitfall class; agent re-seed risk; documentation-system (needle discipline on docs).
- **verify-later:** 016b guide's needle-gate entries; whether the pattern is written up as a standing convention doc.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### SQL needle-gate surgery pattern (guarded, idempotent, reversible DB edits)
- **category:** NEW:sql-change-management
- **status-signal:** deployed
- **status-evidence:** Practised in every w1–w9/slice4a file; codified in notes (Sv): "the needle-gate rule REFINED in 016b_7 (count expectations mechanically from the dump; mismatch = drift OR bad expectation)"; slice4b: "needle discipline applies to docs too."
- **what:** The unit's dominant change-management method for production data edits (templates, layouts, prompts, docs): dump the current bytes; derive byte-exact needles per element (multi-line E'…\n…' needles to disambiguate repeated strings; `position()` where needles contain literal `%`); run a read-only GATE that asserts each needle's presence and mechanically-derived counts; apply nested exact-string `replace()` (or `\1`-anchored regexp_replace) UPDATEs guarded on a pre-state marker so re-runs are 0-row no-ops; RETURNING booleans as immediate post-conditions; separate verify and inverse-rollback files plus a full .bak dump before every mutation. The 019 migrations extend the pattern to agent prompts: sentinel-guarded idempotency, abort-if-fragments-moved, paired down-migration. Sibling rules: `\set ON_ERROR_STOP on` for dependent mutation files; run SQL as files, never pasted into interactive psql.
- **sources:** w2_01_footer_fix.sql; w3b_01_hero_conversion.sql; slice4a_creator_prompt.sql; 019_pcw_prompt_item_fields.sql; running_notes_scheme_to_components(55).md#Sp #Sv #Ss
- **relations:** SQL pitfall class; agent re-seed risk; documentation-system (needle discipline on docs).
- **verify-later:** 016b guide's needle-gate entries; whether the pattern is written up as a standing convention doc.
