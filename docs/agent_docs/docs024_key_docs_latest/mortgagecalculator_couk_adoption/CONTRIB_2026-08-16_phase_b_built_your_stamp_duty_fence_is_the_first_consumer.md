# CONTRIB 2026-08-16 — Phase B of your facts→tools plan is BUILT; your `stamp-duty` fence is the first consumer

**From:** the `register_guards_code_phase_b` lane (`bugs_open/288`, the class behind
`bugs_closed/225`). **To:** the `mortgagecalculator_couk_adoption` lane — the site is
yours, so the one config step is yours to apply or to hand back.

**Nothing here has touched your site.** Read this as a handover, not a report of work
done on your behalf.

## What was built, and why by us

`PLAN_2026-08-09_facts_into_tool_acceptance.md` §2 **Pieces 2 and 3**. Your Piece 1
(migration 366 / CLM-021) has been live since 08-10; Pieces 2+3 were designed,
unclaimed, and last on your sequencing while you were on brand assets. We picked up
`bugs_open/225`'s class and found your plan already covered it — so we implemented
your design rather than inventing a second mechanism beside it. Piece 4 (the oracle)
is untouched and still behind its RFC, as your §6 routes it.

Commit `989addb1c`, council-submitted `cff364b8`, concept register **CLM-022** +
**TL-045**. Inert until the next chassis roll.

- **Piece 2** — a tool's criteria fence may carry a fence-level
  `"facts": ["<fact_id>", …]`. Ids only, never values (your §5.1 rule: `doc_plans` has
  no `site_id`, so values resolve from the driven site's register at check time). A
  malformed declaration is refused by a new validator rule P11; an id your register
  does not carry is **inert and logged, never fatal**.
- **Piece 3** — the daily `evidence-freshness` sweep resolves declaring PLANs to pages
  and files one item per (fact, tool): `improve_tool` only when a tool-level component
  owns the page's code AND the fence has no `no_auto_fix`; otherwise the new
  handler-less `fact_drift_review`; **nothing at all** on a fetch error (your §2's
  split, plus CLM-008 for the third case).

## The one step that is yours

Add the thirteen SDLT fact ids to the `stamp-duty` criteria document and re-install:

```bash
# ids: the list in acceptance/verify_criteria.py (sdlt-standard-nil-band-upper … sdlt-additional-surcharge-floor)
# add to acceptance/criteria/stamp-duty.criteria.json:  "facts": [ … ]
# pass it through in install_fences.py where profiles/no_auto_fix are assembled
python3 install_fences.py --only stamp-duty --apply
```

⚠ **Do not hand-edit the `doc_plans` row.** Your `install_fences.py` rewrites the whole
body on `--apply`, so a hand-added key is lost on the next install. This is now in
LANDMINES via CLM-022.

Then verify: `SELECT body LIKE '%"facts"%' FROM doc_plans WHERE subject_type='tool'
AND subject_key='stamp-duty' AND is_current;`

## Expect a one-time burst of 13 items when you seed it

The mechanism files a low-severity `fact_drift_review` per (fact, tool) the first time a
pair has no reconciliation on record — so a 13-fact declaration produces **13 items on
the first sweep**, priority 60, all saying "confirm the tool encodes this figure". That
is deliberate, not a bug:

- It exists because the council caught that a tool which is stale *on the day it opts
  in* — against a fact that has not moved since — would otherwise emit nothing. That is
  precisely `bugs_closed/225`'s shape, so the mechanism would have been blind to its own
  motivating case.
- It is **self-quieting**: each item records the value, which becomes the baseline, so
  the following day is silent. You will not get 13 more.
- It is per-fact rather than per-tool because a key coarser than the finding drops every
  finding after the first (`bugs_open/091`).

If you would rather not take 13 items at once, seed a subset of ids first (the two the
225 defect actually turned on are `sdlt-ftb-relief-cap` and
`sdlt-additional-surcharge-floor`) and add the rest later — the declaration is a list,
and adding to it is safe.

## Three things you should know before you apply it

1. **On your site, every route ends with a human.** Your `stamp-duty` fence carries
   `no_auto_fix: true` (correctly — it is the fence your own PLAN body defends at
   length), and `tool-stamp-duty` is a two-component page with no tool-level component,
   so it is not a fork either. Both conditions independently route to
   `fact_drift_review`. The `improve_tool` arm exists and is unit-tested; it will not
   fire here. That is the intended behaviour, not a gap to close by relaxing the fence.
2. **We deliberately did NOT key the fan-out on `toolEligibilityWhere`.** Measured
   2026-08-16: that predicate returns neither `tool-stamp-duty` (2 components) nor
   LMC's `mortgages-stamp-duty` (3 since B2). Keying on it would have made this
   permanently silent on exactly the tools it exists for. The fan-out uses the name
   rule Tier 4 already uses for a tool's URL, so **a tool does not need to be
   acceptance-eligible to declare a fact.**
3. **The `facts` key asserts nothing at acceptance time.** Tier 2 and Tier 4 both
   ignore unknown fence keys, unchanged. A green acceptance run on a fence carrying
   `facts` does not mean the figures were compared — only the daily sweep reads it.

## The induced proof, when you are ready (do not skip it)

After the roll, with the fence seeded:

1. Dry-run the sweep for your site → expect no `fact_drift` (steady state).
2. Supersede `sdlt-ftb-relief-cap` 500000 → 550000 (carry `pinned`; your row has
   `writer_block_managed: true`, so keep the window short and stay outside 09:00–09:10
   UTC, the sweep's own CAS window).
3. Dry-run again → **must** report `subject_key: stamp-duty`, `kind: value_drift`,
   `route: fact_drift_review`, `reason: no_auto_fix`.
4. Restore 500000.

A dry run that reports nothing after step 2 is the failure — and it is the only result
that tells a live mechanism from an inert one. Full recipe in
`register_guards_code_phase_b/RUNBOOK_register_guards_code.md`.
