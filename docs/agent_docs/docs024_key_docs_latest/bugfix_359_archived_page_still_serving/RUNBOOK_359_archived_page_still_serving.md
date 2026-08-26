# RUNBOOK — bugs_open/359

Every command that was hard to get right, with its gotcha attached. Change commands HERE,
not in a scrollback.

---

## 1. The census: which retired pages are still serving?

**This is the measurement the bug's §6 candidate 1 asks for, and the arm every fix is verified
against.** It is a script and not a paragraph for the same reason `scripts/probe-page-url.sh` is:
the answer is meaningless without its two controls.

```bash
scripts/audit-archived-still-serving.sh          # human table
scripts/audit-archived-still-serving.sh --json   # machine
```

### The two controls, and why NEITHER is optional

| control | must be | if it is not |
|---|---|---|
| an **invented** URL on the same domain | non-200 | the domain is a catch-all — *every* 200 is meaningless and every archived page would be flagged |
| a **known-good `active` + shipped sibling** page | 200 | the origin is down — every target reads "correctly absent" and the run is a **false all-clear** |

The second is the one specific to this question. For `asset_reference_404` a blinded run
under-reports; here it reports **zero**, and zero is the answer that looks healthy.

### ⚠ `kubectl exec -i` inside a `while read` loop eats the loop's own stdin

The first census printed one row and exited 0 — a plausible-looking answer. `psql` is reached
through `kubectl exec -i`, and `-i` consumed the rest of the here-string.

```bash
# WRONG — silently truncates after the first iteration, exit 0
while IFS='|' read -r a b; do
    sib=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql … -c "…")
done <<<"$ROWS"

# RIGHT — both halves
while IFS='|' read -r a b; do
    sib=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql … -c "…" </dev/null)
done < <(printf '%s\n' "$ROWS")
```

Cross-check the row count against `SELECT count(*)` before believing any loop of this shape.

## 2. The population, straight from the DB

```sql
SELECT s.domain, p.name, p.url, p.deployed_at::date
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status = 'archived' AND p.deployed_at IS NOT NULL
ORDER BY s.domain, p.url;
```

**Do not substitute `build_status`.** Archiving sets `status` and leaves the build columns
untouched (LANDMINES, "Archiving sets `status`…"), so the two axes are independent — which is
exactly why this bug exists. In Go use the shared helpers, never a hand-spelled filter:

```go
datahelpers.PageHasShippedPredicateFor("p")   // BUILD axis:     has it ever been served
datahelpers.PageWantedLivePredicateFor("p")   // LIFECYCLE axis: does the platform still want it
```

`pages.status` takes only `active` and `archived` fleet-wide. **`'deployed'` never occurs
there** — several live predicates spell `status IN ('active','deployed')` and the second
disjunct is dead.

## 3. Would a same-file `active` page make a 200 legitimate?

A 200 at a retired page's URL is not damage if an `active` page derives the SAME FILE. That is
the guard `retract_page_deployment_action.go` applies before it deletes anything, and a
detector that ignores it flags a page the platform is right to serve.

```sql
SELECT s.domain, p.url AS archived_url,
       (SELECT string_agg(a.name || ' [' || a.status || ']', ', ')
          FROM pages a
         WHERE a.site_id = p.site_id AND a.id <> p.id AND a.url = p.url) AS same_url_siblings
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status = 'archived' AND p.deployed_at IS NOT NULL;
```

⚠ `url` equality is the WEAK form. The real rule is equality of the **derived file path**
(`datahelpers.PageFilePathFromURL`), because `/foo/` and `/foo/index.html` are one file.
`[MEASURED 2026-08-26]` no serving archived page has a same-`url` sibling.

## 4. Is a check name live yet?

```sql
SELECT a.type, s.key, s.value->'config'->'checks'
FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(key, value)
WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
  AND s.value->'config' ? 'checks';
```

⚠ **The runner hard-fails on a check name the binary does not register.** The migration that
adds the name must be held back until the image carrying the Go file has rolled — name it
`_HOLD.sql` and apply it by hand after the roll (`SIDECAR_RE` excludes `_HOLD`, and still lists
it). Precedent: migration 368 for `site_unreachable`, stated in that file's own header.

## 5. Is the rotation actually driving the host agent?

```sql
SELECT name, interval_seconds, enabled, last_triggered_at
FROM scheduled_tasks WHERE name = 'site-discovery-rotation-availability';
```

`[MEASURED 2026-08-26]` enabled, 300 s, last fired 15:01Z. Its `pre_query` selects **one** site
whose `site_discovery_rotation.last_selected_at` is older than 4 h, so a site is swept at most
every 4 h and the whole fleet cycles in `300 s × sites`. A detector that has never flagged
anything may simply not have reached the site yet — check this before reading a zero.
