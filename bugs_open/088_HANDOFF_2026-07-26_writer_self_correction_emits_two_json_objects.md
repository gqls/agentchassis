# 088 — the writer self-corrects out loud, emits TWO JSON objects, and the whole page build dies

**Filed:** 2026-07-26, by the bugs_open/073 verification thread, which found it blocking the same
homepage 073 was blocking.
**Severity:** High while it lasts — one occurrence takes out an entire page build (all 8 sections),
and the failure is at iteration 0 so nothing downstream runs. Frequency is the open question (see
*How often* below).
**Class:** contract — a COMPLETE, well-formed response is discarded because it is followed by more
text, and the discard is indistinguishable from a truncation.
**Status:** OPEN, not fixed. Cause fully evidenced and both fix candidates tested below.

**Not `bugs_open/076`** (truncated responses tolerated at unguarded call sites): nothing here is
truncated. The model returned a complete JSON object, then commentary, then a second complete JSON
object. **Not `bugs_open/073`** (an honest empty stat fails the render gate): the field is not
empty, it never arrives at all.

---

## What happens

A page build dies at the first section:

```
step process_sections_loop_iter_0_render_section failed:
failed to execute action render_component:
component "hero" is missing required content field(s) [headline] —
refusing to render an empty section (likely LLM truncation or an unparseable response)
```

Live run, 2026-07-26 14:26:27Z, correlation `d9fd6ed2-28e7-4af2-a49a-a749d71bccd3`,
site `2a8ebf9c-20a2-4c39-b191-840b012371da` (ai-agent-orchestration.com), `/index.html`.

The parenthetical is right about "unparseable" and wrong about the cause, which is why this went
unnoticed: the error blames truncation, and nobody re-read the payload.

## The payload, verbatim

```sql
SELECT collected_data->'generated_content_0'->>'result'
FROM orchestration_states
WHERE correlation_id='d9fd6ed2-28e7-4af2-a49a-a749d71bccd3'
  AND current_step='process_sections_loop_iter_0_render_section';
```

```
{
  "headline": "Multi-agent systems deployed to production in days, not months — on Kubernetes, Kafka, and Postgres",
  "subheadline": "POCs fail in production because orchestration is harder than the model. …",
  "cta_text": "Book a Technical Discovery Call",
  "secondary_cta": "See the Agent Registry"
}

Wait — I must scan for em dashes before returning. Found one in the headline. Rewriting now.

{
  "headline": "Multi-agent systems deployed to production in days, not months. Built on Kubernetes, Kafka, and Postgres.",
  "subheadline": "POCs fail in production because orchestration is harder than the model. …",
  "cta_text": "Book a Technical Discovery Call",
  "secondary_cta": "See the Agent Registry"
}
```

Two complete objects. The second is the model's corrected answer and is **the one we want** — the
first violates the house style rule it was given; the second obeys it. The wrapper stored around
this is `{"type":"text","result":"…"}` — the raw-prose envelope.

## Why — the chain, each link checked

1. **`ParseLLMJSON` rejects it** (`platform/orchestration/actions/json_envelope.go:57`).
   `json.Unmarshal` fails with **`invalid character 'W' after top-level value`** (measured, not
   quoted from memory: the payload was pulled from the DB and run through Go). The payload *does*
   start with `{`, so it passes the "structurally trying to be JSON" guard at :65 and reaches
   `repairJSONStringLiterals` — which repairs escaping, not trailing content, so the second
   `Unmarshal` fails too and the function returns an error.
2. **The caller falls through to the text path** and stores `{"type":"text","result":"<raw>"}`.
   That envelope exists for genuine prose and for truncation, and the file's own header says a
   truncated document "must NOT be silently salvaged". A doubled response is the opposite case —
   complete, twice — and nothing distinguishes them.
3. **`extractContentWithFallbacks` finds no map** (`v3_site_actions.go:4242`): `result` is a string,
   so every candidate path misses and it returns nil. `sectionContentData` stays nil and the render
   context never receives `headline`.
4. **The required-field gate refuses the section** (`missingRequiredLLMFields`, `json_envelope.go:204`
   via `RenderComponentAction`, `v3_site_actions.go:1722`) — correctly, on its own terms: the field
   really is absent. The step fails, the loop fails, the page build fails, and the work item burns an
   attempt.

The gate is not the defect here. The defect is that a complete answer was thrown away upstream of it.

## Where the second object comes from

`page-content-writer`'s section prompt (Voice & Style block, live) says:

> - No em dashes, anywhere, ever. **Before returning, scan your draft for the — character**; every
>   single one must be rewritten as two sentences or a plain trailing clause. A draft containing an
>   em dash is wrong even if it reads well.

The model did exactly that, out loud, after it had already emitted the JSON — and its own correction
line contains an em dash. The Output Format block at the end of the prompt says only:

> Return a JSON object with exactly these keys:

It never says *only* that object, and nothing forbids commentary around it. An instruction to
self-check "before returning", placed in a prompt whose output contract is unfenced, is an invitation
to emit the check.

`[INFERRED]` that the em-dash rule is the specific trigger — one occurrence, and the model names the
rule in its own words, which is as direct as this evidence gets, but it is still n=1.

## How often

In everything `orchestration_states` retains (back to 2026-07-13):

```sql
SELECT count(*) FROM orchestration_states o, LATERAL jsonb_each(o.collected_data) e(k,v)
WHERE k LIKE 'generated_content%' AND v->>'type'='text';                       -- 2 rows
-- both are the two keys of the SAME failed step, i.e. ONE occurrence
```

So: **one occurrence — and it is the only time the raw-text fallback fired at all in 13 days.** The
path built for truncation has, in the retained window, only ever been reached by a complete-but-
doubled response. That is a small sample and should not be read as "truncation never happens"; it is
enough to say the doubled case is not exotic.

## Fix candidates

**A — prompt-side, config-only, live immediately (recommended first).** Fence the output contract in
`page-content-writer`'s section prompt: after "Return a JSON object with exactly these keys", add
that the reply must contain that object and nothing else — no commentary, no corrections, no second
object; if a draft needs revising, revise it before writing the JSON. Optionally reword the em-dash
rule from "Before returning, scan your draft" to a silent check. Cheap, reversible, no image roll.
Does not repair a response that still arrives doubled.

**B — parser-side, durable, needs an image roll.** Teach `ParseLLMJSON` to take the **last complete
top-level JSON value**. **Do not implement the naive form** — it was tried and it breaks the
bug-026/bugs_closed/005 truncation tripwire. Measured, with the real payload
(`probe088.go`, this thread):

| input | naive "last complete value" | strict form |
|---|---|---|
| doubled, both objects complete | accepts the 2nd ✓ | accepts the 2nd ✓ |
| doubled, 2nd object truncated | **accepts the 1st ✗ — ships the superseded answer, truncation invisible** | rejects ✓ |
| single object, truncated | rejects ✓ | rejects ✓ |
| single object, complete | accepts ✓ | accepts ✓ |

The strict form is the naive one plus a single extra condition: after the last value that decoded
cleanly, **if the remaining text still contains a `{` or `[`, reject the whole payload** — something
tried to be a JSON value and did not finish, which is exactly the truncation signature. Both forms
and all four controls are in
`probe088.go` (scratchpad; reproduce by pulling the payload with the query above).

Whichever is taken, the error message at `v3_site_actions.go:1731` should stop asserting
"likely LLM truncation" as the only cause — that phrasing is what made this look like 076.

## How to verify a fix

1. Re-run a full `/index.html` build for ai-agent-orchestration.com and watch it pass
   `process_sections_loop_iter_0_render_section`.
2. **Induce the fault** rather than trusting a green run: hand the parser the stored payload above
   (it is 976 bytes, kept verbatim in this file) and assert the recovered `headline` is the SECOND
   one — the one without the em dash. A fix that recovers the first is a regression dressed as a fix.
3. Assert the truncation control still fails: the same payload with its last ~120 bytes removed must
   still be rejected.

## Related

- `bugs_open/076` — truncated responses tolerated at unguarded call sites. Different mechanism; this
  one is not truncated. They share the raw-text envelope as the place where evidence goes to die.
- `bugs_open/073` (the honest-empty stat) — same gate, same page, different cause. 073's fix
  (migration 217) is live; this is what still blocks that homepage.
- `bugs_closed/005` / bug 026 — why the gate exists at all, and the property fix candidate B must not
  break.
