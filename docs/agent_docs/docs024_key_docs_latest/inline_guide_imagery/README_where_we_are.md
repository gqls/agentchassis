# Where we are — pictures inside guide articles

Plain-prose running log, appended to, newest at the bottom. Started 2026-08-31 when the
lane was picked up; the design work behind it dates from 2026-08-14.

---

## 2026-08-31 — picking this up, and the one thing that is actually in the way

The ask, from a fortnight ago and repeated this week: our guide articles should have
pictures **inside** them — beside and between the paragraphs — not just the one banner
image at the top. The example you gave was the darts grip-styles guide, which talks about
ring grip, razor grip and shark grip and shows none of them.

Two things I expected to be the hard part turn out not to be.

**The pictures themselves are no longer a problem.** The current image model draws darts
correctly. Where our images are wrong they are old ones from July, made by the previous
model, and the lane looking after that site has spent today replacing them.

**The building block already exists.** A week ago another thread built exactly the piece
this plan asked for: a page section that is "some prose plus its own illustration", with
the picture held outside the writing so a rewrite of the words cannot take the picture with
it. Another change on the 26th taught the page planner to notice that this block can hold a
picture at all, which it previously could not distinguish from a plain block of text.

**What is actually in the way is smaller and more annoying than either.** The system can
attach *one* picture to a page, and it knows which page a picture belongs to — but it does
not know which *section* of that page a picture belongs to. The information is recorded
(each picture is filed against a page and a position within it), but the part of the code
that hands pictures to sections throws the position away and gives every section on the
page the same picture. So a guide with six small sections would show the same photograph six
times.

There is a second, related gap: when a page has several sections of the same kind, the
safety net that normally preserves what a page already has stops working for them, because
it identifies sections by type rather than by which one they are. Two sections of the same
type look identical to it, so it gives up and protects neither.

Both are small, both are in one file, and neither has ever mattered before — because until
now no page in the estate has had more than one illustrated section that the framework
itself produced. The one page that does have several (a bees homepage on another site) was
put together by hand, and by my reading its six pictures will disappear the next time
anything rewrites that page's words. I have not asserted that as fact: I have put it
through the diagnosis loop for an independent read rather than trusting my own reading of
the code, and will record what comes back.

So the position is: the parts are built, the pictures are good, and the missing piece is a
join between a picture and a section. That is the next thing to fix, and it is the whole
reason the guides still look like walls of text.

## 2026-08-31 (evening) — the join is built; the pictures themselves are the next job

The missing piece I described this afternoon is now written and committed. A picture filed
against "page X, section 3" is now actually handed to section 3, on both of the two routes
that rebuild a page — and it had to be both, because the second one would otherwise have
undone the first the next time anything re-rendered.

Two things are worth saying plainly about how it was done, because they are the parts that
could have gone quietly wrong.

**I proved the fault before fixing it.** It would have been easy to read the code, believe
it, and write the fix. Instead I wrote a test that asks two sections for two different
pictures, ran it against the old code, and watched it hand back the same picture twice.
That is the difference between "I think this is the bug" and "here is the bug". The same
test now passes, and a second one guards every page that already has a single picture, so
we can tell this change apart from one that quietly broke them.

**I sent it to the diagnosis loop first and it did not agree with me.** It came back "not
confirmed", and it was right: I had bundled two related faults into one description and
offered a page as evidence that only demonstrates one of them. That cost a few minutes and
saved a wrong claim from going into the record. The loop's objection is written up in the
technical notes, with my own error named.

The change is committed but will not do anything until the next fleet build — it is Go
code, so it sits inert until then. The review council has it and I will read the verdict.

**What is still missing is the pictures.** Nothing yet asks for one picture per section of a
guide: fleet-wide there are four such requests in total, across three sites, and none on the
darts site. So the next job is the supply — plan a figure for each small section of the
grip-styles guide (the one you pointed at, six subsections, no images at all today),
generate them, and watch one guide come out right end to end. That is the first thing where
you would actually see a difference on a page.

## 2026-09-02 — it is live, and the reviewers were right about two things

The change shipped with the fleet build on Monday evening, so the join between a picture and a
section now exists in the running system. I checked that by asking the running program whether it
contains the new code, with a control in both directions — the check our own notes recommend for
this (reading a start-up log line) turned out not to work on this service at all, and I have
written down what does.

**The review came back asking for changes, and it was right twice over.**

The first point was the good one. I had argued, at length, that you must not identify a section by
a position number because the system holds two incompatible numbering schemes. Then I wrote the
fix — and the identity of a section ended up being worked out in three separate places, from three
different lists. Same trap, one level along, in my own change. Where those lists disagree — a
section moved, deleted, renamed — the picture does not go missing, which would be obvious. It
lands under the **wrong heading**, and the page renders perfectly.

That is now blocked: the system compares the two lists and, if they disagree at all, simply
declines to attach pictures per section and behaves as it did before. And I proved it by breaking
it on purpose: with the check disabled, the test fails by putting the shark-grip photograph on the
section about ring grips. That failing line is the whole argument.

The second point was more embarrassing and cost nothing to fix. There was already a piece of the
system for counting "which repeat of this thing is this one" — and on one of the two code paths it
was sitting two lines above the one I hand-wrote. I had not looked. It is now used everywhere.

**Where this leaves the actual ask.** The plumbing is done and live. What is still missing is the
pictures: nothing yet asks for one image per section of a guide. The nearest real page is the bees
homepage on another site, which already has six per-section illustrations that a single rebuild
would silently delete — I have written to that thread with the exact fix, which now works for the
first time, and left the decision with them since it is their site. For the darts guides you
pointed at, the remaining step is a page rebuild that composes the guide into small sections and
asks for an image for each, and that is a change to a live page on a site another thread is
actively working, so it wants their agreement rather than my initiative.

## 2026-09-02 (afternoon) — approved, and what the reviews were actually for

The change is approved. It took three rounds, and I want to record what those rounds cost and
bought, because the honest answer is not what I expected.

I would have called the first version finished. The reviewers found two things in it that were
real. The first is the one that stings: I had argued at length that you must not identify a
section by a position number, because the system keeps two incompatible numbering schemes — and
then wrote a fix in which a section's identity ended up worked out in three separate places from
three different lists. The same trap, one level along, inside the change that was fixing it.
Where those lists disagree the picture does not vanish; it lands under the wrong heading, and
everything about the page looks correct. That is now blocked, and I proved it by disabling the
block and watching the test put the shark-grip photograph on the ring-grip section.

The second was simpler: a piece of the system for counting "which repeat of this is this one"
already existed, and on one code path it was sitting two lines above the one I wrote by hand.

The second round failed on something worse, in the sense that it was purely my carelessness: the
first round had told me off for describing a file in prose instead of putting it in the reviewable
list, I fixed that for the file it had named, and then did the identical thing with another one.
The reviewer quoted my own confession back at me. Fixed, and the third round passed.

**Something useful fell out of the waiting.** While the last review ran I measured what a
detector this lane has been planning would actually catch today, rather than assuming it would
catch nothing. It found two live cases on other people's sites: an illustration that was planned,
generated, deployed and is being ignored because the section asks for "an image" and the system
hands back the page's own banner picture — so that page shows its banner three times and its
illustration never — and a second where the field asks for nothing at all, so no part of the
system writes it and the space stays blank for ever. Both are contributed to the existing bug that
covers exactly this, with the queries, and left with the people who own those sites.

**Where the actual ask stands.** The plumbing is approved and half of it is live; the rest goes
out with the next fleet build. The guide pages you asked about still have no pictures in them,
because nothing has yet *asked* for one per section — that step is a rebuild of a live page on a
site another thread is working, and I have given them the exact recipe rather than doing it over
their heads.

## 2026-09-02 (evening) — the honest finding: I spent the effort on the half that wasn't blocking

Two things happened after the approval, and the second one is the important one.

**Someone used it.** The bees-homepage thread read the note I sent them, checked it, and seeded
the six records that make their six illustrations durable — about seven hours after the safety
half went live. That is the first real use of this machinery by anyone.

⚠ **But it has not actually run yet, and I nearly reported it as if it had.** Those records are in
place; the page itself has not been rebuilt since August, so nothing has yet gone and fetched
them. The machinery is loaded and untested. It gets tested the next time that page is rebuilt —
and properly tested the next time its words are rewritten, which is the event this was all built
to survive. I would rather tell you "armed, unproven" than let "someone is using it" stand in for
"it works".

**And then the darts thread asked the question I should have asked a fortnight ago.** They pointed
out that not one page on their site is *shaped* to hold a picture between its sections — all 22 of
their articles and guides are the same three pieces: a banner, one solid block of text, and a
button at the bottom. So I measured it across everything we run:

- **442 article and guide pages.**
- **9** have any section that can carry its own picture.
- **1** has more than one — i.e. exactly one page in the whole estate could show a different
  picture beside each section.

**So the thing standing between you and illustrated guides was never the plumbing I have spent
this fortnight on.** The plumbing is finished, reviewed and live. What is missing is that the
system composes an article as one undifferentiated slab of text, and there is nowhere to put a
picture even when we have one. Seeding more records cannot fix that; there is nothing to seed
them against.

The other thread working on imagery arrived at the same conclusion this afternoon from a
completely different direction — they were asking why some articles get built with no banner at
all — and neither of us was looking for it. That agreement is the most useful thing to come out of
today.

**What that means for the ask.** Getting pictures into the darts guides now means changing how
those pages are *built*, not adding any more capability: re-planning an article into several
smaller sections that can each hold one. That is a change to live pages on a site another thread
owns, they know it is theirs, and they have resized their own work accordingly. It is also the
first page in the estate that would be built that way, so it is an experiment rather than
following a pattern — I have said so to them plainly rather than handing over a confident recipe.

## 2026-09-02 (later) — correcting the entry above: the system HAS started using the new building block

An hour after writing the entry above I checked one more column and it changes the conclusion, so
here it is straight away rather than buried in a later update.

I told you that only one page in 442 can hold a picture beside each section, and implied that the
system never reaches for the piece that can. **The first half is true. The second is wrong.**

The change made on 26 August that taught the page planner to notice this building block **worked**.
Since that date the planner has chosen it on **six sites in seven days** — one of them five hours
after the change went in, and two of them today. It is being adopted steadily and nobody had
noticed, including me, because I counted how many pages have one and never asked when they
appeared.

**What is actually missing is narrower and much cheaper than I made it sound.** Of the nine pages
that use it, eight are *home pages* using it once, as a single illustrated panel. Not one is an
article or a guide. So the planner has learned to reach for an illustrated section once on a
landing page, and has never been asked to build an *article* out of several of them — which is
exactly what your guides need.

That is a question about how the planner composes articles, not about anything missing from the
system, and it now has a working precedent to point at rather than being a from-scratch ask. It
also lines up with what the other imagery thread found today from a completely different angle:
whatever the planner is doing with articles specifically is where the remaining problem lives.

**The correction I'd want you to take from this:** my previous note would have had you believe we
build things nothing uses. On this occasion that was my error, not the system's — the counting was
right and the conclusion I drew from it was not.

## 2026-09-02 (late) — why the new sites have no pictures in the text: the planner was told not to ask

Your critique of the designblog remake reached me this evening — not enough images, infographics
and graphics, on that site and the two beside it, with eighteen more queued. That thread asked me
what to change. I think I can now tell you the actual reason, and it is smaller and more fixable
than anyone expected.

**The system is doing what it was told.** The instructions we give the page planner list five
kinds of picture it may ask for — logo, banner, illustration, icon, infographic. Then, a few
paragraphs later, the same instructions say: *"Use sparingly — most plans will have zero"* for
anything attached to a specific section, set the required minimum at a logo and a banner per page,
and give a worked example whose section pictures are **only icons**. Infographics are permitted,
never required, and never shown in the example.

Across everything we have ever built, the planner has asked for **399 banners, 211 icons, 50
logos, 25 illustrations and exactly one infographic.** That is not a broken system. That is a
system obeying an instruction nobody had read next to the numbers.

**So the fix is a paragraph of English, not an engineering project** — but it belongs to whoever
owns that planner, it affects every site we build from then on, and it costs real money in
generated images, so I have handed them the exact wording rather than changing it myself.

**One honest caveat, because it decides what you'll actually see.** Changing that instruction will
put pictures on the *home and landing* pages, where there is somewhere to put them. It will not
put pictures inside your articles and guides — because, as I found earlier today, an article is
built as a single unbroken block of text with no slots between paragraphs. Those are two separate
problems and I would rather you heard that now than saw one of them fixed and wondered why the
guides still looked the same.

I have said the same to both threads working on this, so nobody reports the first as progress on
the second.

## 2026-09-03 — there is a third layer underneath, and it explains why the guides cannot be fixed from here

I have twice told you the bottleneck had moved. It has moved once more, and this is the bottom of
it — so it is worth setting out plainly what the three layers are, because the answer to "why are
there no pictures in the guides" turns out to be all three at once.

**Layer 1, which I spent a fortnight on: could a picture stay put?** Yes, now. Built, reviewed
across three rounds, live and verified in the running system. This was the half I was asked to fix
and it is done.

**Layer 2, found yesterday: does anything build an article out of small illustrated sections?** No.
An article is composed as a banner, one solid slab of text, and a button. Across 442 article and
guide pages, exactly one page in the whole estate is built the other way, and it was made by hand.

**Layer 3, found this morning by the darts thread and checked independently by me: are the
articles even in the plan?** Mostly not. **Around 84% of articles, guides and tool pages have no
entry in their site's plan at all** — against 2% of home pages and 15% of ordinary content pages.
The split is by *page type*: the structural pages are always planned, the article pages almost
never are. A route builds them from a layout held in the code and never writes the plan down.

That third layer is the one that makes the other two unreachable, because everything I built keys
off the plan. No plan entry means the machinery politely stands aside and the page renders as it
always did.

**And it corrects something I told you yesterday.** I said nine of the thirteen darts blog pages
were ready for this. They are — but only because that thread hand-wrote those plan entries once, in
July, in a single batch, from a note saying *"nobody ever decided what blocks a guide page should
contain."* The fourteen articles written there since have no entries. So my number described one
person's tidying-up, not the system working, and I should not have presented it as the latter.

**What I would want you to take from this.** Getting pictures into the guides needs three things
in order: the articles have to be *planned*, then *composed* into small sections, then the pictures
*attached*. The third is mine and is finished. The second belongs to the other imagery thread. The
first does not appear to belong to anyone, and it is the one that blocks the rest.

## 2026-09-03 (later) — the shape of the missing first step, and the cheap fix that would fool us

Following the third layer I described this morning — most articles never get an entry in their
site's plan — the darts thread and I have now pinned down *why*, and what fixing it would actually
mean. Worth a short note because there is an obvious cheap answer here that would look like
success and quietly stop being true.

**Why articles have no plan entry.** The routine that creates an article writes the page and its
layout into a *cache*, and never into the plan itself. That is not a bug in the sense of something
broken — it simply was never asked to. Only two pieces of code in the whole system write the plan,
and neither of them runs when an article is created. So an article can only ever get a plan entry
if some later, larger re-planning run happens to sweep it up.

**The cheap fix, and why I do not want it presented to you as the fix.** There is a third route:
writing the plan entries by hand, in SQL. It has been used fifteen times across the estate, and
one of those occasions in July is the *entire reason* the darts guides looked ready when I reported
on them yesterday. Doing that again would repair the pages that exist today, in an afternoon, and
change nothing about how the next article is born — the fourteen articles written on that site
since July have no entries, and the fifteenth would not either.

So: **hand-backfilling is a reasonable way to unblock one canary page, and it is not the first
phase.** If it ever reaches you described as "articles are planned now", the right question is
"planned by what, and what happens to the next one?"

**Where that leaves the three steps.** Get articles into the plan (nobody owns this; it needs a
change to how articles are created, not a backfill) → compose them into small illustrated sections
(the other imagery thread) → attach a picture to each section (mine, finished and live). I would
rather you had the honest shape of it than a report that the part I own is done — which it is, and
which on its own changes nothing you can see.

## 2026-09-03 (afternoon) — the pictures work, and now you can see the next problem clearly

Your grip-styles page rebuilt itself this afternoon and I watched the whole thing. **The part I
built works, end to end, on your own page.** That is the first time I can say that without
qualification, so it is worth being precise about what was proved.

The page now has eleven sections instead of three, and five of them carry a different photograph:
ring grip, razor grip, shark grip, smooth barrel, combination grip. Each one is the right
photograph in the right place. I opened all five and checked them by eye — correct dart anatomy,
proper flights, no archery feathers, no screw threads. All five load on the live page.

**The important bit is what happened next, by accident.** Just over an hour after the page was
built, the last of the five images finished generating, and that automatically triggered the whole
page to be rewritten from scratch — every heading, every paragraph, new words throughout. **The
five photographs stayed exactly where they were.** That is the entire thing I was asked to fix in
August. The old page kept its words and any picture in one single field, so rewriting the words
destroyed the picture. Now the pictures are held separately, in the site's plan, and they are
re-attached every single time the page is built. A rewrite happened and cost us nothing.

**And now the bad news, which is on the same page and is not my code.**

The photographs are right. The words next to them are wrong. On the first build, all five sections
were written about the *ring* grip — five times, under five different and correct photographs. So
the section showing a deliberately smooth, polished barrel had a heading about bands of cuts. The
rewrite an hour later replaced that with five near-identical headings all saying some version of
"what your fingers feel", none of which names the grip it is sitting next to. The hidden
descriptions that screen readers announce are wrong in the same way.

**Why.** When the framework asks the writer to produce a section, it hands over a note saying what
that specific section is about — "ring grip: evenly spaced circular grooves", "shark grip: angled
directional cuts", and so on. The darts thread wrote those notes carefully and checked they were
all different. I confirmed today that every one of them reaches the writer. **And then the writer
is never shown them.** The instructions the writer actually receives do not include that note at
all. So it gets five requests that look identical, and writes the same section five times. It is
also never told what is in the photograph, only the file's address.

This is already a known, actively-worked problem — another thread found it independently last week,
predicted in writing that exactly this would happen, and the fix is one prompt change that is
sitting waiting for you to read it. I have added what this page shows to their file rather than
starting a competing account. **What is new is how much worse it is than it looked.** Their examples
were pages where the same paragraph appeared three times: repetitive, obviously fixable, nobody
misled. Here the framework did its half correctly, so identical words became *false captions on
correct pictures*, and the better my half gets, the worse that becomes. Right now two pages in the
estate can suffer it. That number goes up every time this lane succeeds.

**There is a second, quieter lesson in it.** The good first version — where the headings did name
each grip — only happened because the darts thread hand-wrote a long instruction into the rebuild
request. The automatic rebuild an hour later did not have that instruction and had no way to get
it. So a page someone crafted by hand silently got worse, with nobody touching it, in seventy
minutes. The only per-section detail that survives an automatic rebuild is the kind that lives in
the plan — which is the same argument that made the pictures durable, applied to the words.

**Where that leaves your original ask.** You asked for an accurate picture inside each small
section of a guide. The pictures are accurate, they are inside the sections, and they survive
rewrites. The words beside them will be right when that one prompt change lands, and that is not
mine to land. I would rather tell you it is half done and say which half than report success on
the picture and let you find the caption yourself.

## 2026-09-03 (evening) — the darts thread has taken the page back down, and they were right to

A correction to what I wrote a few hours ago, while it is still fresh.

**The grip-styles page has been reverted.** It is back to the three sections it had this morning:
banner, one slab of text, button. The darts thread did it and told me straight away, reporting what
actually happened rather than what either of us hoped. **I think it was the right call and I would
have made the same one.** That page exists to earn search traffic so the site can get affiliate
approval, and seven sections that all say roughly the same thing work directly against that. The
pictures were the only part behaving; everything written around them was wrong.

**What this does and does not change.**

It does not undo what was proved. The page really did build with five different correct photographs,
one per section, and it really did survive a full rewrite of its words an hour later with the
pictures intact. I watched both, and I read them from the system's own records rather than from the
look of the page. That happened, and taking the page down afterwards does not un-happen it.

What it does change is that there is now nothing to go and look at. Anyone reading my note from this
afternoon would go to the page and find the old one. So I have marked every document accordingly.

**The darts thread also proved the writing problem better than I did, and I want that on the
record.** I showed the writer's instructions never mention what each section is about by reading the
configuration. They went one better and read the **actual instructions that were sent**, which the
system stores for every request, and found that four sections which had been given four different
descriptions received one identical instruction — the same request, byte for byte. That is proof
where mine was an argument. I had reached for the second-best evidence without noticing the best
existed, which is worth me writing down because I had spent the afternoon telling other people to
check exactly this sort of thing.

They also measured the damage further than I did. It was not just five similar headings: **every one
of the sections re-wrote the whole article**, each explaining ring, razor and shark grip again from
the start. Six times over.

**One thing we disagree about, and I am not going to pretend otherwise.** They believe the durability
test — can a picture survive the words being rewritten? — never happened, because they took the page
down before anything rewrote it. My records say a full rewrite did happen, automatically, an hour
after the build, and that the pictures were re-fetched from the plan rather than merely left alone.
I have sent them the evidence and asked which of us is reading it wrong. It matters because that
question is the entire point of the work, and I would rather have it settled by them than agreed too
quickly by me.

**Where this leaves you.** Nothing is lost. The five photographs still exist on the site's asset
list, so the page can be rebuilt in minutes the moment the writing problem is fixed — and that fix
is one waiting prompt change belonging to a third thread. Until then, rebuilding it just produces the
same seven repetitive sections again, which is why nobody should.
