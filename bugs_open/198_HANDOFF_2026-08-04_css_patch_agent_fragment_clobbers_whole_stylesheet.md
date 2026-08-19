# 198 — css-patch-agent persists an LLM fragment as the WHOLE stylesheet: theme + repo + live site clobbered on its maiden run

Filed 2026-08-04 by the vigilant_designer_offer_analysis lane, from the A0.4/A0.3b
witnessed runs. **LIVE INCIDENT while open: relojistas.com served a 149-byte
styles.css** (whole-site unstyled) from ~10:00Z until the restore lands.

**090 substitution stated plainly (owner ruling 2026-07-31):** not filed through the
diagnosis loop because the entire causal chain is first-hand, artefact-verified in one
session with no inference: the live workflow config (read from `agent_definitions`), the
four work-item results carrying the fragment, the `css_themes` row shrinking 25,816 → 112
→ 149 chars across versions 2–5 with timestamps matching the four completions, the four
`CSS fix: <no value>` commits at vm-sites, and the live 149-byte serve. Nothing here is
a hypothesis; a 090 run would re-read the same six artefacts.

## Mechanism (all live config, quoted)

`css-patch-agent` workflow: `load_current_css` (css_themes via style_collections) →
`plan_css_fix` (LLM, prompt demands *"patched_css: the complete updated stylesheet with
the fix applied"*) → `save_css_to_db` (`UPDATE css_themes SET css_content = $2` with
`css_fix.result.patched_css`) → `deploy_css` (`git_commit` of `assets/css/styles.css`
with `content_field: css_fix.result.patched_css`) → complete.

**The model returned ONLY the new rule as `patched_css`**
(`"SPAN.ag-eyebrow { color: #7a6010; }"`, its `changes_summary` honest about being a
one-rule change) — and NOTHING between the LLM and the two writers checks size or shape.
The 25,816-char theme became 112 chars in the DB and 149 chars at vm-sites HEAD
(4 commits, 09:56–09:58Z), and the live host synced the fragment (~10:00Z). Three
follow-on runs each saw the already-clobbered "stylesheet" as `current_css` and appended
their own rule — so the final artefact is exactly the four fix rules and nothing else.

This is `bugs_open/012`'s class ("any agent that rewrites a whole artifact can persist a
fragment and report success") arriving through prompt non-compliance instead of
truncation — and the guard the platform already learned to build (the shrink guard) is
absent on BOTH writers. Note `max_tokens: 8000` also makes the "complete stylesheet"
contract structurally unfulfillable for any real theme (~26KB > 8000 tokens' worth in
practice for larger sites): even a compliant model would truncate on a big stylesheet,
hitting the same missing guard via the 012 route. The whole-document-rewrite CONTRACT is
the defect; the model's fragment merely exposed it on day one.

**Secondary defect, same run:** `deploy_css`'s commit_message template renders
`CSS fix: <no value> — <no value>` — `{{.input_data.spec.category}}` resolves in
`plan_css_fix`'s prompt but NOT in `git_commit`'s template context (different context
assembly; missingkey behaviour differs). Cosmetic but it destroyed the audit trail of
what each commit was.

## Impact + containment (all done 2026-08-04, ~10:00–10:30Z)

- All 4 contrast items had already completed before detection — no cancels possible;
  nothing else routes to css-patch-agent today (these were its first-ever items).
- **css_themes RESTORED: version 6, 26,152 chars** = the pre-clobber content (vm-sites
  commit `ee123e31a`, the 09:45 webdesign-agent state — one char newer than live) + the
  four contrast rules appended under a provenance comment. The DB is the durable
  restoration source now.
- **vm-sites HEAD RESTORED by the owner 21:48Z** (commit `f8f2dac`, 26,335 bytes — the
  owner ran the handed-over gh PUT after the session's own outward channels, a gh
  contents PUT and a hand-published `system.adapter.git.requests` message, were blocked
  by the harness permission classifier). **Live site recovers on the next scheduled
  pull** (the vm-sites deploy is a periodic pull, ~≤1.5h lag per the platform's own
  tooling notes); **live VERIFIED RECOVERED 22:13Z 2026-08-04** (26,335 bytes served, contrast-fix block present). The INCIDENT is closed at every layer (DB v6 / repo f8f2dac / live); the DEFECT — the whole-document round-trip with no shrink guard — remains open and is what this file is about.
  The four fixes themselves are correct and preserved in the restore.

## Fix candidates, ordered by what closes the door (the platform's own ranking rule)

1. **Make the fragment unrepresentable: stop round-tripping the whole document through
   the model.** Have `plan_css_fix` return ONLY `css_added` (the new/changed rules) and
   make the writer APPEND/apply it to the loaded stylesheet server-side
   (`current_css.css_content + css_added`). The model can then never destroy what it
   never carried. Also drops the 8000-token structural conflict entirely.
2. **Shrink guard on both writers regardless** (defence in depth, the 012 lesson):
   refuse `save_css_to_db`/`deploy_css` when the new content is < some fraction (say
   50%) of the loaded `current_css` unless an explicit `allow_shrink` is set — the
   `mistyped-deployed-page`/012 family's proven shape.
3. Fix the commit-message template context (or build the message in Go from spec fields
   already forwarded) so the audit trail survives.
4. NOT a candidate: "prompt harder". The prompt already demanded the complete document
   and the flagship model ignored it on the first real input; a comment is not a control
   (a-doc-comment-is-not-an-enforcement-mechanism), and neither is a prompt.

## How to verify a fix

Re-run the A0.3b → sweep chain on the specimen (render audit files contrast items →
sweep promotes → css-patch runs): css_themes length must be ≥ its pre-run length, the
vm-sites diff must touch only the intended rules, and the NEXT render audit must stop
re-filing the fixed pairings (the de-facto verifier designed in A0.3).

## Related

- `bugs_open/012` (the class), 016b §9 "output_tokens == max_tokens means CUT".
- e49f5935 r2's improvement_guardian advisory: "css-patch-agent has never received a
  work item — if it is misconfigured the fix loop stalls silently." It did not stall;
  it fired and the maiden run found this. The witnessed-run discipline is what caught it
  same-hour instead of fleet-wide later.
- The A0.3 r2 spec-contract work is NOT implicated: the prompt received a real finding
  and produced a correct minimal fix — the defect is entirely on the persist side.

## Fix applied 2026-08-05 (candidates 1 + 3; session dispatched at this bug by the owner)

**Candidate 1 is LIVE** (migration `docs/agent_docs/sql_for_agents/318_…`, applied +
recorded 2026-08-05, snapshot `661a27b9` taken first; commit `c48c773c1`,
`Council-Submitted: 5249320e-1d6e-4bb4-94d3-89bd802bd8a4`):

- `plan_css_fix` returns ONLY `css_added` — the model never carries the stylesheet
  again, which also dissolves the max_tokens 8000 structural conflict named above.
- `save_css_to_db` APPENDS server-side (`css_content = css_content || …` with a dated
  provenance comment naming the category). SQL concatenation is monotonic, so the DB
  writer *cannot* shrink — candidate 2's guard delivered by construction, not threshold.
  A 1..8192-char size guard refuses a whole-document echo by matching zero rows.
- NEW `check_saved` step: a refused append routes to `complete_error` LOUDLY — without
  it, zero rows would ride `git_commit`'s "no files → skipped, Success:true" path and
  read as complete.
- `deploy_css` commits `css_saved.css_content` — the DB row AS RETURNED by the UPDATE —
  so repo and DB cannot diverge through this workflow.
- Live-verified post-apply: all five config paths probe correct and `patched_css`
  appears NOWHERE in the config (`position(...)` = 0). The exact installed query was
  proven on the real relojistas theme row inside a rolled-back transaction:
  26,152 → 26,240 chars, v6 → v7, `commit_msg` = "CSS fix: contrast_failure (theme v7)";
  the 9,000-char oversize probe matched **0 rows**. Live row untouched (v6 preserved).

**Candidate 3 is COMMITTED, INERT UNTIL THE NEXT CHASSIS ROLL:** `commit_message_field`
(opt-in, default OFF) on `GitCommitAction` — the `<no value>` mechanism is that
`buildCommitMessage` executes the template against a fixed `{domain, file_count,
filename}` map, never CollectedData. The message now composes in the save step's
RETURNING (`… AS commit_msg`), where params actually resolve. Four tests, field-wins
proven by mutation. Until the roll, the old binary ignores the key and falls back to
the also-updated template `CSS patch: {{.filename}} ({{.domain}})` — honest, less
specific. Both orders safe; registered as DGH-007 same commit; LANDMINES entry added
for the template-context trap.

**Gemini note (measured, not assumed):** this chain does not touch Gemini —
css-patch-agent's 4 incident calls were anthropic/claude-sonnet-4-6 (`llm_call_log`),
and render-audit-agent has no gemini rows; the fleet's gemini traffic is
page-content-writer's generate_content steps. A Gemini outage neither blocks this fix
nor its verification chain.

**STILL OPEN because the file's own verify bar is not yet met:** the witnessed
end-to-end run (render audit files a NEW contrast item → sweep promotes → css-patch
appends → next audit stops re-filing it) belongs to the next real finding — the four
incident pairings are already fixed in v6, so the next audit should file nothing for
them (that non-refiling is itself the verifier of the original fixes). Dispatch
cadence is the vigilant_designer lane's; the door is closed ahead of it, which is what
their handoff ordered ("198's fix candidate 1 before any css-patch dispatch"). Closure
call is theirs once the witnessed run lands; the commit-message half additionally
wants the chassis roll.

## Council round: APPROVED r1 (2026-08-05, correlation 5249320e), 7 advisory objections — dispositions

Verdict read same-day; `complete_approved` at 12:08Z. Every objection either answered
with a measurement or recorded as owed — none required a code change:

- **"Migration premise unverified against live jsonb" (editquality, prior_art_librarian):**
  the full live config was dumped and read before writing; the concern was still REAL —
  the row's updated_at moved at 09:09Z today (content byte-identical, 3,933 chars, so a
  touch not a change) — and the drift guard did its job: matched the exact expected
  shape, UPDATE 1, DO/RAISE verify passed, post-apply probes green.
- **"Other commit_message templates still silently degrade" (bug_historian):** measured —
  19 `commit_message` templates fleet-wide, **zero** reference fields outside
  {domain, file_count, filename} today (survey proven non-vacuous by the 19 count).
  css-patch-agent was the only offender. Whether `buildCommitMessage` itself should
  fail loud on `<no value>` stays an open human trade-off, recorded in DGH-007.
- **"commit_message_field naming collision?" (guardian):** the fleet-wide key search
  returns exactly two users — mine, and `diagnose_prepare_fix_commit` (feature-implementer,
  `stage.commit_message`), which is PRIOR ART for the same name with the same semantics
  (answers reuse_agent/prior_art_librarian too: the convention was followed, not invented).
- **"Truncated css_added passes the size guard" (llm_reliability):** the deciding arm is
  `ai_actions.go:427` — `aiservice.IsTruncated(err)` surfaces a max_tokens cut as an
  ERROR, and plan_css_fix sets no `tolerate_truncation`, so a truncated completion fails
  the step loudly before any writer runs; a tolerated fragment would additionally have
  to survive the `output_format: json` parse + single re-ask. Two closed doors.
- **"snapshot_agent overload/table unverified" (debug_historian):** verified —
  `agent_definitions_backup` id `661a27b9` holds the pre-change config (old
  `patched_css` shape present, `t`). Verify SQL is committed in the migration itself
  (DO/RAISE + `position()` containment over the whole config, not `<>` on a path).
- **"No doc_notes trail for the agent" (tooling_provenance):** written —
  `doc_notes` `930cc916`, subject pipeline/css-patch-agent, per the tool-recreation-handler
  convention.
- **Nil `input_data.spec.category` (editquality, low):** param resolution refuses nil →
  step errors → complete_error. Loud failure through a different door; accepted.

**OWED (named, not half-done):**
1. **Round-trip writer inventory** (architecture seat): enumerate other
   agent_definitions workflows that round-trip a whole artefact through an LLM into an
   unguarded writer — the 012/198 class, quantified. Nobody has this list.
2. **Pod-grep when the chassis rolls** (debug_historian): 
   `kubectl exec <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c resolveCommitMessage'`
   (expect ≥1, every replica) before trusting the commit_msg leg; the config half is
   already live and falls back cleanly meanwhile.

## 2026-08-10 — chassis v1.0.1277 rolled: candidate 3 is now LIVE at the pod

The owed pod-grep (debug_historian's objection) is done: both running chassis replicas
(`agent-chassis-6dc54d77cd-lftkt` / `-v2b59`, image `v1.0.1277`, started 2026-08-09
21:35Z) — `grep -ac resolveCommitMessage /app/agent-chassis` = **2** on each, negative
control (`zz_no_such_symbol_198`) = **0** on each, same exec. Spawned pods run their
spawner's image (bugs_open/066 fix), so css-patch-agent's next dispatch inherits the
binary. Config half re-probed same day: `commit_message_field=css_saved.commit_msg`,
`params[1]=css_fix.result.css_added` — intact. **Both fix candidates are now fully
live. The ONLY thing keeping this file open is the witnessed end-to-end run** (next
real contrast finding → promote → append → next audit stops re-filing), which is the
vigilant_designer lane's to dispatch. DGH-007's register status updated to deployed.

---

## SECOND LIVE INCIDENT — 2026-08-18, noted.co.uk: candidate 3 held, and the OTHER named door was walked through (appended 2026-08-19, noted-rebuild lane)

**The fix worked; the site still lost its stylesheet.** This file's own words from
08-04 — *"the guard the platform already learned to build (the shrink guard) is
absent on BOTH writers"* — candidate 3 (live v1.0.1277, verified 08-10) made
shrink unrepresentable on the **DB** writer. The **git** writer (`deploy_css` →
`git_commit` of `css_saved.css_content`, the whole DB value) still has no guard,
and this incident went through it.

**The new seed — a theme row BORN EMPTY.** `theme-noted-co-uk`
(`07f5cc32…`, `origin='adopted'`, `forked_from=NULL`) was created 2026-08-15
18:17:32, **67 seconds before** webdesign-agent committed the real 17,475-byte
stylesheet to vm-sites (`abe1b617a7`, 18:18:39). The DB row held **empty
css_content** from birth; git held the real file; nothing compares them.
`[INFERRED at the one unwitnessed link: that `fork_theme_from_site` inserted
`renderedCSS=''` in that run — the insert carries rendered content by code, so
either the render was empty or a different path made the row. Verifying step for
the fixing thread: instrument/read the 08-15 run's `fork_theme` inputs, or induce
a fork with empty spec and watch the insert.]`

**Then the guarded append did exactly its job, onto the wrong baseline.** 21
contrast findings on 08-18 (11:36–11:52Z) → 21 correct, guarded appends (v2:
**91 chars**, growing to v22: 2,381) → 21 `deploy_css` commits each replacing the
17,475-byte file with the accumulation. First commit `7a5d4fc0a1` 11:36Z is the
kill; the file's own history shows 17,475 → 91 → … → 2,381. **The DB guard
cannot see the FILE, and the deploy trusts the DB as the one truth.**

**Amplification, worth its own line:** most of the 21 findings were filed by the
contrast auditor looking at a page that had already lost its stylesheet — the
loop was patching damage it had itself deployed, quickly (21 versions in 16
minutes), every run reporting success.

**Detection gap, restated for the fix:** (1) no shrink/size-delta guard on the
git writer — the door this file named and nobody closed; (2) no birth guard —
`fork_theme_from_site` will insert empty `css_content` silently; (3) no
DB↔deployed-file drift check — noted ran 3 days split (git full / DB empty) with
every page green.

**Repair (noted, DONE 2026-08-19, all layers verified):** css_themes v23 =
17,475-byte base (git `abe1b617a7`) + provenance comment + all 21 patches
(20,190 chars); deployed via the git-adapter (`repo_name: "vm-sites"` — NOT the
default) → repo 20,367 bytes → box → live 200/20,367, 98 base-selector lines, 41
patch markers. DB and file now identical, so the next patch cycle deploys a true
whole.

**Fix candidates, ordered by what closes the door (this lane's addition):**
1. **Deploy-side shrink guard** (the one 198 already named): `deploy_css` — or
   `git_commit` generally, opt-in per config — refuses a payload smaller than
   N% of the file it replaces without an explicit override. Closes BOTH known
   routes (LLM fragment AND wrong-baseline) at the last writer.
2. **Birth guard**: `fork_theme_from_site` refuses or `needs_review=true`s an
   insert whose `css_content` is empty/tiny.
3. **Drift check** (detection): a discovery check comparing `css_themes` length
   vs the deployed `styles.css` size per site.
Candidates 1–2 are platform Go → council gate; the noted lane has NOT built them
(ownership respected — this file's lane holds the defect).

## 2026-08-19 — ROUND 2 (idea.uk lane): the fix held, its UNTESTED ARM fired — six more sites clobbered through the guarded door

**090 substitution stated plainly (owner ruling 2026-07-31):** first-hand,
artefact-verified end to end in one session, same shape as this file's original filing —
live workflow config read from `agent_definitions`, vm-sites commit history, `css_themes`
rows, the driving work items, and the served bytes. A 090 would re-read the same
artefacts.

### Mechanism — the append fix assumed `css_themes` holds the stylesheet; on ~11 sites it holds NOTHING

The candidate-1 fix is working exactly as built: the model returns only `css_added`, the
DB append is monotonic, `deploy_css` ships the row as returned — "so repo and DB cannot
diverge through this workflow" (the 08-05 section above). What the 08-05 work did not
test is the **empty base**: `assets/css/styles.css` is written ONLY by a
webdesign-agent design run and was **never sourced from `css_themes`**
(`bugs_open/072`'s finding, quoted at `rerender_single_page_action.go:396`). Fleet
measurement 2026-08-19: **11 of 21 linked themes have `css_content` of 0 bytes** while
their sites serve 13–25KB stylesheets carrying every `:root` variable definition.
relojistas — the one site this file's restore backfilled — is the only site where
"DB = the file" is true.

So on any of those sites, the guarded chain does this: append 100–400 bytes to `''` →
deploy the result **wholesale** over the 20KB file. The DB writer cannot shrink; the
FILE shrinks ~98%. Convergence in the wrong direction. The 08-05 in-transaction proof
ran on the real relojistas row — the populated-base arm only.

### What fired it, and when

`detected-item-promoter` (bugs_open/083's fix, live 2026-08-16/17) began routinely
promoting render-audit `contrast_failure` findings, so css-patch-agent started receiving
real dispatch waves for the first time. Verified clobbers, all at the artefact
(served `/assets/css/styles.css`, 2026-08-19, `:root` count 0 on every one):

| site | served styles.css now | pre-clobber | clobber evidence |
|---|---|---|---|
| **idea.uk** | 428 B | **23,650 B** (vm-sites `8c407a18f`, 08-09) | 4 commits `CSS fix: contrast (theme v2..v5)` 08-17 21:40–21:41Z; **RESTORED, see below** |
| dartsonline.com | 164 B | unmeasured | theme row 164 B, patched 08-18 04:39 |
| vonc.com | 176 B | unmeasured | theme row 176 B, patched 08-18 09:33 |
| cookly.uk | 504 B | unmeasured | theme row 948 B, patched 08-17 17:21 |
| noted.co.uk | 2,381 B | unmeasured | theme row 2,381 B, patched 08-18 11:41–11:52 (14+ items) |
| oufe.com | 1,336 B | unmeasured | theme row 1,336 B, patched 08-18 10:40 |

Reader-visible effect: every `--color-*` definition vanishes. Sections whose templates
carry dark fallbacks (`background: var(--color-background, #0d0d0d)`) go dark while
their text colour vars fail to `inherit` (black) → black-on-black; heroes hard-code
`--hero-ink:#fff` over a background that failed to transparent → white-on-white. On
idea.uk this is **most of the text on every page**.

### It is SELF-AMPLIFYING, and completions lie twice

1. The render audit measures the post-clobber page, finds **1.00:1** pairings (that is
   invisible text), files `contrast_failure` → promoter → **css-patch-agent again** —
   which cannot restore definitions it has never seen (its prompt correctly forbids
   inventing `var()` names, so it emits literal-colour selector patches) and re-deploys
   the still-themeless file. loancash.co.uk took **11 items in 8 minutes**
   (08-18 16:34–16:42Z), every one `complete`.
2. loancash has **no linked style_collection at all**, so those 11 runs exited via
   `complete_no_css` ("cannot patch") — and the items still read `complete`. A
   completion from this handler currently proves nothing about the artefact in either
   arm (`a-complete-work-item-is-not-a-repaired-artefact`).

Note each spec already carries an `acceptance_test` field naming the exact single-
selector re-measurement — written by the audit, read by nothing in this workflow.

### Why the framework did not catch it

- The workflow's only post-write check is `css_saved.count >= 1` — the DB append took.
  Nothing compares against the artefact being replaced.
- The git leg accepted a 23,650 → 428 byte replacement of a site-wide asset without
  comment; there is no shrink guard at the SECOND writer (this file's original
  candidate 2 was delivered "by construction" on the DB side only).
- The audit DOES re-detect — but routes the finding to the agent that caused it, and
  per-site audit eligibility (hourly rotation, LIMIT 1 site, 7-day window) means a
  clobbered site can serve invisible text for days before it is even measured: idea.uk
  was clobbered 08-17 21:41 and never re-audited before the owner saw it 08-19.

### idea.uk RESTORED 2026-08-19 (this file's own relojistas recipe)

- `css_themes` `4734d51c` → **v6, md5 `4841523e47aec4e181fc976aaedd1ae6`**, content =
  vm-sites `8c407a18f` base (23,650 B) + provenance comment + the four legitimate 08-17
  patch rules. Guarded UPDATE (matched md5 `0e8637f5…` of the 428 B state), byte-identity
  verified local-vs-DB by md5. The DB is now the durable restoration source for idea.uk,
  as it already is for relojistas.
- **Deploy leg is framework-native and in flight:** deferred item `01a4dbca` (a real
  08-10 finding; its `parked_reason` condition "restore to detected when 213 is fixed"
  was met 08-15) restored to `detected` as a single-item canary. Promoter → dispatch →
  css-patch run will append its fix and ship **the full restored row** to vm-sites;
  the live host pulls on its ~1.5 h cycle. Verify at the artefact:
  `curl -s https://idea.uk/assets/css/styles.css | grep -c ':root'` → expect 3.
- **The other five sites are OTHER LANES' — same recipe, one run each:** recover the
  pre-clobber file from the site repo's history (the commit before the first
  `CSS fix:` commit on `<domain>/assets/css/styles.css`), append the patch rules the
  clobber-era commits added, write it into the site's `css_themes` row md5-guarded,
  then let (or make) one real css-patch item run. NOTE `sites.github_repo` is empty for
  dartsonline/vonc/cookly/oufe — resolve the repo the way the git-adapter does, not by
  assuming vm-sites.

### Fix candidates for the CLASS, ordered by what closes the door

1. **Make the divergence unrepresentable: backfill `css_themes` from every site's live
   deployed `styles.css` wherever file > row** (the relojistas restore, applied
   fleet-wide, one-time). After that, "deploy the DB row wholesale" is actually safe,
   which is the assumption the 08-05 fix already ships under.
2. **Shrink guard at the second writer**: refuse (or require `allow_shrink`) a deploy
   that replaces `assets/css/styles.css` with content < ~50% of the repo-HEAD file.
   Lives naturally in the git-adapter (it can see the file it replaces) — platform
   seam, council round.
3. **Use the spec's own `acceptance_test`**: a post-deploy step re-measures the one
   pairing at the served page; the item completes only on measured improvement. Also
   closes the loancash arm (a `complete_no_css` exit must not mint `complete`).
4. **Route the mass-1.00:1 signature away from css-patch-agent**: N same-run 1.00:1
   findings is a stylesheet-integrity symptom, not N selector defects.

Candidates 2–4 are the owning lane's to sequence; candidate 1 plus the five per-site
restores can be run by any lane holding this file's recipe.
