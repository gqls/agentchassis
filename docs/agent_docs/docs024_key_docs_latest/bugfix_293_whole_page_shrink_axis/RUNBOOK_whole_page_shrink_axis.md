# RUNBOOK — bugs_open/293, the whole-page shrink floor's axis

Every command here was hard to get right once. The gotcha is attached to the command, not kept
in someone's scrollback.

## Re-run the calibration (this is the thing `section_visible_text.go` says to do and nobody could)

Three exports, then the harness. The harness lives in the repo —
`platform/orchestration/actions/shrink_axis_calibration_test.go` — and SKIPS unless
`SHRINK_CALIBRATION_JSONL` is set, so `go test ./...` is unaffected.

```bash
D=docs/agent_docs/docs024_key_docs_latest/bugfix_293_whole_page_shrink_axis
S=<scratch>                     # 11 MB of live page HTML — never in the repo

$D/export_pairs.sh        $S/pairs_delete_v2.jsonl    # 1,079 EXACT rebuild pairs
$D/export_overwrite.sh    $S/pairs_overwrite.jsonl    #   263 section-editor pairs (positive control)
$D/export_intermediate.sh $S/pairs_intermediate.jsonl # 2,454 weak-join pairs (refusal HUNTING only)
```

⚠ **Chunk the export; do not `json_agg` it.** A single 2 MB row dies mid-stream with
`unexpected EOF` and the harness then fails on a parse error that looks like a bug in the harness
(`NOTES_shared_template_write.md`, 2026-08-17). All three scripts page 10–20 rows at a time.
`export_intermediate.sh` takes ~10 min and needs backgrounding.

⚠ **The tree may not build** because other lanes have WIP in it. Test against committed HEAD:

```bash
H=$S/headtree; rm -rf $H; mkdir -p $H; git archive HEAD | tar -x -C $H
cd $H && SHRINK_CALIBRATION_JSONL=$S/pairs_delete_v2.jsonl \
  go test ./platform/orchestration/actions/ -run 'TestShrinkAxis|TestMinimumSweep|TestPageTotal' -count=1 -v
```

## The controls, and why each one is there

| test | what it would catch |
|---|---|
| `TestCalibrationAxisMatchesGuard` | the harness measuring a THIRD axis nobody ships — it asserts the calibration's tag-stripped measure IS `strippedIncomingBySlot`'s, and that the two axes actually disagree on a CSS-for-prose swap. Runs in the normal suite, no export needed |
| `TestShrinkAxisBlindness` controls | a hollower that quietly did nothing (asserts 0 visible chars left), sections with no prose counted as "protected" (excluded from the denominator), and an empty population (fails rather than reporting a comforting zero) |
| `TestMinimumSweep`'s tail assertion | the sweep's replicated arithmetic drifting from `evaluateSectionShrink` — it must reproduce the shipped decision exactly at the shipped minimum |

## The join, and how to check it is still exact

```sql
-- Disconfirming control: NO live row may be older than the last archived delete of its own
-- (page, slot). A wrong join does not produce a zero here.
WITH lastdel AS (
  SELECT DISTINCT ON (page_id, slot_name) page_id, slot_name, created_at AS del_at
  FROM page_component_history WHERE op='delete' AND slot_name IS NOT NULL
  ORDER BY page_id, slot_name, created_at DESC)
SELECT count(*) FILTER (WHERE pc.created_at < d.del_at) AS must_be_zero,
       count(*) FILTER (WHERE pc.created_at < d.del_at + interval '5 seconds') AS same_transaction
FROM lastdel d JOIN page_components pc USING (page_id, slot_name);
```

⚠ **Never join on `(page_id, slot_name)` alone.** Slot names repeat on 14 pages, so it is a
cartesian product and it manufactures pairs that never existed (it manufactured a refusal on the
first run of this calibration). Check before trusting any pairing:

```sql
SELECT count(*) FROM (SELECT page_id, slot_name, count(*) n FROM page_components
                      WHERE slot_name IS NOT NULL GROUP BY 1,2 HAVING count(*)>1) x;  -- 14
```

⚠ **A long gap means the guard never judged that pair.** On the intermediate export, attribute every
refusal by `gap_s` before calling it a false refusal: a rebuild deletes and re-inserts inside one
transaction (<5 s for 1,109 of 1,123), so a slot that was absent for 1,700 s to 93 hours was a DROP,
and the guard skips drops (`!present → continue`). 7 of the 8 intermediate refusals are this.

## Was the guard even live? (the censoring question, which decides what the population means)

```bash
git log -1 --format='%ad %s' --date=short 2da3e08e5   # per-slot shrink guard: 2026-08-02
```
```sql
SELECT min(created_at)::date FROM page_component_history WHERE op IS NOT NULL;  -- 2026-08-09
```
The guard predates the archive, so **every archived pair is a write the live guard ALLOWED** — which
makes the population right for measuring false refusals and useless for measuring true catches. That
is why the true-catch argument comes from the constructed wipe, not from history.

## Post-roll verification (artefact level, with a demand control)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system logs "$POD" --tail=3000 | grep -m1 'build provenance'
git merge-base --is-ancestor <this-lane's-commit> <the sha it printed> && echo "in the image"
```
⚠ The provenance line is a STARTUP line and scrolls out of reach on a busy service. Empty means
"not in range", not "unstamped" — fall back to the binary probe, and always with both controls:
```bash
kubectl -n ai-persona-system exec "$POD" -- grep -aq "<that-sha>" /proc/1/exe && echo present
kubectl -n ai-persona-system exec "$POD" -- grep -aq "$(git rev-parse HEAD~50)" /proc/1/exe \
  && echo "CONTROL FAILED — should be absent"
```
```sql
-- The refusal, when one happens: the message names the axis, so the row is greppable
SELECT created_at, spec->>'reason' FROM site_work_items
 WHERE item_type='needs_human_review' AND spec->>'reason' ILIKE '%VISIBLE text%'
 ORDER BY created_at DESC LIMIT 5;

-- DEMAND CONTROL — a zero above means nothing unless whole-page rebuilds are actually running
SELECT count(*) AS rebuild_writes, max(created_at) FROM page_component_history
 WHERE op='delete' AND created_at > now() - interval '24 hours';
```
