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
