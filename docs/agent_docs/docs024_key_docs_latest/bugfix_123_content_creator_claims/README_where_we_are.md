# Where we are — content-creator and the claims checker (bugs_open/123)

Plain-prose log, append-only, newest at the bottom. Owner's document.

---

## 2026-08-03, morning — picking this up

We have a service called `content-creator`. It writes free-standing copy — blog
posts, social posts — from a request that arrives over Kafka, and hands the text
back. It is not part of the website build pipeline, which matters more than it
sounds.

Everything we have built to stop the platform publishing things that aren't true
was built on the website side. The evidence register, the banned-claims list, the
gate that runs before a page is written, the newer "claims floor" that refuses to
save a section containing a known falsehood — all of it assumes there is a *site*,
because that is where the register of approved facts lives. `content-creator` has
no site. It has a topic and a prompt. So none of that machinery can be pointed at
it, and nobody had noticed, because coverage is reported per site and this thing
belongs to no site.

On 27 July someone read one of its outputs and found this sentence sitting in it:

> *"Industry data shows that large language models experience hallucination rates
> between 3% and 10%."*

No source. Invented range. Attributed to "industry data", which is the phrasing
that makes it read as though it came from somewhere. That was filed as bug 123 and
has sat unowned since.

**What I checked before starting, because the file is a week old.** The bug is
still real: the service still has no validation of any kind — not a single mention
of "validate", "claims" or "evidence" anywhere in its code. It still writes nothing
to our LLM usage log, so it appears in no per-agent report either. Two
invisibilities on one producer.

**One thing in the file is now overstated, and I would rather say so up front.**
It asks for an owner decision "before blog or social output is published anywhere",
which reads as though something is publishing right now. I checked: every workflow
definition that ever referenced `content-creator` has been deleted. Nothing in the
platform currently calls it. The only way to make it run is to publish a message at
it by hand, which is exactly what happened on 27 July. So this is a loaded gun in a
drawer rather than a fire — real, worth fixing, not urgent. The service itself is
up and running, so a guard put inside it will be on the path of anything that does
dispatch it.

**The good news is that the fix got much cheaper while the bug sat.** The day after
123 was filed, another thread shipped a fleet-wide banned-claims scanner that works
*without* a site — it was written that way deliberately, so that a site with no
register of its own is still protected. That is exactly the site-less entry point
this bug asks for, already built and already in use at two other places. So most of
this job is wiring, not building.

**The gap that is left is the interesting one.** The fleet-wide list catches a site
boasting about itself — "guaranteed accurate", "independently verified", "you can
rely on us". It does not catch an invented statistic, because an invented statistic
isn't a boast, it's a fact-shaped sentence with nothing behind it. Catching that
needs something new: a check for a figure that has been *attributed* — "industry
data shows", "studies find", "N% of" — with no citation anywhere near it. I am
planning that as an opt-in, non-blocking detector, for two reasons. It will have
false positives (a genuinely cited figure looks similar), and our own rule, set by
the owner last weekend, is that new authority on a shared piece of machinery ships
switched off by default so that whoever turns it on is the one who decided to.

Next: the plan, then the council, then the code.

---

## 2026-08-03, evening — done, live, and one thing I did not do

It is fixed, it went through the council, and it is running in production. The
short version: content-creator now checks its own copy before handing it back,
using the same machinery the website side already uses, and writes down anything
it finds. It does not refuse to hand the copy over — more on that below.

**The interesting part was not the code.** It was that the corpus kept telling me
I was wrong, twice, and both times I would have shipped something worse without it.

The first version of the detector looked for a figure with no citation *nearby*.
Run against every page we have — 1,130 pieces of copy — it found four things and
all four were wrong. Every one named its source in the same breath: "Deloitte's
2026 banking research found that…", "Thomson Reuters' survey found…". A detector
that flags cited work as uncited is worse than no detector, because it teaches
whoever reads it to ignore the next one.

That sent me back to some sentences I'd already written down, earlier the same
day, as nine real unsourced statistics live on our sites. Read properly — in the
whole paragraph rather than in the hundred characters the tool prints around a
match — six of the nine were nothing of the kind. One was preceded by a full
University of Melbourne citation. One was a legal definition of what
"representative APR" means, which is simply correct. One was referring back to a
figure the page had already sourced. **I had been reading the tool's output
window, not the copy.** That is now written up as a wrong call and as a trap for
whoever next dry-runs one of these checks, because the window's blind spot is
exactly the width of the thing that would have exonerated the match.

What both mistakes have in common is the fix: a citation usually sits in a
*different paragraph* from the figure it supports, so the check has to look at
the whole document, not one paragraph. And a whole document is exactly what
content-creator has — one piece of prose per request, with nothing earlier that
could be carrying a source. So the thing that makes this detector work for the
producer it was built for is the same thing that makes it quiet on our website
copy. Re-run document-wide: zero false positives across all 1,130 pieces, while
still catching the sentence that started the whole bug.

**Two more mistakes the tests caught and re-reading did not.** The check was
reporting *denials* as claims — "industry data does **not** show that 40%…" was
being flagged, because of a subtle interaction with the shared "is this sentence
a denial?" helper. And one of the messages still described a check the code had
stopped doing. Both would have looked fine to a reader.

**The council approved it first time**, thirteen reviewers, no vetoes, three with
minor points. Four of those were checkable rather than arguable, so I checked them
rather than replying — including one worth knowing: there is a place in the
codebase that sorts findings by type, and a new type lands in its "everything
else" bucket and gets *mislabelled* rather than dropped. Nothing routes our new
type there today, but the next person who wires it up would have hit it. That is
written down now.

**The one thing I did not do, and I would rather say it plainly.** I did not send
a test request into production to watch the guard actually fire. The code is
deployed and I have proved the right binary is running — including proving the
*old* one is gone, which is the check people usually skip. But "deployed" is not
"demonstrated", and the bug report itself makes exactly that point. I have left
the recipe in the closed ticket so it is one command whenever anyone wants it.

**And one open decision, which is not mine.** The new detector — the one built
for this exact incident — ships switched **off** by default. That follows the
standing rule that new checks arrive opt-in, and it is sensible given nothing
currently calls this agent at all. But it does mean that if content-creator is
ever put back to work, the check written for its one known failure would not be
running unless someone turns it on. One of the reviewers flagged that, and they
were right to. It is recorded at the top of the closed ticket rather than buried,
because it is the first thing to decide if this agent comes back.
