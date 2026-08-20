# SUMMARY 2026-08-20 — the rule becomes a check

First in this lane's series. Written because the work reached the point where it can be described
without a list of caveats: the mechanism exists, one half of it is running in production, and the
other half is waiting on a build somebody else will run.

## What we're trying to do

Stop a particular habit of machine writing from reaching our pages: saying what a thing *isn't* in
order to say what it is. *"The registry shows you what's possible, not what survives production."* The
owner read three of our pages, quoted two sentences of exactly that shape, and asked for two things —
fix those pages, and make sure that sort of copy never leaves the framework again.

## Where we've come from

Another team had already done the hard diagnostic work and, in the course of it, corrected themselves
twice in public. Their conclusion is the foundation of everything here: the machine is partly doing as
it is told. The instructions for the complained-of site hand the writer a house tagline built on the
very mannerism, and that phrase goes in about thirteen hundred times and comes back out four hundred.

Two attempts had already been made to fix this with words. A rule was added to the writer's
instructions in July; its own author recorded a fortnight later that it had not worked, because a rule
competing with twelve thousand characters of other instruction loses. A fuller house style shipped in
August and says, today, "say what a thing is rather than what it is not" — and the writer produced the
mannerism again the evening before this work started, for the very site that prompted the complaint.

There is a stronger version of that lesson in our own records, from a team who deleted a rule and left
its worked examples in place and watched the behaviour continue unchanged: *a rule can only name a
form; what goes wrong is an instinct.*

## What we've done

**Made the rule mechanical, at the place where the copy is written.** Every section our framework
writes is now checked for five specific shapes of this mannerism. On the agent that writes almost all
our page copy, anything beyond two on a page — or anything at all in a heading — is sent back once, as
individual sentences, with an instruction to say the thing directly, and the answers are pasted back
in. Nothing can be lost: if the model does not answer, answers badly, or runs long, the original
sentence stands.

**Made every rewrite prove itself before it is accepted.** This is the part that took the longest and
matters most. The obvious design — ask again, keep whichever version scores better — rewards the
machine for swapping "X, not Y" for "X instead of Y". Same habit, new clothes, and the scoreboard says
you fixed it. So each rewrite is checked for the mannerism in nine other grammars, for invented or
dropped numbers, for lost links, for changed markup, for invented names, and — after the review
council caught what we had missed — for claims we cannot support. Every rejection is recorded with its
reason, which as far as we can tell gives us the first instrument we have ever had for *seeing* that
kind of substitution rather than suspecting it.

**Built the other half: a daily check on the instructions themselves.** It is live. Each morning it
reads every site's brief — only the part the writer actually sees, which on the worst site is a
quarter of the document — and reports where a brief hands the writer one of these phrases. Ten of our
twenty-five sites do. It separates three things that look alike and are not: a phrase handed over for
reuse, a contrast used to give guidance, and a required legal disclosure, which it leaves alone
entirely.

**Told the three teams affected before writing any code**, including the one whose site drew the
complaint and the one whose pilot page has the worst brief in the estate.

## Where we are now

The counting half will start working on the next platform build. The database change that switches on
the repair is deliberately parked until then, with its two preconditions written at the top of the
file rather than left in somebody's head. The brief-side check is running now and has already filed
ten findings and closed two of its own when they stopped being true.

**One thing has to be said plainly, because it will otherwise look like failure: this does not change
the three pages the owner read.** That tagline is not the writer's invention — it is in the site's own
instructions, which order it onto four different page types. The check deliberately leaves alone
anything a site's instructions supplied, because the alternative is the platform quietly overruling
what a site has been told to say. It counts it, reports it, and stops. Changing it means editing that
site's brief and rebuilding those pages, which belongs to the team that owns it.

We were wrong four times getting here and two of those were the dangerous kind — mistakes that look
like success. A pattern meant to stop the machine inventing names was broken in a way that rejected
*every* repair, which reads exactly like a strict guard working properly. And the new daily check's
first run flagged "we do not offer refunds" as bad writing; it is a policy, and our own house rules ask
writers to state limits like that plainly. Both were caught by looking at what the thing did on real
data rather than re-reading the code that did it.

## Where we're going

The review council returned our first submission with six objections that changed the code, and the
most valuable one found a hole we had not seen: the repair checked that a rewritten sentence kept every
number, name and link, and never checked whether it had introduced a claim we cannot stand behind.
Asking a machine to say what something *is* is exactly the pressure that produces "the definitive
source". A second round is with the council now.

After that, three things. The platform build, which turns the repair on. A week of traffic, which will
tell us whether the broadest of the five shapes — "rather than", present in nearly half of everything
we write — is a real fleet-wide tic or a pattern we drew too wide; the rejection log answers that
without anyone having to guess. And a conversation with the teams who own the ten briefs, because that
is where the sentence the owner actually read comes from, and no amount of code will take it out.
