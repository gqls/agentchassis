# HANDOFF — bugfix 236 (522 half) — COLD START, read this first

> **UPDATED 2026-08-11 10:30Z.** The 08-10 22:15Z version is superseded in three
> places, each marked below: the rotation has drained (twice), the Anthropic cap
> has **LIFTED** (so the council round is unblocked, not impossible), and the
> pool-site drill variant is **measured** — and cannot prove the self-clear half.

## Status in one paragraph

The fix is **BUILT, COMMITTED, LIVE AND PROVEN ON THE HEALTHY PATH AT FLEET
SCALE**. `check_site_unreachable` is in chassis **v1.0.1284** (re-greped after
the 08-11 09:23Z roll, both replicas, with controls), migration **372 is
APPLIED**, and the rotation has probed **all 22 eligible sites, twice, 0 items**.
**One thing is outstanding: the break-it-on-purpose drill**, which is the only
thing that proves the FILING and SELF-CLEAR paths in production. The council
round is now possible again and is owed. The bug file stays OPEN until the drill
runs.

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
  POSITIVE here, never silence. So `0 items` on serving sites positively rules out
  the egress failure mode. Don't re-test that.
- **The eligible set is 22, not 21.** 21 `deployed` + 1 `active`. The earlier
  "expect 21" was one short; both are correct counts of different things.

## What is live right now `[MEASURED 2026-08-11 09:26–10:30Z]`

| thing | state | proof |
|---|---|---|
| `check_site_unreachable` | in **v1.0.1284** | both running replicas (`7c9d5f74b9-6j5xn`,`-rvrdg`): `site_unreachable` 7, invented control `site_unreachabl3` **0**, pipeline control `asset_reference_404` 13 |
| migration 372 | APPLIED | agent row `checks=["site_unreachable"]`, task enabled 300s |
| rotation | **fully drained, twice** | 22/22 stamped, latest pass 06:07:46Z → 08:03:16Z today |
| findings | **0 items fleet-wide** | `site_work_items WHERE item_type='site_unreachable'` → 0 rows |
| Anthropic API | **RECOVERED ~22:00Z 08-10** | `llm_call_log`: 117 successes 22:00Z, 8 at 08:00Z 08-11; only failure in 14h is an unrelated ollama EOF |

Commits: `4a5d77004` (check + tests + IMP-053), `79feb08e7` (hold released, mig
372, roster fixture), `4864a8754` (delete the stale `_HOLD` copy), `6474f792a`
(notes).

## The two outstanding jobs

### 1. THE DRILL — the only outstanding proof. **NEEDS AN OWNER DECISION**

`0 findings` is also what a silently broken check reports (016b §9). The filing
path and the self-clear have **never run in production**. Both options are costed
in NOTES (2026-08-11 entry); the measurements are done, so this is a judgement
call, not more research.

**Option A — pool site + synthetic clear (recommended; no visitor impact).**
Two fixtures, because one cannot do both — see the trap below.
```sql
-- A1. FILE. All 17 pool domains are *.internal and do not resolve (curl: exit 6,
--     HTTP 000), so this is a TRUE finding on a domain no visitor can reach.
--     Pre-stamp the three CONTENT agents first: that is a hard WHERE exclusion
--     for 7 days (COALESCE(last_selected_at,'-infinity') < now() - 7 days),
--     not merely "sorts last".
INSERT INTO site_discovery_rotation (site_id, agent_type, last_selected_at)
SELECT id, a, now() FROM sites,
  unnest(ARRAY['quality-discovery-agent','design-discovery-agent','completeness-discovery-agent']) a
 WHERE domain='pool-web-tech.internal'
ON CONFLICT (site_id, agent_type) DO UPDATE SET last_selected_at=EXCLUDED.last_selected_at;
UPDATE sites SET status='active' WHERE domain='pool-web-tech.internal';
INSERT INTO site_discovery_rotation (site_id, agent_type, last_selected_at)
SELECT id,'availability-discovery-agent', now() - interval '30 days' FROM sites
 WHERE domain='pool-web-tech.internal'
ON CONFLICT (site_id, agent_type) DO UPDATE SET last_selected_at=EXCLUDED.last_selected_at;
-- wait ONE 300s tick. Expect exactly ONE item: item_type site_unreachable,
-- severity high, handler_agent '', status detected,
-- item_key site_unreachable:<site_id>, spec->>'reason' = 'transport_error'.
-- REVERT: UPDATE sites SET status='pool' ...; then cancel the item WITH PROVENANCE
-- (it can no longer self-clear — see the trap).

-- A2. CLEAR. Hand-insert an item of the exact shape on a HEALTHY real site and
--     watch the real check close it. AllOfType matches on item_type+site_id, so a
--     hand-inserted row is indistinguishable from a check-filed one to the
--     resolver. Force that site to the front, wait one tick, expect it resolved.
```
**THE TRAP, found 08-11 and NOT in the 08-10 version of this file:** `Run()`
returns early when `siteStatus != 'active'|'deployed'`. So reverting the pool site
to `pool` is exactly what STOPS the check probing it — the self-clear can never
fire on A1's item. That is why the clear needs its own fixture (A2).

**Option B — cookly.uk route deletion (the bug file's original protocol).**
Delete route `1e11858e5c1146229c3238351b394146` (`cookly.uk/*` →
`portfolio-sites-router`, zone `ab126cfa3debc8e1cf33fe8b741130bb`), force cookly
to the front, confirm the item, restore the route, confirm it self-clears.
- **What it adds over A, honestly: very little code coverage.** `judgeSiteProbe`
  sends `transport_error` and `http_522` down the **same** branch — the two drills
  differ by one string in `reason`. It adds a fact about *Cloudflare* (that a
  missing route yields non-2xx, not a 200 parking page) and proves file→clear
  chains on one site.
- **Cost:** a real site genuinely offline, ~1–6 minutes depending on tick timing.
- **`[UNMEASURED]`: route DELETE permission is unverified.** The token in
  `~/.cloudflare/404-token.env` reads zones and worker routes fine but **cannot**
  read DNS records, so its scope is workers-only. A 403 on the DELETE is harmless
  and is the cheapest feasibility test — but do not plan around it succeeding.

### 2. THE COUNCIL ROUND — owed, and now POSSIBLE (the cap lifted)

Submission `7177fb02-51c5-4c2a-bb02-10aa27ae85ca` selected its 10-seat panel,
persisted its `fix_plan`, then died at `review_editquality` on an upstream
Anthropic 400 usage-limit — terminal at `complete_invalid`, which is **NOT a
rejection**. `4a5d77004` carries `Council-Submitted: 7177fb02…`, which 098 will
never credit because that run reached no verdict.

**The cap ended ~22:00Z on 08-10, about seven hours after it began — NOT
2026-09-01, which was the API's assertion about a billing period, never an
observation.** Re-verify before spending, one query:
```sql
SELECT date_trunc('hour',created_at), success, count(*) FROM llm_call_log
 WHERE created_at > now() - interval '8 hours' GROUP BY 1,2 ORDER BY 1 DESC;
```
Then resubmit unchanged from `bugfix_236_site_availability/COUNCIL_SUBMISSION_2026-08-10.json`
with `RESUBMIT_CORR=7177fb02-51c5-4c2a-bb02-10aa27ae85ca` so the trail accumulates.
Budget ~30 minutes (dispatch queues behind the fleet).

## Where everything is

- Bug: `bugs_open/236_HANDOFF_2026-08-09_a_live_site_can_serve_522…`.
- Lane docs: `docs024_key_docs_latest/bugfix_236_site_availability/` — PLAN
  (decisions D1–D5), NOTES (the log, incl. six missteps and the 08-11
  blast-radius table), RUNBOOK (every query), README_where_we_are (owner-facing),
  COUNCIL_SUBMISSION json.
- Code: `platform/orchestration/actions/discovery_checks/check_site_unreachable{,_test}.go`.
  The test file header carries the **8-mutation table** — each guard broken and
  the NAMED test that caught it. Re-read that before changing any guard.
- Register: **IMP-053** in `docs026_concept_register/register/improvement-loop.md`.

## The one thing to be careful about

**Do not let this file's "LIVE" heading turn into "done".** What is live is
DETECTION OF AN OUTAGE THAT HAS NOT HAPPENED YET. The check has never fired on a
real failure. Until the drill runs, the honest sentence is: *the machinery is
installed and healthy-path-proven at fleet scale, and its ability to raise an
alarm is proven only in tests.* That is exactly the distinction `bugs_open/236`
was filed about — and two clean full-fleet passes make it MORE tempting to skip,
not less, which is the whole reason it is written here.
