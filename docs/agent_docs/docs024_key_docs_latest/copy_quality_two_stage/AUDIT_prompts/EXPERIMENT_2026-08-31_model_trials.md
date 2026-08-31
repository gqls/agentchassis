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
