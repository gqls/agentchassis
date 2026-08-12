# SUMMARY — where we are, 2026-08-12

*Concept register. Written to be read aloud. Previous milestone:
`SUMMARY_where_we_are_2026-08-10.md`; the series is the record, so this is a new file
rather than an edit of that one. It covers two days rather than one, because the 11th
produced a finding big enough to deserve a read-out and never got one. Both of the
open questions the 10th ended on are now answered — and one of them turned out to be
the wrong question, which is the more useful half.*

## What we're trying to do

Keep one place that answers **"does this already exist, and is it real?"** for every
mechanism, contract, agent, tool and idea in the platform. Documentation says how a
thing works; the register says *that* it exists, who built it, what it is for, what
will mislead you about it, and whether it is genuinely live or merely written down.

Two audiences, and the second is why accuracy is not housekeeping: a session about to
build something, who needs to discover it was built in June; and the council seats
that review our changes, whose prompts are seeded from these entries. A wrong entry
is not misleading prose — it is false evidence inside a machine review.

## Where we've come from

Built in July by sweeping the documentation tree, then checking every concept against
live code and the database. On 4 August we found thirty-four concepts with an entry
and no row in the index — the index being the thing people search, so those read as
*does not exist*. A daily watcher followed. On 9 August every stored count in the
register was retired at the owner's direction, because four different commands
counted the index and all four answers were individually correct.

On 10 August we built the gate: when a commit adds a concept without its index row,
the commit says so, to the person writing it, before it lands. That read-out ended on
two open questions. **First**, staleness — coverage tells us a thing is present, drift
tells us the two halves agree, the gate tells us they shipped together, and nothing
at all told us whether an entry is *still true*. **Second**, a test for the gate
itself: does the missing-row count actually fall to zero, or is the gate simply being
ignored? The advice was to watch the daily row for a week.

Within three hours of that gate shipping, a concept landed with no index row anyway.

## What we've done

**We found out why the gate was ignored, and it wasn't.** It was never heard. The
commit that slipped past it ended by trimming its own output to the last eight lines —
an ordinary way to cut a noisy command down to the part you want — and the warnings
print at the *top*, while the receipt you asked for prints at the *bottom*. So the
trim keeps the receipt and discards the warning, every time.

Then it got larger. Across every commit made through the tool since we added those
warnings in July: **of 2,669 commits touching more than one file, 1,199 — forty-five
per cent — showed the session nothing.** Ninety-five per cent of those were cut by the
session's own trim, across 258 different sessions. It was never one lane's bad habit,
and it was not only our register check going missing: the same trim removed the report
that stands between one session and committing another's half-finished work, plus all
seventeen automated code checks. All computed correctly, all binned. The one hook that
survived prints later, which is precisely why the machinery looked healthy.

The fix was small: after any commit, a hook re-runs those reports against the commit
just made and hands the result to the session directly, outside the command's output,
where no trim can reach it.

**And then the instrument we built to check that fix turned out to be blind to it.**
Run as documented the next morning it reported thirty-eight per cent delivered —
*worse* than the fifty-five we started from. It was counting warnings that appeared in
the command's own output, and the whole point of the fix was to stop sending them
there. Everything the new route delivered was scored as a failure. Taught to look at
both routes, the real figure for the 12th is **every one of thirty-six commits
reached, twenty-three of them only because of the fix.**

That near-miss is worth dwelling on. The wrong number did not fail quietly — it
pointed confidently the other way, on a day the fix was flawless, and it argued for
exactly the conclusion we had just spent a day disproving. The checker had four
safeguards on it and all four passed. All four were watching the old route.

**We also stopped ourselves declaring victory on the second question.** Here is the
shape of it, in order.

*What the signal is:* the count of concepts arriving without their index row. Since
the fix went in, that count is zero, across all seventeen commits that touched the
register.

*The rule for believing a zero:* a clean result only means something if a dirty result
was a realistic possibility. You have to ask what the measurement could have shown if
nothing had improved.

*How this case measures against it:* only **four** commits in that window even added
the kind of entry that could go wrong. At our own historical error rate of about one
in six, four clean commits happen **half the time by luck alone.** So the zero is
the expected outcome whether the fix changed anybody's behaviour or not. It would take
roughly fourteen such commits before a clean run meant anything — days away. Delivery
is proven; whether being *told* changes what people do is untouched.

**On the other flank, staleness, we closed the first of three signals.** Many entries
cite a platform version — "true as of version 1283" — and eighty of them were fifty or
more releases behind. Showing the oldest sounded like a morning's work.

Measuring it first is what saved building the wrong thing. A version in an entry means
one of two opposite things and they look identical: "shipped in version 1029" is a
fact about history and will be true for ever, while "both servers on version 1218 gave
the right answer" is a check, and checks expire. Sorting them by the words around them
failed on three quarters of the cases. A list of a hundred and eleven "stale" entries
where most are permanent facts is a list nobody reads twice.

What worked was already there. The register writes each entry in labelled parts, and
two of those labels are — by its own convention — claims about how things are *right
now*. Keying on the label rather than the sentence needs no cleverness, and it also
turned up something we hadn't looked for: the *evidence* lines are far staler than the
*status* lines. We update what we claim; we don't go back and re-check why.

Then one genuinely sharp result. Some entries quote a container version as proof that
an agent is running. Every one of the 187 live agent records carries the current
version, uniformly — the release rewrites them. So quoting one of those numbers only
ever records the day you looked. That turned a long vague list into a short precise
one, and two entries were fixed off it the same afternoon.

The best illustration is a pair. Two entries both cite version 407, from last
November, 883 releases ago. One is **wrong**: it says a live record still references
that version, and that record now says 1290. The other is **entirely fine**: it
describes what a setup file in the repository says, and that file does still say 407.
Same number, same day, opposite verdicts, and only knowing which artefact is being
described tells you which. So the new tool names the category and refuses to pass
judgement, printing the one-line check instead.

## Where we are now

Entries and index rows agree exactly, and no identifier is used twice. The drift
check's only outstanding finding is a single stored count in a file that has been
another session's half-finished work since the 8th — left alone deliberately for the
fourth time, because tidying our one line out of it would sweep their changes into our
commit.

The warnings that guard this whole estate now reach the people they are written for:
today, all of them, for the first time we can demonstrate. Version lag is visible and
tooled, with its premise honestly narrowed. Two staleness signals remain open —
ninety-six citations pointing at deleted files, and a hundred and fifty-six pointing
at bugs that have moved.

The cold-start document has been replaced, because the old one would now send a
session to redo finished work and to read a clean signal as a vindication it isn't.

## Where we're going

**The two remaining staleness signals, carrying the question that made the first one
work:** is there something to key on that doesn't involve reading English? Version lag
only became trustworthy when it stopped trying to understand sentences and keyed on
structure the register already maintains. That question is worth asking of both
remaining signals before either is built.

**The behaviour question stays open on purpose, and needs no work — only patience.**
In a couple of weeks there will be enough commits for a clean run to mean something.
Until then nobody should argue for making the gate blocking in either direction, and
the standing argument against it — that a check which blocks on a bad day gets
switched off for ever, and a false alarm on a shared tree is everyone's outage — is
untouched by any of this.

**A small authoring rule with real leverage, still not done.** Thirteen of
twenty-nine entries examined cite no commit reference at all, so there is no cheap way
to date them. An entry whose status depends on a release should name its commit: nine
characters when written, a one-command check for ever. That belongs in the gate, where
the error is made, not in another report.

**And the thread through the last three days, which is now one turn longer.** The 10th
said a measurement is only as honest as the thing that could have made it come out
differently. The 11th added that a warning computed perfectly and delivered to nobody
is not a warning. The 12th joins them: **when you move where a signal is delivered,
the thing that measures delivery is the first thing to go stale — and it goes on
producing confident numbers while it does.** Three days, three instruments, each
answering a smaller question than the one being asked, in exactly the same voice. The
only defence we have found is to ask, every time, what result would have proved us
wrong — and to check that our instrument was ever capable of producing it.
