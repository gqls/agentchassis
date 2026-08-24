# What needs your ruling — `bugs_open/326` and its neighbours

**Written 2026-08-24.** Five decisions. Only the first is hard; the rest mostly follow from it or
need a name attached. Each is written as: what the thing is → what the rule says → how this case
measures against it.

---

## 1. The one that matters: should the work-item door be allowed to DESTROY a request?

### What the thing is

When an agent finishes a stage it files a **work item** — a durable row saying "somebody do the
next thing". A brake sits in front of that filing to stop the same request being made over and
over: if an identical request finished within the last three hours, or finished twice in a week,
the brake acts.

**Today, when the brake acts, the request is destroyed.** Under three hours it writes nothing at
all and reports success. Past two attempts it writes the row into a status nothing ever picks
up. Either way the caller is told the same thing it is told when the work is genuinely already in
hand, so nothing downstream can tell "you're covered" from "I threw it away".

### What the rule says

Your ruling of 2026-08-02 §2: new authority on a shared mechanism ships as an **opt-in field with
the unsafe default OFF**, so the decision sits where a reviewer of the *caller* can see it — not
as a global switch and not as a comment.

### How this case measures against it

I proposed making both brake arms **delay** instead of destroying: the row is written, marked
"not before <time>", and picked up later. It reuses machinery that already exists and is already
honoured in three places, changes no status list and no index, and ships with an off-switch.

**The council's guardian vetoed it**, on two grounds, and both are right:

- it changes **default** behaviour for every caller at once (**36** non-test call sites as of
  2026-08-24) with only a global env switch to revert — which is the shape your §2 ruling exists
  to refuse;
- the customer bug **was already fixed without it**, by the config change. So the wider change
  was riding on someone else's urgency.

I have not contested it, and I am not asking you to overrule it. I am asking which of three
shapes you want, because the veto named the alternative but the alternative has a cost nobody has
weighed.

### The three options, costed

| | what it does | what it buys | what it costs |
|---|---|---|---|
| **A** | ship the deferral as proposed | every unclassified caller protected at once, including the 36 Go sites nobody will audit one by one | one day of fleet-wide behaviour change; ~**570** keys (2026-08-24) become live work that previously vanished; dashboards and "is this site clean" counters see rows they did not see before |
| **B** | opt-in per caller, as the guardian suggested | satisfies §2 exactly; blast radius is incremental | **I think this is close to a no-op.** It protects only callers someone has already thought about — which is precisely the set the existing `recurrence_expected` flag already covers. It adds a second lever doing the first lever's job, and the 14 undeclared steps stay unprotected |
| **C** | leave the brake alone; drive the census to zero and make the declaration *mandatory* | the strongest end state — the unclassified case stops being possible | slowest; protects nobody in the meantime; and it cannot reach the 36 Go call sites at all, because the census only sees config |

### The number that should probably decide it

The damage is **growing measurably**: rows born into the dead status went **635 → 661 in one
day** (2026-08-23 → 2026-08-24), i.e. roughly **26/day**. That is the cost of choosing C alone,
per day, until the census is driven to zero.

**My view, offered as a view:** A as an interim with C as the destination, keeping the off-switch
as the retreat. But the veto's whole point is that this should not be decided under bug-fix
urgency, so it is genuinely yours and I have not pre-committed anything to it.

**Where it lives:** `architecture_review/RFC_048_the_anti_churn_brake_may_delay_work_but_may_not_destroy_it.md`,
with the working patch beside it (`RFC_048_proposed_deferral.patch`, applies clean to HEAD, whole
package green, five mutations proven).

---

## 2. Migration 573 — and it becomes a hazard if you decline decision 1

### What the thing is

`573_domain_submitter_refuses_to_report_success_over_nothing_HOLD.sql` makes the front door
**fail loudly** when a submission genuinely queues nothing, instead of reporting success. It
depends on code that only exists in RFC_048's patch.

### What the rule says

A `_HOLD` migration is held back from the runner *for ordering* and applied by hand once its
condition is met. It is not a parking space.

### How this case measures against it

- **If you accept decision 1:** 573 applies by hand after the roll that carries the code. Normal.
- **If you decline decision 1:** its condition can never be met, and it becomes a permanently
  stale `_HOLD` file — **exactly the trap that was just cleaned up on migration 524 today**, where
  a stale held twin sat at HEAD telling readers to hold something that had been live for three
  days.

**So a decline is not "do nothing" — it should be "delete 573".** Say the word and I will, or
leave it to the fresh session with this doc.

---

## 3. Who tells the other 14 lanes?

**14 keyed steps** (2026-08-24) have never declared whether the item they file is a *request*
(repeat is normal) or a *detection* (repeat means the fix is not working). For any of them a
repeat request can still be silently destroyed.

`./scripts/audit-undeclared-recurrence.sh` names them. I deliberately did **not** classify them —
that means judgement calls inside lanes I do not own, and the draft that swept thirteen of them
found its own counter-example (`claims-auditor` genuinely needs the counter; setting it would have
broken it silently).

**The decision is only who does the telling:** one sweep by one lane with the owners consulted, or
each lane told to run the audit for itself. Nothing technical is blocked either way.

---

## 4. The 661 dead rows — already yours, still open

Rows already born into the undispatchable status: **661** as of 2026-08-24, largest populations
`page_rerender` and `improve_tool`.

I did **not** touch them, deliberately: re-promoting them would fire hundreds of renders at once,
and what to do with that landfill is already **your** open decision from RFC_010 / `bugs_open/033`
D2. Overruling it from inside a bug patch is what the failure-ladder header explicitly warns
against.

Flagging only that it is still open, and now growing at ~26/day.

---

## 5. The `bugs_open/345` ownership collision

Two sessions were instructed to take the same ownerless change — one told to *adopt* it, one
(me) told to *claim* it, about an hour apart. Given the shared account, parallel instruction is
the likely explanation.

The code is committed, verified green against HEAD, and inert. The other session remains the
council submitter and their correlation is cited on the commit.

**Nothing is broken and nothing is blocked.** The decision is only whether you want one of us
named as owner going forward, so the next instruction does not fork the same way. Their close
condition (345 stays open until the capability is *live*, not merely committed) is the right one
and I would not change it.
