# Where we are — the deploy path that callers could choose (179)

Plain-prose log, append-only, newest at the bottom.

---

## 2026-08-04, early morning

This is the second bug this session took on. The first one (116, about link
checking) turned out to be waiting on a decision from you rather than on any code,
so I wrote up why and handed it back. This one was a real fix, and it is done bar
the roll.

**The problem, in one sentence.** When the platform publishes an image to a site,
it works out where the file should go from two facts — what the image is for, and
what it is called. Everything that later needs to find that image works the path out
the same way, from the same two facts. That agreement is the whole point: writer and
readers cannot disagree if neither is allowed its own opinion. But the deploy step
had a back door — a caller could pass in a path of its own, and that path won. A
file put there is one that nothing else in the system can find, because everything
else is still working the path out from the two facts.

**What made it worse than the ticket said, and this is the part worth knowing.** The
ticket had measured the risk and found it empty — nobody, anywhere in the system's
whole history, had ever passed one of these paths in. So it read as a door nobody
had opened. But when I looked at *how* the deploy step reads its inputs, it does not
just look where you would expect. It searches the entire bundle of data flowing
through that job, up to twenty levels deep, for anything called by that name. So a
stray value with the right name, sitting in some unrelated sub-result, would have
been picked up and used to redirect where the file was published — **without anyone
choosing it.** The measurement was correct and the conclusion drawn from it was too
comfortable: counting who had deliberately used the door tells you nothing about a
door that can blow open.

**What I did.** Took the back door out entirely, rather than putting a lock on it. If
someone deliberately asks for a custom path, the step now declines and says why,
before it downloads anything or writes anything to the site's repository — declining
politely, so the job finishes and the task is marked resolved rather than retrying
for ever against something that will never let it through. If a value only turns up
via that deep search, it is now ignored rather than refused: refusing would mean a
stray key somewhere could block perfectly good publishing across the whole fleet,
and that would be a worse bug than the one I was fixing.

**And I fixed the class, not just the case.** There is now a test that fails the
build if *anyone anywhere* in the codebase hand-builds one of these path objects
outside the single place that owns them. When I wrote it there was exactly one such
place in the entire codebase — the one I was deleting. So the door cannot be rebuilt
somewhere else by someone who has not read any of this.

**On being sure it works.** I deliberately broke my own fix six different ways to
check each test actually catches something — including moving the safety check to
the wrong place in the sequence, and re-wiring it to the over-eager search I just
described. Each break was caught by exactly one test and no others, which is what
tells you the tests are doing work rather than decorating.

Two things went wrong that are worth recording. Both times, an explanatory comment
I wrote broke one of my own tests — once because naming a function in a comment
above the safety check made the test think the check was in the wrong place, and
once because the check's own comment mentions the thing it deliberately does not
use. Both were the tests being right and me being careless, and both are now
written down where the next person will hit them.

**Where it stands.** The code and the configuration are committed, the database
change is applied and verified, and it has gone to the review council — that was
still running when I wrote this. The one thing left is for the fix to reach the live
system on the next build, and then to prove it there by actually trying a custom
path and watching it be declined. Until that happens the ticket stays open, because
the rule here is that a fix that is committed but not yet running is still
reproducible in production.

## 2026-08-04, later — it is live, proven, and closed. And I have to report a mistake.

The build you deployed carries the fix. I checked it on both running copies, and
because I had taken a "before" reading on the previous build, the change shows in both
directions: the new refusal message appears where it was absent, and the old
"using custom path" message is gone where it was present. That two-way check is
worth more than the usual one-way one.

Then I proved it actually behaves, rather than just being present in the binary. I
sent the deploy step two otherwise identical jobs, differing only in whether they
asked for a custom path, and both pointed at a deliberately non-existent image so
neither could write anything to a real site. The one asking for a custom path was
declined, politely and with a reason naming where the request came from. The one not
asking for it got further and failed later for an unrelated reason — which is the
more useful result, because it shows the new check is not simply refusing everything,
and that it happens early, before anything is written. Neither job left a file behind.

**The bug is closed and the file has moved to the closed folder.**

**Now the mistake, and it is a real one.** Throughout this work I justified deleting
the back door by saying nobody had ever used it — a count across three places in the
database, which I put in the review submission, the register, the database migration
and several commit messages. **That count was worthless.** The way I was searching the
data could never have found a match, because of a formatting detail in how the
database stores it. Every zero I reported was a property of my query, not of the
system.

I found it by accident: after the deploy I re-ran the same count as a routine check,
and it still said zero — minutes after I had myself deliberately created exactly the
thing it was supposed to count. A search that cannot find something I just made with
my own hands is not a search.

I re-did it properly. **The answer is the same** — nothing has ever legitimately used
that back door, so removing it was the right call. But I want to be straight with you
that the fix was right for a reason I had not actually established at the time, and
the review council approved it on evidence that could not have come out any other way.

The search pattern I used was not something I invented — it was recommended in the
bug ticket itself, and it has been copied into at least two other documents. I have
corrected it where it originated, written up the incident, and recorded the trap as a
standing warning, because it applies to any similar count anyone runs against this
database.

---

**2026-08-04, late evening (a fresh session picking up the handoff).** Three things
tonight, all executions of decisions you already made.

First: the vet data collection you restarted this evening worked immediately. Within
a minute of its first tick it was storing practice data again, and — this is the
point of the whole bug — every one of the first four records carries the web address
it actually came from, recorded by the code that fetched the page, not repeated by
the AI. That was the last thing bug 100 was waiting on, so it is now closed. It had
been open since late July, and the fix has been in place for a week; what was
missing was live traffic to prove it, and the restart provided it.

Second: the robot-hands wording fix you authorised. Good news — half of it had
already been fixed by someone else in passing: the catalogue page no longer makes
the "independently verified" claim at all. Only the how-it-works page still did. I
have removed the false phrase there (the surrounding text already describes,
honestly, what the site actually does — sources named and dated), and the page is
queued to republish. I will confirm the change on the live site before calling it
done.

Third: the header test on finetuning.uk you authorised. I picked a blog page that is
half-broken anyway (it has no stored content and is marked for rebuild), and its
current live copy conveniently carries the "before" evidence: a stamp naming the
switched-off header. I have queued a rebuild; if the fix works, the rebuilt page's
stamp will name the correct header from the library instead. Both queued jobs are
waiting on the shared build queue, which is busy but moving.
