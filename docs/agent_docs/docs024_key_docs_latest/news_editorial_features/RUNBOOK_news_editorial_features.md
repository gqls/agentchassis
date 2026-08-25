# RUNBOOK — news editorial features

Every query and command this workstream had to get right, with its gotcha
attached. When one changes, change it **here**.

DB access is always:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

---

## 1. The state-of-the-feed query

One query, everything that matters about where the pipeline stops:

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE topics IS NOT NULL AND topics <> '[]'::jsonb) AS with_topics,
       count(*) FILTER (WHERE entity_ids IS NOT NULL AND array_length(entity_ids,1)>0) AS with_entities,
       count(*) FILTER (WHERE duplicate_of IS NOT NULL) AS grouped,
       count(*) FILTER (WHERE published_page_id IS NOT NULL) AS published
  FROM content_feed_items;
```

**Gotcha:** `array_length(x,1)` returns NULL for an empty array, not 0 — so the
`FILTER` must test `> 0` and tolerate NULL, which `array_length(...)>0` does
(NULL is not true). Testing `entity_ids IS NOT NULL` alone would count an empty
`{}` array as populated.

## 2. Proving a column has no writer

```bash
grep -rn "duplicate_of" --include=*.go platform/ internal/ pkg/
grep -rn "entity_ids"   --include=*.go platform/ internal/ pkg/
```

**Gotcha:** an empty grep is only evidence if the same grep shape finds something
when it should. Run a control in the same breath — `grep -rn "relevance_score"`
over the same paths returns hits, which proves the search is looking where the
writes live. Without that control an empty result is indistinguishable from a
mistyped path.

## 3. Proving an agent definition is absent

Always include a positive control **in the same query**, or a zero row count is
just as consistent with a broken filter:

```sql
SELECT type FROM agent_definitions WHERE type IN
 ('article-rewriter','feed-publisher','feed-lifecycle','news-analyst',
  'story-researcher','analysis-writer','visualization-renderer','data-chart-generator');
-- expect 0 rows

SELECT type FROM agent_definitions WHERE type='feed-triage' LIMIT 1;
-- expect 1 row  <- this is the control
```

**Gotcha:** the live roster query needs the snapshot guard, or you read a
historical row. The full-fidelity form is
`WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL`.

## 4. Proving a table is absent

```sql
SELECT to_regclass('public.event_timeline') IS NOT NULL AS event_timeline_exists,
       to_regclass('public.topic_packages') IS NOT NULL AS topic_packages_exists;
```

`to_regclass` returns NULL rather than raising, so this never aborts a
transaction the way `\d event_timeline` would.

## 5. Reading what a component actually is — never a Go grep

The chart renderers are **rows in `content_components`**, not code. A
`--include=*.go` grep will report them absent, which is exactly the mistake
`register/visualisation-and-charts.md` was written to stop.

```sql
SELECT name, component_level, render_mode, is_active
  FROM content_components
 WHERE name IN ('evidence-chart','evidence-timeseries','mechanism-flow')
    OR name ILIKE '%news%';
```

**Gotcha:** the homepage news component is named **`Latest News Feed`** (spaces,
title case), not `latest-news`. A kebab-case `IN (...)` list silently misses it.
Use `ILIKE '%news%'` when enumerating.

## 6. The triage prompt (where concept extraction actually happens)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps')
  FROM agent_definitions
 WHERE type='feed-triage' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

The `topics` array comes back from this prompt's JSON and is written at
`platform/orchestration/actions/feed_triage_actions.go:245` (parsed at :204).

## 7. Firing the diagnosis loop for this workstream

```bash
SLUG=<slug> ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```

**Gotchas, all of which cost time if missed:**
- The script prints **two** correlation ids. The intake one is *not* the key the
  artifacts are written under. Save the one labelled `RUN_CORRELATION_ID` — the
  dispatch loop mints it and stamps it back as `spec.dispatch_correlation_id`.
- Do **not** also publish by hand. `diagnose-pipeline-trigger` is enabled, so the
  loop claims the item within ~60s; a manual publish is a second full run on a
  correlation that cannot be joined to the first (`bugs_open/124`).
- The symptom must state a **mechanism** and **point at** tables/symbols without
  asserting counts — the loop fetches and cites the numbers itself.
- Poll with the run correlation, not the intake one:

```sql
SELECT current_step, status, updated_at FROM orchestration_states
 WHERE correlation_id::text='<RUN_CORRELATION_ID>' ORDER BY created_at;

SELECT status FROM site_work_items WHERE item_key='needs_diagnosis:<slug>';
```

- A code-only symptom (no site, no pages) makes the script warn *"nothing to key
  coverage on — dispatching blind"* and *"Subject: NONE — persist_note will
  SKIP"*. Both are expected for a platform-wide question, not failures. It means
  the verdict lands in `diagnosis_artifacts` rather than in a per-site
  `doc_notes` row.

## 8. Reading the verdict when it lands

```sql
SELECT body FROM doc_notes
 WHERE categories ? 'diagnosis'
 ORDER BY created_at DESC LIMIT 3;

SELECT * FROM diagnosis_artifacts
 WHERE correlation_id::text='<RUN_CORRELATION_ID>';
```

## 9. Checking the pools are still dormant

```sql
SELECT domain, status FROM sites WHERE status='pool' ORDER BY domain;
-- 17 rows, all *.internal, as of 2026-08-19
```

They are invisible to the fleet sweeps because those only walk deployed sites,
and to the news machinery because that only fires on classified sites with live
pages. Pools have neither. Verify rather than assume before arming one.

---

## 10. Retiring a feature (de-listing) — prepared 2026-08-21, NOT yet exercised

The ratified lifecycle says retirement is **deliberate de-listing, never
deletion**: the page keeps serving at its URL, and stops being pointed at. That
makes a retirement a **removal-shaped chrome change**, which needs two mechanisms
that have never run together. Both are prepared below; neither has retired a real
feature yet, so treat the first run as an experiment and record what happens.

### 10.1 What makes this awkward, in one paragraph

The page being retired is **owned** (so the ordinary reconcile refuses it) *and*
it is the **subject** of the change (its own link must disappear from everywhere
else). So you need the owned-page path for the page itself and the removal-aware
reconcile for every other page — and a marker census will lie to you about both,
because the string legitimately appears on the hub (as a listing entry) and on the
retired page itself (as its own canonical, `og:url` and schema `@id`).

### 10.2 The order

1. **Decide it is retired.** The operational test from the design doc: the
   story's cluster has stopped accruing articles in the feed. Not a clock.
2. **Remove it from the hub listing** — edit the `info-card-grid` `content_data`
   on `insights-index`, re-render locally, `UPDATE` the locked row, assemble-only
   rerender. (Worked example: the two-card update of 2026-08-21.)
3. **Set `noindex` if the page is judged stale enough to harm**, leaving
   `status='active'` so the URL keeps serving. Do **not** set
   `status='archived'` unless you intend it to stop serving — check what that
   does on this platform before assuming.
4. **Propagate to every other page** with the peer lane's removal-aware mode:
   ```bash
   docs/leopardessconsulting/scripts/reconcile_footer_nav.sh \
     <site_id> <domain> --absent '<marker>' [rounds]
   ```
   ⚠ **Anchor the marker on the footer-specific form**, e.g.
   `<li><a href="/insights/<slug>.html"`, not the bare slug — otherwise it
   over-reports on the hub and on the retired page itself. Grade on served bytes.
5. **The retired page itself, and any other owned page**, with:
   ```bash
   docs/leopardessconsulting/scripts/refresh_owned_page_chrome.sh \
     <site_id> <domain> '<marker-that-must-APPEAR>' <page_name> [...]
   ```
6. **Verify at the served artefact**, never at a status: the slug gone from other
   pages' footers, the page itself still 200, the hub no longer listing it.

> **UPDATED 2026-08-21 (later) — step 5 is now usually UNNECESSARY, and a
> retirement probably needs ONE mechanism, not two.**
> `reconcile_footer_nav.sh` was excluding every `rebuild_policy='owned'` page on
> the grounds that `save_sections` refuses them. That is the SECTION path, and the
> script runs in ASSEMBLE mode. Verified: `rerender_single_page_action.go` has
> **zero** `save_sections` references, and its only owned-page branch is
> `loadVerbatimPageHTML` — **owned AND exactly one component AND that component
> `deploy_mode='verbatim'`**. Fleet-wide: **174 owned active pages, 3 verbatim,
> 171 not.** The filter is now narrowed to match, and the dry run on robot-hands
> went from 26 pages (6 skipped) to **33 pages (0 skipped)**.
> **So step 4 now covers owned editorial pages too, and step 5 is only for a
> genuinely verbatim page** — of which this lane has none. Found and measured by
> the `dartsonline_traffic` lane; verified here before the change.

### 10.3 What is vouched and what is not — tested 2026-08-21

`refresh_owned_page_chrome.sh` **safety property: VOUCHED.** Run against
`electric-vs-pneumatic-economics` (owned, 5 permanently-locked components). It
flips the page to `generic`, re-renders in assemble mode, and restores `owned`
from an EXIT trap *before* verifying — "protection first, cosmetics second" in its
own output. After the run: `rebuild_policy='owned'`, all **5 locked rows intact**,
served page **byte-identical at 86,602** with hero and all 10 chart rows.

**Fix property: NOT vouched.** The test page's chrome was already current, so the
run was a no-op by construction — it proves the script does no harm, not that it
propagates a stale footer. **No stale owned page existed to test on.** The first
real retirement is therefore still the first genuine exercise of that half; say so
rather than reporting it as proven.

**The window to know about:** the script sets `rebuild_policy='generic'` for the
duration. On this shared tree another session's wide rebuild could land in that
window. Two things bound it: the EXIT trap restores ownership even on failure or
interrupt, and **permanently-locked `page_components` are preserved rather than
regenerated**, so authored copy has a second layer. Both held in the test.

## 11. Dry-running a template change against a locked row BEFORE accepting it — proven 2026-08-25

When another lane changes a shared component template and the change is queued at one of
our locked rows, "it only changes X" is a **hypothesis**. Our editorial rows were rendered
by hand at seed time, so a platform re-render is not guaranteed to reproduce them. This
settles it offline, with no cluster writes.

**The control is the load-bearing step and it is the one that gets skipped.** Rendering the
new template and eyeballing the diff proves nothing on its own: a harness that renders
differently from the platform produces a plausible-looking delta whether or not the change
is safe. Prove the harness reproduces the CURRENT stored bytes first.

```bash
# 0. Harness (replicates executeGoTemplate: text/template, missingkey=zero,
#    six-function funcmap, "<no value>" stripped).
cp <a previous session's>/scratchpad/build/render.go "$S/dryrun/render.go"
printf 'module dryrun\n\ngo 1.21\n' > "$S/dryrun/go.mod"

PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At"

# 1. The PRE-change template is the component_versions snapshot the converting
#    lane took; the live row is the POST-change one. Do not assume v1 is "old" —
#    check created_at against content_components.updated_at (11 ms apart in the
#    2026-08-23 case, snapshot first).
$PSQL -c "SELECT html_template FROM component_versions
           WHERE component_id='<cid>' AND version_number=1;"        > tpl_v1.html
$PSQL -c "SELECT html_template FROM content_components WHERE id='<cid>';" > tpl_live.html

# 2. Live data + the stored artefact for the row.
$PSQL -c "SELECT content_data::text FROM page_components WHERE id='<pcid>';" > cd.json
$PSQL -c "SELECT rendered_html      FROM page_components WHERE id='<pcid>';" > stored.html

# 3. THE CONTROL. v1 + live content_data must reproduce stored.html BYTE-FOR-BYTE.
#    psql -At appends ONE trailing newline the stored value does not have — strip
#    exactly one byte, do not "normalise whitespace", which would hide a real delta.
go run render.go tpl_v1.html cd.json > ctrl.html
python3 -c "
s=open('stored.html','rb').read(); c=open('ctrl.html','rb').read()
s = s[:-1] if s.endswith(b'\n') else s
print('CONTROL', 'PASS' if s==c else 'FAIL', len(s), len(c))"
# FAIL here means STOP: the diff in step 4 is not evidence about anything.

# 4. Only now: the predicted new output. Bind whatever the platform binds —
#    for instance scope that is InstanceToken(function, occurrence), which returns
#    "c-<function>" for occurrence <= 0 (component_instance_scope.go:102-115).
python3 -c "
import json; d=json.load(open('cd.json')); d['InstanceID']='c-<function>'
json.dump(d, open('cdnew.json','w'))"
go run render.go tpl_live.html cdnew.json > new.html
diff <(fold -w100 ctrl.html) <(fold -w100 new.html)
```

**Read the result as a gate, not a report.** The only acceptable delta is the one the
converting lane described. Anything else — a dropped figure, a reordered series, a
`<no value>`, a changed citation — means do not accept, and say so with the diff.

**Gotchas paid for on 2026-08-25:**
- `content_data` may already carry the key the OLD template read. Ours held
  `ComponentID` set to the **slot name** (not the component uuid), which is why the served
  ids looked like slot names. After the change that key is simply unused — harmless, but it
  means you cannot infer the old template's behaviour from the served id alone.
- The `-At` trailing newline shows up as a 1-byte delta and reads like a real difference.
  Check `tail -c 1 | od -c` before believing it.
- Get the binding token from the FUNCTION, not from the pattern of existing ids. Predicting
  `evidence-timeseries-0` from the slot names was wrong; the function returns
  `c-evidence-timeseries`.
