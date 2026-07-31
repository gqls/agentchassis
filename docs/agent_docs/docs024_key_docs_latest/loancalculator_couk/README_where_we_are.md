# Where we are — loancalculator.co.uk

The owner's running log. Plain prose, append-only, newest at the bottom.

---

**2026-07-30, afternoon.** Started by just looking at the two folders — the source
one in `domains/` and the deployed one in `sites/`. They're the same site: the first
builds it, the second is the copy that gets published. Nothing has touched either
since the 20th of March.

Then I checked whether the site was actually up, and it wasn't. Worth explaining the
symptom because it's a misleading one: the domain resolved fine, the padlock was
valid, the request went out — and then nothing came back at all. It just hung until
it gave up. No error page, no 404, no 500. That looks like a network fault, but it
wasn't.

Every working site of ours has a tiny programme running at Cloudflare's edge that
fetches pages out of our storage bucket. I checked it directly (there's a
`/worker-health` address that answers "Worker is running!" if it's there) —
gamesdesign.co.uk answered it, loancalculator.co.uk didn't. So the edge programme was
never connected to this domain, and requests were still being sent to the old Amazon
storage location from the original March setup, which is long gone. That's why it
hung: the proxy was alive, the thing behind it wasn't.

You fixed that this afternoon by adding the missing route, and the site is back —
confirmed at 16:11, all the pages I tried return properly.

**Then the more interesting bit.** You asked for this to go in through the framework
so it's managed properly, and then for anything the framework's own adoption process
would have got wrong to be reported or fixed. So I read that process carefully, and it
would have done three things we very much don't want.

It would have changed every web address. All our addresses look like
`/tools/standard-calc.html`; the adoption code rewrites them into
`/tools/standard-calc/index.html`. It doesn't matter what the real address was — it
throws it away and builds a new one. That's all 28 addresses broken.

It would have rewritten the calculators. Every page gets marked "recreate", and pages
with interactive bits get handed to a language model to rebuild. Twelve calculators
that work, that do correct arithmetic, regenerated from scratch. Even if it went
perfectly it's a pointless risk; realistically some of them come back subtly wrong.

And the setting that's supposed to prevent exactly this doesn't do anything. There's a
"fidelity" option — how faithfully to keep what you found — and you can pass it, and
it gets written into the record, and then nothing on earth reads it. I checked the
whole codebase rather than trusting the comment that says so: ten mentions of the
word, every one of them about something unrelated. So the promise is in the interface
and the implementation was never written.

**What I want to do about it.** Not work around it for this site. If I hand-port this
one, the site is fine and the framework is exactly as unable to adopt the next one.
So the plan is to build the missing faithful-preservation mode properly — make
"fidelity locked" mean "keep the addresses, keep the bytes, don't regenerate
anything" — and then use it on this site as the proof. It's two focused changes, both
opt-in, so no existing site can be affected either way. They'll go through the review
council before I call them done.

**One thing I found that you may not know about.** The site has real problems of its
own, and they've been live since March. Four pages aren't really pages — they're
loose fragments with no page structure, so they arrive as unstyled text with no
navigation. Ten of the 28 pages have no menu at all, and that includes the main loan
calculator and the overpayment tool, so anyone arriving on those from Google has no
way to get to anything else. The menu that does exist has a link to a calculator page
that doesn't exist, and it's on every page that shows the menu. And the main
calculator has been loading its stylesheet from the wrong place, so it's been
unstyled this whole time.

I'm going to fix all of that first, before adopting, for a slightly non-obvious
reason: the framework learns the site by reading the live pages. If I adopt first, it
faithfully preserves the broken version and we've frozen the bugs in.

**Order of work from here:** fix the site where it sits and publish it; build the two
framework changes and get them reviewed; run the real adoption process with the new
mode; check every page and every calculator still works, including that a later
rebuild can't quietly change them back.

**Something for you to decide, but not now.** The written content quotes interest
rates and a "last updated" date from March. I'm leaving those alone — changing the
copy isn't part of adopting the site, and I'd rather not mix the two. But once it's
managed, keeping figures like that current is exactly what the platform could be
doing for us, and that's worth a conversation after this lands.

**2026-07-30, later.** Two things done, one waiting.

**The site is fixed and back up.** All the problems I listed above are repaired and
live: the four broken pages now render properly, every page has its menu, the dead
menu link is gone (with a forwarding page left at the old address so anything
pointing there still works), the main calculator is styled again, and the address
list search engines read is correct for the first time. I checked all 34 addresses
and every one returns a page. I also confirmed the three dead files I deleted now
genuinely return "not found" — which matters more than it sounds, because it proves
the publish actually reached the storage bucket rather than me reading a stale copy.

**The framework change is written and under review.** It does what I described:
asking for "locked" fidelity now means the platform keeps a site exactly as it
found it — same addresses, same pages, byte for byte — instead of rewriting it.
Two focused changes, both switched off unless explicitly asked for, so no existing
site can be affected. I wrote tests, and then deliberately broke each safety check
to confirm the tests actually catch the failure rather than just passing quietly.
One of those attempts taught me something: my first break didn't change behaviour
at all, so the test passing proved nothing — I had to break it properly before the
check was worth anything.

It has gone to the review council and I'm waiting on the verdict, which usually
takes about half an hour. I'm deliberately not deploying anything until it comes
back, because deploying restarts the system and would kill the review mid-flight.

**One preparation worth mentioning.** The crawler had a cap of 30 pages and this
site has 29 files — and the homepage can count as two addresses, so it could
genuinely have hit the cap and quietly left pages behind. A page lost that way
doesn't error; it just never arrives. I raised the cap to 60 first, after taking a
backup of the old setting and checking the backup really held the old value.

**A mistake of mine, for the record.** I committed the framework change before
submitting it for review, which meant the reference number linking the two was not
available yet and I wrote a placeholder. The system that reports which changes have
been reviewed will therefore not credit this one automatically. Nothing is broken and
the review is happening normally, but the tidy trail is missing. The right order is
submit first, then commit — I've written that into the runbook.

**2026-07-30, late evening.** The site is adopted and healthy, the new mode works,
and one thing needs your decision.

**What went right.** The adoption ran through the real framework and did exactly what
it was supposed to. All 27 pages are now recorded in the platform with their web
addresses **unchanged** — `/tools/standard-calc.html` is still
`/tools/standard-calc.html`, not the rewritten form the old path would have forced.
Every page is marked as ours-not-the-pipeline's, and **not one of the twelve
calculators was handed to a language model to rebuild**. Under the old behaviour all
28 addresses would have changed and all twelve calculators would have been rewritten.

I also proved the safety property rather than asserting it: I deliberately triggered a
rebuild of one page afterwards and confirmed it republished **byte-for-byte identical**
content. So a future maintenance sweep cannot quietly restyle this site.

**What went wrong, and what caught it.** Before letting the deploy run I compared what
the platform had captured against what the site actually serves. **None of the 27
matched** — every stored page was about 9KB bigger, and the same ~9KB on all of them.
That consistency was the clue. The crawler does not return the page the server sends;
it returns the page *after the browser has run the JavaScript*. So the navigation menu
that our site builds in the browser had already been baked into the file, every
relative link had been rewritten to a full address, and the dropdown menu links had
been turned into links back to the same page — which would make clicking "Tools"
reload the page instead of opening the menu.

Three pages slipped out to the live site before I stopped the rest. I restored them
from our own copy and confirmed the live site is correct again; all 27 addresses
return normally. The remaining 24 were cancelled before they could deploy.

Two things are worth saying plainly about that. First, **the check that caught this
was a manual one I ran by hand** — the review council had already told me that check
was "aspirational, not built", and they were right; running it by hand is the only
reason three pages were damaged instead of twenty-seven. It should be part of the
pipeline. Second, restoring the files exposed a separate trap: the publishing step
**silently skipped one of the three files** because the bad copy in storage was newer
than my fix, and the deploy still reported success. I only found it by checking the
storage directly rather than trusting the green tick.

**Where it stands now.** I loaded the correct served bytes into the platform for all
27 pages and verified every one matches exactly, so the record is now both correct and
safe to redeploy. The site is live and unchanged from a visitor's point of view.

**Your decision.** The new "keep it exactly as it is" mode is sound, but its source of
truth — the crawler — cannot deliver exact bytes. Three ways to fix that properly:

1. **Adopt from our own files** rather than by crawling. We already hold the exact
   bytes we publish, so this is the obvious source. I had listed this as an optional
   extra; it now looks like the right answer.
2. **Fetch the pages plainly**, without a browser, just for this mode.
3. **Build the byte-comparison check into the pipeline** so it refuses to deploy when
   what was captured differs from what is served.

My recommendation is 3 regardless of which source you choose, because it is the check
that saved this run, and then 1 as the source. I have not started either — this is a
genuine choice about how the framework should work and it is yours to make.

**Also worth knowing:** the review council looked at the framework change and asked
for revisions. Several were fair and I have already made them — the code now refuses
to continue rather than quietly dropping a page, and it records in the database when
it has skipped the design stage so that is never invisible. Two objections are
genuinely for you rather than me: whether skipping the design cascade for preserved
sites is acceptable (our own documentation says that stage always runs), and whether
this needs a formal architecture review. I have not overridden either.

---

**2026-07-31 — a "before" photograph, and one thing we all had wrong**

The site is fine. All 27 pages still load, the stored copy still matches what visitors
see exactly, and the framework changes from yesterday are genuinely running on the
live machines (I checked by looking inside the running program for the new code, not
by trusting the version number).

Before changing anything I took a proper "before" photograph of the calculators, in a
real browser, driving each one the way a visitor would. I needed this because the test
you set is that the site must start off *"similarly enough with working tools"* — and
the word doing the work there is *still*. Without a before, "still works" is not
something anyone can check. **Eleven calculators work.**

Which brings me to the thing we had wrong. Every note in this project, including my
own, has said this site has twelve calculators. **It has eleven.** The page called
"credit roadmap" is not a calculator at all — it is a page of writing that happens to
live in the tools folder. There is nothing on it to click or type into. I found this
two separate ways and they agree, so I am confident.

It matters more than a miscount looks. If we test "every calculator still works" over
twelve, that test can never pass, because one of the twelve cannot compute and never
could — and a test that always fails gets ignored, which is worse than not having one.
Over eleven it is a real gate. The roadmap page simply becomes an ordinary page the
system is free to improve, which is what it always should have been.

**The next step is bigger than I told you, for a specific and slightly embarrassing
reason.** I said the obstacle to making the site fully editable was that the framework
would wrap a whole page inside another whole page. That is true but it was only half
of it. The other half is that when we adopted the site in "keep it exactly as it is"
mode, we never stored the *furniture* — the top navigation, the page header, the
footer. Nothing needed it, because that mode ships the whole page untouched. So if I
simply flipped the switch to "editable" today, every page would come out not just
malformed but stripped: no navigation, no header, nothing. I found this by reading how
the assembly step actually fetches the furniture before I wrote anything, rather than
by breaking the site and then wondering why.

So the job includes rebuilding the site's shared furniture as proper framework pieces,
not just chopping each page into parts. That is more work, and it is the right work —
it is exactly what makes this site the same kind of thing as the others.

**Two things will visibly change, and I want you to hear them from me first.** Once
pages are assembled rather than copied, every page gains a footer (there is currently
no footer anywhere on the site) and gains a proper mobile setting. That second one is
a genuine repair: four pages — the legal page and three guides — are missing the
instruction that tells a phone to render at phone width, so today they display at
desktop width on a phone. Going editable fixes that on every page at once. The prose
may also get rewritten over time, which you have already said is fine.

**A fork in the road, and it is yours.** A neighbouring project started this morning
is building a combined loan-and-mortgage site, and it copies this site's files. It is
being careful and has explicitly kept out of this site's own adoption, so nothing is
treading on anything. But for the same problem it chose a **different answer**: freeze
the calculator pages permanently and let only the guides evolve. That is much less
work than what I am doing and it is a perfectly reasonable trade.

The reason I have not switched to it is that it would leave eleven of this site's
twenty-seven pages frozen for ever, and your instruction for this site was that it be
*completely* editable. So I am proceeding with the fuller version — preserving each
calculator's working parts exactly while letting everything around them be rewritten.
If you would rather have the cheap version here too, now is the moment to say, because
it is a much shorter road and I would stop building the longer one.

---

**2026-07-31 (later) — I built the splitter, and my own tests told me it worked when it didn't**

Progress, one genuine near-miss, and one thing I need to flag about the homepage.

**The near-miss is the interesting part, so I'll start there.** I wrote the tool that
splits each page into "editable writing" and "the calculator, kept exactly as it is",
and I wrote four tests for it: the calculator's code survives untouched, its styling
survives untouched, no chunk of the page goes missing, and nothing the calculator
depends on gets left behind. All four passed, on all 27 pages, first time.

The splitter was useless. It had put **98% of the text on the calculator pages into the
frozen, untouchable part** — on eleven of the twelve pages, *everything* was frozen.
The site would have ended up exactly as unable to change as it is today, and I would
have had a clean set of test results saying otherwise.

The reason is worth knowing, because it is not a coding slip. Every test I had written
asked *"did anything break?"*. Not one asked *"did anything actually become
editable?"* — which is the entire point of the job. A test that only checks you did no
harm will happily pass a change that does nothing at all. I have added a fifth test
that measures how much of each page ended up editable, and that one failed straight
away. Two rounds of fixing later, the frozen share is down to 25%, and the flagship
loan calculator page went from 90% frozen to 5% — its whole "Basics of Borrowing"
article is now ordinary editable text, with only the calculator itself preserved.

I have written this up for the wider team, because the general version applies to any
job like this: tests that check you preserved something can never fail on a change that
does nothing. So at least one test has to be one that *doing nothing* would fail.

**The homepage has a calculator, and I had missed it.** This morning I told you the site
has eleven calculators, not twelve. That was right about the tools folder and wrong
about the site: the **front page** has a working loan calculator on it too. So it is
twelve — eleven in the tools folder plus the homepage — and the roadmap page still is
not one. The old figure of twelve was the right number by coincidence while pointing at
the wrong page, which is more awkward than simply being wrong.

It mattered practically: my "before" photograph only covered the tools folder, so if the
splitting work broke the homepage calculator, nothing would have told me. Re-taken and
extended — thirteen pages checked, twelve calculators working, one page correctly
recorded as having nothing to click.

**Something you should know about the calculators' appearance.** Eight pages carry their
own private styling rules written into the page itself, and seven of those are
calculators. On the credit health check, two of those rules are what make it show one
step at a time instead of all five at once. So "keep the calculator working" is not just
about keeping its code — drop those rules and the calculator computes every number
correctly while looking completely broken. That is now one of the things the splitter
proves it has preserved.

Relatedly, a neighbouring project reported that this exact page was broken on our live
site. **It isn't** — I checked what the server actually sends, and the rules are there.
Their check only looked at the shared stylesheet and could not see rules written into the
page. Their wider claim that "36 styling classes are undefined" is really 19 for the same
reason. I have written to them with the evidence and a corrected command; their own site
is unaffected. Nineteen is still real, though, and one of them is the FCA warning box —
regulatory text that currently renders unstyled.

**One thing has become a genuine blocker for you rather than me.** The plan said we would
read the site's files from our own deploy repository, platform-side, so any future site
could be adopted the same way. I costed it before building it: the machinery exists and
works, but **the platform's read-only GitHub token cannot see the `sites` repository at
all** — only the main code repository. Widening it needs GitHub admin access, which is
not on this machine. Same shape as the DNS item on the neighbouring project: yours to do,
not mine.

The good news is that it turned out not to block anything here. Decomposition needs
faithful copies of the pages, and we already have those in the database — I re-verified
all 27 are byte-for-byte identical to what visitors receive. So the file-reading work is
no longer a prerequisite; it is an improvement whose value is *the next* site, and it can
wait for you.

**Where this leaves things.** The splitting rule is now proved against all 27 real pages,
offline, with nothing written to the database and no site file touched. What remains
before anything changes for visitors is the shared page furniture I mentioned earlier
(the navigation, header and footer, which the "keep it exactly as it is" mode never
stored), then the switch to editable mode, then letting the normal planning and design
stages run. I have not started any of that, and I would rather not until you have had a
chance to answer the question from my last note about whether you want the full version
here or the cheaper freeze-the-calculators version the neighbouring project chose.

---

**2026-07-31 (evening) — answering your question, and finding that our tool tests prove almost nothing**

You asked why the articles are editable. Three answers, and the third is the real one.

**Nothing on the site is editable today.** Every one of the 27 pages is still in the
"keep it exactly as it is" state. The 75%-editable figure I gave you describes what the
splitting tool *would* produce if we ran it. It is not a description of the live site,
and my earlier note could have been read that way.

**Editability isn't a property of the page — it's a property of how the page is
recorded.** A chunk of a page becomes editable when it is stored as its own piece,
belonging to a page the framework is allowed to rebuild, unlocked, and listed in the
plan so the writing and design stages can see it. Nothing about the HTML makes it
editable or not.

**And the honest answer: the articles are editable because they were what was left
over.** The only instruction I had was that the site must start with working tools, and
freezing the calculators byte-for-byte was the only way I could *guarantee* that. So I
froze everything the calculator's code touches and set the rest free. The articles
weren't chosen — they're the residue. That is a poor reason for a permanent boundary,
which I think is what you were getting at.

And the freeze has a cost I should have put in front of you: the frozen quarter is
exactly the part that can never improve. It can't be restyled, can't be made to work
properly on a phone, can't be fixed when its styling is missing — and it isn't
registered as a tool the framework knows about, so no other page or site can ever reuse
those calculators, and the tool-improvement machinery cannot see them. **So yes, you are
right, and rewriting them is necessary rather than optional.** There are already 36
proper tool components in the framework; these twelve should be joining them.

**But before rewriting anything, I went looking for what would catch a mistake, and the
answer was nothing.** This is the part I want you to know about.

Rewriting a loan calculator means re-implementing arithmetic about people's money. We
have two things that look like tool tests. One checks that the expected elements exist on
the page. The other reports that a tool "responds" when you interact with it. Neither
checks a single number.

Worse, "responds" turns out to be close to meaningless. I built a page with one number
box on it — no code at all, nothing that could possibly calculate anything — and our test
scored it **RESPONDS**. The reason is that the test compares a snapshot of the page
before and after, and the snapshot includes the box you just typed into. So typing
always counts as the page responding. **Which means my own headline this morning — "12
calculators working" — claimed more than I had measured, and I've corrected it.** What
was actually established is that 12 pages are healthy and reactive; not that any of them
computes correctly.

**So I built the missing test.** It drives every calculator in a real browser with three
sets of inputs, derived from each field's own starting values, and records every number
the page produces. Any rewrite then has to reproduce those numbers exactly. I've
captured this for all twelve, and checked it two ways:

- **It's repeatable** — run it again against the untouched site and all twelve match
  exactly. Without that it would be useless.
- **It actually catches things** — which repeatability alone can't show. I took a copy of
  the main loan calculator and made one small error in the interest maths. The page still
  loads, still reacts, still shows believable money. The old test calls it fine. The new
  one catches it in every case and prints the difference: £202.29 a month becomes
  £205.74. That's £207 over the life of the loan, and until today nothing we had would
  have noticed it.

I found and fixed three faults in my own test while building it, and one is worth
repeating because it was the same trap in a new coat: my first version read the page
slightly too early, before the calculator's code had loaded, so it recorded **£0.00 as
the correct answer for everything** — and reported success. A recorded set of answers
like that would have certified a completely broken rewrite as perfect. It now refuses to
save anything unless it can prove each calculator both reacts and produces different
answers for different inputs.

I've also written to the team building the combined loan-and-mortgage site, because
they've just ported 24 calculators and their sign-off test is the "responds" one. Their
own check that the calculator code is copied character-for-character is genuinely
stronger than anything I had, and I told them so — but it can't catch a renamed input
field or a changed starting value, and both would sail through every test they have.
They can use the new one as-is.

**Where that leaves the rewrite.** The safety net now exists, so the work can proceed:
each calculator becomes a proper framework tool, reusable and improvable, with the
requirement that it produces identical numbers to today's version. I have not started
changing anything yet. The open question from my last note still stands and now matters
more, because the rewrite is the bigger of the two roads — say the word and I'll begin.
