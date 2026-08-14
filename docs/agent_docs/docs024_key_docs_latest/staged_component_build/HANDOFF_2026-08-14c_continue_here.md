# HANDOFF — 2026-08-14 (c), fresh chat starts here: the bucket-A pilot LANDED, and re-measuring it found the census counts 11 rows that must NEVER be redeployed

**Supersedes `HANDOFF_2026-08-14b_continue_here.md`.** That file was written ~17:00–17:30Z and
said "nothing was fired"; a later session in the same day **executed the bucket-A pilot**
(`8fed258fd`, 22:35Z) and a services-restore session appended site numbers (`ddd33315b`). Both
postdate it. Nothing in 14b is wrong about *mechanism* — its landmines and its wire-check-first
rule all held, and one of them earned its keep twice during the pilot. What has moved is the
**numbers and the target list**, and this file re-derives both from scratch.

Everything in §1 was measured by me at **2026-08-14 ~22:40–23:00Z**. Re-measure before acting:
bucket B self-drains and a discovery sweep can refile at any time.

## 1. State — verified, with the checks

- **The fix is LIVE and in the running binary, not merely committed.** Fleet
  `agent-chassis:v1.0.1300`. Probed `/proc/1/exe` on a running chassis pod with controls in both
  directions — this is the method the LANDMINES file prescribes (never `strings`, never a
  discovery grep):

  | needle | result |
  |---|---|
  | `discarding an asset_key that is an unresolved path expression` | **PRESENT** |
  | `taking purpose from the asset row rather than the spec default` | **PRESENT** |
  | `ZZZ_MUST_BE_ABSENT_CONTROL` | absent (negative control) |

- **Both original symptom sites are still fixed:** `gaswholesalers.com/assets/images/logo.png`
  **200**, `mortgagecalculator.co.uk/assets/images/hero.jpg` **200**.
- **Census: 98 rows / 12 sites** (`filename LIKE '%asset-key%' OR url LIKE '%asset-key%'`), down
  from 140/14 at the pilot session's start. The drop of 42 = the pilot's 11 + bucket B's
  self-drain, exactly as that session predicted.
- **The LLM cap has RECOVERED** — 14b left this explicitly unsettled ("a partial recovery in
  progress, not a settled fact either way"). Measured on `llm_call_log.success`: last five hours
  run **24, 124, 53, 48, 37 ok against 0,0,1,1,1 failed**. So bucket E's regeneration subset is
  no longer cap-blocked. (Still re-check: a monthly cap can bite again before 2026-09-01.)
- **`RFC_028` and `RFC_029` both still await an owner ruling.** No ruling commit exists for
  either. Nothing to do here.

### ⚠ TWO NEW FINDINGS THAT CHANGE THE TARGET LIST — read before touching the drain

**(1) The census counts 11 rows that must never be redeployed.** The marker query filters on
`filename`/`url` and **not on `assets.status`**:

| status | rows in census |
|---|---|
| active | 87 |
| superseded | 10 |
| retired | 1 |

A `superseded` or `retired` asset has been REPLACED by a newer row. Redeploying its bytes would
push a stale image over a current one — the same class of harm as the live-200 overwrite the
pilot's wire-check caught, arriving by a different door. **The real drain target is 87, not 98.**
Nothing in 248 or in 14b filters on status; the pilot did not hit this because bucket A happened
to contain only one such row.

**(2) Bucket A is NOT remaining work — its members are the pilot's DELIBERATE skips.** Re-bucketed,
the two rows that still show `unresolved` (i.e. "needs a promote") are:

| site | purpose | asset_id | status | `logo.png` at the wire |
|---|---|---|---|---|
| leopardessconsulting.co.uk | logo | `71652e42-…920ad3` | **retired** | **200** |
| finetuning.uk | logo | `9c9de5a0-…a61eea9` | active | **200** |

These are precisely the two the pilot **skipped on purpose** as live-referenced 200s. **A fresh
session reading "bucket A: 2 rows, promote them — the pilot proved that action" would cause the
exact regression the pilot avoided**, and on leopardess it would serve a *retired* asset's bytes.
The census cannot distinguish "not yet done" from "deliberately left alone", and nothing in the
row records the decision. Treat bucket A as **0 rows of work** until someone gives those two a
bookkeeping-only correction path (see §2.3).

### The corrected work-list — ACTIVE rows only, which is the only list worth acting on

| bucket | active rows | what it needs |
|---|---|---|
| **D** — only terminal items exist | **57** | a freshly cloned item at `triaged` (mechanism proven at 2-row scale in 248 §2b, and at 11-row batch scale by the pilot) |
| **E** — no `undeployed_asset` item ever filed | **27** | **per-row wire check FIRST**, before any item is filed at all |
| **B** — an open item exists | **2** | nothing. Self-draining; `build-dispatch-loop` is alive and periodic |
| **A** — `unresolved` | **1** | nothing — it is finetuning's deliberate skip (above) |

Per site (all 98, active + non-active, so it reconciles against the raw census):

| domain | E | D | A | B | total |
|---|---|---|---|---|---|
| dartsonline.com | 0 | 28 | 0 | 0 | 28 |
| fundamentallyai.com | 12 | 4 | 0 | 0 | 16 |
| gamesdesign.co.uk | 14 | 0 | 0 | 0 | 14 |
| leopardessconsulting.co.uk | 0 | 13 | 1 | 0 | 14 |
| vetcomparison.uk | 3 | 4 | 0 | 0 | 7 |
| finetuning.uk | 0 | 5 | 1 | 0 | 6 |
| idea.uk | 0 | 2 | 0 | 2 | 4 |
| lendzy.co.uk | 0 | 4 | 0 | 0 | 4 |
| robot-hands.com | 0 | 2 | 0 | 0 | 2 |
| webdesign.co.uk | 1 | 0 | 0 | 0 | 1 |
| vonc.com | 0 | 1 | 0 | 0 | 1 |
| ai-agent-orchestration.com | 0 | 1 | 0 | 0 | 1 |

**This bucketing sums to 98 exactly** — 14b's summed to 133 of 140 and flagged the gap as
unresolved. The gap was rows matching more than one work item; the query here aggregates per
asset with `bool_or`, so each row lands in exactly one bucket. Query is in NOTES
`## 2026-08-14 (c)` — reuse it rather than re-deriving.

**Note the concentration:** `dartsonline.com` alone is 28 of the 57 bucket-D rows, and three
sites (dartsonline, leopardess, fundamentallyai) hold 45 of them. A per-site pilot is therefore
cheap and informative; a fleet-wide batch is neither.

## 2. What's actually left

1. **Bucket D (57 active rows) — the bulk.** Inherit the pilot's gate **verbatim**: wire-check
   BEFORE filing/promoting, per row, and skip anything already serving 200 at the reader-derived
   path. Add the new status filter: **`AND a.status='active'`**. Pilot one site first —
   dartsonline (28) is the biggest single win but also the biggest blast radius; `robot-hands`
   (2) or `vonc` (1) is the cheaper canary. **This is a real production action across live sites
   and wants a check-in, not a silent bulk fire.**
2. **Bucket E (27 active rows) — wire check first, then decide whether to file at all.**
   14b's reasoning stands and is worth re-reading: `check_undeployed_assets.go` refuses to file
   against this shape for favicon/og_card because doing so would be a FALSE claim, and the
   services-restore contribution found six leopardess `icon_service_*` files serving 200 with
   distinct sizes while their rows carry the placeholder — metadata-only, no redeploy warranted.
3. **The "stale row only" class now needs a decision, not more measurement.** Confirmed members:
   the 2 bucket-A skips, the 10 icons the pilot waved through, the 6 leopardess service icons,
   plus most of bucket E. These rows are *bookkeeping wrong and artefact right*. Redeploying to
   tidy a row is backwards; hand-editing `assets` is off the table (framework rule). **The open
   question — does this class get a bookkeeping-only correction path? — is a small design
   question that should be answered before bucket D starts**, because some of D's 57 will turn
   out to be this class too and the pilot measured that rate as nonzero (2 of 13).
4. **RFC_028 + RFC_029 await the owner.** Not a normal session's task.
5. **The "owned vs generic rebuild policy" is still `[UNMEASURED]`.** 14b checked `sites`
   (`settings`, `locked_at`/`locked_by`, `status`) and found no such column; none of the sites
   are locked. Still unresolved — ask the owner before batching across sites that might need
   separate notice.

## 3. Older items from this lane that have NOT moved — re-verified today, still open

- **The four tracker feeds still 404.** `adoption-tracker.json` and `protocol-tracker.json`
  **404**; `model-directory.json` **200**. Nobody has acted since the evidence was written up on
  08-12. The fix is a config `UPDATE` + the platform's own force-trigger idiom, no image roll,
  and the verification (identical entity counts across two kinds = the 07-26 defect reproduced)
  is written down:
  `model_directory_pipeline/FINDING_2026-08-10_the_tracker_publisher_was_reverted_and_never_re_extended.md`.
  That lane has been dormant since 2026-07-26 — **unowned, free to pick up.**
- **`tool-gas-unit-converter` is still broken and still unparked.** All three items sit at
  `needs_human_review` (two were touched 08-14 by a sweep, not repaired). The blocker is
  unchanged: the page carries `sections=[]` and no plan, so `page-build-handler` correctly
  no-ops, and `required_fields_missing` has no repair handler anywhere in the fleet. Needs a
  build-pipeline re-dispatch decision from the owner, not a handler retry.
- **⚠ A stray `/assets/images/logo.jpg` remains on gaswholesalers.com** (200), left by this
  lane's 08-10 diagnostic dispatch. `logo.png` is now correct and live alongside it; the stray is
  referenced by nothing. It should be removed by whoever next touches that repo — deleting from a
  site repo is a write path nobody has yet exercised deliberately here.

## 4. Landmines worth carrying forward

14b's three still stand (a flat census is not a target list; a periodic loop idling on `triaged`
is not a stall; a check's *exclusion* logic is as load-bearing as its filing logic). Two to add,
both from re-measuring rather than from a symptom:

- **A census keyed on a corruption MARKER cannot tell you whether acting is safe.** It selects
  rows whose bookkeeping is wrong, which is a different set from rows whose *artefact* is wrong,
  and it silently includes `superseded`/`retired` rows whose bytes must never be republished.
  Filter on `status`, and check the wire, before the row count becomes a work-list.
- **Your predecessor's DELIBERATE non-action is invisible to your query.** The two rows that look
  most obviously actionable in bucket A are the two a careful session decided not to touch, and
  the reason lives only in a bug-file contribution. Before acting on a small residual bucket,
  grep the bug file for the row or site — a residual of 1–2 after a pilot is a *skip signal*
  more often than an omission.

## 5. Session-start checklist

1. `git log --oneline -10`; re-read this file FROM DISK; read `bugs_open/248`'s last two
   contributions (bucket-A pilot, and the leopardess services-restore numbers) — this file
   summarises but does not replace them.
2. Confirm the fix is still in the RUNNING binary using the two-way control table in §1. The
   fleet rolls often; `v1.0.1300` will be stale. An absent literal means re-check before
   concluding a regression — the chassis provenance log line scrolls out within hours.
3. Re-run the census **and** the active-only bucket query (NOTES `## 2026-08-14 (c)`). Both
   figures here are ~23:00Z snapshots.
4. `scripts/who-owns.py 248` — **two unrelated bugs share the number 248**; resolve by filename,
   never by number. Also check for a concurrent session mid-batch on the drain.
5. Re-check `llm_call_log.success` before anything that regenerates an image; the cap recovery in
   §1 is hours old, and the monthly cap does not expire until 2026-09-01.
