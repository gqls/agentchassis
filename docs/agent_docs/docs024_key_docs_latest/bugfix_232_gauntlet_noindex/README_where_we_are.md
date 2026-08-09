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
