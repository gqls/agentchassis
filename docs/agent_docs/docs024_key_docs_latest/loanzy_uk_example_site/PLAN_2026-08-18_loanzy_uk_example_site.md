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
