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
