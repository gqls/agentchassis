# RUNBOOK — bugfix 410, silent scan loss

Every command/query this lane had to get right, with its gotcha attached. When one changes,
change it **here**, not in scrollback.

---

## 1. Is the bug still there? (cite by SYMBOL, never by line)

```bash
grep -n "rerender_page_sections: row scan failed" \
  platform/orchestration/actions/rerender_page_sections_action.go
```

⚠ **Do not cite a line number for anything in this file.** 410's own corrections block records
`:1206` expiring the same afternoon it was written — `bd811fa93` moved the Warn ~32 lines. The
`news_editorial` lane moved several test lines again the same day. The distinctive string is
stable; the line is not.

To see the whole loader:

```bash
awk '/^func loadStoredSections/,/^}/' \
  platform/orchestration/actions/rerender_page_sections_action.go
```

## 2. Does the caller propagate the error, or re-swallow it?

The single most important check before writing any guard — a guard whose error the caller
discards is the bug reintroduced as its own fix.

```bash
grep -n "loadStoredSections(ctx" -A 3 \
  platform/orchestration/actions/rerender_page_sections_action.go
```

Today: `return nil, fmt.Errorf("load stored sections: %w", err)`. It propagates.

## 3. The class census — how many places swallow a scan error

```bash
python3 <scratchpad>/scancensus2.py
# SWALLOW (scan err -> continue): 225 production
# PROPAGATE (scan err -> return/break): 491
```

⚠ **The obvious version of this script is wrong and lands within 5% of the right answer.** A
fixed lookahead (`is there a `continue` within N lines of a `.Scan(``) also matches every loop
that legitimately filters *after* a successful scan — a different population that overlaps
heavily. It returned **237** against the correct **225**, and no amount of looking at the number
separates them. The script must **brace-match the `if err != nil {` block that wraps the Scan**
and ask what is inside *that*.

## 4. Blast radius — measure it AT THE SCAN, never at the work item

This is the reviewer's hardest question and it has a real answer. **A `page_rerender` item count
does not size it** (different population, different axis — the `bugs_open/384` lane's point, and
they are right).

Two steps. First, what can be NULL:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "\d page_components" | head -25
```

Then, does anything currently violate it:

```sql
SELECT count(*) AS rows_live,
       count(*) FILTER (WHERE position IS NULL)     AS null_position,
       count(*) FILTER (WHERE content_data IS NULL) AS null_content_data
FROM page_components WHERE build_status IS DISTINCT FROM 'removed';
-- 2026-08-26: 2194 | 0 | 54
```

The argument is the **column-by-column table** (NOT NULL / COALESCE'd / NULL-safe `[]byte`), not
the row count — the query is only its control. `position integer NOT NULL` and `id uuid NOT NULL`
are the load-bearing facts; `content_data` is NULL on 54 rows and scans to nil harmlessly.

⚠ **This expires the moment the projection or the schema changes.** A nullable column added to
that SELECT without a `COALESCE` falsifies it. Re-run; do not quote the stored table.

## 5. Counting work items — THE ROLLING WINDOW TRAP

⚠⚠ **`site_work_items` archives finished rows into `site_work_items_archive`.** Querying only the
live table gives you a slice and it reads as the population. Two lanes made this mistake on the
same figure on the same day and agreed **to the digit**.

**Always UNION both tables:**

```sql
SELECT 'live' AS src, count(*) AS items,
       count(*) FILTER (WHERE spec ? 'reason') AS with_reason
  FROM site_work_items
 WHERE item_type='page_rerender'
   AND created_by IN ('rerender-pages','create_rerender_items')
UNION ALL
SELECT 'archive', count(*), count(*) FILTER (WHERE spec ? 'reason')
  FROM site_work_items_archive
 WHERE item_type='page_rerender'
   AND created_by IN ('rerender-pages','create_rerender_items');
-- 2026-08-26:  live 6,428/3   archive 10,857/200   TOTAL 17,285/203
```

`site_work_items_archive` is **the only `*_archive` table in the schema** (checked by the
`news_editorial` lane 2026-08-26) — so this trap applies to work items and nothing else.
`page_components` has no archive twin; its counts are whole populations.

**And the meta-rule this cost us:** re-running someone else's query tests the *arithmetic*, not
the *choice of table*. Hand over the **population definition** — which tables, which window, what
is excluded — not just the SQL. On receipt, restate the population in your own words before
running anything.

## 6. Do NOT probe `site_work_items` for a column the producer never writes

```bash
awk '/INSERT INTO site_work_items/,/\$[0-9]+\)/' \
  platform/orchestration/actions/create_rerender_items_action.go
```

`create_rerender_items` inserts `site_id, source, pipeline, item_type, severity, summary,
page_id, priority, handler_agent, status, created_by, spec, item_key, batch_id` — **no
`component_id`**, even though `site_work_items.component_id` exists as a real column. A zero there
is guaranteed by construction and is evidence of nothing. Read the INSERT before reading a zero
as a finding.

## 7. Green baseline before touching the package

```bash
go test ./platform/orchestration/actions/ \
  -run 'TestRerenderPageSections|TestAnUnreadableSection|TestNoDynamicallyConstructedItemTypes' \
  -count=1
```

`-count=1` defeats the test cache — without it a "pass" may be a cached result from before your
edit.

## 8. The seven tests that went red on the motivating change

Cite **by name**; they moved twice on 2026-08-26.

- `TestRerenderPageSections_SuccessEntryCarriesTheStoredSlotName` (`save_sections_stored_slot_identity_test.go`)
- `TestRerenderPageSections_FailsWhenComponentUnresolvedByNameOrID`
- `TestRerenderPageSections_ResolvesToolByComponentIDWithoutEscalating`
- `TestRerenderPageSections_ComponentIDWinsOverNameWhenBothResolve`
- `TestRerenderPageSections_InvalidTemplateByID_IsFatalAndNamed`
- `TestRerenderPageSections_EmptyTemplateCarriesWithoutFailing`
- `TestRerenderPageSections_StructuralCarryMakesANotReadySectionRerender`

(last six in `rerender_page_sections_resolve_test.go`)

⚠ **It is SEVEN, not six.** The lane that ran it piped through `tail -6`, fixed the one failure it
showed, re-ran, saw six more, and reported the second run's count — a count formed from a
truncated view, in a bug about instruments that return less than they were given.

## 9. The three precedents to cite at the council

```bash
# 1. closest — same guard, same rows.Scan shape, already live, GRADED
awk '/^func scanBlogArticles/,/^}/' platform/orchestration/actions/rebuild_blog_listing_action.go

# 2. the count guard 410 cites, with its mutation check
sed -n '140,175p' platform/orchestration/actions/validate_page_content_surface_test.go

# 3. the estate's ratchet idiom + the "two layers, one pattern" rule
head -55 platform/orchestration/actions/work_item_type_minting_ratchet_test.go
```

`scanBlogArticles`' graded response was **forced by a gating council objection** (`170147b4`,
bug_historian) — the first cut logged-and-skipped unconditionally. That provenance is the
strongest line available at the gate: it is not a new mechanism, it is one a seat already
required on this exact shape.

## 10. Council gate

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

Save the printed `SUBMISSION_CORR` **into NOTES, not scrollback** — a correlation in a transcript
dies with the session. Budget ~30 minutes, not ~2; a missing orchestration row is latency, not a
dropped dispatch. Find the run by payload:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

Verdict:

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report' ORDER BY created_at;
```

## 11. Committing on a shared tree

```bash
git diff --numstat <file>          # gate on the COUNT before reading lines
git diff HEAD --numstat <file>     # the check that cannot be fooled by a grep pattern
git commit <explicit paths> -m "..."
```

⚠ Another lane is editing a **different region of the same file** (`RerenderPageSectionsAction`'s
per-row loop, ~462-770, vs the loader at ~1199-1250). Different regions is **not** protection from
a same-file passenger — whoever commits second takes both edits. Read `git diff` on the file
immediately before committing. Three passenger events happened on `WRONG_CALLS.md` alone on
2026-08-26, in three different directions; on each, verifying **at HEAD** rather than trusting the
command's own report was the check that settled it.
