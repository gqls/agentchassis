# SUMMARY — idea.uk VM site, 4 August 2026

**What we're trying to do.** idea.uk is a live, card-taking report tool running
on a single rented server. This stream's job was to fix how the internet
reaches that server: put a professional front door in front of it, close a
serious hole in how it answered to domain names, and turn domain plumbing for
the wider portfolio into something a machine can do on instruction rather than
a person clicking through registrar websites.

**Where we've come from.** At the end of July an adjacent workstream, while
answering an unrelated DNS question, discovered that our own runbook believed
idea.uk was behind Cloudflare — and it never had been. There was no web
application firewall, no bot protection, no DDoS absorption; the server's
address was public. Worse, they found that any domain name at all pointed at
our server was served the complete shop, working checkout included — two
unrelated domains briefly did exactly that, and a real, payable order could
have been placed from the wrong one. Their handoff corrected the facts and
left two decisions to this lane's owner: whether to move behind Cloudflare,
and whether to close the hostname hole.

**What we've done.** First, re-measured everything in the handoff rather than
inheriting it — it all held. The owner decided both questions: Cloudflare, and
close the hole now. The hole was closed with a catch-all configuration that
shuts the door on any name that isn't idea.uk, verified against a
sixteen-endpoint baseline, with certificate renewal proven unaffected and the
provisioning script fixed so a rebuilt server keeps the protection. The
visitor-address restoration config was installed early, while it provably did
nothing, which eliminated the one window in the cutover where the rate limiter
could silently break. We built a small tool that speaks Nominet's registrar
protocol directly, so our tag can change .uk nameservers by machine; it has
now made two clean production changes. With it, idea.uk's nameservers moved to
Cloudflare — the registry published in about two minutes and the site never
blinked. Then the proxy was switched on with the strictest encryption mode,
real visitor addresses were proven from two different networks, the server was
firewalled so nothing but Cloudflare can reach its web ports, and all sixteen
endpoints re-tested identical. Along the way loanzy.uk was moved off its
"for sale" parking page onto Cloudflare with its certificate live, and a
request to do the same for webzy.uk was caught in time — we don't own it; the
mistakenly created zone has been deleted. Every misstep is written down where
it happened: a false-alarm certificate watcher, three separate IPv6 traps, a
token lockout my own recovery watcher was unknowingly sustaining, and one
important inversion — with the firewall up, switching DNS back to "grey" is no
longer a harmless rollback, and the runbook now says so.

**Where we are now.** idea.uk sits behind Cloudflare: strict encryption from
edge to origin, per-visitor rate limiting proven to see real addresses, and an
origin that times out for anyone who isn't Cloudflare while the site serves
perfectly through the front door. Everything in that sentence was verified
live, none of it inferred. One confirmation remains outstanding by design: the
first genuine Stripe payment through the new front door — the plumbing is
proven end to end, and the first real order settles the last inch. loanzy.uk
is delegated and certified, serving a correct placeholder error until the
webdesign stream wires its content.

**Where we're going.** This lane's active work is done. What follows: watch
the next real order confirm the payment path; loanzy.uk's content belongs to
the webdesign stream; and Cloudflare's protective features — WAF rules,
Turnstile, bot fighting — are now available but deliberately not yet
configured, which is a decision worth taking on its own rather than a loose
end. The registrar tool stands ready for any future .uk nameserver move on
our tag, with one standing caution: check whose tag the domain is on first.
