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

## 2026-09-02, later — you asked for a second look on Fable. Here is what it changed.

I ran the whole investigation again on a different model, told it to measure before reading my
conclusions, then grade me. The big picture held: the adoption wiped the pages, the rerender
published the empty shells, that specific day's damage was unique to gamedesign.uk, and nothing
in our monitoring can see a site that has no database rows. All reproduced independently.

Seven details were wrong, and I've corrected each one in place with a note saying what caught it.
The ones you'd care about: it's five empty pages, not six (the other two are missing pages, not
empty ones), and across the whole directory it's actually thirteen empty files out of forty-seven,
not just the ones in the menu. Two of the "never populated" pages I named didn't exist yet on the
day I was measuring — my script read a missing file as an empty one. And I called three things
"guards" when only two of them actually refuse to publish; the third just tries to help a page
build. None of that changes the conclusion, but it was wrong and it's fixed.

One correction of the correction: it told me there are 19 site directories in the deploy repo,
and I'd said "20 or more". I counted properly. It's 36, and eight of them have no database row.
So I re-check what a second investigator tells me too.

**The thing it found that I had missed is the one worth your attention.** I'd said this class of
defect was closed because the publish step now refuses to write an empty page. That's true — but
only for new pages. If an empty page is *already* out there and its row has no content blocks,
every rerender since then has quietly "completed" without touching it. There's one live right now
on ai-agent-orchestration.com — the ROI estimator page — with eight completed rerenders since late
August and still an empty body. Two more empties from the same April wave are still serving on
other sites. Those are other lanes' sites so I haven't filed on them; I've recorded them in the
bug file and I'm flagging them to you here. That's a "rerender says done, page still broken"
problem, and we have a closed bug (315) whose description matches it exactly.

Two small extras from the second look: the original hand-written homepage for gamedesign.uk
survives as a stray untracked file in the deploy repo, if the rebuild wants to see what the site
used to say; and about half of gamesdesign.co.uk's page titles still say "GameDesign.uk" — that's
the sibling site's problem, and its lane's.
