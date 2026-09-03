# CONTRIB 2026-09-03, from the 443 lane — apis.uk's own index page is in the subject-less-repeat exposure set

Small, actionable, no reply owed. Found while diagnosing the 443 build-detector's first
cohort (all pre-640 plans; full evidence `bugs_open/443` §10).

**The fact:** apis.uk's current site plan carries `generic-text-block` **×6 on `index`, with
zero subjects on any row** `[MEASURED 2026-09-03 — census, stale by addition]`. It is one of
only three sites fleet-wide whose current plan has a repeated component with no subjects
(the others: seotools ×4 pages, webdesign.co.uk `domains`).

**What that means for you:** every rebuild of your index through page-build-handler will
fire `REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT` ×6, and — after your own 641 lands — the
writer still cannot differentiate those six sections, because the differentiation source
(the subject) is absent from the plan rows, not from the prompt. 641 fixes the prompt half;
your plan rows are the data half.

**The cheap fix, whenever you next touch the site:** write six distinct subjects into the
`site_plan_sections` rows for `index` (phrasing rule: each completes "You'll want to know
___" — the owner's framing C, per your own 641 header), or replan the site post-640 and let
rule 17 do it. The 443 lane will list this in its item-4 remediation hand-offs after
Stage B regardless — this note just gets it to you first, since you own 640/641 and will
read the detector's index fires as noise otherwise.

Nothing here blocks or is blocked by the 641 owner-read gate.
