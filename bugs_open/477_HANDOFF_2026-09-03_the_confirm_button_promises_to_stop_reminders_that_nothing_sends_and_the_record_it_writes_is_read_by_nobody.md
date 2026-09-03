# 477 — the confirm button promises to stop reminders that nothing sends, and the record it writes is read by nobody

Filed 2026-09-03 by `site_delivery_and_editor`, from the owner asking a design question — *"maybe
repeat them in the first follow up email that would normally go a week or so later"* — which turned
out to name a mechanism that does not exist. The question was the check.

## The three facts, measured

**Nothing can send a follow-up.**

| what | count |
|---|---|
| agents that send mail | **1** (`delivery-email-sender`) |
| scheduled tasks targeting any mail-capable agent | **0** |
| Go actions that can send mail | **1** (`send_delivery_email_action.go`) |

**The customer-facing page promises otherwise, twice.** `internal/core-manager/handlers/delivery.go`,
on the button page (`:428`) and again on the success page (`:409`):

> *"Pressing the button below tells us you have moved everything across. **You will not get any more
> reminders about it.**"*

**And the record it writes is inert.** `ConfirmTransfer` redeems the token and stamps
`sites.transfer_confirmed_at` (`handover.go:340-352`). That column, and the `TransferConfirmed` field
built from it, are read **only inside `handover.go` itself** — it is populated into a struct and
nothing outside that file ever consults it. Repo-wide, excluding the file that writes it: zero
readers.

`[MEASURED 2026-09-03]` 60 sites, 1 handed over, **0 confirmed** — nobody has pressed it yet, so
nothing has depended on it so far.

## Why this is worse than the other two of today's copy defects

`bugs_open/475` promises instructions inside a ZIP that has none. This one asks the customer to
**take an action**, tells them what that action prevents, and the thing it prevents does not exist.
They press a button, get a page saying the reminders have stopped, and the truth is that no reminder
was ever going to arrive and no part of the system now behaves differently.

It is also the third instance today of one shape: **customer-facing copy promising a mechanism the
code does not have.** 475 (instructions in the ZIP), 476 (open `index.html` and the site works),
and this. All three were found by a person reading the words and asking whether they were true.
Nothing automated found any of them, and nothing automated could have — the estate has no check that
compares a promise to an implementation.

## What the owner actually asked for, and how it resolves this

He proposed a follow-up email about a week after delivery, repeating the hosting instructions. That
is a good idea on its own merits — the moment a customer needs the instructions is not the moment
they receive them, it is when the free hosting period starts running out.

**And it is the same piece of work as this bug.** A follow-up that exists is a reminder that exists,
and `transfer_confirmed_at` is exactly the right thing to suppress it. Build the follow-up and the
confirm button becomes true; build neither and the button stays a lie; build the button's honesty
alone and there is nothing to suppress.

## Fix candidates, ordered by what closes the door

1. **A scheduled follow-up sender, suppressed by `transfer_confirmed_at`.** Selects sites where
   `handed_over_at` is older than N days, `transfer_confirmed_at IS NULL`, and no follow-up has
   already gone; sends a second email repeating the instructions link and the domain options. Makes
   the existing promise true and gives the confirm button its first consumer. **Needs:** a new agent
   (config, following `delivery-email-sender`'s shape), a `scheduled_tasks` row, and a
   `followup_sent_at` column or work-item to stop it repeating for ever. The mail action already
   exists and is already config-driven, so this is mostly seeding rather than Go.
   ⚠ **A scheduled thing that emails customers is the one to get wrong quietly.** It needs the same
   once-only discipline the handover stamp has, or a scheduler glitch mails somebody nightly.
2. **Change the copy to promise only what is true.** One migration, no new machinery: the button says
   what it records, not what it stops. Honest immediately, and it throws away the owner's idea.
   Worth doing ONLY if (1) is not going to happen soon, because two copy edits are cheaper than one
   wrong promise left standing for a month.
3. **Both, in order** — (2) today so nothing false is live while (1) is built, then (1), then revert
   the copy to the stronger wording once there is something to suppress. **Recommended.** It costs
   one extra migration and means no customer is ever told something untrue.
4. Weakest: leave it, on the grounds that nobody has pressed it. Rejected — `transfer_confirmed_at`
   is 0 of 60 because exactly one site has ever been delivered, and that was tonight.

## How to verify a fix

Deliver to a test address, wait out the follow-up interval, and confirm a second email arrives. Then
press the confirm button on another test site and confirm the follow-up does **not** arrive. The
second half is the one that matters and the one a happy-path test omits: today's negative control is
that pressing the button changes nothing at all.

## Related

- `bugs_open/475` — the delivery email promising instructions the ZIP does not contain.
- `bugs_open/476` — the instructions promising a local double-click that does not work.
- `platform/delivery/handover.go:340` — `ConfirmTransfer`, the write with no reader.
- `internal/core-manager/handlers/delivery.go:409,428` — the two places the promise is made.
