# Gripper dossier — where we are, 2026-07-31

*A new file, not an edit of `SUMMARY_2026-07-27_gripper_dossier.md`. The series is the
record, and this entry exists because that one's central claim needed correcting.*

## What we're trying to do

Robot-hands.com gives away four good gripper calculators and then lets every visitor
leave. The dossier is the thing that doesn't. An engineer describes the part they need to
pick up; we produce a proper written report — a scored shortlist with the arithmetic
printed out, safety factors, environment gating, throughput, mounting compatibility, and
the questions to put to a vendor, every specification traced to where we read it and when.
It arrives as its own page, and the link goes to them by email.

It is the first paid-shaped thing on any of our sites, and it is the pilot for a pattern we
mean to repeat across a thousand domains: free tools bring people in, one signature piece
of work is worth paying for, and the machine that makes it is the same machine everywhere.

## Where we've come from

The cluster half — take a request, score the grippers, write the prose, check the prose is
honest, build the page, publish it — was built between 24 and 27 July. It took six new
pieces of machinery, four database seeds and three rounds of review, and on the way it
turned up two platform defects that had nothing to do with grippers: a whole class of AI
answers being cut off mid-sentence and filed as complete, and no way for a workflow to
declare that it had failed, so a failed job could report success.

The public half — the form, the abuse protection, the email — was designed as a new service
on the island machine and then designed out of existence, because a different thread had
meanwhile shipped a shared public API onto the same box that already had the rate limiting,
the request cap, the key and the site list in it. The owner caught that by asking a
question; nothing in the platform did.

On the 27th all three test fixtures passed, including a deliberately induced failure, and
I wrote that the cluster half was proven end to end.

## What we've done

Since then the two shared pieces the public half genuinely needed have been built,
reviewed and approved: one email sender, and one guard that does per-visitor rate limiting,
CORS and honeypot in a single place. Neither is called by anything yet, which is the honest
state and is written down as such.

And then the owner read the actual report pages for the first time, on the 30th, which had
not happened before — and found two faults in the headroom chart that every check we own
had passed.

That is the part worth dwelling on. One of the numbers beside a bar read `6.42×
(Insufficient` and simply stopped, as though the figure itself were damaged. Underneath,
two small grey captions were printed on top of one another into something unreadable. Both
were invisible to automation by construction: the text in the file was correct and complete,
and it was the drawing area that cut it off. So there was nothing wrong to detect. A reader
of a report full of computed figures, seeing a number end mid-word, would reasonably
conclude our arithmetic was broken.

Chasing it turned up a third fault nobody had asked about and which is the worst of the
three: bars that ran off the end of the scale were drawn exactly as long as each other and
exactly as long as an honest three-times bar, so the chart gave no sign it had stopped
measuring. All three are now fixed, reviewed and live, and this morning I regenerated the
report from the very same inputs as Monday's and looked at the two pictures side by side.
The label reads in full. The captions are on separate lines. Capped bars end in a point,
like an arrow, and the one bar that isn't capped still ends square.

Two smaller things are worth keeping. While fixing it I wrote a test that passed for the
wrong reason — I broke the fix on purpose and the test stayed green, because a second
safety net quietly shortened the label instead. And the honesty checker that guards these
reports refused to publish one of them because the writer had composed the phrase
"IP54-or-better" out of a fact it held; the checker reads that as an invented part number.
The checker is right to be strict and I have not weakened it, but it is fail-closed, so
being right cost us the entire report rather than one sentence. That is filed as a bug with
fix options, and it is intermittent, so it will look like a flaky pipeline to whoever meets
it next.

## Where we are now

The pilot works and, for the first time, that statement rests on somebody having looked at
what it produces rather than on a row of green checks. The chart fix is live and verified on
the page itself.

The public half is still not built, and the reason is no longer technical. It belongs inside
the shared API, which is another thread's code with an open bug against it, so the remaining
step is a decision about order and ownership rather than a piece of work. That decision, and
the tidy-up of the three test pages now live on the site, are both sitting with the owner.

The correction to the 27th is the thing I would want said out loud: "all three fixtures
pass" was true, and it meant less than it sounded. Every automated check we had passed
against a page with a visibly broken number on it. What closed the gap was rendering the
thing and looking at it, which is now written into the workstream's notes as the step that
cannot be delegated to a checker.

## Where we're going

Finish the public half inside the shared API once the ownership question is answered — a
route group, not a fourth copy of a web service. Then generalise the scorer, because it is
Go written for robot-hands specifically, and a paid product that needs bespoke Go per site
does not reach a thousand domains; the pattern to copy already exists elsewhere in the
platform, where the rules live in a table rather than in code.

Two smaller debts stay open and stated rather than implied. The label-guard bug needs a fix
that lets a legitimate recombined phrase through while still rejecting an invented sibling
part number — both halves have to be proven, or the fix is a hole. And one property of the
new shared guard cannot be tested here at all: it needs a connection from a genuinely public
address, and every address a development machine can offer is a private one. That is a
deployment question, and it should not be closed by writing another unit test.
