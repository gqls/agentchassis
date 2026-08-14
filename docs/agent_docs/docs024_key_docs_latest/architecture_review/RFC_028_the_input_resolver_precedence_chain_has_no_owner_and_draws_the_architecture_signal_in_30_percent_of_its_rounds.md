# RFC 028 — `ExtractActionInputs`' precedence chain is a shared contract with no owner, and it draws the architecture signal in ~30% of the rounds that touch it

## STATUS: **OPEN — routed here from a council REVISE, not raised speculatively.**

Filed 2026-08-14 by the `bugfix_209/231` lane, at the architecture seat's own direction.
The seat raised `ARCHITECTURE_SIGNAL: needs_rfc | DEFLECTIONS: unknown` against corr
`41a01378-1211-4987-966d-f8b6e2fddce1` (bugs_open/231 candidate 2, shipped as `d3edb5b89`),
and its objection was **not** that the change is wrong:

> Cost of NOT changing is real and quantified (99 dead config entries, 21 silently diverging
> from stated intent), so I am not objecting to shipping the fix — only to shipping a
> resolver-precedence change without the RFC's staged-rollout/rollback record, given it is
> inert only until the next chassis roll and then live everywhere at once.

**The seat is right on the classification.** By the owner's ruling of 2026-07-29 §1 — an
addition to a shared vocabulary needs an RFC when it changes what the shared mechanism
**GUARANTEES** — candidate 2 is unambiguously architecture-scope: it changes the precedence
rule of the resolver every action in the fleet uses to build its inputs. Not additive-and-inert;
a changed guarantee. So this file exists rather than an argument that it should not.

## 1. What is already settled, and must not be re-litigated here

- **The owner ruled to ship candidate 2**, explicitly, on 2026-08-11 (ruling 2 of three:
  *"Candidate 2: SHIP. 'An explicit config value beats a default' becomes the resolver's
  rule."*). The human break the 2026-07-28 ruling calls for — *"route the seam to
  architecture review on its own merits, and let a human break it"* — **already happened,
  and it happened BEFORE submission, not after a veto.** This RFC records a seam the owner
  chose; it does not reopen the choice.
- **There is no staged rollout available, and that too is an owner ruling, not an omission.**
  The 2026-07-29 ruling retired the ordering exemption's condition (1) precisely because no
  thread on this tree can hold a change back — HEAD is shared, `make build-*` builds from
  committed HEAD, any other session's roll ships your commit — and it went on to rule that
  **we will NOT require a default-OFF switch**, because its cost is a mechanism rotting
  unexercised. So "inert until the roll, then live everywhere at once" is the estate's
  standing deployment model for Go changes, not a gap this change introduced. A reviewer
  asking for a staged rollout here is asking for something the owner has declined twice.

What is genuinely NOT yet recorded, and is below: the **rollback** half, and the
**accumulation** problem that is this RFC's actual subject.

## 2. The measurement the seat asked for and could not make

The seat's `missing` item: *"No count of how many times `platform/orchestration/datahelpers/action_inputs.go`
or `ExtractActionInputs` has previously been sent to a higher layer by council review — check
would clarify whether this file is a repeat deflection site."* It printed `DEFLECTIONS: unknown`.

`[MEASURED 2026-08-14]` — every `fix_plan` artifact naming the file or the function, joined to
its round's final `council_report`:

| | count |
|---|---|
| distinct council rounds touching the resolver | **27** |
| final verdict `approved` | 21 |
| final verdict `revise` | 6 |
| ever `rejected` (guardian veto) | **1** |
| **rounds where a seat raised `needs_rfc`** | **8** |
| fix plans naming it, of 745 fleet-wide | 44 |

```sql
WITH touched AS (
  SELECT DISTINCT p.correlation_id FROM diagnosis_artifacts p WHERE p.kind='fix_plan'
    AND (p.body LIKE '%action_inputs.go%' OR p.body LIKE '%ExtractActionInputs%')
)
SELECT (SELECT count(*) FROM touched) AS distinct_rounds,
       (SELECT count(DISTINCT r.correlation_id) FROM diagnosis_artifacts r JOIN touched t USING (correlation_id)
          WHERE r.kind='council_report' AND r.body LIKE '%needs_rfc%') AS rounds_flagged_needs_rfc;
```

**The answer is the finding: it IS a repeat architecture-signal site — 8 of 27 rounds, ~30%.**
That is not a file being deflected once by an over-eager seat. It is a shared contract that
roughly one round in three notices has no architectural owner, and then ships anyway because
each individual change is well-measured and the cost of blocking it is real. Candidate 2 is
the 28th round and the 9th signal.

⚠ **The 8 is a floor, not a total.** It counts rounds whose report text contains `needs_rfc`,
so it sees the architecture seat's structured signal and misses any round where another seat
made the same point in prose. It also cannot see rounds before the seat existed, or the
submissions this estate's threads never made.

## 3. The rollback record the seat asked for

For candidate 2 specifically (`d3edb5b89` + revision `14e4333f7`):

- **Rollback = one revert commit plus a roll.** No migration, no DB config, no wire format, no
  work-item shape. `git revert` of the two commits restores the previous precedence exactly,
  and the offline detector's re-spec reverts with them — which matters, because a half-revert
  (resolver back, detector forward) would leave the checker describing a resolver that no
  longer exists. **They must revert together, and the reason they can is that they shipped
  together.**
- **Detection that you need to roll back:** `scripts/audit-default-shadowed-keys.sh` moving off
  `0 dead`, or a `Strategy 6: config value's type differs from the spec default's` Warn
  appearing anywhere in the fleet. No live entry mismatches kinds today, so that line firing
  means new config arrived and is being refused.
- **What rollback would cost:** the 99 config entries the fleet now honours go inert again, 21
  of them silently diverging from their stated intent. The four `audit_source` entries are NOT
  among them — `bugs_open/264` repaired that instance independently, by making the field
  Required with no Default, so it does not depend on this change.
- **What rollback cannot undo:** nothing. The change writes no data and migrates nothing.

## 4. The real subject: an eighth precedence arm and no owner

The seat's second objection is the one worth an architecture decision:

> Strategy 3's bridge now also beats a Default via the same dot-discriminator Strategy 6 uses —
> two strategies now share an implicit rule instead of one, increasing the surface a future
> author must re-derive correctly when adding Strategy 7.

Correct, and it generalises. The chain is now **Strategy 0 → 1 → 2 → 3 → nested-object block →
4 → 5 → 6**, eight sequential arms, each opening with a has-value skip, and **two of them now
share a discriminator that is stated in a comment rather than expressed once in code**: a
string containing a dot is a reference, never a value. That discriminator is load-bearing —
it is the guard `bugs_open/248` finding (a) paid 150+ page-visible 404s to learn — and it now
has two call sites and no single definition.

**This is RFC_022's argument, one level down.** The owner ruled there that the trigger for
architecture review should move from "any new opt-in field" to the **accumulated count**,
because *"ten individually inert opt-in fields are a shared action nobody understands, and
this trigger was the only thing that would have noticed the tenth."* Precedence arms accumulate
the same way, and by the same mechanism: each addition is individually well-argued (this one
has a fleet census and a demand control behind it), and the eighth is still an eighth.

RFC_022's counter is built (register WFA-013) and counts optional KEYS per action. Nothing
counts **resolution arms**, and nothing owns the question of whether eight is too many.

## 5. What this RFC asks the owner to decide

1. **Does the resolver's precedence chain get an owner and a stated contract?** Today its
   guarantees live in five places: the strategy comments in `action_inputs.go`, the
   `ActionInputSpec` field docs, `defaultshadow.go`'s class list, CTS-059, and a LANDMINES
   entry. All of those are now consistent — that took a deliberate effort this round and
   nothing keeps them so.
2. **Should the dot-discriminator be expressed once in code** (a named predicate both Strategy
   3 and Strategy 6 call, and which a Strategy 7 author cannot miss) rather than twice in
   prose? Cheap, and it is the concrete half of the seat's objection.
3. **Is there an arm budget**, in RFC_022's sense — a count past which adding a resolution arm
   requires an RFC rather than a council round? The measurement in §2 suggests the estate has
   been answering "no" implicitly, 27 times, while ~30% of those rounds said it should be "yes".

## 6. Deliberately not proposed here

- **Reverting candidate 2.** The owner ruled to ship it; §3 records how, not whether.
- **A default-OFF switch for it.** Declined by the 2026-07-29 ruling as a class.
- **The remaining `dotted_conditional` exposure** (96 live entries whose dotted path falls back
  to its Default silently when it does not resolve). That is `bugs_open/231`'s open half and
  belongs there: resolvability is a runtime fact, so it is a detection question, not a
  precedence one.
- **CTS-059's open review question** — whether a *resolving* dotless string on a defaulted
  field should resolve as a `collected_data` reference instead of being taken literally. Named
  at registration, zero live entries want it, and it is a precedence question this RFC's
  answers should govern rather than pre-empt.

## Sources

- Council round: corr `41a01378-1211-4987-966d-f8b6e2fddce1`, verdict REVISE, 11 seats, 6
  abstained, no truncation, decided by a gating objection from `guardian`. Architecture seat's
  full text quoted above.
- Code: `platform/orchestration/datahelpers/action_inputs.go` (Strategy 6, `LiteralKind`, the
  Strategy 3 bridge), `platform/orchestration/datahelpers/action_inputs_strategy6_test.go`
  (the invariant, with its own vacuity control), `cmd/config-key-audit/defaultshadow.go`.
- Commits: `d3edb5b89` (the seam), `14e4333f7` (the REVISE round's four code fixes).
- Register: **CTS-059**, with the landmine and the open review question.
- Prior rulings this RFC rests on: 2026-07-28 (platform seams and the ordering exemption),
  2026-07-29 §1 and §2 (the guarantee test; condition (1) retired; no default-OFF requirement),
  2026-08-02 §2 (opt-in fields on shared seams), 2026-08-11 (RFC_022's narrowing, and ruling 2
  which shipped this change).
