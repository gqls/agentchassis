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

---

## CONTRIB 2026-08-22 — `bugfix_308_cta_destination_provenance` lane: re-verified, and two things this file does not say

Lane dir: `docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/`.
Opened as its own lane rather than inside `cta_target_content_pass` because that lane's PLAN
is a **content pass** (reword CTA labels so the existing resolver picks better) and carries
this change only as an unstarted phase-1 open question; the owner's 2026-08-18 ruling here is
platform Go. A CONTRIB goes into that lane's NOTES too — it is not competed with.

### 1. Still valid, and larger [MEASURED 2026-08-22, live]

This file's own query, re-run verbatim: **149 (2026-08-17) → 200**. Split by status, which is
the churn as a number:

| status | items | findings |
|---|---|---|
| `complete` | 71 | **112** |
| `unresolved` | 53 | 86 |
| `cancelled` | 2 | 2 |

**112 findings sit on work items the platform marked `complete`** — repairs that ran, reported
success, and left the button where it was.

**⚠ DEMAND CONTROL, and it matters for anyone verifying a fix here.** The detector has been
**silent since 2026-08-19** (3 items that day; 128 on 08-18, 208 on 08-17, 84 on 08-14). So
200 is a **stock, not a flow**, and re-running the census after a fix will return ~200 whether
the fix works or not. A discovery run must be induced first. [INFERRED, not verified] the
cause is `bugs_open/230` (site discovery has no recurring driver).

### 2. `suggested_target` has NO CONSUMER — the gap is one rung deeper than two universes

```
grep -rn "suggested_target" --include=*.go platform/ internal/ pkg/
```
→ three hits: `check_misdirected_cta.go:130`, `check_cta_nonpage.go:79-80`, and one test.
**All writers.** Nothing reads it.

The detector computes the right destination, writes it into the finding, and files a
`page_rerender` carrying only `spec.reason="cta_links_stale"`. `rerender_page_sections_action.go:528`
gates on that reason string and then **re-derives the destination independently**, from a
narrower candidate set. So the two halves do not merely disagree by accident — the half that
knows the answer is never asked. Sharing `ctaClassifyAnchor` (which this check's header says
was done precisely to stop this churn) shared the *classifier* and left both the *candidate
universe* and *the answer itself* unshared.

This is `bugs_open/071`'s shape on another seam (a gate that detects every broken link then
discards the finding), which is an argument for fixing the pattern, not just the pair.

### 3. A THIRD consumer of the candidate loaders, which this file and LNK-033's landmine both miss

`loadContentHubs` / `loadInteractivePages` have **three** non-test callers, not two:

| caller | use |
|---|---|
| `resolve_internal_links_action.go` | build-path resolution (`setCTAField`) |
| `rerender_page_sections_action.go` | repair-path recompute (`applyCTARecompute`) |
| **`render_site_components_action.go:182-190`** | **the site HEADER CTA fallback** (`chooseCTATargets`) |

`grep -c render_site_components` on this file and on `bugs_open/248` → **0** each. LNK-033's
landmine names three breakers of the invariant — widen the loaders, drop `candidatesFromHubs`'
filter, drop `rank()`'s test — as if they were interchangeable. **They are not.**

> **Widening at the LOADERS also silently re-picks every site's header button**, because the
> header derivation reads the same two functions and takes `ordered[0]` by nav_order. **So the
> widening this bug needs must happen at `candidatesFromHubs`, never at the loaders.**

And the obvious instrument is blind to it: [MEASURED] `site_components` carries **0 `cta_url`
keys across all 24 header rows** — the header CTA is never persisted, only rendered. A
`content_data` before/after diff (including the one the RFC_042/355 content-loss work is
building) reads clean while all 24 headers move. Only a rendered-chrome diff would see it.

**Not a falsification of LNK-033, and I checked:** the header legitimately takes `/contact.html`
from the nav's own contact item (`render_site_components_action.go:162-164`), which looks like
a resolver path that CAN mint a utility url — but chrome lives in `site_components`, not
`page_components`, and `site-header` is not in `ctaFieldNames`, so the predicate is never
consulted for it. The scope is exact; it is just only true because of a table boundary that is
nowhere written down.

---

## CONTRIB 2026-08-23 — PHASE A IS LIVE AND PROVEN. This bug STAYS OPEN.

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/`
(handoff: `HANDOFF_2026-08-22_continue_here.md`; milestone:
`SUMMARY_2026-08-23_cta_destination_provenance.md`).

**The owner-ruled candidate 1 — record the provenance — is built, council-reviewed over three
rounds, shipped (`288ce3e7a`), and verified WORKING at the artefact.** `__cta_minted` is a
value-bound record of which CTA url the resolver wrote, registered as **LNK-035**;
`storedCTADestinationIsAuthored` is re-based on it and its SIGNATURE changed so no caller can
reach the utility-area shape test without the mint check.

Measured 2026-08-23, against a baseline of **0** taken immediately after the 2026-08-22 roll:

| check | result |
|---|---|
| rows carrying `__cta_minted` | **11** (hero 6, call-to-action 5) — it MOVED |
| record entries naming the url their field carries | **21 of 21**, 0 mismatched |
| non-CTA components stamped (negative control) | **0** |
| rows with a `secondary_cta_url` that record it | **11 of 11** |

The last line is the live confirmation of a defect found during implementation: both persist
paths merge SHALLOWLY and the record is a nested map, so a naive version recorded one slot and
dropped the sibling's — freezing it. Four of the six `ctaFieldNames` components have two slots.

### Why this does NOT close 308

**Phase A makes the fix safe; it does not make the fix.** The candidate set is unchanged, so the
resolver still cannot offer `/contact.html`, and the 200 findings remain exactly as unrepairable
as when this bug was filed. This file's own verification bar #1 — *the CTA whose copy names
contact must actually reach the contact page, checked at the served page* — is untouched. **308
stays OPEN until Phase B.**

### One constraint Phase B must not lose (this lane's finding, not in the sections above)

`loadContentHubs`/`loadInteractivePages` have **three** non-test callers as of 2026-08-22, and the
third — `render_site_components_action.go:182-190`, the site HEADER's CTA fallback — is named in
no CTA bug file or register entry. **Widen at `candidatesFromHubs`, never at the loaders**, or
every site's header button silently re-picks; and `site_components` holds **0** `cta_url` keys
across all **24** header rows, so no `content_data` diff could ever have shown it.

### ⚠ Phase B is BLOCKED on the estate, not on design

The account's LLM budget is exhausted — the API states access returns **2026-09-01 00:00 UTC** —
and it is not confined to reviews: **112** failed steps as of 2026-08-23 09:49Z, including
`call_content_writer` (live site content generation). Zero LLM-producing orchestrations have
completed since. Council round 4 died unreviewed because of it. Phase B retires LNK-033's
invariant, and the only honest proof it worked is a real button on a real served page — which
needs the repair path running. **Do not start Phase B until the budget resets.**

---

## ⚠ CORRECTION 2026-08-23 (afternoon) — THE BUDGET BLOCK ABOVE IS LIFTED. Phase B is NOT blocked.

The section immediately above ends *"Do not start Phase B until the budget resets"*, and dates the
reset at **2026-09-01**. **That is now false, and a session reading it would stand down for nine
days for no reason** — the exact shape of the "a stale status line prevents the thing it
describes" trap.

**What actually happened** is in `memory/the-fleet-key-is-not-on-the-default-console-org.md`: the
cap was never a spend cap on the fleet's own account. The console the owner lands on by default is
**not the org the fleet's `ANTHROPIC_API_KEY` belongs to**, so the meter read `0% used` while the
API refused calls, and `2026-09-01` was the *wrong* account's reset date — a coincidence (monthly
limits reset on the 1st for everyone) that read as corroboration for two rounds. The owner
identified the correct account this morning.

**Measured at the fleet's own log, not at a console** [MEASURED 2026-08-23 12:25Z]:

```sql
SELECT max(created_at) FILTER (WHERE error_message LIKE '%usage limits%') AS last_cap_error,
       count(*) FILTER (WHERE success) AS ok_24h
FROM llm_call_log WHERE created_at > now() - interval '24 hours';
```

| | |
|---|---|
| last `usage limits` error | **2026-08-23 10:10:40Z** — none since |
| successful calls since | 15 (10:00Z hour) → 40 (11:00Z) → 79 (12:00Z) |
| failures in that window | 2, both `stop_reason=max_tokens` truncations — a different defect |

**Council round 4 has been resubmitted** on the same correlation
(`RESUBMIT_CORR=e4336931-487b-4db3-b4dc-a4b128b3566c`) and is in flight.

**308 still stays OPEN** — nothing here changes that. Phase A remains the whole of what shipped.

---

## CONTRIB 2026-08-23 — PHASE B IS COMMITTED (inert until the roll). 308 STAYS OPEN.

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/`.
Commit `7f85aa814`. Register: **LNK-036** (the shared universe) + **LNK-037** (ambiguity refusal).
Architecture note: `architecture_review/RFC_047_label_match_may_refuse_an_ambiguous_answer.md`.
Full evidence: `…/bugfix_308_cta_destination_provenance/CALIBRATION_2026-08-23_phase_b_widening_report.md`.

### What shipped

`datahelpers.LoadCTALabelUniverse` is now the ONE answer to "which pages may a CTA label name?",
consumed by the detector (`loadCTAMatchIndex`) and BOTH writers. `candidatesFromHubs` is **deleted**.
So the resolver can now mint `/contact.html`, and this file's fix candidate 1 is complete.

**The POSITIONAL pick is NOT widened** and must not be — `rank()` still refuses every utility area
(this file's verification bar #3), and the loaders keep their third consumer, the site HEADER CTA.

**Both writers' KEEP branches had to change, and this is the subtle half.** A MINTED utility
destination — a state that could not exist before this commit — took no keep at all once its label
went generic, and fell to the positional pick. That is `bugs_open/248`'s clobber arriving through
308's own fix. `applyCTARecompute` KEEP #2 lost its area test; `setCTAField` gained the matching
branch. The invariant that replaces LNK-033's:

> **The positional pick may neither CHOOSE nor DISPLACE a utility destination.** Only a confident
> label match puts one there or moves it away.

Verification bar #2 is satisfied: `TestFreshPickRefusesUtilityWhileStoredUtilityIsKept` is rewritten
(arm (b) inverted, arm (d) added), after provenance landed, not before.

### The measurement that changed the design — and it is bigger than this bug

Frozen fleet dump, real `datahelpers` imported, control of 1,266 pairs against the shipping matcher
with **0 disagreements**:

| | |
|---|---|
| fleet CTA writes today → after the widening | **32 → 428** |
| …after the ambiguity refusal | **291** |
| wide-pool matches decided by ALPHABETICAL ORDER alone | **263 of 1,146 (23%)** |
| …of which would overwrite a live CTA | **137** |

**Two families in that 137 are wrong and would have been executed fleet-wide**: finetuning.uk's
`"how we work"` moving OFF the `/how-we-work.html` its copy names (13 findings — the About page's
TITLE reads "…Who We Are and How We Work"), and dartsonline.com's `"Read the guides"` moving off
`/guides/index.html` (6). Both are one-token ties. So `BestLabelMatch` now returns `ambiguous` and
refuses an alphabetical-only win. **A name-tier key and a path-depth key were measured first and
both rejected** — third and fourth rejected tie-break keys across two calibrations.

**Two of this file's own standing suggestions are now answered by measurement:**

- **"recalibrate the stopwords" / add `about`** — DO NOT. It suppresses the four
  `Talk to us about …` → `/about.html` false matches AND the correct `Learn More About Us`.
- **fix candidate 2, the narrow widening** (label match only, utility pages added to the old pool)
  — measured and **WORSE**: 108 writes vs 291, and 26 utility "repairs" the wide pool does not make,
  because a pool that omits the label's real target gives the matcher a monopoly, not a choice.

### Why 308 STAYS OPEN — three reasons, none of them optional

1. **Nothing has touched a served page.** Go-only, inert until the next fleet roll. This file's
   verification bar #1 (the CTA whose copy names contact must actually reach the contact page,
   checked at the served page) is untouched.
2. **41 of the 188 findings can NEVER be closed by this change** [MEASURED 2026-08-23]: of 1,855
   `misdirected_cta` findings fleet-wide only **675 (36%)** sit on a `ctaFieldNames` component with
   the href as a `content_data` value; for this bug's 188, **147 (78%)** are repairable by the
   writers. The rest are anchors in prose or in components outside `ctaFieldNames` and need a
   different mechanism.
3. **A confident false match is not caught.** dartsonline's *"See how each brand differs, spec by
   spec"* → `/about.html` wins on identity overlap (`spec` is in the About page's title), not on a
   tie. No tie rule sees it and no stopword list can.

### A second axis of this bug's own defect, found while building the fix

The detector's universe had no build-state filter, so it could name a page the writers' `validPages`
gate then refuses — the same "suggests what the repairer cannot produce" shape. **43 of 764 live
pages are planned-and-never-deployed and 10 live findings named one** [MEASURED 2026-08-23].
`CTALabelUniverseSQL` carries the predicate now.

### Corrections to this file's own dated claims

- **"200 findings" → 188** (same predicate, 2026-08-23): `complete` 63 items/99 findings,
  `unresolved` 53/86, `cancelled` 2/2, `failed` 1/1. Nothing was fixed; items changed status.
- **"the detector has been silent since 2026-08-19" is FALSE** — 40 items were filed 2026-08-22. Do
  not carry the `[INFERRED]` attribution to `bugs_open/230` forward without re-checking.
- The `⚠ Phase B is BLOCKED` banner above is retracted; see the correction at the end of that
  section. The cap was on the wrong Anthropic account and lifted 2026-08-23 10:10:40Z.

### For whoever verifies this at the roll

```bash
kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "LoadCTALabelUniverse" /proc/1/exe   # + a control that must be ABSENT
```
then induce a discovery run and a `cta_links_stale` rerender on **finetuning.uk** (55 of the 188
findings) and load the page. `page_components` rows whose CTA url is a utility destination should
move from ~0; `misdirected_cta` items per day should FALL, because the two false families stop
being filed. Both directions matter — a fall alone could just mean the detector stopped running.
