# CLOSING SUMMARY — five providers, one pipeline, and a meter on the part we couldn't see (2026-07-28)

*Closes the Gemini content-provider workstream. Every figure below was checked against
the live system on 2026-07-28; the queries are in `RUNBOOK_gemini_content_provider.md`
§10 and in the NOTES entry for this date. Written to be read aloud.*

---

## What we're trying to do

Make the platform genuinely multi-provider — and be able to prove it, rather than assert
it. The narrow job was moving the copywriting agents to Google's Gemini after a failed
attempt in July. The wider job, which is what actually closes here, was establishing that
different parts of the system can run on different suppliers' models, chosen for the work
rather than by default.

## Where we've come from

On 24 July we pointed the writer at Gemini and it returned **zero characters**. We read
that as the model being incapable, reversed the change within the hour, and went back to
Claude.

That reading was wrong and the fault was ours. Newer models think before they answer, in
tokens nobody sees. Anthropic bills that separately; Google takes it out of the *same*
budget you set for the answer. Our code was written for the first convention and passed
the same number to the second. So when we allowed 100 tokens, the model spent 92 thinking
and had 8 left to write with. Not incapacity — arithmetic, and ours.

**And we had no instrument that could tell those two apart.** That is the part worth
remembering.

## What we've done

Fixed the arithmetic — the answer now gets its own budget with headroom reserved for
thinking — and put it through the reviewer council, which approved it. Both content
agents run on Gemini in production and have written real pages on live sites.

Then the thing this summary is really about. Our logging recorded what we asked for and
what came back, **but silently discarded what the thinking cost**. The largest driver of
the bill was invisible to every query we had. We added four columns to record it. The
council approved that unanimously, on the third attempt, and the two rounds before it
were worth having: one found five call sites we'd missed, and the other caught our
submission claiming something its own diff no longer showed.

Two mistakes of ours are recorded rather than tidied away. We called a live page broken
when the fault was in how we were looking at it. And in the very change that fixed *"one
column meaning two things"*, we introduced a column whose name meant something other than
what it held — caught by reading the first four rows instead of trusting our own naming.

## Where we are now

**Five providers, in production, doing different jobs. All verified live today, not
recalled from notes:**

| what | provider & model | evidence |
|---|---|---|
| **Page copy** | Google `gemini-pro-latest` | 24 logged calls; live pages written |
| **Workflows & orchestration** | Anthropic Claude — opus-4-6, sonnet-5, sonnet-4-6, haiku-4-5 | 5,953 calls across 33 agent types; 67 agent definitions |
| **News** | xAI `grok-4-1-fast` | 4 active sources, all fetched this morning |
| **Images** | Google `gemini-3-pro-image-preview` ("banana") and Stability SDXL | 152 assets and 69 assets respectively |
| **Vet-med price scraping** | self-hosted `mistral-small3.1` on ollama | 133 calls, still running today |

> **One correction to the owner's own summary, and it matters because it would have been
> published.** The Companies House reviewer is **not** on ollama — it runs on Anthropic
> `claude-haiku-4-5`. The self-hosted ollama lane is the **vet-med price collector**
> (`mistral-small3.1`). Two scraping-and-extraction jobs, easily swapped in memory; the
> live rows are unambiguous, and this is why we check figures before repeating them.

**Five suppliers — Anthropic, Google, xAI, Stability, and our own hardware** — and the
Grok and image lanes are genuinely different integrations, not a model name swapped in a
config. That is the substantive part: it is a real capability, exercised daily, not a
roadmap item.

**What the meter shows.** Across 11 live page-writing calls: **2,439 tokens of visible
copy against 20,826 tokens of thinking** — roughly **8.5 words of invisible reasoning for
every word we keep**. Thinking is 89.5% of the *billable output*; counted against
everything sent and received it is 25.6%, because the prompt is the larger part at 71.4%.
Both are true and answer different questions, and the first is misleading alone. We have
deliberately not converted any of it to money: rates move, input and output price
differently, and one supplier is on an introductory rate expiring 31 August.

**The decision.** The owner has ruled that we keep Gemini, *because being able to run
different models in our workflows is part of the story we're telling*. The ordering
matters: we measured the cost first and accepted it second. Anyone who later finds that
ratio unacceptable is disagreeing with a deliberate judgement, not catching an oversight.

## Where we're going

Nothing in this workstream is outstanding. `bugs_open/107` and `bugs_open/110` are both
closed and verified on the running binary.

Three things carry forward, none of them Gemini's:

- **The meter changes job** — from decision input to drift watch. It is what will tell us
  if a prompt change quietly doubles the thinking. An occasional look, not a dashboard.
- **The diversity claim is now load-bearing, so it has to stay true.** If some future
  tidy-up proposes consolidating onto one supplier for simplicity, it would be overriding
  a deliberate decision rather than cleaning up an accident.
- **Two page-assembly defects stand, unrelated to any model**: the writer generates each
  section blind to its siblings, so a page's opening and closing blocks can make the same
  point twice; and one page dropped its product list for want of data, leaving copy that
  advertises a sale with nothing to buy underneath.

The lesson worth keeping is not about Gemini at all. **We spent a day concluding a model
couldn't write, when we had starved it and had no way to see that.** The fix was small.
Not being able to see was the expensive part — and the estate above is only defensible
because we can now show what each part of it actually does.
