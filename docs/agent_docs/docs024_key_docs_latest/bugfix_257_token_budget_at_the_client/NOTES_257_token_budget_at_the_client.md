# NOTES — bugs_open/257, budget resolution at the provider client

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-16 — session start: picking the bug

Asked to take `bugs_open/204`. **204 is not available: it is FIXED, LIVE AND
BEHAVIOURALLY VERIFIED END TO END** (2026-08-06, chassis v1.0.1259, canary
`fa89217a`). It sits in `bugs_open/` only because the owner directed on 2026-08-06
*"please leave the bugs that you've found in bugs_open not in the closed bug
file"*, and the file says so explicitly at its head. Nothing to do there.

Went looking for the next unowned bug instead. Method, because "check other
threads are not already looking at this" needs two different instruments:

1. `scripts/who-owns.py <n>` — reads COMMITS, so it is LAGGING and blind to a
   session mid-fix.
2. **Grep the live session transcripts**, which is the only thing that sees
   uncommitted work:
   `ls -t ~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl | head -20`,
   then for each file modified in the last ~4h, count `bugs_open/NNN` mentions.

That found ~20 live sessions. Actively worked right now: 285, 282, 225, 265,
280, 279, 283, 271, 286, 287, 083, 277, 040, 281, 253. **`bugs_open/252` shows up
in EIGHT different sessions at 6-10 mentions each — that is NOT eight sessions
working it**, it is a cross-reference in a shared cold-start doc that everyone
reads at startup. Frequency across sessions is not ownership; the shape to look
for is one session with a dominant count.

`257` appears in **zero** live transcripts, has **one** commit (its own filing,
`406595601`, 2026-08-12), and `who-owns.py` reports no owning workstream. Taken.

> ⚠ `who-owns.py` printed **"VERDICT: OWNED or recently active"** for 257 anyway.
> Its own body contradicts that: `likely OWNING workstream(s): (none identified)`.
> The verdict line appears to trigger on *any* recent commit touching the file,
> and the filing commit is a recent commit. **Read the body, not the verdict.**

## 2026-08-16 — verifying the bug is still valid

Every mechanism claim in §1 still holds. **Every line number in it does not** —
the prompt-caching work (`bugs_open/244`) moved them:

- hardcoded fallback: filed `anthropic.go:109` → actually **`anthropic.go:185`**
- options override: filed `anthropic.go:147` → actually **`anthropic.go:280-281`**

`ai_actions.go:358-364` has not moved. `reasoning/agent.go:127` still passes
`nil` and the package still contains zero occurrences of `max_tokens`.

**Two things the filing got wrong by omission, found by re-running its own census:**

1. **Gemini has the identical hardcoded 2048** (`gemini.go`, `visibleBudget := 2048`).
   The filing's §4 table is Anthropic-only, because that session read one client.
   Fixing one and leaving the other is 016b §9's recurring shape exactly.
2. **There is a tenth direct call site**, added after filing:
   `internal/tools-api/handlers/gripper.go:161`. `llm_options.go`'s SCOPE note had
   predicted *"a THIRD caller should be the extraction, not another paste"* — a
   third caller arrived inside a week. It passes a Go constant so it is unharmed,
   but the prediction held.

## 2026-08-16 — the design, and the question the filing did not ask

The filing treats candidate 1 as the big-blast-radius option wanting "a staged
rollout". Reading the constructors changed that assessment:

**Every provider constructor already receives the whole `ai_service` config map,
and already reads `model` out of it with a default fallback.** It simply throws
`max_tokens` away — the struct keeps `apiKey`, `model`, `httpClient`. So this is
*one more key from a map already in hand*, in the exact shape an existing key
already uses. Not new plumbing.

And the blast radius is not what it looks like. `createAIClient(ctx,
aiServiceConfig)` (`ai_actions.go:287`) and the options build (`:361`) read the
**same merged map**, produced once by `resolveAIServiceConfig` (`:243`). So on the
canonical path the client-side default and the options value can never disagree —
they are one map. Enumerating all seven construction sites gives: **this change
alters zero requests in the fleet today.** It converts a dead config key into a
live one.

## 2026-08-16 — MISSTEP 1: I expected the fakes to break, and their SILENCE was the defect

Adding a `maxTokens` field to the client structs means every test fake that
builds a client as a **struct literal** carries the zero value. I expected the
existing suite to fail loudly, and planned to update the three helpers.

**It passed.** `go test ./platform/aiservice/` → `ok`.

That is worse than failing. It means a client sending `"max_tokens": 0` — a 400
from Anthropic — is **invisible to the entire existing suite**, because nothing
in it asserts that field. On gemini it would be worse still: `visibleBudget = 0`
becomes "reserve tokens of thinking, zero answer".

The right response was NOT to edit the fakes to set the field (that is fixing the
checker to agree with the code). It was to **floor the value at the wire**:

```go
budget := c.maxTokens
if budget <= 0 { budget = DefaultMaxOutputTokens }
```

Three lines, and it means (a) the bad state is unrepresentable for any future
struct-literal client, and (b) **no existing test had to be edited to accommodate
this change** — a much stronger position than adjusting fakes until they agree.
`TestStructLiteralClientCannotSendAZeroBudget` demonstrates the failing case at
the moment the guard was written (register OPP-003's rule).

## 2026-08-16 — MISSTEP 2: a test that could not fail for the reason it claimed

I wrote `TestUnusableConfiguredValuesFallBackRatherThanBeingSent` to prove the
`> 0` guard in `configMaxTokens` — that `max_tokens: 0` is treated as unset
rather than sent as a 400.

Mutation **M9** removed that guard (`case float64: return int(v), true`). The
test **PASSED**. The mutation SURVIVED.

Why: the wire-level floor added in misstep 1 sits in **SERIES** with the
resolver. `configMaxTokens` returns `(0, true)` → `resolvedMaxTokens` returns 0 →
`generate` floors 0 to 2048 → the request carries 2048 → the assertion is
satisfied. The test was passing on the floor, not on the guard it named. Ollama
masks it too, via `if c.maxTokens > 0`.

This is the estate's documented shape — *a mutation that PASSES usually hit a
guard in SERIES* — and I walked straight into it while writing the tests whose
whole purpose was to be disconfirmable.

**Fix:** assert `configMaxTokens` **directly** on its `(int, bool)` pair, where
nothing is in series with it. `TestConfigMaxTokensDistinguishesAChosenBudgetFrom
AnUnusableOne` kills M9. Logged in `WRONG_CALLS.md`.

**The transferable lesson:** when a defensive floor and a validation guard both
exist on one path, an end-to-end assertion can only ever test their CONJUNCTION.
If you want to claim a specific guard is tested, the test has to reach it with
nothing downstream able to rescue the answer.

## 2026-08-16 — mutation results in full (11 run)

| # | mutation | outcome |
|---|---|---|
| M1 | anthropic request body back to literal `2048` | KILLED (5 tests) |
| M2 | gemini `visibleBudget` back to literal `2048` | KILLED (3 tests) |
| M3 | anthropic ctor ignores config (`DefaultMaxOutputTokens`) | KILLED |
| M4 | gemini ctor ignores config | KILLED |
| M5 | ollama drops the config default entirely | KILLED |
| M6 | ollama wrongly imposes the 2048 ceiling | KILLED by the opposed test |
| M7 | struct-literal floor removed | KILLED |
| M8 | resolver accepts float64 only (the `ai_actions.go:361` bug) | KILLED by the `yaml_int` subtest |
| M9 | `> 0` guard removed | **SURVIVED** → drove a real test addition |
| M10 | nil-config branch deleted | SURVIVED — **equivalent mutant** |
| M11 | `resolvedMaxTokens` ignores the bool | KILLED |

**M10 is not a test gap.** Reading a key from a nil map in Go yields the zero
value, so the switch falls through and returns `(0, false)` anyway — deleting the
branch changes no behaviour. Kept in the source with a comment saying exactly
that, so the next reader does not file it as dead code *or* as untested. Recording
an equivalent mutant AS one is the difference between a mutation score and a
mutation claim.

## 2026-08-16 — answering the filing's own [UNMEASURED] question

The filing flagged `reasoning-agent` as hard-pinned to 2048 and said plainly that
whether it truncates in production was **not established**, warning that
`llm_call_log` cannot answer it (that table is written by the bypassed function,
so it reads 0 either way — which would look like "no truncations").

Measured by the route the filing prescribed — **the pod, not the table**:

- 3 replicas, all `1/1 Running`.
- **Their entire log is startup.** Zero request-handling lines, zero
  `TruncatedError`, zero AI calls.
- `grep -rn "system.agent.reasoning.requests"` → the consumer itself, plus **one**
  other hit: `core-manager/admin/system_handlers.go:132`, a hardcoded "known
  topics" list in an admin endpoint. **No producer exists in the codebase.**
- `SELECT count(*) FROM llm_call_log WHERE agent_type ILIKE '%reason%'` → 0
  (uninformative on its own, per the filing; consistent).

**A deployed service consuming a topic nothing writes to.** The hard pin is real;
the burn is zero. Recorded so nobody re-files it as an outage — and so this fix is
not oversold as stopping a live truncation, which it does not.

**Consequence:** the key becomes live (that is the fix) and its VALUE stays 2048.
Raising a budget with no evidence of truncation is precisely what §2.2 of the bug
warns against — two live config changes were made in one day against a key nothing
read. There is now a real lever; pulling it needs a real reason.

## 2026-08-16 — what was deliberately NOT done

**Candidate 2** (replace `ai_actions.go:358-364` with a call to
`llmOptionsFromConfig`), which four council seats asked for. Reasons, so the next
session does not have to rediscover them:

- Path A removes the **harm** the duplication caused. The cost of a caller using
  neither copy was a silent 2048; it is now the configured value. What remains is
  a consistency cost, not a correctness cliff.
- **The two copies are not equivalent and cannot be naively merged.**
  `ai_actions.go`'s outer key is `agentConfig` — the *agent's* whole config;
  `llmOptionsFromConfig`'s is the *step's*. The helper also requires `> 0` (so
  `max_tokens: 0`, today a 400, would become a warning) and folds in
  `budget_tokens`, whose precedence at `ai_actions.go:377` is `aiServiceConfig`-
  only. Each is a behaviour change on a path serving 127 live steps across 55
  agents.
- Bundling a semantic refactor of the fleet's busiest LLM path into a change that
  already touches every LLM call is the trade `llm_options.go`'s own SCOPE note
  refused, for the same reason.

Also not done: raising `reasoning-agent`'s budget (no evidence — see above), and
widening the warning (the bug's candidate 5 explicitly says do not; 205 already
added one and it did not prevent this).

## 2026-08-16 — a side finding, recorded not acted on

`configs/reasoning-agent.yaml` pins `model: "claude-3-opus-20240229"` — a Claude 3
model, which `ResolveModelAlias` passes through unchanged (it is already a full
model name, not an alias). Given the service has no traffic this costs nothing
today, but a session that wires up a producer will be calling a two-generation-old
model. **Not changed here**: it is unrelated to the token budget and choosing a
replacement is a sizing/quality decision, not a bug fix. Flagged so it is written
down somewhere.

## 2026-08-16 — MISSTEP 3: the council submission died at validation and I reported it as in flight

Fired corr `85971036`, saw the trigger script's queue block and verdict instructions, and told the
owner it was submitted and awaiting a verdict. Wrote that correlation into three commit trailers,
MDL-041, the bug file, PLAN, RUNBOOK, SUMMARY and the workstreams memory line.

**It had already failed.** `current_step='complete_invalid'`, `status='COMPLETED'`, no council report.
Two of eight edits named TWO files in one `.file` field:

```
"platform/aiservice/anthropic.go + gemini.go"
"platform/orchestration/actions/llm_options.go + configs/reasoning-agent.yaml"
```

`editProblems()` in `diagnose_persist_fix_plan_action.go` refuses any `.file` containing whitespace, a
leading `/`, or `..`. No review ran; no credits were spent.

**The reason is DERIVABLE but not READABLE — I never got a server error string, because none was
recorded.** Measured: `diagnosis_artifacts` for that correlation → **0 rows of any kind**;
`collected_data` → nothing but `input_data`; `collected_data->>'__step_error'` → empty. The gate's
`persist_submission` step names no `repair_step`, so it fails without writing the refusal note the
action can otherwise produce.

What identified it was **re-implementing the server's predicate in Python and running it over my own
submission** — it flagged exactly 2 of 8 edits, plan bytes 19,093 against a 65,536 cap and 8 edits
against a cap of 8 (so neither of the other plausible causes). A reproduction, not a guess.

**Why my normal habit hid it.** The first two polls returned nothing, and CLAUDE.md says explicitly
that a missing orchestration row is almost always latency and not to retry on that evidence. That rule
is right, and it is exactly what makes this failure look like patience. `status='COMPLETED'` finishes
the job of hiding it: the word that means success everywhere else means "refused and stopped" here.

**Fixed three ways.**
1. **The submission**: restructured to 8 edits, ONE FILE EACH. To stay under the cap I merged the two
   edits that both touched `max_tokens_test.go` (candidate 4's guard lives in the same file as the
   wire-shape tests) and folded the zero-budget floor into the anthropic and gemini edits where it
   physically lives. No content dropped. Resubmitted as **`366efae9-f2a8-46ae-a6a6-b9d68a4cc6ae`**.
2. **The trigger script gains a fifth trap** (`097_TRIGGER_council_review_v1.sh`), mirroring the server
   rule exactly, and printing the offending paths. The four traps already there were each added after
   this same class of silent server-side death — operation allowlist, `grounded_in` must be strings,
   `risks` must be a string, comment-only sketch. This is the fifth.
3. **Demonstrated the detector's failing case at the moment it was written** (register OPP-003): three
   refused shapes (`" + "` whitespace, `../etc/passwd`, `/leading/slash`) all caught, plus a positive
   control that the corrected submission still passes — so the check is not simply refusing everything.

**Consequence I cannot undo:** commits `5cafe18ef`, `171121579`, `665c6773a` carry
`Council-Submitted: 85971036…`, a correlation that will never produce a verdict, so `098` will never
credit them. Forward-only forbids an amend. The honest repair is a follow-up commit naming the live
correlation and saying plainly that it supersedes the dead one.

**The rule worth carrying:** when a job's terminal state is reported by a field that reads `COMPLETED`
regardless of outcome, **the outcome is in a different field** — for the council gate, `current_step`.
One query separates "refused, resubmit now" from "genuinely queued, wait an hour". And do not write a
correlation id into commit trailers and eight documents before you have watched that run reach a
reviewing step.
