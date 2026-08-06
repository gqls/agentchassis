# SUMMARY 2026-08-06 — RFC 012: two of three built, and the specification was one key short

## What we're trying to do

Stop the platform destroying its own agents' work. When one of our agents both *works
something out* and *asks an outside service to do something*, the outside service's reply
overwrites the agent's workings in the permanent record. The status says complete, the
reply looks like the step's output, and nothing indicates anything was lost. Three separate
teams have hit this and each invented a private workaround. The owner ruled three things:
build the shared escape hatch properly, census everything that reads those records so the
deeper fix becomes decidable, and turn a one-off audit of how workflows route their outputs
into a standing check that runs inside the platform.

## Where we've come from

The class was found the hard way. A retraction action computed which pages it was refusing
to delete and why, dispatched its request, and lost every one of those findings — recovered
only because a session read the durable record after an apparently green run. The obvious
fix (park the findings under a side key in memory) was built, tested, and **refuted live**:
the platform reloads fresh state when it parks a step, so anything held only in memory dies
before the database ever sees it. That refutation is what made the ruling specific — the
escape hatch had to be a direct database row, written before the dispatch.

Separately, the same overwrite has a twin that fires without any outside service involved,
and it took every page build in the fleet down for forty minutes. A census found a second,
silent instance of the same shape that nobody had reported. That census is what the owner
asked to make standing.

## What we've done

**The shared writer exists and is live.** The obstacle was a circular dependency: the one
correct database writer lived in a package that couldn't be reached from where the agents
run, so nineteen hand-copied versions had accumulated, with column lists drifted to five
different shapes — nine of them omitting the field that joins a record to the run that
produced it. Moving the writer *below* both sides dissolved the cycle. Old callers compile
untouched; the exemplar conversion's tests pass unchanged, which is the proof nothing
changed shape. It is verified running in both production pods.

**The standing audit is built and, crucially, proven able to fail.** It catches the
outage-causing bug that both simpler versions of the same check miss entirely — and a
deliberate mutation, severing the single routing edge that connects the two steps, makes
the finding vanish. Without that second test, a check that always says "clean" is
indistinguishable from one that works. Against the live fleet it found exactly the two known
problems out of 176 agents and nothing spurious. It carries a signed-off list of those two
so it stays green until something *new* appears.

**And it found a real defect in its own specification.** Rebuilding the audit meant
re-deriving how a workflow can route between steps. The specification listed those routes;
measured against the live fleet, **the list was one short** — a routing key that appears 158
times. Transcribing the prose would have left a blind spot in the very check written to have
none.

## Where we are now

Two of the three rulings are delivered and committed; the third has not started. What
remains: converting the other eighteen copied database writes onto the new shared one;
wrapping the audit in the scheduled job that makes it "online"; and the census itself. Also
owed: one review-council round for the code, and two entries in the concept register.

The remaining work is fully specified rather than merely intended — each of the eighteen
sites is listed with the quirk its conversion must preserve, the scheduled-job pattern to
copy is identified (ship the real binary as an image, rather than the Python
re-implementation an earlier check was forced into), and the two traps a naive census would
fall into are written down.

## Where we're going

A fresh session picks this up from the handoff. The order matters: finish the conversions
(they retire the copy class the ruling was about), then the scheduled job (it makes the
audit real rather than a script someone must remember), then the census — which is a full
session's work on its own and gates a further owner decision about changing the overwrite
itself. Nothing in flight is blocked on anyone; nothing on a live site changes.
