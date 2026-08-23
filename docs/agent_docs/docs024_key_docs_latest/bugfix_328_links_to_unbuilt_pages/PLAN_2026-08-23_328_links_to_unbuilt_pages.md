# PLAN 2026-08-23 — bug 328: stop shipping anchors to pages that have never been served

Design, phasing, decisions **and their reasons**. Corrections live here, marked, never edited away.

## The invariant

**No page is deployed carrying an in-body `<a>` whose target is a page of its own site that would
404 right now and is not arriving — while the authored href survives untouched in `content_data`,
so the anchor returns by itself on the first render after the target ships.**

What that makes unrepresentable: a framework-rendered page shipped with a link the database already
knew, at render time, was dead. What it deliberately leaves representable: a link to a sibling that
has simply not arrived *yet* — the in-flight window — because prevention there is the
no-internal-linking failure `bugs_open/313` (closed 2026-08-19) shows this platform can reach and
not notice.

## Decisions, and why

**1. Suppress on the OUTBOUND path only; never at persistence.** Removing the href from
`content_data` destroys the authoring intent and the link could never come back. Outbound-only
means the suppression is *recomputed every render*, so the whole thing is self-restoring with no
cascade, no repair queue and no second mechanism to keep in step. This is the single property that
makes the rest of the design safe, and it is enforced by a source scan, not by a comment.

**2. A new predicate, because neither existing one is right — and the gap is measured.**
`NeverDeployedPagePredicate` selects 9 pages serving 200; `PageMayBeLinkedPredicateFor` misses the
3 `needs_rebuild` rows that were never built (one of them this bug's own instance). The
discriminator is the rendered-component count: 20 with zero → 20/20 404, 9 with ≥1 → 9/9 200.
⚠ The conjunction is load-bearing: 8 pages have `deployed_at` set and zero components, and a
component test alone would delist all eight.

**3. Opt-in step-config field, default OFF** (owner ruling RFC_010 §2). One loader
(`loadValidPagePaths`) feeds four seams with different write surfaces; widening it would change all
four in one edit, including the one that writes `content_data`. So the new authority is a separate,
narrower set behind `suppress_unshipped_links`, enabled by a held-back migration after the roll.

**4. Two repair arms, decided by reading the markup rather than assuming.** 28 of 36 affected
anchors are classless prose (unlink, keep the words); 8 are classed controls with a label and an
arrow glyph (drop the whole element — LNK-005 correct-or-absent). **Owner's decision, 2026-08-23.**
A plain unlink would have multiplied a standing landmine by eight.

**5. The queue drain ships with it, the escalation does not.** Registering
`unbuilt_internal_link` in `reviewRevalidators` costs one delegation to the verifier that already
exists; teaching the item type a *second remedy* is a write on a failure path and needs its own
round. Deferred on record, not forgotten.

## Phasing

| phase | what | when it is live |
|---|---|---|
| 1 | Go: predicate, suppression pass, policy/loader, both outbound seams, the drain, tests | next fleet roll |
| 1b | `sql_for_agents/575_…_HOLD.sql` — the opt-in keys | **by hand, AFTER the roll** |
| 1c | `page_rerender` items for the 24 pages already serving dead anchors | after 1b |
| 2 | escalation: when the build remedy exhausts, re-render the referrers | deferred, filed |

## Corrections to my own framing, made during the work

> **CORRECTED 2026-08-23 — "63 open items = 14 days of live 404s" was wrong about 42 of the 63.**
> Those targets serve HTTP 200 today; the items are held open by missing `deployed_at` stamps
> (`bugs_open/315`), not by live damage. The real harm is the blast-radius measurement (36 anchors
> / 24 pages / 14 targets). A work item is not evidence about the wire. `WRONG_CALLS.md`.

> **CORRECTED 2026-08-23 — one outbound seam was not enough.** I had established that
> `page-build-handler.deploy_page` reaches `repairOutboundPageLinks` through the `page_renderer`
> role, and concluded one call site covered every publish path. `AssemblePageAction` is a second
> seam that calls neither repair function, used by `pageflow-builder`, `page-rebuild` and
> `site-work-orchestrator`. ⚠ It is invisible to a `jsonb_each` over `workflow->steps` — those
> three are nested in `sub_workflow`s and only a recursive walk finds them.

> **CORRECTED 2026-08-23 — two of my tests were not load-bearing until mutation showed it.** The
> default-OFF test was vacuous (the fail-open branch made the flag-ON mutant produce identical
> output) and the seam scan matched `name(`, so a value-reference walked past. Both rewritten and
> re-proven. See NOTES for the full table.

## Verification (the acceptance contract)

1. Pod-grep both replicas for `PageLinkRefusedPredicateFor` and `suppressUnshippedOutboundLinks`
   after the roll, with a control.
2. Apply 575 by hand; read the keys back (the migration's own `$post$` does this and RAISEs).
3. **At the artefact, not the plan**: re-render loanzy's home page and assert on the SERVED bytes
   that `href="/your-rights.html"` is gone **while `href="/calculators.html"` is still present**.
   The positive control in the same fetch is the whole test — without it, "stopped emitting
   internal links" passes.
4. The bug file's own acceptance test: a build with one page forced to fail, where no deployed page
   links to it and a page that did build is still linked.
