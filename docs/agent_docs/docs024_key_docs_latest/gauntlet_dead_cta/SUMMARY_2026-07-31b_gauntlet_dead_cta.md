# SUMMARY — gauntlet_dead_cta, 2026-07-31b

Written to be read aloud. A new file, as the rule requires. The previous read-out
is `SUMMARY_2026-07-31_gauntlet_dead_cta.md`, which covers the sealed-provocation
work by another thread in this lane; this one covers the share card and the public
record, which it does not mention. The series is the record.

## What we're trying to do

Make vonc.com a site a stranger would argue on. The Gauntlet is the argument: a
timed debate against an AI opponent that files a real position, gets a real
challenge, defends it, and is judged. The question this piece of work answers is
what a person can *do* with a debate once they have had one — because until today
the answer was "look at it, and then close the tab".

## Where we've come from

The owner said, in his own words, that the share card might carry the whole debate,
or there might be two cards, or the card might link through to a record of the full
debate. Three options, and he had not chosen. The first job was to put the choice to
him properly rather than to build one and hope.

Measuring settled it faster than arguing would have. A real debate averages about
3,100 characters. A shareable card is 1200×630 and holds roughly 700 of them
legibly. Two cards carry about 46% of one round. So options one and two were both
*excerpting* strategies wearing different clothes, and only the third actually
carries the argument. He chose the third, staged via the first: ship the better
card now, then give it somewhere to point.

The card shipped this morning — it shows the exchange, what Vonc asked and what the
visitor answered, sized to fit the actual round rather than truncated to fit the
type. Then the back end for the record: two endpoints on the island, a slug per
published round, and a gate that means nothing is readable until the person who
wrote it presses share. That went live at lunchtime and was proven in both
directions, including the direction that matters — that an *un*published round is
not readable.

## What we've done

Today's second half closed the loop. There is now a page at
`vonc.com/tools/gauntlet/round.html?r=<slug>` that shows one whole debate: the
provocation, the position filed against it, Vonc's counter and its challenge, the
defence, and the ruling with its reasons. And the share button now publishes the
round and prints that address on the card, so the picture that travels is a route
back to the argument rather than a dead end.

The button says what it does. It reads "Publish this round and save the card", and
above it a line names exactly what becomes public and says that nobody is named and
no account exists. The owner ruled that the press is the consent; that only works if
the press is informed.

All of it is driven end to end rather than inspected: a browser plays a real round,
presses the real button, and the harness follows the address on the card to the page
and checks that the words there are the words that were typed. Fifty-eight checks
across the two drivers, including the refusals — no address, a malformed address,
and an address that was never published.

## Where we are now

Both halves are live and the loop is closed. Three rounds are published, all of them
ours from testing; no stranger's writing is public.

Two things went wrong today that are worth saying out loud, because both were caught
by measuring rather than reading, and one of them I had already told the owner as if
it were a fact.

I said the share card was off-brand — purple against a red site. It was not. I had
read the *fallback* colour out of the stylesheet, the one that only applies if the
real colour is missing, and it never is. The site is that purple. The card has
matched all along. Nothing but asking the browser could have caught it, because the
wrong value is genuinely there in the source and genuinely never used.

The second was a real defect: the first version of the record page put the visitor's
own argument at 2.06:1 contrast — present, correct, and hard to read — because it
inherited a colour instead of stating one. Every static check passed it. There is now
a contrast measurement in the browser that fails anything under the accessibility
floor, and it immediately found five more, including that the site's own accent
colour is too faint for small text on the site's own background.

## Where we're going

Nothing is blocked. Two open items, both the owner's call and neither urgent.

The rate limit on the record service is one request a second per visitor, shared
with the debate engine — a ceiling chosen for slow paid AI calls, not for reading a
page. A six-request test tripped it. Retuning a shared limiter for a route that has
never served real traffic would be guessing, so it is recorded rather than changed.

And when someone shares the *link* rather than the picture, the social preview will
be generic, because the page assembles itself in the reader's browser and preview
crawlers do not wait. In practice people share the picture, which is its own
preview. Worth knowing before it is reported as a bug.

Beyond that, the next real question is the distribution experiment the owner already
ruled on: he posts the daily provocation and the card where people argue, and real
behaviour decides what this becomes. What changed today is that there is now
something on the other end of the link.
