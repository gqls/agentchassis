# SUMMARY — finetuning.uk service, 2026-08-15

**What we're trying to do.** Sell fine-tuning as a paid service on finetuning.uk:
a customer brings a few hundred examples of how they want a model to write or
answer, we rent a GPU by the minute, train a small adapter on their data, and
give them back a model that talks their way — with a playground hour to try it.
Concierge first (a Stripe payment link and hand-holding for the first ~10
orders), automation after.

**Where we've come from.** The training machinery was built in June against a
70-billion-parameter model and then sat unproven: no run had ever finished
end-to-end with its output safely off the machine. In August we found the GPU
provisioning layer had four serious defects — the worst being that one request
could quietly build several billable machines — so the whole thing was paused
fleet-wide while they were fixed. Those fixes went through review, shipped, and
this morning were proven working on real rented hardware for pennies.

**What we've done (today, mostly).** This afternoon we ran the first complete
training job in the service's intended shape: a small open model (SmolLM2-1.7B,
Apache-licensed), 300 training rows, on the cheap A6000 card. It worked end to
end: the machine sized itself correctly from the vendor's own catalogue, the
training data rendered in the model's real chat format (the silent-mistraining
bug fixed on the 12th earned its keep — only 5 of 300 rows dropped, correctly,
for length), training converged (loss 1.41 → 0.73 over three epochs, 23
minutes), and the finished adapter — 68 MB — was uploaded to our storage and
verified there independently before the machine was destroyed. That last step is
the one that matters: it was the missing proof that "job finished" really means
"output is safe", which is what the automated monitor has been waiting on since
June. The run also flushed out three small real defects (a missing directory on
fresh machines, a hardware check that assumed the big model, and a launch
command that could report success while doing nothing) — all fixed and written
into the runbook the same hour.

**Where we are now.** Whole pipeline proven: provision → size → boot → train →
upload → verify → destroy, about 50 minutes end to end, roughly 30 cents of real
GPU cost for a customer-sized job. Provisioning is re-paused (deliberate; it
stays off between runs). Ten-ish minutes of a session's time fires the next run
from the runbook. Three loose ends: the corrected script bundle still needs
re-uploading to storage (blocked by a session permission today); the automated
monitor's enable-switch is ready to flip but that's the owner's call; and the
duplicate-machine guard, though live and tested, has still never been exercised
by a real slow boot — boot times swing from 16 seconds to nearly 5 minutes day
to day, so one day it will be.

**Where we're going.** The remaining Phase 0 items are converting the adapter to
the format the playground serves (GGUF) and timing a playground session on a
rented box — then Phase 0 is closed and pricing can be set from measured costs
(the flat $1.80/hr in our books overstates the A6000 four- to five-fold; the
first real invoice settles it). After that: the offer page and payment link
(Phase 1), which is blocked on coordinating with the thread that owns the
site's front end, and owner decisions on price, playground booking shape, and
sample datasets.
