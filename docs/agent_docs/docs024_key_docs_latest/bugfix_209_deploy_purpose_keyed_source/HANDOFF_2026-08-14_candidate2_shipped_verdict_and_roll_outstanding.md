# HANDOFF — 2026-08-14. Candidate 2 SHIPPED (`d3edb5b89`), inert until a roll. **Two things outstanding, both cheap: read the verdict, prove it post-roll.** COLD-START HERE.

Supersedes `HANDOFF_2026-08-11b_census_done_owner_decisions.md`. Its Task A and
Task B are both done — **Task A by another lane, by a different route** (see §2).
Evidence and missteps: `NOTES_209…` (08-13 evening → 08-14 morning section). The
bug's own account, which is the fuller one: `bugs_open/231` §"CANDIDATE 2 SHIPPED".
Seam registered as **CTS-059**. Trap written up in `LANDMINES.md`.

## 1. What shipped

**`d3edb5b89`** (2026-08-14 08:44 BST) — bugs_open/231 candidate 2, owner ruling
2026-08-11 #2. **Go change: INERT until the next chassis image roll.** No DB config
touched. `Council-Submitted: 41a01378-1211-4987-966d-f8b6e2fddce1`.

- **Strategy 6** in `datahelpers/action_inputs.go`: for a field still holding only
  its `spec.Defaults` value, an explicit **dotless config scalar of the Default's
  kind** becomes the value, and `Defaulted` provenance is cleared. Dotted strings
  stay dead (a dot means REFERENCE — `bugs_open/248` finding (a)); composites stay
  refused; a kind guard refuses a scalar whose type differs from the Default's; an
  explicit `""` cannot override a **Required** field's Default.
- **Strategy 3 deprecated bridge** now beats a Default when its path resolves.
  Zero live definitions carry a Deprecated alias for a defaulted field, so this arm
  is correctness for the next author, not a live change.
- **Detector re-specified in the same commit** (`cmd/config-key-audit/defaultshadow.go`):
  `static_string` + `non_string_literal` → `live_override`; `type_mismatch` and
  `required_empty_string` are the new dead classes; `deprecated_bridge` → conditional.
  The binary emits a per-finding **`verdict`** and the wrapper groups on it.
- **Two 231-pinned tests flipped**, both now also asserting cleared provenance.

Measured before → after on the same live export: **21 dead-mismatched + 78
dead-matching → 0 dead**, 96 conditional unchanged, **99 live overrides**, exit
1 → 0. Demand control (one live `max_pages: 60` mutated to `"60"`) → exit 1, 1 dead,
`type_mismatch`. The zero is real, not a blinded detector.

## 2. Task A was done by another lane, by a different mechanism — do not redo it

Ruling 1a described a four-line direct config read in `write_audit_findings`. What
is live is `bugs_open/264` / commit `3621ca7cf`: **migration 399** gives each of the
four auditors a resolvable `audit_source`, and the Go spec makes `audit_source`
**Required with no Default**, so a fifth author repeating the mistake fails loudly.
Different route, same outcome, arguably stronger. It is also why candidate 2's
activation set was **empty by construction** rather than by luck.

## 3. Outstanding (in order)

1. **READ THE COUNCIL VERDICT** — `41a01378-1211-4987-966d-f8b6e2fddce1`. The code
   is already on the shared branch, so a REVISE/REJECTED needs acting on, not
   filing. **Do NOT add a `Council-Reviewed:` trailer** — `098` credits the
   correlation automatically once approved, and writing that trailer on an unread
   verdict is the report's dishonesty surface.
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='41a01378-1211-4987-966d-f8b6e2fddce1' AND kind='council_report' ORDER BY created_at;
   ```
   Budget ~30 min from 08:47 BST for dispatch latency; a missing orchestration row
   is latency, not a dropped dispatch — find it by payload, never retry on that
   evidence. Verdict prose: `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;`
2. **POST-ROLL PROOF**, once a chassis image carrying `d3edb5b89` ships. Three checks:
   - the service's own stamp is an ancestor-or-equal of the fix:
     `git merge-base --is-ancestor d3edb5b89 <the stamp>` (read the stamp **per
     service**, not per fleet — `bugs_open/249`);
   - a step carrying a dotless static on a defaulted field logs
     `Strategy 6: explicit config value beat the spec default`;
   - `./scripts/audit-default-shadowed-keys.sh` still exits 0 with **0 dead**.
   **Expect NO `Strategy 6: config value's type differs` Warn anywhere** — no live
   entry mismatches kinds today, so one appearing means new config arrived and is
   being silently ignored. That Warn is the class's early-warning line.
3. **231's remaining half is the 96 `dotted_conditional` entries** — a dotted path
   that fails to resolve still falls to its Default silently. Open by design
   (resolvability is a runtime fact an offline check cannot decide). The one latent
   row is `derive_card_asset entity_type`, benign until phases I5/I6.
4. **CTS-059's open review question**, if anyone wants it: whether a *resolving*
   dotless string on a defaulted field should resolve as a `collected_data`
   reference instead of being taken literally. Deliberately not taken — it is the
   one arm that could replace a typed Default with an object of unknown shape, and
   zero live entries want it. Whoever takes it owns re-measuring the `*_field`
   family first (28 read config directly today, which is what makes it look free).
5. **240 — the sweep half is DONE and PROVEN; only C2/C1 remain.** The previous
   handoff was still waiting for "the first real APPLY run"; `[MEASURED 2026-08-14
   08:50 BST]` from `tail ~/kafka-sweep-240.log` it has run **twice** in APPLY mode
   since the KUBECONFIG fix, both clean: one deleting 1,414 orphaned topics
   (`fail=0`, 1,064 remaining) and the `2026-08-13T11:17:01Z` run deleting 1,096
   (`fail=0`), taking `job.*` topics **1,664 → 570**. Protection held both times
   (568 KEEP against 593 protected correlations, 6-hour window).
   ⚠ **The 00:17 local slot on 08-14 did NOT run** — no log entry — which is the
   already-recorded trap that user crontabs get no anacron catch-up, so a slot the
   machine sleeps through is silently missed. Do not read the gap as a failure.
   Remaining on 240: the **C2 safe subset** (scheduler-scoped transport) and the
   **C1 question**. Neither is started.
6. 209 Phase 3 (retire dead writers) and 236 — open, unowned by this thread.

## 4. Cold-start checks

1. `go test ./platform/orchestration/datahelpers/ ./cmd/config-key-audit/` — green.
   **`./platform/orchestration/actions/` may not COMPILE**, and that is usually not
   yours: on 08-14 another session had `palette_specialised_slots.go:387 undefined:
   colour` mid-edit, and `internal/adapters/thunder/api` is separately broken at
   HEAD (`unknown field Identifier`). Test against `git archive HEAD` with your own
   files overlaid, and prove a pre-existing failure against a **pristine** archive
   before blaming or fixing it.
2. `./scripts/audit-default-shadowed-keys.sh` — expect **exit 0, 0 dead, 96
   conditional, 99 live overrides**. Anything in the DEAD section is new config
   arriving in a class the resolver refuses; investigate before anything else.
   ⚠ **never `... | tail; echo $?`** — that reads `tail`'s status. Redirect first.
3. Deploy state: read the stamp of the service you actually mean; do not assume a
   fleet-wide revision (`bugs_open/249`).

## 5. Traps this lane hit, worth carrying

- **`cmd/config-key-audit/` is still CONTENDED.** The RFC_022 lane's
  `optionalbudget.go` + `optionalbudget_test.go` + 13 lines in `main.go` are
  untracked/uncommitted as of this commit. I touched only `defaultshadow*.go` and
  **deliberately left `main.go` out of my pathspec** — committing their 13 lines
  without their untracked `optionalbudget.go` would break HEAD's build (a reference
  to a function that does not exist at HEAD). Check `git diff cmd/config-key-audit/main.go`
  before you touch it.
- **My commit carries one declared passenger:** `000_concept_index.md` also holds
  their `WFA-013` row. A pathspec commit cannot exclude a same-file edit; it is
  named in the commit message. Their `workflow-authoring.md` entry was still
  uncommitted, so that row referenced an absent entry at HEAD until they committed.
- **A `--default-shadowed-keys` report written before 2026-08-14 describes the
  opposite of production** for 99 of its 195 findings. Read `verdict`, never the
  class name.
- **A long session's sense of "today" is stale evidence.** This one spanned
  midnight and dated a whole day's measurements to 08-13, including inside an
  immutable commit message. Take the date from `git log`/`date`, not the context
  banner (NOTES has the full note).
