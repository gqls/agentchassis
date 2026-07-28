# CONTRIBUTED — what "thinking" actually costs on a live AI system (2026-07-28)

Left here by the Gemini content-provider thread at the owner's request. **Not a
leopardess work item, and nothing in your workstream needs to change.** Contributed
because leopardess sells judgement about AI, and this is a first-party measurement of
something the industry mostly asserts. Make something of it or don't — the owner
suggested an infographic; it's your call, and "not worth a page" is a fine answer.

**Canonical figures and the full working live in one place — read them there, don't
copy them here:**

`docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/SUMMARY_2026-07-28_the_thinking_meter.md`

## The one-paragraph version

Modern AI models "think" before answering — they work through the problem in tokens the
customer never sees, and you are billed for them. We fitted a meter to that on our own
production system. Across eleven real page-writing calls: **2,439 tokens of visible copy,
20,826 tokens of invisible thinking.** Roughly **eight and a half words of reasoning for
every word kept.** Until 28 July we could not see this at all — our logging recorded what
we asked for and what came back, and silently discarded what the thinking cost. The
biggest single driver of the bill was invisible to every query we had.

## ⚠ Before any of this reaches a page

- **"89.5%" needs its neighbour or it lies.** Thinking is 89.5% of *output* tokens, but
  **25.6%** of all tokens in play, because the prompt is the larger part (71.4%). Both
  are true; they answer different questions. Quoted alone, 89.5% reads as "almost all our
  AI spend is wasted", which is **not** what we measured and not what we believe. If only
  one number can fit the design, use **8.5× thinking-to-visible**, which is honest on its
  own.
- **These are TOKENS, not pounds.** We deliberately did not convert. Rates move, input and
  output are priced differently, and one provider is on an introductory rate expiring
  2026-08-31. A money figure here would be an inference dressed as a measurement.
- **Small sample, one agent, one day.** Eleven calls from one writer on 2026-07-28. It is
  a real production measurement, not a benchmark. Do not write "AI systems spend…" — write
  what *we* measured on *our* system on *that date*.
- **This site's own rule stands: no claim ships without an AUDIT row.** This note is
  evidence, not an audit row, and it does not create one.

## The angle that needs no arithmetic

The strongest material here isn't the ratio — it's the story around it, and it needs no
figure to land.

In July we pointed our writer at a new model and it returned **zero characters**. The
obvious conclusion was that the model couldn't do the job, so we reversed the change
inside an hour. That conclusion was wrong. The model had been asked to answer in 100
tokens, and it spent 92 of them thinking — leaving 8 to write with. It wasn't
incapacity. It was our own arithmetic, and **we had no instrument that could show us the
difference between "this model can't write" and "we starved it".**

For a consultancy selling judgement about AI, that is a better line than any cost figure:
*the failure looked like the model's and was ours, and we couldn't tell until we built
something to measure it.* It is also the honest case for measurement over vendor
comparison — the sort of thing a client is usually sold the opposite of.

Both halves are evidenced by artefacts rather than by a calculation someone can dispute:
the zero-character result is in the July record, the fix is council-reviewed, and the
figures above come from a column that did not exist the day before.

---

## ADDENDUM 2026-07-28 — the decision landed, and it is a better angle than the ratio

The owner has ruled that we **keep** the second model: *"I think we can keep gemini
because it is part of the story that we can use different models in our workflows."*

That is worth more to you than the cost figure above, because it is a **position**
rather than a measurement, and it is one most firms cannot honestly claim:

- **We measured what the second model costs, and kept it anyway, on purpose.** The
  ratio came first, the decision second. That ordering is the point — it is the
  difference between a considered trade and an unexamined default.
- **The pipeline genuinely is heterogeneous today.** Two content agents on Google, most
  of the estate on Anthropic, one price-scraping job on a locally-hosted model. Not an
  aspiration or a roadmap item.
- **Not being locked to one supplier is the argument**, and it is more durable than any
  token count: prices move, models get retired, terms change. A pipeline that can move
  between providers has already proved it can, because it has.

**Still bound by the caveats above.** This addendum adds a claim about *our own
decision*, which is safe to state because we made it. It does not license the numbers
to travel without their neighbour, and it does not become a claim about what other
firms do or should do.

---

## CORRECTION + EXPANSION 2026-07-28 — the estate is bigger than the addendum said, and now verified

The addendum above said *"two content agents on Google, most of the estate on Anthropic,
one price-scraping job on a locally-hosted model."* That was **incomplete** — it
understated the story by two whole suppliers. Checked against the live system on
2026-07-28, the real picture is **five providers**:

| what | provider & model | evidence |
|---|---|---|
| Page copy | Google `gemini-pro-latest` | 24 logged calls; live pages |
| Workflows & orchestration | Anthropic Claude (opus-4-6, sonnet-5, sonnet-4-6, haiku-4-5) | 5,953 calls, 33 agent types |
| News | xAI `grok-4-1-fast` | 4 active sources, all fetched that morning |
| Images | Google `gemini-3-pro-image-preview` + Stability SDXL | 152 and 69 assets |
| Vet-med price scraping | self-hosted `mistral-small3.1` on ollama | 133 calls |

**Anthropic, Google, xAI, Stability, and our own hardware.** The news and image lanes are
genuinely different integrations, not a model name swapped in config.

**If you use one line from this whole note, use that table.** "We run five model
providers across one pipeline, each chosen for its job" is a stronger and more checkable
claim than any token ratio — and unlike the ratio it needs no caveat, because it is a
description of what the system does rather than a statistic about it.

**Still bounded.** Every row above was verified on 2026-07-28; providers change, so date
any published version. And a lane being live is not a claim that it is the best available
choice for that job — only that we run it and can show it.
