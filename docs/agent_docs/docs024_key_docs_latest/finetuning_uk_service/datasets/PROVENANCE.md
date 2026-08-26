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
| 1 | email copywriting voice | our own material | ⛔ **BLOCKED by the conflict above** |
| 2 | copy structure | synthetic | ✅ **BUILT** — 80 training + 10 held-out, validated |
| 3 | copy style | our own material | ⛔ **BLOCKED by the conflict above** |
| 4 | support-reply tone | our own material | ⛔ **BLOCKED** — and separately, we hold no real support replies; he would need to send them |
| 5 | product-description house style | synthetic | ⏳ next, unblocked |
| 6 | internal-doc summarisation | our own material | ⏳ unblocked — our own long docs are a fine source for a FORMAT target |

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
