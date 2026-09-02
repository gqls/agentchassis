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

## 2026-09-02 (later the same day) — found the "why", and fixed it

I went to answer my own question about ai-agent-orchestration.com and found the
actual cause, not just the symptom.

The system that picks a layout for a new site is supposed to read a site's
industry information from two places: a quick automated classification, and a
more detailed "identity" profile written earlier in the process. Three of the
four related decision-makers in this system (the ones that pick colours, pick
fonts, and install the finished result) already knew how to fall back to the
detailed profile if the quick classification came up empty. The one that picks
the actual layout — the most consequential of the four — didn't have that
fallback. It only ever looked at the quick classification, and if that came up
empty, it gave up and used the generic default, which is exactly what happened
to ai-agent-orchestration.com in mid-August.

I checked how common this actually is: right now there are four sites in this
condition, two from an old, no-longer-used version of the classifier, and two
(including ai-agent-orchestration.com) whose classification data was just
refreshed again this morning by an unrelated background process, in the same
incomplete shape. Not a fleet-wide emergency, but a real, reproducible bug with
a name and four known instances.

The fix is small and low-risk: make the layout-picker use the same fallback the
other three already use, instead of its own thinner copy. I wrote it, added a
test that reproduces ai-agent-orchestration.com's exact situation and would have
failed before the fix, checked it against the platform's own automated review
process, and wrote up the full case as its own file so nobody has to re-find
this later.

One thing worth mentioning because it's the kind of thing this shared codebase
throws up: while I was mid-edit, another session working on an unrelated feature
(theme presets) happened to be editing the very same function, right next to my
change. I noticed because the file changed under me while I was reading it. I
didn't undo anything — I checked that our two changes were compatible (they
were), flagged to them directly that their half wasn't committed yet so the
whole codebase wouldn't build, and they fixed that within a few minutes. No harm
done, but it's a decent illustration of what working on a codebase several
sessions share at once actually looks like day to day.

I have not triggered anything on the actual affected sites — that's still each
site's own call, not something to do from here. The fix just means that when
someone does decide to re-look at ai-agent-orchestration.com's layout, the
system will finally have the information it needed to make a real choice instead
of just giving up.
