# SUMMARY — site delivery and the customer editor — 16 August 2026

*First summary for this workstream. Written to be read aloud.*

## What we're trying to do

We sell a finished website for £149. The offer says the customer gets the site,
that the files are theirs, and that no changes are included. Today we can build
the site and show it to them on a private preview address — but three things
the offer implies are missing. They cannot get a hosted copy of their own; they
cannot take away the files as a package; and they have no way to change anything
themselves afterwards. This workstream builds those three, in that order, and it
has to do it without breaking a rule we discovered the hard way elsewhere: if
two different things can rewrite the same page, one of them will silently
destroy the other's work.

## Where we've come from

The starting point was an idea and a reference document: publish each finished
site to a free hosting account, and point some kind of editor at it. Reviewing
that document critically changed the shape of the plan rather than the goal. The
proposal wanted a new central engine with its own customer database — but our
platform already is that engine, and building a second one would have duplicated
the very thing that keeps our sites consistent. The proposal also assumed we
would hand the hosting account over to the customer, and the provider we were
looking at can only do that through a manual click in a dashboard, which fails
the owner's clear bar: completely automated. That turned out not to matter,
because of a decision the owner made at the same time — we keep the ability to
update delivered sites. A site we can still write to cannot meaningfully be
handed away, so what the customer owns is the ZIP, exactly as the £149 terms
already said. With that settled, the hosting choice became a straightforward
engineering question, the plan was approved on 14 August in six phases, and the
first — changing the promised build time to "usually ready the next day" — went
live the same evening.

## What we've done

Phase 2, the publishing machinery, is built, reviewed, live, and proven in
production. Every site now has an opt-in switch that is off by default, so
nothing at all changed for any existing site until we deliberately turned one
on. A timer picks the least-recently-checked site that has opted in and asks
whether anything about its built pages has changed since we last published. If
not, it stops. If so, it copies the pages to the site's hosted address, fetches
one back down from the public web, and compares it byte for byte against the
original. Only then does it record the publish as complete. That ordering is
the whole design: we never record success on the strength of an upload
appearing to succeed.

Two things happened on the way that are worth telling, because both are the
process working rather than failing. The reviewers rejected the first
submission, and they were right — the way I was taking a safety backup of a
shared configuration row would have quietly dropped part of it, so a restore
would not have restored. That was fixed and approved on the second pass. Then
the first live run failed on its very first file: the storage service refuses
an upload unless told the size in advance, and we were streaming without one.
Nothing was half-copied and nothing was falsely marked as published, which is
what the design existed to guarantee. We fixed it, then put the mistake back
deliberately to confirm the new test actually catches it.

## Where we are now

Yesterday the canary site failed on file one. Today it copied all eight of its
files, and the page served from its public hosted address is byte-for-byte
identical to the original — checked against a copy taken before any of this
existed, so the comparison could genuinely have come out wrong. An hour later
the timer came round again and correctly did nothing, because nothing had
changed, and it said so in as many words. One site is switched on; every other
site in the estate is untouched and off. We also slowed the timer from every ten
minutes to every hour, because each check was starting a fresh worker just to
answer "nothing changed" — around a hundred and fifty a day for one site — and a
finished site changes a few times a day at most.

The one deliberate gap: the hosting provider the plan recommends as the
eventual primary is registered but switched off, and it refuses loudly if
anyone tries to use it. It needs a key that does not exist yet, and its upload
protocol is only partly documented — writing that code blind, with nothing to
test it against, is exactly how you end up with something that looks reviewed
and has never run. The route we are using instead needs no new credentials and
is the one now proven.

## Where we're going

Next is the ZIP the customer actually owns — the same machinery, aimed at
packaging rather than publishing, and with one specific hazard already written
down: a partly-written archive is a silent broken promise, so it must never be
allowed to truncate quietly. After that, the handover step that marks a site as
delivered, then the emailed link that lets a customer in, then the editor
itself, where the sharpest risk is making it structurally impossible for one
customer to reach another's site.

Two keys remain with the owner and neither blocks us: the Stripe pair for taking
payment, and a hosting key if we later want to move to that provider. Everything
between here and a customer editing their own delivered site is our work.
