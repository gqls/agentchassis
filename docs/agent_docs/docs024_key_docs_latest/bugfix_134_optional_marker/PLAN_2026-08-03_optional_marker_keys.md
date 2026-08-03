# PLAN — bug 134: doc-notation `?` leaked into two live config key names

**Lane opened 2026-08-03.** Source case: `bugs_open/134_HANDOFF_2026-07-28_optional_marker_leaked_into_the_config_key_name.md`.
Single-session scope expected — this dir carries the plan, the council submission
and a NOTES file; no README/SUMMARY unless the work outgrows the session.

## Re-verification (all four figures re-measured live, 2026-08-03 ~11:50 BST)

- Live `product-spec-refresher` row still carries `"category?"` and `"limit?"` in
  `workflow.steps.refresh_specs.config` — query in the bug file, re-run verbatim.
- `orchestration_states` runs for the agent: **0** (all retained history).
- The fleet punctuation-key sweep returns **exactly the same 2 rows** — still the
  only instance.
- `refresh_product_specs_action.go` reads `category` (:211) / `limit` (:215),
  spec at :175-178 has no `CheckConfig`. Seed 156 lines 60-61 carry the bad keys.

Also checked before claiming the bug: `who-owns` (dormant since 07-28), live
transcript symbol grep (`refresh_product_specs|product-spec-refresher`) — hits are
bug-file quotes from other sessions' own bug-picking, not work; the
`needs_diagnosis` queue is empty; the five open `site_work_items` matching
"product-spec" are about a robot-hands page *component*, unrelated.

No fresh 090 run: the filed root cause is fully cited, every figure re-verified
first-hand today, and the mechanism is locally checkable (the extractor reads
exact key names; a `?`-suffixed key is unreadable by construction — grep shows no
Go reads one). That is the named escape hatch of the 2026-07-31 owner ruling,
stated here deliberately.

## Decision — three parts, ordered by what closes the door

1. **Data: fix the seed AND the live row.** Seed 156 (`"category?"` → `"category"`,
   `"limit?"` → `"limit"`) so a replay restores correctness, not the defect; and
   migration **298** for the live row — modelled on 296's shape: `snapshot_agent`
   first, guarded idempotent UPDATE, `doc_notes` record, DO/RAISE verify asserting
   the fixed keys AND neighbours (`site_id` mapping intact, sibling steps intact),
   live `BEGIN;`/`COMMIT;` (the trailing-ROLLBACK landmine), rollback comment.
2. **Contract: `CheckConfig: true` on `RefreshProductSpecsInputSpec`.** The action
   passes its own spec to `ExtractActionInputs`, which is exactly the case
   `action_inputs.go`'s own doc says `CheckConfig` is for — opting in asserts
   nothing new. The next typo on this action becomes a runtime warning and an
   offline `audit-config-keys.sh` finding instead of six silent months.
3. **Class: a `--suspicious-keys` mode on `cmd/config-key-audit`.** A top-level
   step-config key containing `?`, `*`, a space or `:` is doc-notation punctuation,
   never a readable key — true for **every** action, including the ~151 that have
   not opted into SCR-003, which is precisely the population where this instance
   sat invisible since 07-15. Same house pattern as `--unregistered-actions`:
   pure function + fixture test, same stdin export, walked with
   `validation.WalkSteps`, refuse an empty decode, exit 1 on findings. Wired into
   `scripts/audit-config-keys.sh` as a fourth section using the WORKFLOWS_JSON it
   already fetches. Character set kept to exactly the four the bug's sweep used —
   deliberately narrow so the check starts green post-fix and any future finding
   is signal (`narrowing-a-detector-can-make-it-inert` noted: these four are the
   documented doc-notation set, not a guess).

**Not doing:** declaring `category?`/`limit?` in the spec (the WRONG_CALLS
2026-07-28 trap — declaring a dead key silences the detector); touching the
runtime validator's fleet-wide behaviour (advisory offline detection first; a
fleet-wide runtime warning on all agents is a shared-mechanism guarantee change
this bug does not need).

## The behaviour-change call (stated, per the bug's own Ownership section)

Correcting the keys makes caller-supplied `category`/`limit` start resolving.
Nothing depends on the current behaviour: **0 runs ever**, no `scheduled_tasks`
row, and the seed's own header documents `{site_id, category?}` as the intended
input contract — the author meant these to resolve. When no caller passes them,
behaviour is byte-identical (action defaults: `gripper`/20). The 101-residual-2
precedent (leave config honest-but-warning) protected a *live* agent's observed
behaviour; there is no observed behaviour here to protect. Council reviews this
call — that is what the round is for.

## Ordering

None. The deployed image has registered `refresh_product_specs` since 07-15; the
key fix works against the current binary. Go changes ride the next roll
(`CheckConfig` is inert until then; the audit mode runs from source via
`go run`). No ordering constraint is claimed.

## Verification

- `go build ./...`, `go vet` on touched packages, `go test ./cmd/config-key-audit/... ./platform/orchestration/datahelpers/...`
- Apply 298 by hand, register with `--record-only`; re-run the punctuation sweep
  → 0 rows; assert `refresh_specs.config` now maps `category`/`limit` and
  neighbours survived.
- `scripts/audit-config-keys.sh` runs end-to-end; new section prints `none`
  post-migration; `--specs` shows `refresh_product_specs` `opted_in: true`
  (that half only after the next roll — from source it is checkable now).
- Negative control for the new mode: fixture with a `category?` key must be
  flagged; clean fixture must not; deleting the check must fail the fixture test.
