# HANDOFF — 2026-08-18b, fresh chat starts here: the prune is BUILT and under council review; 306+307 are FILED; the gating blast radius is MEASURED; and Phase 2 is **RULED — sequence A (RFC_029 §10.13)**. The next builds are steps 2 and 3 of that path.

**Supersedes `HANDOFF_2026-08-18_continue_here.md`, whose §2 task is DONE and whose §10.10 spec
was CORRECTED before building** — the mis-citation and its consequences are RFC_029 §10.11, the
population split is §10.12, and the residual finding is `bugs_open/306`. Do not build anything
from the 08-18 handoff's §2/§3; this file replaces them.

**Read in this order:** this file → `bugs_open/306` + `bugs_open/307` (the two filings) →
RFC_029 §10.11–§10.12 (corrections + measurement record) → NOTES `## 2026-08-18 (evening)`.

## 1. What is true now (all measured 2026-08-18 evening; evidence cited at each)

- **Build `v1.0.1309` = `f0117fb8b`**, label + digest matched (local RepoDigest == running
  imageID). Only orchestration change vs 1308 is coordinator.go retry timeouts (029 lane) — the
  window reads continuously across the roll.
- **The prune is COMMITTED (`131e6430e`) and the council verdict is APPROVED** (round 1, corr
  `ae0dfb93-9d…`, 4 advisory objections, none high-severity — every one dispositioned, NOTES
  "(verdict)" entry: editquality's ambiguity edge is now a seventh TEST (`595e3fca6`,
  `Council-Reviewed:` trailer), guardian's wider-suite ask ran (9 packages ok from the archive),
  bug_historian's observability point is recorded for step 5's design, nothing alarms on the
  row drop). It removes the ~28% work_item_id class only; the code comment states why it CANNOT
  touch the ~72% `current_page` class. Seven tests, mutation-proved. **It has NOT shipped**:
  needs the next chassis build+roll — that roll is §3 step 1's remaining precondition.
- **`bugs_open/306`** (tie-break undeclared; 13/139 genuinely-different-page candidate sets;
  `tryUnwrapMapPatterns` pattern 1 still ranges an unsorted map, 0/139 can fire). Fix candidates
  ranked in the file; candidates 1+2 (declare the tie-break + sort pattern 1) are small,
  behaviour-identical-today, council-gate material — buildable next.
- **`bugs_open/307`** (outage terminally killed items): owner ruling recorded — "a transient
  blip should return the item to queued". 88/100 exhausted 3 attempts INSIDE the burst (no
  backoff); 12/100 died at attempt_count=1 via an unidentified path (`[UNVERIFIED]` — first
  task in that file). The count-free-retry-on-404 trap is named. **The 100 casualties are
  CANCELLED (owner directive)** with audit notes in `result` — do not re-count them.
- **Migration 417 / 287 / Phase 1 promises: all discharged** (see the 08-18 handoff §1, still
  true; `field=result` silent since 08-17 16:29Z).

## 2. The decision-3 answer (measured — the input to the owner's Phase 2 ruling)

Gating `ensureCoreFields` splits cleanly in two:

- **NEVER gate `domain` / `objective` / `model`.** Undeclared prompt consumers fleet-wide:
  domain **39 steps / 31 agent types**, objective 10, model 6 (regex sweep over
  prompt/prompt_template/system_prompt on all live definitions; query in NOTES).
- **`current_page` / `current_section` / `render_context` are gateable.** Zero undeclared
  prompt consumers. Go-side: one live injection-dependent consumer —
  **`html-developer-chunked`** (3 steps, via `extractPageInfo` in generate_html), soft failure
  (prompt loses its "Page:" line), remediable FIRST with a config edit (add `current_page` to
  its three `input_fields` lists — live immediately, no build). All other Go readers either
  declare the field in their own list (`save_page_sections`, `html-developer`) or read
  `CollectedData` directly (unaffected).
- Gating removes the **63% class-1 rows** (build-dispatch-loop searching for pages nobody
  asked about). It is a guarantee change on a shared mechanism → architecture-scope: an
  RFC_029 amendment with the consumer enumeration above as its evidence. NOT yet ruled — it is
  option 2 of the re-presented Phase 2 sequencing awaiting the owner.

## 3. THE RULED PATH — owner chose sequence A, 2026-08-18 evening (RFC_029 §10.13 is the durable record; this is the working copy)

Five steps, each gated on a STATE the window must show, never a date. Steps 2 and 3 are this
lane's next builds and are independent of each other and of step 1's roll.

| step | what | precondition | done when |
|---|---|---|---|
| 1 | Prune ships; window re-read | council APPROVED on `ae0dfb93` + next fleet roll (owner's) | `work_item_id` class gone; `current_page` class INTACT (that is success — see the code comment) |
| 2 | `bugs_open/306` candidates 1+2: pin the tie-break (depth, then source-class), sort `tryUnwrapMapPatterns` pattern 1's keys | none — buildable NOW; behaviour-identical today, council-gated | tests pin both; mutation-proved |
| 3 | Gate `current_page`/`current_section`/`render_context` in `ensureCoreFields` (requested-only). **`domain`/`objective`/`model` stay UNCONDITIONAL — 39/10/6 undeclared consumers, never gate.** | **config edit FIRST**: add `current_page` to `html-developer-chunked`'s 3 `input_fields` lists (live immediately). Go gate is an RFC_029 AMENDMENT — name and tell consumers per the 2026-07-29 ruling; the NOTES enumeration is the evidence | class-1 63% gone from the window after the roll |
| 4 | Fix surviving shape conflicts AT SOURCE (worked example: `save_page_sections` sees the page OBJECT at `input_data.current_page` vs its NAME STRING at `render_context.current_page` — one rename ends it). **`!` cannot fix these** — Go-side request lists, config-side marker | step 3's window read | window reads zero, or every survivor carries a written safe-by-inspection note |
| 5 | Flip conflicts → refusal at the marked flip sites (`unified_extractor_search_test.go` header) | step 4's gate | §9's "never guess" is mechanical for every future pipeline — the reason the flip happens even on a near-empty population |

Carried correction: my §10.12 claim that `page-build-handler` "reads" the resolved value is
downgraded to `[INFERRED]` — no named reader of the injected copy was found in the decision-3
sweep (NOTES, evening entry, last paragraph).

## 4. Traps for a fresh session (beyond the standing ones in the 08-18 handoff §4, all still live)

- **097's `plan` is an OBJECT** (`summary`/`edits`/`grounded_in`/`risks`), not an array. A
  schema refusal is CLIENT-side (no round spent, safe to fix and re-run); only a published run
  must never be re-triggered.
- **A mutation that breaks the BUILD proves nothing** — an unused-variable compile error is not
  a failing test. Mutate the body to a no-op (`if true { return fields }`) so the package still
  compiles and only the guarded behaviour changes.
- The conflict instrument stores candidate PATHS, never VALUES — "are these the same page?" is
  answered ONLY at `orchestration_states.collected_data` (RUNBOOK four-step method).
- The prompt sweep's three columns (prompt/prompt_template/system_prompt) are the known
  template carriers; a fourth carrier would be a gap in the decision-3 answer. Named residual,
  not closed.

## 5. Session-start checklist

1. `git log --oneline -10`; re-read this file from disk.
2. Verdict already APPROVED and dispositioned (§1) — nothing owed on it.
3. If APPROVED and a build is rolling: verify by label+digest, then re-read the window
   (RUNBOOK) — expect the work_item_id class gone, current_page class INTACT (that is correct,
   not a failed fix; the comment in action_inputs.go says why).
4. Build the ruled path (§3): step 2 (306 candidates 1+2) is buildable immediately; step 3
   starts with the html-developer-chunked config edit, then the gate as an RFC_029 amendment.
