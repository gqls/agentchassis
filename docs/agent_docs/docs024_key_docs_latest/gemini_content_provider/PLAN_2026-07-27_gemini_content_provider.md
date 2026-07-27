# PLAN — putting the content-producing agents on Gemini (second attempt)

*Opened 2026-07-27. Design, phasing, decisions and their reasons. Corrections to
the originating brief live here, marked as corrections.*

---

## What we are trying to do

Point the two content-producing agents at Gemini instead of Claude:

| agent | what it writes | where its provider is configured | live now |
|---|---|---|---|
| `content-creator-agent` | blog posts, social content | kustomize configmap (`configmap-content-creator.yaml`) — a k8s service | `anthropic` / `claude-sonnet-5` |
| `page-content-writer` | the site page sections (the copy on every site we publish) | `agent_definitions` row, `generate_content` step's `ai_service` | `anthropic` / `claude-sonnet-4-6` |

This was attempted on 2026-07-23/24 and reversed. This plan exists because the
reversal was correct on the evidence available at the time **and the evidence was
incomplete** — the observed failure had a cause nobody named.

## Where we've come from — the first attempt, in order

Reconstructed from commits, not from memory. Times are BST.

| when | commit | what happened |
|---|---|---|
| 07-23 11:30 | `014e45ffa` | Gemini added as a selectable provider: `platform/aiservice/gemini.go`, `aiservice.NewClient` factory, content-creator wired to it. Config left on `anthropic` — capability without a switch. |
| 07-23 11:36 | `7b27edfa9` | content-creator configmap flipped to `gemini` / `gemini-2.5-pro`. |
| 07-23 16:25 | `fb6d6ad44` | sweep, "v1.0.1151 — prior to more automated reliance on Claude". **Image tags only.** |
| 07-24 16:53 | `c8896a37d` | `page-content-writer` flipped in the DB to `gemini` / `gemini-2.5-pro`; `about` re-queued as the one-page test. |
| 07-24 16:59 | `5db6a929f` | `page-content-writer` reverted to `anthropic` / `claude-sonnet-4-6`, on the owner's "we have switched back". Voice & Style prompt kept. |
| 07-24 17:11 | `4dd5d6378` | content-creator configmap reverted to `anthropic` / `claude-sonnet-5`, with the findings below. Live ConfigMap had been patched directly just ahead of it. |
| 07-24 17:14 | `3ea9d718c` | `max_tokens` tiers raised (500→1200, 2000→3000, 4000→6000) — a *pre-existing* gap the Gemini work exposed. |

> **CORRECTION to the record (2026-07-27).** `NOTES_brochure_component_library.md`
> states the fleet switch-back "was fleet-level (sweep `fb6d6ad44` … reverted the
> content-creator service)". `fb6d6ad44` contains **no** configmap change — it
> bumps 17 `kustomization.yaml` image tags, the makefile and two docs. In git the
> content-creator provider was reverted only by `4dd5d6378`, twelve minutes
> *after* the writer was reverted to match it. The trigger was the owner's
> instruction, not that commit. Net state ended the same, so nothing was harmed;
> it matters only because a reader would otherwise go to `fb6d6ad44` looking for
> the provider decision and find image tags. Corrected in place at both sites.

## Why it was reversed — the two findings, and the one that was missed

Both were verified against the real pod and the real API on 07-24, so they are
findings, not guesses. The reversal was the right call on them.

**Finding 1 — the pinned model was closed to our key.** `gemini-2.5-pro` and
`gemini-2.5-flash` both answer **404 "no longer available to new users"** for
this platform's API key. Google gates pinned model *generations* by key
provenance, so this is a property of the key, not of the model. `gemini-pro-latest`
(then resolving to `gemini-3.1-pro-preview`) worked. This one was legible: a 404
naming its own cause.

**Finding 2 — the model produced almost no text.** On `gemini-pro-latest`,
thinking could not be switched off (`thinkingConfig.thinkingBudget: 0` → 400) and
"ate a large, variable share of maxOutputTokens before any visible text": at the
100-token twitter tier, **zero** text; at the 500-token short tier, **~85
characters** before truncating. `gemini-flash-lite-latest` had no thinking
overhead and worked cleanly at every budget, but is a real quality step down from
what "pro" was chosen for. The owner chose to revert rather than take that trade
silently. **That judgement was sound. The framing under it was not.**

**Finding 3 — what nobody named.** Finding 2 was read as a property of the
*model*: pro thinks, thinking is mandatory, therefore pro is unusable at our
budgets and the only Gemini option is a quality step down. It is in fact a
property of **our client**. `platform/aiservice/gemini.go` set
`generationConfig.maxOutputTokens` to the caller's `max_tokens` verbatim
(old line 86) and never referenced thinking anywhere in the file.

Gemini's `maxOutputTokens` is a **total output ceiling, and thinking is spent
from it first**. Every `max_tokens` number in this platform was sized against
Anthropic, where — with extended thinking off, which is how every agent here runs
— the entire cap is visible text. So passing that number through does not *cap*
a thinking model's answer, it *starves* it. At the twitter tier we asked a
thinking model to fit its reasoning and its tweet into 100 tokens; it spent them
thinking and had none left to speak with. That is arithmetic, not quality.

The reserve costs nothing to fix, because **`maxOutputTokens` is a ceiling, not a
purchase** — Gemini bills tokens actually produced. Asking for 8k of headroom on
a call that thinks for 300 tokens costs the price of 300 tokens.

The old client also made the failure hard to see: it decoded only
`candidatesTokenCount` from `usageMetadata`, never `thoughtsTokenCount`, so the
tokens doing the damage were invisible at every layer above the transport, and
the error said only `finishReason=MAX_TOKENS` — indistinguishable from a prompt
that wanted to write more.

**This is why the first attempt is worth re-running rather than re-deciding.**
Nothing was learned about Gemini's writing quality on 07-24. The content-creator
tests measured a starved budget. The `page-content-writer` test **never ran at
all** on Gemini (the queued `about` rebuild was still behind the backlog when the
revert landed, per `5db6a929f`) — so for the agent that writes our actual site
copy, there is *no* Gemini evidence in either direction.

## Decisions

**D1 — Fix the client before touching any config.** The provider switch is a
one-line config change; the reason it failed is in Go, and Go changes are inert
until an image rebuild and roll. Flipping config first would reproduce 07-24
exactly. *(Landed 2026-07-27, see NOTES.)*

**D2 — Reserve budget rather than fight thinking.** The client now sends
`maxOutputTokens = caller's max_tokens + thinking_reserve` (default 8192) for any
model assumed to think. This deliberately does **not** depend on getting a
thinking-suppression knob right: the two Gemini generations take different and
mutually incompatible knobs (2.5 an integer `thinkingBudget`, 3.x a
`thinkingLevel` string that rejects the integer with a 400 — which is exactly the
400 seen on 07-24), and guessing wrong fails every call. So by default the client
sends **no** `thinkingConfig` at all and lets the reserve absorb whatever the
model spends. Both knobs are available as opt-in config once the probe says which
one this model accepts.

*Accepted trade, stated plainly:* for a thinking model `maxOutputTokens` can no
longer serve as a visible-length cap. Length has to be instructed in the prompt.
This is the same trade Anthropic extended thinking makes, and the twitter tier's
100 tokens — an intentional platform-side limit per `3ea9d718c` — is now an
instruction rather than a ceiling.

**D3 — Unknown model ⇒ assume it thinks (deny-list, not allow-list).** Only
`flash-lite` and `embedding` are treated as non-thinking, both measured. An
unfamiliar Gemini name is almost always a *newer* one and every generation since
2.5 thinks by default, so an allow-list would under-provision each new model on
the day it ships — and under-provisioning presents as bad copy, not as an error.
(Same polarity lesson as the idea.uk wire-format selector, where an allow-list
meant a model swap alone 400'd every call.)

**D4 — A dead model pin fails at construction, not on every call.** Known-closed
pins (`gemini-2.5-pro`, `gemini-2.5-flash`) are refused by `NewGeminiClient` with
the working replacement named. `ai_service.model` now has **no default** — the
old default was `gemini-2.5-pro`, i.e. a default that had quietly rotted into a
404 on every call, which reads as an outage rather than a config error.
*Not* auto-substituted: silently serving `gemini-3.1-pro-preview` to an operator
who asked for `gemini-2.5-pro` is a different model at a different price behind
the same config, and that silent-default shape is the one this repo keeps filing
bugs about (see `bugs_closed/011`).

**D5 — Thought parts are dropped from the answer.** Gemini returns reasoning and
answer in the *same* `parts` array, and the old loop concatenated every part.
Parts flagged `thought: true` are now skipped. Defensive — with no thinking
summary requested they should not appear — but if they ever did, the model's
reasoning would be spliced into published page copy and nothing above the
transport reads it closely enough to notice. `[UNVERIFIED against a live
response]`: no `thought:true` part has been observed from this API here; the
filter is asserted from the API's documented shape and is safe either way.

**D6 — Probe the key before choosing a model.** `scripts/gemini-probe.sh` exists
because on 07-24 these answers were obtained by hand, mid-incident, and survive
only inside a commit message. It answers all three questions (what this key can
reach · visible text vs thinking at each real tier · which thinking knob is
accepted) in one run, and the tier table *is* the reserve calculation.

**D7 — Model choice is the owner's, and it is now a config line.** The 07-24
revert turned on a quality trade (pro starved vs flash-lite capable). With the
reserve in place that trade may not exist — pro at a provisioned budget was never
tested. The probe produces the evidence; the pick stays an owner decision.

> **ANSWERED 2026-07-27 (owner): quality / provider diversity — go with pro.**
> So the target config for both agents is `provider: "gemini"`,
> `model: "gemini-pro-latest"`, and the thinking reserve is load-bearing rather
> than incidental: pro is exactly the tier that produced nothing on 07-24, and
> the reserve is the only reason to expect a different result this time.
>
> Two consequences that follow from picking pro specifically, both to be settled
> by the probe (P4) and not assumed:
> 1. **The 8192 default may be too small.** It was chosen as generous headroom,
>    not measured against pro's actual thinking on our prompts. The probe's
>    `THINK_TOK` column *is* the reserve calculation — set the reserve from the
>    largest figure observed, with headroom.
> 2. **Cost is now the live risk, not truncation.** Thinking tokens are billed as
>    output. A pro model that thinks for tens of thousands of tokens per section
>    works correctly under this fix and is still the wrong answer commercially.
>    The probe reports thinking tokens per tier, so this is measurable before any
>    site-wide use — and `__usage_thinking_tokens` now makes it trackable after.
>    If the numbers are bad, flash-lite remains the fallback the owner declined
>    on quality, and that trade should be re-put with real figures attached
>    rather than assumed.
>
> Not to be confused with a green light to flip: P3 (roll the image) still comes
> first, because a Gemini config on a chassis that predates the fix reproduces
> 07-24 exactly.

## Phasing

- **P1 — client fix.** `platform/aiservice/gemini.go` + tests + probe script.
  **DONE 2026-07-27**, committed, `go test ./platform/aiservice/` green.
- **P2 — council review.** Platform-code change, so it goes through the gate
  (advisory, ~30 min). **NOT DONE.**
- **P3 — ship it.** Bump `IMAGE_TAG`, `make build-agent-chassis` +
  `build-content-creator-agent`, push, deploy, verify against the running pod.
  Both agents need it: the writer runs in the chassis, content-creator is its own
  service. **NOT DONE — blocked, see below.**
- **P4 — probe the live key.** `scripts/gemini-probe.sh --from-pod <model>`.
  Produces the model shortlist, the tier table and the reserve figure.
  **NOT DONE — blocked.**
- **P5 — flip content-creator** (configmap, live immediately once the image
  carries P1) and generate one real blog post through Kafka. Compare against a
  Claude baseline. **NOT DONE.**
- **P6 — flip `page-content-writer`** (DB, guarded on `updated_at`, backup table
  first — the `bak_agent_definitions_pcw_20260724` pattern from 07-24) and
  rebuild **one** page. Read the copy against the Voice & Style rules before any
  site-wide rewrite. **NOT DONE.**

P5 and P6 are independent and reversible in either order. P6 is the one that
matters commercially — it writes the sites — and it is the one with no evidence
at all, so it should not be skipped on the strength of a P5 result.

## Blocked on

**Cluster credentials.** `kubectl` returns `Unauthorized` in this session
(context `personae-uk001-prod-agent-chassis-cluster`), so P3–P6 cannot start:
no image roll, no pod verification, and no access to `GEMINI_API_KEY`, which
lives only in the cluster secret (`personae-default-secrets`) — there is no
local copy. Everything up to and including the probe script is done and waiting.

## Open question for the owner

Which tier do we actually want, and why did we want Gemini in the first place?
The repo does not record the motive, and it changes the answer. If the motive was
**cost**, `gemini-flash-lite-latest` already worked on 07-24 and needs no reserve
at all. If it was **quality or provider diversity**, `gemini-pro-latest` is the
candidate and the reserve is what makes it viable — but "pro with a provisioned
budget" has genuinely never been tested, so this is an experiment with a real
possibility of the same answer as last time. Either way the flip is now a config
line, so the cost of trying pro first and falling back is one probe run.
