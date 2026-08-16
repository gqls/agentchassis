# Noted rebuild — LIVE. 16 August 2026

**What we're trying to do.** Rebuild noted.co.uk — a note-taking app holding
people's private writing, voice recordings and photographs — as a site fully
controlled by the framework, with a server-side backend so a person can sign in
and reach their notes from any browser, without losing anyone's only copy of
anything.

**Where we've come from.** The old app was browser-only: every note lived in the
one browser it was written in, served as static files from a bucket. A week ago
there was a backend and nothing else — no site, no editor, no migration path.

**What we've done.** The framework built the site unattended. We fixed a silent
delivery failure that had it reaching the repository but never the server. We
built the migration — a page that reads the old app's notes straight out of the
visitor's browser and hands them back as a file, no account needed — and the
editor itself, designed around one clause: it may say "Saved" only when the
server really has saved. Both are tested, and each protection was deliberately
broken once to prove the tests catch it. The owner wrote the privacy copy and it
is on the site word for word. Today the domain was cut over.

**Where we are now.** **noted.co.uk serves the new site.** Ten pages live,
including the editor, the rescue tool, and the privacy page. The old app is
preserved at `/legacy-app/` on the same address, so it still sees anyone's notes.
Proven on the live site after cutover: a note written the old way is found by the
rescue tool and comes back with its text and recording intact; sign-in, save and
retrieval from a second browser all work; a failed save is loud and loses nothing.
The commercial shopfront sharing the machine was unaffected throughout.

**Where we're going.** Watching, not building. Let the grace period run and decide
when to retire the old app. Two small things worth doing before real sign-ups: the
service has no way to delete an account, and the experience patterns are recorded
as `proposed` rather than `verified` until the checking loop runs them green.
