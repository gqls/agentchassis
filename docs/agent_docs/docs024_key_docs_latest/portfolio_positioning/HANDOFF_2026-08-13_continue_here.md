# HANDOFF — fleet build-out: Phase A DONE+LIVE, Phase B (finance directories) planned and ready to execute — 2026-08-13, continue here

Cold-start for a fresh chat picking up this exact thread. Supersedes
`HANDOFF_2026-08-11_continue_here.md` for build-out purposes (that file's §6
mortgagecalculator voice-review queue item is still owed and unchanged — see §6 below).
Read `MEMORY_workstreams.md`'s portfolio-positioning line first, then this file, then
`PLAN_2026-08-12_fleet_buildout.md` (the phase map) and the **Phase B execution plan**
whose content is reproduced in §4 (its working copy lived in the previous session's
plan file; §4 is the durable copy).

## 0. Owner rulings in force (2026-08-12/13 sessions, all still binding)

1. **Nothing new through the build pipeline until the structural-validity gate AND
   RFC_025 are live** — SATISFIED as of 2026-08-13 (see §1). The gate now moves to
   enablement + Phase B.
2. Newsfeeds and directory listings on new sites **from day one**.
3. Generated tools stay on **manual per-site review** (no gate-loosening).
4. Graphs: **`evidence-chart` bar charts only** where a site has verified cited data.
5. Directories: **non-price facts only** (FRN, regulator status, product types,
   underwriter, established year — never APR/rates/premiums). Compliance surface.
6. Phase B kinds: **starter set of 3** — `mortgage-lender`, `savings-provider`,
   `health-insurer` — remaining ~6 added mechanically once proven.

## 1. Phase A: DONE, APPROVED, AND LIVE (verified at the artefact 2026-08-13)

Chassis `v1.0.1295` (rolled ~13:53Z by another lane) carries everything. Verified by
binary probe on BOTH replicas (stamp `69612d692` in `/proc/1/exe`, absent-sha control
clean) + `git merge-base --is-ancestor` for each commit. The log-line recipe was NOT
available — the startup provenance line had rotated out (documented landmine; another
session filed the same finding as "TIME-LIMITED on agent-chassis", commit `6ceeaba1b`).

| piece | commits | council corr | verdict trail |
|---|---|---|---|
| Structural-validity gate (5 checks: `dead_internal_link_live`, `canonical_mismatch`, `structured_data_invalid`, `head_essentials_missing`, `sitemap_entry_dead_live`) | `c66a83e9e`, `9e410ce85`, `a86dd0349` | `51cb66fb-e4bc-46ec-8bbd-a4a561da14a0` | REVISE→REVISE→**APPROVED** (08-13 10:26) |
| RFC_025 (bug 161 fix: `artifact_check` + attestation staleness nudge) | `3129cceea`, `9652f4d52` | `9fd94852-ff79-496b-96b5-78a8d3619162` | REVISE→**APPROVED** (08-12 20:42) |

**No commit carries a Council-* trailer** (all predate their submissions) — resolve
verdicts by correlation in `diagnosis_artifacts`, never by commit message. The durable
records are RFC_025 §10 and NOTES here.

**Not yet done from Phase A** (now unblocked, folded into Phase B's migration wave):
- The 5 structural checks are registered but **not enabled** on any discovery agent —
  enablement migration per the `194`/`215` pattern (plan §B3f).
- No auto-repair routing (deliberate; `canonical_mismatch`→`page-rerender` is future
  work gated on bug 251's fix reaching all four render paths — it's live on ONE).
- RFC_025 short of `IMPLEMENTED` until one real fact uses `artifact_check` (§10 of the
  RFC). The attestation nudge (stage 1) now runs automatically on the daily sweep —
  expect `stale_attestation` items to start appearing for attested facts >180 days old;
  that alone satisfies the "either mechanism" condition once observed.

## 2. The review-round war stories a fresh session must not re-learn

- **Both pieces took multiple REVISE rounds and every gating objection was real** —
  including one where the new `artifact_check` regex would have reproduced the exact
  landmine its own motivating bug documented (bare `10000` matching `100000`), and one
  where a comment's premise about `check_missing_structure.go` collided with the
  vestigial-columns landmine. Verify claims against the live system before writing them
  into headers; the council's code index CAN be stale (it falsely denied
  `preferredPageURL` existed — refuted by direct read, rerender_single_page_action.go:1050).
- **Bug 270 filed** (`bugs_open/270_...missing_structure_check_fires_on_vestigial_columns...`):
  `missing_structure` is LIVE and fires on a predicate that cannot be false — ~25
  completed full-site rerenders since April for nothing. Unowned, not on Phase B's
  critical path, but a standing confounder for rerender-churn investigations.
- `git stash -u` on this shared tree sweeps other sessions' WIP (LANDMINES.md entry
  added 2026-08-12) — use `git stash push -u -- <your paths>` only.

## 3. Phase B design ground truth (verified 2026-08-13 against HEAD; deltas from the
   original design that MUST be respected)

- Adding a directory kind = **six** data changes, not four: `directoryCheckProfiles`
  (check_directory.go:72), `directoryPublishProfiles` (render_directory_action.go:98),
  queryresolve case arms (:116-135, follow `adoption_tracker` NOT the bespoke `model`
  functions), content_components rows, completeness-agent `checks` array enablement
  (194/215 pattern), build-site-planner prompt/component vocabulary.
- **No schema migration needed** — `directory_entities.kind` is unconstrained text.
  Column is `attributes` (NOT `entity_attributes` — that's only the candidate-JSON key).
  Facts (FCA status etc.) are CLAIMS, never attributes (schema's own comment).
- `evaluate_news_feed` wiring precedent is **migration 054** (improvement-loop
  `enrich_news_feed`), NOT 049 — and **291 re-pointed its edges** to `load_audit_state`;
  read the LIVE improvement-loop config before splicing. 049 carries only planner prompt
  rule 11.
- The publish trigger (`SEED_directory_publish_trigger.sql:94`, model-directory-trigger's
  agent_definitions config) is kind-blind in THREE places + `LIMIT 5`. The
  2026-08-10 FINDING's fix was never applied. Phase B inherits this.
- `directory_citation_unverified` + `stale_directory_claim` use CONSTANT item_keys at
  the system pseudo-site (`directory_claims.go:346`, `:611`) — kind-scoping needs 4
  edits (Kind field on `rejectedDirectoryClaim`, populate at 6 literals, group-by-kind
  emit, same for the stale sibling). Read `register/batch-processing.md:49` first.
- The model_directory lane is **dormant 18 days**; its own FINDING says the work is
  free to pick up. Migration numbering: was 400 next, but 5 files are untracked —
  **re-`ls` sql_for_agents/ immediately before claiming numbers.**
- Researcher precedent: `SEED_adoption_researcher_agent.sql` (one agent, multiple kinds
  from one prompt; `exclude_domains` only — its stated anti-marketing-bias reasoning
  transfers directly to finance). Config gotchas both confirmed live: `timeout_seconds`
  INSIDE step config (062), `ai_service` at STEP level (009).

## 4. The Phase B execution plan (approved by owner 2026-08-13)

Naming scheme, stages B0-B4, sequencing, out-of-scope, verification — the full approved
plan text is long; its stages in one line each:

- **B0 (DONE)**: pod-verify v1.0.1295 (done, §1), docs updated, this handoff.
- **B1+B2 (DONE, committed `6f26570e4`, council corr `69a619e6-5152-45d8-ae01-5d30a0c7776f`
  — trailer on the commit this time, verdict pending as of writing; READ IT before
  building on this)**: all three kinds in all three profile tables + 6 queryresolve
  arms; kind-scoped item keys (group-by-kind emit, one txn, empty-kind→legacy key,
  fresh log literals, both emitters); `evaluate_directory_features` action + registry
  entry (its partial matcher is deliberately longest-key-deterministic — the map mixes
  recommending and refusing entries, unlike the news map); DIR-001 register entry +
  index row + ratchet line dropped. Tests: mixed-kind emitter scripts, matcher table,
  map/profile lockstep guard. NOTE: a background agent began this work and was killed
  by an API session limit mid-B1b; its partial work was extracted from its worktree,
  completed, and reviewed line-by-line in the main session. The two bug-231 image test
  failures in the full actions suite are ANOTHER session's in-flight WIP (verified
  passing at clean HEAD in a throwaway worktree).
- **B3 seeds/config, ordered AFTER the B1 roll is pod-verified** (image first, then
  seeds): B3a components seed (6 rows, createElement/textContent XSS discipline from
  SEED_directory_components.sql:16-25, `component_level:'section'`); B3b researcher
  agent + 3 weekly scheduled tasks sharing one concurrency_group; B3c publish-trigger
  three-predicate fix + publisher chain extension; B3d wire evaluate_directory_features
  into improvement-loop (after enrich_news_feed, LIVE edges) AND
  domain-research-classifier (after write_classification_spec — this is what makes
  greenfield builds directory-aware at plan time); B3e planner prompt rule + component
  vocabulary (206's replace()-idiom, snapshot first); B3f enable checks LAST (after
  claims exist), same wave can enable the 5 structural checks.
- **B4 first researcher runs**: force-trigger (`last_triggered_at=NULL`), work the
  `directory_citation_unverified:<kind>` HITL queues, bar = reviewed non-embarrassing
  `found` claims per kind, then spot-check resolver/JSON/HTML agreement.

Naming: kind `mortgage-lender` → SpecKey `mortgage_lender_directory`, components
`mortgage-lender-directory`/`-listing`, queries `mortgage_lender_directory`/`_full`,
page `mortgage-lenders`; same pattern for `savings-provider`, `health-insurer`.

Out of scope: the other ~6 kinds; price facts; per-site filtered directories (would
change the shared guarantee → architecture review); bug 270.

## 5. After Phase B (unchanged from PLAN_2026-08-12)

Phase C: pilot ONE real proposition end-to-end (not lendzy, not a HOLD), manual cost
baseline from llm_call_log/assets, owner sign-off. Phase D: owner decisions (loanzy.uk
conflict with webdesign lane, B8/B9/I10 holds, build order across ~140 domains).
Phase E: fleet dispatch in waves.

## 6. Also still owed (from the 2026-08-11 handoff, unchanged)

The mortgagecalculator.co.uk copy voice review (owner explicitly queued it after the
enforcement-gap work): bring the live homepage + finished title pass to the owner
against `mortgagecalculator_couk_adoption/REFERENCE_2026-08-11_learned_by_correction_house_voice.html`.

## 7. Files of record

This dir: `PLAN_2026-08-12_fleet_buildout.md` (phase map),
`NOTES_portfolio_positioning.md` (full evidence trail, newest at bottom),
`README_where_we_are.md` (owner's plain-prose log, append-only).
Bugs: `bugs_open/270_...` (this session's filing).
RFC: `architecture_review/RFC_025_artifact_sourced_facts_are_trusted_once_registered.md`
(§9 ratified answers, §10 implementation + LIVE status).
Council correlations: §1's table — always resolve by correlation.
Commits this session (2026-08-12→13): `bf9635e5f`, `43f9467b9`, `8bdc3080c`, `071ab8358`,
`dc9f2d87f`, `c66a83e9e`, `fd398a7e1`, `d20c74d67`, `3129cceea`, `72dadd5c6`, `9652f4d52`,
`f23c55f27`, `9e410ce85`, `cea9e0a5e`, `69ce310d8`, `59fb76678`, `a86dd0349`, `fe211e277`,
`69612d692`, plus this handoff.
