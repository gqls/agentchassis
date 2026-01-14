# Workflow Validator

Validates that agent workflow field references match contracts and action expectations.

## Usage

```bash
# Validate agents from a JSON file
go run main.go -agents agents.json

# Validate a single agent from inline JSON
go run main.go -agent-json '{"type":"pageflow-builder","default_config":{...}}'

# Verbose mode (shows field tracing and info-level messages)
go run main.go -agents agents.json -verbose
```

## Getting Agent Data

Export agent definitions from your database:

```sql
-- Export single agent
SELECT json_build_object(
  'id', id::text,
  'type', type,
  'display_name', display_name,
  'default_config', default_config,
  'input_contract', input_contract,
  'output_contract', output_contract
) FROM agent_definitions WHERE type = 'pageflow-builder';

-- Export multiple agents
SELECT json_agg(json_build_object(
  'id', id::text,
  'type', type,
  'display_name', display_name,
  'default_config', default_config,
  'input_contract', input_contract,
  'output_contract', output_contract
)) FROM agent_definitions WHERE type IN ('pageflow-builder', 'page-content-writer', 'deployer-agent');
```

## What It Validates

1. **Step Flow**: Checks that `next_step`, `then_step`, `else_step` reference existing steps
2. **Action Config**: Validates required config keys for each action type
3. **Field References**: Identifies field paths that use `.result` or `.response` suffixes
4. **Input/Output Contracts**: Checks if contracts match what workflow actually uses/produces
5. **Loop Config**: Verifies loop actions have `items_field` or `iterate_over`
6. **Call Agent**: Warns about empty `input_fields`

## Output Categories

- **error**: Something is definitely broken (missing required config, broken step flow)
- **warning**: Something might be wrong (unused contract fields, missing output)
- **info**: Worth reviewing (path patterns that might cause issues)

## Adding Custom Action Specs

Create an `action_specs.json` file:

```json
[
  {
    "name": "my_custom_action",
    "reads_from": ["input_field", "data_field"],
    "writes_to": "output_field",
    "required_cfg": ["required_key"],
    "optional_cfg": ["optional_key"]
  }
]
```

Then run:
```bash
go run main.go -agents agents.json -actions action_specs.json
```