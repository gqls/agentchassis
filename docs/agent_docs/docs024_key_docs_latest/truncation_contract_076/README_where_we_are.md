# Where we are — the truncation contract (076)

Plain-prose running log, append-only, newest at the bottom.

---

**2026-07-26, evening.**

Background in one breath: when we ask a model for an answer and it runs out of
room mid-sentence, we normally treat that as a failure. For a few workflows we
deliberately keep the half-answer instead — a council of sixteen reviewers should
not lose the whole round because the third reviewer overran. Keeping it is only
safe if something further down the line *knows* it is half an answer. Case 076
was that the "something further down" was enforced by nothing at all: any step
could declare "truncation is fine by me" and be believed, with nobody checking
that anyone downstream could tell a fragment from a finished answer.

That is fixed and live. But the fix only speaks up at the worst possible moment —
in production, when a response has actually been cut, by failing that run. It
cannot tell you that a piece of configuration is wrong until the wrongness bites.
The council reviewing the fix said so at the time, and that objection is the piece
of work I have picked up: **read the configuration itself and say so beforehand.**

The awkward part, and it is the whole difficulty, is that there is no moment to
attach the check to. Our agent configuration lives in the database and is live
the instant it is written — there is no build, no deploy, no restart to hang a
gate on. So instead of one gate there are two cheap checks pointing at the same
thing from different sides: one that reads the live fleet and tells you what is
in it right now, and one that runs when someone commits a new agent definition and
tells them then and there. Neither blocks anything, on purpose: a check that can
stop the whole team committing on a bad day gets switched off permanently, and
we have that written down from experience.

Two things I checked before starting rather than after. First, the live fleet: 37
steps currently tolerate truncation and all 37 sit in workflows that do read the
marker, so the new report will say "clean" — which means a clean report is not
evidence of anything, and I will have to deliberately break something to prove
the check works. Second, the handoff's suggestion for the commit-time check
turned out to be aimed slightly wrong: it assumed the bad configuration arrives in
a file that describes the whole workflow, and in reality all three existing files
of that kind are small patches that only name the flag. Had I built what was
suggested, it would have flagged three perfectly correct files on day one, which
is exactly how a check earns a reputation for crying wolf. So it is scoped to the
files that genuinely contain the answer.

---

**2026-07-26, later that evening.**

The two checks are built, and both have been shown to work by being made to fail
rather than by being run and coming back clean.

What exists now: a small script that reads the live fleet and tells you whether
any agent is configured to keep a half-answer that nothing downstream will notice
is half; a check that runs automatically whenever anyone commits a new agent
definition and says the same thing about the file in front of them; and a line at
the end of the migration runner that names the file you just applied and reminds
you to run the first one. None of them can stop anybody doing anything — they
tell you, and that is deliberate.

The proof is the part worth reporting. Today's fleet is clean — 37 places keep
half-answers, all 37 sit behind something that reads them properly — so running
the checker and seeing "clean" tells you nothing at all, because a completely
broken checker prints exactly the same thing. So I created three deliberately
misconfigured agents in the live system, one genuinely wrong and two that only
look wrong, checked that the tool flagged the wrong one and cleared the other
two, and deleted all three. The commit-time check got the same treatment against
every SQL file in the repository — 849 of them, no false alarms — plus seven
hand-built files designed to trip it, which it handled correctly, including one
where the fault was buried in a workflow nested inside another workflow.

Three things I got wrong along the way, all now fixed and all written down in the
technical notes. I claimed our Go code accepts the word "true" as well as a real
yes/no value, without opening the function — it does not, and had I not checked,
the tool would have complained about configuration that is already safe. My first
version quietly scanned 166 of 171 agents, because five agents exist twice in the
database and I was storing them by name, so the second copy of each overwrote the
first — and a duplicate record is exactly where stale configuration hides. And a
second session working from the same handoff found that my parser silently
skipped any entry whose description was long enough to wrap onto a second line;
they reproduced it properly, I confirmed it, and fixed it in a way that handles
the wrapped case rather than just refusing to run.

That collision is worth a sentence of its own, because it is the failure mode we
already know about: two sessions picked the same task off the same handoff within
half an hour, and the tool we have for checking who owns a piece of work reads
committed history, so neither of us was visible to the other until one tried to
write to a file the other had open. They stood down, left their finding in my
notes, and it was a good finding. Nothing was lost this time. The cost was about
half an hour of duplicated thinking.
