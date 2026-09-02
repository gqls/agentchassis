# NOTES — bugfix 426, migration ledger drift check (append-only, newest at bottom)

## 2026-09-02, session start

Resuming `bugs_open/426`. `scripts/who-owns.py 426` shows no session actively fixing
it — it was filed today by the `bugfix_314_council_scope` lane (at the request of,
and crediting, the `dispatch_throughput` lane) and explicitly punted as "a different
kind of problem, deserves its own item; not mine to fix". Their README says so
directly. So this is a genuinely open, unclaimed bug, not a competing effort.

**Bug in one line:** `scripts/migration/run-migrations.sh`'s default no-arg dry run
already computes exactly the set this estate cares about — appliable migrations with
no `schema_migrations` row — and probes each one to say whether it looks like it was
already applied by hand. It is free, read-only-ish (executes+rolls back), and
CLAUDE.md already mandates running it "per session and after every roll". Nobody
drives it on a schedule, so the signal only surfaces when a human happens to run it
interactively. Seven such instances sat unnoticed for a month before the
`dispatch_throughput` lane found them by chance.

## Validity check — re-run the dry run myself, live, 2026-09-02

Per the bug's own §8 instruction ("do this before designing anything"), ran:
`./scripts/migration/run-migrations.sh` (default, no args) in the background — it
takes a long time (each pending file is probed in a doomed transaction via
`kubectl exec`, so ~161 pending files × a few seconds each).

**Pending: 161** files (baseline 124), as of this run.

Confirmed the bug is still live and the population is still growing, not just the
original seven. Spotted in the probe output as it ran (not exhaustive — this is
mid-run sampling, full list to follow):
`234_narrow_aao_concurrency_ban_to_self_claims.sql`,
`235_aao_specs_are_the_source_that_reinjects_the_old_figure.sql`,
`434_enable_finance_directory_and_structural_checks_b3f.sql`,
`441_planner_directory_rule_enumerate_listing_names.sql`,
`491_robot_hands_ifr_facts.sql`, `492_robot_hands_demand_feature_page.sql`,
`493_robot_hands_insights_index.sql`, `494_dartsonline_pdc_calendar_facts.sql`,
`498_robot_hands_compressed_air_feature.sql` — all verdict `ALREADY` (guard raised a
message the probe's `/already/i` match caught). These are new sightings, not the
original seven (which the dispatch_throughput lane already `--record-only`'d).

**[MEASURED 2026-09-02, full run complete, log kept at
`/home/ant/.claude-scratch/.../scratchpad/migration_dry_run.log` this session only —
not durable, so the tallies below are the record]**: of 161 pending files —
**84 CLEAN** (genuinely not yet applied, waiting their turn — the healthy majority),
**34 ALREADY** (probe's guard raised a message matching `/already/i` — the exact
"applied by hand, unrecorded" signal this bug is about), **0 DUP**, **5 NOT_PROBED**
(contain their own ROLLBACK/ABORT etc., refused fail-closed by the probe's own
design), **38 INCONCLUSIVE** (probe hit a different `P0001` — some of these read
like drift/precondition failures rather than "already applied", e.g. `497`/`498`
report the live `pre_query` column not matching what the migration expects at all,
which is a DIFFERENT hazard, not this bug's).

**This is far bigger than the seven the bug file documents.** Those seven were one
lane auditing itself after a heads-up; this is the fleet-wide population on one
ordinary day, and it is **34, not 7** — about 21% of everything currently pending.
None of the seven already-fixed filenames (`582`/`583`/`584`/`637`/`671`/`672`/`673`)
appear in this ALREADY list, confirming their `--record-only` fix is holding (they
correctly dropped out of "pending" once recorded).

This measurement is itself the strongest argument for candidate 1 (drive the dry run
on a schedule): the check ran once, by hand, in one session, and immediately
surfaced 34 live instances of exactly the dangerous state the bug describes, with a
mechanism that already exists and cost nothing to build.

## Contribution from `dispatch_throughput`, in response to the heads-up above

Messaged them the 34-vs-7 measurement. Their reply (verbatim point, credited): **34 is
a floor, not the population.** The probe's `ALREADY` verdict only fires when the
guard's `RAISE` text happens to match `/already/i` — an undocumented vocabulary
contract, which is exactly the bug's own §3 second finding (the `672` case: a
truthful "drifted, investigate" message that actually means "applied by hand,
fine", because the wording doesn't contain "already"). So a hand-applied-unrecorded
file whose guard fires a DIFFERENTLY-worded arm shows up in my run as one of the
**38 INCONCLUSIVE**, not as one of the 34 ALREADY — and INCONCLUSIVE is exactly the
bucket I was about to treat as noise / lower-priority in the report design.

**This changes the report's design, not just its footnote.** The new CronJob's
report should carry (at least) three buckets, not two: ALREADY (vocabulary
matched, high confidence), PROBE-ERROR / P0001-other (guard fired but worded
differently — triage as POSSIBLY applied-out-of-band, not as "broken SQL" or
"safe to ignore"), and genuinely-CLEAN/pending. Collapsing bucket two into "broken
SQL" or silently omitting it from the report would hide exactly the class this job
exists to surface — the same failure shape the bug file already named once.

Also confirmed from them: the runner's own stated contract is that recording stays
a HUMAN act — the cron job must only REPORT, never `--record-only` on anyone's
behalf. Matches my own reading of `run-migrations.sh`'s `--record-only` block
(requires `--note` describing what a human verified) and the bug's own framing
throughout. Not up for reconsideration in the plan.

## Fable planning pass — hit a rate limit, did not complete

Dispatched a `Plan`-type agent with `model: fable` per the task's own request,
with a long, fully-grounded prompt (all research above, plus explicit
instructions to verify claims rather than trust them). It got as far as starting
to read the two precedent CronJobs, then failed:
`API error: You've hit your session limit · resets 4:10pm (Europe/London)`.

Decision: did not wait for the quota reset (unknown how far off that was, and the
bug is real and growing — see the 34-instance measurement above). Wrote the plan
directly instead, building on all the research already done this session, and —
unlike a paper plan — actually PROVED the mechanism against the real cluster before
committing to it (see PLAN §3: two throwaway test pods, deleted after use, that
verified the sparse-clone-plus-direct-psql mechanism end to end, including finding
and fixing a real BusyBox-grep-vs-GNU-grep incompatibility that would otherwise
have shipped as silently ugly output). The full design reasoning lives in
`PLAN_2026-09-02_migration_ledger_drift.md`, not repeated here.

## Build + live test, 2026-09-02

Built `deployments/kustomize/services/migration-ledger-drift-check/` (cronjob.yaml,
check.sh, kustomization.yaml, overlay). `kubectl kustomize` renders cleanly;
`bash -n check.sh` syntax-checks; the parsing/bucketing logic was tested locally
against the REAL captured report from this session's manual dry run (see above) and
reproduced the exact same tallies (161 pending, 84 clean, 34 ALREADY, 0 DUP, 5
not-probed, 38 inconclusive — sums to 161).

Found two real bugs in my own first draft while reviewing before shipping, both
fixed before deploying (not found by testing — found by re-reading my own script
looking for exactly this class of mistake, per this estate's own "a claim about
behaviour is not the behaviour" lesson):
1. The doc_notes INSERT's exit status was never checked — a failed write would
   still print "doc_notes row written". Fixed: now refuses and exits 2 if the
   `psql -f` call fails.
2. If the underlying `run-migrations.sh` run ends abnormally (DB unreachable, a
   hard error) rather than with its own normal "Pending (N):" / "Up to date —"
   ending, the report would have contained neither an ALREADY nor a Pending line,
   so every count would default to zero and the job would have written a false
   "up to date, all clean" row — exactly the "broken check reads identical to a
   clean one" failure this whole job exists to prevent. Fixed: now refuses to
   write anything and exits 2 if the report doesn't end the way a real run always
   does.

**Deployed live** (`kubectl apply -k .../overlays/production/uk_001`) — matches
the precedent of every sibling check here having no `make deploy-*` target and
being applied directly. CronJob confirmed present, schedule `45 7 * * *` UTC, not
yet due to fire on its own. Triggered a manual run immediately
(`kubectl create job --from=cronjob/migration-ledger-drift-check
migration-ledger-drift-check-manual-test-1`) to prove the whole mechanism
end-to-end against production, including the real doc_notes write — the one thing
the throwaway test pods deliberately didn't do.

**Result: the job FAILED** (`BackoffLimitExceeded`, 2 pods). Pods were already
garbage-collected by the fleet's own `agent-job-cleanup` CronJob (`*/10 * * * *`)
by the time I looked, so no pod logs — but querying `doc_notes` directly showed
**two real, well-formed reports had been written**, 44 seconds apart
(13:30:14 and 13:30:58 UTC): pending 158, clean 81, not-probed 5, NEEDS REVIEW 72
(34 ALREADY + 38 INCONCLUSIVE + 0 DUP). The mechanism worked completely — clone,
probe, parse, write — TWICE. The "failure" was `check.sh` exiting 1 because
findings existed, which `single-owner-carriers-check`'s convention treats as a
legitimate "fail loudly" signal (that check's findings are meant to be rare). Mine
are the daily norm (72/158 on the very first run), so exiting 1 triggered
`backoffLimit`'s retry — a full second clone+probe, wastefully — and then still
ended in `Failed` because the SECOND run also (correctly) found the same
findings. **Fixed:** `check.sh` now exits 0 whenever the report was generated and
written, regardless of findings — only genuine operational failure (clone/env/DB/
write) exits non-zero. Full reasoning: PLAN §3b (added after this).

Re-applied (`kubectl apply -k`, new ConfigMap hash, old one deleted), deleted the
failed test job, triggered a fresh manual test
(`migration-ledger-drift-check-manual-test-2`). **Result: `Complete`, single pod,
51 seconds wall-clock (13:44:05→13:44:56 UTC), exit 0.** Read the pod's logs
before cleanup this time (confirmed `NEEDS_REVIEW_N=72`, same shape as before) and
confirmed the doc_notes row via direct query — 3 total rows now under
`subject_key='migration-ledger-drift'` (the two from the failed-exit-code test,
left in place as real, accurate, append-only history, plus this clean one).

**This is now genuinely fixed and live**, not just committed-and-hoping: the
CronJob exists in the cluster, is scheduled (`45 7 * * *` UTC, not yet due to fire
on its own as of this writing), and has been proven twice by manual trigger against
real production data with a real doc_notes write each time. `bugs_open/426` updated
in place (§10) rather than moved to `bugs_closed/`, because candidate 2 (the
probe's message-vocabulary contract) remains a real, deliberately-deferred residual
— see PLAN §1 for why it wasn't bundled into this change.

## Precedent research — what already exists to build on

Three live CronJobs are the direct precedent for "drive a check on a schedule and
make silence legible" (candidate 1 in the bug):

- `deployments/kustomize/services/single-owner-carriers-check/` (RFC 006) — the
  closest shape: `postgres:16-alpine` base image, `apk add python3` at container
  start (no custom image, no IMAGE_TAG bump), connects to Postgres **directly**
  (`psql -h postgres-clients`, `CLIENTS_DB_PASSWORD` secret) rather than via
  `kubectl exec` — its own comment says why: **`ai-persona-app` has no pods/exec
  RBAC on `ai-persona-system`**, so a CronJob literally cannot do what an
  interactive session's `kubectl exec`-based `PSQL_CMD` does. Writes ONE
  `doc_notes` row per run (subject_type `pipeline`), including on a clean result,
  so a missing row means "job didn't run", never conflated with "all clear".
  Exits non-zero on findings (Job shows failed — second, independent signal).
  **Its own commit message states it is OUT OF COUNCIL SCOPE** by
  `scripts/council-scope.sh`'s filter (`deployments/` + a `cmd/` test aren't in
  `^(platform|internal|pkg)/|^cmd/config-key-audit/|^scripts/pattern-check\.py$`)
  and was **not submitted** — direct precedent for this bug's fix too.
  **Has no `make deploy-*` target** (checked: `grep single-owner-carriers-check
  makefile` — nothing) yet **is live in the cluster** (`kubectl get cronjob` shows
  it running on schedule, age 30d) — so it was deployed by a one-off
  `kubectl apply -k`, not routed through a release. Same path is open to this fix.

- `deployments/kustomize/services/bugs-open-staleness-sweep/` (RFC_005 §3.3) — the
  precedent for reading **live repo content from a CronJob with no git clone and no
  kubectl exec**: GitHub Contents/Trees API with a read-only token
  (`GITHUB_READ_TOKEN` secret), ref pinned by an explicit env var
  (`SWEEP_REF`) that a human bumps by hand when the platform's live working branch
  changes — same discipline as `IMAGE_TAG`. `fetch_tree(sha)` lists every blob path
  at a ref; `fetch_raw(path, ref)` gets one file's raw content. Stdlib Python only.

- RFC_022 / `optional-key-budget-check` — the "a clean run must still write a row"
  discipline, ruled by the owner after a blind spot cost real damage. Cited in the
  bug file itself as the shape to copy.

## The one real design problem this bug's fix has that those precedents didn't

`single-owner-carriers-check` only ever needed **live DB state** (one SQL query).
`bugs-open-staleness-sweep` only ever needed **file existence and line counts** —
metadata, not real content, for citation-staleness checking.

This check needs the **actual SQL text** of every pending migration, because the
"applied by hand, unrecorded" signal comes from *executing* each pending file inside
a doomed transaction and reading whatever its own guard `RAISE`s
(`run-migrations.sh:145-192`, the probe). There is no way to get that signal without
running the real file content — metadata alone (a filename list) cannot distinguish
"genuinely not yet applied, waiting its turn" from "applied by hand and never
recorded", which is the entire subject of this bug.

**Reuse constraint, load-bearing:** the bug's own §4 rules out a second copy of the
migration vocabulary (naming pattern, sidecar exclusion, baseline) — that is exactly
the defect `bugs_closed/314` exists to remove, and CLAUDE.md's "reuse existing
machinery" applies directly. So the design should **drive
`scripts/migration/run-migrations.sh` itself, unmodified**, not reimplement its
logic in Python inside a CronJob. The two env vars it already exposes for exactly
this purpose: `PSQL_CMD` (override the `kubectl exec` default with a direct `psql`
invocation, matching what `single-owner-carriers-check` already does) and
`MIGRATIONS_DIR` (point it at a local directory holding the real files).

Getting real file content onto local disk in the CronJob container, without a full
`git clone` (`bugs-open-staleness-sweep`'s comment: the whole repo's `.git` is
262M) is the open question for the plan — options: (a) GitHub Contents API,
one `fetch_raw` per pending file (bounded: ~161 today, well under any rate limit,
same mechanism `bugs-open-staleness-sweep` already uses), needing only filenames
(cheap, from the Trees API) plus content for the pending subset; (b) a **partial,
shallow, sparse** `git clone` (`--filter=blob:none --depth 1 --sparse`) scoped to
`docs/agent_docs/sql_for_agents/` + `scripts/migration/`, which avoids the 262M
problem entirely by never fetching full history and gets real files as ordinary
files with zero re-implementation. Handed both to the fable planning pass to weigh.

## Cross-lane check

Grepped `MEMORY_workstreams.md` and `docs024_key_docs_latest/*/HANDOFF*.md` /
`README_where_we_are.md` for "migration ledger" / "unrecorded migration" /
"record-only" — no other lane is building this mechanism. The only two lanes with
first-hand context are `bugfix_314_council_scope` (filed it, explicitly deferred)
and `dispatch_throughput` (found + fixed the original seven, credited in the bug
file). Neither has open work items against this specific fix as of this session's
start.

## Schedule slot

Surveyed every live CronJob's schedule (`grep schedule: deployments/kustomize/services/*/base/cronjob.yaml`,
cross-checked against `kubectl get cronjobs`). Daily checks cluster at `:20`–`:55`
past 06:00 and 07:00 UTC. **07:45 and 07:50 UTC are the only free slots** in that
range. Picking one for the new job, clear of `database-backup` (02:00) and the
weekly `bugs-open-staleness-sweep` (Sunday 06:00), matching the stated discipline
in `single-owner-carriers-check`'s own comment.
