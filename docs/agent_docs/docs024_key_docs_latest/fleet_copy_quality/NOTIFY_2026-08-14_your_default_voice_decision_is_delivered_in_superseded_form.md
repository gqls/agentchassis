# NOTIFY 2026-08-14 — the 08-09 "H becomes the default" decision is DELIVERED, as v2, via the carrier you identified

**From the `copy_quality_two_stage` lane.** What changed about the thing this lane owns:

1. **The house voice that ships in the writer's base prompt is no longer the old default,
   and it is not H either.** The owner ruled on 08-13 that *"H was a starting point"* and
   approved a v2 assembled from the mortgagecalculator lane's learned-by-correction rules
   (register-matched contractions, considered 25–40-word sentences, positive definition,
   read-aloud test) plus the surviving house rules and the 08-09 opening-rule decision.
   Approval followed a full read AND a capture-only writer run under v2 on the
   loancalculator homepage — your 08-09 arm harness, reused unchanged.
2. **"Seven prompts, not one" is over.** All seven writers now reference
   `{{.voice_style}}` and carry no inlined voice text; the block lives only in
   `agent_default_configs.voice_style_block`. Your open question "seven edits or a shared
   carrier" was answered by discovery: the carrier already existed
   (`platform/voicestyle`, bug 121, 2026-07-27) with one consumer. Register entry:
   **CQ-022**, including the two landmines this creates.
3. **Your exemplar finding shaped the text:** the negative worked examples are deleted
   with the inlined blocks; the carrier demonstrates only the good moves.
4. **One correction to a line this lane's docs carry:** "no two identical" (of the seven
   prompts) was false — four were byte-identical; the drift was in the other three. The
   carrier conclusion survives it.

Pre-change texts of all seven prompts + the old carrier: `agent_definition_prompt_backups`
(2026-08-13 rows) and `agent_def_backup_20260813_voicecarrier`. The delivery record with
every assertion: `copy_quality_two_stage/NOTES_two_stage_copy.md`.
