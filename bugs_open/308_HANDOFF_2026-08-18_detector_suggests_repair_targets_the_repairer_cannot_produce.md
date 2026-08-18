# BUG 308 — the CTA detector and the CTA repairer disagree about which pages exist, so 149 findings name a repair no writer can perform and re-detect for ever

**Filed 2026-08-18** by the `bugfix_248_authored_cta_destinations` lane, which found this
while fixing `bugs_open/248` (slug `cta_recompute_clobbers_authored_contact_links`) and
deliberately did NOT fold it in — see "Why this is separate".
**Status: OPEN, not started.**

## Symptom

A CTA whose copy plainly names the contact page — "Contact our supply team", "Contact Us",
"Book a discovery call" — points at an unrelated tool. The `misdirected_cta` discovery check
sees it, agrees it is misdirected, and files a `cta_links_stale` repair naming
`/contact.html` as the suggested target. The repair runs, completes green, and changes
nothing. The next discovery pass files it again.

## Root cause — two candidate universes, one contract

The detector and the repairer answer "which pages exist as CTA destinations?" from
**different sources**, and nothing reconciles them:

| | source | includes contact/about? |
|---|---|---|
| detector (`loadCTAMatchIndex`, `check_misdirected_cta.go:314-354`) | **all** pages, any `page_type`, minus index/home | **yes** |
| repairer (`candidatesFromHubs` ← `loadContentHubs` + `loadInteractivePages`, `resolve_internal_links_action.go:589-612`) | `page_type='section-index'` + `page_type IN ('tool','game')` | **no** |

So `BestLabelMatch` inside the check can return a contact page as `suggested_target`, while
the same function inside `applyCTARecompute` cannot — the candidate list it is handed does
not contain one. The repair then falls to its keep-branch (the stored tool link is valid, so
it is kept) and the finding survives untouched.

This is precisely the churn the check's own header says it was designed to avoid:

> `ctaClassifyAnchor` is THE definition of "this CTA is misdirected", shared by the discovery
> check and the completion verifier so the two cannot drift. If they drifted, a verifier could
> resolve a defect the next discovery pass immediately re-detects — churn that reads as
> progress.

The *classifier* was shared. The *candidate universe it classifies against* was not.

## Measured (live, 2026-08-17)

```sql
SELECT count(*) FROM site_work_items swi, LATERAL jsonb_array_elements(swi.spec->'findings') f
WHERE swi.item_key LIKE 'misdirected_cta:%'
  AND f->>'suggested_target' ~ '^/(contact|about|privacy|terms|legal)(\.html|/|$)';
```
→ **149 findings**. The largest groups:

| copy | href | suggested | n |
|---|---|---|---|
| Contact our supply team | /tools/tool-breakeven-volume-calculator.html | /contact.html | 26 |
| Book a discovery call | /tools/password-entropy.html | /contact.html | 19 |
| how we work | /how-we-work.html | /about.html | 12 |
| See How It Works | /tools/archetype-taster-quiz/index.html | /about.html | 9 |

These are not false positives. "Contact our supply team" linking to a break-even calculator
is wrong, the detector is right, and the platform currently cannot fix it.

## Why this is separate from `bugs_open/248` and must land AFTER it

248 was the **destructive** direction — a repair overwriting authored contact links — and its
fix rests on a derived-provenance invariant: *no resolver path can produce a utility-area
destination, so a stored valid one is authored.* Closing THIS bug means letting labels
resolve INTO utility pages, which **falsifies that invariant directly**. So:

- **Do not simply widen `candidatesFromHubs`.** The moment the resolver can mint
  `/contact.html`, both of 248's keep-branches start freezing the resolver's own output, with
  no detector left to notice (248 demoted the excluded-area arm). See LNK-033's landmine.
- The prerequisite is **real recorded provenance** on the CTA url field — 248's fix candidate
  1, which 248 was able to avoid and this bug is not.
- There is a **measured false-match risk** in the widening itself: the 2026-08-08 calibration
  deliberately kept `about` OUT of `LabelStopwords` (*"over-narrowing the stopword list is how
  a detector goes quiet on what it exists to catch"*), so a label like "…about your use case"
  will match an About page the moment About becomes a candidate. That calibration needs
  redoing against the widened candidate set, not assumed.

## Fix candidates, ordered by what closes the door

1. **Record provenance on the CTA url field, then widen the candidate set.** Closes the class
   properly: the resolver may target contact, and a keep-branch can still tell its own output
   from an authored value because it no longer has to infer it. Largest change; it is 248's
   candidate 1 and this is the second bug to need it, which is itself the argument.
2. **Widen the candidate set for the LABEL MATCH only, leaving the positional pick excluded,
   plus a stopword recalibration.** Cheaper and probably what the `cta_target_content_pass`
   lane wants. But it breaks LNK-033's invariant unless 248's keep-branches are simultaneously
   re-based on something other than derived provenance — do not do one without the other.
3. **Narrow the DETECTOR to the repairer's universe** so it stops suggesting what cannot be
   built. Cheapest, and wrong: it would silence a true finding. "Contact our supply team"
   pointing at a calculator is a real defect on a live sales page; hiding it because we cannot
   currently repair it is how a detector goes inert.
4. **Route these findings to human review instead of to a repair handler.** Honest and
   available today, but the human queue has no working surface (`bugs_open/033`), so in
   practice it moves 149 items from a loop into a drawer.

## Owner / routing

`docs/agent_docs/docs024_key_docs_latest/cta_target_content_pass/` is owner-commissioned and
its PLAN already carries this as its phase-1 open question — *"whether to widen
`candidatesFromHubs` to guide pages"* — which is the same question one page wider. **Route
there rather than opening a competing lane.**

## How to verify a fix

1. The 149 findings' pages: after a repair pass, the CTA whose copy names contact must
   actually reach the contact page — checked at the served page, not at the work item status.
2. `bugs_open/248`'s tests must all still pass, and if candidate 2 was taken, the asymmetry
   pin (`TestFreshPickRefusesUtilityWhileStoredUtilityIsKept`) will have had to change —
   which is the signal that LNK-033's invariant was consciously retired rather than quietly
   broken. **If that test was edited without provenance landing, the fix is wrong.**
3. A fresh POSITIONAL pick must still never land on a utility page, whatever else changes.
