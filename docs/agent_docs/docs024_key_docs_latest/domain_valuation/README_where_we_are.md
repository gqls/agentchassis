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
