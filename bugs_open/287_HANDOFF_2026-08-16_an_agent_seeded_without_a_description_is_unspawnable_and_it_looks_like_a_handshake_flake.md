# 287 — an agent seeded without a `description` cannot be spawned or resolved: five readers scan a nullable column into a Go string, and the failure looks exactly like the spawn→call handshake flake

**Filed 2026-08-16** by the `bugs_open/279` lane, from the FIRST spawn of
`brief-fidelity-auditor` after migration `419` wired it into the improvement loop.
**Status: data half FIXED AND LIVE (migration `420`); code half FIXED IN CODE, NOT
LIVE (rides the next chassis + core-manager roll); guard test committed.** No `090`
run — first-hand verification substituted (per the 2026-07-31 ruling): the error
string names the column and the Scan, the seed file was read, the readers were
enumerated by a source-scan test that found one my grep had missed, and the fix
was proven by the second dispatch completing end-to-end.

## The mechanism (verbatim error, first-hand)

```
step spawn_bf failed: failed to execute action spawn_agent: failed to get agent
definition: sql: Scan error on column index 3, name "description": converting NULL
to string is unsupported
```

`agent_definitions.description` is **nullable, no default**
(`information_schema.columns`). `AgentDefinition.Description` is a plain Go
`string` (`types.go:76`). Five readers `SELECT … description …` from the table and
`Scan` into it: `spawn_actions.go getAgentDefinition` (:2168 — the spawn path),
`ai_actions.go loadAgentDefinitionForAction` (:1314), `generate_image_actions.go
loadAgentDefinitionForImageAction` (:683), `platform/messaging/processor.go` (:381
— resolves the definition when a MESSAGE reaches an agent, so even a spawned agent
would fail on first message), `internal/core-manager/admin/system_handlers.go`
(:266 — the admin list, which `continue`s past the row so the agent silently
vanishes from the UI). A row with NULL description therefore cannot be spawned,
cannot receive a message, and does not appear in the admin list.

`brief-fidelity-auditor`'s seed
(`brochure_component_library/agents/brief-fidelity-auditor.seed.sql`) lists
`(id, type, display_name, category, default_config, idle_timeout_seconds,
is_active, created_at, updated_at)` — no `description`. `[MEASURED 2026-08-16]` it
was the ONLY live definition with a NULL description (1 of 193); the other 3 NULLs
in the table are inactive `diagnose-wiring-probe-*` scratch rows. So nothing had
ever hit this: the column is nearly always set, and a seed that forgets it IS the
population.

## Why it is nasty (and why it is filed rather than just fixed)

- **It wears the handshake flake's costume.** MEMORY records the spawn→call
  handshake failing ~half the time fleet-wide; a session seeing `FAILED|spawn_bf`
  would reasonably retry and move on. The error string is the tell — read it
  before retrying.
- **Behind `error_step` it is invisible.** `419` wires the auditor with
  `error_step: record_audit_pass` (one auditor must not strand the sweep). Without
  `420`, every improvement sweep would have reported COMPLETED with the auditor
  silently skipped — the exact "everything reports success" shape 279 was about,
  one layer up. **The auditor's 2026-08-13 findings existed, so it WAS run then —
  by some path that did not go through `getAgentDefinition`; the orchestration
  rows are pruned and the path is `[UNKNOWN]`.**

## Fix, both halves

1. **Data (LIVE):** migration `420` sets a description on the live row
   (snapshot taken; refuse-if-already-set guard). The second dispatch then
   completed: 8 findings → 8 routed items, zero minted, `unrouted` empty.
2. **Code (COMMITTED, NOT LIVE):** all five readers now `COALESCE(description,
   '')` — the idiom `fork_theme_composition.go`, `load_component_library_actions.go`
   and `render_js_snippets_for_site_action.go` already use for nullable text.
   `agent_definition_nullable_columns_test.go` scans the whole module for any
   `SELECT … description … FROM agent_definitions` that is not COALESCEd
   (INSERT…SELECT copies excluded — NULL || text stays NULL in the same column),
   with a pattern self-test. **Mutation-verified**: reverting one loader → 1
   failure naming the file. It found a sixth candidate my grep missed
   (`discovery_actions.go`'s variant clone), which turned out to be the copy case —
   that is what the exclusion is for.
3. NOT done: a `NOT NULL DEFAULT ''` on the column. It is the cleaner door but it
   is DDL on a shared table with 209 rows and a seed convention across the estate;
   flagged for an owner call rather than slipped into a lane fix. Until then the
   COALESCE + the test are the guard.

## Verify (post-roll)

`grep -c "COALESCE(description" platform/orchestration/actions/spawn_actions.go`
is not evidence; the evidence is a spawn of an agent whose description you have
NULLed on a scratch row succeeding on the rolled binary — or simply: no
`converting NULL to string` in chassis logs after a seed that forgets the column.
