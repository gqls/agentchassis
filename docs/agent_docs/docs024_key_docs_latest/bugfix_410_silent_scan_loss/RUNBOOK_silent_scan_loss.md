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

## 12. Verify AFTER the roll — the fix is inert until the chassis ships it

Requested by the debug_historian seat (round `c8385154`): the recipe stated up front, with its
known failure modes, rather than left implicit. `loadStoredSections` builds into **agent-chassis**
— verify THAT service, not the fleet (one release can straddle commits and ship several revisions
under one tag).

**Step 1 — ancestry, the primary check.** Ask the pod what it was built from:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 7c443aac6 <the stamped sha> && echo SHIPPED
```

⚠ Test **ancestry, not equality** — the stamp is whatever HEAD was at build time, so `7c443aac6`
is normally an ancestor, not the stamp itself. ⚠ The provenance line is a STARTUP line and
scrolls: on a busy chassis it is out of `--tail` range within hours, and **an empty result means
"not in range", not "unstamped"** — fall back to step 2, which has no shelf life.

**Step 2 — capability probe, three-way, in ONE breath.** A probe whose every result is absent is
uninterpretable, not negative (LANDMINES, the `bugs_open/395` entry) — so all three, together:

```bash
POD=$(kubectl -n ai-persona-system get pod -l app=agent-chassis -o name | head -1)
# the thing under test — the new refusal string (a CAPABILITY literal, better than a sha:
# a sha says what was built, this says what can run):
kubectl -n ai-persona-system exec $POD -- grep -aq "refusing the partial result" /proc/1/exe && echo GUARD-PRESENT
# control that MUST be present in old and new binaries alike (proves the grep mechanism):
kubectl -n ai-persona-system exec $POD -- grep -aq "rerender_page_sections: row scan failed" /proc/1/exe && echo CONTROL-PRESENT
# control that MUST be absent (proves grep -aq can say no):
kubectl -n ai-persona-system exec $POD -- grep -aq "xq410zz-not-a-real-marker" /proc/1/exe || echo CONTROL-ABSENT
```

Expect all three lines. **Never `strings`** (absent from debian-slim; behind `2>/dev/null` its
failure is indistinguishable from "not stamped"). **Never ship a result whose controls are all on
the same side of the answer.** And the same-tag trap stands: a "fresh" roll at an unbumped
`IMAGE_TAG` serves the cached image — the probe above is the check that survives it.

**Step 3 — the guard's live silence is the success state.** Day-one firing rate is zero by
schema, so after the roll expect NO `ScanShortfall` errors anywhere. That zero is only meaningful
alongside step 2's GUARD-PRESENT — a zero from a binary that doesn't carry the guard is the
a-post-fix-zero-needs-a-demand-control trap. The demand control here is the mutation-proved test
suite, which fires the guard on every build.

## 13. Verifying the content-axis extension (2026-08-31) after its first roll

Same three-way form as §12, same controls — only the capability literal differs. The content-axis
refusal (`359503af0`, council `a69d82f2` APPROVED r1) is proven in a binary by:

```bash
kubectl -n ai-persona-system exec $POD -- grep -aq "content_data does not parse into a section object" /proc/1/exe && echo CONTENT-GUARD-PRESENT
```

§12's literal (`refusing the partial result`) verifies only the PARENT guard — both branches ride
it, but a binary can carry the parent without this extension (any build of `7c443aac6..359503af0`
does exactly that). Behaviour-neutral on today's data either way (0 non-object rows), so there is
no urgency to the roll; verify whenever the next one happens.
