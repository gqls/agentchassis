# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-20 morning — what the complaint turned out to be, and what we are building

The owner read three pages on one of our sites and said the copy looked like it had not been through
the framework. Both sentences he quoted do the same thing: they tell you what something *isn't*
before, or instead of, telling you what it is. *"The registry shows you what's possible, not what
survives production."*

It had been through the framework. Another team spent two days finding out why, corrected themselves
twice in public while doing it, and landed somewhere useful: the machine is partly doing as it is
told. The instructions for that site hand the writer a tagline built on exactly that mannerism —
*"deployed to production in days, not months"* — and that phrase goes into the writer more than
thirteen hundred times and comes back out in four hundred pieces of copy. It is, word for word, the
sentence the owner objected to.

But that is only half of it. The writer also does it unprompted: there are phrases in the output that
appear in no prompt at all. So fixing the instructions is necessary and not sufficient.

**What we are not doing is adding another rule to the prompt.** That has been tried twice. The rule is
in there right now — "say what a thing is rather than what it is not" — and the writer produced the
mannerism again yesterday afternoon for the very site that prompted the complaint. The team who
studied this wrote down why, and I think they are right: a rule can name a shape, but what goes wrong
is a habit, and prompts full of prohibitions grow new habits to replace the banned ones.

So we are building something mechanical instead. Between the moment the writer produces a section and
the moment it becomes a page, the section is checked for five specific shapes of this mannerism. If
there are more than two on the page, or any of them in a headline, the offending **sentences** — not
the whole section — are sent back once with an instruction to say the thing directly, and the answers
are pasted back in. Every rewrite is checked before it is accepted, and any rewrite that has simply
found a different way to do the same thing is thrown away and logged.

That last part matters more than it sounds. Three of my own first ideas were wrong and two of them
were wrong in a way that would have looked like success:

- If you ask for the whole section again and keep whichever version has fewer flagged phrases, you
  reward the model for swapping "X, not Y" for "X instead of Y". Same habit, different clothes, and
  the scoreboard says you fixed it.
- If you excuse any phrase that already appears in the prompt — on the reasonable ground that the
  brief asked for it — you accidentally excuse almost everything, because the house style document
  itself uses the words "rather than" six times.
- If you quote the style rule at the model when asking for a rewrite, you hand it a worked example of
  the thing you are trying to remove.

There is one thing I want to say plainly, because it would otherwise look like a failure later: **this
will not change the three pages the owner read.** That tagline was put there deliberately by the
site's own instructions, and the check leaves anything the brief supplied alone — overruling a site's
own stated voice is not something the platform should do quietly. Those pages get fixed by editing
that site's instructions and rebuilding them, which belongs to the team that owns the site. I have
written to all three teams involved today, before touching any code, so nobody is surprised.

We are also, on the owner's instruction this morning, scheduling the second half: a daily check that
reads every site's brief and reports where a brief is handing the writer a phrase built on this
mannerism. The other team built exactly that as a tool a human runs; the owner's view is that an
unrun check goes stale, so it becomes a job that reports every day, including on the days it finds
nothing — because a silent job and a clean one need to look different.

## 2026-08-20 evening — it is built, half of it is live, and one thing needs saying plainly

The mechanical check is written, tested and committed. It counts the mannerism on every page the
framework writes, for every site, whether or not anyone remembers to switch anything on — and on the
one agent that writes almost all our page copy it now also repairs it, by sending the offending
sentences back once and asking for the direct version. Nothing about it can lose a page: if the model
does not answer, or answers badly, or answers too long, the original copy stands.

That half only starts working when the next build of the platform goes out, which is somebody else's
release to run. The database change that switches it on is deliberately parked until then, with the two
things that must be true before anyone unparks it written at the top of the file.

The other half is live today. Every morning at twenty to eight a job reads every site's instructions —
only the part the writer actually sees — and asks whether the instructions themselves hand the writer
one of these phrases. Ten of our twenty-five sites do. It ran twice this afternoon, found twelve, and
then, after I corrected it, closed two of its own findings because they were no longer true. That
self-correction is the bit I am most pleased with: a check that can only ever accuse is a check nobody
can trust.

**The thing that needs saying plainly: this will not change the three pages you read.** The sentence
you objected to — *"deployed to production in days, not months"* — is not the writer's invention. It is
in that site's own instructions, which order it onto the homepage hero, the services hero, the footer
and every meta description. The check deliberately leaves alone anything a site's own instructions
supplied, because the alternative is the platform quietly overruling what a site has been told to say.
So it counts that sentence, reports it, and leaves it. Changing it means editing that site's
instructions and rebuilding those pages, which belongs to the team that owns the site — and they have
the exact queries to do it and to check it worked.

I also want to be honest about what was got wrong along the way, because two of the four mistakes are
the kind that look like success. A regex I wrote to stop the machine inventing names was silently
broken in a way that made it reject *every* repair — which reads exactly like a strict, careful guard
doing its job. And the daily check's first run flagged "we do not offer refunds" as bad writing. It is
not bad writing; it is a policy, and our own house rules ask writers to state limits like that plainly.
Both were caught by looking at what the thing actually did on real data, not by re-reading the code.

The review council rejected my first submission and it was worth every minute. One of its objections
found a genuine hole nobody in this lane had seen: the repair checked that a rewritten sentence kept
every number, link and name — and never checked whether it had introduced a *claim* we cannot support.
Asking a machine to say what something *is*, rather than what it isn't, is exactly the pressure that
produces "the definitive source" and "fully verified". Fixed: every rewrite now goes through the same
banned-claims check the rest of the estate uses before it is allowed anywhere near a page.

## 2026-08-20 late — the review said no, then yes, and the no was the useful half

The internal review council looked at this four times today. It asked for changes twice, then
**refused it outright**, then approved it. I want to record the refusal properly, because it was the
most useful thing that happened to this work.

Twice, reviewers said the same thing: the counting half of the check was switched on for the whole
estate by default, on two pieces of machinery that nearly everything uses, and it had arrived inside a
fix for one agent. Twice I answered by writing a document explaining why that was defensible — and
shipped the code unchanged. The third time they stopped objecting and vetoed it, with a sentence I
have written into our permanent notes: *"we wrote it down and routed it is not the same as it was
contained."*

They were right, and the mistake underneath was mine to make: I had taken a decision you made about
one specific case months ago and read it as a general permission. It is not; it was your call about
that case. So the counting is now switched on only where it is explicitly asked for — which today
means the one writer this fix is about. It costs us something real, and I would rather say so than
bury it: outside that writer, "the copy got better" and "nothing was checking" now look the same
again. Whether it should be switched on everywhere is written up as a decision for you or the
architecture track, and nothing is waiting on it.

The approving round still found one thing worth having. My rule for accepting a rewritten sentence
checked carefully that nothing had been **lost** — no dropped figures, no lost links, no mangled
formatting. It never checked whether something had been **gained**. Asking a machine to say what
something *is*, rather than what it isn't, is exactly the pressure that produces "the definitive
source" and "fully verified", and the only thing standing in the way was a list of banned phrases that
most of our sites have never filled in. That is closed now: a rewrite that reaches for a superlative
the original did not use is refused.

Where it stands tonight: the daily check on our site instructions is live and has found nine sites.
The writing check goes live with the next platform build; the database change that switches it on is
parked with two conditions written at the top of it. And the three pages you read still say what they
said — for the reason I gave this morning, which has not changed.
