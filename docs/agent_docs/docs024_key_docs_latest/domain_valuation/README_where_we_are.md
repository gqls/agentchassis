# Where we are — domain valuation (append-only, newest at the bottom)

## 2026-09-02 — lane opened: gathering the lists

You asked for every domain across Dynadot, Porkbun, Nominet and Spaceship to be
listed and valued — .co.uk, .uk and now the .coms too — with a view to selling
roughly the bottom 500 at keen prices, keeping whole categories together rather
than splitting up, say, the financial names.

Done today: asked each of the four registrar sessions for their full domain
list plus any valuations their registrar can produce; asked the Afternic
session to bring over its current asking prices as a comparison (you've said
they're generally too high, so they're an input, not an answer); and set a
search running over the earlier conversations where we discussed .co.uk/.uk
values, so we start from what was already agreed rather than from scratch.

The one list nobody holds yet is Nominet's — the ~1,500 .uk domains. The
walk failed on a connection blip earlier; when you have a minute, run in the
Nominet session (or here):

    ! python3 scripts/domains/nominet.py login
    ! python3 scripts/domains/nominet.py walk --months 120 > all_domains.txt

## 2026-09-02, late evening — three lists in, categorisation running, method drafted

The registrar sessions moved fast. Dynadot (451 domains), Porkbun (683) and
Spaceship (203) have all delivered their full lists — 1,337 domains in hand,
almost all .com, and about 85% of them parked at Afternic. None of those three
registrars offers valuations through their API except Dynadot, whose appraisal
tool ("Dynappraisal") is being fetched by that session now. Porkbun's
marketplace gave us something useful instead: asking-price comparables — 774
UK listings pulled already, with a .com pull to follow.

A first categorisation pass has sorted a third of the names into families
(financial, home & garden, AI, web design, and so on); a second pass on the
long tail is running now. The valuation method is drafted: every domain gets a
tier with its reasoning recorded, the bottom-500 sale is assembled from whole
weak categories so the financial names (and any kept family) stay together,
and keen prices never go below the £150 you already charge as a transfer-away
fee. The 19 domains carrying live sites are marked keep regardless.

One honest miss: I searched all 646 stored conversations on this machine and
the earlier .co.uk/.uk valuation discussion you remember isn't in any of them —
it likely happened on claude.ai in the browser, or on another machine. What I
did recover: your $12,000 Afternic floor on relojistas.com, the £150
transfer-away ruling, and the domain-value ladder doc. **If you can find that
old conversation (or just paste its conclusions), it would sharpen the
starting point — otherwise we price from the data now arriving.**

Things only you can do, gathered in one place:
1. Nominet walk (the two `!` commands above) — the ~1,500 .uk list.
2. Afternic: export your portfolio CSV into the afternic lane's inbound/
   folder, plus their bulk-upload template — that's the current-prices column
   AND the eventual repricing vehicle.
3. Atom.com: 58 of the Spaceship domains are listed there — an export from
   that dashboard if you want those asks in the comparison.
4. Porkbun: flip the global API-access toggle (Account Settings → API) —
   needed later for repricing writes, not for the valuation itself.
