# RFC 057 — `HandlerCanWriteField` is a new shared contract, and its roster is a NEGATIVE claim nothing re-checks

**Status: DRAFT, raised 2026-08-25 by the `vigilant_designer_offer_analysis` lane.
RETROSPECTIVE (the RFC_002 shape): the mechanism is committed (`af3194204` + `a48c5c942`,
`Council-Submitted: 021cb965`, APPROVED round 1) by the `bugs_open/395` lane and is inert until the
next roll. Register entry **WII-035**.**

**This RFC asks to reverse nothing.** It exists because the council's `architecture` seat named the
trigger explicitly and the 395 lane, correctly, did not decide it alone:

> *"`HandlerCanWriteField` is a shared exported contract with two agreed-upon cross-lane consumers
> from day one — meets the architecture-review trigger test directly and is not covered by the
> RFC_022 opt-in-field exception (which requires zero live consumers)."*

**The seat is right and I am the second consumer**, so the judgement is partly mine and this is my
half of it. The 395 lane took the seat's stated minimum (a contract note in the file) and left the
RFC to me rather than pre-empting it.

---

## 1. What the thing IS, in plain terms

A **work item** is one recorded defect. A **handler** is the agent sent to repair it. A finding can
state, mechanically, which page field would have to change for it to be fixed (`CLM-024`).

`HandlerCanWriteField(handlerAgent, field) (canWrite, known bool, why string)` answers one question:
**can the agent we are about to send actually write that field?** If not, the finding is filed as a
`capability_gap` rather than dispatched at an agent that structurally cannot act.

It has **two callers by design**, in different lanes:

| caller | what it does with the answer |
|---|---|
| the routing seam (`write_audit_findings`, WII-035) | does not ROUTE the finding at an incapable handler |
| the predicate emit gate (`CLM-024`, this lane, not yet built) | STAMPS `field_writable:false` on the stored predicate |

## 2. Why it needed a shared answer rather than two — the near-miss is the argument

The two guards were first agreed **in a chat message between the lanes**, as two independent
implementations. They would have **CANCELLED**: the emit gate was going to REJECT a predicate over an
unwritable field, which moves it into `acceptance_predicate_rejected` — so the routing guard, keying
on the live predicate's `field`, would have gone blind, routed the finding at the incapable handler
exactly as before, and **both guards would have reported success over a mutual blind spot.**

Caught in correspondence, not by either implementation. The fix was structural, not procedural: one
exported helper, and the emit gate **stamps instead of rejecting** so nothing is destroyed and the
composition is order-independent.

⚠ **The council objected to the coordination mechanism itself, and was right to:**

> *"this is the mitigation for exactly the failure mode the founding incident describes (two lanes
> independently reinventing the same mechanism). If that coordination is informal (a chat message,
> not a shared merged branch), there's a real risk both lanes land competing versions."*

That is now mechanical — `TestThePageFieldWriterRosterIsDefinedExactlyOnce` fails the build if
`pageFieldWriters = map[` appears more than once in the package's non-test source. **But the
objection generalises past this instance and is the reason this RFC is worth reading:** two lanes
agreeing in chat is not a control, and the estate has no seam for "these two changes must land
together". Ours worked because one of us happened to raise it.

## 3. The architectural question: DECLARED or DERIVED — and it has a measured answer

The roster is hand-maintained: a map of `(handler, field) → cannot write`, each entry carrying its
own evidence. The obvious objection is that this could be **derived** — ask `agent_definitions` which
handlers have a step that writes the column, and never maintain a list at all. A derived answer
cannot go stale, which is the roster's whole weakness (§4).

**It cannot be derived from config, and this is measurable rather than arguable:**

`[MEASURED 2026-08-25, live]` `upsertPage` (`site_db_actions.go:1235`) writes `pages.meta_description`
— fill-blank only — and it is reached through the action `sync_pages_to_db`. Of the live agents
carrying that step:

| | |
|---|---|
| live agents with a `sync_pages_to_db` step | **3** |
| of those, agents naming `meta_description` anywhere in their config | **1** |

**So 2 of 3 agents that can write the column are INVISIBLE to any config-text derivation.** Which
columns an action touches is a **Go** fact, not a config fact, and no amount of reading
`agent_definitions` recovers it.

**Conclusion: declared is right.** But it follows immediately that **the staleness audit §4 asks for
must resolve steps through the ACTION REGISTRY, not by grepping config text** — otherwise the audit
inherits the same blind spot and goes green while being wrong about 2 of 3 agents. That is the single
most useful thing in this RFC for whoever builds it.

### 3a. ⚠ AND THE INSTRUMENT'S ERROR DIRECTION MUST MATCH THE CLAIM'S DIRECTION

*Added 2026-08-25 by the `bugs_open/395` lane, whose point this is, after re-deriving the premise the
sound way. It is the half §3 was missing, and without it §3 sends the next author to the right table
with the wrong query.*

The sound instrument searches **action names**, whole-config so no nesting can hide one:

```sql
default_config::text LIKE '%"sync_pages_to_db"%'
  OR LIKE '%"save_page_meta_description"%' OR LIKE '%"apply_adoption_plan"%'
```

`[MEASURED 2026-08-25, live; re-run independently by this lane]` → **build-site-planner,
meta-description-backfiller, pageflow-builder, site-adoption-agent, site-work-orchestrator**, plus
**council-gate** and **fix-proposer**, which merely QUOTE those names in prompt text.
**`page-build-handler` and `page-content-writer` are absent** — which is the whole premise of
`bugs_open/395` §9, `WII-033`'s `PromotionOwes`, and `WII-035`.

**Those two false positives are the RIGHT direction, and that is what makes the result usable at
all.** This roster asserts a NEGATIVE — *this handler CANNOT write this field*. An
**over**-inclusive instrument that still returns nothing for `page-build-handler` is strong evidence
for that claim. An **under**-inclusive one returning nothing is no evidence whatsoever, because
absence is exactly what it produces when it is broken.

⚠ **So the same query STOPS BEING FIT FOR PURPOSE the moment the audit's question is phrased the
other way round.** The drift audit §4 asks for is a POSITIVE claim — *has a rostered handler GAINED
the capability?* — and answering that with an over-inclusive search means the first prompt mentioning
an action name raises a false alarm and, worse, a maintainer silences it. **Whoever builds it must
state which direction their instrument errs in and check that it matches the direction of the claim
being made.**

> **The general form, and it is the transferable half of this whole RFC:
> A CONTROL PROVES YOUR INSTRUMENT IS WORKING. IT CANNOT TELL YOU THE INSTRUMENT IS POINTED AT THE
> WRONG THING.** The 395 lane reached §3's conclusion the hard way: answering a `debugging` seat's
> objection ("asserted from a private code read, not independently checkable by SQL") with a SQL
> check that carried a correct demand control and asked the wrong question — *does this config
> mention the column* — then read the answer as *can this agent write the column*. The control passed.
> The two questions differ by two thirds of the population.

## 4. ⚠ THE OPEN RESIDUAL, conceded by its author rather than defended

The council's `constitution` seat:

> *"Staleness re-check for roster entries is described in prose (risks section) but not wired to any
> scheduled verification — the exact enforcement gap 320 §9 was filed about, now reproduced one layer
> down."*

**It is right, and the shape of it is almost funny:** the mechanism exists *because prose does not
bind* — that is the whole finding of the landmine both lanes filed today — and its own staleness rule
was then written in prose.

The roster is a **NEGATIVE capability claim**, and negative claims go stale **BY ADDITION**: the day
`page-build-handler` gains a meta-description step, the roster keeps parking findings that have
become fixable. **The failure direction is safe** (it refuses loudly rather than filing a doomed
item) and it is still wrong.

**The named follow-on is the WII-031 treatment** — declare in Go, then ship a live-drift audit that
fails when a rostered handler has GAINED the capability the roster denies. Per §3 it must read the
action registry. **Nobody has built it, and neither lane has claimed it.** This RFC's ask is that it
be treated as a condition of the roster growing, not as a nice-to-have: **one entry is a measurement,
ten entries with no audit is a stale map with an enforcement mechanism attached to it.**

## 5. What is NOT being asked

- **Not to reverse or re-review WII-035.** It is approved, mutation-proven, and its route-matrix test
  is a biconditional over the whole category universe with both arms asserted non-empty — a standard
  this RFC would like to see copied, not questioned.
- **Not to widen the roster.** It is opt-in per field; an absent field means NOT MEASURED and routes
  exactly as before. `known=false` must never be read as either capability or incapacity.
- **Not a decision on my emit-side stamp**, which is unbuilt. Its own hazard is recorded in the 395
  lane's route-matrix finding and repeated here because it is the same class: **a stamp that lands on
  a record some other gate has already parked can destroy that gate's dedup key and silently merge
  two different gaps.** Their guard would have re-stamped Rule 3's own `cta` capability_gap —
  identical `item_type`, so it looked correct — and only a biconditional test caught it.

## 6. The question for the owner

1. **Is a hand-maintained negative-capability roster acceptable without its drift audit?** §3 says
   derivation cannot replace it; §4 says nothing re-checks it. Today there is one entry and it is
   measured. The risk is entirely in the growth.
2. **Does the estate want a seam for "these two changes must land together"?** §2's near-miss was
   caught by correspondence. The build-time single-definition test fixes this instance; it does not
   generalise, and the founding incident this mechanism exists to prevent is precisely two lanes
   landing competing versions.

## 7. Sources

- `platform/orchestration/actions/write_audit_findings_action.go` (`HandlerCanWriteField`,
  `pageFieldWriters`, `withUnwritableFieldGuard`); register **WII-035**
- council report, corr `021cb965` — `architecture` and `constitution` seats
- `bugs_open/395` §9, `bugs_open/320` §4/§5/§9, register **WII-033** / **CLM-024**
- `LANDMINES.md` — *"a rule that exists only as prose in a bug file gets broken by the next producer"*
  (filed jointly by both lanes, 2026-08-25; §4 is that landmine firing on its own remedy)
- RFC_022 (the narrowing whose third condition fails here), RFC_055 (the sibling accumulation question)
