# Where we are — chrome divergence guard (bug 226)

Append-only, newest at the bottom. Plain prose.

## 2026-08-08 — picked up, plan formed

We took on bug 226. The short version of the bug: the site header and footer
are stored once as finished HTML, and any time someone fixes something directly
in that stored HTML — the way the honesty note on oufe was added — the next
routine rebuild deletes the fix without telling anyone. It happened twice on
oufe and nobody noticed for eight days either time.

The plan has three parts. First, the database itself will keep a copy of the
old header or footer every time one is replaced with something different — so
nothing can ever be silently lost again, no matter who or what does the
replacing. Second, each time the platform writes a header or footer it will
leave a small stamp saying "the machine wrote these exact bytes"; if the bytes
on record stop matching the stamp, we know a person patched it by hand, and the
rebuild will say so loudly and file it for review instead of quietly steaming
over it. Third, the rebuild still goes ahead — locking things down is a
different, existing feature — this is about never losing work and never being
silent about it.

One correction to the bug as filed: it suggested re-running the old render to
compare — that isn't possible with what we store (we keep fingerprints of the
ingredients, not the ingredients). The stamp-the-bytes approach gets the same
answer more cheaply.

Timing matters: a separate fix (bug 117) will trigger a big wave of chrome
rebuilds on the next release. Our database half goes live immediately when
applied, before that wave — so the wave becomes the first thing the new safety
net catches rather than the last thing that slips through it.

Next: council review of the plan, then the code.

## 2026-08-08, later — safety net live, council asked for changes, changes made

The database half is done and live: from this evening, nothing can replace a
stored header or footer with different content without the old version being
kept. We proved it on a real row before trusting it — patched one, watched the
copy appear, put it back, cleaned up.

The reviewer council looked at the whole plan and said "revise". Some of their
worries turned out to be wrong once we measured — for instance, the fear that
only the first site to lose content would ever get a review ticket isn't true,
because tickets are already filed per site. But three worries were right and
we fixed all three: a second loss on the same slot could have been mistaken
for a duplicate of the first and dropped (each loss now gets its own ticket);
a ticket could have been filed even when a protective lock had actually
stopped the rebuild (the ticket now only files when something was really
replaced); and our promise to deal with the same problem on ordinary page
sections "later" wasn't concrete (it is now bug 229, written up properly with
its own evidence).

We resubmitted with the fixes and the measurements. The code is committed and
will be in the next release; the bug stays open until we've seen it working on
the running system, and the checklist for that is written into the bug file.

## 2026-08-09 — it's running, and it already saved something

The new release went out and we checked it properly, on the actual running
machines, including a check that the OLD version's fingerprint is gone — so we
know the right code is live, not just that a release happened.

Better: overnight, before anyone tested anything, two sites had their headers
and footers rebuilt in the normal course of business — and the safety net
quietly kept copies of what was replaced, exactly as designed. That's the
first real proof, unprompted.

The reviewers came back a second time. One embarrassment on our side first: I
had set up an overnight alarm to tell me when their answer arrived, and got
the timezone wrong, so the alarm could never ring — that's written up in the
mistakes log. Their answer: most reviewers now approve; one small code gap
they found is fixed (if the "was this hand-edited?" check hiccups, we now ask
the ledger itself, which always knows); and one genuine disagreement remains
between two reviewers — should we fix the same problem for ordinary page
content right now, or is that properly a separate, differently-shaped job? We
wrote both positions into the new bug file (229) and flagged it as a decision
for you — we did not expand the change on our own authority.

Still to do: one full rehearsal of the whole thing on a test site (patch,
rebuild, watch the alarm ring and the copy appear), and watching the big
rebuild wave from the other fix when it fires. The checklist is in the bug
file and the handoff note.

## 2026-08-09, later — the reviewers approved it

Third time through, the council approved the change — no serious objections
left. The two notes they attached are both "keep an eye on this" rather than
"change this": one is a reminder that our "nothing was lost" reading of an
empty ledger depends on the trigger's own filter staying exactly as written,
and the other is the same page-content question as before, which is yours to
decide (bug 229 lays out both sides).

So: the safety net is live, approved, and has already saved real content
twice without being asked. What remains is the rehearsal and watching the
first big wave go through — both written up step by step for whoever sits
down next.

## 2026-08-09, mid-morning — the rehearsal happened, and it all rang true

We ran the full fire-drill on the darts demo site. First, a stroke of luck:
the big rebuild wave from the other fix had just visited that exact site
twenty minutes earlier, which did half the setup for us (its chrome was
freshly fingerprinted). We then hand-edited the footer the "wrong" way — a
direct database poke, the very move that caused the original loss — and asked
the platform to rebuild.

Everything we promised happened, in order: the alarm rang (once, in the right
place), the hand-edited version was copied to the archive before being
replaced, a review ticket appeared naming exactly which edit was lost, and
the fresh rebuild took its place properly fingerprinted. Then the control
run: rebuild again with nothing touched — silence, no copies, no tickets,
which is just as important (a smoke alarm that goes off every time you cook
is one you unplug). Even the hand-edit itself left a tidy archive copy of
what it replaced, which proves the net catches the sneaky direct-edit route
too, not just the official ones. We cancelled our own test ticket afterwards
with a note, so nobody wastes time reviewing a deliberate drill.

The wave itself has now begun, by the way — three sites through so far, no
errors, the fingerprint count creeping up as designed. That part just takes
however long it takes; the checking queries are written down for whoever
looks next. In substance, this one is done: built, reviewed, approved,
rehearsed, and already earning its keep.

## 2026-08-09, early afternoon — you asked if the wave had finished; it hasn't, and I had it wrong

Short answer: no, and my summary this morning gave you a rosier picture than
the facts support. I said the fingerprinting was rolling out at about a site an
hour and would simply fill in. Two and a half hours later it has gone from 3
slots to 6, out of 57 — and one of those two sites got done by accident, by an
unrelated job that happened to rebuild its header and footer on other business.

What I got wrong is worth saying plainly, because it is a thinking error rather
than a typo. I kept counting how many things had been *found* and treated that
as evidence that things were being *fixed*. Those two look identical if you only
count — a pile that is growing looks the same whether it is being worked through
or just growing. The moment I looked at how long each item had waited, it was
obvious: the three that got done were picked up within five minutes each, and
the newest one has been sitting untouched for over three hours.

The reason is not a fault in our safety net, which is behaving perfectly — no
errors at all since it went live. It is that the thing filing these jobs is a
brand-new hourly checker another thread switched on this morning, and it is
deliberately "look, don't touch": it writes down what it finds and stops there.
The part that would pick those jobs up and act on them is a known open problem
of its own (bug 83), untouched on purpose.

So where does that leave us. The protection itself is finished and proven — a
hand-edit is archived at the moment it is destroyed on every site, fingerprinted
or not, so nothing can be silently lost anywhere today. What the fingerprint
adds is the *alarm*: on a site that has one, we also get told. That alarm is
live on two sites and will reach the others gradually as ordinary rebuild
traffic touches them, with no schedule and no timetable I can honestly give you.
I have rewritten the checklist item accordingly, because as I had written it, it
was waiting for an event that was never going to arrive. If you want the
remaining 17 sites fingerprinted sooner, that is a one-line job to dispatch and
I can do it on your word — but it is a rebuild of every site's chrome, so it is
your call rather than mine.
