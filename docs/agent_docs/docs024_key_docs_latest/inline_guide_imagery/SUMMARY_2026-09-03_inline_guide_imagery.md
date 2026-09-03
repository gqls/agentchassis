# SUMMARY — inline guide imagery, 2026-09-03

First summary for this lane. Written at the milestone where the mechanism was proved working
end to end on a live page, which is also the moment it became clear that the owner's ask has a
second half this lane does not own.

---

## What we are trying to do

The owner asked, on 2026-08-13 and again on 2026-08-31, that the guide articles carry explanatory
pictures **inside** the body of the article rather than only a banner at the top. He named the
example himself: the grip-styles guide on the darts site should show what a ring grip, a razor grip
and a shark grip actually look like, next to the paragraphs describing them.

The lane reframed that as a durability problem rather than a content one. Putting a picture into an
article body is easy and it does not last. An article's prose lives in one field owned by a language
model, and the next time anything rewrites that field the picture goes with it. So the real
requirement is a figure that is **attached to a section rather than written into it**, and that
re-attaches itself every time the page is rebuilt.

## Where we have come from

The starting position was that a page could carry only one picture per *kind*. If five sections on
a page all asked for an illustration, all five resolved to the same image. The plan could already
record which section a figure belonged to — there is a field for it — but nothing ever read that
field. It was used to filter rows and then discarded.

Three other things had to be understood before the fix could be safe. The estate carries two
incompatible ways of numbering a page's sections, so a figure bound in one scheme and read in the
other lands on the wrong section — and a wrongly-bound figure renders and deploys looking perfectly
correct. There are two separate code paths that render a page, and a fix applied to one of them
would have overwritten the other's work. And a person can pin a section onto a live page that the
plan does not know about, which shifts everything after it.

## What we have done

The binding was built, reviewed and shipped. A section-scope imagery row now binds to the one
section its ordinal names. The ordinal is translated once, against the plan that minted it, into an
identity of the form "this slot, this repeat" — never a position number — and both render paths then
count repeats using the estate's existing shared counter. Where the plan's section order and the
live page's order disagree, the whole binding **stands down** rather than guessing, because a figure
bound one section out is indistinguishable from a correct one.

It went through the review council three times. Every one of those rounds found a real defect that
would otherwise have shipped: the ordinal's identity had been computed by three separately-written
pieces of arithmetic; a hand-rolled counter sat two lines below the shared one it should have used;
and one round answered an objection about the single file it named instead of the category it meant.
The change carries an approved verdict at round three, twelve seats.

Along the way the lane measured the thing everyone assumed and nobody had checked: what actually
composes an article. The answer reframed the whole ask. Almost no article page in the estate is
built out of small illustrated sections, and most article pages have no entry in their site's plan
at all. The capability was never the bottleneck; the composition was.

## Where we are now

**The mechanism is proved, on the owner's own page, at the served bytes.** The grip-styles guide was
recomposed into eleven sections by the darts thread this afternoon, five of them illustrated, and
each of the five resolved its own correct photograph in the right place. All five load. All five are
anatomically right, checked by eye.

The decisive test then happened by accident and it passed. An hour after the page was built, the
last image finished generating, and that automatically triggered a full rewrite of the page — every
heading and paragraph replaced. **The five photographs stayed exactly where they were.** That is the
durability property the lane exists for, observed rather than argued: a rewrite of the kind that
used to destroy an inline figure now costs nothing.

**And the same page shows that the words beside those pictures are wrong.** On the first build all
five sections were written about the ring grip, under five different and correct photographs, so the
section showing a smooth polished barrel carried a heading about bands of cuts. The rewrite replaced
that with five near-identical headings, none naming the grip it sits next to. The hidden text that
screen readers announce is wrong in the same way.

The cause is measured and is not this lane's code. The framework hands the writer a note saying what
each specific section is about, and those notes are correct and distinct and do reach the writer.
The writer's instructions never include them. So it receives five requests that look identical and
writes the same section five times. Another thread found this independently, predicted in writing
that this is what would happen, and the fix is a single prompt change awaiting the owner's reading.
What this page adds is the severity: their examples were pages repeating a paragraph, which is dull
but misleads nobody. Here the framework got its half right, so identical words became false captions
on correct pictures — and that gets worse, not better, as this lane succeeds.

## Where we are going

The lane's own remaining work is small and honest. One render path — the re-render path, as opposed
to the rebuild path — has still never been exercised on a page carrying several figures, and it is a
genuinely separate arm of the safety check. That is the next thing to watch rather than to build.

The ask as a whole now has three layers and this lane owns only the top one. Attaching a picture to
a section is done, reviewed and live. Composing an article out of small illustrated sections has
just been done by hand for the first time, on grip-styles, and belongs to the editorial thread.
Getting articles into the site plan at all — the layer that makes the other two reachable — is
nobody's today, and it should not be closed by backfilling plan rows by hand, because that repairs
the pages that exist and changes nothing about how the next article is born.

The one thing this lane will keep pressing is the coupling. Per-section pictures are worth exactly
as much as per-section subjects. One of those two shipped, and the gap between them is visible on
the page the owner asked about.

---

> ## ⚠ CORRECTION APPENDED 2026-09-03, EVENING — the proving page has been REVERTED
>
> **The body of this summary is left exactly as written**, because the series is the record of what
> was believed at each milestone and overwriting it would destroy that. This footer is appended
> rather than woven in, for the same reason. But do not read the sections above as live state.
>
> `dartsonline_traffic` reverted `grip-styles` hours after this was written: **3** plan sections,
> **0** section-scope imagery rows, `page_components` back to hero/article-body/call-to-action. The
> five illustration assets remain `active`, so it is re-runnable in minutes. **Their call, and the
> right one** — that lane wins search traffic for affiliate approval, and seven near-identical
> sections work against it. Their words: *"the imagery was the only part that behaved."*
>
> **What still stands:** the binding was proved on a real page and read at the artefact, and a
> revert does not retract a measurement. **What is now stale:** every present-tense sentence above
> about the served page. **What is void:** this lane's pre-registered prediction about the re-render
> path, which can no longer be resolved on that page. **apis.uk is again the only armed page.**
>
> **RESOLVED the same evening, in this lane's favour and confirmed from both sides — see NOTES §20.**
> That lane withdrew the objection, having found the fault in their own instrument (an `llm_call_log`
> query bounded to the window where they expected the build to end, so the second writer run fell
> outside it), and confirmed at their own 16:49Z page snapshot: five distinct illustrations, one per
> section. **The durability property is proven, not merely the binding.** The one clause that
> survives is that no item of type `content_rewrite` was ever fired, so that dispatch path is
> untested; the register now says the phrase names the event, not the item type. The paragraph below
> is left as written, because it is what was believed when the disagreement was open.
>
> **One claim above is disputed by that lane.** They hold that the durability property — figures
> surviving a rewrite of the prose — remains unexercised, because they reverted before anything
> rewrote over the figures. This lane's measurements say a second full writer run rewrote every
> section's prose 69 minutes after the build and re-derived the figures from the plan rather than
> carrying them forward. The evidence has been put to them; it is recorded as an open disagreement
> rather than settled here. **NOTES §19b** has both positions and the numbers.
>
> **And their instrument was better than this lane's** for the writer defect: `llm_call_log.prompt_rendered`
> holds the prompt actually sent, and one byte-identical prompt went to four of the five
> same-component sections. This summary's account of that defect is correct and was reached the
> weaker way. Logged in `WRONG_CALLS.md`.
