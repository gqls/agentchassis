# PLAN 2026-08-25 — DNS plan B: own authoritative DNS + Cloudflare for SaaS

**Status: SCOPED, NOT STARTED. No infrastructure built, deliberately** — see §7.

**Why this file exists.** The owner ruled **"let's start on plan B"** on 2026-08-21 during the
scale review. The `dispatch_throughput` lane routed execution here and wrote, in its own
handoff: *"belongs to the domain programme lane; pointer left in site_delivery_and_editor
NOTES; do not build here; **check they picked it up**."* Nobody had, for four days. This is
the pickup.

Own authoritative DNS was separately ruled **GO** on 2026-08-17
(`PLAN_2026-08-17_delivery_architecture_decisions.md`, owner decision 4) as part of the
**domain programme**, not the scale review.

---

## 1. What plan B is, in plain terms

Today every domain we manage gets **its own Cloudflare zone**, in **one Cloudflare account**,
with a Worker (`portfolio-sites-router`) routing requests to the B2 bucket. That is "plan A".
It works, it is free, and it is fine for tens to hundreds of domains.

Plan B replaces the DNS half with two things:

1. **Our own authoritative DNS** — a boring self-hosted nameserver pair, one template zone
   file per domain, fully automatable. We need this *anyway*: being the registrar means we
   must host DNS, and zone-per-domain on Cloudflare is how we do that today.
2. **Cloudflare for SaaS custom hostnames** on **one** zone — customer domains attach as
   custom hostnames instead of as zones, with automated TLS. ~$0.10/hostname/month past the
   first 100.

What it buys: no zone sprawl, no per-zone API-limit pressure, a much smaller token blast
radius, and a path past the account zone cap.

## 2. Current state, measured today rather than carried forward

| thing | value | how |
|---|---|---|
| Cloudflare zones on the account | **41** `[MEASURED 2026-08-25]` | CF API `/zones?per_page=1`, `result_info.total_count` |
| zone plan tier | **41 of 41 "Free Website"** `[MEASURED 2026-08-25]` | CF API `/zones`, `plan.name` |
| sites in the database | **51**, every one `build_status='pending'` `[MEASURED 2026-08-25]` | `SELECT build_status, count(*) FROM sites GROUP BY 1` |
| growth since the research doc | **39 → 41 in 7 days**, i.e. **~2/week** | research doc measured 39 on 2026-08-18 |
| existing CF automation | `scripts/cloudflare/{add_www_redirect,deploy_worker}.sh` + `worker.js`, token at `~/.config/cloudflare/portfoliotoken` | repo |

## 3. ⚠ The urgency is contingent, and the cap behind it is UNVERIFIED

The scale review's argument was: *at 50 domains/day the ~1k zone cap is ~3 weeks of
promotion away, so plan B's readiness may need to PRECEDE the first big promotion.* That
reasoning is sound. Two things about it need saying plainly before anyone spends weeks:

**(a) The cap figure is inherited, not established.** Every statement of "~1,000 zones" in
this estate traces to the 2026-08-18 research doc; I could not verify it from here — the
Cloudflare token is zone-scoped, so the `/accounts` endpoint returns nothing and the account's
own limits are not readable with it. **[UNVERIFIED]** — and it is the entire justification for
the programme's timing. **Verifying it is a prerequisite, not a formality**: the difference
between a 1,000-zone cap and no practical cap is the difference between "build this soon" and
"build this when we are the registrar and want to be". Cheapest routes: ask Cloudflare support
directly, or read the limit off the dashboard's account page (a human with dashboard access can
answer this in a minute; a zone-scoped token never will).

**(b) At today's rate the cap is years away; at promotion rate it is weeks.** 2 zones/week
puts 41 → 1,000 roughly nine years out. 50/day puts it ~19 days out. So the trigger is not a
date — **it is "promotion is scheduled"**. And promotion cannot start while the webdesign.uk
shopfront is parked, which it is, pending the owner's copy/design revision and two undecided
product questions (`SUMMARY_2026-08-25_webdesign_uk_build_service.md`).

**What follows, and it is a recommendation not a decision:** plan B should be *designed and
decided* now — that is cheap, and it is what "start on plan B" most needs — but the
**infrastructure build should be timed off the promotion decision, not off the calendar.**
D12 in the scale review's decision table asked exactly this ("calendar-based if promotion is
planned"), and the honest answer today is that promotion is not planned yet.

## 4. The design, and the decisions inside it that are not mine

Plan B is **two independent halves**, and conflating them is the first mistake available:

### Half 1 — authoritative DNS (needed regardless of hosting, ruled GO 2026-08-17)

Every domain we register needs DNS whether the customer hosts with us, on Netlify, or nowhere.
Three zone templates were already named in the 08-17 ruling: **ours / netlify /
choose-a-home page**. EPP sets NS at registration.

Open shape questions, all needing an owner or architecture answer:

| # | question | why it is not a build detail |
|---|---|---|
| D-a | **What runs the nameservers?** (e.g. a managed authoritative provider vs self-hosted BIND/Knot/PowerDNS pair) | "Self-hosted" was the 08-17 word, but a nameserver pair is a **new standing production dependency with its own uptime obligation** — if it is down, every customer domain is down, including ones we do not host. That is a bigger blast radius than anything this estate currently runs. |
| D-b | **Where do they live?** Two hosts, genuinely independent (different provider/region), or the existing box + one more? | The whole point of a pair is that one failure does not take DNS out. A pair on one provider is one failure. |
| D-c | **Who writes zone files, and from what?** | The estate's answer should be the database — `sites.domain` plus a template — with the same "the register is the wire" discipline. This is the part that is genuinely automatable and cheap. |
| D-d | **Migration of the 41 existing zones** | The 08-17 ruling already says: "first customer domains may ride zone-per-domain until the pair is live (migration = one proven EPP call each)". So this is incremental, not a cutover. |

### Half 2 — Cloudflare for SaaS (needed only for domains WE host)

Custom hostnames on one zone, automated TLS, ~$0.10/hostname/month past 100. Only our-hosted
sites need it; a customer who moves to Netlify or elsewhere does not.

| # | question | why |
|---|---|---|
| D-e | **Does CF for SaaS require a paid plan tier, and which?** | All 41 zones are Free today `[MEASURED]`. CF for SaaS is not a free-plan feature as far as I can establish, and I have not verified the tier or price from inside the estate. Same rule as §3(a): do not build against an unverified commercial constraint. |
| D-f | **How does the Worker route change?** | Today `worker.js` keys objects on `<hostname><path>`. Custom hostnames change what `hostname` is at the edge, and `add_www_redirect.sh`'s header already records that zones are **not** uniform (24 wildcard-routed, 12 apex-only, some not served by this worker at all `[MEASURED 2026-08-18]`). A migration that assumes uniformity breaks the non-uniform ones silently. |

## 5. What I recommend, in order

1. **Verify the two commercial facts** (§3a cap, §4 D-e tier/price). Both are owner-or-dashboard
   answers, both are cheap, and both gate whether the rest is worth doing now. **Nothing should
   be built before these.**
2. **Decide D-a and D-b** — the nameserver shape and where it lives. This is an architecture
   question with a real blast radius and it should go to the architecture seat / an RFC, not be
   settled inside a build. A self-hosted authoritative pair is a new standing production
   dependency for *every customer domain*, which is a larger commitment than anything currently
   running here.
3. **Then, and only then, build half 1's automation** (D-c): zone template + writer driven from
   the database. That part is boring, cheap and reversible, and it is the piece that pays off
   even if half 2 never happens.
4. **Half 2 last**, timed off the promotion decision.

## 6. What would make this urgent again, stated so it is checkable

- The shopfront unparks **and** promotion is scheduled → the 19-day clock from §3(b) starts.
- Zone count crosses a threshold worth watching. At 2/week there is no threshold worth
  watching yet; re-measure with the one-liner in §2 rather than trusting this file.
- The cap turns out to be materially lower than ~1,000 (§3a) → re-plan immediately.

## 7. Why nothing is built in this commit

Three reasons, in order of weight:

1. **The premise is unverified** (§3a). Building an alternative to a cap nobody has confirmed
   is how an estate acquires a standing dependency it did not need.
2. **The shape needs an owner ruling** (§4 D-a/D-b). A nameserver pair for every customer
   domain is architecture-scope by this repo's own test — a shared mechanism whose blast radius
   is "every domain we manage" — and the guardian seat's standing position is that such a thing
   arriving inside a build is the defect, not the shortcut (`bugs_closed/124`).
3. **The trigger is contingent and currently not met** (§3b). The launch is paused; promotion
   is not scheduled.

Doing the cheap, decidable part now and holding the expensive, undecided part is the whole
content of "start on plan B" as I read it. If the owner means "build the nameservers now", that
is a different instruction and this plan is ready to take it — §4's questions are what I would
need answered first.
