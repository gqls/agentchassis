# SUMMARY 2026-08-19 — the phone-dialling button: protection is live, the repair is one decision away

*(First summary for this lane. Written at the milestone where the detection and protection
halves went live and the remaining work reduced to a single owner decision.)*

## What we're trying to do

A button on the webdesign.uk home page has words about one thing and a link to another: the
copy talks about answering a few questions with the Brief Starter tool, and clicking it opens
the phone dialler. The owner found it himself, which is the part that matters — no queue
caught it, no check saw it, and it had survived several complete rewrites of the page.

So the job was never "fix the button". It was to answer why a system that already had a fix
for misdirected buttons produced this one, and to close that at the framework level, since
whatever is wrong here is wrong for every site.

## Where we've come from

The lane opened on 18 August and found four separate faults stacked on top of each other,
each of which alone would have been enough to keep the button broken:

- The part of the system whose whole job is to work out where each button should point **was
  computing the right answer and having it thrown away** — the step that hands finished
  sections to the page writer looks for that answer under a name it is not stored under, and
  falls back silently to the previous build's values. That is now `bugs_open/312`, and it is a
  recurrence: the same seam broke in June in the opposite direction.
- The checker for "the words promise one thing, the link does another" **skipped phone and
  email links entirely**, so this exact button was invisible to it.
- The repair machinery, had it ever touched the button, would have **destroyed genuine "call
  us" buttons** elsewhere rather than fixing this one — an accident confirmed in production on
  another site.
- The phone numbers themselves were written in a form phones cannot dial.

The fix was built, calibrated against the whole fleet before shipping (a first-draft detector
would have raised 226 alarms of which ~211 were wrong; narrowed, it raised 17, all of them
real), reviewed by the council — REVISE on round one for a form problem, APPROVED on round two
— and committed with three switches deliberately left off, because throwing the wiring switch
before the protective code was live would have started clobbering good buttons fleet-wide.

## What we've done

Today the protective code **rolled into production**, and we verified it is genuinely running
rather than assuming it from a version number. Two of the three switches are now thrown: the
detector for this class of fault is live on the completeness checker, and the mechanism that
tells the copy writer what a button actually points at is armed.

We also closed out the four checkable objections the council raised but nobody had answered.
Three held up under checking; one did not, and it was ours: our own submission had asserted
two contradictory things about existing code, and the safety-critical half happened to be the
correct one. That near-miss was luck rather than design, so it is written down.

And we replaced the single traced build with a fleet-scale proof, using a sharper instrument
we did not have before. The link resolver writes a small note recording what each button
points at, and nothing else produces those notes — so their absence downstream *is* the
discard, provably, even on builds where the link happens to be correct already. Across every
recent build we could examine: 26 answers computed, **zero survived**.

## Where we are now

The button is still wrong on the live site, and it changed for a fourth time yesterday — new
words, same phone link. That is the diagnosis confirming itself, not a setback.

Everything protective is live and verified. The only remaining step is the third switch, which
makes the system actually use the link resolver's answers. Its technical precondition is now
satisfied — both protective halves are demonstrably in the running binary — but applying it
changes how buttons are written on every site at that instant, so it wants a human's go-ahead
and one site watched closely as it happens.

One further thing turned up while doing the safety check that was meant to gate all of this.
**The estate's standard method for confirming new code is live does not work most of the
time.** The service announces its version at startup, but that announcement scrolls out of the
log within hours, and the documented fallback returns a confidently wrong answer — we got
"your fix is not there" for code that certainly was. Thirty-two migrations promise to perform
that check; none can enforce it. We wrote up what actually works, and filed `RFC_040`
proposing to make it mechanical.

## Where we're going

1. The owner's decision on applying the wiring fix, with `leopardessconsulting.co.uk` as the
   canary — four hand-written contact buttons that must survive untouched.
2. Then rebuild the webdesign.uk home page through the normal pipeline and check the button by
   its text, never by searching the page for the correct address — the navigation and footer
   already link that tool correctly, so a page-wide search passes today while the button stays
   broken.
3. Two smaller answers still outstanding: whether that button should end up as a phone button
   with honest wording or a link to the tool, and confirmation of the intended phone number.
4. Separately, and on its own track: whether `RFC_040` is worth building.
