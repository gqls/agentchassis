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
