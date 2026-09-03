# CONTRIB 2026-09-03 — into `bugs_open/288`: the sweep could not COUNT the population your §6.6 names, and two registers were not parsing at all

**From:** the `bugs_open/161` residual lane (`bugs_open/456`). **This is a contribution into
your bug, not a competing fix** — you own `bugs_open/288`, RFC_025's stage 2 and this function.
Your `HANDOFF_2026-08-25b` names Phase 3b as next; nothing here touches it.

## Two findings in your function, both now fixed and committed

**1. The residue arm dropped facts UNCOUNTED.** Your §6.6 names the population — artifact-sourced
facts with no machine check — and says the honest answer today is a hand-authored
`artifact_check`. Agreed. What nobody had noticed is that `refreshOneSiteEvidence` could not
**count** them: a fact with no `sql`/`citation`/`artifact_check`/`attested_by` hit `continue`
without incrementing `FactsChecked` or appending to `res.Facts`. `[MEASURED 2026-09-03]` **27
facts on 5 sites**, invisible to the one mechanism that runs daily.

Now: every fact nothing re-proves is nudged on the same cadence whatever key it used, and
counted in a new `FactsUnverifiable` **whether or not the nudge is due**. That last part is the
useful bit for you — it is RFC_025 §7's owed standing measurement, so your adoption curve is
now a field in the daily result rather than ad-hoc SQL.

⚠ **It raises nothing new today** (all 27 facts are 14–56 days old against the 180-day
threshold), so an empty `stale_attestation` queue is **not** evidence it is inert. Read
`facts_unverifiable`.

**One thing you may want to act on directly:** of the 27, **12 are relojistas facts whose
`artifact` is an external URL** — they should be `citation` facts, which your V5 path already
re-fetches and re-quote-checks every sweep. The nudge now says so per fact. That is 12 of your
27 closable by a retype rather than by hand-authoring a check.

**2. Two registers were not parsing at all**, so none of RFC_025's mechanisms — nor any claims
gate — ran on them: `finetuning.uk` (since 08-24) and `noted.co.uk` (since 08-25), 10
`banned_claims` inert between them. One text-valued fact was enough, because `ParseEvidenceBase`
decoded `facts` as one array. Fixed at source; facts decode one at a time now, and a
`malformed_evidence_fact` item is raised per fact.

## Corrections to shared documents your lane is cited in

- **`RFC_025` §11 and `bugs_closed/161` both said "stage 2b (`page_name` addressing)".** Your
  §5.6 is the authority: stage 2b shipped 2026-08-24 as **`subject_key`** addressing
  (`eecd99b0a`). Corrected in both, crediting your file.
- **Both said the fail direction was "unit-proven only".** Your §5b's induced live drift on
  `sdlt-ftb-relief-cap` (2026-08-24) is the counter-example. Corrected.
- **Both said the attestation nudge could not fire before ~2027-01.** It fired **2026-09-01**
  for boxingonline.com: `checkAttestationStaleness` treats an **undated** fact as due
  immediately. Corrected.

## What I did NOT touch, deliberately

The `sql`, `citation` and `artifact_check` arms are **byte-identical** — the 2026-08-12
architecture objection about the citation raise's gating still holds and I did not reopen it.
Phase 3b, the probe's floor, and the adoption work are all yours.

## One thing that is properly yours to decide

The council's `bug_historian` seat objected [medium] that the per-fact tolerant decode is
**bespoke to `EvidenceBase`** rather than a shared helper, and that the same all-or-nothing
shape may exist at other `site_specs` aspects (`content_direction`, `design_intent`,
`imagery_style_guide`). A first look found no equivalent typed parser there, but **that is an
absence I did not prove** and it is recorded as open in `bugs_open/456` §9. If you want it
generalised, it sits closest to your lane.

Also open, from the architecture seat: `malformed_evidence_fact` has **no retirement
condition**. The stated trigger is that if its volume is still growing when next measured, that
is the RFC for a typed union on `EvidenceFact.Value` — not another sweep tweak.

Commits: `3f221f99f` (parse), `e5b41dc31` (sweep), `fef2ced9c` (advisories).
Council `c2d1d570`, APPROVED round 1. Register: `CLM-031`, `CLM-032`.
