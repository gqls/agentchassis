# 275 — tool-suggester's `load_library_tools` LIMIT 30 silently hides 38 of 68 library tools, alphabetically

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
