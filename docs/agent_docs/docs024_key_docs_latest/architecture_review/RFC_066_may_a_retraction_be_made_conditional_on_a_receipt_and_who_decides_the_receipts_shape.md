# RFC_066 — may a retraction be made CONDITIONAL on a receipt, and who decides that receipt's shape before a second consumer arrives?

**Status: OPEN — filed at the architecture seat's request.**
Raised 2026-09-03 by session "bugs_open/469" after council round 1 on `009fabca`
**APPROVED** the change (13 reviewers, 4 advisory objections, none high) while the
`architecture` seat returned `ARCHITECTURE_SIGNAL: needs_rfc` at MEDIUM.

**The seat's objection, quoted, because it is the whole reason this file exists:**

> *"Receipt/Evidence on ResolvedFinding is a new shared contract for ALL discovery checks
> (author explicitly frames it as reusable for the other nine no-closer checks), decided
> inside a single-symptom bug fix rather than written up with blast radius across those
> future adopters."*

It is right, and I had handed it the call rather than claiming the RFC_022 exception.
**The code has shipped** (`fc9cad600`, inert until the next roll) — this estate reviews
after the fact by design (owner ruling 2026-07-29 §2: nothing on a shared HEAD can hold a
change back), so the question is not whether to revert but what the contract should be
before a second consumer exists.

---

## 1. What the thing IS, in plain terms

A **discovery check** is a rule that looks at a site and files a work item when something is
wrong. There are 71 of them.

A **retraction** is a check closing one of its own items because it looked again and the
problem is gone. Only 19 of the 71 can do this; the rest file and never close.

The retraction machinery has one safety rule, written down since it was built: **close only
on a positive observation** — never because you found nothing, since a check that was broken
and a site that is healthy both produce "nothing".

**That rule is not enough for one shape of check, and that is what this RFC is about.** If a
check's finding is *"two stores disagree"*, then the finding stopping and the damage
completing can be **the same event**: the stores agree again precisely because one of them
overwrote the other. A closer obeying the rule to the letter would still, correctly and
positively, certify a destroyed page as resolved.

## 2. The rule that follows, and what shipped

**A check may retract a finding whose resolution DESTROYED something only by recording what
was destroyed, in the same transaction.**

Shipped as two opt-in fields on `ResolvedFinding` (nil = today's behaviour for all 19
existing retractors) with the enforcement inside `resolveWorkItems`, the one implementation
both callers share:

- `Receipt *WorkItemSpec` — written FIRST, same transaction; the retraction is **WITHHELD**
  if it can neither be inserted nor confirmed present.
- `Evidence map[string]interface{}` — nested at `result->'resolution_evidence'`.

## 3. How the current case measures against it

One consumer, in the same commit: `check_section_source_drift` (`bugs_open/469`). The
motivating damage is real — `robot-hands.com/gripper-catalog` lost a section a human
deliberately added, and the warning about it sat open for 37 days and then closed looking
resolved.

**What makes this architecture-scope rather than one check's business** is that I said, in
the submission, that the seam is reusable for the other **nine** flag-only checks that have
no closer. The seat took me at my word, and should have.

## 4. What the owner is actually being asked

**Q1. Is "a retraction may be made conditional on a receipt" the right shared contract, or
should the coupling live per-check?**
Recommended: **shared.** `resolveWorkItems` has exactly **2** non-test callers
(`discovery_checks.go`, `work_item_retraction.go` — verified at source, not assumed), and a
control in either protects only that one. Per-check means the tenth author must remember.

**Q2. Should the receipt's SHAPE be fixed now, or left free until a second consumer arrives?**
This is the one I genuinely do not know, and it is the seat's real point.

- Today `Receipt` is a whole `WorkItemSpec`, so each adopting check invents its own item
  type, key shape and spec keys. That is maximum freedom and maximum divergence: ten
  adopters could produce ten incompatible records of "something was destroyed".
- The alternative is a **typed** receipt — a fixed `{what_was_lost[], evidence_copied,
  recovery_pointer}` shape the seam enforces — which is answerable now and constrains a
  design nobody has needed yet.

Recommended: **leave it free, and revisit at the SECOND adopter**, on the explicit condition
that this RFC is the place it gets revisited. Costed honestly: the cost of leaving it free is
exactly the accumulation RFC_022's own budget mechanism exists to catch, and I would rather
that be a stated debt with a named trigger than a silent one.

**Q3. Does adopting this seam on any of the nine remaining no-closer checks require coming
back here, or is the contract now settled enough to use?**
Recommended: **the SECOND adopter comes back**, the third onwards does not. One adopter is a
worked example; two is a pattern, and two is when the shape question in Q2 becomes
answerable from evidence rather than from taste.

## 5. Blast radius across the nine candidate adopters, since the seat asked for it

`[MEASURED 2026-09-03]` 19 of 71 checks populate `Resolved`; 18 file at least one flag-only
item and **10 of those have no closer at all** — this change is the first. The other nine:
`backend_entry_orphaned`, `decision_guards`, `content_duplication`,
`contact_form_undeliverable`, `integrity`, `image_source_unsatisfiable`, `palette_contrast`,
`voice_tells`, `unverified_claims`.

**Not all nine want a receipt, and that is the important part of the answer.** A receipt is
only meaningful where the resolution can be DESTRUCTIVE — where "the finding stopped
reproducing" has a branch that means "somebody's work is gone". Of the nine, the ones whose
findings could resolve by destruction rather than by repair are the divergence-shaped ones;
a `palette_contrast` finding that stops reproducing means the contrast was fixed, and there
is nothing to record. **So the honest blast radius is smaller than "nine", and nobody has
enumerated which — that enumeration is work this RFC does not do and should not pretend to.**

## 6. What I am NOT asking

Not asking to revert: the change is APPROVED, mutation-proved (18 mutations run, each
killing a named test, sources restored and diffed byte-identical), and every other seat
approved or objected only advisorily. Not asking to widen it: no second adopter is proposed
here.

## 7. Relations

- `bugs_open/469` — the motivating damage and the closer
- Council `009fabca-71c8-4f7b-9b23-f0b6605eb531` — APPROVED round 1, this seat's objection
- Register **WII-039** (the seam) / **WII-040** (the first receipt type)
- Owner ruling 2026-08-02 §2 (opt-in field, unsafe default OFF) — which this follows
- Owner ruling 2026-07-29 §1 (an addition needs an RFC when it changes what the shared
  mechanism GUARANTEES) — the test this arguably meets: it ADDS a guarantee rather than
  narrowing one, which is why the seat graded it MEDIUM rather than HIGH
- **RFC_022** — the optional-key budget, whose accumulation argument is exactly Q2's shape
- `RFC_064` (the `site_plan_sections` writer, the `427` lane's) — the WRITE half of the same
  bug; this is the RECORD half
