# DESIGN 2026-08-25 — the sample datasets a prospect can try before handing over their own documents

**Owner decision 5, 2026-08-25, verbatim:** *"yes we should offer a range of datasets, thinking
about what they might use us for like for finding their 'copywriting voice' for their emails, or
their copy structure or for their style of copy or for perhaps a dozen other tasks they might use
us for, we could give them example data and example honest worked examples."*

So: **yes, a range, keyed to TASKS rather than to industries**, each shipping example data *and* a
worked example. This file is the design. **It deliberately does not generate any data yet** — see
§4, which names the one thing only the owner can answer and which gates every dataset below.

## 1. What a "dataset" is here, concretely

One dataset = three artefacts, and the third is the one that does the selling:

1. **The training set** — 60–200 short paired examples in the format the trainer already takes.
2. **A held-out set** — 10 examples never trained on, so the worked example can be honest about
   generalisation rather than showing the model reciting.
3. **The worked example** — the same prompt run through the base model and the fine-tuned model,
   both outputs shown verbatim, **including where the fine-tuned one is no better**. That last
   clause is the owner's word "honest" and it is the whole point: a page of cherry-picked wins is
   the register he has already rejected twice.

## 2. The six starters

Ordered by how close they sit to what he actually named. Six rather than twelve because six can be
built well and a dozen thin ones is the same mistake as a dozen thin pages.

| # | task | training pair | why a prospect cares |
|---|---|---|---|
| 1 | **Email copywriting voice** | a brief → an email written in the house voice | his first named example; the most common "we all write differently" complaint |
| 2 | **Copy structure** | a raw content dump → the same content in the house section order | teaches shape, not words — the clearest demonstration that a model learns FORM |
| 3 | **Copy style** | a stiff draft → the house-style rewrite | the nearest thing to "make it sound like us" |
| 4 | **Support-reply tone** | an inbound question → the reply as this company answers it | the highest-volume repetitive writing in most SMEs |
| 5 | **Product-description house style** | spec bullets → a description in the range's voice | e-commerce, and it is measurable (length, attribute order) |
| 6 | **Internal-doc summarisation** | a long document → the summary in the house format | the "we drown in documents" case, and it needs no brand voice at all |

Numbers 2 and 6 are deliberately not voice tasks. If every demo is "sound like us", a prospect
whose problem is structure or volume never sees themselves.

## 3. What makes the worked example honest rather than promotional

- Show the **base-model output first**, unedited. On several of these it will be perfectly decent,
  and saying so is what makes the rest credible.
- Show at least one case where fine-tuning **did not help**, and say why (usually: the task was
  instruction-following, not style, and a prompt would have done).
- State the **training cost in real terms** for that dataset — rows, minutes, the £99 it sits
  inside — so the demo cannot imply a bigger job than it was.
- No quantified outcome claims. `evidence_base.governing_rule` on this site already forbids any
  quantified client outcome that the owner has not attested and that is not a registered fact, and
  a demo page is not an exemption.

## 4. ⚠ THE BLOCKING QUESTION — whose data, and it is the owner's alone

Decision 5 as parked asked *"and if so, whose/what data"*. His answer settled **whether** and
**what for**; it did not settle **provenance**, and nothing below can be built without it, because
every option carries a different promise:

| option | what it costs | what it risks |
|---|---|---|
| **A. Synthetic** — generate a fictional company and write both sides | most work; no licensing question at all | a fabricated company is exactly the "AI sounding" register he rejected, unless written carefully |
| **B. Our own material** — finetuning.uk's / websy's real emails, replies and copy | least work, unambiguously ours, and honest by construction | exposes our own internal writing; he may not want that public |
| **C. Public-domain / open-licensed corpora** | free and defensible | reads as generic, and the licence must be carried on the page |
| **D. A customer's, with permission** | the most persuasive by far | needs a customer, a written permission, and it cannot be first |

**Recommendation: B for datasets 1, 3, 4 and 6, A for 2 and 5.** B is honest by construction — a
model trained on our own replies demonstrating our own voice is the claim, not a simulation of one
— and it needs no permission from anyone but him. A covers the two where our own material does not
exist in enough volume.

**Not proceeding until he answers**, because generating data under the wrong provenance is work
that has to be thrown away, and publishing someone else's text is not a mistake that can be
re-rendered out.

## 5. Sequencing

1. Owner answers §4.
2. Build datasets + held-out sets + worked examples as **files** (data, not site copy — the lane's
   standing copy hold does not bite them).
3. Run each through the real trainer, so every worked example is a real run and not a mock-up.
4. The page that presents them waits for `copy_quality_two_stage` to ship its register
   improvements — same hold as everything else outward-facing on this site.

## 6. What is already registered and can be stated on that page

`evidence_base.facts[]` now carries the four operational facts from his 2026-08-25 answers, so the
demo page can state them without a fresh attestation: `ft-booking-hours` (9–5 UK weekdays, other by
arrangement), `ft-deletion-window` (deleted within a week of a request), `ft-retention-default` (30
days after handover), `ft-playground-hour` (one hour included, expires 30 days).
