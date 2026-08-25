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

> ⚠ **`spec->>'page_name'` IS NOT A UNIQUE PREDICATE — always filter by producer too.** Several
> producers file rows for the same page, with **different spec shapes**, and an unordered
> `LIMIT 1` over a page-name predicate is a *sampling method, not a lookup*. `[MEASURED
> 2026-08-24]` on one page of one site it matches **3 rows from 2 producers**:
>
> | producer | item_type | spec keys |
> |---|---|---|
> | `reconcile_site_plan` | `needs_page` | `page_name, page_role, plan_id, reason` |
> | `rerender-pages` | `page_rerender` | `domain, filename, page_id, page_name` |
>
> This is not hypothetical: a peer lane read the `rerender-pages` shape and reported it as the
> reconcile shape, then proposed a closure-test fix built on `spec->>'page_id'` — a key present
> on the rerender rows and on **0 of 134** reconcile rows. Add `AND created_by='<producer>'` to
> every lookup, and **when you describe a population, put the population in the `WHERE`.**
>
> The reconcile-side type discriminator is **`spec->>'page_role'`** (present 134/134); the
> authority is `pages.page_type` via a join on `(site_id, spec->>'page_name')`.

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
> **⚠ CORRECTED 2026-08-24 POST-ROLL — this check as written CANNOT FIRE, and it was wrong in
> two lanes' docs at once.** Measured after v1.0.1334 went live: the page is still 404 and will
> stay so. (1) **Reconcile does not run on a cadence** — it runs inside a build/publish
> pipeline, and `sites.last_reconciled_at` for garden-tools.uk was 2026-08-23 20:15, i.e. a
> quiet site never re-reaches the fixed code. (2) **The parked row blocks its own re-mint** —
> the fix routes what is MINTED, and `loadOpenPageItems` counts `needs_human_review` as OPEN, so
> reconcile skips the page as "queued" and the new routing never applies to it.
>
> **The honest proof is the MINT, not the page, and it needs the key freed first:**
> 1. close the parked row to a TERMINAL status (`cancelled`/`wont_fix` — never `complete`, which
>    asserts work that did not happen), reason in `error`, which releases the `item_key` from
>    both `idx_swi_dedup` and `loadOpenPageItems`;
> 2. trigger a build/publish for the site so `reconcile_site_plan` actually runs;
> 3. assert on the NEW ROW: a fresh `needs_page:<page>` carrying
>    `handler_agent='directory-build-handler'`, `created_by='reconcile_site_plan'`, with nobody
>    having set the handler by hand. The page building and the link going live follow from it.
>
> **A hand re-triage (step 4 above) fixes the PAGE and proves NOTHING about this fix** — setting
> `handler_agent` yourself only re-demonstrates that `directory-build-handler` works, which has
> been known since 2026-08-08. Do not report one as the other.
>
> Step 1 is an operator action on another lane's site — ask, do not assume.

~~404 before the roll, 200 after, with no hand re-triage — that is the closure proof.~~

> **OWNERSHIP: this check is THIS lane's, and it is deliberately double-owned.** The
> `loanzy_uk_example_site` lane also carries it (their handoff §3a, with the same procedure) and
> will run it if that session is still alive when the roll lands. **Do not treat that as
> delegation and drop it here** — their own docs record a handoff outliving the work it asked
> for, and a session ending is exactly the failure mode that leaves an owed action unowned. Two
> owners on a one-command check is cheap; nought owners is how this lane's original bug sat for
> fifteen days. Verify at the SERVED page, never at `build_status` or the work item.
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

## 7. Prove the fix at the MINT — the closure query, corrected 2026-08-25

**The version in `HANDOFF_2026-08-24` can be passed by a page a human repaired.** `handler_agent`
is a mutable column and re-pointing a parked row is the documented operator escape hatch, so
`created_by='reconcile_site_plan' AND handler_agent='directory-build-handler'` matches both a row
the fix minted and a row somebody fixed by hand. `[MEASURED 2026-08-25]` all **three** rows in
existence that match it, live `UNION` archive, all history, are hand re-routes.

**Use the mint fingerprint as the GATE.** The fixed emit stamps `page_type` into `spec`; the old
one did not; and an `UPDATE` of `handler_agent` cannot add a spec key.

```sql
-- PASS requires BOTH: the fixed code minted this row, AND it routed correctly.
SELECT s.domain,
       swi.spec->>'page_name'                        AS page,
       COALESCE(p.page_type, swi.spec->>'page_role') AS page_type,
       swi.item_type,
       COALESCE(NULLIF(swi.handler_agent,''),'(empty)') AS handler,
       swi.status,
       CASE
         WHEN COALESCE(p.page_type, swi.spec->>'page_role') IN ('entity-directory','section-index')
              AND swi.handler_agent='directory-build-handler' THEN 'PASS'
         WHEN COALESCE(p.page_type, swi.spec->>'page_role')='entity-page'
              AND swi.item_type='capability_gap' AND COALESCE(swi.handler_agent,'')='' THEN 'PASS'
         WHEN COALESCE(p.page_type, swi.spec->>'page_role') IN ('entity-directory','entity-page','section-index')
           THEN 'FAIL'
         ELSE 'n/a'
       END AS verdict
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
LEFT JOIN pages p ON p.site_id = swi.site_id AND p.name = swi.spec->>'page_name'
WHERE swi.created_by = 'reconcile_site_plan'
  AND swi.spec ? 'page_type'          -- ← GATE 1: the fixed emit stamps this; the old one did not.
  AND swi.updated_at < swi.created_at + interval '1 second'
                                      -- ← GATE 2, ADDED 2026-08-25. Proves NOTHING has written to
                                      --   the row since it was minted. Both gates or neither —
                                      --   see the two holes below.
  AND swi.created_at > '<the build start>'
ORDER BY verdict DESC, page;
```

> **⚠ CORRECTED 2026-08-25 (same day), by an adversarial review of this section.** As first
> written, §7 gated on `spec ? 'page_type'` alone and this RUNBOOK called it airtight because
> *"a hand re-route cannot add a spec key"*. That sentence is true and **insufficient**, for two
> independent reasons, and I had the discriminator for both in hand and dropped it — the very
> measurement in `bugs_open/206` §2 that exposed the original defect computes
> `updated_at > created_at + interval '1 second'`, and the shipped query had no `updated_at`
> clause at all.
>
> **Hole 1 — the stamp dates the ROW, not the HANDLER value.** `handler_agent` stays mutable after
> a stamped mint. If the fixed binary ever *mis*-routes — the exact failure this test exists to
> catch — and an operator then repairs the row, it carries the stamp AND the right handler and
> reads `PASS`. The fix takes credit for a human repair, one generation later, in the same shape.
>
> **Hole 2 — this lane's own operator recipe forges gate 1 without touching a spec key.**
> Reconcile's `capability_gap` spec **already contains `page_type`**
> (`reconcile_site_plan_action.go`, the `gapSpec` block), and the recipe eleven lines below it
> (step 3, added by this lane in `0baa8a107`) says: *promote this row in place — set
> `item_type='needs_page'`, `status='triaged'` and `handler_agent` to a handler that can actually
> build it.* A promoted gap row is `created_by='reconcile_site_plan'`, stamped, `needs_page`, at
> `directory-build-handler` — **every §7 PASS condition, entirely by hand.**
> `[MEASURED 2026-08-25]` prospective, not live: **0** stamped `capability_gap` rows exist
> anywhere, so nothing satisfies it today. It arms itself the moment the fixed code files its first
> gap, which is exactly when this test starts being used.
>
> **Gate 2 closes both**, because `trg_site_work_items_updated_at` is a `BEFORE UPDATE … FOR EACH
> ROW` trigger (verified in `pg_trigger` 2026-08-25) — *any* write bumps `updated_at`, so
> `updated_at ≈ created_at` means nothing has touched the row since the emit wrote it.
>
> **The cost of gate 2, stated: it expires.** A legitimately minted row that is then claimed or
> completed also bumps `updated_at`. So **read the mint promptly, while the row is still
> `triaged`** — that is the window in which this proof exists. If `updated_at` has already moved,
> that row cannot serve as proof; find a fresher one rather than relaxing the gate.
>
> **The durable fix is a code change, not a better query:** stamp the routed handler into the spec
> at mint (`"handler": route.handler`) and assert `spec->>'handler' = handler_agent`. Then the
> column carries its own provenance and no operator action can forge agreement. Named as a
> follow-up; not smuggled into an approved change.

Two gotchas attached, both learned the hard way:

- **`section-index` is now a PASS case** (it routes to `directory-build-handler` since
  2026-08-25). The 08-24 version of this query listed it as expected-to-stay-parked. It is not,
  except where §7a bites.
- **Keep the un-gated version for detecting FAIL.** The stamp is absent exactly when the fix did
  not mint the row, so gating on it hides the failures. Run the gated query to confirm success and
  the un-gated one (join `pages.page_type`, read `handler_agent`) to find damage. Two questions,
  two instruments.

**Sanity-check the fingerprint's population is still empty before you trust a first PASS:**

```sql
WITH allrows AS (SELECT created_by,spec FROM site_work_items
  UNION ALL SELECT created_by,spec FROM site_work_items_archive)
SELECT (spec ? 'page_type') AS stamped, count(*) FROM allrows
WHERE created_by='reconcile_site_plan' GROUP BY 1;
-- [MEASURED 2026-08-25] f | 508 ... and no `t` row at all. The first `t` is necessarily the fix.
```

### 7a. Why a parked page may STILL not move: the `unresolved` gap

`loadOpenPageItems` (`reconcile_site_plan_action.go:713`) treats `unresolved` as **open**, so
reconcile skips the page as "already queued" and the new routing never reaches it — while
`idx_swi_dedup` does **not** cover `unresolved`, and both claim gates filter
`status IN ('triaged','approved')`, so the row is undispatchable too. Nothing frees it.
`[MEASURED 2026-08-25]` one live instance: `adversecreditmortgage.co.uk` `blog-index`.
Check before concluding a routing fix failed:

```sql
SELECT item_key, item_type, handler_agent, status, created_at
FROM site_work_items WHERE site_id=(SELECT id FROM sites WHERE domain='<domain>')
  AND item_key = 'needs_page:<page>';
```

## 8. Enumerate every producer of a `needs_page:` item (the "is there another copy?" check)

Three council seats asked this and it is worth having as a command rather than a memory.

```bash
# every Go site that mints the key
grep -rn '"needs_page:' --include=*.go platform/ internal/ cmd/ | grep -v _test.go
# every Go site that names a build handler as a literal (a hardcoded routing answer)
grep -rn "page-build-handler\|directory-build-handler" --include=*.go platform/ internal/ cmd/ | grep -v _test.go
```

`[MEASURED 2026-08-25]` **six** minting sites, of which two consult `builderForPageType` and three
hardcode `page-build-handler` (`rerender_page_sections_action.go:1424`,
`apply_adoption_plan_action.go:731`, `discovery_checks/check_incomplete_page_group.go:202`).
**Do not assume a hardcoded handler is a latent 206** — check whether it actually produces the
signature before filing anything:

```sql
WITH allrows AS (
  SELECT site_id,created_by,item_key,status,error FROM site_work_items
  UNION ALL SELECT site_id,created_by,item_key,status,error FROM site_work_items_archive)
SELECT a.created_by, p.page_type, a.status, count(*) AS rows,
       count(*) FILTER (WHERE a.error ILIKE '%no sections ready to build%') AS with_206_signature
FROM allrows a JOIN pages p ON p.site_id=a.site_id AND p.name = replace(a.item_key,'needs_page:','')
WHERE a.item_key LIKE 'needs_page:%'
  AND p.page_type IN ('entity-directory','section-index','entity-page')
GROUP BY 1,2,3 ORDER BY 5 DESC, 4 DESC;
```

`[MEASURED 2026-08-25]` the three hardcoding producers: 26 typed-page rows, **0** with the
signature. ⚠ **The join reads `pages.page_type` as it is NOW**, not as it was at mint — a page can
be re-typed by hand. That cannot manufacture the zero, but it matters if you recount the totals.

## 7b. ⚠ Gate 1 has EXPIRED, and it never proved what the wording implied (2026-08-25, from the `bugs_open/381` lane's build)

Three corrections to §7, all settled at the commit or in live data. **Read this before running §7.**

### (i) The "population is empty" argument is spent

§7 said the stamp's population was empty (508 reconcile-minted rows, none stamped) so a first hit
was necessarily the fix. `[MEASURED 2026-08-25]` the `bugs_open/381` lane's greenfield build of
`homegarden.uk` minted **21 stamped rows** — `reconcile_site_plan | stamped=t | 21`. The population
is no longer empty. It emptied by **time**, not by absence: no reconcile had run since the roll.
This was always going to expire; §7 said so. It has.

### (ii) The stamp proves the **08-24** code minted the row — NOT that the swap shipped

The wording "the first stamped row is necessarily the fix" was ambiguous about *which* fix, and the
answer matters. `git log -S` settles it: the `"page_type": routeType` stamp was added by
**`d1aa231aa` (2026-08-24 11:50)**, live since `v1.0.1334`. **Today's swap (`efec862f4`) never
touched that emit.** So a stamped row tells you reconcile consulted `builderForPageType`; it says
**nothing** about whether `section-index` routing is live. For that, read the handler on a
`section-index` page — or check the pod's start time against the commit (§7c).

### (iii) The killer: a build can satisfy every gate and still not DISCRIMINATE the fix

`homegarden.uk` carried 21 pages: **17 `section-index`, 2 `content`, 1 `landing`, 1 `blog-post`.
Zero `entity-directory`, zero `entity-page`.** Every one of those types routes to
`page-build-handler` under **both** the old hardcoded literal and the new map. So the mint is fully
*consistent* with the fix and **cannot distinguish it from the bug**. Stamped, untouched, correct —
and worthless as proof.

**So "the proof arrives free on the next greenfield build" is WRONG as written.** It requires a site
carrying an **`entity-directory`** page (routes to `directory-build-handler` — the only type where
old and new disagree today), or an **`entity-page`** (files a deferred `capability_gap` with an
empty handler instead of a doomed dispatch). Check the plan for one of those **before** treating a
build as the closure artefact:

```sql
SELECT page_type, count(*) FROM pages
WHERE site_id=(SELECT id FROM sites WHERE domain='<domain>') GROUP BY 1 ORDER BY 2 DESC;
-- no entity-directory and no entity-page  ⇒  this build cannot close 206, whatever else it shows
```

### (iv) One caveat RETIRED, free, from the same build

§7's worry that the query's `created_by='reconcile_site_plan'` filter might read empty on a
successful build minting through the other door: `[MEASURED 2026-08-25]` on a real greenfield build,
`reconcile_site_plan` minted all 22, and **`WriteBuildItemsAction` did not appear as a `created_by`
at all**. Reconcile is the greenfield door. Keep the check — one build is not a guarantee — but the
filter is not the live hazard it looked like.

## 7c. Does the running fleet carry a given commit? Two timestamps, not a grep

`[LEARNED THE HARD WAY 2026-08-25 — see NOTES misstep 5]`

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o custom-columns='NAME:.metadata.name,IMAGE:.spec.containers[0].image,START:.status.startTime'
git log -1 --format=%cI <your-commit>
```
**A binary cannot contain code committed after it started.** That is free, local, and has no failure
mode. Preferred when it settles the question — and it usually does.

Only if you need finer resolution, read the stamp and test **ancestry**:
`kubectl … logs -l app=agent-chassis --tail=3000 | grep -m1 'build provenance'` then
`git merge-base --is-ancestor <your-commit> <the stamp>`. On a busy service that line has scrolled;
an empty result means "not in range", not "unstamped".

**Three things NOT to do**, each of which returned a confidently wrong answer in one command here:
- **Do not probe a symbol both versions carry.** `builderForPageType` is PRESENT on a binary that
  predates the swap by 31 minutes — it shipped on 08-24 and dates nothing.
- **Do not grep the binary for an ANCESTOR's sha.** A build stamps **one** sha. Grepping for
  `d1aa231aa` returns absent on a binary that demonstrably contains that code. Ancestry is a git
  question, not a grep question.
- **Do not use 40 zeros as a negative control.** It came back **PRESENT** — it matches Go's internal
  tables. The one element in the command that existed to reveal a broken method agreed with it.
