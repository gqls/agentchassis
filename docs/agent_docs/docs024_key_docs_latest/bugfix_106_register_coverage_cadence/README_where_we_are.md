# Where we are — bug 106, the register's blind spot

Append-only. Newest at the bottom. Plain prose.

---

**2026-07-28.**

We keep an index of everything the platform can already do — the concept register.
Sessions are told to check it before concluding that some capability does not
exist, so that we do not rebuild things we already have.

It stopped being updated on 13 July. Since then about two-thirds of the work we
have done is missing from it, and three whole subsystems were discovered missing
only because somebody happened to be working right next to the hole. That is the
dangerous shape: an index with a hole in it does not say "I don't know" — it reads
as "that doesn't exist". In one case in late July a session concluded exactly that
and was about to build something we already had.

The thread that filed this bug had already fixed most of it within hours. Every
file in the register now carries a "I stopped looking on this date" stamp, so
nobody can mistake it for complete. And they built a detector that lists what the
register has never heard of.

What was missing was that the detector only ran when somebody remembered to run
it — which is the same "found it by chance" problem the bug was about, just moved
one step back. Their own note said so, which I thought was an unusually honest
piece of writing.

So: the detector now runs by itself, at the moment a gap is created. When a commit
adds a new area of work the register has never heard of, it says so, to the person
who just created it, who can close the gap in about ten seconds. I preferred that
to a nightly job — a nightly job tells you a week later, and it tells nobody in
particular.

**One thing I checked rather than assumed.** The house rule here is that a new
automatic check has to prove how often it will fire before it ships, because a
check that cries wolf gets ignored within a week. Mine fires four times in the last
fifteen hundred commits — which is *quieter* than any of the existing checks, and
that could equally mean it is precise or that it is broken. Those look identical
from the number alone. So I looked at all four: every one is a real gap, and two of
them are the exact two the detector itself found by hand back on the 27th. Then I
deliberately created fake gaps to confirm it actually fires, and confirmed both of
the ways to silence it work.

**What I did not fix, and it is worth someone's time.** The register can be
complete but *wrong*. Twice today I was misled by an entry that was accurate when
written and stale by the time I read it — one of them contributed directly to the
bug I fixed earlier this session going unnoticed. Coverage and accuracy are
different problems, and only coverage now has a watcher. I have written that down
rather than quietly widening this fix, because a coverage check that starts
auditing accuracy becomes slow, noisy, and ignored.
