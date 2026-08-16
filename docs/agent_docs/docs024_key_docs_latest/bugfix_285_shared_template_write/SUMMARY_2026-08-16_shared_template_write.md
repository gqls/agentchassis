# SUMMARY 2026-08-16 — bugs_closed/285, the tool-improver's shared-template write (read-out at closure)

**What we're trying to do.** Make it impossible for a fix aimed at ONE page to rewrite a template
that many pages share — the accident that armed 115 pages twice in nine days — and make sure the
one place it did reach a reader is repaired.

**Where we've come from.** The 281 lane built the fix on 15 August (a fence at the writer, producers
re-routed to human review, the loader pinned to the page); it was council-approved but not running.
The bug file believed no page had ever served the bad markup and that one other writer was still
unfenced.

**What we've done.** Found that one page HAD served it — a learn article emptied and given fake
download links because the improver's delivery went to an arbitrary page — and restored it from the
platform's own archive, byte-proven, then re-deployed and checked the live page. Confirmed the new
build carries the fence, then made the fence refuse on purpose (a request carrying the template's
own bytes) and read the refusal in the error log. Turned the hand-written census of template writers
into a build-time test that fails if a new writer appears unfenced or an exemption goes stale;
council-approved; broken on purpose two ways to prove it bites. Corrected three claims in the
record: the "unfenced writer" was mis-described; the "recovery" version was the poison; and "no tool
has a plan" was counted in the wrong table (87 do).

**Where we are now.** 285 is closed: fixed, live, seen refusing, casualty off the site. The measured
"or worse" branch is on record: one regeneration of that shared template drove 154 re-renders and 73
attempted LLM rebuilds of imported pages, all refused by the owned-page guard downstream.

**Where we're going.** One decision for the owner (README, PLAN §D2(b)): whether findings on the
63 imported tools should be fixed automatically per PAGE by a writeback at the finding's scope —
designed, not built, because building it without that decision leaves a mechanism nothing
exercises. The 281 lane's first post-roll sweep census is still theirs to take.
