# SUMMARY 2026-08-25 — bugfix_380_claims_fail_open (milestone: shipped, rolled, verified, closed)

## What we are trying to do

Stop the framework inventing what a business *does*. On a site with no verified facts on record, the
three mechanisms that were meant to keep the copywriter honest all switched themselves off — the
planner stopped assigning facts, the writer was shown nothing, and the claims auditor exited without
reading a page and reported success. The sites that knew least were checked least. The owner's ruling:
aspirations must never be written as present-tense practice; say only what is sourced, and source more.

## Where we have come from

`bugs_open/380` was filed on 2026-08-24 from the garden-tools.uk greenfield build, whose largest page
described a product-testing operation that had never tested a product. The same inversion had been
diagnosed on 2026-07-20 by the claims_verification lane, with a fix designed ("cold audit") and never
built. When this lane took the bug the auditor turned out to be an orphan — no seed, no schedule, one
LLM call in its life — and the bug's own "mint an empty register" candidate could not have worked,
because every consumer keys on the facts, not the row.

## What we have done

- **Auditor (597, 601):** the opt-in gate is deleted; with no facts the audit runs cold (every business
  assertion unsupported, first-person practice claims reported first, honest disclosures spared); a DB
  error fails the run instead of completing it; every run leaves a receipt; and the page-text extraction
  — which a PostgreSQL regex quirk had been silently gutting — strips per component.
- **Planner (598):** factless sites get the object form with `facts: []` on every section, which turns
  on the writer's strong arm with no Go change; methodology pages may not be briefed as practice.
- **Writer (599):** a no-register / no-operating-history arm, released after the owner read the plaintext.
- **Clock (600):** one site an hour, every site every seven days, never-audited first.
- **Go (practice-claims family):** a deterministic detector for "we test / buy / weigh / receive samples",
  exempted by a signed `operating_history` attestation, recorded at warning through its own entry point
  and pinned out of the refusing set; attestation-only registers can no longer arm the number scan.
- Both council slices APPROVED round 1. Seven landmines, three wrong-calls, three register entries, nine
  lane notes, RFC_003's owner answers.

## Where we are now

Everything is live and verified at the running system (2026-08-25): chassis `v1.0.1337` carries the Go
commit (binary probe with controls, 1/1/0); the rotation audited 15 sites overnight with a receipt for
each (10 with findings, 4 clean); the writer arm rendered on 150 of 150 content calls; the cold audit's
first two findings on garden-tools were the owner's own quoted sentences. The bug is CLOSED and moved to
`bugs_closed/`. One verify-later remains: the first rebuild of a register-less page that carries a
practice sentence should show a `practice_claim` warning in its validation result — the three builds
since the roll had nothing to flag.

## Where we are going

Nothing further on this bug. What it uncovered belongs elsewhere: `bugs_open/386` (a refreshed fact
convicts pages still rendering the old value — the rotation now amplifies it), `bugs_open/033` (the
human-review queue the findings feed has no working surface), six other agents that complete as success
on a missing target (census handed to the 354 lane), four other queries carrying the same regex trap
(landmined), wiring real research into greenfield builds ("source more" — needs the owner's decision
because unattended research breaks agritec's mandatory review), and a shared visible-text SQL definition.
