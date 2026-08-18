# PLAN 2026-08-17 — `bugs_open/248` (slug `cta_recompute_clobbers_authored_contact_links`)

> **Refer to this bug BY SLUG.** `248` is a documented ambiguous number — `bugs_closed/248`
> is the unrelated *undeployed-asset-repair* case, and most commit messages saying "248"
> mean that one. `git log` the FILE PATH, never the number.

## The defect in one line

One set — `areasExcludedFromCTA = {about, contact, privacy, terms, legal}` — makes three
different decisions, and only the first of them is a statement the set can support.

| # | Site | The decision | Verdict |
|---|---|---|---|
| 1 | `chooseCTATargets` → `rank()`, `resolve_internal_links_action.go:434` | a **freshly picked** CTA destination must never be a utility page | correct |
| 2 | `applyCTARecompute` keep-branch, `rerender_page_sections_action.go:771-776` | an **already-stored** link to a utility page is untrustworthy → overwrite | **the bug** |
| 3 | `check_misdirected_cta.go:182` | a deployed href in a utility area is an **unknown destination** → file for a human | false-positive engine |

Job 1 is about what the platform should *generate*. Jobs 2 and 3 reuse it as evidence of
what a human *meant*. It was never evidence for that.

> **CORRECTED 2026-08-18 (credit: the `bugs_open/299` session, via cross-session message;
> verified here at the live config and at 150 retained runs before propagating).** The
> paragraph below claims the build path was destroying links on regeneration. It was not:
> `page-content-writer` reads the resolver output from a path that resolves in **0 of 150**
> runs (`bugs_open/312`), so `setCTAField`'s output never reaches a page and its missing
> keep-branch was **inert**. The fix is a PRE-POSITIONED guard for the moment 312 unholds —
> which is a stronger reason to ship it, not a weaker one, but it is a different claim.

**Fourth consumer, and nobody had filed it:** `setCTAField`
(`resolve_internal_links_action.go:308-336`), the build/regeneration writer, has no
keep-branch at all. Fixing only the repair path leaves every authored contact link dying on
the next full page regeneration — which is exactly the re-check the leopardess lane recorded
as owed and unverified on 2026-08-14.

## The design decision, stated so it can be attacked

**Provenance derived from the code's own constraints, rather than recorded in a field.**

No resolver path can produce a utility-area destination:

- the **positional pick** cannot, because `rank()` drops those candidates (`:434`);
- the **label match** cannot, once `candidatesFromHubs` applies the same filter (Edit 5 —
  today it does not, and its doc comment falsely claims it does).

Therefore a stored url that **is a valid page** *and* **lands in a utility area** was not
written by either writer — it is authored. That is bug 248's fix candidate 1 (give the field
provenance) obtained without a schema change or a backfill.

`validPages` membership is load-bearing, not decoration: an **invalid** `/contact.html` is
bug 203's phantom, and replacing that is the repair, not the bug.

## Decisions and their reasons

- **Both writers, not just the reported one.** Scope confirmed with the owner 2026-08-17.
  The bug file's own recommended fix (its candidate 2) touches only `applyCTARecompute`;
  that leaves the build path destroying the same links.
- **The detector arm is DEMOTED, not deleted.** Keep emitting the finding, stop emitting the
  `cta_names_unknown_destination` work item. Deleting outright would remove the only signal
  that can see a fabricated-but-*valid* contact link (the phantom arm is blind when the page
  exists; the misdirect arm is blind when the copy names nothing). Demoting keeps the signal
  and stops the queue noise. **This is more conservative than the option approved** — the
  approved option was deletion, and the adversarial pass argued deletion away.
- **A "keep" WRITES; it does not merely return.** On regeneration whether a bare return
  preserves the value depends on `plan_sections`' carry, which misses for non-`deployed`
  rows, conflicted duplicate slots and mismatched slot names. A bare return can therefore
  *drop* the button it was added to protect.
- **The label-match branch stays AHEAD of the keep in both writers.** This is forced by the
  bug file's own verification bar #2: a fabricated contact url whose label names a real page
  must still be recomputed. It is also the source of a known residual (below).

## Known residuals — recorded, not fixed

- **Label overlap still beats an authored contact link.** `BestLabelMatch` claims a label at
  overlap ≥ 1, so "Talk to us about pricing" → a Pricing page wins over an authored
  `/contact.html`. Forced by the ordering above. Test documents it; not closed by this fix.
- The self-link gap and the both-fields-same-target quirk (the bug file's 2026-08-15 sibling
  observations) are untouched.
- **The 149 unrepairable findings get their own bug file** — the detector's match index
  holds every page (contact included) while the repairer's candidate set holds only
  section-index + tool/game, so a finding suggesting `/contact.html` names a repair no writer
  can perform, and re-detects for ever. Routed to the owner-commissioned
  `cta_target_content_pass` lane, whose PLAN already holds this as its phase-1 open question.
  Deliberately NOT folded in here: it *adds* a capability with a measured false-match risk
  ("about" is deliberately not a stopword, per the 2026-08-08 calibration), whereas this
  commit is purely protective.

## Corrections to the originating brief

> **CORRECTED 2026-08-17 — the bug file's containment argument has expired.** It states the
> discovery and improvement schedulers are disabled, so the defect is "mostly dormant".
> Measured live today: `site-discovery-rotation-completeness` (the check's host) is
> **enabled**, `detected-item-promoter` runs every **900s** with no human in the loop, and
> `misdirected_cta`-keyed repairs dispatched **today**. `016b` §9 rule 4 predicted exactly
> this: a defect gated by a switch someone else can flip is not contained.

> **CORRECTED 2026-08-17 — the queue figures in the bug file cannot be re-run as written.**
> It reports "192 detected / 95 unresolved / 63 failed" for a `misdirected_cta` queue.
> `item_type='misdirected_cta'` returns **zero rows**: the check files `item_type
> ='page_rerender'` with `item_key LIKE 'misdirected_cta:%'` and `spec.reason
> ='cta_links_stale'`. Query by item_key.
