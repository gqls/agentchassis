# RUNBOOK — bugfix 191 chrome link policy

Every command here had a gotcha. The gotcha is attached to the command.

## R1 — Reproduce / verify the defect (the CORRECTED query)

`bugs_open/191`'s own verification SQL **over-reports twice**. Use this instead:

```sql
SELECT s.domain,
       substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"') AS cta_href,
       p.build_status, p.deployed_at IS NOT NULL AS ever_deployed
  FROM site_components sc
  JOIN sites s ON s.id = sc.site_id
  LEFT JOIN pages p ON p.site_id = sc.site_id
       AND p.url = substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"')
 WHERE sc.slot_name = 'header'
   AND substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"') IS NOT NULL  -- ⚠ THE FIX
   AND p.deployed_at IS NULL;
```

⚠ **Trap 1 — the NULL-join inflates the answer.** Without the added line, a header
with **no** `header-cta` at all produces a NULL `p.*` row, which satisfies
`p.deployed_at IS NULL` and comes back as a hit. Run the original and you get **6
rows, 4 of them with an empty `cta_href`**. Those 4 are not sites with a broken
button; they are sites with no button.

⚠ **Trap 2 — the DB cannot answer the question.** Of the 2 real rows,
`lendzy.co.uk/tools/price-cap-checker/index.html` **serves HTTP 200**.
`deployed_at IS NULL` means "no recorded deploy", **not** "does not serve"
(`bugs_open/098`'s durable half). **The confirmed live 404 was 1, not 6 and not 2.**

```bash
# ALWAYS follow the SQL with this. It is the only authority.
for u in <each surviving cta_href>; do
  printf '%s -> ' "$u"; curl -s -o /dev/null -w '%{http_code}\n' --max-time 20 "$u"
done
```

## R2 — Size the escape population (the query that must be SPLIT)

```sql
SELECT CASE WHEN pages_total = 0 THEN 'no pages at all (never built)'
            WHEN shipped = 0     THEN 'HAS pages, NONE shipped -> takes the escape'
            ELSE 'has shipped pages -> strict' END AS bucket,
       count(*) AS sites
FROM (
  SELECT s.id, s.domain, count(p.id) AS pages_total,
         count(*) FILTER (WHERE NOT (p.deployed_at IS NULL
                          AND COALESCE(p.build_status,'') <> 'deployed')) AS shipped
    FROM sites s LEFT JOIN pages p ON p.site_id = s.id
   GROUP BY s.id, s.domain
) t GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **The gotcha is the `CASE`, and it is the whole point.** The obvious version
counts `shipped = 0` and returns **19 of 38** — which reads as "half the fleet
bypasses the filter" and would have killed the design. Splitting on
`pages_total = 0` shows the truth: 18 sites have no pages at all (chrome renders
nothing either way), 19 are strict, and **1** actually takes the escape. A site with
no pages is not a site opting out of filtering.

## R3 — Blast radius by grep (run these before believing any plan, including a model's)

```bash
grep -rn "loadResolverPageSet(" platform --include="*.go"   # expect: def + 3 call sites
grep -rn "loadFetchablePageSet(" platform --include="*.go"  # expect: def + applyNavVisibility only
grep -rn "deployed_at IS NULL" platform --include="*.go" | grep -v datahelpers | grep -v queryresolve | grep -v _test
```

⚠ The third one's 3 hits today are **comments and one human-readable fix string**,
not code. Read them before counting them as hand-spelled predicates.

## R4 — Test, build and MUTATE without dirtying the shared tree

```bash
S=/home/ant/.cache/claude-mut-191
rm -rf "$S"; mkdir -p "$S"
git archive HEAD | tar -x -C "$S"
cp <each changed file> "$S/<same path>"          # HEAD + your patch, nobody else's WIP
cd "$S" && TMPDIR=/home/ant/.cache/gotmp-191 go test ./platform/orchestration/...
```

⚠ **`TMPDIR` is not optional.** `/tmp` on this box is a **16 GB tmpfs shared by every
session** and it was 100% full: `go build` fails with
`compile: writing output: write $WORK/b001/_pkg_.a: no space left on device`, which
reads like a broken toolchain and is not. The session scratchpad is on the same
tmpfs, so it is no escape — use a path under `/home`.

⚠ **Mutate by changing a CONDITION, never by deleting a block.** Deleting the
`case deployedPages == 0:` arm produced `[build failed]`, which is **not** the test
going red — a build failure proves nothing about the guard. `sed -i 's/case
deployedPages == 0:/case deployedPages == -1:/'` keeps the mutant compiling, so the
RED that comes back is the guard firing.

The four mutations and what each must turn red:

| mutation | must fail |
|---|---|
| CTA call site reverted to `loadResolverPageSet` | both source scans |
| deployment predicate dropped from the fetchable SQL | 5 tests (mock expectation stops matching) |
| `deployedPages == 0` → `== -1` | first-build escape tests |
| `Allows` returns `true` unconditionally | the never-deployed test + the nav suite |

⚠ And the standing control: **`nav_visibility_test.go` must pass UNEDITED.** If the
refactor needs it edited, the refactor changed behaviour and is wrong.

## R5 — Concept register

```bash
python3 scripts/test-concept-register-drift-local.py
```

⚠ **It reads at `HEAD`, not the working tree.** It cannot see your uncommitted entry
and will report the pre-change state as clean — a pass here says **nothing** about
your addition. Bump the index headline count in the same commit as the entry and the
row, then re-run *after* committing.

## R6 — LANDMINES sync

```bash
./scripts/landmines-sync.py --apply     # --check exits 1 if drifted
```

⚠ `--check` reports ~228 "to insert/refresh" as its normal steady state; that is the
tool refreshing owned rows, not 228 entries of drift. Read the `content changed` and
`orphaned` lines instead.

## R7 — Council gate

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_191_chrome_link_policy/submission_191_r1.json
```

`SUBMISSION_CORR = 78b0b7ff-f88d-402b-8f8f-ca4ae01c2d30` (round 1, 2026-08-04).
Budget **~30 minutes**, not 2 — the council itself is 2–5 min, the dispatch queues.
Find the run by payload, never by a printed id:

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '78b0b7ff-f88d-402b-8f8f-ca4ae01c2d30';
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='78b0b7ff-f88d-402b-8f8f-ca4ae01c2d30' AND kind='council_report' ORDER BY created_at;
```

## R8 — After the next chassis roll (OWED, the bug is not proven until these pass)

```bash
# both replicas, one exec each — a roll is not evidence your fix shipped
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "LoadChromeLinkPolicy"'          # expect >0
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "headerPages := loadResolverPageSet"'  # NEGATIVE control, expect 0
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "loadFetchablePageSet"'          # POSITIVE control, expect >0
```

⚠ This change **removes** a string, so a real negative control exists here — take it,
rather than inventing one (`bugs_open/153`).

Then re-run `nav-updater` on `mortgagecalculator.co.uk` (locked, 25 unbuilt pages and
1 deployed — it recreates the condition on demand), re-run **R1**, and **curl**.
