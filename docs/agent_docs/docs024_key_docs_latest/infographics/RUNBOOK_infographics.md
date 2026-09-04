# RUNBOOK — infographics

Full path: `docs/agent_docs/docs024_key_docs_latest/infographics/RUNBOOK_infographics.md`
Every command here was needed to get an answer right, with its gotcha attached. Change it HERE.

DB access: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## 1. The two-route census — ALWAYS RUN BOTH ARMS

⚠ **Running only arm A returns `1` and reads as "the estate has no infographics". It is one of two
routes and it is the minority one by 45×.** Four sessions have made this exact measurement and
stopped there.

```sql
-- ROUTE A: the diffusion picture
SELECT kind, count(*) AS rows, count(DISTINCT sp.site_id) AS sites
  FROM site_plan_imagery spi JOIN site_plans sp ON sp.id = spi.plan_id
 GROUP BY 1 ORDER BY 2 DESC;

-- ROUTE B: the code-rendered component (the answer that is actually shipping)
SELECT cc.name, count(pc.id) AS instances, count(DISTINCT p.site_id) AS sites,
       min(pc.created_at)::date AS first_use, max(pc.created_at)::date AS last_use
  FROM content_components cc
  LEFT JOIN page_components pc ON pc.component_id = cc.id
  LEFT JOIN pages p ON p.id = pc.page_id
 WHERE cc.name IN ('checklist','comparison-table','period-calendar',
                   'mechanism-flow','evidence-chart','evidence-timeseries')
 GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **Schema traps, both hit on 2026-09-04:** `content_components` has **no `site_id`** column (a
component is global; do not write `site_id IS NULL`), and `page_components` has **no `deleted_at`**
(do not add a soft-delete predicate — it errors, and "fixing" it by dropping the join silently
changes the population).

**Keep the component list in step.** It is hand-maintained here and in the PLAN. A new structured
component that nobody adds makes route B read low, which is the same failure this runbook opens
with. Source of truth for what exists:
`docs/agent_docs/docs026_concept_register/register/visualisation-and-charts.md` (VIZ-001…019).

## 2. Is the instruction even exercisable? — the disjoint-sets check

The reason "zero infographics since 718" explains nothing. **Re-verified first-hand 2026-09-04**;
the `framework_prompts_positive_voice` lane reached it independently.

```sql
-- (a) sites CAPABLE of an infographic: a current plan AND registered facts
SELECT count(*) FROM sites s
 WHERE EXISTS (SELECT 1 FROM site_plans sp WHERE sp.site_id=s.id AND sp.is_current)
   AND EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id=s.id AND ss.aspect='evidence_base'
               AND ss.is_current AND jsonb_array_length(COALESCE(ss.data->'facts','[]'::jsonb))>0);
-- → 21  [MEASURED 2026-09-04]

-- (b) of those, how many have planned ANY imagery since migration 718 (2026-09-02)
SELECT count(DISTINCT sp.site_id)
  FROM site_plan_imagery spi JOIN site_plans sp ON sp.id=spi.plan_id
 WHERE spi.created_at >= '2026-09-02'
   AND EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id=sp.site_id AND ss.aspect='evidence_base'
               AND ss.is_current AND jsonb_array_length(COALESCE(ss.data->'facts','[]'::jsonb))>0);
-- → 0   [MEASURED 2026-09-04]  ⇒ the two sets are DISJOINT
```

⚠ **`jsonb_array_length` on the facts array is the load-bearing part, not `aspect='evidence_base'`.**
Two of the seven planning sites (apis.uk, gamedesign.uk) **carry the aspect with a facts array of
length 0**. A check written as "has an `evidence_base` aspect" therefore counts them as capable, is
constant across the whole comparison, and **cannot come out otherwise** — that exact error produced
a void test on 2026-09-04 that was then quoted by a second lane and built into a causal account.
Both retracted. `WRONG_CALLS.md`, same date.

## 3. Read the live planner prompt — never quote it from a doc

⚠ **A verbatim quotation is a measurement of a mutable string and goes stale exactly like a count.**
Migration 718 replaced the sentence every document was quoting, **on the same day** one lane read
it; that lane then quoted its own reading a day later and it reached the owner as evidence.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A \
  -c "SELECT default_config::text FROM agent_definitions
       WHERE id='f263eaa1-61e1-446e-9410-648e12b7875b';" > planner.txt
wc -c planner.txt   # 39,431 B [MEASURED 2026-09-04]
```

Then unescape before matching — the column is JSON, so `\n` and `\"` are literal two-character
sequences and a multi-line regex silently matches nothing:

```python
t = open('planner.txt').read().replace('\\n','\n').replace('\\"','"')
```

Confirm the row is the one you mean *by content*, not by id alone: `grep -c sparingly` must be **0**
(the pre-718 anchor) and `Content-carrying imagery is EXPECTED` must be present.

## 4. Verify at the served artefact, with a control

A `build_status='deployed'` row is not a rendered graphic.

```bash
curl -s --max-time 25 "https://<domain>/<path>" -o /tmp/p.html -w "HTTP %{http_code} bytes=%{size_download}\n"
grep -o 'class="[^"]*checklist[^"]*"' /tmp/p.html | sort -u      # or comparison-table__, mechanism-flow__
curl -s --max-time 15 -o /dev/null -w "CONTROL %{http_code}\n" "https://<domain>/this-does-not-exist-xyz.html"
```

⚠ **The control is not optional: a parked or catch-all domain 200s every path**, so a bare 200 on
the target proves nothing. Worked example 2026-09-04:
`websitepromotion.co.uk/blog/website-launch-promotion-checklist.html` → 200 / 80,415 B / 48 `<li>` /
`checklist__item` present; invented sibling → **404**.

## 5. Before routing work at a neighbour

```bash
./scripts/who-owns.py <number|slug>          # advisory, ~0.3s, reads COMMITS
grep -rn "<mechanism>" bugs_open/ bugs_closed/ | head
```

⚠ `who-owns.py` is **blind to uncommitted sessions** — a lane mid-fix looks unowned. `bugs_open/114`
is **owned and active** (14 commits in the last 3 days): contribute into the file, do not compete.
Re-run at every phase boundary; every ownership check is lagging.

## 6. Lane-local conventions

- **Route A** = diffusion `site_plan_imagery.kind='infographic'`. **Route B** = code-rendered
  component. Always say which. A bare "infographic" in this lane's docs means route A only.
- Any count carries the date it was counted (`**N** as of <date>`), per the owner ruling.
- This lane **specifies**; peers **build**. See PLAN §4 for the "will not do" list.
