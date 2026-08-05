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
