# noted.co.uk — second read-out: the machine is chosen

2026-08-10, later the same day. Written to be read aloud.

The first read-out today (`SUMMARY_2026-08-10_noted_rebuild.md`) ended with three
open questions. **One of them is now closed**, and closing it settled more than it
looked like it would. This is the read-out to that point.

---

## What we're trying to do

Unchanged. noted.co.uk is a note-taking app — type a note, record your voice, add
a photo. Everything currently lives in the browser you wrote it in, so clearing
your browsing data destroys it and you can't pick up on your phone what you
started on your laptop. We're rebuilding it on the framework, moving the notes
onto a server we control so people can sign in and find their notes anywhere, and
making the framework genuinely own every part of it — **decomposed** into the
framework's own parts, not swallowed whole as a few opaque lumps.

## Where we've come from

This morning the app was live but invisible to us: sitting inside the framework's
own storage bucket since January, with no database record, in no repository, and
not mentioned anywhere in the code.

We put a notice on it telling people to save their work — and found, just before
shipping it, that the Backup button saved only the text. No voice recordings, no
photos. So the notice went up together with a fix, because otherwise it would
have sent people to a button that quietly loses the one thing they can't
recreate. That's all live and verified now, and the app is in version control for
the first time.

## What we've done since

**The owner made the hosting call, and it turned out to be an easy one.** The
ruling was: prefer Mythic Beasts over Hetzner where it's an easy choice, and keep
the apis.uk machine for the API and nothing else.

We run four machines. Two are Hetzner — the idea.uk payments box and the
relojistas traffic site — and the ruling deprioritises both. Of the two Mythic
Beasts machines, the apis.uk one is now explicitly off the table. That leaves
exactly one candidate: **the webdesign.uk box in Cambridge.** No judgement left in
it, and it happened to be what we'd recommended anyway.

**We measured the box rather than trusting the paperwork.** Its own planning
documents specify 4 GB of memory. The machine actually has **7.8 GB, of which 6.2
is free**, two cores, and **47 GB of its 50 GB disk unused**. It's close to idle.
There's plenty of room. Postgres isn't installed yet, which is the main piece of
work to add.

**We also found something on that box worth passing to the team that owns it.**
The web design chat service is listening on every network interface rather than
just the machine's own loopback. Nothing is exposed — the firewall is on, denies
everything inbound except SSH, and we confirmed from outside that the port really
is unreachable. But that box's own documentation says services bind to loopback
*so that a firewall mistake exposes nothing*, and this one doesn't. So for that
service the firewall isn't a second line of defence, it's the only one, and a
single mistaken firewall rule would put a service that spends real money on the
public internet. Not ours to fix, but very much worth someone knowing.

## Where we are now

The live app is safe, warns people, can genuinely save everything, and is under
version control. The framework knows the domain and has its foundational specs.
**And we now know exactly which machine this is being built on**, which unblocks
the next phase.

The one real objection to that machine is that it already serves the web design
shopfront — a live commercial front door — and putting a consumer app holding
user data beside it couples the two. Our own history argues against that: we've
twice bought a separate machine specifically to avoid this kind of coupling, once
for a saving of about three euros a month.

The owner's ruling points the other way and the ruling stands, so we've answered
the coupling instead of ignoring it. **The recordings and photos will not live on
that machine's disk** — they go to the object storage we already run, and the
database on the box holds only text and bookkeeping. That's better design
regardless, and it removes the one thing that could actually take the shopfront
down: noted's storage grows without limit and the shopfront's doesn't, so a
shared 50 GB disk was the real risk. With the media held elsewhere, the box is
running a small text database that will stay small for years.

Two things still need deciding, and both are yours rather than technical.

**The privacy promise has to change.** Today the site says "we can't see your
notes, read your text, or listen to your recordings", which is true only because
there is no server at all. The moment you can sign in from another browser, it
stops being true. We've blocked the old wording at the framework level so it
can't be copied forward by an agent that mistakes it for house style — but what
replaces it is a decision about the product. The honest, cheap position is "your
notes are on our server so you can reach them anywhere; we could technically read
them; we don't". Anything stronger means actually building end-to-end
encryption, which is real work and is currently blocked as a claim precisely
because we haven't done it.

**And migration needs a person to act.** We hold no copy of anyone's notes and
can't reach into their browsers. The only path across is someone exporting from
the old app and importing into the new one. Today's backup file was designed to
be exactly what the new server will accept, so the path exists — but it's the
reason the notice on the live site matters.

One boundary hasn't moved and shouldn't be forgotten: **the framework doesn't
write server code and has nothing for end-user accounts.** Every backend we run
is hand-written Go, and the one component that does have accounts is for platform
operators and isn't reachable from a public site. So the notes server will be
written by hand, like the other three. Everything else — the pages, the editor,
its behaviour, the checks over it — can be framework-owned.

## Where we're going

Add Postgres to the Cambridge box and set up its backups. Write the notes server:
accounts, sessions, notes, media to object storage, and an import that accepts
exactly the backup file we shipped this morning. Bind it to loopback, behind
nginx, on its own systemd unit and its own database role — the lesson from what
we found on that box today.

Then, before rebuilding the app itself, write its behaviours down as contracts
the framework can test — signing in and finding your notes, capturing a thought
offline, attaching a recording, getting your data back out. That mechanism
already exists and is what makes "the framework checks it" a real claim rather
than a hopeful one.

Then build the site through the pipeline, properly decomposed, and only then move
the domain across — keeping the old app reachable for a while, because what's in
people's browsers is still the only copy that exists.
