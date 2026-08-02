# Where we are — the lost seed scope (bug 174)

Plain prose, append-only, newest at the bottom.

---

## 2026-08-02, morning — what this bug actually was

When you fire a diagnosis at the platform you can tell it *which bits of code to
look at*. That's the `SEED_SCOPE` setting. It turns out that for at least the last
week, the platform has been quietly ignoring you.

Not failing. Ignoring. You'd name three functions, and the diagnosis would go and
read a completely different set of functions that a search engine picked out for
itself, then come back with a confident answer about those. Nothing errored,
nothing warned, and the report looked entirely normal. The only way to find out
was to already know what you'd asked for and go digging in the database.

We know of three real diagnoses this happened to. Two of them were other people's
actual investigations — one into a broken image deploy, one into the scheduler.
Both were aimed at chosen code, both silently re-aimed somewhere else, and
neither author had any way of knowing. Whatever those runs concluded about "the
code shows X" was about code nobody chose.

## The bit that was more interesting than the ticket said

The ticket that reported this had already worked out the cause: there's a relay
step between you and the diagnosis engine, and it wasn't passing the setting
along. It listed the fix as "add the setting to the relay's list of things to
pass on".

That would not have worked. There were **three** places the setting got dropped,
not one, and they had to be fixed together:

1. The relay reads the job ticket and copies out the fields it cares about. It
   never copied this one — so there was nothing for step 2 to pass along even in
   principle. The ticket didn't mention this one.
2. The list the ticket *did* name.
3. And then a type problem. The setting is a list of names. On the way through
   the database layer it gets flattened into a piece of text that *looks* like a
   list but isn't one, and the code at the far end quietly treated "text I can't
   read as a list" as "nobody gave me anything".

So if I'd done what the ticket said, the setting would have been dropped anyway —
in a new place, just as silently. That's the thing worth remembering about this
one: the proposed fix for a silent-failure bug would itself have failed silently.

## Why nobody noticed for a week

The diagnosis engine has a sensible fallback: if you don't tell it what to look
at, it works it out from a code search. That's correct behaviour and we want to
keep it.

But it means a *lost* setting and an *absent* setting look identical from the
inside. The engine can't tell "the operator didn't ask for anything" from "the
operator asked and it got confiscated in transit", so it does the reasonable
thing and falls back — producing a perfectly good run that answers a different
question.

A fallback that's this well-behaved is exactly what turns a lost setting into an
invisible one. So beyond fixing the three gates, I made the engine record *which
route it took*, and print a warning into the report when it used the fallback:
"this scope was not chosen — if you did pass one, it didn't arrive." It can't
tell you which of the two happened, and it doesn't pretend to. It puts the
question in front of the one person who knows the answer.

## The check I nearly got wrong

The ticket also asked for an automated check so this class of thing can't drift
again. I built the obvious version — "every step must pass on everything the next
step says it accepts" — and then measured it before trusting it.

It produced 31 complaints, and the ones I looked at were all legitimate
behaviour, not bugs. I tightened it. It produced 3. Better.

Then I checked *which* 3, and none of them was this bug. The check was blind to
the exact thing it was written for, because of a detail in how this particular
relay decides who to call. I'd have shipped a green light that could never turn
red.

So the check is narrower and honest now: it only makes the assertion where the
question is genuinely answerable, it names which relays it's asserting about, and
it separately reports relays it *isn't* covering rather than letting silence read
as approval. There are three such relays on the whole platform; one is covered,
two are listed as not covered. And when a check we claim to be running stops
matching reality, that's reported louder than an actual fault — because a check
that quietly stopped running is worse than one that found something.

That last part isn't theoretical. The very first time I ran it against the live
system it matched nothing at all, because of a naming mismatch I'd introduced.
It told me so. If I hadn't added that category ten minutes earlier it would have
printed "all clear".

## Where it stands

The configuration half is applied and live now. The code half is committed and
goes out with the next chassis build. The council reviewed it and approved it
first time round, with six advisory comments — four of which were fair and I've
acted on all four, including one that caught me reusing nothing where a perfectly
good helper already existed, and one that caught me quoting a blast-radius number
I'd never properly measured. The real number turned out to be 1, not 14, which
makes the risk smaller than I'd told them.

Two other sessions were working in the same files within the same hour, so a
chunk of the care here went into not standing on their work — including
renumbering my own database change when someone else claimed the same number
eight minutes after me.
