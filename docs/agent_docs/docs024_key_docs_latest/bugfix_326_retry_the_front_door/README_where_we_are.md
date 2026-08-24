# Where we are — bugs_open/326, the front door

Plain prose, append-only, newest at the bottom.

## 2026-08-23 — the bug was real, but not for the reason anyone thought

The complaint was simple: submit a domain, watch the build fail halfway, submit it
again — and you get a green tick and no build. That much is true, and it is the
customer path, so it matters.

The reason given in the bug file was that the system refuses to file the same piece
of work twice, whatever state the old one is in. That turns out to be wrong, and I
can show it: the database index that does the refusing explicitly ignores finished
work. A completed job cannot block a new one. So the thing everyone was about to fix
was innocent.

What actually happened is a different safety mechanism sitting just above it. If the
same piece of work finished less than three hours ago, the system quietly declines to
file it again — and "quietly" is the whole problem. It returned exactly the same
answer it gives when the work is genuinely already running, so the submitter reported
success. The build in question was re-submitted two hours and twenty-eight minutes
after the first one started. Twenty-eight minutes later and it would have worked.

That also means the recovery procedure in the old handoff — renaming seventy-eight
rows by hand in the database — was almost certainly unnecessary. It was surgery for a
three-hour timer. The lane that wrote it has accepted that and corrected its own
notes.

There is a second, worse version of the same mechanism. If a piece of work has
finished twice in a week, the third attempt gets filed into a status that nothing
ever picks up. It looks, to anyone reading a dashboard, like somebody made a
decision. Six hundred and thirty-five of those exist right now.

The interesting part is that none of this is new or undocumented. We hit it in July,
wrote it up, fixed it, and left ourselves a switch to turn it off for the kind of
work where a repeat is normal — asking for the next stage of a build, asking for a
page to be re-rendered. The July fix reached the places that lane happened to touch.
Nobody counted how many other places needed it. Nineteen out of twenty-one still
did not have it, including every step of the customer build pipeline.

So I have not fixed the five steps in front of me and called it done. Three changes:
the mechanism now *delays* work instead of destroying it, so even a step nobody has
classified is only slowed down; the result now says which of the two things happened,
so nobody can misread it again; and there is a check that lists every step nobody has
classified, so the next person does not have to remember.

The five build steps are also classified properly, and that part is live already —
it is a configuration change, so it took effect on apply rather than waiting for a
release.

One thing I want to flag because it is uncomfortable: I made the same class of
mistake three times today while investigating this. First I ran a query that could
only ever have given me the answer I expected — it asked about the present tense of
something that happened in the past. Then I counted the affected steps with a search
that only looked one level deep and missed nearly half of them, in the same session
where I wrote a tool whose documentation warns about exactly that. Then I checked a
file for a leftover entry with the error output switched off, from the wrong
directory, so "no matches" actually meant "no such file". All three are written
down: the first and third in the fleet-wide log of wrong calls, the second as a
correction in this project's notes. The second was only caught because the tool I
built disagreed with my own count, which is a decent argument for building the tool.

## 2026-08-23, later — what is still outstanding

The code half is committed but not live: it needs an image build and a fleet roll,
and one final configuration change has to wait until after that roll because
applying it early would take the front door down completely. That file says so in
large letters at the top.

The review council has the change and I am waiting on its verdict.

There is also a live test running that I did not set up: another session started a
real greenfield build of a new domain this afternoon, and has agreed to capture the
evidence if it needs to re-submit — rather than doing what the old handoff told
people to do. Either outcome is useful. If the build fails, we get the three-hour
mechanism caught in the act on a live build, which is evidence this bug has never
had. If it succeeds, they will re-submit afterwards, past the three hours, which
demonstrates the "can never be retried" claim in the bug's title is simply false.

Two things I have deliberately left alone, and I want to be explicit rather than
have them look like oversights. The six hundred and thirty-five dead records are not
being cleaned up here: re-activating them all at once would fire hundreds of page
rebuilds simultaneously, and the question of what to do with them is already an open
decision of yours from earlier in the month. And the other sixteen unclassified
steps belong to other people's work; the new check names them, and those lanes
should make their own call rather than have me guess from outside.

## 2026-08-23, evening — the review council said no to half of it, and it was right

I put the change through the review council and it came back REJECTED. Eleven of the
fourteen reviewers approved, two raised points, and one — the guardian seat, which
can veto on its own — refused it. Its reasoning was better than mine and I want to
record that plainly rather than dress it up.

Its argument: the customer's problem is fixed by the configuration change alone. The
larger change I bundled with it — making the mechanism delay work instead of
destroying it, everywhere, for everyone — is a separate decision about how the whole
system should behave, and I had attached it to an urgent bug fix, where it would get
waved through on the urgency of something else. It called that "an architecture change
dressed as a point fix", which is exactly what it was. I had the evidence in front of
me and I bundled it because it was there.

So: the configuration fix is applied and live. Re-submitting a domain after a
finished build now queues work instead of quietly doing nothing. I checked that at
the artefact rather than trusting the migration's own report — the new audit went
from nineteen unclassified steps to fourteen, and none of the remaining fourteen is
in the customer build path. The protection against two people submitting the same
domain at once is untouched, and it is enforced by the database rather than by
configuration, so it cannot be undone by an edit.

The larger change is written, tested, and *not* committed. On this repository,
committing is shipping — any other session's next build picks up whatever is on the
shared branch — so committing something the council refused would have been shipping
it by the back door. Instead it is preserved as a patch file next to a written-up
proposal that lays out three options with their costs, including the one I would pick
and why that is a view rather than a finding. Someone else decides.

I also took the change back out of the shared working files. That was the fiddly part:
another session has unfinished work in the same file, and the careless way to undo
mine would have destroyed theirs. I removed exactly my own lines and then checked
theirs was still there, character for character.

Two of the reviewers' smaller points were real defects and are fixed. Both migrations
could have taken a corrupted backup if anyone ran them twice, and both would have
happily edited the wrong row for any agent that has two active configurations — four
agents on the system do, though none of the five I touched. Both now refuse rather
than guess.

## What is still open, in one place

- **Fixed and live:** the customer path. Re-submission works.
- **Written, refused, waiting on a decision:** the framework-wide protection, so that
  a step nobody has classified cannot have its work silently destroyed. Fourteen
  configuration steps and thirty-six places in the code are still exposed.
- **Waiting on a release:** one further configuration change that makes the front door
  say so out loud when it genuinely cannot start a build. It is deliberately held
  back, because applying it before the code ships would stop the front door working
  entirely. The file says so at the top in large letters.
- **Not mine, flagged to their owners:** fourteen unclassified steps in other people's
  areas, now listed by name by the new check.

## 2026-08-23, 19:23Z — it works, on a real build, and we watched it happen

The proof arrived by accident, which is the best way for it to arrive.

A new domain was being built from scratch this evening as an unrelated test. That build died
partway through on a different fault entirely. So the person watching did the natural thing —
submitted the domain again — and this is precisely the situation the bug was about: a failed
build, a re-submission, and just over two hours since the first attempt. That is inside the
three-hour window that would have silently swallowed it this morning.

The work was queued. A new record, a new identity, and when I checked again a few minutes later
it had already been picked up and was running. Two of us checked, separately, and we agreed in
writing beforehand what each possible outcome would mean — which mattered, because I had already
managed to invalidate their original prediction without noticing.

I have also written down the two things that could have made this a false result and did not:
the check we thought would matter turned out to be unused, and the test came within about forty
seconds of colliding with a leftover item that would have blocked it for an entirely different
reason. Following my own earlier instruction, it would have collided. Their care caught that,
not mine.

So the customer path is fixed and demonstrated rather than merely believed. What is still open
is everything outside that path, which is what the refused proposal is about.

## 2026-08-24, afternoon — your ruling is carried out, and the council was faced honestly

You ruled "D and E now, with the census running alongside", and both halves are in.

E first, because it was independent: the eight places in the code that ask for the next stage
of work — re-adopting a site, re-running the design stage, re-requesting an image, re-seeding a
build — now say so explicitly, so the churn brake leaves their repeats alone. I read every
candidate before touching it and refused three, including one whose own comments had already
decided the opposite and one that shares its identity with a fault detector. The refusals are
written down next to the changes so nobody later "completes" the list.

D was the delicate one, because this council rejected its wider form yesterday and the rejection
was right. So it went back through the council as its own submission, opening with the veto
rather than burying it, answering each of its three grounds with measurements rather than
argument. The change itself is narrow: the three-hour arm that used to destroy a repeated
request now writes it down with a "not before" time — waiting out exactly the remainder of the
same three hours — while the two-strike arm is left alone on purpose, because a third of what it
stops is a fixer that keeps claiming success without fixing, and hurrying that along would help
nobody. There is an off switch, shipped on, that restores the old behaviour exactly, and it is
itself tested.

One deliberate courtesy that paid for itself twice: another session was mid-flight in the same
function, so I asked rather than raced, they landed first, and their work and mine compose
cleanly — their parked rows can never even reach my new code, by their design and my check.

None of this is live yet. Code changes wait for the next build and roll; the five build-chain
declarations from yesterday are config and have been live since they were applied. The one
standing warning is unchanged: the held migration 573 must not be applied — the code it needs
was part of the un-ruled remainder, and applying it early would stop the front door entirely.
