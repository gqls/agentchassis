# SUMMARY — noted.co.uk, 2026-08-25

**What we're trying to do.** Make noted.co.uk a notes product a person can
trust with anything — words, photos, GIFs, video, voice — on any device, with
every page framework-owned and every promise on the site backed by a mechanism
that actually exists.

**Where we've come from.** A hand-built, browser-only app with no accounts. We
rebuilt it on the framework, gave it a real engine with accounts, cut the
domain over on the 16th of August, and retired the old app with nothing lost.
Until this week a note was text only.

**What we've done.** In the last two days the owner's pasteboard vision went
live in two stages. First, media: any photo, GIF, video or audio can be pasted,
dropped or picked into a note, plays or displays in place, and can be removed —
under the same honesty contract as the text (nothing claims to be stored until
the server says so). Second, the board: an "Arrange on board" view where the
text and every media item are tiles you drag and resize — built for touch
first, and the same arrangement scales between a phone and a desktop. Under it
all, media bytes moved from the server's disk to a private Backblaze bucket
with a tightly-scoped key, which is what makes paid storage a settings change
later; the 50 MB allowance stays as the valve. Everything is proven at the
artefact: mutation-verified test suites on both halves, and a live rehearsal on
the real site — a picture uploaded through the page into the real bucket, read
back by a second browser, deleted, bucket confirmed empty.

**Where we are now.** The product does what its front page implies: writing,
media, arrangement, on phone or desktop, with degraded states that fail loudly
and lose nothing. The formal review thread on the header button is closed with
every advisory answered. Two things are promised or implied but not yet built:
closing an account (planned in detail, waiting on two owner decisions —
immediate deletion versus a grace period, and the honest sentence about
backups), and stage 3 of the pasteboard (editing media in place: crop, rotate,
captions).

**Where we're going.** Build account deletion once the owner picks; then stage
3 if the owner still wants it after living with the board; and when real users
arrive, the paid storage tier the Backblaze move was for.
