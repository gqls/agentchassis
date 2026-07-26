# Where we are — bugfix 006

Plain-prose running log. **Append to it; never rewrite or reorder it.** The owner writes here too;
if something here turns out to be wrong, add a dated correction below rather than editing it away.

Started 2026-07-26, later than it should have been — the workstream had been running since 07-20
without one.

---

## 2026-07-26 — all three problems are now answered, and the case is closed

Bug 006 was never really one bug. It was three unrelated things that happened to be noticed on the
same afternoon back in July, written up together, with a note at the top saying to hand each one to
a different person. That note turned out to be the most useful sentence in the file.

**The first one has fixed itself, and I can't prove how.** We had two copies of the machine that
publishes websites, so that if one dies the other carries on. One of them had been broken and
restarting in a loop for over three weeks — six thousand times — which meant we had no spare. Today
both copies are healthy, the broken one is gone, and its replacement has been running for five days
without a single restart. So the risk is genuinely gone.

What I can't tell you is *why*. The original problem was a misconfigured machine underneath, and
that machine is no longer in the cluster. Somebody may have fixed it, or the whole batch of
machines may simply have been replaced during normal maintenance. From where I sit those look
identical. I've written that down honestly rather than claiming a fix, and I've left a note saying
that if the same error ever appears again, it's this old problem coming back rather than a new one.

**The second one is fixed and working, with a tail we agreed to leave alone.** These are the
contact forms on the generated sites — for a long time, anyone filling one in was sending their
message into a void. That's fixed at the source: any new page that gets built now produces a
working form, and there's an automatic repair loop that fixes old ones when it next visits a site.

Today twelve sites have a contact form. Three of them are now working. Nine are still broken and
will fix themselves whenever the system next looks at them, which is what you decided yesterday —
no forcing it, no batch runs. The one genuinely new thing since then is oufe.com: it's a brand-new
site, and its form was *born* correct. That's the first proof we've had that the fix works when
building something from scratch, not just when repairing something old.

**The third one is what I actually built today.** This is the least visible of the three and it was
quietly wasting real money.

When the system decides a page needs work, it hands the job to a worker and marks the job "taken".
The worker does the work and reports back, and the job gets marked "done". Sometimes the worker
finishes the job but the "done" message never arrives — the worker's machine dies, or the reply
gets lost. The job then sits there looking untouched, and forty minutes later the system gives up
waiting, assumes nothing happened, and **does the whole job again from scratch**.

There was already a safety net for this: after fifteen minutes, before giving up, the system checks
whether the work looks done and marks it finished if so. The problem is that this check had to be
written out by hand, separately, for each *kind* of job — and only three of the eighteen kinds had
ever had one written. So for the other fifteen, the safety net simply wasn't there. Over the last
two weeks that's eighty-four jobs re-done unnecessarily. Eleven of them eventually gave up
altogether and were recorded as failures, despite the work having been finished correctly the first
time.

Writing fifteen more hand-written checks was the obvious move and the wrong one. It turned out the
system already keeps a record that answers the question for *every* kind of job at once: each
worker leaves behind a record of its own run, and that record carries the job's identity. So
instead of fifteen new checks there's now one, and it covers every kind of job — including kinds
nobody has invented yet.

I was careful about two things. First, the new check is deliberately no cleverer than the message
it's replacing: if the worker said it finished, we accept that, exactly as we would have if the
message had arrived normally. Being stricter would have meant continuing to redo finished work,
which is the whole problem. Second, there are three kinds of job where a separate quality check can
*refuse* to mark them done, and my check can't run that quality check — so those three are left out
and carry on the old way. There's an automated test that will break the build if anyone adds a
fourth of those and forgets to update the list.

**It's live now.** This one was a settings change rather than a code change, which means it took
effect immediately rather than waiting for the next software release.

**I tested it by breaking it on purpose.** I planted two fake jobs: one where the worker had
finished, one where the worker had failed. The live system correctly finished the first and
correctly left the second alone. I also deliberately introduced four different faults into the fix
itself and confirmed each one made the safety checks scream, because a check that has never failed
isn't really a check. Then I deleted the fake jobs.

**One thing I got wrong, and it's worth saying.** I had a very convincing theory about the cause —
that a cleanup routine was killing the supervisor while the worker was still busy. It explained
everything I'd seen. Before writing it down I measured how long workers actually take: at most
eight minutes, and the cleanup routine doesn't touch anything under thirty. The theory was simply
wrong. It cost me a minute to check and would have cost a lot more to unpick if it had reached a
handoff stated confidently.

**Where that leaves us.** All three parts are answered, so the case has moved to the closed pile,
with the two loose ends written at the top of it rather than buried: the nine contact forms still
waiting their turn, and the fact that today's fix catches the problem after it happens rather than
preventing it. Preventing it — working out why those "done" messages go missing at all — belongs to
a different piece of work that's already underway and already partly fixed.

I've also sent the change to the review council. That's advisory and doesn't hold anything up; if
it comes back with objections I'll deal with them as follow-ups, and I'll only claim it was
reviewed if it's actually approved.
