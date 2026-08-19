# Where we are — bug 315, "the database says the page was published and it wasn't"

Plain-prose running log for the owner. Append only, newest at the bottom.

## 2026-08-19, morning — starting on it

**The bug in one sentence.** When a page is rebuilt, the database writes a timestamp called
`deployed_at` that is supposed to mean "this page is now live" — but nothing in the system ever
checks whether the page actually reached the website, so the timestamp gets written either way.

The lane that found it (the one rebuilding the webdesign.co.uk tools) hit the real-world version:
a tool page was rebuilt four separate times, all four rebuilds reported success, the timestamp was
refreshed each time, and the public website carried on serving the *old* tool for about six hours.
Then it published itself with nobody doing anything. A second lane found the same shape on a
different site — a page marked active that has never existed on the web at all, with three
successful rebuilds behind it.

**Is it still real?** Yes. I re-measured everything this morning rather than trusting the file:
42 live pages across 14 sites have no content in them at all, and two of those are marked as
successfully deployed. Those numbers are unchanged from yesterday.

**Whose is it?** The lane that filed it says plainly, twice, that it is not theirs to fix — they
found it while doing something else. So nobody is working on it and I am not stepping on anyone.

## What I have found so far, and one thing I did not expect

I traced the whole path a page takes from "rebuilt" to "on the website". It goes: the platform
writes the new page, hands it to a small service that commits it to a shared GitHub repository, and
GitHub then copies the changed folders up to the storage bucket the public site is served from.
The phrase the platform's own documentation uses is **"commit is deploy"** — there is no separate
publishing step, which is why the timestamp has nothing to wait for.

Two things stand out.

**First, the timestamp is written in the wrong place — and in two cases, at the wrong time.** There
are five places in the system that stamp a page as deployed. Three of them stamp it just after
handing the page to the commit service, but they never look at what that service came back and
said — so if it came back saying "there was nothing to commit", the page is still marked deployed.
The other two stamp the page as deployed **before the commit has even been requested.** So this is
not a subtle race. There is no arrangement of these five workflows in which that timestamp could
honestly mean "this page is live".

**Second, and this is the one I did not expect: the machinery to do this properly was already built
and then never plugged in.** The database has a column on pages for a content fingerprint, and
another on page sections for the commit that deployed them. Both exist. Both are completely empty —
zero rows out of 786 and zero out of 1,775 respectively — and searching the entire codebase,
including the tests, finds *no code that writes to either of them*. Somebody designed exactly the
traceability this bug is asking for, and it was never wired up.

That changes what the fix is. It is not "invent a way to prove a page published". It is "start
writing the two things the schema already has room for, then compare them against what the website
is actually serving". That is a much cheaper piece of work, and it is what the house rules ask for
anyway — reuse what exists before building something new.

**A related thing I corrected while I was in there.** The platform's own concept register — the
document other parts of the system, including the automated code reviewers, read as authoritative —
states that commit references *are* recorded on pages and work items for traceability. That is
wrong in all three of its parts: there is no such column on pages, none on work items, and the one
column that does exist has never been written to. I have marked the correction in place, because a
reviewer reading that line today would conclude the traceability already exists and would push back
on a proposal to add it.

## One operational note that is not about this bug

At about 10:25 this morning the fleet's AI provider started refusing requests with "you have
reached your specified API usage limits". It knocked over my first diagnosis run. Before reporting
that as an outage I checked the history: the same message has appeared on five separate days over
the last month and the fleet carried on each time, and the system re-queued my run by itself within
two minutes. So it looks like an intermittent spend ceiling rather than the hard lockout the wording
suggests. Worth knowing about, not worth acting on from here.

## Where I am going next

I have asked for a full fix plan and I will bring it back with a recommendation rather than a menu.
The shape it will take: make the commit step report what it actually committed, write that down in
the columns that already exist, and then add a periodic check that compares what we believe we
published against what the website is really serving. The last of those is the piece that would
have caught this bug the first time it happened, six hours before anyone noticed.
