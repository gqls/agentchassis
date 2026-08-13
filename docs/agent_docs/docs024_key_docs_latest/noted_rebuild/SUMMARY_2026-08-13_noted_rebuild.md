# Noted rebuild — where things stand, 13 August 2026

**What we're trying to do.** Rebuild noted.co.uk — a note-taking app holding
people's private writing, voice recordings and photographs — as a site fully
controlled by the framework, with a server-side backend so a person can sign in
and reach their notes from any browser. Do it without breaking the live app or
losing anyone's only copy of anything.

**Where we've come from.** The old app is browser-only: every note lives in the
one browser it was written in, and it has been live since January. By the 11th we
had the backend built, public and tested at app.noted.co.uk, backups encrypted
and drilled, and the framework build dispatched but not started.

**What we've done.** The framework built the entire site unattended overnight —
five pages, imagery, re-renders. We found and fixed a silent delivery failure
(the box's sync fetched only the shopfront's folder; now it derives its list so
sites can't drift out of it again). Every call-to-action now points at the real
app. We built the migration page through the framework's own tool path: it reads
the old app's notes straight out of the visitor's browser and hands them back as
a file, no account needed — tested against real browser data, with proof it
cannot write, delete, or leave anything behind. The owner wrote the privacy copy;
it is on the site word for word, registered in the framework as the only
permitted wording. A companion guide explains the migration, written under a
corrected instruction after the original guidance was found to be teaching
writers the exact phrase it banned.

**Where we are now.** Eight pages built and deployed to our own machine —
including privacy and the rescue tool — all served correctly, none public. The
live domain still shows the old app and its wind-down notice, untouched. The
site's work queue is empty. The plan, structure and page records the framework
needs for future upgrades and maintenance are all in place and accurate.

**Where we're going.** Three things stand between here and launch: the owner's
decision on whether the privacy page discloses backup retention for deleted
notes; a by-hand test that a failed save fails loudly rather than losing text
(the one protection for the unrecoverable case); and the cutover itself —
repointing noted.co.uk from the old bucket to the box, keeping the old app
reachable for a grace period, and re-running the browser-storage probe on the new
origin the moment it flips.
