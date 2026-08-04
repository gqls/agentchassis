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
- **vm-sites HEAD + live site NOT yet restored** — the session's outward-write channels
  (gh contents PUT; a hand-published `system.adapter.git.requests` commit message) were
  both blocked by the harness permission classifier; handed to the owner with exact
  commands. The four fixes themselves are correct and preserved in the restore.

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
