# Where we are — the header button that pointed at a page that wasn't there

Append-only, newest at the bottom. Plain prose.

---

**2026-08-04, morning.**

Picked up bug 191 from the open pile. It was filed by the mortgagecalculator lane
the day before and marked unowned, and nothing in the fleet was working it.

The complaint is small to describe and slightly embarrassing to look at. Our header
has a row of navigation links and, beside them, a call-to-action button. Both are
built in the same run, by the same code, from the same list of pages. But the links
and the button were checked against **different rules** about which pages are
allowed to be linked to. The links got the careful rule: only pages that have
actually been published. The button got a sloppier one: any page we have *planned*,
whether or not it exists yet.

So on mortgagecalculator.co.uk the navigation had quietly shrunk to its one live
page — correctly, because the rest aren't built — while the button next to it
pointed at a stamp-duty page that has never been published. I checked it by hand:
it returns "404 not found". And because this header is copied onto every page of
the site, that's a dead button everywhere, not in one place.

Worse, this header is written **once** and then left alone. There's no process that
comes back later and re-renders it when the missing page finally goes live. So it
doesn't heal.

**What I decided to do about it.**

The tempting fix is one line: make the button use the careful rule. I didn't do
that, because it doesn't explain how the two rules came to disagree in the first
place, and it wouldn't stop the next person doing the same thing.

The real reason is that the careful rule was never available to be reused. It was
written *inside* the navigation code, tangled up with the navigation's own
decisions, so anyone building a different bit of the header couldn't call it even if
they wanted to. They reached for the nearest other thing, which was the rule meant
for page *content* — where being loose is actually correct, because content gets
rewritten constantly and fixes itself.

So I lifted the careful rule out into something with a name — a "chrome link
policy" — and pointed both the navigation and the button at it. Now they can't
disagree, because there's only one of them. And I left a note on the loose rule
saying, in effect, "this one is for page content, and here is why the header must
not use it", with an automated check that fails the build if someone wires a header
link to the loose rule again.

**One judgement call I want to flag, because it's arguable.**

On a brand-new site nothing has been published yet, so the careful rule would reject
*everything*. The navigation already has an escape hatch for this: on a first build
it stops filtering, on the grounds that an empty menu frozen into a new site's
header forever is worse than a temporarily optimistic one.

The bug report suggested the button should do the opposite — show nothing at all on
a first build. I disagreed and gave the button the same escape hatch as the
navigation. The reason: this header is written once and never revisited, so "no
button" on a first build isn't temporary, it's permanent. And having the list and
the button beside it answer the same question differently is *precisely the bug I
was fixing*. I'd rather they be wrong together for ten minutes than right and wrong
side by side forever.

I've written that up for the review council explicitly, as the one thing in this
change they should push back on if they think I've got it wrong.

**Something I nearly got wrong, and it's worth telling.**

To justify that decision I wanted to know how many sites would actually hit the
"brand new, don't filter" escape. The first number I got was **19 out of 38 — half
the fleet**. That would have sunk the argument: it would mean the escape isn't a
rare first-build case at all, it's the normal state.

It was the wrong number, because I'd asked the wrong question. "Sites with nothing
published" includes 18 sites that have **no pages at all** — shells that have never
been built, where the header renders nothing either way and the rule is irrelevant.
Once I separated those out, the real answer was **one site**. One genuinely young
site takes the escape.

The design was fine. But it was fine for a reason I hadn't checked yet, and the
first number I pulled looked exactly like proof that it wasn't. I've logged that in
the wrong-calls file, because the mistake wasn't the query — it was writing a filter
straight off the question in my head without asking what else it would sweep up.

**Where it stands.**

The code is written, tested and committed. Every safety check in it has been proved
by deliberately breaking the code and watching the test go red — a passing test
proves nothing unless you've seen it fail. It's gone to the review council; I
committed without waiting for the verdict, which is the house rule, using the
trailer that makes no claim about the outcome.

It is **not live yet**. Go changes only take effect when a new image is built and
rolled out, and that happens on someone else's schedule. Until then the fix is real
but dormant, so I've left a short list of checks for whoever is around after the
next roll: confirm the new code is actually in the running pods, re-run the header
builder on mortgagecalculator, and curl the button. Only then is the bug genuinely
dead.
