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

> **CORRECTED 2026-07-20 — the warning above was right, and the number was wrong by 2.4×.**
> The real parse (`scripts/parse_gates.py`, below) gives **171 UNGATED anchors across 41
> components**, not 70/37. The undercount is not the lookback window: `regexp_matches(…,'g')`
> returns **non-overlapping** matches, and the greedy `.{0,60}` prefix of each match eats the
> *previous* anchor — so in runs of adjacent anchors (nav lists, footer link columns, i.e.
> exactly where CTA anchors cluster) roughly every other anchor is consumed and never counted.
> **Use R9 only for a direction-of-travel reading. For any worklist, run the parse.**

## R9b — the real template parse (use this for worklists)

```bash
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT jsonb_agg(jsonb_build_object('name',name,'function',function,'tpl',html_template))
FROM content_components WHERE is_active AND html_template IS NOT NULL;" > components.json
python3 scripts/parse_gates.py components.json     # writes parsed_anchors.json + a summary
```

Tokenises each template, maintains an `{{if}}/{{range}}/{{with}}` … `{{end}}` block stack, and
marks an anchor **gated only when an enclosing block's condition references the same field**.
Result 2026-07-20: 189 `href="{{.X}}"` anchors — 18 gated / 14 components, **171 ungated / 41**.
After migration 181: 22 gated, **169 ungated**, of which **152 / 29 components** are the CTA
worklist.

> **Read the range/CTA split the script prints — do not use the raw ungated total.**
> An anchor inside a `{{range}}` is an **item link**, not a CTA: the field belongs to the ranged
> item (`{{range .items}}<a href="{{.url}}">`), fed by a query-provided list. Different class,
> different fix. 17 of the 169 are these, across 13 components (`url`, `affiliate_url`).
> **This bites twice:** it inflates the P2.1 worklist by ~10%, and a migration post-condition
> written as *"no ungated `{{.x_url}}` anchor remains in this component"* will trip on a
> range-scoped `.url` and roll back an otherwise correct change — which is exactly what 181's
> first draft did before it was caught. Gate by **exact needle**, per component, per anchor.

> **Then resolve placements before editing anything** — 20 of the 41 are dormant library stock
> holding ~80 of the anchors. `page_components` joins by `slot_name = function` (or `name`);
> `site_components` joins by **`component_id`, which IS populated there** — the R6 gotcha about
> `component_id` being unpopulated applies to `page_components` only. Live-placed: 21 components.

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

## R14 — Watching the observe stage (LIVE in v1.0.1140, 2026-07-20)

> **Renumbered R12→R14 by the bugfix-023 session, 19:10.** Two threads appended to
> this file within hours and both claimed R12 (mine at 14:00: section-authority; yours
> at 19:06: this one). Nothing referenced either number, so no pointer broke. Sequence
> is by commit order, so the later entry moved. Content untouched.

Deploy verified in-pod (never the tag): all four markers present in
`/app/agent-chassis` — `cta derivation delta`, `cta ownership conflict`,
`uncovered cta url field`, `DeriveCTAURLFields`.

Collect the evidence the flip round is gated on:

```bash
# all chassis-family deployments; widen --since as needed
for dep in agent-chassis business-intel vet-intel; do
  kubectl logs -n ai-persona-system deploy/$dep --since=24h 2>/dev/null \
    | grep -E "cta derivation delta|uncovered cta url field|cta ownership conflict"
done
```

**How to read each stream — and what silence means (this matters):**

| line | fires when | silence means |
|---|---|---|
| `cta derivation delta` (Info) | `internal-link-resolver` runs on a section whose schema-derived CTA set differs from `ctaFieldNames` | no CTA-bearing page build has run — NOT "no gap"; the gap is structural (map covers 5 of 33 functions) and will log on the first build touching an unmapped component |
| `uncovered cta url field` (Warn) | build touches a component with a query-resolved `*_url` and no sibling in any form | same as above |
| `cta ownership conflict` (Info) | a rerender's fresh `ResolvedData` would replace a **differing** stored value on a derived CTA field | genuinely no conflict on rendered traffic — equal values are correct silence. Unmapped fossil fields re-resolve to the same value every time (the resolver never wrote them), so they are *expected* to be silent here while loud in the delta stream |

> **Gotcha — spawned pods.** Builds run in short-lived spawned chassis pods;
> `deploy/agent-chassis` logs miss them once reaped. For a thorough sweep use
> the label selector while pods are alive, or rely on the delta stream
> accumulating across days before the flip review — one build's logs are a
> sample, not the census.

> **Gotcha — do not judge the flip on conflict-log volume alone.** The delta
> stream is the coverage evidence; the conflict stream is the damage evidence.
> A quiet conflict stream with a loud delta stream still justifies the flip
> (unmapped fields are unrepairable today even when not actively clobbered).

## R15 — Live link audit: what actually 404s on the deployed sites (`bugs_open/049`)

**The only trustworthy census.** Everything else in this runbook reads
`page_components.rendered_html`, which is **demonstrably not what ships** — leopardess
`/tools/ai-agent-roi-estimator.html` stores two `<a href="">` anchors and the live page has zero.
Stored HTML over-reports. This fetches the artefacts.

```bash
# 1. the page list (any set of sites)
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -tAF'|' -c "
SELECT s.domain, p.url FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.status='active' AND s.domain IN ('finetuning.uk', ...) ORDER BY 1,2;" > pages.txt

# 2. fetch every page, extract internal hrefs from SHIPPED html, test every distinct target
./scripts/live_link_audit.sh          # writes live_anchors.txt + target_status.txt
```

Result 2026-07-20 across 7 sites: 180 pages, 3,386 anchor instances, **68 unique targets 404,
312 anchor instances broken, on 117 of 180 pages**. Runtime ~4 minutes.

Breakdowns worth having (from `target_status.txt` + `live_anchors.txt`):

```bash
grep '^404' target_status.txt | cut -d'|' -f2 | sort | uniq -c | sort -rn      # unique targets/site
grep '^404' target_status.txt | cut -d'|' -f2,3 | sort -u > bad_targets.txt
awk -F'|' 'NR==FNR{b[$1"|"$2]=1;next} b[$1"|"$3]{print $1"|"$3}' bad_targets.txt live_anchors.txt \
  | sort | uniq -c | sort -rn | head -25                                        # worst targets
```

> **Gotcha — a 301 is not a break.** relojistas.com and idea.uk redirect several extension-less
> paths at the Cloudflare edge and they resolve. 5 of 45 extension-less candidates were fine for
> this reason. Classify on the **status code**, never on the shape of the path.

> **Gotcha — `site_components` is trustworthy where `page_components` is not.** The stale chrome
> in `049` matched the live artefact exactly on all three sites; the page-level staleness above did
> not. Do not generalise "stored HTML is stale" into "ignore stored HTML" — check which table.

## R16 — List-section empty-state audit (`bugs_open/054`, migration 185)

The standing check that a list component degrades gracefully when its query-sourced
`items` resolves empty. Same coarse regex the bug uses.

```sql
SELECT function, (html_template ~ '\{\{ *if [^}]*items *\}\}') AS has_if_guard
  FROM content_components
 WHERE is_active AND html_template LIKE '%{{range .items}}%'
 ORDER BY has_if_guard, function;
```

`has_if_guard=f` ⇒ the component ranges over `items` with no enclosing guard, so an empty
list renders a **blank container**. Fix per migration 185: wrap in
`{{if .items}}{{range .items}}…{{end}}{{else}}<p class="…-empty">{{if .empty_state_text}}{{.empty_state_text}}{{else}}<English fallback>{{end}}</p>{{end}}`
and add an `empty_state_text` `source:llm` field (translatable — do NOT hardcode English,
bugs_open/026). Result 2026-07-21 after 185: **7 of 7 guarded** (was 5 of 7).

Or run the packaged lint (advisory, exit 1 on any unguarded component):

```bash
python3 scripts/check_list_empty_states.py
```

> **Gotcha — the guard regex is coarse.** It proves *some* `{{if …items…}}` exists in the
> template, not that it **encloses the range**. Good enough to flag a candidate; read the
> template before editing. A precise check needs the block tokeniser (`scripts/parse_gates.py`
> shape).

> **Gotcha — do NOT "fix" this in the resolver.** `min_items:1`/`required:true` on a
> query-sourced array is silently ignored: `plan_sections_action.go:1288-1321` `continue`s
> before the required-branch, so the empty list always reaches the template. Changing that is
> `bugs_open/054` fix-candidate 2 — a separate, riskier change (it can mask a real "resolver
> errored" as "empty"). This runbook entry is only about the template guard.
