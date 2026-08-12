# BUG 253 — `BestLabelMatch` scores by RAW overlap count, so a long marketing `nav_label` that incidentally contains another page's distinctive words ties with that page — and an alphabetical tie-break then picks the wrong one

> **FIX HANDOFF (2026-08-11):** a fresh session should start from
> `docs/agent_docs/docs024_key_docs_latest/bugfix_203_phantom_cta_cleanup/HANDOFF_2026-08-11_fix_bug_253_label_match_scoring.md`
> — it carries the 2-minute local repro, the two-candidate-pool trap, worked scoring
> numbers for the options below, the calibration method and baseline, and the process
> steps. This file stays the case record.

> **FIXED AT COMMIT `f1819861f` (2026-08-11).** `BestLabelMatch` now ranks a candidate by
> identity-token overlap (label tokens present in the candidate's own `name`/`title`) ahead
> of total-token overlap (name+title+`nav_label` combined, the old and only signal);
> interactive-preference and the alphabetical-`Name` final tie-break are unchanged. A second
> ranking key — smaller token-set-size as a further tie-break, meant to replace the
> alphabetical tie-break — was implemented and calibrated, then **dropped before shipping**:
> it flipped already-correct picks on live data (all 9 CTAs on gaswholesalers.com, driven by
> a stray hyphen tokenising "Break-Even" into two words) for reasons unrelated to
> correctness, so the shipped ranking keeps the original alphabetical `Name` tie-break.
> Calibrated twice against the live fleet, both candidate pools the shipping code actually
> uses: **detector pool** 784 examined / 697 matched / 347 newly-resolved / OVERRIDDEN
> 208→205, 18 changed picks (2.3%), each individually inspected and read as an improvement or
> a defensible identity-grounded pick; **resolver pool** 784/401/196/44, OVERRIDDEN unchanged
> at 44, 0 changed picks (its `contentHub` candidates carry no `nav_label`, so identity==total
> in that pool structurally — 253's mechanism cannot regress it). Full numbers and the
> gaswholesalers/vetcomparison/robot-hands regression traces:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_203_phantom_cta_cleanup/CALIBRATION_2026-08-11_label_match_identity_report.txt`.
> Submitted to the council gate before commit: `Council-Submitted: ccef36de-6757-4777-91db-37864b018622`.
> `bugs_open/248` is unaffected by this fix and remains the other blocker on draining the
> `misdirected_cta` repair queue.

> **FIXED AND LIVE — CLOSED 2026-08-12.** Moved to `bugs_closed/` per the restored fixed-AND-live
> bar (owner, 2026-08-12: "if it is fixed and live it should be moved"). Both checks done
> directly, not inferred from a roll announcement:
> 1. **Build provenance, at the artefact, not git.** The startup "build provenance" log line
>    had scrolled out of `--tail=3000` on both live `agent-chassis` pods (a known landmine —
>    it is a startup-only line on a busy service). Fell back to the binary probe: extracted
>    every 40-hex substring from `/proc/1/exe` (one bounded `grep -aoE`, not a blind
>    single-string discovery grep, which the landmine warns matches Go's internal digit
>    tables on every service) and cross-checked each candidate against real commit hashes in
>    this repo's history with `git cat-file -e` — a spurious binary-table match cannot also be
>    a real, existing commit sha in this specific repo. Exactly one candidate matched a real
>    commit: `da5a7eb8ff12f78a3569d8474363445013a77557` (2026-08-12 15:37:47+01), which
>    `git merge-base --is-ancestor f1819861f8316aea502364fb85a4cf006c24b9a5
>    da5a7eb8ff12f78a3569d8474363445013a77557` confirms is a descendant of this fix's commit.
>    Both `agent-chassis` pods started 2026-08-12 14:55Z, after that build. **The fix is live.**
> 2. **Live control, dispatched directly** (the scheduled discovery rotation cannot be
>    targeted): `completeness-discovery-agent` re-run against robot-hands.com
>    (orchestration `c7d41ab0-cc65-42ea-a15b-c14a291fa7a0`, COMPLETED). Read from
>    `collected_data->'discovery_result'->'findings'`, not `site_work_items` (the dedupe
>    trap — fresh findings that match an existing `detected` row file nothing new there).
>    **`how-to-specify-a-gripper` — the exact page this bug was filed against — carries ZERO
>    `misdirected_cta` findings in this run**, down from the 3 that motivated filing this bug.
>    17 `misdirected_cta` findings remain fleet-wide on this run, all a DIFFERENT, unrelated
>    pattern (a `/matchmatrix.html` vs `/tools/matchmatrix/index.html` vs
>    `/matchmatrix-methodology.html` vs `/how-it-works.html` naming confusion) — genuinely
>    separate site content ambiguity, not a 253 recurrence; not investigated further here as
>    out of this bug's scope.
>
> **Council verdict**: the first council run failed on infrastructure, not content — `error`
> column read `reaper: stale EXECUTING_STEP for >4h; step=review_bug_historian`, i.e. one seat
> hung and was reaped, no verdict was ever produced. Resubmitted under the SAME trail
> correlation (`RESUBMIT_CORR=ccef36de-6757-4777-91db-37864b018622`), new run orchestration
> `101c11d9-50d7-4499-98b2-894138213094`, still pending at time of writing. The commit's
> existing `Council-Submitted:` trailer needs no amend — `098`'s report resolves by
> correlation at report time, so approval on this resubmission credits the same commit
> automatically. Whoever next touches this lane: check the verdict
> (`SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
> correlation_id='ccef36de-6757-4777-91db-37864b018622' AND kind='council_report' ORDER BY
> created_at;`) and act on REVISE/REJECTED if it isn't APPROVED — closing this file does not
> retire that obligation.

**Filed 2026-08-11** by the `bugfix_203_phantom_cta_cleanup` lane. Found while running
detection manually to build a repair list — **this defect is why that repair list was not
safe to act on**, and the rollout was halted. **Status: OPEN, not started.**

Third distinct defect in `datahelpers.BestLabelMatch`. Siblings: the category-before-overlap
priority bug (**fixed**, `3bc0486d7`, approved `6cb8c72b-…`) and `bugs_open/248` (a recompute
destroys authored `/contact.html` links). Filed on first-hand verification with a mechanical
repro rather than a `090` run, per CLAUDE.md's 2026-07-31 ruling — the cause is one scoring
expression, and the repro is decisive either way.

## Symptom

`misdirected_cta` flags a link whose text and href **already agree**, and suggests replacing
it with an unrelated page. On `robot-hands.com/how-to-specify-a-gripper` (three anchors, all
three slots), text *"Gripper Safety Factor Calculator"* → href
`/tools/gripper-safety-factor-calculator/index.html` — the correct, active tool page at
exactly that URL — is reported misdirected with
`suggested_target: /tools/gripper-payload-calculator/index.html`.

## Reproduction (mechanical, 2026-08-11, at HEAD)

Real rows, verbatim from live `pages` for robot-hands.com, through the same
`NewLabelMatchCandidate(id, name, title, url, interactive, name, title, nav_label)`
construction `loadCTAMatchIndex` uses:

```
label tokens: [gripper safety factor calculator]
  candidate gripper-payload-calculator              interactive=false overlap=4
  candidate tool-gripper-payload-calculator         interactive=true  overlap=4
  candidate tool-gripper-safety-factor-calculator   interactive=true  overlap=4
BestLabelMatch -> "tool-gripper-payload-calculator" (/tools/gripper-payload-calculator/index.html)
```

**All three tie at 4.** Among the interactive pair the final tie-break is
`c.Name < bestPtr.Name`, and `tool-gripper-payload-calculator` sorts before
`tool-gripper-safety-factor-calculator` — so the wrong page wins on the alphabet.

## Root cause

`BestLabelMatch` (`platform/orchestration/datahelpers/label_match.go`) scores a candidate by
**how many of the LABEL's tokens it contains**, and nothing else. It never asks how much of
the *candidate* that represents. A candidate's token set is the union of its `name`, `title`
and `nav_label` — and `nav_label` here is a 14-word marketing sentence:

| page | the text that supplies its tokens |
|---|---|
| `tool-gripper-safety-factor-calculator` | name + title `"Gripper Safety Factor Calculator \| Tools"`, nav_label **empty** |
| `tool-gripper-payload-calculator` | nav_label `"Gripper Payload Calculator — Validate Capacity with **Safety Factor** \| Robot-Hands.com"` |

The payload page mentions "safety factor" as *prose about what it does*. That is enough to
absorb every token of a label that names a different page. The signal a human uses instantly
— safety-factor matches **4 of its 5** tokens, payload matches 4 of ~15, and the match is
against payload's *description* rather than its *name* — is invisible to a raw count.

**A bigger token set is strictly an advantage under this scoring.** Any page whose
`nav_label` is a rich sentence competes for every label on the site, which is the opposite of
what a "does this text name this page?" test should do.

## Blast radius

**Not yet measured fleet-wide** `[UNMEASURED]` — deliberately, because the query is not
cheap to get right (it needs the matcher, not SQL, exactly as the 2026-08-08 calibration
found). What IS measured, 2026-08-11:

- One manual `completeness-discovery-agent` run over robot-hands.com produced
  **misdirected_cta findings on 16 of its pages**, and the three inspected in detail on
  `how-to-specify-a-gripper` are all this false positive.
- The `misdirected_cta` queue fleet-wide holds **192 `detected` / 95 `unresolved` /
  63 `failed`** items. The lane's 2026-08-07 note that this queue is "substantially false
  positives" now has a second concrete mechanism (the first being `248`'s excluded-area
  behaviour).

Every site with tool pages carrying descriptive `nav_label`s is exposed, and that is the
normal shape of a generated site.

## Why this blocks the repair work specifically

The repair path recomputes rather than applying `suggested_target`, but it recomputes with
**this same function**, so it lands on the same wrong page. Acting on the 16 robot-hands
findings would have rewritten three correct safety-factor links to the payload calculator.
**Detection is currently the only stage running by hand; do not promote or dispatch
`misdirected_cta` items until this is fixed** (and see `248` for the second, independent
reason not to).

## Fix candidates, ordered by what closes the door

1. **Score by a normalised measure, not a raw count** — e.g. overlap ÷ |candidate tokens|
   (or an F1/Jaccard blend of that with today's overlap ÷ |label tokens|). Under any of
   these, safety-factor (4/5) beats payload (4/15) decisively and no tie-break is reached.
   Closes the class: a verbose candidate stops being advantaged. Needs recalibration against
   the fleet (the 2026-08-08 method — the real function over live data via
   `kubectl port-forward`, not a SQL approximation — is written up in the lane's NOTES).
2. **Weight the token SOURCES**: `name` and `title` are identity; `nav_label` is
   description. Scoring a name/title hit above a nav_label-only hit fixes this case
   directly and is a smaller change than (1). Requires `NewLabelMatchCandidate` to keep
   per-source token sets instead of unioning them — a real but contained change to a
   struct whose `tokens` field is deliberately unexported.
3. **Require the match to cover the candidate's distinctive name tokens** — i.e. a
   candidate cannot be "named" by a label that misses a word in its own name
   ("payload"). Cheap and targeted; risks being too strict where titles and names diverge
   legitimately.
4. **Break ties by candidate-token-set size (smallest wins) instead of alphabetically.**
   A one-line change that fixes exactly this transcript and nothing more — the alphabet is
   plainly the wrong discriminator, so this is worth doing *even alongside* a real fix, but
   it does not address near-ties that are not exact ties.

**Do not narrow by adding "safety"/"factor"-style words to `LabelStopwords`** — they are
genuinely distinctive here, and the standing landmine (narrowing past an invented false
positive makes a rule inert) applies directly.

## How to verify a fix

1. The repro above: `"Gripper Safety Factor Calculator"` must resolve to
   `/tools/gripper-safety-factor-calculator/index.html`.
2. It must not regress the two fixes already shipped in this function — the full existing
   `label_match_test.go` plus `TestBestLabelMatchOverlapBeatsCategory` and
   `TestBestLabelMatchInteractiveTiesBreakToInteractive` must keep passing.
3. Re-run `completeness-discovery-agent` against robot-hands.com and confirm the
   `how-to-specify-a-gripper` findings drop from 3 to 0 (it is the cleanest available
   control: those three anchors are known-correct).
4. Recalibrate before shipping — a scoring change moves every match on the fleet, which is
   exactly what the 2026-08-08 calibration exists to measure.

## Related

- `bugs_open/248` — the other reason the `misdirected_cta` queue cannot be drained.
- `bugs_closed/023` — original label/URL pairing case; `chooseCTATargets` was label-blind.
- `bugfix_203_phantom_cta_cleanup/NOTES_phantom_cta_cleanup.md` (2026-08-11) — the session
  trail, including why the manual-detection plan stopped here.
