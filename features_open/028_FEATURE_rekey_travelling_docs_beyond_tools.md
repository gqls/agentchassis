# 028 FEATURE — `RekeyTravellingDocs` covers only tools, so every other subject type orphans its docs on rename

**Raised:** 2026-07-30 by the `staged_component_build` lane, at the council gate's
instruction. Two seats asked for this as a **tracked ticket rather than a policy comment**
when reviewing `subject_type='component'` (council `e5673868-7c5b-489c-931a-7ba59b959b91`):

- **`bug_historian`, medium, gating-adjacent:** *"Before this PR, component renames were
  not a doc-subject concern at all; after it, they are, and the only defence is a comment
  nobody is forced to read. This is fixing the vocabulary (symptom) while leaving the
  shared rename/regeneration mechanism generic and exploitable at the exact new call site
  this PR creates."*
- **`architecture`, low (with an approve verdict):** *"Subject-key immutability for
  components is enforced by a code comment only… Acceptable for this scope but should be
  tracked as a follow-up, not silently absorbed."*

**Status:** FILED, unowned, not designed. This entry exists so the deferral is visible
instead of living in a Go comment.

## The mechanism, measured

`RekeyTravellingDocs` (`platform/orchestration/datahelpers/travelling_docs_rekey.go:29`)
moves a subject's `doc_plans` + `doc_notes` rows from an old key to a new one. It takes
`subjectType` as a parameter, so it is *capable* of handling any subject type.

**It has exactly one caller, and that caller hardcodes `"tool"`** — verified by
`git grep -n "RekeyTravellingDocs" -- '*.go'`, 2026-07-30:

```
platform/orchestration/actions/rename_tool_identity_action.go:115:
  plansMoved, notesMoved, err := datahelpers.RekeyTravellingDocs(ctx, tx, "tool", oldFn, newFn)
```

So for **every other** subject type — `pipeline`, `experience`, `action`,
`experience-pattern`, and now `component` — a rename of the underlying thing leaves its
travelling docs stranded under the old key. **No error, no warning, no failed work item.**
The docs simply stop being found, and `load_doc_context` returns nothing, which reads
exactly like "this subject has no PLAN yet".

## Why it is not theoretical

`bugs_open/136` is a live, **half-landed** `*_domain` → `*_pipeline` rename: three
instances, nothing setting the new key, correct by luck. That is precisely this shape, on
a subject type that already exists. The population is not "tools that might one day be
renamed" — it is "renames that are already in flight".

For components specifically the key is `content_components.function`, and
`create_work_item:118-121` already carries a shim for exactly this class of rename, which
is evidence that component-function renames are a real operational event rather than a
hypothetical one.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Make orphaning impossible rather than detectable.** Route every rename through one
   helper that rekeys travelling docs as part of the same transaction — i.e. give the
   other subject types the equivalent of `rename_tool_identity`'s call, or invert it so
   the rename path *cannot* commit without the rekey. Closes the door; costs a survey of
   every rename path per subject type.
2. **A cheap detector, worth building regardless of (1).** A discovery check for
   travelling docs whose `subject_key` resolves to nothing —
   `doc_plans WHERE subject_type='component' AND subject_key NOT IN (SELECT function FROM
   content_components WHERE is_active)`, and the equivalent per type. This is the only
   candidate that also finds the orphans that **already exist** from past renames, which
   nobody has ever counted. Start here if only one gets built.
3. **Enforce immutability in the write paths** (the register's own rule for
   `experience_patterns`: names immutable once approved, supersede under a new name).
   Weakest, because it constrains authors rather than fixing the mechanism, and it cannot
   help the renames already in flight.

**(2) first, then (1).** (3) is what the `component` submission chose as an interim, and
both objecting seats were right that it is not sufficient on its own.

## How to verify a fix

Rename a component's `function` in a scratch site, and assert its PLAN and NOTES follow.
Then run candidate (2)'s query across all six subject types and record the orphan count
per type **as a baseline** — that number has never been measured, and it is the honest
size of the problem.

## Cross-links

`features_open/027` (the lane that surfaced this), concept register **DOC-068** (records
this as its open review question), `bugs_open/136` (the live rename),
`platform/orchestration/actions/doc_subjects_common.go` (where the interim policy lives),
council correlation `e5673868-7c5b-489c-931a-7ba59b959b91`.
