# RUNNING NOTES — Experience Loop

*Append-only. Newest entry last. Companion to PLAN_experience_loop.md and
RUNBOOK_experience_loop.md. House rule: every entry states what was done, what
was verified (and how), and what it changes for the next actor.*

---

## 2026-07-17 — Workstream ACTIVE: defaults accepted, RUNBOOK drafted (session "vonc4")

**Owner input**: "defaults accepted" — resolving PLAN §7 / HANDOFF §3d:

1. **Gauntlet in MVP = minimal-real** — a playable timed round against the daily
   provocation, client-side scoring, no leaderboard; every fabricated number
   stripped. Demotion to coming-soon only if the feasibility critic proves
   minimal-real can't be honest and small.
2. **Provocation detail pages = static daily-emitted pages** — consistent with
   the daily JSON emitter design. NOTE (verified this session): the daily
   emitter itself is NOT BUILT — `/data/provocations.json` is a hand-committed
   static sample; the pipeline exists only as
   `docs/social001_vonc_tiktok_social/PLAN_spark_provocation_pipeline.md`.
   The council must scope this dependency in the MVP cut.
3. **Pilot fully autonomous on vonc** — artifact-verified checkpoints recorded
   here (CP1–CP5, defined in RUNBOOK §2), no approval gates.

**Done this session**: machinery survey (4 read-only exploration passes over the
acceptance ladder, council pattern, build/render side, claims verification) +
live-DB schema checks; RUNBOOK_experience_loop.md drafted from the findings;
PLAN header flipped PROPOSED → ACTIVE.

**Load-bearing facts established (each verified against code or live DB, refs in RUNBOOK §1):**

- `doc_plans`/`doc_notes` CHECK constraints allow only `('tool','pipeline')` —
  `subject_type='experience'` requires a migration on BOTH tables (RUNBOOK T1.1).
- browser-runner has NO navigation step action; each URL runs in a fresh browser.
  Tier-4 journeys are a genuine extension, not configuration (RUNBOOK T5.1).
- The arena's orphaning is a THREE-way key mismatch: criteria load by
  `content_components.function`, URL resolution by `pages.name = function`,
  travelling docs by `subject_key = function`. The page rename to `tool-arena`
  broke the middle link; re-keying doc_plans alone would NOT reconnect the sweep
  (RUNBOOK T4.2).
- `ReconcileSitePlanAction` emits `needs_page:<name>` → page-build-handler for
  EVERY plan page including tool-owned ones — the tool-role exclusion that
  `check_incomplete_page_group.go` already encodes (TP-004) is missing there.
  That asymmetry is guard rail 1's primary edit (RUNBOOK T2.1).
- No page-ownership marker column exists anywhere; TL-001 protection today is
  heuristic guards in `SavePageSectionsAction` + a manual park step.
- No code re-keys travelling docs on rename; no automated renamer exists either
  (the arena rename was manual). Root cause of the drift class:
  `create_tool_component_action.go:207` hardcodes `/tools/{function}.html`
  instead of calling `CanonicalisePage` (RUNBOOK T2.2).
- Claims verification V1a/V1b code is committed but NOT in any deployed image;
  vonc has NO evidence_base row, so the gauntlet's fabricated numbers are
  currently outside every enforcement lane (RUNBOOK T2.4).
- Migration sequence (`docs/agent_docs/sql_for_agents/`) runs through 161 with
  known collisions at 151 and 157; next free number at this snapshot = 162 —
  RE-CHECK at execution time.
- Concurrent-session state at drafting time: uncommitted WIP by another session
  in `platform/orchestration/actions/create_rerender_items_action.go` (a
  guard-rail-1 touchpoint) and an untracked live-probe test in
  `internal/adapters/browserrunner/` — re-check both before editing.

**Next actor starts at**: RUNBOOK §3 phase table — Phase 1 (foundations
migration) then Phase 2 (guard rails).

---

## 2026-07-17 (later, same session) — Phase 1 DONE, Phase 2 guard rails 1–4 DONE (code), image roll A next

Owner said "please go ahead"; execution began. Everything below is committed
per task and verified as stated; RUNBOOK §8a mirrors this as a state table.

**Phase 1 — DONE.** Migration **163** (not 162 — that number was taken
mid-flight by another session's toolgen-rerender tail, itself good news: it
closes TP-002). Both subject_type CHECKs now allow `experience`; applied +
ledgered; probe insert verified. Commit `378054bad`.

**Guard rail 1 — DONE (DB live, Go awaits image).** Migration **164**:
`pages.rebuild_policy generic|owned`, 38 pages seeded owned (36 `page_type=
'tool'` fleet-wide + vonc `provocations-index`/`provocation`). Go: reconcile
now emits `owned_page_review` (needs_human_review, NO handler) instead of
`needs_page`→page-build-handler for tool/game-role or owned pages — retires
the manual park step; `save_page_sections` hard-refuses owned pages before
its heuristic guards. Rerender/assembly deliberately not gated. Commit
`fb89f1071`.

**Guard rail 2 — DONE (code).** `RekeyTravellingDocs` datahelper (refuses
two-current collision), `rename_tool_identity` action (atomic function +
slot_name + doc re-key + rename note; reports pages.name coupling and stale
js_snippets/nav refs), and `create_tool_component` now derives page identity
via `CanonicalisePage` (kills TL-003 flat-URL drift at birth;
sanitiseFunction's tool- prefix keeps pages.name == function). Commit
`aabd38161`. The arena re-attach (T4.2) is this action's first deliberate use.

**Guard rail 3 — DONE (code).** `IsNoopHref`/`DeadControlAnchors` (+unit
tests, green): bare `#`, `#!`, javascript:void — the class ClassifyLinkScope
correctly files under anchor scope, which is why the gauntlet's dead CTAs
were invisible to phantom_internal_links. Wired as (a) Tier-2 built-in
`shell-dead-controls` and (b) `dead_controls` discovery check
(page_components only — chrome nav toggles are legit #+JS; runtime-fill
shells exempt; emits `dead_control`, needs_human_review, no handler).
Enable SQL **165 written but NOT applied** — image-first ordering. Commit
`f2824a713`.

**Guard rail 4 — DONE (live).** Migration **166**: vonc `evidence_base`
seeded — facts deliberately EMPTY (nothing quantitative is assertable until
EXPERIENCE_PLAN data contracts exist), 9 banned patterns from a live-page
harvest, allowed_entities for the fictional product nouns. **claimscan
baseline: 14 findings across 49 components** — the full known gauntlet set
PLUS three previously unknown fabrications: `14,203 Happy Customers`
(gauntlet-cta component on about AND index), `10K+ Players Scored` (index),
and mangled stat labels on about (`1 Daily Clients Served`, `4 Hours Awards
Won`). All queued for the T4 strip. Also found: `unverified_claims` was
already pre-enabled on quality-discovery-agent by its own workstream —
V1b activates the moment the claims image ships, no further seed needed.
Commit `c437682a6`.

**Concurrency observed this session** (all handled): migration 162 claimed
between drafting and execution; the build default inverted fleet-wide
(default target now builds committed HEAD — RUNBOOK §0 updated); cluster
rolled 1128→1130 by other sessions; `create_rerender_items_action.go` WIP
belongs to the toolgen-tail session (file untouched by us, as planned).

**Next**: T2.5 — image roll A + CP1.

---

## 2026-07-17 (later) — ✅ CP1 REACHED: guard rails live in prod and proven on vonc

**Image roll A came free.** Another session built+deployed **v1.0.1134** from a
HEAD that already included all four guard-rail commits — so no separate roll
was needed (my local 1132 build is moot; I did not push it). Verified in-pod,
NOT by tag: `strings /app/agent-chassis` in pod `agent-chassis-6d85fff446-54jzc`
(container name is `agent-chassis`, NOT `agent` — that's only the intel pair)
shows rebuild_policy×4, owned_page_review×4, rename_tool_identity×4,
RekeyTravellingDocs×2, shell-dead-controls×1, dead_controls×6, CanonicalisePage×2.
Deployment AND agent_definitions.image_tag both v1.0.1134.

**Migration 165 applied** (image-first satisfied — dead_controls symbol confirmed
in-pod): `dead_controls` now in completeness-discovery-agent's checks array.

**CP1 proofs (all live on vonc, artifact-verified):**

1. *Binary*: the 7 symbols above are in the running 1134 binary.
2. *dead_controls check fired* — completeness discovery (corr
   4cedb4fb, completeness-discovery-agent row COMPLETED) emitted two live
   `dead_control` items: index `brief-explanation` "Get Started" → `#` and
   "Learn More" → `#`, both needs_human_review, no handler. Genuine new finds
   (the index had dead CTAs too, not just the gauntlet). The gauntlet's own
   dead CTAs weren't caught here only because tool-gauntlet is
   build_status=needs_rebuild, not deployed — it's covered by the claims lane
   and rebuilt in T4.
3. *owned_page_review routing* — a SCOPED reconcile (corr 4c0c4acf; ran ONLY
   reconcile_site_plan via a one-step envelope, NOT build-site-planner, to
   avoid the re-plan clobber) emitted `owned_page_review` (needs_human_review,
   NO handler) for `provocation`, `tool-gauntlet`, `tool-archetype-taster-quiz`
   — exactly the tool/owned pages that previously went to needs_page →
   page-build-handler. `tool-arena` (deployed at current plan version) correctly
   skipped. **Zero needs_page emitted for any owned page.** The manual park of
   needs_page:provocation is now mechanical; the stale 07-12 park item
   (01674b35) was cancelled as superseded.
4. *save_page_sections refusal*: code-verified in the 1134 binary (the
   rebuild_policy='owned' guard returns before the DELETE+reinsert); not fired
   against a live owned page — the reconcile proof already demonstrates the
   marker is read and enforced, and firing a destructive save purely to see it
   refuse is not worth the risk.

**Guard-rail bonus already banked**: between the claimscan baseline (14 findings,
3 previously unknown) and the dead_controls sweep (2 more on index), the rails
surfaced 5 defects nobody had catalogued — before any experience work began.

**Phase 2 CLOSED.** Next: Phase 3 — experience-planner agent + challenge council,
run to convergence on vonc-spark-game (CP2).
