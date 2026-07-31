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
