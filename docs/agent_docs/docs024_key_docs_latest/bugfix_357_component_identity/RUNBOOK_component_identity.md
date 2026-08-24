# RUNBOOK — `bugs_open/357` component identity

Every query/command that was hard to get right, with its gotcha attached. Change it HERE, not in
scrollback.

---

## The population (357's own query, unrestricted)

```sql
SELECT s.domain, p.name AS page, pc.created_at::date, pc.slot_name,
       length(pc.rendered_html) AS html_len
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE cc.name = 'hero'
  AND position(left(cc.html_template, position('{{' in cc.html_template) - 1) in pc.rendered_html) = 0
ORDER BY pc.created_at;
```

> ⚠ **This returns 22, while `bugs_open/357`'s own population table lists 9.** The nine are the
> subset that ALSO had a parked `required_fields_missing` work item (all single-component pages).
> Do not read the difference as 13 new rows overnight — check `created_at` before concluding
> anything about rate.

> ⚠ **`left(tmpl, position('{{' in tmpl) - 1)` is empty when a template STARTS with `{{`**, and
> `position('' in anything)` is 1, not 0 — so such a component can never be flagged by this test.
> Silent false-negative, not an error.

## The narrow predicate — use THIS for a guard, not the one above

```sql
WITH x AS (
  SELECT pc.id, cc.name AS comp, s.domain, p.name AS page,
     substring(cc.html_template  from 'data-component="([^"{]+)"') AS tmpl_attr,
     substring(pc.rendered_html  from 'data-component="([^"]+)"')  AS html_attr
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
)
SELECT * FROM x WHERE tmpl_attr IS NOT NULL AND html_attr IS DISTINCT FROM tmpl_attr;
```

> ⚠ **`[^"{]` in the template pattern is load-bearing.** Without excluding `{`, a template whose
> attribute is itself interpolated (`data-component="{{.kind}}"`) captures the Go template
> expression and then "disagrees" with every row it ever rendered. Excluding `{` makes the test
> skip those components instead of convicting them.

> ⚠ **`IS DISTINCT FROM`, not `<>`.** With `<>`, a NULL `html_attr` — which is the whole
> pathological class here (stored HTML carries no attribute at all) — yields NULL and the row is
> silently dropped from the result. Using `<>` returns **0 rows** and reads as "nothing is wrong".

**Always print the agreement count in the same breath** — it is the demand control. A guard census
that reports "0 disagreements" proves nothing unless the same query shows the ~1,550 rows where the
comparison ran and agreed.

```sql
SELECT count(*) FILTER (WHERE tmpl_attr IS NOT NULL AND html_attr = tmpl_attr)          AS agree,
       count(*) FILTER (WHERE tmpl_attr IS NOT NULL AND html_attr IS DISTINCT FROM tmpl_attr) AS flagged,
       count(*) FILTER (WHERE tmpl_attr IS NULL)                                        AS not_testable
FROM x;
```

## Which writer wrote a `page_components` row (fingerprints)

No writer stamps itself, so read the marks it leaves:

| writer | `position` | `content_brief` | `build_status` |
|---|---|---|---|
| `save_page_sections_action.go` | `i+1` | **written** (`"{slot} section"`) | `deployed` |
| `deploy_tool_action.go` | `2` | never | `deployed` |
| `create_tool_component_action.go` | `2` | never | `deployed` |
| `adopt_verbatim.go` | `0` | never | `approved` |
| `create_report_page_action.go` / `rebuild_blog_listing_action.go` | — | check before relying | — |

> ⚠ **`rendered_html_digest` is NOT a writer fingerprint.** The INSERT writes `md5($3)`
> unconditionally; the column postdates older rows (`bugs_open/229` / IMP-052). Rows split by
> digest-present/absent split by **DATE**. I nearly filed that as evidence of a second writer.

## Is a page's tool still being served? (the only proof that counts)

```bash
curl -s "https://<domain>/<page>.html?cb=$(date +%s)" -o /tmp/p.html -w 'http=%{http_code} bytes=%{size_download}\n'
grep -c 'class="tool-page"'      /tmp/p.html   # the tool is present
grep -c 'data-component="hero"'  /tmp/p.html   # 0 here means NO hero rendered at all
```
`bugs_closed/287`: a `complete` work item is not a repaired artefact. Assert the tool's own markup,
not the item status.

## Diagnosis loop for this lane

Intake `f7aedef7-0bee-4c68-8cde-c86ac552e3e2` → **`RUN_CORRELATION_ID=e580b34a-d284-4f80-ac96-81af1c4adaba`**
(the run id is the one the dispatch loop mints and stamps back, NOT the intake id the script prints
first — artifacts are written under the run id).

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<RUN_CORRELATION_ID>';
```
Budget ~30 minutes: the council/diagnosis itself takes 2–5 min, the dispatch queues behind the fleet.
A missing row is latency, not a dropped dispatch — do not retry on that evidence.

---

## ARMING PHASE 2 — the exact order, and why it is that order (written 2026-08-24)

**Nothing below has been run.** Phase 2 is committed, council-APPROVED and confirmed
live in the running binary, but its flag defaults OFF, so the mint is still open.
This is the owner's decision to take, not a session's.

### Precheck — is the code actually there?

Probe the RUNNING binary for the capability, with BOTH control arms. Never `strings`
(absent from the debian-slim images), never a discovery grep for "some 40-hex string".

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for probe in "adopt fragment: bound an unidentified fragment" \
             "self-contained tool section" \
             "zzz_never_shipped_literal_357_qqq"; do
  kubectl -n ai-persona-system exec "$POD" -- grep -aq "$probe" /proc/1/exe \
    && echo "PRESENT: $probe" || echo "ABSENT : $probe"
done
```
Wanted: first PRESENT (the new capability), second PRESENT (positive control — a
pre-existing literal, proving the probe can find anything at all), third ABSENT
(negative control — proving it is not matching everything). **[VERIFIED 2026-08-24
13:40Z: exactly that, on `agent-chassis-698cfdf4f5-2dws6`.]** A probe without both
controls returns the same answer on every service and is worthless.

### Step 1 — seed FIRST, arm second. The order is load-bearing.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/577_seed_adopted_fragment_component_HOLD.sql
```
**Why this order and not the other:** armed-but-unseeded is a REACHABLE state and it
degrades to `component_id` NULL — the honest unknown, but also a NEW population the
council's `bug_historian` seat objected to on `bugs_closed/039`'s precedent (a section
naming no component). Seeding first means adoption succeeds from the first page.
Idempotent (`WHERE NOT EXISTS`), and its verify block RAISEs rather than SELECTs, so a
wrong result aborts instead of committing.

### Step 2 — arm ONE carrier, not all eight, and watch it

Eight live agent types carry `save_page_sections` as of 2026-08-24: `fix-proposer`,
`page-build-handler`, `pageflow-builder`, `page-rebuild`, `page-rerender`,
`required-fields-missing-handler`, `site-work-orchestrator`, `tool-recreation-handler`.
**`tool-recreation-handler` is the right canary** — it is the producer that emits the
`<div class="tool-page">` fragment this whole bug is about, so it exercises the new
path on the first run rather than eventually.

Set `adopt_unidentified_fragments: true` in that agent's `save_page_sections` step
config. Take a snapshot of the row first; DB config is LIVE IMMEDIATELY, with no image
tag to roll back.

### Step 3 — the check that says it worked, and the one that says stop

```sql
-- WORKED: a newly built tool page's first row is adopted, stamped, and STILL named by its plan
SELECT p.name, pc.slot_name, cc.function AS component,
       (pc.content_data->>'body' = pc.rendered_html) AS regenerable,
       (pc.component_version_id IS NOT NULL)         AS stamped
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
 WHERE pc.created_at > now() - interval '1 hour' AND cc.function = 'adopted-fragment';
```
Wanted: `component = adopted-fragment`, `regenerable = t`, `stamped = t`, and
`slot_name` **unchanged** (it will read `hero`, and that is correct — renaming it is
what arms the carry-forward landmine).

**STOP conditions, each invisible to a "is the tool still there?" check:**
- the page's `page_components` **row count goes UP by one** after a rebuild — the
  landmine fired through a path this design believes it never touches;
- a `357` population row appears **WITH** a stamp — the Layer 2 splice hygiene failed;
- a new tool row lands with `component_id` NULL — adoption is failing; read the
  `adopt fragment:` log lines, which say which arm refused and why.

Before/after at the artefact, per page:
```sql
SELECT position, slot_name, md5(rendered_html), component_id FROM page_components
 WHERE page_id = '<page>' ORDER BY position;   -- run BEFORE, and again AFTER one rebuild
```

### Step 4 — only then phase 3

`578_retype_mislabelled_tool_rows_HOLD.sql` re-types the existing rows. It refuses to
run until an organically adopted row with a stamp exists, so steps 1–3 are its
precondition and it enforces that itself rather than trusting this runbook. It targets
**all 22**, the six `rebuild_policy='owned'` pages INCLUDED (owner instruction
2026-08-24, correcting my earlier reading that 'owned' meant a person had claimed the
page — it means the page belongs to a tool, it is set in code, and 172 of 704 pages
carry it). Those six are the ONLY rows phase 2 can never heal: the owned-page guard
returns at `save_page_sections_action.go:186` and adoption runs at `:397`.

**Do not run phase 3 before phase 2 is armed:** the population refilled 12 rows on
08-23, so repairing it first is repairing a set that immediately renews.

### ARMED 2026-08-24 — what was actually done, and how to undo it

Owner instruction: "arm it". Applied by hand, in this order:

```bash
# 1. seed (idempotent; verify block RAISEs, it does not merely SELECT)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/577_seed_adopted_fragment_component_HOLD.sql
# 2. arm — refuses to commit unless the seed exists AND all six steps come back armed
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/579_enable_adopt_unidentified_fragments_HOLD.sql
```

**⚠ THE SECTION ABOVE SAID "ARM ONE CARRIER, NOT ALL EIGHT". THAT ADVICE WAS WRONG
AND IS SUPERSEDED — read this before following it.** It was written before the
surface was enumerated. Enumerating it (recursively; three of six steps are nested
in `sub_workflow`s and a top-level `jsonb_each` misses them):

| carrier | path | `html_field`? | armed |
|---|---|---|---|
| page-build-handler | `workflow,steps,save_sections` | yes | ✔ |
| pageflow-builder | `…build_pages_loop,config,sub_workflow,steps,save_sections` | yes | ✔ |
| page-rebuild | `…build_pages_loop,config,sub_workflow,steps,save_sections` | yes | ✔ |
| site-work-orchestrator | `…build_items_loop,config,sub_workflow,steps,save_sections` | yes | ✔ |
| tool-recreation-handler | `workflow,steps,save_sections` | yes | ✔ |
| page-rerender | `workflow,steps,save_sections` | **no** | ✔ (carry half only) |

**FIVE can mint, not one.** The mint happens on the HTML-parsing path, and the save
falls through to it whenever the metadata path yields nothing — so any carrier with
an `html_field` can reach it. `tool-recreation-handler` is merely the most obvious
(it declares `expects_no_sections_metadata`), not the only one. Arming it alone
would have left four minting: the "one call site gets the rigorous fix while the
mechanism stays generic elsewhere" shape that migration 575's own `bug_historian`
objection names, and that gated this lane's phase 2 round 1.

`page-rerender` cannot mint (no `html_field`) and is armed for the carry half alone —
without it an adopted row loses its identity the first time a rerender carries its
bytes.

**Verify the armed state independently rather than trusting the migration's output:**
```sql
SELECT ad.type, COALESCE(step.value->'config'->>'adopt_unidentified_fragments','(unset)') AS armed
  FROM agent_definitions ad,
       LATERAL jsonb_path_query(ad.default_config, 'strict $.**') AS step(value)
 WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
   AND jsonb_typeof(step.value)='object' AND step.value->>'action'='save_page_sections'
 ORDER BY 1;   -- expect six rows, all `true`
```

**To undo:** `579_..._HOLD_ROLLBACK.sql` removes the key from all six paths; absent
means OFF and there is no third state. ⚠ It does NOT un-adopt rows already adopted —
those are true, regenerable, and their bytes were never touched. They simply stop
having their identity carried, so they drift back toward the mislabelling one
rebuild at a time rather than reverting at once.

**Baseline captured before arming**, for the row-count comparison the landmine
requires: `scratchpad/baseline_before_arming.txt` — population 22, and the per-page
row counts and slot lists for all 22 affected pages. **A page whose row count goes UP
by one is the landmine firing**, and it is invisible to any check that only asks
whether the tool is still present.
