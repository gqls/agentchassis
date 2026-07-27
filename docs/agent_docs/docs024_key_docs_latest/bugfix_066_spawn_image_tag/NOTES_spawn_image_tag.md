# NOTES — 066 spawn image tags (append-only, newest at the bottom)

## 2026-07-27 — session open

`scripts/who-owns.py 066` said OWNED by `fixloop_eg_dartsonline` (165 commits/14d). Read
their docs before touching anything: their `PLAN` and `HANDOFF_2026-07-25` both say the same
thing — *"Hand `bugs_open/066` to whoever owns deploys. It is the one shakeout bug still able
to bite silently."* So it is handed off, not being worked. No open `site_work_items` matched
`image_tag`; no other bug file in `/bugs_open/` or `/bugs_closed/` carries the mechanism.
Proceeded.

## The first reading of the makefile was wrong, and it reached the owner

I built a table of "five copies of the tag and who maintains each", concluded the deploy-time
sync was *"wired into the roll zero times"*, and offered it to the owner as the evidence for
choosing a route. They chose on it.

**It was false.** `deploy-agents` calls `update-agent-images-v2` at `makefile:1028`, which
runs `UPDATE agent_definitions SET image_repository = …, image_tag = …`. I found out when I
went to *edit* the tail of that recipe — 111 lines below the target header, past ~20
near-identical `kubectl apply -k` blocks.

What I had actually run was `grep -n "IMAGE_TAG\|image_tag\|deploy-agents" makefile` (output
dominated by 40 `newTag:` sed lines) and `sed -n '917,1010p'` on a recipe that runs to 1035.
The check that would have caught it in one command: **`grep -n "agent_definitions" makefile`**
— grep the *object of the claim*, not the *subject of your search*. This is
`narrow-filter-defines-the-conclusion` again; I have it in memory and did not apply it.

Told the owner immediately rather than quietly correcting the plan. Their decision ("both
halves") did not change, but half 2 changed shape: from *add a sync* to *consolidate five
unscoped ones*.

**The correction improved the fix.** "Someone forgot to wire a target" is a wiring problem
with a wiring answer. "The sync ran and the rows went stale anyway" is a class problem: a
deploy-time sync is a property of one deploy *path*, and `kubectl rollout undo` — the case
where spawned pods most need to follow the chassis *down* — can never be covered by a
forward-only write. That is what makes spawn-time resolution the right answer rather than
merely a tidier one.

## Two live observations that changed the design

1. **The drift runs in both directions.** At 13:44:56 the rows were updated to v1.0.1173;
   the chassis pod's `startTime` was 13:45:31. For 35 seconds the *row* led the *Deployment*,
   and spawned pods created 13:49–14:04 were on v1.0.1173 while the chassis served
   v1.0.1172. The original case file only ever describes the row *trailing*. Both directions
   are now unit-tested, because a fix that only corrects "row is stale" would have been
   silently half-right.

2. **The dead env vars are the argument.** Inside a pod running v1.0.1173:
   `AGENT_IMAGE_TAG=v1.0.82`, `agent_image_tag=v1.0.44`. `grep -rn "AGENT_IMAGE" --include=*.go .`
   returns nothing — no Go code has ever read either. And `v1.0.44` is exactly the default in
   `scripts/deploy/deploy-agents.sh` line 4, so that script has been run at least once and
   nothing has corrected the value since. This is why the fix does not add a third copy.

## Missteps and dead ends

- **Wrote the first version of `agent_image.go` before checking RBAC.** It happened to be
  fine (`pods: get` is already granted) but I designed the "read the Deployment" variant
  first, which would have needed an RBAC change. `kubectl auth can-i --as=…` is ~1s and
  should have come before the design, not after.
- **The drift script's own census printed nothing on its first run** — an aggregate in
  `GROUP BY`, hidden because `psql_q` had `2>/dev/null`. A broken query and "nothing to
  report" looked identical. Ironic given what the script is for; the redirect is gone and
  there is a comment saying why. Then proved the *failing* branch with a bogus tag
  (`v1.0.9999` → 180 rows), because a passing check that cannot fail proves nothing.
- **Nearly built a pin as a schema column plus migration.** The survey killed it: zero rows
  pin today. Deferred to `default_config.pin_image_tag` with a named reversal trigger (the
  first real pin). Noted because the instinct to add the column was strong and wrong.
- **Considered having the chassis write the running tag back over the rows at startup.**
  Rejected: rolling deploys would have old and new pods fighting over the column, and every
  spawned pod runs the same binary, so it would fan out fleet-wide on every spawn.
- **`gofmt -l` on the actions package lists ~11 files.** None are mine — pre-existing, other
  sessions. Confirmed my edit was the only diff in `spawn_actions.go` and that HEAD was
  gofmt-clean before running `gofmt -w` on it, so I did not reformat someone's in-flight work.
- **`go build ./...` fails** on `cmd/reasoningset` (another session's uncommitted WIP) and on
  a stray `docs/.../working_dir` with two package names. Neither is this change; build
  `./platform/... ./internal/... ./pkg/...`.

## State at the end of the session

Committed `c0d7c3a71` (7 files, scope block verified). Council submitted, corr
`3e146ef2-a072-40a8-86be-f6cd940a95f9` — lane was clear, run started immediately rather than
the usual ~30 min queue. **066 stays OPEN**: the fix is inert until a chassis roll past
v1.0.1174, and the defect is reproducible in production until then. The four post-roll
verification steps are written into the case file so whoever rolls can close it.

## Council gate — APPROVED round 1, and four objections worth acting on

Corr `3e146ef2-a072-40a8-86be-f6cd940a95f9`. `approved`, "5 advisory objection(s) — none
high-severity", `unreadable: 0`, `abstained: 7` (the relevance filter, not silence — the
abstention count on the 16-seat gate is filtered seats, so only `unreadable: 0` says every
seat that ran could read the plan). Lane was clear; the run started immediately and finished
in ~8 minutes rather than the ~30 the RUNBOOKs warn about.

Acted on rather than filed, because four of them were right:

- **constitution (medium ×2)** — the SQL was built by string interpolation. Both sites now
  bind with `psql -v` / `:'var'`. **This immediately produced a real trap:** `psql` performs
  `:'var'` interpolation only on input it *lexes* (stdin / `-f`) and sends a `-c` string
  as-is, so the parameterised `-c` form fails with `syntax error at or near ":"`. Reaching
  for `-v` with `-c` **looks** parameterised in review and is not parameterised at all. Both
  forms verified; the makefile now pipes via `printf`. In the RUNBOOK.
- **guardian (low)** — no log on the "repository didn't match, falling back to the row" path.
  Correct, and it is exactly what a registry-host spelling mismatch would do: revert silently
  to the bug being fixed. Now a `Warn` naming the row's repository and every image this pod runs.
- **prior_art_librarian (medium)** — "spawn_group.go:248 is a dead override" was an absence
  claim with no lookup attached, *while the claim right next to it carried a grep*. That is
  the same shape I had already logged against myself in `WRONG_CALLS.md` **this session**, one
  file away, and I still did it. Checked properly: four `"image_tag"` hits in Go, three are
  struct tags, the spawn_group write is the only one and nothing reads it. Claim survives.
- **reuse_agent (medium)** — no evidence the tree was searched for an existing self-pod
  helper. There is none (the only `Pods().Get` calls are job-existence and gate-log fetch),
  **but the search found a house convention I had ignored**: `os.Getenv("POD_NAME")`, used by
  `agentbase/agent.go` and three adapters. Now preferred, hostname kept as fallback — and the
  fallback is the live path, because `POD_NAME` is unset on the chassis Deployment. The seat
  was right for a reason I had not anticipated: not "you duplicated something" but "you
  invented a convention next to an existing one".
- **debug_historian (low)** — enumerate `status`/`is_active` before trusting the blast radius
  (the `sites.status` lesson). 105 experimental+active, 70 active+active, 5
  experimental+**inactive** = the 180. Recorded in the script as a deliberate finding: the 5
  inactive rows are in scope on purpose, since an inactive row can be reactivated.
- **editquality (medium/low)** — edits 5–7 are hygiene bundled into a fix that disclaims them
  as the cause. Accurate. It was the owner's explicit call to do both halves in one pass;
  recorded, not unpicked (forward-only).

One fidelity note for future submissions: the guardian flagged that the drift log string
referenced by `check-agent-image-drift.sh` "isn't in the agent_image.go sketch". It *is* in
the file — my sketch was abridged. **An abridged sketch is a different claim than the code**,
which is the `grounded_in` quote-fidelity rule applied to `sketch`, and I had not thought of
it that way before.

Committed `e96d42226` with `Council-Reviewed: 3e146ef2-…`. The trailer is on the follow-up
commit rather than on `c0d7c3a71` (the fix itself) because forward-only forbids amending —
the 098 report's commit↔verdict join will find the trailer, but a reader should know the
verdict covers both commits.
