# RUNBOOK — 417 logo text policy

Every command here was needed and had a gotcha. Gotchas are attached, not appended.

## Is the exemplar fix (669) actually applied? Ask the ROW, not a migration tracker
There is no `schema_migrations_agents` table — that guess costs a round trip. And
`schema_migrations` has no `version` column. Read the artefact instead:
```sql
SELECT CASE WHEN default_config::text LIKE '%no text outside the wordmark itself%'
              THEN 'LICENCE STILL PRESENT (669 NOT applied)'
            WHEN default_config::text LIKE '%a text-free mark%'
              THEN '669 APPLIED'
            ELSE 'NEITHER — drifted' END, updated_at
FROM agent_definitions
WHERE type='build-site-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## The census — count the LICENCE, never the prohibition, and never a literal
⚠ The exemplar phrase itself matches `no text`, so "does the prompt forbid text?" scores the
contradictory prompts as SAFE. **And counting the licence by literal is still a literal** — the
model rewords it ("other than" vs "outside"). The binding census is a human read of the concept.
The mechanical pre-filter that is *honest about being a floor*:
```sql
SELECT s.domain, spi.id, spi.created_at, left(spi.prompt,200)
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id=spi.plan_id
LEFT JOIN sites s ON s.id=sp.site_id
WHERE spi.kind='logo' AND sp.is_current
  AND spi.prompt ILIKE '%wordmark%' AND spi.prompt NOT LIKE '%migration 6%';
```
⚠ `site_plan_imagery` has **no `updated_at`** column — naming it errors the whole query.

## Post-roll: did the guard REACH the generation? (the only decisive check)
```sql
SELECT a.id, s.domain, a.created_at,
       (a.origin_prompt LIKE '%Render a text-free mark%') AS text_free,
       (a.origin_prompt LIKE '%the exact wordmark%')      AS wordmark
FROM assets a JOIN sites s ON s.id=a.site_id
WHERE a.asset_key='logo' AND a.created_at > '<the roll>'
ORDER BY a.created_at DESC;
```
**Neither true = the guard was UNREACHED on that path** — check `kind` arrival first (two legacy
parents map no `kind`; LANDMINES). ⚠ `assets` has **no `asset_role`** column; it is `asset_type`,
and the useful key is `asset_key='logo'`.

## Rehearse a migration before applying it
`sed 's/^COMMIT;$/ROLLBACK;/' <file> | psql` — proves the row count inside the real transaction
without committing. The `DO`/`RAISE EXCEPTION` guard is what makes the count assertable; a verify
block of bare `SELECT`s cannot stop the `COMMIT`.

## Council submission — the schema, which the dry run will teach you for free
`DRY_RUN=1 097_TRIGGER_council_review_v1.sh <file>` spends nothing. Types that bit me:
- `plan.edits[].operation` ∈ `modify|add|remove|config_change` — **`create` is invalid, use `add`**.
- `plan.risks` must be a **STRING** (prose block).
- `plan.grounded_in` must be an **ARRAY of strings** — do not "fix" it to match `risks`.
- an edit's `sketch` must not be **comment-only** (every non-blank line starting `//`/`--`/`#` is refused).
One command answers all of it at once:
`python3 -c "import json;d=json.load(open(F));print({k:type(v).__name__ for k,v in d['plan'].items()})"`

## Verify the change against committed HEAD, isolated from other sessions' dirty files
```
./scripts/verify-head-builds.sh --with <each file> --test
```
⚠ It prints `FAILED` and **exits 0**, so `&&` chaining will not catch it. HEAD is independently
red in ~23 places. **Run the bare control** (`./scripts/verify-head-builds.sh --test`) and diff
the FAIL sets — the claim you can actually make is "every failure in my set is in the control's".
