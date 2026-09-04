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

Later the same evening: the reconciliation check I called "the next build
item" above is now written. Every six hours the system will ask Thunder
directly "what machines are you charging us for?" and compare that against
our own records — if Thunder is billing for something we have no record of,
it raises a high-priority flag with instructions for whoever sees it (kill
the machine at the console, then work out how it got created without a
record). It deliberately does not try to fix anything itself: a machine we
have no record of is exactly the machine our cleanup tools cannot reach, and
guessing would risk killing something real. It also knows the difference
between a genuine stray and a machine that is simply mid-setup (new machines
get a half-hour of benefit of the doubt — but a machine whose age Thunder
won't state gets flagged regardless, so nothing can hide). The code is
written, reviewed by the automated panel (verdict due shortly), and switches
on with the next software release plus one database step afterwards — until
that first scan is verified, the ten-second bedtime check stands.

One more thing tonight, and it needs you: **the AI account ran out of
credits at half past six this evening** (UK time: about half past seven).
Every automated process that thinks — the review panel, the site builders,
the diagnosis loops, all of it, across every project — has been failing
since that moment. Nothing is broken; it is a billing top-up on the
Anthropic account (Plans & Billing page). The review of tonight's new
billing-watchdog code was mid-flight when the credits ran dry: the first
round had come back asking for one improvement (a fair catch — I was
writing my own database entry instead of using the shared, safer routine;
now fixed), and the second round died on the empty account before any
reviewer saw it. Once you top up, the resubmission is one saved command —
it's written in the technical notes next to this file.

Saturday midday. The credits are topped up (thank you) and the new software
release went out this morning, so the billing watchdog I wrote yesterday is
now switched on for real — and it earned its keep in its first ten minutes,
though not the way I expected. Its very first live run failed: I had written
one wiring label in the wrong place in the configuration, copying the shape
of an older, working component — and it turns out that older component has
carried the same wrong label since it shipped, harmlessly, because nothing
ever needed to read what it labels. Mine did need to. Fixed within the hour,
re-run, and the watchdog now completes cleanly: it asked Thunder directly
what we're being billed for (nothing), checked that against our books
(twenty-three machines, all long since shut down, nothing live), and agreed.
That trap is now written up so nobody copies it again. The review panel's
second round also caught me flat-out wrong about one thing: I'd claimed our
migration folder has no ledger tracking what's been applied — it does, I
just hadn't read the tool's own header — so the watchdog's database change
is now properly recorded there too. Third review round is running now on
the finished, live state. The lane's next big items are unchanged: the paid
rehearsal run (wants you present) and, further out, the page and payment
link once the front-end thread is coordinated with.

Saturday, early afternoon. The review panel approved the billing watchdog —
fourth round, and that closes it. Two reviewers still raised a point each,
neither serious enough to block, and I checked both rather than let them
sit, because each was one command.

The first was a fair worry that turned out fine: the watchdog writes a note
into our documentation table using direct database code rather than going
through the usual software route, and I had only checked that the database
would accept it, not that the software agrees the category is a real one.
It does — "pipeline" is on the approved list. No problem, but I hadn't
actually looked, and the reviewer was right to say so.

The second one caught me properly, and it's the more useful of the two. I
had written that the watchdog was "live, both copies checked" — which was
true, and was the wrong thing to have checked. Our main program runs under
about thirty-four different job names on the cluster, not two, so looking at
the two obvious ones can pass while most of the fleet is running older
software. When I counted properly this morning, the fleet had in fact moved
on to a newer release than the one I tested on — another session shipped it
while I was writing yesterday's notes. So I re-checked against what is
actually running now: all four long-running services carrying the program
have my code, including two I would never have thought to look at because
they're named after completely different things. The nine remaining older
ones are short-lived job pods that finish and disappear, and they carry the
code too, because it first shipped on the release they're pinned to. So
nothing was wrong — but the sentence I'd written couldn't have told me that,
and that's the part worth fixing. It's logged with the other wrong calls.

Worth noting the review credit sorted itself out: because I tagged the
commits as "submitted for review" rather than waiting to claim approval, all
four got credited automatically the moment the verdict landed, with no
rewriting of history. That mechanism did what it promised.

So the watchdog is finished and I'm treating it as closed. Next is the paid
rehearsal — the real end-to-end run on a rented GPU, a pound or two, and the
one that wants you around when it happens.

Tuesday afternoon. We tried the rehearsal. No model got trained, about ten
pence got spent, and I'd call it one of the more useful afternoons this lane
has had — because everything that stopped us would otherwise have stopped a
paying customer instead, quietly, some weeks from now.

I started by reading the training script we were about to run, rather than
running it. Two things were wrong with it. Back in July we taught it to accept
a choice of model, and recorded that as "ready for small models" — but the
script also had the *conversation format* of the big Llama model hard-written
into it in two places, and nobody had noticed those needed to change too. If
we'd run it as planned, it would have finished happily and handed us a trained
model that had been taught with the wrong punctuation, so to speak. It wouldn't
have crashed. It would just have been subtly rubbish, and we'd have paid for it
and possibly not spotted why. That's fixed, and I've added a check that refuses
to start if the format and the model disagree — and I tested the check by
proving it correctly rejects the old broken combination, because a safety check
you haven't seen fail isn't a safety check.

The second was smaller: our copy of a script had drifted from the live one in a
way that would have made the big training runs save their progress five times
less often. Also fixed.

Then we tried to rent a GPU, and got three surprises in a row. First, the
default number of processor cores we ask for is one the supplier no longer
accepts for nine of the eleven machine types — including every cheap one. Only
the most expensive machine works with our defaults, which is a slightly
comical way to fail. Second, we only wait five minutes for a machine to boot,
and the cheap machine we want for the playground takes longer than that — so
we create it, pay for it, and then our own tidy-up deletes it right before it
becomes useful. Twice.

The third one is why I stopped. Renting a machine was creating *more machines
than we asked for*. Two requests turned into three real, billing GPUs. The
cause is that our system waits on the phone, so to speak, for five minutes
while the machine boots — but the messaging layer behind it gives up after
sixty seconds and hands the same request to someone else, who dutifully rents
another one. Nothing in the code counts how many times this can happen; it's
paced only by how slow the supplier is. On a machine that boots slightly
faster, each duplicate would have been left running rather than cleaned up.

So I hit the emergency stop — provisioning is paused across the whole system —
and then checked that the stop actually works rather than assuming it: two more
duplicate requests arrived in the next minute and a half and were both refused,
nothing was created, and the supplier now shows no machines running at all. You
chose to leave it paused until it's fixed, which I agree with. Nothing else has
rented a GPU since June, so it costs us nothing to leave off.

Two things I got wrong today and have corrected in the notes. I told you the
playground machine's floor price was 43p an hour rather than 35p — that was me
doing arithmetic on an assumption about core charges, not a measured price, and
the truth is somewhere between the two until we see a real invoice. And I should
be clear that the "no machine got left running" claim is from asking the supplier
directly, not from our own records — our own records wouldn't have known, which
is itself one of the findings.

Where that leaves us: the training side is now in better shape than it was this
morning and is ready. The renting side needs a code fix and a redeploy, which is
your release to run, and I've written both bug reports with the fixes ranked. I
also sent the duplicate-machine problem to the diagnosis system for an
independent second opinion on the exact cause, because I'm confident about what
happened and less confident about precisely which bit of the messaging layer
does it, and that's a distinction worth keeping honest.

---

**2026-08-12, evening.** I picked this up to do the fix we'd queued: stop one
request for a machine turning into several machines. It's done and committed,
but the interesting part is that **the explanation we'd written down this
morning was wrong**, and it was wrong in a way that would have cost us.

This morning's account was: the messaging system delivered the same message
several times while the handler was busy, so we built several machines. That is
not what happened. What actually happens is this. When we ask for a machine, the
step that asks is willing to wait **ten minutes** for an answer. If no answer
arrives in ten minutes, the system does not give up — it **asks again**, from
scratch, as a brand new request. It will do that up to four times. Nothing in
the machine-renting code recognised the second, third and fourth asks as being
the same original request, so each one rented another machine. Four asks, four
machines.

The ten-minute gaps we'd measured and read as "the message was redelivered" were
simply that ten-minute patience running out, over and over.

**How I caught it, and why I'm confident this time.** Every request we send is
recorded in a table with its own id. If a message is genuinely delivered twice,
both copies carry the *same* id — that's what redelivery means, the identical
message arriving again. I looked, and the four attempts had **four different
ids**. That rules the old explanation out rather than just failing to support
it. The timings settle it beyond doubt: each attempt was closed off at precisely
the moment its ten minutes expired, and the next one went out about a second
later. That's a stopwatch running down, not a network hiccup.

**Why it nearly cost us.** This morning's report suggested, as the cheap first
move, changing some messaging timeouts — no rebuild needed, quick win. That would
have done **nothing at all**, because it adjusts a mechanism that was never
involved, and we'd have spent a build cycle discovering it while the underlying
problem carried on. I've struck that recommendation out.

I should be straight about my own part in this, because it's the same shape of
error. Reading the code fresh, I decided this morning's suggested fix used the
*wrong identifier* to spot duplicates and that I'd found a flaw in it. My
reasoning was fine; it just rested on this morning's wrong story about what was
happening. When I checked the story instead of acting on my objection, it turned
out the original choice of identifier was right all along — for a reason nobody
had written down. **Had I "corrected" it, we'd have shipped a safety guard that
could never once have triggered, and every test would have passed.** I've written
that trap up where the next person will hit it.

**One more thing we now understand.** When the machine-renting service failed, it
did send back an error, promptly — but that error never stopped the ten-minute
clock. The system waited out the full ten minutes anyway and then asked again. If
that error had been heard, we'd have failed once and built **one** machine, not
four. So that loose end from this morning isn't a footnote; it's half the reason
this happened. I haven't fixed it — it's a different part of the system and it
deserves its own investigation rather than a guess.

**Where that leaves us.** The fix is written, tested and committed, and I've put
it through the review council. It is **not live**: it needs a rebuild and one of
your fleet releases, and a small database change applied first. Renting stays
switched off until then, which is right — a fix that isn't in the running program
hasn't fixed anything. And this fix stops the *duplicate* machines; it does not
yet make the machine we actually want start up in time. That's the other bug
(258), still open, and now safer to fix because duplicates can't pile up while we
do.

---

**2026-08-13.** You said the new chassis probably included the thunder adapter. It
did — I checked rather than assumed, because a fleet release can straddle several
commits and ship different code per service. The adapter is running v1.0.1295,
built from a commit that contains yesterday's duplicate-machine fix, 154 commits
back. So **the fix that stops one request renting four machines is live.**

Before doing anything with that, I checked one more thing: is the *other* problem
still there — the one where we give a machine only five minutes to start up and
then delete it for being slow? It is. I checked it in the exact code that was
deployed, not in my own working copy. Which meant that if we'd switched renting
back on this evening, we would have rented a machine, waited five minutes,
destroyed it, and reported a failure. Money for nothing, on a run we already knew
would fail. That's why I asked you rather than just proceeding, and you chose to
fix that first. I think that was right.

So that's what I did. Two changes:

**The machine size is no longer our guess.** We were asking for 4 CPUs by default,
and the supplier rejects 4 for nine of its eleven single-GPU machines — every
cheap one. The only ones that accepted it were the most expensive box on the menu.
So "rent a machine with our defaults" effectively meant "rent the dearest one, or
fail". That's also why nobody noticed for so long: anyone who tested with the
expensive machine saw it work fine. Now we ask the supplier what it will accept
and take the cheapest valid option. I re-checked their published list myself
rather than trusting yesterday's note about it, since it's their menu and they can
change it.

**The startup deadline is now a setting, not baked into the program.** You can
change how long we wait without me rebuilding anything. Default is nine minutes,
up from five.

**And I found a trap in my own change, which is the part I want to flag.** I asked
what happens if someone sets that waiting time *too high* — expecting the answer
to be "we wait too long". It isn't. Past about ten minutes, the surrounding system
gives up first and asks again; the new duplicate guard correctly refuses the
second request; and the result is that **the job reports failure while a real,
running, billed machine carries on with nobody watching it.** So being more
patient produces a machine we've rented and forgotten. I've set the default safely
below that line, written down the order to change things in if you ever need it
higher, and put the warning in five places including the operator's runbook —
because it's one database command away from anyone trying to be helpful.

**Where we are.** Both fixes are written, tested and committed. Neither is live:
they need a rebuild and one of your releases. Renting stays switched off until
then. Nothing was spent today.

**One snag, and it needs you.** The cluster login token expired at seven past
seven this evening, so for the last few hours I've had no access. Two things are
waiting on that, neither urgent: the small database change that carries the new
setting, and sending the change to the review council. Both are written and ready.
When you refresh the token they'll take a couple of minutes.

---

**2026-08-15.** The new build has both fixes in it — I checked the running service's
own record of what it was built from, and both yesterday's duplicate-machine fix and
the two machine-sizing fixes are in there.

**But one of them wasn't actually working, and it's worth explaining why, because it
nearly slipped past me.** The fix that stops us destroying a machine for being slow
to start needs two things: the new program, and a small database setting that tells
it how long to wait. The program shipped on Wednesday. The database setting didn't —
that had been blocked on the expired login token. So for about forty-four hours we
had a fix that every check said was deployed, and which was quietly still using the
old five-minute limit. Nothing failed; the program is written to carry on sensibly
if the setting is missing, which is right, and is also exactly what made it
invisible. I've applied the setting now (nine minutes), and confirmed the program
can actually read it rather than assuming.

I've written that up as a general warning for everyone: **"it's deployed" answers
whether the code is there, not whether the behaviour is on.** For anything that
depends on a setting, you have to check both.

**Where we are now.** Everything is in place: both fixes live, the setting applied
and sanity-checked against the surrounding timeout, the change sent to the review
council, and all the training material still staged in storage from before. Nothing
has been spent and no machine has been rented since Wednesday.

**What's left is the actual test run** — switch renting back on, ask for one cheap
machine with no size specified (that's the point of the sizing fix), and see it come
up and stay up. That's the money step, about four to ten pence, and it's the first
run that has a real chance of working end to end. It also finally answers something
we still don't know: how long one of these machines actually takes to start. Every
previous attempt was killed at five minutes, so all we can honestly say is "more
than five".

I've left detailed instructions so this can be picked up in a fresh conversation
without losing anything — what to run, what to watch, and the three separate things
that run needs to prove. One caution I've flagged prominently: if the machine starts
quickly, that will *not* prove the duplicate-machine guard works, because that guard
only fires when something is slow. It's live and thoroughly tested, but it has still
never been seen doing its job on a real request, and I'd rather that stayed written
down than got quietly counted as done.

---

## 2026-08-15, later the same day — picked it back up, and checked rather than trusted

Came back to this fresh and went through the handoff's own checklist against the
running system instead of reading it forward. Almost everything held. One thing had
already gone out of date, within about three hours of being written.

**Somebody else rebuilt the fleet while this was sitting paused.** The handoff
recorded which build of the GPU service was running, and by the time I looked it was
a different one — a newer build, shipped by another session working on something
completely unrelated. That is normal here and not a problem in itself, but it does
mean the note saying "our two fixes are in the running service" was, strictly, about
a service that is no longer running. So I re-checked it against the build that is
actually live now. **Both fixes are still in.** Also worth saying: I checked in a way
that could have come out "no" — I included a case that should fail, and it did fail,
which is what makes the "yes" worth anything.

**The review verdict came back APPROVED.** That was the outstanding question from
this morning. Nine reviewers looked at the sizing fix and the timeout fix; eight
approved outright and one raised a records-keeping quibble about a rollback script.
Nothing high-severity, nothing blocking.

What I think is genuinely interesting is *what they agreed on*. **Four separate
reviewers, independently, objected to the same thing** — that the safety margin
protecting us from the duplicate-machine bug is currently held in place by a comment
and a test, not by the machinery itself. In plain terms: there are two timers that
have to stay in a particular order, they live in two different places owned by two
different bits of the system, and nothing stops a future session changing one without
knowing about the other. If that ever happens, the failure is the nasty one — it
looks like a clean failure while a machine we're paying for keeps running unattended.
I checked the actual numbers today and they are correct with the intended margin. The
gap is about durability, not about today.

**Nothing is running and nothing is being billed** — I confirmed that with the vendor
directly, not just our own records. One thing that gave me a moment's fright: our
table of machines has 23 rows in it, where I expected none. They are all long since
shut down, the most recent back in June. It's a history table, not a list of what's
switched on.

I also cleared three small loose ends the reviewers had asked about — whether any
other part of the code still had the old wrong machine-size baked in (it doesn't),
whether we'd accidentally rebuilt something that already existed (we hadn't), and
whether the new setting could break another part of the service that reads the same
configuration (it can't). Those were cheap to check and are now written down rather
than left as open questions.

**So we are exactly one step from the first real test run, and that step costs money**
— a few pence, but it's a live machine at a vendor, so I've stopped and asked rather
than just doing it. That question is with you now.

---

## 2026-08-15, end of day — the bigger card timed, the training run held at the start line, and one correction of my own

You approved three things this afternoon; here is where each landed.

**The bigger card (a100xl) also starts in seconds** — about 12 to 17 of them, and
the sizing logic picked the right processor count for it too, on a completely
different catalogue entry. So the fix works across the menu, not just on one card.
Cost of finding that out: under five pence.

**But I have to take back something I told you this morning.** I said machines start
in sixteen seconds and our nine-minute patience window was absurdly long. Then,
rereading Tuesday's notes for an unrelated recipe, I found this same card was
measured *twice on Tuesday* taking nearly five minutes and still not ready. Same
machine type, three days apart, twenty times slower. So the truth is: **start-up
time swings wildly from day to day**, our patience window is doing its job on the
slow days, and the duplicate-machine guard *could* still get its natural test on a
slow day — I was wrong to say it never would. I've corrected this everywhere I
wrote it down, with the correction visible rather than papered over.

**The training run got to the very start line and is parked there.** Everything
free was done and checked: the training scripts and data verified in storage, the
delivery addresses minted, the launch command written out and double-checked
against the code that will receive it. A machine was actually started — and your
message about credits arrived moments later, so I shut it down before it did any
work (four seconds' worth of charge, about a penny) and put the safety catch back
on. **Nothing is running, nothing is billing, and the whole day's vendor cost is
about three pence real** (our own books say eleven, because they still price every
card at a flat rate — that discrepancy is written up and waiting on a real invoice).

When you're ready to spend again, the next session opens the runbook and fires the
training run in minutes — no research, no rebuilding. The one thing already in
flight when your message arrived was the small automated review of the note I added
to the shared traps file this morning; that one couldn't be recalled, but it's a
single small job, not a machine.

> **Dated correction, 2026-08-15:** the section above was accidentally inserted
> *ahead* of the one below it, and ate its heading — the heading is restored here
> and the mistake is owned rather than smoothed over. Read the section below
> ("we rented a GPU") FIRST: it happened earlier in the day. The pattern-check on
> the commit is what caught it.

---

## 2026-08-15, later — we rented a GPU, it worked, and it cost about six pence

You said do the provision test only. Done. It worked, first try, no drama. Here is
what happened and the two things that surprised me.

**The run.** I unpaused at 14:26:50, asked for one A6000, and deliberately did *not*
tell it how many processors to use — that was the whole point, because the bug we
fixed was that the system used to guess a number the vendor rejects. It worked out
the right answer by itself, asked the vendor for the machine, the machine came up,
and the job finished. I shut the machine down by hand, checked with the vendor that
it was really gone, and put the safety catch back on at 14:30:24. **The whole thing
was unpaused for three and a half minutes.** Nothing is running now; I confirmed that
with the vendor directly rather than trusting our own records.

**Both fixes are proven working on a real, paid machine, not just in tests.** The
sizing fix worked out "6 processors" from the vendor's own published menu, and the
vendor's records confirm it built the machine with exactly that. The timeout fix read
its setting from the database like it was supposed to, instead of falling back to the
old hardcoded five minutes. That five-minute fallback is the thing that quietly wasted
44 hours earlier this week, so seeing it read the real setting is the result that
mattered most.

**Surprise one: these machines start in about sixteen seconds.** We have been carrying
"more than five minutes" as the honest figure for months, because every previous
attempt got killed at the five-minute mark before it finished. It turns out an A6000
is up in roughly a quarter of a minute. I want to be careful here — the older slow
cases were a *different, bigger* type of card, so it may genuinely be slow for those.
But for the card we actually plan to use, our nine-minute patience setting is about
thirty-three times longer than it needs to be.

That has an awkward consequence. **The duplicate-machine guard now has almost no way
of ever proving itself in normal use.** It only wakes up when a machine takes so long
that the system gives up waiting — and if the machine is ready in sixteen seconds,
that will never happen. So the honest position is unchanged from this morning: that
guard is live, well tested, and still has never been seen doing its job for real. The
only way to see it work is to deliberately engineer a slow case, and doing that
carelessly creates precisely the dangerous situation it exists to prevent — a machine
running and billing while the system reports failure and nobody is watching. **I don't
think a session should decide that on its own. It's a decision for you.**

**Surprise two: the cost figure in our own records is not what the vendor charges.**
It says six and a half pence. That number is our own estimate — we multiply the time
used by a single flat rate of $1.80 an hour that we apply to every type of card,
regardless of which one it actually was. The A6000 is advertised at around 35 to 43
cents an hour, so the real charge for that run was probably nearer **one and a half
pence**. We are over-estimating by roughly four or five times.

That is the safe direction to be wrong in — our £30 daily ceiling trips early rather
than late, so it protects us harder than we thought. But it does mean the "spend so
far" number is a worst case, not a bill, and **the question of what an A6000 really
costs us is still open. Only an actual invoice will settle it.**

**One near-miss worth telling you about,** because it is the kind of thing that wastes
an afternoon. Halfway through, I checked the logs to confirm the sizing fix had run,
and the logs came back empty. On the face of it that meant the fix had not run at all,
and the obvious next move would have been to go hunting for a bug in it. Before doing
that I searched for a line I had already read with my own eyes two minutes earlier —
and *that* came back empty too. Which is impossible, so the problem was my search, not
the system. Asking the same question a slightly different way returned the full logs
and both proofs, exactly as they should be. Nothing was wrong. I have written that
trap up in the shared traps file, because an empty search result reads like an answer
and it usually isn't one.

**Where that leaves us.** The provisioning half of Phase 0 is genuinely finished and
proven on real hardware. The remaining half — actually driving a training run on the
machine, and getting the trained result safely into storage — is untouched and still
staged. That is the next natural piece of work whenever you want it, and it will cost
more than six pence because the machine has to stay up long enough to train something.

---

## 2026-08-15, evening — the first complete training run, start to finish, and the proof that was missing since June

Credits came back and you said carry on, so the training run went ahead. **It
worked, completely, first full attempt** — though it stumbled twice at the start in
ways that were cheap to hit now and would have been expensive to hit later, which
is exactly what a first run is for.

The stumbles, briefly: fresh machines don't have the working directory the scripts
assume, and the setup script refused our mid-size card because it still assumed the
huge model from June. Both fixed within minutes, both written into the runbook, and
the second one is committed to the repository. One subtler catch: the launch
command *claimed* success while actually having done nothing — the way it was
chained meant "launched" printed regardless. That's now restructured so the word
only prints when it's true, and it's in the traps file.

Then the run itself: about five and a half minutes installing software, a
two-minute rehearsal on 20 examples to catch format problems before the real
spend, then the real thing — 300 examples, three passes, 23 minutes. The training
data rendered correctly in the model's own chat format, which is the thing the
August 12th fix was for; the training loss halved, which is what you want to see;
and the finished 68 MB adapter was uploaded to our storage automatically.

**Then the step that mattered most: before destroying the machine, I fetched the
uploaded file from storage independently and checked it was really, fully there.
It was.** That closes the question open since June — whether "job finished" can be
trusted to mean "the output is safe off the machine". It can, and now it's proven
rather than believed. This also means the automated babysitter (which watches runs
and cleans up machines when they finish) is safe to switch on — I've left that
switch for you, since it's a standing fleet behaviour rather than a one-off.

The bill for the complete rehearsal of the customer experience: **about 30 cents
of real GPU time, 50 minutes end to end.** Our own books say $1.50 because they
still price every card at the flat $1.80 rate — same known discrepancy as this
morning, waiting on a real invoice.

Still open: re-uploading the corrected script bundle to storage (my upload was
blocked by a session permission — either allow it or run the §2 recipe yourself),
converting the adapter to the playground's format, and timing a playground
session. Then Phase 0 is done and we can price the service from measured numbers.

---

## 2026-08-17 — Phase 0 is finished

Two things remained: converting the trained model into the format the playground
serves, and rehearsing the playground itself. Both are done, and the rehearsal
answered the question the whole booking design was waiting on.

The conversion failed on its first try yesterday in an annoying way — the
conversion tool reported success but had written the file into a different
folder than the one it was told, so our safety check (correctly) refused to
believe it, and the file was lost when the machine shut down. While I was away,
the automatic janitor noticed the idle machine and shut it down at its two-hour
limit — the first time it has ever done that for real, and it worked exactly as
designed, which is quietly one of the best results of the week. Today's re-run
looked for the file everywhere rather than trusting the tool's word: eight
minutes, converted, uploaded, and independently verified in storage — a 1 GB
file ready for the playground.

Then the rehearsal: I started a fresh machine using the vendor's
ollama-flavoured template, pulled the converted model from storage, and timed
every step. **From pressing go to the model answering its first question: about
three and a half minutes.** After that first answer it responds in about a third
of a second at ~140 words-worth a second — a perfectly pleasant chat experience.
The practical rule this gives us: start the machine about ten minutes before a
customer's booked hour (that covers the vendor's slow days too) and they will
never see a cold model.

The whole of Phase 0 — every rehearsal, every failure, every measurement across
both days — cost about **$1.12 of real GPU money**. Everything is switched off
and the safety catch is back on. What's left is business rather than
engineering: your call on the price (the first invoice will settle what the
cards really cost us), whether to switch the automatic run-babysitter on, and
coordinating the sales page with the thread that owns the site's front end.

---

## 18 August 2026 — price direction, positioning direction, and the machinery armed

Your call on price, recorded: we don't fix it finally until we've walked real
customers through the whole journey. Start at **£99** and see how we go — it's a
business audience, so the credible range is low hundreds; too cheap reads as low
value, too dear kills take-up; reducing later is easy, raising later is hard.
And we watch what competing service companies charge.

Your positioning sketch, recorded as the working hypothesis: **we do a techie
thing without sounding techie** — genuinely helpful in tone (even though it's
automated underneath), and visually just enough "tech authority" to be credible
without looking like the tech companies, which would scare off exactly the
businesses we want. You've asked for more thinking on this together; it's an
open strand, not a settled decision.

Meanwhile three mechanical things moved. The run-babysitter (the monitor that
watches training jobs and cleans up machines when they genuinely finish) is now
switched ON — the proof it needed landed on Friday, and the provisioning safety
catch stays on regardless. The mystery from the front-end thread — repairs that
report "done" while changing nothing — has been formally handed to the diagnosis
loop, which reads the real code and evidence and comes back with a cited verdict.
And your invoice confirmed our cost estimates to the cent, so the £99 starting
price sits on measured ground: roughly £1 of GPU per customer job.

---

## 18 August 2026, later — position like the expensive agencies, price like an invitation

Your rulings today, recorded: the positioning stands — we present like the
serious, expensive agencies but keep the £99 price; the cheapness must never
read as a disadvantage, and the way we earn that is by **describing the process
in detail** so the buyer can see we are what we say we are. If the messaging
can't carry it we could raise the price later, but we try the price we like
first. And my sample copy was rightly rejected — too obviously machine-written.
The page's words will come through the framework, not from a session's keyboard.

Two consultations are now formally lodged with the specialist threads, exactly
as you asked: the **copy quality** thread has been asked how to encode the
friendly-expansive, glossary-backed register into the site's own specification
(so the tone is a property of the site, not of one draft), and the **offer
analysis** thread has been asked to rank the benefits — which one leads, which
true claim would be the wrong opener, and whether process-transparency really
answers "£99? what's the catch". Their trades, their call; our facts are
attached. The page won't be seeded until both answer.

---

2026-08-24. The offer page is on its way. Three things happened today, in order.

First, the homework other threads left us paid off. The copy team's answer told us where
wording guidance actually lands (worked examples beat rules, and the writer reads one
particular formatted field), and a sister site (apis.uk) ran the experiment we would
otherwise have run blind: example sentences get COPIED ONTO THE PAGE as content unless you
add an explicit "these are style samples, not content" note — and with that note plus a
named subject for every section, the AI-sounding constructions the owner dislikes went
from 12 to 0 on their served page. We took all of that as given rather than rediscovering it.

Second, the seeding itself. The site's brief was quietly working against us: four of its
five "write like this" examples were themselves built on the "X, not Y" pattern we're trying
to stamp out, and the site's own quality checker had flagged four such phrases on 08-20.
All cleared today: the examples are rewritten positive-first with the style-samples note
attached, the key differentiators are rewritten as gains with the agreed lead ("your
company's voice, in a model you own") in first position, two dead spec fields nothing reads
are retired, and — a thing that would have bitten later — the site's claims register said
"no numbers may be stated on this site", which would have stopped the offer page naming its
own price. The £99 (owner's decision) and the ~$5,000 market anchor (from the research) are
now registered, so the page may say them and nothing else numeric.

Third, the page. We found the old "don't add new pages" warning had expired (the bug behind
it was fixed and closed), so the offer page went in through the framework's normal route: a
planned page plus a work order for the builder, with a brief that gives each of the six
sections one named subject — the offer, how it works (send examples, trained overnight,
chat with it next day), exactly what you get, why £99 can be enough, a small glossary, and
a gentle book-a-call ending. The brief explicitly forbids the "a real person checks every
run" line you flagged; the safe form "run by people, not left to a queue" is in. The
builder picks work up on a one-minute tick; we're watching it and will check the finished
copy for the old tells before anything else happens.

Still waiting on: the benefit-ordering thread (no reply yet — the question has shrunk to
"what is the first differentiator, written as a gain", and we've made that call for now),
and your three Phase 1 decisions (playground booking shape, sample datasets, Stripe posture).

Later the same day. The builder wrote the page — and a quality gate then blocked it, wrongly.
The gate keeps a list of "template leftovers" to catch (things like "[Your Company]" that a
lazy template forgets to fill in), and one entry on that list was simply the bare words "your
company" — which is also the second word-pair of our agreed headline, and of half the privacy
copy on this site. We pulled the records: that entry has fired 46 times in its life, every
single one a false alarm, 41 of them on this site — it has been quietly killing our page
builds for three weeks and nobody saw, because blocked builds land in a review queue with no
screen. We fixed the list (the genuinely-suspicious forms are still caught), the review
council approved it first time, and the fix ships with the next release; the page then
rebuilds on its own.

The good news is the writing itself. We checked the blocked page's copy word by word: the
constructions you flagged as AI-sounding are at zero, none of our example sentences were
copied onto the page, the only number on the page is £99, and the promise we agreed not to
make isn't there. One habit survived — sentences shaped like "X rather than Y" — and we could
trace it directly: the instructions we hand the writer USE that shape twenty-one times while
telling it not to. We've cleaned up the instructions we own; the shared fleet wording is the
copy team's call and we've sent them the numbers.

Evening, same day. Three good things and one decision for you.

The fix shipped with the evening release, the page rebuilt itself, and it's live:
https://finetuning.uk/your-own-model.html. The headline is the line we agreed, the price is
stated once and plainly, the promise we retired stays absent — instead the page says a person
runs each training while volumes are small "and we'll say plainly if that changes", which is
honest in exactly the way we wanted. The glossary answers "what is GGUF" in one sentence. The
page isn't yet in the site menu — that rebuild is queued and should follow on its own.

We also ran the copy editor over it once, deliberately. It found something real: two sections
were re-telling the three-step story the "How it works" section had already told, and it
proposes rewriting them — one into "what you get and what you can do with it" (adding a
contact link in the body text), one into a practical "what to send us" list. Reading it, the
rewrite is good. The automated checker flags it — partly for reasons that look like checker
sharpness rather than real faults, and partly one genuine call: the rewrite swaps two
subheadings for a bullet list, which the checker counts as structure loss.

**The decision for you:** approve or decline that rewrite (work item 8003c51a, parked for
review as designed — nothing applies without you). If you'd rather read it on the page than
in a database row, say so and we'll stage it somewhere readable.

Still on the list, in your order: the playground booking shape, sample datasets, the terms
questions — and Stripe last, as you said.

One more thing prepared tonight so it's ready when you are: the terms questions. Extending the
site's terms and privacy pages for customer training data needs four decisions that are yours,
because each one is a commitment we then have to keep:

1. **Retention** — after a customer's model is trained and handed over, how long do we keep
   their training documents, and the trained file? (Options range from "deleted once the
   playground hour is done" to "kept while they remain a customer". Whatever you pick, the
   terms will say it and the page copy can then say it too.)
2. **Deletion on request** — do we promise deletion on request at any time, and how fast?
3. **Where the data lives during training** — the honest sentence today is: it goes to a rented
   GPU machine for the training run and to our storage for the handover. Are you happy for the
   terms to name that plainly?
4. **Playground hours** — are they part of the product in the terms (one booked hour included,
   more purchasable), and do unused hours expire?

The licence side needed no decision: we verified the three model families' licences in writing
today (Llama's community licence, Mistral's Apache 2.0, Phi's MIT) and registered them, so the
new technical page can state them exactly. Nothing in the terms work blocks the site as it
stands — the £99 page is live without it — but the first paying customer makes it due.

2026-08-25. Your verdict on the two new pages, recorded: the copy fails the "would a person
actually say this" test — "very AI sounding", "so methodical like AI" — with three quoted
specimens (the licence-summary paragraph on the technical page, the "comes down to three
steps" line, and the whole "Who is actually running this" section). The rest you rated "not
so bad to be fair"; facts and claims are fine. You also flagged that the front page cards are
all negatively framed, and said the whole site could be rewritten in better language.

> **CORRECTION (mine, 2026-08-25):** my note of yesterday evening praised the "Who is
> actually running this" section as "honest in exactly the way we wanted". You've overruled
> that, and the miss is instructive: that section passed every automated tell-check we ran
> (zero em dashes, zero "not just", zero of the owner-named constructions) and still reads as
> AI. The checklist is not your ear. I've said this to the copy team in the escalation, as
> evidence their acceptance test needs to change shape.

What was done with your verdict: it went, verbatim, to the copy quality lane as you asked —
both pages were framework-written end to end, so this is their machinery's ceiling, and the
escalation says plainly that it "will need to substantially improve", with the three
specimens, the front-page card measurements (four of six differentiator card HEADLINES are
built on "X, not Y" — those render from the old specs and would fix mechanically on a
rebuild), and everything they need to reproduce it.

What was deliberately NOT done: no rebuilds fired. A front-page rebuild today would fix the
card headlines (yesterday's re-seeding) but the prose would come back in the same register —
we measured that ceiling across three builds yesterday. Rewriting the site before the
machinery improves would spend money to reproduce the problem at scale. The two new pages
stay up (facts right, price right, links right) unless you'd rather they came down.

2026-08-25, later. Three things happened since this morning, and none of them needs you to do
anything today except the decisions already on your list.

The copy team has answered. Your verdict on our two pages is now the first item on their work
list. It turns out a second complaint from you landed with them the same day — the one about
the garden site — and it came with instructions: raise their game a lot, go and re-read the
whole accumulated discussion before proposing any fix, and audit every single writing
instruction we hold, in the database and in the code, for whether it is teaching an AI style of
writing. They have done the re-reading and they have scoped the audit. What they have NOT yet
done is change anything that would make a page read differently. So we are still holding: no
rebuilds here, no rewrite, because a rewrite today would cost money and come back in the same
voice you rejected.

The rewrite you were asked to approve or decline: I now recommend not approving it as it
stands, and here is the honest reason. The copy team ran it through their own checker and it
turned out to be worse than I described last night. I re-ran that check myself rather than take
their word for it. The rewrite does not just swap two subheadings for a bullet list — one of
the two edits deletes an entire numbered list from the page, three items, and the other removes
both its headings and half its paragraphs. Each edit cuts about a third to a half of the words.
The prose may well read better; that is not the point. We have been bitten before by an editing
step that quietly threw away part of a page and reported success. So the options are: leave it
parked until the copy machinery improves and fold it into the proper rewrite (my
recommendation), or ask for the same edits again with the list and the headings kept, or drop
it. Applying it as it stands would take a list off a live page.

One tidy-up: the validator bug that blocked this site's builds for three weeks is now formally
closed. I checked it the hard way rather than trusting the release — the live page is serving
the exact sentence that used to be rejected ("Your company's voice, in a model you own"), and
it was published after the fix went out, so it cannot be running the old rule.

Everything else waits on you, in your order: the rewrite decision above, the header slot, the
booking shape, sample datasets, the four terms questions, and Stripe last.

2026-08-25, same evening. Your answer on the rewrite, recorded: hold it, and fold it into the
proper rewrite when the copy machinery is better. So it stays parked exactly where it is —
nothing applied, nothing thrown away, and the page is untouched. Next move on it is not ours
until the copy team ships something that changes how the writing reads.

2026-08-25, later still. All seven of your answers are in and acted on, and the eighth thing you
mentioned turned out to be the most interesting.

**The unreadable pages.** You said a couple of pages have no hero image and that this had made the
copy unreadable. That is exactly right, and the chain is worth writing down because it is not
obvious. When a page HAS a hero image, the banner is that image with a dark tint over it, and white
text on it reads fine. When a page has NO hero image, the banner falls back to being painted a
solid colour instead — and on this site that fallback is broken. The colour it reaches for is not a
plain colour at all, it is a two-tone blend, and the place the code puts it only accepts a plain
colour. So the browser throws the whole instruction away and paints nothing. The heading stayed
white, the page behind it is cream, and white on cream is not readable. So: no image, therefore
unreadable. Your sentence was the diagnosis.

Two things came out of that. First, it is not our site's fault and not only our site: seven other
sites use the same two-tone colour, and I found the same fault live on robot-hands.com and
gaswholesalers.com. Second, there was a worse version of it that you did not mention because you
could not see it — some of the buttons are white text on a white button, so the label is
completely invisible. That one is on the £99 page too.

The banner half is fixed and live on all three sites. I checked it by loading the actual pages, not
by trusting a "done" message — which mattered, because the first three re-renders reported success
and had changed nothing at all. The invisible buttons need a software release before they can be
fixed; the fix is written and waiting for it.

**The header.** You said I could move About, Case Studies, How We Work or Contact. I moved Contact —
it was the only one that costs nothing, because the "Get Started" button in the same header already
goes to the contact page. That turned out not to be enough, and the reason is worth knowing: the
header is not ordered by anything we set per site. The system has a fixed list of page names it
considers important — home, services, about, tools, pricing, case studies and so on — and it fills
eight slots from that list. "Your Own Model" is not a name it knows, so freeing a slot just let
Pricing move up instead. To actually get the offer page in, a second page from your list had to
come out, so I moved How We Work as well — it overlaps with Services and with the Approach page in
the footer. Both moved pages are still reachable from the footer. If you would rather have How We
Work back and lose Case Studies instead, that is one command.

**Your terms answers are registered** as facts the writing system can quote: bookable 9 to 5 UK on
weekdays with other times by arrangement, deletion within a week of a request, documents and model
kept 30 days after handover by default, and one playground hour included that expires after 30
days. One of the four is still open — whether the terms may say plainly that documents go to a
rented GPU machine for the training run and to our storage for the handover. I did not want to put
words in your mouth on that one.

**The sample datasets.** I have designed six, keyed to jobs rather than industries: email voice,
copy structure, copy style, support replies, product descriptions, and summarising long documents.
Two of those deliberately have nothing to do with voice, so a prospect whose problem is volume or
structure can see themselves. Each one gets real example data, a held-back set, and a worked
example that shows what the plain model produced FIRST — and at least one case where fine-tuning
did not help. A page of only wins is the tone you have twice told us to stop using.

I have not generated any of it yet, because there is one question only you can answer: whose
material do we use? My recommendation is our own writing for four of the six — our real emails,
replies and copy — because a model trained on our own words demonstrating our own voice is the
actual claim rather than a simulation of it, and it needs nobody's permission but yours. The other
two would be invented. The alternatives are open-licensed text, which reads as generic, or a
customer's with written permission, which cannot be first.

**The copy machinery has not moved yet**, so the rewrite you told us to wait for is still waiting,
and the site is still on hold for anything that would spend money reproducing the tone you rejected.

Stripe is still last, and yours.

2026-08-26. Your three answers are done.

**The terms** can now be written in full. All four commitments are registered as facts the writing
system can quote, including the one you have just settled — that the terms may say plainly that
documents go to a rented GPU machine for the training run and to our storage for the handover.

**The nav problem is filed** as a bug, with your suggestion as the recommended fix. The evidence
turned out to be better than I expected: somebody has already hit this once and their workaround is
still in the code — three page names hardcoded into a fleet-wide list to force one site's pages into
its header — and I measured today that those very pages are missing again, because that site is
back at the eight-slot limit. So the existing approach demonstrably does not hold. Eight pages
across five sites currently say "put me in the header" and are not there.

**The datasets.** I have built the machinery that checks a dataset before it costs anything, and the
first dataset through it — the one that teaches structure rather than voice — is done: ninety
examples, eighty for training and ten held back. The checker exists because our first training run
quietly lost five of its three hundred rows to a length limit, and we only found out after the GPU
was paid for.

But building it turned up a problem with what you approved, and I would rather raise it than work
around it. "Our own material" means our own published copy — and our own published copy is exactly
the writing you rejected twice as sounding like AI. For the three datasets whose whole point is to
demonstrate a voice, training on it would teach a model the register you have told us to stop using,
and then show that to prospects as a selling point.

The dataset about summarising documents is unaffected, because its target is a format rather than a
voice, and so is the second invented one. So four of the six can proceed.

For the other three, the only material we hold that is written in a voice you actually approve is
**your own** — the notes in this file read plainly and directly, which is what the copy team is
trying to reach. I am not going to use your personal writing for a customer-facing demo without you
saying so. The alternative is to wait until the site is rewritten and harvest it then.

2026-08-26, later. The images are not done, and nothing was changed — which is the right outcome
given what turned up.

First, some good news: most of what the site had recorded as missing images isn't missing. There
were eleven such findings on file since July. I checked them against the live site rather than the
database, and every picture loads, nothing is broken, and no image tag is empty. The check that
raised them compares a filename against an internal list and never actually asks the web — so those
eleven have been wrong all along.

The real problem is narrower: nine pages have no hero image at all. Those are the same nine that
were falling back to the plain coloured banner, which is why the unreadable-heading fault showed up
on them. Worth knowing that every other page on the site uses the same single picture, so simply
filling the gaps would have given you thirty-five pages sharing one image. Each of the nine had its
own concept written instead.

Then the thing that stopped it. To give a page an image, the system rebuilds that page from scratch
through the writer — which rewrites the words. There is no way to add a picture without that. Worse,
the internal job it creates is labelled as if it were the harmless kind; someone has been caught by
that before, and their rebuild invented a contact email address that didn't exist.

You reversed before anything had run, and I have confirmed nothing did: the nine jobs never started,
the rebuild step never fired once, and the pages are byte-for-byte what they were. I had also taken
a full copy of all nine pages beforehand, which is still there.

The nine image briefs are written and banked. When you have the copy machinery sorted, one rebuild
does the new words and the new images together — which is a better result than doing it twice, and
it costs nothing to wait.

I have filed the underlying problem as a bug: you should be able to add a picture to a page without
having its words rewritten. The fix looks straightforward — the lighter rebuild already understands
"an image just landed", it simply never gets told.

2026-08-30. Copy quality goes back to the copy team, as you asked, and this thread returns to the
fine-tuning service itself. A fresh handoff is written so a new conversation can pick up cleanly.

Worth knowing before anything else: the cluster login has expired again — that's the three-day
timeout, and it's yours to refresh. Nothing on the database side can be done until it is.

One correction to something I told you on Tuesday, and it's in your favour. I said the nine images
had all failed to appear. That was true when I looked and it isn't true now. Careers has its image.
And when I checked the storage today, all nine images are there, published and loading properly —
eight of them are simply not yet pointed at by their pages. So the pictures exist and are paid for;
what's missing is one line of plumbing per page. That's now the cheapest thing on the list rather
than the most expensive, and it doesn't touch any copy.

What the copy experiment settled, so nobody repeats it: with the bad examples stripped out of the
brief and your rule sitting in it, the writer still produced the comparison habit about five times a
page. Two independent counts agreed. So it isn't imitation and it isn't a missing instruction — the
model simply prefers that shape, and the only thing that will stop it is cutting it mechanically
after the writing. Your rule was right; every single instance shortened with nothing lost. The copy
team have that and the decision that follows from it.

On the service itself: all four of your terms commitments are registered as facts the system can
quote, but the terms and privacy pages themselves still need writing — that's the next real piece of
work and it's unblocked. The six sample datasets are built. The booking shape is decided but not
built. Stripe stays last and yours.

---

**2026-09-02 — the playground page is up, and it taught us something about the site.**

The booking page is live at finetuning.uk/playground.html. That was the last of your seven
decisions still unbuilt, so apart from Stripe — which is yours and last — they are all done now.
The page explains the included hour, when it can be booked (your 9 to 5, weekdays, other times by
arrangement), and what happens to documents afterwards. It only states things we have written
down as facts, so there is nothing on it we cannot stand behind.

It did not go smoothly, and two parts of that are worth telling you.

**The first attempt built nothing at all, and looked like it had worked.** The system reported the
job complete, no errors anywhere, and the page simply did not exist. That was my mistake: I put
the page's structure in the wrong place — in the instruction rather than on the page record. Once
it was in the right place it built properly. I have written the trap down, because the failure is
invisible: nothing goes red, so there is nothing to go looking for.

**The second thing is a genuine fault in the system, not a mistake of mine, and you will see it on
the page.** The middle of the page says the same thing three times. Three sections, three headings,
all more or less "what you do in the hour".

I found out why, and it is not the writing being careless. When a page has several sections of the
same kind, something has to tell each one what it is supposed to be about, so the second does not
just repeat the first. That mechanism exists — but it only works for sites set up in a particular
way, and this site is not one of them. It never has been. So the three sections were each handed
the same instruction and each wrote the most obvious answer to it. Given how the site is built,
that was the only thing that could have happened.

I have not patched over it by rewriting the three sections myself. Two reasons. It would mean
hand-writing content on a site whose whole point is that the framework writes it — the thing you
ruled against in August. And this page is currently the clearest demonstration we have of a fault
that affects six of our sites and 186 pages between them, so quietly tidying it away would remove
the evidence. It is filed as bug 443 with the options for fixing it properly.

The same underlying problem turned out to be blocking another team as well. They are building a
way to get images onto pages automatically, and their approach depends on the same missing piece
this site lacks. Their test site has it, so their tests would have passed while six real sites
got nothing. I have told them, with the numbers.

**Where that leaves you.** Nothing needs a decision from you today. The page is live and honest;
it repeats itself in the middle, and that is logged as a fault to fix rather than something to
live with. If you would rather it not sit there repeating while the fix is worked out, say so and
I will take the page down or cut it to fewer sections — both are reversible, and both are your
call rather than mine.

**2026-09-02 (later) — it turns out the new page was not the problem. Two older ones are worse.**

After I wrote the note above, another session picked up the bug and asked me whether I had any
context I had not written down. That prompted me to check something I should have checked before
filing it: whether the other pages built the same way have the same fault.

They do, and they have it worse.

Your main offer page — the £99 one, your-own-model — has three sections in the middle whose
headings are "How it works", "How it works", and "How it works". Identical, three times. The
technical page does the same with "The model and its licence". The new playground page, the one I
told you about earlier, is actually the mildest of the three, because at least its three headings
are worded differently from each other.

The offer page has been like that since 27 August. Six days, on the page the whole service points
at. I did not spot it, and I had been on that page repeatedly.

I want to be careful about one thing, because it would be easy and wrong to join it up. When you
reviewed the site in August and said it sounded like AI, that was on 25 August, and this copy was
written on the 27th. So this repetition is not what you were reacting to. They are two separate
problems that happen to sit on the same page.

But it does mean something for the copy work. If someone rewrites that page to sound more human,
it will still say "How it works" three times, because that is not a writing problem — the system
asks for the same section three times and has no way to tell the three apart. I have told the copy
team so they do not spend effort on a page where the most obvious flaw is out of their reach.

The good news is that the fault is narrower than I feared. Across everything we run, eleven pages
have the shape that triggers it — four of yours, four on gaswholesalers, three on
ai-agent-orchestration. Three other sites have none at all. So this is a real bug worth fixing
properly, not a fire.

Someone else has taken the fix on and is doing it framework-wide rather than patching our site,
which is the right way round. Nothing needed from you. I would only say that the offer page now
has a visible flaw on it that I did not know about this morning, and if you would rather it did
not sit there while the fix goes through, that is worth telling me — the same choice as the
playground page, and the same answer applies: I can shorten it, but I should not hand-write it.

**2026-09-02 (end of evening) — two decisions taken, one pick left over.**

You decided the planning question: the six older sites move onto the newer planning tables, and
because this is a one-off tidy-up of known sites — no new site will ever take the old route — a
direct database insert is acceptable rather than building machinery for it. That is recorded in
the proposal document itself. One duty survives your decision: before anyone touches a live site,
we prove on a single small site that creating its plan does not set off a wave of page rebuilds.
The design says it should not; nobody has watched it not do so yet.

And you redirected the prompt work, usefully. Rather than approving the block of instructions as
written, you asked for it to be turned around: no negatives (your elephant rule), and written in
the voice we want back rather than as commands about voice. I drafted three versions that differ
only in the scene they set for the writer — answering a person who just asked, writing for a
reader partway down the page, or simply naming the subject and noting the rest of the page covers
everything else. My recommendation is the middle one. The three are written up side by side for
you to pick from next time, filled in with real values so you are reading what the machine would
read, not a form with blanks.

Nothing else moved tonight. The playground page is live, the repeated-headings fix waits on the
prompt you are now reshaping, and the rebuild of your four affected pages waits behind that.

2026-09-02, late. You read the three prompt framings and said they all sounded a bit AI. Fair. Each
one set up a little scene and then padded it. I redrafted against the style prompt you built for the
pitch deck, three plainer versions, and you picked C: the one that speaks the way the page would,
"You'll want to know what to bring to the session. That's what this section is for." followed by a
short list of what the other sections cover.

Before sending your words to the apis.uk lane, who own that file, I rendered the block the way the
machine will, against real data from today's runs. That found something that would have bitten us:
the list of other sections is not visible to the prompt as things are configured today. It would
have rendered the line "also covers, each in its own section:" with nothing under it, and no error
anywhere. The same shape of failure as the bug we are fixing. The fix is one extra name in the
writer's input list, shipped in the same change, and I have told apis.uk it must be checked, not
assumed.

One thing for you to decide, not urgent. Your sentence reads perfectly when the subject is written
the way we write ours ("what to bring", "how the hour works"). On the sites where the planner
writes the subjects, it currently writes things like "Brief description of the sister-site
relationship with gamesdesign.co.uk", and "You'll want to know Brief description of…" reads badly.
Either the planner is nudged to write subjects as the thing a reader wants to know, or the sentence
is loosened. I'd nudge the planner. Your words stay exactly as you picked them either way, and you
will get the exact final text back for a fresh read once apis.uk have put it into the migration.

2026-09-03, morning. You told me a chassis build was on its way within the hour. Checking what
that changes, I found last night's roll had already shipped the piece we were waiting for: the
running service can now carry a subject per section on sites like this one. So the wait is over,
and it was over before the handoff said it would be. I have corrected that note.

I have written the subjects for three of the four affected pages, in the phrasing your chosen
prompt needs, and held off the rebuild until your new build has landed and settled, because a
rebuild running through a roll gets killed. After that, the front-door page gets rebuilt first.
Its three identical "How it works" headings will still be identical after this step; that is
expected, because the prompt change itself is still with the apis.uk lane. What this step proves
is that the subjects reach the writer at all.

And the prompt change is written up as a starting document for the new thread you want to open on
all the prompts in the framework. It has your words, the four drafts and why three were rejected,
how a prompt actually reaches the model here, and a count of every live prompt as of this morning:
141 of them, and only 7 read the one shared voice block.

2026-09-03, mid-morning. You read the three offer pages and the homepage and gave me four things.

The missing link from the offer page to the playground is fixed at the source, so the next
rebuild of that page carries it. The homepage verdict has gone to the copy quality lane in your
words, with your ruling on the "not tied to one provider" line: keep the first half, drop the
rest. The technical-details brief asked for exactly the three-model listing you found unhelpful,
so that brief needs rewriting before the page is rebuilt with the new prompt.

The tool is the big one. The plan from July designed the playground as a live chat with your
model, served from a GPU box for a booked hour, and in August we proved that works by hand:
about three and a half minutes from asking for the box to the first reply. What went live last
night was only the booking page. Making the chat real needs three pieces: the chat box on the
page (we have one in the library to copy), a route in the tools service to carry the messages,
and the model server behind it. The one choice that is yours: should a visitor be able to try a
demo model on the public page without booking, or does the chat only work inside a booked hour?
I'd do both. The demo can run on the small server we already have in the cluster.

Separately, the prompt change is written, reviewed and rehearsed by the apis.uk lane, and is
waiting only on you reading the final words. They are in my message.

2026-09-03, 10:15. You said rebuild the homepage now, and keep "We're not tied to one provider"
with the rest of that sentence cut. Both are done as instructions: the homepage is queued for a
full rewrite through the writer, which is the only route that runs the negation check. One slip
on my side, caught before it mattered: the homepage never had a proper brief on file, only audit
notes, and my first attempt copied one of those. It was rewritten as a plain brief before the
queue picked it up. The technical-details rebuild is still queued too, behind another site's
backlog. I'll report both when they land. The copy lane has the exact "before" for
technical-details pinned, so the "after" can be measured rather than eyeballed.

2026-09-03, 10:50. Both rebuilds have landed. The homepage now says "We're not tied to one
provider." and stops there, which is what you asked for. Of the nine negation shapes the check
found while writing, six were repaired and three were left because the repair would have gutted
the sentence or the model had no rewrite to offer, and it is built to leave a sentence alone
rather than damage it. The one shape you named that the check does not yet look for is the
"so" clause, and the page still has plenty of those; the copy lane is treating that as a
judgement call rather than a pattern, and it is the next thing to decide with them.

The technical-details page came back with every section knowing its own subject, which proves
the fix that lets sites like this one carry subjects at all. Its headings still repeat because
the prompt change is waiting on your read. It also came back with one misspelt closing tag,
which I've filed rather than patched, and a third shorter than before.

2026-09-03, 11:05. You ruled on the two "where is the line" questions. Cuts like the one the
check refused are to be accepted, and the "so" clause is to be repaired only when it is bolted
on to a sentence that has just explained a term. Both have gone to the copy quality lane, who
own the check. One thing they found after I'd asked: the check's idea of "gutted" is a length
rule, anything under forty percent of the original is refused, and your own sentence is at
thirty. So the homepage reads right today because the brief told the writer, and the check
itself would have refused that cut. Changing that number is a code change through review, and
they'll do it on your ruling rather than on a guess.

2026-09-03, 11:15. You chose both: a public demo anyone can try, and booked hours on a GPU with
the customer's own model. Your question was what the demo costs in model fees, and the answer
is nothing: it runs our own small fine-tuned model on a machine we already pay for in the
cluster, with no call to Anthropic or anyone else. The only thing that costs real money by the
hour is the GPU box, and that is reserved for booked sessions at about thirty-five cents an
hour, billed by the minute. The plan for building the playground is now written, in five steps,
starting with getting the demo model loaded and timed.

2026-09-03, 11:45. Your direction for the site, in your words: "As a whole, I'd like the
finetuning site to be very much focused around this tool. We can still have the other tools, but
much of the "what else we do as a company" should now move to leopardess consulting or other "me"
sites. For finetuning.uk I'd like this tool shown prominently on the home page and I want in the
future example after real example of what we've done and before and after examples. And I'd like
to host those same models so they can try them (at maybe a couple of pounds for an hour or
something that covers our costs say 5x) We can talk details later."

Recorded as the direction in the plan and as a new milestone summary. Nothing has been moved or
rebuilt on the back of it yet; the playground build carries on as step one, and the site
refocus, the examples catalogue and the pricing wait for the details conversation. On the
pricing: five times our measured GPU cost comes to about £1.40 an hour, so a couple of pounds
covers it with room.

2026-09-03, 11:55. Two more things from you: "the gpu for the examples could be a big one so it
feels snappy - this might change our pricing estimate. Eventually we may have third parties
submitting their models and they might have a page of their own and we'd show examples or their
results similar to how we'd show our own." Both recorded in the plan. On the money: the small GPU
we tested costs 35 cents an hour and already answered in a third of a second with a small model;
the big one is $1.09 an hour on the same invoice, so five times that is about £4 an hour rather
than £1.40. The big card buys bigger models and more people at once, not speed on a small model.
The third-party idea means the examples pages are built as "model pages" from the start, with an
owner on each, so ours are just the first ones.

2026-09-03, 12:40. A correction to what I told you about the demo's cost. I said it would run on a
machine already in the cluster. The service that has to call it, the tools service behind
tools.apis.uk, does not run in the cluster; it runs on a small rented box that, by design, never
calls into the cluster. That box has one processor and one gigabyte of memory, so it cannot run
the model itself. So the demo needs a small machine of its own that the tools box can reach:
still no per-answer fees, but roughly ten to twenty pounds a month for the box. That is the one
placement choice for you, and it is the only thing between the route, which is written, tested
and with the council, and a working demo. The model I loaded this morning still gave us the
speed measurement, so that work stands.

2026-09-03, 14:10. The chat route passed review, on the fourth attempt, with every seat approving.
The three rounds it took were worth it: they caught a real trap in how migrations are filed (one of
the files I was copying had already been applied to the wrong database back in August), pushed the
duplicated security wiring for the two chat tools into one place, and made me check a claim I had
given you as fact rather than as a guess. That claim was that the tools box cannot reach the model
server we already run in the cluster. It turns out to be true, but I had asserted it before testing
it, and it is the reason the demo needs a machine of its own. So the placement question stands, and
it is now the only thing between a reviewed, tested chat route and a working demo on the page.

**2026-09-03, early evening (a fresh session picking up the handoff).** I checked where the playground
actually stands rather than trusting the handoff, and the answer is: nothing has moved on your side
yet, which is fine, it only went to you an hour ago. The island still runs the old tools-api image,
the five `PLAYGROUND_*` settings are not in its env file, and a test call to the new chat route from
outside gets "not found" (I checked with a route that does exist, which correctly says "forbidden"
for a wrong origin, so the machine and its security gate are fine; the route simply is not in the
running program). Migration 641, the prompt change that stops the repeated headings, is also still
unapplied by the prompts lane.

Two things I could do without you, so I did them.

First, the new tools-api image is built and checked: it is tagged `v1.0.1359-playground`, built from
the committed code, and I extracted the program from it and confirmed the playground handler is in
there (with a positive and a negative control, per the runbook). I tried to copy it onto the island
so it would be sitting there waiting; the session's safety classifier refused, as it refused the env
edit earlier today. So the swap is three commands for you, written out in the runbook under
"Playground tool, step 2". The reason it is yours and not mine: restarting tools-api also restarts
the robot-hands.com and vonc.com tools.

Second, the technical-details page. You said it was "an unhelpful page listing on 3 types of
model", and the reason is that the brief asked for exactly that. I have rewritten the brief: the
page now says we choose a small open-weight model to suit the job and tell you in writing which one
and which licence before training starts, that we only use models whose licence lets a business
your size use the result commercially at no charge, and that the exact terms come with the
handover. No list of families, no licence small print on the page. It also now points at the
playground. The rebuild is written as a file that REFUSES to run until 641 is live, so nobody can
fire it early by accident; when the prompts lane applies 641, this lane runs it.

One thing I need your call on, and it is not urgent until the route is live. The plan was to take
the library's existing chat box and point it at the new route. I read the library chat box: it is
a one-message-at-a-time box that talks to a page's own server, not a conversation box that talks
to tools.apis.uk and shows the reply as it streams. Pointing it elsewhere is not enough; the code
has to be different. Two ways to get that code: (a) have the framework's tool generator write it
from a brief describing the route, which is the "everything through the framework" route and goes
through the same checks every other tool gets; or (b) do what the robot-hands and vonc widgets did,
where a person wrote the widget's JavaScript by hand and it was installed through a migration.
Both of those hand-written widgets work today; the generated route is the one your August ruling
points at. My recommendation is (a), with (b) as the fallback if the generator cannot produce a
working streaming client after a couple of rounds, in which case that is a bug to file. Say if you
would rather go straight to (b).

**2026-09-03, evening, after you ran the deploy.** The chat route is live. A call from outside
with the finetuning.uk origin streams a reply from the demo model on the Hetzner box: about 3.9
seconds for the first call while the model loads, then under a second for a follow-up, with the
first word arriving in under half a second. Other origins are refused, and the robot-hands and
vonc tools came through both restarts fine.

Two things went wrong on the way and both are worth knowing. First, my instructions said to put
the five settings in the island's env file, and that was not enough: the compose file lists every
setting it passes into the container by name, so the container never saw them and the service
booted saying "playground not mounted". I added the block, copied the file over (after proving the
island's copy matched the one in the repo, so nothing else changed) and restarted once more. The
recipe is corrected and the trap is written up so the next tenant does not hit it. Second, when the
env paste went in indented, the fix-up read back the end of that file into this chat, which
included the gripper mailbox password. I did not repeat it, but it is in this transcript now, so
please rotate that password when convenient.

What is next on the playground is the widget on the page, which needs your call between the two
routes I described in the entry above (generated by the framework, or hand-written like the
robot-hands one). Until then the demo is reachable only by a direct call to the API, which is fine
for testing and invisible to visitors.

**2026-09-03, later in the evening. Stage B ran, and the result is half a win; your four answers are
acted on.**

Migration 641 went live at about half past seven, after you read it. I sent both pages through the
rebuild straight away. Technical-details came back at 19:35 and is live. The good news: all six
headings are now different, the broken closing tag is gone, and the three-model listing you called
unhelpful is gone, because the brief no longer asks for it. The bad news: the three middle sections
still all talk about the same thing, the model and its licence, under three different headings. The
repetition moved from the headings into the text.

I read the exact instructions the writer was given for each section. The new opening-line mechanism
is there and correct: each section got its own line and a list of the other five. But nothing in the
instructions tells the writer to open on that line or to leave the other five alone, and a separate
block labelled "incorporate this into the content" hands every section the whole page brief. Given
three identical section types and the whole brief each time, the writer wrote the meatiest part three
times. That is a prompt change, so it is the prompts lane's; I have sent them the evidence and a one-line
suggestion, and you would read the bytes again if they take it. The 443 bug lane has the same note.

Your-own-model did not save. The writer followed its new opening line and wrote a shorter, cleaner
hero, and the framework's guard against truncated saves refused the whole page because that hero was
just under half the size of the one it replaced, which had been padded with tool links in August. The
guard is doing its job on the wrong case. It retries by itself at ten past eight; if it fails three
times I will bring you the choice.

Your answers: the chat box will be generated by the framework, with the hand-written route as fallback
(webdesign.uk's box, I checked, was placed by hand, so it is a precedent for the fallback rather than
the plan). Leopardess has a note about the copy that may move, and nothing moves until you or the
domain thread say where. The catalogue shape is drafted as a discussion document in this folder,
`DESIGN_2026-09-03_examples_catalogue_shape.md`, built from your sentences: what an entry is, how one
gets in, accounts and the agreement click, removal, pricing by GPU class as a choice, and the ways
people could cheat with a control for each. It ends with questions for you and a build order. The
password rotation is left as you said.

**2026-09-03, on the three domains.** You said rationale.uk, egret.co.uk and proverb (.co.uk and .uk)
could each take one angle of what you offer, narrower and in more depth, proverb first. I have written
that down where the finetuning plan keeps your direction, and added it to the note leopardess has, so
they know the copy leaving finetuning.uk is more likely heading to proverb than to them. I have not
started any of those sites: each is its own site row and brief through the framework, and that belongs
with the domain thread you are talking to. When proverb's brief is being written, the finetuning lane
can hand over the list of what it would give up, page by page.

**2026-09-03, late evening: both rebuilt pages are live; the chat box exists and is one step from the
page.** Your-own-model saved on its second try, without anyone loosening the guard: the writer's second
hero was longer, and that was enough. Both pages now have all-different headings, and both show the
same remaining problem in the same place: the three middle sections each say the same thing under
different headings. Two pages, one cause, and the prompts lane has it as a formal diagnosis.

The chat box was generated by the framework on the second run. The first run refused because the
playground page was a content page and the framework will not quietly turn a live page into a tool
page; I made that one-column change, since the playground page is the tool's page, and kept the old
value for rollback. The generator then attached the chat box as a seventh section and queued a page
re-render to publish it. It also queued a rewrite of the page's text with its stock "tool page" brief,
which would have replaced the six booking sections you approved; I cancelled that before it ran. The
box is visible on the site as soon as the re-render goes through, and I have a real-browser test ready
that types a question into it and waits for the streamed answer.

**2026-09-03, 21:35 BST: the playground answers.** The chat box is on the playground page, second
section down under "Try the demo model", and it works: I drove the live page in a real browser,
typed "In one sentence, what is fine-tuning?", pressed Send, and the model's answer streamed into the
box word by word from your Hetzner server. Nothing on the page reaches anywhere except our own route.
That is the whole chain you asked for this morning, live: the model, the route, the page.

Three things for you to look at when you have a minute. The words on the chat box itself (the heading,
the one-line intro, the disclosure) are the ones I put in the brief; change any you dislike and it is
a small rebuild. The generator also wrote a companion page, "Understanding playground chat", at
/guides/playground-guide.html; it reads sensibly but nothing links to it, so it is invisible until you
say whether you want it. And the playground page's own text is still the booking copy from yesterday;
folding the tool into it properly, as the centre of the page, is a rebuild I would want you to read
the brief for first.

**2026-09-03, 22:00 BST. Your three questions, and something you need to know before we publish
comparisons.**

Cost: no, the playground does not cost you money per use. It runs on the Hetzner box you already pay
for, on its processor, with no paid AI service anywhere in the path, and each visitor is limited to
60 messages an hour. A GPU only costs money when one is hired for a booked hour, and nothing on this
page hires one.

GPU: no, it is the box's two ordinary processor cores; I read that from the box itself. It feels
quick because the model is very small, about a gigabyte, and replies are capped at 150 words.

The steps and the examples: before writing "here is what the model does for you", I measured what it
does. The demo model was trained on your own emails, in the shape "write this email in my voice,
situation: ..." and "reply to this in my voice". When a visitor types a general question, that is not
what it was taught, and it answers like any small untrained model would. So I ran the exact kind of
prompt it was trained on through both the untrained model and yours, side by side, on prompts it had
never seen. The result is honest and not flattering: on one email it improved a little (shorter,
"holiday" instead of "vacation"); on another it got worse; on two it simply repeated the message back;
on a summary it gave up after the title. A few hundred examples cannot teach a voice, and this run
shows the gap rather than the product. All of it, word for word, is in
`COMPARISONS_2026-09-03_base_vs_finetune_demo_model.md`.

So the decision is yours: publish the comparison now with a plain explanation ("this is what a first,
small fine-tune does; the effect is modest and uneven"), which is educational and true; or improve the
model first (more of your writing, and a training fix for the repeating) and publish when it shows
something. I would not put the comparison on the page under the offer's promise as it stands. Tell me
which, and the page brief follows, for you to read before it runs.

**2026-09-03, 22:05 BST.** Decision taken: train it better first, with someone else's writing that has
a defined character, because your own sample is too small. Nothing is published from tonight's
comparison, the chat box stays live as it is, and the page text waits for the model that will actually
show something. When you have a corpus in mind, the things that decide whether it will work are the
right to use it (public domain is the safe class), whether it is the kind of writing the demo asks for
(letters, replies, short pieces suit "write this in my voice" far better than long fiction), and how
much of it there is (we want a couple of hundred short pieces, not twenty). The repeating fault in the
last run gets fixed before the next one, and this time the run measures the model on writing it has
never seen before anyone calls it a success.

**2026-09-03, 22:30 BST, the homepage cards and infographics.** Two things are now in motion, both
through the framework. First, the design critic is looking at the site: it screenshots the pages at two
widths and writes a senior designer's report, and its brief already asks the questions you asked, about
card and grid composition, cards that restate their neighbour, and whether the site could be mistaken
for a template. Second, the request in your words, with the page's section list, the library's
carousel-like alternatives and the rule that your approved copy must not be regenerated, has gone to
the editorial-design-uplift lane, whose whole job is making page families look far better with imagery,
graphic treatments and charts. No single agent does "research, choose and apply" in one go, and the
choice of components is not something I should make by hand, so that lane is where the imaginative part
lives. They will write to you here or in their own log.

**2026-09-03, 22:35 BST.** The explanation and the example pair are in the chat box's component now,
under the box, with your sentence as the framing, and the six sections of the page are untouched
(checked by hash). It reaches the live page when the queued re-render runs; I'll test it again in the
browser then.

On the homepage: the design-uplift lane answered quickly and honestly. The card and carousel choice is
not theirs (their remit is the editorial feature pages) and it turns out nobody owns "swap a page's
components", so I will do it, after the design critic's report lands, one slot at a time starting with
the case studies, and checking your words are byte-for-byte unchanged each time. They will do the
infographics. One thing you should know: the framework's site planner is currently instructed to
produce almost no infographics ("use sparingly, most plans will have zero"), and across every site ever
built it has produced exactly one. Making it produce them everywhere is a change to that instruction,
with eighteen site rebuilds queued behind it, and is the planner owners' call, so I have not touched
it. The narrow route for this page is to author its infographic entries by hand, which is what the
uplift lane will do. Words and figures in those graphics will be real HTML text over real facts, not
pictures of text, because the claims checks cannot read text inside an image.

**2026-09-03, 22:50 BST, three small things for you on the infographics.** The uplift lane has
scoped what they would draw: a concept diagram in the "what fine-tuning is" section (your documents,
the small trained part, a model you host), and the £99 against roughly $5,000 comparison, which is
backed by the research figure registered on 2026-08-18. They would not draw one for the departments
section, because a list of departments is a taxonomy and a graphic there is decoration unless it shows
something measurable. They also noticed the page has no "three steps" section as such; the steps are
described inside the fine-tuning section. So: do you want a dedicated three-steps section added to the
homepage, or is a diagram inside the existing section enough? And do you want the narrow route (these
graphics authored for this page by hand, reversible) or the fleet-wide change to the planner's
instruction? They will not write anything to the site until you have said, and until their own
session is cleared to.

**2026-09-03, 22:55 BST.** The explanation is live under the chat box: your sentence about five
articles and a handful of short emails, three prompts to try, and the same email prompt answered by the
untrained model, by yours, and what the person actually wrote, names removed. I tested it again in a real
browser with one of the page's own suggested prompts and the reply streamed in, a generic business
email, exactly as the page now says to expect. The design critic has started on the site.

**2026-09-03, 23:05 BST, the design critic's verdict and what I propose.** The critic looked at eight
pages at two widths. Site-wide it found the same thing you did: every page runs the same beat, navy
hero, cream three-column cards, navy CTA, then a huge footer, with the same card icon treatment and the
same hero picture on five pages, so the site reads as templated. On the homepage itself it liked the
hierarchy, called the six-card grid solid and the orange-edged blocks the strongest section of the
whole site, and faulted exactly one thing: the case studies, four cards that fall into three plus one
orphan. The full report is in the lane folder.

So the first card change I propose is that one: the case studies become a swipeable carousel, which
removes the orphan by construction and is the structure you asked for. It is not a simple switch,
because the carousel component takes its cards in a different shape and is rendered by an agent, so the
words could drift; I would copy your four cards' text across verbatim and check every title and excerpt
byte-for-byte on the live page afterwards, with a rollback ready. It is the first time anyone has
swapped a slot's component on a live page here, so I have written the steps down rather than doing it
at eleven at night. The simpler alternative the critic itself suggests is a card count that divides by
three, so three or six case studies instead of four, which is a content decision for you. Say "carousel"
or "three" or "six" and it happens first thing.

**2026-09-03, 23:30 BST, two imagery facts worth knowing.** The critic saw the same hero picture on
five pages; the whole site has thirty-eight hero sections and two pictures, thirty-five of them the
same one. Separately, ten pages already have their own hero image generated and deployed that nothing
displays. For four of them (use cases, case studies, approach, contact) the fix is wiring, owned by the
imagery lane's mechanism from bug 412, and I would only ask for it on your yes since it changes four
live pages. For the other six, tool and guide pages, the components cannot show an image without a
change that has already been tried once fleet-wide and rolled back because it showed the same picture
twice; leave those. And the ten "broken image" alarms that have sat on your case-study cards since July
were stale: every image returns fine, checked against a deliberately invented one that does not, so I
closed them and told the lane that owns the detector.

**2026-09-03, 23:40 BST, a correction to the hero images.** I said the fix for the four pages was
wiring. More precisely: the wiring mechanism is already built and running in the fleet, switched off
behind a setting that no agent yet turns on, and the lane that owns turning it on is the imagery-wiring
lane, by a decision recorded on 2 September. The reason this matters: those same pages were wired by
hand once already, on 26 August, nine of nine proven at the time, and today only three of the nine
still have their image. Hand-wiring decays within a week when images are regenerated. So the honest
question for you is not "wire four pages" but "arm the built mechanism", which is that lane's to carry
and now has the nine-to-three evidence attached. A stop-gap on the four is a legitimate choice too, as
long as it is called one.

**2026-09-04, 13:00 BST: the case studies are a swipeable carousel on the live homepage.** The five
cards' titles, text, categories and client lines are on the page word for word (checked by machine
against what was stored), the other five sections are byte-for-byte what they were, and a real browser
confirmed the row scrolls. The framework's re-render queue had gone quiet for three quarters of an hour,
so I used the documented one-page bypass; on my first go I pointed it at the playground page by
mistake, which re-rendered that page unchanged, and the second go did the homepage. What the carousel
cannot carry, and no longer shows in that section: the one-sentence intro, its own "Start your project"
button (the page still ends with one), and the five card pictures, which are still on the case-studies
page. Two things for you to say: whether your "more than three cards" rule should now reach the
six-card "what we build" grid and the five-card departments grid, which the critic called solid; and
whether you want the pictures back, which means the image-card carousel that no site has used yet.
