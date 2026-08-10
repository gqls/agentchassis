# PLAN 2026-08-10 — bugs_open/213: a verifier speaks for an ITEM, not for a NAME

**Lane opened** 2026-08-10. **Bug:** `bugs_open/213_HANDOFF_2026-08-07_two_producers_share_an_item_type_and_the_verifier_implements_only_one.md`.

## The defect

`site_work_items.item_type` is the join between **who filed an item** and **what
predicate is re-run before it closes**. Two producers file `hardcoded_section_colors`;
only one wrote the registered verifier. The other producer's items are graded against
a predicate that has nothing to do with the defect they describe. The verifier answers
**its own question correctly** and returns `Resolved: true`, so the item closes
`complete` with the defect untouched.

Nothing errors. Nothing is mis-written. **The specification is the defect.**

## Re-verification on pickup (2026-08-10) — the bug has WORSENED

The bug file recorded 7 producer-B `complete` rows on 2026-08-07. Re-run today:

```sql
SELECT status,
       count(*) FILTER (WHERE spec->>'audit_source' = 'design-audit') AS producer_b,
       count(*) FILTER (WHERE spec->>'audit_source' IS NULL)          AS producer_a,
       count(*) AS total
FROM site_work_items WHERE handler_agent='color-variable-fixer' GROUP BY 1;
```

| status | producer B | producer A | total |
|---|---|---|---|
| `complete` | **11** (was 7) | 2 | 13 |
| `detected` | 0 | 3 | 3 |
| `unresolved` | 0 | 5 | 5 |

[MEASURED 2026-08-10] **Zero producer-B items have ever failed to close** — 0
unresolved, 0 detected, across the type's whole life. Every row that ever failed to
close is producer A's (8 of 8). That asymmetry is the finding: a route where one
producer's items never fail is not a route with a good producer; it is a route whose
grader cannot see that producer's defect.

**Fleet scope** [MEASURED 2026-08-10]: across all 11 registered verifier item_types,
counting `count(DISTINCT COALESCE(spec->>'audit_source','<none>'))`,
`hardcoded_section_colors` is the **only** verified item_type with two producers
today (21 rows, 13 complete). Live exposure is one route; the structural hole is
general. This is a disconfirmable zero — the same query returns 2 for the one route
that has the shape, so it can distinguish.

## The pre-flight that decided the design [MEASURED 2026-08-10]

The scope test below keys on `spec.check`. That is only safe if it actually separates
the two producers, so it was measured before being designed in — and it could have
come out otherwise:

```sql
SELECT status, (spec ? 'check') AS has_check_key, spec->>'audit_source', count(*)
FROM site_work_items WHERE item_type='hardcoded_section_colors' GROUP BY 1,2,3;
```

| status | `spec ? 'check'` | audit_source | count |
|---|---|---|---|
| complete | **f** | design-audit | 11 |
| complete | **t** | *(null)* | 2 |
| detected | **t** | *(null)* | 3 |
| unresolved | **t** | *(null)* | 5 |

**10 of 10 producer-A rows carry `spec.check`; 0 of 11 producer-B rows do.** A clean
partition with no overlap. `git show 62a79c8ac` confirms producer A's check has
written `spec.check` since its first commit, so the positive shape match is safe for
every historical row, not just today's.

## The decision

**Primary fix: make the producer part of the join**, in two ordered halves.

### Half A — producer B gets its own item_type, `dark_section_audit`

Vacates the only live collision and restores the invariant the registry silently
assumes: one item_type, one predicate. One line in `designItemTypes`, plus the
coverage-guard classification.

### Half B — a verifier declares the REMIT its predicate speaks for

An opt-in `Grades` scope test on `VerifierPolicy`, consulted at the completion gate
before the verifier runs. An item the verifier disclaims is neither completed nor
graded: it routes into the existing RFC_017 fail-closed attempt machinery with
`_verification.status = "out_of_scope"`.

### Why this closes the door rather than narrowing it

The defect is that `verifiers[itemType]` treats a **name** as proof that the
verifier's predicate is the item's predicate. Half B makes the false-verified state
unrepresentable at the one choke point every completion passes
(`verifyBeforeComplete`), for **every** producer channel — code, DB config,
hand-filed rows — because the scope test reads the **row itself**
(`VerifyTarget.Spec`, which the gate already carries), not an enumeration of
producers.

That is precisely what dodges the refutation of the bug file's candidate 3.
Candidate 3 asked *"who filed this?"* against a list that live config can outrun —
any agent definition can file any `item_type` from DB config with no code change.
`Grades` asks *"is this the item my predicate re-runs?"*, which is answerable
per-row, always current, and grades a well-shaped item from an **unknown** producer
correctly.

RFC_010 narrowing 1 supplies the sharpest supporting argument: *"a work-item type
with no automated consumer is not the kind of shared vocabulary whose guarantees
change when a producer is added"*. A registered verifier **is** an automated
consumer — so a verified item_type is exactly where convergence changes a guarantee,
and that must be mechanically enforced, not documented.

## Candidates not taken, and why

| candidate | disposition |
|---|---|
| Registry keyed on `(item_type, producer)` | **REJECTED.** Every producer-identity source fails. `created_by` is ruled out (it bottoms out at the literal `generic`, which carries 20+ item_types). `spec.audit_source` is optional and its absence proves nothing. A new mandatory producer field re-creates candidate 3 at the lookup: an unregistered pair must either complete (the bug, for novel producers) or fail closed against a code-side pair list that live config outruns. The refuted shape, one column over. |
| Creation-time guard | **PARTIALLY ABSORBED.** The build-time form ships in Half A as a test. A runtime creation guard is redundant once Half B lands — completion-side catches every channel including hand-filed rows, which creation-side cannot. |
| Consume `spec.acceptance_test` at completion | **FOLLOW-ON**, as the future verifier for `dark_section_audit` via the `criteria_check` vocabulary (RFC_002). Not primary: it needs an LLM/browser evaluation on the completion path, which this estate deliberately keeps free even of HTTP probes, and it covers only items carrying the field. |
| Re-check the 11 by hand | **REMEDIATION, not the fix.** Planned, but "operators must remember the verifier does not mean what the item says" is the defect, not a repair. |

## Ordering (Go is inert until a roll; DB config is live immediately)

1. Commit Half A (Go + coverage entries + register entry). Nothing changes live.
2. Commit Half B with its register entry in the same commit. Submit to the council
   gate before/alongside, per the 2026-07-29 ruling ("review here is after the fact,
   by design").
3. Image roll.
4. Migration `374`, applied **after** the roll: defensively re-type any still-open
   producer-B rows (0 today, but audits run between commit and roll), then the
   remediation inserts gated on grading.

A pre-roll-filed B row reaching completion post-roll before the sweep runs is caught
by Half B — blocked loudly as `verifier_scope_mismatch`, not silently closed.

## Corrections to the originating bug file

> **CORRECTED 2026-08-10:** the bug file's §4 table reads `complete`: producer B = 7.
> That was true on filing; it is **11** today, and four more closed clean in the three
> days the file sat open. The bug file's own warning stands and is reinforced — a
> count of `complete` is not a count of false-completes, and each of the 11 still
> needs grading against its own `acceptance_test`.
