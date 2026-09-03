# RUNBOOK — improvement loop

Every command here had a gotcha attached when I first ran it. The gotcha is the point.

`PSQL` below means:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## Is the loop running at all?

```sql
SELECT name, enabled, interval_seconds, last_triggered_at, last_completed_at
FROM scheduled_tasks WHERE name = 'improvement-sweep';
```

⚠ **Do not answer this from any document.** Several standing docs and one auto-memory
entry still carry the 2026-07-29 owner ruling that the sweep is off deliberately.
Migration `389` re-enabled it. The row is the fact; the ruling is history.

## What the loop did, and whether it did anything

Orchestration rows are purged in roughly a day, so **this window is all the evidence
there is** — a question you ask on Monday about Friday has no answer.

```sql
SELECT date_trunc('day', created_at)::date d, status, current_step, count(*)
FROM orchestration_states
WHERE owner_agent_type = 'improvement-loop' AND created_at > now() - interval '2 days'
GROUP BY 1,2,3 ORDER BY 1 DESC, 4 DESC;
```

⚠ **The column is `owner_agent_type`, not `agent_type`** — there is no `agent_type` on
this table and the query errors rather than returning zero, which is the good case.

⚠ **`execution_path` is EMPTY on these rows.** It looks like the natural way to ask
"which steps ran", and it will silently tell you "none of them" for every run. Read
`collected_data`'s keys instead:
`SELECT jsonb_object_keys(collected_data) FROM orchestration_states WHERE orchestration_id = '…';`

## Did the audit half run, or was the site skipped?

```sql
SELECT collected_data->'audit_state'->>'audit_due'      AS audit_due,
       collected_data->'audit_state'->>'not_converging' AS not_converging,
       count(*)
FROM orchestration_states
WHERE owner_agent_type = 'improvement-loop' AND created_at > now() - interval '2 days'
GROUP BY 1,2;
```

⚠ **`complete_clean` does NOT mean "site is clean".** It is also the terminus for a
site whose fingerprint has not changed (audit skipped, migration 291) and for a site
whose entire finding pile was held back as unroutable. Three very different states,
one step name. `audit_state` is what separates them — that separability was
`bugs_open/171`'s explicit requirement, so use it.

## How much work is held back, and of what kind

Per run:

```sql
SELECT collected_data->'triage_result'->>'promoted'       AS promoted,
       collected_data->'triage_result'->>'not_promotable' AS held,
       collected_data->'triage_result'->'not_promotable_by_type'
FROM orchestration_states
WHERE owner_agent_type = 'improvement-loop' ORDER BY created_at DESC LIMIT 5;
```

⚠ **Do not sum `not_promotable` across runs.** It is a per-run count of the site's
standing pile, so the same rows are counted once per visit. Summing gave me 3,866 for a
backlog whose true size is 1,385. For the standing figure, count the rows:

```sql
SELECT item_type, count(*), count(DISTINCT site_id) sites, min(created_at)::date oldest
FROM site_work_items
WHERE status = 'detected' AND (handler_agent IS NULL OR handler_agent = '')
GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **`handler_agent IS NULL` alone is not the predicate.** Migration `217` normalised
NULL to `''` and made the column NOT NULL; every live reader spells it
`(handler_agent IS NULL OR handler_agent = '')`. Ask only about NULL and you get zero,
which reads exactly like a clean estate.

⚠ **`site_work_items` is a rolling window** — `work-item-archiver` runs daily. Any
claim about a *population* over time must `UNION site_work_items_archive`. The query
above is deliberately about the *standing* pile, which is live-table-only by definition.

## Is a finding TRUE? Probe the page, never the row

```bash
for u in <recorded page_url> <invented-url-on-the-same-domain>; do
  code=$(curl -s -o /tmp/p.html -w '%{http_code}' -m 20 "$u")
  echo "$u | http=$code | $(grep -o -i -m1 '<title>[^<]*</title>' /tmp/p.html || echo NO-TITLE)"\
       "| footer=$(grep -c -i '<footer' /tmp/p.html) | bytes=$(wc -c </tmp/p.html)"
done
```

⚠ **The invented URL is not optional.** A parked domain answers 200 to every path —
boxingonline.com returns a 114-byte redirect stub for `/`, `/about.html` and anything
else you type. Without the control you cannot tell "our page is broken" from "this
domain is not ours to serve".

⚠ **`spec.missing` on an existing row is of unknown age.** `insertWorkItem` writes with
`dropOnConflict`, so a re-run of the check drops the fresh row and leaves the original
spec in place; and `head_essentials_missing` only retracts when *all* essentials are
present. A row whose skip link is still missing keeps its first-ever missing-list for
ever, however much of it has since been repaired. **Read the page, then the row.**

## Fire one loop run by hand

The sweep picks its own site (`ORDER BY s.updated_at ASC NULLS FIRST LIMIT 1`), so you
cannot choose one by waiting. Dispatch directly at the agent instead — and read
`scripts/kafka-publish-lib.sh` first rather than hand-rolling a `kcat -P`, which exits 0
having sent nothing.

---

## Before the skip-link re-render wave — the two-stage gate

The Go change is inert until the fleet image rolls, and then inert per page until that
page re-renders. **Stage 2 must not start until stage 1 is proven at the artefact**, and
this gate exists because the council's `debug_historian` seat objected (medium, corr
`3c71ec77`) that the original plan named no such gate. A fan-out fired against a pod that
is still running the old binary re-renders every page on the estate to exactly the same
bytes: expensive, green, and indistinguishable from success.

**Stage 1 — is the code actually running?** Ask the service, not git and not the tag:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 5cfd41bc0 <the sha that prints>   # exit 0 = it shipped
```

⚠ That is a STARTUP line and scrolls off a busy service, so an empty result means "not in
range", not "not stamped". Fall back to the binary probe, **with both controls in the same
breath** — a sha that must be present and one that must be absent:

```bash
kubectl -n ai-persona-system exec <pod> -- grep -aq "<expected-sha>" /proc/1/exe   # must hit
kubectl -n ai-persona-system exec <pod> -- grep -aq "<a-sha-not-in-the-build>" /proc/1/exe  # must miss
```

**Stage 1b — does the running code emit the link?** The cheapest positive proof is one
page, not a census: re-render a single page, fetch it, and check for the pair.

```bash
curl -s https://<domain>/<page>.html | grep -c 'class="skip-link"'    # expect 1
curl -s https://<domain>/<page>.html | grep -c 'id="content"'         # expect 1, NOT 2
curl -s https://<domain>/<page>.html | grep -c 'data-skip-link'       # expect 1 — the CSS
```

⚠ **All three, not just the first.** The link without the CSS is the failure that a whole
test suite missed until it was mutated: the page still contains a skip link, and it is
VISIBLE at the top of a client site.

**Stage 2 — the fan-out**, only once stage 1b has passed on a real page.

## What "it worked" looks like afterwards, and what it does not

The findings retract themselves: `HeadEssentialsMissingCheck` re-probes the live page each
run and emits a `ResolvedFinding` on `len(missing) == 0`, which `resolveWorkItems` closes
by `site_id + item_type + item_key`. It does **not** filter on `handler_agent`, so these
flag-only rows are retractable — verified in code, not assumed.

⚠ **Do not read a flat count as the fix failing.** A row clears only after its page has
re-rendered AND the structural check has next run over that site (its rotation is hours,
not minutes). Expect the 867 to drain over days, not on the roll.

⚠ **Expect a residual of exactly 10 rows on 5 sites**, named in NOTES §(gg): pages with no
`page_components` rows, which this path does not build. `[MEASURED 2026-09-02]` 968 of 978
rows carry `spec.assembled = true`, so the covered fraction is 99.0%. **If the count
plateaus near 10, that is the expected floor, not a stall.**

---

## Migration ledger — is a file you APPLIED actually RECORDED?

```sql
SELECT filename, applied_at, applied_by FROM schema_migrations WHERE filename LIKE '722%';
```

⚠ **Empty means unrecorded, not unapplied — and the artefact tells you which.** 722 was
applied by hand on 09-03 and this returned nothing all day; `pg_trigger` said the trigger was
live. A hand-apply and a `--record-only` are ONE motion:

```bash
./scripts/migration/run-migrations.sh --record-only <file.sql> --note '<what you verified live>'
```

⚠ **The full dry run (`run-migrations.sh` with no flags) probes every pending file and took
>300s on 2026-09-03** (14 pending, some from July). `--no-probe` lists in seconds. To probe
ONE file, copy it alone into a scratch dir and point the runner at it:
`MIGRATIONS_DIR=<dir> ./scripts/migration/run-migrations.sh` — the probe executes it verbatim
in a doomed transaction and reports `ok … ran to its own COMMIT` or the error.

## Retractions — the vocabulary

A `head_essentials_missing` row the check re-probes clean goes to **`status='complete'`**,
not `retracted`. `[MEASURED 2026-09-03]` there is no `retracted` status on this type at all;
a query for one returns 0 and reads exactly like "the drain stopped".

```sql
SELECT count(*) FILTER (WHERE status='detected') AS open,
       count(*) FILTER (WHERE status='complete' AND completed_at > '2026-09-02 20:56:43Z') AS retracted_post_roll
  FROM site_work_items WHERE item_type='head_essentials_missing';
```

## Growth posture — who is held, for how long, and is the born-held rule still on

```bash
scripts/audit-growth-posture-hold.sh              # the daily report, by hand; no doc_notes row
scripts/audit-growth-posture-hold.sh --days 14    # a different threshold for this run
scripts/audit-growth-posture-hold.sh --write      # also record the row, as the CronJob does
scripts/audit-growth-posture-hold.sh --self-test  # fixtures only, no cluster
```

Exit 0 clean, 1 findings, 2 refused to look. It is THE SAME `check.py` the CronJob runs.

⚠ **"age=unknown" on a hand-held site is expected, not broken.** The row has no
`growth_posture_set_at`; the check bounds the age below by the first day it saw the hold and
says so. The fix is the lane stamping the row (three-line recipe in register WDS-020), not a
smarter guess.

⚠ **The 722 demand test is still unrun** — no site has been created through the trigger.
When one is, check it directly, and check the RECORD too now that 752 is in:

```sql
SELECT domain, created_at,
       settings->'maintenance_profile'->>'growth_posture'        AS posture,   -- must read hold
       settings->'maintenance_profile'->>'growth_posture_set_by' AS set_by,    -- trg_sites_born_holding_growth
       settings->'maintenance_profile'->>'growth_posture_set_at' AS set_at     -- ≈ created_at
  FROM sites ORDER BY created_at DESC LIMIT 3;
```

⚠ **copyonline.co.uk will read open in that query and it is not the trigger failing** — it
was created ~100s before 722 applied (NOTES §(qq)).
