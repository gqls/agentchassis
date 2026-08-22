# SUMMARY — 2026-08-22: the release coverage gate, inverted

Written to be read aloud.

## What we're trying to do

Make it impossible for one of our own services to be quietly left behind by a release.

Every service runs from a container image, and a release rebuilds those images and points
the cluster at the new ones. A service that is not on the release's list is simply never
moved again — and nothing complains, because the pod keeps running perfectly. It is just
running last month's code. There is a check that is supposed to catch that. This work is
about the fact that the check could not see the very thing it was built for.

## Where we've come from

In August a service called `render-audit-adapter` was found sitting eighty-six versions
behind the rest of the fleet, serving a binary with a known credential leak, for months.
The lane that fixed it did the right thing: it replaced four hand-written lists with one,
and built a gate that walks the deployment folder and refuses to release if a service is
missing from it.

Then, chasing the same shape, they found six more frozen services and the owner ruled all
six into the release. But their own write-up recorded, honestly, that the gate could never
have found those six — because of how the gate asks its question. They filed that as its
own bug, `318`, three days ago.

Two more services fell into the hole in the three days since. Both were created by people
who had the owner's ruling in front of them.

## What we've done

**First, we found the next release was already broken and fixed it.** Three of the newest
services had been added to the list of images the release *ships* and to none of the lists
that say how to *build* them. The release would have built twenty-two images, taken six
minutes doing it, and then stopped dead trying to upload one that had never been built —
before the deploy step, so nothing would have reached the cluster. We did not fix that by
adding three names, because two hand-written lists that must agree is the shape of the bug
itself. The build list is now *derived* from the ship list. There is one list.

**Then we inverted the gate.** It used to ask "is this service's image one the release
builds?" and skip the service when the answer was no — which is exactly what being left
behind means, so the one case it existed to catch was the one case it stepped over. Now an
image of ours that no release builds is a failure, and the only way out is to say so
explicitly, in a list that names what moves that service instead.

**We moved it out of the makefile into Go**, and that was not tidiness. The review council
only reads certain directories and the makefile is not one of them, so the old gate had
never been reviewed by anyone and could not have been. It also could not be tested safely:
proving it worked meant editing a file forty sessions share, and someone did exactly that
this week and had their half-finished edit committed by a stranger before they could put
it back.

**The council approved it first time**, and two of its three notes were good enough to
change the code rather than argue with. One reviewer noticed that our new scanner read
only the first image in a file and silently ignored a second — which is precisely the
disease this whole change exists to cure, reproduced inside the cure. We checked: no file
on the estate has two images today, so it was a trap waiting rather than a live fault. We
removed the limit rather than adding a warning about it.

**And we added a second, gentler check that fires at the moment the mistake is made**,
rather than at the next release. We measured it against every commit in the last five
weeks that created one of these files: twenty-one of them. It fires on seven, and all
seven are the known incidents. It stayed silent on the other fourteen — including one
created correctly by another team an hour earlier, which is a pass that could have failed.

## Where we are now

The gate is built, reviewed, and in place. It judges thirty-one files and passes. The
exemption list was going to be empty and the gate's first run found a real entry for it:
the admin dashboard is genuinely outside the backend release and genuinely still shipped,
and nothing anywhere on disk had ever said so.

One thing we tried and could not build, which is worth knowing. The owner asked for a rule
that fires when a service's version is older than the last change to its own code. The
file that records a service's version is edited by the release itself and never committed,
so its history is fiction — twenty-six of them are uncommitted right now, and the chassis
reads eighty-four versions behind what it is actually running. That trap is now written
down where the next person will find it before they build on it.

## Where we're going

Two things, and one of them is the owner's.

The **owner's decision** is what to do about that rule. The honest version asks the
running service what code it is made of, which we have been able to do since the
build-provenance work earlier this month. But now that every service of ours is either in
the release or explicitly excused, the rule could only ever apply to the excused ones —
and there is one. So it may be worth less than it was when he asked for it, and that is
his call, not ours.

The **remaining work** is a daily census that looks at the cluster rather than the
filesystem, because there are two shapes no file-based check can see: a service that
exists on disk and was never switched on, and one running in the cluster with nothing on
disk describing it. We found one of each while measuring. That earns its own round.

And the gate has not yet met a real release. The reviewer who signed it off said plainly
that this is the single point every deploy passes through and that the first run should be
watched. That is right, and it is the close condition.
