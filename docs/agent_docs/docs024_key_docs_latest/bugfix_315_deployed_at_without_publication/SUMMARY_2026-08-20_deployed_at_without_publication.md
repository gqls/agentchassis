# Summary — bug 315: the fingerprint is built and reviewed; arming it broke the fleet and is on hold

2026-08-20. Written to be read aloud. Supersedes nothing — the 08-19 summary stands as what we
believed then, and the distance between them is the point.

## What we're trying to do

Make it possible to answer one question honestly: **has this page actually reached the website?**
The database has a timestamp everything treats as that answer, and it is the answer to a different
question.

## Where we've come from

The 08-19 summary ended with a decision waiting on the owner and nothing shipped. Since then: the
owner ruled (record a fingerprint of the bytes we send, at page level, and leave the section-level
column alone); both halves were built; the review council put it through three rounds; and it was
switched on.

**The review earned its keep twice.** Round one judged the plan and found we had promised more than
we were delivering. Round two judged the code and found a claim in it that was **flatly untrue** — we
had justified a piece of the design by borrowing a safety guarantee from a shared helper, and the
helper's own comment states that guarantee as an *intention* and then, four lines later, states that
the current version does the opposite. We had read far enough to find a sentence that suited us and
stopped. Round three approved the corrected version.

## What we've done

The deployment service now reports what it actually did — the reference of the commit it made, and a
fingerprint of the exact bytes of each file — instead of returning a constant that identified
nothing. The page-marking step reads that, refuses to mark a page deployed when the deployment
reported doing nothing, and records the fingerprint.

Then we switched it on, and **it broke every page publish in the estate for thirty-three minutes**.
Eight jobs failed and a hundred and twenty-three queued rebuilds stopped draining. The cause was a
one-line mistake: the file describing what settings each component accepts contains two such
descriptions forty lines apart, and the new setting went into the wrong one. Another session found
it, traced it to the line, wrote it up as `bugs_open/336`, and restored service by running our own
rollback file.

## Where we are now

Service is healthy and confirmed so, not assumed. The mistake is corrected in the code — by that
other session, not by us. The new checking is **switched off and must stay off** until the corrected
code reaches a running build; switching it on today would reproduce the outage exactly, and the
switch-on instructions have been rewritten so they cannot be followed against the wrong build.

**The uncomfortable part, and it is the most useful thing this lane has produced.** After switching it
on we checked that the setting was present in all three places, got the three expected answers, and
wrote "verified". That is a status check. We never asked whether a page could still be published —
and it already couldn't. Our own notes ninety minutes earlier said *"config being right is not the
artefact — that is this bug's entire lesson, and it applies to the fix as much as to the defect."*

**So this bug's exact defect was committed by the lane fixing this bug, at the moment of fixing it.**
That is not a tidy irony; it is evidence about how the failure works. It is not caused by ignorance of
the rule — we had written the rule down that morning — but by looking for a change's *benefit* rather
than its *damage*. The number we were watching sat at zero, we correctly refused to call it success,
and then read it as "nothing has run yet" when it meant "nothing can run". Those are indistinguishable
in that column.

## Where we're going

**One build**, which picks up the correction. Then the switch can be thrown, and the first thing done
afterwards is to look for damage rather than for success — that ordering is now written into the
migration itself.

After that the platform will finally be recording *what it sent*, which is the thing that makes the
question answerable at all. The last piece — a scheduled check comparing that record against what the
website actually serves — is designed and unbuilt. It is the piece that would have caught the original
bug six hours before anyone noticed, and it is worth nothing until the fingerprints exist, which is
why it has stayed last.

One guard we have deliberately not claimed: a test that every setting a component reads is declared in
its own list, and that no list declares a setting its component never reads. It would have prevented
the outage outright. It belongs with whoever owns that seam — this lane has just demonstrated it
should not be the one grading its own homework there.
