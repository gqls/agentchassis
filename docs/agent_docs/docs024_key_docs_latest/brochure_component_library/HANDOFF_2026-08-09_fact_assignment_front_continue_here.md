# HANDOFF 2026-08-09 — fact-assignment front (bug 151 / RFC_016): cold-start for a fresh chat

**Supersedes `HANDOFF_2026-08-08_continue_here.md`.** Written ~14:30 BST after
re-verifying every premise against the shared tree (**341 commits** landed since
this front's last commit) and the live cluster. Everything below was checked
today unless marked otherwise.

**This is ONE OF TWO fronts in this directory. Do not confuse them.**
- **This file = the fact-assignment front** (bug 151 candidate 1, RFC_016,
  planner/writer prompts, seeds 327/328/329/330/333).
- **`HANDOFF_2026-08-09_sweep_front_continue_here.md` = the fundamentallyai
  sweep front** — a different live thread, same site, same directory. It owns
  the improvement sweep, linkability (`1c2e25c8f`, council-submitted
  `9da24d85`), and `bugs_open/210`. **Read it before touching the site**: we
  share one acceptance site and our actions land on each other.

Site id, needed everywhere: **`199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`**.

## 1. Verified live state (re-checked 2026-08-09)

| thing | state |
|---|---|
| Chassis | **v1.0.1274**, both replicas (was 1262). My Go half **survives the roll**: pod-grep `assigned fact id matches no current evidence_base fact`→1, `fact assignment composed an empty writer block`→1, `normalised object-form sections`→1, negative control `extractSectionNames`→**0** |
| Migration 327 | applied — `site_plan_sections.assigned_fact_ids` present |
| Seed 329 | live — roster block in `build-site-planner` |
| Seed 333 | live — `EVERY section entry on EVERY page` in the same prompt |
| Slice B (328/330) | **still held**, both `_HOLD`-named at HEAD; live configs clean: `page-content-writer` has no `facts_scoped` (0), `page-build-handler` no `section_facts` (0) |
| Current plan | `8ee5807b` (2026-08-07), 71 sections, **17 carrying assignments** |
| Site pages | **25 active, ALL `deployed`**; 4 archived |
| HEAD build | `go build ./platform/orchestration/actions/` **OK** |
| HEAD tests | **RED, not ours** — `TestValidDocSubjectTypes_LockstepWithMigrationCheck` fails at clean HEAD (another lane's `decision` subject-type work, `bugs_open/064`). Reproduced today. Don't chase it; don't let it mask yours |

## 2. Where this front actually stands

Owner decided all three RFC_016 questions on 2026-08-08 (§5, recorded):
rule **ratified** (+ §1a scope clarification), sliced order **approved**,
§3a = **option (a)**. Option (a) is **built and live** (seed 333) and **proven
at the emission** — the planner now assigns facts richly per page
(index/stat-band got F1+F2+F4).

**The blocker is structural and is the whole job now.**
`reconcilePlanWithRealised` Pass B2 (`v3_site_actions.go:3031`, header `:5118`)
restores realised sections over the LLM's for every **deployed** page — correct
and deliberate (bugs 001/037/050/051 lineage: a built page must not be silently
recomposed) — and fact assignments ride *inside* the LLM's section entries, so
they are discarded with them. **All 25 active pages on the acceptance site are
deployed**, so today candidate 1 reaches none of them. Re-measured today; this
is stronger than the "nearly all" in yesterday's handoff.

**Candidate 1b** (designed, NOT built — RFC_016 §3b): (i) prompt — for deployed
pages the planner is shown the realised section list and assigns facts to those
names; (ii) Go — Pass B2 carries facts onto restored sections by component-name
match, logging misses durably. (i) makes (ii)'s match near-total; (ii) alone
carries whatever coincides.

## 3. Do these, in this order

1. **`bugs_open/215` — fix it first.** It gates every clean observation replan,
   and today it got worse, not better (see §4). Dedup by canonical name inside
   `WriteSitePlanAction` (no dedup exists on that path — re-read at HEAD today:
   canonicalise loop `write_site_plan_action.go:274-315`, insert `:355-381`).
   Platform code ⇒ council gate. Independent of everything else here.
2. **Build candidate 1b.** Small, but the Go half touches a bug-laden merge —
   read Pass B/B2's header before editing, and mutation-test the name-match
   carry. Goes **in** the Slice B council round (RFC_016 §3b), never as a bug
   patch.
3. **Revise + submit the Slice B round.** Draft is
   `COUNCIL_DRAFT_slice_b_2026-08-08.json` with a HOLD note inside — **do not
   submit as-is**; add 1b's edits and a fresh compliance observation
   (post-215-fix replan: facts must SURVIVE onto restored sections).
4. ~~**Owner's read of the v4 writer prompt**~~ — **DONE. APPROVED by the owner
   2026-08-09** ("that prompt looks good to me"), on
   `sql/page_content_writer_prompt_v4_2026-08-06.txt`. This was the compliance
   seat's round-1 ask and the last human gate on seed 330; **it is cleared —
   do not re-ask.** Note what it does and does not license: the owner approved
   the v4 TEXT. Seed 330 still applies only after the Slice B council verdict,
   and only after 328 (a v4 prompt without the wiring key sees no
   `facts_scoped` and behaves as v3). **If the text changes for any reason
   after this date, the approval does not travel with it — re-ask.**
5. **Then**: un-`_HOLD` 328/330 → apply 328 then 330 → rebuild flagged pages →
   census: overlap pairs fall on fundamentallyai, the five fact-blind sites
   must not move (the disconfirming half).

Standing/unowned: `bugs_open/214` (imagery scope_refs, small, independent);
Monday contact-sheet cron (08-10, the other front's file tracks it).

## 4. What changed under this front while it slept (the re-check that mattered)

- **My 2026-08-07 replan created three phantom 404 page rows**, which the sweep
  front found and hand-archived on 08-08 (their §2b), also cancelling four work
  items pointing at them. Each was a canonical/stem twin of a page already live
  under the other spelling (`tool-llm-cost-calculator` vs `llm-cost-calculator`,
  `tool-tools` vs `tools`, `ai-readiness-checker-guide` vs
  `tool-ai-readiness-checker-guide`). Verified today: all three created
  `08-07 08:24:22`, `planned`, `deployed_at IS NULL`, now `archived`.
  **This is bug 215's second damage mode** — the same dual-identity defect
  expressed quietly instead of as a crash — and while those rows existed they
  were valid internal-link targets, which is the ammunition behind the other
  front's linkability fix. Recorded in `bugs_open/215`.
- **A replan of this site therefore costs another front real cleanup work.**
  Co-ordinate before firing one, and expect to sweep phantoms afterwards until
  215 is fixed.
- **`bugs_open/215`'s named colliding PAIR was corrected today** — I had read it
  from `llm_plan`/`validate_plan`, but the writer reads `page_plan`/`site_plan`
  (`site_db_actions.go:749-782`). The collision and the missing dedup are
  proven; the pairing is inference and the run's row has expired, so it is
  [UNVERIFIABLE] for that incident. A reproduction must read **`site_plan`**.
  `WRONG_CALLS` 2026-08-09 — the same error class as the 08-08 entry, one day
  later.
- `v3_site_actions.go` was edited by the 210 lane (`2c3efc9f5`, PBP-038) — in
  `UpdatePageStatusAction` only; **Pass B2 and the reconcile passes are
  untouched**, so candidate 1b's target is as described. Checked hunk by hunk.

## 5. Traps (this front's, still live)

- **Completed orchestration rows expire in ~24h.** Pin raw evidence into a doc
  the day you cite it — two claims in this lane are now permanently
  unverifiable for exactly this reason.
- **Cite the line where the failing code READS the data**, not the key you had
  open (`collected_data->'validate_plan'` is post-merge; `llm_plan.result` is
  the emission; the plan writer reads `site_plan`). Two WRONG_CALLS entries in
  two days.
- **A replan on this site IS a build dispatch** (reconcile files `needs_page`,
  claimed within minutes) **and now also a phantom-page generator** (§4).
- **Never `run-migrations.sh --apply`** — the pending list carries other lanes'
  files plus the two `_HOLD` slices.
- **Live row is truth for prompts.** Dump before editing; seed 333 was built
  byte-exact from a live dump by `scratchpad/gen_seed_333.py` (regenerate from
  a fresh dump if you re-run it — the anchors are asserted 1/1/1 and will
  refuse a drifted row).
- Rollback for the prompt seeds: `agent_definitions_bak_329` / `_bak_333`.

## 6. Commit trail (this front)

`d6e9dcf06` decisions + seed 333 applied · `47620cb53` bug 215 filed ·
`c589779a3` the Pass B2 correction across RFC/151/NOTES/README/WRONG_CALLS ·
`9b61d04b1` 08-08 handoff + held council draft · this file's commit (215
correction + WRONG_CALLS 08-09 + this handoff).
