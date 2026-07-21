# Where we are — bug 020 (the tool that invented a directory)

Plain-prose running log, append-only, newest at the bottom.

---

**2026-07-21 (bugfix 020 thread).**

The problem, in plain terms: when we adopt a website that has an interactive tool
on it (a calculator, a game, a searchable directory), the platform rebuilds that
tool from scratch. That mostly works. But when the tool's whole point is the *data
behind it* — like vetcomparison's search box over 2,109 real vet practices — the
rebuilder got confused. We'd told it to make the tool "self-contained", and it
took that to mean "make up the data too". So instead of keeping the link to the
real list of practices, it wrote a little program that *invents* fake practice
names and fake postcodes in the browser, and shipped that to the live public site.
It even left a comment admitting it: "we generate a large, realistic, deterministic
dataset". Everything reported "complete". Nothing warned.

The vetcomparison team already cleaned up their own site. My job was the underlying
platform bug, so it can't happen to the next site.

I fixed it in two halves.

The first half is live right now. I rewrote the instructions we give the rebuilder.
There's now a big "Data Integrity" rule at the top that says, in effect: never
invent records; "self-contained" means keep all your *code* in one file, it does
NOT mean bake in the *data*; if the original tool loaded a list from somewhere,
load from that same place; and if you genuinely can't reach the data, show an empty
"no data yet" state and stop — an empty tool is fine, a fake one is a serious
fault. I also made the earlier "analysis" step write down exactly where the
original tool got its data, so the rebuilder is told plainly. This is a database
change, so it's effective immediately with no deploy.

But a written instruction isn't enough on its own — the old rule was ignored, and
you can't trust "it'll obey this time". So the second half is a mechanical check
that doesn't rely on the rebuilder behaving. It reads the finished tool and, if it
spots the tell-tale signs of invented data, it refuses to publish and instead
raises a "needs human review" flag with the evidence attached. The hard part was
making it *precise*: lots of perfectly good tools use randomness (every dice game
does), so I couldn't just flag "uses random numbers". The check only trips when it
sees the actual fingerprint of this bug — the original tool loaded real data, and
the rebuild threw that away and manufactured a fake list instead. I wrote eleven
tests: it catches the real vetcomparison fabrication (even with the confessing
comment removed), and it leaves a dice game, a calculator, a faithful rebuild, an
honest empty state, and a name-generator tool alone.

That second half is written, tested, and committed, but it's Go code, so it only
comes alive when the platform's main program is next rebuilt and rolled out —
that's an owner decision, not something I do mid-stream. I've sent the change to
our reviewer council for a second opinion, and I've staged the final wiring step so
it's ready to switch on the moment the new build is live.

So: the invention is much less likely to happen at all now (half one, live), and
once the next build ships it becomes genuinely unable to publish (half two). Until
that build ships, I'm leaving bug 020 marked OPEN — the honest bar here is "fixed
AND live", and half of it isn't live yet.

One more thing worth saying out loud, from a sister thread the same day: the locks
we'd put on vetcomparison's corrected components to protect them turned out not to
survive a full page rebuild — the rebuild replaces the component rows wholesale, so
a lock attached to the old row just vanishes. It was harmless this time, but it's a
good reminder of exactly why 020's fix has to live in the rebuilder's contract and
in a publish-time gate, not in a flag stuck on a row that a rebuild throws away.
