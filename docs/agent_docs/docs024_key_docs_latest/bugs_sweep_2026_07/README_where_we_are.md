# Where we are — bugs sweep (plain prose, append-only)

## 2026-07-28

We pointed a session at the open-bugs folder and asked it to work through them, preferring
fixes that help the whole platform over ones that patch a single site.

Four bugs are properly finished and closed: a page assembly that reported success while
producing nothing, tool pages that were publishing their internal build instructions to
Google, a field that was documented and read by no code at all, and a publish script whose
main branch could never run. Three more are fixed and shipped but stay open because
something real is left over — in one case two duplicate news pages on a live site that no
code change can tidy up.

Three things we learned are worth more than the fixes.

**We had already decided the news URL question, and I re-opened it by mistake.** The
convention is written into the code and there is a site already doing it correctly. I even
quoted the file that says so, as evidence for something else, and still asked for a ruling.
Escalating feels like the careful thing to do, and it isn't free — it writes "undecided"
into the record of a decided question.

**One planned fix turned out to be impossible, and finding that out was the work.** We
wanted the system to spot a page that is secretly the news page under the wrong label. It
can't: on robot-hands there is a catalogue page and a news page that look *identical* to
any check we can write. Shipping the fix would have re-labelled the catalogue and broken it.

**The diagnosis tool is being lied to by its own code index.** It searched for a function,
was told "found nothing, and this is a real answer, not a gap", and gave up. The function
exists. The index is pinned to a snapshot from four days ago, about a thousand commits
behind. Until that is fixed, asking the diagnosis loop anything about recent code wastes
the run — which is exactly what happened to us.

I also broke something myself: deploying a new build killed a review that was running at
that moment, and then I misread a dead job as a slow one for over an hour. Both are written
up so the next person checks what is in flight before deploying.
