# SUMMARY — 2026-08-19 · `bugs_open/302` CLOSED: proven in production, four-way

Second and final read-out for this lane. Supersedes nothing — the 08-18 summary said *"installed,
not yet seen working"*; this one says proven and closed, which is the inflection.

## What we're trying to do

Stop the platform stamping repair jobs "done" when nothing was repaired.

## Where we've come from

The owner ruled on 2026-08-08 that a completion check which *cannot run* must not pass the job — "I
could not check" is not "I checked and it is fixed". A sibling check written five days later did the
opposite: when it could not read the repair agent's report, it waived the very assertion its opt-in
roster exists to make, and completed the job. Another lane filed that as `302`, scoped it, and handed
it on.

## What we've done

Measured it first, which corrected three of the filing's own claims — the checker registry holds
thirteen types not eleven; seven of the eleven bad cases belonged to a different bug that shipped a
fix the day before (939 broken payloads before that release, **zero** across 1,880 completions
after); and the fix the filing recommended is one the estate had already declined in writing.

Found the sharper defect: **five of the eleven waived completions were jobs the gate had already
refused one attempt earlier**. It was not failing to grade — it was reversing its own refusal.

Fixed the arm as a per-type declaration whose zero value is not a policy: undeclared is a build
failure, and undeclared at runtime abstains rather than blocks. Proved it by breaking it six ways.
Council **APPROVED round 1**, and both advisory objections earned their keep — one caught a false
sentence in our own submission, the other asked for the blast radius by query and turned up a live
scheduler reading this job type as 86.7% successful, a figure that is an artefact of exactly the
false greens the fix removes.

Then proved it in production, four-way: no report → **refused** with the new reason; a readable
all-zero report → refused with the **other** reason (so the two stayed distinct); a report with real
work in it → **completed**; a job type not on the roster → **completed**. The harness spawned no
repair agent and touched no site; teardown was verified at the data, with the live scheduler's view
of the real success ratio unchanged.

Closed `bugs_open/201` on the way, on evidence rather than on its own account.

## Where we are now

**`302` is closed.** Fixed, live on `v1.0.1314`, proven, council-approved, registered as an amendment
to WII-017 with WII-011 carrying the new fifth status value.

The cost of the fix is stated rather than glossed: a refused job burns its attempts and waits for a
human, because the mechanism that would release it has never run and this producer's scheduler is
off. That is the cost the owner knowingly accepted for the 08-08 ruling.

## Where we're going

Nothing is queued here; three things outlive the lane and none is this lane's to decide.

- **`bugs_open/317`**, filed so closing `302` loses nothing: the claimed-item-timeout sweep can
  complete one of these jobs with neither gate running, because its exemption list is locked to the
  other gate's list. Zero occurrences, and only because both carriers are disabled — re-enabling one
  re-arms it.
- **A semantics decision** on `needs_design_review`: an analysis blob may legitimately be the
  deliverable, in which case "changed nothing" is a success and no rule should be written.
- **A measurement** nobody has taken: what a successful `spacing_fix` / `responsive_fix` report
  actually looks like. Until then their roster entries would be a guess about somebody else's handler,
  which is the one thing that roster forbids.
