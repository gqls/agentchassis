# EXPERIMENT (pre-registered) — model trials on a constant post-fix prompt, 2026-08-31

**Owner instruction (ruling 11):** *"Let's try different models until we find the best. I think
Claude Fable will probably be too expensive but let's try it as it might give us a benchmark for
the other models, then try Grok and Gemini next."* Motivated directly by the canary's P2a result:
the comparison-shape carrier is the MODEL PRIOR, so the model is the variable left to try.

## Subject — one prompt, held constant

`llm_call_log` id `79257fb4-fcfa-4ff6-9923-dc4e7fcd2b6a` — finetuning.uk, the `about-content`
section of the APPROACH page (the worst-scoring page of the canary), built 2026-08-26 19:17 on
`claude-sonnet-5`. This prompt is FULLY post-fix: template post-627/628, brief post-646/647
(de-demonstrated) AND post-19:03 (the owner's comparison rule present as a positive instruction).
**Production baseline on this exact prompt (sonnet): the shipped section scored NEG=5** (round-2
canary scoring; the page totalled 10).

## Arms

Fable first (the benchmark), Grok and Gemini next per the owner's ordering:
- `claude-fable-5` ×2 (same API/protocol as production; max_tokens 16000, default temperature)
- next round: xAI Grok ×2, Gemini ×2 (keys confirmed present on the pod for all three)

## Metrics, defined now

Battery (six negation shapes + plainly/honest) over the JSON output's flattened text — same
scorer as the canary — plus a READ of each output. Primary comparison: NEG per output vs the
sonnet baseline's 5, on the identical prompt. n=2/arm detects only gross differences; this is a
SCREEN, not a rate estimate.

## Pre-registered readings

- A model at NEG ≤1 with a competent read = a genuinely different prior; promote it to a wider
  screen (more sections, more sites) before any production claim.
- A model ≈5 = same prior behaviour; expense decides nothing.
- The read can VETO a low count (a model that avoids comparisons by writing vaguely fails — the
  owner's target is benefit-led concreteness, not absence of a shape).
- Family caveat, stated in advance: Fable shares a family with sonnet; a similar count is
  expected and would NOT refute the trial premise — it is the benchmark the owner asked for.

## Results

*(appended after the runs)*

**Fable arm run 2026-08-31 (2 calls, ~12,950 tokens in / ~1,850 out each):**

| arm | NEG (six shapes) | plainly | honest | words | read |
|---|---|---|---|---|---|
| production sonnet (shipped) | **5** | 0 | 0 | ~600 (section) | good prose, comparison tic |
| claude-fable-5 #1 | **0** | 1 | 0 | 603 | passes |
| claude-fable-5 #2 | **0** | 0 | 0 | 537 | **passes decisively** |

**The read (the half that can veto, and it does the opposite):** F2 achieves zero not by
vagueness but by STATEMENT — *"A model we fine-tune on your documents belongs to you"*, *"No
vendor pays us, so the choice is made on fit alone"*, *"You decide what leaves, and the default
is nothing"*, *"a model you can host anywhere is a model nobody can take away from you"*. It is
benefit-led ("pays for itself", "shortens the approval conversation considerably"),
non-presumptive ("If that sounds like a sensible way in…"), keeps every fact (£99, $5,000, RAG,
LoRA, three internal links), and carries none of the self-narration class. **This is ruling 12's
target register, produced without ruling 12's machinery.**

**Reading, per the pre-registration:** NEG ≤1 with a passing read → PROMOTE to a wider screen.
The family caveat INVERTED — same family, opposite behaviour — which sharpens the mechanism
question honestly: on this prompt the owner's comparison rule was PRESENT, so the difference may
be instruction-following capability rather than raw prior (Fable obeyed the rule sonnet ignored).
Either way the operative conclusion is the same: **the model is the lever that works.** One
residual: F1's single "plainly" — ruling 9's ban postdates this prompt; the wider screen should
use a current prompt carrying it.

**Next arms (owner's ordering):** Grok ×2, Gemini ×2 on the same prompt (keys confirmed on the
pod: XAI/GROK + GEMINI). Then, if Fable holds at a wider screen: the cost question is the
owner's, with the token counts above as the input.

---

**RETROACTIVE CORRECTION to the Fable rows (owner recalibration, same day):** the owner read the
F samples and named DENSITY a first-class fault — *"a small first project lets you judge the work
on results"* is "a riddle"; models compress and *"we need to put in effort to expand the words
more… every time"*. F2's "passes decisively" is downgraded to **register-pass / density-fail**,
and the read column below carries density explicitly. Instrument note stands: our read is not his
ear either.

**Gemini arm run 2026-08-31 (`gemini-3.1-pro-preview`, ×2, same prompt, maxOutputTokens 16000,
default temperature; pod-side key, generateContent):**

| arm | NEG (battery) | read | words* | facts kept | verdict |
|---|---|---|---|---|---|
| gemini-3.1-pro #1 | **1** ("rather than forcing a single platform") | negation tic at the HIGHLIGHT surface — the exact surface of the owner's finetuning screenshot — plus an implied-competition title ("…that *actually* belong to you"); otherwise statement-led, simple sentences, density fine | 330 | £99+$5k kept; links legitimate (all 3 in the allowed list — checked, not assumed) | **register-marginal** |
| gemini-3.1-pro #2 | **0** | **battery-zero by EVASION, read vetoes**: "We do not tie you to a single provider", "Automation does not mean losing control", "Off-the-shelf AI speaks in generalities" — definition-by-negation in a `do not <verb>` form neither the battery nor the gate's seven shapes match. AND **invented substance**: "protecting you from sudden price changes and service outages" — zero grounding in the prompt (`grep -c` 0 for both) | 442 | **£99/$5,000 DROPPED; 0 internal links** (allowed 3) | **FAIL — register + grounding** |

*words = scorer count over the raw JSON, identical measure across arms (F1 628 / F2 562 on the
same measure). Gemini writes at roughly half Fable's length — under-delivery, not concision: the
missing half is the facts.

**Reading, per the pre-registration:** neither Gemini arm promotes. G1 sits at NEG=1 but the tic
lands on the highest-value surface; G2 is the case the pre-registration warned about in its
sharpest form — a low count achieved by lexical evasion, vetoed by the read, compounded by the
one fault class we treat as disqualifying for a production writer (invention; both lanes rejected
Fable's expansions for exactly this). **The screen's standing after three models: sonnet NEG=5
(prior carries the register), Fable NEG=0 with grounding but density-fails the owner's ear,
Gemini ≈NEG=0-by-evasion with a grounding fault sonnet does not have.**

**Scanner note filed, not actioned:** the `do not <verb>` definitional form (G2's carrier) is an
EIGHTH candidate shape for `ScanDefineByNegation` — but the gate's vocabulary follows evidence
from production copy, and production runs on sonnet, whose tic is the comma form. If the writer
model ever changes, re-derive the shape list from that model's actual output first.

**Grok arm BLOCKED — owner action needed [MEASURED 2026-08-31, from the pod]:** both `XAI_API_KEY`
and `GROK_API_KEY` draw **403 Forbidden** on `https://api.x.ai/v1/responses` (the platform's own
endpoint, `feed_actions.go:745`) and on chat/completions. Discriminants: garbage key → 400
(unknown key), no auth + empty body → 422 (request parsed) — so the endpoint is reachable and the
keys are **recognised and refused**: a disabled or unfunded xAI account, not a wrong endpoint and
not an IP block. `llm_call_log` has **zero** Grok rows ever (`model_resolved ILIKE '%grok%'`) —
the xAI path has never run live (MDL-040's exact trap: a capability with no live caller has an
untested environment dependency). BusyBox wget drops the 4xx body, so the refusal reason itself
is unreadable from the pod. Only the owner can check/fund the xAI console.

---

> **CORRECTED 2026-08-31 (same night, by the owner's contradiction — "We use Grok daily for the
> news"):** the Grok block above stands in its conclusion and was WRONG in its instrument and
> incomplete in its diagnosis. (1) "`llm_call_log` has zero grok rows ever — the xAI path has
> never run live" measured the wrong table: `FetchLLMNewsAction` calls x.ai over raw HTTP and
> never writes `llm_call_log`. The right instrument (`orchestration_states`, provider=xai) shows
> the arm LIVE since 2026-08-30 14:55Z, 28 dispatches/day — and **zero items ever delivered**:
> every call draws the 403 whose body the Go client captured verbatim: *"Your team d443dd72-…
> has either used all available credits or reached its monthly spending limit."* So "disabled or
> unfunded" resolves to **out of credits / at the monthly cap**, and the conclusion (no
> successful Grok call in the platform's history) survives re-measurement at the right
> instrument — by luck, not by method (WRONG_CALLS row filed). (2) The discovery underneath:
> `fetchViaResponsesAPI` converts every API failure into an empty COMPLETED result nothing
> records and nothing reads — a total provider outage is indistinguishable from a quiet news
> day, masked by the RSS arm still delivering. Filed as **`bugs_open/418`** with fix candidates.
> (3) The Grok trial arm therefore stays blocked on ONE owner action: fund/raise the cap on xAI
> team `d443dd72-09cf-4ba7-8209-1395f0edb4f0`. The moment it funds, the runner recipe above works
> (the platform's own endpoint + `grok-4-1-fast`, or a stronger sibling if `/v1/models` lists one
> post-funding). (4) Probe practice from the same night: a multi-line Go raw-string literal
> false-absences a single-line grep — the truncation-trial literal read 0 in the fresh binary
> until the needle was cut at the line wrap ("SENTENCE BEFORE THE COMPARISON" ×1 present). Probe
> needles must be single-source-line substrings.

---

## Grok arms — RUN 2026-09-04 11:38–11:50Z (the block above is lifted)

**What changed:** the xAI team was funded by the owner on 2026-09-03 — `[MEASURED 2026-09-04]` the
news arm's first delivering run is `2026-09-03 15:06:22Z` (`orchestration_states`, provider=xai,
`write_items.written > 0`); 45 items over 10 runs since, zero 403s. So the one owner action the
correction above waited on has happened, and the arm ran on the same stored prompt as every other
arm. `grok-4-1-fast` (the news arm's recommended model) is **not** in `/v1/models` today; the list
carries `grok-4.20` (reasoning / non-reasoning), `grok-4.3`, `grok-4.5`, `grok-4.6` (newest,
2026-08-06) and the imagine/build families. Two arms were run: the strongest sibling, and the
cheapest non-reasoning one as the cost/latency counterpoint. n=2 each, as pre-registered.

**Instrument note, before the table.** The 08-31 rows quote sonnet at NEG=5 on "six shapes". Re-scoring
the same stored sonnet section TODAY with the production scanner (`datahelpers.ScanDefineByNegation`
at HEAD `763c8002f`) gives **8** (2 `x_not_y`, 5 `rather_than`, 1 `instead_of`) — the scanner has
gained shapes since 08-31, so the 5 and the 8 are two instruments over one text, not a change in the
text. Every number below is the production scanner over the model's flattened JSON output with tags
stripped, sonnet re-scored the same way in the same run; the 08-31 Fable/Gemini rows are NOT on this
instrument and their outputs were not stored, so they cannot be re-scored (fixed from today:
`TRIAL_OUTPUTS_2026-09-04_grok_arms_verbatim.md`). `ScanContrastNeighbours` = 0 on every arm.

| arm | NEG (prod scanner) | shapes | words | £99 / $5,000 | RAG / LoRA | links (all allowed?) | highlights | wall | xAI cost |
|---|---|---|---|---|---|---|---|---|---|
| sonnet (shipped, re-scored) | **8** | 2 x_not_y · 5 rather_than · 1 instead_of | 695 | **dropped / dropped** | kept | 3 (yes) | 4 | 37.5s | — |
| grok-4.6 #1 (reasoning, 18,873 reasoning tokens) | **0** | — | **1,015** | kept / dropped | kept | 13 (yes) | 3 | 315s | $0.140 |
| grok-4.6 #2 (reasoning, 14,420) | **0** | — | 627 | dropped / dropped | kept | 3 (yes) | 3 | 233s | $0.097 |
| grok-4.20 non-reasoning #1 | **4** | 3 x_not_y · 1 rather_than | 288 | dropped / dropped | kept | 1 (yes) | **[] (none)** | 5s | $0.011 |
| grok-4.20 non-reasoning #2 | **5** | 2 x_not_y · 2 rather_than · 1 negative_reveal | 340 | dropped / dropped | kept | 5 (yes) | **[] (none)** | 5s | $0.011 |

Words = flattened output, tags stripped, `split()` — sonnet on the identical measure. `plainly` = 0
and `honest*` = 0 on every arm (rule 19 held everywhere). Token counts: sonnet 12,954 in / 2,713
out; grok-4.6 8,905 in / 20,418 and 15,317 out (the reasoning is ~92% of that); non-reasoning
8,167 in / 440 and 489 out. Cost is xAI's own `cost_in_usd_ticks` (1 tick = 1e-10 USD; checked
against the list prices: grok-4.6 is $2/M in, $6/M out); sonnet's price is not asserted here.

**The read — grok-4.6 (the half that can veto; it does not).** Both runs reach zero by STATEMENT,
not evasion, and #1 is the longest output any arm has produced: *"That order matters because a
model aimed at a muddled process will still produce a muddle, and you will have paid for it."* ·
*"A review gate sits at the points that carry risk: a quote before it goes to a customer, a filing
before it is submitted, a summary before it goes out."* · *"A process with no owner, or with rules
that change every time someone asks, will give you a faster version of the same confusion."* (#2).
Benefit-led, non-presumptive (*"If that's your situation, pick one process and start a
conversation. We'll tell you when the work is a poor fit."*), simple sentences, and — the thing
Fable failed on — **it expands rather than compresses**: #1 at 1,015 words is 46% longer than the
shipped sonnet section and reads as explanation, not riddle. Ruling 13's density fault is not
present in #1; #2 (627 words) is sonnet-length. **Grounding, checked term by term against the
prompt** (Llama/Mistral/Phi, DoRA, GDPR, 10–250 employees, the sectors, Bulk Data Collection,
"run locally"/in-browser, review gates, quality sweeps, the blog titles — all present in the
prompt): no invented service, number, client or guarantee. **Two caveats, both illustrative rather
than factual:** #2's *"can be explained to an insurer, an auditor, or a client who issues a
questionnaire"* is a benefit scenario with no grounding in the prompt (the same CLASS as Gemini
G2's invention, milder in degree — no claimed outcome, but not in the brief); #1's gate examples
(quote / filing / summary) are illustrations the brief does not supply. **Facts:** the $5,000
market anchor is dropped by every arm INCLUDING the shipped sonnet section (the 08-31 table did
not record sonnet's fact column; it is recorded now); £99 survives in grok-4.6 #1 only. The
`highlights` field is populated on both (3 each), on-register.

**The read — grok-4.20 non-reasoning.** The sonnet prior, verbatim: *"Start with the work, not the
technology"* (an H3), *"They want concrete time savings and fewer errors, not science projects"*,
*"belong to you, not to a rented API"*, *"This is not a limitation of the technology. It is how…"*
(negative_reveal), *"standard parts of how we work, not add-ons"*. Plus an ungrounded
competitor-disparagement line (*"ChatGPT alone cannot reliably handle their documents … without
making things up"* — not in the brief), an EMPTY `highlights` array on both runs, and ~300 words
against sonnet's 695 — under-delivery of the Gemini kind. **≈ sonnet: expense decides nothing;
FAIL on the read as well.**

**Reading, per the pre-registration:** grok-4.6 is NEG ≤1 with a passing read on both runs →
**PROMOTE to a wider screen**, the second model to earn it after Fable, and the first that does
not density-fail. grok-4.20 non-reasoning does not promote.

**What the four arms now line up as `[INFERRED — n=2 per arm, one prompt; a screen, not a rate]`:**
the two arms that reason before writing (Fable, grok-4.6) both score zero; the two that do not
(sonnet as shipped, grok non-reasoning) score 4–8, and produce the SAME shapes. That is consistent
with the 08-31 Fable reading — the comparison rule was in the prompt and the difference is whether
the model obeys it — and it cuts across vendor. ~~If it is right, the cheapest fix is not a vendor
switch but **the production writer with extended thinking enabled**, which the production writer
does not use today (`[MEASURED 2026-09-04]` `thinking_tokens > 0` on 0 of the last 7 days'
`page-content-writer` rows — see the query in the RUNBOOK).~~

> **CORRECTED 2026-09-04 ~12:00Z, twenty minutes after it was written, by the API reference (the
> `claude-api` skill), not by a measurement.** Two things were wrong. (1) The premise: on
> `claude-sonnet-5` **adaptive thinking is ON when `thinking` is omitted**, and the production
> client omits it (`platform/aiservice/anthropic.go:278-300` only sends `thinking` when a
> `budget_tokens` option is configured) — so the shipped NEG=8 section was written WITH reasoning.
> (2) The instrument: `llm_call_log.thinking_tokens` is populated from parsed thinking-block text,
> and Sonnet 5 returns thinking blocks EMPTY by default (`display: "omitted"`), so that column
> reads NULL whether or not the model reasoned — the "0 of 6,724" measured the column, not the
> behaviour. The tell was already in the row: sonnet's `output_tokens` is 2,713 for ~1,100 tokens of
> visible JSON. The cheap check that would have caught it: read the API reference before asserting
> what a model does by default. Consequence for the inference above: **"reasons before writing"
> does NOT separate the arms** — sonnet reasons and still scores 8. What is left of the hypothesis
> is a DOSE question (Grok-4.6 spent 14–19k reasoning tokens; sonnet's default is adaptive at
> effort `high`), so arm 5 is re-specified below as a within-model dose test. Also found on the way:
> `anthropic.go`'s `budget_tokens` path is a **latent 400 on Sonnet 5** (the API rejects
> `{type:"enabled",budget_tokens}` on 4.7+ models) — see NOTES for whether any live config can reach it.

**Arm 5, re-specified (pre-registered before running):** `claude-sonnet-5` on the same prompt,
(a) `thinking: {type:"adaptive", display:"summarized"}` + `output_config.effort: "max"` ×2 — the
most reasoning the model will do; (b) `thinking: {type:"disabled"}` ×2 — none. Production sits
between them (adaptive, default effort). Readings stated in advance: (a) ≈8 ⇒ reasoning dose is not
the lever within sonnet, the prior is; (a) ≤1 ⇒ a config change on the existing writer is the
cheapest fix and a vendor switch is unnecessary; (b) ≫8 ⇒ the default reasoning is already doing
work and the count is a floor of the prior. Costs and latency stay the owner's call either way, and
the latency column matters — grok-4.6 spent 4–5 minutes per section against sonnet's 38 seconds.

**Runner recipe (works, key never leaves the pod):** `RUNBOOK_two_stage_copy.md` §"Offline model
replay". Scorer: same file, §"Score a replay with the production scanner".

## Arm 5 — RUN 2026-09-04 11:56–13:24Z: the reasoning-dose hypothesis is REFUTED within sonnet

Six calls, `claude-sonnet-5`, same stored prompt, three doses. Production omits `thinking`
entirely (`platform/aiservice/anthropic.go:278-300`), which on this model means **adaptive thinking
at default effort** — so the shipped section is the middle rung, and 5c reproduces it with
`display: "summarized"` so the reasoning is visible rather than inferred.

| arm | request | NEG (prod scanner) | words | stop | out tokens | thinking summary |
|---|---|---|---|---|---|---|
| 5b #1 | `thinking:{type:"disabled"}` | **9** | 593 | end_turn | 1,192 | — |
| 5b #2 | same | **6** | 626 | end_turn | 1,241 | — |
| 5c #1 | `adaptive` + `display:summarized` (= production) | **9** | 663 | end_turn | 3,169 | 2,334 chars |
| 5c #2 | same | **7** | 648 | end_turn | 1,856 | 761 chars |
| sonnet shipped 08-26 (re-scored) | production, implicit adaptive | **8** | 695 | — | 2,713 | not stored |
| 5a #1/#2 | `adaptive` + `effort:"max"`, `max_tokens` 16,000 | **NO TEXT** | 0 | **max_tokens** | 16,000 | 17.4k / 15.8k chars |
| 5a #1/#2 (retry) | same at `max_tokens` **40,000** | **NO TEXT** | 0 | **max_tokens** | 40,000 | — |

**Result 1 — the dose does not move the count.** Thinking off: 9 and 6. Thinking on at production's
dose: 9 and 7. The shipped section: 8. Every arm sits in the same 6–9 band, and 5c's summaries prove
the reasoning happened (2,334 and 761 characters of it, listing the very headings it then wrote).
**The pre-registered reading for this outcome was written before the run: "(a) ≈8 ⇒ reasoning dose is
not the lever within sonnet, the prior is." That is what came out.** So yesterday's cross-vendor
pattern is NOT explained by "models that reason before writing score zero" — sonnet reasons and
still scores 8. What separates grok-4.6 and Fable from sonnet is the prior, which is the conclusion
the 08-25 canary reached (P2a: the carrier is the MODEL PRIOR) and which this now survives a direct
within-model test rather than resting on a cross-vendor correlation.

**Result 2 — `effort: "max"` is UNUSABLE for this writer prompt, and it fails in the shape a battery
cannot see.** Four calls, two token budgets, four times `stop_reason: max_tokens` with the entire
budget spent on thinking and **zero characters of text**. Scored naively that is NEG=0 — a perfect
score, from an empty output, on the arm that would look most attractive in a table. This is
`bugs_open/012`'s rule arriving from the other side: `output_tokens == max_tokens` means CUT, not
finished. **Any future arm must assert `stop_reason == "end_turn"` and a non-zero word count before
a single count is read** (the RUNBOOK now says so, and the scorer prints stop and words first).

> **FOLLOW-UP 2026-09-04, prompted by the `bugs_open/257` lane: it is a SHAPE, not a size, and I had
> filed it as a size.** They pushed back on my calling result 2 a sizing question before anyone
> checked whether it reproduces below `max`. It does. `[MEASURED, n=8 at `effort: "xhigh"`,
> `max_tokens` 16,000]` **1 of 8 unusable**, and that one returned **3,488 characters of real text
> cut mid-JSON** rather than nothing — the more dangerous shape, because a lenient consumer can
> repair-and-persist a fragment. The seven that completed ran 5,636–15,057 output tokens, the best
> of them at **94% of the ceiling**. Production's default (`high`) was **0 of 6**, never above 23%
> of the same ceiling. ⚠ n=6 cannot license "high is safe" — it would miss a ~5% rate — so the
> claim is "no evidence at high", and what IS established is that headroom differs by an order of
> magnitude between two adjacent effort settings on one prompt. **Effort is not a quality dial that
> can be raised independently of `max_tokens`.** Landmine corrected the same day; the correction
> made it worse, which is the right direction for a correction to run.

**What this leaves for the owner's model question.** Within sonnet there is no config that fixes the
register: off, default and max all land at 6–9 or produce nothing. The lever is the model. grok-4.6
scored 0 twice with the facts kept and the length UP; Fable scored 0 twice but density-failed his
ear; Gemini reached ≈0 by evasion with invention. **On the evidence so far the shortlist is one
model long, and it is grok-4.6** — pending his read of the verbatim outputs
(`TRIAL_OUTPUTS_2026-09-04_grok_arms_verbatim.md`), the cost table above, and the latency (4–5 min
per section against sonnet's 38s, which matters for a fleet that renders many sections per build).

**Instrument corrections banked from this run** (both in `WRONG_CALLS.md`, checks in the RUNBOOK):
`llm_call_log.thinking_tokens` cannot see Sonnet 5 reasoning at all (blocks return empty by default,
so the column reads NULL while the model thinks); and a batch of Anthropic 400s I attributed to
BusyBox `wget` was the 11:21–11:57Z fleet credit outage — the same wget recipe returns 200 now.
