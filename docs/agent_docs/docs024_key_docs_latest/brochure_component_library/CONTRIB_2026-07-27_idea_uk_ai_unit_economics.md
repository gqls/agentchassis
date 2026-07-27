# CONTRIBUTED — idea.uk AI unit economics, for fundamentallyai copy (2026-07-27)

Left here by the idea.uk VM-site thread at the owner's request. **Not a brochure work
item and nothing in your build needs to change.** Contributed because fundamentallyai
argues about AI in production and now has a first-party measurement to point at.

**Canonical figures + full working — read them there, don't duplicate them here:**

`docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/EVIDENCE_2026-07-27_ai_unit_economics.md`

## The one-paragraph version

A real £29 product, measured on current frontier models: a bespoke, web-researched,
source-cited report on a submitted business idea, produced end to end in **7 min 25 s**,
with **$0.641 of model spend measured across two of its five model calls**. The other
three were not logged, so the true total is higher and was not captured on that run.

## ⚠ Before any of this reaches a page

- **$0.641 is a FLOOR, not a report's cost.** As a total it is simply false.
- **A complete figure lands on the next real customer order** — the logging gap was fixed
  and deployed the same day. **Wait for it.** It will be a measurement, not a bound.
- `claude-sonnet-5` is on an **introductory rate ending 2026-08-31**. Attach a measurement
  date to anything published.
- Directly relevant to this thread: `bugs_open/043` exists because generated page copy
  invents quantitative claims, and the `evidence-chart` component was built specifically so
  a figure on a page is **resolved data, not LLM content**. A floor typed into prose is the
  failure mode you already built machinery against. **If this becomes a chart, it wants the
  complete post-fix figure driving it via config — not this one.**
- One live trap from your own notes: **`<svg>` text is invisible to the claims gate.** A
  number rendered inside a chart's SVG will not be caught by the checker, so it needs to be
  right before it goes in, not after.

## The angle that needs no arithmetic

The durable claim isn't the cost, it's that the pipeline **declined to pad**: it searched
for further ideas, judged none good enough, and said so in the report instead of
manufacturing filler — and told the customer the paying demand was on the seller side,
against our own interest. That is an evidenced product-integrity story, and it survives any
argument about arithmetic.
