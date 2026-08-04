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

---

**2026-07-31, later — all eleven calculators rewritten, and the check is now the
platform's rather than mine**

You asked for three things: make the numeric check part of the tools workflow,
rewrite the rest of the tools, and write a summary. All three are done. The
summary is `SUMMARY_2026-07-31_calculators_that_prove_their_own_arithmetic.md`.

On the check: it is now a real part of the platform, not a script I run. It drives
a calculator with fixed inputs and insists on the exact answer — £303.44, not
"something shaped like money" — and it runs on the normal schedule in a real
browser. I deliberately did not give it a cheap version that only checks the boxes
exist, because a weaker check wearing this name would recreate the exact false
confidence we had just spent a day proving was there.

On the tools: eleven, not twelve. The homepage calculator and the standard loan
calculator turn out to be the same calculator sharing one piece of code, so one
component serves both. Every one of the eleven now reproduces its recorded values
exactly.

**The gate paid for itself several times over, on my own work.** Three worth
telling you about. A piece of ordinary text — the phrase "58-day", in quotes —
got dropped into a calculator's code by the templating system and produced an
error that killed the whole tool; it showed £0.00 for every input while passing
every structural check we have. The same layout rule broke two tools in opposite
directions because their original stylesheets happened to disagree, which nothing
about reading either page could have told me. And my first attempt to improve the
harness silently moved its test click onto the navigation menu on nine of twelve
pages — the gate had stopped testing the calculators at all, and only comparing
two full measurements against each other showed it.

I also found real defects and split them by a rule I want to state plainly,
because I had to apply it against myself. Fixes that provably cannot change what
a visitor sees went in — a "clear" button that was wiping all browser storage
rather than its own, two tools counting checkboxes belonging to other parts of the
page, a restore that failed silently. Anything that changes what the page actually
says did not, however worthwhile. One calculator prints three decimal places on a
money figure. Another computes nothing at all on a 0% deal, which is a real
advertised product. A third signals its whole verdict by colour alone, which a
colour-blind reader cannot see — I wrote the fix for that one, the gate caught
that it changed the page's text, and I reverted it. Each is queued as its own
change. **The test is whether a change is visible, not whether it is worth
making**, and being inconsistent about that is how a port quietly becomes an
unreviewable rewrite.

**The platform change went to the review council and came back "revise", on a
point I had missed and four reviewers spotted.** A check the running system does
not recognise yet is skipped — and a set of skipped checks passes. So installing
one of these too early reports green having tested nothing: the precise failure I
built it to eliminate, reproduced by the fix for it. My defence had been "none are
installed yet", which is a fact about today rather than a guard. I have written
the guard. On its first run it caught a fault in *itself* — the standard recipe
for checking what is deployed does not work on that particular service — and said
so rather than telling me my change had not shipped.

Nothing is live yet. Next: the components go into the database and the pages get
rebuilt from them, which is the point at which the site actually becomes editable.
Then the check ships and gets switched on per calculator, in that order.

Two things still need you, both unchanged. **The GitHub token cannot see the
repository holding the site's source** — that needs someone with GitHub admin.
And the question from last week is still open: full decomposition here, or the
cheaper freeze-the-calculators split the neighbouring lane chose. The rewrite did
not depend on the answer; the next step does.

---

**2026-08-02 — you said full decomposition, so that is what this is.**

Nothing is live yet. Everything below is proven offline and waiting on one
deliberate step. I want to flag three things before I ship, because two of them
are changes you should know about and one of them is a near miss.

**The site chrome was already broken, and nobody could see it.** Something outside
this lane created the site's header, footer and page-head yesterday morning, then
queued a rebuild of all 27 pages. Every rebuild reported success and the site did
not change by a single byte — because these pages currently ship their stored
bytes and skip the assembly step entirely. So the chrome sat there unused and
wrong: it pointed at a stylesheet that does not exist (a spelling difference, one
letter), its navigation bar had no links in it at all, and its icon and social-card
images were both missing. The first page I decomposed would have gone out
unstyled and unnavigable. I have replaced all three. The navigation is now
extracted from the existing nav script rather than retyped, so it is the same
menu, just delivered by the server instead of assembled in the reader's browser —
which is also better for search engines and removes a visible flicker on load.

My own note in the runbook said the precondition was "chrome exists". That was
the wrong question. It existed. The right question is whether it *resolves*, and
I have rewritten that check.

**Two changes to what readers see, both deliberate.** First, every page gains a
footer. The pages never had one, and I would not add one on a whim — but your
legal page currently has *zero* links to it from anywhere on the site, which means
in practice it does not exist for a reader or for Google. The footer fixes that.
Second, and this one I want to be explicit about: **the homepage gains the "late
repayment can cause you serious money problems" warning.** It was on the standard
calculator page but not the homepage. I actually tried to strip it out, in the
name of changing as little as possible — and the tooling refused me, because when
I built that component I marked the warning as required and wrote down why: it is
a regulatory warning that belongs next to a credit promotion, and the homepage is
one. It was right and I was wrong. If you would rather it were not there, it is a
one-field change and I will make it.

I did *not* carry the other two lines across — "current market average is 7.9%"
and "updated for the 3.75% base rate". Those are dated facts, and putting them on
a second page just doubles the number of places you have to remember to correct.

**The near miss is worth telling you about, because of how it was caught.** All
six automated checks passed the homepage — the calculator produced identical
numbers, no text had been lost, no links were broken. Then I looked at a
screenshot and saw three paragraphs that had never been on that page. Every check
was asking whether anything had been *lost*; none was asking whether anything had
been *added*. That check now exists, but it exists because I looked at the page.
The same lesson keeps recurring in this work and it is worth writing down plainly:
a test that compares numbers cannot see words, and a test that checks for loss
cannot see gain.

I also found a real hole in the splitting rule itself. On the homepage it
classified the calculator's entire results box — the big monthly figure and the
two totals — as ordinary editable text. The reason is that this one page keeps its
arithmetic in a shared script file rather than in the page, and the rule only ever
read scripts written inside the page. So the safety proof passed while quietly
getting the most important part of the homepage wrong. Fixed, and the rule now
refuses to run at all if it cannot read a script the page depends on.

Where that leaves us: the 27 pages decompose into editable prose plus a proven
calculator component, all twelve calculators produce identical numbers in their
new assembled pages, and a guide page is pixel-identical to today apart from the
new footer. The next step writes it to the database, one page first — and that
first page has a job beyond itself, because it is the test of whether my offline
model of the assembler is actually right. If it disagrees, the other 26 stay put.

The GitHub token still cannot see the repository holding the site's source. That
one needs someone with GitHub admin and is unchanged.

---

**2026-08-02, later — it works, and the check I was most worried about came back clean.**

All 27 pages are now stored as editable pieces rather than as frozen documents:
51 blocks of ordinary prose and 12 calculators, each a proper component. The first
page has been rebuilt and is live, and the rest are queued behind it.

The thing I want to report is the check, because it was the one part of this I
could not honestly claim to have proved.

To test the decomposition without shipping it, I had to write my own copy of the
platform's page-assembler — the code that stitches the header, the sections and
the footer into a finished page. That is a slightly uncomfortable thing to do: if
my copy is wrong in the same way my test is wrong, everything passes and nothing
is actually verified. So I made my copy write down, in advance, the exact page it
predicted the real system would produce, and then compared the two once the real
system had run.

They matched **byte for byte** — 16,649 bytes, identical checksum, down to the
punctuation of a machine-readable block buried in the page header. So the earlier
result ("all 27 pages check out, all 12 calculators produce identical numbers")
can now be read as a fact about the site rather than as a fact about my model of
the site. That distinction sounds pedantic and it is the whole difference between
testing something and agreeing with yourself.

The live page looks exactly like the old one, with the footer added.

**A correction to what I told you earlier.** I said the render was stuck behind
about 325 other jobs and would probably not happen until tomorrow. It took three
hours. My mistake was reading the age of the jobs finishing at that moment as if
it were the wait — but those are by definition the oldest jobs in the queue, so
their age tells you how long the tail is, not how long you will wait. I have
corrected the runbook. It also means the message I asked you about — permission to
push a job to the front — was not as necessary as I made it sound. You chose to
write all 27 anyway, which turned out to be the faster route regardless.

Everything is reversible per page with one command, and every original is backed
up in the database.

---

**2026-08-02, end of the day — all twenty-seven pages are live and every check is green.**

```
pages rebuilt and live      27 of 27
identical to prediction     27 of 27
calculators still correct   12 of 12
failures                    none
```

The site is now stored as sixty-three editable pieces instead of twenty-seven
frozen documents. Every block of text can be rewritten on its own, every
calculator is a proper component that can be restyled or improved, and nothing
computes a different number than it did this morning.

One point about that last claim, because it is easy to overstate. The final check
compares the live calculators against a record taken from the live calculators, so
by itself it only proves the site agrees with itself. What makes it mean something
is that the new record was compared field-by-field against the one taken from your
original hand-built site — ninety-four fields present in both, none of them
changed. Self-consistency and a clean comparison against the original together are
equivalence. Either one on its own is not, and it would have been easy to report
the easy half.

Everything is reversible one page at a time, and every original is still in the
database.

Two things wait on you: the **GitHub token still cannot see the repository holding
the site's source**, and the site is still marked as ours-to-hold rather than
ours-to-rebuild — flipping that is what would let the improvement loop start
touching it, and it is now a real choice rather than a blocked one.

---

**2026-08-02, later still — the site is now ours to rebuild, and the calculators are locked.**

Done as you asked: all twenty-seven pages are flipped from ours-to-hold to
ours-to-rebuild. The improvement loop can now touch this site.

Before flipping I went and read what that flag was actually holding back, because
it turned out to be holding back two specific things and both would have landed on
your calculators. One is the planner, which treats a held page as off-limits and
files it for a person instead of rebuilding it — without that, a page the plan
thinks is missing goes to the generic page builder, which is on record for
producing "a widget-less prose page where an interactive tool belongs". The other
is the section-saver, which refuses held pages outright, because the way it saves a
page is to delete all its pieces and put them back.

Three measurements decided what to do about that.

There is **no plan for this site** — none has ever been created — so the planner
has nothing to act on today. **Nothing is scheduled to walk sites**: of twenty-six
running scheduled jobs, the only one that touches sites dispatches work that
already exists rather than creating any. So the flip is not a starting gun; like
every step before it, it does nothing until something makes work for this site.

And the deciding one: **the section-saver's delete already respects locks.** Its
delete says "and this piece is not locked". So a locked piece survives, and the
blocked attempt raises a review item rather than vanishing.

So the twelve calculators now carry a permanent lock and the fifty-one blocks of
prose do not — text is rewritable, arithmetic is not. I checked that the lock
actually bites rather than assuming it: on the standard calculator page, all five
prose blocks come back writable and the calculator comes back not.

The site did not change by a single byte. A policy change is not a rebuild.

One cost worth naming now rather than discovering later. Those four calculator bugs
still on the list — the three-decimal money, the car finance tool that does nothing
at zero per cent, the consolidation checker that counts a debt towards the balance
but not the interest, the verdict told only by colour — now each need an explicit
unlock before the calculator can be edited. I think that is right. Changing a
calculator whose arithmetic we have proved should take a deliberate act, and this
way an attempt to change one shows up instead of just happening.

---

**3 August 2026 — the four calculator bugs, and the check that nearly wasn't one**

Took the four bugs that were left on the list. The money printed to three decimal
places; the car finance tool did nothing at all if you told it the deal was
interest-free; the consolidation checker counted a debt towards what you owe but
not towards what it costs you; and the "better option" verdict was told by colour
alone, so a colour-blind reader got nothing. All four are now fixed and proved.

The thing worth telling you is what happened when I went to check them.

We have a harness that drives every calculator and compares the answers against a
recorded baseline. It has been the backbone of this whole lane. I fixed the four
bugs, ran it, and it came back green. And for two of the four, **that green meant
nothing at all** — it would have been green whether I had fixed them or not.

The reason is simple once you see it. The harness works out what to type into
each box by taking the number already in it and scaling it — the same, double,
half. That is a sensible way to drive any calculator without hand-writing a
script for each one. But it means it never types anything far from what the page
already shows. The car finance bug only appears at zero per cent, and no doubling
or halving of 8.9 ever gets you to zero. The consolidation bug only appears when
a box is left *empty*, and the harness fills every box it can find. So neither
bug was ever reachable, and the harness had been reporting a confident pass over
both of them all along.

So I wrote a second, smaller check that types the awkward values on purpose. And
because a check nobody has watched fail is worth very little, it also rebuilds
each calculator **as it was before the fix** and requires the same test to come
out differently. All four now do. Three deliberate "nothing should change here"
cases confirm I did not break the working paths.

Then the same trap caught me again, within the hour. That before-and-after check
originally rebuilt the old version from "the latest commit". Which was correct
right up until I committed the fix — after which "the latest commit" *contained*
the fix, both sides became identical, and the check quietly stopped being a check.
I only noticed because I had already changed it to compare the actual readings
rather than just pass/fail; the earlier version would have reported all four as
proven. It now points at a fixed, named commit that cannot move underneath it.

One correction to something we had written down. The note on the consolidation
bug said it made consolidating look *better* than it really was. It is the other
way round — leaving a debt's interest out understates what your current debts
cost, so the tool overstates how bad consolidating looks. It was talking people
out of consolidations that might have suited them. I did not work that out by
argument; I ran the old version and read what it said.

And one near miss I want on the record, because it would have looked exactly like
success. Two of the fixes needed a new piece of text — the "Better option" badge,
and the message explaining why a comparison is being withheld. Each component
carries a list of its fields with a default value for each, and I assumed the page
would fall back to that default. **It does not.** The renderer only ever looks at
what is stored against the page itself; a field it has not been given comes out as
an empty string, with no error. The badge would have shipped as an empty box. The
page would have rendered, the numbers would have been right, and every automatic
check we have would have passed — while the entire accessibility fix was simply
absent. Caught it one step before it went out, and wrote a small tool that fills
those gaps in and refuses to overwrite anything already there.

Both of those traps are now written into the fleet-wide landmines file, because
neither is about this site.

On the consolidation fix there was a real choice to make and I want to flag it
rather than bury it. If you half-fill a debt row, the tool could either quietly
leave that debt out of the sums, or refuse to give you a verdict until you finish
filling it in. I chose to refuse: it still shows you the total you owe, which is a
fact and does not depend on any rate, but the three interest figures become
dashes and the verdict box tells you what is missing. Leaving the debt out would
answer a different question from the one you asked without saying so, and this is
a page about being mis-sold. If you would rather it just excluded the row, that is
a small change.

Where it stands right now: the first of the four pages is queued to rebuild and I
am waiting on it. When it lands I will check it on the live site, put the lock
back on, and send the other three the same way.

**3 August 2026, later — all four are live, and the deploy route was the hard part**

All four calculator fixes are on the live site and I have checked them by driving
the real pages, not a copy: eight test cases, all passing, on production. Nothing
else on the site moved — ten of the twelve calculators are byte-for-byte what they
were, and the two that changed did so only in the ways intended. No arithmetic
anywhere on the site is different except the rounding we set out to fix.

But getting there turned up something bigger than the four bugs, and you should
know about it because it is not really about this site.

I updated the calculators, followed our own written procedure to rebuild the page,
and the job came back saying **complete**. The page was unchanged. Not broken —
unchanged, exactly as before, with the fixes sitting in the database doing nothing.

The cause is a mismatch nobody had reason to notice. When the rebuilder wants to
know which calculator a slot on the page holds, it looks the slot up **by name**.
When we took this site apart in July we named its slots by position — "prose-0",
"tool-2" — deliberately, so that if a paragraph ever went missing the error message
would say which one. Those names don't match anything, so the rebuilder finds
nothing, quietly keeps whatever was already there, and reports success. Every
signal it produces is identical to a job that worked.

That is the part worth escalating. It is not that this site is awkward. It is that
**a rebuild that did nothing at all is indistinguishable from one that worked** —
there is no error, no warning that surfaces, no difference in the job's status. I
have written it up as bug 182 and put it through the diagnosis loop for a second
opinion. I also counted how far it reaches: six sites, and this one is the extreme
case at 100%. The other five are more insidious — they rebuild most of a page and
silently freeze one or two pieces, so nothing looks wrong at all.

There is a working route, and it is not a hack: it is the same route all 27 pages
were originally published through. Build the calculator with the same engine,
write it to the page, let the assembler put the page together. I have written it up
properly with a safety check that refuses to write if it cannot first reproduce
what is already there.

One more near miss on the way. The tool I wrote to compare "with the fix" against
"without the fix" was caching its results by which calculator it was building — but
not by *which version*. So it built the old one, then built the new one, got the
cached old one back, and told me they were identical. It got the right answer for
the two fixes that added new text and the wrong answer for the two that only
changed code — which is to say it was wrong precisely where being wrong reads as
"nothing to do, your fix is already there". Fixed, and both of those traps are now
in the fleet-wide landmines file.

Where it stands: four bugs fixed and proved, twelve calculators locked again, a new
baseline recorded, one platform bug filed. Next on the list is the header's link
list, which is still hand-maintained and will go stale the first time a page is
added or removed.

Update, same day, from the thread that picked up the platform bug: it's fixed and
live. The rebuilder now looks a section up by the identity the page itself already
carries, not just by its name, so this site's positional slots resolve properly.

While proving that, they found a second problem in the same neighbourhood: firing
the rebuild on the locked calculator duplicated it on the page instead of leaving
it alone. They fixed the live page back to normal in the same session and filed
the new problem separately (bug 189) rather than folding it into this one. The
practical upshot for us: don't fire the documented rebuild on any of our other
locked calculator sections until that second bug is closed — each one would hit
the same duplication the first time. The four fixes we already shipped through the
offline route aren't affected; that route doesn't go anywhere near the part that's
broken.

**3 August 2026, evening — your two decisions, done**

Both carried out. The duplicate calculator page is gone, and the consolidation
checker keeps refusing to give a verdict it can't stand behind.

On the page: it now returns "not found", the sitemap has gone from 27 entries to 26,
and I checked the other 26 pages are all still serving normally. Before deleting it I
checked three separate places for anything linking to it — the page bodies, the
header and footer, and the navigation tables — and ran the same check against a page
I *knew* was linked, to prove the check actually works rather than just returning
nothing. Clean on all three.

One thing that made this more than a delete. Your sitemap — the file that tells
Google what pages exist — was written once when the site was adopted and the platform
has never touched it since. There is a purpose-built tool for retiring a page, built
by another thread this week, and it deliberately leaves sitemaps alone, because on
most sites the platform regenerates them. On yours it doesn't. So using that tool
would have deleted the page and left the sitemap still pointing at it — which is
exactly the bug that tool was built to fix. I removed the page and its sitemap entry
together instead, and I've written that up for the thread that built it, because it
probably affects every adopted site rather than just this one.

Two honest notes about my own work. First, when I went looking for links to the page,
my initial check found two — and both turned out to be *comments* in the page source,
one of them written by me three hours earlier. Had I trusted it I would have refused
to delete a page nothing actually linked to. Second, when I pushed the deletion, it
was rejected twice because other work was pushing to the same place at the same
moment — and the command I'd written cheerfully reported success anyway, because it
was reading my own copy rather than the real one. Both fixed, both written down.

And a small tidy that fell out of it. The footer carried a note in its source
explaining which pages were orphaned and why — visible to anyone viewing the page
source, on all 27 pages. The moment the page went, that note was describing something
that no longer exists. It's now two lines pointing at our own documentation. That's
the third time today the same mistake has surfaced: engineering notes written into
something that gets published. The existing pages will pick up the corrected footer
as they naturally rebuild; I didn't think an invisible comment justified forcing 26
rebuilds.

Nothing is outstanding on either decision. The next real piece of work is the header's
link list, which I looked at earlier and which turns out to need a decision from the
platform side before it can be automated safely.

---

## 2026-08-04, evening — a second check on a second update, and one job left half done

The platform updated twice today rather than once. This morning's update was checked
and passed; a second one landed at about half past eight this evening, so the same
check was owed again. It passed too: the three pages I tested come back
character-for-character identical to what the old version produced, all eleven
calculators still give exactly the right answers, and all twenty-six pages are up.

This time the check was done properly, and the difference is worth explaining because
this morning's was not. A check like this compares "before" against "after", so the
"before" has to be a clean picture. This morning it wasn't — those pages had a small
correction of mine already queued up inside them, waiting to appear, so they were
always going to come out different and the check couldn't have told me anything. I
noticed afterwards and wrote it down. This evening I used the only three pages on the
site that had already absorbed that correction and so had nothing else pending in
them. Nothing to predict, nothing to explain away.

I also found a mistake in the notes I handed over this morning, and it was mine. They
told the next person that a particular platform bug was still open and that a certain
routine operation was therefore harmless on this site. Both had stopped being true the
day before — and my own working notes, in the same folder, said so plainly several
pages up. I'd built this morning's summary by carrying paragraphs forward from
yesterday's without checking them against my own notes. The important part isn't that
the fact changed; it's that the danger reversed. That operation used to do nothing
here. It now does something actively damaging — it duplicates a calculator on the
page — and anyone following my summary would have concluded the opposite. Corrected,
with the reasoning left visible rather than tidied away.

Then the honest bit. I'd decided to finish off the footer correction I judged not
worth the effort yesterday, because I found two better reasons for doing it: the old
note names a page we deleted, so twenty-three live pages are publishing something
untrue in their source, and it's the very thing that spoiled this morning's check. I
started carefully — two test pages first, to prove the footer is the only thing that
would change — and then the platform's page-rebuilding queue stopped moving. Not
because of anything I did: there are over two hundred rebuild jobs queued from another
piece of work on two other sites, and none of them are being picked up either, though
the rest of the system is running normally.

So I stopped. I did not queue the remaining twenty-one, because adding jobs to a queue
that isn't moving helps nobody and makes it harder for the next person to tell my work
from the backlog. I did not re-send the two that are waiting, because there's a
documented trap there — a job that looks lost is usually just queued, and re-sending
creates duplicates. And I did not start diagnosing why the queue is stuck, because
that's a different piece of work on someone else's territory and guessing at it in
writing is how bad theories get inherited.

What's left is small and precisely written down: two jobs waiting in the queue, a
check to run on them when they land, and twenty-one to fire after that. Whoever picks
this up should read the status of those two before doing anything else.

**Later the same evening — the queue freed up briefly, and what I found changes the
size of this job.** Correcting what I wrote an hour ago: I said I'd stopped and left
twenty-one jobs unqueued. The queue started moving again on its own after about a
quarter of an hour, my two test pages went through, and I've now queued the rest. But
the important part is what the two test pages showed.

I'd expected them to change in exactly one way — the footer note. One of them did. The
other changed in three ways: the footer, plus it gained a "canonical" tag and lost an
empty description tag. Those two extra things are a genuine improvement the platform
started producing on the 2nd of August, and pages only pick it up when they next
rebuild.

So I checked all twenty-six pages properly, and **twenty of them have no canonical tag
and carry an empty description tag**. That matters for search engines: a canonical tag
is how a page tells Google "this is my real address", and an empty description tag is
worse than none at all. This has been true for two days on a site whose whole purpose
is being found.

Which means my judgement yesterday — "an invisible comment, not worth twenty-six
rebuilds" — was wrong, and wrong in an instructive way. I had decided what was waiting
to happen without ever checking what was waiting to happen. On a system where
improvements land continuously and only reach a page when that page rebuilds, whatever
is pending on a stale page is *everything that has shipped since it last rebuilt* —
which grows every day and can't be guessed at. I guessed, and I guessed low.

The other thing worth saying: I ran two test pages rather than one, and they disagreed
with each other. Had I run only the calculator page I'd have concluded "just the
footer" and been wrong about twenty pages. That check cost about two minutes.

Twenty-one jobs are now queued. The queue is moving in fits and starts because of that
other backlog, so they may take a while. Everything needed to confirm they worked is
written down, including a check that needs nothing kept from tonight.

**Finished, later that night.** All twenty-one queued jobs went through — first time,
no retries, which says the jobs were always fine and only the queue was slow. All
twenty-six pages now have a proper canonical tag, a correct description tag, and the
corrected footer. The calculators still give exactly the right answers, every page is
up, and the deleted page is still deleted.

One detail I'd want you to know, because it nearly went the other way. Before trusting
the "all clear", I ran the check against a page I *knew* was still broken, to confirm
the check could actually detect a problem. It could — it flagged all three faults. A
check that has never been seen to fail is not evidence of anything, and an hour earlier
tonight I'd been caught by exactly that: I'd read a run of zeroes as good news when the
zeroes actually meant I'd downloaded an error message instead of a web page.

The thing I'd flag for the future isn't this site — it's that this will quietly happen
again. Every time the platform improves how pages are built, every page that doesn't
rebuild is left behind, and nothing announces it. Tonight that was twenty pages missing
a tag that tells Google which address is real, for two days, on a site whose whole job
is being found. Nobody did anything wrong to cause it. It's just what "applies on next
rebuild" means when pages don't rebuild often. Worth deciding whether we want something
that watches for it rather than someone noticing by accident.
