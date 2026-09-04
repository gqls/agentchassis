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

---

# CONTRIB 2026-09-04 — taken up, half fixed, and three things this file got wrong

Added by the session that picked this up (`bugs_open/477` lane,
`docs/agent_docs/docs024_key_docs_latest/bugfix_477_delivery_followup/`). The filing lane
(`site_delivery_and_editor`) confirmed it was unclaimed before I started and has recorded the claim
in their own `HANDOFF_2026-09-03b_continue_here.md` §6b.

**Every figure in the original was re-measured `[MEASURED 2026-09-04 11:43Z]` and all four held**:
6 grep hits for `transfer_confirmed_at`, all in `handover.go`; 1 mail-capable agent; 0 scheduled
tasks targeting it; 60 sites / 1 handed over / 0 confirmed.

## 1. THERE WAS A THIRD SURFACE, and it is the only one a customer has actually read

This file measured two surfaces, both Go. The same false promise was also in the **delivery email
itself**, read at the live row rather than from the seed:

> `agent_definitions`, `delivery-email-sender`, `…->'send_email'->'config'->>'body_template'`:
> *"Once your site is off our hosting, press the button here **so we stop reminding you**"*

That one is **config**, so it was fixable the same day with no roll — and it is the one that reached
a human, because the idea.uk rehearsal email carried it. **FIXED AND LIVE**: the
`site_delivery_and_editor` lane took it on being told and shipped migration **`776`**, anchored on
the verbatim line and aborting if it had moved; `ILIKE '%stop reminding you%'` over live agent rows
now returns **0**. The paragraph was kept and only the promise replaced (*"press the button here to
tell us you have moved"*), because unlike the two pages that sentence had no other reason to press.

## 2. TWO OF THE FIX CANDIDATES WERE WRONG ABOUT THEIR OWN COST

> **Candidate 2 is NOT "one migration, no new machinery".** `renderConfirm`'s copy is hardcoded Go
> string literals (`internal/core-manager/handlers/delivery.go:405-432`). There is no DB row and no
> config key, so it needs an image and a **core-manager roll**, and `make release` is the owner's.
> The delivery *email's* copy IS config; the confirm *page's* is not. Two surfaces, two mechanisms.

> **Candidate 1 is NOT "mostly seeding rather than Go".** `SendDeliveryEmailAction` calls
> `delivery.Claim` → `StampHandover`, which claims `WHERE handed_over_at IS NULL` and returns
> `ErrAlreadyDelivered` otherwise (`platform/delivery/prepare.go:261-267`). Every site a follow-up
> targets is already handed over, so **the existing action refuses, by design, exactly the population
> the follow-up exists for.** A follow-up needs its own action and its own claim.

## 3. WHAT IS DONE

**Step A — the copy is honest.** Committed `76ec663d3`. The sentence deleted from both pages, not
reworded: the button's stated motivation was the false part, and a replacement motivation is the same
defect one wording along. Guarded by a tripwire test proven by mutation on **both arms separately**.
⚠ **INERT until a core-manager roll.**

**Step B — the follow-up sender, BUILT AND SEEDED DISABLED.** `delivery.ClaimFollowup` +
`send_followup_email` action + migrations `774` (the `sites.followup_sent_at` claim column) and `775`
(agent + schedule, `enabled = false`). The at-most-once property is the `UPDATE … WHERE
followup_sent_at IS NULL` claim, and **`transfer_confirmed_at IS NULL` is re-checked inside that same
statement** — which is where this bug actually closes: a customer who presses the button between the
scheduler's SELECT and the dispatch landing is not emailed, and only the UPDATE can promise that.
Proven against real Postgres in a rolled-back transaction, five cases plus a second-claim test, **with
a negative control** (drop the suppression predicate and the confirmed site claims — so the refusals
are the predicate's doing and not a broken fixture).

## 4. ⚠ AND IT CANNOT SEND TO ANYONE YET — a gap this bug uncovered rather than caused

`[MEASURED 2026-09-04]` **`build_queue` has ZERO rows for idea.uk**, the only site ever handed over.
It was our own rehearsal site, delivered to an address typed into the dispatch by hand, so it never
went through the order pipeline that writes `build_queue` — which 651's header (corrected
2026-08-31, `bugs_open/420`) names as the ONLY permitted recipient source, `sites.email` being
explicitly forbidden.

Found by running the pre_query with a **demand control** — the same query at `interval '0 days'`,
which *had* to return idea.uk and returned nothing. Without that control the zero reads as "nothing
due", which is what this failure would have looked like for ever.

**The obvious fallback fails too, and that was measured rather than assumed.** The delivery run does
record the address it used (`orchestration_states.collected_data->'input_data'->>'customer_email'` =
the address idea.uk went to) — but the **oldest row in that whole table is under 24 hours old**
(6,662 rows, oldest 2026-09-03 11:47Z; `stale-orchestration-reaper` runs every 180s). A follow-up
due in seven days would look for it six days after it was reaped.

> **THE REAL GAP: the estate has no durable record of who a delivered site was delivered to.**
> `build_queue` holds it only for order-originated sites; `orchestration_states` holds it for about a
> day. **The structural fix is to stamp the recipient onto the site row in the same statement that
> stamps `handed_over_at`** — a change to `platform/delivery/prepare.go`, which is the
> `site_delivery_and_editor` lane's surface, so it is routed to them and NOT made here.
> `775`'s verify block reports the gap on every apply (`1 of 1` today) so it cannot go quiet.

## 5. STILL OPEN — three, all needing the owner

1. **The interval.** "a week or so" is not a number. The action REFUSES to run without one rather
   than defaulting, so his answer cannot be quietly inherited from the placeholder in `775`.
2. **⚠ The first working run emails HIM.** idea.uk's delivery address is `aaa@designconsultancy.co.uk`.
   He must be told before, not after.
3. **The durable recipient record** (§4), which blocks the sender for every site delivered outside
   the order pipeline.

**Step C — restoring the stronger wording on the confirm page — is deliberately NOT done**, and the
tripwire test is what makes doing it a conscious act. It should happen when the sender is live, and
the same is true of the email's *"so we stop reminding you"*, which `776`'s header and rollback both
say should come back then.

## 6. COUNCIL: both rounds APPROVED (2026-09-04)

| round | correlation | verdict |
|---|---|---|
| step A — honest copy | `b9fc0004-74a4-4e17-8e5a-0e9c82d32052` | **APPROVED**, 1 advisory |
| step B — the sender | `3555a7a1-cf53-4b3b-91ba-4907a2e43ae4` | **APPROVED**, 4 advisories, none high-severity |

> ⚠ **Step A's commit `76ec663d3` carries a DEAD correlation** (`Council-Submitted: eee96972…`). That
> run was killed by the fleet's Anthropic credit exhaustion (11:21Z–~12:00Z) — `COMPLETED` at
> `complete_invalid`, no `council_report`, `__step_error` = *"no reviewer produced a readable opinion
> (6 abstained, 11 unreadable) — a council with no opinions cannot decide"*. Forward-only forbids an
> amend, so `098` will read that commit as unreviewed for ever. **`b9fc0004-…` is its live,
> approved replacement.** No `Council-Reviewed:` has been back-dated onto a later commit to tidy the
> report; the mapping is here instead.

Objections acted on: a cross-reference comment added **in `prepare.go`** naming the duplicate link
builder (three seats raised it); the post-claim errors now **name the site id** and the RUNBOOK gained
a "stamped but never sent" recovery section (the failure was already durably logged, but triage reads
the work item, not the log); and the placeholder guard is now explicitly documented as a **local** fix
after checking that no shared template-safety mechanism exists. Two objections were factually wrong,
both because my submission's `sketch` fields showed only a fragment of the migration — the seats read
the plan, not the tree. Full dispositions: this lane's `NOTES_delivery_followup.md`.

## 7. THE BLOCKER IN §4 IS CLOSED — and it was worse than §4 said (2026-09-04)

§4 established that the recipient is recorded nowhere durable. The `site_delivery_and_editor` lane
then measured the half I had missed, and it inverts the shape of the problem:

> `[MEASURED 2026-09-04, verified at the live row]` idea.uk carries
> `sites.email = 'idea.uk@contactforsales.com'` — a **site mailbox**. The delivery went to
> `aaa@designconsultancy.co.uk`.

**There is no NULL to warn anybody.** The column a reader reaches for is populated, well-formed,
plausible and wrong, so someone answering a support or refund question is confidently misled. An
absence sends you looking; a misleading value does not. That is the argument for a table rather than
a documented convention, and §4 had not made it.

**FIXED. Migration `778` (`site_deliveries`) is APPLIED, and the Go that writes it is committed
(`698b144fa`).** Owner ruling, relayed: a dedicated record rather than `sites.delivered_to`, at the
placement this lane proposed — `sites` is read by a great many things and `bugs_open/420` exists to
control which address lives where.

- **The recipient is written in the SAME STATEMENT that claims the handover** — a CTE on
  `StampHandover`'s claim, selecting from the rows the claim actually won. So the record cannot exist
  without the delivery, and no follow-on write can be lost. Doing it as a second call would have
  rebuilt, in the commit that fixes it, the identical shape as the follow-up's own "stamped but never
  sent".
- **`StampHandover` REFUSES an empty recipient**, before touching the database. Handing a site to a
  customer without knowing which customer is now unrepresentable rather than discouraged.
- **Proven against real Postgres, rolled back, with a demand control**: a won claim stamps and
  records; a LOST claim records nothing; **the same INSERT ungated DOES fire**, so that zero is the
  claim gate's doing and not a broken insert; a second claim neither stamps nor records.
- **⚠ The backfill expired the same day and was caught with hours to spare.** The only
  machine-readable copy of that address was the delivery run's `orchestration_states` row, and that
  table retained **1 day 02:11** when measured at 13:59Z. Applied 14:50Z, captured **1 of 1**.
  Tomorrow it would have been a human typing it out of a document.
- **⚠ ORDERING is the one way this breaks production**: without the table, every delivery fails at
  the claim. `778` was applied before the code could ship. Council round `62a99103` was submitted
  before applying; the seats therefore audit a migration that went live mid-round, which is stated
  rather than hidden — the alternative was a dirty tree holding a change that breaks deliveries the
  moment another session sweeps it, which is not hypothetical on this tree.

**And §5's first open question is answered.** The owner ruled the follow-up interval at **THREE DAYS**
(verbatim: *"I think the follow up should be 3 days"*), replacing `775`'s placeholder. The copy also
gained two things from his own hosting run: that setup takes about forty minutes and that slowness is
not a fault, and a paragraph telling the customer to **open their new address in a private window** —
because a Netlify Drop site is private by default and looks perfectly public to whoever uploaded it,
so a customer can press the confirm button in good faith with a site nobody can reach, and
`transfer_confirmed_at` would then suppress the one message that might have told them.

**Still open:** the schedule stays **disabled**, because the interval being settled is not the same as
somebody deciding to email a real person today — and the first real run emails the owner.
