# PLAN — 426, migration ledger drift check on a schedule

**Status:** decided, implementing same session. A `fable`-model planning pass was
requested but the agent hit a session rate limit mid-run (resets 16:10 Europe/London)
before producing anything usable — see NOTES. Rather than block the fix on that
quota, this plan was made directly, grounded in first-hand research AND a live
mechanism test against the real cluster (not paper design) — see §3.

## 1. Decision: build candidate 1 only; candidate 3 is not a substitute; candidate 2 is deferred

**Candidate 1** (bug's §5.1 — drive the existing dry run on a schedule, make its
silence legible) is what gets built. Reasons, not just the bug's own framing:

- It is the only candidate that covers **both** paths that produced the 34 live
  instances measured this session (see NOTES) — the plain-file path (most of them)
  and the `_HOLD`-rename path candidate 3 targets. Candidate 3 alone would have
  caught roughly none of today's 34.
- "Do nothing" (candidate 4) is refuted by the same measurement: the population is
  not shrinking on its own, it grew from 7 to (at least) 34 in under two weeks.

**Candidate 2** (contract the probe's message vocabulary away from prose-matching
`/already/i`) is a REAL, independent defect — confirmed twice now, once in the bug
file's own `672` case and once by `dispatch_throughput`'s direct reply this session
(see NOTES) — but it is **deferred, not folded into this commit**. Reasons:

- It means editing `scripts/migration/run-migrations.sh` itself, a script every one
  of ~30 concurrent sessions relies on interactively, and whose `probe_file()`
  behaviour hundreds of already-written migration guards implicitly depend on
  (any guard already written to say "already applied" must keep matching). That is
  a wider, higher-stakes blast radius than a brand-new, additive CronJob, and
  deserves its own change and its own review, not a rider on this one.
- **It doesn't need to block this fix**, because the mitigation is available at the
  REPORT layer instead of the runner layer: the runner already buckets every
  pending file into one of five verdicts (`CLEAN` / `ALREADY` / `DUP` /
  `NOT_PROBED` / `INCONCLUSIVE`) via distinct, already-existing output-line
  prefixes. `dispatch_throughput`'s point — that a hand-applied-unrecorded file can
  surface as `INCONCLUSIVE` rather than `ALREADY` when its guard's wording doesn't
  happen to contain "already" — is answered by **not discarding INCONCLUSIVE as
  noise**. The new job's report treats `ALREADY` + `DUP` + `INCONCLUSIVE` together
  as one "NEEDS REVIEW" bucket (sub-labelled by verdict, so a human can tell a
  probable-already-applied file from a probable-genuine-drift file), and only
  `CLEAN` + `NOT_PROBED` as non-actionable. This captures the practical harm today,
  with zero changes to the shared script, and zero new risk to concurrent sessions.
- Record candidate 2 back into `bugs_open/426` itself as the explicit residual, so
  it isn't lost — not silently dropped, a stated deferral with a reason.

## 2. The one real design problem, and how it's solved

Every existing "quiet mechanism" precedent on this estate needed either pure DB
state (`single-owner-carriers-check`) or file metadata / prose (`bugs-open-
staleness-sweep`). This check needs the **real SQL text** of every pending
migration, because the "applied by hand, unrecorded" signal only exists by
*executing* each file inside a doomed, self-rolling-back transaction and reading
its own guard's `RAISE` — there's no metadata-only substitute.

**Decision: a shallow, partial, sparse `git clone`, not the GitHub Contents API,
and not a re-implementation of the runner's vocabulary.**

- `docs/agent_docs/sql_for_agents/` is **16MB, 1,151 files** — measured this
  session (`du -sh`). A `--filter=blob:none --sparse --depth 1` clone of just that
  directory plus `scripts/migration/` (32KB) is one command, fetches only the blobs
  actually checked out, and — **measured live against the real GitHub repo from
  inside a throwaway pod in `ai-persona-system` this session** — completed in
  **1.8 seconds**. This is nothing like the 262M-`.git` problem
  `bugs-open-staleness-sweep`'s own comment rules out; that concern is about
  cloning **history**, and `--depth 1` never fetches any.
- This gets **real, current file content** for the *entire* migrations directory in
  one shot, with **zero re-implementation** of which files are migrations, which
  are sidecars, or what the baseline number is — `run-migrations.sh` itself decides
  all of that, unmodified, exactly as CLAUDE.md's "reuse existing machinery" and the
  bug's own §4 require. The alternative (GitHub Contents API, one fetch per file)
  would need to *already know* which files are pending before fetching them — i.e.
  duplicate the vocabulary to decide what to fetch — or fetch all 1,151 files
  individually, which is strictly worse than one clone.
- The runner is then invoked **completely unmodified**: `PSQL_CMD` overridden to a
  direct `psql` command (see §3 — no `kubectl exec`, matching every existing
  CronJob precedent) and `MIGRATIONS_DIR` pointed at the local sparse checkout.

## 3. Proven live, not just designed — the mechanism test this session

Built two throwaway test pods in `ai-persona-system` (`postgres:16-alpine`, deleted
immediately after each test — no persistent footprint) to verify the actual
mechanism against the real cluster and the real GitHub repo, rather than trust the
design on paper. Both later deleted.

**Test 1 — clone + direct DB connection:**
```
apk add --no-cache bash git
AUTH_B64=$(printf 'x-access-token:%s' "$GITHUB_READ_TOKEN" | base64 | tr -d '\n')
git -c http.extraHeader="Authorization: Basic ${AUTH_B64}" \
  clone --filter=blob:none --sparse --depth 1 --branch "$MIGRATION_CHECK_REF" \
  "https://github.com/${REPO_OWNER}/${REPO_NAME}.git" /tmp/repo
cd /tmp/repo && git sparse-checkout set docs/agent_docs/sql_for_agents scripts/migration
```
Result: clone in 1.82s, 1,141 real files present (not stubs — confirmed by `head -1`
on a known migration, matched the real file), `git --version` 2.54.0 (has
sparse-checkout support). `PGPASSWORD=... psql -h postgres-clients ...` returned `1`
— direct DB connectivity confirmed, no `kubectl exec`, no RBAC needed (matches
`bugs-open-staleness-sweep`'s comment about `ai-persona-app` having no pods/exec
RBAC in this namespace).

**Header form matters — `Authorization: Basic base64(x-access-token:$TOKEN)` via
`http.extraHeader`, NOT a URL-embedded token, NOT `Authorization: bearer $TOKEN`.**
Tried the bearer form first; it failed (`fatal: could not read Username for
'https://github.com'` — git didn't recognise it as auth and tried to prompt
interactively, which fails non-interactively). The Basic-auth form worked first
try. **This is a real, first-hand-verified gotcha, not a style choice — get the
header form wrong and the clone fails, non-obviously, with an error that looks like
a missing credential rather than a malformed one.**

**Test 2 — the unmodified runner, wired through this mechanism:**
Ran `bash /tmp/repo/scripts/migration/run-migrations.sh --no-probe` (fast path,
skips the actual per-file probing but exercises the exact same PENDING computation,
idempotency lint, and PSQL_CMD wiring the full run uses) with `PSQL_CMD="psql -h
postgres-clients -p 5432 -U clients_user -d clients_db"` and `MIGRATIONS_DIR=/tmp/
repo/docs/agent_docs/sql_for_agents`. It ran to completion, exit 0, produced the
same `Pending (N):` / idempotency-warning output shape as an interactive session's
own run.

**Found and fixed a real environment defect in the process, not a hypothetical
one:** `postgres:16-alpine`'s BusyBox `grep` does not support `-P` (PCRE), and
`run-migrations.sh`'s `lint_idempotency()` uses `grep -oiP`. Without GNU grep, every
pending file's lint call fails and dumps BusyBox's `Usage: grep [...]` help text
into the output — once per pending file, so potentially 100+ times, drowning the
actual report. **Fix: `apk add --no-cache bash git grep`** (Alpine's `grep` package
*is* GNU grep, unlike the `grep` applet BusyBox otherwise provides) — confirmed
clean after adding it: real idempotency warnings, no BusyBox usage-text spam.
This is exactly the kind of thing that would have shipped broken and silently ugly
if this plan had gone from design straight to a cronjob.yaml without running it for
real first.

## 3b. Deployed and run against production — a second real defect found and fixed

After building the full `check.sh` + `cronjob.yaml` (§4) and reviewing the script by
re-reading it looking specifically for "claims of success that aren't checked"
(this estate's own recurring lesson), two defects were fixed BEFORE deploying:
the doc_notes `psql -f` call's exit status was never checked (a failed write would
still print "row written"), and a `run-migrations.sh` run that ended abnormally
(DB unreachable, hard error) rather than with its normal `Pending (N):` / `Up to
date` line would have produced a false "all clean" report — exactly the "broken
check reads identical to a healthy one" failure this whole job exists to prevent.
Both now refuse loudly (exit 2) instead.

Then deployed for real (`kubectl apply -k .../overlays/production/uk_001`,
matching the precedent that every sibling check here has no `make deploy-*`
target and was applied directly) and triggered a manual run
(`kubectl create job --from=cronjob/migration-ledger-drift-check ...`) to prove
the whole thing end to end against production — the one thing the throwaway test
pods deliberately hadn't done, because it means a real doc_notes write.

**Result: the mechanism worked completely correctly** — confirmed by reading the
doc_notes row directly, not by trusting the Job's status (this estate's own
"trust the rendered artefact, not the status" lesson). Real pending count 158,
81 clean, 5 not-probed, 72 in NEEDS REVIEW (34 ALREADY + 38 INCONCLUSIVE, 0 DUP —
matches the shape of this session's earlier manual measurement almost exactly,
with the small drift expected from other sessions recording files in the
meantime). Full rendered report is real, readable, and actionable — filenames plus
the runner's own verbatim guard detail for every entry.

**But the Job itself showed `Failed`, and that surfaced a real design defect,
not a fluke.** `check.sh`'s first version exited 1 whenever NEEDS-REVIEW was
non-empty, copying `single-owner-carriers-check`'s "exit non-zero on findings =
second signal via Job status" convention. That convention assumes findings are
RARE. Here they are the daily norm (72 of 158 on the very first run), so exiting 1
made `backoffLimit`'s retry re-run the ENTIRE clone-and-probe a second time
(confirmed: two pods, two full runs, two doc_notes rows 44 seconds apart), and
still ended in `BackoffLimitExceeded` — a Job that will read as broken, forever,
on every single ordinary day. **Fixed in §1/§4's exit-code decision, corrected
after this test, not assumed from the start:** exit 0 whenever the report was
generated and written, regardless of findings; only genuine operational failure
(clone/env/DB/write) exits non-zero. The doc_notes row, not the Job's exit code,
is this check's actual signal — consistent with every other quiet-check on this
estate being read by querying `doc_notes`, not by watching Job history.

Re-deployed with the corrected `check.sh` (via the ConfigMap regenerating a new
hash suffix and the CronJob picking it up — `kubectl apply -k` again) before
declaring this done. The two now-stale doc_notes rows from the pre-fix test runs
were left as-is (append-only log, real events, not misleading — both correctly
reported the true state of the ledger, they just came from a run that then
mis-reported ITS OWN completion status upward).

## 4. Concrete build

New service: `deployments/kustomize/services/migration-ledger-drift-check/`

```
base/
  cronjob.yaml
  check.sh              # the wrapper: clone, invoke run-migrations.sh, parse, report
  kustomization.yaml     # configMapGenerator for check.sh
overlays/production/uk_001/
  kustomization.yaml     # namespace: ai-persona-system, labels — copy single-owner-carriers-check's overlay verbatim
```

**Why a shell script, not Python** (unlike every sibling precedent): the whole
point is to invoke `run-migrations.sh` (bash) unmodified. A Python wrapper would
just be `subprocess.run(["bash", ...])` with extra ceremony around it for no
benefit — the report-parsing (grep on `!!`/`??`/`--`/`ok`/`Pending (` prefixes) is
no harder in POSIX shell than in Python, and it keeps one language in the container
instead of two.

**Image:** `postgres:16-alpine` (psql pre-installed, matches every sibling).
`apk add --no-cache bash git grep` at container start — no custom Docker build, no
`IMAGE_TAG` bump, ever, same as every sibling check.

**Env:**
- `MIGRATION_CHECK_REF` — **no default**, same discipline as `bugs-open-staleness-
  sweep`'s `SWEEP_REF`: a human bumps it by hand when the platform's live working
  branch changes. Set to `"087_towards_multiple_domains"` today (the current
  branch, confirmed via `git branch --show-current` this session).
- `REPO_OWNER=gqls`, `REPO_NAME=agentchassis` — copied verbatim from
  `bugs-open-staleness-sweep`'s defaults (same repo, same convention).
- `PG_CLIENTS_HOST=postgres-clients`, `CLIENTS_DB_PASSWORD` (secret) — copied from
  `single-owner-carriers-check`.
- `GITHUB_READ_TOKEN` (secret) — copied from `bugs-open-staleness-sweep`. Both
  secrets already live in `personae-platform-secrets` (confirmed this session —
  no new secret provisioning needed).

**Schedule:** `45 7 * * *` UTC. Surveyed every live `cronjob.yaml`'s `schedule:`
this session — the 06:00–08:00 UTC band is where every daily check clusters, and
`07:45`/`07:50` are the only two free minutes in it. Clear of `database-backup`
(02:00) and the weekly `bugs-open-staleness-sweep` (Sunday 06:00).

**`activeDeadlineSeconds`: 1800** (30 min) — generous relative to siblings
(300–900s), because this job's workload (probing every pending file) has no
comparable precedent and only grows as more sessions add migrations daily. Direct
`psql` (no `kubectl exec` per call) should be materially faster than an interactive
session's dry run, but there is no live measurement of the FULL probe path's
duration under this mechanism yet (see §6 residual). If the job is regularly
running close to the deadline, that is itself a signal worth alarming on, not
silently extending — revisit the schedule/budget rather than just raising the
number.

**doc_notes:** `subject_type='pipeline'`, `subject_key='migration-ledger-drift'`
(confirmed free this session against live `doc_notes` — no existing key resembling
"migration"). One row per run, **including clean runs**, body = the buckets +
counts + (for NEEDS-REVIEW items) the filename and the verbatim runner detail line,
so a human can go straight to `--record-only` without re-deriving anything.

**Exit code — CORRECTED after the live test (§3b), not left at the first design.**
Originally: exit non-zero when NEEDS-REVIEW is non-empty, matching
`single-owner-carriers-check`'s "second signal via Job status" convention. The
FIRST real run against production (§3b) showed why that convention doesn't
transfer here: that sibling's findings are meant to be RARE (a bug), so failing
loudly and eating a `backoffLimit` retry is cheap. This check's findings are the
ORDINARY DAILY STATE (measured live: 72 of 158 pending on the very first run) —
exiting 1 on them made `concurrencyPolicy: Forbid` + the retry re-run the entire
clone-and-probe a SECOND time every single day, wrote a second near-duplicate
doc_notes row 44 seconds later, and left the Job permanently red with nothing a
human would ever clear. Fixed: exit 0 whenever the report was generated and
written, regardless of findings — the doc_notes row is the durable signal (RFC_022's
own convention: missing row = didn't run, present row = ran, read its body).
Genuine operational failure (clone/env/DB/write) still exits 2, well before
this point, and is what the retry budget is actually for.

## 5. Council scope — confirmed out, matching precedent exactly

Read `scripts/council-scope.sh` in full this session.
`COUNCIL_SCOPE_CODE_RE='^(platform|internal|pkg)/|^cmd/config-key-audit/|^scripts/pattern-check\.py$'`.
Every file this plan adds sits under `deployments/kustomize/services/migration-
ledger-drift-check/` — outside that regex — and nothing here edits
`scripts/migration/run-migrations.sh` (candidate 2 is deferred, see §1). This is
the **identical** scope situation `single-owner-carriers-check` was in; its own
commit message states plainly "Out of council scope by the gate's own filter... so
not submitted." Same treatment here — no council submission for this commit.

## 6. Risk to concurrent sessions, and what's still owed

- The probe **executes** each pending file's SQL before rolling back — brief row
  locks, possible sequence advances (documented in `run-migrations.sh`'s own header
  comment). This is not a NEW class of risk: CLAUDE.md already mandates every human
  session run this exact dry run "per session and after every roll", so ~30
  concurrent sessions are already expected to trigger it routinely. One more
  scheduled run daily does not change the risk profile, just the cadence and who's
  watching.
- **Residual, explicitly not resolved by this plan:** the full probe path's real
  duration under direct-psql (not `kubectl exec`) was not measured this session —
  only `--no-probe` was tested live (§3), to avoid burning another ~10+ minutes of
  cluster/API time re-running the exact probe already run once manually this
  session (see NOTES's 161-file, ~34-ALREADY measurement). First real scheduled run
  (or a manual `kubectl create job --from=cronjob/migration-ledger-drift-check
  ...`, the same pattern every sibling's `-now` Makefile target uses) IS that
  measurement, and `activeDeadlineSeconds` should be revisited from its actual
  wall-clock time, not assumed.
- The pending population will keep growing as long as sessions keep adding
  migrations faster than they're applied — this job doesn't fix that, it only makes
  the "applied but unrecorded" subset of it visible. That's the whole and correct
  scope of candidate 1; a growing backlog of genuinely-not-yet-applied files is a
  separate, expected, healthy state this job must NOT flag as a problem (hence
  bucketing `CLEAN` separately and never treating raw `Pending (N)` as the finding).

## 7. Verification (bug's own §6)

- **Positive control: already satisfied, no synthetic plant needed.** Today's real
  34 `ALREADY`-verdict files (list in NOTES) are a live positive control. First real
  run of the new job should name at least that many (allowing for drift between
  measurement time and first run — some may get `--record-only`'d in the interim,
  which is the correct outcome, not a discrepancy to chase).
- **Negative control:** confirm a day with zero NEEDS-REVIEW findings still writes a
  `doc_notes` row (the CLEAN-bucket report, not silence). Check by direct query
  after the first scheduled run, same as `single-owner-carriers-check`'s own stated
  convention.
- **Candidate 2's verification (deferred, not owed by this plan):** left for
  whoever picks up the message-vocabulary fix — bug file's own §6 already states it
  (induce `672`'s drift arm, confirm reclassification).
