# SUMMARY — the prevention holds, and the cleanup turns out to be a content job (2026-08-14)

*Duplicate-pages front of the brochure lane. Follows
`SUMMARY_2026-08-09_the_first_look_paid_off_and_the_machine_half_is_blocked.md`, which is the
last read-out on this lane; that one covers the camera front and is parallel to this, not
superseded by it.*

## What we're trying to do

Some of our sites carry the same page twice. There is a plain version at an address like
`/llm-cost-calculator.html`, and a second version under `/tools/`, and both are live, and the
machine has been quietly maintaining both. It is untidy for a visitor, it splits whatever
search traffic the page earns, and it means our own records disagree with the website.

There are two halves to fixing it. **Stop new duplicates appearing**, and **clean up the
seven pairs that already exist.** This read-out is mostly about discovering that those two
halves are very different sizes.

## Where we've come from

The prevention half was built earlier and switched on quietly for one site, fundamentallyai,
so it could be watched before being trusted. It is still watching: it has not fired once,
because nothing has asked it to — that site's page list has not been rebuilt since. That is
the expected reading rather than a disappointment, and it is unchanged from the last summary.

The cleanup half is where everything since has happened, and it started with a nastier
discovery: **pages we had deliberately retired were coming back to life.** Two on
fundamentallyai had been retired by hand and were serving to the public again days later.
That mattered directly, because the whole cleanup plan is "keep one version, retire the
other" — and if retiring does not stick, the plan undoes itself.

## What we've done

**We found out why retired pages were returning, and it was not what anyone assumed.** Four
separate parts of the system can rebuild and republish a page, and none of them checked
whether the page was retired. The telling detail: for one page the part we suspected actually
behaved *correctly* — it stopped and asked for a human to review — and sixteen minutes later
a completely unrelated part rebuilt and published it anyway. A fix aimed at the obvious
culprit would have bolted a door that was already shut.

**We fixed it, and it is live.** Two locks: one stops a retired page's file reaching the
website, the other stops our database recording a publication that did not happen. Both were
needed — with only the first, the records would still have lied. It went through the internal
review panel, which approved it, and it is running in production on both servers.

The review was worth having. It doubted one of my claims because I had proved it with a text
search rather than a proper audit. Re-checking properly, that claim held — but the
neighbouring one would not have, because a helper writes that value in a form no text search
could find. Harmless here, but I would have been leaning on it.

**We measured how widespread the problem was, and I got that wrong first.** My initial check
said eighteen pages. Visiting them showed **five**. The other thirteen were retired properly
months ago; our database keeps a "last published" date that never gets cleared, and I was
reading that as "live now". Had we "fixed" the problem and measured success with my query, we
could have claimed a drop from eighteen to thirteen having changed nothing at all.

**You decided all seven pairs**, and then reversed two of them when I found that a step I had
relied on does not exist — see below. Both rulings are recorded.

**We executed the first pair, and the platform stopped us — correctly.** The page was archived,
then the request to remove its file was refused: *"still linked from live content"*. It is
linked from the site's navigation, from the footer that appears on every page, and from a blog
article. Nothing was deleted.

## Where we are now

**The prevention is done and holding.** Retired pages can no longer be republished. It has not
had to refuse anything yet, so it is proven present rather than proven firing — the first real
test will come from the cleanup itself.

**The cleanup is blocked at the first pair, and the blockage is the interesting part.** The
refusal was not a bug; it is deliberate design. Our platform treats a navigation entry as a
pointer it can safely fix by itself, but treats a link inside written copy — including the
site footer — as *content*, which it will not rewrite on its own authority. So finishing even
one pair needs a change to the footer and a change to an article.

**That resizes the whole job.** The first pair was chosen deliberately as the easiest of the
seven: no plan surgery, and the version being retired was a 727-word stub with no working
calculator in it. It still needs a footer edit and an article edit. There is no reason to
expect the other six to be tidier, and three of them need extra work besides. The procedure we
wrote describes eight mechanical steps per pair; that is not what this is.

**Two things we relied on turned out not to exist.** There is **no redirect mechanism** — the
step that said "point the old address at the new one" writes to a table nothing reads, so
retiring an address simply produces a "not found". That is what made you reverse two of the
seven, and you were right to. And the check that told us "nothing links to these pages" reads
a table that is **empty across the entire company**, so it returns that reassuring answer for
every page that has ever existed. The first pair we actually tested had three inbound links.

Right now the first pair's older version is marked retired but still serving, with the
navigation and footer still pointing at a page that works. Nothing is broken for a visitor.
One command puts it back exactly as it was.

## Where we're going

**The question for you is whether the cleanup is still worth its price.** It is no longer a
tidy-up; it is a content project across four sites. The case for doing it is unchanged — the
duplicates are untidy and split traffic. The case for leaving them is that they are stable,
nobody has complained, and the prevention now stops the problem growing. That is a genuine
choice rather than a formality, and it is the reason this read-out exists.

If we do proceed, the shape per pair is now known: mark the page retired, ask the system to
remove it (which refuses and hands back the full list of what links to it), repair those links
through the content pipeline, ask again, then check the result on the live site. The one thing
that cannot be skipped is the repair, because the platform will not let us create a dead link.

Two smaller threads stay open. The prevention will not be considered finished until it has
actually refused something once — the cleanup is the natural way to make that happen. And the
finetuning site's pair stays decided but untouched, because that site has a separate known
fault that has to be fixed first.
