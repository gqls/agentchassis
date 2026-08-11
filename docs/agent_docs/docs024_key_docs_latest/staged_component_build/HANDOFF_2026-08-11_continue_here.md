# HANDOFF — 2026-08-11, fresh chat starts here: both storage proofs LANDED; what's left is decisions and batch-8 tail

**Supersedes `HANDOFF_2026-08-10b_continue_here.md`** for state and work-list. Still binding
from the chain: the 08-09 handoff's §0 (shared-228) and §2 (rerender traps), the 08-08
handoff's §3 (interactive-fence line), and the 08-10 handoff's ADDENDUM (batch-8
requalification, the `computed_values` corrections, the two-session coordination traps —
read it before authoring any new fence).

## 1. State (verified 2026-08-11 ~10:00Z)

- **59 subjects proven end-to-end: 54 sections + 5 tools** (setup-builder by session A;
  grip-force + matchmatrix by session B, 08-10). Naming contract: PASS, 54 canonical /
  28 testable / 10 backlog / 0 BROKEN (08-10 figures — re-run before quoting).
- Fleet: **v1.0.1284** (chassis pods up 09:26Z, browser-runner 09:23Z, all markers green).
  Re-grep at session start.
- The whole-fleet tag bump (`kustomization.yaml` × 19 services) sits **uncommitted** in the
  tree — the owner's release, not this lane's to commit. Pathspec your commits around it.

## 2. THE TWO PROOFS FROM 08-10b §2 — BOTH DONE 2026-08-11 (morning)

Full evidence: NOTES `## 2026-08-11 (morning session)`, and the UPDATE blocks appended to
both bug files. Compressed:

- **bugs_open/243 (candidate 1): FIXED + LIVE + PROVEN.** Driven, not waited for — the
  due-sweep suppresses any tool with a verdict <7 days old, so it could never have produced
  the proof this week. Work item `ae33ed59…` (tool-setup-builder@dartsonline) → spawned pod
  `agent-tool-acceptance-agent-649a6c11-q9mlk` → run `0ee53904…` **`complete`**, no step
  error, 15/0/9, **first-ever `tool-acceptance-agent` rows in `llm_call_log`** (look /
  anthropic / claude-sonnet-5 / success / 2 images / 0 dropped).
- **bugs_open/245: proven and the overlay half executed.** Spawned pod's four storage env
  vars all `secretKeyRef → personae-storage-secrets`; a real authenticated B2 READ happened
  inside that pod (the screenshots the vision model consumed). Both greps re-run (0 direct
  readers; `prepare_training_data`'s unconditional client noted in the bug file). Overlay
  lines 76–98 **removed**, tombstone comment in place, kustomize builds clean with 0
  credential names. **Residual**: the removal reaches the standing deployment at the next
  `apply -k`/release — after it, check `env | grep -c B2_` → 0 on each chassis replica.

**The find of the day: the vision half's FIRST run found a real defect** — near-invisible
low-contrast text on several dartsonline setup-builder options and the CTA, desktop AND
mobile, while all 15 selector checks pass. Recorded in 243's update. Two consequences below
(§3 items 1 and 4).

## 3. Decisions and follow-ups, in order of who owns them

**Owner decisions (also summarised in the chat that cut this handoff):**

1. **243 candidate 3 — make vision findings visible.** Today's run is the worked example:
   vision names a defect, the run reads green, nothing is raised, the text sits in
   `collected_data->'look'` where nobody looks. Recommend: build it (small chassis change,
   normal council gate).
   > **MEASURED 2026-08-11 (parallel session) — the gap is total, not partial.**
   > `grep -rn "render-critique" --include=*.go platform/ internal/ pkg/` → **0 hits**. The
   > only match for `critique` in live `agent_definitions` is `tool-acceptance-agent`, the
   > producer. And this run's note is the **first and only `render-critique` row in all
   > history** (`min = max = 2026-08-11 09:43:14`). So it is not that the finding is filed
   > somewhere unread — the category has no consumer at all, and the verdict
   > `## Tier-4 acceptance PASSED` was written in the same second. Raises the priority of
   > this candidate: the restored eyes currently write to a channel nothing reads.
2. **243 candidate 2 — the manual/inline path** still loses the vision half by design
   (08-08 ruling keeps the standing chassis bucket-less). (a) accept + document / (b) route
   manual triggers through the spawn path / (c) reopen the ruling. The 08-10b handoff's
   inline-path sibling (chassis `deploy_image_asset` always fails, bugs_open/248 context)
   hangs on the same ruling.
3. **FIRECRAWL_API_KEY** — the one remaining value-copy in the spawn block
   (`spawn_actions.go:2649-2653`), same shape 245 just closed for storage keys. Convert to
   secretKeyRef too, or accept for a SaaS key? Recorded in 245's update.
4. **Batch-8 blocked tail**: 8 loancalculator naming mismatches (their lane's call),
   fuel-budget-forecaster (gaswholesalers logo 404), gas-unit-converter (known-broken page).
   > **CORRECTED 2026-08-11 (parallel session) — "there is no way round it" is FALSE, and
   > this is probably not a naming decision at all.** Every handoff in this chain since
   > 08-10 has said the Tier-4 lookup can only find a page by name because the live agent
   > config has no `url_field`. The first half is right; the conclusion is not. **The code
   > has always supported `url_field`, and it is checked FIRST**
   > (`tool_acceptance_actions.go:163-166`, covered by
   > `tool_acceptance_actions_test.go:377-380`):
   > ```go
   > if uf := datahelpers.GetStringField(config, "url_field", ""); uf != "" {
   >     pageURL = datahelpers.ExtractNestedFieldString(params.CollectedData, uf)
   > }
   > if pageURL == "" && params.DB != nil && siteID != "" { /* the name lookup */ }
   > ```
   > The name lookup is the **fallback**, guarded on `pageURL == ""`. So adding
   > `"url_field": "input_data.spec.page_url"` to the live `request_browser_run` step
   > config (verified absent today; `function_field` already reads `input_data.spec.function`,
   > so `spec.*` is the work item's own spec) is:
   > - **additive and inert** — nothing reaches it until a work item carries
   >   `spec.page_url`; every existing run falls through to today's behaviour unchanged;
   > - **DB config, so live immediately** — no image, no roll;
   > - enough to unblock **all 9** unresolvable placements at once, including
   >   `tool-loan-repayment`, which sits on `index` and therefore **could never be fixed by
   >   renaming anything**.
   >
   > That reframes the owner question. It was "which name is canonical, and who pays the URL
   > change on a finished site?" It is now "do we teach the dispatcher to name the page it
   > means, instead of making 26 live pages match a lookup?" Recommend the latter, through
   > the normal council gate (additive-and-inert ⇒ gate, not RFC, per the 07-29 ruling).
   > Renaming may still be wanted on gamesdesign for its own reason — restoring that site's
   > `tool-` convention, 15 precedents — but that is a tidiness decision, no longer a blocker.

**Routable to other lanes:**

5. ~~**dartsonline setup-builder contrast defect** → fixloop/darts. Possibly an instance of
   the known colour-churn landmine (`generic_theme` misfires; pin via
   `design_intent.palette.reference_values`) — check that first.~~
   > **CORRECTED 2026-08-11 (parallel session) — measured, and both halves of that were
   > wrong.** It is **not** colour churn, and it is **not** one site. NOTES
   > `## 2026-08-11 (parallel session)` has the working; the short form:
   > - **Mechanism**: the component's own rule
   >   `.db-option input:checked + label { background: var(--color-primary); color: var(--color-surface); }`
   >   uses `--color-surface` as its *text-on-primary* colour, i.e. it assumes surface always
   >   contrasts with primary. There is no `.db-option label` base colour rule, which is why
   >   exactly one option per group — the `checked` default — is affected. The palette is not
   >   churned or misfired; **both tokens hold the values the site intends**. The component's
   >   assumption about what they MEAN is what fails.
   > - **dartsonline**: `--color-primary #1A1F2E` on `--color-surface #1E2436` = **1.06:1**
   >   (AA needs 4.5:1; `--color-text` on primary would be 14.65:1).
   > - **The idiom is fleet-wide**: 9 active components / 7 functions, live on 8 domains.
   >   Contrast computed from each site's own served stylesheet: **6 of 8 are fine** — so the
   >   idiom is not wrong, it is *unguarded*. The second casualty is
   >   **mortgagecalculator.co.uk at 2.95:1** (`#b59230` on `#ffffff`), failing AA and even
   >   AA-large, on `tool-bridging-compound` and `tool-rate-scenarios`.
   > - **So the routing changes**: this is not a darts ticket. It is either a component-template
   >   fix (use a token that means "on-primary", or state the contrast requirement) or a
   >   palette-contract check at build time. It touches two lanes' sites, so it is an owner
   >   call on scope — see the owner log entry of the same date.

**This lane's next work:**

6. **`tool-llm-cost-calculator`** — last authorable batch-8 subject; MUST be fork-aware
   (4 forks share the `doc_plans` key). 08-10 ADDENDUM §D has the detail.
7. **`tool-bayesian-ranking`** — needs the RUNBOOK §11 two-row rename first (restores
   gamesdesign's own `tool-` convention; 15 precedents on that site).

## 3b. OWNER DECISIONS LANDED 2026-08-11 (in chat) — and what was done with them same-day

1. **Vision findings visible (243 c3): YES, build it.** The parallel session's measurement
   stands: `render-critique` has NO consumer anywhere in Go, and the PASSED stamp lands the
   same second as the finding. **This is the next build for a fresh session** — spec:
   - Read the judge end-to-end first (`tool_acceptance_actions.go` ~700–1060) — the
     no_auto_fix branch, the improve_tool dedup, and the site-chrome comment at ~773 ("real,
     user-visible defects that no tool edit can fix") are the constraints to respect.
   - A vision result must (a) mark the run's SUMMARY distinctly — `vision: ok /
     finding(<gist>) / skipped(<reason>)` — so a green-with-eyes is distinguishable from a
     green-without; and (b) on a finding, raise something a reader will meet: given the
     ~773 comment, a vision finding is usually a SITE/COMPONENT defect, so it should follow
     the site-chrome path (visible item, not tool-improver), never weaken no_auto_fix.
   - Platform code → council gate; if a new note category or consumer seam is added,
     concept-register entry in the same commit and name the consumers (07-29 ruling).
2. **Manual path (243 c2): option (b), an orchestrator wrapper. DONE.**
   `tool_acceptance_run.sh` rewritten to insert the due-sweep's own work item so manual
   runs SPAWN (both halves run); preflights refuse the old quiet no-ops loudly. RUNBOOK
   §10 box has the mechanics. Proof run: work item `4ef3c11a…` (see NOTES for the result).
3. **FIRECRAWL_API_KEY: convert. DONE** — moved into `agentenv.providerKeyNames`
   (secretKeyRef, both spawners; also fixes the remote-spawner never injecting it — the
   112 drift class). Commit `f56abaadf`, `Council-Submitted: 6f13c5ce…` — **read that
   verdict** (session-start item). Inert until the next roll; verify by pod-spec capture.
4. **loancalculator / dartsonline-contrast / gaswholesalers-logo: ALL THREE DECIDED
   (owner, 2026-08-11, in chat) and the lanes notified same-day** (four NOTIFY files:
   `loancalculator_couk/`, `fixloop_eg_dartsonline/`, `mortgagecalculator_couk_adoption/`,
   `js_snippets_news_gaswholesalers/`). The decisions, and the work they sanction:
   - **loancalculator → url_field route, NO renames.** New sanctioned work for THIS lane:
     add `"url_field": "input_data.spec.page_url"` to the live `request_browser_run` step
     config (DB, live immediately, through the council gate) AND make the spec producers
     carry `page_url` (`check_tool_acceptance_due.go` — Go, rides a roll; and the
     `tool_acceptance_run.sh` wrapper — script). Unblocks all 8 + the `index` case.
   - **contrast → BOTH: fix the shared component template (proper on-primary pairing) +
     a build-time palette-contract check.** New sanctioned work for this lane; rerender
     the affected pages (dartsonline setup-builder; mortgagecalculator bridging-compound
     + rate-scenarios) when the template fix lands. Both site lanes asked NOT to
     hand-patch CSS meanwhile.
   - **gaswholesalers → their lane deploys the logo properly** (NOT via the 248-broken
     auto-repair); fuel-budget-forecaster's S6 unblocks when the logo serves 200.

## 3c. AND THE REST LANDED TOO — 2026-08-11 afternoon (the session the owner decided in)

All five §3/§3b items now have their answer; full working NOTES `## 2026-08-11 (parallel
session)` (both blocks) and the bug files. State only:

1. **Vision findings (243 c3): BUILT** — `record_vision_finding` + `vision_finding` item
   type (TL-041), commit `e6d1ac6dc`, `Council-Submitted: 310dee45…` (**read that verdict**,
   session-start item). Inert until the next chassis roll; then apply
   `383_tool_acceptance_vision_findings_visible_HOLD.sql` BY HAND after pod-grepping
   `record_vision_finding` ≥1 on both replicas, and `--record-only` it. Proof owed once
   both halves live: a `FINDINGS: reported` critique → exactly one `vision_finding` row;
   a `FINDINGS: none` critique → none.
2. **Contrast (owner: fix the shared component): FIXED + LIVE on 8 of 9 pages** —
   migration `382` (9 templates → the guaranteed `--color-text` fill; per-site proof
   10.35:1–17.85:1), rerenders verified at the artefact. The 9th (gaswholesalers
   fuel-cost-estimator) is `rebuild_policy='owned'` and `save_page_sections` REFUSED the
   generic save by design — left refused: that site was legible all along (17.06:1), and
   the fixed template rides the next tool-pipeline rebuild of that page.
   **Two discoveries riding this fix:** the estate already HAS an on-primary token
   (`--color-primary-text`, all 8 sites) — future harmonisation candidate; and
   mortgagecalculator's own `--color-primary-text` on `--color-primary` is **2.95:1**, a
   palette-level defect routed to their lane
   (`mortgagecalculator_couk_adoption/CONTRIB_2026-08-11_…_contrast.md`, with rollback
   pointers for the two pages this fix re-rendered on their site).
3. **loancalculator (owner: probably both): `url_field` is LIVE** — migration `384`,
   verified at the live row, inert until a work item carries `spec.page_url` (structurally
   safe: absent field → "" → the name-lookup fallback unchanged). Producer half
   deliberately deferred WITH the reason (the due-sweep only raises items for
   PLAN-carrying tools, which all resolve by name); the wrapper's optional `page_url`
   argument belongs to the wrapper's owner (RUNBOOK §10). Renames: gamesdesign's
   `tool-bayesian-ranking` §11 rename is still batch-8 work on its own merits (site
   convention, 15 precedents); the loancalculator renames are now optional tidiness for
   that lane, NOT a blocker — `tool-loan-repayment` on `index` was never renameable anyway.
4. **gaswholesalers logo-404**: not this session's; unchanged.

## 3d. POST-ROLL, 2026-08-11 ~13:00Z — v1.0.1286: everything in §3c that was waiting on a roll has now LANDED

Fleet is **v1.0.1286** (chassis + browser-runner, pods up 12:02–12:03Z) — §1's 1284 is
superseded. Full working: NOTES `## 2026-08-11 (parallel session, afternoon-2)`.

1. **243 c3 is FULLY LIVE AND ITS NEGATIVE ARM IS PROVEN.** Pod-grep 6/0 both replicas →
   migration **383 applied by hand + recorded** (renamed from `_HOLD` post-apply: the
   runner refuses to record uppercase sidecars. ⚠ the number 383 collided with another
   session's `383_rfc022_…` file — the ledger is filename-keyed, both stand; resolve by
   filename like bug numbers). Live row verified all three ways. Proof run: wrapper item
   `3bec5e4f…` → spawned run `c3139293…` `complete`, verdict green 15/0, the model wrote
   the machine line (`FINDINGS: none`), `file_vision_finding` returned
   `{filed:false, verdict_line:"none"}`, **0 `vision_finding` rows, 0 collateral items**.
   The positive arm stays unit-pinned until a genuine finding occurs.
   **Bonus, independent-instrument confirmation of 382:** the critique explicitly reports
   "no … contrast failures" on the very page that measured 1.06:1 the day before — the
   vision pass is describing the NEW chips and finding them legible.
2. **Council 310dee45 came back REVISE round 1 — answered, resubmitted round 2 on the SAME
   correlation** (run `73cb0a29`, mid-review at cut time; **read the round-2 verdict**,
   session-start item). The two objections that improved the work: the medium (a failed
   filing left no durable trace) is FIXED in code — commit `3ed587049`, insert failure now
   writes a render-critique doc_note with error + full critique, pinned by
   `TestVisionInsertFailureLeavesDurableNote`; the high (needs_human_review is historically
   a silent sink, 033/083) is answered by measurement — the admin dashboard HAS a
   'Needs Review' approve/edit surface and 033's row-cap display bug is fixed there; the
   remaining gap is CADENCE, which is 033's remit and shared by every producer of that
   status. The round-1 overclaim "exactly one reader" is withdrawn for "one queue, deduped,
   displayed, closable — cadence tracked in 033".
3. **245 is DONE IN FACT, residual included**: `env | grep -c '^B2_|^AWS_ACCESS|^AWS_SECRET'`
   → **0 on both 1286 replicas**. Recorded in the bug file; stays in `bugs_open/` per owner
   practice. (Firecrawl corr `6f13c5ce`: **APPROVED** on its second round — verdict read.)
4. **Unrelated but adjacent — migration 389** (another session, owner decision): the weekly
   render audit's 220 `contrast_failure` items are PARKED so improvement-sweep can drain
   page re-renders. Different pipeline from 382's template fix; no interaction. Expect the
   dispatch rotation to be busier than usual while the re-renders drain.

**What is actually left on this lane's plate:** the round-2 verdict (`73cb0a29`); the
`vision_finding` positive arm (arrives with the first genuine finding — check
`SELECT * FROM site_work_items WHERE item_type='vision_finding'` after any acceptance run
whose critique reports); batch-8 tail unchanged (`tool-llm-cost-calculator` fork-aware,
`tool-bayesian-ranking` after its §11 rename, loancalculator once their lane authors
golden-derived PLANs — 384 is live and waiting).

## 3e. THE WRAPPER'S `page_url` HALF — done by the wrapper's owner, 2026-08-11 ~14:00Z

§3c item 3 deferred this piece "to the wrapper's owner"; that is the session that rewrote
`tool_acceptance_run.sh` this morning, and it is now done — **not** as an optional argument.
The wrapper resolves the page **by component placement** and always writes `spec.page_url`,
so `url_field` (384) actually has a producer on the manual path. Exact-name pages still win
the tie, so nothing changes for tools that already resolved. Foreground-tested:
`tool-loan-repayment` — the `index` case no rename could ever fix — now resolves and prints
the route note, stopping honestly at its missing PLAN. New refusal added for an empty
`pages.url` (the one case neither route can resolve). Commit `585e37dad`.

**The naming gate is therefore gone from BOTH halves of the manual path; what blocks the
eight loancalculator tools is a fence to author, not a name.** The sweep producer is still
deliberately unchanged (its population all resolves by name — §3c's reasoning stands).

Proof status, stated honestly: run `a457a96a…` exercises the new wrapper end-to-end and
shows the extra spec key breaks nothing. It **cannot** prove the url route was TAKEN — for
tool-setup-builder both routes give the same URL. **The discriminating test is the first
fence authored for a loancalculator tool**, whose page name cannot resolve at all; whoever
does that gets the route proof for free.

⚠ **New LANDMINE from verifying 384** (`LANDMINES.md`, 08-11): a step's NAME is not its
ACTION. `steps->'request_browser_run'` is NULL — the step is keyed **`request_run`** (and the
vision step `look`). Path-read, `? 'key'` and `jsonb_pretty` all return NULL together, which
reads exactly like "the migration never landed". Enumerate `jsonb_object_keys(…->'steps')`
FIRST.

## 4. Standing defect list

Items 1–8 unchanged from `HANDOFF_2026-08-09` §4. Item 9 (243) → **all three candidates
done** (c1 proven; c2 wrapper live; c3 live + negative arm proven, round-2 verdict pending).
Item 10 (batch-8 naming gate) → dissolved by 384 (§3c item 3); gamesdesign rename still
wanted on its own merits. 245 → **done in fact, residual proven** (§3d item 3). 248 (both
files) — read before touching asset deploys.

## 5. Session-start checklist

1. `git log --oneline -10`; re-read this file FROM DISK. The tree carries the owner's
   uncommitted release bump (§1) and possibly other sessions' WIP.
2. Pod-grep chassis + browser-runner (RUNBOOK §4 markers). No dispatch within 300s of a
   restart.
3. **Read the round-2 council verdict** — find by payload:
   `collected_data->'input_data'->>'fix_correlation_id' = '310dee45-…'` (or the run id
   `73cb0a29…`). REVISE → objections come answered; act, resubmit on the same correlation.
4. Re-run the census + `CHECK_naming_contract.sh` before quoting any batch-8 figure.
5. `who-owns.py` + live-transcript grep before writing at robot-hands, loancalculator,
   gamesdesign, or anything touching 248's deploy surface.
6. **AND grep the LANE's own recent transcripts, not just `git log`** — twice on 08-11 a
   session was minutes from re-implementing work another session had in flight, once with
   the owner's decision in hand. The recipe that worked:
   `CUT=$(date -u -d '30 minutes ago' +%Y-%m-%dT%H:%M:%SZ); find ~/.claude/projects/-home-ant-projects-agentchassis/ -name '*.jsonl' -newermt "$CUT" | xargs grep -lc '<the symbol you are about to touch>'`
   **A sanctioned task is not a claimed task**: the owner decides in ONE chat, the lane is
   worked by several. Claim in this file before the first edit.
