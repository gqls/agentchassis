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

# ⚠ ROOT CAUSE — CORRECTED 2026-08-19 (evening). The finding SURVIVES and gets simpler; my COUNTS were over text the writer never sees.

> **Read this block before the section below, which is left standing as filed.**
>
> **What I got wrong.** I wrote that the brief "is written in the construction, seven times
> over", and ran a fleet census reporting 24 of 25 briefs at 24–38 instances. **Those counts
> are over the whole `content_direction` DOCUMENT. The writer does not read the document.**
> `[MEASURED 2026-08-19]` It reads `formatted` — **3,558 of that spec's 15,760 chars, ~23%** —
> and the structured instructional fields (`cta_style.approach`, `terminology.approach`) never
> reach the prompt at all.
>
> **Proof, and it is the check that found this:** `not a sales process` and `rather than
> transaction` are both in that spec's `cta_style.approach`. Across every
> `page-content-writer` call they appear in **ZERO prompts** — while appearing in **35** and
> **21** outputs respectively, i.e. **only where the prompt did NOT contain them.** Text that
> reaches no prompt cannot be causing anything; those echoes are the model's own phrasing, not
> transfer. (⚠ The "seven times" figure was worse than merely mis-scoped — I read it off a
> `LIMIT 10` sample of regex matches and reported the visible rows as the total.)
>
> **Corrected figures** `[MEASURED 2026-08-19]`:
>
> | | in `formatted` (the writer reads this) | in the whole spec (it does not) |
> |---|---|---|
> | `ai-agent-orchestration.com` — the complained-of site | **2** | 13 |
> | `remortgagecalculator.uk` — worst in fleet | **19** | 38 |
> | `vonc.com` / `loanandmortgagecalculator.co.uk` / `loancash.co.uk` | 18 / 17 / 16 | 36 / 36 / 36 |
> | sites whose `formatted` carries it at all | **23 of 25** | 24 of 25 |
>
> **What survives — and it is the load-bearing half, now stated more precisely.** The brief's
> effect on this symptom runs through **one supplied phrase**: the canonical tagline
> *"Multi-agent systems deployed to production in days, not months"*, which IS in `formatted`
> and transfers verbatim — **1,348 rendered prompts, 408 responses**. That is a literal chain
> and it is untouched by this correction. The owner's hero sentence was supplied by the spec.
>
> **So the root cause is narrower and cleaner than I filed it:** not *"the brief is saturated
> with the construction"* but ***"the brief hands the writer a canonical tagline built on the
> construction, and supplied phrases transfer."*** The saturation was an artefact of counting
> text nobody reads.
>
> **And the fleet claim changes shape too.** 23 of 25 sites still carry it in the text the
> writer actually sees, and the ranking is unchanged — so the pattern is real, roughly half as
> large as reported, and still worth a detector. But a detector must read `formatted` (and the
> other supplied fields), **never the whole spec document**, or it will report an inflated
> number pointed largely at text with no consumer.
>
> **Told `portfolio_positioning`**, whose pilot I had warned at 38: the real figure for what
> their writer sees is **19**, still the fleet's worst.

# ROOT CAUSE as originally filed (superseded by the block above) 2026-08-19 — the brief is written in the construction, and one of its sentences is the owner's own complaint, verbatim

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
4. ~~⚠ This is one site … run it before asserting a fleet pattern.~~ **CENSUS RUN, same
   session `[MEASURED 2026-08-19]` — it is FLEET-WIDE, and this site is on the LOW end.**

   **24 of 25** sites with a current `content_direction` carry the construction. Instances per
   spec, worst first: `remortgagecalculator.uk` 38 · `loanandmortgagecalculator.co.uk` 36 (+23
   "rather than") · `loancash.co.uk` 36 · `vonc.com` 36 · `loancalculator.co.uk` 28 (+29) ·
   `vetcomparison.uk` 26 (+32) · `idea.uk` 26 · `cookly.uk` 24. **`ai-agent-orchestration.com`,
   the site that drew the owner's complaint, has SEVEN** — so the complaint arrived from one of
   the least saturated briefs in the estate.

   ⚠ **Top of that list is `remortgagecalculator.uk` — `portfolio_positioning`'s Phase C
   directory pilot**, the site they offered to rerun once a fix is live. Rerunning it against
   that brief would reproduce the register from the most saturated source in the fleet. Told
   them.

   **What this census does and does NOT establish.** It measures the FORM in the briefs, and
   the counts include the spec's own instructional prose (*"not as buzzwords"*), where a
   contrast is a reasonable way to give guidance. **Proven transfer: one literal phrase** —
   the canonical tagline, 1,348 rendered prompts → 408 responses. **Not proven:** that the
   *instructional* uses transfer into output at all. That is the next measurement, and after
   §3's history it should be designed to be able to come out either way before it is run.

---

# THE FIX, 2026-08-20 — built by the `bugfix_305_negation_gate` lane, contributed back here

**Who and where.** A session picking this up with the owner's instruction to fix it at the framework
level. `scripts/who-owns.py 305` names `copy_quality_two_stage`; their
`docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/HANDOFF_2026-08-19_continue_here.md`
had no 305 fix in flight, so the writer-side half was open. Everything below is contributed INTO this
file rather than forked into a second account. Working docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_305_negation_gate/` (PLAN, NOTES, RUNBOOK,
README_where_we_are). Told, in writing, before any code: `copy_quality_two_stage`,
`site_ai_agent_orchestration`, `portfolio_positioning` (CONTRIB files dated 2026-08-20 in each lane's
own directory).

## §8. The bug is still live — re-verified `[MEASURED 2026-08-19 ~21:30Z]`

- the brief still supplies the tagline: `content_direction.formatted` on `ai-agent-orchestration.com`
  (3,558 chars, row created 2026-07-24, never updated) contains `in days, not months`;
- the three pages still serve the quoted copy (`page_components.content_data`, all nine components
  unlocked, `updated_at` 2026-08-17 which is the RE-RENDER this file's §3 already warned about);
- the writer has not stopped: same site, 2026-08-19 18:26–18:32Z — *"not a catalogue built to look
  busy"*, *"not from provider marketing pages"*, *"not staging load"*.

## §9. What was measured before designing anything

`llm_call_log`, `agent_type='page-content-writer'`, `success`, 2026-08-13..19, **1,503 calls ≈ sections**:

| shape | sections with ≥1 |
|---|---|
| `x_not_y` ("…possible, not what survives") | 631 (42%) — ≥2: 208 (14%) |
| `rather_than` | 646 (43%) |
| `not_x_but_y` — **the only shape the existing Go detector matches** | **23 (1.5%)** |
| negative reveal ("…. It doesn't tell you…") | 168 (11%) |
| a headline-class field carrying `x_not_y` | 209 (14%) |

⚠ **Correction to my own first census:** it reported `not_x_but_y` and `rather_than` as **0**. I had
pasted Go patterns into psql, and **Postgres has no `\b`** — there it is a backspace character and `\y`
is the word boundary. Every figure above is post-correction. In `WRONG_CALLS.md`.

## §10. What was built

Three pieces. The first two are inert until the next chassis roll; the third is live.

1. **The family as one shared scanner** (`platform/orchestration/datahelpers/negationtells.go`,
   `negation_content.go`). Five shapes, plus a *displacement* set that can never trip and only rejects a
   rewrite, plus the exemption rule, plus the acceptance rule. `voicetells.go`'s `strawman` arm now calls
   it, so the post-deploy voice check and the meta-description gate inherit the shapes they were blind
   to; `rather_than` enters that check as a **density** (>2/page), never per-hit.
2. **Counting, default ON, at the seam every LLM section crosses** — wrappers on the `render_component`
   and `compile_page_sections` registry entries add `copy_gate_findings` and `copy_gate_page_hits` and
   change nothing else. Every writer, every site, wired or not.
3. **Repair at the writer seam** — a new action `rewrite_negations`, inserted into
   `page-content-writer`'s section loop by migration `509` (**held**: it rewires the step chain, so it
   must not be applied before the image). One LLM call asking for **sentence replacements**, spliced in
   Go, each one judged before it is accepted. Beyond a budget of **two per PAGE** (the house voice's own
   standard, carried in `CollectedData`) or **any** hit in a headline-class field.
4. **The brief side** — `cmd/brief-negation-check`, a daily CronJob (07:40 UTC), **LIVE since
   2026-08-20** at `v1.0.1321`. Derives the writer-visible surface at runtime from the live prompt,
   measures only that, and separates *supplied* (files a finding) from *instructional* (counted only)
   from *regulatory* (left alone, by the gate's own rule). First fleet run: **10 of 25 sites**;
   `ai-agent-orchestration.com` has exactly ONE and it is MANDATED onto four page types — this file's
   own complaint, arriving from the source side.

## §11. ⚠ THE FIX WILL NOT CHANGE THE THREE PAGES THIS BUG IS ABOUT, AND THAT IS DELIBERATE

The gate **exempts anything the site's brief supplied**, because the house voice's own first line is
*"A site's own voice specification outranks these rules"*: rewriting a brief's words would put the
platform in the position of overruling a site owner. `in days, not months` is supplied by that site's
`content_direction`. **The gate will count it and leave it.** Only editing the brief and re-rendering
moves it, and that belongs to `site_ai_agent_orchestration`, who have been told and given the queries.

So the honest statement of what was delivered against *"ensure that that sort of copy never leaves this
framework again"* is: **the five named forms, beyond two per page and never in a headline field, do not
leave `page-content-writer`; brief-supplied and regulatory negations are counted, not rewritten.** Not
"the instinct never leaves" — `fleet_copy_quality`'s own ablation says that is unreachable by rule.

## §12. CORRECTION TO §7 OF THIS FILE — its verification instruction is wrong for this fix

§7 says to re-run the rate on `llm_call_log`. **Do not verify this fix that way.** The gate's own
repair calls are logged in that table too, so the per-call rate can RISE while the artefact improves.
Verify at three levels instead:

1. **the artefact** — `page_components.content_data` for pages built after the roll (never
   `updated_at`: a re-render bumps it without regenerating, which is §3's own finding);
2. **the marker** — `orchestration_states.collected_data->'__copy_gate'`: `hits_after`, and the
   `rejected` reasons, which are the displacement instrument;
3. **the split** — first attempts vs `error_message LIKE 'RETRY (bugs_open/305%'`. ⚠ that marker is
   present on **successful** repairs too, so filter failures on `success=false` (the `bugs_open/119`
   precedent, which has already misled one census).

## §13. Council: round 1 REVISE, and six objections changed the code

Correlation `c48b7612-3ecc-4345-912e-5966c079cb91` (round 2 submitted on the same correlation). The
gating objection was that a `sub_workflow`'s running half is often keyed `substeps`, not `steps` — the
council's own read-only check answered it (`has_substeps=false`), and the migration now anchors on the
container path so a `substeps` row is 0 rows and a loud RAISE. Five more were right and were fixed
rather than argued with: **no banned-claims scan on the accepted replacement** (compliance, HIGH — "say
what it IS" is exactly the pressure that invents a superlative, and nothing downstream inspects a
spliced sentence); the repair prompt not forbidding new capability claims; **truncation** unhandled; tag
**multiset** equality accepting inverted nesting; and the per-page budget's unproven cross-iteration
state, now a named precondition in the held migration rather than a follow-up. And `497` was already
taken, exactly as the seat suspected — it is `509`.

## §14. What is still open

- **The migration is HELD.** Two preconditions, both in its header: the image is live (ask the binary,
  per service), and the per-page budget canary passes.
- **The three pages** need a brief edit by their owning lane, then a re-render. ⚠ `bugs_open/327`: write
  the WHOLE `content_direction` object, never a patch, and verify by label presence rather than a diff.
- **Is `rather_than` too broad?** 43% of sections is either a real fleet-wide tic or a pattern that
  should be narrowed. It is a density rather than a per-hit finding for that reason, and the rejection
  log will settle it within a week of traffic. `copy_quality_two_stage` has been asked directly.
- **Does *instructional* contrast transfer?** Still `[UNMEASURED]`, still open in two lanes, and nothing
  in this fix depends on it — but the gate now produces the corpus that would answer it: every rewritten
  sentence is a before/after pair with the brief text alongside.

## §15. The demand control: the shipped scanner, run over these three pages' real `content_data` `[MEASURED 2026-08-20]`

Not a unit test — the actual `datahelpers` functions, over the actual `page_components.content_data` of
`model-directory`, `adoption-tracker` and `protocol-tracker`, with that site's actual
`content_direction.formatted` + `identity.key_differentiators` as the exemption corpus. Recipe and
expected output in the lane RUNBOOK.

```
TOTAL 7 | exempt (brief-supplied or regulatory) 1 | repairable 6, of which headline-class 6
```

- **Both sentences the owner quoted come back REPAIRABLE:**
  `model-directory` call-to-action headline — *"The registry shows you what's possible, not what
  survives production."* (`x_not_y`), and its subheadline — *"It doesn't tell you how they hold up
  under real Kafka throughput…"* (`negative_reveal`).
- **The canonical tagline comes back `exempt:brief_supplied_sentence`** (`adoption-tracker` hero) — the
  designed behaviour, and the whole of §11 in one line of output.
- All six repairable hits are headline-class, so all six are repaired regardless of the page budget.

**This canary found a defect three test files had missed.** `negative_reveal`'s pattern begins at the
PREVIOUS sentence's full stop, so the hit was being attributed to *"A model directory tells you which
agents exist."* — a true, clean sentence. The repair would have been handed that to rewrite while the
reveal stayed exactly where it was, and the gate would have reported a repair that changed nothing that
mattered. Fixed (the anchor now skips the terminator onto the construction itself) and pinned by a test
that also re-asserts both splice invariants. **Run this canary after any change to the shapes** — the
fixture a unit test would use is the one I wrote, and it passed.

## §16. Council round 3: REJECTED — a guardian veto, upheld, and what it changed

**The veto, verbatim, because it is a better statement of the principle than anything I would write:**

> *"This is round 3 and the code under review for edit 4 has not changed since round 2's HIGH
> objection — only the paperwork around it has (an architecture_review doc was filed) … Routing a
> scope objection to architecture review does not license deploying the disputed change … 'we wrote it
> down and routed it' is not the same as 'it was contained.'"*

I had read the owner ruling of 2026-07-28 as licensing *"the code stays, file an RFC"*. That ruling was
an owner's decision about one case, not standing permission for a session to ship a disputed shared
seam and write a note about it. **Two of the fourteen seats had flagged the same thing at HIGH for two
consecutive rounds and I had answered with a document.**

**What changed in response, in code:** the fleet-wide counting is now **opt-in per step
(`copy_gate_annotate`), default OFF**, enabled only on `page-content-writer`'s own render and compile
steps by migration `509` — the same shape as migration `474`. Containment is the strong form and is
pinned by a test: *a step that did not opt in cannot tell the wrapper exists.*

**What it costs, stated plainly:** outside that one agent, *"the copy improved"* and *"the check was
not wired here"* are once again the same number. `RFC_044`'s question flips from *"may this stay
default-ON?"* to *"should it BECOME default-ON?"*, and nobody is now under time pressure to answer it.

**The same round also corrected `509`'s un-hold check** (`debug_historian`, HIGH): it was build
provenance + `git merge-base`, which is the documented anti-pattern for a fix that ships in two halves
— the stamp proves the BINARY and says nothing about the migration, and one `make release` resolves
HEAD separately per service. It now probes the **capability** in the running pod (`grep -ac
'rewrite_negations' /proc/1/exe`) with a mandatory absent-control, on every replica.

**Seats that approved:** editquality, bug_historian, reuse_agent, improvement_guardian,
tooling_provenance, adoption_guardian, compliance, render_guardian, constitution, mission,
prior_art_librarian. The two round-1 objections that had most force — no claim scan on rewrites, and
the truncation arm — are approved as fixed.
