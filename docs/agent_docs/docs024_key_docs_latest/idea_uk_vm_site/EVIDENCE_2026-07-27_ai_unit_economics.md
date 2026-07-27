# EVIDENCE — what a production AI report actually costs us (idea.uk, 2026-07-27)

**Canonical source for the idea.uk unit-economics figures.** Contributed for the
leopardess-consulting and fundamentallyai copy threads, which asked for it. Point at this
file; do not copy the numbers into a second document that can drift from it.

> ## ⚠ READ THIS BEFORE PUTTING ANY NUMBER IN COPY
>
> **The measured figure is a FLOOR, not a total.** It covers two of the five model calls in
> one report. It is not "what a report costs" and must not be published as though it were.
>
> A complete, self-measuring figure arrives automatically on the **next real customer order**
> — the logging gap that caused this was fixed and deployed the same day (`fb10b2659`, 6th
> deploy). **The right move for a copy thread is to wait for that number, not to dress this
> one up.** It is days away, not weeks.
>
> Both of these threads produce outward-facing prose, and this repo has an entire workstream
> (`bugs_open/043`) about generated copy inventing quantitative claims, plus a standing rail
> on the oufe lane: *no figure in any brief or spec*. A floor presented as a total is the
> exact failure those exist to prevent.

## What was actually measured

One full engine run on the deployed production binary (`idea internal` — the authenticated
no-order, no-billing path), 2026-07-27, models `claude-opus-5` (assess/generate/verify) and
`claude-sonnet-5` (critique/score), taken from the binary's compiled-in defaults.

- **Wall clock:** 445 s (7 min 25 s), exit 0
- **Output:** an 11,347-character report on a submitted business idea, with live web search,
  cited sources, and an honest "no further idea cleared the bar" refusal
- **Model calls logged:** 2 of 5

| | tokens | $/M | cost |
|---|---:|---:|---:|
| cache write | 37,033 | 6.25 | 0.2315 |
| cache read | 171,265 | 0.50 | 0.0856 |
| input | 1,908 | 5.00 | 0.0095 |
| output | 12,561 | 25.00 | 0.3140 |
| **MEASURED FLOOR** | | | **$0.641** |

Rates are Anthropic list, taken from the `claude-api` skill on the day, not from memory:
Opus 5 $5 in / $25 out per MTok; cache write 1.25× input and cache read 0.1× input at the
5-minute TTL the engine uses.

## Why it is only two of five calls

The usage log was gated on cache activity. A call whose system prompt falls under the
cacheable minimum (**512 tokens on Opus 5, 1024 on Sonnet 5**) caches nothing, so it logged
nothing, so its tokens are unrecoverable from that run. Both Sonnet steps were invisible.
Fixed the same day — logging is now unconditional — but **this particular run cannot be
completed retrospectively.**

## What is assertable, and what is not

| Claim | Status |
|---|---|
| "$0.641 of model spend was measured across two of the five calls in one production report" | ✅ **Measured.** Cite freely. |
| "The whole report is under roughly $1.30" | ⚠️ **[INFERRED]** — the measured calls are the expensive ones (Opus at `xhigh`, 32k max-tokens); the unmeasured are Sonnet at lower caps. Reasonable, defensible, **not measured**. Mark it if used. |
| "A report costs us $0.64" | ❌ **False.** That is the floor presented as a total. |
| "Model cost is a low single-digit percentage of a £29 sale" | ⚠️ **[INFERRED]**, and rests on the bound above. |
| Any precise per-report figure | ❌ **Not yet.** Wait for the next real order. |

## The bit with an expiry date on it

`claude-sonnet-5` is on an **introductory rate ($2/$10 per MTok) that ends 2026-08-31**,
reverting to $3/$15 — a 50% rise on the critique-and-score half of the bill. Immaterial at
current volume, but **any published figure should carry its measurement date**, because this
one has a known step change five weeks out. A dated figure ages honestly; an undated one
becomes wrong silently.

## The genuinely interesting angle for copy (no numbers required)

The defensible, durable story here is not the cost — it is the **shape**:

- A bespoke, web-researched, source-cited report on a stranger's business idea, produced
  end to end by an AI pipeline in **seven and a half minutes**, sold at £29 with a human
  reviewing before it goes out.
- The pipeline **declines to pad**: on this run it searched for further ideas, judged none
  good enough, and said so in the report rather than manufacturing filler. That is a
  product-integrity claim, and it is demonstrable.
- The report is willing to be **unflattering where it costs the sale** — it told the
  submitter the paying demand was on the seller side and that a dozen competitors give the
  same advice away free.

Those are all directly evidenced by the artefact and need no contested arithmetic. If a
figure is wanted, take it from the next real order once the automatic per-order record has
run — by then it will be complete, and it will be a measurement rather than a bound.

**Full working:** `RUNNING_NOTES_idea_uk_vm_site.md` §X.25.
