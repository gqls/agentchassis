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

## 2026-09-04 — checked, Appleby domains pulled out: 2,888 in the sheet now

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

**The sheet you'd upload now**: `outbound/SEDO_IMPORT_2026-09-04_draft3.xlsx`
— 2,888 domains, all make-offer, no prices. That's ready whenever you are.

## 2026-09-04 (later) — williama.co.uk out, and the wyke/pastured-egg names pulled

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

**Current sheet**: `outbound/SEDO_IMPORT_2026-09-04_draft5.xlsx` —
**2,879 domains**, still all make-offer with no prices. This supersedes
the draft3 file above.

## 2026-09-03 (correction) — the two dates and three filenames just above were wrong

Small correction to what I wrote above: I'd mis-dated that whole stretch
of work as "2026-09-04" — it actually happened today, 2026-09-03 (I'd
picked up the wrong date from a filename in another thread's folder
instead of checking the clock myself). The content above is accurate,
only the dates and the filenames were off by one day. The actual files on
disk are named with today's date:
`outbound/SEDO_IMPORT_2026-09-03_draft3.xlsx` and
`outbound/SEDO_IMPORT_2026-09-03_draft5.xlsx` (draft5 is still the
current one to upload — 2,879 domains, unchanged). If you already
downloaded or looked for a file with "09-04" in the name, that's why it
wasn't there.

## 2026-09-03 (urgent) — copyonline.co.uk pulled from the sheet, one more check tightened

You told another thread just now that copyonline.co.uk is your site but
might become your wife's — a keeper, not something for sale — so I've
pulled it out straight away.

Worth explaining honestly how it got in, because the real reason is
different from what it looked like at first: it wasn't a hole in the
logic, it was timing. copyonline.co.uk was only just set up in our system
— 30 minutes after I'd last checked which domains are live sites, and I
reused that same check across a few sheet regenerations instead of
re-running it each time. I've now made re-checking mandatory before every
future sheet, so a newly-started site can't slip through the same way
again. I also checked whether anything else new had appeared in the same
window — nothing had; copyonline.co.uk was the only one.

**Current sheet**: `outbound/SEDO_IMPORT_2026-09-03_draft6.xlsx` —
**2,878 domains**, everything else unchanged from before, this one domain
removed. This is now the one to upload.

## 2026-09-03 (urgent) — found a whole second group of live sites that weren't being protected

The valuation thread swept the domains more carefully and found something
important: some of your older sites — built before the current
Cloudflare setup — sit on a different hosting system entirely (the older
"Clook" nameservers), and my checks had only ever looked at the newer
system. So 33 of them were sitting in the sheet as ordinary for-sale
stock, completely unprotected. Two in particular I want you to know about
by name before anything moves further: **wpx.uk**, which carries your own
email address, and **designconsultancy.co.uk**, your own company's
domain. Two more, **leopardess.co.uk** and **leopardess.uk**, sit right
next to a client's domain I'd already kept off the list for exactly the
reason you'd expect — I don't think "all live sites" was meant to include
your own email or a client relationship, so I've held those out rather
than assume, alongside the rest of the 33.

**Current sheet**: `outbound/SEDO_IMPORT_2026-09-03_draft7.xlsx` —
**2,839 domains**. Everything above about pricing and withdrawals still
applies; this just removes those 39 newly-found domains too.

## 2026-09-03 (later still) — you said list live sites too, priced high: that's now its own separate piece of work

You told me: list the live sites as well, and price them high —
webdesign.uk could be worth over a million within a year. I've taken that
seriously, but I haven't just dumped them into the current sheet, because
the current sheet has every price left blank — putting a domain you think
is worth seven figures into a sheet with "make an offer, no price" would
undersell it badly, arguably worse than not listing it at all.

So: live-site listings are now their own piece of work, separate from the
portfolio sheet. The valuation thread needs to work out real asking
prices for each one before anything goes anywhere — I've asked them to
take that on. Once real prices exist, building that sheet (or the
individual Sedo listings) is quick.

One thing I still need from you, whenever you have a moment: does "list
all live sites" really mean ALL of them, including your own email domain
and company site, and the client's neighbouring names? My working
assumption is no — those specific ones stay off the list until you say
otherwise — but I'd rather have that confirmed than guessed.

## 2026-09-03 (later still) — one more site held out, and please double-check which webdesign you meant

Two more things before any live sites go anywhere.

**leopardessconsulting.co.uk — I want your explicit yes on this one
specifically, not folded into "all live sites."** This is your client's
actual site, and it's the exact case you wrote down back on 24 July when
you said "buy this site" on a paying client's site would be a
relationship breach. I don't think your "list all live sites" answer had
that specific site in mind — you were answering a question about
relojistas.com — so I've kept it off the list and I'd like you to say
this one's name specifically if you do want it included, rather than me
assuming a broad answer covers it.

**webdesign.uk and webdesign.co.uk are two different domains.** Your
example — "could be worth over a million" — was webdesign.uk, an
18-page site. webdesign.co.uk is a different, much bigger site (155
pages) — it's your actual webdesign business's shopfront. Worth
confirming which one (or both) you meant, since pricing or listing the
wrong one would be an expensive mix-up.

## 2026-09-03 (later still) — your list applied, cartoon.co.uk's price floor noted, one domain still needs an answer

Went through your keep/release list against the 39 I'd found. It matched
cleanly except for one: **2v.uk** — you didn't say either way, so I've
left it excluded for now rather than guess.

Everything else is done: your 17 keeps stay off the list, your 21
releases are back in the ordinary sheet as normal for-sale stock (not the
high-price tier — nothing you said suggested those were special, just
that they didn't need protecting).

Noted what you said about cartoon.co.uk — over £5,000 paid for it, so it
must not go out underpriced. It's currently sitting at no price, same as
everything else in the sheet, and I've made sure the valuation thread has
that figure before they set any number on it. Worth mentioning: if
there's a real cost behind that one you hadn't told me, there may be
others among the 21 I don't know about — happy to hold any more you think
of the same way.

**Current sheet**: `outbound/SEDO_IMPORT_2026-09-03_draft8.xlsx` —
**2,860 domains**.

## 2026-09-03 (later still) — both open questions closed

Got it: 2v.uk stays off permanently, and both webdesign domains are in
play for the high-value listing (not just the one you mentioned first).

I've also noted what you said about them becoming the same endpoint one
day — passed that along to the valuation and about-page threads, since
it matters for pricing and for which one eventually carries the site.
Nothing for me to act on there myself, just making sure it's on the
record so nobody prices or plans around the current two-separate-sites
picture as if it were permanent.

## 2026-09-03 (later still) — a correction, and your call on the webdesign pair

Small correction on my part: I had it backwards earlier — webdesign.uk
(not .co.uk) is the actual shopfront, and it's the same domain you gave
as your seven-figure example, not two separate things. Checked it
properly against the record before telling you, so that's now right.

You said quote them as a pair, which matches what the valuation thread
independently recommended too — selling either one alone would hand part
of the eventual merged value to whoever buys it, since the two would no
longer be under your control together. They'll now price the two as one
combined figure rather than two separate numbers.

## 2026-09-03 (later still) — no Buy Now prices anywhere, and now it can't happen by accident

Checked every sheet I've built so far, all eight versions, not just the
current one — none of them has ever carried a Buy Now price or any price
at all. Every row has genuinely been blank.

You also asked if we can ban this except when you do it manually — done.
The tool that builds the sheet now refuses outright if anything tries to
set a Buy Now price or a fixed asking price, unless a specific switch is
turned on for that one run. That switch isn't something I'd flip myself
from a conversation — it only exists for you to use deliberately, when
you've actually decided a specific run should carry real prices. I tested
it both ways: it genuinely stops the sheet being built when the switch is
off, and works normally when it's on. Minimum offer amounts (the floor
under "make an offer," like cartoon.co.uk's £5,000) are unaffected —
those exist to protect against lowballs, which is the opposite of what
this is guarding against.

## 2026-09-03 (later still) — the built-site listings are unblocked, and the floor policy is settled

You said you'd rather bear with lowball offers than have a visible
minimum on the higher-value sites that anchors what buyers think they're
worth — so the piece that was holding relojistas.com, both webdesign
domains, and everything else I'd been keeping back purely because it
didn't have a real price yet is gone. I asked you whether that meant
truly no minimum or a small nominal one to filter spam, and you said
truly none for now.

Since "no minimum, make an offer" is exactly the same shape every other
domain in the sheet already has, I folded all of those sites straight
into the main sheet rather than keeping a separate list — nothing left
waiting on a price any more, except the specific ones you've named to
stay off entirely (the Appleby names, Wyke Farm/pastured egg,
copyonline.co.uk, and leopardessconsulting.co.uk).

Then you clarified further, and it's a genuinely useful distinction:
Sedo's own minimum field can carry a real number whenever you give me
one or we agree one together — that's completely separate from whatever
the actual website shows, which never displays a price at all. So
"blank for now" isn't a permanent rule, it's just where things sit until
a real figure exists — and when one does, it goes straight into Sedo,
not derived from any automatic scoring, just what you tell me or what we
agree.

**Current sheet**: `outbound/SEDO_IMPORT_2026-09-03_draft9.xlsx` —
**2,943 domains**. This is essentially the whole estate now, minus only
what you've specifically asked to keep off.
