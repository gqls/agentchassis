# NOTES — council-gate cost

Running technical log. Append-only, newest at the bottom. Missteps are the
point, not an appendix.

## 2026-08-10 — how this started, and the first thing I got wrong

The owner hit an Anthropic account spend cap and said: *"please switch off the
improvement loop for now as it is eating my credits."*

**I did what was asked and it turned out to be aimed at the wrong thing.** That
matters more than the fix, so it goes first.

Disabled `site-discovery-rotation-{completeness,design,quality}` (three
`enabled` booleans, reversible, `improvement-sweep` was already off since
2026-05-02). Then measured what the improvement loop actually costs, and it was
**already inert and nearly free**:

- its two big auditors (`content-quality-auditor`, `visual-design-auditor`)
  last called an LLM at **2026-08-09 14:55** — over 24h earlier;
- the discovery agents made **zero** LLM calls in 24h;
- **zero** work items fleet-wide sat in `triaged`/`approved`, the only statuses
  the dispatcher claims — so 544 `detected` discovery findings were piling up
  with nothing able to consume them.

The real spender, measured the same way:

```
agent_type          | calls | in_tok     | out_tok
council-gate        |   119 | 11,632,762 | 370,298     <-- ~85% of fleet spend
page-content-writer |    89 |    817,617 |  96,847
experience-planner  |    19 |    163,768 |  56,213
```

**Input is 97% of the volume.** So `max_tokens`, output caps and (mostly) model
choice are all the wrong lever. Recorded in the register at IMP-016 with an
explicit "do not cite this as a cost saving; it saved approximately nothing."

### MISSTEP 1 — an unparenthesised OR silently dropped my time filter

Writing the by-agent breakdown I wrote:

```sql
WHERE created_at > now() - interval '24 hours'
  AND agent_type LIKE '%discovery%' OR agent_type LIKE '%audit%'
```

`AND` binds tighter than `OR`, so the second branch carried **no time filter at
all**. It returned `content-quality-auditor 7167 calls` and
`visual-design-auditor 4056` — all-time figures wearing a 24h label, and they
were about to become the headline of my answer ("the auditors are what's
burning your credits"). They were 24h+ stale.

Caught because the numbers were implausibly large next to the totals. The check
that settles it in one command is a **positive control**: run the same query
with no time filter and compare. If the two agree, the filter isn't filtering.

> **The general shape:** a filter that silently doesn't apply produces a number
> that is *real*, just answering a different question. It cannot look wrong,
> because nothing about it is malformed.

## 2026-08-10 (later) — why council-gate is expensive, measured rather than guessed

17 seats, each `claude-sonnet-5`, each with its own `ai_service` block in
`agent_definitions` (so models ARE individually tunable, live, no deploy —
my first query looked at `config.model` and found nothing; the real path is
`config.ai_service.model`).

Per-seat input is uniformly **~100k tokens**:

```
review_prior_art     17 calls  1,723,607 in   -> ~101,388/call
review_reuse_agent   17 calls  1,718,405 in   -> ~101,082/call
review_constitution  17 calls  1,707,270 in   -> ~100,428/call
```

That uniformity is the clue. Chasing it:

1. **The seats share their whole evidence body.** Took two seats from the same
   `orchestration_id` and sampled 400-char slices at 10,20,…,90% of one; all
   nine were found verbatim in the other via `position()`.
2. **Their common prefix is ZERO characters.** `left(a,n) = left(b,n)` fails at
   n=500. Common *suffix* is 0 too.
3. **Why**: every seat template is
   `[persona 4k–32k][shared block 172 chars of placeholders][more seat text]`.
   The seat persona sits at the top and pushes the shared body to a different
   offset per seat. Anthropic caching is a **prefix** match, so nothing could
   ever have cached regardless of client configuration.
4. The 27,200-char length gap between the two sampled seats is accounted for
   **exactly** by their template sizes (31,854 vs 4,654) — confirming the
   interpolated content is identical and only the template differs.
5. All 17 interpolate the **same three placeholders**
   (`.schema_hint.text`, `.input_data.rationale`, `.plan_persisted.plan_json`),
   and the 172-char block containing them is **byte-identical across all 17**
   (md5 `574d945d97706890d6595a0f24c9a38f`), appearing exactly once in each.

**And the fact that makes it fixable at all: the seats run SEQUENTIALLY**, not
in parallel — `review_editquality → review_constitution → review_mission →
review_prior_art → …gates… → review_guardian → council_decide` via `next_step`.
The documented fan-out hazard (N parallel requests all miss, because none can
read a cache the others are still writing) does not apply to a chain. Seat 1
writes; seats 2–17 read.

Also confirmed: **prompt caching had never been implemented anywhere in the
estate** — repo-wide grep for `cache_control`/`CacheControl` across `platform/`,
`internal/`, `pkg/` returned **0 files**, and the response parser discarded the
cache counters entirely.

### The stray boilerplate — investigated, then deliberately NOT changed

Every council prompt ends with a JSON-format instruction whose worked example is
a *site-classification* payload (`{"site_type": "brochure", …}`). It looked like
template bleed. It is `getJSONOutputInstructions()` in `ai_actions.go:1220`,
appended to **any** step with `output_format: json` fleet-wide.

I initially pitched this to the owner as a cost saving. **That was wrong and I
corrected it before acting**: it is ~120 tokens × ~340 calls/day ≈ **$0.08/day**.
The owner's ruling — *"don't remove the boilerplate if it's useful"* — is right,
and it is useful: it is what stops the model wrapping JSON in markdown fences,
which would break parsing fleet-wide. Left exactly as is. The example mismatch
remains a minor prompt-quality wart on a shared function, not a cost item.

### MISSTEP 2 — `git checkout` to revert a mutation test wiped the real change

To prove the new tests actually bite, I mutated the implementation and re-ran
them (both mutations failed the right test with the right message — good). To
revert the mutation I ran `git checkout platform/aiservice/anthropic.go` — which
restored the **committed** file and therefore deleted my actual implementation
along with the mutation. The test file survived only because it was untracked.

Cost: a few minutes re-applying two edits I still had in context. Could have
been much worse if the change had been larger or unrecorded.

> **The check:** to undo a deliberate mutation, undo the *mutation* (re-edit the
> line back, or work on a copy). `git checkout <file>` reverts to HEAD, and on
> uncommitted work HEAD is not where you were. Mutation-test on a file you have
> already committed, or keep the mutation in a scratch copy.

## What shipped

- **LCO-008 / `CacheBreakpointMarker`** — opt-in cache breakpoint in
  `platform/aiservice/anthropic.go`. Absent the marker the request body is
  byte-identical to before, so the seam's ~40 other callers are structurally
  unaffected (mutation-proven test). 1h TTL, chosen because the 17-seat chain
  runs 2–5 min and sits right on the 5-minute boundary; a partial mid-chain miss
  is the hardest kind to notice because it still works.
- **Migration 376** — `cache_creation_input_tokens` / `cache_read_input_tokens`
  on `llm_call_log`. Not bookkeeping: once a caller opts in, `input_tokens`
  means the **uncached remainder**, so every existing cost query would
  understate by ~95% *in the flattering direction*. NULL (old binary) is kept
  distinct from 0 (no cache used).
- **Migration 377** — hoists the shared block to the front of all 17 seat
  templates and inserts the marker. Verified before applying: block appears
  exactly once per seat, net length change is +25 chars (marker + 2 newlines,
  i.e. arithmetic proof it is a *move*), all three placeholders survive, and —
  the one that matters — **exactly 1 distinct cacheable prefix across 17 seats**.

Council submission: `b54f173e-ebd4-45c4-954a-dfc70005e62c`, committed with
`Council-Submitted:` since the verdict had not landed. Deliberately submitted
the client seam **alone**, not bundled with the template reorder: the guardian
seat has previously and rightly vetoed shared mechanisms that arrived buried
inside a larger change.

### A migration guard is only evidence if it can fail

Both 376 and 377 verify with `DO`/`RAISE`, not a block of `SELECT`s —
`ON_ERROR_STOP` does not treat a result set as an error, so a SELECT-based
"verification" commits green whatever it finds. For 376 I **induced** the guard
first (asked it to expect 3 columns; it aborted with exit 3) before trusting the
real run.

## Open / next

- Verdict on `b54f173e…` — read it, act on REVISE/REJECTED (code is already on
  the shared branch; that is by design here, not an oversight — see the
  2026-07-29 owner ruling that review is after the fact).
- 377 not yet applied: **a council run was in flight** through the very agent it
  rewrites. Applying mid-chain would assemble one verdict from two prompt
  generations with nothing downstream showing it.
- Build + roll the chassis, then prove at the artefact: seat 1
  `cache_creation > 0`, seats 2..17 `cache_read > 0` and `input_tokens`
  collapsing ~100k → ~5k. **A zero `cache_read` is the failure mode**, and it
  looks exactly like success.
- Not done, offered and not taken up: downgrading mechanical seats to Haiku 4.5.
  Worth revisiting once caching is proven, since the two compose (caches are
  model-scoped, so a mixed roster just means separate cache entries per model).

## 2026-08-10 (evening) — council APPROVED, and the two objections that were real

Verdict on `b54f173e-ebd4-45c4-954a-dfc70005e62c`: **approved, 5 advisory
objections, none high-severity**. Two were real defects and were fixed before
the code shipped. Both were invisible to the tests as first written, which is
the part worth inheriting.

1. **`ttl:"1h"` needed a beta header this client does not send** (edit-quality,
   medium). The first caller to opt in would have got a **400, not a cache
   hit** — and because council-gate reviews every platform change, that 400
   removes the review path for the whole estate, not just the saving. Removed
   the `ttl` field entirely; the 5-minute default needs no header on any model.
   **The asymmetry is the argument**: a too-short TTL is a worse saving, visible
   instantly in `cache_read_input_tokens`; an unsupported `ttl` is an outage.
2. **A marker at position 0 leaked the literal marker text to the model**
   (edit-quality, low) — a real bug I introduced. **My test passed the whole
   time**, because it asserted the *type* of the content (still a string, true)
   and never its *content*. A type-only assertion sails straight through this
   entire class of defect. Now stripped, asserted, and mutation-proven against
   the exact bug the seat described.

Also answered `bug_historian`'s objection by measuring instead of deferring:
only two `scheduled_tasks` pre_queries reference `llm_call_log`, and **both read
`output_tokens` only** — so the `input_tokens` meaning-change affects zero live
automated consumers. (`total_tokens` would be affected, per migration 246, and
nothing reads it.) That audit is the sort of thing CLAUDE.md says to do *before*
submitting rather than leave for a reviewer; the council was right to ask.

### Applying 377 mid-flight was safe, and the check that proved it generalises

The migration originally carried a "do not apply while a council run is in
flight" banner. **That was a guess, and it was wrong.** One query settled it:

```sql
SELECT orchestration_id,
       (workflow_plan->'steps'->'review_mission'->'config' ? 'prompt_template'),
       length(workflow_plan->'steps'->'review_mission'->'config'->>'prompt_template')
FROM orchestration_states WHERE status NOT IN ('COMPLETED','FAILED','CANCELLED');
```

Both live runs carried their **own copy** of every seat template inside
`orchestration_states.workflow_plan`. An orchestration executes from that
captured plan, so an `agent_definitions` edit cannot reach a run already under
way. Confirmed empirically afterwards: council calls at 19:33–19:38, after 377
landed, still start with `# Council reviewer: …` (the OLD ordering) and carry no
marker.

On a tree this many sessions share, that is the difference between "config edits
are safe" and "wait for a quiet window that may never come". **What it does not
license:** editing a prompt and expecting an in-flight run to pick it up — it
will not.

### The intermediate state, stated precisely (I was sloppy about this once)

I wrote in a commit message that on the current binary the marker is
"stripped-or-ignored". **That is not accurate.** Until the image rolls, a council
run started *after* migration 377 will snapshot the new templates and the old
binary will send `<!--CACHE_BREAKPOINT-->` to the model as **literal text**. It
is an HTML comment sitting between the plan and the seat persona, so it is
semantically inert and no reviewer has commented on one — but it is genuinely
present, not removed. Once v1.0.1282 is live it becomes a boundary and never
reaches the model.

`cache_creation_input_tokens` / `cache_read_input_tokens` read **NULL** on every
call until the roll, which is the designed signal for "this binary predates
cache support" — already useful for reading fleet state, and the reason those
columns were deliberately not defaulted to 0.

### Where this stops being mine

Deploys here are **whole-fleet, one tag, run by the owner**
(`make release redeploy-agents ENVIRONMENT=production REGION=uk001`). A
single-service roll at its own tag fragments the fleet. `IMAGE_TAG` is bumped to
**v1.0.1282** and committed; production runs v1.0.1280. Verification after the
roll is mine and is written up in the RUNBOOK — the one that matters is a
**non-zero `cache_read_input_tokens` on the 2nd+ seat of a run**, because a zero
there is the failure mode and looks exactly like success.
