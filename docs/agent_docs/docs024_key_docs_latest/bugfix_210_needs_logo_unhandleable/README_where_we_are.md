# Where we are — the logo/hero image jobs that can never run

Append-only, newest at the bottom. Plain prose.

---

## 2026-08-09 — picking this up, and what turned out to be going on

I took on the bug about `needs_logo` jobs failing. The short version of what was already known:
when the system notices a site is showing the stand-in logo instead of a real one, it raises a
job to generate one. That job hands the image generator a prompt. If nobody ever wrote a logo
prompt for that site, the job is raised anyway — with no prompt — and the generator refuses the
handover and dies. It has never once worked in that state.

The person who filed it did the right thing and stopped short of fixing it, because the obvious
fix is a trap: you can make the prompt optional in one character, and then the generator quietly
falls back to the sentence *"Generate content based on the provided context."* and paints
whatever that inspires — and the system saves the result **as the site's logo**. A loud failure
becomes a silently wrong brand. So the easy fix is worse than the bug.

Digging in, four things came up that change the picture.

**One: the job that started all this should never have existed.** The check that spots a
stand-in logo looks for the text `/assets/images/logo.png` anywhere in the page. On
fundamentallyai.com the only place that text appears is in a link to *someone else's* logo, on
someone else's website — a partner logo in a portfolio strip. The check saw the address and
assumed the site was talking about itself. Across the whole fleet that is the **only** match for
a stand-in logo there has ever been, and it is a false alarm.

**Two: the problem is nonetheless real, and there is a genuine case.** mortgagecalculator.co.uk
really is showing the stand-in hero image on six pages, really has no hero image of its own, and
its job really did die the same way. So this is not just a false-alarm story.

**Three: it is much more widespread than "two sites".** The check tries to recover a prompt from
the site's plan. Thirty-four of our thirty-nine sites have no usable plan prompts at all — most
have no current plan record whatsoever. So the "recover the prompt" path fails on about
seven-eighths of the fleet. The only reason we have seen two failures instead of thirty-four is
that the *other* condition — the site having no image at all for that slot — is currently rare.
Thirteen sites are one deleted image away from the same dead end. And a second, entirely
separate part of the system (the one that sets sites up in the first place) can raise exactly
the same broken job; it just has not happened to yet.

**Four: the check is blind to the place a logo actually lives.** Logos sit in the site's shared
header, not in the page content, and this check only reads page content. Its twin check was
fixed for this months ago and this one was left behind. I want to be careful not to oversell
that: four sites do show the logo address in their header, and all four already have a real
logo, so today it is costing us nothing. It is a hole waiting rather than a hole leaking.

**What I want to do about it**, in order of how firmly each one shuts the door:

- Stop counting someone else's web address as our own page showing a stand-in. That alone
  removes the only false alarm we have.
- Read the site header as well as the page content, so a genuine missing logo is actually seen.
- **Make the image generator refuse to generate from the meaningless fallback sentence.** This
  is the one I care most about. It does not matter which part of the system forgot the prompt,
  or which part someone writes next year — at the single point where every image is made, a
  meaningless prompt stops being something that can produce a saved brand asset. I checked
  before proposing it: of the 344 images we have with a recorded prompt, **none** was made from
  that sentence, so refusing breaks nothing that has ever worked. (Fifty-five older images have
  no prompt recorded at all, so I cannot see those — worth saying rather than glossing.)
- Stop raising a generation job when we know there is no prompt for it. Raise something a person
  can act on instead.

**What I am deliberately not deciding.** The real question underneath is: when a site needs a
logo and nobody ever said what it should look like, who decides? That is not mine to answer, and
the filing session was right to leave it. I did find something that bears on it: the codebase
already deliberately refuses to auto-generate any styling direction for logos — there is a note
from May about generated logos getting contaminated by the site's photographic style — and the
intended life of a logo is "made once, approved by a human, then locked". So sending an
unplanned logo to a person is not us giving up; it is the path the code already assumes. That is
worth having in front of you when you decide.

Next: this goes to the review council, then I build it.

## 2026-08-09, later — council verdict and what shipped

The council approved it first time round, which I did not take for granted, and two of its
twelve reviewers earned their keep.

One caught a flaw in **how I had measured**, not in what I proposed. My case for the big change
rested on "only one part of the system does image generation, so making it stricter is safe" —
and the way I counted only looked at the top level of each agent's configuration, missing
anything nested deeper. That is a known trap here and I walked into it. I re-counted properly.
The answer held: still exactly one. And I checked the new method could actually have told me
otherwise — asked the same question about a different, common operation and it found seven
things the old method missed. So the safety claim survived a real test rather than a
sympathetic one.

The other asked me to stop calling something a "stated residual risk" and go and look at it. I
had said 344 of our images have a recorded prompt and none was made from the meaningless
fallback, while 55 have no prompt recorded and I couldn't see them. Fair challenge. Looking:
47 of those 55 were never *generated* at all — they are derived from another image (cards from
a hero, favicons from a logo), so they never go near the code I changed. The remaining 8 are
all from January to early March. There was no hole.

Separately, and before the council came back, the codebase's own tests caught a mistake in my
fix. I had routed the "needs a human" item to something called `human-review`, which sounded
right. A test objected that no such agent exists, and it was correct — the platform's real
convention is to leave the handler blank, which 433 live rows do and only 12 use the name I had
picked. Fixed. Two council reviewers raised the same point independently, so three separate
checks agreed; the tests just got there first.

Two things I got wrong along the way are written up properly in the fleet-wide log, because
both were near-misses rather than harmless. I nearly restored an unrelated piece of work from
five branches ago into our shared working copy — a `git stash` command failed quietly and the
next one grabbed somebody else's parked work instead of mine. It only stopped because of a
lucky collision. And I spent a while convinced I had broken a test that another session had
already fixed, because I compared "before" and "after" copies of the code taken a few minutes
apart, and the code moved underneath me in between.

**What is done, and what is not.** The code is committed and reviewed. It is **not live** —
nothing changes until the next chassis image is built and rolled, and I have written down
exactly how to prove it arrived (look inside the running program, not at the version tag).
The open question is still yours: when a site needs a logo and nobody ever said what it should
look like, who decides? I found something that bears on it — the system already refuses, on
purpose, to invent styling for logos, and expects a person to approve one — so sending it to a
person matches how the rest of the code thinks. But I have not decided it, and it should not be
decided by whoever happens to be fixing the next bug.
