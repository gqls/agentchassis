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
