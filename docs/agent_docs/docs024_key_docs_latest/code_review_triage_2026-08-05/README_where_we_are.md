# Where we are — the 2026-08-05 code review

Plain prose, append-only, newest at the bottom. This is the owner's log.

---

## 2026-08-05, afternoon — the triage got actioned, and two of the fifteen findings were wrong

This morning a session ran a code review over the working diff, got fifteen findings, and
sorted them by which lane owned them. It deliberately fixed nothing, on the reasoning that all
three lanes were still active and you should contribute into a lane rather than compete with
it. That was the right call at 11:02. By 11:03 it had stopped being true — two of the three
lanes closed their bugs and moved to `bugs_closed/`, and neither of their closing handoffs
mentions the code review at all. So ten of the fifteen findings were nobody's. That is what
made it reasonable to just fix them.

Here is what actually happened, in the order it mattered.

**The 140MB of binaries was the easy one, and it turned out we had done this before.** Four
compiled Go binaries were sitting untracked in the repo root, unmatched by any ignore rule. I
went to add them and found the `.gitignore` already has a section for exactly this, written
after a 93MB binary was committed by accident on 2026-07-20. So this was not a new risk, it
was the same accident lining up a second time. Four lines appended to the block that was
already there, with the reason it already gave: our builds use `git archive`, so a tracked
binary at the root gets extracted into *every* service's build context.

**Two findings were simply wrong, and one of them I nearly got wrong the same way.** The
morning's triage had already caught one false positive — a finding that asserted no live agent
sets a particular config key, when exactly one does. I found a second. A finding said we write
unboundedly to an error-log table with nothing cleaning it up. I measured it: the table had a
month of history and my search for cleanup code found nothing, so I wrote down "no reaper".
That was false. There is a reaper, it runs every hour, it had run minutes before I looked, and
it deletes anything older than thirty days. The "month of history" I read as evidence of
neglect *was the thirty-day boundary*, drawn to the minute — the oldest surviving row was
twenty-one minutes inside it.

What makes that worth telling you rather than quietly fixing: the number I relied on could
never have come out any other way. A thirty-day-old oldest row is what a working thirty-day
reaper produces, and it is also what no reaper at all produces. It does not distinguish
between them, and I did not stop to ask what a disconfirming result would have looked like. I
also searched only our Go code, when the cleanup is written in SQL and lives in a database
column. Both errors are in `WRONG_CALLS.md` now. The finding is recorded as a false positive
rather than filed as a bug — the same outcome the morning session reached for its one, by the
same route.

**The most valuable finding was a name collision, and it had already bitten us.** Yesterday's
work gave a config key called `require_sections_metadata` the meaning "refuse to save this
page". That exact spelling was already live, in the same package, meaning something much
milder — "warn me that a check could not run". Worse, one of our agents, `page-build-handler`,
carries both steps in a single definition, so the same word would have meant "warn" on one
line and "refuse the save" a few lines later. The trap is that the natural way to roll a
setting out across the fleet is a bulk update by key name, and that would have armed a hard
refusal on our highest-traffic page-save path.

The proof it is a real trap and not a theoretical one: I found a comment in our own codebase
that had *already* been confused by it, describing the two keys as though they were one thing
set on one step. They are not. I renamed the new key, which cost nothing because we had
deliberately shipped it switched on for nobody, and I corrected the comment in place rather
than deleting it, because a wrong comment that was caused by the collision is evidence for
fixing the collision.

**One finding asked for something impossible.** It said a database write here should call our
shared logging helper instead of writing its own SQL. It cannot: the shared helper lives in a
package that imports this one, so calling it would be a circular import. About twenty files in
that package each write their own copy for that reason. The finding's premise — that the
package's own convention forbids the duplication — was backwards; the duplication *is* the
convention, and it is forced. The half of that finding that was real (the row could not be
linked back to the run that produced it) is fixed.

**Two findings were assigned to the wrong lane, and the method is the reason.** The morning's
triage worked out ownership by asking git who last changed each *file*. Two lanes touched one
file twenty-six minutes apart yesterday, so that method gave the wrong answer for two
findings: it named the still-open lane when the lines in question belong to the closed one.
Asking git who wrote the specific *lines* settles it. Both are fixed rather than handed to a
lane that does not own them.

**What I did not do.** One finding is in a file another session has open work in right now —
their uncommitted changes are sitting in the working tree, and the very comment the finding is
about is part of what they have not committed yet. Touching it would have edited their work
mid-flight and swept it into my commit. I confirmed the finding is correct and worked out the
exact corrected line numbers, and left it for them. And one finding, F6, I could not action at
all: it is named in the triage's ownership table but never described anywhere, and the
original review output was not saved, so there is no record of what it actually claimed. I
have flagged that rather than guess.

**Where it stands.** Twelve findings resolved: nine fixed and committed, two recorded as false
positives with the measurements that refute them, one confirmed and handed back to its owner
with corrected line numbers. One cannot be recovered. All the code changes went to the review
council; the first verdict came back approved, and its one criticism was fair — I had claimed
two functions had no callers anywhere on the strength of a text search, which the reviewers
rightly said they could not verify. I redid it as a proof that could actually fail: rename the
functions, rebuild, and see whether anything breaks. Nothing did, which is what "no callers"
should mean.

## 2026-08-05, evening — the fixes are live, and one owed check now means something different

The chassis rolled to a new build and the code changes from this morning's work are live on
both replicas. I did not take the roll as proof of that — our own notes record that an image
carries no record of which commit built it, so a roll tells you something changed, not what. I
checked inside the running binary instead, on each replica separately, looking for two strings
that only exist because of these changes and one deliberately misspelled string that should not
exist at all. Both real strings present on both replicas, the misspelled one absent on both.
One of the two only exists after the very last code change, so its presence dates the image
past all of it.

One honest gap in that check, which I would rather write down than paper over: the strongest
version of this test also looks for a string the change *deleted*, and confirms it is gone.
This change set cannot supply one — most of what I changed was comments and internal names,
neither of which appear as text inside the binary, and the one piece of SQL I removed appears
in fifty-one other places. So I can prove the new code is in there; I cannot prove by the same
means that the old code is out.

The more interesting consequence is a check we already owed ourselves. When yesterday's work
shipped, it left a note to look back 24 hours later and count how often a particular
"this page just lost its structured content" warning had fired, with a clear rule attached: if
it fires for the re-render agent, the warning is wrongly designed and we must not roll it out
any further. One of today's fixes widened exactly the condition that warning tests. So that
check now asks a broader question than it did when it was written, and I have set out in the
notes what each possible answer will mean before seeing it — which is the part that stops a
result being read the way we would like to read it.

Right now the count is zero, which is the stated pass. It is also a weak pass and I have
recorded it as one: the warning has never fired in any version, so zero is what a correct
change and a broken one would both produce. The fleet is definitely running, but no page-save
traffic has crossed the window yet. The real read-out is due tomorrow evening.

---

**2026-08-06, morning.** Picked this up about thirteen hours before that read-out is due, to
get it ready rather than to take it early. Four things worth telling you, and one of them is a
mistake in yesterday's own instructions.

First, the system was rebuilt and restarted again overnight, for reasons nothing to do with
this work. That quietly broke yesterday's proof — not the fix, which is still in there and I
have re-checked it on both of the new machines, but the *evidence*. Yesterday's note named two
specific machines, and those no longer exist. Anyone reading it later would have been quoting
something that had vanished. So there is a general lesson written down now: after somebody
else's rebuild, you have to go and look again, because your proof has an expiry date you don't
control.

Second, and this is the mistake: yesterday I wrote down a way of checking whether the thing
we're measuring had actually happened at all, and that check cannot work. It counts entries in
an *error* log. When the operation we care about runs perfectly it writes no error, so the
answer is zero whether it ran a thousand times or never. I would have read a healthy system as
a silent one. I have replaced it with a check that looks at what the operation actually leaves
behind, and that one says clearly: it ran, on ten pages, and every single one kept its
structured content. That turns tonight's zero from "we learned nothing" into "nothing went
wrong, and here is the traffic that proves we'd have seen it if it had".

Third, a piece of luck worth noticing. The rule attached to tonight's check says: if this
warning fires for the *re-render* agent specifically, the design is wrong and we stop. It turns
out the re-render agent is the **only** one that has run since the change went live — twenty-three
times. So tonight's reading lands squarely on the case the rule was written about, which is
better than it might have been.

That third point came with a deadline I nearly missed. The records showing *which* agent ran
are themselves deleted after twenty-four hours — meaning they would have disappeared at almost
exactly the moment tonight's check was due to be read. I have taken and saved that evidence now.
The general form of it: if you schedule a check for a day later, and the supporting evidence
lives somewhere that clears itself out after a day, you have to grab the evidence early or you
will arrive to find the answer and no way to interpret it.

Fourth, my own slip, caught rather than shipped. Trying to confirm those twenty-three runs had
really completed, I asked the question in a way that could only ever have answered "no" — I
looked for a name in a list, and the field is actually a plain count, so the comparison was
meaningless and came back empty every time. I only noticed because it contradicted the other
measurement. It is a small thing, but it is the same shape as most of the errors in this
project's log of wrong calls: a check that cannot come out any other way isn't a check. Written
up where those go.

Where that leaves tonight: the reading itself is unchanged and still due at about 20:45. What is
different is that a zero will now mean something. I have also confirmed the second, smaller
outstanding check is settled — I did it by inspecting the configuration rather than the logs,
because the restart had wiped the logs down to half an hour, and configuration cannot expire the
way a log can. All six callers are correctly set up, so that warning genuinely cannot fire until
someone adds a new one.
