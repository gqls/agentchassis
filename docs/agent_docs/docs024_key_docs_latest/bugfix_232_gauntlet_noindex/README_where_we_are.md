# Where we are — the Gauntlet round pages and search engines

Plain prose, append-only, newest at the bottom.

---

## 2026-08-09, afternoon — what this is and what I did

The Gauntlet is the thing on vonc.com where a visitor argues a point against the AI and
gets a verdict. If they like the result they can press share, and the round becomes a
permanent public page. Another thread noticed this morning that **nothing was stopping
Google indexing those pages**, and wrote it up as bug 232. I picked it up.

It matters less for what it is today than for what it becomes. Right now the published
rounds are basically our own test traffic. But the moment a stranger publishes a round
that argues about a real, named person, that person's name becomes a search term that
returns our page. A link someone hands you is one kind of problem; the same words coming
back from a name search is a much bigger one. And the fix costs us **nothing** — shared
links keep working exactly as they did, which is the whole appeal. We only lose search
discovery, which for this page we never wanted.

**The two obvious ways to do it both turned out to be impossible**, and finding that out
was most of the afternoon. The tidy way is to have the web server add a "don't index
this" instruction to the response. We don't have a web server in front of that page —
the site is static files sitting in storage with Cloudflare in front, and no code of
ours runs when someone asks for the page. There *are* some server config files in the
repo that look like they'd do the job, but they're for a different service on a
different machine; using them would have produced a fix that never runs. The second way
is to put the instruction in the page's own template. That doesn't work either: the
template for that page is only the middle of the document, and the instruction has to go
in the header section, which comes from somewhere else entirely.

So the fix goes where the header is actually assembled. There's a function that builds
each page's header, and it already does four things of exactly this shape — it puts in
the title, the description, some structured data, and a canonical link. I added a fifth,
and it's switched on by a new per-page flag that is **off by default everywhere**. One
page in six hundred and thirty has it on. Every other page comes out byte-for-byte the
same as before, which I measured rather than assumed.

**The thing I'd want someone to push back on**, and I've asked the reviewers about it
explicitly: I made this a small change rather than a big one. There's a real, larger
problem sitting next to it (below), and I deliberately didn't fix it. That's a judgement
call and reviewers may say I drew the line in the wrong place.

## The uncomfortable thing I found on the way

There are **two** different bits of code that build a page's header, and only one of them
gets the improvements. The other one is still very much in use — three of our agents use
it. So the last three header improvements we made (structured data in July, canonical
links a week ago, and now this) all only work on one of the two paths.

For this bug that's more serious than it sounds, because it means the database can say
"this page is hidden from search" while the actual page has nothing of the sort on it,
if it happened to get rebuilt by the other route. **The record and the reality can
disagree, and nothing would tell you.** I've written that up prominently in two places
so the next person hits the warning before they hit the surprise, but I have not fixed
it — it's a bigger architectural question about whether those two bits of code should
become one, and quietly folding that into a small bug fix is exactly the kind of thing
we've agreed not to do.

## Where it stands right now

The database change is **live**. The code is **written, tested and committed but not yet
running** — it only takes effect the next time the whole system gets rebuilt and
redeployed, which happens on its own schedule. After that someone needs to trigger a
rebuild of that one page and then check the live page actually carries the tag. I've
written down the exact commands, including the traps (you have to defeat Cloudflare's
cache to see the truth, and you have to check the page's body rather than its headers,
which is the mistake the original bug write-up half-makes).

There's also a second, smaller half: the round data is also served by an API, at its own
web address, and *that* one **can** take the proper server-side instruction. I've written
that too, but it lives on a separate machine with its own deploy process, so it isn't
live either, and I've not gone and deployed to another host without being asked.

I've sent the change to the review council and haven't seen the verdict yet.

## One small process note, because it says something about how we work

Twice this afternoon I added something to a shared document, and within a minute or two
another session committed that file with my text inside their commit. Nothing was lost
and it doesn't really matter — but the second time, the file looked completely untouched
when I checked it, which naturally reads as "my change didn't save". If I'd trusted that,
I'd have written the whole thing again and we'd have had it twice. The lesson is small
and worth having: on this setup, "the file looks clean" doesn't mean "your edit vanished"
— it might mean someone already committed it. Ask what's in the committed version, not
whether your copy looks modified. I've logged it.

## 2026-08-10, morning — it's live, and it's proven properly

The rebuild went out overnight, so I went and checked the actual page. **It works.** The
published Gauntlet round page now carries the "don't index this" instruction, and Google
and the others will honour it. Shared links are completely unaffected, exactly as intended.

I want to be clear about *how* I checked it, because "I looked at the page and the tag was
there" would not actually have been proof of much. Two things had to be true. First, that
the new code is genuinely running — I checked the actual program on every machine running
it, not the version label, which can lie. Second, and this is the one people skip: I also
re-built a **different** page, one that is *not* flagged, on the same new code. It came out
with **no** tag, and deployed perfectly normally. That's the half that matters. If I'd only
checked the flagged page I couldn't tell the difference between "the switch works" and
"the new code puts that tag on everything". Now I can.

## Two things went wrong on the way, and one of them is worth your attention

**The documented procedure for rebuilding this page didn't work.** The script our own
runbook tells you to use hung for about ten minutes and then failed, having never actually
started the job. I checked first whether the system was just busy — it wasn't, the queue
was completely empty and my job was the only thing in it. There's a second, more direct
route that skips the part that hung, and that worked in about twenty-five seconds. I've
written that down prominently, because the next person will reach for the same broken
script I did.

**The more interesting problem: that failure left no trace in the place we look for
failures.** We have a note on file saying that when this kind of hang happens, the job
table under-reports it and you should count them in the error log instead. Today it was the
exact opposite — the job table said "failed", the error log said nothing at all. So the
under-counting goes *both* ways, which means a clean error log doesn't prove this is
healthy. That matters beyond my little fix, because there's an open bug about these hangs
and part of the evidence for "this particular component is fine" is a clean error-log
count. I've recorded it as a contribution to that bug rather than going off and
re-diagnosing it myself, since someone already owns it.

## What's left

One thing, and it isn't mine to do unasked. The round data is also served by an API on a
**separate machine** with its own deployment process — SSH, rebuild, restart. The code for
that half is written and committed, but the overnight rebuild did nothing for it, because
it isn't part of the same system. It needs someone to deploy it deliberately. It's the
smaller half — the page itself, which is what a person actually finds in a search, is done.
Worth knowing: another thread has since added a separate safety check to that same file, so
whoever deploys it will be shipping both changes at once and should talk to them.

Everything else is finished: the fix is live, the review council's questions are all
answered with measurements, and the notes and warnings are filed where the next person will
trip over them rather than have to go looking.
