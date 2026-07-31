# RUNBOOK — chrome component selection (bugs_open/118)

Every query below was run live on 2026-07-31 and each one is here because it was
either hard to get right or it changed what I believed.

## R1 — What each of the three predicates picks, side by side

This is the query that makes the bug undeniable in one screen. Run all three
against the same library; if they disagree, that IS the defect.

```sql
-- A: render_site_components' old fallback — NO predicate at all
SELECT 'A' AS site, name, is_active, forked_from IS NOT NULL AS forked, component_level
FROM content_components WHERE function='site-header' ORDER BY name LIMIT 1;

-- B: link_site_components' old lookup — is_active only
SELECT 'B', name, is_active, forked_from IS NOT NULL, component_level
FROM content_components WHERE function='site-header' AND is_active ORDER BY name LIMIT 1;

-- C: GetComponentByFunction — active + unforked, and (until this fix) NO ORDER BY
SELECT 'C', name, is_active, forked_from IS NOT NULL, component_level
FROM content_components WHERE function='site-header' AND is_active AND forked_from IS NULL LIMIT 1;
```

2026-07-31: A → `header-bold-gradient` (`is_active=f`), B → `header-leopardess`
(a FORK), C → `site-header` (`component_level='section'`). Three predicates,
three answers, three different kinds of wrong.

**Gotcha.** C has no `ORDER BY`, so what it returns is whatever the plan happens
to give. Do not record its answer as a fact without saying that — run it a few
times, and read it as "what it happens to return today", which is exactly the
property the fix removes.

## R2 — The chrome pool, with the column that decides it

```sql
SELECT function, component_level, name, is_active, forked_from IS NOT NULL AS forked
FROM content_components
WHERE function IN ('site-header','site-footer','head')
ORDER BY function, component_level, name;
```

**The column people miss is `component_level`.** `site-header` and `site-footer`
are `component_level='section'` — page-section components that merely share the
chrome function name. `head` has exactly one `'head'`-level row and it is
`is_active=false`, which is why `head` has no eligible chrome at all.

The whole vocabulary, because the chrome half of it is not obvious:

```sql
SELECT component_level, count(*) FROM content_components GROUP BY 1 ORDER BY 2 DESC;
-- section 127 | tool 60 | site 12 | header 5 | footer 1 | head 1 | element 1
```

Chrome is `('site','header','footer','head')` — four values, because the
vocabulary grew twice. `'section'`, `'tool'` and `'element'` are page-body levels.

## R3 — Blast radius of an ORDER BY: measure it, never argue it

Before adding `ORDER BY name` to a `LIMIT 1`, prove which functions even have a
choice to make, and prove the ordered answer equals today's answer:

```sql
WITH elig AS (
  SELECT function, name FROM content_components WHERE is_active AND forked_from IS NULL
), multi AS (
  SELECT function FROM elig GROUP BY 1 HAVING count(*) > 1
)
SELECT m.function,
       (SELECT count(*) FROM elig e WHERE e.function=m.function) AS eligible_rows,
       (SELECT c.name FROM content_components c
         WHERE c.function=m.function AND c.is_active AND c.forked_from IS NULL LIMIT 1) AS today_unordered,
       (SELECT c.name FROM content_components c WHERE c.function=m.function
         ORDER BY (c.is_active AND c.forked_from IS NULL) DESC, c.name LIMIT 1) AS new_ordering
FROM multi m ORDER BY 1;
```

2026-07-31: **2 rows** — `site-header` and `site-footer` — and `today_unordered`
= `new_ordering` for both. That is what makes "the fleet's answer is unchanged"
a measurement rather than a hope.

## R4 — Which sites the selection code can actually reach

The single most important query in this file, because it is what refuted the
filed bug's "changes the rendered footer on every site".

```sql
SELECT s.domain,
  max(CASE WHEN sc.slot_name='header' THEN 1 ELSE 0 END) AS hdr,
  max(CASE WHEN sc.slot_name='footer' THEN 1 ELSE 0 END) AS ftr,
  max(CASE WHEN sc.slot_name='head'   THEN 1 ELSE 0 END) AS hd
FROM sites s LEFT JOIN site_components sc ON sc.site_id=s.id
GROUP BY 1 ORDER BY 1;
```

A site with a row for a slot never reaches the selection code — the row pins the
choice by `component_id`. 2026-07-31: all 14 real sites have all three;
`loancalculator.co.uk` (created 2026-07-30) has none; the rest are
`pool-*.internal` / `system.internal` placeholders.

**Gotcha.** `sites` has **no `deleted_at` column** — `WHERE deleted_at IS NULL`
errors. Count from `sites` directly and read the `pool-*.internal` rows as the
placeholders they are.

## R5 — What each site is actually pinned to

```sql
SELECT sc.slot_name, cc.name AS component, cc.is_active, cc.component_level, count(*) AS sites
FROM site_components sc JOIN content_components cc ON cc.id = sc.component_id
GROUP BY 1,2,3,4 ORDER BY 1,5 DESC;
```

2026-07-31: 11 footers on `footer-4-column` (`is_active=f`), 7 headers on
`header-bold-gradient` (`is_active=f`), 9 heads on `Document Head`
(`is_active=f`). **This is the residual the code fix does not touch.**

## R6 — The detector that already knows, and the repair that cannot repair

```sql
SELECT id, status, left(summary,80), created_at::date
FROM site_work_items
WHERE item_type='deactivated_component' AND status NOT IN ('complete','cancelled','rejected')
ORDER BY created_at DESC;
```

Items since 2026-07-17, two of them `[unresolved after 2 attempts]`. Their
`HandlerAgent` is `rerender-pages`, which re-renders **the component the row
already points at**. Read that before concluding nobody noticed — somebody's
code noticed, three months of items ago.

## R7 — Verify at the artefact, never at the library

A chrome fix is invisible until `render_site_components` runs (`bugs_open/117`),
so the library and the page can disagree for days. Ask the page:

```sh
curl -s https://relojistas.com/index.html | grep -o '<h[34]>[^<]*</h[34]>'
# 'Our Services' -> footer-4-column (is_active=false)
# 'Explore'      -> footer-theme-chrome (is_active=true)
```

Pick a string only ONE candidate carries. `<h3>{{.logo_text}}</h3>` is in both
`footer-4-column` and `footer-theme-chrome`, so it discriminates nothing:

```sql
SELECT name, substring(html_template from '<h[34]>[^<]*</h[34]>') FROM content_components
WHERE function='site-footer' ORDER BY name;
```

## R8 — Building and testing on a shared tree

The working tree carries other sessions' in-flight edits, so `go test ./...` in
the repo can fail for reasons that are nothing to do with you (2026-07-31:
`save_page_sections_action.go:499 declared and not used: floorDetail`, another
lane mid-edit). `make build-*` builds from committed HEAD, so test what will
actually ship:

```sh
SP=$HOME/.cache/ac-118-headtest; rm -rf "$SP"; mkdir -p "$SP"
git archive HEAD | tar -x -C "$SP"
cp <your changed files> "$SP/<same paths>"
cd "$SP" && go build ./cmd/agent-chassis/ && go test ./platform/orchestration/actions/ -count=1
```

**Gotcha, and it cost a confusing half hour:** do NOT put this under `/tmp` on
this box. `/tmp` is a 16G tmpfs and was at 94%; the extraction silently produced
a **0-byte `go.mod`** and `go` reported *"missing module declaration"*, which
reads like a repo problem and is a disk-full problem. `wc -c "$SP/go.mod"` right
after the extract is the cheap check.

`cmd/reasoningset` does not build at HEAD (pre-existing, `declared and not
used`), so `go build ./...` is red for reasons that predate any of this. Build
`./cmd/agent-chassis/` specifically.

## R9 — Proving a test is not vacuous

Delete the thing the test is about, watch it go red, put it back:

```sh
cp platform/orchestration/actions/component_library.go /tmp/cl.bak
# remove both ORDER BY clauses
go test ./platform/orchestration/actions/ -run 'Chrome|GetComponentByFunction' -count=1
#   --- FAIL: TestResolveChromeComponentOrdersEligibleFirstThenByName
#   --- FAIL: TestGetComponentByFunctionIsOrdered
cp /tmp/cl.bak platform/orchestration/actions/component_library.go   # green again
```

## R10 — Council submission

```sh
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
# SUBMISSION_CORR for this change: 5bc232d6-590a-4476-a6b1-4fb6f61751c6
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<corr>';
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

## R11 — Proving the fix is LIVE (the council's debug_historian objection, and it was right)

The submission carried no deploy-verification step. This is it. **Never git, never
the image tag** — a same-tag rebuild ships the node's stale cached binary, and a
roll is not evidence your commit is in the image (`bugs_open/153`).

Grep the RUNNING pod's binary for a symbol this change ADDED, **with a positive
control in the same exec** so a zero can be told from a broken grep:

```sh
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system exec "$POD" -- sh -c '
  echo -n "ResolveChromeComponent (NEW, want >0): "; strings /app/agent-chassis | grep -c "no eligible component for function"
  echo -n "chrome level whitelist (NEW, want >0): "; strings /app/agent-chassis | grep -c "component_level IN (.site., .header."
  echo -n "positive control (want >0):            "; strings /app/agent-chassis | grep -c "RenderSiteComponentsAction"
'
```

A zero on the first two with a non-zero control means the image predates the
commit. A zero on all three means the grep is wrong, not the deploy — that is
what the control is for. **Run it on every replica**, not one: `logs deploy/X`
and a single `exec` both read one pod of N.

Then, and only then, the behavioural proof — build a site that has no chrome
assignment and read which component it got:

```sql
SELECT sc.slot_name, cc.name, cc.is_active, cc.component_level
FROM site_components sc JOIN content_components cc ON cc.id=sc.component_id
JOIN sites s ON s.id=sc.site_id WHERE s.domain='loancalculator.co.uk';
-- before this fix: header-bold-gradient / footer-4-column / Document Head, all is_active=f
-- after:           header-theme-chrome  / footer-theme-chrome, both is_active=t
--                  head stays Document Head — no eligible head component exists (by design, logged at ERROR)
```

`loancalculator.co.uk` is the whole live test population, which is the point: this
change cannot touch a site that already has an assignment.
