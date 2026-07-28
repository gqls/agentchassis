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

---

## 2026-07-27, triage sweep after the v1.0.1174 roll — the blocker was never 029

Fleet rolled to **v1.0.1174 at 15:11 UTC**. Chassis binary built 14:58 UTC; last Go
commit before it `e96d42226` at 14:52 UTC, so every commit of the 107/110 work is in
the running image. (Clock trap: this box is BST, `kubectl` is UTC. All times here UTC.)

### What is now verified live, against pods rather than git

| check | evidence |
|---|---|
| 107 client fix in the binary | `agent-chassis-5994dc6d6c-pt8v9`: `grep -c "thinking consumed the entire output ceiling"` → **1** |
| negative control | pre-fix format string `no text content in response (finishReason=%q)` → **0** |
| 110 candidate 1 in the binary | `__sent_wire_max_output_tokens` → **1**, `__sent_visible_budget_tokens` → **0** (the rename is the only pre/post discriminator) |
| images | `agent-chassis` and `content-creator-agent` both `v1.0.1174` |
| P6 flip | writer step = `{"model": "gemini-pro-latest", "provider": "gemini", "max_tokens": 8000, "api_key_env_var": "GEMINI_API_KEY"}` — the 8000 survived |
| no shadowing `ai_service` | exactly ONE `ai_service` in the whole `page-content-writer` definition; root and `config` both ABSENT (the `bugs_closed/009` shape, checked not assumed) |
| Gemini calls ever made through the chassis | `SELECT count(*) FROM llm_call_log WHERE provider='gemini'` → **0** |

### MISSTEP, mine, and it sent the previous session's handoff at the wrong bug

I wrote (in the handoff, in `107`, and in commits `bfbbb7cfa` / `5bd32602a`) that
**`bugs_open/029` had halted every build since 19 July** and that P7 was blocked by it.
**That was false, and it was already false when written.**

```sql
SELECT date_trunc('day',created_at)::date AS day, status, count(*)
FROM orchestration_states WHERE owner_agent_type='build-dispatch-loop'
GROUP BY 1,2 ORDER BY 1 DESC;
--  2026-07-27 | COMPLETED | 30
--  2026-07-26 | COMPLETED | 62
--  2026-07-24 | CANCELLED |  2
```

One query, and it contradicts the claim outright. Page builds were completing too:
`ai-agent-orchestration.com/model-directory` COMPLETED at 02:27, 08:27 and 14:27 on
07-27. And `029`'s own corrected diagnosis (`23e58e1bf`, filed 07-21, six days before
I cited it) already said the trigger is **an image roll** — a transient window, not a
standing outage. I cited a bug's headline without reading its correction.

**What the `grip-styles` run actually did.** It was claimed and it ran, 15:46 UTC, in
`agent-page-build-handler-8bf4fb08-8hfvq`:

```
spec_sections : {"count": 0, "source": "none", "sections": []}
plan_sections : {"reason": "no sections to plan", "ready_count": 0}
check_has_ready_sections : {"condition_met": false, "next_step_override": "mark_no_ready_sections"}
mark_no_ready_sections   : {"status": "needs_human_review"}
```

dartsonline has **no `site_plan` aspect** in `site_specs` (its aspects are audience,
briefing, classification, content_direction, design_intent, identity,
resolved_composition, strategy, submission, vertical_landscape) and
`pages.sections` for `grip-styles` is `[]`. `load_page_sections_from_spec` reads the
first and falls back to the second, so it found nothing. **A bad target, not a broken
pipeline.** Note `orchestration_states.error` was NULL and `status` was `COMPLETED` —
the failure is only visible in `collected_data`, exactly as the "`COMPLETED` with
`error` NULL" trap warns.

Not `bugs_open/087` either: 087 is the `page-rebuild` path and explicitly records
`page-build-handler` as unaffected. I checked before reaching for it.

### The real blocker, found by asking one more question — filed as `bugs_open/112`

Having established the pipeline works, I went looking for what would happen at the
next build, and checked whether the chassis could even reach Gemini. The main pod can:

```
agent-chassis-5994dc6d6c-pt8v9 : GEMINI_API_KEY PRESENT len=53
```

But `page-content-writer` does not run there. It runs in its **own spawned pod**
(`orchestration_states` → `__execution_context__.sender.pod_name` →
`agent-page-content-writer-47953cd4-k8bmn`, and four more). And a live spawned pod says:

```
agent-build-dispatch-loop-b62d9c1a-pnnt9 | v1.0.1174 | GEMINI ABSENT
```

Its env list: `ANTHROPIC_API_KEY`, `GROK_API_KEY`, `FIRECRAWL_API_KEY`, ... and no
Gemini. Cause is in Go, not config — `platform/orchestration/actions/spawn_actions.go`
:2440-2518 builds the spawned pod env as an **explicit allow-list**, injecting
`ANTHROPIC_API_KEY` and `GROK_API_KEY` by name from `personae-default-secrets`. There
is no `GEMINI_API_KEY` block. `grep -rn "GEMINI_API_KEY" --include=*.go` returns
**two hits in the whole tree**, both inside `platform/aiservice/gemini.go` (a comment
and an error string).

The secret itself is fine — `personae-default-secrets` does contain `GEMINI_API_KEY`.
The main Deployment picks it up via `envFrom: secretRef`; spawned pods do not use that
`secretRef` at all. **Two provisioning routes for one credential, and only one was
updated.** `content-creator-agent` is a standalone Deployment with its own explicit
`GEMINI_API_KEY` patch, which is precisely why P5 passed and P6 cannot.

`gemini.go:157-160` fails loudly (`API key environment variable 'GEMINI_API_KEY' is
not set or empty`), and `generate_content` has **no `error_step`**, so the failure
takes the whole page build with it.

### Time-sensitive, because the flip is already armed

`scheduled_tasks`: `model-directory-publish` → `model-directory-trigger`,
`interval_seconds=21600`, enabled, `last_triggered_at` 07-27 14:25 UTC, so **next due
~20:25 UTC**. That build ran fine at 14:27 on `claude-sonnet-4-6` (71
`page-content-writer` rows in `llm_call_log`, step names
`process_sections_loop_iter_N_generate_content`) — seven minutes before the P6 flip
landed at 14:34. The next one will hit Gemini, and `112`.

So P6 did not just fail to get verified; it **armed a fleet-wide page-build failure**
that has not fired yet only because no build has reached the writer since. Either fix
`112` or revert the writer's `ai_service` to anthropic before ~20:25 UTC.

**[UNVERIFIED] and worth stating:** I did not induce the failure. The prediction is
read off the code path and the pod environment, not observed. The 20:25 build will
settle it either way, and that is a cheap observation to make rather than a risk to run.

### One thing the cost data already said

Independently of `112`, the 5x2 comparison above found Gemini spends ~1,815 thinking
tokens per section against Claude's zero — roughly **10x the billable output tokens**,
at 2.7x latency, on an agent that runs per section across the whole estate. If the
owner's answer to that is "not worth it", then `112` need never be fixed for this
workstream's sake and the writer simply stays on Claude. That is an owner call, and it
is now the *first* question, not the last one.

## 2026-07-27 — four-model bake-off (Gemini · Sonnet 4.6 · Grok · Fable 5), and the dryness is partly OUR PROMPT

Owner: stay on Gemini and watch costs; prefers the Gemini copy because it sounds
less like AI, but finds it **"very dry… a straight up list of facts without much
personality."** Asked to put the same prompt to Grok (news API key exists) and to
Fable 5.

Same harness as the 5×2 run: real writer `prompt_template`, identical material,
`max_tokens: 8000`, 5 runs each. Grok via `https://api.x.ai/v1/chat/completions`,
model `grok-4-1-fast` (the model our news lane already uses; the platform itself
calls `/v1/responses` — `feed_actions.go:742` — because it needs the web_search
tool array, which a plain completion does not). Fable 5 per the `claude-api`
skill: **send no `thinking` field at all** (it is always on; an explicit
`disabled` or `budget_tokens` both 400), no sampling params, and check
`stop_reason == "refusal"` before reading content.

| metric | gemini-pro-latest | claude-sonnet-4-6 | grok-4-1-fast | claude-fable-5 |
|---|---|---|---|---|
| valid JSON / all keys | 5/5 | 5/5 | 5/5 | 5/5 |
| fenced ``` | 0/5 | **5/5** | 0/5 | 0/5 |
| em dashes · filler | 0 · 0 | 0 · 0 | 0 · 0 | 0 · 0 |
| **negative-frame sentences** | 0.0 | **0.4** | 0.0 | **0.0** |
| raw HTML tags | 0.4 | 3.6 | 0.0 | 1.6 |
| contractions | 1.2 | 3.6 | **0.0** | 1.6 |
| mean words/sentence | 7.6 | 14.9 | **4.8** | 12.5 |
| chars | 422 | 557 | **262** | 527 |
| billable output tokens | **2,425** (2,311 thinking) | 172 | 76 (+675 reasoning) | 767 |
| latency | 17.1s | 4.3s | 6.9s | 13.2s |

**Grok is the owner's complaint amplified, not fixed.** Verbatim: *"We build
multi-agent systems. The systems produce finished work. That work includes
research. It includes copy. It includes imagery. It includes layout."* — 4.8 words
per sentence, zero contractions, 262 chars. It obeys "one idea, one sentence"
hardest of the four and is by far the driest. It also breaks our own **rule 13**
(don't repeat a sentence template for cadence) while obeying rule 1.

**Fable 5 is the direct fix for dryness.** Verbatim: *"If you're weighing whether
to build this capability or buy it, the fleet is work you can inspect before you
decide."* That has a point of view, a contraction, and a reason to care — and
**0 negative-frame sentences**, where Sonnet averages 0.4 and reliably reaches for
*"Not assistants. Not chatbots."* It is also the only model besides Gemini that
never fences its JSON. Cost: **~$0.073 per hero section** (in 3,444 / out 767 tok
at $10/$50 per MTok, rates from the skill) vs Gemini's **~$0.024** and Sonnet's
**~$0.010**. So ~3x Gemini, ~7x Sonnet.

> **Token counts and cost run in OPPOSITE directions here, which is worth not
> misreading.** Gemini spends **3x** the billable output tokens of Fable (2,425 vs
> 767) and costs **1/3** as much, because its per-token rate is a fifth. A "which
> model is cheaper" question cannot be answered from the token column alone.

**The finding that matters most: the dryness is partly our own prompt.** The Voice
& Style block is almost entirely **prohibitive** — no em dashes, no filler, no
negative frames, no exclamation marks, one idea per sentence. It says almost
nothing about *having something to say*. Pushed to its limit, "one idea, one
sentence" produces exactly the staccato fact-list the owner is objecting to, and
**Grok proves the mechanism**: it obeys hardest and is driest.

Tested it. Added **one** positive clause to the same prompt, changed nothing else,
re-ran Gemini 3x:

> *Say why it matters to the reader, not just what is true. At least one sentence
> should give them a reason to care that they could not have guessed from the facts
> alone. Write like someone with a point of view who has done this work, not like a
> specification being read out.*

Result: **422 → 611 chars mean** (528/636/670), still 0 em dashes, mean words per
sentence 8.0–9.5 (still short). And the copy acquired an argument:

> *"Chaining language models together is easy for a single demo. Making them
> recover from errors takes months of engineering time. You don't have to build
> that architecture from scratch."*

That is the personality the owner said was missing, on the **same model, at no
extra cost**. `[n=3, one section, one page]` — not proof, but it is the cheapest
thing to try before paying 3x for Fable.

**Residual tic in the warmed output:** it now opens twice — *"We build multi-agent
systems that produce finished work… We construct multi-agent systems that execute
entire workflows."* A near-duplicate restatement. Worth a dedupe clause if this
goes further.

**Recommendation to the owner:** keep Gemini, add the positive clause, and re-read
one page. Reach for Fable 5 only if that still reads flat — it is the best prose of
the four and it is ~3x the cost. **Grok is out on this evidence.**

## 2026-07-27 — Voice & Style v4 APPLIED, and the choppiness was our rule 1

Owner: use Gemini with the improved prompt, make it the default for all content
unless overridden — and asked why Gemini won't write *"...easy for a single demo
but to try to have them recover..."*, i.e. why it never joins clauses and always
takes the most obvious word.

**Answer: our own rule 1 forbade it.** The live block read:

> `- One idea per sentence. If a sentence chains two or more ideas with commas or dashes, split it.`

That is an instruction to split exactly the sentence the owner wanted. Gemini
follows the block most literally of the four models tested, so it landed driest —
and Grok, which follows it harder still, produced *"It includes copy. It includes
imagery. It includes layout."* **The choppiness was not a model trait. It was
obedience.**

**v4, tested before applying** (5 runs, `gemini-pro-latest`, same prompt/material):

| | v3 (live) | v4 |
|---|---|---|
| chars | 422 | **637** |
| mean words/sentence | 7.6 | **12.1** |
| sentence-length SD | — | 5.7 |
| conjunction-joined clauses per run | ~0 | **2.2** |
| em dashes · negative frames · filler | 0 · 0 · 0 | **0 · 0 · 0** |

The requested sentence appears verbatim: *"Chaining models together is easy for a
single demo, but getting them to recover from errors takes months."*

Four changes: rule 1 **rewritten** (not appended to — see below); "say why it
matters"; "don't always take the most obvious word"; "don't restate your opening"
(that last one fixes the near-duplicate opening the earlier warm test produced).

> **Rule 1 had to be REPLACED, and this is the reusable bit.** Appending a softer
> rule below a stricter one leaves the prompt self-contradictory, and these models
> resolve that by obeying the earlier, more literal instruction. A prompt is not a
> pile of preferences; a contradicted rule is a bug, not a nuance. The apply script
> asserts the old wording is **gone**, not merely that the new wording is present.

**Applied to `page-content-writer`** via `APPLY_voice_v4_page_content_writer.sql`
(transactional, backs up to `bak_agent_definitions_pcw_voice_v4`, and RAISEs to roll
back unless six conditions hold — including that the em-dash rule survived).
Result: `OK: rule 1 rewritten, 3 rules appended, em-dash rule intact (16149 chars)`.
Live immediately; 12,570 → 16,149 chars.

**"Default for all content" is TWO places, not eighteen.** 30-day LLM call counts:
`page-content-writer` **2,330**, `content-gap-planner` 9, `content-reviewer` 2, and
the other **fifteen** `content-creator-*` / `content-writer` agent definitions
**zero**. They are dormant. Pasting the block into fifteen unused agents would
create fifteen copies of a contract nobody exercises — the drift class, for no gain.

The real second place is **`content-creator-agent`**, the blog/social service. It
does not write to `llm_call_log` at all, so it appears in no usage query, **and it
had no house style whatsoever** — every blog post and tweet this platform has ever
produced was written with none. Fixed in Go (`buildEnhancedContentPrompt` now
appends `voiceStyleBlock(config)`), **inert until the next content-creator roll**.

**The override mechanism, since the owner asked for one:**
`core_logic.voice_style_block` — present and non-empty wins; **present and empty
means explicitly off**; absent means inherit the default. The empty case is
deliberately distinct from absent, because "unless overridden" needs a way to say
*off*, and an absent key cannot express that. Same present-but-empty distinction
`bugs_closed/009` had to add a guard for.

> **I created a second copy of the rules, and filed it rather than hiding it.** The
> block now exists as a DB literal (page-content-writer) **and** a Go `const`
> (content-creator) — two hand-maintained copies of one contract, in two substrates,
> changed by two different mechanisms. That is the council-roster drift class
> verbatim. **`features_open/026`** proposes the single-source fix (prefer candidate
> 1: one row both read). Both sites carry a `[KNOWN DUPLICATION]` comment pointing
> at it. **Until then: change both, or neither.**

**Not yet done:** no page has been rebuilt on v4. The measurements above are the
harness, not a real section with `site_specs`/`brief`/`link_context` loaded. And
`bugs_open/112` still blocks the writer from reaching Gemini in a spawned pod until
that image rolls.

## 2026-07-27 — the duplication FIXED, refiled as bugs_open/121, and three corrections from the owner

Owner: file 026 as a **bug** the architecture council should have caught; the
override I built is not what was meant; the example wording is poor; **one place
for the prompt, and not in Go.** Then: fix it.

**Refiled as `bugs_open/121`, `features_open/026` deleted.** It is a defect with a
wrong output (two voices on one estate), not something we want to build.

**The architecture seat has never reviewed anything.** Measured:
`SELECT count(*) ... FROM diagnosis_artifacts WHERE kind='council_report'` filtered
for the seat → **0 mentions, 0 reviews**, across every report ever written. It is
seeded and live, but its own handoff records why it is silent: rate-limited on
owner-approved specs, both owned by other threads. So my duplication went through a
ten-seat council on `a1a5cf20` with `unreadable: 0` and the one seat whose remit is
"two things that must stay identical" was not among them. **A seat that has never
fired is not coverage — it is worse than an absent seat, because the roster counts
it.**

**Four missteps, mine, recorded in 121 in full:**
1. **I created the duplicate inside the commit message warning about duplication.**
   `d39995125` says *"I CREATED A SECOND COPY OF THE RULES AND FILED IT RATHER THAN
   HIDING IT"* — I saw it, described it accurately, and shipped it anyway behind a
   `[KNOWN DUPLICATION]` comment. **A comment naming a defect is not a mitigation.**
   It felt responsible, which is exactly why it was the wrong call.
2. **I filed it as a feature**, which put a defect in the queue nobody treats as
   urgent.
3. **I invented an override the owner never asked for** — a `voice_style_block`
   config switch with a present-but-empty opt-out, and a paragraph defending the
   empty-vs-absent distinction. The owner meant *"a request has its own prompt in
   the request"*. I designed against an imagined requirement and then justified it.
   **Cheap check: when a directive contains a word like "override", ask what it
   means before building the mechanism.**
4. **I put prose in Go.** Prompt text a non-engineer may want to edit does not
   belong somewhere that needs a compile and a fleet roll.

**The fix, applied:**
- **Migration 240 (APPLIED)** — canonical block in `agent_default_configs`,
  `config_name='voice_style_block'`, **2,499 chars**, guard verified: refuses if the
  text is implausibly short or contains an em dash (the rule teaching its own
  opposite).
- **`platform/voicestyle`** — reader + 60s cache, **no text**. Takes a fetcher
  closure because the chassis holds `*sql.DB` and content-creator holds a
  `pgxpool.Pool`. An unavailable block degrades to "no house voice", never to a
  failed generation: losing the voice on one page is recoverable, failing every
  content build is not.
- **Chassis** — injects `{{.voice_style}}` into the template data of *every*
  `execute_llm_prompt`, so any prompt opts in by naming it, and does **not**
  overwrite a value a step already supplied. That is the request-level override the
  owner actually meant.
- **content-creator** — reads the same row; the Go const is deleted.
- **Migration 241 (WRITTEN, NOT APPLIED)** — swaps page-content-writer's literal for
  the placeholder.

> **241's gate is not ceremonial, and this is the trap worth carrying.** The prompt
> renderer is `missingkey=zero`. An unresolved `{{.voice_style}}` renders as
> **nothing** — no error, no log line. Applying 241 before the chassis carries the
> injection would silently delete the house voice from every page build, and the
> only symptom would be that the writing got worse. The file carries the two
> pod-greps that must pass first.

**Wording fixes in the canonical text:** the example now reads *"building them to
recover from errors"* rather than *"getting them to recover"*, and the underlying
point is promoted to its own rule — *"Name the action, not a vague gesture at it."*
The owner thought this one was unfixable by prompt and suggested leaving it; it was
worth fixing anyway, because **a rule's own example is the strongest signal in a
style prompt** — it is the thing the model imitates most directly, so a vague
example teaches vagueness.

**Still owed:** 241 after the roll; then rebuild one page and READ it, because a
blanked placeholder and a working one look identical until you read the output.

## 2026-07-27 — the chassis did not need rolling; content-creator did. One copy of the voice now exists, and it is live.

Owner: roll the chassis when it's ok. **It was already done, by someone else.**

`make build-*` builds from committed HEAD, so another session's **v1.0.1177** build
at 20:24 (`403f67920`) picked up my commits without either of us coordinating.
Pod-grep on the running chassis: `voice_style_block` 1, the exact SQL 1,
`__sent_wire_max_output_tokens` 1. **Rolling again would have killed two councils
mid-flight for nothing** (two `EXECUTING_STEP` at 2-3s idle when I checked).

**Proved it runs, not just that it shipped.** A pod-grep proves the code is in the
binary, never that it is on the path. Chassis logs, last 30 min, on a live council
run: `"available_fields":[..."voice_style"...]` and
`"voice_style":"These rules outrank any instinct toward \"compelling mark...`. That
is the canonical row, read from the DB, injected into real template data.

**Migration 241 was REFUSED BY ITS OWN GUARD on the first attempt, and that refusal
paid for itself twice.**

```
ERROR: template is only 289 chars - too much was cut. ROLLING BACK.
```

I had assumed the Voice & Style block was the final section and truncated at the
anchor. It is not: the block sits at **char 272 of 16,150** and ends at
`## Company Context`. The row was left untouched.

> **The refusal surfaced a SECOND defect I had already shipped.** The v4 apply
> appended the three new rules to the end of the **template**, not the end of the
> **block** — they landed ~11,500 chars away, after the JSON output instructions,
> and one still read *"the word-weight rule above"*, a reference that no longer
> resolved anywhere near it. **My v4 guard passed because it asserted the strings
> were PRESENT, not that they were POSITIONED.** An assertion that checks presence
> and not position passes a misplacement silently. 241 v2 now asserts
> `position('{{.voice_style}}' in t) <= 500`.

241 v2 does both jobs — block → placeholder, and delete the orphaned tail. Applied:
`OK: literal replaced by placeholder, 11840 chars remain` (16,150 − 3,633 block −
695 tail + placeholder). **There is now exactly one copy of the house voice**:
canonical row 2,499 chars; page-content-writer literal **0**; Go const **0**.

**content-creator was the roll that was actually needed** — it was on v1.0.1174
without the reader, so every blog and social post was still being written with no
house voice. Built from HEAD at **v1.0.1178**, pushed, deployed via its own overlay
(there is no per-service push/deploy target; `push-backend` is fleet-wide, so
`docker push` + `kubectl apply -k` on that one overlay is the surgical route). It
was idle, 0 requests in 20 min, so nothing was interrupted.

Pod-grep, with the discriminating control: reader present (`voice_style_block` 2,
the SQL 1, the warning string 1) and **the block TEXT absent (0)** — that zero is
the proof the prose is no longer compiled into Go, and it is absent *because of* the
change rather than incidentally.

**End-to-end, live blog generation on gemini-pro-latest, 212 words:**

| check | result |
|---|---|
| `house voice block unavailable` warnings | **0** |
| em dashes · exclamations · negative-frame opens | **0 · 0 · 0** |
| joined clauses | **1** |
| mean words/sentence | 14.1 (SD 3.0) |

The joined clause is the owner's own request, in live blog copy: *"Technical leaders
often prioritize initial accuracy, but they'll get better results building error
recovery instead."* Contractions present. Not staccato.

> **AND THE SAME OUTPUT FABRICATED A STATISTIC.** *"Industry data shows that large
> language models experience hallucination rates between 3% and 10% depending on the
> task."* No source, invented range, stated as fact. The house voice governs how copy
> READS and says nothing about whether it is TRUE — and **content-creator has no
> claims gate at all**: the fabrication machinery (`043`, the evidence_base, the
> claims checkers) lives on the site/page path, which this agent does not touch.
> Applying the voice to content-creator is what surfaced it. **Needs an owner call
> before any blog output is published anywhere** — flagged, not fixed, and not mine
> to bundle into this change.

---

## 2026-07-27, ~20:00–20:15 UTC — P7 done: page-content-writer wrote a live page on Gemini

### First, a correction to the handoff I was following

> **CORRECTED 2026-07-27 evening — `HANDOFF_2026-07-27b` §1 is wrong about why the
> `grip-styles` build failed, and its suggested retry could never have worked.**
> The handoff says the 15:14 build failed because the writer could not reach Gemini
> (`bugs_open/112`), making it "stale evidence" now that 112 is live. It reads the
> failure from the timing rather than from the row. The row says:
>
> ```
> page-build-handler no-op: no sections ready to build (empty spec sections, or all
> sections deferred for missing data) — the target section was NOT rebuilt
> ```
>
> It never reached an LLM call of any provider. Caught by reading
> `site_work_items.error` before re-queueing. Cheap check that would have caught it
> at writing time: read `.error`, not `.updated_at`.

**Why it could not work, and why the named fallbacks could not either.**
`page-build-handler.load_spec_sections` reads `site_plan_sections` (authoritative),
falling back to `pages.sections`. For dartsonline:

```sql
SELECT sps.page_name, count(*)
FROM site_plan_sections sps
JOIN site_plans sp ON sp.id=sps.plan_id JOIN sites s ON s.id=sp.site_id
WHERE s.domain='dartsonline.com' AND sp.is_current GROUP BY 1;
-- 9 rows: about, brands-index, contact, guides-index, index,
--         new-arrivals, sale, shipping-returns, shop-index
```

`grip-styles` is absent, and **so is every `blog-post` page** — all 10 have
`sections = []` and no plan rows. The handoff's fallbacks (`tungsten-guide`,
`steel-tip-vs-soft-tip`, `beginners`) are in that set. Two of them already tried and
parked: `tungsten-guide` and `beginners` sit in `needs_human_review` from 2026-07-22
and 07-20, alongside `board-setup`, `barrel-weight` and `brand-detail` — five items,
same no-op, pre-dating this workstream. **This is not a Gemini problem and not new.**

### A name collision that nearly produced a false correction

`agent_definitions` has `type='content-creator'` (v1 + v2), whose `ai_service` is
`claude-haiku-4-5`/`anthropic` — zero occurrences of "gemini" in either config. That
is **not** the content-creator that was flipped. The flipped one is the standalone
k8s service, which builds its client from `cfg.Custom["ai_service"]` in a **configmap**
(`internal/agents/contentcreator/agent.go:101`), verified live:

```
kubectl -n ai-persona-system get configmap content-creator-agent-configmap \
  -o jsonpath='{.data.content-creator-agent\.yaml}' | grep -A4 '  ai_service:'
  provider: "gemini"   model: "gemini-pro-latest"   api_key_env_var: "GEMINI_API_KEY"
```

I was one step from filing "content-creator was silently reverted". **Two live objects
share the name; the `agent_definitions` row is the orchestration workflow, not the
service.** Check which one before asserting anything about "content-creator".

### The run: work item `df744e27`, orchestration `af2d066b`

Target chosen as the only never-deployed page with plan sections and no competing
open work item (`shop-index`/`brands-index` both had open items — coverage check).
Owner approved publishing. Dispatch needs `status='triaged'` + `approval_mode='auto'`
+ `pipeline='build'` (`load_work_item_actions.go:558`); `build-pipeline-trigger`
fires every 120s and picked it up in ~2 min.

**Gemini calls, from `llm_call_log` — the first ever recorded for this provider:**

| step | model | max_tokens | in | out | ok | ms |
|---|---|---|---|---|---|---|
| `process_sections_loop_iter_0_generate_content` | gemini-pro-latest | 8000 | 4227 | 87 | t | 9608 |
| `process_sections_loop_iter_1_generate_content` | gemini-pro-latest | 8000 | 4160 | 79 | t | 16476 |

Two notes on that table:

1. **`max_tokens` reads 8000, not 16192 — so `110` candidate 1 is LIVE**, not inert
   as `HANDOFF_2026-07-27b` and RUNBOOK §5 both state. The runbook's own test ("if
   you see 16192 that row predates the roll") is the thing that settles it. The
   chassis in production already carries it. **[VERIFIED]** by the rows above.
2. **87 and 79 output tokens is correct, not starvation.** These steps emit a small
   JSON content object (headline / subheadline / cta_text), not prose. `success=t`
   and no `error_message` naming thinking — which RUNBOOK §5 correctly identifies as
   the authoritative truncation signal for a thinking model.

**Before this run, `llm_call_log` had ZERO gemini rows in 7 days, fleet-wide.** So
every Gemini claim in the workstream up to now rested on direct API probes and on
content-creator (which does **not** write to `llm_call_log` — it is a standalone
service and never goes through `ai_actions.go`). This run is the first evidence the
provider works through the chassis orchestration path end to end.

### The artefact — read, not inferred

`section_plan`: `ready_count 2` (hero, call-to-action), `skipped_count 1`,
`deferred_count 0`. **`product-grid` skipped: `"on_missing=skip_section triggered"`** —
no product data resolved. Live at `https://dartsonline.com/sale.html`, 21,821 bytes,
header + footer present.

| check | result |
|---|---|
| em dashes · exclamations · filler | **0 · 0 · 0** |
| negative-frame openings | **0** |
| contractions | present ("It's") |
| "why it matters" sentence | **yes** — *"It's easier to test different weights and grip profiles when the gear costs less."* |
| **fabricated statistics** | **0** — no percentages, no deadlines, no invented urgency **on a sale page** |
| site's own story survived | yes — tungsten barrels, flights, shafts, grip profiles, grouping |

**Defect found, not Gemini's: the two blocks duplicate each other.** Hero =
*"Find Your Next Set on Clearance"*; CTA = *"Find your next setup in the clearance"*,
and both subheads cover discounted tungsten + finding weight/grip. Each section is
generated in its own loop iteration with no sight of its siblings' output, so nothing
in the chain can notice. Structural, provider-independent, unfiled.

**Worse defect, also not Gemini's: a Sale page with nothing to buy.** The product
grid was the whole commercial point and it was dropped for missing data, leaving two
blocks of copy asserting things are marked down. Each step behaved defensibly;
the page is still a shop page that cannot sell.

### Misstep: I called a live page dead

I probed `https://dartsonline.com/sale` → **404** and was about to attribute it to
`bugs_open/098` ("`deployed_at` set but not fetchable") or `120` (merge commit skips
deploy). What stopped me was checking **`new-arrivals`, a page I had not touched**,
deployed cleanly on 07-26 — also 404. A page I did not touch cannot have been broken
by my build, so the fault was in my probe. **This site serves `.html` extensions**:
`/sale.html` → 200, `/about.html` → 200, extensionless → 404 for all. The page had
been live and correct from 20:10:21.

Cheap check that caught it: **probe an untouched peer before believing a bug.** Logged
to `WRONG_CALLS.md`.

---

## 2026-07-27, ~22:45–23:05 UTC — `110` candidate 2: thinking cost is now a column

Owner asked for this directly after P7 ("yes please go ahead with 110"). Commit
`ca4071c82`; migration `245_llm_call_log_thinking_telemetry.sql` **applied and
verified live**; council corr `913a86e0-8847-4346-b688-5decfcb8e312`.

### What was actually missing

`gemini.go` computes four values and writes them into the options map. **No reader
outside `platform/aiservice/`** — they reached no column and no query. Thinking bills
as output and the writer runs once per section across the estate, so the one number
that drives the cost decision could not be selected. `016b` §9: *a field is only as
live as its LAST reader.*

### The bug file's own column list was stale, and checking it changed the design

`110` names `visible_budget_tokens` as one of the four to add. **It is already
`max_tokens`** — candidate 1 made `__sent_max_tokens` its sole feed so the column
means the caller's answer-budget for every provider. Adding it would give one meaning
two column names: **the defect 110 exists to close, reproduced a third time inside
the fix for it.** Added `wire_max_output_tokens` instead.

The bug file's key name for it is wrong too. Grepped, with the stale name as a
**negative control**:

```
__sent_wire_max_output_tokens      written_by_gemini=1 read_by_logger=1
__sent_thinking_reserve_tokens     written_by_gemini=1 read_by_logger=1
__usage_thinking_tokens            written_by_gemini=1 read_by_logger=1
__usage_total_tokens               written_by_gemini=1 read_by_logger=1
negative control __sent_visible_budget_tokens: 0
```

**Why the control mattered here specifically:** my test asserts these literals, and I
wrote both the test and the fixture. A test I author asserting a key I chose proves
nothing about `gemini.go` — that is the *"I tested a shape against a fixture I wrote"*
trap. The grep is the only thing in this change that actually pins the contract, and
there is no compiler check on it: a typo yields nil, which is **indistinguishable from
a provider that reports nothing**, so the column would sit quietly NULL forever.

### The one real design decision: NOT using the package's own helper

Every other int in `LLMCallLogParams` goes through `nullIfZero()`. Using it here would
map `0 → NULL` and collapse *"a thinking model that spent no thinking"* into *"this
provider has no thinking"* — **the empty-vs-absent confusion this bug is an instance
of**, and precisely what migration 243's header warns about one table over. So the four
params are `interface{}`, nil **only** when the key was absent. A reported 0 survives
as 0; anthropic/ollama rows are honestly NULL.

### Read on the failure path too — the valuable half

`gemini.go` sets the usage keys **before** it inspects `finishReason`, so a call cut
short by thinking still carries the count explaining why. A failed row with
`thinking_tokens` ≈ `wire_max_output_tokens` is the **starvation signature**, readable
without the `output_tokens` arithmetic that 110 §"Consequence 1" shows cannot express
truncation for a thinking model. 107 was misdiagnosed as an incapable model exactly
because this number was invisible. [VERIFIED by reading gemini.go:404–411 against the
`RefusalError`/`finishReason` checks that follow it.]

### Applying one migration without sweeping eight others

`run-migrations.sh --apply` applies **all** pending files, and the queue held eight
belonging to other threads (229, 230, 234, 235, 236, 240, 241, 242). There is no
single-file flag. `MIGRATIONS_DIR` **is** overridable, so:

```bash
SB=<scratch>/mig245; mkdir -p $SB
cp docs/agent_docs/sql_for_agents/245_*.sql $SB/
MIGRATIONS_DIR=$SB ./scripts/migration/run-migrations.sh          # dry run: 1 pending
MIGRATIONS_DIR=$SB ./scripts/migration/run-migrations.sh --apply  # applies + records
```

This keeps the runner's own safety machinery (doomed-transaction probe, ordering,
`schema_migrations` ledger) while touching only my file. **Added to the RUNBOOK.**

### Still owed

**The Go half is INERT until the next chassis roll, so `110` stays OPEN.** Until then
every new column is NULL on every row — honest, because the data genuinely was not
captured, and **no backfill is possible**: the values were never persisted anywhere to
backfill from. Post-roll check is item 3 of the bug's own §"How to verify": the four
columns non-NULL on a Gemini row, `thinking_tokens` in the 2,764–2,878 range measured
for the writer's real prompt.

### Council round 1 was REAPED, not decided — and the lane is slower than 030 claims

**Run 1, corr `913a86e0`**: submitted 2026-07-27 22:04 UTC, reached
`review_constitution`, then **wedged for over four hours** and was killed by the
watchdog:

```
error: reaper: stale EXECUTING_STEP for >4h; step=review_constitution
__step_error: (null)          -- nothing; the step never errored, it never finished
```

**No LLM call was ever made.** `llm_call_log` between 23:00 and 03:00 UTC holds six
rows, none from a council seat — so the seat wedged *before* reaching the model, which
is the spawn→call handshake shape ([[spawn-call-handshake-races]]), not a truncation
or an unparseable verdict. **`__step_error` being NULL is the tell**: a step that
failed leaves an error, a step that hung leaves nothing.

**It was not just me.** Two council runs were reaped in the same window — mine at
`review_constitution` (02:07) and another thread's at `review_guardian` (23:23).

**But the lane is NOT broken, and I checked before concluding it was.** Unfiltered
over the day: **23 COMPLETED vs 2 FAILED** (~8%), which is in line with the ~11%
harness-failure rate already recorded for council rounds. So this was bad luck, not a
new defect — and *not* something to file. Resubmitted the identical plan as
**corr `fa4ec9c8`**; nothing was revised, so it is a fresh run rather than a
`RESUBMIT_CORR` revision, which would have implied objections that do not exist.

**Dispatch latency, measured, and it contradicts a closed bug.** Run 2 was published
**02:40 UTC** and its orchestration row was created **06:56:48 UTC** — **4h16m**
publish→start. `bugs_open/030` is CLOSED as "publish→start 1s vs ~18min"
([[dispatch-queue-serialisation-workstream]]). Either that gain does not hold under
this load, or the two reaped runs held the lane until something cleared them.
**[UNMEASURED]** which — I have not traced the group's head-of-line state, and I am
recording the observation rather than the explanation. Anyone budgeting "~30 minutes"
for a council verdict from the CLAUDE.md figure should know a 4-hour wait is reachable.

### Independent corroboration of the Gemini writer, from a build I did not run

At **02:29 UTC**, `page-content-writer` ran three sections on `gemini-pro-latest` —
115, 51 and 98 output tokens, all `success=t` — for **`model-directory` on
ai-agent-orchestration.com**, deployed 02:31:36. That is the model-directory
workstream's own build, not mine, on a different site.

This is worth more than my own verification: it is the Gemini writer working in
production **for a thread that was not testing it**, which is the strongest form of
the untouched-peer check. Three independent page builds on Gemini have now completed.

**And all three rows show the four new columns as NULL** — migration 245 is live, the
Go that fills it is not. Exactly the documented pre-roll state, confirmed rather than
assumed:

```
 created_at | agent_type | provider | max_tokens | wire | reserve | thinking | total
 02:30:02   | page-content-writer | gemini | 8000 |  |  |  |
 02:29:43   | page-content-writer | gemini | 8000 |  |  |  |
 02:29:33   | page-content-writer | gemini | 8000 |  |  |  |
```

### v1.0.1180 is deployed and does NOT carry 110 candidate 2 — and two of my markers were vacuous

Chassis pod `agent-chassis-5987b8749b-4p4bl`, image **v1.0.1180**, started
**2026-07-27T22:06:22Z**, 0 restarts. My commit `ca4071c82` landed at **22:02:16Z** —
**four minutes** before the pod started, i.e. almost certainly after the image build
had already begun. `make build-*` builds from committed HEAD, so a commit that lands
mid-build is simply not in the artifact.

**My first two pod-grep markers were VACUOUS and would have told me the opposite:**

| marker | count | what it actually matched |
|---|---|---|
| `wire_max_output_tokens` | 1 | `__sent_wire_max_output_tokens` — gemini.go, pre-existing since 107 |
| `thinking_reserve_tokens` | 5 | `__sent_thinking_reserve_tokens` + the config key, all pre-existing |
| `total_output_tokens` | **0** | nothing — `__usage_total_tokens` does NOT contain this substring |

**I chose the column names to echo the option keys, which is good for reading the
code and useless for proving it shipped.** Three of my four column names are
substrings of strings 107 already put in the binary. Only `total_output_tokens`
discriminates, plus any multi-column fragment of the INSERT.

**The decisive check was a positive control in the SAME statement.** Absence alone
could mean the binary lacks the INSERT, the grep is wrong, or `strings` missed it:

```
rag_context_used                                 1   <- pre-existing, same INSERT
prompt_variant                                   1   <- pre-existing, same INSERT
work_item_id, vertical                           1   <- pre-existing, same INSERT
total_output_tokens                              0   <- mine
thinking_tokens, total_output_tokens             0   <- mine
wire_max_output_tokens, thinking_reserve_tokens  0   <- mine
```

The INSERT is in the binary; my columns are not. **Conclusive: 110 candidate 2 is
still INERT, and the bug stays OPEN.** It needs an image built from a commit at or
after `ca4071c82`.

> **This is why "a new chassis was deployed" is not an answer to "is my change
> live".** Two facts that both looked like yes — a bumped tag and a fresh pod — and
> the change was still absent. The pod-grep is the only thing that settled it, and it
> only settled it once the marker was one my change ALONE creates.
