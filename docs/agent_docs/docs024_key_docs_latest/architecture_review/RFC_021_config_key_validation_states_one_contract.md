# RFC 021 — the config-key validator is a four-state precedence machine built from four bug tickets; write the contract down, and decide who inventories blast radius before a live hard-fail ships

**Raised 2026-08-10 by the bugfix_234_dead_spec_key lane, at the council's direction.**
Round-1 verdict on corr `3eb0d1f1-6929-4131-bbef-c636256aa667`: **REJECTED, hard veto from
`guardian`** — while `architecture` **objected but approved proceeding** ("point fix on
individual merits; architecture underneath deserves a short RFC before the next
increment") and seven other seats approved. The seats disagree about the same change in
the same round, which is the RFC_002 shape, and per the owner rulings of 2026-07-28/29
the code stays while the precedent gets fixed **here, by a human**. The change is
committed (`d278d7b25`), registered (SCR-007, same commit), and rides v1.0.1278 unless
the owner directs otherwise (§4).

## 1. What exists now (the contract nobody wrote down)

`checkStepConfigKeys` (platform/validation/workflow.go) runs on every step of every
message and now implements FOUR states, shipped by four lanes over three weeks:

| state | declared via | consequence at validation | shipped by |
|---|---|---|---|
| **removed** | `RemovedConfigKeys` (map key→message) | **hard error**, message names the replacement; independent of opt-in | bugs_open/234 |
| **strict-unknown** | `StrictConfig: true` + opted in | **hard error**, generic, names the keys | bugs_open/101 (mechanism), 234 (first flip) |
| **unknown** | opted in via `CheckConfig`/`ConfigKeys` | once-per-pod Warn | bugs_open/101 |
| **deprecated** | `Deprecated` (reference alias) / `DeprecatedConfigKeys` (setting alias) | none here — recognised on purpose; the read site warns | pre-existing / bugs_open/136 |

**Precedence, as implemented and test-pinned:** removed → strict/unknown → (deprecated
and recognised keys are silent). Removed keys are excluded from the unknown set (no
double-report under the softer label). `ConditionalKeys` annotates recognised keys for
the offline report only. The offline mirror is `scripts/audit-config-keys.sh` +
`cmd/config-key-audit` (sections: REMOVED IN USE / UNKNOWN / DEPRECATED / CONDITIONAL /
SUSPICIOUS / UNDECLARED; exit 1 on removed, unknown or suspicious).

Each increment was individually sanctioned (opt-in field, unsafe default OFF — owner
ruling 2026-08-02 #2; additive-and-inert — 2026-07-29 #1) and individually
mutation-tested. **What was never reviewed is the whole**, and the `architecture` seat is
right that increment #5 will face the same "additive, opt-in, doesn't relocate" argument
with no accumulated record to check it against. This section IS the written contract the
seat asked for; a future increment edits this table in the same commit.

## 2. The guardian's veto, stated fairly

Not about correctness — about **how a live hard-fail reached the shared validator**:

- `ActionInputSpec` is consumed by every registered action; a new field plus an
  `UnknownConfigKeys` carve-out changes shared validation semantics platform-wide.
- `checkStepConfigKeys` is dispatch-adjacent plumbing running fleet-wide per message; a
  hard-fail branch there is not a contained point fix.
- `StrictConfig` on `create_work_item` converts any FUTURE undeclared key on the fleet's
  busiest work-item action from warning to outage, "on behalf of pipelines this plan does
  not name or consult".
- Safest contained alternative (guardian's words): ship the OFFLINE audit now; hold the
  live hard-fail + strict flip for a review that inventories every producer.

Counterpoints already on the record: the hard-fail fires only for (a) a key an action's
own maintainer has declared retired, or (b) an unknown key on an action whose maintainer
declared the contract complete after an all-depths fleet census (re-run fresh 2026-08-10
11:15Z: 0 `spec` carriers, 0 unrecognised keys on any live create_work_item step). The
warn-only alternative is the recorded failure mode that let four dead keys ship and
survive on this one action, one of them costing months of silently-lost behaviour.

## 3. Questions for the owner

1. **Does a live hard-fail on the shared validator require a named-producer inventory
   before each adoption** (guardian's position), or is the audit-clean census + the
   declaring action's own contract sufficient license (the shipped position)? The answer
   becomes the adoption protocol for every future `StrictConfig` flip and
   `RemovedConfigKeys` declaration.
2. **Keep or split v1.0.1278's enforcement?** Options: (a) it rides the next roll as
   committed — the audit ran clean, the canary check is in the lane RUNBOOK; (b) a
   follow-up commit reverts the validator hard-fail + strict flip to warn-only, keeping
   the offline audit, until Q1 is answered. Forward-only either way; (b) is the
   guardian's contained alternative, at the cost of the class staying open on the one
   action that just demonstrated the damage.
3. **The named siblings**: `mark_page_needs_attention.notes_field` /
   `validation_issues_field` (356's adjudicated-dead, left standing to be REPORTED).
   Tracked in `bugfix_136_config_key_aliases/HANDOFF_2026-08-09_deferred_items.md` as
   RemovedConfigKeys candidates. Adopting them is increment #5 and waits on Q1.

## 4. Disposition of the round's other objections (all answered, for the record)

- `editquality` (M): census cited rather than fresh → re-run at commit-gate AND again
  2026-08-10 11:15Z, both clean; henceforth the protocol is census-at-commit, recorded.
- `bug_historian` (M): siblings named but untracked → now tracked (§3.3).
- `prior_art` (M): "no `config[\"spec\"]` read" rests on prose → confirmed by body read
  (the action reads spec_data/spec_paths/spec_literal only, :247-296) and pinned both
  directions by `create_work_item_config_contract_test.go`, which fails if the action
  gains a read the spec omits or vice versa.
- `architecture` (missing): a single written spec of the states → §1.

- **sources:** council report corr `3eb0d1f1…` (diagnosis_artifacts, kind=council_report);
  SCR-003/006/007; bugs_open/101, 136, 234; commit `d278d7b25`
