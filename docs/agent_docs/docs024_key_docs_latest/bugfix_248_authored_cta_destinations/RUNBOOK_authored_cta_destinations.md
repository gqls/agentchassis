# RUNBOOK — `cta_recompute_clobbers_authored_contact_links`

Commands that were hard to get right, with the gotcha attached.

## The at-risk census (from the bug file, still correct)

```sql
SELECT count(*) FROM page_components pc
WHERE COALESCE(pc.content_data->>'cta_url', pc.content_data->>'primary_cta_url',
               pc.content_data->>'secondary_cta_url')
      ~ '^/(contact|about|privacy|terms|legal)(\.html|/|$)';
```
24 (08-10) → 20 (08-17) → 18 (08-18). **A falling count is not repair** — two of those
components lost their contact link to this defect.

**⚠ The count overstates the reachable population by ~a third.** Only components whose
`function` is in `ctaFieldNames` can be touched by either writer. Narrow it or you will size
the fix against `system-stats` and `tool-*` slots no writer has ever written:

```sql
SELECT cc.function, count(*)
FROM page_components pc LEFT JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id
WHERE COALESCE(pc.content_data->>'cta_url', pc.content_data->>'primary_cta_url',
               pc.content_data->>'secondary_cta_url')
      ~ '^/(contact|about|privacy|terms|legal)(\.html|/|$)'
GROUP BY 1 ORDER BY 2 DESC;
```
`content_components` has **no `deleted_at`** column and its function names are NOT unique —
`hero` and `call-to-action` each have two rows, one `is_active=false`.

## Finding this defect's work items — the item_type is NOT what the bug file says

The bug file reports "the `misdirected_cta` queue holds 192 detected / 95 unresolved". You
cannot re-run that: `item_type='misdirected_cta'` returns **zero rows**. The check files
`item_type='page_rerender'` with the check name in the KEY:

```sql
SELECT status, count(*), max(updated_at) FROM site_work_items
WHERE item_key LIKE 'misdirected_cta:%' GROUP BY 1 ORDER BY 2 DESC;
```

## Attributing a lost CTA to THIS mechanism rather than to a bad regeneration

The 08-17 CONTRIB to the bug file says the two are indistinguishable, because both land as a
`save_page_sections_overwrite` history row. They are distinguishable when a work item is
involved — join on the SECOND, not the day:

```sql
-- 1. when did the component change, and to what
SELECT pc.updated_at, pc.content_data->>'primary_cta_url'
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='<domain>' AND p.name='<page>' AND pc.slot_name='<slot>';

-- 2. what completed in the same second
SELECT swi.item_key, swi.status, swi.updated_at, swi.spec->>'reason'
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id JOIN pages p ON p.id=swi.page_id
WHERE s.domain='<domain>' AND p.name='<page>' AND swi.updated_at > '<just before>';
```
`reason='cta_links_stale'` is the tell: `applyCTARecompute` is the only CTA behaviour that
reason triggers.

**⚠ `page_component_history` must be joined on `page_id` ALONE.** Every history row holding a
contact url has `component_id IS NULL`, so a `(page_id, component_id)` join returns ~nothing
and reads as "no damage found" (this cost the 08-17 CONTRIB a false negative before it caught
it). Retention is not the constraint — the table goes back to 2026-03-16.

**⚠ The archived row holds the value being REPLACED, not the new one.** So the newest history
row for a clobbered component still shows `/contact.html` while the live row shows the tool.

## Reading a component's live schema — NEVER from the seed

An adversarial review nearly killed this fix by citing `docs/agent_docs/sql_for_tables/
005_content_components.sql`, which still shows `"fallback": "/contact.html"` on
`hero.cta_url`. The seed records what the component WAS. Migrations 091/098 changed it.

```sql
SELECT function, key, val->>'source', val->>'on_missing', val->>'fallback'
FROM content_components cc, LATERAL jsonb_each(cc.input_schema->'fields') AS f(key,val)
WHERE cc.function IN ('hero','call-to-action','archetype-grid','archetype-combinations',
                      'gauntlet-cta','content-block-about')
  AND key LIKE '%cta%url%';
```
All ten carry `source=renderer`, `fallback=NULL` (2026-08-18). The utility-area fallbacks that
DO exist live on `site-header`/`site-head` — chrome, a different path (LNK-030).

## Building and testing while the shared tree does not compile

Another session's uncommitted WIP had `tool_acceptance_actions.go` calling a function it had
not written yet, so `go build ./platform/...` failed on code that was nothing to do with this
lane. Build against committed HEAD with only your own files overlaid:

```bash
SB=<scratchpad>/build248
rm -rf "$SB" && mkdir -p "$SB" && git archive HEAD | tar -x -C "$SB"
for f in $(git status --short | grep -E "^ M platform|^\?\? platform" | sed 's/^...//'); do
  cp "$f" "$SB/$f"; done
cd "$SB" && go build ./platform/... && go test ./platform/orchestration/actions/...
```
This is also the honest test: it is what `make build-*` will actually compile.

## The mutation check (do not skip — a green suite proves nothing here)

Every one of these must go RED, and name the right test:

| revert | expect red |
|---|---|
| keep #1 in `applyCTARecompute` | `TestApplyCTARecomputeKeepsAuthoredContactLink` |
| the authored branch in `setCTAField` | `TestSetCTAFieldKeepsAuthoredContactLink` |
| the `candidatesFromHubs` filter | `TestFreshPickRefusesUtilityWhileStoredUtilityIsKept` |

## Council submission — four schema rejections worth knowing before you write one

`097_TRIGGER_council_review_v1.sh` validates client-side and each failure costs a round trip:

1. `operation` must be one of `modify|add|remove|config_change`. Not `add_function`,
   not `create`.
2. `plan.risks` must be a **string**, not an array.
3. `plan.grounded_in` must be an **array of strings** (so risks and grounded_in disagree
   about shape — do not copy one from the other).
4. **A sketch whose every non-blank line starts `#`, `//` or `--` is refused as
   "comment-only".** A markdown doc edit sketched as `### LNK-033 — …` trips this, because
   `#` is a comment prefix. Describe the change in prose instead.
