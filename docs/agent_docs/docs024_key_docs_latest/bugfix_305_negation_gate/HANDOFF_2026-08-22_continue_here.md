# HANDOFF 2026-08-22 — continue here (`bugfix_305_negation_gate`)

**Supersedes `HANDOFF_2026-08-21_continue_here.md`**, which is kept for its history: its three
layered banners are the record of the fix being wrong twice and then right, and the middle one
("THE GATE DOES NOT YET CHANGE PAGES") is FALSE as of today. Read this file for state; read that one
for how we got here.

> ## ▶ ONE-LINE STATE
> **The copy gate detects, selects, rewrites, AND changes pages — all four proven on real copy, at
> the artefact, on 2026-08-22.** The lane's engineering is DONE. What remains for the BUG is the
> damage half (the owner's three pages), which is another lane's and partly a decision only the owner
> can make.

---

## 1. What is live, and how each was checked

| thing | state | how it was checked (not "it should be") |
|---|---|---|
| the gate's code | **LIVE on chassis `v1.0.1326`**, both replicas | `grep -ac` in `/proc/1/exe`: `rewrite_negations` 8, `copy_gate` 7, `invented_superlative` 1; controls `rewrite_negationz` / `invented_superlativz` **0** |
| migration `509` (the chain) | applied 2026-08-21 10:28Z | chain reads `generate_content → rewrite_negations → render_section` |
| migration `517` (the `ai_service`) | applied 2026-08-21 10:40Z | the repair had none and was live-but-blind; re-read after today's deploy and still present |
| migration `548` (`render_section` reads the repair) | **applied 2026-08-22 09:19Z**, recorded 09:20:25Z | `UPDATE 1`, verify `DO` passed; live row re-read as `content_from = copy_gate.result`, and re-read AGAIN after the 15:10Z deploy |
| **the repair reaching the stored page** | **PROVEN 2026-08-22 10:00Z** | see §2 |
| the same-field splice race (§20) | **CLOSED at the artefact** | six rewrites in ONE field, six landed; `hits_after` = `hits_before − len(rewritten)` exactly |
| `aiservice.ClassifyTruncation` (MDL-043) | committed, **inert until the next roll** | council `a696e2a3` APPROVED round 1; three mutation-proven tests |
| `brief-negation-check` | daily 07:40 UTC | behavioural probe (reports "9 of 25" + a regulatory column) |
| the owner's three pages | **STILL SERVING IT** — 6 of 9 components, 2026-08-22 | components unchanged since 08-17 20:34Z; brief unchanged since 07-24 |

## 2. The proof, so nobody re-derives it

`loanzy.uk/tool-interest-rate-stress-test`, rebuilt 09:57:01Z (post-548), saved 09:59:25Z.
`copy_gate_2`: **`hits_before 8 → hits_after 2`, 6 rewritten, 0 rejected**; two other sections' hits
**exempt** (brief-supplied — §11's designed behaviour, visible in the same run).

In `page_components.content_data`: **0 of 6 removed constructions present, 6 of 6 replacements
present.** The second count is the demand control — it proves the test could have come out otherwise.

**The control that makes it decisive:** the same query (`RUNBOOK` §8) on
`tool-loan-repayment-calculator`, built 09:10Z — *before* 548 — and whose save was also **accepted**,
returns `stored_matches_PRE=true, POST=false` (the §22 defect). Post-548 it returns `PRE=false,
POST=true`. Same site, same morning, same instrument, opposite answer.

⚠ **Do not verify this with a literal from an earlier run.** A rebuild REGENERATES the copy, so an
earlier run's sentence does not exist in the new one. §8 compares each run against its OWN
`generated_content_<N>.result` vs `copy_gate_<N>.result`. I lost time to this.

## 3. What is left — and only the first is this lane's

1. **NOTHING for the defect half.** Items 1–3 of the old handoff are closed (roll, `548`, artefact
   proof), as are its items (2) superlative-guard and (3) the reuse follow-up.
   The only open thread is a **watch**: no gate run has happened yet on `v1.0.1326`. The code is
   binary-probed present on both replicas, so it has not been REACHED rather than failed.
   > ⚠ **CORRECTED 2026-08-22 18:27Z — I first wrote "this is traffic, not a fault". It is a fault,
   > just not ours.** The Anthropic account hit its usage cap at **18:15:35Z** and the fleet stopped:
   > 116 successful LLM calls in the hour to 18:15:51, then 1 success against 8 usage-limit failures.
   > Page builds now fail at `generate_content`, upstream of the gate. **`bugs_open/243`
   > (resolve by slug: 243-anthropic-cap) — third occurrence in 22 days; only the owner can clear it
   > (credit), else the API states 2026-09-01.** `ai-agent-orchestration.com/adoption-tracker` — one
   > of THIS BUG'S three pages — failed on it at 18:22:02.
   > **So the first-gate-run-on-1326 check is blocked until the account is restored**, and its
   > absence must NOT be read as a gate defect. Recovery: `SELECT max(created_at) FROM llm_call_log
   > WHERE success;` must move and keep moving. First run should carry `has_result=true`, like the 12
   > runs across 6 domains measured on the previous build.
2. **The damage half — another lane's, and now SPLIT in two** (bug file §24):
   - the **6 repairable hits, including both sentences the owner quoted**, need only an ordinary
     rerender — no brief change. This is new: before today a rerender would have reported success and
     changed nothing.
   - the **canonical tagline** (`adoption-tracker` hero) needs `content_direction` edited first. That
     is **D2**, the owner's call.
   ⚠ **Do NOT fire those rerenders yet.** `ai-agent-orchestration.com` is mid-repair by its own lane:
   migration `557` (11:29Z today) rewrote its `evidence_base` because a rebuild had failed twice on a
   claims error, `560` bound its case-study images, and two of the three pages carry open
   `claims_unverified` items filed 16:02Z. A rerender now would likely fail at the claims gate and
   read as this bug.
3. **`bugs_open/366`** (filed today) — `cmd/reasoningset` treats unreported usage as a complete
   answer. Not ours to fix; it changes which rows reach an eval corpus.

## 4. Decisions that are not a session's to make

Unchanged from the previous handoff and still open: **D1** fleet-wide counting annotation (RFC_044),
**D2** the nine briefs, **D3** is `rather than` a tic or ordinary English (43% of sections),
**D4** the `negation_density` threshold, **D5** routing for `brief_supplies_negation` findings.
Full text in `HANDOFF_2026-08-21_continue_here.md` — do not re-litigate them in a session.

## ⏱ UPDATE 2026-08-23 — WHAT CHANGED, AND THE ANSWER TO "CAN WE CLOSE IT"

**The cap cleared** (recovery 2026-08-22 21:42:28Z, a 3h27m outage — `bugs_open/243`, and the owner
acted again rather than the calendar). **Chassis `v1.0.1328`** is live, and it carries yesterday's
`ClassifyTruncation` (probed 1 on both replicas, control 0) — so MDL-043 is no longer inert. Both
config keys survived this deploy too, re-read from the live row.

**The gate ran across the fleet overnight and this morning**, `result=true` on every marker.

**The damage half moved a long way** (bug file §25): the site's lane rebuilt all three pages, and
**`model-directory` — the page carrying BOTH sentences the owner quoted — saved clean at 10:48Z.**
Both quotes are absent from the stored `content_data`, checked as literals with a control. The other
two pages failed at `validate_content` (19 and 5 errors — the claims layer, their lane's work).

**But a NEW defect was found today, in this lane's own mechanism** (bug file §26): 15 of 49 markers
did not reconcile (`targets ≠ rewritten + rejected`), 12 accounting for none of their targets, because
the repair loop ranges over the model's answer rather than over the targets. Fixed, mutation-proven,
council `f3046f0c` **submitted — verdict pending**, and **INERT until the next roll**.

### So: NOT YET — and the blocker is no longer the owner's pages

The estate's bar is *fixed AND live*, and *"a fix committed but inert until the next roll stays OPEN,
because the defect is still reproducible until it ships"*. By that bar this lane has one open,
reproducible defect: the repair log under-reports. It is the honest reason to hold, and it is a
**one-roll wait plus a verdict**, not a dependency on anybody's decision.

⚠ **And it has a consequence worth stating: `D3` must not be decided yet.** That decision was to be
settled from the rejection log, and the log is exactly what was incomplete.

**After that roll**, closure rests only on things that are NOT this defect — `D2` (an owner decision
about a brief the gate exempts by design) and another lane's claims validation — at which point
closing with the residual filed is defensible, on the `bugs_open/198` → `352` precedent.

## 5. Can `bugs_open/305` close? — the ORIGINAL assessment, superseded by the block above

The **defect** is fixed, live and proven. The **damage** is not: the owner asked for two things and
"fix the affected pages" is unmet. Same shape as `bugs_open/327`. What would close it: the aiao lane
edits `content_direction` (⚠ whole object, never a patch — `bugs_open/327`; verify by label presence,
never by diffing, because `formatted` regenerates in random key order) and rerenders the three pages.
Then the §15 canary should return 0 repairable hits with the tagline gone rather than exempt.

## 6. Standing cautions

- **A brief-supplied phrase passes the gate BY DESIGN.** Check the marker's `exempt` count before
  concluding the gate is broken.
- **A marker over a run whose SAVE was refused still describes a page that never changed.** Check the
  parent for `complete_error` (e.g. `bugs_open/253`'s component floor) before blaming the gate. This
  is the surviving half of the fifth-face landmine; the rest is now history.
- **`\y`, never `\b`, in Postgres.** It cost this lane two wrong figures.
- **Place a run either side of migration `548` by `created_at`, not `updated_at`** — a run that
  started before 09:20:25Z can carry a later `updated_at`.
- **`cmd/gatecanary` must never become a real command** — scratch-tree only, `RUNBOOK` §7.
- **The register is a review, not paperwork.** Writing MDL-043 sent me to MDL-042, which contradicted
  my own change and caught a real error (requested vs applied ceiling) minutes after I submitted it.

**Migrations this lane owns:** `509`, `517`, `548` (all applied and recorded).
**Council:** `c48b7612` (the gate, APPROVED r4) and `a696e2a3` (the truncation helper, APPROVED r1).
