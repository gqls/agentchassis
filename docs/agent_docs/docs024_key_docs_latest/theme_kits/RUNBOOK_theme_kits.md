# RUNBOOK — theme kits

Every command here was hard to get right at least once, and the gotcha is attached to
the command rather than kept in someone's scrollback. **Change a command HERE when it
changes.**

---

## 1. Is any of this actually live?

Three independent facts — binary, schema, adoption — and this lane has got each of them
wrong at least once. **Check all three; none implies another.**

### Binary

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=400 | grep -m1 'build provenance'
```

⚠ **This is a STARTUP line, so it scrolls.** On `agent-chassis` it was already out of
reach of `--tail=3000` hours later. **An empty result means "not in range", NOT
"unstamped".** Fall back to the binary probe, which has no shelf life:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
for n in apply_theme_kit page_archetypes fork_theme_from_site zzz_not_a_real_action_zzz; do
  kubectl -n ai-persona-system exec "$POD" -- grep -aq "$n" /proc/1/exe \
    && echo "$n PRESENT" || echo "$n absent"
done
```

⚠ **Always run BOTH controls** — `fork_theme_from_site` is a pre-existing action that
must be PRESENT, and `zzz_not_a_real_action_zzz` must be ABSENT. Without them a probe
that matches everything (or nothing) reads as a clean answer. **Never `strings`**: it is
absent from the debian-slim images, and behind the customary `2>/dev/null` its failure
is indistinguishable from "not stamped".

### Schema

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT to_regclass('public.theme_kits'), to_regclass('public.page_archetypes');"
```

⚠ **Migrations do NOT ride a build** — applying them is a separate action, and a chassis
carrying the code against missing tables **degrades silently by design**
(`loadSiteThemeKitDefaults` errors are swallowed and every consumer falls through to
existing behaviour). Nothing breaks; nothing works either.

### Adoption

```sql
SELECT count(*) FROM site_specs WHERE aspect='theme_kit_adoption' AND is_current;
```

**0 as of 2026-09-03.** A durable zero means the mechanism is built but undriven, which
is not evidence it works.

---

## 2. Applying migrations for this lane

⚠ **`--apply` takes EVERY pending file, not yours.** Applying 689 and 691 with a bare
`--apply` would have swept a dozen other lanes' migrations into a live database. Scope
it instead:

```bash
D=$(mktemp -d)
cp docs/agent_docs/sql_for_agents/689_theme_kits.sql \
   docs/agent_docs/sql_for_agents/691_per_site_palettes_for_three_sites_on_a_shared_library_row.sql "$D"/
MIGRATIONS_DIR="$D" ./scripts/run-migrations.sh --apply
```

⚠ **The dry run's success message reads like the opposite of what it means.** With no
`--apply` the runner reports *"ran to its own COMMIT without error (everything rolled
back)"*. That is NOT applied. This lane wrote "deployed" into the concept register on
the strength of that line.

⚠ **A verify block of bare `SELECT`s cannot stop a `COMMIT`** — `ON_ERROR_STOP` ignores
a non-empty result. Use `DO` / `RAISE`, and induce the failure once to prove the guard
fires.

---

## 3. The measurements this lane keeps needing

**Chrome eligibility — use the predicate as WRITTEN, not as remembered.**
`chromePinEligibleSQL` (`component_library.go:334,375`) is
`is_active AND component_level IN ('site','header','footer','head')` and has **no
`forked_from` filter**.

```sql
SELECT function, count(*) AS eligible_rows
  FROM content_components
 WHERE is_active AND component_level IN ('site','header','footer','head')
 GROUP BY function ORDER BY 2 DESC;
```

⚠⚠ **`content_components` has BOTH a `name` AND a `function` column, holding
near-identical vocabularies by design. SELECT BOTH. Never filter on one and conclude
about the other** — that is how this lane retracted a true claim, having read
`WHERE function LIKE '%theme-chrome%'` → 0 rows as "these components do not exist" when
`header-theme-chrome`/`footer-theme-chrome` are `name` values whose `function` is
`site-header`/`site-footer`. **A `LIKE` probe returning 0 rows is evidence about the
column you queried and nothing else.** The query that actually answers it:

```sql
SELECT name, function, component_level, is_active, forked_from IS NULL AS unforked,
       (is_active AND component_level IN ('site','header','footer','head')) AS chrome_eligible
  FROM content_components WHERE function IN ('site-header','site-footer')
 ORDER BY function, chrome_eligible DESC, name;
```

**11 rows as of 2026-09-03, exactly 3 chrome-eligible:** `header-theme-chrome`,
`footer-theme-chrome`, and **`header-leopardess` — an ACTIVE FORK of one client's
header**, eligible because the predicate has no `forked_from` filter. The rows *named*
`site-header`/`site-footer` are `section`-level and ineligible. **So a function-name
subquery for `site-header` is ambiguous, and the extra row is one client's fork** — which
is why the seed hardcodes UUIDs.

⚠ **When a claim in a doc looks false, resolve the ARTEFACT by id before retracting it.**
Migration 689 names both UUIDs three lines above the comment this lane called wrong, and
`WHERE id IN (…)` would have settled it in one query. Grepping for the propagation is
also cheap and was what finally caught it: **70 files name these components and migration
339 has `RAISE EXCEPTION` drift guards on updating them.** A component that does not
exist does not have drift guards.

**Component function collisions — the number is meaningless without the predicate.**

```sql
-- 84 as of 2026-09-03: raw, every row
SELECT count(*) FROM (SELECT function FROM content_components
  GROUP BY function HAVING count(*)>1) x;
-- 3 as of 2026-09-03: canonical (site-header, site-footer, tool-agent-complexity-estimator)
SELECT count(*) FROM (SELECT function FROM content_components
  WHERE is_active AND forked_from IS NULL GROUP BY function HAVING count(*)>1) x;
```

⚠ The column is `forked_from`, not `forked_from_component_id`. Distinct functions are
**425 raw / 410 canonical as of 2026-09-03** — the "364" in older documents is stale by
addition.

**What the four kits actually pin** (the query that found the chrome no-op):

```sql
SELECT tk.name, hc.function, hc.component_level, fc.function, fc.component_level
  FROM theme_kits tk
  LEFT JOIN content_components hc ON hc.id = tk.header_component_id
  LEFT JOIN content_components fc ON fc.id = tk.footer_component_id
 ORDER BY tk.name;
```

All four return `site-header` / `site-footer` at `component_level='site'`, which is
exactly what `ChromeSlotFunction()` (`component_library.go:386`) hardcodes for an
unpinned site.

⚠ **Any "N of X ever" figure must union the archive.**
`site_work_items` is a rolling window and terminal rows move to
`site_work_items_archive`. This lane quoted 1 where the truth was 2. `site_specs` does
NOT archive (it versions in place under `is_current`), so the same union is useless
there. Cheap guard before quoting any "ever" figure:

```sql
SELECT table_name FROM information_schema.tables WHERE table_name LIKE '%work_item%';
```

⚠ **`psql` through `kubectl exec` regularly takes 1–3 minutes here.** A 120 s timeout
will be moved to the background mid-query. Put the SQL in a file and pipe it with
`-f -`, and run it in the background rather than fighting the timeout.

---

## 4. Council gate for this lane

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

**Always dry-run first — it validates everything and the scope admission for free.**
Then drop `DRY_RUN=1`. For a resubmit after REVISE, keep the trail on one correlation:

```bash
RESUBMIT_CORR=bed139b2-f512-436a-9ba8-ff2fbfade8ef \
  ./docs/.../097_TRIGGER_council_review_v1.sh <submission.json>
```

Traps this lane hit or verified:

- **≤8 edits, one file per edit.** Naming two files in one edit passes every local check
  and is refused server-side. **Merge two edits on the SAME file** to stay inside the
  cap — that is the sanctioned move, and it is how the `candidates` fix folded into the
  layout edit.
- **A sketch whose every non-blank line is a comment is refused** ("a fix plan proposes
  changes, not observations"). Observations belong in `rationale` / `grounded_in`.
- **`.plan.risks` must be a STRING**, not an array.
- **A figure in `grounded_in` must carry its predicate as a RUNNABLE query.** Round 1
  said "3 collisions after the canonical predicate" in prose; the reviewer reconstructed
  a different query, got 84, and reviewed that. Both numbers were right under their own
  predicate.
- **The sketch is what gets reviewed, not the repository.** Round 1's rationale claimed a
  typography guard that exists in the code and was absent from the sketch. Verdict:
  REVISE, correctly.

Resolve a verdict:

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<CORR>';
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='<CORR>' AND kind='council_report' ORDER BY created_at;
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

⚠ Budget **~30 minutes**, not 2 — the council runs in 2–5 minutes but the dispatch
queues behind the fleet. A missing row is almost always latency, not a dropped dispatch.
⚠ **Never write `Council-Reviewed:` on a verdict you have not read** — 098 buckets that
as MISMATCH. Use `Council-Submitted:` before the verdict lands; it is credited
automatically if the correlation later approves.

---

## 5. Editing the Go in this lane

```bash
gofmt -l <file> | wc -l          # gate on the COUNT
go build ./platform/orchestration/actions/
```

⚠ **`gofmt -l` exits 0 whether or not it lists a file**, so
`gofmt -l f && echo BAD || echo OK` always prints BAD. Gate on `| wc -l`, and prove the
check discriminates with a deliberately malformed control file.

⚠ **Adding a comment INSIDE a Go map literal breaks gofmt's alignment group** and makes
it rewrite every neighbouring key, turning a one-line change into a noisy diff. Put the
comment above the `return`.

⚠ **Go changes are inert until an image is rebuilt and rolled; DB config is live
immediately.** `make build-<service>` builds from committed HEAD, so **commit, then
build.**
