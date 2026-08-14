# HANDOFF — 2026-08-14b. Candidate 2 **LIVE on v1.0.1298**, both replicas, stamp-proven. **Round 2 with the council. ONE behavioural check still owed.** COLD-START HERE.

Supersedes `HANDOFF_2026-08-14_candidate2_shipped_verdict_and_roll_outstanding.md`
(that one's §3 items 1–2 are now done/superseded; its §5 traps still hold).
Full evidence: `bugs_open/231` §POST-ROLL. Missteps: `NOTES_209…` (08-14 afternoon).
Owner-facing prose: `README_where_we_are.md`. Architecture question: **RFC_028**.

## 1. State in one table

| thing | state |
|---|---|
| Seam (Strategy 6) | **LIVE**, chassis `v1.0.1298`, both replicas, artefact-proven |
| Council round 1 | **REVISE** (corr `41a01378`, 11 seats, 6 abstained, no truncation, gated by `guardian`) |
| Round 1's 4 code objections | **CLOSED** in `14e4333f7` |
| Council round 2 | **SUBMITTED**, same corr `41a01378` — **verdict NOT read** |
| Architecture `needs_rfc` | **ROUTED**, not argued down → `RFC_028` (`bc39e7bf5`) |
| Detector, live | **0 dead**, 96 conditional, 99 live overrides, exit 0 (185 agents) |
| Behavioural proof that Strategy 6 FIRES | **⚠ NOT DONE — the one thing owed** |
| `--report` mode | built, **UNDRIVEN** (no CronJob; deliberate, see §4) |

Commits: `d3edb5b89` seam · `14e4333f7` revise round · `bc39e7bf5` RFC_028 ·
`01f983411`/`df1ed4f94`/`e62068807` docs.

## 2. THE ONE CHECK OWED — and why the obvious way to do it lies

**Observe `Strategy 6: explicit config value beat the spec default` actually firing.**

Do **NOT** conclude anything from `kubectl logs --tail=N`. `[MEASURED 2026-08-14]` a
chassis pod started 08:58:03Z retained **243 lines spanning 92 SECONDS**
(13:51:23Z → 13:52:55Z). The Strategy 6 line was absent — and that reading is
**worthless**, because the control says so: **Strategy 0's pre-existing Info line is
absent from the same window too.** 241 of the 243 lines are `level:info`, so it is not
a level filter. The absence measures retention, not the resolver.

Two ways that do work:

```bash
# (a) stream both replicas and wait for real traffic to hit one of the 99 live overrides
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  kubectl -n ai-persona-system logs -f "${p#pod/}" --since=1s 2>/dev/null \
    | grep --line-buffered -E "Strategy 6: explicit config value beat|Strategy 0: Resolved config path" &
done
```
Include the **Strategy 0** line in the same grep. It is the liveness control: if
neither appears, you have learnt nothing; if Strategy 0 appears and Strategy 6 does
not over sustained traffic, *that* is a finding.

```bash
# (b) pick a live override and drive the agent that carries it
./scripts/audit-default-shadowed-keys.sh 2>/dev/null | sed -n '/=== LIVE/,$p' | head -20
```

**Also expect NO `Strategy 6: config value's type differs` Warn anywhere.** No live
entry mismatches kinds today, so one appearing means new config arrived and is being
silently refused. But it inherits the same retention blindness — **do not report its
absence as a pass on its own.**

## 3. Read the round-2 verdict

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='41a01378-1211-4987-966d-f8b6e2fddce1' AND kind='council_report'
ORDER BY created_at;   -- round 1 = revise; a SECOND row is round 2
```
Prose: `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;`
Progress: `SELECT current_step, status FROM orchestration_states WHERE collected_data->'input_data'->>'fix_correlation_id'='41a01378-1211-4987-966d-f8b6e2fddce1';`

**If APPROVED:** add `Council-Reviewed: 41a01378-1211-4987-966d-f8b6e2fddce1` to your
next commit in this lane. **Never write that trailer on an unread verdict.**
**If REVISE again:** round 1 was worth it — four real defects — so revise, do not
defend. Resubmit with `RESUBMIT_CORR=41a01378-…`.

Round 2's submission JSON is kept at
`scratchpad/council_231_round2.json` (session `4de3e004`) — if that scratchpad is gone,
rebuild from the commit bodies; nothing is lost.

## 4. `--report` is built and UNDRIVEN — the ordering is on whoever wires it

`config-key-audit --default-shadowed-keys --report` reads the fleet straight from
Postgres and writes one `doc_notes` row per run, clean or not. It exists because the
`bug_historian` seat objected that Strategy 6's three rejection arms report only
through zap — an objection §2 above then confirmed empirically.

**No CronJob runs it, and I deliberately did not ship the overlay.** The image must
exist BEFORE the overlay is applied, because this fleet reports `ImagePullBackOff` as
a Job still **RUNNING**, never FAILED — the trap `removed-config-keys-check`'s own
`cronjob.yaml` records hitting on its first rollout. To wire it, copy that service
line-for-line: `build-`/`push-` makefile pair (see makefile ~199–208), `base/cronjob.yaml`,
`overlays/production/uk_001/`. Schedule in the config-integrity window (06:20
single-owner, 06:25 removed-keys, 06:35 discovery-staleness — so 06:30).

## 5. What else is open on 231

- **The 96 `dotted_conditional` entries** — a dotted path that fails to resolve still
  falls to its Default silently. Open **by design**: resolvability is a runtime fact an
  offline check cannot decide. The one latent row is `derive_card_asset entity_type`,
  benign until phases I5/I6.
- **CTS-059's open review question** — whether a *resolving* dotless string on a
  defaulted field should resolve as a `collected_data` reference rather than being taken
  literally. Zero live entries want it; it could replace a typed Default with an object
  of unknown shape. Whoever takes it owns re-measuring the `*_field` family first.
- **RFC_028's three asks for the owner**: does the precedence chain get an owner; should
  the dot-discriminator be one named predicate instead of two prose copies; is there an
  arm budget in RFC_022's sense. Measured there: **27 rounds, 8 `needs_rfc`, 1 veto.**

## 6. Not this lane's, unchanged

- **240**: sweep half DONE and proven (two clean APPLY runs, `job.*` 1,664 → 570).
  Remaining: C2 safe subset + the C1 question. A gap in
  `~/kafka-sweep-240.log` is a slept cron slot, not a failure.
- 209 Phase 3 (retire dead writers) and 236 — open, unowned by this thread.

## 7. Traps carried forward

- **`cmd/config-key-audit/main.go` is CONTENDED** (the RFC_022 lane's 13 lines +
  untracked `optionalbudget*.go`). I never touched it — `--report` parses `os.Args[2:]`
  inside `defaultshadow.go`, exactly as `emitRemovedKeyCarriers` does. Committing
  `main.go` without their untracked file would break HEAD's build.
- **A `--default-shadowed-keys` report written before 2026-08-14 describes the opposite
  of production** for 99 of its 195 findings. Read `verdict`, never the class name.
- **`| tail` then `echo $?` reads `tail`'s status**, not the script's. Cost me a wrong
  "exit 0" on the first cold-start check.
- **A second-hand deploy fact is still a claim.** Another session had already recorded
  `v1.0.1298 stamped bc39e7bf5` (`8dd925576`); I re-probed both replicas with a
  two-sided control anyway. It was right — but the negative arm is what makes the
  positive one mean anything, because a discovery grep for "some 40-hex string" matches
  Go's internal digit table.
- **A long session's sense of "today" is stale evidence** — this lane backdated a whole
  day's measurements once already (`WRONG_CALLS.md` 2026-08-14).
