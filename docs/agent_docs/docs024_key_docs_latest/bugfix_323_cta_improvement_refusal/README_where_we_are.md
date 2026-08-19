# README — where we are (owner's plain-prose log, append-only, newest at the bottom)

## 2026-08-19, afternoon — picking up bug 323

**What the bug is, in plain terms.** Five of our auditing agents look at a site and sometimes
say "the buttons on this page are wrong" — the two buttons go to the same place, or the label
says "browse the catalogue" and the link goes to a calculator, or the labels are all "Learn More"
when the brief asked for task-specific verbs. Every one of those findings is handed to the same
repair agent, the component-template-fixer. That agent has, since March, had a line in its code
that says, for this kind of job: "I can't do this — it needs a writer, not a CSS patch — mark it
for review." And then the job is stamped **complete** anyway, because nothing on the completion
path reads that "mark it for review" message. 993 jobs over five months, across 22 sites; not one
was ever actually done by that agent. And once a job is complete, the same finding is suppressed
if the auditor raises it again.

**Is it still happening?** Yes. 34 of these in the last week, 12 sites, all closed green.

**The thing the earlier lane missed, and it matters for the fix.** The 302 lane, which found this,
wrote that the fixer's "I did nothing because it was already right" and "I did nothing because I
can't" look identical in the data. They don't. The fixer has always marked its *refusals* with a
separate flag (`action: needs_review`) and its *nothing-to-do* results without one. I checked every
row it has ever produced: the flag appears on 470 refusals and on zero nothing-to-do rows; the
nothing-to-do reasons ("already has flex CSS") appear 299 times and never carry the flag. So the
handler has been telling us clearly all along — we just never wired a listener. There is even a
comment in the code claiming the flag "stops the dispatch loop recording the work item as done".
It doesn't. Nothing reads it.

**Did the work get done some other way?** Partly, and only for one kind. Where the finding is "the
two buttons go to the same place", a separate deterministic mechanism (the internal-link resolver
and its "stale CTA links" rerender) fixes destinations on its own schedule — robot-hands.com's
homepage was corrected about two hours after the auditor flagged it, by that route, not by the job.
Where the finding is about the *words* on the button, nothing touches it; the only thing that
might is a whole-page rewrite triggered by some other finding, which is accidental.

**What I've done so far.** Re-measured everything (archive-inclusive), read the real code path end
to end, confirmed nobody else owns this, and fired the diagnosis loop (~16:00Z) to check my read
before I commit to it. Set up this lane's docs.

**The shape of the fix I'm heading for — three layers, smallest blast radius first:**
1. *Today, config only:* teach the fixer's own workflow to honour its own flag — when it refuses,
   park the job visibly for a human instead of completing it. This is exactly what the page
   builder already does ("park the work item visibly instead of letting the dispatch loop stamp
   it complete"). It covers every refusal the fixer makes, not just this one.
2. *Next chassis build:* stop sending CTA findings to an agent that refuses them. Until a
   capable agent exists, file them the way the estate already records "found work I have no
   handler for" (your July ruling on bug 077) — a roadmap row, not a dispatch.
3. *Next chassis build:* a build-time check so that a routing rule can never again point a
   category at a fix type the fixer's own code refuses. This is the part that stops it recurring.

**What I am NOT doing, and why.** Not building the writer that would actually do CTA copy work.
That is the same missing piece two other lanes (277 and 301/083) asked the copy-editor lane
about *today* — an LLM that turns "this component is wrong in this way" into a small field edit
for the section-editor. It's wanted in three places now, which is the argument for building it
once, properly, with the owner's say on whether it may write to live pages without a human. I'll
tell that lane this is a third customer.

**Open question for you (not blocking):** when the fixer refuses a job, do you want it parked for
a human (my layer 1, matching the page builder) — knowing the human-review queue is already ~980
rows — or would you rather it simply be closed as "won't fix — no handler" so it doesn't sit
there? I've gone with parked-visibly, because that is the estate's existing answer and it keeps
the finding's detail for whoever builds the writer.

## 2026-08-19, evening — done today, in plain terms

**The diagnosis loop agreed with my reading** (confirmed, three passes, same lines of code).

**The quick half is live and proven.** The fixer's own workflow now reads its own "I can't do this"
flag and parks the job for a human instead of closing it. I didn't just change the config and
read it back — I sent the real fixer a throw-away job and watched it: the job went from "claimed"
to "needs human review", the error text says why, the maintenance note reads "refused:
cta_improvement" with the handler's own reason. Then I deleted the throw-away job and its note.
Measured beforehand: this would have caught all 470 historical refusals and blocked none of the
299 genuine "nothing to do" results.

**The durable half is committed and waiting for the next chassis build.** CTA and navigation
findings stop being sent to the fixer at all; they become the estate's standard "we found work we
have no handler for" record (your ruling on 077), with the auditor's suggestion and acceptance test
kept on the row for whoever builds the handler. And a build-time test now refuses to let anyone
route a category at the fixer for a fix type the fixer's own code turns down — I proved the test
bites by putting the old routing back and watching it fail. The council review is running; the
commit carries the "submitted" trailer.

**Told the copy-editor lane** they now have a third customer (after 277 and 301) for the small
"one component, one defect → field edit" writer, with the CTA demand measured (34 a week across 12
sites) and one trap a CTA writer must know (button *destinations* are owned by the link resolver
and get re-resolved on render — a writer should edit the *text*, not the URL).

**One thing for you to be aware of:** until the chassis rolls, CTA findings still go to the fixer
and now land in the human-review queue (that's the honest state, and it's already ~980 rows).
After the roll they become one "no handler" roadmap row per site per category instead. If you'd
rather they went straight to the roadmap row now, that is a one-line config change I can make —
say the word.

**Still open:** read and act on the council verdict; verify the Go half after the roll.

**20:31Z — the review panel approved it, first round, no serious objections.** Two small
things they asked for I did straight away: the test now calls the fixer's own "which fix type is
this?" logic instead of keeping its own copy (so the two can't drift), and the hand-off with the
283 lane — whose pending change touches the same agent — is now written down where they will find
it, not just in my prose. One thing they flagged that I'm keeping as a stated trade-off: after the
roll, the "no handler" roadmap row keeps only the first CTA finding's detail per site; later ones
on the same site are counted but not stored. That is how the 077 shape works everywhere, and the
auditors re-raise every run, so the detail comes back the moment a real handler exists. Nothing
left for you to decide tonight.
