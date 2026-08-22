# NOTES — bugs_open/316, the news-feed cap serves the alphabet

Append-only, newest at the bottom. Technical log: evidence, commands, what the system said,
and every misstep.

---

## 2026-08-22 — session start: is the bug still valid?

**Ownership checked first.** `scripts/who-owns.py 316` returns OWNED-or-recently-active and names
`bugfix_275_silent_row_caps` as the likely owner. Read that lane's cold start — it is **CLOSED as of
2026-08-19 10:30Z** and its handoff lists 316 explicitly as one of *"three tickets that stand on their
own — they are not this lane and nobody owns them"*, suggested order **third**. So: not owned, free to
take. `git log` on the bug file shows one commit (`a996dbd73`, the filing) and nothing since.

**The defect is still live.** [MEASURED 2026-08-22 09:52Z, live `clients_db`]

```
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'find_news_sites')
FROM agent_definitions WHERE type='content-feed-trigger'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

still ends `... ORDER BY s.domain LIMIT 5`. Verbatim text saved to
`PREFIX_find_news_sites_query_2026-08-22.sql` in this directory — it is the fixture the detector's
positive control will be built from, and once the migration lands the live row can no longer supply it.

⚠ **`agent_definitions.updated_at` for this row reads 2026-08-22 08:36:05Z — today.** Recorded because a
fresh `updated_at` on a shared tree is exactly the shape that should make you look before assuming your
target is untouched.

> **CORRECTED same day, before writing anything:** this paragraph first read *"That is NOT another session
> fixing this … there are **zero snapshot rows** for `content-feed-trigger`"*. The zero was from
> `agent_definitions WHERE is_snapshot`, which is **one of two** places a snapshot can live —
> `snapshot_agent()`'s two-arg overload writes to `agent_definitions_backup` instead. Caught by grepping
> `LANDMINES.md` for my own footprints before touching anything (§*"`snapshot_agent` has TWO overloads
> writing to TWO different tables"*), not by a symptom. `agent_definitions_backup` is **also** empty for
> this agent, so the conclusion stands — but it stood on a check that would have returned the same zero
> had it been false. Logged in `WRONG_CALLS.md`.
>
> **The 08:36Z bump remains unexplained** and is left that way rather than tidied: no migration in
> `schema_migrations` touched this agent (549 at 09:56Z is the day's latest, unrelated), both snapshot
> tables are empty. The evidenced claim is the narrower one — **the step's query text is byte-identical
> to the one `bugs_open/316` quotes on 08-19**, still `ORDER BY s.domain LIMIT 5` — so the bump is inert
> with respect to this defect. That is not the same as "nobody touched it", and only the first is claimed.
> ⚠ **Re-read the live row immediately before applying the migration**, for the same reason.

## The bug has WORSENED, and the new figure is much sharper than the filed one

Census of the retained `orchestration_states` window (~2 days), by the method `bugs_open/275` established
(`collected_data`, not the logs — the chassis log retains 15-90 s):

| run (UTC) | n | domains returned |
|---|---|---|
| 08-22 08:38 | 5/5 | dartsonline, gaswholesalers, mortgagecalculator, relojistas, vetcomparison |
| 08-22 02:38 | 5/5 | ai-agent-orchestration, dartsonline, fundamentallyai, relojistas, robot-hands |
| 08-21 20:37 | 5/5 | dartsonline, gaswholesalers, mortgagecalculator, relojistas, vetcomparison |
| 08-21 14:37 | 5/5 | ai-agent-orchestration, dartsonline, fundamentallyai, relojistas, robot-hands |
| 08-21 08:37 | 5/5 | ai-agent-orchestration, dartsonline, fundamentallyai, gaswholesalers, mortgagecalculator |

**Every run at the cap, 5 of 5** — the bug file's central claim, reproduced on a fresh window.

**`webdesign.co.uk` appears in ZERO of the five.** It is alphabetically LAST of the nine eligible sites.
It has been continuously due since 2026-08-21 08:45Z and was therefore eligible at four consecutive runs
(14:37, 20:37, 02:38, 08:38) and selected at none of them. Last fetched **2026-08-21 02:45Z** — over 31
hours on a **6-hour** cadence, i.e. **419% of its own cycle**. The filed file recorded this same site at
**7%**. The starvation is not stable; it compounds.

`relojistas.com` (the file's sentinel, 3 h cadence) is currently **served** — it appears in 4 of 5 runs.
Its alphabetical rank is 6, so it sits just past the old cut and wins whenever fewer than five
alphabetically-earlier sites are due. That is the same mechanism, not a contradiction: the alphabet does
not decide who is late in every window, it decides who is late **whenever contention happens**, and the
site it hurts most is whichever one is furthest down the alphabet at that moment. Today that is
`webdesign.co.uk`, decisively.

## Fleet census: how big is the class?

Over every live, active, non-snapshot agent definition, every `query_database` step whose query carries a
`LIMIT`, extracting the trailing `ORDER BY … LIMIT n`. [MEASURED 2026-08-22]

Findings, classified:

- ~19 steps are `LIMIT 1` fetch-one / claim-one, correctly ordered (`created_at DESC`, or
  `priority ASC, created_at ASC FOR UPDATE SKIP LOCKED`). Not this class — LCO-009 excludes them for the
  same reason.
- `build-pipeline-trigger.find_dispatchable_site` — `wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1`.
  Correct FIFO.
- `visual-design-auditor.load_design_context` (LIMIT 5) and `fix-proposer.load_last_bundle` (LIMIT 2) —
  the LIMIT is **inside a subquery**; the outer result is one row. Correctly ignored, as LCO-009 already
  records.
- `meta-description-backfiller.load_pages_missing_meta` — `ORDER BY p.name LIMIT 25`. Alphabetical **and**
  capped, so it looks like the same defect — but its candidate set is **CONSUMED**: a page that gets a
  meta description leaves `WHERE COALESCE(p.meta_description,'') = ''`. Alphabetical order there is a
  batching delay, not starvation. **This near-miss is the useful one** — it is what forced the
  distinction below.
- `model-directory-trigger.find_directory_sites` — `ORDER BY random() LIMIT 12`. Recurring set, but
  `random()` does not systematically starve, and it returns 3-4 against a cap of 12 so the cap never
  binds. This is the bug file's negative control and it holds up.
- **`content-feed-trigger.find_news_sites` — `ORDER BY s.domain LIMIT 5`. The only live member of the
  dangerous shape.**

**The distinction the census forced** (and it narrows register LCO-009's *"coverage is eventual, not a
defect"* gloss more precisely than the bug file did):

> A cap on a query whose candidate set is **replenished by the clock** — rows never leave the set, they
> merely acquire a later due-time — starves the tail **permanently** under a static `ORDER BY`.
> A cap on a query whose candidates are **consumed** — a row leaves the set once served — is only a
> batching delay, and there "coverage is eventual" is true.

`find_news_sites` is the first kind. `load_pages_missing_meta` is the second. The count alone cannot tell
them apart, which is why LCO-009's WARN cannot.

## The platform already has the right idiom — one layer down

Not a new convention. Two Go call sites already order feed work by its due-time:

- `platform/orchestration/actions/dispatch_feed_sources_action.go:101` — `ORDER BY next_fetch_at ASC NULLS FIRST LIMIT $n`
- `platform/orchestration/actions/feed_actions.go:1016` — `ORDER BY next_fetch_at ASC NULLS FIRST`

Both select **sources within a site**. The **site**-selection query, which lives in config rather than in
Go, is the single layer that skipped it. That materially changes the fix's argument: it applies an
existing platform convention to the one place that missed it, rather than proposing a new one.

## A trap in the bug file's own fix candidate 1

The file proposes `ORDER BY min_next_fetch_at NULLS FIRST`. Taken literally that is **wrong**, and it
would be worse than the alphabet.

The eligibility predicate has two arms:

```
(NOT EXISTS (active content_sources for this site)   -- arm A
 OR EXISTS (an active source with next_fetch_at IS NULL OR <= NOW()))   -- arm B
```

A site matching **arm A** has no active sources at all, so `min(next_fetch_at)` is NULL **and it is
permanently eligible** — nothing a fetch does can advance a timestamp it does not have. Under
`NULLS FIRST` such a site would win **every run for ever**: a head-of-queue squatter, which starves
everyone deterministically instead of merely starving the alphabetical tail.

[MEASURED 2026-08-22] Zero sites are in that state today — all nine eligible sites carry 1-9 active
sources — so this is a latent trap, not a live one. The arm is in the query deliberately, so the fix must
answer it rather than delete it.

Second, smaller trap in the same expression: a source with `next_fetch_at IS NULL` has **never been
fetched** and is genuinely maximally overdue, but SQL `min()` skips NULLs, so a bare
`min(cs.next_fetch_at)` hides it behind its siblings' timestamps.

## A thing I checked rather than asserted

The repo seed `docs/agent_docs/sql_for_agents/090_b_content_feed_trigger.sql` carries
`WHERE s.deleted_at IS NULL`; the live query does not. That looks like drift worth restoring — it is not.
`sites.deleted_at` **does not exist** (`ERROR: column s.deleted_at does not exist`). The live query is
correct and the seed is stale. Recorded because "live config dropped a guard the seed has" is a
convincing-looking finding that took one query to refute, and the memory note is right: the seed is
history, the live row is fact.

## The starvation claim, checked against the trigger's OWN predicate (not my paraphrase of it)

The claim "`webdesign.co.uk` was eligible at four consecutive runs and picked at none" is only worth
anything if it was eligible by the **trigger's** definition. My lateness query uses a looser filter (it is
the denominator, deliberately), so it cannot settle this. Checked directly: [MEASURED 2026-08-22]

| | `webdesign.co.uk` | `ai-agent-orchestration.com` (control) |
|---|---|---|
| `news_feed.recommended` | true | true |
| deployed pages | **128** | 38 |
| active sources | 5 | 5 |
| sources due at 08-21 14:37Z | **5** | **0** |
| sources due at 08-22 08:38Z | **5** | **0** |

`webdesign.co.uk` satisfied **every arm** of the trigger's eligibility predicate at both runs and was not
selected at either. The claim holds.

**The control is the point of the table.** `ai-agent-orchestration.com` returns **0 due** at those same
two instants, which is what makes this a measurement rather than a formality: the query can come out
"not eligible", and for one of the two sites it does. Had both columns read 5 I would have learned
nothing — I would only have shown that my predicate matches everything.

## ⚠ THERE ARE TWO CAPS, IN SERIES — and the bug file only knows about one

`find_news_sites` caps at `LIMIT 5`. The step that consumes its output, `process_sites`, is a `loop`
action over `news_sites.rows` and carries **`max_iterations: 5`**. [MEASURED 2026-08-22]

```
SELECT default_config->'workflow'->'steps'->'process_sites'->'config'->>'max_iterations' AS loop_cap,
       substring(default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query'
                 from 'LIMIT ([0-9]+)') AS query_cap
FROM agent_definitions WHERE type='content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- loop_cap | query_cap
--        5 | 5
```

**This does not change the ordering fix** — reordering 5 of 9 is unaffected by a second gate at 5.

**It changes the capacity advice completely, and it is the trap in the bug file's fix candidate 2**
(*"raise the cap to >= 10"*). Raising only the `LIMIT` produces **no change in throughput whatsoever**:
the query would return 10 rows and the loop would silently process the first 5 and stop. Worse, it would
look like a fix that had been applied — the cap-hit census (`jsonb_array_length` of the step's own
output) would go from 5-of-5 to 10-of-10 and stop reporting a cap hit, while exactly as many sites got
refreshed as before. **The instrument would report relief that did not happen.**

So the owner-facing arithmetic needs restating: supply is `4 runs/day x min(query LIMIT, loop
max_iterations)`, and **both literals must move together or neither moves**.

This is the "guard in series" shape the estate's own memory warns about, found here by reading the
consumer of the step rather than the step. Recorded as a **correction to `bugs_open/316`'s fix candidate
2**, which names only the query cap.

Each loop iteration spawns and calls a `content-feed-orchestrator` per site
(`spawn_agent` -> `call_agent`, 600 s timeout), so the cap is a real spend gate, not a formality — which
is exactly why the capacity half is the owner's decision and not ours.

## Supply, counted rather than calculated — and the retention trap in doing so

The bug file derives supply as `4 runs/day x cap 5 = 20`. Counted directly from what the trigger
actually issued: [MEASURED 2026-08-22]

```
 day        | runs | site_slots_issued
 2026-08-22 |    2 |                10
 2026-08-21 |    3 |                15
```

⚠ **Do NOT read those daily totals as the throughput.** The `orchestration_states` window is ~2 days and
both ends are truncated — 08-21 is missing its earlier runs and 08-22 is only part-way through the day.
A reader taking "15 on 08-21" at face value would conclude the fleet supplies 15/day and that the
oversubscription is worse than filed. It is not; the window is short, not the fleet slow.

**What the count actually establishes, and it is the load-bearing part:** every run issued **exactly 5**
slots — 5 of 5 runs, no run under cap — and the run timestamps are 6-hourly (08:37, 14:37, 20:37, 02:38).
So supply is 4 x 5 = **20/day**, and the arithmetic in the bug file is confirmed rather than merely
restated. Demand re-derived from live rows the same day is **42** against 9 eligible sites, so
**2.10x** stands exactly.

## Council admission: the migration carries the submission, and `cmd/` does not

`scripts/council-scope.sh` (the single source, read 2026-08-22):

- `COUNCIL_SCOPE_CODE_RE='^(platform|internal|pkg)/'`
- `COUNCIL_SCOPE_MIGRATION_RE='^docs/agent_docs/sql_for_agents/[0-9]{3}_[A-Za-z0-9_]+\.sql$'`

So `docs/agent_docs/sql_for_agents/552_*.sql` admits the submission. **`cmd/` matches neither**, which
means a detector living in `cmd/config-key-audit/` would not, on its own, be admissible — it rides in on
the migration.

[OBSERVED, not filed] That generalises: **every check binary in this estate lives in `cmd/`**, so each
one has only ever reached the council bundled with a `platform/` or migration change. Recorded as an
observation about the gate's reach, not as a defect of this lane's — it is not mine to widen, and
`DRY_RUN=1` on the 097 trigger tests admission for free before spending anything.

Also noted while checking the class: register LCO-009 describes `bugs_open/242`
(*"a capped render audit is indistinguishable from a complete one"*) as *"the same class in a different
subsystem, **still open**"*. It is **closed** — `bugs_closed/242_…`, commit `03640f491`, *"fixed AND live
since v1.0.1288"*. A dead pointer to fix while editing that entry, and an instance of the class the
estate measured at 71.5% of directory-prefixed bug pointers already dead.

⚠ And it does **not** rescue the detector's justification: 242 has no ordering angle (no `ORDER BY`,
`alphabet` or `starv` hit anywhere in the file). The ordering class stays at **one live member**. The
count class LCO-009 already covers is larger, but that is a different check and it exists.
