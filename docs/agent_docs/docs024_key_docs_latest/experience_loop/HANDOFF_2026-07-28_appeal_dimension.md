# HANDOFF — teaching the experience loop to reason about APPEAL

**Opened 2026-07-28** by the `gauntlet_dead_cta` thread, at the owner's request:
*"Please put the whole UI to the experience loop to try and make it more enticing.
We may have to improve the experience loop to be able to handle this sort of
request."*

**Finding: it cannot, and this is a design boundary, not a bug.** Start a fresh
thread from this file.

## 1. The question that was asked, and the honest answer

The owner asked for three things on `vonc.com/tools/gauntlet`: make "Today's
Provocation" obviously the thing to respond to, make "Take a position" obvious,
and make the text bigger. Then: put the whole UI through the experience loop to
make it **more enticing**.

The first three were done by hand and are live (see §5). The fourth cannot be
routed to the loop as it stands.

## 2. Why it cannot — evidence, not opinion

`agent_definitions` where `type='experience-planner'` (live, `is_active`).

**The plan it composes has exactly five sections** (from the `compose` step's
`prompt_template`, quoted verbatim):

1. **Journeys** — "every step names: page, control (a real CSS selector), action,
   and the OBSERVABLE outcome. No step may end at `#`."
2. **Promise ledger** — "CTA copy → the page/state the destination must deliver."
3. **Data contracts** — what the feed must contain.
4. **MVP cut + LATER** — an ordered, gated step list.
5. **Acceptance criteria** — a fenced ```criteria block.

**Its five review seats**: `review_mvp`, `review_honesty`, `review_journeys`,
`review_contracts`, `review_feasibility`.

**Its entire acceptance vocabulary is five check types**:
`selector_exists`, `asset_loads`, `interaction`, `page_status_ok`,
`no_horizontal_overflow`.

Every one of those is a **binary existence-or-behaviour assertion**. The loop's
governing question is *"does this control do the thing it promises?"* — its HARD
RULES are all about not fabricating: no dead controls, no invented numbers, no
simulation. That is a question about **integrity**.

"Enticing" is a question about **hierarchy, emphasis, rhythm and restraint**.
There is no seat that would look, no section that could hold the answer, and no
check type that could assert it. Submitting the request today would return a plan
about whether the buttons work — which is already true and already verified.

**Do not "just try it".** A run costs credits and ~30 minutes of queue, and the
predictable output is a correct plan answering a question nobody asked.

## 3. What extending it actually requires

The loop is well-built for its remit; this is additive, not a repair.

**a. A sixth section, or a sibling plan type.** Appeal is not a journey, so it
cannot be smuggled into §1. It needs its own home — something like a
**presentation ledger**: for each screen, what is the ONE thing a visitor must
notice first, what is second, and what must recede. That is expressible as an
ordered list, which is the loop's native shape.

**b. At least one new seat.** `review_hierarchy` (does the stated first-thing
actually dominate?) and arguably `review_restraint` (is anything competing?).
Seat them with `099_SYNC_gate_roster.py` — **do not hand-patch the gate**; the
roster is mirrored and drift is the exact class the council reviews for.

**c. New check types, and this is the hard part.** The existing five are all
"does X exist / did X happen". Appeal needs *comparative* and *computed* checks.
The good news: the ones that matter are mechanical, and this workstream has
already written and used every one of them by hand:

| candidate check | how it is computed | already used at |
|---|---|---|
| `computed_style` | `getComputedStyle(sel).prop` vs expected | the type-scale change, 2026-07-28 |
| `relative_size` | fontSize(a) ≥ N × fontSize(b) — "the provocation must dominate its own label" | measured by hand today |
| `contrast_ratio` | WCAG luminance of fg vs actual painted bg | `bugs_open/083` council round; the vonc palette split |
| `first_paint_order` | bounding-box order of named elements at a viewport | — |
| `not_below_fold` | is the primary action visible without scrolling, per profile | — |

`contrast_ratio` is the strongest first candidate: it is fully objective, it has
a published threshold (4.5:1 body, 3:1 large), and **it would have caught a real
live defect** — vonc's link colour sat at 3.71:1 for weeks and nothing flagged it.

**d. A rail, or this becomes a taste engine.** Every check above must reduce to a
number with a threshold. If a seat's objection cannot be expressed as "X must be
≥ N times Y" or "≥ 4.5:1", it does not belong in an automated loop — it is a
human design judgement and should be routed to the owner, not adjudicated.

## 4. Warnings for whoever builds this

- **The 16-seat gate counts filtered seats as ABSTAINED.** `"abstained": 8` of
  `"reviewers": 8` does NOT mean nobody voted — read `body.reviews[]`, never the
  counters. This misled me once already (corr `e004fd81`).
- **The compose prompt is near its length ceiling by design**: "aim for about
  14,000 characters, never exceed 20,000. A plan that hits the cap is DESTROYED."
  Adding a sixth section means *cutting* elsewhere or raising `max_tokens` — and
  a truncated plan fails outright with `stop_reason=max_tokens`.
- **A bare re-fire after an escalation starts blind** — compose/load_context read
  no prior `council_report` history. Fold previous objections into the Decisions
  channel first (the 197/207/209 pattern).
- **Dispatch latency is ~30 min**, and a submission fired within ~300s of a
  chassis restart is silently dropped. Check the pod's `startTime`.

## 5. What was done by hand instead (live now, so don't redo it)

`vonc.com/tools/gauntlet`, component `gauntlet-interface`
(`5da50747-7936-4b8f-a66d-c1ea98919c75`, **used on 1 page — checked before
editing**):

| element | was | now |
|---|---|---|
| "Today's Provocation" | 13px | **20px** |
| the provocation itself | 32px | **46px** (35px mobile) |
| provocation body | 19px | 23px |
| "1 · Your position" | 13px | **24px** |
| the input | 18px | 21px |

Verified live, desktop + mobile, no horizontal overflow. Site-wide the base was
already raised 16→20px earlier the same day.

**This is precisely the work that has no automated guard.** Nothing in the fleet
can tell whether it made the page more enticing — only that it did not break.
That gap is what this handoff exists to close.

## 6. First move for the new thread

Do NOT start by building seats. Start by answering: **can `contrast_ratio` be
added to the browser-runner's check types alone?** It is one check, fully
objective, needs no new seat, and would immediately earn its keep across the
fleet (the brochure workstream measured 101 unreadable pairs on one site).
If that lands cleanly, the machinery for the comparative checks is proven and the
seat work becomes worth doing. If it does not, that is the real blocker and it is
better found on one check than on six.

Related: `bugs_open/083` §4 candidate 2 is a separate fleet-wide item; and the
browser-runner's `stepDelay = 300ms` (`run_checks_action.go:199`) still makes the
existing acceptance ladder unable to test an 8–23s AI call — **that is a
prerequisite for any acceptance work here.**
