# Where we are — gamedesign.uk

Plain-prose log, append-only, newest at the bottom. Owner's document.

---

## 2026-09-02 — what was actually wrong, and why

You asked me to look up the old threads for gamedesign.uk and fix it, then said the first
thing to fix is why the adoption broke it. Here is the whole story.

**The site is serving empty pages, and has been since April.** Six of its nine pages give
you a header, a footer, and nothing in between — including the homepage and the tools page,
which was supposed to be the whole point of the site. Two more pages, privacy and terms,
are linked from every footer and don't exist at all. This has been live to the public for
about four and a half months.

**The platform doesn't know the site exists.** There is no record of gamedesign.uk in the
database — no site, no pages, nothing. The files are just sitting in the storage bucket
answering requests on their own. That is why nobody noticed and why nothing fixed it
automatically: every repair mechanism we have starts from a database row, and there isn't
one.

**What the adoption did.** Back in April we ran the adoption process on gamedesign.uk —
the thing that reads an existing site and rebuilds it in our system. It was pointed at
gamedesign.uk as both the source and the destination, so it read the live site and wrote
back over the top of it. The first thing it did was wipe the existing pages and recreate
them as empty placeholders, ready for the content step to fill in. That is by design; the
pages are meant to be empty for a few minutes.

**The mistake was that the empty placeholders got published.** The rerender step ran on
those blank pages, produced a page consisting of a header and an empty body, committed it
over the good HTML, and deployed it. At the time nothing checked whether a page had any
content before publishing it. The content step that was supposed to fill the pages back in
never finished. So the blank version is what the public got, and it stuck.

I can show you the exact moment. In the site's own file history, the homepage had 5,977
characters of content on 14 April and zero on 16 April, in a commit called "Rerender:
index.html" that deleted 278 lines and added 6.

**It wasn't a general bug — it was specific to this site that day.** Seven sites were being
rerendered on 16 April. I checked every HTML file touched on all of them: gamedesign.uk
lost content in 4 of its 11 files, and the other six sites lost content in 0 of their 139.
So the rerender machinery was working fine everywhere else. It was the adoption running
against this particular site that caused it.

**The April thread saw it and misjudged it.** Its handoff note lists the empty pages as
problem "P3" and says the cause is that the content jobs hadn't run yet, and the fix is to
let them finish. That was a reasonable read at the time. But the blank pages were already
public by then, and the content jobs never completed, so "wait for it to resolve itself"
turned into four months of a broken site.

**The good news: this can't happen again.** Three separate guards have gone in since, all
after the damage:

- 12 May — the renderer now refuses to publish a page that is just a header, an empty
  body and a footer.
- 8 June — a page that arrives with no sections during an adoption borrows its layout from
  a sibling page instead of rendering blank.
- 27 July — a page that has content blocks but renders to nothing now fails loudly instead
  of quietly reporting success.

The first one is the one that would have stopped this. I read the actual code rather than
trusting the comment, and it does what it says.

So the defect you asked about is real, is understood, and is already closed. gamedesign.uk
is a casualty from before the fix that nobody could repair afterwards, because deleting its
database row took away the only handle anyone had on it.

**Two things are still open.** First, the site itself is still broken and still public —
fixing that is the rebuild, which I'd started setting up before you redirected me. Second,
there's a gap in our monitoring: we have a check that finds retired pages still serving to
the public, but it works by looking at pages marked "archived" in the database.
gamedesign.uk's rows were deleted outright, not archived, so the check cannot see it. A
whole site can disappear from our records and keep serving to the public with nothing
noticing. That looks worth filing as its own bug — I haven't yet.

**On the rebuild direction.** I asked the portfolio positioning thread as you suggested,
and they came back with a clear steer: gamesdesign.co.uk stays the "authority" side — free
calculators and guides for solo devs and students — and gamedesign.uk takes the
professional practice side: how working studios actually run game design. Process,
workflow, balance sign-off, pipelines, hiring, tooling reviews, opinion. Written for leads
and producers rather than learners. They've asked that it avoid the free-tool and
guide-library formats entirely and link to the sibling instead, so the two don't compete.
They also flagged domains to stay clear of, and have written the register rows.
