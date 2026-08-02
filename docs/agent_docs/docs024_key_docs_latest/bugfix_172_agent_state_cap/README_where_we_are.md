# Where we are — the diagnosis loop's agent-state evidence (bugs_open/172)

*Plain-prose log, append-only, newest at the bottom.*

---

**2026-08-02, evening.**

I picked up bug 172 off the open pile. It was one of the few nobody else was in the
middle of — I checked by grepping every live session's transcript for the code
symbols, not just the bug number, because the bug number turns up everywhere (it is
in the index every session loads) and that makes an unowned bug look busy.

The bug is about our own debugging machinery. When the system diagnoses itself, it
builds a dossier for the model that has to decide what went wrong, and one section
of that dossier says "here are the agents your symptom mentioned, and here is what
they've been doing". The complaint was that this section quietly drops agents when
too many are named, while its heading still claims to list them all — so the model
deciding the diagnosis can be told nothing about an agent and read that silence as
"this agent wasn't involved".

The person who filed it had measured it and found it wasn't actually biting yet: it
allows five agents and we'd never seen more than four named. Latent. I re-measured
and that still holds.

**Then the job changed.** Before fixing something I like to look at what it actually
produces, and the number that mattered wasn't the one in the ticket. The dossier
also lists each named agent's recent AI calls, capped at ten lines. Ten lines looked
healthy — plenty of evidence, in most dossiers. But when I counted how many
*different agents* those ten lines belonged to, the answer was always one.

Every time. Twenty-three dossiers named two, three or four agents and had call
history in them, and in all twenty-three, all the lines belonged to a single agent.

The cause is one line of database code. It asks for "the ten most recent calls by any
of these agents" rather than "ten each". So the busiest agent takes all ten and the
others show up with nothing — and nothing on the page distinguishes "this agent has
been quiet" from "we never looked". I reproduced it directly: ask about three agents
together and you get ten rows, all from one of them, while another with **eighteen
thousand** calls to its name renders not a single line.

So the ticket's cap was asleep and its neighbour, three lines further down, had been
awake since at least the 20th of July. I fixed both, because they are the same
mistake twice and fixing only the named one leaves the live one behind a file that
now looks audited.

**What I changed, in plain terms.** Each named agent now gets its own allowance of
call history instead of them all sharing one. When agents *are* dropped, the section
says so and names them, instead of quietly shortening the list. When an agent has no
call history at all, it now says that out loud, so silence stops being ambiguous.
And the list of agents is now sorted, which sounds trivial but wasn't: without it,
two identical diagnoses could quietly examine different agents and both report
success.

**Two things I got wrong, both worth telling you about.**

The first is small and the compiler caught it. The second is not. I wrote a test to
prove the sorting fix worked, watched it pass, and nearly moved on. Then I deleted
the sorting fix on purpose to check the test would notice — and it didn't. The test
was using a fake database that hands back rows in whatever order *the test itself*
chose, so it could never have detected the real thing going missing. It was checking
my own answer against itself. The ticket had actually warned about this in one line
and I'd read it and still walked into it. Fixed properly now, and written up in the
fleet's wrong-calls log, because the lesson generalises: when you use a stand-in for
the real system, ask whether the thing you're testing could even survive the
substitution.

**The council reviewed it and approved it first time**, with four advisory notes.
One of them was a genuine catch and I'm glad it went through: my new "this agent has
no call history" message asserted more than it could support. We renamed how agents
are recorded back in July, so an agent can have plenty of history filed under its old
name and none under the new one. My message would have handed the diagnosing model a
confident "made no calls" — which is precisely the false reassurance the whole fix
exists to remove, reintroduced by the fix's own wording. Now it states what's
actually true and names the rename. That's fixed and committed.

Another reviewer made a fair architectural point: this is the third time we've
patched this same class of bug in this corner of the system one instance at a time,
and nothing checks for the others. So I went looking, and found a fourth — three more
silent caps in the neighbouring file, where the same function already announces one
of its own caps eight lines earlier and simply doesn't announce these. I've filed
that as bug 181 rather than widening this change.

While filing it I made the same mistake twice in ten minutes: I ran a measurement
that returned a clean zero and read it as "this never happens", when in fact the
query hadn't looked at anything at all — the population was empty. Caught it both
times by asking "how many did it examine?", and I've written the guard into the new
bug file so the next person doesn't repeat it.

**Where it stands right now.** The fix is written, tested, reviewed, approved and
committed. It is **not yet running in production** — that needs the chassis to be
rebuilt and rolled, and when I went to do it, another session had a council review
mid-flight. Rolling would have killed it and cost them the round. So I'm holding.
Our own rule is that a bug stays open until the fix is actually live, not merely
committed, so 172 stays open until then and I'd rather say that plainly than tidy it
away. Everything needed to finish it — the build, the two pod checks, and which
strings to grep for — is written down in the runbook.

**One thing you may want to know about the existing dossiers.** Every diagnosis
dossier written before this ships has the flaw in it, and we keep them for about a
month. If anyone reads one and concludes "agent X made no AI calls", that conclusion
is not supported. I've added a warning to the fleet's landmines file so it surfaces
automatically for anyone who touches that data.
