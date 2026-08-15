# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-15 — the repair handler you asked for this morning is built

This morning you ruled that `required_fields_missing` — the "this page is missing required
content fields" finding — should get a repair handler fleet-wide instead of piling up in the
human-review queue. This session built it.

Before building, we measured what the 44 open findings actually are, and that changed the
design. Most of them (35) are NOT pages missing content — they are pages that serve perfectly
well today, but whose content lives as one stored block of HTML rather than as structured
fields. Automatically "repairing" those would regenerate the section from a template and
throw away the served page — the exact accident we've had before. Six more point at
components that no longer exist. Only a handful are genuinely repairable by a rebuild: one
component with fields that are truly empty, one generic page with no section plan, and your
gas unit converter (a tool page, which the platform deliberately refuses to rebuild with the
generic builder, because that clobbers tools).

So the handler is a router, on the same pattern as the image routers you asked for on the
12th. For each finding it asks the database what is true NOW and takes one of four actions:
if the finding is out of date, it closes it with the evidence written on it; if the fields
are genuinely empty, it files a targeted rewrite that edits the existing copy rather than
replacing it; if the page has no plan and is safe to rebuild, it files that rebuild; and for
the two classes that genuinely need a human (the stored-HTML pages, and tool pages like the
gas converter), it parks the finding back in your review queue — but now carrying its
classification, the evidence, and the safe options, instead of sitting there as a bare
mystery. Parked findings are pinned so the system cannot keep re-raising duplicates of them.

State right now: the classification was dry-run over all 44 findings and every prediction
checked out; the change went to the council for review; the handler is written and ready to
seed. Next steps are: seed it (inert), run four representative findings through it as a
canary, then point the remaining 40 at it. The gas converter itself will come back to you as
a parked decision naming the tool pipeline as the repair route — this handler routes it
honestly rather than overriding the no-generic-rebuild-of-tools rule.

## 2026-08-15 evening — built, live, backlog routed; two decisions need you

The repair handler is done and working. Every one of the 44 parked findings has now been
through it or is queued for it: the dead ones are being closed with the evidence written on
them, and the ones a machine must not touch (the stored-HTML pages, the tool pages, the
image-slot fields) are parked back in your review queue carrying their classification, the
danger, and the safe options — instead of sitting there as bare mysteries. New findings of
this type route themselves from birth now (that change rode another lane's release this
afternoon). Along the way the review process caught two genuine design errors before they
could do harm — a repair that would have asked a prose writer to invent image addresses
(the system refused it, and we made that refusal a routing rule), and a page rebuild that
would have quietly produced nothing (measured, and made a routing rule too).

The council reviewed this four times and improved it each time, but it has not approved it,
and the remaining objections are ones the reviewers themselves disagree about — which by our
own rules means they are yours to settle, not mine to argue again:

1. **One reviewer insists new findings should be born "unclaimable" and promoted by the
   triage stage; another accepts what we did.** The trouble is the triage stage's machinery
   has been switched off since May (bug 083) — honouring the contract today means findings
   sit stranded forever, which is what you ruled against. If you want the contract honoured,
   the real fix is rebuilding the promoter, and that is a separate piece of work. My change
   is one line to revert the day that happens.
2. **This is the third single-purpose router of its kind** (the two image routers from the
   12th are its siblings). Several reviewers want one shared engine instead of a growing
   family. I filed that as RFC_030 with the case for doing it as a consolidation of all
   three; it needs a decision on whether and when to schedule it.
