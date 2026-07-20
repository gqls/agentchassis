# 032 — the completion verifier reads a DELETED component as a successful fix

> ## CLOSED 2026-07-20 19:00 BST — FIXED AND LIVE
>
> Fixed in `a467baa11` (the conservative floor from § "Fix candidate": return an
> error, never a verdict, so the gate's fail-OPEN policy turns a false success
> into a visible unknown). Image rolled 18:58:33 BST; verified against the
> running binary, not git and not the tag (pod `agent-chassis-5567d99bd6-5snzn`):
>
> ```
> strings /app/agent-chassis | grep -c "cannot verify: component"   -> 1
> ```
>
> Regression tests pin both directions via sqlmock: absence must not claim
> success, and a present-but-empty component must still fail closed.
>
> **One thing deliberately NOT done, and still worth doing.** The stronger option
> in § "Fix candidate" — if the page still EXPECTS the component (a
> `plan_sections` entry, a slot reference) then absence is *deletion*, not
> ambiguity, and `Resolved: false` is the honest answer — remains open. It is the
> `empty_sections_loop_integrity` thread's call and this floor does not preclude
> it. Reopen or file a follow-on rather than treating the error-return as the
> finished shape.
>
> The coverage half of the finding still lives in `bugs_open/021`, unchanged:
> `RegisterVerifier` is called for a handful of item types out of ~50, so most
> items still complete on the handler's self-report.


**Filed:** 2026-07-19 by the reasoning-dataset thread.
**Found by:** the council gate's `bug_historian` seat, during review of a plan
that proposed *copying* this behaviour to two more item types
(`SUBMISSION_CORR=66dbd0dd-de5f-4f50-acd3-f5f3d817dbd9`, verdict REVISE).
**Severity:** latent, silent, content-loss-masking. Not an outage.
**Status:** OPEN. One-branch fix; needs the owning thread's call on policy.
**Owner:** the `empty_sections_loop_integrity` thread (it built the completion
gate and the only registered verifier).

---

## The defect

`VerifyEmptySectionResolved` (`platform/orchestration/actions/discovery_checks/check_empty_sections.go:205`)
is consulted by the completion gate before an `empty_section` work item is
stamped `complete`. When the component row is gone it reports **success**:

```go
	if err == sql.ErrNoRows {
		// Component removed — nothing left to be empty.
		return VerifyResult{Resolved: true, Detail: "component no longer exists"}, nil
	}
```

The comment states the assumption plainly, and the assumption is unsafe. **A
missing `page_components` row is equally the signature of this platform's most
repeated failure** — a rebuild silently deleting a component. This repo has paid
for that class at least twice already (`bugs_open/021` cites "rendered_html tools
deleted twice independently, the page-assembler visible-content filter needing
two independent fixes"; `bugs_open/012` is the truncation-overwrite instance).

So the verifier cannot distinguish:

| what happened | what it reports |
|---|---|
| the empty section was genuinely fixed or deliberately removed | `Resolved: true` ✅ |
| **a rebuild silently deleted the whole component** | `Resolved: true` ❌ |

The second case is a content-loss incident being recorded as a verified fix — by
the very mechanism that exists to stop self-reported completions being trusted.
It is the `bugs_open/012` shape one layer up: the guard adopts the blind spot
that caused the incident it guards against.

## Why it matters more than the row count suggests

Live 2026-07-19: 4,570 items `complete`, **5** carrying a `result._verification`
record, **9** ever `verified` (all `empty_section`, none since 2026-07-14). Small
today — but this is the *reference implementation*. It is explicitly the pattern
new verifiers are told to follow (`verifiers.go:17-19`), and a plan to copy it to
`hardcoded_section_colors` and `phantom_internal_link` was in council review when
this was caught. The blast radius is every verifier not yet written.

> The seat's own words, worth keeping because they name the trap precisely:
> *"the detection layer improves, but the new check adopts the same blind spot
> that caused the original content-loss incidents."*

## Fix candidate (deliberately conservative)

Return an **error**, not a verdict. The gate's documented policy is to fail OPEN
on verifier error — *"verifier errors → fail OPEN (complete, recording the error
under `result._verification`)"* (`complete_work_item_verification.go:14-21`). So:

```go
	if err == sql.ErrNoRows {
		// Component row absent. AMBIGUOUS: equally a legitimate removal and a
		// rebuild silently dropping the component. Never report success.
		// Error => gate fails OPEN (item still completes, nothing wedges) and
		// result._verification records that verification could not be made.
		return VerifyResult{}, fmt.Errorf("cannot verify: component %s no longer exists (fixed or silently deleted — indistinguishable here)", componentID)
	}
```

**Item flow is unchanged** — the item completes either way. What changes is that
a false success becomes a *visible unknown*, which is the difference between data
that misleads and data that is merely incomplete.

**The stronger option, for the owning thread to weigh:** if the page still
expects that component (a `plan_sections` entry, a slot reference), absence is
not ambiguous at all — it is deletion, and could return `Resolved: false`. That
is a better answer and a bigger change; the error-return above is the safe floor
and does not preclude it.

**Do NOT** return `Resolved: false` unconditionally — a legitimately removed
component would then burn an attempt and, at `max_attempts`, strand a genuinely
fine item in `failed`.

## How to verify the fix

```bash
PSQL='kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db'

# 1. Before: any completed empty_section item whose component row is gone was
#    verified on the unsafe branch.
$PSQL -c "SELECT w.id, w.status, w.spec->>'component_id' AS comp
          FROM site_work_items w
          WHERE w.item_type='empty_section' AND w.status IN ('complete','verified')
            AND NOT EXISTS (SELECT 1 FROM page_components pc
                            WHERE pc.id::text = w.spec->>'component_id')
          LIMIT 20;"

# 2. After the image rolls, the same class must record a verification error
#    rather than a silent success:
$PSQL -c "SELECT result->'_verification' FROM site_work_items
          WHERE item_type='empty_section' AND result ? '_verification'
          ORDER BY updated_at DESC LIMIT 5;"
```

Go change — **inert until an image is rebuilt and rolled**. Verify against the
running pod, never git:
`kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "cannot verify: component"'`

## Related — do not fork these

- **`bugs_open/021`** — *the completeness guard covers ONE durable-write path.*
  Same family, already filed from a council `bug_historian` objection on the
  identical shape. **The coverage half of this finding belongs there, not here**:
  `RegisterVerifier` has been called exactly **once** in the codebase
  (`check_empty_sections.go:38`), so ~49 item types complete on the handler's
  self-report. An instance note has been appended to 021 rather than filed
  separately.
- **`bugs_open/012`** — the truncation-overwrite instance of the same "trust the
  artefact, not the status" family.
- **016b §9** — transferable pattern added (*a verifier that treats a missing
  target as success cannot distinguish repair from deletion*).
