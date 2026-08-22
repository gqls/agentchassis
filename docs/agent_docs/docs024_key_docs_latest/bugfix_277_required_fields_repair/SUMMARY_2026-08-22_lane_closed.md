# SUMMARY — 2026-08-22. The lane is closed

## What we were trying to do

Make sure that when the platform finds something wrong with a live page, something can actually fix
it. The lane began with one bug — a whole class of findings (`required_fields_missing`) that no
handler in the fleet owned, so they piled up at "needs a human" for ever — and picked up a second on
the way (`083`: findings written as `detected` that no promoter ever collected, because the only
promoter lived inside a task switched off since May).

## Where we've come from

Both halves turned out to be the same shape seen twice: **detection works, and the estate thins out
at repair.** 083's fix was a promoter of its own, with doors added as the evidence arrived — hold a
pair that has never succeeded, count `verified` as success, report what you held, read archived
history, escalate to a human after three days. 277's was a router, then a repair for the class the
router could not serve.

The hard part of 277 was a correction of our own making. We first blamed page ownership; that was
wrong. The real reason a group of pages could not be repaired was that their **stored content cannot
reproduce the page being served**, so every repair route we had — all of which rebuild from stored
content — was inapplicable by construction. That took three wrong costings to see, and the fix was a
shape the estate did not have: a small mechanical edit to the finished HTML itself.

## What we've done

**083 is closed.** The mechanism ran end to end twice — most convincingly on a route that had not
existed 24 hours earlier: one row promoted by hand at 13:21, complete at 13:25, and the promoter
released the other six on its very next tick with no human involved. Seven pages repaired, verified
at the bytes a visitor is served, and still holding across three chassis rolls.

**277 is closed.** The router is live and classifies everything with nothing unrouted. Both repair
shapes exist and are proven at the served page. The file's own worked example — the one it insisted
must "carry real content" before anyone declared victory — serves its full conversion table and every
value its finding calls missing.

**The backfill is where the lane earned its keep, and it did so by refusing.** Asked to fill in the
missing data for 27 components, we wrote 3. The platform counts a component as rebuildable the moment
its data holds *any one* field the template wants, so filling in some of the data flips a page to
"rebuildable" and the next rebuild blanks everything we did not fill. Partial is not smaller; it is
destructive. So the recovery works backwards from the served page and refuses to write anything unless
re-rendering reproduces those exact bytes. Fifteen could not meet that bar because the component
versions that built them no longer exist. Nine we refused outright: each is a whole working tool
stored in a slot claiming to be a page banner, and filling those in would have armed a rebuild that
destroys them.

## Where we are now

Two bugs closed with their evidence and their caveats — including one mechanism that has never fired
and is described as untested rather than proven, and 41 undispatched items that belong to other bugs
and are named as such so no one reads the close as "the queue is empty".

One new bug is open (`357`, the nine mislabelled tool pages). Its cause is genuinely unknown: the
automated diagnosis came back honestly unable to narrow it, and we recorded that as neither confirmed
nor refuted rather than filling the gap with a plausible story. Two of the three things it said it
lacked we answered from our own measurements; the third — which piece of code labels these slots — is
narrowed to six candidate writers and one clearly-labelled lead.

## Where we're going

Nothing in this lane. `357` is the successor and it has a live, proven precedent to copy: another lane
already solved "a page that is one stored blob" by decomposing it properly, and its scripts carry the
backup convention and the restore path. Whoever takes it should read those before writing new ones,
and should get the cause established first — repairing nine pages while something is still minting
them would be the mistake this estate keeps logging.

The one deliberate loose end is a decision already made by closing 277: the fifteen pages built from
vanished templates stay as they are. They serve correctly today. Rebuilding them would change what
visitors see on four sites, which is a decision about pages, not a data repair — and it is written
down where the next person will find it.
