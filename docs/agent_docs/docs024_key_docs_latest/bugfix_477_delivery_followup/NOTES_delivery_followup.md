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
the delivery actually used (`input_data.customer_email` = the address idea.uk went to). But
`SELECT min(created_at), count(*) FROM orchestration_states` = **2026-09-03 11:47:55Z, 6,662 rows** —
under 24 hours of history, with `stale-orchestration-reaper` running every 180s. A follow-up due in
seven days would look for that row six days after it was reaped. I nearly wrote a COALESCE onto it.

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
