# RUNBOOK — the template-variable ↔ `input_fields` lint

## Run it

```bash
scripts/audit-template-input-fields.sh          # human-readable
scripts/audit-template-input-fields.sh --json   # the raw report
```

Exit: `0` no unreachable roots · `1` unreachable roots found · `2` could not determine.

⚠ **Read a refusal from EMPTY stdout, never from the exit code.** `go run` folds the tool's
exit 2 into its own exit 1, so exit codes cannot separate "found findings" from "refused to
run". The wrapper already does this; if you call the binary directly, do the same.

## Run it against a file instead of the cluster

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow',
                                    'agent_prompt_template', default_config->'prompt_template'))
FROM agent_definitions
WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active
  AND default_config ? 'workflow';" > /tmp/fleet.json

go run ./cmd/config-key-audit --template-input-fields < /tmp/fleet.json
```

⚠ **The `agent_prompt_template` projection is NOT optional and NOT cosmetic.** Every other
audit script's export omits it. Feed one of those in and the mode **refuses** (exit 2) rather
than running — because without it the agent-level prompt tier is invisible, and a run blind to
6 live render sites reports fewer findings and reads exactly like a clean fleet. If you see

```
no agent row carried an 'agent_prompt_template' key across N agents
```

you piped the wrong export. `jsonb_build_object` emits the key as `null` when an agent has no
prompt_template, so a correct export always carries it.

## Read the output

Three kinds, and **only `unreachable_root` exits 1**:

| kind | means | act on it? |
|---|---|---|
| `unreachable_root` | `input_data` is NOT in `input_fields`, so config alone decides it: the variable resolves on **no row, ever** | yes — add the root to `input_fields`, or delete the dead template block |
| `conditional_root` | `input_data` IS in `input_fields`, so ExtractFields also promotes every key of the runtime `input_data` map to the root — undecidable from config | read it, don't chase it |
| `declared_unread` | an `input_fields` entry no template variable reads; costs a whole-tree extraction every run | usually a template that lost a reference |

## Gotchas that cost time here

- **A dotted `input_fields` entry supplies its LAST segment.** `["a.b"]` makes `{{.b}}` work
  and `{{.a.b}}` NOT work. If a finding looks absurd, check this first.
- **The injected roots are per ACTION, not shared.** `voice_style` is available to
  `execute_llm_prompt` and NOT to `execute_vision_prompt`; `vision_image_manifest` is the
  other way round. Do not "fix" a finding by assuming the union.
- **A clean run over few templates is not a clean fleet.** The header line prints
  `templates checked`; on 2026-09-03 that was 139 of 1,474 steps. The mode refuses outright
  if it is 0.
- **Parse failures are printed and are NOT findings.** A template the binary could not parse
  is a template it did not CHECK. The report carries them as a first-class list for exactly
  that reason.

## Re-measure the estate's shape

```bash
# how many steps render a template at all, and via which tier
scripts/audit-template-input-fields.sh --json | python3 -c '
import json,sys; o=json.load(sys.stdin)
print(o["agents_scanned"],"agents /",o["steps_walked"],"steps /",o["templates_checked"],"templates",
      "(",o["templates_agent_tier"],"agent-tier )")'
```

## Tests

```bash
go test ./cmd/config-key-audit/            -run 'TemplateRoots|LiveWriter|Classification|AgentTier|RefusesAnExport|RangeBody'
go test ./platform/orchestration/datahelpers/ -run 'AlwaysEnsured|DottedInputField|SpeciallyHandled|InputDataPromotes|TemplateRootsFor'
go test ./platform/orchestration/actions/  -run 'DeclaredLLMInjected|VisionInjected|InjectedRootsAreNot|RendersPromptTemplate|TemplateRootsAvailableTo'
```

⚠ `go test ./cmd/config-key-audit/` also runs `TestShippedRegistryIsSelfConsistent`, which
**fails at committed HEAD** for an unrelated reason (`TOOL_PAGE_HELD_NO_TOOL_SOURCE`'s `why`
does not name its retention window — `bugfix_450`'s registry entry). Confirmed at
`cc572ea14` in a clean extract via `scripts/verify-head-builds.sh --test`, so it is not this
lane's. Scope your `-run` or you will chase it.
