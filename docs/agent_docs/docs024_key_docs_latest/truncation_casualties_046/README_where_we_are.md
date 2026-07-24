# Where we are — bug 046, the truncation casualties

Plain-prose log, newest at the bottom. Append; don't rewrite.

---

**2026-07-21 (bugfix-046 thread)**

The story so far: a while back we fixed the *cause* of a nasty bug where a
half-finished AI generation would overwrite a good tool component with a cut-off
one — the tool's JavaScript ends mid-sentence and the page breaks. We fixed the
cause, and by hand we repaired the *one* tool we happened to be watching at the
time. What we never did was go and look for the others. Bug 046 is that census:
**nine components, across six of our sites, were still broken and still live** —
serving broken JavaScript to real visitors — and nothing we had would ever tell
us, because the broken version was sitting in both the "recipe" and the
"baked" copy, so there was nothing to compare against and notice.

I picked this up today and did two things.

**First, the lasting fix: I taught the platform to notice.** There's now a
"sweep" that walks every site's components and flags any whose HTML is a cut-off
generation — an opening `<script>` with no closing one, and so on. I calibrated
it against every live component we have: it catches exactly the nine bad ones and
zero good ones. It doesn't try to auto-repair them, on purpose — because our
tool-rebuilder has a nasty habit of *inventing* data when it can't find the real
thing (that's a separate known bug, 020), so quietly telling it to "rebuild all
nine" could ship made-up numbers to more sites. Instead each one it finds becomes
a tracked review item, and it tells the reviewer up front whether there's a clean
older version to restore (cheap) or whether it needs a full rebuild. That code is
written, tested, and committed — it'll go live the next time a new server image
is rolled out and the one-line "switch it on" is applied.

**Second, the cheap repair: I fixed one tool at the source.** Only one of the nine
— the grip-force calculator on robot-hands.com — had a clean older version saved,
so I restored it. Our count of broken components dropped from nine to eight.

The honest caveat, which I want to be clear about: **restoring the recipe does not
by itself fix the live page.** The page still has to be "re-rendered" from the
good recipe, and that re-rendering pipeline has its own open bug (024) that
another thread is actively working. So as of right now the grip-force page is
still broken to a visitor; what I've done is make sure that when it *does*
re-render, it'll come out right. I deliberately did not go poke the re-rendering
machinery — it's someone else's live workbench and going in blind risks making
things worse.

The remaining eight all need a genuine rebuild (no clean old version exists), and
that's the fabrication-risk area, so I've left those as tracked items for a human
to steer rather than firing them off automatically.

So: the class of bug is now *visible* going forward, one casualty is repaired at
source, and the rest are catalogued with clear next steps. What still needs an
owner decision: (a) roll the image + switch the sweep on, and (b) how to rebuild
the eight without the fabrication trap.

---

**2026-07-22 (bugfix-046 thread)**

The new chassis build went to production, which is what the sweep I wrote was
waiting for. I checked it was genuinely in the running server (not just tagged),
switched it on, and — importantly — didn't just trust that it was wired up: I ran
it for real against one of the broken sites (vonc.com) and watched it correctly
flag the broken "arena" tool, with the right details and the right "here's how to
fix it" note, while correctly leaving alone a broken-but-unused component that no
visitor can reach. So the detection half of this is now genuinely done and
working, not just written.

What's left is unchanged and still needs owner steering: the eight remaining
broken tools need rebuilding (the fabrication-risk area, so a human should drive
it), and the live pages need re-rendering through the pipeline another thread
owns. But from today, any new casualty like this will surface itself as a tracked
item automatically — which is the thing that was missing when this bug was filed.

---

**2026-07-24 (bugfix-046 thread)**

Good news on two fronts. While I was away, the two things that were blocking the
actual repair both got fixed by other threads: the "how do we get a fixed tool
onto the live page" problem (bug 024) is solved — there's a sanctioned path now —
and the "the tool-rebuilder invents fake data" problem (bug 020) is fixed too. So
the road is clear.

I used that to finish off the grip-force calculator on robot-hands.com properly.
Last time I'd repaired its recipe but the live page was still broken. Today I
pushed the repair through the sanctioned delivery path and confirmed on the live
site: the broken (never-closed) script is gone, the tool's JavaScript is now
whole. That's the first one completely fixed, front to back — and it proves the
delivery recipe works, which is the thing that matters for the rest.

The remaining eight are the harder ones: none has a clean old version, so each
needs a genuine rebuild by the tool-generator (now safe, thanks to the fake-data
fix), then the same delivery step I just proved. That's a bigger, LLM-heavy job
that changes eight live customer tools and produces new designs for each, so I've
stopped here to check how you'd like to proceed rather than firing off eight
rebuilds unprompted.
