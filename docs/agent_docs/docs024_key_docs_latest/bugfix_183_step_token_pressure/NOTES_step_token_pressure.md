# NOTES — bugs_open/183 step token pressure (append-only, newest at the bottom)

## 2026-08-06 — picking the bug

Ranked the open backlog by reference-heat over transcripts touched in the last 4h,
then confirmed the top candidates by **fix-site symbol** grep (not bug number — a
number matches every session that ran `ls bugs_open/`).

- **186** (thunder NULL instance_ip): rejected. `instances.go` at 107 hits in a
  session 0.4h old, `lookupOne`/`InstanceIP` alongside it. Someone is in that file.
- **189** (locked positional slot duplicates): rejected. `matchLockedRow` and
  `extractSectionsFromMetadata` are hot in the active 204/182 lane's council
  argument — a session literally quoting `matchLockedRow` as "a narrow,
  already-disclosed exception" in a submission.
- **183**: taken. `who-owns.py` → no owning workstream; zero open `site_work_items`;
  the only `domain-research-classifier` heat is an adjacent Firecrawl-audit lane
  whose own text says **"domain-research-classifier has NO owner"**. That is the
  false-positive case from my own memory file working in reverse: high mention
  count, but reading the hits showed a lane declining to touch it.

## Validity re-check before starting

Cap is now **32000**, unshadowed (root `ai_service` block NULL), last run 2026-08-04
succeeded at 4295 tokens. So the ACUTE symptom is gone and the bug is still correctly
OPEN: its own candidates 2–4 are about the class, not the incident.

## MISSTEP 1 — I walked into the trap my own memory file documents, on my first query

First fleet measurement filtered `WHERE success AND output_tokens IS NOT NULL`. That
is **exactly** the `llm_call_log` trap recorded in LANDMINES and in my memory index:
a truncated call logs `output_tokens = NULL`, so the filter silently deletes every
truncation from the population — and it is blindest on the steps that truncate most.

What it produced: `classify_and_extract@6000` at "p95 93.9%, max 5642", reading as a
step under mild pressure. The truth in the same window was **5 truncations**. The
corrected query moved it to `p95 100.0%, 5 truncated`, and the fleet table changed
shape entirely — `extract_and_reconcile@2048` went from **absent** to the largest
truncation count in the estate (63).

The check that would have caught it instantly, and which is now the rule for this
lane: **before trusting a zero, induce a non-zero.** I knew the truncations existed
(they are quoted in 183's own evidence section) and still wrote a query that returned
none of them.

## MISSTEP 2 — my plan asserted the 14-day window was clear of the agent_type relabel

Written into the PLAN as fact: "the 14-day window is now clear of the relabel date."
It is not. The relabel was 2026-07-26; 14 days back from 2026-08-06 is **2026-07-23**,
inside the `generic` era. Caught while re-reading my own plan, not by a query.
Corrected in place. The fix was structural rather than waiting for the calendar: key
on `(step_name, cap)` and treat `agent_type` as a display label only.

## MISSTEP 3 — `DISTINCT ON (agent_type, step_name)` resurrected a retired cap

Consequence of the same relabel, and it only appeared when I widened to 90 days. The
"current cap" CTE originally took the latest call **per (agent_type, step_name)**.
Over a 90-day window the pre-relabel `generic` rows form their own group whose "most
recent call" is months old — so `classify_and_extract@6000`, a cap retired on 08-02,
came back in the LIVE run as a current pair.

The tell was that I already knew the answer: the live check must NOT flag this step
(cap 32000, last run at 13%). Getting the expected-absent case wrong is what exposed
it. Fixed by taking the latest call **per step_name** only, with the trade written
into the seed's header and a standing check for it in the RUNBOOK (today: 0
step_names held at different caps by different agents).

## MISSTEP 4 — the check I first wrote could not have flagged its own bug in time

The clone of FIX-058 used a 14-day window. Measured as-of 2026-08-01 (the day before
183's step first truncated), `classify_and_extract@6000` had run **3 times** in 14
days — under any sane n floor, so **silent**. The check would have said nothing until
after the sites burned, then flagged T.

Same window widened to 90 days, same date: **n=15, p95 90.0%, peak 94.0% → flags P**,
a full day before the first truncation. That is the difference between a leading and
a lagging indicator, and it was invisible until I pinned `now()` and asked what the
check would have said *then*.

This is the `5d1df2777` lesson in a new costume — a mechanism that looks correct,
runs clean, and is **inert on the exact case it was built for**. Two departures from
FIX-058 came out of it: 90-day window, and n ≥ 5 rather than n ≥ 20 (183's step had 9
in-window calls on the day it burned; at n ≥ 20 the check cannot see its own bug).

Also, trivially: `round(double precision, integer)` does not exist in this Postgres —
cast the percentile `::numeric` first.

## Verification actually run (2026-08-06)

1. **Seeded, then executed the pre_query VERBATIM out of the live row** (not my file
   — the row is what the scheduler runs). It inserted note `3186dcfa`, 10 pairs.
2. **Dedup proven**: re-running the identical pre_query returned **0 rows**. Event,
   not heartbeat, demonstrated rather than asserted.
3. **Known-case test, window pinned**: as-of 2026-08-02 18:00 → `classify_and_extract@6000`
   top of list (T, n=21, 5 truncations). As-of 2026-08-01 00:00 → flags **P** before
   any truncation. Live → correctly **absent**.
4. **Negative control**: `stage_implement@32000` (p95 24.4%) never flags on pressure —
   it appears only under T, from a single real truncation. `generate_tool_html`
   (raised to 32000) does not appear live; it does at 8000 in history. The retired
   population stays retired.

## The first run found a live bug — `bugs_open/205`

Top line was `extract_and_reconcile@2048`, 64 truncations, **the largest in the
fleet**, and no bug file mentioned it. Followed it:

- The step sets **no `max_tokens` at any level** and no root block exists — so it
  runs on the transport's hardcoded `2048` (`platform/aiservice/anthropic.go:109`).
- 100% failure since 08-05 ~03:00; 0 of 54 truncated on 08-04.
- **The shape is not cap drift.** Grouping by `md5(prompt_rendered)`: 46 calls of one
  byte-identical prompt, then 18 of a second, **zero successes on either**, while
  every other prompt in the window succeeded at 468–639 output tokens. Distinct
  `correlation_id` per call ⇒ fresh dispatches, not one orchestration's retries.
  Driver: `scheduled_tasks` `vet-batch-verify`, every 300s, enabled.
- Census run rather than left for a reviewer: **8 of 126 active LLM steps set no cap
  at any level**; only 2 have run in 90 days; the other 6 are latent.

**The lesson worth keeping, and it is now a LANDMINE on this check:** a truncation
COUNT does not tell you the SHAPE. 64 truncations looked like the worst cap-sizing
problem in the estate; it is one bad record on a loop, where a bigger cap is the
wrong first move. `md5(prompt_rendered)` group-by is the one-query discriminator, and
it belongs beside every flag this check raises.

## Council gate

Attempted per CLAUDE.md. The gate's scope is `platform/`, `internal/`, `pkg/`
(owner ruling 2026-07-17) and the trigger **refuses docs/config-only submissions
client-side** so they never spend credits — this change is a live `scheduled_tasks`
row plus docs and touches no Go. Recorded rather than forced: `FORCE=1` exists but
submitting out-of-scope work to burn seats on it is the behaviour the refusal is
there to prevent. Review here follows the RFC_006 precedent instead — the seed, its
reasoning, the register entry and the first note are all readable, and the pinned
known-case test is the disconfirmable part.

## 2026-08-06 (later) — cap raised to 64000 on owner instruction

Applied via `SQL_2026-08-06b_classifier_cap_32000_to_64000.sql`, guarded both ways
(refuses if a root `ai_service` block appears; RAISEs unless exactly one row reads
64000 after). Verified live: cap 64000, root block NULL, model `claude-sonnet-4-6`,
prompt still 16,950 chars and `input_fields` still 4 — so only the cap moved.

**Why 64000 and not 128000** (the model's real ceiling — checked, not recalled:
Sonnet 4.6 and Sonnet 5 both cap at 128K output, 1M context):

- The chassis **does not stream**. `platform/aiservice/anthropic.go:72` is one
  `http.Client{Timeout: 600s}`, and the comment beside it records having hit
  "Client.Timeout exceeded" at ~600,0xx ms. A cap the model can actually fill is
  bounded by wall-clock here, not by the API limit.
- 64000 is **already live and exercised** on this chassis — `recreate_tool`, 77 calls
  in 90 days, peak 11,888 — so it is a proven operating point.
- 128000 would make this step the fleet's **only** 128000. That is the exact singleton
  shape that hid this bug for months when it was the only 6000: no sibling to compare
  against, nothing to notice when it drifts.

Also measured while choosing: `verdict@32000` has peaked at **31,860 — 99.6% of cap**
over 305 calls. 32000 is genuinely reachable in this fleet, so the raise is not
theoretical headroom.

## MISSTEP 5 — I wrote an unverified structural claim into the bug file

Asked to explain the structural split, I wrote into `bugs_open/183` that the four
sections "are not independent — three of them read the classification the same call
produces." **I had not opened the prompt.** I had measured its LENGTH (16,950 chars)
earlier and let that stand in for having read it.

Then I opened it. What the template actually says is different in a way that changes
the design:

- `identity` is the upstream section, not `classification` — the prompt requires every
  `content_direction` field to be "specific to THIS industry" and the palette to be
  derived from the industry, and industry is an `identity` field.
- The real cross-section coupling is `classification.suggested_style` ↔
  `design_intent.style_direction` — the **same enum**, with "style_direction must agree
  with the palette you emit".
- The three adoption-branch consistency rules point at the **inputs** (the adopted
  archetype/content_direction/design_intent), not at this call's own output, so they
  survive any split unchanged.

Corrected in both files, marked as a correction in the DESIGN doc.

**The check, and why this one stings:** *reading a field's LENGTH is not reading the
field.* This is the third instance in one session of the same family — measure
something adjacent to the question, then answer the question from it. The other two
(the `output_tokens IS NOT NULL` filter, the 14-day window) were caught by running a
query. This one was caught only because the task happened to require me to open the
file. Had the owner asked "raise the cap" and nothing else, the wrong claim would have
shipped in a bug file and been inherited by whoever builds candidate 3.

**A design built on a misread dependency graph fails in a specific, expensive way**:
four parallel calls (what my wrong version implies) produce a dark palette under a
`modern-light` classification, and nothing catches it until a site renders.

## 2026-08-06 (later still) — the limits census, and a blind spot I shipped

Owner asked for the limits in one place. Full write-up:
`SUMMARY_2026-08-06b_the_real_limits.md`. Three things came out of measuring rather
than recalling:

1. **Context is not a constraint and I would have guessed wrong.** Peak input across
   30 days is 126,195 tokens against a **1M** window — 12.6%. Every truncation this
   platform has had is an OUTPUT collision with our own configured cap, never with the
   model. Worth saying out loud because "we hit a token limit" slides very easily into
   "we need more context", and we have an order of magnitude spare.

2. **The real ceiling is wall-clock, not tokens.** The chassis does not stream
   (`anthropic.go:72`, one `http.Client` at 600s; gemini the same; **ollama at 120s**).
   Output generation is remarkably linear — Sonnet 5 holds ~98 tok/s from 8k to 32k
   output; Sonnet 4.6 runs 47–82 tok/s. So 600s converts to **~58,800 tokens on Sonnet
   5 and ~28,000–42,000 on Sonnet 4.6**. A `max_tokens` above that cannot be reached;
   the clock fires first. In 90 days zero calls exceeded 500s, but the peak was
   495,177 ms — **82.5% of the limit**, so the margin is thinner than "600 seconds"
   sounds.

3. **MISSTEP 6 — I shipped a monitor with a blind spot and only found it by being
   asked an adjacent question.** LCO-007 counts truncations from `error_message`
   matching `response truncated:` / `stop_reason=max_tokens`. A call that dies on the
   **clock** instead matches neither AND logs `output_tokens = NULL`, so my own WHERE
   clause (`output_tokens IS NOT NULL OR <truncation match>`) **excludes it from the
   population entirely**. A step that degrades from truncating to timing out therefore
   looks to my check like a step that got better.

   Not hypothetical: 35 rows carry `api.anthropic.com/v1/messages": context canceled`,
   most recent 2026-08-04.

   **What caught it:** nothing in my own verification. I tested the check against
   truncations, pinned the clock, ran a negative control — and every one of those tests
   was inside the failure vocabulary I had already chosen. The blind spot was only
   visible from a question I was not asking (*what are the timeouts?*). **The check I
   should have run: enumerate the DISTINCT error shapes the population can contain,
   before writing the predicate that filters them.** One `GROUP BY left(error_message,60)`
   over `llm_call_log` would have shown the timeout family sitting next to the
   truncation family.

   Family: this is the fourth instance this session of *measuring something adjacent to
   the question and answering from it*, and the second where a verification suite
   confirmed only what its author already believed. Fix recommended (add the timeout
   strings to the pre_query's vocabulary) but **not applied** — it changes a live shared
   check and belongs in a deliberate edit, not a footnote.
