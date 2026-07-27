# HANDOFF — Gemini content provider, continue here (2026-07-27, evening)

**Supersedes `HANDOFF_2026-07-27_continue_here.md`.** Read this, then `NOTES` for
the evidence and `RUNBOOK` for the commands. Bugs: **107** (the starvation fix),
**110** (telemetry), **121** (voice single-source), **123** (fabrication).

---

## Where we are in one paragraph

Both content agents run on `gemini-pro-latest` in production and are proven on live
generations. The house Voice & Style rules exist in exactly **one** place and are
live on both. The starvation defect that got Gemini reversed in July is fixed,
council-approved, and shipped. What is left is verification on a real page, one
telemetry migration, and a fabrication gap that is not really about Gemini at all.

## State

| thing | state |
|---|---|
| `content-creator-agent` | **LIVE on `gemini-pro-latest`**, image **v1.0.1178**. Proven: 264-char tweet at the 100-token tier that returned ZERO on 07-24; 1,292-word blog post, no truncation |
| `page-content-writer` | **LIVE on `gemini-pro-latest`**. Its own prompt verified against the model (valid unfenced JSON, all keys, `finishReason=STOP`) |
| Chassis | **v1.0.1177** — already carried my code before I was asked to roll (another session built from committed HEAD). Do **not** roll redundantly |
| Voice & Style v4 | **LIVE, single-sourced.** Canonical row `agent_default_configs.config_name='voice_style_block'`, 2,499 chars. page-writer literal **gone**, Go const **gone** |
| `bugs_open/107` | Fixed, council-APPROVED (`a1a5cf20`), shipped. **Real page built and read 07-27 20:15 UTC (`sale`, live) ⇒ the last condition is met — CLOSE IT** |
| `bugs_open/110` | Candidate 1 **LIVE, verified** (Gemini rows log `max_tokens 8000`, not 16192 — this table previously said inert). **Candidate 2 owed** — see "the cost question" below |
| `bugs_open/112` | **FIX CONFIRMED LIVE** in v1.0.1177 (`ProviderKeyEnv` present in the binary), so spawned pods now get `GEMINI_API_KEY`. **The writer can reach Gemini.** Someone should close it |
| `bugs_open/121` | Fixed. One copy of the voice. Migrations 240 + 241 both applied |
| `bugs_open/123` | **OPEN, unowned, needs an owner call** — see below |

## Do this next, in this order

### 1. ~~Re-queue the page build~~ — DONE 2026-07-27 20:15 UTC, but NOT this way

> **CORRECTED 2026-07-27 evening — the diagnosis below is wrong, and the retry it
> recommends could never have worked.** `grip-styles` did not fail because the writer
> could not reach Gemini. Read `site_work_items.error`, not the timing:
> `"page-build-handler no-op: no sections ready to build … the target section was NOT
> rebuilt"`. It never reached an LLM call of any provider. **`grip-styles` has no rows
> in `site_plan_sections` — and neither does any other `blog-post` page on
> dartsonline**, including all three "safe sibling" fallbacks named below. Two of them
> (`tungsten-guide`, `beginners`) already tried and are parked in `needs_human_review`
> from 07-20/07-22, with three more. Re-queueing would have added a sixth.
>
> **What was done instead:** built `sale` (the only never-deployed page WITH plan
> sections and no competing work item), owner-approved to publish. **Live at
> `https://dartsonline.com/sale.html`** — copy clean on every check, and **zero
> fabricated statistics on a sale page**. First-ever Gemini rows in `llm_call_log`.
> Full record in NOTES under "P7 done". Two defects found, neither Gemini's: the hero
> and CTA duplicate each other, and `product-grid` was skipped for missing data so the
> Sale page has nothing to buy.
>
> **Also corrected by that run: `110` candidate 1 is LIVE, not inert.** The Gemini
> rows log `max_tokens = 8000`, not the reserve-inflated 16192 — which is RUNBOOK §5's
> own test for a post-fix binary. The state table below and RUNBOOK §5 both say inert.

*Original text, left for the record:*

`dartsonline/grip-styles` was queued at 15:14 and is now `needs_human_review`,
`attempt_count 1`, last touched **15:46** — i.e. it *did* get dispatched and failed,
**before** the v1.0.1177 roll and **before** `112`'s fix was live. Its failure is
therefore stale evidence: the writer could not reach Gemini at all when it ran.

```sql
-- inspect first; do not assume the old failure reason still applies
SELECT status, attempt_count, updated_at, spec FROM site_work_items
WHERE created_by = 'gemini-p7-verification';
```

Re-queue it (or a sibling `planned` blog-post page on dartsonline — `tungsten-guide`,
`steel-tip-vs-soft-tip`, `beginners` are all `planned` and never deployed, so a bad
result costs nothing), then **read the copy**. `complete` is not proof; read the
artefact. Check: em dashes 0, filler 0, no negative-frame openings, at least one
"why it matters" sentence, and that the page's own story survived.

Watch for a `*TruncatedError` naming thinking — that means the 8192 reserve is too
small for the writer's real context-loaded prompt. It is **not** a sign the fix
failed; raise `thinking_reserve_tokens` in that step's `ai_service`.

**Builds may still be blocked by `bugs_open/029`** (hung spawns saturate the dispatch
group). The owner has a thread on it. If the item sits at `triaged` with
`attempt_count 0` and never moves, that is 029, not this workstream.

### 2. The cost question you cannot currently answer

The owner asked to keep an eye on costs. **You cannot, from a query.** `110`
candidate 2 is unbuilt: `__usage_thinking_tokens`, `__usage_total_tokens` and
`__sent_wire_max_output_tokens` are written by the client and read by **nothing** —
`llm_call_log` has no columns for them. Thinking is where essentially all of
Gemini's cost lives, so the one number that matters is invisible.

That is a migration plus a small Go change (add the columns, read the four `__`
fields in `ai_actions.go`, pass them through `llm_call_logger.go`). **This is the
highest-value unbuilt item in the workstream**, because every cost decision from here
depends on it.

Measured figures to sanity-check against once it lands: thinking on the real
12,570-char writer prompt was **1,576–2,878 tokens**; a hero section came to ~2,425
billable output tokens on Gemini against 767 on Fable and 172 on Sonnet.

### 3. `bugs_open/123` — an owner call, not a code task

A live content-creator generation wrote *"Industry data shows that large language
models experience hallucination rates between 3% and 10%"* — invented, uncited,
phrased to read as sourced. The claims assessor **cannot** be pointed at this
output: it needs a `SiteID` and loads the evidence base from `site_specs`, and
content-creator's text has no site, no page, no evidence base.

**The urgency question, partly answered:** nothing subscribes to
`system.agent.content-creator.responses` as a dedicated consumer, but the generic
orchestration awaited-response machinery reads it (`platform/orchestration/helpers.go`
— `awaited.ResponsesTopic`). So the text **does** flow back into orchestration state,
and what happens next is workflow-defined. **I did not trace which workflows call
content-creator.** That trace is the remaining cheap step and it decides whether 123
is urgent or merely important.

Do **not** "fix" this by adding "never invent statistics" to the voice block. `043`
exists because prompt rules were not enough, and the voice block governs how copy
reads, not whether it is true.

## Landmines, ordered by what they would cost you

1. **Assert POSITION, not just presence.** My v4 guard checked the new rules were
   *in* the 16,150-char template and passed while they sat 11,500 chars from the
   block they belonged to. Only a *later* migration refusing itself surfaced it. On
   any patch to a big text blob, assert where.
2. **`jsonb_set` with a literal object is a REPLACE.** The writer's `max_tokens: 8000`
   lives inside the `ai_service` block; a wholesale replace drops it and the client
   falls back to 2048 — a 4x cut, invisible in the diff. Use `||`.
3. **The prompt renderer is `missingkey=zero`.** An unresolved `{{.voice_style}}`
   renders as *nothing*, silently. That is why 241 was gated on a pod-grep. If page
   copy ever loses the house voice with no error, check this first.
4. **Neither Gemini thinking knob CAPS thinking.** `thinkingBudget` is a soft target
   the model overshoots (128 → 483 spent; 32768 → 783). It is a cost lever, never a
   substitute for the reserve. Only `thinkingBudget: 0` is refused.
5. **The writer's step is NOT top-level.** It is at `workflow → steps →
   process_sections_loop → config → sub_workflow → steps → generate_content`. The
   shorter path returns **NULL with no error**. `steps` is an object: `jsonb_each`.
6. **A pod-grep proves the binary, never the path.** For the voice injection the
   proof was a log line showing `voice_style` in live template data, not the symbol
   in the binary.
7. **`fuel_budget` is a required Kafka header** on content-creator requests and is not
   in the payload schema. Without it: `'fuel_budget' header not found`.
8. **Only `page-content-writer` and `content-creator` matter.** The other fifteen
   `content-creator-*` agent definitions had **zero** LLM calls in 30 days. Do not
   paste anything into them.

## Numbers not to re-derive

Measured 2026-07-27 against the live key, `gemini-pro-latest` (→ `gemini-3.1-pro-preview`).

- Four-model bake-off, 5 runs each, identical prompt/material — data:
  `DATA_2026-07-27b_four_model_comparison_5x4_runs.json`, script `RUN_model_comparison.py`.
- **It is mostly the prompt.** Filler and em dashes were 0 for all four models; the
  block was doing that work, not the model.
- **Grok (`grok-4-1-fast`) is the dryness complaint amplified** — 4.8 words/sentence,
  zero contractions, 262 chars. Ruled out.
- **Fable 5 writes best** (0 negative frames vs Sonnet's 0.4) at ~**$0.073**/section
  vs Gemini **$0.024** and Sonnet **$0.010**.
- **Token counts and cost run opposite ways**: Gemini spends 3x Fable's billable
  output tokens and costs a third as much.
- **v4 vs v3 on Gemini**: chars 422 → 637, mean words/sentence 7.6 → 12.1, joined
  clauses ~0 → 2.2, em dashes still 0.
- `gemini-2.5-pro` / `-flash` still **404** for this key; the model listing advertises
  them anyway.

## Open elsewhere, touching this work

- **`bugs_open/029`** — hung spawns halt builds fleet-wide. Owner has a thread. This
  is what gates a real page build.
- **`features_open/025`** — a provider-independent character cap for the twitter tier
  (item (a) only; item (b) was superseded by `110`).
- **`bugs_open/121`'s coda** — the architecture seat has **0 reviews, ever**, and a
  twenty-line pre-commit hook made the architectural observation it exists to make.
  Offered there as a first case it could review.
