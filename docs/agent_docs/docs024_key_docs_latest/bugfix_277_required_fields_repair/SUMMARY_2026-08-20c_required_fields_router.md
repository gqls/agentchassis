# SUMMARY — 2026-08-20c — you said yes, so the repair that edits finished pages now exists; it waits on a release and one supervised first run

**Written because the previous read-out (`SUMMARY_2026-08-20b`) ended on a decision that was yours,
and you made it the same afternoon.** That file stays as written; this is the position after acting
on your answer.

## What we are trying to do

Unchanged: when the system inspects a site it has built and files a note about something wrong, that
note must reach something that can actually fix it — and we must be able to point at a repaired page
and say so.

## Where we have come from

By lunchtime today we knew the seven stubborn findings — backticks showing around code words on
developer-tool pages — could not be repaired by anything we had, and we finally knew *why*: the words
live only in the finished page, and every repair we own works by rebuilding a page from its source
data, which for these pages holds nothing. Fixing them meant building a new kind of repair, one that
edits the finished page directly. Whether seven mild defects justified that build was your call, and
you called it: yes.

## What we have done

**Built it, this afternoon, and kept it deliberately small.** The new repair does exactly one thing:
it finds backticked code words in the visible text of a page and wraps them in proper code
formatting — the thing the original author meant. It cannot touch scripts, styling, existing code
blocks or anything inside the page's machinery; that boundary is enforced by making the repair see
the page exactly the way the detector that files these notes sees it, so the two cannot disagree.
It rides the same page-editing machinery the estate already trusts, with all its existing safety
rails, and the switch that permits it is off everywhere except the one agent that needs it — your
standing rule for new powers on shared machinery. The system also now decides *by itself* which
repair a finding needs, using the test we wrote into the escalation instructions yesterday: it asks
whether the page's source data could rebuild the page, instead of asking who owns it — the question
that three days of wrong fixes taught us to ask. If in any doubt, it routes the old way, which ends
at a person.

The change passed our tests — including one run against the real bytes of a live tool page, chosen
because it contains both the defect and a booby trap: a piece of genuine code that a careless
version of this repair would have corrupted. It went to the review council this afternoon and is
written up in the shared register so other threads know it exists. Along the way we found and
recorded a genuine trap in the HTML library itself, the kind that passes every ordinary test and
corrupts pages in production.

## Where we are now

Nothing is repaired yet, and that is expected. The code travels with the next fleet release. After
that, the dispatcher will rightly refuse to trust a brand-new route until it has one success on
record — so the first repair must be run deliberately, watched end to end, and verified on the
served page itself. That one success unlocks the remaining six, which then flow on their own. The
step-by-step instructions for that first run are written down beside this file.

## Where we are going

The supervised first run, after the next release — it doubles as the proof this lane has owed from
the start: a page that was wrong, repaired, verifiable in the bytes a visitor receives. Then the
honest remaining work is unchanged: the 27 jobs whose pages have no content at all, which need a
different repairer entirely, and the two smaller debts named in the handoff.

Full state: `HANDOFF_2026-08-20b_continue_here.md` beside this file.
