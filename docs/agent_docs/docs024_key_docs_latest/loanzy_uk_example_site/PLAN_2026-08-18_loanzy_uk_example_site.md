# PLAN 2026-08-18 — loanzy.uk as webdesign.uk's first one-shot example site

## What this lane is for

Build **one site on `loanzy.uk` using nothing but a webdesign.uk customer prompt**, so that
webdesign.uk can finally show a prospect the pair it currently cannot: *this prompt produced
this site*.

## Why it exists (the constraint it lifts)

webdesign.uk's page lead was approved 2026-08-18 — proposal **F, "show the work, promise
nothing"**: *you can see exactly what you get: real sites built with this system, and the
exact prompt that produced each one*. Its **precondition is the load-bearing half**: the
`writer_block` forbids "see our examples", any portfolio link, or implying a showcase
**until a gallery exists**, because the four sites the copy may name
(`any_site_type_examples`: noted.co.uk, cookly.uk, dartsonline.com, vetcomparison.uk) were
**not built by the one-shot route**. That is a copy-honesty judgement, not a technical
blocker — the `cross_site_domain` guard refusing the links is a separate, smaller thing
(and an allow-list already exists for it: `loadAllowedReferenceDomains`,
`validate_page_content.go:1462`).

So the missing artefact is a site that the one-shot route really did produce. This lane
makes it.

## The owner's decision, in his words (2026-08-18 chat, this session)

1. *"We can put loanzy back in the queue with the rest of the finance domains."* —
   opening position, **superseded** by what follows.
2. *"yes we could use it as an example site"*
3. *"which would mean no prior registry entry"*
4. *"and have it built only from the webdesign.uk prompt"*

Recorded as **P10** in `portfolio_positioning/REGISTER_positioning.md` (L9), which also
records what P10 does **not** do: `loanzy.uk` keeps its L9 claims-table line, but that claim
is mechanical (it stops another proposition drifting onto the domain and keeps
`check_register.py` whole) and is **not** a direction.

## The rule this lane is built on, stated so it can be broken visibly

> **Every input to the build must be traceable to a sentence in the customer prompt.**

A paying customer supplies a prompt and nothing else. They do not supply a positioning
entry, a hand-tuned `evidence_base`, an imagery style guide written by us, or a second
round of steering. If we give `loanzy.uk` any of those, the gallery item overstates the
product — which is the exact dishonesty the owner's own deferral was protecting against.
Where the prompt attests nothing, the seed carries **nothing**: `facts[]` empty plus a
`writer_block` forbidding figures (the `oufe` precedent; webdesign.uk's own seed could
populate facts only because the owner had attested them).

**Anything we add beyond the prompt gets written into NOTES as a deviation, in the same
session it happens.** A silent deviation makes the published pair a lie.

## Phases

**Phase 1 — the prompt (owner input; BLOCKING).** One customer-shaped prompt, in a
customer's register — no framework vocabulary, no register jargon. It is a published
artefact: it appears next to the site in the gallery. Candidates and the compliance
constraint are in §Open questions.

**Phase 2 — build.** Seed from the prompt alone, then
`082_submit_domain_unified.sh loanzy.uk --email … --mission-file …`. Standard gated
pipeline, no interventions beyond what a customer's build would get. Every HITL item the
build raises is answered the way a customer's build would be answered — or left, and
recorded as left.

**Phase 3 — serve it.** `loanzy.uk` needs no DNS work: the zone is active, both worker
routes exist, apex and www already answer (see RUNBOOK). This is the one domain in the
estate where the lane's open "who owns DNS" question does not bite. Deploy, then verify at
the served page, never at item status.

**Phase 4 — hand the pair to the webdesign lane.** Prompt + URL + what it cost + what it
got wrong. Their lead's precondition is a **gallery**, so one pair is the first row of one,
not the whole thing; the copy stays as it is until they decide the gallery is live.

## What would make this lane's output worthless

- Steering the build after dispatch and not saying so.
- Publishing the pair while quietly omitting the HITL answers we supplied.
- A site that reads as a regulated financial firm (see §Open questions) — a compliance
  problem on a live UK domain, on the lane whose product is other people's trust.

## Open questions

1. **The prompt itself** — owner's call, Phase 1 blocking.
2. **Subject matter vs the domain name.** `loanzy.uk` sounds like a lender. A demo brand
   that reads as a **lender or broker** is a fake regulated firm on a live UK domain, and
   the register's own L10 entry is emphatic about the line (NOT a lender, NOT a broker, no
   applications, no lead-gen). Two clean ways out: pick a subject that is not
   regulated-financial at all — which also demonstrates the copy's new *"any sort of site"*
   claim — or make the demo status visible on the page. Note the new commercial position
   makes the mismatch *realistic*: customers are hosted under a domain we provide, so an
   example site living on one of our spare domains is what a real customer's site looks
   like before they rent or buy one.
3. **Does the pair need the cross_site_domain allow-list?** Only when webdesign.uk copy
   links to `loanzy.uk`. `loadAllowedReferenceDomains` is opt-in per site and in production
   use on `fundamentallyai.com` — so it is a config change, not a build.

---

## REVISION 2026-08-18 (afternoon) — Phase 1 was struck, then reinstated by evidence

**Phase 1 ("the prompt", owner input, blocking) was struck** when the owner asked for the
framework to determine the direction itself with no prompt at all. That run happened, and
its result **reinstates Phase 1 on evidence rather than preference**: given only the string
`loanzy.uk`, the classifier produced a UK credit-broking business — lender panel, eligibility
checker, per-referral revenue — and one page of it deployed to the live domain before the
build was stopped. Full account: `SUMMARY_2026-08-18_the_no_prompt_build_put_a_credit_broker_live.md`.

**The owner has chosen to re-run with one short prompt** naming a business that is not in a
regulated trade. Still one customer input, still no positioning entry, still the framework
doing everything else.

### The phases now

**Phase 0 — containment BEFORE dispatch (new, and non-negotiable).** The next build does not
start until publication is something we control rather than something we race. Concretely:
either the domain has no worker route while the build runs, or the build targets a domain
that does not resolve, and either way the hold is verified at the URL before dispatch. Today's
sequence — dispatch, discover, then try to stop a queue that had already handed work out —
is what put a page on the internet. **Also: `page-build-handler` deploys each page itself
(`deploy_page`); there is no single publication step to hold, and cancelling queued items does
nothing to an item already CLAIMED.**

**Phase 1 — one short prompt (owner).** Non-regulated subject. It is published beside the
site, so it is written as a customer would write it.

**Phase 2 — build, unchanged in principle**: seed from the prompt alone, dispatch, no
steering, deviations logged.

**Phase 3 — clean up today's run.** `bugs_open/304`: the retracted page is still in the
bucket because `Deploy to B2` skips a domain whose directory no longer exists. Needs the
owner or a permission grant — see NOTES for the two commands.

**Phase 4 — hand the pair to the webdesign lane** (unchanged).

### Shipped since the original plan

- **CGV-032 / migration `464`** — the classifier may not propose a regulated business model
  unless a mission explicitly asks for one (owner instruction). Applied and verified live in
  config; **unexercised** until a finance-shaped domain is submitted.
- **`bugs_open/304`** — retracting a site's last page cannot unpublish it.
- **A LANDMINE correction** (the b2 remedy is available on this box, and an agent may still be
  blocked from running it) and a new LANDMINE (cancelled ≠ stopped; deploy lives in the page
  builder; retraction refuses a live page).

### Recorded, not acted on

**The council gate structurally cannot review this change.** `097_TRIGGER`'s client-side
scope is `^(platform|internal|pkg)/`, so a seam that ships as DB config — which is where a
large share of fleet behaviour actually lives — is refused before it costs anything. CGV-032
is therefore registered but unreviewed, which is the honest state, not an oversight. Whether
config-only seams should be reachable by the gate is a question for the architecture lane.
