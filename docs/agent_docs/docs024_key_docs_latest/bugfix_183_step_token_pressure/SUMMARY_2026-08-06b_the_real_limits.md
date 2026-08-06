# SUMMARY 2026-08-06b — what actually limits an LLM call on this platform

Written because the owner asked for the limits in one place. Every number here is
measured against the live system on 2026-08-06 and the query is given, so it can be
re-run rather than believed. **The headline is not the one I expected**: neither the
context window nor the model's output limit is what constrains us. The clock is.

## The four limits, and which one bites

| # | Limit | Value | How close we run | Binding? |
|---|---|---|---|---|
| 1 | **Context window** (input) | 1,000,000 tokens | peak **126,195** = 12.6% | **No — not close** |
| 2 | **Model output ceiling** | 128,000 tokens | peak **31,860** = 25% | No |
| 3 | **Configured `max_tokens`** | per step, 2,048–64,000 | routinely 90–100% | Sometimes — this is `bugs_open/183` |
| 4 | **HTTP client timeout** | **600 seconds** | peak **495s = 82.5%** | **Yes — this is the real ceiling** |

## 1. Context is not a problem, and it is worth saying so plainly

We are using about an eighth of the context available to us.

```sql
SELECT model_resolved, count(*), max(input_tokens) AS peak_in,
       percentile_cont(0.95) WITHIN GROUP (ORDER BY input_tokens)::int AS p95_in
  FROM llm_call_log WHERE created_at > now() - interval '30 days' AND input_tokens IS NOT NULL
 GROUP BY 1 ORDER BY 2 DESC;
--  claude-sonnet-5   | 6054 | 126195 | 68707
--  claude-sonnet-4-6 | 3100 | 105595 |  6300
--  gemini-pro-latest |  934 |   8149 |  6419
```

Sonnet 4.6 and Sonnet 5 both carry a **1M-token context window**. Our heaviest single
call used 126K. So "we ran out of room for input" is not a failure mode we have, and
any future design that wants to send *more* context — more research, more of the
scraped site, more prior specs — has an order of magnitude of room to do it in.

**This is the opposite of the intuition.** Every truncation incident on this platform
has been about *output*, and it is easy to slide from "we hit a token limit" to "we
need a bigger context window". We don't.

## 2. Output is where the pressure is — but the model is not the limit either

The model will write up to **128,000 tokens** of output. Our largest observed single
completion is **31,860** (`verdict`, diagnose-agent, at a 32,000 cap — 99.6% of *its*
cap, but only 25% of what the model would do).

So every truncation we have ever had was a collision with **our own configured
`max_tokens`**, never with the model's ability. That is what makes it a config bug
class rather than a capability problem, and it is why the fix is always either a
bigger number or a smaller unit of work.

## 3. The real ceiling: 600 seconds ÷ ~98 tokens per second

This is the finding that reframes the rest, and it is the reason the cap raise stopped
at 64,000 instead of going to the model's 128,000.

**The chassis does not stream.** Every Anthropic call is one blocking HTTP request:

- `platform/aiservice/anthropic.go:72` — `http.Client{Timeout: 600 * time.Second}`,
  with a comment beside it recording a real `"Client.Timeout exceeded"` at ~600,0xx ms.
- `platform/aiservice/gemini.go:185` — the same 600s.
- `platform/aiservice/ollama.go:55` — **120s**, five times tighter.

Output generation runs at a strikingly stable rate. Measured over 30 days, bucketed by
output size:

| model | output size | median latency | **median tokens/sec** |
|---|---|---|---|
| claude-sonnet-5 | 8,000–12,000 | 94s | 97.6 |
| claude-sonnet-5 | 16,000–20,000 | 172s | 99.4 |
| claude-sonnet-5 | 29,000–31,860 | 308s | 98.2 |
| claude-sonnet-4-6 | 500–4,000 | 23s | **46.8** |
| claude-sonnet-4-6 | 4,000–8,000 | 65s | 82.1 |
| claude-sonnet-4-6 | 16,956–16,959 | 242s | 70.0 |

Sonnet 5 holds ~98 tok/s almost perfectly across the whole range. **So the 600-second
timeout converts directly into a token ceiling:**

- **Sonnet 5: 600s × 98 ≈ 58,800 tokens.**
- **Sonnet 4.6: 600s × ~70 ≈ 42,000 tokens** — and at its slower observed rate of
  ~47 tok/s, closer to **28,000**.

**A `max_tokens` above that number cannot be reached.** The clock fires first. This is
why 128,000 was never a real option: on our slowest path it is four times more output
than we have wall-clock to produce.

We have not actually hit it — in 90 days, **zero calls exceeded 500s** and the peak was
495,177 ms — but 495s is 82.5% of the limit, so the margin is thinner than the raw
"600 seconds" suggests.

## 4. The consequence I did not anticipate, stated plainly

Above roughly **28,000 output tokens on Sonnet 4.6**, the cap stops being the operative
limit and the clock takes over. That changes the *failure mode*, and not for the better:

| | a cap that binds first | a clock that binds first |
|---|---|---|
| Error | `stop_reason=max_tokens` | `context canceled` / client timeout |
| Partial text | **preserved** (`TruncatedError.Partial`) | **lost entirely** |
| Legible to the new monitor? | **yes** — counted as a truncation | **no** — invisible |
| Time to fail | proportional to the cap | always ~10 minutes |

**So the cap-pressure check I shipped today (LCO-007) has a blind spot, and it is
better to write it down than to discover it later:** it counts truncations from
`error_message` matching `response truncated:` / `stop_reason=max_tokens`. A call that
dies on the clock instead matches neither, *and* logs `output_tokens = NULL`, so it is
excluded from the population entirely rather than counted as pressure. A step that
moves from truncating to timing out would look like a step that got *better*.

This is not hypothetical wiring — 35 rows carry
`Post "https://api.anthropic.com/v1/messages": context canceled`, most recently
2026-08-04. (Note those are caller-side cancellations, a different mechanism from the
600s HTTP timeout; both produce the same blind spot.)

## 5. What this means for the classifier cap we just raised

`classify_and_extract` runs on **Sonnet 4.6** and its observed maximum output is
**6,590 tokens** — about 94–140 seconds. It sits at roughly 11% of the wall-clock
ceiling and 10% of its new 64,000 cap. It is not close to any limit on any axis.

**The 64,000 number is deliberately larger than the clock allows.** That is not an
error and I would not change it: for this step both failure modes are equally fatal
(183 rules out salvaging a partial here, because the trailing `design_intent.palette`
would go silently absent), so the extra headroom costs nothing and the runaway case is
~10× the observed maximum. But the honest statement is that **the effective ceiling for
this step is ~28,000–42,000 tokens of wall-clock, not the 64,000 in the config.**

## 6. What I would do about the blind spot (not done, recommend)

Three options, cheapest first:

1. **Teach the monitor to see timeouts.** Add `context canceled` / `Client.Timeout` to
   the pressure check's error vocabulary and score them as cap-reaching. One edit to a
   live `scheduled_tasks` pre_query, no code, no roll. Closes the blind spot for every
   step at once. **Recommended.**
2. **Stream the Anthropic client.** Removes the wall-clock ceiling properly and unlocks
   the model's real 128K. This is a platform change to a seam every agent uses —
   architecture scope, an RFC, and not worth it while nothing is near the limit.
3. **Nothing.** Defensible today: no step is near 28,000 except `verdict`, which peaked
   at 31,860 and is the one population where this could bite first.

## 7. The one number to watch

**`verdict` (diagnose-agent) at cap 32,000, peak output 31,860, 305 calls.** It is the
only step in the fleet routinely producing output large enough for the clock and the
cap to be in the same conversation. If anything is going to expose §4 in practice, it
is that step — and because it runs on Sonnet 5 at ~98 tok/s, 31,860 tokens is ~325s,
still comfortably inside 600s. It is fine today. It is the canary.
