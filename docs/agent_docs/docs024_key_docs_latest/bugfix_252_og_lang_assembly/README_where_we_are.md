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

---

## 2026-08-20, afternoon — the fix is written, and one check stopped me shipping a fresh mistake

The code is done, tested and committed; it is with the review council now, and nothing is switched on
yet. Three things you should know, one of which is a decision I need you to confirm.

**The thing I most want to tell you: your instruction to show you the list of sites caught a real
error before it shipped.** You said to opt all the UK sites into British English. My plan was to add
the `.com` sites too, on the bug file's own argument that our `.com` sites are British. Before writing
that list I checked each site rather than assuming — and **relojistas.com is a Spanish-language
publication**. Its recorded location is España, its tagline and every heading on the live page are in
Spanish. Declaring it British English would have been false metadata stated more confidently than the
plain "English" it says today — which is precisely the fault this whole bug is about. It now gets
Spanish (`es-ES`) and everything else gets British English. Twenty-four sites British, one Spanish,
the internal placeholder sites left alone entirely.

I'd like you to confirm that Spanish call, since it goes slightly beyond what you asked for — you said
"UK sites", and this is me correcting a non-UK one while I'm in there. My reasoning is that leaving it
saying "English" is a known-false value I'd be walking past. Easy to reverse either way.

**Second: what the fix actually does.** When a page is assembled it now removes whatever the site's
shared page-header claims about *that page's* identity and states the page's own — its title, its
description, and its address. The address is worked out by the same piece of code that produces the
"canonical" tag, so the two can't drift apart; before, they were two separate calculations kept in
step by a comment. Pages with no description written yet simply get no description tag rather than an
empty one, which is deliberate: staying silent is better than a page describing itself as nothing.
That's a good half of pages today, and the other lane's description-writing work is what shrinks it.

A nice side effect: this repairs the duplicated tags and the wrong addresses **at page-assembly time**,
so it doesn't need every site's header rebuilt first. That mattered more than it sounds — a code
change doesn't cause headers to rebuild, so a fix that depended on rebuilding them would have sat
inert.

**Third: I got something wrong in my own design and the test caught it.** I'd decided the new step
must run *before* the existing description step, wrote the reason into the code as settled fact, and
built a test to pin it. Then I deliberately swapped the order to check the test would fail — and it
passed, which means it was pinning nothing. Investigating that showed the order does matter and I had
it **backwards**: my order caused the page description to be written into the *image* tag, a tag this
change isn't supposed to touch. Swapped, corrected the comment, and the test now genuinely fails if
anyone swaps it back. It's written up in the wrong-calls log with the cheap check that would have
found it sooner. Nothing shipped wrong — but I'd rather you know the fix had a real fault in it that
was caught by testing the test, not by me being careful.

**On the independent diagnosis run** I mentioned: it came back "not confirmed", and the reason is
useful rather than worrying. It re-derived the mechanism from the source code exactly as I had, then
couldn't find any actual page to check it against — because it looks for pages' headers in three
database columns that have been empty for the entire fleet for months, and the one place the evidence
does live gets cut off before the relevant part. In other words it can read code and configuration
but it cannot see what a site actually serves. That's a genuine blind spot in one of our own tools, so
I've written it down as a trap for the next person: a "not confirmed" from that tool about anything a
visitor can see means "I couldn't look", not "you're probably wrong". The evidence for this bug never
depended on it — I'd already fetched the live pages.

**What's left**, in order: the council verdict (I'll act on it, including if it wants changes); then a
fleet release before any of this does anything; then I prove the new code is genuinely running in the
pods; then two canary pages checked by eye; then the two database changes; then it spreads across all
26 sites through machinery we already have. The database changes are deliberately held back until the
release has happened — if they land first, the system would rebuild everything with the *old* code,
mark it all as done, and go quiet with every site still wrong.

---

## 2026-08-20, late afternoon — approved by the council, and the new build does not contain it

**The review council approved it, first time round**, with five advisory comments and no serious
objections. Twelve reviewers looked at it; six had nothing to say. That's a good result, but the more
useful part is that four of them asked me to *check* something rather than change anything, and all
four checks were worth running:

- One asked me to prove, rather than assert, that only one thing reads the shared page-header before I
  delete a line from it across the whole fleet. It holds — but I had been asserting it, and that was a
  fair hit.
- One asked what happens to the history table when those bytes change. There's a trigger that archives
  the previous version; it fires once per site and *keeps* the old copy, so it's the safety net doing
  its job rather than a cost.
- One asked whether removing two lines from a template would leave a setting pointing at nothing. It
  doesn't — **and the reason turned out to explain the original bug.** Those two lines referred to
  settings that were never actually defined for that template, so they could only ever produce an
  empty value. That's precisely why four sites were serving a blank tag. I'd worked out that cause by
  reading the code; now it's proven.
- One pointed out I'd said "we'll check the new code is really running" without saying *how*, and that
  a version number proves nothing. Quite right, and it mattered immediately — see below.

**Now the important bit: the chassis build deployed today does not contain this fix.** It was built at
11:18 this morning; the code was committed at 15:03. So it was cut about four hours before the fix
existed.

I didn't want to conclude that from timestamps alone, so I asked the running program directly whether
it contains the new function — along with two controls: something that must be there, and something
that can't be. The control that must be there was found, the impossible one wasn't, and **the new
function is absent.** The check is trustworthy and the answer is no.

**Practically, that means nothing is switched on and nothing should be.** The two database changes stay
held back, and there's no point canarying a page yet — it would just re-render with the old code and
look like the fix doesn't work. **What I need is a chassis build cut from the current code.** I haven't
done that myself: builds and releases are whole-fleet and yours to run.

Once that's done the rest is quick — prove the new code is running the same way I just did, check two
pages by eye, apply the two database changes, and let the existing machinery spread it across the
sites in waves.

**One aside worth your time:** the command our own documentation recommends for checking what a
service is running gave me 2.4 megabytes of a different piece of work's notes. The service logs whole
review payloads, and those payloads happen to quote the very phrase the command searches for. It's a
known trap, but it fired on the exact recommended command, so I've written it down as something to fix
in the instructions rather than just a thing to know.

**And, as you asked, I've written up everything I caught along the way** that isn't this bug — six
things, in `docs/agent_docs/docs024_key_docs_latest/bugfix_252_og_lang_assembly/FINDINGS_2026-08-20_errors_caught.md`.
The two I'd act on first:

1. **webdesign.co.uk — our largest site at 117 pages — serves no page-header element at all.** Its
   header component is a fragment with the opening and closing tags missing. Browsers cope, so it
   looks fine, but every tool we have that adds something to a page header looks for that closing tag
   — and they don't all behave the same way when it's absent. One of them silently does nothing, which
   is why that site quietly gets no share-preview tags whatsoever. It will keep excluding itself from
   every future improvement, invisibly.
2. **Migration numbering has no allocator and collisions are now routine.** Two numbers were already
   used twice before I started; while I was writing my two files, five consecutive numbers were taken
   by three other sessions, including both of mine. I renumbered. Today it's confusion; the day two
   files with the same number touch the same data it's a real conflict.
