# RUNBOOK — directory build handler / page-builder routing (`bugs_open/206`)

The commands this lane had to get right, with the gotcha attached to each. Created 2026-08-24
(the lane had run since 08-06 without one — a gap in the standing five). **Update commands
HERE when they change, not in your scrollback.**

---

## 1. Census the defect class — and the WRONG way to do it

The class is "a build item reached a handler that cannot build the page". Its signature is the
handler's own error text.

```sql
-- CORRECT: join to pages.page_type, dedup on the item id
WITH noop AS (
  SELECT DISTINCT ON (swi.id) swi.id, swi.site_id, swi.status, swi.created_by,
         p.page_type,
         (p.sections IS NULL OR jsonb_array_length(COALESCE(p.sections,'[]'::jsonb))=0) AS layoutless
  FROM site_work_items swi
  LEFT JOIN pages p ON p.site_id=swi.site_id
       AND (p.id=swi.page_id OR p.name=swi.spec->>'page_name')
  WHERE swi.error ILIKE '%no-op: no sections ready to build%'
  ORDER BY swi.id, (p.id=swi.page_id) DESC NULLS LAST
)
SELECT COALESCE(page_type,'(no pages row)') AS page_type, count(*) AS items,
       count(*) FILTER (WHERE layoutless) AS layoutless,
       count(*) FILTER (WHERE status='needs_human_review') AS parked,
       count(DISTINCT site_id) AS sites
FROM noop GROUP BY 1 ORDER BY 2 DESC;
```

> ⚠ **Do NOT filter on `swi.spec->>'page_type'`.** `reconcile_site_plan`'s mint writes a spec of
> `{page_name, page_role, plan_id, reason}` with **no `page_type` key**, so that filter returns a
> confident **zero** for the population it exists to count. It did, in this lane, and the zero
> reached a council submission (`WRONG_CALLS` 2026-08-24).
>
> ⚠ **`DISTINCT ON` is load-bearing.** The `page_id OR name` join fans out — the naive version
> returned 134 for a population of 87.
>
> ⚠ **Two causes wear this one error string.** A `layoutless` page has no layout for any builder
> to fill (this bug). A page WITH a layout whose sections are all deferred fails the same way for
> an unrelated reason — 28 of 29 `content`-type hits were that. Split on `layoutless` before
> theorising.

**Prior art before you file anything**: grep by the SYMPTOM, not the bug number — this population
is named under four different numbers, and `who-owns.py` cannot find it.

```bash
grep -rln "no sections ready to build" bugs_open/ bugs_closed/
```

---

## 2. Is a handler real? (the load-bearing existence check)

Never assert that an agent exists, or that one does not, without this. Both directions matter:
routing to a non-existent handler mints dormant work; treating an existing one as missing parks
buildable pages.

```sql
SELECT type, is_active, COALESCE(is_snapshot,false) AS snap, deleted_at IS NULL AS not_deleted
FROM agent_definitions
WHERE type IN ('directory-build-handler','page-build-handler','tool-builder','entity-page-builder')
ORDER BY type;
-- 2026-08-24: directory-build-handler|t|f|t · page-build-handler|t|f|t
--             tool-builder, entity-page-builder → NO ROWS (the absence is real)
```

The three live filters are `is_active`, `NOT is_snapshot`, `deleted_at IS NULL` — a row alone is
not enough.

---

## 3. Can a `deferred` row be dispatched? (and how to check any status claim)

Two independent gates filter status, and **both** must be read — `claim_work_item_action.go:102`
and `load_work_item_actions.go:711`, each `AND status IN ('triaged', 'approved')`.

Then the live demand control — but note what it does and does not license:

```sql
SELECT count(*) AS deferred_with_handler,
       count(*) FILTER (WHERE attempt_count > 0) AS ever_attempted,
       count(DISTINCT handler_agent) AS distinct_handlers
FROM site_work_items WHERE status='deferred' AND COALESCE(handler_agent,'') <> '';
-- 2026-08-24: 262 | 0 | 16
```

> ⚠ **This control varies STATUS and is uniform on REGISTRATION** — all 16 of those handlers are
> registered agents. It says nothing about `deferred` + an *unregistered* handler, which was the
> actual question a council seat asked. Add the registration axis explicitly:
> ```sql
> SELECT swi.handler_agent, count(*),
>        (SELECT count(*) FROM agent_definitions ad
>         WHERE ad.type=swi.handler_agent AND ad.is_active AND ad.deleted_at IS NULL) AS registered
> FROM site_work_items swi
> WHERE swi.status='deferred' AND COALESCE(swi.handler_agent,'')<>'' GROUP BY 1;
> ```

---

## 4. Re-triage a parked build item — **AFTER the roll, never before**

The routing fix changes what is MINTED. It does not touch the rows already parked: they hold
their `needs_page:<name>` key, `needs_human_review` is non-terminal, so `idx_swi_dedup` covers
them and reconcile skips them as "queued" (`loadOpenPageItems`' status filter excludes only
complete/verified/rejected/wont_fix/failed/cancelled).

So the five legacy rows need one manual re-triage each. **Do not run this before the image
carrying `builderForPageType` is live** — the row still names `page-build-handler`, so a
re-triage now re-runs the no-op and burns an attempt of three.

```sql
-- 1. LOOK first (never re-triage blind — read the error and the attempt count)
SELECT id, status, handler_agent, attempt_count, max_attempts, left(error,120)
FROM site_work_items WHERE id = '<item>';

-- 2. Re-point and re-queue
UPDATE site_work_items
SET handler_agent = 'directory-build-handler', status = 'triaged', error = NULL
WHERE id = '<item>' AND status = 'needs_human_review';
```

Then let `build-pipeline-trigger` (120s) pick it up — **do not hand-dispatch**; the point of the
fix is that these are ordinary items.

> ⚠ **`priority ASC` — LOWER dispatches FIRST** (`load_work_item_actions.go:750`,
> `ORDER BY wi.priority ASC, wi.created_at ASC`). Priority dominates within a site; `created_at`
> only breaks ties. "Bumping" a number UP starves the item — it cost this lane 45 minutes on
> 08-08 and a wrong correction to a peer on 08-24. Planner-minted build items sit at `10+i`, so
> compare against the competing ROWS, never against the column default of 100.
>
> ⚠ **No dispatch within ~300s of a chassis pod restart** — the spawn is silently dropped.

Watch the PAGE, not the status:

```sql
SELECT name, build_status, deployed_at FROM pages
WHERE site_id = '<site>' AND name = '<page>';
```

...then `curl` the URL and read it. `complete` is not proof.

---

## 5. Prove the code is actually live

**Do NOT use build provenance for this** (a same-tag rebuild serves the node's stale cached
binary, and the `build provenance` startup line scrolls out of reach within hours). Ask the
running binary for this change's own symbols, with controls in the same breath:

> ⚠ **Enumerate the pods properly — a label selector can miss most of them** (council round 6,
> `debug_historian`; the documented case is `-l app=<x>` returning 2 pods of 41 that actually run
> the binary, because one image serves many labels). Count what you got before trusting a clean
> grep, and check **every** pod it returns, not the first:

```bash
# see what you are actually about to check — and how many there are
kubectl -n ai-persona-system get pods -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image \
  | grep agent-chassis
# then probe EVERY one, not `head -1`
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  P=${POD#pod/}; echo "== $P"
  # must be PRESENT (the symbol, and the route literal it writes)
  kubectl -n ai-persona-system exec $P -- grep -ac 'builderForPageType' /proc/1/exe
  kubectl -n ai-persona-system exec $P -- grep -ac 'directory-build-handler' /proc/1/exe
  # NEGATIVE CONTROL — must be 0, or a probe that matches everything proves nothing
  kubectl -n ai-persona-system exec $P -- grep -ac 'builderForPageTypeXYZZY' /proc/1/exe
done
```

> ⚠ **Never `strings`** — absent from the debian-slim image, and behind `2>/dev/null` its failure
> is indistinguishable from "not stamped". ⚠ **Check every replica**, not one: a partial roll
> makes them disagree. ⚠ Per SERVICE, not per fleet.

**The best behavioural proof is on a site nobody set up for it.** `garden-tools.uk` is a live,
unaided greenfield build (the `loanzy_uk_example_site` lane, 2026-08-23/24), deliberately left
unrepaired: its `/brand-directory/index.html` is an `entity-directory` page **linked from its own
home page**, and that link is one of 9 measured dead links on the site. Post-roll, that link
should go live **without anyone touching the site**:

```bash
curl -s -o /dev/null -w '%{http_code}\n' https://garden-tools.uk/brand-directory/index.html
```
404 before the roll, 200 after, with no hand re-triage — that is the closure proof, and it is
stronger than re-triaging a page yourself because nothing about it was arranged to succeed.
⚠ Its `/buying-guides/index.html` is `section-index` and will STILL 404 after the roll — that is
the deliberate narrowing, not a failure of the fix.

Then the queue-side proof: a `reconcile_site_plan` run that mints a typed page's item carrying
the right handler.

```sql
SELECT spec->>'page_name', spec->>'page_type', handler_agent, item_type, status, created_at
FROM site_work_items
WHERE created_by='reconcile_site_plan' AND created_at > '<the roll>'
ORDER BY created_at DESC LIMIT 20;
```
A `section-index` or `entity-directory` row at `directory-build-handler` = live. An
`entity-page`/`tool` row as a `deferred` `capability_gap` with an **empty** `handler_agent` = the
gap arm live.

---

## 6. Council gate — the two things that cost time

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <sub.json>
RESUBMIT_CORR=<corr> ./docs/.../097_TRIGGER_council_review_v1.sh <sub.json>   # same corr, trail accumulates
```

Read the verdict properly — the `doc_notes` summary TRUNCATES each objection mid-sentence, and
the severity is not in it. Go to the report body:

```sql
SELECT r->>'reviewer', o->>'severity', o->>'edit', o->>'problem'
FROM (SELECT body::jsonb AS b FROM diagnosis_artifacts
      WHERE correlation_id='<corr>' AND kind='council_report'
      ORDER BY created_at DESC LIMIT 1) x,
     jsonb_array_elements(x.b->'reviews') r,
     jsonb_array_elements(COALESCE(r->'objections','[]'::jsonb)) o
WHERE o->>'severity' IN ('high','medium') ORDER BY 2,1;
```

> ⚠ `metadata` holds only the tally (`decision`, `reviewers`, `abstained`); **the reviews live in
> `body`**, which is `text` and needs `::jsonb`.
> ⚠ The verdict note's "reviewers' checks, answered" block is worth reading even when you
> disagree with the objections — this lane's round-1 checks returned a population **8× larger**
> than its own census and that discrepancy led to the actual bug.
