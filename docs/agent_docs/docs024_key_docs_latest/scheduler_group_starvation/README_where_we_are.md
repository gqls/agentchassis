# Where we are — scheduler maintenance tasks that stopped running

## 2026-07-21 (bugfix-048 thread)

Four of our background housekeeping jobs had quietly not run since **2 May** — 79
days. Nothing errored, nothing alerted, and the admin view showed them as healthy,
so nobody noticed until another thread tripped over it while chasing a different bug.
The four are the "maintenance" jobs: one that un-blocks stuck work items, one that
cleans up old error logs, one that retires work items left half-done for two days,
and one that archives finished work.

The cause turned out to be a subtle accounting bug in the scheduler. These four jobs
are set up to run one-at-a-time (they share a single "slot"). The first job in the
queue, when it runs and finds there's nothing for it to do — which is its *normal*
resting state — grabbed the slot, did nothing, and then never handed the slot back
or recorded that it had run. Because it never recorded a run, it stayed first in
the queue, so on the very next check it grabbed the slot again and did nothing again.
It could never lose its place by doing nothing, and doing nothing is exactly what
kept it there. The other three jobs just kept being told "the slot is busy". For 79
days.

Interestingly the jobs would briefly spring back to life whenever there genuinely
*was* something to un-block: the first job would then do real work, record it, and
drop to the back of the queue, letting the others have a turn — until things went
quiet again and it re-jammed. So it was intermittently self-healing, which is part
of why it hid so well.

The fix is two small changes to the scheduler, together: don't grab the slot until
a job has actually decided to do something, and always record "I ran" even when a
job looked and found nothing to do. That makes the jobs take turns properly and
makes "when did this last run" mean what everyone assumes it means.

Built and rolled the scheduler (it's a separate program from the main agents) at
version v1.0.1146 and checked it on the live system: all four jobs are running
again, the whole roster of 14 scheduled jobs is now healthy, and the small backlog
that had built up cleared itself. Done and verified.

One thing I deliberately did **not** do: add an alarm that would shout if a job goes
silent like this again. That belongs with a separate open item (044) about us having
the alerting capability but nothing wired into it. Worth doing, but it's a different
piece of work.
