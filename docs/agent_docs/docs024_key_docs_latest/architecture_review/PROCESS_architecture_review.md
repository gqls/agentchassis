# The architecture-review track (RFC process)

**Created 2026-07-25 by owner decision**, after the council gate's guardian
vetoed the bug-003 delivery-guarantee redesign twice — not on technical grounds
(those it conceded in round 2) but because "a coordinated rewrite of the
fleet's delivery guarantee … across five-plus packages" is, by the gate's own
charter, an architecture change dressed as a fix, and the platform had no
review track sized for one. This document is that track. It is deliberately
lightweight: one file per RFC, the owner as the deciding authority, the
existing council gate reused for what it is good at.

## When a change needs an RFC (the trigger test)

Taken from the guardian's charter language, which two verdicts have now
exercised. A change needs an RFC when ANY of these hold:

- it changes a **shared contract**: a dedupe key, a delivery guarantee, a
  state machine consumed by more than one package, a wire/message shape;
- it changes or removes an **exported symbol** other packages depend on
  (signature changes count);
- it lands coordinated edits across **many packages at once** (rule of thumb:
  three or more of `platform/*`'s top-level packages);
- it needs a **staged or reversible rollout** to be safe (the change and its
  verification cannot both fit in one deploy step).

A point fix that happens to be large does not need an RFC; a small change
that rewrites a contract does. When in doubt, the cost of an RFC is one
document — write it.

## What an RFC contains

One markdown file, `RFC_NNN_<slug>.md`, in this directory. Status line at the
top: `DRAFT` → `RATIFIED` (owner) → `IMPLEMENTED` → (possibly) `RETIRED`.
Sections, all required:

1. **Problem + evidence** — the defect or need, with live figures and
   file:line citations. Point at the bug file; don't fork it.
2. **Design** — what changes, per package, with the load-bearing mechanics
   (the two or three queries/orderings everything rests on).
3. **Alternatives considered** — including the do-less option, each with the
   evidence that ruled it out. An alternative dismissed without evidence is
   an objection waiting to happen.
4. **Blast radius, named** — derived mechanically (`go list -deps` per cmd
   target, compile-proof for removals), not qualitatively. Name the binaries
   whose behaviour changes, and the ones that merely relink.
5. **Staged rollout plan** — the order things ship, what is watched at each
   stage, and the induced-fault tests (not just happy-path greps). Name the
   canary if there is one.
6. **Rollback plan** — what undoes each stage; schema must tolerate the
   previous binary (image-first rollback), or say loudly why not.
7. **Acceptance evidence** — the measurements that will retire the RFC's
   risk: pod-greps of created literals, behavioural probes, week-later stats.

## The flow

1. Write the RFC (status DRAFT). Commit it.
2. **Owner ratifies** — the owner is the architecture authority; a ratified
   RFC records the decision the way PLAN decisions (D1, D2…) already do.
   Objections and revisions happen in the file, visibly, like any working doc.
3. Optionally, exercise the **council gate** on the RFC-shaped plan (097; the
   `plan.edits` name the real files, the rationale points at the RFC). The
   gate remains advisory; its seats are good at catching what one author
   missed. A guardian veto on a ratified RFC is input for the owner, not a
   block.
4. Implement **in the RFC's stages**. Each stage that fits the point-fix
   shape goes through the normal council gate as usual.
5. When acceptance evidence is in, mark the RFC IMPLEMENTED, with the
   evidence inline or pointed at.

## Numbering and index

- `RFC_001_at_least_once_delivery.md` — bug-003 delivery-guarantee redesign
  (the case that created this track).

Claim the next number by adding a line here in the same commit as the RFC —
the same collision discipline as migrations, and this list is the ledger.
