# SUMMARY 2026-08-24b — the web admin console and the customer-links door

Second read-out of the day — the morning one (`SUMMARY_2026-08-24_…`) marked the Builds
screen being built; this one marks the exposure work being finished and live.

## What we're trying to do

Give the owner one browser-reachable place to watch and steer the website factory — every
site, every build step by step, every instruction the writing agents follow — and give
customers the links they need, without either door exposing the cluster that runs it all.

## Where we've come from

The console went public earlier this week at admin.apis.uk, behind Cloudflare's login
wall, through an outbound-only tunnel, over an encrypted leg that can reach exactly four
services. The customer confirmation links, though, still lived on the shopfront domain,
protected only by an accident — a parking redirect nobody chose as a security control.
And a link click alone could confirm a contract handover, which meant a corporate mail
scanner opening links could confirm things no human decided.

## What we've done

Today, four things. The Builds screen — the feature this workstream exists for — was
built, tested and approved by review: a per-site timeline of each build's stages with
durations, that refuses to trust the platform's optimistic status labels, plus warning
badges on page sections rebuilds have silently overwritten, and an instructions editor
that now says which documents are enforced rules versus wishes, and can no longer be
tricked into silently disabling a site's claims checking.

The customer links got their own front door, and it is live tonight: links.webdesign.uk,
deliberately public, exposing exactly one token-shaped path. We hardened it until nothing
malformed can even cross into the cluster, put a rate limit at Cloudflare's edge, and
then attacked it from outside to prove every layer: junk paths die at the box, a
properly-shaped link travels the whole way through, and hammering the address gets
blocked after a handful of requests — watched happening, not assumed.

The owner ruled the remaining hazard closed rather than accepted: confirmation will
require a real button press on a page, so a scanner's click can never confirm anything,
and no customer email goes out before that page exists. And we mapped every public path
properly: the portfolio sites people can visit never touch the cluster at all — they are
static pages served from storage — so the cluster's public surface is exactly two doors,
one login-walled, one token-guarded, both chosen.

## Where we are now

The links door is live, verified, and idle — no customer tokens exist yet, so there is
nothing to guess and nothing at risk. The Builds screen is committed and approved but not
yet visible: it waits for the cluster's next routine software roll, then the console
image follows it — that order matters and is written down. The cluster access key expired
on its usual three-day cycle this evening, which pauses cluster-side work until the owner
renews it.

## Where we're going

Next: the button-press confirmation page gets built and reviewed, which unblocks the
first customer delivery email. The reviewers owe a pass over the whole exposure posture
now that the second public door exists — the measured map and a concrete containment
candidate are ready for them, and the round fires once the access key is renewed. Then
the console shows its Builds tab after the roll, and the workstream returns to growing
the "contribute" side as the owner asks for it.
