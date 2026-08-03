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
