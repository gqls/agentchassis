# RFC 024 — nine CronJob meta-checks, no shared harness: the duplication three seats have now flagged twice (and everyone, including me, undercounted)

**Status: OPEN — raised 2026-08-11 by the `bugfix_213` lane (D3). Nothing built, nothing
changed. This is a routing document: three council seats asked for it, and "recorded not
actioned" was already my own answer once.**

## 1. Why this exists rather than another line in a risks block

The council approved `verifier-remit-check` (corr `fc082c4a-4b00-4835-8ffe-11a55e53f47a`,
round 2) with three seats independently saying the same thing:

> `reuse_agent [medium]`: "This is the FIFTH standalone Go CronJob meta-check binary …
> each reimplementing its own doc_notes-on-every-run, dedup and retraction plumbing. The
> plan's own risks block names this exact concern (architecture seat, round 1, 'recorded
> not actioned') but ships the fifth instance anyway."

> `architecture [medium]`: "Fifth near-duplicate standalone CronJob meta-check service
> (own dockerfile, overlay, doc_notes convention, dedup) with no shared harness — the
> author's own risk list flags this and does not act on it. Not blocking this fix, but the
> pattern should get a consolidation pass before a sixth instance ships."

> `tooling_provenance [low]`: "… it is the fourth instance of the same reinvention and
> should not be waved through indefinitely."

**The same signal fired in round 1 and I answered it by writing it down.** Writing it down
again would be the same non-answer, so it is filed here instead — where a decision can be
taken, or refused on the record.

## 2. The population, named [MEASURED 2026-08-11]

Scheduled meta-checks with their own image or ConfigMap script, their own CronJob, their
own overlay, and their own copy of the reporting convention:

| job | language | schedule (UTC) | what it censuses |
|---|---|---|---|
| `bugs-open-staleness-sweep` | Python (ConfigMap) | Sun 06:00 | repo prose vs live state |
| `single-owner-carriers-check` | Python (ConfigMap) | 06:20 | `agent_definitions` — RFC_006 carriers |
| `removed-config-keys-check` | **Go image** | 06:25 | `agent_definitions` (RFC_012) |
| `site-discovery-staleness-check` | Python (ConfigMap) | 06:35 | discovery rotation freshness |
| `component-fallback-check` | Python (ConfigMap) | 06:40 | component templates (CGV-029) |
| `concept-register-drift-check` | Python (ConfigMap) | 06:50 | register vs code |
| `component-render-check` | **Go image** | 06:55 | rendered output (CGV-030) |
| `shared-output-fields-check` | **Go image** | 07:10 | `agent_definitions` (RFC_012 d) |
| `verifier-remit-check` | **Go image** | 07:25 | `site_work_items` producer shapes (WII-015) |

That is **nine** jobs, four of them Go images (from three `cmd/` packages), not five.
Every one of the nine writes a doc_notes row per run — verified, not assumed:
`grep -rln "INSERT INTO doc_notes" cmd/ deployments/kustomize/services/*/base/*.py`
returns 3 Go files and 5 Python ones, covering all nine jobs (two share the
`config-key-audit` binary).

**The seats undercounted, and so did I** — my own submission said "the fifth". That
nobody involved could name the set from memory is itself the argument: this is a
population being maintained by copy, with no list.

## 3. What is actually duplicated

Each Go job independently implements:

1. **`dbConn()`** — the identical `PG_CLIENTS_HOST` / `CLIENTS_DB_PASSWORD` /
   `sslmode=disable` block, plus the identical `kubectl exec` fallback for a session at a
   terminal. [MEASURED] `grep -l PG_CLIENTS_HOST cmd/*/*.go` returns **four** files across
   three packages — `config-key-audit` carries it twice (`fleetdb.go`, `sharedoutputs.go`).
   The comment explaining *why* — ai-persona-app has no `pods/exec` RBAC — is copied too.
2. **doc_notes-on-every-run** — one row per run, clean or not, so a missing row means the
   job did not run. Same insert, different `subject_key`/`categories`/`source`.
3. **Exit-code discipline** — 0 ran / 1 findings / 2 could not run. Stated in each header.
4. **The CronJob YAML** — `imagePullSecrets: docker-hub-creds` (whose omission produces a
   Job that reports *Running* for ever), `concurrencyPolicy: Forbid`, `backoffLimit: 1`,
   an `activeDeadlineSeconds` justified in a comment, and a schedule chosen by reading the
   other seven jobs' schedules out of their YAML.
5. **The makefile quartet** — `build-/push-/deploy-/<name>-now`, plus an overlay whose
   `newTag` must be moved in lockstep with the build (the trap `debug_historian` gated my
   round 1 on).

Only (1)–(3) are library-shaped. (4) and (5) are a **template** problem, not a library one.

## 4. What it costs today, stated honestly

- **Not correctness.** Each job works, and the duplication is copy-not-fork: the copies
  have not diverged in behaviour, only in the comments around them.
- **It costs the SCHEDULE.** Every new job picks its slot by reading the others' YAML. Two
  jobs already chose 07:10 independently (`shared-output-fields-check` has it; this lane's
  first draft claimed it too, and moved to 07:25 only because a human read the file).
- **It costs the CONVENTION.** doc_notes-on-every-run exists because a silent check is
  indistinguishable from a stopped one — but nothing enforces it. A ninth job that forgets
  is invisible, and the check that would notice does not exist.
- **It costs the REVIEW.** Three seats now spend part of every such round re-litigating
  the shape instead of the change.

## 5. The options, costed

1. **Do nothing; keep copying.** Cheapest per job; the schedule collision and the
   unenforced convention keep growing. This is the status quo and it is what the seats
   object to.
2. **A tiny shared package** — `internal/cronchecks` with `DB()`, `Report(subjectKey,
   categories, source, body)` and the exit-code constants. ~80 lines, no behaviour change,
   removes (1)–(3). Does nothing for the YAML or the makefile.
3. **(2) plus a manifest generator** — one YAML template rendered per job from a small
   spec (name, schedule, deadline), so `imagePullSecrets` and `concurrencyPolicy` cannot be
   forgotten and the schedule table lives in one file that can be checked for collisions.
4. **A single meta-check binary with subcommands**, the `config-key-audit` shape — one
   image, one CronJob per subcommand. Highest consolidation; highest blast radius, because
   one bad build then stops nine checks instead of one, and their failure modes stop being
   independent (which is the reason `component-render-check` was deliberately split from
   `component-fallback-check` — see that job's YAML header).

**This lane's recommendation is (2) then (3), and explicitly NOT (4)**: the independence of
these jobs' failure modes is load-bearing and was itself an earlier design decision. But it
is not this lane's call, and the point of filing is that the next author should not have to
re-derive the argument from three council objections again.

## 6. Cheapest disconfirming check for whoever picks this up

If the duplication is really only cosmetic, this is one grep:

```bash
grep -l "PG_CLIENTS_HOST" cmd/*/*.go                      # the copies of dbConn
grep -rn "INSERT INTO doc_notes" cmd/ deployments/kustomize/services/*/base/*.py
grep -rn "schedule:" deployments/kustomize/services/*/base/cronjob.yaml   # the collision surface
```

A ninth job that omits the doc_notes row, or lands on an occupied minute, is the finding
that would make (3) urgent rather than tidy.
