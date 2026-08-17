# NOTES — bugs_open/297, tool-recreation-handler's related-context cap

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-17 — picked up, ownership checked two ways

`who-owns.py 297` → only the filing commit (275 lane spinning the census out); no owning
workstream. Live-transcript grep for the slug + `tool-recreation-handler` across ~25 recent
`.jsonl` sessions: the only real hit counts are the 275 lane itself (last message: "both filed…
nothing of mine uncommitted") and the 291 lane (tool-auditor). No competing fixer.

## 2026-08-17 — validity re-check, from the live DB not the bug file

- One active `tool-recreation-handler` row (id `8701375f…`, v1) — not a duplicate-active-row type.
- Query unchanged: ends `ORDER BY p.nav_order LIMIT 10`.
- Census: 25 sites (was 24 at filing), 19 over cap, median 26, worst 107 (webdesign.co.uk).
  Verbatim live query against the worst site returns exactly 10 rows today.

## 2026-08-17 — the measurement flipped the remedy away from 275's shape

Expected to repeat 275 (bound the dominant column, mark the truncation). Measured instead:

- Rendered line is `- name (page_type): title`; `rr.summary` is selected and NEVER rendered —
  and is nearly absent anyway (21 of 727 pages, max 48 chars).
- Column extremes: name ≤ 66, title ≤ 144, type ≤ 16. Nothing to bound.
- Whole-population rendered block at the worst site: **8,810 chars (~2.2k tokens)** vs 735 capped —
  in a prompt already carrying the page's full raw HTML.

So the honest fix is simply: **no cap, no bounding, no marker** — there is nothing to truncate.
Wrote that down BEFORE the council could ask why I didn't cargo-cult `left(…,200)` from 445.

## 2026-08-17 — found a second live defect in the query: join fan-out

The plain `LEFT JOIN research_results … result_type='adoption_page'` has no one-row guarantee.
Measured: page `0747e2fc…` (`index` of site `00ff3af5…`) has 2 adoption rows and sits at
nav_order 1 — **today's prompt on that site lists `index` twice inside the visible 10.** With the
cap gone this door opens wider (N research rows → N lines), so the same edit closes it:
`LEFT JOIN LATERAL (… ORDER BY r.created_at DESC LIMIT 1)` — newest per page, shape-preserving,
indexed (`idx_research_page`, `idx_research_created`).

Read-only validation of the proposed text: worst site → 107 rows = population; fan-out site →
40 rows = population, duplicate gone.

## 2026-08-17 — scope decisions, stated so the reviewers don't have to ask

- No Go change: LCO-009's detector is committed (`eb137faed`) and covers the class at the shared
  point; it rides the next roll.
- `rr.summary` kept in the SELECT although unrendered — 275's own `category` reasoning (dropping
  a column a future consumer might read = scope creep for negligible saving). Adjacent finding.
- Prompt template untouched.
- Inner `LIMIT 1` is the fetch-one idiom — outside the silent-cap class by LCO-009's stated
  design (n=1 excluded; end-anchored regex ignores mid-query subquery LIMITs — both arms
  vindicated on live cases in 275's round 2).

## 2026-08-17 — council submitted (FORCED, with the reason stated), then applied

`SUBMISSION_CORR = 4b9265c3-f6f4-4ed6-a038-f6aaf10b52d8`.

**The gate's client-side scope filter is `^(platform|internal|pkg)/` and this change is
config-only**, so it needed `FORCE=1`. I did NOT do that silently: the rationale opens with the
scope note. The justification is 275's own round — the sharpest objection there
(`bug_historian`, unmarked column truncation) was against the MIGRATION half, not the Go half, so
a config migration to a live shared agent is exactly a shape this gate has caught things in. The
ruling's purpose is that *docs and site content* never spend credits; a live agent's SQL is
neither.

## 2026-08-17 — a risk I raised, then closed rather than leaving for a reviewer

Writing the submission's risks block I noted the LATERAL's `ORDER BY r.created_at DESC` and asked
whether it wanted `NULLS LAST` — `created_at` is nullable, and a NULL sorts FIRST under plain
DESC, which would let an untimestamped row win the "newest" tie for ever.

**I measured instead of asking: 0 of 21 adoption_page rows are NULL today.** So the guard costs
nothing right now, and it makes the bad state unreachable rather than merely unlikely — which is
the "rank by what closes the door" test. Added `NULLS LAST` to the query, the verify block's
literal, the file header and the submission before firing it. **A risk you can close for free is
not a risk to hand a reviewer.**

## 2026-08-17 — applied, verified, recorded

Applied by hand (own file only — `--apply` takes every pending file, and 452 sits in the dir as a
`_HOLD`). Snapshot captured, both gates passed, `UPDATE 1`, post-state verify passed, COMMIT.

| check | result |
|---|---|
| live query | LATERAL present, **no multi-row LIMIT** |
| worst site | **107 rows = full population** (was 10) |
| fan-out site | population rows, duplicate `index` gone |
| snapshot | `agent_definitions_backup` 16:21:26Z (NOT an `is_snapshot` row — the landmine) |
| ledger | `--record-only` with a note, never a hand INSERT |

**A live reminder of the bug's own premise, caught in passing:** the fan-out site's population read
41 while I was validating and **42** ten minutes later, at verify time — another session was adding
pages underneath me. A fixed constant against a growing population is exactly what candidate 4
warned about; this fix leaves no constant.

**Owed, not claimed:** `llm_call_log.prompt_rendered` confirmation needs the next real recreation
run (most recent call 2026-08-11). The query-level disconfirming pair is what is asserted today.

## 2026-08-17 — checked the objection 275 drew, before a seat could ask it

275's `editquality` seat asked whether a downstream filter silently drops what a widened prompt
gains (the *"widening a planner's MENU changes nothing"* landmine). The analogue here is
input-side: **does anything clip the rendered prompt, so the extra rows never reach the model?**

**Read the code, not the config: no.** In `ExecuteLLMPromptAction`
(`platform/orchestration/actions/ai_actions.go:329`) the template is rendered and the result is
passed on whole. Every `TruncateString` call in that file is a **log preview** (350/300/400 chars);
the entire truncation apparatus — `tolerate_truncation`, `__truncated`, `bugs_open/076`'s refusal —
is about the **response's output tokens**, not the input. There is no input-side character or token
cap anywhere on the path.

So the widening is real: more rows in the query means more lines in the prompt the model actually
sees. Worth having checked rather than assumed — if an input cap HAD existed, this fix would have
moved a silent row cap one layer down instead of removing it, which is exactly 275's misstep-4
shape.

## 2026-08-17 — council ROUND 1: **APPROVED**, 14 reviewers, 3 abstained, 4 advisory objections

`decision: approved`, *"approved with 4 advisory objection(s) — none high-severity"*,
`gated_by_truncation: false`, `unreadable: 0`. Corr `4b9265c3-f6f4-4ed6-a038-f6aaf10b52d8`.

⚠ **A method note on reading the verdict.** At 16:35 the `council_report` already said `approved`
while `orchestration_states` still read `review_guardian EXECUTING_STEP`. I did **not** claim the
verdict then — `review_guardian -> council_decide` in the step chain, so a guardian veto is still
reachable at that point. The run went terminal at `complete_approved` (COMPLETED) a minute later,
and that is what I read. **An artifact is not a verdict until the run that can override it has
finished.**

Every objection below was answered with a measurement, not an argument. Three of them changed what
I know; two are misses of mine.

### bug_historian (MEDIUM) — "you traded a row cap for unbounded prompt growth, with no guard"

The sharpest objection, and the one I had left as an OWED item — which is exactly what it called
out: *"the author's own 'OWED' framing quietly stands in for a fix."* Fair. It named the check:
has `analyze_tool` ever truncated?

**Measured, and the objection's premise does not hold — there IS a guard, and it already watches
this step:**

| check | result |
|---|---|
| `analyze_tool` calls (all history) | **129**, max output **7,735** of an 8,000 cap, **0 truncations** |
| any step of this agent setting `tolerate_truncation` | **none** → a truncated response **ERRORS** the step (`bugs_open/076`'s machinery), it is not silently persisted |
| `fleet-step-token-pressure` (LCO-007) | **enabled, 6-hourly, last completed 2026-08-17 16:36:39Z** — verified in `scheduled_tasks`, not trusted from the register |
| what that monitor already says | `N analyze_tool@8000 — n=102, p95 72.8%, peak 96.7%, truncated 0` — **already on its flagged list as a near-miss** |

So the "silent-loss vector with no equivalent guard" is: (a) not silent — truncation errors loudly
here; (b) not unguarded — a standing 6-hourly fleet check already classifies this exact step and
would reclassify it `N → T` the moment it truncates; (c) not yet realised — zero truncations in 129
calls. **The residual is real and now stated with a number instead of a shrug: peak output is 96.7%
of cap, 265 tokens of headroom.** The trip-wire is LCO-007's own note, and the follow-up if it ever
flips to `T` is in that note's runbook (raise the cap or shrink the unit — and check the prompt-shape
query first, because a stuck retry loop looks the same).

⚠ **The monitor's newest note is dated 2026-08-15 and that is BY DESIGN** — it writes only when the
flagged SET changes (md5 digest, 30-day dedup). "No recent note" is not "not running"; the liveness
check is `scheduled_tasks.last_completed_at`, which is today.

### guardian (MEDIUM) — "blast radius is ASSERTED, not checked against the fleet" — **MISSTEP 1, mine**

Correct, and it is CLAUDE.md's own rule (*"enumerate the consumers — asserting it without the query
is itself the objection"*). I verified the consumer **within** this agent's config and wrote "the
ONLY consumer", which is a fleet-wide claim from a single-agent check.

**Enumerated now:** four live agents mention `related_pages` — and the other three are a **different
field**: `input_data.spec.related_pages`, the 1-3 page-name cross-link list `tool-suggester` attaches
to a suggestion and `tool-generator`/`tool-deployer` consume. **No other agent has a
`load_related_context` step**, and my step's `related_pages` lives only in its own orchestration's
collected data. The claim survives — but it is now enumerated, and it was luck that it did.

**Worth its own line:** two unrelated fields share one name across four agents. A future session
grepping `related_pages` fleet-wide will find all four and can easily conclude the field is shared.

### guardian (MEDIUM) — "453 is claimed by filename, not the ledger"

True when submitted (review is after the fact by design); resolved since: `--record-only` recorded
it with a note. Ledger max was 451, and 452 sits in the dir as a `_HOLD` sidecar.

### tooling_provenance (MEDIUM) — "no check of doc_notes for this subject before editing" — **MISSTEP 2, mine**

Also correct, and it is the standing memory rule *"grep LANDMINES for the SYMBOL you are about to
trust"* — the SessionStart hook only matches files already dirty, so an agent-type footprint is
never shown. I grepped `agent_definitions` (the table) and **not `tool-recreation-handler` (the
symbol)**.

**Grepped now: six landmine entries name this agent. None contradicts the change** — they cover
`recreate_tool` + the evidence register (untouched), `load_page_record` (untouched),
`expects_no_sections_metadata` (untouched), tool CSS vars (untouched), the adoption URL rewrite
(background), and the two instrument traps I had already honoured independently
(`owner_agent_type` returns a confident 0; `internal-linker` vs `internal-link-resolver`).
**"Nothing blocking" is only knowable after running it**, and the cost was three seconds.

### reuse_agent (LOW) — "is the unguarded fan-out join a class, patched once?"

The right question, and one query answers it: **fleet-wide, exactly ONE step's query config
references `research_results` at all** — this one. Not a class; nothing else to sweep.

### editquality / prior_art_librarian (LOW) — "we cannot see `research_results` from the council's schema tier"

Both flagged the duplicate-row and nullable-`created_at` claims as unverifiable *by them* and asked a
human to confirm. Confirmed here, with the queries in the RUNBOOK: 21 `adoption_page` rows, 0 with
NULL `created_at`, and page `0747e2fc…` carries exactly 2 — which is the duplicate.

### The one thing I checked that nobody asked for, and would have mattered most if it had failed

Whether anything clips the rendered prompt input (the previous NOTES entry): it does not. Had an
input cap existed, this fix would have moved the silent cap one layer down.
