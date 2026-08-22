# PLAN 2026-08-22 — bugs_open/358, the unread finding codes

Design, phasing, decisions **and their reasons**. Corrections to the originating bug file live
here, marked as corrections — never silently edited away.

## 0. What the bug file got right, and three things this plan corrects

`bugs_open/358` is a census, and its census holds: re-run today it reproduces (NOTES, RUNBOOK).
Three of its supporting claims do not, and each changes the design.

> **CORRECTION 1 (2026-08-22) — "the `resolved` workflow has never been used once" is no longer
> true, by one day.** 358 §1 reports 45,426 rows, `resolved = false` on every one. Live now:
> **48 resolved rows**, all stamped today 10:40 UTC by `resolved_by =
> 'content-loss-check:healed'` / `':row_gone'` (`cmd/content-loss-check/main.go:358`). The
> `bugfix_238_regeneration_key_loss` lane's checker (`cba51ad1d`) is the first user of the triple
> in the table's history. It also introduced a new code, `CONTENT_KEY_LOSS` (72 rows), written and
> consumed by the same binary. **Why it matters:** the bug file's strongest structural claim —
> that a reader has only ever shipped *with* its writer, never after — now has a fourth
> instance, and the fix has a working exemplar to copy rather than a pattern to invent.
> [MEASURED, RUNBOOK query 3.]

> **CORRECTION 2 (2026-08-22) — `RESOLVER_CONFLICTING_CANDIDATES` is NOT unruled, and it is the
> bug file's headline example.** 358 §4 B1 says: *"`RESOLVER_CONFLICTING_CANDIDATES` first: 9,615
> rows in five days is either the estate's loudest unheard alarm or its most expensive no-op, and
> nobody has ruled which."* Somebody has. The code and its sibling `RESOLVER_MAPPING_BYPASSED`
> are **Phase-1 instrumentation under `architecture_review/RFC_029`**, with a stated observation
> window, an owner, and an end condition. The concept register records the architecture seat's
> scope note verbatim (`register/contracts-and-standards.md:511`): *"this sink is Phase-1
> instrumentation for these TWO finding types — a second finding type or a second consumer gets a
> fresh architecture look."* The population has **six dated reads** (RFC_029 §10.5, §10.6, §10.7,
> §10.9, §10.11, §10.12) and an **owner ruling on 2026-08-18** (§10.13) sequencing Phase 2 on its
> evidence. The no-dedup design 358 reads as waste is deliberate: *"frequency is the population
> §9's disconfirmation clause needs."*
> **Why it matters:** the triage 358 §4 asks for has only three outcomes (consume / demote /
> keep-as-human-evidence) and **none of them fits this code.** A fourth disposition is needed —
> see §2. Filing 9,617 rows of live, owner-ruled instrumentation as "the estate's loudest unheard
> alarm" would have been the plan's first mistake.

> **CORRECTION 3 (2026-08-22) — `agenterrors` is NOT "the ONE writer", and this is the load-bearing
> correction.** `platform/orchestration/agenterrors/agenterrors.go:3` declares itself *"The ONE
> writer against agent_error_log"*, and RFC_012 (owner ruling 2026-08-06) really did retire
> nineteen hand-copied INSERTs in `platform/orchestration/actions/` into it. But an exhaustive
> grep across every language in the tree finds **five** insert paths, not one:
>
> | path | writes | goes through the seam? |
> |---|---|---|
> | `platform/orchestration/agenterrors/agenterrors.go:89` | most codes | — it IS the seam |
> | `platform/orchestration/actions/store_generated_component_action.go:1439` | `component_validation_rejected`, `component_validation_orphan_schema_field`, `component_validation_unknown_template_var` | NO — own INSERT, **deliberately** (the council's edit-quality and guardian seats objected to consolidating it; the comment at `:1428` records the ruling) |
> | `internal/agents/contentcreator/claims_guard.go:184` | `CONTENT_CREATOR_CLAIMS_DETAIL` | NO — and it **cannot**: the agent holds a `*pgxpool.Pool` (`internal/agents/contentcreator/agent.go:92`) and `agenterrors.Write` takes a `*sql.DB`. A **type-level** barrier, not laziness |
> | `docs/agent_docs/sql_for_agents/214_build_dispatch_watchdog.sql:108` | `BUILD_DISPATCH_STALLED` | NO — SQL, inside a scheduled `pre_query` |
> | `cmd/content-loss-check/main.go:292` | `CONTENT_KEY_LOSS`, `STRUCTURAL_KEY_CARRY_MISS`, `CONTENT_DATA_REGRESSION` | NO — standalone binary |
>
> Command: `grep -rn "INSERT INTO agent_error_log" --include='*.go' --include='*.sql'
> --include='*.py' --include='*.sh' .` (excluding `.git` and `_test.go`). [MEASURED.]
> **Why it matters:** the obvious fix — put the registry check inside `agenterrors.Write`, where
> every row supposedly passes — would be **blind to four of the five writers**, including one
> that is structurally unable to ever use the seam and one that lives in SQL. It would look
> complete and be 4/5ths incomplete. That is precisely the failure mode this bug is about, so
> building it would be the bug fixing itself into a new form.
>
> Independent grading requested: `090` filed this session on exactly this claim, run correlation
> `c965bfec-993a-4b2b-88ba-d44549c81df1` (intake `be708758-15b5-4a8c-8097-6a8be0d8d3b6`). The
> claim above is first-hand and re-runnable; the loop grades it because the whole design rests
> on it.

## 1. Restating the defect so the fix has something to close

The bug is not "sixteen codes lack readers". Readerless is a *symptom* and, for some codes, a
legitimate state. The defect is one rung up:

> **A finding code can enter the estate with no declared disposition, and nothing notices.**

Every property 358 measured follows from that: the count grows (nineteen codes and rising — one
more, `CONTENT_VALIDATION_WARNING_DETAIL`, was added by `0ce242d9c` **the same morning the bug
file was written**); the reader is never added later; retention erases the backlog before anyone
counts it; and the two places the estate *did* try to keep a roster have both gone stale (§3).

Framing it this way is what makes a fix possible. "Build sixteen readers" is not a fix — 358 §5
rules it out itself, correctly, because each needs domain judgement. "Never ship a code without a
declared disposition" is a fix, and it is enforceable.

## 2. The four dispositions (Correction 2 forces the fourth)

| disposition | means | what an entry must prove |
|---|---|---|
| `consumed` | an automated reader selects rows by this code and acts on them | a reader `file:line`, **verified by the checker to exist and to contain the code** |
| `instrumented` | deliberately written for a stated observation window, with an owner and an end condition | the owning doc, and a `review_by` date the checker enforces |
| `human-evidence` | legitimate hand-run forensics; the 30-day window is knowingly accepted | a reason that names the retention window it accepts |
| `operational` | failure plumbing (`UNKNOWN`, `TIMEOUT`, `LLM_API_ERROR`, …); generic newest-N diagnostic reading IS its correct consumption | nothing beyond the classification — 358 §2.3 already rules these out of scope and is right |

Plus one non-disposition: `unruled` — declared, not yet decided. It is a **visible backlog with a
count**, not an exemption, and §5 says exactly what the check does with it.

**The design rule these come from** (`optional_explicit_wire_acks.json`, RFC_029 §10.15): *"an
entry with a blank `downstream` is ignored and warned about, because an ack satisfiable by typing
the key is no ack."* Every field above is chosen so it cannot be satisfied by typing: the
`consumed` reader is verified against the file, the `instrumented` date expires, the
`human-evidence` reason must name the window.

## 3. Why a registry, and what it retires

The estate has **already tried twice** to keep a roster of live codes, both times by hand, both
times inside a test file, and **both are stale today**:

- `platform/orchestration/actions/discovery_checks_error_log_test.go:141-149` — a hard-coded
  `taken` list of nine codes, asserting the new code collides with none.
- `platform/orchestration/actions/save_sections_content_data_links_test.go:150-157` — a second
  hard-coded list of eight, plus a **prefix-disjointness** check with a real justification:
  *"the estate has two such queries today (`tool_crosslink_not_emitted%`,
  `component_validation_%`), so prefix-disjointness is a real property, not a stylistic one."*

Neither list contains `RESOLVER_CONFLICTING_CANDIDATES`, `PLAN_SECTION_NAME_DROPPED`,
`CONTENT_KEY_LOSS`, or anything else added since. **Two hand-maintained rosters that must stay
identical is the exact drift class CLAUDE.md names** (it is why `099_SYNC_gate_roster.py` exists).
A registry subsumes both: the two tests read it instead of their own lists, and the
prefix-disjointness property survives as a property OF the registry, checked once for all codes
rather than once per new code against a snapshot of the others.

## 4. The design — DB-authoritative, with a source-side early warning

Two halves, deliberately layered so the **robust** one carries the guarantee and the **fragile**
one only buys time.

### 4a. The guarantee: the live table is the authority

```
SELECT DISTINCT error_code FROM agent_error_log;
```

This sees **every** writer — Go through the seam, Go bypassing it, pgx, SQL inside a scheduled
`pre_query`, a `cmd/` binary, and any code that arrives from `agent_definitions` config rather
than source. It is immune to every blindness in Correction 3, and to the free-text and
constant-vs-literal traps in 358 §3/§8. Nothing is parsed, so no comment can become load-bearing.

The check asks two questions:

1. **Is every code observed live present in the registry, with its disposition's evidence
   satisfied?** An unregistered code is a **finding, exit 1** — this is the ratchet.
2. **Is every registered code still observed?** Report-only. A registered code absent for the
   whole window is a *possibly-dead writer*, which is bookkeeping to tidy, never a defect to page
   on (the `StaleAck` precedent, `cmd/config-key-audit/optionalbudget.go:73-77`).

**Its one blind spot, stated:** a code that has never fired in 30 days is invisible to it. That
blind spot is harmless *by construction* — an unfired code produces no unread findings and costs
nothing. The bug is about codes that fire. So the authoritative half is blind to exactly the set
of cases that do not matter, which is the property that makes it the right authority.

### 4b. The early warning: a Go test that reads the same registry

A conservative source scan (`ErrorCode: "LITERAL"`, and `const xxxErrorCode = "LITERAL"` used as
one) fails the build on a literal absent from the registry. This is the fragile half — it will
miss the ~7 sites that pass a variable, and it can be fooled. **That is acceptable, and the
layering is the point:** anything it misses is caught by 4a within a day of first firing. It buys
an author immediate feedback at commit time instead of a day later. It must never be described as
the guarantee, and the plan says so at the file head so a later reader does not mistake it for one.

> This inversion — fragile check as early warning, robust check as guarantee — is deliberate and
> is the design's answer to 358 §3.2 ("a 'zero readers' verdict produced by literal-grep alone
> answers the question it encoded, not the one asked"). We keep the grep, and we stop letting it
> answer.

### 4c. Explicitly rejected: a runtime guard in `agenterrors.Write`

The seam is the obvious place and it is the wrong one, for the reason in Correction 3: it is
blind to four of five writers, one of which cannot ever use it. A check there would report clean
while `component_validation_*` (157 rows), `BUILD_DISPATCH_STALLED` and the whole
`internal/agents/` pgx family walked past. **Rejected on measurement, not taste.**

### 4d. Normalisation, decided explicitly (358 §8's trap)

`create_tool_cross_link_items.go` emits colon-suffixed variants
(`tool_crosslink_not_emitted:tool_page_will_not_go_live`, `:no_related_pages`). The registry keys
on the code **up to the first colon**, and each entry lists its observed `raw_variants`. Stated
here and in the registry's `_doc` so it is a decision on the record, not an accident that lets a
family double-count as compliance.

## 5. Where it lives, and why nothing new is built

`cmd/config-key-audit` is the estate's audit binary. It already has **fifteen modes**, the
exit-code discipline (`0` clean / `1` findings / `2` could not determine, discriminated by empty
output because `go run` collapses the child status — LANDMINES), the acks-file idiom used twice
(`optional_key_budget_acks.json`, `optional_explicit_wire_acks.json`), and — decisively — a
**direct-Postgres route already built for exactly this reason**: `cmd/config-key-audit/fleetdb.go`
exists because *"the ai-persona-app service account has no pods/exec RBAC in this namespace"*, so
a CronJob cannot shell to `kubectl exec`. A new mode inherits all of it.

- `--finding-codes` — new mode, new file `cmd/config-key-audit/findingcodes.go`.
- Registry: `docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json`,
  beside its two sibling acks files.
- `unruled` entries are **counted and listed, exit 0**. Only an *unregistered* code exits 1.
  Reason: a check that fails from day one on a pre-existing backlog is a check that gets ignored,
  and the estate has a memory entry about exactly that. The crisp signal is "someone shipped a
  code nobody declared"; the backlog is a number that should go down.

## 6. Phasing

| phase | what | why separable |
|---|---|---|
| **1** | registry seeded from the live census + `--finding-codes` mode + its tests + the two hand-rolled roster tests repointed at the registry + concept-register entry | one coherent task; inert (a read-only audit binary); the register entry ships in the SAME commit per the 2026-07-29 ordering exemption's surviving condition (2) |
| **2** | the daily CronJob (`finding-code-registry-check`), image before overlay | the schedule is what makes it real — *"detection works; SCHEDULE and DISPATCH do not"*. Separated only because the image must exist before the overlay pins its tag (LANDMINES, `component-render-check`) |
| **3** | B1's per-code rulings, filled in against the registry, `unruled` count going to zero | needs owner judgement per code; the registry is the vehicle, and the count is the progress metric |

## 7. Acceptance, with the control for each

| claim | test | control that could have failed |
|---|---|---|
| the ratchet fires | insert one row with code `TEST_UNREGISTERED_X`, run `--finding-codes` → finding, exit 1 | delete the row, re-run → exit 0. **Mutation-proved both ways**, against the live table, not a fixture |
| the reader field is not satisfiable by typing | point a `consumed` entry at a `file:line` that does not contain the code → the checker rejects the entry | the true `file:line` passes |
| the source-side test is non-vacuous | add `ErrorCode: "TEST_UNREGISTERED_Y"` **to a copy of the tree**, run the test → fails | registry entry added → passes. ⚠ **mutate a copy, never the live file** — `WRONG_CALLS.md` 2026-08-22 records a session losing exactly this to a concurrent commit |
| the roster tests still bite after repointing | flip a registry code to collide with another → both repointed tests fail | unmodified registry → both pass |
| the census is disconfirmable | the count of live codes must be derived from `SELECT DISTINCT`, never from the registry's own length | a check that enumerated the registry and asked "are these registered?" returns the same answer whatever is true — the 2026-08-03 `[MEASURED]` trap |

## 8. Scope: council gate, not an RFC

Measured against the 2026-07-29 owner ruling (an RFC is owed when a change alters what a shared
mechanism **guarantees**, not merely because the mechanism is shared): this adds a read-only
audit mode and a declaration file. It changes nothing about what `agent_error_log` guarantees,
adds no runtime authority, and no code path's behaviour changes. Under RFC_022's narrowing it is
not architecture-scope either — there is no new opt-in field and no new authority on a shared
seam. **Normal council gate**, submitted before/alongside the commit, with the concept-register
entry in the same commit (condition (2), which survives).

One adjacent ruling to respect rather than trip: the architecture seat's 2026-08-16 scope note on
the resolver sink (*"a second finding type or a second consumer gets a fresh architecture look…
and never a second parallel sink"*). This plan adds **no** sink and **no** finding type — it
declares the ones that exist. Named here so the seat can see it was read.

## 9. Explicitly not in scope

- Building readers for the readerless codes (358 §5 — each needs domain judgement, and a blanket
  consumer would be a fourth way of feeling detected).
- The three codes `cmd/content-loss-check` owns — the `bugfix_238_regeneration_key_loss` lane's,
  in flight. They get registry entries as `consumed`, pointing at that lane's reader; nothing else.
- Changing retention. 358 §5 is right that 14/30 may well be correct; what is wrong is codes whose
  whole output falls inside the window undeclared. Do not touch the windows before phase 3.
- Consolidating the `store_generated_component_action.go` INSERT into the seam. A council round
  already ruled against it on blast-radius grounds (the comment at `:1428` records both objecting
  seats). The registry does not need it consolidated — it reads the table, not the seam.
