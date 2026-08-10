# Where we are — council-gate cost

Plain-prose log for the owner. Append-only, newest at the bottom.

## 2026-08-10 — you asked me to switch off the improvement loop; it wasn't the problem

You hit a spending cap and asked me to turn off the improvement loop because it
was eating credits. I turned it off — three switches, reversible in one line,
nothing deleted.

Then I measured it, and it turned out the improvement loop was **already
switched off in every way that matters**. Its main scheduler had been off since
May. Its two big reviewing agents hadn't made a single AI call in over a day.
And 544 findings it had produced were sitting in a queue that nothing was able
to pick up, because the step that promotes them was disabled too. So it was
idling, not spending.

What was actually spending was the **council gate** — the review panel that
checks platform changes before they ship. In 24 hours it made 119 AI calls
consuming **11.6 million tokens of input**, about **85% of all the AI spend
across the whole fleet**. Everything else put together was a rounding error next
to it.

I want to be plain that I got the framing wrong twice on the way here, and both
times it would have cost you money or trust if it had gone unchecked:

- I nearly reported that the "auditor" agents were the culprit. That number came
  from a query where I'd written `AND … OR …` without brackets, which silently
  threw away the "last 24 hours" part. The figures were real but were totals
  going back weeks. I caught it because they looked too big.
- I told you a stray piece of boilerplate in the prompts was worth removing as a
  cost saving. When I actually measured it, it was about **8p a day**. You said
  don't remove it if it's useful — and you were right, it is: it's the
  instruction that stops the AI wrapping its answers in formatting that would
  break everything downstream. It's untouched.

## What's actually wrong with the council gate, and what I've done

The council has 17 reviewers. Each one gets the same ~100,000 words of context —
the plan being reviewed, the rationale, the database schema. They genuinely all
need it. The waste is that **the identical material is sent 17 separate times
and paid for at full price every time.**

Anthropic sells a fix for exactly this: send the shared part once, and the other
16 reviewers read it back at a tenth of the price. Two things were stopping us:

1. **Nobody had ever built it.** I searched the whole codebase — prompt caching
   had never been implemented anywhere.
2. **Even if we had, it wouldn't have worked**, because of how the prompts are
   assembled. Caching only works on a shared *beginning*. Each reviewer's
   prompt started with its own job description and only then got to the shared
   material — so no two prompts started the same way, and there was nothing to
   cache. I checked: the shared part is genuinely identical across all 17, but
   the overlap started at a different point in each one.

One piece of luck: the 17 reviewers run **one after another**, not all at once.
That matters, because reviewers running simultaneously can't reuse each other's
cached copy — they'd all be paying full price before the first one finished
writing it. Running in sequence is the arrangement that makes this work.

So I've done three things:

- **Taught the system to cache**, in a way that is off by default. Every other
  agent that shares this code is untouched unless it explicitly opts in — I
  tested that specifically, by deliberately breaking it and confirming the test
  caught it.
- **Started recording the cache figures**, which were being thrown away. This
  matters more than it sounds: once caching is on, the "input tokens" number
  everything currently measures stops meaning what it used to, and every
  existing cost report would understate the true usage by about 95% — in the
  flattering direction. I'd rather not hand you a saving that's partly an
  accounting illusion.
- **Rewritten all 17 reviewer prompts** so the shared material comes first. This
  is prepared and verified but **not yet applied**, because a council review was
  running through those very prompts while I worked. Changing them mid-review
  would produce a verdict assembled from two different versions, and nothing
  would show that had happened.

Expected saving once it's live: roughly **80% off the dominant cost** — about
**$18 a day now, and closer to $29 a day** after Anthropic's introductory
pricing on this model ends on 31 August.

I've put the change through the council gate itself, which is a bit circular but
is the rule for platform changes. That review is what's currently running.

**One honest caveat**: I haven't been able to verify the exact cache request
format against the live API, because doing so would spend council credits on the
very thing we're trying to rein in. Everything is tested against a simulated
API. The first real review after this goes live is the moment to watch, and I've
written down exactly what "working" and "silently not working" look like — they
are easy to confuse, because a cache that isn't working still returns perfectly
good answers, just at full price.

If you'd rather have a cheaper, cruder saving in the meantime, the other lever
is still available and takes minutes: move the more mechanical reviewers onto a
cheaper model. It's config-only and reversible. Say the word.
