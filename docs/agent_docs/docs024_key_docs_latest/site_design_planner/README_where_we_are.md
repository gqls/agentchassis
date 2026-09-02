# Where we are — site design planner

The owner's running log. Plain prose, append-only, newest at the bottom.

---

## 2026-09-02 — picking this thread up

You renamed this session "site design planner" and asked me to pick up that
thread if it already exists, or take it on myself if it doesn't. I looked — there
wasn't one. No other session has this as its dedicated territory, and there's no
folder of notes anywhere already tracking it as its own piece of work.

So here's what "site design planner" actually is, in plain terms, and where it
stands.

When a site gets built, its visual identity — which layout it uses, what colours,
what fonts — isn't picked by one big AI decision. It's split into stages on
purpose. First something reads the site's brief and works out broad style
signals (dark or light, formal or playful, that kind of thing). Then a small,
deliberately non-AI component called **site-design-planner** takes those signals
and matches them against a curated library — about eighteen hand-built layouts,
a set of colour palettes, a set of type systems — and picks the best fit,
weighted so a more specific match beats a vague one. It writes that choice down.
Only after that does a separate AI step actually paint the page with the chosen
palette and generate the stylesheet.

The reason it's split this way, rather than one AI agent doing everything, is
reliability: a fixed library plus a deterministic matcher can't go wrong the way
a free-form AI generation can, and if nothing in the library fits well, the
system is built to say so honestly (a "this needs a human to look" flag) rather
than force a bad match silently.

This piece of the platform is mature — it's been running since April and been
through several rounds of real fixes. The most recent significant one, closed
about three weeks ago, was a genuinely nasty bug: on dark-themed sites, several
colour slots the palette didn't explicitly set were silently falling back to
values written for light sites, so on some pages text became almost invisible
against its background — one page had text at a contrast ratio of about 1.2:1,
which is essentially unreadable. That's fixed everywhere now, verified on the
actual served pages, not just in the code.

Right now there are only three small loose threads sitting in this system's
queue, and I want to be upfront that two of them aren't really mine to touch:
one belongs to the loancalculator.co.uk site, which already has its own active
session working it, and I'd just be stepping on their toes. Another looked like
a live stuck item on adversecreditmortgage.co.uk, but when I dug in, it turned
out to be a stale leftover — that site actually got its design fixed through a
different path a week later, and the leftover queue entry is one of about 230
unresolved items on that site left over from an unrelated billing outage a
couple of weeks ago, nothing to do with design at all.

The one real open question is about ai-agent-orchestration.com: back in mid-
August, when the system tried to pick a layout for it, its classification data
was empty, so it fell back to a generic default layout rather than picking
something tailored — and it correctly flagged that as "needs a human to decide"
instead of guessing. Nobody's looked at that flag since. I haven't investigated
it yet either — that's the natural next thing to do if you want me to keep
going on this.

Nothing has been changed yet. I wanted to get my bearings and write this down
before touching anything, especially since a couple of other sessions are
working closely-related pieces (there's a "build-site-planner" — a different,
similarly-named component — being worked by another thread right now; I checked
in with them so we don't collide).

Let me know if you want me to chase the ai-agent-orchestration.com layout
question next, or if you had something more specific in mind for this thread.
