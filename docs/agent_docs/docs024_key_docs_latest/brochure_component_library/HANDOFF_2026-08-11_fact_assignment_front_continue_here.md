# HANDOFF 2026-08-11 — fact-assignment front: the census is AUTHORISED. Cold-start for a fresh chat

**Supersedes `HANDOFF_2026-08-10b_…` as the entry point — but that file remains
the detailed playbook: its §4 (replan/rebuild/census steps + queries) and §5
(traps) are NOT restated here. Read this file, then execute 08-10b §4.**

Written 2026-08-11 morning, after the owner ruled the four follow-up decisions
(`DECISIONS_2026-08-11_four_follow_up_rulings.md`) and rolled **v1.0.1284**
(verified: all three detection literals on both replicas, NEG control 0,
pods `agent-chassis-7c9d5f74b9-*` started 09:23Z — the ~300s dispatch cooldown
is long past).

Site id: **`199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`**.

## What changed since 08-10b

1. **The replan is AUTHORISED (owner, 08-11): run it now.** The phantom-404
   side effect is an accepted cost — the sweep front hand-archives afterwards.
   Coordination with the sweep front's handoff before dispatch is still
   mandatory (08-10b §4 step 1); "authorised" ≠ "uncoordinated".
2. **`features_open/012` is scheduled**: the field-based recompose fix is this
   lane's next round **immediately after the census** — when you finish the
   census, that round is your next job, not a judgement call.
3. **Platform ruling recorded (nested contracts)**: nested additions to a
   declared object input are register-named, not re-declared. PBP-037 now
   carries the `section_plan` seam's enumerated nested contract — if the census
   work adds any key to `sections_ready` items, name it there in the same
   commit.
4. **The invented-commitments clause is approved** — see the new-seeds queue.

## The job queue, in order

1. **Execute 08-10b §4 verbatim**: sweep-front check → work-item queue check →
   dispatch the replan (NO `recompose_pages`) → read the plan (composition
   preserved? `assigned_fact_ids` landing? carry/absent error rows ~0?) →
   rebuild flagged pages → scoped sections state only their assigned facts →
   **the disconfirming half: the five fact-blind sites must not move**.
   Close the loop in NOTES/README; the fact-overlap pair count on
   fundamentallyai must FALL (9 pairs pre-round — re-derive, don't trust).
2. **The 012 round** (scheduled): surface `recompose_pages` to the planner as a
   prompt-visible field. The spec travels at `input_data.spec.recompose_pages`
   in `collected_data`, so exposure may be config-only (an `input_field` on
   `plan_site` + a template block marking released pages "REDESIGN REQUESTED —
   propose a new composition"). Own council round. On success, retire the
   prose escape's load-bearing status in seed 362's instruction (a follow-up
   seed) and update the LANDMINES entry + `features_open/012`.
3. **Two small seeds, buildable any time** (each its own narrow round or one
   combined small round — submitter's judgement):
   - **Commitments clause** (owner-approved, ruling 4): extend STRICT RULE 5 of
     `page-content-writer`'s prompt with: *"…and do NOT invent commitments,
     guarantees, warranties, or service promises (response times, refunds,
     availability) not stated in this prompt."* Derive the template byte-exact
     from the LIVE row (the seed-330 pattern: base64 + jsonb_set, drift guard,
     em-dash census; the nested path is quoted in PBP-037/seed 330). Do NOT
     edit the v4 file in the brochure sql dir and seed from it — it is already
     one drift behind live the moment anything else touches the row.
   - **Guidelines-corpus seed** (ruling 3's seat-visible half): add the
     nested-contract ruling to the corpus the guidelines seat reads — the
     `sql_for_agents/247_council_guidelines_column_rename_rule.sql` pattern.
     Until it lands, the ruling lives in DECISIONS_2026-08-11 + PBP-037 only.

## Standing state (verified 08-11, details in 08-10b §1)

Seeds 362/328/330 applied + row-verified · round APPROVED corr `a06ff850`
(verdict pinned) · three durable detections live and twice roll-survived ·
compliance read countersigned · consumption still ZERO until step 1 runs ·
rollback tables `bak_362/_328/_330` exist.

## Owner items still open (context, not blockers)

- 215 revisit trigger: `PLAN_PAGE_MERGE_LOSSY` count non-zero ⇒ revisit
  richer-wins. Currently 0.
- Compliance findings 3.5/3.6 (edit-mode legacy claims; testimonial
  trade-dress): recorded, unscheduled.
