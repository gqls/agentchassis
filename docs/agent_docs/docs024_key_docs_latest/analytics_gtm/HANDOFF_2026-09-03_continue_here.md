# HANDOFF — analytics / GTM / GA4 / consent · continue here (supersedes HANDOFF_2026-08-25)

**Written 2026-09-03 ~17:15 BST, session "google".** This lane owns **everything Google** (owner
ruling 2026-08-25, verbatim in the 08-25 handoff). Fleet background:
`docs/agent_docs/docs024_key_docs_latest/039_REFERENCE_traffic_and_tracking.md` (read its two
dated addenda first — they mark the GA4 history epoch and the consent step-change).

> ## ▶ ONE-LINE STATE `[ALL MEASURED 2026-09-03 16:45–17:10 BST]`
> **GA4 is LIVE** (`GTM-PQ3WCTBD` v3 → `G-Y26N29T4KH`, history began 2026-09-02 ~20:11Z).
> **Consent (Option A) is SHIPPED and proven in production** — 25 of 38 tagged heads carry the
> banner block, the rest converge via their own rebuilds; live behavioural proof on noted.co.uk
> 5/5. **The new-site seeder (STY-061) is on HEAD and building but INERT until the next fleet
> roll**, and its council round is IN FLIGHT: round 1 REVISE (answered with a complete census),
> round 2 **FAILED SILENTLY** at `review_architecture` (16:07:33Z, no `__step_error` — the
> spawn→call handshake class), **round 3 dispatched 17:10 BST** (`RUN_ORCH_ID
> dc6d0538-5b39-44c0-a5af-bffa175da4b2`). **Migration 733 is validated and DELIBERATELY UNAPPLIED
> until the verdict.** Customer container `GTM-TH5XGNQ4`: empty, and empty = correct.

## 0. The one in-flight thing: council correlation `45ae3ad3-b10d-4f27-a412-a9a79f0f0cab`

```sql
SELECT current_step, status, updated_at FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '45ae3ad3-b10d-4f27-a412-a9a79f0f0cab'
 ORDER BY updated_at DESC LIMIT 1;
SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%45ae3ad3%' ORDER BY created_at DESC LIMIT 1;
```
- **APPROVED** → apply migration 733 via the runner, SCOPED (other threads' files are pending —
  never bare `--apply`): copy `733_network_default_analytics_container.sql` alone into a temp dir,
  `MIGRATIONS_DIR=<tmp> scripts/migration/run-migrations.sh --apply` **inline assignment**
  (LANDMINES: an assignment on its own line scopes nothing). Then the arc is done pending the next
  fleet roll; note the future `Council-Reviewed` credit is automatic (098, the trailer is already
  `Council-Submitted` on `fe7359158`/`1501718a5`).
- **REVISE** → objections come with reviewers' checks answered; revise, resubmit
  `RESUBMIT_CORR=45ae3ad3…`. Round 1's pattern: evidential gaps, and it was right both times.
- **FAILED again** → do NOT cancel the row (memory: handshake failures are ~half of spawns);
  resubmit on the same correlation — a failed run produced no verdict and is judged fresh.
- Submission JSON (for edits): the session scratchpad copy is gone with the session — rebuild from
  `bugs_open/397` §6.2 + the STY-061 register entry, which carry every fact in it.

## 1. Verify before doing anything — one command

```bash
docs/agent_docs/docs024_key_docs_latest/analytics_gtm/scripts/check_gtm_state.sh --all
```
Expected now: estate container **PUBLISHED → G-Y26N29T4KH** · customer container (run with
`CONTAINER=GTM-TH5XGNQ4`) **0 tags — which is CORRECT for that one; a tag appearing there is the
webdesign.uk DECISION §5 re-ruling trigger** · `--sites` shows `gtm=` and `consent=` per domain ·
`--db` census A/C (B and D must stay 0; a D reappearing = a site born before the seeder rolls —
c2 with `-v UNTAGGED=1` still works and now refuses `mode:'none'` sites).

## 2. What is DONE and proven (dates are the evidence trail in NOTES §27–§31)

- **GA4 publication** (owner's click, 09-02): container v3, one Google Tag. GA4 history starts
  there; Cloudflare is the only source with a past (039).
- **Consent Mode v2 + banner (STY-060)**: inside `{{if .gtm_container_id}}`, before the loader, in
  all three head templates; defaults all-denied; Accept/No-thanks equal; withdrawal wipes `_ga*`;
  choice in `localStorage.cc_v1`. **Proven 26/26 locally then 5/5 on live noted.co.uk.** Fail
  direction = denied. GA4 UNDERCOUNTS from each site's convergence — the consent gate, not lost
  traffic (039 addendum).
- **Durable tagging**: 397's two c2 waves; every artefact-only site restored; the loss mechanism
  and the 7/7-vs-10/10 natural experiment are in `bugs_open/397` §10.
- **Seeder (STY-061)**: `seedAnalyticsDefault` in `EnsureSiteRecordAction` — network-value opt-in,
  no-current-row guard, `mode default|custom|none`, system/test pseudo-sites excluded, non-fatal
  by contract; sqlmock tests; HEAD builds (`verify-head-builds.sh ./platform/orchestration/...`).

## 3. Owner actions open

1. **The 11 `chrome_divergence_overwritten` review rows** — diagnosed (397 §10-addendum), the
   status write was permission-blocked in this session; approve and any session closes them.
2. **cv1.co.uk GA4 membership** (lampenkap RULED 08-26: keep; cv1 still open).
3. **Search Console** — a GCP service account; grant Tag Manager API on the same account and
   container automation unblocks too. Nothing exists on our side (re-checked 09-03).
4. Nothing else: publish is done; the second container is done.

## 4. Owed builds, in order

1. **Cookie/privacy policy pages per site** (the PECR "information" half; banner text is the
   interim notice) — through the framework (`llm_fields` first — the framework writes the
   content), ~30 sites + nav; needs its own PLAN section before starting.
2. **§6.3 detector** — a `config-key-audit`-fleet daily check for head-artefact-vs-spec-key
   disagreement (manual form: `check_gtm_state.sh --db`). Remember the RFC_022 lesson: keep the
   cron literal in step and run the parity test.
3. **webdesign.uk's open question** (their CONTRIB reply pending): should hosted customer copies
   suppress the banner while their container is empty? Mode-aware suppression is small; decide
   before their first customer build.

## 5. Traps paid for since the 08-25 handoff — do not re-pay

- **Nothing on a page tells you whether GA4 records. Only `gtm.js` does** — and parse it by
  counting `"function":"__googtag"` FILE-WIDE, never a blob keyed to a neighbouring literal: the
  2026-09 format dropped `"predicates"` and the old parser called the owner's first-ever publish
  FAILED (`WRONG_CALLS.md` 2026-09-02). A pre-09-02 copy of the check script has the bug.
- **A value backfilled into `rendered_html` is a cache** (LANDMINES 2026-08-25) — write the spec
  key; expect the stale_chrome rebuild; MERGE into existing rows.
- **Writing `site_config` = firing a rebuild** (fingerprinted). Size in pages first.
- **snap Chromium** refuses a profile under a hidden home dir — use `~/snap/chromium/common/`;
  `chromium.chromedriver` speaks plain HTTP (no websocket dep) and is how both consent proofs ran.
- **097 submissions**: `plan` is an OBJECT (`summary/edits/grounded_in/risks`); a sketch whose
  every line starts `#`/`--` is refused as comment-only (a markdown heading trips it).
- **A timestamp you didn't read from the artefact is a guess** — the c2 apply time was corrected
  by a peer reading the row stamp (`WRONG_CALLS.md` 2026-09-02, second entry).
- All 08-25 traps stand (`HANDOFF_2026-08-25_continue_here.md` §4, and its banner stack is the
  history of this arc — kept, superseded).

## 6. Cross-lane

- `webdesign_uk_build_service` — owns intake/ZIP/copy for customer builds; their DECISION doc
  carries the mode table and `GTM-TH5XGNQ4`; my CONTRIB 2026-09-03 in their dir asks the banner
  question and states the seam is built.
- `site_delivery_and_editor` — ZIP path (DGH-011); consumes `mode` when built.
- `apis_uk_bees_homepage` — split ruling 08-25 (everything Google here); their site is durable and
  converged.
- `dartsonline` — two `page_rerender …_section_data_resolved` items failed 09-02 (`result={}`);
  their machinery, their pages serve pre-consent bytes until rerendered.
- The 11 review rows' sites — all answered in 397 §10-addendum; do not "review" them as hand-edits.
