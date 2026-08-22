# Where we are — imagery that gets made and never used (bug 114)

Plain-prose log for the owner. Append-only, newest at the bottom.

---

## 2026-08-22 — what this is, what we found, and what has changed today

**The complaint, originally.** Back in July you said there was not enough imagery on a
site. There was: twenty-one images had been generated and put on the server, and the site
referenced three of them. That became bug 114, and since then five other threads have
added evidence to it without anyone taking the fix. I have taken the framework half —
the part that would stop it happening again on every site, rather than repairing one more
site by hand.

**First: it is still true, and it is bigger.** Ten sites carried the bad default when
somebody last measured in July; today it is eighteen. Across the fleet there are 94
"content hero" images — one made specifically for one page — and only 23 of them are
actually on the page they were made for. On mortgagecalculator, where ten were generated
in one night on your instruction back on the 15th, two are on their pages and eight are
not.

**Then the thing I did not expect.** One of those sites, fundamentallyai, had been
repaired by hand on 29 July. Someone found the site was pointing at a picture file that
did not exist, worked out what it should point at, and fixed it. Today it is pointing at
the missing file again. Nobody undid it.

It turns out the pipeline was undoing it. Every time the system stores a generated image,
it also writes down "the site's main picture is at this address" — and it was working out
that address from the *kind* of image rather than from the image itself. So a picture made
for one particular page would announce itself as the site's main picture, at an address
nothing had ever been saved to. Generate a hero for the About page, and the whole site's
default quietly re-points to a file that does not exist.

The clearest evidence is almost silly: I listed what every site had recorded as its main
picture, and **every site had recorded exactly the same answer**. Same for icons, same for
logos. Eighteen different sites, one identical address — because the address was never
derived from the site at all. Two of those addresses (`content_hero.jpg`, `icon.jpg`) are
names the system is incapable of ever producing, and I checked: they return "not found" on
every site that lists them.

**And the safeguard already existed.** Whoever built the image handler had written
"don't touch the site's main picture settings" into its configuration. The instruction was
sitting there, correctly, all along — and no code ever read it. The setting was designed,
written down, and never wired up.

**What has changed today.** Two commits, both waiting for the next system build:

1. The store step now honours that instruction, in both directions, and works out the
   address the same way the thing that actually saves the file does. A picture made for one
   page can no longer redefine the whole site. A picture that genuinely *is* the site's
   main one still can — and now records the address that exists rather than the one that
   does not, which is exactly the repair fundamentallyai was given by hand.
2. When a page falls back to the site-wide default, it now says so in the logs, and says
   whether it even had the chance to find its own picture first.

The second one needs explaining, because it is a fix for a different kind of problem. On
mortgagecalculator, two pages out of eight got their own picture and six did not, on the
same night, through the same machinery. I could not work out why, and neither could our
diagnosis system, because the detailed records of that night have since been deleted on
the normal retention schedule. Falling back to the site default is perfectly correct
behaviour for an older site that has no pictures of its own — and it is also exactly what
this bug looks like. The two were indistinguishable after the fact. So rather than guess,
I have made the system say which of the two it is. Next time it happens it is one search
instead of a dead end.

**I was wrong twice today, and both are written down.** I thought the difference was which
of two processes handled each page — our diagnosis loop disproved that. I then thought a
page that already had a stale setting would keep it — which fitted the eight pages in
front of me perfectly, until I looked at every page on the fleet and found ten that had
that stale setting and got the right picture anyway. Both are in the lane's notes and in
the fleet-wide log of wrong calls, along with the cheap check that would have caught each
sooner. The second one is the useful lesson: the refuting evidence was one query away, and
I had already written the wrong reading into my notes before running it.

**What I have deliberately not done.**

- Not repaired the eighteen sites yet. The repair is written but held back until the fix
  above is actually running, because applying it first just invites the next generated
  image to undo it — which is the mistake that cost us the July repair.
- Not run a test page-rebuild on mortgagecalculator. The plan called for it; on reading
  what that actually triggers, it would have rewritten the page's text while another
  thread is actively working that site. Reading the existing data answered the same
  question better, and gave me eight cases instead of one.
- Not started generating the missing case-study images. Somebody would have to widen what
  the generator considers, which is fleet-wide picture-generation spend — **your call, not
  mine.** Same for five stalled jobs whose pages have nothing built to re-render.

**One thing you should know about, which is not mine to fix.** The part of the system that
was supposed to finish this job — link each generated picture to its page — only runs when
a particular daily sweep visits a site. That sweep has not run since 11 August. Its four
sibling sweeps are all current. We have a watchdog that reports this **every single day**,
and the report goes nowhere anybody reads. I have recorded the measurement against the bug
that owns the watchdog rather than fixing it here, but it is the reason a one-off batch of
generated images never gets connected to anything, and it is why the remaining work in
this lane is designed to link the picture at the moment it is made rather than waiting for
a sweep that may never come.

**Next.** Parts two and three: link each image to its page when it is generated, file the
follow-up work immediately rather than waiting for the sweep, and add a check that notices
"this page has its own picture sitting unused". All specified in the lane plan. Part one
is submitted to the review council and is waiting on a verdict.
