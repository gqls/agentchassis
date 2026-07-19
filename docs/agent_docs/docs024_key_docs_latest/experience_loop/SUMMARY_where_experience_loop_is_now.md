Summary to read aloud — updated 2026-07-19

What we're trying to do

We're building a layer that owns the experience, rather than the individual pieces. Everything the platform checked before looked at one thing at a time: is this page built, does this tool work, does this button point at a real page. Nothing looked at the whole thing a visitor actually does — whether a button keeps the promise its label makes, whether a journey across several pages reaches an end, whether the numbers printed on a page are real.

Where we've come from

This started with three broken things on our test site. Clicking an entry in the archive took you back to the same page. A tool sat saying "Loading" forever. And a third page looked like a working game but was scenery — the buttons did nothing, and it displayed twelve thousand competitors and a live leaderboard of people who don't exist. Three different faults, one common cause: every check we had was looking at one artifact in isolation, so all three were invisible.

So we built two things. First, four guard rails, to stop the underlying classes of fault recurring anywhere. Second, a planner that writes out what an experience is meant to be — the journeys, the promises, the data it needs — and a council of four critics that attack that plan until it holds up. One critic checks that every journey actually finishes. One checks it can really be built. One checks nothing is invented. One argues for cutting scope.

What we've done

The guard rails went live first and immediately turned up five problems nobody had catalogued, including invented numbers on two more live pages we hadn't known about.

Then the council refused to approve a plan, seven runs in a row. Every single refusal was correct, and each one was pointing at a fault in our own tooling rather than in the plan. One critic objected five rounds running that it couldn't confirm a component existed — our query had been showing it one of the five components on the site. Another time it caught that the plan document was being cut off halfway through, by our own size limit, in the very section the builders would need.

We fixed those, and then found two more faults of our own. A single flaky critic could destroy an entire twenty-minute run, so a critic that fails is now simply counted as not voting — except the one that checks for invented numbers, which still stops everything, because a plan must never be approved with that check quietly skipped. And the critic whose job is cutting scope had been ignored four rounds running, because our own prompt introduced it as "advisory" while the voting rules made its objections blocking. We had told the writer that objection was optional, so it treated it as optional.

Where we are now

**The council has approved a plan.** Unanimously, all four critics, on the eighth run. That was the milestone this phase existed to reach.

Two details make it trustworthy rather than lucky. First, we confirmed no critic abstained — since we'd just made abstention possible, an approval with a silent seat would have been worthless, and this one has all four voting. Second, the approval came for the right reason: the plan now defers the piece the scope critic kept asking us to drop, labels it honestly as coming soon, and where it keeps something that critic questioned, it says in writing why. That is the behaviour we asked for, in the plan's own words.

One thing we have NOT proven: the flaky-critic fix has never actually run, because no critic has failed since we made it. It's in place and correct as written, but untested in anger, and I'd rather say so than count it.

Where we're going

Next is building the thing itself, which is now unblocked — the approved plan is the live one, where before it was an unapproved draft nobody was allowed to build from. After that we extend the automated browser checks so they can follow a journey across several pages instead of testing one page at a time.

Two things to watch during the build. The site's homepage is queued for a generic rebuild, and the council itself spotted that our planned edits could be wiped out by it — that needs sequencing before any work starts, not after. And separately, this site's header, footer and page-head all point at library entries marked inactive, with the repair job sitting unclaimed in a queue since the 11th. Nothing is visibly broken, because the already-built versions are still being served, but it's stale. That one turned out to be the same underlying disease as a bug another thread diagnosed this morning — the system detects a problem, files it, and nothing ever picks it up — so the evidence went to their bug rather than into a second one.

The honest headline for this phase: we spent it discovering that our reviewer was more trustworthy than our tooling, and the last thing it caught was us not listening to it. Every one of those faults is now fixed, and the reviewer has signed off.
