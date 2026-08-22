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

## The fix, dry-run against live data before applying anything

Both queries run read-only at the same instant, 2026-08-22 ~10:35Z.

**OLD (`ORDER BY s.domain LIMIT 5`)** — `webdesign.co.uk` is **rank 5 of 5**, i.e. last, which is where
it has been for four consecutive runs:

```
ai-agent-orchestration.com
dartsonline.com
fundamentallyai.com
robot-hands.com
webdesign.co.uk
```

**NEW (`ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5`)** — `webdesign.co.uk` is **rank 1**:

```
webdesign.co.uk              <- the starved site, promoted from last to first
ai-agent-orchestration.com
dartsonline.com
fundamentallyai.com
robot-hands.com
```

**The two results contain the SAME FIVE SITES.** That is the check that matters and it is not a
coincidence — the eligibility predicate is byte-for-byte unchanged, so membership is provably identical
and only the order moves. If the new query had returned a different *set*, the rewrite would have
changed what the trigger considers, which is not what this fix is for.

⚠ **Right now exactly 5 sites are due, so nothing is being cut at this instant** and the ordering costs
nobody anything. The ordering only decides an outcome when more than five are due — which is the
condition that held at all five retained runs. So this dry-run demonstrates the *promotion*, not the
relief; the relief is verified after the fact against the lateness query.

## Prior art I should have found sooner: SCH-025, and the property it says a due-ordered queue must have

Found while writing the register entry, which is later than it should have been — the register is exactly
what you are meant to consult *before* concluding a mechanism does not exist.

**SCH-025 (`site_discovery_rotation`)** already solved this class on this estate: starvation in a
recurring scheduled queue, fixed by picking *the least-recently-selected* site from a stamp table. It
names `IMP-010` as "the starvation this design answers" and `SCH-008` / `bugs_open/048` as the shape.

**This does not make the fix redundant, and I checked rather than assumed.** SCH-025 had to *invent*
rotation state because discovery had none. The feed pipeline already has it —
`content_sources.next_fetch_at` **is** the rotation stamp. Adding a second stamp table would be the
duplicate-mechanism smell; reading the one that exists is the cheap correct answer. So: same class, same
remedy, lighter instrument because the state was already there and only the reader was ignoring it.

**The transferable warning in SCH-025 is the one worth acting on:**

> *"the stamp records **selection, not completion**, so a site whose run fails cannot pin the rotation
> head (the SCH-008/`bugs_open/048` starvation shape)"*

That is a real hazard for due-ordering. If `next_fetch_at` only advanced on SUCCESS, a permanently
failing site would sit at the head of the queue for ever under my fix — the same squatter I rejected
`NULLS FIRST` for, but *reachable* rather than latent. **So I read the writers rather than assuming:**

| writer | what it does to `next_fetch_at` |
|---|---|
| `dispatch_feed_sources_action.go:272-279` | **on DISPATCH, optimistically** — `NOW() + fetch_interval`, before the work completes, commented *"to prevent re-dispatch before completion"* |
| `feed_actions.go:1112` (failure path) | `NOW() + (fetch_interval * LEAST(error_count + 1, 4))` — advanced **with exponential backoff** |
| `feed_actions.go:1124` (success path) | `NOW() + fetch_interval` |

**All three advance it.** There is no path on which a site is served and its key stays put, so a failing
source cannot pin the head — it is pushed forward at dispatch and pushed *further* on failure. The fix
inherits SCH-025's required property for free.

This also sharpens the argument for the fix generally: the ordering state was **already being maintained
correctly by the platform all along**. `find_news_sites` was the one reader that ignored it.

## Council verdict on the migration — APPROVED, and it found a real hole anyway

Corr `e6e8b923-f614-4a1e-97d8-bf40fb5e3cc3`. **APPROVED**, *"approved with 2 advisory objection(s) —
none high-severity"*, `gated_by_truncation: false`, 9 reviewers.

> **CORRECTION to my own reading, same hour:** I first reported this as *"8 of 9 seats abstained"*, from
> `metadata->>'abstained' = 8` in the summary object. That is wrong — the report body carries **nine
> substantive reviews** with written notes. Whatever the `abstained` counter is counting, it is not
> "seats that did not review". Recorded because I nearly told the owner the approval was thin when it was
> the opposite, and because reading a summary counter instead of the artefact it summarises is the same
> mistake as trusting a status over an artefact.

**The one that mattered — `debug_historian`, medium, and it is right:**

> *"The pre-state guard only anchors on the QUERY'S TAIL … then overwrites the ENTIRE query string with a
> hardcoded new literal. If any other part of the live WHERE/predicate had drifted since the author's
> snapshot … the guard would pass and the wholesale replacement would silently revert that unrelated
> change. Needle-gate discipline calls for gating on the full known pre-state."*

That is exactly the hazard I *thought* I was guarding against — this row's unexplained 08:36Z
`updated_at` bump — and my guard only covered the region I was changing. A concurrent edit to the
news_feed test, the deployed-page `EXISTS`, or either arm of the eligibility predicate leaves the
`ORDER BY` tail untouched, so my guard would have passed and my rewrite would have thrown that change
away. **Fixed:** the guard is now `q IS DISTINCT FROM $pre$<the whole captured query>$pre$`, and the
literal is the verbatim pre-state already committed as `PREFIX_find_news_sites_query_2026-08-22.sql`.
Verified against live before applying — the equality returns `t`, so the migration will not spuriously
abort.

Also taken: **`editquality`, low** — the duplicate-active-row abort now names the colliding `version`s,
because aborting safely is not much use if the operator then has to go and find out which rows collided.

**Raised and answered rather than changed:**
- **`editquality` + `debug_historian`, medium (both):** the `_ROLLBACK` file is described in the rationale
  but absent from the `edits` array. It **exists and is committed** (`95635d09b`) — `_ROLLBACK` is
  explicitly out of council scope, so it could not have been an edit. A submission-hygiene point, and a
  fair one: the rationale should have said "committed alongside, out of scope" rather than describing it
  as if reviewable.
- **`guardian`, low:** should have used `operation: "config_change"` rather than `"add"`. Noted for next time.

**Three checks the seats asked for, run rather than asserted:** [MEASURED 2026-08-22]

| asked by | check | result |
|---|---|---|
| `guardian` | does any OTHER agent carry this query text? | **1 row** — `content-feed-trigger.find_news_sites` only |
| `prior_art_librarian` | re-run "zero live instances of arm A" | **0** |
| `prior_art_librarian` | does `check_has_sources` genuinely cover arm A? | **yes, and stronger than I claimed** |

The third is worth spelling out because the seat was right to doubt it: `check_has_sources` reads
`seed_result.has_sources`, i.e. it reads the output of a step that has *already run*. The orchestrator's
`start_step` is **`seed_sources`** (`seed_content_sources`) → `check_has_sources` → `dispatch_sources` /
`complete_no_sources`. So seeding is the **first thing that happens to every selected site**, not a
branch reached only by arm-A sites. Arm A really is the provisioning path, and the NULLS LAST reasoning
stands on a verified premise rather than a plausible one.

## Applied 2026-08-22 10:54Z, and verified

**Applied by hand** (`psql -f`), **not** via `run-migrations.sh --apply` — that takes EVERY pending file,
and two other sessions had `552`/`553` pending. Recorded afterwards through the supported out-of-band
path: `run-migrations.sh --record-only 554_… --note "…"`, so the ledger is honest about how it was
applied (`applied_by='record-only'`).

Apply output: snapshot captured → `BEGIN` → `DO` (pre-guards) → `UPDATE 1` → `DO` (verify, including the
`EXPLAIN`) → `COMMIT`. Exit 0.

**The snapshot holds the PRE-change value**, which is the landmine's check rather than "does a snapshot
exist": `agent_definitions_backup`, taken 10:54:08Z, `holds_the_OLD_query = t`.

**The detector pair — this is the demand control, and it is why the zero means something:**

| | agents scanned | undecoded | findings |
|---|---|---|---|
| **before** the migration | 194 | 0 | **1** — `content-feed-trigger.find_news_sites` |
| **after** the migration | 194 | 0 | **0** |

Same binary, same command, same scan size. Only the config changed between them. Both outputs committed
(`CONTROL_prefix_…` / `CONTROL_postfix_…`).

⚠ **The post-fix run FAILED on its first attempt, and failed in the right direction.** The `kubectl exec`
export was truncated mid-stream (`unexpected EOF`), and the detector **refused** — exit 2,
*"stdin is not a JSON array of agents … A truncated kubectl exec exits 0, so a short read arrives here
looking like a small fleet."* Had it parsed leniently it would have printed a clean report over a partial
fleet, which is exactly the blind-green this whole class of check exists to avoid. **The refusal path
earned its place on its first real use**, unplanned. The retried export was 1,216,896 bytes against the
pre-fix 1,216,689 — comparable, so the second read was whole.

**The installed query, executed as the runtime will execute it** (read back out of the live row and
`EXECUTE`d, rather than re-running my own copy of the SQL):

```
selected: webdesign.co.uk            <- was rank 5 of 5, and absent from 5 of 5 runs
selected: ai-agent-orchestration.com
selected: dartsonline.com
selected: fundamentallyai.com
selected: robot-hands.com
```

### ⚠ What is NOT yet verified, and cannot be until 14:37Z

Everything above shows the **query** now orders correctly. None of it shows the **starvation relieved** —
that needs a real trigger run. The trigger is 6-hourly (08:38, 14:37, 20:37, 02:38) and the next one is
**14:37Z**. After it, run the lateness query in the RUNBOOK and check:

1. `webdesign.co.uk` appears in the 14:37Z run's `news_sites` rows (it is currently rank 1, so it should);
2. its `last_fetched_at` moves off 2026-08-21 02:45Z;
3. **the overdue set no longer correlates with alphabetical rank** — that is the bug file's own
   disconfirming test, and it is the one that matters. If the same names are still late, the fix did not
   land.

⚠ **Do not read a single good run as the fix working.** Nine sites, five slots, 2.10x oversubscribed:
after `webdesign.co.uk` is served, whoever is then most overdue goes next. The claim to test is that
lateness ROTATES, not that it disappears — it cannot disappear while demand exceeds supply, and saying it
had would be claiming the capacity half was fixed too.

## Council round 2 — the detector. APPROVED, and one objection was a real runtime risk

Corr `703dbe2f-a078-4a40-825e-fb7773a1d95b`. **APPROVED**, *"approved with 1 advisory objection(s) —
none high-severity"*, **11 reviewers**, `gated_by_truncation: false`.

**The one that could have bitten — `editquality`, medium:** the `--report` path writes a `doc_notes`
row, and *"`doc_notes.subject_type` is CHECK-constrained to eight values, and daily-check inserts
routinely fail live despite passing locally"*. If that write failed, the CronJob would run, find
nothing, and record nothing — and a missing row is defined in this estate to mean *the job did not run*.
The failure would be invisible and would invert the signal.

[MEASURED 2026-08-22] The constraint allows
`tool, pipeline, experience, action, experience-pattern, landmine, component, decision`, and
`writeDocNote` uses **`'pipeline'`** — 1,878 live rows already do. **Proved rather than inferred**: the
exact insert was executed against live `clients_db` inside a transaction and then **rolled back**
(`after rollback, rows=0`), so the write path is confirmed with no misleading row left behind. A real row
written now would later read as "the cron ran on 08-22", and the row's entire meaning is that a run
happened.

**What actually made it safe is worth naming: I reused `writeDocNote` instead of writing my own insert.**
The seat's concern was correct about the class; it missed because the shared helper already had the value
right. That is the "reuse existing machinery" rule paying off in a place I had not thought about.

**Other objections, all low, and what was done:**

| seat | objection | response |
|---|---|---|
| `guardian` | confirm `QueryRowCap` does not collide with anything `cmd/config-key-audit` already imports from `actions` | **checked**: the binary uses exactly two symbols from that package, `GlobalActionRegistry` and `QueryRowCap`. No collision (and it compiles, which is the stronger proof) |
| `debug_historian` | no deploy-verification step; a CronJob image is subject to the same-tag-rebuild trap like any other | **taken** — pod-binary recipe with a negative control added to the RUNBOOK above |
| `prior_art_librarian` | the live-fleet figures are asserted; re-confirm `max_iterations=5` and the `ORDER BY` before merge, since `agent_definitions` can change under a session | **re-run**: `max_iterations=5 | query_limit=5`, unchanged |
| `tooling_provenance` | no travelling `doc_plans`/`doc_notes` record read or left for the `config-key-audit` subject | noted, not done — worth a lane's attention if this tool family keeps growing |
| `editquality` | a daily CronJob is disproportionate for a class with one live member | acknowledged in the submission and unchanged. It is the fair objection and it stays on the record; SCH-027's `verify-later` names the honest denominator so a later reader can judge it |

**Submission hygiene, twice now.** Both rounds objected that files described in the rationale were absent
from the `edits` array — the `_ROLLBACK` in round 1, and the test file, dockerfile, CronJob and makefile
entries in round 2. All of them exist and are committed; several (`_ROLLBACK`, `cmd/`, `makefile`,
`deployments/`) are **out of council scope** and could not have been edits. **The lesson is about
wording:** say "committed alongside, out of scope, not reviewable here" rather than describing a file as
if the reviewer can see it. Describing an invisible file as part of the safety net reads, correctly, as a
gap.

⚠ **Commit `d7be8db66` (the detector's Go) will list as UN-REVIEWED in the 098 report** despite this
approval: it carries the in-scope `platform/` file but predates the submission and so has no trailer,
and forward-only forbids an amend. The work *was* reviewed — corr `703dbe2f` — and this note is the
resolution for anyone reading that list. The correct habit, which I used on the later commits and not on
that one, is `Council-Submitted: <corr>` at commit time.

## Capacity: OWNER DECISION 2026-08-22 — "increase the capacity with both caps together". Migration `556`, applied 11:2xZ

Council **APPROVED unanimously** — corr `2cfe6fbd-c7da-4f63-ba22-9883305c38df`, *"all reviewers
approve"*, 10 reviewers, one low objection.

**Both literals 5 → 10 in ONE migration**, nested `jsonb_set` so they cannot diverge, with a verify block
asserting both moved. That pairing is the whole point: raising the query `LIMIT` alone moves throughput
by nothing while the cap-hit census flips from *"5 of 5, at the cap"* to *"10 of 10, under the cap"* and
stops reporting — the instrument would show relief that never happened.

**Why 10:** 9 eligible sites, so any cap ≥ 9 stops binding; 10 gives one slot of headroom and still binds
at an eleventh site, which is wanted — that is when LCO-009's WARN should fire and ask again.

### ⚠ The bug file's headline figure is a POOL, and the pool framing is misleading

*"42 demanded vs 20 supplied, 2.10×"* implies a bigger cap could close the gap. **Per site it cannot.**
The trigger fires every 6 h (`scheduled_tasks.content-feed-refresh`, `interval_seconds = 21600`), so no
site can be served more than **4×/day at any cap**. [MEASURED 2026-08-22]

| | wants/day | ceiling at a 6-hourly trigger | after `556` |
|---|---|---|---|
| 7 sites (6 h cadence) | 4 | 4 | **fully satisfied** |
| `dartsonline.com` (4 h) | 6 | 4 | capped **by frequency** |
| `relojistas.com` (3 h) | 8 | 4 | capped **by frequency** |

So the residual shortfall is exactly **6 fetches/day**, it belongs entirely to those two sites, and it is
a **trigger-frequency (or cadence) decision, not a cap decision**. Deliberately not taken.

### Cost, measured

Supply 4 × 5 = **20** → 4 × 9 = **36** site-refreshes/day (+80%). The LLM component is `feed-triage`:
**114 calls / 7 days, avg 2,780 in / 1,992 out, 544k tokens** = ~78k tokens/day, scaling to ~140k
(+~62k/day).

> ⚠ **MISSTEP, caught by its own tell.** My first version of that query was
> `WHERE created_at > now() - interval '7 days' AND agent_type ILIKE '%feed%' OR agent_type ILIKE '%triage%'`
> — `OR` binds looser than `AND`, so it read `(recent AND feed) OR (triage, ALL TIME)` and returned
> **911 calls** with `first_seen 2026-03-30`. **A March date inside a 7-day window is the tell**, and it
> is the only reason I looked. Parenthesise every mixed `AND`/`OR`, and sanity-check that the returned
> range fits the window you asked for.

### Verified after apply

- both caps read **10**, ordering from `554` **intact** (`ORDER BY due_at ASC NULLS LAST` still present);
- the class detector re-run over the live fleet: **194 agents, 0 findings** — the cap change did not
  reintroduce the ordering defect;
- the installed query, executed as the runtime will, now returns **5 rows against a cap of 10**. Before
  `556` every run returned exactly 5 of 5. **Returning fewer rows than the cap IS the observable that
  the cap has stopped being the constraint** — the remaining 4 sites are simply not due yet.
- guard proven to **abort**, not merely exist: applying the body twice in one transaction raises
  *"not byte-identical to migration 554's installed query"* and leaves the live row at 5/5.

### ⚠ A THIRD cap exists on a different axis, and it is one source from binding

Raised by the council's `guardian` seat (low) and checked rather than deferred:
`DispatchFeedSourcesAction` reads `max_dispatches` with a **default of 10** — how many SOURCES are
dispatched **per site** — and `content-feed-orchestrator.dispatch_sources` sets only `site_id`, so the
default applies. The busiest eligible site currently has **9 active sources**. 9 < 10, so it does not
bind today, but **one more source on that site and it silently caps** on an axis nobody is watching.
Not fixed here (different axis, no live effect); recorded so it is not rediscovered from a symptom.
