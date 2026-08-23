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
