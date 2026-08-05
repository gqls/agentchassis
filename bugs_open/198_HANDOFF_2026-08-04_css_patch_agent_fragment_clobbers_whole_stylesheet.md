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
