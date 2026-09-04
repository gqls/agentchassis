# Where we are — the writer prompt that told the model the wrong thing

Plain-prose log, append-only, newest at the bottom.

## 2026-09-03 — what this was, and what it turned out to be

Six of our sites had pages that were planned, approved and then simply never appeared.
They sat there marked active, with other live pages linking to them, for weeks. On
loanzy.uk one of them was `/your-rights.html` — the consumer-rights page, on a site whose
whole licensed position is explaining borrowers' rights. Anyone clicking that link got a
404.

The builds were failing 119 times in a fortnight, always the same way. The error said the
content writer had produced a sentence where the page design called for a list. So it read
like an unreliable writer: the AI was being asked for a structured list of decision
points and kept typing prose instead.

**It was not the writer. It was us, telling it the wrong thing, in the instructions we
generate for it.**

When we ask the writer to fill in a page section, we don't hand it the design directly. We
build an example for it to copy — "here is the shape of the answer I want" — and that
example is generated automatically from the component's design. The generator only ever
carried the *names* of the fields, never their shapes. So where the design said "branches
is a list of outcomes, each with a label and a body", the example we actually sent said:

    "branches": "..."

That is a sentence. We showed the writer a sentence, so it wrote a sentence, and our own
safety check then correctly refused the page for containing a sentence. Every single time.
That is why there were 119 failures and not one lucky success — it was never a matter of
the model having a bad day. We were asking for the wrong thing perfectly consistently.

The proof is a single database row: the exact instruction we sent on 2 September at 19:07,
and the writer's obedient answer, stored side by side.

**Why nobody spotted it sooner is the part worth remembering.** Every piece of evidence
pointed at the writer. The error message quoted the writer's output. The failure count
counted writer failures. The component's design, when you open it, is completely correct —
and it was correct all along. The one thing that was wrong is the instruction, and the
instruction doesn't exist anywhere you can read it: it is assembled fresh each time. You
have to go and look at what was actually sent.

## What we changed

The generator now works out the real shape, however deeply nested, and shows it properly —
along with the explanatory note the designer wrote for that field, which the old generator
also threw away. So the writer is now shown:

    "branches": [{ "body": "...", "label": "..." }]

and told, in words, that it may leave it out entirely when a step has no decision point —
because empty is legitimate here and we did not want to overcorrect into badgering the
writer to fill things in that aren't in the source material.

Exactly one component on the whole estate is affected today. I measured that across all
three of the ways components can describe their fields, because measuring only the obvious
one would have under-reported. Every other page's instructions are unchanged, byte for
byte, and there is a test that proves it rather than an assurance that says so.

## Where it stands right now

The database half is live — that went in this morning and I verified it by reading the row
back, not by trusting the tool that applied it. The code half is committed but **does not
take effect until the next chassis release**; today's staged release does not contain it.
The two halves are deliberately safe in either order, and I proved that rather than
reasoning about it, because "it'll be fine" is exactly the sort of claim that turns out not
to be.

It has gone to the reviewer council; the verdict was still pending when I wrote this.

## One thing I got wrong, and how it was caught

The database change had a safety guard on it — a check that refuses to run if the edit
doesn't do exactly what I said it would. I got the arithmetic in that guard wrong. It
would not have damaged anything; it would have refused my own correct change, which is
worse than it sounds, because when a guard fires the natural reaction is to doubt the
change rather than the guard.

I caught it by rehearsing the whole thing twice before running it for real — once as
arithmetic, once as the actual database commands inside a transaction I threw away
afterwards. The rehearsal also proved the undo script puts everything back exactly as it
was. That practice is now written into the runbook, because the general lesson is one I'd
got away with before by luck: when you assert what an edit will change, you have to count
what it *removes* as well as what it adds.

## What is not done

This stops it happening again. **It does not rebuild the pages that are already stuck** —
that is deliberate, and it is two further pieces of work that remain open:

1. Nothing repairs a build that already failed this way. It fails, retries identically,
   and is eventually marked terminal. There is no path that says "regenerate that one field,
   here is the error".
2. Nothing tells anyone that an active page which other live pages link to has never been
   built. That silence is why these sat for weeks. On the evidence here it is the more
   valuable of the two.

The good news on the stuck pages is that most will heal themselves once the code ships: a
page whose item was marked `failed` gets picked up again automatically by the routine
sweep. The exception is a page branded "unresolved after 2 attempts" — that state is
deliberately kept open, which means it *blocks* the sweep from re-minting, so those need a
person to close them. I have told the portfolio-positioning lane, who are holding
advertise.co.uk ready to test, exactly which of those two situations they are in (the easy
one) and to wait for the release before firing anything.

## 2026-09-03, early afternoon — it works, and now we know what it doesn't fix

The release went out and the fix is working. I checked it properly this time rather than
taking the first encouraging sign.

Six pages have been written since the new code started running. All six were given the
corrected instructions. Four of them have gone all the way through to being built and
published, on three different sites. Every one of them stored the content in the right
shape. None stored it in the old broken shape.

I also went and looked at one of the finished pages on the live web — the advertising
regulation map on advertise.co.uk, which is one of the pages this bug had left stranded.
It loads, and the decision points that used to be an unusable blob of text are now seven
properly laid-out branches on the page. I checked a made-up address on the same site at
the same time to be sure the site wasn't just returning something for every URL, which
would have made the first result meaningless. It wasn't.

That page is worth singling out for another reason: nobody fired anything at it. It
repaired itself, on the routine sweep, exactly as predicted.

### The thing I nearly got wrong

My first count said three pages had failed *after* the fix went in. That would have meant
it was leaking. It hadn't failed at all.

The queue keeps the last error message on a job even after the job has moved on, and any
later touch of that job updates its timestamp — so an old failure resurfaces looking like
a new one. The giveaway was that two of the three jobs were marked *complete* while still
carrying a failure message. Counting the actual build attempts instead of the job records
gives the true answer: the last real failure was at 12:23 and there have been none since.

I have written that up as a trap for other people, because it will catch anyone checking
whether a fix worked, and it fails in the direction that makes you distrust a good fix.

### The reviewers approved it, and asked three fair questions

The review council approved the change. It raised four points, none serious, and three of
them were worth actually answering rather than waving through, so I did:

- *Does this fix one spot in a problem that exists elsewhere?* No. There is exactly one
  place in the system that builds these instructions, and it is the one fixed.
- *Does the count of affected components include the per-site copies people make?* Yes —
  and rather than just say so, I ran the count in a way that showed 27 such copies were in
  the population being searched. Only the one original component is affected.
- *You cited a test that wasn't visible in what you submitted.* The test exists; my
  submission was at fault for not showing it. Fair, and the same lesson as last round.

### The part that isn't finished, and it is bigger than I thought

The fix stops this happening again. It does not go back and rebuild what is already broken,
and I now know exactly how much that is.

Seventy-three pages' worth of work was hit by this. **Fifty-two of them — just over
seven in ten — cannot recover on their own, ever.** When a job fails twice it gets branded
"unresolved", and that brand deliberately keeps the job open, which has the side effect of
blocking the system from ever trying again. There are 251 of those branded records sitting
across four sites: remortgagecalculator, farmerinsurance, loanzy and cv1.

This does not improve with time. A related counter I was watching *is* decaying nicely —
one site dropped from six at-risk pages to two while I was working — but the branded
records are permanent until somebody clears them.

So the honest position is: the door is shut, and about three in ten of the damaged pages
will now heal themselves. The other seven in ten need someone to clear the brand.

**I have not done that, and I would like your view before I do.** It means changing 251
records across four sites that other people are working on. The reason I was waiting was
that we wanted a real page to build successfully first, before clearing anything — that has
now happened four times over, so the evidence side is settled. What is left is a judgement
call about touching other lanes' sites in bulk, and that is yours rather than mine.

Loanzy's three original stranded pages are still stranded, and they are the illustration of
what leaving it costs: they are linked from live pages, they have never been published, and
nothing in the system will ever try them again by itself.

## 2026-09-03, late afternoon — I was wrong about the stuck pages, and they fixed themselves

I need to correct what I wrote a few hours ago, because you made a decision based on it.

I said 52 pages were permanently stuck behind a brand that only a person could clear, and
asked whether you wanted them cleared. You said to do one site first and then the rest.

When I went to do that, I looked at the first site's records before touching them — which is
the point of doing one first — and there was nothing there. They had already cleared. So had
the second site's, all 19 of them. Both happened at 16:08 this afternoon, automatically,
with nobody involved.

**Twenty of the fifty-two repaired themselves while the document saying they never would was
sitting in the repository.**

### What I got wrong

The brand does not block anything for these particular records. I took a rule that is true
for one kind of job and applied it to a different kind without checking they were the same
kind. They were not. For the kind actually involved here, the system treats that brand as
"finished" and is free to try again whenever it likes.

There is also, it turns out, already an automatic repair for exactly this situation: when a
page that other pages were linking to finally gets built, the system goes back and closes
those link complaints by itself. That is precisely what happened. The fix made the pages
buildable, and the existing machinery did the rest.

So the honest position is better than the one I gave you: the repair is happening on its own,
at roughly twenty pages in three hours, and it needs no intervention.

### What is actually left

Two sites have not repaired — loanzy and farmerinsurance, 32 pages between them. But not
because anything is blocking them. It is that nothing has rebuilt their pages yet, so the
automatic repair has nothing to act on. Loanzy's three pages have not been touched since
August. What they need is a build, not record-clearing.

**Nothing was changed and nothing needs undoing.** I have not touched a single record.

### The thing I want to flag, because it is the useful part

Your choice to do one site first is the only reason this was caught. If we had done all four
at once I would have run the whole operation — pointless everywhere, and on two of the four
sites it would have been surgery on records that had already sorted themselves out.

That is worth remembering as a general point: doing the small version first was not just
about limiting damage. It was that it forced me to actually look at the thing before acting,
and looking is what showed the reasoning was wrong.

I have written the mistake up properly in the shared log of wrong calls, including the
one-column check that would have caught it in the first place.

## 2026-09-04, late morning — you asked me to trigger the builds; there were none left to trigger

Good news, and it is the third time this week the answer has been "it already sorted itself
out, go and look again before doing anything".

**The page this whole bug was about is live.** Loanzy's `/your-rights.html` — the one that was
active, linked from other live pages, and had never been published since 18 August — went
live at 4:36 this morning. I loaded it: it serves, it is 116KB, and the decision-point boxes
that were the whole problem render properly, five of them. A second loanzy page came back six
minutes later, and farmerinsurance's claims page at 1:18am. All three by themselves.

So I have not triggered any builds, because there is nothing left to build.

**Nothing has failed this way in about 23 hours**, which is the strongest evidence yet that
the fix is holding — that window covers the overnight run when most building happens.

### The queue rows that are still showing

There are still 175 alarming-looking records across the two sites. They are stale, not stuck.

They are all complaints of the form "this page links to something that was never published".
They all point at the two pages that have now been published. The check that clears them runs
once a day and last ran at 16:06 yesterday — **nine to twelve hours before those pages went
live**. So it was right when it ran, and it simply has not run since. It is due again around
16:06 today and should clear all 175 by itself.

I checked that properly rather than assuming, because I have been wrong twice on this bug by
predicting instead of measuring: the check does look at records in this state, it processes up
to 1,500 at a time rather than the 50 I first feared, and the test it applies now passes for
both pages. And there is a precedent from yesterday — two other sites in exactly this
position cleared themselves the same way.

### One thing that is genuinely still broken, and it is not mine

Loanzy's `/guides/index.html` has been sitting unpublished since 18 August. **It is not this
bug** — it has no content at all, so it never got as far as the stage that was broken. It is
a different fault on someone else's site, and it was hidden behind this bug's noise. I have
flagged it to the loanzy lane along with a farmerinsurance page in a similar state, and left
both alone.

### Lanes told

I have written to the loanzy lane, which owns loanzy.uk and is the closest thing
farmerinsurance.uk has to an owner — it has no lane of its own. I told them what happened,
what needs nothing from them, and the one page that is theirs to act on.

I also explicitly retracted yesterday's wrong claim in that note, so they do not pick it up
second-hand and start clearing records that do not need clearing.

## 2026-09-04, late afternoon — all of it cleared itself, and nobody touched anything

The daily tidy-up ran at 16:08 and closed every remaining record in about three minutes.

**The count is now zero.** Yesterday afternoon there were 251 records I described to you as
permanently stuck and asked whether to clear by hand. This morning there were 175. There are
now none, across all six sites. Two hundred and seventy-five closed in total.

Nothing failed in the way this bug describes for over a day, and **no person touched a single
record at any point**. The repair machinery the system already had was sufficient the whole
time. Every intervention I proposed would have been unnecessary — worth saying plainly, since
I asked you to decide on one.

### A near-miss this afternoon, which is the useful part

The fleet upgrade restarted things at 16:01, and the tidy-up was due at 16:06 — inside a
five-minute window where a scheduled job can be silently swallowed and simply never run.

I warned another session about that risk in some detail, because the failure leaves no trace.
They checked at about 16:06, saw nothing, and reported it as dropped. It was heading for
someone being asked to trigger it manually. In fact it ran at 16:08 — a couple of minutes
late, because the restart delayed it rather than eating it.

**My warning caused that.** I described the invisible failure vividly enough that an absence
looked like proof of it. I had also given them the correct waiting time in the same message —
"if it's still missing after 16:10, then it's dropped" — but the story travelled further than
the number attached to it.

That is worth remembering as a general point about warning people: the alarming half of a
warning spreads faster than the part that tells you how long to wait before believing it.

That session, to their credit, refused to make the change on my behalf when they thought it
was needed — they said it needed a human rather than one Claude session doing an action
another session's settings had blocked. That was exactly right, and it is why nothing hasty
happened while we were both wrong about what we were looking at.

### Where that leaves it

The bug's original problem is finished: fixed, live, reviewed, the damaged pages rebuilt and
serving, and the queue drained to nothing. What remains in the file are two separate gaps that
were never this defect — and one of them, the fact that nothing notices a page which live
pages link to and which has never been published, is the reason this went unseen for three
weeks.
