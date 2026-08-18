# SUMMARY — finetuning.uk, 2026-08-18

**What we're trying to do.** Sell fine-tuning as a paid service on
finetuning.uk: a business sends a few hundred examples of how it writes or
answers; we train a small AI model in their voice on a GPU rented by the
minute; they get the finished model as a file they own, a before-and-after
comparison, and a booked hour chatting with it. Concierge-first, automated
underneath. The agreed message: **"your company's voice, in a model you own."**

**Where we've come from.** The training machinery was built in June and sat
unproven; August found and fixed four serious provisioning defects, then proved
the whole pipeline on real hardware in two days of cheap, measured rehearsals —
training, conversion to the playground's format, and the playground itself,
every stage timed. A first failed conversion attempt was cleaned up by the
automatic reaper's first-ever real intervention, which worked unattended,
exactly as designed.

**What we've done.** Phase 0 is complete and the vendor's invoice confirmed our
cost tracking to the cent: the entire programme of rehearsals cost $1.12 of
GPU. A customer-shaped job costs about £1 of compute and 50 minutes; a
playground machine is conversational three and a half minutes after we press
go, and sub-second once warm. Commercial direction is set: start at **£99**
(business audience, low hundreds credible), positioned like the expensive
agencies but priced as an invitation — believability carried by describing the
process in detail, which we can do because every step is measured. The
competitive research found our exact cell — genuine fine-tune, you own the
file, under £100, no technical skill needed — unoccupied. The two site threads
(service backend and front end) merged into one.

**Where we are now.** Engineering is ahead of the shop window. The site's
automated design-repair path was found to complete "repairs" without repairing;
the diagnosis loop refused my first broad explanation (correctly) and the
narrow, code-verified cause is filed: repair-type work items have no registered
verifier, so completion is accepted unchecked (`bugs_open/302`). The offer
page's words are deliberately not written yet: the copy-quality thread has been
asked how to encode the friendly, expansive, glossary-backed register into the
site's own specification, and the offer-analysis thread has been asked which
benefit leads. One claim was caught before it could reach copy — "a real
person checks every run" is not yet an operational promise and is marked
unverified.

**Where we're going.** Fix 302 so repairs are provable; get the two specialist
answers and seed the offer page through the framework in the agreed register;
then Phase 1 — the page, the £99 payment link, and the first concierge
customers, whose end-to-end journeys tell us whether the price and the
positioning hold.
