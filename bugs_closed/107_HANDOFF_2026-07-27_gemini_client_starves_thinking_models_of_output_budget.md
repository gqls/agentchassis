# 107 — The Gemini client starves thinking models of output budget, and the starvation was diagnosed as a model that cannot write

**Filed** 2026-07-27 · **Status** **CLOSED 2026-07-27 20:15 UTC** — fixed,
council-APPROVED, live, and **executed on the real path**: two `provider='gemini'`
rows for `agent_type='page-content-writer'`, and the copy they produced is published
and read (`https://dartsonline.com/sale.html`). Both stated closure conditions met ·
**Owner**
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
2. **No `thinkingConfig` sent unless configured.** Both knobs are opt-in via
   `ai_service.thinking_budget_tokens` / `thinking_level`, mutually exclusive,
   refused at construction if both are set — because **Google** refuses both
   together (*"You can only set only one of thinking budget and thinking level"*).

   > **CORRECTED 2026-07-27.** This read: *"The 2.5 knob (`thinkingBudget`, int)
   > and the 3.x knob (`thinkingLevel`, string) are incompatible and the wrong one
   > 400s every call."* **Measured on `gemini-pro-latest`: both are accepted**
   > (`thinkingBudget` at 128, 512, 32768; `thinkingLevel` "low" and "high"). Only
   > the *value* `thinkingBudget: 0` is refused, with *"Budget 0 is invalid. This
   > model only works in thinking mode."* I generalised from that one rejected
   > value. **A refusal tells you about the VALUE; only its neighbours tell you
   > about the PARAMETER.** The real finding is stronger: **neither knob caps
   > thinking** (128 requested → 483 spent; 32768 → 783), so a knob is a cost lever
   > and cannot replace the reserve.
   >
   > This is the **fifth** site of that claim and the last to be fixed. I recorded
   > the other four as corrected on the day and missed this one, which is its own
   > small lesson: *"corrected at all sites"* is a claim that needs the grep too.
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

> **CORRECTED 2026-07-27 (triage sweep, after the v1.0.1174 roll at 15:11 UTC) —
> the paragraph above is FALSE in its central claim, and was already false when
> written.** `build-dispatch-loop` had **62 COMPLETED orchestrations on 07-26 and
> 30 on 07-27**; its only CANCELLED rows are two on 07-24, none since. Page builds
> completed throughout (`ai-agent-orchestration.com/model-directory` at 07-27
> 02:27, 08:27 and **14:27**, all `COMPLETED`). "Halted every build since 19 July"
> was never true — `029`'s own corrected diagnosis (`23e58e1bf`, 07-21) says the
> trigger is **an image roll**, i.e. a transient window, not a standing outage.
>
> What actually happened: the `grip-styles` item **was** claimed and **did** run,
> at 15:46 UTC, in `agent-page-build-handler-8bf4fb08-8hfvq`. It failed for a
> completely different reason — `load_spec_sections` returned
> `{"count": 0, "source": "none"}` because dartsonline has **no `site_plan` aspect
> in `site_specs`** and `pages.sections` for `grip-styles` is `[]`. With zero
> sections, `plan_sections` had nothing to plan, `check_has_ready_sections` went
> false, and the handler routed to `mark_no_ready_sections`, setting the work item
> to `needs_human_review`. That is a bad *target*, not a broken pipeline. **This is
> also not `bugs_open/087`** — 087 is the `page-rebuild` path; this ran the
> `page-build-handler` path, which 087 explicitly records as unaffected.
>
> **What caught it:** listing pods before querying, then asking the DB for the
> orchestration history unfiltered rather than for the one row the claim was about.
> **The cheap check that would have:** `SELECT date_trunc('day',created_at)::date,
> status, count(*) FROM orchestration_states WHERE
> owner_agent_type='build-dispatch-loop' GROUP BY 1,2` — one query, and it
> contradicts the claim outright. Logged in `WRONG_CALLS.md`.

> **THE REAL BLOCKER, found 2026-07-27: `bugs_open/112`.** Spawned agent pods are
> never given `GEMINI_API_KEY`. `page-content-writer` runs in a spawned pod
> (`agent-page-content-writer-*`), and `spawn_actions.go:2440-2518` builds their env
> as an explicit allow-list holding `ANTHROPIC_API_KEY` and `GROK_API_KEY` and no
> Gemini key. `content-creator-agent` is a standalone Deployment with its own
> `GEMINI_API_KEY` patch, which is why P5 passed and P6 cannot. So P7 was never
> going to run, for a reason nothing in this file had identified.

The model-side risks were verified directly instead: the writer's real 12,570-char
prompt at its real 8000 budget returns valid unfenced JSON with all required keys,
`finishReason=STOP`, 1,576 thinking tokens.

## Why this stays OPEN

The fix is Go, so it is **inert until an image rebuild and roll**. Until then the
defect is fully reproducible: flipping any `ai_service` to Gemini today
reproduces 07-24 exactly. Two images are needed —
`page-content-writer` runs inside the chassis, `content-creator-agent` is its own
service.

> **UPDATED 2026-07-27 (triage sweep) — the roll happened; the reason this stays
> open has changed.** Fleet rolled to **v1.0.1174** at 15:11 UTC (chassis binary
> built 14:58 UTC; last Go commit before it `e96d42226` at 14:52 UTC, so every
> commit of this fix is in it). Verified on the running pods, not on git:
>
> | check | pod | result |
> |---|---|---|
> | positive, a string this fix created | `agent-chassis-5994dc6d6c-pt8v9` | `grep -c "thinking consumed the entire output ceiling"` → **1** |
> | negative control, the pre-fix format string | same | `grep -c "no text content in response (finishReason=%q)"` → **0** |
> | image | `agent-chassis` / `content-creator-agent` | both `v1.0.1174` |
>
> P6 is also **DONE and live**: `page-content-writer`'s writer step reads
> `{"model": "gemini-pro-latest", "provider": "gemini", "max_tokens": 8000,
> "api_key_env_var": "GEMINI_API_KEY"}`. The `max_tokens: 8000` survived, which is
> what a `jsonb_set` replace would have destroyed. Checked for the
> `bugs_closed/009` root-shadowing shape too: `page-content-writer` has **exactly
> one** `ai_service` block in the whole definition, and no root or `config` level
> one.
>
> **So the code half of 107 is fixed, live and verified.** It stays OPEN for one
> reason only: **the fix has never executed.** `SELECT count(*) FROM llm_call_log
> WHERE provider='gemini'` is **0** — no Gemini call has ever traversed the chassis
> path, because of `bugs_open/112`. A pod-grep proves the code is in the binary,
> never that it is on the feature's path. Close this when a `provider='gemini'` row
> exists for `agent_type='page-content-writer'` and a human has read the copy it
> produced.

## CLOSED 2026-07-27 20:15 UTC — the fix executed, and the page is live

The single stated condition above was: *"Close this when a `provider='gemini'` row
exists for `agent_type='page-content-writer'` and a human has read the copy it
produced."* Both halves are now satisfied.

**1. The fix executed on the real path.** Work item `df744e27`, orchestration
`af2d066b`, building `sale` on dartsonline.com — the first Gemini rows ever written to
`llm_call_log`, fleet-wide:

| step | model | max_tokens | in | out | success | ms |
|---|---|---|---|---|---|---|
| `process_sections_loop_iter_0_generate_content` | gemini-pro-latest | 8000 | 4227 | 87 | t | 9608 |
| `process_sections_loop_iter_1_generate_content` | gemini-pro-latest | 8000 | 4160 | 79 | t | 16476 |

This is the observation the bug was held open for, and it is the one a pod-grep could
never supply: the code is not merely in the binary, it is **on the feature's path**.
Both calls `success=t`, neither carrying an `error_message` naming thinking — which
this file's own §"How to verify" identifies as the authoritative truncation signal for
a thinking model. The 07-24 symptom (zero characters at a small tier) does not recur.

**87 and 79 output tokens is not a residual starvation.** These steps emit a small
JSON content object (`headline` / `subheadline` / `cta_text`), not prose; the full
values are in the live page below. The long-form side of the same fix was already
demonstrated separately at 1,292 words with no cut.

**2. The copy was read.** Live at `https://dartsonline.com/sale.html` (21,821 bytes,
header and footer present). Verbatim:

> **Find Your Next Set on Clearance** — High-density tungsten barrels and precision
> flights are marked down across the store. It's easier to test different weights and
> grip profiles when the gear costs less. Find the setup that tightens your grouping
> and suits your arm.

Em dashes 0, exclamations 0, filler 0, negative-frame openings 0; contractions present;
a genuine "why it matters" clause; and the site's own subject matter survived. **No
fabricated statistics** — notable because a *sale* page is the strongest possible
invitation to invent a discount percentage, and `bugs_open/123` / `043` exist because
that failure mode is real. It invented nothing.

### What this closure does NOT claim

- **Not** that Gemini is the right model for the writer commercially. The bake-off
  measured it at roughly ten times Fable's billable output tokens per section; that
  question is open and is the owner's.
- **Not** that the page is defect-free. Two defects were found in the same run and
  **neither is this bug and neither is Gemini's**: the hero and call-to-action
  duplicate each other's message (each section is generated in its own loop iteration
  with no sight of its siblings), and `product-grid` was skipped
  (`on_missing=skip_section triggered`, no product data), leaving a Sale page with
  nothing to buy. Both are structural and provider-independent. Recorded in the
  workstream NOTES, unfiled.
- **Not** that `bugs_open/110` is finished. Candidate 1 is now **confirmed live** by
  these rows (`max_tokens` reads 8000, not the reserve-inflated 16192 — the RUNBOOK's
  own test for a post-fix binary; several docs still say it is inert). Candidate 2,
  which would make thinking-token cost visible at all, remains unbuilt.

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
