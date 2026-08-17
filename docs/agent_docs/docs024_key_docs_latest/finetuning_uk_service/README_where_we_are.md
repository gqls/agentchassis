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
