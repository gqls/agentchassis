# BUG 067 — feature-designer's repropose cap made every revise cycle fatal

**Filed:** 2026-07-24 · gauntlet_dead_cta / feature-builder B4 shakeout · **CLOSED same day (config-only fix, live immediately)**
**Severity:** high — the designer could only ever ship round-1 approvals; any council
REVISE killed the run, so council feedback was structurally unusable.

## Symptom
Two designer runs died after a round-1 REVISE:
- corr `c2a9fd27` (2026-07-23): stuck at `repropose` ~4h, then FAILED with **no
  recorded error** (older image; undecoded death).
- corr `7773219b` (2026-07-24): died at `repropose` with the decoded cause:
  `stop_reason=max_tokens (output_tokens=16000 reached the configured cap, 25896 chars
  recovered)`.

## Root cause
`feature-designer`'s `repropose` step configured `max_tokens: 16000` while a full
staged plan runs ~26k characters — the reviser must re-emit the whole plan, so **every
revise cycle hits the cap**. `compose` was already at 32000; the two steps drifted.
Same class as the experience-loop's run-10 compose death (seed 176) — a whole-artifact
re-emitter whose cap was sized for something smaller.

## Fix (live)
Migration `201_designer_repropose_cap.sql` (applied + ledgered, snapshot taken):
repropose `max_tokens` 16000 → 32000, matching compose. Config-only, no image.

## Verify
The next designer REVISE round that completes a repropose (fix_plan artifact count ≥2
on one corr) is the behavioural proof — expected on corr `ffb74056` (round 5) or the
next revise the council issues. Until then the fix is applied-and-asserted (the
migration's DO block read back 32000), not yet behaviourally exercised.

## Note
The 07-23 instance shows why the v1.0.1138-class stop_reason decode matters: the same
defect was a silent 4-hour hang on the older path and a clean named error on the new
one. Cap sizing for whole-artifact re-emitters belongs in review checklists — any step
that re-emits an artifact must have max_tokens ≥ the artifact's realistic ceiling.

## ADDENDUM 2026-07-24 (same day) — the `design` and `reframe` steps had the same cap

Round 5 (corr `ffb74056`) died at the INITIAL `design` step, stop_reason=max_tokens at
16000 — the same defect one step upstream, exposed when the capability spec grew. 201
fixed only `repropose`; **migration 202** raises `design` + `reframe` to 32000 so all
three whole-artifact emitters match. Reviewer seats stay at 8000 (verdict JSON only).
Lesson sharpened: when a cap defect is found on ONE step, sweep EVERY step of that
agent for the same sizing before closing — I closed 067 one step too narrow.
