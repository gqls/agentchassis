# SUMMARY — live-object declaration drift · 2026-08-26

*(A new file, not an edit of `SUMMARY_2026-08-22_…`. The series is the record.)*

## What we are trying to do

Some of how this platform behaves is not written in code. It sits in the database: a scheduled job's
query, a trigger, a rule about which values a column may hold, an agent's workflow. Our normal safety
net — a test that runs when we build — cannot see any of that, because there is no database when the
tests run.

So people bridged the gap by having tests read the *migration file* that originally created the
thing. That looks sensible and cannot work. A migration is a historical record: once it has been
applied we are forbidden to edit it, because we store a fingerprint of the file to prove what ran. So
the file is frozen for ever, while the live object it created keeps moving with every later change.
A test comparing today's code against a frozen file is asking a question whose answer can never
change. It passes whatever the database does.

The aim is to give the live objects a description that is *allowed* to change, tie the code to it,
and then have something check that description against the real database, every day.

## Where we came from

Filed on 22 August with seven such tests found and all seven objects measured. Nothing had actually
drifted — this was filed for the open door, not for damage. Two halves were built within days: a
declarations file the tests read, and a daily job that compares those declarations to the live
database.

Then it stalled, for an ordinary and slightly embarrassing reason. The daily job was deployed on the
23rd, and every document describing this work went on saying it was not deployed. The top of the bug
file announced that the main task was still ahead of us. So the finished half of the work sat
uncommitted in the shared workspace for two days, and two other people's work sat blocked behind it,
both of them having noticed, written a note explaining they were waiting, and correctly not touched
it.

## What we have done

Landed the blocked work, after checking every new declaration against the real database — including
one checked twice, once against a deliberately wrong question, to be sure it was capable of saying
no. That unblocked both waiting parties, and we absorbed one of their outstanding items so they would
not have to queue again behind the same counter.

Corrected the documents that had been telling everyone the opposite of the truth, and said plainly in
each what the staleness had cost.

Then we broke the checker on purpose, which is where the real finding came from. It turned out to
notice things being *removed* from the database and to be blind to things being *added* — and
"the live thing has quietly grown past what we wrote down" is precisely the problem this whole piece
of work exists to catch. So the checker had the very hole it was built to close, in its own subject
area.

That is now fixed. It counts as well as looks, proven by telling it to expect seven where the
database has eight and watching it object. And there is a new test that stops it coming back: anyone
adding a looking-only check must add the counting half or write down why their case does not need it.
Six such explanations exist, and two of them are honest admissions that a gap remains.

## Where we are now

Both halves are live. The daily job runs at seven each morning, and this morning it reported
inspecting ten things rather than the five it inspected on Monday — which is how we know the new work
genuinely reached production, rather than trusting a version number.

The bug stays open, deliberately, for three real leftovers. Two function-body checks still cannot see
an added value, and we have said so rather than implying coverage we do not have. A paragraph inside
the database describes itself in terms that stopped being true on the 19th, and our checker cannot
catch that one: the rule it sits beside is correct, and only the description lies. And one other
lane still owes a line that is theirs to add, in their own order.

We also lost a morning's review to something outside this work: the API ran out of credit, every
council review across the estate failed silently, and the failure looks exactly like "still waiting".
The reviews we need have to be resubmitted now that credit is back.

## Where we are going

Resubmit the review of this morning's change. Then the two honest gaps: the function-body checks, and
the self-describing paragraph in the database, which needs a change to a column another lane owns.

Two things worth carrying beyond this lane. First, a checker that has never been seen failing is not
yet known to work — and when we finally broke this one, one of our attempts *correctly* did nothing,
which briefly looked like a fault and was actually us asking the wrong question. Second, and less
comfortably: the day after writing a warning about a command that silently discards unsaved work, I
ran that command on my own unsaved work and lost a test. Writing the warning does not protect you
from the thing it warns about.
