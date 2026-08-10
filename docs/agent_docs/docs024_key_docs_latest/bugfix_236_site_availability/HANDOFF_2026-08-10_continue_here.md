# HANDOFF — bugfix 236 (522 half), 2026-08-10 22:15Z — COLD START, read this first

## Status in one paragraph

The fix is **BUILT, COMMITTED, LIVE AND RUNNING**. `check_site_unreachable` ships
in chassis **v1.0.1283** (pod-verified, both replicas, with a negative control),
migration **372 is APPLIED**, and the rotation has already completed its first
real probe. Three things are outstanding, in priority order: **(1) the
break-it-on-purpose drill**, which is the only thing that proves the FILING path
in production; **(2) the council round**, which died on an upstream API cap and
must be resubmitted; **(3) let the cold-start rotation finish** and read the
result. The bug file stays OPEN until (1) is done.

## Do not re-derive these

- **The bug is by SLUG, not number.** `bugs_open/236_HANDOFF_2026-08-09_a_live_site_can_serve_522…`.
  The OTHER 236 (`hero_and_logo_deployed_lose_image_url`) is a different case.
  Same trap on 243: the LLM-cap bug is `243_…anthropic_account_usage_limit…`, not
  the tool-acceptance one.
- **This lane owns candidate 1 only.** Candidate 2 (zone/route conformance) stays
  with `domains_cloudflare_rollout`. Candidate 3 is not taken.
- **Migration 372's number collides with another lane's applied migration**
  (`372_provocation_generator_token_budget.sql`). **Do NOT renumber.** Both are
  recorded in `schema_migrations` by FILENAME, both applied cleanly. Renumbering
  an applied migration makes it pending and re-runs it.
- **The healthy-path zero is not a blind zero.** If pod egress were blocked, the
  probe would transport-error and FILE an item — a blocked network is a false
  POSITIVE here, never silence. So the first run's `0 items` on a serving site
  positively rules out the egress failure mode. Don't re-test that.

## What is live right now

| thing | state | proof |
|---|---|---|
| `check_site_unreachable` | in the binary | both replicas of v1.0.1283: `site_unreachable` 7, `+site_unreachable` 1, invented control `site_unreachabl3` **0**, pipeline control `asset_reference_404` 13 |
| migration 372 | APPLIED | `schema_migrations`; agent row `checks=["site_unreachable"]`, task enabled 300s |
| `availability-discovery-agent` | live, no LLM steps | orchestration `4ec82e1c` COMPLETED 22:03:49Z |
| `site-discovery-rotation-availability` | ticking | `last_triggered_at` 22:03:47Z, 1 site stamped |
| first probe | clean | robot-hands.com: `checks_run:[site_unreachable]`, `checks_failed:[]`, 0 items |

Commits: `4a5d77004` (check + tests + IMP-053), `79feb08e7` (hold released, mig
372, roster fixture), `4864a8754` (delete the stale `_HOLD` copy), `6474f792a`
(notes). Plus contributions to the 230 lane and the webdesign.uk chat lane.

## The three outstanding jobs

### 1. THE DRILL — the only outstanding proof (do this first)

`0 findings` is also what a silently broken check reports (016b §9). The filing
path and the self-clear have **never run in production**. The bug file's own
protocol:

> delete `cookly.uk/*`'s worker route, confirm the checker files a work item
> within its interval, restore the route, confirm the item closes.

**Before doing that, consider the safer variant** — it was designed but not run
because the fleet had no non-serving in-scope site: **induce on a POOL site**
(pool domains are unrouted by design, so a finding there is TRUE, and nobody
visits them). Sketch, modelled on IMP-051's induce-and-revert:

```sql
-- 1. pick a pool site whose domain genuinely does not serve (CHECK with curl first)
-- 2. flip it into scope and force it to the front of the rotation
UPDATE sites SET status='active' WHERE domain='<pool domain>';
UPDATE site_discovery_rotation SET last_selected_at = now() - interval '30 days'
 WHERE agent_type='availability-discovery-agent'
   AND site_id=(SELECT id FROM sites WHERE domain='<pool domain>');
-- 3. wait one 300s tick; expect ONE site_unreachable item, handler_agent='',
--    severity high, item_key site_unreachable:<site_id>
-- 4. REVERT: UPDATE sites SET status='pool' ...; cancel the item with provenance
```

**RISK to weigh before choosing this variant, and it is the reason it was not
just run:** the three CONTENT rotations also select `status IN ('active','deployed')`
and order by `last_selected_at NULLS FIRST` — an unstamped site sorts FIRST, so
flipping a pool site to `active` can pull quality/design/completeness discovery
onto it within the hour. Mitigate by stamping all four agent types for that site
before flipping, or accept the cookly route-deletion drill instead (shorter blast
radius in time, larger in visibility — it takes a real site down for ~1 minute).
**This is a judgement call for the next session; both options are costed here.**

### 2. THE COUNCIL ROUND — owed, currently impossible

Submission `7177fb02-51c5-4c2a-bb02-10aa27ae85ca` selected its 10-seat panel,
persisted its `fix_plan`, then died at `review_editquality` on an upstream
Anthropic **400 usage-limit** — terminal at `complete_invalid`, which is **NOT a
rejection**. The commit `4a5d77004` carries `Council-Submitted: 7177fb02…`, which
098 will never credit because that run reached no verdict. **A FRESH submission
is owed once `bugs_open/243` clears** (API says 2026-09-01, but it is an owner
billing action and may lift sooner — check before assuming).

The submission JSON is ready to re-fire unchanged:
`bugfix_236_site_availability/COUNCIL_SUBMISSION_2026-08-10.json`. Resubmit with
`RESUBMIT_CORR=7177fb02-51c5-4c2a-bb02-10aa27ae85ca` so the trail accumulates.
**Check the LLM state first, one query — do not spend a round into a wall:**

```sql
SELECT date_trunc('hour',created_at), success, count(*) FROM llm_call_log
 WHERE created_at > now() - interval '8 hours' GROUP BY 1,2 ORDER BY 1 DESC;
-- a zero in the `t` column is the finding; the `f` rows only prove the fleet tried
```

### 3. LET THE ROTATION DRAIN, THEN READ IT

`LIMIT 1` per 300s tick ⇒ 21 sites in ~105 minutes from 22:03Z, i.e. **complete
by ~23:50Z on 2026-08-10**, then it settles to the 4-hour cooldown. Read it:

```sql
SELECT count(*) AS stamped FROM site_discovery_rotation
 WHERE agent_type='availability-discovery-agent';            -- expect 21
SELECT s.domain, wi.summary, wi.status FROM site_work_items wi
  JOIN sites s ON s.id=wi.site_id WHERE wi.item_type='site_unreachable';
```

**Expected: 21 stamped, 0 items** (all 21 sites served 200 when measured at
15:45Z and again in the first probe). **If an item DOES appear, it is probably
REAL** — read the `reason` in its spec (`http_522` / `transport_error` /
`empty_body` / `not_html`) and check the site by hand before assuming a false
positive. Also worth a look: any `title_absent` or `delegated` FINDINGS in the
orchestration `collected_data` — those are deliberately unfiled, and
`mortgagecalculator.co.uk` is a known `title_absent` (a divergent render, someone
else's defect).

## Where everything is

- Bug: `bugs_open/236_HANDOFF_2026-08-09_a_live_site_can_serve_522…` (status
  banner at the top is current as of the commit before this handoff).
- Lane docs: `docs024_key_docs_latest/bugfix_236_site_availability/` —
  PLAN (decisions D1–D5), NOTES (the log, incl. four missteps), RUNBOOK (every
  query), README_where_we_are (owner-facing prose), COUNCIL_SUBMISSION json.
- Code: `platform/orchestration/actions/discovery_checks/check_site_unreachable{,_test}.go`.
  The test file header carries the **8-mutation table** — each guard broken and
  the NAMED test that caught it. Re-read that before changing any guard.
- Register: **IMP-053** in `docs026_concept_register/register/improvement-loop.md`
  (three LANDMINEs in it, incl. "this check finds NOTHING today, and that is
  measured, not assumed").
- Fleet lessons logged this session: `WRONG_CALLS.md` ×2 (migration-number
  collisions; misreading `complete_invalid` when the landmine already existed) and
  a recurrence appended to the `complete_invalid`/usage-cap entry in `LANDMINES.md`.

## The one thing to be careful about

**Do not let this file's "LIVE" heading turn into "done".** What is live is
DETECTION OF AN OUTAGE THAT HAS NOT HAPPENED YET. The check has never fired on a
real 522. Until the drill in §1 runs, the honest sentence is: *the machinery is
installed and healthy-path-proven, and its ability to raise an alarm is proven
only in tests.* That is exactly the distinction `bugs_open/236` was filed about.
