# RUNBOOK — CTA label/destination agreement (bugs_open/399)

Every query here was hard to get right once. The gotcha is attached to each.

## THE READING OBLIGATION — the rate, not the row

⚠ **This is the deliverable.** Nobody should read 155 individual `CTA_LABEL_MISMATCH` records;
somebody must notice if 14.6% becomes 30%, or falls to 2% after a prompt change. A record nobody
reads is the class `bugs_open/410` was filed for. Read this monthly, and record each reading dated
in `NOTES_cta_label_agreement.md`.

```sql
SELECT date_trunc('day', occurred_at)::date          AS day,
       count(*)                                       AS pages_recorded,
       sum((context->>'contradicts')::int)            AS contradictions,
       sum((context->>'ambiguous')::int)              AS ambiguous,
       count(DISTINCT agent_type)                     AS producing_agents
FROM agent_error_log
WHERE error_code = 'CTA_LABEL_MISMATCH' AND occurred_at > now() - interval '14 days'
GROUP BY 1 ORDER BY 1;
```

> **STATUS 2026-08-31: `645` IS APPLIED — the rate is readable, but only FORWARD.**
> All six `save_page_sections` steps are armed as of **`2026-08-31 15:09:38Z`** (the ledger's
> `applied_at` for `645_audit_cta_label_agreement_remaining_writers.sql`).
>
> ⚠ **Bound the window on that timestamp, not on "the last 14 days".** The `interval '14 days'` in
> the query above reaches back **before** the second arming and silently mixes two instruments: the
> **145 records banked before it** (`page-build-handler` 61, `page-rerender` 84) came from a
> two-of-six-writer instrument and carry exactly the fleet-wide bias this staging existed to avoid.
> Averaging across the boundary reproduces the bias it was staged to prevent. Use:
>
> ```sql
> AND occurred_at > (SELECT applied_at FROM schema_migrations
>                    WHERE filename = '645_audit_cta_label_agreement_remaining_writers.sql')
> ```
>
> ⚠ **The first forward records are still not a rate.** The four new writers were armed minutes ago
> and the 391 lane's re-resolve burst has not yet landed. Give it a full cycle across all six before
> quoting a percentage — and see the burst warning below.

The arming was staged on purpose (council `guardian` seat, corr `e9bda035`): `643` armed the two
primary writers as a canary, `645` armed the remaining four once that canary fired from **both**
producers (it did: 61 and 83 records, 2026-08-31). Between the two migrations the record was a
**smoke test, not a measurement** — an instrument armed on half its writers reports a rate that reads
fleet-wide and is silently biased, which is the whole reason the census found six steps rather than
the obvious two.

⚠ **`producing_agents` must reach ≥2** once both paths have run (`page-build-handler` and
`page-rerender`). **One producer means the coverage claim is failing silently** — that is the whole
reason migration 643 censuses six steps rather than arming the obvious two.

⚠ **A pre-roll zero is not a clean fleet.** The pass is inert until an image carrying
`cta_label_audit.go` rolls; an older binary ignores an unknown config key. Confirm the binary first:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
```

An empty result means "scrolled out of range", not "unstamped" — fall back to the binary probe with
BOTH controls (a sha that must be present and one that must be absent).

## The baseline census, re-runnable without the cluster's help

This is the token-overlap census that sized the class **before** the mechanism existed. Keep it: it
is the only measure that covers rows written before the audit was armed.

⚠ **It is NOT the same predicate the audit uses** and will not produce the same number. The audit
asks the matcher's ranked question; this asks a token-overlap question. Do not compare them
directly, and do not carry the 14.6% forward as if the audit had produced it.

⚠ The label field is **not derivable from the title key by a string rule** — the five live pairings
are `cta_target_title`↔`cta_text`, `secondary_cta_target_title`↔`secondary_cta` (bare stem!),
`primary_cta_target_title`↔`primary_cta`, `cta_primary_target_title`↔`cta_primary_label`,
`cta_secondary_target_title`↔`cta_secondary_label`. Production reads the component schema instead
(`datahelpers.DeriveCTAURLFields`).

```sql
WITH pairs(tkey,lkey) AS (VALUES
  ('cta_target_title','cta_text'),('secondary_cta_target_title','secondary_cta'),
  ('primary_cta_target_title','primary_cta'),('cta_primary_target_title','cta_primary_label'),
  ('cta_secondary_target_title','cta_secondary_label')),
raw AS (SELECT s.id site_id, s.domain, p.name page, pc.updated_at,
         pc.content_data->>pr.lkey label, pc.content_data->>pr.tkey title,
         pc.content_data->>replace(pr.tkey,'_target_title','_url') url
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  CROSS JOIN pairs pr WHERE pc.content_data ? pr.tkey
    AND COALESCE(pc.content_data->>pr.tkey,'')<>'' AND COALESCE(pc.content_data->>pr.lkey,'')<>''),
j AS (SELECT r.*, d.name dn, d.title dt, d.nav_label dnav
      FROM raw r LEFT JOIN pages d ON d.site_id=r.site_id AND d.url=r.url),
tok AS (SELECT *,
  ARRAY(SELECT left(w,5) FROM unnest(regexp_split_to_array(lower(regexp_replace(label,'[^a-zA-Z0-9]',' ','g')),'\s+')) w
        WHERE length(w)>3 AND w NOT IN ('your','with','this','that','from','more','here','free','best','about','into','over','than','they','have','what','when','will','make','find','view','read','learn','start','click','today','online','page','site','week','take','need','know','ready','first','full','back','only','just','next','using','check','explore')) lt,
  ARRAY(SELECT left(w,5) FROM unnest(regexp_split_to_array(lower(regexp_replace(concat_ws(' ',split_part(title,'|',1),dn,split_part(dt,'|',1),dnav),'[^a-zA-Z0-9]',' ','g')),'\s+')) w
        WHERE length(w)>3 AND w NOT IN ('your','with','this','that','from','more','here','free','best','about','into','over','than','index','page','site','online')) tt
  FROM j)
SELECT count(*) FILTER (WHERE NOT (lt && tt)) AS mismatched, count(*) AS pairs,
       round(100.0*count(*) FILTER (WHERE NOT (lt && tt))/count(*),1) AS pct,
       count(DISTINCT domain) FILTER (WHERE NOT (lt && tt)) AS sites
FROM tok;
```

⚠ **Sample before you quote the count.** My first version compared the label to `_target_title`
alone and over-reported: several "mismatches" were correct destinations whose *marketing* title
shares no word with a perfectly good label. Add `ORDER BY random() LIMIT 20` and read them.
Known residual false positives: ≤3-char tokens are dropped, so `"Read what to do if you can't pay"`
→ `/cant-pay.html` reads as a mismatch.

## Is the writer actually being told?

```sql
SELECT count(*), count(*) FILTER (WHERE position('Destination (fixed)' in prompt_rendered) > 0)
FROM llm_call_log
WHERE created_at > now() - interval '3 days' AND agent_type = 'page-content-writer';
```

⚠ Use `position(... in ...)`, not `LIKE '%...%'` — the parenthesised literal is awkward to escape
through `psql -c`, and the whole query times out over a 7-day window. Keep it to 3 days.
`llm_call_log` is the **training corpus** — read only, never prune.

## Which save steps are armed

```sql
SELECT a.type, x AS armed
FROM agent_definitions a,
     LATERAL jsonb_path_query(a.default_config, 'strict $.**.audit_cta_label_agreement') x
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;
```

⚠ Use `jsonb_path_query` with recursive descent, **not** a top-level `workflow.steps` read: four of
the six live `save_page_sections` steps sit inside a loop's `sub_workflow` and a top-level read
reports them as absent — which reads as "not armed" and is indistinguishable from "does not exist".

**Expected answer since 2026-08-31: all six true** (`page-build-handler`, `page-rerender`,
`page-rebuild`, `pageflow-builder`, `site-work-orchestrator`, `tool-recreation-handler`).

> ### ⚠ THE MIXED-ANSWER CONTROL IS GONE — carry two known-false types
>
> While `645` was held, this census had a **mixed** expected answer (2 armed, 4 not) and that mixture
> was its own control: an all-false result meant you had the wrong spelling, and a matching mix meant
> the predicate discriminated. After `645` the expected answer is **all-true**, and an all-true result
> is indistinguishable from a predicate that matches anything at all. So add types that MUST read
> false:
>
> ```sql
>   AND a.type IN ('page-build-handler','page-rerender','page-rebuild','pageflow-builder',
>                  'site-work-orchestrator','tool-recreation-handler',
>                  'content-writer','council-gate')   -- the last two are the control
> ```
>
> Six true and two false is the answer that means something. Six true alone is not.

> ### ⚠ THE CONFIG KEY AND THE GO FILE ARE THE SAME WORDS IN A DIFFERENT ORDER
>
> Key: **`audit_cta_label_agreement`**. File: **`cta_label_audit.go`**. `LIKE '%cta_label_audit%'`
> returns false on **every** writer, armed ones included — which reads exactly like "the migration
> never applied", and the next move after that reading is to re-apply an applied migration.

> ### ⚠ AND THE WORK-ITEM TYPE IS NOT THE CHECK NAME
>
> The discovery check is **named** `misdirected_cta` (`check_misdirected_cta.go:64`) and **files**
> `item_type='cta_names_unknown_destination'` (`:352`). Querying
> `WHERE item_type='misdirected_cta'` returns **zero rows in all of history**, live table and
> archive both — which reads as "this check has never found anything" rather than "you asked for a
> type that does not exist". Cost a real detour on 2026-08-31 while discharging §6's owed
> comparison. Also: `site_work_items` is a rolling window — closed rows move to
> `site_work_items_archive`, so any historical count must `UNION ALL` both tables or it reads zero
> for anything already dealt with.

## ⚠ What this pass does NOT see

`ApplySectionEditAction` (`section_editor_actions.go`) writes `page_components.content_data`
**directly** and never passes through `SavePageSectionsAction`, so it is outside this pass. It is
live: 144 `section_edit` work items, newest 2026-08-26, of which **3** name a CTA field
`[MEASURED 2026-08-26]`. Re-run before quoting:

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE spec::text ~ 'cta_|_cta') AS mentions_a_cta_field,
       max(created_at)::date AS last
FROM site_work_items WHERE item_type ILIKE '%section_edit%';
```

If that CTA share grows materially, widening the pass to the section-editor path is the follow-up —
it was left out because 3-in-144 did not justify a third seam while the first two were unproven.

## Tests

```bash
go test ./platform/orchestration/datahelpers/ ./platform/orchestration/actions/ \
        ./platform/orchestration/actions/discovery_checks/ -count=1
scripts/verify-head-builds.sh --with <changed files> --test    # never hand-roll `git archive`
```
