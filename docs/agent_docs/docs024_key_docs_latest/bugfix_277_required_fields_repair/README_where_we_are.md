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

---

## 2026-08-17, late morning — the promoter was letting one broken handler waste work, and we caught it by not believing a graph

The fresh build you deployed is `v1.0.1305`. I checked it properly rather than trusting the
version number: the image itself records which commit it was built from, and I confirmed that
same commit is inside the running program, using a second commit that had to come back *absent*
as a control. It carries the change this lane was waiting on. Nothing here was blocked on it.

Then I re-measured everything before acting on it, and two of the numbers turned out to mean
the opposite of what they looked like.

**The queue looked like it had refilled.** The count of unclaimed findings had gone from 4 to
82 overnight, which reads like the new promoter had stalled. It hadn't. Seventy-seven of those
82 are a kind of finding that deliberately has no handler attached — they are flags for a human
to read, and sitting in that queue is where they are *supposed* to live. Forty of them were put
back there on purpose this morning by another session working a neighbouring bug. Once you
count only the findings that actually have a handler, the number is five, and all five are the
promoter correctly refusing to act until a person has run one by hand. So the mechanism is fine
— but the *measurement* we had written down for judging it was not, because it lumped together
two groups that mean opposite things. I've rewritten that test in the bug file so the next
person isn't misled the way I nearly was.

**The promoter was, however, doing something genuinely wrong.** When we built it we wrote down a
risk: it decides a handler is trustworthy if that handler has ever succeeded at that kind of
job even once. We flagged that as thin at the time. It has now bitten. One handler had
succeeded once and failed twenty-eight times, and the promoter — seeing the one success — kept
handing it more work: six items, five of which failed. That is not a disaster, but it is
wasted machine time and a queue full of failures that look like new problems.

I've closed that door. A handler now has to be succeeding at least a quarter of the time before
the promoter will feed it, and only once it has a real track record — so a brand-new handler
still gets its careful first run, exactly as before. I did not pick the "quarter" out of the
air: I listed every handler's success rate, and there is a clean gap in the data — one handler
at 3%, and the next worst at 41%. Anywhere between those two isolates the broken one and
touches nothing else. As it stands the new rule blocks nothing at all today; it is a door, not
a repair.

**The part worth telling you about is how nearly we missed it.** My first attempt to measure
handler reliability said every single handler had a 100% success rate. That is obviously too
good, which is the only reason I looked again. The cause was a genuine trap in the database:
failed jobs never record a "finished at" time, only successful ones do — so a query that asks
"how did this handler do before now?" using that timestamp silently counts *zero failures for
everybody*. It doesn't error, it doesn't come back empty; it comes back as a clean, uniformly
excellent table. I've written that up as a warning other sessions will see automatically before
they touch the same column, because it would flatter any handler-reliability report we ever run.

**The other thing I want to correct on the record.** One of the three tests we set for declaring
this bug fixed was already passed six days before we built the fix — and both of the last two
updates said it was still pending, because nobody re-checked it, they just copied it forward. If
I'd re-measured a day later I would probably have written it up as proof our fix worked. It
isn't; it was the *old* mechanism, run by hand back in July. The fix is still well-evidenced —
I verified four repaired pages by actually loading them in a browser and reading them, and in
every case the page was written seconds *before* the job was marked done, which is the
fingerprint of real work — but that particular test proves nothing about it and I've said so in
the bug file rather than quietly banking it.

(Small thing, same theme: my first attempt to load those four pages returned "not found" on all
four, which looked briefly like the repairs had never reached the live site. It was my own
mistake in how I typed the addresses. Loading the site's front page first — which worked
fine — is what told me the problem was my question, not the site.)

**Where that leaves things.** The repair router and the promoter are both live and behaving.
The promoter now has both door-closers the review round asked for, plus the one it needed and
nobody asked for. Next: the review round itself, which I'm submitting now with all of this as
evidence.
