# SUMMARY 2026-08-18 — site delivery and editor

**What we're trying to do.** Turn a finished framework-built site into
something a customer actually owns and can take anywhere: a live hosted copy,
a downloadable archive of the whole site, a delivery email that hands over
every key, and eventually a login where they can edit their own content.
Ownership means three doors: download the site as a file, have it placed into
their own free Netlify account, or pay us (deliberately expensive) to keep
hosting it — plus their own domain, which we register and whose £10-a-month
retention is a separate fee.

**Where we've come from.** Phase 2 — the publishing machinery that mirrors a
built site to a hosted copy and re-publishes only when something actually
changed — was proven live two days ago on the noted.co.uk canary. The next
day the owner settled the whole delivery architecture in five discussion
rounds: separate fees, no free custom-domain serving, Netlify connect offered
during the build wait, the delivery email as the account surface for now, and
our own nameservers as part of the domain programme.

**What we've done.** Built, reviewed and proved the ZIP deliverable — the
single artefact all three ownership doors hand over. The code gathers every
file of the built site, packs it into one archive, checks its own work twice
(the archive really contains every file byte-for-byte, and the stored copy is
exactly the right size — a silently shortened download was the failure we
most wanted to design out), and returns a download link that lasts seven
days. The review council approved it first round; chasing one of its
advisory comments uncovered and fixed a real bug that would have made every
normal request fail. Yesterday's fresh build shipped it; today the
configuration was switched on and the whole chain ran against the live
canary site: the archive downloaded, contained exactly the site's eight
files byte-for-byte, the link worked while valid and refused once expired,
and an oversized-site alarm was deliberately triggered and fired while the
archive still completed in full. Very large sites will never get a cut-down
archive: everything, or a loud alert.

**Where we are now.** Phases 2 and 3 are both complete and live-proven.
Cutting an archive is on demand only — nothing does it on a schedule, and
nothing will until the delivery email asks for it. Nobody outside the team
can reach any of this yet; the pieces exist and work, but no customer-facing
flow triggers them.

**Where we're going.** Phase 4 is the handover: a timestamp that marks a site
as delivered, and the delivery email that carries every link — the ZIP
download, the Netlify connect button, the hosting payment link, the domain
subscription, and the customer billing page Stripe hosts for us. After that,
Phase 4b places sites into customers' own Netlify accounts, and Phases 5–6
are the customer login and editor. First revenue still waits on the owner:
Stripe keys and the webhook exposure; the domain programme waits on the
second Nominet TAG.
