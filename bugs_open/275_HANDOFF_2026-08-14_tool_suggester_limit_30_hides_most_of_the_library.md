# 275 — tool-suggester's `load_library_tools` LIMIT 30 silently hides 38 of 68 library tools, alphabetically

**Filed 2026-08-14, webdesign_uk_build_service lane**, found while shipping
migration 406 (the requires-backend gate) through the same query. Not run
through the 090 loop — stated substitution: the claim is direct arithmetic on
the live config and live data, both quoted below, with no mechanism inference;
every figure is reproducible by one query each.

## The defect

`tool-suggester`'s `load_library_tools` step (agent_definitions,
`{workflow,steps,load_library_tools,config,query}`) ends:

```
ORDER BY display_name LIMIT 30
```

Measured 2026-08-14 against the live library:

```sql
SELECT count(*) FROM content_components
WHERE component_level='tool' AND forked_from IS NULL
  AND is_active AND html_template != '';
-- 68
```

So the LLM `suggest_tools` step is shown the first **30 of 68** library
masters by `display_name` — **38 tools can never be suggested for any
site**, and which 38 is an accident of alphabetical order, not a judgement of
relevance. The cap is invisible in output: the suggester returns plausible
suggestions either way, so nothing ever looks wrong (a silent cap — the
"no silent caps" class).

## Why it matters

- Suggestion quality fleet-wide is judged against under half the library.
- The exclusion is alphabetical: a rename can move a tool in or out of the
  visible set, which will read as the suggester "deciding" differently.
- The library grows (68 and rising — 27 as of 2026-07-20 per
  `plan_sections_action.go`'s calibration comment), so coverage decays
  further with every added tool.

## Fix candidates (ranked by what closes the door)

1. **Remove the arbitrary cap and send a compact list** — the query selects
   five short columns; 68 rows of that is small. If the real constraint is
   prompt size, cap by TOKENS at the prompt assembly, not by row count in the
   dark. Closes the door: the visible set is the library.
2. **Rank before capping** — if a cap must stay, `ORDER BY` something
   meaningful (usage_count, avg_quality_score, category diversity), so the
   cut is a judgement rather than the alphabet. Door stays ajar (still a
   silent cap) but the damage is chosen, not accidental.
3. Do nothing to the query; document the cap in the step description. Weakest
   — a doc comment is not an enforcement mechanism.

Whoever fixes: migration on the same step migration 406 touched
(`sql_for_agents/406_tool_suggester_requires_backend_gate.sql` is the worked
example, including the snapshot + DO/RAISE verify pattern), and mind that 406's
gated query is now the base text — do not restore the ungated query by
copying an older sketch.

## How to verify a fix

Disagreeing pair: pick a tool sorting past position 30 by display_name
(e.g. anything starting late in the alphabet from
`SELECT display_name FROM content_components WHERE component_level='tool'
AND forked_from IS NULL AND is_active ORDER BY display_name OFFSET 30`),
run tool-suggester for any site, and confirm the late-alphabet tool appears
in the `library_tools` input of the `suggest_tools` LLM call
(`llm_call_log`, rendered prompt — the 2026-08-09 method). Before the fix it
cannot appear; after, it can.
