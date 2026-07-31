# Where we are — making a refused plan recoverable (bugs_open/099)

Plain-prose running log. Append only, newest at the bottom.

---

## 2026-07-30, afternoon — what this is and why I picked it

I was asked to find the next open bug that no other session is working on, and fix
it properly rather than locally. There are 63 files in `bugs_open`. The tree is busy
— sixty-odd commits today alone — so "is anyone on this?" is a real question and the
tool we have for it (`who-owns.py`) is too cautious to answer it: it says "owned or
recently active" for nearly everything, because it counts any passing mention in an
active workstream's notes. What actually separated the live cases from the parked
ones was simply asking git when each bug's own file was last touched. That gave a
clean break: a cluster last edited on the 27th and 28th, and everything from the
29th onward still moving.

I picked **099**. Its file says, in its own words, that the durable fix "remains
**not done**". Nobody had touched it for three days.

## The bug, in ordinary terms

There is a step in the fix/feature loop whose job is to check the shape of a plan
before anything acts on it — the right number of stages, no contradictory edits, that
sort of thing — and then save it. If the plan fails one of those checks, the step
fails, and the whole run is thrown away.

The trouble is what gets thrown away. The plan is produced by a language model that
has just spent real money thinking about the problem, and the *design* is often
perfectly good. It falls over on a bookkeeping rule that the instructions it was
given never mentioned. The specific one that started this: if the same file appears
in two edits of a single stage, the plan is rejected. Nothing in the designer's brief
said so. So a sensible decomposition — "add the helper", "add the thing that calls
it", both in one file — produces a plan that is coherent, inside every stated limit,
and unsaveable. It cannot know.

And it fails quietly. The run reports as *completed*. The error column in the
database is empty. The actual reason is buried in a sub-field nobody reads. So a
dashboard shows a clean run, and a good design is gone.

Somebody already fixed this once, in July, by adding that one rule to the designer's
instructions. That worked, and it is still working. But there are about a dozen rules
in that validator, and it gains more over time. Fixing them one at a time means
writing the same fix again for every rule, in every agent that uses this step, and it
comes undone the moment the validator changes. The bug file itself flags this and
names the real fix.

## What I built

Instead of teaching the model each rule, I made the refusal recoverable. When a plan
fails a shape check now, three things happen:

1. The rejected plan and the exact complaints are **written down** — durably, so a
   design the loop gave up on can be read back by a person. Today it evaporates.
2. The step hands the complaints **back to the designer** with a prompt that says, in
   effect: your plan was not reviewed and not rejected on its merits, it failed a
   structural check before anyone saw it, here is precisely what to fix, change
   nothing else.
3. It does that a **bounded** number of times — once, by default — and then fails
   exactly as it does today.

The important part is that this does not care *which* rule was broken. It works for
the dozen that exist and for every rule added later, with no change to any prompt.
That is the difference between fixing the case and fixing the class.

Two things I was careful about. The gate is not loosened: an invalid plan is still
never saved, and nothing reaches a reviewer that has not passed the check. And a plan
that arrived **cut off** — because the model ran out of output budget mid-sentence —
does not get retried. That is a different fault with its own history here, and
looping on it would just burn money.

## The bit where the bug file was wrong, and I nearly followed it

The bug file's recommendation is to feed the problem back into a step called
`repropose`, which already exists. I started doing that, then read what `repropose`
actually says to the model. Its prompt is written for a *council review* — "a council
reviewed your previous plan and asked for revision", followed by the reviewers'
comments and their verification queries.

But the shape check runs **before** any council. So on the first failure there are no
reviews, no query results, and nothing to show. The prompt would have rendered three
empty sections and told the model a committee had objected when no committee had seen
it. I built a separate repair step instead, and wrote the correction into the bug file
rather than quietly designing around it.

## A mistake of my own, caught by luck of habit

I wrote the new prompt to include the spec for context, using a field called
`spec_row.body`. There is no such field — it is `summary` and `spec_text`. A wrong
field name here does not error. It renders as **nothing**, and the step runs happily
with an empty context section, looking fine. I only caught it because I went to check
the name before applying the change rather than after. That is in `WRONG_CALLS.md`,
because the lesson is not "check field names", it is that this whole class of mistake
is invisible by construction.

Two other traps turned up while building, and both are now written into the
fleet-wide landmine file: a database column that only accepts five specific values
(so inventing a sixth compiles fine and blows up later, on the one code path nobody
tests first), and a counting query that silently returns zero for ever if one field
happens to be empty — which would have turned my bounded retry into an infinite one,
with no error to show for it.

## Where it stands right now

The code is written, 11 tests pass, and the database half is written and dry-run
against the live system. Nothing is applied yet — the code has to be in a running
image before the config change means anything, and that ordering is deliberate.

Next: put it through the review council, build and roll the image, apply the config,
and then prove it on the **failing** case rather than a happy one. A green run proves
nothing here; the whole bug lives on the branch where the plan is bad.

---

## 2026-07-31, afternoon — it works, and the run that failed was the useful one

It's live. Not because I rolled anything — I deliberately didn't, because a rebuild
kills any review that's mid-flight and I had one running — but because my change was
already committed, and this repo builds from whatever is committed. So somebody else's
rebuild carried it out to the fleet without either of us doing anything. I checked it
had really landed by looking inside the running program on both machines rather than
trusting the version number, and it had.

Then I broke a plan on purpose to watch the new machinery catch it. It did, exactly as
designed: the broken plan was refused, the full design — about ten thousand characters
of it — was written down where a person can read it back, an operator-visible record
was filed, and the run was handed to the new repair step. Before today all of that
would have been a silent deletion.

And then the run died anyway, which is the most useful thing that happened.

The repair produced a corrected plan and got one field's *shape* wrong — it wrote a
list where a piece of text was expected. I had decided earlier that this kind of
mistake should be fatal, lumping it in with "the model ran out of room and stopped
mid-sentence". Those are not the same thing at all. One is a half-finished answer that
it would be daft to retry; the other is a complete answer in the wrong format, which
is about the most mechanically fixable error there is. So the loop saved the design,
recorded it, handed it back — and then threw the run away for a reason it should have
been able to fix.

That boundary was my own assumption, which is exactly why none of my fifteen tests
could have caught it. A test only checks the line you drew. It took one real run
against the real system to show that the single most likely repair failure was the one
case I'd ruled out. It's fixed now, and waiting on the next rebuild.

Something I want to flag, because it's a pattern rather than an incident: **three
separate times today a check would have told me everything was fine for a reason that
had nothing to do with what I was checking.** The bug report's own instructions for
verifying a fix would have passed without the new code running at all. The first
setting I chose to trigger the test turned out not to trigger anything, and I only
found that by going and reading what the system actually produces instead of assuming.
And the string I used to prove the deployment had landed came back zero — not because
the deployment failed, but because my own edit had removed that exact string. I'd have
gone hunting a build problem that didn't exist if I hadn't checked a second string at
the same time.

None of those were caught by being careful. They were caught by measuring the specific
thing rather than the thing next to it, which is slower and is the only method that
works.

One more, on how this place runs: I had two jobs going and another session rebuilt the
fleet underneath them. Both died instantly, and neither reported an error — they just
sat there looking busy with a timestamp that had stopped moving. I'd deliberately
avoided rebuilding, to protect exactly those jobs, and it made no difference, because
the rebuild that kills your work doesn't have to be yours. There's nothing to fix
there, it's just how a shared machine works, but it's worth knowing before you plan
around a long-running job.

I'm not closing the bug yet. The mechanism is live and I've watched every part of it
work. But I haven't yet seen a repair go all the way through and save a design in
production, and "the plumbing works" isn't the same claim as "it saved something".
That needs the next rebuild and one more deliberate break.
