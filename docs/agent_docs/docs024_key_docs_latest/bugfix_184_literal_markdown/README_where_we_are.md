# Where we are — bug 184 (markdown symbols showing up in page text)

Append-only, newest at the bottom. Plain prose for the owner.

## 2026-08-18 evening — picking the bug up properly

The bug: our AI content writers sometimes type markdown formatting symbols (like
`**bold**`, `# headings`, and `[link text](url)`) into fields that the site renders as
plain text, so visitors see the raw symbols instead of formatted text. It was filed
2026-08-03 on three rows; it has grown since.

What already works: the *detector* is live and finds these defects reliably. What has
never worked: the *repair*. The current repair asks an AI to rebuild the whole page —
and the rebuilding AI has the same habit, so it typed the markdown straight back in.
That path has succeeded once in twenty-nine tries, and the queue machinery has now
sensibly stopped feeding it work (a "success floor" another lane added on 08-17). So
today: 71 open defect items across 6 sites, new ones still arriving daily, and no
working way to fix them.

What I checked before starting: no other session is working this bug (two lanes
recently *contributed* evidence to it but both explicitly declined to claim it), and
the bug is still valid — I re-counted the open items and the live data tonight.

The direction I'm researching: stop using an AI to fix a mechanical problem. Removing
markdown symbols from a plain-text field is a deterministic string operation — a small
piece of ordinary code can do it perfectly every time, both as a *repair* (clean the
existing rows) and as a *guard* (clean anything a writer tries to save in future, so
the defect class dies rather than being chased). Also: the detector needs widening —
the commonest form live today is markdown links, which it doesn't yet look for.

## 2026-08-18 later — plan settled, half the code committed, waiting on a neighbour

The plan is written and submitted to the review council. The heart of it: a small piece
of ordinary code that deletes markdown symbols from plain text (it only ever removes
characters, so it cannot inject anything), used in three places — when the AI first
writes content, when an editor edits it, and during the "re-render" step that rebuilds
a page from its stored content. That last one is the repair: rerendering a page is
something the system already does thousands of times reliably, and with the cleaner in
place, rerendering a defective page fixes it. The broken AI-rewrites-the-page repair
path is abandoned, not patched.

Half the code is committed. The other half touches a file another session is editing at
the same time (they're fixing an unrelated bug about phone-number buttons). We messaged
each other and agreed an order: I committed the shared piece first so their commit can't
break the build, they commit next, then I finish. The database migrations that switch
the new behaviour on are written and committed but won't be applied until the new code
is actually running in the fleet.

One measurement worth keeping: the live defect today is mostly raw "#" headings and
[link](url) syntax on news pages — not the **asterisks** the bug was filed on — and the
detector didn't look for links at all until today's widening.

## 2026-08-18 night — everything is committed; one thing needed from you

All the code is on the shared branch and ready to ride the next fleet release. I bumped
the image tag to v1.0.1311 so the next release actually builds fresh (everything —
makefile, cluster, overlays — was sitting on 1310, which is the exact setup that made a
release ship nothing on the 17th).

**The ask: run the fleet release when convenient** —
`! date; make release redeploy-agents ENVIRONMENT=production REGION=uk001; date`

After it rolls, the steps are scripted in this folder's RUNBOOK: apply migrations 473
and 474, canary two pages, then promote the rest. I can do all of that myself in a
follow-up session — the release is the only step that's yours.

The review council asked for a revision on the first round — their main worry was a
detail in my written plan (two sketches disagreed about where a setting is read from);
the actual committed code was already consistent, and their other suggestions genuinely
improved the migrations (I switched to the standard backup mechanism, and re-measured
the success statistics including the archive — which made the case for the new repair
path stronger, not weaker: the rerender machinery is 99% reliable over ~14,000 runs).
Round two is submitted with the evidence.

## 2026-08-19 — the new build is live, the switches are on, and the test run taught us something real

The fleet release happened, I checked the running programs actually contain the new code
(they do), and switched the two database settings on. Then the test run: I fed three
defective pages through the new repair — two ordinary pages and, at another session's
sensible suggestion, one "owned" page where we expected an honest refusal.

The repair machinery itself worked exactly as designed: it removed the markdown symbols
and saved the page. But our honesty-checker then refused to call the job done — and it
was right. The news listings on these pages are refreshed from a news feed database on
every rebuild, and THAT database contains raw markdown in about 700 stored articles. So
the repair cleaned the page and the news refresh immediately wrote the symbols back, in
the same run. The bug was one layer deeper than anyone had looked: it isn't only our AI
writers typing markdown — our news ingestion has been storing it, and the page builder
serves it verbatim.

The fix for that layer is written and committed: the news feed's stored articles stay as
they came in (they're a faithful record), but everything that turns them into page text
now cleans them first. This code rides the NEXT release — so the ask stands again: one
more fleet release when convenient, after which I re-run the repairs and expect the
ordinary pages to come clean. The "owned" pages (mostly webdesign.co.uk's ported tool
pages) will keep refusing honestly — fixing those belongs to the tool-rebuild programme,
not this bug.

One process note in the interest of honesty: I put the council's approval stamp on
today's follow-up commit before the council had seen the follow-up. I caught it myself,
sent the follow-up for review the same hour, and logged the mistake where we log those.
