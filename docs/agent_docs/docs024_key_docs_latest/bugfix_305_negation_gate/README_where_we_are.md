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
