# SUMMARY 2026-08-09b — webdesign.uk build service

Fifth in the series (08-04, 08-04b, 08-06, 08-08, 08-09, this one). Written the
same day as the last entry, because the read-out genuinely moved: Phase 4 went
from a todo list item blocked on a key to built, deployed, and proven live.
Cold-start doc: `HANDOFF_2026-08-09b_continue_here.md`.

## What we're trying to do

webdesign.uk sells complete websites to small and medium UK businesses at one
fixed price, built by the machinery it sells, hosted on our own box. The site
is the first half of the product; a chat box that opens a real conversation
with a visitor, rather than a form, is the second — and it's the one piece we
write by hand, because nothing in the framework generates backend services.

## Where we've come from

By this morning the five-page site was built, reviewed twice, and parked by
the owner as "ok for now." The one thing standing between the plan and the
next phase was a practical one: the chat service needed its own Anthropic API
key, kept deliberately separate from the platform's own credential, and the
owner had to create that himself in a console neither of us had used together
before. That took a few wrong turns — the first attempt landed on the wrong
website entirely, since Anthropic's chat product and its developer platform
are two different sites that happen to share a login — before we found the
right page and the owner created a properly isolated workspace and key.

## What we've done

With the key in hand, we built the chat service in one sitting, following the
plan's own design brief closely: a small, hand-written Go program, no
generated backend, no external dependencies beyond the standard library,
running on its own port on the box behind the same tunnel that already serves
the site.

The five safety controls the plan called for from day one are all in the
service, not bolted on after: a limit on how often one visitor can start a
conversation, a hard cap on how many turns one conversation can run, a daily
spending ceiling that refuses new calls once it's reached, a log of every call
with its real cost, and every message kept as a proper record — which is the
whole point of the service, since what people actually ask it is the evidence
for what the framework should build next.

Two things were worth catching on the way, because both were the same shape of
mistake in two different places. This box reaches the internet through a
private tunnel rather than sitting directly on the web, which means the
usual way of finding a visitor's real address doesn't work here — it would
have quietly treated every visitor as coming from the same place, defeating
the whole point of a per-visitor limit. We used the address Cloudflare itself
stamps on the request instead, which can't be faked by whoever's asking, and
then found the same trap waiting a second time in the web server's own
configuration and avoided it there too. And separately: an early line of the
service's own bookkeeping code used the wrong template for today's date — a
typo that would have silently broken the daily spending count from the very
first day. Caught it on a second read, before any test ran, well before it
could have mattered.

Before calling either control finished, we didn't just watch the tests pass —
we broke each one on purpose, confirmed the test caught the break, then put it
back. That's the difference between a check that exists and a check that
works: a test that would pass whether or not the thing it's guarding is
actually there proves nothing.

Then we proved the whole thing for real rather than trusting any of it in
isolation. Built the program, copied it to the box — which has no programming
tools installed, by design, so it has to arrive as a finished binary — wired
it into the web server, and started it. It refused to start three times in a
row until the contact details it needs for its own failure path were in
place, which is exactly the caution we wanted from it. Once those were added,
a real message sent through the actual public address came back with a real,
sensible reply from the AI, and the record it kept of that exchange showed the
exact right cost for what was asked and answered. We even checked the visitor
address it had logged against an outside service that told us our own address
independently — they matched exactly, which is about as good a proof as one
network can give that the address isn't a stand-in or a mistake.

## Where we are now

The site and the chat service are both live on the private preview, though the
chat service isn't reachable from the page yet — there's no button pointing at
it. Nothing public has changed; the domain still forwards elsewhere and the
preview remains the only way in.

One loose end, not urgent: proving that a second visitor from a genuinely
different network gets logged with a genuinely different address, which needs
two connections to test properly and only one was available today. The check
that matters most — that the address is real and not a stand-in — is already
done; this would only add the second half of that proof.

## Where we're going

The input box on the page itself, wired to talk to the service that now
exists. That's what the nine buttons on the site with nowhere to point have
been waiting for. Once it's there, the owner reviews the whole thing — site and
chat together — and on his approval, the switch that makes it all public.
