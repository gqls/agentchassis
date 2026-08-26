# Sample datasets — provenance, status, and one conflict in the approved plan

**Owner decision 2026-08-26:** *"Sample datasets as you suggest"* — approving
`DESIGN_2026-08-25_sample_datasets.md` §4's recommendation: **our own material** for datasets 1, 3,
4 and 6; **synthetic** for 2 and 5.

## ⚠ THE CONFLICT, found while building, and it affects three of the four he approved

"Our own material" means finetuning.uk's published copy. **That copy is the register the owner
rejected on 2026-08-24 and again on 2026-08-25** — *"very AI sounding"*, *"so methodical like AI"*,
*"the whole site could be rewritten in better language"*.

So for the three datasets whose TARGET is a voice — **1 (email copywriting voice)**, **3 (copy
style)**, **4 (support-reply tone)** — training on our own published copy would teach a model to
produce exactly the register he has twice told us to stop producing, and the worked example would
then demonstrate it to prospects. Our own material is the honest source; it is currently the wrong
teacher.

**This does not affect dataset 6** (internal-doc summarisation): its target is a summary in a house
FORMAT, not a voice, and our own long documents are fine for that.

> ## ✅ RESOLVED 2026-08-26 — permission GRANTED, and it was never the constraint
>
> **Owner, 2026-08-26: *"yes we can use my writing"*** — option 1 below, approved.
>
> **And then the material turned out not to exist.** See §"The corpus measurement" at the foot of
> this file. The honest total of attributable owner prose in the estate is **117 words** plus **one**
> before/after pair. The three voice datasets are still blocked — **on material, not on permission**,
> and only he can unblock them now by sending it. What to send is listed at the end.

**Three ways forward, and this one is the owner's call:**

1. **Use his own writing** — the `README_where_we_are.md` entries and his own messages are plain,
   direct and human, and are the voice the whole copy-quality lane is trying to reach. It is the
   only material we hold that is demonstrably in a voice he approves. **Needs his say-so**: it is
   his personal writing and these datasets are customer-facing.
2. **Wait** for `copy_quality_two_stage` to ship its improvements, rewrite the site, then harvest
   the rewritten copy. Correct, and it puts the three voice datasets behind the same hold as
   everything else outward-facing.
3. **Synthesise them too.** Cheapest and worst: a demo of "sound like us" built from invented
   material demonstrates nothing about us.

Recommendation: **1 if he is willing, otherwise 2.** Not proceeding on any of the three until he says.

## Status of the six

| # | dataset | provenance | state |
|---|---|---|---|
| 1 | email copywriting voice | our own material | ✅ **BUILT** 2026-08-26 — 26 training + 5 held-out, from his own emails; briefs describe the situation only, never the phrasing |
| 2 | copy structure | synthetic | ✅ **BUILT** — 80 training + 10 held-out, validated |
| 3 | copy style | our own material | ✅ **BUILT** 2026-08-26 — 13 training + 3 held-out. Target verbatim his; the INPUT is written deliberately in the register he rejected, carrying the same facts, so the model learns to rewrite and not to delete |
| 4 | support-reply tone | our own material | ✅ **BUILT** 2026-08-26 — 16 training + 4 held-out. Target verbatim his; the INBOUND side is RECONSTRUCTED (the originals were never supplied) and labelled as such |
| 5 | product-description house style | synthetic | ✅ **BUILT** 2026-08-26 — 80 training + 10 held-out, validated. Target register checked clean of `rather than` / `not just` / `isn't`: the target side IS the register, so it had to be written in the one we want |
| 6 | internal-doc summarisation | our own material | ✅ **BUILT** 2026-08-26 — 8 training + 2 held-out. ⚠ Inputs are his articles but the TARGETS are written by this lane to a fixed shape, so it teaches STRUCTURE, not his register. A worked example from it must not claim otherwise |

## The harness, and why it refuses things

`build_dataset.py` validates before anything is uploaded. Every refusal it makes is induced in its
own `--self-test`, plus a positive control that a clean dataset still builds — a validator that has
only ever been seen to pass is a validator nobody has seen.

It refuses: malformed JSON · anything other than exactly one user turn then one assistant turn ·
an **empty or whitespace assistant turn** (trains the model to say nothing, and reads as a good row
in every dashboard) · a row over 12,000 characters · a dataset with no declared provenance.

**The length cap is the one with a price attached.** The phase-0 run (`RESULTS_2026-08-15`)
uploaded 300 rows and trained **295** — five were dropped by the trainer's response-marker filter
and the drop surfaced only as a row count in a stage summary, *after* the GPU was paid for. That is
the failure this refuses for free.

`provenance` is a required field precisely because no machine can check it. Whose words these are
is a question a person has to answer before anything is published.

---

## The corpus measurement, 2026-08-26 — and the near-miss that produced it

Permission was granted, so the next step was to find his writing. It is not there, and the way that
was nearly missed is worth more than the finding.

**The wrong measurement.** 268 `README_where_we_are.md` files (**5.85 MB**) are "the owner's
document" by convention. Extracting every quoted string from any paragraph mentioning the owner gave
**2,580 utterances / 31,164 words across 259 lanes** — comfortably a training set.

**It was almost entirely OUR OWN prose.** An unfiltered sample, every 60th row:
`feed items collected: 11,513` · `does this exact string appear in the facts?` · `, because our own
tooling treats a trailing`. Session writing, error strings and technical fragments that sessions had
quoted **for emphasis while describing their own work**. The extractor tested for QUOTING; I was
reading it as AUTHORSHIP. Caught by printing the sample rather than the total — a count cannot show
you that every row is the wrong kind of thing. Logged in `WRONG_CALLS.md` 2026-08-26.

**The honest figures `[MEASURED 2026-08-26]`:**

| source | attributable owner prose |
|---|---|
| explicitly-marked owner speech (`**OWNER … verbatim:** "…"`) fleet-wide | **16 utterances, 117 words** — all short instructions ("keep the explanatory copy") |
| `travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v2.md` | **1** genuine before/after pair he judged — preserved as `_owner_voice_seed/` |
| that exercise's full ground truth (`held_out_ground_truth.md`) | its own README says *"not copied into the repo, ephemeral"* — gone |
| `pitch_pdf_source/deck.html` (3,371 words) | **unattributable** — carries NEITHER the v1 phrase nor the hand-edited target phrase, so it is a third version of unknown authorship |
| everything else | ours, or unattributable |

**117 words and one pair is not a training set for anything.** Composing the rest and calling it his
voice would be precisely the fabrication these datasets exist to avoid — and it would be published
to prospects under his name.

## What he needs to send, per dataset

Volumes are the design's own (60–200 pairs); 30–60 real items is enough to start and can be extended.

| # | dataset | what to send |
|---|---|---|
| 1 | email copywriting voice | **30–60 emails he has actually written** — any subject, business or otherwise. The brief side can be reconstructed from each email; the email itself is the target |
| 3 | copy style | **30–60 pieces of his own copy**, ideally each with the stiffer draft it replaced. `_owner_voice_seed/` is exactly one pair of this shape and is the template — about forty more of those |
| 4 | support-reply tone | **30–60 real replies to customers or enquiries**, with the inbound message where possible |

Datasets **2** (built), **5** and **6** need none of this and are unaffected.

---

## All six built, 2026-08-26 — and the honest sizes

| # | dataset | training | held out | teaches |
|---|---|---|---|---|
| 1 | email voice | 26 | 5 | his register |
| 2 | copy structure | 80 | 10 | shape |
| 3 | copy style | 13 | 3 | his register |
| 4 | reply tone | 16 | 4 | his register |
| 5 | product descriptions | 80 | 10 | a plain register |
| 6 | doc summaries | 8 | 2 | structure |

**The three voice datasets are SMALL — 26, 13 and 16 rows against a design target of 60–200 — and
that is a real limit, not a rounding.** It is the direct consequence of the corpus measurement: the
material is what he had to hand, and padding it with invented correspondence would destroy the one
property that makes these worth showing a customer. **More of his real emails and articles is the
only thing that raises them.**

The two synthetic sets (2 and 5) are full size because generation is free there and costs no honesty.

## Before any of this trains

Every dataset passes `build_dataset.py`, which refuses malformed rows, an empty assistant turn, a
row over the length cap that silently cost the phase-0 run five of its 300, and a missing
provenance declaration. Re-run it after any edit — it is cheap and it has already caught things.
