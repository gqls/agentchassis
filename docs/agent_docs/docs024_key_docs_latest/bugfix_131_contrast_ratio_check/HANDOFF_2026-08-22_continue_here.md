# HANDOFF — continue here (contrast_ratio check lane, `bugs_open/131` vonc slug)

**Written 2026-08-22 end of day.** Cold start: read this, then `NOTES_contrast_ratio_check.md`
(newest at the bottom) and `PLAN_2026-08-22_contrast_ratio_check.md`. ⚠ **131 is an ambiguous
number** — this lane is the **vonc gauntlet usability audit** slug; `bugs_closed/131` is an
unrelated og-image case.

## 0. THE ONE THING THAT BLOCKS EVERYTHING ELSE (owner action)

`claude-sonnet-5` is returning **HTTP 400 "You have reached your specified API usage limits. You
will regain access on 2026-09-01 at 00:00 UTC."** First seen `2026-08-22 18:00:14Z`.

That model carries **81 step-configs across 18 active agent types (as of 2026-08-22)** — **all 17
`council-gate` seats and all 21 `fix-proposer` seats**, plus `feature-designer`,
`experience-planner`, `experience-approval-council` and the tool generate/improve/recreate chain.
**So the council gate and the fix loop are down** until the cap is lifted or the model repointed.
⚠ The fleet still reads healthy in aggregate (69 COMPLETED / 6 FAILED in the first hour) because the
haiku dispatch and checker lanes are unaffected — **do not read throughput as evidence the council
works.** ⚠ A root-level `ai_service.model` census shows ONE agent and badly understates it; count
the **step** overlays (`jsonb_each(default_config->'workflow'->'steps')`) — the `bugfix_257` trap.

## 1. Where the work actually stands

| piece | state |
|---|---|
| `contrast_ratio` Tier-4 check | **LIVE on `browser-runner-adapter` v1.0.1326**, proven at the artefact (provenance `27b932aca`; both commits ancestors, with a negative control; binary needles 1/1/1/0) |
| council rounds 1–2 | r1 REVISE (found a real blind-pass defect), **r2 APPROVED** — corr `7e2391ec` |
| the bounded-backdrop refinement | **committed, NOT in the live image** — needs the next `browser-runner-adapter` build |
| council round 3 (the refinement) | **NO VERDICT** — a seat died on the quota above. 7 of 7 completed seats approved with no objections, incl. the seat that gated r1. **Do NOT write `Council-Reviewed:`** |
| fences / seeds / planner prompts | **still name the check NOWHERE**, deliberately (LANDMINES:512) |

## 2. Next actions, in order

1. **When the quota clears** (or the model is repointed): resubmit the refinement —
   `RESUBMIT_CORR=7e2391ec-47d0-4820-afde-b4cc475714e5 ./…/097_TRIGGER_council_review_v1.sh
   docs/…/bugfix_131_contrast_ratio_check/council_submission_2026-08-22_r3.json`. The 7 harvested
   approvals are in NOTES if you want to shorten the round's framing; only `review_architecture`
   never ran.
2. **Rebuild + roll `browser-runner-adapter`** (it is its OWN image — a chassis roll does nothing;
   `make build-browser-runner-adapter`, bump `IMAGE_TAG`, and the overlay's `newTag` in the SAME
   commit). Then re-prove with the RUNBOOK's three-needle grep.
3. **Re-run the witness** — `scripts/witness_contrast_ratio.sh` — and expect **`failed:1`** where
   today it gives `passed:3 failed:0`. That flip IS the proof the refinement works live. (The reply
   topic does not auto-create; read the adapter's `run_checks complete` log line for the counts.)
4. **Only then** Phase 2 vocabulary: planner prompts and `259_experience_approval_council.sql`'s
   deferral-honesty list. Not before — an unknown check type is SKIPPED and a skipped fence can
   record a PASS.
5. **Phase 3 needs a human decision, and here is the number for it:** **145 firm contrast failures
   across 24 live homepages (as of 2026-08-22)**, only 3 of them from the refinement. If
   `contrast_ratio` joins standing fences today it fails most sites. ⚠ 145 is a FLOOR — homepage-only
   is the exact sampling error `bugs_closed/122` warns about (its own runs differed ~2 orders of
   magnitude).

## 3. Bug 131 itself — what is owed and by whom

Stays OPEN. Not this lane's to close:
- **Item D** (phone column, **65.6%** — worse than the 74% originally complained about) was **never
  decided**; three documents disagree. Contributed to the gauntlet lane's design pass
  (`gauntlet_dead_cta/HANDOFF_2026-08-12_design_pass_START_HERE.md` §8).
- **Item A** decayed to 2.48:1 and its remedy is **restated** in the bug file: not another accent
  re-pin — the section's *body* text (`gi-eyebrow`, `gi-rules-label`, the rules list) is below AA on
  every reading of its backdrop.
- **Item H**: engineering complete 07-31; the residue is the owner's own distribution leg.
- **Nobody has recorded who closes 131.**

## 4. Tools in this lane (all committed under `scripts/`)

`witness_contrast_ratio.sh` (drives the deployed adapter directly — **no** acceptance judge, so it
files nothing at another lane's surface) · `dump_probe_test.go.txt` + `run_deployed_probe.py`
(extract the probe FROM SOURCE and run it live — no hand-copying, so scan-vs-deployed drift is
impossible) · `induced_backdrop_controls.py` (four cases; the discriminating pair is same-colours
flat-vs-`url()`) · `blast_radius.py` (**reads targets from `sites`, never from recall** — a
hand-typed domain cost me a near-filed bug against someone else's site).

## 5. Four wrong calls are logged in `WRONG_CALLS.md` today

A "(run <date>)" receipt typed before the query ran · the same error again eight hours later in a
paragraph answering a reviewer about unverified claims · **inheriting a suppression rule without
testing what it suppresses** (the big one — it made the check blind on its own founding case) · a
near-filed bug against `dartsonline.co.uk`, a domain I typed from memory, "confirmed" by three tools
that were all asked about the wrong host.
