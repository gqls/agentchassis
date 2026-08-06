# DESIGN 2026-08-06 — the structural split of `classify_and_extract` (bugs_open/183 candidate 3)

**Status: DESIGNED, NOT BUILT.** Written because the owner asked what the split
actually is. Nothing here is implemented; two numbers must be measured before anyone
writes it (§6).

> **OWNER RULING 2026-08-06 — NOT NOW, accepted.** The owner read §8 and took the
> recommendation. The split is **not** rejected on merit; it is deferred because the
> class is monitored and not biting. **What makes deferral safe is LCO-007, so the
> deferral expires if the monitoring does** — if `fleet-step-token-pressure` is ever
> disabled, or if it flags `classify_and_extract`, this decision is void and the
> design is the answer. Do not re-litigate it on the strength of the cap alone: the
> cap has now been raised four times and each raise moved the cliff.

## 1. What the step does today

One `execute_llm_prompt` call, one ~17,000-character prompt template, one JSON document
with **four top-level sections**. It is step 6 of 15 and precedes all four
`write_*_spec` steps. Structure read from the live definition 2026-08-06:

| # | Section | What it holds | Rough weight |
|---|---|---|---|
| 1 | `identity` | company name, industry, sub-industry, services, contact, key people, USPs, target audience, competitors | small |
| 2 | `classification` | `site_type`, `category`, 4–10 `industry_tags`, `tone_suggestion`, `suggested_style`, `page_count_estimate` | small |
| 3 | `content_direction` | voice (5 fields), sentence style, persuasion, content depth, **≥8 `writing_rules`**, **≥6 `things_to_avoid`**, **≥6 `things_to_emulate`**, heading/paragraph/CTA style, terminology with 5–10 key terms, compliance rules | **largest by far** |
| 4 | `design_intent` | style direction, colour/typography mood, **8 mandatory `palette.reference_values` hex slots**, 2 typography stacks, avoid-list | large |

The failure that filed this bug: the document is emitted in that order, so a cut lands
in section 3 or 4 — and `design_intent.palette.reference_values` is **the field the
composition pipeline and design renderer actually read**. The most valuable part of the
document is the last thing written.

## 2. The dependencies — measured from the prompt, not assumed

> **CORRECTION, same day.** My first draft of this said "three sections read the
> `classification` the same call produces." **That is not what the prompt says**, and I
> wrote it into `bugs_open/183` before checking. Corrected here and there. What the
> prompt actually states is below. The near-miss is logged in NOTES — it is the exact
> failure this lane already logged twice, caught on the third try only because I opened
> the file.

Real couplings, quoting the live template:

- **`identity` is upstream of 3 and 4.** *"Every field in content_direction must be
  specific to THIS industry, not generic writing advice"* and *"derive the palette from
  the mission/adoption/research and the industry"*. Industry is an `identity` field.
- **`classification` ↔ `design_intent` share an enum.** `classification.suggested_style`
  and `design_intent.style_direction` are both
  `professional-dark|modern-light|bold-creative`, and the prompt requires
  *"style_direction must agree with the palette you emit"*. Two sections, one decision.
- **The adoption branch adds three more**, but they point at the **inputs**, not at this
  call's own output: `site_type`, `tone_suggestion` and `suggested_style` must each be
  consistent with the *adopted* archetype / content_direction / design_intent that were
  passed in. Those constraints survive any split unchanged.

**So the shape is: `identity` first, then fan out — with `classification` and
`design_intent` needing to agree on one field.** It is not four independent calls, and
anyone who builds it as four parallel calls will get a dark palette under a
`modern-light` classification.

## 3. The proposed decomposition

```
  classify_and_extract                    (identity + classification — small, cheap)
        |
        +-- write_identity_spec           (unchanged, now fed directly)
        +-- write_classification_spec     (unchanged, now fed directly)
        |
        +-- generate_content_direction    (NEW call: identity + classification in)
        |        \-- write_content_direction_spec
        |
        +-- generate_design_intent        (NEW call: identity + classification in,
                 \-- write_design_intent_spec        style_direction pinned by
                                                     classification.suggested_style)
```

Three generations instead of one. The first is small and nearly incompressible. The two
that follow are each bounded by a single document, run against a **short** prompt (the
17k template splits — each child carries only the rules for its own section plus the
identity/classification result), and can run concurrently.

## 4. Why this is worth doing — and the reason is NOT the smaller cap

The cap is a red herring, and the platform says so in its own words
(`platform/aiservice/truncation.go:26-29`): *"whatever the number, the step that writes
most approaches it on the work most worth doing."* We have now raised this cap four
times (6000 → 16000 → 32000 → 64000). Each raise moved the cliff.

**The real property the split buys is that failure stops being all-or-nothing.**

| | today | after the split |
|---|---|---|
| One section overruns | **all four specs lost**, site unclassified, `max_attempts` burned | that one section fails; the other three are already written |
| Retry cost | redo the entire classification and all four documents | regenerate one document |
| A bad palette | rerun everything | rerun `generate_design_intent` alone |
| Blast radius of a prompt edit | every section | the section whose rules you edited |

That last row is the quiet win. Today, changing the `writing_rules` floor from 8 to 10
re-risks the palette, because they are the same completion.

## 5. Costs, stated honestly

- **Three calls instead of one.** More total input tokens (identity/classification is
  re-sent to both children), more latency unless the two children are run concurrently,
  more `llm_call_log` rows. Probably *cheaper* overall than today's retry behaviour —
  today a truncation re-runs the whole document up to 3 times — but that is an
  [UNMEASURED] claim until §6 is done.
- **A new consistency surface.** `suggested_style` must reach
  `generate_design_intent` and be honoured. Today one completion keeps them consistent
  for free; afterwards it is a contract that can drift. **This is the part to get right**
  — pass it as an input field and state it as a constraint in the child prompt, do not
  hope.
- **Three prompt templates to maintain instead of one**, with the shared build-standard
  and mission preamble duplicated or factored out.
- **Partial spec sets become representable.** A site can now sit with 3 of 4 specs. Every
  downstream reader that assumes "specs exist ⇒ all four exist" needs checking. That is
  the RFC trigger, not the token count.

## 6. Two things to measure BEFORE writing any of it

Both are cheap, and the design is not decidable without them:

1. **Per-section token weight.** Take ~20 successful `classify_and_extract` completions
   from `llm_call_log.response_text`, parse the JSON, and measure the serialized size of
   each of the four sections. The whole design assumes `content_direction` and
   `design_intent` dominate. **If they do not — if the weight is spread evenly — the
   split buys much less** and the honest answer is to keep the raised cap and the
   monitor. Do not skip this because the schema *looks* lopsided; the schema is not the
   output.
2. **Whether a smaller unit actually reduces truncation risk, or just moves it.** The
   `content_direction` child, alone, still has floors of ≥8 + ≥6 + ≥6 list entries plus
   ten prose sub-objects. It is entirely possible that section 3 alone approaches a
   16000 cap on a verbose industry — in which case the split has relocated the problem,
   not removed it, and `content_direction` needs its own decomposition or a shorter
   spec. Measure it from the same 20 completions.

## 7. Governance

**This needs an RFC** under the OWNER RULING of 2026-07-29 §1: the test is not "is it
additive" but "does it change what the shared mechanism GUARANTEES". It does — the four
`write_*_spec` steps today cannot observe a partial spec set, and afterwards they can.
That is a guarantee change on a seam other pipelines read, which is exactly the
architecture-scope case.

Consumers to **tell, not merely measure** (2026-07-29 §3): the adoption lane, the
composition pipeline (`site-design-planner` reads `category`/`industry_tags`), and the
design renderer (reads `palette.reference_values`). The useful message to them is what
changed about their guarantee — "you may now see 3 of 4 specs" — not a list of new steps.

## 8. Recommendation

**Not now, and the monitoring is what makes waiting safe.** With the cap at 64000 the
observed maximum (6,590) has ~9.7× headroom, and `fleet-step-token-pressure` (LCO-007)
now announces regrowth toward the ceiling instead of leaving it to be discovered by a
burned site. Restructuring a live 15-step agent that adoption lanes are actively using,
to fix a class that is currently monitored and not biting, is the wrong order.

**The trigger to build it** is written into the bug: if LCO-007 ever flags
`classify_and_extract` again, the cap is no longer the answer and this is.
