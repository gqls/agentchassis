# Where we are — the broken social preview (bugs_open/131, og-card)

Plain-prose log, append-only, newest at the bottom.

---

## 2026-07-28, evening — we can fix it today, but looking at the result found something worse

**The problem, in one line.** When anyone shares a link to one of our sites — WhatsApp, Slack,
X, LinkedIn, iMessage — the page tells the app "here is a preview picture", and on 11 of our 14
sites that picture does not exist. So the share renders blank. It has never worked.

**The good news is that we already built the fix and forgot.** The bug file said someone should
go and find out how the two working sites got their picture, "before building a new generator".
I did, and the answer is that there is already a generator: it takes the site's logo, sits it on
the site's brand colour, and produces the 1200×630 card that social apps want. It has been
sitting in the system since 11 July. Two sites got it; nobody ever ran it for the others.

So the expensive-sounding option — write code, get it reviewed, rebuild, redeploy the fleet —
turned out to be unnecessary for the actual problem. **We just needed to run the thing.** I ran
it for relojistas as a test and it took eighteen seconds. The picture is now live.

**And then I looked at the picture, and that is where this gets interesting.**

Every automated signal said success. The job said complete. The URL returned 200. The file was
a valid PNG at exactly the right size. The database rows were written. All green.

The image is a picture of a **brand specification sheet** — the "Relojistas" name printed
*twice*, side by side, once on cream and once on black, the way a designer shows you a logo
will work on light and dark backgrounds. It is not a logo. It is a page *about* a logo.

That is not the generator's fault. The generator did precisely what it promises. **The problem
is that the thing we have stored as relojistas' "logo" is that specification sheet.**

**Which means the real finding is not about social previews at all.** That same file is what the
website uses as its header logo, on every page. It is 1408 pixels wide and gets squeezed down
to about 81 pixels in the header — two copies of the name crammed into a space barely wider than
a thumbnail, both far too small to read. I checked the stylesheet in case it was being cropped
cleverly. It is not; there is no crop. This has been live for weeks, on the site I described in
my own handoff as "finished".

I want to be straight about why I missed it. Everything I checked on this site was about what a
*crawler* receives — the tags, the status codes, the robots file, the sitemap, the feed. That
is genuinely "checking the live site", and it has nothing whatever to say about what a *person*
sees when the page loads. The two do not overlap, and I only had one of them.

**Where that leaves the rollout.** I was going to run the same job for the other eleven sites
tonight. I have stopped, because I would be putting eleven generated images onto live sites
without looking at any of them, and the one I did look at was wrong. I checked two other sites'
logos and they are proper, sensible marks — so this is a per-site problem, not a fleet-wide one,
and the rollout is probably fine for most of them. But "probably fine" is not something I want
to discover after the fact on your brand.

**Did I make relojistas worse?** No. Both files previously returned "not found", and the page
already falls back to that same specification sheet for the tab icon, so the icon is no worse
than it was. A share now previews *something* rather than nothing. But that is a low bar and I
am not claiming it as a win.

**What I need from you.** Two decisions, and they are both about the logo rather than the code:

1. relojistas' logo needs to become an actual logo — the left-hand half of that sheet, cropped.
   Then the header, the tab icon and the social card all come right at once. I can do the crop,
   but it is your brand and the current file was approved at some point, so I would rather ask.
2. Whether to run the job for the other eleven sites. My suggestion is yes, but that I look at
   every card afterwards and show you any that are wrong, rather than assume.

Separately, and much less urgent: the system's own watchdog for this ("an asset was generated
but never deployed") has fired five times, and every single one was for robot-hands — the one
site whose picture was already working. It has never once flagged any of the eleven broken ones,
because it can only see assets that were *generated*, and theirs never were. A watchdog that
starts from the list of things that exist cannot tell you about a thing that is missing.

---

## 2026-07-29, morning — you answered, and the plan is now concrete

You gave me three answers: do the storage work **through the chassis** (no credentials handed
to me — the platform does its own writes), apply the corrected relojistas logo **everywhere**
(the social card, the tab icon and the page header all come right together), and **generate
fresh logos** for the two sites whose stored "logo" turned out to be junk — with the promise
that I look at every generated image and show you before anything goes live.

One new finding while checking: idea.uk's problem is worse than a broken pointer. I pulled down
the logo its own pages serve, and it is another AI mangle — lettering that reads more like
"IBTA" than "IDEA". So even if I repaired its plumbing, the picture at the end would still be
wrong. It needs a new logo, same as gaswholesalers, which is why I've put it in that bucket.

And one course-correction on my own yesterday's advice: I suggested protecting leopardess's
hand-made card by putting a lock on a database row. I read the actual code this morning and
that lock would not protect the picture — the file is committed to the site before the lock is
ever consulted. So the protection has to go into the deriving code itself: teach it to refuse
to overwrite an artefact that has been marked as approved. I'm folding that into the same small
code fix that stops tab icons coming out squashed, since they are two faces of the same rule:
the machine must not destroy or distort things a person has signed off.

Order of work from here: land that small code fix first (it goes through the reviewer council,
a build and a deploy), because the relojistas re-run should produce its tab icon *after* the
squashing is fixed, not before. While that is in review, get the corrected relojistas logo into
storage via the cluster, and fire the two fresh-logo generations so there is something to show
you. Nothing goes live on any of the three sites without the pictures having been looked at.

---

## 2026-07-29, late morning — relojistas' logo is fixed and live, and I got the deploy wrong once on the way

**The headline: relojistas.com now shows its actual name in the header.** Not the two-up
specification sheet squeezed into a thumbnail — the real wordmark, legible, on every page. I
looked at the live file to confirm it rather than trusting the byte count.

Getting there took a detour worth telling you about, because it is the kind of mistake that
looks like success. I published the corrected logo, the deploy ran green, the cache purge
reported success — and the site kept showing the old picture. I decided the cause was a
slow intermediate server somewhere in the chain, and I was wrong. **We have two different
publishing routes, and each site is on one of them:** some sites are served from a virtual
machine, others from cloud storage. relojistas is a virtual-machine site and I had published
to the storage route, into a folder that exists but that nobody reads. One database column
would have told me. Republished to the right place and it was live in two minutes.

I am flagging it rather than quietly fixing it for two reasons. First, **idea.uk is on the
same virtual-machine route**, and that site now belongs to the other session, so they needed
warning before making the same mistake — I've written it into the shared notes. Second, I had
labelled my wrong explanation as a guess rather than a finding, and that label is the only
reason it did not end up recorded as fact. That habit earned its keep today.

**On the code fix:** the reviewers approved it, but not first time, and their objection was a
good one. My original version only honoured an "approved, do not overwrite" mark on assets
whose status was exactly *active* — and it turns out that status field is free text with no
rules, so an approved item filed under any other status would have been quietly overwritten
anyway. I removed the condition entirely: an approval now blocks overwriting whatever else is
true of the record. They also asked, correctly, whether any *other* part of the system has the
same flaw. It does — the content-card generator — so I measured it (nothing is currently at
risk), wrote it up as its own bug, and left it for a separate fix rather than quietly widening
this one.

**Where it stops for now:** the fixed code is built and reviewed but not yet running, because
publishing the new build to the image registry needs a permission I do not have in this
session. Everything downstream of that — regenerating relojistas' social card and tab icon
with the corrected logo, and the sweep to repair the squashed tab icons on the other nine
sites — is waiting on that one command. Nothing is broken while it waits; the sites are all in
the state they were this morning, plus relojistas' header being right.
