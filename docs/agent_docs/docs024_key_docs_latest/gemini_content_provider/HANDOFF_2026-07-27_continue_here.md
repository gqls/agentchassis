# HANDOFF — Gemini content provider, continue here (2026-07-27)

**Read this first, then `PLAN` for the why and `RUNBOOK` for the commands.**
Bug cases: **`bugs_open/107`** (the starvation fix) and **`bugs_open/110`** (a
regression the 107 fix introduced into `llm_call_log` — candidate 1 applied, inert
until the next roll). Follow-up: `features_open/025` (item (a) only; item (b) was
superseded by 110).

---

## One paragraph

The content agents were moved to Gemini on 2026-07-23/24 and reversed, on a
verdict that Gemini pro "can't write" — zero characters at the twitter tier. That
verdict measured **our own starved token budget**: Gemini's `maxOutputTokens` is a
*total* output ceiling with thinking spent from it first, while Anthropic's
`max_tokens` (thinking off) is all visible text, and our client passed the
Anthropic-sized number straight through. The client is fixed, council-APPROVED,
shipped in **v1.0.1173**, and **BOTH agents are now live on `gemini-pro-latest`** —
content-creator proven on two real generations, the writer proven on its real prompt
returning valid JSON. One thing is left, and it needs a decision rather than a
command: **rebuild one real page and read the copy**, on a target the owner
nominates.

## State

| phase | state |
|---|---|
| P1 client fix | **DONE** — `8a2b5dea0`, `f4f2336a3`, `17136ce3c`, `b08387fde` |
| P2 council gate | **DONE — APPROVED round 1**, corr `a1a5cf20-a70d-48c3-8fda-842d2a91b651` |
| P3 build + roll | **DONE** — both images `v1.0.1173`, pod-grep verified on both binaries |
| P4 probe the key | **DONE** — tier tables in `NOTES`; falsified one of my own claims |
| P5 flip content-creator | **DONE AND PROVEN** — 264-char tweet at the tier that returned ZERO on 07-24; 1,292-word blog post, no truncation |
| P6 flip page-content-writer | **DONE** — script ran clean, `max_tokens: 8000` preserved (what a `jsonb_set` replace would have deleted), style block intact |
| **P7 read the copy** | **PARTIAL — needs an owner decision on the target.** The writer's real 12.5K prompt at its real 8000 budget returns **valid unfenced JSON, all keys, finish=STOP, 1,576 thinking tokens**, and the copy scores **0 em dashes / 0 filler / fact-first opening** (Claude+v1 got em dashes 19→14). But **no real page has been rebuilt** — see below |
| **110 candidate 1** | in code, **INERT until the next chassis roll** |

`bugs_open/107` stays **OPEN** until P6/P7 land, per the "fixed AND live with no
residual" bar.

> # CORRECTED 2026-07-27 (triage sweep, after the v1.0.1174 roll) — READ THIS BEFORE THE REST
>
> **Two things in this handoff below are wrong, and one big thing is missing.**
>
> 1. **`bugs_open/029` is NOT the blocker, and the "halted every build since 19
>    July" claim is false.** `build-dispatch-loop` completed 62 orchestrations on
>    07-26 and 30 on 07-27. Page builds completed all through 07-26/27. The
>    `grip-styles` item **was claimed and did run**, at 15:46 UTC.
> 2. **It failed on a bad target, not a broken pipeline.** `load_spec_sections`
>    returned `{"count": 0, "source": "none"}`: dartsonline has no `site_plan`
>    aspect in `site_specs`, and `pages.sections` for `grip-styles` is `[]`. Zero
>    sections, so the handler routed to `mark_no_ready_sections` and set the work
>    item to `needs_human_review`. (Not `bugs_open/087` either — that is the
>    `page-rebuild` path; this was `page-build-handler`.)
> 3. **THE ACTUAL BLOCKER IS NEW: `bugs_open/112`.** Spawned agent pods are never
>    given `GEMINI_API_KEY`. `page-content-writer` runs in a spawned pod, and
>    `spawn_actions.go:2440-2518` builds their env as an explicit allow-list with
>    `ANTHROPIC_API_KEY` and `GROK_API_KEY` and no Gemini key.
>    `content-creator-agent` is a standalone Deployment with its own key patch —
>    **that is the whole reason P5 passed and P6 cannot.**
>
> **P6/P7 are therefore live-but-armed-to-fail, and this is time-sensitive.** The
> flip is in the DB, `generate_content` has **no `error_step`**, and the scheduled
> `model-directory-publish` fires every 6h (last 14:25 UTC, **next ~20:25 UTC**).
> When it fires, every section fails at client construction and the page build
> fails with it. Either fix `112` or revert the writer to anthropic first.
>
> **What IS confirmed live and verified** (running pods, v1.0.1174, not git): the
> 107 client fix (positive grep 1, negative control 0), the 110 candidate-1 rename
> (`__sent_wire_max_output_tokens` 1, `__sent_visible_budget_tokens` 0), and the P6
> row itself with `max_tokens: 8000` intact and no shadowing `ai_service`.

## BLOCKED, and not by us: `bugs_open/029` has halted every build since 19 July

> **The heading and section below are superseded — see the correction above.**

**P7's page build cannot run.** Queued `dartsonline/grip-styles` (planned, never
deployed) as a proper `needs_page` build item on 2026-07-27. It was never claimed.
`build-pipeline-trigger` detects it correctly (`pending_sites: "1"`, and it is the
only eligible build item fleet-wide), then hangs at `spawn_dispatch` awaiting a
response. **`build-dispatch-loop`'s last orchestration is CANCELLED, 2026-07-19.**

That is `bugs_open/029` — *hung spawns saturate the dispatch group and silently halt
ALL builds fleet-wide*, filed 19 July, **still OPEN**, the consequence half of `003`.
Nothing to do with the model swap. **Do not start a competing fix**; it is filed and
owned. The work item stays queued and will build when 029 lifts.

Two traps if you retry this:
- `build-pipeline-trigger`'s `seed_queue` step reporting `{"total":0,"seeded":0}` is
  a **red herring**. That step seeds brand-new sites from `build_queue`; zero is
  correct. The stall is one step later, at `spawn_dispatch`.
- `grip-styles` has **0 `page_components`**, so a rerender has nothing to re-render.
  A `planned` page needs the full build path. See `bugs_open/087`.

## The one thing left, and it needs YOUR call, not a command

**Rebuild one real page through `process_sections_loop` and read the copy.** I
stopped short of this deliberately: every candidate page belongs to a workstream
that is mid-flight, and mutating their live estate to answer my question is not my
decision.

- `fundamentallyai` / `about` — **the ideal comparison**, because the brochure
  thread snapshotted `about_copy_before.txt` as a Claude baseline for exactly this.
  But that thread is actively working `085`/`109` on that site.
- `vonc.com` — the designated test site, but the gauntlet thread has just got it to
  `claimscan 0/49`, and regenerating copy could reintroduce fabrications.
- Pool sites — **have no pages at all** (`pool-ai-agents.internal` → 0 rows), so the
  usual scratch target does not exist for a page build.

So: **name a page, or authorise `fundamentallyai/about` after a word with the
brochure thread.** Then read the artefact — `complete` is not proof the work
happened — and check em dashes, filler, fact-first openings, and that the page's own
story survived (bug-056 vigilance).

Watch for a `*TruncatedError` naming thinking: it means the 8192 reserve is too
small for the writer's *real* (context-loaded) prompt — raise
`thinking_reserve_tokens` in the same `ai_service` block. It is **not** a sign the
fix failed. On my harness thinking was 1,576, i.e. the reserve is ~5x what was
needed, but a real run carries `site_specs`, `brief`, `existing_content` and
`link_context` on top.

### What P7 already established, and what it does not

**Does:** the writer's prompt shape works on Gemini; it returns **valid, unfenced
JSON with all required keys** (the real risk, since `output_format: json` means a
fence or a mid-string cut would break the section writer invisibly); no truncation
at 8000; and **the Voice & Style block transfers to a different model family**
rather than being tuned to Claude — an untested risk until now.

**Does not:** justify "Gemini writes better than Claude". It is **n=1 vs n=1** on
different content, and my harness had no `site_specs`/`brief`/`existing_content`/
`link_context`, no chassis `appendOutputInstructions`, and no multi-section
coherence test.

## Landmines, ordered by what they would cost you

1. **`jsonb_set` with a literal object is a REPLACE, not a merge.** The writer's
   `max_tokens: 8000` lives *inside* the `ai_service` block being patched. My first
   version of this SQL would have dropped it, falling back to the client's 2048 —
   a 4x cut, invisible in the diff, surfacing days later as truncated sections and
   sending you to the reserve for a cause that isn't there. Use `||`. The script
   does, and asserts the 8000 afterwards.
2. **The writer's step is NOT top-level.** It is at
   `workflow → steps → process_sections_loop → config → sub_workflow → steps →
   generate_content → config → ai_service`. The shorter path returns **NULL with no
   error**, so a wrong path is indistinguishable from an absent value. `steps` is an
   *object*: `jsonb_each`, not `jsonb_array_elements`.
3. **Neither thinking knob CAPS thinking.** `thinkingBudget` is a soft target the
   model overshoots (128 → 483 spent; 32768 → 783). It is a **cost lever, not a
   correctness one** — never treat one as a substitute for the reserve. Only
   `thinkingBudget: 0` is refused; both knobs together are refused by Google.
4. **The 07-24 "Gemini writes badly" verdict is not evidence of anything.** Every
   measurement was of a starved budget, and `page-content-writer` was never
   exercised on Gemini at all (flipped 16:53, reverted 16:59, its test rebuild
   still queued). Do not cite it either way.
5. **A negative control must be absent *because of* your change.** I grepped
   `datahelpers.GetIntField` expecting 0, got 1, and it proved nothing — the symbol
   exists all over the tree. The valid control was the old format string
   `no text content in response (finishReason=%q)` → 0.
6. **`fuel_budget` is a required Kafka header** on content-creator requests and is
   not in the payload schema. Without it you get `'fuel_budget' header not found`
   on the errors topic. Add `-H fuel_budget=100000`.
7. **No `Council-Reviewed:` trailer exists for this work, and that is correct.**
   The verdict post-dates the commits and the tree is forward-only, so `098` will
   list them UNREVIEWED — a known false negative (`016b` §8.2, the
   `bugs_closed/011` shape). Adding one via a later commit would attach the
   approval to code the council never saw.

## Numbers you should not re-derive

Measured against the live key, 2026-07-27, `gemini-pro-latest`
(→ `gemini-3.1-pro-preview`). Re-measure with `scripts/gemini-probe.sh --from-pod`
if prompts grow materially — thinking scales with prompt complexity, it is not a
constant.

- Real 12,570-char writer prompt: thinking **1,576** tokens with genuine material,
  **2,764–2,878** with stub placeholders. Hence 8192 (3–5x headroom).
- Trivial prompt: thinking 786–1,145.
- **Thinking expands to fill a small ceiling**: at `maxOutputTokens=100` it spent
  92 and left 4 tokens of text. That is the 07-24 failure, reproduced exactly.
- `thinkingLevel: "low"` → ~1,080; `thinkingBudget: 512` → ~940. Both roughly a
  third of default, if cost ever needs trimming.
- `gemini-2.5-pro` / `-flash` still **404** for this key. `gemini-3-pro-preview` is
  retired outright. **The models listing advertises all three** — the listing is
  not evidence of reachability.
- Live proof: tweet 264 chars / 66 tokens / 12.6s at the 100-token tier; blog post
  8,726 chars / 1,292 words / 2,181 tokens / 35.4s at the 6000 tier, no truncation.
  Cost metadata resolves at the Gemini rate (0.00066 for 66 tokens).

## Open items that are not P6

- **`features_open/025`** — two council recommendations, both real and neither
  mine to fix inside this workstream: a provider-independent **character** cap for
  content-creator's twitter tier (because `max_tokens` no longer bounds visible
  length for a thinking model, and a clamp in the provider client would truncate
  JSON mid-string), and teaching `llm_call_log`'s
  `output_tokens == max_tokens` CUT heuristic about the reserve split, since it now
  silently under-reports for Gemini.
- **Cost at scale is unmeasured.** Thinking bills as output. `page-content-writer`
  runs per section across a whole site, and its prompt is 4x the size of the one I
  measured. `__usage_thinking_tokens` is now recorded per call — read it after the
  first real page rather than projecting from my probe.
- **`text-embedding-004`** is `[UNVERIFIED]` as still reachable — same retirement
  class as the 2.5 pins. Made configurable rather than changed; nothing here is
  known to call `GenerateEmbedding` on Gemini.
- **The `[UNVERIFIED]` Gemini rates** in content-creator's `estimateCost` are
  carried over from the 2.5 generation and have never been checked against
  Google's price list. The live `0.00066` figure confirms the *key* now matches,
  not that the *rate* is right.
