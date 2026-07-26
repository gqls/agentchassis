# SUMMARY 2026-07-26 — bugfix 006 closed

The closing read-out for this workstream. A new file, not an edit of anything earlier: the series
is the record. Written to be read aloud.

---

## What we're trying to do

Close `bugs_open/006` — a case filed on 2026-07-16 that was never one bug. It bundled three
unrelated faults noticed during the same afternoon of idea.uk work, with an instruction at the top
to route each to its own chat: a publishing machine that had lost its redundancy, contact forms
across the whole fleet that silently discarded every message, and a job-scheduling loop that kept
re-doing work it had already finished.

The aim was not to make the file tidy. It was to get each of the three to the point where the
defect is genuinely gone from production, and to say plainly which parts are not.

## Where we've come from

The case sat open for ten days, and most of that time went on the contact forms — the part with
real users on the other end of it.

That part began with a **wrong cause**, and the wrongness is instructive. The file said forms were
posting to a dead `/contact` endpoint, and cited a `.sql` file as the emitter. Measured against the
live database, **no form on the fleet had ever posted to `/contact`.** The cited file was a backup
dump of a table, not source code. The real fault was that the form's destination came from content
the language model wrote, and no code anywhere set or checked it — so whatever the model happened
to emit is what shipped. That mistake survived four days and is logged in `WRONG_CALLS.md`.

From there it went the long way round, which was the right way: a fix at the render seam that
converts a dead form into a `mailto:` using the site's own address, and that **refuses to invent an
address** for sites which have none, on the grounds that a mailbox nobody reads makes a form look
repaired while still losing the message. A council review then found a genuine hole in that promise
— two render paths synthesise a fallback address before rendering, so the guard would have
fabricated exactly what it swore not to. That was fixed at the single point both paths pass
through. Then a detection check, then an automatic repair loop, then proof end-to-end on vonc.com
with no human touching it.

Two of the three were therefore in hand by 07-25. The third — the job loop — had been diagnosed
and left.

## What we've done

**Section A** was checked, not fixed. Both copies of the publishing runner are healthy, zero
restarts, no crash-looping pod anywhere in the namespace, and the replacement sits on a different
machine from the one that was broken. The single point of failure is gone. **How** it went is
`[INFERRED]` and the file says so: the faulty machine is no longer in the cluster, which makes
"somebody repaired it" and "the pool rolled and replaced it" indistinguishable from outside. A
reopen trigger is written in rather than a claim of victory.

**Section B** needed no new work, only re-measurement. Three of twelve forms now deliver; nine
await their next organic cycle, by the owner's ruling of 07-25. The new fact is `oufe.com` — a site
built *after* the fix, whose form was correct from birth. Everything before that had proved repair;
this is the first evidence the fix holds on the creation path.

**Section C** is what was built. The scheduler's safety net for a lost completion had to be
hand-written per kind of job, and covered three of eighteen — so for the other fifteen there was no
net at all. Eighty-four jobs re-run in fourteen days, eleven recorded as failures with the work
already done.

The fix is one branch instead of fifteen. Every dispatched job's identity is already carried into
its worker, so the worker's own run record answers "did this finish?" for every kind of job,
including kinds not yet invented. Two decisions shaped it. First, **the new check is deliberately
no stricter than the message it replaces** — a recovery that second-guesses the write it stands in
for keeps re-running finished work, which is the defect. Second, the three kinds of job whose
quality check can *refuse* completion are excluded, because that check is Go and this is SQL; a
test that reads the migration file itself will break the build if a fourth is added and the list is
not updated.

It went live immediately — settings, not code — and it was proved by breaking it: four faults
induced into the fix, each watched to make the guards fail with the right message, then two planted
jobs swept by the **production scheduler**, one finished correctly and one correctly left alone.

One theory was **refuted** along the way, and it was the confident one: that a cleanup routine was
killing the supervising loop while its worker was still busy. Measuring worker runtimes — eight
minutes at most, against a thirty-minute threshold — killed it in under a minute, before it reached
a handoff.

## Where we are now

`006` is **closed** and in `/bugs_closed/`, with both residuals stated at the top of the file
rather than buried in it:

- nine live contact forms still serving a dead action until their next organic re-render —
  a decision, not an oversight;
- and the **cause** of a lost completion write, which belongs to `bugs_open/003` and is already
  partly fixed and live there. Section C repairs the consequence, reliably, for every kind of job;
  it does not prevent the loss.

The council review of the C change is submitted and queued. It is advisory and blocks nothing; the
`Council-Reviewed:` trailer will only be claimed if it comes back approved.

## Where we're going

Nothing further is planned in this lane, and that is the intended end state — the file's own
instruction was to route each part to its own chat, and each part now has an owner or an ending.

Three things to watch rather than to do:

1. **A contact form still showing `#contact` after a discovery cycle has run on that site** is a
   real signal and should be chased. A form still broken because nothing has visited yet is not.
2. **The new completion marker** — `Auto-completed: handler orchestration completed after claim` —
   should start appearing against job kinds that had never recorded a single auto-completion. If it
   stays empty while timeouts continue, the branch is not reaching them and this closure was
   premature.
3. **`bugs_open/003`** is where the real prevention lives. Every job this new net catches is one
   that should not have needed catching.
