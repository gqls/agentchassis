# README — where we are: the gripper dossier pilot (robot-hands.com)

*Owner's running log. Plain prose, append-only, newest at the bottom.*

---

**2026-07-24 — designed, and the first piece built**

This is the first real build of the per-site AI idea (the per_site_ai
workstream): a paid-shaped, produced deliverable on robot-hands.com. A visitor
chats with a small assistant on a new page, describes their pick-and-place
application, leaves their email, and gets back a link to a proper engineering
report — every gripper in the site's index scored against their actual
application with the physics shown, every figure traced to a manufacturer
datasheet, and an honest "nothing in this index fits" when that's the truth.

The design is settled and written down (DESIGN doc in this directory). The
shape in one line: the public half lives on the little isolated island server
that the gauntlet work already built (the public never touches our cluster,
and the visitor's email never leaves the island); the report-building half
runs inside the platform and publishes the report as a page on the site; the
island notices the page appear and emails the link.

Decisions you made today: shared sender address for the emails; testing on
the live site approved (with cleanup); soft-launch without a nav link first;
a £/$50-a-month cap on the new AI key. The tunnel to the island is live —
your authorisation click happened. The one thing still only you can do:
issue that second spend-capped API key.

Built today: the scoring engine — the server-side port of MatchMatrix v2's
physics, reading the same verified figures the tool uses, with the whole
verdict logic (Match / Marginal / Insufficient data / No match), the
conflict-note maths, and the fact block that will keep the report's prose
honest (the writer may only use numbers from it). Tests all pass, including
the case where nothing fits and the case where a figure isn't published —
both must be said plainly, never papered over.

---

**25 July.** The whole cluster-side half of this is now built, tested and
committed. What's left before a real visitor could use it is the island
service (the public-facing bit), an image roll, and the end-to-end tests.

The most useful thing that happened today was the review council telling me I
was wrong, twice, about things that mattered.

The first round came back "revise". The gap it found: my honesty gate could
catch an invented *number* and an invented *model number*, but not an invented
*vendor name*. If the writer had padded the shortlist with "you might also look
at Piab", nothing would have stopped it — the check needs digits to see a name,
and the general-purpose fabrication scanner is deliberately switched off on
these pages (it compares against the site's own figures, and every number on a
report is calculated fresh for that one customer, so it would reject every
honest report). I'd written that gap down as an accepted limitation. The
council said, in effect: that's precisely the thing you claim this feature
does, so don't accept it quietly. It was right. That's now closed, and the
list of vendors it checks against is read from our own product data, so it
grows by itself as we index more.

It also caught me copying rather than reusing: my new "pull from the island"
code was a near line-for-line copy of the code that pulls from the traffic
probe box. I've now extracted the shared half so it exists once, and moved the
live traffic-probe code onto it too — which is the riskiest change in the
batch, since that one is running in production.

The second round found something bigger, and this one is worth your attention.
When we ask an AI model for a long piece of writing and the answer gets cut off
at the length limit, the platform *keeps* the half-answer, marks it, and reports
success — on the reasoning that a marked half-answer is better than a hard
failure. That's a fair trade, but only if whoever receives it reads the mark.
The council asked how many places read it. Nobody had checked. The answer:
**118 places ask a model for something, across 58 agents; 5 read the mark.**
Two orchestrations are carrying a truncation marker right now. So a cut-off
answer can quietly become a finished-looking piece of work almost anywhere in
the system. I've written that up as its own bug (076) rather than bolting a
patch onto this feature — the obvious fix is to make it fail loudly by default,
but that would change behaviour in 113 places at once, and nobody has measured
what would break. That measurement should come first.

I also caught two of my own errors worth recording. I stamped "reviewed by the
council" on a commit without reading the verdict — it was a "revise", not an
approval. It's in the permanent history now and I've logged it in the
wrong-calls file; the check was one query and took thirty seconds when I
finally ran it. And the report page would have gone out to a paying customer
completely unstyled: the way these pages are assembled means a component's
styling has to travel *with* the rendered page, and robot-hands.com has no
styling for a page type that didn't exist until this week. I've written a test
that renders a report and checks every single style name is actually defined —
it found two more I'd missed straight after I'd checked by eye.

Nothing here is live yet. The report pipeline is committed but inert until the
next image roll, and both scheduled jobs are seeded switched off, on purpose:
the plan is to prove the builder on a hand-made request first, and only then
let real visitor requests start flowing.

Still waiting on you for the one thing: the spend-capped API key for the
island. Everything up to that point can carry on without it.

---

## 27 July — you caught something no machine did, and the reason why turned out to matter more

You asked how the dossier would fit with the other tools before carrying on.
That question found a real problem. I had designed the visitor-facing half of
this as a new service on the island machine — its own database, its own AI key,
its own rate limiter, its own web address. One day after I wrote that, another
thread put `tools-api` on that same machine: a shared public API built from the
start to serve many tools across many sites, with all four of those things
already in it. Mine would have been a second copy of the lot, on a box with one
processor and 2GB of memory. There was also a wrong web address already
committed, which the machine would have rejected on every attempt, quietly,
forever.

Nothing caught it. No hook, no council, no check. You did, two days later.

What I want to tell you is that the reason is more interesting than the mistake.
I did search for something like this before designing it, on the 24th, and the
search was **right** — `tools-api` genuinely did not exist that day. The problem
is that nothing ever looks again. Every review we have judges a proposal against
a snapshot of the world, and this codebase moves fast enough that a design
document outlives the world it was written in. That is the actual failure, and
it is not fixed by being more careful.

Chasing it turned up three things, all of which I checked myself.

The first is that the decision lived in a document, and our review machinery
refuses to read documents. That refusal is deliberate and sensible — reviewing
every design doc would cost real money, and we wrote 72 of them this month. But
it means a document that decides to build a new service is refused by the one
thing that would have argued with it.

The second is that we already have a council member whose entire job is asking
"does this already exist?" It asks its questions about the code into thin air:
the step that answers them was deliberately left off that council, on the
reasoning that the author will go and look. The whole reason that seat exists is
that authors don't look.

The third is the one I'd actually fix. The search index that member relies on
records what our functions are *called* but never what they *do* — names and
signatures only, never the contents. So a search for a web address, a table
name, or any piece of text inside a function comes back empty, and the member
whose job is checking claims that something doesn't exist is being handed
made-up emptiness. Its own worked example, written in the documentation, cannot
work. I've filed that.

You asked whether the answer is a council member, the diagnosis loop, or the
architecture council. My honest answer is none of the three. A council member is
the wrong tool because "does this exist" is a question a search can settle, and
we already have two members holding that job — they need their instrument
repaired, not company. The diagnosis loop genuinely cannot do it; it is built to
narrow onto one bug and refuses to widen. And the architecture council already
exists — there's a thread on it that you were ruling on today. I've added our
measurements to it rather than starting a fourth thing, which would have been
funny in a bad way.

What I'd actually build is small. A document that proposes a new program gets
told which programs already exist. I ran it against the last 1,500 commits of
real history: it fires about once in 150, which is normal for our checks, and it
fires on the exact commit that opened this workstream — two days before you
asked. More to the point, it is free, so it runs *again* every time the document
is touched, and on the 25th and 26th it would have printed `tools-api` in that
list, newly arrived. That repetition is the whole value. No council re-reads
itself for free two days later.

One sentence I'd like you to rule on, because it settles a lot of arguments
cheaply: **divergence is allowed when it is parameterised and forbidden when it
is copied.** A second way of doing something is fine if it's a row in a table;
it's not fine if it's a second copy of the code. There's a lovely example of the
rule being broken — two exports of essentially the same thing sit ten lines apart
in the same file, one of them explicitly labelled "nothing site-specific may be
hardcoded here", and nobody noticed.

On scale, the number that matters: nine of our 296 pieces of pipeline machinery
exist for two sites out of a thousand, and five of those nine shipped in one
week. That is the thing that doesn't reach 1,600 domains — not the cluster, not
the hosting. A paid product per site that each needs its own hand-written code
is a staircase we can't climb. We already do this correctly elsewhere, for
company data, where a new industry is a row in a table rather than new code.

You've said finish the pilot first and generalise after, which I think is right
— prove the shape works on one site before paying to abstract it from a single
example. I've also put together a consolidation list, and the most useful thing
on it may be the item I'm recommending we *don't* do: eight copies of the same
small health-check server look like the tidiest possible win, and they turn out
not to be identical at all, so merging them is eight risky changes to the thing
that tells Kubernetes our services are alive, for no benefit whatsoever at any
number of domains. I'd rather close that one off as "won't do" than leave it
sitting there looking available.

The genuinely missing capability, and it surprised me: we cannot send email.
Not anywhere in the code we actually build and deploy. The only working mailer
in the estate lives in idea.uk's box, outside the build. Everything that emails
a customer a link — including this dossier — depends on that being fixed once,
properly, rather than copied a third time.

---

## 27 July, evening — it works. Two real dossiers are live on robot-hands.com.

The whole chain ran, end to end, in production. A request goes in; the physics
gets scored server-side against the ten real grippers; the model writes the
prose around those numbers; the honesty gate checks every figure and name; a
page is built, committed and deployed; and a little status file appears beside
it saying "ready". You can click both of them:

- `robot-hands.com/reports/d1a371be-04a5-4ee6-b744-d64c6fd9e7c4.html`
- `robot-hands.com/reports/29c3f8aa-3246-4a81-be8a-1e6b237cc467.html`

Neither is linked from anywhere on the site, which is deliberate.

The first is the ordinary case — a 2.5 kg steel blank. What I checked for is a
string that could not appear by luck: the actual arithmetic, substituted with
that request's own numbers, printed on the page. It's there. I also checked a
formula that should *not* be there, and it isn't, because a test that can only
pass isn't a test.

The second is the one I care about more. It asks for something we cannot do —
half a tonne of glass, washed down at high pressure. The report says so, in
plain words: *"No gripper in this index meets the requirement."* It doesn't
hedge, doesn't offer a near-miss, and doesn't recommend a purchase. And it
still counts as a **success** and still gets delivered, because an honest no is
a real answer and the machinery treats it as one. That was the part most likely
to go wrong and it went right.

The failure path got tested too, though not on purpose. I made three mistakes
today and two of them drove the pipeline into its failure branch — which is
exactly where I'd want to find out that it works. Each time it stopped, refused
to publish anything, and marked the job **failed** rather than quietly claiming
success. That distinction is the whole reason I added that piece last week.

The three mistakes, since they're the useful part:

The scheduled job was pointed at the wrong internal address — one that nothing
listens on. This is a nasty one because everything upstream looked perfect: the
scheduler said "message sent successfully", said "task triggered", and updated
its own timestamps. All true, and all meaningless, because nobody was on the
other end. You can only see it by looking downstream and finding nothing there.
Eighteen of our eighteen live jobs use the right address; mine used the default
that comes with the database column, which turns out to be a trap.

I'd also skipped a step: the gripper measurements themselves were never loaded,
so the scorer had nothing to score. It failed with an error naming the exact
file I'd missed, which is the kind of error message worth having.

And my test request used a made-up reference instead of a proper one, which the
page builder refused — correctly, since that reference becomes the page's web
address.

One correction I want to own, because it's the same lesson as this morning's.
I looked at a run twenty seconds in, saw it sitting at a step, and wrote into
*another team's bug file* that it was stuck. It wasn't — it finished ninety
seconds later. The tell was in a table I had written into that same file
minutes earlier. I corrected it before anyone acted on it, but wrong evidence
in someone else's open investigation is worse than no evidence, and I'd rather
record that than tidy it away.

Also worth knowing: after the job said "complete", the second page was still a
404 for about two minutes while it made its way out to the CDN. I nearly wrote
that down as a failure. "It's deployed" and "you can read it" are not the same
statement.

Where that leaves us. Everything inside our own system is proven. What's left
is the public-facing half — the form people actually fill in, and the email —
and that now belongs inside the shared tools service rather than in a service
of its own. Thank you for the key; it only gates that last part, so it hasn't
been holding anything up. Before we build it, I'd still want the email sender
and the abuse guard moved into shared code, because that's the exact moment
we'd otherwise end up with two of each.

I've switched the lane off again now the tests have run. The two pages are left
up for you to read; say the word and I'll clear them out.

---

**28 July, later that evening.** The review board came back approving the
gripper honesty fix. It's worth telling you how, because the interesting part
isn't the approval.

When the board looked at it the first time it said "revise", and the note I left
myself for the next session said the sticking point was a paperwork one — some
fields we'd added weren't formally declared anywhere. I picked the work up
tonight, went and read the board's actual record rather than my own summary of
it, and my summary was wrong. The reviewer who blocked it was blocking for a
completely different reason, and a much better one. It asked: you've built a
gate that spots a dishonest report — but does anything actually *stop* the
report? Or does it just make a note and carry on? It pointed at two of our own
open bugs where exactly that had happened before.

That is the right question to ask us, and I couldn't answer it from memory. So I
went and traced it. The gate refuses, and the refusal is a real refusal: the
engine's default when a step fails is to stop, and the step that builds the page
sits directly behind the gate, so a failed check can never reach it. It's the
opposite of the bug it was being compared to. I could show that in three
places rather than assert it.

Two things I want on the record because they're the sort of thing that quietly
costs you money later. First, my note to myself would have sent tonight's
session off answering the wrong objection entirely — I'd have done a tidy piece
of paperwork and the real question would have gone untouched, and we'd have
burned another review round finding that out. The lesson is small and dull:
don't summarise a verdict, quote it. I've written that up properly.

Second, the same note told me to "fix" a function name in the submission that
turned out to be correct already. Had I followed my own instructions I'd have
introduced a mistake into something a reviewer had merely asked me to *confirm*.

On the paperwork point that I'd originally thought was the blocker — I did go
and look at it properly, and the answer is that the mechanism people kept
pointing at doesn't fit this case. It's designed for one agent calling another;
ours is two steps inside a single job passing notes to each other. Half of it
isn't even read by anything: one of those two columns has no reader anywhere in
the codebase, which I checked rather than assumed. I said so plainly in the
submission instead of quietly declaring something untrue to make a warning go
away, and the three reviewers who'd raised it all withdrew.

There is a real gap underneath it — we have no way to declare "this step hands
that step these fields" — but inventing one inside a bug fix is precisely the
mistake we agreed to stop making, so I've written it down for a proper
architecture discussion rather than smuggling it in.

One reviewer left a caution I want to pass on rather than bury: my claim about
that gap is now the reason we're *not* doing something, and it hasn't been
independently checked by anyone but me. It asked that someone verify it before
it gets treated as settled. That seems right, and I've flagged it so the next
person doesn't inherit it as fact.

Nothing about the change itself moved between the two rounds. Same fix, same
six edits. All that changed is that four claims I'd made were replaced by
evidence. That's a fair description of what the board is for.

---

**2026-07-29 — I was wrong about the security bug I filed yesterday, and the
real one is more interesting.**

Yesterday I filed a bug saying that a visitor to the gauntlet tool could lie
about their own address — tell our server they were coming from somewhere else —
and so slip past the limit on how often they can use it, while also poisoning the
record we keep of who did what. I was confident, because we had proved exactly
that on idea.uk a few days earlier and the code here looked the same. I did write
down honestly that I hadn't actually tested it, and that someone should before
quoting me.

This morning I tested it. **It isn't true.** I sent the tool two requests
pretending to be an address in a range reserved for documentation, once each way
a visitor could try it. Both went through, and both were recorded as coming from
the real place. A visitor cannot lie to us about this.

**But the same test turned up something else, and this one is real.** Our server
isn't recording visitors' addresses at all. It records the address of the machine
sitting in front of it — the same value, every single time. I checked the whole
table: eighty-three visits recorded since the twenty-fifth, and every one carries
an identical entry. Not "mostly the same" — literally one value, for every visit
we've ever had.

Two things follow. The limit on how often the tool can be used is supposed to be
per visitor; because everyone looks identical, it's actually one shared limit for
the entire internet. One busy person uses up everybody's allowance, and nobody has
to be malicious for that to happen. And the column where we record who visited is
carrying no information whatsoever — if we ever need to ask "was that the same
person?", the answer we've been storing is the same word, eighty-three times. It
never looks broken, which is why nobody spotted it: it's always filled in, always
well-formed, always the same.

The reason my original claim was wrong is worth saying, because it's the sort of
mistake that repeats. I reasoned about our own program and stopped there. But
whether a visitor can lie about their address isn't decided by our program — it's
decided by the two pieces of plumbing in front of it, Cloudflare and a small
routing program on the island machine. Between them they clean up the lie before
our code ever sees it. Neither of those is in this repository; one is a config
file on a machine, the other is a supplier's service. So the answer simply wasn't
in the place I was looking, and no amount of care reading our code would have got
me there.

**One practical consequence you should know about.** The next job on this list was
to adopt a shared piece of security code we built and got approved last week, and
plug it into this tool. **It wouldn't have fixed this.** It would have arrived at
the same useless constant by a different route — and worse, it would have *looked*
like a fix, because the change would be sitting right there in the code where the
problem is. I've written down what would actually fix it: the visitor's real
address does reach us, just in a different label that our code currently ignores,
and it's one a visitor can't forge because Cloudflare rejects any request that
tries. That's a small change, but it belongs to the thread that owns this tool,
not to me, so I've put the evidence in the bug file rather than reaching into
their service.

I'd also flag the shared security code itself. It's fine where it's used today,
but it carries a note explaining why it's safe that describes the plumbing on
idea.uk, not the plumbing here. Anyone adopting it elsewhere would inherit the
reassurance without the thing that makes it true. That's worth fixing in the
package rather than in each place that uses it.

Cost of finding all this: about half an hour, three test visits recorded in their
table, and one throwaway copy of the routing program run on my own machine to
watch what it does to a request. Nothing was changed on the live service.

---

**2026-07-29, later — the thing I flagged is fixed, and the reviewers passed it first
time.**

I said yesterday that the shared security code explains why it is safe by describing
the plumbing on one of our two sites, and that anyone adopting it elsewhere would
inherit the reassurance without the thing that makes it true. That is now fixed, and
the fix is about honesty rather than cleverness: the code no longer guesses what is in
front of it — it makes the caller say. There are three ready-made answers and
deliberately no default, because a default that is right in one place and quietly
wrong in another is the whole of the original problem. The old behaviour is one of the
three answers, unchanged, so the fix we proved against a real production incident
earlier this month is still proved.

It went through the review council and was approved on the first round. It also
settled a question I had raised against my own submission — whether tightening a
shared mechanism needs the heavier architecture review — in favour of the normal route,
on the grounds that hardening a shared thing's contract *before* anything uses it is
the cheapest moment in its life to do it. Both reviewers who asked for something wanted
confirmation rather than argument, so I gave them measurements instead of prose.

One reviewer asked me not to let a loose end go stale, so I closed it the same day
rather than writing it down as a future task. The safety of the new arrangement rests
on a check about where a request came from, and that check's exact edge was implied
rather than stated. It is now written into tests, including one driven through a real
network connection rather than a typed-out address, and I deliberately broke the thing
those tests guard to confirm they would actually notice. That turned up something worth
knowing before this gets deployed anywhere new: a proxy sitting behind certain kinds of
shared internet address is not trusted, and its answer gets ignored. It fails safely,
but it limits where the code can go, and it is better pinned down now than discovered
by someone in production.

I should say that my first attempt at breaking those tests was worthless. The command I
used to damage the code silently failed to change anything, then the tests passed, and
for a moment I read that as "nothing broke". It is the most misleading possible result:
a test that passes because the sabotage never happened looks exactly like a test that
passes because the code is robust. I redid it with something that checks it actually
made the change first.

---

**2026-07-30 — the work is done; the last step is a conversation I can't have on my
own.**

Everything on this piece of work that is mine to do is finished, committed and
approved. What is left is that both remaining users of the shared code live inside one
service that belongs to another thread. So on the 29th I wrote up the evidence and a
ready-to-apply patch — following that service's own existing conventions rather than
imposing ours — and put it in their directory, which is what our convention says to do.

This morning I checked what happened to it. Nothing. They have been busy, six commits
since, but none of them near that service. So I looked at why, and found something more
useful than a reminder would have been: their own start-here document contains a line
saying, roughly, "the consolidation people may be in touch, nothing owed yet." I
checked when that line was written. Five hours *before* my patch arrived — and it has
survived four later edits of the same file untouched. So the next person picking up that
thread would read "nothing owed" with a finished patch sitting two files away.

That is nobody's fault, and it is the useful finding of the day: putting a file in the
right place is not the same as delivering it. Nothing in how we work tells a thread that
a new document in its own directory is meant for it. I have added a dated note to their
start-here document — added, with none of their words changed — saying the contact has
arrived, what the finding is, and that it concerns their service whether or not they care
about our shared code. Because it does: the visitor address that tool stores has been the
same single value in all eighty-three records since it went live, so the limit they think
is per-visitor is one shared bucket, and the identity they store has never told anyone
apart. That is measured, not reasoned, and no attacker is involved.

What I am not going to do is apply the patch myself. That service is theirs, they have an
open bug against it, and reaching into someone else's code is exactly what our
contribute-don't-fix rule exists to stop. So the choice is yours: leave it for them to
pick up, or decide it has waited long enough and route it. It is a live defect in a public
endpoint rather than a tidy-up, which is the only reason I am raising it rather than
letting it sit.

---

**2026-07-31, later that morning.** The regenerated gripper report is done and I have looked
at it. Both things you spotted in the chart are fixed, and I checked them the only way that
works here — by rendering the page and looking at the picture, side by side with the old one,
same inputs on both.

The number beside the longest bar used to read `6.42× (Insufficient` and just stop, as though
the text itself were damaged. It now reads `6.42× (Insufficient data)` in full. The two little
grey captions under the chart used to be printed on top of each other into something
unreadable; they are now on two separate lines. And the third thing, which you did not ask for
but which was the same underlying mistake: the two bars that run off the end of the scale used
to be drawn exactly as long as each other and exactly as long as an honest three-times bar, so
you could not tell a capped bar from a real one. They now end in a point, like an arrow,
and the bar that isn't capped still ends square. So the chart now tells you when it has
stopped measuring.

Worth saying why this needed a human eye. Nothing automated could see either fault. A clipped
label is still the correct text in the file — it is the drawing area that cuts it off — so
every check we have passes while the reader sees a corrupted number in a report full of
computed figures. The natural conclusion for a reader would have been that our scoring is
broken. That is the expensive kind of bug and the cheap kind to miss.

One small thing I noticed and have deliberately not touched: on five of the rows the number
sits on top of one of the dashed vertical guide lines. It is slightly untidy rather than
misleading, and it was exactly the same on the old page, so it is not something the fix broke.
Say the word if you want a pass at it, otherwise I will leave it.

The page is `robot-hands.com/reports/bf3765d6-befe-43a8-b1cd-ca5c210f39e9.html`. It took eight
and a half minutes, against twenty-seven for the first one on Monday.

Two things are now waiting on you and I am not going to move on either unasked. The first is
the tidy-up: there are now three test pages live on that site plus the test rows behind them,
and you asked to see them before anything was cleared away — that is now everything on the
list. The second is unchanged from yesterday: which order you want the shared limiter work
done in relative to the other thread's distribution work.

---

**2026-07-31, same morning — and this one is good news I did not expect to be writing.**
The shared-code question I put to you yesterday has resolved itself, properly.

You routed it to the other thread ahead of their distribution work, on the grounds that the
identity column is what the experiment gets measured on. They took it that evening, wrote it
up as a small shared piece rather than a patch buried in a handler, deployed their side of
the island this morning, and put it through review. It is live.

What matters is how they proved it, because it is better than the test I had asked for. I had
said: show me more than one distinct visitor key, because simply checking that the bad value
has gone also passes on unfixed code. They did something stronger. Before sending a single
request they worked out, independently, what the stored value *ought* to be if the fix
worked — the visitor's real address as Cloudflare reports it, put through the same hashing
the code uses. Then they sent a real request and the stored value was exactly that. So the
column did not merely change; it changed to the number predicted in advance. A changed but
unpredicted value would only have told us "not the old constant".

I checked it myself rather than take their word for it, and it holds, with one detail they
did not mention that I think seals it. The old constant appears on ninety-five records, the
last of them yesterday afternoon. The correct key appears on three, the first of them
ninety-eight seconds after they swapped the image. Not one record on either side of that
moment carries the wrong key for its era. That is a clean cutover.

One honest caution, which I have written into the file so nobody quotes it too enthusiastically
later. All three of the new records are from one machine, so what is proven is that we now
store the visitor's own address instead of a constant — not yet that two people arriving at
once get told apart. Real traffic settles that and nothing here can. It is a cutover result,
not a distribution result, and the difference will matter when you look at the experiment's
numbers.

I have closed that bug. They deliberately left the decision to me, since the file was mine,
and the bar is met — fixed, live, and proven on the real thing rather than at the tag. I also
found and fixed one small piece of dishonesty that had crept into our own records as a result
of the good news: our register now said the shared guard "has its first real consumer", which
is true, but the sentence underneath still implied the whole package was in use. Only one of
its three parts is. The rate limiter itself is untouched — what changed is which visitor it
counts you as, not how it counts. That is now written down precisely.

The one thing I promised to clean up and am now deliberately *not* cleaning up: three test
records I left in their table last week. They have become the evidence. They are part of the
old block that makes the before-and-after visible, so deleting them would destroy the proof
of the thing we just proved. They stay, and I have said so where someone tidying that table
will see it.

So of the two shared pieces, the guard is done and in use, and the email sender is still
waiting — its only customer is the gripper form, which lands in the same service, so it
inherits the same ownership question. Nothing needed from you on that today.

---

**2026-08-04.** You asked what's next, I answered "start on the email sender, it's a small
self-contained job," and then I actually went and checked that before touching anything —
and it wasn't. The email sender has nobody to call it yet. Its only two intended uses are the
gripper form, which doesn't exist as a public thing yet, and the other team's paid-report
service, which lives on a different machine we don't build from here. Wiring it in with
nobody using it would just be code that looks finished and does nothing — I've seen that
trap before and didn't want to walk into it myself.

So the real next step is the same one flagged before: someone needs to build the public
gripper form as part of the other team's existing public tool service, not as a new thing of
its own. That's their service, so it's the same kind of ask as the visitor-identity fix you
routed to them a few days ago — and that one worked cleanly. I've written down exactly what's
needed so it's ready to hand over whenever you want to route it the same way.

One separate bit of good news while I was checking: the intermittent bug that killed the
first attempt at showing you the fixed chart page has since been properly fixed and is live —
so regenerating a report for real, when the time comes, won't run into that again.

---

**2026-08-05.** You said go ahead, so I've written up the actual proposal for the other
team — what needs building, in their own service, so the public form and the email sender
both get somewhere to plug into.

One thing worth knowing: their own design note said the service was already built to host
more than one tool. I checked, and it isn't yet — it's wired for their one tool only. That's
not a problem, just means the honest size of the ask is "restructure a bit, then add ours,"
not "slot ours into a shelf that's already there." I wrote that down plainly rather than let
it surprise them halfway through.

The rest of the write-up is deliberately concrete rather than a wishlist: which of their
existing pieces we can reuse untouched, which need widening, the exact message format the
cluster is already sending and waiting to be answered, and where in their code the one email
actually gets sent from. Two things still need you specifically: a spending-capped key for
this tool's own AI calls, and the email server login details — neither blocks them starting,
both block the first real message going out.

It's saved and ready. Whenever you want to hand it to them the way you handed them the
visitor-identity fix, it's there.

---

**2026-07-31, later still.** The cleanup is done, and I checked it properly rather than
trusting the status.

Removing a test page turned out to need two separate steps, not one, because of something
we'd already learned the hard way on this same site: marking a page "archived" in our
database does not take it off the internet. The actual page is a file, pushed by our build
system to the storage that really serves it, and that file does not care what our database
says about it. So I did both: removed the four test entries and the three page records from
the database first, then removed the eight files themselves — the three report pages, their
data sidecars, the failed one's status file, and one stray test file from the very first day
that had never been on anybody's list.

I did the database part myself. The file part needed a permission I don't have by default, so
I handed you the exact commands rather than find a way round it, and you ran them.

Then I checked, rather than assumed. All five of the removed pages now return "not found".
The two ordinary pages I checked as a control — the homepage and one of the calculators — are
still there and unaffected. The build system's own log confirms the redeploy that removes
them actually ran and succeeded, not just that the commit was made.

So that's fully closed. Nothing from this pilot is left live except the real thing: the
scoring engine and page-builder, proven and ready for whenever the public half is built. The
only two things still open are the ones from before — the shared email sender waiting for a
customer, and which order you want the visitor-identity fix relative to the other team's work,
which I gather is already settled since they took it and it's live.

---

**2026-08-10, out-of-order note.** Two entries dated 2026-08-04 and 2026-08-05 appear earlier
in this file, above this one — I added them after an older paragraph instead of at the true
end, so they read as if they came before this 07-31 entry when they actually came after it.
Nothing in them is wrong, just misplaced; I'm noting it here rather than moving anything,
since this file doesn't get rewritten. Read them out of position: 08-04 (the email-sender
correction) and 08-05 (the proposal for the other team) both belong chronologically after
everything above this note.

You've since created the spend-capped key for this tool — I checked it actually works (a free
check, no cost), and tightened its file permissions on the way past since it was readable by
more than just you on this machine. The email side is still waiting on the actual server
details from your hosting panel.

---

**2026-08-15.** You sent over the email server details, password included, straight in our
conversation. Worth knowing for next time: that's the kind of thing better saved to a file on
your end and just pointed at, the way you did with the AI key, rather than typed out where it
sits in plain view — I don't think anything bad has come of it here, just flagging it as a
habit for later.

Either way, I got it off the page and into a proper file straight away, and did the same
permissions check as before — it had the same too-open default the AI key file did, fixed the
same way. Then I actually tried logging into the mail server with it, rather than assuming a
line of text that looks like the right shape means it works. It does — real account, real
server, confirmed without sending an actual email out.

So both things you needed to provide are done now, checked rather than taken on trust. Nothing
is wired up yet — that still waits on the other team building the actual public form — but
whenever they do, both credentials are sitting ready and known-good.

---

**2026-08-16.** The public half — the chat box, the "send me the report" button, the bit that
emails you the link — is now written, in the same service the vonc debate tool runs on, exactly
the way the 5 August proposal asked. It is tested locally against a real database and run as a
real process here; it is **not on the island yet** and nothing public has changed.

Three things I found that neither the design nor the proposal knew, and each changed what I
built:

1. The design said the chat would collect fields called one thing and the cluster would
   translate them. I checked the live pipeline: it doesn't translate anything — it reads the
   fields by the names *it* expects, straight from what the island sends. So the chat now
   records the cluster's names from the first question. Built the other way, every request would
   have failed after the visitor was already promised an email.
2. The proposal said the island could check the cluster's secret key against a column in its
   sites table. That column doesn't exist on the island (its sites table is deliberately tiny).
   The key lives in the island's environment file instead, and has to be pasted identically
   into the cluster's config — the same "two places, by hand" the original config file always
   said.
3. The cluster-side config that points at the island still had the *old* wrong address in it
   from July. Fixed the file; it was never applied, so nothing was broken.

Two design choices worth knowing about. The new tool is **switched off unless its own AI key
is present** — so swapping the image onto the island is safe on its own, and switching the tool
on is a separate, reversible step (delete the key, restart, it's gone). And the emails: I wrote
first-draft wording for "your report is ready" and "sorry, we couldn't" because none existed
anywhere. Please read them before this goes live; they're short and easy to change.

What's left is all on-the-box or cluster-side, in a numbered order in the island runbook:
put the two credentials on the island (I have not sent them anywhere), check the mail port is
open *from the island* (the 15 August check was from this machine), build and swap the image,
create the tables, apply the cluster config with the same key, switch the pull task on, then
build the page and widget on robot-hands.com. I'd like your go-ahead before I touch the island.

---

**2026-08-16, later.** I did the one item on that list that needed no permission and no
secrets: checking whether the island can actually reach the mail server.

This mattered more than it sounds. When you sent the mail settings, I checked they worked —
but I checked from *this* machine, not from the server that will actually be sending. Our own
notes carry a hard-won warning from months back that some hosting providers silently block the
exact mail port we're configured to use, and specifically that the cPanel screen advertising
that port can be misleading. Which is precisely the route we took. If that warning had applied
here, the failure would have been an ugly one: someone requests a report, we build it, we tell
them it's coming, and the email quietly never sends.

**Good news: it works.** The island reaches the mail server on the configured port, the server
answers, and it offers to accept a login from that machine specifically. So the settings as
they stand are right and nothing needs changing. I did this without putting your password on
the island — it was purely a "can these two machines talk" check, so the actual credential
step is still untouched and still waiting on you.

Two things I corrected while there. Our internal note about mail ports was written from a
different hosting provider and stated too broadly — as if no cloud server can use this port.
That's true of the provider it was written on, not of ours, and the note is now corrected so
the next person doesn't design around a limit that doesn't apply. And I'd written in a few
places that the credentials would go into the cluster's secret store; that was wrong — this
service runs on the island, so they belong in the island's own config file. The other session
spotted it, and I've fixed it where I originally wrote it.

Everything else on that list still needs your go-ahead, and still needs the credentials put on
the box by you or someone you've authorised.

---

**2026-08-20.** The credentials are on the island now. You ran the write; I checked it before
and after.

Worth telling you about the thing we nearly walked into, because it would have been genuinely
horrible to debug. The password your hosting panel issued starts with a dollar sign. It turns
out the software that passes settings into our service treats a dollar sign as the start of a
placeholder — so it would have quietly swallowed the first few characters and handed the mail
server a shortened password. Nothing would have complained. The service would have started
normally, the setting would have looked present and correct, and the only symptom would have
been an email failing to send — at which point the obvious conclusion is "the password is
wrong", and someone spends an afternoon re-checking a password that was right all along. On
this pipeline that failure lands *after* we've told a visitor their report is on the way.

I found it by testing with a made-up password of the same shape rather than yours, and by
reading the value from inside the running container rather than trusting the tool's own summary
— which, it turns out, displays it differently from how it actually delivers it. The fix is to
double the dollar sign in the settings file, which the script now does automatically. I've
written it up so the next person hits the warning rather than the bug.

The write itself: I backed up the existing settings file first (it holds the other tool's live
passwords, so a bad edit would have taken that down), and afterwards confirmed all seven
settings are present, nothing was duplicated, the other tool's passwords are untouched, and the
running service never so much as restarted. I also confirmed the other tool still works
properly — and I'll admit I got that check wrong the first time, testing against a domain I'd
guessed rather than looked up, which gave a "refused" answer that told me nothing. Looked up the
real one, retested, got a clean pass.

Your password never appeared on screen at any point, including in my own checks — where I
needed to confirm it arrived intact, I compared fingerprints of it rather than the thing itself.

Next up is creating the database tables and swapping in the new version of the service. Those
are bigger steps than this one, and I'd want your go-ahead before starting.

---

**2026-08-25.** You asked me to check where things stood before going further. Good instinct —
five days had passed and I found something that would have caused a confusing mess.

First the state: nothing had moved since we put the credentials on the island. They're still
there, intact and correctly locked down. The island is still running the old version of the
service, which is expected — we haven't swapped it yet.

The next step was to create the database tables the new feature needs. Before running the
command from our own checklist, I checked it against the actual database — and it was wrong.
It tried to record the change in a column that doesn't exist. Because of how the command was
chained together, what would have happened is: the tables get created, then the record-keeping
step fails with an error. You'd be looking at an error message while the change had *already*
gone through, and our own records would say it never happened. The next person would likely
run it again. I've fixed the command and split it into separate steps so that can't happen.

I also checked the things that command depends on, rather than trusting them: the script only
adds things (it deletes nothing), it's safe to run twice, it refuses to run against the wrong
database, and — the one I'd have kicked myself for missing — the site identifier hardcoded in
it genuinely matches robot-hands.com's real identifier. If that had been wrong, every request
would have been filed against a site that doesn't exist, and nothing on the island would have
complained.

I couldn't run it myself: my own safety guard blocks changes to live databases, the same one
that stopped me deleting those test files back in July. That's working as intended and I
didn't try to get around it. The exact command is ready for you to run — it's the first thing
in the handoff.

Speaking of which: I've rewritten the handoff. That document had grown to about four hundred
lines of accumulated history, and anyone opening it fresh was landing in July. There's now a
clearly-marked "start here" section at the end with a simple table of what's done and what
isn't, the one command to run next, and the handful of things that would trip someone up if
nobody warned them.

Where we are: two of seven steps done, and the next one needs about thirty seconds of your
time.

## 2026-08-25 (evening) — everything I can do from here is done; two short runs of yours ship it

I picked the lane back up this evening and checked the island first: exactly as the
afternoon handoff left it — the database change (436) still not applied, the service
still on the old build from three weeks ago, your seven secret entries still intact
on the box.

Two things worth telling you plainly:

**I caught a config regression before it shipped.** The deployment copy of the
island's service configuration kept in the repo had fallen behind the live one: your
rate-limiter loosening from the end of July (done directly on the box) was never
copied back into the repo, while the new gripper settings existed only in the repo.
Following the runbook as written would have shipped the gripper settings AND quietly
put your rate limiter back to the strict defaults — no error anywhere, you'd only
have noticed when visitors started getting turned away again. The two copies are now
merged, and the runbook tells future sessions to compare against the box before
copying anything over.

**The safety guard held again, so the island steps are yours.** Same as last time,
my harness refuses to let me change anything on the island directly — it even
refused to let me save a ready-made script containing those commands. I've stopped
pushing at that; it's the guard doing its job. What I could do, I did: the new
service build is made, tested, and proven to contain the gripper code, and it's
sitting in a file ready to copy across.

**What's left for you — two things, a few minutes:**

1. Type `! bash ~/.config/gripper-dossier/ship-step1-migrate.sh` in this chat — that
   applies the database change, records it in the ledger, and checks its own work.
2. Then the three copy-paste commands in the handoff's evening block (they copy the
   merged configuration and the new build across, and restart the service — a blip of
   a few seconds on the public tools site).

Once those run, I take it from there: verifying the right build is actually running,
the four public checks, and the cluster half (the report-pull switch stays OFF until
the checks pass — that rule is written down and I'll follow it).

## 2026-08-26 — it's live on the island. One thing stands between us and switching it on: API credit

You ran both halves this morning and they went exactly to plan: the database change
applied and checked itself, the new build went across, and the service came back up
in seconds with the gripper machinery mounted.

I then verified everything from my side. The right build is genuinely the one
running (checked at the container, not the label on the box). Your rate-limiter
settings survived the swap — that's the regression I caught yesterday, confirmed
fixed. And five of the six public checks pass: the chat door refuses strangers and
admits robot-hands.com, the cluster's pull door refuses without the key and answers
properly with it, and the existing vonc.com tool still works untouched.

The sixth check — an actual conversation turn — failed, and the reason is refreshingly
simple: **the Anthropic account behind the gripper's dedicated key has no credit on
it.** The service handled it exactly as designed (polite "assistant unavailable" to
the visitor, precise reason in the log). Nothing to fix in the code or on the box.

Worth being straight about: when we said the key was "verified live" back on the
15th, that check used a free endpoint — deliberately, to avoid spending — which
means it proved the key works but couldn't possibly notice an empty balance. Lesson
logged: verify a paid capability with the smallest paid call.

**Over to you, one item:** put credit on (or raise the cap for) the account that key
belongs to — Plans & Billing in the Anthropic console. One caution from a previous
episode: if the page you land on already shows credit, you're likely looking at the
wrong organisation — the key's "Last used" (today, 09:55 our time) identifies the
right one.

The moment that's done, tell me: I run one real conversation turn, and if it
answers, I flip the last switch on the cluster and we watch a request travel the
whole way — chat to emailed report. That's also the day I write the milestone
summary we agreed on.

## 2026-08-26, later — it works. All of it.

You put credit on the key, and this morning the system had its first two real users
(both were me, wearing a visitor's hat).

The second request is the one to savour: a three-turn chat about aluminium castings,
a submission, and half an hour later the dossier link arrived by email — a real
96 KB report page carrying the visitor's actual numbers, built, checked and
published by the pipeline with nobody touching anything. That's the whole promise of
this pilot, demonstrated on the production system.

The first request proved the other branch: I deliberately left one detail vague, the
quality gate refused to publish a report that hedged about it, and the visitor got a
courteous apology email instead. The safety net works — but it also showed me the
first real product flaw: the chat happily accepts a vague answer that the
report-writer downstream is forbidden to hedge about, so a vague visitor currently
gets an apology when they should get one more question. Filed with fix directions
(bug 409); worth fixing before the site widget goes up, because real visitors will
be vague.

Two honest confessions from the morning, both already logged: I briefly called the
chat's behaviour a bug when it was actually being smarter than me (it refused my
"correction" of a value it had recorded correctly — I'd misunderstood what the field
meant). And when a build stalled for twenty minutes during a Kafka wobble, I
declared it dead thirty seconds before it finished on its own; no harm done, the
race fell our way, and the lesson is written where the next session will read it.

The milestone summary you can read aloud to someone is
SUMMARY_2026-08-26_gripper_dossier.md in this folder. Next build item: the widget on
robot-hands.com. The soft-launch call — quiet link or none — remains yours.

---

**2026-09-03, midday — the missing button, and why it was missing**

You were right that the page looked incomplete. The widget was never broken; it was
never given a chance to run.

The site puts its shared JavaScript file in the page's `<head>`, and the browser runs
that file the moment it reaches it — before it has read the rest of the page. Our
widget's first act was to look for the spot on the page where it should draw itself.
At that moment that spot didn't exist yet, because the browser hadn't got to it. The
widget was written to quietly give up if it can't find its spot, so it did. No error,
no warning, nothing in any log — just no button, for a week.

The fix is four lines: instead of looking immediately, the widget now waits until the
browser says the page is fully read, and only then draws itself. The site's own image
carousel has done exactly this for a long time; ours was the only interactive piece on
the site that didn't, which is a fair sign it was an oversight rather than a design
choice.

One awkward detail worth telling you, because it nearly bit. We cap this widget at 8KB
— it's downloaded on every page of the site, so it earns its size. The fix pushed it
over. The note the previous session left me said there was plenty of room; when I
measured it myself there was almost none, because the command that note recommended
had a flaw that made it skip the very line it was measuring. Had I trusted it, the
update would have been rejected. I've corrected the note and written the mistake up,
because the lesson is more useful than the incident: measure a limit where the limit
is actually enforced, not with a convenient approximation of it.

To get the room back I removed a comment line that was pure duplication — the system
already prints the widget's name and description directly above it — and joined up the
styling text, which changes how it reads in the source but not one character of what
it does. I checked that second point rather than assuming it: I ran both versions and
compared the result, character for character.

And I did not just read the fix and declare it correct. That is the exact mistake that
led to last week's false "it's live" — the code was present in the right file, the
page had the right placeholder, and neither fact meant a button appeared. So this time
I built a small simulator of a browser's page-loading order and ran both versions
through it. The old widget draws nothing, before or after loading finishes — the real
failure, reproduced. The new one waits, then draws its Start button. That old-version
run is the important half: a test that only ever shows the new code succeeding proves
nothing.

Where it stands: the fix is written, applied to the live database and committed, and
it has been sent through the reviewer council. The last mechanical step is a rebuild
of the site's JavaScript file, which is sitting in a queue — our site is currently last
in line, because that queue is ordered by which site has been waiting longest, and ours
only joined at lunchtime. Nothing is wrong; it just waits its turn.

Then it needs you. A button on a real page in a real browser is the only proof that
counts here, and it is the one thing I can't check myself. When the rebuild has gone
out I'll tell you, and the ask is simply: reload the page and see whether a Start
button appears under the explanatory text.

The soft-launch decision — quiet link, or none — is still yours and isn't affected by
any of this.
