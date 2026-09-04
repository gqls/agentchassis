# PLAN 2026-09-04 — make the confirm button true

Lane opened 2026-09-04 to work `bugs_open/477`: the customer-facing confirm page promises
*"You will not get any more reminders about it"*, nothing in the estate sends a reminder, and
`sites.transfer_confirmed_at` — the record the button writes — has no reader outside the file that
writes it.

**Ownership.** `scripts/who-owns.py 477` reads **OWNED** off the filing commit alone
(`01f46c87a`, `site_delivery_and_editor`, 2026-09-03 22:16). I asked that lane's live session
directly before touching anything. Their answer: *"477 is yours. Not claimed, nobody is building it,
and no objection to candidate 3."* They have recorded the claim in their own
`HANDOFF_2026-09-03b_continue_here.md` §6b and are staying out of `delivery.go` and `handover.go`.

---

## 1. What is actually true, re-measured today (not carried from the bug file)

| claim | how checked | result |
|---|---|---|
| `transfer_confirmed_at` has no reader outside `handover.go` | `grep -rn "transfer_confirmed_at\|TransferConfirmed" --include=*.go .` | **CONFIRMED** — 6 hits, all in `platform/delivery/handover.go` (`:147` field, `:189`, `:215` populate, `:346` write). Zero in tests. |
| exactly one agent can send mail | `agent_definitions` where `default_config::text LIKE '%send_delivery_email%'`, live, non-snapshot, not deleted | **CONFIRMED** — `delivery-email-sender`, active |
| no scheduled task targets it | `SELECT count(*) FROM scheduled_tasks WHERE target_agent_type='delivery-email-sender'` | **CONFIRMED — 0** (of 107 rows) |
| 60 sites, 1 handed over, 0 confirmed | `SELECT count(*), count(handed_over_at), count(transfer_confirmed_at) FROM sites` | **CONFIRMED `[MEASURED 2026-09-04 11:43Z]`** — 60 / 1 / 0 |

`[MEASURED 2026-09-04]` The one handed-over site is **idea.uk**, from last night's owner-authorised
rehearsal. It is the only row a follow-up sender could select, and **its delivery address is the
owner's own** (`aaa@designconsultancy.co.uk`). See §5.

## 2. TWO CORRECTIONS to the bug file's own fix candidates

Both were found by reading the code the candidates rest on. Neither is a criticism of the filing —
the bug file is accurate about the defect; these are about the cost of the remedies.

> **CORRECTION 1 — candidate 1 is NOT "mostly seeding rather than Go".** The bug file says the mail
> action already exists and is config-driven, so a follow-up sender is mostly a seed. It is not.
> `SendDeliveryEmailAction` calls `delivery.Claim` (`send_delivery_email_action.go:156`), and `Claim`
> calls `StampHandover`, which claims the row with `WHERE handed_over_at IS NULL` and returns
> **`ErrAlreadyDelivered`** for any site already handed over (`platform/delivery/prepare.go:261-267`).
> Every site a follow-up would target is by definition already handed over, so **the existing action
> refuses, by design, exactly the population the follow-up exists for.** A follow-up sender needs its
> own Go action.

> **CORRECTION 2 — candidate 2 is NOT "one migration".** The bug file offers the honest-copy fix as
> *"One migration, no new machinery"*. The confirm-page copy is **hardcoded Go**:
> `internal/core-manager/handlers/delivery.go:405-432`, string literals inside `renderConfirm`. There
> is no DB row and no config key. So it is a Go edit, an image, and a **core-manager roll** — and
> `make release` is the owner's. (The *delivery email* copy is config, deliberately; the *confirm
> page* copy is not. Two surfaces, two mechanisms, and the bug file conflated them.)

The consequence for sequencing: candidate 3's first step is **not** the cheap one. It is inert until
a roll, exactly like `bugs_open/466`. That does not change the order — it changes what "today" buys.

## 3. The plan — candidate 3, in three steps

**Step A (this session): make the copy true by deleting the false clause.**
Remove *"You will not get any more reminders about it."* from both the button page
(`confirmButton`) and the success page (`confirmOK`). Invent nothing to replace it: the button's
stated motivation was the lie, and the smallest possible diff is also the easiest to revert at step
C. Add a test that fails if the promise returns while no sender exists.

**Step B: the follow-up sender.** A new action, a new agent, a `scheduled_tasks` row modelled on
`zip-link-refresh` (the estate's only site-selecting scheduled sender pattern), and a once-only
claim. Designed here, held for the owner — see §5.

**Step C: restore the stronger wording** once B is live, deleting the step-A tripwire test in the
same commit. The test is what makes C deliberate rather than accidental.

## 4. Step B design — the once-only claim is the whole risk

The lane that filed this put it best: *"a scheduled thing that emails customers is the one to get
wrong quietly."* `delivery-email-sender` is safe because the handover stamp claims the row. A
scheduled sender has **no equivalent unless one is built**.

- **Claim by UPDATE, not by SELECT-then-send.** Copy the pattern at `handover.go:157-205` and its
  comment: `UPDATE sites SET followup_sent_at = $2 WHERE id = $1 AND followup_sent_at IS NULL
  RETURNING …`, and send only if the row was claimed. The row can be claimed by at most one
  statement by construction. A timestamp comparison must take no part in the decision — that exact
  mistake is written up in `StampHandover`'s own comment.
- **Selection:** `handed_over_at < now() - interval 'N days'` AND `transfer_confirmed_at IS NULL`
  AND `followup_sent_at IS NULL`. This is what gives `transfer_confirmed_at` its first reader and is
  the half of the bug that makes the button mean something.
- **Copy is CONFIG**, following `send_delivery_email_action.go:8`, and it must **carry
  `{{instructions_link}}`, not restate the hosting steps**
  *(> **RENAMED 2026-09-04:** this said `{{instructions_url}}` when written. Agreed with the
  `bugs_open/475` lane: the estate's convention is a `*_url` CONFIG KEY feeding a `*_link`
  PLACEHOLDER — `send_delivery_email_action.go:127-129` does it three times — so the placeholder is
  `{{instructions_link}}` and the config key stays `instructions_url`. Done while `775` was still
  seeded-and-unapplied, when it cost nothing; after it is applied a rename costs a migration.)* — the rule from `bugs_open/475`: anything
  that can go out of date lives on the page, never in a copy the customer already holds. A follow-up
  email is a fourth exit for that same body of copy, not a new one.
- **Keep the promise-vs-artefact guard.** `send_delivery_email_action.go:125-132` refuses to send
  when the template names a `{{placeholder}}` this dispatch cannot fill. A hardcoded URL would
  escape it. Inherit it.
- **Recipient is an open question, not a detail.** `sites.email` is the site's *published* contact
  address, not the ordering customer's. The delivery email took `input_data.customer_email` from the
  dispatch. A scheduled sender has no dispatcher to supply it, so where the address comes from must
  be settled before anything sends. **[UNRESOLVED]**

## 5. What is held for the owner, and why

**He has not ruled.** All he said was *"maybe repeat them in the first follow up email that would
normally go a week or so later"* — a suggestion in passing, which is what prompted the bug. Do not
cite a ruling. Three things need his answer before B can send anything:

1. **The interval.** "a week or so" is not a number.
2. **The recipient rule** (§4, last bullet).
3. **⚠ The first working test WILL email him.** idea.uk is the only selectable row and its delivery
   address is his. He must not meet that unannounced.

So B ships **built and disabled**: the action, the column and the claim in code; the agent and the
`scheduled_tasks` row seeded with `enabled = false`. One flag, after he rules.

## 6. Gate and process

`internal/` and appliable migrations are both in council-gate scope. Submit **before** applying, not
after — the fair objection the delivery lane earned on `751` yesterday. Migrations: **rename first,
record second** (sessions race for numbers; recording then renumbering mints a ledger row naming a
file that no longer exists).
