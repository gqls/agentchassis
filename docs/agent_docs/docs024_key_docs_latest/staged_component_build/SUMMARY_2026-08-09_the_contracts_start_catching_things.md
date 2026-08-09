# SUMMARY — 9 August 2026: the contracts have started catching things

## What we're trying to do

Every reusable piece of a website we build — the hero banner, the news list, the contact
block, each interactive tool — should have a written contract saying what it must do, and
that contract should be enforced automatically by a real browser against the real live
page. Not a description of the piece. A set of checks, each one of which has been
deliberately broken first to prove it can catch a fault, so that when it says "pass" the
word means something.

The point is not tidiness. It is that this estate builds pages by machine, and a machine
that rewrites a component has no way of knowing what that component was for. The contract
is the only thing standing between "the rewrite ran" and "the rewrite was correct".

## Where we've come from

In late July we designed a seven-stage build ladder, and you cut it to three funded gates
because the evidence did not support the other four. Those three were built, proven, and
put into production. By 5 August the machinery worked end to end and we ran a calibration
batch of five pieces to find out what each one costs: about fifteen minutes for a simple
page section, thirty to forty-five for an interactive one. You were offered three paces
and chose the third — everything.

Since then the work has been production-line: read the real live page, write the contract
against what it actually does, break every check on purpose to prove it can fail, write it
into the database, then fire it at the live cluster with a deliberately wrong page mixed
in to prove the system refuses it. Roughly nine minutes per simple piece once warm.

By the morning of 8 August every simple page section in the estate that *could* be proven
had a proven contract — forty-two of them, plus two tools.

## What we've done since

**The interactive pieces have started, and they needed a new rule.** Seven of them now have
contracts: the news archive, the homepage news teaser, the case-studies grid, the blog
index, the contact block, the games index and the AI-readiness quiz.

The rule they forced is this. For a static piece, checking the markup is there is enough.
For an interactive one it is not: you can delete the component's entire script and every
"is it there?" check still passes, so the contract would happily certify a dead panel. So
each interactive contract now carries at least one check that **only** passes if the code
actually ran — the news page has to display its item count, which is written only after it
fetches the feed; the filter bars have to genuinely move when a filter is clicked; the quiz
has to advance a screen when Start is pressed and unlock Next when a question is answered.

One detail in that is worth repeating because it nearly slipped past: on the blog index the
contract clicks a filter that is *not* the one already selected when the page loads.
Clicking the pre-selected one would have passed with the script deleted. That distinction —
between a check that watches something change and a check that watches something that was
already true — is the whole difference between a real contract and a decorative one.

**And the contracts started earning their keep by finding things.**

The serious one: the **contact block does not send anything, and tells the visitor it
has.** You fill it in, press Send, it pauses as if talking to a server, and then says in
green: "Your message has been sent. We'll be in touch shortly." Then it clears what you
typed. There is no destination on the form, and the script the visitor's browser actually
downloads contains no code that sends anything at all. The pause is a timer. Its
validation *is* real — mistype your email and it tells you properly — which is exactly why
nobody caught it: the form looks wired up. It is live on robot-hands' contact page and on
leopardess's quiz page. Every other form on the estate has a real destination; this is the
only one.

We deliberately did **not** write "the message was sent" into that component's contract.
Doing so would have made our own quality system start certifying the claim as correct, and
we would have baked the lie in permanently.

Smaller finds along the way: five missing case-study images on finetuning; a games index
whose JavaScript is dead in two different ways at once and harms nothing; and several
database records claiming a component sits on a page that, when fetched, plainly does not
have it.

## Where we are now

**Forty-nine page sections and two tools have proven, live-tested contracts.** Every simple
section that can be proven is done. Seven of the seventeen interactive ones are done.

Three things are blocking the rest, and each has a clear unblock:

1. **Eight sections are written but stuck behind missing site images** — the same missing
   hero images our own checker found on 31 July and nobody was ever sent to repair. One
   image fix per site releases its subjects. It remains the highest-value small repair on
   the board and it has been sitting there for over a week.
2. **Ten interactive sections and ten tools left**, at roughly half an hour each.
3. **A handful belong to other active workstreams** and are deliberately left alone.

There is also a correction on the record from this session. The contact-form bug was first
written up as affecting three live pages; it is two. The rest of that write-up was verified
against the real pages and the real scripts on purpose — and then the one figure a reader
uses to judge urgency was taken on trust from a database table, three paragraphs below an
explanation of why that table is not evidence. It is corrected, and the lesson is logged
where we log these.

## Where we're going

The immediate decision is yours: **what do we do about the contact form?** Either give it a
real destination — we have two working patterns already in the estate to copy — or take the
form off and leave the contact details, which are correct and complete. The recommendation
is the first; where enquiries should land is not ours to decide.

After that the line continues in the same shape: the remaining interactive sections, then
the ready tools, with the image-blocked eight released the moment those assets are fixed.
Nothing about the method needs to change. It is producing contracts that pass, and — more
usefully — it is producing findings that nothing else in the system was going to produce.
