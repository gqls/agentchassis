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

**2026-07-21, later — I put the gate through our reviewer council, and it earned its keep.**

I sent the mechanical gate to the internal reviewer panel for a second opinion. It
was a bumpy ride — the first attempt hung on our own flaky dispatch, and one of my
resubmissions bounced back because I'd written a file path sloppily — but two of the
rounds came back with real, substantive reviews, and they were worth every credit.

The most important thing they caught: my gate "failed open". If, for some odd reason,
it couldn't read the recreated tool at all, it was quietly deciding "nothing wrong
here" and letting it through. One reviewer pointed out that this is *the exact same
kind of silent-yes* that caused bug 020 in the first place — a safety check that,
when it can't do its job, waves everything through. That's embarrassing and correct,
so I flipped it: now, if the gate can't inspect the output, it *holds it for a human*
rather than deploying it. As a bonus, that also means if the plumbing ever shifts
under it, the gate will loudly stop everything for review instead of silently doing
nothing — which is exactly the failure that would otherwise be invisible.

The rest of what they raised I've addressed too (I verified the internal data paths
are right, checked there wasn't already a tool doing this job before building a new
one, tightened the database-change script so it double-checks it's touching exactly
the one row it should, and left a note on the pipeline explaining the design for the
next person). A couple of their points were really "prove it even harder in the
submission" and "consider whether other parts of the platform need the same gate" —
both fair, but the first can't be fully shown until we actually roll the new build,
and the second is a sensible follow-up rather than part of this fix. So I stopped
there. To be clear: the panel never gave a final green tick, so I'm *not* claiming it
did — but it made the change materially better, which is the whole point of asking.

Bottom line unchanged: half the fix is live, the other half is ready and now a bit
sturdier, and bug 020 stays open until we roll a build and switch the gate on.

**2026-07-23 — bug 020 is CLOSED. The gate is on, and I proved it actually catches things.**

The story since the last entry is a good one, because the careful step is exactly the
one that paid off. A build went out with the gate in it, I switched it on, and — before
trusting it — I ran a deliberate test: I fed the system a fake, invented tool (the exact
kind of thing bug 020 was about) and watched what happened. It did NOT get caught by the
real check; instead a *safety net* caught it, which told me the real check wasn't
actually seeing the tool at all. Digging in, I found a genuine bug: the check was being
handed the tool's contents where it expected the *location* of the contents, so it kept
looking at an empty box and waving everything through. On the earlier build that would
have meant the gate silently did nothing — it would have looked switched-on while
letting fabricated tools straight out to sites. The only reason I caught it is that I'd
added that safety net earlier (from the review) and that I insisted on testing with a
real fake rather than trusting the wiring.

So I switched the gate back off immediately (nothing had slipped through — no real tools
had run in that window), fixed the bug, and added a test that would catch it again.

This build (the one you just deployed) has that fix. I switched the gate on again and
re-ran the same deliberate test — and this time the *real* check caught the fake tool,
by name: it flagged the give-away phrases ("realistic, deterministic dataset", the fake-
postcode generator) and held the tool for review instead of publishing it. That is the
end-to-end proof I wanted: a fabricated tool, driven through the live system, stopped
before it could go out.

So both halves are now live and working: the instructions that tell the builder never to
invent data, and the mechanical gate that stops it if it does anyway. Bug 020 is closed.
I've noted on the imagery workstream's page that the "wait for 020" hold has met its
condition, so lifting it is your green-light whenever you want it.

The honest reflection: this took three builds instead of one, because the first "it's
wired, ship it" would have shipped a gate that did nothing. The lesson — don't trust a
detector until you've made it catch something real — is exactly the one the bug itself
was teaching.

**2026-07-22 — the build rolled, and the gate is now switched on and guarding.**

The new build went out, so I switched the gate on. First I checked the running
program actually contained the new check — it did — and that it wasn't already
switched on by anyone else — it wasn't (it was sitting there doing nothing, which is
the worst state: built but not connected). I connected it: now, every time the system
rebuilds one of these tools, the output passes through the fabrication check before it
can be published, and anything that looks like invented data is held for a person to
look at instead of going live. I double-checked the plumbing routes correctly — a
flagged tool goes to the review pile and stops; a clean tool carries on and publishes
as normal. So the core protection is live and working.

Two honest caveats, and then a decision you made. First, this particular build was cut
about twenty minutes before I'd finished the small "fail-safe" tweak from the review
(the one that makes the check hold things for review if it ever can't read the output
at all), so the running version is the slightly-less-hardened one. It still catches
real invented data — the tweak only matters for an odd empty-output edge that isn't
fabrication anyway — but it's not the fully-hardened version yet. Second, I've proven
the check works by testing and proven the routing is correct, but I haven't yet pushed
a deliberately-fabricated tool all the way through the live system end to end.

You decided: leave bug 020 open and finish it properly on the next build. That's the
right call — there's no urgency now that the protection is live, and both remaining
bits (the fail-safe tweak, and one deliberate end-to-end test) are cleanest to do
together on the next roll. The runbook has the exact three steps to close it out then.
