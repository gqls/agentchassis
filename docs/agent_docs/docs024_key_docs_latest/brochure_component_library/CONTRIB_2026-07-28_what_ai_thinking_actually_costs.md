# CONTRIBUTED — what "thinking" actually costs on a live AI system (2026-07-28)

Left here by the Gemini content-provider thread at the owner's request, for
**fundamentallyai.com**. **Not a work item for this thread, and nothing in the brochure
component library needs to change.** Contributed because fundamentallyai.com is where we
talk about how AI systems actually behave, and this is a first-party measurement of
something the industry mostly asserts. The owner suggested an infographic — that's a
suggestion, not a brief. "Not worth a page" is a fine answer.

**Canonical figures and the full working live in one place — read them there, don't
copy them here:**

`docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/SUMMARY_2026-07-28_the_thinking_meter.md`

## The one-paragraph version

Modern AI models "think" before answering — they work through the problem in tokens the
reader never sees, and you are billed for them. We fitted a meter to that on our own
production system. Across eleven real page-writing calls: **2,439 tokens of visible copy,
20,826 tokens of invisible thinking.** Roughly **eight and a half words of reasoning for
every word kept.** Until 28 July we could not see this at all: our logging recorded what
we asked for and what came back, and silently discarded what the thinking cost — so the
biggest single driver of the bill was invisible to every query we had.

## ⚠ Before any of this reaches a page

- **"89.5%" needs its neighbour or it lies.** Thinking is 89.5% of *output* tokens, but
  **25.6%** of all tokens in play, because the prompt is the larger part (71.4%). Both
  are true and they answer different questions. Alone, 89.5% reads as "almost all AI
  spend is wasted", which is **not** what we measured. If the design has room for one
  number, use **8.5× thinking-to-visible** — it is honest unaccompanied.
- **These are TOKENS, not pounds.** Deliberately not converted: rates move, input and
  output are priced differently, and one provider is on an introductory rate expiring
  2026-08-31.
- **Small sample, one agent, one day.** Eleven calls from one writer on 2026-07-28. Write
  what *we* measured on *our* system on *that date* — not "AI systems spend…".
- **This site has an open fabrication problem** (`bugs_open/043`, `bugs_open/123`): its
  generated copy has previously invented statistics and phrased them as sourced. A page
  built from this note is exactly the shape that goes wrong — a real figure, adjacent to
  plausible-sounding ones nobody measured. Any number on the page must trace to the
  SUMMARY above; anything that doesn't, isn't ours.

## An illustration that is already true, if you want one

There is a natural visual here and it doesn't need embellishing: **a bar where the
visible slice is a tenth of the whole**, captioned with what the invisible part is.
Nothing needs exaggerating for it to be striking — the honest ratio is the striking part,
which is rarer than it sounds.

## The angle that needs no arithmetic

In July we pointed our writer at a new model and it returned **zero characters**. We
concluded the model couldn't do the job and reversed the change within the hour. That
conclusion was wrong. The model had been given 100 tokens to answer in, and it spent 92
of them thinking — leaving 8 to write with. It wasn't incapacity; it was our own
arithmetic, and **we had no instrument that could tell "this model can't write" from "we
starved it".**

That's the piece worth publishing, with or without the ratio: *the failure looked like
the model's and was ours, and we could not tell the difference until we built something
to measure it.* Both halves are evidenced by artefacts rather than a calculation someone
can dispute — the zero-character result is in the July record, the fix is
council-reviewed, and the figures come from a database column that did not exist the day
before.

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
