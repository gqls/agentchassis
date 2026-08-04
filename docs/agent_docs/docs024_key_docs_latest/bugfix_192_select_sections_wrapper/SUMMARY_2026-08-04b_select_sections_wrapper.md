# SUMMARY 2026-08-04b — `bugs_closed/192`: closed, fixed at source, live

New file, not an edit of `SUMMARY_2026-08-04_…`. That one ended "still open, deliberately";
this one is the closing read-out, and the difference between them is the record.

---

## What we're trying to do

Stop every page build in the fleet failing, and fix it at a level that closes the *class* —
because the thing that broke is a pattern the platform recommends in writing and will use
again.

## Where we've come from

The morning summary ended with the outage over but the ticket deliberately open: the
config workaround was live, the real fix was committed and inert, and the bar here is
*fixed **and** live*. A chassis roll has since happened.

## What we've done

**Closed it, on evidence.** The fix is live on `v1.0.1250`, pod-verified on **both**
replicas with a positive control — never git, never the tag. A `page-build-handler` run
created after the roll carries the section plan **flat, with no wrapper**, and both handler
and writer ran to COMPLETED. The temporary shim has been retired, so the correct path is
carrying builds on its own.

Along the way the council pushed back once, and it was right. Its gating objection was that
we had fixed the one place that tripped over the mechanism and left the mechanism itself
unguarded — and it asked for a fleet-wide audit nobody had run. **That audit found a second
live instance**, in a different pipeline, that had never produced a reported failure. It is
fixed too, and tested.

Then, cleaning up, the same class nearly caught us again: another lane's change landed in
the middle of ours, and removing our temporary path *by position* would have silently
deleted theirs. We removed it by value instead. It landed in between exactly as feared.

## Where we are now

**`bugs_closed/192`.** Council APPROVED. Live and verified. The defect is not reproducible.

Four things outlive the bug:

- **A landmine, where it fires:** reusing a key as your step's `output_field` *replaces*
  that key — it does not annotate it. The wrong result looks exactly like the right one,
  because all the data is still there one level down.
- **A new safety switch** any pipeline can take: a step can now declare a field
  non-optional, so a data-contract break fails where it happens instead of two steps later
  under an error naming the wrong file. Off by default.
- **Three wrong calls, written down**, including two of our own: we gave the council the
  right number with the wrong reasoning, and we discovered that the two obvious versions of
  the detector both report "clean" on the very bug that motivated them.
- **An architecture question routed, not dropped** — into an existing open RFC on the same
  mechanism rather than a competing new one.

## Where we're going

Nothing is owed by this lane. Three things sit with other people, all named and none
blocking:

1. The **`178` lane** re-runs its own end-to-end check, which this outage had blocked and
   which is now unblocked, with completed items waiting as subjects.
2. The **`webdesign.uk` lane** is unblocked and has been told; its shopfront build is one
   command, left for them because it publishes a live page and the timing is theirs.
3. A **human** decides the architectural question in `RFC_012`: whether the coordinator
   should warn when a step's output collides with an existing key, and whether the audit
   that found the second instance should run on a schedule rather than when somebody
   happens to think of it.

One loose end belongs to nobody yet, and it is worth someone's morning: an agent that is
spawned on every single page build has not made a language-model call in the four and a
half months our logs retain. That is not this bug, and it may be nothing — but nobody has
looked.
