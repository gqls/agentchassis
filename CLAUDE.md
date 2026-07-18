# CLAUDE.md — agentchassis

Many Claude sessions work this repo and this cluster **concurrently**: one working
tree, one branch, one image tag sequence, one live database. You can see your own
actions; you cannot see any other session's, in-flight or committed, except by
looking. Full problem statement and evidence:
`docs/agent_docs/docs024_key_docs_latest/multi_session_coordination/HANDOFF_2026-07-16_multi_session_coordination.md`

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

## Council review of platform changes (advisory, live 2026-07-17)

Any thread can put a change through the fix loop's own reviewer council before
committing it. Advisory: it records a verdict, it cannot block you. Scope is
`platform/`, `internal/`, `pkg/` — docs and site content are refused client-side
and never spend credits. Full runbook + submission schema:
`docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/RUNBOOK_council_gate.md`.

- **Submit**, from a JSON file holding `rationale` (the real why — reviewers judge
  the plan against it) and a `plan` (≤8 edits, each with file/operation/rationale/
  sketch; real diff hunks welcome; plus `grounded_in` evidence quotes):
  `./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>`
  Save the printed `SUBMISSION_CORR`. A run takes ~2 minutes.
- **Verdicts.** APPROVED → commit with a trailer line `Council-Reviewed: <id>`
  (that trailer is what makes the coverage report's commit↔verdict join exact).
  The `<id>` may be **either** the gate's `SUBMISSION_CORR` **or** the
  orchestration id of a **fix-proposer** council run (`RUN_ORCH_ID`) — a fix the
  fix loop's own council approved counts as reviewed, and the report resolves
  both (prefix match, so a short id is fine).
  REVISE → the objections come back with the reviewers' own read-only checks
  already answered; revise and resubmit with `RESUBMIT_CORR=<corr>` so the trail
  accumulates. REJECTED → a guardian veto; its notes name the safest contained
  alternative. Read it: `SELECT body FROM doc_notes WHERE categories ?
  'council-gate' ORDER BY created_at DESC LIMIT 1;`
- **Cost is relevance-gated**, so submitting is cheaper than it looks: two seats
  always run (edit-quality, guardian); the rest — 11 as of 2026-07-18, and
  growing — fire only when your edited paths match their footprint. One council
  run per coherent task, not per iteration.
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

## Diagnosis before debugging (opt-in, by judgement — not a gate)

The same loop can diagnose a bug *before* you fix it — read the real code + live
DB, form a cited theory, and follow the evidence to the cause (which often lives
in shared infra named nothing like the symptom). This is the one thing the
council gate above cannot do: the gate reviews the fix you wrote; only the
diagnosis loop tells you the cause isn't where you're looking.

**It is NOT a gate and NOT a default.** For a bug you can see, debug directly —
you have full context and will out-diagnose the loop faster and for free (this
platform's bugs mostly dissolve under grep + a schema read; that is the whole
reason the loop's value is unattended cited diagnosis, not discovery). Reaching
for the loop on every bug front-loads minutes + credits before you know it is
hard, and risks diagnosing a premise you are one commit from changing.

**File it to the loop first (090 trigger) when ANY of these hold:**
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
stale); one coherent bug per run. Minutes + credits per run; the 090 trigger
already refuses if another thread has open work on the target (FORCE=1 overrides
after you read the findings).

**You often do not need to do this yourself.** The immune system already sweeps
every recorded failure fleet-wide (triage + silent-check) and routes genuine
platform-wide code bugs into the diagnosis queue automatically — so the
cross-cutting class is largely covered without a manual run. Check the queue
before filing: `SELECT summary, status FROM site_work_items WHERE
item_type='needs_diagnosis' AND status='awaiting_diagnosis';`

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
  tagged `IMAGE_TAG`. Nothing downstream of the build records whether it came
  from a commit — one more reason the build itself must be the committed one,
  and verified against the pod.
- Bump `IMAGE_TAG` (makefile ~line 16) for every build — a same-tag rebuild
  ships the node's stale cached binary.
- Verify a deploy against the **running pod**, never git, never the tag:
  `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<your symbol>"'`
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
- **When you diagnose something durable**: file the case in
  `/bugs_open/NNN_HANDOFF_<date>_<slug>.md` (evidence, root cause, fix candidates,
  how to verify) AND add the transferable pattern to 016b §9. The case file is for
  the fixing thread; the §9 entry is so nobody re-walks it.
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
- Schema first: `\d <table>` before writing SQL; read the function before
  changing it.
- Go changes are inert until an image is rebuilt and rolled; DB config is live
  immediately.
- DB access:
  `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
