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

## 2026-09-02, evening — where this stands, and what I need from you

The new chassis build is running and I checked it carries the guards (it does; the previous one
did too, so nothing changed for this site). I've written the handoff and a runbook so a fresh chat
can pick this up cold: `docs/agent_docs/docs024_key_docs_latest/gamedesign_uk_rebuild/HANDOFF_2026-09-02_continue_here.md`.

The diagnosis is finished. The rebuild is prepared and waiting on you. Before anything gets
dispatched I need: a contact email for the site; whether it should look different from the tools
site; whether to clear the old files first (I'd say yes — they're tool pages, which the new
direction forbids); and a yes or an edit on the brief.

Separately, three things that aren't this site's but came out of the second look: a live "rerender
says done, page still empty" defect on ai-agent-orchestration.com that matches a bug we closed;
who should own the monitoring gap I filed as 432; and seven other domains serving with no
database row. Those are your calls, not blockers for this lane.

## 2026-09-02, ~17:10 — your rulings are done and the build is running

Everything you decided is executed. The site row exists with the email you gave. Theme kits
supplied a look that is the opposite of the tools site on three counts at once — light warm
paper instead of near-black, serif headings instead of sans, an earth-brown accent instead of
cyan — and those values are seeded where the pipeline reads them first. One caveat from theme
kits that you ruled on yourself today: a seeded palette is a starting point, not a lock; the
renderer may move off it, and if it does we look at what it produced rather than clamp it.

The old broken pages are gone from the public internet — every old address now gives an
honest "not found" instead of an empty page — and the build was dispatched at 17:07. It will
take hours; the classifier queues behind the rest of the fleet.

The monitoring gap I filed as 432 now has its check, built and run: it looks at what the
storage bucket is actually serving and asks the database whether it knows each domain. First
run found ten domains serving with no record — three more than I'd counted from the deploy
repo, because the bucket is wider than the repo. gamedesign.uk and oxenunity.com now have rows;
the other eight are your adoption backlog for after this one.

315 is reopened and handed to the ai-agent-orchestration lane. The name clash is with the new
gamesdesign.co.uk session, with "GamesDesign.co.uk" as the recommended name.

## 2026-09-02, about 18:10 — it's live

gamedesign.uk is serving a real site. Four pages — home, about, contact, articles — with proper
copy, a sitemap, your contact address in the footer, and every link working. The homepage opens
"Game design, examined as a practice, not a pitch," which is about as close to the brief as I could
have hoped. The classifier, in its own notes, called it "editorial — not a tool platform; those
live on the sister domain." It took fifty minutes from dispatch to a styled homepage.

The look is what you asked for: warm paper, dark ink, a rust accent, serif headings and body. The
exact colours are not the ones we seeded — the renderer took the classifier's near-identical
values instead and warmed the rest. That is the "starting point, not a lock" rule you set today
doing what it says, and I've passed the numbers to theme kits rather than argued with them.

Two things are parked on purpose and need a word from you: an article slot the planner made with
no article to put in it, and three buttons that have nothing to point at until articles exist.
Both are the system declining to ship something empty. And two things are simply absent: a
favicon, and privacy and terms pages — the plan didn't include them and nothing links to them,
but a public site probably wants a privacy notice. Say the word and I'll get them planned.

## 2026-09-02, about 20:00 — done

Your four answers are all carried out. The article slot is cancelled, there are no privacy or
terms pages, the favicon is live — made from the logo, showing in the tab — and the calculator
question turned out to belong to a different site where the tool already exists. The sibling has
stopped using this site's name and I checked that on its pages. gamedesign.uk has nothing left
waiting on you. What carries on after this lane is the monitoring check I built for the gap that
caused all this (it runs by hand today; it should run on a schedule), and the eight other domains
we're serving without a record of, which you said to adopt later with care.

## 2026-09-02, about 21:30 — you were right, and the cause was mine

You looked at the site and it was wrong for what it is: a games site with no games in it, no game
pictures, an articles page with no articles that explains what the articles would be like, and a
hero that was a dark gradient over a missing file. I had checked that pages existed and links
worked and called it done. I had not read it the way you did.

The cause is plain. The picture guide I wrote for the site banned game imagery — the image prompts
literally say "no game imagery" — and the brief asked for a restrained journal. The system did what
it was told. I've said so in the bug file, added the whole thing to the errors list as its own
category, and rewritten the guide and the brief: games on every page, big bold game art, real
released games named and described as a player sees them, playable illustrations inside articles,
and a rule that no page may ever describe its own brief. The rebuild was dispatched at 20:11 and
picks up tonight's planner change that expects pictures by default.

I ran the checkers as you asked. They saw everything you saw — twenty-seven findings, including
"the articles index writes about itself and lists zero articles" — and filed every one in a mode
that acts on nothing. That is now written down where it can be fixed.

The designblog lane had the same review from you tonight; I joined their routes rather than opening
new ones, and between us and two other threads we found that the "wrong hero image" problem is
fleet-wide: seven components, over a hundred and fifty instances, sixty-odd pages wearing their
homepage's picture. The components thread owns that fix.

Three things need a word from you: the checkers say the contactforsales.com address reads as a
placeholder to senior studio people; the site has no named author and I'm not allowed to invent
one; and there's no newsletter or feed for readers to come back by.

## 2026-09-03, morning — where the rebuild actually is

The rebuild I dispatched last night did not rebuild anything. Re-submitting a site that is already
live runs the research and strategy steps and then stops on purpose — a safety catch so that a
routine refresh can't tear a live site down. Nobody told me and I didn't look; I spent the evening
watching for a plan that was never coming. The new brief and picture guide did land — the
classifier now describes the site as "bold, game imagery, the sensibility of a magazine" — so the
direction is set. I've now queued the actual rebuild deliberately, and it's waiting its turn behind
sites with older backlogs, which the throughput lane says can take hours. Nothing is wrong with it;
it just hasn't been picked up yet.

The fresh build you deployed carries everything this site depends on, including the fix that stops
an articles page shipping with no articles and the components fix for the wrong hero pictures.
The handoff is written for a new chat: docs/agent_docs/docs024_key_docs_latest/gamedesign_uk_rebuild/HANDOFF_2026-09-03_continue_here.md.
Your three questions from last night still stand, plus one from the improvement-loop lane: should
a brand-new site start on "hold" so the machinery can't plant things on it before anyone has read
a page?

## 2026-09-03, late morning — the rebuild ran, and I found out why the articles keep not appearing

The rebuild you were waiting for did get picked up, about an hour after the last note was
written. It wrote a new brief and a new plan. **The plan has four pages and no articles — the
same as before.** So the thing you called a major error is not fixed, and I want to be straight
about that rather than let the imagery improvements make it look fixed.

**Correction to what I wrote last night.** I said the fresh build "carries the fix that stops an
articles page shipping with no articles". That was too strong. What shipped is a *detector*: it
notices an articles page with nothing to list, writes down what is missing, and — because those
pages already exist on your site — deliberately leaves them alone. It records the debt. It does
not create articles. Both of those things happened today exactly as designed.

**The cause, and it is not really a mystery any more.** The planner is the part that decides what
pages the site will have. I read what it actually wrote down while deciding, and it says: the
articles "are satisfied by the blog infrastructure; individual posts are not planned as static
pages here." In plain terms: it assumed some other part of the system writes the articles later,
so it did not plan any. **There is no other part.** Articles on every other site here are just
ordinary pages that the planner planned — your sister site gamesdesign.co.uk has thirteen of them
that were made exactly that way.

**Two things make this worth your attention rather than just mine.** The first is that it is not
only this site: the same planner wrote almost the same sentence about designblog.co.uk the day
before, which is the site you told me suffers from the same problems. So that instinct of yours
was right, and this looks like the shared cause for the articles half of it. The second is that I
cannot fix it by writing a better brief. The brief already told it, in plain words, "the site
launches with real articles, not a description of what the articles will be like" — I checked
that those exact words reached the planner — and it planned none anyway. So the fix has to be
made where the planner's own instructions live, which affects every site, not just this one. I
have written all of it up for the thread that owns that area rather than changing it myself,
because a change there touches all forty sites.

**What I have let happen meanwhile.** The rest of the rebuild is still running — the new game
imagery, the corrected hero pictures, the lighter look. I have let it finish, because the page
layout it is filling in is the one your site already has, so it makes things better and nothing
worse. It just cannot add articles. When it lands I will read the pages and tell you how they
look; the imagery and the wrong-hero problem should both be genuinely fixed.

Your five open questions from last night are all still open. Nothing today needed a decision from
you, and none of them is blocking.

## 2026-09-03, midday — the articles cause is fixed at the source, and I need one decision from you

Two good things and one thing I need you to settle.

**The first good thing: the cause I found this morning is already fixed for the whole estate.**
The designblog lane took the evidence and changed the planner's own instructions — it is now told,
in the prompt, that no later editorial pass will write the articles for it and that it must plan
real launch pieces itself. That went live around midday. Our site's plan was written about twenty
minutes before that, so it does not benefit yet; a re-plan picks it up, and when it does the
articles half of your complaint should come free rather than needing another argument.

**A correction I owe you, because I got something wrong and it travelled.** I said this morning
that the system had no machinery at all for writing articles later. That was too strong, and
another thread caught it within the hour. There *is* such a mechanism, properly built and wired —
it simply has not run since April. So the planner was not inventing something; it was pointing at
something real that stopped working four months ago and nobody noticed. That is a more useful
thing to know, and it is now written into the fix so the wording stays true. My overstatement had
already been quoted into a live change before it was caught, which is my fault for stating an
absence more confidently than I had checked; I have logged it against myself and told the lanes
concerned. Nobody has yet asked why that mechanism went quiet in April — I think someone should,
but it is not this site's job.

**The second good thing: your checks caught a real contradiction, and refused to build.** The
homepage build stopped rather than shipping. The writer had written a line pointing readers to the
contact page, and the rules you set yesterday say this site has no contact page. So it blocked.
That is the machinery doing exactly what you asked it to.

**And that is the decision I need.** Three of your own instructions are now pulling against each
other. Your brief says the site has no contact page, no form and no email. But the contact page
was already built, and the framework deliberately preserves pages that already exist, so it is
still there in the menu. And the rule against mentioning it means any copy that refers to it stops
the build. Nothing can resolve that except you, and until it is resolved the homepage will not
build.

So it is no longer the smaller question I asked last night about whether to keep that
placeholder-looking email address. It is: **do you want the contact page gone, or do you want to
keep it and let the site refer to it?** Either is easy to do. I just should not pick for you,
because both are live and permanent choices about how the site talks to readers.

One smaller thing worth knowing, no action needed. Because the site has no articles yet, the
homepage's "featured article" panel had every one of its six pieces of information come back empty
— picture, summary, category, author, reading time, date. It does not break the page; it just
quietly leaves the panel hollow. A re-plan that creates real articles fills it.

## 2026-09-03, early afternoon — your ruling is in, the site is re-planning

Done, and the re-plan is running.

**Your contact decision is applied.** The contact page stays, and the address is now
gamedesignuk@contactforsales.com. I also had to lift two rules the site was carrying from
yesterday's opposite decision — one banned any mention of a contact page or form, the other banned
any email address anywhere. Those were what stopped the homepage building this morning, so with
them gone that blockage is cleared. I left the third rule from yesterday alone: the site still says
plainly that it is written by an AI, and still may not invent a human editorial team. You did not
ask me to change that, so I did not.

**The site is now re-planning from scratch.** This matters more than usual, because the plan it
had was written about twenty minutes before the fix for the missing-articles problem went live.
This new plan is the first one written with that fix in place, so it is also the first real test of
whether the fix works. If articles come back this time, that is the whole morning's problem solved
at the source. If they do not, I will know within the hour and the lane that shipped the fix needs
to hear it straight away.

**One thing I will have to check afterwards, and I want to flag it rather than quietly fix it
later.** The About and Contact pages already exist and were not marked for rebuilding, and their
current text still contains the old email address. Because we have just removed the rule that
banned email addresses, nothing will now flag the old one as wrong. So it could sit there looking
fine. I will read the actual pages once the rebuild finishes and get it corrected if it has stuck.

**And I have written up the loose end from this morning as a proper bug.** The second, older
mechanism for writing a site's articles ran thirteen times in March and April and then stopped dead
on 24 April. Nothing has noticed in four months. I have not tried to work out why — that is a real
investigation and it is not this site's job — but I have recorded everything I measured, including
the most likely innocent explanation so whoever picks it up tests that first. It is worth someone's
time because if that mechanism were working, the whole problem you spotted this morning might not
arise at all.

Nothing needs a decision from you right now. Your other questions from last night still stand.
