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

## 2026-09-02 (evening) — your importer spreadsheet decoded and mapped

You sent over Sedo's example spreadsheet for bulk domain listing. I've
decoded it and written up what each column means, and matched it
column-for-column to the API call we'll use — so the same sheet can drive
either route: you uploading it on Sedo's website, or us pushing the data
through the API. The write-up is in the RUNBOOK, §6.

Two things worth knowing from Sedo's own docs: adding a domain doesn't take
effect instantly (it goes through their checks first, and any failures come
to you by email), and every domain added gets parking switched on
automatically — even ones marked not-for-sale. That second one matters
because some of your domains are parked elsewhere, so "add it to Sedo"
stays a per-domain decision, exactly like the nameserver question.

Still waiting on the account and the two API codes from Sedo (your steps:
RUNBOOK §1–§3). Everything on my side is ready — once the codes are sealed
in, the first call will list one domain, checked end to end, before any
batch.

## 2026-09-02 (late evening) — your spreadsheet is ready: 1,318 domains

You told me you already have the account (info@designconsultancy.co.uk)
and partnership status, and asked for the full portfolio in Sedo's import
format. Done — the file is
`outbound/SEDO_IMPORT_2026-09-02_draft1.xlsx` in this folder, with a
plain CSV next to it if you want to eyeball it first.

What's in it: 1,318 domains — everything Dynadot (451), Porkbun (683) and
Spaceship (203) hold, checked complete by each of those threads today.
What's NOT in it, deliberately:
- Your 19 live websites (websitedesign.com, boxingonline.com,
  relojistas.com, vetcomparison.uk and so on) — I've fenced out anything
  actually serving a site. The list is in the same folder; say the word if
  any of those SHOULD be for sale and I'll put them in.
- The ~1,500 .uk names at Nominet — still needs you to run the two
  Nominet commands (the nominet thread has them ready). The moment that
  lands I'll cut draft 2 with everything.

Prices: every domain is set to "make offer", for sale, with NO price yet —
the valuation thread is working through the portfolio now (Dynadot's own
appraisals are being pulled tonight as one input). Two honest caveats if
you upload this version straight away: (1) make-offer with no minimum
invites lowball offers — the minimums arrive with the valuations, as a
second upload that updates the same domains; (2) Sedo switches parking on
for every domain you add — that changes nothing at Afternic unless
nameservers move (which nobody is doing), but you'll see them all as
"parked" in your Sedo dashboard.

Your three remaining moves, whenever suits: upload the sheet (or wait for
the priced version), run the Nominet walk, and send the API-access email
(RUNBOOK §2) so future changes don't need manual uploads at all.

Small update, same evening: the Dynadot count was double-checked against
your control panel — complete, and grew to 453 with the two domains you
added today (overhead-cranes.com, paper-cups.com). The sheet now holds
1,320. Since you add domains from time to time, we'll re-pull all the
registrar lists right before any version you actually upload.

## 2026-09-03 — the full portfolio is in: 2,895 domains, live sites double-checked

The Nominet list arrived — 1,606 .uk names (their first-ever successful
pull; they fixed three bugs in their tooling to get it, so treat tonight's
count as freshly proven rather than long-settled). That's now folded into
the sheet, which is why the numbers below are bigger than last time.

Before folding it in, I went back and checked which domains are actually
live sites more carefully than the first pass. Last time I only caught
sites by looking at which nameservers pointed at Cloudflare — that missed
anything registered through Nominet directly, plus a couple of odd cases
(one domain, adversecreditmortgage.co.uk, shows as a live site in our own
records but its nameservers currently point at a marketplace — the
nameserver check alone would have called that "safe to list", which would
have been wrong). So this time I checked two sources and combined them:
Nominet's own record of which domains they've switched over to Cloudflare,
and a direct read of which domains our system has actually built and
deployed a site for. Together that's **50 domains** now kept off the
sheet, not 19.

**New sheet total: 2,895 domains** — everything at Dynadot, Porkbun,
Spaceship and Nominet, minus those 50 live sites. Still all set to "make
offer, for sale, no price" until the valuations land.

Your remaining moves, whenever suits: upload the sheet (or wait for the
priced version), send the API-access email (RUNBOOK §2), and — if you want
to double-check my live-site list before uploading — it's
`EXCLUDED_live_2026-09-03.txt` next to the sheet, one domain per line.

## 2026-09-03 — checked, Appleby domains pulled out: 2,888 in the sheet now

You asked me to double-check the sheet and pull out the Appleby domains
before you upload. Both done.

**On prices**: checked every single row directly, not just what I
intended to build — there are no Buy Now listings anywhere in the sheet,
and no minimum offer amounts set on anything. Every domain is currently a
plain "make an offer" with no floor at all. That's the interim state we
agreed while the valuation work is ongoing — minimums land as a second
upload once those numbers are ready, so right now nothing stops a very
low offer coming in on anything in the sheet.

**Appleby domains**: pulled all 7 you actually hold — anthonyappleby.com,
appleby.cv, katherineappleby.co.uk, kathyappleby.co.uk, kathyappleby.com,
oliverappleby.co.uk, williamappleby.co.uk. (Three similar-looking names
turned up in the valuation work as comparison domains, not ones you own —
left those alone, they were never in the sheet.) I kept this as a separate
list from the live-sites one, so a future refresh of the live-sites list
won't accidentally lose your Appleby instruction.

**The sheet you'd upload now**: `outbound/SEDO_IMPORT_2026-09-03_draft3.xlsx`
— 2,888 domains, all make-offer, no prices. That's ready whenever you are.

## 2026-09-03 (later) — williama.co.uk out, and the wyke/pastured-egg names pulled

You confirmed williama.co.uk should come out with the Appleby names — done.
You didn't get back to me on the other four person-name domains (ianstirling.com,
kapoor.uk, keeler.uk, anne-marie.co.uk) the valuation thread flagged as
having no obvious family connection, so I've left those exactly as they
were — in the sheet, for sale — rather than guess either way.

Then you asked me to also remove anything like Wyke Farm or pastured egg.
I checked every registrar and registry list fresh rather than assume I
already had them all, and found eight: two Wyke Farm domains with a
hyphen (wyke-farm.co.uk/.uk — the non-hyphenated wykefarm.co.uk/.uk were
already out because they're one of your live sites), plus six pastured-egg
names at Porkbun (pasturedegg.co.uk, pasturedegg.uk, thepasturedegg.com,
and three thepasturedeggcompany.* variants). All eight are out now.

**Current sheet**: `outbound/SEDO_IMPORT_2026-09-03_draft5.xlsx` —
**2,879 domains**, still all make-offer with no prices. This supersedes
the draft3 file above.
