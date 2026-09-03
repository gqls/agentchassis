# Where we are — mortgagecalculator.co.uk

Plain-prose log, append-only, newest at the bottom.

---

**2026-07-31, evening.**

Picked this up from the handoff the loanandmortgagecalculator lane wrote. Its
headline warning is real and I have now confirmed every part of it myself:
mortgagecalculator.co.uk is live and served from Backblaze, but it has never been in
the `sites` git repo. That combination is dangerous. When the platform rebuilds a
page it doesn't just upload it — it *commits* the page into that repo, and the
deploy job then makes the bucket match whatever the repo folder contains. If the
folder has just been created and holds one file, the bucket gets emptied down to
that one file and the site goes down. The job reports success while doing it.

So the first job was to get the real site into the repo, completely, before anything
can trigger that.

The previous lane couldn't finish this because it had no Backblaze credentials and
said so honestly. That has changed — the credentials are on this machine now, just
not in the place their check looked. So I could list the bucket properly instead of
guessing at it, which is what the whole thing hinged on. There are 29 real files.
The copy in `~/projects/domains/.../gemini/02` matched 28 of them exactly, which
confirms the previous lane's finding about which local folder is the true one.

Two things caught me out, and both are worth you knowing about because they are the
kind of thing that would have been discovered much later and much more expensively.

The first: the `robots.txt` file the website serves is **not** the file in storage.
Cloudflare adds a block of its own text to it as it goes out the door. The stored
file is 491 bytes; what you get in a browser is 2,327. I had been about to save the
browser version into the repo as if it were the original — which would have left the
site permanently carrying a duplicated copy of Cloudflare's text inside a file whose
entire purpose is to be read literally by search engines. The previous lane looked at
the last five lines of that file, and the added block is at the top, so they saw a
perfectly ordinary-looking file. I caught it by printing the whole thing. The fix was
to take the file out of storage directly rather than off the website.

The second is more embarrassing and more instructive. Before pushing anything I ran
a simulation of the deploy, to see exactly what it would add and delete. The
simulation printed "deletions: none, uploads: none" — which is precisely what a
completely safe deploy looks like. It was wrong. The command had actually failed
outright (I'd used an option spelling from an older version of the tool), and my way
of summarising the output turned "the command didn't run" into "the command found
nothing to do". Those two look identical on screen. I only noticed because I'd also
printed the command's exit status in the same block. I've written this up properly,
because a check that reassures you when it fails is worse than no check.

The simulation, once it actually ran, says the deploy will upload all 29 files and
delete 35 things. That sounds alarming and isn't. Thirty of those deletions are old
superseded copies that Backblaze keeps, each one paired with re-uploading the same
file; five are empty placeholder markers that no other site in the system has
either. Nothing that the website actually serves gets removed. I checked that
directly rather than reasoning about it: every file in the bucket is present in what
I'm about to push, so there is nothing for the delete step to remove.

On your instruction about fidelity — you asked for "high" rather than the strictest
setting, so the site comes in as something we can edit. I need to be straight with
you about what that word does, because the name is misleading. In the code there are
only two behaviours, not five. "locked" preserves the site exactly. **Everything
else, including "high", is the same single path: it invents a new address for every
page — `/repayment.html` becomes `/tools/repayment/index.html` — and hands every
page to an AI to rewrite from scratch, including the thirteen mortgage calculators
that were all tested working today.** The script that submits the job says so in its
own comments: high and the rest are "recorded but modulating nothing".

I put that in front of you with the code quoted and you chose "high" knowing it, so
that is what I'm building towards and I've recorded it as a decision rather than
leaving it to look like a mistake later. I do want to flag one consequence plainly:
every existing web address on that site changes, so anything linking to it, and
anything Google has indexed, points at pages that will no longer exist at those
addresses. If you'd like the old addresses to keep working, that is a redirect job
we should plan deliberately rather than discover.

Where I've stopped: the 29 files are committed into the `sites` repo on this
machine, but **I have not pushed**. Pushing is what triggers the live deploy, and
although I've simulated it and believe it's safe, it touches a live site, so I'd
rather you said go. Nothing has been sent to the cluster and the live site is
untouched.

One more thing to decide before we adopt, from the previous lane's work and still
true under "high": three pages can't be reached by following links from the
homepage, so the crawl won't find them and they won't be adopted. One is the 404
page, which is correct. The other two are real guides — one has no link to it
anywhere, and the other is linked under a filename that doesn't exist
(`mortgage-scorecard.html` vs the actual `your-mortgage-scorecard.html`, which is
also a live 404 for anyone clicking it today). Fixing those links before we adopt
means 22 pages come in instead of 20. Fixing them afterwards doesn't. It's about an
hour's work.

---

**2026-07-31, later that evening.**

Both things you approved are done and live, and the adoption is now running.

The site is in the `sites` repo. I pushed it and watched the deploy actually run
rather than trusting the green tick: the job correctly identified
mortgagecalculator.co.uk as the thing that changed, synced it, and afterwards every
one of the 29 files on the live site had exactly the same fingerprint as before.
Nothing moved. That was the point — the site now has a complete copy in the repo, so
the "platform commits one file and the deploy empties the bucket" failure can't
happen any more. That was the single most dangerous thing about this domain and it's
closed.

Then the link fixes. The homepage's "Read the Guide" button pointed at a filename
that has never existed, so it was a dead link for real visitors and it also hid the
scorecard guide from the crawler. The buy-to-let guide had no link to it from
anywhere at all. Both fixed, and I checked the result by re-walking every link on the
site from the homepage rather than eyeballing it: 22 of 23 pages are now reachable,
and the only unreachable one is the 404 page, which is correct. One file changed on
the live site; the other 28 are untouched.

I hit two mistakes worth telling you about, because both produced a confident wrong
answer rather than an obvious error.

The first: I ran a check to preview what the deploy would do, and it came back
saying the deploy would touch *every* site we own rather than just this one. That
would be a serious problem. It wasn't true. The `grep` command inside my session
isn't the standard one — it's a substitute tool with a slightly different pattern
engine, and it silently failed to match something the real one matches. The actual
deploy server uses the normal version and got it right both times. I'd been about to
report a fault in the deployment pipeline that doesn't exist.

The second: my pre-flight check asked "is another one of us already working on this
domain?" and came back with 41 jobs in progress. That looks like a clear stop sign.
It was the wrong site — `loanandmortgagecalculator.co.uk` literally contains
`mortgagecalculator.co.uk` inside it, so a "contains" search picks up both. All 41
belonged to the sister site. Our domain had nothing. I've written both of these into
the shared warnings file, because the dangerous part in each case was that the wrong
answer looked completely plausible.

The adoption itself is now submitted and running, at the "high" setting you chose.
Before firing it I had to wait a few minutes — another session was mid-way through
deploying a new version of the platform, and jobs sent during that window get thrown
away silently.

One thing I worked out while waiting, which you should know before the results land,
because it isn't quite what "rebuild the site" sounds like. The rebuild doesn't
replace pages where they are. It gives almost every page a new address —
`/repayment.html` becomes `/tools/repayment/index.html` — and **it does not delete
the old one**. The homepage is the single exception: it keeps its address, so the
new AI-written homepage lands directly on top of the current one.

So the likely end state is: a new homepage, new versions of every page at new
addresses, and all the original hand-built pages still sitting there at their old
addresses, still working, but with nothing linking to them any more. Anything Google
has indexed, and any link anyone has ever shared, points at the old set. Nothing in
the platform notices or reconciles that.

That's fixable — it's a redirect job — but it's worth deciding deliberately rather
than discovering. I'll show you what actually comes out before anything gets
published, and we can decide then whether to redirect the old addresses, keep both,
or something else.

---

**2026-08-03, early hours.**

I came back to read the classifier's output and start the positioning work. The
classifier had failed — three attempts, all of them, about nine hours earlier.

It turned out not to be anything about your site. The step that does the thinking
was being cut off mid-sentence because it had been given a budget too small to
finish the document it is asked to write. It had been running about two hundred
words under that ceiling for four months, and it finally went over. Two sites hit it
the same afternoon: yours failed all three attempts, the other one scraped through on
its second by a couple of hundred words. That is the whole difference between the two.

I checked how big that budget is compared with everything else in the system, and it
is the smallest of its kind anywhere — every comparable step has at least a third
more room, most have nearly three times as much. So I raised it, re-ran it, and it
completed first time. The proof is in the numbers rather than in the green tick: the
finished document needed about ten per cent more room than the old ceiling allowed,
which matches exactly how far short the failed attempt fell. I have written the whole
thing up as a fleet bug because it will have been quietly blocking every site
adoption, not just ours.

Two things I believed along the way turned out to be wrong, and both were the
comfortable kind of wrong.

The first was that this was a quirk of *adopted* sites like ours — the classifier is
told to preserve everything we captured from your existing site, so I assumed it had
more to write and overran. Neat story, and false: I checked all 54 runs and adopted
sites actually overrun slightly *less* often. If I had stopped at the neat story I
would have filed this as our problem instead of everyone's.

The second was more serious. The record showed someone had edited that part of the
system a few hours earlier, which on this shared setup means "another colleague is
working here, back off". It was a bulk update that touched 184 things at once — a
sweep, not a person — and it happened *after* our failures anyway. I very nearly
stood down from a fix that was mine to make.

Then the positioning. Here I have to correct the brief I was working from, and it is
worth explaining because it affects the sister sites too.

The brief said to give this site a "divergence rule" — a written statement of what
the site is *not* — mirroring one the sister site had been given. I went looking for
where that gets used and **it is used nowhere.** Nothing in the system reads it. It
is a note in a drawer. Worse, the next time the classifier runs it would throw the
note away, because it does not know that field exists.

What *does* work is the plain "who is this site for" field, and when I read the
sister site's version of that, the divergence is already written there in prose:
*"Not the single-subject researcher. A visitor who only wants a mortgage repayment
figure ... is served better by a single-subject site."* So the right thing was
already being done in the field that works — the separate "divergence rule" was
redundant. I have written ours the same way, so the three sites now each say plainly
what they cover and where to send the visitor they do not serve. Yours is mortgages
and property-secured lending; anything about personal loans or car finance points at
the other two.

I should flag one weakness rather than let you discover it. **None of this is
protected.** There is a "pin" setting on these records that looks exactly like it
would stop an automated agent overwriting your decisions, and it does not — the code
that does the overwriting never looks at it, and quietly discards it. So if the
classifier runs again it will overwrite the positioning I just wrote. I have recorded
that, and for now I have used the thing that *does* work.

Which is this: **I have put a hold on the whole site.** There is a proper lock for
exactly this purpose and, oddly, nothing on the estate was using it. The reason I
switched from holding the 24 jobs individually to holding the site outright is that
this is a chain — each stage creates the next one, and the dispatcher picks work up
every two minutes. Holding items one at a time is a race I would eventually lose,
usually at three in the morning. One switch, reversible, and nothing on your site can
move until we decide it should.

For the record, the chain runs: classify → research the market → decide a strategy →
write a brief → **plan the pages**. Everything before that last step only writes
notes. Page planning is the first point at which your live site could change, and
that is where I have drawn the line.

So: the site is intact and untouched, the classifier has finished properly, and the
positioning is set. What I need from you is the go-ahead on how to release the build.
My recommendation is to let one page rebuild first — deliberately *not* the homepage,
because the homepage is the only page that keeps its address and therefore the only
one that overwrites something live. We look at what comes out, and then you decide
whether to run the rest.

---

**2026-08-03, later.**

Three fixes and one experiment. Taking the fixes first.

I raised that token budget again, from where I'd put it to about five times what the
job actually needs, so we're not back here in a month. The site lock now genuinely
locks. And six of your guide pages had a broken link I should have spotted earlier:
clicking the logo in the header took you to a page that doesn't exist. One missing
"../" in the address — the line directly underneath it was already correct. Fixed and
live, and I re-checked all 29 files afterwards to be sure nothing else moved.

On the lock: I proved it properly this time rather than just re-reading the setting.
I ran the system's own "which site should I build next?" question twice against the
same live data — once the way it used to ask it, once the way it asks it now. The old
way picks **your site, first in the queue**. The new way skips it. So without that fix
your site would have been rebuilding itself while we were talking.

Then the experiment: I built one page — the first-time buyer guide — deliberately not
the homepage. It worked mechanically: new page live at its new address, old page still
serving, homepage untouched throughout.

But I have to own a mistake in how I ran it. **I built a page before the site had a
stylesheet.** When I froze everything last night, two of the jobs I froze were "work
out the site's colours and layout" and "generate the site stylesheet". So the page
came out with no styling and no menus, and my first reaction was that the rebuild
produces broken orphan pages. That would have been a serious conclusion and it was
wrong. I checked a page on the sister site built by the same machinery — it has its
menus, and its stylesheet loads fine. The tooling is capable; I ran the steps out of
order. The right sequence is design first, then pages, and I'll do it that way.

Two links on the new page also point at pages that don't exist yet. I nearly wrote
those up as invented, but they're both real pages the system has planned and simply
hasn't built. Not a fault.

**One thing is a genuine fault**, and it's the useful result from the experiment.
The opening line of the new page reads: *"Banks evaluate your application using a
\*\*Decision Engine\*\*"* — with the asterisks visible. The writer used a formatting
convention the page doesn't understand, and nothing catches it. It's on two other
sites of ours as well. It's small, but the reason it matters is that every automatic
check passes it: valid page, complete content, marked as successfully deployed. It was
only found because someone read the sentence. Written up so we can add a check.

Where that leaves us: your site is locked and, this time, actually held. The live site
is exactly as it was apart from the six link fixes. Nothing is queued that can move
without you.

What I'd suggest next is a proper first pass rather than another single page: let the
composition and stylesheet jobs run, then rebuild two or three pages so you can judge
them looking as they're meant to look. Still nothing near the homepage until you've
seen that and said go.

---

## 3 August, around midday — the test page came back looking like a real page

Short version: it worked, and the thing we were unsure about is now settled.

You'll remember the first rebuilt page — the first-time-buyer guide — came out with
no header, no navigation and no footer. It was a bare column of text. I'd said at the
time that I thought the cause was ordering: I'd built the page before the site had
its styling and structure in place, and a page gets its layout baked in at the moment
it's built, so giving it a stylesheet afterwards doesn't retro-fit the missing parts.

That was the theory. It's now confirmed. I rebuilt that one page and it came back
with a proper header, navigation and footer, and the styling resolving. It went from
about 8,900 characters to about 20,500. It went live at 11:06 this morning. You can
look at it:

  https://mortgagecalculator.co.uk/guides/first-time-buyer/index.html

So the order is: structure and styling first, then pages. That's now proven rather
than assumed, which means the remaining pages can be rebuilt as a batch when you're
ready, and they should come out looking right first time.

**Your live site is untouched.** I re-checked every file against the copy in the safety
repo after making the change, and everything is byte-for-byte identical except
`robots.txt` (which Cloudflare rewrites — expected, and covered before) and the test
page itself, which is the change I meant to make. The site is locked again and nothing
is queued that can move on its own.

### What was actually wrong

Worth recording because it wasn't what I first thought. Chrome — the header, nav and
footer — is stored in one table, and your site simply had no rows in it. A sibling
site had three. That was the entire difference. There's a job that creates those rows
and it had never run for us, because I'd deferred the jobs ahead of it. I ran it, and
it produced the header, footer and page-head in one go.

Along the way I nearly misread this. There are three columns on the pages table with
promising names like "rendered_header", and they were all empty for your site. That
looks like a smoking gun. They're empty for **every page on every site we own** —
they're leftovers that nothing writes any more. If I'd only looked at your site I'd
have "fixed" a column that does nothing.

### Two things you should see before we do the rest

**One: the navigation currently has a single link — Home.** This is deliberate and I
think correct, but you should know about it. The system refuses to put a link in the
header if the page it points at hasn't been built yet, because the header appears on
every page and a dead link would then be everywhere. Right now only one page is
built, so there's almost nothing legitimate to link to. As we build the rest, the
navigation fills in by itself. It's not something I need to fix, but the test page
will look sparse until there are more pages, and I didn't want you to judge it as
broken when it's actually being careful.

**Two: I found a real fault, and it's small but it will spread.** The "Get Started"
button in the header points at a page that doesn't exist yet, so it's a 404 right
now. The navigation links are checked against "has this page actually been built" —
but the button next to them is checked against a different, looser test that lets an
unbuilt page through. Two checks, same function, a few lines apart, different rules.
Because the header goes on every page, this one broken button would appear on all of
them once we build the batch.

I should be straight about one thing here: I first measured this against our records
and concluded two of our sites were affected, and I was about to tell you that. Then
I actually fetched the other site's button target and it works fine — our records say
"never deployed" but the page serves perfectly well. So the honest answer is that the
faulty logic is real, and the confirmed broken button is on this site only. I'd rather
give you the smaller true number than the bigger one I hadn't checked.

Nothing about that blocks the batch, and it'll want fixing before we build the rest,
or that button ships broken onto every page.

### Where that leaves your decisions

Unchanged, and both still yours:

- **The homepage.** Still the only page whose address doesn't change, so it's the only
  one that would overwrite what's live. Still deferred, still needs your go-ahead.
- **The old addresses.** 22 of the 23 original pages move to new addresses when
  rebuilt, and the old files keep serving alongside them. Nothing reconciles that
  automatically.

What I'd suggest: have a look at the test page now it's styled. If it reads right to
you, the sensible next step is a batch of two or three more guides — not the homepage —
so you can see them as a set with the navigation filling in. I'd fix the Get Started
button first so it doesn't ship broken onto each one.

---

## 4 August, late evening — your homepage was rebuilt overnight, and I've put it back

I need to tell you something went wrong, and then that it's fixed.

**Your homepage was replaced.** Last night at 19:45 the platform rebuilt `/index.html`
and deployed it over the live original — the one page we had both agreed nobody would
touch until you'd seen a styled rebuild and said go. You have it back now: I restored the
original at your instruction and checked it at the wire. It's the file you had, byte for
byte, 28 links and all.

**Nothing else was affected.** I checked all 33 files of your site against the safety
copy: everything matches except `robots.txt`, which Cloudflare rewrites and always shows
as different. Your calculators — repayment, affordability, simple, overpayment — were
never touched and are serving fine.

### Why the lock didn't stop it, which is the part I got wrong

I told you the site was locked and that nothing could move without you. The lock was on
the whole time — it never lapsed, and I re-checked it. But it turns out the lock only
controls the *queue*: the scheduler that picks sites and works through their to-do lists.
It does not stop a job that's fired at the site directly.

That distinction matters because firing directly is exactly what I was doing all through
yesterday — it's how I rebuilt the test page while keeping everything else frozen. I used
that door deliberately, and then told you the building was locked. Another session, fixing
the very bug I'd filed about your Get Started button, needed a live site to test against.
Yours was the example named in the bug report, so they used it, and the homepage came along
with it.

I've corrected the handoff notes so the next person doesn't inherit the same wrong belief,
and written it up as a mistake rather than quietly fixing it, because "locked" reading as
"safe" is the kind of thing that will catch someone else.

### One thing worth knowing about the rebuilt version

It wasn't broken, and it's worth saying why I still recommended putting the old one back.

The rebuild was clean — properly styled, no jargon glitches, and the broken "Get Started"
button I flagged yesterday was **gone**, which is the proof that the fix works. But its
internal links dropped from 28 to 4. The system refuses to link to pages that haven't been
built yet, and most of your site hasn't been rebuilt. So the front door stopped pointing at
your calculators. They still worked, but a visitor landing on the homepage couldn't find
them.

That's the same lesson as yesterday, one level up: **the homepage can only be rebuilt after
enough of the site exists for it to link to.** It's not a fault in the rebuild, it's an
ordering problem, and it's a good argument for leaving the homepage until last rather than
first.

### Where that leaves us

Your site is exactly as it was. The good news from yesterday still stands: the ordering is
proven, the test guide looks right, and the button fault I found has been fixed properly by
another session and is live.

I haven't started the batch of two or three guides yet — finding the homepage changed took
priority. That's the next thing, and it's unaffected by any of this, because those pages go
to new addresses and can't overwrite anything of yours.

One honest caveat I'd rather you heard from me: I can hold the queue, but I can't stop
another session firing at your site directly, and neither can the lock. What I can do is
keep a restore point and check the pages afterwards — which is what caught this.

---

## 5 August — three more guides live, and your calculators now have a proper referee

Two pieces of good news, then the state of play.

**The three guides you approved are live** — remortgaging, buy-to-let and negative equity,
all styled, all with proper navigation, all at new addresses so your original pages are
untouched and still serving. That's four rebuilt pages now including the test one, and
they look like a set.

**The bigger news is about your calculators.** You asked whether we could have something
that checks the arithmetic before the rebuild touches them, and told me to search for
anything that already did this before building it. That search paid off twice over.

First: the checker already existed. Another of our site teams built exactly this a week
ago — a way of recording what a calculator answers for a fixed set of inputs, down to the
penny and the pound sign, and then having the platform re-ask those questions on a
schedule forever. If a rebuilt calculator's repayment figure drifts by a fraction, it
fails loudly. I verified the checking machinery is actually live in the platform rather
than trusting the docs, and it is. So no new agent was needed — building one would have
duplicated a tested, approved mechanism.

Second: when I ran it against your twelve calculators, it refused to certify one of them —
the investor page — claiming its arithmetic "ignores its inputs". **Your calculator was
right and the checker was wrong.** That page computes ratios — rental yield, loan-to-value
— and the checker varied its test inputs by doubling everything at once. Double a rent and
double a price and the yield is identical; that's what a ratio is. The checker read
"answer never changes" as "broken". I fixed the checker itself (it now also varies fields
by different amounts, which a ratio does respond to), proved the fix didn't disturb the
other team's twelve calculators (all twelve still match their recorded answers exactly),
and re-ran ours: all twelve now certified, answers recorded.

One honest footnote: the recorded answers prove a rebuild computes *what your originals
compute* — not that your originals were right in the first place. They're your trusted
code, so that's the correct baseline, but it's worth saying once.

**The rebuild of the twelve calculators is running now.** Each new version goes to a new
address — nothing overwrites your working pages — and each will be compared, number for
number, against the recorded answers before we treat it as done. One detail found during
recording: your original calculate buttons have no internal names, which the enforcement
system needs, so the rebuilt versions will gain them — that's the one deliberate
difference, and it's invisible to visitors.

Nothing needs a decision from you right now. The homepage remains yours and untouched.

---

**2026-08-08 — why three calculators came back empty, and what's been done about it.**

You'll remember twelve calculator rebuilds ran, and three of them produced nothing at
all — the overpayment calculator, the portfolio dashboard, and the fact-finder game.
I've now got to the bottom of all three, and none of them failed because the rebuild
itself went wrong.

Two of them were killed by our own quality checker. It scans a finished page for
leftover template text — things like "[Name]" that an AI sometimes leaves behind —
but it was reading the page's program code as well as its visible text. Calculator
code is full of harmless phrases that happen to start the same way, so the checker
declared two perfectly good calculators "unfinished", threw them away, and — worse —
still marked the job as done, which is why nothing flagged it at the time. The same
thing bit one other site the same day. I've fixed the checker so it only reads what a
visitor would actually see, proved the fix against the exact three lines it wrongly
convicted (and proved it still catches real leftover template text), and put the
change through the review council. It's committed and will take effect on the next
platform release — the three failed rebuilds should then be re-run.

The third, the portfolio dashboard, was stopped by a different guard — one that
checks a rebuilt tool hasn't invented fake data (a real incident on another site a
few weeks back). Your original portfolio page doesn't load any data — visitors type
in their own properties — so I suspect this was also over-caution, but the record of
exactly what it objected to has since been cleaned away, so the honest answer is:
re-run it and watch. That's queued for after the checker fix lands.

Separately, while digging, I found the "mark the job done even though the output was
thrown away" behaviour is its own defect — the workshop's paperwork says a failed
check should still hand the work over, but the live settings lost that instruction in
a way nobody can now trace. I've written that up for a deliberate decision rather
than quietly changing it, because "what should happen when a rebuilt tool fails its
checks" is a judgement call, not a typo.

The site itself is untouched by all of this: still locked, your original pages all
verified byte-for-byte again, the nine live rebuilt calculators still serving. The
arithmetic verification of those nine (the ID-alignment work) is the next job after
the three re-runs.

**2026-08-08, evening — all twelve tools are now live.** The checker fix shipped in
this afternoon's platform release (I verified the running system actually carries it,
not just that a release happened), and I re-ran the three stalled rebuilds. All three
came back clean. The overpayment calculator and the fact-finder quiz — the two that
the faulty checker had wrongly convicted — passed validation this time, which is the
fix proving itself on the exact cases that motivated it. The portfolio dashboard,
the one stopped by the fake-data guard, passed that guard cleanly this time with
nothing flagged at all; whether its earlier conviction was genuine we'll never know
(the record was purged before anyone could read it), but the version now live was
checked and passed. There's one open loose end from that earlier conviction — a
"needs a human to look at this" ticket from the 5th that's now about a version of
the page that no longer exists. It's harmless, but closing it is your call rather
than mine.

I verified the results the hard way: all three pages serve on the live site with the
site's own header and footer, each one contains the right calculator (a mix-up
between pages was a real risk we'd seen signs of), and a byte-for-byte comparison of
the entire live site against our copy shows the only difference anywhere is the
robots.txt file Cloudflare rewrites — meaning your original pages are all still
exactly as they were. The site is locked again.

One wrinkle worth knowing about: the overpayment rebuild's paperwork says "completed
with errors" — the deploy itself worked, but the internal message saying "done"
failed to get back to the workflow that sent it, a known intermittent fault in the
platform's plumbing that we have on the books already. The page is live and correct;
only the paperwork grumbled.

Also corrected today: the write-up I'd done of the "failed check throws the work
away" defect had the mechanism wrong. The lost-instruction theory didn't survive
contact with the code — the instruction is there and the system does read it; the
real problem is that the failure path hands over to a step that then can't find the
work (it looks in a box that's only filled on success) and shrugs "nothing to save"
as if that were a success. Same behaviour seen from outside, different — and
actually simpler — thing to fix. The decision about what SHOULD happen on a genuine
failure is still yours to make; nothing has been quietly changed.

Next: the ID-alignment pass across all twelve calculators, then the arithmetic
verification — none of the twelve is yet proven to compute the same answers as your
originals, and that remains the biggest open item on this adoption.

**2026-08-08, late evening — the calculators now speak the same language as your
originals, and the first honest comparison found real differences to fix.** After
this afternoon's twelve went live, I ran a second pass asking each rebuild to keep
the exact internal names your original calculators use — that's what lets our
checker drive both versions side by side and compare the actual numbers. Nine of the
twelve took the change and are live; three were stopped by the platform's own safety
guards and kept their earlier versions (nothing was lost or broken). One of those
three was stopped by the fake-data guard again — and this time I caught it in the
act: it convicted a comment in the code that says "no fabricated data — starts
empty". The guard read the words "fabricated data" and couldn't see the "no". I've
written that up as a platform bug with the exact evidence; the same guard probably
convicted the same tool the same way on the 5th.

The real news is what the side-by-side comparison found now that it can finally
compare like with like: six of the calculators get DIFFERENT ANSWERS from your
originals — not because anyone's arithmetic is sloppy, but because the rebuilds
compute the standard textbook way while your originals have their own way of
working things out (for example, your repayment calculator answers £1,390 a month
where a textbook calculation says £1,169). For adoption, your originals are the
contract — the rebuilds must match them, and the next pass will instruct each
rebuild to copy the original's calculation method exactly rather than treating it
as inspiration. One happy resolution: the stamp-duty calculator's worrying
"£0 result" turned out to be our test pressing the wrong buyer-type option because
the rebuild had reordered the drop-down — its actual sums check out correctly for
every band we drove. The whole site was byte-verified against our copy again
tonight; your original pages remain untouched and the site is locked.

**2026-08-08, night — correction to the entry above, and it's good news: the
rebuilt calculators are NOT getting wrong answers.** Your question — "explain why
it's all different" — made me check by hand, and the checking overturned my earlier
report. The comparison tool I used has a quirk I hadn't appreciated: it doesn't feed
both calculators the same numbers. It reads whatever starting values each page
happens to ship with and drives multiples of those — so your original repayment
calculator was tested with £250,000 at 4.5%, while the rebuild (which ships
£200,000 at 5% as its example values) was tested with those instead. Both answered
their own question correctly; I mistook the two different questions for six broken
calculators. When I drove the rebuilt repayment calculator by hand with your
original's exact inputs, it answered £1,389.58 against your original's £1,390 —
the same answer, to the pound, just shown with pence. The other "differences"
dissolve the same way (one rebuild ships no example values at all, so the test
poured meaningless 1000s into every box — that's the "1200% yield").

Two real things survive: the bridging-loan calculator genuinely models the loan
differently from your original (your original uses the retained-interest structure
bridging lenders actually quote; the rebuild compounds it another way) — and per
your ruling, the improvement loop should decide which is right rather than blindly
copying either. And the comparison tool itself needs the fix implied above: replay
the exact same inputs on both sides, which it already records but doesn't yet use.
That's queued as the first job for the next session, because per your direction the
checker's job is to prove results don't differ on identical inputs.

Also actioned from your message: the site is UNLOCKED, and the stance changes —
correctness beats faithfulness to the originals, the improvement loops own tool
quality, and everything runs from the framework. The handoff for the next thread
says all of this.

**2026-08-08, night (continued) — your "supply both calculators" direction is
recorded and is now the plan for the bridging-loan case.** Where two calculators
are right in different ways — as with bridging, where your original quotes the
retained-interest structure lenders actually use and the rebuild compounds the
interest another way — we won't pick a silent winner. The main tool page keeps
one model, and the other becomes its own clearly signposted page that explains
the difference and who each version is for. That page will be built through the
framework like everything else, so the improvement loops can keep it honest.
First, though, this session is fixing the comparison tool so it feeds both
calculators identical inputs — that's what tells us which tools agree already
and which genuinely differ.

**2026-08-08, late night — the comparison tool now asks both calculators the
same question, and the answers are in.** I rebuilt the checker so it feeds every
rebuilt calculator the exact same numbers your originals were tested with, and
then compared the answers. The headline: not one rebuilt calculator gets a wrong
answer. Six of the nine testable tools agree with your originals outright —
sometimes to the penny where your originals round to the pound, and two of the
"differences" turned out to be my test's own blind spot (your investor and
equity-release pages each have two Calculate buttons, and the test only ever
pressed the first one, so it recorded zeros your originals never actually show
a user).

Three tools genuinely differ, and in each case both sides are right about
different things — exactly the situation your "supply both, well signposted"
direction covers. The bridging calculator: yours quotes the retained-interest
structure lenders use; the rebuild compounds the interest. The rate forecaster:
yours models rates changing over the life of the mortgage (a genuinely cleverer
model — I checked it to the penny); the rebuild answers the simpler "what if my
rate were X" question. The fee analyser: yours counts every pound leaving your
pocket during the deal; the rebuild counts only the true cost — interest and
fees — treating repaid principal as money you keep. All three are queued for the
both-calculators treatment.

One genuine error was found — in the original, not the rebuild. Your stamp duty
calculator gives first-time buyers relief on homes priced between £500,000 and
£625,000, but under the rules in force since April 2025 the relief disappears
entirely above £500,000. On a £595,000 home yours quotes £14,750 where the
correct bill is £19,750 — a £5,000 under-quote. The rebuild gets this right. Per
your ruling we improve past the original rather than copying its mistake — the
rebuild already does.

**2026-08-09 — the three both-ways calculators are in the improvement queue,
and your legislation question has a pleasing answer.** Five jobs are filed and
live: the bridging page switches to the retained-interest structure lenders
quote, with a new companion page for the compound-interest way of charging;
the rate forecaster adopts your original's cleverer over-time model, with a
new companion page for simple what-if comparisons; and the fee analyser will
show both cost figures side by side — the true cost of borrowing and the total
cash out the door — each with a one-line explanation. Every job carries a
worked example the new calculator must reproduce exactly, so a wrong
implementation fails loudly instead of looking plausible.

On legislation: we don't need to build a watcher — the platform already has
one. Every day it re-checks each site's registered facts against their
official sources, re-fetching the cited page and confirming the exact quoted
sentence still appears; if the wording moves or vanishes, it raises a review
item. What mortgagecalculator.co.uk was missing was any registered facts for
it to watch. That's fixed: the site now carries the stamp duty rules — the
standard bands, the first-time-buyer thresholds, the £500,000 cliff your
original calculator missed, and the additional-property surcharge — each
quoting GOV.UK verbatim. From tomorrow the platform checks them daily, so
when the Treasury next moves the thresholds, we hear about it from the
machine rather than from a wrong calculator.

Publishing the rules on the site, as you suggested, is a good idea and the
natural next step: a "current stamp duty rates" page whose numbers can only
come from those registered facts, so it stays correct by construction. That's
queued behind one mechanical check about how new guide pages get created.
The remaining gap, being designed properly rather than rushed: connecting the
registered facts to the calculators themselves, so a tool that encodes an
out-of-date threshold fails its acceptance check automatically.

---

**2026-08-09 (afternoon) — the five jobs built, and the facts-to-calculators
design is written up.**

First, the good news you don't have to wait for: all five of this morning's
jobs finished by quarter past eleven. The bridging page, the rate forecaster
and the fee analyser have been rebuilt on the models we agreed, and the two
new companion pages — the compound-interest bridging calculator and the rate
scenario comparison — exist. I have not yet re-driven them against the
originals to confirm the numbers now match; that is the next practical job and
it is what will tell us whether the rebuilds actually landed the models or
merely claim to.

Second, the design you asked for: connecting the registered facts to the
calculators. I spent this session measuring what the platform already does
rather than sketching, and the picture is sharper than expected — and slightly
worse in one place, better in another.

Worse: our own twelve rebuilt calculators are, right now, invisible to the
platform's tool-checking machinery altogether. Not failing it — not in it. The
checks are attached to a per-tool planning document, and the rebuilt tools
never got one, so they have never had a single automated check run against
them in their lives. The two brand-new companion pages built this morning
*do* have one, because they came through a different route. So before any of
the fact-checking work can bite here, those twelve need their documents
created. That is a small job and it turns on the existing checks too, which is
a straight win regardless.

Better: the most valuable single change turns out to need no code at all. The
agent that rebuilds a calculator is already handed the site's registered facts
— it has been all along — and simply is not shown them in its instructions. So
it invents its constants, or copies whatever a human typed into the job spec,
while the correct, source-cited numbers sit unused in its own context. Telling
it to read them is a configuration change, live the moment it is applied. That
does not *guarantee* a correct calculator — a model shown a fact can still
ignore it — so it does not replace the checking work, but it is the cheapest
improvement available and it makes every later fix a rebuild rather than a patch.

The rest of the design is four pieces in order: show the builder the facts;
let each tool declare which facts it encodes; make a fact that changes reach
the tools that encode it; and finally compute the expected answers from the
register itself, so a calculator running a superseded threshold fails
automatically. That last piece is the one that genuinely closes the door, and
it is the one that needs a proper architecture review — because it changes
what a passing check *means*, from "this calculator still does what it did"
to "this calculator agrees with the published rules". Those are different
promises and the platform currently only makes the first.

One caution I want on the record. All of this rests on the registered facts
being right. The daily check proves that GOV.UK still says what we quoted —
that is provenance, not correctness. It cannot tell a correct threshold from a
confidently wrong one. So before we let the register start overruling the
arithmetic in the calculators, those stamp duty facts deserve a human read.

Full design, with the measurements and the traps:
`PLAN_2026-08-09_facts_into_tool_acceptance.md`.

---

**2026-08-10 — the daily legislation check has now actually run, and the three
rebuilt calculators are right.**

Three good pieces of news and one honest caveat.

First, the legislation watch is no longer just switched on — it has run and
worked. At two minutes past ten this morning the platform re-fetched the GOV.UK
stamp duty page, found all four of our quoted sentences still there word for
word, and stamped each fact as re-verified. That matters more than it sounds:
the risk on day one was never that the law had changed overnight, it was that
our own quotes wouldn't survive the platform's way of reading the page, and
we'd get four false alarms. We got none. So from here, an alarm means something.

Second, the three calculators we agreed to rebuild have landed, and I checked
them against the originals rather than taking the pipeline's word for it. The
bridging loan calculator now matches the original exactly. The rate forecaster
reproduces the original's cleverer over-time model to the penny. The fee
analyser does what you asked for — both cost figures on one page — and I drove
the live page to read them: the new figure agrees with the original's £26,841.44
to the penny, and the second, stricter figure comes out at £17,384.79 exactly as
specified. Nothing regressed.

One of those checks nearly went wrong in an instructive way. My comparison tool
only compares numbers that appear on *both* the old and new pages. The fee
analyser was deliberately built to show an *extra* number — so the tool marked
it as a disagreement, because the figure that agrees with the original is the
new one it can't see. Any calculator we improve by adding an output will look
broken to that comparison. I've written the warning down where the next person
will hit it.

Third, the cheapest piece of the facts-into-calculators work is now live. The
agent that rebuilds a calculator is told, in its instructions, that a registered
fact overrides both the original code and the specification — and it is handed
the current stamp duty rules to work from. Before applying it I checked the
thing that could have gone badly wrong: six of our sites have no registered
facts at all, and a clumsy change here would have broken calculator rebuilds on
all of them. I tested that case explicitly and it behaves exactly as before.

The caveat: that last change is applied but not yet *proven*. Telling a model to
prefer a fact is not the same as watching it do so. The next job is to rebuild
one calculator and read the code it produces, checking it uses £500,000 rather
than the old £625,000. Until that's done it's a reasonable expectation, not a
result, and I'd rather say so than let it read as finished.

Still open and unchanged: the stamp duty calculator's dropdown values need
pinning before it can be checked automatically; twelve of our calculators have
no checking document at all, which is why none of them has ever had an automated
check run against it; and the big piece — computing expected answers from the
registered facts themselves — still needs its architecture review.

---

**2026-08-10, evening.** The proof I promised in the last entry is done, but not
the test I said I'd run — because when I looked properly, that test couldn't
have failed.

The plan was: rebuild the stamp duty calculator and check it uses £500,000
rather than the old £625,000. So I read the calculator we already had, built
two days ago, before any of this. It already used £500,000. The model simply
knows the stamp duty rules. So the test would have come back green whether or
not our registered facts had reached it — it would have measured the model's
memory, not our register. A check that can only come out one way isn't a check.

What can only come from our register is our own *wording*. So that became the
test: does the calculator's code quote the sentences we wrote when we registered
each fact? The old version: not once, anywhere. The new one: every one of them,
sitting as a comment beside the number it licenses. That is the register
reaching the calculator, and nothing else explains it.

Then the rebuild taught me something I didn't ask it to. It **dropped a rule**.
The old calculator knew that the additional-property surcharge only applies to
purchases of £40,000 or more; the new one applied the surcharge at any price.
Nothing failed, nothing was flagged, and the tool was quietly wrong at the
bottom end. The cause is the instruction we ourselves added: *don't state a rule
that isn't in the register*. The £40,000 rule was real law that we had never got
round to registering — we'd registered four stamp duty facts, and that wasn't
one of them.

**So the register cuts both ways: what it leaves out can be taken away.** Every
register is incomplete, so this matters beyond one number. I've written it up
where any session touching a register will read it before, not after.

The fix was the obvious one, and it doubles as the strongest evidence yet that
any of this works. I re-registered the stamp duty rules properly — four facts
became thirteen, one for each band edge and each rate, each quoting the exact
sentence from GOV.UK, including the £40,000 rule — and then rebuilt the same
calculator from a specification I did not change by a single character. The
£40,000 rule came back, with our sentence beside it, and the code actually uses
it rather than just declaring it. Change the register, and the calculator
changes. That is the thing we've been trying to build.

Two smaller results. The dropdown values are pinned now, so the automated
comparison can finally drive the stamp duty calculator: it reports that on a
£595,000 first-time purchase the original site says £14,750 and ours says
£19,750 — ours is right, and the original under-quotes by £5,000. That finding
is still waiting on you. And on the mechanics: quotes from GOV.UK are now
lifted out of the fetched page by machine and checked with the same code the
daily watch uses, rather than typed by me and hoped over.

One outage worth knowing about, since it wasn't ours and it cleared: for about
two hours this afternoon every AI call across the whole fleet failed on an
account usage limit whose message said access returns on 1 September. It came
back at ten past seven. It had already been written down elsewhere as a
three-week outage, and I've corrected that, because other people are reading it
and planning around it.

Still open, unchanged: twelve calculators have no checking document, so none has
ever had an automated check run; and the big piece — computing the expected
answers from the registered facts themselves — still needs its architecture
review before it's built.

---

**2026-08-10, later that evening — the calculators now have checking documents,
and finding out what a checking document actually is took most of the work.**

The open item at the end of the last entry was "twelve calculators have no
checking document, so none has ever had an automated check run". Eight of them
have one now. The other four turn out not to be eligible for the checking
system at all, or not to have anything honest to put in one — I'll come to
those, because they're the more interesting half.

The job sounds mechanical: drive each calculator, write down what it answers,
save that as the document the platform checks against in future. It is not
mechanical, and the reason is the thing this whole lane exists to fix. **What a
calculator currently prints is not the same as what it ought to print.** If you
record today's answers and call them the expected answers, then the day the
calculator is rewritten wrongly, the check goes green — and worse, the day
someone notices it was always wrong, the checking document is sitting there
vouching for it. That is precisely how the original stamp duty calculator ran an
expired tax rule for sixteen months with every check we owned passing it.

So nothing went into a checking document until it had been **recomputed from
somewhere that is not the calculator's own code**. Eighty numbers, three
different kinds of "somewhere else", and I've kept them labelled separately
rather than blurring them into one word like "verified":

- Fifty-six come from the published formula — the standard mortgage repayment
  identity, compound interest, running an amortisation schedule month by month.
- Four come from **our own evidence register**: the stamp duty bands are built
  out of the thirteen facts we registered this morning, each one quoting its
  sentence from GOV.UK and re-checked every day. Not from a second copy of the
  tax table typed into a script.
- Twenty come from the calculator's own design decisions — how long the fixed
  period is before the rate changes, what it counts as "total cost". These are
  weaker and I've said so in the file. They'll catch someone breaking the sum;
  they can't catch a design decision that was wrong from the start.

All eighty agree. But eighty agreements prove nothing on their own, so there's a
control: I changed the register's first-time-buyer cap back to the **old,
expired** £625,000 figure and re-ran. The expected answer for a £595,000
purchase immediately became **£14,750 — the original site's wrong number, to the
pound.** That is the run that shows the register is genuinely driving these
expectations rather than sitting decoratively beside them. Put the expired law
back in and the expired answer comes back out.

**Four things I'd been told, or had assumed, were wrong — all in the direction
of making the job look easier than it was.**

The first would have been invisible. The platform doesn't look up a checking
document by the page's name; it strips a prefix and looks up a shortened name.
Had I filed these under the obvious name, the system would have gone on
reporting "this calculator has no checking document" for ever, with no error
anywhere, and it would have looked exactly like not having done the work.
I found the rule in the code first — and then found the platform's own notes
confirming it, filed days ago under the very names I'd chosen.

The second: three of the twelve can't be checked by this system at all, for
structural reasons (one page has two content blocks instead of one, two aren't
classed as tool pages). Writing documents for them would have produced rows that
look like coverage and are never read.

The third is the one I'd want you to know about even though nothing went wrong.
Switching a calculator's checking document on also switches on a second, older
check — and that one can fail a page for reasons the document says nothing
about, then hand the page to an automated rewriter. The block it would hand over
is shared by **252 pages across 18 other sites**. Our neighbouring lane had
written down a reason why this can't happen; I checked it against the code and
the reason is incomplete. So before installing anything I ran those checks
against all twelve live pages using the platform's own code, and all twelve
pass — then deliberately fed it a broken page to confirm the check can actually
fail, because twelve passes from a test that has never failed aren't worth much.
Nothing is at risk today. But what's holding it off is that the pages happen to
be clean, not a guard, and I've written that down where it will be read.

The fourth: one calculator, the portfolio one, got no document. The capture tool
fills each field by scaling the page's own default value, and that form has no
defaults, so it typed a mortgage term of a thousand years into every test. The
calculator quite correctly refused. Every "answer" it recorded was the error
message. A checking document built from that would spend the next year proving
our error message still works.

**And one thing I got wrong myself, worth telling you because of how it looked.**
My checker reported one calculator wrong by £1,923. It was not; my own number
parser was. The page writes a negative amount with the minus sign *outside* the
pound sign, and my code was looking for a minus attached to a digit, so it read
a fall of £961 as a rise of £961. It only came to light because the independent
formula disagreed. Had I used the same faulty parser on both sides, it would
have agreed with itself and quietly written a sign error into the permanent
record.

**Where this leaves us.** Eight calculators now have checking documents, and I
filed a real check run against the stamp duty one to prove the documents actually
execute rather than just existing. But there's a catch I want to flag rather than
paper over: **seven of the eight are invisible to the scheduler.** They serve
perfectly well — I fetched every one — but their database rows never got stamped
with a deploy time, and the scheduler skips anything unstamped. It's a known
condition; the platform's own source notes it and names nine pages on this site
as the example. I have not gone in and stamped them, because that would be
recording a deploy event I didn't witness, on rows sitting in a queue that may
belong to another lane. It's a small decision but it's yours: either we stamp
them, or every check on this site has to be filed by hand.

**Later still, 08-10 — all eight passed.**

I fired every one of the eight checking documents at the real system rather than
leaving them sitting there looking installed. All eight came back green, between
19:05 and 19:16, four checks each, in a real headless browser against the live
pages. Nothing was flagged, and nothing was handed to any automated rewriter —
I re-checked that afterwards as well as before.

Two of those greens are worth a sentence each. The rate forecaster one proves the
plumbing end to end: it asserts an amount with a proper typographic minus sign,
and that character had to survive being written by a script, stored as JSON,
pushed through the database, pulled back out, sent across the message bus and
compared letter-for-letter in a browser. It did. And the equity release one
passed where the neighbouring site's equivalent calculator *failed* this morning
— their test was reusing state from an earlier step and reported someone aged 130.
Ours sets every field explicitly each time, so it can't inherit.

The open question from the last entry stands unchanged: seven of the eight are
invisible to the scheduler, so today they only run when someone asks. That's the
one thing I'd like a steer on.

**2026-08-11, afternoon — the equity release change you asked for is filed, with
one correction to how it was written down this morning.**

I've filed the rebuild request for the equity release calculator with the
original page's age table written into the instructions, exactly as decided this
morning. But one thing in this morning's note was recorded the wrong way round,
and you should know about it: the two numbers were swapped. The ORIGINAL page
gives £124,000 for a 65-year-old with a £400,000 home — its table says 31% at
age 65. It's our REBUILT version that currently says £120,000, using a
straight-line formula the generator made up. This morning's note had those
figures attached to the opposite sides.

"Match the original" therefore means the calculator will now show £124,000 in
that example, not £120,000. I went ahead because the instruction itself was
clear — use the original's table — and that's what has been filed. But if what
you actually wanted was the £120,000 figure itself (say, because you know it to
be closer to current lender policy), tell us and we'll put the straight-line
version back; it's one small change and nothing else depends on it. Worth
remembering: no visitor reaches these rebuilt pages yet anyway — the homepage
still links to the original versions.

One more small thing: the original page's own comment claims "about 30% at 65"
while its code uses 31%, so the page disagrees with itself — that's probably
where the mix-up started.

**Later the same afternoon — done, end to end, and checked.**

The equity release calculator has been rebuilt and is live with the original's
age table, and its checking document has been rewritten to match and run: all
four checks passed in a real browser. The safety net now covers the maximum
cash figure for the first time (it couldn't before, because the old rebuild's
formula was invented), and it also now checks that anyone under 55 is politely
refused. We proved the checker can still say no by running it against the old
page's numbers — it flagged exactly the three cases where the two formulas
disagree and nothing else.

Two things you may care about. First, the automatic picture-review of the
passing page spotted that the Calculate button's label is nearly invisible
(pale text on a pale button) on the new version — that's been filed
automatically for a human to look at, through the mechanism you approved this
morning. Second, our work request sat unnoticed for over an hour because the
fleet's build queue serves whichever site has the oldest waiting work, and
other sites have weeks-old backlogs — we dispatched it by hand (a documented,
safe nudge) and have written up both the trap and the remedy for other lanes.

**2026-08-11, evening — the homepage copy: I found out why, and it wasn't the
writing model.**

You asked me to put the homepage back on Gemini because the copy had gone back to
AI slop, then said to leave the model alone. Leaving it alone is right, and here is
why: the model was never the problem. Two other things were, and both are now
written down.

**First, the copy you're reading is about two hours old.** At 17:41 this evening an
automatic review of the site ran, looked for the homepage's content, found nothing,
and concluded the site had no reason for anyone to choose it. It then commissioned
a rewrite whose written instruction was to explain why this calculator beats
MoneySavingExpert and Which, with a pass mark of "claims a benefit other mortgage
calculators don't have". Ten minutes later the framework built and published a new
homepage that does exactly that. "See what the bank's decision engine sees before
you apply" is not a model being clever. It is that instruction being obeyed.

The reason the review found nothing is worth knowing, because it will happen again
on any site we adopt. The review looks in our database, not at the website. Until
this evening our homepage wasn't in the database at all — it was the original page
we adopted in July, sitting in the bucket and serving perfectly. So the review saw
an empty homepage, and everything it concluded followed from that.

**Second, the voice you object to was written down in the site's own settings, and
had been since we adopted it.** When we took the site on, the framework studied the
original and recorded how it wrote. That record says, in as many words: be
challenging rather than reassuring, never soften bad news, write in the lender's
voice to sound like an insider, and put invented labels in quote marks — its own
examples are "Flight Risk" and "The Inheritance Destroyer". It also explicitly
forbade writing in a reassuring tone, and listed warm phrasing as something the
site would never say. Emoji on the cards were mandated too.

So every writer we pointed at this site was being told to do the thing you don't
want. It was faithfully copying the original author's style, which is what we asked
it to do back in July.

**What I've changed.** I've rewritten those instructions: the customer's own words,
never the lender's voice, warm and calm, plain sentences under twenty words, no
clever or ironic headings, no invented labels, no emoji, no urgency, and a flat ban
on comparing us with other websites. I didn't invent the rules — I used the
readability standard you set earlier today on the other project ("readable by a
five year old", short sentences, ordinary words) and our own house style guide for
plain writing. Everything else in those settings, including all the compliance and
disclaimer rules and the boundaries with the loan sites, I left exactly as it was
and checked afterwards that I hadn't dropped any of it.

**Two things I have deliberately not done, and I'd like your call on one.**

I haven't rewritten the homepage yet. This morning another team found that the
mechanism which rewrites a homepage's words also strips out its layout — it kept
84% of the text and none of the styling, turning a designed page into a flat list.
I'd rather drive that once, carefully, and check the page's structure survived,
than fire it tonight and hand you a page that reads better and looks broken.

And here's the thing I need you to decide. Half the words on the homepage aren't the
homepage's. Each tool and guide card shows the target page's own title, exactly as
stored — which is why you're looking at "Stamp Duty Calculator 2026 — UK SDLT Rates
| MortgageCalculator.co.uk" as a heading, pipe, domain name and all, and at "The
Unvarnished Truth" and "The Mortgage Prisoner Trap" on the guide cards. No amount of
rewriting the homepage touches those, because they live on the other pages. There's
already a short plain label on every card ("Stamp Duty", "Buy-to-Let") that we could
show instead. Switching to it is a change to a component shared by 252 pages across
18 sites, so it isn't mine to make quietly. The alternative is to change those pages'
titles, which affects how they appear in Google. Which would you prefer?

**2026-08-11, late — where the evening ended.**

Everything you asked for is live. The homepage reads the way you wanted it, the
three invisible Calculate buttons are fixed and I've had that confirmed
independently, and the automatic process that kept rewriting the homepage behind
our backs is switched off with its leftover instructions closed.

The one thing I'd flag as unfinished is small and mechanical: all thirty-one pages
have their new titles stored, but a page only picks its title up when it's next
rebuilt. The homepage has done that, so the cards you see are all correct. The
individual tool and guide pages will still show their old title in the browser tab
until each one is rebuilt, which is a routine pass I can run whenever you want it.

I've also written the evening up properly for whoever picks this up next, because
a fair amount of what we learned was about our own rules rather than about the
site. Four times tonight the fault turned out to be in a rule I'd written, and each
one was caught by you reading the live page rather than by any check I'd built. The
one that surprised me most was your point about density: the presumptive heading
wasn't wrong on its own, it was wrong six times in a row. That's a better rule than
the one I had, and it's now written down the way you put it.

The copywriting summary you asked for is saved in two places, one of them outside
the repository, and it's also on a private web page you can share if it's useful to
anyone else.

---

**2026-08-13 (agent).** You asked why the framework didn't put the hero, nav and
card images right, told us to build the missing handlers but not switch them on,
and this morning said to carry on — noting the site has no hero or logo at all.
That last observation checks out, and it is worse than it looks: every single
brand file the pages ask for — hero, logo, favicon, the social-share card — comes
back "not found", even though the system's own records say the favicon and share
card jobs finished successfully, twice. The pattern from the whole investigation
holds: the machinery keeps reporting success while the thing it was supposed to
deliver never arrives where the pages look for it.

Where we are: the root cause for the hero (a one-line configuration slip that
files images under a garbled name) is diagnosed and still live; the fix is written
up and is the very next step. The hero image itself already exists and looks fine —
it is just filed under the wrong name, so putting it right costs nothing new. The
logo is a decision rather than a repair: you rejected the two machine-made logos
last week and preferred the old one, but we can't yet find where the old one
lives — that needs a quick hunt, and if nothing turns up it is your call whether
the machine tries again. The two new handler agents you asked for exist and are
switched off, exactly as you instructed, until that configuration slip is fixed.

This chat was getting too long to work well, so the full technical
continuation plan is written down for a fresh session to pick up without losing
anything.

---

**2026-08-14 (evening).** The site has its branding back, and it came through the framework.

What you'll see: the gold roundel logo is live again — it's now properly registered in the
platform as the site's logo (taken from the original site's own file, exactly as you asked:
"carry on with the original logo"). From it the system derived a favicon (the tab icon) and a
social sharing card, both live. The hero image is the one you said "this is ok" about — the
plain navy one with the five icons and no text. A different session's bug-retest had put a
newer generated hero up earlier today, but that one had a big wordmark baked into the image,
which fought the headline text the page lays over it, so I checked with you and swapped to
your earlier choice. The newer one is kept, not deleted, same as the others. The clunky
"you don't need to sign up for any of it" line is being replaced with "It's all free, and
there's nothing to sign up for" — the edit is queued and should be live shortly.

Why the framework hadn't done any of this by itself, in one breath each: the design detection
loop has been switched off fleet-wide since the 10th (a deliberate cost pause while a deploy
bug was wasting every image it touched — that bug is now fixed and live); the favicon jobs it
had filed earlier were marked "done" by a separate platform bug without ever running; and the
hero had used up its two automatic attempts. The two router agents you asked for are now
assigned and proven working on real items. Detection is running once, tonight, for this site
only — turning it back on for the whole fleet is a cost call that's yours to make.

One thing worth knowing: while doing this I caught, live, the platform bug that falsely marks
jobs "done" (it swaps in another job's paperwork). It's now evidenced with traceable IDs in
the bug files (213/274) — the fixing lane has what it needs.

---

**2026-08-16 (morning).** Checked back in after the weekend roll.

The good: everything from Thursday night is still up, and the header now shows the gold
roundel as an actual image instead of the text — the site's own chrome re-render picked the
logo up on its own once the asset existed. That was the piece I'd said was "parked behind the
nav bug"; it wasn't blocked after all. The nav has About in it now too. Someone else's
improvement pass rewrote the hero sentence again on Friday evening; the new one ("No sign-up,
no upsell, and no personal data collected…") is a sentence a person would say, so I've left it.

The platform fix you asked me to confirm (the one that was marking jobs "done" with the wrong
paperwork) IS in the live build and IS working — I measured it: zero of those failures against
859 finished jobs since the roll. But the same roll brought in a *different* wrong-paperwork
shape — jobs now get stamped "done" with the record of the agent being *started* rather than
what it *did*. The work itself is fine; only the receipt is wrong. It's about three-quarters
of all completions since Friday. I've filed it (bug 287) and sent it through the diagnosis
loop; the fix belongs to whoever owns the coordinator, not this lane.

The honest disappointment: the ten tool-page hero images you told me to let run did run — ten
images, all generated, all uploaded — and **not one of them is on a page**. The tool pages
re-rendered and fell back to the site hero; nothing links the new images to their pages. That's
a known class of bug (114) and I've added these ten as a clean, measured example. I've left the
images in place so whoever fixes the wiring has something to test with. Nothing more of that
kind should be generated on this site until it's fixed — it's spend with no visible result.

## 2026-08-16 (Sunday afternoon) — the page titles turned out to be already done, and I found eight links on the site that go nowhere

Picked the site up again from the Saturday-morning notes. Two jobs were left on the list. The
first one was already finished, and the second one turned into something worth telling you about.

**The titles are done.** The list said thirty pages were still showing their old browser-tab
titles. I checked every page on the site against what it is supposed to say, and all twenty-seven
that are live are correct. They fixed themselves: the pages were rebuilt over Friday and Saturday
for other reasons, and a rebuild picks up the new title on its way past. Worth saying plainly —
if I had trusted the list instead of checking, I would have rebuilt thirty pages to change
nothing. The list was right when it was written five days ago and nobody had looked since.

**Eight links on the site lead to a page that isn't there.** I went round every internal link on
every live page and fetched each one. Four destinations are dead:

- the **Mortgage Scorecard Simulator** — four different pages send you to it, and it is also in
  the menu at the top and bottom of every page;
- **"The Secret Scorecard"** guide — two pages send you to it;
- the **lender restrictions** guide — one page sends you to it;
- one link to the **rate forecaster** that is written in a slightly different form from the
  address the page actually lives at, so it 404s even though the page is perfectly fine.

The first three were planned back on 31 July and never built, but the copy on the site talks
about them by name as though they exist — so a reader following the sentence gets a 404.

**Why nothing flagged this.** The part of the system that checks links has not looked at this
site since 9 August. It is switched off across the whole fleet — that was already known and
written down by three other people working on other things, so it isn't news, but it is the
reason. Two of the eight were caught on 9 August, before it went off, and have been sitting in a
"needs a human to decide" queue ever since. The other six arrived after it stopped looking.

**You chose to build the three missing pages** rather than re-point the links at pages that
already exist. That is the better answer and it is also the tidier one mechanically: the links
are already correct, so building the pages fixes seven of the eight without editing a single
sentence of the copy we have spent weeks getting right. As it happens the system had already
started building the scorecard page by itself this afternoon — I found it mid-build — so I have
left that alone and queued the two guides behind it.

**One thing I am deliberately not fixing by hand.** The eighth link, the rate-forecaster one, is
a symptom of something broader: on this site, and on the other sites hosted the same way, an
address ending in a slash always fails. `/guides/` fails; `/guides/index.html` works. Both work
on the sites hosted the other way. So anyone who types the address, or shares it, or follows an
old link without the exact ending, gets a 404 — and that is invisible to every internal check we
have, because the software treats the two forms as the same page. The right fix is at the hosting
layer, once, for every site of this kind — not a hand-edit of one link here. I have written it up
where the next person will hit it. **It is your call whether that is worth doing; it is a change
to how the sites are served, not to the site itself.**

### Later the same afternoon — two of the three pages are built and live; the third is blocked by a known platform bug

The two guides are done and on the site: **"Where you stand before you apply"**
(`/guides/mortgage-scorecard/`) and **"What might limit your options"**
(`/guides/lender-restrictions/`). The framework wrote both — I only un-parked the two build jobs
that had been sitting in the queue since 31 July. They read in the right voice, they have the right
titles, and the three links that used to lead nowhere now land on them.

**The Scorecard Simulator did not build, and it is not a fluke.** The system had actually started
building it by itself this afternoon, before I got to it — and its own quality gate refused the
result, because the page came back with raw template code in it (`{{if …}}`, `{{end}}`) instead of
finished HTML. That is a known bug, filed on 12 August by another lane, root cause proven, not yet
fixed. The gate did the right thing: nothing broken was published. I have added our case to that
bug, with the measurement that it has now happened eleven times across four sites and that six of
those eleven are ours.

**One thing I got wrong today, and it is worth you knowing how.** I wrote up the trailing-slash
problem as a new discovery. It was not — it was already written down in our own traps file, and that
entry even named this site as an example. The reason I missed it is that I searched for it in my
words ("trailing slash", "directory-form") and it was filed in someone else's ("a /section/ URL").
I found out by accident, forty minutes later, when I went looking for where the fix would live. I
have corrected it everywhere I claimed it, logged it in the wrong-calls file, and the rule I took
from it is: search for prior work by the *name of the thing* (the file, the function), never by the
words you would use to describe the symptom. Two people describing the same bug rarely choose the
same words; they cannot avoid choosing the same filenames.

**A caveat on today's link fix, stated plainly.** Building the two guides fixed three dead links —
not seven, as I expected. The two new pages each link to the Scorecard Simulator themselves, because
the site's own brief names it as a page that should exist, so the writer keeps referring to it. Dead
links went from eight to seven, not from eight to one. The site is slowly accumulating references to
one page it cannot build, and that will keep happening until that bug is fixed. Nothing here is
broken by it; it just means the tidy-up is not finished and cannot be finished from this end.

## 2026-08-17 (Monday morning) — the link checker is back on, and the stray design file is gone

Three things you asked for. Two were changes and they are done; the third was an explanation.

**The link checker is running again.** This is the part of the system that reads every page on a
site and checks that the links actually go somewhere. It had been switched off since 10 August,
which is why eight dead links sat on this site unnoticed. Before turning it on I checked what
turning it on actually costs, because that was the question you paused it over: it takes **one
site per hour**, and only sites it has not looked at for a week, so it works through the backlog
of 22 sites over about a day and then goes quiet. It picked up its first site a minute after I
enabled it.

One honest caveat. The checker only *finds* things; a separate mechanism, which has been switched
on since you paused this one, now automatically promotes what it finds into actual work. So over
the next day you should expect a fair amount of repair work to start moving across the estate —
on one site last week a single pass found about 77 things. That is the system doing what it is
for, but it is more activity than "just switch the checker back on" sounds like, so I would rather
you heard it from me now than noticed it tomorrow. Turning it off again is one command and it is
written down in the handoff.

**The stray GIMP file is gone.** That is the 175 KB design master that has been sitting publicly
downloadable on the site since we adopted it, flagged three times and never dealt with. It nearly
got dealt with badly: it turned out the copy on the web server is *copied up from a folder on this
machine*, and it had been re-uploaded from there as recently as Saturday evening — so deleting it
from the web server alone would have quietly put it back on the next sync. I removed both. It is
now a 404, nothing on the site linked to it, and I kept four identical copies in different places
(including the deploy repository's own history) so it is not lost in any sense.

**The trailing-slash fix I explained rather than made.** Nothing changed at Cloudflare or in the
code — it is written up for you and for whoever reviews it, and it stays your call.

**Still waiting on a decision that is not mine:** another team has left a question at the top of
our handoff, about adding one line of configuration to our stamp-duty checker so it gets told when
the tax thresholds it relies on change. They deliberately did not do it themselves because it
would add items to our queue. It has been waiting since Saturday and it needs a yes, a no, or a
"do the harmless trial version".

## 2026-08-21 (Friday morning) — the thing we were blocked on got fixed by someone else, and the site is now down to one broken link

**First, a correction to the entry above.** The last thing I wrote to you said the trailing-slash
fix was "explained rather than made" and that it stayed your call. That was true when it was
written, on the Monday morning. You approved it later and **it shipped on the Tuesday**, which
means the paragraph above is now wrong and has been for three days. It is live on every site we
host in that way, not just this one, and it was checked properly before and after. I am leaving
the old paragraph where it is rather than tidying it away, because the record of what we believed
when is worth more than a clean page.

**Now today.** I picked this site up cold after a three-day gap. In that time roughly fourteen
hundred changes landed on the shared codebase from other people working on other things, so before
touching anything I went back and re-measured every number the handover note carried. Four of the
figures had moved. One of the four is the thing we had been stuck behind since the start of the
month.

**The blocker is gone, and we did not fix it.** There is one page on this site — the Scorecard
Simulator — that the system has been unable to build since we adopted the site. The reason was a
fault deep in the page-building machinery: when the writing stage got one field slightly wrong, the
renderer quietly fell back to a cruder method that left raw programming instructions sitting in the
page, and the safety check then correctly refused to publish it. That fault was filed as a bug back
on the 12th, another team took it on, and **they closed it yesterday**. So the thing that was
standing in our way has been removed by somebody else while we were away from this site.

I did not take that on trust. Before acting on it I checked that the fix is genuinely running on
the machines serving this site — not merely committed, which is a different thing — by asking the
running program directly whether it contains the new code and no longer contains the old. It does,
and I ran two deliberate control checks alongside it so I could be sure the test itself was capable
of coming out wrong. It was.

**So I have restarted that page's build.** It is running as I write. There are two possible
outcomes and both are useful. Either the page builds, in which case six broken links on the site
die at once. Or it fails again — but the whole point of the other team's fix is that a failure now
names the single field that is wrong instead of burying it in twenty vague complaints, so even a
failure hands us something we can actually repair.

**The rest of the site is in better shape than I expected.** I fetched all twenty-nine published
pages and checked every internal link on them — a thousand and thirty links, going to thirty-three
distinct places. **Thirty-two of the thirty-three work.** The only one that does not is the
Scorecard Simulator page above, and I checked it three times to be sure it was not a caching
fluke. The site has also quietly grown from twenty-seven pages to thirty-two.

**One thing I want to flag, because it is a small design failure rather than a bug.** The system
*did* notice all six of those broken links. It filed a separate note for each one, naming the exact
page and the exact block of text containing the link. Then it put all six in a queue that nothing
reads and nobody looks at. So the information was produced correctly and then parked. I have
written that up for the team who filed the related bug, because it changes what the fix should be:
they had assumed the system was blind to this, and it is not blind, it is just unwired. Fixing an
unwired thing is much cheaper than building a new one.

While checking, I also found that one of those six notes is **stale** — it describes a broken link
on the contact page that no longer exists, in either the published page or our stored copy of it.
Something repaired that link and nothing went back to close the note. Worth knowing, because it
means a note sitting in that queue is not proof the problem is still there.

**Two things still waiting on you, unchanged from before.** The thirteen stamp-duty items I
deliberately left open, which need a yes or no on whether closing them is wanted. And the question
another team left at the top of our handover about one line of configuration on our stamp-duty
checker — that has now been waiting since the 16th.

**Two small ones I would add to that list.** The contact page cannot finish rebuilding because it
is asking for a real business email address to display, and nobody has given it one. And a tool
page called "simple" is missing its headline. Both are things only you can answer; neither is
urgent, and neither is breaking anything visible today.

## 2026-08-21 (Friday, later) — the page still will not build, but we now know exactly why, and the reason is somebody else's to fix

I said this morning there were two possible outcomes and both were useful. We got the second one.

**The page failed to build, twice, and the failure now names the exact problem.** Where it used to
say "twenty blockers" — which tells you nothing — it now says: on the component that draws the
step-by-step flow diagram, a field called `branches` is supposed to be a structured list of
outcomes, and the writing stage keeps putting a sentence of prose there instead.

**I ran it twice deliberately, and that is the finding.** If it had failed once and worked the
second time, this would be random bad luck and the answer would be "keep retrying". It failed both
times, on the same field, in the same way — only the position within the page moved. So the writer
gets this **reliably** wrong, and retrying is not a route to a working page. I stopped at two
rather than spending more on attempts that have no reason to succeed.

**It is not our component's fault.** I checked the specification the writer is working from before
assuming anything, and it is clear and correct. If anything the problem is that the specification
*describes* the field in a way that reads like an instruction to write a sentence — and the writer
does what it is told rather than what the type says. That is a known argument another team has
already made in the abstract; this is the first live case of it I can hand them.

**So I have handed it over, with a case they can reproduce on demand.** The team that owns the
writing half of this problem now has the exact error, the exact item to re-run, and — as it
happens — the only live example of this failure anywhere on the estate at the moment.

**I also found a second, separate fault while doing it, and filed it.** When the build failed, the
system marked the job **complete**. Not failed, not "needs a human" — complete, with nothing built
and the page still missing. That matters more than it sounds: a job that reports success is a job
nobody goes back to. This exact shape was found and fixed last month for a different cause, and the
fix was to add a guard for *that* cause specifically. Yesterday's fix from the other team created a
*new* cause, and the guard does not cover it. I filed it as its own bug with the evidence.

I tried to put that second fault through our automated diagnosis loop first, as our rules ask, but
the loop broke twice on its own infrastructure — once on an AI service limit and once on a genuine
bug in the code that is supposed to *record* failures. So I verified it myself instead and said so
plainly at the top of the bug, rather than quietly skipping the step.

**Where that leaves the site.** Thirty-two of the thirty-three internal links work. The one that
does not is the Scorecard Simulator, linked from six pages, and it will stay broken until the
writing team fixes the field. I have put our job back into the "needs a human" queue where it
honestly belongs, instead of leaving it marked complete.

**And one honest note about my own work today.** I set up an automatic watcher to tell me when the
second build finished, and it told me the build had finished thirty seconds after starting — before
it had even begun. I had deliberately kept the previous run's error message so I could compare the
two, and my watcher was looking for the word "failed", which was sitting right there in the old
message. It matched history rather than the event. No harm done, I caught it in the same breath,
and I have written up the check that prevents it. I mention it because a watcher that lies
confidently is exactly the sort of thing that quietly corrupts a day's conclusions.

## 2026-08-21 (Friday, late afternoon) — the contact page is reworded and live

You asked me to reword it rather than supply an email address. It is done and serving.

**What it said before.** The page invited people to get in touch four separate times — the heading
promised "a place here" for questions the tools don't answer, then "tell us here", then a promise
of a reply if you "write to us", then an invitation to report a wrong figure and say what you'd
entered. There was no email address, no phone number and no form anywhere on it. Every one of
those four invitations was a dead end.

**What it says now.** It opens by saying most answers come from the guides and calculators rather
than from writing in, and then says plainly that there is no form or email address on the page. It
keeps the part that was always true and worth keeping — that this site works out figures rather
than giving advice, and that a decision about your own borrowing needs a lender or a broker. Then
it sends people where they can actually be helped: the common questions page and the guides.

**One nice thing I didn't ask for.** I did not write this copy — I wrote the brief and the
framework wrote the words, which is the rule. It found a destination I hadn't thought of: if a
figure from a calculator looks wrong, it now points you at the page explaining the assumptions
behind our figures, which sets out what each calculation is based on and where it can fall short.
That is a better answer than the one I would have written, and it is the argument for briefing the
system rather than typing the sentences myself.

**I checked it properly rather than trusting the system's word.** I swept the live page for all
seven of the old promise phrases — every one now returns nothing — and confirmed there is no email
address, no form, and no email-shaped text anywhere in the page's code. Nothing was invented.

**One judgement I've left to you rather than making it myself.** The page is still *titled*
"Contact us", and the link to it in the footer still reads "Contact". A page called "Contact us"
that tells you there's no way to contact us is arguably fine — it is the page people go looking
for, and it now gives them an honest answer instead of a dead end. But it is a fair thing to
disagree about, and changing it touches the footer of every page on the site, so I have not done it
on my own initiative. Say the word either way.

**And one small confession.** My automatic check briefly reported that an email address had
appeared on the page, which would have meant the system invented one. It hadn't — my checker was
counting "@" symbols and had picked up three of them inside the page's stylesheet, where they are
part of the styling language and not text anyone sees. I caught it within a minute by searching
the page properly. No harm done, but it is the second time today a checker of mine has told me
something confident and wrong, so I have written both up.

## 2026-08-24 (Monday) — the "not financial advice" line is now on every page

You asked for the disclaimer on every page. It is going on now, and the way it is done matters
more than the line itself.

**I did not add it to 32 pages.** The framework already has a mechanism built for exactly this: a
per-site setting called "compliance lines" whose entire purpose is a short statement that appears
in the footer of every page. It was built in early August for other sites and had never been
switched on here. So this is one setting, in one place, and every page picks it up — including
every page built in future, without anyone remembering to do anything.

**Why I asked you for the wording rather than writing it.** Two reasons. That setting is
configuration, not page copy, so the framework's writer cannot produce it — there is no generator
to ask. And a disclaimer about financial advice is a legal statement made in your name, which is
not mine to invent. You chose the site's own voice, which reuses the sentence the framework itself
wrote for the contact page last week, so the line reads like the rest of the site rather than
bolted-on boilerplate. You also chose to say nothing about being regulated, which I think is right
— I have no way to verify your status either way, and a wrong claim about that is worse than
silence.

The footer now carries, on every page:

> This site works out figures rather than giving financial advice.
> Any decision about your own borrowing needs a lender or a broker who can look at your full
> circumstances.

**One thing you should know, because it is a real change and you did not ask for it.** The pages
have to be rebuilt for the new footer to appear on them. I rebuilt one first as a test, and it came
back **two and a half times bigger** — from about 25KB to about 60KB. My disclaimer accounts for
roughly 400 bytes of that. The rest is a 34KB block of styling that the site's design system now
puts into every page, and which this site had not picked up because its pages had not been rebuilt
since the 20th.

I checked whether I had broken something or whether this is just the site catching up, and it is
catching up: every site on the estate rebuilt in the last three days carries a block of that size
or bigger — one is 62KB. The sites still on the small version are precisely the ones last rebuilt
on the 20th, which is what this one was. The words on the page did not change; I compared them and
the only differences are the footer links and the new disclaimer.

I went ahead rather than stopping to ask, because stopping halfway would have left one page on the
new design and twenty-eight on the old, which is worse than either. **But if a heavier page matters
to you, say so** — it is the design system's doing rather than mine, and it affects every site, so
it is worth raising once rather than per site.

**One thing I found while doing it, which I have written up for whoever picks it up.** Three of the
tool pages have a section with no stored content behind it. That is harmless the way I rebuilt
them, but anyone who rebuilds those three a slightly different way would silently have their words
regenerated. Two of the three are the pages another team's automated check called "serving broken"
last week and then closed without changing anything. That cluster deserves a look; it is not part
of this job.

**Confirmed done, same afternoon.** All **30** pages that the site actually serves now carry the
disclaimer — checked twice, independently, with repeat probes rather than one look. Nothing failed.
The two pages that are not live yet (the Scorecard Simulator and one guide) will pick it up
automatically the first time they are built, because the setting lives with the site rather than
with the pages.

One small thing worth knowing, since it is the second time this week: one of my automatic checks
briefly reported a page as missing the disclaimer when it was fine. I said at the time that I knew
why — that the page had got too big for my check's time limit — and then could not reproduce it, so
I have withdrawn that explanation rather than leave a tidy-sounding reason in the record. It was a
one-off failed fetch. What matters is that my checks fail by crying wolf rather than by saying all
is well when it is not, which is the right way round.

---

## 2026-09-02, later — you asked me to verify the tools and fold in the images. Here is what I actually found.

**Short version.** The tools are all up and none of them is obviously broken, but I cannot honestly
tell you they work, because the machine we built to check them is currently looking at half of them
and is being lied to about the other half. The images are all sitting there, correctly made, and
there are two separate reasons none of them reaches a tool page — neither of which is the reason
this morning's handover note said it was. I have fixed nothing on the live site yet, on purpose,
and I explain why below.

**A note on this morning's handover, because it matters.** The session before me wrote a careful
document saying both of your requests needed re-aiming. It was right that they needed re-aiming and
wrong about where to aim instead — on both counts. I say this not to score a point but because it
happened for a reason worth knowing: that session reasoned about the *settings* attached to the
tool pages without ever looking at what those pages are actually made of. When I looked, the picture
changed completely. There was also a message from another team sitting unread in our folder that
would have told them, and I did not read it until after I had worked it out the slow way. **Read the
incoming post before trusting the outgoing summary** — that is the lesson, and it is mine as much as
theirs.

### The tools

All eighteen calculators are live and serving. I fetched every one and checked that each button and
box the page's own code reaches for actually exists on that page — no broken wiring anywhere, and no
half-finished template text leaking through. That is a genuine result and it is not nothing.

**But it is not a verdict, and I want to be plain about why.** That check can only see code that
names a thing directly. Two of the eighteen do it a more roundabout way, so for those two my check
saw nothing at all and reported no problems — which is not the same as there being none. The only
honest test is to open each calculator in a real browser and use it. We have exactly that machine.
It is not working, for two separate reasons.

**Reason one: it isn't looking at half of them.** The system has a rule for which tools it is
allowed to check. Nine of our eighteen pages fail that rule — not because anything is wrong with
them, but because each one has a second block of text sitting next to the calculator, and the rule
only admits pages where the calculator is the *only* thing on the page. Seven of those nine still
have a full set of written tests attached to them, from the middle of August, that nothing has read
since. The tests are fine. Nobody is running them.

**Reason two: where it is looking, it is being told the wrong thing.** Back in August the platform
renamed every part inside every interactive tool, to stop two tools on one page clashing. Sensible
change, done properly. Nothing updated the tests, and nothing taught the checker about the rename.
So the test says "click the box called X", the box is now called something slightly longer, the
checker looks for exactly X, doesn't find it, and reports the tool broken.

**This is not just us.** Across the whole estate in the last six weeks there are 187 of these
"the page is missing a part" failures. I checked each one against the tool's own code: **134 of
them, across 99 different tools, name a part that is definitely there** — just under its new name.
Roughly three in four of those alarms are false. And they are not harmless: each one raises a job to
go and fix the tool, so an automatic repairer goes and edits a calculator that was never broken. I
watched one of those happen this afternoon on another team's site while I was writing this up.

I have put this into the platform's own diagnosis process rather than just asserting it — it is a
big claim about shared machinery and it deserves an independent check before anyone acts on it.

### The images

Every one of your eighteen tool pages already has its own picture, made for it, live on the server,
correct. Fourteen of the eighteen pages show no picture at all. There are two different reasons, and
this morning's note guessed a third that does not exist.

**Ten of the fourteen have nowhere to put a picture.** This is the thing worth understanding, and it
is odd. On those ten pages, the slot that is supposed to hold the banner at the top of the page
actually contains *the entire calculator*. It was filed in the wrong drawer when the site was
adopted, and it has stayed there working perfectly ever since. So the page has no banner, the
picture setting attached to that slot does nothing, and — this is the important part — **anyone who
"fixed" this by making the banner render would delete the calculator and replace it with a title
strip.** This morning's note pointed the next session at exactly that. Another team has already
spotted the same thing from their end and has a careful repair ready to run; I have written to them
about one consequence they had not costed.

**The other four have a banner that structurally cannot show their picture.** This one is a
genuine, small, fleet-wide defect and I am fairly pleased with it. The banner used on tool pages
draws a background image, but its settings sheet does not declare that it has an image — and the
part of the system that goes and finds the right picture for a page only runs for things that
declare one. So it silently falls back to the generic site picture, for ever. The code even has a
comment predicting this exact outcome. Across the estate that is **54 pages on 21 sites** whose own
picture exists and cannot be shown. The fix is four lines of settings, copied from the banner used
everywhere else, and it cannot make anything worse. I have handed it to the team now working the
imagery bug rather than applying it myself, because it touches 21 sites that are not ours.

### What I have and have not changed

I closed one stale job — a warning about a missing script that genuinely no longer matters, because
nothing on the site asks for that script any more. I checked that properly, and my first attempt at
checking it was worthless in a way I nearly missed: the thing I was comparing against gave the same
empty answer as the thing I was testing, which proves nothing at all. I redid it with a real
comparison.

Otherwise **I have changed nothing on the live site.** The two real fixes both belong to other
teams' territory — one is a shared component across 21 sites, the other is a repair another team has
already designed and pinned. Barging in would break their work, and in one case would fight a safety
check they built specifically to stop sessions like me doing exactly that.

### What I need from you

**One decision.** The ten pages whose calculator sits in the banner slot cannot get a picture until
they are put back in the right drawer. The other team's repair does that, and it is being piloted on
one of our pages today. After it lands, giving those pages a proper banner and their own picture is
straightforward. **Do you want me to wait for that, or to hand the whole tool-imagery job to that
team as part of their repair?** Waiting is my recommendation — it is a few days, and doing it in the
wrong order means doing it twice.

Everything else on my side is either done or handed to the team that owns it.

---

## 2026-09-02, evening — with the login working again, I got real answers. Two of them are about my own mistakes.

**The headline: four of the calculators are now confirmed working by a real browser, and I found
out that "working" means less than I thought it did.**

### I was wrong twice this afternoon, and both errors flattered my own theory

I told you this morning that the machine which checks the calculators was being lied to, and that
this affected every calculator it could see on this site. **Only two, not eight.** I had assumed
that because a page had been renamed internally, the test written for it must be out of date. That
only follows if the test actually names one of the renamed parts, and most of them don't — they
point at things the rename never touched. When I tested it properly instead of reasoning about it,
five of the tests turned out to be perfectly good.

The second error was in the bug report I filed. I measured the problem by looking at the master
copy of each calculator, when what actually gets tested is the published page. Measured the right
way the problem is bigger (sixteen calculators affected, not ten) and the cause is different: some
published pages are simply old, built before the rename, so they still carry the old names even
though their master copy was updated. I corrected the bug report the same day, before anyone acted
on it.

**Both mistakes made my own story look tidier than the truth.** That is the direction to be
suspicious in, and I have written it down where I keep that sort of thing.

### One calculator was recorded as broken and had been fixed nine hours later

The bridging-loan compound calculator's most recent verdict said it failed. Looking at the
timestamps: it failed at 11:57 one morning, an automatic repair finished at 12:21, and the page was
rebuilt at 23:07 — and then nothing ever re-checked it. **So the failure was newer than the fix on
paper, and older than it in reality.** I ran a fresh check this evening and it passes: all nine
tests, on desktop and on mobile, including the exact one that had failed. It took forty-eight
seconds.

That is worth remembering as a general thing: **a failure notice outlives its own repair whenever
the thing that fixes it doesn't re-run the test.**

### The uncomfortable finding: the tests don't check the arithmetic

This is the one I would most like you to know about. There are two kinds of test on this site and
they check opposite halves of the problem.

The older tests, written by hand for the original calculators, check **the sums are right** — feed
in a mortgage, confirm the monthly payment is the exact expected figure. They check nothing else:
not whether the page loads, not whether it works on a phone.

The newer tests, written automatically, check **the calculator is alive** — the page loads, nothing
errors, it fits a phone screen, and something appears when you press the button. They do not check
a single number.

So when I tell you a calculator "passes", ask which kind. Two of the four passing tools pass a test
that would be equally happy if the calculator printed a wrong figure with confidence.

I checked whether this is just us: across the whole estate the automatic writer has produced **170**
of these tests and **107 of them assert no expected value at all**. The reason is small and fixable
— the instructions that agent works from list the test types it can use, and the one that checks a
number simply isn't on the list. It has never been told it exists. One team's tests do use both
kinds together, so it plainly works; nobody told the generator about it.

I have filed that, with a warning attached: the fix must not be "let the generator make up the
expected answers", because it would then pin whatever the calculator does today, bugs included. The
expected figures have to come from somewhere other than the calculator — which is exactly what the
hand-written tests on this site already do.

### Where the calculators actually stand

- **Four confirmed working** — but see the caveat above about what the checks cover.
- **Two genuinely blocked**, by the two faults I filed today: the test looks for parts under their
  old names, and the automatic repairer that should fix it can't start because of a separate flaw
  in how the job is created. That second one accounts for **62% of all failed repair jobs across
  the estate**, so it is not just us.
- **Three have no test at all** — writing them is real work, needing a trustworthy source for each
  expected figure, and I have not started it.
- **Nine cannot be tested yet** and need the other team's repair to land first.

### What I have not done

I still haven't touched the images, for the reasons in my last note — both fixes belong to other
teams and one of them is mid-flight. **The question I asked you earlier still stands**: wait for
their repair, or hand them the whole imagery job. Nothing else is waiting on you.

---

## 2026-09-03 — the big repair landed overnight, and it quietly switched off the site's own testing

**What happened while we were away.** You applied the migration another team had prepared — the one
that puts each calculator into its own properly-labelled box instead of leaving it filed under
"page banner". It worked. All eleven moved, the pages look and behave exactly as before, and that
team has closed their bug.

**The good news is bigger than they claimed.** Before it, the automated checking system could only
see nine of our eighteen calculators; the rest failed its "is this really a tool?" test because of
how they were filed. **It can now see all eighteen.** That is a real gain and it was not in their
notes.

**The bad news is the thing I warned them about yesterday, and it happened exactly as described.**
The checking system finds each calculator's test by name, and the migration changed the names. All
eight of our tests were left pointing at names that no longer exist. Nothing errors when this
happens — the system just quietly finds no test and moves on. **Including the one calculator that
was the only properly-tested thing on the site the day before.**

I have repaired all eight. Before changing anything I checked five things that could each have
stopped me — that no other name was already taken, that no other website shares these tests, that
every part each test looks for is genuinely on the live page, and, as a control, that the *old*
tests really do fail now (they do). The last check ran inside the change itself and would have
undone the whole thing if the count came out wrong. No expected figure was altered; only the
addresses.

**And a thing nobody predicted, which I think is the most useful finding of the day.** The migration
gave each calculator a new internal template that renames every element on the page — but left the
already-published pages alone. Those two states agree only until something rebuilds the page. At
around a quarter to nine this morning a routine rebuild wave rebuilt five of the ten, and their
element names changed. **So right now the site is half-converted: five calculators renamed, five
not.** The calculators themselves are fine — I checked every one for broken internal links and found
none — but each of those five rebuilds silently broke that calculator's test, and the other five
will do the same whenever they next rebuild.

That reframes the fault I filed yesterday. I had it as a pile of stale tests to fix once. **It is
not a pile — it is a tap that is still running.** Five broke in four minutes this morning with
nobody intending it. I have said so in the bug, because it changes which fix is worth doing: the
only stable one is teaching the checker to accept both naming styles.

It also retires a reassurance that sounded solid. That team verified "the bytes are unchanged", and
they were — at the moment they ran it. It stopped being true eleven hours later, because the
migration had changed *the template the next rebuild would use*. **A check that nothing moved cannot
promise that nothing will move.** I have written that into their closed bug as information, not as a
complaint — the migration did what it said it would.

**One more repair, unglamorous but it had stopped the whole toolkit.** The script this lane uses to
check its own test figures had died. It reads the site's register of verified facts and expects
every one to be a number; five new citation entries have no number, and it crashed on the first.
That took the installer down with it. Fixed, and — importantly — I proved it still *fails* when it
should: I deliberately fed it a wrong stamp-duty figure and it caught it and refused. A checker that
only ever passes is worth nothing.

**Where the calculators stand.** Four are confirmed working by a real browser. Eight more now have
live tests again and I have queued fresh runs for them; those are still waiting in the fleet queue,
which is normal and slow, so the results will land after this note. Two are genuinely blocked by the
two faults I filed yesterday. Five have no test at all — and one of those, the portfolio calculator,
**must not get one yet**: our own tooling told me its expected figures have never been independently
worked out, only copied from what the page prints. Pinning those would enshrine whatever it does
today, right or wrong. That is the tooling refusing me, and it was right to.

**Your instruction on the images is done.** The whole job is handed to the team working that bug,
with everything we found and a correction they needed: the reason I gave you yesterday for not
touching ten of those pages — that there was nowhere to put a picture — **stopped being true when
the migration landed**. There is now. It is their call.

**Nothing is waiting on you.** The copy-quality question from last week still needs an answer when
you have a moment, but nothing is blocked on it.

---

## 2026-09-03, later — I got the cause wrong this morning, and I want to correct it plainly

**The correction.** In my note earlier today I told you the overnight migration had given each
calculator a new internal template that renames every element, and that this was why five
calculators changed and five did not. **The migration did not do that.** It copied each calculator
across exactly as it was, which is what its own checks said and what its team claimed. I checked
which migration had run most recently and assumed it was responsible. That is a guess dressed as a
finding.

**What actually did it** was a separate automatic sweep, which runs across the whole estate looking
for interactive tools whose internal element names could clash if two appeared on one page. At
twenty to eight this morning it found the eleven freshly-copied calculators, correctly noticed they
had not been through that renaming, and converted them between about half past eight and a quarter
to nine. Each conversion then asked for the page to be rebuilt, and five of those rebuilds ran.

**Why I am making a point of this rather than quietly fixing it.** I had written the wrong version
into six different documents, including a note to the migration's own team telling them their safety
check had expired. It had not. I have retracted that in full, in their file, with the evidence.
Getting a cause wrong is ordinary; letting it spread to five other places and to another team's
record is the part worth flagging.

**And the true version is more useful than mine was.** Three separate automatic actions, each
completely correct on its own — copy the calculator across as-is, rename elements that needed
renaming, publish the result — combined to switch off the testing on those calculators. **No one did
anything wrong, and nothing in the system could have seen it coming**, because each step only knows
about its own job. That will happen again on the next site that gets the same treatment, which makes
it predictable rather than freakish, and that is worth far more than a culprit.

**Where things stand right now.** I caught something before it did damage: six page rebuilds are
queued, and they target the five calculators that have not yet been converted. When they run, three
of the tests I repaired this morning stop matching their pages — and I had test runs queued directly
behind those rebuilds, which would have recorded three failures against calculators that work
perfectly. **I have held those three back.** The order is: let the rebuilds happen, re-point the
three tests at what the pages actually serve, then run them.

**The five safe test runs are still waiting in the queue.** Not stuck — waiting. This site was last
served by the job scheduler at ten to nine this morning, and there are eighteen sites sharing it. I
also spent a while convinced the whole testing system had broken estate-wide, because the numbers
looked exactly like that; it had not, and I have written down what talked me out of it. **Nothing is
broken and nothing needs your attention** — I will report the results when the queue reaches us.

---

## 2026-09-03, evening — I broke the tests this morning and did not notice until they failed

**What happened.** The queue finally moved this afternoon. All the page rebuilds went through, every
one of the eighteen calculators is now on the new naming scheme, and the five tests I had queued came
back. **All five failed — and all five failures were mine, not the calculators'.**

**The mistake.** A test has two halves: the boxes it fills in, and the answers it checks. This
morning, when I repointed the tests at the renamed elements, I fixed the first half and missed the
second. So each test filled in the right boxes, pressed the right button, and then went looking for
an answer field under a name that no longer existed. From the outside that looks exactly like a
broken calculator. **The calculators were fine throughout.**

**The part that actually worries me** is not the slip — it is that my own check said everything was
fine. I had verified the repair before shipping it, with a control, and both passed. They passed
because the check was written from the same wrong assumption as the repair: it looked for the answer
fields under the same name I had got wrong, so it never examined them at all. **A check built on the
same misunderstanding as the thing it is checking will always agree with it.** I have written that up
properly, because it is the most useful thing to come out of today.

**How I fixed it, and why I think it holds this time.** I repointed all eight tests — the five I had
half-repaired, plus three I had deliberately held back this morning and which had gone stale in the
meantime exactly as I predicted. Then I verified a completely different way: instead of looking for
particular named fields, I read every element reference out of the raw test text, so no naming
mistake can hide from it. I also checked the old versions fail that same test — they do, on between
one and nine references each — which proves the repair actually changed something rather than
agreeing with itself. And I read the finished tests back out of the database rather than trusting the
files on my disk, because "the file I wrote is right" and "the record is right" are two different
claims and I had just been caught conflating two things that looked identical.

**No expected figure was altered at any point today**, in either pass. Only the addresses.

**One piece of luck worth naming.** These tests carry a flag meaning "if this fails, tell a human,
never send the automatic repair robot". I set that months ago for an unrelated reason — because an
automatic repairer can only turn a failing sums-test green by changing the sums. **That flag is the
only reason this stopped at five wrong entries in a log** instead of five working calculators being
"repaired" into agreement with a broken test.

**Where it stands.** Eight fresh test runs are in flight. I will report them when they land. The
calculators have been working correctly all day and were never at risk; what was broken, and is now
fixed, was our ability to tell.

**A small irony I have recorded rather than buried.** Earlier this week I filed a fault report about
exactly this: tests pointing at old element names and recording false failures against working tools.
Then I produced five more of them by hand, inside the repair for it. Knowing the trap did not help,
because my error was a level below it — I never checked where the test format actually keeps those
references.
