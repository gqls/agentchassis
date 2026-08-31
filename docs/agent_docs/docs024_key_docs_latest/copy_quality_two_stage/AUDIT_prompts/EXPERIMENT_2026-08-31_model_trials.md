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
