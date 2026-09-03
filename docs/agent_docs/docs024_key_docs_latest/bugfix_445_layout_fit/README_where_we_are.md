# Where we are — bugs_open/445, the layout library gap

The owner's running log. Plain prose, append-only, newest at the bottom.

---

## 2026-09-03 — what this was, and what it turned out to be

You pointed at three sites you'd just had built — designblog.co.uk, advertise.co.uk and
websitepromotion.co.uk — and said the design was exactly the same as all the others and it
should be different. A previous session traced that to a real gap in the layout library: we
have eighteen hand-built page layouts, and none of them is built for what those three sites
actually are, which is a magazine-style content site whose real offering is a set of
interactive tools. So all three landed on the closest thing available, `magazine-grid`. That
was written up as bug 445, and the proposed fix was to design the missing layout.

That diagnosis was right as far as it went. What I found underneath it is worse, and it is the
reason this took a day rather than an afternoon.

**The system had no way of telling us the library was short.** There is a mechanism for exactly
this — when the layout picker can't find a good match it is supposed to raise a note saying "a
human should look at whether we need a new layout". It has existed since June. In the whole
history of this estate it has fired **twice**, out of sixty-three thousand pieces of work, and
both times it was for the same trivial reason: a site that arrived with no classification data
at all. **It has never once looked at the library and said "we're missing something."**

The reason is a small piece of arithmetic. The picker scores each layout, partly on how many of
the site's own descriptive tags it matches, and partly on bonuses — a bonus for being the right
broad category, a bonus if a word appears in the layout's written description, a bonus for
being light rather than dark. The "we're short" alarm was wired to fire only when the *total*
score came out at zero. But the bonuses get added whether or not a single tag matched. So a
layout that matches **nothing at all** about a site still scores above zero, and the alarm
stays silent.

That is not hypothetical. Four live sites are recorded, by the system itself, as having matched
*zero* tags and simultaneously as a successful library match. webdesign.uk is one of them.

**And the fix for this was written down in April and never built.** The original design document
for this part of the system specifies a "match score" between 0 and 1, and a threshold of 0.5
below which the layout is considered a poor fit. Neither the score nor the threshold was ever
implemented. Not one of the thirty-three sites has that score recorded. With no score to
compare against, the alarm was left wired to the only thing available — total-is-zero — which
isn't a measure of fit at all.

**Then a second session found the half I'd missed, and it is the more embarrassing one.** There
is a step that reads our layout library and hands the list of tags to the AI that classifies
each new site, so it can describe sites in words our layouts actually understand. That step
runs. It fetches the list. And then the list is thrown away one line later, because of a
missing entry in a permissions list — so what the AI has actually been shown, every time, is
the word `null` where the list should be, followed by an instruction to match against it, and
then: "if nothing fits, invent a new tag — an unmatched tag will trigger a review rather than
silently failing."

So the AI was told to invent tags against an empty list, on a promise the code could not keep.
It obeyed perfectly. Eighty-seven per cent of the words it has invented match nothing in our
library, and **four strings** end up deciding which layout almost every site gets. Seven sites
share a single one of them, and it describes between seven and ten per cent of what those sites
say about themselves.

Two separate broken links in one loop. The first produced the mess; the second is why nobody
saw it — because "most of our tags match nothing" is precisely what the silent alarm existed to
report.

## 2026-09-03 (later) — what is fixed, and the part I want to be honest about

Three things have landed today, two of them mine.

The other session fixed the thrown-away tag list; that went live at 11:39. I fixed the alarm:
it now measures what fraction of a site's own description the chosen layout actually addresses,
records that number permanently, and raises the note when it is poor rather than only when it
is zero. That is committed and will be in the next build of the system. I also corrected the
instructions given to the classifier so they no longer promise something that doesn't happen,
no longer offer our own layout names as examples of tags (which is why twelve sites have a
layout's name sitting in their tag list doing nothing), and now ask for words describing a
site's *shape* rather than its industry — which matters, because nine of our eighteen layouts
are tagged with industry words no site ever uses, and are therefore unreachable.

**Deliberately, I changed nothing about which layout any site gets.** Only what gets measured
and recorded. Nobody's site moves.

**The part I want to flag rather than bury.** You asked, in effect, for the missing layout to be
designed, and an owner ruling today put that on this thread too. Before recommending a design I
simulated it — added a hypothetical "content hub with tools" layout to a faithful copy of the
scoring system and re-ran all thirty-three sites. It does break the sameness up: four of the
seven move onto it and get a genuinely better fit. **But two of them, including designblog.co.uk
— the site you actually complained about — end up no better off.** For those two the new layout
is a different near-miss rather than a fit.

I'd rather tell you that now than deliver a layout and let you find it. It also happens to be
the argument for the work I did first: after the library grows, we still need to be able to see
who remains badly served, and until today we could not.

The archetype itself is still owed, and I'd want to choose its tags by simulation against the
sites you haven't built yet as well as the ones you have — the other lane has given me the
seventeen remaining remakes for exactly that. That is the next piece of work.

## 2026-09-03 (evening) — the missing layout is built and live

You said go ahead with the archetype. It is done and live — but nobody's site has moved onto
it, deliberately, because you chose "fix forward only" earlier today.

The new layout is called `content-hub-tools`. It is what those three sites actually are: a
reading-first content site whose real product is a set of interactive tools. Structurally it
is a single editorial column, narrower than either of the two layouts it sits between, with
full-width "tool shelves" that interrupt the reading flow — the tools live inside the
publication rather than on a separate index page or in a sidebar — and a way to drop one tool
straight into the middle of an article at reading width.

**Its tags were chosen by simulation, not taste.** I tried four candidate tag sets against every
built site with the scorer that reproduces the system's own numbers. The first — just the
words from the bug's title — fixed four of the seven sites and left designblog.co.uk, the one
you complained about, no better off. The one I used adds the words those sites already use
about themselves ("editorial guides", "long-form content", "research publication", "content
platform"), and it fixes six of the seven. designblog goes from 7% to 16% of its identity
addressed; apis.uk from 9% to 19%. Two other sites would also move to it — a design-practice
publication, and a farm insurance guide that was sitting on a directory layout matching
literally nothing about itself — and I checked both rather than just letting the simulation
have them. One site, oufe.com, does not move under any tag set; its own description leads with
"interactive platform", so the tools layout keeps it, which is fair.

**The migration refuses to create a layout no site can reach.** Nine of our eighteen existing
layouts are unreachable — nobody's tags match theirs — and the guard I built in makes it
impossible to add a tenth by accident. Fourteen current sites can reach this one.

**What happens next, without anyone doing anything:** the next site to be built and classified
will see these tags in its list (that list was invisible to the classifier until this morning),
and if it is this shape, the matcher will pick this layout. The other lane's copyonline.co.uk
is the first candidate and they have promised to tell me what it does.

**What is still owed:** the fleet-wide sweep that checks every site's fit daily (it needs the
shared cron harness first, as you decided), the reusable "is this layout reachable" guard as a
function rather than a block inside one migration, and reading the two council verdicts.
