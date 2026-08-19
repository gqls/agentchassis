# PLAN — bugs_open/321: loop-filed work items collide on a site-wide `item_key`

Lane created 2026-08-19. Bug: `bugs_open/321_HANDOFF_2026-08-19_tool_suggester_loses_most_of_its_suggestions_to_a_site_wide_item_key.md`.

## The defect, in one paragraph

`create_work_item_action.go:225-260` builds `item_key` as `<item_key_prefix>_<domain>` —
site-wide. `idx_swi_dedup` is `UNIQUE (site_id, item_key)` over non-terminal rows. A step
that files N items per site inside a loop therefore loses iterations 2..N silently (the
loop reports success either way). Measured on tool-suggester: **40 suggestions → 11 items,
~72% lost**, and it is a race, not a clean cap. The remedy already exists in the same
function — the optional `item_key_suffix_field` config key (unresolved = deliberate hard
error) — and the affected steps simply do not set it.

## Scope decision: the CLASS, not the instance

Fleet census (2026-08-19): 20 live `create_work_item` steps carry `item_key_prefix`; the
defect exists only where the step is nested inside a `loop` — six steps, of which
**four lack the suffix**:

| agent | step | suffix path chosen | why that path |
|---|---|---|---|
| tool-suggester | `create_novel_item` | `current_suggestion.function` | 239/239 historical suggestions carry a non-empty unique `function`; 0 intra-answer duplicates |
| tool-suggester | `create_library_item` | `current_suggestion.function` | same evidence |
| component-quality-auditor | `create_work_item` | `current_component.component_id` | the step's own `spec_paths` already HARD-REQUIRES this exact path, so the suffix adds zero new failure modes; component_id is the unique id (two components can share a `function`) |
| internal-linker | `create_rewrite_item` | `current_link.source_page` | same structural argument — `spec_paths.page_name` hard-requires the identical path. Page-level granularity is correct: two links into one page dedupe to one rewrite item deliberately |

`tool-auditor`'s two steps already set the suffix (`tool_data.page_id`) — the proven idiom.

The two "latent" steps (component-quality-auditor, internal-linker) have never filed an
item — but internal-linker's loop is dead only because of a broken conditional that the
`bugs_open/313` lane's migration 490 is about to fix, so its collision goes LIVE when that
lands. Fixing all four now is the class fix; both latent fixes are risk-free by the
spec_paths argument above.

## Decisions and their reasons

1. **`continue_on_error: true` on tool-suggester's `create_items_loop`, in the same
   migration as the suffix.** The suffix's unresolved-path hard error would otherwise
   abort the whole loop (0 items — worse than today's partial loss): the loop is the only
   one of the four without the flag, and `routeToErrorStepOrFail` falls to `failWorkflow`
   because no substep carries an `error_step`. Two fable reviews disagreed on whether a
   skipped iteration is durable; settled by reading the code: `skipToNextLoopIteration`
   PERSISTS `{loop}_iter_{N}_error` to `orchestration_states.collected_data`
   (`loop_error_handler.go:141-149,185-188`) and the aggregation folds `status:"error"`
   into `items_created` (`loop_actions.go:505-511`), an output field of the workflow. So
   the skip is durably recorded, converges the loop onto the estate norm (the other three
   all set it), and the N-suggestions→N-items pairing check is the backstop.
2. **`recurrence_expected` stays OFF.** Post-fix, a re-run can repeat a key and
   `insertWorkItem`'s two-strike block brands the THIRD item on a repeated key
   `unresolved`. Measured: only 4 of 214 historical (site,function) pairs ever reached a
   third suggestion (194×1, 16×2, 3×3, 1×4) — bounded, and arguably correct (a tool that
   failed twice should not be retried for ever). Duplicate-rebuild waste at steady state
   ≈ 10% of suggestions repeat a function on the same site across runs (25/239), each
   waste ≈ one tool-generator chain (~12.5k in / 18k out tokens, ~$0.30); the component
   layer is idempotent on `function` (`create_tool_component_action.go:225-246` returns
   `already_exists`), so no duplicate pages.
3. **No throttle** (owner, 2026-08-19): all suggestions become items; volume bounded
   upstream (prompt caps at 8, loop `max_iterations: 10`, dispatch `max_items: 5`/pass,
   add_tool priority 120/130 sorts behind default-100 work). Report real volume after the
   canary.
4. **Framework control = offline detector + runtime Warn** (owner chose detector+guard;
   the guard's first shipment is a zero-authority Warn — see below). New mode in
   `cmd/config-key-audit` (the estate's config-auditor family, WalkSteps traversal, no
   second descent), wrapper script, daily CronJob writing one `doc_notes` row per run
   including clean runs, parity pinned by test.
5. **Runtime refusal deferred, Warn ships first.** A default-OFF refusal protects only
   authors who already know the class; a default-ON refusal is RFC scope; and either is a
   second implementation of the detector's rule that can drift (`bugs_open/144` pattern).
   Instead: in `CreateWorkItemAction`, when the insert deduped (`!inserted`) AND the step
   config carries the framework-injected `loop_var_name` (`loop_expansion_handler.go:155`)
   AND no suffix is configured — log Warn naming the class. Zero config surface, no
   behaviour change; it is also the only net under the detector's blind spot (a suffix
   that resolves but is loop-invariant). Escalating it to a refusal is a follow-up once
   the detector proves the fleet stays clean.

## Phases

- **Phase 1** — migration (next free number at write time; 493 as of 12:30Z) setting the
  four suffixes + the one `continue_on_error`. 484's shape: snapshot_agent, DO/RAISE
  pre-gate (incl. md5 anchor `68889d441446cb03a9dec2968919eb3b` on tool-suggester's
  `create_items_loop` subtree, re-verified after the 12:14 fleet-wide `updated_at` touch),
  id-scoped `jsonb_set`, in-transaction DO/RAISE post-verify asserting 484's edits
  survive, rollback sidecar. Config-only → live on apply, outside the council gate (484's
  recorded precedent). Prove fail-loud by applying twice: second run must abort at the
  pre-gate.
- **Phase 2** — detector: `cmd/config-key-audit/loopitemkeys.go` + tests + `scripts/
  audit-loop-sitewide-item-keys.sh` + CronJob service dir + makefile targets. Proven by
  firing against the PRE-fix config (must report the 4), then clean after Phase 1.
- **Phase 3** — the Warn in `create_work_item_action.go` (+ council submission for the Go
  halves, `Council-Submitted:` trailer).
- **Phase 4** — canary: next tool-suggester run must produce N items for N suggestions
  (pairing query in RUNBOOK); tripwire query for the hard-error arm for a week; report
  actual build volume to the owner.
- **Phase 5** — register (WFA entry), LANDMINES extension of the "key coarser than its
  finding" entry, bug file update; close only when fixed AND live AND verified at the
  artefact.

## Coordination

- `bugs_open/313` session (owns internal-linker, migration 490 in flight): messaged
  12:30Z — my edit is on a disjoint subtree; offered to drop internal-linker from my
  migration if they prefer to fold it into 490's rollout. **Awaiting reply; will not
  apply the internal-linker leg until they answer or their 490 lands, whichever first.**
- `bugfix_275` lane (tool-suggester's 484, applied 10:23Z): lane is CLOSED per its own
  handoff; my pre-gate asserts their edits survive rather than racing them.
- `bugs_open/184` NOT touched — live session mid-canary owns it.
