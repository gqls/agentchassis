# REVIEW 2026-08-26 — `billing_orders.external_reference` as a cross-subsystem join key

Written at the council `architecture` seat's advisory request (corr `aa5a40a2`,
APPROVED round 1): the reasoning below crossed three top-level packages and lived
only in migration comments and a submission's risks field. This is the short
blast-radius + rollback account, in one place humans read.

## What the key is

The chat bot on the webdesign.uk box mints `BR-XXXXXX` when it stores a
visitor's approved brief (CHAT-011). The customer quotes it at payment; the
owner's dashboard create-order call records it in
`billing_orders.external_reference` (migration 659, APPLIED 2026-08-26); the
`collect_external_orders` action (PAY-010) releases a brief into `build_queue`
only when a PAID billing row carries its reference. **Owner ruling 2026-08-26:
the reference joins payment to brief; the brief itself never enters billing,
because briefs change and the reference does not.**

## The chosen semantics, and why each is deliberate

- **Non-unique.** Billing's own recovery bias (service.go: a checkout that
  fails after order creation is retried as a NEW order, never re-typing a
  voucher) means two rows legitimately share a reference. A UNIQUE constraint
  would turn that recovery into an error at the worst moment.
- **Newest-paid-wins.** The collector's gate is `status='paid' ORDER BY paid_at
  DESC LIMIT 1`. With non-unique references this is the only deterministic
  read, and any paid row is sufficient licence — the amount was computed
  server-side either way.
- **Typo-silent, visibly.** A mistyped reference at order creation joins
  nothing. The failure mode is a brief that stays uncollected and re-listed
  every 15 minutes, counted in the collector's `awaiting_payment` output —
  visible, recoverable (fix the order row), and never a wrong build.

## Blast radius (enumerated, not asserted)

- **Writers of the column:** exactly one — `Repository.CreateOrder`
  (`INSERT ... NULLIF($4,'')`). The webhook path (`ProcessEvent`) never touches
  it. Callers of the exported `CreateOrder` chain: `HandleCreateOrder` only —
  verified by grep across `internal/ cmd/` and by the whole repo compiling at
  HEAD `d2e8cfded` (`verify-head-builds.sh` OK) after the signature change.
- **Readers:** `ListOrders` (display) and the collector's paid-gate SELECT.
  Nothing else names the column as of 2026-08-26 (grep across the repo).
- **Deploy order (staged, both directions stated):** column before binary —
  the pre-659 binary ignores the column; the post-659 binary's
  `INSERT ... RETURNING` names it, so 659 had to apply first and did. Rollback
  is the mirror: binary back first, then `659_ROLLBACK` — dropping the column
  under the new binary breaks every order creation.

## The acknowledged gap: refund timing

The gate reads `status='paid'` only. Refunds are DECISION_2026-08-25 Option A
(webhook-as-truth, unbuilt): until that lands, a refund issued between payment
and collection would still release the build. The window is one polling
interval (≤15 min) and the schedule ships DISABLED (661's verify asserts it),
so today the gap is theoretical. **It must be re-examined when the refunds
build lands** — the natural close is the collector also refusing references
whose newest event is `charge.refunded`, which is one predicate on the same
query. Recorded here so the refunds lane inherits it explicitly; also in
PAY-010's risks and the 661 header.

## Rollback of the whole programme

Every stage is independently reversible and idempotent: disable/delete the
scheduled task (661_ROLLBACK — safe mid-flight, every collector branch is
idempotent and unacked briefs simply stay on the box); drop the column
(659_ROLLBACK, binary-first ordering above); the box half keeps storing briefs
regardless, losing only the automated collection.
