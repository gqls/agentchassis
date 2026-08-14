# SUMMARY — 2026-08-14 · the planner can see the calculators

*(the framework-rebuild thread; the copy/voice thread and the 227 platform job are
separate and already summarised)*

## What we're trying to do

loancalculator.co.uk was built by hand and then adopted into the platform, so today
it is a working, indexed, twenty-seven-page site whose most valuable parts — twelve
interactive loan calculators whose arithmetic has been independently proven — are
protected by permanent locks. The aim is to rebuild the whole site through the
framework, so that every page becomes something the platform can audit, regenerate
and improve, without losing the calculators and without moving a single page
address. The brief for that rebuild is approved, with those two things pinned.

## Where we've come from

One thing stood between us and firing the rebuild: the site planner — the agent that
decides what each page contains — could not *see* calculators. Its menu of building
blocks listed text and layout components only, so any plan it wrote for a calculator
page simply left the calculator out, and the build would then shove the locked
original to the foot of the page. Half of the fix (teaching the build to pair a
planned section with its locked original) went live and passed review two days ago;
the other half — the menu itself — was deliberately not rushed, because it hangs on
one wire: the query needs to know which site is asking, and if that lookup ever
fails, no site on the platform can plan anything. There was also a puzzle: the
planner appeared never to have run at all, anywhere, ever — no trace in the run
records — while finished plans kept appearing.

## What we've done

The puzzle dissolved first: the run-records table only keeps about two days of
history, and the planner is so rarely used (three times, ever) that its records are
always gone before anyone looks. Nothing exotic — a short shelf and a rare visitor.

The wire was then settled three ways rather than guessed: the planner's own workflow
creates the site record two steps before the menu is read, in code that cannot skip
it; the step that writes finished plans has used the exact same lookup forever, so
every plan ever written is a successful test of it; and we ran the new query against
the live database and confirmed that every site without the opt-in flag gets
byte-for-byte the menu it got before. The change went in with a snapshot, guards
that abort the transaction if anything is off — guards we first deliberately
triggered, to prove they fire — and the flag was switched on for this one site only.

The review council sent it back once, and the round earned its keep: the gating
objection was that I had *stated* a protective fix existed elsewhere without
attaching proof, while older warnings still said the opposite. The proof existed; it
is now attached (including a probe of the running software itself), the stale
warnings are corrected in place, and the second round approved with nothing
high-severity. Two smaller things came out of it too: a filed follow-up for counting
how many opt-in flags accumulate on a site's spec, and a single authoritative census
of the five that exist today.

Then the trial: one real planner run against the live site. The mechanics all
proved out — the planner's menu carried all eleven calculator components, the site
lookup resolved, nothing failed anywhere, the locked rows were untouched to the
byte, and all twenty-seven pages still serve perfectly.

The honest record includes one false alarm of mine, logged where such things are
tallied: mid-session I briefly reported that a routine platform re-render had
stripped the calculators off the live site. I had been checking pages at addresses I
constructed from their internal names rather than their real URLs, so every "page" I
was reading was a 404 error — which greps clean for anything you search. The
database disagreeing with the wire is what sent me back to look.

## Where we are now

The planner can see the calculators; nobody else's behaviour changed; the site is
clean and serving. The trial also produced two findings, both contained and both
more valuable than a simple pass would have been.

First, the planner *invented* two pages we never asked for — an "about" page and a
guides index — and queued work to build them and generate imagery, including a step
that would have published the two of them as empty shells. All of that is parked
with reasons written on it, and the two empty page entries are archived; nothing
reached the live site, and reversing the containment is one command if the owner
actually wants those pages.

Second, and the real discovery: a plain replan writes page compositions only for
pages it invents. For the twenty-seven pages that already exist, it proposes no
layout at all. So the question we most wanted the trial to answer — "will the
planner keep the calculators in place when it recomposes a page?" — never came up,
and *cannot* come up in a trial of this shape. It will be answered the first time a
built page is genuinely recomposed, which is the rebuild itself; the locks and the
pairing fix are the safety net there, and the worst case they permit is a calculator
displaced on a page, never deleted.

While we waited, the platform's routine maintenance also touched the site — an
expected one-off re-render wave after a fleet fix, plus one new FAQ guide page the
gap-planner created on its own. All verified harmless: claims still cut, calculators
intact, locked chrome held. One consequence for later: the wave is stamping
fingerprints on page rows, so the rebuild's before/after audit must lean on the
backups taken on the eleventh, not on the live table.

## Where we're going

The rebuild is ready to dispatch and is waiting on two answers from the owner, not
on code. One: how the twenty-six existing pages should be regenerated — the trial
says a plain replan will not do it, so the honest route is to name the pages to be
recomposed explicitly, in both the dispatch and the briefing text, and that is
where the calculator-placement question finally gets its answer. Two: whether the
mission brief's "keep the pages" instruction is trusted to stop the planner
inventing pages, given that an unbriefed run invented two. After that: release the
eight non-calculator locks, fire the mission, and judge the result against the
pinned checks — addresses unchanged, calculators in place, and nothing new that
nobody asked for.
