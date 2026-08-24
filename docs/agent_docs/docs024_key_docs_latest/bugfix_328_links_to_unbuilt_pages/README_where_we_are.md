# Where we are — bug 328, dead links to pages that never got built

Plain prose, append-only, newest at the bottom.

## 2026-08-23, late afternoon — the bug is real, and the reason it survives is not what we thought

The complaint is simple. When one page of a site fails to build, every other page that links to
it still ships with the link. The reader clicks it and gets a "page not found". A site with four
good pages and no link to a missing fifth just looks small; the same site with two dead links
looks broken, and that is a judgement customers make about the whole product rather than about
one page.

I checked whether it is still happening before doing anything else. It is. The loanzy.uk home
page is live and healthy, and it carries three links to two pages that have never existed — two
of them to the same missing page. Same on mortgagecalculator.co.uk, where the missing page has
been advertised from six other pages since late July. Across the whole fleet there are 63 open
records of this, on seven different sites, all pointing at 13 pages that could not be built. The
oldest has been live for a fortnight.

Then the part that changes what we should build. The bug file, updated two days ago by another
lane, says the platform already NOTICES these links — it files a record naming the linking page
and quoting the link — and that the records simply sit in a queue nobody reads. If that were
true, the fix would be plumbing: wire the queue to something.

It is not true, and the rows say so. Those records were picked up. A builder was dispatched at
each one and did run. It failed — 48 times with "no sections ready to build" and 10 times with
"content validation failed" — because the record's only instruction is *build the missing page*,
and the missing page is precisely the thing that cannot be built. So the platform detected the
problem correctly, dispatched correctly, tried the one repair it knows, and that repair is the
one repair that cannot work here. Nobody ever told the pages doing the linking.

The other half of the picture is in the code, and it is a one-line gap with a wide reach. There
is already a shared piece of machinery that cleans dead links out of a page just before it
ships — it runs on the build path, both re-render paths, and at the point where content is
saved. It asks "does this link point at a real page?" and it answers that by looking for a row
in the pages table. A page that was planned and never built HAS a row. So to every one of those
four checks, a link to a page that has never existed on the web looks perfectly fine.

We have already solved exactly this problem once, for the header and footer. That fix asks the
harder question — has this page ever actually been served? — and it is careful about two cases
where the honest answer is "don't touch anything": when the lookup fails, and when the site has
not published anything yet, because on a brand-new site "never served" is true of everything and
means nothing. This bug is the same fix for the body of the page, and the hard part is that
middle case: during a first build, pages go out one at a time, so almost every link points at a
page that has not shipped *yet*. Strip those and you publish a site with no internal links at
all, which is a failure we have shipped before and not noticed. So the design has to tell "not
coming" apart from "not here yet", and it has to do it with something we can actually query.

I have handed that design question to a second model with all the evidence, and I will bring the
answer back through the reviewer council before anything is committed.

## 2026-08-23, evening — built, committed, and with the council; plus two things I got wrong on the way

The fix is written and committed. What it does, in one sentence: just before a page's HTML leaves
for publication, we now check every link it makes to another page of the same site, and if that
page has never existed on the web and is not on its way, the link comes out.

Three things about that are worth saying plainly, because they are the difference between this
working and this being another well-intentioned change nobody notices.

**We do not touch what the page is made of.** The link stays in the stored source. Only the copy
being published loses it. That sounds like a detail and it is actually the whole design: it means
the day the missing page finally gets built, the link comes back on its own, the next time that
page is rendered. Nothing has to remember to put it back. No queue, no repair job, nothing to go
wrong later. It also means the existing detector still sees the link in the database and keeps
reporting it, so we have not made the problem invisible by fixing it.

**Deciding "would this link break" turned out to be harder than it looks, and the obvious answer
was wrong.** The database has a column that says when a page was last published, and the natural
move is to treat an empty one as "never published". I measured it instead of assuming, and nine
pages across the fleet have an empty column and are *serving perfectly well* — they were published
and the column simply never got filled in. Removing links to those would have broken nine working
pages to fix fourteen broken ones. There is an existing, narrower rule the platform already uses,
and that one has the opposite problem: it misses three genuinely dead pages, one of which is the
exact page this bug was filed about. So neither existing rule would do, and I had to find a third
signal. It is whether the page has any actual content stored against it: twenty pages with none
were all dead, nine with some were all alive. Twenty-nine out of twenty-nine.

**The scale is small and precise, which is what I wanted to see before shipping it.** Across every
page on every site, there are 3,193 links between pages. This change removes 36 of them. Every one
of those 36 is a link that gives the reader a "page not found" today.

### The two mistakes

**I measured 56 pages and got the opposite of the truth.** My first sweep said nineteen supposedly
unbuilt pages were serving fine, which would have meant the platform's existing rule was badly out
of date and my whole approach was wrong. All nineteen were one domain — a site parked at its
registrar, which answers "200 OK" for *every* address, including one I made up on the spot to test
it. The check that caught it costs one extra request per site and I nearly didn't do it. That is
now written down as a trap, because the next person to check a list of our URLs will hit the same
domain and get the same confident wrong answer.

**I wrote "fourteen days of live broken links" into our notes when about two thirds of that was
not true.** The platform has 63 open records of this problem, so I quoted 63. Checking them, 42
point at pages that are working now — old records left open by a separate bookkeeping bug. The
real number had to be measured a completely different way. A record in a queue tells you what
something noticed once; it is not evidence about what the website is doing today. That one was
caught by a second model reviewing my write-up, after I had already put the number in a document.

Both are logged in the shared mistakes file, which is the point of having one.

### Where it goes next

The code cannot do anything until the next platform rebuild — that is normal here. After it, there
are three small steps: switch it on (a database change I have written and deliberately held back so
it cannot go live before the code that reads it), re-publish the 24 pages that are serving broken
links today, and then check the result on the actual website rather than in the database. That last
check has a deliberate second half: as well as confirming the broken link is gone, I have to confirm
a *working* link is still there. Without that, a change that simply deleted all the links would look
like a success.

It has gone to the reviewer council and is being reviewed now. I told the reviewers up front about
the one thing I expect them to object to: this is the third time we have solved this same shape of
problem with a purpose-built piece rather than a general one, and our own records flagged after the
second time that the third should probably prompt a wider rethink. I do not think that should stop
this fix, but it should be said by me rather than found by them.

## 2026-08-24 — the new build is out, the switch is on, and the first test page is queued

The platform rebuild landed this afternoon, so the three things that were waiting on it have now
happened.

**First, I checked the new code is genuinely running** rather than assuming it because a build went
out. The service normally announces which version it is at startup, but that message had already
scrolled off, so I asked the running program directly on both machines — and, importantly, asked it
two control questions in the same breath: one thing that must be there and one thing that cannot
possibly be there. Both machines answered correctly on all five checks. Without those controls a
row of "yes" answers is worthless, because a broken instrument says yes to everything.

**Second, I switched it on.** That is a small database change I had deliberately held back so it
could not go live before the code that reads it. It backed up all five affected settings first, made
the change, checked its own work, and I then read the result back with a different query than the one
inside it. All five are on.

**Third, the problem got measurably worse overnight, which is the clearest evidence yet that this
needed fixing.** Yesterday there were 36 dead links across 24 pages. Today there are **48 across 28**
— and a site that did not exist in yesterday's count, garden-tools.uk, arrives with nine of them on
four pages. The bug report predicted exactly this: the count grows with how much work the platform
does, because every new page the framework writes can add another link to a page that was never
built. It is not a backlog that sits still while you fix it.

**One judgement call worth explaining.** My plan said to re-publish all 24 affected pages so the fix
reaches what is already live. Before doing that I checked how recently each had been published — and
26 of the 28 were published *today*. Re-publishing a page brings in every change the platform has
made since it last ran, not only mine, so doing 28 of them would have been a lot of unnecessary
churn on customer sites for something the normal publishing cycle will carry within a day anyway. So
I have queued exactly one: the loanzy.uk home page, which is the original example in the bug report
and was published two hours ago, so it carries almost no accumulated drift.

I have recorded what that page serves right now, so the comparison is honest: three bad links to
remove, and — the part that actually matters — five good links to the calculators page that must
still be there afterwards. A change that simply deleted every link would pass a test that only looked
for the bad ones disappearing.

The page is in the publishing queue and I am waiting for it to run.
