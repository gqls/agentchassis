# NOTES — Gemini content provider

*Technical running log. Append-only, newest at the bottom. Record missteps and
wrong turns, not just conclusions. Mark unverified claims
`[INFERRED]`/`[UNMEASURED]`/`[ASSUMED]`.*

---

## 2026-07-27 — workstream opened; the first attempt reconstructed from commits

Owner asked why the Gemini switch was reversed and to revisit it. No workstream
dir existed (the first attempt left its record in commit messages and in another
workstream's NOTES), so this one was opened.

**Grepped before filing** (CLAUDE.md): `bugs_open/`, `bugs_closed/` and
`docs024_key_docs_latest/` for `gemini` — no open bug and no existing workstream
covers the text-generation provider. The Gemini hits in `imagery/` and
`016b §9` are the **image** lane (Banana / `gemini-3-pro-image-preview`), a
different subsystem; the hits in `per_site_ai/` are Gemini used as a strategy
advisor in chat, not as a platform provider. No duplication.

**The six commits that are the whole record** — `014e45ffa` (provider added),
`7b27edfa9` (content-creator flipped), `c8896a37d` (writer flipped),
`5db6a929f` (writer reverted), `4dd5d6378` (content-creator reverted, with the
findings), `3ea9d718c` (max_tokens tiers raised). Timeline in PLAN.

**Misstep, mine, caught immediately.** I started by believing the brochure
workstream's account that the fleet switch-back was sweep `fb6d6ad44`. Checking
`git show --stat` first: that commit has no configmap change at all — 17
`kustomization.yaml` image-tag bumps, the makefile, two docs. Then
`git log -p` on just the configmap, filtered to `provider:`/`model:` lines,
gave the actual answer in one command:

```bash
git log --format="%h %ad %s" --date=format:"%m-%d %H:%M" -p -- \
  deployments/kustomize/services/content-creator-agent/overlays/production/uk_001/configs/configmap-content-creator.yaml \
  | grep -E "^[0-9a-f]{9} |^[+-] *(provider|model):"
```

Exactly two commits ever changed the provider: `7b27edfa9` in, `4dd5d6378` out.
The writer was reverted at 16:59 citing a fleet revert that, in git, landed at
17:11. Cheap check that settled it: **`git log -p` on the one file, not the
sweep's message.** A sweep commit's subject line describes the *sweep's* intent,
not its contents. Corrected in place in the brochure NOTES and recorded in PLAN.

**The finding that reopens this.** `4dd5d6378` attributed the empty/truncated
output to the model: pro thinks, thinking can't be disabled, so pro is unusable
at our budgets. Reading `platform/aiservice/gemini.go` as it stood shows the
cause is ours — old line 86:

```go
generationConfig["maxOutputTokens"] = maxTokens   // maxTokens = caller's max_tokens
```

and the word "thinking" appeared **nowhere** in the file. Gemini's
`maxOutputTokens` is a total output ceiling that thinking is drawn from first;
every `max_tokens` in this platform was sized against Anthropic with thinking
off, where the whole cap is visible text. So the 100-token twitter tier asked a
thinking model to fit reasoning *and* tweet into 100 tokens. Zero visible text is
the arithmetic working correctly.

Two things kept it invisible: `usageMetadata` decoding read only
`candidatesTokenCount`, never `thoughtsTokenCount` (so the tokens doing the
damage were invisible above the transport), and the truncation error said only
`finishReason=MAX_TOKENS`, which looks identical to a prompt that wanted to write
more.

**Status of the 07-24 evidence, restated honestly:** the content-creator tests
measured a starved budget, not writing quality. The writer test **never ran** on
Gemini at all — the queued `about` rebuild was still behind the backlog when the
revert landed (`5db6a929f`). So for the agent that writes our site copy there is
no Gemini evidence in either direction, and the flip is an open experiment rather
than a settled question.

## 2026-07-27 — client fixed (P1); what is asserted and what is not

`platform/aiservice/gemini.go` rewritten around the budget:

- `maxOutputTokens = caller's max_tokens + thinking_reserve` (default 8192) for
  any model assumed to think. Deny-list polarity: only `flash-lite` and
  `embedding` are treated as non-thinking, both measured on 07-24. An unknown
  Gemini name is assumed to think, because an unfamiliar name is almost always a
  newer model — the same polarity lesson as idea.uk's wire-format allow-list.
- No `thinkingConfig` sent unless configured. The 2.5 knob (`thinkingBudget`,
  integer) and the 3.x knob (`thinkingLevel`, string) are incompatible, and 07-24
  already caught the 400 from sending the wrong one. The reserve makes the
  default case work with no knob at all.
  > **CORRECTED later the same day — see the P4 probe entry below.** Both knobs are
  > accepted; only the *value* `thinkingBudget: 0` is refused. The "incompatible
  > generations" story was my generalisation from one rejected value. The decision
  > stands on better ground: neither knob CAPS thinking.
- `thoughtsTokenCount` / `totalTokenCount` decoded; thinking written back as
  `__usage_thinking_tokens`. `__usage_output_tokens` stays **visible** tokens so
  the field keeps the same meaning across providers.
- Parts flagged `thought: true` are skipped. Gemini returns reasoning and answer
  in the same `parts` array and the old loop concatenated both.
- Known-closed pins refused at construction with the replacement named;
  `ai_service.model` now has **no default** (the old one was `gemini-2.5-pro`,
  i.e. a default that had rotted into a 404 on every call).

`scripts/gemini-probe.sh` written so the 07-24 answers stop living in a commit
message: model reachability for this key, visible-vs-thinking per real tier,
and which thinking knob the model accepts.

**Verified:** `gofmt` clean; `go build ./platform/... ./internal/...` clean;
`go test ./platform/aiservice/` green including 11 new Gemini tests, which pin
the reserve at the exact tier (100) that produced zero text in production and
assert the truncation message names thinking as the consumer.

**NOT verified, and these are the ones that matter:**
- `[UNMEASURED]` that any of this makes Gemini produce usable text. The tests
  prove the client sends the right numbers; only the live probe and a real
  generation prove the numbers were the problem. **The reserve theory could be
  right about the mechanism and still leave pro unusable** — e.g. if it thinks
  for tens of thousands of tokens on our prompts. The probe's tier table
  measures that directly.
- `[UNVERIFIED]` that `thinkingLevel` is the accepted knob on
  `gemini-3.1-pro-preview`. Inferred from the 07-24 400 on `thinkingBudget: 0`
  plus the API's generational split. The client sends neither by default
  precisely so this inference is not load-bearing; the probe settles it.
- `[UNVERIFIED]` that a `thought: true` part ever appears from this API. The
  filter is asserted from the documented response shape and is harmless if such
  parts never arrive.
- `[UNVERIFIED]` the Gemini rates in content-creator's `estimateCost` table.
  The old keys (`gemini-2.5-pro`/`-flash`) were unreachable model names, so they
  could never match and every Gemini call silently costed at the Claude fallback
  rate. Re-keyed to the floating pointers with the 2.5-era rates carried over
  and marked `[UNVERIFIED]` inline — not checked against Google's price list.
- `[UNVERIFIED]` that `text-embedding-004` is still reachable. Same retirement
  class as the 2.5 pins. Made configurable rather than changed, since nothing
  here is known to call `GenerateEmbedding` on Gemini.

**Blocked:** `kubectl` is `Unauthorized` in this session, so P3–P6 (image roll,
pod verification, probe, both flips) cannot start. `GEMINI_API_KEY` exists only
in the cluster secret — there is no local copy, so the probe cannot be run from
here either. Everything up to the probe is done and waiting on credentials.

## 2026-07-27 (later) — cluster auth restored; P4 probe RUN, and it corrected me twice

Owner restored `kubectl` and asked for the probe + the council gate.

**Model reachability, re-verified today** (not carried over from 07-24). Key read
from pod `content-creator-agent-84564dfb67-vjq5g`:

| model | result |
|---|---|
| `gemini-2.5-pro` | **404** "no longer available to new users" |
| `gemini-2.5-flash` | **404** "no longer available to new users" |
| `gemini-3-pro-preview` | **404** "no longer available" (retired outright, different message) |
| `gemini-3.1-pro-preview` | OK |
| `gemini-pro-latest` | OK (→ 3.1-pro-preview) |

**The listing advertises models the key cannot call.** `models?pageSize=200`
returns 42 `generateContent` models including `gemini-2.5-pro` and
`gemini-3-pro-preview`, both of which 404. So the probe's own warning was right
and worth keeping: **a model appearing in the listing is not evidence the key can
reach it.** The `geminiRetiredPins` construction guard is therefore correct, and
correct *today*, not just on 07-24.

**Tier table, `gemini-pro-latest`, trivial prompt** — the 07-24 failure
reproduced exactly:

| max_tokens | finish | visible tok | thinking tok | chars |
|---|---|---|---|---|
| 100 | MAX_TOKENS | 4 | 92 | 23 |
| 500 | MAX_TOKENS | 19 | 477 | 107 |
| 1200 | STOP | 38 | 1145 | 224 |
| 3000 | STOP | 37 | 888 | 213 |
| 6000 | STOP | 44 | 786 | 228 |

**Thinking expands to fill a small ceiling** (92 of 100, 477 of 500) and settles
at ~800–1,150 once the ceiling is comfortable. That is the mechanism, measured.

**Tier table on the REAL 12,570-char writer prompt** (placeholders filled), which
is the figure that sizes the reserve:

| config | max_tokens | thinking tok | visible tok |
|---|---|---|---|
| none | 8000 | **2,764** | 99 |
| none | 3000 | **2,878** | 103 |
| `thinkingLevel: low` | 8000 | 1,080 | 57 |
| `thinkingBudget: 512` | 8000 | 940 | 55 |

So the 8192 default carries ~3x headroom on the real workload. **It is now
measured rather than chosen** — recorded in the constant's comment.

> **CORRECTED 2026-07-27 — my own claim, falsified by the probe.** PLAN D2, the
> NOTES entry above, the `016b` §9 pattern and the commit message all said the two
> generations take *incompatible* knobs: "3.x takes a `thinkingLevel` string and
> rejects the integer with a 400". **False.** On `gemini-pro-latest`:
> `thinkingBudget: 512` → **ACCEPTED**; `128` → ACCEPTED; `32768` → ACCEPTED;
> `thinkingLevel: "low"`/`"high"` → ACCEPTED. Only `thinkingBudget: 0` is
> rejected, and its message says exactly why: *"Budget 0 is invalid. This model
> only works in thinking mode."*
> The 07-24 observation was right and narrow (that one value 400s). **The
> generalisation was mine**, built from a single rejected value, and it is the same
> error shape as the one this whole workstream exists to correct — reasoning from
> one refusal to a structural claim without testing the neighbours. Cheap check
> that caught it: three more values of the same parameter, ~30 seconds.
> No harm done to the code: the client sends neither knob by default precisely so
> this inference was never load-bearing. The comments asserting it are corrected.

**And a second correction, this one to the guard's rationale rather than the
guard.** Sending both knobs together IS refused — *"You can only set only one of
thinking budget and thinking level"* — so the mutual-exclusion check is right. But
my stated reason for it (two generations, two knobs) was wrong. Right check, wrong
why, which is a worse state than it looks: it would have survived any review that
read the reason instead of testing the behaviour.

**The finding that actually changes the plan: neither knob CAPS thinking.**
`thinkingBudget` is a soft target the model overshoots freely — 128 requested →
483 spent; 512 → 903/940; 32768 → 783. It reduces thinking substantially (2,764 →
~940 on the real prompt) but bounds nothing. So a knob is a **cost lever, not a
correctness one**: it cannot replace the reserve, and any plan that says "just set
thinkingBudget and drop the reserve" is wrong.

**A probe fault I nearly reported as a finding.** The first knob run printed all
three knobs as `REJECTED: contents is not specified`. That was my script: `jq
--argjson` was fed jq syntax (`{thinkingConfig:{thinkingLevel:"low"}}`, unquoted
keys) instead of JSON, so jq emitted nothing, curl posted an empty body, and the
API complained about the *missing prompt* — which reads exactly like the API
refusing the knob. **A malformed request and a refused parameter produce the same
shape of "no".** Fixed, and the script now reports a request that fails to BUILD
as `PROBE FAULT (NOT a verdict)` rather than letting it masquerade as one. Had I
believed it, I would have "confirmed" my own falsified claim with my own bug.

**Path correction, found by running it.** RUNBOOK §0's original query for the
writer's provider returned four NULL columns and no error: `generate_content` is
not a top-level step. Real path:
`workflow → steps → process_sections_loop → config → sub_workflow → steps →
generate_content → config → ai_service`. **A jsonb `->` path that returns NULL has
not told you the value is absent — it may have told you the path is wrong**, and
the two are indistinguishable without walking the keys. Also: `steps` is an
*object* keyed by step name, so `jsonb_array_elements` errors ("cannot extract
elements from an object") while `jsonb_each` works.

**Two live facts about the writer that change the framing:**
1. Its step budget is **`max_tokens: 8000`**, not one of content-creator's
   100/1200/3000/6000 tiers. At 8,000 with thinking around 2,800 there is room to
   spare — so `[INFERRED, from a context-stripped probe]` the 07-24 starvation
   probably would **not** have bitten the writer. The starvation is a
   *small-tier* defect, and content-creator's twitter/short tiers are where it
   lived. For the writer the fix is insurance plus the thinking-token visibility
   and the thought-part filter. Marked INFERRED because a real run carries site
   specs, brief, existing content and link context — a far bigger prompt than the
   12.7K template alone, and thinking scales with prompt complexity.
2. Its prompt template is **12,570 chars**, not the 7.8K recorded in the brochure
   NOTES. Grown since. Re-measure rather than quoting.

**Visible-output figures from these probes are NOT quality evidence** and must not
be quoted as such: the template's placeholders were filled with
"(context omitted for probe)", so the model had nothing real to write about. The
55–103 visible-token counts measure that, not its writing. One run at tier 3000
returned 2 characters with `finishReason=STOP` — almost certainly an empty JSON
object, and exactly the sort of number that would become "Gemini writes nothing"
if lifted out of context. The **thinking** figures are the usable output here.

## 2026-07-27 — council gate fired (P2)

**`SUBMISSION_CORR = a1a5cf20-a70d-48c3-8fda-842d2a91b651`** (council-gate,
5 edits: the three `gemini.go` areas, content-creator's cost table, the test
file). Queued behind one in-flight run (`council-gate-orchestrate-0727-1309`, at
`review_reuse_agent`); consumer live, last step advanced 16s before submission.
Do NOT re-fire — a duplicate spends the same credits and lands further back in the
same lane.

Verdict:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='a1a5cf20-a70d-48c3-8fda-842d2a91b651' AND kind='council_report'
ORDER BY created_at;
```

**Three defects in my own submission, caught by assertions I put in the builder
rather than by reading it back.** Worth recording because the submission is the
only view the reviewers get, and all three would have cost a round:

1. **One edit's `sketch` was EMPTY (0 chars).** My hunk selector matched patterns
   against diff hunk *headers* (`@@ … @@`), which carry only line numbers and the
   enclosing function — so a hunk whose *body* contained the change was never
   selected. Reviewers judge the sketch; an empty one draws confident objections
   about code they cannot see (the recorded failure in the gate RUNBOOK). Now
   matches hunk BODIES and **asserts non-empty before submitting**.
2. **A quote attributed to the wrong source.** I cited *"The queued `about`
   rebuild never ran under Gemini (still triaged behind the backlog)"* as commit
   `5db6a929f`'s message. It is not in that message — it is in the brochure NOTES
   file that commit *added*. Caught by asserting the quote appears in the text it
   is attributed to. **A quote that is verbatim-but-mis-sourced is still a false
   citation**, and reviewers cannot open files to check.
3. **Two quotes silently joined across line breaks.** Commit messages wrap, so
   `404 "no longer available\n   to new users"` is not the string I pasted.
   Grepping for the joined form returned 0 — which is how I found it. Quotes now
   preserve the wrapping and say they do.

The cheap check behind all three: **build the submission with a script that
asserts, rather than assembling it by hand and proof-reading.** Proof-reading
found none of these; three `assert`s found all three in one run.

## 2026-07-27 — council verdict: APPROVED, and what each objection got

**APPROVED, round 1**, corr `a1a5cf20-a70d-48c3-8fda-842d2a91b651`. Decided
*"approved with 4 advisory objection(s) — none high-severity"*.
`reviewers: 10, abstained: 6, unreadable: 0` — 10 + 6 = 16 = the live seat count,
so **every seat is accounted for and no opinion was lost**. (Read `unreadable`,
not `abstained`: abstentions are the relevance filter working.) Ten seats fired,
six were filtered out. Approving seats: guidelines, diagnosis_guardian,
constitution, mission, prior_art_librarian, llm_reliability. Objecting-but-not-
vetoing: editquality, reuse_agent, guardian, debug_historian.

**Every objection, and what it got. Three changed something.**

1. **reuse_agent (medium) — `geminiConfigInt` may duplicate
   `datahelpers.GetIntField`, and the plan showed no evidence I checked.** The
   seat was right that I had not checked. I have now: `GetIntField` exists at
   `platform/orchestration/datahelpers/data_helpers.go:1560`, and it is **not**
   applicable, for two reasons now recorded *in the code at the helper*:
   it returns the **default on an unrecognised type** rather than an error (so
   `thinking_reserve_tokens: "8192"` — a string, the likeliest hand-editing slip
   — would silently become the default, which is the silently-ignored-config-key
   failure this platform has a standing rule against); and it lives in the
   orchestration layer, so importing it would point a provider-client leaf at
   orchestration (no cycle today, but the arrow is backwards). **The good outcome
   is not the answer, it is that the code now carries the check** — the next
   reader and the next council see it was made.

2. **reuse_agent + prior_art (procedural) — no precedent check, and this file was
   already the subject of a revert.** Done: two prior council reports mention
   Gemini (`d35844da` 07-20, `e996bf0a` 07-18) and **both are the imagery lane** —
   `provider.go`'s routing switch and the `"gemini"` alias, i.e.
   `internal/adapters/imagegenerator` / `bugs_closed/011`. Neither touches
   `platform/aiservice/gemini.go` or text generation. So: no prior council pass on
   this file. That is an answer, not an absence-of-evidence shrug.

3. **guardian (low) — name `estimateCost`'s blast radius rather than assume it.**
   Done: exactly one consumer, `internal/agents/contentcreator/agent.go:299`, same
   file. Nothing outside it reads the table or its output.

4. **guardian (missing) — confirm no other `agent_definitions` row is on Gemini.**
   Done: `SELECT count(*) … provider": "gemini"` → **0**. Risk 6 holds.

5. **editquality + guardian (medium) — `max_tokens` is no longer a hard visible
   length cap, with no compensating clamp.** The substantive one, and I am **not**
   fixing it here. Filed as `features_open/025` with the reasoning: a clamp in the
   provider client would truncate returned text, and the writer sets
   `output_format: json`, so cutting mid-string yields an unparseable artifact —
   worse than a long one. And a *character* limit tied to a publishing platform
   does not belong in an LLM transport: Anthropic has the identical exposure the
   moment extended thinking is turned on there. The cap belongs in
   content-creator, on the tweet text, in characters, for every provider. 025 also
   carries the question guardian asked that I only partly answered: which callers
   enforce their own truncation, and which rely on `max_tokens`.

6. **llm_reliability (low) — `llm_call_log`'s `output_tokens == max_tokens` CUT
   heuristic goes quiet for Gemini thinking models.** Disclosed in the submission
   and in the code, and the seat's real point was ownership: *"should be called out
   to whoever owns that heuristic/dashboard, not just noted in code comments."*
   Also in `025`, with two options and a note that the rule is quoted in CLAUDE.md
   and 016b — a rule that silently means something different per provider is worse
   than either version of it.

7. **debug_historian (medium) — the plan names no pod-grep deploy verification.**
   Correct about the *plan*; the step does exist, in `RUNBOOK` §4 (a string this
   change creates, plus a negative control) and in `bugs_open/107`'s "How to
   verify". I left it out of the submission, which is the reviewable artifact, so
   the objection stands as written. Cross-referenced rather than re-invented.

8. **editquality (low) — edits 3/4 are scope creep past the diagnosed mechanism.**
   Fair, and accepted as a judgement call rather than changed: the retired-pin
   refusal is what makes the primary fix *reachable at all* (you cannot exercise a
   reserve against a model that 404s), which the seat itself allowed. The cost-table
   re-key is genuinely adjacent — it stays because it is three lines and was
   discovered by the same read, and unbundling it now would cost a round to gain
   tidiness.

9. **prior_art (low) — "the string 'thinking' appeared nowhere" describes a prior
   commit's state that the code index cannot see, with no diff/blame attached.**
   A fair point about *evidence form*. The claim is checkable with
   `git show 8a2b5dea0~1:platform/aiservice/gemini.go | grep -c thinking` → 0, and
   the seat noted the conclusion does not hinge on it.

**No `Council-Reviewed:` trailer is possible, and this is a known false negative.**
The approved code is in `8a2b5dea0`, `f4f2336a3` and `17136ce3c`, all of which
predate the verdict, and the repo is forward-only — no amends. So the `098`
coverage report will list them as UNREVIEWED. Same shape as `bugs_closed/011`'s
round-9 approval, already recorded in `016b` §8.2. The verdict is real and its
correlation is recorded here, in `bugs_open/107` and in the PLAN; the join is just
not machine-exact. **A trailer added by a later commit would be worse** — it would
attach the approval to code the council never saw.

## 2026-07-27 — P3 verified, P5 PROVEN IN PRODUCTION, P6 blocked on a permission

**P3 — both images live.** `v1.0.1173`, rolled 13:45:31Z (chassis) and 13:45:35Z
(content-creator). Pod-grep on `agent-chassis-5f85dff548-8d2tq`, five strings this
change CREATED, all → 1: `thinking consumed the entire output ceiling`,
`__sent_visible_budget_tokens`, `__usage_thinking_tokens`, `deliberately no
default`, `You can only set only one of thinking budget`. Negative controls: the
old format string `no text content in response (finishReason=%q)` → **0**, and an
invented string → 0. Content-creator's own binary: same positives, plus
`gemini-flash-lite-latest` → 1 with the old `gemini-2.5-flash":` key → **0**,
which proves the cost-table re-key shipped too.

> **My third negative control was worthless and I nearly reported it as one.**
> I grepped `datahelpers.GetIntField` expecting 0 and got 1 — but that symbol
> exists all over the tree and is in the binary regardless of my change, so it
> could never have discriminated. A negative control has to be a string that is
> absent *because of* the change. The old format string was the valid one.

**P5 — content-creator flipped to `gemini`/`gemini-pro-latest` and PROVEN.**
Configmap edited in git, `kubectl apply`'d, deployment restarted; live values
re-read (`provider: "gemini"`, `model: "gemini-pro-latest"`,
`api_key_env_var: "GEMINI_API_KEY"`). New pod started clean with **0 restarts** —
which is itself evidence the new construction validation passed (model present,
not a retired pin, no double knob).

Two real generations through Kafka, both on `gemini-pro-latest`:

| tier | request | result |
|---|---|---|
| `max_tokens: 100` (twitter) | `generate_social_media` | **264 chars of real tweet**, 66 tokens, 12.6s |
| `max_tokens: 6000` (long) | `generate_blog_post` | **8,726 chars / 1,292 words**, 2,181 tokens, 35.4s, no truncation |

**The 100-token tier is the whole point: on 2026-07-24 that exact tier returned
ZERO characters. It now returns a publishable tweet.** The defect is fixed end to
end in production, not just in a test.

`estimated_cost` came back `0.00066` for 66 tokens — i.e. 0.010/1000, the
`gemini-pro-latest` rate. Before the re-key it would have silently used the Claude
fallback of 0.003. So that fix is live and observable too.

**Two frictions worth recording for whoever fires the next one:**
- The first publish was rejected with `'fuel_budget' header not found` — an
  `-H fuel_budget=100000` header is required and is not in the payload schema.
  The message reached `handleMessage` and errored to
  `system.agent.content-creator.errors`; it did **not** vanish.
- `Agent instance not found, using default configuration` is expected for an
  ad-hoc probe (a random `agent_instance_id` has no DB row). It means
  `GetDefaultConfig` supplied the tiers, so these runs exercised the shipped
  defaults, which is what we wanted.

**P6 — NOT DONE, blocked on a permission, not on knowledge.** The `UPDATE` to the
live `agent_definitions` row was refused by the tool-permission classifier. The
statement is written, reviewed and ready as
`P6_FLIP_page_content_writer.sql` in this directory, with a transaction and a
`DO $$` block that RAISEs (rolling back) unless all four post-conditions hold.
Backup `bak_agent_definitions_pcw_20260727` is already created.

> **Writing that script found a bug in my own RUNBOOK §7 that would have quietly
> cut the writer's budget by 4x.** The original SQL replaced the whole `ai_service`
> object with `{"provider","model","api_key_env_var"}`. **`max_tokens: 8000` lives
> inside that same block** — so the replace would have dropped it and
> `NewGeminiClient` would have fallen back to its 2048 default. Invisible in the
> diff; it would have surfaced days later as truncated page sections, and I would
> have gone looking at the reserve. Caught by reading the row before writing it:
> a query for the *step's* `max_tokens` returned NULL, because the key is one level
> in. **`jsonb_set` with a literal object is a REPLACE, not a merge** — use `||`
> on the existing object whenever the block has siblings you did not enumerate.
> Corrected in the RUNBOOK, and the script asserts `max_tokens = 8000` after the
> write rather than trusting it.

Also noted: the writer row's `updated_at` was `13:44:56.343485+00` — the
architecture-review seat's re-seed, minutes before I read it. The guard is not
theoretical on this row.

## 2026-07-27 — the owner's question turned a "feature" into a regression of mine (110)

Owner asked whether the truncation-detection problem needed its own bug listing.
Answer: **yes**, and reading the schema instead of my own intent showed why it is a
**regression I introduced**, not a gap we never filled. Filed `bugs_open/110`.

**Fact 1 — `llm_call_log.max_tokens` was about to carry two meanings.** It is fed
solely from `options["__sent_max_tokens"]` (`ai_actions.go:390-392` →
`llm_call_logger.go`). `anthropic.go:118` and `ollama.go:101` set that to the
caller's answer-budget. `gemini.go` set it to the caller's budget **plus the 8192
reserve**. One column, two definitions, split by provider — **which is precisely
107's own finding, reproduced one layer up by the fix for it.** Each layer reads
correctly on its own: `__sent_max_tokens` genuinely *did* record what was sent.
That is why the shape recurs, and why "I just fixed this class" is no protection.

**Fact 2 — `features_open/025` item (b) proposed a repair the schema cannot
support.** It said to compare `usage_output_tokens` against
`sent_visible_budget_tokens`. **Neither column exists** — the table has
`max_tokens` and `output_tokens`. I had written both names from memory of my own
field names, never checked against `\d llm_call_log`, and the RUNBOOK carried the
same two invented names. Superseded by 110; item (a) stands.

**Fact 3 — an overclaim, mine, in a committed bug file and a council submission.**
`107`, corr `a1a5cf20` and three commit messages say this change makes thinking
tokens *"visible to logging"* / *"surfaced"*. **Half true, and the false half is
the load-bearing one.** All four new fields —
`__sent_visible_budget_tokens`, `__sent_thinking_reserve_tokens`,
`__usage_thinking_tokens`, `__usage_total_tokens` — have **no reader outside
`platform/aiservice/`** (grep) and no column to land in. Thinking is visible in the
*error message* and the in-process options map, and **nowhere a query can reach**.
`016b` §9's *"a field is only as live as its LAST reader"*; same shape as `101`'s
inert `scrape_web` keys.

**Why nobody caught it, including a ten-seat council one of whose seats discussed
these exact fields:** *"writes the field"* and *"the field is readable"* look
identical in a diff. The reviewers could see the write. Nothing in the submission
showed the absent reader, and I did not think to look. **The check is one grep of
the field name with your own package excluded** — the same check `101` earned for
config keys, applied to telemetry fields.

**Timing, and it is lucky rather than clever.** `llm_call_log` today:
`anthropic 43,586 · ollama 808 · gemini 0`. Zero because **`content-creator` does
not write to that table at all** — only `platform/orchestration/actions/*` does —
so P5's two proven generations logged nothing there. The first Gemini row arrives
when `page-content-writer` runs on Gemini, i.e. **the moment P6 lands**. So the
window to fix this before any wrong row exists was still open when the owner asked
the question.

**Candidate 1 applied** (`__sent_max_tokens` = the caller's visible budget;
wire total moved to `__sent_wire_max_output_tokens`, still unpersisted). Three
lines, no migration, tests updated to assert the *meaning* rather than the number.
**It inverts a decision I made deliberately in `8a2b5dea0`**, where I argued the
field name says "sent" so it should carry the wire value — I optimised for one
field's local honesty and broke the cross-provider comparability of the column the
platform-wide truncation rule depends on. Cross-provider meaning wins in a column
called `max_tokens`.

> **It is INERT until the next chassis roll.** v1.0.1173 still logs the inflated
> total, so if P6 runs before that roll, its first rows carry `max_tokens = 16192`
> where the caller asked 8000. Those rows are identifiable —
> `provider='gemini' AND max_tokens = 16192` — and are recorded here so nobody
> later reads them as a mystery. **I am not holding P6 for it:** the owner's actual
> question is whether Gemini writes acceptably, and delaying that to protect a
> telemetry column whose wrong rows are self-identifying would be the wrong
> priority. Stated so it is a decision, not an oversight.

## 2026-07-27 — P6 DONE (writer on Gemini), P7 partly done and deliberately stopped short

**P6 applied.** `P6_FLIP_page_content_writer.sql` ran clean:

```
--- BEFORE --- {"model":"claude-sonnet-4-6","provider":"anthropic","max_tokens":8000,...}
UPDATE 1
--- AFTER  --- {"model":"gemini-pro-latest","provider":"gemini","max_tokens":8000,...}
NOTICE: OK: gemini/gemini-pro-latest, max_tokens 8000 preserved, style block intact.
COMMIT
```

**`max_tokens: 8000` survived** — which is precisely what the `jsonb_set` replace
would have deleted. The assertion block earned its place on its first run: it is the
difference between knowing that and assuming it. `tmpl_chars` still 12,570, style
block intact. Backup: `bak_agent_definitions_pcw_20260727_p6`.

**P7 — I did NOT rebuild a real page, and that is a deliberate stop, not an
omission.** The candidates were all live pages owned by other workstreams
mid-flight: `fundamentallyai`'s `about` (brochure, actively working 085/109 — and
the site whose baseline snapshot makes it the *ideal* comparison), and `vonc.com`
(gauntlet, which has just got that site to `claimscan 0/49` — regenerating copy
could reintroduce fabrications and undo it). Pool sites have **no pages**
(`pool-ai-agents.internal` → 0 rows), so the usual scratch target does not exist
for a page build. Mutating someone's live estate to answer my question is not mine
to decide.

**What I tested instead, because it isolates the risky part.** The writer's real
12,570-char `prompt_template`, placeholders filled with genuine material and the
section schema appended, at its real `max_tokens: 8000`, direct to
`gemini-pro-latest`:

| measure | result |
|---|---|
| finish reason | **STOP** (not MAX_TOKENS) |
| visible / thinking | 120 / **1,576** tokens — the 8192 reserve is ~5x what was needed |
| output | **valid JSON, unfenced, all four required keys present** |

JSON adherence was the real risk: the writer sets `output_format: json`, and a
thinking model that wrapped its answer in a ``` fence or truncated mid-string would
break the section writer in a way a prose test would never show. It did neither.

**The copy, measured against the Voice & Style rules rather than eyeballed:**

| rule | result |
|---|---|
| no em dashes | **0** |
| no exclamation marks | 0 |
| filler words (crucially/seamless/robust/leverage/delve…) | **none** |
| fact-first opening, no negative frame | **yes** — "FundamentallyAI is an AI consultancy." |
| one idea per sentence | 6 sentences, mean 9.2 words |
| one rough edge left standing | **yes** — "Our systems can produce incorrect output." |
| contractions | **absent — the one rule it missed** |

For context, the Claude + v1-prompt test on 07-24 got em dashes **19 → 14**
("down, NOT gone — rule partially obeyed"). This run scored **0**.

> **Do not read that as "Gemini obeys the style prompt better than Claude."** It is
> **n=1 against n=1**, on *different* prompt content and a different section, and my
> harness differs from a real run in four ways: no `site_specs` / `brief` /
> `existing_content` / `link_context` (I substituted a Material block), no
> `appendOutputInstructions` from the chassis, no `process_sections_loop`, so no
> multi-section coherence, and no deploy. **What it does establish** is narrower and
> still worth having: the writer's prompt shape works on Gemini, returns parseable
> JSON of the right shape, does not truncate at 8000, and the style block transfers
> to a different model family rather than being tuned to Claude. That last one was a
> live risk nobody had tested.

**Still owed, and it needs an owner decision on the target:** one real page build
through `process_sections_loop` on a site the owner nominates, then read the copy
and check the page's own story survived (bug-056 vigilance). Until then P7 is
**partial** and `bugs_open/107` stays OPEN.

**Also owed at the next chassis roll:** `110` candidate 1 is inert, so any writer
run before that roll logs `max_tokens = 16192` where the caller asked 8000. Those
rows are self-identifying (`provider='gemini' AND max_tokens = 16192`).

## 2026-07-27 — P7's page build is BLOCKED by bugs_open/029, which has nothing to do with Gemini

Owner nominated dartsonline. Queued `grip-styles` (status `planned`, **never
deployed**, so a bad result costs nothing) as a `needs_page` / `page-build-handler`
item, `pipeline='build'`, `status='triaged'`, work item
`9fdb87b4-56bf-4981-a899-318e6294c08d`.

**It never got claimed.** Eight minutes, `attempt_count` still 0. The chase:

- `build-pipeline-trigger` fires every 120s and is enabled. Its `pre_query` returned
  `pending_sites: "1"` — my site, correctly detected. So detection works.
- The site is unlocked, and my item satisfies every clause of that query
  (`pipeline='build'`, `triaged`, `attempt_count 0 < max_attempts 3`). It is the
  **only** eligible build item fleet-wide.
- Its `seed_queue` step reports `{"total":0,"seeded":0,"skipped":0}` — **a red
  herring I nearly chased.** That step is `seed_build_queue`, which creates records
  for brand-new sites from `build_queue`. Zero is correct there.
- The runs get to `spawn_dispatch` (`spawn_agent` → `build-dispatch-loop`) and sit
  at `AWAITING_RESPONSES`. Three of them, idle 134s / 284s / 430s.
- `build-dispatch-loop` orchestrations: **the most recent is CANCELLED, 2026-07-19.**
  Nothing has run since.

That is `bugs_open/029_..._hung_spawns_saturate_dispatch_group_and_halt_builds_fleetwide`
verbatim: *"builds simply stop happening everywhere, and the scheduler keeps firing
every 30 seconds into a full pool."* **Filed 2026-07-19, still OPEN**, and it is the
consequence half of `003` (spawn loses the child response). Corroborating:
`page-build-handler` runs on 07-26 FAILED at `spawn_content_writer` and
`call_content_writer`.

**So P7 cannot complete through the normal build path, for reasons that predate this
workstream by eight days and are unrelated to the model swap.** I am not fixing 029
— it is a filed, owned fleet outage, and starting a competing fix is exactly what the
coordination rules forbid. The work item stays queued; it will build when 029 lifts.

Also relevant if someone retries this: `grip-styles` has **0 `page_components`**, so
even a working rerender has nothing to re-render — that is `bugs_open/087`'s territory
(*page_rebuild writer has no section plan and nothing builds one*). A page that is
`planned` needs the full build path, not the rerender path.

**What was produced instead, on the owner's actual brief.** content-creator is on
Gemini and working, so the in-depth darts content went through that: **1,458 words,
8,570 chars, 2,146 tokens, 43.5s, est. £0.021**, corr
`16f7535d-6f8c-4149-b650-44764e1342ec`. Saved as
`SAMPLE_2026-07-27_darts_technique_gemini_output.md`. **Not published and not
claim-checked.**

Audited it rather than admiring it. The brief forbade invented statistics, results
and quotes, and that held: **no tournament results, no dates, no quotes**, and every
number is game arithmetic (501, 15-dart legs, treble 20, a 141 finish on double 12 —
which checks out) rather than a statistic about a person. It does make **technique**
claims about seven named living players. Those are checkable in principle and
**unchecked by us**, which is fine for a sample and would not be fine on a page.

> **Misstep, mine, caught in the same breath.** My first extraction of the article
> printed `Letâs` and `bedâusing`, and I was about to report an encoding bug in the
> pipeline. It was **my** extraction: I decoded the JSON payload with
> `unicode_escape`, which mangles UTF-8. Re-read with plain `json.loads` on the log
> line and the text is clean — zero mojibake. **A garbled artefact is evidence about
> your reader until you have proved otherwise.**

## 2026-07-27 — the model-vs-prompt question, measured (5 runs each, same prompt)

Owner asked which model's content is better, or whether it's much of a muchness and
really about the prompt. Neither of my earlier samples could answer that: Gemini n=1
on a hand harness, Claude n=1 from 07-24 on an *older* prompt version and a different
page. Comparing them would have been a guess wearing a number.

So: same prompt (the real writer `prompt_template`, placeholders filled identically),
same material, same visible budget (8000), **5 runs each**, mechanically scored.
Script `RUN_model_comparison.py`, raw output `DATA_2026-07-27_model_comparison_5x2_runs.json`.

| metric | gemini-pro-latest | claude-sonnet-4-6 |
|---|---|---|
| valid JSON, all keys | 5/5 | 5/5 |
| em dashes | 0.0 | 0.8 (0–2) |
| filler words | 0.0 | 0.0 |
| exclamations | 0.0 | 0.0 |
| **negative-frame sentences** | **0.0** | **1.4 (0–2)** |
| **"not X, it's Y" construction** | **0.0** | **0.8 (0–1)** |
| contractions | 0.8 (0–1) | **4.8 (4–5)** |
| raw HTML tags in fields | 0 | 6.0 |
| chars / sentences / mean words | 382 / 7.4 / 8.2 | 602 / 6.4 / 15.2 |
| **billable output tokens** | **1,921** (106 visible + 1,815 thinking) | **188** |
| latency | 12.3s (9.7–15.4) | 4.6s (4.1–5.5) |

**The answer to the question as asked: mostly the prompt.** Both models produce
recognisably house-style copy. Filler, exclamations and hype are at zero for both.
Neither invents a negative headline. On the things the style block was written to
stamp out, the two are close, and both are far from where the copy was before the
block existed.

**Where they differ, they differ consistently, and not in the direction I expected.**
Gemini scored **0 negative-frame sentences across 5 runs; Claude scored 7** — with
the exact construction the owner personally caught across three rounds of refining
that prompt: "Not assistants. Not chatbots." "That's not a case study." Rule 3 is the
rule this house style is *most* about, and Claude persistently breaks it.

Claude wins the contractions rule (4.8 vs 0.8) and writes 58% more, in sentences
nearly twice as long. Whether "one idea per sentence" at 8.2 words reads better than
15.2 is the owner's call, not a metric.

> **My first scoring pass missed the negative-frame result entirely**, because it
> only checked the *headline* for a negative opening. The violation was in the
> *subheadline* and *body*. I found it by reading the samples, not by running the
> scorer. **A mechanical score is only as good as the field it looks at**, and a
> metric that returns 0/5 for both models is exactly as convincing as one that
> found nothing to look at.

**Claude fences its JSON on 5/5 runs and Gemini never does.** Not a risk: the
pipeline strips ```json at `v3_site_actions.go:2806`, which is why Claude has been
serving the writer for months without trouble. Recorded so nobody counts it as a
Gemini advantage. The raw `<p>` tags Claude emits into JSON string fields are a
different matter and depend on the component template.

**The finding that should decide this workstream is cost, not quality.** Gemini
spends **1,815 thinking tokens per section** against Claude's zero, so ~**10x the
billable output tokens** for one hero section, at **2.7x the latency**. Thinking
bills as output. `page-content-writer` runs *per section, per page, across the
estate*, so that multiplier lands on every build. `[UNVERIFIED]` per-token rates put
it at roughly £0.019 vs £0.003 per section, but the **token ratio is measured and the
rates are not** — trust the 10x, not the pounds.

**Caveats, so this isn't over-read.** n=5. My harness still lacks `site_specs`,
`brief`, `existing_content` and `link_context`, and the chassis'
`appendOutputInstructions`; a real section carries more context and may think longer,
not less. And "better prose" is not decidable by any of these counters — ten samples
are in the JSON for a human read.
