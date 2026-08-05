# Where we are — the repair agent that never repaired anything

Plain prose, append-only, newest at the bottom. The owner maintains this file too — append,
never rewrite, and add a dated correction rather than editing anything above.

---

**2026-08-05.** This lane exists because of a wrong turn on a different job, so it is worth
saying how it started.

I was finishing off bug 194 and needed to prove one last thing by running a real job on a real
site. You suggested ai-agent-orchestration.com. Chasing that, I hit three walls one after
another, and the third turned out to be a bug somebody else had already filed on Monday — bug
201. You then asked me to fix that instead, which is what this is.

**What was wrong.** When one of our automated checks finds a small defect on a page — markdown
symbols showing up as literal asterisks, a fabricated phone number, a page with no content —
it files a repair job and names the agent that should do the repair. Three of those checks
named the *writer* agent directly. Every single one of those repairs failed. Twelve were tried
last week; eleven died outright and the twelfth reported success having changed nothing.

**Why, and it is not what the error message said.** The error read "planned its own sections
and none are ready", which sounds like the page's content wasn't available yet. It actually
means something much sillier: **nobody had told it which parts of the page to work on.** The
writer, called directly, expects the caller to hand it a list of the page's sections. The
checks don't send one — I confirmed that every one of the fourteen repair jobs ever created
this way lacks that field entirely. So the writer looks at an empty list, correctly concludes
there is nothing to write, and gives up. The message describes the state it ended in rather
than the input it never got, which is why this reads as a content problem and is really a
wiring problem.

**The fix, and why it is a small one.** There is a second agent — the page *build handler* —
which fetches the section list from the site's own plan rather than trusting whoever called
it. It cannot fail this way. It is also the agent that ends up calling the writer anyway, once
it has done the planning. So the repair is to point those three checks at the build handler
instead. That is three lines.

I did not want to take that on trust, so I checked it two ways. The build handler is doing
this job successfully in production right now — thirty-two completed repairs on pages that
were already built, which is exactly the case that was supposed to be doubtful. And two other
checks were **already migrated** the same way earlier; their file headers still record it.
These three were simply the ones nobody got to.

**One of the three was plainly just wrong**, independently of all that: it files a job of a
type that our own planner also files, and the planner sends that type to the build handler.
The same kind of job was going to two different places depending on who asked.

**Something I considered and rejected**, recorded because it looks like the obviously better
idea. We have a purpose-built agent for editing a single slot on a page, which is a scalpel
where the build handler is a hammer — it rebuilds more of the page than strictly needs
rebuilding. I went and read it. It has no writing step at all; it applies an edit that someone
else has already composed, and what these checks produce is an *instruction* to a writer, not
finished replacement text. Using it would have meant building a new agent to sit in front of
it. That is a much bigger change than this bug deserves, so I have written the argument down
for whoever wants to build it later rather than half-starting it now.

**One consequence I want to flag, because it is the sort of thing that gets "tidied up" into a
fault.** After this change, nothing in the system files jobs for that writer any more. There is
an old orchestration loop that was waiting for exactly those jobs, and it will now never
receive any. It has never actually run, and the only jobs it could ever have received were the
ones that were failing — so I have made an existing dead end visible rather than created one.
The tempting next move is to repoint that loop at the build handler, and it should **not** be
done: another part of the system is already picking those jobs up, and two consumers of the
same queue is how you get work done twice. I have written that warning into the bug file.

**Where it stands.** The change is committed and has gone to the review council; I have not
read the verdict yet. It does nothing until the next time the system's image is rebuilt and
rolled out — the version running right now predates it — so nobody should read the commit as
the problem being over. There is also a second half to this bug, the job that claimed success
while doing nothing, which the original filer explicitly said to leave until after this fix,
and I have left it.

I have written down the three separate ways the verification could give a false pass, because
I walked into two of them on the previous job today.

---

**Later the same day — the review came back approved, and it caught something I had played
down.** Fifteen reviewers, approved, five advisory notes and nothing serious enough to block.
It came back in about six minutes rather than the half hour I had budgeted.

The approval is not the interesting part. Two of the notes were worth acting on and one of them
corrected me.

**The one that corrected me.** I had written that sending these repairs through the build
handler was "heavier than ideal" — a hammer where a scalpel would do. A reviewer pointed me at
a warning already on file that says something stronger, and having read it, the reviewer is
right: **the build handler cannot see the page's existing words.** It is handed the instruction
("remove the markdown symbols") and nothing to work from, so it writes the section again from
nothing. The old wording is gone. Not corrupted — replaced.

I have corrected my own note rather than quietly reword it, because the difference matters to
you. **"Heavier than ideal" and "the paragraph gets rewritten" are not the same sentence.**

Does that change the decision? I don't think so, and here is the reasoning rather than a
justification. The alternative is not "edit the one field" — that capability does not exist
anywhere in the system today; the same warning says so explicitly. The alternative is what we
have now: the repair fails every single time and the defect stays on the page for ever,
reprinting each time the page is rebuilt. So the real choice is between a rewritten paragraph
and a permanently broken one. For the third of the three checks — pages with no content at all
— writing from scratch is simply the correct thing anyway.

There is also a trap for anyone who tries to improve this: there **is** a setting that looks
like it would make the writer see the existing text, and turning it on makes things quietly
worse, because it feeds it an old snapshot from when the site was first copied rather than
what is on the page now. I have flagged that where someone would find it.

**What this really tells us** is that the precise fix — change one field, leave the rest alone —
needs a piece of plumbing we have never built. I had listed that as a nice-to-have. It is more
than that, and I have written it up properly.

**The other note worth acting on** was a warning that dispatchers in this system have two gates,
and that fixing the one you can see leaves work stuck at the second. That was a fair challenge
and I went and checked rather than argued: the second gate passes everything these repairs need,
and an audit across all 175 agents found no gaps. Clear.

**One note I have left open rather than closed**, because closing it would be dishonest: a
reviewer observed that I have proved the build handler works for *similar* jobs, not for these
exact three. That is true. It is why the verification insists on seeing the page's stored
content actually change, rather than trusting a "completed" label — this bug's own second half
is a job that reported success having done nothing.

**And one for the record.** A reviewer pointed out this is the **fifth** time this same mistake
has shipped — five different checks, each naming an agent that could not do the work, each
fixed by editing one word. The safety net we have only checks that the name is a real agent,
not that the agent can actually handle the job it is being sent. So the sixth one will pass
review too. I have written that up as a proposal with three costed options; the cheapest is a
few hours' work and would catch it at build time. That is a decision for you, not something I
should quietly start.

