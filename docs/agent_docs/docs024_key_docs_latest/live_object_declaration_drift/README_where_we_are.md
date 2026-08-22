# Where we are — live-object declaration drift

Plain prose, append-only, newest at the bottom. The owner's document.

---

## 2026-08-22 — asked to fix bug 317; it was already fixed, and something else turned up

I was asked to fix bug 317. It isn't in the open pile any more — another session fixed it three
days ago and filed it away yesterday.

I checked that rather than believing it. There are three things that had to be true, and all
three are: the live database setting that the bug was about now has the right fourteen entries
in it, those fourteen are exactly the fourteen the code says they should be, and the automated
guard that keeps the two in step passes when run against the committed code. So 317 is
genuinely done, and I have not touched it.

While checking, I found something else, and it is worth explaining because it is a shape rather
than a single fault.

We keep some of the platform's behaviour in the database rather than in code — a scheduled job's
query, a database trigger, a rule about which values a column will accept. Because that
behaviour is not in the code, an ordinary test cannot see it. What we have done instead, in
several places, is write a test that reads **the migration file** — the file that was run once,
long ago, to put that behaviour into the database — and check the code agrees with the file.

The problem is that a migration file is a receipt, not a description. Our own rule, and a good
one, is that you must never edit a migration after it has been applied, because the system
records a fingerprint of the file and editing it would make that record a lie. So the file is
frozen at the moment it was written, while the live setting carries on being changed by every
later migration. The test therefore checks the code against a photograph, and passes whether or
not the photograph still resembles the thing.

The example I hit makes it vivid. The guard for 317 reads a migration file and pulls the list of
fourteen values out of it — but in that file the list only appears **inside a comment**. It has
to, because the migration edits the live setting by find-and-replace rather than writing the
whole thing out. So the specification our test checks against is a sentence somebody typed in a
comment. And the live setting has since been changed again by a further migration whose filename
the guard doesn't even look at.

Nothing is broken today. I checked: the file and the live setting still agree, in this case and
by luck rather than by machinery. What is missing is anything that would tell us if they stopped
agreeing. I looked for such a thing specifically and there is none — the migration system records
a fingerprint of the file and never looks at the live object, and nothing anywhere in our code
asks the database what its triggers actually contain.

It is not one guard, either. I counted seven places doing this, covering scheduled jobs, two sets
of database triggers, a column rule, and a workflow definition.

So this is the lane: work out the general fix rather than patching the one guard. I have sent it
to the diagnosis loop for an independent read, and asked our planning model for a design. My own
view before either comes back is that the fix cannot be a test, because tests must not need a
live database, and it cannot be a commit hook, because at commit time the migration hasn't been
applied yet — we already ruled on exactly that point in a different case a few weeks ago, and the
answer there was a daily check running against the real system. I expect we land somewhere near
that, but I would rather see the plan than assume it.

One honest note: I got something wrong on the way and it is written up. I counted the values in
the code two short of the live list and briefly thought the guard was failing. I had grepped for
one spelling of a function call when there are two. The lesson is small and general — I measured
my guess about the interface rather than the interface — and I have logged it.

## 2026-08-22, later — I went looking for damage and did not find any. That is the honest headline.

I said earlier that seven of our tests check the code against a photograph rather than against the
live system. The obvious next question is whether any of those photographs have stopped resembling
the thing, so I checked six of the seven directly against the live database — the scheduled job's
setting, two database triggers, two column rules, and two workflow definitions.

**All six still agree.** Nothing has drifted. Nothing is broken today.

I want to be plain about that rather than dress it up, because it would have been easy to. Each of
those six checks could have come out the other way, and one of them very nearly looked like it had
— I found that a third database trigger now exists on one table where the file our test reads
describes only two. But the part the test actually checks, the shared logic behind those triggers,
still matches.

So the case for doing something is not "this is broken". It is narrower and, I think, stronger.

Three of these seven tests **cannot fail** in the way they are written to fail. They check that a
migration file still contains a particular line — but we have a firm rule that a migration file must
never be edited once it has run, because the system stores a fingerprint of it. So the thing the
test is watching for is something our own rules have already made impossible. The test is green, and
would stay green if the live database were changed out from under it tomorrow. Two others quietly
pass if the file they want has merely been renamed.

There is also a good example already in our code of how this should look. One of the seven, the one
covering which values a couple of columns will accept, doesn't name a fixed file at all — it scans
every migration we have and takes the most recent one that sets the rule. That is much harder to
slip past, and it is the pattern I would want to generalise rather than invent something new. What
even that one lacks is the last step: asking the live database what it actually thinks.

The independent diagnosis run is still going, on its third pass. The design work is still coming
back. When both land I will file this properly and put it to the review council. I do not expect to
be told this is urgent, and I would not argue if someone said it should wait behind more pressing
work — but it is the kind of thing that is very cheap to fix now and expensive to discover later,
because the failure it permits is one where everything looks fine.

## 2026-08-22, evening — the independent check did not settle it, and then handed me the best finding anyway

Two things to report, and the second is the interesting one.

**First: the independent diagnosis run came back inconclusive, and I am not going to dress that up.**
It ran for about twenty-three minutes, went round five times, and stopped because it hit its own
iteration limit without reaching a conclusion. Its verdict field says "NOT CONFIRMED". That is not
the same as being told I am wrong — it did not refute anything — but it is certainly not support,
and the bug I have filed rests on my own checking, not on that run.

It failed for a reason that is my fault, and I have written it down. The tool takes an optional
"start by looking here" list, and I did not give it one; it even printed a warning saying it was
"dispatching blind", which I read past. Without that list it guesses where to look by searching for
code whose *names* match the words in my description — and the whole point of this bug is that the
problem lives in test files and old migration files, which have no such names. So it went looking at
a decommissioning routine, an idle monitor and an email-claim function, none of which have anything
to do with anything, and spent all five of its attempts there. The general lesson is worth more than
the wasted run: the kind of bug that most needs an outside opinion is exactly the kind this tool
cannot find on its own.

**Second: while doing that, it quoted something back at me that I had missed, and it is the sharpest
thing in the whole file.**

We have a firm rule here: don't trust the old migration file, go and read what the database actually
says now. It is good advice and I have been following it all day. But when the run quoted the live
setting back to me, I finally read the *comment* sitting directly above the list — rather than just
checking the list itself, which is what I had been doing.

That comment tells you the rule is "the twin of the verifier registrations", and it names a specific
test that supposedly keeps the two in step.

Both statements are wrong. The rule changed three days ago — it is now the combination of *two*
checks, not one — and the test it names has been deleted; it was replaced by a differently-named one
that had to be moved to another part of the codebase to do its job.

Here is why that matters more than it sounds. That sentence is the *original cause* of the bug I was
asked to look at this morning. Someone read it, believed it, and built a half-complete safeguard on
the strength of it. When that was fixed three days ago, the fix corrected the *list* and left the
*sentence* saying the old thing. So anyone who now does the correct thing — ignores the old file,
reads the live system — is handed the very same wrong explanation that caused the original problem,
along with a pointer to a test they will not be able to find.

It also spoils the neat version of my own story. I had been about to say "the files are stale, the
live system is the truth". That is not right either: the live system is carrying its own out-of-date
explanation of itself, and a checker that simply compared the live list against a written-down list
would sail straight past this, because the *list* is correct and it is the *prose* that lies. I have
written that into the bug as a constraint on any fix rather than quietly leaving it out, because it
is the kind of thing that makes a tidy solution look better than it is.

The design proposal is now with the review council. Nothing is built yet, and I would rather wait for
their verdict than start.
