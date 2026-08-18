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

---

## OWNER RULINGS 2026-08-18 — this bug's direction is now decided

Recorded by the `bugfix_248_authored_cta_destinations` lane, which filed this bug and put the
question to the owner in the same session.

### 1. **Build the provenance record. Fix candidate 1 is the chosen route.**

Asked whether to keep deriving provenance from the resolver's constraints (cheap, fragile) or
to record it properly, the owner ruled: **"Keep a provenance record."**

So candidates 2, 3 and 4 above are **not** the plan. In particular candidate 2 (widen the
label-match candidate set and recalibrate stopwords) must NOT ship on its own — it falsifies
`bugs_open/248`'s live invariant (LNK-033) the instant it lands, and 248's two keep-branches
would begin freezing the resolver's own output with the detector arm that would have noticed
now demoted. **Provenance first, then the widening.**

### 2. **No opt-out flags. "Don't add any new flags that let other agents ignore things — this leads to bugs."**

A design constraint on whatever is built here, and it rules out the shape this estate reaches
for by habit: a `skip_provenance` / `ignore_authored` / `force_recompute` config key on the
step, so a caller in a hurry can switch the new rule off.

Read it against the 2026-08-02 owner ruling (new authority on a shared seam ships as an opt-in
field with the unsafe default OFF) — **they are not in conflict, and the difference is worth
being precise about.** That ruling is about a *capability* whose widest branch needs a caller
to opt in. This one is about an *escape hatch* from a rule that is already load-bearing. A
field that turns a protection OFF is not the same object as a field that turns a capability
ON, even though both are "a new key with a default". If the provenance work needs a new key,
it must be the second kind.

The 248 fix already conforms, and is a usable precedent: `storedCTADestinationIsAuthored` is a
predicate with no config surface at all, and the detector demotion is unconditional. Nothing
was added that a caller can switch off.

### What this means for sequencing

`bugs_open/312` (the wiring defect that makes `setCTAField`'s output discarded) is held and
gated on 248's two keep halves. 248's rerender half is proven live at `v1.0.1310`; its build
half is pre-positioned. Provenance work should assume the build path WILL be live and design
for both writers from the start, rather than treating the build path as dormant.
