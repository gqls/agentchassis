# RUNNING NOTES — multi-session coordination: decisions taken and machinery built

**Session:** "session coordination", 2026-07-16, continuing
`HANDOFF_2026-07-16_multi_session_coordination.md`. All four §7 decisions resolved.
Everything below is committed per task with explicit pathspecs (practising §5).

> **UPDATE 2026-07-17 — the inverted ref build is PROVEN in production, fleet-wide.**
> `make release` was run at `IMAGE_TAG=v1.0.1133` (owner-directed full-fleet
> release). All 14 backend services built from committed `HEAD` via the
> `git archive` ref builds — first real end-to-end exercise of the inversion,
> closing the "never driven through a real docker build" open item. Verified:
> the live `agent-chassis` pod digest (`sha256:ed57728c…8aacde2`) is identical
> to the pushed registry image AND to a separate earlier standalone build of the
> same commit — i.e. the ref build is deterministic. Whole fleet READY=AVAIL.
> Coordination lessons that bit DURING this deploy (all live demonstrations of
> the problem this thread exists for):
> - **The tag was a moving target.** We built `137589be2` intending tag
>   `v1.0.1132`; between build start and finish another session pushed AND
>   deployed a different `v1.0.1132` to prod. Our local `1132` became a
>   clobber-live footgun — neutralised by retagging to the free `v1.0.1133`
>   and dropping the local `1132` tag. LESSON: pick the tag at PUSH time, not
>   build time; a registry check (`docker manifest inspect`) is mandatory
>   pre-push. `IMAGE_TAG` churned 1130→1131→1132→1133 across sessions in one hour.
> - **`make deploy-agents` is FLEET-WIDE, not chassis-only**, and would have
>   ImagePullBackOff'd 8 services lacking a `1133` image had we not first built
>   the whole fleet. There is no chassis-only deploy target — the mixed-tag
>   fleet (chassis 1132, others 1130) was the symptom. `make release` (build all
>   → push all → deploy all) is the coherent whole-fleet path.
> - **Deployed into a busy cluster by owner decision.** ~8 live orchestrations
>   were in flight; the `redeploy-agents` restart dropped them (the 300s
>   rebalance landmine). Flagged and explicitly accepted before firing. Do a
>   `orchestration_states` quiet-check before any fleet restart.
> - Pre-flight that paid off: `go build ./cmd/...` from a clean HEAD archive
>   confirmed all 14 binaries compile BEFORE the ~10-min release, so no
>   half-finished rollout.

## Decision 1 — pre-dispatch coverage check: BUILT and PROVEN LIVE

`090_TRIGGER_needs_diagnosis_v1.sh` (commit `4accef4e3`) now refuses to dispatch
when open work already touches the target, unless `FORCE=1`. It runs before the
intake INSERT, so a refusal writes nothing. Semantics copied verbatim from
`silentCoverageClause` — `status NOT IN ('complete','cancelled','rejected')`,
matched by `page_id`, `spec->>'page_id'`, or `item_key` segment. All probes
exclude the script's own `ITEM_KEY` (same-slug re-runs are the documented
idempotency, not a collision), and fail closed if the DB is unreachable.

Three probes:

1. **`PAGES=<name-or-url,...>`** (new env) — page-keyed, searched across **all**
   sites, because the 2026-07-16 incident spanned finetuning.uk AND
   gamesdesign.co.uk and `RUNTIME_SITE` is singular. BLOCKS.
2. **`SEED_SCOPE` file overlap** — any open `pipeline='diagnose'` item whose
   `spec->'seed_scope'` shares a file (the `:Symbol` suffix is ignored). BLOCKS.
   **This resolves the handoff §6.1 open question: the coverage key for a
   code-only diagnosis is the seed_scope file set, at file granularity.**
   Plus a non-blocking ADVISORY when seed files are dirty in the local tree
   (someone is mid-edit; the diagnosis reads `REF`, not the tree).
3. **Site-level fallback** (site resolved, no `PAGES`) — open items on the
   target site touched within `RECENT_WINDOW` (default 2 hours) BLOCK; the
   older open backlog is a one-line FYI count only. Rationale: blocking on a
   weeks-old parked `needs_human_review` would train operators to always
   `FORCE`, which kills the check.

Validation evidence (all against the live cluster, 2026-07-16 afternoon):

- **Incident replay:** the page probe matches the `json-leak-fix-retry` batch
  rows on both sites (status filter relaxed, since they are now closed) — the
  check would have caught the collision that cost correlation
  `781ea4f7-996d-4b41-be0e-96473c4a7996`.
- **Live refusal:** `PAGES=contacto RUNTIME_SITE=relojistas.com` refused with
  exit 1 and zero rows written, surfacing TWO different actors' open items on
  that one page (`rerender-pages` via `page_id`, `image-build-handler` via
  item_key segment) — a real collision-in-waiting, found on the first live run.
- **Site fallback:** relojistas.com blocked on 13 recently-touched open items,
  5 older listed FYI-only.
- **Happy path:** code-only selftest inserted and was cleaned up
  (`DELETE ... item_key='needs_diagnosis:coverage-check-selftest'`).
- The `FORCE=1` branch was not exercised live (it would fire a real Kafka
  dispatch); it is a trivial print-and-continue.

## Decision 2 — commit hygiene: ADOPTED, carried by a new root CLAUDE.md

`CLAUDE.md` created at repo root (commit `fc773d084`) — no CLAUDE.md existed, so
every session now loads the practice at start: commit per task with an explicit
**pathspec** (`git commit <paths> -m ...`, which ignores whatever other sessions
have staged in the shared index — strictly safer than add-then-commit), never
`add -A`/`add .`/`commit -a`, forward-only, and re-run `git status` before
acting on a stale snapshot. Hook ENFORCEMENT deliberately not built — per §6.3,
habit first; revisit only if bundled commits keep appearing in the log.

## Decision 3 — deploy blast radius: STRUCTURAL FIX, announcements deferred

Ruling: the structural fix dominates the doc_notes announcement channel, for the
handoff's own reasons — an image built from a committed ref **cannot** bundle
WIP, whereas an announcement only works if every session reads it at the moment
of risk. No doc_notes deploy channel was built (and the `subject_type` CHECK
question therefore stays moot).

Built instead (makefile, commit `0c7b17616`):

- `make build-agent-chassis-ref REF=<ref>` — `git archive` into a clean /tmp
  context, builds from committed state only, prints the resolved sha for
  attributability. `REF` defaults to `HEAD`. Verified: the archive context is
  complete (go.mod, configs/agent-chassis.yaml, dockerfile, 725 .go files);
  the dockerfile needs nothing gitignored. Not exercised through a full
  `docker build` (multi-minute, and the default tag is another session's
  staged release) — first real use should watch it.
- The existing `build-agent-chassis` keeps working-tree behaviour (deliberately
  NOT flipped under other sessions' feet) but now prints a RED report of exactly
  which uncommitted changes the image will sweep in (40 at test time), and
  points at the ref target.

Combined with per-task commits, the standing mitigation "don't deploy verified
fixes, wait for someone's release" is obsolete: commit your task, then
`build-agent-chassis-ref` ships exactly the committed state.

## Decision 4 — file-claim announcements: DEFERRED (per handoff §7.4)

Weakest case, highest ceremony, most likely ignored. Revisit only if collisions
persist after 1–3 bed in. The cheap substitute already shipped: the 090
advisory prints when your seed files are dirty in the tree, and CLAUDE.md tells
sessions to re-run `git status` before trusting a snapshot.

## Landmines met while building (§8 was right)

- The tree changed under this very session: three other-session commits landed
  mid-build (including the v1.0.1124→v1.0.1125 tag bump), a file that was dirty
  at session start (`platform/kafka/consumer.go`) was committed by someone else
  while the advisory was being tested against it, and MEMORY.md was rewritten
  externally. Every "snapshot goes stale" claim in the handoff self-demonstrated.
- `git status --cached` is not a flag (`git diff --cached --name-only` is);
  pathspec commits sidestep the shared-index problem entirely.

---

# Round 2 — follow-ups closed, and the practice tested by fire

## The makefile refactor was swept mid-edit — Symptom D, live, on this thread

While generalising the makefile (below), another session's commit
**`69d6f3ecc` "updates to vet med export for vetcomparison.uk and
directory_export_action for products"** — an unrelated task — **committed this
thread's half-finished refactor**. At that moment only 2 of 14 services had the
`wip_report` call; HEAD briefly carried a partial refactor under a commit
message about vet med exports. (`git log -S "define wip_report" -- makefile`
attributes it there, not to this thread.)

**The finding this forces, which the handoff's §6.3 reasoning missed:**
commit-per-task is **not self-protecting**. Adopted unilaterally it stops *you*
sweeping *others'* WIP; it does nothing to stop *their* `git add -A` sweeping
*yours*. The asymmetry matters:

- The only self-protection available today is **commit early, commit narrow** —
  a long-lived dirty tree is shared mutable state, not a private workspace.
  This is now in `CLAUDE.md` (`42f90aee2`).
- It is also the strongest argument yet for the enforcement hook §6.3 deferred.
  **Not built, deliberately, and flagged for the owner rather than assumed:** a
  repo-wide git hook rejecting broad adds would change every concurrent
  session's git behaviour without their knowledge, and could block another
  thread's commit mid-flight. That is an owner call, not a side effect of this
  thread. Evidence for it is now on the record if the owner wants it.

Corroborating churn in one session: `IMAGE_TAG` went v1.0.1124 → v1.0.1125 →
v1.0.1126 under us, and three unrelated commits landed on the branch.

## Deploy blast radius — now covers every backend service (`031e2f074`)

The round-1 fix covered agent-chassis only. Generalised, without pasting
anything 14 times:

- **`build-%-ref` pattern rule** — ONE rule gives all 14 backend services a
  committed-state build (`make build-<service>-ref [REF=<ref>]`). Guards: the
  service must have a `build/docker/backend/<service>.dockerfile`, and `REF`
  must resolve to a real commit (`git rev-parse --verify '<ref>^{commit}'`) —
  a ref build that silently accepted a non-commit would defeat its own purpose.
  Frontends are excluded deliberately: they build from `frontends/<app>` with
  their own Dockerfile and context, so the convention does not hold.
- **`$(call wip_report,<service>)` macro** — the round-1 inline snippet, now
  factored out and called by all 14 working-tree targets, naming the correct
  per-service `-ref` alternative in its message.
- Verified: `make -n` expands both correctly (the multi-line `define` collapses
  to a single shell line via backslash continuations — the classic make trap,
  checked rather than assumed); frontends carry no `wip_report` (`grep -c` = 0);
  no target got a duplicate call.

## 084 bare trigger — pointer added, coverage check still absent by design (`63d51441e`)

084 stays the no-record escape hatch. Its header now states plainly that it has
no coverage check, that 090 does, and gives the queue query to run by hand.

**Found while doing it — NOT fixed, flagged for the fixloop thread:** there are
**two divergent tracked copies** of 084. `./084_TRIGGER_diagnose_v1.sh` (root)
is the evolved one — anchor note, 3b subject support, correct `agent-type` pod
label. `./scripts/initial_messages/310_analysis_adapter/084_TRIGGER_diagnose_v1.sh`
is stale and greps the **wrong pod label** (`agent_type`), so its follow-up
command silently returns nothing. A session running the wrong copy is exactly
this thread's class of problem, but reconciling them is the fixloop thread's
call, not a coordination-thread drive-by. The pointer went on the root copy only.

---

# Round 3 (2026-07-17) — the default is now inverted; the two halves connect

## The gap round 2 left, stated plainly

Round 2 shipped both halves of the deploy fix but they **did not connect**.
CLAUDE.md said "commit your task when ready"; the build (`docker build ... .`)
ignored git entirely and tarred the tree. So a session could commit perfectly
and *still* ship five other threads' mid-edit code, because those files were
still dirty in the shared tree. The structural fix (`-ref`) existed but was
opt-in — precisely the "a note nobody reads is worse than nothing" failure the
original handoff warned about, rebuilt in a new shape. The owner caught it.

## The fix — invert the default (`eed45ed4b`)

`make build-<service>` now builds from committed `HEAD` via `git archive`; the
old working-tree behaviour moved to an explicit `build-<service>-tree`. Now the
commit is load-bearing by default: your commit is *what ships*, and you get the
safe build by not thinking about it.

- **Structure:** two make `define`s, `ref_build` / `tree_build`, plus two
  pattern rules (`build-%-ref` kept as an alias, `build-%-tree` new). Each of
  the 14 named backend targets is now one line: `$(call ref_build,<svc>)`.
  `wip_report` removed (folded into `tree_build`).
- **Guards (both block before docker, verified at runtime not just -n):**
  the service must have a `build/docker/backend/<svc>.dockerfile`, and `REF`
  must resolve to a real commit.
- **Fail-safe direction (the crux):** forget to commit → the ref build omits
  *your* change and says so in yellow (N uncommitted changes NOT in this image)
  → wasted cycle, caught by pod-grep. The OLD default silently shipped everyone
  *else's* untested change to prod. Failing toward "missing my change" beats
  failing toward "shipped their WIP".
- The `UNSHIPPED` note is suppressed when `REF` is pinned (≠ HEAD): pinning a
  ref is an explicit statement of what ships, so dirty-tree noise is wrong there.
- Verified: `build-<svc>` (HEAD ref), `build-<svc> REF=<sha>` (pinned, no note),
  `build-<svc>-tree` (tree + red warning), `-ref` alias, all four aggregates
  (`build-backend/agents/adapters/all`), and `make help` all expand/parse.

## Deploy scripts inherit the safe default — no edit needed, but note the change

`scripts/deploy/deploy-agents.sh` (`make build-agent-chassis push-agent-chassis`)
and `scripts/deploy/deploy-production.sh` (`make build-all push-all`) now build
**committed state**. Correct for deploy scripts, and no code change required.
**Workflow consequence to know:** you must now COMMIT before running
deploy-agents.sh, or it ships HEAD without your change (loudly — the yellow note
fires). This aligns with what today's bug handoffs already tell sessions
(`bugs_open/008`, `009`: "never build from the working tree"); another session
already shipped v1.0.1127 via the ref build before this inversion landed.

## Frontends still excluded — deliberate, unchanged

`build-user-portal` / `build-agent-playground` build from `frontends/<app>` with
their own Dockerfile and context; the backend dockerfile convention doesn't hold,
so they have no `-ref`/`-tree`. Their blast radius is one subdirectory, not the
whole tree. Out of scope; flag if they ever bundle WIP that matters.

## Open follow-ups (small, none blocking)

- **Two divergent copies of 084** (Round 2 above) — fixloop thread's call.
- **Enforcement hook for commit hygiene** — evidence now strong (§Round 2), but
  it is an owner decision because it changes other live sessions' git behaviour.
- Ref builds cover backend services only; frontends would need their own rule
  if they ever start bundling WIP that matters.
- `build-agent-chassis-ref` has still never been driven through a real
  `docker build` (multi-minute; the tag belongs to another session's release).
  The context assembly is verified, the build itself is not — **first real user
  should watch it**.
- The 090 self-exclusion is exact-key only; a *differently-slugged* second
  intake at the same target is precisely what the probes are for — do not
  "improve" it to fuzzy-match.
