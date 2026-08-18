# SUMMARY 2026-08-18b — the guard holds, and loanzy.uk is live as an education site

The morning's summary (`SUMMARY_2026-08-18_the_no_prompt_build_put_a_credit_broker_live.md`)
ended with a live page presenting a credit broker that does not exist. This one is written the
same evening because the read-out genuinely changed: the same domain, given the same absence of
instruction, now serves the opposite kind of site.

## What we are trying to do

Give webdesign.uk one honest pair — a prompt and the site it produced — because its approved
lead promises the visitor can see exactly that, and is forbidden from mentioning examples until
such a pair exists. loanzy.uk is the domain carrying the attempt.

## Where we came from

Given only its domain name, the framework decided loanzy.uk was a UK credit-broking business:
an eligibility checker, a panel of FCA-regulated lenders, and revenue from per-referral fees.
One page of it reached the live web before the build was stopped, and stopping it exposed two
mistakes of ours and two of the platform's — recorded in the morning summary, `WRONG_CALLS.md`,
`LANDMINES.md` and `bugs_open/304`.

The owner's instruction was then narrow and correct: *hint to the classifier that this isn't a
valid option unless the brief specifically asks for it.*

## What we have done

**The rule is live** (migration `464`, register entry `CGV-032`): a regulated business model —
credit broking, lending, debt advice, mortgage or insurance or investment arranging, payment
services, claims management — is not an available answer unless a mission explicitly asks for
one. A domain name tells the classifier the SUBJECT, never that the site may broker or advise.

**It was proved as an A/B, not asserted.** Same domain, same site row, no mission either time,
the previous specs superseded so the classifier could not read its own answer back. The only
difference was the prompt. Services went from *Personal Loan Matching · Loan Comparison Tool ·
Eligibility Checker · Lender Lead Facilitation* to *Loan Explainers · Loan Cost Calculator ·
Borrowing Guides · Glossary · Rights and Regulations*. Both runs are committed verbatim as
`EVIDENCE_run1_*` and `EVIDENCE_run2_*`.

**Then a genuine from-scratch build**, triggered through the front door with the domain string
and nothing else. It is live now.

## Where we are now

`https://loanzy.uk/` serves *"Loanzy UK — Honest Loan Calculators & Plain-English Guides"*.
The home page says, unprompted: **"We do not arrange loans, introduce you to lenders, or take a
cut when you borrow anywhere."** `get-help.html` sends readers to StepChange, MoneyHelper and
National Debtline, and states plainly that debt advice *"is a regulated service we are not
authorised to provide"*. `calculators.html` advertises that nothing you type leaves your
browser. Every regulated-sounding term on the site — representative APR, lender, apply — appears
in an explanatory sentence, never as an offer, and the only outbound links on the site go to
free debt-advice charities.

The contrast worth keeping: run 1's business model was capturing a borrower's details and
selling the introduction. This site's selling point is that it collects nothing. Same domain,
same silence from us.

**Three things are unfinished, and none is about the guard.**
The consumer-rights page did not build: it was refused by content validation with 20 identical
`{{end}}` blockers — a known platform defect, `bugs_open/260`, which a second lane hit on two
pages the same day. Our contribution is that it fires on greenfield builds and that its real
cost is a page that never exists plus a dead link on a live page, which no scan of stored
components can see. The guides index built nothing because nothing supplied guides for it to
index — the builder refused rather than shipping an empty shell, which is correct. And the
navigation is still run 1's: the footer offers *Check Eligibility* and *Lenders*, pointing at
pages that no longer exist, because the nav rebuild refused to run while the page corpus was
incomplete. The guard fixed what the framework WRITES; it cannot reach what was already STORED.

Eight tool pages are queued for the owner-aware build path rather than the generic one, and the
imagery run is still filling in.

## Where we are going

Re-triage the nav rebuild once the corpus settles, so the site stops advertising an eligibility
checker it does not have. Watch `bugs_open/260` for the rights page. Then judge whether this is
the example pair webdesign.uk should show — the honest question being not "is it impressive"
but "is this what a customer's one-shot build really produces", because that is the only claim
the gallery is allowed to make.

One control is still owed on the guard itself: that a brief which DOES ask for a regulated model
still gets one. Until that runs, we have proved it declines, not that it still complies.
