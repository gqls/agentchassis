# 161 — the evidence register ratifies the claim it was built to catch

**Filed 2026-07-31** while working `bugs_open/HANDOFF_2026-07-31_checker_layer_remaining_items.md`
§1 (149's C1 recognition gap). It answers §1's question and **falsifies two of its three
fix candidates**, so read this before acting on that file.

> ## RE-VERIFIED 2026-09-03 — the fix HOLDS; three statements below are now FALSE; and the residual is filed as `bugs_open/456`
>
> Resumed by a session asked to check whether this bug was still valid. **It is not — the
> defect is fixed and stayed fixed** — but four things are worth the next reader's time.
>
> **1. Re-verified today rather than carried forward.** `gd-trials` is corrected and still
> carries the first real `artifact_check` fleet-wide; the daily sweep has rewritten
> gamesdesign's register **every day 08-24 → 09-02**; the canary component
> `15f1f798-…` still holds `Math.min(val, 10000)` with **zero** `Math.random`. All six
> repaired pages serve **200 at their recorded `pages.url`** with the false claim absent, the
> deliberately-spared loot-table guide still carries its honest technique teaching, and an
> invented URL 404s. ⚠ **Do not compose those URLs from `pages.name`** — three different URL
> forms are in use on this one site, and doing so returns 404 on all seven plus the control
> (`WRONG_CALLS.md` 2026-09-03; use `scripts/probe-page-url.sh`).
>
> **2. `verified_at` for `gd-trials` now reads 2026-09-02**, so the daily `artifact_check` is
> still running and still passing. The mechanism this bug produced is alive, not merely shipped.
>
> **3. THREE CLAIMS IN THE CLOSE-OUT BLOCKS BELOW ARE FALSE, all corrected here:**
> - ~~"RFC_025 stage 2b (`page_name` addressing)"~~ — **stage 2b shipped 2026-08-24 as
>   `subject_key` addressing** (`eecd99b0a`, `bugs_open/288` §5.6). `page_name` was never the
>   design; the label was coined in this close-out and copied into RFC_025 §11.
> - ~~"the fail-direction of the check live (unit-proven only)"~~ — **an induced live drift was
>   proven 2026-08-24** on mortgagecalculator.co.uk's `sdlt-ftb-relief-cap` (`bugs_open/288`
>   §5b), the same day this was written.
> - ~~"the attestation nudge's first possible firing is ~2027-01 (all attested facts younger
>   than 180 days)"~~ — **wrong on its own code, and it fired on 2026-09-01.**
>   `checkAttestationStaleness` treats a fact with no usable `verified_at` as **due
>   immediately** (*"an undated attestation is not evidence of freshness, it is the absence of
>   the one signal this check has"*). boxingonline.com's undated `business_name` fact raised a
>   `stale_attestation` item on 09-01. The claim reasoned from the facts that existed on 08-24
>   and treated the threshold as the only path to the queue.
>
> **4. THE RESIDUAL IS REAL AND IS THIS BUG'S MECHANISM INVERTED — `bugs_open/456`.**
> This bug was *a register that vouches for a claim it caused*. `[MEASURED 2026-09-03, all 27
> live registers]` **two registers do not parse at all**, so their sites' entire claims layer,
> `banned_claims` included, is OFF: `finetuning.uk` (3 bans, since 08-24) and `noted.co.uk`
> (7 bans, since 08-25). One text-valued fact was enough, because `ParseEvidenceBase` decoded
> `facts` as one array and every caller reads a parse error as "no register". Fixed at source
> (`3f221f99f`); full case, with a before/after control on the live registers, in
> `bugs_open/456`. **The 28 artifact-sourced facts this file's close-out counted are also
> still uncounted by the daily sweep** — its residue arm drops them without incrementing
> anything — which is the same "trusted once registered" one level further out.
>
> **This file stays CLOSED.** Its own defect is fixed, live and driven; 456 is a different
> bug found while re-checking it, not a reopening.

> ## STATUS 2026-08-24 — thread resumed; step 6 VERIFIED AT THE SERVED URL; the structural fix went and came back ratified, built, live and (today) driven
>
> The 07-31 thread went quiet after the status block below; no commit has touched this
> file since. Resumed 2026-08-24. What has happened since 07-31, verified today rather
> than carried forward:
>
> **1. Step 6 (the served pages) is DONE — verified at the served URL, with controls.**
> All 6 affected pages have `deployed_at` between 2026-08-19 and 2026-08-23, well past
> the 15:28 repair. Curled the live site 2026-08-24: the false claim ("10,000 Monte
> Carlo trials per query" / any 10,000-adjacent Monte Carlo attribution) is absent from
> all six. Controls: an invented URL returns **404** (the domain genuinely serves — not
> parked); the deliberately-spared `tool-loot-table-balancer-guide` still carries its
> honest technique-teaching mentions ("Run Monte Carlo simulations across at least
> 10,000 simulated players" — reader advice, not a claim about our tools), proving the
> grep sees what is there; and `tool-spawn-rate-balancer-guide`'s one remaining mention
> is the corrected copy stating the tools compute the distribution exactly *rather
> than* sampling. So the per-site defect is fixed AND live AND served.
>
> **2. Fix candidate 3 became RFC_025 and is ratified, council-approved, and live.**
> `architecture_review/RFC_025_artifact_sourced_facts_are_trusted_once_registered.md`
> (opened and RATIFIED 2026-08-12 by the portfolio_positioning lane; council APPROVED
> round 2, corr `9fd94852-ff79-496b-96b5-78a8d3619162`) shipped BOTH halves of
> candidate 3: **stage 1** — the `stale_evidence`-style cadence item for prose-sourced
> facts, as a 180-day `stale_attestation` nudge on `attested_by` facts; **stage 2** —
> `source.artifact_check`, the grep-shaped assertion (`component_id` + `pattern` +
> `must_be_present`) that lets an artifact-sourced fact fail like a sql-sourced one.
> Live on the fleet since chassis v1.0.1295, 2026-08-13, pod-verified per that RFC's
> §10. The council's round-1 REVISE also closed this file's own landmine *inside* the
> mechanism: a bare-numeric pattern (the `10000`-matches-`100000` trap from "Traps this
> cost me") is refused at parse time.
>
> **3. The residual found on resumption: the mechanism had never been driven.** Measured
> 2026-08-24: **zero** facts fleet-wide carried `artifact_check`, and no
> `stale_attestation` item has ever been raised (expected — every attested fact is
> younger than 180 days; earliest possible firing ~2027-01). RFC_025 §5.3's own staged
> plan names the canary: retype `gd-trials` itself with a real check on the input-clamp
> line. **Migration `585_bug161_arm_artifact_check_canary_on_gd_trials.sql`** (this
> session, council corr `a9e1a0de-ff04-4193-83dc-ad67f2d4d83d`) does exactly that:
> pattern `Math\.min\(val,\s*10000\)` against tool-drop-rate-simulator's hero component
> `15f1f798-51fb-41d0-8a07-18148b39a293` (still untouched since 2026-06-05; still zero
> `Math.random` — the decisive negative holds). The daily `evidence-freshness` sweep
> covers gamesdesign despite its zero sql facts (`resolveEvidenceSites` has no sql
> filter — the "structurally blind to 4 of 9 registers" finding below was about what
> the sweep could *check*, not which sites it visits, and stage 2 is what makes the
> visit able to check something). `verified_at` is deliberately left at 2026-07-31 so
> the first sweep's bump is a demand-control proof the check ran.
>
> **4. Blast radius re-censused, because the 07-31 count is stale by growth: 19
> registers, 294 facts as of 2026-08-24** (was 9/102). Of the 262 non-sql facts, 185
> are `citation`-sourced (machine-reverified every sweep via V5), 133 `attested_by`
> (nudged at 180 days via stage 1), and **28 are `artifact`-sourced with no
> `artifact_check` yet** — the remaining adoption surface, deliberately human-paced
> per the ratified RFC ("facts are retyped per-site, not migrated in bulk").
>
> **CLOSE CRITERIA: apply 585 after its council verdict, dispatch/await one
> evidence-freshness pass on gamesdesign, confirm `verified_at` bumped (check ran,
> passed), then move this file to `bugs_closed/` and mark RFC_025 IMPLEMENTED.**

> ## CLOSED 2026-08-24 — fixed AND live AND driven; every close criterion met the same day
>
> - **585 council-APPROVED round 1** (corr `a9e1a0de-ff04-4193-83dc-ad67f2d4d83d`, 2
>   advisories, both answered with edits: every guard comparison made NULL-proof via
>   `IS DISTINCT FROM`, `jsonb_typeof(source)='object'` added as a named precondition,
>   `_ROLLBACK` sidecar written — the NULL-blind `<>` arm is logged in `WRONG_CALLS.md`).
>   Applied by hand ~11:20Z after a doomed-transaction rehearsal ran the whole file to
>   its own COMMIT; recorded in the ledger via `--record-only`.
> - **The canary ran and PASSED, live, the same day.** A single-site
>   `evidence-freshness` dispatch (orch `ac49d67e-3f86-4034-a666-64737ed1b001`,
>   `sites_checked=1`, published with an asserted kcat receipt) executed the check:
>   per-fact outcome `fresh` / tolerance `artifact_check`, `verified_at` bumped
>   `2026-07-31 → 2026-08-24` (the demand control this migration deliberately built
>   in), register rewritten by `evidence-refresher` — **and the `artifact_check` key
>   survived that rewrite**, verified at the artefact, so the check re-arms daily.
> - **Binary capability probed, not assumed:** both chassis replicas carry the
>   `artifact_check.pattern` literal in `/proc/1/exe` (present=2 each, absent-string
>   control=0), answering the council's prior_art_librarian.
> - **What closes with this bug:** the self-ratifying-register defect (per-site fix
>   served; structural mechanism live and now exercised on the motivating case). The
>   fact that motivated the mechanism is the first fact the mechanism guards.
> - **What does NOT close with it, and where it lives:** the fail-direction of the
>   check live (unit-proven only; the council's bug_historian noted a follow-up
>   induced-drift test would be worth an owner-sanctioned run); RFC_025 stage 2b
>   (`page_name` addressing) and the prose half of the encoded-figure class —
>   both tracked in `bugs_open/288`; the qualitative-credential gap ("built BY a
>   shipped designer") — tracked in `bugs_open/149`; the remaining **28** (as of
>   2026-08-24) artifact-sourced facts with no check — adoption is per-site and
>   human-paced by ratified design; the attestation nudge's first possible firing
>   is ~2027-01 (all attested facts younger than 180 days).

> ## STATUS 2026-07-31 (later) — FIXED AT SOURCE AND ARMED. Open only until the served pages catch up.
>
> Owner authorised the repair ("take on bugs_open/161") and chose the **coherent rewrite**
> plus **arm the patterns**. What is done, each verified at the artefact rather than at the
> COMMIT:
>
> | step | state |
> |---|---|
> | **1. the false fact** | **CORRECTED, LIVE.** `270_bug161_…sql`. Supersede-then-insert, old row kept as history. `gd-trials` claim → "maximum attempts modelled per query", `context_terms` no longer include "monte carlo"/"simulation"/"trial", source now cites `Math.min(val, 10000)`. All 4 facts preserved. |
> | **2. the writer whitelist** | **CORRECTED, LIVE.** `writer_block` now states the tools compute exact probability and **explicitly prohibits** attributing sampling to them, while still permitting the technique to be taught. |
> | **3. every repo reseed vector** | **CORRECTED.** `218_evidence_facts_for_043_sites.sql:139`, `SQL_2026-07-24_evidence_base_four_sites.sql:48`, and `SQL_2026-07-24_gamesdesign_index_stats_traced.sql` (both its trace note and its `stat3_label`). `bk_site_specs.sql` deliberately untouched — a backup should record what was. |
> | **4. the published copy** | **REPAIRED in the DB**, 14 replacements across 10 components, `content_data` **and** `rendered_html`, every replacement asserted to have matched in both before and to be absent after. |
> | **5. detection** | **ARMED, LIVE.** `271_bug161_…sql`, two per-site `banned_claims` patterns with order-enforcing guards. |
> | **6. the served pages** | ⚠ **STALE — the only thing outstanding.** See below. |
>
> **The measurement that proves 5 is precise, not just present** — `cmd/claimscan` over the
> complete live corpus (67 components, export row count asserted 67/67, no base64
> truncation stubs):
>
> | configuration | findings |
> |---|---|
> | current register, original copy | **0** ← the whole bug: 10 live falsehoods, scanner clean |
> | corrected register, original copy | **0** ← correcting the register flags NOTHING (see the correction below) |
> | corrected register **+ patterns**, original copy | **18 findings on exactly 10 components** — all 10 false ones, and it **spares** the one legitimate component |
> | corrected register + patterns, **repaired copy** | **0** |
>
> **⚠ STEP 6, stated plainly: the fix is in the database and NOT yet on the website.** The 6
> affected pages were last deployed **12:51–12:53**; my repair landed **15:28**. So the served
> HTML still asserts the falsehood. `build_status='needs_rebuild'` is **not** a route — that
> queue is dead (44 pages, oldest stuck since **2026-04-23**). What will fix it: the
> automated `rerender-pages` sweep queued a full-site gamesdesign rerender at **15:28:33**
> (34 items, `reason` NULL → `check_rerender_mode`'s else branch → `render_page`,
> "assemble stored HTML"), which is exactly the mode that redeploys the corrected
> `rendered_html` without regenerating anything. The dispatcher is alive (1,261 complete,
> last touched 15:38) but ~100 minutes behind, so it was left to drain rather than
> double-dispatched into another lane's sweep. **Verify with:**
> ```sql
> SELECT p.name, p.deployed_at > '2026-07-31 15:28:35+00' AS carries_the_fix
> FROM pages p WHERE p.site_id='e33263f4-74f8-494f-b191-546845dbbddf'
>   AND p.name IN ('guide-skinner-box','guide-rng-design','tool-spawn-rate-balancer-guide',
>                  'game-auto-battler','game-economy-simulator','guide-fairness-in-rng');
> ```
> If it has not drained, fire `page-rerender` for those 6 page ids directly with a `reason`
> **not** in (`image_landed`, `section_data_resolved`, `cta_links_stale`) so it stays
> assemble-only. **Then confirm at the served URL, not at `rendered_html`.**
>
> ### Corrections to this file's own first draft, all found by re-measuring
>
> 1. **"three deployed components" was wrong — it is ELEVEN carrying the phrase, TEN false.**
>    I undercounted from reading truncated psql output. The ten: `game-auto-battler`,
>    `game-economy-simulator`, `guide-fairness-in-rng` (×2), `guide-rng-design` (×2),
>    `guide-skinner-box` (×2), `tool-spawn-rate-balancer-guide` (×2).
>    `tool-loot-table-balancer-guide` carries only general technique teaching and reader
>    advice — **true, and deliberately left alone.**
> 2. **"Doing this ARMS the gates against the surviving components" was WRONG.** Measured: 0
>    findings before and after the register fix. Every affected page is `guide`, `game` or
>    `blog-post` — all in `editorialPageTypes`, exempt from `ScanUnregisteredNumbers`. The
>    file said this in one section and the opposite in another; the second was right. What
>    arms detection is `banned_claims`, because `ScanBannedClaims` "runs on every surface".
> 3. **The 043 audit's trace note was wrong TWICE, not once**, and this is the sharpest thing
>    here. It wrote: *"stat3 'Monte Carlo Trials 10,000' — TRACES: the deployed drop-rate
>    simulator/tuner … run 10000 iterations per query in their shipped JS. TRUE; kept."* It
>    caught three of four stats as fabricated and marked this one TRUE. **(a)** "Monte Carlo"
>    is false — no randomness at all. **(b)** "10000 iterations per query" is **also** false:
>    the tuner's CDF is sized `maxKills = Math.max(1, kph * hours)` — from the *user's*
>    inputs — and the only real 10000 is `Math.min(val, 10000)`, a clamp. **The trace found
>    the digits and took the technique word from the copy it was auditing.** That is the
>    mechanism by which an audit launders the thing it is auditing.
> 4. **A multi-session hazard, recorded because it was luck.** The `rerender-pages` sweep
>    queued its rerenders at 15:28:33, *inside* my repair window (15:28:04–15:28:34). Had my
>    repair finished seconds later, that sweep would have deployed the FALSE copy and read as
>    a completed rerender. On a shared tree, a data repair races the sweeps that publish it.

**Status: was OPEN/unowned when filed. Structural + cross-cutting. The register was
human-owned by design, so the original filing changed nothing and asked for a ruling; the
ruling came and the fix above is applied.**

> ## Diagnosis loop: UNVERIFIABLE at the iteration cap — every substantive element corroborated, blocked by a harness truncation
>
> Run per the **owner ruling of 2026-07-31** (a `bugs_open/` file asserting a cross-cutting
> or structural root cause must go through the loop, or the session must state plainly why
> it substituted equivalent first-hand verification). Intake
> `93cc6cef-39c6-42b0-9861-ab80a235740e`, run `08ff91c4-dfa7-4226-a039-e80a08e44cc1`,
> 5 iterations, ~14 minutes.
>
> **Verdict as returned, verbatim and not softened:** `status = UNVERIFIABLE`,
> `conclusion = NOT CONFIRMED (stopped: iteration-cap)`, `stopped_by = iteration-cap`.
> **It is not a REFUTED and it is not a CONFIRMED.** Read what it actually established
> before using either word:
>
> **What the loop reached independently, in its own words:**
> - iteration 2 — *"The mechanism half is directly shown: `numberSupported` (claims.go)
>   compares only val against `f.Value`/`tolerance`/`context_terms` and **never reads
>   `Source.Artifact`, `Source.AttestedBy`, or `f.Claim` wording** — an artifact- or
>   attestation-sourced fact's number is trusted once registered."*
> - it found the platform's own doc on `EvidenceSource`: *"Artifact: code path or URL
>   evidence — **checked for presence in the register, not re-proved.**"* And the
>   corresponding code at `refreshOneSiteEvidence`:
>   `if query == "" { continue // artifact/attested facts are checked for presence, not re-proved }`
> - iteration 4 — *"one concrete mismatched fact (gd-trials: claimed '10,000 Monte Carlo
>   trials', but **the shipped tool's own doc-comment describes a geometric-distribution
>   closed-form model, not a trial-based simulation**) that was in fact published and
>   persisted repeatedly."* It cited that doc comment directly: *"Cumulative distribution
>   modelled via geometric distribution with optional hard pity cap."*
>
> So it corroborated the mechanism AND the specific false fact, from the code and the live
> DB, without being shown this file. That is independent agreement on both halves.
>
> **Why it still would not confirm, and this is the transferable part.** Its own iteration-5
> note: the tuner script *"has now been fetched twice and truncates at the identical point
> … this looks like a fixed truncation on the `rendered_html` column itself, not a
> fetch-window problem."* Measured against my complete export: **the loop saw 10,704 of
> 21,527 characters — 49.7%.** The `geometric distribution` doc comment sits at offset
> 10,649, fifty-five characters *inside* the cut, which is exactly why it could cite the
> comment. But `Math.pow` is at 13,047 — **beyond the cut** — so the loop could never read
> the actual computation, and `Math.random`'s **absence** (the decisive negative evidence)
> **cannot be established from a half-read column at all.** You cannot prove an absence
> from a truncated read.
>
> **So the loop declined to confirm for a correct reason**, and the gap is one a direct
> export closes: `psql` with the row's own `length()` asserted against the exported bytes
> (21,478 in the DB, 21,566 exported including psql's headers; stderr empty, no
> `unexpected EOF`). That is the substitute this filing relies on, declared as the ruling
> requires — **not** "first-hand verification instead of the loop", but the loop run, its
> corroboration recorded, and the one thing it structurally could not read supplied by a
> method that can.
>
> **A harness finding worth its own attention:** any hypothesis turning on the ABSENCE of a
> symbol in a large artefact column is **unconfirmable by the diagnosis loop today**,
> because its `data_request` path truncates around ~10.7KB and an absence claim needs the
> whole column. This is not `bugs_open/012`'s max_tokens truncation — it is on the read
> path, silent, and reproducible. Now in `LANDMINES.md`.

---

## The one-sentence version

A site's evidence register is simultaneously (a) the whitelist of claims the writer is
**instructed** to make and (b) the authority every claims gate checks published claims
**against** — so a false fact in the register is **self-ratifying**: it causes the claim,
then vouches for it, and no gate in the layer can ever object.

`gamesdesign.co.uk` has one. It has been live since 2026-07-24 and is **still served today**.

## What is false

`site_specs` / `aspect='evidence_base'` / `is_current`, fact `gd-trials`:

```json
{"id":"gd-trials","claim":"Monte Carlo trials per query","value":10000,"kind":"metric",
 "tolerance":"exact","context_terms":["trial","monte carlo","simulation"],
 "source":{"artifact":"the figure is hard-coded in the shipped drop-rate tool JavaScript"},
 "verified_at":"2026-07-24",
 "writer_line":"{value} Monte Carlo trials per query in the drop-rate tools"}
```

**Neither drop-rate tool performs any Monte Carlo simulation, and neither contains any
randomness at all.** [VERIFIED 2026-07-31, by reading the shipped component HTML/JS out of
`page_components.rendered_html`, export length asserted against the DB's own `length()`]

| tool | page id | what it actually computes | `Math.random` | `monte` | the `10000` in it |
|---|---|---|---|---|---|
| `tool-drop-rate-tuner` | `b381f0db-…` | `Math.pow(1-p, k)` — closed-form geometric survival, then a CDF array indexed by kill count (`:452-463`) | **0** | **0** | none — the apparent match is `if (pity <= 0 \|\| pity > 100000)` |
| `tool-drop-rate-simulator` | `0f9ed454-…` | says "binomial" 6× | **0** | **0** | `return Math.min(val, 10000)` — an **input clamp** on attempts |

A Monte Carlo method *is* random sampling. With zero randomness in either artefact there is
no trial count to be "hard-coded", so the fact's own cited source does not support it.

**The number 10,000 is real; its stated MEANING is wrong.** It is the maximum attempts the
simulator will model, which is exactly what another session independently concluded when it
repaired the homepage on 2026-07-30 (see the timeline).

## It was false when it was registered — not stale since

This matters, because "went stale" would be `refresh_evidence_base`'s job and "false on
arrival" is nobody's.

- `tool-drop-rate-simulator`'s component was created **2026-06-05 20:12:52** and its
  `updated_at` is **the same timestamp** — untouched for seven weeks before the register was
  written, and untouched since. So the JS I read today is byte-for-byte the JS that existed
  on 2026-07-24. [VERIFIED]
- `tool-drop-rate-tuner` has **zero rows** in `page_component_history` mentioning
  `monte carlo` or `Math.random` — the tool artefact never carried either. [VERIFIED]

## The sequence, and it runs the opposite way round to the obvious guess

Timestamps are from `page_component_history.created_at` and `site_specs.created_at`.

| when | what |
|---|---|
| **2026-06-06 16:59** | gamesdesign's homepage copy already asserts "Monte Carlo" — **seven weeks before any register existed**. Origin is outside this bug. |
| 2026-06-22 → 2026-07-20 | four more saves, claim carried forward each time |
| **2026-07-24 14:20:34** | the homepage is saved again, still asserting it |
| **2026-07-24 14:22:40** | `bugs_closed/043`'s wave-1 audit **creates the register** — 2 minutes and 6 seconds later — registering "Monte Carlo trials per query" as a verified fact and attributing it to the shipped tool JavaScript |
| 2026-07-29 17:26–18:10 | the rewrite wave witnessed in `149` C1 restates it across many components |
| **2026-07-30 14:31** | another session repairs the homepage stat card to `stat2_label:"Max Attempts Modelled"`, description "computes **exact binomial probabilities** for up to ten thousand". **That session got the truth right.** But it corrected the copy only — not the register, not the migration |
| **today** | the register still says "Monte Carlo", so it now **contradicts the repaired page**, and will re-supply the false line to the next rewrite |

**So the audit that was supposed to give writers "a register of true ones to reach for"
(`bugs_closed/043`) appears to have derived this fact from the very copy it was auditing,
and then cited an artefact that does not contain it.** The two-minute gap is
circumstantial, not proof of method — but the artefact evidence above is decisive on the
outcome either way: what got registered was not what the artefact said. [VERIFIED on
outcome; the derivation-from-copy route is [INFERRED] from the timing]

## Why no gate can catch it — the structural part

Both halves of the layer consult the same register, so both are disarmed by the same row.

**1. It is fed to the writer as an instruction.** `refresh_evidence_base_action.go:16-18`
— `writer_block` is "consumed by the page-content-writer prompt … so the numbers the writer
is permitted to assert can never quietly rot". gamesdesign's live `writer_block` (600 bytes)
reads, verbatim:

```
NUMBERS (state only these, with their listed meaning; as of 2026-07-24):
- 11 interactive design tools live, all client-side and free
- 10,000 Monte Carlo trials per query in the drop-rate tools (the figure is in the shipped tool code)
- 4 configurable inputs in the drop-rate tuner: drop chance, kills per hour, pity timer, target hours
- 10 guides & articles live (5 blog posts + 5 guides)
NOT TRACKED / DOES NOT EXIST, NEVER STATE: user counts, accuracy-gap percentages, …
```

The writer did **not** invent this claim. It was handed to it under "state only these",
with a parenthetical asserting the artefact backs it.

**2. It is the authority the gates check against.** `datahelpers/claims.go:931`
`numberSupported(10000, window)` walks `eb.Facts`; `gd-trials` has `context_terms`
including `"monte carlo"`, the window contains it, tolerance is `exact`, `10000 == 10000`
→ **supported → skipped**. Every consumer of that one function is disarmed at once:

- the prose scan (`ScanUnregisteredNumbers` → `claims.go:778`)
- the stat-field audit (`claims_stats.go:327`, `ScanStatClaims`)
- the persistence floor shipped for 149 C1 (`save_sections_claims_guard.go`, same engine by
  design — `:110-114`)

**This is why the handoff's §1 measurement returned 0 findings, and the 0 was CORRECT.**
§1 read that silence as a vocabulary blind spot in `businessClaimContextRe`. It is not:
the number is registered, and a scan that skips a registered number is working.

## Consequence for the handoff's §1 candidates — two of the three are inert

Measured against the motivating case, not argued:

1. **"Widen `businessClaimContextRe` to cover technical/product vocabulary" — INERT.**
   Widening makes the number *reach* `numberSupported`, which then matches `gd-trials` and
   correctly skips it. Net effect on the witnessed fabrication: **zero**. It also buys the
   false-positive risk `claims.go:612-617` already records from the last widening. **Do not
   spend a council round on this.**
2. **"A structural stat rule rather than a lexical one" — INERT, and largely already built.**
   `ScanStatClaims` is *already* purely structural (no lexical gate at all) — §1's premise
   that this needs building is wrong. It also calls `numberSupported`, so it is disarmed by
   the same row. On §1's sharper sub-question — *"check whether the stat audit ran on that
   page and what it said; if it stayed silent that is a more serious finding"* — the answer
   is that it would have stayed silent **and been right to**, for this claim.
3. **"A claim-vs-source diff at rewrite time" — the only one that survives**, and only for
   the *other* half of the witnessed case (below). It would not flag the Monte Carlo claim,
   which is unchanged from the register's own words.

## The other half of the witnessed case still stands

§1's second bullet is correct and untouched by this bug: **"built *by* a shipped
live-service designer"** (from "built *for* live-service and tabletop designers") is a
fabricated **human credential** — non-numeric, matching no banned pattern. gamesdesign has
**0 `banned_claims`**. [VERIFIED] The engine has no shape for a qualitative personnel claim.
That remains a real gap and is *not* what this bug is about.

## Blast radius — measured fleet-wide, 2026-07-31

```sql
SELECT s.domain, jsonb_array_length(COALESCE(sp.data->'facts','[]'::jsonb)) AS facts,
       (SELECT count(*) FROM jsonb_array_elements(sp.data->'facts') f
          WHERE NOT (f->'source' ? 'query' OR f->'source' ? 'sql')) AS prose_sourced,
       COALESCE(sp.data->>'writer_block_managed','(absent)') AS wb_managed
FROM site_specs sp JOIN sites s ON s.id=sp.site_id
WHERE sp.aspect='evidence_base' AND sp.is_current ORDER BY 2 DESC;
```

| domain | facts | prose-sourced | wb_managed |
|---|---|---|---|
| oufe.com | 36 | 35 | (absent) |
| leopardessconsulting.co.uk | 18 | 9 | true |
| fundamentallyai.com | 15 | 8 | true |
| relojistas.com | 13 | **13** | (absent) |
| ai-agent-orchestration.com | 7 | 2 | (absent) |
| robot-hands.com | 5 | 2 | (absent) |
| **gamesdesign.co.uk** | 4 | **4** | (absent) |
| vonc.com | 4 | 0 | (absent) |
| finetuning.uk | 0 | 0 | (absent) |

**9 registers · 102 facts · 73 (72%) prose-sourced.** `refresh_evidence_base` only refreshes
a value from a `sql`/`query` source, and it "never rewrites the human-authored WORDS"
(`refresh_evidence_base_action.go:20-23`) — so **no mechanism re-checks a prose-sourced fact,
ever, and none checks any fact's `claim` wording against its artefact.**

Confirming detail: the live `stale_evidence` work items name **oufe, leopardess,
ai-agent-orchestration, fundamentallyai** — precisely the four registers that have
sql-sourced facts. gamesdesign, relojistas, vonc and robot-hands can never raise one.
[VERIFIED] The freshness mechanism is structurally blind to 4 of 9 registers, and
`relojistas` (13/13 prose) and `gamesdesign` (4/4) are wholly outside it.

## What is still being served

> **CORRECTED 2026-07-31 — this table listed THREE components; the real count is ELEVEN
> carrying the phrase, of which TEN are false.** I undercounted from truncated psql output.
> All ten are now repaired in the DB (see the status block); the full list and the one
> deliberately-spared component are there. The table below is left as filed, understated.

`page_components`, `build_status='deployed'`, 2026-07-31:

| page | slot | text |
|---|---|---|
| `tool-spawn-rate-balancer-guide` (`blog-post`) | `call-to-action` | "The drop-rate tuner runs **10,000 Monte Carlo trials per query**. You set the 4 configurable inputs…" |
| `tool-spawn-rate-balancer-guide` | `article-body` | contains it |
| `tool-loot-table-balancer-guide` (`blog-post`) | `article-body` | contains it |

Both pages are `page_type='blog-post'`, which `editorialPageTypes` (`claims.go:721-723`)
exempts from the prose number scan — so even correcting the register does **not** make these
three components get flagged. They need a direct repair. The homepage is already correct.

## Fix candidates — ordered by what makes the bad state unrepresentable

**The register is human-owned by design ("Truth decisions stay human",
`refresh_evidence_base_action.go:20`), so 1 and 2 are OWNER CALLS, not a thread's to take.**

1. **Correct the fact, in both places, or a reseed reintroduces it.** The false row is in the
   **committed migration**: `docs/agent_docs/sql_for_agents/218_evidence_facts_for_043_sites.sql:139-143`.
   Fixing only the live row leaves the repo able to restore it. Proposed correction, matching
   what the artefact and the 07-30 repair both say: `claim` → "maximum attempts modelled per
   query", `writer_line` → "{value} maximum attempts modelled in the drop-rate tools",
   `context_terms` → drop `"monte carlo"` and `"simulation"`, keep `"attempt"`/`"trial"`,
   `source.artifact` → "`Math.min(val, 10000)` input clamp in tool-drop-rate-simulator".
   ~~**Doing this ARMS the gates against the three surviving components** — that is the point,
   but it means the next build on those pages starts objecting, so sequence it with 3.~~
   > **CORRECTED 2026-07-31 — FALSE, and measured to be false: 0 findings before the register
   > fix and 0 after.** Every affected page is `guide`/`game`/`blog-post`, all in
   > `editorialPageTypes`, so `ScanUnregisteredNumbers` never runs on them. Correcting the
   > register stops the writer being *instructed* and stops the engine *vouching* — it does
   > not detect the already-published copy. **`banned_claims` is what arms detection**, because
   > `ScanBannedClaims` runs on every surface (`claims.go`, and it says so explicitly). Two
   > patterns are now live; the sequencing that DOES matter is the opposite of what this
   > paragraph said — repair the copy FIRST, because a banned claim is BLOCKER severity and
   > would otherwise make 6 pages unsaveable with the falsehood still published.
2. **Then repair the three deployed components.** They assert a technique the tools do not
   use. `blog-post` exemption means no gate will raise them; they need naming explicitly.
3. **Make "the artefact backs this" checkable rather than prose.** The generalisable fix, and
   the reason this is a bug and not a typo: today `source.attested_by` / `source.artifact`
   are free text that nothing can verify, while `source.query` is machine-checked every
   sweep. 73 of 102 facts are in the unverifiable class. Options, cheapest first:
   - a **`stale_evidence`-style item for prose-sourced facts on a cadence** ("this fact has
     never been machine-verified; re-attest it") — turns silence into a queue;
   - **`source.artifact_check`**: an optional grep-shaped assertion (path/table + a pattern
     that must be present) so an artefact-sourced fact can fail like a sql-sourced one. This
     is a **new shared vocabulary key on a shared mechanism** — architecture scope under
     CLAUDE.md's seam rules, RFC before code.
4. **Reject: "make the audit that seeds registers read the artefact."** Right instinct, wrong
   lever — it is a prompt/procedure change, and a procedure is not an enforcement mechanism.
   It also cannot help the 102 facts already registered.

## How to verify any fix

- **Unit, on `numberSupported`,** with `gd-trials` corrected: "10,000 Monte Carlo trials per
  query" must become **unsupported** while "10,000 maximum attempts modelled" stays
  supported. A fix asserted only on the second half is half a fix.
- **Mutation-check it** — revert and confirm the new test fails.
- **Re-run `cmd/claimscan` over gamesdesign** before and after. Note it scans prose and stat
  fields via `ParseEvidenceBase`, and **has no stat-claim path of its own** (`0` hits for
  `ExtractStatClaims`/`ScanStatClaims`; positive control `ParseEvidenceBase` = 12)
  [VERIFIED 2026-07-31] — so it predicts the prose half only. Do not read a clean claimscan
  as clearing the stat fields.
- **Serve-side, not stored-side**: confirm the three components are gone from what is
  actually served, per `bugs_open/`'s standing rule that stored HTML is not the artefact.

## Traps this cost me, now in LANDMINES.md

- **A registered fact makes a green claims gate meaningless as evidence of truth.** "The
  scan returned 0 findings" is compatible with "the claim is false and the register says
  otherwise". Always ask *which fact matched* before reading a pass as clearance.
- **`page_components` has no provenance column and `page_component_history.source` does not
  rescue it** — the column is `save_page_sections_overwrite` for **12,386 of 12,416 rows**,
  every pipeline writing the same literal; the rest are hand-typed operator strings.
  [VERIFIED] So the handoff §3's "`ApplySectionEditAction` cannot be bounded from
  `page_components`" **stands**, and history does not bound it either. Saves the next thread
  the query.
- **Grepping a tool for `10000` matches `100000`.** Both of this bug's apparent
  "the figure is in the code" hits were a `pity > 100000` bound and a `Math.min(val, 10000)`
  clamp. Print the match in context; a bare count would have confirmed the false fact.

## Relations

- `bugs_closed/043` — the parent. Its remediation seeded this register; its line 299 asserts
  "the writer_block lists each site's **true** countables", which is the assumption this
  bug falsifies. **Not reopening it** — 043's fix works; this is a defect in one seeded row
  and in the absence of any check on seeded rows.
- `bugs_open/HANDOFF_2026-07-31_checker_layer_remaining_items.md` §1 — answered here; two of
  its three candidates are inert.
- `bugs_open/149` C1 — the ⚠ banner's "the floor would not have caught it" is **correct but
  for the wrong reason**: not narrow vocabulary, but a register that vouches for the claim.
- `bugs_open/102` — the `blog-post` editorial exemption that keeps the three surviving
  components unflagged even after a register fix.
- `bugs_open/151` — "a rewrite is the moment unbacked claims get laundered". Same family,
  one level up: here the *register* does the laundering.
- `bugs_open/147` — robot-hands' "independently verified". Adjacent but different: that one
  the engine **caught**, because it is a banned claim rather than a registered one.
- `CLM-003` / `CLM-014` / `CLM-018` — concept register, `claims-verification.md`.
