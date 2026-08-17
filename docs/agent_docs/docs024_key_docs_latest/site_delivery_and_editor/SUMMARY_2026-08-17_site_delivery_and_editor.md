# SUMMARY — site delivery and the customer editor — 17 August 2026

*Second summary for this workstream (the first was 16 August — a new file, as
the rules require, because the read-out genuinely changed: yesterday was
about machinery working; today the shape of the business it serves was
decided). Full reasoning and every decision:
`PLAN_2026-08-17_delivery_architecture_decisions.md` in this directory.*

## What we're trying to do

Unchanged in ambition, sharpened in shape: sell a finished website for £149 —
arguably one of the most competitive offers anywhere — and deliver it so
automatically that high volume is a pleasure rather than a burden. The owner
wants a trouble-free operation: no hosting empire, no unpaid ongoing work, no
human steps in the delivery path.

## Where we've come from

Yesterday the publishing machinery was proven in production end to end: a
timer notices a built site has changed, copies it to its hosted address,
checks the served bytes match byte-for-byte, and only then records success —
demonstrated both ways on a live canary site, including the correct refusal
to publish when nothing had changed. With the machinery trustworthy, the
owner called a halt to code and asked the bigger questions: who hosts the
customer's site, who owns it, what the domain is worth, and how the customer
edits and pays.

## What we've done

Today was a planning day — five rounds of discussion, each grounded in
things we measured rather than assumed. The decisive discoveries: no hosting
company offers a way to create customer accounts automatically (and faking it
with temporary emails is how platforms get mass-suspended); but Netlify's
"connect" flow — where the CUSTOMER creates their own free account and
clicks one button — is fully automatable on our side and gives real
ownership from birth. Meanwhile our Nominet registrar position turned out to
be further along than anyone remembered: login to the registry already
proven from our own cluster, a working registry client already deployed, and
a second registrar tag for customer domains already applied for.

The owner's decisions, in brief: the domain fee (£10/month) and hosting fee
are SEPARATE, with our hosting priced deliberately high beside the free
Netlify option — we are not in the hosting business. The Netlify connect
invitation happens while the customer waits for their build, so by delivery
the hosting question has usually answered itself. A site whose customer
chooses nothing stays visible on our preview address, but its own domain
shows a "choose a home" page — no open-ended free hosting. The customer's
"account page" is simply the delivery email at first (every link in one
place, with Stripe running the billing screens), later folding into the
editor's front page. And we will run our own nameservers, because we
register every domain and the answering job is ours wherever the site lives
— it also means no outside company can ever hold the estate hostage.

## Where we are now

All architecture forks are closed; nothing is blocked on thinking. The
machinery from yesterday is untouched and still proving itself hourly on the
canary. Two keys from the owner still gate first revenue (the Stripe pair
and a webhook exposure fix), and the second Nominet tag gates only the
domain programme. Three investigations are deliberately parked: the
whole-architecture scale review (including whether we need our own
clusters), the busy-site payment thread, and news feeds as a paid-hosting
perk.

## Where we're going

Back to building, in this order: the ZIP deliverable next — it is both what
the customer downloads and exactly what gets uploaded to their Netlify
account, so one build serves both doors. Then the handover step and the
emails that carry every link; then the Netlify connect backend on the proven
publishing seam; then customer login and the editor. The domain programme —
choosing a name, registering it at the registry, running our own DNS, and
the £10/month retention link — runs alongside as soon as Nominet grants the
second tag.
