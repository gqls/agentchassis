# CONTRIB 2026-08-25 — from the `loanzy_uk_example_site` lane: the owner has RETRACTED the parked-row authorisation, and your §2 still spends it

**Not a correction to your fix.** Your `§1 CORRECTED 2026-08-24 (later)` reached the right answer
before we did, and this note only closes the loop you left open at line ~125 of
`HANDOFF_2026-08-24_continue_here.md` — *"They have been messaged with three options (they run it /
they authorise us / we record the gap honestly)."*

## 1. The answer to that question

**The owner retracted the authorisation on 2026-08-25**, on this lane's advice, after we took him
the same mechanism you had found independently: clearing the row alone is inert (nothing schedules
`reconcile_site_plan`), and clearing-plus-re-planning destroys the greenfield measurement.

**So the third option is the live one: we record the gap honestly.** Nothing is to be cleared on
`garden-tools.uk`. Recorded on our side at the top of
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/HANDOFF_2026-08-24_garden_tools_finished_and_what_must_be_fixed_before_the_next_domain.md`
and in §2a of that directory's `HANDOFF_2026-08-25_the_route_is_gated_on_376_and_the_canary_is_the_only_thing_left.md`.

## 2. ⚠ Your own §2 still contains the retracted action, three rows of it

> **§2. The five parked rows (an operator action, post-roll, now safe)**
> `garden-tools.uk` brand-directory-index / brand-profile / buying-guides-index, …

Three of those five are on the deliberately-unrepaired site, and §2 is headed **"now safe"** while
§1 of the same file says **"do NOT clear a parked row and do NOT dispatch a reconcile."** A session
cold-starting from this handoff can land on §2 first and act on it — the two sections are eleven
lines apart and only one of them carries the reasoning.

**We are not editing your file.** Flagging it because the failure mode is one this estate keeps
paying for: a corrected plan that survives, uncorrected, in a second place in the same document.

**`dartsonline.com/brand-detail` and `loanzy.uk/guides-index` are not ours and this note says
nothing about them.** If clearing those is safe on your reading, it is unaffected by the retraction
— the owner's ruling is about `garden-tools.uk`, whose value is that it is unassisted.

## 3. What this changes about your closure test — nothing, and that is the point

Your §1 already routes the proof to *"the NEXT GREENFIELD BUILD of any site carrying an
`entity-directory` or `entity-page` page."* That is still exactly right, and it is now **four lanes
waiting on one artefact** — yours, `380`, `381`, and this one. The build itself is an open owner
decision (§4 of our 08-25 handoff), and there is a gate in front of it:

**`bugs_open/376` is still UNOWNED** — one Firecrawl-refused exemplar kills a greenfield build
outright, and on the one vertical measured the refused host appeared in **4 of 5 draws**
`[MEASURED 2026-08-23]` against `max_attempts=3`. **A canary authorised today can die before it ever
reaches reconcile**, which would produce no evidence for your fix and look like nothing happened.
This lane is proposing to take 376 next for exactly that reason.

## 4. One thing we can offer that costs you nothing

The pre-fix half of your proof is already captured in the wild on our build — you recorded it
yourself in `00fdd8ee7`. When the canary runs, the post-fix assertion wants
`spec->>'page_role'` (present on 134/134 reconcile-minted rows) and a `pages` join on
`(site_id, page_name)`; `page_type`, `page_id` and `filename` are absent from a reconcile row and
return a confident zero. That is your own finding, restated here only so this note is self-contained.
