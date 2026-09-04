> **⚠ CLOSED 2026-09-04 — SUPERSEDED by `HANDOFF_2026-09-04b_continue_here.md`.**
> Every decision this file put to the owner was answered the same day and every one is acted on:
> (1) the induced check is ARMED on two callers, (2) candidate 2 is ruled "leave it", (3) direct-caller
> observability is `bugs_open/480`, (4) "fix it properly" shipped as the budget ladder plus migration 769,
> which is APPLIED. **Read this file for its analysis of rounds 1–2 only; its §4 asks for work that is
> done, and its §5 action list is spent.**

# HANDOFF 2026-09-04 — continue here (bugs_open/257, round 2: SHIPPED, APPROVED, LIVE — and three decisions waiting)

**Supersedes** `docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/HANDOFF_2026-09-03_continue_here.md`,
which asked for work that is now done. Read that file only for its analysis, and read its own banner
first: **its §4 census is wrong in three places.**

**Status:** round 2 is committed, council APPROVED, and **LIVE on `v1.0.1360`** (owner's roll,
2026-09-03 22:06:29Z). `bugs_open/257` **stays OPEN** — three residuals below are untouched, and two of
them are owner decisions rather than work.

---

## 1. What this bug is, in plain terms

When the platform asks a language model to write something, it must say how long the answer may be. That
number lives in configuration under `ai_service.max_tokens`, so an operator can raise it for a step that
needs longer output without touching code.

**The original defect (2026-08-12):** only ONE function read that key. Any code that talked to a
provider directly ignored it and used a hardcoded floor. **Fixed 2026-08-16 ("Path A", register
MDL-041): the provider clients now resolve the budget from the config they were constructed with.**

**Round 2's defect (found 2026-09-03):** two actions written AFTER that fix each re-implemented the rule
by hand and each ended in a hardcoded `2000`. **A literal DEFEATS Path A** — an explicitly supplied
option wins at the wire (`platform/aiservice/anthropic.go:307`), so those callers could never inherit
the configured number; passing `nil` was strictly safer than passing `2000`. And in `offer-analyser` the
literal was numerically EQUAL to the configured value, so `llm_call_log.max_tokens` read `2000` whether
the configuration was honoured or dropped. **That is fixed and live.**

---

## 2. What shipped

Five direct model callers in `platform/orchestration/actions` now resolve their budget through the
package's one resolver, `llmOptionsFromConfig` (`llm_options.go`):

| file | before | after |
|---|---|---|
| `rewrite_negations_action.go` | hand-rolled block + literal `2000` | resolver |
| `repair_ordering_register_action.go` | hand-rolled block + literal `2000` | resolver |
| `companies_house_llm_review_action.go` | `map[string]interface{}{}` — an EMPTY map | resolver |
| `execute_vision_prompt_action.go` | float64-only read, **no `> 0` guard** | resolver |
| `feed_actions.go` (`fetchViaPerplexity`) | `"max_tokens": 4096` in a raw HTTP body | `source_config`, 4096 as default |

**New guard:** `platform/orchestration/actions/llm_budget_call_sites_test.go` — three tests, AST-based
(`go/parser` asked for **no comments**, so this package's prose cannot become a needle for its own test):

1. every `GenerateText`/`GenerateWithImages` call must be handed a non-nil, non-empty options map, and
   its enclosing function must call `llmOptionsFromConfig`;
2. no numeric literal may be written to `max_tokens`/`budget_tokens` anywhere in the package, in either
   shape (`options["max_tokens"] = 2000`, or a map-literal entry);
3. `ai_actions.go` — the ONE exemption, because it *is* the canonical resolver — must still assign a
   budget from a non-literal, so the exemption cannot outlive its reason.

Non-vacuity counters `t.Fatal` on zero model calls or zero budget writes; today they log
`audited 10 model call sites` and `audited 8 budget-key writes`. **Mutation-proven, four mutations, in
an isolated HEAD checkout** — reinstating the literal, hand-rolling the map, passing `nil`, and renaming
the exempt file each fail with the intended message.

### Commits

| commit | what |
|---|---|
| `51357cf51` | the fix + the guard (`Council-Submitted:`) |
| `54ee2d261` | the council's two substantive objections answered in the code (`Council-Reviewed:`) |
| `9a30ade68` | lane docs, bug file, LANDMINES (new entry + a correction), WRONG_CALLS |
| `034870dba` | register MDL-041 dated update |
| `459d8949f` | PLAN records round 2 + the corrections to the brief |
| `3310780dc` | the APPROVED verdict, objections, and the round's RUNBOOK commands |
| `1dc25585e` | closed the 09-03 handoff |
| `083e1af0f` | LIVE on v1.0.1360, with the unprovability stated |

**Council `c8660cfb-690d-4dd2-8b1f-25828305133e` — APPROVED**, `decided_by: "all reviewers approve"`,
6 seats deciding, no high-severity objection, 16 minutes from dispatch.

---

## 3. It is LIVE — how that was proven, and the ONE thing the proof cannot do

**Live on `v1.0.1360`.** Pods `agent-chassis-ffc9ddff9-jvw92` / `-k866t`, started 2026-09-03
22:06:35/58Z.

⚠ **THE CHANGE IS BEHAVIOUR-NEUTRAL BY CONSTRUCTION, SO NO POST-ROLL READING CAN DISTINGUISH THE NEW
CODE FROM THE OLD.** Every live step reached by the five call sites declares its budget under
`ai_service`, so the helper supplies the same explicit number the hand-rolled block did. **The bug's own
§6 check cannot discriminate either** — it says to watch a step whose declared budget exceeds the old
literal send the larger number, and `page-content-writer` sent 16000 **before** the roll too, because
its hand-rolled block read the config correctly; the literal only applied when nothing was configured.
**Do not record the 16000 as evidence the fix shipped.** This is §2(c)'s blindness reappearing inside
the verification plan, one level out from where the last handoff aimed it.

**What actually proves it — ancestry off another lane's artefact proof:**

```bash
# the 453 lane proved PRC-003 (commit 681b0ee65) present in the v1.0.1360 binary with FOUR controls,
# including the strongest kind: the OLD literal "TEMPLATE RENDERED WITH MISSING DATA" ABSENT.
git merge-base --is-ancestor 51357cf51 681b0ee65 && echo "our fix is inside theirs"
```

Passes. Any build containing `681b0ee65` (committed 17:33:01+01:00) contains `51357cf51`
(17:21:53+01:00). **Not first-hand evidence** — it is their probe plus one `git` command, and it is
recorded that way. Their proof: register `PRC-003` in
`docs/agent_docs/docs026_concept_register/register/prompt-composition.md`.

⚠ `54ee2d261` is **not** an ancestor of `681b0ee65`. It is test-file comments only, so it cannot affect
the binary; whether the build also carries it is unknown and immaterial.

**Nothing regressed** [MEASURED 2026-09-04], all calls since 22:06:58Z:

| agent | step | sent | calls | max out | truncated |
|---|---|---|---|---|---|
| offer-analyser | `repair_hierarchy_register` | 2000 | 4 | 125 | 0 |
| offer-analyser | `repair_ordering_register` | 2000 | 14 | 706 | 0 |
| page-content-writer | `rewrite_negations` (loop iters 0–4) | 16000 | 184 | 5,937 | 0 |
| tool-acceptance-agent | `look` | 4000 | 8 | — | 0 |

208 calls, every one at its configured value, zero truncations — counted from
`error_message ILIKE '%stop_reason=max_tokens%'`, **never** from `output_tokens >= max_tokens`
(a truncated call has `output_tokens` NULL; fleet-wide landmine). `ch-llm-reviewer` and
`design-critique-agent.critique` had no post-roll traffic.

---

## 4. ⚠ THE THREE DECISIONS — these are the owner's, not a session's

### DECISION 1 — run the induced check that would prove the fix behaviourally? (new, and cheap)

**What it is.** There is exactly one way to tell the new code from the old by watching it run. The new
resolver reads the **step's own** config before the `ai_service` block; the old hand-rolled code never
looked at step level at all. So: put a budget where only the new code looks, and see which number goes
out.

```sql
-- on page-content-writer's rewrite_negations step, ADD a TOP-LEVEL config.max_tokens
-- (sibling of ai_service, NOT inside it). ai_service.max_tokens is 16000.
--   old code -> sends 16000 (step level invisible to it)
--   new code -> sends 15999 (step level wins)
-- then read the next llm_call_log.max_tokens for that step, and REVERT.
```

**Cost and risk.** One jsonb write. One token of difference in a budget nothing is close to (largest
observed output on that step: 5,937 of 16,000). Reversible in seconds. No image, no roll.

**What it buys.** It converts "live" from an inference (ancestry off somebody else's probe) into a
first-hand measurement — and it would be **the first live exercise of the step-level precedence arm at
all**: [MEASURED 2026-09-03] no active agent declares a step-level budget today, so that capability is
believed rather than observed.

**Why it is yours.** It writes to a live production agent's configuration and takes effect immediately.

**Recommendation: do it.** It is the smallest experiment that can fail, and this lane has now twice been
bitten by checks that could not.

### DECISION 2 — merge the last two copies of the precedence rule? (candidate 2, open since August)

**What it is.** Two places still decide "which budget wins": `ai_actions.go:357-360` (the canonical path,
serving most live steps) and `llmOptionsFromConfig`. Round 2 took the count from five back to two; it did
not answer whether the last two should become one.

**Why it has not been done.** Two real obstacles, both measured, neither fatal: `platform/orchestration/actions`
imports `platform/aiservice`, so pushing the resolver down is a genuine Go import cycle; and the two are
**not behaviourally equivalent** — `ai_actions.go`'s outer key is the AGENT's config, the helper's is the
STEP's. Merging them changes behaviour on a path serving 127 live steps across 55 agents.

**Options.** (a) Leave it: two copies, now with a test that stops a third. (b) Extract a shared resolver
into `datahelpers` (no cycle) and have both call it — a real change on a hot path, wants its own council
round and its own blast-radius measurement. (c) Decide the two rules should differ permanently and write
that down as a contract rather than an accident.

**Recommendation: (a) for now, and revisit only if a third caller needs it.** The guard removes the cost
of leaving it. What is NOT acceptable is leaving it undecided and undocumented, which is how it got to
five copies.

### DECISION 3 — is direct-caller observability part of this bug, or its own lane?

**What it is.** `llm_call_log` is written by the function these callers bypass, so a direct caller is
invisible to every truncation instrument the estate has unless it logs itself. Four of six direct callers
inside `platform/orchestration` now do (which is why round 2 could be measured at all); the two
provocation actions and everything outside the orchestration layer do not.

**Why round 2 sharpens it.** More reporting is not automatically more truth: a logged number that a
hardcoded default could equally have produced tells you nothing. That was §2(c) and it is now fixed at
these five sites, but the general shape stands.

**Recommendation: its own lane.** It is a different blast radius (every direct caller in the estate,
including `internal/agents/*` and `tools-api`) and a different kind of change (adding writes, not removing
literals). Keeping it inside 257 has kept it unstarted for three weeks.

### ...and one live-config finding that also needs a ruling (arrived with round 2)

[MEASURED 2026-09-03] **Four `site-adoption-agent` steps declare a budget in a place nothing reads** —
a TOP-LEVEL `config.max_tokens`, outside `ai_service`:

| step | asks for | actually sends |
|---|---|---|
| `analyze_site` | 32000 | 16000 |
| `derive_content_direction` | 6000 | 16000 |
| `classify_archetype` | 4000 | 16000 |
| `generate_design_intent` | 4000 | 16000 |

All four run at the agent's root `ai_service` 16000. `ai_actions.go:357` reads the AGENT's config then
`ai_service`; nothing looks at a step's top level. Max observed output 7,708, so **nothing is truncating
today** — but `analyze_site` asked for double and got half, and the other three asked for less and got
more (so this is also a spend question, not only a capability one).

**Not changed by this lane**, because moving them under `ai_service` is live the instant it applies and
changes what those steps cost. **This is the original 257 defect from the configuration end.** It wants
a ruling: honour the declared numbers, delete them as mistakes, or leave and document.

**(The other three top-level declarations, `html-developer-chunked`'s, ARE read** — by
`getMaxTokens(config, 16000)` at `html_actions.go:27`, which then synthesises a whole `ai_service` block,
hardcoding model and provider too. That last part is a separate smell nobody has looked at.)

---

## 5. What a next session should actually DO

Ordered. Nothing here is urgent; nothing is burning.

1. **If DECISION 1 is granted**, run the induced check (SQL sketch above), read the next
   `llm_call_log.max_tokens` for that step, revert the config, and record it in the bug file's live
   section. ~10 minutes including waiting for a call.
2. **If DECISION 3 says "own lane"**, open it with the standing five under
   `docs/agent_docs/docs024_key_docs_latest/<new lane>/` and take the census from `bugs_open/257` §5
   step 3 — but re-run it, do not quote it (see the traps below).
3. **Extend the guard beyond this package, or state that we will not.** `internal/agents/reasoning`,
   `internal/agents/contentcreator` and `tools-api` (defend/gripper/position, hardcoded 8192) all still
   hardcode budgets and no test watches them. A Go test is package-scoped by nature, so this needs either
   one test per package or a `scripts/pattern-check.py` check. **Note the estate's own rule: council
   scope now includes `scripts/pattern-check.py`.**
4. **Leave `bugs_open/257` open** until decisions 2 and 3 are ruled. Its bar is *fixed AND live* for the
   defect, which round 2 meets, but the file carries named residuals.

---

## 6. Traps this round hit — do not re-hit them

- **A census keyed on the client interface cannot see a provider called over raw HTTP.** Four censuses
  over three weeks all missed `feed_actions.go`'s hardcoded 4096, because they all grep
  `\.GenerateText(`. Census the CONCEPT:
  `grep -rnE '"(max_tokens|max_output_tokens|maxOutputTokens|num_predict|max_completion_tokens)"' --include=*.go . | grep -v _test.go`
  (**51 hits as of 2026-09-03**, against 12 from the interface grep). Full entry in `LANDMINES.md`.
- **A binary probe is a handful of KNOWN values, never an enumeration.** One
  `grep -aoFf <474 candidate shas> /proc/1/exe` returned **empty after ~110s** — killed, not answered —
  and an empty result there reads exactly like *"none of these commits is in this image"*. The
  must-be-absent control had passed beforehand and did not save me: **a control proves the probe
  discriminates, not that it completed.**
- **The `build provenance` log line had scrolled** out of `--tail=3000` ten hours after pod start. That
  means *not in range*, not *unstamped*.
- **Do not poll the council on `orchestration_states`** — that jsonb scan times out at 100s and presents
  as a `kubectl` hang (exit 143). Use the indexed table:
  `SELECT created_at, kind, metadata->>'decision' FROM diagnosis_artifacts WHERE correlation_id='<corr>' ORDER BY created_at;`
  **This lane's own RUNBOOK says so in bold and I hit it twice anyway** — I read the runbook for the
  submission schema and not for the polling section.
- **`097` refuses a comment-only sketch** client-side (*"a fix plan proposes changes, not
  observations"*). A documentation edit's sketch must be written as a **diff**, with `+` prefixes.
  `DRY_RUN=1` catches it for free.
- **The working tree may not compile, and it is not yours.** Use
  `scripts/verify-head-builds.sh --with <file> [--with …] --test <pkgs>`.

⚠ **`HEAD` DOES NOT CURRENTLY PASS ITS OWN TESTS IN THIS PACKAGE, AND IT IS NOT THIS LANE'S.** Confirmed
again 2026-09-04 at HEAD `3ea1552bc`: `TestFindingCodeScanEveryWriteIsRegistered` (undeclared error code
`FAIL_WORK_ITEM_MESSAGE_TEMPLATE_FALLBACK`) and `TestTemplateExecutorsAreDeclared` (undeclared
`renderFailWorkItemMessage`). Both symbols live in `fail_work_item_message_template.go`; none of the
seven files this lane touched mentions either. It belongs to that lane — the fix is a registration
decision, not a typo. **A fresh session will hit this in its first `go test` and must not spend an hour
on it.**

---

## 7. Files

- **Bug:** `bugs_open/257_HANDOFF_2026-08-12_the_token_budget_contract_has_one_entry_point_and_using_it_is_optional.md`
  — read §2026-09-03b (what shipped) and the §ROUND 2 IS LIVE section (proof + the induced check).
- **Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/`
  — `NOTES_257_token_budget_at_the_client.md` (technical log, newest at the bottom),
  `README_where_we_are.md` (the owner's plain-prose log),
  `RUNBOOK_257_token_budget_at_the_client.md` (every command this round needed, with its gotcha),
  `PLAN_2026-08-16_…` (Path A + a round-2 section),
  four `SUMMARY_*` files (the series; newest `SUMMARY_2026-09-04_…`),
  `COUNCIL_SUBMISSION_2026-09-03_257_round2_no_hardcoded_budgets.json`, and this file.
- **Code:** `platform/orchestration/actions/llm_options.go` ·
  `platform/orchestration/actions/llm_budget_call_sites_test.go` ·
  `platform/aiservice/max_tokens.go` · `platform/orchestration/actions/ai_actions.go:357-360`
  (candidate 2's other copy)
- **Register:** `MDL-041` in `docs/agent_docs/docs026_concept_register/register/model-infrastructure.md`
  (now LIVE + the round-2 update); `PRC-003` in `…/prompt-composition.md` (whose binary proof this lane
  borrowed).
- **Fleet-wide:** `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` (three 257 entries, two written
  2026-09-03 and one new: the census/raw-HTTP one) ·
  `docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md` (the census-grouped-before-it-was-read entry).
- **Related bugs:** `bugs_open/205` (the unconfigured instance) · `bugs_open/305` (migrations 517/569) ·
  `bugs_open/337` (the escalation seam) · `bugs_closed/076`, `012`, `046` (the silent-truncation lineage).
