# HANDOFF — RFC_012 lane · **START HERE** · written 2026-08-08 (night)

**Supersedes `HANDOFF_2026-08-06_continue_here.md`**, which is now a long historical document.
Read this file first. Go there only for the back-story on why §1a is shaped the way it is (its
§1/§1b), the (d) CronJob runbook, and the 090/hero-logo thread.

Then: `RUNBOOK_rfc012_await_findings.md` (every command with its gotcha — the "provenance seam"
section is new and is what you will actually use) and `NOTES_rfc012_await_findings.md` (the
missteps; they are the point, and there are thirteen).

---

## THE LANE IS CLEAR. Everything owed is delivered, approved and LIVE.

| piece | id | state |
|---|---|---|
| Three owner rulings (RFC_012 second sitting) | `3851e90b5` | **ALL DELIVERED** |
| B core + 18 conversions (`agenterrors`) | `5f49b4cfd`, `f930de86b` | **LIVE** |
| (d) detector, offline + the CronJob | `abf5e8266`, `867037f5a`, `22ed9aa04` | **LIVE AND PROVEN** |
| (a)/(a′) reader census | `40992cbce` | **DELIVERED** |
| **Provenance hardening** (the merge split) | `f993554f6` | **LIVE on v1.0.1268** |
| **Provenance hardening** (step_name symmetry) | `0dc2d71a2` | **LIVE on v1.0.1268** |
| Council rounds 3a, 3b, 4, 5 | 4 correlations | **ALL APPROVED** |
| Landmine verifier | `1ffae4bf` | NEEDS_HUMAN_REVIEW → **dispositioned false alarm** (index lag) |

**Nothing is in flight. No verdict is outstanding. Nothing is broken.**

### How the hardening was proven live (do this after any roll; the recipe is in the RUNBOOK)

Two independent layers, and you want both:

1. **In the binary**, both chassis replicas, picked by image TAG **and**
   `--field-selector=status.phase=Running` (the fleet spawns ephemeral per-job pods on the same
   image; the first two I picked were `Succeeded`/`NotFound` by the time I exec'd):
   control `Failed to write to agent_error_log`=1, `recorded as unattributed rather than credited
   to the running step`=1, `names an agent_type but no step_name`=1 ← **this one separates the
   second commit from the first**, `provenance_running_agent_type`=1, a phrase in no version=0.
2. **In production behaviour**, which is the stronger claim: **305 rows from 26 distinct
   `agent_type`s** written since the roll, and `agent_type='unattributed'` = **0**. The row count
   is the positive control that makes the zero mean something — the detector alone reads 0 against
   a dead path, an empty table or a wrong predicate.

⚠ **Never verify with an anchored needle.** `strings /app/agent-chassis | grep -c "^unattributed$"`
returns **0 on a binary that carries it** — the Go linker packs constants into contiguous blobs, so
`strings` emits them concatenated dozens-to-a-line. It fails in the direction that reads as "your
fix did not ship". Landmine filed; needle on a distinctive full PHRASE.

---

## THE ONE REMAINING JOB — §1a: hoist the `RunAgentType` ladder

**This is the whole of the lane's outstanding work, and it is architecture-scope.**

### What is wrong

When a call site asks the shared error-log door for "the running step's provenance", what it
actually gets is usually the filler `generic`. Measured 2026-08-08:

- all **25** live `REVIEW_SUPERSEDED_BY_PASSING_SAVE` rows carry `agent_type='generic'`, **0**
  carry anything else;
- fleet-wide `generic` = **559 rows across 25 distinct `step_name`s** — the widest step spread of
  any `agent_type`, against `vet-practice-verifier`'s 9,696 rows over 5. That scatter is the
  fingerprint of a placeholder being passed around, not an agent;
- the table's main investigation index is `(agent_type, occurred_at DESC)`, so this degrades the
  one column every investigation starts from.

### The estate already has the answer and only one consumer uses it

`types/context.go:62` documents `ExecutionContext.RunAgentType` as *"the RESOLVED real agent type
whose workflow is executing … as opposed to the dispatch-path sender which is often 'generic'"*,
citing `bugs_open/060`. It is set by `platform/messaging/processor.go:1828`.
`coordinator.determineOwnerAgentType` (`coordinator.go:3465`) is the **only** reader, with this
ladder: `RunAgentType` → `Sender.AgentType` → `os.Getenv("AGENT_TYPE")` → log Error, return
`"generic"`. Corroboration that the estate treats `generic` as a non-identity:
`agent_error_log_test.go:187` puts it in the same set as `""` for values that must not pass a
review gate.

`runningStepProvenance` in `actions/log_action_error.go` implements only the middle rung.

### The fix, and the reachability is MEASURED not assumed

Hoist the ladder onto `*types.ExecutionContext` — e.g. `func (ec *ExecutionContext)
ResolvedAgentType() string` — then have `determineOwnerAgentType` delegate to it **and**
`runningStepProvenance` call it. One ladder, two consumers, no drift.

Both facts you need were checked, because a council seat rightly objected that I had asserted the
first without asking the compiler:

```bash
# actions CANNOT import orchestration — the cycle is real, so delegation is not an option
go list -f '{{range .Imports}}{{println .}}{{end}}' ./platform/orchestration \
  | grep -x github.com/gqls/agentchassis/platform/orchestration/actions        # -> found
# types IS imported by BOTH, so it is a valid hoist target
go list -f '{{range .Imports}}{{println .}}{{end}}' ./platform/orchestration/actions \
  | grep -cx github.com/gqls/agentchassis/platform/orchestration/types         # -> 1
go list -f '{{range .Imports}}{{println .}}{{end}}' ./platform/orchestration \
  | grep -cx github.com/gqls/agentchassis/platform/orchestration/types         # -> 1
```

### The traps, in the order you will hit them

1. ⚠ **`os.Getenv("AGENT_TYPE")` is a rung of the coordinator's ladder and does NOT belong on a
   `types` method.** Decide where that rung lives before you write a line, or you will have moved
   a drift problem rather than fixed one. Options: keep it coordinator-side and have the shared
   method stop one rung short; or pass a fallback in. Neither is obviously right — this is the
   design decision of the job.
2. ⚠ **This changes what `agent_error_log.agent_type` MEANS for every existing reader**, which is
   the OWNER RULING 2026-07-29 §1 trigger ("does it change what the shared mechanism
   GUARANTEES?"). **Read that ruling and consider an RFC rather than the council gate.** Round 4's
   `architecture` seat ruled the *previous* change `point_fix` precisely because it added no
   column and changed no wire shape; this one alters the semantics of a populated column, so do
   not assume the same answer transfers.
3. ⚠ **Name the other consumers and TELL them** (OWNER RULING 2026-07-29 §3) — measuring that
   nothing breaks is not the same as establishing that `owner_agent_type`'s owners would agree.
4. ⚠ **Test by MUTATION.** Every test in the actions package pins codes and messages; the
   provenance pins I added (`log_action_error_test.go`) are the only ones on `agent_type`/
   `step_name`, so extend those. Six mutations and their expected failure sets are in the RUNBOOK.
5. **Do NOT expect the 559 `generic` rows to change retroactively** — this is write-path only.

### Scope discipline, learned the hard way on this lane

`bugs_open/060` is a *different* bug (no durable agent-run record; usage_count dead). Do not fold
it in. This lane's round 2 was REJECTED for bundling an unrelated detector into a bug patch, and
that is why §1a was carved out of the hardening rather than shipped with it.

---

## Other threads, both explicitly NOT this lane's

- **The hero/logo silent breaks.** 090 came back **UNVERIFIABLE** (`dce40cf4`) — which means the
  evidence would not grow, **not** that the premise is false. It refuted a rival hypothesis and
  confirmed the raw two-level read is the primary path, with a `hero_url`/`logo_url` fallback
  running after it. `hero_deployed` appears in **0** of 1,667 retained `orchestration_states` rows,
  so it cannot be observed after the fact. **The cheapest decisive test is a CANARY:** run a page
  build on `pageflow-builder` or `site-work-orchestrator` and read
  `collected_data->'hero_deployed'` while it is in flight, then check whether the page carries a
  hero and a logo. One run settles what five diagnosis iterations could not. Detail in the 08-06
  handoff §2.
- **`search_results.results.0.url` can never resolve** — `vet-practice-verifier`'s
  `fallback_url_field` uses an array index and `ExtractNestedField` does map access only. That
  fallback has never fired. Belongs to the vet lane.
- **Dead config keys survive indefinitely** — `commit_from` configured in 6 agents, read by
  nothing; 4 HITL `output_format` templates never rendered. No drift check notices.

---

## Two working rules this session paid for, worth carrying into the new chat

- **`LANDMINES.md` was swept into another session's commit TWICE in ninety minutes.** It is
  fleet-wide, append-only and edited by ~30 sessions, so it is the highest-collision file in the
  repo. **Append to it and commit that file ALONE, immediately** — do not batch it with the code
  commit it documents. And after any commit: `git show <sha> --numstat -- <path>` then
  `git show <sha> -- <path> | grep -c "<text YOU added>"`. A non-zero numstat is not evidence;
  mine was `11 0`, entirely another session's lines. Full write-up in `WRONG_CALLS.md`.
- **My tool channel rewrites plain ASCII.** A typed `''` landed on disk as `U+201D`; `gofmt`,
  the compiler and 12 tests were all silent. Harmless in a comment, a silent behaviour change in a
  string literal. Finish every file with
  `grep -o -P '[^\x00-\x7F]' <file> | sort | uniq -c` and account for each character
  (`—`/`§`/`⚠`/`→` are intentional here), plus
  `grep -n -P '[\x{2018}\x{2019}\x{201C}\x{201D}]' <file>` — smart quotes never are.
- **A fresh landmine's verifier verdict carries no information.** It reads a code index pinned to
  an older commit, so an entry naming a NEW symbol is *guaranteed* `NEEDS_HUMAN_REVIEW` the day it
  is filed. Settle it with `git grep` at HEAD plus a pod-grep and write the disposition into the
  entry, or the next reader inherits an unresolved scare.

## Milestone read-outs
`SUMMARY_2026-08-06_two_of_three_built.md` · `SUMMARY_2026-08-08_shipped_proven_and_the_survey_done.md`
· `SUMMARY_2026-08-08b_all_three_delivered.md` ·
**`SUMMARY_2026-08-08c_the_seam_is_strict_now.md`** — the current one. Each is a NEW file; never
edit an earlier one.
