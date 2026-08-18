# Where we are — loanzy.uk

Plain prose, append-only, newest at the bottom.

## 2026-08-18

Nothing has ever been built on loanzy.uk. The domain itself is in good shape — it has a
Cloudflare zone, a certificate, and both worker routes, so it answers instantly with a small
404 that means "no site here yet" rather than anything being broken. That puts it ahead of
most of the portfolio, where domains still point at registrar parking pages.

The FCA-rules site you remembered is real, but it is lendzy.co.uk, not loanzy.uk — one letter
apart, and built the day after loancash.co.uk as a second run of the same brief. Both are
live and were redeployed this morning.

The decision this session: loanzy.uk becomes webdesign.uk's first example site, built only
from a customer prompt, with no positioning entry written for it. That matters more than it
sounds. The webdesign.uk page lead approved today is "show the work, promise nothing" — the
promise being that you can see real sites and the exact prompt that produced each one. The
copy is currently forbidden from mentioning examples at all, because the four sites we could
point at were not built the way a customer's site gets built. Building loanzy.uk from a
prompt and nothing else produces the first honest pair.

What is needed next is the prompt itself, which is yours to write, because it gets published
next to the site. One caution worth a sentence: the domain sounds like a lender, and a
convincing fake lender on a live UK domain is a compliance problem rather than a demo. Either
the demo is not a financial firm — which also shows off the new "any sort of site" line — or
the page says plainly that it is an example.

## 2026-08-18, later

We tried it your way — the framework given nothing but the domain name — and it answered
confidently and dangerously. It decided loanzy.uk should be a loan comparison and matching
platform: a panel of FCA-regulated lenders, an eligibility checker, and money coming from
lenders paying us for each borrower we refer. That is credit broking, which is a regulated
trade. It even found two unrelated companies trading as "Loanzy" elsewhere in the world and
treated their business as evidence for what ours should be.

I stopped the build, but not cleanly, and one page got out before I did. It sat on the live
domain saying we search a panel of FCA-authorised lenders. The account of exactly how that
happened, including the three separate mistakes I made trying to stop it, is in the summary
document written this afternoon — you asked for it in full and it is in full.

Two things came out of it that are worth more than the site would have been. The first is
that the system already knew: two steps before any page was written, the briefing agent had
noted that an FCA authorisation number "must be obtained before launch" and that the lender
panel was not confirmed. It wrote that down and carried on, because nothing gates on a note.
The second is that our unpublish path cannot unpublish the last page of a site — the deletion
empties the folder, the deploy skips folders that do not exist, and both halves report
success. That is filed as a bug.

The classifier now carries the rule you asked for: a regulated business model is not an
available answer unless the brief explicitly asks for one, and a domain name tells it the
subject, never that it may broker or advise. That is live in configuration, though nothing has
been built through it yet, so it is not yet proven.

The page from this morning is still up. Both ways of removing it were blocked by my own
sandbox — one deletes production data, the other writes to the deploy repository — so that
one needs you.

## 2026-08-18, evening — it is live, and it is the opposite site

loanzy.uk is up. Given nothing but its own name — no brief, no facts, no contact details — the
framework has built a loan education site: calculators that run in your browser, a glossary, a
page pointing people at free debt advice, and a home page that says in its own words "we do not
arrange loans, introduce you to lenders, or take a cut when you borrow anywhere."

This morning the same domain with the same silence from us produced a credit broker with an
eligibility checker and a lender panel. The only thing that changed in between is the rule you
asked for, which now tells the classifier that a regulated business is not an available answer
unless the brief asks for one. I proved it by running the same thing twice and changing nothing
else, and kept both answers on file so the comparison can be checked rather than taken on trust.

Three things are unfinished and none of them is about that rule. The rights page did not build:
it hit a known platform bug that leaks raw template syntax into the page, which another team hit
on two pages today, so I added our case to it rather than filing a duplicate. The guides page
built nothing because there are no guides yet, and the builder refused to ship an empty shell,
which is the right call. And the menu is still the old one — it offers "Check Eligibility" and
"Lenders", pages that no longer exist — because the navigation rebuild sensibly refuses to run
while half the site is missing. That last one is the only thing on the live site that
contradicts itself, and it is next.
