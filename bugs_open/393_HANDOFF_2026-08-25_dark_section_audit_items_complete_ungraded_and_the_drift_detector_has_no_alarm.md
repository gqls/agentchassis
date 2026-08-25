# 393 — every `dark_section_audit` item completes UNGRADED by the no-change gate, and the row that says so has no reader

**Filed** 2026-08-25 by the `bugs_open/358` lane, on the owner's ruling of 2026-08-25 (decision 4:
commission a reader). Every figure `[MEASURED 2026-08-25]`, queries inline. Related closed lineage:
`bugs_closed/302` (item types with no registered verifier pass no-op completions) and
`bugs_closed/213` (verifier/producer join) — this is the same territory one layer down: the gate
EXISTS here and cannot read what it is given.

## 1. The defect, in plain terms

The no-change gate (`complete_work_item_no_change.go:504`) grades a completing work item by reading
the handler's result. When it **cannot parse that result**, it lets the item complete **ungraded**
and writes a durable row saying so (`NO_CHANGE_GATE_UNREADABLE_RESULT`) — the right failure
direction. But nothing reads the row, so a handler whose result shape drifts away from what the
gate expects is **permanently exempt from grading, silently**, and the only evidence expires on the
retention clock.

## 2. Evidence — the entire population is ONE item type

```sql
SELECT count(*), string_agg(DISTINCT substring(error_message from 'item_type ''([^'']+)'''), ', ')
  FROM agent_error_log WHERE error_code='NO_CHANGE_GATE_UNREADABLE_RESULT';
```

`[MEASURED 2026-08-25]` → **11 rows, every one `dark_section_audit`** (window 2026-08-14→08-17).
This is not a diffuse "shapes drift sometimes" problem: **one handler's result shape does not match
what the gate can read, on every completion.** The gate's own message carries the shape it saw
(`error_message` suffix "— item completed ungraded by this gate: <shape>").

So this file is two asks in one, and the first is small:

## 3. What is asked for

**A. Fix the one known instance.** Read the `dark_section_audit` handler's result shape against
what `complete_work_item_no_change.go` expects (the row's `<shape>` suffix says what arrived), and
make one of them match the other. Until then, every `dark_section_audit` completion is ungraded.

**B. The commissioned reader (the class fix).** Something automated that selects
`NO_CHANGE_GATE_UNREADABLE_RESULT` rows and acts. The natural shape is the estate's daily-check
convention: a check that GROUPs the window's rows by item_type and **fails on any type not on an
acknowledged list** — i.e. the `shared_output_fields_ack` ratchet pattern: a known drift is
acknowledged with a reason; a NEW drifting type is a finding the morning after it first appears,
instead of never. (A `[MEASURED]` reference point: this instance sat unread for 11 days and was
found only because `bugs_open/358`'s census happened to sweep the table.)

**Acceptance**: A — zero new `NO_CHANGE_GATE_UNREADABLE_RESULT` rows for `dark_section_audit`
after the shape fix rolls, over a window in which `dark_section_audit` items demonstrably completed
(the demand control: completions > 0, ungraded = 0). B — a synthetic row with a novel item_type
turns the check red; the acknowledged list turns it green; **mutation-proved both ways**. Registry
follow-up: flip `NO_CHANGE_GATE_UNREADABLE_RESULT` to `consumed` with `reader`/`reader_sink` in the
same commit that ships B (`DBG-075` verifies both).

## 4. Traps

- **The gate is not broken** — it did the right thing loudly. Do not "fix" it to grade unreadable
  results; fix the shape, and give the loud row a reader.
- The 11 rows die on the retention clock; extract the `<shape>` suffix before building against it.
- A resolved row lives 14 days, not 30/365 — extract first if B resolves.
