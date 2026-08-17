# PLAN 2026-08-16 — bugs_open/275: the tool-suggester cap, and the SILENT-CAP CLASS behind it

Bug: `bugs_open/275_HANDOFF_2026-08-14_tool_suggester_limit_30_hides_most_of_the_library.md`.

Ownership checked two ways before starting (the `bugs_open/257` method): `who-owns.py 275` names
`webdesign_uk_build_service` as the FILING lane, not an active fixer; and a grep of ~25 live session
transcripts (the only instrument that sees uncommitted work) shows **no session on 275**. Actively
worked elsewhere right now: 251, 270, 153, 253, 286, 284, 287, 083, 277, 072, 145, 283, 285, 289,
225, 288, 271, 268, 278, 149, 177, 098, 248, 113, 122.

## 1. Still valid, and WORSE than filed

| | at filing (2026-08-14) | measured 2026-08-16 |
|---|---|---|
| library masters | 68 | **74** |
| shown to the LLM | 30 | 30 |
| **hidden** | 38 | **44** |

The query still ends `ORDER BY display_name LIMIT 30`, with migration 406's requires-backend gate in
the base text exactly as the bug warned.

**Refinement the bug did not make — the gate barely narrows anything.** Only **1 of 74** masters
carries `requires-backend`, and only **3 of 40** sites have the capability. So the effective picture is
**43 of 73 hidden for a no-backend site, 44 of 74 for a backend one** — the cap, not the gate, is doing
all the hiding.

## 2. The real constraint is PROMPT SIZE, and it is `description` — measured, not assumed

The bug's candidate 1 says *"if the real constraint is prompt size, cap by TOKENS at the prompt
assembly, not by row count in the dark."* That is exactly right, and the data says which knob:

| column | total chars across 74 rows |
|---|---|
| `description` | **29,832 (80% of the payload)** |
| `id` | 2,664 |
| `display_name` | 2,100 |
| `function` | 1,828 |
| `category` | 785 |

`description`: median **374**, mean 403, max **2,526**; 50 of 74 exceed 200 chars.

### The decisive comparison

| variant | rows | chars |
|---|---|---|
| **TODAY** — 30 rows, `description` uncapped | 30 (41% of library) | **16,421** |
| **PROPOSED** — all 74, `description` ≤ 200 | **74 (100%)** | **20,376** |
| all 74, `description` ≤ 300 | 74 | 25,146 |

**The whole library for +24% payload (~4,000 chars, ~1k tokens).** That is the argument for this fix,
and it is arithmetic on live data, not a preference.

**200 chars was checked for MEANING, not just size** — read the first 200 chars of the three longest
descriptions: *"The Arena is Spark's competitive mode, v1 as a fully self-contained client-side
experience (no fetch calls, no backend). Four elements, in order: (1) TODAY'S PROVOCATION…"* and
*"Companion to the Bridging Loan Calculator, per the owner ruling of 2026-08-08: where two calculation
models are both right in different ways, supply BOTH…"*. The first 200 chars carry what the tool IS,
which is what a relevance judgement needs.

### How the prompt actually renders it (read, not assumed)

```
{{range .library_tools}}- {{.display_name}} ({{.function}}, id: {{.id}}): {{.description}}
```

⚠ **`category` is SELECTED and never RENDERED** — 785 chars of dead payload. Deliberately left alone:
removing a column another consumer might read is scope creep with a non-zero risk, for a 2% saving.
Recorded, not acted on.

## 3. THE CLASS — and this is where the framework leverage is

The user's standing instruction is to prefer a fix applicable to the framework over the individual
case. Bug 275 is one instance of: **a row-count `LIMIT` feeding an LLM prompt is a SILENT cap — the
model returns plausible output whether it saw 30 of 74 or all 74, so nothing ever looks wrong.**

Census of every literal `LIMIT` in a query-shaped step config across live agents — **26 hits**. Most
are `LIMIT 1` (fetch-one / claim-one idiom, by design). The **multi-row** caps, which are the ones that
can silently bite:

| agent | step | cap |
|---|---|---|
| `tool-suggester` | `load_library_tools` | **30** ← this bug |
| `internal-linker` | `load_candidate_pages` | 15 |
| `model-directory-trigger` | `find_directory_sites` | 12 |
| `tool-recreation-handler` | `load_related_context` | 10 |
| `content-feed-trigger` | `find_news_sites` | 5 |
| `visual-design-auditor` | `load_design_context` | 5 |
| `fix-proposer` | `load_last_bundle` | 2 |

**Every one of these runs through ONE function**: `QueryDatabaseAction`
(`platform/orchestration/actions/database_actions.go:11`), registered as `query_database`. It already
logs `zap.Int("count", len(results))` and says nothing about whether that count hit a ceiling.

## 4. DECISION — two parts, deliberately separable

### Part A (framework, Go): make every silent cap in the estate VISIBLE, at the one shared point

In `QueryDatabaseAction`, parse the query's trailing literal `LIMIT n` and, when
`len(results) == n && n >= 2`, log a **WARN** naming the step, the limit and the row count.

- **Observational ONLY. No behaviour change, no new config key, no new authority on the seam** — it
  cannot alter a single query result. That is deliberate: this is a shared action serving 26+ live
  steps, and the owner ruling of 2026-08-02 §2 / RFC_010 says new authority on a shared seam ships
  opt-in with the unsafe default OFF. A log line takes no authority at all.
- **`n == 1` is excluded, and that is the whole signal-to-noise decision.** 19 of the 26 hits are
  `LIMIT 1` — the fetch-one/claim-one idiom, which returns 1 row *by design* and would warn on every
  execution for ever. A cap that "bites" at one row is not a truncation, it is a lookup.
- **Why equality rather than a definitive count:** knowing "30 of 74" would need a second `COUNT(*)`
  per step (doubling DB work fleet-wide) or rewriting the SQL to `LIMIT n+1` and trimming — the latter
  changes what a shared action returns, which is a far bigger blast radius than this bug justifies.
  Equality is a cheap, honest *suspicion* signal: it is exactly true that the result may be truncated,
  and it names the step so a human can check in one query. The `n+1` probe is recorded as the
  follow-up, not smuggled in.
- **Known false positive, stated rather than discovered:** a population that happens to equal the cap
  exactly warns while nothing is hidden. That is the correct trade for a warning — a false "check
  this" costs one query; a false silence costs 44 tools.

**Reuse checked BEFORE writing** (the `reuse_agent` seat's objection on `bugs_open/257`, learned):
`grep` over `platform/` finds no existing LIMIT-parsing or result-cap detection. The `Truncate*`
helpers are all string truncation; `markTruncated` is LLM `stop_reason` handling; `LimitedRead` is
HTTP body bounding. Nothing to reuse — checked, not assumed.

### Part B (this bug, migration): give the suggester the whole library

Rewrite `load_library_tools`: `left(description, 200) AS description`, drop `LIMIT 30`, **keep
migration 406's requires-backend gate byte-for-byte** (the bug explicitly warns not to restore an
older ungated sketch).

⚠ **DB config is LIVE IMMEDIATELY** (CLAUDE.md) — unlike Part A, this needs no roll and has no
image-based rollback. Snapshot first; `406_tool_suggester_requires_backend_gate.sql` is the worked
pattern, including the `DO`/`RAISE` verify block (a verify block of bare `SELECT`s cannot stop a
`COMMIT`).

## 5. Adjacent findings — recorded, not acted on

- **One library master has an EMPTY `display_name`** (and one an empty `category`). An empty string
  sorts FIRST, so that row has always occupied a slot in the visible 30 while telling the LLM nothing —
  its description is developer-facing implementation notes (*"Parameterised calculator component
  (Track B2, bugs_open/263): panel, ids and scripts live in this template…"*). Data quality, not this
  bug's mechanism, and fixing content is a different lane's call.
- `category` selected but never rendered (§2).

## 6. How to verify

- **Part A:** unit tests on the LIMIT parser and the warn condition, each **mutation-proven**; and the
  `n == 1` exclusion asserted explicitly, because that is the arm most likely to be "simplified" away.
- **Part B:** the bug's own disconfirming pair — take a tool sorting past position 30 by
  `display_name`, confirm it CANNOT appear in the `suggest_tools` rendered prompt before, and CAN
  after (`llm_call_log.prompt_rendered`). Plus the payload arithmetic re-measured post-migration.
