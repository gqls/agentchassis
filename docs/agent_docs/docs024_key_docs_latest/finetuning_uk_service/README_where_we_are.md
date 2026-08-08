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

## 3 August 2026 — the key in the cluster doesn't work, and the safety check now does

Two things to report, and they point in opposite directions.

The first is a problem. You said we now have a Thunder Compute key, so I went and
checked rather than taking it on trust, and the key sitting in the cluster is
being refused by Thunder — it comes back "unauthorised". I checked it from inside
the machine that would actually use it, so this is not a question of my laptop or
my typing. I also made sure the *check* wasn't the thing at fault: I compared what
I sent against what our own code sends, and I tried the same request with a
deliberately fake key and with no key at all. Both were refused in exactly the
same way, which is how I know the refusal means "this key is bad" rather than
"you asked the wrong question". And the key in the cluster is not the one you
generated on the 31st — the two don't match — so I think the new one simply never
got pasted in.

The reassuring half of that: a key that can't authenticate can't rent a GPU
either. Nothing can be running up a bill right now, and I confirmed separately
that we have nothing running at all and have spent nothing in the last day. What
it *does* cost us is the one check that could spot a machine running at Thunder
that our own records know nothing about — that check has to ask Thunder directly,
so it's blind until the key works. That is the reason to fix it, more than the
GPU work being blocked.

The second thing is progress. The safety net you asked about — the automatic
sweep that shuts down forgotten machines — has been repaired and the repair is
live. You'll remember the finding: it had been running every fifteen minutes for
months and had never actually shut anything down, because the way it looked for
machines missed three of the situations that actually cost money. That fix is now
applied and I've confirmed all three are covered.

Applying a fix isn't the same as proving it works, so I'm now doing exactly that:
I've put a fake, deliberately-fake-named machine into our records, backdated so
the sweep should consider it long overdue, and I'm watching to see whether the
sweep finds it and acts. Two things make this the safest possible moment to run
that test — we have nothing real running, and the key doesn't work, so the test
physically cannot destroy anything of ours at Thunder. The flip side is that it
proves the half that was broken (noticing, and deciding to act) but not the final
step of actually telling Thunder to shut down. That last step has to be re-proved
once a working key is in.

One more thing worth knowing, from your other thread rather than mine. That lane
found that the machinery which is supposed to pick up detected problems and act
on them has been switched off since May — problems get spotted and then sit
there. My plan for taking customer orders was going to ride on that same
machinery, so I've noted that it would have inherited a queue nobody empties.
Better to find that now than after someone has paid.

Nothing I've touched today goes anywhere near the site itself, so your other
thread and this one aren't going to collide.

## 8 August 2026 — the key works, the safety net is proven, and the prices came down

Three pieces of good news, each checked rather than hoped.

First, the key. Your new Thunder Compute token is in and working — I checked
from inside the actual machine that uses it, and Thunder answered properly
where a week ago it refused. The important change is HOW it got there: the key
now lives in the Terraform configuration, which is the system that owns that
secret. That also finally explains last week's mystery — the key we thought we
had installed kept not working because Terraform, quite correctly, kept putting
its own older key back. Fighting the mechanism looked exactly like a broken
key. The procedure is written down now, and it starts from the file where you
keep the token.

Second, the safety net. The automatic sweep that shuts down forgotten GPU
machines has now been proven all the way to the end. Last week it correctly
noticed an overdue machine but choked on our own database before it could act
— and even that test couldn't reach Thunder itself, because the key was bad.
Today, with the working key, I staged another fake overdue machine and watched
the sweep notice it, call Thunder for real, handle Thunder's "no such machine"
answer gracefully, work out the cost, and close the record — thirty-two
seconds from trigger to done. The database bug it tripped over last week has
been fixed, reviewed and approved (the review panel passed it first time), and
that fix goes live with the next software release. Total spend across all this
testing: nothing. No real machine was ever created.

Third, the prices. The A100 we budgeted at $1.80 an hour is now $1.09. Better
still, there's a smaller card (an A6000, plenty for the try-it-out playground)
at thirty-five cents an hour — so a customer's two-hour playground session
costs us about seventy cents, comfortably inside a few-pounds price. One
correction to my earlier plan: the mid-range card I named ("L40S") doesn't
actually exist in Thunder's catalogue — the real options are the A6000 and the
L40, and the A6000 is cheaper anyway.

What's next is the rehearsal run: one small model trained end to end, timed
and priced for real. That was always the gate before putting a price on the
page, and nothing blocks it any more.

---

**2026-08-08, evening.** The software release went out this afternoon, and the
database fix I mentioned this morning is now confirmed working in production.
I proved it the same way as before: staged a fake overdue machine — this time
deliberately missing the network address that used to crash the sweep — and
watched the whole thing complete in about thirty seconds: the sweep picked it
up, the lookup that used to fall over sailed through, Thunder was asked for
real and answered "no such machine", the cost was worked out and the record
closed. Cleaned up the test row afterwards; still nothing spent across all of
this testing. One honest footnote: the usual trick for proving a release
contains a particular change (searching the running program for tell-tale
text) doesn't work for this change, because it altered only how data is typed,
not any message text — so the live test above isn't a nice-to-have, it's the
only real proof, which is exactly why it was designed in. The cleanup sweep is
now fully proven end to end, including the awkward case. The remaining gap is
unchanged: a machine Thunder is billing us for that our records never heard of
would still be invisible — that reconciliation check is the next build item,
and the ten-second manual check remains the net until then. Otherwise the lane
is clear for the rehearsal run, which wants you around when it happens.
