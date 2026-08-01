# RUNBOOK — bugs_open/170, the chrome pin

The commands that were hard to get right, with the gotcha attached.

## R1 — the state of every chrome pin in the fleet

The one query that answers "is this bug still real?". Joined from
`style_collections`, **not** from `sites` — join from `sites` and you see 4 rows
and conclude 4 collections; there are 4 collections but two of them
(`bold-gradient`, `minimal-light`) are used by **zero** sites and are invisible
that way, while `professional-dark` is one collection serving three.

```sql
SELECT scol.name AS collection,
       hc.name AS header_pin, hc.is_active AS hdr_active, hc.forked_from IS NOT NULL AS hdr_fork,
       fc.name AS footer_pin, fc.is_active AS ftr_active,
       (SELECT count(*) FROM sites s WHERE s.style_collection_id = scol.id) AS sites_using
FROM style_collections scol
LEFT JOIN content_components hc ON hc.id = scol.header_component_id
LEFT JOIN content_components fc ON fc.id = scol.footer_component_id
WHERE scol.header_component_id IS NOT NULL OR scol.footer_component_id IS NOT NULL
ORDER BY 1;
```

## R2 — pin predicate vs pool predicate, side by side

The check that stops you copying `chromeEligibleSQL` into the pin path. It is
positive and negative control in one query: the three deactivated pins must come
back false on both, and `header-leopardess` must come back **true on the pin
predicate and false on the pool one**. If they agree on every row, you have
collapsed the asymmetry and your first live action will be deleting a client's
bespoke header.

```sql
SELECT s.domain, hc.name AS header_pin,
       (hc.is_active AND hc.component_level IN ('site','header','footer','head')) AS pin_predicate,
       (hc.is_active AND hc.forked_from IS NULL
        AND hc.component_level IN ('site','header','footer','head'))              AS pool_predicate
FROM sites s
JOIN style_collections scol ON s.style_collection_id = scol.id
JOIN content_components hc ON hc.id = scol.header_component_id
ORDER BY 1;
```

## R3 — what the extended detector would file, before you ship it

Run this rather than trusting the check to be bounded. It should return one row
per genuinely wrong pin and **nothing** for the legitimate fork. It returned 7 on
2026-08-01.

```sql
SELECT s.domain, p.slot_name, p.name AS deactivated_pin, p.collection
FROM sites s, LATERAL (
  SELECT 'header' AS slot_name, hc.name, scol.name AS collection,
         (hc.is_active AND hc.component_level IN ('site','header','footer','head')) AS eligible
  FROM style_collections scol JOIN content_components hc ON hc.id = scol.header_component_id
  WHERE scol.id = s.style_collection_id
  UNION ALL
  SELECT 'footer', fc.name, scol.name,
         (fc.is_active AND fc.component_level IN ('site','header','footer','head'))
  FROM style_collections scol JOIN content_components fc ON fc.id = scol.footer_component_id
  WHERE scol.id = s.style_collection_id
) p WHERE NOT p.eligible ORDER BY 1,2;
```

## R4 — PREPARE every new query before shipping it

`go build` cannot parse SQL, and a mocked test proves only that the Go compiles.
`PREPARE` is the cheapest thing that reads the query against the live schema.
Use `-v ON_ERROR_STOP=1` or a later statement's success hides an earlier failure.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 <<'SQL'
PREPARE q(uuid) AS <your query>;
\echo '--- PARSED ---'
SQL
```

**Gotcha:** a predicate built by Go concatenation (`(` + chromePinEligibleSQL("hc.") + `)`)
must be expanded by hand for the PREPARE — you are testing the string the driver
sends, not the one in the source.

## R5 — build and test when the shared tree does not compile

The working tree is shared and frequently carries another session's half-finished
edit; on 2026-08-01 it was `ai_actions.go:347 declared and not used`. A red
`go test` in the tree tells you nothing about your own change, and a green one is
not evidence either.

```bash
T=<scratchpad>/verify170
rm -rf "$T" && mkdir -p "$T" && git archive HEAD | tar -x -C "$T"
cp <only the files you changed> "$T/<same paths>"
cd "$T" && go build ./... && go test ./platform/...
```

**Gotcha:** check `df -h /tmp` first — the archive tree is ~1GB and /tmp is a 16G
tmpfs that other sessions are also using (it was at 89% on 2026-08-01).

## R6 — prove a scan test actually fires (and that a FAIL is an assertion)

A mutant that breaks the build prints the same FAIL as a mutant that was caught.
Separate the two steps, always.

```bash
cat > "$T/platform/orchestration/actions/zz_induce_170_tmp.go" <<'EOF'
package actions
// a consumer written the way the originals were
func zzInduce(ctx context.Context, db interface{}, coll *StyleCollection) (*Component, error) {
	return GetComponentByID(ctx, db, *coll.HeaderComponentID, nil)
}
EOF
cd "$T"
go build ./platform/orchestration/actions/ && echo "MUTANT COMPILES CLEANLY"   # step 1
go test ./platform/orchestration/actions/ -run TestNoConsumerDereferences      # step 2
rm "$T/platform/orchestration/actions/zz_induce_170_tmp.go"                    # step 3
```

## R7 — find a diagnosis run's artifacts, and read the failure mode when there are none

The 090 trigger prints TWO correlations. The artifacts are under the **run**
one (`RUN_CORRELATION_ID`), not the intake one.

```sql
SELECT kind, iteration, length(body), source_agent, created_at
FROM diagnosis_artifacts WHERE correlation_id LIKE '<run-corr-prefix>%' ORDER BY created_at;
```

**Gotcha, and it cost this lane a wait:** four `bundle` rows and **no**
`iteration_note` / `council_report` / `doc_note` is not "still running" — it is a
run that completed without concluding. Check the last bundle for
`0 of N in-scope symbol(s) rendered with a body`: the bundle body cap is 60,000
chars and `component_library.go` is 93,905 bytes, so the symbol under test was
omitted from every iteration. The column is `body`, not `content`.
