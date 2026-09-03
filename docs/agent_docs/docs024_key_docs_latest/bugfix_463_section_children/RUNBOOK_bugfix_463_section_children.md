# RUNBOOK — bugfix 463

Every command here had a gotcha attached. The gotcha is the point.

## Prove the bug is still live, before touching anything

```sql
-- the fleet population 463 holds unfillable. NOT "hubs Pass C emptied" - it is
-- hubs that CANNOT be filled while the bug stands.
WITH hubs AS (
  SELECT p.site_id, p.name, '/' || split_part(ltrim(p.url,'/'),'/',1) || '/' AS prefix
    FROM pages p
   WHERE p.page_type IN ('section-index','blog-index','news-index','entity-directory')
     AND p.url LIKE '/%/index.html')
SELECT count(*) AS hubs, count(*) FILTER (WHERE kids = 0) AS childless
  FROM (SELECT h.*, (SELECT count(*) FROM pages c
                      WHERE c.site_id=h.site_id AND c.url LIKE h.prefix||'%' AND c.name<>h.name) AS kids
          FROM hubs h) t;
-- 2026-09-03: 78 hubs, 53 childless, across 21 sites.
```

## ⚠ The census trap that cost me an inverted conclusion

Modelling a Go helper in SQL: **count the Go function's branches first and give your
predicate the same number.** `sectionStemOf` takes the first URL segment *only when the
trimmed url contains a `/`*, and otherwise falls back to the name. A single `split_part`
expression applies the first arm to rows that take the second and returns a plausible
non-zero answer. Mine said "5 asymmetric hubs"; the truth was 0 of 83. Full account in
`WRONG_CALLS.md`, 2026-09-03.

```sql
-- correct: encode BOTH branches
WITH h AS (SELECT name, url, btrim(url,'/') AS trimmed,
                  regexp_replace(name,'-index$','') AS name_stem FROM pages
            WHERE page_type IN ('section-index','blog-index','news-index','entity-directory'))
SELECT count(*) FILTER (WHERE go_stem = path_seg) AS agree,
       count(*) FILTER (WHERE go_stem <> path_seg) AS disagree
  FROM (SELECT CASE WHEN position('/' in trimmed) > 0 THEN split_part(trimmed,'/',1) ELSE name_stem END AS go_stem,
               CASE WHEN position('/' in trimmed) > 0 THEN split_part(trimmed,'/',1)
                    ELSE regexp_replace(trimmed,'\.html?$','') END AS path_seg FROM h) g;
```

## Is the field the planner would need even asked for? Read the LIVE row, never the seed

```sql
SELECT (default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template') ILIKE '%parent_section%',
       length(default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template')
  FROM agent_definitions
 WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- 2026-09-03: false, 32191. The prompt has no such field, so the value can only ever be absent.
```

## ⚠ orchestration_states censuses time out — scope them or they never return

`collected_data ? 'validate_plan'` over the whole table is a full scan with a TOAST detoast
per row (~9,600 rows per 2 days). Two attempts died at the 2-minute tool limit. Filter on
`updated_at` (the only useful index) and keep the window to a couple of days, or answer the
question from `site_plan_pages` instead, which is small.

## Mutation-test a guard, don't assert its absence

A rule about NOT acting cannot be proven by an expectation that nothing happened — a test
that passes because the branch was never reached looks identical to one that passes because
the guard held.

```bash
# Pass C: force the old first-segment rule back and confirm the suite reproduces the bug
cp platform/orchestration/actions/v3_site_actions.go /tmp/.bak   # NOT git stash - it is banned and blocked
# ...edit the branch to `if false`...
go test ./platform/orchestration/actions/ -run TestPassC -count=1
# expect: dropped_collision = 5, naming all five gamedesign articles; every KEEP case fails
cp /tmp/.bak platform/orchestration/actions/v3_site_actions.go
```

For the `RealisedIdentity` guard the mutation is in the test itself — same input twice, marker
set and cleared, asserting the url does not move and then that it does. See
`TestValidateRoles_RealisedEntryIsNotDerived`.

## ⚠ Choosing a test fixture: /guides/, /tools/ and /games/ are NOT neutral

`ValidateRoles` rule 5 (`nestedRoleFromURL`) retypes a page under those three directories
before rule 6 is reached, so a fixture there measures rule 5, not your change. I used
`/guides/welcome.html` first and it silently tested the wrong thing. Use `/articles/`.

## Verify HEAD, and check whether HEAD is ALREADY red before blaming yourself

```bash
scripts/verify-head-builds.sh --test ./platform/orchestration/...          # HEAD alone
scripts/verify-head-builds.sh --test --with <file> [--with <file>] ./...   # HEAD + your change
```
On 2026-09-03 HEAD was already failing two tests (`TestFindingCodeScanEveryWriteIsRegistered`,
`TestTemplateExecutorsAreDeclared`) from another lane's commit `83407cd37`. Running HEAD alone
first is what separated "my change broke it" from "it was already broken" in one step.
**Never hand-roll `git archive HEAD | tar`** — that recipe is why this box runs out of space.

## Council gate

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```
Dry run tests admission for free. Budget ~30 minutes, not ~2 — the council takes 2–5 but the
dispatch queues behind the fleet. Commit with `Council-Submitted: <corr>`; never write
`Council-Reviewed:` on a verdict you have not read.

## After appending to LANDMINES.md

```bash
./scripts/landmines-verify-dispatch.sh      # sync AND arm the verifier
```
Not `landmines-sync.py --apply` — that consumes the "new entry" status first, and the verifier
then never checks your entry.
