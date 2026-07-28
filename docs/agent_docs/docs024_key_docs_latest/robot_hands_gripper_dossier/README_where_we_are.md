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
