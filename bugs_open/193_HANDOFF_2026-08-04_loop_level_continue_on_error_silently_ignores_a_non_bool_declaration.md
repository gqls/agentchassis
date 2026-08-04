# 193 — the LOOP-level `continue_on_error` silently ignores a non-bool declaration, and the substep-level twin beside it warns

**Filed 2026-08-04** by the `bugs_open/173` lane, **at the council gate's `bug_historian`
seat's direction** (corr `549e25fb-acc1-4806-a2a7-95bf73cca806`, round 1, severity **medium**).
**Status: OPEN, UNOWNED.** Latent — nothing is failing today because of it.

> **Why this is a file and not a paragraph in someone's risk section.** It was a paragraph in
> someone's risk section: `173`'s submission named this as a KNOWN RESIDUAL and deferred it.
> The seat's objection is the reason it is now a ticket, and it is worth quoting because the
> reasoning generalises:
>
> *"`loop_actions.go:66` has the identical declared-and-inert defect this edit fixes (bare
> `.(bool)` assertion silently ignoring a malformed declared value) at the LOOP level — the
> plan names this explicitly as a KNOWN RESIDUAL and defers it as 'a separate submission.'
> This is precisely the documented recurring shape 'one call site of a shared judgement gets
> the rigorous fix; the sibling stays heuristic' (§9 index). The plan should at minimum file
> the sibling as a tracked follow-up bug rather than leaving it as prose in this plan's risk
> section, where it will not surface to anyone auditing the mechanism later."*
>
> That last clause is the whole point. A deferral recorded only in the deferring document is
> invisible to the next person auditing the mechanism, because they will read the mechanism,
> not the document that once touched it.

## The defect

`platform/orchestration/actions/loop_actions.go:63-68`:

```go
// continue_on_error: failed iterations are skipped rather than
// failing the entire workflow
continueOnError := false
if coe, ok := config["continue_on_error"].(bool); ok {
    continueOnError = coe
}
```

A loop step declaring `continue_on_error: "true"` — a **string**, the likely JSON-authoring
mistake — takes the `ok == false` branch and silently keeps `false`. No error, no warning,
no audit finding. The author has written the key, the config-key audit accepts it
(`continue_on_error` is in `datahelpers.frameworkStepConfigKeys`), and the loop is strict
anyway.

**The asymmetry is the sharp part, and it is NEW as of 2026-08-04.** `bugs_open/173`'s fix
gave the *substep*-level read of the same key exactly the loud treatment this one lacks
(`resolveSubstepContinueOnError` warns and names the substep and the offending type). So the
same key, in the same file's call chain, is now **loud when a substep mistypes it and silent
when a loop does** — which is a worse state to leave a shared mechanism in than uniform
silence, because the first thing a reader infers from "no warning" is "no problem".

## Blast radius, measured

**No live loop currently declares the key as a non-bool. [MEASURED 2026-08-04]**

```sql
SELECT a.type, s.key AS loop_step,
       jsonb_typeof(s.value->'config'->'continue_on_error') AS declared_type,
       s.value->'config'->>'continue_on_error' AS declared_value
FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.value->>'action'='loop' AND s.value->'config' ? 'continue_on_error';
```

All 10 declaring loops declare it as `boolean` (9 `true`, 1 `false`). **The positive control
that makes that 0 meaningful:** the same query without the type filter returns those 10 rows,
and the fleet has 18 loop steps in total — so the predicate reaches loop configs and a
`string` row was reachable had one existed.

**So this is a latent trap, not a live fault.** It costs nothing today and will cost one
confusing afternoon the first time somebody hand-writes a loop config.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Mirror the substep resolution: warn on a present-but-non-bool declaration** and keep the
   `false` default. Smallest change, restores the symmetry, and the wording can be lifted from
   `resolveSubstepContinueOnError`. Does not change behaviour for any live loop (measured
   above), so it is inert on arrival.
2. **A shared parse helper used by both call sites**, so the loop-level and substep-level
   reads cannot drift again. Strictly better than (1) on the "one implementation, two callers"
   rule 016b §9 states — *"prefer one implementation with two callers over two implementations
   with a test that they match"* — and it is the reason this bug exists at all. Slightly more
   surface: the two reads have different fallbacks (`false` vs the loop's value), so the
   helper takes the fallback as a parameter.
3. **Reject at validation time instead**, in `ValidateWorkflow` — a `continue_on_error` that
   is not a bool is a definition error and could be caught before any message arrives. But
   note 016b's warning of 2026-07-29 about exactly this move: a new hard rule that measures
   zero impact today is a rule about the sample, not the system, and this one would newly
   reject a definition that currently runs (harmlessly). Warning-level only, if taken.

**Prefer (2).** It is the option that makes a *future* divergence unrepresentable rather than
correcting the current one, and the divergence is the actual defect — the silence is only its
symptom.

## How to verify a fix

Do not grade this on a build. Give a loop `continue_on_error: "true"` (string) and confirm the
warning fires and names the loop; then give it `true` (bool) and confirm no warning and
tolerant behaviour. **Mutation-check it**: with the fix reverted, the string case must fail the
test — a test that passes against both spellings is asserting nothing (`WRONG_CALLS.md`
2026-08-03).

## Related

- `bugs_open/173` — the substep-level twin, fixed 2026-08-04 (WFA-008). Its landmine entry in
  `LANDMINES.md` already carries the "presence is not truth" trap for anyone editing either read.
- `016b` §9 — "one call site of a shared judgement gets the rigorous fix; the sibling stays
  heuristic", the family this belongs to.
- The wider class: a config key that is declared, accepted by the audit, and read by nothing —
  `bugs_open/134`, and WFA-005's `output_format` case.
