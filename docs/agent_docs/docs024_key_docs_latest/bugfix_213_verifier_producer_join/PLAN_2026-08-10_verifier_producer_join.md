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

---

## OWNER DECISIONS 2026-08-11 — three rulings, taken after the fix went live

Put to the owner once the instance fix was pod-proven, because all three are about
what the fix does NOT cover. Recorded here as decisions with their reasoning, per the
PLAN doc's remit.

### D1 — `dark_section_audit`'s gap: **BUILD THE VERIFIER, AND SETTLE THE ROUTING**

The problem put to the owner: Half A stops design-audit items being graded by the
*wrong* predicate, but the new type has **no verifier of its own**, so those items now
close **ungraded** rather than **mis-graded**. Both close clean. Half A moved the
failure for that half rather than removing it — two council seats
(`bug_historian`, `improvement_guardian`) said so independently, and they were right.

**Ruled: do both halves of the remedy.**

1. **A verifier over `spec.acceptance_test`**, via the `criteria_check` vocabulary
   (RFC_002). This makes that field load-bearing for the first time — it is currently
   read by **nothing** on the completion path (grepped: zero consumers outside the
   `improve_tool` family). All 11 live acceptance tests are mechanical and
   browser-checkable (computed background-colour, contrast ratio, absence of an inline
   `<style>`), so the vocabulary fits without inventing anything.
   - ⚠ **The hard constraint, and it is a real one:** this puts a browser/computed-style
     evaluation on the **completion path**, which this estate has deliberately kept free
     even of HTTP probes (`verifier_coverage_test.go:171` records the standing objection).
     Do not smuggle that in as an implementation detail — it is the design question, it
     needs its own council round, and `contrast_failure` is the precedent to argue from
     (same posture, same browser dependency, currently answered by re-detection rather
     than by a verifier).
2. **Settle the handler routing.** `designRouting["dark_section"]` still points at
   `color-variable-fixer`. **Do not assume that is wrong for all 11** — see the
   correction below.

> **CORRECTION 2026-08-11 to this lane's own earlier claim.** The bug file, WII-013 and
> my own commit messages all say the fixer "provably cannot touch producer B's typical
> defect". That is **verified for exactly one item** — gamesdesign's
> `var(--color-primary, #1a1a2e)`, an already-`var()` fallback outside
> `ReplaceHardcodedColors`' remit, which is *why* it passed. Reading the other ten
> acceptance tests, several name inline `style` attributes and `rgba(0,0,0` literals
> that may well be **inside** that remit. So "the handler cannot fix these" is
> `[UNVERIFIED]` as a generalisation and must be measured per item before the routing
> is changed. Generalising from the one worked instance is precisely the move this bug
> exists to punish.

### D2 — the 11 mis-closed items: **LET THE ROTATION RE-DETECT**

**Ruled: do nothing bespoke.** Justified by a measurement taken the same day, which
also corrects this lane's earlier assumption:

> **CORRECTION 2026-08-11.** The handoff written yesterday implies re-detection is
> doubtful, echoing `bugs_open/093`'s finding that `improvement-sweep` is disabled
> (`enabled=f`, last fired 2026-05-02 — still true). **But `bugs_open/230`'s rotation
> supersedes that for this producer.** `site_discovery_rotation` carries
> **`design-discovery-agent` across 22 sites, last selected 2026-08-10** — live and
> driving. So any still-present defect re-files itself on a ≤7-day period, now under
> `dark_section_audit`, with a fresh dedup key (different `item_type` ⇒ no collision
> with the old rows). Re-detection is real, not hypothetical.

Consequences accepted with the ruling:
- The historical `complete` rows stay as they are — they are the honest record of what
  the machine did.
- **This is only sound once D1 lands.** Until `dark_section_audit` has a verifier, a
  re-detected item closes unverified, so the rotation regenerates the finding and then
  loses it again. **D2 depends on D1; it is not independent.** If D1 slips, revisit.
- We therefore never get a false-complete *count* for the original 11. That is
  acceptable and should be stated rather than quietly dropped: "11 closed" remains an
  upper bound on the damage, with 1 confirmed broken and 1 confirmed clean.

### D3 — the class-level hole: **BUILD THE DETECTOR**

**Ruled: build it.** The `architecture` seat's non-blocking objection, quoted because
it is the sharpest framing available: *"`VerifierPolicy.Grades` is opt-in — the NEXT
converging producer on any of the other 10 verified item_types reproduces this exact
bug unless a human remembers to write a Grades function, which is precisely the
discipline that already failed once here."*

Shape (the seat's own suggestion, and it needs no new mechanism): a periodic check
flagging any **verified** `item_type` accumulating rows with more than one
spec-shape / `audit_source` and **no** `Grades`. The query already exists in this
lane's RUNBOOK as the fleet check; the work is turning it into a scheduled check that
files a work item rather than a query a session has to remember to run.

Design notes for whoever builds it:
- **Do not key it on `created_by`** (bottoms out at `generic`) or on a producer list
  (refuted — `bugs_open/213` §5.3). Key it on the **spec shape**, the same principle
  as `Grades` itself.
- The predicate is disconfirmable today: it returns 2 for `hardcoded_section_colors`
  and 1 for every other verified type, so it can distinguish and is not vacuous.
- It must not fire on the type it is about to fix — once `Grades` is registered for a
  type, that type is answered and should drop out of the finding.

---

## D3 — BUILT 2026-08-11. What was decided while building it, and why

The ruling said "key it on the **spec shape**". Implementing that turned up two
design decisions the ruling could not have contained, both settled by measurement
rather than by argument:

1. **A distinct-key-set COUNT is not a producer count.** [MEASURED 2026-08-11] it
   reads 2 on four single-producer types (`empty_section`, `literal_markdown`,
   `page_canonical_collision`, `truncated_component`) because producers add and drop
   optional keys over their life. So the shape axis has to CLUSTER, and the
   clustering — not the shape — is the load-bearing part of the detector.
2. **The cluster rule is the overlap coefficient, not Jaccard.**
   `page_canonical_collision`'s two real shapes are 11 keys and 3 sharing 2: J=0.167
   (a phantom second producer) against an overlap of 0.667. The threshold of 0.5 is
   not tuned — every same-producer pair in the fleet scores ≥0.667 and the one true
   cross-producer pair scores 0.000, so 0.5 sits in an empty band. **If that band
   ever closes, the threshold stops being defensible; it is a thing to re-measure,
   not to inherit.**

Rejected axes, each measured on live data and each firing on a type with exactly
one producer: `created_by` (2–3 values), the `source` column (2 on
`page_canonical_collision`), and the VALUE of `spec.check` (2 on the same). The
last was considered as a defence against the council's `editquality` objection and
would have made the detector worse.

**Where it runs, and why not inside `discovery_checks`** (the `reuse_agent` seat's
objection, answered with three checkable facts rather than with precedent):
invocation there is gated on live agent config (`enabledChecks` — a new check file
is inert until `agent_definitions` names it, the exact inert-by-omission failure
this kind of detector exists to avoid); retraction is site-scoped
(`resolveWorkItems(…, dctx.SiteID, …)`); dedup is `(site_id, item_key)`. A
fleet-level finding has no site to be scoped to. It is therefore a daily CronJob Go
image — Go, not Python, because both halves of its question (who is registered, who
declares a remit) are compiled-in state, and a mirrored list would go stale exactly
when a new verifier lands.

**What it does NOT close, on the record:** two producers that share an
`audit_source` label AND overlap ≥50% of their spec keys merge into one family and
are invisible to it; so is a convergence that has not yet filed a row. No
row-shaped test can see either, and a producer-list test is refuted
(`bugs_open/213` §5.3). This narrows the class; it does not close it.

### D1 got sharper while D3 was being built — and not by argument

[MEASURED 2026-08-11] `dark_section_audit` already carries **14 rows, all created
today, 13 already `complete`.** The rotation re-detected within a day of the roll,
exactly as D2 assumed. But the type still has no verifier, so those 13 closed
**ungraded**. D2's stated dependency on D1 is therefore no longer hypothetical:
at the current rate the machine re-finds these defects and loses them again on a
≤7-day cycle, ~13 items at a time. That is the strongest argument for D1 to date
and it was not visible when the rulings were taken.
