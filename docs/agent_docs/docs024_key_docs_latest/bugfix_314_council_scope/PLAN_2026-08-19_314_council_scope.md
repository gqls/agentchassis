# PLAN 2026-08-19 — bugs_open/314: the council gate cannot see that a migration is config

**Lane:** `bugfix_314_council_scope` · **Bug:** `bugs_open/314` · **Session:** named
`bugs_open/taken 313 and 298 and 314`, owner-directed 2026-08-19 ("prepare a plan to fix this bug
preferring a robust solution at all times that is applicable to the framework as a whole in
preference to the individual case").

## What we are fixing

`097_TRIGGER_council_review_v1.sh` refused any submission whose edits all sat under `docs/` —
including `docs/agent_docs/sql_for_agents/NNN_name.sql`, which is where this estate's live agent
config ships. The 2026-07-17 owner ruling it quotes is about **subject matter** (prose does not
spend council credits); it was implemented as a **path test**. A migration is not prose: it
rewrites what a live agent does, is live the moment it applies, and has no image tag to roll back —
so it reaches production faster than any Go change, and it is the half nobody reviewed.
**Measured in the bug: 152 of 227 (67%)** migration-shipping commits over a fortnight were
config-only, and so unreviewable without `FORCE=1`.

## Owner decisions taken this session

- **Fix candidate 1, implemented robustly** (not the one-line regex widening the bug sketched).
- **No new `090` diagnosis-loop run.** Substitution stated plainly, per the 2026-07-31 ruling: the
  mechanism is four lines of shell read at source (`SCOPE_RE` at :87, the jq test and `exit 2` at
  :146-153), the bug file's §7 already records the same substitution for the filing, and the
  population figure is one `git log` loop any reader can re-run. Nothing structural is asserted
  here that a loop would test. The council round below reviews the fix itself.

## The framework finding that changed the shape of the fix

The bug names one file. There were **three** hand-maintained copies of the scope:

| consumer | was | role | failure semantics it needs |
|---|---|---|---|
| `097_TRIGGER_council_review_v1.sh:87` | `SCOPE_RE='^(platform\|internal\|pkg)/'` | ADMISSION gate | fail **LOUD** — with no scope it must neither admit prose nor lock everyone out |
| `scripts/council-coverage-nudge.sh:58` | inline `grep -E` | commit-msg nudge | fail **SILENT** — advisory, must never block a commit |
| `098_REPORT_unreviewed_commits_v1.sh:76` | `SCOPE_PATHS=(platform internal pkg)` | coverage report | fail **LOUD** — a report against an unknown scope is worse than none |

Widening one and leaving two is how a rule becomes folklore: the gate would admit migrations while
the nudge stayed silent on them and the report kept mis-bucketing them. So the fix is **one shared
definition, three callers** — `scripts/council-scope.sh`.

## Why a shared fragment and not a lockstep drift-check

The estate's own doctrine decides it: `pattern-check.py`'s `check_register_coverage` docstring
("Two hand-maintained copies of one matching rule is precisely the drift class this platform keeps
filing bugs about… There is one implementation; this calls it"), the CLAUDE.md `099` roster-mirror
trap, and the standing dedup-index/Go-list lockstep. A drift check would bless four copies and add a
fifth artefact. All three consumers are bash and already resolve (or can resolve) the repo root, so
sourcing is cheap. **The differing failure semantics live at each call site**, which is the only
place they can — see the table above.

**One copy could not be absorbed, deliberately.** The migration vocabulary belongs to
`scripts/migration/run-migrations.sh` (`SIDECAR_RE` at :65, the appliable-name grep at :283).
Sourcing the fragment *from the runner* was considered and **rejected**: it would couple the review
gate to the APPLY path, so a syntax error in the review gate's fragment would stop migrations being
applied — a bigger blast radius than the bug. Instead the fragment carries verbatim copies held
honest by `council_scope_drift_warn()`, a `grep -qF` verbatim-anchor check (the estate's own idiom —
the `381`+`383` worked pair) that WARNs on every 097/098 run and can never block one.

## The definition

In scope = `^(platform|internal|pkg)/` **OR** (`^docs/agent_docs/sql_for_agents/[0-9]{3}_[A-Za-z0-9_]+\.sql$`
**AND NOT** `_[A-Z][A-Z0-9_]*\.sql$`).

Two arms ORed, and the migration arm is match-then-reject-sidecar rather than one clever regex —
which is the runner's own idiom, because a trailing-`_TOKEN` negative class is unwritable in ERE.
Keeping the arms separate also means a `platform/` path with an uppercase-suffixed `.sql` cannot be
knocked out by the sidecar rule. The `[0-9]{3}` anchored straight after `sql_for_agents/` also
excludes the `sql_for_agents_v1`/`_v2` archive trees for free.

## Cost — the caveat §5 asked someone to price before this went in

- Council-gate volume, measured 2026-08-19 from `llm_call_log`: **11–43 runs/day** over 9 days
  (median ~24–33).
- Worst case if every config-only commit became a round: +152/fortnight ≈ **+11/day**. Realistic,
  at one round per coherent task spanning several commits: **+4–6/day, i.e. +15–25%**.
- Two mitigations already live: **prompt caching since v1.0.1283, measured 74% saving**
  (`council_gate_cost` lane), and relevance gating — only footprint-matching seats fire beyond the
  two always-on ones (edit-quality, guardian).
- **Tripwire, stated rather than assumed:** if sustained gate volume exceeds **150% of the 9-day
  median for a fortnight**, raise fix candidate 2 (a cheaper config-specific seat set) as a
  follow-up — do **not** revert admission. Candidate 2 is weakened but not refuted by §9's finding
  that `FORCE=1` did not degrade the review: the full roster reviews config well, so the case for a
  separate roster is cost, not quality.

## Why not the other candidates

- **2 (separate config roster):** §9 measured that all ~16 seats reviewed forced config rounds
  substantively — "it only needs to change *admission*, not the review itself". Building a second
  roster is work the evidence says is unnecessary today. Kept as the costed fallback above.
- **3 (relocate migrations out of `docs/`):** touches the runner, `MIGRATIONS_DIR`, every runbook
  and every session's muscle memory on a shared tree — high disruption for the same admission
  outcome. And §9 records that no seat in three rounds objected to a migration living under
  `docs/`; the mismatch is entirely in the pre-filter.
- **4 (status quo):** gives up the ability to tell a skipped review from an impossible one, which
  is the audit hole the bug is about.

## RFC question (owner ruling 2026-07-29 §1), answered

**No RFC.** The trigger is a change to what the shared mechanism *guarantees*. The gate's guarantees
to its consumers are untouched: an admitted submission still gets a full round, verdict artefacts
still key on the correlation, the trailer and 098 join semantics are unchanged. What changes is
**admission policy** — which submissions may spend credits. Ruling §3 ("other consumers must be
told, not merely measured") is discharged by the CLAUDE.md, RUNBOOK and LANDMINES edits: those *are*
the telling.

## Deliberately not done

- `scripts/pattern-check.py:1248`'s docstring quotes the old `SCOPE_RE` while arguing that its grep
  exists *because* docs are refused on cost grounds. The widening admits only `NNN_*.sql`
  migrations, never `.md` design docs, so its premise survives intact. No edit.
- **A fourth copy of the migration vocabulary exists** and is out of this fix's remit:
  `pattern-check.py:384-385` (`MIGRATION_DIR`, `MIGRATION_NAME = ^\d{3}_[a-z0-9_]+\.sql$`, used by
  `check_unguarded_migration_insert`). It is not a scope consumer — it is an idempotency lint — but
  it is **already drifted from the runner**: lowercase-only, so a runner-valid migration with an
  interior capital is invisible to that lint. Recorded here and in the close-out as a follow-up.
- Tooling (`scripts/`, and the 097/098 scripts themselves) is **not** admitted to scope. That is why
  this fix's own submission needs `FORCE=1`.
