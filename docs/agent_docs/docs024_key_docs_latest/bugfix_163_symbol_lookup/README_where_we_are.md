# Where we are — the code lookup that could never find a symbol

Plain prose, append-only, newest at the bottom.

---

**2026-08-03, evening.**

We have a code index — a searchable table of every function and type in the codebase — and
two things read it: the diagnosis loop, which is how we establish causes we intend to write
down, and the council's review tier, which is how a proposed change gets checked against
reality. A third thing reads it too: the verifier that checks whether our LANDMINES entries
still describe the system.

All three ask it questions in the same format. One of those question types has never worked.

If you ask it "does `ReadSymbolBody` exist in `internal/analysis/symbolbody.go`?", it chops
your question into words — `internal`, `analysis`, `symbolbody`, `go`, `ReadSymbolBody` — and
then demands that **every one of those words appear inside the function's name**. No function
is called "internal". So the answer comes back empty, always. Not sometimes, not for hard
cases: by construction, for every question of that shape.

That would be bad enough. What makes it worse is what the answer *says*. It does not say "I
could not answer that". It says: *"searched the names of 4,992 indexed symbols. The query was
RUN and matched none; this is not an unanswered question."* A confident denial, in the same
voice it uses when it has genuinely looked.

The verifier that reads those denials has produced twenty verdicts since 31 July. Sixteen say
"needs human review". **None has ever confirmed anything.** And three separate workstreams
have now written in their notes that they deliberately did *not* run the verifier on their
work, because they knew it would come back unable to confirm — so the damage is not just wrong
answers, it is people routing around a check we built and paid for.

The fix is small and it goes in one place. Split the question at the colon: the part before it
is a file path, so look for it in the *path* column; the part after it is a name, so look for
it in the *name* column. That is all. The same edit repairs all three readers, because they
share the one function.

Two things made it more than a one-line change. First, when we looked at how people actually
write these references, a third form turned up that nobody had accounted for: about a dozen
entries write a **line number** after the colon rather than a symbol name — `spawn_actions.go:3066`.
A naive fix would send "3066" off to be matched against function names and reproduce the
original bug wearing a different hat. So the fix has to notice a line number and treat the
whole thing as a file question.

Second, we decided that when a symbol genuinely is not at the path you named, the tool should
say so *and then tell you where it is*. The failure that started all this was a symbol being
reported absent when it existed perfectly well — the file had simply been described
differently. A tool that answers "not there, but here it is over here" cannot make that
mistake, and it costs one extra query only when the first one misses.

One piece of housekeeping worth recording, because it nearly cost us the bug. Another
workstream had scanned live sessions in early August, seen one talking a great deal about "the
landmine verifier", and concluded this bug was already being worked on — so they dropped it.
It was not: that session was working on the landmine *corpus*, and had never once opened the
function where the fault lives. The bug then sat untouched for two more days. The lesson is
narrow and useful: when you check whether someone else is on a bug, search for the **function
you would have to edit**, not the name of the thing that is broken. Everyone adjacent says the
latter; only the person fixing it says the former.
