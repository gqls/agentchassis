# Where we are — the unclickable article index (bug 309)

*(Owner's plain-prose log. Append-only, newest at the bottom.)*

## 2026-08-18 — picked up, cause found, and it's a family not a one-off

The Platform Log page on fundamentallyai.com lists six articles as cards — but not
one of them is clickable. The cards look finished (picture, date, title, excerpt)
and just have no link. Another session found and filed this yesterday; today's
session took it on and traced it end to end.

What happened: the card section on that page was built months ago by an AI
component generator, and the generator told the cards to fetch their links from a
data drawer called "blog" in the site's records. **That drawer has never existed —
not on this site, not on any site.** The rest of each card gets written by the
content writer, so everything else filled in fine; the link is the one part that
had to come from the missing drawer, and the system's rule for a missing value is
"quietly leave it out". So the page shipped looking complete and linking nowhere.

Two things follow. First, the fix for this page is easy and clean: we already have
a properly built article-list component that gets its links by asking the database
"which articles does this site actually have?" — we swap the page onto that and
rebuild it. That also fixes the card that currently advertises an archived article,
because the ask-the-database route only lists real, live pages.

Second, and more important: we checked whether other components make the same
mistake, and **ten other invented "drawers" are referenced by eleven components**,
plus seven made-up database queries that don't exist either. Most of those
components aren't currently on any page, but three are (one page on the
leopardess site, two on gaswholesalers). Nothing stops the generator inventing
more. So the real fix is a rule at the door: when a new component is stored, its
data sources must be checked against what actually exists — made-up sources get
sent back with a message saying what the real options are. That's the change going
to the council.

A machine diagnosis of the same page was fired by the previous session and is
still queued; we'll read it when it lands as a cross-check on our tracing.

## 2026-08-18, evening — the door has a lock now; the page fix is with you

Since this morning's note: it turned out a second session had picked up the same
bug at almost the same minute — it confirmed the same cause through the automatic
diagnosis loop and is the one currently asking you which way to fix the page
itself (its "candidate 1", rebuilding the card list to ask the database for the
site's real articles, is the one we'd also recommend — it fixes both affected
sites' pages and the archived-article card in one move).

This session built the door-lock: new components are now checked, the moment they
are stored, against the real list of data sources the platform can actually
serve. A component asking for a data drawer that doesn't exist anywhere gets sent
back with a message naming the real options, instead of shipping and failing
silently months later. The review council asked for two hardenings before
approving — proof the check is genuinely called (we now have a test that fails if
anyone unplugs it — verified by unplugging it), and a permanent record whenever
the check has to run half-blind because the database was briefly unreachable.
Both are done and the revised change is back with the council. The lock takes
effect on the next software release.

One honest limit, written down where the next person will look: the lock guards
the machine-generated route. A component added by hand-written database script
walks around it — the same gap that eventually forced a database-level rule for
an earlier problem of this shape. If phantom sources reappear via hand seeds,
that is the next move.

## 2026-08-19 — the lock is live and proven; this lane's work is done

The new software release went out this morning and we checked it the careful way:
the running service's own binary carries our changes (checked with both a
should-be-there and a should-not-be-there probe, so the check itself can fail).
The door-lock is live. It hasn't been triggered by a real generation yet — that's
expected; the first component that tries a made-up source will be its live proof.

Overnight the other session applied the page repair (the card list now asks the
database for the site's real articles — and, satisfyingly, that repaired component
passes our new lock by construction). But re-rendering the page hit a different
safety guard, and the guard is right: five of the eight articles have no summary
text to put on their cards. Digging into why turned up something much bigger than
this page — more than half the site pages across the whole fleet have no summary
text, because no part of the pipeline was ever asked to write one. That's filed as
bug 320 with the options laid out for you.

So: this lane is closed — the read-out is in SUMMARY_2026-08-19. The one decision
waiting on you is bug 320's: how summaries should get filled in. Once you pick,
finishing the Platform Log page is two mechanical steps written out in the bug
file, and any session can run them.

---

## 2026-08-22 — the last piece of this bug, and one honest gap it leaves you

Picking this up fresh today, the first job was to check whether anything here is still
true, because both halves of this bug were declared done three days ago and this tree
moves fast. Both held. The Platform Log page still serves its eight articles with working
links on every card, and the lock we put on component creation has not let a single bad
one through since it went live.

But there was a door we never shut, and it is the bigger one.

**What the lock does and does not do.** When the system invents a new page component, it
declares where each piece of its data comes from. Our lock checks those declarations at
the moment the component is created, and refuses one that points at data the platform
cannot supply — which is the exact fault that left six articles with no links on them for
four months. The trouble is that components do not only arrive that way. They are
routinely created or edited *directly in the database*, by hand or by a migration, and
those never pass the lock at all. It is not a hole we can plug: it is simply not on that
door.

**So I went and counted what is already inside.** Sixty-nine fields, across seventeen
components, are asking for data that does not exist and never has. Six of those
components are live on forty-six real pages right now. Each one is a piece of content
quietly not appearing — no error, no broken layout, just an absence that looks like a
design choice. That is the same silent failure this bug is named after, still happening,
in seventeen more places.

**What I have built is a nightly check that asks the same question of everything already
in the database.** The important part is that it does not re-implement the rule — it
*calls* the very same code the creation lock uses. That sounds like a detail and it is
the whole point: two copies of a rule drift apart, and then you have two answers and no
way to know which is right. One rule, two doors.

**The part that needed care was not letting it become useless.** Sixty-nine existing
problems cannot be fixed in one go — each needs a judgement about that particular
component, and fixing *one* of them last week took a decision from you and ran into a
safety guard on the way. But a check that is red every single morning is a check people
stop reading. So the sixty-nine are recorded in a frozen list that the check treats as
"already known". Three things stop that list becoming a way of hiding problems: it is
written so precisely that it excuses only the exact sixty-nine and nothing else — add a
new fault to one of those same components and it goes red immediately; it refuses to
accept new entries at all, so nobody can quiet a future problem by adding a line; and it
only affects the pass/fail light, never the report, so the check still names all
sixty-nine every morning whether or not anyone has acted. I tested all three by
deliberately breaking them and confirming the tests caught it.

**The gap I am leaving you, stated plainly.** The check is written, reviewed and
committed, but **it has not run yet** — its image has to be built and shipped with the
next fleet release. I want to be clear about why that matters more than it sounds: on
this cluster, a job whose image is missing shows up as *still running*, not as failed. So
if anyone applies it early, it will sit there looking perfectly healthy having never once
executed. I have written the build-then-deploy order and the checks that prove it really
ran into the runbook, and I have left the bug open on that single step rather than
calling it done.

**One more thing worth your time.** The sixty-nine now have their own file, bug 362, with
the six live ones listed first. Nothing about them is urgent in the sense of breaking a
page — they have been like this for months. But they are real missing content on live
sites, and now that they are counted and watched, they can be worked through steadily
rather than rediscovered by accident the next time someone notices a card with no link.

**A note on how today went, since it is the kind of thing worth knowing.** Two separate
sessions collided on one file this afternoon and briefly broke the shared build; the other
session spotted it, fixed it the careful way rather than the fast way, and messaged me
about it. It also caught a claim of mine that had gone stale within hours of my writing
it. Both got fixed properly and both are written down. I mention it because it is the
system working as intended, not despite the collisions but through them.
