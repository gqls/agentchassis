# RUNBOOK — identity source resolver (`bugs_open/072`)

Every query here was run on 2026-07-31 against the live `clients_db`. Gotchas are
attached to the query that has them, not collected at the bottom.

DB access: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## §1 — Where does each site's contact identity actually live?

**This is the query that flipped the bug.** It puts the three candidate stores
side by side. Run it before believing anything about which store is populated.

```sql
SELECT s.domain,
  COALESCE(NULLIF(s.email,''),'(none)')                              AS sites_email,
  COALESCE(NULLIF(s.phone,''),'(none)')                              AS sites_phone,
  CASE WHEN ss.data->>'email' <> '' THEN 'Y' ELSE '-' END            AS spec_flat,
  CASE WHEN ss.data->'contact'->>'email' <> '' THEN 'Y' ELSE '-' END AS spec_nest
FROM sites s
JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'identity' AND ss.is_current
WHERE s.domain NOT LIKE 'pool-%'
ORDER BY 4, 1;
```

**Gotchas, all three load-bearing:**

- **`WHERE domain NOT LIKE 'pool-%'` is required.** `sites` holds 14
  `pool-*.internal` rows (industry pools, no content) alongside the real sites.
  Counting them turns "12 of 15 have an email" into "12 of 29" and makes the
  defect look three times worse than it is. **A count off `sites` with no domain
  filter is answering a different question.**
- **`ss.data->>'email' <> ''` is null-safe by accident, deliberately relied on.**
  If the key is absent, `->>` yields NULL and `NULL <> ''` is NULL → falsy. So
  absent and empty both read as `-`, which is exactly what the resolver does
  (`navigateMap` treats `""`, `[]` and nil as not-found). Do not "fix" this to
  `IS NOT NULL`: that would count an explicit `"email": null` as present, which
  the resolver does not.
- **`is_current` is not optional.** `site_specs` keeps superseded rows;
  `idx_site_specs_current` is UNIQUE on `(site_id, aspect) WHERE is_current`.
  `ensureSpecs` (`plan_sections_action.go:122`) filters on `is_current = true`, so
  any query that does not is describing a different resolver than the one running.

## §2 — Does the nested sub-object exist but hold nothing?

The distinction the bug file missed. `contact_keys` non-empty with every value
`-` is the "shape present, facts absent" state.

```sql
SELECT s.domain,
  CASE WHEN ss.data->'contact'->>'email' <> '' THEN 'Y' ELSE '-' END AS nest_email,
  CASE WHEN ss.data->'contact'->>'phone' <> '' THEN 'Y' ELSE '-' END AS nest_phone,
  (SELECT string_agg(k,',') FROM jsonb_object_keys(COALESCE(ss.data->'contact','{}'::jsonb)) k) AS contact_keys
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE ss.aspect = 'identity' AND ss.is_current ORDER BY 1;
```

`COALESCE(…,'{}'::jsonb)` matters: without it, `jsonb_object_keys` on a missing
`contact` key errors out on the one site that has none, and you lose the whole
result set rather than one row.

## §3 — Fleet census: which declared source paths resolve at all?

Replicates `resolveSpecPath` + `navigateMap` in SQL: first segment is the aspect,
the rest is a `#>` path. **Found the P3 finding (74 of 100 paths name an aspect no
site has).**

```sql
WITH srcs AS (
  SELECT DISTINCT f.value->>'source' AS src
  FROM content_components c, jsonb_each(c.input_schema->'fields') f
  WHERE f.value->>'source' LIKE 'site_specs.%' AND c.is_active),
parsed AS (
  SELECT src, split_part(substring(src from 12), '.', 1) AS aspect,
         string_to_array(substring(src from 12), '.') AS segs FROM srcs),
sitelist AS (SELECT DISTINCT s.id, s.domain FROM sites s JOIN site_specs x ON x.site_id = s.id),
ev AS (
  SELECT p.src, p.aspect, s.domain,
    (SELECT ss.data #> p.segs[2:] FROM site_specs ss
      WHERE ss.site_id = s.id AND ss.aspect = p.aspect AND ss.is_current LIMIT 1) AS val,
    EXISTS (SELECT 1 FROM site_specs ss
      WHERE ss.site_id = s.id AND ss.aspect = p.aspect AND ss.is_current) AS aspect_exists
  FROM parsed p CROSS JOIN sitelist s),
agg AS (
  SELECT src, aspect,
    count(*) FILTER (WHERE val IS NOT NULL AND val NOT IN ('null'::jsonb,'""'::jsonb,'[]'::jsonb)) AS n_resolve,
    count(*) FILTER (WHERE aspect_exists) AS n_aspect
  FROM ev GROUP BY 1,2)
SELECT CASE WHEN n_resolve > 0 THEN 'A: resolves somewhere'
            WHEN n_aspect = 0 THEN 'C: ASPECT DOES NOT EXIST on any site'
            ELSE 'B: aspect exists, leaf never resolves' END AS category,
       count(*) AS distinct_paths, string_agg(DISTINCT aspect, ', ') AS aspects
FROM agg GROUP BY 1 ORDER BY 1;
```

**Gotchas:**

- **`substring(src from 12)` strips the literal `site_specs.`** (11 chars + 1).
  Hardcoded; if the prefix ever changes this silently mis-parses every row rather
  than erroring.
- **`segs[2:]` — slice from 2, not 1.** Element 1 is the aspect, which is matched
  in the WHERE, not navigated. Off-by-one here yields 0 resolutions fleet-wide and
  looks like a catastrophic finding.
- **`val NOT IN ('null'::jsonb,'""'::jsonb,'[]'::jsonb)` replicates `navigateMap`'s
  emptiness rule.** Drop it and 79 never-resolving paths become ~20, because
  `"email": null` counts as resolved. **The Go function's emptiness semantics are
  part of the measurement, not a detail.**
- `sites.deployed_at` **does not exist** — it is `last_deployed_at`. The
  `EXISTS (… site_specs …)` join is used instead to mean "a real, worked site".

## §4 — Verifying the fix after the chassis rolls

The fix is Go, so it is inert until a rebuild + roll. **Do not verify against git
or the image tag.**

1. **Pod-grep with a positive control in the same exec** (a roll is not evidence
   your build shipped — `bugs_open/153`):
   ```sh
   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
     'strings /app/agent-chassis | grep -c "resolved from the canonical sites row"; \
      strings /app/agent-chassis | grep -c "plan_sections: loaded site_specs"'
   ```
   First count > 0 = my change is in. Second > 0 = the control, proving the grep
   and the binary path are sound. **A zero on the first with a zero on the second
   means the grep is broken, not that the fix is missing.**

2. **Induce the failing case**, on a site with an empty spec and a populated
   column (oufe, robot-hands, vetcomparison, vonc, webdesign). Rebuild its contact
   page and confirm **three** components, with the email in the third:
   ```sql
   SELECT slot_name, left(content_data::text, 120) FROM page_components pc
   JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
   WHERE s.domain = 'vonc.com' AND p.name = 'contact' ORDER BY slot_name;
   ```
3. **Negative control** — `gamesdesign.co.uk` has no contact fact in any store and
   must STILL withhold the section. If it starts rendering, the fallback is
   fabricating a value and the change is wrong.
4. **Confirm which store won**, from the chassis log line the fix emits:
   `plan_sections: site_specs path resolved from the canonical sites row` with
   `requested` and `resolved_from`. A build where the section appears but no line
   fired resolved it literally — check you actually induced the failing case.

## §5 — Running the package tests on a shared tree

`go test ./platform/orchestration/actions/` failed twice on 2026-07-31 for
reasons that were **not mine and not at HEAD** — another session's uncommitted
files. To test without touching their work, use a build overlay:

```sh
printf 'package actions\n' > /tmp/stub_test.go
cat > /tmp/overlay.json <<'EOF'
{"Replace": {"/abs/path/to/their_broken_test.go": "/tmp/stub_test.go"}}
EOF
go test -overlay=/tmp/overlay.json ./platform/orchestration/actions/ -run 'YourTests' -v
```

`-overlay` is read-only w.r.t. the tree — nothing on disk changes, so it cannot
sweep or clobber another thread's file. Prefer it to moving files aside.

Note the corollary: **`derive_brand_head_assets_test.go` does not compile at HEAD**
(`a22010eaa` widened `lockedBrandHeadKeys` to return `assetLockSet` and left a
`map[string]bool` assertion behind). Another session fixed it in the working tree
while I was looking at it. Until that is committed, the actions test package is
un-runnable from a clean checkout — so a green local `go test` here is not
evidence about HEAD.
