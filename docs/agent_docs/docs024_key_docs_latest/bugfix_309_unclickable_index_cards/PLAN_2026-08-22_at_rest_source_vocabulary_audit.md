# PLAN — 2026-08-22 — the at-rest source-vocabulary audit (bugs_open/309's last half)

**Phase 2 of this lane.** Phase 1 (`PLAN_2026-08-18`) shipped the BIRTH gate and the case
repair; both are live and both were re-verified today before any of this was designed
(NOTES, 2026-08-22). This plan builds the half the concept register has been naming as
owed since the day the gate shipped.

## What we are fixing, in one sentence

`sourceVocabularyIssues` refuses a phantom data source when a component is BORN, and
nothing has ever asked the same question of the **285 active components already in the
database** (`[MEASURED 2026-08-22]`; 184 of them declare an `input_schema.fields` object
at all) — where **69 fields across 17 components** declare a source that resolves
nowhere, 46 live page instances deep.

## Why this is the framework-wide fix and not the individual one

The individual case (fundamentallyai.com's article index) was repaired by migration 478
and is still serving 8 linked cards today. The *class* is not the six orphaned articles;
it is the shape: **a declared source that resolves nowhere produces markup that is
complete and data-less, and is indistinguishable from success at every stage.** The birth
gate closes one door into that state. This closes the room.

Three facts make the at-rest sweep the load-bearing half rather than a tidy-up:

1. **The birth gate cannot ever see the other write paths.** A component is routinely
   inserted or altered by a hand-written migration or by hand SQL, which never passes
   through `store_generated_component_action`. Already a recorded LANDMINE (85dbf889d),
   and it is exactly how the motivating component got there.
2. **The one runtime detector of this silence is write-only and cannot see dormant
   components.** `STRUCTURAL_KEY_CARRY_MISS` records precisely this omission at render
   time — `[MEASURED 2026-08-22]` **28 rows all-history**, first 2026-08-11, **last
   2026-08-17**, i.e. nothing written for five days.

   > **CORRECTED 2026-08-22 (later, same day).** This paragraph originally called it one
   > of the codes `bugs_open/358` measures as having **no automated consumer**. That is
   > **no longer true and was already false when written**: `cmd/content-loss-check`
   > consumes it as of `cba51ad1d`, committed that morning by the `bugfix_238` lane, and
   > `[MEASURED]` **8 of its 28 rows are now `resolved`**. Caught by the `358` lane and
   > **re-verified here at the source and the table before accepting it** — a peer's
   > report is another doc. The mistake was inheriting a census from another lane's file
   > instead of running its query: `358` was filed the same day and its consumer half went
   > stale within hours. Logged in `WRONG_CALLS.md`.

   **The decision not to route findings into `agent_error_log` stands, on the limit the
   correction does not touch:** that writer only fires when a page is **BUILT**, so it can
   never see a component that is never built — the **eleven dormant components in this
   audit's own baseline are permanently outside its reach**, however faithfully its rows
   are now consumed. A disposition says who READS a code; it says nothing about whether
   the WRITER can see the population it exists to catch. A scheduled check also needs a
   carrier that speaks on a CLEAN run, which is what the `doc_notes` convention is for.
3. **The register already instructs this shape.** CLC-018
   (`docs026_concept_register/register/component-lifecycle.md:252`): *"`sourceVocabularyIssues`
   (pure, reusable — a future daily audit of EXISTING config should call IT)"* and
   *"build it ON `sourceVocabularyIssues`, not on a second predicate, or they drift."*

## Decision 1 — the carrier: a Go-native check image, calling the guard's own function

**Chosen: a new `--component-source-vocabulary` mode in `cmd/config-key-audit`, shipped in
a purpose-built check image on a daily CronJob.**

The competing shape is the Python-script-in-`postgres:16-alpine` mirror
(`component-fallback-check`, `optional-key-budget-check`, `single-owner-carriers-check`),
pinned against the Go rule by a shared fixture and a parity test. It was rejected, and the
reason is not taste:

> **A Python mirror IS a second predicate.** It can only ever *detect* drift; calling the
> function makes drift **unrepresentable**. This estate ranks fix candidates by what makes
> the bad state unrepresentable, and the register's own entry asked for the call, not a
> copy.

The "owner ruled for Python" reading that nearly sent this the other way is real but
**local to 2026-08-14 and superseded by practice**: `build/docker/backend/` now holds
**ten** two-stage Go check images against **three** Python checks, and the most recent
check added to this tree (`capped-schedule-ordering-check`, `c81b73b9e`, this week) is
Go-native, `CMD ["./config-key-audit", "--capped-schedule-ordering", "--report"]`. The
objection the owner was answering — *"a git clone of a 262M repo + go mod download + a
compile, in a job with uncertain egress"* — is about compiling **inside the job**, which
a pre-built image does not do.

**Machinery reused rather than rebuilt:** `dbConn()` and `writeDocNote()` from
`cmd/config-key-audit/fleetdb.go`; the mode/exit-code/refusal discipline of
`emitCommitShaExposure`; the acks-file loader shape of `loadCommitShaExposureAcks`; the
dockerfile, makefile targets and kustomize base+overlay of the capped-schedule sibling.

**Exports required** (a pure rename inside `package actions`; call-site behaviour
identical, so the birth gate cannot change): `sourceVocabularyIssues` →
`SourceVocabularyIssues`, `loadKnownSpecAspects` → `LoadKnownSpecAspects`. `main.go`'s
blank side-effect import of the package becomes a named one.

## Decision 2 — the two vacuity refusals, which are NOT symmetrical

Exit **2** means *the check did not run*, and must never read as a pass. Two conditions,
and they fail in opposite directions:

- **Zero active components decoded** → the classic silence: "0 findings" would read clean.
- **Zero `site_specs` aspects** → the opposite, a **flood**: an empty aspect set marks
  every `site_specs.*` source in the estate as phantom. Equally wrong and equally a broken
  read, because `SELECT DISTINCT aspect FROM site_specs` returning nothing never means the
  estate has no aspects.

**And one deliberate divergence from the birth gate, stated rather than smuggled:** the
gate FAILS OPEN when the aspect set is unreadable (it must not block all component
generation; it records `SOURCE_GUARD_ASPECT_SET_UNAVAILABLE` durably — `[MEASURED]` **0
rows all-history**, so the read has never failed). The audit **fails closed, exit 2**. An
audit has nothing to block, and a daily report that silently skipped a third of its rule
is the blind-pass landmine one rung higher — the trap this lane's sibling shipped and had
to correct on 2026-08-21 (`85c70c24a`).

## Decision 3 — the ratchet: a frozen, shrink-only baseline keyed on the exact finding

69 pre-existing findings cannot be zeroed in one commit — each is a per-component
judgement, and repairing *one* of them (migration 478) took an owner ruling and a
shrink-guard incident. A job that is red from day one trains people to ignore it. So:

`component_source_baseline.json` — **generated mechanically from the census query, never
hand-typed** — one row per finding:

```json
{ "component_id": "<uuid>", "component": "info-card-grid", "field": "carousel",
  "source": "config", "class": "prefix_outside_vocabulary",
  "live_instances_at_baseline": 32, "baselined": "2026-08-22",
  "route": "bugs_open/362_…" }
```

A live finding is grandfathered **iff the exact 4-tuple `(component_id, field, source,
class)` matches.** Four red conditions, none of which is a standing backlog:

1. **A finding not in the baseline** — a new component, a new field, a *changed* source
   string, or a previously-clean field going bad. This is the job's purpose: the
   migration/hand-SQL door the birth gate cannot watch.
2. **A component dormant at baseline (`live_instances_at_baseline = 0`) now has a live
   instance.** Grandfathering the eleven dormant components is **conditional on their
   staying dormant**: deploying one is a new page acquiring a known silent field-drop.
   Growth on the six already-live ones (info-card-grid 32→33) does *not* go red — that is
   more of already-routed damage, and red-on-every-instance is how a check gets ignored —
   but current-vs-baseline counts are reported every run.
3. **A stale baseline entry** matching nothing live (repaired or deactivated). The message
   names the exact line to delete. **This is the ratchet's pawl:** the file shrinks
   mechanically as repairs land and can never accumulate dead entries that later mask a
   re-offence.
4. **The guard's own unit suite failing** — reported as "do not trust today's result".

### Why this is not the allow-list-silences-your-own-detector trap

That trap is on this lane's own record (a memory landmine, and `scripts/pattern-check.py`'s
`COMPONENT_WRITE_ALLOWED` note about converting a live debt into a false all-clear). Three
structural properties, not discipline:

- **Narrowest possible key.** An entry matches exactly one frozen finding. It cannot
  swallow a future offender the way a name- or class-keyed allow-list would — and the
  demand control below proves that on the motivating case itself.
- **Append-closed, enforced by test.** A unit test asserts every entry's `baselined` date
  is `2026-08-22` and fails with *"the baseline is closed — new findings are fixed, not
  baselined"* on any other value. Growing it means falsifying a date in a diff-visible
  file or amending the test: both deliberate, visible, and the owner's call, not a
  session's.
- **Grandfathered is not silent.** The daily `doc_notes` row reports the whole
  grandfathered population every run — count, class split, live-vs-dormant, live findings
  listed first as "repair owed". **The baseline governs the EXIT CODE and never the
  REPORT.** A detector that still names all 69 every morning cannot go quiet.

## Decision 4 — schedule: a slot that is free in the REPO, not merely in the cluster

`kubectl get cronjobs` is the wrong instrument: `capped-schedule-ordering-check` is
committed at `"5 7 * * *"` and **not yet deployed**, so the cluster does not show it and
07:05 looks free while it is claimed (and already collides with `content-loss-check`).
Enumerated from the repo instead, genuinely free: **07:00, 07:20, 07:30, 07:35, 07:45,
07:50.** Top-of-hour is avoided — hourly jobs fire there and other lanes rebuild pages
hourly, so a config check should not measure mid-rebuild state.

## Decision 5 — what this deliberately does NOT do

- **The 69 repairs.** Each is a per-component judgement. They are *routed*, not silently
  deferred: a new `bugs_open/` file is the owning ledger, every baseline row points at it
  by path, and every repair forces a visible baseline trim — **the file's shrink history
  IS the burn-down.**
- **The eleven dormant components are not deactivated.** Deactivation is itself a repair
  judgement; some are awaiting regeneration, which the birth gate will now force to
  repoint sources. Red condition 2 means leaving them costs nothing silently.
- **No work-item production and no auto-repair.** Detection→dispatch is the estate's
  known-broken half (`bugs_open/083`; `bugs_open/358` for this exact channel). A red job
  plus a daily `doc_notes` row is the carrier this estate actually reads.
- **No widening of the rule.** Nested `items` sub-schemas stay out of scope on both sides,
  inherited from the guard's own stated scope.
