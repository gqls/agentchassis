# HANDOFF — 2026-08-08 (evening) — `bugfix_136_config_key_aliases`, cold start

**Read this first, then `bugs_open/136` §§6–9** (the numbered sections at the foot; §1–5 are
the original 2026-07-28 filing plus my first update, and **§1's harm claim is superseded by
§6** — do not quote it).

## Where it stands in one paragraph

`bugs_open/136` (config says `*_domain`, code reads `*_pipeline`) is **fixed, approved, and
live on v1.0.1267**. The fix is a framework seam — a declared alias vocabulary for literal
config settings — registered as **SCR-006**. The binary is proven to carry it. **The runtime
behaviour is NOT witnessed, and the witness I designed cannot work.** That is the only thing
still owed, and §9 names the two observations that would actually settle it. The bug stays in
`bugs_open/` (owner ruling 2026-08-06: a finished bug stays there).

## What is proven, and what is not — do not blur these

| claim | status | evidence |
|---|---|---|
| the binary carries the change | **PROVEN** | pod-grep across the roll, `new` 0→1, `pos_control` 1 throughout, `neg_control` 0. Recipe in §8 / RUNBOOK |
| the declarations join real live definitions | **PROVEN** | `./scripts/audit-config-keys.sh` UNKNOWN KEYS **4 → 1** against production `agent_definitions` |
| the alias is honoured at runtime | **UNWITNESSED** | and NOT obtainable from `create_work_item` — see below |
| `run_discovery_checks` / `triage_detected_items` behaviour | **UNWITNESSED** | quiesced by owner ruling; no live proof available and none should be manufactured |

## The one live thing still owed, and why the obvious route is a dead end

**Do not try to witness this via `create_work_item`.** All nine live `item_domain` carriers
set `"build"`, which is exactly `create_work_item`'s hardcoded default, so the row it writes
reads `pipeline='build'` whether the alias fired or not. The observation cannot come out
otherwise. I nominated it as the witness before checking that; it is logged in
`WRONG_CALLS.md` as this bug's own defect reproduced inside its own verification.

Two things WOULD discriminate:

1. **Find where `create_work_item` actually logs.** The deprecation warning fires on every
   call from those nine agents, so a single located log line settles it. I swept all 23 pods
   carrying the image (`--since=3h`) and got 0 — **but `create_work_item` also appears 0
   times in those same logs**, so the positive control is zero and the sweep proves nothing
   except that I looked in the wrong place. Answer "where do these executions log?" first.
2. **Set one carrier's `item_domain` to `"content"`** and watch the next row's `pipeline`.
   Cheap, reversible, and the only single-step discriminator that exists. It is a definition
   edit on another lane's agent, so check ownership first — I did not do it.

## What shipped (4 commits, all pathspec'd, no passengers)

- `3f93456fd` — the seam + 4 adopters + audit surface + SCR-006 + LANDMINE (13 files)
- `51fb9383f` — bug-file correction + the lane's standing five
- `d3f0e93bc` — council trail + two corrections to my own claims
- `4756535cc` — live verification + the third wrong call

**Council: APPROVED**, `433de2c0-682f-4d8d-8c48-28637309f1ba`, 12 seats, 2 advisory, none
high. The `architecture` seat confirmed the scope call explicitly: additive/inert/opt-in
struct field + same-commit register and LANDMINE entries = **normal gate, not an RFC**. The
same-commit registration is what earns that, not the size of the diff.

## The mechanism, in case you need to extend it

`ActionInputSpec.DeprecatedConfigKeys` (old setting key → canonical) + `ResolveConfigSetting`
in `platform/orchestration/datahelpers/config_key_aliases.go`. Opt-in, inert until a spec
names it. **The landmine to know before touching any of this:** `ActionInputSpec.Deprecated`
is a *different mechanism* — it resolves the old key's **value as a dot-path into
collected_data** (Strategy 3), so putting a literal setting there silently takes the default
*and* silences the detector, which is worse than not declaring it. That asymmetry is the
whole reason two of the three renames never got a shim.

## Still owed, deliberately not done (all recorded on the bug file §5)

1. The **four mislabelled rows** are not repaired (2 `complete`, 2 `detected` — and `detected`
   is a queue with no consumer, `bugs_open/083`).
2. **`summary_template` is still biting** — two human-review items captioned with their own
   `item_type`. **Not an alias case**: aliasing it to `summary` would ship a raw
   `{{.input_data.topic}}` to a human reviewer. Needs a render-or-literal decision. **This is
   the highest-value remaining item and it is an owner call, not a code question.**
3. `plan_sections.domain` — genuinely dead, left reported because `page-build-handler` is hot.
4. `create_work_item`'s full opt-in — blocked on adjudicating (2), `spec_fields`, `domain`.
   Note `priority` **is** read, via `GetIntField`; the bug file's §3 read-list was wrong.
5. `resolveAgentTypeForSpawn` hand-rolls `group_type` → `agent_type`, the same class —
   a convergence candidate, **measured at zero live carriers**, so not urgent.

## Traps this lane paid for (all in `WRONG_CALLS.md`, 2026-08-08 ×3)

- **`grep 'config\["'` cannot see a key read through a helper.** Grep the *key name*.
- **A `$n` in a predicate is an unmeasured input.** `WHERE pipeline = $2` looks like it
  partitions on pipeline; its one caller always passes `"build"`, so it partitions on nothing.
  Read the callers before citing a query as evidence of harm.
- **Nominate a witness only after writing down the disconfirming observation.**
- **`-l app=agent-chassis` returns 2 pods; 23–25 run the image.** Enumerate by IMAGE for any
  deploy proof, and pair a new-string grep with a positive control — source-uniqueness proves
  a string *would be* new, never that it is spelled right in the binary.
- **`go build` at HEAD vs `go test` in the tree is not a comparison** — `go build` skips vet.

---

## UPDATE 2026-08-08 (night, next session) — the one live thing owed above is DELIVERED

The runtime behaviour is **WITNESSED**: `bugs_open/136` §11. Neither of the two
observations proposed above was the route —

1. "Find where `create_work_item` logs" is answered and is a dead end: the executions log
   to pod stdout, and an active chassis pod retains **<1 second** of log (measured 0.4s;
   LANDMINES entry added). The §9 sweep was measuring retention, not behaviour.
2. Editing a live carrier was not needed. A throwaway lane-owned definition
   (`alias-witness-136`) filed a born-`cancelled` item with `item_domain: "content"` —
   non-default, so the observation could come out otherwise — and the row landed
   `pipeline='content'` at 22:25:39Z on v1.0.1268. Recipe in the lane RUNBOOK; definition
   deactivated after the run.

The proven/unproven table above should now read: runtime = **WITNESSED**. Remaining items
are §5-as-amended-by-§10 only (all deliberately deferred; the highest-value one,
`summary_template`, was resolved by owner decision A+D in §10).
