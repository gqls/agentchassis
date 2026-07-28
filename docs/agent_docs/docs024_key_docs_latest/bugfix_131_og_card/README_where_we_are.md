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
