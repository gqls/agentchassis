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
