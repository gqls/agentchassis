# SUMMARY 2026-08-03 — the work item that could never be done

**What we're trying to do.** Stop the platform asking itself for work it has
made impossible. Every time the tool generator built a new interactive tool,
it filed a follow-up request — "write some introductory prose around this
widget" — and every single one of those requests, nine out of nine over three
weeks, died unworked in the human-review queue. Worse, one of them silently
blocked two real pieces of work, the same coupling that stalled the whole
fleet twice the day before this lane opened.

**Where we've come from.** The bug was filed yesterday by a triage session
that measured the consequence precisely but honestly declined to guess the
cause. This lane picked it up this morning — after first yielding bug 182 to
another session found twenty minutes ahead of us — and verified the cause: two
sibling code paths create tool pages, one declares which page sections exist
and one forgot to, and both then file the identical "write the sections"
request. The handler looks up the declared sections, finds none, and parks the
request. The request was unsatisfiable at the moment it was created. The
clean proof: the same two code paths also file a "write the companion guide"
request for a page that DOES declare its sections — and those requests
succeed.

**What we've done.** Both paths now go through one shared gate that asks, at
the moment of filing, the exact question the handler will ask later: does this
page declare anything a writer could write? If not, no request is filed, and
the skip is visible in the run's output rather than buried in a log. The
review council approved the plan first round; its four advisories were all
acted on (one caught a real SQL wildcard bug in our cleanup script). The nine
dead requests were retired honestly — marked "won't fix", never "complete",
because completing them would have released their dependents on a lie. The fix
is live in production and proven on the running binaries, not just the tag.

**Where we are now.** Closed. One deliberate loose end is parked in plain
sight: two blocked items were NOT released, because the diagnosis of a
neighbouring bug (178) completed mid-fix and showed that dispatching that kind
of item currently destroys page content — so their dead dependency now serves
as a visible interlock, documented in 178's own file for its owner to release
with their fix.

**Where we're going.** Three threads leave this lane. Bug 187 (filed at the
council's direction) asks which of five other request-emitters have the same
disease — at least 24 more requests are parked with the identical symptom, but
some may be legitimate. The design question the fix deliberately did not
answer — should generated tool pages get prose around the widget at all? —
sits with the owner as TL-009. And the fix's everyday behaviour will witness
itself within days: the next novel tool generated should file no request at
all, and the next library-tool fork should file one exactly as before.
