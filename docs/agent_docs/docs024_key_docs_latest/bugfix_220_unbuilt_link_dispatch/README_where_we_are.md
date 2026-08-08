# Where we are — bugfix 220 (plain prose, append-only, newest at the bottom)

## 2026-08-08 evening — picked up, plan settled

The improvement loop had a nasty habit: when it found a link on a live page pointing
at a page that was never built, it "fixed" it by rebuilding the page the link sits on
— which was never broken — and then marked the job done. The missing page stayed
missing, the link stayed a 404, and the next sweep found it again and did the same
thing. Every cycle looked green and cost a full page build.

The team that found this (while closing bug 206) wrote it up but didn't take it, so
this lane has. Three things need doing and two of them are small config/code changes:
tell the dispatcher to hand the worker the ID of the page the job is actually about
(it already stores it — it just never passed it on; the sibling dispatcher already
does exactly this); teach the page loader that when that ID is handed over
explicitly, it wins over the page name (today the name wins, which is the quiet root
of the whole thing); and register a completion check so a job of this kind can only
be marked done when the missing page is genuinely live, or the link genuinely gone.
That last piece slots into the verification framework another lane made strict
today, so the timing is good.

One deliberate deferral: pointing directory-type pages at the new directory builder
(rather than the generic page builder) is left as a recorded follow-up — with the
completion check in place, that case now fails loudly instead of lying, and the
directory team's own machinery is the proper fix for it.
