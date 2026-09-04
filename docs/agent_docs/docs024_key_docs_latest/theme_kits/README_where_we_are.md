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

---

## 2026-09-03, last entry — the review found something real, and it is the thing you should know about

The internal review council came back a second time asking for changes, and this time it
was not about how I had written the submission. **It found a genuine fault in the running
code, and it is the fault that matters most.**

**A look applied to a brand-new site gets thrown away.**

Here is the mechanism in plain terms. When a new site is submitted, a step runs that reads
the brief and writes down the site's design intentions. If we have applied one of our looks
to that site beforehand, that step overwrites it. Not partially, not with a warning — it
simply replaces the values and the look's fonts and colours are gone. The record of what we
asked for is still in the database as a superseded row, so nothing appears broken, and the
site that comes out is plausible because the step read the same brief we did.

**Which parts survive:** the page layout does, because it is stored somewhere that step
does not touch. The colours do not, but as I explained above colours never reached the page
anyway, so nothing changes there. **The fonts do not survive, and fonts are the one thing a
look was actually delivering.**

**So the fault is precisely the inverse of what you asked for.** You said a site should be
able to start from a look and change it later. What we built works if the site has already
been through the pipeline, and quietly fails if it hasn't. Applying a look to a finished
site is fine. Applying it to a new one — the case you described — loses the fonts.

There is also a trap I should flag because I built it. I added a way for a person to pin a
value so nothing overwrites it. **That pin does not help here.** It is honoured by the piece
of code I wrote and ignored by the step that does the overwriting, and worse, the pin marker
itself survives while the values do not — so afterwards the record says "a human chose
this" sitting on top of values a human did not choose. That is a worse state than having no
pin at all, and it is now written down in three places so nobody trusts it.

**I have not fixed it, and I want to be straight about why.** There are three ways to fix
it. One changes what that classification step is allowed to overwrite, which affects every
site we build and is not a change I should make on my own authority at the end of an
afternoon. One works around it by writing the fonts somewhere else that happens to survive
— but that only survives by accident, because of a separate bug we already have open, so it
would be building on a defect. The third, and the one I would choose, is simply to refuse
to apply a look to a site that hasn't been classified yet, and say so plainly instead of
silently losing the values. That one is small and honest, and it still changes the behaviour
of something live, so it should go through the same review rather than being slipped in
today.

It is all written up where the next session will find it, including which fix I recommend.

**Nothing is on fire.** No site has a look applied, so nobody is affected today. But it does
change my answer to the question I asked you earlier about whether a "look" is the right
idea at all — because with this, three of the four things a look bundles don't work on a new
site, and the fourth is the page layout, which we can already choose without any of this
machinery.

---

## 2026-09-04 — you asked about the layouts we built and never used

Short answer: **the layouts are the most solid thing in this whole area, they are not
superseded, and about half of them have never been used. But there is a second thing
buried in them that was designed, half-built and then forgotten, and it explains a
problem we have been circling for days.**

### The layouts themselves

We have nineteen. All of them are switched on. Ten are in use and nine have never been
used by any site.

I checked whether they are actually different from each other, because a library of
near-duplicates would not be worth much. They are genuinely different: nineteen distinct
stylesheets, nineteen distinct sets of structural settings, each between fourteen and
thirty kilobytes of real CSS. Somebody did the work properly.

More importantly, **a layout is the one design decision that survives all the way to the
page.** Colours get overwritten by the design step, as I explained the other day. Page
shapes barely apply. But the code is explicit that the layout always wins on structure and
the design step cannot override it. So when a site gets a layout, it really does get that
layout.

**So no, layouts are not superseded, and they are not the thing that is broken.** They are
the one piece of this machinery that does what it says.

### Why half of them sit unused

Another team already found the main reason and it is not ours to redo: the layouts are
tagged with industry words like "bakery", "law", "artisan", while the step that classifies
a new site emits words about what a site *does* — "publication", "tool portal". So the
matching never fires for the industry-tagged ones. They are unreachable rather than
unwanted. That team is building a proper scoring tool for it.

### The thing I found, which nobody had recorded

Back in April, when this system was designed, the plan said something sensible that I had
not seen anyone mention since:

> *"Different layouts often need structurally different headers and footers: a
> comparison-aggregator needs a header with a prominent search input; an
> ecommerce-storefront needs a header with a cart icon; a docs-sidebar needs a fixed left
> nav. These are structural differences, not stylistic variations."*

So they gave each layout two fields: its own preferred header and its own preferred
footer. The idea was that choosing a layout would also bring the right furniture with it.

**Those two fields are empty on all nineteen layouts, and no code anywhere reads them.**
They were created in April and never connected to anything. The design document even says
why they were left empty at the time — the headers themselves had not been built yet, and
they would be filled in "when the components land".

**The components landed.** Five of them are alive and usable right now, including the
search header, the cart header and the disclaimer footer. Nobody ever went back and joined
them up.

### Why that matters, and it is not just tidiness

Three of the four layouts that design named a special header for are in the never-used
group. That is not the reason they are unused — the tagging problem above is. But it means
**fixing the tagging alone would put sites onto layouts that then render with the generic
header.** You would get the right shape with the wrong furniture, and it would look like
the layout was a poor choice when the real gap is one field nobody filled in.

It also explains the chrome puzzle I described to you: every site has an identical header
partly because the mechanism designed to vary it by layout was never wired.

### What I have not done, and why

I have not wired it up. Filling in nineteen rows is trivial; the part that would take
judgement is that nothing reads the field, so somebody has to decide where layout-chosen
furniture sits relative to the two other places we can already express the same choice. We
would be adding a third answer to one question. That is a design decision, not a task, and
it is adjacent to a question you already have in front of you about whether "kits" are the
right idea at all.

My honest read: **the layout is the natural place for this.** The April argument is right
— a documentation layout needs a sidebar nav because it is a documentation layout, not
because of who the client is. That is a property of the archetype. If we ever consolidate
the three ways of choosing a header, I would consolidate onto the layout.

I have written the full account into the layout team's file so it reaches the people
making that decision, and corrected our internal reference where it described the field as
though it worked.
