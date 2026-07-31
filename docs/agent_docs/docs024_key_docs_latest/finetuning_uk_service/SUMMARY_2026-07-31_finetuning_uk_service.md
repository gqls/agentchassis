# finetuning.uk service — milestone summary, 31 July 2026

*First summary of this workstream, written the day it opened — the design
milestone. Plain prose, to be read aloud. Technical entry point:
`PLAN_2026-07-31_finetuning_uk_service.md`.*

## What we're trying to do

Make finetuning.uk sell the thing its name promises, in miniature: for a few
pounds, a customer gets a real fine-tuned small model — trained on their data or
on a sample of ours, delivered as a side-by-side comparison against the base
model, the model file itself, and a booked session on a GPU to talk to it. The
price covers the computer time and a little of the owner's time. The point is
proof, not profit: this site can demonstrate the thing it advises on.

## Where we've come from

Three separate pieces of past work make this cheap. In the spring the platform
built a complete GPU fine-tuning pipeline — rented machine, training script,
results streamed to storage, cost caps and a reaper — and used it once, to train
one large model for about twenty dollars. In July, a sibling site (idea.uk) took
its first real card payment through payment machinery written for exactly this
kind of small paid product. And the platform already has one approved, working
pattern for letting an outside server hand work to the cluster without the
cluster ever being exposed to the internet. None of these were built for this
product, and all three are exactly what it needs.

## What we've done

Dug through all of it and written the plan, which the owner has approved with
four directions of his own: protect the main cluster by fronting everything with
the already-guarded public tools server; get a fresh Thunder Compute token
before any GPU is started; treat the GPU as strictly stop-start, with everything
worth keeping stored by us and reloadable; and run the try-it playground on a
GPU in named hours rather than keeping anything on overnight. The working
documents now exist, the training script has been adjusted so a small model is a
setting rather than a rewrite, and the token-replacement steps are written out
ready.

## Where we are now

Designed, not built. Everything before the first pound: the owner must mint the
new Thunder token, and then one cheap rehearsal run — a small model, maybe
twenty minutes of GPU — gets measured end to end. That rehearsal also settles a
proof the training pipeline itself has been waiting on since June: that a
finished run's output is safely in storage before the machine is destroyed.
Price comes after that measurement, not before.

## Where we're going

Token, rehearsal, measurements. Then a page on the site and a plain payment
link, first orders run by hand, capped at ten. Only after real people have paid
does the automation get built: the order-taking surface on the tools server, the
cluster pulling jobs from it, results flowing back through storage, playground
bookings driving GPU start and stop. The open questions for the owner are the
final price, which sample datasets we are proud to offer, and confirming that
card payments living on the tools server is a change he accepts.

## The one-sentence version

A demonstration fine-tuning service for a few pounds is mostly assembly of
things already built and proven separately — the design is settled and approved,
and the first step is a fresh GPU token followed by one cheap, measured
rehearsal run that also pays off an old debt in the training pipeline.
