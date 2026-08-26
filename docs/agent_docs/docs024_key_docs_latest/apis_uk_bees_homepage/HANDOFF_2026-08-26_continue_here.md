# HANDOFF 2026-08-26 — continue here

**Supersedes `HANDOFF_2026-08-25_continue_here.md`** (whose top correction block records the
08-25 owner split — still worth reading for the page's traps; its Google half is dead, see below).

> ## ▶ ONE-LINE STATE
> The page is done and locked. **Everything Google left this lane on 08-25** (owner ruling;
> `analytics_gtm` is the cold-start for tracking, `bugs_open/397`). The lane's remaining build,
> **per-section subjects, SHIPPED on 08-26**: Go + migration committed (`35905c547`), migration
> 638 applied, council **round 1 REVISE → round 2 RESUBMITTED same day** (`4bd35ed8`, same correlation —
> verdict pending), three config seeds `_HOLD` awaiting the next chassis roll — nothing else is
> open here except the image-accuracy A+C build, not started.

## 1. Per-section subjects — what a next session owes, in order

Full design + falsifiers: `PLAN_2026-08-26_per_section_subjects.md` · register **PBP-049** ·
council submission `scratchpad` copy is gone with the session — the artifacts are keyed under the
correlation.

1. **Read the council verdict — ROUND 2** (round 1 came back REVISE at 09:20 UTC, gating seat
   `prior_art_librarian`; all eight objections were answered the same morning — tests + seed
   guards in commit `52085b410`, per-seat answers + previously-unreachable evidence in the
   resubmission — and round 2 went out ~11:35 BST on the SAME correlation after two Kafka
   publish failures, evidence in `bugs_open/040`). Budget ~30 min from dispatch; a missing row
   is latency, not a drop — never resubmit on that evidence:
   `SELECT current_step, status FROM orchestration_states WHERE
    collected_data->'input_data'->>'fix_correlation_id' = '4bd35ed8-5f72-4a2f-9cbf-3860847837f4';`
   then `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;`
   - APPROVED → nothing to do; commit `35905c547` carries `Council-Submitted:` and `098` credits
     it automatically. REVISE → revise and resubmit with `RESUBMIT_CORR=4bd35ed8…`. REJECTED →
     read the guardian's contained alternative; the code is on the shared branch, forward-only.
2. **After the next chassis roll** (any session's — a roll ships this commit), apply the seeds
   **in order, by hand** (they are `_HOLD`; the runner skips them):
   `639` (wiring) → `640` (planner rule 17) → `641` (writer prompt v5).
   Every file carries drift-guarded anchors and aborts rather than mis-applies.
3. **`641` is double-gated: the owner must read the inserted block first** (quoted in full in the
   seed header — 4 lines). RFC_016 §5.2: the v4 approval attaches to its committed text and voids
   on edit. Do not apply 641 on an old approval.
4. **Then un-defer apis.uk's two `content_rewrite` items** (swarm, pollination — their rows name
   this chain as the unblock condition). ⚠ The locks refuse page re-renders by design, and adding
   a section needs plan surgery or a replan — that step is a decision, not a mechanical follow-on:
   write the two subjects into `site_plan_sections` for apis.uk (or replan), un-defer, and expect
   to settle `pages.build_status` afterwards (a render re-queues; settle → verify → settle again).
5. **Falsifiers before claiming anything works**: a plan with no subjects must be byte-identical
   end to end (NULL column, `omitempty` items); a replan with a repeated component must produce
   DISTINCT non-NULL `site_plan_sections.subject` on the repeats; after 641, same-named sections
   must differ in TOPIC, not wording. Adoption query (also the copy_quality lane's control):
   `SELECT count(*) FILTER (WHERE subject IS NOT NULL), count(*) FROM site_plan_sections sps
    JOIN site_plans sp ON sp.id=sps.plan_id WHERE sp.is_current;` — 0/N until the chain is live.

**Stated gaps, so nobody re-derives them:** the Pass B/B2 carry's object-realised arm is untested
(realised lists are plain strings everywhere today; the string arm is mutation-proven); the
carry's `unmatched` list stays facts-worded and drops an unmatched subject silently; seeding
`page_components.content_brief` from the subject at save time was considered and deferred.

## 2. Known-red at HEAD, not ours — expires when the 396 lane acts

`TestFindingCodeScanEveryWriteIsRegistered` fails on **clean HEAD** (actions package):
`WORK_ITEM_STATUS_OVERRIDE_REFUSED` shipped in `2b46afbe6` with no `finding_code_registry.json`
declaration. CONTRIB filed into `deferred_work_item_park/`. Re-establish before trusting:
`scripts/verify-head-builds.sh --test ./platform/orchestration/actions/ 2>&1 | grep OVERRIDE` —
no hit = fixed, delete this section.

## 3. RFC_022 parity — settled, and the lesson

`35905c547` grew plan_sections' Optional list 7→8 and I did NOT regenerate the cron literal in
the same commit — the **333 lane caught it at HEAD before I did** (parity test). Settled in
`339474ca4`: literal regenerated from a committed-HEAD extract, overlay re-applied, verified at
the mounted ConfigMap (`plan_sections: 8` with an unchanged neighbour as control). **If you touch
any ActionInputSpec Optional list in this lane again, check.py + overlay ride the same commit.**

## 4. Everything else this lane once held

- **Google/tracking** → `analytics_gtm/HANDOFF_2026-08-25_continue_here.md` (owner split, verbatim
  in the 08-25 handoff's correction block). apis.uk is in their `c2` rebuild wave's bucket B; they
  were asked to tell us when it runs so we re-verify the served bytes.
- **Image accuracy A + C** — designed 08-24 (see the 08-25 handoff §3 bullet 3), owner-approved,
  NOT started. D is live. C is agent config (`execute_vision_prompt` on `visual-design-auditor`),
  canary one apis.uk image first.
- **The page itself**: 200 / 67,877 B, `<h1>A closer look at bees</h1>`, 7 permanent locks,
  `tools.apis.uk` probe via `POST /api/v1/tools/gauntlet/round` (root 404 is by design).
  `[UNEXPLAINED, 08-25]` `sites.build_status='pending'` while the page row is `deployed` — no
  served effect, left alone, still true when checked 08-26 morning.

## 5. Design checks are visiting apis.uk again — measured, and the source is now settled

> **CORRECTED 2026-08-26 ~11:00:** the 00:40 visit was the **improvement-loop** (owner re-enabled
> it ~21:18Z on 08-25, loanzy lane's phased plan; it dispatches design-discovery + completeness as
> children — apis.uk got full cycles at 00:39, 04:47 and 08:40Z), NOT the rotation, which
> `webdesign-tool-rebuilds` re-enabled separately at 09:20Z (`bugs_open/401`). **Both are active
> now**, so expect findings roughly every 4h from the loop plus rotation visits. Their lane
> corrected its own NOTES on our timestamps (their commit `7baa7a4f1`).

Measured here rather than assumed: **six findings landed on apis.uk at 00:40:15–56 UTC**,
all status `detected`, **all with `handler_agent` empty — so NONE are promotable** by
`detected-item-promoter` (it requires `COALESCE(handler_agent,'')<>''`, plus the mig-629 origin
door). No auto-dispatch today; re-check that premise if anyone maps handlers onto these types.

| finding | read |
|---|---|
| `head_essentials_missing`: index missing `skip_link`, `footer` | **the footer half is BY DESIGN** (owner: no footer, no email — the empty `site_components.footer` row is the mechanism). Do NOT let anything "repair" it. The `skip_link` half is a real accessibility gap — a legitimate small fix, but see the trap below |
| `image_url_404` ×2: chrome references `/assets/images/favicon.png` + `og-card.png`, no active asset | real, mild (404 favicon/og-card). Fix = deploy the two assets, NOT edit chrome |
| `prerequisite_missing` ×2 (page_research, feed_sources) · `structure_floor_unmet` (1 of 6 structures) | the single-page-by-design shape failing fleet norms — expected; annotate rather than "fix" if they nag |

⚠ **THE TRAP for any of these:** the locks guard `page_components` only. Any repair that
regenerates chrome (head/footer render) **strips the GTM tag** — apis.uk is `bugs_open/397`
bucket B (artefact-only) until `analytics_gtm`'s `c2` applies. So: assets over chrome edits,
coordinate anything chrome-shaped with that lane, and settle `pages.build_status` after any
repair (a render re-queues).
