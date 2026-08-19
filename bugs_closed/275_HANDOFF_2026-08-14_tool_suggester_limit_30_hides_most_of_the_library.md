# ✅ 275 (CLOSED 2026-08-19 — fixed, live, and proven at the artefact) — tool-suggester's `load_library_tools` LIMIT 30 silently hides 38 of 68 library tools, alphabetically

**Filed 2026-08-14, webdesign_uk_build_service lane**, found while shipping
migration 406 (the requires-backend gate) through the same query. Not run
through the 090 loop — stated substitution: the claim is direct arithmetic on
the live config and live data, both quoted below, with no mechanism inference;
every figure is reproducible by one query each.

## The defect

`tool-suggester`'s `load_library_tools` step (agent_definitions,
`{workflow,steps,load_library_tools,config,query}`) ends:

```
ORDER BY display_name LIMIT 30
```

Measured 2026-08-14 against the live library:

```sql
SELECT count(*) FROM content_components
WHERE component_level='tool' AND forked_from IS NULL
  AND is_active AND html_template != '';
-- 68
```

So the LLM `suggest_tools` step is shown the first **30 of 68** library
masters by `display_name` — **38 tools can never be suggested for any
site**, and which 38 is an accident of alphabetical order, not a judgement of
relevance. The cap is invisible in output: the suggester returns plausible
suggestions either way, so nothing ever looks wrong (a silent cap — the
"no silent caps" class).

## Why it matters

- Suggestion quality fleet-wide is judged against under half the library.
- The exclusion is alphabetical: a rename can move a tool in or out of the
  visible set, which will read as the suggester "deciding" differently.
- The library grows (68 and rising — 27 as of 2026-07-20 per
  `plan_sections_action.go`'s calibration comment), so coverage decays
  further with every added tool.

## Fix candidates (ranked by what closes the door)

1. **Remove the arbitrary cap and send a compact list** — the query selects
   five short columns; 68 rows of that is small. If the real constraint is
   prompt size, cap by TOKENS at the prompt assembly, not by row count in the
   dark. Closes the door: the visible set is the library.
2. **Rank before capping** — if a cap must stay, `ORDER BY` something
   meaningful (usage_count, avg_quality_score, category diversity), so the
   cut is a judgement rather than the alphabet. Door stays ajar (still a
   silent cap) but the damage is chosen, not accidental.
3. Do nothing to the query; document the cap in the step description. Weakest
   — a doc comment is not an enforcement mechanism.

Whoever fixes: migration on the same step migration 406 touched
(`sql_for_agents/406_tool_suggester_requires_backend_gate.sql` is the worked
example, including the snapshot + DO/RAISE verify pattern), and mind that 406's
gated query is now the base text — do not restore the ungated query by
copying an older sketch.

## How to verify a fix

Disagreeing pair: pick a tool sorting past position 30 by display_name
(e.g. anything starting late in the alphabet from
`SELECT display_name FROM content_components WHERE component_level='tool'
AND forked_from IS NULL AND is_active ORDER BY display_name OFFSET 30`),
run tool-suggester for any site, and confirm the late-alphabet tool appears
in the `library_tools` input of the `suggest_tools` LLM call
(`llm_call_log`, rendered prompt — the 2026-08-09 method). Before the fix it
cannot appear; after, it can.

---

## ✅ FIXED 2026-08-16 — the cap is gone and the class is now detectable (lane: `bugfix_275_silent_row_caps`)

Commit `eb137faed` (Go) + migration **445** (applied, live). Council corr
`b684a399-bb4d-4b1f-82f0-fe1429ebdceb` (`Council-Submitted:` — read the verdict before treating it as
reviewed).

### Two refinements to this file's own analysis

1. **Worse than filed, exactly as predicted.** 74 masters now, not 68 — **44 hidden, not 38**. Two days.
2. **The 406 gate is NOT the cause, and nobody had checked.** Exactly **1 of 74** masters carries
   `requires-backend`, and **3 of 40** sites have the capability. The gate narrows almost nothing; the
   cap was doing all of the hiding. Worth knowing, because if the gate HAD been narrowing heavily the
   right fix would have been a different one.

### The fix is candidate 1, done the way candidate 1 actually specifies

This file says: *"if the real constraint is prompt size, cap by TOKENS at the prompt assembly, not by
row count in the dark."* The constraint IS prompt size — 74 rows is 37,209 chars — and the measurement
names the knob:

| column | chars across 74 rows |
|---|---|
| **`description`** | **29,832 — 80% of the payload** |
| `id` 2,664 · `display_name` 2,100 · `function` 1,828 · `category` 785 | |

`description`: median 374, mean 403, max 2,526; 50 of 74 over 200.

So **bound `description`, not coverage**: `left(description, 200)` and drop `LIMIT 30`.

| | rows | chars |
|---|---|---|
| before | 30 (41% of library) | 16,421 |
| **after (measured live)** | **73** | **20,101 — +22.4%** |

73 not 74 on the no-backend path because the gate correctly excludes the one `requires-backend`
tool — 406's gate preserved byte-for-byte, which this file warns most about losing.

**200 was checked for MEANING, not just size** — the first 200 chars of the longest descriptions still
say what the tool IS. And the prompt template was read rather than assumed:
`- {{.display_name}} ({{.function}}, id: {{.id}}): {{.description}}`.

### Verified live

| check | result |
|---|---|
| `LIMIT` gone | ✅ |
| `left(description, 200)` present | ✅ |
| 406's requires-backend gate intact | ✅ |
| **tools sorting past position 30, previously unreachable** | **44 now visible** (first: *Early Settlement Estimator*) |

⚠ **Still owed — this file's own end-to-end proof.** The disconfirming pair asks for a late-alphabet
tool to be absent from `llm_call_log.prompt_rendered` before and present after. That needs a real
tool-suggester run, which has not happened yet. The config and population checks above are necessary,
**not sufficient**.

### THE CLASS — the larger half, and it is why this took a Go change too

A row-count `LIMIT` feeding an LLM prompt is **silent by construction**: the model returns plausible
output whether it saw 30 rows or 74. Census over live `agent_definitions`: **26** literal LIMITs in
query-shaped step configs; 19 are `LIMIT 1` (fetch-one idiom); **seven are multi-row caps** that can
bite —

| agent | step | cap |
|---|---|---|
| `tool-suggester` | `load_library_tools` | 30 ← fixed here |
| `internal-linker` | `load_candidate_pages` | 15 |
| `model-directory-trigger` | `find_directory_sites` | 12 |
| `tool-recreation-handler` | `load_related_context` | 10 |
| `content-feed-trigger` | `find_news_sites` | 5 |
| `visual-design-auditor` | `load_design_context` | 5 |
| `fix-proposer` | `load_last_bundle` | 2 |

All 26 run through `QueryDatabaseAction`, which now **WARNs when a result set reaches its query's own
trailing literal `LIMIT`**, naming the step (register **LCO-009**). Observational only — it cannot
change a query, a result or a row. `LIMIT 1` is excluded, or the channel would be noise.

⚠ **NOBODY HAS CHECKED WHETHER THE OTHER SIX BITE.** Each is one query. The WARN will answer it in
production once the Go half rolls — **it is NOT live yet** (Go is inert until the next chassis roll;
the migration half IS live).

### Adjacent, recorded not acted on

- **One library master has an EMPTY `display_name`** (and one an empty `category`). An empty string
  sorts FIRST, so that row has always occupied a slot in the visible 30 while telling the model
  nothing — its description is developer-facing notes. Content quality, another lane's call.
- `category` is selected and never rendered — 785 chars of dead payload, left alone deliberately.
- **`bugs_open/242` is this same class in another subsystem** (*"a capped render audit is
  indistinguishable from a complete one"*) and is still open.

### ⚠ THE OTHER CAPS: CHECKED — and TWO ARE THE SAME DEFECT, one worse than this one

The section above said *"nobody has checked whether the other six bite"*. Checked, 2026-08-16, same
session. **The naive census was itself wrong in two ways, and the corrected picture is:**

| step | cap | uncapped population | bites? | kind |
|---|---|---|---|---|
| `tool-suggester.load_library_tools` | 30 | **74** | ✅ fixed here | **CORPUS** |
| **`tool-recreation-handler.load_related_context`** | 10 | **107** (worst site) | **YES** | **CORPUS** |
| **`internal-linker.load_candidate_pages`** | 15 | **68** (worst site) | **YES** | **CORPUS** |
| `content-feed-trigger.find_news_sites` | 5 | 9 | yes, but | work queue |
| `model-directory-trigger.find_directory_sites` | 12 | 3 | **no** | work queue |
| `fix-proposer.load_last_bundle` | 2 | — | **n/a** | LIMIT is INSIDE a subquery |
| `visual-design-auditor.load_design_context` | 5 | — | **n/a** | LIMIT is INSIDE a subquery |

**Correction 1 — only FIVE of the "seven" are whole-result caps.** `fix-proposer.load_last_bundle`
wraps its `LIMIT 2` in a subquery and `string_agg`s the result into **one** row; `visual-design-auditor`
does the same inside a correlated subquery. The detector's end-anchored regex correctly ignores both,
and flagging them would have been **wrong** — their outer result is not bounded by that number. A real
case vindicating the anchor, found by checking rather than by argument.

**Correction 2 — and this is the part that matters for reading the new WARN: NOT EVERY MULTI-ROW CAP
IS A DEFECT.** The distinguishing question is not the size of the cap, it is what the rows ARE:

- **A WORK QUEUE** takes N per run and the rest arrive next run. `find_news_sites` (5 of 9) and
  `find_directory_sites` (12 of 3, `ORDER BY random()`) are this. Coverage is *eventual*, and the cap
  is a batch size. **Not a defect.**
- **A CORPUS shown to a model** takes N and the rest are **never seen on that run**, and the model
  answers confidently either way. That is this bug.

**The SQL cannot tell you which**, which is exactly why LCO-009's warning is deliberately dumb and the
reader makes the call. Expect it to fire on the work-queue steps; that is not a false positive in the
mechanical sense, it is the check asking a question only a human can answer.

### The two new instances — NOT fixed here, and needing an owner call

Both feed an LLM a truncated corpus, which is precisely this bug's mechanism:

1. **`tool-recreation-handler.load_related_context`, `LIMIT 10` against up to 107 pages.** The context a
   tool recreation is given. **Worse in ratio than 275 itself** (9% vs 41%).
2. **`internal-linker.load_candidate_pages`, `LIMIT 15` against up to 68.** The candidate set an LLM
   picks internal links from — so a page's link targets are chosen from at most 15 of 68, ordered
   `p.name`, i.e. alphabetically. Same accident, same invisibility.

**✅ 297 IS NOW FIXED AND LIVE (2026-08-17, `bugfix_297_tool_recreation_context` lane) — its file is now `bugs_closed/297_…`, council APPROVED r1 (`4b9265c3`).** Migration
453 drops its cap entirely — and the remedy **inverted this lane's shape**: nothing needed bounding
there (one short line per row; whole population at the worst site = 8,810 chars vs 735), so there is
no truncation and no marker. It also closed a defect the census had not looked for — the step's plain
`LEFT JOIN research_results` has no one-row guarantee and was already duplicating a page. Worth
carrying into 298: **measure the payload before assuming 445's column-bounding remedy transfers.**

**✅ NOW FILED (owner directed, 2026-08-17): `bugs_open/297` and `bugs_open/298`.** Re-measured from scratch before filing, and the numbers moved: 297's cap bites at **19 of 24 sites** (median site sees 10 of 26, worst 10 of 107 — **worse than 275 itself**), and 298's at **8 of 24** (median 12, under its own cap, so most sites are unaffected). ⚠ **298 is filed with a deliberately WEAKER claim**: `llm_call_log` has **zero** rows for `internal-linker` in all history, so whether its cap has ever shaped a link decision is **UNMEASURED** and that file says so rather than guessing. ⚠ My earlier count for 298 omitted the query's own `HAVING COUNT(pc.id) > 0` and over-counted — corrected in the ticket.

### ✅ COUNCIL APPROVED — round 2, corr `b684a399-bb4d-4b1f-82f0-fe1429ebdceb` (2026-08-17)

*Verdict READ before this line was written.* `decision: approved`, *"approved with 3 advisory
objection(s) — none high-severity"*, **`gated_by_truncation: false`**, 4 seats abstained. Round 1 was
REVISE, gated by `debug_historian`.

**Round 1 changed the shipped work** — it caught that 445 traded a silent ROW cap for a silent COLUMN
cap. Migration **446** (applied, live) now marks truncated descriptions ` […truncated]`: 49 of 73 rows
carry it, payload 20,738 chars, the signal costing ~3%.

**The three advisory residuals:**

1. **`editquality` (medium)** — *"446 never checks the live query text is currently 445's output before
   mutating it."* **The file DOES, at lines 52-56** (`position('left(description, 200)' in q) = 0` →
   `RAISE`). My SKETCH omitted it — the second time in two rounds I left a safety line out of a sketch,
   after writing that exact lesson into the round-2 rationale. Logged in `WRONG_CALLS.md`; the code
   needed no change.
2. **`editquality` (low)** — *"`LIMIT 30 -- note` is a false negative — silent under the very mechanism
   meant to end silent caps."* **Real and now FIXED**: the regex tolerates trailing `--` and `/* */`
   comments, with tests both ways (a comment must not HIDE a cap, and prose mentioning a limit must not
   INVENT one). Mutation-proven — reverting to the comment-blind pattern fails
   `TestATrailingCommentDoesNotHideACap`.
3. **`editquality` (missing)** — *"the schema_migrations INSERT is asserted as done outside the diff."*
   Correct: 445 and 446 were recorded by hand (`INSERT ... ON CONFLICT DO NOTHING`), not by an edit in
   the plan. Both rows exist; the point stands that a hand-recorded ledger entry is invisible to review.
4. **`bug_historian` (approve, with an observation worth carrying)** — Part A is WARN-only, and this
   lane's own census found **two live analogous caps**. Observation is not remediation: the warning
   makes them visible, it does not fix them. Those two remain open (above) and are an owner call.

**Trailer discipline:** the code commits carry `Council-Submitted:`; only post-verdict commits carry
`Council-Reviewed:`, written after reading the verdict body.

## §2026-08-17 — LCO-009 is LIVE at v1.0.1307, and the "BEFORE" half of this file's own proof is now done AT THE ARTEFACT

**The detector shipped.** v1.0.1307 (pods 17:05Z), verified at the binary with controls: the added
string **present**, a known-present control **present**, a plausible fake **absent**. ⚠ The previous
"fresh build" (14:42Z) shipped **nothing** — same-tag cached image — so a pod restart is not evidence;
see the `A same-tag rebuild leaves the OLD binary running` landmine.

### The disconfirming pair, "before" half — PROVEN, from a stored prompt

This file asks for a late-alphabet tool to be **absent** from `suggest_tools`' rendered prompt before
the fix and **present** after. The before half needs no new run — it is in `llm_call_log`:

Taking the most recent pre-fix prompt (2026-08-15 20:29:52) and ranking the library **as it stood at
that moment** (71 eligible masters, not today's 74):

| | value |
|---|---|
| library tools appearing in the prompt | **29** |
| eligible at that time | **71** |
| appearing **and** sorting past rank 30 | **0** |
| highest rank present | **exactly 30** |

**The model was shown the first 30 alphabetically and nothing beyond it** — 41 tools unreachable in
that specific, real run. That is the defect proven at the artefact rather than inferred from the SQL.

⚠ **Rank against the library AS IT WAS, not as it is.** Ranking today's 74 against that two-day-old
prompt reported *"1 tool past rank 30"* — a confound, because tools added since shift the ordering,
and it would have muddied the claim into "the cut was not strictly alphabetical". Constrain the
ranking CTE with `created_at <= <the prompt's timestamp>`.

### Still owed: the "after" half

**Zero `suggest_tools` runs since migration 445 went live at 11:22Z** (last was 2026-08-15). The after
half needs one real run; then re-run the query above against the new prompt and assert a rank > 30 tool
IS present. Until then this file's verification is **half done, and the half that is done is the half
that proves the bug — not the fix.**

### The live cap census is NOT yet answerable, and the zero is uninformative

`EQUALS the query's LIMIT` has fired **0 times** since the roll. **That is not evidence the detector is
silent** — the demand control says so: only **5** `query_database` completions have occurred since
17:05, all from `agent-landmine-verifier`, whose `load_entry` is `LIMIT 1` and correctly excluded. **No
capped step has run at all.**

`content-feed-refresh` (`find_news_sites`, cap 5, population 9) is on a 6-hourly schedule and last
fired 14:31Z, pre-roll — it is the first expected positive. The command:

```bash
for p in $(kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.spec.containers[0].image}{"\n"}{end}' \
           | grep 'agent-chassis:v1.0.130[7-9]' | awk '{print $1}'); do
  kubectl -n ai-persona-system logs "$p" --since=6h | grep "EQUALS the query's LIMIT"
done
```
**Always report the demand control beside the result** — a zero with no capped step executed says
nothing at all.

## §2026-08-18 — the live cap census RAN, and the answer is "no capped step has executed"

Detector re-verified in the running binary at **v1.0.1309** (pods 15:45Z, tag bumped, new digest;
added string present, known-present control present, plausible fake absent).

**Swept all 41 pods running the chassis image, 24h window:**

| | |
|---|---|
| cap WARNs | **0** |
| `query_database` completions (**DEMAND CONTROL**) | **21** |

**And every one of the 21 attributed, because the log line carries `step_name`:**

| step | runs | rows returned | capped? |
|---|---|---|---|
| `find_dispatchable_site` | 9 | 0,0,1,1,1,1,0,1,1 | `LIMIT 1` — excluded by design |
| `notify_scheduler` | 6 | all 0 | no cap |
| `load_entry` | 3 | 1,1,1 | `LIMIT 1` — excluded |
| `notify_scheduler_idle` | 3 | all 0 | no cap |

**NONE of the five whole-result caps ran.** So the zero WARN is still **uninformative** — it is not
evidence the detector is silent, it is evidence nothing it watches has executed. Reporting the 0
without the attribution table above would be a blind pass.

### ✅ The `LIMIT 1` exclusion is now EMPIRICALLY vindicated, not just argued

`find_dispatchable_site` returned **exactly 1 row on 5 of its 9 runs** — i.e. it sat on its own
`LIMIT 1` five times in 24 hours, on an unusually quiet fleet. **Without the `n >= 2` exclusion the
detector would have emitted 5 false warnings from that single step alone**, and it is a
dispatch-loop step that runs continuously. The signal-to-noise decision was made on reasoning
(`query_row_cap.go`); this is the measurement that confirms it, and it is a bigger margin than expected.

### ⚠ A REAL LIMITATION, found by running the census rather than by design review

**The WARN is a log line, so its history dies with the pod.** Observable window = time since the last
pod restart. Measured today: pods restarted **15:45Z**; `content-feed-refresh` (cap 5, population 9)
last fired **14:32Z** and `model-directory-publish` **12:15Z** — *both before the restart*. Their next
runs are ~20:32Z and ~18:15Z.

So with rolls landing roughly daily and the capped steps on **6-hourly** schedules, **a capped step that
fires shortly before a roll is invisible for ever.** The detector is correct and live; whether it will
actually *catch* the caps in practice depends on a race between roll frequency and schedule period that
nobody has characterised.

**This strengthens the case for the recorded follow-up** (the `LIMIT n+1` probe): a definitive result
written somewhere durable — a `doc_notes` row or an `llm_call_log`-style column — survives a roll,
where a log line does not. Noted as LCO-009's `verify-later`, not smuggled in here.

### Still owed, unchanged

**275's "after" half: 0 `suggest_tools` runs since migration 445 went live** (11:22Z on 08-17). The
before half is proven at the artefact (§2026-08-17); the after half needs one real run.

## §2026-08-18 (evening) — the cap census IS answerable, from a channel that was there all along

> **This section supersedes "§2026-08-18 — the live cap census RAN, and the answer is 'no capped step
> has executed'" and the limitation section under it.** Both were run honestly and both were reading a
> channel that cannot hold the answer.

### The instrument was the problem, twice over

**1. The window is 15–90 seconds, not "time since the last pod restart".** The container log rotates on
**size**, and the coordinator emits whole-state dumps (mean 2.2 KB/line; worst single line **183 KB**),
so it is wiped continuously while the pods keep running. [MEASURED] 2026-08-18, both chassis pods up
since 18:00Z with **0 restarts**: oldest retrievable line **3 s / 15 s / 21 s / 34 s / 91 s** across
five samples. No aggregator exists — `platform/logger/logger.go:37` is `OutputPaths: ["stdout"]`.

**2. The detector was not in the fleet for most of the window that census covered.** [MEASURED] the
oldest surviving replicaset carrying `v1.0.1309` — the first release containing the detector — was
created **2026-08-18 15:45:31Z**. A "last 24h" sweep run that afternoon therefore spanned mostly hours
with no detector running at all.

Two sufficient causes, one indistinguishable clean zero. **Neither of them is "no cap fired".**

⚠ This was **already in `LANDMINES.md`** — filed 2026-08-08 (0.4 s under load) and again 2026-08-14
(~90 s, with the liveness-control recipe, and the explicit note that *a log-only signal is not a durable
control on this fleet*). This lane built a log-only detector with both entries on file. The design
review, the council round and the author all missed it; grepping LANDMINES for `kubectl logs` before
choosing the medium would have caught it in one command.

### The durable census — no new code needed

`QueryDatabaseAction` writes its result to the step's `output_field`, which lands in
`orchestration_states.collected_data` and survives rolls. Every fact the WARN reports is there, per run,
retroactively. [MEASURED] over the ~2 days that table retains (5,701 rows; only 25 older than 2 days):

| agent | step | cap | runs measurable | max rows | **hit the cap** |
|---|---|---|---|---|---|
| `content-feed-trigger` | `find_news_sites` | 5 | 4 | 5 | **3** |
| `internal-linker` | `load_candidate_pages` | 15 | 2 | 15 | **1** |
| `model-directory-trigger` | `find_directory_sites` | 12 | 5 | 4 | **0** |

```
2026-08-17 20:31Z  content-feed-trigger   5 of cap 5   HIT
2026-08-17 22:22Z  internal-linker        7 of cap 15  under
2026-08-18 01:01Z  internal-linker       15 of cap 15  HIT
2026-08-18 02:32Z  content-feed-trigger   4 of cap 5   under   <- same agent, same query: not a constant
2026-08-18 08:32Z  content-feed-trigger   5 of cap 5   HIT
2026-08-18 14:32Z  content-feed-trigger   5 of cap 5   HIT
```

**Two independent negative arms**, which is what makes the positives mean anything: `model-directory-trigger`
never exceeded 4 against a cap of 12 across five runs — the negative control this file predicted — and
`content-feed-trigger` itself came back under its cap once.

⚠ **Handle both output formats or you re-create the blind zero in SQL.** A first pass counting only
JSON arrays reported **0 of 4** for `content-feed-trigger`; `output_format: object` puts the number in
`->>'count'`, and the true answer was 3 of 4.

**So: the caps in this file's class are now MEASURED to bite in production**, which no census had shown
before. What remains unobserved is the **WARN itself** — every one of those six runs predates 15:45Z.

### What has happened since the detector went live

Exactly **one** capped step has executed: `model-directory-trigger` at 18:15Z, **4 rows against a cap of
12** — no warning due, and the durable row is what says so (the log had already rotated, so the WARN's
*absence* was not observed either; an unobserved absence is not evidence).

**First genuine positive opportunity: `content-feed-trigger` at ~20:32Z.** [MEASURED] eligible
population **6 against a cap of 5** at 18:35Z — though the predicate includes `next_fetch_at <= NOW()`,
so it moves. A streaming capture was attached from 18:42Z (deadline 20:45Z), because with a 15–90 s
window that is the only way to read a WARN. **Whatever the stream catches, the `collected_data` row will
answer it whenever anyone next looks** — that is the point of the durable channel.

### Still owed: 275's "after" half, unchanged

[MEASURED] `suggest_tools` has run **0 times** since migration 445 (11:22Z on 08-17); last run in all
history **2026-08-15 20:29Z**, and it has been quiet on 08-16, 08-17 and 08-18. Its historical cadence
is 1–9 runs on roughly half of days, so this may clear on its own.

**Everything short of a real run is now verified at the live row, not at the migration file:** the live
`load_library_tools` config carries **no `LIMIT`** and bounds `description` to 200 chars with the
` […truncated]` marker; executing that exact query today returns **76 tools, of which 46 rank past
position 30** — 46 the pre-fix query could not have shown — with **54 descriptions marked truncated**
and a longest description of 213 chars. What is still unproven is only that a *prompt* carried them.

### §2026-08-18 20:32Z — the cap FIRED (first since the detector went live), and the WARN was still not witnessed

**Durable channel:** `content-feed-trigger-orchestrate-0818-2032`, created 20:32:54Z,
`news_sites.count = 5` against cap **5** — hit, five real domains in the payload, read while the run was
still in flight.

**Log channel: 0 anchored WARN lines**, against 133 anchored completions as a liveness control.
**That is not evidence the detector stayed silent** — the step's *unconditional* completion line is
missing too, so both of its lines went together and the failure is at the observation layer. Ruled out:
wrong pod (`processing_node` was a followed pod, and all recent trigger runs use the deployment pods),
stream disconnect (no reconnect marker; lines captured 30 s either side from both pods), old binary
(all 62 chassis-image pods on v1.0.1310), and `logs -f` being lossy in general (fidelity test: 561
streamed vs 291 retained over the same 60 s). **The miss is unattributed and recorded as such.**

**⚠ Instrument defect for whoever tries next:** `-l app=agent-chassis` matches **2 of 62** chassis-image
pods; the other 60 are `app=dynamic-agent` ephemerals that `agent-job-cleanup` deletes within minutes.
It did not cause tonight's miss, but a capped step running inside one would be unwitnessable by any
means.

**The conclusion the lane should carry:** the WARN is a hint for someone already watching;
`collected_data` is the record, and it answered the same question in one query, retroactively, with
controls.

### §2026-08-19 — five for five, and the class analysis had a blind spot this file should own

New build `v1.0.1314` (07:52Z) verified at the binary with **yesterday's stamp as a third control** —
absent, so this shipped new code rather than a cached image. Detector still an ancestor of the stamp.

**Census over the retained window: `content-feed-trigger` hit its cap on ALL FIVE runs**
(08-18 08:32/14:32/20:32, 08-19 02:33/08:33); `model-directory-trigger` **0 of 4** (4 rows against a cap
of 12). ⚠ Yesterday's "3 of 4" became "5 of 5" only because the one under-cap run aged out of the ~2-day
retention — **the fleet did not change; the window did.**

**And the cap census cannot answer the question that turned out to matter: WHO gets cut.**
`find_news_sites` orders `BY s.domain`, alphabetically and stably, so the same names win every
contention. Measured against each site's own configured cadence: ranks 1–5 are **0% late**, ranks 6–9
are **all late**, and the split is exactly the cap boundary. The worst-hit site is the one that asked to
be refreshed most often. The queue is also **2.10× oversubscribed** (42 demanded/day vs 20 supplied),
which removing the cap would not close (36 vs 42).

**Filed as `bugs_open/316`, and it narrows a claim this file's own work put into register LCO-009** —
that a work-queue cap means "coverage is eventual, not a defect". That was reasoning, not measurement,
and it told future readers to dismiss the case this bug turned out to have. **For every capped step, read
the `ORDER BY` as well as the count.**

## ✅ §2026-08-19 09:44Z — THE "AFTER" HALF IS PROVEN AT THE ARTEFACT, and the same run exposed a regression this fix caused

**Owner authorised a deliberate dispatch** (waiting could not have worked — `check_missing_tools.go`
applies a 30-day cooldown and every candidate site was evaluated 08-10..08-15, so the next natural run
was mid-September). Dispatched via a direct `orchestrate` message to `gamesdesign.co.uk` after the work
item route proved to be ~158 items deep and idling.

### The disconfirming pair is now COMPLETE

| | tools in prompt | past rank 30 | highest rank present |
|---|---|---|---|
| **BEFORE** (last pre-fix prompt, 71-tool library) | 29 | **0** | **exactly 30** |
| **AFTER** (2026-08-19 09:44:26Z, 81-tool library) | **80** | **51** | **81** |

Ranked against the library **as it stood at each prompt's timestamp**. The three highest-ranked tools in
the new prompt are `Write` (79), `XP Curve Designer` (80) and `Your notes from the old Noted` (81) — the
alphabetical tail the old `LIMIT 30` could never reach. Migration 446's truncation marker is present in
the prompt (`[…truncated]`), and the payload cost of the whole library is 33,818 chars against 24,327 for
thirty tools — **+39% for 2.7× the menu**, which is what bounding `description` bought.

**So the defect this file describes is fixed, live, and proven at the artefact.**

### ⚠ And the same run FAILED — filed as `bugs_open/319`

`suggest_tools` returned `stop_reason=max_tokens`: the answer hit the step's own 3,000-token cap and the
step errored. **This fix caused it.** A bigger menu means a longer answer, and the answer budget was
never raised to match — across 59 historical calls the output high-water mark was **2,921, i.e. 97.4% of
the cap**, *before* the menu tripled. The margin was already nearly gone and nobody (including me) looked.

**Nothing was corrupted** — the platform failed closed, discarded the truncated text, and never reached
`create_items_loop`, so **zero `add_tool` items** were created. That is CLAUDE.md's
`output_tokens == max_tokens` rule doing its job.

**The lesson for this file: bounding an INPUT payload can move the failure to the OUTPUT budget.** This
bug measured the prompt per column and never asked what the answer costs. See `bugs_open/319` for the fix
candidates and the one query that would have caught it before migration 445 shipped.

## ✅ CLOSED 2026-08-19

**Fixed, live, and proven at the artefact**, which is the bar. The disconfirming pair is complete
(before: 29 tools, 0 past rank 30, highest exactly 30 — after: 80 of 81, 51 past rank 30, highest rank
81), the truncation marker from migration 446 is present in the live prompt, and a full suggester run
now completes end to end (`COMPLETED`, 2026-08-19 10:25Z).

**Three descendants stay open and are NOT this bug:**

- `bugs_open/319` — the answer budget this fix overran. **Also fixed and live** the same day
  (migration 484), verified with the strict inequality: `success=true`, output 1,761 against a 6,000
  cap. Closed alongside this one.
- `bugs_open/321` — ~72% of the model's suggestions collide on a site-wide `item_key` and are silently
  dropped. **Open.** This is why widening the menu has produced fewer built tools than it should.
- `bugs_open/316` — the news-feed cap serves the alphabet (ranks 1–5 never late, 6–9 always). **Open.**
  Found by the same census method this lane built.

**And the lesson this bug paid for twice:** it bounded the INPUT payload carefully — per column, to
find `description` was 80% of it — and never asked what the ANSWER costs, nor what happens to the
answer once given. Both were one query away. `319` and `321` are the two halves of that omission.
