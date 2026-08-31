# CONTRIB from the loanzy lane, 2026-08-31 — OWNER ASK: carousels as a default pattern. READ THE OFFER-ANALYSIS MEASUREMENT BEFORE BUILDING.

The owner, reviewing farmerinsurance.uk (his words in substance): *"it would be nice to have
nicer components — using carousels rather than just lists and lists of separate cards. Maybe
we should make different types of carousels (with the same cards if we want to) as the
default, because scrolling down on a mobile with card after card is not a good user
experience."* Measured on farmer's homepage: 0 carousels, 72 card-class elements.

**Before building variants, read the sibling CONTRIB in this directory:**
`CONTRIB_2026-08-31_from_the_offer_analysis_lane_two_carousels_already_exist_and_the_planner_cannot_see_them.md`
(commit 97fcf0e22). Their measurement: two carousel components are ALREADY live in the
library (0 and 1 instances vs the plain grid's 42 across 21 sites), and `component_expresses`
carries **no traversal token** — the planner cannot distinguish a carousel from a grid, the
same defect shape as bugs_open/381 / IMG-074 (208-to-6). Their sequencing point stands:
**variants built without the vocabulary join the library as unchosen weight** — build and
vocabulary must land together. The vocabulary widening is fleet-visible planner input =
council scope, and the offer-analysis lane has put it to the owner with the measurement.

So the concrete asks for this lane, sequenced: (1) wait for the owner's word on the
vocabulary arm; (2) then carousel VARIANTS (same cards, different traversal treatments) as
the owner described; (3) the brief-fidelity seat's held verdicts on farmer ("Editorial hub
layout — featured guide(s) prominent" unmet) are the acceptance evidence sitting ready —
site `99cae989…`, queue query in loanzy_uk_example_site/RUNBOOK. Full review:
loanzy_uk_example_site/OWNER_REVIEW_2026-08-31_farmerinsurance_first_review_and_routing.md §5.

> **UPDATED same day, before commit — the offer-analysis lane's §6 (their commit `fa549fd76`)
> INVERTS the sequencing above, and theirs is the measured one. Read their §6 before my (1)-(3):**
> `info-card-grid` — the 42-instance grid the carousels lose to — **already has a declared
> boolean `carousel` schema field** gating a horizontal prev/next mode, ON for exactly 1 of 42
> (`case-studies-grid`: 0 of 4, same field from mig 559). So the owner's ask is closest to **a
> switch that exists and is off**: the flag is the cheapest lever, dedicated-carousel VARIANTS
> come later if at all, and the vocabulary token (derive from `semantic_tags`, 3/3 genuine —
> NOT a template grep, which mis-tags the grids themselves as carousels) is second. Two
> decisions stay the owner's: who sets the flag and on what evidence (`source: static` — nothing
> resolves it), and whether carousel-as-default is right at all (items behind an interaction
> are items many readers never see).
