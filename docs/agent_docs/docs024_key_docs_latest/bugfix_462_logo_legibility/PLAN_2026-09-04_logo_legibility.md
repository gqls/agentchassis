# PLAN — 462, logo legibility (lane opened 2026-09-04)

**The bug:** `bugs_open/462_HANDOFF_2026-09-03_a_logo_can_be_perfectly_rendered_correctly_deployed_and_illegible_and_no_check_can_see_it.md`.
Read §3 (candidates), §7 (the owner's ruling), §8 (the sweep) and §9 (routing) — in that order.
Everything below assumes them.

## Why this lane exists

462 was filed by `bugfix_417_logo_text_policy` and worked there until 2026-09-04, when the owner put
it on its own session. The transfer is recorded in `bugs_open/462` §5 — **and nowhere a tool can
read it**, which is the same structural gap §5 already documents about the original filing.

417 keeps: invented wordmarks, the fence residual. **462 owns:** whether a mark can be seen.

## What is settled, and must not be re-opened

| | |
|---|---|
| **Fix shape** | **Candidate 2 — report it afterwards** (owner, 2026-09-03). NOT candidate 1, a fail-closed contrast statistic at store time. The owner ruled against the §3 ranking with the facts in front of him. |
| **websitepromotion.co.uk** | Its illegible logo **stays** (owner, 2026-09-03). It is the detector's test case, not an open repair. The estate knowingly serves one illegible logo. |
| **The two preserved PNGs** | In the 417 lane dir. **Evidence, not a rollback option** — the pre-regeneration artefact exists nowhere else. |
| **Measure after matting, against the header** | Never against the keyed generation ground (§6). A pre-matte check passes a white-on-magenta mark happily. |

## Where the work stands

**Built and proven:** `scripts/audit-logo-legibility.py`. Two arms, each with a named artefact behind
it; `--self-test` runs offline and fires on both preserved websitepromotion marks. Re-verified by
this lane on takeover (2026-09-04 13:25Z): 6/6 self-test cases, fires on both live findings, control
passes.

**The population is the real finding.** 34 active logo assets; only **7** can be judged at all,
because 22 have a background baked in (pre-`424`) and the header is not their backdrop. So "how
widespread is 462" is not yet answerable for most of the estate.

**Open:** routing. §9 works it and the answer is uncomfortable — see below.

## The decision this lane WAS waiting on — **RULED 2026-09-04: option (A)**

> **The owner ruled option (A): apply the standing check, defer the filer.** It is live. The fork
> below is kept as written because the reasoning is what a later reader needs when (B) comes back on
> the table — not because it is still open.
## The decision this lane is waiting on

§9e's fork. Stated once here so it is not re-derived:

- **(A) standing report only** — schedule the sweep, a human reads it. No item type, no handler.
- **(B) file, with provenance routing** — generated marks converge onto
  `needs_imagery:site:-:logo` at `image-build-handler`; uploaded marks go somewhere a human sees,
  *not* `needs_human_review`.
- **(C) neither** — leave it hand-run. This is the status quo and it decays.

**This lane recommends (A) now, (B) when there is a target it may act on.** The reasoning is §9d:
an automatic filer has **zero** legitimate targets today (one finding the owner ruled hands-off, one
an operator upload with nothing to regenerate from), while its only automatic remedy is an
irreversible re-roll by a generator that has no legibility criterion — the thing §6 measured making
a logo worse.

## Phasing

1. **Take over cleanly** — transfer recorded, corrections made, detector re-verified at the artefact
   rather than inherited on trust. ✅ 2026-09-04.
2. **Work the routing question and put the fork to the owner.** ✅ 2026-09-04 (`bugs_open/462` §9).
3. **Make the sweep standing** — the half of "report it afterwards" that needs no further decision.
   Check-fleet CronJob; needs `PG_CLIENTS_HOST` direct Postgres (no `pods/exec` RBAC in-cluster),
   an image rather than a ConfigMap script (Pillow + outbound HTTPS), and one `doc_notes` row per
   run *including on clean results*, so "looked and found nothing" stays distinguishable from
   "stopped running". ✅ **BUILT 2026-09-04 (`c0e2900ff`) and deliberately NOT APPLIED** —
   `deployments/kustomize/services/logo-legibility-check/`, daily 08:15 UTC. Applying it is one
   command and is **pending the owner's answer**, because applying it *is* choosing option (A).
   ✅ **RULED AND APPLIED 2026-09-04 — option (A).** Live at 08:15 UTC daily. Two defects that only
   a scheduled run exposes were found and fixed on the first in-cluster run (462 §10a): the exit
   code, and buffered logs. Verified at the artefact — fleet numbers reproduced, `doc_notes` row
   landed.
4. **Routing** — only on the owner's answer. If (B): the concept-register entry must carry the
   producer set, the `item_key` shape **and the provenance branch**, in the shipping commit
   (owner ruling 2026-08-02).
5. **Option (a), the render-audit home** — still the destination, still unbuilt. The reason is
   staleness, not coverage: this sweep reads a *declared* theme token, and `bugs_open/396` rewrites
   theme rows, so a pass here decays into a false pass. Keep the thresholds in one place so (a) can
   reuse them rather than inventing a second, silently different rule.

## Decisions this lane has taken, with reasons

- **2026-09-04 — do not add an absolute-pixel arm** (`legible_ink_min_px`). Drafted and removed by
  the 417 lane: no artefact motivated it, and it risks false positives on small marks. Both shipped
  arms have a named artefact behind them. **If it is re-added, add the case first.**
- **2026-09-04 — take no verdict on baked-background marks.** The sweep reports `baked_bg`,
  `baked_max` and `baked_legible_frac` and stops there. There is no known-bad artefact to calibrate
  a threshold against, and choosing a number would manufacture 22 sites' worth of verdicts out of
  nothing. **A stated blind spot is better than an invented threshold.**
- **2026-09-04 — routing branches on `assets.origin_type`, whatever else is decided.** A generated
  mark and a human's uploaded mark are different findings with different remedies. The sweep does
  not read `origin_*` today; anything that files must.
