# NOTES — Gemini content provider

*Technical running log. Append-only, newest at the bottom. Record missteps and
wrong turns, not just conclusions. Mark unverified claims
`[INFERRED]`/`[UNMEASURED]`/`[ASSUMED]`.*

---

## 2026-07-27 — workstream opened; the first attempt reconstructed from commits

Owner asked why the Gemini switch was reversed and to revisit it. No workstream
dir existed (the first attempt left its record in commit messages and in another
workstream's NOTES), so this one was opened.

**Grepped before filing** (CLAUDE.md): `bugs_open/`, `bugs_closed/` and
`docs024_key_docs_latest/` for `gemini` — no open bug and no existing workstream
covers the text-generation provider. The Gemini hits in `imagery/` and
`016b §9` are the **image** lane (Banana / `gemini-3-pro-image-preview`), a
different subsystem; the hits in `per_site_ai/` are Gemini used as a strategy
advisor in chat, not as a platform provider. No duplication.

**The six commits that are the whole record** — `014e45ffa` (provider added),
`7b27edfa9` (content-creator flipped), `c8896a37d` (writer flipped),
`5db6a929f` (writer reverted), `4dd5d6378` (content-creator reverted, with the
findings), `3ea9d718c` (max_tokens tiers raised). Timeline in PLAN.

**Misstep, mine, caught immediately.** I started by believing the brochure
workstream's account that the fleet switch-back was sweep `fb6d6ad44`. Checking
`git show --stat` first: that commit has no configmap change at all — 17
`kustomization.yaml` image-tag bumps, the makefile, two docs. Then
`git log -p` on just the configmap, filtered to `provider:`/`model:` lines,
gave the actual answer in one command:

```bash
git log --format="%h %ad %s" --date=format:"%m-%d %H:%M" -p -- \
  deployments/kustomize/services/content-creator-agent/overlays/production/uk_001/configs/configmap-content-creator.yaml \
  | grep -E "^[0-9a-f]{9} |^[+-] *(provider|model):"
```

Exactly two commits ever changed the provider: `7b27edfa9` in, `4dd5d6378` out.
The writer was reverted at 16:59 citing a fleet revert that, in git, landed at
17:11. Cheap check that settled it: **`git log -p` on the one file, not the
sweep's message.** A sweep commit's subject line describes the *sweep's* intent,
not its contents. Corrected in place in the brochure NOTES and recorded in PLAN.

**The finding that reopens this.** `4dd5d6378` attributed the empty/truncated
output to the model: pro thinks, thinking can't be disabled, so pro is unusable
at our budgets. Reading `platform/aiservice/gemini.go` as it stood shows the
cause is ours — old line 86:

```go
generationConfig["maxOutputTokens"] = maxTokens   // maxTokens = caller's max_tokens
```

and the word "thinking" appeared **nowhere** in the file. Gemini's
`maxOutputTokens` is a total output ceiling that thinking is drawn from first;
every `max_tokens` in this platform was sized against Anthropic with thinking
off, where the whole cap is visible text. So the 100-token twitter tier asked a
thinking model to fit reasoning *and* tweet into 100 tokens. Zero visible text is
the arithmetic working correctly.

Two things kept it invisible: `usageMetadata` decoding read only
`candidatesTokenCount`, never `thoughtsTokenCount` (so the tokens doing the
damage were invisible above the transport), and the truncation error said only
`finishReason=MAX_TOKENS`, which looks identical to a prompt that wanted to write
more.

**Status of the 07-24 evidence, restated honestly:** the content-creator tests
measured a starved budget, not writing quality. The writer test **never ran** on
Gemini at all — the queued `about` rebuild was still behind the backlog when the
revert landed (`5db6a929f`). So for the agent that writes our site copy there is
no Gemini evidence in either direction, and the flip is an open experiment rather
than a settled question.

## 2026-07-27 — client fixed (P1); what is asserted and what is not

`platform/aiservice/gemini.go` rewritten around the budget:

- `maxOutputTokens = caller's max_tokens + thinking_reserve` (default 8192) for
  any model assumed to think. Deny-list polarity: only `flash-lite` and
  `embedding` are treated as non-thinking, both measured on 07-24. An unknown
  Gemini name is assumed to think, because an unfamiliar name is almost always a
  newer model — the same polarity lesson as idea.uk's wire-format allow-list.
- No `thinkingConfig` sent unless configured. The 2.5 knob (`thinkingBudget`,
  integer) and the 3.x knob (`thinkingLevel`, string) are incompatible, and 07-24
  already caught the 400 from sending the wrong one. The reserve makes the
  default case work with no knob at all.
- `thoughtsTokenCount` / `totalTokenCount` decoded; thinking written back as
  `__usage_thinking_tokens`. `__usage_output_tokens` stays **visible** tokens so
  the field keeps the same meaning across providers.
- Parts flagged `thought: true` are skipped. Gemini returns reasoning and answer
  in the same `parts` array and the old loop concatenated both.
- Known-closed pins refused at construction with the replacement named;
  `ai_service.model` now has **no default** (the old one was `gemini-2.5-pro`,
  i.e. a default that had rotted into a 404 on every call).

`scripts/gemini-probe.sh` written so the 07-24 answers stop living in a commit
message: model reachability for this key, visible-vs-thinking per real tier,
and which thinking knob the model accepts.

**Verified:** `gofmt` clean; `go build ./platform/... ./internal/...` clean;
`go test ./platform/aiservice/` green including 11 new Gemini tests, which pin
the reserve at the exact tier (100) that produced zero text in production and
assert the truncation message names thinking as the consumer.

**NOT verified, and these are the ones that matter:**
- `[UNMEASURED]` that any of this makes Gemini produce usable text. The tests
  prove the client sends the right numbers; only the live probe and a real
  generation prove the numbers were the problem. **The reserve theory could be
  right about the mechanism and still leave pro unusable** — e.g. if it thinks
  for tens of thousands of tokens on our prompts. The probe's tier table
  measures that directly.
- `[UNVERIFIED]` that `thinkingLevel` is the accepted knob on
  `gemini-3.1-pro-preview`. Inferred from the 07-24 400 on `thinkingBudget: 0`
  plus the API's generational split. The client sends neither by default
  precisely so this inference is not load-bearing; the probe settles it.
- `[UNVERIFIED]` that a `thought: true` part ever appears from this API. The
  filter is asserted from the documented response shape and is harmless if such
  parts never arrive.
- `[UNVERIFIED]` the Gemini rates in content-creator's `estimateCost` table.
  The old keys (`gemini-2.5-pro`/`-flash`) were unreachable model names, so they
  could never match and every Gemini call silently costed at the Claude fallback
  rate. Re-keyed to the floating pointers with the 2.5-era rates carried over
  and marked `[UNVERIFIED]` inline — not checked against Google's price list.
- `[UNVERIFIED]` that `text-embedding-004` is still reachable. Same retirement
  class as the 2.5 pins. Made configurable rather than changed, since nothing
  here is known to call `GenerateEmbedding` on Gemini.

**Blocked:** `kubectl` is `Unauthorized` in this session, so P3–P6 (image roll,
pod verification, probe, both flips) cannot start. `GEMINI_API_KEY` exists only
in the cluster secret — there is no local copy, so the probe cannot be run from
here either. Everything up to the probe is done and waiting on credentials.
