# Where we are — loanandmortgagecalculator.co.uk

The owner's running log. Plain prose, append-only, newest at the bottom.

---

**2026-07-31, morning.** Started by finding out what we actually have. You asked me
to put mortgagecalculator.co.uk onto loanandmortgagecalculator.co.uk, so the first
job was to be sure which copy of the mortgage site is the real one.

There are two copies in your domains folder, and they are not the same. The live
site is served from the `gemini/02` subfolder, not from the top level. I checked
this properly rather than by looking at dates — I fetched all 23 live pages and
compared them byte for byte against both copies. All 23 match `gemini/02` exactly.
Worth saying plainly: if I had taken the obvious top-level folder, I would have
built the new site from the wrong material and every check afterwards would have
told me it was fine.

**Then I tried to prove the calculators work, and the testing tool lied to me.**
We have a tool that loads each calculator in a real browser and checks whether it
does what it promises. I pointed it at the mortgage site and it reported that all
fourteen interactive pages were broken — "nothing a visitor can touch". They are
not broken; they all work.

What saved me was the number, not carefulness. Fourteen out of fourteen identical
failures is not a description of a website, it is a description of a broken
instrument. A site with fourteen genuinely dead calculators would fail in fourteen
different ways. So I checked the tool instead of believing it, and the tool was
looking for the calculators inside an HTML element that this site doesn't use. It
had been checking one of our other sites for months, where every page does use that
element, so nobody had noticed.

I fixed it, and I proved the fix two ways: the previously-invisible calculators now
pass, and I re-ran the old version and the new version side by side over four pages
of the site it was originally built for, comparing nine results each. Nothing moved.
So I have made it see more without changing any verdict it was already giving.

Then the baseline came out clean: thirteen of thirteen mortgage calculators working.

**Something you may want to know about the old site.** It has had four broken things
live since it was built. Six of the nine guides have a "Home" link that goes to a
page that doesn't exist. The homepage links to a guide that doesn't exist either —
the file is called something slightly different. Two guides have nothing linking to
them at all, so nobody can find them. And there is no sitemap, with a leftover
placeholder comment in the robots file where the sitemap line should be. I have not
touched the old site — that wasn't the brief — but say the word and it is an hour's
work.

**On your point about the two sites evolving in different directions.** That changed
the design, and for the better. I had been planning to copy the site across and
write better guides. What you actually want is two sites with different audiences,
where the difference is recorded in the framework rather than being something I
maintain by hand.

So the new site is positioned as the **whole-borrowing-picture** site, and
mortgagecalculator.co.uk stays the narrow mortgage authority. The new one answers
the questions that span both subjects — how your car finance reduces what a
mortgage lender will offer, whether to consolidate debt into a remortgage, whether
the next thousand pounds should go on the deposit or clear a loan. That is a real
difference and not a paraphrase: no single-subject site can answer those questions,
which is exactly why they are worth owning.

You chose the combined site, so it has both halves: twelve mortgage calculators and
eleven loan ones, twenty-three in total, on one domain with one design.

**What is built.** All twenty-three calculators are ported and every one is verified
working in a real browser. The important thing about how I did it: the calculators'
arithmetic is copied byte for byte, and the build **refuses to run** if any of it
changed. I only rewrote the wrapper — the titles, the navigation, the footer, the
links. And I deliberately broke the check to make sure it actually catches a change,
because a safety check that has never gone off isn't a safety check.

Thirteen guides are entirely new writing. None is adapted from either site. One
editorial decision worth flagging: I have avoided quoting any current interest rate
or tax band anywhere. The old sites hard-code "3.75% base rate" and a March date,
and copy like that is wrong within weeks with nothing to tell the reader. The new
guides explain how things work, which doesn't go stale, and send people to the
calculator for numbers.

**Three problems I hit while porting, and they are instructive.**

One was mine. My build rebuilds the page header from scratch, and three of the
mortgage calculators keep their shared maths file linked in the header rather than
in the body. So the build quietly threw that link away. One of the three then failed
loudly, which is how I found it. The other two carried on looking perfectly fine
while missing the same file — those are the ones that would have shipped. I fixed it
properly by adding a second check that fails the build if anything a calculator
depends on goes missing, rather than fixing three files by hand.

The second was not mine, and it is a nice example of a thing that is genuinely hard
to see. One of the loan tools is a five-step questionnaire. Its code moves a marker
from one step to the next and relies on the stylesheet to hide the steps you are not
on. **That stylesheet rule was never written** — not on the loan site, not anywhere.
So all five steps have been showing at once on the live site, and the tool has been
visibly broken, and no amount of reading the code would ever find it because the
code is correct. The missing half is a class name that appears nowhere except as
text inside one instruction. I have written the missing rules and it works now. It
turned out to be one of thirty-six styles those tools use that were never defined.

The third was the testing tool again, twice more. It called a working tool dead
because it had no way to tick a checkbox, and it called the questionnaire dead even
after I had fixed it, because it only knew how to look for changes in a fixed list
of places and the questionnaire responds by showing a different section. Both fixed,
both re-checked for side effects, nothing moved. Three of the four failures on this
site were the instrument rather than the site, which is a ratio worth remembering.

**One tool I did not bring over,** and I want to be explicit rather than have it
just be absent: the loan site has a "6-month credit roadmap" filed with the
calculators, and it is not a calculator. It is under two kilobytes with no controls
and no code — a short article in the wrong folder. The subject is covered properly,
and in much more depth, by one of the new guides.

**Where this stops, and it needs you.** The site is built, checked, and the files
are already uploaded to our storage. It is not visible yet, because the domain is
still parked at whoever you registered it with — it currently bounces to a
registrar holding page. Two things need doing in the Cloudflare dashboard and only
you can do them: add the domain as a zone and repoint the nameservers, then add a
Workers Route for it pointing at the same little program every other site uses.
There are no Cloudflare credentials on this machine at all, so this is not something
I can script my way around.

The good news is that the storage side is already correct and won't need redoing —
the way our setup works, files are filed under the domain name, so the moment that
route exists the site simply appears. I verified all fifty-two files uploaded.

**After that, two things remain.** Adopting it into the framework so it is managed,
which I will do with the calculators locked so nothing can ever regenerate working
arithmetic, and the guides left open so the platform can keep improving them. And
recording the two sites' different audiences as framework settings, which is the
part that makes your "evolve in different directions" instruction stick without
anyone having to remember it.

**One caveat I want on the record.** I have verified that the ported calculators
respond correctly and that their code is byte-identical to the originals. I have
**not** verified that each one produces the same answer for the same input as its
original did. Byte-identical code with its dependencies present is strong evidence,
but it is not the same thing as checking the output, and I would rather say so than
let "verified" cover more than it earned. Per-calculator acceptance checks are the
obvious next job and there is already a pattern in the repo for them.

---

**2026-07-31, evening. It is live — and going live found three things I had got wrong.**

You put the domain into Cloudflare and pointed it at the storage, and the site came up
straight away. I checked all fifty-two files by fetching them and comparing them
byte-for-byte against what I had built: fifty-one identical. The fifty-second is the
`robots.txt`, and that one is fine — Cloudflare adds its own block to the top of every
site's robots file, and it does the same on your other domains. My own instructions are
still there underneath.

**Then I checked the live site properly, and it was not as clean as I told you it was.**
Three things, all mine, and I would rather set them out plainly than have you find them.

**One: the three main section links were broken on every single page.** The links to
"Mortgage tools", "Loan tools" and "Guides" pointed at `/loans/` and so on, and every one
of those returned a Not Found. Worse than that, those three addresses were also in the
sitemap I hand to Google, and three pages told Google that their own official address was
one of the broken ones.

The reason is a genuine quirk of how your sites are hosted. There is no real web server —
files are pulled straight out of storage by name. A normal web server, asked for
`/loans/`, knows to look for `/loans/index.html`. Storage does not: it looks for a file
literally called `loans/`, and there isn't one. Every site you own behaves this way; mine
was the only one that actually *linked* like that.

**And here is the part I am least pleased about.** My own link checker caught this. It
told me `/loans/` was dead, along with about sixty other links that genuinely were fine.
I fixed the false ones — and while I was at it I taught the checker to treat `/loans/` as
meaning `/loans/index.html`, because that is what the test server on my machine does. So
I took a true warning and taught the instrument to stop giving it. The site then passed
every check I had, for a day, while being broken.

The lesson I have written down for the whole fleet is short: **when a checker disagrees
with the live site, change the checker to match the live site, never the other way round.**
And I was right about fifty-seven of the sixty, which is exactly why I stopped looking.

**Two: the structured data on all thirteen guides was invalid.** Every guide carries a
small hidden block that tells Google what the page is — its title, description and date.
Mine had one character wrong in how it was generated, which made all thirteen unreadable.
Google throws away structured data it cannot parse without saying anything, so nothing
complained; the guides were simply not eligible for the richer search results they should
have been. Same underlying mistake as the first one: I had checked that the block was
*there*, never that it was *valid*.

**Three: the site claimed twenty-four calculators and has twenty-three.** When I dropped
the one that turned out not to be a calculator, I never updated the wording. It also said
"all 12 loan calculators" where there are eleven. Small, but it is a false statement on a
finance site, which is the last place for one.

**All three are fixed and live, and I fixed them so they cannot come back.** Rather than
correcting thirteen separate places, there is now one definition of how a section link is
written, and the counts are counted from the actual list of calculators instead of typed
by hand. The build now refuses to produce a page that links to a folder, or that carries
unreadable structured data — and I deliberately broke each of those four safety checks to
confirm they actually stop the build. A check that has never once gone off is not a check.

I also wrote a single verification command that tests the live site rather than my
machine, and it caught a fourth mistake within a minute of existing.

**The site is now in the framework.** All forty-one pages are registered, and every one is
marked so that nothing can regenerate it — the calculators' arithmetic is untouchable by
design. Nothing was handed to an AI to rewrite: I checked that explicitly, because that
was the one outcome worth being paranoid about.

**But the way the framework copies a site in is not safe, and I am glad I gated it.** To
adopt a site, the platform crawls it — and what a crawler captures is the page *after* the
browser has run all its code, not the file as it was written. All forty-one pages came
back changed. Mostly cosmetic, but two matter: the "skip to content" link at the top of
every page — the first thing a keyboard or screen-reader user hits — had been turned into
a link that reloads the whole page instead of jumping down it. And the amortisation
calculator came back eleven kilobytes bigger, because the crawler had captured the
year-by-year table the page builds when you open it, and baked it in.

So I held the platform's forty-one "publish this" jobs the moment they were created —
under two minutes' notice, so it had to be automatic, not me watching — replaced all
forty-one with the real files, and then let one through as a test. It republished the page
with **no change at all**, which is exactly the property worth having: the framework can
now rebuild the site without altering it.

**A caution I want recorded for whoever touches this next.** There are now two things
that can write those forty-one files: my build scripts, and the framework. They agree
today because I just made them agree. If someone changes the build scripts and does not
re-sync the framework's copy, the next framework rebuild will quietly undo their change.
There is a one-command fix for that, and it is written down in the runbook.

**And the divergence is now recorded as real settings, not a note.** This is the part you
asked for — that the two sites evolve apart, managed by the framework rather than by
someone remembering. The new site's positioning is written into the fields the content
system actually reads when it writes a page: who this site is for (someone whose loan or
car finance and their mortgage affect each other), what is in scope, and an explicit rule
that when a subject could be written either as plain single-topic explanation or as a
crossing-point question, it must always be the crossing-point version.

**Two honest caveats on that.** First, I checked which settings are actually *read*, and
the most obvious one is a dead end: the field literally named "audience" is filled in on
twenty-nine of the thirty-three sites in your estate and **nothing anywhere reads it.** My
own earlier plan had named it as one of three places to write this. That third of the work
would have looked done and done nothing.

Second, and more important for what you asked: **there is nothing in the platform that
detects two of your sites drifting back together.** I looked. The duplicate-content
checking that exists only ever compares a site against itself. So these settings are the
entire mechanism — they steer new writing, but nothing will raise a hand if the two sites
converge again. If you want that guard, it does not exist yet and would need building.

**What is left.** The thirteen guides are registered but still frozen, so the framework
cannot yet improve them — handing them over needs them broken into sections first, because
of how the page assembler works, and that is a proper piece of work rather than a flag to
flip. And two small things only you can do in Cloudflare: turn on "Always Use HTTPS" (the
site currently answers on plain `http://` as well, which means the same pages exist at two
addresses — the exact thing we are trying to avoid), and decide whether you want `www` to
work, which it currently does not on any of your sites.

---

## 6 August 2026 — the voice you chose is now on the site, and the widgets survived it

You picked the gentle explanatory register (trial H), approved four sample rewrites,
and said "do the whole site — I'll check it then". This is where that stands.

**The blocker was structural, not editorial.** All forty-one pages were frozen whole
documents — one stored file each, which the framework ships byte-for-byte and never
looks inside. No writer could touch a word of them. So before any copy could change,
each page had to be broken into parts the system can edit: the text in editable
blocks, the calculator in its own locked block. That is the "proper piece of work"
the last entry said the guides needed, and it is now built and proven on both kinds
of page.

**Two pages are live in the new voice.** One guide (how loans cut what you can
borrow) and one calculator page (debt consolidation). I picked two deliberately
rather than one, because a single test page can agree with you by luck.

**Before touching anything, I recorded what all twenty-three calculators compute.**
A real browser drives each one and writes down every answer. That is the only defence
against the failure mode that actually matters here: rewriting the words around a
calculator and silently breaking the arithmetic, which would look fine on screen.
After the consolidation page was rebuilt in the new voice, every one of its numbers
came back identical, down to the pound.

**And the rebuilt pages came out byte-for-byte as predicted.** I built an offline
model of what the framework would produce, predicted both pages exactly, then compared
against what actually went live. No difference at all, on either. That means the
remaining pages can be done with confidence rather than hope — and each one still gets
checked the same way.

**One honest note about what changed beyond the words.** Each page used to carry its
own copy of the header, footer and page furniture. They now share one copy, which is
the point of the exercise, but it costs three small things: the navigation no longer
highlights which section you are in, each page's social-media preview tags are gone
(the framework emits its own structured data and a canonical link instead, which is
the part search engines actually use), and the guides' hand-written article markup
goes with them. I judged those worth it to make the site editable. If you disagree
about any of them, say so — they are recoverable.

**The legal page is exempt from the voice, as your rules require, and I have not
touched a single compliance line anywhere.** The FCA risk warnings, the disclaimers
and the debt-help signposting are copied across byte-for-byte and the tooling refuses
any rewrite that alters them.

**What is left.** Thirty-eight pages of copy are being written now, in the same
register, each one checked automatically for invented figures, lost links, broken
anchors and tampered compliance text before it can go anywhere near the site. Then
they go up in batches with the same verification each time. One practical wrinkle:
the step that writes to the live database needs your approval each time it runs, so
you will see a few of those come past — or you can allow it once in settings and I
will run the batches through.

---

## 2026-08-08 — we now check whether the calculators are RIGHT, and two of them are not

You asked, back on the 6th, for "a comprehensive check on the calculators that
they produce validated output". I had to tell you at the time that we did not
have one. We do now, and it found real problems on the first run.

**The distinction that mattered.** Everything we had built until now recorded
what each calculator answered on a particular day and checked it still answers
the same. That catches a calculator we break. It cannot catch a calculator that
was wrong from the day it was written, because the recording faithfully captured
the wrong answer and every check since has been confirming it.

So the new work computes the answers independently — from the standard loan
formula, and from the stamp duty tables published by HMRC — in code that has
never looked at the website's own workings. Then it types numbers into the real
pages in a real browser and compares. Being genuinely independent is the whole
point: if I had written the checker by reading the site's code, it would have
agreed with every mistake in it.

**The first problem: the stamp duty calculator is running a tax rule that
expired sixteen months ago.** First-time buyers used to get relief on properties
up to £625,000. That was a temporary rule, and it ended on 31 March 2025 — since
then the relief stops at £500,000, and above that you pay the ordinary rates on
the whole price. Our calculator still uses £625,000. For anyone buying between
£500,000 and £625,000 it quotes exactly £5,000 too little.

The uncomfortable detail is that the page itself says, in its own text just
above the calculator, that the temporary period ended in March 2025. The words
are right, the table of rates beside them is right, and only the arithmetic is
out of date. Whoever wrote it clearly wasn't sure — the code still carries their
comments saying "rules vary" and "let's use standard rates for safety". There is
a second, smaller version of the same thing: it charges the buy-to-let surcharge
on properties under £40,000, where the surcharge doesn't apply at all.

**I have not changed the numbers.** These are tax figures a real person would
budget against, and changing what we tell them is a decision for you, not for
me. The finding is written up with the HMRC pages cited in `bugs_open/225`.

**The second problem: type a 0% interest rate and six of the seven loan
calculators break.** That is not a made-up input — 0% credit cards, interest-free
car finance and employer loans are all ordinary things, and the car finance
calculator is exactly where somebody would type it.

There's a tidy explanation. The site has one shared file with the loan formula
written properly, including the special case where the rate is zero. Every
calculator in the mortgages section uses it, and every one of those passed.
Every calculator in the loans section has its own private copy of the formula,
and not one of those copies handles the zero case. So it isn't seven separate
mistakes, it's one mistake copied six times.

Three of them show "£NaN" on screen, which at least looks broken. The other
three are worse: they quietly show you the answer to whatever you typed *before*,
with no warning at all. The clearest way I found to demonstrate that doesn't
involve reading any code — type the same three numbers into the same three
boxes, but get there in a different order, and you get a different answer.
£143.47 one way, £429.81 the other, same numbers on screen.

The one I'd want fixed first is the loan comparison tool. Give it a 0% loan and
a 5% loan and it tells you the 5% one is cheaper. Confidently, in a green box.

**Something I got wrong, which is worth telling you about.** My first version of
the checker accused the rate forecaster of miscalculating. It hadn't — my
formula was the naive one and the page was doing the more careful thing, working
out what's left of the mortgage at each stage rather than starting from scratch.
The page was right and my checker was wrong, and I only found that out by going
and doing the sums a second way. I've kept the wrong version in the code as a
deliberately-labelled wrong answer, so if anybody ever rewrites that page badly,
it'll be recognised rather than just failing.

There were three more like that — including one where my own reporting quietly
downgraded the stamp duty finding from "this is wrong" to "this is a matter of
opinion", because I'd used the same mechanism for both. That one is worth
remembering: the tool I built to explain findings was capable of explaining the
most important one away.

**Where that leaves us.** Eighteen of the twenty-three calculators now have an
independent check on their arithmetic, run on the live site. Five of them —
the credit score, the damage checklist, the application tracker, the mortgage
scorecard, and the portfolio dashboard — don't have a right answer to check
against, because they're our own scoring judgements rather than sums. I haven't
invented one for them. Instead I check things that must be true whatever the
scoring is: answering worse can't score better, percentages stay between 0 and
100, the same answers give the same result, saved data comes back unchanged. All
of those pass, and I'd want you to read them as weaker evidence than the
arithmetic ones, because they are.

**What I need from you:** a decision on the stamp duty numbers. Everything else
I can get on with.

---

**2026-08-09 (bugfix 224 session).** The six loan calculators that got a 0%
interest rate wrong are fixed and live. You asked a session to take this on by
name, so I've treated that as the go-ahead the validation report was waiting
for on this one.

What was wrong, in plain terms: if someone typed 0% into any of the loan-side
calculators — and 0% is a real thing in this market, car finance and balance
transfers advertise it — three of them printed "£NaN", one of those three
confidently recommended the WRONG loan, and three quietly left the previous
answer on screen so it looked fresh. The mortgage-side tools were always fine,
because they all share one well-written piece of code. The loan tools each had
their own private copy of the maths, and every private copy was missing the
same case.

The fix is the one the report recommended: delete the private copies and make
all the loan tools use the same shared code the mortgage tools use. There is
now one place this arithmetic lives, so this class of bug can't come back by
one page drifting. The tools also now always show an answer (or a clear blank)
— never a leftover one. Nothing changed for normal interest rates: I checked
every tool against its old answers before and after, to the penny.

Proof it's right: the validation suite you commissioned went from 23 failures
to zero on these tools, with all its self-checks passing, and the stamp-duty
failures still showing (that's the separate decision still waiting on you —
bug 225).

Two of our own tools would have bitten us tonight and are now fixed: the
repair script that keeps the database and the repo in step would have
destroyed the decomposed consolidation page if run as documented, and the
deploy script crashed halfway through its own bookkeeping. Both are mended
and written up.

---

**2026-08-09 (later).** Ran the emit-criteria step. This is the one that hands
the checking over to the platform: instead of the calculators only being
verified when someone remembers to run the harness, their correct answers get
written into the acceptance record so the system re-checks them on its own
schedule, unprompted.

It half worked, and the half that didn't is worth your attention. Seven tools
now have that automatic coverage. **Ten don't — and not because they're wrong.**
On nine of the mortgage calculators, and on the debt consolidation tool, the
button the user actually presses ("Calculate Tax", "Calculate Savings", and so
on) has no identifier in the HTML. The automatic checker has to be able to say
"press this button", and it can't name a button that has no name — so it
refuses, rather than guessing and writing a check that drives the tool
differently from how we tested it. That refusal is the right behaviour.

The awkward part: **the stamp duty calculator is on the refused list.** The tool
that was quietly wrong for sixteen months is the one we can't yet watch
automatically. The fix is small — an identifier on each button — but it's a
change to those pages, so I haven't done it as part of tonight's work.

I also double-checked the numbers before letting them be written down as "the
right answers", by recomputing every one of the 52 from the independent maths.
Six came out wrong at first. They were my error, not the site's: the test drives
some vectors with fractional terms like 6.9 years, and I'd rounded that to whole
months where the calculators don't. Corrected, all 52 agree. I've kept the six
failures on record, because a checker that has never failed isn't evidence of
anything.

**Nothing is switched on yet.** The checks are written and I've verified the
system is capable of running them, but installing them is a separate step and a
separate decision — partly because of the ten uncovered tools above.

---

**2026-08-10.** All seventeen calculators are now watched automatically, and the
watching has already paid for itself.

The system picked this site up on its own at twenty past three this morning, ran
the fourteen calculators that were due, and **found one genuinely broken**. The
equity release tool: put in an eligible age, it shows you a figure; change the age
to one that doesn't qualify, and it left the old figure sitting there as though it
still applied. Same fault as the loan calculators we fixed yesterday, but hidden
behind an age check rather than an interest-rate check, which is why the earlier
sweep walked straight past it.

That prompted me to stop trusting my own summary. I'd written that this kind of
fault was dealt with; it wasn't. So I searched for the *shape* of the mistake
rather than the version of it I already knew, and found three more — the bridging
loan tool showing a cost for a deal it had just declared impossible, and the
investor tool keeping an old percentage on screen after you clear a box. All four
are fixed. **Ten in total across the two sites, and only six were the interest-rate
version.**

The other thing worth telling you: the safety catch worked. When that tool failed
this morning, the system's normal reaction would have been to send an automated
writer in to change the calculator until the check passed — on arithmetic about
consumer credit. I'd blocked that yesterday, and the record shows it refusing and
asking for a human instead. It has now been tested by a real failure rather than
just asserted.

Last piece: the debt consolidation page was the one tool that couldn't join the
automated checks, because of how it's put together. It can now, and it passes.
Seventeen of seventeen.

Nothing is outstanding. There's a full handoff (`HANDOFF_2026-08-10_continue_here.md`)
if you want to pick this up in a new conversation. The only thing I'd flag as a
choice rather than a task: the same class of fault has never been looked for on
mortgagecalculator.co.uk or loancash.co.uk, which come from the same family of
pages.

---

## 2026-08-10 — the two sites are unlocked, except for the twenty pages where unlocking would have destroyed the calculator

You asked me to unlock loanandmortgagecalculator.co.uk and loancash.co.uk and make
everything on them fully editable and upgradable. I've done most of it, and
stopped short of one part on purpose. Here's the honest picture.

**First, a correction to how this was described to you.** "Locked" didn't mean the
content couldn't be edited. It could — the framework's page-rerender and
section-editor paths were never blocked, and that's how we've been changing these
pages all week. What the flag blocked is the framework rebuilding a page *from
scratch*. So the before-state was "the pipeline may not rewrite this page", not
"nobody can touch it". Worth being clear about, because "we unlocked the content"
would have overstated both halves.

**What's now unlocked: 39 pages.** All the guides, the index pages, the legal
pages — 24 on loanandmortgagecalculator, 15 on loancash. Those are fully in the
framework's hands now and can be rebuilt and upgraded like any other page. Done
as a proper migration, with a check that refuses to run if the numbers don't match
what I measured, and I deliberately broke that check first to prove it actually
stops things rather than just printing a reassuring line.

**What I did not unlock: the 20 pages that hold a calculator.** This is the part I
want to explain, because it's the opposite of what it looks like.

On those pages the entire page — text, layout and the calculator's code — lives as
a single block in the database. And the way the framework rebuilds a page is:
write new content, **commit it to the repository the website is published from**,
and only *then* check whether it was allowed to. The check comes after the
publish. So flipping that flag on a calculator page doesn't produce a warning; it
produces a page where the calculator has been replaced by prose, already live.

The flag I'd have been removing is the only thing preventing that. It exists
because exactly this happened on another site once. And two of those twenty are
the stamp duty and loan calculators we've just spent three days proving correct.

**The right way to do it, which is real work rather than a flag.** Split each of
those pages into its text parts plus a separate, protected calculator part — then
the text is fully editable by the framework and the calculator survives being
upgraded. One of the 20 (`loans/consolidation`) is already in that shape, so we
have a working example. There's one specific trap to handle first: unless the
calculator's slot is renamed to something the framework expects, it survives the
rebuild but gets silently moved to the bottom of the page, underneath all the new
text. I've written that up rather than discovering it on a live page.

**Something else I found that you should know about.** Neither site has a "site
plan" — the framework's description of what the site should contain. Unlocking a
page says the framework *may* rebuild it; the plan is what it rebuilds *from*.
With no plan, those 39 pages are now permitted and still idle. Seeding one is the
next step, and there's a known bug to read first.

**I ran the checks again afterwards.** All 170 arithmetic checks on the 18
calculators still pass, with the controls re-run in the same session so a green
result means something. I also checked loancash.co.uk's three tools for the same
0% bug — clean, and that's the first time anything has ever checked them.

**And that turned up the next thing worth doing.** loancash.co.uk has no
arithmetic checking at all, and two of its three tools are the dangerous kind:
they hardcode the FCA payday-lending caps — 0.8% a day, £15 default fee, 100%
total cost — with nothing verifying those numbers against the regulator. That is
precisely the shape of the stamp duty problem: a rule with a date on it, sitting
in a page nobody checks. I'd want to point the same method at it next.

Everything is written up in a handoff so this can continue in a fresh
conversation without losing the thread.

---

## 2026-08-10, evening — you asked for decomposition, and the job is bigger and easier than the last note implied

You asked for the components on both sites to be decomposed so the framework can
fully control them, and for the handoff updated so a fresh chat can pick it up.
The handoff is written (`HANDOFF_2026-08-10c_continue_here.md`). Two things came
out of checking the ground first, and they pull in opposite directions.

**The job is bigger than it looked.** Unlocking those 39 pages this morning gave the
framework *permission* to rebuild them. It did not give it anything to work with.
Fifty-seven of the fifty-nine pages across the two sites are still one solid block
each — the whole page, text and calculator together, as a single item in the
database. The framework can't control a part it can't see, so decomposition is the
actual work, and it applies to the pages we unlocked this morning just as much as to
the calculator ones we didn't. Only two pages on the estate are in the finished
shape today. Both are useful: one is a calculator page done properly, one is a plain
page done properly, so we have a worked example of each rather than guessing.

**The job is also easier than this morning's note said.** That note flagged a known
platform bug as the thing standing in the way — the short version being that once a
page is split up, the framework loses track of the pieces and can't rebuild it. If
that were still true, decomposition would be a one-way door and I'd have come back
to you before starting. It isn't true. That bug and its twin were both fixed and
shipped on the 6th of August, four days before the note that called them blockers.
They're still sitting in the "open bugs" folder because you asked us to leave found
bugs there rather than filing them away — which is fine, but it means the folder a
bug sits in tells you nothing about whether it's fixed. I checked the running system
directly rather than taking the paperwork's word for it, and both fixes are live.

**One near-miss worth telling you about**, because it's the kind of thing that
quietly derails a day. My first check on one of those fixes came back "not there" —
which would have meant the fix had been undone and the whole task was blocked. It
was my query that was wrong, not the system; a second check on the same row a moment
later showed the fix present. I'd have reported a blocker that didn't exist. I've
written down what the reliable check is so the next session doesn't repeat it.

**How I'd sequence it.** Plain pages on the main site first — no calculator to
break, and the tooling for them already exists and has been tested here. Then the
calculator pages, one at a time with the arithmetic re-checked between each, because
that failure is silent and the thing that breaks is a live loan calculator. The
second site last, since it has no tooling of its own and we'll know much better what
that tooling needs after doing the first.

**One thing I want to flag rather than bury.** Decomposition gets you fine-grained
editing — change one paragraph without touching the calculator. It does not by
itself get you "press a button and the framework rebuilds the page", because neither
site has a site plan for it to rebuild *from*. That's a separate decision with its
own risks, and I've kept it out of this task rather than quietly bundling it in.

**And the thing I still think is the highest-value item on the estate isn't this
one:** loancash.co.uk's tools hardcode the FCA payday caps with nothing checking
them against the regulator, which is exactly the shape of the stamp duty error that
was sixteen months out of date. It's independent of all the above. Say the word and
I'll point the same method at it.

---

**2026-08-10, evening — before starting the decomposition I found six calculators
that had been left unguarded, and I've put the guard back.**

You said yes to decomposition, so I picked the lane up and re-ran the safety checks
at the top of the handoff before touching anything. They all passed. Then I listed
the pages to see what I'd be working through, and the list didn't match what the
handoff said.

Here's the background in one paragraph. Every page on these two sites carries a flag
saying whether the automated pipeline is allowed to rebuild it from scratch. For the
guide pages that's fine — a rebuild just rewrites prose. For a calculator page it is
not, because those pages are still one indivisible block of HTML with the calculator's
code inside them, so "rebuild the page" means "write new prose and throw the
calculator away". The flag is the only thing stopping that. This morning's migration
deliberately lifted the flag on the pages with no calculator and deliberately left it
on the twenty that have one.

**It missed six.** It decided which pages had calculators by searching the stored HTML
for two particular spellings — `onclick` and `addEventListener`. Six of these pages
don't use either. Four of them wire up their buttons a different way (`oninput`,
`onsubmit`, `onchange`), and two of them keep their code in a shared JavaScript file
that the page merely links to, so the give-away words aren't in the page at all. They
are: compare-loans, interest-rate-stress-test, loan-vs-savings, settlement-calculator,
damage-checker and fact-finder.

**This was not theoretical, and that's the part I want you to see.** Those six pages
were all flagged "needs rebuild", all still one indivisible block, and each had a
pending job queued against it. And the rebuild has already been attempted on them:
back on the 9th, twenty of these pages went through the full rebuild and were stopped
at the last moment by exactly the guard we're talking about — the error message is
still in the database, and it says in plain terms "a generic section save would clobber
it, refusing to overwrite". That refusal is why those calculators still exist. This
morning's change took it away for six of them, for about seven hours.

I've put it back — a new migration, applied and recorded, with a note saying what it
did and why. Nothing else changed. I checked afterwards, from the database rather than
from the migration's own say-so, that exactly those six moved and that none of the
genuine guide pages were caught by mistake.

**The uncomfortable detail worth knowing.** This morning's migration *did* have a
safety check designed to catch precisely this mistake — "make sure no calculator page
got unlocked" — and it was tested by deliberately breaking it to watch it complain.
It still didn't catch this, because the check looked for calculators the same way the
original decision did. Two things that are blind in the same way will always agree
with each other. I've written that up as a standing warning for other sessions, since
it isn't specific to this site.

I also managed the same class of mistake myself, half an hour later: my own check
compared pages by their web address, and it turns out both sites have a page at
`/guides/jargon-buster.html`, so it couldn't actually tell them apart. I found that
by deliberately breaking my own check rather than by reading it — which is the whole
argument for doing that step.

**What this does to the plan.** Nothing was decomposed or undone; six pages simply
moved from the easy pile to the careful pile. The easy pile (prose only, no
calculator) is 17 pages rather than 23. The careful pile is 22 rather than 16. The
route for the careful pile is unchanged and is the one you approved: break the page
into parts, lock the calculator part so a rewrite can't touch it, then allow rebuilds
— one page at a time with the arithmetic re-checked between each.

Next, unless you'd rather I did something else: start the 17-page pile, beginning with
the one page that is already in the finished shape, so the first thing I prove is the
loop itself rather than a conversion.

---

## 2026-08-10, late evening — Track A is loaded and ready to fire, and I want to be straight about what it gets you

You said go ahead with Track A, so I've done everything up to the point where it
starts changing live pages, and left that last step for the fresh chat. The brief is
`HANDOFF_2026-08-10d_track_a_prose_decomposition.md`.

**What I actually did rather than planned.** I ran the tooling for real, in read-only
mode, against the live database. It builds the plan for all seventeen pages cleanly
and reports zero calculators among them. Then I ran the prediction step across all
seventeen — it produces the exact page each one would become, and all seventeen
passed. So this isn't "here's a sequence that should work"; the only untried part is
the button that writes.

**I checked the calculator question properly this time, and I'm glad I did.** Earlier
today another session found that six calculator pages had been unlocked by mistake
this morning and re-locked them. Before I knew that, I'd run my own check for the
same problem — and it came back clean. It came back clean because the fix had already
landed ninety minutes earlier. Had I run it this morning it would have given me the
same reassuring answer while six calculators sat exposed, because I'd used the same
flawed test that caused the problem in the first place. So I redid it against the
hand-written list of which pages are calculators, and the result is now a proper
proof: all twenty-three calculators are locked, and not one of them is in the
seventeen I'm about to touch.

**One thing I broke on purpose, and it told me something useful.** There's a check
that runs before each page is written. I wanted to know whether it would catch a
corrupted plan, so I deliberately corrupted one and ran it. **It passed.** That check
guards the destination — it makes sure nobody else has edited the page since our
baseline — but it does not verify that the new content is faithful to the old. That
job is done earlier, by the tool that builds the plan, which refuses outright if the
text doesn't match. So we are covered, but by a different thing than I'd assumed, and
anyone who trusts the wrong one will be trusting nothing.

**Now the honest part about what Track A delivers.** The word "decomposition" suggests
a page gets broken into many editable pieces. For these seventeen it doesn't: each one
becomes a single editable block of text. That's the tool working as designed, not a
fault. What you gain is real but narrower than it sounds — the framework can rewrite
that block, the page stops being an opaque lump it refuses to touch, and the page
becomes properly rebuildable. What you don't gain is the ability to say "change the
third paragraph" and have the framework know what that means.

There's a related point I'd rather you heard from me than discovered later. Right now
every one of these pages is built from a single shared template that is also used by a
hundred and fifty-four pages across three of our sites — and that template has been
damaged once already. After this work, they'll use a different shared template, used
by twenty-nine pages across two sites. That's better. It is not the same as each page
owning its own. Only the calculator pages get that, and only in Track B.

So: ready to go, seventeen pages, smallest and least-visited first, homepage last,
checking each one before moving on.

---

## 2026-08-11 (morning) — Track A started, and the safety net turned out to be broken

Picking up where the last session left off, with your go-ahead on Track A. I re-checked
everything the brief claimed before touching anything, and it all held — the right
pages, the right protections, the right code live on the servers. Two things did not
hold, and both are worth telling you about, because one of them was the thing that was
supposed to save us if this went wrong.

**First, a mistake I nearly made myself.** The most important check in this whole job
is the one proving that none of the seventeen pages I'm about to change is a
calculator. I ran it and it came back perfect — three empty lists, exactly the answer
the brief recorded. It was worthless. The hand-written list of calculators writes them
one way (`loans/compare-loans`) and the database writes them another
(`/loans/compare-loans.html`), so the two lists had no word in common and could never
have overlapped no matter what was true. The check agreed with the right answer while
being incapable of disagreeing with anything. What saved me was printing one extra
line I didn't strictly need — a list that was also supposed to be empty and came back
with twenty-three entries. That mismatch is the only reason I looked twice. I've fixed
the check, and it now genuinely proves the point: all twenty-three calculators are
protected, none of them is in my seventeen. I've written this up in our log of wrong
calls, because it's a good example of a test that passes for no reason.

**Second, and more serious: the undo button did not work.** Every page I change gets
its previous state copied into a backup table first, so any page can be put back
exactly as it was. When I ran the very first page, it stopped dead with a database
error — the backup step couldn't run at all, because the main table has gained a
column since that backup table was created back on the fifth. The good news is it
failed *before* changing anything, so nothing was half-done. The bad news is that
nobody would have discovered this until they needed it.

Then I found the worse half. The backup was written so that every run copies in any
row it hasn't already seen. That sounds sensible and isn't: once a page has been
converted, the *next* run copies the page's **new** content into the backup as well,
sitting alongside the old. Restoring that page would then put **both** versions back on
it at once — producing exactly the corrupted page this whole process is designed to
avoid, delivered by the mechanism meant to protect us. It had already happened to one
page, back on the fifth. Had I converted my seventeen one at a time as planned, I'd
have quietly ruined the undo for about sixteen of them, and we'd only have found out at
the worst possible moment.

I've fixed both, repaired the damaged backup (keeping the stray copy rather than
deleting it), and then — instead of assuming the fix worked — I converted the first
page, put it back, and checked it was byte-for-byte the original. It was. So the undo
button is now something we've watched work, not something we believe in.

**Where that leaves us.** Two pages are converted and live: the legal page and the
guides index. Both came out byte-for-byte identical to what we predicted offline
before touching the site, which is the strongest evidence we can get that nothing
drifted. I also did something the brief hadn't asked for: before changing anything, I
downloaded all seventeen pages as they were and compared the words a reader actually
sees against what our conversion would produce. All seventeen matched exactly. So
whatever else changes in the markup, nobody reading these pages will see a different
word.

**One thing you should know is changing, because it is a genuine loss.** These pages
currently carry three tags that control what appears when someone shares a link on
WhatsApp, Facebook or LinkedIn — the preview title, description and address. The
converted pages lose them, keeping only two generic ones. They also change from
declaring themselves British English to plain English. Neither is a surprise: both
were written down and accepted back on the fifth of August, when the decision applied
to two pages. It now applies to nineteen, and eventually to all fifty-nine, so I'd
rather say it out loud again than let it pass on the strength of a decision made when
it was smaller. Nothing a reader sees on the page changes; it's the sharing preview
and the language tag. It is fixable, but the fix is a change to the shared platform
rather than to this site, so it isn't part of this job.

Fifteen pages to go — the twelve guides, the two section indexes, and the homepage
last, on its own.

## 2026-08-11 (midday) — Track A is done: all seventeen pages converted and live

All seventeen are converted and serving, and every one of them came out
**byte-for-byte identical** to what we predicted offline before touching the site.
Not "looks right" — identical, to the character, on all seventeen. The site now has
no old-style frozen pages left outside the calculators.

I did them in a deliberate order rather than all at once. The legal page first,
because it's the smallest and least visited. Then the guides index, because it's a
different *shape* — thirteen of the pages are narrow-column guides and four are
full-width hub pages, and the narrow ones passing told me nothing about the wide
ones. Once each shape had proved itself on the live site, the rest of that shape
followed. The homepage went last and on its own, as planned.

**The arithmetic is untouched, and I checked rather than assumed.** None of these
seventeen carries a calculator, so the sums should not have moved — 170 checks pass,
none fail. More importantly I ran the two controls in the same sitting: one that
proves the checker can read numbers off a page at all, and one that feeds it a
deliberately wrong answer to confirm it actually fails when it should. A green run
without those is just a green light with no bulb behind it.

**One genuine problem found, and it isn't ours.** Our own site checker flagged that
the homepage now tells search engines its official address is
`loanandmortgagecalculator.co.uk/index.html` rather than plain
`loanandmortgagecalculator.co.uk/`. Both show the same page, but that tag is exactly
how you tell Google which of the two to treat as the real one — and it's now naming
the version nobody links to.

Before blaming our work I checked ten other sites of ours. **Nine of the ten already
have the same problem**, and have done for a while. It comes from shared platform
code that builds the address by gluing the page's filename on the end, with nothing
to say "the front page is just a slash". Our homepage was in the correct minority
only because nothing had ever rebuilt it — this work moved it onto the same path as
everything else.

I'll admit a wrong turn here, because it nearly became a wrong bug report. The tenth
site was serving the correct address, and I assumed it must be doing something
special. It isn't — I checked, and my explanation was simply wrong. The real reason
is that its homepage has never been rebuilt at all, so it's still serving its old
hand-made file. It wasn't an exception to the rule; it was a page the rule hadn't
reached yet. Had I shrugged at "nine out of ten, close enough", I'd have filed a
report whose one disagreeing case was actually the strongest thing supporting it.
It's written up as a bug for the platform, with the fix and its blast radius.

**One thing I fixed rather than tolerated.** Our site checker was set to fail
whenever a page lacked those social-sharing tags I mentioned last time. Since no
converted page can have them, that check was on course to be permanently red — two
pages yesterday, nineteen today, all fifty-nine eventually. A checker that is always
red is one everybody learns to ignore, and it takes its genuine findings down with
it. It now reports the count as a known, accepted loss on converted pages, and still
fails properly on a hand-built page that's genuinely missing them. I tested that by
breaking a page on purpose to make sure it still complains.

**What's next, and what I'd flag before Track B.** Track B is the twenty-two
calculator pages, and it is a different risk class entirely — those pages do sums
people rely on. There's one thing the last handoff marked as reasoned-but-unmeasured:
what happens to a locked calculator when the page is rebuilt. Track A couldn't test
it, because none of these pages has a locked component. It should be measured on a
single calculator page before Track B goes anywhere near the rest.

---

**2026-08-11, afternoon — you asked me to look over it all again. I did; it holds.**

Three things checked, from the live system rather than from anyone's notes,
including mine.

First: the six calculator pages I re-protected yesterday evening are still
protected, and the protection did what it was for — the seventeen prose pages got
converted around them without any rebuild going near a calculator.

Second: the conversion work itself. The counts on the live database match what the
other thread reported exactly — every prose page on the calculator site is now in
the new component form, no page was left half-done, and the calculators' arithmetic
checks still pass. The one caveat in their report is right and worth keeping: their
"the live page matches our prediction" proof is only valid until the next software
roll, and has to be re-run after each one.

Third: one scare that turned out to be nothing, worth explaining because of *why*
it was nothing. A note elsewhere in the repo says a certain sync tool "would have
reverted migration 377" — and 377 is my re-lock, so that read as a close call. It
isn't: two different pieces of work were both numbered 377 on the same evening, and
the note is about the other one. Duplicate numbers turn out to be common in that
folder — sixty-odd cases going back months — so the rule we already use for bug
numbers applies: trust the full filename, never the bare number.

One correction of my own from last night: I'd called the loancash regulatory-cap
check the most valuable unstarted job. Someone did it today, properly, against the
regulator's own handbook — and the caps are fine: unchanged since 2015, correctly
implemented. My reasoning-by-analogy with the stamp-duty bug didn't hold, because
stamp duty rates move every Budget and this cap has moved once in eleven years. The
genuinely unchecked thing on that site is now the complaint-deadline tool, whose
rules do move.

Where that leaves the decisions: the big one is still yours — whether to start on
the twenty-two calculator pages. The measurement you authorised came back
reassuring (a rebuilt page keeps its calculator in place as long as the rebuild is
told the calculator's slot exists, which the normal path always is; the danger case
is only the not-yet-seeded "site plan" route). Nothing in my pass argues for
delay beyond what the other thread already said: one live end-to-end run on the
already-converted consolidation page first, then one page at a time.

---

**2026-08-11, late afternoon — you asked for the homepage to be rewritten by the
content agent, with a before/after comparison. The agent ran; a safety rail
stopped its version going live; here's the comparison.**

The rewrite went through the framework's normal route, and the framework itself
refused to save the result — a guard that stops a page losing more than half its
text in one save. That refusal is good news twice over: the live homepage is
untouched, and we still got the agent's full answer out of the failed run's
records, so you can compare without anything having shipped.

The short version of the comparison: the agent writes better *sentences* — its
four paragraphs are plainer and more human than what's live, which speaks well of
the recent voice work. But it wrote an *essay* where the homepage is also the
site's *directory*: its version drops from 617 words to 235, from 35 links to 3,
and loses every heading including the main one. Eighteen pages — the stamp duty,
repayment and affordability calculators among them — would lose their only link
from the front page.

The likely reason isn't the prompt alone: this page has never been given a
content brief in the database, so the agent had nothing telling it "this page is
a directory — keep the calculator cards". My recommendation is to write that
one-paragraph brief, run it again, and compare a second time. That tells us
cleanly whether the prompt needs revisiting or whether it just needed to be told
what the page is for.

Side-by-side, with everything above laid out:
https://claude.ai/code/artifact/ca0d8274-929b-42c0-95e1-18b982343cc7

---

**2026-08-11, evening — your five decisions are recorded and the first two are done.**

The homepage rewrite is live. Second attempt, this time with a proper brief on
file — the first attempt had nothing telling it what the page was for, and that
turned out to be the whole story. With the brief, the agent kept all twelve
calculator cards, linked every guide on the site, and came out slightly *longer*
than the old copy, in a plainer voice. It took the liberty of swapping two of the
twelve featured cards for two others; nothing lost its place on the site, but if
you want the original two back on the front page, that's a one-line note. The
old copy is snapshotted and restorable. The side-by-side page is updated with
what's actually live now — and it answers your prompt question: the prompt was
fine; the page needed a brief. Every page we decompose from here should get one.

The canonical bug is fixed in code, tested, and submitted for review — it takes
effect at the next software roll. One judgement call made on measurement: only
the homepage form gets normalised, because the "tidier" directory-style addresses
for section pages turn out not to work at all on our hosting, and a fix that
pointed at broken addresses would have been worse than the bug.

Still to come, in order: the twenty-two calculator pages (starting with the
simplest one as the proving run), then the site plan seeded and iterated until
it matches the site's current size — with your no-shrink rule as the acceptance
test — then the social-tags work once the canonical fix has rolled, then the
complaint-deadline checker.

---

**2026-08-11, night — the design is back, and it was my miss, not the agent's.**

You were right: the homepage had lost its design. What happened: when I briefed
the rewrite I described what the page should *say* and how it should be
*organised*, but not the visual building blocks the site's stylesheet knows how
to draw — the hero banner, the card grid, the buttons. The writing agent
delivered exactly what was asked, so the page came out as plain text. My
comparison read the words and missed it entirely; you saw it in minutes.

Only the homepage was affected — I checked the rest of the converted pages and
they kept their look throughout.

The fix went through the same framework route: the page's brief now includes the
design vocabulary, and the rewrite ran again. The homepage now serves the agent's
copy inside the site's own design — hero, twelve calculator cards in their grids,
buttons — verified on the live page, not assumed. Both lessons are written where
the next job will find them: every page brief carries the design vocabulary from
now on, and every before/after comparison counts the design markup, not just the
words.

One admin note: the side-by-side page has moved to a new link (the old one got
stranded by an account-context change): 
https://claude.ai/code/artifact/70514218-28e4-44ce-936b-07a012c74330

---

**2026-08-13 — your "everything editable and reusable" question, answered with a
working page.**

You asked whether decomposing the shared page shapes so all text and widgets stay
editable was a viable route. It is, and it's now proven rather than argued: the
Quick Mortgage Payment Check page has been rebuilt the new way and is live.

How it works, in plain terms: the calculator's machinery — the panel, the input
grid, the working parts — now lives in a template that the writing agents simply
cannot touch. Every piece of visible text on it (the heading, the three input
labels, the button, the result caption, the link text) is a named field with its
own guidance, including a warning on the labels that they are load-bearing: renaming
what a number means silently changes what people believe they're calculating.

The proof was a full round trip through the real machinery, not a claim: I changed
the heading field through the framework's own editor, watched the live page update
— with the calculator untouched and its arithmetic re-checked — then reverted it,
and the page came back byte-for-byte identical to where it started. That's the
whole requirement demonstrated: text editable, widget safe, and reuse comes free
because a second page can use the same component with its own words in the fields.

One design correction along the way, worth knowing: the "permanent lock" we'd been
putting on calculator components turns out to be incompatible with editing — the
editor refuses locked components, by design. So on the new shape nothing is locked;
the protection is that the template holds the machinery and the writers can only
fill in the words. Twenty-one calculator pages remain to convert, the three odd-
shaped ones last.

---

2026-08-14, afternoon. A re-check found the record wrong, not the site. This
morning two sessions fixed the same broken calculator within six minutes of each
other: one patched it by hand, and six minutes later the session that owns the
calculator conversions rebuilt the same page properly, from the last clean copy in
the site's history. The rebuilt version is what's live, and that's the right
outcome — it is the original fix, restored through the framework rather than
hand-stitched. The calculator is correct either way: every arithmetic check
passes, including the 0% interest case that started all this.

The morning write-up, though, said the hand-patched version was still live and
quoted its test score — and both statements had stopped being true about six
minutes after they were measured. The check meant to confirm it looked for
features both versions share, so it couldn't tell them apart. I've corrected the
handoff, notes and summary, and logged the lesson: when checking "is my fix still
live", compare the actual bytes, not a marker both versions carry.

One nuance now on record so nobody "fixes" it: the restored version rounds the
monthly payment to the penny before working out the totals — the way lenders
actually bill — so six checks now read "matches the billed convention" rather
than "matches to the last decimal". That is the healthy, historical state of this
page, not a fault.

---

**2026-08-14 — sixteen calculators are now in the new editable shape, and the site's
arithmetic is fully verified clean. Getting there surfaced two mistakes of mine you
should know about.**

The good news first: sixteen of the twenty-two calculator pages now work the way you
asked — every piece of visible text is an editable field, the working parts live in a
template no writing agent can touch, and the whole site passes its complete
arithmetic check, all 170 assertions, same as before any of this started.

The two mistakes. First, my initial batch shipped all fifteen pages without their
JavaScript — I had proven the visible part byte-perfect and never noticed the
scripts travelled separately. My checks then passed the broken pages twice, for two
different bad reasons. Second, and subtler: when I rolled pages back during earlier
problems, the "restore" used a safety copy taken on the 5th of August — three days
before your stamp duty week's 0% interest fix. So restoring a page quietly
un-fixed it, and one calculator served the old wrong-at-0% behaviour for about two
days. The only thing that caught it was the full arithmetic sweep — which is exactly
why it now runs as the final gate on every batch, no exceptions. Both mistakes are
written up in the shared trap lists so no other thread repeats them, and the site is
verified clean end to end as of this entry.

Left to do here: the five odd-shaped calculator pages, and the two pages converted
under the older scheme. Then the site plan work you ruled on, which is next.

---

**2026-08-15 — the five odd-shaped calculators are done bar the publishing step.**

The five pages I flagged last time are converted. Every one of the twenty-one
calculators is now built the way you asked: all the visible words are editable fields,
and the working parts sit in a template no writing agent can reach. The five are sitting
in the database, fully checked. They are not on the live site yet, and that is the one
thing still outstanding.

Why they were awkward, in plain terms. Each of these five has a single panel where the
heading and the machinery are tangled together rather than stacked one after the other.
The tool that pulls a page apart walks *into* a panel like that looking for the seam,
and in doing so it dissolves the panel — the calculator loses its box and its inputs
stack up in one column. There was already a guard that spotted this and refused all
five, which is the guard doing its job. The answer you ruled on in August was to take
the panel whole and let the tangled words become editable fields, and that is what I
built. It is worth saying why that is safe now and was not safe in early August: back
then "whole" meant the words got locked away where nobody could edit them. Under the
new scheme the panel is a template and the words are unlocked fields, so the same
change that used to freeze copy now frees it.

The publishing step is queued behind other sites. The system deploys one site at a time
in a fixed order and this one sits near the end; when I last looked there were still
forty-eight pages ahead of it, down from eighty-eight, so it is moving, just not
quickly. Nothing is broken while it waits — the live pages carry on serving the previous,
correct version. Whoever picks this up next has one command to run when the queue
reaches us, and a written-down set of before-readings so they can prove the change
actually published rather than assume it.

Two mistakes of mine, both caught before they cost anything, both written up.

The first is the one worth your attention. I checked whether it was safe to roll these
five pages back to an old safety copy, and I got the answer completely backwards. The
page names in the database use a hyphen where the file on disk uses a slash — so I was
asking about files that have never existed. Asking git about a file that isn't there
gets you silence, and silence looks exactly like "this file hasn't changed". Five pages,
five silences, and I wrote down "all safe". In fact four of the five had changed, because
they carry the interest-rate fixes from the week before. Had I acted on that reading I
would have re-broken the same arithmetic we spent two days repairing last week. I found
it because five out of five looked too tidy to be true, not because any check caught it.
It is now filed as a standing trap with the one-line check that would have caught it.

The second is smaller and slightly embarrassing: I announced the queue had been stuck
for an hour and went hunting for a broken scheduler. The database keeps time in UTC and
this machine is on British Summer Time. The items were four minutes old.

I also found a real fault in our own checking tool while reading it. The tool that
verifies "the published page matches the source" was pointed at the *wrong* source — the
one we abandoned last week precisely because it contains reverted arithmetic. It made no
difference to these five pages, which are identical in both, but it would have mattered
on the standard loan calculator, which is exactly the page the arithmetic fix was about.
Fixed so the reference can only be set in one place and cannot drift apart again.

Left after this: publish and verify the five, then the last two pages that are still on
the older scheme — and those turn out to be two different problems rather than one, which
I have written down. Then the site-plan work you ruled on.

**2026-08-15, later — the five are live, and all twenty-one calculators are now done.**

They published at about ten to ten and I have checked them properly. Every one of the
twenty-one calculator pages now works the way you asked: all the visible words are
editable fields, the working parts sit in a template no writing agent can reach, and the
panels are intact. The whole site's arithmetic still comes out right — a hundred and
seventy checks, no failures, exactly the same as before any of this started.

I checked the publishing actually happened rather than assuming it. Before the change I
took a fingerprint of each of the five live pages; afterwards all five had changed, and
the calculator panel was present on every one. That matters because one of the obvious
ways to check would have been wrong: two of the five have no separate text section at
all, so the marker you would naturally look for is legitimately absent on them, and
anyone using it as the test would have concluded those two had failed to publish.

Then something worth telling you about. The final safety check — the one that deliberately
feeds the calculators wrong answers to confirm it can still spot a mistake — announced
that it had failed, and said in effect "the checker is asleep, eight things passed that
should have failed". That is exactly the alarm you want, so I took it seriously and looked
at all eight. It turned out to be a false alarm, and I can say why with confidence.

Seven of the eight are checks that read *words* rather than numbers — a page saying
"Option A is Cheaper", or "2 Years 3 Months". Feeding the test wrong *numbers* cannot
possibly disturb a check that is reading words, so those seven were never going to fail
and counting them as a problem was a mistake in the alarm, not in the calculators. The
eighth is nicer: the test proves itself by asserting an answer nothing could compute, and
it uses the number 100 for that. One of our test cases is a loan of £300,000 against a
property worth £300,000 — where the correct answer genuinely *is* 100%. So the test's
"impossible" number happened to be the true one, on that one case.

None of the eight was on a page I had touched. I have fixed the alarm so both kinds are
excluded and labelled, checked that the ordinary run is completely unaffected, and
confirmed the numbers now add up exactly. This is the second time this alarm has cried
wolf — it happened once before in August for a related reason — and that matters more
than it sounds: an alarm that goes off wrongly is one people learn to ignore, and this is
the last check standing between us and shipping a broken calculator.

One honest limitation, which I have written down rather than quietly fixed: that safety
check only ever tested the checks that compare *numbers*. It has never been able to test
the seven that read words, so if one of those quietly stopped working, nothing we
currently have would catch it. That is a real gap and it is now on the record.

Left to do: the last two pages still on the older scheme — and those are two different
problems rather than one, which I have written up. Then the site-plan work you ruled on.

---

2026-08-15, afternoon (second session). The last two calculators are done. These were
the two oldest-style pages — their calculator machinery was frozen solid back on
5 August as a safety measure, because back then "editable" and "safe" couldn't both be
true at once. Under the new design they can: the working parts (the arithmetic, the
buttons' wiring) live in a template no copy-editor can reach, and the visible words
(headings, labels, button text) are now editable fields, like every other calculator
on the site. The conversion was proven byte-for-byte before it touched anything —
after deploying, both pages serve exactly the same bytes as before, verified at the
live site, and every calculator on the site still passes the independent arithmetic
check (170 checks, none failing). One of the two pages had no template at all behind
it — its calculator was pasted directly into the page row — so it got a proper
template made for it. That means every one of the 23 calculators is now built the
same way, editable the same way, protected the same way. The remaining ideas on the
list: actually demonstrate reusing one calculator on a second page with different
words (nobody has done it yet, it's the cheapest proof of the reuse goal), and the
bigger "teach the planner what this site is" work.

Later the same afternoon. The reuse idea is no longer just a design claim — it has been
demonstrated and tidied away again, which is what was decided: prove it, don't leave a
stray page around. We took the repayment calculator's template, made a second copy of
just its words (different labels, different button text, same arithmetic), put it on a
hidden test page, and checked the sums on that page against our independent calculator —
all twelve checks right. Then we removed the test page through the platform's own
page-removal machinery: the page now returns "not found", the real calculator pages are
untouched, and the whole site still passes all 170 arithmetic checks. One honest wrinkle:
the database keeps a small archived record of the test page (its history table is
designed never to forget a page existed), so internal counts now say "41 active pages
plus one archived". That archived stub is actually useful — it is standing proof that one
calculator template can serve two pages with two different sets of words, which was the
owner's original ask. The one remaining big item on this site is teaching the planner
what the site looks like, so a future rebuild would recreate it faithfully.
