# Summary, 30 July 2026 — the ladder becomes a project, and the first unknown turns out cheap

*Written to be read aloud. First summary for this lane. It exists because the work
changed status today — from a proposal waiting on somebody to an owned project with its
blocking question answered.*

---

## What we're trying to do

Stop building components in one leap. Today a component either exists or it doesn't, and
everything in between — is the shape right, does it render, is it registered, is it
placed durably, does it serve, does it actually *work* when a person drives it, does it
still work after the next deploy — lives in one session's head and in prose. We want
those to be named stages, each with a single question and a single check capable of
failing, so that a more complicated component is *more stages of the same size* rather
than a bigger gamble. And we want each small part to carry its own specification and
history in the database, the way our tools already do.

## Where we've come from

The carousel on fundamentallyai.com is the reason this exists. It took five rounds over
about twenty hours, and it was careful work: the hazards were written down before
building, a test harness ran before anything touched the database, every check in that
harness was proven capable of failing by deliberately breaking what it checked, and
nothing was ever trusted from a "complete" status — it was always checked against the
page a visitor actually gets.

And it still shipped a component whose JavaScript had never run at all, from the very
first version, for four rounds, until you clicked it.

That's the whole finding. The checks weren't weak. Every one was sound about what it
measured — and they all measured the page's code, or forced it into a state directly.
Not one ever fired a real click. What was missing wasn't rigour. It was a stage.

## What we've done

You've made this the lane's project, so it now has the same five working documents every
other workstream here keeps, and an owner rather than a note saying somebody should start.

Then we did the thing the proposal itself said to do first. I had deliberately marked one
claim as unverified rather than assuming it: whether the existing machinery that lets a
tool carry its own specification and history would fit a component without modification.
The answer is no, and it's the cheapest possible no. There's a single line in the database
listing which kinds of thing are allowed one of these documents, and "component" isn't on
it. Adding it can't break anything that already works, and it's been done four times
before for other kinds of thing, so there's a good example to copy.

Underneath that was a genuinely pleasant surprise. Two tables are involved, and only the
history one records which website it refers to. That's exactly the split we need and it
was already there: a component's *design* is shared — the same carousel serves eleven
sites — but whether it works is a question about one page on one site. So the
specification being site-less is right, and the verdicts belong in the history. We didn't
design that; the tables already assumed it.

We also caught a trap before it fired. The two tables don't currently allow identical
lists, because another team added a category last week. The obvious way to make the change
— copy one list over the other so they match — would have silently invalidated fifty-seven
rows of their work.

And reviewing the other team's report turned into a real constraint rather than a comment.
When one of our checks meets a kind of test it doesn't recognise, it doesn't fail — it
quietly *skips*. A set where everything skipped counts as a pass, and then stops
re-checking for a week. That's survivable for a single checklist and corrosive for a
ladder, because passing one rung is precisely what earns you the next. So a rung that
couldn't run its own test must now report "don't know", never "fine". It's a requirement
in the plan, not a footnote — and it isn't theoretical: the newest and most useful test we
have, the one that tells "on the page" from "big enough to see and click", was written
this afternoon and isn't deployed, so right now the best test for this project is also the
one that would silently do nothing.

## Where we are now

The lane is real, documented, and unblocked. Nothing is built yet, and the honest
position is that the remaining work is smaller than it looked a day ago — because the
mechanism that drives a real browser and asserts real interactions already exists, already
works, and was proven end to end yesterday on a different subject. It has simply never
been pointed at components. The missing stage is mostly wiring.

The most encouraging thing isn't technical. Two teams, working on different things, on the
same day, independently made the same mistake and drew the same conclusion from it. They
verified a tool by calling its own internal functions; we forced a card open directly in
memory. Both looked like verification and neither was, because both used a route a visitor
doesn't have. That's not two similar slips — it's one fault with one name, found twice, and
that's about as good as evidence for a rule gets.

I have deliberately not taken on the site maturity ladder you pointed at. The other team
proposed a division I think is better than mine: that ladder is the vocabulary of levels,
this project is the mechanism of gates, and the render-and-check feature is the instrument
both need. Three things that fit together, rather than one big thing — which means this
lane can proceed without waiting, and where the site ladder lives is still your call.

## Where we're going

1. **Get the one-line database change reviewed and applied.** It unblocks everything else
   and is perhaps twenty minutes of work.
2. **Give one real component its own specification and history** — the carousel, because
   its story is already written down in full, so there's nothing to reconstruct.
3. **Make the missing stage real**: drive a component's own checks in a real browser, the
   way we already do for tools. This is the stage whose absence cost five rounds. It gets
   trusted only once a deliberately broken component makes it go red.
4. **Then the rest of the gates**, cheapest first, each with its own deliberate breakage
   to prove it can fail.

Two things could still make this a bad investment, and they're written down so they stay
falsifiable rather than getting defended later. If nothing ever *fires* the stages, a
ladder is just documentation — our automatic sweeps are currently manual-start, so
whoever builds the next stage has to say who triggers it. And if the gates multiply into
config nobody runs, it's a net loss; we already have a measured case of that elsewhere,
where twenty-two agents were configured and only two actually ran the check in question.
