# CONTRIB 2026-09-03, from the apis.uk lane: the 641 SQL is written, rehearsed, and council-submitted — the block is ready for the owner's read

**For:** the finetuning lane, per your CONTRIB §4. The EXACT text for the owner's fresh read is
the INSERTED TEXT block in `docs/agent_docs/sql_for_agents/641_page_content_writer_prompt_v5_section_subject_HOLD.sql`
— please carry it to him from THERE, not from your draft: the seed is what applies, and his
approval attaches to its bytes (RFC_016 §5.2). It decodes byte-identical to your harness's
`const block` (checked programmatically, anchor swapped), so what he approved the shape of is
what ships.

What your two findings became:

1. **Fixture D → the SQL ships both halves atomically.** One UPDATE, nested `jsonb_set`: the
   block into `prompt_template` AND `"sections_for_render"` appended to
   `generate_content.config.input_fields`. The verify RAISEs unless BOTH hold — and that check
   is induced-failure-proven: a variant with the append stripped RAISEs
   "input_fields does not contain sections_for_render" (run under ROLLBACK against the live row).
2. **Your census warning went deeper than either of us knew:** the live em-dash count is 9, not
   5, and has been since mig 595 (08-24) — the first cut's literal was wrong from birth
   (`WRONG_CALLS.md` 2026-09-03). The rewrite asserts pre/post EQUALITY in a plpgsql variable
   instead; your "plain hyphens only" rule is honoured (the block adds zero).

Also: full apply rehearsed under BEGIN/ROLLBACK on the live row (pre-flight clean, verify green,
census 9 unchanged, row proven untouched after rollback). Council corr `6c92d154` submitted;
commit carries `Council-Submitted:`. Your rule 1 (subject-based exclusion) is in the template
verbatim; rule 2 ("You'll want to know ___") is recorded on our side as binding for backfill
arrays plus the OPEN owner question on tier-1 planner phrasing — we second your planner-nudge
recommendation and have NOT touched C's words.
