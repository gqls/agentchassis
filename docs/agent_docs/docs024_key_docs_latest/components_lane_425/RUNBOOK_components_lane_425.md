# RUNBOOK — components lane (`bugs_open/425`)

Every command here was hard to get right once. The gotcha travels with the command.

## Read the shape of the deck class (the lane's central census)

```sql
SELECT (pc.content_data->'articles'->0 ? 'excerpt') AS new_shape, count(*)
  FROM page_components pc
 WHERE pc.slot_name='content-listing' AND pc.content_data ? 'articles' GROUP BY 1;
```
⚠ **This drains.** 5/12 on 09-02, **9 new / 8 old at 2026-09-03 15:05Z**. Never quote a figure from
a document; re-run it. A count here goes stale by SUBTRACTION as the fleet re-renders, which is the
opposite direction from the estate's usual staleness-by-addition and just as wrong.

## Attribute a write — the join that works, and the two that do not

```sql
SELECT to_char(h.created_at,'MM-DD HH24:MI:SS') AS at, h.source, h.source_item_id,
       (h.content_data->'articles'->0 ? 'excerpt') AS present_BEFORE_this_write,
       wi.item_type, wi.handler_agent, wi.spec->>'reason' AS reason, wi.created_by
  FROM page_component_history h
  LEFT JOIN site_work_items wi ON wi.id = h.source_item_id
 WHERE h.page_id = '<page uuid>' AND h.content_data ? 'articles'
 ORDER BY h.created_at;
```
- ⚠ **Join on `page_id`, never `component_id`.** That column holds a **`page_components`** id, not a
  library id, and is NULL on 44,555 of 45,285 rows. Filtering it by a `content_components` id
  returns zero and reads as "the table is blind here".
- ⚠ **`present_BEFORE_this_write` is the load-bearing column.** The table archives the state each
  write REPLACED. Omit it and you cannot tell a writer that PRODUCED a value from one that
  inherited it — the error that cost this lane a day.
- ⚠ **A `save_page_sections_overwrite` row proves a SAVE RAN, not that content CHANGED** (the INSERT
  has no `IS DISTINCT FROM`). The trigger-written `artefact_archive_trigger` rows *are* change-gated.
- ⚠ **The sections path DELETE/re-INSERTs**, so a new `page_components.id` on every re-render is
  expected and is not evidence of new content.

## Dispatch a re-render, and choose a target that can fail

```sql
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec, priority,
                             handler_agent, status, created_by, batch_id, pipeline, approval_mode)
SELECT s.id, 'side_effect', 'page_rerender', 'low',
  '<what this tests, and what each outcome would mean>',
  '{"domain":"<d>","reason":"template_changed","page_id":"<uuid>","page_name":"<name>"}'::jsonb,
  80, 'page-rerender', 'triaged', 'bugs_open/425', '<batch uuid>', 'build', 'auto'
  FROM sites s WHERE s.domain='<d>' RETURNING id, created_at;
```
⚠ **Before filing, prove the baseline LACKS what you are testing for** — on a path that persists
`stored ⊕ fresh`, a preserved value and a re-resolved one are the same bytes, so a baseline that
already carries it makes the result predetermined. This lane spent two dispatches learning that:

```sql
SELECT (content_data->'articles'->0 ? 'excerpt') AS baseline_already_has_it,
       to_char(updated_at,'MM-DD HH24:MI') AS untouched_since
  FROM page_components WHERE page_id='<uuid>' AND slot_name='<slot>';
```
⚠ **`reason` must be one of the five recognised literals** (`section_data_resolved`,
`template_changed`, `cta_links_stale`, `image_landed`, `literal_markdown`). A prose sentence takes
`else_step` exactly as NULL does — assemble mode, which re-ships stored bytes and reports success.
⚠ **Coverage first** (CLAUDE.md): check for open items on the target, and check `sites.locked_at`.

## Read a dispatch without fooling yourself

```sql
SELECT wi.status, to_char(wi.claimed_at,'HH24:MI:SS') AS claimed,
       to_char(wi.completed_at,'HH24:MI:SS') AS completed, left(wi.error,200) AS err
  FROM site_work_items wi WHERE wi.batch_id='<batch uuid>';
```
⚠ **Print the item status in the same breath as the artefact reading.** A timed-out watcher prints
the unchanged baseline, which is indistinguishable from a result. Treat an unmoved `updated_at` as
"did not run", not as "no change".
⚠ **The run's OWN counts cannot grade the run.** `section_count`, `rerendered`, `carried` and
`escalated` are byte-identical between a working re-render and one that resolved nothing — measured
by the `bugs_open/384` lane on one page, where the broken 05:06/05:08 runs and the working 12:54 run
all reported `4 / 4 / 0 / false`. **Only the artefact discriminates.** Read `content_data` and the
rendered bytes; never accept a count as evidence the repair happened.
⚠ **Expect `rendered_html` to SHRINK on a successful repair.** Stripped site-name suffixes plus
collapsed empty elements can outweigh the added excerpt text — robot-hands.com
`/learning-center-hub` went 7,681 → 6,753 B on a clean repair and did not trip the shrink floor. An
acceptance check asserting "html grew" files a success as a regression.
⚠ **Find items by `batch_id`, not `id`** — a batch uuid is what the summary and handoffs quote, and
`WHERE id IN (<batch uuid>)` silently returns nothing.

## Measure the rendered slots — the predicate that discriminates

```sql
SELECT (SELECT count(*) FROM regexp_matches(pc.rendered_html,'article-card__excerpt','g'))       AS total,
       (SELECT count(*) FROM regexp_matches(pc.rendered_html,'article-card__excerpt">\s*</p>','g')) AS empty
  FROM page_components pc WHERE pc.page_id='<uuid>' AND pc.slot_name='content-listing';
```
⚠ **Count the element and its emptiness SEPARATELY.** `0 total` = the slot collapsed (682 working);
`n total, 0 empty` = filled; `n total, n empty` = the pre-682 defect. Counting only the empty form
returns 0 for the first and second cases alike.
⚠ **Never take a class count from a page carrying an orphan row** — its bytes are a mixture of
template eras no re-render can normalise:
`SELECT count(*) FROM page_components WHERE page_id=$1 AND component_id IS NULL;`

## Confirm at the served bytes

```bash
scripts/probe-page-url.sh <domain> <page-name>        # reads pages.url, runs both controls
```
⚠ Never compose a URL from `pages.name`. ⚠ A domain that does not resolve from this machine makes
the script's sibling control FAIL, and it then refuses to answer rather than reporting damage —
`boxingonline.com` has no A record from here, so its served readings came from a peer lane.
⚠ Pair any "the suffix is gone" claim with a control: count the suffix string page-wide and confirm
the remaining occurrences are NOT deck titles (on garden-tools `/care` the single hit is an
unrelated `<li>` label).

## Ask the running service what it is

```bash
kubectl -n ai-persona-system get deploy agent-chassis \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
for pod in $(kubectl -n ai-persona-system get pods -l app=agent-chassis \
             -o jsonpath='{.items[*].metadata.name}'); do
  for sha in <expected> <must-be-absent> 0000000000000000000000000000000000000000; do
    kubectl -n ai-persona-system exec "$pod" -- grep -aqs -- "$sha" /proc/1/exe \
      && echo "$pod ${sha:0:9} PRESENT" || echo "$pod ${sha:0:9} absent"
  done
done
git merge-base --is-ancestor <your-commit> <the stamp>
```
⚠ **The `build provenance` log line scrolls out of reach within minutes** on this service —
`--tail=5000` on a busy chassis pod returns none of it. An empty result there means "not in range",
not "unstamped"; fall back to the binary probe, which has no shelf life.
⚠ **Run all three shas.** The all-zero control returns PRESENT on every pod (it matches Go's
internal digit table), which is exactly why a bare discovery grep for "some 40-hex string" gives
the same wrong answer everywhere.
