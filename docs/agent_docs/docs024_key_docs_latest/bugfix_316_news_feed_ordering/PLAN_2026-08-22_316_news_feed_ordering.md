# PLAN — bugs_open/316, ordering fairness in the news-feed trigger

Design, phasing, decisions **and their reasons**. Corrections live here, marked as corrections.

---

## Status: 2026-08-22, evidence complete, plan being drafted

Everything in this section is settled and measured; see `NOTES_316_news_feed_ordering.md` for the
queries and `RUNBOOK_316_news_feed_ordering.md` for how to re-run them.

### The two defects are separable and only one is mine to fix

1. **Unfairness — a defect.** `content-feed-trigger.find_news_sites` ends `ORDER BY s.domain LIMIT 5`.
   Fix by ordering on the schedule. Config-only, live on apply, no image roll.
2. **Capacity — an owner spend decision, NOT a defect fix.** 42 fetches/day demanded against 20 supplied
   (2.10x); removing the cap entirely still leaves 36 vs 42 (1.17x). [MEASURED 2026-08-22, reproduces the
   filed figure exactly.] **Out of scope. Present the arithmetic; change nothing.**

### Decisions taken, with reasons

- **Order on the due-time, not on `random()`.** `random()` is what the sibling `model-directory-trigger`
  uses and the bug file calls it "the cheap version of the same idea". It is cheaper and it is worse:
  it makes starvation unbiased rather than absent, and a site can still lose several draws running. The
  platform's own Go layer already orders this exact kind of work by `next_fetch_at ASC NULLS FIRST` in
  two places, so due-ordering is the **existing convention**, applied to the one layer that skipped it.
- **The fix must answer the `NOT EXISTS(active sources)` arm explicitly.** The bug file's fix candidate 1
  (`ORDER BY min_next_fetch_at NULLS FIRST`) taken literally creates a permanent head-of-queue squatter
  — see NOTES. This is a **correction to the originating bug file's proposed remedy**, recorded here
  rather than silently worked around.
- **Prefer a framework-level answer over the one-line fix, but size it against the evidence.** The
  dangerous shape has exactly **one** live member today. The argument for building a detector anyway is
  `query_row_cap.go`'s own, in its header: *"one function, and the whole class becomes visible at once,
  including caps nobody has written yet."* The argument against is over-engineering for n=1. This is the
  central open design question the plan must settle, and it must be settled with reasons, not taste.
- **Reuse `cmd/config-key-audit`, do not build a new service shape.** It already hosts ~10 live-config
  audits behind flags, already has a direct-Postgres route for CronJobs (`fleetdb.go`), and already
  traverses with `validation.WalkSteps`. A new check that is a new binary would be new machinery for a
  question the existing binary is built to ask.
- **Sequencing is a correctness property, not housekeeping.** Any detector must be shown FIRING on the
  motivating case before the migration lands. Once the migration applies, the live row can no longer
  supply the positive control — which is why the pre-fix query text is already captured verbatim to
  `PREFIX_find_news_sites_query_2026-08-22.sql`.

### Open at the time of writing
- The exact corrected SQL and its NULL semantics.
- Whether to build the detector, and at what size.
- Council submission.

---

## The plan, 2026-08-22 (written after the evidence; fable was unavailable — see the note at the end)

### Part 1 — the defect fix: migration `552`, config-only, live on apply

Replace `find_news_sites`'s query. The `ORDER BY` cannot simply be swapped, because `next_fetch_at` lives
inside the `EXISTS` subquery and is not in scope — the source aggregate has to be lifted into a derived
column:

```sql
SELECT site_id, domain FROM (
  SELECT DISTINCT s.id::text AS site_id, s.domain,
         (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz))
            FROM content_sources cs
           WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at
  FROM sites s
  JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true
    AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true
  WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed')
    AND (NOT EXISTS (SELECT 1 FROM content_sources cs
                      WHERE cs.site_id = s.id AND cs.is_active = true)
         OR EXISTS (SELECT 1 FROM content_sources cs
                     WHERE cs.site_id = s.id AND cs.is_active = true
                       AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))
) q
ORDER BY due_at ASC NULLS LAST, domain ASC
LIMIT 5
```

**The eligibility predicate is byte-for-byte unchanged.** Only the ordering changes, plus the derived
column it needs. That is deliberate: it keeps the blast radius to "who goes first", and it means the set
of sites the trigger considers is provably identical before and after.

Three decisions inside that ORDER BY, each of which could have gone another way:

1. **`COALESCE(cs.next_fetch_at, '-infinity')` inside the `min()`.** A source with `next_fetch_at IS NULL`
   has **never been fetched** and is therefore maximally overdue — but SQL `min()` skips NULLs, so a bare
   `min(cs.next_fetch_at)` would hide that source behind its siblings' timestamps and rank the site as if
   its never-fetched source did not exist. The sentinel makes "never fetched" sort first, which is what it
   deserves.

2. **`NULLS LAST` for `due_at` — i.e. a site with NO active sources sorts LAST.** This is the decision
   the bug file got wrong, and it needed the *orchestrator* to settle rather than the trigger.
   `content-feed-orchestrator` carries `check_has_sources` -> **`seed_content_sources`**, so arm A of the
   eligibility predicate (`NOT EXISTS (active sources)`) is the **provisioning** path: it exists so a
   newly-classified news site gets its sources seeded. That cuts both ways —
   - `NULLS FIRST` (what the bug file proposes) provisions promptly, but if `seed_content_sources` ever
     fails or yields no *active* source, the site stays in arm A with a NULL key and wins **every run for
     ever**, burning a slot on a failing seed and starving all eight other sites. Silent and unbounded.
   - `NULLS LAST` cannot squat. Its cost is that under sustained oversubscription a newly-classified site
     waits for provisioning.
   **Chosen: `NULLS LAST`**, because the failure it risks is *bounded and obvious* (a new site visibly has
   no news) while the one it avoids is *unbounded and silent*, and because — measured — the state has
   **zero live instances**: no news-feed site with a deployed page lacks an active source, including none
   stuck with only inactive ones. ⚠ **This is a genuine behaviour change for that case**: today such a
   site sorts alphabetically among everyone, so an early-alphabet unprovisioned site would currently be
   served first and after this will be served last.
   **The real defect underneath is that provisioning and refresh share one capped queue**, and a priority
   tweak papers over it. Recorded as a follow-up, not smuggled into this fix.

3. **`domain ASC` as the final tie-break**, purely for determinism. Among arm-B sites exact key ties are
   effectively measure-zero (timestamps), so this is not a reintroduction of the alphabet; among arm-A
   sites it *is* alphabetical, which is acknowledged rather than hidden and is bounded by the same
   zero-instance measurement.

**Why not `ORDER BY random()`** (what the sibling `model-directory-trigger` uses, and what the bug file
calls "the cheap version of the same idea"): it makes starvation *unbiased* rather than *absent*. A site
can still lose several draws running, and nothing about the result is reproducible when you come to check
it. Due-ordering is also not a new convention here — the platform's own Go layer already orders this exact
work by `next_fetch_at ASC NULLS FIRST` at `dispatch_feed_sources_action.go:101` and
`feed_actions.go:1016`. Config is the one layer that skipped it.

**Migration shape** — the `549` house pattern: `snapshot_agent()` before `BEGIN`, `DO`/`RAISE` pre-state
guards, the `UPDATE`, `DO`/`RAISE` post-state verify, `COMMIT`. The pre-state guard **gates on the live
query still ending `ORDER BY s.domain LIMIT 5`**, so a concurrent edit aborts the migration instead of
being silently overwritten — the mechanical answer to this row's unexplained 08:36Z `updated_at` bump.

### Part 2 — the framework half: a detector for the class

**Should this be built at all? The class has exactly ONE live member.** Argued rather than assumed:

- **For.** This is the reasoning `query_row_cap.go` already carries in its own header for the sibling
  check — *"one function, and the whole class becomes visible at once, including caps nobody has written
  yet"*. The shape is fully visible in config, so it is checkable offline, before a run pays for it. And
  the cost is mostly boilerplate: **every piece of machinery already exists** (`cmd/config-key-audit` hosts
  ~10 live-config audits behind flags, `fleetdb.go` gives a CronJob a direct-Postgres route,
  `validation.WalkSteps` does the traversal, and there are 14 precedent check services).
- **Against.** n=1 is a thin denominator, and a check that never fires trains readers to ignore it.
- **Settled: build it.** The deciding point is not the count but *how the defect was found* — by a human
  reading a query three days after a different bug happened to census it, and it had by then been
  starving one site for over a day. Nothing in the estate would have surfaced it otherwise, and
  LCO-009's WARN, which looks directly at this step, **cannot** see it: it counts rows, and the row count
  is identical whether the ordering is fair or not.
- **A repo-side / pre-commit check is NOT an option**, and this is already settled estate doctrine
  (RFC_006, and `single-owner-carriers-check`'s own docstring): at commit time a migration is unapplied,
  and config on this platform is routinely changed directly in the database with no commit at all. The
  only place the question has a true answer is live `agent_definitions`, on a clock.

**The rule, stated mechanically.** A `query_database` step is a **finding** when all three hold:

- **(a) capped** — the query carries a **trailing literal `LIMIT n`, `n >= 2`**, by exactly LCO-009's
  existing definition (end-anchored, tolerating a trailing `;` and SQL comments; a parameterised `LIMIT $2`
  and a LIMIT inside a subquery deliberately do not match);
- **(b) clock-replenished** — the query compares a column against `NOW()` / `CURRENT_TIMESTAMP`, i.e. its
  candidate set refills as time passes rather than draining as work is done;
- **(c) statically ordered** — its top-level `ORDER BY` names **no column appearing in (b)'s predicate**,
  and is not `random()`. **No `ORDER BY` at all also counts**, because an uncapped-order capped query
  returns rows in a heap/index order that is stable in practice, which starves exactly the same way.

Clean otherwise. **The predicate is the conjunction — (a) alone is LCO-009's job, not this one.**

**Stated false negatives** (a check's blind spots belong in its header, not in a later bug file): a
parameterised `LIMIT`; a LIMIT inside a subquery; a due predicate that reaches the query as a bound
parameter rather than a literal `NOW()`; and any ordering that is fair in expectation but not in fact.
**Stated false positives:** a `NOW()` comparison that is a sanity bound rather than a schedule, and a
static order over a set that drains for reasons the SQL cannot show.

**The effective cap is reported, not just the query's LIMIT.** Reporting `LIMIT 5` alone would have been
misleading here, because `process_sites` caps the same fan-out again at `max_iterations: 5`. The mode
resolves the consuming `loop` step (the one whose `items_field` dots into the capped step's
`output_field`) and reports `min(query LIMIT, loop max_iterations)`. A loop cap on its own is **not** a
finding — the finding is still (a)&(b)&(c) — but a reader who acts on the reported number needs the
honest one, and this is the trap that would otherwise make "raise the cap" a no-op.

**Where it lives.**

| file | what |
|---|---|
| `platform/orchestration/actions/query_row_cap.go` | export the trailing-LIMIT parse so there is **one** definition of "capped" |
| `cmd/config-key-audit/cappedscheduleordering.go` | the mode, `--capped-schedule-ordering` |
| `cmd/config-key-audit/cappedscheduleordering_test.go` | tests incl. mutation tests |
| `cmd/config-key-audit/main.go` | flag dispatch |
| `build/docker/backend/capped-schedule-ordering-check.dockerfile` | the image |
| `makefile` | build/push/deploy targets **and `RELEASE_IMAGES`** |
| `deployments/kustomize/services/capped-schedule-ordering-check/…` | daily CronJob + overlay |
| `docs026_concept_register/register/…` | the entry, in the same commit that ships it |

⚠ **`RELEASE_IMAGES` is not optional and not an afterthought.** The makefile comment says in capitals
that a new check service must be added in the commit that creates it; two check services were born
outside the list in the last two days and were invisible to `check-release-coverage`, because the gate
only polices overlays pinning an image already in the list. For a `config-key-audit` image a frozen image
is a frozen action *inventory*, so its clean report degrades silently.

⚠ **Do not duplicate the LIMIT regex.** `bugs_open/144`'s shape — two hand-written implementations that go
blind in the same direction and then agree with each other — is exactly what a second copy would create,
and every other mode in this binary is written to avoid it.

### The sequencing constraint, which is a correctness property

The detector must be **shown firing on the motivating case before the migration lands**. Once `552`
applies, the live row can no longer produce the positive control, and a detector whose only evidence is a
post-fix zero has been silenced by its author's own action — a documented failure class here.

The control does **not** require a fleet roll, because the binary reads a fleet export from **stdin**:

1. Build the mode and its tests. The unit fixture is the **verbatim pre-fix query**, already captured to
   `PREFIX_find_news_sites_query_2026-08-22.sql` precisely so the migration cannot destroy it.
2. Run the mode locally against a **live** fleet export, before applying anything. **Expected: exactly one
   finding, `content-feed-trigger.find_news_sites`.** Capture the output verbatim.
3. Apply migration `552`.
4. Re-run **the same binary, the same command**, against a fresh live export. **Expected: zero findings.**
   Only the config changed between (2) and (4), which is what makes the zero mean something.

The CronJob and image are the durable half and are **inert until the owner's next fleet roll** — stated
plainly, because "committed" is not "live" here.

### What is NOT in scope

**Capacity.** Demand 42/day against supply 20/day, 2.10x, re-derived from live rows today; removing the
cap entirely still leaves 36 vs 42. And **both caps must move together or neither moves** — the query's
`LIMIT` and `process_sites.max_iterations` are in series at 5 and 5. Raising only the first changes
throughput by nothing while making the cap-hit census stop reporting, so the instrument would show relief
that never happened. This trades ingest/LLM spend against freshness and is the owner's call; the
arithmetic is the input, and nothing here changes either literal.

**Separating provisioning from refresh.** Recorded as the design issue underneath decision 2 above.

### Verification, at the artefact

The bug file's disconfirming pair is already established: today ranks 1-5 are on time and 6-9 are late,
with `webdesign.co.uk` at 419% of its own cadence and absent from 5 of 5 runs. After the fix, re-run the
lateness query (RUNBOOK) and **the overdue set must not correlate with alphabetical rank**; if the same
names are still late, the fix did not land. **`webdesign.co.uk` is the sentinel**, not `relojistas.com` —
relojistas is currently served in 4 of 5 runs and would show nothing.

Live-on-apply: the migration. Waits for a roll: the detector image and its CronJob.

### Council

Two coherent tasks, two submissions, per the one-run-per-task rule:

- **A — the fix.** Admitted by `docs/agent_docs/sql_for_agents/552_*.sql` (migrations are in scope since
  2026-08-19). `_ROLLBACK` is out of scope and is not counted.
- **B — the detector.** Admitted by the `platform/` helper edit. Note `cmd/`, `makefile` and
  `deployments/` match neither scope regex, so the platform file is what carries it.

`DRY_RUN=1` tests admission for free before spending anything.

### Note on how this plan was produced

It was to be drafted by a `fable` agent given the full evidence pack. That agent **terminated on a session
limit before returning anything** (resets 14:10 BST). Rather than hold the work for three hours the plan
above is mine, written from the same evidence — which is in `NOTES` and reproducible from `RUNBOOK`.
Recorded because "who wrote this" is part of how much weight a later reader should give it.
