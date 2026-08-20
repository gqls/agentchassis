# Where we are — bug 252, the share-preview and page-language half

Plain prose, append-only, newest at the bottom. The owner's running log.

---

## 2026-08-20, morning — picking it up, and finding the bug had changed shape

You asked me to look at bug 252 — the one about pages losing their own "share preview" details and
every page claiming to be written in American English. Two things are worth telling you before any
of the fix detail.

**First: the bug is still real, but it is not the bug we filed.** When it was written on 11 August,
the complaint was that an assembled page carried *nothing* of its own for social sharing — share a
page on LinkedIn or WhatsApp and you'd get the site's generic details rather than that page's. Since
then a different piece of work (the share-card and favicon job) added a block of tags to each site's
shared page-header. That block includes a line saying "the address of this page is…" — and it fills
it in with the site's **home page**, because at the moment it runs there is no page to ask.

So today, every one of our assembled pages doesn't merely lack its own identity; it actively states
the wrong one. I checked a real page rather than trusting the database: `about.html` on
ai-agent-orchestration.com tells a share scraper it is the homepage, while the *canonical* tag right
below it — a separate mechanism we fixed last week — correctly says it is the about page. The page
contradicts itself. Four of our sites are worse still: they carry the tag twice, once blank and once
filled with the site name, because the new block cannot see that the site's page-header template had
already left an empty one there.

That matters practically, not just tidily: a missing tag is silence, and a scraper falls back to
guessing sensibly. A **wrong** tag is followed. It is also the reason the original fix idea in the
bug file no longer works — it proposed leaving blanks for us to fill in, and the sites that have
blanks already have them overridden a few lines later. So the shape of the fix changed: instead of
filling blanks, assembly will now remove whatever the shared header claimed about page identity and
state the page's own. That happens to repair all 22 affected sites and both duplicated tags in one
move, without needing to rebuild any site's header first.

Scale, re-measured rather than repeated from the file: **700 assembled pages across 26 sites**. The
file said 503 across 23; that was true in August and the estate has grown.

**Second: the language half is unblocked and I need one small confirmation from you.** You already
decided in August how this should work — the language belongs in the site's page-header, not in a new
database column, and our Go code should stop hard-coding "English". That is what I am building. Today
you also told me to switch all the UK sites over to British English straight away rather than shipping
the ability and leaving it switched off, which I agree with — an unused mechanism quietly rots.

The confirmation I need is about the `.com` sites. The bug file's own argument is that our `.com`
sites are British too, which is why guessing the language from the web address was rejected. I will
write the change so it names each site explicitly and you can see the list, and I'll put the list in
front of you before it is applied — because the decision "is this site British?" is yours, not a
pattern match. The internal pool sites are excluded; they serve nobody.

**On coordination**, since several sessions work this tree at once: nobody else is editing this. You
mentioned another thread might be working on share-preview copy, so I swept for it specifically —
commits, uncommitted work, every lane's notes, and the live work queue. What I found is adjacent but
different: one lane has been fixing *page descriptions* (the sentence under a search result), which
is now scheduled and running, and the share-card lane closed its own bug yesterday and spun its
leftovers into a new file, `bugs_open/322`. That new file overlaps mine by design — it describes the
same block of tags from the other end. I am taking the part that lives at page assembly and leaving
its other four items alone, and I've written that division into both files so neither of us
duplicates the other.

One planning assumption of mine also expired within the hour, which is worth knowing because it is
the normal condition here rather than an incident: I had built the plan around the code not compiling,
because another session had a half-finished change in one of the files I need. By the time I started
work they had committed, the code compiles, and both files are clean. I kept the cautious version of
one step anyway — it costs nothing — but the lesson is that any statement in a plan about another
session's work needs re-checking at the moment you act on it, not when you write it.

**What happens next**, in order: I write the code and its tests; it goes through the review council
(which as of yesterday also reviews database migrations, so the whole change goes in one submission);
then it needs a fleet release before anything is switched on. The ordering there is load-bearing and
easy to get wrong — the database half takes effect the instant it is applied, while the Go half does
nothing until the new image is running. If they land in the wrong order the system would rebuild
every site's header using the *old* code, mark it all as freshly done, and go quiet with the fleet
still wrong. So the migrations are named to be held back, and I will prove the new code is actually
running in the pods before applying them. After that: two canary pages checked by eye, then the
existing staleness machinery carries it across all 26 sites in waves on its own.

I have also asked the diagnosis loop to check my reasoning independently while I write the code. I
don't expect it to overturn anything — the mechanism is plain in the source — but the claim here is
cross-cutting, which is exactly the case where our own rule says spend the run rather than assert.
