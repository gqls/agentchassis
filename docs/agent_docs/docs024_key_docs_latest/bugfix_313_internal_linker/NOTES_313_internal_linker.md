# NOTES — bugfix 313/298 (internal linker dead branch + cap)

Append-only, newest at the bottom. Missteps are the point — record them.

## 2026-08-19 — session start, re-verification, design

- **Ownership checked before anything** (`who-owns.py 313` / `298`): both filed by
  `bugfix_275_silent_row_caps`, whose handoff carries a "LANE CLOSED 2026-08-19 10:30Z" banner and
  lists 313→298 as standing tickets **nobody owns**. Diagnosis queue
  (`item_type='needs_diagnosis' AND status='awaiting_diagnosis'`): empty. No competing thread.
- **Bug still valid, re-verified live 2026-08-19** (not trusted from the 08-18 filing):
  - live row `93cffe67-baf4-4fb1-bec9-ba546fb24a54`, step listing via `jsonb_each`:
    `load_candidate_pages` `output_format=array`, `check_candidates`
    `condition='candidate_pages.count > 0'`, `then_step=load_specs`,
    `else_step=complete_no_candidates`. Matches seed `101` exactly.
  - `llm_call_log WHERE agent_type='internal-linker'`: **0 rows, all history**.
  - 20 open (`unresolved`) work items name it as handler.
  - 298 census re-run with the query's own predicate (incl. the HAVING): 26 sites with
    candidates, **8 over the cap**, median 11.5, worst **69**. (Filing said 24/8/12/68-9 — moved
    slightly with the fleet, same shape.)
  - `updated_at = 2026-08-19 07:51` on the row is the v1.0.1314 roll-window bulk write, NOT a
    targeted edit — the 313 file's own landmine; the step config read is what settles it.
- **Code re-read at the cited lines** (unchanged since the CONFIRMED verdict `c4aa3559`):
  `database_actions.go:129` returns the bare slice for array; `:132-145` object branch =
  `rows`/`count`/`columns` **+ first row flattened to top level** (this is why
  `target_page.page_id` works — worth knowing: a column literally named `count`/`rows`/`columns`
  would clobber the metadata; our columns don't collide, and the four legit `.count` readers in
  the fleet may be *relying* on the flatten when they alias `count(*) AS count`).
  `conditional_branch_action.go:275-284`: numeric arm, `ToFloat64(nil)` fails → WARN →
  `return false, nil`. `:397-412`: strategy 5 requires `map[string]interface{}`.
- **No 090 re-run, substitution stated** (owner ruling 2026-07-31): 313 has a CONFIRMED verdict
  already (first iteration, 08-18); a second run on the same mechanism is a duplicate round. The
  substitute is the first-hand re-verification above. (Sought the verdict's doc_notes row by
  correlation for extra colour: not present under `%c4aa3559%` — orchestration_states prunes ~2
  days and the note body evidently doesn't carry the corr string. The bug file's quotation of the
  verdict is the record.)
- **Template/render path checked before designing the fix** (the half that trades a dead branch
  for a broken prompt): `ExecuteLLMPromptAction` → `ExtractFields` (generic arm:
  `extractSingleField` + `UnwrapDeep`; a `{rows,count,columns}` map has no `type`/`result` keys so
  passes unchanged) → `RenderPromptTemplate` = plain Go text/template. So
  `{{range .candidate_pages.rows}}` + `{{.name}}` works; nothing shape-special in between.
- **Design inputs measured, not assumed:**
  - Fleet census of `conditional` step config keys (top-level AND nested, both spellings):
    exactly `condition`/`then_step`/`else_step`, 145 uses each → a declared ActionInputSpec is
    complete; the new key makes 4.
  - `conditional` registers **no** ActionInputSpec today; unknown config keys are ignored at
    execution (bugs_open/101 header in cmd/config-key-audit/main.go) → opting `check_candidates`
    into the flag pre-roll is inert, not breaking. No `_HOLD` needed.
  - `evaluateStringCondition` has two external callers (both test files) → keep its signature,
    add a strict variant.
  - Latest migration number in the dir: 487 → this lane takes **488**. (Re-check at commit time —
    other sessions allocate concurrently.)
- Worked shapes copied rather than reinvented: migration **484** (snapshot → count-gate →
  pre-state gate incl. refuse-to-double-apply → id-scoped jsonb_set → DO/RAISE verify with
  controls); **446** (`[…truncated]` marker); WFA-009 (the opt-in-required precedent and its
  scope argument); WFA-010 (`GetBoolFieldLoud` for the new key).

### Correction discipline for this file
Anything above that later turns out wrong gets a dated `> **CORRECTED:**` block here, and if it
was a durable claim in a commit/bug file, a WRONG_CALLS.md row.

## 2026-08-19 (later) — implementation record

(appended as work lands; see RUNBOOK for the exact commands)
- **Implementation landed (all uncommitted so far):** migration `490_…fail_loud.sql` + ROLLBACK;
  `conditional_branch_action.go` opt-in `fail_on_non_numeric` (wrapper keeps the 3-arg lenient
  signature; spec registered for both names); `cmd/config-key-audit/conditionalshape.go`
  (`--array-producer-conditions`) + tests + `scripts/audit-array-producer-conditions.sh`.
  Tests green; both flag guards MUTATION-PROVEN (default flipped → lenient test RED; error arm
  disabled → fail-loud + AND-propagation tests RED; both reverted, clean run green). First live
  run of the audit: **191 agents, 145 conditionals, exactly 1 finding = internal-linker** — the
  motivating case, independently re-deriving 313's census.

> **CORRECTED 2026-08-19 (same session):** the line above saying "this lane takes **488**" was
> stale within ~3 hours — 488 was taken TWICE concurrently (301 lane's applied migration + 320
> lane's renumbered HOLD) and 489 applied too, exactly the "486/487 taken concurrently" event
> repeating. Renumbered everything to **490** BEFORE committing, so no damage. The check that
> caught it: re-listing the dir + `schema_migrations` before commit, which the NOTES entry itself
> prescribed. Number allocation on this tree has ~hours of shelf life.

- **Pre-existing red test at HEAD, not ours:** `optional_budget_cron_parity_test.go` fails on
  clean `git archive HEAD` — `save_page_meta_description` (320 lane, commit `aeccfc595`) declares
  5 optional keys, cron literal says 0. The 320 lane is ACTIVE today → contributing the finding to
  them, not fixing competitively.

## 2026-08-19 (afternoon) — council submitted, 321 interaction, pre-commit checks

- **Council: SUBMITTED, corr `aef24a7f-2992-4d4f-a6e0-422cd77fcca3`** — admitted without FORCE=1
  (the platform/ files put it in scope; the config-only half alone would have been refused —
  bugs_open/314). Committing with `Council-Submitted:` per the 2026-07-30 norm; budget ~30 min for
  the verdict, find the run by payload not printed id.
- **The pre-existing parity failure resolved itself**: `go test ./cmd/config-key-audit/` is green
  at the current tree — a session fixed the `save_page_meta_description` literal in the interim.
  The planned contribution to the 320 lane is MOOT; nothing to do.
- **Cross-session message from the 321 lane (item_key collisions), and it names a REAL residual of
  this fix:** `create_rewrite_item` carries `item_key_prefix: internal_link` with NO
  `item_key_suffix_field`, so once 490 revives the loop, an N-link plan files ONE content_rewrite
  item and silently drops N−1 (idx_swi_dedup collision — 321's exact class; plan_links asks for
  1–3 links, so the loss is up to ~2/3 of the agent's output). **They fix it independently**
  (their migration, next free number, single jsonb_set on the create_items_loop subtree — fully
  disjoint from 490's paths and gates; apply order immaterial, confirmed both sides). Replied
  2026-08-19: proceed independently; joint disconfirming check = an N-link plan produces N items
  keyed `internal_link_<domain>_<source_page>`. **Until theirs applies, expect ~1 item per plan —
  do not misread that as 490 failing.**
- **Fresh chassis build v1.0.1315 (12:15Z) predates this commit** — the Go halves (WFA-019 flag,
  WFA-018 audit changes) ride the NEXT build; the audit is runnable from the tree meanwhile
  (`go run`), and migration 490 is live-on-apply regardless.
- **Pre-commit checks:** same-file passenger scan on the four shared modified files —
  `git diff --numstat` shows additions only, counts matching my edits exactly (4/2/21/46+5).
  **HEAD + only-my-files** build: `git archive HEAD` + my six code files → build green, tests
  green (`cmd/config-key-audit`, `actions`) — my change does not lean on other sessions'
  uncommitted work.
