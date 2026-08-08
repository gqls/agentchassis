# NOTES — `bugfix_136_config_key_aliases` (append-only, newest at the bottom)

## 2026-08-08 — session 1: picking the bug

Triage of `bugs_open/` against 32 live `.jsonl` transcripts. The useful discriminator was
**not** `who-owns.py`, which returns "OWNED or recently active" for almost everything on this
estate because almost every workstream directory is active. Two things worked:

1. `git log -1 --format=%cd -- bugs_open/<file>` per bug, sorted — a bug whose file has not
   moved in ≥9 days is a real signal.
2. Grepping the **tail** of each recently-touched transcript (`tail -c 400000`) for
   `bugs_open/NNN` and taking the top few by count. A session's *current* focus is legible
   that way; a whole-file grep is not, because every session that ran `ls bugs_open` matches
   everything.

Rejected after checking: `093` (its own last update says it is no longer a code task — it is
blocked on `083`, and `083` is hot in two sessions) · `211` (`who-owns` OWNED, the 122 lane
committed to it today, and one session had 62 hits on it) · `085`, `181`, `185`, `189`, `203`
(fixed and live; only site-level verification owed) · `114`, `126`, `146` (owning lanes active
within 8 days) · `040` (lane dormant, but the bug is infrastructure and the metrics predate
any claim).

**A trap worth naming**: `b5a58a2b` showed 20 hits on `136`'s vocabulary, which read as a
second session on this bug. Every hit was the auto-memory file
`bugfix-136-domain-pipeline-rename.md` being loaded as context, not work. **Grep the hits,
do not count them** — a memory file mentioning a bug looks exactly like a session working it.

## 2026-08-08 — the bug was still valid, and worse than filed

Re-measured before planning. The three named instances all still present in the live audit
(4 UNKNOWN KEYS, three of them this bug's). Two figures had moved since 2026-07-28:
`target_domain` is carried by **1** step, not 3 (migration 286 removed two under RFC 006), and
`item_domain` by 9, not 7.

Then the finding that changed the priority. §2a's `[MEASURED] Nothing is mislabelled today`
is **false as of 2026-08-04**: `content_duplication` and `page_canonical_collision` — both of
which propagate `dctx.Pipeline` — have joined `completeness-discovery-agent`, and four rows
are filed under `design` on an agent whose config says `content`. The bug file predicted this
exact trigger in the sentence immediately after the measurement. **A figure that depends on
which checks are registered where has a shelf life of days here**, and §2a did not mark it.

## 2026-08-08 — the framework finding, which is the whole shape of the fix

Reading `action_inputs.go` to see whether `ActionInputSpec.Deprecated` could carry these
renames: **it cannot, and would be worse than nothing.** Strategy 3 does
`ExtractNestedField(collectedData, config[oldKey])` — the old key's *value* is a dot-path.
That is a reference alias. A literal setting placed there resolves nothing, takes the
default, and — because `UnknownConfigKeys` recognises `Deprecated` keys on purpose — silences
the detector as well.

This answers the question the bug file asked and could not answer: *"They wrote it on one
action and not on the other two."* There was no honest declaration available, so two of three
authors correctly did nothing. The per-action shim was never the fix; it was the symptom.

## 2026-08-08 — missteps

1. **`grep 'config\["'` to enumerate an action's config reads.** Missed `priority`, which
   arrives via `GetIntField`. Caught by the planner before it reached a doc. Logged in
   `WRONG_CALLS.md`; the check is in the LANDMINE that shipped with the fix. I made this
   bug's own mistake — a key read in a way the obvious search cannot see — while writing its
   fix.
2. **A mutation that did not compile is not a mutation proof.** My first attempt at killing
   `TestResolveConfigSettingPrecedence` (`if canonical == key` → `if false`) left `old`
   unused, so the package failed to build and the test "failed" for the wrong reason. A build
   failure proves nothing about a test: it proves the compiler works. Redone as
   `canonical == key && false`, which compiles and makes the rule inert — and the test then
   failed on exactly the row that exists for the rule.
3. **`go build` at HEAD vs `go test` in the tree is not a comparison.** Chasing whether a
   `thunder` failure was mine, I built the package from `git archive HEAD` and it passed —
   which looked like proof I had broken it. `go build` does not run vet, and the failure was
   a vet diagnostic. `go vet` at HEAD reproduces it identically. Match the instrument before
   attributing the difference.

## 2026-08-08 — what shipped, and what was deliberately left

Committed `3f93456fd`, 13 files, `Council-Submitted: 433de2c0-682f-4d8d-8c48-28637309f1ba`.
Five guards, each killed by a deliberate mutation and restored:

| mutation | test that died |
|---|---|
| alias loop made inert (`&& false`) | `TestResolveConfigSettingPrecedence` — the alias-only row |
| precedence swapped (alias beats canonical) | same test — the both-set row |
| `DeprecatedConfigKeysInUse` gated on `checksConfig()` | `TestDeprecatedConfigKeysInUseIndependentOfOptIn` |
| alias loop removed from `UnknownConfigKeys` | `TestUnknownConfigKeysRecognisesDeprecatedConfigKeys` |
| triage action body reverted to the direct config read, **spec left declaring the alias** | `TestTriageDetectedItemsHonoursTargetDomain` (sqlmock, asserts the value reaches the UPDATE) |

That last one is the one worth keeping: it is the only test in the set that catches a
declaration which has stopped describing the code, which is this bug's own shape.

**Live verification that needed no roll**, run against production `agent_definitions`:
UNKNOWN KEYS **4 → 1**. The survivor, `plan_sections: domain`, is left on purpose —
`page-build-handler` is hot with several sessions today and the UNKNOWN line is the honest
record until its owner deletes the key.

Not done, all recorded on the bug file: the four mislabelled rows are not repaired (two sit
in `detected`, a queue with no consumer — `bugs_open/083`); `summary_template` is still
biting and is **not** an alias case (aliasing it to `summary` would ship a raw
`{{.input_data.topic}}` to a human reviewer); `create_work_item`'s full opt-in; every
definition edit.

## 2026-08-08 — council APPROVED, and the objections were better than the verdict

`433de2c0` — approved, 2 advisory objections, none high-severity, 12 seats, ~8 min. Full
trail in `bugs_open/136` §7. The `architecture` seat confirmed the scope call explicitly:
additive/inert/opt-in field + same-commit LANDMINE and register entry = **normal gate, not an
RFC**, per the 2026-07-29 owner ruling. The registration is what earns that, not the size of
the diff.

Three seats independently found things I had to go and fix or answer:

- **`guardian`**: the consumer search was "a grep over Go, not over every consumer that might
  exist in a definition". True — a `query_database` step filtering `pipeline='design'` would
  not appear in a Go grep. Closed both surfaces; nothing names `design` or `content`.
- **`debug_historian`**: the pod-grep recipe was positive-only and used `-l app=agent-chassis`.
  **Measured: that selector returns 2 pods and 25 RUNNING pods carry the image (34 including
  non-Running) — 8% coverage.** RUNBOOK recipe replaced: enumerate by IMAGE, and carry a
  positive control (`Using deprecated config pattern`, Strategy 3's long-live warn) plus an
  invented negative. Pre-roll baseline banked: `new=0 pos_control=1 neg_control=0`.
- **`reuse_agent`**: "confirm no other ad-hoc alias shim exists that should have converged."
  It does. `resolveAgentTypeForSpawn` (`spawn_actions.go:3154-3163`) hand-rolls
  `group_type` → `agent_type`, a literal-setting alias, this exact class. **Measured before
  deciding: `group_type` is set by ZERO live steps**, so it guards a shape nobody writes.
  Recorded as a convergence candidate with no live exposure rather than pulled in.

**`editquality` was right about the submission and wrong about the world**, and the gap is
worth naming: it objected that the parity test, the behaviour-preserving test, the LANDMINE
and the register entry are *"claimed in the rationale but absent from the edits list — if they
exist they should be edits; if not, the risk mitigations are fictional."* All four shipped in
`3f93456fd`. The cause is the schema's **8-edit cap**: I spent all eight slots on code and
described the rest in prose. **A reviewer can only see the edits list. A mitigation named only
in the rationale reads as fiction, and it is not the reviewer's job to assume otherwise.**
Next submission: spend a slot on the test file, or say in terms "shipped in the same commit,
not listed, cap reached".

## 2026-08-08 — the second misstep, and it is the same shape as the first

Logged in `WRONG_CALLS.md`. I claimed the four mislabelled rows were doing measurable damage,
citing `countDispatchableWorkItems`' `WHERE ... AND pipeline = $2`. **That query has one
caller and it always passes `"build"`** — so it cannot distinguish `design` from `content` and
cannot be evidence about confusing them. Every live pipeline-filtering consumer, Go and
definition alike, names `build`, `reports` or `diagnose`. The rows are mislabelled; the
demonstrated cost today is nil.

What caught it: going to answer the *opposite* question — whether anything would BREAK when
the fix moved those rows. The enumeration that proves a change safe is the same enumeration
that grades the harm claim, and I had done it in only one direction.

**Both of today's missteps are one shape**: I trusted the code in front of me over the
enumeration of what reaches it. A `config["` grep hid a key read through a helper; a `$2` in a
predicate hid a caller that always passes a constant. **A parameterised filter tells you what a
query CAN discriminate, never what it DOES.** Corrected in place at `bugs_open/136` §6, which
is the version to quote — §1's "the harm is not cosmetic" is superseded and says so.

Both corrections went in *after* the council had my overstated version. The `guardian` seat
found the same gap independently, from the submission alone, which is the argument for
submitting before you are certain rather than after.

## 2026-08-08 (late evening, cold-start session) — §9's "where do these executions log?" ANSWERED, and the answer kills log-sweeping as a witness

The executions log exactly where they should: **stdout of an ephemeral per-agent pod**
(`agent-<type>-<agent_id[:8]>-<hash>` — the tool-improver row of 21:18:16Z was filed from
`agent-tool-improver-3f5db2ad-pqjn9`, which was still Running when I looked). The sweep in §9
found nothing because **two independent mechanisms destroy the evidence within minutes**:

1. **Container log rotation eats the lines within seconds.** That pod's retrievable log
   *starts at 21:18:21.613Z* — five seconds AFTER the work-item row landed (21:18:16.17). The
   chassis's `DEBUGaa` state dumps are so large (223 retrievable lines ≈ tens of MB) that the
   rotation window during an active orchestration is measured in seconds. The
   `create_rerender_item` execution WAS logged and was already rotated away by the first
   opportunity to grep. A `--since=3h` sweep is theatre against a <1-minute retention.
2. **`agent-job-cleanup` deletes Completed agent pods within minutes** (a CronJob, visible
   completing every few minutes), so for short-lived carriers the pod itself is gone. Measured
   live: the domain cascade ran classifier→exemplar→strategist→briefing 22:09→22:17Z, filing
   four carrier rows, and at 22:18:13Z **zero** of those pods existed.

So §9's option 1 (find the logs) is answered and closed: the logs exist, and no after-the-fact
sweep can reach them. A 5s polling watcher over carrier pods is armed this session as a
belt-and-braces (`alias_witness_watcher.sh` in the session scratchpad), but rotation can
outrun even that during a dump-heavy step.

## 2026-08-08 (late evening) — the deterministic witness, designed and fired BEFORE reading the result

§9's option 2 (edit a live carrier to `item_domain: "content"`) touches another lane's agent
and risks a real row landing outside `pipeline='build'`, which `countDispatchableWorkItems`
and the stale-item reaper DO filter on. Instead: a **one-shot agent definition I own**
(`alias-witness-136`, category `test`), whose single `create_work_item` step carries ONLY the
deprecated `item_domain` key with the NON-default value `content`, filing an item that is
**born `status='cancelled'`** — outside `idx_swi_dedup`'s partial index, outside every
dispatcher's status predicate, and NOT `detected` (which `triage_detect_items` launders to
`pipeline='build'` — 090 trigger header, note 4). Anchored to `system.internal`, the 090
pseudo-site. No LLM steps, no credits.

**The disconfirming observation, written down before firing** (the rule this lane's third
wrong call bought): row `pipeline='content'` → alias honoured at runtime; row
`pipeline='build'` → the alias fell through to the default in production — each outcome
refutes the other, unlike every observation available from the nine live carriers, whose
configured value equals the default.

Fired 2026-08-08 ~22:40Z, `WITNESS_CORR=6e89cad7-29e1-4a9d-9d9c-c0a66b79d0cb`, publish
confirmed by the `PUBLISH_OK` marker pattern (payload in the container COMMAND — the kcat
stdin trap). Dispatch queues behind the shared generic lane: budget ~30 min. Cleanup owed
after the read: deactivate the definition; the witness row stays (it is the evidence and is
inert).

## 2026-08-08 (night) — the witness landed: pipeline='content'. WITNESSED.

Row `1d46cf49` at 22:25:39Z: `pipeline='content'`, `status='cancelled'`,
`item_key=alias_witness_136_eac60db8`. The config carried only `item_domain: "content"`;
the default is `"build"`; the row reads `content`. The alias is honoured at runtime in
production on v1.0.1268. Orchestration `8a46f229` COMPLETED on
`agent-chassis-778b7c77c7-rd27g` — the generic lane runs on the long-lived chassis
deployment pods, not an ephemeral pod, which is worth knowing for the next dispatch trace.
Dispatch→row ~3 minutes (quiet lane; budget 30 anyway).

The corroborating warn line was already unreachable ~2 minutes later on that same still-
running pod — measured the retention: oldest retrievable line was **0.4 seconds old**. That
number closes §9's log question permanently and became a LANDMINES entry.

Cleanup done: definition `alias-witness-136` deactivated + soft-deleted 22:27:46Z
(`remaining_active=0` verified). The witness row stays — born-terminal, no consumer, it is
the evidence. Bug file §11 written. The one thing the 2026-08-08 evening handoff said was
owed is delivered.

**Provenance footnote on the LANDMINES entry (2026-08-08, minutes later):** the "chassis pod
retrievable log <1s" entry reached HEAD as a same-file passenger in ANOTHER session's commit
`1eae32644` (fix(226), 22:30:58Z) — it landed between my append and my commit, so my own
commit `745ed93d8` says "New LANDMINES entry" while carrying zero LANDMINES diff. The entry
is live at HEAD and synced to doc_notes; only the attribution is odd. This is the
`a-pathspec-commit-still-takes-a-same-file-passenger` mechanism running in the direction
that costs nothing.
