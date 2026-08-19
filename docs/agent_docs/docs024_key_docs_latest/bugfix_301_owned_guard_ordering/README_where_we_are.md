# Where we are — bug 301 (the platform pays for work it then throws away)

*Owner's plain-prose log. Append-only, newest at the bottom.*

## 2026-08-19 — lane opened

The bug, in plain terms: some pages on our sites are marked "owned" — they belong to an
interactive tool or widget, and the generic page-building pipeline is not allowed to
overwrite them. There is a guard that enforces this, and the guard works. The problem is
where it sits: it is the *last* step of the page-build workflow. So when the system decides
to rebuild one of these protected pages, it first pays for the expensive part — an AI call
that writes all the page content, plus a link-resolution pass — and only *then* asks "am I
allowed to save this?", gets told no, and throws the work away. On one site that was 39
paid-for-and-discarded builds in a single night.

Two other threads have already done useful halves of this: one made the refusals visible at
all, and one made a refusal stop counting as a "failure" against the handler's competence
score (that shipped this morning). What nobody had done yet is move the question to the
front of the queue — ask "am I allowed?" *before* paying for the AI call. The thread that
found the bug offered it to whoever picks it up, so this lane is picking it up.

The fix is small and deliberately cautious: the step that loads the page's details (right
at the start of the workflow) learns an optional switch — "refuse owned pages here". Only
the page-build workflow turns the switch on. The tool pipeline, which is *supposed* to work
on owned pages, never gets the switch and doesn't change. The late guard also stays, as a
belt-and-braces backstop. Right now there are 146 queued jobs that would each burn a wasted
AI call under the old ordering, so the fix has a real queue waiting for it.

I checked the bug is still happening today (it is — I can see this morning's wasted chains
in the database), filed the independent diagnosis run to double-check my reading of the
workflow before committing to it, and drafted the change. Next: the automated council
review, then commit, so the fix rides the next fleet release.

## 2026-08-19, later — the fix is live and working, and the reviewers made it better documented

The new fleet build went out early this afternoon and it carries our change. Within about
ninety minutes the fix did its job for real, three times: three repair jobs aimed at protected
tool pages were turned away instantly — at the "load the page" step — with the right closing
status, and crucially the expensive AI writing step never ran for any of them. Before today,
each of those three would have cost a full AI writing pass that was then thrown away. Ordinary
pages carried on as before: their builds still run the writer normally (some of them are being
rejected later by an unrelated, long-standing content-quality check, at about the usual daily
rate — not something our change touches).

The automated review came back asking for revisions rather than approving first time. Its main
catch was fair: when I applied the database half of the change, I skipped the house rule that
says "take a snapshot of the live configuration before you modify it". No harm was done —
I happened to have saved an exact copy of the old configuration minutes before, and that copy
is now safely in the repository, plus a proper snapshot has been taken — but the reviewer was
right that I'd relied on luck rather than the rule, and I've written the mistake up in the
shared mistakes log so the next person avoids it. The other review points were mostly "show us,
don't tell us" — the actual committed files already had the safety checks the reviewers wanted;
my summary just hadn't shown them. I've resubmitted with everything they asked for measured and
attached, and the second verdict is due shortly.

The independent double-check I'd ordered on my reading of the workflow came back "ran out of
steam — neither confirmed nor refuted". That's now moot in the best way: the live system has
demonstrated the behaviour directly, which is stronger evidence than any code reading.

What's left before this can be closed: read the second review verdict, and see one ordinary
page build complete end-to-end on the new version (none has yet, only because of that
unrelated content-quality rejection plus low traffic — worth one more look tomorrow).

## 2026-08-19, evening — closed

Both outstanding reads came back the right way, so this bug is now closed and its file has
moved to the closed folder.

The second review verdict is **approved** — eleven reviewers approved, none raised anything
serious, and the three advisory comments were questions I could answer by checking rather than
by arguing. One reviewer asked whether our fix leans on a mechanism that an older open bug says
is unreliable; that older bug was closed three weeks ago. Another asked for an independent
confirmation that the workflow engine really honours the setting the way we placed it; I read
the engine's code and it does — it checks our placement first. A third asked which pods had
been checked and whether the image could be a stale cached one; I re-checked this evening and
every one of the 22 pods that can run this workflow is on the same build, down to the exact
image fingerprint, and the two main pods (which had been replaced since the earlier check)
both carry the change.

And the thing I was waiting to see happened: two ordinary page builds have now gone all the
way through on the new version — writer, save, publish — so "ordinary pages still work" is
observed, not inferred. Meanwhile a fourth tool page was turned away at the load step this
evening with no AI writing run, and this one left a fresh review row stamped by the new step
(the earlier three had been deduplicated onto existing rows, which is why I'd only had the
indirect evidence before). Refusals at the old, late position since the roll: zero.

**One decision for you, which I have flagged and deliberately not taken.** This fix stops the
waste; it does not stop the *cause*. Twelve places in the code still send "rewrite the content
of this page" jobs to the generic page builder without first checking whether the page is a
tool page the builder is not allowed to touch. So tool pages keep accumulating jobs that can
only ever be refused — 142 of them sitting in the queue tonight. That upstream defect is now
the "not addressed here" footnote of TWO closed bug files (295 and this one), and no open bug
file carries it as its subject, which is exactly how a known problem gets buried. Options:
(a) file it as its own small open bug now, cross-referenced to the Tier 2 repair discussion, or
(b) hand it to the Tier 2 / copy-quality lane that owns the adjacent "how do we actually repair
an owned page" question. My recommendation is (a): "jobs are routed to the wrong handler" is a
routing defect with a small fix; "how should these pages be repaired" is a design question —
different bugs, and the first should not wait for the second.

## 2026-08-19, later — you chose (a), and it is filed

Filed as bug **333** (open, unowned). Two things I got more precise while filing it, and one
I got wrong above and am correcting here rather than editing it away:

- **"Twelve places" was an undercount.** My first search skipped two folders. The real figure
  is thirty places across twenty-five files, plus one agent whose routing lives in its
  configuration rather than in code. None of them checks whether the page is a tool page before
  sending the job to the generic builder — and, as a control, the same search *does* find the
  two producers that already do check, so "none" is a real result, not a blind one.
- **The biggest offender is the tool pipeline itself.** For four months, the code that builds a
  tool page has been asking the generic page builder to write the prose around the widget — on a
  page the same pipeline has just marked as "not the generic builder's to touch". That is not a
  forgotten check; it is two rules in conflict, and the bug file names it as a separate design
  decision for the tool lane rather than trying to settle it.
- **What the fix should look like.** The shared place where every work item is written already
  has one "is this routable?" check at the door (for handlers that don't exist). The natural fix
  is a second check beside it: if the handler is the generic builder and the page is owned, don't
  file the job at that handler — record it visibly as "needs a route", per finding, so it can be
  counted and acted on rather than silently refused. The bug file orders four candidates and says
  why that one closes the door and the others don't.
- **The boundary with the other lane.** The 277 lane reached the same residual from the other
  side tonight ("an owned page with a real defect has no route at all"). 333 owns "stop sending
  jobs to a handler that refuses"; 277 owns "what route can actually repair an owned page". I
  have left a note in their file pointing at 333 so neither lane re-derives the other's half.
