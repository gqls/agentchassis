# finetuning.uk service — milestone summary, 31 July 2026 (second of the day)

*A new file, not an edit of `SUMMARY_2026-07-31`. That one closed on the design
being approved. This one exists because the preparation is now finished and
committed, the choice of models has been settled by licence evidence, and the
whole workstream has narrowed to a single action only the owner can take.*

## What we're trying to do

Make finetuning.uk sell, in miniature, the thing its name promises: for a few
pounds a customer gets a genuinely fine-tuned small model — trained on their own
data or a sample of ours, delivered as a before-and-after comparison, the model
file itself, and a booked hour or two on a GPU to chat with the result. The
price covers computer time and a little of the owner's time. It is a
demonstration, and the site will say so plainly.

## Where we've come from

This morning the design was settled and approved. The plan leans on three
things the platform already proved separately: the GPU training pipeline built
in the spring (used once, for real, on a large model); the payment machinery
that took its first real card payment on a sibling site in July; and the
approved pattern by which an outside server hands work to the cluster without
the cluster ever being reachable from the internet. The owner set four
directions: front everything with the already-guarded public tools server, get
a fresh Thunder Compute token before any GPU starts, treat the GPU as strictly
stop-and-start with everything worth keeping stored by us, and run the
try-it-out playground on a GPU in named hours rather than keeping anything on
overnight.

## What we've done

Since the morning, the groundwork. The workstream's working documents exist and
are committed. The training script now accepts a small model as a setting
rather than a rewrite, with its defaults untouched — and, deliberately, the
live copy in storage has not been replaced yet, so nothing about the running
system has changed; that swap happens as part of the first rehearsal. The exact
steps for replacing the Thunder token are written out and were checked against
the live cluster, down to which secret holds the key and how to prove the new
token works from inside the system rather than from a laptop.

And the model menu is now decided by evidence rather than preference. We
checked the licence of every candidate in writing. Three come out clean and
will be the opening offer: Mistral 7B, Phi-3.5-mini, and the small Qwen — three
different sizes, all permissively licensed, and a Mistral already runs on our
own machines today. Two traps were found and written down: the 3-billion
version of Qwen — the one size you would naturally reach for — carries a
research-only licence its siblings do not, and Meta's Llama would oblige every
customer's downloaded model to be renamed "Llama-something" and our site to
display "Built with Llama". Llama is out for now; it isn't worth putting those
obligations on a customer's deliverable.

## Where we are now

Everything that can be done without spending money is done. The whole
workstream now waits on one thing: **the owner minting a new Thunder Compute
token**, for which the runbook has ready-to-paste steps. Nothing else is in the
way. No price has been set, no page has been written, and that is deliberate —
both depend on measurements we have not yet made.

## Where we're going

Token first. Then one cheap rehearsal — a small model, likely twenty-odd
minutes of GPU, about a pound — measured end to end: what training costs, how
long the playground takes to wake up, what the cheaper GPUs rent for. That same
run settles a proof the training pipeline has owed itself since June: that a
finished run's output is safely in storage before the machine that made it is
destroyed. Price comes from those measurements. Then a page on the site and a
plain payment link, first orders run by hand and capped at ten; automation only
after real strangers have paid. Open for the owner alongside the token: the
final price, which sample datasets we are proud to publish, and confirming that
card payments living on the tools server is a change he accepts.

## The one-sentence version

The preparation is finished and committed, the three models are chosen on
licence evidence rather than habit, and the entire project is now queued behind
one five-minute action — a fresh GPU token — after which a one-pound rehearsal
run sets the price.
