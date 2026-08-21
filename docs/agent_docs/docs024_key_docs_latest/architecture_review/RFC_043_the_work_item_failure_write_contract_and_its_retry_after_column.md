# RFC_043 — the work-item failure-write contract, and the `retry_after` column four read sites must honour

**Raised 2026-08-20** by the `bugfix_307_terminal_write_contract` lane, **at the architecture
seat's own direction and not as a gate**. Council corr `4cdec68b-fa17-436d-8e25-8c422ee6c8c5`,
round 1, verdict REVISE (gated on an unrelated code defect, since fixed):

> `architecture` (medium, advisory): *"New shared retry state machine + schema-column contract
> (`retry_after`) consumed across 4 read sites and 2 media (Go/SQL), armed by default for 9 live
> consumers — meets the architecture trigger test and is disqualified from the RFC_022 opt-in
> exception. Should get a recorded architecture_review entry (blast radius/rollback) even though
> not blocking."*
>
> *"The 08-18 owner ruling authorises the BEHAVIOUR; it doesn't exempt the MECHANISM implementing
> it from architectural review — those are separate questions. The plan's own diligence
> substantially satisfies what an RFC would ask for, so this is advisory, not a stop-ship: record
> it via architecture_review, don't gate the round on it."*

**The seat is right on both halves, and the second half is the interesting one.** I had treated
the owner's ruling of 2026-08-18 ("a transient blip should return the item to queued") as
settling the question. It settles *what should happen to the item*. It says nothing about
whether a new shared state machine plus a schema column read from two media is the right shape
for making it happen — and that is a different question, decided by different people, on
different evidence.

**This RFC asks for no decision it can block on.** The mechanism is built, committed and half
applied; review here is after the fact by design (owner ruling 2026-07-29 §2 — no thread on this
tree can hold a seam back, so honesty about that is worth more than a pretence of gating). What
it asks is that the shape be on the record where the *next* person to touch this seam will find
it, with its blast radius and its rollback stated rather than reconstructed.

## 1. Why this is architecture-scope by the estate's own test

Against the 2026-07-29 §1 trigger — *"an addition to a shared vocabulary needs an RFC only when it
changes what the shared mechanism GUARANTEES"* — this qualifies on three counts, and I did not
think so when I built it:

1. **It changes a guarantee, not just a capability.** Before: a work item that fails is written
   terminal by whichever writer got there, immediately, and any status it held is replaced. After:
   an item's failure is counted, delayed, and refused if the row records a deliberate decision.
   Every existing caller's behaviour changes without any caller opting in.
2. **A schema column is a contract with future readers.** `retry_after` is honoured today by four
   read sites in two media (`ClaimWorkItemAction`, `LoadWorkItemsAction`, `build-pipeline-trigger`'s
   `pre_query`, and its `find_dispatchable_site` selector inside `agent_definitions`). Any fifth
   reader of "is this item dispatchable" that does not know about the column is wrong by omission
   — the exact drift migrations 213 and 285 already exist to repair on this very query.
3. **It is armed for 9 live consumers.** Five agents carrying `update_work_item_status` with
   `status:'failed'`, four dispatch loops carrying `fail_work_item`'s ladder branch.

**And it is explicitly disqualified from RFC_022's narrow exception**, which requires all three of:
opt-in, unsafe-default-OFF, and zero live consumers naming it. This is armed by default and every
existing consumer's behaviour changes. The seat applied that test correctly and I should have
applied it myself.

## 2. The shape, stated plainly

One Go helper, `applyWorkItemFailureLadder`, is the single failure-path write. It does four things
that were previously done four different ways or not at all: counts the attempt against
`max_attempts`; stamps a not-claimable-before time whose minutes come from `reaper_policies`;
refuses to overwrite a status recording a deliberate decision; and, when the fleet is visibly
failing the same way at once, returns the item to the queue without consuming an attempt — but
never without a cooldown.

Two vocabularies sit adjacent in that file **because they must differ**: the failure-path guard
omits `failed`/`unresolved` (moving through them *is* the ladder), the completion-path guard
includes them (stamping `complete` over a failed row is the silent overwrite WII-003 closed). The
first version of the change used one list for both; that is what the council gated on.

## 3. Blast radius, measured

- **Consumers:** 9 live (enumerated in WII-024 and in the submission, via the nested
  `jsonb_path_query(default_config,'$.**.steps')` walk — a top-level `jsonb_each` returns **zero**
  rows for these actions).
- **Population the behaviour change reaches** [MEASURED 2026-08-20, archive-inclusive]: **401 of
  558** failed items in 14 days died before exhausting their budget (**72%**), 398 with
  `handled_by` NULL. These are the items that stop dying early. (The first figure I published,
  141 of 270, was read from the live table alone and understated it — the `work-item-archiver`
  drains terminal rows after ~7 days. The guardian seat caught it.)
- **Dispatch impact at apply time:** zero. `retry_after` is NULL on every pre-existing row, so the
  four predicates are tautologies until the chassis rolls. Verified rather than argued:
  `held_by_new_predicate = 0` immediately after migration 506, with the trigger still firing on
  its 60 s tick.
- **What it does NOT reach:** `fail_work_item`'s `status_override` branch (4 agents —
  `bugs_open/033` D2's territory), the six `needs_human_review` and six `complete` steps on
  `update_work_item_status`, and WII-019's ownership-refusal substitution.

## 4. Rollback

Three independent levers, deliberately not one:

1. **Env disarms, no redeploy** — `DISABLE_WORK_ITEM_DECISION_GUARD`,
   `DISABLE_WORK_ITEM_RETRY_BACKOFF`, `DISABLE_WORK_ITEM_TRANSIENT_RELEASE`,
   `DISABLE_WORK_ITEM_BURST_RELEASE`. Separate because the behaviours fail differently: a
   misbehaving burst detector must not force an operator to give up the guard as well.
2. **Config rollback** — `506_..._ROLLBACK.sql`, which since the `debug_historian` objection
   carries a **pre-state md5 gate** on both read sites and aborts rather than clobbering a later
   lane's edit (mutation-tested: it fires with expected/found md5s).
3. **Schema rollback** — `505_..._ROLLBACK.sql`, which refuses to drop the column while any live
   `pre_query` or agent config still names it. ⚠ Order matters: 506's rollback first.

**The residual risk in rollback is silence, and it is worth naming here.** The binary tolerates the
column being absent (latches the pre-migration statement shape on SQLSTATE 42703 and logs once), so
a rolled-back 505 leaves the backoff **inert while everything still reports success** — the
pre-307 behaviour, restored invisibly. That is in LANDMINES with its check.

## 5. The open questions this RFC actually wants answered

1. **Is a `reaper_policies` row the right home for retry numbers on this queue?** RFC_018 invited
   exactly this ("adopt reaper_policies for its numbers first, executor second") and this is its
   second consumer. But its first consumer *also* uses its executor function; this one takes only
   the numbers. Two consumers with different depths of adoption is how a shared mechanism becomes
   two mechanisms. **RFC_018's own stopping point said to decide the generalisation when a second
   queue arrived. It has arrived.**
2. **Should the completion-path guard converge?** Three writers now guard `complete`, two with an
   inlined seven-status list and one (this change) with a named eight-status constant that adds
   `cancelled`. I deliberately did not edit the siblings — that changes their behaviour — but
   three near-identical lists is the drift class this estate keeps filing bugs about.
3. **Where does a SQL-resident sweep fit a Go-resident contract?** `claimed-item-timeout` runs a
   fifth copy of the ladder in a `pre_query` and cannot call Go. Filed as `bugs_open/341` at three
   seats' insistence, with candidates ranked — but the general question (how does this estate stop
   a `scheduled_tasks` sweep and a Go action diverging?) is bigger than that bug.
4. **Is "armed with env disarms" the right default for an owner-mandated behaviour change?** Two
   live precedents point opposite ways — WII-019 (opt-in, unsafe default OFF) and WDS-018 (armed,
   env disarm, *"the owner has ruled against default-OFF switches that rot unexercised"*). I chose
   WDS-018 because this implements a ruling rather than granting caller-licensed authority. The
   `guidelines` seat agreed on record. It is still a judgement worth ratifying or overturning
   explicitly, because the next author will copy whichever answer stands.

## 6. Relations

Register **WII-024** (the contract) · `bugs_open/307` §8/§8b · `bugs_open/341` (the fifth writer) ·
**RFC_018 / SCH-024** (`reaper_policies`, whose invitation this accepts) · **RFC_022** as narrowed
2026-08-11 (the opt-in exception this is disqualified from) · **WII-003** (the guard pattern) ·
**WDS-018** / `bugs_closed/291` (the armed-with-kill-switch precedent) · **RSH-006/007** (the
classifiers layered here) · council corr `4cdec68b-fa17-436d-8e25-8c422ee6c8c5`.

---

# OWNER RULINGS, 2026-08-21 — all four questions answered

Given in response to §5, in the owner's own order.

## 1. `reaper_policies`: adopt the numbers only — LEAVE the executor until a THIRD consumer appears

> *"leave it until a 3rd consumer appears."*

So `site_work_items` stays a numbers-only consumer and `RFC_018`'s "generalise when a second queue
arrives" is **superseded on this point**: two consumers is not enough evidence to design a shared
executor, and an abstraction fitted to two is usually fitted to neither. **RFC_018's stopping point
should be read as "consider at two, decide at three".**

What this obliges: nothing now, and one thing later — whoever brings the third consumer inherits
the generalisation, and should find this ruling rather than re-deriving it. Noted in `SCH-024`.

## 2. The three completion-guard lists SHOULD converge — as their own change, with their own review

> *"yes they should become one with it's own change and review."*

Confirms the instinct that kept them apart here. The two older lists are hand-typed literals inside
`CompleteWorkItemAction` and `failUnverifiedCompletion`; the third is `workItemCompletionGuardStatuses`,
which adds `cancelled` and `deferred`. **Converging them is a behaviour change to two working code
paths dressed as tidying** — precisely the shape that goes wrong when it rides inside another bug's
patch, and precisely why it gets its own round.

Unowned and unscheduled. Whoever takes it: the blast radius is every caller of both functions, the
delta is `cancelled` + `deferred`, and the disconfirming question is whether any live flow
legitimately completes a `cancelled` row.

## 3. Shared contracts live in GO; a SQL sweep that needs one gets MOVED, not mirrored

> *"I agree to your rule that shared contracts live in Go."*

**The general rule this estate has been paying for in instalments.** `341` and `344` were both
instances within two days: each time the question was "duplicate the rule in SQL, or move the
sweep", and each time duplication was the cheaper-looking answer. It is now ruled out.

Note what the rule does NOT say: a `pre_query` may still hold logic that is *only* its own — the
`claimed-item-timeout` exclusion list, its staleness thresholds. The rule bites when a rule is
**shared** with Go, because that is when two spellings can disagree. And it has already saved work
once: `341` §5c stopped a mirror being written into a sweep that turned out not to need it.

## 4. Armed-by-default STANDS — with a condition, since the owner deferred this one to me

> *"I defer to your best judgement on this."*

**Ruling: keep armed-by-default, and require every kill switch to be EXERCISED BY A TEST before it
counts as justification for shipping armed.**

The reasoning, and it turns on evidence from this lane rather than preference. I shipped armed on
the WDS-018 precedent — *"the owner has ruled against default-OFF switches that rot unexercised"* —
and offered the env disarms as the safety that made it acceptable. Then the acceptance found that
**`DISABLE_WORK_ITEM_RETRY_BACKOFF` was itself broken**: setting it would have dropped `$4` from
every failure write (SQLSTATE 42P18) and taken down the whole failure path. **The safety valve I
justified the posture with did not work, and nothing would have told an operator until they reached
for it mid-incident.**

That is not an argument against arming. It is the *same* argument the owner used against default-OFF
switches, one level down: **an untested disarm rots exactly as an unexercised feature does**, and it
rots in the worst possible place — the thing you reach for when something is already wrong. So the
posture survives and acquires the obligation that makes it honest.

Concretely, for anything shipping armed with a disarm:
- each switch is driven by a test that asserts the disarmed behaviour, not merely that the flag is read;
- the test asserts the *statement or effect*, not a mock's bookkeeping — this lane's 42P18 defect was
  invisible to fifteen sqlmock tests because a mock never PREPAREs;
- shipping armed is justified by the disarm's test, not by the disarm's existence.

**Status here: satisfied.** All four switches are exercised (`TestFailureLadder_EachNewBehaviourDisarmsIndependently`,
plus the three-trigger table test from the 42P18 repair). ⚠ Verified by CONSTANT name — a first
check by literal string reported "1 of 4" because the tests reference `envDisable…` identifiers, not
the `"DISABLE_…"` literals. A coverage claim about env switches is a claim about how the tests
*reference* them.

**Open, and deliberately not decided here:** whether this becomes a general estate rule or stays
this lane's practice. It reads as general — but that is a ruling, and this RFC has had its four.
