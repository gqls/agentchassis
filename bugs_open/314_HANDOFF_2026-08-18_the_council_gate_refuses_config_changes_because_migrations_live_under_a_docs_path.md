# 314 — the council gate cannot review a config-only change, because its scope check is a PATH regex and this estate's config ships as SQL files under `docs/`

**Filed:** 2026-08-18 · **Branch:** `087_towards_multiple_domains` · **Status:** OPEN, diagnosed
with evidence, not fixed
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
