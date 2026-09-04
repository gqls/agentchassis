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
