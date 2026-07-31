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

## 31 July 2026, later — everything that doesn't cost money is done

Finished the preparation and committed it. The working documents exist; the
training script takes a small model as a setting now rather than needing a
rewrite, with its old behaviour untouched; and the token-replacement steps are
written out and checked against the live cluster — including the detail that the
secret holding it carries eleven other providers' keys, so it must be patched
one key at a time rather than recreated.

The one substantive decision made since this morning is which models we offer,
and it was settled by reading licences rather than by preference. Three are
clean and will be the opening menu: Mistral 7B, Phi-3.5-mini and the small Qwen
— three sizes, all permissive, and a Mistral already runs on our own machines.
Two things we would plausibly have reached for turned out to be traps. The
three-billion Qwen carries a research-only licence that its sibling sizes do
not, which is exactly the sort of thing you find out afterwards. And Meta's
Llama would require every customer's downloaded model to be renamed
"Llama-something" and our own site to display "Built with Llama" — obligations
we would be passing on to a customer for a five-minute demo. Llama is out for
now.

Deliberately not done: no price, no page. Both wait on the rehearsal run,
because a price guessed before measuring is just a number we would have to
defend later.

So it now sits on you: mint the new Thunder Compute token, and the runbook's
first section has the steps ready to paste. After that the next spend is about a
pound.
