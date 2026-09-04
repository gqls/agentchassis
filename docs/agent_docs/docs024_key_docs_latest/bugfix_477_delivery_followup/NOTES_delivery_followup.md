# NOTES — bugs_open/477, the confirm button and the follow-up that does not exist

Append-only, newest at the bottom. Technical log: what was tried, what the system actually said, and
every misstep.

---

## 2026-09-04 — session opens

**Ownership first, because `who-owns.py` said OWNED and it was wrong.** `scripts/who-owns.py 477`
returned `VERDICT: OWNED or recently active`. Its own evidence block showed why that was a false
positive: the only commit it found was the FILING commit (`01f46c87a`, 2026-09-03 22:16), and
"likely OWNING workstream(s)" was `(none identified)`. The filing lane's live session
(`site_delivery_and_editor`) answered directly: *"477 is yours. Not claimed, nobody is building it."*

> **The check that resolves it is asking, and it took one message.** `who-owns.py` reads COMMITS, so
> filing a bug and never queueing it produces a permanent OWNED reading with nobody behind it. The
> filing lane has since added a §6b to their handoff recording that this session owns 477, and wrote
> the general lesson themselves: **filing a bug is not the same as queueing it.**

**Everything in the bug file re-measured before use** (`[MEASURED 2026-09-04 11:43Z]`), because a
count goes stale by addition and this one is a day old. All four claims held: 6 grep hits for
`transfer_confirmed_at`, all in `handover.go`; 1 mail-capable agent; 0 scheduled tasks targeting it;
60 sites / 1 handed over / 0 confirmed.

**Two things in the bug file's fix candidates turned out to be wrong**, both found by reading the
code the candidates rest on rather than the candidates themselves:

1. Candidate 1 says a follow-up sender is *"mostly seeding rather than Go"* because the mail action
   exists. `SendDeliveryEmailAction` calls `delivery.Claim` → `StampHandover`, which claims
   `WHERE handed_over_at IS NULL` and returns `ErrAlreadyDelivered` otherwise
   (`platform/delivery/prepare.go:261-267`). **The existing action refuses, by design, exactly the
   population a follow-up exists for** — every site it would target is already handed over.
2. Candidate 2 says the honest-copy fix is *"One migration, no new machinery"*. The confirm-page copy
   is **hardcoded Go** (`internal/core-manager/handlers/delivery.go:405-432`), not config. So it
   needs an image and a core-manager roll. The *delivery email* copy IS config
   (`send_delivery_email_action.go:8`, stated there as a deliberate property) — two surfaces, two
   mechanisms, and the bug file conflated them.

**Step A applied.** The sentence *"You will not get any more reminders about it."* deleted from both
`confirmOK` and `confirmButton`, replaced with nothing, with the measured reason and the restore
condition written into the comment at the site.

**The tripwire test was MUTATED before being believed** — both arms, separately:

```
### MUTATION 1: restore the promise on the BUTTON page only
--- FAIL: TestNoConfirmPagePromisesRemindersStopWhileNothingSendsThem (0.00s)
### MUTATION 2: restore the promise on the SUCCESS page only
--- FAIL: TestNoConfirmPagePromisesRemindersStopWhileNothingSendsThem (0.00s)
### RESTORED
ok  github.com/gqls/agentchassis/internal/core-manager/handlers  (uncached, -count=1)
```

A single mutation would have proven only that *one* arm is live. The first restored run reported
`(cached)`, which is not evidence of anything — re-run at `-count=1` before believing a green.

**Verified against committed HEAD, not just locally:**
`scripts/verify-head-builds.sh --with internal/core-manager/handlers/delivery.go --with …delivery_test.go --test ./internal/core-manager/...`
→ `OK — tests pass against HEAD 8ec048689`.

**Council gate.** `DRY_RUN=1` first, which cost nothing and caught a real error: my `operation`
values (`copy_change`, `add_test`) are not in the allowed set (`modify|add|remove|config_change`).
Submitted after fixing: **`SUBMISSION_CORR=eee96972-a562-45e1-b28e-80526213c82d`**.

> ⚠ **A missing council run today may be the FLEET'S CREDIT, not latency.** The `farmerinsurance_uk`
> lane recorded (commit `8fa6d4eb9`) that the fleet's Anthropic credit ran out at **11:21:12Z** —
> 32 successes then 18 straight failures across council-gate, diagnose-agent and landmine-verifier,
> with no second provider. My submission went out at ~11:50Z, i.e. **after** that. So the standing
> advice "a missing orchestration row is almost always latency, do not retry" has a third reading
> today: it may be neither latency nor a dropped dispatch. Check the run's actual status before
> concluding anything, and do not read a null result as a verdict.

## 2026-09-04 (afternoon) — step B, and the two things that nearly shipped wrong

**A THIRD SURFACE, found by reading the seed and then checking the live row.** 651's seed for
`delivery-email-sender` carries a `body_template` ending *"press the button here so we stop reminding
you"*. Verified at the LIVE `agent_definitions` row, not from the seed (the seed is history; the live
row is fact). So the false promise was in three places, not two — and the third is the only one a
human has read, because the idea.uk rehearsal email carried it.

It is also the only one that was fixable the same day: the email copy is CONFIG. I did not write SQL
against it, because it sits in the `site_delivery_and_editor` lane's surface and they had told me
they were writing a `body_template` migration for `bugs_open/475`. I messaged them instead. Their
reply corrected my premise — **that migration does not exist and is blocked** on the owner's Netlify
run — so folding it in would have parked the false clause behind work that may not land this week.
They shipped it standalone as migration **`776`** within the hour, and it is live.

> **The lesson is not "coordinate", it is what coordinating cost and bought.** One message cost about
> a minute and prevented two lanes writing migrations over one jsonb path. It also produced a better
> fix than mine: I proposed deleting the clause, as I had on the two pages. They pointed out that the
> email's sentence had **only** that clause as its reason to press, so deleting it leaves "press the
> button here:" with no answer to *why* — the pages each stated the reason separately, which is why
> deletion worked there and not here.

### The two near-misses, both caught by a control rather than by review

**1. `publish_target` is not a URL.** My first `pre_query` derived `live_site_url` from
`COALESCE(NULLIF(s.publish_target,''), 'https://' || s.domain)`. `publish_target` holds a **worker
name** — boxingonline.com's value is `b2worker` — and `publish_project` holds the host. Both are
EMPTY on idea.uk. So that COALESCE would have emailed a customer the word "b2worker" as their website
address. Caught by looking at the actual column values before trusting the name. The shipped version
is `'https://' || s.domain`, which is exactly what the delivery email itself used
(`RUNBOOK_site_delivery_and_editor.md:425`).

**2. THE DEMAND CONTROL IS THE ONLY REASON I KNOW THE SENDER SELECTS NOBODY.** I ran the pre_query
and got 0 rows, which is correct — idea.uk was handed over less than 7 days ago. Then I ran the
identical query at `interval '0 days'`, where it *had* to return idea.uk. It returned **nothing**.

`build_queue` has **ZERO** rows for idea.uk. It was our own rehearsal site, delivered to an address
typed into the dispatch by hand, so it never went through the order pipeline that writes
`build_queue` — and 651's header (corrected 2026-08-31, `bugs_open/420`) names that as the ONLY
permitted recipient source, `sites.email` being explicitly forbidden.

> Without the control, a sender that can never select anybody produces **the same 0** as a sender
> with nothing due. It would have sat there looking healthy for ever. This is the
> `a-post-fix-zero-needs-a-demand-control` lesson arriving in a new place: the zero was not a
> post-fix zero, it was a **pre-launch** zero, and it needed the control just as much.

**And the fallback fails too — measured, not assumed.** `orchestration_states` records the address
the delivery actually used (`input_data.customer_email` = the address idea.uk went to). But that table is a QUEUE, not a
history: `sql_for_agents/466` deletes terminal rows whose `updated_at` is over **24 hours** old, so a
follow-up due days later reads a reaped row. I nearly wrote a COALESCE onto it.
*(This paragraph originally sized that with `SELECT min(created_at)`. See the correction below — that
is the wrong column and it over-states the margin.)*

**So the real gap: the estate has no durable record of who a delivered site was delivered to.**
Routed to the `site_delivery_and_editor` lane, since the fix belongs in the statement that stamps
`handed_over_at`. `775`'s verify reports the gap on every apply (`1 of 1` today) so it cannot go
quiet.

### What was proven, and how

- **Real Postgres, rolled back**: five fixture sites (due / confirmed / already-sent / too-early /
  window-closed) each claim exactly what they should; the winning site re-claimed returns 0 (the
  at-most-once property itself); and a **negative control** — drop only `transfer_confirmed_at IS
  NULL` and the confirmed site DOES claim, so the refusals are that predicate's doing and not a
  fixture that could never have claimed anything. Rollback verified: 60 sites before and after.
- **Three mutations, each caught**: drop the suppression predicate (2 tests fail), drop the
  at-most-once predicate (3 fail), drop `ErrFollowupSuppressed` from the action's quiet-refusal
  switch (1 fails).
- **A migration dry run caught a real defect before it applied**: a PL/pgSQL variable named
  `is_nullable` is ambiguous against `information_schema.columns`' own column, and Postgres refuses
  the whole block. `774`'s verify would have failed on first apply. The variables are `v_`-prefixed
  now and the comment says why.
- **`verify-head-builds.sh`**: builds `./...` with all five files overlaid on HEAD, and
  `./platform/delivery/...` tests pass against HEAD.

> ⚠ **HEAD IS ALREADY RED IN TWO PACKAGES, NEITHER OF THEM MINE**, and I checked by running them
> against clean HEAD with no overlay rather than assuming: `platform/orchestration/actions` fails
> `render_seam_one_spelling_test.go:246` (UNDECLARED template executor `renderFailWorkItemMessage`,
> introduced by `83407cd37`, the `bugs_open/440` lane) and `cmd/config-key-audit` fails
> `findingcodes_test.go:329`. My own packages' tests pass. **I am not claiming a green package** —
> any session running `verify-head-builds.sh --test` on those two today will see red that is not
> theirs either.

Council: step B submitted, `SUBMISSION_CORR=3555a7a1-cf53-4b3b-91ba-4907a2e43ae4`.

## 2026-09-04 (mid-afternoon) — both rounds APPROVED, and what the advisories were owed

**Step A: `b9fc0004-74a4-4e17-8e5a-0e9c82d32052` — APPROVED**, 1 advisory objection.
**Step B: `3555a7a1-cf53-4b3b-91ba-4907a2e43ae4` — APPROVED**, 4 advisory objections, none
high-severity. Both `complete_approved`.

> ⚠ **STEP A'S FIRST CORRELATION (`eee96972-…`) IS DEAD AND ITS TRAILER CAN NEVER RESOLVE.** That
> run was killed by the credit outage; the commit `76ec663d3` carries
> `Council-Submitted: eee96972…`, forward-only forbids an amend, and `098` will therefore read that
> commit as unreviewed for ever. **The live replacement is `b9fc0004-…`, and it is APPROVED.** I have
> NOT back-dated a `Council-Reviewed:` onto a later commit touching the same files to tidy the
> report — that is precisely the dishonesty surface the trailer exists to protect. The mapping lives
> here and in `bugs_open/477`.

### Disposition of every step-B objection

| seat | objection | disposition |
|---|---|---|
| editquality | *775 shows only a `scheduled_tasks` INSERT; the matching `agent_definitions` row is not visible* | **Objection is mistaken, and the sketch is why.** `775` does insert the agent row (`INSERT INTO agent_definitions … 'delivery-followup-sender' … WHERE NOT EXISTS`); my `sketch` field showed only the schedule half. No code change. **The lesson is mine, not the seat's:** a sketch is what the reviewer reads, so omitting half the migration from it manufactures an objection. |
| editquality | *`followup_after_days` has no visible landing place once the owner rules* | **Mistaken for the same reason.** It is in `775`'s agent config (`'followup_after_days', 7`) as a placeholder. Re-read the file: the landing place exists and the RUNBOOK names it. No change. |
| editquality · tooling_provenance · guardian (**three seats**) | *`ConfirmTokenURL` duplicates `prepare.go`'s `tokenURL`; a comment in one file protects nobody reading the other* | **ACTED ON.** Added a cross-reference comment **in `prepare.go`** naming the duplicate, why they are apart, and the instruction to collapse them. `prepare.go` was clean in the tree at the time (checked before editing) and the edit is comment-only, 9 lines added, 0 deleted. |
| bug_historian | *the placeholder guard may reinvent a shared template-safety mechanism (016b §9 case 7)* | **CHECKED, then stated.** `grep -rn "missingkey" platform/ internal/` finds **only comments describing the hazard** — no shared guard exists. Also: this seam is not `text/template` at all, it is a `strings.Replacer` over a closed vocabulary, so a template guard would not cover it. Written into the action as a comment rather than left implied. |
| bug_historian | *a post-claim send failure leaves `followup_sent_at` set with no email and no way for an operator to find it* | **HALF RIGHT, and the half that was right is fixed.** The failure IS durably logged — `routeToErrorStep` writes `__step_error` and `agent_error_log` together (`coordinator.go:3997-4002`). But `v3_site_actions.go:6230` records that **triage reads the work item, not the log**, so the row was findable only by someone who already suspected. The three post-claim errors now **name the site id** and point at the RUNBOOK, and `RUNBOOK_delivery_followup.md` has a "stamped but never sent" section with the join and the deliberate-recovery warning. |
| tooling_provenance | *no durable cross-lane record; coordination was by message* | **ACTED ON.** `CONTRIB_2026-09-04_from_the_477_lane_no_durable_record_of_who_a_site_was_delivered_to.md` written into the `site_delivery_and_editor` lane directory, plus the register entry `EMAIL-003` and the `bugs_open/477` §4 contribution. The message was how it was raised in time; the file is how it survives the session. |

**Two of six objections were factually wrong, and both for the same reason — my `sketch` fields
showed a fragment.** Worth carrying: the seats read the plan, not the tree, so an abbreviated sketch
is not a neutral summary, it is the evidence. That is a cheaper lesson than it sounds — a REVISE on
either would have cost a round.

## 2026-09-04 (late afternoon) — the blocker is closed, and the record was captured with hours to spare

**The gap turned out to be worse than "no record", and the peer lane found the half I had missed.**
I measured that the recipient is stored nowhere durable. They measured what a reader *gets*:
`[MEASURED 2026-09-04, verified by me at the live row]` idea.uk carries
`sites.email = 'idea.uk@contactforsales.com'` — a **site mailbox** — while the delivery went to
`aaa@designconsultancy.co.uk`.

> **An absence makes you go and look. A misleading value does not.** There is no NULL to warn anyone:
> somebody answering a support or refund question finds a well-formed, plausible address and is
> confidently wrong. That difference is the whole argument for a table over a documented convention
> about which column to read, and I had not made it — I had only established the absence.

**Owner ruling (relayed by the `site_delivery_and_editor` lane, and recorded as second-hand rather
than as my own judgement):** a dedicated record, not `sites.delivered_to`, written in the same
statement as `handed_over_at`. The reason is the one that made me route the question instead of
answering it — `sites` is read by a great many things and `bugs_open/420` exists to control which
address lives where. **The placement he chose is the one this lane proposed**, which is the argument
for routing a question you are not entitled to answer rather than guessing well.

### The two hazards this created, and how each was handled

**1. The ordering hazard is real and it is the one that breaks production.** The new
`StampHandover` writes `site_deliveries` inside the claim, so without the table **every delivery
fails at the claim**. Migrations are live-on-apply, Go is inert until a roll, so the safe order is
table-then-code — but on this tree an uncommitted Go change is not safe either: it can be swept into
another lane's commit and shipped by their build. That is not hypothetical; my `LANDMINES` and
`WRONG_CALLS` entries were swept into `d6077796a` earlier today.

> So: **778 applied first, commit immediately after.** The cost is that the council seats are
> auditing a migration that went live mid-round, which I stated in the commit rather than hid. The
> alternative was a dirty tree holding a change that breaks deliveries if anyone commits it, and that
> is worse. Submitted before applying, which is the part of the rule that protects the review.

**2. The backfill had 5h31m left — NOT the ~7 hours I first read.** Applied 14:50Z with 4h40m to
spare; captured **1 of 1 recoverable**. Tomorrow it would have been a human typing an address out of
a document and hoping the document was right.

> **CORRECTED 2026-09-04 — I measured the wrong thing, and a peer lane re-measuring is what exposed
> it.** I sized the margin as `now() - min(created_at)` over the whole table (1d02:11 → "about seven
> hours"). **That does not measure retention.** It reports the oldest SURVIVOR's birthday, while the
> policy keys on a different column entirely: `sql_for_agents/466` deletes `WHERE status IN
> ('COMPLETED','FAILED') AND updated_at < now() - INTERVAL '24 hours'`. A recently-touched old row
> keeps `min(created_at)` pinned — `[MEASURED 14:29Z]` the oldest survivor was 1d02:41 by birthday
> and had been **updated 22:24 ago**, comfortably inside the window.
> **The true deadline was that row's `updated_at` (19:30:40Z) + 24h = 2026-09-04 19:30:40Z.** Both
> lanes had been quoting ~20:56Z, which is 1h26m too generous.
> **The right question is never table-wide** — ask the row:
> `SELECT updated_at + interval '24 hours' - now() AS time_left FROM orchestration_states WHERE …`.
> The peer read the drift between two of these readings as *"retention is elastic"*. It is not
> elastic; it is a flat 24 hours from a column neither of us had read. **Both of us mis-measured the
> same way and neither reading would ever have come out right**, which is what makes it an instrument
> error rather than a stale figure. Corrected in both landmine entries, the `778` header, the bug
> file, the CONTRIB and the RUNBOOK.

### What was proven, and the control that makes it mean something

Real Postgres, rolled back, four cases:

```
PASS: a won claim stamps AND records, in one statement
PASS: a lost claim records nothing
CONTROL PASS: the ungated insert DOES fire, so the zero in test 2 is the claim gate doing it
PASS: a second claim neither stamps nor records — the recipient of record is the first one
```

The control is the load-bearing one. "A lost claim records nothing" is a zero, and a broken INSERT
gives the same zero — so the same statement without the claim gate is run and must fire. It did.

**Why one statement rather than an INSERT after the claim:** a claim that succeeds and a follow-on
write that is lost leaves a delivered site whose recipient nobody knows — the identical shape to the
follow-up's own "stamped but never sent". Building a second instance of that defect inside the commit
that fixes the first would have been a poor day's work.

### The pre-commit architecture signal fired, and I think it is right to have fired

`.githooks/pre-commit` flagged **"exported symbol removed/changed"** (`StampHandover`'s signature)
and **"migration + platform code in one commit — needs a staged rollout order"**, and said it meets
the RFC trigger test.

My reading, recorded so a later reader can disagree with the reasoning rather than guess at it: the
staged-rollout half is **satisfied and stated** (778 applied before the code could ship, named in the
commit and in both files). The exported-symbol half is a **genuine guarantee change** — a handover
now *cannot* happen without a recorded recipient — but the consumer set is enumerated and is **one**
(`Claim`, whose own consumer set is one: `send_delivery_email_action.go`). Under the owner ruling of
2026-07-29 §1, a change needs an RFC when it changes what a shared mechanism GUARANTEES to its
consumers; here the guarantee is strengthened, the single consumer is updated in the same commit, and
nothing else can observe it.

**I am not treating that as settled by my own say-so.** The round in flight
(`62a99103-3097-4aeb-9aeb-be0d190c534e`) has the architecture seat in its footprint, and if it rules
otherwise the RFC gets written. That is the mechanism the estate has for exactly this question, and
using it is cheaper than being confident.

### Round 3 verdict: APPROVED (`62a99103-3097-4aeb-9aeb-be0d190c534e`), and the architecture seat ruled the scope question

**`ARCHITECTURE_SIGNAL: point_fix | DEFLECTIONS: 0`** — so the pre-commit hook's RFC trigger is
answered by the seat rather than by my own reading, which is what I said I would do rather than
settle it myself. The guardian seat's note is worth keeping too: *"Contained: new table, one narrow
package, one production caller, no changes to orchestrator/messaging/dispatch core or shared wire
formats. The migration-before-roll ordering hazard is stated plainly in the plan itself rather than
glossed over, which is the right way to carry a risk."*

Approved with 3 advisory objections (8 raised across seats, none high-severity). Dispositions:

| seat | objection | disposition |
|---|---|---|
| editquality | *editing `775` in place changes the FILE, not the LIVE ROW, if it was already applied* | **The RULE is right; the PREMISE is false, and I checked rather than argued.** `775` has never been applied — `SELECT count(*) FROM scheduled_tasks WHERE name='delivery-followup-send'` → **0**, and `sites.followup_sent_at` does not exist either. So the in-place edit genuinely changes what will be applied. **Added to the RUNBOOK as a state table plus the query**, with the explicit note that this stops being true the moment `775` is applied, after which the interval or the letter needs a NEW numbered migration. A good objection that would have been right on a different day. |
| prior_art_librarian | *`customer_access_tokens` may already carry a recipient — the DORMANT-MACHINERY shape* | **Answered by looking.** Thirteen columns, none of them a recipient: `id, site_id, purpose, token_hash, issued_at, expires_at, single_use, used_at, use_count, revoked_at, created_by, stored_url, stored_url_expires_at`. `created_by` is the ISSUER (`'delivery-email'`), not the customer — it tracks a token's lifecycle, not an identity. **The objection was fair even though the answer is no**: I asserted a new table without ruling out the existing one, and a peer lane telling me it had no recipient column is not the same as my having looked. Recorded in the RUNBOOK with the one-line check. |
| debug_historian | *the ordering hazard is named but no POD-LEVEL verification is described; "migrations are live, Go is inert" is correct reasoning and has still caused real incidents* | **ACTED ON, and it is the best objection of the round.** The RUNBOOK gains an ordering section that checks the running binary by ANCESTRY (`service_binary_capabilities` → `git merge-base --is-ancestor 698b144fa`) rather than by argument, states the dangerous combination explicitly (a descendant pod with no table), and notes that a `778` rollback is the only thing that could recreate it — which is why that rollback refuses while the table holds rows. |
| guardian | *confirm no other config carries the `followup_after_days` placeholder* | **Checked: 0 and 0** across `agent_definitions` and `scheduled_tasks`. No second copy to diverge from the ruling. |
| debug_historian | *no rollback artifact named for `778`* | **Mistaken, and it is my sketch's fault again** — `778_..._ROLLBACK.sql` exists, refuses while the table holds rows, and refuses while the schedule is enabled. It was not in the edit list. **Third round running that an elided sketch manufactured an objection**; I keep paying the same small tax. |
| guardian · prior_art_librarian | *the "one production caller" claim needs an attached symbol search, not an assertion* | **Fair on process even though the count is right.** `grep -rn 'StampHandover('` → one caller; `delivery.Claim(` → one production caller plus six test sites. Attaching the output, not the number, next time. |
| debug_historian | *`DISTINCT ON` needs a deterministic ORDER BY* | Already deterministic in the file — `ORDER BY 1, created_at DESC` (newest run per site wins). Elided by the sketch, same tax. |

> **THE PATTERN ACROSS ALL THREE ROUNDS, and it is mine to fix: four of the objections I have
> received were manufactured by an ABBREVIATED `sketch` field.** The seats read the plan, not the
> tree. An omitted half does not read as omitted, it reads as absent — and answering that costs a
> round when it lands as a REVISE. Cheaper: paste the real hunk.

## 2026-09-04 (late) — two commitments this lane now owes another lane

Recorded because a cold-start reader of this lane would not otherwise know either, and both are
things that look like tidy-ups a future session might "helpfully" do or skip.

> **⚠ SUPERSEDED WITHIN THE HOUR — read the correction at the foot of this file before acting on
> commitment 1. The dependency it describes is gone: converting this caller is now THIS lane's work,
> on its own schedule, and the pre-substitution can stay indefinitely.**

**1. DO NOT delete the `{{instructions_link}}` pre-substitution in `send_followup_email_action.go`
until the `bugs_open/475` lane says their shared vocabulary landed — or that it did not.** It looks
like a redundant hand-patch around a shared filler. It is the only thing filling that token today,
and removing it early leaves a gap where neither lane owns the fill. Their round:
`c8ed56d2-74ea-4bcc-a0a4-73050c436693` (submitted before writing code, six edits; this lane's action
is edit 3 and its `{{zip_link}}` never-reason is carried explicitly).

**2. This lane owes the `tokenURL` / `ConfirmTokenURL` collapse, once `778` has settled.** Three
council seats flagged the duplicate on 2026-09-04. The 475 lane scoped it OUT of their round and
referred it here, with a better statement of the reason than the one in my own file's comment:
*"bundling it would have demonstrated the failure it fixes"* — a same-file edit from two lanes is how
one lane's work gets lost, which is precisely what the duplicate exists to avoid.

**What I gave their round, as a reviewer, and why it was worth reading their plan against my code
rather than against itself:** two requirements their design did not state, both invisible from their
side.

- **The same token has different PROVENANCE per caller.** Their edit 5 resolves the instructions URL
  from `Links.Instructions`, populated inside `Claim` — and my caller never calls `Claim`. It calls
  `ClaimFollowup` and hand-builds its own `Prepared`. Three tokens already differ this way
  (`{{live_site}}` input vs cfg, `{{confirm_link}}` minted by me post-claim vs inside `Claim`, and now
  the instructions link). A `Vocabulary` modelling provenance per TOKEN cannot be satisfied by my
  caller at all.
- **"Pre-claim" names a different statement in each caller** — `Claim` at `send_delivery_email_action.go:154`,
  `ClaimFollowup` at `send_followup_email_action.go:188`. A shared helper documented as "call Check
  pre-claim" is ambiguous across two callers, and the failure is silent: a Check that runs after MY
  claim still passes its own assertions while having stamped `followup_sent_at`.

> **The general form, for this lane's own future reviews: read the other lane's plan against YOUR
> code, not against their plan.** Both findings were invisible from inside their design and obvious
> from inside mine, and neither would have surfaced by reading the submission carefully.

**And their point 3 found a defect in MY docs** — the enable-check named the commit that added the
action rather than the one that renamed the placeholder. Fixed in `b92beae38`; the reasoning is in
the RUNBOOK and in `775`'s own header, because the person applying it reads the migration.


## 2026-09-04 (evening) — CORRECTION: the two commitments above have changed, and one has inverted

The `bugs_open/475` round came back **REVISE, gated HIGH by the guardian seat**, and the objection
was that their submission edited **this lane's file**. Their edit 3 converted
`send_followup_email_action.go` to the shared `Fill`. The seat's point, near enough: it touches a file
explicitly said to belong to another lane, mid-flight on its own migration, in the same submission
that changes the shared vocabulary that file depends on — *"exactly the concurrent-edit collision
pattern the plan itself warns about elsewhere"*, since they had scoped `handover.go` OUT for that very
reason one edit earlier.

**Verified rather than taken on trust:** their resubmission (`9af520228`,
`COUNCIL_SUBMISSION_2026-09-04b…r2.json`) has **four edits and none of them touches this lane's
file** — `vocabulary.go`, `send_delivery_email_action.go`, `vocabulary_test.go`, `prepare.go`.

### What this lane now owes, replacing the list above

- **Commitment 1 is RETIRED, not merely relaxed.** There is no window where neither lane owns the
  fill, because their change binds a caller only *when that caller adopts `Fill`*. The hand-kept
  guard slice and the `{{instructions_link}}` pre-substitution keep working untouched. **Keep them for
  as long as suits this lane.** Their adoption model is copied deliberately from `ActionInputSpec`,
  whose own comment records that a fleet-wide version was rejected because *"an over-strict validator
  is a considerably worse bug than the inert key it is chasing"* and that adoption is *"driven by the
  coverage report, not by a flag day"*.
- **NEW: converting this caller to `delivery.Fill` is THIS LANE'S WORK.** Not theirs, not optional
  forever, and not urgent. When it happens, use `AssertNeverProduces(t, f, "{{zip_link}}")` from their
  test helper — it exists because the architecture seat objected that `NeverReason` otherwise depends
  on a human reading a string correctly for ever, which is that round's own thesis turned back on its
  author. That assertion is what makes "a scheduled follow-up can never mint a presign" fail a test
  rather than be read past.
- **The coupling landmine (`d67f08ff4`) STAYS, and deleting it is THIS LANE'S COMMIT.** They removed
  their deletion edit. With only one caller converted the trap is still live in this file, and
  **removing a warning from a real trap is worse than leaving a stale one.** It goes when the second
  caller adopts `Fill` — i.e. in the same commit as the conversion above.

> **The lesson is one this lane gave them and then received back.** I told them a comment in one file
> protects nobody reading the other; the seat told them that editing another lane's file inside a
> round that changes what that file depends on is the collision their own plan warned about. Both are
> the same rule — **the person who bears a risk should be the one who takes the change** — and the
> useful part is that neither of us spotted it in our own work, only in each other's.

## 2026-09-04 (evening) — round 4 APPROVED, and a defect caught in an unapplied file

**`ac9eb6b4-ae6a-4486-96ef-6f07dcf7b09c` — APPROVED, "all reviewers approve", no objections.** The
vocabulary conversion stands: both letter senders derive their guards from `delivery.Vocabulary`,
`fillTemplate` is gone, and the coupling landmine is retired in place.

**Four council rounds this lane, all approved:** `b9fc0004` (honest copy), `3555a7a1` (the sender),
`62a99103` (the recipient record, architecture seat `point_fix`), `ac9eb6b4` (the conversion).

### The last defect of the day, and it was mine, in a file that had never been applied

The `site_delivery_and_editor` lane found that `775`'s `pre_query` composed
`live_site_url` as `'https://' || s.domain`. **That is right only for a domain already pointed at us.**
For a customer who has bought a build and not yet transferred theirs — the normal case — it composes
a URL that does not resolve. Owner ruling the same afternoon: *"the delivery email should say it is on
the ugg2 subdomain not paper-cups.com."*

`[MEASURED 2026-09-04, by me, not taken from the report]` `publish_project` set on **2 of 60**;
idea.uk NULL; **no code writes it** — the only non-test reference is `b2worker.go:63`'s refusal, which
also refuses a `publish_project` equal to the site domain, i.e. the code's own statement that the two
are different things.

> **THE SHARPER HALF, and it is about how I check rather than what I found.** My own header, **two
> lines above the defect**, warns that a `COALESCE` over `publish_target` would have mailed a customer
> the word `b2worker`. **The fix I chose to avoid that trap landed on the other wrong answer.** I
> checked the column I was about to use and never checked the one I fell back to. The trap was one
> step along from where I was looking — which is the general form of every column-name failure this
> lane logged today, arriving in my own file after I had spent the afternoon telling three lanes to
> check the values behind a name.

**No `COALESCE` fallback to the domain**, and I nearly wrote one: it is correct for the two pointed
sites and quietly wrong for every customer who has not transferred — the same defect wearing a guard.
Sites with no serving host are excluded and **counted**, so `775` now prints `GAP 2` on every apply
alongside the recipient gap. Both read `1 of 1` today, and the sender consequently selects nobody.
**That is correct behaviour, loudly.**

**It was caught because the file had never been applied.** Every other instance of this class this
week reached a customer or a live row first.
