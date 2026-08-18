# 305 — the v2 house voice does NOT suppress define-by-negation, and the rate did not fall when it shipped

**Filed 2026-08-18** by `copy_quality_two_stage`, from an owner report routed in by
`portfolio_positioning` (their CONTRIB of the same date, in this lane's directory).

> ## ⚠ ROOT CAUSE IS NOT ASSERTED HERE — it is with the diagnosis loop
> `090` run filed 2026-08-18, **`RUN_CORRELATION_ID=57b2dcd2-2ded-473c-9f2e-617176f39c15`**.
> Per the owner ruling of 2026-07-31, a `bugs_open/` file making a cross-cutting or
> structural claim is not "filed" until it has been through that loop. What is below is
> **measurement and location only**. The licensing hypothesis in §4 is a HYPOTHESIS and is
> labelled as one; do not quote it as a cause.

## 1. What the owner said

He reviewed three live directory pages on `ai-agent-orchestration.com` — `model-directory`,
`adoption-tracker`, `protocol-tracker` — and said the copy *"looks like it didn't go through
the framework"*. His verbatim examples:

> *"The registry shows you what's possible, not what survives production."*
> *"…tells you which agents exist. It doesn't tell you how they…"*

**He gave the `copy_quality_two_stage` lane both halves**: *"ensure that that sort of copy
never leaves this framework again"*, and fixing the affected pages.

## 2. It DID go through the framework, and the self-undermining shape is the point

`model-directory`'s `call-to-action` `content_data` `[MEASURED 2026-08-18]`:

```
headline:    "170+ agents defined, 174 agent types. The registry shows you what's
              possible, not what survives production."
subheadline: "A model directory tells you which agents exist. It doesn't tell you how they
              hold up under real Kafka throughput…"
```

A **model-directory page** whose call-to-action tells the reader the directory does not tell
them the important thing. The `hero` on the same page closes *"…not a vendor roadmap"*. The
directory **listing** components are fine — they render from the cited register and read
plainly. It is the surrounding `hero` and `call-to-action` copy, which is LLM-authored.

## 3. ⚠ CORRECTED 2026-08-18, SAME SESSION — MY CENTRAL MEASUREMENT DOES NOT SUPPORT THIS FILE'S TITLE

> **The rate comparison below cannot answer the question it was built to answer, and I
> published it before checking that.** Read this block before §3.
>
> I reported pre-v2 **2.72** → post-v2 **2.85** hits per 1,000 words and called it "unchanged
> to slightly worse". Then I controlled for date drift by comparing ADJACENT equal-length
> windows (08-07..08-12 vs 08-13..08-18) and got **4.35 → 2.85** — a 34% FALL, i.e. the
> opposite conclusion from the same table.
>
> Both are artefacts. The weekly series for the same agent `[MEASURED]`:
>
> | week | 06-15 | 06-22 | 06-29 | 07-06 | 07-13 | 07-20 | 07-27 | 08-03 | 08-10 | 08-17 |
> |---|---|---|---|---|---|---|---|---|---|---|
> | per 1,000 words | 4.27 | 1.86 | 2.85 | 2.89 | 3.11 | 4.23 | **0.38** | 2.94 | 4.08 | 2.92 |
>
> **Week-to-week variance runs 0.38 to 4.27 with no trend.** An effect the size of either
> claim is invisible against that. So the honest statement is: **this method cannot detect
> whether v2 changed the rate**, and neither "it did not reduce it" (the title) nor "it
> reduced it 34%" is supportable. The all-history baseline was low because it averaged
> months of different sites, page types and content lengths; the adjacent window was high
> because the preceding week happened to be high.
>
> **What survives, and it is what actually matters:** the writer demonstrably still produces
> the construction — 2.85 per 1,000 words is not zero, and the owner objected to live output
> from it. §5's urgency does not depend on the comparison at all.
>
> **What a real measurement needs:** hold site AND page-type constant (the mix moves weekly),
> or measure per-PAGE rather than per-call, and state the effect size the sample could
> detect before running it. A within-site pre/post query is running and is the right shape;
> it was too slow to finish inline.
>
> **The title of this file overstates its own evidence and is left standing deliberately**,
> with this correction directly beneath it, because renaming it would hide that the error
> was made. Cause remains with the loop (`57b2dcd2`) — ⚠ and note the `090` symptom
> statement was authored FROM the refuted framing, so read its verdict against this block.

## 3. The measurement as originally filed (superseded by the block above)

**First finding — the copy is OLDER than the symptom's timestamp.** The producing call is
`llm_call_log` `31a81e3c-f8e0-4783-b213-8e089779f564`, `page-content-writer`, step
`process_sections_loop_iter_2_generate_content`, **2026-08-08 08:51:33Z** — found by
searching `response_text` for the owner's own sentence. The components' `updated_at` of
2026-08-17 20:08Z is a **re-render**, not a fresh generation. So this copy predates both the
identity-spec fix (08-12) and the v2 carrier (08-13), and `portfolio_positioning`'s framing
of "five days after the identity-spec fix" describes when it was RE-RENDERED, not when it was
written.

**That would make it a fossil — except the writer has not stopped.** Same agent, same table,
split at the 2026-08-13 v2 flip `[MEASURED 2026-08-18]`:

| era | calls | mean words | `X, not Y` per 1,000 words |
|---|---|---|---|
| pre-v2 (before 08-13) | 19,651 | 222 | **2.72** |
| post-v2 (08-13 onward) | 1,338 | 223 | **2.85** |

Method: `agent_type='page-content-writer' AND success`, regex `,\s+not\s+\w` over
`response_text`, normalised by word count. **Mean response length is identical (222 vs 223),
so the rate is not a length artefact.** By presence-per-call the gap is wider (33.6% → 41.6%)
but that figure is the less defensible of the two; the normalised rate is the one to quote.

**So the v2 house voice did not reduce this construction.** It is unchanged to slightly
worse. That is the finding; it is not yet an explanation.

## 4. HYPOTHESIS ONLY — the instruction names the construction

v2's own text discusses when *a matched contrasting pair is earned*, i.e. the shape is named
inside the instruction meant to discourage it, and this estate has a measured precedent for
prompt text scoring as the behaviour it describes. `voicetells.go` independently codifies the
same shape as a tell (`strawmanCommaRe`, the defining-by-negation check). **Not tested.** The
diagnosis loop has the question; the obvious fix may be the wrong one, which is exactly why
`portfolio_positioning` did not attempt it and why I have not edited the block.

## 5. Why this is urgent rather than cosmetic

The directory planner now plans `hero → listing → call-to-action` for **every site that opts
into a directory kind** (migrations `433`/`441`, live and council-approved, register
`DIR-001`), and the fleet plan is ~140 domains. Every one draws `hero` and `call-to-action`
copy from the path measured above. **Fixing the three existing pages without fixing the
writer reproduces it N times.**

## 6. What has NOT been done, deliberately

- **No edit to the voice block or any writer prompt.** The cause is with the loop.
- **No rerender of the three pages.** They are `ai-agent-orchestration.com`'s — another
  lane's site, with a live session. The evidence is intact on purpose: a rerender destroys
  the before-image, and `portfolio_positioning` preserved it for the same reason.
- ⚠ **Separate defect noticed in passing, not this bug:** `model-directory`'s
  `cta_url` and `primary_cta_url` both point at `/tools/password-entropy.html` while the
  button reads *"Book a Technical Discovery Call"*. That site carries 27 open
  `cta_names_unknown_destination` items. Belongs to whoever owns that site's CTA resolution.

## 7. How to verify a fix

Re-run §3's split with the cut at the fix date rather than 08-13, on
`page-content-writer` only, normalised per 1,000 words, and require the rate to FALL. ⚠ A
rate that falls on a *different* agent population proves nothing — hold the `agent_type` and
the word-count normalisation fixed, or the comparison drifts.
