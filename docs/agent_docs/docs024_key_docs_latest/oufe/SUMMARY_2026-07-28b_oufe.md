# SUMMARY — oufe.com, 28 July 2026 (evening)

Second summary of the day, and a separate file because the read-out genuinely
changed: the first chart went live, the owner looked at it, and what came back
was more useful than the chart.

Written to be read aloud.

---

## What we're trying to do

Build a specialist publication about how corporate finance works when a company
is in trouble — restructuring, distressed debt, who gets paid and who does not.
It is written for people who work in or around that field and are short of time.
The thing that makes it different from a research note is that you can move the
assumptions yourself and watch the answer change.

Everything about how it is built serves one constraint: the site is assembled
with a lot of machine assistance, and machines invent things fluently. So a
figure only reaches a page if it is registered with a source, and the site says
openly that it can be wrong. On a site about named real companies, that honesty
is not modesty. It is the product.

## Where we've come from

Earlier today the site got the things it had been missing. The legal pages that
had been approved but never actually published went live, so there is now a
correction route a reader can reach and a privacy notice behind the contact form.
The tool was properly tested in a browser for the first time and passed. The
flagship Thames Water page stopped being one long block of text and gained a
diagram of how a restructuring plan actually forces a dissenting class to accept
it.

And we built the machinery for charts: a way for the evidence register to hold
figures that carry the date they apply to, and a renderer that draws them.

## What we've done since

**The first real chart is on the Thames Water page.** It shows the drawn debt by
class as it stood at the last financial year end before the plan was sanctioned —
Class A at £14.7 billion, Class B at £1.4 billion, and a further £1.7 billion of
hedging exposure alongside them. It sits directly under the tool, so the
illustrative sliders now have the real shape of the thing beside them.

Getting there produced a more interesting result than the chart. The task was to
register a *series* — something moving over time. The verified figures do not
form one. Every debt figure sits at a single date, and the two percentage figures
we could find measure different things: one is the average bill increase, the
other is an increase above inflation. Plotting those together would have drawn a
trend line through quantities that are not comparable. It would have looked
authoritative and been false. So we published what the data supports and left the
time-series renderer unused.

**Then the owner looked at it and it was unreadable.** Not subtly — the section
heading was invisible against the background, and the chart's own title was light
text on a white card. Three failures, and our contrast checker had reported the
page clean an hour earlier.

The checker was measuring only links and buttons. Everything else on the page —
headings, labels, captions, the chart's own text — was never looked at, and the
tool reported that as *passing* rather than as *not checked*. That is now fixed:
it measures every element that renders text.

**Both underlying faults were ones we had already written down.** The bug file
from yesterday records that this site's primary colour is identical to its
surface colour, and describes the white card as a bug "waiting to happen". We
then shipped a component that draws a card, on that palette, and did not connect
the two. Writing a hazard down is not the same as checking for it.

**And fixing it broke something else that the numbers still could not see.**
Making the card dark fixed the text and made two of the three bars invisible,
because the default bar colour was now exactly the card colour. No contrast tool
caught that, because bars contain no text. It was found by taking a screenshot of
our own fix — which is the habit that should have caught the original problem an
hour earlier.

**The 403 problem is solved.** The Ofwat and Parliament sources that refused us
earlier were not asking for an account. They block simple automated fetchers but
serve the page normally to a real browser. We already had a headless browser for
the contrast checking, and pointing it at Ofwat returns the page cleanly. So the
data we wanted is reachable after all, and the time-series chart has somewhere to
get its numbers.

## Where we are now

Eight pages live. The claims scanner reports nothing unsourced. The chart is
readable, the bars are visible, the footer statement runs the full width instead
of hiding in a column, and every one of those was confirmed by looking at a
picture of the page rather than by trusting a number.

The site now has a working pattern that did not exist this morning: a real figure,
registered with the sentence it came from and the date we read it, drawn as a
chart that cannot contain a number we did not verify.

What is still owed is volume. One chart and one diagram on one case is a
demonstration, not a publication. The owner is right that we need many more of
both, and more tools.

## Where we're going

Next is the Ofwat data, now that we can reach it: the price determinations are a
genuine series — the allowed revenue moving across a review — and that is the
first honest use for the time-series chart.

After that, the bigger idea: take a piece of writing, pull out the premises it
actually rests on, and give the strongest one or two their own page with their own
graph and tool. That is already in the plan from earlier in the month, and the
advice we recorded is to hand-build the first one before automating anything,
because the expensive judgement is choosing *which* premise deserves a page.

The lesson from today is small and awkward. Almost everything that went wrong was
something we had already written down and then not checked — a warning in a
handoff, a hazard in a bug file, a rule in a validator nothing called, a checker
whose scope we chose ourselves. **A measuring instrument you built yourself is the
hardest one to distrust, because when it agrees with you it may only be agreeing
with your blind spot.** The fix is dull and reliable: look at the thing, not at
the number that says the thing is fine.
