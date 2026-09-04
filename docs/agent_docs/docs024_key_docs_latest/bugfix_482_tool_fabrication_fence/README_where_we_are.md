# Where we are — the tool fabrication fence (bugs_open/482)

Plain-prose running log, append-only, newest at the bottom. Owner's document.

---

## 2026-09-04 — picking the bug up, and finding it was a bigger one than filed

Bug 482 was filed this morning by the `calendar` lane after the owner spotted that
boxingonline.com's fight-countdown tool was counting down to fights that had already
happened. It was filed unowned: the lane that found it said plainly that fixing it was not
their job, and the two other lanes that cited it said the same. Nobody was working it, so
this session picked it up.

**The bug is real and still live.** Both tools are still active, still serving. The countdown
has six boxing matches written directly into its own JavaScript — real fighters, real venues,
completely invented pairings — and every one of them is dated 2025. Today is September 2026,
so every option the tool offers is over a year in the past, and the tool's own logic replies
"this fight has started or concluded" to all six. It is not partly broken; there is no option
that works. Meanwhile the site genuinely does hold a real, checked, upcoming fight — Canelo
vs Mbilli on 31 October — and the countdown has no connection to it whatsoever.

**Then the thing that changed the shape of the job.** Looking for what *should* have caught
this, I found that the platform already has a fabrication detector. It was built in July, for
almost exactly this problem, on almost exactly this kind of site. It has been reviewed by the
council, it is careful and well-written, and it is switched on — but only on one of the three
routes by which a tool's content can be written. Tool "birth", the route these two tools came
through, does not consult it at all.

So I ran that detector by hand against the offending tools. It scores them clean.

**That result got sharper the more tools I fed it.** The detector does not fail to notice
these tools. On three of them it *does* notice — it records "large literal record array (~30
entity objects)" and then returns a verdict of *not fabricated* anyway. The reason is a
condition it inherited from the job it was built for: it was designed to catch a tool that
was *rebuilt* from an original, so before it will convict, it insists on checking that the
original loaded real data and the rebuild dropped it. A tool that was born rather than
rebuilt has no original, so that check can never pass, and the verdict is thrown away. **The
evidence is computed correctly and then discarded.** That is a much smaller thing to fix than
"build a new checker", which is what everyone including me assumed was needed.

**And the census found something worse than the bug I was sent to fix.** I ran a fleet-wide
count first and got a reassuring answer: one tool out of 335. That answer was wrong, and it
was wrong because of how I asked. I searched for *dates*, because the bug in front of me was
about dates. The `427` lane searched for something better — records that name a real thing
and claim a checkable fact about it — and found several. The worst is on **vetcomparison.uk's
home page**: thirty invented veterinary practices, with invented postcodes and invented
websites, live and deployed since yesterday. The tool describes them in its own code as a
"verified sample". It carries a note to visitors saying the details may change and they
should confirm anything important directly with the practice, and it points at the RCVS
register while saying so. It is inviting the public to go and check the details of practices
that do not exist.

I have told that site's thread, with the evidence, and touched nothing of theirs — taking a
live commercial site's content down is that thread's call and the owner's, not mine.

There is a bitter footnote. The bug that caused the fabrication detector to be built in the
first place was filed about vetcomparison.uk. The detector was written for that site, and
that site is still doing it — in a shape the detector was never taught to convict.

**Where this leaves the job.** Not "write a checker for invented boxing fights". The two
things worth doing are (1) stop throwing away a verdict we already correctly compute, and
route every tool-writing path through it rather than one of three, and (2) make it impossible
for a *future* route to appear unfenced without somebody noticing — the platform already has
a neat trick for exactly that, used elsewhere, which fails the build when a new writer turns
up that hasn't been signed off. I have a plan under adversarial review now.

**Deliberately not decided by me:** what happens to the tools that are already live and
wrong. Repair them from real facts, withdraw them, or ship them honestly empty. On
boxingonline that is being put to the owner; on vetcomparison it is that site's thread. One
thing I have flagged: the two boxingonline tools want *different* answers. The countdown has
two real fixtures available so repairing it is genuinely possible. The comparator needs
fighter data that does not exist anywhere on the estate, so its only honest options are to
withdraw it or show it visibly empty. Offered a single yes/no, the right answer for one is
the wrong answer for the other.
