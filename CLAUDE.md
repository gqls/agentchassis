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

## Platform conventions

- Go, not Python. British English. Structural fixes over patches. Reuse existing
  machinery before building new.
- Schema first: `\d <table>` before writing SQL; read the function before
  changing it.
- Go changes are inert until an image is rebuilt and rolled; DB config is live
  immediately.
- DB access:
  `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
