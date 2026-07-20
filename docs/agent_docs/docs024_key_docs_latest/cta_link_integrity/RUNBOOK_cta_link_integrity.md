# RUNBOOK — CTA / link integrity

Every query below was run on 2026-07-19 and returned the results quoted in
`NOTES_cta_link_integrity.md`. Gotchas are attached to the command that needs them.
When one changes, change it **here**.

DB prefix used throughout:
```bash
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c "<SQL>"
```

---

## R1 — Fleet-wide dead-control census

The headline number. Classifies every anchor on every active page.

```sql
WITH anchors AS (
  SELECT s.domain,
         (regexp_matches(pc.rendered_html, '<a[^>]*href="([^"]*)"', 'g'))[1] AS href
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
  WHERE p.status='active' AND pc.rendered_html IS NOT NULL
)
SELECT CASE
   WHEN href = ''                 THEN '1. empty href=""'
   WHEN href IN ('#','#!')        THEN '2. bare #'
   WHEN href LIKE '#%'            THEN '3. fragment #anchor'
   WHEN href LIKE 'javascript:%'  THEN '4. javascript:'
   WHEN href LIKE 'http%'         THEN '5. external'
   WHEN href LIKE 'mailto:%'
     OR href LIKE 'tel:%'         THEN '6. mailto/tel'
   ELSE '7. internal path' END AS kind,
  count(*) AS anchors, count(DISTINCT domain) AS sites
FROM anchors GROUP BY 1 ORDER BY 1;
```

> **Gotcha.** `page_components` only. It does **not** cover `site_components`
> (header/footer), so nav dead links are invisible to this census — run it again against
> `site_components` for the full picture.

## R2 — Resolve a slot_name to the component actually serving it

**The gotcha that cost the most time.** `page_components.slot_name` resolves by
`content_components.function`, **not** `.name`, and `component_id` is unpopulated for most
rows fleet-wide. Querying by `name` returns zero rows and looks like an orphan.

```sql
SELECT name, function, is_active, length(html_template) AS tpl,
       (SELECT count(*) FROM jsonb_each(input_schema->'fields')) AS n_fields
FROM content_components
WHERE function = '<slot_name>' OR name = '<slot_name>';
```

This is how `bayesian-ranking-hero-tool` was found to be served by
`bayesian-ranking-hero-tool_pre_037` — an `is_active` backup snapshot.

## R3 — Dump a component's CTA schema

```sql
SELECT k AS field, v->>'source' AS source,
       COALESCE(v->>'required','') AS req,
       left(COALESCE(v->>'fallback',''),40) AS fallback
FROM content_components c, jsonb_each(c.input_schema->'fields') AS e(k,v)
WHERE c.name = '<component name>' AND k ~ 'cta|url|label|btn'
ORDER BY k;
```

> **Gotcha.** `input_schema->'fields'` is a JSON **object**, not an array.
> `jsonb_array_elements` fails with `cannot extract elements from an object` — use
> `jsonb_each`.

Read it for the two red flags: a `*_label` with `source:static` whose `*_url` sibling is
absent, and any `*_url` with `source:llm` + `required:true`.

## R4 — Show the template's CTA anchors (is it gated?)

```sql
SELECT (regexp_matches(html_template, '<a[^>]*class="[^"]*btn[^"]*"[^>]*>[^<]*', 'g'))[1]
FROM content_components WHERE name = '<component name>';
```

An anchor with no enclosing `{{if .x_url}}` is class **C** — an empty value renders
`href=""`.

## R5 — Which sites/pages use a component

```sql
SELECT s.domain, p.name AS page, pc.slot_name
FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.slot_name = '<slot_name>' AND p.status='active'
ORDER BY s.domain, p.name;
```

Run before any component fix — these are shared, and the blast radius is the point.

## R6 — Slot names with no library component at all

```sql
SELECT pc.slot_name, count(*) AS instances, count(DISTINCT s.domain) AS sites,
       string_agg(DISTINCT s.domain, ', ') AS domains
FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE p.status='active'
  AND NOT EXISTS (SELECT 1 FROM content_components cc WHERE cc.name     = pc.slot_name)
  AND NOT EXISTS (SELECT 1 FROM content_components cc WHERE cc.function = pc.slot_name)
GROUP BY pc.slot_name ORDER BY instances DESC;
```

> **Gotcha.** Do **not** use `LEFT JOIN content_components ON cc.id = pc.component_id` and
> read NULLs as orphans — `component_id` is unpopulated fleet-wide and you will "find"
> ~100 false orphans. I did; it is corrected in NOTES.

## R7 — The CTA work-item queue (and why it never drains)

```sql
SELECT item_type, status, count(*)
FROM site_work_items
WHERE site_id='<site_id>' AND item_type ~ 'cta|link|control|nav'
GROUP BY 1,2 ORDER BY 3 DESC;
```

Anything at `needs_human_review` is inert: `TriageDetectedItemsAction` never promotes it,
no `handler_agent` claims it, and `load_work_item_actions.go:804` excludes it from re-open
queries. Read the summaries — they are usually correct and unread:

```sql
SELECT item_type, left(summary,110), created_at::date
FROM site_work_items
WHERE site_id='<site_id>'
  AND item_type IN ('unresolved_cta','cta_names_unknown_destination')
ORDER BY created_at DESC;
```

## R8 — Verify a button against the live artefact

Never trust `page_components.rendered_html` alone — verify what actually ships.

```bash
curl -s -o /tmp/pg.html -w "http=%{http_code} bytes=%{size_download}\n" \
  "https://<domain>/<path>.html"

# the CTA anchors as rendered
grep -oE '<a[^>]{0,300}class="[^"]*btn[^"]*"[^>]*>[^<]{0,60}' /tmp/pg.html

# does a fragment target actually exist?
grep -c 'id="guide-start"' /tmp/pg.html     # 0 = dead fragment

# does an external host resolve?
getent hosts <host> || echo NXDOMAIN
```

> **Gotcha.** Tool and guide pages are **not** at the site root. Check `pages.url` first —
> leopardess tools live at `/tools/<name>.html` and guides at `/guides/<name>.html`;
> curling `/<name>.html` returns 404 and looks like a missing page.

## R9 — Ungated CTA anchors fleet-wide (the Phase 2 worklist)

```sql
WITH a AS (
  SELECT name,
         (regexp_matches(html_template,'(.{0,60})<a[^>]*href="\{\{\.([a-z_]*url)\}\}"','g')) AS m
  FROM content_components WHERE is_active
)
SELECT CASE WHEN m[1] ~ '\{\{ *if' THEN 'gated' ELSE 'UNGATED' END AS gating,
       count(*) AS anchors, count(DISTINCT name) AS components
FROM a GROUP BY 1;
```

**Result 2026-07-19: 75 UNGATED anchors across 38 components; 14 gated across 12.**
So roughly **84% of URL-bound CTA anchors in the library violate LNK-005** — an empty or
unresolved value renders `href=""` rather than rendering nothing. Drop the aggregation to
get the per-component list.

> **Gotcha.** This is a 60-character-lookback heuristic, not a parse. A `{{if}}` opened
> further back than 60 chars reads as UNGATED (false positive), and a nested/unrelated
> `{{if}}` inside the window reads as gated (false negative). Good enough to size the
> problem and to prove the direction; **re-derive the exact list with a real template
> parse before mass-editing.**

## R10 — Backup rows still serving live traffic

```sql
SELECT name, function, is_active
FROM content_components
WHERE is_active AND (name ~ '_pre_[0-9]+$' OR name ~ '_bak' OR name ILIKE '%backup%')
ORDER BY name;
```

**Result 2026-07-19: 16 active `_pre_037` rows**, including `bayesian-ranking-hero-tool`,
`blog-listing`, four `header-*` variants, `footer-with-disclaimer`, and five `tool-*`
components.

Then check whether any of them is ambiguous:

```sql
SELECT function, count(*) AS active_rows, string_agg(name,' | ' ORDER BY name) AS names
FROM content_components WHERE is_active
GROUP BY function HAVING count(*) > 1 ORDER BY count(*) DESC;
```

**Result: none of the `_pre_037` rows collide.** They are the *sole* active row for their
function — i.e. they are not stale duplicates shadowing a canonical row, they *are* the
canonical row, wearing a name that says otherwise. Whatever migration 037 intended to
replace them with never landed. Resolution is therefore unambiguous; the defect is
misleading naming plus un-reviewed pre-migration content (which is how the Bayesian
labels survived). **Do not "clean up" these rows by deleting them — that would delete the
live component.** Rename or supersede deliberately.

The five genuine function collisions are per-site forks (`site-header`,
`tool-llm-cost-calculator`, `site-footer`, `tool-bayesian-ranking`, `tool-meme-generator`)
and are expected.

## R11 — Retrieve a diagnosis-loop verdict

The terminal verdict does **not** live in `diagnosis_artifacts` (that holds only the evidence
bundles) and lands in `doc_notes` **only if** `SUBJECT_TYPE` *and* `SUBJECT_KEY` were both set
at trigger time — they are a subject gate on `persist_note`. Without them the run still
completes and still diagnoses; the verdict is in the orchestration's collected data:

```bash
CORR=<correlation-id>
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
 "SELECT collected_data::jsonb -> 'diagnosis' FROM orchestration_states
  WHERE correlation_id='$CORR'::uuid AND collected_data::jsonb ? 'diagnosis' LIMIT 1;" > diag.json
```

Keys: `status` (CONFIRMED/…), `is_fix`, `stopped_by`, `summary`, `conclusion`,
`evidence_trail`. **Read `conclusion`, not `summary`** — `summary` is a stock line
("Diagnosis CONFIRMED — see conclusion"), the substance and the citations are in `conclusion`.

Progress while it runs (poll by correlation id, never by `created_at`):

```sql
SELECT status, current_step, EXTRACT(EPOCH FROM (NOW()-last_activity))::int AS since_s
FROM orchestration_states WHERE correlation_id='<corr>'::uuid ORDER BY created_at DESC LIMIT 1;
```

> **Gotcha — the trigger's `REF` default.** It is `main`. On this repo `main` can be hundreds
> of commits behind the working branch: on 2026-07-19 `origin/main` carried a 2-entry
> `ctaFieldNames` map while the branch carried 6. **Diff the symbol you are diagnosing across
> refs before firing**, and pin `REF` to a branch you have verified carries the current code:
> `git show origin/<branch>:<path> | sed -n '/^var ctaFieldNames/,/^}/p'`

> **Gotcha — closing the intake.** The item parks at `status='awaiting_diagnosis'` and no
> sweep selects that status, so it stays open (and blocks the 090 coverage check for the same
> target) until closed by hand:
> `UPDATE site_work_items SET status='complete', completed_at=now() WHERE item_key='needs_diagnosis:<slug>' AND status NOT IN ('complete','verified');`

## R12 — Section-authority check, per site, before touching page composition

Run BOTH before any page_components/pages.sections edit — the answer differed across all
three sites fixed so far (leopardess: aspect current but page absent → sections governs;
finetuning: no current plan → sections governs; robot-hands: plan current AND lists the
page, with a THIRD composition → only page_components edited):

```sql
SELECT sp.id, sp.is_current FROM site_plans sp
JOIN sites s ON s.id=sp.site_id WHERE s.domain='<domain>' AND sp.is_current;

SELECT page_name, string_agg(component_name,' · ' ORDER BY ordering)
FROM site_plan_sections WHERE plan_id='<plan id>' AND page_name='<page>' GROUP BY 1;
```

> **Gotcha.** `site_plans` has no `status` column (`is_current` boolean) and
> `site_plan_sections` keys on `plan_id`, not `site_plan_id`.

## R13 — Verifying a redeploy: poll for the GOOD content, not the bad content's absence

A mid-deploy fetch of a B2-backed page can return a **310-byte JSON `NoSuchKey` error
body** (HTTP 404) — which contains no markup at all and therefore PASSES any
"broken-marker absent" check. This false-green bit this workstream on 2026-07-20.

```bash
until curl -s -o page.html -w '%{http_code}' "$URL" | grep -q 200 \
  && grep -q '<distinctive-good-marker>' page.html \
  && ! grep -qE '<bad-marker>' page.html; do sleep 10; done
```

Status + good marker + bad-marker absence, all three. Dispatch-to-deploy latency under a
busy chassis was ~35 min (queue, not a drop — pod uptime 5h, nothing restarted).
