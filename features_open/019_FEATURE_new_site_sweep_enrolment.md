# 019 FEATURE — auto-enrol new sites into the discovery/audit cadence

**Raised:** 2026-07-24, from the fundamentallyai finding: the site had **zero**
discovery-check work items while every other live site has dozens-to-hundreds
(robot-hands 314, leopardess 185, finetuning 147) — the ~40-check sweep had
simply never been pointed at the new site.
**Priority: LAST of the site-quality automation set (owner ruling 2026-07-24):
the improvement loop is not currently running, so enrolment would feed a queue
nothing drains — build 016/017/018 first; wire enrolment when the loop is back.**
**Status:** specified, deferred.

## The gap

Discovery sweeps (`run_discovery_checks` via design-discovery-agent) and the
audit agents run against sites they are explicitly pointed at (scripts,
scheduled tasks with fixed site lists, hand triggers). Site onboarding
(`domain-submitter` → … → build) never registers the new site with any sweep
cadence. Every new site starts invisible to the immune system and stays
invisible until a human remembers.

## What it is (when picked up)

1. On first successful build completion (the `reconcile_site_plan` /
   first-deploy milestone), automatically create the site's sweep enrolment —
   whatever form the cadence uses (a scheduled_tasks row, a recurring work
   item, or membership in the fleet sweep's site list).
2. Backfill: enrol the currently-unenrolled live sites (at minimum
   fundamentallyai.com; audit the full list at build time — do not trust this
   file's snapshot).
3. Sequencing guard: enrolment lands together with (or after) the improvement
   loop being back on — an enrolled site whose findings nobody drains is
   work-item noise (and the dedup index means stale findings can block fresh
   ones).

## Evidence query (re-run at build time)

```sql
SELECT s.domain, count(wi.id) FILTER (WHERE wi.source ILIKE '%discovery%'
  OR wi.item_type IN ('broken_nav_link','voice_tells','phantom_link','dead_control')) AS findings
FROM sites s LEFT JOIN site_work_items wi ON wi.site_id=s.id
WHERE s.status='deployed' GROUP BY s.domain ORDER BY findings;
```
