# CLAUDE.md — agentchassis

Many Claude sessions work this repo and this cluster **concurrently**: one working
tree, one branch, one image tag sequence, one live database. You can see your own
actions; you cannot see any other session's, in-flight or committed, except by
looking. Full problem statement and evidence:
`docs/agent_docs/docs024_key_docs_latest/multi_session_coordination/HANDOFF_2026-07-16_multi_session_coordination.md`

## Explaining decisions and mechanisms to the owner (owner note, 2026-08-12)

When explaining a ruling, mechanism, or decision to the owner — in chat or in docs — don't
compress a definition, the rule that governs it, and its application to the current case into
one dense paragraph. Split it: say what the thing IS in plain terms first (no unexplained
jargon or acronyms), THEN state the rule, THEN say how the current case measures against it.
If a paragraph needs a second read to find the one number that mattered, it is doing too much.

## Git — commit per task (owner ruling, 2026-07-16)

- Commit with an **explicit pathspec**: `git commit <paths your task touched> -m "..."`.
  The pathspec on **`commit`** is the load-bearing part, not the one on `add`: a
  pathspec commit takes those files from the working tree and **ignores the
  index**, so whatever another session has left staged cannot ride along. A bare
  `git commit -m` sweeps their staged files in no matter how careful your `add` was.
- **New files must be `add`ed first** — `git commit <path>` fails on an untracked
  file ("pathspec did not match any file(s) known to git"):
  `git add docs/x/NEW.md && git commit docs/x/NEW.md -m "..."`. Name it twice;
  the `add` makes it trackable, the path on `commit` still excludes everything else.
- Never `git add *`, `git add -A`, `git add .`, or `git commit -a`. One commit per
  task, message names the task.
- A deliberate tidy-up **is** allowed, but must say so: `-m "sweep: leftover docs
  from concurrent threads"`. What destroys review/bisect/revert is four threads'
  work arriving under one thread's message, not breadth as such.
- Forward-only: no resets, no amends, no rebases. Another session may commit
  between your add and your commit — check `git log` before assuming HEAD is yours.
- Your session-start `git status` is a snapshot; it goes stale within minutes.
  Re-run it before acting on it.
- **The auto-memory directory is versioned too** (2026-07-20). Every Write/Edit
  under `~/.claude/projects/*/memory/` is auto-committed into a git repo inside
  that directory by `scripts/memory-git-snapshot.py` (a PostToolUse hook; commits
  one file by pathspec, attributed by session). It exists because a session
  overwrote another session's memory file with `cat >` and it was
  **unrecoverable** — the same mistake in this repo would have cost nothing,
  because git had a copy. Recover with `git -C <memory dir> log/show`. Note the
  hook cannot see a `cat >` that never goes through Write/Edit, so it captures
  the *previous* state, not the clobbering write — which is exactly what you need
  to restore. **Read before write on any file you did not create; prefer the
  Write tool, which refuses an unread file, over a shell redirect, which does
  not.**
- **`git stash` is FORBIDDEN on this tree (owner ruling, 2026-08-14) — and mechanically
  blocked.** A bare `git stash` revert-sweeps EVERY session's dirty tracked files at once,
  not just yours: measured 2026-08-12, one stash took **38 files across ~10 lanes** and put
  the 18 production overlay manifests back ~50–100 releases — after which `git status` read
  clean and the tree matched HEAD, so the next `apply -k` would have been a silent fleet
  rollback. Git has no pre-stash hook, so the ban is enforced at the session harness
  instead: a `PreToolUse` hook (`scripts/block-git-stash.py`, wired in
  `.claude/settings.json`, self-test `--self-test`) refuses every **mutating** form —
  bare/push/save/pop/apply/drop/clear/branch — in any session, however the command is
  compounded. Read-only `git stash list` / `git stash show` stay allowed: they are how you
  READ a stash an earlier session left. Need what a stash holds? Extract **by path** —
  `git show 'stash@{N}:<path>' > <path>` — **never pop**: a shared-tree stash holds
  everyone's WIP, and popping dumps it all into your tree to be committed under your name.
  Wanted a clean tree for yourself? That is what committing per task (above) and
  `git worktree` are for. Full trap + per-file recovery: grep `git stash` in `LANDMINES.md`.
- **Your uncommitted work is not safe, and this practice does not make it safe.**
  Committing per task stops *you* sweeping up *others'* WIP; it cannot stop a
  session that still runs `git add -A` from sweeping up *yours*, half-finished,
  into a commit about something else entirely. This is not hypothetical — it
  happened to this file's own makefile change on 2026-07-16 (`69d6f3ecc`,
  a vet-med-export commit).
  **So: commit each task the moment it is coherent, narrowly.** A long-lived
  dirty tree is not a private workspace — it is shared, mutable state.
  If your work does get swept into someone's commit: nothing is lost, forward-only
  still holds. Finish the task and commit the remainder; say so in the message.
- **Every commit now prints a yellow "commit scope" block** listing what it
  actually contains, grouped by area (`.githooks/pre-commit` →
  `scripts/commit-scope-report.sh`). **This is advisory, not an error** — it never
  blocks. Read it: any file listed that is not part of your task belongs to another
  session, and a pathspec commit is how you leave it out. It deliberately applies
  no threshold — breadth does not predict damage (the commit that swept another
  thread's work was an unremarkable 16 files), so it reports rather than judges.
  It cannot see a *same-file* passenger — if two sessions edit one file, whoever
  commits takes both edits, and no hook can prevent that.
- **Checking that committed HEAD still builds: `scripts/verify-head-builds.sh`
  (OPP-008). Do NOT hand-roll `git archive HEAD | tar` — that recipe is why this
  machine keeps running out of space.** Each extract is ~450 MB, and the `rm -rf`
  in every pasted copy is the **setup** half: it clears the tree the run is about
  to use, so it reclaims a tree of *the same name*, and every run picks a new one
  (one session left `headtree`, `headtree2`, `headtree3`, `headfinal`, `ht5`,
  `ht6` — ~2.8 GB in a morning). **73 documents as of 2026-08-24** still spell the
  recipe out; 66 of them never delete anything. The script writes to disk, points
  the Go linker at disk, refuses a tmpfs target *by filesystem type*, and deletes
  its tree on exit.
  `scripts/verify-head-builds.sh [targets]` after committing;
  `--with <file> [--test]` to build your change against HEAD *before* committing.
  **`/tmp` is a 16 GB tmpfs, i.e. RAM** — a full one presents as
  `link: mapping output file failed: no space left on device`, which reads like a
  compiler fault and is not one. Reap abandoned scratch on **both** filesystems
  with `scripts/scratch-report.py [--days N] [--reap]` (OPP-005; dry-run by
  default, `--self-test` proves its guards). Why the disk half matters as much as
  `/tmp` — and why `df -h /tmp` alone will tell you it is fixed when it is not:
  `docs/agent_docs/docs024_key_docs_latest/tmpfs_exhaustion/`.

## Council review of platform changes (advisory, live 2026-07-17)

Any thread can put a change through the fix loop's own reviewer council before
committing it. Advisory: it records a verdict, it cannot block you. Scope is
`platform/`, `internal/`, `pkg/`, **plus appliable DB migrations
(`docs/agent_docs/sql_for_agents/NNN_name.sql`) — widened 2026-08-19, `bugs_open/314`:
a migration IS the running system, live the moment it applies, with no image tag to
roll back, and 67% of migration-shipping commits could previously only be reviewed
with `FORCE=1`.** Prose, site content and the SQL that is **not the change** (`_ROLLBACK`, `_VERIFY`,
`_SUPERSEDED`) are still refused client-side and never spend credits. **`_HOLD.sql` IS in
scope** — it is the change, held back from the runner for ordering and applied by hand
(excluding it was a real defect in the first cut, caught by the council).

**Two further widenings, both owner rulings, both TARGETED at where detector logic
sits rather than loosening a directory:** `cmd/config-key-audit/` (2026-08-23 — the
check fleet keeps its rules there, and both commits shipping one such check contained
zero in-scope files), and **`scripts/pattern-check.py` (2026-08-24)**. The second is the
same argument one level along: `[MEASURED 2026-08-24]` that one file is **2,058 lines
carrying 22 checks**, against **2,220 lines for every other `audit-*`/`check-*` script
under `scripts/` combined** — so it is about half the non-Go detector surface, and the
half that runs most often, firing on **every commit in every session** via
`.githooks/pre-commit` rather than once a night. Nothing else under `scripts/` is in
scope: `scripts/kafka-publish-lib.sh` is refused, and so is
`scripts/audit-advisory-findings.py` even though it imports this file's checks.

**The scope is single-sourced in `scripts/council-scope.sh`** for the **decision**, and
097 plus the commit-msg nudge read only that. ⚠ **But "single-sourced" does NOT mean one
edit widens it.** `098_REPORT` has to **enumerate** candidate commits before it can judge
them, and `git log` takes pathspecs, not regexes — so it carries `SCOPE_PATHS`, a
hand-kept array, as a pre-filter. **A path added to the regex and not to that array is
INVISIBLE to the coverage report** — not listed as unreviewed, absent, which reads as
nothing to report. Measured 2026-08-23, the day `cmd/` was added: **22 in-scope commits
across four lanes sat in no bucket at all.** **Widen both, in one commit.**
`DRY_RUN=1 097_TRIGGER…` tests admission for free. Full runbook + submission schema:
`docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/RUNBOOK_council_gate.md`.

**Approval is now reachable, so this is a real norm — not a formality (owner
ruling 2026-07-24: strengthen advisory, defer enforcement).** Until 2026-07-22 the
decision rule ignored objection severity, so an APPROVED verdict was effectively
unreachable (~5%) and submitting felt pointless. That is fixed and live
(`bugs_closed/057`, chassis v1.0.1149): approval ran **~80%** over the two days after
— a sound platform change now passes, usually quickly, in ~30 minutes. **So: put
platform-code changes through the gate before (or alongside) committing them.** It
stays ADVISORY — it cannot block your commit, and PR-mode enforcement is still
deferred — but a platform-code commit with no `Council-Reviewed:` trailer now draws
an advisory `commit-msg` nudge and is listed by the `098` coverage report. One run
per coherent task, not per iteration.

- **Submit**, from a JSON file holding `rationale` (the real why — reviewers judge
  the plan against it) and a `plan` (≤8 edits, each with file/operation/rationale/
  sketch; real diff hunks welcome; plus `grounded_in` evidence quotes):
  `./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>`
  Save the printed `SUBMISSION_CORR`. **Budget ~30 minutes, not ~2.** The council
  itself takes 2–5 minutes, but the dispatch queues behind the fleet: measured
  2026-07-20, publish→run start was **29 minutes** under normal load. A missing
  orchestration row is almost always latency, not a dropped dispatch — do not
  retry on that evidence (it costs a duplicate round), and find your run by
  payload, not by the printed id:
  `SELECT current_step, status FROM orchestration_states WHERE
   collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';`
- **Committing before the verdict lands? Use `Council-Submitted: <corr>`** (added
  2026-07-30, from `bugs_open/138`'s round). It asserts nothing, so it can never be
  a false claim, and `098` resolves the correlation at **report** time — the commit
  is credited automatically once the verdict turns approved, with no amend
  (forward-only forbids one). This closes a real hole: the 2026-07-20 rule "commit
  the moment it is coherent, never hold code for a verdict" and a trailer that can
  only be written *after* approval were **mutually unsatisfiable**, so a thread that
  did everything right still read as un-reviewed for ever. **Never write
  `Council-Reviewed:` on a verdict you have not read** — that is a MISMATCH, which
  is the coverage report's dishonesty surface. Both trailers on one commit →
  `Council-Reviewed:` wins.
- **Verdicts.** APPROVED → commit with a trailer line `Council-Reviewed: <id>`
  (that trailer is what makes the coverage report's commit↔verdict join exact).
  **Use `SUBMISSION_CORR`** — the correlation is the key the artifacts are
  written under, so it always resolves. (A council *run* id also resolves, but
  only if it is the id the artifacts actually carry; take that from the DB, not
  from a trigger's printout. Prefix match, so a short id is fine.)
  REVISE → the objections come back with the reviewers' own read-only checks
  already answered; revise and resubmit with `RESUBMIT_CORR=<corr>` so the trail
  accumulates. REJECTED → a guardian veto; its notes name the safest contained
  alternative. Read it: `SELECT body FROM doc_notes WHERE categories ?
  'council-gate' ORDER BY created_at DESC LIMIT 1;`
- **Cost is relevance-gated**, so submitting is cheaper than it looks: two seats
  always run (edit-quality, guardian); the rest — 14 of 16 as of 2026-07-20, and
  still growing — fire only when your edited paths match their footprint (a real
  submission that day drew 10 of 16). One council run per coherent task, not per
  iteration. Read the live count rather than trusting this line:
  `SELECT jsonb_array_length(default_config->'workflow'->'steps'->'council_decide'->'config'->'review_fields')
   FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;`
- **Coverage** (who reviewed, who didn't):
  `./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/098_REPORT_unreviewed_commits_v1.sh [days]`
- **If you add or change a council seat**, seat `fix-proposer` as usual, then run
  the mirror — do not hand-patch the gate:
  `python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py`
  (dry run; `--apply` writes, taking a snapshot first). It copies every
  `review_*`/`gate_*` step and the footprint map from the live `fix-proposer` row,
  swapping the diagnosis context for the submitter's rationale. Two
  hand-maintained rosters that must stay identical is exactly the drift class this
  council reviews for — the roster changed five times in 18 hours, so it is
  mechanical now.
  > **⚠ `--apply` SUSPENDED 2026-08-11 — DRY RUN ONLY.** The mirror regenerates all
  > 17 gate prompts and its transform **predates migration 377**, which hoisted the
  > shared evidence block to the front of every `council-gate` seat and inserted
  > `<!--CACHE_BREAKPOINT-->` so Anthropic prefix caching can fire. `--apply`
  > reverts that on all 17 seats and takes with it a **measured 68% saving** on what
  > was ~85% of fleet LLM spend. Its dry run prints
  > `drift (steps that would change): [all 17]`, which reads as "the gate has
  > drifted" and actually means **"the gate is AHEAD of the mirror"** — every
  > untouched seat is exactly +37 chars, and that is 377's arithmetic, not
  > divergence.
  > **Until `099` learns about 377:** migrate `fix-proposer` as usual, then mirror
  > into `council-gate` with a **second, surgical migration anchored on a verbatim
  > line**, guarded to abort if the breakpoint moves or the shared prefix
  > fragments. Worked pair: `381` + `383` (RFC_022's narrowing). Health check
  > (**17 seats marked, 1 distinct prefix**) and the full trap: `LANDMINES.md`,
  > footprint `099_SYNC_gate_roster.py`. **A documented exception, not a licence to
  > hand-patch — fix `099` and it ends.**

### RFC_022 — the architecture seat's trigger, narrowed (OWNER RULING, 2026-08-11)

**An opt-in field whose unsafe default is OFF, and which no live consumer names, is
NOT architecture-scope.** Do not expect `needs_rfc` for that shape, and do not treat
it as one when you are the reviewer. This makes the 2026-08-02 §2 ruling below
self-consistent: it *prescribes* opt-in-default-OFF as the way to ship new authority
on a shared seam, and the seat's shape-based trigger was firing on exactly that
remedy — so following the rule and ignoring it drew the same signal, which is how a
signal stops discriminating. All three conditions must hold (opt-in; the **unsafe**
side is the default; zero live consumers name it), and **enumerate the consumers —
asserting it without the query is itself the objection**.

**The ruling is option (3) with option (1) as the interim, and the interim is not the
destination.** What (1) gives up is **accumulation**: ten individually inert opt-in
fields are a shared action nobody understands, and this trigger was the only thing
that would have noticed the tenth. (3) gets it back by triggering on the accumulated
optional-key **count** — ~~but that counter (a sweep over `RegisterActionInputSpec`
declarations per action) **is not built**. Whoever builds it closes `RFC_022`; until
then the estate runs with a stated blind spot.~~ **CORRECTED 2026-08-17: RFC_022 is
CLOSED and the whole mechanism is LIVE — this paragraph told every session the
opposite for three days.** The counter was built 2026-08-13 (`cmd/config-key-audit
--optional-key-budget` / `scripts/audit-optional-key-budget.sh [--json] [N]`,
register **WFA-013**); the owner ruled **N = 10** on 2026-08-14; and the automatic
half runs daily (`50 6 * * *` UTC, CronJob `optional-key-budget-check`, live since
2026-08-14 and writing ONE `doc_notes` row per run — on clean results too, so a
MISSING row means the job did not run and must not read as "nothing is wrong").
**So: do not build it, and do not repeat the blind-spot claim.** What an author owes
now is only to keep the cron's literal in step — two actions (`retract_asset_files`,
`publish_site`) entered the registry counted as **ZERO** and were invisible to the
check until 2026-08-17; the parity test (`cmd/config-key-audit/optional_budget_cron_parity_test.go`)
catches it, so RUN IT, and after editing `check.py` re-apply the kustomize overlay or
the cluster keeps the old literal. An action past N owes ONE review of its accumulated
surface, recorded in `architecture_review/optional_key_budget_acks.json` (the source of
truth) with the review it points at; three are acknowledged today. Live in both rosters
via migrations `381` (fix-proposer) and `383` (council-gate), updated by `402`/`403`.

### Platform seams and the ordering exemption (OWNER RULING, 2026-07-28)

A change that adds or alters a **shared** mechanism is an **architecture-scope**
change even when it is additive, small and well tested — a new namespace or
reserved key on a shared action, a new contract, anything whose blast radius is
"every pipeline that uses X". The guardian seat will veto one that arrives inside
a bug patch, and it is **right to**: `bugs_closed/124` shipped the `$ctx.`
parameter namespace that way and drew a REJECTED verdict on exactly that ground.

**The ruling is that the code stays and the precedent gets fixed.** So, from now:

- ~~**A platform seam may ship ahead of its review only when BOTH hold.** (1) There
  is a **real, stated ordering constraint** — the config or data half cannot be
  applied against the old binary without breaking something live. Name it in the
  commit message; "it was convenient" is not one.~~ **Condition (1) SUPERSEDED by
  the owner ruling of 2026-07-29 (below); condition (2) STANDS and is now the
  whole of the requirement.** (2) The seam is **registered in
  the concept register in the same commit that ships it**, with its landmine and
  the open review question written down. Not "later" — later is how a seam becomes
  folklore.
- **Measure the blast-radius claim before you submit; do not ask the reviewer to.**
  124's submission listed "verify no collision" as a risk for the council to check.
  That is not evidence. The check was one query and the answer was decisive
  (63 `params` entries fleet-wide, exactly one `$`-prefixed — the new one). **"No
  collision is possible" is a query, not an argument.**
- **A veto on SCOPE is not answered by resubmitting with better measurements.** It
  is a judgement about *how* a capability reached production. Record it where the
  change lives, route the seam to architecture review on its own merits, and let a
  human break it — especially when seats disagree with each other, which they did
  here: the guardian's contained alternative was precisely what the `reuse_agent`
  seat objected to in the same round.

### OWNER RULING 2026-07-29 — three answers, and one of them retires a rule above

Raised by `architecture_review/RFC_002_criteria_check_type_vocabulary.md` after the
council gate's `architecture` and `guardian` seats reached opposite defensible
conclusions about the same change in the same round.

1. **An addition to a shared vocabulary needs an RFC only when it changes what the
   shared mechanism GUARANTEES** — not merely because the vocabulary is shared. The
   worked case: two new criteria check types made the Tier 2 evaluator able to
   **refute** (fail a page for what it serves) where its stated rule had been
   "confirm, never refute". *That* is the RFC trigger. A type that only adds an
   opt-in capability, reachable by nothing until a document names it, goes through
   the normal council gate. So the opening paragraph of this section — "even when it
   is additive, small and well tested" — is **narrowed**: additive-and-inert is not
   the same as additive-and-guarantee-changing, and only the second is
   architecture-scope.

2. **CONDITION (1) OF THE ORDERING EXEMPTION IS RETIRED, because it asks for
   something no thread can supply.** It assumed a thread could hold a change out of
   the fleet and was choosing not to. On this tree it cannot: HEAD is shared,
   `make build-*` builds from committed HEAD, and any other session's roll ships
   your commit. RFC 002's own case is the proof — the change was submitted before it
   was committed, explicitly disclaimed an ordering constraint, and went live anyway
   on another session's build. **The only mechanism that actually holds a seam back
   is a default-OFF switch, and the owner has ruled we will NOT require one** (its
   cost is a mechanism rotting unexercised, which this platform has been bitten by
   before). So: **review here is after the fact, by design.** Do not claim an
   ordering constraint you do not have; do not pretend you could have waited.
   What is still required is condition (2) — registration in the same commit — plus
   submitting to the gate before or alongside the commit.

3. **A shared mechanism's OTHER consumers must be told, not merely measured.**
   Measuring that zero existing documents are affected proves nothing breaks; it
   does not establish that the other pipeline's owners would have agreed. Name the
   consumers in the submission, and tell them — the useful message is what changed
   about their guarantee, not a list of your new keys.

Worked example, with the evidence and the three options costed:
`docs/agent_docs/docs024_key_docs_latest/bugfix_124_double_dispatch/REVIEW_2026-07-28_ctx_namespace.md`.

### OWNER RULING 2026-08-02 — two narrowings, from `architecture_review/RFC_010`

Raised by the `bugs_open/175` lane after the council's `architecture` seat objected on
record (not to block) that a new four-outcome collision seam plus a work-item contract
change belonged in this track rather than in a bug's risks block.

1. **Converging N producers onto ONE `item_type` / `item_key` does NOT need an RFC** —
   *provided the producer set is named in the concept-register entry and the shared
   `item_key` shape is stated there.* De-duplication is behaviour this estate wants
   cheap, and taxing it with an architecture round makes the second producer quietly
   invent its own key instead. The condition is what keeps it honest: a reader must be
   able to see who files the type, and how it dedupes, without reading every call site.
   This narrows the section above — a work-item type with **no automated consumer** is
   not the kind of shared vocabulary whose *guarantees* change when a producer is added.

2. **New authority on a shared seam ships as an OPT-IN FIELD, not a documented
   contract.** RFC_010's seam could re-type a never-served page, and was safe only
   because every caller's role is a compile-time constant — a rule stated in a doc
   comment and enforced by review. Five council seats independently flagged it. The
   ruling: when a seam's widest branch is licensed by "callers must all be X", make X a
   field with the unsafe default OFF. It costs about four lines, it moves the decision
   to where a reviewer of the CALLER can see it, and it is the only version that
   survives a session that did not read the helper. **A comment is not a control on a
   tree this many sessions share.**

## Diagnosis before debugging (the DEFAULT for any durable claim)

The same loop can diagnose a bug *before* you fix it — read the real code + live
DB, form a cited theory, and follow the evidence to the cause (which often lives
in shared infra named nothing like the symptom). This is the one thing the
council gate above cannot do: the gate reviews the fix you wrote; only the
diagnosis loop tells you the cause isn't where you're looking.

**The test is not how hard the bug feels — it is what you are about to assert.**
File it BEFORE you commit to a root cause whenever the claim is durable: a
mechanism, a structural property of the platform, a cause that lives outside the
symptom, or a fix that changes behaviour beyond the one site in front of you.
Debug directly only when the fix is local and **self-evidencing** — you can watch
it fail, change it, watch it pass, and nothing outside that file depends on your
being right.

> **CORRECTED 2026-07-19 — this section used to say the opposite, and the premise
> was tested and failed.** It read: *"It is NOT a gate and NOT a default … you
> have full context and will out-diagnose the loop faster and for free."*
> That day a thread with full context filed a confident structural claim about
> the two rerender paths, built from grep hits whose functions it had never
> opened. The loop **refuted it in 9.5 minutes** by reading the one function the
> thread had skipped (`rerenderLoadSections`), and the refutation held on
> re-check. The same day, the experience-loop council escalated **eight runs
> running** — every escalation correct, each exposing a defect in our own harness
> rather than in the plan under review.
> **Confidence is not a signal.** The wrong claim felt obvious; that is exactly
> why "obvious" cannot be the gate. Full context is no protection, because the
> failure mode is not missing information — it is not looking.

**Always file when ANY of these hold** (the strongest cases, unchanged):
- the cause is still non-obvious after a quick look (grep + read the function);
- you suspect it is cross-cutting / platform-wide, or that the cause is NOT where
  the symptom is (a local fix would then paper over a shared defect — how BUG A
  and BUG B were found);
- you want a cited, auditable diagnosis (a class of bug, a regression you'll be
  asked to justify, a fix that will change behaviour fleet-wide).

```
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```
Symptom-authoring that earns a gradable verdict: state the MECHANISM, then POINT
at the tables/symbols where the evidence lives — assert neither rows nor counts
(the loop fetches and cites them); no downstream-consequence clauses (they go
stale); one coherent bug per run. The 090 trigger already refuses if another
thread has open work on the target (FORCE=1 overrides after you read the
findings).

**Still not a blocking gate, and the cost is real** (minutes + credits per run) —
but spend it *before* you assert, not after you are contradicted. A refuted
hypothesis costs one run. A wrong root cause that reaches a handoff costs every
thread that then believes it, and those are expensive to unpick precisely because
they arrive stated with confidence. **A REFUTED verdict is a success, not a
waste** — it is the cheapest possible place to be wrong. Record it as a visible
correction where you made the claim (see the working-docs rules), naming what
caught it.

**Check the queue first — it may already be filed, by a sweep or by a thread.**
The immune system sweeps every recorded failure fleet-wide (triage +
silent-check) and routes genuine platform-wide code bugs into the diagnosis queue
automatically, so part of the cross-cutting class arrives without a manual run:
`SELECT summary, status FROM site_work_items WHERE item_type='needs_diagnosis'
AND status='awaiting_diagnosis';`
This is a **de-duplication** check, not a reason to skip filing: the sweep only
sees failures the platform already *recorded*. A wrong belief you are about to
write into a handoff has no failure row and no sweep will ever catch it. Also
grep `/bugs_open/` and `/bugs_closed/` for the mechanism before filing — on
2026-07-19 that turned a would-be duplicate into a sharper finding, and then into
a correction when the loop refuted it.

**OWNER RULING 2026-07-31 — a `bugs_open/` file asserting a cross-cutting or
structural root cause is not "filed" until it has been through the loop above,
or the filing session states plainly why it substituted equivalent first-hand
verification.** This turns the "always file when ANY of these hold" list from
guidance a session can reasonably skip under time pressure into a stated norm
with a named escape hatch (a real substitute, but a declared one, not a silent
omission). Raised because `bugs_open/155` was filed on rigorous first-hand
verification alone — reproduced the failure, read the exact code path, confirmed
the fix — without running `090`, which by this section's own criteria it
should have. Run afterward: **CONFIRMED**, first iteration, independently
re-reading the same functions and citing the same lines. See
`architecture_review/RFC_005_targeted_review_for_docs_that_feed_the_fleet.md`
for the fuller discussion (it also covers why this does NOT extend to a
council review of `bugs_open/`/`016b` themselves — wrong tool for prose, and it
would dilute the architecture seat's signal for the platform-code case that
most needs it).

## Working docs — the standing five (owner directive, 2026-07-18; cadence added 2026-07-19)

Any workstream that will outlive one session keeps **five living documents** in
its own directory under `docs/agent_docs/docs024_key_docs_latest/<workstream>/`.
Create them at the START, not at handoff time — the point is that a doc exists
to update while the work is happening. Update them **as you go**; a doc written
only at the end is a report, and reports lose the wrong turns, which are the
expensive part.

**The cadence column is the load-bearing part.** These are not end-of-session
deliverables. Writing them is part of doing the work, not part of finishing it.

| doc | what it holds | when you write it |
|---|---|---|
| `PLAN_<date>_<slug>.md` | design, phasing, decisions **and their reasons**. Corrections to the originating brief live here, marked as corrections — never silently edited away. | at the start; then the moment a decision, correction or resizing lands |
| `RUNBOOK_<slug>.md` | the commands. Every query/command you had to get right, with its gotcha attached. When one changes, change it HERE, not in your scrollback. | the moment a command was hard to get right — not later |
| `NOTES_<slug>.md` | running record, append-only, **newest at the bottom**. What was tried, what the system actually said, and **every misstep: dead ends, wrong turns, mistaken diagnoses, and your own earlier claims in that file that turned out false.** The missteps are not an appendix — they are the point. | at least once per session, and again each time you get something wrong |
| `README_where_we_are.md` | the owner's running **plain-prose log**, append-only, newest at the bottom. What you'd say out loud: what was found, what broke and why, what you decided, what you need a choice on. No jargon, no tables of field names, no file:line unless it genuinely helps a non-specialist. | **frequently — at every natural break where you stop to summarise, present a choice, or explain a bug.** Roughly: if you wrote a substantial reply in chat, it belongs here too |
| `SUMMARY_<date>_<slug>.md` | the milestone read-out, so the owner can talk about the subproject to someone else. Five parts, in this order: **what we're trying to do · where we've come from · what we've done · where we are now · where we're going.** Plain prose, written to be read aloud. **Every summary is a NEW FILE — never an edit of the last one** (see below). | **at real milestones only, and NOT on a clock.** When the read-out would genuinely differ: a phase done, a design changed, a blocking question answered. If "where we are now" would say much the same as the last one, don't write one — expect days between summaries on a steady workstream |

**`README_where_we_are.md` is the owner's document.** He maintains it too.
**Append to it; never rewrite or reorder it**, and never edit his words — add a
dated correction below instead, the same way NOTES corrections work. If it looks
like a pasted chat transcript, that is because it is one and that is fine; match
the register, don't tidy it. (A session mistook it for a stray file and
overwrote it on 2026-07-19 — flagging a file as suspicious is not permission to
replace it.)

Keep the three prose docs distinct or they collapse into one drifting account:
**NOTES** is the technical log (evidence, commands, what the system said),
**README_where_we_are** is the plain-English history (what happened, in order),
**SUMMARY** is current state only (no chronology — that's the other two).

**Write a NEW summary each time; never overwrite the last one** (owner directive,
2026-07-19). Each file stays current-state-only, but **the series is the record.**
`SUMMARY_2026-07-19_x.md`, then `SUMMARY_2026-07-22_x.md`; for a second on the
same day, suffix the date — `SUMMARY_2026-07-19b_x.md`. The concept register
already works this way (`SUMMARY_where_we_are_2026-07-16/17/17b/18.md`) — follow it.

**Rarity is part of the design** (cadence cut 2026-07-20, the day after the
new-file rule went in — the two rules only work together). A daily summary plus a
never-overwrite rule produces a shelf of near-identical files that nobody reads
and that bury the two or three which actually marked a turn. The series is only a
record if each entry is an inflection. **The five headings are the test**: if
answering them would produce substantially the last summary again, the milestone
has not happened yet — put the material in NOTES or `README_where_we_are` and wait.

Why: a summary is what we believed at a milestone, written for the owner to say
out loud to someone else. Overwriting it destroys the only record of how the
understanding actually moved — and that trajectory is often the most useful thing
in the directory, because a summary that turned out to be wrong is evidence about
*how* we get things wrong, which no other doc captures at that altitude. NOTES
records missteps you caught yourself; the summary series shows the ones you did
not catch at the time. **This rule was written because a session overwrote a
summary the same morning it wrote it** — the replacement was accurate and it was
still a loss, so "the new one is better" is not the test.

Rules that make them worth the effort:

- **Record what was wrong, not just what is right.** A wrong turn you diagnosed
  is the most valuable line in the file — it is the one thing the next thread
  cannot rederive. Correct claims in place, visibly (`> **CORRECTED <date>:** …`),
  and say what caught the error.
- **Ground every figure against the live system** before repeating it from
  another doc. Volumes, counts and statuses go stale within days; a figure
  carried forward unchecked is how a stale premise gets diagnosed as a bug.
- **A verified fact needs its evidence inline** — the query, the file:line, the
  pod output. "Verified" without the check is a claim, not a verification.
- **Mark the UNVERIFIED ones too.** The rule above only makes a checked claim
  look checked; it does nothing to make an unchecked one look unchecked, and
  that asymmetry is how every entry in `WRONG_CALLS.md` got written. An
  inference stated in the same voice as a finding *is* the error. So mark it:
  `[INFERRED]`, `[UNMEASURED]`, `[ASSUMED]` — inline, where the claim is. Typing
  the marker is itself the check, because most of the time you will go and do
  the query instead. A durable claim with no marker and no evidence is the
  shape to distrust in your own writing and in anyone else's.
- **The marker is not the whole check — a `[MEASURED]` figure is only evidence if
  the measurement could have come out otherwise.** Dated and marked is not the
  same as disconfirmable. Before recording a figure, name what the disconfirming
  result would have looked like: a detector run against a tree that already
  carries its own fix, or a census filtered on the very column it exists to
  test, produces the same number regardless of what is true. Both happened,
  dated and marked, in one 2026-08-03 session (`WRONG_CALLS.md`) — the marker
  rule was followed in full and neither claim could ever have come out false.
- **A COUNT OF THINGS MUST CARRY THE DATE IT WAS COUNTED (owner ruling, 2026-08-22).**
  Any "N writers / callers / call sites / instances of X" claim — in a bug file, an
  RFC, a register entry, a landmine or a commit message — is written `**N** as of
  <date>`, never a bare `N`. **A census does not go wrong; it goes STALE, by
  ADDITION, and it reads as current for ever.** The worked case: `RFC_008` counted
  ten writers of `page_components.rendered_html` on 2026-08-02 and named two as the
  open gap. `create_tool_component_regenerate.go` was **born 2026-08-19** — so every
  document quoting "ten writers" was correct when written and wrong by birthday,
  including the landmine whose own text warns that *"the SET of writers grows while
  you are not looking"*. The eleventh was found by `DBG-076`, not by any reader of
  those documents (`bugs_open/362` §6a).
  **The date is what makes the staleness mechanically checkable**, which is the whole
  reason for the rule: `git log --since=<census date> --diff-filter=A -- <dir>` lists
  what was added since, and a non-empty result means re-run the census before quoting
  it. Without the date there is no `--since`, and the claim can only be re-derived by
  hand. So the cheap half of this is free and you owe it today: **put the date next to
  the number.**
- **Log the wrong calls: `docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md`.**
  Fleet-wide, append-only — a claim you wrote down that turned out to be false,
  what caught it, and the cheap check that would have. One row is an anecdote;
  the **tally** of skipped checks is the point, because a check that keeps
  appearing is one worth automating. That tally is what earned
  `check_append_only_docs` its place in `scripts/pattern-check.py`. Distinct
  from 016b §9: that file records how the *system* fails, this one how *we* do.
- **Log the landmines: `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`**
  (created 2026-07-29 at the owner's direction). Fleet-wide, append-only, **any
  thread may append.** **It is the system of record** (owner ruling D10,
  2026-07-29) and is **synced into `doc_notes`** by `scripts/landmines-sync.py` so
  council seats and agents can read it — **so append to the file, never hand-write
  a landmine row into `doc_notes`.** The 40 LANDMINE lines still in `MEMORY*.md`
  stay there until delivery is working; that duplication is deliberate, not
  untidiness. A landmine is the rung between the two files above: **a
  trap that fires when you TOUCH a particular file, table, command or service,
  where the wrong result looks exactly like the right one.** WRONG_CALLS is the
  incident, retrospective; a landmine is the distilled check, prospective, and
  you read it *before* you have a symptom. The test for an entry is strict:
  **would a session touching this thing, with no symptom and no suspicion, get it
  wrong without the entry?** If it needs a symptom first it is 016b §9.
  - **You get the path-shaped ones automatically**: a `SessionStart` hook
    (`scripts/landmines-session-start.py`, wired in `.claude/settings.json`)
    matches entries against the files already dirty in the tree. New sessions
    only — a running one needs `/hooks` or a restart. **Still grep it yourself for
    table, command and symbol footprints**, which cannot match a path:
    `grep -n "<path-or-table>" …/LANDMINES.md`.
  - **After you append, run `./scripts/landmines-verify-dispatch.sh`** — it runs the
    sync (so the `doc_notes` rows follow) AND arms the verifier for the new/changed
    entries. ~~`./scripts/landmines-sync.py --apply`~~ alone consumes the "new entry"
    status first, so the verifier then never checks your entry (LANDMINES, "Running
    `landmines-sync.py --apply` before `landmines-verify-dispatch.sh`…" — corrected
    here 2026-08-15 after a session followed this line and hit exactly that). If you
    already applied: `./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>'`
    per entry. `landmines-sync.py --check` exits 1 if the rows have drifted.
  - **Every entry carries a `footprint`** (the path/table/symbol it guards) so
    entries convert mechanically to `doc_notes` rows —
    `architecture_review/PROPOSAL_D9_landmines_as_a_footprinted_corpus.md`, open
    as **D10**, proposes exactly that and this file is written as its staging
    format, not a competitor.
  - **Do NOT drain `MEMORY.md`'s landmine lines into it.** D10 §5: moving them
    before delivery exists removes protection we have today. Both run in
    parallel; duplication is correct until D10 is decided.
- **Point at bugs, don't restate them.** Durable defects belong in `/bugs_open/`
  (see Debugging below) — link them; do not fork a second account that drifts.
- **Grep before you file.** `/bugs_open/` and the workstream dirs first: several
  threads work concurrently and may have found it hours ago. This has already
  prevented duplicate bug files.

## Dispatching work at the cluster

- **Checking the pod does not check the queue.** Before firing a diagnosis or fix
  at a target, check for open work items already touching it — another session
  may have a fix in flight (this cost a real diagnosis run on 2026-07-16):
  `... FROM site_work_items WHERE status NOT IN ('complete','cancelled','rejected') AND <target match>`
- The 090 needs_diagnosis trigger performs this coverage check itself and refuses
  on a hit; `FORCE=1` overrides after you have read the findings.

## Building & deploying images

- **`make build-<service>` builds from committed `HEAD`** (inverted 2026-07-17):
  a `git archive` into a clean context that structurally cannot bundle anyone's
  WIP — yours or another session's. So the safe build is the one you get by not
  thinking about it. **Commit your task, then build.** All 14 backend services;
  frontends build from their own context and are unaffected.
- If you forgot to commit, the build prints how many uncommitted changes it is
  **leaving out** and continues — so you get an image missing *your* change (a
  wasted cycle, caught by the pod-grep below) rather than one silently carrying
  everyone else's untested work to production. Commit and rebuild.
- `make build-<service> REF=<ref>` pins a specific commit. `make build-<service>-tree`
  is the deliberate escape hatch that builds the **working tree**, WIP and all —
  only when you actually want uncommitted code in the image.
- `push-*` and `deploy-*` are git-blind: they ship whatever image is locally
  tagged `IMAGE_TAG`. ~~Nothing downstream of the build records whether it came
  from a commit~~ — **CORRECTED 2026-08-11: the binary now says so itself.**
  Every backend binary and image carries the commit it was built from
  (`bugs_open/153`, register **BLD-019**, live since `v1.0.1283`). The build
  must still be the committed one; you can now *check* rather than assume.
- Bump `IMAGE_TAG` (makefile ~line 16) for every build — a same-tag rebuild
  ships the node's stale cached binary.
- **Ask the service what it is running. Do not hunt for markers, and do not use
  `strings`** (rewritten 2026-08-11; the old `strings … | grep -c "<your symbol>"`
  recipe that stood here produced three confidently wrong readings in one day.
  **REORDERED 2026-09-04, `bugs_open/463` lane:** the scrolling log line stood
  first here and the table below was named nowhere in this file, though
  `LANDMINES.md` documents it in nine places — following this section as written
  cost that lane an hour. The corpus had the answer; the file every session loads
  unasked did not):
  ```sql
  SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
   WHERE kind='build' AND pod_name LIKE '<deployment>-%' ORDER BY started_at DESC;
  ```
  A running pod publishes the commit it was built from (register **BLD-023**), so
  **ask this first** — no exec, no binary path, no log window, nothing to install.
  **Filter by `pod_name`, not by the `service` column**: that column also carries
  rows for other pods sharing the image (`agent-landmine-verifier-*` and friends),
  which may have rolled at a different time and be running something else. Then
  **"did my fix ship?" is a query, not an inference:**
  `git merge-base --is-ancestor <your-commit> <the stamp>`. So you no longer need
  to plant a string literal in your change in order to date it later.
  - **⚠ It is a TWO-HOUR WINDOW, not a history** (`RetentionWindow`,
    `platform/buildcapability/buildcapability.go`). It answers *what is running
    NOW*, with no shelf life — and it answers a question about the **PAST** with
    today's survivors, **silently**: a pod whose `started_at` precedes your event
    reads as proof that binary served it, and proves only that this pod outlived
    the prune. If the thing you are dating is more than two hours old, **stop** —
    corroborate against something that is not pruned, e.g.
    `kubectl -n ai-persona-system get rs -l app=<svc> --sort-by=.metadata.creationTimestamp`
    (how 463 established its 22:07:19Z roll time). Full trap: `LANDMINES.md`,
    "…IS A TWO-HOUR WINDOW…".
  - **Per SERVICE, not per fleet** (`bugs_open/249`). Until the pinning fix in
    `release` has rolled, a release could straddle other sessions' commits and
    ship several revisions under one tag — `v1.0.1284` shipped three. Read the
    stamp of the service you actually mean.
  - **The service's own startup line still works — it is second because it
    SCROLLS**, not because it is wrong:
    `kubectl -n ai-persona-system logs -l app=<service> --tail=300 | grep -m1 'build provenance'`.
    On a busy service it is already out of reach hours later (`agent-chassis`,
    measured 2026-08-11: absent from `--tail=3000`), and reading the whole log
    OOM-kills the tool. **An empty result there means "not in range", not
    "unstamped".**
  - **If you doubt the stamp, probe the binary — but verify a KNOWN value:**
    `kubectl -n ai-persona-system exec <pod> -- grep -aq "<expected-sha>" /proc/1/exe`.
    Never `strings` (absent from the debian-slim images, and behind the customary
    `2>/dev/null` its failure is indistinguishable from "not stamped"), and never
    a *discovery* grep for "some 40-hex string" (it matches Go's internal digit
    table and returns the same wrong answer on every service). **Always run a
    control in the same breath** — a sha that must be absent, and one that must
    be present. Both traps are in `LANDMINES.md`.
  - **⚠ A capability probe cannot see code that nothing CALLS — so for INERT
    code, verify by ANCESTRY, never by literal.** Grepping the binary for a
    literal your change added is evidence only for code with a live caller; the
    linker drops an unreachable one, so a built-and-inert phase-1 module probes
    ABSENT with both controls clean (measured 2026-09-02, `bugs_open/440`: two
    fresh pods read PHASE1A-ABSENT while their stamps carried the commit by
    ancestry). Worse, it will spontaneously start hitting the day the first
    caller lands, which mis-dates the ship to *that* roll. Use the stamp plus
    `merge-base`, above.
- Image first, then seeds (a seed naming an unregistered action fails at runtime).
  No orchestration dispatch within ~300s of a chassis pod (re)start — the spawn
  is silently dropped.

## Debugging — read the guide, then file what you learn

- **Before debugging, read the guide**:
  `docs/agent_docs/docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md`
  (§ "Durable invariants" first, then §9 patterns; 016 holds the back-catalogue).
  Most stalled investigations skipped an invariant already written down there.
- **Open bugs live in `/bugs_open/`** (repo root; was `aaa_fails_to_mend/`). §10 of
  the guide indexes them. **Grep that index for the mechanism before filing a new
  bug** — 005/008/009/012 turned out to be one truncation-and-config family found
  by four separate threads.
- **Closed cases live in `/bugs_closed/`** (split out 2026-07-19 so `/bugs_open/`
  answers "what is biting prod right now"). The bar is **fixed AND live** — a fix
  committed but inert until the next roll stays OPEN, because the defect is still
  reproducible until it ships. **Grep BOTH directories before filing.** Numbering
  is one sequence across both dirs, never reassigned, and **several numbers name
  two unrelated cases** (016, 017, 083, 112, 131, **146** as of 2026-07-29, **410** as of 2026-08-26, **420** as of 2026-08-31 — two lanes filed within hours, and BOTH fired on the same day's canary/paid-build work, **456** as of 2026-09-03 — two lanes again within hours (`one_undecodable_fact_disarms_a_whole_evidence_register` and `writer_emitted_a_malformed_closing_tag`), and the second lane's own commit message already says a bare "456 §4" meaning the other one — and the list grows)
  — so a bare number is ambiguous and most commit messages saying "083" mean the
  *other* one. **Resolve by slug, and `git log` the FILE PATH, not the number.**
- **Before routing work AT an existing bug, check who owns it:**
  `scripts/who-owns.py <number|slug>` (advisory, ~0.3s, no cluster calls) — it
  names the workstream most likely working it and separates that from dirs merely
  citing it. "Grep before you file" covers a NEW bug; this is the same failure
  mode for an existing one. On 2026-07-20 a thread promoted `bugs_open/023` with
  implementation direction while its owner was six council rounds in with code
  already shipped. If it says OWNED: contribute into the bug file, do not compete.
  It reads COMMITS, so a session mid-fix is invisible — check the tree too.
- **When you diagnose something durable**: file the case in
  `/bugs_open/NNN_HANDOFF_<date>_<slug>.md` (evidence, root cause, fix candidates,
  how to verify) AND add the transferable pattern to 016b §9. The case file is for
  the fixing thread; the §9 entry is so nobody re-walks it.
- **When you BUILD a new reusable mechanism, register it** in
  `docs/agent_docs/docs026_concept_register/register/<category>.md` — one entry,
  `status` / `what` / `sources` / `relations`, and say so explicitly if it is
  built but never exercised ("deployed" would overstate it). **The bar: another
  workstream could call this and would not know it exists** — a new exported
  helper, agent, seat, check type or delivery path. NOT bug fixes, copy, or
  site-specific work. Update the index count and drop any matching line from
  `102_coverage_ratchet.txt`.
  **Routing, because there are now five destinations:** how the SYSTEM fails (you
  have a symptom) → 016b §9 · how WE were wrong → `WRONG_CALLS.md` · **what will
  mislead you when you TOUCH something (no symptom yet) → `LANDMINES.md`** · what
  THIS workstream is doing → the standing five · **what EXISTS and is callable →
  the concept register.**
  Why this earns its line: extraction froze 2026-07-13 and **67% of workstreams
  postdate it** (`bugs_open/106`), so the index sessions are told to consult
  *before concluding something does not exist* has been going blind ever since.
- Two rules that keep catching real damage:
  - **`output_tokens == max_tokens` means the completion was CUT**, not finished.
    Any agent that rewrites a whole artifact can persist a fragment and report
    success (`bugs_open/012`: a 10,272-char component saved back as 1,253 chars of
    CSS). Check the artifact's structure after a rewrite, not just the status.
  - **Trust the rendered artefact, not the status.** `complete` is not proof the
    work happened; verify against the DB row or the live page.

## Platform conventions

- Go, not Python. British English. Structural fixes over patches. Reuse existing
  machinery before building new.
- **EVERY SITE GOES THROUGH THE FRAMEWORK. Never hand-build one (OWNER RULING,
  2026-08-04).** No hand-authored HTML uploaded to the bucket, however small,
  however temporary, however much faster it would be. Seed the site row and its
  specs (`SEED_*.sql`, the `oufe` file is the worked example), then dispatch
  `scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh
  <domain> --email … --mission-file …`. Raised because a session hand-wrote the
  **webdesign.uk shopfront** and shipped it to `portfolio-sites/` — on the lane
  whose entire product is framework-built sites. Two reasons it is a rule and
  not a preference: a hand-built page **demonstrates nothing** on a site selling
  this capability, and it silently opts out of every control the pipeline
  applies (`evidence_base` claims gating, banned-claim sweeps, the discovery
  checks, imagery style, rerender). **The framework being slower is not a
  reason** — the fast path produces an artefact nobody can audit and nobody can
  rebuild. If the framework cannot yet do it, that is a bug to file, not a
  licence to hand-roll.
- Schema first: `\d <table>` before writing SQL; read the function before
  changing it.
- Go changes are inert until an image is rebuilt and rolled; DB config is live
  immediately.
- DB access:
  `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
