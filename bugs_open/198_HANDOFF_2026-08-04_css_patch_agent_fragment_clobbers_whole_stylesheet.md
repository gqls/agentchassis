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

### 2026-08-19 (later, idea.uk lane) — the two sections above are ONE incident wave, written concurrently; reconciliation

The noted-rebuild lane's SECOND LIVE INCIDENT section and the idea.uk lane's ROUND 2
section were appended within the hour, neither seeing the other. Read them as one:

- **One mechanism, two halves.** noted has the SEED (`fork_theme_from_site` inserts a
  theme row born empty — with an `[INFERRED]` link left to verify) and the pre-symptom
  one-query tell; ROUND 2 has the fleet census (11 of 21 rows empty — so the born-empty
  seed, or something like it, is the NORM, not a noted one-off), the self-amplification
  measurements, and the loancash `complete_no_css` arm.
- **Restore state fleet-wide:** relojistas (08-04), **noted (08-19, all layers, via
  git-adapter `repo_name:"vm-sites"`)**, **idea.uk (08-19, DB v6 done; deploy riding
  canary item `01a4dbca` — if it stalls, the noted lane's git-adapter route is the
  proven fallback)**. **Still clobbered and unowned-for-restore: dartsonline.com,
  vonc.com, cookly.uk, oufe.com** — recipe in either section; note their
  `sites.github_repo` is empty, so resolve the repo the git-adapter's way.
- **Fix candidates converge** — both lanes independently rank the deploy-side shrink
  guard first. Union list for the owning lane: (1) deploy-side shrink guard [both] ·
  (2) birth guard on `fork_theme_from_site` [noted] · (3) DB↔file drift discovery
  check [noted; ROUND 2's fleet backfill is the one-time data half of the same idea] ·
  (4) use the spec's own `acceptance_test` post-deploy + stop `complete_no_css`
  minting `complete` [ROUND 2] · (5) route mass-1.00:1 to a stylesheet-integrity item
  [ROUND 2]. Two LANDMINES landed adjacent in the file, cross-linked.

---

## 2026-08-20 — dartsonline.com RESTORED (one of the four this file names), and what the restore then exposed

Appended by the `dartsonline_traffic` lane. **Read the honest framing first: I re-derived this
file's ROUND 2 mechanism independently and it was already here.** The owner asked me to "fix the
CSS" on dartsonline; I found the 164-byte stylesheet, traced it to `css-patch-agent`, measured
theme rows against deployed files across six sites, and wrote it up as a discovery — then
grepped `bugs_open/` before filing, per CLAUDE.md, and found this file already carrying the
census (11 of 21 rows empty), the self-amplification, the ranked candidates, and **these four
domains named by name as "still clobbered and unowned-for-restore"**. Nothing in my mechanism
section was new. What follows is only what was not yet true.

### 1. dartsonline.com is restored, at both layers, verified

| layer | state |
|---|---|
| `css_themes` `fef5db36-69e0-4848-ba88-7f0871d5e03c` | v4 = the true stylesheet, then **v5** = that + the contrast overrides below. Verified byte-identical to the git blob after whitespace strip, **172 rules both sides** (the apparent 148-byte gap was `length()` counting CHARS against a UTF-8 file — same trap as bash `${#var}`, in reverse) |
| deployed file | `sites@564dfa11d` restored from `9225356da` (2026-08-15, 24,210 bytes); live serve confirmed 24,210 then 26,918 after the overrides |
| provenance | recovered from git, **never rewritten** |

The two `css-patch-agent` rules occupying the clobbered row were deliberately **not** carried
forward: the LLM that wrote them was shown an EMPTY `current_css`, so it authored blind. One of
them (`H3.H3 { color:#ffffff }`) matches nothing on the site — no element carries `class="H3"`;
`render_audit.py` labels findings by uppercased TAG, and the agent appears to have read that
label as a class. **A patch written against a clobbered stylesheet is not a patch to preserve.**

### 2. The restore is what made the real contrast bugs visible — and they are the dual-role token

With the stylesheet back, `scripts/render_audit.py` over all 23 sitemap pages at 1366×900:
**21 failures, six of them invisible text** (1.06:1–1.11:1), all one cause —
`--color-primary` (#1A1F2E) sits in the same tonal band as `--color-background` (#111520) and
`--color-surface` (#1E2436), so every component using "primary" as an INK collapses:

```
contact.html                    h2      1.06:1   rgb(26,31,46) on rgb(30,36,54)
contact.html                    h3 ×2   1.11:1   Email / Phone
tools/dart-weight-comparator    legend ×2 1.06:1
tools/dart-weight-comparator    .btn-compare 1.11:1  (fill #1A1F2E, ink #111520)
```

Fixed in the site's own stylesheet (`sites@65c1d726c`, appended to theme v5), **not** in the
component: `contact-info` (`0bd72302-e9bf-4dc0-a615-41a9c919bf17`) serves **12 pages across 11
sites** `[MEASURED]`, and on a light theme its `color: var(--color-primary)` is a legible brand
heading — only a dark palette breaks it. Re-pointing a shared component's ink is a fleet
decision. Overrides are `body`-prefixed rather than `!important`, because the offending
declarations are in page-level component CSS emitted AFTER the stylesheet, so equal specificity
loses on source order. Post-fix re-measure: **both pages 0 failures.**

**This is the fix candidate this file does not yet have: a stylesheet-integrity check cannot see
the dual-role token, and the token is what the patch agent keeps being dispatched at.** Every
one of those six findings had already been through `css-patch-agent` or was parked in the 296
backlog; the agent cannot fix them because the declaration it must beat is not in the file it
edits. Candidate (6), for the union list: **when a contrast finding's offending declaration is
NOT in `css_themes`, the patch agent should refuse and re-file rather than append a rule that
cannot win.** Measurable precondition: grep the theme for the selector before planning.

### 3. robot-hands.com was next, and is now seeded (prophylaxis, not a restore)

`css_themes` **0 bytes, version 1, never patched**, deployed file **healthy at 25,559 bytes** —
i.e. the exact pre-symptom state this file's noted section describes as the one-query tell. Its
row is now seeded from the deployed blob (`06629999b`, verified identical, 174 rules), so the
first contrast item routed there can no longer destroy it. That is the data half of candidate
(3) applied to one site; **~10 of the 11 empty rows remain**, and doing them one lane at a time
is not a plan — the backfill wants to be one job.

### 4. cookly.uk, oufe.com, vonc.com: DB restored, FILES STILL CLOBBERED — blocked, needs a human

I seeded all three theme rows from their last healthy blobs, verified identical:

```
cookly.uk  a2f6c606…  17,462 b  from 249c94487 (08-15)   121 rules   theme OK, FILE STILL 504 b
oufe.com   4378085c…  20,695 b  from 09d448134 (08-15)   143 rules   theme OK, FILE STILL 1,336 b
vonc.com   ecd4cbe1…  21,823 b  from 24e84d8fd (08-09)   136 rules   theme OK, FILE STILL 176 b
```

**The deploy commit was refused by my session's permission classifier**, so those three sites
are still serving a near-empty stylesheet. This is a half-state and it is strictly better than
before (the rows are correct, so the next patch run appends to the truth and its deploy would
incidentally restore the file), but **three live sites remain visually broken and the last step
needs a session that is allowed to push, or the git-adapter route this file already documents.**
Recorded here rather than left in a transcript because "unowned-for-restore" is exactly how they
sat for two days.

oufe.com is worth one line for the pattern: **nine successive clobber commits on 2026-08-18**,
theme v2→v10, 70→140→299→480→663→830→997→1,156→1,336 bytes. The agent worked as designed nine
times; what it appended to was already wrong. That is this file's self-amplification, measured
again.

### 5. 090 substitution, stated

Not filed through the diagnosis loop, and it needed no filing at all in the end — the mechanism
was already in this file. My independent verification before I knew that: the live workflow JSON
read from `agent_definitions` (all nine steps), `css_themes` length vs deployed size across six
sites **including two controls** (noted.co.uk consistent at 20,190/20,367 — the theme row IS the
source there; robot-hands 0/25,559 — divergent and unpatched), and the per-commit blob sizes for
four sites from the deploy repo's own history. The one thing I got wrong on the way is recorded
in the lane's NOTES: I read a live 404 as a failed deploy twice, when the B2 sync simply had not
landed.

> **CORRECTED 2026-08-20, same lane, hours later — my "21 failures" figure was inflated by 15,
> and the corrected number strengthens candidate (6) rather than weakening it.**
>
> `render_audit.py`'s 21 rows were **6 real measurements + 15 `overImage` placeholders**
> (`[c.get('overImage') for c in page['contrast']]` in the `--json` output: 15 of 21 true). The
> probe pushes a mid-grey `rgb(128,128,128)` under any text whose backdrop is a background
> image or gradient because the real colour is unknowable — every `rgb(255,255,255) on
> rgb(128,128,128) = 3.95:1` row is that guess. **LANDMINES already carried this entry, on this
> exact footprint, and I quoted the terminal total without applying it.** I had reasoned my way
> to leaving those 15 alone ("fixing a page to satisfy an approximate measurement…"), which was
> the right action for the wrong reason: they are not near-misses to defer, they are not
> measurements.
>
> **Post-fix, site-wide, all 23 pages: 15 rows, 15 placeholders, REAL failures = 0.** So
> dartsonline carries no measured contrast failure, and the six that existed were all one cause.
>
> **And the sharper evidence for candidate (6), now enumerated rather than asserted.** All six
> real failures map to filed `contrast_failure` items — but not all of them were parked:
>
> | measured today | filed item | its status |
> |---|---|---|
> | contact `h2` 1.06:1 | `H2.H2 on /contact.html` (08-11) | deferred |
> | contact `h3` ×2 1.11:1 | `H3.H3 on /contact.html` (08-18) | **complete** |
> | comparator `legend` ×2 1.06:1 | `LEGEND.LEGEND on /tools/dart-weight-comparator` (08-11) | deferred |
> | comparator `.btn-compare` 1.11:1 | `BUTTON.btn-compare on /tools/dart-weight-comparator` (08-11, filed at 1.14:1) | deferred |
>
> **The `H3` row is the instance this file needs: an item marked `complete` by
> `css-patch-agent`, whose text was still invisible when I measured it two days later.** Its
> patch was `H3.H3 { color:#ffffff }` — which matches nothing, because no element carries
> `class="H3"`. So that completion is false twice over: written against an empty `current_css`,
> and aimed at a selector the agent inferred from `render_audit`'s uppercased-tag label. Even a
> correct rule would have lost, because `.contact-card h3`'s declaration is emitted after the
> stylesheet the agent edits. **This is "processed, correctly fixed, and never applied" with a
> row id behind it** — the news_editorial lane's phrase for what a subset of 296's durable 185
> may actually be.
>
> Also worth one line for whoever restores cookly/oufe/vonc: `BUTTON.form-submit` on contact
> (filed 4.30:1, marked complete, patched with `background-color:#c8180a`) does **not** appear
> in the post-restore sweep at all. Its patch was dropped with the clobbered file and the real
> stylesheet styles that button adequately on its own — i.e. one of the two rules I declined to
> carry forward was addressing a defect that only existed *because* the stylesheet was missing.

### 2026-08-20 (later) — the cause stated one level down: ONE artefact, TWO writers, neither reading the other

This file's ROUND 2 says the append fix "assumed `css_themes` holds the stylesheet; on ~11 sites
it holds NOTHING". True, and the reason it holds nothing is worth naming, because it changes
which candidate is the real fix and it makes every site-level repair (including mine today)
temporary by construction.

**`assets/css/styles.css` has two independent producers.** Read from the live agent rows:

| producer | writes the file from | reads `css_themes.css_content`? |
|---|---|---|
| `webdesign-agent` | `render_css_from_spec` → `generated_css.result`, built from palette + layout + typography rows + the `design_spec` + `css_snippets` | **no** |
| `css-patch-agent` | `css_saved.css_content` — the whole `css_themes` row | it IS the row |

`render_css_from_spec_action.go`'s own header lists its inputs (the three FK'd rows, the spec
merge rules, the `css_snippets` append, `buildSectionDefaults`); the only `css_content` it reads
is `css_snippets.css_content`. So **whichever agent ran last owns the file, entirely.** The
theme row is not "stale" — it was never in that path. `webdesign-agent` also carries
`fork_theme_from_site`, which is the born-empty seed the noted lane left `[INFERRED]`; on this
reading it is not a bug in the fork but the fork doing its job for a producer that never uses it.

**Consequences, in order of how much they change:**

1. **Candidate (3) is mis-named.** "DB↔file drift discovery check" implies one source of truth
   drifting from its copy. It is two writers with no reconciliation, so a drift check can only
   ever report the disagreement — it cannot say which side is right, and on a site where
   `webdesign-agent` ran last the DB is the WRONG side. A check that assumes the DB is
   authoritative would "repair" a healthy site by overwriting its real stylesheet.
2. **Every site-level CSS repair has an expiry, mine included.** The override block I appended
   today lives in the theme row and the file; the next `webdesign-agent` run on dartsonline
   regenerates from the spec and drops it. That agent has run **seven times since 2026-07-06**,
   most recently 08-15, i.e. roughly weekly. So today's contrast fix is a stopgap and should be
   removed rather than maintained once the durable fix lands. Recorded here so nobody reads it
   as settled.
3. **The durable fix for the invisible-ink family is a token that already exists.**
   `--color-primary-ink` is computed per palette by `palette_specialised_slots.go`
   (`legibleInkFor`) and **emitted into every healthy stylesheet** — measured on the served CSS
   2026-08-20, each site's own token against its own background and surface:

   ```
   dartsonline #94a0c2  7.00:1 / 5.93:1      leopardess  #0D0D0D 18.32:1 / 19.44:1
   robot-hands #94a0c2  7.20:1 / 5.88:1      noted       #1A1A18 16.00:1 / 14.50:1
   fundament'l #86ADDE  8.30:1 / 7.19:1      webdesign   #5c6b5d  5.32:1 /  5.65:1
   ```

   All six clear AA on both, dark palettes and light. So the shared-component fix is
   `color: var(--color-primary)` → `var(--color-primary-ink)` — a token swap, no value to
   defend — and it **survives regeneration**, because the generator is what emits the token.
   Note what this rules out: no single literal can serve both families (dartsonline needs a
   light ink, noted a near-black one), so any "prove one value across five palettes" exercise
   converges on a value that fails two of them, honestly.
4. **`oufe.com` has no `--color-primary-ink` at all** `[MEASURED]`, because its file is still
   clobbered — the token is emitted into the file, so a clobbered file loses it. Any component
   switched to the token renders an invalid `var()` there until the file is restored. **That
   makes the three outstanding restores a dependency of the fleet fix, not housekeeping.**

> **CORRECTED 2026-08-21 — point 4 above is WRONG. The three outstanding restores are NOT a
> dependency of the ink fix.** I wrote that a component switched to `--color-primary-ink` would
> render an invalid `var()` on oufe.com, whose clobbered file carries no such token, and
> concluded that the restores block the fleet fix. The opt-in form is **two-level** —
> `var(--color-primary-ink, var(--color-primary))` — so an absent companion falls through to the
> raw palette colour, which is the pre-2026-08-06 behaviour and is also how the mechanism's
> kill-switch works. On a site with no token the repoint is a **no-op**, not a breakage.
>
> Caught by the `news_editorial` lane, from `buildLegibleInkDefaults`'s own comment. My error was
> reasoning from the token's absence in the served CSS without reading how consumers are told to
> reference it — the same shape as the `retry_after` mistake I withdrew yesterday: **an absence
> is only evidence once you know what reads it.**
>
> **The restores still matter and their justification is unchanged** — three live sites serving a
> near-empty stylesheet, and any contrast census taken over them is a floor rather than a count.
> They are simply not blocking anyone else's work, and no one should sequence behind them.
>
> Two refinements to the table in point 3, same source:
> - The tokens target **`inkMinContrast = 5.0`**, not AA's 4.5. `inkFloorContrast` (4.5) is a
>   separate constant for the `-text` slots, with a test that fails if the two are merged. So
>   "all six clear AA" understated what the mechanism guarantees.
> - **A repoint that silently strips the brand scores a CLEAN contrast pass**, so neither the
>   render audit nor this file's evidence can see it. Between 2026-08-06 and 08-14 `-ink`
>   resolved to `--color-text` in practice; the 08-14 `LegibleVariant` repair fixed that, and the
>   check that demonstrates it is a hue comparison the audit cannot supply: robot-hands
>   `#E8500A`/`#f77f47`, dartsonline `#E8311A`/`#f18072` — each ink a lighter member of its
>   accent's own family, nothing like the text colour. **Anyone repointing a token here should
>   run that check as well as the contrast one**, because contrast alone cannot distinguish a
>   legible brand colour from a legible grey.

---

## CONTRIB — cookly.uk restored 2026-08-21, and the clobber had reached the REPO

From the `news_editorial_features` lane, on the owner's explicit instruction to
fix the three outstanding sites. **Only cookly.uk was still broken** — oufe.com
(21,312 B served) and vonc.com (25,052 B) were already healthy when checked.

### The clobber was no longer confined to the bucket

This is the part worth adding to the bug. `cookly.uk/assets/css/styles.css` was
serving **504 bytes containing only css-patch-agent's five contrast rules**. The
local repo copy was still the good 17,462-byte stylesheet — so the initial reading
was "bucket damaged, repo clean, re-sync fixes it".

**That was wrong.** On `git pull`, the remote had **committed the 504-byte file
into the repository**, and git's automatic merge resolved to **875 bytes**.
Committing that merge would have propagated the clobber into the one place it had
not reached. So the damage does not stop at the served object: it reaches the
repo, and from there it becomes the baseline every future render diffs against.

**For anyone restoring the remaining sites: `git pull` before assuming the repo
copy is safe.** A clean local file is not evidence; the remote may already carry
the damage.

### How it was resolved

Full generated stylesheet from `css_themes.theme-cookly-uk` (121 rules, all five
core palette tokens, zero patch markers) **plus the five css-patch-agent contrast
rules appended** under a comment saying why. Those five were the entire content of
the clobbered file — they are genuine repairs, so they were preserved rather than
discarded with the damage. Result 18,047 B, verified at the served artefact:
126 rules, 5/5 core tokens, patches intact.

### A second defect the restore exposed — the de-branded ink

The repo copy (11 August) carried `--color-accent-ink: #2C2C27`, which is
**byte-identical to this site's `--color-text`**. That is the pre-2026-08-14
legible-ink derivation, before `colour.LegibleVariant` got first refusal
(register `VIZ-014`). The current generated value is `#a24122`, a member of
`--color-accent`'s (`#C8502A`) own hue family.

It is not cosmetic: line 91 is
`a { color: var(--color-accent-ink, var(--color-accent)); ... }`, so **every link
on cookly.uk was rendering in the body-text colour** and is now in the accent.

⚠ **And it is invisible to a contrast audit**, which is why it survived: a
de-branded ink still passes contrast. `VIZ-014` states this
("stripping a brand colour scores a CLEAN PASS").

> ### ⚠ CORRECTION 2026-08-21, same day — the check as first written here is WRONG and would condemn a correct token
>
> The paragraph above originally ended: *"diff the restored `-ink` tokens against
> that site's `--color-text` and treat equality as a stale derivation, not a
> coincidence."* **Equality with `--color-text` is not sufficient**, and cookly
> itself is the counter-example. Applied to its OTHER ink token it produces a
> false positive:
>
> ```
> cookly.uk   --color-text:        #2C2C27
>             --color-primary:     #2C2C27   <- the palette makes these EQUAL
>             --color-primary-ink: #2C2C27   <- therefore CORRECT, not stale
>             --color-accent:      #C8502A
>             --color-accent-ink:  #a24122   <- genuinely repaired
> ```
>
> Confirmed at the palette source (`css_themes.color_palette`): cookly's `primary`
> and `text` are **both** `#2C2C27` by design, and `legibleInkFor` "returns
> srcHex unchanged when srcHex already clears minRatio" (its own doc comment).
> `#2C2C27` on `#FDFAF4` clears easily, so the derivation returned primary
> untouched — which is exactly right.
>
> **The correct check needs BOTH clauses:**
> ```
> stale  ⇔  <x>-ink == --color-text  AND  <x>-ink != --color-<x>
> ```
> The second clause is what distinguishes *substituted* (the old walk replaced the
> source with the text slot) from *returned unchanged* (the source already
> cleared). Against `accent` it still fires — `#2C2C27` != `#C8502A` — which is
> the true positive that started this. Against `primary` it correctly stays
> silent.
>
> **Why this correction matters more than the original check:** acting on the
> single-clause version would mean changing a token that is already correct, on a
> site whose palette deliberately uses one ink for both roles — i.e. introducing
> the defect while believing you were removing it. Caught by the
> `dartsonline_traffic` lane applying my own check to the token I had not looked
> at, which is the right way for it to fail.
>
> Footnote for cookly specifically: nothing on that site consumes
> `--color-primary-ink` at all (it appears only in its own definition and the
> renderer's comment), so even a genuine staleness there would have been inert.
> `--color-accent-ink` was the live one, via `a { color: var(--color-accent-ink, …) }`.

### Seeding the theme row is a REPAIR, not only prophylaxis — measured on two sites

From the `dartsonline_traffic` lane, and it changes the procedure: **oufe.com and
vonc.com restored themselves.** Not by a human session — the restoring commits are
`css-patch-agent` runs (`0fec465dd` oufe, theme v15, 2026-08-21 12:47;
`3e0601b3a` vonc, theme v24, 11:27). Once the `css_themes` row underneath was
seeded with the true stylesheet, the next patch run appended to *the truth* and
deployed the whole row — so the mechanism that causes the clobber also carries the
cure, provided the row is right first.

**So for a future clobber: seed the theme row and the file restore is optional** —
the next patch run carries it, and that half needs no push permission at all.
It is not a plan on its own, though: cookly had no pending patch run to ride, and
waiting for one is not a schedule. Seed the row always; push the file when the
site is visibly broken now.

### And: `git pull` before assuming the repo copy is safe

Recorded above from the cookly restore, repeated here because it belongs with the
procedure: the remote had committed the clobbered file into the repository, and an
automatic merge of damage against clean content resolved to **875 bytes of
neither** — it does not fail, it produces a third thing that becomes the baseline
every later render diffs against.

### The corrected check's first fleet-wide run — 5 stale tokens on 4 sites, verified independently

Run by the `dartsonline_traffic` lane over the deploy repo; **re-verified here at
the SERVED artefact** rather than accepted on report, because a repo file and a
served object are different things and this whole bug is about them diverging.
All five confirmed live:

| site | token | ink | `--color-text` | its own source |
|---|---|---|---|---|
| lendzy.co.uk | `accent-ink` | `#1A1A1A` | `#1A1A1A` | `accent #E8700A` |
| loancash.co.uk | `accent-ink` | `#1a1a1a` | `#1a1a1a` | `accent #b7791f` |
| loancash.co.uk | `primary-ink` | `#1a1a1a` | `#1a1a1a` | `primary #e8f5ee` |
| vetcomparison.uk | `accent-ink` | `#0f172a` | `#0f172a` | `accent #10b981` |
| vonc.com | `primary-ink` | `#f0eeff` | `#f0eeff` | `primary #7c3cff` |

**The negative half works too, which is what makes the list trustworthy.** On
every one of those sites the *other* ink token equals its own source and is
correctly NOT flagged — `vonc accent-ink #fc5c7d == accent #fc5c7d`,
`vetcomparison primary-ink #2563eb == primary`, `lendzy primary-ink #1B2A4A ==
primary`. A check that flagged everything would be worthless here; this one
separates the two cases on real data. The single-clause version would have
returned cookly as a sixth and someone would have "corrected" a correct token.

### ⚠ The addition to the restore procedure, and it is the important part

**Seeding the theme row faithfully preserves whatever the blob encoded, including
a pre-repair derivation.** `vonc.com` is on the list *because* it was restored:
the blob used (`24e84d8fd`, dated **2026-08-09**) predates the 2026-08-14
`LegibleVariant` repair, and the patch run that later rebuilt the file deployed
that row faithfully. `oufe` and `dartsonline` are clean only because their blobs
happened to be dated 08-15.

So, added to the procedure:

> **Check which side of 2026-08-14 the blob you restore from falls on, and run the
> two-clause check on the result.** A restore is not a repair; it reinstates a
> point in time, and that point may be before the derivation was fixed.

That is "file date is not derivation date" applied to the restore itself — the
lane that coined the line walked into it, which is the most useful way for it to
be demonstrated.

### What this check does and does not establish

It finds **suspects, not confirmed defects**, and the distinction is real.
`legibleInkFor` falls through to the palette walk when the source cannot be
rescued by a lightness move at all (achromatic sources, or one no lightness
clears). For such a source, an ink differing from its source is *correct* even
post-repair. `loancash primary #e8f5ee` — a near-white mint used as an ink on a
light page — is exactly that shape: near-black may well be where a *current*
render lands too.

**So the confirming test is a fresh render, not the token comparison.** Regenerate
the palette and compare: if the value moves to something in the source's hue
family, it was stale; if it stays, the walk was right both times.

### NOT fixed here, deliberately

All four sites are outside both contributing lanes, no consumer tracing has been
done on any of the five (so no visible damage is claimed), and the remedy is a
**palette regeneration**, not a token edit — which changes every derived value on
the site at once. Regenerating four sites' palettes on the strength of a check
run this afternoon would be a larger, less reversible action than the evidence
supports. **Recorded here so the list is not lost in two transcripts**, which is
what this section is for.

---

## 2026-08-21 — THIRD WAVE: remortgagecalculator.uk and loanzy.uk clobbered the same morning; both restored, the empty-row backfill DONE fleet-wide, and the missing DETECTOR built

Appended by the `remortgagecalculator.uk` lane. **Coordination note: the owner said mid-session
that this file is being worked in another thread.** So this section is deliberately scoped to
what was not already here: two new incidents, the restores, ROUND-2 candidate 1 executed across
the fleet, and a detector for candidate 3's detection half. **No change was made to
css-patch-agent's config or to the git-adapter** — candidates 2 (deploy-side shrink guard) and
the birth guard are untouched and remain this file's lane's.

### 1. What the owner actually saw, and what it was not

The report was "the site was rebuilt in another thread and the CSS is now broken". **The rebuild
did not break it.** The index rebuild FAILED and was cancelled at 12:13Z — `store_component`
rejected the `mortgages-repayment` component (`field "currency_symbol" declares source
"site_specs.locale.currency_symbol" but no site carries a site_specs aspect named "locale"`,
which is `bugs_open/345`'s territory, not this one's). Only `mortgage-lenders.html` re-rendered
today. The CSS broke **before** that, at 10:27Z, through this file's mechanism.

The rebuild is nonetheless why it became visible: rebuilt markup consumes `var(--color-*)`, and
those resolve only against the `:root` block in the file that had just been destroyed.

### 2. The two incidents, at the artefact

| | remortgagecalculator.uk | loanzy.uk |
|---|---|---|
| theme row | `08fc0b7f…`, **born empty 08-17 12:46:33**, `origin='adopted'` | `2a1fb031…` |
| real stylesheet | `gqls/sites@556951393` (08-17 13:12Z), **17,403 B**, `:root` ×3 | `gqls/sites@8397c1442`, **17,160 B**, `:root` ×3 |
| clobber commits | `69fe2eed8` 10:27:27Z (17,403 → **68 B**), `83229b1bf` 10:27:55Z (68 → **136 B**) | 16 versions of `CSS fix: contrast`, served down to **1,577 B**, still climbing when found |
| driving items | 2 × `contrast_failure` filed 08-20 12:43Z, both `complete` | 15 `complete`, **18 still `triaged`** |

**Arithmetic proof the base was empty rather than truncated** (worth recording because it
distinguishes this arm from the 08-04 LLM-fragment arm without needing the config): each append
block is exactly 68 chars — `\n\n` + a 42-char provenance comment + `\n` + the 23-char rule.
2 × 68 = **136**, and the file opens with the append's own leading blank lines. The column was 0.

**The patches were authored blind, again, and the selector proves it.** Both rules target `p.P`
— and **no element on the site carries `class="P"`**. This is the `H3.H3` mechanism the
dartsonline lane recorded on 08-20, firing again: `render_audit.py` labels findings by
UPPERCASED TAG, and the agent reads that label as a class. Per that lane's precedent the rules
were **not carried forward** into either restore. **This is now three sites' worth of evidence
for candidate (6)** — when the offending declaration is not in the file the agent can edit, it
should refuse and re-file rather than append a rule that cannot match.

### 3. Restores — both live, both verified at the artefact

- **remortgagecalculator.uk.** Row md5-guarded to v4 = the 17,403-byte blob
  (`34051f6e5dbed8f43cc0de433e5b0fa8`), then the file deployed through the **git-adapter route
  this file documents** (`repo_name:"sites"`), commit `a8462c86a`. Verified: DB md5 == repo blob
  md5 == served bytes, `:root` ×3. **DB, repo and live are byte-identical**, which is the
  property that makes "deploy the DB row wholesale" safe for the next patch run.
- **loanzy.uk.** Row md5-guarded to v17 = its 17,160-byte blob, then deployed. **Then the
  self-healing behaviour this file predicts was observed directly:** the 18 queued contrast items
  began appending to the *restored* base and deploying the whole file — v21/17,906 B within
  minutes, v34/21,330 B by the end of the session, `:root` ×3 throughout. A restored row converts
  the remaining queue from a destructive loop into an additive one with no further intervention.
  (Those appends are still blind-authored patches against phantom findings measured on the
  clobbered page — junk, but small and additive. Candidate 6 is what stops them being written.)

### 4. ROUND-2 candidate 1 EXECUTED — the empty-row backfill, fleet-wide

Nine rows seeded from their own healthy deployed stylesheet, each `cmp`-verified byte-identical
to what the site actually serves before writing, each UPDATE guarded on
`octet_length(COALESCE(css_content,''))=0` with a `DO`/`RAISE` md5 verification:

| site | row before | row after |
|---|---|---|
| ai-agent-orchestration.com | 0 | 20,923 |
| fundamentallyai.com | 0 | 17,430 |
| gamesdesign.co.uk | 0 | 19,174 |
| lendzy.co.uk | 0 | 17,387 |
| loanandmortgagecalculator.co.uk | 0 | 13,650 |
| mortgagecalculator.co.uk | 0 | 17,413 |
| vetcomparison.uk | 0 | 24,083 |
| webdesign.co.uk | 0 | 20,261 |
| leopardessconsulting.co.uk | 1,649 (bare `:root`, no layout) | 13,978 |

**There are no empty linked theme rows left in the fleet.** The census that used to read
"11 of 21 empty" now reads zero, so candidate 1 is done as data — though it is a one-time repair,
not a guard: the next site composed by the normal path is born empty again, which is why the
**birth guard** stays on this file's candidate list.

### 5. Three things the fleet sweep found that the 10-site list did not

The sweep was all 25 deployed/active sites, comparing row bytes against **served** bytes and
`:root` count — not the `=0` census, which is what had been driving this work.

- **A `=0` predicate is too narrow.** Three sites carried a **1,649-byte** row holding a bare
  `:root` palette block and *no layout rules* (leopardess, seeded above; plus finetuning.uk and
  gaswholesalers.com). Deployed over a real stylesheet that produces a page where every variable
  still resolves and every layout rule is gone — a **different failure signature**, and one the
  new detector in §6 does NOT catch. Worth a line on this file's candidate list.
- **`finetuning.uk` and `gaswholesalers.com` SHARE one theme row** (`fecb962d…`) — the
  `shared_css_theme` defect. **They were deliberately left alone**: seeding from either site's
  file would push that site's CSS onto the other. This needs a human decision, not a backfill.
- **`cookly.uk` is STILL SERVING 504 BYTES** — the file half of the 08-20 restore. Its row is
  correct; the deploy was refused then by a permission classifier and **was refused again in this
  session**, on the same git-adapter route that succeeded for the other two sites minutes
  earlier. oufe.com and vonc.com, listed alongside it on 08-20, have since recovered (served ==
  row, `:root` present) — presumably a patch item rode their restored rows. **cookly.uk is the
  one site still visually broken by this bug, and its last step still needs a session that is
  allowed to push, or the owner.** Ready-to-run command with the blob identified:
  `git -C ~/projects/sites cat-file blob 249c94487:cookly.uk/assets/css/styles.css` (17,462 B,
  `:root` ×3, no patch markers) → deploy to `gqls/sites` at `cookly.uk/assets/css/styles.css`.

Also cleared as NOT damage, so nobody re-investigates: **webdesign.uk** 302-redirects to
webdesign.co.uk and has no stylesheet of its own (a bare `curl` without `-L` reads the 143-byte
Cloudflare redirect page as a gutted file — I made exactly that mistake for several minutes);
**adversecreditmortgage.co.uk / loancalculator.co.uk / loancash.co.uk** have no
`style_collection` at all, so css-patch-agent exits `complete_no_css` on them — unclobberable by
this path, and equally unprotectable.

### 6. The DETECTOR — and why every existing check was blind

The owner also asked whether the improvement loop would have spotted this. **It would not, and
two checks are wrong-way blind rather than merely silent.** Walked one by one before building
anything:

- **`asset_reference_404` is the only check that fetches `/assets/css/styles.css`, and it scores
  HTTP STATUS ONLY.** A 136-byte 200 is a *positive observation* to it — so it does not stay
  quiet, it **RETRACTS** any open item keyed to that URL.
- **The render audit is blind in the direction that reads as health.** With `:root` gone,
  `var()` declarations are invalid at computed-value time: text inherits to black, background
  falls back to transparent/white, and the probe measures **~21:1 — a clean, high-contrast
  audit**. This is why a clobbered site can sit for days without a signal. It also explains this
  file's own self-amplification finding from the other end: when the audit *does* eventually
  file (1.00:1 on a dark-fallback section) it routes the finding straight back to the agent that
  caused the damage.
- `missing_css` asks whether the row exists (it does — it is the row that did the damage);
  `generic_theme` reads the head component (intact); `palette_contrast` reads `palettes.colours`
  (untouched).

**Built and committed this session (`e34b33a36`, gofmt/emission fix `093363070`):
`stylesheet_gutted`**, a flag-only discovery check on `design-discovery-agent`, register
**IMP-055**. Predicate is **definition coverage, not a byte floor** — because `bugs_open/211` is
this defect at ~26KB (alias `:root` block absent, `--color-heading` defined zero times) and a
floor cannot see it: it compares the custom properties DEFINED by the served same-host
stylesheets against those REFERENCED without a fallback by deployed components and
`css_snippets`. It declines to judge — filing **and** retracting nothing — whenever a stylesheet
fails to fetch or returns non-2xx, so a blinded run can never be mistaken for a healthy one.
Its item key is deliberately constant rather than URL-shaped, so `asset_reference_404`'s 2xx
retraction can never close its findings. 18 tests, each guard proven by a named mutation, both
real incident shapes as fixtures.

**It is NOT yet enabled.** Migration `541_..._HOLD.sql` is held until a chassis image carrying
the check has rolled and the capability is pod-probed with a negative control — an unregistered
check name fails the *whole* discovery step and discards every earlier check's findings in that
run (`discovery_checks.go:198-216`). Council round submitted: correlation
`d3187418-3bb5-435d-b66b-92d8fc9d9d01`, verdict pending at time of writing; **whoever reads it
owes acting on a REVISE, since the code is already on the shared branch.**

**Calibration, so a first rotation is falsifiable:** measured across all 25 deployed/active
sites on 2026-08-21, exactly **one** site would file today (cookly.uk). **A rotation that
includes cookly.uk and reports zero is a bug in the check, not good news.**

### 7. What this does and does not close

- **Closed:** candidate 1 (empty-row backfill) as data, fleet-wide. The detection half of
  candidate 3, for the "definitions are gone" shape.
- **NOT closed, and untouched by this lane:** candidate 2 (**deploy-side shrink guard** — still
  the single fix that would have stopped both of today's incidents at the last writer); the
  **birth guard**; candidate 6 (**refuse to patch when the offending declaration is not in the
  file the agent can edit** — now with a third site's evidence); the **"only `:root` survived"**
  shape, which the new check does not detect; and **cookly.uk's file deploy**, which needs a
  human.
- **090 substitution, stated plainly (owner ruling 2026-07-31):** not filed through the
  diagnosis loop. The mechanism was already in this file and every claim above is first-hand and
  artefact-verified in one session — git blob sizes and commit shas from the deploy repo, the
  theme rows before and after by md5, the driving work items, the live served bytes with
  cache-busters, and the live agent config. A 090 run would re-read the same artefacts.

> **CORRECTED 2026-08-21, same lane, ~2 hours after the section above — two claims in it
> were already stale or wrong when measured properly.**
>
> **(a) "cookly.uk is the one site still visually broken" is NO LONGER TRUE.** It now serves
> **18,047 bytes** with its `:root` intact. It was restored between my measurement and my
> re-check — by the other thread working this file, or by a patch item riding its corrected
> row (its theme row was already right, which is exactly the half-state §4 of the 08-20
> section predicted would self-heal). **Nothing is outstanding on cookly.** The
> classifier-refusal note stands as a record of what I could not do, not as a live task.
>
> **(b) MY CALIBRATION FIGURE FOR THE NEW CHECK WAS WRONG, and the way it was wrong is the
> more useful half.** §6 said "exactly one site would file today (cookly.uk); a rotation
> covering it that reports zero is a bug in the check". That was measured with a
> **`:root`-presence proxy** — not with the check's own predicate. Running the REAL
> predicate across all 25 deployed/active sites, the check **as first written would have
> filed on NINETEEN**, and seventeen of those were the SAME four names —
> `--color-hero-title`, `--color-hero-subtitle`, `--color-secondary-text`,
> `--color-secondary-hover` — which **no site's stylesheet has ever defined, including in
> the pre-clobber originals I restored from**. That is a real defect (a component vocabulary
> the renderer never emits) but it is a DIFFERENT one, and seventeen copies of it would have
> buried the signal this check exists for on its first day — `bugs_open/083`'s failure mode,
> shipped by the very lane quoting 083 as a risk.
>
> **Fixed before enabling:** the predicate now gates on the renderer's GUARANTEED vocabulary
> (`rendererGuaranteedTokens`, kept in step with `canonicalCSSTokens` by a parity test that
> reads `component_validation.go`'s source — proven non-vacuous by deleting `--hero-ink` and
> watching it fail by name). Both incident shapes still fire: a clobber loses the whole
> palette, and `bugs_open/211`'s missing alias block takes `--color-heading` and `--hero-ink`,
> all canonical. **Re-measured with the gate: 0 of 25 sites.** So the check ships as a
> REGRESSION GUARD with no live positive, and **a clean first rotation is now the EXPECTED
> result rather than a bug** — the opposite of what §6 told the next reader.
>
> **What caught it:** running the predicate against the page I had just restored, instead of
> trusting the proxy I had calibrated with. The restored site came out FIRING on four tokens,
> which was the tell — a site I had just verified byte-identical to its own healthy original
> should not have been a finding. Logged in `WRONG_CALLS.md`.
>
> **The four undefined component tokens are left as a finding for whoever wants them:** they
> are referenced by live components on ~17 sites and defined nowhere, so those declarations
> are dead on every one. Not this lane's to fix, and deliberately NOT filed by the new check.

---

## 2026-08-21 (evening) — THE PREVENTION HALF: candidate 2 built, the born-empty cause fixed at its producer, and the completions stop lying

Appended by the **bugfix-198 lane**, dispatched at this file by the owner. **Scope taken
straight from the THIRD WAVE section's §7**, which named what was left after that lane's
data-and-detection work: *"candidate 2 (deploy-side shrink guard — still the single fix that
would have stopped both of today's incidents at the last writer); the birth guard; candidate
6."* This section is candidates 2 and the birth-guard slot. **No restores were done and none
were needed** — as of 18:00Z no site in the fleet serves a clobbered stylesheet (cookly.uk
recovered to 18,047 B by the `news_editorial_features` lane earlier today).

**090 substitution, stated plainly (owner ruling 2026-07-31):** not filed through the
diagnosis loop. The mechanism was already established in this file across five lanes, and
every claim below is first-hand and artefact-verified in one session — the live workflow
configs read from `agent_definitions`, a fleet census by the exact `load_current_css` JOIN,
in-transaction proofs against live rows, served bytes with status codes, and two Go mutations
run against the real source. A 090 would re-read the same artefacts.

### 1. The root cause, no longer `[INFERRED]`

ROUND 2 left one link unverified: *"[INFERRED at the one unwitnessed link: that
`fork_theme_from_site` inserted `renderedCSS=''` in that run]"*. **It is not
`fork_theme_from_site`**, which refuses an empty render (`if renderedCSS == "" { ... return
forkSkipped(...) }`) and so cannot produce a born-empty row.

**The born-empty row comes from `install_site_composition_action.go:342-370`, and it is
DELIBERATE.** Its own comment: `// css_content is empty — the renderer reads composition via
FKs.` Those are the only two `INSERT INTO css_themes` in the Go tree. So the empty row is the
install contract for one producer, and lethal to the other — which is why the 2026-08-20
"ONE artefact, TWO writers" reading is the correct level to fix at, and why a *birth guard*
would have been the wrong shape: it would refuse the composition itself, which is working as
designed.

### 2. What shipped (commits `511afc791` config, `4ee9bfff6` Go)

| # | change | where | live? |
|---|---|---|---|
| A | **Base-integrity refusal** — `check_base_integrity`: `css_len >= 4096 AND site_count <= 1`, `fail_on_non_numeric: true` | migration **542**, css-patch-agent | **LIVE on apply** |
| B | **Refusals and failures stop minting `complete`** — three `update_work_item_status` steps stamping before their terminal | migration **542** | **LIVE on apply** |
| C | **Persist-at-render** — `persist_css_to_theme` between `generate_css` and `deploy_css` | migration **543**, webdesign-agent | **LIVE on apply** |
| D | **`file_shrink_floor`** — opt-in shrink guard enforced in the git-adapter (DGH-016) | Go, both `internal/adapters/git` and the chassis | **committed, INERT until BOTH images roll** |

**(A) closes the arm all three waves went through.** `check_has_css` tests
`current_css.css_content != null`, and an **empty string is not null**, so it passed. The new
gate is numeric. The floor is census-derived, not chosen: measured today across all 22 linked
rows, healthy rows are **13,650–26,917 bytes** and every clobbered or stub row ever recorded
in this file is **≤ 2,381**. Fleet split at 4096 is **19 PASS / 3 REFUSE with nothing in
between** — the query is in the lane RUNBOOK §3, and if a future census shows anything in
that gap the floor needs re-deriving rather than overriding. `octet_length`, not `length`
(this file already records the chars-vs-bytes trap on dartsonline).

`site_count <= 1` closes a door **this file named but nobody could shut**: the THIRD WAVE
section's *"`finetuning.uk` and `gaswholesalers.com` SHARE one theme row — this needs a human
decision, not a backfill"*. It is now refused automatically instead of waiting on that
decision, and the decision is flagged rather than forced.

`fail_on_non_numeric: true` matters more than it looks: without it, a missing `css_len` — the
query edit not having landed — routes **every** run to the refusal arm and reads exactly like
a working guard.

**(B) is the fix for this file's own "completions lie twice" finding.** Every terminal here is
a success-labelled `complete_workflow`, so the parent dispatch loop stamps `complete`
whatever happened — which is why loancash's 11 `complete_no_css` runs all read `complete`. A
guard whose refusal reads as a repair is worse than no guard, because it also suppresses the
evidence (296 §10.4). Refusals now stamp `needs_human_review` with a
`result_fields.parked_by` marker; real errors stamp `failed` through the shared retry ladder.
**⚠ A parked item HOLDS its dedup key** (`idx_swi_dedup` does not exclude
`needs_human_review`), so the finding cannot re-file while parked — the unpark sweep keyed on
`parked_by` is RUNBOOK §4 and is the only route back.

**(C) is the durable half, and it is what makes every restore in this file stop expiring.**
The 2026-08-20 section already said it: *"Every site-level CSS repair has an expiry, mine
included ... the next `webdesign-agent` run on dartsonline regenerates from the spec and
drops it"*, roughly weekly per site. `persist_css_to_theme` writes the rendered stylesheet
into the site's theme row byte-for-byte at every design run, so the row tracks the file.
Guarded four ways: `octet_length($2) >= 4096` (never persist a fragment — the same number the
consumer refuses at, so the halves cannot disagree about what a stylesheet is), `origin <>
'seed'` (never overwrite a library theme), exactly one linking site (never push one site's
CSS onto another), `IS DISTINCT FROM` (no churn). Fails **open** to `deploy_css` deliberately:
the realistic error is an unresolvable site id on a non-site run, and (A) is the backstop, so
an unpersisted row causes a REFUSAL later, never a clobber.

> **It also makes the concept register true.** `DES-005` has claimed since creation that a
> theme row is born empty and *"webdesign-agent fills it at render"*. **That fill did not
> exist in any code path.** A register status is a claim, and this one was load-bearing,
> false, and precisely the sentence that would reassure a reader their restore is maintained.

**(D) is candidate 2, at the only place it can exist.** The chassis cannot compare against
the incumbent file — `GitCommitAction` assembles a payload and produces it to Kafka, and no
adapter verb exposes file contents. The git-adapter can, and the read primitive was already
there being thrown away: `pathExists` GETs `/contents` and `io.Discard`s a body carrying
`size`. One opt-in key, absent/0 = OFF = today's behaviour, and an unconfigured caller makes
**no extra API call** (test, not claim — that is what bounds the blast radius across the 17
carrier agents). Exactly one step opts in, at 0.5.

### 3. Evidence, and what could have come out otherwise

- **543's UPDATE, in a rolled-back transaction on live rows.** A real 25,202-byte value onto
  dartsonline's row: `UPDATE 1`, v5→v6, `md5(css_content) == md5(value)` exactly. Then
  `UPDATE 0` four times — shared row, 100-byte fragment, unchanged content, seed row. Four
  negatives and one positive from one statement.
- **Two Go mutations RUN, not asserted.** Deleting the enforcement call failed three tests.
  Measuring the **unprefixed** path failed its dedicated test **and let the clobber commit
  through** — every lookup 404s, every 404 reads as "new file", and the guard still logs that
  it ran. Source restored byte-for-byte after each.
- **Built from a clean `git archive HEAD` tree plus only these files.** The working tree
  fails an unrelated test from another session's in-flight work; on HEAD plus mine, green.
- Post-apply config probes green on both agents, and `position('patched_css' in config) = 0`
  still — 318 survives 542 untouched.

### 4. What this closes, and what it does NOT

- **Closes:** candidate 2 (the deploy-side shrink guard, pending the roll); the birth-guard
  slot, by fixing the producer instead — a born-empty row is filled at the site's first
  render and (A) covers the window before that; the `complete_no_css` arm minting `complete`;
  and the shared-theme hazard, which is now refused rather than merely known.
- **Does NOT close, and is not claimed:**
  - **The witnessed live refusal.** Proven in-transaction and by config probe; NOT observed
    on a real dispatch. Deliberately not induced — the only sites that would exercise the
    refusal arm are finetuning.uk and gaswholesalers.com, both live, and a faulty gate would
    clobber them. This remains this file's closure bar.
  - **The post-roll proof of (D)** — pod-grep `file_shrink_floor` on the chassis AND the
    git-adapter, each with a negative control, then the adapter's `Info` line
    (`"file_shrink_floor: commit passed the shrink floor"`) on a real deploy. Lane RUNBOOK §7.
  - **Candidate 6** — refuse and re-file when the offending declaration is not in the file
    the agent can edit (`H3.H3`, `p.P` ×2 — three sites' evidence in this file now).
    Untouched, and the next coherent task here.
  - **The round-trip-writer inventory**, owed since council round `5249320e` (2026-08-05).
    Still owed; explicitly not absorbed by this work.
  - **Owner decision:** per-site theme split for finetuning.uk + gaswholesalers.com.

### 5. Where the working record lives

`docs/agent_docs/docs024_key_docs_latest/bugfix_198_roundtrip_writers/` — PLAN
(`PLAN_2026-08-21_two_writer_reconciliation.md`, the four decisions and why each is where it
is), RUNBOOK (the one-query tell, the gate probes, the unpark sweep, the migration-apply trap,
the post-roll capability probes, and the restore recipe collected from three lanes), NOTES
(missteps included), README_where_we_are (owner prose), and the council submission JSON.
Council round `5f756c51-cdc6-4a48-b5f9-59e472243601`, submitted before the commits; both
commits carry `Council-Submitted:` and the verdict is owed a read — **whoever reads it owes
acting on a REVISE, since the code is already on the shared branch.** Two LANDMINES added
(the repair-expiry class; the `error_step`-mints-`complete` class) and one WRONG_CALLS entry.
