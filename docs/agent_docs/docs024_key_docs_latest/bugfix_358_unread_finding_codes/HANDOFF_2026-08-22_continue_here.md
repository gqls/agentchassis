# HANDOFF 2026-08-22 — bugs_open/358, continue here

**Read this first, then `SUMMARY_2026-08-22_unread_finding_codes.md` if you want the owner-facing
read-out.** Everything below is `[MEASURED]` on 2026-08-22 unless marked otherwise. **Re-measure
before quoting any number** — see §0, which is not boilerplate here.

---

## 0. Re-measure first, and this lane is the reason why

Every count in these docs decays on a **30-day sliding window**, and it moved *during the session
that wrote them*:

| when | distinct codes observed | registered-but-unobserved |
|---|---|---|
| ~11:00 UTC | 43 | 10 |
| ~14:00 UTC | **42** | **11** |

`REVIEW_SUPERSEDED_BY_PASSING_SAVE` — 25 rows, all 2026-07-23, the oldest in the table — was
**deleted by the retention job between those two runs**. Its entire output, erased unread, in the
three hours it took to document the fact that this happens. That is the bug, observed live, and it
is the single most useful thing to tell anyone who asks why this matters.

**`TRUNCATION_DEGRADED_REVIEW` is next: 41 rows, last written 2026-08-02, so it dies ~2026-08-25.**
If anything is ever wanted from it, extract before then. Three days from this writing.

One command re-establishes the state:

```bash
./scripts/audit-finding-codes.sh          # exit 0 = every observed code is declared
```

---

## 1. What this lane is, in one paragraph

`agent_error_log` carries deliberately-written **finding codes** — a detector's record of
something it noticed and will not fix. `bugs_open/358` measured that most have no automated
reader and are deleted at 30 days unresolved (14 if resolved — **marking a row resolved makes it
die faster**). The fix shipped here is *not* a reader per code: some codes are legitimately human
evidence, some are time-boxed instrumentation with an owner, and operational plumbing is
*correctly* consumed by generic newest-N diagnostic reads. The fix is that **a code cannot enter
the estate with no declared disposition and nobody notice.**

## 2. Status — what is DONE, and what is NOT

**DONE, live, council-approved** (`be1fd678-0836-4f32-90a6-8927b2463fee`, round 2, 13 reviewers,
4 abstained, no gating objection). This is `bugs_open/358` §4 **candidate B2**:

| thing | path |
|---|---|
| the registry | `docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json` |
| the check (mode 17) | `cmd/config-key-audit/findingcodes.go` + `findingcodes_test.go` |
| hand-run wrapper | `scripts/audit-finding-codes.sh` |
| the shared roster + source-side early warning | `platform/orchestration/actions/finding_code_roster_test.go` |
| register entry | `docs026_concept_register/register/debugging.md` → **DBG-075** |

**NOT DONE — three things, in order of who decides them:**

1. **Phase 2, the daily CronJob** (`finding-code-registry-check`) — *mine to build, not started.*
   Until it exists this protects only whoever runs it by hand. Follow
   `optional-key-budget-check`'s shape exactly (`deployments/kustomize/services/
   optional-key-budget-check/base/`): **image before overlay**, one `doc_notes` row per run
   **including clean ones** — a missing row must mean "the job did not run", never "nothing is
   wrong". ⚠ LANDMINE: `make deploy-<x>` ships nothing on its own; the overlay pins the tag. Bump
   `newTag` in the same commit as the rebuild and read the artefact, not the make target.
2. **§4 candidate B1 — the 32 `unruled` codes.** *Owner's to rule.* This is the real remaining
   work; see §5.
3. **Nothing else.** Do not "finish" B2 — it is finished for the `agent_error_log` half, and §6
   says what it deliberately does not cover.

## 3. The one thing you must not get wrong

> **`platform/orchestration/agenterrors/agenterrors.go:3` says "The ONE writer against
> `agent_error_log`". IT IS NOT — a check placed there covers one writer in five (as of
> 2026-08-22) while reading completely clean.**

RFC_012 (owner ruling 2026-08-06) really did retire nineteen hand-copied INSERTs into it.
**Five paths insert rows, as of 2026-08-22.** Per the owner ruling of that date, **re-census
rather than quoting that number** — a census does not go wrong, it goes stale by ADDITION and
reads as current for ever:

```bash
grep -rn "INSERT INTO agent_error_log" --include='*.go' --include='*.sql' --include='*.py' --include='*.sh' .
git log --since=2026-08-22 --diff-filter=A -- platform/ internal/ cmd/   # non-empty ⇒ re-census
```

| path | goes through the seam? |
|---|---|
| `agenterrors.go:89` | it IS the seam |
| `actions/store_generated_component_action.go:1439` | **no** — own INSERT, kept **deliberately**; a council round's edit-quality and guardian seats objected to consolidating it, reason at `:1428` |
| `internal/agents/contentcreator/claims_guard.go:184` | **no, and it CANNOT** — holds a `*pgxpool.Pool` (`agent.go:92`) against `Write`'s `*sql.DB`. A **type-level** barrier |
| `sql_for_agents/214_build_dispatch_watchdog.sql:108` | **no** — SQL, inside a scheduled `pre_query` |
| `cmd/content-loss-check/main.go:292` | **no** — standalone binary |

That is why the check's authority is `SELECT DISTINCT error_code` and not a source scan. Three
further traps, all of which have already produced a wrong answer in this lane:

- **Grep the CONSTANT, not just the literal.** `page_build_failure_guard.go:131` binds a const to
  `$1`; a literal-grep verdicts its code unread. This is `358` §3.2 and it caught the bug's own
  author.
- **Codes also arrive as POSITIONAL arguments** to `LogActionError(ctx, params, siteID, domain,
  action, code, …)` — `tool_birth_instance_scope_refused`, `RETRACTION_AUDIT`,
  `component_write_shared_blocked`, `PLAN_PAGE_SAME_NAME_IDENTITY_HELD`. An `ErrorCode:` grep sees
  none of them. This is a **fourth** blindness beyond the constant one 358 records.
- **`error_code` is free text**, and one writer emits colon-suffixed variants. The registry keys on
  the code **up to the first colon**; `raw_variants` lists what is actually written. Any new
  `GROUP BY` must state its normalisation or a family double-counts as compliance.

**There is no index on `error_code`** (`\d agent_error_log` — four indexes, none on it). Matters if
you build a per-code consumer at fleet scale.

## 4. Four corrections to `bugs_open/358` itself, all measured, all in the bug file

The census holds. Four supporting claims did not:

1. **"`resolved` has never been set on any of 45,426 rows"** — expired within the day. 48 rows
   now, all stamped by `cmd/content-loss-check` (`cba51ad1d`), the first use of that column in the
   table's history.
2. **`RESOLVER_CONFLICTING_CANDIDATES` is NOT unruled** — it is the bug file's headline "nobody has
   ruled which" example, and it is Phase-1 instrumentation under `architecture_review/RFC_029`
   with an owner, **six dated reads** (§10.5–§10.12) and an **owner ruling 2026-08-18** (§10.13).
   Its no-dedup design is deliberate. *This forced a fourth disposition the bug file's triage had
   no slot for.*
3. **`agenterrors` is not the one writer** — §3 above. The load-bearing one.
4. **`BUILD_DISPATCH_STALLED`'s "closed loop" is not live** — §2.2 lists it as one of three codes
   with an automated reader. Migration `214` was **never applied** (0 rows in `schema_migrations`
   for `214%`, no `build-dispatch-watchdog` task). Both halves live inside that unapplied task's
   `pre_query`. **Its zero rows read as "quiet" and mean "absent"** — which is exactly why the
   registered-but-unobserved direction is *reported* and is *report-only*.

## 5. The actual next task — B1, the 32 undecided codes

Run `./scripts/audit-finding-codes.sh` for the live list. Each needs **one ruling**, recorded by
changing its `disposition` in the registry:

| disposition | means | the entry must then carry |
|---|---|---|
| `consumed` | something automated selects by this code and acts | `reader` as `file:line` — **the check opens that file and confirms the code is in it** |
| `instrumented` | deliberate, time-boxed measurement | `owner` doc + `review_by` date **that expires** |
| `human-evidence` | hand-run forensics only | `why`, and it **must name the 30-day window** it accepts |
| `operational` | failure plumbing; generic newest-N reading IS its correct use | nothing further |

Every required field is chosen so it **cannot be satisfied by typing** — RFC_029 §10.15's lesson
from `optional_explicit_wire_acks.json`: *"an ack satisfiable by typing the key is no ack."*

**Suggested order** — not by row count, which is the wrong axis (the loudest code turned out to be
the best-governed one). Start with codes whose whole population dies soonest, because that is the
only irreversible thing here: `TRUNCATION_DEGRADED_REVIEW` (~08-25), then the rest at leisure.

**⚠ If B1 ever builds a triage flow that marks rows resolved: EXTRACT FIRST.** `resolved = true`
halves a row's remaining life from 30 days to 14. `cmd/content-loss-check` gets this right (heals,
then resolves) and is the exemplar to copy.

## 6. Scope boundary — what B2 does NOT cover, and who is waiting

`d795e10f5` routed an **RFC_008** open question at "358 B2" on the same day. It is a **different
channel** and the answer is not coming from here:

| RFC_008 expects from B2 | delivered? |
|---|---|
| no new `agent_error_log` code ships undeclared | **YES** |
| a lint tying `rendered_html` writers to the repair seam | **NO** — different seam, different table |
| measure whether `scripts/pattern-check.py` advisory findings get read | **NO** — that channel writes to a terminal and leaves **no durable row**, so the question has nothing to query. Giving it one is the prerequisite |

The RFC_008 lane has since **marked that trigger UNARMED** rather than waiting (their commits
`fb046bca2` + `ca677bbd8`), so this is settled on both sides — do not re-adopt it.

**Also proposed for `scripts/pattern-check.py` by the `bugs_open/362` lane, unowned:** a
`bugs_open/NNN` reference while `NNN` sits in `bugs_closed/` is a stale-status suspect. It is
**not** a second rule inside B2 — B2 went DB-authoritative precisely because source scans miss
positional codes, so there is no corpus-scanning helper for it to live in. Both sides record this.

## 7. A stated limit — `consumed` is NOT coverage

Contributed by the `bugs_open/309` lane, which decided **not** to route its findings here and
cited 358 in its own code for why. `STRUCTURAL_KEY_CARRY_MISS` has a reader and 8 of its 28 rows
are already resolved — **and its writer only fires when a page is BUILT**, so a component that
never gets built is invisible to it for ever (their at-rest audit found eleven such).

**A disposition says who READS a code. It says nothing about whether the WRITER can see the
population it exists to catch.** Stated in the registry `_doc` and on that entry. Do not let the
registry be read as a clean bill of health.

## 8. Traps for whoever picks this up

- **A zero over this table proves 30 days, never "never".** Any all-history claim needs a source
  that outlives the window (`page_component_history`, git, the writers' own tests).
- **`go run` collapses the child's exit status** — the vacuity refusal is discriminated by **empty
  stdout**, never by branching on exit code 2, which would be dead code. Compiled binary gives 2;
  `go run` gives 1. Same discipline as `audit-optional-key-budget.sh`.
- **Mutate a COPY, never the shipped registry** — pass `--registry <copy>`. Several sessions have
  this tree open; `WRONG_CALLS.md` 2026-08-22 records a session mutating a shared file in place to
  prove a guard and another session committing it mid-window.
- **The roster test reads a `docs/` file at `go test` time.** The council's guardian seat flagged
  this new coupling (advisory, medium). Measured as currently unexercised — nothing in the tree
  runs `go test` in a stripped context, and `.dockerignore` strips `*.md`, not this `.json`. **If it
  ever bites, CARRY THE FILE — do not hard-code the list back into the package**, which would be
  the third hand-maintained roster and the exact drift this retired. The failure message says so.
- **After committing any file another session is also in, build the COMMITTED tree:**
  `rm -rf /tmp/h && mkdir /tmp/h && git archive HEAD | tar -x -C /tmp/h && (cd /tmp/h && go build ./...)`.
  This lane broke HEAD twice in one commit by trusting a green working tree
  (`WRONG_CALLS.md` 2026-08-22, and the memory entry `a-pathspec-passenger-can-be-half-written`).
- **Use a quoted heredoc for every commit message carrying prose** — `-m "$(cat <<'EOF' … EOF)"`.
  Backticks inside `-m "…"` execute; this lane lost two words from a commit message that way, and
  forward-only means they cannot be amended back.

## 9. The five living docs

| doc | what it holds |
|---|---|
| `PLAN_2026-08-22_unread_finding_codes.md` | the design, the four dispositions, §8a's scope boundary, acceptance table with controls |
| `RUNBOOK_unread_finding_codes.md` | every query and command, gotcha attached |
| `NOTES_unread_finding_codes.md` | evidence log + **every misstep**, newest at the bottom |
| `README_where_we_are.md` | the owner's plain-prose log |
| `SUMMARY_2026-08-22_unread_finding_codes.md` | the milestone read-out, written to be read aloud |

Bug file: `bugs_open/358_HANDOFF_2026-08-22_agent_error_log_finding_codes_are_write_only_and_expire_unread.md`
(carries the four corrections). Index row: `016b_debugging_guide_8_consolidated.md` §10.

## 10. Correlations, for the trail

- **Council** `be1fd678-0836-4f32-90a6-8927b2463fee` — round 1 REVISE (gating objection was
  right: a helper called by two edits and defined by none), round 2 **APPROVED**.
- **090 diagnosis** `c965bfec-993a-4b2b-88ba-d44549c81df1` — **UNVERIFIABLE (scope-not-narrowing)**,
  and that is **not a refutation**. Its static tier is `.go`-only and index-backed: the SQL writer
  was outside its corpus entirely, and its `pgxpool` search returned 0 from an index holding no
  `type` declarations ("unrepresentable-not-absent"). **Worth knowing before filing a `090` about
  anything that is not Go.** The five-writer claim rests on first-hand reads of all five files.
