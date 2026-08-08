# The provocation gate: calibrated on content, one hole on safety (2026-08-08)

## What we're trying to do

vonc.com promises a fresh provocation every day, five times over, on its own home page.
It has been serving the same one since 26 July. The job is to make that promise true
without a person writing a piece every morning: a pool of provocations, a gate that
decides which are good enough to publish, a generator to top the pool up, and a
scheduler to put one out each day.

The gate is the part that matters, because it is the only thing standing between an
automated writer and a live website. It has to do two different jobs at once. It has to
recognise a good provocation — an argument someone could genuinely disagree with — and it
has to refuse the bad ones: invented statistics, party politics, and abuse.

## Where we've come from

The pipeline was built, reviewed by the council and shipped into the chassis by the start
of August. Then we tried to prove the gate actually worked, and that turned out to be the
hard part.

The first attempt scored nine out of nine and was worthless. The test had been built by
writing the provocation bodies by hand, and then checking whether the gate could spot
features the same session had just written in. When the test was rebuilt from the owner's
real prose, the score fell to four out of nine. The gate was rejecting five of the
owner's own published provocations.

Two rounds of diagnosis followed. The first found that the gate demanded every
provocation put a counter-argument, a rule invented from a handful of examples rather
than the whole set — and the owner's answer was that he prefers the one-sided ones, so
the rule went. The second found the deeper problem: the check meant to catch *invented*
facts was actually penalising *uncited* ones, which is simply the register argumentative
writing is written in. A provocation that has to footnote every generalisation is not
writable. That was narrowed to fabrication — the test is whether something was invented,
not whether it was cited — and the change was committed but not yet built into an image.

## What we've done

The change is now live, and we checked it properly: not by trusting the version number,
but by looking inside the running binary on both copies of the service for a phrase the
change added and for the old phrase it deleted. Added phrase present, deleted phrase
gone, on both. Another session's routine rebuild had carried it out to the fleet.

Then we ran the calibration nine times rather than once, because we had learned the
judge gives slightly different answers to identical text.

**The content half is settled.** All nine runs scored the same: eight of the owner's nine
provocations approved, every time. The ninth fails because its body text is genuinely
empty in the database — a true statement about a gap in the pool, not a fault in the
gate, and the framework's job to fill rather than a session's. The two that were failing
for "overstated generalisation" now pass. The deliberately fabricated one, citing a study
from an institute that does not exist, is caught in every run. That is the half we most
needed to keep, and narrowing the rule did not cost it.

**The safety half has a hole.** Alongside the nine real provocations we test four written
to be refused. One is pure abuse: repeated name-calling, no argument. On the third of the
nine runs, the gate approved it — and not because it failed to notice. The judge's own
written note on that run describes the piece as "pure repeated insult with no actual
argument", and then answers that it is safe. The gate reads the answer and ignores the
note.

The reason is structural rather than accidental. The entire safety decision rests on a
single yes/no field from the model, and nothing compares that field to the model's own
reasoning. There is a guard against the reply arriving truncated or malformed, and it
works. There is no guard against a reply that arrives complete and confidently wrong.

## Where we are now

The gate can be trusted to recognise the owner's work: nine runs out of nine agree, and
that result is now cheap to reproduce — three scripts do the reset, the run and the
scoring, so nobody has to retype anything or remember that the gate never re-judges a
row it has already judged.

It cannot yet be trusted to refuse abuse. One failure in nine runs is enough to matter
and rare enough to hide. We are being deliberately careful about that number: a single
occurrence tells you the fault is real, not how often it happens.

That finding also cost us a rule we had been relying on. The inherited standard was "run
it three times and require all three to pass". Runs four through nine were six clean runs
in a row — any three of them would have declared this gate ready. We only saw the failure
because it happened to land on run three. On the rough arithmetic, "three clean runs"
would certify a gate with this fault about seven times in ten. For something that must
never happen, counting clean runs buys far less than it feels like, and we have written
that down rather than leave the next person to inherit the comfortable version.

Nothing is wired to publish, so none of this has reached a reader. The site still shows
26 July.

Two smaller things are worth recording because they are the same mistake twice. The
database stores these provocations across five different prose columns, and reading the
wrong one makes the owner's work look like it is missing. That is what caused the
hand-written test back on 5 August — that session concluded eight of nine provocations
had no text stored, when in fact eight of nine *do*, one column over, marked as written
by a human. Today the same trap briefly suggested the whole pool had been wiped. And the
warning note written to protect the next person from it named the two columns in the
wrong order — caught within the hour by an automated checker that read the actual code,
after our own verification had passed on a set of rows where the two readings cannot
disagree.

## Where we're going

Three things, in order.

The safety hole needs a decision, and it is the owner's rather than ours, because the
options trade cost against certainty. The narrowest fix is a mechanical check for abuse
that runs before the model is consulted at all — the existing party-politics check works
exactly that way and has never once varied across nine runs. Asking the model several
times and refusing on any objection is more thorough and costs more per candidate, and
still never reaches zero. Doing nothing is defensible too, but only if a human approves
each provocation before it publishes, and then it has to be recorded as an accepted risk
with the rate attached, not described as a gate that handles safety. Our recommendation
is the mechanical check, because it removes the randomness instead of sampling it.

The empty provocation body needs generating by the framework, which is the standing
instruction: the system writes the content, not us.

Then the pipeline can be wired to publish — one agent definition and one scheduled task.
The architecture reviewer has already asked that the wiring go through its own council
round rather than ride on the gate's existing approval, and the safety finding belongs in
that submission as an input to the decision.
