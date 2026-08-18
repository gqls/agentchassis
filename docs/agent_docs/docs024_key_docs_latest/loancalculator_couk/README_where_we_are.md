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

---

## 2026-08-05 — the copy goes back through the framework, and why it couldn't yet

You asked for two things today: the gentler writing style on this site too, and for the
copy to be produced by the framework rather than typed by me. The second one is the
interesting story.

The style is seeded and live. I took the prompt the other session developed with you
this morning, adapted the worked examples to this site's own copy, and put it where the
writing pipeline actually reads from. One detail worth knowing because it nearly caught
me: the writer reads exactly *one* field, a flattened text version of the style spec,
and if you edit the spec without regenerating that flattened copy the change is
completely invisible while looking perfectly applied. The script I wrote refuses to run
unless it can first reproduce the existing flattened copy exactly, which is how it
proves it's regenerating it correctly rather than mangling it.

Then I hit the thing you'd already put your finger on. The prose was labelled in the
system as "authored" — meaning a human wrote this, don't regenerate it — and the writing
pipeline skips anything with that label entirely. You said, correctly, that the label is
just wrong: that prose came from a different AI, outside this framework and without its
checks. So it isn't authored in any sense that matters, and correcting the label is a
fix rather than a loosening. I changed it, and checked carefully that I changed only
that one: there were exactly two such labels in the whole system, and the other one
belongs to a deliberate "copy this site exactly, don't touch a word" mode that would be
genuinely damaged by the same edit.

**And then it still didn't work, for a deeper reason.** Rather than assume, I ran one
real page through the pipeline to see. It refused — politely and clearly, which is to
its credit — saying it couldn't find anything to build. The cause is that when we broke
this site into editable pieces we named them by position: "prose-0", "prose-1". The
build pipeline looks up pieces by *type* name, not by the direct link it already holds
to the actual component. So it looks for something called "prose-0", finds nothing, and
gives up. Not one page — **none of them**. All 57 pieces on this site are invisible to
it, and four other sites are partly affected.

This is the same fault we fixed three days ago in the neighbouring function, and the
commit that fixed it actually edited this file while it was there — it just fixed one of
the two places. That's a recurring shape here and it's now written down as such.

One thing I'd flag as more than a nuisance: when the build pipeline can't find a piece,
it doesn't just fail, it asks the system to *build a new component* with that name. My
single test page generated four such requests, for components that already exist. A
full run would have generated over a hundred. I cancelled them all before anything acted
on them, but it's worth knowing that's what a failed attempt leaves behind.

So: nothing on the site has changed, nothing is broken, and the style is loaded and
waiting. What's needed next is a small fix to the build pipeline — the same fix, one
function over — and then the rewrite can actually run. That's proper code rather than
configuration, so it goes through the review gate. I've written it all up in detail
rather than starting it at the end of a long session.

Two smaller things for when you next look. Your choice to put the new style into the
*base* prompt for all future sites turns out to be bigger than it sounded: that prompt
isn't one thing, it's seven copies across seven agents, and they've already drifted
apart from each other. So it's either seven edits that will drift again, or one shared
source they all read. I haven't picked; that's a design decision worth making
deliberately. And the base prompt currently tells writers to "start with the fact",
which is close to the opposite of the new style's "start where the reader is standing".
They agree on the important half, so it's reconcilable, but it's a change to an existing
instruction rather than an addition, and the reviewers will rightly want that said out
loud.

---

## 8 August — starting the rewrites you approved

You've seen the new voice on one page and said to go ahead, so that's what's happening
now. Two things worth saying before the copy starts changing.

**The thing that was blocking this is genuinely fixed.** Back on the 5th I couldn't
rewrite a single page, because the build pipeline looked up the pieces of a page by
type name and this site's pieces are named by position — "prose-0", "prose-1" — so it
found nothing and gave up on all 57 of them. That got fixed properly in the code (by
the bug lane, not here), reviewed, and proved on a real page: both pieces resolved, the
prose actually changed, and none of the junk "please build me a component called
prose-0" requests appeared. That proved page is the one you looked at.

**I re-checked everything this morning rather than trusting the earlier check**, because
a new build of the system went out while I was setting up, and because another session
had edited the writing agent's settings the day before. The style is still loaded and
still correct; the labels that let the pipeline touch this prose are still right; the
two settings the fix depends on are still there. One of those checks looked like it had
been wiped and hadn't — I'd looked in the wrong place in a nested settings file, which
gives you exactly the same answer as "someone reverted it". I've written that trap down,
because on this row that revert was a real risk and I'd have believed it.

**How it will run.** Twenty-five pages left, in batches of four or five rather than all
at once, so a bad batch costs a batch. Every page goes through the framework — the same
route as the page you approved, with the instructions copied across by machine rather
than retyped, so what you saw is exactly what each page gets. The calculators are safe
twice over: the system never sends a calculator to the writer at all, and the locks you
have on those twelve rows would stop an overwrite even if it did. I'll be checking after
each batch that the calculator rows haven't so much as changed their timestamps, and at
the end that all twelve still compute the same answers they did before.

**One page I want to flag now rather than after.** The legal page is copy, so by your
instruction it's in scope, and I'll do it — but it's the one page where a nicer-sounding
rewrite could be a worse page, if a disclaimer or a statutory reference comes out softer
than it went in. So it goes last, and I'll read its before-and-after myself line by line
instead of just confirming it changed. If anything shifts in meaning rather than in
phrasing, I'll put that one page back as it was and tell you.

Everything is backed up first — every piece of copy on the site, with its original text
saved, so any page or the whole site can go back exactly as it was.

---

## 8 August, later — the rewrites are done: 23 of 26 pages, calculators proven untouched

**Where it landed.** Twenty-three of the site's twenty-six pages now read in the new
voice. All twenty-six are up and healthy. The calculators are the thing I was most
careful about and they are provably fine: all twelve locked calculator blocks are
byte-for-byte what they were before I started, and the harness that drives every
calculator and checks its arithmetic reports **all eleven reproducing their expected
values exactly**. Nothing about the maths has moved.

**The copy itself is doing what you liked about the sample.** It opens where the reader
is standing rather than with a flat assertion, it explains a thing before naming it —
"you'll need to ask your lender for the total amount required to close the account
today; lenders call this the Settlement Figure" — and it quietly drops the salesy
absolutes. "Overpaying is almost always a smart financial move" became "For most
people, overpaying is worth doing."

**Three pages I could not convert, and it is not a copy problem.** The home page, the
car finance calculator and the interest rate stress test are refused by the platform's
own content checker, every single time. The reason turned out to be nothing to do with
the writing: each of those three pages contains a calculator whose code carries a
developer's note explaining how it works, and one of the words in that note is on the
checker's list of "things a machine writes when it talks about its own task instead of
doing the work". The checker reads the whole page including the invisible notes, sees
that word, and refuses. Those notes have been on the live site since 3 August and are
harmless in themselves. I proved it rather than guessed: exactly the three pages with
that word in their calculator failed, and all nine without it passed. **Those three
pages are untouched and working** — they simply still read in the old voice. I've
written it up as a bug for the platform to fix, and the moment it is fixed those three
take about three minutes each.

**One thing I broke and fixed, worth telling you about.** On four pages the rewrite
deleted the styling that lays the calculator out. When the site was broken into editable
pieces, some pieces that got labelled "prose" were actually the page's stylesheet, and
rewriting prose in those slots threw the styling away — while every safety check in the
system said fine, because the calculator itself was untouched and only its appearance
collapsed. I caught it on the first page, put that page back, then found the other three
and put those back too, and confirmed each one on the live site rather than in the
database. All four are correct now. I've written the trap down so the next person to
rewrite a site like this checks for it before starting rather than after.

**One decision I'd like you to make.** On several calculator pages the intro text was
almost empty — a single line, sometimes thirty characters. The framework didn't just
restyle those, it filled them out, in some cases to a page of new explanation. It reads
well and nothing in it trips the site's claim checks, but it is **new material rather
than a rewrite of yours**, and on a money site that is your call, not mine. If you'd
rather those slots stayed short, say so and I'll trim them back; if you're happy, they
stay. Everything is backed up either way.

**Not touched, as agreed:** the other finance sites are still waiting on your review,
and the fleet-wide change to the house voice is still a separate decision that needs a
proper design round rather than being slipped in here.

---

**Saturday 8 August, afternoon — the last three pages are unblocked, but the fix
needs a release before they can be rebuilt.**

This morning I said three pages could not be rebuilt because of a platform bug, and
that when it was fixed each page would take about three minutes. The bug is now
fixed, reviewed and approved. It is not yet *live* — the code has to go out with the
next whole-fleet release, which is yours to run. Until then those three pages sit
exactly as they were: serving fine, in the old voice, unchanged and healthy.

**When you next run a release, that fix goes with it**, and I can finish the last
three pages and re-baseline the calculator golden the same afternoon. The command is
the usual one:

```
date; make release redeploy-agents ENVIRONMENT=production REGION=uk001; date
```

I have set the version tag to `v1.0.1265` ready for it. There is no rush from my side
and nothing degrades while it waits.

**What the bug turned out to be is worth two minutes of your time, because I got it
wrong first and the wrong version would have looked fixed.**

The checker that reads every page before it saves has a list of phrases that should
never appear in customer copy — the sort of thing a machine writes when it talks
about its task instead of doing it. Sensible check. But it was reading the *whole
file*, including the parts no reader ever sees: the code, and the notes the
programmer leaves in the code. Our three calculators each carry a careful note
explaining how the arithmetic works and why. Those notes mention the word the checker
objects to — because that word is the name of the thing they are explaining. So the
checker refused to save the page, and its error message blamed the machine for
something a human had written five days earlier.

This morning I wrote down that those notes were in one kind of comment, and proposed
the obvious fix: ignore that kind of comment. **They are in a different kind.** Two
of the three files contain none of the sort I named. Had I built the fix I proposed,
it would have passed review, shipped, changed nothing, and I would have recorded the
bug as fixed. What caught it was asking the database to pull out just the comments
and count — it found none of the offending word, where counting the whole file found
three. Two counts that had to agree, and did not.

The real fix is smaller and better: the checker now reads only the words a visitor
actually sees. I proved it on the exact pages that failed — the system had kept the
precise files from this morning's failures, so I could run the new checker over the
bytes that broke it rather than over an example I invented. All three now pass, and
the check still catches the thing it is for.

**Two side-effects, one of which you may want to act on.**

While measuring whether this could break anything elsewhere, I found the same checker
is quietly blocking a page on **webdesign.co.uk** — the tools index. Its copy says
*"LocalBusiness schema, as an AI-builder prompt"*, which is a perfectly good
description of what that tool does, and the checker sees the phrase "as an AI" and
refuses. Nothing is broken today and the page serves normally; but the next time
anyone asks that page for a content change, it will fail. My fix does **not** solve
that one — that copy really is visible text — so I have written it up separately and
told that lane rather than quietly reword their page to suit a scanner. It needs its
own small fix.

The other: the reviewers approved my change but caught a genuine gap — I had measured
one thing and not another, and was leaning on the answer being obvious. It was, but
they were right that I had not checked. I have since checked it properly. Worth
recording because that is the review working, not the review being awkward.

**Nothing else has changed.** The other 23 pages are still in the new voice, the
calculators are still proven untouched, and the question I asked you yesterday — the
one about whether the framework was right to *fill* those near-empty pages with new
explanatory copy rather than just restyling them — is still open and still yours.

---

**Saturday 8 August, later — your ruling on the homepage, and I have taken the
caution back out.**

I had flagged the homepage as needing a decision before rebuilding it, on the grounds
that it still carries the original hand-built copy and that a framework rebuild would
replace the paragraphs you liked as well as the opening you objected to. You have
ruled the other way:

> "I am happy for the whole site to be built through the framework because that is
> what I am judging."

That is the right call and I have removed the flag from the handoff rather than leave
a caution you have overruled. The reasoning generalises, so I have written it down
next to the instruction: **keeping hand-built copy because it reads well protects the
wrong thing.** What is being judged here is what the framework produces. A page that
holds on to its hand-written paragraphs is a page that tells you nothing about that —
and if the rebuild makes the homepage worse, that is the finding, not an accident to
be avoided. The pre-run copy is backed up either way, so nothing is lost by trying it.

So when the release goes out, all three remaining pages go through together — the
homepage, the car finance calculator and the interest rate stress test — with no
special handling for any of them.

---

**Saturday 8 August, evening — it is done. All twenty-six pages are in the new
voice, and the calculators still give exactly the same answers.**

Your release went out and carried the fix. I checked it had actually arrived rather
than trusting that a deployment happened: I looked inside the running program for a
phrase the fix adds and a phrase it removes, on both copies of the service, and then
did the same on an older one still running the previous version to make sure the check
could tell the difference. It could.

Then the last three pages went through the framework — the homepage included, as you
ruled. All three built cleanly. This morning the same three failed every time.

**What I checked before telling you it worked.** Not the status — a job can report
success and have saved nothing. For each page: that the text in the database is
genuinely new (the rows were replaced, not left alone), that every number, price,
percentage and legal reference from the old copy is still there, that the locked
calculators were not touched, and that the live page on the internet is serving the
new words and none of the old ones.

I also deliberately broke the checker to make sure it was capable of failing — I fed
it a version of the "before" record that should have made it complain, and it
complained about all four rows. A test that has never failed isn't a test.

**The calculators.** Before re-recording what "correct" looks like, I ran the existing
record against the live site: all eleven tools produce identical numbers, including
the two whose pages had just been rewritten. That order matters — if I had re-recorded
first, I would have quietly enshrined whatever the rebuild did, and the one step that
could have caught a problem would never have run. Only then did I take the new
reference recording, which now also covers a test case the old one predated.

**One thing I nearly got wrong, and it is the sort you would never have caught.** When
I counted how many pages were in the new voice, I got twenty-five out of twenty-six,
and was about to tell you one page had been missed. It hadn't. The page in question is
the very first one we did — the one you reviewed and approved — and it was rewritten a
day before the main run, so the date I used to count "done" excluded it. The count was
right and the question was wrong. I have written that down, because a wrong number
attached to the word "complete" is the kind that gets repeated.

**Where that leaves us.** Twenty-six of twenty-six pages in the new voice, every page
loading, every calculator unchanged. The platform bug is fixed, live and proven, and
I have left the record of it open for you rather than filing it away.

**Two things still want you, neither urgent.** The question from yesterday about
whether the framework was right to *fill* the near-empty pages with new explanatory
copy rather than just restyling them — still open, still yours. And the separate
problem I found on webdesign.co.uk, where a page cannot currently be rebuilt because
its own copy says "as an AI-builder prompt" and the checker objects to the phrase.
That one needs a small fix of its own and I have left it with that lane.

---

**Saturday 8 August, later still — you have answered the last open question: keep the
explanatory copy.**

So nothing is trimmed, nothing is re-run, and the pages stay exactly as the framework
built them. That was the last thing owed on this site, so the workstream is finished.

For the record of what you decided, because in six weeks the question will be harder to
reconstruct than the answer: several of the calculator pages had almost nothing on them
— a line or two, sometimes barely a sentence. When the framework rewrote the site it
did not just restyle those, it **wrote them properly**, turning a stub into eight or
nine hundred words of explanation. That was more than the brief asked for. The brief
said change the voice and add nothing. So there were two fair readings — the pages are
better and readers are served, or the machine exceeded its instruction and we should
put it back. You have taken the first.

I have also taken your answer as covering the two smaller things I listed under the
same question: the writer expanded "Consumer Credit Act" to "Consumer Credit Act 1974",
which is correct but still an addition, and it reworded two headings, one of which was
a main page heading. They are the same act by the same writer and were asked as part of
the same question. If you meant only the long copy and want those two put back, say so
and it is a few minutes' work — but I have not held anything up waiting to find out.

**Where that leaves the site.** All twenty-six pages in the new voice, every page
loading, every calculator giving identical answers to before, and the reference
recording refreshed. The platform bug that blocked the last three is fixed, live and
proven. The only thing left with our name on it is the separate problem on
webdesign.co.uk, which belongs to that lane and is written up for them.

---

## 2026-08-08, late — the whole site now speaks in the new voice, and the homepage says what you approved

The platform fix we were waiting on arrived in tonight's build, so the last three pages
could finally be rebuilt. That's done. All twenty-six pages are now in the new voice —
nothing is left in the old one — and the homepage opens with the copy you signed off.
Both phrases you objected to are gone from it. The calculators still give exactly the
right answers and every page is up.

Two things worth telling you, because both are mistakes of mine rather than the system's.

I told you twice this week that the platform fix wasn't live yet. I was checking for a
piece of text that only exists in the fix's *test* file, never in the shipped program, so
my check was always going to say "not there" whether it was or not. What settled it in
about a minute was simply trying to rebuild one of the blocked pages and watching it work.
When a fix doesn't add any new visible text, there's nothing to look for, and trying it is
the only real test.

And when I applied your homepage copy, it appeared three times on the page. I'd written
the instruction as "replace the opening block" — but the system writes a page one section
at a time, and each section only ever sees itself. So three different sections each
decided they were the opening block and each inserted your paragraph. Two of them had
their own content destroyed in the process, which I restored from backup. The page is
correct now.

That's the fourth time this week the same thing has bitten from a different direction. A
rule, a worked example, a stored instruction and now a page-level note have each been
obeyed by every section, because obeying everything uniformly is the only thing a
section-by-section writer can do. It's the clearest argument yet for giving the writer the
whole page instead — which is on the list, and which I'd now put above almost everything
else on it.

One small thing still owed, and it's my fault. Restoring those two sections brought back
their old wording, and one of them still says "your exact monthly repayments" and "the
true total cost" — the same register you objected to, one line below the paragraph we
fixed. It needs one careful rewrite, written to that section rather than to the page, and
I've left precise instructions.

---

**Saturday 8 August, late — you asked me to pick up the planner bug and leave it ready
for a fresh session. It is written, tested as far as it can be tested without spending a
run, and not switched on.**

A reminder of what the bug is, because it is a good one. There is an agent whose job is
to plan an "experience" — a whole journey across a site rather than a single page. When I
pointed it at this site to settle the debt page ordering, it came back with a confident,
detailed plan about a completely different site: a game, its daily provocation, its
timed round. The reason turned out to be simple and slightly embarrassing. Somebody had
written that other site's situation directly into the agent's instructions — its broken
pages, the decisions you had made about it, the exact file its widgets read. Every time
the agent runs, for any site, it is told that *that* is the problem it is fixing. It went
unnoticed for three weeks because nobody had ever run it anywhere else. Sixty-one plans
exist; fifty-nine are that one site's.

**Three things I found while writing the fix, and each one made the fix bigger.**

First, I said yesterday it was in one place. Then I measured and told you three. It is
five. My own count was wrong twice, for a daft reason worth admitting: I searched for the
word "gauntlet" in lower case, and in two of the files it is written "Gauntlet" with a
capital. So two whole sections came back clean when they were not. I have logged that
where we log wrong calls; it is a recurrence of a rule we already have written down.

Second — and this is the one that matters — **the reviewers are contaminated too.** This
agent writes a plan and then a small council of critics judges it. I said yesterday that
the council was the part that worked, because it correctly refused the nonsense plan.
That was too generous and I have corrected it. Three of the four critics are themselves
told what the *other* site's data file is called, what its core loop is, and what counts
as a fabricated number *there*. The refusal was right, but the plan was so obviously
wrong that both a fair critic and a contaminated one would have rejected it, so it tells
us nothing about whether the critics are sound. The practical danger is the reverse case:
once the fix lands, a *correct* plan for this site could still be objected to by a critic
looking for a feed and a timer this site was never going to have — and that would look
exactly like the fix having failed.

Third, the hardcoded facts have gone stale as well as being about the wrong site. One
critic — the one that can veto on its own for dishonesty — is told that the other site has
no verified facts at all, so any number in a plan must be invented. That was true when it
was written in July. It stopped being true at nine o'clock this morning, when that site
gained four verified facts. So that critic is now telling itself four real facts do not
exist. Nothing goes back and updates a fact pinned inside an instruction when the world
moves; that is an argument for the change I have made even if the original bug had never
happened.

**What the fix does.** It takes the site-specific brief out of the shared instructions and
puts it where it belongs: attached to the site, as data. Each experience can have its own
brief; the agent reads whichever one applies. If a site has no brief — which will be the
normal case — the agent is told so explicitly, and told not to borrow anyone else's. The
other site keeps its brief word for word; it just moves house, in the same single
operation that removes it from the shared prompt, so there is no moment where it is lost.

**What I have deliberately not done: switched it on.** I ran the whole thing against the
real database with the final "save" turned into "discard" — so every safety check, every
edit and the verification all actually executed, and then it was thrown away. It passed,
and I confirmed afterwards that it left no trace. That proves the change is
mechanically sound. It does **not** prove the agent then writes a better plan, because
that costs a real run of about half an hour, and the fresh session should spend it and
watch the result rather than inherit my word for it. Two paren-level errors in the SQL
were caught by that dry run and by nothing else, which is the argument for doing it that
way round.

There is a second, separate fault in the same area that I have written up but not touched:
when the council rejects a plan, the rejected plan still becomes the official one. It was
demoted by hand at the time so nothing is building from it. Fixing it properly is a choice
between two approaches and I have laid both out rather than picking one, because one of
them changes shared machinery and should go through review.

Everything is in `HANDOFF_2026-08-09_continue_here.md`. The site itself remains finished
and untouched by any of this.

---

**2026-08-09, morning.** The last line owed on the homepage is done, and it turned up
something worth knowing.

The line was the one under "Standard Loan Calculator": *"Calculate your exact monthly
repayments and see the true total cost of borrowing."* You had already objected to
exactly that language in the paragraph above it — too strong, positioning us as the
accuracy authority when nobody was asking. The paragraph got fixed on Thursday; this
line one row below it did not, because it was put back by a restore after that session's
rewrite went wrong. It now reads: *"Enter your loan amount, rate and term below to see
how the monthly figure and the total cost move together."* The framework wrote it, not me.

Before touching it I checked whether it really was the only line left in that register
anywhere on the site. It was. There is one other page that says "understanding exactly
what you're signing", which is about the reader's own contract rather than a boast about
our sums, so I left it alone.

**The thing worth knowing.** Thursday's session worked out why its rewrite went wrong:
the writer only ever sees one section at a time and never its neighbours, so a page-level
instruction gets applied by every section that thinks it might qualify. Its advice was to
phrase instructions conditionally — "if this section is the one I mean, do this,
otherwise leave yourself alone". I did that. **It still leaked.** The heading section
immediately above the target read "the introduction under the Standard Loan Calculator
heading", concluded that meant it, and wrote its own version of the new sentence. If I
had relied on the wording alone, the page would now say much the same thing twice and I
would have repeated Thursday's mistake with a better-written prompt.

What saved it was that I had locked every section on the page except the one I wanted
changed, so there was physically nowhere else for the writing to land. Two of the three
locked sections came back untouched; the third came back rewritten and was refused. So
the honest version of Thursday's lesson is: word it carefully **and** lock the neighbours.
Wording alone is not a control.

I also nearly got the diagnosis wrong in a way worth recording. The system files a notice
every time a locked section blocks a change, and three of those appeared. My first thought
was "there's the proof it leaked". It isn't — that notice fires whether the writer changed
the section or handed it back identical, so it would have appeared regardless. The actual
proof needed a different check, comparing what the writer proposed against what was there
before. I've flagged that one for the fleet, because the notice looks exactly like a
detector for this and is not one.

Everything else checks out: all 26 pages serving, all 11 calculators still producing
identical numbers to the saved reference, and the opening paragraph you approved is
untouched and appears once. I've left that paragraph permanently locked, since it is your
copy and the site has already destroyed it once.

One correction to the file above: Friday night's note ended "the site itself remains
finished and untouched". Two sessions wrote handoffs fourteen minutes apart that night and
neither knew what the other had found — one said the site was finished, the other recorded
this outstanding line. The second was right. Both files now say so.

---

**Saturday afternoon, 9 August.** The planner fix is applied and it works. This is the one
where our experience planner had another site's homework written into its prompt, so any
plan it wrote for us came back describing that site's pages. The repair moves each site's
briefing out of the shared prompt and into the database, where it belongs, so the planner
reads *your* site's brief and nobody else's.

It is live now, and we proved it twice over. Asked to plan the "struggling with repayments"
experience for us, it produced a plan about loans and debt with no trace of the other site —
the first time that has ever happened. Then we asked it to plan the other site's game, and it
correctly produced that site's plan from that site's brief, which we had moved across word for
word so nothing was lost. Both plans went to the review council and both were approved. That
last part matters more than it sounds: three of the four reviewers had also been holding the
other site's rules as their general standard, so we had expected them to complain about a
missing feed and a missing timer that our site was never going to have. They didn't.

**The thing worth telling you about is that our own test was broken, and it passed.** We had
written down, in three separate places, how to prove the fix worked: check that the prompt
contains the phrase "no brief on file". Ours did, so we ticked it off. But that phrase is also
sitting in the instructions we send the model every single time — so the check would have said
"passed" even if the whole mechanism had never been connected, which is exactly the failure it
was written to catch. We only found out because the second test, on the other site, came back
saying "no brief on file" for a site that definitely has one. For a minute that looked like the
fix had failed. It hadn't; the test had. We've corrected it everywhere it was written down and
logged it for the fleet, because it is a nasty shape: the reassuring answer, arriving on the
first try, from a check nobody would look at twice.

One side effect I want to flag, because it is someone else's site. Running that second test
caused the other site's plan of record to be replaced with a freshly written one — the system
saves a plan *before* the council has voted on it, which is a known second fault in this same
bug and is still unfixed. I have put their original plan back the way I found it and left the
new one on file, marked up, in case they want it. But it means this fault is worse than we had
written down: we thought it could leave a *rejected* plan sitting there as official. In fact
any test run at all quietly overwrites whatever was there.

That second fault is the one thing still outstanding, and it needs a decision from you at some
point, because the two ways to fix it are genuinely different — one is a small config change
to this planner only, the other is a change to shared machinery that every plan-writing agent
uses. I have not touched it. Everything else on this job is done and written up.

---

**Monday morning, 10 August.** Two things this morning. First, the fix from Saturday survived
the new build — I checked it against the freshly rolled system rather than assuming, and while
I was there I finally pinned down something Saturday's notes had left as a mystery. Every
deploy quietly touches almost every agent record in the database, which had looked like it
might be overwriting our configuration changes on each release. It isn't: it only stamps which
image each agent should run. I measured it column by column across the roll — 189 records had
their image tag updated, four had configuration changes, and all four belong to other people's
work. **So a database-only fix survives a rebuild.** That is worth knowing generally, because
it is the thing that makes a config change safe to ship without waiting for a build.

Second, you chose the smaller of the two routes for the remaining fault — the one where a plan
was being saved as official before the review council had voted on it. That is now done and
live. The plan is only written once the council approves; a vetoed or escalated run leaves
nothing behind. It also fixes something we hadn't framed as a fault: a run that took three
attempts used to leave three plans behind, each briefly official. Now it leaves one.

One honest caveat, and I want to be plain about it because it is the same trap I fell into on
Saturday. I have a test running that proves the approved path writes exactly one plan instead
of three. That is real evidence, but it is evidence about the path that was already working.
**The whole point of this fix is what happens when the council says no — and I cannot make the
council say no on demand.** Both runs so far were approved. So I have recorded this as proven
for the approved case and still owed for the rejected one, rather than calling it finished. It
needs either a natural rejection to come along, or a deliberately impossible experience seeded
to force one.

The one thing I'd flag for later: running the planner to check something *changes that site's
official plan*. There is no dry-run. That is how Saturday's test quietly replaced the other
site's plan, which I put back. The larger of the two routes you didn't pick would also have
fixed that, so it may be worth revisiting if this bites again.

**Monday afternoon, 10 August.** The thing I said was still owed on Saturday and again this
morning — proof about what happens when the review council says *no* — is now done, and it went
better than I expected.

I couldn't wait for a rejection to turn up naturally, so I made one. I wrote a brief for an
imaginary experience on the loan site that a client might genuinely ask for and that we
genuinely cannot build: a live board showing what other visitors were offered by lenders in the
last hour, a counter of how many people in your area are looking right now, and each visitor's
own history following them from laptop to phone. Everything on our sites is served as flat
files with no server behind them, so none of that is possible. I filed the brief the same way a
real one would be filed, and fired the planner at it.

**The council vetoed it on the first round, exactly as it should have** — four objections,
three of them serious: there is no server to write to, no way to recognise the same person on a
second device, and putting the lender's API key in the page would hand it to anyone who looked.
And the point of the exercise: **while all that was happening, nothing was saved**. I watched
the count of stored plans the whole way through. It stayed at nothing while the plan was being
drafted, stayed at nothing while it was being vetoed, stayed at nothing while the system had
another go — and only became one when the second attempt was approved. Under the old behaviour
that same run would have left two plans behind, and the first of them — the one the council
threw out — would have been the official one. That is precisely the fault we set out to fix,
reproduced deliberately and now absent.

Two smaller things fell out of it. One is a piece of luck I'd rather call a near miss: when the
system has a second go at a plan, it turns out to write to the same place the first attempt
used, which is what makes moving the save to the end safe. I had checked that for one of the two
retry routes; this run checked the other, and if it had gone the other way we'd have been
quietly saving the *rejected* draft. The other is housekeeping — three descriptions inside the
planner still told the old story, one of them stating flatly that a rejected plan stays official.
That is now the opposite of the truth, and a future reader would have gone looking for something
that no longer exists. Corrected and live.

There is still one narrow gap, and I'm leaving it open rather than rounding it up. I've proved
that a *vetoed* draft is never saved. I haven't yet watched a run that gets vetoed and then
gives up entirely, because — and this is the interesting bit — **the system is built so that
almost never happens**: after a veto it is explicitly told to shrink the idea to something
honest and try again, and it only gives up on a second refusal. So "wait for a natural
rejection" was never going to deliver it. I know how to force it and had it set up.

**Which is where I have to flag something bigger than this job.** At 14:51 today the Anthropic
account hit its spending limit. Every agent on the estate that talks to the model is now failing
— I watched my run and another session's review council die within the same minute on the same
error. **The message says access returns on 1 September**, which I assume is not what you want,
so it will need the cap raising on the billing side. Nothing is broken and nothing is lost;
everything queued will simply fail until it's lifted. It also means the last small piece of
this job can't be finished today.

---

**2026-08-10 — the rebuild is decided, and the first framework fault is already fixed
(in code, not yet live).**

You asked for the site to be rebuilt entirely through the framework so we can see what
the framework itself gets wrong. Before touching anything I checked what a rebuild would
actually do, and found the first fault without running it: the framework cannot express
this site's addresses. Its planner insists every tool lives at /tools/name/index.html,
while this site serves /tools/name.html — and rather than refusing, it would quietly
move twenty-four of the twenty-six pages to new addresses and leave the old ones stale.
That's now written up as bug 241, and the code half of the fix is committed: the
framework can now be told "this site uses flat addresses", switched off by default so no
other site changes behaviour. It's gone to the review council; the verdict wasn't back
when this session ended.

Nothing on the live site has been touched. The rebuild itself — backups, releasing the
locks as you decided, re-submitting the site through the pipeline, and then the audit of
what comes out — is laid out step by step in the 2026-08-10 handoff, with the wiring of
the address fix as the first job. The two wrong claims (the footer's "shows its own
arithmetic" and the guide's "month-by-month breakdown") are still live; they die with
the rebuild, and if the rebuild drags we should just correct the two lines in the
meantime — your call.

**Monday evening, 10 August.** The spending cap was lifted some time around six o'clock — the
system started working again on its own, and I've checked rather than assumed: everything was
running normally from 18:00 onwards. A new build also went out while I was away. I re-checked
all three of our fixes against it before touching anything, and all three are still in place.
That's now the third time a database-only fix has survived a rebuild, which I think we can
stop treating as a question.

**And the last gap is closed.** This morning I could show that a plan the council *vetoes*
never gets saved. What I couldn't show was the other half: a run that gets vetoed and then
gives up entirely, leaving nothing behind. I explained why that's hard — the system is built to
have a second go rather than give up — so I forced it by telling the council it only had one
round to decide in, then put the setting back straight afterwards.

It went exactly as it should. The system drafted a full plan — ten and a half thousand
characters of it — the review panel vetoed it on the grounds that it needs a server we don't
have, and the run stopped there. **Nothing was written.** The site's existing plan was left
untouched, down to its last-modified timestamp.

The detail that makes this proof rather than a coincidence is that **a real plan existed at the
moment it gave up**. That was the flaw in this afternoon's attempt, which looked like a pass:
that run had died before writing anything at all, so of course nothing was saved. This one had
a finished plan sitting in hand, ready to save, in the exact place the saving step reads from —
and it still didn't save it. Under the old behaviour that vetoed plan would have become the
site's official plan, and the run would then have ended in failure with a rejected document as
the record.

So **the whole of this fault is now fixed and proved on both halves**. I've tidied up after
myself: the imaginary experience I invented to bait the veto is deleted, and the one plan it
produced is marked as a test artefact and is no longer official — I kept it rather than
deleting it because it happens to be the evidence for the first half of the proof.

## 2026-08-11 evening — the rebuild is unblocked; everything before the button is done

The URL problem you asked us to fix first is fixed and running. The framework code that
respects this site's flat web addresses went live this afternoon (it rode another thread's
release, which saved us a deployment), and I've switched it on for this site and checked
the switch took. So the thing that made a rebuild dangerous — the framework silently moving
24 of the 26 page addresses — can no longer happen here.

The reviewers looked at the work twice. First pass: approved, with the fair note that the
code alone changes nothing until it's wired in and switched on — which is what I then did.
Second pass caught something real: if this site were ever RE-imported through the adoption
route, the import would silently wipe the switch I'd just set, and the bug would quietly
come back. I've fixed that too (an import now keeps settings it didn't create), it's
committed, and the third review round is in progress. That fix isn't needed for THIS
rebuild — it protects the future.

Everything on the runway before the rebuild is done and checked: the sixteen open
maintenance jobs aimed at the calculators are parked (seventeen actually — one more was a
job that creates such jobs); four separate backups exist, each one verified, including a
snapshot that provably captured all twenty content locks; and the exact lock list is
written down so it can be re-applied if we ever want it.

Two things wait on you:

1. **The mission brief** — the framework needs the site described to it as if this were any
   fresh commission, and that description steers everything it writes. My draft is in
   `MISSION_DRAFT_2026-08-11_for_owner_review.md`. The two sentences worth your attention:
   the one that tells it to keep the existing pages and addresses, and the one that forbids
   claiming a capability a calculator doesn't have — that second one exists because of the
   two false claims that started all this.
2. **The locks.** You said release everything, and I will — but I'm releasing them at the
   moment we press the button, not before. There are twenty-one other open jobs (page
   rewrites and re-renders) that would be free to alter the untouched baseline in the gap
   between unlocking and rebuilding, and that would muddy the before/after comparison the
   whole exercise is for.

Say the word on the brief and I fire the rebuild.

## 2026-08-12 — the four decisions, written down

Re-checked this morning, all still true: the flat-URL setting is on, all twenty locks are
still held, the seventeen calculator jobs are still parked, and both false claims are still
serving (I re-fetched the pages to be sure — the homepage still has no table of any kind,
so the "month-by-month breakdown" claim is still false today).

**1. The two false claims have now been live for three days.** The rebuild replaces them,
but only when it runs. If you're approving the brief today, leave them. If the rebuild is
going to sit any longer, I'd cut the false half-sentence from each — a deletion, not a
rewrite, so nothing hand-authored gets added to a site whose whole point is that the
framework writes it.

**2. Does the brief tell the framework to keep the existing pages and addresses?** My draft
says yes. That's the bigger of the two questions, because it decides what the exercise can
teach. Keeping them means we learn how the framework WRITES; letting it choose freely means
we also learn how it PLANS, at the cost of possible duplicate pages (a page it names
differently doesn't move — it appears twice, which is a bug another thread is mid-fix on).
I recommend keeping them, because both faults we're chasing are writing faults.

**3. Something I've measured since you said "release everything".** The planner physically
cannot see the calculators — it only reads section and element components, and calculators
are a third kind. So unlocking them buys almost no extra purity: the planner was never
going to rewrite them. What unlocking DOES do is let the page-save wipe them, and we know
that's not theoretical, because the locks have already turned away ten such writes, five of
them aimed straight at calculator slots. The arithmetic is recoverable either way, so this
costs rework rather than the asset. Your call stands unless you want to narrow it; I'd keep
the twelve calculator locks for the first pass and release the other eight.

**4. One I'll just do unless you stop me:** park the twenty-one other queued jobs (page
rewrites and re-renders) alongside the seventeen, so that anything different after the
rebuild was done by the rebuild and not by the maintenance queue.

## 2026-08-12 later — all four done except one, and the last one turned out to be two jobs

Parked the maintenance queue (and a correction: I told you twenty-one jobs, it was
actually forty-five — twenty-one was just the two kinds I'd named. All parked, all
reversible).

Your instruction to fix the planner turned out to be two separate changes, and finding
that out was the useful part of the day.

The first one is done and reviewed: when a rebuild composes a page, the code that
protects a locked calculator was only recognising it **by the name of its slot**. Our
calculators sit in slots called things like "tool-2", while a fresh plan describes pages
in proper component names — so the protection would have missed them. It now recognises
them by which component they actually are, which is how the neighbouring piece of code
has always worked. Without this, fixing the second half alone would have put **two**
calculators on every page: the new one in place, the protected original pushed to the
bottom.

I found this partly because another thread had already worked it out on the 10th and
left a test explaining it — their conclusion was sharper than mine, and it says plainly
that seeding a fresh plan is the dangerous act, which is exactly what we're about to do.
I've written back to them.

The second one — making the planner actually see the calculators — I have designed but
deliberately **not** done, because three measurements say it isn't a one-line change.
The component library is shared across the whole platform with no per-site column, so
the obvious edit would offer every site all eighty-one calculators, including other
sites'. Twenty-one sites already use tool components, so even a careful version changes
behaviour beyond us. And the step fails hard if its site lookup ever comes back empty,
which on a single shared setting means no site anywhere can plan. That is a change to
make carefully with a review, not one to slip in before pressing a button.

**So there's one decision left, and it's a timing one.** If we fire today, the twelve
calculators survive — the locks do hold — but each lands at the *bottom* of its page
until I put it back. If we wait a day or so for the second half, the rebuild produces
the site properly first time. I'd wait. The reason we kept those locks was to protect
the calculators, and firing into a known twelve-page mess to save a day spends exactly
what we were protecting. The cost of waiting is that the two false claims stay up a bit
longer — so if we do wait, say the word and I'll cut those two half-sentences now.

## 2026-08-12 evening — both false claims are gone from the live site

Checked on all twenty-six pages just now: the footer sentence claiming every calculator
shows its own arithmetic is gone, and the guide's promise of a month-by-month breakdown
is gone. I also checked the opposite direction on every page — the rest of the footer is
still there, and no page came back short or broken — because "I found nothing" is worth
very little unless you can show the same check finding something.

I did it as deletions rather than rewrites. The footer now reads "Independent UK
borrowing tools." and stops; the guide now says use the calculator, and that like any
calculator it only knows the figures you give it. Nothing new was written, which matters
on a site whose whole point is that the framework writes the words.

Two things I want to flag because they cost time and one of them nearly cost more.

First, I fixed the sentences in two places each — the finished page and the underlying
content record. The claim was sitting in both, and had I only fixed the visible one, the
next time the framework rebuilt that page it would have written the false sentence
straight back.

Second, I got something wrong and caught it late. After triggering the rebuild of the
pages I checked whether it had started, saw nothing, and concluded the instruction had
been dropped. It hadn't — it had run perfectly, three quarters of an hour before I
looked. My check asked "did anything happen in the last fifteen minutes", which was the
wrong question after the session had been interrupted; the answer it gave was true and
useless. I was one step from re-running a whole-site redeploy that had already succeeded.
It's written up so the next person asks by name rather than by clock.

One genuine unknown, recorded as unknown: an earlier batch of twenty-six queue entries I
created by hand disappeared from the database outright. I can't find anything in the code
that deletes those, and everything else of mine is untouched. They were the wrong shape
and wouldn't have run anyway, so nothing was lost — but I'd rather write down that I
don't know than invent a reason.

The rebuild itself is still waiting on the planner work, as agreed. Nothing about today
changes that.

## 2026-08-13 — everything shipped is now running; one measurement left before the button

Today's build carries both protection fixes — I checked the running programs themselves,
not just the version number, both times it rolled. The reviewers approved the lock fix
with notes but no objections of substance. And the deeper worry from my review turned
out fine when measured: the protected calculators and the components a fresh plan would
name are literally the same records, so the pairing fix will catch them.

The one thing I refused to do today is the last one: pointing the planner at the
calculators requires telling it where to find the site's identity, and the place I was
told to look — records of past planner runs — turns out to be empty, all-time, even
though plans are visibly being written. I know the likeliest answer, but a wrong guess
here stops every site on the platform from planning, so I'm not guessing. The next
session catches one live planner run, reads the answer off it, and then it's a small
reviewed change followed by the rebuild itself.

Handed over cleanly: HANDOFF_2026-08-13_planner_half_continue_here.md has the whole
state, what's proven, and the exact next steps.

2026-08-14 (evening) — The planner change is in. The one thing blocking the rebuild
was teaching the site-planner to see a site's own calculators, and the one thing
blocking THAT was proving which piece of run-state reliably carries the site's id —
because getting it wrong wouldn't just break this site, it would stop every site on
the platform from planning. That proof is now done three ways: the planner's own
workflow creates the value two steps before it's needed (in code that cannot skip
it), the step that writes finished plans has used the exact same value forever (so
every plan ever written is a successful test of it), and we ran the new query
against the live database and confirmed sites without the opt-in flag get
byte-for-byte what they got before. The change is applied and live, the flag is set
on loancalculator only, and the whole thing went to the review council (submission
508fe8eb). The mystery of the planner runs that seemed to leave no trace also fell:
the runs table only keeps about two days of history, and the planner runs so rarely
(three times ever) that its records are always gone before anyone looks.

Two things happened to the site while we weren't looking, both harmless but worth
knowing: the platform's routine maintenance created one new page (a loan FAQs
guide — it's live and looks right), and a fleet-wide re-render wave (expected,
one-off, caused by a fix that rolled yesterday) refreshed about a third of the
pages. The calculators, the cut claims, and the locked footer all came through
intact — the locks did their job and blocked the three attempts to overwrite the
protected chrome. One embarrassment on our side, recorded properly: I briefly
believed the wave had stripped the calculators off the live site, because I was
checking pages at addresses I'd guessed instead of the real ones — the "pages" I
was reading were 404 errors. The database said otherwise, which is what sent me
back to look. Next: council verdict, then a one-site trial replan, then the
rebuild.

2026-08-14 (night) — The review board approved the planner change on the second round
(the first round asked, fairly, that I attach proof for things I'd merely stated; the
proof existed and now it's attached). Then we ran the trial: one planner run on the
live site with the new behaviour switched on. The good news is complete on the
mechanics — the planner saw all eleven calculator components in its menu, looked up
the right site, nothing broke anywhere on the platform, the locked calculators
weren't touched, and every real page still serves perfectly. Two surprises, though,
both caught and boxed in. First, the planner took it upon itself to invent two pages
we never asked for (an "about" page and a guides index) and queued up work to build
them plus generate imagery; I've parked all of that and shelved the two empty page
entries — nothing reached the live site. Second, and more useful: a plain replan
turns out not to propose page layouts for pages that already exist — so the question
we most wanted answered ("will the planner keep the calculators in place when it
rebuilds a page?") simply never came up. It can only be answered by the rebuild
itself, or by explicitly asking for specific pages to be recomposed. The locks and
the pairing fix remain the safety net either way.

So: the planner work you asked for is done, reviewed, and proven as far as a trial
can prove it. Before the rebuild fires there are two things I'd like your view on:
how we want the 26 existing pages regenerated (the explicit per-page route is now
clearly the honest one), and whether you're comfortable that the mission brief's
"keep the pages" instruction is strong enough given that a bare replan invented two
pages on its own. Everything is written up for the next session either way.

2026-08-14 (late night) — You answered all three questions: regenerate all 26 pages
by naming them explicitly, trust the "keep the pages" instruction backed by the
immediate after-the-planner check, and yes to both of the pages the planner invented
— the about page and the guides index are wanted, so they've been restored rather
than left shelved.

Before firing I checked the machinery rather than assuming, and four things came out
of that which change the shape of tonight's launch, though not its substance. First,
the ordinary submission route has nowhere to carry the "recompose these 26 pages"
list — so the rebuild happens in two stages: the full rebuild first (new brief, new
plan, the two new pages built), then a second, targeted planner run that names all
26 existing pages for redesign. Second, the platform deliberately refuses to
automatically rebuild pages that carry working tools — for eleven calculator pages
the redesign will be QUEUED FOR YOUR REVIEW rather than executed, which is the
platform protecting the calculators, not a fault; the other fifteen or so pages
rebuild on their own. Third, the two restored pages can only be built through the
work tickets the trial run created and we parked — so those get un-parked at the
right moment (after the new plan lands, so they're built to the new brief, not the
old one). Fourth, a page that exists but has never been served is exactly what the
planner is happy to fill in, so restoring the two rows is all the preparation they
need. One more thing found on the way: the platform was updated fleet-wide this
evening (after the handoff was written), and the site came through it serving all
27 pages cleanly — re-checked before touching anything.

2026-08-15 (morning) — The rebuild is FIRED and moving. The two wanted pages are
restored, the eight non-calculator locks are off (the twelve calculator locks stay),
and the submission went through cleanly — the research stage picked it up within
half a minute and finished within a few. Two mistakes of mine, both caught and fixed
within minutes, both written up in the fleet's wrong-calls log: I told the brief the
homepage carries the "credit roadmap" tool when it actually carries the standard
repayment calculator (fixed in the brief and in the stored copy before anything
downstream read it), and I timestamped the launch from this machine's clock, which
turns out to run about ten hours behind the cluster's — the database's own clock is
the one that counts. Watching now for the new site plan to land; the moment it does,
the two new pages' build tickets get un-parked and the invention check runs.

2026-08-15 (mid-morning) — A pause, understood and resolved. The research and
strategy stages finished quickly, then the pipeline went quiet: it turns out a
safety gate added last week deliberately stops a live site's strategy refresh from
triggering a full re-plan — on a site that's serving, "someone refreshed the
strategy" must not mean "rebuild everything". Sensible in general; our case is the
one exception, because rebuilding everything is precisely what we're here to do. The
fix was to file the one work ticket the gate had withheld, exactly the way the
webdesign.uk lane did on the same seam last week. The chain is moving again —
briefing next, then the plan. One consequence worth remembering: resubmitting a live
site will ALWAYS need this extra manual step; that's now written into the lane notes.

2026-08-15 (late morning) — A second, smaller stall, also understood: the ticket I
filed by hand sat unnoticed for nearly two hours because hand-filed tickets enter
the queue one state earlier than the dispatcher looks — the platform's own workflows
skip that state when they file tickets, and the sweeper that normally promotes
things is running days behind, fleet-wide. Nudged ours into the visible state; rule
recorded: file by hand = file it already-promoted. Meanwhile your new platform build
(v1.0.1301) rolled through cleanly — nothing of ours was in flight when it did. A
full handoff for the next session is cut and committed; the fire's state, the
checkpoint script, the phase-two script, and the four decisions that will be yours
to make are all in it.

2026-08-15 (midday) — The plan landed, and with three pieces of genuinely good
news. First: the planner invented NOTHING this time — the "keep the pages"
instruction held, which was the thing we couldn't test in the trial. Second: the
plan is exactly the twenty-nine pages we asked for, and the about page has a real
layout proposed. Third, and the big one: the rebuild didn't need the planned second
stage — the pipeline queued the whole regeneration itself: fifteen ordinary pages
are rebuilding right now, and all eleven calculator pages were routed straight to
your review queue, which is the platform refusing to touch working tools without a
human — exactly what we'd want. The twelve calculator locks are confirmed intact.
The homepage is the one to watch: it rebuilds automatically around its locked
calculator, which is the first real test of the protection work this whole effort
was built on. One decision has grown: the design tickets you parked on the 12th are
now the only thing stopping the site's look (chrome, styles, favicon, imagery) from
being regenerated too — un-parking them is your call, since parking them was too.
A watcher is on the rebuild; next update when it settles or if anything fails.

2026-08-15 (afternoon) — Fourteen of the fifteen pages rebuilt cleanly and are
serving — checked at the live site, not just the tickets. The fifteenth, the
homepage, stopped itself: it turns out pages of its kind only rebuild from an
explicit layout in the plan, and the plan didn't carry one, so it queued itself for
review and changed nothing — the calculator never came under threat. That same fact
un-shelved the second launch stage: it exists precisely to put layouts in the plan
for the homepage and the eleven calculator pages, so I've fired it for just those
twelve (not the fourteen already rebuilt — no point churning them). It writes plans
only; no live page changes until the review tickets are worked, which remains your
call. One embarrassment, logged in the wrong-calls file: I fell into the exact trap
I'd written up two hours earlier — the two new pages' build tickets sat invisible
for three hours because my own script filed them one state too early; they're
building now.

2026-08-15 (mid-afternoon) — The question this whole effort was built to answer
now has a measured answer, and it's the uncomfortable one: given full freedom and
a menu that demonstrably contained every calculator, the planner designed eleven
calculator pages and the homepage WITHOUT their calculators. Nothing live was
touched — the review gate and the twelve locks did exactly their job, which is
why we built them — but the redesigns it proposed must not be applied as they
stand, and I've marked the review tickets accordingly. The "why does the planner
ignore the tools" question is now with the platform's own diagnosis loop; its
verdict will say whether the fix is in the planner's instructions or somewhere
deeper. Separately, all the ordinary pages are converging on their final form (a
second, harmless build pass is running), and the two new pages are building. The
site has served every real page cleanly throughout.

2026-08-15 (early evening) — The diagnosis came back and it exonerates the planner:
I had the mechanism backwards, and I've corrected the record everywhere I'd written
it. The planner DID put every calculator exactly where it belongs — its raw output
proposes the repayment calculator as the homepage's second section, and the right
tool on each calculator page. What eats them is a validation step downstream: its
list of acceptable section names was never taught about tool components when the
menu was widened last week, so it silently deletes every tool the planner places,
believing it's cleaning up an invalid name. Classic case of upgrading one half of a
handshake. Filed as bug 282 with the fix spelled out (teach the validator using the
menu's own rule, as one shared piece of code so they can never drift apart again).
The redesign tickets stay held until that fix ships and a fresh plan proves the
calculators land; everything else about the rebuild stands. The about page is built;
the guides index needs the same fix path as the homepage.

---

## 2026-08-17 (midday) — the calculators are back in the plan: none of twelve, to eleven of eleven

Short version: the thing that was blocking this site is fixed, and I have proved it
on the actual site rather than in principle. The calculators are not yet back on the
rebuilt pages, and the reason for that is a fleet-wide billing gate, not anything
wrong with this site.

**What was wrong, in plain terms.** When the site was rebuilt on 15 August, the
planner correctly decided which calculator belonged on each page — and then a
checking step further down the line threw every one of those decisions away. It was
checking proposed page contents against a list that simply did not include
calculators, so each calculator silently failed the check and was dropped before
anything was saved. The result was a set of tool pages whose plans described
everything except the tool.

**It is fixed, and the fix is running.** Another thread fixed it yesterday: the
checking step now accepts exactly what the planner was offered, rather than
re-deciding for itself. I confirmed the running system genuinely carries that fix —
not by trusting a version number, but by reading the commit stamped into the running
program and asking git whether the fix was in it, with a control that could have
failed.

**Then I re-ran the planning step and counted.** Before: none of the twelve pages had
its calculator in the plan. After: **every one of the eleven pages that owns a
calculator now has exactly its own calculator, in the right place.** The twelfth page
(Credit Roadmap) has no calculator of its own, and the planner gave it a copy of the
Credit Health Check one — that is a content decision for you, not a fault, and I have
left it alone. The twelve protected copies of the calculators were untouched
throughout, and I checked that rather than assuming it.

**I also finished the outstanding check on the calculators' arithmetic.** The tool
that records what each calculator computes reported "11 of 11 diverged", which sounds
alarming and is not. Breaking it down: of 1,340 values compared, **not one changed**.
Every difference is the new page furniture — the old hand-built navigation menu has
gone and a new questions-and-answers block has arrived. The sums are identical. I
re-recorded the baseline so future runs compare against today's pages.

**One thing I should flag as a genuine loss.** The old site had a "Tools" menu in the
header listing nine calculators. The rebuilt header has only Home and About. The
calculators are still reachable — each page links to eight of the others in its body
— but they are no longer in the menu. This is the framework behaving as designed: it
deliberately keeps individual tool pages out of the top menu and expects a single
parent "Tools" listing page to represent them. **This site has no such page** (it has
one for Guides). So the question for you is whether we should create a Tools listing
page, the same shape as the Guides one. I have not done it.

Related, and smaller: the Guides entry is missing from the menu too, but for a
different and more mundane reason — the Guides page itself has never been built and
currently returns "not found", which is the one outstanding page on the site (28 of
29 serve correctly). The menu data already contains the Guides link, so building the
page should bring it back.

**Why the pages have not actually been rebuilt yet.** The rebuild work was queued
correctly — fifteen page jobs and some image jobs — and then nothing happened,
because the whole fleet's job queue is currently shut. The cause is the Anthropic
account's spend limit: it was hit briefly at about 11:08 this morning, and a health
record flipped to "unavailable". Everything that claims a job checks that record
first, so no job anywhere in the estate can be claimed while it says unavailable —
and that record is only re-checked once an hour. Meanwhile the API itself recovered
within minutes and is working fine right now. So the estate is idle because of a
stale flag, not because it is broken. It should clear itself at 12:09:53, and I am
watching for that.

To be clear about what is and is not affected: **the eleven-of-eleven result above
does not depend on any of that.** The plan is written and verified. The queue only
governs when the pages get rebuilt from it.

This is already known — three other threads hit the same wall this morning and it is
written up as bug 243, with the spend itself as bug 244. I spent about twenty minutes
re-diagnosing it before finding their write-ups, which is my own fault: I should have
searched first. Worth knowing that this is the second time today that the fleet's
own "is everything healthy" check said yes while nothing could actually run.

---

## 2026-08-17 (late afternoon) — the calculators came out right; I also put fourteen duplicate pages on the site

Two things happened and they need separating, because one is the win we were after and
the other is a mess I made getting there.

**The win, and it is the whole point of the last three days.** The homepage rebuilt
this afternoon and the loan repayment calculator is now sitting where it belongs — as
the second thing on the page, composed into the layout by the framework, rather than
bolted on at the end as it was before. Every one of the eleven calculators is in the
plan on its own page. All twelve protected copies survived untouched. And the tool that
records what each calculator computes now passes cleanly on all eleven: "all 11 tools
reproduce their golden values exactly". So the site is genuinely being built by the
framework now, calculators included, which is what we set out to prove.

**The mess.** The same planning run that fixed the calculators also quietly rewrote the
guides section. Our fourteen guides live at /guides/… and are typed as guides. The new
plan replaced all fourteen with blog posts at /blog/… — same topics, same slugs, new
URLs — and dropped the real guides from the plan entirely. It then queued fourteen
builds for these new pages.

I misread those fourteen queued jobs as harmless. The previous session had described
its own batch of fifteen jobs as harmless re-stamping of existing pages, and I carried
that sentence across to a different run without checking it. The job names actually
said "can-i-overpay", not "guide-can-i-overpay" — the missing word was in output I
pasted into my own notes and read twice.

I did catch it, but late. There is a check written for exactly this — a fingerprint of
the site's page list, meant to be run the moment a new plan lands — and when I ran it,
about thirty-five minutes after the fire, it told me straight away that the page list
had changed. By then the fleet-wide queue I told you about this morning had re-opened,
on the hourly timer whose exact firing time I had calculated myself, and the builds had
gone. **All fourteen duplicates are now live.** The real guides are untouched and still
serving. I tried to stop it about two minutes after realising, and the write was
refused by the permission system; by the time it reached you the window had shut.

**What I need from you, and it is a question about the site rather than a fix.** The
plan currently says this site's articles live at /blog/. Nobody chose that. So: do the
guides stay at /guides/ and we retract the fourteen new pages, or does the site
actually move to /blog/ and we retire the guides with redirects? I have not touched
either, because the answer changes what gets cleaned up. Worth knowing that retracting
published pages has no clean path in the framework today — it is the same problem
already flagged as bugs 80 and 81, a wrongly-typed page that is live with no way to
withdraw it — so this may need a decision from you about method as well as intent.

The site's own immune system spotted the duplication without being asked, reporting
"14 blog posts deployed but not linked from blog listing page", and responded by
queueing a re-render of all forty-three pages. That is why there is a lot of churn in
the queue right now. Most of it is failing on a git error when it tries to publish,
which is not losing any work — the pages themselves are fine — but it is noise.

**One more thing, and it matters for anything you deploy today.** The fresh chassis
build has not actually reached the cluster. A new image was built at 15:30 from a much
later commit, but it was pushed under the *same* version tag as the one already
running, so the machines kept the copy they had cached. The running program is the one
from this morning — two hundred commits behind what you built. The version number
looks right, which is exactly why this trap is in our notes. It needs the tag bumping
and a proper fleet release, which is your command to run. Nothing in this site's work
depends on it; the calculator fix was already live.

Where the site stands: forty-three pages active, forty-two serving. The Guides index is
still the one page that will not build — it has no sections composed for it — and it is
still the only genuine 404.

### Correction, same day, about an hour later — the calculators ARE in the navigation

I told you above that the rebuilt site had lost the calculators from its navigation and
that you might need a new "Tools" listing page. **The first part is wrong and I want to
correct it before you act on it.**

The framework did put all eleven calculators into the site's navigation — in the
**footer**, not the header. That is what it is designed to do with a page it keeps out
of the top menu: bar it from the main menu, keep it in the footer. I have now checked
this at three levels and they agree: the navigation data has eleven footer entries, the
regenerated footer itself contains all eleven links, and a guide page that republished
at 16:21 this afternoon serves a footer listing every calculator.

**Why I got it wrong.** I looked at the footer on pages that had been published at
13:44. The site's shared header and footer were regenerated at 13:47 — three minutes
afterwards. A page keeps whatever header and footer it had when it was last published,
so those pages simply had not caught up yet. The re-renders that would have brought
them up to date are the ones failing on the git error, so the out-of-date version was
sitting there stable enough to look like the finished state.

So the real change is narrower than I said: the old site had a **Tools dropdown in the
header**; the new one has **all eleven calculators in the footer**. Nothing is missing
from the navigation. Whether a footer list is good enough, or whether you want them
back in the header — which needs that Tools listing page — is a genuine choice, but it
is a preference now rather than a repair.

The Guides entry is a different matter and still outstanding: it is in the navigation
data but the Guides page itself still will not build, so there is nothing to link to.

---

## 2026-08-18 — you chose /guides/, and it turns out the framework already knew how

You said you'd prefer /guides/ but would take whatever the most natural fix for the code
was. Those turned out to be the same thing, which is the best possible answer.

**What I found.** The framework already has a switch for precisely this problem: keep a
page where it is actually being served rather than re-deriving its address from what kind
of page the planner decided it was. It was written back on 10 August — while planning this
very site's rebuild — by someone who spotted the risk before it happened. The code comment
describing the danger reads, almost word for word, like a description of what went wrong
yesterday: it warns that the planner will re-derive a blog post's address under /blog/ and
so move a live page that is serving from /guides/.

It is deliberately off by default, because switching it on changes real addresses on a live
site, so each site has to opt in. **This site had never opted in.** I have now switched it
on and checked it took: three flags set, and the two things that had to survive the change —
the site's flat-URL setting and its 27-page adoption record — both intact.

**And there's a reason the planner picked "blog post" that isn't its fault.** The framework
can only express a guide's address as /guides/something/index.html. It has no way at all to
say /guides/something.html, which is the shape this site actually uses. The only page kinds
that produce the flat shape are blog posts and entity pages. So when the planner tried to
describe our guides, the closest thing it could say was "blog post" — and that puts them
under /blog/. It reached for the only expressible option.

**What this does and does not do.** It stops the *next* plan moving these pages. It does not
undo yesterday: the fourteen duplicate pages are still published. Removing them still needs
the publishing path to work, and that is still broken across the whole estate, so it is
waiting on infrastructure rather than on you. The good news, which I checked, is that they
are almost invisible: not in the sitemap, not in any menu, and there is no blog index page
linking them, so the only way to reach one is to type its address.

**One more thing I found and have not fixed**, because it needs a decision beyond this site.
There is a guard in the planner written for exactly our situation — it spots when the
planner re-proposes a page that already exists under a different name and throws the
duplicate away. Its own example in the code is the same shape as ours. It cannot ever fire
on a site that already has a plan, because of how its list is built. So for every
established site, a guard that looks like protection is switched off by construction. I've
written that up; it deserves its own bug.

**Also corrected today, my own error.** Yesterday I told you the recompose had produced "no
no-op pages" as a good sign. I checked how that signal is recorded and it has never recorded
anything, anywhere, in the system's whole history — so my reassurance was worth nothing. The
sign wasn't wrong, it just couldn't have been anything else. I'd applied exactly that kind of
scepticism to another number the same morning and then failed to apply it to my own.
