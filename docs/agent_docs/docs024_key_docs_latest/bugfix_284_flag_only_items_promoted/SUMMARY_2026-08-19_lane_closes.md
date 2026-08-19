# Summary — 2026-08-19: the lane closes, and what it hands on

## What we were trying to do

Findings the system makes about its own sites were being detected and then going
nowhere. Some were promoted into a work queue and immediately stamped as blocked;
others were filed under a category no handler had ever been registered for, so they
simply died where they were written. The job was to find out why, fix it at the
mechanism rather than the symptom, and prove the fix on the live system rather than
in a test.

## Where we came from

Four separate faults, all wearing the same disguise — work that looked handled and
was not. A router inventing category names nothing could route. An auditor that
produced good findings nobody ran. Flag-only findings promoted into a queue that
then rejected them. And an agent seeded without a description, which could not be
started at all and looked like an intermittent network flake for weeks.

## What we have done

All four are fixed, live, and verified against the running system rather than the
code. The repairs went in at three depths: the code that produced the bad rows, the
data that was already wrong, and a database constraint so the bad state cannot be
written again by any route, including someone typing SQL by hand.

Along the way the owner asked three questions and got three answers, one of which
was "nothing to build, this already exists and the instructions were three days out
of date" — so the instructions were corrected instead. He also settled a question
about the fundamentallyai site's writing: two or three articles per tool is fine.

Yesterday the same thread found and confirmed a fifth fault, on that site's Platform
Log index: it lists six articles and links none of them. That is now diagnosed to the
mechanism, agreed with a second team working the same bug, and half-fixed.

## Where we are now

The lane is finished. The last thing we were waiting on resolved itself this morning:
the constraint we added could only be proven safe by argument until something real
passed through it, and overnight forty-two items were promoted through it correctly,
every one properly routed. That is the evidence we wanted, and it arrived from normal
traffic rather than a test.

The index fix is the one thing that outlives this lane. The configuration change is
live and proven — the page's data now resolves eight real articles with working links
— but the page itself is unchanged, because a safety guard refused the write. It was
right to: five of the eight articles have no summary text stored, so the new cards
would have shipped half-empty. The old cards hid that by having the model invent
copy, which is also why the live page currently shows publication dates that are
about eighteen months wrong.

## Where we are going

One thing, and it is not ours: those five summaries need writing by the framework,
not by a person and not by us. Once they exist, re-running the same command finishes
the job — we have measured that it would then clear the safety guard comfortably.

The reason it has not fixed itself is worth saying plainly, because it is this lane's
own subject one more time: the system has already noticed the missing summaries, six
hundred and six times across the estate — and every one of those notices is sitting in
a category with no handler attached. The fault we spent this lane fixing is the same
fault now standing between us and the last step.
