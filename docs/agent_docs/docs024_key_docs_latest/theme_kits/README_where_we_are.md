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
