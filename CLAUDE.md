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
- **Every commit now prints a yellow "commit scope" block** listing what it
  actually contains, grouped by area (`.githooks/pre-commit` →
  `scripts/commit-scope-report.sh`). **This is advisory, not an error** — it never
  blocks. Read it: any file listed that is not part of your task belongs to another
  session, and a pathspec commit is how you leave it out. It deliberately applies
  no threshold — breadth does not predict damage (the commit that swept another
  thread's work was an unremarkable 16 files), so it reports rather than judges.
  It cannot see a *same-file* passenger — if two sessions edit one file, whoever
  commits takes both edits, and no hook can prevent that.

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
- **Closed cases live in `/bugs_closed/`** (split out 2026-07-19 so `/bugs_open/`
  answers "what is biting prod right now"). The bar is **fixed AND live** — a fix
  that is committed but inert until the next image roll stays OPEN, because the
  defect is still reproducible until it ships. **Grep BOTH directories before
  filing**, and note the two traps in `/bugs_closed/README.md`: numbering is one
  sequence shared across both dirs and is never reassigned (so a stale
  `bugs_open/NNN` or `aaa_fails_to_mend/NNN` pointer resolves by number in the
  other dir), and **`016` and `017` are each used by two different cases** — a
  bare number is ambiguous, resolve by slug.
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
