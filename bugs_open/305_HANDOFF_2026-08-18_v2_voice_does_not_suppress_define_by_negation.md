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

## 4. ⚠ REFUTED 2026-08-18 by the diagnosis loop AND by a direct check — the licensing hypothesis is DEAD

**`090` verdict (`57b2dcd2`): `UNVERIFIABLE` — "stopped: scope-not-narrowing", no fix
proposed, "hand to a human with the full trail; do NOT auto-conclude."** A refuted
hypothesis at the cost of one run is the cheapest place to be wrong, and this one refuted §4
below from a path I had not taken.

**What the loop established (its own evidence, independent of mine):** the exact rhetorical
shape was already being produced for the **model-directory page on 2026-08-06** — a week
before the v2 flip — in the wording *"actually running in our production deployments, not a
catalogue of what's theoretically possible"* and *"deployed and doing work, not sitting in a
demo repo"*. It reached that by finding a different generation than the 08-08 one I found.
**So v2's wording cannot be the mechanism that introduced the construction.**

**The gap it named, and I have now closed it.** It asked: *"it would also help to know what
`voice_style_block` text (if any) was actually active on 2026-08-06 — the current
data_request only returns the present (post-flip) config text, not a version history"*. There
IS a version history, in a backup **this lane created when it shipped v2** (CQ-022's rollback
list): `agent_default_configs_bak_20260813_voicecarrier`. `[MEASURED 2026-08-18]`

| voice block | contains "contrasting pair" | chars |
|---|---|---|
| pre-v2 (backup taken 2026-08-13, pre-update) | **NO** | 2,499 |
| current v2 | yes | 6,032 |

**The construction was being produced under a block that never mentions it.** Naming it
cannot be what licensed it. §4's hypothesis is dead, and dead by two independent routes.

⚠ *Provenance of the pre-v2 text:* the live row was created 2026-07-27 and carries exactly
one recorded update (2026-08-13, the flip), and the backup was taken immediately before that
update — so absent an **unrecorded** intermediate edit (`updated_at` shows only the last one)
that text was live on 08-06. Corroborated by CQ-022's own history: built 07-27, one consumer,
untouched until the flip.

**What is still open** — the loop was explicit that it could not close the original symptom:
the same before/after comparison for **`adoption-tracker` and `protocol-tracker`** (this
bundle had no generation history for either; symbol bodies came back "unavailable" as a
tooling failure, and its own data request died with `canceling statement due to statement
timeout`).

⚠ **Possible confound I caused:** I was running an expensive correlated query over
`llm_call_log` × `orchestration_states` on the same database during that window (cancelled
after ~7 minutes). I cannot show it caused the loop's statement timeout, and the message is a
`statement_timeout` rather than my cancellation — but **do not run heavy exploratory queries
while a diagnosis run is in flight**, which is a lesson regardless of whether it bit here.

## 4b. The original hypothesis, kept as the record of what was refuted

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


---

# ROOT CAUSE FOUND 2026-08-19 — the brief is written in the construction, and one of its sentences is the owner's own complaint, verbatim

**This closes the gap the loop named** (*"the same before/after-2026-08-13 comparison for those
two pages"*) **and supplies a cause the loop did not reach.** First-hand verification, stated
plainly per the 2026-07-31 ruling: the evidence below is literal string matching between a
spec, a rendered prompt and a stored output — not a rate, not an inference, and not the
correlation this file already had to withdraw.

## A. All three flagged pages are pre-v2. The symptom contains no post-v2 output at all.

Dated by searching `llm_call_log.response_text` for each page's own distinctive sentence
`[MEASURED 2026-08-19]`:

| page | component | copy written |
|---|---|---|
| `model-directory` | call-to-action | **2026-08-08** (and the loop found the same shape on 08-06) |
| `adoption-tracker` | call-to-action | **2026-07-26 15:50** |
| `protocol-tracker` | call-to-action | **2026-07-26 15:33** |
| `adoption-tracker` | hero ("in days, not months") | **first seen 2026-04-10**, in **251** calls |

All predate the v2 carrier (08-13). **The owner's complaint is entirely about pre-v2 copy**,
which is why every attempt to explain it through v2's wording failed — including mine.

## B. The cause: `content_direction` is itself written in the construction, and it hands down the tagline verbatim

The site's current `content_direction` spec (`is_current`, created 2026-07-24) — the page
brief the writer reads — carries the shape **seven times in one document**:

- *"Use concrete stack references (Kubernetes, Kafka, Postgres) naturally, **not as buzzwords**"*
- *"…thinking 'these people actually run this in production' **rather than** 'these people have read about this tech'"*
- *"CTAs should initiate a technical conversation, **not a sales process**"*
- *"Frame the next step as an engineering discussion, **not a purchase decision**"*
- *"verbs that imply collaborative technical engagement **rather than** transaction"*
- *"Headings should make a claim or describe a condition, **not just label a section**"*
- and the **canonical tagline it supplies**: *"Multi-agent systems deployed to production **in days, not months** — on Kubernetes, Kafka, and Postgres"*

**The causal chain is literal, not statistical** `[MEASURED 2026-08-19]` — across all 21,078
`page-content-writer` calls, `in days, not months` appears in **1,348 rendered prompts** and
**408 responses**. The brief hands the writer the phrase; the writer uses it. That is the
hero sentence the owner read.

**Every one of the site's five writer-visible specs carries the shape** (`identity`,
`strategy`, `content_direction`, `audience`, `briefing` — all match `,\s+not\s+`).

## C. What this corrects, in both directions

- **`portfolio_positioning` were right that the 08-12 root cause does not fit — and right for
  a reason that turns out to be beside the point.** They checked `identity.key_differentiators`
  and found them *positive in sentiment* ("Fast deployments (minutes instead of weeks)"). True
  — **and that differentiator is itself a contrast.** Positive in CONTENT is not free of the
  CONSTRUCTION, and it is the construction the owner objected to.
- **So the lane's 08-12 finding survives in corrected form.** It said the negativity comes from
  a negatively-worded differentiator. The general statement is stronger and simpler: **the
  writer reproduces the RHETORICAL FORM of its brief, independent of the brief's sentiment.**
  This is the estate's own measured principle — *the example is the instruction; the rule is
  commentary* — arriving from the input side: here the instructions ARE the examples.
- **And it explains why no voice-block change could have fixed it.** The fleet carrier is a
  general instruction; the brief is a specific, repeated, site-scoped exemplar carrying the
  literal sentence. `305 §4`'s licensing hypothesis was not merely unproven, it was looking at
  the wrong document.

## D. What follows for the fix — and it is NOT a writer change

The owner gave this lane both halves (*"ensure that that sort of copy never leaves this
framework again"* + fix the pages). On this evidence:

1. **The durable half is the SPEC, not the writer.** `content_direction` for this site should
   be rewritten out of the construction — including the canonical tagline, which is the single
   highest-leverage string on the site (1,348 prompts). ⚠ This is a **site-config** change on
   another lane's site: coordinate, do not unilaterally edit.
2. **A detector is expressible and cheap:** flag a `content_direction`/`identity` spec whose
   own text carries the shape more than once or twice. It is the same count
   `count_negation_tells.py` already does, pointed at the SPEC instead of the page — and it
   catches the fault at the place that causes it, before a single page is written.
3. **The three pages are cleanup and stay behind (1).** Re-rendering them against an unchanged
   brief regenerates the same register from the same source.
4. ⚠ **This is one site.** Whether other sites' briefs carry the shape is a one-query census
   nobody has run — and given this file's own history with over-general claims, **run it before
   asserting a fleet pattern.**
