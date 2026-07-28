# 134 — a doc convention for "optional" leaked into the real key name, so two config keys are inert

Filed 2026-07-28 by session "bugsearch 3", found with `cmd/config-key-audit --specs`
(SCR-005) while working the `bugs_closed/101` coverage ratchet. **Latent, not
biting:** the agent has never run.

## The defect

`product-spec-refresher/refresh_specs` (action `refresh_product_specs`) carries this
live config:

```json
{
    "limit?":    "input_data.limit",
    "site_id":   "site_record.site_id",
    "category?": "input_data.category"
}
```

The action reads `category` and `limit` — no question marks
(`refresh_product_specs_action.go:177,211,215`):

```go
Optional: []string{"category", "limit", "delay_ms", "llm_model"},
...
category := inputs.Get("category")
limit    := inputs.GetInt("limit", 20)
```

So **neither key ever resolves.** `limit` silently takes its hard-coded default of
20 and `category` is empty, whatever the caller passes in `input_data`. No Go code
anywhere reads a `?`-suffixed key: `grep -rn '"limit?"\|"category?"' --include=*.go
platform/ internal/` returns nothing.

## Where it came from — the seed, which is the part that matters

`docs/agent_docs/sql_for_agents/156_product_spec_refresher_agent.sql`. Line 15 is a
comment describing the agent's input contract:

```sql
-- {site_id, category?}. Requires the chassis image that registers
```

`category?` there is **documentation notation** — the common convention for "this
field is optional". Forty-five lines later the same notation appears inside the
actual JSON:

```sql
"category?": "input_data.category",
"limit?": "input_data.limit"
```

**Fix the seed, not just the live row.** A replay of 156 restores the defect, and
the seed is what a rebuild reads.

## Severity: latent

- `SELECT count(*) FROM orchestration_states WHERE ... 'product-spec-refresher'` → **0**
- no `scheduled_tasks` row matches `%spec%`

Nothing has ever run this agent, so nothing depends on the current behaviour and the
fix cannot regress anything. It also means the defect would have surfaced the first
time someone tried to pass a `category` and quietly got all categories.

## Fleet-wide sweep — this is the ONLY instance

Worth keeping; it is one query and it caught this:

```sql
SELECT ad.type AS agent, e.k AS step, v->>'action' AS action, ck.key AS suspicious_key
FROM agent_definitions ad,
     jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v),
     jsonb_object_keys(v->'config') AS ck(key)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
  AND jsonb_typeof(v->'config')='object'
  AND (ck.key LIKE '%?' OR ck.key LIKE '%*' OR ck.key LIKE '% %' OR ck.key LIKE '%:%')
ORDER BY 1,2;
```

2 rows, both above. `?`, `*`, a space and `:` are the punctuation that schema-doc
conventions use and JSON keys should not.

## Fix candidates, ordered by what closes the door

1. **Opt `refresh_product_specs` into the config-key contract** —
   `CheckConfig: true` on `RefreshProductSpecsInputSpec`. This is the door-closing
   fix and it is one line: the action is exactly the "category B" case the ratchet
   report separates out (spec exists, live step carries keys the spec does not
   list), so opting in makes the validator warn `unrecognised config key: category?`
   at runtime and `scripts/audit-config-keys.sh` report it offline. **Do this even
   if you also do 2** — otherwise the next typo is invisible again.
2. **Correct the two keys in the seed AND the live row.** Trivial and safe here
   because the agent has never run. Live: `agent_definitions.default_config`;
   repo: seed 156 lines 60-61 (and line 15's comment is fine — it is a comment).
3. **Do not "fix" this by adding `category?`/`limit?` to the spec.** That is the
   `WRONG_CALLS.md` 2026-07-28 mistake in its purest form — declaring a dead key
   makes it *recognised* and silences the detector, leaving the behaviour broken and
   the report clean.

## Ownership

`product-spec-refresher` is not this lane's agent and `scripts/who-owns.py
refresh_product_specs` finds no bug file. Seed 156 has no stated owner. The keys are
**inert**, so correcting them IS a behaviour change (`limit` would start honouring
the caller instead of defaulting to 20) — small, obviously intended by the seed's own
comment, and with zero runs behind it, but somebody's call rather than a silent
side effect. Same shape as `bugs_closed/101` residual 2, where the owner ruled to
leave the config honest-but-warning rather than change behaviour.

## Related

- `bugs_closed/101` — the parent class: a config key nothing reads is
  indistinguishable by inspection from a live one. This is a fresh instance found by
  101's own tooling, which is the tooling working.
- `bugs_open/133` — same session, same class one level down: a *message* that
  describes something that did not happen.
- concept register SCR-003 (`CheckConfig`), SCR-004, SCR-005 (`--specs`, what found it).
