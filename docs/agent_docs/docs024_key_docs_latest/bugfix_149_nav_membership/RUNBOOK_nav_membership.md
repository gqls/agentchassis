# RUNBOOK — nav membership (bugs_open/149 Group A)

Every command here was needed to get something right, with its gotcha attached.
DB access throughout:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## R1 — the defect: a discovery-raised `nav_drift` item and what it named

```sql
SELECT s.domain, w.created_at::date, w.status, w.updated_at,
       w.spec->>'affected_pages' AS affected
FROM site_work_items w JOIN sites s ON s.id = w.site_id
WHERE w.item_type = 'nav_drift' AND w.source = 'discovery'
ORDER BY w.created_at;
```

**Gotcha: filter on `source='discovery'`.** 14 of the 17 `nav_drift` items in the
fleet's whole history were fired by hand by threads ("Rebuild site chrome:
sites.email cleared…"). Reading all 17 tells you nothing about the detector,
because most of them are humans using the item type as an action request. Only 3
came from a discovery agent, and those 3 are the evidence.

## R2 — did the completed item actually repair anything?

```sql
SELECT p.name, p.url, COALESCE(p.page_type,'content') AS page_type,
       COALESCE(p.in_header,false) AS ih, COALESCE(p.in_footer,false) AS if_,
       EXISTS (SELECT 1 FROM site_nav_items ni
                WHERE ni.site_id = p.site_id AND ni.status = 'active'
                  AND (ni.url = p.url OR ni.page_id = p.id)) AS in_nav
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND p.name IN ('bayesian-ranking','tool-drop-rate-tuner',
                 'tool-loot-table-balancer','tool-xp-curve-designer');
```

**Gotcha: match on `url = p.url` OR `page_id = p.id`, not one of them.**
`site_nav_items.page_id` is `ON DELETE SET NULL` (016_nav_tables.sql), so a nav
item can outlive its page with `page_id` NULL and only the URL to match on;
equally, hand-written rows exist whose URL shape differs from the page's. Checking
one column understates presence and would have made this defect look worse than it
is.

**The control that makes this a mechanism and not a cadence story** — same query,
`domain='robot-hands.com'`, `p.name IN ('learning-center','news')`. Both come back
`in_nav = t`. Same check, same handler, same action; the only difference is that
their URLs are not under a child prefix.

## R3 — the classifier replicated in SQL (for censuses)

```sql
CREATE TEMP VIEW cls AS
SELECT p.id, p.site_id, s.domain, p.name, p.url, p.status, p.build_status,
       COALESCE(p.page_type,'content') pt,
       COALESCE(p.in_header,false) ih, COALESCE(p.in_footer,false) if_,
       (p.url ILIKE '/tools/%' OR p.url ILIKE '/blog/%' OR p.url ILIKE '/guides/%'
        OR p.url ILIKE '/articles/%' OR p.url ILIKE '/case-studies/%'
        OR p.url ILIKE '/news/%' OR p.url ILIKE '/resources/%'
        OR p.url ILIKE '/insights/%') child_url,
       COALESCE(p.page_type,'content')
         IN ('blog-index','entity-directory','section-index','news-index') sect_idx,
       (lower(p.name) IN ('privacy','terms','cookies','disclaimer','privacy-policy',
                          'terms-of-service','cookie-policy','terms-and-conditions')
        OR lower(p.name) ~ '^(privacy|terms|cookie|disclaimer|legal)') is_legal,
       lower(p.name) IN ('404','sitemap','robots') is_system
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status IN ('active','deployed','pending');
```

**Gotcha, and it cost a wrong answer before it cost anything else: replicate the
Go by reading it top-to-bottom, BRANCH ORDER INCLUDED — not by listing its
predicates.** `classifyPagesForNav` tests `legalNames`/`isLegalPage` **before** it
looks at any flag, so a legal page needs no `in_header`/`in_footer` at all. My
first census modelled the flag test as universal and reported
`loancalculator.co.uk /legal.html` as a row the fix would stop reproducing. It is
reproduced, by the legal branch, and always was. Caught by re-reading the function
rather than by anything the query said — a replica that is wrong in the same
direction as your hypothesis returns a plausible number.

Same class: `status IN ('active','deployed','pending')` is `loadPagesForNav`'s
filter, not a guess. Read the loader, not just the classifier.

## R4 — regression surface: rows the derivation would stop reproducing

Run R3 first, then:

```sql
WITH nav AS (
  SELECT ni.id, ni.site_id, s.domain, ng.group_type, ni.label, ni.url, ni.page_id
  FROM site_nav_items ni
  JOIN site_nav_groups ng ON ng.id = ni.group_id
  JOIN sites s ON s.id = ni.site_id
  WHERE ni.status = 'active'
)
SELECT n.domain, n.group_type, n.url, c.name, c.ih, c.if_,
       (c.id IS NOT NULL AND NOT c.is_system
        AND (c.is_legal OR NOT (c.child_url AND NOT c.sect_idx))
        AND (c.is_legal OR c.ih OR c.if_))                       AS reproduced_before_fix,
       (c.id IS NOT NULL AND NOT c.is_system
        AND (c.is_legal OR c.ih OR c.if_))                       AS reproduced_after_fix
FROM nav n
LEFT JOIN cls c ON c.site_id = n.site_id AND (c.id = n.page_id OR c.url = n.url)
ORDER BY 1, 3;
```

Measured 2026-07-31: **7 rows `false` before, 1 `false` after.** The survivor is
leopardess `/tools/password-entropy.html` (`in_header=f in_footer=f`) — a nav row
asserting a membership its page does not declare, not reproducible either way, and
**not a regression introduced by this fix**. Repair is to set the flag, which is
that lane's call.

## R5 — additions: what appears on the next rebuild

```sql
SELECT c.domain, count(*) AS new_items,
       count(*) FILTER (WHERE c.build_status = 'deployed') AS would_ship_now,
       string_agg(c.name, ', ' ORDER BY c.name) AS pages
FROM cls c
WHERE c.child_url AND NOT c.sect_idx AND NOT c.is_system AND NOT c.is_legal
  AND (c.ih OR c.if_)
  AND NOT EXISTS (SELECT 1 FROM site_nav_items ni
                   WHERE ni.site_id = c.site_id AND ni.status = 'active'
                     AND (ni.url = c.url OR ni.page_id = c.id))
GROUP BY 1 ORDER BY 2 DESC;
```

26 items / 9 sites / ceiling 5 per site on 2026-07-31. **Re-measure, do not quote:
the page surface moves daily.**

## R6 — nav group sizes, to answer "will a footer flood?"

```sql
SELECT s.domain, ng.group_type, count(*) items
FROM site_nav_items ni JOIN site_nav_groups ng ON ng.id = ni.group_id
JOIN sites s ON s.id = ni.site_id
WHERE ni.status = 'active' GROUP BY 1,2 ORDER BY 1,2;
```

**Gotcha: do not try to do this per-site with correlated subqueries against
`sites`** — my first attempt filtered `sites.status='active'` and returned exactly
one row, because that is not how site liveness is recorded here. A plain
`GROUP BY` over the join answers it without inventing a filter. (Memory: a zero
from a filter you invented is not evidence.)

## R7 — is a nav item actually in the SERVED chrome?

```bash
curl -s "https://gamesdesign.co.uk/index.html?cb=$RANDOM" \
  | sed -n '/<header/,/<\/header>/p' | grep -oE 'href="[^"]*"'
```

**Gotcha, and it is the whole reason this fix ships no visible change: a nav row is
not a link.** Chrome is a stored artefact (`bugs_open/117`/`118`). gamesdesign and
fundamentallyai both hold `tools`/`primary` nav rows written by `addToolToNav`, and
neither appears in the served header — nothing has re-rendered chrome since they
were written. So "the row exists" and "the link ships" are different questions and
this command is the only one that answers the second.

Also: `grep -o -E '<nav[^>]*>.*</nav>'` returns **nothing** even when the nav is
there, because the markup spans lines. Use the `sed` range.

## R8 — nav-updater's real workflow (before assuming what the handler does)

```sql
SELECT key AS step, value->>'action' AS action
FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps')
WHERE a.type = 'nav-updater' AND a.is_active
  AND COALESCE(a.is_snapshot,false) = false AND a.deleted_at IS NULL;
```

`populate_nav_tables` → `render_site_components` → `create_rerender_items` →
`get_pages_for_rerender`. **This is why the fix is one predicate and not a new
route**: the handler already derives, re-renders chrome and propagates to every
deployed page. It only lacked a page to place.

## R9 — proving the test catches the defect, not just that it passes

```bash
# 1. temporarily restore the pre-fix predicate in classifyPagesForNav:
#      if isChildPageURL(page.URL) && !isSectionIndexType(page.PageType) { continue }
#      neverPrimary := neverPrimaryTypes[page.PageType]
go test ./platform/orchestration/actions/ -run TestChildPageURLCannotVetoADeclaredFlag
# expect: 4 FAILs of the form `in utility = false, want true ... utility=[]`
# 2. revert immediately, then:
go test ./platform/orchestration/actions/
```

**Do this, and keep the window to seconds.** A guard test that has never been seen
to fail is indistinguishable from one asserting nothing. And on this tree the
working copy is shared — `make build-*` builds from committed HEAD so a broken
working tree cannot reach an image, but another session running `go build` inside
that window sees your temporary breakage.

## R10 — build without touching another session's makefile

```bash
make build-agent-chassis IMAGE_TAG=v1.0.1215      # committed HEAD, tag overridden
```

**Gotcha:** `IMAGE_TAG` in the makefile is `?=`, so a command-line override works
and you do **not** have to edit the file. That matters here because the makefile is
routinely dirty with another session's tag bump (it was at `v1.0.1214`, uncommitted,
while HEAD said `v1.0.1206`). Editing it would put your change inside their
uncommitted diff.

**And the ordering gotcha that is easy to get wrong:** a chassis roll KILLS an
in-flight council run and an in-flight diagnosis run (both execute inside these
pods). Build freely — that touches nothing — but do not roll until your verdicts
have landed.

## R11 — verify the fix at the pod, with a control

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
        -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec -n ai-persona-system $POD -- sh -c '
  strings /app/agent-chassis | grep -c "declares no nav membership";
  strings /app/agent-chassis | grep -c "nav_membership_declared";
  strings /app/agent-chassis | grep -c "classifyPagesForNav: skipping child page"'
```

Expect `≥1`, `≥1`, **`0`** — the third is the string this fix REMOVED, and asserting
on a removal is what distinguishes "my image shipped" from "an image shipped".
Check **both** replicas: `logs deploy/X` and a single-pod exec each read one pod
of N.

## R12 — does a `triaged` work item actually get dispatched? (the council's medium objection)

The `bug_historian` seat's objection was that `RequestNavRebuild` moves the silent gap one
hop downstream: a work item sitting in `triaged` with nothing promoting it delivers nothing.
It is a measurement, not an argument. Three queries, and **re-run them rather than trusting
the answer below** — the whole point of the objection is that a lane can go quiet.

```sql
-- 1. the tightest answer: every item of the type this request uses
SELECT count(*) AS all_history,
       count(*) FILTER (WHERE claimed_by IS NOT NULL) AS claimed,
       count(*) FILTER (WHERE status = 'complete')    AS complete
FROM site_work_items WHERE item_type = 'nav_drift';

-- 2. is the lane that claims them alive, day by day?
SELECT created_at::date AS day, count(*) items,
       count(*) FILTER (WHERE claimed_by IS NOT NULL) claimed
FROM site_work_items
WHERE pipeline = 'build' AND created_at > NOW() - INTERVAL '7 days'
GROUP BY 1 ORDER BY 1 DESC;

-- 3. is the trigger enabled and firing?
SELECT name, target_agent_type, enabled, last_triggered_at
FROM scheduled_tasks WHERE target_agent_type ILIKE '%pipeline-trigger%';
```

Measured 2026-07-31: **`nav_drift` = 17 raised / 17 claimed / 17 complete**;
`build-pipeline-trigger` **enabled**, last fired `09:18:21Z`; the `build` lane claimed
**1,580 of 1,664** items over 7 days with claims on every day.

> **⚠ CORRECTED the same afternoon — those figures are true and the conclusion drawn
> from them was FALSE, so add query 4 and read it FIRST.** Claiming had stopped
> **fleet-wide at 13:21** that day, two hours before I quoted the numbers above; 9 items
> were stalled and my own item sat `triaged` and unclaimed for 15 minutes. A 7-day rate
> cannot answer "will this be claimed now", and a near-perfect 7-day rate is exactly
> what a lane that died two hours ago looks like.
>
> ```sql
> -- 4. THE LIVENESS QUESTION, asked directly
> SELECT max(claimed_at) AS newest_claim_anywhere,
>        round(EXTRACT(EPOCH FROM (NOW() - max(claimed_at)))/60) AS mins_since
> FROM site_work_items;
> -- and the loop's own terminal step, which is the real tell:
> SELECT max(updated_at) FROM orchestration_states WHERE current_step = 'complete_idle';
> ```
>
> **`scheduled_tasks.last_triggered_at`/`last_completed_at` advancing is a
> fire-and-forget stamp** — it proves the scheduler fired, never that a dispatch-loop
> orchestration was created. That trap is written down in
> `robot_hands_checker_gaps/NOTES_checker_gaps.md`, which I had read the same morning to
> find out what that lane owned. Logged in `WRONG_CALLS.md`. When the lane IS stopped,
> the bypass is `TRIGGER_nav_rebuild.sh` in this directory.

**The gotcha, and it is the reason the objection was worth answering rather than waving
away: `detected` and `triaged` are two different queues and the fleet's famous
"263 detected / 0 triaged" figure is about the FIRST one.** Nothing promotes `detected` →
`triaged` on a schedule (that is `triage_detected_items`, run by three agents, none with an
enabled scheduled task — the `robot_hands_checker_gaps` lane's finding). The
`triaged` → claimed → complete lane is a different mechanism and is demonstrably alive.
`RequestNavRebuild` is born `triaged` precisely to start on the working side of that seam.

**Do not read query 1's `claimed` column off a status snapshot instead.** A grouped
`count(*) FILTER (WHERE claimed_by IS NOT NULL)` per *status* shows `triaged | 98 | 0`,
which reads as "nothing claims triaged items" and is an artefact: a claimed item has by then
MOVED to `complete`, so the zero is definitional. Count by `item_type` or by `pipeline` over
a window, never by current status.

## R13 — do not roll while another session's council is mid-round

```sql
SELECT left(correlation_id::text,8) AS corr, current_step, updated_at,
       round(EXTRACT(EPOCH FROM (NOW()-updated_at))/60) AS mins_idle
FROM orchestration_states
WHERE status IN ('EXECUTING_STEP','AWAITING_RESPONSES')
  AND updated_at > NOW() - INTERVAL '60 minutes'
ORDER BY updated_at DESC;
```

A chassis roll kills these — they execute inside those pods, and a council round is 4–9
minutes of LLM calls with no resume. Two other sessions had rounds at `review_bug_historian`
when this change was ready to ship on 2026-07-31, so the roll waited. A `current_step` of
`review_*` is somebody's council; `mins_idle` under ~5 means it is live, not stale.
**Build immediately (it touches nothing); roll on a quiet queue.**
