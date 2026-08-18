# NOTE for the webdesign.uk / site-delivery joint lane — the gallery's missing artefact is being built

**From:** the `loanzy_uk_example_site` lane (new, 2026-08-18) · **Needs a reply? No.** One
thing to decide later, named at the end.

## What changed for you

The owner ruled today (**P10**, recorded in `portfolio_positioning/REGISTER_positioning.md`
under L9, commit `f21530d37`): **`loanzy.uk` becomes webdesign.uk's first example site, built
ONLY from a webdesign.uk customer prompt, with no positioning entry written for it.** His
words, in order: *"yes we could use it as an example site"* → *"which would mean no prior
registry entry"* → *"and have it built only from the webdesign.uk prompt"*.

This is aimed straight at your lead's precondition. Proposal F's `writer_block` forbids "see
our examples", any portfolio link, or implying a showcase **until a gallery is live**, because
the four sites in `any_site_type_examples` were not built by the one-shot route — your own
copy-honesty judgement, which the `cross_site_domain` allow-list has no bearing on. A site
built from a prompt and nothing else is the first pair that would populate such a gallery.

## What this supersedes on your side

Your NOTES (2026-08-03 correction) records *"loanzy.uk … is the real new member"* of this
lane's domains, and **P9 (2026-08-15)** gave the domain to you outright. P10 keeps it with
you in substance — it is still your example site — but the *use* is now fixed and the
positioning register will never write a proposition for it. Nothing you have built is
affected: `loanzy.uk` has **0 rows in `sites`, 0 work items**, and serves a 9-byte 404 from
`portfolio-sites-router` (measured 2026-08-18). Nothing was ever wired to it.

## What the domain gives you for free

Zone `18c86604a6066bdb717e11ff28effb48` **active**; worker routes `loanzy.uk/*` **and**
`*.loanzy.uk/*` → `portfolio-sites-router`; apex and www both proxied and answering. No DNS
work is needed — unusual in this estate, where the portfolio lane's open question is literally
"who owns DNS" because ~140 domains still point at registrar parking pages.

## The rule the build will be held to

**Every seeded input must trace to a sentence in the customer prompt.** No hand-tuned
`evidence_base`, no imagery guide we wrote, no second round of steering; empty `facts[]` plus a
figure-forbidding `writer_block` where the prompt attests nothing (the `oufe` precedent — and
the reason webdesign.uk's own seed could carry facts is that the owner had attested them).
Deviations get logged in that lane's NOTES as deviations. If we quietly steer the build, the
published pair overstates the product, which is the exact thing your deferral was protecting.

## The one thing for you to decide, later

Whether one pair is enough to call a gallery live, and therefore whether the `writer_block`
precondition is met. **That is your call, not ours** — this lane will hand over the prompt,
the URL, the cost and what the build got wrong, and change no copy. If you do link it,
`loanzy.uk` will need adding to that site's `allowed_reference_domains`
(`loadAllowedReferenceDomains`, `validate_page_content.go:1462` — opt-in per site, in
production use on `fundamentallyai.com`), or the `cross_site_domain` guard will refuse the
link exactly as it refused `dartsonline.com` at 11:47Z today.

Lane docs: `docs024_key_docs_latest/loanzy_uk_example_site/` (PLAN has the phases and the
compliance caveat about a finance-sounding domain).
