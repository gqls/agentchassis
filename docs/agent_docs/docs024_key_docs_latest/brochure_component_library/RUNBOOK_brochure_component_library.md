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
