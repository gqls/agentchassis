# SUMMARY — bug 046 truncation casualties, detection now live (2026-07-22)

## What we're trying to do
Some interactive tools on our live sites were built from a half-finished AI
generation — their JavaScript is cut off mid-sentence, so the tool is broken and
the rest of the page after it breaks too. We'd fixed the *cause* of this a while
back, but never went looking for the tools already damaged. Bug 046 is that
clean-up: find them all, repair what we can, and — most importantly — make the
platform able to *notice* this class of damage on its own, because right now
nothing does.

## Where we've come from
The original truncation fix repaired exactly one tool, by hand, because that's
the one someone happened to be watching. A census then found **nine** damaged
components still live across six sites, serving broken JavaScript, and invisible
to every check we had — because the damage sat identically in both the "recipe"
and the "baked" copy, so there was nothing to compare and flag.

## What we've done
1. **Built a detector.** A new "sweep" that walks each site's components and
   flags any whose markup is a cut-off generation. Tuned against every live
   component: it catches exactly the nine bad ones and zero good ones. It does
   **not** try to auto-repair (our tool-rebuilder can invent fake data — a
   separate known bug), so instead each find becomes a tracked review item that
   says up front whether there's a clean older version to restore or a full
   rebuild is needed.
2. **Repaired the one cheap casualty.** The grip-force calculator was the only
   one with a clean older version saved, so we restored it. Count dropped 9 → 8.
3. **Shipped and switched the detector on.** It reached production in the latest
   server build; we confirmed it's genuinely running, enabled it, and then ran it
   for real against a broken site and watched it correctly flag the broken tool
   (with the right repair advice) while correctly ignoring a broken-but-unused
   component. So it's proven working, not just deployed.

## Where we are now
The detection half is **done and operational**: from now on, any tool damaged
this way surfaces itself as a tracked item the moment its site is next swept.
One casualty is repaired at source. Eight remain damaged.

The honest caveat: repairing the *recipe* doesn't fix the *live page* by itself —
the page has to be re-rendered, and that re-render pipeline has its own open bug
another team owns. So the live pages are still broken to a visitor; what we've
guaranteed is that they'll come out right when they do re-render.

## Where we're going
Two things need an owner's steer: (a) how to rebuild the eight remaining tools
without tripping the fake-data trap, and (b) getting the repaired recipes
delivered to the live pages through the re-render pipeline (the other team's
bug 024). Neither is safe to automate blindly, which is exactly why the detector
surfaces them for a decision rather than firing off rebuilds itself.
