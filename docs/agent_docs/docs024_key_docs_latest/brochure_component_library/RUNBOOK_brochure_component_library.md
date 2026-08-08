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
  site_id, item_type, item_key, status, severity, priority, summary, spec,
  handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
VALUES ((SELECT id FROM sites WHERE domain='fundamentallyai.com'),
  'page_rerender',
  'page_rerender:<page>:<tag>',
  'triaged','medium',30,'Republish <page>: <why>',
  jsonb_build_object('domain','fundamentallyai.com','page_id','<page_id>',
                     'filename','<path/index.html>','page_name','<page>'),
  'page-rerender','operator:<workstream>','operator:<workstream>',0,3,NOW(),NOW());
```

> **CORRECTED 2026-08-03 (twice in one morning).** (1) The column list above had
> drifted from the live schema twice — first `category`, then `pipeline`; the
> table has neither. The form above (severity + priority, item_key in the
> `page_rerender:<page>:<tag>` shape) is copied from a live row and inserted
> successfully three times today. (2) **The claim that this queue item
> "re-renders from stored content_data" is WRONG as written — with no `reason`
> in the spec it is ASSEMBLE-ONLY.** Measured: a 10:58 `content_data` edit, the
> queue item complete at 11:00:33, and the section's `rendered_html` still
> carried the OLD markup until the direct script ran at 11:04:39. A completed
> item + a deployed page that still serves your old content is exactly what an
> assemble-only run of stale sections looks like. For a content_data or
> template change, use `scripts/rerender_page_sections_direct.sh` (verified,
> §routes table below) — or add `'reason','section_data_resolved'` to the spec
> jsonb, which the agent wires into the rerender_sections pre-pass
> [UNVERIFIED — the direct script is the proven path].

Do NOT use `needs_page` for a republish — that routes into
the full LLM pipeline and rewrites all the copy (016b §9,
"`spec.reason` does not make `needs_page` scoped"). `needs_page` IS the right
route for a genuinely new planned page (used 2026-08-03 for the two guide
builds: spec `{"reason":"not_built","page_name":"<name>"}`, handler
`page-build-handler`, mirroring the completed `needs_page:capabilities` item).

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

## Re-rendering a page after a TEMPLATE or content_data change (no LLM)

Two different things are called "re-render" and only one of them picks up a
component template change:

| what you changed | what you need | how |
|---|---|---|
| nothing (just republish) | assemble-only: stitches STORED rendered_html | `049b_deploy_single_page.sh <page_id> <site_id> <domain>` |
| a component's `html_template`, or a section's `content_data` | the `rerender_sections` pre-pass: every section re-rendered from stored `content_data` through the CURRENT template, no LLM | `scripts/rerender_page_sections_direct.sh <page_id> <site_id> <domain> <page_name>` |

**`049b` with its 4th argument does not work** — it builds `spec` as
`{"reason": …}` and the agent reads `page_name` from `input_data.spec.page_name`,
which `rerender_page_sections` declares Required. Every such dispatch fails in
under a second with `missing required fields: [page_name]`. Confirmed on two of
our dispatches and two from another session on a different site. The script above
is the corrected form; the 049b owner has a note with the one-line fix.

**Before firing the reason path, check for NULL `content_data`** — if any section
has none, the whole page escalates to the content writer and the copy IS
regenerated:

```sql
SELECT slot_name, content_data IS NULL AS null_cd FROM page_components WHERE page_id='<id>';
```

**The work-item route was NOT usable on 2026-07-26**, and the reason is not
established: `build-pipeline-trigger` selected fundamentallyai.com on three
consecutive runs (18:12, 18:15, 18:17), reached `spawn_dispatch`, sat in
`AWAITING_RESPONSES`, and the two `page_rerender` rows stayed `triaged` for 15+
minutes with no child orchestration. Only 3 build items were queued fleet-wide and
our site had the lowest `site_id`, which is what that trigger's
`ORDER BY wi.site_id … LIMIT 1` selects on — so it was being picked, not starved.
[UNDIAGNOSED] Recorded because a future thread will otherwise read the silence as
"the queue is backed up", which the depth said it was not.

## Counting em-dashes so the number means something (added 2026-07-27)

The obvious query is wrong, and it was wrong in the handoff for a day. Counting `—`
in `page_components.rendered_html` sums **two different populations**: what the
content LLM wrote, and literals baked into the component's `html_template`. No
writer-prompt round and no `content_data` post-pass can move the second kind, so a
page whose count "did not improve" may have improved entirely and be pinned by
template text.

Always split by origin:

```sql
WITH pc AS (
  SELECT p.name AS page,
         length(pc.rendered_html)  - length(replace(pc.rendered_html,'—',''))  AS em_rendered,
         length(cc.html_template)  - length(replace(cc.html_template,'—',''))  AS em_template
  FROM pages p JOIN sites s ON s.id = p.site_id
       JOIN page_components pc ON pc.page_id = p.id
       JOIN content_components cc ON cc.id = pc.component_id
  WHERE s.domain = 'fundamentallyai.com'
)
SELECT page, sum(em_rendered) AS total,
       sum(em_template)                AS from_template,
       sum(em_rendered - em_template)  AS from_words
FROM pc GROUP BY page
UNION ALL SELECT 'TOTAL', sum(em_rendered), sum(em_template), sum(em_rendered - em_template) FROM pc
ORDER BY 4 DESC;
```

`from_words` is the only column a writer-prompt change can move. **Do not** attribute
by comparing `em_rendered` to `content_data` alone: a component can carry both, and
`content_data` also holds resolved data (chart facts, query results) whose punctuation
is not the writer's either.

Gotcha: a `<style>` comment inside a component template counts, and *ships* — it is an
HTML comment, not a Go template comment, so it costs bytes on every render and shows up
in every text metric taken from `rendered_html`.

## Council gate: reading a REVISE properly (added 2026-07-27)

`metadata` on the `council_report` artifact has only four keys — `decision`,
`abstained`, `reviewers`, `unreadable`. **The objections are in `body`, not
`metadata`**, and `body` is a single JSON string, so the psql way to read it is to
dump it to a file and grep, not to try `jsonb_array_elements` on `metadata->'reviews'`
(that returns zero rows and looks like "no reviews ran"):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
  "SELECT body FROM diagnosis_artifacts
    WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report';" > /tmp/report.txt
grep -o '"reviewer":"[a-z_]*","verdict":"[a-z]*"' /tmp/report.txt
```

`decided_by` names the seat that gated it. Read `unreadable` before trusting
`abstained`: on this 16-seat gate the abstentions are footprint filtering, which is
normal and cheap — `unreadable: 0` is what says no seat failed to run.

Resubmit on the SAME correlation so the trail accumulates:
`RESUBMIT_CORR=<corr> ./…/097_TRIGGER_council_review_v1.sh <submission_r2.json>`

## Measuring what a page actually looks like (added 2026-07-27)

The tool that found 101 unreadable text pairs in two minutes. Run it before
believing a page is fine, and after any palette or component-template change.

```bash
scripts/render_audit.py --sitemap https://<domain>/index.html      # whole site
scripts/render_audit.py --width 1280 https://<domain>/about.html   # desktop
scripts/render_audit.py --json out.json --quiet https://...        # for diffing
```

Exit status is 1 on any failure, so it works as a gate.

**Read the "slow-loading image(s) re-checked OK" note, and never skip the
re-check.** A headless render fires every image request at once and our own
origins throttle the burst, so the browser's "this failed to load" is evidence
of a LOAD failure, not a missing file. Measured here: **41 reported broken, 35
of them served 200 on an unhurried retry**, 6 were real. Acting on the first
number sends someone regenerating assets that are already live. The tool
re-checks each one serially over HTTP before reporting; if you ever reimplement
this, reimplement that too.

**A contrast check cannot see a brand regression.** Making `--color-primary`
light fixed 98 failures and turned three heroes pale blue — dark ink on light
blue, comfortably passing AA, and completely wrong for a "deep navy dominant
field" site. Screenshot after every palette change:

```bash
CH=~/.cache/ms-playwright/chromium-*/chrome-linux64/chrome
$CH --headless=new --no-sandbox --hide-scrollbars --window-size=390,1400 \
    --virtual-time-budget=8000 --screenshot=out.png https://<domain>/index.html
```

## Correcting a palette, and why you cannot just regenerate the stylesheet

`assets/css/styles.css` is written by exactly one path: webdesign-agent's
`deploy_css` step. Reaching it means running `analyze_design` first — an LLM
that emits a fresh `color_scheme` on every run, which **wins over the palette
row** for the eight core slots (`corePaletteKeys`). The pin that is supposed to
hold it steady, `design_intent.palette.reference_values`, is handed to the model
as *"starting points, not exact targets … you may adjust them"*. **It is
advisory by construction, and a memory landmine describes it as a pin — that is
wrong and this is the correction.**

Proof the drift is real, and the method for checking it on any site: render the
layout template locally against the palette row and diff it against the served
file.

```bash
# layout css_template + structure_tokens + typography_set fonts + palette colours
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A <<'SQL'
SELECT jsonb_build_object('layout_template', l.css_template, 'structure_tokens', l.structure_tokens,
                          'typography', ts.fonts, 'palette_new', p.colours)::text
FROM layouts l, typography_sets ts, palettes p
WHERE l.id='<layout_id>' AND ts.id='<typography_set_id>' AND p.id='<palette_id>';
SQL
```

The template uses only `{{palette}}`, `{{typo}}`, `{{token}}` and a single
`{{with palette "heading" ""}}`, so a regex substitution reproduces it exactly —
which is how the drift was proven: every structural rule matched byte-for-byte
while all five core colours differed by a shade.

**Correct the data AND the artefact, or the fix is temporary.** Update
`palettes.colours`, `css_themes.color_palette` and
`design_intent.palette.reference_values` together (they are three copies), then
publish the stylesheet directly:

```bash
docs/.../brochure_component_library/scripts/deploy_stylesheet_direct.sh \
    <domain> <local-styles.css> [assets/css/styles.css]
```

That script publishes one file through the git-adapter
(`system.adapter.git.requests`, `repo_name: sites`) without running a design
pass. **Note for the next session: this needs a Kafka publish, which a
restricted permission mode may refuse — it did on 2026-07-27, and the colour fix
sat ready but unpublished as a result.**

## Where the palette slots actually come from

```sql
-- every palette slot every active layout declares, and how many declare it
SELECT slot, count(DISTINCT l.name)
FROM layouts l, LATERAL regexp_matches(l.css_template, '\{\{\s*palette\s+"([a-z_]+)"', 'g') m(slot)
WHERE l.is_active GROUP BY slot ORDER BY 2 DESC;
```

18 of 18 layouts declare the same 17; then a long tail of 60+ names in 1–3
layouts each. A per-site generated palette supplies **8**. The gap is
`bugs_open/113`.

## Re-rendering a page so a template or content_data change reaches the served HTML

The work-item route, with `reason='section_data_resolved'` — this is the one
that re-renders every section through the CURRENT template with no LLM. Without
that reason you get assemble-only, which restitches the STORED html and your
template change never appears.

```sql
INSERT INTO site_work_items (site_id, item_type, item_key, status, pipeline, summary, spec,
  handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
SELECT p.site_id, 'page_rerender',
       'page_rerender_' || p.name || '_<site>_<what>_<yyyymmdd>',
       'triaged', 'build', 'Republish ' || p.name || ': <why>',
       jsonb_build_object('domain','<domain>','page_id',p.id::text,'page_name',p.name,
                          'filename', p.name || '.html', 'reason','section_data_resolved'),
       'page-rerender','operator:<workstream>','operator:<workstream>',0,3,NOW(),NOW()
  FROM pages p WHERE p.site_id='<site>' AND p.name IN (...);
```

**Check for NULL `content_data` first** — a page carrying one escalates to the
content writer and the copy IS regenerated:

```sql
SELECT p.name, count(*) FILTER (WHERE pc.content_data IS NULL)
FROM pages p LEFT JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id='<site>' GROUP BY p.name;
```

Measured 2026-07-27: 11 such items went from `triaged` to `complete` in well
under the 7–9 minutes this runbook previously recorded — 030 is fixed and the
lane is fast now.

## `kubectl exec` inside a `while read` loop eats the loop's input

```bash
while IFS='|' read -r fn id; do
  kubectl ... psql -c "..." > "$fn.html"     # <-- consumes stdin, loop runs ONCE
done < list.txt
```

Every file after the first is silently missing. Redirect the inner command's
stdin (`< /dev/null`) or iterate over an array instead. Cost 5 minutes and a
confusing "No such file or directory" from a grep over files that should exist.

## CSS backgrounds are invisible to the link/img census (added 2026-07-29)

The census above greps `href="…"` and `<img … src="…"`. A dead
`background-image: … url('/assets/images/hero.jpg')` passes both — it is neither an
href nor an img src, and because the url() sits under a gradient the page degrades to
a flat colour band instead of a visible broken image. Three pages shipped exactly this
(114 family). Sweep backgrounds too:

```bash
grep -hoE "url\('[^']*'\)" page_* | sed "s/url('//;s/')//" | grep '^/' | sort -u \
  | while read -r t; do
      echo "$(curl -s -o /dev/null -w '%{http_code}' "https://fundamentallyai.com$t") $t"; sleep 0.5
    done
```

Gotchas: probe serially with a sleep — a burst of ~20 curls self-throttles the origin
and returns EMPTY 200s (check `wc -c` on every fetched page before extracting from it);
and stored-but-unrendered values don't reach `rendered_html`, so also check
`page_components.content_data->>'hero_url'` — the calculator row held a dead value no
crawl could ever see.

## Placing a section component so it SURVIVES (added 2026-07-29, cost four re-renders)

A page's section list is written down in **three** places, and a re-render honours
all three. Miss any one and the section is dropped at the next rebuild while the
work item reports `complete`:

1. `site_plan_sections` — `(plan_id, page_name, ordering, component_name)`. The
   durable placement. `component_name` holds the component's **function**
   (`info-card-grid`, not its section_type `info-card`).
2. `pages.sections` — the jsonb array of section names on the page row.
3. `page_components.slot_name` — **the one that is easy to miss, because it is
   nullable and nothing complains.** `rerender_page_sections` builds its schema
   lookup from slot names (`loadComponentSchemas(ctx, db, names…)` where `names`
   are `slot_name`s, `rerender_page_sections_action.go:226-232`). A NULL slot
   finds no component, so the action logs *"component not found, carrying stored
   HTML"* and carries the row's stored HTML — which for a freshly inserted row is
   empty — and then `getPageSections` drops the empty section from the assembled
   page (`rerender_single_page_action.go:645`). Two correct behaviours compose
   into a silent disappearance.

```sql
-- all three, in one transaction
UPDATE site_plan_sections SET component_name='<function>'
 WHERE plan_id='<plan>' AND page_name='<page>' AND component_name='<old>';
UPDATE pages SET sections = (SELECT jsonb_agg(CASE WHEN s='<old>' THEN '"<function>"'::jsonb
       ELSE to_jsonb(s) END ORDER BY ord) FROM jsonb_array_elements_text(sections)
       WITH ORDINALITY AS t(s,ord)) WHERE id='<page_id>';
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, build_status)
SELECT '<page_id>', cc.id, <pos>, '<function>', $J${...}$J$::jsonb, 'pending'
  FROM content_components cc WHERE cc.function='<function>' AND cc.is_active;
```

**Never key a placement edit on a `page_components.id` you read earlier** — those
ids are regenerated by every re-render, so the UPDATE/DELETE matches nothing and
reports `UPDATE 0` without failing. Key on `(page_id, function)`.

## Rebundling `/assets/js/snippets.js`

`scripts/rebundle_js_snippets_direct.sh <site_id> <domain>`. There is **no work
item route** to `site-asset-renderer` (no row with that `handler_agent` has ever
existed), so the generic Kafka entry point is the only manual trigger. Do **not**
substitute a `needs_design` item: that routes to `webdesign-agent`, which
regenerates `styles.css` and therefore re-rolls the palette.

Verify by the artefact — the bundle's own header counts its snippets:
```bash
curl -s https://<domain>/assets/js/snippets.js | head -4   # "N active snippet(s)"
```

## Measuring a state the page only reaches on interaction

`scripts/render_audit.py` renders a **local copy** of the page, so a query string
(`?open=<key>`) never reaches `window.location` and an "open state audit" silently
measures the CLOSED page — the same number, a different question. Use
`scripts/probe_reveal_open_state.py <url>`, which forces every `<details>` open in
the DOM and reads computed colours. It prints the count of revealed bodies it
measured, so a run that failed to open anything is distinguishable from a clean one.
Note the site is behind Cloudflare: fetch with a browser `User-Agent` or get a 403.

## The council numbers behind `tool-review-council-simulator` (added 2026-07-30)

The tool's rates are a **dated snapshot baked into the template at build time** (static
page, no API to call at load). Re-measure with these before quoting them anywhere, and
re-render if they have moved. The source is `diagnosis_artifacts` where
`kind='council_report'` — `body` is TEXT holding JSON, so cast it: `body::jsonb`.

**Per-seat objection rate at each severity threshold** — this is the tool's engine:

```sql
WITH r AS (SELECT (body::jsonb) AS j FROM diagnosis_artifacts WHERE kind='council_report'),
rev AS (SELECT rv->>'reviewer' AS seat, COALESCE(rv->'objections','[]'::jsonb) AS objs
        FROM r, jsonb_array_elements(j->'reviews') rv),
sev AS (SELECT seat,
   EXISTS (SELECT 1 FROM jsonb_array_elements(objs) o WHERE o->>'severity'='high') AS hi,
   EXISTS (SELECT 1 FROM jsonb_array_elements(objs) o WHERE o->>'severity' IN ('high','medium')) AS med,
   jsonb_array_length(objs) > 0 AS any FROM rev)
SELECT seat, count(*) AS fired,
       round(100.0*count(*) FILTER (WHERE any)/count(*),1) AS pct_any,
       round(100.0*count(*) FILTER (WHERE med)/count(*),1) AS pct_med_plus,
       round(100.0*count(*) FILTER (WHERE hi)/count(*),1)  AS pct_high
FROM sev GROUP BY seat ORDER BY fired DESC;
```

**Approval rate before/after the 2026-07-22 decision-rule fix** (`bugs_closed/057`):

```sql
WITH r AS (SELECT created_at, (body::jsonb)->>'decision' AS d
             FROM diagnosis_artifacts WHERE kind='council_report')
SELECT CASE WHEN created_at < '2026-07-22' THEN 'pre-fix' ELSE 'post-fix' END AS era,
       count(*) AS runs, count(*) FILTER (WHERE d='approved') AS approved,
       round(100.0*count(*) FILTER (WHERE d='approved')/count(*),1) AS pct
FROM r GROUP BY 1;
```

**Does a medium objection block?** (it does not, and this is what fixed a false label):

```sql
WITH r AS (SELECT (body::jsonb) AS j FROM diagnosis_artifacts WHERE kind='council_report')
SELECT j->>'decision' AS decision, count(*) AS runs,
  count(*) FILTER (WHERE EXISTS (SELECT 1 FROM jsonb_array_elements(j->'reviews') rv,
      jsonb_array_elements(COALESCE(rv->'objections','[]'::jsonb)) o
      WHERE o->>'severity'='high')) AS with_high,
  count(*) FILTER (WHERE EXISTS (SELECT 1 FROM jsonb_array_elements(j->'reviews') rv,
      jsonb_array_elements(COALESCE(rv->'objections','[]'::jsonb)) o
      WHERE o->>'severity'='medium')) AS with_medium
FROM r GROUP BY 1;
```

Measured 2026-07-30: approved 110 runs, **99 with a medium objection, 1 with a high**;
rejected 15, **all 15 with a high**. So high blocks, medium is advisory.

### Gotchas that cost real time here

- **Two different denominators, both real.** `doc_notes WHERE categories ? 'council-gate'`
  returns **284** rows but 18 are threads' own notes rather than verdicts (verdict bodies
  start `COUNCIL GATE — <VERDICT>`), so the verdict denominator is **266**, starting
  07-17. `diagnosis_artifacts` has **362** reports starting 07-10, because they include
  the fix-loop's own council runs. Per-seat data exists ONLY in the 362.
- **`(round N)` in a verdict note is always 1.** All 266 say `(round 1)`. There is no
  rounds-to-approval distribution in that field; do not model one.
- **`pct_any` exceeds the seat's `verdict='object'` rate** (guardian: 89.3% vs 70.2%)
  because a seat can attach advisory objections to an `approve`. Pick the one that
  matches your question and say which.
- **`diagnosis_artifacts` has `body` and `metadata`, NOT a `content` column.** `\d` it
  first; `jsonb_object_keys(content)` fails.

### Re-rendering this page after a template edit

```bash
python3 components/tool-review-council-simulator/install.py --emit   # first install only
# template change thereafter: UPDATE content_components ... then queue a rerender with
# reason='section_data_resolved' (see "Re-rendering a page so a template change reaches
# the served HTML" above). Assemble-only will NOT pick up a template change.
python3 scripts/probe_council_simulator.py --url https://fundamentallyai.com/tools/review-council-simulator.html
```

Measured 2026-07-30: two of three re-renders completed in **under 2 minutes**; the third
sat `triaged` for over 11. Same route, same payload, 5x the latency. Budget the 10
minutes the runbook says and do not re-dispatch on the strength of a slow one.

## Looking at a page, or at one element of it (added 2026-07-31)

```bash
scripts/look.py <url>                          # whole page
scripts/look.py <url> --selector '.trp'        # + a crop of that element
scripts/look.py <url> --selector '.trp' --profile mobile
scripts/look.py <url> --selector '#rcs-results' --pad 40 --out ~/results.png
```

Prints the element's measured box, so a run that found nothing is distinguishable from
one that did, and writes `~/look.png` plus `~/look-<n>-crop.png`.

**Use this instead of hand-rolling `chromium --headless --screenshot`.** Doing it by hand
cost six wrong crops in one session. The four traps it encodes, all measured:

- **snap chromium cannot read or write outside `$HOME`.** A temp file under `/tmp` gives
  `ERR_FILE_NOT_FOUND`; a screenshot written to `/tmp` silently does not appear.
- **Never chase the document height.** Sections sized in `vh` grow with the viewport, so
  "measure the document, render that tall" diverges — measured on `/model-fine-tuning.html`
  as 1000 → 2854 → 4152 → 6141 → 6453, never settling. The script measures and shoots at
  ONE fixed generous height (4000, grown only if the element falls outside), so both
  numbers come from the same geometry.
- **`--screenshot` photographs from the top and IGNORES scroll position.** So
  `scrollIntoView` moves your measurement and not your image. An earlier version of the
  script did exactly that and produced a confidently blank crop.
- **A `file://` copy does not execute the page's cross-origin scripts; an
  `http://127.0.0.1` copy DOES.** To measure an element you must inject a probe, which
  means serving your own copy — over `file://` the site's real JS never runs and the
  layout is wrong. This is also why `probe_council_simulator.py`-style harnesses report
  the arrows and sibling-close of `teaser-reveal-panel` as broken when they are fine: it
  is `file://` that is special-cased, not cross-origin. **Serve over loopback and the
  problem disappears** — the cleanest fix available for any future interaction probe.

Also retries the fetch with backoff: the origin self-throttles under a burst and answers
with a connection reset or an EMPTY 200, and the second is worse because it looks like a
page.

## Proving a deploy on an image that has no `strings` (added 2026-07-31)

CLAUDE.md's recipe is `strings /app/<binary> | grep -c "<symbol>"`. **The
browser-runner-adapter image has no `strings` binary**, so that pipeline feeds grep
nothing and prints a confident `0` — for your symbol *and* for anything else. Use
`grep -ac` on the binary directly, and always in the same exec as a control:

```bash
kubectl exec -n ai-persona-system <pod> -- sh -c '
  grep -c -a "capture_renders" /app/browser-runner-adapter   # the string your change ADDED
  echo "--control--"
  grep -c -a "criteria_json"   /app/browser-runner-adapter'  # a LONG string that predates it
```

Read it this way: **a positive is conclusive; a zero is only meaningful if the
control is positive.** Go compiles short string literals to immediate comparisons
that never reach rodata, so a zero can also mean "too short to be findable"
(`LANDMINES.md`:503 measured this — `selector_count`, 14 bytes, returned 0 on a
binary that fully supported it). Pick a long marker, or a whole sentence.

## Mutation-proving a test on this shared tree (added 2026-07-31)

Never run the loop in the working tree. Another session's half-finished edit to a
*different file in the same package* will break the build mid-run, and **a build
failure and a caught mutant both print `FAIL`** — so a `grep FAIL` summary scores
the invalidated mutants as successes. Measured: four of six, 2026-07-31 evening.

```bash
SCRATCH=<your scratchpad>
rm -rf $SCRATCH/headtree && mkdir -p $SCRATCH/headtree
git archive HEAD | tar -x -C $SCRATCH/headtree          # untracked WIP CANNOT follow you here
cp <your changed .go files> $SCRATCH/headtree/<same paths>
cd $SCRATCH/headtree && go test ./<pkg>/ -count=1        # green baseline, then mutate HERE
```

Two rules inside the loop, both learned the hard way:

- **Build before you test, and say so out loud.** A mutant that does not compile
  proves nothing:
  `go build ./<pkg>/ 2>&1 | head -3 | grep -q . && echo "!! DID NOT COMPILE"`
- **Mutate with Python, not `sed -i`.** Anything containing `|` or `/` makes sed
  parse garbage, print an error that scrolls past, and leave the file untouched —
  which then reads as "the guard is redundant" (`LANDMINES.md`, *"a mutation that
  never happened…"*). The working script is `scratchpad/mutate_renders.sh`.

The export is worth doing anyway: it re-verifies your change against what HEAD will
actually build, which a working-tree `go test` cannot tell you.

## Turning `capture_renders` on, when the chassis carries it (added 2026-07-31)

> **DONE 2026-08-02 ~19:02 — kept here as the worked procedure, not as pending work.**
> Two things changed in the doing and both are corrections to what is written below:
> step 1's control was **weak** (see the negative-control note), and step 2 should
> never have been a bare `UPDATE` (it became seed `292`). Read the corrections.

The order is load-bearing. **Image first, then the key** — DB config is live
immediately and the Go half is not, so writing this before the roll gives you a step
config that reads switched-on while the running binary drops the field.

```bash
# 1. Confirm the chassis carries the caller half (controls in the SAME exec, and
#    on EVERY replica -- `kubectl logs deploy/x` and a single pod both lie by
#    sampling). What was actually run on 2026-08-02, on both:
for p in $(kubectl get pods -n ai-persona-system -l app=agent-chassis \
             -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== $p ==="
  kubectl exec -n ai-persona-system "$p" -- sh -c '
    for s in "capture_renders" "request_browser_run" ".response.data.screenshots"; do
      printf "%-32s %s\n" "$s" "$(grep -acF -- "$s" /app/agent-chassis)"
    done'
done
# expected: capture_renders 1 | request_browser_run 8 | .response.data.screenshots 0
```

**CORRECTION to step 1 as first written: `judge_acceptance_results` is a POSITIVE
control, and a positive control cannot prove your change shipped.** It proves the
grep works and the binary is readable — it reads identically on an image built
*before* the caller half. The control that carries the argument is a **negative**
one: `.response.data.screenshots` was a real concatenated literal in the old binary
and the refactor deleted it (four hand-built path strings became one
`envelopePaths(field, key)`), so **0** there is only possible on a post-change
binary. Use `grep -acF`, not `strings | grep` — `strings` is absent from several
of these images and returns a confident 0 for target *and* control.

```sql
-- 2. Only then, the one key. tool-acceptance-agent is the ONLY live agent
--    referencing request_browser_run; its steps are
--    ensure_site_record -> load_docs -> request_run -> judge.
--    DO NOT run this as a bare UPDATE (see the correction below):
--    it shipped as docs/agent_docs/sql_for_agents/292_acceptance_runs_photograph_a_page_that_passes.sql
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
       '{workflow,steps,request_run,config,capture_renders}', 'true'::jsonb, true)
 WHERE type='tool-acceptance-agent' AND deleted_at IS NULL;

-- 3. Verify at the ARTEFACT, not the config: a passing acceptance run must leave a
--    "Rendered:" line in its own doc_note. A note with no such line means the flag
--    did not reach the adapter, or object storage is unconfigured.
SELECT created_at, left(body, 400) FROM doc_notes
 WHERE subject_type='tool' AND categories ? 'acceptance-run'
 ORDER BY created_at DESC LIMIT 3;
```

**CORRECTION to step 2: a bare `UPDATE` was the wrong instrument, and the reason
generalises.** A DB-only write leaves no record of who set the key, when, against
which binary, or why — and the next person to read `default_config` finds a key with
no provenance at all. It shipped as a numbered delta seed instead, following seed
`147`'s shape (`snapshot_agent` first, `jsonb_set`, a `doc_note`, a `DO`/`RAISE`
verify block). **Two guards, and the second one is the transferable idea: it asserts
a NEIGHBOUR key** (147's `profiles`), because a guard that only checks the key it
just wrote cannot distinguish a surgical write from one that replaced the whole
`config` object. **Both were induced before the real apply** — mutants built with
`sed`, `COMMIT` swapped for `ROLLBACK`, each raising its own message. Note the
mutant must not be self-satisfying: mutating `'true'::jsonb` globally would have
flipped the guard's expectation too and passed.

## Dispatching an acceptance run by hand (added 2026-08-02)

The natural trigger (`tool_acceptance_due`) has a **7-day cooldown** per subject, so
after any recent run there is nothing to wait for — you must queue one yourself.
Insert the work item the discovery check would have emitted; `build-pipeline-trigger`
(every 120s) picks up anything `status='triaged' AND pipeline='build'` on an
**unlocked** site.

```sql
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec,
   page_id, component_id, priority, handler_agent, status, created_by, item_key)
VALUES ('<site_id>', 'operator:brochure_component_library', 'build',
        'acceptance_run', 'low', '<why you are running it>',
        '{"function":"<fn>","component_id":"<cc_id>","page_id":"<page_id>",
          "page_name":"<page>","check":"manual"}'::jsonb,
        '<page_id>', '<cc_id>', 90, 'tool-acceptance-agent', 'triaged',
        'operator:brochure_component_library',
        'acceptance_run:<fn>:<site_id>');
```

Three things that will bite:
- **`idx_swi_dedup` is a UNIQUE index on `(site_id, item_key)`** excluding terminal
  statuses — an existing non-terminal item for the same function rejects the insert.
  Check first; a collision is the queue telling you a run is already coming.
- **Check `sites.locked_at IS NULL`.** The trigger's `pre_query` skips locked sites,
  and a `pre_query` returning no rows **skips silently** — the item just sits.
- **Priority 90 is deliberate** (after builds/rerenders, so acceptance tests the NEW
  page). Expect to queue behind any priority-10 rerender another lane has in flight.

**A render is not a verdict.** `Renders` is empty of `failing_checks` by
construction and must never become signal — if something starts branching on a
render's presence, the two-list design has been undone.

## The weekly contact sheet (owner-approved cadence, added 2026-08-04)

The owner asked for the TL-035 render contact sheet on a cadence. What runs:

```
crontab -l          # 53 8 * * 1 → weekly_contact_sheet_refresh.sh, Mondays 08:53
tail ~/acceptance_renders/refresh.log     # every run logs here — check FIRST
```

`scripts/weekly_contact_sheet_refresh.sh` does three things: an auth pre-check
(the kubeconfig token expires every ~3 days; on a dead token it push-notifies
the owner instead of silently failing), regenerates
`~/acceptance_renders/contact_sheet.html` via `contact_sheet.py --limit 8`, and
push-notifies that the sheet is fresh.

**The claude.ai gallery page does NOT auto-refresh** — measured 2026-08-04:
headless `claude -p` carries no Artifact tool (checked the roster, not
assumed). To refresh it, say in any interactive session: *"republish
~/acceptance_renders/contact_sheet.html as an artifact updating
https://claude.ai/code/artifact/14a45889-e1f0-46e9-969a-08295cc36650"*.
If that URL 404s or errors as deleted, the owner removed it — publish fresh,
then update `ARTIFACT_URL` in the wrapper and this section (it has already
happened once: the 08-03 page 95bd1577… was gone by 08-04).

Gotchas the first runs paid for:
- **`/snap/bin` must be on the cron PATH** — kubectl is a snap; without it the
  auth pre-check reports "token expired" for command-not-found.
- **Push messages truncate ~200 chars on phones** — keep the actionable part
  at the FRONT of the message.
- To stop the cadence: `crontab -e` and delete the line. To change it, edit
  the schedule there — the script itself is cadence-agnostic.

---

## improvement sweep (added 2026-08-05)

Site id for every query below: `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`.

### 0. Pre-flight — MANDATORY, the script's own header says so

Triage promotes EVERY `detected` row and dispatches it to a handler that changes live pages,
so a stale row becomes live page churn. Review them against the live artefact, then cancel
with the evidence recorded:

```sql
SELECT id, item_type, item_key, left(summary,70) FROM site_work_items
WHERE site_id='<site>' AND status='detected' ORDER BY item_type;

UPDATE site_work_items SET status='cancelled', updated_at=now(),
  result = result || '{"cancelled_by":"<session>","reason":"<measured evidence>"}'::jsonb
WHERE id='<row>' AND status='detected';   -- the status guard prevents a double-cancel race
```
**Cancelling loses no signal** — the sweep re-runs the full audit chain and re-files anything
still true. Cancel when the row's CLAIM is false; KEEP it when the claim holds even though its
evidence text has drifted (e.g. a finding naming a component that has since been replaced, on
pages that still share an identical pattern).

### 1. Fire it

```bash
./run_improvement_sweep_once.sh fundamentallyai.com   # save the printed SWEEP_CORR
```
Not within ~300s of a chassis pod restart — the spawn is silently dropped. Check first:
`kubectl -n ai-persona-system get pods -l app=agent-chassis -o custom-columns=NAME:.metadata.name,START:.status.startTime`

### 2. Did it run? — `orchestration_states` has **NO `agent_type` column**

Getting this wrong is how a completed sweep reads as a lost dispatch (WRONG_CALLS 08-05).
Run it by hand before wrapping it in any watch loop, and never send stderr to `/dev/null`:

```sql
SELECT orchestration_id, current_step, status, created_at, updated_at, error
FROM orchestration_states WHERE correlation_id='<SWEEP_CORR>' ORDER BY created_at;
```
A healthy sweep is ~14 rows, all `COMPLETED`, `error` NULL. A row count of 0 shortly after
firing is usually latency, not loss — publish→start has been measured at up to ~29 min.

### 3. The 291 gate — audited, skipped, or not-converging

```sql
SELECT jsonb_pretty(collected_data->'audit_state') FROM orchestration_states
WHERE correlation_id='<SWEEP_CORR>' AND collected_data ? 'audit_state' LIMIT 1;
```
`audit_due=true` + `not_converging=false` ⇒ the FULL audit chain ran (design audit + the
discoveries), which is what makes the run worth its spend. A third run at an unchanged
fingerprint files one `capability_gap` roadmap row instead of reporting clean — correct
behaviour (`bugs_open/171`), not a failure.

### 4. What it left behind

```sql
SELECT status, item_type, item_key, left(summary,60), left(COALESCE(error,''),110)
FROM site_work_items WHERE site_id='<site>'
  AND status IN ('failed','blocked','deferred','needs_human_review') ORDER BY status;
```
`blocked` with "No handler_agent set" = the finding has no worker. Sometimes by design
(flag-only `capability_gap`), sometimes a finding that can never drain — read which.
A FAILED step can show COMPLETED with `error` NULL elsewhere; read `__step_error` too.

### 5. Proving the designer and copy improver are in the chain

Do not infer this from step names — resolve the spawn targets:

```sql
SELECT k, COALESCE(default_config#>>ARRAY['workflow','steps',k,'config','agent_type'],'(none)')
FROM agent_definitions, jsonb_object_keys(default_config->'workflow'->'steps') k
WHERE type='design-audit-agent' AND is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL AND k LIKE 'spawn%';
-- expect: spawn_visual_auditor -> visual-design-auditor
--         spawn_content_auditor -> content-quality-auditor
```

### 6. Checking the served artefact — gate on bytes before reading a grep

An empty body and a page genuinely missing an element both give `grep -c` = 0
(WRONG_CALLS 08-05):

```bash
b=$(curl -s --retry 3 --retry-all-errors "https://<domain><path>")
n=$(printf '%s' "$b" | wc -c)
[ "$n" -lt 2000 ] && echo "SUSPECT FETCH ($n bytes)" || printf '%s' "$b" | grep -c '<header'
```

---

## Linking a NEWLY BUILT page into the site chrome (added 2026-08-08)

The case: `platform-log-index` built and deployed 2026-08-07 11:07, `in_footer=true`,
and **no live page linked to it**. Chrome is a STORED artefact (REB-006 / `bugs_open/117`)
— a page build never refreshes it — so the stored footer, rendered 09:38, predated the page
by 90 minutes and simply had no such link. A new page can be live, navigable in config, and
still orphaned to every visitor.

**Do NOT reach for `nav-updater`.** Its first step `populate_nav_tables` does
`DELETE FROM site_nav_items WHERE site_id=$1` and rebuilds from `pages`, and
`classifyPagesForNav` skips child-path URLs (`/tools/`, `/blog/`, `/guides/`, …) unless the
page_type is a section-index type — `tool` is not one. Read the full LANDMINES entry
including its **2026-07-31 narrowing** before acting on the headline.

### The three steps that work

```bash
# 0. PRE-FLIGHT — every nav target must already serve, or you publish 404s site-wide.
#    The nav table can hold rows for pages whose build failed.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT url, label, status FROM site_nav_items WHERE site_id='<site>' ORDER BY position;"
# then curl every url and require 200 before continuing.

# 1. Refresh the stored chrome FROM the existing nav tables (no populate step, no deletion).
./docs/leopardessconsulting/scripts/orchestrate_safe.sh nav-link-fixer \
  '{"site_id":"<site>","domain":"<domain>"}'
# verify the link actually landed in the stored html BEFORE propagating:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT slot_name, updated_at, rendered_html LIKE '%<url-fragment>%' AS has_link
    FROM site_components WHERE site_id='<site>' ORDER BY slot_name;"
# expect the link in the slot whose nav flag the page declares — a page with in_footer=t
# and in_header=f SHOULD show footer=true, header=false. That asymmetry is correct, not a
# half-failure.

# 2. Propagate to already-deployed pages in ASSEMBLE mode, reconciling until each serves it.
./docs/leopardessconsulting/scripts/reconcile_footer_nav.sh <site> <domain> '<marker>' 3
```

### Why assemble mode, and the trap that makes it load-bearing

`page-rerender` regenerates section HTML from `content_data` **only** when
`spec.reason='section_data_resolved'`. Omitting the reason takes the else branch, which
re-assembles the STORED section HTML with FRESH chrome — all a nav change needs. Sending the
reason instead runs `rerender_page_sections`, whose pre-check escalates the WHOLE page to the
content writer when any section is missing a schema-required `source:"llm"` field. Chrome
propagation must never carry that risk.

Two more, both from the same landmine: the assemble branch needs **`page_id`**, not
`page_name` (`rerender_single_page` errors `page_id not found in input`); and pages with
`rebuild_policy='owned'` are excluded, because `save_sections` refuses them — firing at them
only produces FAILED orchestrations.

### The complementary case: a content_data edit must NOT use assemble mode

Assemble mode re-ships stored section HTML, so a `content_data` edit is invisible to it.
To serve an edited field, use `rerender_page_sections_direct.sh` (reason
`section_data_resolved`) — and check first that **no** section on that page has NULL
`content_data`, or the whole page escalates to the writer:
```sql
SELECT slot_name, content_data IS NULL FROM page_components pc
 JOIN pages p ON p.id=pc.page_id WHERE p.site_id='<site>' AND p.name='<page>';
```

### Authored vs sourced fields — check before editing either

`SELECT k, v->>'source' FROM content_components, jsonb_each(input_schema->'fields') e(k,v)
 WHERE name='<component>' AND is_active;`
A field with a static `source` (e.g. `site_specs.evidence_base.facts`) is **re-resolved on
every render and overwrites whatever you author** — which is exactly why the capabilities
chart self-corrects from the register. A `source:"llm"` field is authored and persists, but a
future content-writer run can regenerate it, so a hand-corrected number there is not durable.
