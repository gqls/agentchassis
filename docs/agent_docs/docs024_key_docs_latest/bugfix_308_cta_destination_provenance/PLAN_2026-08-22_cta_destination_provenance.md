# PLAN — bugs_open/308: a recorded CTA destination provenance stamp, then one shared candidate universe

**Opened 2026-08-22.** Design drafted by a `fable` planning agent against a grounded brief;
every load-bearing claim re-checked here by hand, and two of its conclusions corrected (§7).

## 1. What this is, in plain terms

When the platform decides whether it may rewrite a stored CTA link, it answers *"did a person
choose this link, or did we?"* by **inference**: "we are incapable of producing a contact-page
link, so a stored contact-page link must be a person's." That is LNK-033, and it is why
`bugs_open/308` cannot be fixed — the moment we let the machinery target contact pages (which
is the fix), the inference becomes false and every keep-branch starts freezing our own output.

This plan replaces the inference with a **record**. When the resolver writes a CTA destination
it writes down, beside the value, *which url it wrote*:

```json
"__cta_minted": { "cta_url": "/contact.html" }
```

A stored link is then authored when it is a valid page **and the record does not name its
current value**.

**The rule this must satisfy** (owner, in the bug file, 2026-08-18): build the real provenance
record — candidate 1; provenance first, then the widening; and **no opt-out flags**, nothing
that lets another agent switch a protection off.

**How it measures against that rule.** The stamp is a record, not a flag: no config key is
added anywhere and there is nothing to switch off. The two behaviour-changing sites are the two
resolver writers themselves. The widening ships only after the stamp is pod-verified live, in a
separate commit that amends LNK-033 visibly.

## 2. The one decision that makes the rest cheap: the stamp is VALUE-BOUND

It records *which url* was minted, per field — not a bare "the resolver wrote this".

`authored == valid && (no stamp for this field || stamp names a DIFFERENT value)`

This is what makes eight of RFC_042's nine `content_data` writers correct with **zero code
changes**, and it is not a detail — a boolean stamp gets the killer case backwards:

- a **REPLACE** writer drops value and stamp together → authored. Correct: it *is* authoring.
- a **MERGE** writer (section-editor field update) writes a new value while the old stamp
  survives → the stamp **mismatches** → authored. **Under a boolean stamp the surviving stamp
  would licence the recompute to clobber a human's edit** — i.e. `bugs_open/248` again.
- a path that carries the **value without the stamp** → false *authored* → a freeze. The **safe
  direction**: the frozen value is a valid page, and the label-match branch still runs ahead of
  every keep in both writers, so a frozen link whose copy names another page is still repaired.

## 3. Where it is stored, and why that is safe

Inside `page_components.content_data`, one reserved key per section. No schema change
(`content_data` is jsonb), no migration for the mechanism.

**The load-bearing assumption, DISCHARGED by measurement rather than left as a risk.** The
design dies if an undeclared key cannot survive into stored `content_data`. It survives:

```sql
SELECT cc.function, count(*) AS rows_storing_target_title,
       bool_or(cc.input_schema::text LIKE '%_target_title%') AS schema_declares_it
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.content_data::text LIKE '%_target_title%' GROUP BY 1;
```

| function | rows | schema declares it |
|---|---|---|
| hero | 223 | **t** |
| call-to-action | 215 | **t** |
| content-block-about | 12 | **f** |
| gauntlet-cta | 2 | **f** |
| archetype-combinations | 1 | **f** |
| archetype-grid | 1 | **f** |

**16 rows across four component functions store a resolver-written key their schema does not
declare.** [MEASURED 2026-08-22] The check could have come out otherwise — had all 844 rows sat
on the two declaring components it would have proven nothing, which is why it is grouped by
function and not counted in total.

The `__` namespace is measured clean fleet-wide (0 `__`-prefixed keys in any `content_data`).

**Co-ordination with the RFC_042/355 content-loss lane, not competition:** the stamp is
deliberately **not** schema-declared, so `cmd/content-loss-check`'s differ (scoped to
schema-declared non-LLM keys) ignores it and stamp churn can never read as key loss. A CONTRIB
goes into their NOTES naming the reserved key and asking that `__`-prefixed keys stay out of any
future widening of that differ.

**Backfill: none, deliberately.** Absence of a stamp preserves today's exact semantics (the
derived utility-area rule keeps applying to unstamped rows) — the bug file's own "default to
derived preserves today's behaviour" arm. Defaulting to authored would freeze every CTA in the
fleet, not just the 200 findings.

## 4. Phasing

**Phase A — the stamp and the re-based predicate.** One commit, one council round,
behaviour-preserving.

- New `platform/orchestration/datahelpers/cta_provenance.go`: `CTAMintedKey`, `SetCTAMinted`,
  `CTAMintedCovers` (normalised compare). Pure functions, no DB, no config.
- `setCTAField` stamps on **every branch that writes a url**, and — see §7 — re-stamps at the
  **unresolved fallthrough**, where a carried value would otherwise ship unstamped.
- `storedCTADestinationIsAuthored` re-signed to take `(stored map[string]interface{}, field
  string, validPages)` and conjoin `!CTAMintedCovers(...)`, **so no caller can use the shape
  test without the mint check** — the signature is the enforcement.
- `applyCTARecompute` stamps on its mint branches; **KEEP #2/#3 and their ordering untouched**
  (LNK-034's mutation-proven seam).
- Inert until rebuild+roll. Behaviour-identical: with zero stamps in the DB and the narrow
  candidate set, the resolver never mints utility, so the new conjunct never fires.
- **Gate before Phase B:** on a canary, induce one `cta_links_stale` rerender and one full
  build (the build writer is live since 312's `e10a2c887`), then
  `SELECT count(*) FROM page_components WHERE content_data ? '__cta_minted';` 0→N, with a
  negative control (non-CTA components stay stamp-free).

**Phase B — one shared universe, the widening, recalibration, the completion verifier.** Gated
on A pod-verified; the calibration report is the shipping gate.

- New `datahelpers/cta_universe.go`: `LoadCTALabelUniverse` — one SQL spelling for both sides,
  and it fixes the detector's own gap in passing (it lacks `PageMayBeLinkedPredicateFor`, so it
  can suggest a never-built 404). `ctaClassifyAnchor` moves to `datahelpers.ClassifyCTAAnchor`.
- **`candidatesFromHubs` is deleted — its utility filter IS the invariant being retired.**
- **The positional pick (`chooseCTATargets`, `loadContentHubs`, `loadInteractivePages`) is
  UNTOUCHED.** This is not tidiness: those loaders are shared with the **site header CTA
  fallback** (`render_site_components_action.go:182-190`, this lane's own finding). Widening at
  the loaders silently re-picks every site's header button, and no `content_data` diff can see
  it because `site_components` stores no `cta_url` on any of its 24 header rows. **Widen at the
  assembly seam; never at the loaders.** It also keeps verification bar #3 intact.
- **Recalibration with evidence, before commit.** A `fleetharness`-tagged harness runs
  `ClassifyCTAAnchor` over fleet anchors with OLD vs NEW universe and reports every delta.
  It must explain the known case: "how we work" → `/about.html` (n=12) is an alphabetical
  **tie** loss today. Candidate rule, shipping only if the harness supports it: *a utility-area
  candidate wins only on strictly greater identity overlap, never on a tie* — kills the
  how-we-work class, keeps "Contact our supply team" (contact 1 vs tool 0, strict). Any rule
  ships **inside the shared matcher**, so both sides move in lockstep; no config key. Output:
  `CALIBRATION_<date>_cta_widened_universe_report.md` in this dir.
- **Completion verifier** (separable commit): `VerifyMisdirectedCTAResolved` — re-runs the
  detector's own predicate against live rows before a `page_rerender` may complete. **This is
  what turns the 112 "complete and unchanged" outcomes into refusals.** Spec-only fast path for
  every other reason (it fires on the fleet's busiest pipeline), with its own mutation proof.
- **Lockstep source-scan test**: the three CTA files may obtain label candidates only via the
  shared loader.
- Same commit as the widening: the LNK-033 amendment (§6) — condition (2) of the ordering
  exemption.

**Phase C — drain the stock, verify at the artefact.** §8. No Go.

## 5. Migrations

Highest today is **554**; numbers race between lanes, so re-derive at write time.

- **None for the mechanism** — jsonb content, two Go writers, no config keys, no checks array.
- **`555_requeue_misdirected_cta_stock.sql`** (+ `_ROLLBACK`), Phase C **only after the Phase B
  image is stamp-verified per service**: a status flip is live instantly, and flipping under the
  old binary re-runs the broken repair and adds strikes. Flips the ~53 `unresolved` items back
  to dispatchable; leaves the 71 `complete` alone (the induced discovery re-run covers them).
  Not `_HOLD` — nothing orders against another migration; the ordering constraint is against the
  *image*, enforced by hand-applying post-verification and stated in the header.
  [UNVERIFIED] that `unresolved` is non-dispatchable and `detected` is — read the dispatch
  loop's status predicate before writing it.

## 6. Register and who must be told

- **LNK-033 AMENDED, visibly and dated:** the invariant *"no resolver path can produce a
  utility-area destination"* is **RETIRED**, owner-ruled; provenance is now RECORDED, not
  derived; the landmine paragraph is superseded for the label-match route and **stands for the
  positional route and schema fallbacks**. The asymmetry-pin paragraph is rewritten.
- **New LNK-035** — the value-bound `__cta_minted` stamp, with the freeze residual as its
  landmine and the "REPLACE drops it by design" note so no future census reports it as loss.
- **New LNK-036** — `LoadCTALabelUniverse` + `ClassifyCTAAnchor` + the lockstep test + the
  verifier; carries the "chrome may not use this" note. (Confirm free ids at commit time —
  LNK-031/032 have collided once.)
- **Consumers to be TOLD, enumerated by query** (this lane's NOTES, re-run at commit): the
  loaders' three callers incl. `render_site_components_action.go` (chrome; untouched, told
  anyway); `check_cta_nonpage.go` (shares the classifier — its calibrated scope must be
  re-checked); the **`cta_target_content_pass`** lane (told already, 2026-08-22); the
  **248/299** lanes (their keeps re-based, their tests are the tripwires); the **238/355** lane
  (reserved key); and `016b §9` for the transferable pattern.

## 7. Where I corrected the drafted plan

1. **The stamp fix belongs in `setCTAField`, not in PBP-039's carry.** My own earlier NOTES
   entry said the carry must carry the stamp; the planning agent said the carry structurally
   cannot and must not. **Both were wrong, and the call site settles it** —
   `setCTAField` is handed `existing`, a *fresh DB read*, not the carry's output, so the stamp
   is in its hand directly. The leak is `setCTAField`'s own final fallthrough, which writes
   nothing. Corrected in NOTES and `WRONG_CALLS.md`.
2. **The plan's headline `[UNVERIFIED]` — that `__`-prefixed keys persist — is discharged by
   measurement** (§3), not carried as a risk. It was the assumption the whole design rests on.

## 8. Verification, at the artefact

**Nothing counted before a discovery run is induced.** The detector has been silent since
2026-08-19 (a stock, not a flow), so the census returns ~200 whether the fix works or not.

1. After Phase B rolls and 555 applies: **curl the served pages** for the four modal finding
   groups (26 + 19 + 12 + 9). The anchor reading "Contact our supply team" must carry
   `href="/contact.html"`.
2. **Controls in the same breath:** a 248-protected authored contact button unchanged;
   webdesign.uk's faq `tel:` button unchanged (KEEP #3); and on a canary, no CTA whose label
   names no utility page gains a utility href (bar #3).
3. Induce a discovery run: re-detection must file **0** for the repaired anchors — *and the run
   must be shown to have executed*, via findings > 0 on a known-dirty control page. A zero from
   a run that did not happen is this lane's own demand-control lesson.
4. The verifier's first refusals are the honest failure surface: they must appear as refused
   completions, not as `complete`.

## 9. Not verified, stated

- [UNVERIFIED] dispatchability of `unresolved` vs `detected` (gates 555's form).
- [UNVERIFIED] which discovery agents carry `misdirected_cta`, and the induced-run command.
- [UNVERIFIED] the exact function folding `resolved_data` into persisted `content_data` — §3's
  measurement proves the *outcome* (undeclared keys arrive) without my having read the writer.
- [INFERRED] `bugs_open/230` as the cause of the detector's quiet. The design does not depend
  on it; the verification plan does, which is why §8 induces a run rather than waiting.
- [ASSUMED] no template ranges over `content_data` keys, so the stamp cannot render. Cheap
  check: fleet grep of `rendered_html` for `__cta_minted`, expecting 0, after Phase A.
- Migration number and register ids race with concurrent lanes; re-derive at commit time.
