# Where we are — Nominet domain management

## 2026-09-02 — the lane opens, and immediately finds four domains going dark

You asked for a single home for everything Nominet, so this is it: the tag, the
EPP connection, the domain list, registrations, transfers, and moving domains'
nameservers.

First thing found: the nameserver change you made at Nominet for the four remake
domains — advertise.co.uk, designblog.co.uk, seotools.co.uk and
websitepromotion.co.uk — points them at Cloudflare, but no Cloudflare zones
exist for them yet. Until a zone exists, Cloudflare refuses to answer for those
names, so as the internet's caches empty over the next day or two each domain
stops resolving entirely. advertise.co.uk is already unreachable from here. The
old Drupal site appearing to still work earlier today was just stale cache.

The fix is ready and safe to re-run: one command creates the four zones and
wires them the same way as every other portfolio domain. My session isn't
allowed to run it (it involves the Cloudflare key), so please run, in this chat:

    ! scripts/domains/cf-zone-bootstrap.sh advertise.co.uk designblog.co.uk seotools.co.uk websitepromotion.co.uk

That makes the four domains answer again (serving the portfolio router; the
remade sites appear on them as each build lands).

Two things only you can move, when you have a moment:

1. **The domain list.** Your ~1,500 .uk domains at Nominet have never been
   listed out — every other registrar got counted today, Nominet is the one
   still missing. Easiest: export the CSV from Nominet Online Services and drop
   it somewhere I can read. Alternatively run the prepared walk (runbook §2).
2. **The second tag.** The application went in on 11 August and we've heard
   nothing since — has anything arrived from Nominet?

Small tidy-up, no rush: the old read-only Cloudflare key (the file called
`token`) has stopped working entirely. Everything now uses the newer
`portfoliotoken`, so nothing is broken — but if you meant to keep two working
keys, that one needs remaking.

## 2026-09-02 (later) — Nominet now has one tool, like the other registrars

You asked for the same shape as the Dynadot and Porkbun lanes, so there is now
a single command for everything Nominet: `scripts/domains/nominet.py`. It can
test the connection, list your domains, check availability, show a domain's
details, and move nameservers (dry-run first, always). Registering new domains
stays in the separate careful tool that spends money.

Its offline checks all pass and it reaches Nominet from the cluster fine, but
no logged-in command has run yet — my session isn't allowed to touch the
credentials. Its first real outing doubles as the domain list you're owed:

    ! python3 scripts/domains/nominet.py login
    ! python3 scripts/domains/nominet.py walk --months 120 > all_domains.txt

The first line proves the connection; the second lists every .uk domain on the
tag into a file (a couple of minutes). One improvement over the old plan: the
old recipe only looked twelve months ahead for renewal dates, but domains can
be registered for up to ten years, so it could have quietly missed some — this
walks the full ten.

## 2026-09-02 (late evening) — the four domains are back, and advertise.co.uk is publicly launched

You ran everything, and it all worked. The wrinkle along the way: Cloudflare
gave the four new zones a different pair of nameservers (betty/ivan) from the
one the older sites use (alexis/leah), so the nameserver change you'd made at
Nominet pointed at the wrong pair and the zones could never come alive. The
new Nominet tool moved all four to the right pair — its first real write, and
it behaved exactly as designed — and on the next check all four zones went
active immediately.

All four domains now serve real pages over https, checked properly (actual
page content fetched from Cloudflare's own servers, not just a status code).
That means the advertise.co.uk remake — built earlier today — is now live on
its own domain. A detail passed to the remake programme: the other three
domains are serving full-looking sites even though their remake briefs are
still waiting on you — worth a glance at what they're showing in the meantime.

Still open from earlier: the domain list (the walk hit a one-off connection
blip — just run it again when convenient) and the second-tag question.

## 2026-09-03 — the domain list is done: 1,606 domains, and three real bugs found along the way

You ran the walk. It didn't work first go, and that turned out to be useful:
nobody had ever actually run this successfully before, only proven the
connection worked. Fixing it took three rounds — a message shape Nominet's
system rejects, a permission that has to be requested at login rather than
when you use it, and — the one that mattered most — my code was reading
domain names from the wrong place in Nominet's replies. That last one is the
dangerous kind of bug: it would have returned "zero domains" for every month
with no error at all, so a full run could have quietly told you your entire
domain list was empty and looked like a clean success. I found it by
refusing to believe "zero domains this month" without checking a few more
months first, and I've now built in a permanent check so that specific
failure can never happen silently again.

With all three fixed, the real walk ran clean: **1,606 domains**, a bit more
than your ~1,500 estimate from a couple of weeks ago — normal growth, not a
mistake. Two other sessions working on your domains — one valuing the whole
portfolio, one building a Sedo listing sheet — were waiting on exactly this
list, so I delivered it to both as soon as it was ready. The valuation
session cross-checked it against a completely separate export from your own
Afternic dashboard and it lined up almost perfectly, which is good
independent confirmation the number is real.

That same cross-check also settled your question from this morning about
which domains might have expired — nine names showed up nowhere in any of
our lists. Five are genuinely gone and would need re-buying, not just
finding a forgotten password; two are undetermined because that particular
domain ending doesn't publish public ownership records; the other two turned
out already accounted for once I looked closer.

Later the same evening you asked me to release 50 domains to your tag. I
checked first rather than guessing, and the registry already showed all 50
as yours — so there was nothing for that specific action to do. It turned
out you meant something different (moving legal ownership, not the
registrar record), and you handled that yourself directly. Worth remembering
for next time: "release" and "transfer" are two different things at Nominet,
easy to say interchangeably but not the same underlying action — I'll ask
which one is meant if it's ever ambiguous again, rather than assume.

Still open: the second-tag question (nothing heard since 11 August), and
tidying the domain list against what's parked where (low priority).
