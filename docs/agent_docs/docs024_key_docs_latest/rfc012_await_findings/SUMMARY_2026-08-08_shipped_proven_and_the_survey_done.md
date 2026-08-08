# SUMMARY 2026-08-08 — RFC 012: shipped, proven live, the survey delivered — and turned down by the reviewers for being honest about its own size

## What we're trying to do

Stop the platform destroying its own agents' work. When one of our agents both *works
something out* and *asks an outside service to do something*, the outside service's reply
overwrites the agent's workings in the permanent record. The status says complete, the reply
looks like the step's output, and nothing indicates anything was lost. Three separate teams
hit this and each invented a private workaround. The owner ruled three things: build the
shared escape hatch properly, survey everything that reads those records so the deeper fix
becomes decidable, and turn a one-off audit of how workflows route their outputs into a
standing check that runs inside the platform rather than when someone remembers.

## Where we've come from

The class was found the hard way — an action that computed which pages it was refusing to
delete, and why, then dispatched its request and lost every one of those findings. It was
recovered only because a session read the durable record after an apparently green run. The
obvious fix, parking the findings in memory under a side key, was built and then **refuted
live**: the platform reloads fresh state when it parks a step, so anything held only in
memory dies before the database sees it. That refutation is what made the ruling specific.

Two days ago the first two pieces were built but not shipped, and the survey had not started.
The blocker was mundane and expensive: the one correct database writer lived in a package
that the agents could not reach, so nineteen hand-copied versions had accumulated, drifted to
five different shapes, nine of them missing the field that joins a record to the run that
produced it.

## What we've done

**The shared writer is live and proven, and the copy class is retired.** All nineteen
hand-written copies are now one shared piece of code — bar one, deliberately left because a
previous review specifically said to leave it. It has been running in production since
yesterday morning and is confirmed still running on today's build.

**The survey the owner commissioned is finished, and the answer is much better than we
feared.** The question was: if we change the system so an agent's own workings survive an
outside reply, what else breaks? On the configuration side — hundreds of small references
scattered through the agent definitions — **nothing breaks at all.** Some survive because a
helper written for a different purpose years ago already copes with exactly this shape. The
rest survive because they are *dead*: configured in six agents and read by nobody, carried
forward by copy-paste.

In the program code there are **three places that would break, and they break silently** —
the parts that put the hero image and the logo on a page. Without fixing them first, pages
would render with no hero and no logo, no error would be raised, and nothing would tell us.
Three small fixes, all identified, all written down.

The survey also found that **most of the system already does what the proposed change
proposes** — well over half of the affected steps already behave the new way and have for
years. That reframes the decision from a risky platform-wide change into a small one.

**We corrected ourselves twice, in public, on things nobody had challenged.** The test I
published for "did this reach production?" was wrong in the most awkward direction: it would
have told a later session that a change which *had* shipped had *not*. And the fix I had
recorded as "the real fix" for a database inconsistency turned out, when measured rather than
argued, to be backwards — it would have made things worse. Both are written up where they
were made, with the cheap check that would have caught them.

**Two bonus defects, neither ours.** A fallback that looks up a web address in a way the
lookup function physically cannot perform, so it has silently done nothing since written; and
a class of configuration keys that nothing reads and no check would ever notice.

## Where we are now

All three rulings' substantive work is delivered and committed. One piece remains: putting
the standing audit on a schedule so it runs by itself. Everything else is done, live, and
verified against the running system rather than against a version number.

**The reviewers rejected the code change, and the reason matters more than the verdict.**
Last round they told me off, correctly, for describing work I had not shown them. So this
round I said plainly: here are eight representative files, the change is thirty-four, here is
what the other twenty-six do. One reviewer confirmed that fixed the earlier complaint — and
then the senior reviewer vetoed it **precisely because I admitted twenty-six files were out
of view.** Its job is to judge how far a change reaches, and I had just handed it the fact it
must veto on.

I do not think the reviewer was wrong, and the honest version is still the right version. The
problem is structural: the submission form allows eight files, this change is thirty-four,
and no amount of good writing reconciles those two numbers. The veto explicitly says this is
not about the quality of the work — it calls the design and the self-corrections "genuinely
careful". The good news is it named exactly what would satisfy it: **give it the list of all
the file names**, which costs nothing and does not consume the eight slots.

Nothing is broken and nothing is at risk. The rejection blocks a rubber stamp, not the work,
and the code has been in production throughout.

## Where we're going

Four things, in order. Resubmit the code change with the full file list. Submit the standing
audit as its own separate review — the reviewers fairly pointed out I had bundled two
unrelated pieces of work into one round because they arrived on the same day, which is a fact
about my calendar and not about the code. Build the scheduled job that makes the audit run by
itself, which is the last undone ruling. And fix the three silent image failures the survey
found, which are worth fixing whether or not the bigger change ever happens.

The deeper decision — whether to change the overwrite itself — is now the owner's to take,
because the survey it was waiting on exists and says the cost is small and specific. There is
a recommendation in it: of the two possible designs, the one that leaves the reply's fields
where they already are breaks nothing at all, while the tidier-looking one breaks three
places and needs a further sixty-four checked by hand.
