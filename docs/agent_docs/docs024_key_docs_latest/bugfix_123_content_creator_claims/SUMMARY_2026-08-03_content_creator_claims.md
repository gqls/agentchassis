# SUMMARY 2026-08-03 — the claims layer reaches a producer that has no site

Milestone read-out for `bugs_open/123` → `bugs_closed/123`. Written to be read
aloud.

---

## What we're trying to do

Stop the platform publishing things that are not true. We have built a lot of
machinery for that over the last few weeks — a register of approved facts per
site, a list of claims no site may make about itself, a gate that runs before a
page is written, and a floor that refuses to save a section containing a known
falsehood.

All of it was built on the website side, and all of it starts from the same
question: *which site is this?* That is how it finds the register of facts to
check against.

## Where we've come from

`content-creator` is a service that writes free-standing copy — blog posts,
social posts — from a request that arrives over Kafka. It belongs to no site. It
has a topic and a prompt, and it hands text back.

So none of the machinery above could be pointed at it, and nobody noticed,
because coverage is reported per site and this thing belongs to no site. Our own
code says so: the claims floor's source file lists, as its scope boundary,
*"bugs_open/123's content-creator path has no site and no page row, so this seam
cannot reach it at all."*

On 27 July someone read one of its outputs and found this sitting in it:

> *"Industry data shows that large language models experience hallucination rates
> between 3% and 10% depending on the task."*

No source. Invented range. Attributed to "industry data", which is the phrasing
that makes it read as though it came from somewhere. That was filed as bug 123
and sat unowned for a week.

## What we've done

**Checked the bug was still real before working it**, because the file was a week
old. It was: the service still had no validation of any kind, and still wrote
nothing to our LLM usage log, so it appeared in no per-agent report either — two
invisibilities on one producer.

**Corrected the bug's own severity downward.** It was filed HIGH and asked for an
owner decision "before blog or social output is published anywhere". Every
workflow definition that ever referenced this agent has been deleted. Nothing
calls it. It is a loaded gun in a drawer, not a fire — which also means the fix
could not regress anything.

**Found that half the job had already been done by someone else.** The day after
123 was filed, another thread shipped a fleet-wide claim scanner that works
*without* a site, deliberately, so that a site with no register of its own is
still protected. That is exactly the entry point this bug asks for. So the first
half was wiring, not building.

**Built the second half, and let the corpus design it.** The existing list
catches a site boasting about itself; it does not catch an invented statistic,
because an invented statistic is not a boast. The new detector looks for a figure
attributed to unnamed evidence in a document that cites nothing anywhere.

That last clause — *anywhere* — is the whole design, and we got it wrong twice
before measuring. Checking each paragraph on its own, against all 1,130 pieces of
copy we have live, the detector found four things and every one was wrong: each
named its source in the same sentence. Going back over that, six of nine
"unsourced statistics" recorded earlier the same day turned out to be cited, or
legal definitions, or references back to a figure the page had already sourced.
We had been reading the tool's hundred-character output window rather than the
copy. Checked document-wide instead, it produces **zero** false positives across
the entire estate while still catching the sentence that started the bug.

**Shipped it as record-and-annotate, never refuse.** The service writes down what
it finds and attaches it to the reply; it never withholds the text. Refusing would
change what callers observe on failure, which our architecture reviewer has
already ruled needs its own review round, and the reply path belongs to another
live piece of work.

**Council approved it first time** — thirteen reviewers, no vetoes, three minor
points. Four were checkable rather than arguable, so we checked them.

**It is live**, on the build that rolled this evening, and verified against the
running binary in both directions — the new strings present *and* the old one
gone, which is the half people skip.

## Where we are now

Closed. The bug is in `bugs_closed/`, the mechanism is in the concept register as
CLM-019, two new traps are in the landmines file, and the wrong call is logged.

Two things are honestly incomplete and both are written at the top of the closed
ticket rather than buried:

1. **Deployed is not demonstrated.** We did not send a test request into
   production to watch the guard fire, so its behaviour in the real system rests
   on the test suite rather than on a real run. The recipe is one command and it
   is written down.
2. **The detector built for this incident is switched off by default.** That
   follows our standing rule that new checks arrive opt-in, and it is reasonable
   while nothing calls this agent — but a reviewer was right to name it, and it
   is the first decision to take if the agent is ever put back to work.

## Where we're going

Nothing is owed on this bug. Three things are worth someone's attention later,
in descending order of usefulness:

- **The next producer with no site will pay this cost again.** Our architecture
  reviewer made the point well: the claims layer has exactly one way in, and what
  we built is a hand-wired guard living inside one agent. A chat agent, a summary
  agent, an email drafter — each would re-derive the same four pieces. That is a
  design question, not a bug, and it is recorded as one.
- **Wiring the new check type into the website-side gates is an architecture
  review**, not a follow-up tweak. Said plainly in the register so it cannot slip
  through as "just adding a case to a switch".
- **content-creator still writes nothing to the usage log**, which is the second
  invisibility this bug named and we did not fix. It is a different seam.
