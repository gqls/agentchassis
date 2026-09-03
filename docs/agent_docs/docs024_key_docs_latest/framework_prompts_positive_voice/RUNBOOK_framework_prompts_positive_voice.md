# RUNBOOK — framework prompts, positive voice

Commands this lane had to get right, each with its gotcha. Update HERE when one changes.

## Read a live prompt (the row is the fact; the seed is history)

```bash
# every prompt_template string, largest first (jsonb_path_query walks any depth)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT type, length(q.v #>> '{}') FROM agent_definitions,
 LATERAL (SELECT jsonb_path_query(default_config,'strict \$.**.prompt_template') AS v) q
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND jsonb_typeof(q.v)='string'
ORDER BY 2 DESC;"
# the writer's prompt, whole
... -c "SELECT p.v #>> '{}' FROM agent_definitions, LATERAL (SELECT jsonb_path_query(default_config,'strict \$.**.prompt_template') AS v) p
WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```
Gotcha: inside a double-quoted bash string the JSONPath `$` must be written `\$` or bash eats it and
psql sees `strict .**.prompt_template` (a syntax error that reads like a jsonpath fault).
The writer's `prompt_template` lives at `workflow.steps.process_sections_loop.config.sub_workflow.steps.generate_content.config.prompt_template`;
`default_config->'workflow'->'steps'->'generate_content'` returns NULL, silently.

## The house voice row and the build standard row

```bash
... -c "SELECT config_name, length(config->>'text'), updated_at FROM agent_default_configs WHERE config_name IN ('voice_style_block','build_standard_block');"
... -c "SELECT config->>'text' FROM agent_default_configs WHERE config_name='voice_style_block';"
```
Gotcha: `updated_at` on the voice row reads 2026-08-13 although migration 628 rewrote the text on
2026-08-25 (628 never bumped it). Anchor on the TEXT (a 628 phrase), never on the timestamp.
Which prompts read it: `... LIKE '%{{.voice_style}}%'` over the enumeration above (7 of 141 on 2026-09-03).

## Model census and writer volume

```bash
... -c "SELECT m.v->>'model', count(*), count(DISTINCT type) FROM agent_definitions,
 LATERAL (SELECT jsonb_path_query(default_config,'strict \$.**.ai_service') AS v) m
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND jsonb_typeof(m.v)='object' GROUP BY 1 ORDER BY 2 DESC;"
... -c "SELECT step_name, model, count(*), sum(input_tokens), sum(cache_read_input_tokens), sum(output_tokens)
FROM llm_call_log WHERE agent_type='page-content-writer' AND created_at > now() - interval '7 days' GROUP BY 1,2 ORDER BY 3 DESC;"
```
Gotcha: `step_name` is per loop iteration (`process_sections_loop_iter_N_generate_content`); sum
across N for "the writer step". Date every figure: 5,058 generate_content calls / 38.39M in / 6.94M out
for the 7 days to 2026-09-03; cache reads 0.

## Council run by correlation (find by payload, never by the printed id)

```bash
... -c "SELECT current_step, status, updated_at FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' LIKE '6c92d154%' ORDER BY updated_at DESC LIMIT 2;"
```

## Test-render a block (this lane's copy of the finetuning harness)

```bash
cd docs/agent_docs/docs024_key_docs_latest/framework_prompts_positive_voice/render_test && go run . | tee OUTPUT.txt
```
Builds the template exactly as `datahelpers.RenderPromptTemplate` does (Funcs, Parse, Execute, default
options: a missing map key prints `<no value>`; a `range` over a key absent from `input_fields` renders
EMPTY with no error). Fixtures are real `orchestration_states` rows. Copied from
`finetuning_uk_service/render_test_641/` on 2026-09-03; theirs is not edited from here.

## Who owns a prompt

`scripts/who-owns.py` resolves BUG numbers and slugs only; a prompt name returns "No bug file matches",
which reads as unowned and means "not a bug". For a prompt: `git log --oneline -5 -- docs/agent_docs/sql_for_agents/NNN_*`
on its latest seed, then the lane HANDOFF that commit names.
