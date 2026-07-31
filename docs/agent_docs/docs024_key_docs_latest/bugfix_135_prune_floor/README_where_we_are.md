# Where we are — the code index could delete itself, and nothing would have said so

Append-only, newest at the bottom, plain prose.

---

**2026-07-31, afternoon.**

We keep an index of the codebase — every function, method, struct and interface in
the repo, about 5,000 of them, with the source text of each one. It is what lets a
reviewing agent answer "does this code do X" without reading the whole tree, and
several of our review and diagnosis loops lean on it.

The way it gets refreshed is: fetch the repo, parse it, write down everything you
found, then **delete everything you did not find**. That last step is right — code
gets deleted, and the index has to notice. The problem is that nothing checked
whether the run had actually looked at the whole repo before it started deleting.

If the download arrived truncated, or a directory had moved, or the parser returned
a short answer, the run would not fail. It would quietly write down the few hundred
things it had seen, then delete the four thousand it had not. And the index would
not look broken afterwards — it would look *small*, and it would answer questions
about the missing code with a confident "there is no such thing here". We made that
answer *more* confident earlier this month, when we fixed a related complaint about
empty results being ambiguous. So the lie got stronger while the hole stayed open.

Nobody has been bitten by this yet. It was found by one of our own review seats
reading the code during a review of something else entirely, and filed as a
latent problem before it fired. The same pattern has bitten us twice before
elsewhere, both times destroying content on a live site.

**What we changed.** The delete now has to earn the right to run. Before deleting
anything, the run compares what it just confirmed against what is already stored —
separately for each kind of symbol, and separately again for the count of source
files it saw. If any of those has dropped by more than half, the delete is refused
outright, nothing is removed, and the refusal is written somewhere that survives
(our logs only keep about two days, which is less time than it would take anyone
to notice a thin index).

Two details worth saying out loud, because they are the difference between a guard
and a guard-shaped inconvenience:

- **The threshold can be lowered or switched off.** A big refactor really can delete
  half of something, and that is a legitimate event. The refusal message says how
  to override it. A guard with no exit gets deleted by the first person it blocks.
- **When we refuse to delete, we keep stale rows on purpose** — and that means the
  index is briefly part-old and part-new. Our freshness banner reads one row to
  decide "the index describes commit X", so it would have announced the new commit
  for a half-old index. So the read side now says plainly when the index spans more
  than one commit. Trading a loud catastrophe for a quiet inaccuracy would not have
  been a fix.

**A thing that went wrong today, which is worth recording.** Our test suite for
this whole area would not compile, and I concluded it had been broken for two days
by a commit on Tuesday. That was wrong. Another session is working in the same tree
right now, mid-refactor, and its half-finished files were sitting in front of me
looking exactly like committed history. The give-away was that fixing the first
compile error immediately produced a second one — two unrelated breakages in one
package in five minutes is a sign that you are looking at somebody's desk, not at
the record. I undid my "fix" (it would have broken the shared branch had I
committed it), rebuilt a clean copy of the branch with only my own files added, and
everything passed. That is also the only way to make a truthful claim about your own
work on a tree this busy.

**Where it stands.** The change is written, tested, registered, and has gone to the
review council. It is committed, so it ships with the next image. The one thing
still outstanding is the thing that actually matters: **watching the refusal
happen.** A run over a healthy repo tells you nothing — the guard is supposed to be
invisible then. So the last step is to deliberately make the index look
half-missing and confirm the delete refuses, then clean up and confirm a normal run
still prunes. Until that has happened I would describe this as shipped and
unproven, not done.

