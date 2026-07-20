# HANDOFF — Claims verification: resume from here

**Start a fresh chat from exactly this file.** It is the cold-start entry point:
current state verified against the live system on 2026-07-20, the one action that
remains, the traps, and the open decisions.

**Read in this order:** this file → `SUMMARY_2026-07-19_claims_verification_complete.md`
(what to tell someone else) → `RUNBOOK_claims_verification.md` (every command, with
its gotcha) → `PLAN_2026-07-16_claims_verification.md` (decisions and why) →
`NOTES_claims_verification.md` (the technical log, including missteps).
`README_where_we_are.md` is the OWNER's plain-prose log — **append only, never
rewrite**. The originating design is `SPEC_claims_verification.md`.

---

## What this is, in one paragraph

The platform writes site copy with an LLM and nothing ever compared a **claim** to
**evidence** — so it shipped invented case studies, an invented founder, a
fabricated agent taxonomy, and (on another site) fabricated prices with legal
exposure. This layer gives each opted-in site a machine-readable **evidence base**
(verified facts with sources and tolerances, that site's own banned fabrications,
and the entities its copy may name), then checks claims against it at three points:
a build-time gate, a post-deploy discovery check, and the writer's prompt. A
judgement lane (LLM) catches unsupported prose, and a freshness pass keeps
live-sourced numbers current. **Every judgement terminates at a human — there is no
code path that rewrites a factual claim.**

Pilot site: leopardessconsulting.co.uk, `4851f6fc-71cf-4160-a270-e03d6d3e0732`.

## State — verified live 2026-07-20

| Phase | What | State |
|---|---|---|
| V0 | `evidence_base` as structured data | **LIVE** — spec row `00b01b38`, rev 6, pinned, 18 facts (18 with `writer_line`), 19 banned patterns, `writer_block_managed=true` |
| V1a | Build gate in `validate_page_content` | **LIVE in prod** |
| V1b | Discovery check `unverified_claims` | **LIVE** — enabled on `quality-discovery-agent` |
| V2 | Writer whitelist in the page-content-writer prompt | **LIVE** — proven 5/5 prompts, 0 blockers |
| V3 | `claims-auditor` agent (prose lane) | **LIVE, not scheduled** — both paths proven (clean site → `[]`, no items; no-evidence-base site → zero LLM calls) |
| V4 | `refresh_evidence_base` (freshness) | **LIVE** — in image `v1.0.1139`, dry-run proven against all 3 evidence-base sites, scheduled task `evidence-freshness` applied (daily, enabled) |

Leopardess is clean: last full scan 0 findings / 95 components; `content_data`
sweep clean; all nine original resurrections fixed and deployed; the poisoned
specs that caused them are cleaned.

## ⇢ START HERE: one live finding needs an owner ruling

**V4 ran and immediately caught a real problem.** Dry run 2026-07-20
(orchestration `8ec0743e`), 3 sites swept, 9 sql-facts checked, 0 errors:

| fact | published | live | verdict |
|---|---|---|---|
| **`C1-records-verified`** | **2,767** | **2,291** | **DRIFTED DOWN — exact tolerance breached** |
| C1-records-enriched | 937 | 937 | fresh |
| C1-ch-vet-mirror | 5,798 | 5,798 | fresh |
| C2-feed-items-collected | 6,262 | 6,663 | updated (gte, grew) |
| C2-feed-items-scored | 5,228 | 5,585 | updated (gte, grew) |
| C4-agent-definitions-catalogue | 157 | 166 | updated (gte, grew) |
| C4-agent-definitions-active | 60 | 63 | updated (gte, grew) |
| C4-orchestration-state-records | 90,790 | 102,863 | updated (gte, grew) |
| C6-sites-deployed | 9 | 11 | updated (gte, grew) |

**The ruling needed.** leopardessconsulting.co.uk publishes *"2,767 business
records verified against Companies House"*. The live count is **2,291**. Verified
directly — `business_intel.businesses` now reads `verified: 2291, dismissed: 874,
pending: 238, seed_import: 16`, so ~476 records were reclassified OUT of verified
since 2026-07-16. **The site is currently overclaiming.** Options: republish with
the new figure; re-word to a growth-proof form ("more than 2,000 records verified",
which `gte` tolerance would then protect); or investigate why 874 were dismissed
before changing any copy. **Do not auto-edit the copy** — this is the human-terminal
case the layer exists to create.

Where the claim appears: grep the deployed pages for `2,767` (RUNBOOK §4 shows the
`content_data`-vs-`rendered_html` query — fix `content_data` too, or the next
re-render restores it).

**Note on how the register behaves here:** on the first REAL (non-dry) run the
register re-syncs to 2,291 — for a sql-sourced fact the query IS the source of
truth — and raises a `stale_evidence` work item so a human rules on the COPY. So
after the first scheduled pass, expect the register to say 2,291 while the site
still says 2,767, with an item open against it. That is the designed behaviour, not
a bug.

**Also observed:** vonc.com and relojistas.com now have `evidence_base` specs
(0 sql-facts each), so other threads have begun opting sites in. The sweep already
covers them with no seed change.

## Open decisions — owner's, not the next thread's

1. **Cadence for the claims auditor (V3).** Recommendation now evidence-based:
   **schedule it where prose dominates the claim surface, leave it manual where
   numbers do.** Leopardess is numbers-dominant and self-describing, so V1 does the
   work there and V3 returned a literal `[]` — run it after significant content
   builds, not on a clock. gaswholesalers is the opposite: no machine-verifiable
   ground truth at all, so V3 is not a supplement there, it is the only lane that
   functions.
2. **Second site: gaswholesalers.com** (owner chose it 2026-07-20 over
   vetcomparison, which another thread is working). Recon done, question resolved:
   the site asserts a business the owner is not in — **all 174 measured operational
   assertions are false**. It is now the **cold-audit pilot**, with a measured
   expected result to grade against. See
   `PLAN_2026-07-20_gaswholesalers_second_site.md`. The repositioning/rewrite and
   its new AI-influence page are routed OUT to `features_open/006`; the deferred
   freemium chatbot is `features_open/007`. **This thread does not write that copy.**
3. **Re-file bug 019 to the diagnosis loop?** See below; needs `FORCE=1` or the
   intake item closed. My view: not worth it until `bugs_open/003` is fixed.

## Next build work for THIS thread (from the 07-20 owner direction)

1. **Cold-audit posture** — the layer cannot audit a site with an empty register
   (V3 gates on `facts_text`), which is backwards for the site that needs it most.
   Add `"posture": "cold_audit"` to satisfy the gate with zero facts and tell the
   prompt to report every unsupported operational assertion. Benchmark: ~174 on
   gaswholesalers, concentrated in `pricing-transparency` (19),
   `supply-terms-and-eligibility` (17), `who-we-serve` (17), `service-areas` (15).
2. **V5 — researched, cited, re-verifiable external facts. NOW SPEC'D:**
   `SPEC_V5_researched_citations.md`. Owner's requirement (2026-07-20): the site
   must "consistently use numbers that are verified from web deepsearch cited
   references, so not manual but part of the chassis' capability". The key finding
   is that the enforcement lanes already exist — V1 flags unregistered numbers, V2
   whitelists registered ones, V4 re-checks them — so this is an **acquisition**
   problem, not an enforcement one. The one genuinely new idea: a citation is
   verified **deterministically** by re-fetching the URL and asserting the stored
   verbatim quote still appears in it, which kills hallucinated references at
   acquisition time and gives free re-verification later. Build order in §6.
3. **Forward mode.** That page is the first chance to run the layer the other way
   round — register first, copy second. Every use so far has been remediation.
4. **Open question that shapes V5:** should the claims layer gate the planned
   chatbot's responses (`features_open/007`)? If yes, verification becomes
   *pre-response* rather than post-publish — materially harder. Do not plan beyond
   V4 without an answer.

## Bugs this workstream found (all filed, none fixed here)

- **`bugs_open/019`** — one truncated reviewer (`output_tokens == max_tokens`)
  voids an entire council round, discarding every other seat's review. Found
  submitting this workstream's own V4 change. **Independently reproduced by a
  second thread** in normal use. Independent diagnosis was ATTEMPTED and NOT
  OBTAINED (the run was eaten by 003) — the file says so explicitly; do not read it
  as loop-verified.
- **`bugs_open/003`** — spawn-lost-child-response. A fresh occurrence is appended
  (2026-07-19): it killed a diagnosis run, the sweeper did not retry
  (`retry_version=0` on an expired request), and the orphaned intake item makes the
  090 coverage check refuse the retry — a stall that blocks its own remedy.
- **Fixed en route:** `checkpoint_for_review` was documented since creation but
  never registered, so any workflow naming it failed with "requires a topic".
  Registry entry added (`0540698a4`). `refresh_evidence_base` registered in
  `06376bcbf`.

## Traps specific to this workstream

1. **Fix `content_data`, not just `rendered_html`.** `content_data` is the render
   source; fixing only the HTML means the next re-render restores the claim.
2. **Route copy fixes through `page_rerender`, never `needs_page`.** A
   `needs_page` regenerates via the writer, which can reintroduce a claim if any
   spec still carries it.
3. **`regexp_matches` without the `'g'` flag returns only the FIRST match per
   row.** A spec sweep looked clean after one pass and was not. This cost a full
   cleanup round.
4. **The specs feeding the writer are the real surface.** The original nine
   resurrections were caused by `content_direction` / `site_plan` / `strategy` /
   `briefing` literally instructing the fabrication — including a `writing_rules`
   entry telling the writer to cite "least-privilege IAM policies". Cleaning pages
   without cleaning specs regresses on the next rebuild.
5. **Banned claims stay OUT of the writer prompt** (don't-think-of-an-elephant).
   The whitelist goes in; the blacklist is enforced deterministically at the gate.
   Do not "helpfully" add the banned list to the prompt.
6. **A clean check leaves no positive trace.** Checks emit findings only. Verify a
   clean pass by `status=COMPLETED` **plus** zero new work items — never by looking
   for a success line.
7. **Track cluster work by the identifier the trigger gave you**, against the
   table that stores it (`orchestration_states.correlation_id`,
   `diagnosis_artifacts.correlation_id`, pod logs as fallback). **"No row" means
   "not yet", never "lost".** I made this mistake three times in one session,
   including reporting another thread's council run as my own — every council run
   is `owner_agent_type='generic'`, so recency cannot distinguish threads.
8. **`config.agent_type: "generic"` in a dispatch is a no-op agent** — it loads
   the generic definition instead of running your inline workflow. Send
   `config.workflow` alone.
9. **Column names that cost monitor restarts:** `llm_call_log.prompt_rendered`
   (not `rendered_prompt`); `agent_error_log.occurred_at` (not `created_at`);
   `site_work_items` has no `attempts`; `awaited_requests` has no `created_at`.

## Key identifiers

| thing | value |
|---|---|
| site | leopardessconsulting.co.uk `4851f6fc-71cf-4160-a270-e03d6d3e0732` |
| evidence base (current) | spec row `00b01b38-d166-4212-9f56-a48901655a3d`, rev 6, pinned |
| deployed image at handoff | `docker.io/aqls/agent-chassis:v1.0.1139` (contains V4) |
| council submission (voided) | `c9ca40d5-73c2-4fcf-a6e2-d8ee12e7bf60` |
| lost diagnosis run | `46253496-f8e0-471f-9ae0-29c9e630ada5` |
| V4 code | `06376bcbf` · V3 agent `0540698a4` · docs `fda25624c` |

## What NOT to do

- Do not rewrite `README_where_we_are.md` — it is the owner's log; append only.
  (A session overwrote another workstream's copy on 2026-07-19 after judging it a
  stray file.)
- Do not auto-fix a `claims_unverified` or `stale_evidence` item. They are
  human-terminal by construction — there is no handler agent, deliberately.
- Do not add banned patterns to any prompt.
- Do not apply the freshness seed before a dry run passes.
- Do not extend the layer to another site by editing shared machinery: everything
  is opt-in on `evidence_base` presence, and a site without one must stay untouched.
