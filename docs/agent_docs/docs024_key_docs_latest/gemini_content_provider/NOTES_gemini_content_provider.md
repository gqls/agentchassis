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

## 2026-07-27 (later) — cluster auth restored; P4 probe RUN, and it corrected me twice

Owner restored `kubectl` and asked for the probe + the council gate.

**Model reachability, re-verified today** (not carried over from 07-24). Key read
from pod `content-creator-agent-84564dfb67-vjq5g`:

| model | result |
|---|---|
| `gemini-2.5-pro` | **404** "no longer available to new users" |
| `gemini-2.5-flash` | **404** "no longer available to new users" |
| `gemini-3-pro-preview` | **404** "no longer available" (retired outright, different message) |
| `gemini-3.1-pro-preview` | OK |
| `gemini-pro-latest` | OK (→ 3.1-pro-preview) |

**The listing advertises models the key cannot call.** `models?pageSize=200`
returns 42 `generateContent` models including `gemini-2.5-pro` and
`gemini-3-pro-preview`, both of which 404. So the probe's own warning was right
and worth keeping: **a model appearing in the listing is not evidence the key can
reach it.** The `geminiRetiredPins` construction guard is therefore correct, and
correct *today*, not just on 07-24.

**Tier table, `gemini-pro-latest`, trivial prompt** — the 07-24 failure
reproduced exactly:

| max_tokens | finish | visible tok | thinking tok | chars |
|---|---|---|---|---|
| 100 | MAX_TOKENS | 4 | 92 | 23 |
| 500 | MAX_TOKENS | 19 | 477 | 107 |
| 1200 | STOP | 38 | 1145 | 224 |
| 3000 | STOP | 37 | 888 | 213 |
| 6000 | STOP | 44 | 786 | 228 |

**Thinking expands to fill a small ceiling** (92 of 100, 477 of 500) and settles
at ~800–1,150 once the ceiling is comfortable. That is the mechanism, measured.

**Tier table on the REAL 12,570-char writer prompt** (placeholders filled), which
is the figure that sizes the reserve:

| config | max_tokens | thinking tok | visible tok |
|---|---|---|---|
| none | 8000 | **2,764** | 99 |
| none | 3000 | **2,878** | 103 |
| `thinkingLevel: low` | 8000 | 1,080 | 57 |
| `thinkingBudget: 512` | 8000 | 940 | 55 |

So the 8192 default carries ~3x headroom on the real workload. **It is now
measured rather than chosen** — recorded in the constant's comment.

> **CORRECTED 2026-07-27 — my own claim, falsified by the probe.** PLAN D2, the
> NOTES entry above, the `016b` §9 pattern and the commit message all said the two
> generations take *incompatible* knobs: "3.x takes a `thinkingLevel` string and
> rejects the integer with a 400". **False.** On `gemini-pro-latest`:
> `thinkingBudget: 512` → **ACCEPTED**; `128` → ACCEPTED; `32768` → ACCEPTED;
> `thinkingLevel: "low"`/`"high"` → ACCEPTED. Only `thinkingBudget: 0` is
> rejected, and its message says exactly why: *"Budget 0 is invalid. This model
> only works in thinking mode."*
> The 07-24 observation was right and narrow (that one value 400s). **The
> generalisation was mine**, built from a single rejected value, and it is the same
> error shape as the one this whole workstream exists to correct — reasoning from
> one refusal to a structural claim without testing the neighbours. Cheap check
> that caught it: three more values of the same parameter, ~30 seconds.
> No harm done to the code: the client sends neither knob by default precisely so
> this inference was never load-bearing. The comments asserting it are corrected.

**And a second correction, this one to the guard's rationale rather than the
guard.** Sending both knobs together IS refused — *"You can only set only one of
thinking budget and thinking level"* — so the mutual-exclusion check is right. But
my stated reason for it (two generations, two knobs) was wrong. Right check, wrong
why, which is a worse state than it looks: it would have survived any review that
read the reason instead of testing the behaviour.

**The finding that actually changes the plan: neither knob CAPS thinking.**
`thinkingBudget` is a soft target the model overshoots freely — 128 requested →
483 spent; 512 → 903/940; 32768 → 783. It reduces thinking substantially (2,764 →
~940 on the real prompt) but bounds nothing. So a knob is a **cost lever, not a
correctness one**: it cannot replace the reserve, and any plan that says "just set
thinkingBudget and drop the reserve" is wrong.

**A probe fault I nearly reported as a finding.** The first knob run printed all
three knobs as `REJECTED: contents is not specified`. That was my script: `jq
--argjson` was fed jq syntax (`{thinkingConfig:{thinkingLevel:"low"}}`, unquoted
keys) instead of JSON, so jq emitted nothing, curl posted an empty body, and the
API complained about the *missing prompt* — which reads exactly like the API
refusing the knob. **A malformed request and a refused parameter produce the same
shape of "no".** Fixed, and the script now reports a request that fails to BUILD
as `PROBE FAULT (NOT a verdict)` rather than letting it masquerade as one. Had I
believed it, I would have "confirmed" my own falsified claim with my own bug.

**Path correction, found by running it.** RUNBOOK §0's original query for the
writer's provider returned four NULL columns and no error: `generate_content` is
not a top-level step. Real path:
`workflow → steps → process_sections_loop → config → sub_workflow → steps →
generate_content → config → ai_service`. **A jsonb `->` path that returns NULL has
not told you the value is absent — it may have told you the path is wrong**, and
the two are indistinguishable without walking the keys. Also: `steps` is an
*object* keyed by step name, so `jsonb_array_elements` errors ("cannot extract
elements from an object") while `jsonb_each` works.

**Two live facts about the writer that change the framing:**
1. Its step budget is **`max_tokens: 8000`**, not one of content-creator's
   100/1200/3000/6000 tiers. At 8,000 with thinking around 2,800 there is room to
   spare — so `[INFERRED, from a context-stripped probe]` the 07-24 starvation
   probably would **not** have bitten the writer. The starvation is a
   *small-tier* defect, and content-creator's twitter/short tiers are where it
   lived. For the writer the fix is insurance plus the thinking-token visibility
   and the thought-part filter. Marked INFERRED because a real run carries site
   specs, brief, existing content and link context — a far bigger prompt than the
   12.7K template alone, and thinking scales with prompt complexity.
2. Its prompt template is **12,570 chars**, not the 7.8K recorded in the brochure
   NOTES. Grown since. Re-measure rather than quoting.

**Visible-output figures from these probes are NOT quality evidence** and must not
be quoted as such: the template's placeholders were filled with
"(context omitted for probe)", so the model had nothing real to write about. The
55–103 visible-token counts measure that, not its writing. One run at tier 3000
returned 2 characters with `finishReason=STOP` — almost certainly an empty JSON
object, and exactly the sort of number that would become "Gemini writes nothing"
if lifted out of context. The **thinking** figures are the usable output here.
