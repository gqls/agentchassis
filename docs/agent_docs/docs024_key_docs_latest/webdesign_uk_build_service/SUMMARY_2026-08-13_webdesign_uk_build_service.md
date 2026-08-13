# SUMMARY 2026-08-13 — the chat bot now tells the truth from the database, and the "sell a build" loop is scoped

Milestone read-out for the owner, written to be read aloud. Current-state only;
the chronology is in NOTES and README. Supersedes nothing — a new file, per the
standing-docs rule.

## What we're trying to do

Sell a done-for-you website through webdesign.uk: a visitor turns up, has a
short conversation with a chat assistant about their business, and — the part
that doesn't exist yet — pays and gets a real site built for them
automatically. The near-term job has been to make the shopfront and the chat
assistant honest, safe, and cheap to run; the next job is to close the loop
from "conversation" to "paid site."

## Where we've come from

The five marketing pages were built through the framework and parked "ok for
now." A small hand-written chat service went live on its own VM — it holds a
real conversation, has all its safety controls (rate limits, spend ceiling,
falls back to real contact details if the AI is unavailable), and its
transcripts are the demand signal the whole phase exists to collect. But it had
a structural weakness: the facts it stated — price, terms, refund policy — were
copied into its code by hand, frozen at the moment it was written, with no link
back to where those facts actually live. That weakness bit twice: once when a
£75 deposit went live in the database while the bot still promised a full
refund, and again this week when the whole site moved to a new £149 price while
the bot carried on quoting the old £1,200.

## What we've done

Built and then switched on a small service inside the cluster whose only job is
to hand the chat bot one site's facts, read live from the database, over a
private encrypted tunnel — without exposing the database itself, which stays as
locked-down as it always was. As of today the chat bot reads its prices and
terms from the database in real time. The moment it went live it corrected the
£1,200-versus-£149 contradiction a visitor would otherwise have hit: the bot now
says £149, one-off, pay-after-approval, no refund, with no trace of the retired
offer. Change a fact in one place now and the bot follows within minutes, with
no rebuild — and if the service ever can't reach the database, the bot refuses
to start rather than quietly revive its old baked-in figures.

Along the way the tunnel work turned up a real latent fault — it looked
connected but silently forwarded nothing, because a single system setting was
off, and it had never been caught because no one had ever pushed real traffic
through it end to end. Found, fixed, proven with live traffic. The cluster-side
service went through the platform's own review panel and was approved, after one
revision round that correctly caught a weak test.

## Where we are now

The shopfront sells the £149 offer consistently, on every page and in the chat.
The chat assistant is live, correct, and now speaks from the single source of
truth. What it still does **not** do is take money or build anything — it has
the conversation and stops, because it was only ever built to collect the
conversation. The pieces needed to go further all exist but as separate,
disconnected parts: a payment system that is built but has no keys yet; a build
pipeline that works but is triggered by hand; automatic hosting that is proven;
and dispatch reliability that was fixed this week. None of them are joined up.

## Where we're going

Closing the loop from conversation to paid, built site. The recommended shape,
recorded for the owner and for the sibling workstream that owns most of it: take
the money first — it fits the £149 no-refund terms cleanly and makes every
automated build a paid, accountable one rather than something a stranger can
trigger for free. The concrete chain is four connections: switch on payments
(needs Stripe keys — the one thing genuinely blocked on the owner), route the
payment confirmation into the cluster (the new tunnel can carry it), fire the
build when payment clears, and turn the chat transcript into a build brief.
Hosting the result is essentially free once a site is built. The immediate next
step is the owner's — the Stripe keys — after which this becomes the sibling
lane's build work, picked up cleanly from the 2026-08-13 handoff.
