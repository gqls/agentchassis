# NOTES — bugs_open/393 (append-only, newest at the bottom)

## 2026-08-26 (a) — lane opens; the perishable evidence extracted FIRST

The bug file warns the 11 `NO_CHANGE_GATE_UNREADABLE_RESULT` rows die on the retention clock, so
the first act was the extraction: `EVIDENCE_2026-08-26_unreadable_rows_extract.txt` (11 rows,
2026-08-14 14:24 → 08-17 12:44, all `dark_section_audit`, all from `build-dispatch-loop`).

One full row (391 chars) says more than the bug file could: the gate's declared counters are
`response.fix_result.total_fixed` / `response.text_color_result.total_fixed` — **color-variable-
fixer's two repair steps** — and the payload it read had top-level keys
`[agent_id agent_type role topics]`: **a SPAWN RECORD, not a handler result** (the
`bugs_closed/287` family). The row's own `context.remedy` names the fix path.

## 2026-08-26 (b) — [MEASURED] the whole population, live ∪ archive, and part A dissolves

Every `dark_section_audit` item ever (39 rows), by phase:

| phase | rows | result shapes seen |
|---|---|---|
| 08-11 → 08-17, handler `color-variable-fixer` | 30 | **four distinct foreign shapes**: `spacing,typography,color_scheme,design_notes` (a design-token blob); `approach,new_page,reasoning,…` (another page's triage decision); `role,topics,agent_id,agent_type` (spawn record, bugs_closed/287); plus real `response,…` envelopes |
| 08-19 → 08-23, handler `css-patch-agent` | 2 | one `needs_human_review` (held: `grading,held_by,held_reason`), one `failed` (full envelope). **ZERO completes.** |
| 08-25 21:37 → 23:40 (post-roll), **no handler** | 6 | all `deferred`, `created_by='visual-design-audit'`, summary prefixed `[verdict, not dispatched]`, spec carries `filing_mode` / `not_dispatchable` / `deferred_by` / `deferred_reason` / `release_recipe` / `routed_handler` / `routed_status` |

**So part A ("make the shape match") is not actionable and not currently a defect:**

1. The 11 unreadable events split exactly as the roster comment records: 7 spawn records —
   **fixed** by `bugs_closed/287` (rolled v1.0.1307, 08-17); the rest, color-variable-fixer's
   foreign shapes — **retired** when the owner moved routing to css-patch-agent on 08-19.
2. The roster entry is `LicenceVoided`, both carriers `enabled=false` — nothing reaches it.
3. The type's NEW traffic (last night's six) is **deliberately not dispatched**: verdicts filed
   `deferred` with no handler and a `release_recipe` in the spec. Nothing can complete, so
   nothing can complete *ungraded*.
4. The roster's own precondition ("re-measure css-patch-agent's reply shape on THIS type") cannot
   be satisfied: css-patch-agent has **zero** completions of this type to measure.

**What remains is exactly part B** — the commissioned reader — so the *next* type that drifts into
unreadable-and-completing is a finding the morning after it first appears, not 11 days later by a
census accident. Note the subtlety for B's design: the record fires on the **abstain** arm
(`recordUnknownNoChangeShape`, `complete_work_item_no_change.go:485-513`); a type at
`unreadableRefuses` blocks instead. So a NEW row of this code always means "a rostered
abstain-policy type completed ungraded" — precisely the population the reader exists for, with no
false positives from the refuse arm.

## 2026-08-26 (c) — a discovery en route, NOT this bug: dark_section_audit verdicts are being filed-and-held

The six born-`deferred` rows are a designed holding pattern (`[verdict, not dispatched]`,
`release_recipe`, `routed_handler` in the spec) — presumably the 08-19 ruling's follow-through or
last night's improvement-loop migrations (623/624/625). Not graded here as good or bad; recorded
because whoever eventually releases them re-opens the question this lane's part A dissolved, and
the roster comment's precondition (re-measure, rewrite or delete + exclusion row together) becomes
live again at that moment. **The release recipe's owner should read
`complete_work_item_no_change.go`'s dark_section_audit comment before releasing.**

## 2026-08-26 (d) — part B built, proven at the wire, committed; and BOTH directions of the same-file-passenger landmine in one morning

**The reader shipped**: `--ungraded-completions` mode + CronJob (07:35 UTC, slot proven free
against the repo census with my own manifest excluded), acks file in-image, registry flipped
`consumed`, DBG-077 registered. Wire proofs: live run 11/1/acked/exit-0 over 49,046 rows; novel
type → exit 1; zero aliveness → exit-2 refusal. Three mutation proofs, file restored
byte-identical. Council `0871db60`, verdict pending at write time.

**The passenger landmine fired in BOTH directions within one hour:**

1. **My WIP rode another lane's commit.** My registry flip (working tree, uncommitted) was swept
   into `a0ec90eb9` at 10:16 — their pathspec commit of `finding_code_registry.json` took the file
   as the tree held it. Consequence worth remembering: **from 10:16 until my commit ~40 minutes
   later, HEAD's registry named a reader file HEAD did not contain**, so
   `TestShippedRegistryIsSelfConsistent` was RED at HEAD through no commit of mine. A registry
   entry and its reader must land together — and on this tree, "together" is threatened not only
   by forgetting but by a THIRD PARTY committing the shared file between your edit and your commit.
   The defence is the standing one: commit the moment the pair is coherent, narrowly.
2. **Another lane's WIP rode mine, named.** The makefile carried an uncommitted
   `IMAGE_TAG v1.0.1340 → v1.0.1341` bump from a release session; my pathspec commit takes the
   whole file, so it rode along — named in the commit message rather than silent, per the estate's
   accepted handling.

**Remaining for 393**: the council verdict (`0871db60`), then the next fleet release builds
`ungraded-completions-check` and applies its CronJob. Bar for closing: the check RUNS on schedule
and writes its `doc_notes` row (`subject_key='ungraded-completions'`) — fixed AND live. First
scheduled run: the first 07:35 UTC after the release.
