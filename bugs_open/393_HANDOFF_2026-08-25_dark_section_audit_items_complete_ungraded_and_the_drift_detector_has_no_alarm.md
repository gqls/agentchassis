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

---

# APPENDED 2026-08-26 (`bugfix_393_ungraded_completions` lane) — A dissolved on measurement, B built and submitted

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_393_ungraded_completions/` (evidence extract,
NOTES with the queries).

## A — "fix the one known instance" has no instance left to fix

The perishable evidence was extracted FIRST (all 11 rows, full messages + context — the extract is
committed in the lane, so the retention clock no longer matters). Then the whole population was
measured, live ∪ archive (39 `dark_section_audit` items ever):

- **The 11 unreadable events are fully attributed and both causes are gone**: 7 were spawn records
  stored as the item's result (`bugs_closed/287`, fixed and rolled v1.0.1307 on 08-17); the rest
  were color-variable-fixer's foreign shapes (a design-token blob, a foreign triage decision),
  retired when the owner rerouted the type to css-patch-agent on 08-19.
- **Nothing can currently complete ungraded**: the gate's roster entry is `LicenceVoided` with both
  carriers `enabled=false` (its own comment carries the full story), and the type's new traffic —
  six rows filed the night of 08-25 — is deliberately undispatched (RFC_056 `filing_mode=record`:
  `deferred`, no handler, release recipe in the spec).
- **The §3A ask ("make one shape match the other") is unsatisfiable as stated**: css-patch-agent
  has ZERO completions of this type to measure a shape from, and the roster forbids writing
  CounterPaths from a guess. The obligation transfers to whoever releases the held rows — the
  roster comment's precondition (re-measure, then rewrite or delete entry + claim-timeout
  exclusion together), and the new acks file's `dark_section_audit` entry restates it so the
  releaser cannot miss it.

⚠ One correction to §2's query: the clean grouping key is `context->>'item_type'` — the writer
stamps it into context; the `error_message` regex form works but is the fragile spelling.

## B — the commissioned reader: BUILT, proven at the wire, submitted

`config-key-audit --ungraded-completions` + daily CronJob `ungraded-completions-check`
(07:35 UTC), the commit-sha-exposure ratchet shape: rows grouped by item_type, **any type not
acknowledged-with-diagnosis in `architecture_review/no_change_unreadable_acks.json` fails the
run**; the acks file ships in-image from committed HEAD so an unreviewed ack is unrepresentable;
`--report` writes ONE `doc_notes` row per run, clean runs included, body stating scope.

Acceptance, as this file stated it, measured:
- **synthetic novel item_type → RED**: exit 1 through the real emit path, remedy text naming
  `complete_work_item_no_change.go`; **acknowledged list → GREEN**: the live hand-run returns
  11 rows / 1 type / acknowledged / exit 0 over a 49,046-row `agent_error_log`;
  **mutation-proved both ways** (ack-always-true → the novel arm reds; ack-lookup-dropped → the
  acked arm and the hollow-ack CONTROL red) plus hollow-ack-accepted → red.
- **Registry follow-up done in the same commit**: `NO_CHANGE_GATE_UNREADABLE_RESULT` flipped to
  `consumed` with `reader`/`reader_sink`; `TestShippedRegistryIsSelfConsistent` (DBG-075) green.
- One extra guard the file did not ask for, because zero-is-healthy here inverts the family's
  vacuity logic: an `agent_error_log` that reads as EMPTY is a **blind read**, refused with exit 2
  — never a clean report.

Register: **DBG-077**. Council: `Council-Submitted: 0871db60-8f34-4075-9d64-58a94f52eaa5`.
**Not closable yet**: the CronJob rides the next fleet release (bar: fixed AND live — the check
must run on schedule and write its doc_notes row before this file moves to `bugs_closed/`).
