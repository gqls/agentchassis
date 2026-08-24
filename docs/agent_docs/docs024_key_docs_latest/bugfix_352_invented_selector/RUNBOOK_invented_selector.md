# RUNBOOK — bugs_open/352, the invented selector

Every command I had to get right, with its gotcha attached. When one changes, change it HERE.

---

## §1 The census predicate — and why a naive one is wrong

The population is "findings whose selector is `TAG.TAG`", i.e. the tag-name fallback.

```sql
-- the predicate. Note the BACKREFERENCE \1 and that it is CASE-SENSITIVE.
spec->>'selector' ~ '^([A-Za-z0-9]+)\.\1$'
```

**Gotchas, all of which cost me a wrong number or would have:**

- **Do NOT use `^[A-Z0-9]+\.` to find them.** That matches *every* row, because the tag component
  is always uppercase — `SPAN.calc-eyebrow` (a real class) matches it too. It answers "is the tag
  uppercase", not "is the class a copy of the tag". My first attempt did exactly this and reported
  452 of 452.
- **Case-sensitivity is load-bearing, keep it.** The backreference makes `H3.H3` match and `H3.h3`
  not. Since the fallback copies `el.tagName` verbatim (uppercase per the HTML DOM), a
  case-*insensitive* variant would start catching genuine lowercase classes that happen to equal
  their tag.
- `split_part(sel,'.',1) = split_part(sel,'.',2)` is an equivalent predicate (the 198 lane used it
  to reproduce these figures independently) and is easier to read. It differs on a multi-class
  selector, which this producer never emits — one class token maximum, `contrastSelector` takes
  `tokens[0]`.

Run from the repo root; `psql` is only reachable through the pod:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "…"
```

⚠ **psql eats a leading backslash** in some quoting paths, and `$` inside a double-quoted bash
string needs escaping (`\$`) or bash eats the anchor and the regex silently matches more. A regex
that has lost its `$` still returns rows, so this failure is invisible. Sanity-check any count
against `SELECT DISTINCT spec->>'selector'`.

## §2 The four figures, and how to get each

```sql
-- (a) the whole population by status, with the invented subset in the same pass
SELECT status, count(*) AS n,
       count(*) FILTER (WHERE spec->>'selector' ~ '^([A-Za-z0-9]+)\.\1$') AS invented
FROM site_work_items WHERE item_type='contrast_failure' GROUP BY status ORDER BY n DESC;

-- (b) the OPEN, still-failing population — the brochure lane's "durable 185"
--     ⚠ use workItemClosedStatuses, NOT workItemTerminalStatuses. They differ.
--     ⚠ AND READ THE COLUMN NAMES: `sites` below counts sites across the WHOLE open
--       population, not across the `invented` subset in the FILTER next to it. Two
--       counts in one row and only one of them filtered. Run 2026-08-24 it returned
--       181 / 73 / **16** — while this lane has always quoted **13**, which is the
--       invented subset's site count and needs the second query below.
SELECT count(*) AS open_still_failing,
       count(*) FILTER (WHERE spec->>'selector' ~ '^([A-Za-z0-9]+)\.\1$') AS invented,
       count(DISTINCT site_id) AS sites_ALL_OPEN_not_just_invented
FROM site_work_items WHERE item_type='contrast_failure'
  AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled');

-- (b2) the invented subset's OWN site count — this is the 13
SELECT count(*) AS open_invented, count(DISTINCT site_id) AS sites_invented
FROM site_work_items WHERE item_type='contrast_failure'
  AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
  AND spec->>'selector' ~ '^([A-Za-z0-9]+)\.\1$';

-- (c) is it still PRODUCING, or is this history? A count with no date is unanswerable.
SELECT min(created_at)::date, max(created_at)::date,
       count(*) FILTER (WHERE created_at > now() - interval '7 days') AS last7d
FROM site_work_items WHERE item_type='contrast_failure'
  AND spec->>'selector' ~ '^([A-Za-z0-9]+)\.\1$';
```

**⚠ THE ARITHMETIC GOTCHA, because it bit me.** Do not add the per-status figures in your head to
get "how many are at risk". Derive it in the query:

```sql
count(*) FILTER (WHERE status NOT IN ('complete','verified','rejected','wont_fix','cancelled'))
```

I added `unresolved` **26** (the status total) instead of **15** (the invented subset) off my own
printed table and published 84 for 73. Every input was correct and on screen. → `WRONG_CALLS.md`.

**⚠ `site_work_items` is a ROLLING WINDOW** — closing a row can archive it out of the table you
queried, so a figure for "how many were ever X" cannot be taken from here. Date every count.

## §3 The two status sets — which one you want, and when

`platform/orchestration/actions/work_items_common.go`:

| set | line | members | used for |
|---|---|---|---|
| `workItemTerminalStatuses` | `:42` | complete verified rejected wont_fix cancelled **failed unresolved** | dedup / `ON CONFLICT`; interpolated into `insertWorkItem` |
| `workItemClosedStatuses` | `:85` | complete verified rejected wont_fix cancelled | retraction's "already settled" filter |

**The asymmetry is deliberate and documented at `:97` — do not "fix" it in passing.** It has a real
consequence for any rekey: `unresolved` is terminal, so it does **not** hold the dedup slot and
cannot collide; `deferred` is in neither set, so it **does** hold the slot and **can** collide with
`idx_swi_dedup`. Get this backwards and the migration throws a unique violation on the 58 rows that
matter most.

⚠ And the 42P10 trap: the `ON CONFLICT … WHERE` clause MUST imply `idx_swi_dedup`'s predicate or
*every* keyed insert fails fleet-wide with SQLSTATE 42P10. The Go list and the index are one
contract. Read the comment at `:49-54` before touching either.

## §4 Reading the producer and its siblings

```bash
grep -n "tagName" internal/adapters/browserrunner/*.go scripts/render_audit.py
```

Four hits, and only one is the bug — see NOTES for the table. The important one to NOT change is
`run_checks_action.go:1140`, which names a **component**, not a selector.

⚠ **`render_audit_action.go` is over 60 KB.** A `090` diagnosis run on a symbol in a file that big
returns bundles and **no verdict**, and that looks exactly like a run still in progress
(LANDMINES). Do not spend a run on this file expecting a verdict.

## §5 Verifying the claim that a bare selector is site-wide

Do **not** infer this from the code — read the agent's actual prompt, which is the contract:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -c "
SELECT substring(v->'config'->>'prompt_template' from 1 for 1400)
FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') AS e(k,v)
WHERE a.type='css-patch-agent' AND a.is_active AND COALESCE(a.is_snapshot,false)=false
  AND a.deleted_at IS NULL AND k='plan_css_fix';"
```

It says, verbatim: *"The platform APPENDS your rules to the END of the stylesheet above"* and
*"Repeat the offending selector exactly as it appears above (or more specifically) so your override
wins."* One stylesheet per site (198 proved 22/22 `css_themes` rows unshared), so **an appended
bare `p { … }` applies to every page that loads it.** That is the whole basis of the
"naive fix is a regression" finding, and it is measured at the prompt rather than assumed.

⚠ **Read the live row, not the seed.** `SEED_*.sql` records what the agent *was*; config drifts.
And count **step** overlays with `jsonb_each(default_config->'workflow'->'steps')` — a root-level
`ai_service.model` census shows ONE agent and badly understates it (the `bugfix_257` trap).

## §6 Where the fix ships — TWO images, and the chassis is the wrong half

```bash
for d in cmd/*/; do grep -rq "adapters/browserrunner" "$d" && echo "IMPORTS: $d"; done
```

Result: **`cmd/browser-runner-adapter/` only.**

| change | image | build |
|---|---|---|
| `internal/adapters/browserrunner/render_audit_action.go` | browser-runner | `make build-browser-runner-adapter` |
| `platform/orchestration/actions/write_render_audit_findings_action.go` | agent-chassis | `make build-agent-chassis` |

⚠ **`render-audit-adapter` runs the browser-runner IMAGE** (makefile:107) — so its overlay `newTag`
must move with the browser-runner's, in the same commit, or the two drift. All three overlays sat at
`v1.0.1332` and `IMAGE_TAG` at `v1.0.1333` on 2026-08-24.

⚠ **"the chassis rolled" is not evidence this fix is live.** Half of it is not in that image. Prove
per SERVICE:

```bash
kubectl -n ai-persona-system logs -l app=browser-runner-adapter --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <my-commit> <the stamp>
```

An empty grep means "scrolled out of range", **not** "unstamped" — it is a startup line. Fall back
to the binary probe, and **always run a control in the same breath** (a sha that must be absent and
one that must be present). Never `strings` (absent from the debian-slim images) and never a
discovery grep for "some 40-hex string" (it matches Go's internal digit table and returns the same
wrong answer on every service).

## §7 Council gate — working again, contrary to a sibling lane's handoff

`bugfix_131_contrast_ratio_check/HANDOFF_2026-08-22_continue_here.md` §0 says the gate is DOWN
(`claude-sonnet-5` capped to 2026-09-01, all 17 seats on that model). **Stale as of 2026-08-24.**
Check before believing either way:

```sql
SELECT current_step, status, created_at,
       collected_data->'input_data'->>'fix_correlation_id' AS corr
FROM orchestration_states
WHERE collected_data->'input_data' ? 'fix_correlation_id'
  AND created_at > now() - interval '18 hours'
ORDER BY created_at DESC LIMIT 6;
```

2026-08-24 gave four `complete_approved` and two `complete_revise`, newest 13:01Z; 47 COMPLETED
runs over three days. ⚠ Do **not** read fleet throughput as evidence the gate works — the haiku
dispatch lanes are unaffected by a sonnet cap and keep the aggregate healthy. Query the gate's own
rows, as above.

Budget **~30 minutes**, not ~2: the council takes 2–5 min but the dispatch queues behind the fleet.
Find your run by **payload**, not by the printed id. A missing row is almost always latency — do
not retry on that evidence.

## §8 Migration numbering

```bash
ls docs/agent_docs/sql_for_agents/ | sed -n 's/^\([0-9]\{3\}\)_.*/\1/p' | sort -u | tail -5
```

Highest on 2026-08-24 is **585**, so **586** is next free. ⚠ Re-run this immediately before writing
the file — several lanes are filing migrations daily and the number goes stale in hours.

⚠ `run-migrations.sh --apply` takes **every** pending file, including other lanes' backlog. Scope
the directory, and dry-run once per session and again after every roll. An ordering-critical file
must be named `_HOLD.sql` — a banner does not hold it back.

⚠ A verify block made of `SELECT`s **cannot stop the `COMMIT`** (`ON_ERROR_STOP` ignores a
non-empty result). Use `DO` / `RAISE`, and **induce the failure once** to prove the guard fires.

## §9 The ownership check that gave the wrong answer

`scripts/who-owns.py 352` → **"OWNED or recently active … Do not start a competing fix."** The
evidence it cites is the commit that **FILED** the bug. Read the commit it names and ask what that
commit did; then `ListAgents` + one `SendMessage` to the lane. That is the only source that is not
lagging, and it took seconds. → `WRONG_CALLS.md`.

## §10 ⚠ AFTER 587 APPLIES, EVERY CENSUS IN THIS LANE RETURNS ZERO — BY DESIGN, NOT BY DRIFT

Raised by the `bugs_open/198` lane, 2026-08-24, and it is the sharpest trap this lane leaves behind.

> ## ⚠ STATE AS OF **2026-08-24 19:11:22 UTC — 587 IS APPLIED**. You are on the FAR side.
>
> **The whole of the rest of this section has switched tense.** It was written in advance,
> in the future tense, for the reader who would apply the file. That reader has been: the
> migration was applied by hand at **2026-08-24 19:11:22 UTC** and reported `UPDATE 73`.
>
> **So the open-population census in §2 now returns 0, and that is the SUCCESS condition,
> not a refutation of the earlier numbers.** Use the recovery query below — it is the one
> that keeps returning 73 for ever. Post-application checks, all fresh:
> `open_invented = 0`, `withdrawn = 73`, `withdrawn_without_prior_status = 0`,
> `falsely_completed = 0`, and the recovery query returns **deferred 58 + unresolved 15 = 73
> across 13 sites**, matching the pre-application census exactly.
>
> **The superseded banner, kept because the trap it describes is still real for anyone
> reading an older copy of this lane's docs:**
>
> > ~~**STATE AS OF 2026-08-24: 587 IS COMMITTED AND *NOT APPLIED*. The 73 are LIVE.**~~
> > ~~[MEASURED 2026-08-24, after both lanes independently ran it] `withdrawn_by_587 = 0`,
> > `carrying_prior_status = 0`, `contrast_failure` total **452**, `open_invented_now` **73**.~~
> >
> > ⚠ **And one figure in that superseded banner was already wrong when written: the total was
> > not 452 at the time it was labelled.** 452 was the total as of *before* 15:31:50 UTC; by
> > 16:55, when the banner was dated, a pre-roll audit had filed 47 more and the true total was
> > 499. The measurement was sound and the LABEL carried the writing time, not the measuring
> > time. Arithmetic that pins it: today's total 509 − the 10 post-roll rows − those 47 = 452.
> > → `WRONG_CALLS.md` 2026-08-24. **Date a figure when you MEASURE it, not when you write it up.**
>
> **A committed migration is indistinguishable from an applied one from inside the repo**, and that
> is how the 198 lane came to write *"587 withdraws the 73"* in the present tense into two files
> without checking the cluster. `_HOLD` is the only tell, and it lives **in the filename** — not in
> the SQL, not in `git log`, not in any query you would think to run. See LANDMINES.

The estate's dated-count rule exists because a census goes stale **by ADDITION** and keeps reading
as current. This is the **SUBTRACTION** case, and it is worse, because the number does not merely
drift — it goes to **zero**, which reads as *"this never happened"* or *"the earlier census was
wrong"* rather than *"we fixed it"*.

**Five documents of mine, plus two of the 198 lane's and one of the brochure lane's, quote `73`**
(and `181`, and `108`). The moment `587_retire_invented_contrast_selectors_HOLD.sql` is applied,
§2's predicate returns **0** for the open population, permanently. Neither figure was ever wrong.

**So: any figure in this lane is `<n> as of 2026-08-24, BEFORE migration 587`.** If you are checking
one of them, the question is *which side of 587 are you on*, and that is answerable:

```sql
-- WHICH SIDE OF 587 AM I ON? Run this FIRST, before any census below.
SELECT count(*) FILTER (WHERE result->>'cancelled_by' = 'migration_587') AS withdrawn_by_587,
       max((result->>'cancelled_at')::timestamptz)                       AS applied_at
  FROM site_work_items WHERE item_type = 'contrast_failure';
-- withdrawn_by_587 = 0  → pre-migration; §2's numbers should still reproduce.
-- withdrawn_by_587 > 0  → post-migration; §2 returning 0 is the SUCCESS condition.
```

**To recover the population ONCE 587 HAS BEEN APPLIED** (it has not been, as of 2026-08-24) — the
rows are withdrawn rather than deleted, and each carries the status it held:

```sql
SELECT result->>'pre_352_status' AS status_before_587, count(*), count(DISTINCT site_id) AS sites
  FROM site_work_items
 WHERE item_type = 'contrast_failure' AND result->>'cancelled_by' = 'migration_587'
 GROUP BY 1 ORDER BY 2 DESC;
-- EXPECT, if the census held: deferred 58, unresolved 15 — 73 across 13 sites.
```

That query is the one to quote in future, because it **keeps returning 73 for ever** while the
open-population census correctly falls to zero. ⚠ Note the two answer different questions and only
one of them is stable — do not substitute one for the other, which is the whole error this section
exists to prevent.

**The `complete` population (108) is NOT touched by 587** and stays queryable by §2's predicate
without the status filter. It is the larger and more damaging number — already-recorded false
repairs, against 73 that were only ever *at risk* — and it never moves, because those rows were
closed long before this lane existed.
