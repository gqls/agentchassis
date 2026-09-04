# PLAN (PRELIMINARY) 2026-09-04 — handwriting models: read my hand, write my hand, sell the font

**Status: a starting draft for the owner to expand in his own lane.** Nothing here is built, dispatched
or costed against a real quote. Written at his request, 2026-09-04: *"I'd like to do the handwriting
model, maybe all three, and we can sell the font too. How would we learn the customer's pen strokes? To
begin with the recognition, then the image based then the pen strokes if we can do it, and then the
font. I'll start a lane, please write a preliminary plan for it that I can expand there."*

Everything marked `[ASSUMED]` or `[TO VERIFY]` is exactly that. The one thing this plan asserts
strongly is the §6 refusal, and it is the part to decide first.

---

## 1. What the product actually is

Four artefacts, in the owner's order, each usable on its own and each feeding the next:

| # | Artefact | What the customer does | What they get |
|---|---|---|---|
| 1 | **Recogniser** | photographs a page of their notebook | the text, typed |
| 2 | **Image writer** | sends a few photographed words | a picture of new text in their hand |
| 3 | **Stroke writer** | writes for ten minutes on a tablet | new text drawn stroke by stroke, at any size, in any colour, as SVG |
| 4 | **Font** | nothing more | a real .ttf/.otf they install and keep |

And the showpiece, which is why this belongs on finetuning.uk rather than anywhere else: **the voice
model writes the words, the stroke model writes the letters.** "Write a thank-you note to a customer
who complained, in my voice, in my hand" produces a page that looks handwritten by that person and
sounds like them. Nobody else is selling that pair.

## 2. Phase 1 — recognition (read my handwriting)

The easy one, and the one that proves the data-collection loop.

- **Model:** fine-tune an existing open handwritten-text-recognition model. `TrOCR-base-handwritten`
  (Microsoft, MIT-licensed weights on Hugging Face) is the obvious start; a CRNN + CTC model is the
  lighter alternative. `[TO VERIFY]` the exact licence of any checkpoint before it ships in a paid product.
- **Data, and why it is free:** the customer copies out a passage **we** supply. Because we chose the
  passage, we already have the ground truth — no transcription labour. 200–400 lines is a normal
  fine-tuning set for one hand; that is roughly ten to fifteen pages, an hour of writing.
- **Preprocessing:** page photo → deskew → line segmentation → per-line images. All classical, all open
  (OpenCV, or a small line-detector).
- **CPU:** `[ASSUMED, to measure]` TrOCR-base is ~330M parameters, so a line is a second or two on the
  Hetzner box; the whole page a minute. Acceptable for "upload a notebook page, get a document back".
- **⚠ Licence trap:** IAM Handwriting Database, the standard academic set everyone fine-tunes on, is
  **free for non-commercial research only**. It must not be in the training set of anything we sell.
  Base checkpoints trained on it are a grey area — `[TO VERIFY]` per checkpoint, and prefer one whose
  training data is stated.

## 3. Phase 2 — image-based writing (write my hand, from photographs)

- **Model:** few-shot styled handwriting generation — give the model 15–50 photographed words in a
  hand, plus target text, and it renders new words in that hand. This is an active research area with
  several open implementations; `[TO VERIFY]` which have permissive licences and usable weights.
- **Data:** the same photographs as phase 1, cropped to words. No new collection.
- **Weakness, stated plainly:** these models produce *pictures*. Letter joins are where they fail, the
  output is a fixed resolution, and you cannot recolour or resize it cleanly. It is the middle rung, and
  its real value is that it needs no tablet.
- **CPU:** heaviest of the four. `[TO MEASURE]` — the 3.8 GB demo box may not hold it; this phase may
  need the bigger box or a per-job GPU minute.

## 4. Phase 3 — stroke-based writing, and the owner's actual question

**"How would we learn the customer's pen strokes?"** — by recording the pen, not the page. The answer
is a web page, which is a thing this estate already knows how to build and host.

**Collection.** A canvas in the browser, using **Pointer Events**. Every `pointermove` gives x, y and a
timestamp; a stylus also gives **pressure** and **tilt**. That covers an iPad with an Apple Pencil, a
Samsung with an S Pen, a Wacom tablet, a Surface pen, and a finger on any phone (worse, but usable).
No app, no install, no store review. The same page can run on the customer's own device during a
booked hour.

**What we ask them to write, and why:** roughly fifteen minutes, in three parts.
1. **Coverage** — pangrams and a word list chosen so every letter appears in several *contexts*
   (start, middle, end, and after the letters that change its shape). Handwriting is contextual: the
   `o` in "book" is not the `o` in "of".
2. **Joins** — the common bigrams and trigrams (th, er, ing, tion, qu…), because the join is the style.
3. **Their own material** — name, sign-off, a short note they would actually send. This is what makes
   the demo feel like them.

**What we store:** each stroke as a sequence of points, resampled to a fixed rate (say 100 Hz),
normalised for size against the writing line, and reduced to offsets — `(dx, dy, pen_up)` per point,
the standard representation since Graves's 2013 handwriting network. A page of writing is tens of
kilobytes. It is text, not images, which is why this rung is the good one.

**Model.** A small sequence model with a mixture-density output over the next `(dx, dy, pen_up)`,
conditioned on the character sequence — a few million parameters, not a few hundred million. Trained
once on a permissively-licensed base corpus, then **primed or lightly fine-tuned per customer**, which
is what makes ten minutes of writing enough. Generation is a second or less on CPU.

**Output is a path, not a picture.** SVG at any size, any pen width, any ink colour, and it can be
animated to write itself on screen. That last is worth a lot on a website selling this.

**⚠ Same licence trap, harder:** the standard on-line stroke corpus (IAM-OnDB) carries the same
non-commercial research licence. Either the base model is trained on data we can use commercially, or
we collect our own — a few paid contributors writing for an hour each is a real and affordable option,
and it is also the honest one to describe on the site.

## 5. Phase 4 — the font, which is the thing people will actually buy

- From strokes (best) or images (workable), build a genuine **.ttf/.otf**: outline each glyph, add
  **several variants per letter** and contextual alternates so the same word does not repeat identically,
  plus ligatures for the common joins.
- Tooling is open and mature (`fonttools`, FontForge). This is a one-off offline job per customer, no
  serving cost, no model at inference time.
- **A font is not the stroke model, and both are sellable:** a font repeats a fixed set of shapes; the
  stroke model draws every letter afresh. Price them differently — the font is the keepsake, the model
  is the service.

## 6. The refusal that has to be decided before anything is built

**A handwriting synthesiser is a signature forger unless it is deliberately not one.** This product
must, from the first line of code:

- **exclude signatures** — do not accept a signature as training material, do not generate one on
  request, and say so on the page;
- treat a person's handwriting as **personal data of an identifying kind**: explicit consent, a stated
  retention period, deletion on request, and no training of a shared model on one customer's hand
  without separate permission;
- **watermark or otherwise mark generated images** `[OPEN QUESTION]` — worth deciding, because "this
  page was written by a machine in my hand" is a claim someone will need to make later;
- take a view on **whose** handwriting may be uploaded: their own only, or anyone's? The honest answer is
  their own, verified as far as we reasonably can.

This is not a compliance footnote. It is the difference between a charming product and one that ends up
in a story about forged documents, and the owner should decide the stance before a lane spends money.

## 7. What it would cost to find out

| Step | Spend | Time |
|---|---|---|
| Collect the owner's own hand (phase 1 data) | £0 | ~1 hour of his writing |
| Fine-tune the recogniser once, a6000 | `[EST]` under £1 of GPU | an afternoon |
| Build the stroke-collection page | £0 (framework) | a day |
| Collect strokes, prime a base stroke model | £0–small | ~15 min of his writing |
| Base stroke model on licensable data | the real unknown | `[TO SCOPE]` |
| Font pipeline from strokes | £0 (open tools) | a few days |

**Recommended first move, which costs an hour and settles the whole idea:** the owner writes out the
supplied passage, we fine-tune the recogniser, and he photographs a page of his own notebook and gets it
back as text. If that works, the collection loop, the data format and the CPU story are all proven, and
phases 2–4 are variations on a pipeline that already exists.

## 8. Where this sits

Recorded as a product idea in `finetuning_uk_service/PLAN_2026-07-31_finetuning_uk_service.md`
(DIRECTION, 2026-09-03). This file is the preliminary plan for the owner's own lane; when that lane
opens, move this file into it and leave a pointer here. Nothing in it depends on the finetuning.uk
playground work, and nothing in the playground work waits on it.
