# Where we are — chrome divergence guard (bug 226)

Append-only, newest at the bottom. Plain prose.

## 2026-08-08 — picked up, plan formed

We took on bug 226. The short version of the bug: the site header and footer
are stored once as finished HTML, and any time someone fixes something directly
in that stored HTML — the way the honesty note on oufe was added — the next
routine rebuild deletes the fix without telling anyone. It happened twice on
oufe and nobody noticed for eight days either time.

The plan has three parts. First, the database itself will keep a copy of the
old header or footer every time one is replaced with something different — so
nothing can ever be silently lost again, no matter who or what does the
replacing. Second, each time the platform writes a header or footer it will
leave a small stamp saying "the machine wrote these exact bytes"; if the bytes
on record stop matching the stamp, we know a person patched it by hand, and the
rebuild will say so loudly and file it for review instead of quietly steaming
over it. Third, the rebuild still goes ahead — locking things down is a
different, existing feature — this is about never losing work and never being
silent about it.

One correction to the bug as filed: it suggested re-running the old render to
compare — that isn't possible with what we store (we keep fingerprints of the
ingredients, not the ingredients). The stamp-the-bytes approach gets the same
answer more cheaply.

Timing matters: a separate fix (bug 117) will trigger a big wave of chrome
rebuilds on the next release. Our database half goes live immediately when
applied, before that wave — so the wave becomes the first thing the new safety
net catches rather than the last thing that slips through it.

Next: council review of the plan, then the code.

## 2026-08-08, later — safety net live, council asked for changes, changes made

The database half is done and live: from this evening, nothing can replace a
stored header or footer with different content without the old version being
kept. We proved it on a real row before trusting it — patched one, watched the
copy appear, put it back, cleaned up.

The reviewer council looked at the whole plan and said "revise". Some of their
worries turned out to be wrong once we measured — for instance, the fear that
only the first site to lose content would ever get a review ticket isn't true,
because tickets are already filed per site. But three worries were right and
we fixed all three: a second loss on the same slot could have been mistaken
for a duplicate of the first and dropped (each loss now gets its own ticket);
a ticket could have been filed even when a protective lock had actually
stopped the rebuild (the ticket now only files when something was really
replaced); and our promise to deal with the same problem on ordinary page
sections "later" wasn't concrete (it is now bug 229, written up properly with
its own evidence).

We resubmitted with the fixes and the measurements. The code is committed and
will be in the next release; the bug stays open until we've seen it working on
the running system, and the checklist for that is written into the bug file.
