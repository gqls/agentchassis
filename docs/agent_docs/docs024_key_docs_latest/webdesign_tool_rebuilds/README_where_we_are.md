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
