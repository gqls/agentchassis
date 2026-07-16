# RUNNING NOTES — multi-session coordination: decisions taken and machinery built

**Session:** "session coordination", 2026-07-16, continuing
`HANDOFF_2026-07-16_multi_session_coordination.md`. All four §7 decisions resolved.
Everything below is committed per task with explicit pathspecs (practising §5).

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

## Open follow-ups (small, none blocking)

- `084_TRIGGER_diagnose_v1.sh` (the bare ad-hoc trigger) still has NO coverage
  check — deliberate for now (084 is the no-record escape hatch), but worth a
  one-line pointer at 090 next time it is edited.
- The WIP report + ref build cover agent-chassis only; the other ~14 service
  targets still build silently from the tree. Extend if/when those bite.
- The 090 self-exclusion is exact-key only; a *differently-slugged* second
  intake at the same target is precisely what the probes are for — do not
  "improve" it to fuzzy-match.
