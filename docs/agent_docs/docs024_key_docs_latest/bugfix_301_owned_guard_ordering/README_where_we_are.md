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
