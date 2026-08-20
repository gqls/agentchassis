# 314 — the council gate cannot review a config-only change, because its scope check is a PATH regex and this estate's config ships as SQL files under `docs/`

> ## ✅ CLOSED 2026-08-20 — FIXED, LIVE, and the fix's own defect found by the gate it widens
>
> **Fixed by the `bugfix_314_council_scope` lane** (session `bugs_open/taken 313 and 298 and 314`,
> owner-directed). Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_314_council_scope/`.
> **Owner chose fix candidate 1, implemented robustly.** Commits `49ad608a2` (the fix) and
> `738fcbb96` (the correction below). **Council `85fac99c-c947-46e8-afcc-c2b03568cc24`: APPROVED,
> round 1, 12 reviewers, none high-severity.**
>
> **Live on commit** — these are shell scripts read from the tree at invocation, so there is no
> image, tag or roll between the commit and the behaviour. (`debug_historian` made exactly this
> point in the round: the "verify against the running pod" discipline does not transfer here, and
> the control matrix is the correct analogue.)
>
> ### What shipped — and it is not the one-line widening §5 sketched
>
> §5 candidate 1 named one file. There were **three** hand-maintained copies of the scope:
> `097:87` (admission), `scripts/council-coverage-nudge.sh:58` (the commit-msg nudge), `098:76`
> (this report's own pathspec). Widening one would have left the gate admitting migrations while
> the nudge stayed silent on them and the report mis-bucketed them. So: **one definition,
> `scripts/council-scope.sh`, three callers** — with deliberately OPPOSITE failure semantics at
> each call site, because a missing definition means something different to each: **097 fails
> LOUD** (exit 1, *not* the refusal's exit 2 — "I cannot tell" must not be reported as a scope
> decision), **the nudge fails SILENT** (exit 0, zero bytes — it runs from `commit-msg` on every
> commit in every session, so a bug there could block the fleet committing), **098 fails LOUD**.
> All three branches were *induced* in a scratch repo, not reasoned about.
>
> The runner keeps its own vocabulary on purpose: sourcing the fragment FROM
> `run-migrations.sh` would couple the review gate to the APPLY path, so a syntax error in the
> gate would stop migrations being applied — a bigger blast radius than this bug.
>
> **`DRY_RUN=1` added to 097**, after all validation and the scope decision but before any
> correlation is minted. §6's required POSITIVE control was previously unobtainable without
> spending a real round, so this filter could only ever be half-tested. It is free now, in both
> directions, permanently. Registered as **FIX-061**.
>
> ### §6's controls, all satisfied — plus the arm §6 did not ask for
>
> | control | result |
> |---|---|
> | POSITIVE: config-only submission ACCEPTED without `FORCE=1` | ✅ exit 0 |
> | POSITIVE: `_HOLD.sql` submission accepted (see correction) | ✅ exit 0 |
> | POSITIVE: `platform/*.go` still accepted (regression) | ✅ exit 0 |
> | NEGATIVE 1: `.md` under `docs024_key_docs_latest/` still REFUSED | ✅ exit 2 |
> | NEGATIVE 2: `_ROLLBACK` / `_VERIFY` / `_SUPERSEDED` sidecars refused | ✅ exit 2 |
> | NEGATIVE: `README.md` inside the migrations dir refused | ✅ exit 2 |
> | **DISCONFIRMING** (not in §6): the same migration against the PRE-FIX script | ✅ exit 2 |
>
> That last row is the one that matters and §6 did not ask for it: the negatives prove the check
> was not deleted, but **only the pre-fix run proves the change did anything.**
>
> **One stated deviation from §6:** it asks that negatives refuse "with the unchanged message".
> The message *is* changed, deliberately — one that still said "docs and site content" after the
> widening would mislead the next author. The controls assert refusal + exit 2 + the new message.
>
> ### The gate found a real defect in the change that widens the gate
>
> `editquality`, medium severity: **excluding `_HOLD.sql` was wrong.** A `_HOLD` is a migration
> held back from the runner **for ordering** and applied **by hand** — live config reaching
> production, merely not ledger-recorded. I had excluded it by reusing the runner's `SIDECAR_RE`,
> which answers *"will `--apply` run this?"*, as a proxy for *"is this the change?"* — **which is
> this bug's own defect, committed inside its own fix.** Verified first-hand before acting
> (157 `_ROLLBACK` / 12 `_HOLD` / 7 `_VERIFY` / 2 `_SUPERSEDED`; the `_HOLD` files carry real
> `UPDATE`/`jsonb_set` writes), corrected in `738fcbb96`, and recorded in `WRONG_CALLS.md`.
> **This is the strongest available argument for the change itself** — and it is §9's finding
> reproduced live: the config-shaped round is where this council earns its keep.
>
> A second defect surfaced while fixing the first: the new suffix census ended in a `grep -v`
> that exits 1 when the census is CLEAN, which under the callers' `set -euo pipefail` killed the
> admission gate outright, silently, **only in the healthy case**. Caught by tracing an
> unexplained `exit 1` — not by the control matrix, whose harness was simultaneously destroying
> `$?` with a `$(basename …)` in the same `printf`. Both in `WRONG_CALLS.md` with the mechanical
> checks.
>
> ### Answers to the open questions this file left
>
> - **§5's costing caveat, priced:** council runs **11–43/day** (median ~24–33, 9 days, measured
>   from `llm_call_log`). Realistic increase **+4–6/day (+15–25%)**, not the +11 worst case, since
>   one round covers a task spanning several commits. Prompt caching is live at a measured **74%**
>   saving, and relevance gating limits seats per round. **Tripwire:** if sustained volume exceeds
>   **150% of the 9-day median for a fortnight**, raise §5 candidate 2 (a cheaper config-specific
>   roster) — do **not** revert admission.
> - **Candidate 2 is not built**, on §9's own evidence: `FORCE=1` did not degrade the review, so
>   what needed changing was admission, not the roster. Kept as the costed fallback above.
> - **§4's "098 inherits the blind spot":** fixed. The pathspec now covers the migrations
>   directory, with the exact predicate applied as a per-commit post-filter (a git pathspec cannot
>   express the sidecar exclusion). Population over 14 days: **411 → 583**, UNREVIEWED 69 → 149.
>   Verified both ways: three `494_*_HOLD.sql`-only commits excluded before the correction,
>   migration commit `5315c8a19` included.
> - **No RFC** (owner ruling 2026-07-29 §1): admission policy changed; the gate's *guarantees* —
>   round semantics, verdict artefacts, the trailer/098 join keys — did not. The `architecture`
>   seat independently reached the same conclusion (`ARCHITECTURE_SIGNAL: point_fix`). §3's
>   "consumers must be told" is discharged by the CLAUDE.md, RUNBOOK and LANDMINES edits.
> - **§7's process note stands:** no `090` run, deliberately, for the same reasons this file gives.
>
> ### Residual, deliberately left — and the LANDMINES entry stays live for it
>
> **Tooling is still out of scope.** `scripts/`, and the 097/098 scripts themselves, are not
> admitted — the widening covers appliable migrations only. So a change to the gate still needs
> `FORCE=1`, **with the force explained in the first paragraph of the rationale**. This fix's own
> submission was in exactly that position, which is why the LANDMINES entry was *amended* rather
> than retired. Whether tooling should be admitted is a separate question nobody has asked yet.
>
> **A FOURTH copy of the migration vocabulary exists and is already drifted:**
> `scripts/pattern-check.py:384-385` (`MIGRATION_NAME = ^\d{3}_[a-z0-9_]+\.sql$`, lowercase-only).
> It is an idempotency lint, not a scope consumer, so it is out of this fix's remit — but the
> drift is now concrete rather than suspected: it cannot see
> `482_ROLLBACK_claim_timeout_exclusion.sql`, a live runner-appliable migration, nor any `_HOLD`
> file. **A candidate for its own small ticket.**

**Filed:** 2026-08-18 · **Branch:** `087_towards_multiple_domains` · ~~**Status:** OPEN, diagnosed
with evidence, not fixed~~ → **CLOSED 2026-08-20, see banner above**
**Severity:** medium — it does not break anything running; it means two thirds of the estate's
behaviour changes reach production with the review gate declining to look at them.
**Class:** instrument gap / mis-targeted filter.
**Filed by:** the `bugfix-277/083` lane, recording **owner decision 6 of 2026-08-18** (*"as
recommended — raise as its own item"*), after both of that lane's config rounds needed `FORCE=1`.

---

## 1. What the gate does, in plain terms

`097_TRIGGER_council_review_v1.sh` is how any thread puts a change through the reviewer council
before committing it. Before it spends any credits it asks one question: *does this submission
touch code we review?* The answer is a regex over the **file paths** in the plan
(`097_TRIGGER_council_review_v1.sh:87`, read 2026-08-18):

```sh
SCOPE_RE='^(platform|internal|pkg)/'
```

and the check passes if **at least one** edit matches (`:146`). If none does, it refuses:

> `REFUSED: no edit touches the review scope (platform/, internal/, pkg/ — owner ruling 2026-07-17).`
> `Docs and site content do not spend council credits. FORCE=1 to override.`

## 2. The rule that is actually being applied, and why it misfires here

The rule the owner set in 2026-07-17 is about **subject matter**: prose does not spend council
credits. That is right, and the architecture-seat decisions doc argues it well —
`DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` §8d cites this exact line and concludes
the refusal is *correct*, because 72 DESIGN/PLAN/SPEC docs were created in one month and reviewing
them would cost real money.

**But the check implements that rule as a PATH TEST, and on this estate a large part of the running
system is config that lives under a `docs/` path.** Every DB migration — the file that changes what
a live agent does — is `docs/agent_docs/sql_for_agents/NNN_name.sql`. So a change that rewrites a
`scheduled_tasks.pre_query`, re-points an `agent_definitions` workflow step, or adds a seat to the
council's own roster is classified by that regex as prose, and refused with a message about site
content.

**The gate is not declining to review config. It cannot see that it is config.**

## 3. How much of the estate this is — measured, and it could have come out small

[MEASURED 2026-08-18] every commit in the preceding 14 days that ships a numbered migration,
split by whether it also touches `platform/`, `internal/` or `pkg/` — i.e. whether it would have
passed the scope check:

| | commits |
|---|---|
| shipping a numbered migration | **227** |
| of those, **config-only** (no in-scope file — the gate REFUSES) | **152 (67%)** |
| mixed, would pass the gate on the strength of its Go half | 75 (33%) |

Two thirds. And the 33% pass for the wrong reason: the check is `length > 0`, so a submission whose
Go half is one line and whose config half rewrites a fleet-wide scheduled task is admitted on the
one line.

⚠ **What this number is NOT.** It is not 152 unreviewed *changes* — some are docs commits that
happen to carry a migration, some lanes submitted with `FORCE=1` anyway, and one change often spans
several commits. It bounds the exposure; it does not enumerate it. The honest claim is that the
config path is refused **by construction**, and that the population it applies to is large.

## 4. Why it matters more than "a script says no"

- **`FORCE=1` is available, and that is the problem, not the mitigation.** The escape hatch has no
  audit trail distinguishing "the author knew better" from "the author wanted past a refusal". A
  gate whose normal use requires an override stops being a gate and becomes a formality — and
  CLAUDE.md was recently strengthened on exactly the grounds that the gate should be a real norm.
- **The `098` coverage report inherits the blind spot.** It joins commits to verdicts; a
  config-only commit that could never have been submitted still reads as un-reviewed, which is
  indistinguishable from a thread that skipped the gate.
- **The riskiest changes on this estate are config.** A migration is live the moment it COMMITs —
  no build, no roll, no image tag. A Go change waits for a chassis roll and can be reverted by
  rolling back. The half with the shorter fuse is the half nobody reviews.
- **The council reviews its own roster through this gate.** `099_SYNC_gate_roster.py` and the
  `fix-proposer`/`council-gate` mirror migrations are config, so a change to who reviews cannot
  itself be reviewed without `FORCE=1`.

## 5. Fix candidates, ordered by what closes the door

1. **(Preferred) Widen the scope test to name config by what it IS, not where it lives.** Add
   `^docs/agent_docs/sql_for_agents/[0-9]+_.*\.sql$` to `SCOPE_RE`. It is one line, it admits
   exactly the numbered migrations, and it still refuses prose — including the `_HOLD`/`_ROLLBACK`
   sidecars if the pattern is written to exclude them. **Makes the bad state unrepresentable**: a
   config-only submission stops needing an override.
   ⚠ **Cost, stated:** it also makes ~152 commits-worth of change per fortnight newly *eligible*,
   which is a real credit increase. Relevance gating means only footprint-matching seats fire, but
   the two always-on seats (edit-quality, guardian) run on every submission. Someone should cost
   that before it goes in, and it may argue for candidate 2.
2. **A distinct, cheaper config review** — a small seat set for migrations (guardian + a
   config-specific reviewer that knows about `pre_query` drift, verify blocks that cannot fail, and
   the `agent_definitions` snapshot convention) rather than routing config through the code roster.
   More work; better matched to what actually goes wrong in migrations.
3. **Move migrations out of `docs/`** to a top-level `migrations/` and add that to `SCOPE_RE`.
   Cleanest conceptually — config is not documentation — but it touches
   `scripts/migration/run-migrations.sh`, every runbook, the `MIGRATIONS_DIR` default, and every
   session's muscle memory, on a tree many sessions share. High disruption for the same outcome as
   candidate 1.
4. **Do nothing; keep using `FORCE=1`.** Rejected as the primary answer, but named because it is
   the status quo and it is not absurd: the override works, and the seats do review the plan once
   it is admitted. What it gives up is the ability to tell a skipped review from an impossible one.

## 6. How to verify a fix

Positive **and** negative control, or "the gate accepts config now" is equally consistent with
having disabled the scope check:

- **Positive:** a submission whose every edit is a numbered migration is ACCEPTED without `FORCE=1`.
- **Negative:** a submission whose every edit is a `.md` under `docs024_key_docs_latest/` is still
  REFUSED with the unchanged message. Without this, candidate 1 is indistinguishable from deleting
  `SCOPE_RE`.
- **Second negative:** a submission of only `_ROLLBACK.sql` / `_HOLD.sql` sidecars is refused —
  those are not applied by the runner and are not the change.

## 7. On process, stated rather than omitted

Per the owner ruling of 2026-07-31: **first-hand verification substituted for a `090` run,
deliberately.** There is no structural claim here that needs a diagnosis loop — the mechanism is
four lines of shell read at source (`SCOPE_RE` at `:87`, the `jq` test at `:146`, the refusal at
`:148`), the population figure is one `git log` loop a reader can re-run, and the prior art
(§8d of the architecture-seat decisions doc) was read and is cited rather than paraphrased. Nothing
here is inferred.

## 8. Related

- **Owner decision 6, 2026-08-18** —
  `docs024_key_docs_latest/bugfix_277_required_fields_repair/HANDOFF_2026-08-18c_continue_here.md` §5.
- `architecture_review/DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` §8d — the same
  line, argued from the docs side, concluding the refusal is correct **for prose**. This bug does
  not contradict it; it says the implementation does not distinguish prose from config.
- `bugs_closed/096` — the last defect found in this gate's dispatch path (head-of-line blocking).
- CLAUDE.md, "Council review of platform changes" — the norm this gap makes unsatisfiable for
  config authors.
- The two rounds that motivated it: council corrs `05a3d1c8` and `8dc58e2a` (`bugs_open/083`), both
  config-shaped, both submitted with `FORCE=1`.

---

## 9. Corroborating evidence from the `bugs_open 294` lane (added 2026-08-19, contributed not competing)

Three council rounds from a different lane, on the same branch, the same week. **Two of the three
needed `FORCE=1`**, and they are useful to this file because of *what the council did with them*
once it was allowed to look.

| round | correlation | in scope? | verdict |
|---|---|---|---|
| `463` reaper `RUNNING` arm, r1 | `860d87d9` | **no** — only file was `docs/agent_docs/sql_for_agents/463_*.sql` | **REVISE**, gated by `debug_historian` at **HIGH** |
| `463` r2 (same trail) | `860d87d9` | **no** | **APPROVED**, 2 advisories |
| `464` reaper `INITIALIZED` arm | `e973d2aa` | **no** | **APPROVED**, 4 advisories |
| delete `WorkflowMonitor` (Go) | `25fa8173` | **yes** — `platform/`, `cmd/`, `test/` | **APPROVED**, zero objections |

**The point this file can use: the in-scope round was the one with nothing to say, and the
out-of-scope rounds were the ones that changed the artefact.** The Go deletion — which the gate
admitted without argument — drew **zero objections**. The config-only change the gate would have
**refused** drew a HIGH-severity gating objection that was correct and materially improved the
work: my verify block asserted the rewritten `pre_query` text with `LIKE` substring checks, which
prove a needle is present and prove *nothing* about whether the assembled SQL parses — and a
`pre_query` parses only when the reaper next ticks, so a typo would have committed happily and
taken out **all five arms** minutes later. The fix (an `EXECUTE`-in-a-discarded-sub-block parse
check, plus an `md5` concurrency gate) exists only because that round ran.

So the current filter is not merely mis-scoped in the abstract: **it declines to look at exactly
the class of change where this council has demonstrably found the most severe defect**, and admits
the class where it found none. That is the inverse of what a relevance filter should do, and it is
a measured instance rather than an argument.

Two smaller notes for §5's candidates:

- **`FORCE=1` did not degrade the review.** All 16-ish seats ran normally and the verdicts were
  substantive, so whatever the fix is, it only needs to change *admission*, not the review itself.
- **The reviewers themselves were unbothered by the path.** No seat in any of the three rounds
  objected to a migration living under `docs/`; `editquality` and `debug_historian` reviewed the
  SQL on its merits. The mismatch is entirely in the pre-filter, which supports candidates that
  widen the regex over candidates that relocate the migrations.

*Source: `bugs_closed/294` (the `RUNNING` gap) and its sibling `464`. Not proposing a fix here —
this file's §5 already owns that, and its owner decision is recorded at the top.*
