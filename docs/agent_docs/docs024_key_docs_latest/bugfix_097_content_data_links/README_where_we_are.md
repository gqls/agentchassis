# Where we are — the broken card links (bug 097)

Plain prose, append-only, newest at the bottom. No jargon where I can avoid it.

---

## 2026-08-02, morning — what the bug was, and what was left of it

Back in July somebody noticed that oufe.com had gone live with six links on its
home page that led nowhere, and one of the buttons in the site's own top menu
gave a "page not found". A bug was written up. Since then, other threads did a lot
of good work on it: today, when one of our pages is built or rebuilt, we go
through the finished page, check every link against the list of pages that really
exist, and either correct it or turn it back into ordinary text. So the pages we
publish are clean.

What nobody had done is the other half, and it is the half the bug report itself
kept pointing at. Fixing the finished page is like proofreading the printout. The
**manuscript** — the stored data each section is rebuilt from — still has the wrong
link in it. So every time we rebuild that page, the bad link comes back, gets
quietly corrected again on the way out, and nobody is ever told that the section
was written with a link to a page that does not exist.

## Why nothing caught it

Because of how we were asking the question. The system had a hand-written list:
"these six kinds of section have a button, and the button's address is in a field
called one of these two names." Anything else was invisible.

That works until somebody builds a section that holds *several* links — a grid of
cards, say, where each card has its own "read more". Those live one level down,
inside the card, and the list-based check simply cannot reach them. It is the same
story for a section that numbers its cards one to five: each has an address field,
none has the paired label field the newer check looks for, so that check skips
them too.

Adding those two sections to the list would fix those two sections. The next new
one reopens the hole. That is what the bug report meant by ranking the fixes "by
what closes the door".

## What I found when I measured it

I ran the real code over every stored section on every live site — 885 of them.
**52 stored links point at a page that does not exist**, spread over seven sites
and four kinds of section, and **not one of those four was covered by either
check**. Some examples:

- idea.uk's cards point at `/about`, `/report` and `/tools`. Those pages **do**
  exist — they are just at `/about.html`, `/report.html`, `/tools.html`. The link
  was written from what someone meant rather than from the real address.
- gaswholesalers.com has cards pointing at `/pricing`, `/products`, `/delivery`
  and `/eligibility`. Those pages were never built at all.
- Two sites advertise five case studies each that were never written.

So it splits neatly: **19 of the 52 are near misses** where the real page is right
there, and **33 are pages that simply do not exist**.

## What I changed

One pass, added where every section is saved, that looks at the stored data rather
than the finished page. It finds link fields wherever they are — however deeply
buried, whatever they are called — and checks each against the real list of pages.

The trick that makes it general is that it uses two different questions. "Might
this field hold a web address?" is answered by the field's **name**. "Is this
actually a link to one of our own pages?" is answered by the **value**. That
second question is what keeps pictures and links to other websites out of it,
without me having to write down a list of the fields to ignore — a picture's
address ends in `.jpg`, an outside link starts with `https://`, and the existing
shared code already knows that. Running it over the whole fleet, it picked up
nothing it should not have, and left 872 of the 885 sections completely untouched.

Then it does two different things depending on which kind of problem it found:

- **The near misses it corrects**, in the stored data, to the real address. That
  is safe: it can only turn a link that goes nowhere into one that goes somewhere
  real, and it can only use an address the database gave it.
- **The genuinely missing pages it reports and leaves alone.** I want to be
  explicit that this is a choice. I could have emptied those fields, which would
  make the card render as plain text. But that throws away what the writer
  intended, in the one copy where it cannot be recovered — and there is an open
  disagreement between two of our own reviewers about whether we should be doing
  that even on the finished page. Deleting more, in a more permanent place, while
  that argument is unresolved, is not something this fix should decide on its own.

## Two things I got wrong along the way

I wrote the safety checks and then, as I should, tried breaking each one on
purpose to confirm a test would catch it. One of them I could delete entirely and
every test still passed. It turned out I had written **two** things doing the same
job, so each was hiding the other's absence — which means neither was ever really
tested. I deleted the spare rather than leave it in looking useful.

Separately, two of my break-it-on-purpose attempts didn't compile, and the word
"FAIL" in the output made them look like they had proved something. They hadn't. A
broken build is not a failing test. I redid them properly.

## Where this leaves things

The 52 bad links do not need a clean-up operation. Each one is corrected or
recorded the next time its page is saved in the normal course of things, so it
sorts itself out as the sites are worked on.

The change is committed and has gone to our review council. It does not take
effect on the live system until the next time the software is rebuilt and rolled
out, so the bug stays open until then — that is the rule, and it is the right one:
until it is actually running, the problem is still reproducible.
