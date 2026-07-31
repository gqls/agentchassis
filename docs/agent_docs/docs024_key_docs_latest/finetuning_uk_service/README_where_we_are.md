# finetuning.uk service — where we are (append-only, newest at the bottom)

---

## 31 July 2026 — the workstream opens with a settled design

We want finetuning.uk to stop just talking about fine-tuning and actually do it:
a customer pays a few pounds, sends us a small dataset (or picks one of ours),
and gets back a genuinely fine-tuned small model — a before-and-after comparison,
the model file itself, and a booked hour or two on a GPU to chat with it. It is a
demonstration priced to cover costs, not a business trying to compete with the
big platforms, and the site will say so plainly.

The good news from today's dig: almost everything hard already exists. The GPU
training pipeline was built in May and June and trained a real (large) model
once; the payment machinery was proven on idea.uk in July with a real card
payment; and there is already a safe, approved way for an outside machine to
hand work to the cluster without the cluster ever being reachable from the
internet. What has never existed is the product surface joining them.

Decisions made today, with you: customers can bring their own data or use a
sample of ours; the try-it-out playground runs on a rented GPU for hours you or
the customer name — nothing runs overnight, because we store everything and can
start again from storage; the public-facing part lives on the little "island"
server that already runs our public tools behind its shields, so the main
cluster stays completely unreachable; and we go concierge-first — a page and a
payment link, orders run by hand — before automating anything.

Two things block the first real run. You need to create a new Thunder Compute
API token (the old one is being replaced) — the exact steps are written up in
the runbook, step 1. And we then do one cheap rehearsal run with a small model
to measure what it actually costs and how long it takes, because the price
should come from a measurement, not a guess. That rehearsal doubles as the
missing proof that a finished training run's output is safely in storage before
the machine that made it is destroyed — a proof the training system has been
waiting on since June.

One honest caveat recorded: taking card payments on the island changes what the
island was — it was built to hold no important credentials at all. Stripe keys
are money keys, not keys to our systems, and the cluster stays sealed either
way; but it is a real change and it is written down as one.
