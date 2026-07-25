# AUDIT — does the paid £29 tool do what /report.html says it does? (2026-07-25)

Owner question, verbatim intent: *"Has the paid-for tool changed? Does it do what we say it does?"*
— quoting /report.html's own description (single-idea research; sources you can check; AI use
clearly indicated; honest refusal; the six coverage areas).

Method: read the tool's actual source (`docs024_key_docs_latest/idea.uk/golang_files/` —
`engine.go` 806 lines, `prompts.go` 210, `service.go` 1051, `billing.go` 154) against the live
/report.html copy. Every claim below cites where in the code it is or is not delivered.

## Has the tool changed?

**No.** The engine is, and has been through its whole visible history (engine.go header: "the
ideation method (v2) … Go port of idea_method_runner.py"; ~30 `.orig` iterations in the folder),
an **ideation engine**: it takes `(domain, audience, assets)` and **generates AI-product ideas
for the customer's business**, then filters them. It has never been a single-idea assessment
service. What has moved is the **copy**: /report.html's current description was authored during
the chassis site build and describes a different product than the binary delivers. The copy
drifted from the product, not the product from the copy.

## What the tool ACTUALLY does (engine.go:386 RunMethod, prompts.go)

1. **Audience framing + challenge** — states the audience and willingness-to-pay, proposes up to
   3 better-fit alternatives, carries one forward (may not be the stated one).
2. **Generate** — 12–24 candidate AI-product ideas across five lenses (demand, generalist-failure,
   frontier, outcome, asset×capability sweep). Claude Opus.
3. **Cut** — a **different model, different vendor where configured** (GPT-4o if OPENAI_API_KEY
   set) kills candidates with a concrete free substitute or seller-bundled support. "Most
   candidates should die here."
4. **Verify** — **real web search** (web_search tool, up to 12 uses, extended thinking): does the
   data feed exist, do competitors already offer it, is willingness-to-pay evidenced. Candidates
   whose premise fails are dropped.
5. **Score** — six factors incl. a separate operator-risk axis; hard gate (Defensibility ≥3 AND
   Willingness ≥3); Risk=1 (regulated-profession territory) is dropped and surfaced separately;
   ranked report with "a cheap first test" per idea, plus "Didn't make the cut" and "Set aside on
   risk" sections with reasons.

Payment is real Stripe checkout, webhook as source of truth (billing.go). A human operator
reviews every request before the engine runs (service.go /op flow; decline emails politely at no
charge), and the T&Cs disclose AI use and hallucination risk prominently and honestly
(service.go:940).

## Claim-by-claim against the live copy

| /report.html claim | Verdict | Evidence |
|---|---|---|
| "produced for a single idea you submit" / "the problem **your idea** addresses… where **the idea** is defensible" | **NOT DELIVERED** | `EngineFunc(domain, audience, assets)` — there is no "your idea" input to the method. The report intro the customer receives says the opposite: *"You asked us to find AI product ideas for {domain}"* (engine.go:674). The form's field is even labelled "The business **or idea**" — an idea submitted there is treated as a domain to ideate around, not assessed. |
| "competitor and substitute analysis" | **DELIVERED** (per generated idea) | Cut step names the concrete free substitute; verify step web-searches competitors. But it is applied to *our* candidates, not to the customer's idea. |
| "Where we cite a figure or a claim, we explain its source so you can check it yourself" | **NOT DELIVERED** | The report renders `Findings` as 1–2 plain sentences with **no sources**; the verify prompt explicitly instructs *"Do not list strings of product or vendor names"* (prompts.go:142-143). Nothing in the rendered report is checkable by the reader. |
| "Where we use AI processing… that is clearly indicated" | **PARTIAL** | The T&Cs disclose it fully and well. The report itself never mentions AI. |
| "draws on publicly available market data" | **PARTIAL** | Web verification is real, but no market *data* (figures) appears in the report. |
| "Not every idea warrants a report… if research turns up nothing worth building on, we say so plainly" | **DELIVERED** | "No idea cleared the bar" is a real rendered outcome ("That is a real result, not a dead end", engine.go:616); empty cut/verify results short-circuit to an honest note; refusal is encoded as a correct outcome in the system prompt. |
| "a considered next step — a specific, affordable action to test real demand" | **DELIVERED** | "A cheap first test" per advancing idea, with a mandatory liability warning where Risk ≤ 2. |
| "£29" / human review | **DELIVERED** | Stripe checkout; operator approves/declines every order before the engine runs. |

**Also found while auditing:** `reportContact()` falls back to `idea-uk@leopardess.uk`
(engine.go:685) — the stale address already flagged in the site DB. If `CONTACT_EMAIL` is not set
in `/etc/idea/idea.env` on the box, the paid report's "email us" line carries the dead address.
Owner can confirm on the box: `grep CONTACT_EMAIL /etc/idea/idea.env`.

## Bottom line

The tool is genuinely good at what it actually is — a multi-model, web-verified, honestly-gated
**idea finder** with human review and real refusal outcomes. The copy sells a different product:
a **single-idea assessment with checkable citations**. Two of the copy's specific promises
("for a single idea you submit", "we explain its source so you can check it yourself") are not
true of what a paying customer receives today.

## Options (owner decision — this is the sales promise)

- **A. Fix the copy** (cheap, DB-only, no deploy): rewrite /report.html to describe the real
  product. Honest immediately; loses the "we assess YOUR idea" pitch.
- **B. Extend the engine toward the copy** (code + owner deploys the binary): add a first step
  that assesses the submitted idea itself (the cut/verify/score machinery is reusable on a
  customer idea), and carry web-search source URLs into the rendered report so "check it
  yourself" becomes true. Larger, but the copy is arguably the better product.
- **C. Both, staged**: copy now (A), engine later (B), copy updated again when B ships.

Nothing here needed to block the rest of the day's work; the pipeline pages funnel to
/report.html regardless of which way this lands.
