# HANDOFF — Gemini content provider, continue here (2026-07-27)

**Read this first, then `PLAN` for the why and `RUNBOOK` for the commands.**
Bug case: `bugs_open/107`. Follow-ups: `features_open/025`.

---

## One paragraph

The content agents were moved to Gemini on 2026-07-23/24 and reversed, on a
verdict that Gemini pro "can't write" — zero characters at the twitter tier. That
verdict measured **our own starved token budget**: Gemini's `maxOutputTokens` is a
*total* output ceiling with thinking spent from it first, while Anthropic's
`max_tokens` (thinking off) is all visible text, and our client passed the
Anthropic-sized number straight through. The client is fixed, council-APPROVED,
shipped in **v1.0.1173**, and **content-creator is live on `gemini-pro-latest` and
proven on two real generations**. One thing is left: flipping
`page-content-writer`, which needs a live DB write my tooling refused.

## State

| phase | state |
|---|---|
| P1 client fix | **DONE** — `8a2b5dea0`, `f4f2336a3`, `17136ce3c`, `b08387fde` |
| P2 council gate | **DONE — APPROVED round 1**, corr `a1a5cf20-a70d-48c3-8fda-842d2a91b651` |
| P3 build + roll | **DONE** — both images `v1.0.1173`, pod-grep verified on both binaries |
| P4 probe the key | **DONE** — tier tables in `NOTES`; falsified one of my own claims |
| P5 flip content-creator | **DONE AND PROVEN** — 264-char tweet at the tier that returned ZERO on 07-24; 1,292-word blog post, no truncation |
| **P6 flip page-content-writer** | **NOT DONE — the only outstanding item** |
| P7 read the copy | **NOT DONE** — the real test of the question the owner asked |

`bugs_open/107` stays **OPEN** until P6/P7 land, per the "fixed AND live with no
residual" bar.

## The one command left

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db \
  < docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/P6_FLIP_page_content_writer.sql
```

My tool-permission classifier refused this write; nothing else is blocked. The
script is transactional, backs the row up, is guarded on the `updated_at` I read
(`2026-07-27 13:44:56.343485+00`), and **RAISEs to roll itself back** unless
provider, model, `max_tokens: 8000` and the Voice & Style block all verify after
the write.

If it reports `UPDATE 0` / raises on provider: another session wrote the row.
**Re-read `updated_at`, put the new value in the script, do not retry blind.**

### Then, and do not skip it

Rebuild **one** page and **read the copy**. `complete` is not proof the work
happened — read the artefact. Compare against the Claude baseline the brochure
workstream snapshotted (`about_copy_before.txt`): em dashes, filler words,
fact-first openings, and whether the page's own story survived. Only then consider
anything site-wide.

Watch for a `*TruncatedError` naming thinking. That means the 8192 reserve is too
small for the writer's *real* prompt — raise `thinking_reserve_tokens` in the same
`ai_service` block. It is not a sign the fix failed.

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

- Real 12,570-char writer prompt: thinking **2,764–2,878** tokens. Hence 8192
  (~3x headroom).
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
