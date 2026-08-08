# HANDOFF 2026-08-08 — decisions landed, prompt half proven, and the carrier gap found: candidate 1b is the work now

**Supersedes `HANDOFF_2026-08-07b_continue_here.md`.** Written ~11:00 BST;
liveness verified at write time. Read RFC_016 §3b and NOTES 2026-08-08 before
believing anything the 08-07 docs say about "planner uptake" — that
conclusion was corrected (WRONG_CALLS 2026-08-08).

## State in one paragraph

The owner decided all three RFC_016 questions (§5, recorded 2026-08-08):
section-entry rule RATIFIED (+§1a scope clarification), sliced order
APPROVED, §3a = option (a). Seed `333` (rule 17: object form + facts key
mandatory per page) is LIVE, applied + recorded, commit `d6e9dcf06`,
rollback = `agent_definitions_bak_333`. The compliance replan under it
(corr `1cb17b11`) proved at the RAW emission that the prompt half WORKS
(index stat-band = F1+F2+F4) — and then failed at write on `bugs_open/215`
(canonicalised page-name collision, pre-existing, transactional, no damage;
plan `8ee5807b` from 08-07 still current). Diagnosing the failure exposed the
real blocker: **`reconcilePlanWithRealised` Pass B2 discards the LLM's
section entries — and the fact assignments inside them — for every DEPLOYED
page** (v3_site_actions.go:3031, header :5118). Nearly the whole site is now
deployed, so candidate 1 as shipped cannot reach the motivating pages.
Candidate **1b** (sketched RFC_016 §3b, NOT implemented) is the work:
(i) prompt — deployed pages re-emit their realised section list verbatim,
facts assigned to those names; (ii) Go — Pass B2 carries facts onto restored
sections by component-name match, misses logged durably.

## Open, in the order I would take them

1. **Fix `bugs_open/215`** (dedup by canonical name inside WriteSitePlanAction,
   stub loses to composed entry, ~20 lines + unit test — the failed run's raw
   pages array is a ready fixture). It gates every clean observation replan,
   and rule 17 v2 plausibly raises the collision odds. Platform code →
   council gate (can ride the Slice B round or go alone — it is independent).
2. **Design + build candidate 1b.** Both halves small; the Go half touches a
   heavily bug-laden merge (001/037/050/051 lineage — read Pass B/B2's header
   first, and the section-collision passes). Mutation-test the name-match
   carry. This is architecture-adjacent: it goes IN the Slice B council round
   (RFC_016 §3b says so), never slipped into a bug patch.
3. **Slice B round, revised**: draft is
   `COUNCIL_DRAFT_slice_b_2026-08-08.json` (HOLD note inside; do not submit
   as-is). Add 1b's edits + a fresh compliance observation (post-215-fix,
   post-1b replan: expect facts to SURVIVE onto restored sections). Owner's
   read of the v4 plaintext still owed — gates 330's apply only.
4. **After the round + human read**: un-`_HOLD` 328/330 → apply 328 then 330 →
   rebuild flagged pages → census: overlap pairs fall on fundamentallyai,
   five fact-blind sites do not move.
5. Standing: Monday 08-10 contact-sheet cron; `bugs_open/214` fix candidate 1
   (unowned, independent).

## Traps beyond the 08-07b list (all still apply)

- **Completed orchestration rows expire in ~24h.** Pin any raw `llm_plan`
  evidence into a doc THE DAY you cite it. The r1 mechanism claim is
  permanently unverifiable because this wasn't done.
- **`collected_data->'validate_plan'` is post-merge, not the emission** — the
  §9 entry (2026-08-08) is the general form. Raw = `llm_plan.result`.
- **A replan currently dies whenever the LLM emits both a stem and its
  canonical page name** (215) — a FAILED write_site_plan with SQLSTATE 23505
  on `idx_site_plan_pages_name` is THAT, not concurrency.
- The permission classifier transiently denied one kcat dispatch this
  morning; the identical retry passed. Transient, per its own message.
