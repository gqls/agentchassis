# The concept register — where we came from, where we are, where we're going

*2026-07-16. The calm, read-aloud version. For the technical detail see
`006_VERIFICATION_stage2.md` (verification method and findings) and
`PLAN_concept_register.md` (the living plan). Turn-by-turn log:
`RUNNING_NOTES_concept_register.md`.*

---

## Where we came from

Agentchassis has been building for months, and the paper trail shows it —
something like four thousand documents scattered across the `docs` folder:
handoffs, runbooks, running notes, plans, from a dozen different workstreams
and eras. Somewhere in that pile is a complete answer to "what does this
system actually do, and does it actually do it" — but nobody could read four
thousand files to find out.

So we built a register. One agent swept every single file and pulled out
every distinct concept it could find — every mechanism, every responsibility,
every behaviour anyone had ever described — and tagged each one with what the
documents *claimed* about it: is this built? Half-built? Just an idea that
never happened? Replaced by something else? Quietly abandoned? We ended up
with about sixteen hundred of these, sorted into just over a hundred
categories that emerged naturally from the material rather than being forced
onto it.

That gave us a map. But a map drawn purely from what the documents *say* is
only as good as the documents — and documents drift out of date the moment
code changes and nobody goes back to update the paragraph about it.

## Where we are: verified, coordinated, and growing a new capability

So the next step was to check the map against the territory. We took every
single one of those sixteen hundred concepts and checked it against the real,
running code — not the docs' word for it, grep the actual file, read the
actual function, confirm the actual database table. About one in thirteen
turned out to be wrong in some way: some things the docs called "finished"
were actually dead code nobody had wired up; some things called "abandoned"
had quietly come back to life; a chunk of them weren't code claims at all,
just working practices and design principles that had been mislabeled for
lack of a better category. We fixed all of it, and — this is the part worth
saying twice — we didn't just accept the first correction that came back for
each one. Every proposed fix went through a second, independent check whose
whole job was to try to prove the first check wrong. About half the proposed
corrections got overturned that way. The ones that survived, we trust.

Then this week we turned outward and checked in with the diagnosis-and-fix
loop — the project this whole register exists to serve. That project just hit
a real milestone of its own: it's now fully built, all four of its phases
live, watching the fleet, sorting real failures from noise, catching the
ones nothing else would ever notice, and reporting all of it in one place
every day. In catching up with it, we found a piece of new work that had
shipped after our original sweep finished, so we brought the register current
— and, in doing that homework, we caught something useful: the very first
real bug the fix-loop team had lined up to work on had, it turned out,
already been quietly fixed by a different team working in parallel the same
day. We flagged that rather than letting anyone spend a real run solving a
problem that no longer existed.

## Where we're going

The whole reason this register exists is to grow that fix-loop's review
panel. Right now, every proposed fix only gets judged by two reviewers. The
plan has always been to add more — someone who remembers whether we've built
this before, someone who's seen this exact bug shape recur, someone who knows
the compliance rules — but until now nobody knew which one to build first.

We do now, and not by guessing: we counted, across the entire history of the
project's documentation, which ideas keep getting independently rediscovered
by people who didn't know someone else had already found them. That's the
strongest possible signal for "this matters and people keep forgetting it."
Two answers came out clearly. One is about reuse — a huge number of the
project's own hard-won lessons say some version of "we built this twice
because nobody checked first." The other is a specific, recurring failure
shape — content silently vanishing during a rebuild — that we found has
independently bitten this platform at least five separate times across
completely different parts of the system, most recently in the very bug the
fix-loop was just about to work on.

We're building that second one first, because it's the one with a real, fresh
example sitting right in front of us: a reviewer whose entire job is to
recognise "we've seen this shape of bug before" the moment it starts to
happen again — before it becomes the sixth occurrence instead of getting
caught at the fifth.
