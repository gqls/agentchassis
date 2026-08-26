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

## Tests

```bash
go test ./platform/orchestration/datahelpers/ ./platform/orchestration/actions/ \
        ./platform/orchestration/actions/discovery_checks/ -count=1
scripts/verify-head-builds.sh --with <changed files> --test    # never hand-roll `git archive`
```
