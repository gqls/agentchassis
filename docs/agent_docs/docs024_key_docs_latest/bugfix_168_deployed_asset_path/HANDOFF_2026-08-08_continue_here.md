# HANDOFF — 2026-08-08 — cold start. The sweep lane is DONE and self-running. Two known items remain.

> # SUPERSEDED FOR STATE — read `HANDOFF_2026-08-08b_continue_here.md` instead.
> **2026-08-08, later the same day.** The sweep gained a fifth covered type (`voice_tells`,
> `ef80216be`), the dedup figures in §4.1 moved again (now **55 pairs / 184 rows**), and §4.3's
> candidate list did not survive its own CLOSER check — `content_rewrite` is drained by a real fix
> pipeline and is NOT a retraction candidate. The reasoning and the traps below are still good.


**Read this file only.** It supersedes `HANDOFF_2026-08-04_continue_here.md` for state; that file
stays for its reasoning and its traps. Plain-prose account: `SUMMARY_2026-08-06b_…`. Missteps and
working: `NOTES_deployed_asset_path.md` (newest at the bottom).

**Nothing is half-applied. Nothing needs a decision. The lane can be dropped here with no cleanup.**

---

## 1. State — verified 2026-08-08, chassis `v1.0.1262`, both replicas

| thing | state |
|---|---|
| `bugs_closed/168` (asset path) | CLOSED, live since `v1.0.1229` |
| RFC_010 retraction seam + first adopter | LIVE, working |
| **Sweep starvation (§0 of the old handoff)** | **FIXED, LIVE, PROVEN UNATTENDED.** See §2 |
| Stopgap: `max_items` 500→1500 | LIVE. Migration `321`, commit `b14609e05` |
| Durable fix: selection filters to judgeable types | LIVE. `0e4e79124`, council `f64da546` APPROVED r1 |
| `scheduled_tasks.review-queue-revalidate-daily` | `enabled=t`, 86400s, firing on its own |
| §0b's "individually skipped" row | RESOLVED, suspected cause REFUTED |
| **Decision 2's dedup half** | **OPEN, blocked. §4.1 — the only real work left here** |
| **Armed-but-inert cap in a sibling check** | **OPEN as a tripwire. §4.2** |
| RFC_010 Q1 (two-strike) | Owner-ruled accept-as-is, tracked, not work |

## 2. THE PROOF, because it can no longer be re-derived from the database

**`orchestration_states` retention is ~24h.** The run that proved this closed 20 previously
unreachable rows has **already aged out** — there is exactly **1** reval run in the table today. The
figures below are the only surviving record. Do not go looking for them in the DB; they are gone.

**Unattended, measured from the closures rather than inferred:**

```sql
SELECT completed_at::date, count(*) FROM site_work_items
WHERE resolution_path='auto:revalidated' GROUP BY 1 ORDER BY 1 DESC;
--  2026-08-08 | 3    <- scheduled, nobody watching
--  2026-08-07 | 1    <- scheduled, nobody watching
--  2026-08-06 | 21   (20 = the one hand-dispatched drain of the unreachable backlog)
--  2026-08-04 | 33   (pre-fix history)
```

Today's scheduled run: `scanned 151 · capped_at 1500 · cap_binding false · resolved 3 ·
still_holds 37 · unknown 111 · uncovered_backlog 625`. Live at the same moment: **148 judgeable,
625 uncovered, 773 parked** — so `151 − 3 = 148`, `uncovered_backlog` matches the live count to the
row, and the sweep looked at the right 151 of 773.

⚠ **`resolved 3` is NOT degradation from 20.** The 20 was a one-off flush of a fortnight's
unreachable backlog; 1 and 3 are steady state. **Watch `cap_binding`, not `resolved`** — if
`cap_binding` is ever `true`, judgeable work is being left behind and the starvation is back.

**How to re-verify at any time** (the strings are ASCII-only on purpose; `N=${N:-0}` is not
optional because `grep -c` exits 1 and prints nothing on zero):

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | sed 's|pod/||'); do
  A=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'judgeable rows were left unexamined'" 2>/dev/null|tail -1); A=${A:-0}
  B=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'so this sweep cannot judge it'" 2>/dev/null|tail -1); B=${B:-0}
  P=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'auto:revalidated'" 2>/dev/null|tail -1); P=${P:-0}
  echo "$POD capwarn=$A refusal=$B positive_control=$P"    # want 1 / 1 / non-zero
done
```
Confirmed 1/1/2 on `v1.0.1257` (when first proven) and again on `v1.0.1262`. **None of these builds
were mine** — the change rides along at HEAD, which is why it is worth re-greping after any roll.

## 3. What the fix actually is, in case you need to reason about it

`loadParkedReviewItems` used to select the oldest N parked rows of **any** type. Types with no
revalidator return `unknown`, which is deliberately **non-terminal**, so they stayed parked, stayed
oldest, and were re-selected every run — only ~104 of 500 head slots ever turned over. **64
judgeable rows sat permanently beyond the cap: never reached, not reached slowly.** They were also
the *newest* covered rows, which inverts the selection's own oldest-first rationale.

Now the selection filters to the types `reviewRevalidators` covers, **derived from that map**, so
registering a revalidator widens the selection in the same edit and the two cannot drift. The
coverage gap is one `GROUP BY` over the whole parked set (the old shape reported it as *smallest*
exactly when the backlog was worst). `cap_binding` is logged at WARN when a pass fills its cap.

**If you add a revalidator, you no longer have to think about the cap** — registering the type is
what makes the selection load it. A type *not* in the map is never loaded, and is counted in
`uncovered_types` instead.

## 4. What is left

### 4.1 Decision 2's dedup half — the only substantive work, and it is blocked

Owner ruled `unresolved` is OPEN, so it must leave `idx_swi_dedup`'s exclusion list. Two things make
this a project rather than an index swap, and **both are unchanged**:

1. **Duplicate rows block the unique index.** Re-measured **2026-08-08 against the PROPOSED
   predicate** (today's index minus `unresolved`): **53 colliding `(site_id, item_key)` pairs across
   180 rows** — up from 48 / 135 on 2026-08-03. Largest contributors: `undeployed_asset` 48 rows,
   `improve_tool` 30, `needs_internal_links` 29. `CREATE UNIQUE INDEX` fails against this
   population; the cleanup is a prerequisite needing the "which copy do I keep, and does discarding
   the rest lose a true finding?" judgement.
2. **The ordering is asymmetric and one direction breaks the fleet.** Unchanged — see §4.3 of
   `HANDOFF_2026-08-04_continue_here.md` for the `42P10` / `23505` argument in full. Go-first breaks
   every keyed insert fleet-wide; index-first stops `unresolved` rows being conflict targets.

⚠ **READ THE INDEX, DO NOT RECONSTRUCT IT.** Measuring this on 2026-08-06 I wrote the exclusion list
from memory of the phrase "terminal statuses" and got **75 pairs / 227 rows** — nearly recording a
56% growth that was an artefact of my own query. The live predicate excludes **three more** statuses
(`wont_fix`, `failed`, `unresolved`):

```sql
SELECT pg_get_indexdef(oid) FROM pg_class WHERE relname='idx_swi_dedup';   -- the only authority
-- correct measurement, PROPOSED predicate = live index minus 'unresolved':
SELECT count(*) AS pairs, sum(n) AS rows_involved FROM (
  SELECT site_id, item_key, count(*) AS n FROM site_work_items
  WHERE item_key IS NOT NULL
    AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','cancelled'])
  GROUP BY 1,2 HAVING count(*) > 1) q;
```

### 4.2 The armed-but-inert cap — a tripwire, not a task

`discovery_checks/check_image_source_unsatisfiable.go:167` is still `return result, nil` inside its
per-pass cap, and still populates `Resolved` **0** times (verified 2026-08-08). **It is correct
today** — a noise bound *should* stop early while a check can only file. **The commit that adopts
the retraction seam there is the commit that must change it to `break`**, or retraction goes silently
inert on exactly the badly-shaped sites carrying the most stale items. In `LANDMINES.md` with the
re-run census command.

### 4.3 The honest big one: 625 parked rows nothing can re-check

Not this lane's next move, deliberately. Closing it means teaching the sweep more item types, one at
a time, each independently reviewable. **Before adopting any of them, run the CLOSER check as well
as the producer check** — this lane shipped a duplicate closer once and 14 council seats accepted the
false premise:

```sql
SELECT status, count(*), left(result::text,120) FROM site_work_items
WHERE item_type='<your type>' AND status IN ('complete','verified') GROUP BY 1,3 ORDER BY 2 DESC;
```
A `result.revalidation` block means `revalidate_review_queue` already owns that type — extend it
rather than writing a second closer. Candidates by size: `cta_names_unknown_destination` (~123, but
**owned by `bugs_open/023`** — do not touch), `content_rewrite` (~34), `voice_tells` (~25),
`needs_sprite_css` (~10).

### 4.4 Or take the next unowned bug from `bugs_open/`

The original standing task. `scripts/who-owns.py <n>` **plus** a grep of live `.jsonl` transcripts —
the script reads commits and is blind to a session mid-fix.

## 5. Traps specific to this lane

- **`scheduled_tasks.input_data` is INERT for this action.** `max_items` *and* `item_type` come from
  the **step config** (`params.StepConfig.Config`); the `sweep` step has no `input_mapping`. So
  "several scheduled rows, one per type" **cannot work** — they would all read the same config. The
  old handoff recommended exactly that, two paragraphs after documenting the trap. `WRONG_CALLS.md`
  2026-08-06.
- **`last_triggered_at` is written at publish time and is NOT proof an agent ran.** Take the chain in
  `HANDOFF_2026-08-04` §0c, or measure the closures as §2 does.
- **A row missing a stamp its siblings have may never have been LOADED.** §0b suspected an
  `item_key` prefix drift; refuted — the row was simply beyond the cap, and its "judged siblings"
  were two weeks older, so the comparison was an age comparison in a predicate's clothing.
- **`uncovered_types` changed meaning on 2026-08-06** — from "unjudgeable rows inside the cap" to
  "every parked unjudgeable row". Larger, and the true figure. Comparisons across that date compare
  two different measurements.
- **Uncovered rows are no longer stamped at all**, so absence of `result.revalidation` no longer
  distinguishes *never scanned* from *no revalidator exists*. Raised by the council; the per-type
  answer is in `uncovered_types`. The review_queue_drain seed's documented verification queries are
  blunted for uncovered types (that lane has been told).
- **`/tmp` is a near-full 16G tmpfs**; the Go linker writes there. Use
  `TMPDIR=/home/ant/.cache/buildtmp go build ./...`.
- **Test against `git archive HEAD`** when the shared tree is broken by another session — and delete
  the extraction afterwards.
- **Use `git commit -F <file>`**; backticks in `-m` execute.

## 6. Correlations

| what | id |
|---|---|
| selection filter council (APPROVED r1, 11 seats, 4 advisory) | `f64da546-e1a4-42e4-98c0-d94cf42af71c` |
| the hand-dispatched drain that closed 20 (payload since aged out) | `267fe850-5c22-43fe-b1d1-266e261a3a40` |
| today's unattended run | `1ac359c4-81d1-4826-89bb-7f060bcdd612` |
| RFC_010 Decision 1 council | `846f4f3d-8958-4e4c-be81-d5f02e20852d` |
| `168` council / diagnosis (REFUTED — read why in NOTES) | `abd9b119-…` / `ae9404bd-…` |

## 7. Commits, in order

`b14609e05` mig 321 (cap stopgap) · `0e4e79124` the selection filter + tests · `246763083` docs,
WRONG_CALLS, LANDMINES, consumer notices · `c21b7e216` the live proof · `45c37b4f8` landmine
disarmed-for-its-instance · `9e4caa2b3` SUMMARY 2026-08-06b + my "nothing open" correction.
