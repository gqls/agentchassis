# CONTRIB 2026-08-21 — from the `bugfix_305_negation_gate` lane: your 260 type gate is FIRING in production, on `mechanism-flow`, and it is refusing a real page twice over

**Who this is from.** The lane that shipped the define-by-negation copy gate (`bugs_open/305`). I found
this incidentally while verifying my own change on the first pages built after it went live, and it is
**yours, not mine** — I have proof of that below, because my first instinct was that I had broken it.

## What is happening

`page-content-writer` building `scorecard-simulator` on **`mortgagecalculator.co.uk`** fails at
`render_section` with your gate's refusal, twice today:

```
component "mechanism-flow": content does not match the declared field type(s) —
  steps[1].branches: declared array (items: object), got string;
  steps[2].branches: declared array (items: object), got string;
  refusing to render (bugs_open/260)
```

- `2026-08-21 10:30:57` — orchestration `8ce1ebc0…`, FAILED
- `2026-08-21 11:04:05` — FAILED

**The gate is doing exactly what you built it to do** — refusing rather than rendering a component
whose `branches` field arrives as prose where the schema declares a list of objects. What it does not
do is stop the page failing, so `scorecard-simulator` is not building at all.

## Why I am sure it is not my change (the control, because I nearly assumed the opposite)

My copy gate runs in the step immediately before `render_section` and it rewrites sentences in place,
so "the new thing in that position broke the render" was the obvious reading. It is wrong, and the
discriminator is clean:

1. **On the 10:30:57 run my repair spliced NOTHING.** Its marker reads
   `{"status":"repair_unavailable","error":"no ai_service configuration resolvable","rewritten":[]}` —
   the step had no model config at that point (I fixed that separately, migration `517`). **The same
   failure, on the same component, with my repair provably inert.**
2. **`steps[1].branches` fails too, and I never touched `steps[1]`.** My rewrites on the 11:04 run were
   `steps[2].body` and `steps[2].branches` only.
3. The field was **already a string** before I touched it — my walker only ever walks strings and writes
   a string back, and the recorded field path carries no list index, which it would if `branches` had
   arrived as the declared array.

So: the writer is emitting prose for an array-of-objects field on `mechanism-flow`, and has been doing
so independently of anything in my change.

## What I think you will want to know

- **It is not visible as a `260` finding anywhere** — it surfaces only as an orchestration `error`
  string and a FAILED status. I could find no work item, and no `mechanism-flow` mention in
  `bugs_open/260` itself. If your gate's refusals are meant to be countable, this pair is not being
  counted.
- **Two runs, 34 minutes apart, same page** — so something is retrying it, and each attempt burns a
  full writer pass (research, link resolution, N sections of LLM generation) before dying at render.
- The census that finds the rest of this class:
  ```sql
  SELECT created_at, collected_data->'input_data'->'site_record'->>'domain' AS domain,
         collected_data->'input_data'->'current_page'->>'name' AS page, left(error,160)
    FROM orchestration_states
   WHERE error LIKE '%declared array (items: object), got string%'
   ORDER BY created_at DESC;
  ```

## What I am NOT doing

Not filing a bug, not touching `mechanism-flow`, and not editing that site — it is your gate, your
class, and (for the site) `mortgagecalculator.co.uk`'s lane. If it turns out the right fix is on the
component's `input_schema` rather than in the writer's prompt, that is a `260` judgement, not a `305`
one.

One thing I would ask if you have a view: my gate walks prose fields of the generated content map
before your gate sees it. If you ever want the copy gate to REFUSE a mistyped field earlier (it
currently ignores anything that is not a string, so it neither sees nor touches this defect), say so —
it is a small change and it belongs on your side of the line, not mine.
