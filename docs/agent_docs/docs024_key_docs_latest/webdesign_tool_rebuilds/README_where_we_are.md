# Where we are — webdesign tool rebuilds

2026-08-15. The owner asked us to speed up replacing the 63 imported ("ported") tools on
webdesign.co.uk with tools the framework builds and manages itself. Today we found out how
native tools actually get built (a suggester files a work item, a generator builds it), and
discovered nothing was replacing the imported ones — the generator was only adding new tools.
We also cleaned up two hazards first: a shared template that a fixer bot had overwritten with
the wrong tool's markup (restored; nothing had broken yet), and the one page where a native
tool had been dropped next to its imported twin, showing both a leftover and a raw template
tag to visitors (repair queued, leftover retired).

The plan: prove the replacement recipe on one simple tool (the aspect-ratio calculator — its
rebuild is queued), then work through the simple tools one at a time. The rich, hand-built
apps (mind-map studio, meme studio, mini-CMS and friends) are deliberately excluded — an
AI rebuild from a one-line description would quietly lose what makes them good; they wait for
the faithful-conversion route or a per-tool decision.

## 2026-08-16 morning — what happened overnight, in plain words

The pilot did not fail because of the flaky agent handshake I guessed at in the handoff. It failed
because the tool-generator's "save the tool" step can only ever CREATE a page — it has no way to
say "attach this to the page that already exists". Since our whole plan is "rebuild the tool at the
same address", every rebuild would have hit this wall. I had the diagnosis loop confirm it (it did,
first go), built the fix as an opt-in switch (off by default, so nothing else changes), put it
through the council (approved), and committed it. It is NOT live yet: it needs the next chassis roll,
and then a small config seed that turns the switch on for the generator only. Both steps are
written down. Note: at 09:52 this morning the cluster was still running last night's build, so if
a newer one was pushed it had not landed anywhere I could see.

Two other things fell out. First, the "two odd audit items" were not two — a whole class of
audit-fix items had been filed with an EMPTY spec for weeks (233 of 233), because four agent
configs wrote the spec in a shape the platform silently ignores. Fixed at the source last night;
the three items I re-armed all ran and applied within minutes. Second, the ab-test page: the native
copy of that tool turned out to be a hollow shell (its labels and headings were never filled in),
and the page was serving 47 raw template tags. I put the old working tool back on that page and
queued a re-publish. ab-test will be rebuilt properly through the new route, second in line after
aspect-ratio.

The 285 close-out (proving the shared-template fence refuses) is being done by its own lane right
now — I saw the refusal land at 09:59 and left it to them.

## 2026-08-16 afternoon — the owner has settled the rich apps: rebuild them too

The one open question on this lane was what to do with the handful of hand-built applications —
the mind-map studio, meme studio, logic architect, the mini-CMS, the pasteboard — and the tools
whose code sits in separate script files. The plan had them set aside, because a rebuild from a
written description is a fresh implementation rather than a copy, and something is usually lost.
That trade was put to the owner and he has accepted it: **rebuild them as well.** So every one of
the 63 imported tools is now in scope for this route, and nothing here waits on the faithful-copy
conversion any more.

Three things stay exactly as they were, and matter more for these apps than for the calculators:
the description we hand the generator has to be written from watching the live tool work, not from
reading its page; each rebuild is judged before the old one is switched off, because we have
already had one come back as an empty shell; and the old version is retired rather than deleted, so
a disappointing rebuild can be put back in one step. My suggestion, which is the lane's to take or
leave: do these last and one at a time, once the simple tools have proved the process, so the owner
is only asked to look at the difficult ones after the easy ones are known to work.

## 2026-08-16 late afternoon — the first rebuilt tool is built and graded; the old one is switched off

The aspect-ratio calculator is the first tool the framework has rebuilt in place, and the part that
had never worked before now works: the generator attached its new tool to the page that already
existed, instead of trying to create a second page at the same address and dying. That was the whole
point of yesterday's fix, and it did it in 54 seconds.

Before switching anything off I graded what it built, because last time we assumed and got an empty
shell. This one is real: it reduces a width and height to a simplified ratio the proper way, it works
the other direction too (give it a target ratio and one dimension and it gives you the other), and
the 16:9 / 4:3 / 1:1 / 21:9 shortcut buttons are there. All the wording is written into the tool
itself rather than left as blanks for something else to fill in — which is exactly what went wrong
with the A/B test tool.

So the old imported version of that tool is now switched off. Switched off, not deleted: its content
is untouched, and I checked its fingerprint before and after to prove the bytes did not change, so
putting it back is a single flip if anyone dislikes the new one. The page will rebuild itself with
just the new tool on it shortly — it is sitting in a queue about twenty jobs deep. I will check the
live page once it has run; until then the rebuild is done but unproven, and I am not treating it as
finished.

Two things I got told to do that turned out to be impossible as written, both now corrected in the
plan. First, we were meant to record an "archive row" for each tool as the undo handle — there is no
such row, because the archive only records changes to a page's content, and switching a tool off
changes only its status. The undo handle is the old tool's own row, which is better anyway. Second,
the address we were checking the live page at, `/tools/aspect-ratio/`, is a 404 — the real page is
at `/tools/aspect-ratio/index.html`. That matters more than it sounds: every one of our "is it clean
now?" checks passes perfectly against a 404 page, because a page that does not exist contains none of
the things we are looking for.

While waiting I counted up what is left, and found a couple of traps worth knowing about. There are
97 imported pages on the site but only 63 are tools — the rest are learning pages, and one of the
63-looking ones is just the tools index listing. And there are two separate groups of exactly 13
tools that are easy to mistake for each other: 13 whose code lives in separate files (the awkward
ones), and 13 carrying a marker from an earlier repair effort. Only 4 tools are in both groups. If
someone used one list as a shortcut for the other they would get nine tools wrong in each direction.

I also read the generator's code and found something that would have bitten us on the second rebuild:
when it checks "does this tool already exist?", it does not care whether the existing one is switched
off. So the A/B test tool, whose failed rebuild we withdrew this morning, would block its own
replacement — and the run would finish successfully having built nothing at all. That is now written
down with the one-line fix, before it cost us a cycle rather than after.
