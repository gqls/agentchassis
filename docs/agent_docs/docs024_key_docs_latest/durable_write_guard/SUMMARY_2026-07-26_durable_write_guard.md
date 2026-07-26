# SUMMARY — the durable-write guard (bug 021), closed

*Written 2026-07-26, at the close of the work. First summary in this workstream:
there was nothing worth reading aloud until the thing was finished and proven.*

---

## What we're trying to do

Stop the platform quietly destroying its own work.

When an AI writes a page component, a tool, or a block of copy, that output gets
saved into the database as the durable copy — the thing every future page build
reads from. If the AI's output was cut off halfway, we would save the fragment,
report success, and carry on. Nobody would notice until a live page came out
broken, and by then the good version was often gone.

We had already fixed one instance of that. The question this workstream existed
to answer was the bigger one: **where else does the same shape happen, and can we
close the class rather than the instance?**

## Where we've come from

The bug was filed by the thread that fixed the first instance, and it did the
right thing: instead of declaring victory, it wrote down "this exact shape is
probably elsewhere" and named two places it thought were exposed.

**Both of those guesses turned out to be wrong, and finding that out was the most
valuable day of the work.** One of the two places has no code anywhere that writes
to it — the columns are dormant, so a guard there would protect a write that never
happens. The other is not a durable copy at all; it is regenerated from the copy
we already protect, so if it is ever damaged we simply rebuild it.

Chasing those two would have produced a guard that blocked legitimate work at a
seam that was never at risk. What the evidence pointed at instead was somewhere
nobody had looked: not the moment we *overwrite* something, but the moment we
*first create* it. And that one was not hypothetical — eight broken tools were
sitting on six live sites at that moment, every one of them born broken and saved
without complaint.

A second, separate problem was attached to the same bug later. We have a mechanism
that is supposed to ask "did that job actually fix the thing it claimed to fix?"
before marking work complete. It had been built properly and then used **once** —
one job type out of roughly seventy. Everything else was taken at its word.

## What we've done

**The first half — refuse to save a truncated creation.** We now check the
structure of anything the AI produces at the moment it is first saved, using the
*same* check the system already applies when loading it later. A tool that would
be rejected on the way out can no longer get in. If something is refused, it is
recorded and routed to a human rather than silently dropped.

We proved it by deliberately breaking things. We fed it a tool cut off mid-script:
refused, logged, nothing saved. We fed it a healthy tool: sailed through. That
matters more than it sounds — a guard that only ever says no is easy to write and
useless, because people switch it off.

**The second half — ask whether the fix actually worked.** We found the real
reason that "ask if it worked" mechanism had only ever been used once, and it was
not, as everyone assumed, that authors kept forgetting. **It was unusable.** The
mechanism was never told *which site* it was checking, and most of these checks
are impossible to write without that. It was not a discipline problem; it was a
missing parameter. We fixed that, then wrote the check for the job type that had
been blocked on it, and added a guard that now **breaks the build** if anyone adds
a new kind of job without either checking it or writing down why not.

The subtle part is what the check actually asks. The obvious version — "re-run the
scan that found the problem" — would have been a quiet disaster. Our scanner flags
any hardcoded colour; our fixer only changes a specific narrow kind. Across the
fleet the scanner currently flags 32 things on 8 sites, and **on five of those
eight sites not a single one is something the fixer was ever built to touch.** The
obvious version would have refused every completion on those five sites for ever,
retried them, and finally filed them as failures — punishing hardest the sites
where nothing was wrong. So the check asks the right question instead: *would the
fixer itself still change anything?*

We proved that one both ways too, on the live system, having worked out the
correct answer in advance so the test could actually fail. On a site with three
genuinely fixable components it refused the completion and named the right
component. On a site with eight flagged and none fixable it let the job through.

## Where we are now

**Bug 021 is closed.** Both halves are live in the current build, and neither was
signed off on the basis that the code shipped — both were demonstrated firing, and
demonstrated *not* firing when they shouldn't. We also found the new check had
already quietly passed a real job on a live site that morning, with nobody
watching.

An independent review panel approved the work first time round. Its one
substantive objection turned out to be wrong — and wrong because of something we
did: when we quoted a database query as evidence, we shortened it and dropped the
one line the objection turned on. The reviewer reasoned perfectly from what we
showed it. That is now written down as a rule, because reviewers cannot open our
files and only ever see what we paste.

One loose end was deliberately not swept up with the closure. Our scanner and our
fixer disagree about what counts as a problem, which leaves eight items in the
backlog labelled as though a fixer kept failing on them, when in truth there was
never anything for it to do. It costs nothing and breaks nothing — it just makes
the backlog lie a little. It is filed separately, with three options written out
and none chosen, because choosing is a judgement call for whoever owns that
scanner.

We also kept an honest record of getting things wrong. Four missteps this session,
all the same underlying error: reasoning confidently from evidence that was there
but incomplete. Once I decided a message had been lost when it was merely queued,
"fixed" it, and wrote the false fix into our runbook before disproving it. The
fleet-wide ledger of wrong calls took four increments from this session and **no
new category** — which says the class is already named, and what is missing is not
more care but reaching for the one-line check.

## Where we're going

Nothing here is blocking, and there is no work in flight.

The natural next step belongs to the other workstream: the same "did it actually
work?" check for the next job type, which is now the highest-volume one we take on
trust. Its handler needs reading first — that lesson has been paid for twice.

Deliberately **not** next: raising the coverage number for its own sake. Three
checks out of seventy-seven sounds thin, and it is the right number. Every other
job type carries a written reason for why it is not checked, the build enforces
that, and the one time somebody wrote a check to improve the ratio it would have
broken 1,849 items. The list is the backlog; it names its own next candidate.

The reusable thing that came out of all this is not the fix — it is the test rig.
We can now take any part of the system, hand it a payload we control, and watch it
succeed or fail on the live cluster within seconds. It has been used on two
completely different problems already, and it is written up so the next person
does not build a third one.
