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

## 2026-08-16 — council ROUND 1 verdict: REVISE, and two of the three objections were right

Corr `366efae9` reached `complete_revise` / `COMPLETED` at 16:40:39Z — **the full panel ran this
time** (`select_panel` → seat reviews → `council_decide` → `route_checks` → `run_checks` →
`compose_verdict_checked` → `append_verdict` → `complete_revise`). `decided_by`: *gating objection from
`reuse_agent`*. 4 seats abstained.

**How to read the state fast, because this is what the earlier monitor got wrong:** query
`orchestration_state_audit` by `orchestration_id` (indexed). The `orchestration_states` lookup keyed on
`collected_data->'input_data'->>'fix_correlation_id'` is a jsonb scan that **times out at 100s**, which
is why my first monitor ran a full hour and emitted nothing at all. An empty result from a timing-out
query is indistinguishable from an empty result from a query that ran.

### The objections, and what each turned out to be

**1. `reuse_agent`, HIGH (gating) — "did you check `datahelpers.GetIntField` / `ExtractNestedFieldInt`
before hand-rolling a type switch?"** **It was right that I had not.** Having checked:
`GetIntField(m, key, default) int` handles float64 and int only — no int64, no `json.Number`;
`ExtractNestedFieldInt(data, path) int` handles float64/int/int64, returns 0 for missing. **Both
collapse "chose nothing" and "chose 0" into one int**, and that distinction is the entire purpose of my
`(int, bool)` pair — it is what lets anthropic/gemini apply a floor while **ollama omits the field
entirely**. A helper returning `int` cannot express two opposite policies over one answer. Neither
rejects a negative or a zero either, and `max_tokens: 0` is a hard 400. **Kept the helper; wrote the
check into the source** so the next reader does not re-ask.

**2. `reuse_agent`, MEDIUM — "this is a THIRD resolver; show the existing one could not be shared."**
**Measured, and decisive:** both existing resolvers live in `package actions`, and
`platform/orchestration/actions` **imports** `platform/aiservice` (`ai_actions.go:13`). So `aiservice`
importing either is a **genuine import cycle** — Go refuses to compile it. Not a preference I declined;
not available. `datahelpers` would not cycle, but `go list -deps ./platform/aiservice` returns **only
itself** — zero internal dependencies — so that would be the package's first dependency on the
orchestration layer. Residual stated rather than hidden: three readers of one key remain; this change
removes the *consequence* of disagreement, not the duplication (candidate 2).

**3. `bug_historian`, MEDIUM — "`GenerateWithImages` is named in the landmine beside `GenerateText` and
the plan never mentions it."** **The right question.** The answer is that both providers already route
vision through the same `generate()` this change fixes (`anthropic.go:111`, `gemini.go:305`) — so the
budget arrived by construction. **But "it shares the path" is an assertion about code someone can
refactor on a Tuesday**, and this estate's own rule is that asserted-not-verified IS the objection.
Added three vision tests (both providers inherit the configured budget through nil options, plus an
unconfigured negative control). **Mutation-proven independently:** reverting anthropic's shared budget
to a literal fails `TestAnthropicVisionInheritsTheConfiguredBudgetThroughNilOptions` on its own.

**4. `editquality`, LOW ×3 — "edits 6/7/8 are comment-only, not substantive."** Accepted as stated, and
kept. Edit 6 is not decoration: this change makes `llm_options.go`'s central sentence FALSE, and a
false explanation beside working code is how folklore starts here. Edit 8 is the same defect in a guard
test's failure message, which teaches its wrong lesson every time it fires. None of them claims to be
the fix — the fix is edits 1-4, exactly as the seat says.

**5. `editquality`, MISSING — "no edit addresses the `llm_call_log` blindness the diagnosis names."**
**Correct, and structurally out of scope at this layer:** the provider clients hold `apiKey`, `model`,
`maxTokens` and an `http.Client` and **no database handle of any kind** (grep for `sql.`/`pgx`/`DB`
across `platform/aiservice` returns one comment and nothing else). A bypassing caller has no DB either.
The clients do keep writing `__sent_max_tokens` back into the options map — the sole feed for
`llm_call_log.max_tokens` when a caller persists — so this change keeps that column honest for the
budgets it newly enables. Making direct callers visible to the fleet instrument needs a writer at the
**caller's** layer; flagged for the human decision the seat asks for, alongside candidate 2.

**Round 2 resubmitted on the SAME trail** (`RESUBMIT_CORR=366efae9…`, run orch
`8a88b982-cef0-40a9-ab62-5577244afd50`). Two of three substantive objections produced real work — a
genuine reuse check I had skipped, and three tests covering a path I had not thought about. That is the
round paying for itself.

## 2026-08-16 — council ROUND 2: APPROVED

Same trail (`366efae9`), run orch `8a88b982`. `decision: approved`, `decided_by: "approved with 3
advisory objection(s) — none high-severity"`, **`gated_by_truncation: false`** — checked deliberately,
because the architecture seat has a recorded history of truncated reviews and a truncation-gated
verdict is not the same thing as a considered one. 4 seats abstained.

Every round-1 objection was answered with a measurement or a code change rather than an argument, and
`editquality`'s own note says so: *"Round 2 answers all three substantive objections with measurement
or code rather than argument, per the stated bar."* Its assessment of the shape is worth keeping:
*"Edits 1-4 are the real fix and sit on the causal path (client-side budget resolution replacing the
optional ExecuteAIStepAction gate) … the ollama asymmetry is deliberate and justified, not an
oversight."*

### The three advisory residuals

1. **`editquality` (low, edit 7)** — the yaml edit is doc-only like 6 and 8 but was not grouped under
   my own "kept deliberately, not a fix" concession, so it reads as a functional fix. **Accepted.**
   Fair, and the cheapest kind of honesty error to make: the rationale did disclose the value is
   unchanged, but grouping is what makes that visible without reading.
2. **`bug_historian` (low, edit 5)** — the provider scan guards only providers reachable from
   `factory.go`'s `case` list; a client built outside the factory escapes it. **MEASURED, and mostly
   closed:** five production sites bypass the factory (`rag_actions.go:374`, `reasoning/agent.go:71`,
   `tools-api` defend/gripper/position) and **all five are covered anyway** — the per-provider tests
   call `NewAnthropicClient`/`NewGeminiClient`/`NewOllamaClient` **directly**, not through the factory,
   so the constructor axis was already bound behaviourally. The real residual is narrower: a future
   provider whose constructor exists but never reaches the factory switch. Recorded in the test.
3. **`bug_historian` (low, edit 2)** — struct-literal constructions outside test files were never
   enumerated the way the call sites were. **MEASURED and closed:** exactly three non-test hits, and
   **all three are the constructors themselves**. No production code builds these clients as a bare
   struct.

Residuals 2 and 3 are a good illustration of why the round is worth its cost even when it approves:
both were answerable in one grep each, neither had been run, and one of them ("did you enumerate this
axis the way you enumerated the other one?") is exactly the question I would not have thought to ask
myself, having already done the enumeration I *did* think of.

**Also flagged:** `bugs_closed/076` (*truncated_llm_responses_tolerated_at_113_unguarded_call_sites*)
was absent from `grounded_in` despite being the closest title in the family. With `012` and `046` it is
the silent-truncation lineage this bug belongs to; cited in the bug file now.

**Trailer discipline:** the code commits carry `Council-Submitted:`, which is correct for a pre-verdict
commit — `098` credits them automatically now the correlation is approved, with no amend. Only the
post-verdict commit carries `Council-Reviewed:`, written after reading the verdict body, never before.

## 2026-08-16 (evening) — LIVE at v1.0.1305, and MISSTEP 4: my census was understated 5×

Fleet rolled 22:07:55Z. **The chassis roll carried `reasoning-agent:v1.0.1305` as well**, so the
"needs its own image" caveat I recorded is satisfied — a whole-fleet `make release` covers all 14
services, which I had treated as an open question rather than checking.

**Deploy proven at the binary.** `reasoning-agent` stamps `6a782274b`; `git merge-base --is-ancestor
5cafe18ef 6a782274b` passes. The `agent-chassis` provenance line had **already scrolled out of
`--tail=400`** — exactly the documented landmine — so the `/proc/1/exe` probe settled it, with a
fabricated sha as the negative control in the same exec. The deployed `reasoning-agent.yaml` carries
the 257 annotation, again with a fabricated-string control.

**`reasoning-agent` is healthy** — 3 pods, 0 restarts, 0 error lines. Worth checking rather than
assuming: this is the first build in the service's life where that YAML `max_tokens` is actually
parsed, and it arrives as an `int`, which is precisely the type a `float64`-only reader drops.

### MISSTEP 4: "13 agents configure a budget" was really 68, and the LIVE DATA caught it

Reading `llm_call_log` after the roll, `feed-triage / score_relevance` showed **8192**. My census had
recorded feed-triage at **4000**. One query resolved it: root `ai_service.max_tokens = 4000` **and** a
step-level `workflow.steps.score_relevance.config.ai_service.max_tokens = 8192`, which
`resolveAIServiceConfig` overlays over the root (`ai_actions.go:77-87`).

**My census only ever read the root block.** Depth-aware: 193 active agents · 3 root · 10 agent-level ·
**67 step-level** · **68 distinct**. A 5× understatement, and it went into the bug file, the council
submission and three commit messages.

**What makes this worth writing down is that I had already read the warning.** `LANDMINES.md` carries
this trap twice, for this exact key — *"the cap is `config.ai_service.max_tokens`; `config.max_tokens`
reads NULL for every row"* and the depth-aware-census entry (134 LLM nodes, not 126). I grepped
LANDMINES for `max_tokens` at the start of this lane and read both. **Knowing the trap existed did not
stop me, because I was asking a different question and never noticed it was the same object.** That is
a more uncomfortable failure than not having looked.

**The conclusion is unaffected, and the reason is the uncomfortable part.** The blast-radius argument
never rested on the count — it rests on `createAIClient` (`:287`) and the options build (`:361`) reading
one merged map that already includes the step overlay. So the number was decorating an argument that did
not need it, and **nothing failed when it was wrong**. A load-bearing wrong number gets caught; a
decorative one survives until someone quotes it for a different purpose.

### The live check of the central claim, and its honest limit

0 steps changed `max_tokens` across the roll boundary; 36 post-roll calls across 5 steps as the demand
control (so the zero is not a blind pass); 0 truncations either side, counted on
`error_message ILIKE '%stop_reason=max_tokens%'` — never on `output_tokens >= max_tokens`, per the
landmine. **Early and partial: 5 steps is not the fleet.** Re-run after a day of normal load.

### What is still NOT verified, recorded so nobody over-reads the LIVE banner

The distinguishing behaviour — a nil-options caller whose config exceeds 2048 actually sending >2048 —
**has no live exerciser.** `reasoning-agent` is the only nil-options caller, its budget is deliberately
2048, nothing publishes to its topic, and its config is baked into the image (no ConfigMap volume), so
manufacturing the case needs a yaml edit plus a rebuild and would be synthetic. It is proven
deterministically by the mutation-tested wire-shape tests instead. **The bug file now says explicitly
not to write "behaviourally verified end to end" without doing that work** — which is a claim the 204
lane could legitimately make and this one cannot.

---

## 2026-09-03 — lane resumed after 18 days quiet; research only, no code

**Session:** picked up at the owner's request ("check if it already has an active thread, if so leave
it"). Ownership checked two ways before starting, because one is not enough:
`scripts/who-owns.py 257` reported this lane `[quiet 14d]` (its VERDICT line still says "OWNED or
recently active" — read the body, not the verdict, as this file already records), and a grep of the
live session transcripts modified in the last three days found no other session working it. The two
hits were a session merely *reading* the bug file and one printing a `who-owns` report. **Resumed.**

**No code was written or committed.** The fleet build `v1.0.1359` that landed during this session is
another lane's and carries nothing from here.

### The bug is still valid, and the class has come back in a new shape

Path A holds. What is new: two actions written *after* Path A landed each hand-roll the budget rule and
end in a hardcoded literal —
`rewrite_negations_action.go:537-543` (`a5d5d0728`, 2026-08-20) and
`repair_ordering_register_action.go:436-440` (`f7156fb54`, 2026-08-31).

The structural point I did not expect going in: **a literal fallback is strictly worse than passing
`nil`.** Path A works by having the client fall back to its construction config when the caller supplies
no budget. These callers always supply one, so `options["max_tokens"]` wins at the wire
(`anthropic.go:307`) and Path A never runs. The fix cannot reach the callers most likely to need it.

Copies of one precedence rule: **five** as of 2026-09-03, up from the three the bug file records.
Neither new copy reads the step's own config; neither forwards `budget_tokens`. Both live in the same Go
package as `llmOptionsFromConfig`, whose own SCOPE note says in writing that a third caller should be
the extraction rather than another paste.

### MISSTEP 5: I almost attributed four live truncations to the Go literal, and the evidence could not have told me either way

**The claim I nearly wrote.** `llm_call_log` shows four `rewrite_negations` calls truncating at
`max_tokens=2000, output_tokens=2000, stop_reason=max_tokens` (2026-08-21 → 08-23). The Go literal on
line 542 is `2000`. That reads as an open-and-shut live instance of 257: the config was dropped, the
literal applied, the page lost its repair.

**Why it is wrong.** Migration `517` (2026-08-21) gave the step its own `ai_service` block with
`max_tokens: 2000`, deliberately small and with a written rationale. Migration `569` (2026-08-23) raised
it to 16000 after the 305 lane measured a page with ten targets repairing **zero**. So the truncations
are the *configured* value being too small — a real defect, already found and already fixed the correct
way, by configuration — not the literal.

**The uncomfortable part, and the reason this is a lesson rather than a slip.** The configured value and
the Go fallback were **the same number**. The wire value, `llm_call_log.max_tokens`, and the error text
are all identical under both hypotheses. **No query over the fleet's own instrument could have separated
them.** I settled it by reading two migration files. Had 517 chosen 2048, or the Go literal chosen 1500,
one query would have answered it.

**What generalises:** when a hardcoded fallback is numerically equal to the configured value it is meant
to stand in for, the instrument that would detect the defect returns the same reading either way. This
is not "hard to measure" — it is **unmeasurable by construction**, and it looks like a clean result.
Logged in `WRONG_CALLS.md` and filed as a `LANDMINES.md` entry, because the next person to check whether
a budget "arrived" will hit it with no symptom to warn them.

`repair_ordering_register` is the live case: `offer-analyser` declares 2000 on both its steps, the Go
literal is 2000, [MEASURED 2026-09-03] 29 calls all at 2000, 0 truncations, max output 600. Not burning
— its answers are nowhere near the cap — but **whether the config is being read at all is currently
unknowable from the data.**

### Censuses re-run today, because the bug file's are stale by addition

- **Direct call sites: 12**, not the 9 in §4. That census has now been wrong three times (missed the
  Gemini client, missed a site that arrived within days, predates these two). The grep is in the
  handoff.
- **Live config, 208 active agents:** 149 steps declare `config.ai_service.max_tokens`; 10 agent-level;
  3 root; **7 declare a top-level `config.max_tokens` outside `ai_service`** (`site-adoption-agent` x4,
  `html-developer-chunked` x3) which may be a sixth spelling nobody reads; 0 declare a value `<= 0`;
  1 declares `max_tokens_ceiling` (bug 337's seam).
- **A live config trap with no traffic:** the 10 agent-level entries are all `content-creator-*` agents
  whose *step* blocks say 8000 while the agent top says 1500–2000. `ai_actions.go:357` gives the
  agent-level key priority, so the smaller number wins. **None of those ten has ever written an
  `llm_call_log` row**, so it is latent, not burning. Recorded, not acted on.

### What the next session should do

The plan is in `HANDOFF_2026-09-03_continue_here.md` §5, unsubmitted to the council. Short version:
replace both hand-rolled blocks with `llmOptionsFromConfig` (three behaviour changes, all measured inert
on today's fleet), then widen the existing call-site guard from two named files to the whole package.
⚠ The test must not land before the fix, or HEAD breaks for every other session on this tree, and the
two current violations must not be allow-listed to make it green.

---

## 2026-09-03 (later the same day) — round 2 SHIPPED: the fix, the guard, and three sites the census could not see

**Session:** picked up `docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/HANDOFF_2026-09-03_continue_here.md`
and implemented its §5 steps 1 and 2 in one commit. **Commit `51357cf51`.** Council submission
`c8660cfb-690d-4dd2-8b1f-25828305133e` (trailer `Council-Submitted:`, verdict pending at the time of
writing).

### What shipped

All five direct model callers in `platform/orchestration/actions` now resolve their budget through
`llmOptionsFromConfig`:

| file | before | after |
|---|---|---|
| `rewrite_negations_action.go` | hand-rolled block + literal `2000` | resolver |
| `repair_ordering_register_action.go` | hand-rolled block + literal `2000` | resolver |
| `companies_house_llm_review_action.go` | `map[string]interface{}{}` — an EMPTY map | resolver |
| `execute_vision_prompt_action.go` | float64-only read, **no `> 0` guard** | resolver |
| `feed_actions.go` (Perplexity) | `"max_tokens": 4096` in a raw HTTP body | `source_config`, 4096 as default |

Plus `llm_budget_call_sites_test.go` (new, package-wide AST audit) and a dated in-place correction to
`llm_options.go`'s SCOPE note, whose prediction — *"a THIRD caller should be the extraction, not another
paste"* — was ignored twice.

### ⚠ THE CENSUS IN THE HANDOFF'S §4 WAS WRONG IN THREE PLACES, AND THE TEST IS WHAT FOUND IT

This is the misstep worth recording, and it is mine as much as the handoff's — I set out to implement a
plan whose §4 table I had read and believed.

1. **`companies_house_llm_review_action.go:133` was filed as "reads config, no literal — acceptable".**
   It does not read config for the budget at all: it passed a literal **empty options map**. Not a burn
   since Path A, but the classification was wrong, and had I only implemented "step 1 as written" it
   would have stayed wrong.
2. **`execute_vision_prompt_action.go:212` was filed the same way.** It reads `max_tokens` float64-only
   **with no `> 0` guard**, so a configured `max_tokens: 0` would be sent verbatim — a hard 400 from
   Anthropic, per `platform/aiservice/max_tokens.go`. Latent, not live ([MEASURED] no active agent
   declares a non-positive budget), but it is a defect the census recorded as acceptable.
3. **`feed_actions.go:607` was not in the census at all**, in any of its four runs. `fetchViaPerplexity`
   builds a raw HTTP chat-completions body and never touches `platform/aiservice`, so **every census
   this bug has ever run was structurally blind to it** — they all grep `\.GenerateText(`. It carried
   `"max_tokens": 4096`.

**What generalises, and it is now a `LANDMINES.md` entry:** a census of "who hardcodes a budget" that is
keyed on the client interface can only find callers who use the client. A provider reached over raw HTTP
is invisible to it, and the census reads as complete. The AST audit found it in the first run because it
asks a different question — *is a numeric literal written to a budget key anywhere in this package* —
which does not care how the request is transported.

### The guard, and why it is not an allow-list

`TestEveryModelCallResolvesItsBudgetFromConfig` / `TestNoHardcodedTokenBudget` /
`TestTheCanonicalBuilderStillResolvesFromConfig`, in
`platform/orchestration/actions/llm_budget_call_sites_test.go`.

- **AST, not source text** (`go/parser` asked for no comments at all). A source scan would make this
  package's comments load-bearing — the documented failure shape where a needle matches your own prose
  and the test passes vacuously.
- **`ai_actions.go` is the ONE exemption**, because it *is* the canonical resolver and cannot call the
  helper (import cycle + different precedence level; that is candidate 2). The exemption is not blind:
  the third test asserts `ai_actions.go` still assigns a budget from something that is not a literal,
  so the moment it stops resolving from config the exemption stops being granted.
- **The two live violations were FIXED in the same commit, never allow-listed.** Allow-listing a
  detector's own motivating case is a documented failure shape here.
- **Non-vacuity counters**: the tests `t.Fatal` if they see zero model calls or zero budget-key writes.
  Today they log `audited 10 model call sites` and `audited 8 budget-key writes`.

### [MEASURED 2026-09-03] Mutation proof — four mutations, in an isolated HEAD checkout, never in the tree

| mutation | result |
|---|---|
| reinstate `options["max_tokens"] = 2000` in `rewrite_negations_action.go` | FAIL — *"rewrite_negations_action.go:553 assigns a hardcoded number to a token budget"* |
| replace the resolver call with a hand-built map in `repair_ordering_register_action.go` | FAIL — *"called from runRegisterRepair, which never calls llmOptionsFromConfig"* |
| pass `nil` options at `companies_house_llm_review_action.go` | FAIL — *"is handed a nil options map"* |
| rename the exempt file constant to `ai_actions_renamed.go` | FAIL on all four `ai_actions.go` call sites **and** the exemption's own repoint message |

A guard that has never been seen to fail is a guard nobody has tested. All four restored, all green.

### Verification route used, and one trap avoided

`go test` in the working tree **could not build**: another session's untracked
`recommended_type_reconciliation_test.go` does not compile. That is not a reason to touch their file —
`scripts/verify-head-builds.sh --with <each file> --test` builds committed HEAD with my files overlaid,
which is exactly the question. After committing, `scripts/verify-head-builds.sh ./...` → **OK, HEAD
496be18c1 builds** (HEAD moved four times during the session; other lanes are committing fast).

⚠ **HEAD does not currently pass its own tests, and it is not this change.** Two failures in
`platform/orchestration/actions` come from another lane's committed work:
`TestFindingCodeScanEveryWriteIsRegistered` (undeclared error code
`FAIL_WORK_ITEM_MESSAGE_TEMPLATE_FALLBACK`) and `TestTemplateExecutorsAreDeclared` (undeclared
`renderFailWorkItemMessage`). Control run: `git grep` puts both symbols in
`fail_work_item_message_template.go` at HEAD, and zero of the seven files I touched mention either.
Left for that lane — the fix is a registration decision, not a typo.

### One more live finding, recorded and NOT acted on

[MEASURED 2026-09-03] The handoff's open question about the 7 top-level `config.max_tokens` entries is
answered, and the answer splits:

- **`html-developer-chunked` x3 ARE read** — by `getMaxTokens(config, 16000)` in `html_actions.go:27`,
  which then synthesises a whole `ai_service` block. (That block also hardcodes model and provider,
  which is a separate smell for another day.)
- **`site-adoption-agent` x4 are DEAD CONFIG.** `analyze_site` 32000, `derive_content_direction` 6000,
  `classify_archetype` 4000, `generate_design_intent` 4000 — and **every logged call from all four runs
  at 16000**, the root `ai_service` value. `ai_actions.go:357` reads the AGENT's config then
  `ai_service`; nothing looks at the step's top level. So an operator asked for double on `analyze_site`
  and got half. Max observed output 7,708, so nothing is truncating today.

Not changed here: moving those four under `ai_service` is a live behaviour change with cost
implications the moment it is applied, and it is the owner's call, not a side effect of a code commit.

### Council verdict: ✅ APPROVED, `c8660cfb-690d-4dd2-8b1f-25828305133e`, 16 minutes after dispatch

`decided_by: "all reviewers approve"`, `gated_by_truncation: false`, 6 seats decided, 6 abstained. Seats
that spoke: `editquality`, `reuse_agent`, `guidelines`, `guardian`, `adoption_guardian`,
`llm_reliability`, `debug_historian`. No high-severity objection. Answered in commit `54ee2d261`
(trailer `Council-Reviewed:`, written only after reading the verdict):

- **`reuse_agent`, MEDIUM — "two mechanisms answering the same question, never unified once both
  exist".** The new package audit sits beside `TestNoProvocationActionCallsAModelWithAnEmptyOptionsMap`
  with no stated plan to consolidate. **A fair hit, and the objection is about the missing DECISION, not
  the code.** They differ by exactly one assertion: the old test fails if EITHER named provocation file
  stops calling `GenerateText` at all — a per-FILE alarm the package-wide version cannot give among ten
  call sites. Its empty-map half IS redundant. Both files now say this, in both directions, so the next
  reader can retire it deliberately rather than find two guards and guess.
- **`guardian`, MEDIUM — verify the Path A premise; a landmine on record contradicts it.** Answered by
  reading, not by a change: that landmine's heading is half false **and its own first bullet says so**
  (*"⚠ CORRECTED 2026-08-16 — THE HEADING ABOVE IS NOW HALF FALSE"*). The seat read the heading.
  Verified anyway — `platform/aiservice/anthropic.go:75` constructs with
  `maxTokens: resolvedMaxTokens(config)`, and `git merge-base --is-ancestor 5cafe18ef HEAD` passes.
  **And the sharper answer: this change's blast radius does not DEPEND on Path A.** Every live step
  reached by the five call sites declares a budget, so the helper supplies an explicit `max_tokens` in
  every live case; Path A governs only the hypothetical unconfigured caller.
  ⚠ **Worth carrying:** a corrected-in-place landmine still misled a reviewer, because the heading is
  what gets read. I did NOT rewrite that heading — it is the entry's `doc_notes` slug, and changing it
  orphans the synced row. So the residual is real and belongs to whoever owns the landmine sync, not to
  a session with an editor.
- **`guardian`, LOW — enumerate `fetchViaPerplexity`'s callers instead of asserting the widening is
  safe.** Done, and it should have been in the submission: exactly one, `feed_actions.go:419`, updated
  in the same commit. The seat was objecting to *"asserted, not shown"*, which is the same objection
  this lane drew in August and the correct one both times.
- **`guardian`, LOW** — a future legitimate direct model call now fails the build. Intended; the test
  now says what to do about it.
- **`editquality`, LOW x2** — the `feed_actions` sketch did not show the call-site update (true of the
  sketch; the commit does it), and edit 7 is comment-only (true, and deliberate: it corrects a note
  whose prediction had been ignored twice).

### One misstep this session, small but exactly the documented kind

I polled the council with the `orchestration_states` jsonb scan **twice**, and both attempts died at
exit 143 after the 100s timeout — while **this lane's own RUNBOOK, thirty lines above where I later
appended, says in bold that this query times out and that an empty result from a timed-out query is
indistinguishable from an empty result from a query that ran.** I read the RUNBOOK for the submission
schema and not for the polling section. The indexed `diagnosis_artifacts` query answered in under a
second. Cost: ~4 minutes and one background task. Recorded because the failure mode is *reading the
document for the thing you came for* — the trap was already written down by this same lane.
