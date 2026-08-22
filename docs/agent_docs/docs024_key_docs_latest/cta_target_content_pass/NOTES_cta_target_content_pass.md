# NOTES — cta_target_content_pass (append-only, newest at the bottom)

## 2026-08-15 — lane created (spun out of bugfix_268_cta_buttons_fleet)

- Owner commission received in the 268 lane's session 3; wording and
  context in PLAN. Population measured same day (RUNBOOK query): 16 sites
  with ≥6 rows on their modal target; baseline table in the 268 lane's
  session log and the PLAN.
- Nothing dispatched yet. Phase 1 = finetuning.uk canary per the RUNBOOK
  recipe. Caveats inherited from 268 are listed in the RUNBOOK — read them
  before dispatching, especially the label-match tie class (253) and the
  248 quirks.
---

## CONTRIB 2026-08-18 from `bugfix_248_authored_cta_destinations` — your CTA guarantee changed, commit `53a8d3c1d`

Telling you rather than only measuring that nothing broke (the 2026-07-29 owner ruling: a
shared mechanism's other consumers must be told). **Nothing in this lane's files was edited.**

**What changed for you.** `areasExcludedFromCTA` ({about, contact, privacy, terms, legal}) was
answering three questions with one answer. It now answers only the first:

1. a FRESH positional pick still never lands on a utility page — **unchanged**;
2. an ALREADY-STORED valid utility destination is now **KEPT** by both writers
   (`applyCTARecompute` and, newly, `setCTAField`), via `storedCTADestinationIsAuthored`;
3. the `misdirected_cta` check's "lands in an excluded area" arm still emits its finding but
   **no longer files a `cta_names_unknown_destination` work item**.

**The label-match branch is untouched and still runs first**, so a stored contact url whose
label names a real page is still recomputed — 248's verification bar #2, pinned by tests on
both writers.

**One thing you may have believed that was never true:** `candidatesFromHubs`' doc comment
claimed its inputs arrive pre-filtered by `chooseCTATargets`' `rank()`. `rank()` filters a
local copy and never mutates its inputs, and both call sites passed the raw loader output.
It now really does filter, which is what makes the derived-provenance invariant exact.

**⚠ The landmine this creates for anyone widening the candidate set** (register **LNK-033**):
the keep-branches rest on "no resolver path can emit a utility-area url". Widen
`loadContentHubs`/`loadInteractivePages`, drop `candidatesFromHubs`' filter, or add a
utility-area schema `fallback` to a `ctaFieldNames` component, and both writers start freezing
the resolver's own output — with the detector arm that would have noticed now demoted. That is
filed as `bugs_open/308`, which is routed here and must land with recorded provenance, not
before it.

---

## CONTRIB 2026-08-22 from `bugfix_308_cta_destination_provenance` — your phase-1 open question is being answered elsewhere, and one answer is already decisive

**Nothing in this lane's files was edited.** Telling you rather than only measuring, per the
2026-07-29 owner ruling.

Your PLAN's phase-1 open question is *"whether to widen `candidatesFromHubs` to guide pages"*.
`bugs_open/308` is that question one page wider, and the **owner ruled it on 2026-08-18**:
fix candidate 1 — build a real provenance record first, then widen. Candidate 2 (widen the
label-match set + recalibrate stopwords) is explicitly **not** the plan on its own, because it
falsifies `bugs_open/248`'s live LNK-033 invariant the instant it lands. So if this lane was
holding the question open pending canary quality, it is now decided independently of the
canary, and the provenance half is being built in the lane named above.

**One finding you can use immediately, whatever the provenance work does.** Enumerating the
seam's consumers turned up a **third** caller of `loadContentHubs`/`loadInteractivePages` that
neither `bugs_open/308`, `bugs_open/248` nor LNK-033's landmine names:
`render_site_components_action.go:182-190`, **the site header's CTA fallback**.

> So a widening done at the **loaders** silently re-picks **every site's header button** — the
> header derivation reads the same two functions and takes `ordered[0]` by nav_order, so a page
> newly admitted to the loader result can outrank today's pick without going near a utility
> area. **Widen at `candidatesFromHubs`, never at the loaders.**

And the natural check would not catch it: [MEASURED 2026-08-22] `site_components` carries
**0 `cta_url` keys across all 24 header rows** — the header CTA is never persisted, only
rendered — so a `content_data` before/after diff reads clean while all 24 headers move.

**Also worth knowing for your step 2**, which relies on `applyCTARecompute` label-matching the
new wording: the detector already computes the correct destination and writes it into the
finding as `suggested_target`, and **nothing in the tree reads it** (3 grep hits, all writers).
The rerender is handed only `spec.reason="cta_links_stale"` and re-derives from the narrower
set. That is why a reworded label whose best match is a guide/contact page will still not be
honoured today, even after your rewrite lands.

Lane dir: `docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/`.
