# Where we are — imagery that gets made and never used (bug 114)

Plain-prose log for the owner. Append-only, newest at the bottom.

---

## 2026-08-22 — what this is, what we found, and what has changed today

**The complaint, originally.** Back in July you said there was not enough imagery on a
site. There was: twenty-one images had been generated and put on the server, and the site
referenced three of them. That became bug 114, and since then five other threads have
added evidence to it without anyone taking the fix. I have taken the framework half —
the part that would stop it happening again on every site, rather than repairing one more
site by hand.

**First: it is still true, and it is bigger.** Ten sites carried the bad default when
somebody last measured in July; today it is eighteen. Across the fleet there are 94
"content hero" images — one made specifically for one page — and only 23 of them are
actually on the page they were made for. On mortgagecalculator, where ten were generated
in one night on your instruction back on the 15th, two are on their pages and eight are
not.

**Then the thing I did not expect.** One of those sites, fundamentallyai, had been
repaired by hand on 29 July. Someone found the site was pointing at a picture file that
did not exist, worked out what it should point at, and fixed it. Today it is pointing at
the missing file again. Nobody undid it.

It turns out the pipeline was undoing it. Every time the system stores a generated image,
it also writes down "the site's main picture is at this address" — and it was working out
that address from the *kind* of image rather than from the image itself. So a picture made
for one particular page would announce itself as the site's main picture, at an address
nothing had ever been saved to. Generate a hero for the About page, and the whole site's
default quietly re-points to a file that does not exist.

The clearest evidence is almost silly: I listed what every site had recorded as its main
picture, and **every site had recorded exactly the same answer**. Same for icons, same for
logos. Eighteen different sites, one identical address — because the address was never
derived from the site at all. Two of those addresses (`content_hero.jpg`, `icon.jpg`) are
names the system is incapable of ever producing, and I checked: they return "not found" on
every site that lists them.

**And the safeguard already existed.** Whoever built the image handler had written
"don't touch the site's main picture settings" into its configuration. The instruction was
sitting there, correctly, all along — and no code ever read it. The setting was designed,
written down, and never wired up.

**What has changed today.** Two commits, both waiting for the next system build:

1. The store step now honours that instruction, in both directions, and works out the
   address the same way the thing that actually saves the file does. A picture made for one
   page can no longer redefine the whole site. A picture that genuinely *is* the site's
   main one still can — and now records the address that exists rather than the one that
   does not, which is exactly the repair fundamentallyai was given by hand.
2. When a page falls back to the site-wide default, it now says so in the logs, and says
   whether it even had the chance to find its own picture first.

The second one needs explaining, because it is a fix for a different kind of problem. On
mortgagecalculator, two pages out of eight got their own picture and six did not, on the
same night, through the same machinery. I could not work out why, and neither could our
diagnosis system, because the detailed records of that night have since been deleted on
the normal retention schedule. Falling back to the site default is perfectly correct
behaviour for an older site that has no pictures of its own — and it is also exactly what
this bug looks like. The two were indistinguishable after the fact. So rather than guess,
I have made the system say which of the two it is. Next time it happens it is one search
instead of a dead end.

**I was wrong twice today, and both are written down.** I thought the difference was which
of two processes handled each page — our diagnosis loop disproved that. I then thought a
page that already had a stale setting would keep it — which fitted the eight pages in
front of me perfectly, until I looked at every page on the fleet and found ten that had
that stale setting and got the right picture anyway. Both are in the lane's notes and in
the fleet-wide log of wrong calls, along with the cheap check that would have caught each
sooner. The second one is the useful lesson: the refuting evidence was one query away, and
I had already written the wrong reading into my notes before running it.

**What I have deliberately not done.**

- Not repaired the eighteen sites yet. The repair is written but held back until the fix
  above is actually running, because applying it first just invites the next generated
  image to undo it — which is the mistake that cost us the July repair.
- Not run a test page-rebuild on mortgagecalculator. The plan called for it; on reading
  what that actually triggers, it would have rewritten the page's text while another
  thread is actively working that site. Reading the existing data answered the same
  question better, and gave me eight cases instead of one.
- Not started generating the missing case-study images. Somebody would have to widen what
  the generator considers, which is fleet-wide picture-generation spend — **your call, not
  mine.** Same for five stalled jobs whose pages have nothing built to re-render.

**One thing you should know about, which is not mine to fix.** The part of the system that
was supposed to finish this job — link each generated picture to its page — only runs when
a particular daily sweep visits a site. That sweep has not run since 11 August. Its four
sibling sweeps are all current. We have a watchdog that reports this **every single day**,
and the report goes nowhere anybody reads. I have recorded the measurement against the bug
that owns the watchdog rather than fixing it here, but it is the reason a one-off batch of
generated images never gets connected to anything, and it is why the remaining work in
this lane is designed to link the picture at the moment it is made rather than waiting for
a sweep that may never come.

**Next.** Parts two and three: link each image to its page when it is generated, file the
follow-up work immediately rather than waiting for the sweep, and add a check that notices
"this page has its own picture sitting unused". All specified in the lane plan. Part one
is submitted to the review council and is waiting on a verdict.

---

## 2026-08-22, later — the review passed, and the second half is built

**The review council approved it**, on the third round. That is worth a sentence because
the first two rounds both changed the work rather than rubber-stamping it. Round one
asked, reasonably, how I knew that the address the storing step records and the address
the deploying step publishes are actually the same — I had asserted it and not shown it.
They are the same function, one calling the other, but that was an argument, so I made it
a test instead. Round two found something I had genuinely missed: the safeguard is
declared per workflow *step*, and one step takes the image's name from the job it is
handling, so a future job could still declare a single page's picture to be the whole
site's. Every such job today is legitimate, and reaching that step at all requires the job
to have said "this is a brand update" — but nothing in the code prevents it, so it now
logs a warning naming the picture. Anyone who does it in future is findable in one search
rather than invisible.

**And the second half is built.** The step that was meant to connect each generated
picture to its page only ever ran when a daily sweep visited the site — the sweep that
has not run since 11 August. That connection now happens at the moment the picture lands,
in the same breath as the page re-render, so a batch of images finishes its own job
instead of waiting for something that may never arrive.

**One thing I decided not to build.** The plan I wrote this morning said to add a new
setting so the storing step could record which page a picture belongs to. I dropped it
after actually reading the two places that read that information: both of them only ever
look at a different kind of image (the small "card" version used on listing pages). So the
setting would have had nothing reading it — a new knob that does nothing, which this
system has a specific rule against accumulating. The page's own picture was already found
by name; what was missing was only the card, and that is what the change above now
triggers.

**Nothing is live yet.** All of it is program code, so it takes effect at the next system
build. What is owed then, in order: check the running services are actually on the new
build; run the two controls (a page-specific image must *not* move the site default, a
genuine site image must); then apply the held repair to the eighteen sites; then watch the
mortgagecalculator batch — ten pictures, no connections since 15 August — do the whole
journey by itself. That last one is the real test, and it is the honest place to say the
bug is fixed. Not before.

---

**2026-09-02 — picked the thread back up after eleven days, and the news is mostly good.**

The machinery we built in August has been doing its job on its own. The moment a page's
picture finishes, the system now files the follow-up that makes the small "card" version
for listing pages — that has happened **193 times** since we left, without anyone asking,
and every single one of the 193 produced a properly connected card whose file is really
there on the site. The daily sweep that had been dead since 11 August also came back to
life (someone else fixed that), so there is now a belt AND braces. And the poisoning we
stopped — where storing any picture stamped a wrong site-wide default — has not recurred
anywhere. Three of the four things we said had to be true before calling this fixed are
now true and measured.

The fourth was "something should notice this state so it never silently returns", and
that is what I am building today. While checking the ground for it I found two things
worth saying plainly:

First, **most tool pages cannot show a big header picture at all** — about three quarters
of them. On those pages the slot where the picture would go is occupied by the calculator
itself (a known, separately-tracked defect in how tool pages are stored). So some of the
"generated and never shown" pictures were never showable on their page — though they are
still useful, because the listing-page cards are cut from them. The new detector says
which pages are which, instead of lumping them together.

Second, an existing detector has been quietly making the mess worse: it notices "nobody
references this picture" but prescribes the wrong medicine (re-uploading the file, which
changes nothing), and the system's own anti-repeat brake then shelves the item. There are
**1,651** shelved items of that kind today. The new detector names the right medicine per
case; what to do about the old one's backlog is a question I will put to you rather than
decide.

Also housekeeping: the August data-repair migration was submitted for review and the
review never actually ran (lost in dispatch, not rejected) — resubmitting it properly.
And four more dead site-wide defaults (icon, content-hero, illustration, sprite-sheet
paths that point at files which have never existed) get the same careful deletion the
hero one got, now that the thing that kept re-creating them is gone.

---

**2026-09-03 morning — the new build went out, and the missing smoke alarm is now live
and has already rung once, correctly.**

This morning's build carries yesterday's work. I checked the running services really
carry it (asked the services themselves, with a decoy check to prove the question
works), then switched on the two held pieces: the clean-up that deletes the four dead
site-wide image addresses nobody reads (done — 27 removals across 18 sites, every row
backed up first), and the new detector that notices "this page has its own picture and
shows the generic one instead".

Within ten minutes the detector produced its first real report, on idea.uk: six pages
that hold a generated picture no part of the page can display — filed as one tidy
summary card, not six separate nags, with the date on the count and the right owner
named. The sweep that runs it completed normally, so it broke nothing on the way.

What I'd like before we call this bug closed: let the daily sweep visit the rest of
the sites (about a day) and check the reports appear where we already know they
should. If they do, the bug file moves to the closed pile — everything it names is
then either fixed and live, or handed to the specific bug file that owns the
remainder. The delivery mechanism itself (actually putting those orphaned pictures ON
their pages) is deliberately switched off until one careful test run on a quiet site
tells us why the old delivery sometimes missed — that test is written up and waiting,
and your ruling last night on the six plan-less sites gave it a clear runway.

---

**2026-09-03 evening — the detector is working better than I could prove this morning,
and one of the three things I said we were waiting for turns out to be impossible.**

Short version: the new detector has now run on fifteen sites and got every single one
right. I was wrong about what we were still waiting for.

The detector sorts each neglected picture into one of three situations. This morning I
said we needed to see all three happen live before closing the bug. Two have happened.
The third — the case where a page has a picture slot but a calculator is sitting in it —
**cannot happen any more**, because it was fixed the night before the detector switched
on. Your migration on the mortgage calculator site moved the calculators out of the
picture slots. That was the only place on the estate in that situation. So I have been
waiting all day for a report that will never come, and saying so in every update.

I checked this properly rather than assuming. Across the whole estate: 800 pages have a
picture slot, 45 pages have a calculator or game embedded, and **zero** have both in the
same place. I also re-ran the exact count from yesterday's notes: it said sixteen pages
were in that broken state, and today it says none — and the "healthy" number went up by
almost exactly the same amount, which is what you'd expect if they were repaired rather
than deleted.

**The good news is much bigger than that correction.** I found a way to prove the
detector actually ran on each site, rather than assuming silence meant it looked and
found nothing. It ran on sixteen sweeps today, never failed once, and on all fifteen
sites it agreed exactly with an independent count I did by hand. Seven of those fifteen
were sites where the right answer was to say nothing at all — and it correctly said
nothing. That matters more than the reports it did file: anything can raise an alarm,
but a thing that knows when to stay quiet is actually working.

**One practical correction.** I had told you the sweep would visit every site in about a
day. It won't — the schedule turns out to be one site every three hours, and only sites
it hasn't touched in a week. So full coverage is more like a few days. Nothing is broken;
my arithmetic was wrong.

**Where that leaves the bug.** I'm not closing it tonight, but only one thing is left,
and it's automatic: four sites that should produce a report haven't been visited yet, the
biggest being webdesign.co.uk. I've written down exactly what each of them should say
when its turn comes — so whoever picks this up can check it in one query instead of
working it out again. If they match, the file moves to the closed pile.

**Something new landed for us today that isn't part of this bug.** The mortgage
calculator lane has handed us the whole picture job for that site — eighteen tool pages
whose images are all generated and paid for and simply not shown. Their note says the
change that used to be unsafe is now safe, thanks to that same migration. That is real
work with a spend decision attached, and I've not started it; flagging it so it doesn't
get lost between the two lanes.

---

**2026-09-04 — the bug turned out to be a different, bigger thing, and the fix I was expected
to ship would not have worked.**

I picked this up expecting to finish it. The plan I inherited said one thing was left: switch
on a mechanism that was already built and waiting. I have not switched it on, and I think that
is the right call. Here is why, in order.

First, the housekeeping. Three other threads were working on pictures today, so before touching
anything I asked them who owned what. The one called "imagery" said it was a different job and
listed exactly which files it was in, so I knew we would not collide. The finetuning thread —
the one that brought you the missing hero image on the case studies page — handed this over
properly: *"Take it. My CONTRIB was a handover, not a claim."*

**The important thing they told me is about your own words.** You said *"let's use the hero
images somehow, we don't need a stop gap though."* You were answering a straight choice between
patching four pages by hand and switching on the built mechanism. You said no to the patch and
asked for the result. **You have never been shown the mechanism itself, or the review that says
it is flawed.** So I have treated your instruction as "get the pictures onto the pages" and left
myself free to pick how — which turned out to matter, because the built mechanism has a problem.

**What I found instead.** Every picture slot on every page carries a written instruction from
the component that owns it, and it always says the same thing: *find the right image, and if you
cannot find it, use this specific stand-in.* There is no third option. What I found is **123
picture slots, across 32 of your sites, where neither happened** — no image and no stand-in, just
nothing. The instruction was ignored. Of those, 86 are showing the site's one generic header
photo instead, which means **every "about us" and "contact" page across thirty-odd sites is
currently showing the same picture**. The other 37 show nothing at all.

**Why it keeps coming back, which is the part I care most about.** There is a step in the page
build that clears a page's contents and writes them again from scratch. If the incoming batch
does not happen to mention the picture, the picture is simply dropped — and the finished HTML
still has it, because that was assembled a moment earlier. So the page looks right and the
record behind it is wrong, and the next rebuild makes the page match the wrong record. This is
exactly why the repair we did on 26 August faded: it fixed nine pages, and eight days later only
three were still fixed. Another thread measured that decay before I arrived; I can now say what
caused it. **It means any hand-repair has a shelf life, and so does the mechanism I was going to
switch on** — which is precisely what its reviewers objected to, and they were right.

**What I have proposed instead**, and it is with the review council now. Rather than adding
another thing that writes pictures onto pages, make the page-writing step obey the instruction
that is already written down: if the incoming batch has no picture, keep the one the page already
had, and if there was none, use the stand-in the component asks for. The neat part is that this
needs no cleverness at all — the stand-in is written in plain text in the component's own
settings, so honouring it is a matter of reading, not working anything out. It ships switched
off, and switching it off again is a one-line change with no rebuild.

**Two things I want to be straight with you about.**

One: this stops the rot and guarantees *a* picture. It does **not** guarantee each page gets
*its own* picture. Only 18 of the 123 actually have their own image sitting ready; for the other
105, the generic one genuinely is the right answer according to their own settings. Getting the
right picture onto those 18 is a separate, smaller job.

Two: my fix does not cover everything. The admin screens, two command-line tools, and about 25
of our database scripts also write to this same place, and nothing I can write in the application
will constrain a database script. Since our hand-repairs *are* database scripts, the very thing
that caused the decay sits outside my fix's reach. The only place that would catch everything is
the database table itself, and that is a big enough change that it belongs in a proper written
proposal rather than smuggled in with a bug fix. I have said so in the submission rather than
letting a good-looking result imply more than it does.

**And a confession, because it belongs here.** I got four measurements wrong this afternoon
before I got one right — including publishing a count to three other threads before I had
checked it properly (it was 123, I told them 121). Every one was the same mistake in a different
coat: a test that could not have told me I was wrong. The one that finally worked, I checked
against a case I already knew the answer to before I trusted it. All four are written up. I
mention it because the finding underneath is now load-bearing for a fleet-wide change, and you
should know it survived four bad attempts rather than arriving clean.
