# 107 — The Gemini client starves thinking models of output budget, and the starvation was diagnosed as a model that cannot write

**Filed** 2026-07-27 · **Status** FIXED IN CODE, **council-APPROVED**, **INERT
until the next image roll** — so it stays OPEN · **Owner**
`gemini_content_provider` workstream ·
**Council** APPROVED round 1, corr `a1a5cf20-a70d-48c3-8fda-842d2a91b651`
(10 reviewers, 6 filtered, `unreadable: 0`; 4 advisory objections, none
high-severity). **No `Council-Reviewed:` trailer is possible** — the verdict
post-dates the commits and the repo is forward-only, so `098` will read
UNREVIEWED: a known false negative, same shape as `bugs_closed/011` round 9
(`016b` §8.2). Two objections became `features_open/025`; the rest are answered in
the workstream NOTES ·
**Related** `bugs_closed/008` (stop signals undecoded), `bugs_closed/009`
(root `ai_service` shadowing), `bugs_closed/011` (a capability believed missing
because one enum routed elsewhere)

---

## Symptom

Pointing the content-producing agents at Gemini produced almost no text:

| tier (`max_tokens`) | visible output | `finishReason` |
|---|---|---|
| 100 (twitter) | **zero characters** | `MAX_TOKENS` |
| 500 (short) | **~85 characters** | `MAX_TOKENS` |

Measured against the live API on 2026-07-24 with `gemini-pro-latest` (then
resolving to `gemini-3.1-pro-preview`), via a real Kafka request to
`content-creator-agent`. `gemini-flash-lite-latest` worked cleanly at every
budget tested but is a quality step down from the pro tier that was chosen.

The conclusion drawn was that the model thinks by default, thinking cannot be
disabled (`thinkingConfig.thinkingBudget: 0` → 400), and therefore pro is
unusable at our budgets. The provider switch was reversed fleet-wide
(`4dd5d6378`, `5db6a929f`).

## Root cause

**Gemini's `maxOutputTokens` is a total output ceiling and thinking is spent from
it before any visible text. Anthropic's `max_tokens`, with extended thinking off
— which is how every agent in this platform runs — is entirely visible text.
Every `max_tokens` value here was sized against the second definition and passed
verbatim into the first.**

`platform/aiservice/gemini.go`, as it stood (line 86 of the pre-fix file):

```go
generationConfig["maxOutputTokens"] = maxTokens   // maxTokens = options["max_tokens"]
```

The word "thinking" appeared nowhere in the file. So the twitter tier asked a
thinking model to fit its reasoning *and* a tweet into 100 tokens. It spent them
thinking and had none left. **Zero visible text was the arithmetic working
correctly** — not a capability limit.

Two decisions in the same file kept it looking like a model verdict:

1. **The diagnostic figure was in the response and discarded.** The usage decoder
   read `candidatesTokenCount` only, never `thoughtsTokenCount`, so the tokens
   doing the damage were invisible at every layer above the transport — including
   in `llm_call_log`.
2. **The error named the wrong candidate cause.** `finishReason=MAX_TOKENS` alone
   is indistinguishable from a prompt that wanted to write more, and that reading
   points at the model.

A third, independent defect in the same file: the response loop concatenated
**every** part in `candidates[0].content.parts`. Gemini returns reasoning and
answer in that same array, with reasoning flagged `thought: true`. Nothing above
the transport inspects generated copy closely enough to notice reasoning spliced
into a published page. `[UNVERIFIED]` that such a part has ever arrived from this
API — asserted from the documented response shape, fixed defensively.

## Evidence

- `git log -p -- deployments/kustomize/services/content-creator-agent/overlays/production/uk_001/configs/configmap-content-creator.yaml`
  — exactly two commits ever changed the provider: `7b27edfa9` in (07-23 11:36),
  `4dd5d6378` out (07-24 17:11).
- `4dd5d6378`'s message carries the live-API measurements quoted in the table
  above, and the `thinkingBudget: 0` → 400.
- The pre-fix `gemini.go`: `grep -c thinking` → 0.
- `5db6a929f` — `page-content-writer` was flipped to Gemini at 16:53 and reverted
  at 16:59, its `about` test rebuild **still queued**. So the writer half was
  never exercised on Gemini at all.

## Fix (in code, 2026-07-27, `platform/aiservice/gemini.go`)

1. **Reserve, don't fight.** `maxOutputTokens = caller's max_tokens +
   thinking_reserve` (default 8192) for any model assumed to think. A ceiling is
   not a purchase — Gemini bills tokens produced — so headroom is nearly free.
2. **No `thinkingConfig` sent unless configured.** The 2.5 knob
   (`thinkingBudget`, int) and the 3.x knob (`thinkingLevel`, string) are
   incompatible and the wrong one 400s every call. Both are opt-in via
   `ai_service.thinking_budget_tokens` / `thinking_level`, mutually exclusive,
   refused at construction if both are set.
3. **Deny-list polarity.** Only `flash-lite` and `embedding` are treated as
   non-thinking; an unrecognised model is assumed to think, because an
   unfamiliar Gemini name is almost always a newer one.
4. **Thinking decoded.** `thoughtsTokenCount` / `totalTokenCount` decoded and
   written back as `__usage_thinking_tokens` / `__usage_total_tokens`.
   `__usage_output_tokens` stays **visible** tokens so the field means the same
   thing across providers.

   > **CORRECTED 2026-07-27 — this said "Thinking made visible", and that was an
   > overclaim.** Thinking is visible in the **error message** and in the
   > in-process options map. It is **NOT persisted**: `__usage_thinking_tokens`,
   > `__usage_total_tokens`, `__sent_visible_budget_tokens` and
   > `__sent_thinking_reserve_tokens` have **no reader outside
   > `platform/aiservice/`** (verified by grep) and `llm_call_log` has no columns
   > for them, so no query, dashboard or diagnosis bundle can see them. Worse,
   > `llm_call_log.max_tokens` is fed from `__sent_max_tokens`, which for Gemini is
   > the **inflated total** — so that column is about to mean two different things
   > by provider. **Filed as `bugs_open/110`**, which supersedes
   > `features_open/025` item (b). Nobody caught the overclaim — including a
   > ten-seat council, one seat of which discussed these exact fields — because
   > *"writes the field"* and *"the field is readable"* look identical in a diff.
5. **The truncation message names the consumer**, and says which setting to raise
   when visible output is 0 and thinking is non-zero.
6. **`thought: true` parts dropped** from the answer.
7. **Dead pins refused at construction** with the replacement named
   (`gemini-2.5-pro`/`-flash` answer 404 "no longer available to new users" for
   this key); `ai_service.model` has **no default** — the old default was
   `gemini-2.5-pro`, a default that had rotted into a 404 on every call.

Also: `internal/agents/contentcreator/agent.go`'s cost table was keyed on those
two unreachable pins, so it could never match and every Gemini call silently
costed at the Claude fallback rate. Re-keyed to the floating pointers; the rates
are marked `[UNVERIFIED]` inline.

`scripts/gemini-probe.sh` written to answer, against the live key in one run:
what this key can reach · visible-vs-thinking output at each real tier · which
thinking knob the model accepts. Those answers were obtained by hand on 07-24
and survived only inside a commit message.

## Blocked on an unrelated fleet outage

The end-to-end page-build verification (P7) **cannot run**: `bugs_open/029` (hung
spawns saturate the dispatch group, halting all builds fleet-wide, filed 2026-07-19,
still OPEN) has stopped `build-dispatch-loop` since 19 July. A properly queued
`needs_page` item for `dartsonline/grip-styles` was detected and never claimed. This
predates this bug by eight days and is unrelated to the provider switch.

The model-side risks were verified directly instead: the writer's real 12,570-char
prompt at its real 8000 budget returns valid unfenced JSON with all required keys,
`finishReason=STOP`, 1,576 thinking tokens.

## Why this stays OPEN

The fix is Go, so it is **inert until an image rebuild and roll**. Until then the
defect is fully reproducible: flipping any `ai_service` to Gemini today
reproduces 07-24 exactly. Two images are needed —
`page-content-writer` runs inside the chassis, `content-creator-agent` is its own
service.

## How to verify

1. **The binary shipped.** Pod-grep a string this change *created*, with a
   negative control (`RUNBOOK_gemini_content_provider.md` §4):
   `strings /app/agent-chassis | grep -c "thinking consumed the entire output ceiling"` → >0.
2. **The reserve reached the wire.** `llm_call_log.sent_max_tokens` for
   `provider='gemini'` should exceed the caller's budget by the reserve
   (RUNBOOK §5). **Note:** the "`output_tokens == max_tokens` means CUT"
   heuristic goes quiet for a thinking model, because visible output is compared
   against a total including the reserve. The typed `*TruncatedError` on
   `finishReason=MAX_TOKENS` is the authoritative signal — check errors, not
   arithmetic.
3. **The failing branch, induced.** Green output alone proves deployment, not
   correctness. Set `thinking_reserve_tokens: 0` on a scratch config and confirm
   the truncation error names thinking and the reserve setting. Then restore.
4. **The thing it was for.** Probe the live key, pick a model on the tier table,
   flip content-creator, generate one real post; then flip
   `page-content-writer`, rebuild **one** page and read the copy. `complete` is
   not proof — read the artefact.

## What is NOT claimed

That Gemini writes acceptably for us. The tests prove the client now sends
sensible numbers; nothing here measures output quality. **The mechanism can be
right and pro still impractical** — if it thinks for tens of thousands of tokens
on our prompts, the reserve works and the cost does not. The probe's tier table
measures that directly, and it has not been run: `kubectl` is `Unauthorized` in
the filing session and `GEMINI_API_KEY` exists only in the cluster secret.

Full context, phasing and the open owner question:
`docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/`.
Pattern: `016b` §9 *"The same parameter name on two providers is not the same
parameter"*. Wrong call recorded in `WRONG_CALLS.md` (2026-07-24 entry).
