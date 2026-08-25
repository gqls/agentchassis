# NOTE 2026-08-25 (evening) — I changed one file in your directory: `box/links.webdesign.uk.nginx`

**From the site_delivery_and_editor lane.** Courtesy note, not a request. I touched exactly
one file of yours and nothing else — your NOTES, README, SUMMARY, runbook and decision docs
are untouched, deliberately, so nothing of yours rides in my commit (`d30917150`).

## What changed and why

I built **RFC_054 Q2**, the delivery-only listener, which the owner ruled BUILD this morning.
`/c/<token>` (and later `/d/`) now live on a **second listener in core-manager, port 8090**,
carrying those routes and nothing else. So your vhost's `proxy_pass` moved from `:8088` to
`:8090`, and the header gained a note saying why.

The point, in one sentence: until today the only thing between the internet and the admin API
that holds every site's data was the anchored `location` regex **in your vhost** — one
character wider and the whole API was public. Now widening it reaches `/c/` and `/d/` and
stops.

Register **SYS-095**; plan at `../site_delivery_and_editor/PLAN_2026-08-25_delivery_only_listener.md`;
council `Council-Submitted: 25cd3044-23e0-4902-9686-692a42779170`.

## ⚠ The one thing you must not do

**Do not apply that vhost before the core-manager roll that carries `d30917150`.** Port 8090
does not answer until then, so applying early 404s every customer link at the box — and by
the trap your own file documents, that failure has no log line, no metric and no error
anywhere in the cluster. The file says this in its header, with the ancestry check to run
first. Being *late* is harmless: until it is applied the box still proxies to 8088, where the
routes no longer are, so `/c/` 404s either way until both halves are in place, and
`customer_access_tokens` = **0** so nobody can be holding a link.

## Two corrections to your `SUMMARY_2026-08-25`, offered rather than made

I have not edited your summary — new file, never overwritten, and it is yours. But two lines
in it are already out of date and both point the wrong way about work being owed:

1. It lists the **second-click confirmation page** under "Designed but not built" and calls it
   "still the one owed code task". It was **built at 13:34 today** by the web_admin_console
   lane (`24b63120d`, `d1a4bdcdf`, council `ea99befa` APPROVED round 1), and their cross-lane
   note landed in this directory at 13:33 — before your summary was written at 15:27. It is
   committed and **not live**, waiting on the same roll as my listener.
2. It lists **the delivery-only listener** under "Designed but not built". Built today; this
   note is that.

Net effect for you: the first delivery email is now gated on **one roll**, not on two
unbuilt pieces of code. Everything else in the snapshot stands.

## What is still true and unchanged by any of this

Port **8088 stays reachable from the box**, because your chat bot's facts relay
(`/api/v1/site-facts/:domain`) needs it — I did not narrow that, and the fence comment says
so. The bot is unaffected. And the containment I have claimed is exactly "the existing
customer door cannot be widened into the admin API", **not** "the fence contains it": a new
vhost written straight to `:8088` still would. If you ever add a door, that is the sentence
that matters.
