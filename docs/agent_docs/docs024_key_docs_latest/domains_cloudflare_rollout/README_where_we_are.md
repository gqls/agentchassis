# Where we are — putting the whole domain portfolio behind Cloudflare

## 2026-08-03

We want almost every domain we own — the ones on our Nominet tag, plus the ones
at Dynadot, Porkbun and Spaceship — set up on Cloudflare's free plan, with each
domain pointing at our portfolio router worker, the same way dartsonline.com
already works.

Where we've got to: Cloudflare is done and proven — we made a limited-permission
API key, checked it works, and read the existing setup with it (36 domains are
already on Cloudflare; they all follow the same simple pattern, which we'll copy).
The machine can reach Nominet's EPP service and all three registrars' APIs, so
nothing is blocked on connectivity.

What's needed to actually run it, all on the account-owner side:
1. Nominet: the TAG name and EPP password, and add IP 151.226.83.138 to the EPP
   allowlist in Online Services.
2. Dynadot: an API key (Tools → API in their control panel).
3. Porkbun: an API key + secret (porkbun.com/account/api), ideally restricted to
   that same IP.
4. Spaceship: an API key + secret (API Manager), read+write on domains only.
5. A decision on which domains to skip ("almost all" needs a list or a rule).
6. A decision on www: right now the pattern doesn't make www.domain work at all —
   one extra DNS record per domain would fix that. Do we want it?

Once those land the run is mechanical: list every domain, create the missing
Cloudflare zones, add the record and worker routes, then switch each domain's
nameservers over at whichever registrar holds it, and finally check every single
one came up active rather than trusting the batch said "ok".

## 2026-08-04

You gave us the skip list and the rule behind it (everything static until a
domain gets a real dynamic service, then we deliberately move it), and said www
should work by pointing at the main site. All recorded; the zone template now
includes www.

Two problems surfaced, both about addresses, and they change two earlier asks:

1. Your office internet connection changes its address — it has already changed
   since Saturday, both kinds (IPv4 and IPv6). So locking the Cloudflare key to
   an address was my bad advice: that lock has now shut us out entirely, even
   though the key itself is fine. Please edit the token in the Cloudflare
   dashboard and REMOVE the IP filter list (keep everything else) — the tight
   permissions and expiry are what actually protect us.
2. Same for Nominet: the address you kindly added has already gone stale. The
   good news is the server cluster has five fixed addresses that never rotate.
   Please add these to the Nominet EPP allowlist instead:
   134.213.168.26, 134.213.168.37, 134.213.168.44, 134.213.168.54,
   134.213.168.56 — then EPP will run from the cluster and never break this way.

Also still needed: the Nominet TAG name (the login username that goes with the
password you provided), and the three registrar keys when you get a moment. A
subtle trap got written down for future sessions: both services happily pass
their "is it working?" checks even when the address lock is the thing that's
broken — we now know to test with a real call instead.

## 2026-09-02 — Spaceship is connected

You created the Spaceship API key and put it in the credentials file; I checked it
works and pulled the full list. 203 domains sit at Spaceship, essentially all
parked at the two marketplaces (144 pointing at aftermarket.com nameservers, 58 at
atom.com, plus one already on Cloudflare). Nothing has been changed anywhere —
reading only so far.

There is now a small command-line tool (scripts/domains/spaceship.py, alongside
the Porkbun and Dynadot ones another session built today) that lists domains and,
when we're ready, repoints nameservers to Cloudflare. The repointing side hasn't
been tried yet — the first real repoint will be its test.

On renewals: 17 of the 203 expire before the New Year, six of them within about
three weeks, but all 17 have auto-renew on, so nothing lapses by itself. The 14
domains with auto-renew off all run until at least June 2027.

Of the three registrar keys you owed the rollout, Spaceship is now in. Dynadot and
Porkbun are still outstanding — same shape when you get a moment: create the key,
drop it in the credentials file named in the runbook, in a separate terminal.

Update, 4 September: all three registrar keys are now in — the paragraph above is
out of date. Spaceship and Dynadot went in on the 2nd, and you created the Porkbun
key the same day. Porkbun is connected and working: we listed all 683 domains
there, none of them are on Porkbun's own sale marketplace, and the valuation work
got three data files out of it (your domain list, UK asking-price comparables,
and .com comparables for your category keywords).

One switch left, and it's in your hands: in your Porkbun account settings there
is an "enable API access for all domains" toggle, still off as of today. Nothing
about listing needs it — but until it's on, we cannot change any DNS or move any
Porkbun domain onto Cloudflare. One click covers all 683; there's no need to do
it per domain. When you flip it, tell any session and it will confirm the door
is open within seconds.
