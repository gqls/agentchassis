# HANDOFF — provocation pipeline, 2026-08-10 (evening)

**Supersedes `HANDOFF_2026-08-10_continue_here.md` for "what to do next".** That file's
§3.3, §3.4 and §4 are unchanged and are NOT repeated here — read it for those. Its
merge banner explains the lane takeover.

---

## 1. State in one paragraph

**The generator works.** It has produced eight candidates against a real model, the
gate approved seven and rejected one for a fabricated experimental result. Those seven
are sitting in the pool as `status='approved'`, **undated and unstamped** — they cannot
publish, and the only thing standing between them and the site is the owner reading
them. The shelf still ends **2026-08-15**; nothing has been added to the published
schedule yet. RFC_020 §5.3 is live. §5.2 is still built-and-not-live.

## 2. What to do next, in order

### 2.1 THE ONLY BLOCKING ITEM — the owner approves (or rejects) the seven

They are listed in `README_where_we_are.md` under 2026-08-10 evening with their
teasers. Approving is one statement per row:

```sql
UPDATE provocations
   SET human_approved_at = now(), human_approved_by = '<owner>'
 WHERE domain='vonc.com' AND slug IN ('<slug>', ...);
```

**Reject by retiring, never by deleting** (`status='retired'`) — a deleted row's slug
becomes reusable and the generator will happily propose it again, since
`loadRecentTitles` reads what is in the table.

Then date them, which is a dispatch and not an UPDATE:

```
agent_type provocation-scheduler-manual   (mig 321; max_assign 6)
```

⚠ **It dates forward only, starting tomorrow.** It will not fill the gap behind today
and must not be made to (PLAN §10.5).

### 2.2 Then re-run the generator until the shelf is deep enough

```
./docs/agent_docs/docs024_key_docs_latest/provocation_pipeline/builder/run_generation_round.sh
```

Four candidates per run, ~7/8 approval observed. It is idempotent on slugs
(`ON CONFLICT DO NOTHING`) and each run sees the previous run's titles, so batches do
not repeat themselves. **Read the run's own guards** — they refuse on a young chassis,
on a schedule appearing, and on a gate model that has drifted from the calibration.

### 2.3 Consider giving it a schedule — but read migration 371's header first

371 seats it operator-invoked and **asserts the absence of a `scheduled_tasks` row**.
That assertion was correct while the path was unevidenced. It now has evidence: two
runs, eight candidates, no truncation, one correct rejection. **The commit that removes
that assertion must carry the run that justified the cadence** — that is the bar 371
sets for itself, and a weekly run producing four is the obvious first proposal.

### 2.4 Owed to the council — the extraction (`65d153f0`, REVISE)

Four seats independently (`reuse_agent`, `constitution`, `prior_art_librarian`,
`architecture`) want `ai_actions.go:351-372`'s option-building **extracted** so that
path and `llmOptionsFromConfig` share one implementation, rather than the second copy
that shipped. `architecture`'s framing is the one to act on: the options-map contract
has exactly one enforced entry point and using it is **optional**, so any future action
reaching for `client.GenerateText` reproduces this defect by construction — twice so
far (`bugs_open/205`, then this). It explicitly does not block. **It asks for either a
census of every direct `GenerateText` call site, or moving budget resolution into the
provider client so the choice disappears.** The census cannot be done with the code
index (it does not index function bodies); it needs a grep.

## 3. What was verified, and how [MEASURED 2026-08-10 evening]

| claim | evidence |
|---|---|
| fix is live on v1.0.1283 | pod-grep, **both** replicas: 3 added strings = 1 each; **2 strings the change REMOVED = 0** (`"A one-sided piece is rejected"`, `"makes the case in a short paragraph"`) — a removed string is the stronger control |
| the token budget now reaches the API | two runs, eight candidates, **no truncation**; bodies 420–453 chars, complete sentences. Before the fix, every run died at `output_tokens=2048` |
| the gate discriminates on real generated text | `your-dog-does-not-love-you-it-loves-your-fridge` **rejected**: *"Swap in a stranger who reliably feeds and walks the same dog, and the tail wags just as hard within a week" — invented/unverified experimental result*. That is the 08-08 "invented, not uncited" narrowing firing on live content rather than on the fixture |
| nothing can publish without the owner | all seven are `publish_on IS NULL` **and** `human_approved_at IS NULL`; the feed and the scheduler both require the stamp |
| RFC_020 §5.3 live | `"Every report is read"` **0 before** the rerender, **1 after**, same cache-busted URL. Both §5.4 markers still 1 |
| blast radius of the new hard-refuse | census: `generate_provocations`/`gate_provocation` appear in exactly **two** active agents, both this lane's, both vonc.com |

## 4. Traps this session paid for

- **A config key that nothing reads looks exactly like one that works.** I changed live
  config **twice** against `ai_service.max_tokens` before reading the call site. Run 3's
  error still said `output_tokens=2048` — the disconfirmation was in the error text
  after the first fix, and I read it as "still too small" rather than "the number never
  moved". `[MEASURED]` the wrong thing twice.
- **`contactforsales` is a FALSE marker on the round record page** — it already appears
  once from the site footer, so grepping for the address would have reported the report
  route as shipped before anything shipped. Grep the *copy*, not the address.
- **"Mirrors ExecuteAIStepAction" was false** and I wrote it in code and in a council
  submission. `ai_actions.go`'s outer key is the **agent's** config (`:180`, `:219`),
  not the step's. Corrected in place in `llm_options.go`; `WRONG_CALLS.md` has it.
- **The chassis keeps under a second of logs.** Two investigations this session died on
  that. If you need the model's raw reply, capture it at dispatch time or not at all —
  and note these two actions bypass `ExecuteAIStepAction`, so they write **no
  `llm_call_log` row** either.

## 5. `[UNRESOLVED]` — 2048 tokens produced 84 characters

Before the fix, one run reported `output_tokens=2048 … 84 chars recovered`. 2048 tokens
of JSON should be ~8,000 characters. Something consumed the budget without emitting
text and the obvious candidate is extended thinking — but this path never passes
`budget_tokens`, which is the only way the Anthropic client enables it
(`anthropic.go:118-137`). **I never established what it was.** It has not recurred
since the budget was raised, so it may stay invisible. If a future run truncates
oddly, start here rather than assuming the budget is simply too small.

## 6. Where everything lives

- Seat + guards: mig **371** (operator handle, no schedule, gate-model assertion),
  **372** (`max_tokens` 8000 — inert until v1.0.1283, live now), **373** (`count` 4)
- Runner: `builder/run_generation_round.sh` · scheduler: mig **321**
- Code: `llm_options.go` (new), `provocation_generator_action.go` (prompt + exemplars),
  `provocation_gate_action.go` (options map)
- Commits: `36b2dc54e` fix · `8a37910e4` lane merge + NOTES · `bc09b8292` RFC_020 §5.3 ·
  `b5d9aea2f` council response + the correction
- Council: **`65d153f0` REVISE** — core fix approved outright by the gating seat; the
  extraction is what remains
- Process: `gauntlet_dead_cta/PROCESS_report_and_takedown.md` · RFC_020 §7 has build
  status for all four items
