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

## 2026-08-21 evening — live, and the useful part was watching it fail

Both halves are running now. The daily check on our site instructions has been going since the 20th;
the writing check went live yesterday morning on the fresh build, and the newest build this evening
carries the last two corrections.

What I want to tell you about is not that it works, but the two ways it was wrong, because both were
invisible to every test we had.

**Three minutes after it went live it was doing nothing at all.** It found the mannerism correctly on
the first page and then could not repair it, because it could not find a model to call — this writer
keeps that setting on a different step from the one we added. Detecting perfectly, repairing nothing,
and reporting a status that looked orderly. We caught it in one query only because a reviewer had
insisted two days earlier that "the machine broke" and "nothing needed changing" must not look the
same. I nearly waved that objection through as paperwork. It paid for itself within a day.

**Then the first busy page found a real bug in my repair.** Where several of these phrases sit in one
block of text, each rewrite was being applied to the original version of that block, so they wrote over
one another — six rewrites accepted, one actually applied, and the report claimed six. That is fixed
and live.

**And I nearly certified it with a check that could not fail.** I asked whether each rewrite's text
appeared in the stored page. Five of six said yes. It was meaningless: these edits trim the end of a
sentence, so the beginning reads the same whether the edit landed or not. The real question — is the
phrase we removed actually gone? — said three of the six were still sitting there. I have written that
particular lesson down four times in this area now, and I still did it, on my own work, with an answer
I wanted. It is logged as such.

In between those, it did the job well: on a real page it found five instances, left a regulatory
sentence alone, allowed two under the per-page allowance, and rewrote two — *"the result breaks down by
area rather than giving you one verdict:"* became *"the result breaks down by area:"*, with nothing else
on the page touched.

One check is still outstanding and it is only waiting for ordinary traffic: a busy page built since
this evening's release, so I can watch a repaired sentence arrive in the stored page rather than in a
status. The fleet has been quiet since three o'clock.

Nothing has changed about the three pages you read. That tagline is in that site's own instructions,
which order it onto four page types, and the check deliberately leaves alone anything the instructions
supplied. Nine of our twenty-five sites are in the same position, and that is a decision about
positioning rather than a thing code can settle.

---

2026-08-22, morning. The missing piece went in. Last night we found that the check was doing its
rewriting and then losing the result on the way to the page — the sentence was fixed in memory and
the page was built from the unfixed copy. The code change that makes the fixed copy travel with the
page went out in this morning's release (I checked the running software itself on both servers, not
just the release notes), and I then switched the page-builder to read from the fixed copy. Both
halves are now live.

What's left is the proof: one ordinary page that trips the check, built from now on, where the
rewritten sentence is what the stored page actually says. The system is busy this morning, so that
should happen on its own; I'll read the first one that qualifies. Until a page proves it, treat
"the check now changes pages" as expected rather than established.

Later the same morning. A page came through that tripped the check and I could follow it the whole
way: it found six of these constructions, rewrote one — "…is the right next step, rather than trying
to borrow your way around it" became "…is the right next step" — and left everything else on the page
alone. I could see the rewritten sentence in the section the system built and in the finished page it
assembled. That is the part that was broken last night, and it is working.

The page still didn't get saved, and for a reason that has nothing to do with copy: a separate safety
guard refuses a save when a page's layout suddenly loses half its structure, and this page's stored
version is four days old while the layout builder has changed since. So the guard did its job and
threw the whole save away. I've written that up for the people who own that guard, because it will
refuse every rebuild of those older pages until someone decides which version of the layout is the
right one.

So the honest position: the check now works end to end inside the build, and I still owe you one page
where the improved sentence lands in the stored site. I've left the exact test written down so it is
a single query when the next suitable page comes through, and I've deliberately not forced one — that
site is being rebuilt by another thread this morning and I'd be treading on their work.

Ten o'clock, and the last piece landed. The same page came round again on a retry, and this time
everything worked: the check found eight of these constructions, rewrote six of them, left two alone
because that site's own instructions supplied them, and the page saved. I then went and read the
stored page itself: none of the six phrases it removed are there, and all six of its replacements
are. So the improved copy is now what the site actually holds — that is the thing I owed you, and it
is done.

The rewrites are the sort of thing you would do with a pencil. "…rather than paying down the loan
itself" became "…", leaving "interest,". "…based on your budget, rather than theirs" became "based
on your budget". Nothing else on the page moved.

Two corrections to what I told you earlier today. The save that was refused an hour ago was not a
sign that older pages are stuck — the very next attempt at the same page produced a good layout and
saved normally, so that guard was doing its job on one bad roll of the dice and I over-read it. I've
withdrawn the question I'd sent to the people who own that guard, and logged the mistake.

Nothing has changed about the three pages you originally complained about. That still needs the
wording in that site's own instructions to be changed, which is a decision for whoever owns that
site, not something the check will ever do on its own.

Evening, after the new release went out. First thing I did was check the release had not undone us:
a deploy writes to the same configuration table our fix lives in, and if it had been reset the check
would have gone quietly back to being useless with nothing to show it. It had not — I read both
settings back off the live system, and asked the running software itself, on both servers, whether it
still contains the code. It does.

Then I did the one job still on our list. When this change was reviewed, the panel left a note asking
us to move a small piece of shared logic somewhere reusable, and to look at the other places doing the
same thing rather than move just one. Doing the looking changed the answer. It turns out the platform
already has a much better way of noticing a cut-off answer — the AI provider tells us directly — so
rather than build a rival, I wrote the new helper as a clearly-labelled fallback for the two
situations where that signal is not available, and said so in its documentation so nobody reaches for
the worse one. The review panel approved it first time, all reviewers.

One reviewer made a fair criticism: I had written that a problem I found in someone else's tool was
"reported to them", and that is not a real thing — nothing tracked it, so it would have evaporated.
I have filed it properly now.

I also caught an error of my own about ten minutes after submitting the work, and I caught it by
writing the documentation entry. Describing how my change related to a neighbouring one forced me to
read the neighbour, and the neighbour showed I had been comparing against the wrong number — the
limit we asked for rather than the limit actually applied. Harmless on the AI provider we use, but
wrong, and now fixed. Worth saying because writing things up is usually treated as the chore after
the work; this time it was the thing that found the bug.

Your three pages are still as they were. But something changed today that matters for them: until
this morning, rebuilding those pages would have looked like it worked and changed nothing. Now a
rebuild would genuinely fix six of the seven bits of copy involved, including both sentences you
quoted at me. The seventh is the tagline, which comes from that site's own written instructions and
needs a decision from you rather than a fix from me. I have not rebuilt them yet, and deliberately:
another thread is in the middle of repairing that site's factual claims today, and a rebuild fired
now would most likely fail on that rather than on anything to do with copy.
