# RUNBOOK — the unowned-bug queue

Commands that were hard to get right, with the gotcha attached.

## Is a bug actually unowned?

```bash
python3 scripts/who-owns.py <number>          # necessary, NOT sufficient
grep -rl "bugs_open/<number>" docs/ bugs_open/ # then read what each lane SAYS
```
⚠ `who-owns.py` reads commits, so it reports **who is citing the bug**, not who is fixing it. All
three bugs here returned `VERDICT: OWNED` and all three were unowned for the fix. Read the owning
lane's own words for what it **declined**.

## The 451 census — and the discriminator that actually works

```sql
-- the parked population, fleet-wide (NOT one item_key)
SELECT count(*), count(DISTINCT item_type), count(DISTINCT item_key), count(DISTINCT site_id)
  FROM site_work_items WHERE summary LIKE '[unresolved after%';

-- how many were parked by a SUCCESS (reconstructed; the ladder records no evidence)
WITH parked AS (SELECT id, site_id, item_key, created_at FROM site_work_items
                 WHERE summary LIKE '[unresolved after%')
SELECT count(*) FILTER (WHERE s.n_complete >= 1) AS had_a_completed_strike,
       count(*) FILTER (WHERE s.n_complete >= 2) AS both_strikes_complete, count(*) AS total
FROM parked p CROSS JOIN LATERAL (
  SELECT count(*) FILTER (WHERE w.status='complete') AS n_complete
    FROM site_work_items w
   WHERE w.site_id=p.site_id AND w.item_key=p.item_key AND w.id<>p.id
     AND w.status IN ('complete','failed')
     AND w.created_at < p.created_at AND w.created_at > p.created_at - interval '7 days') s;
```

⚠ **DO NOT use a time-gap census to decide whether a `complete` cleared the condition.** I did, it
returned a clean 0, and it could not have returned anything else for the mechanism under test — the
check runs daily, so persistence and drift both re-file once per tick. The discriminator is the
CONTENT of the drift:

```sql
-- same drifted keys twice = persistence (brake is right); different keys = new drift
SELECT site_id, created_at, spec->'drifted' FROM site_work_items
 WHERE item_key='stale_chrome' ORDER BY site_id, created_at;
```
Cross-reference with open `chrome_render_failed` items per site — `render_site_components` records
a *degraded success* there, and that is precisely a `complete` with the drift still present.

## 457 — the two censuses, and which one is right

```sql
-- WRONG for this defect: sweeps pages where different slots legitimately share a position
SELECT page_id, position, count(*) FROM page_components GROUP BY 1,2 HAVING count(*)>1;

-- RIGHT: the same slot duplicated
SELECT page_id, slot_name, position, count(*) FROM page_components
 WHERE component_id IS NULL GROUP BY 1,2,3 HAVING count(*)>1;
```
⚠ And a count on this population needs the **time of day**, not just the date — three censuses
inside 45 minutes gave 14/7, 14/7, 15/8.

## 433 — sizing a backfill without guessing

```sql
SELECT CASE WHEN url LIKE '/assets/images/%' THEN 'deployed' ELSE 'source only' END,
       count(*) FROM assets
 WHERE status='active' AND (mime_type IS NULL OR mime_type='') GROUP BY 1;
```
⚠ Then check the purposes against `storage.ImagePurposes`. `illustration`, `infographic` and an
empty purpose are **not keys**, so `GetImageConfig` falls through to `default` silently — a
purpose-derived backfill would guess on 40 rows. Enumerate; never `IN (SELECT …)`.

## Council

```bash
DRY_RUN=1 ./docs/agent_docs/.../097_TRIGGER_council_review_v1.sh <submission.json>   # free
./docs/agent_docs/.../097_TRIGGER_council_review_v1.sh <submission.json>             # save SUBMISSION_CORR
```
⚠ The **commit-msg hook rejects `Council-Submitted: pending`** — the trailer is a join key for the
098 report, so it must be a real correlation or absent entirely. Committing before you submit needs
no trailer at all.

## Committing on this tree

⚠ **Do not put backticks in a Go raw string**, including inside a `--` SQL comment you add to one.
The string ends at the first backtick and the package stops compiling.

⚠ `gofmt -l <dir>` lists **other sessions' unformatted WIP** too. Format only your own files by
name, never the directory.

⚠ Gate an append on the COUNT, not on reading the diff: `git diff --numstat <file>` must show
`N 0`. And re-run `git status` on the exact path immediately before committing — a file that was
dirty ten minutes ago may be clean now (v3_site_actions.go was), and one that was clean may not be.
