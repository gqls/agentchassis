# Where we are — the fix-loop family

*The owner's running plain-prose log for the three workstreams that share this
directory: the **fix loop** itself, the **feature builder**, and the **council
gate**. Append-only, newest at the bottom. Never rewrite or reorder what is
already here, and never edit someone else's words — add a dated correction
underneath instead. If it reads like a pasted chat transcript, that's because it
is one, and that's fine.*

*(Started 2026-07-19 by the council-gate thread, because CLAUDE.md's standing-five
directive asks for this file and this directory didn't have one. The fix loop's
older prose log is `README_so_far.md` — left exactly as it is; this is not a
replacement for it.)*

---

## 2026-07-19 — council gate: live, adopted by three threads, and it has started catching things

The council gate is the fix loop's reviewer council opened up as a service, so
any working session can put a change through it before committing. It went live
on the 17th and it is now being used by threads other than the one that built it.

**Where it stands.** Thirteen reviewers. Two of them always sit — one checking
the edits are real and minimal, one guarding the rest of the platform, with the
only power to block. The other eleven only wake up when a change actually
touches their territory, which is what keeps this affordable at the pace people
work. A thread submits a change with its reasoning, gets a verdict in about two
minutes, and if approved commits with a one-line trailer so a report can later
say who was reviewed and who wasn't.

**Adoption happened without anyone being told to do it.** Within a few hours,
three separate threads had used it: the imagery thread submitted, got asked to
revise, revised, and resubmitted on the same correlation — exactly the intended
loop. The feature-builder thread submitted a stage-loop controller. A third
carried the trailer on a commit. That matters because the gate is advisory: it
cannot make anyone use it, so use is the only real measure of whether it earns
its place.

**It has already caught a real thing, on its first run.** Our own test
submission proposed a change that rested on an assumption about how records are
tagged. The reviewers are allowed to query the live database before deciding,
and their queries showed the assumption was simply false — the change would have
produced a permanently empty section, silently, with no error. That is the
platform's most persistent failure shape, and it was caught before a line of
code existed.

**Three things about the tooling turned out to be wrong, and I'd rather record
them than the successes.** First, the coverage report was quietly lying: a
plumbing mistake meant it stopped counting at the first reviewed commit, so it
was reporting four commits when there were forty-one, and looking perfectly
healthy while doing it. Another thread found that. Second, the same report
accused an honest commit of faking its review, because it only recognised one of
the two kinds of identifier a thread might legitimately cite. Third — and this is
the one worth remembering — the verdict that commit pointed at was *deleted*
between two runs of the report, by a documented practice that told people to
clear old council records before a fresh run. So a properly reviewed commit
became indistinguishable from a false claim. That advice is now retired (it was
also obsolete), and the report distinguishes "we can't find your evidence" from
"you didn't have any", because those are very different accusations.

**A related bug, flagged by the owner and audited by the reasoning-dataset
thread.** The fix loop's reviser — the step that rewrites a plan after the
council objects — had been receiving *blank* reviews for some time, through a
subtle templating fault. It was revising against nothing while looking like it
was working. That's fixed and now proven in live traffic. The same audit showed
a second, quieter version of the problem: the reviser only ever saw six of the
thirteen reviewers, because each new seat had to be threaded into the prompt by
hand and nobody had. Rather than list all thirteen — which would break again on
the fourteenth — the reviser now reads the council's report as a single
document, so new seats reach it automatically. That fix is applied but has not
yet had a chance to run, and I'm watching for the first one rather than claiming
it works.

**What I'd like a decision on, when you're ready.** The gate is advisory by
design and can't intercept a commit. The open question is whether it ever should
— platform changes riding branches, with only approved work merging. That
changes how every session works, so it waits for you and for evidence from the
advisory period. The coverage number is the input to that decision, and it is
currently very low: most platform commits still go unreviewed, which is exactly
what you'd expect a day in.
