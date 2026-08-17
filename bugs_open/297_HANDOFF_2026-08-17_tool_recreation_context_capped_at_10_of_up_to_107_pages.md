# 297 — `tool-recreation-handler` is shown 10 of up to 107 related pages, and the cut is by `nav_order`

> **FIXED AND LIVE 2026-08-17** — migration `453_tool_recreation_whole_site_context.sql`, applied
> and ledger-recorded, by the `bugfix_297_tool_recreation_context` lane. The cap is **gone, with
> nothing bounded in its place**, and a second live defect in the same query (join fan-out) is
> closed with it. **Council APPROVED round 1** (corr `4b9265c3-f6f4-4ed6-a038-f6aaf10b52d8`,
> 14 reviewers, 3 abstained, *"4 advisory objections, none high-severity"*,
> `gated_by_truncation: false`) — see "What the council changed" at the foot.

**Filed 2026-08-17** by the `bugfix_275_silent_row_caps` lane, from the census `bugs_open/275`'s fix
made possible. **Same defect class as 275, different step, and worse in ratio.**

## The defect

`tool-recreation-handler`'s `load_related_context` step (`agent_definitions`,
`{workflow,steps,load_related_context,config,query}`) ends:

```sql
SELECT p.name, p.title, p.page_type, rr.summary
FROM pages p
LEFT JOIN research_results rr ON rr.page_id = p.id AND rr.result_type = 'adoption_page'
WHERE p.site_id = $1 AND p.name != $2
ORDER BY p.nav_order LIMIT 10
```

The rows become the *related context* an LLM is given when recreating a tool. A row-count `LIMIT`
feeding a prompt is **silent by construction**: the model produces a plausible tool whether it saw
10 related pages or 107, so there is no wrong output to notice and no error to catch. This is
`bugs_open/275`'s mechanism exactly — see it, and register **LCO-009**, for the class.

The cut is by `nav_order`, so what survives is "whatever sits highest in the navigation", which is a
judgement about menus, not about what a tool needs to know.

## Measured 2026-08-17 (live `clients_db`)

Population per site, using the query's own predicate (all pages of the site bar the subject):

| | value |
|---|---|
| cap | **10** |
| sites | 24 |
| **sites where the population EXCEEDS the cap** | **19 of 24** |
| median site | **26** |
| worst site | **107** |

At the median site the model sees **10 of 26 (38%)**; at the worst, **10 of 107 (9%)**. For
comparison, `bugs_open/275` — the bug that prompted this census — was 30 of 74 (41%). **This one is
worse, at more sites.**

## It is LIVE, not latent — the agent demonstrably runs

Checked, because a dormant agent would be a different severity:

- `llm_call_log` for `agent_type='tool-recreation-handler'`: **290 calls, most recent 2026-08-11.**
- `site_work_items` with `handler_agent='tool-recreation-handler'`: 4, most recent 2026-08-11.

⚠ `SELECT ... FROM orchestration_states WHERE owner_agent_type='tool-recreation-handler'` returns
**0** and is the WRONG instrument — it does not carry the handler's name. Use `llm_call_log`.

## Fix candidates, ordered by what closes the door

1. **Bound the PAYLOAD, not the row count** — the shape that worked for 275. Measure which column
   dominates (there, `description` was 80% of the bytes), bound that, and drop the row cap. Closes the
   door: the visible set becomes the population. **Do the measurement first** — 275's `LIMIT 30` was
   not arbitrary, and "just remove it" would have tripled that prompt.
2. **Rank before capping.** If a cap must stay, order by something that means "related" — shared
   entities, page_type affinity, link graph distance — rather than `nav_order`. Weaker: still a silent
   cap, but the cut becomes a judgement instead of a menu artefact.
3. **Ask whether the step needs a corpus at all.** `rr.summary` is a research summary per page; if the
   recreation only needs the site's *shape*, a compact aggregate may beat 10 full rows.
4. **Do NOT just raise the number.** 10 → 30 leaves 5 sites still cut and re-creates this file in a
   month, because the population grows and the constant does not. That is precisely how 275 arrived.

## How to verify a fix

- Re-run the census above: **0 sites over the cap**, or no cap at all.
- The disconfirming pair: pick a page sorting past position 10 by `nav_order` on a site with >10
  pages, fire a tool recreation, and confirm it appears in the `plan`/recreate step's rendered prompt
  (`llm_call_log.prompt_rendered`). Before the fix it **cannot** appear; after, it can.
- ⚠ **If you bound a column instead of rows, MARK the truncation.** 275's first fix cut descriptions
  silently and the council caught it — an unmarked cut is the same defect one level down
  (`016b` §9; migration 446 is the worked remedy, `[…truncated]`).

## Filing basis (owner ruling 2026-07-31)

**No `090` run; substitution stated plainly.** This file asserts no new mechanism — the mechanism is
`bugs_open/275`'s, already diagnosed and **council-approved** (corr `b684a399`), and registered as
LCO-009. What is new here is arithmetic on live data: the query text read from `agent_definitions`,
the population counted with that query's own predicate, and the agent's activity confirmed in
`llm_call_log`. Every figure is reproducible by one query. **Grepped `bugs_open/` and `bugs_closed/`
before filing** — nothing covers this step under any spelling.

## What was done (2026-08-17, `bugfix_297_tool_recreation_context`)

Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_297_tool_recreation_context/`.

**Candidate 1 was followed and the measurement inverted its remedy.** The instruction was to
measure which column dominates the payload, bound that, and drop the row cap — 275's shape. Here
**nothing needed bounding**: the prompt renders one short line per row
(`- {{.name}} ({{.page_type}}): {{.title}}`, read from the live template); across all 727 pages
`name` ≤ 66 chars, `title` ≤ 144 (p99 ≈ 114), `page_type` ≤ 16; and the **whole population at the
worst site renders as 8,810 chars (~2.2k tokens) against 735 today**, inside a prompt that already
embeds the original page's entire raw HTML. So the cap simply went. **No truncation marker, because
nothing is truncated** — stating that so the absence of 446's marker machinery reads as a decision,
not an omission. Candidate 4 is honoured: no constant is left to outgrow.

**A second live defect in the same query, found by measuring rather than reading.** The plain
`LEFT JOIN research_results … result_type='adoption_page'` has no one-row guarantee and no unique
constraint behind it. Page `0747e2fc…` (the `index` of site `00ff3af5…`, nav_order 1) **already has
two adoption rows, so that site's prompt was listing `index` twice inside the visible 10** — today,
before any change. Removing the cap widens that door, so the same edit closes it with
`LEFT JOIN LATERAL (… ORDER BY r.created_at DESC NULLS LAST LIMIT 1) ON true`: newest summary per
page, exactly one row per page, row shape unchanged (`name, title, page_type, summary`).
`NULLS LAST` because `created_at` is nullable and a NULL sorts first under plain `DESC` — 0 of 21
rows are NULL today, so the guard costs nothing and makes the bad state unreachable rather than
merely unlikely.

**The inner `LIMIT 1` is not a new silent cap**: it is the fetch-one idiom, which LCO-009 excludes
by design, and LCO-009's regex is end-anchored so a subquery LIMIT is correctly ignored (both arms
vindicated on live cases in 275's round). The migration's post-state verify asserts no **multi-row**
LIMIT survives while permitting this one.

**No Go change**: the framework half of this class is already committed (`eb137faed`, LCO-009) and
rides the next chassis roll. `rr.summary` stays selected though never rendered — 275's `category`
reasoning, and it keeps the row shape byte-compatible.

### Verified live 2026-08-17

| check | result |
|---|---|
| live query text | carries the LATERAL, **no multi-row `LIMIT`** |
| worst site (webdesign.co.uk) | **107 rows = its full population** (was 10) |
| fan-out site `00ff3af5…` | population rows, **duplicate `index` gone** |
| snapshot | `agent_definitions_backup`, 2026-08-17 16:21:26Z |
| ledger | `453_…` recorded via `--record-only` (never a hand-written INSERT) |

**The widening is real, checked at the code not assumed.** The analogue of 275's *"does a
downstream filter drop what you gained"* objection is input-side here: nothing clips the rendered
prompt. In `ExecuteLLMPromptAction` (`platform/orchestration/actions/ai_actions.go:329`) the
rendered template is passed on whole; every `TruncateString` there is a log preview, and the whole
truncation apparatus (`tolerate_truncation`, `__truncated`, `bugs_open/076`) governs the
**response's output tokens**, not the input. Had an input cap existed, this fix would have moved
the silent cap one layer down rather than removing it.

⚠ **Still owed, not claimed:** end-to-end confirmation in `llm_call_log.prompt_rendered` needs the
next real recreation run (most recent call was 2026-08-11). What is asserted here is the
query-level disconfirming pair — a page past nav position 10 **could not** appear before and
**does** now.

## What the council changed (round 1, APPROVED with 4 advisory objections)

Three answers below are stronger than what was submitted, and two of the objections are misses of
mine. Full round in the lane NOTES.

**The prompt-growth question, answered with numbers instead of an OWED note.** `bug_historian`
(medium) argued the fix trades a row cap for unbounded prompt growth *"with no equivalent guard"*,
citing this platform's history of silent LLM-output truncation. Measured:

| check | result |
|---|---|
| `analyze_tool` calls, all history | **129**, peak output **7,735** of an 8,000 cap, **0 truncations** |
| `tolerate_truncation` on any step of this agent | **none** → a truncated response **ERRORS** (`bugs_open/076`), it is never silently persisted |
| `fleet-step-token-pressure` (register **LCO-007**) | **enabled, 6-hourly, last completed 2026-08-17 16:36:39Z** — checked in `scheduled_tasks`, not trusted from the register status line |
| what that standing monitor already says | `N analyze_tool@8000 — n=102, p95 72.8%, peak 96.7%, truncated 0` — **already flagged as a near-miss** |

So the vector is neither silent nor unguarded, and has not fired. **The residual, stated with a
number: 265 tokens of headroom at peak (96.7%).** The trip-wire is LCO-007 reclassifying this step
`N → T`; its runbook holds the response (and says to check prompt SHAPE first — a stuck retry loop
looks identical to genuine cap drift). ⚠ That monitor writes a note only when its flagged SET
changes, so "no recent note" is not "not running" — liveness is `last_completed_at`.

**Blast radius: I asserted it, the guardian was right to ask, and it is now enumerated.** Four live
agents mention `related_pages`; the other three use a *different* field — `input_data.spec.related_pages`,
the cross-link list `tool-suggester` attaches to a suggestion for `tool-generator`/`tool-deployer`.
**No other agent has a `load_related_context` step.** The claim survives, but "the ONLY consumer" had
been a single-agent check stated fleet-wide. ⚠ **Two unrelated fields share one name across four
agents** — a fleet-wide grep for `related_pages` will find all four and suggest a shared field.

**The fan-out is not a class.** `reuse_agent` (low) asked whether other steps join `research_results`
unguarded. Fleet-wide, **exactly one step's query references that table at all** — this one.

**A landmine grep I skipped.** `tooling_provenance` (medium) noted no `doc_notes` check for this
subject before editing. I had grepped the TABLE (`agent_definitions`) and not the SYMBOL
(`tool-recreation-handler`). Six landmine entries name this agent; **none contradicts the change**
(they cover `recreate_tool`/evidence register, `load_page_record`, `expects_no_sections_metadata`,
tool CSS vars, the adoption URL rewrite, and the two instrument traps this file already honoured).
"Nothing blocking" is only knowable after running it.

Rollback (config is live on apply, so it was written before the apply):
`453_tool_recreation_whole_site_context_ROLLBACK.sql`, gated to refuse unless the row still carries
453's text.

## Related

`bugs_open/275` (the sibling, fixed and approved) · register **LCO-009** (the detector — once the
chassis rolls, `QueryDatabaseAction` will WARN when this query fills, which is how this becomes
visible without a census) · `bugs_open/298` (the third instance) · `bugs_open/242` (same class, render
audit) · `016b` §9 (silent caps).
