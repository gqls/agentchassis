# Where we are — the chrome lock gate (bugs_open/069)

Plain-prose log, append-only, newest at the bottom.

## 2026-07-26, evening

**What the job was.** When someone edits a site's header, footer or `<head>` block in the admin
dashboard, the system marks that slot "locked" — meaning *automation, leave this alone*. The lock
worked for ordinary page sections (that was bug 058, fixed a few days ago). It did not work for the
site chrome: the four bits of code that rewrite a header or footer never looked at the lock at all.
So a human could correct the footer, lock it, and the next automatic re-render would quietly throw
their work away. That was bug 069, and it was deliberately left over from 058 to keep that change
small enough to review.

**What we found on the way.** Three things worth knowing.

First, the bug report was wrong about one of the writers. It said one of the two statements in the
re-render code only ever *creates* a row, so it could be ignored. It doesn't — it also repoints an
existing row at a *generic, off-the-shelf* template. That turned out to be the worst case of the
whole bug: it fires precisely on the slots that have no template assigned, which is a common state.
We now block it, and we tested exactly that case.

Second, we found a completely separate defect while auditing. If you roll a site back to an earlier
snapshot, the rollback deleted every lock on the site and put nothing back — and the "safety
snapshot" it takes first couldn't help, because it had never recorded locks either. Thirty-nine real
locks were exposed to that. You said to fix it in the same session, so we did: snapshots now record
locks, and a rollback carries the *current* locks across untouched. The rule we settled on is worth
stating in one line: **a rollback restores content; it never locks or unlocks anything.** Restoring
the old lock state would have quietly released anything locked since the snapshot — the same bug in
different clothes. That is filed and closed as bug 088.

Third, and this is the honest part: **nothing had actually broken yet.** No chrome slot on any site
is currently locked, and nobody has ever edited chrome through the dashboard. So this was a fix to a
trap, not to damage. That matters for how we proved it — a normal run would have looked fine whether
the fix worked or not, so we had to *cause* the failure deliberately and watch it be refused.

**How we proved it.** We built the new chassis image, checked the running pod really contained the
new code (not just that the tag looked right), then made three fake chrome slots on the dartsonline
test site — none of them a real slot, so nothing the live site serves could be touched. Two were
locked, one wasn't. We forced a re-render at all three. The locked ones came out byte-for-byte
identical, down to the timestamp; the unlocked one was rewritten properly (41 characters to 3,429),
which is the half people forget to check — it's easy to "fix" this kind of bug by never writing
anything. Two review items appeared in the queue, one per locked slot, flagged for a human with no
robot assigned to them. Then we deleted the test data and checked twice that nothing was left behind.

**One thing didn't go smoothly.** Our first attempt at that test sat in a queue for twenty minutes
behind other work, and while it waited another session restarted the chassis — which silently ate the
queued message. The tooling told us (correctly, at the time) "it's queued, don't re-send". It was
queued right up until the restart threw it away. We've written that down: "it's queued" needs a
second question — *has the thing that would process it restarted since I sent it?*

**Where it stands.** Both bugs are fixed, live, and closed. The code is in chassis v1.0.1170 and
re-verified on v1.0.1171. The database half of bug 088 was live the moment it was applied.

**One thing you may want to know about.** The advisory review council came back "revise", but not
because of an objection: ten of its twelve seats approved, and the verdict was decided by a seat that
crashed instead of answering. One reviewer did make a fair point we acted on — a piece of
copy-pasted logic had nothing keeping it in step with the original, so we added a test that fails if
the two ever drift, and then deliberately broke it to make sure the test actually fails. We have not
put a "reviewed by council" marker on the commits, because approval is what earns that and we didn't
get one.

**Still open, and named rather than hidden:** several other checks still scan chrome without looking
at locks, so a locked-but-stale header can still generate findings that its fixer will now politely
decline. It's bounded (the system gives up after two attempts) and it needs a wider change to fix
properly — one shared copy of the lock rule, which means resolving a code-structure knot first.
