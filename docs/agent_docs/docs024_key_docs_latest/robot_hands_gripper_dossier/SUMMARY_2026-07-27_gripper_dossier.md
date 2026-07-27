# Gripper dossier — where we are, 2026-07-27

*A new file, not an edit of any earlier summary. The series is the record.*

## What we're trying to do

Robot-hands.com sells nothing today. It has four free calculators on it that
work out gripper physics — payload, grip force, cycle time, and a matrix that
scores ten real grippers against a job. They are good, they are honest, and they
are free, which means they bring people to the site and then let them leave.

The dossier is the thing that doesn't let them leave. An engineer describes the
part they need to pick up, and we produce a proper written report: a scored
shortlist of grippers with the arithmetic printed out, safety factors,
environment gating, throughput projections, mounting compatibility, and a list
of questions to put to the vendor. Every specification traced back to where we
read it and when. Delivered as its own page, and the link emailed to them.

It is the first paid-shaped thing on any of our sites, and it is the pilot for a
pattern we intend to repeat: free tools bring people in, one signature piece of
work is worth paying for, and the machine that produces it is the same machine
everywhere.

## Where we've come from

The calculators were built and verified in July. The site was cleaned of
invented statistics and fenced so it cannot invent more. That left the top of
the funnel missing, and this workstream opened on 24 July to build it.

The design split into two halves. The **cluster half** — take a request, score
the grippers, write the prose, check the prose is honest, build the page, deploy
it — and the **public half**, which is the bit that visitors actually touch: the
form, the abuse protection, and the email.

The cluster half was the hard one and it is done. It needed six new pieces of
machinery, four database seeds, and three rounds of council review. Along the
way it turned up two genuine defects in the platform that had nothing to do with
grippers: a whole class of AI responses being cut off mid-sentence and saved as
if complete, and a missing way for a workflow to declare that it had failed —
so a failing job could report success and be filed as done beside the evidence
that it hadn't worked.

## What we've done

The cluster half is built, reviewed, and committed. It is inert until the next
image roll, which is by design.

The honesty machinery is the part I'd point at. The report is written by a
model, and a model will happily produce a number that looks right. So the
scoring step emits a block of facts, and a separate deterministic check then
refuses to publish if the prose contains a number that isn't in that block, a
part code we never saw, a vendor we never assessed, an empty section, or any
sign the model's answer was truncated. It cannot be talked round, because it
isn't a model — it's arithmetic.

Two of my own errors are worth recording because they nearly shipped. I stamped
"reviewed by the council" on a commit without reading the verdict, which was
"revise" rather than "approved" — and that then became subtler, because the
same submission was later approved for a *different* version of the plan, so the
stamp now points at a real approval for code it doesn't describe. And the report
page would have gone to a paying customer completely unstyled, because the way
these pages are assembled means the styling has to travel with the page, and
robot-hands.com had no styling for a page type that didn't exist until this week.

## Where we are now

The public half has been redesigned out of existence, and that is the most
important thing on this page.

I had specified a new service on the island VM — its own database, its own AI
key, its own rate limiter, its own web route. One day after I wrote that, a
different thread shipped `tools-api` onto the same machine: a shared public API
built from the start to serve many tools across many sites, with the rate
limiting, the request-size cap, the key and the site list already in it. My
service would have been a second copy of all of that, on a machine with one
processor and 2GB of memory.

Nothing caught it. Not a hook, not the council, not a lint. The owner caught it,
two days later, by asking how the dossier would fit with the other tools.

That is a better outcome than it sounds, because the cause turned out to be
worth more than the near-miss. I *had* searched for prior art, on the 24th, and
the search was correct — `tools-api` did not exist yet. The failure is that
nothing in the platform ever looks again. Every review we own judges a proposal
against a snapshot, and this repository moves fast enough that a design document
outlives the world it was written in.

Chasing that produced three findings, all verified. The council refuses to read
documents, for good reasons of cost — which means a document that decides to
build a new service is refused by the one mechanism that would object. The
council seat whose whole job is asking "does this already exist?" asks its
code-shaped questions into a void, because the step that answers them was
deliberately left off that council. And the code index behind that seat never
records what functions actually *do* — only their names and signatures — so a
search for any route, table name or piece of text comes back empty, and the seat
whose job is policing claims of absence is being handed manufactured absence.
Its own documented example cannot work.

Nothing in the report pipeline is live. Both scheduled jobs are seeded switched
off deliberately: prove the builder on a hand-made request first, then let real
visitors in.

## Where we're going

Finish the pilot as it stands. Roll the image, apply the seeds, run three tests
including one deliberately induced failure, and prove the whole shape works end
to end on one site. Then generalise — because the scoring engine is Go code
written for robot-hands specifically, and a paid product per site that each needs
its own bespoke Go does not reach a thousand domains, let alone the sixteen
hundred on the list.

The public half moves into `tools-api` as another route rather than another
service. Before that happens, three things it genuinely lacks should be built
once and shared, not forked: the ability to send email — which does not exist
anywhere in our built code, only in one VM app outside the build — a single
rate-limiter and abuse guard, and the status poller.

And the divergence check gets built and measured, because it is cheap: a
document that proposes a new binary gets told which binaries already exist. Run
against real history it fires roughly once in every 150 commits, and it fires on
the exact commit that opened this workstream — two days before a human noticed.
