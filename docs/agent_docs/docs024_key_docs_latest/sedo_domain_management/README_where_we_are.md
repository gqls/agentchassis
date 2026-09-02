# Sedo domain management — where we are (append-only, newest at the bottom)

## 2026-09-02 — lane opened

You asked for Sedo to be set up so Claude can manage your domains there.

Good news first: Sedo does have an API, and it covers everything you'd
want — listing your domains, putting them up for sale, changing prices,
and reading parking statistics. I've already proven from our own cluster
that we can reach it and that it answers properly.

I've built the tool that will make the calls, and tested every part of it
that can be tested without an account. It's designed so your Sedo password
and keys never appear in any chat transcript — they live in a sealed
cluster secret, and the calls fire from inside the cluster.

What only you can do (about 15 minutes of your time, then a wait):
1. Create the Sedo account at sedo.com — keep the password 16 characters
   or shorter (their API caps it there).
2. Register for their partner programme (their precondition for API use).
3. Email api@sedo.com from the account's email asking for API access —
   there's a ready-made draft in the RUNBOOK, §2. They reply with two
   codes (a Partner ID and a SignKey).
4. When the codes arrive, follow RUNBOOK §3 — three copy-paste commands in
   your own terminal that seal the four values into the cluster. After
   that, any session can manage the domains on your say-so.

Nothing else moves until Sedo approves the access request. One thing I've
deliberately NOT set up: automatically moving any domain's parking or
nameservers to Sedo — a couple of your domains are parked at Dan/Afternic
and other work depends on that, so re-pointing stays a per-domain decision.
