# HANDOFF — bugfix 213, continue here

> **UPDATED 2026-08-11.** The instance fix is **LIVE and PROVEN** on `v1.0.1284`, both
> replicas, by the sanctioned binary probe with a two-sided control. The council
> APPROVED it round 1. **Three owner rulings (D1–D3) were taken on 2026-08-11 and are
> the whole of the remaining work** — they are in `PLAN…md` §"OWNER DECISIONS" with
> reasoning, and summarised as next steps below.
>
> **Read the two corrections in that PLAN section before quoting this lane**: (a) "the
> colour fixer cannot repair these" is verified for ONE of the 11 and UNVERIFIED for
> the rest; (b) re-detection is LIVE (230's rotation carries `design-discovery-agent`
> across 22 sites), which contradicts what an earlier draft of this handoff implied.

**Cold-start for this lane.** Read this, then `NOTES…md` for the missteps, the mutation
matrix and the verification history. `PLAN…md` has the design, the rejected candidates,
and the owner rulings.

---

## NEXT STEPS — the three owner rulings, in dependency order

### D3 — ✅ BUILT 2026-08-11 (`cmd/verifier-remit-check`, register **WII-015**)

**Done, and the code is committed.** A daily CronJob Go binary, linked against the
live verifier registry, that files an undispatchable `verifier_remit_gap` work item
for any verified item_type carrying more than one producer shape with no declared
`Grades`, and closes its own findings on a positive re-observation. Council trail
corr `fc082c4a-4b00-4835-8ffe-11a55e53f47a` (round 1 REVISE — every objection
answered by measurement, see NOTES; round 2 submitted).

**DEPLOYED AND PROVEN 2026-08-11.** CronJob live at `25 7 * * *` UTC on image
**`v1.0.1289`**, verified as an unbroken chain rather than inferred: the image's
`org.opencontainers.image.revision` label is commit `74ac4ed3a`, the running pod's
`imageID` digest is byte-identical to the digest `docker push` printed, and the Job
wrote its own artefact — `SELECT … FROM doc_notes WHERE source='verifier-remit-check'`
(18:45 and 18:54 UTC): *12 item_types evaluated, 0 findings, `hardcoded_section_colors`
named as answered by its declared remit*. `verifier_remit_gap` rows: **0**, correctly.
Council **APPROVED round 2**; the one code change it produced (a constant census SQL,
no interpolation) is in `74ac4ed3a`, and the other seven objections were each answered
by a measurement recorded in NOTES.

**What is left on D3 — one thing, and it is a waiting job, not a task:**
- **Deployed is not exercised.** The detector has never filed a row, because nothing
  currently qualifies. Its own `--ignore-remit` control reproduces the original bug as a
  live finding, so the zero is a zero that looked — but the first REAL filing is still
  unobserved. If you want it exercised deliberately: register a second producer's shape
  on any verified type, or watch for the next convergence.
- Also raised out of this round and left open on purpose: **`RFC_024`** — there are
  **nine** CronJob meta-checks with no shared harness, and three council seats have now
  asked for a consolidation pass twice.

The detector fires today only under `--ignore-remit` (its built-in disconfirmability
control, writes refused), because the one two-producer type now declares a remit.
**Deployed is not exercised**, and the two are different claims.

<details><summary>The original D3 brief, kept for the reasoning</summary>

#### BUILD the class-level detector
A periodic check flagging any **verified** `item_type` accumulating rows with more than
one spec-shape / `audit_source` and **no** `Grades`. The query is already written, in
this lane's RUNBOOK §"Find every verified item_type with more than one producer"; the
work is turning it into a scheduled check that files a work item.
- Key it on the **spec shape**. NOT on `created_by` (bottoms out at `generic`), NOT on
  a producer list (refuted, `bugs_open/213` §5.3).
- It must drop a type from the finding once that type registers a `Grades`.
- It mints a new item_type ⇒ **the `verifier_coverage_test.go` obligation applies**
  (classify in `itemTypesWithoutVerifiers` AND `liveItemTypes`, same commit), and if you
  give it a verifier, the `sql_for_agents/220` claim-timeout exclusion must move in
  lockstep — `TestRegisteredVerifiersMatchClaimTimeoutExclusion` enforces both ways.
- Disconfirmable today: returns 2 for `hardcoded_section_colors`, 1 for the other ten.

</details>

### D1 (the big one, spans sessions) — the acceptance_test verifier + settle routing

> **URGENCY RAISED 2026-08-11 by measurement, not by argument.** `dark_section_audit`
> already carries **14 rows, all filed today, 13 already `complete`** — the rotation
> re-detected within a day of the roll, exactly as D2 assumed, and every one of those
> 13 closed **ungraded** because the new type still has no verifier. D2's stated
> dependency on D1 is no longer hypothetical: the machine is re-finding these defects
> and losing them again on a ≤7-day cycle, ~13 at a time.
**This is the gap this lane created**: `dark_section_audit` has no verifier, so its
items now close *ungraded* rather than *mis-graded*. Both close clean.
1. Build a verifier over `spec.acceptance_test` using `criteria_check` (RFC_002). That
   field is read by **nothing** today. All 11 live acceptance tests are mechanical and
   browser-checkable.
   ⚠ **The design question, not an implementation detail:** this puts a browser /
   computed-style evaluation on the **completion path**, which this estate has
   deliberately kept free even of HTTP probes (`verifier_coverage_test.go:171` records
   the standing objection). Own council round. Argue from `contrast_failure`, which has
   the same browser dependency and is currently answered by re-detection instead.
2. **Measure before re-routing.** Check each of the 11 acceptance tests against
   `ReplaceHardcodedColors`' actual remit. Only gamesdesign's already-`var()` case is
   confirmed outside it; several others name inline `style` / `rgba(0,0,0` and may be
   inside. Do not generalise from the worked instance — that is the move this bug punishes.

### D2 — the 11 mis-closed items: NOTHING BESPOKE. Let the rotation re-detect.
`site_discovery_rotation` carries `design-discovery-agent` across 22 sites, last
selected 2026-08-10, so a still-present defect re-files itself on a ≤7-day period under
the new type with a fresh dedup key. Historical `complete` rows stay — they are the
honest record of what the machine did.
- ⚠ **D2 DEPENDS ON D1.** Until the new type has a verifier, a re-detected item closes
  unverified — the rotation finds the defect and loses it again. If D1 slips, revisit.
- Accept, and state rather than quietly drop: we never get a false-complete *count* for
  the original 11. "11 closed" stays an upper bound, 1 confirmed broken, 1 confirmed clean.

### Still owed on the fix itself
**The gate has not FIRED.** `SELECT count(*) … WHERE result->'_verification'->>'status'
='out_of_scope'` = **0**. Deployed ≠ exercised, and `bugs_open/213` stays OPEN until a
`hardcoded_section_colors` item without `spec.check` reaches completion and lands
`triaged`/`failed` with the scope-mismatch error. Migration `374` is **not needed** —
0 in-flight rows, re-measured post-roll; do not ship an empty migration.

---

## Verification recipe for this service (learned the hard way, 2026-08-11)

CLAUDE.md's new first step — `kubectl logs -l app=<service> | grep 'build provenance'`
— **returns nothing on agent-chassis**: the log had 13 lines and retention is seconds,
so the startup stamp is long gone. Use the sanctioned fallback, and note `strings` is
now BANNED (it produced three wrong readings in one day) and a tag is not a commit
(`v1.0.1284` straddles three revisions, `bugs_open/249`):

```bash
kubectl -n ai-persona-system exec <pod> -- sh -c '
  probe() { if grep -aq "$1" /proc/1/exe; then echo "PRESENT  $1"; else echo "ABSENT   $1"; fi; }
  probe "verifier_scope_mismatch"                    # Half B
  probe "verification_unavailable"                   # POSITIVE control (live since RFC_017)
  probe "zzz_this_string_must_never_exist_213"       # NEGATIVE control
'
```
No `2>/dev/null` (it makes a missing tool look like a missing symbol) and never a
discovery grep for "some 40-hex string" (it matches Go's internal digit table).

---

## What is done

| | state |
|---|---|
| Half A — `dark_section` → its own item_type `dark_section_audit` | **committed** `2d151c41f` |
| Half B — `VerifierPolicy.Grades` remit contract + gate enforcement | **committed** `2d151c41f` |
| Coverage-guard classification (both maps) | **at HEAD**, but see the passenger note below |
| Tests (join guard, over-correction guard, out_of_scope reason) | **committed**, mutation-proven |
| WII-013 register entry + index row | **committed** `3c72619fc`, `3895be34e` |
| Two LANDMINES entries + `landmines-sync.py --apply` | **committed** `3c72619fc`, synced |
| Standing five (PLAN/NOTES/RUNBOOK/README) | **committed** |
| Council | **APPROVED r1**, verdict read, all 4 advisory objections actioned or answered (`5d482297e`) |

**Build state, verified:** `go build ./...` clean and
`go test ./platform/orchestration/actions/...` green against a clean
`git archive HEAD` extraction, run after the final commit. ⚠ **Do not judge this by
building the live working tree** — it was concurrently broken by another session's
in-flight `save_page_sections_action.go` / `save_sections_decision_gate.go`, which are
nothing to do with this lane.

> **Passenger note, so `git show` does not mislead you.** `2d151c41f`'s message
> describes seven files and the commit contains **six**. My edits to
> `discovery_checks/verifier_coverage_test.go` were swept into another session's commit
> `d644723b8` ("RFC_015 round 1 revisions") before I committed — the documented
> same-file-passenger hazard. **Nothing is lost**: the content at HEAD is exactly the two
> intended edits (verified, 4 occurrences of `dark_section_audit`, correct text) and
> forward-only means it stays where it is. Find it with
> `git log -S dark_section_audit -- <path>`.

## What is NOT done — in priority order

### 1. The roll, then prove it at the artefact
The fix is **NOT in `v1.0.1283`** (that image predates the commit). After the next roll:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# ✓ NEW — expect 0 before the roll, 1 after
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'verifier_scope_mismatch'"
# ✓ POSITIVE CONTROL — live since RFC_017, must read 1 in the SAME exec
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'verification_unavailable'"
```
Repeat on **every replica** — a label greps a subset (`bugs_open/153`). The control is
what proves the grep and the binary rather than your spelling.

Then the behavioural half, which no grep gives: a `hardcoded_section_colors` item whose
spec carries **no** `check` key must land `triaged`/`failed` with `error` beginning
*"completion blocked: the verifier registered for this item_type does not grade this item"*.

```sql
SELECT status, attempt_count, error, result->'_verification'
FROM site_work_items WHERE result->'_verification'->>'status'='out_of_scope'
ORDER BY updated_at DESC;
```

### 2. Migration `374`, AFTER the roll (not before)
Defensively re-type any still-open producer-B rows filed between the commit and the roll:

```sql
UPDATE site_work_items
   SET item_type='dark_section_audit',
       item_key=replace(item_key,'_hardcoded_section_colors_','_dark_section_audit_')
 WHERE item_type='hardcoded_section_colors'
   AND spec->>'audit_source' IS NOT NULL
   AND status NOT IN ('complete','failed','rejected','wont_fix');
```
**0 rows qualify today — re-measured after the verdict, still 0.** The guardian seat
raised exactly this as a state-transition side effect; the population is empty, so
**do not ship an empty migration**. Re-check after the roll and skip if still 0. Pre-roll it would be pointless
(an audit could immediately re-file old-type rows); post-roll it is idempotent and final.
A pre-roll-filed row that reaches completion post-roll is caught by Half B anyway —
blocked loudly, not silently closed. Next free number was 374; **re-check, it moves.**

### 3. Grade the 11 closed producer-B items — the part needing judgement
**A `complete` count is NOT a false-complete count.** Two verdicts already exist:
gamesdesign.co.uk confirmed **FALSE** at the served artefact (bug §3); relojistas.com
measures **CLEAN** in `bugs_open/122`'s `BASELINE_2026-08-06_render_audit.txt`. Nine
unknown. Enumerate them with the bug file's §4 query filtered to
`spec->>'audit_source'='design-audit' AND status='complete'`.

Grade each against **its own `spec.acceptance_test`** at the served artefact (browser /
computed styles — the 122 lane's render-audit machinery). **Do NOT re-run the verifier**:
it will pass again, for the same correct reason (bug §6).

For each confirmed-unrepaired, **insert a fresh row** (`item_type='dark_section_audit'`,
`status='detected'`, spec copied plus `reopened_from=<old id>` and
`reopen_reason='bugs_open/213 false complete'`, item_key recomputed under the new type so
dedup and two-strike apply). **Leave the historical rows' status alone** — `complete` is
the honest record of what the machine did.

**Acceptance measure for the whole fix:** the 11-vs-0 asymmetry must become capable of
disappearing — after one post-roll audit cycle, a producer-B item must be able to reach
`unresolved`/`failed`/`detected`. Until one does, the fix is deployed but unexercised,
and those are different claims.

### 4. Two things deliberately left open, on the record
- **`designRouting["dark_section"]` still points at `color-variable-fixer`**, whose
  transform provably cannot touch producer B's typical defect (an already-`var()`
  fallback — that is *why* gamesdesign's item passed). That is a routing decision for the
  design-audit route owner, not this bug. Named in WII-013 so it is not lost.
- **`spec.acceptance_test` is still read by nothing** on the completion path (grepped:
  zero consumers outside the `improve_tool` family). The candidate verifier for
  `dark_section_audit` is `criteria_check` (RFC_002) over that field. That is the
  follow-on that would turn a declared gap into real verification.

## Traps that cost me time today

1. **Every ownership check reads COMMITS.** `who-owns.py`, `git log`, "no workstream dir
   exists", and the session-start `git status` snapshot all share one blindness. Run
   `git status --short <the bug's cited paths>` **first**. An untracked new file beside a
   bug's paths is the strongest ownership signal there is. This cost ~1h on `bugs_open/214`
   (full account in `WRONG_CALLS.md`, distilled into LANDMINES).
2. **Mutate one half at a time.** Reverting Half A alone left the guard GREEN, because
   Half B independently covers the route. Had I mutated both at once and stopped, I would
   have claimed "the test guards the fix" when it guards *at least one half of it*.
3. **`git archive HEAD` must be given the repo** — `git -C <repo> archive`, not a bare
   `git archive` after `cd`-ing into the scratchpad, which silently leaves you with an
   almost-empty directory that then "builds fine".
4. **Council schema**: `.plan.summary` is required and is a different field from
   `rationale`; `operation` must be `modify|add|remove|config_change` (`create` is
   rejected — a new file is `add`). The `commit-msg` trailer gate correctly blocks
   `Council-Submitted: pending` — submit first, commit with the real correlation.
5. **`site_plan_pages` vs `site_plan_sections`** (from the 214 detour, but general): a
   page with no sections looks like a missing page. Resolve page existence against the
   table the *consumer* reads.

## The 214 debt I incurred and discharged

I investigated `bugs_open/214` for an hour before discovering its owner mid-fix, and stood
down. I contributed the measurements rather than competing — appended to that bug file
2026-08-10. **If anyone picks 214 up: the filed census is section-scope only and
understates it roughly fourfold.** Page scope has 28 unresolvable refs of 162, its
consumer join is `scope_ref = $page` *exactly* (not the section join's tolerant `LIKE`),
19 of 22 current-plan orphans have active generated assets, and gamesdesign.co.uk's about
page serves two `<img src="/assets/images/hero.jpg">` that **404** while the commissioned
`hero-about.jpg` sits deployed at 202,259 B. That is not this lane's work — do not adopt
it here.
