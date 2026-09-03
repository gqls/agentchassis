# Where we are — theme kits

This is the plain-English running log for this piece of work, newest at the bottom.
It is the owner's document as much as anyone's. **Append to it. Never rewrite or reorder
it, and never edit anyone else's words** — add a dated note below instead.

---

## 2026-09-03 — the first entry, written a day late

This file should have existed on day one and didn't. Everything below is the honest
version of where the theme-kit work actually stands.

**What we were asked for.** A set of named "looks" a new site can start from — colours,
fonts, page shapes, header and footer — which the site is then completely free to change.
You also wanted to be able to make one of these looks out of a site we had already built.

**What got built and is now running.** A registry of four named looks, an action that
stamps one onto a site, and a table of reusable page shapes that replaces a list of page
layouts that had been hardcoded in the program. It all went live on 2 September. The
database side and the program side were both checked properly this time, because the
first attempt at reporting this said "deployed" when neither half was true.

**Nothing uses it.** Not one site has adopted a kit. That is not a fault — it was never
switched on for anything. But it does mean everything below is measured on the code and
the library, not on a site anyone can look at.

**The uncomfortable finding, and it is the important one.** A kit bundles four things:
colours, fonts, page shapes, and the header and footer. **Three of those four cannot
change how a site looks.**

Colours can't reach the page. When a site is rendered, a later step writes the eight main
colours itself and overwrites whatever the kit chose. We proved this on a real site: we
hand-picked a palette, the system recorded it correctly, and the site served none of
those eight colours. That is the system working the way you asked it to work — the design
step has the final say — but it does mean colours are not something a kit can deliver.
The thing that actually moves colour is the brief we write for the site.

Page shapes barely apply. Of the 1,083 live pages, about 94% don't use the default page
shape at all — the page planner chooses for itself. So the new table changes what happens
in fewer than one page in eighteen, and even that is an upper estimate.

The header and footer are a no-op, and I only found this today. All four kits point at
the same header and footer that a site with no kit gets anyway. So adopting a kit changes
nothing there. I had previously reported this area as the most promising lever, which was
wrong twice over — first because I measured the component library instead of the data
behind it, and now because the kits themselves don't even vary it.

**So what does a kit actually do? It picks the page layout.** And here it gets narrower
still. Two of our four kits pick a layout the system would have picked by itself, and one
of those two is the layout it falls back to when it can't decide — so that kit dresses up
the default as a choice. Only one of the four, the soft editorial look, reaches something
the system cannot otherwise reach at all.

**That one is worth keeping, but for an honest reason.** It is the only route to that
look, not because we designed it as a route, but because the way we tag layouts is
broken: layouts tagged with industry words like "bakery" or "law" never get chosen, while
layouts tagged with words describing what a site *does* get chosen constantly. Another
piece of work is building a proper scoring tool for this. Until then I am deliberately
not guessing at which looks should become kits, because picking by taste is how we got
these four.

**What I got wrong, and how it came out.** A review of my submission to our internal
review council came back asking for changes, and it was right: I had claimed a fix in my
write-up that my own summary of the code didn't show, so the reviewer had no way to check
it. While fixing that I re-ran some measurements and found I had also written down two
straightforwardly false facts — I said a particular header component wasn't usable when
it is, and I named two components as the right ones when they don't exist at all. Both had
been sitting in an internal reference document that other reviewers read as fact.

The pattern is the thing to notice. In all three cases my conclusion was right and my
reason for it was false. That is more dangerous than being plainly wrong, because the
right answer stops you checking. Nothing external caught these — a reviewer queried a
different sentence, and re-measuring for that swept them up by accident.

**What I need from you, when you have a moment.** Two questions, neither urgent.

First, there is a dead feature we have both been circling: the ability to specify a
palette when a site is submitted. It doesn't work, and both lanes that touched it agree
it should either be built properly or removed. Building it would not put a colour on a
site by itself, for the reason above, so it only makes sense alongside changing how
rendering merges colours — which is the proposal we withdrew. My honest recommendation is
to remove it.

Second, and bigger: given that three of the four dimensions can't move, is a "kit" the
right idea at all? It may be that all the value here is in making more layouts reachable,
which is a different and cheaper piece of work. I am not proposing we throw this away —
it is built, live and harmless — but I would rather ask now than build a Phase 2 on it.

**One thing that is genuinely fine.** Nothing here can hurt anything. No site is using a
kit, an unused kit does nothing, and every site can still override every value at any
time. The three sites we did touch on 2 September were checked colour-by-colour before
and after and look exactly as they did.

---

## 2026-09-03, later the same afternoon — I have to correct the entry above

The paragraph above says I "named two components as the right ones when they don't exist
at all". **That was wrong, and it is the most instructive thing that happened today.**

The two components exist. What happened is that this table has two columns holding almost
the same set of words — one is the component's name, the other is the job it does — and I
searched the wrong one, found nothing, and concluded the components were fictional. So I
retracted a statement that was correct, and the retraction went into an internal reference
document and into a live review submission before I noticed.

What caught it was not a check. I was about to tell another team that a claim they might
have inherited was false, and grepped the codebase first to see where it had spread. It
had not spread. Instead I found seventy files naming those components, and a database
migration carrying explicit guards against overwriting them. **A thing that does not exist
does not need guards against being overwritten.**

Everything is now corrected forward in the internal documents, with the wrong version left
visible next to the right one, because the mistake is more useful than the fact.

The reason I am writing this to you rather than burying it in the technical log is that it
changes the count and the character of today's errors. It is four, not three. And a
retraction is the expensive direction to be wrong in, because it reads as "someone went and
checked" — so it carries more weight than the original claim, and the next reader stops
there.

**Nothing about the substance changes.** The finding that matters — that all four of our
looks point at the same header and footer a site gets anyway, so that part of a kit changes
nothing — is still true, and I have now double-checked it from two directions. So is
everything above about colours not reaching the page and page shapes barely applying. The
two questions I asked you for a decision on stand exactly as written.

---

## 2026-09-03, end of the session — a third question, and it was nobody's

There is a design question this work keeps running into that no one has actually written
down, and I would rather hand it to you than leave it in a handoff again. It was raised
when we built gamedesign.uk and has been sitting unowned since.

**Should the record of a site's design describe colours the public never sees?**

Here is the shape of it in plain terms. When we compose a site we write down a formal
record of its design — which layout, which palette, which fonts, and where each of those
choices came from. That record is validated: it will reject an unknown value, and it
insists we say how each decision was reached. It is, deliberately, the trustworthy
account of what the site is.

The problem is that for the eight main colours, it isn't true. The rendering step writes
those itself and overwrites whatever the record says. So we maintain a careful, validated,
audited description of colours that never reach anybody's screen. We proved that on a real
site: the record said one thing, the site served another, and both were behaving correctly.

Three ways out, and I am not going to pick one for you.

Leave it, and accept that the record describes intent rather than outcome — which is
defensible, but then something should say so, because today it reads as fact and other
lanes have quoted it as fact.

Stop recording core colours there at all, so the record only claims what it controls.
Cheap and honest, and it loses the ability to ask "what did we intend here?" later.

Make the record win, so the colours in it are the colours served. That is the change we
drafted and then withdrew at your direction, and I am not reopening it. It would put the
authority back where the record is, at the cost of the design step's freedom — and you
have already ruled that the design step should keep that freedom.

**My recommendation is the second one**, with a line in the record saying what it no longer
claims. It is the only option that makes the document honest without taking authority away
from the step you decided should have it.

This does not block anything. Nothing is broken while it stays open. But it is the third
time this question has come up in a week, and each time it costs somebody an afternoon
discovering it for themselves.
