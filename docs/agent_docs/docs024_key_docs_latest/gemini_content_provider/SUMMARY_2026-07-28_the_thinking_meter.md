# SUMMARY — we fitted a meter to the part of the bill nobody could see (2026-07-28)

*Written to be read aloud. Figures measured 2026-07-28 on chassis v1.0.1182; the
queries that produced every number are in `RUNBOOK_gemini_content_provider.md` §10.*

## What we're trying to do

Our sites are written by a machine. A model is handed a brief and the surrounding
context of a page, and it writes the actual sentences a visitor reads. For most of this
year that model has been Claude. In July we tried moving the two content-writing agents
to Google's Gemini, partly for quality and partly so we aren't dependent on one
supplier. The aim of this workstream was to find out whether that move works, what it
costs, and whether the copy is any good.

## Where we've come from

The first attempt, on 24 July, was reversed within the hour. Gemini had been asked to
write a tweet-length piece and returned **nothing at all** — zero characters. Longer
pieces came back stunted. The obvious reading was that the model couldn't do the job,
and on that evidence the switch was undone and everything went back to Claude.

**The obvious reading was wrong, and the fault was ours.** Newer models "think" before
they answer: they work through the problem in tokens you never see. Anthropic bills that
thinking separately. Google does not — it takes the thinking out of the *same* budget
you set for the answer. Our code had been written for Anthropic and passed the same
number straight through. So when we told Gemini "you may use 100 tokens", it spent 92 of
them thinking and had 8 left to write with. Zero characters wasn't incapacity. It was
arithmetic, and it was our arithmetic.

## What we've done

We fixed the arithmetic — the answer now gets its own budget, with headroom reserved on
top for thinking — and put it through the reviewer council, which approved it. Both
content agents now run on Gemini in production, and the writer has produced real pages
on live sites.

Then we found the thing this summary is really about. Our logging recorded what we asked
for and what came back, **but not what the thinking cost** — the client calculated it and
threw it away. So the single biggest driver of the bill was invisible to every query,
dashboard and report we have. We added four columns to record it, and the council
approved that too, unanimously, on the third attempt.

Along the way we got two things wrong and caught them. We claimed a live page was broken
when the fault was in how we were looking at it. And in the very change that fixed "one
column meaning two different things", **we introduced a column whose name meant something
other than what it held** — caught by reading the first four rows rather than trusting
our own naming.

## Where we are now

The meter is fitted and reading. Eleven real calls from the page writer, measured today:

| | tokens |
|---|---|
| the prompt we send | 57,983 |
| **visible copy the model wrote** | **2,439** |
| **thinking, which we also pay for** | **20,826** |
| whole-call total | 81,248 |

**For every word of copy we keep, the model wrote roughly eight and a half words'
worth of thinking that nobody will ever read** — and we pay for those too. Thinking is
**89.5% of the billable output**, and it ranged from 1,163 to 2,901 tokens on a single
page section.

**One number needs its neighbour or it misleads.** That 89.5% is a share of *output*
tokens only. Counting everything sent and received, thinking is **25.6%** of the tokens
in play, because the prompt is much the larger part at **71.4%**. Both figures are true
and they answer different questions. Quoted alone, 89.5% invites the reading "almost all
our spend is wasted on thinking", which is not what we measured and not what we believe.

We have also deliberately **not** converted any of this into pounds. Rates move, one of
our providers is on an introductory price that expires at the end of August, and input
and output tokens are priced differently. The token ratio is measured. A money figure
would be an inference wearing a measurement's clothes.

## Where we're going

The question this was built to answer is now answerable and remains open: is Gemini
worth it for the writer? An earlier bake-off put Gemini at roughly ten times a rival's
output tokens per section while costing less per token — which is exactly why guessing
was hopeless and measuring was necessary. That's a commercial decision, and it is the
owner's.

Two smaller things stand. The writer generates each section without sight of its
siblings, so a page's opening and closing blocks sometimes make the same point twice.
And on one page the product list was dropped for want of data, leaving copy that
advertises a sale with nothing underneath it to buy. Neither is about Gemini; both are
about how pages are assembled.

The wider lesson is the one worth carrying: **we spent a day in July concluding a model
couldn't write, when what had actually happened was that we'd starved it and had no
instrument to see that.** The fix was small. Not being able to see was the expensive
part.
