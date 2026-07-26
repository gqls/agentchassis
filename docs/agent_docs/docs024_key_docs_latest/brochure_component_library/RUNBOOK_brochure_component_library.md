# RUNBOOK — brochure component library / fundamentallyai.com

Every command here had a gotcha attached. The gotcha is the point — when one
changes, change it HERE, not in your scrollback.

Site: `fundamentallyai.com`, site_id `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`,
current plan_id `81741260-6447-492c-bf98-4b3c185f8e7b`.

DB access (the only route):
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

---

## Rebuilding a page (the whole minefield)

**Do not re-queue a historic work item.** `stale-work-item-reaper` (hourly) parks
any `triaged` build item whose **`created_at`** is 48h+ old — row age, not time
waiting (`bugs_open/070`). A five-day-old row is born eligible, so your
minutes-old request gets stamped `[stale: triaged 48h+]` and parked. Observed
here 12 minutes after a re-queue.

**Insert a fresh row instead.** `unresolved` is in `idx_swi_dedup`'s terminal set,
so the same `item_key` inserts cleanly beside the parked one.

```sql
INSERT INTO site_work_items (
  site_id, item_type, item_key, status, pipeline, summary, spec,
  handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
SELECT site_id, item_type, item_key, 'triaged', pipeline,
       'Rebuild <page>: <why, in words a stranger can act on>',
       spec, handler_agent, 'operator:<workstream>', 'operator:<workstream>',
       0, max_attempts, NOW(), NOW()
FROM site_work_items WHERE item_key='needs_page:<page>'
  AND site_id=(SELECT id FROM sites WHERE domain='fundamentallyai.com') LIMIT 1;
```

- **`created_by` is NOT NULL with no default** — a copy-INSERT that omits it dies
  with 23502. Name it.
- Also set `pages.build_status='needs_rebuild'` for a full rebuild.
- A hand INSERT **bypasses** the Go-side two-strike suppression in
  `insertWorkItem`. Deliberate; know you are doing it.
- With fresh rows, queueing several at once is safe — the reaper was the whole
  reason batches used to park. Verified 2026-07-25 (contact claimed in ~4 min
  alongside three siblings).

**A page with no plan sections cannot build**, and fails fast rather than
loudly: `plan_sections` → `check_has_ready_sections` → `mark_no_ready_sections`
in ~38s, no LLM spend, item → `needs_human_review`. Check first:

```sql
SELECT page_name, count(*) FROM site_plan_sections
WHERE plan_id='81741260-6447-492c-bf98-4b3c185f8e7b' GROUP BY 1 ORDER BY 1;
```

Placing sections (this is also what makes a component survive future rebuilds):

```sql
INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name) VALUES
 ('<plan>','<page>',0,'hero'), ('<plan>','<page>',1,'generic-text-block'), ...;
```
`idx_site_plan_sections_key` is UNIQUE(plan_id, page_name, ordering) — to insert
in the middle, shift high (+100), insert, bring down (−99). An in-place
`ordering+1` collides.

---

## Republishing a page after a DIRECT data edit (no content regeneration)

**DISPATCH LATENCY IS MINUTES, AND EVERY ROUTE LOOKS DEAD UNTIL IT ISN'T.** This
is the single most misleading thing in this runbook, so it comes first.

> **CORRECTED 2026-07-25 (same day I wrote the opposite).** I first recorded here
> that `049b_deploy_single_page.sh` "silently failed to ingest — four calls
> produced zero rows". **That was wrong.** All four landed:
> correlations `693bb6a7`, `ffd0856f`, `90767a97`, `7f2d33ad` all have
> orchestration rows with the right `page_id`, created **17:12–17:13** for
> dispatches fired around 17:05. I queried at ~17:07 and again at ~17:10, saw
> nothing, and concluded silent failure. It was **~7 minutes of latency.**
> The 086 route behaved the same way: dispatched ~17:11, row appeared **17:20:52**
> (~9 minutes). Neither script is unreliable. My patience was.

So: **budget ~10 minutes before concluding a dispatch failed**, and when you do
check, get the window right — I then wasted four more polls on a monitor filtering
`created_at > '18:00'` while the clock read 17:54, which cannot match anything and
reports exactly like a dead dispatch. Find the row by **payload**, and by a window
that starts *before* you dispatched:

```sql
SELECT status, current_step, created_at,
       initial_request_data->'config'->'workflow'->>'start_step' AS start_step
FROM orchestration_states
WHERE initial_request_data->'input_data'->>'page_id' = '<page_id>'
  AND created_at > '<a few minutes BEFORE you dispatched>'
ORDER BY created_at DESC LIMIT 3;
```
`start_step = 'spawn_rerender'` identifies an 086-script dispatch; a NULL
`start_step` with your correlation id is a 049b one. That is how to tell which of
your own attempts actually did the work — without it I credited the 086 script
for a republish the queued work item may well have done.

**All three routes work.** Prefer the **work-item queue** — it needs no Kafka
envelope and leaves a durable, inspectable row — unless the queue is backed up
(measured 98 `triaged` build items fleet-wide at 17:50 on 2026-07-25, which parks
an operator republish behind everyone else's work; that is when a direct dispatch
earns its keep):

```sql
INSERT INTO site_work_items (
  site_id, item_type, item_key, status, pipeline, summary, spec,
  handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
VALUES ((SELECT id FROM sites WHERE domain='fundamentallyai.com'),
  'page_rerender',
  'page_rerender_<page>_199733a8-ac9c-4c30-b2ce-65ecdac6f3bd_assemble_<tag>',
  'triaged','build','Republish <page>: <why>',
  jsonb_build_object('domain','fundamentallyai.com','page_id','<page_id>',
                     'filename','<path/index.html>','page_name','<page>'),
  'page-rerender','operator:<workstream>','operator:<workstream>',0,3,NOW(),NOW());
```

`item_type='page_rerender'` + handler `page-rerender` re-renders from **stored
`content_data`** (no LLM). Do NOT use `needs_page` for this — that routes into
the full LLM pipeline and rewrites all the copy (016b §9,
"`spec.reason` does not make `needs_page` scoped").

**Verify by payload, never by the printed correlation id:**
```sql
SELECT status, current_step, created_at FROM orchestration_states
WHERE initial_request_data->'input_data'->>'page_id'='<page_id>'
ORDER BY created_at DESC LIMIT 1;
```

---

## Internal-link census (the check the gate does but does not keep)

The site is **`.html`-based**: `/capabilities` 404s, `/capabilities.html` is 200.
Extension-less internal hrefs are broken, always.

**Use this form. Do NOT use `href="(/[^"#?]*)"`** — excluding `#` from the
character class silently drops every anchored href, which is how 21 broken links
passed a census, a fix and a post-check that all shared the pattern. Capture the
whole href, *then* strip the fragment:

```sql
WITH hrefs AS (
  SELECT p.name AS page, unnest(regexp_matches(pc.rendered_html, 'href="(/[^"]*)"', 'g')) AS href
  FROM pages p JOIN sites s ON s.id=p.site_id JOIN page_components pc ON pc.page_id=p.id
  WHERE s.domain='fundamentallyai.com'
), split AS (
  SELECT page, href, split_part(split_part(href,'#',1),'?',1) AS path FROM hrefs
)
SELECT path, count(*) AS n, string_agg(DISTINCT page, ', ') AS on_pages
FROM split
WHERE path <> '' AND path !~ '\.(css|js|png|jpg|jpeg|svg|ico|webp)$'
  AND NOT EXISTS (SELECT 1 FROM pages p2 JOIN sites s2 ON s2.id=p2.site_id
                  WHERE s2.domain='fundamentallyai.com' AND p2.url = split.path)
GROUP BY path ORDER BY n DESC;
```
Zero rows = every internal link resolves **in the database**. That is not the same
as on the live site: the served artefact can lag a data fix by however long the
republish takes, so confirm with a crawl of the served pages (below) before
reporting anything fixed. **Gotcha in my first version:** I
filtered the target lookup on `build_status='deployed'`, which mislabelled
`/contact` as an invented page when `/contact.html` was serving 200 — the page
row was `needs_rebuild` while the artefact was live. Don't filter on build_status
when asking "does this URL exist".

Recovering what the gate found on a build that **passed** (warnings are not
persisted — `bugs_open/071`; this dies with `collected_data` at ~24h):
```sql
SELECT jsonb_pretty(jsonb_agg(DISTINCT jsonb_build_object(
         'type', i->>'type', 'sev', i->>'severity', 'value', i->>'value')))
FROM orchestration_states os,
     jsonb_array_elements(os.collected_data->'validation_result'->'issues') i
WHERE os.orchestration_id::text LIKE '<oid-prefix>%';
```

Blocker/error detail on a build that **failed** (this one IS persisted):
```sql
SELECT occurred_at, jsonb_pretty(context) FROM agent_error_log
WHERE domain='fundamentallyai.com' AND error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'
ORDER BY occurred_at DESC LIMIT 1;
```
`agent_error_log` has **`occurred_at`, not `created_at`**.

---

## Editing rendered content safely

Replace **quoted exact** strings, not bare paths: `'"/contact"'` → `'"/contact.html"'`.
The leading `"/` stops `/contact` also matching `/contact.html` (→ `.html.html`)
and `/review-council` matching `/multi-agent-review-council`. Apply to
`rendered_html` **and** `content_data`, or the next re-render undoes it. Snapshot
first: `CREATE TABLE bak_pc_<slug>_<date> AS SELECT pc.* FROM page_components pc …`.

Full worked example: `sql/`-adjacent script used on 2026-07-25 for the link
repair (14 replacements, post-check included).

## Shell traps hit on this workstream

- **`grep -rn` goes silent-binary** when another session leaves a NUL byte in the
  tree — it found nothing for a string that was there. **Use `git grep`.**
- `printf` breaks on rendered HTML containing `%` and `)` — build SQL with
  `echo`/`cat` heredocs only.
- The Bash tool's working directory **persists between calls**. A relative `cd`
  that worked last call fails this call; `cd X && cat >> f << EOF` then silently
  writes nothing (the heredoc is consumed, the `&&` short-circuits). Use absolute
  paths for appends.
- Backticks inside `git commit -m` execute in bash — use `-F -` with a heredoc.

## Live crawl — the only independent witness

The database says what *should* be served; this says what *is*. Retry before
condemning: rapid cache-busted requests throttle the origin, and a throttled
request returns `000` or a spurious `404` that reads exactly like a broken link.

```bash
for pg in /index.html /capabilities.html /about.html /contact.html \
          /model-fine-tuning.html /multi-agent-review-council.html \
          /blog/self-correction-leopardessconsulting.html; do
  B=$(curl -s --max-time 25 "https://fundamentallyai.com${pg}?cb=$RANDOM")
  # any internal href WITHOUT .html is broken on this site, anchored or not
  n=$(echo "$B" | grep -oE 'href="/[a-z-]+(#[a-z-]+)?"' | grep -vE '\.html' | sort -u | tr '\n' ' ')
  echo "${pg}: ${n:-clean}"
done
```

For a full link check, follow every href three times before calling it broken,
and put `sleep 1` between requests. A tight loop over ~60 links without that
produced 21 false failures on 2026-07-25.

---

## The evidence-chart component (added 2026-07-26)

**Edit the source, never the SQL.** `register.sql` is generated, and a hand-edit
silently diverges from the template the harness tested:

```bash
B=docs/agent_docs/docs024_key_docs_latest/brochure_component_library
python3 $B/scripts/gen_component_register_sql.py $B/components/evidence-chart
```

**Validate the template before it goes near the DB**, for every page case,
including the ones that should render nothing:

```bash
for p in index capabilities ""; do
  go run $B/scripts/render_component_template.go \
      $B/components/evidence-chart/template.html \
      $B/components/evidence-chart/sample_data.json "$p"
done
```

The sample data deliberately contains the awkward cases: a chart belonging to
another page, a dangling `fact_id`, a zero value, a round million (which prints
as `1e+06` under Go's default float formatting and is invalid CSS), and a fact
nothing references. **Add a row to the sample for any case you add to the
template** — the one defect that reached the DB this session was a fact-sourced
denominator whose use site still read the old variable, and the harness caught it
only because the sample exercised that path.

**Check the register after any evidence_base edit** (eight defect checks, none of
which share the seed's logic; all should return zero rows):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db < $B/sql/evidence_base_charts_2026-07-26_VERIFY.sql
```

**Check the live pages against the register** (not against the template):

```bash
$B/scripts/verify_evidence_chart_live.sh fundamentallyai.com index capabilities
```

### Traps specific to this component

- **A charted fact's `value` must be a JSON number.** `html/template` CSS-filters
  a string under `printf "%.4f"` to `ZgotmplZ`, which kills the bar silently.
- **A SQL-sourced fact must not carry `display`.** `refresh_evidence_base`
  rewrites `value` and `verified_at` but never `display`, so it drifts.
- **A point label must contain one of its fact's `context_terms`**, or the claims
  gate reports our own charted figure as an unregistered number (it reads a
  ±70-character window, and block elements delimit it).
- **Never chart a `tolerance: gte` fact** — those say "state a FLOOR, never the
  exact number", and a bar states it exactly.
- **Text inside `<svg>` is invisible to the claims gate** (`claims.go:137`). This
  component keeps figures in HTML text for that reason; anything that later moves
  them into SVG has to replace that check.

## Queueing a rebuild: the dedup slot is occupied more often than you think

`idx_swi_dedup` is UNIQUE `(site_id, item_key)` WHERE status is **not** in
(`complete`, `verified`, `rejected`, `wont_fix`, `failed`, `unresolved`,
`cancelled`). **`needs_human_review` is not in that list**, so a page whose old
`needs_page` row was parked for review still holds the slot, and the RUNBOOK's
copy-INSERT recipe above dies on a unique violation.

Give the fresh row its **own key** — the handler reads `spec.page_name`, not the
key:

```sql
item_key = 'needs_page:<page>:<what-changed>-<yyyymmdd>'
```

**Measured 2026-07-26: queued 17:45:36, claimed 17:48:00 — 2m24s.** That is much
faster than the 7-9 minutes this runbook records above, because `bugs_open/030`
(cron sharing the dispatch lane) was closed the same day. The old figure is left
in place deliberately: it was true when written, and the lane can back up again.
Check the depth rather than assuming either number:

```bash
./scripts/dispatch-queue-depth.sh
```

**One page builds at a time in this lane.** Both rebuilds were queued together;
the second stayed `triaged` until the first finished. That is the known residual
of 030 (one job at a time per lane), not a dropped dispatch.
